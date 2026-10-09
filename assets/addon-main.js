		// UserAgent override
		Object.defineProperty(navigator, 'userAgent', {
			get: () => '__WA_USER_AGENT__'
		});
		Object.defineProperty(navigator, 'appVersion', {
			get: () => '__WA_USER_AGENT__'
		});

		// Native Notification Polyfill for Windows Desktop Toast
		(function() {
			window.Notification = function(title, options) {
				options = options || {};
				var body = options.body || '';
				if (window.sendNativeNotification) {
					window.sendNativeNotification(title, body);
				}
				this.title = title;
				this.onclick = null;
				this.onclose = null;
				this.onerror = null;
				this.onshow = null;
			};
			window.Notification.permission = 'granted';
			window.Notification.requestPermission = function(callback) {
				var p = Promise.resolve('granted');
				if (typeof callback === 'function') {
					callback('granted');
				}
				return p;
			};
		})();

		// External link protection: Open external links in default system browser
		(function() {
			var origOpen = window.open;
			window.open = function(url) {
				if (url && typeof url === 'string') {
					try {
						var u = new URL(url, window.location.href);
						if (!isWhatsAppInternalUrl(u) && (u.protocol === 'http:' || u.protocol === 'https:')) {
							if (window.openExternalLink) {
								window.openExternalLink(u.href);
							}
							return null;
						}
					} catch(e) {}
				}
				return origOpen.apply(this, arguments);
			};

			function isWhatsAppInternalUrl(u) {
				if (!u) return true;
				if (u.protocol === 'blob:' || u.protocol === 'data:' || u.protocol === 'javascript:') return true;
				var host = (u.hostname || '').toLowerCase();
				if (!host) return true;
				if (host === 'web.whatsapp.com' || host.endsWith('.whatsapp.com') || host.endsWith('.whatsapp.net') || host.endsWith('.fbcdn.net') || host.endsWith('.facebook.com')) {
					return true;
				}
				return false;
			}

			document.addEventListener('click', function(e) {
				var el = e.target;
				while (el && el.tagName !== 'A') {
					el = el.parentElement;
				}
				if (!el || !el.href) return;
				if (el.hasAttribute('download') || el.download) return; // Allow native downloads!
				try {
					var u = new URL(el.href, window.location.href);
					if (!isWhatsAppInternalUrl(u) && (u.protocol === 'http:' || u.protocol === 'https:')) {
						e.preventDefault();
						e.stopPropagation();
						if (window.openExternalLink) {
							window.openExternalLink(u.href);
						}
					}
				} catch(err) {}
			}, true);
		})();

		// Feature 2: Unread Message Observer & Taskbar Flash Trigger
		(function() {
			var lastUnread = -1;
			function checkTitle() {
				var title = document.title || '';
				var match = title.match(/^\((\d+)\+?\)/);
				var count = match ? parseInt(match[1], 10) : 0;
				if (count !== lastUnread) {
					lastUnread = count;
					if (window.onUnreadCountChanged) {
						window.onUnreadCountChanged(count);
					}
				}
			}

			var observer = new MutationObserver(checkTitle);
			function attachTitleObserver() {
				var titleEl = document.querySelector('title');
				if (titleEl) {
					observer.observe(titleEl, { subtree: true, characterData: true, childList: true });
					checkTitle();
				} else {
					setTimeout(attachTitleObserver, 500);
				}
			}
			attachTitleObserver();
		})();

		// Helper: Run callback when DOM (document.body) is available (guaranteed single execution)
		function whenDOMReady(fn) {
			if (document.body) {
				fn();
				return;
			}
			var executed = false;
			var iv = null;
			function trigger() {
				if (executed) return;
				executed = true;
				if (iv) {
					clearInterval(iv);
					iv = null;
				}
				document.removeEventListener('DOMContentLoaded', trigger);
				fn();
			}
			document.addEventListener('DOMContentLoaded', trigger, { once: true });
			iv = setInterval(function() {
				if (document.body) {
					trigger();
				}
			}, 50);
			setTimeout(function() {
				if (iv) {
					clearInterval(iv);
					iv = null;
				}
			}, 15000);
		}

		// Feature: Modular Privacy Engine (Anti-Pusing / Custom Blur)
		var privacyConfig = {
			active: false,
			blurContacts: true,
			blurPreview: true,
			blurMessages: true,
			blurMedia: true,
			blurAvatars: false,
			blurIntensity: 5
		};

		try {
			var savedCfg = localStorage.getItem('wa_privacy_config');
			if (savedCfg) {
				privacyConfig = Object.assign({}, privacyConfig, JSON.parse(savedCfg));
			}
		} catch(e) {}

		function savePrivacyConfig() {
			try {
				localStorage.setItem('wa_privacy_config', JSON.stringify(privacyConfig));
			} catch(e) {}
			updatePrivacyStyles();
		}

				// Impeccable Design System Styles (Custom Scrollbars, Switches, Tabs, Animations)
				// Impeccable Design System Styles (Tokens, Scrollbars, Switches, Dock, Modal, Toast)
		function injectImpeccableStyles() {
			if (!document.head && !document.body) return;
			if (document.getElementById('wa-impeccable-styles')) return;
			var style = document.createElement('style');
			style.id = 'wa-impeccable-styles';
			style.textContent = WA_ADDON_CSS;
			var target = document.head || document.documentElement || document.body;
			if (target) target.appendChild(style);
		}

		function updatePrivacyStyles() {
			if (!document.body && !document.head) return;
			var style = document.getElementById('wa-privacy-style');
			if (!style) {
				style = document.createElement('style');
				style.id = 'wa-privacy-style';
				var target = document.head || document.documentElement || document.body;
				if (target) target.appendChild(style);
			}
			var px = (privacyConfig.blurIntensity || 4) + 'px';
			var selectors = [];

			if (privacyConfig.blurContacts) {
					selectors.push(
						'body.wa-privacy-active #pane-side [data-testid="cell-frame-title"]',
						'body.wa-privacy-active #main header [data-testid="conversation-info-header-chat-title-name"]',
						'body.wa-privacy-active #main [data-testid="author"]'
					);
				}
				if (privacyConfig.blurPreview) {
					selectors.push(
						'body.wa-privacy-active #pane-side [data-testid="cell-frame-secondary"]'
					);
				}
				if (privacyConfig.blurMessages) {
					selectors.push(
						'body.wa-privacy-active #main [data-testid="msg-container"]',
						'body.wa-privacy-active #main .message-in',
						'body.wa-privacy-active #main .message-out'
					);
				}
				if (privacyConfig.blurMedia) {
					selectors.push(
						'body.wa-privacy-active #main [data-testid="image-thumb"]',
						'body.wa-privacy-active #main [data-testid="video-content"]',
						'body.wa-privacy-active #main [data-testid="msg-container"] img:not(.emoji)',
						'body.wa-privacy-active #main [data-testid="msg-container"] video'
					);
				}
				if (privacyConfig.blurAvatars) {
					selectors.push(
						'body.wa-privacy-active #pane-side [role="row"] img:not(.emoji)',
						'body.wa-privacy-active #main header img:not(.emoji)'
					);
				}

				if (selectors.length === 0) {
				style.textContent = '';
				return;
			}

			var hoverSelectors = selectors.map(function(s) {
				return s + ':hover';
			});

			style.textContent = [
				selectors.join(',\n') + ' {',
				'  filter: blur(' + px + ') !important;',
				'  transition: filter 0.15s ease-in-out !important;',
				'}',
				hoverSelectors.join(',\n') + ' {',
				'  filter: none !important;',
				'}',
				'#wa-addon-modal, #wa-addon-modal *, #wa-addon-dock-btn, #wa-addon-dock-btn *, #privacy-mode-toast, #wa-addon-toast, #wa-direct-chat-modal, #wa-lock-overlay, #wa-change-pin-modal {',
				'  filter: none !important;',
				'}'
			].join('\n');
		}

		whenDOMReady(function() {
			injectImpeccableStyles();
			updatePrivacyStyles();
			if (privacyConfig.active) {
				document.body.classList.add('wa-privacy-active');
			}
		});

		function showPrivacyToast(active) {
			if (!document.body) return;
			var toast = document.getElementById('privacy-mode-toast');
			if (!toast) {
				toast = document.createElement('div');
				toast.id = 'privacy-mode-toast';
				toast.className = 'wa-toast';
				document.body.appendChild(toast);
			}
			var label = active ? ('Mode Privasi Aktif (' + privacyConfig.blurIntensity + 'px)') : 'Mode Privasi Nonaktif';
			toast.innerHTML = '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#00a884" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg><span>' + label + '</span>';
			toast.classList.add('show');
			clearTimeout(toast._timer);
			toast._timer = setTimeout(function() {
				toast.classList.remove('show');
			}, 2200);
		}

		window.togglePrivacyMode = function(forceState) {
			if (!document.body) return false;
			var isActive = typeof forceState === 'boolean' ? forceState : !document.body.classList.contains('wa-privacy-active');
			if (isActive) {
				document.body.classList.add('wa-privacy-active');
			} else {
				document.body.classList.remove('wa-privacy-active');
			}
			privacyConfig.active = isActive;
			savePrivacyConfig();
			showPrivacyToast(isActive);
			if (window.onPrivacyModeToggled) {
				try { window.onPrivacyModeToggled(isActive); } catch(e) {}
			}
			var toggleCb = document.getElementById('wa-cc-privacy-toggle');
			if (toggleCb) toggleCb.checked = isActive;
			return isActive;
		};

		window.addEventListener('keydown', function(e) {
			if (window.self !== window.top) return;
			if ((e.altKey && (e.code === 'KeyP' || e.key === 'p' || e.key === 'P')) ||
				(e.ctrlKey && e.shiftKey && (e.code === 'KeyP' || e.key === 'p' || e.key === 'P'))) {
				if (e.preventDefault) e.preventDefault();
				if (e.stopPropagation) e.stopPropagation();
				window.togglePrivacyMode();
			}
		}, true);

		// Global Toast Notification Helper
		function showAddonToast(msg) {
			if (!document.body) return;
			var toast = document.getElementById('wa-addon-toast');
			if (!toast) {
				toast = document.createElement('div');
				toast.id = 'wa-addon-toast';
				toast.className = 'wa-toast';
				document.body.appendChild(toast);
			}
			var cleanMsg = (msg || '').replace(/[\u{1F300}-\u{1F9FF}]|[\u{2600}-\u{26FF}]|[\u{2700}-\u{27BF}]/gu, '').trim();
			toast.innerHTML = '<span class="wa-status-dot" style="flex-shrink:0;"></span><span>' + cleanMsg + '</span>';
			toast.classList.add('show');
			clearTimeout(toast._timer);
			toast._timer = setTimeout(function() {
				toast.classList.remove('show');
			}, 2200);
		}

		// Feature: Direct Chat by Number (Ctrl + N)
		var openDirectChatByNumber;
		(function() {
			if (window.self !== window.top) return;

			function createDirectChatModal() {
				if (!document.body || document.getElementById('wa-direct-chat-modal')) return;
				var modal = document.createElement('div');
				modal.id = 'wa-direct-chat-modal';
				modal.className = 'wa-modal-backdrop';
				modal.innerHTML = '<div class="wa-modal-box" style="width:380px;">' +
					'<div class="wa-modal-header" style="padding:14px 18px;">' +
					'  <div class="wa-header-left">' +
					'    <div class="wa-header-icon" style="width:32px;height:32px;">' +
					'      <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"/></svg>' +
					'    </div>' +
					'    <div>' +
					'      <h3 class="wa-header-title" style="font-size:14px;">Chat ke Nomor Baru</h3>' +
					'      <p class="wa-header-subtitle" style="font-size:11.5px;">Kirim pesan tanpa simpan nomor kontak</p>' +
					'    </div>' +
					'  </div>' +
					'</div>' +
					'<div style="padding:18px 20px;">' +
					'  <input id="wa-direct-phone-input" class="wa-input" type="text" placeholder="Contoh: 08123456789 atau +628..." style="width:100%;margin-bottom:14px;">' +
					'  <div style="display:flex;justify-content:flex-end;gap:8px;">' +
					'    <button id="wa-direct-cancel-btn" class="wa-btn wa-btn-secondary">Batal</button>' +
					'    <button id="wa-direct-submit-btn" class="wa-btn wa-btn-primary">Buka Chat</button>' +
					'  </div>' +
					'</div></div>';
				document.body.appendChild(modal);

				var input = document.getElementById('wa-direct-phone-input');
				var cancelBtn = document.getElementById('wa-direct-cancel-btn');
				var submitBtn = document.getElementById('wa-direct-submit-btn');

				function submit() {
					var raw = input ? input.value.trim() : '';
					if (!raw) return;
					var cleaned = raw.replace(/[^\d+]/g, '');
					if (cleaned.startsWith('+')) {
						cleaned = cleaned.substring(1);
					} else if (cleaned.startsWith('0')) {
						cleaned = '62' + cleaned.substring(1);
					}
					if (cleaned.length < 7) {
						alert('Nomor telepon tidak valid!');
						return;
					}
					modal.style.display = 'none';
					input.value = '';
					showAddonToast('Membuka chat: +' + cleaned + '...');
					window.location.href = 'https://web.whatsapp.com/send?phone=' + cleaned;
				}

				if (submitBtn) submitBtn.onclick = submit;
				if (cancelBtn) cancelBtn.onclick = function() { modal.style.display = 'none'; };
				if (input) {
					['keydown', 'keyup', 'keypress', 'input'].forEach(function(evtName) {
						input.addEventListener(evtName, function(e) {
							if (e.stopPropagation) e.stopPropagation();
						});
					});
					input.onkeydown = function(e) {
						if (e.key === 'Enter') submit();
						if (e.key === 'Escape') modal.style.display = 'none';
					};
				}
				modal.onclick = function(e) {
					if (e.target === modal) modal.style.display = 'none';
				};
			}

			openDirectChatByNumber = function(raw) {
				if (!raw) return;
				var cleaned = raw.replace(/[^\d+]/g, '');
				if (cleaned.startsWith('+')) {
					cleaned = cleaned.substring(1);
				} else if (cleaned.startsWith('0')) {
					cleaned = '62' + cleaned.substring(1);
				}
				if (cleaned.length < 7) {
					alert('Nomor telepon tidak valid!');
					return;
				}
				showAddonToast('Membuka chat: +' + cleaned + '...');
				window.location.href = 'https://web.whatsapp.com/send?phone=' + cleaned;
			};

			window.openDirectChatModal = function() {
				whenDOMReady(function() {
					createDirectChatModal();
					var modal = document.getElementById('wa-direct-chat-modal');
					var input = document.getElementById('wa-direct-phone-input');
					if (modal) {
						modal.style.display = 'flex';
						setTimeout(function() { if (input) input.focus(); }, 50);
					}
				});
			};

			window.addEventListener('keydown', function(e) {
				if ((e.ctrlKey || e.metaKey) && (e.code === 'KeyN' || e.key === 'n' || e.key === 'N') && !e.shiftKey) {
					if (e.preventDefault) e.preventDefault();
					if (e.stopPropagation) e.stopPropagation();
					window.openDirectChatModal();
				}
			}, true);
		})();

		// Feature: Zoom Controls (Ctrl + / Ctrl - / Ctrl 0)
		var currentZoom = 100;
		var applyZoom;
		(function() {
			if (window.self !== window.top) return;

			try {
				currentZoom = parseInt(localStorage.getItem('wa_zoom_level') || '100', 10);
				if (isNaN(currentZoom) || currentZoom < 75 || currentZoom > 150) currentZoom = 100;
			} catch(e) {}

			applyZoom = function(val) {
				currentZoom = Math.min(150, Math.max(75, val));
				if (document.body) {
					document.body.style.zoom = currentZoom + '%';
				}
				try {
					localStorage.setItem('wa_zoom_level', currentZoom.toString());
				} catch(e) {}
				var label = document.getElementById('wa-cc-zoom-label');
				if (label) label.textContent = currentZoom + '%';
				showAddonToast('🔍 Zoom: ' + currentZoom + '%');
			};

			whenDOMReady(function() {
				if (document.body) {
					document.body.style.zoom = currentZoom + '%';
				}
			});

			window.addEventListener('keydown', function(e) {
				if (e.ctrlKey || e.metaKey) {
					if (e.key === '=' || e.key === '+' || e.code === 'Equal') {
						if (e.preventDefault) e.preventDefault();
						if (e.stopPropagation) e.stopPropagation();
						applyZoom(currentZoom + 5);
					} else if (e.key === '-' || e.key === '_' || e.code === 'Minus') {
						if (e.preventDefault) e.preventDefault();
						if (e.stopPropagation) e.stopPropagation();
						applyZoom(currentZoom - 5);
					} else if (e.key === '0' || e.code === 'Digit0' || e.code === 'Numpad0') {
						if (e.preventDefault) e.preventDefault();
						if (e.stopPropagation) e.stopPropagation();
						applyZoom(100);
					}
				}
			}, true);
		})();

		// Feature: Eye Comfort, Night Mode (Sepia) & OLED Pure Black & Compact Chat List
		var eyeComfortConfig = {
			theme: 'default', // 'default', 'oled', 'warm'
			compact: false
		};
		try {
			var savedEye = localStorage.getItem('wa_eye_comfort_config');
			if (savedEye) eyeComfortConfig = Object.assign({}, eyeComfortConfig, JSON.parse(savedEye));
		} catch(e) {}

		function saveEyeComfortConfig() {
			try { localStorage.setItem('wa_eye_comfort_config', JSON.stringify(eyeComfortConfig)); } catch(e) {}
			updateEyeComfortStyles();
		}

		function updateEyeComfortStyles() {
			if (!document.body && !document.head) return;
			var style = document.getElementById('wa-eye-comfort-style');
			if (!style) {
				style = document.createElement('style');
				style.id = 'wa-eye-comfort-style';
				var target = document.head || document.documentElement || document.body;
				if (target) target.appendChild(style);
			}
			var lines = [];
			if (eyeComfortConfig.theme === 'oled') {
				lines.push('body, #app, .app, [data-theme="dark"], #main, header, footer { background: #000000 !important; background-color: #000000 !important; }');
				lines.push('div[data-asset-chat-background-dark] { opacity: 0.02 !important; }');
				lines.push('#pane-side, div[role="region"], div[role="navigation"] { background-color: #000000 !important; }');
				lines.push('.message-in, div[class*="message-in"] { background-color: #111111 !important; }');
				lines.push('.message-out, div[class*="message-out"] { background-color: #003e2c !important; }');
			} else if (eyeComfortConfig.theme === 'warm') {
				lines.push('html { filter: sepia(0.25) saturate(0.92) brightness(0.96) !important; }');
			}
			if (eyeComfortConfig.compact) {
				lines.push('div[role="listitem"] > div { padding-top: 4px !important; padding-bottom: 4px !important; min-height: 52px !important; }');
				lines.push('div[role="listitem"] { margin-bottom: 0px !important; }');
			}
			style.textContent = lines.join('\n');
		}

		whenDOMReady(function() {
			updateEyeComfortStyles();
		});

		// Feature: Custom Quick Replies Manager
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

		function saveQuickReplies(list) {
			try { localStorage.setItem('wa_quick_replies', JSON.stringify(list)); } catch(e) {}
		}

		function addQuickReply(key, text) {
			key = (key || '').trim();
			text = (text || '').trim();
			if (!key || !text) return false;
			if (!key.startsWith('/')) key = '/' + key;
			var list = loadQuickReplies();
			list = list.filter(function(it) { return it.key.toLowerCase() !== key.toLowerCase(); });
			list.unshift({ key: key, text: text });
			saveQuickReplies(list);
			return true;
		}

		function deleteQuickReply(key) {
			var list = loadQuickReplies();
			list = list.filter(function(it) { return it.key.toLowerCase() !== key.toLowerCase(); });
			saveQuickReplies(list);
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

		var waQRTypedWord = '';

		function getEditorUserText(editable) {
			if (!editable) return '';
			var lexicalSpans = editable.querySelectorAll('span[data-lexical-text="true"]');
			if (lexicalSpans && lexicalSpans.length > 0) {
				var text = '';
				lexicalSpans.forEach(function(s) { text += s.textContent; });
				return cleanZeroWidth(text).trim();
			}
			var pTags = editable.querySelectorAll('p');
			if (pTags && pTags.length > 0) {
				var pText = '';
				pTags.forEach(function(p) { pText += (p.innerText || p.textContent || ''); });
				return cleanZeroWidth(pText).trim();
			}
			return cleanZeroWidth(editable.innerText || editable.textContent || '').trim();
		}

		function placeCaretAtEnd(editable) {
			editable.focus();
			var sel = window.getSelection();
			var spans = editable.querySelectorAll('span[data-lexical-text="true"]');
			var targetNode = (spans.length > 0) ? spans[spans.length - 1] : (editable.querySelector('p') || editable);
			var range = document.createRange();
			var tn = (targetNode.lastChild && targetNode.lastChild.nodeType === Node.TEXT_NODE) ? targetNode.lastChild :
			         (targetNode.firstChild && targetNode.firstChild.nodeType === Node.TEXT_NODE) ? targetNode.firstChild : null;
			if (tn) {
				range.setStart(tn, tn.textContent.length);
				range.collapse(true);
			} else {
				range.selectNodeContents(targetNode);
				range.collapse(false);
			}
			sel.removeAllRanges();
			sel.addRange(range);
		}

		// Selects the trailing "/command" the user typed so one insertText replaces it.
		// WhatsApp's editor (Lexical) reads the selection from beforeinput, so a single
		// replace is reliable; repeated execCommand('delete') and synthetic input events are not.
		function selectSlashCommand(editable, slashCmd) {
			editable.focus();
			var sel = window.getSelection();
			var len = slashCmd.length;
			var node = null, endOffset = 0;
			if (sel && sel.rangeCount) {
				var r0 = sel.getRangeAt(0);
				if (r0.collapsed && r0.startContainer.nodeType === Node.TEXT_NODE && editable.contains(r0.startContainer)) {
					var before = r0.startContainer.textContent.substring(0, r0.startOffset);
					if (before.toLowerCase().endsWith(slashCmd.toLowerCase())) {
						node = r0.startContainer;
						endOffset = r0.startOffset;
					}
				}
			}
			if (!node) {
				var walker = document.createTreeWalker(editable, NodeFilter.SHOW_TEXT);
				var t;
				while ((t = walker.nextNode())) {
					var i = t.textContent.toLowerCase().lastIndexOf(slashCmd.toLowerCase());
					if (i !== -1) { node = t; endOffset = i + len; }
				}
			}
			if (!node) return false;
			var range = document.createRange();
			range.setStart(node, endOffset - len);
			range.setEnd(node, endOffset);
			sel.removeAllRanges();
			sel.addRange(range);
			return true;
		}

		// Paste is the one insertion path WhatsApp's editor handles for multi-line text
		// (insertText drops newlines); fall back to insertText if the paste is not consumed.
		function insertIntoEditor(editable, text) {
			editable.focus();
			var handled = false;
			try {
				var dt = new DataTransfer();
				dt.setData('text/plain', text);
				var pasteEv = new ClipboardEvent('paste', { clipboardData: dt, bubbles: true, cancelable: true });
				editable.dispatchEvent(pasteEv);
				handled = pasteEv.defaultPrevented;
			} catch (err) {}
			if (!handled) document.execCommand('insertText', false, text);
		}

		var waQRLastApply = 0;

		function applyQuickReplyToEditable(editable, matchKey, textToInsert, typedWord) {
			if (!editable) return false;
			var now = Date.now();
			if (now - waQRLastApply < 400) return true; // ignore double-fired triggers
			waQRLastApply = now;

			if (window.waExpandVars) textToInsert = window.waExpandVars(textToInsert);

			var userText = getEditorUserText(editable);
			var m = userText.match(/(?:^|\s)(\/[\w-]*)$/);
			var slashCmd = m ? m[1] : (typedWord || matchKey);

			if (selectSlashCommand(editable, slashCmd)) {
				// Lexical learns about the new selection from an async selectionchange event,
				// so inserting immediately would replace the old (collapsed) selection.
				setTimeout(function() {
					insertIntoEditor(editable, textToInsert);
				}, 100);
			} else {
				placeCaretAtEnd(editable);
				setTimeout(function() {
					for (var k = 0; k < slashCmd.length; k++) {
						document.execCommand('delete');
					}
					insertIntoEditor(editable, textToInsert);
				}, 100);
			}
			showAddonToast('⚡ Template: ' + matchKey + ' diterapkan');
			return true;
		}

		function applyQuickReply(item, target) {
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
				var slashCmd = m ? m[1] : (waQRTypedWord || item.key);
				var kIdx = before.lastIndexOf(slashCmd);
				if (kIdx !== -1) {
					target.setRangeText(textToInsert, kIdx, selEnd, 'end');
				} else {
					target.setRangeText(textToInsert, selEnd - slashCmd.length >= 0 ? selEnd - slashCmd.length : 0, selEnd, 'end');
				}
				target.dispatchEvent(new Event('input', { bubbles: true }));
				showAddonToast('⚡ Template: ' + item.key + ' diterapkan');
				return true;
			}

			var editable = (target.closest && target.closest('[contenteditable="true"]')) ||
			               (target.isContentEditable ? target : null) ||
			               document.querySelector('footer div[contenteditable="true"]') ||
			               document.querySelector('div[contenteditable="true"][data-lexical-editor="true"]') ||
			               document.querySelector('div[contenteditable="true"]');
			if (!editable) return false;

			return applyQuickReplyToEditable(editable, item.key, textToInsert, waQRTypedWord);
		}

		// Floating Autocomplete Popup Controller for Quick Replies
		var waQRPopup = null;
		var waQRMatches = [];
		var waQRSelectedIndex = 0;
		var waQRTarget = null;

		function ensureWAQRPopup() {
			if (waQRPopup && waQRPopup.parentNode) return waQRPopup;
			waQRPopup = document.createElement('div');
			waQRPopup.id = 'wa-qr-popup';
			waQRPopup.className = 'wa-qr-popup';
			waQRPopup.style.display = 'none';
			(document.body || document.documentElement).appendChild(waQRPopup);
			return waQRPopup;
		}

		function hideWAQRPopup() {
			if (waQRPopup) {
				waQRPopup.style.display = 'none';
				waQRMatches = [];
				waQRSelectedIndex = 0;
				waQRTarget = null;
				waQRTypedWord = '';
			}
		}

		function showWAQRPopup(target, matches, queryWord) {
			if (!matches || matches.length === 0) {
				hideWAQRPopup();
				return;
			}
			waQRTarget = target;
			waQRMatches = matches;
			waQRSelectedIndex = 0;
			waQRTypedWord = queryWord || '';

			var popup = ensureWAQRPopup();
			var rect = target.getBoundingClientRect();
			var bottomPos = (window.innerHeight - rect.top + 10);
			if (bottomPos < 50) {
				var footer = document.querySelector('footer');
				if (footer) {
					var fRect = footer.getBoundingClientRect();
					bottomPos = (window.innerHeight - fRect.top + 10);
				} else {
					bottomPos = 90;
				}
			}

			var leftPos = rect.left;
			if (leftPos + 390 > window.innerWidth) {
				leftPos = window.innerWidth - 400;
			}
			if (leftPos < 16) leftPos = 16;

			popup.style.bottom = bottomPos + 'px';
			popup.style.left = leftPos + 'px';

			renderWAQRPopup();
			popup.style.display = 'flex';
		}

		function renderWAQRPopup() {
			if (!waQRPopup) return;
			var html = '<div class="wa-qr-header">' +
				'<div style="display:flex;align-items:center;gap:6px;"><span class="wa-status-dot"></span><b style="color:var(--wa-text);">Balas Cepat</b></div>' +
				'<span class="wa-qr-badge">Tekan [Tab] / [Enter] / [Klik]</span>' +
				'</div>' +
				'<div class="wa-qr-list">';

			for (var i = 0; i < waQRMatches.length; i++) {
				var it = waQRMatches[i];
				var isAct = (i === waQRSelectedIndex);
				var cleanSnippet = it.text.replace(/\n/g, ' ').substring(0, 56);
				if (it.text.length > 56) cleanSnippet += '...';
				html += '<div class="wa-qr-item ' + (isAct ? 'active' : '') + '" data-idx="' + i + '">' +
					'<div class="wa-qr-item-key"><span>' + it.key + '</span></div>' +
					'<div class="wa-qr-item-text">' + cleanSnippet + '</div>' +
					'</div>';
			}
			html += '</div>';
			waQRPopup.innerHTML = html;

			var items = waQRPopup.querySelectorAll('.wa-qr-item');
			for (var j = 0; j < items.length; j++) {
				(function(idx) {
					var el = items[idx];
					el.addEventListener('mousedown', function(ev) {
						if (ev.preventDefault) ev.preventDefault();
						if (ev.stopPropagation) ev.stopPropagation();
						var chosen = waQRMatches[idx];
						var tgt = waQRTarget;
						if (chosen) {
							applyQuickReply(chosen, tgt);
						}
						hideWAQRPopup();
					});
				})(j);
			}
		}

		function tryExpandQuickReply(e) {
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
							showAddonToast('⚡ Template: ' + item.key + ' diterapkan');
							return true;
						}
					}
				}
				return false;
			}

			var editable = (target.closest && target.closest('[contenteditable="true"]')) ||
			               (target.isContentEditable ? target : null) ||
			               document.querySelector('footer div[contenteditable="true"]') ||
			               document.querySelector('div[contenteditable="true"][data-lexical-editor="true"]') ||
			               document.querySelector('div[contenteditable="true"]');
			if (!editable) return false;

			var fullText = getEditorUserText(editable);
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

			applyQuickReplyToEditable(editable, matchItem.key, textToInsert, matchItem.key);
			return true;
		}

		// Input listener for slash typing in WhatsApp
		document.addEventListener('input', function(e) {
			var target = e.target;
			if (!target) return;

			var isInput = (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA');
			var isCE = target.isContentEditable || (target.closest && target.closest('[contenteditable="true"]'));
			if (!isInput && !isCE) {
				hideWAQRPopup();
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
					textBefore = getEditorUserText(host);
				}
			}

			var m = textBefore.match(/(?:^|\s)(\/[\w-]*)$/);
			if (!m) {
				hideWAQRPopup();
				return;
			}

			var query = m[1].toLowerCase();
			var all = loadQuickReplies();
			var matches = all.filter(function(it) {
				return it.key.toLowerCase().startsWith(query);
			});

			if (matches.length > 0) {
				showWAQRPopup(target, matches, m[1]);
			} else {
				hideWAQRPopup();
			}
		}, true);

		// Keydown listener for WhatsApp keyboard navigation & quick replies
		document.addEventListener('keydown', function(e) {
			var key = e.key;
			var code = e.code;
			var isTab = (key === 'Tab' || code === 'Tab' || e.keyCode === 9);
			var isEnter = (key === 'Enter' || code === 'Enter' || e.keyCode === 13);
			var isEscape = (key === 'Escape' || code === 'Escape' || e.keyCode === 27);
			var isUp = (key === 'ArrowUp' || code === 'ArrowUp' || e.keyCode === 38);
			var isDown = (key === 'ArrowDown' || code === 'ArrowDown' || e.keyCode === 40);
			var isSpace = (key === ' ' || code === 'Space' || e.keyCode === 32);

			if (waQRPopup && waQRPopup.style.display !== 'none' && waQRMatches.length > 0) {
				if (isDown) {
					if (e.preventDefault) e.preventDefault();
					if (e.stopPropagation) e.stopPropagation();
					waQRSelectedIndex = (waQRSelectedIndex + 1) % waQRMatches.length;
					renderWAQRPopup();
					return;
				}
				if (isUp) {
					if (e.preventDefault) e.preventDefault();
					if (e.stopPropagation) e.stopPropagation();
					waQRSelectedIndex = (waQRSelectedIndex - 1 + waQRMatches.length) % waQRMatches.length;
					renderWAQRPopup();
					return;
				}
				if (isEscape) {
					if (e.preventDefault) e.preventDefault();
					if (e.stopPropagation) e.stopPropagation();
					hideWAQRPopup();
					return;
				}
				if (isTab || isEnter) {
					if (e.preventDefault) e.preventDefault();
					if (e.stopPropagation) e.stopPropagation();
					var selected = waQRMatches[waQRSelectedIndex];
					var tgt = waQRTarget || e.target;
					if (selected) {
						applyQuickReply(selected, tgt);
					}
					hideWAQRPopup();
					return;
				}
			}

			if (isTab || isEnter || isSpace) {
				var expanded = tryExpandQuickReply(e);
				if (expanded) {
					hideWAQRPopup();
				}
			}
		}, true);

		document.addEventListener('click', function(e) {
			if (waQRPopup && waQRPopup.style.display !== 'none') {
				if (!waQRPopup.contains(e.target)) {
					hideWAQRPopup();
				}
			}
		}, true);

		// Feature: In-App Notes (several notes, searchable). Stored in localStorage as wa_notes.
		var WA_NOTES_KEY = 'wa_notes';
		var waNotes = { list: [], active: null };

		function newNote(text) {
			return { id: 'n' + Date.now().toString(36) + Math.random().toString(36).slice(2, 5), text: text || '', updated: Date.now() };
		}
		function saveNotes() {
			try { localStorage.setItem(WA_NOTES_KEY, JSON.stringify(waNotes)); } catch(e) {}
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
					showAddonToast('📋 Catatan disalin ke clipboard');
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

			showActive();
			renderList();
			return drawer;
		}

		window.toggleScratchpad = function() {
			whenDOMReady(function() {
				var drawer = createScratchpadDrawer();
				if (!drawer) return;
				var isOpen = drawer.style.right === '0px';
				drawer.style.right = isOpen ? '-360px' : '0px';
				if (!isOpen) {
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

		// Feature: Status / Story Media Downloader
		function setupStatusDownloader() {
			setInterval(function() {
				var dialog = document.querySelector('div[role="dialog"], div[data-animate-status-v3="true"]');
				var btn = document.getElementById('wa-status-dl-btn');
				if (!dialog) {
					if (btn) btn.style.display = 'none';
					return;
				}
				var statusContainer = document.querySelector('div[data-animate-status-v3="true"], div[role="dialog"] [data-testid="status-v3-main"]');
				if (!statusContainer) {
					var v = dialog.querySelector('video');
					if (v && v.offsetHeight > 200 && (v.src || v.currentSrc)) {
						statusContainer = v.parentElement;
					}
				}
				if (!statusContainer) {
					if (btn) btn.style.display = 'none';
					return;
				}
				if (!btn) {
					btn = document.createElement('div');
					btn.id = 'wa-status-dl-btn';
					btn.title = 'Unduh Media Status WhatsApp';
					btn.style.cssText = 'position:fixed;top:20px;right:90px;z-index:9999999;background:#00a884;color:#111b21;padding:8px 14px;border-radius:20px;font-family:Segoe UI,sans-serif;font-size:12px;font-weight:700;display:flex;align-items:center;gap:6px;cursor:pointer;box-shadow:0 4px 12px rgba(0,0,0,0.6);user-select:none;';
					btn.innerHTML = '<span style="font-size:14px;">⬇️</span><span>Unduh Status</span>';
					btn.onclick = function(e) {
						e.preventDefault();
						e.stopPropagation();
						downloadCurrentStatus();
					};
					document.body.appendChild(btn);
				}
				btn.style.display = 'flex';
			}, 2000);

			function downloadCurrentStatus() {
				var video = document.querySelector('div[role="dialog"] video, div[tabindex="-1"] video, video');
				var img = document.querySelector('div[role="dialog"] img[src*="blob:"], div[tabindex="-1"] img[src*="blob:"], div[role="dialog"] img');
				
				var src = '';
				var isVideo = false;
				if (video && (video.src || video.currentSrc) && video.offsetHeight > 200) {
					src = video.currentSrc || video.src;
					isVideo = true;
				} else if (img && img.src && img.offsetHeight > 200) {
					src = img.src;
				}

				if (!src) {
					showAddonToast('⚠️ Media status tidak ditemukan');
					return;
				}

				showAddonToast('⏳ Mengunduh media status...');
				var ext = isVideo ? 'mp4' : 'jpg';
				var d = new Date();
				var dateStr = d.getFullYear() + '' + (d.getMonth()+1) + '' + d.getDate() + '_' + d.getHours() + '' + d.getMinutes() + '' + d.getSeconds();
				var filename = 'WA-Status-' + dateStr + '.' + ext;

				fetch(src).then(function(r) { return r.blob(); }).then(function(blob) {
					var blobUrl = URL.createObjectURL(blob);
					var a = document.createElement('a');
					a.href = blobUrl;
					a.download = filename;
					document.body.appendChild(a);
					a.click();
					setTimeout(function() {
						document.body.removeChild(a);
						URL.revokeObjectURL(blobUrl);
					}, 1000);
					showAddonToast('✅ Berhasil diunduh: ' + filename);
				}).catch(function() {
					var a = document.createElement('a');
					a.href = src;
					a.target = '_blank';
					a.download = filename;
					a.click();
					showAddonToast('✅ Media dibuka di unduhan!');
				});
			}
		}

		whenDOMReady(function() {
			setupStatusDownloader();
		});

		// Feature: Anonymized Screenshot Mode
		var isAnonMode = false;
		function updateAnonymizeMode(active) {
			isAnonMode = active;
			var style = document.getElementById('wa-anon-style');
			var banner = document.getElementById('wa-anon-banner');

			if (active) {
				if (!style) {
					style = document.createElement('style');
					style.id = 'wa-anon-style';
					var target = document.head || document.documentElement || document.body;
					if (target) target.appendChild(style);
				}
				style.textContent = [
					'#pane-side [data-testid="cell-frame-title"], #pane-side span[title], #main header [data-testid="conversation-info-header-chat-title-name"], #main [data-testid="author"], header span[dir="auto"] {',
					'  color: transparent !important;',
					'  text-shadow: 0 0 10px rgba(255,255,255,0.85) !important;',
					'}',
					'#pane-side [role="row"] img:not(.emoji), #main header img:not(.emoji) {',
					'  filter: blur(12px) grayscale(1) !important;',
					'}',
					'span[dir="ltr"] {',
					'  color: transparent !important;',
					'  text-shadow: 0 0 8px rgba(255,255,255,0.7) !important;',
					'}'
				].join('\n');

				if (!banner) {
					banner = document.createElement('div');
					banner.id = 'wa-anon-banner';
					banner.style.cssText = 'position:fixed;top:0;left:0;width:100vw;background:#00a884;color:#111b21;padding:8px 18px;z-index:999999999;display:flex;align-items:center;justify-content:space-between;font-family:Segoe UI,sans-serif;font-size:13px;font-weight:600;box-shadow:0 4px 12px rgba(0,0,0,0.5);box-sizing:border-box;user-select:none;';
					banner.innerHTML = '<span>🔒 Mode Screenshot Anonim Aktif - Identitas kontak disamarkan. Ambil tangkapan layar (Win + Shift + S), lalu klik tombol selesai.</span>' +
						'<button id="wa-anon-done-btn" style="background:#111b21;color:#00a884;border:none;padding:5px 14px;border-radius:6px;cursor:pointer;font-weight:700;font-size:12px;">✓ Selesai</button>';
					document.body.appendChild(banner);
					var doneBtn = document.getElementById('wa-anon-done-btn');
					if (doneBtn) {
						doneBtn.onclick = function() {
							updateAnonymizeMode(false);
						};
					}
				}
				banner.style.display = 'flex';
				showAddonToast('🔒 Mode Screenshot Anonim: AKTIF');
			} else {
				if (style) style.textContent = '';
				if (banner) banner.style.display = 'none';
				showAddonToast('Mode Screenshot Anonim dinonaktifkan');
			}
		}

		window.toggleAnonymizeMode = function() {
			updateAnonymizeMode(!isAnonMode);
		};

		// Feature: Voice Note Speed Multiplier ([ / ]) & Picture-in-Picture (Alt + V)
		var audioSpeed = 1.0;
		var changeAudioSpeed;
		(function() {
			if (window.self !== window.top) return;

			try {
				audioSpeed = parseFloat(localStorage.getItem('wa_audio_speed') || '1.0');
				if (isNaN(audioSpeed) || audioSpeed <= 0) audioSpeed = 1.0;
			} catch(e) {}

			document.addEventListener('play', function(e) {
				if (e.target && (e.target.tagName === 'AUDIO' || e.target.tagName === 'VIDEO')) {
					e.target.playbackRate = audioSpeed;
				}
			}, true);

			changeAudioSpeed = function(deltaOrValue) {
				if (typeof deltaOrValue === 'number' && deltaOrValue >= 0.5 && deltaOrValue <= 3.0) {
					audioSpeed = deltaOrValue;
				} else {
					var rates = [0.5, 0.75, 1.0, 1.25, 1.5, 1.75, 2.0, 2.25, 2.5, 3.0];
					var idx = 2; // 1.0 default
					for (var i = 0; i < rates.length; i++) {
						if (Math.abs(rates[i] - audioSpeed) < 0.05) {
							idx = i;
							break;
						}
					}
					idx += deltaOrValue;
					if (idx < 0) idx = 0;
					if (idx >= rates.length) idx = rates.length - 1;
					audioSpeed = rates[idx];
				}

				try {
					localStorage.setItem('wa_audio_speed', audioSpeed.toString());
				} catch(e) {}

				var mediaEls = document.querySelectorAll('audio, video');
				mediaEls.forEach(function(m) {
					try { m.playbackRate = audioSpeed; } catch(e) {}
				});
				showAddonToast('⏩ Kecepatan Suara: ' + audioSpeed + 'x');
			};

			window.addEventListener('keydown', function(e) {
				var active = document.activeElement;
				if (active && (active.tagName === 'INPUT' || active.tagName === 'TEXTAREA' || active.isContentEditable || (active.getAttribute && active.getAttribute('contenteditable') === 'true'))) {
					return;
				}
				if (!e.ctrlKey && !e.altKey && !e.metaKey) {
					if (e.key === ']' || e.code === 'BracketRight') {
						if (e.preventDefault) e.preventDefault();
						changeAudioSpeed(1);
					} else if (e.key === '[' || e.code === 'BracketLeft') {
						if (e.preventDefault) e.preventDefault();
						changeAudioSpeed(-1);
					}
				}
				// Alt + V for Picture-in-Picture
				if (e.altKey && (e.code === 'KeyV' || e.key === 'v' || e.key === 'V')) {
					if (e.preventDefault) e.preventDefault();
					var videos = document.querySelectorAll('video');
					for (var v = 0; v < videos.length; v++) {
						var vid = videos[v];
						if (!vid.paused || videos.length === 1) {
							if (document.pictureInPictureElement) {
								document.exitPictureInPicture();
								showAddonToast('📺 Picture-in-Picture ditutup');
							} else if (vid.requestPictureInPicture) {
								vid.requestPictureInPicture();
								showAddonToast('📺 Picture-in-Picture aktif');
							}
							break;
						}
					}
				}
			}, true);
		})();

		// Feature: App Lock with PIN (Ctrl + L & 5-minute Inactivity Auto-Lock)
		var storedPin = '1234';
		(function() {
			if (window.self !== window.top) return;

			var isLocked = false;
			try {
				storedPin = localStorage.getItem('wa_app_pin') || '1234';
				isLocked = localStorage.getItem('wa_is_locked') === 'true';
			} catch(e) {}

			window.updateStoredPin = function(pin) {
				if (pin && pin.length >= 4) {
					storedPin = pin;
					try { localStorage.setItem('wa_app_pin', pin); } catch(e) {}
				}
			};
			if (window.getAppPin) {
				window.getAppPin().then(function(p) {
					if (p && p.length >= 4) {
						window.updateStoredPin(p);
					}
				});
			}

			var inactivityTimer = null;
			var INACTIVITY_TIMEOUT = 5 * 60 * 1000;

			function resetInactivityTimer() {
				clearTimeout(inactivityTimer);
				inactivityTimer = setTimeout(function() {
					window.lockWhatsApp();
				}, INACTIVITY_TIMEOUT);
			}

			['mousedown', 'keydown', 'touchstart'].forEach(function(evt) {
				window.addEventListener(evt, resetInactivityTimer, { passive: true });
			});
			resetInactivityTimer();

			function createLockOverlay() {
				var existing = document.getElementById('wa-lock-overlay');
				if (existing) return existing;
				if (!document.body) return null;
				var overlay = document.createElement('div');
				overlay.id = 'wa-lock-overlay';
				overlay.className = 'wa-modal-backdrop';
				overlay.style.cssText = 'position:fixed !important;inset:0 !important;width:100vw !important;height:100vh !important;background:#0c1317 !important;z-index:2147483646 !important;align-items:center !important;justify-content:center !important;display:none;';
				overlay.innerHTML = '<div class="wa-modal-box" style="width:330px;padding:30px 26px;text-align:center;align-items:center;">' +
					'<div class="wa-header-icon" style="width:48px;height:48px;border-radius:50%;margin-bottom:14px;">' +
					'  <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>' +
					'</div>' +
					'<h2 style="margin:0 0 4px 0;font-size:18px;font-weight:600;color:var(--wa-text);">WhatsApp Terkunci</h2>' +
					'<p style="margin:0 0 18px 0;font-size:12.5px;color:var(--wa-text-muted);">Masukkan PIN untuk membuka akses</p>' +
					'<input id="wa-pin-input" class="wa-input" type="password" maxlength="6" style="width:180px;height:42px;text-align:center;font-size:22px;letter-spacing:8px;margin-bottom:12px;">' +
					'<div id="wa-pin-error" style="color:var(--wa-danger);font-size:12px;min-height:18px;margin-bottom:10px;"></div>' +
					'<button id="wa-pin-unlock-btn" class="wa-btn wa-btn-primary" style="width:100%;height:38px;font-size:13px;margin-bottom:12px;">Buka Kunci</button>' +
					'<div style="font-size:11px;color:var(--wa-text-dim);">Default: 1234 • Ctrl+L untuk mengunci</div>' +
					'<div style="margin-top:10px;"><a id="wa-pin-change-link" href="javascript:void(0)" style="color:var(--wa-primary);font-size:12px;text-decoration:none;font-weight:500;">Ganti / Ubah PIN</a></div>' +
					'<div style="margin-top:14px;padding-top:12px;border-top:1px solid rgba(255,255,255,0.08);width:100%;display:flex;gap:6px;justify-content:center;">' +
					'  <button id="wa-lock-btn-tg" class="wa-btn wa-btn-secondary" style="font-size:11px;padding:5px 12px;border-radius:14px;color:#24A1DE;border-color:rgba(36,161,222,0.3);">' +
					'    <span style="font-weight:bold;">Telegram</span> (Ctrl+2)' +
					'  </button>' +
					'</div>' +
					'</div>';
				document.body.appendChild(overlay);

				var btnLockTG = overlay.querySelector('#wa-lock-btn-tg');
				if (btnLockTG) {
					btnLockTG.onclick = function(e) {
						if (e.preventDefault) e.preventDefault();
						if (e.stopPropagation) e.stopPropagation();
						if (window.switchToTelegram) window.switchToTelegram();
						else if (window.openTelegramWindow) window.openTelegramWindow();
					};
				}

				var pinInput = document.getElementById('wa-pin-input');
				var unlockBtn = document.getElementById('wa-pin-unlock-btn');
				var errEl = document.getElementById('wa-pin-error');
				var changeLink = document.getElementById('wa-pin-change-link');

				if (changeLink) {
					changeLink.onclick = function(e) {
						if (e.preventDefault) e.preventDefault();
						window.openChangePinModal();
					};
				}

				function unlock() {
					var val = pinInput.value;
					if (val === storedPin) {
						overlay.style.display = 'none';
						pinInput.value = '';
						errEl.textContent = '';
						try { localStorage.setItem('wa_is_locked', 'false'); } catch(e) {}
						showAddonToast('🔓 WhatsApp terbuka');
						resetInactivityTimer();
					} else {
						errEl.textContent = 'PIN salah! Coba lagi.';
						pinInput.value = '';
						pinInput.focus();
					}
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
					pinInput.addEventListener('input', function(e) {
						var current = (pinInput.value || '').trim();
						if (current.length >= 4 && current === storedPin) {
							unlock();
						}
					});
				}

				return overlay;
			}

			function createChangePinModal() {
				var existing = document.getElementById('wa-change-pin-modal');
				if (existing) return existing;
				if (!document.body) return null;
				var modal = document.createElement('div');
				modal.id = 'wa-change-pin-modal';
				modal.className = 'wa-modal-backdrop';
				modal.innerHTML = '<div class="wa-modal-box" style="width:340px;padding:26px 24px;text-align:center;">' +
					'<div class="wa-header-icon" style="width:40px;height:40px;border-radius:50%;margin:0 auto 12px auto;">' +
					'  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 2l-2 2m-1.5 1.5L10 13l-4 1 1-4 7.5-7.5m1.5-1.5l2-2"/><circle cx="7.5" cy="16.5" r="3.5"/></svg>' +
					'</div>' +
					'<h3 style="margin:0 0 4px 0;font-size:16px;font-weight:600;color:var(--wa-text);">Ubah PIN WhatsApp</h3>' +
					'<p style="margin:0 0 16px 0;font-size:12px;color:var(--wa-text-muted);">Tentukan 4-6 angka PIN keamanan Anda</p>' +
					'<div style="text-align:left;margin-bottom:10px;">' +
					'<label style="font-size:11px;color:var(--wa-text-muted);display:block;margin-bottom:4px;">PIN Saat Ini (Default: 1234):</label>' +
					'<input id="wa-cp-old" class="wa-input" type="password" maxlength="6" placeholder="PIN Lama" style="width:100%;">' +
					'</div>' +
					'<div style="text-align:left;margin-bottom:10px;">' +
					'<label style="font-size:11px;color:var(--wa-text-muted);display:block;margin-bottom:4px;">PIN Baru (4-6 angka):</label>' +
					'<input id="wa-cp-new" class="wa-input" type="password" maxlength="6" placeholder="PIN Baru" style="width:100%;">' +
					'</div>' +
					'<div style="text-align:left;margin-bottom:12px;">' +
					'<label style="font-size:11px;color:var(--wa-text-muted);display:block;margin-bottom:4px;">Konfirmasi PIN Baru:</label>' +
					'<input id="wa-cp-confirm" class="wa-input" type="password" maxlength="6" placeholder="Ulangi PIN Baru" style="width:100%;">' +
					'</div>' +
					'<div id="wa-cp-error" style="color:var(--wa-danger);font-size:12px;min-height:16px;margin-bottom:12px;"></div>' +
					'<div style="display:flex;gap:8px;">' +
					'<button id="wa-cp-cancel-btn" class="wa-btn wa-btn-secondary" style="flex:1;">Batal</button>' +
					'<button id="wa-cp-save-btn" class="wa-btn wa-btn-primary" style="flex:1;">Simpan PIN</button>' +
					'</div></div>';
				document.body.appendChild(modal);

				var oldInput = document.getElementById('wa-cp-old');
				var newInput = document.getElementById('wa-cp-new');
				var confirmInput = document.getElementById('wa-cp-confirm');
				var errEl = document.getElementById('wa-cp-error');
				var cancelBtn = document.getElementById('wa-cp-cancel-btn');
				var saveBtn = document.getElementById('wa-cp-save-btn');

				function save() {
					var oldVal = oldInput ? oldInput.value : '';
					var newVal = newInput ? newInput.value : '';
					var confirmVal = confirmInput ? confirmInput.value : '';

					if (oldVal !== storedPin) {
						errEl.textContent = 'PIN lama salah!';
						if (oldInput) { oldInput.value = ''; oldInput.focus(); }
						return;
					}
					if (!/^\d{4,6}$/.test(newVal)) {
						errEl.textContent = 'PIN baru harus 4-6 angka!';
						if (newInput) newInput.focus();
						return;
					}
					if (newVal !== confirmVal) {
						errEl.textContent = 'Konfirmasi PIN tidak cocok!';
						if (confirmInput) { confirmInput.value = ''; confirmInput.focus(); }
						return;
					}

					storedPin = newVal;
					try { localStorage.setItem('wa_app_pin', newVal); } catch(e) {}
					if (window.syncAppPin) {
						window.syncAppPin(newVal);
					}
					modal.style.display = 'none';
					oldInput.value = '';
					newInput.value = '';
					confirmInput.value = '';
					errEl.textContent = '';
					showAddonToast('✅ PIN berhasil diubah!');
				}

				if (saveBtn) saveBtn.onclick = save;
				if (cancelBtn) cancelBtn.onclick = function() { modal.style.display = 'none'; };
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
								modal.style.display = 'none';
							}
						});
					}
				});

				return modal;
			}

			window.openChangePinModal = function() {
				whenDOMReady(function() {
					var modal = createChangePinModal();
					if (modal) {
						modal.style.display = 'flex';
						var oldInput = document.getElementById('wa-cp-old');
						var errEl = document.getElementById('wa-cp-error');
						if (errEl) errEl.textContent = '';
						if (oldInput) {
							oldInput.value = '';
							setTimeout(function() { oldInput.focus(); }, 50);
						}
					}
				});
			};

			window.lockWhatsApp = function() {
				whenDOMReady(function() {
					var overlay = createLockOverlay();
					if (overlay) {
						if (overlay.style.display === 'flex') {
							// Already locked and visible! Never wipe input or re-focus if user is currently typing
							return;
						}
						overlay.style.display = 'flex';
						try { localStorage.setItem('wa_is_locked', 'true'); } catch(e) {}
						var input = document.getElementById('wa-pin-input');
						if (input) {
							input.value = '';
							setTimeout(function() {
								if (document.activeElement !== input) {
									input.focus();
								}
							}, 50);
						}
					}
				});
			};

			window.addEventListener('keydown', function(e) {
				if ((e.ctrlKey || e.metaKey) && (e.code === 'KeyL' || e.key === 'l' || e.key === 'L')) {
					if (e.preventDefault) e.preventDefault();
					if (e.stopPropagation) e.stopPropagation();
					window.lockWhatsApp();
				}
			}, true);

			if (isLocked) {
				whenDOMReady(function() {
					window.lockWhatsApp();
				});
			}
		})();

		// Feature: Quick Unread Chats Filter
		var isUnreadFilterActive = false;
		window.toggleUnreadFilter = function() {
			isUnreadFilterActive = !isUnreadFilterActive;
			var filterStyle = document.getElementById('wa-unread-filter-style');
			if (!filterStyle) {
				filterStyle = document.createElement('style');
				filterStyle.id = 'wa-unread-filter-style';
				(document.head || document.documentElement || document.body).appendChild(filterStyle);
			}
			if (isUnreadFilterActive) {
				filterStyle.textContent = '#pane-side [role="listitem"]:not(:has(span[aria-label*="unread"])):not(:has(span[aria-label*="belum"])):not(:has([data-icon*="unread"])) { display: none !important; }';
				showAddonToast('✉️ Filter Chat Belum Dibaca: AKTIF');
			} else {
				filterStyle.textContent = '';
				showAddonToast('Filter Chat Belum Dibaca: NONAKTIF');
			}
			var btn = document.getElementById('wa-unread-filter-btn');
			if (btn) {
				btn.style.background = isUnreadFilterActive ? '#00a884' : '#111b21';
				btn.style.color = isUnreadFilterActive ? '#111b21' : '#00a884';
			}
			var ccBtn = document.getElementById('wa-cc-unread-toggle');
			if (ccBtn) {
				ccBtn.className = isUnreadFilterActive ? 'wa-btn wa-btn-primary' : 'wa-btn wa-btn-secondary';
				ccBtn.textContent = isUnreadFilterActive ? 'Nonaktifkan Filter' : 'Aktifkan Filter';
			}
			return isUnreadFilterActive;
		};

		// Feature: In-App Control Center Modal & Sidebar Integration
		(function() {
			if (window.self !== window.top) return;

			window.syncAutoStartUI = function() {
				if (window.getAutoStartStatus) {
					window.getAutoStartStatus().then(function(res) {
						var autoCb = document.getElementById('wa-cc-autostart-cb');
						if (autoCb) autoCb.checked = !!res;
					});
				}
			};

			function createControlCenterModal() {
				var existing = document.getElementById('wa-addon-modal');
				if (existing) return existing;
				if (!document.body) return null;

				var modal = document.createElement('div');
				modal.id = 'wa-addon-modal';
				modal.className = 'wa-modal-backdrop';
				modal.innerHTML = WA_CC_HTML;

				document.body.appendChild(modal);

				// Tab Switching
				var tabBtns = modal.querySelectorAll('.wa-tab-btn');
				var tabPanes = modal.querySelectorAll('.wa-tab-pane');
				tabBtns.forEach(function(btn) {
					btn.onclick = function() {
						var tabId = btn.getAttribute('data-tab');
						tabBtns.forEach(function(b) { b.classList.remove('active'); });
						btn.classList.add('active');
						tabPanes.forEach(function(p) {
							if (p.id === 'wa-tab-content-' + tabId) {
								p.style.display = 'flex';
							} else {
								p.style.display = 'none';
							}
						});
					};
				});

				// Wire up Stealth Checkboxes
				var unreadToggleBtn = document.getElementById('wa-cc-unread-toggle');
				if (unreadToggleBtn) {
					unreadToggleBtn.onclick = function() {
						window.toggleUnreadFilter();
						syncPrivacyUI();
					};
				}

				var dualAccBtn = document.getElementById('wa-cc-dual-acc-btn');
				if (dualAccBtn) {
					dualAccBtn.onclick = function() {
						if (window.openDualAccount) {
							window.openDualAccount();
							showAddonToast('Membuka WhatsApp Akun Ke-2...');
						}
					};
				}
				var filterUnreadBtn = document.getElementById('wa-cc-filter-unread-btn');
				if (filterUnreadBtn) {
					filterUnreadBtn.onclick = function() {
						window.toggleUnreadFilter();
					};
				}
				var tgBtn = document.getElementById('wa-cc-tg-btn');
				if (tgBtn) {
					tgBtn.onclick = function() {
						modal.style.display = 'none';
						if (window.switchToTelegram) {
							window.switchToTelegram();
							showAddonToast('Beralih ke Telegram Web...');
						} else if (window.openTelegramWindow) {
							window.openTelegramWindow();
							showAddonToast('Beralih ke Telegram Web...');
						}
					};
				}
				var splitBtn = document.getElementById('wa-cc-split-btn');
				if (splitBtn) {
					splitBtn.onclick = function() {
						modal.style.display = 'none';
						if (window.openSideBySideView) {
							window.openSideBySideView();
							showAddonToast('Mengatur Tampilan Berdampingan...');
						}
					};
				}

				// Wire up close
				var closeBtn = document.getElementById('wa-cc-close');
				if (closeBtn) closeBtn.onclick = function() { modal.style.display = 'none'; };
				modal.onclick = function(e) { if (e.target === modal) modal.style.display = 'none'; };
				window.addEventListener('keydown', function(e) {
					if (e.key === 'Escape' && modal.style.display === 'flex') {
						modal.style.display = 'none';
					}
				});

				// Wire up privacy toggle & checkboxes
				var privBtn = document.getElementById('wa-cc-privacy-btn');
				var cbContacts = document.getElementById('wa-cc-blur-contacts');
				var cbPreview = document.getElementById('wa-cc-blur-preview');
				var cbMessages = document.getElementById('wa-cc-blur-messages');
				var cbMedia = document.getElementById('wa-cc-blur-media');
				var cbAvatars = document.getElementById('wa-cc-blur-avatars');

				function syncPrivacyUI() {
					if (cbContacts) cbContacts.checked = privacyConfig.blurContacts;
					if (cbPreview) cbPreview.checked = privacyConfig.blurPreview;
					if (cbMessages) cbMessages.checked = privacyConfig.blurMessages;
					if (cbMedia) cbMedia.checked = privacyConfig.blurMedia;
					if (cbAvatars) cbAvatars.checked = privacyConfig.blurAvatars;

					var isAct = document.body && document.body.classList.contains('wa-privacy-active');
					if (privBtn) {
						privBtn.textContent = isAct ? 'Status: Aktif' : 'Status: Nonaktif';
						privBtn.className = isAct ? 'wa-btn wa-btn-primary' : 'wa-btn wa-btn-secondary';
					}

					[3, 5, 8].forEach(function(p) {
						var btn = document.getElementById('wa-blur-btn-' + p);
						if (btn) {
							var sel = (privacyConfig.blurIntensity === p);
							btn.className = sel ? 'wa-segmented-btn active' : 'wa-segmented-btn';
						}
					});

					var unreadToggleBtn = document.getElementById('wa-cc-unread-toggle');
					if (unreadToggleBtn) {
						unreadToggleBtn.className = isUnreadFilterActive ? 'wa-btn wa-btn-primary' : 'wa-btn wa-btn-secondary';
						unreadToggleBtn.textContent = isUnreadFilterActive ? 'Nonaktifkan Filter' : 'Aktifkan Filter';
					}

					var anBtn = document.getElementById('wa-cc-anon-btn');
					if (anBtn) {
						anBtn.className = isAnonMode ? 'wa-btn wa-btn-primary' : 'wa-btn wa-btn-secondary';
						anBtn.textContent = isAnonMode ? 'Mode Anonim: Aktif' : 'Aktifkan';
					}

					// Sync Eye comfort
					var thDef = document.getElementById('wa-theme-default');
					var thOled = document.getElementById('wa-theme-oled');
					var thWarm = document.getElementById('wa-theme-warm');
					var cbCompact = document.getElementById('wa-compact-cb');
					if (thDef) thDef.className = (eyeComfortConfig.theme === 'default') ? 'wa-theme-card active' : 'wa-theme-card';
					if (thOled) thOled.className = (eyeComfortConfig.theme === 'oled') ? 'wa-theme-card active' : 'wa-theme-card';
					if (thWarm) thWarm.className = (eyeComfortConfig.theme === 'warm') ? 'wa-theme-card active' : 'wa-theme-card';
					if (cbCompact) cbCompact.checked = !!eyeComfortConfig.compact;

					// Sync Quick replies list
					renderQuickRepliesUI();

					// Sync Autostart UI
					if (window.syncAutoStartUI) window.syncAutoStartUI();
				}

				function renderQuickRepliesUI() {
					var listEl = document.getElementById('wa-qr-list');
					if (!listEl) return;
					listEl.innerHTML = '';
					var items = loadQuickReplies();
					items.forEach(function(item) {
						var row = document.createElement('div');
						row.className = 'wa-row';
						row.style.cssText = 'background:var(--wa-bg);border:1px solid var(--wa-border);padding:6px 12px;border-radius:var(--wa-radius-sm);';
						row.innerHTML = '<div style="display:flex;align-items:center;gap:8px;overflow:hidden;"><b style="color:var(--wa-primary);font-size:12px;font-family:monospace;">' + item.key + '</b><span style="color:var(--wa-text-muted);font-size:12px;text-overflow:ellipsis;overflow:hidden;white-space:nowrap;">' + (item.text.length > 32 ? item.text.substring(0,32) + '...' : item.text) + '</span></div>' +
							'<button class="wa-del-qr wa-close-btn" style="width:24px;height:24px;color:var(--wa-danger);" title="Hapus shortcut"><svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg></button>';
						var delBtn = row.querySelector('.wa-del-qr');
						if (delBtn) {
							delBtn.onclick = function() {
								deleteQuickReply(item.key);
								renderQuickRepliesUI();
								showAddonToast('Shortcut ' + item.key + ' dihapus');
							};
						}
						listEl.appendChild(row);
					});
				}

				// Wire up Add Quick Reply
				var qrAddBtn = document.getElementById('wa-qr-add-btn');
				var qrNewKey = document.getElementById('wa-qr-new-key');
				var qrNewText = document.getElementById('wa-qr-new-text');
				if (qrAddBtn) {
					qrAddBtn.onclick = function() {
						var k = qrNewKey ? qrNewKey.value : '';
						var t = qrNewText ? qrNewText.value : '';
						if (k && t) {
							if (addQuickReply(k, t)) {
								if (qrNewKey) qrNewKey.value = '';
								if (qrNewText) qrNewText.value = '';
								renderQuickRepliesUI();
								showAddonToast('Shortcut ' + k + ' disimpan');
							}
						}
					};
				}

				// Wire up Theme buttons
				var thDef = document.getElementById('wa-theme-default');
				var thOled = document.getElementById('wa-theme-oled');
				var thWarm = document.getElementById('wa-theme-warm');
				var cbCompact = document.getElementById('wa-compact-cb');
				if (thDef) thDef.onclick = function() { eyeComfortConfig.theme = 'default'; saveEyeComfortConfig(); syncPrivacyUI(); };
				if (thOled) thOled.onclick = function() { eyeComfortConfig.theme = 'oled'; saveEyeComfortConfig(); syncPrivacyUI(); };
				if (thWarm) thWarm.onclick = function() { eyeComfortConfig.theme = 'warm'; saveEyeComfortConfig(); syncPrivacyUI(); };
				if (cbCompact) cbCompact.onchange = function() { eyeComfortConfig.compact = cbCompact.checked; saveEyeComfortConfig(); };

				// Wire up Productivity buttons
				var spBtn = document.getElementById('wa-cc-scratchpad-btn');
				if (spBtn) spBtn.onclick = function() { modal.style.display = 'none'; window.toggleScratchpad(); };
				var anBtn = document.getElementById('wa-cc-anon-btn');
				if (anBtn) anBtn.onclick = function() { modal.style.display = 'none'; window.toggleAnonymizeMode(); };

				// Wire up Autostart Checkbox
				var autoCb = document.getElementById('wa-cc-autostart-cb');
				if (autoCb) {
					autoCb.onchange = function() {
						if (window.setAutoStartStatus) {
							var want = autoCb.checked;
							window.setAutoStartStatus(want).then(function(ok) {
								if (!ok) {
									autoCb.checked = !want;
									showAddonToast('Gagal mengubah Mulai Otomatis');
									return;
								}
								showAddonToast(want ? 'Mulai Otomatis: Aktif' : 'Mulai Otomatis: Nonaktif');
							});
						}
					};
				}

				if (privBtn) {
					privBtn.onclick = function() {
						window.togglePrivacyMode();
						syncPrivacyUI();
					};
				}

				[cbContacts, cbPreview, cbMessages, cbMedia, cbAvatars].forEach(function(cb) {
					if (cb) {
						cb.onchange = function() {
							privacyConfig.blurContacts = cbContacts.checked;
							privacyConfig.blurPreview = cbPreview.checked;
							privacyConfig.blurMessages = cbMessages.checked;
							privacyConfig.blurMedia = cbMedia.checked;
							privacyConfig.blurAvatars = cbAvatars.checked;
							savePrivacyConfig();
						};
					}
				});

				[3, 5, 8].forEach(function(p) {
					var btn = document.getElementById('wa-blur-btn-' + p);
					if (btn) {
						btn.onclick = function() {
							privacyConfig.blurIntensity = p;
							savePrivacyConfig();
							syncPrivacyUI();
							showAddonToast('Intensitas blur diubah ke ' + p + 'px');
						};
					}
				});

				// Wire up direct chat inside modal
				var phoneInp = document.getElementById('wa-cc-phone');
				var phoneBtn = document.getElementById('wa-cc-phone-btn');
				if (phoneBtn) {
					phoneBtn.onclick = function() {
						var val = phoneInp ? phoneInp.value : '';
						if (val) {
							modal.style.display = 'none';
							if (phoneInp) phoneInp.value = '';
							openDirectChatByNumber(val);
						}
					};
				}
				if (phoneInp) {
					phoneInp.onkeydown = function(e) {
						if (e.key === 'Enter' && phoneInp.value) {
							modal.style.display = 'none';
							openDirectChatByNumber(phoneInp.value);
							phoneInp.value = '';
						}
					};
				}

				// Wire up App Lock buttons
				var lockBtn = document.getElementById('wa-cc-lock-btn');
				if (lockBtn) {
					lockBtn.onclick = function() {
						modal.style.display = 'none';
						window.lockWhatsApp();
					};
				}
				var cpBtn = document.getElementById('wa-cc-changepin-btn');
				if (cpBtn) {
					cpBtn.onclick = function() {
						modal.style.display = 'none';
						window.openChangePinModal();
					};
				}

				// Wire up zoom
				var zIn = document.getElementById('wa-cc-zoom-in');
				var zOut = document.getElementById('wa-cc-zoom-out');
				var zRes = document.getElementById('wa-cc-zoom-reset');
				if (zIn) zIn.onclick = function() { applyZoom(currentZoom + 5); };
				if (zOut) zOut.onclick = function() { applyZoom(currentZoom - 5); };
				if (zRes) zRes.onclick = function() { applyZoom(100); };

				// Wire up audio buttons
				var audioBtns = modal.querySelectorAll('.wa-audio-btn');
				audioBtns.forEach(function(ab) {
					ab.onclick = function() {
						var spd = parseFloat(ab.getAttribute('data-spd'));
						changeAudioSpeed(spd);
					};
				});

				// Wire up reload & logout
				var relBtn = document.getElementById('wa-cc-reload-btn');
				if (relBtn) relBtn.onclick = function() { window.location.reload(); };

				var logBtn = document.getElementById('wa-cc-logout-btn');
				if (logBtn) {
					logBtn.onclick = async function() {
						var msg = 'Keluarkan akun WhatsApp dari komputer ini?\n\n' +
							'Anda perlu scan QR lagi untuk masuk. Pengaturan add-on (pin, balasan cepat, catatan) tetap disimpan.\n\n' +
							'Komputer ini tetap tertaut di HP sampai Anda menghapusnya di WhatsApp > Perangkat tertaut.';
						if (!confirm(msg)) return;
						try {
							// Keep add-on settings (our keys all start with "wa_"), wipe everything WhatsApp stores.
							var keep = {};
							for (var i = 0; i < localStorage.length; i++) {
								var k = localStorage.key(i);
								if (k.indexOf('wa_') === 0) keep[k] = localStorage.getItem(k);
							}
							localStorage.clear();
							sessionStorage.clear();
							Object.keys(keep).forEach(function(k) { localStorage.setItem(k, keep[k]); });
							if (indexedDB.databases) {
								(await indexedDB.databases()).forEach(function(d) { if (d.name) indexedDB.deleteDatabase(d.name); });
							}
							if (window.caches) {
								(await caches.keys()).forEach(function(k) { caches.delete(k); });
							}
							if (navigator.serviceWorker) {
								(await navigator.serviceWorker.getRegistrations()).forEach(function(r) { r.unregister(); });
							}
						} catch(e) {}
						window.location.href = 'https://web.whatsapp.com';
					};
				}

				modal._syncUI = syncPrivacyUI;
					if (window.waBindCoreUI) window.waBindCoreUI(modal);
				return modal;
			}

			window.openAddonControlCenter = function() {
				whenDOMReady(function() {
					var modal = createControlCenterModal();
					if (modal) {
						if (modal._syncUI) modal._syncUI();
						modal.style.display = 'flex';
					}
				});
			};

			// Shortcuts: Alt + M (Control Center), Ctrl + 1 (WhatsApp), Ctrl + 2 (Telegram), Ctrl + 3 (Side-by-Side)
			window.addEventListener('keydown', function(e) {
				if (e.altKey && (e.code === 'KeyM' || e.key === 'm' || e.key === 'M')) {
					if (e.preventDefault) e.preventDefault();
					if (e.stopPropagation) e.stopPropagation();
					window.openAddonControlCenter();
				} else if (e.ctrlKey || e.metaKey) {
					if (e.key === '1' || e.code === 'Digit1') {
						if (e.preventDefault) e.preventDefault();
						if (window.switchToWhatsApp) window.switchToWhatsApp();
						else if (window.openWhatsAppWindow) window.openWhatsAppWindow();
					} else if (e.key === '2' || e.code === 'Digit2') {
						if (e.preventDefault) e.preventDefault();
						if (window.switchToTelegram) window.switchToTelegram();
						else if (window.openTelegramWindow) window.openTelegramWindow();
					} else if (e.key === '3' || e.code === 'Digit3') {
						if (e.preventDefault) e.preventDefault();
						if (window.switchToSplitView) window.switchToSplitView();
						else if (window.openSideBySideView) window.openSideBySideView();
					}
				}
			}, true);

			// 2. Inject Sidebar & Floating Messenger Tabs Dock
			function injectAddonLaunchers() {
				if (!document.body) return;
				if (document.getElementById('wa-messenger-tabs') && document.getElementById('wa-addon-rail-btn')) return;

				// Floating Messenger Tabs Dock at bottom-left
				if (!document.getElementById('wa-messenger-tabs')) {
					var dock = document.createElement('div');
					dock.id = 'wa-messenger-tabs';
					dock.className = 'wa-dock';
					dock.innerHTML = [
						'<button id="wa-tab-cc" class="wa-dock-item" title="Pusat Kontrol (Alt+M)">' +
							'<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="4" y1="21" x2="4" y2="14"/><line x1="4" y1="10" x2="4" y2="3"/><line x1="12" y1="21" x2="12" y2="12"/><line x1="12" y1="8" x2="12" y2="3"/><line x1="20" y1="21" x2="20" y2="16"/><line x1="20" y1="12" x2="20" y2="3"/><line x1="1" y1="14" x2="7" y2="14"/><line x1="9" y1="8" x2="15" y2="8"/><line x1="17" y1="16" x2="23" y2="16"/></svg>' +
							'<span>Fitur</span>' +
						'</button>',
						'<div class="wa-dock-sep"></div>',
						'<button id="wa-tab-wa" class="wa-dock-item active" title="WhatsApp Aktif (Ctrl+1)">' +
							'<span class="wa-status-dot"></span>' +
							'<span>WhatsApp</span>' +
						'</button>',
						'<div class="wa-dock-sep"></div>',
						'<button id="wa-tab-tg" class="wa-dock-item" title="Buka Telegram Web (Ctrl+2)">' +
							'<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>' +
							'<span>Telegram</span>' +
						'</button>',
						'<div class="wa-dock-sep"></div>',
						'<button id="wa-tab-split" class="wa-dock-item" title="Mode Berdampingan 50:50 (Ctrl+3)">' +
							'<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="12" y1="3" x2="12" y2="21"/></svg>' +
							'<span>Berdampingan</span>' +
						'</button>'
					].join('');

					window.syncDockActiveTab = function(mode) {
						var d = document.getElementById('wa-messenger-tabs');
						if (!d) return;
						var bWA = d.querySelector('#wa-tab-wa');
						var bTG = d.querySelector('#wa-tab-tg');
						var bSp = d.querySelector('#wa-tab-split');
						if (bWA) bWA.classList.toggle('active', mode === 'wa');
						if (bTG) bTG.classList.toggle('active', mode === 'tg');
						if (bSp) bSp.classList.toggle('active', mode === 'split');
					};

					var btnCC = dock.querySelector('#wa-tab-cc');
					var btnWA = dock.querySelector('#wa-tab-wa');
					var btnTG = dock.querySelector('#wa-tab-tg');
					var btnSplit = dock.querySelector('#wa-tab-split');

					if (btnCC) {
						btnCC.onclick = function(e) {
							e.preventDefault();
							e.stopPropagation();
							window.openAddonControlCenter();
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

					document.body.appendChild(dock);
				}

				// Try injecting into WhatsApp vertical icon navigation rail
				if (!document.getElementById('wa-addon-rail-btn')) {
					var container = window.waSel ? window.waSel.one('navRail') : null;
					for (var i = 0; i < 1; i++) {
						if (container) {
							var railBtn = document.createElement('button');
							railBtn.id = 'wa-addon-rail-btn';
							railBtn.title = 'Pusat Kontrol (Alt+M)';
							railBtn.style.cssText = 'background:none;border:none;cursor:pointer;width:40px;height:40px;border-radius:50%;display:flex;align-items:center;justify-content:center;color:#00a884;margin:4px auto;outline:none;transition:background 0.2s;';
							railBtn.innerHTML = '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="4" y1="21" x2="4" y2="14"/><line x1="4" y1="10" x2="4" y2="3"/><line x1="12" y1="21" x2="12" y2="12"/><line x1="12" y1="8" x2="12" y2="3"/><line x1="20" y1="21" x2="20" y2="16"/><line x1="20" y1="12" x2="20" y2="3"/><line x1="1" y1="14" x2="7" y2="14"/><line x1="9" y1="8" x2="15" y2="8"/><line x1="17" y1="16" x2="23" y2="16"/></svg>';
							railBtn.onmouseover = function() { railBtn.style.background = 'rgba(0,168,132,0.18)'; };
							railBtn.onmouseout = function() { railBtn.style.background = 'none'; };
							railBtn.onclick = function(e) {
								e.preventDefault();
								e.stopPropagation();
								window.openAddonControlCenter();
							};
							container.appendChild(railBtn);
							break;
						}
					}
				}
			}
			window.ensureDock = injectAddonLaunchers;
			whenDOMReady(function() {
				injectAddonLaunchers();
				setTimeout(injectAddonLaunchers, 2000);
				setTimeout(injectAddonLaunchers, 5000);
				setTimeout(injectAddonLaunchers, 10000);
			});
		})();
