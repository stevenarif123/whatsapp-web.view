// Download panel, shared by the WhatsApp and Telegram windows. Downloads are saved through the Windows
// Save As dialog by Go (downloads.go), which reports progress here via window.waDownloadProgress.
(function() {
	if (window.self !== window.top) return;

	var rows = {};

	function formatBytes(n) {
		if (!(n >= 0)) return '?';
		if (n < 1024) return n + ' B';
		if (n < 1024 * 1024) return (n / 1024).toFixed(0) + ' KB';
		if (n < 1024 * 1024 * 1024) return (n / 1024 / 1024).toFixed(1) + ' MB';
		return (n / 1024 / 1024 / 1024).toFixed(2) + ' GB';
	}

	function getPanel() {
		var panel = document.getElementById('wa-dl-panel');
		if (panel) return panel;
		var root = document.body || document.documentElement;
		if (!root) return null;
		panel = document.createElement('div');
		panel.id = 'wa-dl-panel';
		panel.style.cssText = 'position:fixed;right:16px;bottom:16px;width:300px;display:flex;flex-direction:column;gap:8px;z-index:2147483645;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;pointer-events:none;';
		root.appendChild(panel);
		return panel;
	}

	function button(label, onClick) {
		var b = document.createElement('button');
		b.textContent = label;
		b.style.cssText = 'background:none;border:1px solid rgba(255,255,255,0.18);color:#e9edef;border-radius:6px;padding:3px 9px;font-size:11px;cursor:pointer;font-family:inherit;';
		b.onclick = function(e) {
			e.preventDefault();
			e.stopPropagation();
			onClick();
		};
		return b;
	}

	function createRow(p) {
		var row = document.createElement('div');
		row.style.cssText = 'pointer-events:auto;background:rgba(17,27,33,0.97);border:1px solid rgba(134,150,160,0.25);border-radius:10px;padding:10px 12px;box-shadow:0 10px 30px rgba(0,0,0,0.55);color:#e9edef;font-size:12px;';
		var top = document.createElement('div');
		top.style.cssText = 'display:flex;align-items:center;gap:8px;';
		var name = document.createElement('div');
		name.style.cssText = 'flex:1;min-width:0;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;font-weight:600;';
		var actions = document.createElement('div');
		actions.style.cssText = 'display:flex;gap:6px;flex-shrink:0;';
		top.appendChild(name);
		top.appendChild(actions);
		var bar = document.createElement('div');
		bar.style.cssText = 'height:4px;background:rgba(255,255,255,0.08);border-radius:2px;margin:8px 0 5px 0;overflow:hidden;';
		var fill = document.createElement('div');
		fill.style.cssText = 'height:100%;width:0%;background:#00a884;transition:width 0.2s ease;';
		bar.appendChild(fill);
		var info = document.createElement('div');
		info.style.cssText = 'color:#8696a0;font-size:11px;';
		row.appendChild(top);
		row.appendChild(bar);
		row.appendChild(info);
		var panel = getPanel();
		if (panel) panel.appendChild(row);
		return { el: row, name: name, actions: actions, bar: bar, fill: fill, info: info, state: '' };
	}

	function removeLater(id, ms) {
		setTimeout(function() {
			var r = rows[id];
			if (!r) return;
			if (r.el.parentNode) r.el.parentNode.removeChild(r.el);
			delete rows[id];
		}, ms);
	}

	window.waDownloadProgress = function(p) {
		if (!p || !p.id) return;
		var r = rows[p.id];
		if (!r) r = rows[p.id] = createRow(p);
		r.name.textContent = p.name || 'Unduhan';
		r.name.title = p.name || '';
		var known = p.total > 0;
		var pct = known ? Math.min(100, Math.round(p.received / p.total * 100)) : 0;

		if (p.state === 'progress') {
			r.fill.style.width = known ? pct + '%' : '30%';
			r.info.textContent = known ?
				formatBytes(p.received) + ' / ' + formatBytes(p.total) + ' (' + pct + '%)' :
				formatBytes(p.received) + ' diunduh';
			if (r.state !== 'progress') {
				r.actions.textContent = '';
				r.actions.appendChild(button('Batal', function() {
					if (window.cancelDownload) window.cancelDownload(p.id);
				}));
			}
		} else if (p.state === 'completed') {
			r.fill.style.width = '100%';
			r.info.textContent = 'Selesai • ' + formatBytes(p.received);
			r.actions.textContent = '';
			r.actions.appendChild(button('Buka folder', function() {
				if (window.showDownloadInFolder) window.showDownloadInFolder(p.id);
			}));
			removeLater(p.id, 10000);
		} else {
			r.fill.style.background = '#ef4444';
			r.info.textContent = p.state === 'cancelled' ? 'Dibatalkan' : 'Gagal mengunduh';
			r.actions.textContent = '';
			removeLater(p.id, 5000);
		}
		r.state = p.state;
	};
})();
