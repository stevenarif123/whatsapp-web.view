package webview2

// Download handling, added for whatsapp-desktop (not part of upstream go-webview2).

import (
	"time"

	"github.com/jchv/go-webview2/pkg/edge"
)

// DownloadInfo describes a download that is about to start.
type DownloadInfo struct {
	ID            int
	URI           string
	MimeType      string
	SuggestedPath string // where WebView2 would save the file by default
	TotalBytes    int64  // -1 when unknown
}

// DownloadProgress reports a running or finished download.
type DownloadProgress struct {
	ID       int
	Path     string
	Received int64
	Total    int64  // -1 when unknown
	State    string // "progress", "completed", "cancelled" or "failed"
}

type download struct {
	op           *edge.ICoreWebView2DownloadOperation
	path         string
	bytesHandler *edge.DownloadOperationEventHandler
	stateHandler *edge.DownloadOperationEventHandler
	bytesToken   edge.EventToken
	stateToken   edge.EventToken
	lastReport   time.Time
}

// downloadSeq numbers downloads across all webviews of the process.
var downloadSeq int

const downloadReportInterval = 250 * time.Millisecond

// SetDownloadHandler routes the page's downloads through choose, which runs on the UI thread and may
// show a modal dialog. It returns the path to save to, or "" to cancel the download. The default
// download UI is hidden; progress (optional) is called on the UI thread instead. It returns false when
// the WebView2 runtime is too old to support download events, in which case downloads keep their
// default behaviour.
func (w *webview) SetDownloadHandler(choose func(DownloadInfo) string, progress func(DownloadProgress)) bool {
	c, ok := w.browser.(*edge.Chromium)
	if !ok {
		return false
	}
	w.downloadStarting = edge.NewDownloadStartingEventHandler(func(args *edge.ICoreWebView2DownloadStartingEventArgs) {
		op := args.DownloadOperation()
		if op == nil {
			return
		}
		deferral := args.GetDeferral()
		if deferral == nil {
			op.Release()
			return
		}
		downloadSeq++
		info := DownloadInfo{
			ID:            downloadSeq,
			URI:           op.Uri(),
			MimeType:      op.MimeType(),
			SuggestedPath: args.ResultFilePath(),
			TotalBytes:    op.TotalBytesToReceive(),
		}
		args.AddRef()
		// The event handler must return before a modal dialog runs, so choose is called from the queue.
		w.Dispatch(func() {
			path := choose(info)
			if path == "" || !args.PutResultFilePath(path) {
				args.PutCancel(true)
				op.Release()
			} else {
				args.PutHandled(true)
				w.trackDownload(info.ID, op, path, progress)
			}
			deferral.Complete()
			args.Release()
		})
	})
	return c.AddDownloadStarting(w.downloadStarting)
}

func (w *webview) trackDownload(id int, op *edge.ICoreWebView2DownloadOperation, path string, progress func(DownloadProgress)) {
	if w.downloads == nil {
		w.downloads = map[int]*download{}
	}
	d := &download{op: op, path: path}
	w.downloads[id] = d

	report := func(force bool) {
		state := op.State()
		if state == edge.COREWEBVIEW2_DOWNLOAD_STATE_IN_PROGRESS && !force && time.Since(d.lastReport) < downloadReportInterval {
			return
		}
		d.lastReport = time.Now()
		p := DownloadProgress{ID: id, Path: path, Received: op.BytesReceived(), Total: op.TotalBytesToReceive(), State: "progress"}
		switch state {
		case edge.COREWEBVIEW2_DOWNLOAD_STATE_COMPLETED:
			p.State = "completed"
		case edge.COREWEBVIEW2_DOWNLOAD_STATE_INTERRUPTED:
			p.State = "failed"
			if op.InterruptReason() == edge.COREWEBVIEW2_DOWNLOAD_INTERRUPT_REASON_USER_CANCELED {
				p.State = "cancelled"
			}
		}
		if progress != nil {
			progress(p)
		}
		if state != edge.COREWEBVIEW2_DOWNLOAD_STATE_IN_PROGRESS {
			w.finishDownload(id)
		}
	}

	d.bytesHandler = edge.NewDownloadOperationEventHandler(func(*edge.ICoreWebView2DownloadOperation) { report(false) })
	d.stateHandler = edge.NewDownloadOperationEventHandler(func(*edge.ICoreWebView2DownloadOperation) { report(true) })
	d.bytesToken, _ = op.AddBytesReceivedChanged(d.bytesHandler)
	d.stateToken, _ = op.AddStateChanged(d.stateHandler)
	report(true)
}

// finishDownload unregisters the handlers of a finished download and drops the operation.
func (w *webview) finishDownload(id int) {
	d, ok := w.downloads[id]
	if !ok {
		return
	}
	delete(w.downloads, id)
	d.op.RemoveBytesReceivedChanged(d.bytesToken)
	d.op.RemoveStateChanged(d.stateToken)
	d.op.Release()
	// Keep the handler objects alive until WebView2 has let go of them on this turn of the loop.
	w.Dispatch(func() { _ = d })
}

// CancelDownload stops a running download started by this webview.
func (w *webview) CancelDownload(id int) {
	if d, ok := w.downloads[id]; ok {
		d.op.Cancel()
	}
}

// SetZoomFactor sets the browser zoom (1.0 = 100%) and reports the factor now in effect.
func (w *webview) SetZoomFactor(factor float64) float64 {
	c, ok := w.browser.(*edge.Chromium)
	if !ok || c.GetController() == nil {
		return 0
	}
	ctl := c.GetController()
	_ = ctl.PutZoomFactor(factor)
	got, _ := ctl.GetZoomFactor()
	return got
}
