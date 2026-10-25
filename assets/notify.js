// Notification polyfill, shared by the WhatsApp and Telegram windows. Page notifications become Windows
// toasts (notify.go). When a toast is clicked, Go calls window.waNotificationClicked with the
// notification's ID and the page's own click handler runs, which opens the chat like in a browser.
(function() {
	var recent = {};
	var order = [];
	var nextId = 1;

	function NativeNotification(title, options) {
		options = options || {};
		this.title = title;
		this.body = options.body || '';
		this.tag = options.tag || '';
		this.data = options.data;
		this.onclick = null;
		this.onclose = null;
		this.onerror = null;
		this.onshow = null;
		this._listeners = {};

		var id = nextId++;
		recent[id] = this;
		order.push(id);
		if (order.length > 100) delete recent[order.shift()];

		if (window.sendNativeNotificationEx) {
			window.sendNativeNotificationEx(String(title || ''), String(this.body), id);
		} else if (window.sendNativeNotification) {
			window.sendNativeNotification(String(title || ''), String(this.body));
		}
	}
	NativeNotification.prototype.addEventListener = function(type, fn) {
		if (typeof fn !== 'function') return;
		(this._listeners[type] = this._listeners[type] || []).push(fn);
	};
	NativeNotification.prototype.removeEventListener = function(type, fn) {
		var list = this._listeners[type];
		if (list) this._listeners[type] = list.filter(function(f) { return f !== fn; });
	};
	NativeNotification.prototype.close = function() {};
	NativeNotification.permission = 'granted';
	NativeNotification.requestPermission = function(callback) {
		if (typeof callback === 'function') callback('granted');
		return Promise.resolve('granted');
	};
	window.Notification = NativeNotification;

	// Toast clicked: run the page's handlers. Without any, fall back to opening the chat by name.
	window.waNotificationClicked = function(id, chat) {
		var n = recent[id];
		if (n) {
			var ev = { type: 'click', target: n, currentTarget: n, preventDefault: function() {} };
			var handled = false;
			try {
				window.focus();
				if (typeof n.onclick === 'function') { n.onclick.call(n, ev); handled = true; }
				(n._listeners.click || []).forEach(function(fn) { fn.call(n, ev); handled = true; });
			} catch(e) {}
			if (handled) return;
		}
		if (chat && typeof window.waOpenChatByName === 'function') window.waOpenChatByName(chat);
	};
})();
