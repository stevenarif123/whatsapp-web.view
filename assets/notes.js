// In-app notes (Alt+N), shared by the WhatsApp and Telegram windows. The notes are kept by Go
// (notes.go) so both windows show the same notes; localStorage keeps a copy for instant display.
// The drawer uses the add-on stylesheet (WA_ADDON_CSS), which Telegram also receives.
(function() {
	if (window.self !== window.top) return;

	function notesToast(msg) {
		if (typeof showAddonToast === 'function') showAddonToast(msg);
		else if (typeof window.tgShowToast === 'function') window.tgShowToast(msg);
	}

	function whenBody(fn) {
		if (document.body) { fn(); return; }
		document.addEventListener('DOMContentLoaded', fn, { once: true });
	}

	// WhatsApp injects the stylesheet itself; Telegram gets it here, with its own accent colour.
	function ensureStyles() {
		if (!document.getElementById('wa-impeccable-styles') && typeof WA_ADDON_CSS === 'string') {
			var style = document.createElement('style');
			style.id = 'wa-impeccable-styles';
			style.textContent = WA_ADDON_CSS;
			(document.head || document.documentElement).appendChild(style);
		}
		if (/(^|\.)telegram\.org$/.test(location.hostname) && !document.getElementById('wa-notes-tg-theme')) {
			var theme = document.createElement('style');
			theme.id = 'wa-notes-tg-theme';
			theme.textContent = ':root { --wa-primary: #24A1DE; --wa-primary-hover: #1d8cc4; --wa-primary-text: #6ec4f0; --wa-primary-subtle: rgba(36, 161, 222, 0.13); }';
			(document.head || document.documentElement).appendChild(theme);
		}
	}

	var WA_NOTES_KEY = 'wa_notes';
	var waNotes = { list: [], active: null };

	function newNote(text) {
		return { id: 'n' + Date.now().toString(36) + Math.random().toString(36).slice(2, 5), text: text || '', updated: Date.now() };
	}
	var pushTimer = null;
	function saveNotes() {
		try { localStorage.setItem(WA_NOTES_KEY, JSON.stringify(waNotes)); } catch(e) {}
		clearTimeout(pushTimer);
		pushTimer = setTimeout(function() {
			if (window.saveSharedNotes) window.saveSharedNotes(JSON.stringify(waNotes));
		}, 400);
	}
	// Takes the notes saved by Go (possibly edited in the other window). Notes that only exist in this
	// page's localStorage (from before notes were shared) are sent to Go the first time.
	function pullNotes(cb) {
		if (!window.getSharedNotes) { if (cb) cb(); return; }
		window.getSharedNotes().then(function(raw) {
			var parsed = null;
			try { parsed = raw ? JSON.parse(raw) : null; } catch(e) {}
			if (parsed && Array.isArray(parsed.list) && parsed.list.length) {
				waNotes = parsed;
				if (!waNotes.list.some(function(x) { return x.id === waNotes.active; })) waNotes.active = waNotes.list[0].id;
				try { localStorage.setItem(WA_NOTES_KEY, JSON.stringify(waNotes)); } catch(e) {}
			} else if (waNotes.list.some(function(x) { return (x.text || '').trim(); })) {
				if (window.saveSharedNotes) window.saveSharedNotes(JSON.stringify(waNotes));
			}
			if (cb) cb();
		}).catch(function() { if (cb) cb(); });
	}
	function loadNotes() {
		waNotes = { list: [], active: null };
		try {
			var raw = localStorage.getItem(WA_NOTES_KEY);
			if (raw) {
				var parsed = JSON.parse(raw);
				if (parsed && Array.isArray(parsed.list)) waNotes = parsed;
			} else {
				var legacy = localStorage.getItem('wa_scratchpad_notes');
				if (legacy) {
					var first = newNote(legacy);
					waNotes = { list: [first], active: first.id };
					saveNotes();
				}
			}
		} catch(e) {}
		if (!waNotes.list.length) {
			var n = newNote('');
			waNotes = { list: [n], active: n.id };
		}
		if (!waNotes.list.some(function(x) { return x.id === waNotes.active; })) waNotes.active = waNotes.list[0].id;
	}
	function activeNote() {
		return waNotes.list.filter(function(x) { return x.id === waNotes.active; })[0] || null;
	}
	function noteTitle(n) {
		var line = (n.text || '').split('\n').map(function(s) { return s.trim(); }).filter(Boolean)[0];
		if (!line) return 'Catatan kosong';
		return line.length > 38 ? line.slice(0, 38) + '…' : line;
	}
	function noteDate(ts) {
		try { return new Date(ts).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' }); } catch(e) { return ''; }
	}

	function createScratchpadDrawer() {
		var existing = document.getElementById('wa-scratchpad-drawer');
		if (existing) return existing;
		if (!document.body) return null;
		loadNotes();

		var drawer = document.createElement('div');
		drawer.id = 'wa-scratchpad-drawer';
		drawer.className = 'wa-scratchpad';
		drawer.innerHTML = '<div class="wa-modal-header" style="padding:14px 18px;">' +
			'<div class="wa-header-left">' +
			'  <div class="wa-header-icon" style="width:30px;height:30px;">' +
			'    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>' +
			'  </div>' +
			'  <h3 class="wa-header-title" style="font-size:14px;">Catatan</h3>' +
			'</div>' +
			'<button id="wa-sp-close" class="wa-close-btn" title="Tutup" aria-label="Tutup">' +
			'  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>' +
			'</button>' +
			'</div>' +
			'<div class="wa-note-tools">' +
			'  <input id="wa-note-search" class="wa-input" type="text" placeholder="Cari catatan…" style="flex:1;min-width:0;">' +
			'  <button id="wa-note-new" class="wa-btn wa-btn-primary wa-btn-sm">+ Baru</button>' +
			'</div>' +
			'<div id="wa-note-list" class="wa-note-list"></div>' +
			'<div style="flex:1;padding:12px 18px 0;display:flex;flex-direction:column;min-height:0;">' +
			'<textarea id="wa-sp-textarea" class="wa-input" style="flex:1;width:100%;resize:none;line-height:1.5;"></textarea>' +
			'<div style="display:flex;align-items:center;justify-content:space-between;margin:10px 0;font-size:11px;color:var(--wa-text-muted);">' +
			'<span id="wa-sp-count">0 karakter</span>' +
			'<div style="display:flex;gap:6px;">' +
			'<button id="wa-sp-copy" class="wa-btn wa-btn-secondary wa-btn-sm">Salin</button>' +
			'<button id="wa-sp-clear" class="wa-btn wa-btn-danger wa-btn-sm">Hapus</button>' +
			'</div></div></div>' +
			'<div style="padding:10px 18px;border-top:1px solid var(--wa-border);font-size:11px;color:var(--wa-text-dim);text-align:center;">' +
			'Pintasan <b style="color:var(--wa-text-muted);">Alt + N</b> untuk membuka / menutup catatan' +
			'</div>';
		document.body.appendChild(drawer);

		var ta = drawer.querySelector('#wa-sp-textarea');
		var count = drawer.querySelector('#wa-sp-count');
		var listEl = drawer.querySelector('#wa-note-list');
		var search = drawer.querySelector('#wa-note-search');
		ta.placeholder = 'Tulis draf pesan, nomor resi, to-do list, catatan telepon di sini...\n\nTersimpan otomatis.';

		function renderList() {
			var q = (search.value || '').trim().toLowerCase();
			var items = waNotes.list.filter(function(n) { return !q || (n.text || '').toLowerCase().indexOf(q) !== -1; });
			items.sort(function(a, b) { return b.updated - a.updated; });
			listEl.innerHTML = '';
			if (!items.length) {
				var empty = document.createElement('div');
				empty.className = 'wa-list-empty';
				empty.textContent = q ? 'Tidak ada catatan yang cocok' : 'Belum ada catatan';
				listEl.appendChild(empty);
				return;
			}
			items.forEach(function(n) {
				var row = document.createElement('div');
				row.className = 'wa-note-item' + (n.id === waNotes.active ? ' active' : '');
				var title = document.createElement('span');
				title.className = 'wa-note-title';
				title.textContent = noteTitle(n);
				var when = document.createElement('span');
				when.className = 'wa-note-date';
				when.textContent = noteDate(n.updated);
				row.appendChild(title);
				row.appendChild(when);
				row.onclick = function() { selectNote(n.id); };
				listEl.appendChild(row);
			});
		}
		function showActive() {
			var n = activeNote();
			ta.value = n ? n.text : '';
			count.textContent = ta.value.length + ' karakter';
		}
		function selectNote(id) {
			waNotes.active = id;
			saveNotes();
			showActive();
			renderList();
			ta.focus();
		}

		ta.oninput = function() {
			var n = activeNote();
			if (!n) return;
			n.text = ta.value;
			n.updated = Date.now();
			saveNotes();
			count.textContent = ta.value.length + ' karakter';
			renderList();
		};
		search.oninput = renderList;
		search.addEventListener('keydown', function(e) { if (e.stopPropagation) e.stopPropagation(); });
		ta.addEventListener('keydown', function(e) { if (e.stopPropagation && !e.altKey) e.stopPropagation(); });

		drawer.querySelector('#wa-note-new').onclick = function() {
			var cur = activeNote();
			if (cur && !cur.text.trim()) { search.value = ''; selectNote(cur.id); return; }
			var n = newNote('');
			waNotes.list.push(n);
			search.value = '';
			selectNote(n.id);
		};
		drawer.querySelector('#wa-sp-close').onclick = function() { drawer.style.right = '-360px'; };

		var copyBtn = drawer.querySelector('#wa-sp-copy');
		copyBtn.onclick = function() {
			if (!ta.value) return;
			navigator.clipboard.writeText(ta.value).then(function() {
				notesToast('📋 Catatan disalin ke clipboard');
				copyBtn.textContent = 'Tersalin';
				setTimeout(function() { copyBtn.textContent = 'Salin'; }, 1800);
			});
		};
		drawer.querySelector('#wa-sp-clear').onclick = function() {
			var n = activeNote();
			if (!n) return;
			if (n.text.trim() && !confirm('Hapus catatan ini?')) return;
			waNotes.list = waNotes.list.filter(function(x) { return x.id !== n.id; });
			if (!waNotes.list.length) waNotes.list.push(newNote(''));
			waNotes.active = waNotes.list[0].id;
			saveNotes();
			showActive();
			renderList();
		};

		drawer._refresh = function() { showActive(); renderList(); };
		showActive();
		renderList();
		return drawer;
	}

	window.toggleScratchpad = function() {
		whenBody(function() {
			ensureStyles();
			var drawer = createScratchpadDrawer();
			if (!drawer) return;
			var isOpen = drawer.style.right === '0px';
			drawer.style.right = isOpen ? '-360px' : '0px';
			if (!isOpen) {
				pullNotes(drawer._refresh);
				var ta = document.getElementById('wa-sp-textarea');
				if (ta) setTimeout(function() { ta.focus(); }, 100);
			}
		});
	};

	window.addEventListener('keydown', function(e) {
		if (e.altKey && (e.code === 'KeyN' || e.key === 'n' || e.key === 'N')) {
			if (e.preventDefault) e.preventDefault();
			if (e.stopPropagation) e.stopPropagation();
			window.toggleScratchpad();
		}
	}, true);
})();
