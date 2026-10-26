// Reports WhatsApp Web's state to Go every 30 s (watchdog.go), which reloads the page when the reports
// stop or the page stays on the loading screen. "qr" (not logged in) is not a problem to fix.
(function() {
	if (window.self !== window.top || location.hostname !== 'web.whatsapp.com') return;

	function state() {
		if (!navigator.onLine) return 'offline';
		if (document.getElementById('pane-side')) return 'chats';
		if (document.querySelector('canvas[aria-label], [data-ref] canvas, [data-testid="qrcode"]')) return 'qr';
		return 'loading';
	}

	function report() {
		if (typeof window.waHeartbeat === 'function') {
			try { window.waHeartbeat(state()); } catch (e) {}
		}
	}

	setInterval(report, 30000);
	setTimeout(report, 5000);
	window.addEventListener('online', report);
	window.addEventListener('offline', report);

	// Shown after Go reloaded the page on its own
	window.waOnAutoRecovered = function(reason) {
		var msg = 'WhatsApp dimuat ulang otomatis (' + reason + ')';
		if (typeof showAddonToast === 'function') showAddonToast('🔄 ' + msg);
		try {
			var log = JSON.parse(localStorage.getItem('wa_recovery_log') || '[]');
			log.unshift({ at: Date.now(), reason: reason });
			localStorage.setItem('wa_recovery_log', JSON.stringify(log.slice(0, 20)));
		} catch (e) {}
	};
})();
