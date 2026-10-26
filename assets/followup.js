// "Belum dibalas": chats whose last message came from the other side and has waited longer than a set
// time. Sources, newest information wins:
//   - chat list rows: a delivery/read tick in the preview means the last message is yours (answered);
//     otherwise the row's timestamp tells since when the other side has been waiting
//   - the open chat: the last message carries a tick when it is yours; a group's header has a
//     subtitle listing its members
//   - incoming notifications (notify.js), for chats that are not on screen
// Muted chats and groups are left out by default. Nothing is sent to WhatsApp.
//
// WhatsApp Web draws its icons as SVGs whose <title> names the icon (wds-ic-read, ic-notifications-off,
// ic-group-filled, ...); older versions used data-icon attributes, which are still checked.
(function() {
	if (window.self !== window.top) return;

	var CFG_KEY = 'wa_followup_cfg', STATE_KEY = 'wa_followup';
	var DEFAULT_CFG = { enabled: true, minutes: 60, notify: true, includeGroups: false };
	var SHOW_DAYS = 3;   // older waits are taken as finished conversations ("ok, makasih")
	var FORGET_DAYS = 7;

	function loadJSON(k, def) { try { var v = JSON.parse(localStorage.getItem(k)); return v == null ? def : v; } catch (e) { return def; } }
	function saveJSON(k, v) { try { localStorage.setItem(k, JSON.stringify(v)); } catch (e) {} }
	function toast(msg) { if (typeof showAddonToast === 'function') showAddonToast(msg); }
	function cfg() { return Object.assign({}, DEFAULT_CFG, loadJSON(CFG_KEY, {})); }

	// state: chat -> { since, notified, ignored, group, muted }
	var state = loadJSON(STATE_KEY, {});
	if (!state || typeof state !== 'object') state = {};
	// Entries saved without a first-seen time count as seen now, so they never cause backlog toasts.
	Object.keys(state).forEach(function(k) { if (!state[k].seen) state[k].seen = Date.now(); });
	function save() { saveJSON(STATE_KEY, state); }

	// parseRowTime turns the time shown in a chat row into a timestamp (approximate for older days).
	var DAYS = { minggu: 0, senin: 1, selasa: 2, rabu: 3, kamis: 4, jumat: 5, "jum'at": 5, sabtu: 6,
		sunday: 0, monday: 1, tuesday: 2, wednesday: 3, thursday: 4, friday: 5, saturday: 6 };
	function parseRowTime(text, now) {
		text = String(text || '').trim().toLowerCase();
		var d = new Date(now);
		var m = text.match(/^(\d{1,2})[.:](\d{2})(?:\s*([ap])\.?m\.?)?$/);
		if (m) {
			var h = parseInt(m[1], 10) % 24;
			if (m[3] === 'p' && h < 12) h += 12;
			if (m[3] === 'a' && h === 12) h = 0;
			var t = new Date(d.getFullYear(), d.getMonth(), d.getDate(), h, parseInt(m[2], 10)).getTime();
			return t > now + 60000 ? t - 86400000 : t;
		}
		if (text === 'kemarin' || text === 'yesterday') return new Date(d.getFullYear(), d.getMonth(), d.getDate() - 1, 12).getTime();
		if (DAYS.hasOwnProperty(text)) {
			var back = (d.getDay() - DAYS[text] + 7) % 7 || 7;
			return new Date(d.getFullYear(), d.getMonth(), d.getDate() - back, 12).getTime();
		}
		m = text.match(/^(\d{1,2})\/(\d{1,2})\/(\d{2,4})$/);
		if (m) {
			var a = parseInt(m[1], 10), b = parseInt(m[2], 10), y = parseInt(m[3], 10);
			if (y < 100) y += 2000;
			var dayFirst = a > 12 || !/^en-us/i.test(document.documentElement.lang || navigator.language || '');
			var day = dayFirst ? a : b, month = dayFirst ? b : a;
			return new Date(y, month - 1, day, 12).getTime();
		}
		return null;
	}
	window.waParseRowTime = parseRowTime;

	var TICK_ICON = /^(wds-ic-)?(read|delivered|sent|pending)$|(status|msg)-(check|dblcheck|time)/;
	var MUTED_ICON = /notifications-off|muted/;

	// iconNames lists the icons inside el: SVG titles and (older WhatsApp) data-icon values.
	function iconNames(el) {
		var names = [];
		el.querySelectorAll('svg title').forEach(function(t) { names.push((t.textContent || '').trim()); });
		el.querySelectorAll('[data-icon]').forEach(function(i) { names.push(i.getAttribute('data-icon') || ''); });
		return names;
	}
	function hasTick(el) { return iconNames(el).some(function(n) { return TICK_ICON.test(n); }); }
	window.waHasTick = hasTick;

	function rowInfo(row) {
		var name = window.waRowName ? window.waRowName(row) : '';
		if (!name) return null;
		var icons = iconNames(row);
		var out = {
			name: name,
			mine: icons.some(function(n) { return TICK_ICON.test(n); }),
			muted: icons.some(function(n) { return MUTED_ICON.test(n); }) || !!row.querySelector('[aria-label="Chat dibisukan"], [aria-label="Muted chat"]'),
			group: icons.indexOf('ic-group-filled') !== -1,
			time: null
		};
		var stamp = row.querySelector('[data-testid="chat-list-timestamp"]');
		if (stamp) out.time = parseRowTime(stamp.textContent, Date.now());
		var spans = stamp ? [] : row.querySelectorAll('span, div');
		for (var i = 0; i < spans.length && out.time === null; i++) {
			var el = spans[i];
			if (el.children.length) continue;
			var t = (el.textContent || '').trim();
			if (t && t.length <= 12 && t !== name) out.time = parseRowTime(t, Date.now());
		}
		return out;
	}

	// markIncoming records a message from the other side at time t. The wait counts from the first
	// unanswered message; exact is false when t is only "now" because the row showed no time yet, and
	// such a guess gives way to the row's time. A newer message brings back a chat the user ignored.
	function markIncoming(chat, t, exact, extra) {
		var s = state[chat];
		if (!s) {
			state[chat] = Object.assign({ since: t, last: t, exact: exact, seen: Date.now(), notified: false, ignored: false }, extra || {});
			return;
		}
		if (exact && !s.exact) {
			s.since = t;
			s.exact = true;
		}
		if (t > (s.last || s.since) + 60000) {
			s.last = t;
			s.ignored = false;
		}
		if (extra) Object.assign(s, extra);
	}
	function markAnswered(chat) { if (state[chat]) delete state[chat]; }

	function scanList() {
		if (!window.waSel) return;
		window.waSel.all('chatRows').forEach(function(row) {
			var info = rowInfo(row);
			if (!info) return;
			if (info.mine) { markAnswered(info.name); return; }
			var extra = { muted: info.muted };
			if (info.group) extra.group = true;
			markIncoming(info.name, info.time || Date.now(), info.time !== null, extra);
		});
	}

	function scanOpenChat() {
		var main = document.getElementById('main');
		var chat = main && window.waOpenChatName ? window.waOpenChatName() : '';
		if (!chat) return;
		var rows = main.querySelectorAll('[data-id]');
		var last = null;
		for (var i = rows.length - 1; i >= 0; i--) {
			if (rows[i].querySelector('[data-testid="msg-container"]') || /^(true|false)_/.test(rows[i].getAttribute('data-id') || '')) { last = rows[i]; break; }
		}
		if (!last) return;
		var id = last.getAttribute('data-id') || '';
		var subtitle = main.querySelector('header [data-testid="chat-subtitle"]');
		var group = /@g\.us/.test(id) || !!(subtitle && /,/.test(subtitle.textContent || ''));
		if (id.indexOf('true_') === 0 || hasTick(last)) { markAnswered(chat); return; }
		if (state[chat]) state[chat].group = group;
		else markIncoming(chat, Date.now(), false, { group: group });
	}

	window.waOnNotification = function(title) {
		title = String(title || '').trim();
		if (!title || title === 'Radar Online') return;
		markIncoming(title, Date.now(), true);
		save();
	};

	// waiting lists the chats past the limit, longest wait first.
	function waiting(now) {
		var c = cfg();
		if (!c.enabled) return [];
		var limit = c.minutes * 60000;
		return Object.keys(state).map(function(chat) { return Object.assign({ chat: chat }, state[chat]); })
			.filter(function(e) {
				return !e.ignored && !e.muted && (c.includeGroups || !e.group) &&
					now - e.since >= limit && now - e.since <= SHOW_DAYS * 86400000;
			})
			.sort(function(a, b) { return a.since - b.since; });
	}
	window.waFollowUpWaiting = function() { return waiting(Date.now()); };

	function formatWait(ms) {
		var min = Math.floor(ms / 60000);
		if (min < 60) return min + ' mnt';
		var h = Math.floor(min / 60);
		if (h < 24) return h + ' jam';
		return Math.floor(h / 24) + ' hari';
	}

	function ignore(chat) {
		if (state[chat]) { state[chat].ignored = true; save(); }
		refresh();
	}

	// Dock button with the count, and a popup with the list
	function ensureDockButton(count) {
		var dock = document.getElementById('wa-messenger-tabs');
		if (!dock) return;
		var btn = document.getElementById('wa-tab-followup');
		if (!btn) {
			var sep = document.createElement('div');
			sep.className = 'wa-dock-sep';
			btn = document.createElement('button');
			btn.id = 'wa-tab-followup';
			btn.className = 'wa-dock-item';
			btn.title = 'Chat yang belum dibalas';
			btn.innerHTML = '<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg><span>Balas</span><b class="wa-followup-count" style="min-width:16px;padding:0 4px;border-radius:8px;background:#ef4444;color:#fff;font-size:10px;line-height:16px;text-align:center;"></b>';
			btn.onclick = function(e) { e.preventDefault(); e.stopPropagation(); showListPopup(btn); };
			dock.appendChild(sep);
			dock.appendChild(btn);
		}
		var badge = btn.querySelector('.wa-followup-count');
		badge.textContent = count > 99 ? '99+' : String(count);
		badge.style.display = count ? '' : 'none';
	}

	function showListPopup(anchor) {
		var old = document.getElementById('wa-followup-pop');
		if (old) { old.remove(); return; }
		var list = waiting(Date.now());
		var pop = document.createElement('div');
		pop.id = 'wa-followup-pop';
		pop.className = 'wa-modal-box';
		pop.style.cssText = 'position:fixed;z-index:2147483600;width:300px;max-height:60vh;overflow:auto;padding:12px;';
		var r = anchor.getBoundingClientRect();
		pop.style.left = Math.max(8, r.left) + 'px';
		pop.style.bottom = (innerHeight - r.top + 8) + 'px';
		var head = document.createElement('div');
		head.style.cssText = 'font-size:13px;font-weight:600;color:var(--wa-text);margin:0 4px 8px;';
		head.textContent = list.length ? list.length + ' chat belum dibalas' : 'Semua chat sudah dibalas';
		pop.appendChild(head);
		list.forEach(function(e) { pop.appendChild(listRow(e, function() { pop.remove(); })); });
		document.body.appendChild(pop);
		setTimeout(function() {
			document.addEventListener('mousedown', function close(ev) {
				if (!pop.contains(ev.target) && ev.target !== anchor && !anchor.contains(ev.target)) { pop.remove(); document.removeEventListener('mousedown', close, true); }
			}, true);
		}, 0);
	}

	function listRow(e, after) {
		var row = document.createElement('div');
		row.className = 'wa-row';
		row.style.cssText = 'gap:8px;';
		row.innerHTML = '<div class="wa-row-text" style="min-width:0;"><div class="wa-row-title" style="overflow:hidden;text-overflow:ellipsis;white-space:nowrap;"></div><div class="wa-row-desc"></div></div>' +
			'<div class="wa-row-actions"><button class="wa-btn wa-btn-primary wa-btn-sm">Buka</button><button class="wa-btn wa-btn-secondary wa-btn-sm">Abaikan</button></div>';
		row.querySelector('.wa-row-title').textContent = e.chat;
		row.querySelector('.wa-row-desc').textContent = 'Menunggu ' + formatWait(Date.now() - e.since);
		var btns = row.querySelectorAll('button');
		btns[0].onclick = function() { if (after) after(); if (window.waOpenChatByName) window.waOpenChatByName(e.chat); };
		btns[1].onclick = function() { ignore(e.chat); row.remove(); toast('Diabaikan sampai ada pesan baru: ' + e.chat); };
		return row;
	}

	// Control Center (Chat & alat tab)
	window.waRenderFollowUp = function() {
		var c = cfg();
		var en = document.getElementById('wa-cc-followup-cb');
		var mins = document.getElementById('wa-cc-followup-min');
		var notif = document.getElementById('wa-cc-followup-notify');
		var groups = document.getElementById('wa-cc-followup-groups');
		var list = document.getElementById('wa-followup-list');
		var count = document.getElementById('wa-followup-count');
		function set(patch) { saveJSON(CFG_KEY, Object.assign(cfg(), patch)); refresh(); window.waRenderFollowUp(); }
		if (en) { en.checked = c.enabled; en.onchange = function() { set({ enabled: en.checked }); }; }
		if (mins) { mins.value = String(c.minutes); mins.onchange = function() { set({ minutes: parseInt(mins.value, 10) || 60 }); }; }
		if (notif) { notif.checked = c.notify; notif.onchange = function() { set({ notify: notif.checked }); }; }
		if (groups) { groups.checked = c.includeGroups; groups.onchange = function() { set({ includeGroups: groups.checked }); }; }
		var items = waiting(Date.now());
		if (count) count.textContent = String(items.length);
		if (list) { list.innerHTML = ''; items.forEach(function(e) { list.appendChild(listRow(e)); }); }
	};

	function refresh() {
		var now = Date.now();
		var list = waiting(now);
		ensureDockButton(list.length);
		var c = cfg();
		// Notify only chats that passed the limit while the app was watching: no toasts for the backlog
		// found on the first scan or when older chats scroll into view.
		var limit = c.minutes * 60000;
		var due = list.filter(function(e) { return !e.notified && e.since + limit >= (e.seen || 0) - 60000; });
		list.forEach(function(e) { state[e.chat].notified = true; });
		if (c.notify && due.length && typeof window.notifyFollowUp === 'function') {
			if (due.length <= 2) {
				due.forEach(function(e) {
					window.notifyFollowUp(e.chat, 'Menunggu balasan ' + formatWait(now - e.since) + '. Klik untuk membuka chat.');
				});
			} else {
				var names = due.slice(0, 3).map(function(e) { return e.chat; }).join(', ') + (due.length > 3 ? ', +' + (due.length - 3) + ' lainnya' : '');
				window.notifyFollowUp('', due.length + ' chat belum dibalas: ' + names);
			}
		}
		// Forget old waits; they are not coming back as a to-do.
		Object.keys(state).forEach(function(k) { if (now - state[k].since > FORGET_DAYS * 86400000) delete state[k]; });
		save();
	}

	window.waFollowUpRefresh = refresh; // for tests

	var timer = null;
	function schedule() {
		if (timer) return;
		timer = setTimeout(function() {
			timer = null;
			try { scanList(); scanOpenChat(); refresh(); } catch (e) {}
		}, 1500);
	}

	function boot() {
		new MutationObserver(schedule).observe(document.body, { childList: true, subtree: true });
		setInterval(schedule, 60000);
		schedule();
	}
	if (document.body) boot(); else document.addEventListener('DOMContentLoaded', boot, { once: true });
})();
