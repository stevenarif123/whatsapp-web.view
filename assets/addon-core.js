// Add-on core: selector registry + diagnostics, unlimited local pins, online radar.
// Appended to the WhatsApp init script by main.go. Relies on globals defined there:
// whenDOMReady(), showAddonToast(), privacyConfig.
(function () {
	if (window.self !== window.top) return;

	// ------------------------------------------------------------------
	// 1. Selector registry. WhatsApp Web changes its DOM often, so every
	//    element we depend on lists several candidate selectors, tried in
	//    order. The Diagnostik tab reports which candidate still matches.
	//    need: 'app' = always present, 'chat' = needs an open chat.
	// ------------------------------------------------------------------
	var WA_SEL = {
		paneSide: { label: 'Panel daftar chat', need: 'app', sel: ['#pane-side'] },
		chatRows: { label: 'Baris chat', need: 'app', sel: [
			'#pane-side [role="row"][data-testid^="list-item-"]',
			'#pane-side [role="row"]',
			'#pane-side [role="listitem"]',
			'#pane-side [data-testid="cell-frame-container"]'
		] },
		rowTitle: { label: 'Nama di baris chat', need: 'app', sel: [
			'[data-testid="cell-frame-title"] span[title]',
			'span[title]',
			'span[dir="auto"]'
		], within: 'chatRows' },
		unreadBadge: { label: 'Badge belum dibaca', need: 'unread', sel: [
			'#pane-side [data-testid="icon-unread-count"]',
			'#pane-side [aria-label*="belum dibaca"], #pane-side [aria-label*="unread"]'
		] },
		chatHeader: { label: 'Header chat terbuka', need: 'chat', sel: [
			'#main header',
			'header[data-testid="conversation-header"]',
			'div[role="main"] header'
		] },
		headerTitle: { label: 'Nama kontak di header', need: 'chat', sel: [
			'[data-testid="conversation-info-header-chat-title"]',
			'span[dir="auto"][title]',
			'span[dir="auto"]'
		], within: 'chatHeader' },
		messages: { label: 'Gelembung pesan', need: 'chat', sel: [
			'#main [data-testid="msg-container"]',
			'#main [data-id]',
			'#main .message-in, #main .message-out'
		] },
		composer: { label: 'Kotak ketik pesan', need: 'chat', sel: [
			'footer div[contenteditable="true"]',
			'#main div[contenteditable="true"]',
			'div[contenteditable="true"][data-lexical-editor="true"]'
		] },
		searchBox: { label: 'Kotak pencarian chat', need: 'app', sel: [
			'[data-testid="chat-list-search-container"] input',
			'input[role="textbox"][data-tab="3"]',
			'#side div[contenteditable="true"][role="textbox"]',
			'#side input[type="text"]'
		] },
		navRail: { label: 'Rel navigasi kiri', need: 'app', sel: [
			'[data-testid="navbar-primary-section"]',
			'[role="navigation"]',
			'nav'
		] },
		navFooter: { label: 'Bagian bawah rel navigasi', need: 'app', sel: [
			'[data-testid="navbar-footer-section"]',
			'[data-testid="navbar-primary-section"]'
		] },
		mainScreen: { label: 'Wadah tata letak utama', need: 'app', sel: [
			'[data-testid="wa-web-main-screen"]'
		] },
		statusViewer: { label: 'Penampil status', need: 'status', sel: [
			'div[data-animate-status-v3="true"]',
			'[data-testid="status-v3-main"]',
			'div[role="dialog"]'
		] }
	};

	function qAll(selector, root) {
		try { return Array.prototype.slice.call((root || document).querySelectorAll(selector)); } catch (e) { return []; }
	}

	// Returns { els, index, selector } for the first candidate that matches.
	function resolve(key, root) {
		var def = WA_SEL[key];
		if (!def) return { els: [], index: -1, selector: '' };
		if (def.within && !root) {
			var parent = resolve(def.within);
			root = parent.els[0] || null;
			if (!root) return { els: [], index: -1, selector: '' };
		}
		for (var i = 0; i < def.sel.length; i++) {
			var els = qAll(def.sel[i], root);
			if (els.length) return { els: els, index: i, selector: def.sel[i] };
		}
		return { els: [], index: -1, selector: '' };
	}
	function one(key, root) { return resolve(key, root).els[0] || null; }
	function all(key, root) { return resolve(key, root).els; }

	window.WA_SEL = WA_SEL;
	window.waSel = { one: one, all: all, resolve: resolve };

	function toast(msg) { if (typeof showAddonToast === 'function') showAddonToast(msg); }
	function loadJSON(k, def) {
		try { var v = localStorage.getItem(k); return v ? JSON.parse(v) : def; } catch (e) { return def; }
	}
	function saveJSON(k, v) { try { localStorage.setItem(k, JSON.stringify(v)); } catch (e) {} }
	function esc(s) { return String(s).replace(/[&<>"]/g, function (c) { return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' }[c]; }); }

	// ------------------------------------------------------------------
	// 2. Diagnostics
	// ------------------------------------------------------------------
	function chatOpen() { return !!document.querySelector('#main'); }
	function statusOpen() { return !!document.querySelector('div[data-animate-status-v3="true"], [data-testid="status-v3-main"]'); }

	window.waRunDiagnostics = function () {
		var rows = [];
		Object.keys(WA_SEL).forEach(function (key) {
			var def = WA_SEL[key];
			var r = resolve(key);
			var state, note;
			if (r.els.length) {
				state = r.index === 0 ? 'ok' : 'warn';
				note = r.index === 0 ? r.selector : 'fallback: ' + r.selector;
			} else if (def.need === 'chat' && !chatOpen()) {
				state = 'warn'; note = 'buka sebuah chat untuk memeriksa';
			} else if (def.need === 'unread' && !document.querySelector('#pane-side [role="row"]')) {
				state = 'warn'; note = 'daftar chat belum termuat';
			} else if (def.need === 'unread') {
				state = 'warn'; note = 'tidak ada chat belum dibaca saat ini';
			} else if (def.need === 'status' && !statusOpen()) {
				state = 'warn'; note = 'buka sebuah status untuk memeriksa';
			} else {
				state = 'bad'; note = 'tidak ditemukan - fitur terkait mungkin tidak jalan';
			}
			rows.push({ name: def.label, state: state, note: note, count: r.els.length });
		});
		var bridge = [
			['Notifikasi native', typeof window.sendNativeNotification === 'function'],
			['Buka link eksternal', typeof window.openExternalLink === 'function'],
			['Autostart', typeof window.getAutoStartStatus === 'function']
		];
		bridge.forEach(function (b) {
			rows.push({ name: 'Jembatan Go: ' + b[0], state: b[1] ? 'ok' : 'bad', note: b[1] ? 'terhubung' : 'tidak tersedia', count: b[1] ? 1 : 0 });
		});
		return rows;
	};

	function renderDiagnostics(modal) {
		var list = modal.querySelector('#wa-diag-list');
		var summary = modal.querySelector('#wa-diag-summary');
		if (!list) return;
		var rows = window.waRunDiagnostics();
		var bad = rows.filter(function (r) { return r.state === 'bad'; }).length;
		var warn = rows.filter(function (r) { return r.state === 'warn'; }).length;
		list.innerHTML = rows.map(function (r) {
			return '<div class="wa-diag-row" data-state="' + r.state + '"><span class="wa-diag-dot"></span>' +
				'<div style="min-width:0"><div class="wa-diag-name">' + esc(r.name) + '</div><div class="wa-diag-note" title="' + esc(r.note) + '">' + esc(r.note) + '</div></div>' +
				'<span class="wa-diag-count">' + r.count + '</span></div>';
		}).join('');
		if (summary) {
			summary.textContent = bad ? (bad + ' elemen tidak ditemukan' + (warn ? ', ' + warn + ' perlu perhatian' : '')) :
				(warn ? 'Semua oke, ' + warn + ' butuh konteks (buka chat/status)' : 'Semua elemen dikenali');
		}
	}

	// ------------------------------------------------------------------
	// 3. Unlimited pinned chats (local only; WhatsApp's own pins untouched)
	// ------------------------------------------------------------------
	var pinState = { enabled: loadJSON('wa_pins_enabled', true), names: loadJSON('wa_pinned_chats', []) };
	var PIN_SVG = '<svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"><path d="M16 3l5 5-3.2 1.1-3.6 3.6.5 4.4-1.7 1.7-3.4-3.4L5 21l-1-1 4.6-5.6L5.2 11l1.7-1.7 4.4.5 3.6-3.6z"/></svg>';

	function savePins() { saveJSON('wa_pinned_chats', pinState.names); saveJSON('wa_pins_enabled', pinState.enabled); }
	function isPinned(name) { return pinState.names.indexOf(name) !== -1; }
	function rowName(row) {
		var t = one('rowTitle', row);
		if (!t) return '';
		return (t.getAttribute('title') || t.textContent || '').trim();
	}

	function fireClick(el) {
		['mousedown', 'mouseup', 'click'].forEach(function (type) {
			el.dispatchEvent(new MouseEvent(type, { bubbles: true, cancelable: true, view: window }));
		});
	}
	// WhatsApp's click handler sits on the row's inner frame, not the row itself.
	function openRow(row) {
		fireClick(row.querySelector('[data-testid="cell-frame-container"]') || row.querySelector('[role="gridcell"]') || row);
	}

	function findRowByName(name) {
		var rows = all('chatRows');
		for (var i = 0; i < rows.length; i++) { if (rowName(rows[i]) === name) return rows[i]; }
		return null;
	}

	function openChatByName(name) {
		var row = findRowByName(name);
		if (row) { openRow(row); return; }
		// Not rendered (virtualised list): search for it, then open the match.
		var box = one('searchBox');
		if (!box) { toast('Kotak pencarian tidak ditemukan. Cek tab Diagnostik.'); return; }
		box.focus();
		if (box.isContentEditable) {
			document.execCommand('selectAll', false, null);
			document.execCommand('insertText', false, name);
		} else {
			var setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value').set;
			setter.call(box, name);
			box.dispatchEvent(new Event('input', { bubbles: true }));
		}
		var tries = 0;
		var iv = setInterval(function () {
			tries++;
			var hit = findRowByName(name);
			if (hit || tries > 12) {
				clearInterval(iv);
				if (hit) { openRow(hit); }
				else { toast('Chat "' + name + '" tidak ditemukan'); }
				setTimeout(function () {
					box.focus();
					document.execCommand('selectAll', false, null);
					document.execCommand('delete', false, null);
					box.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', code: 'Escape', bubbles: true }));
				}, 250);
			}
		}, 150);
	}

	function ensurePinBar() {
		var pane = one('paneSide');
		if (!pane) return null;
		var bar = document.getElementById('wa-pin-bar');
		if (!bar) {
			bar = document.createElement('div');
			bar.id = 'wa-pin-bar';
			bar.style.cssText = 'position:sticky;top:0;z-index:20;';
		}
		if (bar.parentElement !== pane) pane.insertBefore(bar, pane.firstChild);
		return bar;
	}

	function renderPinBar() {
		var bar = ensurePinBar();
		if (!bar) return;
		var show = pinState.enabled && pinState.names.length > 0;
		bar.classList.toggle('has-pins', show);
		var key = show ? pinState.names.join('\u0001') : '';
		if (bar._key === key) return;
		bar._key = key;
		bar.innerHTML = '';
		if (!show) return;
		pinState.names.forEach(function (name) {
			var chip = document.createElement('button');
			chip.className = 'wa-pin-chip';
			chip.title = name;
			chip.innerHTML = PIN_SVG + '<span></span>';
			chip.lastChild.textContent = name;
			chip.onclick = function () { openChatByName(name); };
			bar.appendChild(chip);
		});
	}

	// Per-chat blur: chats listed here stay blurred (hover to peek) without needing Privacy Mode.
	var blurState = { names: loadJSON('wa_blur_chats', []) };
	var EYE_SVG = '<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/></svg>';
	function isBlurred(name) { return blurState.names.indexOf(name) !== -1; }
	function saveBlur() { saveJSON('wa_blur_chats', blurState.names); }

	function ensureRowButton(row, cls, html) {
		var btn = row.querySelector(':scope > .' + cls);
		if (btn) return btn;
		btn = document.createElement('button');
		btn.className = 'wa-pin-toggle ' + cls;
		btn.innerHTML = html;
		row.appendChild(btn);
		return btn;
	}

	function decorateRows() {
		var rows = all('chatRows');
		rows.forEach(function (row) {
			var name = rowName(row);
			if (!name) return;
			if (getComputedStyle(row).position === 'static') row.style.position = 'relative';
			row.classList.add('wa-pin-host');

			// Blur toggle (always available)
			var eye = ensureRowButton(row, 'wa-blur-toggle', EYE_SVG);
			if (!eye._wired) {
				eye._wired = true;
				eye.addEventListener('click', function (e) {
					e.preventDefault(); e.stopPropagation();
					var n = eye._name;
					if (isBlurred(n)) { blurState.names = blurState.names.filter(function (x) { return x !== n; }); toast('Blur dilepas: ' + n); }
					else { blurState.names.push(n); toast('Chat diburamkan: ' + n); }
					saveBlur(); refreshPins();
				}, true);
			}
			eye._name = name;
			eye.title = isBlurred(name) ? 'Lepas blur chat ini' : 'Selalu buramkan chat ini';
			eye.classList.toggle('on', isBlurred(name));
			row.setAttribute('data-wa-blur', isBlurred(name) ? '1' : '0');

			// Pin toggle
			var btn = row.querySelector(':scope > .wa-pin-btn');
			if (!pinState.enabled) { if (btn) btn.remove(); return; }
			if (!btn) {
				btn = ensureRowButton(row, 'wa-pin-btn', PIN_SVG);
				btn.addEventListener('click', function (e) {
					e.preventDefault(); e.stopPropagation();
					var n = btn._name;
					if (isPinned(n)) { pinState.names = pinState.names.filter(function (x) { return x !== n; }); toast('Pin dilepas: ' + n); }
					else { pinState.names.push(n); toast('Dipin: ' + n); }
					savePins(); refreshPins();
				}, true);
			}
			btn._name = name;
			btn.title = isPinned(name) ? 'Lepas pin' : 'Pin chat ini';
			btn.classList.toggle('pinned', isPinned(name));
		});
	}

	function openChatName() {
		var h = one('chatHeader');
		var t = h && one('headerTitle', h);
		return t ? (t.getAttribute('title') || t.textContent || '').trim() : '';
	}

	function renderBlurList() {
		var list = document.getElementById('wa-blurchat-list');
		var c = document.getElementById('wa-blurchat-count');
		if (c) c.textContent = String(blurState.names.length);
		if (!list) return;
		list.innerHTML = '';
		blurState.names.forEach(function (name) {
			var row = document.createElement('div');
			row.className = 'wa-row';
			row.innerHTML = '<div class="wa-row-title" style="overflow:hidden;text-overflow:ellipsis;white-space:nowrap;"></div><button class="wa-btn wa-btn-secondary wa-btn-sm">Lepas</button>';
			row.firstChild.textContent = name;
			row.lastChild.onclick = function () {
				blurState.names = blurState.names.filter(function (x) { return x !== name; });
				saveBlur(); refreshPins();
			};
			list.appendChild(row);
		});
	}

	function refreshPins() {
		renderPinBar();
		decorateRows();
		var c = document.getElementById('wa-pin-count');
		if (c) c.textContent = String(pinState.names.length);
		renderPinList();
	}

	function renderPinList() {
		var list = document.getElementById('wa-pin-list');
		if (!list) return;
		list.innerHTML = '';
		pinState.names.forEach(function (name) {
			var row = document.createElement('div');
			row.className = 'wa-row';
			row.innerHTML = '<div class="wa-row-title" style="overflow:hidden;text-overflow:ellipsis;white-space:nowrap;"></div><button class="wa-btn wa-btn-secondary wa-btn-sm">Lepas</button>';
			row.firstChild.textContent = name;
			row.lastChild.onclick = function () {
				pinState.names = pinState.names.filter(function (x) { return x !== name; });
				savePins(); refreshPins();
			};
			list.appendChild(row);
		});
	}

	var pinTimer = null;
	function schedulePins() {
		if (pinTimer) return;
		pinTimer = setTimeout(function () { pinTimer = null; try { refreshPins(); } catch (e) {} }, 300);
	}

	// ------------------------------------------------------------------
	// 4. Online radar (only the chat currently open exposes presence)
	// ------------------------------------------------------------------
	var radar = loadJSON('wa_radar_config', { enabled: false, names: '' });
	var radarSeen = {};
	var RE_TYPING = /(sedang mengetik|mengetik|typing|merekam|recording)/i;
	var RE_ONLINE = /^(online|daring)$/i;

	function radarWanted(name) {
		var list = String(radar.names || '').split(',').map(function (s) { return s.trim().toLowerCase(); }).filter(Boolean);
		if (!list.length) return true;
		var n = name.toLowerCase();
		return list.some(function (x) { return n.indexOf(x) !== -1; });
	}

	function radarTick() {
		if (!radar.enabled) return;
		var header = one('chatHeader');
		if (!header) return;
		var titleEl = one('headerTitle', header);
		var name = titleEl ? (titleEl.getAttribute('title') || titleEl.textContent || '').trim() : '';
		if (!name) return;
		var texts = qAll('span', header).map(function (s) { return (s.getAttribute('title') || s.textContent || '').trim(); });
		var online = texts.some(function (t) { return RE_ONLINE.test(t); });
		var typing = texts.some(function (t) { return t.length < 40 && RE_TYPING.test(t); });
		var prev = radarSeen[name];
		radarSeen[name] = { online: online, typing: typing };
		if (!prev || !radarWanted(name)) return;
		var msg = null;
		if (typing && !prev.typing) msg = '✍️ ' + name + ' sedang mengetik';
		else if (online && !prev.online) msg = '🟢 ' + name + ' sedang online';
		if (msg) {
			toast(msg);
			if (typeof window.sendNativeNotification === 'function') window.sendNativeNotification('Radar Online', msg);
		}
	}

	window.waRadar = { tick: radarTick, config: radar };

	// ------------------------------------------------------------------
	// 4a. Compact layout for narrow windows. WhatsApp Web's main container has min-width: 748px, so in a
	//     narrower window the chat pane is cut off and the page scrolls sideways. Below NARROW_PX we show
	//     one pane at a time (chat list, or the open chat with a back button), like the native app.
	// ------------------------------------------------------------------
	var NARROW_PX = 900;
	var compactEnabled = loadJSON('wa_compact_narrow', true);

	function markLayout() {
		var side = document.getElementById('side');
		var screen = one('mainScreen');
		if (!side || !screen) return;
		var left = side.parentElement;
		var root = left && left.parentElement;
		if (!root || root.parentElement !== screen) return;
		if (root.getAttribute('data-wa-layout') !== 'root') root.setAttribute('data-wa-layout', 'root');
		if (left.getAttribute('data-wa-pane') !== 'list') left.setAttribute('data-wa-pane', 'list');
		// The root also holds the nav rail, overlays and toasts; only the pane right after the list is the chat.
		var chat = left.nextElementSibling;
		if (chat && chat.getAttribute('data-wa-pane') !== 'chat') chat.setAttribute('data-wa-pane', 'chat');
	}

	function ensureBackButton() {
		var b = document.getElementById('wa-narrow-back');
		if (b || !document.body) return b;
		b = document.createElement('button');
		b.id = 'wa-narrow-back';
		b.title = 'Kembali ke daftar chat';
		b.setAttribute('aria-label', 'Kembali ke daftar chat');
		b.innerHTML = '<svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 18 9 12 15 6"/></svg>';
		b.addEventListener('click', function () { window.waCloseChat(); });
		document.body.appendChild(b);
		return b;
	}

	// Escape closes the open conversation in WhatsApp Web.
	window.waCloseChat = function () {
		var target = document.querySelector('#main footer [contenteditable="true"]') || document.querySelector('#main') || document.body;
		['keydown', 'keyup'].forEach(function (type) {
			target.dispatchEvent(new KeyboardEvent(type, { key: 'Escape', code: 'Escape', keyCode: 27, which: 27, bubbles: true, cancelable: true }));
		});
	};

	function syncNarrow() {
		if (!document.body) return;
		var narrow = !!compactEnabled && window.innerWidth < NARROW_PX;
		var root = document.documentElement;
		if (root.hasAttribute('data-wa-narrow') !== narrow) {
			if (narrow) root.setAttribute('data-wa-narrow', ''); else root.removeAttribute('data-wa-narrow');
		}
		var open = document.getElementById('main') ? '1' : '0';
		if (document.body.getAttribute('data-wa-chat-open') !== open) document.body.setAttribute('data-wa-chat-open', open);
		if (narrow) { markLayout(); ensureBackButton(); }
	}
	window.waSyncNarrow = syncNarrow;

	// ------------------------------------------------------------------
	// 4b. Unread filter. WhatsApp ships its own "Belum dibaca" chip, so we
	//     drive that. Rows are positioned with transform, so hiding them
	//     would leave gaps; the fallback dims non-unread rows instead.
	// ------------------------------------------------------------------
	function findFilterTab(re) {
		var tabs = qAll('[role="tablist"][aria-label="chat-list-filters"] [role="tab"]');
		if (!tabs.length) tabs = qAll('#side [role="tab"]');
		for (var i = 0; i < tabs.length; i++) {
			if (re.test((tabs[i].textContent || '').trim())) return tabs[i];
		}
		return null;
	}
	function unreadActive() {
		var chip = findFilterTab(/^(belum dibaca|unread)/i);
		if (chip) return chip.getAttribute('aria-selected') === 'true';
		return !!document.getElementById('wa-unread-dim-style') && document.getElementById('wa-unread-dim-style').textContent !== '';
	}
	function setUnreadButtons(on) {
		/* eslint-disable-next-line no-undef */
		try { isUnreadFilterActive = on; } catch (e) {}
		var ccBtn = document.getElementById('wa-cc-unread-toggle');
		if (ccBtn) {
			ccBtn.className = on ? 'wa-btn wa-btn-primary' : 'wa-btn wa-btn-secondary';
			ccBtn.textContent = on ? 'Nonaktifkan Filter' : 'Aktifkan Filter';
		}
	}
	window.toggleUnreadFilter = function (force) {
		var on = typeof force === 'boolean' ? force : !unreadActive();
		var chip = findFilterTab(/^(belum dibaca|unread)/i);
		var allTab = findFilterTab(/^(semua|all)/i);
		var dim = document.getElementById('wa-unread-dim-style');
		if (chip && allTab) {
			fireClick(on ? chip : allTab);
			if (dim) dim.textContent = '';
		} else {
			if (!dim) {
				dim = document.createElement('style');
				dim.id = 'wa-unread-dim-style';
				(document.head || document.documentElement).appendChild(dim);
			}
			dim.textContent = on ? '#pane-side [role="row"]:not(:has([data-testid="icon-unread-count"])) { opacity: 0.2; }' : '';
		}
		setUnreadButtons(on);
		toast(on ? 'Filter belum dibaca: aktif' : 'Filter belum dibaca: nonaktif');
		return on;
	};

	// ------------------------------------------------------------------
	// 4c. Quick reply variables, lock-on-hide, chat export
	// ------------------------------------------------------------------
	window.waExpandVars = function (text) {
		var name = openChatName();
		var d = new Date();
		var map = {
			nama: name,
			depan: name ? name.split(/\s+/)[0] : '',
			tanggal: d.toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' }),
			hari: d.toLocaleDateString('id-ID', { weekday: 'long' }),
			jam: d.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' }).replace('.', ':')
		};
		return String(text).replace(/\{(nama|depan|tanggal|hari|jam)\}/gi, function (m, k) {
			var v = map[k.toLowerCase()];
			return v ? v : m;
		});
	};

	// Called from Go when the window is hidden to the tray or via the Boss Key.
	window.waOnHidden = function () {
		if (loadJSON('wa_lock_on_hide', false) && typeof window.lockWhatsApp === 'function') window.lockWhatsApp();
	};

	// Builds a plain-text transcript of the messages currently rendered in the open chat.
	window.waBuildChatExport = function () {
		var rows = qAll('#main [data-id][data-testid^="conv-msg-"]');
		if (!rows.length) rows = all('messages');
		var lines = [];
		rows.forEach(function (row) {
			var pre = row.querySelector('[data-pre-plain-text]');
			var head = pre ? pre.getAttribute('data-pre-plain-text').trim() : '';
			var parts = qAll('[data-testid$="selectable-text"], .selectable-text.copyable-text', row).filter(function (el) {
				return !el.closest('[data-testid="quoted-message"]');
			});
			// Keep only outermost matches so nested spans are not counted twice.
			parts = parts.filter(function (el) { return !parts.some(function (o) { return o !== el && o.contains(el); }); });
			var body = parts.map(function (el) { return (el.innerText || '').trim(); }).filter(Boolean).join('\n');
			if (!body) {
				if (row.querySelector('[data-testid="image-thumb"]')) body = '<gambar>';
				else if (row.querySelector('[data-testid="video-content"], video')) body = '<video>';
				else if (row.querySelector('audio, [data-icon="audio-play"], [data-testid="audio-play"]')) body = '<pesan suara>';
				else body = '<lampiran / stiker>';
			}
			lines.push((head ? head + ' ' : '') + body);
		});
		return lines;
	};

	window.waExportChat = function () {
		var lines = window.waBuildChatExport();
		if (!lines.length) { toast('Buka sebuah chat dulu'); return 0; }
		var title = openChatName() || 'chat';
		var text = 'Ekspor chat WhatsApp: ' + title + '\r\n' + 'Diekspor: ' + new Date().toLocaleString('id-ID') + '\r\n' +
			'Hanya pesan yang termuat di layar (' + lines.length + ' pesan)\r\n\r\n' + lines.join('\r\n') + '\r\n';
		var blob = new Blob(['\uFEFF' + text], { type: 'text/plain;charset=utf-8' });
		var url = URL.createObjectURL(blob);
		var a = document.createElement('a');
		a.href = url;
		a.download = ('WhatsApp - ' + title).replace(/[\\/:*?"<>|]/g, '_').slice(0, 80) + '.txt';
		document.body.appendChild(a);
		a.click();
		setTimeout(function () { a.remove(); URL.revokeObjectURL(url); }, 2000);
		toast('Mengekspor ' + lines.length + ' pesan');
		return lines.length;
	};

	// ------------------------------------------------------------------
	// 5. Control Center wiring (called from main.go after the modal exists)
	// ------------------------------------------------------------------
	window.waBindCoreUI = function (modal) {
		var pinCb = modal.querySelector('#wa-cc-pin-cb');
		var radarCb = modal.querySelector('#wa-cc-radar-cb');
		var radarNames = modal.querySelector('#wa-cc-radar-names');
		var diagBtn = modal.querySelector('#wa-diag-run');
		var exportBtn = modal.querySelector('#wa-cc-export-btn');
		var lockHideCb = modal.querySelector('#wa-cc-lock-hide-cb');
		var compactCb = modal.querySelector('#wa-cc-compact-narrow-cb');

		if (pinCb) pinCb.onchange = function () {
			pinState.enabled = pinCb.checked; savePins(); refreshPins();
			toast(pinState.enabled ? 'Pin chat: aktif' : 'Pin chat: nonaktif');
		};
		if (radarCb) radarCb.onchange = function () {
			radar.enabled = radarCb.checked; saveJSON('wa_radar_config', radar); radarSeen = {};
			toast(radar.enabled ? 'Radar online: aktif' : 'Radar online: nonaktif');
		};
		if (radarNames) radarNames.onchange = function () { radar.names = radarNames.value; saveJSON('wa_radar_config', radar); };
		if (diagBtn) diagBtn.onclick = function () { renderDiagnostics(modal); };
		if (exportBtn) exportBtn.onclick = function () { window.waExportChat(); };
		if (compactCb) compactCb.onchange = function () {
			compactEnabled = compactCb.checked;
			saveJSON('wa_compact_narrow', compactEnabled);
			syncNarrow();
			toast(compactEnabled ? 'Mode ringkas: aktif' : 'Mode ringkas: nonaktif');
		};
		if (lockHideCb) lockHideCb.onchange = function () {
			saveJSON('wa_lock_on_hide', lockHideCb.checked);
			toast(lockHideCb.checked ? 'Kunci saat disembunyikan: aktif' : 'Kunci saat disembunyikan: nonaktif');
		};

		var prevSync = modal._syncUI;
		modal._syncUI = function () {
			// Older saved configs may hold a blur size that is not one of the presets.
			try {
				var sizes = [3, 5, 8], cur = privacyConfig.blurIntensity || 5;
				var best = sizes.reduce(function (a, b) { return Math.abs(b - cur) < Math.abs(a - cur) ? b : a; });
				if (best !== privacyConfig.blurIntensity) {
					privacyConfig.blurIntensity = best;
					if (typeof savePrivacyConfig === 'function') savePrivacyConfig();
				}
			} catch (e) {}
			setUnreadButtons(unreadActive());
			if (prevSync) prevSync();
			if (pinCb) pinCb.checked = !!pinState.enabled;
			if (radarCb) radarCb.checked = !!radar.enabled;
			if (radarNames) radarNames.value = radar.names || '';
			if (lockHideCb) lockHideCb.checked = !!loadJSON('wa_lock_on_hide', false);
			if (compactCb) compactCb.checked = !!compactEnabled;
			refreshPins();
		};
	};

	// ------------------------------------------------------------------
	// 6. Boot
	// ------------------------------------------------------------------
	whenDOMReady(function () {
		new MutationObserver(function () { syncNarrow(); schedulePins(); }).observe(document.body, { childList: true, subtree: true });
		window.addEventListener('resize', syncNarrow);
		syncNarrow();
		schedulePins();
		setInterval(radarTick, 1500);
	});
})();
