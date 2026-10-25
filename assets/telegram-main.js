	// UserAgent override
	Object.defineProperty(navigator, 'userAgent', {
		get: () => '__WA_USER_AGENT__'
	});
	Object.defineProperty(navigator, 'appVersion', {
		get: () => '__WA_USER_AGENT__'
	});

	// External link protection: Open external links in default system browser
	(function() {
		var origOpen = window.open;
		window.open = function(url) {
			if (url && typeof url === 'string') {
				try {
					var u = new URL(url, window.location.href);
					if (!u.hostname.includes('telegram.org') && (u.protocol === 'http:' || u.protocol === 'https:')) {
						if (window.openExternalLink) window.openExternalLink(u.href);
						return null;
					}
				} catch(e) {}
			}
			return origOpen.apply(this, arguments);
		};
	})();

	// In-App Tab Switcher (Floating Glassmorphic Dock for Telegram)
	(function() {
		function injectTabs() {
			var root = document.documentElement || document.body;
			if (!root) return;

			if (!document.getElementById('tg-dock-style')) {
				var style = document.createElement('style');
				style.id = 'tg-dock-style';
				style.textContent = [
					'.wa-dock { position:fixed !important;bottom:16px !important;left:16px !important;z-index:2147483647 !important;display:flex !important;align-items:center !important;gap:4px !important;background:rgba(23,33,43,0.96) !important;backdrop-filter:blur(24px) !important;-webkit-backdrop-filter:blur(24px) !important;border:1px solid rgba(255,255,255,0.16) !important;padding:4px 8px !important;border-radius:24px !important;box-shadow:0 8px 32px rgba(0,0,0,0.7) !important;user-select:none !important;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,Helvetica,sans-serif !important;pointer-events:auto !important;visibility:visible !important;opacity:1 !important; }',
					'.wa-dock-item { display:flex !important;align-items:center !important;gap:6px !important;padding:6px 13px !important;border-radius:18px !important;font-size:12px !important;font-weight:500 !important;color:#9db2c6 !important;background:transparent !important;border:none !important;cursor:pointer !important;transition:all 0.15s cubic-bezier(0.16,1,0.3,1) !important;outline:none !important;white-space:nowrap !important; }',
					'.wa-dock-item:hover { color:#ffffff !important;background:rgba(255,255,255,0.08) !important; }',
					'.wa-dock-item.active { background:rgba(36,161,222,0.22) !important;color:#24A1DE !important;font-weight:600 !important;cursor:default !important; }',
					'.wa-dock-sep { width:1px !important;height:14px !important;background:rgba(255,255,255,0.12) !important;margin:0 2px !important; }',
					'.wa-dot-wa { width:7px !important;height:7px !important;border-radius:50% !important;background:#00a884 !important;display:inline-block !important;box-shadow:0 0 6px rgba(0,168,132,0.6) !important; }',
					'.wa-dot-tg { width:7px !important;height:7px !important;border-radius:50% !important;background:#24A1DE !important;display:inline-block !important; }',
					'.tg-qr-popup { position:fixed !important; z-index:2147483647 !important; background:rgba(23,33,43,0.98) !important; backdrop-filter:blur(20px) !important; -webkit-backdrop-filter:blur(20px) !important; border:1px solid rgba(255,255,255,0.16) !important; border-radius:10px !important; box-shadow:0 16px 40px rgba(0,0,0,0.75) !important; width:380px !important; max-width:90vw !important; max-height:260px !important; display:none; flex-direction:column !important; overflow:hidden !important; font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,Helvetica,sans-serif !important; user-select:none !important; pointer-events:auto !important; }',
					'.tg-qr-header { padding:8px 12px !important; background:rgba(15,22,29,0.95) !important; border-bottom:1px solid rgba(255,255,255,0.08) !important; display:flex !important; align-items:center !important; justify-content:space-between !important; font-size:11px !important; color:#9db2c6 !important; }',
					'.tg-qr-badge { font-size:10px !important; padding:2px 6px !important; border-radius:4px !important; background:rgba(36,161,222,0.18) !important; color:#24A1DE !important; font-weight:600 !important; }',
					'.tg-qr-list { overflow-y:auto !important; padding:4px !important; display:flex !important; flex-direction:column !important; gap:2px !important; max-height:210px !important; }',
					'.tg-qr-item { display:flex !important; flex-direction:column !important; padding:8px 10px !important; border-radius:6px !important; cursor:pointer !important; transition:all 0.12s ease !important; border:1px solid transparent !important; }',
					'.tg-qr-item:hover, .tg-qr-item.active { background:rgba(255,255,255,0.08) !important; border-color:rgba(255,255,255,0.12) !important; }',
					'.tg-qr-item.active { background:rgba(36,161,222,0.22) !important; border-color:rgba(36,161,222,0.4) !important; }',
					'.tg-qr-item-key { font-size:12.5px !important; font-weight:600 !important; color:#24A1DE !important; display:flex !important; align-items:center !important; gap:6px !important; }',
					'.tg-qr-item-text { font-size:11.5px !important; color:#9db2c6 !important; white-space:nowrap !important; overflow:hidden !important; text-overflow:ellipsis !important; margin-top:2px !important; }'
				].join('\n');
				(document.head || root).appendChild(style);
			}

			var bar = document.getElementById('tg-messenger-tabs');
			if (!bar) {
				bar = document.createElement('div');
				bar.id = 'tg-messenger-tabs';
				bar.className = 'wa-dock';
				bar.innerHTML = [
					'<button id="tg-tab-wa" class="wa-dock-item" title="Kembali ke WhatsApp (Ctrl+1)">' +
						'<span class="wa-dot-wa"></span><span>WhatsApp</span>' +
					'</button>',
					'<div class="wa-dock-sep"></div>',
					'<button id="tg-tab-tg" class="wa-dock-item active" title="Telegram Aktif (Ctrl+2)">' +
						'<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>' +
						'<span>Telegram</span>' +
					'</button>',
					'<div class="wa-dock-sep"></div>',
					'<button id="tg-tab-split" class="wa-dock-item" title="Mode Berdampingan 50:50 (Ctrl+3)">' +
						'<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="12" y1="3" x2="12" y2="21"/></svg>' +
						'<span>Berdampingan</span>' +
					'</button>',
					'<div class="wa-dock-sep"></div>',
					'<button id="tg-tab-privacy" class="wa-dock-item" title="Mode Privasi: blur chat (Alt+P). Shift+klik untuk ganti intensitas">' +
						'<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>' +
						'<span>Privasi</span>' +
					'</button>',
					'<div class="wa-dock-sep"></div>',
					'<button id="tg-tab-notes" class="wa-dock-item" title="Catatan, sama dengan di WhatsApp (Alt+N)">' +
						'<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>' +
						'<span>Catatan</span>' +
					'</button>',
					'<div class="wa-dock-sep"></div>',
					'<button id="tg-tab-lock" class="wa-dock-item" title="Kunci Telegram (Ctrl+L)">' +
						'<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>' +
						'<span>Kunci</span>' +
					'</button>'
				].join('');

				var btnWA = bar.querySelector('#tg-tab-wa');
				var btnTG = bar.querySelector('#tg-tab-tg');
				var btnSplit = bar.querySelector('#tg-tab-split');
				var btnLock = bar.querySelector('#tg-tab-lock');
				var btnNotes = bar.querySelector('#tg-tab-notes');

				if (btnNotes) {
					btnNotes.onclick = function(e) {
						e.preventDefault();
						e.stopPropagation();
						if (window.toggleScratchpad) window.toggleScratchpad();
					};
				}

				if (btnLock) {
					btnLock.onclick = function(e) {
						e.preventDefault();
						e.stopPropagation();
						if (window.lockTelegram) window.lockTelegram();
					};
				}

			if (btnWA) {
				btnWA.onclick = function(e) {
					e.preventDefault();
					e.stopPropagation();
					if (window.switchToWhatsApp) window.switchToWhatsApp();
					else if (window.openWhatsAppWindow) window.openWhatsAppWindow();
				};
			}

			if (btnTG) {
				btnTG.onclick = function(e) {
					e.preventDefault();
					e.stopPropagation();
					if (window.switchToTelegram) window.switchToTelegram();
					else if (window.openTelegramWindow) window.openTelegramWindow();
				};
			}

			if (btnSplit) {
				btnSplit.onclick = function(e) {
					e.preventDefault();
					e.stopPropagation();
					if (window.switchToSplitView) window.switchToSplitView();
					else if (window.openSideBySideView) window.openSideBySideView();
				};
			}

				root.appendChild(bar);
			} else if (!bar.isConnected || bar.parentElement !== root) {
				root.appendChild(bar);
			}
		}

		window.ensureDock = injectTabs;

		window.syncDockActiveTab = function(mode) {
			var bWA = document.getElementById('tg-tab-wa');
			var bTG = document.getElementById('tg-tab-tg');
			var bSp = document.getElementById('tg-tab-split');
			if (bWA) bWA.classList.toggle('active', mode === 'wa');
			if (bTG) bTG.classList.toggle('active', mode === 'tg');
			if (bSp) bSp.classList.toggle('active', mode === 'split');
		};

		// Shortcuts: Ctrl+1 (WhatsApp), Ctrl+2 (Telegram), Ctrl+3 (Split View)
		var handleShortcuts = function(e) {
			if (e.ctrlKey || e.metaKey) {
				if (e.key === '1' || e.code === 'Digit1') {
					if (e.preventDefault) e.preventDefault();
					if (e.stopPropagation) e.stopPropagation();
					if (window.switchToWhatsApp) window.switchToWhatsApp();
					else if (window.openWhatsAppWindow) window.openWhatsAppWindow();
				} else if (e.key === '2' || e.code === 'Digit2') {
					if (e.preventDefault) e.preventDefault();
					if (e.stopPropagation) e.stopPropagation();
					if (window.switchToTelegram) window.switchToTelegram();
					else if (window.openTelegramWindow) window.openTelegramWindow();
				} else if (e.key === '3' || e.code === 'Digit3') {
					if (e.preventDefault) e.preventDefault();
					if (e.stopPropagation) e.stopPropagation();
					if (window.switchToSplitView) window.switchToSplitView();
					else if (window.openSideBySideView) window.openSideBySideView();
				}
			}
		};
		window.addEventListener('keydown', handleShortcuts, true);
		document.addEventListener('keydown', handleShortcuts, true);

		// Initial injection attempts
		if (document.body || document.documentElement) {
			injectTabs();
		}
		document.addEventListener('DOMContentLoaded', injectTabs, { once: true });
		setTimeout(injectTabs, 500);
		setTimeout(injectTabs, 1500);
		setTimeout(injectTabs, 3000);

		// Periodic health-check interval to survive dynamic SPA DOM re-renders
		setInterval(injectTabs, 3000);

		function showTgToast(msg) {
			var root = document.documentElement || document.body;
			if (!root) return;
			var toast = document.getElementById('tg-toast-msg');
			if (!toast) {
				toast = document.createElement('div');
				toast.id = 'tg-toast-msg';
				toast.style.cssText = 'position:fixed;bottom:70px;left:50%;transform:translateX(-50%);background:rgba(23,33,43,0.96);color:#ffffff;padding:8px 18px;border-radius:20px;font-size:12.5px;font-weight:500;box-shadow:0 8px 24px rgba(0,0,0,0.6);border:1px solid rgba(255,255,255,0.12);z-index:2147483647;pointer-events:none;transition:opacity 0.2s ease,transform 0.2s ease;opacity:0;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;';
				root.appendChild(toast);
			}
			toast.textContent = msg;
			toast.style.opacity = '1';
			toast.style.transform = 'translateX(-50%) translateY(0)';
			clearTimeout(toast._timer);
			toast._timer = setTimeout(function() {
				toast.style.opacity = '0';
				toast.style.transform = 'translateX(-50%) translateY(6px)';
			}, 2500);
		}

		// Feature: Quick Replies in Telegram
		(function() {
			var DEFAULT_QUICK_REPLIES = [
				{ key: "/rek", text: "BCA: 1234567890 a/n Akun Bisnis\nMandiri: 0987654321 a/n Akun Bisnis" },
				{ key: "/alamat", text: "Jl. Mawar No. 123, Kel. Sukajadi, Kota Bandung, Jawa Barat 40162" },
				{ key: "/halo", text: "Halo! Terima kasih telah menghubungi kami. Ada yang bisa kami bantu hari ini?" },
				{ key: "/terimakasih", text: "Terima kasih banyak atas pesan dan kerja samanya! Semoga harimu menyenangkan." }
			];

			function loadQuickReplies() {
				try {
					var raw = localStorage.getItem('wa_quick_replies');
					if (raw) {
						var parsed = JSON.parse(raw);
						if (Array.isArray(parsed) && parsed.length > 0) return parsed;
					}
				} catch(e) {}
				return DEFAULT_QUICK_REPLIES.slice();
			}

			function cleanZeroWidth(str) {
				return (str || '').replace(/[\u200B-\u200D\uFEFF\u200E\u200F\u202A-\u202E]/g, '');
			}

			function getRawCharIndex(raw, cleanIndex) {
				var cleanIdx = 0;
				for (var i = 0; i < raw.length; i++) {
					var c = raw.charCodeAt(i);
					var isZW = (c >= 0x200B && c <= 0x200D) || c === 0xFEFF || (c >= 0x200E && c <= 0x200F) || (c >= 0x202A && c <= 0x202E);
					if (!isZW) {
						if (cleanIdx === cleanIndex) return i;
						cleanIdx++;
					}
				}
				return raw.length;
			}

			var tgQRTypedWord = '';

			function getTgEditorUserText(editable) {
				if (!editable) return '';
				return cleanZeroWidth(editable.innerText || editable.textContent || '').trim();
			}

			function placeTgCaretAtEnd(editable) {
				editable.focus();
				var sel = window.getSelection();
				var range = document.createRange();
				var tn = (editable.lastChild && editable.lastChild.nodeType === Node.TEXT_NODE) ? editable.lastChild :
				         (editable.firstChild && editable.firstChild.nodeType === Node.TEXT_NODE) ? editable.firstChild : null;
				if (tn) {
					range.setStart(tn, tn.textContent.length);
					range.collapse(true);
				} else {
					range.selectNodeContents(editable);
					range.collapse(false);
				}
				sel.removeAllRanges();
				sel.addRange(range);
			}

			function applyTgQuickReplyToEditable(editable, matchKey, textToInsert, typedWord) {
				if (!editable) return false;
				var userText = getTgEditorUserText(editable);
				placeTgCaretAtEnd(editable);

				var m = userText.match(/(?:^|\s)(\/[\w-]*)$/);
				var slashCmd = m ? m[1] : (typedWord || matchKey);

				var delCount = slashCmd.length;
				for (var k = 0; k < delCount; k++) {
					document.execCommand('delete');
				}

				document.execCommand('insertText', false, textToInsert);
				editable.dispatchEvent(new InputEvent('input', { bubbles: true, cancelable: true, inputType: 'insertText', data: textToInsert }));
				if (typeof showTgToast === 'function') {
					showTgToast('⚡ Template: ' + matchKey + ' diterapkan');
				}
				return true;
			}

			function applyTgQuickReply(item, target) {
				if (!item || !item.text) return false;
				target = target || document.activeElement;
				if (!target) return false;

				var textToInsert = item.text;
				var isInput = (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA');

				if (isInput) {
					var val = target.value || '';
					var selEnd = target.selectionEnd || target.selectionStart || val.length;
					var before = val.substring(0, selEnd);
					var m = before.match(/(?:^|\s)(\/[\w-]*)$/);
					var slashCmd = m ? m[1] : (tgQRTypedWord || item.key);
					var kIdx = before.lastIndexOf(slashCmd);
					if (kIdx !== -1) {
						target.setRangeText(textToInsert, kIdx, selEnd, 'end');
					} else {
						target.setRangeText(textToInsert, selEnd - slashCmd.length >= 0 ? selEnd - slashCmd.length : 0, selEnd, 'end');
					}
					target.dispatchEvent(new Event('input', { bubbles: true }));
					if (typeof showTgToast === 'function') {
						showTgToast('⚡ Template: ' + item.key + ' diterapkan');
					}
					return true;
				}

				var editable = (target.closest && target.closest('[contenteditable="true"]')) ||
				               (target.isContentEditable ? target : null) ||
				               document.querySelector('.input-message-input[contenteditable="true"]') ||
				               document.querySelector('div[contenteditable="true"]');
				if (!editable) return false;

				return applyTgQuickReplyToEditable(editable, item.key, textToInsert, tgQRTypedWord);
			}

			var tgQRPopup = null;
			var tgQRMatches = [];
			var tgQRSelectedIndex = 0;
			var tgQRTarget = null;

			function ensureTgQRPopup() {
				if (tgQRPopup && tgQRPopup.parentNode) return tgQRPopup;
				tgQRPopup = document.createElement('div');
				tgQRPopup.id = 'tg-qr-popup';
				tgQRPopup.className = 'tg-qr-popup';
				tgQRPopup.style.display = 'none';
				(document.body || document.documentElement).appendChild(tgQRPopup);
				return tgQRPopup;
			}

			function hideTgQRPopup() {
				if (tgQRPopup) {
					tgQRPopup.style.display = 'none';
					tgQRMatches = [];
					tgQRSelectedIndex = 0;
					tgQRTarget = null;
					tgQRTypedWord = '';
				}
			}

			function showTgQRPopup(target, matches, queryWord) {
				if (!matches || matches.length === 0) {
					hideTgQRPopup();
					return;
				}
				tgQRTarget = target;
				tgQRMatches = matches;
				tgQRSelectedIndex = 0;
				tgQRTypedWord = queryWord || '';

				var popup = ensureTgQRPopup();
				var rect = target.getBoundingClientRect();
				var bottomPos = (window.innerHeight - rect.top + 10);
				if (bottomPos < 50) bottomPos = 90;

				var leftPos = rect.left;
				if (leftPos + 390 > window.innerWidth) {
					leftPos = window.innerWidth - 400;
				}
				if (leftPos < 16) leftPos = 16;

				popup.style.bottom = bottomPos + 'px';
				popup.style.left = leftPos + 'px';

				renderTgQRPopup();
				popup.style.display = 'flex';
			}

			function renderTgQRPopup() {
				if (!tgQRPopup) return;
				var html = '<div class="tg-qr-header">' +
					'<div style="display:flex;align-items:center;gap:6px;"><span style="width:6px;height:6px;border-radius:50%;background:#24A1DE;display:inline-block;"></span><b style="color:#ffffff;">Balas Cepat Telegram</b></div>' +
					'<span class="tg-qr-badge">Tekan [Tab] / [Enter] / [Klik]</span>' +
					'</div>' +
					'<div class="tg-qr-list">';

				for (var i = 0; i < tgQRMatches.length; i++) {
					var it = tgQRMatches[i];
					var isAct = (i === tgQRSelectedIndex);
					var cleanSnippet = it.text.replace(/\n/g, ' ').substring(0, 56);
					if (it.text.length > 56) cleanSnippet += '...';
					html += '<div class="tg-qr-item ' + (isAct ? 'active' : '') + '" data-idx="' + i + '">' +
						'<div class="tg-qr-item-key"><span>' + it.key + '</span></div>' +
						'<div class="tg-qr-item-text">' + cleanSnippet + '</div>' +
						'</div>';
				}
				html += '</div>';
				tgQRPopup.innerHTML = html;

				var items = tgQRPopup.querySelectorAll('.tg-qr-item');
				for (var j = 0; j < items.length; j++) {
					(function(idx) {
						var el = items[idx];
						el.addEventListener('mousedown', function(ev) {
							if (ev.preventDefault) ev.preventDefault();
							if (ev.stopPropagation) ev.stopPropagation();
							var chosen = tgQRMatches[idx];
							var tgt = tgQRTarget;
							if (chosen) {
								applyTgQuickReply(chosen, tgt);
							}
							hideTgQRPopup();
						});
					})(j);
				}
			}

			function tryExpandTgQuickReply(e) {
				var key = e.key;
				var code = e.code;
				var isTab = (key === 'Tab' || code === 'Tab' || e.keyCode === 9);
				var isEnter = (key === 'Enter' || code === 'Enter' || e.keyCode === 13);
				var isSpace = (key === ' ' || code === 'Space' || e.keyCode === 32);
				if (!isTab && !isEnter && !isSpace) return false;

				var target = e.target;
				if (!target) return false;

				var replies = loadQuickReplies();
				if (!replies || replies.length === 0) return false;

				var isInput = (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA');
				if (isInput) {
					var start = target.selectionStart || 0;
					var val = target.value || '';
					var textBefore = val.substring(0, start);
					for (var i = 0; i < replies.length; i++) {
						var item = replies[i];
						var kLower = item.key.toLowerCase();
						if (textBefore.toLowerCase().endsWith(kLower)) {
							var prevCharIdx = textBefore.length - kLower.length - 1;
							if (prevCharIdx < 0 || /\s/.test(textBefore.charAt(prevCharIdx))) {
								if (e.preventDefault) e.preventDefault();
								if (e.stopPropagation) e.stopPropagation();
								var keyStart = start - item.key.length;
								var rep = item.text;
								if (isSpace && !rep.endsWith(' ')) rep += ' ';
								target.setRangeText(rep, keyStart, start, 'end');
								target.dispatchEvent(new Event('input', { bubbles: true }));
								if (typeof showTgToast === 'function') {
									showTgToast('⚡ Template: ' + item.key + ' diterapkan');
								}
								return true;
							}
						}
					}
					return false;
				}

				var editable = (target.closest && target.closest('[contenteditable="true"]')) ||
				               (target.isContentEditable ? target : null) ||
				               document.querySelector('.input-message-input[contenteditable="true"]') ||
				               document.querySelector('div[contenteditable="true"]');
				if (!editable) return false;

				var fullText = getTgEditorUserText(editable);
				if (!fullText) return false;

				var matchItem = null;
				for (var j = 0; j < replies.length; j++) {
					var it = replies[j];
					var k = it.key.toLowerCase();
					if (fullText.toLowerCase().endsWith(k)) {
						var idx = fullText.toLowerCase().lastIndexOf(k);
						if (idx === 0 || /\s/.test(fullText.charAt(idx - 1))) {
							matchItem = it;
							break;
						}
					}
				}

				if (!matchItem) return false;

				if (e.preventDefault) e.preventDefault();
				if (e.stopPropagation) e.stopPropagation();

				var textToInsert = matchItem.text;
				if (isSpace && !textToInsert.endsWith(' ')) {
					textToInsert += ' ';
				}

				applyTgQuickReplyToEditable(editable, matchItem.key, textToInsert, matchItem.key);
				return true;
			}

			// Input listener for slash typing in Telegram
			document.addEventListener('input', function(e) {
				var target = e.target;
				if (!target) return;

				var isInput = (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA');
				var isCE = target.isContentEditable || (target.closest && target.closest('[contenteditable="true"]'));
				if (!isInput && !isCE) {
					hideTgQRPopup();
					return;
				}

				var textBefore = '';
				if (isInput) {
					var pos = target.selectionStart || 0;
					textBefore = (target.value || '').substring(0, pos);
				} else {
					var sel = window.getSelection();
					if (sel && sel.rangeCount) {
						var r = sel.getRangeAt(0);
						var node = r.startContainer;
						if (node && node.nodeType === Node.TEXT_NODE) {
							textBefore = cleanZeroWidth(node.textContent.substring(0, r.startOffset));
						}
					}
					if (!textBefore) {
						var host = isCE ? (target.isContentEditable ? target : target.closest('[contenteditable="true"]')) : target;
						textBefore = getTgEditorUserText(host);
					}
				}

				var m = textBefore.match(/(?:^|\s)(\/[\w-]*)$/);
				if (!m) {
					hideTgQRPopup();
					return;
				}

				var query = m[1].toLowerCase();
				var all = loadQuickReplies();
				var matches = all.filter(function(it) {
					return it.key.toLowerCase().startsWith(query);
				});

				if (matches.length > 0) {
					showTgQRPopup(target, matches, m[1]);
				} else {
					hideTgQRPopup();
				}
			}, true);

			// Keydown listener for Telegram keyboard navigation & quick replies
			document.addEventListener('keydown', function(e) {
				var key = e.key;
				var code = e.code;
				var isTab = (key === 'Tab' || code === 'Tab' || e.keyCode === 9);
				var isEnter = (key === 'Enter' || code === 'Enter' || e.keyCode === 13);
				var isEscape = (key === 'Escape' || code === 'Escape' || e.keyCode === 27);
				var isUp = (key === 'ArrowUp' || code === 'ArrowUp' || e.keyCode === 38);
				var isDown = (key === 'ArrowDown' || code === 'ArrowDown' || e.keyCode === 40);
				var isSpace = (key === ' ' || code === 'Space' || e.keyCode === 32);

				if (tgQRPopup && tgQRPopup.style.display !== 'none' && tgQRMatches.length > 0) {
					if (isDown) {
						if (e.preventDefault) e.preventDefault();
						if (e.stopPropagation) e.stopPropagation();
						tgQRSelectedIndex = (tgQRSelectedIndex + 1) % tgQRMatches.length;
						renderTgQRPopup();
						return;
					}
					if (isUp) {
						if (e.preventDefault) e.preventDefault();
						if (e.stopPropagation) e.stopPropagation();
						tgQRSelectedIndex = (tgQRSelectedIndex - 1 + tgQRMatches.length) % tgQRMatches.length;
						renderTgQRPopup();
						return;
					}
					if (isEscape) {
						if (e.preventDefault) e.preventDefault();
						if (e.stopPropagation) e.stopPropagation();
						hideTgQRPopup();
						return;
					}
					if (isTab || isEnter) {
						if (e.preventDefault) e.preventDefault();
						if (e.stopPropagation) e.stopPropagation();
						var selected = tgQRMatches[tgQRSelectedIndex];
						var tgt = tgQRTarget || e.target;
						if (selected) {
							applyTgQuickReply(selected, tgt);
						}
						hideTgQRPopup();
						return;
					}
				}

				if (isTab || isEnter || isSpace) {
					var expanded = tryExpandTgQuickReply(e);
					if (expanded) {
						hideTgQRPopup();
					}
				}
			}, true);

			document.addEventListener('click', function(e) {
				if (tgQRPopup && tgQRPopup.style.display !== 'none') {
					if (!tgQRPopup.contains(e.target)) {
						hideTgQRPopup();
					}
				}
			}, true);
		})();
	})();

	// Feature: Privacy mode (blur) for Telegram. Alt+P or Ctrl+Shift+P toggles it, the dock button does the
	// same and Shift+click cycles the blur strength. Hovering a blurred element reveals it.
	(function() {
		var LEVELS = [3, 5, 8];
		var cfg = { active: false, intensity: 5, contacts: true, preview: true, messages: true, media: true, avatars: false };
		try {
			var saved = localStorage.getItem('tg_privacy_config');
			if (saved) cfg = Object.assign({}, cfg, JSON.parse(saved));
		} catch(e) {}

		function save() {
			try { localStorage.setItem('tg_privacy_config', JSON.stringify(cfg)); } catch(e) {}
		}

		function buildCSS() {
			var P = 'body.tg-privacy-active ';
			var sel = [];
			if (cfg.contacts) sel.push(P + '.chatlist-chat .peer-title', P + '.chatlist-chat .user-title', P + '.chat-info .peer-title', P + '.bubble .peer-title', P + '.bubble .reply-title');
			if (cfg.preview) sel.push(P + '.chatlist-chat .dialog-subtitle');
			if (cfg.messages) sel.push(P + '.bubble-content');
			if (cfg.media) sel.push(P + '.bubble .media-photo', P + '.bubble .media-video', P + '.bubble .album-item-media', P + '.bubble .attachment');
			if (cfg.avatars) sel.push(P + '.chatlist-chat .avatar', P + '.chat-info .avatar', P + '.bubbles-group-avatar-container .avatar');
			if (!sel.length) return '';
			return sel.join(',\n') + ' { filter: blur(' + cfg.intensity + 'px) !important; transition: filter 0.15s ease-in-out !important; }\n' +
				sel.map(function(s) { return s + ':hover'; }).join(',\n') + ' { filter: none !important; }';
		}

		function applyStyle() {
			var root = document.head || document.documentElement;
			if (!root) return;
			var style = document.getElementById('tg-privacy-style');
			if (!style) {
				style = document.createElement('style');
				style.id = 'tg-privacy-style';
				root.appendChild(style);
			}
			var css = buildCSS();
			if (style.textContent !== css) style.textContent = css;
		}

		function syncBody() {
			if (!document.body) return;
			document.body.classList.toggle('tg-privacy-active', !!cfg.active);
			var btn = document.getElementById('tg-tab-privacy');
			if (btn) btn.classList.toggle('active', !!cfg.active);
		}

		function toast(msg) {
			var root = document.documentElement || document.body;
			if (!root) return;
			var t = document.getElementById('tg-toast-msg');
			if (!t) {
				t = document.createElement('div');
				t.id = 'tg-toast-msg';
				t.style.cssText = 'position:fixed;bottom:70px;left:50%;transform:translateX(-50%);background:rgba(23,33,43,0.96);color:#ffffff;padding:8px 18px;border-radius:20px;font-size:12.5px;font-weight:500;box-shadow:0 8px 24px rgba(0,0,0,0.6);border:1px solid rgba(255,255,255,0.12);z-index:2147483647;pointer-events:none;transition:opacity 0.2s ease,transform 0.2s ease;opacity:0;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;';
				root.appendChild(t);
			}
			t.textContent = msg;
			t.style.opacity = '1';
			t.style.transform = 'translateX(-50%) translateY(0)';
			clearTimeout(t._timer);
			t._timer = setTimeout(function() {
				t.style.opacity = '0';
				t.style.transform = 'translateX(-50%) translateY(6px)';
			}, 2200);
		}

		window.toggleTgPrivacy = function(force) {
			cfg.active = typeof force === 'boolean' ? force : !cfg.active;
			save();
			applyStyle();
			syncBody();
			toast(cfg.active ? 'Mode Privasi aktif (' + cfg.intensity + 'px)' : 'Mode Privasi nonaktif');
			return cfg.active;
		};

		window.cycleTgPrivacyIntensity = function() {
			var i = LEVELS.indexOf(cfg.intensity);
			cfg.intensity = LEVELS[(i + 1) % LEVELS.length];
			save();
			applyStyle();
			toast('Intensitas blur: ' + cfg.intensity + 'px');
		};

		function onKey(e) {
			if ((e.altKey || (e.ctrlKey && e.shiftKey)) && (e.code === 'KeyP' || e.key === 'p' || e.key === 'P')) {
				if (e.preventDefault) e.preventDefault();
				if (e.stopPropagation) e.stopPropagation();
				window.toggleTgPrivacy();
			}
		}
		window.addEventListener('keydown', onKey, true);

		document.addEventListener('click', function(e) {
			var btn = e.target && e.target.closest && e.target.closest('#tg-tab-privacy');
			if (!btn) return;
			e.preventDefault();
			e.stopPropagation();
			if (e.shiftKey) window.cycleTgPrivacyIntensity();
			else window.toggleTgPrivacy();
		}, true);

		// The dock is re-rendered by Telegram's SPA, so re-assert state periodically.
		applyStyle();
		syncBody();
		document.addEventListener('DOMContentLoaded', function() { applyStyle(); syncBody(); }, { once: true });
		setInterval(function() { applyStyle(); syncBody(); }, 1000);
	})();

	// Feature: Presence guard. While the app is locked, Telegram Web must not see the page as visible,
	// focused or used, otherwise it keeps reporting the account as online. Runs at document start, before
	// Telegram's own scripts register their listeners. The lock section below toggles it.
	(function() {
		var guarded = false;
		try { guarded = localStorage.getItem('tg_is_locked') === 'true'; } catch(e) {}

		function define(obj, prop, getter) {
			try { Object.defineProperty(obj, prop, { get: getter, configurable: true }); } catch(e) {}
		}
		define(document, 'hidden', function() { return guarded ? true : Document.prototype.__lookupGetter__('hidden').call(document); });
		define(document, 'webkitHidden', function() { return guarded ? true : Document.prototype.__lookupGetter__('webkitHidden').call(document); });
		define(document, 'visibilityState', function() { return guarded ? 'hidden' : Document.prototype.__lookupGetter__('visibilityState').call(document); });
		define(document, 'webkitVisibilityState', function() { return guarded ? 'hidden' : Document.prototype.__lookupGetter__('webkitVisibilityState').call(document); });
		var origHasFocus = Document.prototype.hasFocus;
		document.hasFocus = function() { return guarded ? false : origHasFocus.call(document); };

		// Real input and focus events must not reach Telegram while locked; the lock UI keeps working.
		var BLOCKED = ['focus', 'blur', 'visibilitychange', 'webkitvisibilitychange', 'mousemove', 'mousedown', 'mouseup',
			'click', 'wheel', 'keydown', 'keyup', 'keypress', 'touchstart', 'touchmove', 'pointermove', 'pointerdown'];
		function inLockUI(t) {
			return !!(t && t.nodeType === 1 && t.closest && t.closest('#tg-lock-overlay, #tg-change-pin-modal'));
		}
		BLOCKED.forEach(function(name) {
			window.addEventListener(name, function(e) {
				if (!guarded || !e.isTrusted || inLockUI(e.target)) return;
				e.stopImmediatePropagation();
			}, true);
		});

		window.__tgSetPresenceGuard = function(on) {
			on = !!on;
			if (guarded === on) return;
			guarded = on;
			try {
				document.dispatchEvent(new Event('visibilitychange'));
				window.dispatchEvent(new Event(on ? 'blur' : 'focus'));
			} catch(e) {}
		};
	})();

	// Feature: Telegram App Lock with PIN (Ctrl + L & 5-minute Inactivity Auto-Lock)
	// The PIN is checked by Go (hashed, with a lockout after repeated wrong PINs); this page never sees it.
	(function() {
		var isLocked = false;
		try {
			isLocked = localStorage.getItem('tg_is_locked') === 'true';
			localStorage.removeItem('tg_app_pin'); // left over from when the PIN was kept in the page
		} catch(e) {}

		// Latest status from Go: { isDefault, length, wait, remaining }. null until the first answer.
		var pinInfo = null;
		function refreshPinInfo() {
			if (!window.getAppPinStatus) return Promise.resolve(null);
			return window.getAppPinStatus().then(function(st) {
				pinInfo = st;
				updatePinHints();
				if (st && st.wait > 0) showLockoutCountdown(st.wait);
				return st;
			}).catch(function() { return null; });
		}
		window.onAppPinChanged = refreshPinInfo;
		refreshPinInfo();

		function updatePinHints() {
			var hint = document.getElementById('tg-pin-hint');
			if (hint) hint.textContent = pinInfo && pinInfo.isDefault ? 'PIN awal: 1234 • Ctrl+L untuk mengunci' : 'Ctrl+L untuk mengunci';
		}

		function formatWait(s) {
			return s >= 60 ? Math.floor(s / 60) + ' mnt ' + (s % 60) + ' dtk' : s + ' dtk';
		}

		var countdownTimer = null;
		function showLockoutCountdown(seconds) {
			var input = document.getElementById('tg-pin-input');
			var btn = document.getElementById('tg-pin-unlock-btn');
			var errEl = document.getElementById('tg-pin-error');
			if (!input || !errEl) return;
			clearInterval(countdownTimer);
			var left = seconds;
			function tick() {
				if (left <= 0) {
					clearInterval(countdownTimer);
					input.disabled = false;
					if (btn) btn.disabled = false;
					errEl.textContent = '';
					input.focus();
					return;
				}
				input.disabled = true;
				if (btn) btn.disabled = true;
				errEl.textContent = 'Terlalu banyak PIN salah. Coba lagi dalam ' + formatWait(left) + '.';
				left--;
			}
			tick();
			countdownTimer = setInterval(tick, 1000);
		}

		var inactivityTimer = null;
		var INACTIVITY_TIMEOUT = 5 * 60 * 1000;
		function resetInactivityTimer() {
			clearTimeout(inactivityTimer);
			inactivityTimer = setTimeout(function() {
				window.lockTelegram('auto');
			}, INACTIVITY_TIMEOUT);
		}
		['mousedown', 'keydown', 'touchstart'].forEach(function(evt) {
			window.addEventListener(evt, resetInactivityTimer, { passive: true });
		});
		resetInactivityTimer();

		function showTgToast(msg) {
			var root = document.documentElement || document.body;
			if (!root) return;
			var toast = document.getElementById('tg-toast-msg');
			if (!toast) {
				toast = document.createElement('div');
				toast.id = 'tg-toast-msg';
				toast.style.cssText = 'position:fixed;bottom:70px;left:50%;transform:translateX(-50%);background:rgba(23,33,43,0.96);color:#ffffff;padding:8px 18px;border-radius:20px;font-size:12.5px;font-weight:500;box-shadow:0 8px 24px rgba(0,0,0,0.6);border:1px solid rgba(255,255,255,0.12);z-index:2147483647;pointer-events:none;transition:opacity 0.2s ease,transform 0.2s ease;opacity:0;font-family:-apple-system,BlinkMacSystemFont,\"Segoe UI\",Roboto,sans-serif;';
				root.appendChild(toast);
			}
			toast.textContent = msg;
			toast.style.opacity = '1';
			toast.style.transform = 'translateX(-50%) translateY(0)';
			clearTimeout(toast._timer);
			toast._timer = setTimeout(function() {
				toast.style.opacity = '0';
				toast.style.transform = 'translateX(-50%) translateY(6px)';
			}, 2500);
		}

		function createLockOverlay() {
			var existing = document.getElementById('tg-lock-overlay');
			if (existing) return existing;
			var root = document.documentElement || document.body;
			if (!root) return null;

			var overlay = document.createElement('div');
			overlay.id = 'tg-lock-overlay';
			overlay.style.cssText = 'display:none;position:fixed;inset:0;width:100vw;height:100vh;background:rgba(15,20,28,0.95);backdrop-filter:blur(18px);-webkit-backdrop-filter:blur(18px);z-index:2147483646;align-items:center;justify-content:center;user-select:none;font-family:-apple-system,BlinkMacSystemFont,\"Segoe UI\",Roboto,sans-serif;';
			overlay.innerHTML = '<div style="background:#17212b;border:1px solid rgba(255,255,255,0.1);border-radius:14px;width:330px;padding:30px 26px;text-align:center;box-shadow:0 24px 60px rgba(0,0,0,0.8);display:flex;flex-direction:column;align-items:center;box-sizing:border-box;">' +
				'<div style="width:48px;height:48px;border-radius:50%;background:rgba(36,161,222,0.15);border:1px solid rgba(36,161,222,0.3);display:flex;align-items:center;justify-content:center;color:#24A1DE;margin-bottom:14px;">' +
				'  <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>' +
				'</div>' +
				'<h2 style="margin:0 0 4px 0;font-size:18px;font-weight:600;color:#ffffff;">Telegram Terkunci</h2>' +
				'<p style="margin:0 0 18px 0;font-size:12.5px;color:#708499;">Masukkan PIN untuk membuka akses</p>' +
				'<input id="tg-pin-input" type="password" inputmode="numeric" maxlength="6" style="width:180px;height:42px;text-align:center;font-size:22px;letter-spacing:8px;margin-bottom:12px;background:#0e1621;border:1px solid rgba(255,255,255,0.16);color:#ffffff;border-radius:8px;outline:none;box-sizing:border-box;font-family:inherit;">' +
				'<div id="tg-pin-error" style="color:#ef4444;font-size:12px;min-height:18px;margin-bottom:10px;"></div>' +
				'<button id="tg-pin-unlock-btn" style="width:100%;height:38px;font-size:13px;margin-bottom:12px;background:#24A1DE;color:#ffffff;font-weight:600;border:none;border-radius:8px;cursor:pointer;outline:none;transition:background 0.15s ease;font-family:inherit;">Buka Kunci</button>' +
				'<div id="tg-pin-hint" style="font-size:11px;color:#6c7883;">Ctrl+L untuk mengunci</div>' +
				'<div style="margin-top:10px;"><a id="tg-pin-change-link" href="javascript:void(0)" style="color:#24A1DE;font-size:12px;text-decoration:none;font-weight:500;">Ganti / Ubah PIN</a></div>' +
				'<div style="margin-top:14px;padding-top:12px;border-top:1px solid rgba(255,255,255,0.08);width:100%;display:flex;gap:6px;justify-content:center;">' +
				'  <button id="tg-lock-btn-wa" style="background:#242f3d;color:#ffffff;border:1px solid rgba(0,168,132,0.35);padding:5px 12px;border-radius:14px;font-size:11px;cursor:pointer;outline:none;font-family:inherit;">' +
				'    <span style="color:#00a884;font-weight:bold;">WhatsApp</span> (Ctrl+1)' +
				'  </button>' +
				'</div>' +
				'</div>';
			root.appendChild(overlay);
			updatePinHints();

			var btnLockWA = overlay.querySelector('#tg-lock-btn-wa');
			if (btnLockWA) {
				btnLockWA.onclick = function(e) {
					if (e.preventDefault) e.preventDefault();
					if (e.stopPropagation) e.stopPropagation();
					if (window.switchToWhatsApp) window.switchToWhatsApp();
					else if (window.openWhatsAppWindow) window.openWhatsAppWindow();
				};
			}

			var pinInput = document.getElementById('tg-pin-input');
			var unlockBtn = document.getElementById('tg-pin-unlock-btn');
			var errEl = document.getElementById('tg-pin-error');
			var changeLink = document.getElementById('tg-pin-change-link');

			if (changeLink) {
				changeLink.onclick = function(e) {
					if (e.preventDefault) e.preventDefault();
					window.openChangePinModal();
				};
			}

			var verifying = false;
			function unlock() {
				var val = (pinInput ? pinInput.value : '').trim();
				if (verifying || !val || !window.verifyAppPin) return;
				verifying = true;
				window.verifyAppPin(val).then(function(st) {
					verifying = false;
					pinInfo = st;
					if (st.ok) {
						overlay.style.display = 'none';
						pinInput.value = '';
						errEl.textContent = '';
						try { localStorage.setItem('tg_is_locked', 'false'); } catch(e) {}
						if (window.__tgSetPresenceGuard) window.__tgSetPresenceGuard(false);
						showTgToast('🔓 Telegram terbuka');
						resetInactivityTimer();
						if (st.isDefault) window.openChangePinModal({ forced: true });
						return;
					}
					pinInput.value = '';
					if (st.wait > 0) {
						showLockoutCountdown(st.wait);
						return;
					}
					errEl.textContent = 'PIN salah! Sisa ' + st.remaining + ' percobaan sebelum dijeda.';
					pinInput.focus();
				}).catch(function() {
					verifying = false;
					errEl.textContent = 'Gagal memeriksa PIN. Coba lagi.';
				});
			}

			if (unlockBtn) unlockBtn.onclick = unlock;
			if (pinInput) {
				pinInput.addEventListener('keydown', function(e) {
					if (e.ctrlKey || e.metaKey) {
						if (e.key === '1' || e.code === 'Digit1') {
							if (e.preventDefault) e.preventDefault();
							if (e.stopPropagation) e.stopPropagation();
							if (window.switchToWhatsApp) window.switchToWhatsApp();
							return;
						} else if (e.key === '2' || e.code === 'Digit2') {
							if (e.preventDefault) e.preventDefault();
							if (e.stopPropagation) e.stopPropagation();
							if (window.switchToTelegram) window.switchToTelegram();
							return;
						} else if (e.key === '3' || e.code === 'Digit3') {
							if (e.preventDefault) e.preventDefault();
							if (e.stopPropagation) e.stopPropagation();
							if (window.switchToSplitView) window.switchToSplitView();
							return;
						}
					}
					if (e.key === 'Enter') {
						if (e.preventDefault) e.preventDefault();
						unlock();
						return;
					}
					if (e.stopPropagation) e.stopPropagation();
				});
				['keyup', 'keypress', 'input'].forEach(function(evtName) {
					pinInput.addEventListener(evtName, function(e) {
						if (e.ctrlKey || e.metaKey) return;
						if (e.stopPropagation) e.stopPropagation();
					});
				});
				// Submit as soon as the PIN has its full length, so Enter is optional
				pinInput.addEventListener('input', function(e) {
					var current = (pinInput.value || '').trim();
					if (pinInfo && pinInfo.length && current.length === pinInfo.length) {
						unlock();
					}
				});
			}

			return overlay;
		}

		// Options for the open change-PIN modal: forced (cannot be cancelled), thenLock (lock after saving)
		var changeOpts = {};

		function createChangePinModal() {
			var existing = document.getElementById('tg-change-pin-modal');
			if (existing) return existing;
			var root = document.documentElement || document.body;
			if (!root) return null;

			var modal = document.createElement('div');
			modal.id = 'tg-change-pin-modal';
			modal.style.cssText = 'display:none;position:fixed;inset:0;width:100vw;height:100vh;background:rgba(0,0,0,0.7);backdrop-filter:blur(10px);-webkit-backdrop-filter:blur(10px);z-index:2147483647;align-items:center;justify-content:center;user-select:none;font-family:-apple-system,BlinkMacSystemFont,\"Segoe UI\",Roboto,sans-serif;';
			modal.innerHTML = '<div style="background:#17212b;border:1px solid rgba(255,255,255,0.12);border-radius:14px;width:340px;padding:26px 24px;text-align:center;box-shadow:0 24px 60px rgba(0,0,0,0.85);box-sizing:border-box;">' +
				'<div style="width:40px;height:40px;border-radius:50%;background:rgba(36,161,222,0.15);border:1px solid rgba(36,161,222,0.3);display:flex;align-items:center;justify-content:center;color:#24A1DE;margin:0 auto 12px auto;">' +
				'  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 2l-2 2m-1.5 1.5L10 13l-4 1 1-4 7.5-7.5m1.5-1.5l2-2"/><circle cx="7.5" cy="16.5" r="3.5"/></svg>' +
				'</div>' +
				'<h3 id="tg-cp-title" style="margin:0 0 4px 0;font-size:16px;font-weight:600;color:#ffffff;">Ubah PIN Telegram</h3>' +
				'<p id="tg-cp-desc" style="margin:0 0 16px 0;font-size:12px;color:#708499;">Tentukan 4-6 angka PIN keamanan Anda</p>' +
				'<div id="tg-cp-old-row" style="text-align:left;margin-bottom:10px;">' +
				'<label style="font-size:11px;color:#8293a4;display:block;margin-bottom:4px;">PIN Saat Ini:</label>' +
				'<input id="tg-cp-old" type="password" inputmode="numeric" maxlength="6" placeholder="PIN Lama" style="width:100%;background:#0e1621;border:1px solid rgba(255,255,255,0.14);color:#ffffff;padding:8px 12px;border-radius:6px;font-size:12.5px;outline:none;box-sizing:border-box;">' +
				'</div>' +
				'<div style="text-align:left;margin-bottom:10px;">' +
				'<label style="font-size:11px;color:#8293a4;display:block;margin-bottom:4px;">PIN Baru (4-6 angka):</label>' +
				'<input id="tg-cp-new" type="password" inputmode="numeric" maxlength="6" placeholder="PIN Baru" style="width:100%;background:#0e1621;border:1px solid rgba(255,255,255,0.14);color:#ffffff;padding:8px 12px;border-radius:6px;font-size:12.5px;outline:none;box-sizing:border-box;">' +
				'</div>' +
				'<div style="text-align:left;margin-bottom:12px;">' +
				'<label style="font-size:11px;color:#8293a4;display:block;margin-bottom:4px;">Konfirmasi PIN Baru:</label>' +
				'<input id="tg-cp-confirm" type="password" inputmode="numeric" maxlength="6" placeholder="Ulangi PIN Baru" style="width:100%;background:#0e1621;border:1px solid rgba(255,255,255,0.14);color:#ffffff;padding:8px 12px;border-radius:6px;font-size:12.5px;outline:none;box-sizing:border-box;">' +
				'</div>' +
				'<div id="tg-cp-error" style="color:#ef4444;font-size:12px;min-height:16px;margin-bottom:12px;"></div>' +
				'<div style="display:flex;gap:8px;">' +
				'<button id="tg-cp-cancel-btn" style="flex:1;height:34px;background:#242f3d;color:#ffffff;border:1px solid rgba(255,255,255,0.1);border-radius:6px;font-size:12px;font-weight:500;cursor:pointer;outline:none;">Batal</button>' +
				'<button id="tg-cp-save-btn" style="flex:1;height:34px;background:#24A1DE;color:#ffffff;border:none;border-radius:6px;font-size:12px;font-weight:600;cursor:pointer;outline:none;">Simpan PIN</button>' +
				'</div></div>';
			root.appendChild(modal);

			var oldInput = document.getElementById('tg-cp-old');
			var newInput = document.getElementById('tg-cp-new');
			var confirmInput = document.getElementById('tg-cp-confirm');
			var errEl = document.getElementById('tg-cp-error');
			var cancelBtn = document.getElementById('tg-cp-cancel-btn');
			var saveBtn = document.getElementById('tg-cp-save-btn');

			function close() {
				if (changeOpts.forced) return;
				modal.style.display = 'none';
			}

			var saving = false;
			function save() {
				var needOld = !(pinInfo && pinInfo.isDefault);
				var oldVal = oldInput ? oldInput.value.trim() : '';
				var newVal = newInput ? newInput.value.trim() : '';
				var confirmVal = confirmInput ? confirmInput.value.trim() : '';

				if (needOld && !oldVal) {
					errEl.textContent = 'Masukkan PIN lama!';
					if (oldInput) oldInput.focus();
					return;
				}
				if (!/^\d{4,6}$/.test(newVal)) {
					errEl.textContent = 'PIN baru harus 4-6 angka!';
					if (newInput) newInput.focus();
					return;
				}
				if (newVal === '1234') {
					errEl.textContent = 'Jangan pakai 1234, pilih PIN lain!';
					if (newInput) { newInput.value = ''; newInput.focus(); }
					return;
				}
				if (newVal !== confirmVal) {
					errEl.textContent = 'Konfirmasi PIN tidak cocok!';
					if (confirmInput) { confirmInput.value = ''; confirmInput.focus(); }
					return;
				}
				if (saving || !window.changeAppPin) return;
				saving = true;

				window.changeAppPin(needOld ? oldVal : '', newVal).then(function(st) {
					saving = false;
					pinInfo = st;
					if (!st.ok) {
						if (st.wait > 0) {
							errEl.textContent = 'Terlalu banyak PIN salah. Coba lagi dalam ' + formatWait(st.wait) + '.';
						} else if (st.error === 'wrong') {
							errEl.textContent = 'PIN lama salah! Sisa ' + st.remaining + ' percobaan sebelum dijeda.';
						} else {
							errEl.textContent = 'Gagal menyimpan PIN. Coba lagi.';
						}
						if (oldInput) { oldInput.value = ''; oldInput.focus(); }
						return;
					}
					var thenLock = changeOpts.thenLock;
					changeOpts = {};
					modal.style.display = 'none';
					oldInput.value = '';
					newInput.value = '';
					confirmInput.value = '';
					errEl.textContent = '';
					updatePinHints();
					showTgToast('✅ PIN berhasil diubah!');
					if (thenLock) window.lockTelegram();
				}).catch(function() {
					saving = false;
					errEl.textContent = 'Gagal menyimpan PIN. Coba lagi.';
				});
			}

			if (saveBtn) saveBtn.onclick = save;
			if (cancelBtn) cancelBtn.onclick = close;
			[oldInput, newInput, confirmInput].forEach(function(inp) {
				if (inp) {
					['keydown', 'keyup', 'keypress', 'input'].forEach(function(evtName) {
						inp.addEventListener(evtName, function(e) {
							if (e.stopPropagation) e.stopPropagation();
						});
					});
					inp.addEventListener('keydown', function(e) {
						if (e.key === 'Enter') {
							if (e.preventDefault) e.preventDefault();
							save();
						}
						if (e.key === 'Escape') {
							if (e.preventDefault) e.preventDefault();
							close();
						}
					});
				}
			});

			return modal;
		}

		// opts.forced: right after unlocking with the default PIN, a new PIN must be set before continuing.
		// opts.thenLock: the user asked to lock while the default PIN was still set; lock once a PIN exists.
		window.openChangePinModal = function(opts) {
			changeOpts = opts || {};
			var modal = createChangePinModal();
			if (!modal) return;
			var isDefault = !!(pinInfo && pinInfo.isDefault);
			var title = document.getElementById('tg-cp-title');
			var desc = document.getElementById('tg-cp-desc');
			var oldRow = document.getElementById('tg-cp-old-row');
			var cancelBtn = document.getElementById('tg-cp-cancel-btn');
			if (title) title.textContent = isDefault ? 'Buat PIN Sendiri' : 'Ubah PIN Telegram';
			if (desc) desc.textContent = isDefault ?
				'PIN masih 1234 (bawaan) yang bisa ditebak siapa saja. Buat PIN 4-6 angka sendiri.' :
				'Tentukan 4-6 angka PIN keamanan Anda';
			if (oldRow) oldRow.style.display = isDefault ? 'none' : '';
			if (cancelBtn) cancelBtn.style.display = changeOpts.forced ? 'none' : '';
			modal.style.display = 'flex';
			var errEl = document.getElementById('tg-cp-error');
			if (errEl) errEl.textContent = '';
			['tg-cp-old', 'tg-cp-new', 'tg-cp-confirm'].forEach(function(id) {
				var el = document.getElementById(id);
				if (el) el.value = '';
			});
			var first = document.getElementById(isDefault ? 'tg-cp-new' : 'tg-cp-old');
			if (first) setTimeout(function() { first.focus(); }, 50);
		};

		// reason: 'auto' (inactivity), 'restore' (was locked when the app closed), or nothing for an
		// explicit lock (Ctrl+L, dock button, tray).
		window.lockTelegram = function(reason) {
			if (pinInfo && pinInfo.isDefault && reason !== 'restore') {
				// A lock that anyone can open with 1234 protects nothing
				if (reason === 'auto') return;
				showTgToast('Buat PIN sendiri dulu sebelum mengunci');
				window.openChangePinModal({ thenLock: true });
				return;
			}
			var overlay = createLockOverlay();
			if (overlay) {
				if (overlay.style.display === 'flex') {
					return;
				}
				overlay.style.display = 'flex';
				try { localStorage.setItem('tg_is_locked', 'true'); } catch(e) {}
				if (window.__tgSetPresenceGuard) window.__tgSetPresenceGuard(true);
				var input = document.getElementById('tg-pin-input');
				if (input) {
					input.value = '';
					setTimeout(function() {
						if (document.activeElement !== input) {
							input.focus();
						}
					}, 50);
				}
				refreshPinInfo();
			}
		};

		window.addEventListener('keydown', function(e) {
			if ((e.ctrlKey || e.metaKey) && (e.code === 'KeyL' || e.key === 'l' || e.key === 'L')) {
				if (e.preventDefault) e.preventDefault();
				if (e.stopPropagation) e.stopPropagation();
				window.lockTelegram();
			}
		}, true);

		document.addEventListener('keydown', function(e) {
			if ((e.ctrlKey || e.metaKey) && (e.code === 'KeyL' || e.key === 'l' || e.key === 'L')) {
				if (e.preventDefault) e.preventDefault();
				if (e.stopPropagation) e.stopPropagation();
				window.lockTelegram();
			}
		}, true);

		if (isLocked) {
			if (document.body || document.documentElement) {
				window.lockTelegram('restore');
			} else {
				document.addEventListener('DOMContentLoaded', function() {
					window.lockTelegram('restore');
				}, { once: true });
			}
		}
	})();


	// Feature: unread count for the taskbar badge and tray tooltip. Telegram Web reports it through the
	// Badging API (navigator.setAppBadge, as for an installed web app); the "(N)" title prefix is a fallback.
	(function() {
		var last = -1;
		function report(n) {
			n = Math.max(0, parseInt(n, 10) || 0);
			if (n === last) return;
			last = n;
			if (window.onUnreadCountChanged) window.onUnreadCountChanged(n);
		}
		var badgeSeen = false;
		try {
			navigator.setAppBadge = function(n) { badgeSeen = true; report(n === undefined ? 1 : n); return Promise.resolve(); };
			navigator.clearAppBadge = function() { badgeSeen = true; report(0); return Promise.resolve(); };
		} catch(e) {}
		setInterval(function() {
			if (badgeSeen) return;
			var m = (document.title || '').match(/^\((\d+)\+?\)/);
			report(m ? m[1] : 0);
		}, 2000);
	})();

	// Toast used by the shared notes drawer (notes.js)
	window.tgShowToast = function(msg) {
		var root = document.documentElement || document.body;
		if (!root) return;
		var toast = document.getElementById('tg-toast-msg');
		if (!toast) {
			toast = document.createElement('div');
			toast.id = 'tg-toast-msg';
			toast.style.cssText = 'position:fixed;bottom:70px;left:50%;transform:translateX(-50%);background:rgba(23,33,43,0.96);color:#ffffff;padding:8px 18px;border-radius:20px;font-size:12.5px;font-weight:500;box-shadow:0 8px 24px rgba(0,0,0,0.6);border:1px solid rgba(255,255,255,0.12);z-index:2147483647;pointer-events:none;transition:opacity 0.2s ease,transform 0.2s ease;opacity:0;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;';
			root.appendChild(toast);
		}
		toast.textContent = msg;
		toast.style.opacity = '1';
		clearTimeout(toast._timer);
		toast._timer = setTimeout(function() { toast.style.opacity = '0'; }, 2500);
	};

	// Feature: chat to a phone number that is not in the contacts (Ctrl + N). Telegram only finds an
	// account by number when its owner allows that in their privacy settings.
	(function() {
		function normalize(raw) {
			var cleaned = String(raw || '').replace(/[^\d+]/g, '');
			if (cleaned.charAt(0) === '+') cleaned = cleaned.substring(1);
			else if (cleaned.charAt(0) === '0') cleaned = '62' + cleaned.substring(1);
			return cleaned;
		}

		function createModal() {
			var existing = document.getElementById('tg-direct-chat-modal');
			if (existing) return existing;
			var root = document.body || document.documentElement;
			if (!root) return null;
			var modal = document.createElement('div');
			modal.id = 'tg-direct-chat-modal';
			modal.style.cssText = 'display:none;position:fixed;inset:0;background:rgba(0,0,0,0.65);backdrop-filter:blur(8px);-webkit-backdrop-filter:blur(8px);z-index:2147483646;align-items:center;justify-content:center;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;';
			modal.innerHTML = '<div style="background:#17212b;border:1px solid rgba(255,255,255,0.12);border-radius:14px;width:340px;padding:24px;box-shadow:0 24px 60px rgba(0,0,0,0.85);box-sizing:border-box;">' +
				'<h3 style="margin:0 0 4px 0;font-size:16px;font-weight:600;color:#ffffff;">Chat ke nomor baru</h3>' +
				'<p style="margin:0 0 14px 0;font-size:12px;color:#708499;">Tanpa menyimpan kontak. 0812… otomatis menjadi 62812…</p>' +
				'<input id="tg-direct-phone" type="text" inputmode="tel" placeholder="Contoh: 08123456789 atau +628…" style="width:100%;height:36px;background:#0e1621;border:1px solid rgba(255,255,255,0.14);color:#ffffff;padding:0 12px;border-radius:8px;font-size:13px;outline:none;box-sizing:border-box;margin-bottom:6px;">' +
				'<div id="tg-direct-error" style="color:#ef4444;font-size:12px;min-height:16px;margin-bottom:10px;"></div>' +
				'<div style="display:flex;gap:8px;">' +
				'<button id="tg-direct-cancel" style="flex:1;height:34px;background:#242f3d;color:#ffffff;border:1px solid rgba(255,255,255,0.1);border-radius:8px;font-size:12px;cursor:pointer;">Batal</button>' +
				'<button id="tg-direct-open" style="flex:1;height:34px;background:#24A1DE;color:#ffffff;border:none;border-radius:8px;font-size:12px;font-weight:600;cursor:pointer;">Buka Chat</button>' +
				'</div></div>';
			root.appendChild(modal);

			var input = modal.querySelector('#tg-direct-phone');
			var err = modal.querySelector('#tg-direct-error');
			function close() { modal.style.display = 'none'; }
			function submit() {
				var num = normalize(input.value);
				if (num.length < 7) { err.textContent = 'Nomor telepon tidak valid!'; input.focus(); return; }
				close();
				input.value = '';
				window.tgShowToast('Membuka chat: +' + num + '…');
				// Telegram Web K opens tg:// links passed as tgaddr; resolve?phone= looks the number up.
				location.hash = '#?tgaddr=' + encodeURIComponent('tg://resolve?phone=' + num);
			}
			modal.querySelector('#tg-direct-open').onclick = submit;
			modal.querySelector('#tg-direct-cancel').onclick = close;
			modal.onclick = function(e) { if (e.target === modal) close(); };
			['keydown', 'keyup', 'keypress', 'input'].forEach(function(evt) {
				input.addEventListener(evt, function(e) { e.stopPropagation(); });
			});
			input.addEventListener('keydown', function(e) {
				if (e.key === 'Enter') { e.preventDefault(); submit(); }
				if (e.key === 'Escape') { e.preventDefault(); close(); }
			});
			return modal;
		}

		window.openDirectChatModal = function() {
			var modal = createModal();
			if (!modal) return;
			modal.style.display = 'flex';
			var err = modal.querySelector('#tg-direct-error');
			if (err) err.textContent = '';
			var input = modal.querySelector('#tg-direct-phone');
			if (input) setTimeout(function() { input.focus(); }, 50);
		};

		window.addEventListener('keydown', function(e) {
			if ((e.ctrlKey || e.metaKey) && !e.shiftKey && !e.altKey && (e.code === 'KeyN' || e.key === 'n' || e.key === 'N')) {
				e.preventDefault();
				e.stopPropagation();
				window.openDirectChatModal();
			}
		}, true);
	})();
