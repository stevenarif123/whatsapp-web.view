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

	function decorateRows() {
		var rows = all('chatRows');
		rows.forEach(function (row) {
			var name = rowName(row);
			if (!name) return;
			var btn = row.querySelector(':scope > .wa-pin-toggle');
			if (!pinState.enabled) { if (btn) btn.remove(); return; }
			if (!btn) {
				btn = document.createElement('button');
				btn.className = 'wa-pin-toggle';
				btn.innerHTML = PIN_SVG;
				btn.addEventListener('click', function (e) {
					e.preventDefault(); e.stopPropagation();
					var n = btn._name;
					if (isPinned(n)) { pinState.names = pinState.names.filter(function (x) { return x !== n; }); toast('Pin dilepas: ' + n); }
					else { pinState.names.push(n); toast('Dipin: ' + n); }
					savePins(); refreshPins();
				}, true);
				if (getComputedStyle(row).position === 'static') row.style.position = 'relative';
				row.appendChild(btn);
			}
			btn._name = name;
			btn.title = isPinned(name) ? 'Lepas pin' : 'Pin chat ini';
			btn.classList.toggle('pinned', isPinned(name));
			row.classList.add('wa-pin-host');
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
	// 5. Control Center wiring (called from main.go after the modal exists)
	// ------------------------------------------------------------------
	window.waBindCoreUI = function (modal) {
		var pinCb = modal.querySelector('#wa-cc-pin-cb');
		var radarCb = modal.querySelector('#wa-cc-radar-cb');
		var radarNames = modal.querySelector('#wa-cc-radar-names');
		var diagBtn = modal.querySelector('#wa-diag-run');

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
			refreshPins();
		};
	};

	// ------------------------------------------------------------------
	// 6. Boot
	// ------------------------------------------------------------------
	whenDOMReady(function () {
		new MutationObserver(schedulePins).observe(document.body, { childList: true, subtree: true });
		schedulePins();
		setInterval(radarTick, 1500);
	});
})();
