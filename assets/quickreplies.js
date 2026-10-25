// Quick replies shared by the WhatsApp and Telegram windows. Go keeps the list (notes.go); each page
// keeps a copy in localStorage['wa_quick_replies'], which the quick reply code reads synchronously.
// The copy is refreshed when the page loads and whenever its window gets focus, so a reply added in
// WhatsApp works in Telegram right away.
(function() {
	if (window.self !== window.top) return;
	var KEY = 'wa_quick_replies';

	function localList() {
		try {
			var parsed = JSON.parse(localStorage.getItem(KEY) || 'null');
			return Array.isArray(parsed) ? parsed : null;
		} catch(e) { return null; }
	}

	window.waPushQuickReplies = function(list) {
		if (window.saveSharedQuickReplies) window.saveSharedQuickReplies(JSON.stringify(list || []));
	};

	function pull() {
		if (!window.getSharedQuickReplies) return;
		window.getSharedQuickReplies().then(function(raw) {
			var shared = null;
			try { shared = raw ? JSON.parse(raw) : null; } catch(e) {}
			if (Array.isArray(shared)) {
				try { localStorage.setItem(KEY, JSON.stringify(shared)); } catch(e) {}
				if (typeof window.waOnQuickRepliesChanged === 'function') window.waOnQuickRepliesChanged();
			} else {
				// Nothing shared yet: offer this page's own list (from before replies were shared).
				var mine = localList();
				if (mine && mine.length) window.waPushQuickReplies(mine);
			}
		}).catch(function() {});
	}

	// Bindings are installed before page scripts but may not be ready on the very first tick.
	setTimeout(pull, 0);
	setTimeout(pull, 1500);
	window.addEventListener('focus', pull);
})();
