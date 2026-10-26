// WhatsApp tools: clickable phone numbers in messages, a formatting toolbar for the composer, and chat
// labels with per-contact notes. Uses helpers exposed by addon-core.js (waSel, waRowName, waOpenChatName,
// waOpenChatByName) and addon-main.js (waInsertIntoEditor, waOpenChatByNumber). None of the tools changes
// WhatsApp's own DOM text: phone numbers are marked with the CSS Custom Highlight API.
(function() {
	if (window.self !== window.top) return;

	function toast(msg) { if (typeof showAddonToast === 'function') showAddonToast(msg); }
	function loadJSON(k, def) { try { var v = JSON.parse(localStorage.getItem(k)); return v == null ? def : v; } catch (e) { return def; } }
	function saveJSON(k, v) { try { localStorage.setItem(k, JSON.stringify(v)); } catch (e) {} }
	function sel(key, root) { return window.waSel ? window.waSel.one(key, root) : null; }
	function selAll(key, root) { return window.waSel ? window.waSel.all(key, root) : []; }
	function esc(s) { return String(s).replace(/[&<>"']/g, function(c) { return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]; }); }

	function addStyle() {
		if (document.getElementById('wa-tools-style')) return;
		var st = document.createElement('style');
		st.id = 'wa-tools-style';
		st.textContent = [
			'::highlight(wa-phone) { text-decoration: underline; text-decoration-style: dotted; text-decoration-color: #00a884; color: #00a884; }',
			'.wa-tool-pop { position: fixed; z-index: 2147483600; background: var(--wa-bg-elevated, #182229); border: 1px solid var(--wa-border-strong, rgba(134,150,160,.28)); border-radius: 10px; box-shadow: 0 10px 30px rgba(0,0,0,.45); padding: 4px; display: flex; gap: 2px; font-family: var(--wa-font, "Segoe UI", sans-serif); font-size: 12.5px; color: var(--wa-text, #e9edef); }',
			'.wa-tool-pop.vertical { flex-direction: column; min-width: 190px; }',
			'.wa-tool-pop button { background: none; border: none; color: inherit; font: inherit; padding: 6px 10px; border-radius: 6px; cursor: pointer; text-align: left; white-space: nowrap; }',
			'.wa-tool-pop button:hover { background: rgba(134,150,160,.18); }',
			'.wa-tool-pop .wa-fmt { min-width: 30px; text-align: center; padding: 5px 7px; }',
			'.wa-tool-pop .wa-pop-title { padding: 6px 10px 4px; color: var(--wa-text-muted, #8696a0); font-size: 11.5px; }',
			'.wa-label-dots { position: absolute; left: 76px; bottom: 3px; display: flex; gap: 3px; pointer-events: none; z-index: 2; }',
			'.wa-label-dots i { width: 7px; height: 7px; border-radius: 50%; display: block; }',
			'.wa-label-dots b { font-size: 9px; line-height: 7px; color: var(--wa-text-muted, #8696a0); font-weight: 400; }',
			'#wa-label-bar { display: flex; flex-wrap: wrap; gap: 6px; padding: 6px 12px; background: inherit; }',
			'#wa-label-bar:empty { display: none; }',
			'.wa-label-chip { display: inline-flex; align-items: center; gap: 6px; border: 1px solid var(--wa-border-strong, rgba(134,150,160,.28)); background: transparent; color: inherit; border-radius: 14px; padding: 3px 10px; font-size: 12px; cursor: pointer; font-family: inherit; }',
			'.wa-label-chip i { width: 8px; height: 8px; border-radius: 50%; display: inline-block; }',
			'#wa-chat-tag { position: fixed; z-index: 2147483000; display: flex; align-items: center; gap: 6px; max-width: 46%; background: var(--wa-bg-elevated, #182229); border: 1px solid var(--wa-border, rgba(134,150,160,.14)); border-radius: 14px; padding: 3px 10px; font-family: var(--wa-font, "Segoe UI", sans-serif); font-size: 11.5px; color: var(--wa-text, #e9edef); box-shadow: 0 4px 14px rgba(0,0,0,.25); cursor: pointer; opacity: .92; }',
			'#wa-chat-tag:hover { opacity: 1; }',
			'#wa-chat-tag .wa-tag-note { color: var(--wa-text-muted, #8696a0); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }'
		].join('\n');
		(document.head || document.documentElement).appendChild(st);
	}

	var popup = null;
	function closePopup() { if (popup) { popup.remove(); popup = null; } }
	function openPopup(x, y, build, vertical) {
		closePopup();
		popup = document.createElement('div');
		popup.className = 'wa-tool-pop' + (vertical ? ' vertical' : '');
		build(popup);
		document.body.appendChild(popup);
		var r = popup.getBoundingClientRect();
		popup.style.left = Math.max(8, Math.min(x, innerWidth - r.width - 8)) + 'px';
		popup.style.top = Math.max(8, Math.min(y, innerHeight - r.height - 8)) + 'px';
		return popup;
	}
	function popButton(parent, label, onClick, cls) {
		var b = document.createElement('button');
		b.textContent = label;
		if (cls) b.className = cls;
		b.addEventListener('mousedown', function(e) { e.preventDefault(); }); // keep the composer selection
		b.addEventListener('click', function(e) { e.preventDefault(); e.stopPropagation(); onClick(); });
		parent.appendChild(b);
		return b;
	}
	document.addEventListener('mousedown', function(e) { if (popup && !popup.contains(e.target)) closePopup(); }, true);
	document.addEventListener('keydown', function(e) { if (e.key === 'Escape') closePopup(); }, true);

	// ------------------------------------------------------------------
	// 1. Phone numbers in messages: highlighted, click for "chat" / "copy"
	// ------------------------------------------------------------------
	var PHONE_CANDIDATE = /(^|[^\w+])(\+?\d[\d\s().-]{7,19}\d)(?![\w])/g;

	// normalizePhone returns international digits (62812...) for numbers that look like phone numbers:
	// Indonesian mobile (08..., 628..., +628...) or any number written with a leading +.
	function normalizePhone(raw) {
		var plus = raw.trim().charAt(0) === '+';
		var d = raw.replace(/\D/g, '');
		if (d.length < 9 || d.length > 15) return null;
		if (/^08\d{7,12}$/.test(d) && !plus) return '62' + d.slice(1);
		if (/^628\d{7,11}$/.test(d)) return d;
		if (plus) return d;
		return null;
	}
	window.waNormalizePhone = normalizePhone;

	var phoneMatches = [];
	var highlightOK = typeof CSS !== 'undefined' && CSS.highlights && typeof Highlight === 'function';

	function scanPhones() {
		if (!highlightOK) return;
		var main = document.getElementById('main');
		phoneMatches = [];
		if (!main) { CSS.highlights.delete('wa-phone'); return; }
		var ranges = [];
		var texts = main.querySelectorAll('span.selectable-text, [data-testid$="selectable-text"]');
		for (var i = 0; i < texts.length && i < 600; i++) {
			var walker = document.createTreeWalker(texts[i], NodeFilter.SHOW_TEXT);
			var node;
			while ((node = walker.nextNode())) {
				var t = node.textContent;
				if (t.length < 9 || !/\d{3}/.test(t)) continue;
				PHONE_CANDIDATE.lastIndex = 0;
				var m;
				while ((m = PHONE_CANDIDATE.exec(t))) {
					var num = normalizePhone(m[2]);
					if (!num) continue;
					var start = m.index + m[1].length;
					var r = document.createRange();
					r.setStart(node, start);
					r.setEnd(node, start + m[2].length);
					ranges.push(r);
					phoneMatches.push({ range: r, number: num, text: m[2] });
				}
			}
		}
		if (!ranges.length) { CSS.highlights.delete('wa-phone'); return; }
		var h = new Highlight(); // takes ranges one by one, not an array
		ranges.forEach(function(r) { h.add(r); });
		CSS.highlights.set('wa-phone', h);
	}

	function phoneAt(x, y) {
		var pos = document.caretRangeFromPoint ? document.caretRangeFromPoint(x, y) : null;
		if (!pos) return null;
		for (var i = 0; i < phoneMatches.length; i++) {
			var r = phoneMatches[i].range;
			try {
				if (r.startContainer === pos.startContainer && r.isPointInRange(pos.startContainer, pos.startOffset)) {
					// Make sure the pointer is really over the number, not just past the end of the line
					var rects = r.getClientRects();
					for (var j = 0; j < rects.length; j++) {
						var b = rects[j];
						if (x >= b.left - 1 && x <= b.right + 1 && y >= b.top - 1 && y <= b.bottom + 1) return phoneMatches[i];
					}
				}
			} catch (e) {}
		}
		return null;
	}

	document.addEventListener('click', function(e) {
		if (!phoneMatches.length || e.button !== 0) return;
		if (!e.target.closest || !e.target.closest('#main')) return;
		var s = window.getSelection();
		if (s && !s.isCollapsed && String(s).length > 0) return; // the user is selecting text
		var hit = phoneAt(e.clientX, e.clientY);
		if (!hit) return;
		e.preventDefault();
		e.stopPropagation();
		openPopup(e.clientX + 4, e.clientY + 10, function(p) {
			var title = document.createElement('div');
			title.className = 'wa-pop-title';
			title.textContent = '+' + hit.number;
			p.appendChild(title);
			popButton(p, 'Chat ke nomor ini', function() {
				closePopup();
				if (window.waOpenChatByNumber) window.waOpenChatByNumber('+' + hit.number);
			});
			popButton(p, 'Salin nomor', function() {
				closePopup();
				navigator.clipboard.writeText('+' + hit.number).then(function() { toast('Nomor disalin: +' + hit.number); });
			});
		}, true);
	}, true);

	// ------------------------------------------------------------------
	// 2. Formatting toolbar: select text in the composer to format it
	//    with WhatsApp's own markers (*bold* _italic_ ~strike~ ```mono```).
	// ------------------------------------------------------------------
	var INLINE = [
		{ label: 'B', title: 'Tebal', mark: '*', style: 'font-weight:700' },
		{ label: 'I', title: 'Miring', mark: '_', style: 'font-style:italic' },
		{ label: 'S', title: 'Coret', mark: '~', style: 'text-decoration:line-through' },
		{ label: '</>', title: 'Monospace', mark: '```', style: 'font-family:monospace' },
		{ label: '`', title: 'Kode sebaris', mark: '`', style: 'font-family:monospace' }
	];
	var BLOCK = [
		{ label: '❝', title: 'Kutipan', prefix: function() { return '> '; } },
		{ label: '•', title: 'Daftar berbutir', prefix: function() { return '- '; } },
		{ label: '1.', title: 'Daftar bernomor', prefix: function(i) { return (i + 1) + '. '; } }
	];

	// wrapInline adds (or removes, when already there) a marker around each non-empty line. Spaces stay
	// outside the markers, as WhatsApp only formats text that starts and ends right at a marker.
	function wrapInline(text, mark) {
		var lines = text.split('\n');
		var all = lines.every(function(l) {
			var t = l.trim();
			return !t || (t.length > 2 * mark.length && t.slice(0, mark.length) === mark && t.slice(-mark.length) === mark);
		});
		return lines.map(function(l) {
			var m = l.match(/^(\s*)(.*?)(\s*)$/);
			if (!m[2]) return l;
			var core = all ? m[2].slice(mark.length, m[2].length - mark.length) : mark + m[2] + mark;
			return m[1] + core + m[3];
		}).join('\n');
	}

	function prefixLines(text, prefixFn) {
		var n = 0;
		return text.split('\n').map(function(l) { return l.trim() ? prefixFn(n++) + l : l; }).join('\n');
	}
	window.waWrapInline = wrapInline;
	window.waPrefixLines = prefixLines;

	function composerSelection() {
		var s = window.getSelection();
		if (!s || s.isCollapsed || !s.rangeCount) return null;
		var composer = sel('composer');
		if (!composer || !composer.contains(s.anchorNode) || !composer.contains(s.focusNode)) return null;
		var text = s.toString();
		if (!text.trim()) return null;
		return { composer: composer, text: text, rect: s.getRangeAt(0).getBoundingClientRect() };
	}

	function applyFormat(fn) {
		var cs = composerSelection();
		if (!cs || !window.waInsertIntoEditor) return;
		window.waInsertIntoEditor(cs.composer, fn(cs.text));
		closePopup();
	}

	var fmtTimer = null;
	document.addEventListener('selectionchange', function() {
		clearTimeout(fmtTimer);
		fmtTimer = setTimeout(function() {
			var cs = composerSelection();
			if (!cs) { if (popup && popup.classList.contains('wa-fmt-bar')) closePopup(); return; }
			if (popup && popup.classList.contains('wa-fmt-bar')) return;
			var p = openPopup(cs.rect.left, cs.rect.top - 46, function(p) {
				INLINE.forEach(function(f) {
					var b = popButton(p, f.label, function() { applyFormat(function(t) { return wrapInline(t, f.mark); }); }, 'wa-fmt');
					b.title = f.title + ' (' + f.mark + 'teks' + f.mark + ')';
					b.setAttribute('style', f.style);
				});
				BLOCK.forEach(function(f) {
					var b = popButton(p, f.label, function() { applyFormat(function(t) { return prefixLines(t, f.prefix); }); }, 'wa-fmt');
					b.title = f.title;
				});
			});
			p.classList.add('wa-fmt-bar');
		}, 180);
	});

	// ------------------------------------------------------------------
	// 3. Labels and contact notes (stored locally, included in backups)
	// ------------------------------------------------------------------
	var DEFAULT_LABELS = [
		{ id: 'l1', name: 'Pelanggan', color: '#25d366' },
		{ id: 'l2', name: 'Belum bayar', color: '#f59e0b' },
		{ id: 'l3', name: 'Sudah kirim', color: '#3b82f6' },
		{ id: 'l4', name: 'Penting', color: '#ef4444' }
	];
	function labelDefs() { var d = loadJSON('wa_label_defs', null); return Array.isArray(d) ? d : DEFAULT_LABELS.slice(); }
	function chatLabels() { var m = loadJSON('wa_chat_labels', {}); return m && typeof m === 'object' ? m : {}; }
	function contactNotes() { var m = loadJSON('wa_contact_notes', {}); return m && typeof m === 'object' ? m : {}; }
	function labelsOf(chat) {
		var ids = chatLabels()[chat] || [];
		return labelDefs().filter(function(d) { return ids.indexOf(d.id) !== -1; });
	}

	function setChatData(chat, ids, note) {
		var map = chatLabels();
		if (ids.length) map[chat] = ids; else delete map[chat];
		saveJSON('wa_chat_labels', map);
		var notes = contactNotes();
		note = (note || '').trim();
		if (note) notes[chat] = note; else delete notes[chat];
		saveJSON('wa_contact_notes', notes);
		refreshLabels(true);
	}

	function decorateRowsWithLabels() {
		var notes = contactNotes();
		selAll('chatRows').forEach(function(row) {
			var name = window.waRowName ? window.waRowName(row) : '';
			if (!name) return;
			var labels = labelsOf(name);
			var key = labels.map(function(l) { return l.color; }).join(',') + (notes[name] ? '|n' : '');
			var box = row.querySelector(':scope > .wa-label-dots');
			if (!key) { if (box) box.remove(); return; }
			if (box && box._key === key) return;
			if (!box) {
				box = document.createElement('div');
				box.className = 'wa-label-dots';
				if (getComputedStyle(row).position === 'static') row.style.position = 'relative';
				row.appendChild(box);
			}
			box._key = key;
			box.innerHTML = '';
			labels.slice(0, 4).forEach(function(l) {
				var i = document.createElement('i');
				i.style.background = l.color;
				i.title = l.name;
				box.appendChild(i);
			});
			if (notes[name]) { var b = document.createElement('b'); b.textContent = '✎'; box.appendChild(b); }
		});
	}

	function renderLabelBar() {
		var pane = sel('paneSide');
		if (!pane) return;
		var bar = document.getElementById('wa-label-bar');
		if (!bar) {
			bar = document.createElement('div');
			bar.id = 'wa-label-bar';
		}
		var pinBar = document.getElementById('wa-pin-bar');
		var anchor = pinBar && pinBar.parentElement === pane ? pinBar.nextSibling : pane.firstChild;
		if (bar.parentElement !== pane) pane.insertBefore(bar, anchor);
		var map = chatLabels();
		var counts = {};
		Object.keys(map).forEach(function(chat) { (map[chat] || []).forEach(function(id) { counts[id] = (counts[id] || 0) + 1; }); });
		var defs = labelDefs().filter(function(d) { return counts[d.id]; });
		var key = defs.map(function(d) { return d.id + d.color + d.name + counts[d.id]; }).join('|');
		if (bar._key === key) return;
		bar._key = key;
		bar.innerHTML = '';
		defs.forEach(function(d) {
			var chip = document.createElement('button');
			chip.className = 'wa-label-chip';
			chip.innerHTML = '<i></i><span></span>';
			chip.firstChild.style.background = d.color;
			chip.lastChild.textContent = d.name + ' (' + counts[d.id] + ')';
			chip.onclick = function(e) {
				e.preventDefault(); e.stopPropagation();
				var chats = Object.keys(chatLabels()).filter(function(c) { return (chatLabels()[c] || []).indexOf(d.id) !== -1; }).sort();
				var r = chip.getBoundingClientRect();
				openPopup(r.left, r.bottom + 4, function(p) {
					var t = document.createElement('div');
					t.className = 'wa-pop-title';
					t.textContent = d.name;
					p.appendChild(t);
					chats.forEach(function(c) {
						popButton(p, c, function() { closePopup(); if (window.waOpenChatByName) window.waOpenChatByName(c); });
					});
				}, true);
			};
			bar.appendChild(chip);
		});
	}

	// Small tag at the top of the open chat: its labels and the first line of its note.
	function renderChatTag() {
		var main = document.getElementById('main');
		var chat = main && window.waOpenChatName ? window.waOpenChatName() : '';
		var tag = document.getElementById('wa-chat-tag');
		var labels = chat ? labelsOf(chat) : [];
		var note = chat ? (contactNotes()[chat] || '') : '';
		if (!chat || (!labels.length && !note)) { if (tag) tag.remove(); return; }
		var header = main.querySelector('header');
		var hr = (header || main).getBoundingClientRect();
		var mr = main.getBoundingClientRect();
		if (!tag) {
			tag = document.createElement('div');
			tag.id = 'wa-chat-tag';
			tag.title = 'Label & catatan (Alt+K)';
			tag.onclick = function() { window.waOpenLabelEditor(); };
			document.body.appendChild(tag);
		}
		var key = chat + '|' + labels.map(function(l) { return l.id; }).join(',') + '|' + note;
		if (tag._key !== key) {
			tag._key = key;
			tag.innerHTML = '';
			labels.forEach(function(l) {
				var c = document.createElement('span');
				c.className = 'wa-label-chip';
				c.style.cssText = 'cursor:inherit;padding:1px 8px;font-size:11px;';
				c.innerHTML = '<i></i><span></span>';
				c.firstChild.style.background = l.color;
				c.lastChild.textContent = l.name;
				tag.appendChild(c);
			});
			if (note) {
				var n = document.createElement('span');
				n.className = 'wa-tag-note';
				n.textContent = '✎ ' + note.split('\n')[0];
				tag.appendChild(n);
			}
		}
		tag.style.top = Math.round(hr.bottom + 8) + 'px';
		tag.style.right = Math.round(innerWidth - mr.right + 16) + 'px';
	}

	window.waOpenLabelEditor = function() {
		var chat = window.waOpenChatName ? window.waOpenChatName() : '';
		if (!chat) { toast('Buka sebuah chat dulu'); return; }
		var modal = document.getElementById('wa-label-modal');
		if (!modal) {
			modal = document.createElement('div');
			modal.id = 'wa-label-modal';
			modal.className = 'wa-modal-backdrop';
			modal.innerHTML = '<div class="wa-modal-box" style="width:360px;padding:22px;">' +
				'<h3 style="margin:0 0 2px 0;font-size:16px;font-weight:600;color:var(--wa-text);">Label & catatan</h3>' +
				'<p class="wa-lm-chat" style="margin:0 0 14px 0;font-size:12.5px;color:var(--wa-text-muted);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;"></p>' +
				'<div class="wa-lm-labels" style="display:flex;flex-direction:column;gap:8px;margin-bottom:12px;"></div>' +
				'<textarea class="wa-input wa-lm-note" rows="4" placeholder="Catatan untuk kontak ini, misalnya alamat, ukuran, atau riwayat pesanan" style="width:100%;height:auto;padding:8px 11px;resize:vertical;margin-bottom:12px;"></textarea>' +
				'<div style="display:flex;gap:8px;"><button class="wa-btn wa-btn-secondary wa-lm-cancel" style="flex:1;">Batal</button><button class="wa-btn wa-btn-primary wa-lm-save" style="flex:1;">Simpan</button></div>' +
				'</div>';
			document.body.appendChild(modal);
			modal.onclick = function(e) { if (e.target === modal) modal.style.display = 'none'; };
			modal.querySelector('.wa-lm-cancel').onclick = function() { modal.style.display = 'none'; };
			var ta = modal.querySelector('.wa-lm-note');
			['keydown', 'keyup', 'keypress', 'input'].forEach(function(n) { ta.addEventListener(n, function(e) { e.stopPropagation(); }); });
			ta.addEventListener('keydown', function(e) { if (e.key === 'Escape') modal.style.display = 'none'; });
		}
		modal.querySelector('.wa-lm-chat').textContent = chat;
		var box = modal.querySelector('.wa-lm-labels');
		box.innerHTML = '';
		var mine = chatLabels()[chat] || [];
		labelDefs().forEach(function(d) {
			var row = document.createElement('label');
			row.style.cssText = 'display:flex;align-items:center;gap:10px;cursor:pointer;font-size:13px;color:var(--wa-text);';
			row.innerHTML = '<input type="checkbox"><i style="width:10px;height:10px;border-radius:50%;display:inline-block;"></i><span></span>';
			row.firstChild.checked = mine.indexOf(d.id) !== -1;
			row.firstChild.value = d.id;
			row.children[1].style.background = d.color;
			row.lastChild.textContent = d.name;
			box.appendChild(row);
		});
		modal.querySelector('.wa-lm-note').value = contactNotes()[chat] || '';
		modal.querySelector('.wa-lm-save').onclick = function() {
			var ids = [].slice.call(box.querySelectorAll('input:checked')).map(function(i) { return i.value; });
			setChatData(chat, ids, modal.querySelector('.wa-lm-note').value);
			modal.style.display = 'none';
			toast('Label & catatan disimpan: ' + chat);
		};
		modal.style.display = 'flex';
	};

	// Control Center: edit the label list (Chat & alat tab)
	window.waRenderLabelManager = function() {
		var list = document.getElementById('wa-label-defs');
		if (!list) return;
		var defs = labelDefs();
		list.innerHTML = '';
		defs.forEach(function(d, idx) {
			var row = document.createElement('div');
			row.className = 'wa-row';
			row.innerHTML = '<input type="color" style="width:28px;height:24px;border:none;background:none;padding:0;cursor:pointer;">' +
				'<input class="wa-input" type="text" maxlength="24" style="flex:1;margin:0 8px;">' +
				'<button class="wa-btn wa-btn-secondary wa-btn-sm">Hapus</button>';
			var color = row.children[0], name = row.children[1];
			color.value = d.color;
			name.value = d.name;
			['keydown', 'keyup', 'keypress', 'input'].forEach(function(n) { name.addEventListener(n, function(e) { e.stopPropagation(); }); });
			color.onchange = function() { defs[idx].color = color.value; saveJSON('wa_label_defs', defs); refreshLabels(true); };
			name.onchange = function() { if (name.value.trim()) { defs[idx].name = name.value.trim(); saveJSON('wa_label_defs', defs); refreshLabels(true); } };
			row.children[2].onclick = function() {
				var map = chatLabels();
				Object.keys(map).forEach(function(c) { map[c] = map[c].filter(function(id) { return id !== d.id; }); if (!map[c].length) delete map[c]; });
				saveJSON('wa_chat_labels', map);
				defs.splice(idx, 1);
				saveJSON('wa_label_defs', defs);
				refreshLabels(true);
				window.waRenderLabelManager();
			};
			list.appendChild(row);
		});
		var add = document.getElementById('wa-label-add');
		if (add) add.onclick = function() {
			defs.push({ id: 'l' + Date.now().toString(36), name: 'Label baru', color: '#8b5cf6' });
			saveJSON('wa_label_defs', defs);
			window.waRenderLabelManager();
		};
	};

	// ------------------------------------------------------------------
	// Refresh loop
	// ------------------------------------------------------------------
	var timer = null;
	function refreshLabels(force) {
		if (force) { var bar = document.getElementById('wa-label-bar'); if (bar) bar._key = null; }
		try { decorateRowsWithLabels(); renderLabelBar(); renderChatTag(); } catch (e) {}
	}
	function schedule() {
		if (timer) return;
		timer = setTimeout(function() {
			timer = null;
			try { scanPhones(); } catch (e) {}
			refreshLabels(false);
		}, 400);
	}

	window.addEventListener('keydown', function(e) {
		if (e.altKey && !e.ctrlKey && (e.code === 'KeyK' || e.key === 'k' || e.key === 'K')) {
			e.preventDefault(); e.stopPropagation();
			window.waOpenLabelEditor();
		}
	}, true);

	function boot() {
		addStyle();
		new MutationObserver(schedule).observe(document.body, { childList: true, subtree: true, characterData: true });
		window.addEventListener('resize', function() { renderChatTag(); });
		schedule();
	}
	if (document.body) boot(); else document.addEventListener('DOMContentLoaded', boot, { once: true });
})();
