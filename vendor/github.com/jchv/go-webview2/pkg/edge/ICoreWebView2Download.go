package edge

// Download support (ICoreWebView2_4.DownloadStarting and ICoreWebView2DownloadOperation).
// Added for whatsapp-desktop; not part of upstream go-webview2. IIDs and vtable order are taken from
// WebView2.h of the WebView2 SDK.

import (
	"math"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	COREWEBVIEW2_DOWNLOAD_STATE_IN_PROGRESS = 0
	COREWEBVIEW2_DOWNLOAD_STATE_INTERRUPTED = 1
	COREWEBVIEW2_DOWNLOAD_STATE_COMPLETED   = 2

	COREWEBVIEW2_DOWNLOAD_INTERRUPT_REASON_USER_CANCELED = 26
)

func comBool(b bool) uintptr {
	if b {
		return 1
	}
	return 0
}

// comString reads a CoTaskMem-allocated string returned by a getter and frees it.
func comString(p *uint16) string {
	if p == nil {
		return ""
	}
	s := windows.UTF16PtrToString(p)
	windows.CoTaskMemFree(unsafe.Pointer(p))
	return s
}

// ICoreWebView2_4

type iCoreWebView2_4Vtbl struct {
	iCoreWebView2_3Vtbl
	AddFrameCreated        ComProc
	RemoveFrameCreated     ComProc
	AddDownloadStarting    ComProc
	RemoveDownloadStarting ComProc
}

type ICoreWebView2_4 struct {
	vtbl *iCoreWebView2_4Vtbl
}

func (i *ICoreWebView2_4) Release() {
	_, _, _ = i.vtbl.Release.Call(uintptr(unsafe.Pointer(i)))
}

func (i *ICoreWebView2) GetICoreWebView2_4() *ICoreWebView2_4 {
	var result *ICoreWebView2_4

	iidICoreWebView2_4 := NewGUID("{20D02D59-6DF2-42DC-BD06-F98A694B1302}")
	_, _, _ = i.vtbl.QueryInterface.Call(
		uintptr(unsafe.Pointer(i)),
		uintptr(unsafe.Pointer(iidICoreWebView2_4)),
		uintptr(unsafe.Pointer(&result)))

	return result
}

// ICoreWebView2DownloadStartingEventArgs

type _ICoreWebView2DownloadStartingEventArgsVtbl struct {
	_IUnknownVtbl
	GetDownloadOperation ComProc
	GetCancel            ComProc
	PutCancel            ComProc
	GetResultFilePath    ComProc
	PutResultFilePath    ComProc
	GetHandled           ComProc
	PutHandled           ComProc
	GetDeferral          ComProc
}

type ICoreWebView2DownloadStartingEventArgs struct {
	vtbl *_ICoreWebView2DownloadStartingEventArgsVtbl
}

func (i *ICoreWebView2DownloadStartingEventArgs) AddRef() {
	_, _, _ = i.vtbl.AddRef.Call(uintptr(unsafe.Pointer(i)))
}

func (i *ICoreWebView2DownloadStartingEventArgs) Release() {
	_, _, _ = i.vtbl.Release.Call(uintptr(unsafe.Pointer(i)))
}

// DownloadOperation returns the operation with a reference the caller must Release.
func (i *ICoreWebView2DownloadStartingEventArgs) DownloadOperation() *ICoreWebView2DownloadOperation {
	var op *ICoreWebView2DownloadOperation
	hr, _, _ := i.vtbl.GetDownloadOperation.Call(uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(&op)))
	if int32(hr) < 0 {
		return nil
	}
	return op
}

func (i *ICoreWebView2DownloadStartingEventArgs) ResultFilePath() string {
	var p *uint16
	hr, _, _ := i.vtbl.GetResultFilePath.Call(uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(&p)))
	if int32(hr) < 0 {
		return ""
	}
	return comString(p)
}

func (i *ICoreWebView2DownloadStartingEventArgs) PutResultFilePath(path string) bool {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return false
	}
	hr, _, _ := i.vtbl.PutResultFilePath.Call(uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(p)))
	return int32(hr) >= 0
}

func (i *ICoreWebView2DownloadStartingEventArgs) PutCancel(cancel bool) {
	_, _, _ = i.vtbl.PutCancel.Call(uintptr(unsafe.Pointer(i)), comBool(cancel))
}

// PutHandled true hides the default download UI.
func (i *ICoreWebView2DownloadStartingEventArgs) PutHandled(handled bool) {
	_, _, _ = i.vtbl.PutHandled.Call(uintptr(unsafe.Pointer(i)), comBool(handled))
}

func (i *ICoreWebView2DownloadStartingEventArgs) GetDeferral() *ICoreWebView2Deferral {
	var d *ICoreWebView2Deferral
	hr, _, _ := i.vtbl.GetDeferral.Call(uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(&d)))
	if int32(hr) < 0 {
		return nil
	}
	return d
}

// ICoreWebView2Deferral

type _ICoreWebView2DeferralVtbl struct {
	_IUnknownVtbl
	Complete ComProc
}

type ICoreWebView2Deferral struct {
	vtbl *_ICoreWebView2DeferralVtbl
}

// Complete finishes the deferred event and releases the deferral.
func (i *ICoreWebView2Deferral) Complete() {
	_, _, _ = i.vtbl.Complete.Call(uintptr(unsafe.Pointer(i)))
	_, _, _ = i.vtbl.Release.Call(uintptr(unsafe.Pointer(i)))
}

// ICoreWebView2DownloadOperation

type _ICoreWebView2DownloadOperationVtbl struct {
	_IUnknownVtbl
	AddBytesReceivedChanged       ComProc
	RemoveBytesReceivedChanged    ComProc
	AddEstimatedEndTimeChanged    ComProc
	RemoveEstimatedEndTimeChanged ComProc
	AddStateChanged               ComProc
	RemoveStateChanged            ComProc
	GetUri                        ComProc
	GetContentDisposition         ComProc
	GetMimeType                   ComProc
	GetTotalBytesToReceive        ComProc
	GetBytesReceived              ComProc
	GetEstimatedEndTime           ComProc
	GetResultFilePath             ComProc
	GetState                      ComProc
	GetInterruptReason            ComProc
	Cancel                        ComProc
	Pause                         ComProc
	Resume                        ComProc
	GetCanResume                  ComProc
}

type ICoreWebView2DownloadOperation struct {
	vtbl *_ICoreWebView2DownloadOperationVtbl
}

func (i *ICoreWebView2DownloadOperation) Release() {
	_, _, _ = i.vtbl.Release.Call(uintptr(unsafe.Pointer(i)))
}

func (i *ICoreWebView2DownloadOperation) stringProp(proc ComProc) string {
	var p *uint16
	hr, _, _ := proc.Call(uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(&p)))
	if int32(hr) < 0 {
		return ""
	}
	return comString(p)
}

func (i *ICoreWebView2DownloadOperation) int64Prop(proc ComProc) int64 {
	var v int64
	hr, _, _ := proc.Call(uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(&v)))
	if int32(hr) < 0 {
		return -1
	}
	return v
}

func (i *ICoreWebView2DownloadOperation) int32Prop(proc ComProc) int32 {
	var v int32
	_, _, _ = proc.Call(uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(&v)))
	return v
}

func (i *ICoreWebView2DownloadOperation) Uri() string      { return i.stringProp(i.vtbl.GetUri) }
func (i *ICoreWebView2DownloadOperation) MimeType() string { return i.stringProp(i.vtbl.GetMimeType) }
func (i *ICoreWebView2DownloadOperation) ResultFilePath() string {
	return i.stringProp(i.vtbl.GetResultFilePath)
}

// TotalBytesToReceive is -1 when the size is unknown.
func (i *ICoreWebView2DownloadOperation) TotalBytesToReceive() int64 {
	return i.int64Prop(i.vtbl.GetTotalBytesToReceive)
}
func (i *ICoreWebView2DownloadOperation) BytesReceived() int64 {
	return i.int64Prop(i.vtbl.GetBytesReceived)
}
func (i *ICoreWebView2DownloadOperation) State() int32 { return i.int32Prop(i.vtbl.GetState) }
func (i *ICoreWebView2DownloadOperation) InterruptReason() int32 {
	return i.int32Prop(i.vtbl.GetInterruptReason)
}

func (i *ICoreWebView2DownloadOperation) Cancel() {
	_, _, _ = i.vtbl.Cancel.Call(uintptr(unsafe.Pointer(i)))
}

func (i *ICoreWebView2DownloadOperation) AddBytesReceivedChanged(h *DownloadOperationEventHandler) (EventToken, bool) {
	var token EventToken
	hr, _, _ := i.vtbl.AddBytesReceivedChanged.Call(uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(h)), uintptr(unsafe.Pointer(&token)))
	return token, int32(hr) >= 0
}

func (i *ICoreWebView2DownloadOperation) RemoveBytesReceivedChanged(token EventToken) {
	_, _, _ = i.vtbl.RemoveBytesReceivedChanged.Call(uintptr(unsafe.Pointer(i)), uintptr(token.value))
}

func (i *ICoreWebView2DownloadOperation) AddStateChanged(h *DownloadOperationEventHandler) (EventToken, bool) {
	var token EventToken
	hr, _, _ := i.vtbl.AddStateChanged.Call(uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(h)), uintptr(unsafe.Pointer(&token)))
	return token, int32(hr) >= 0
}

func (i *ICoreWebView2DownloadOperation) RemoveStateChanged(token EventToken) {
	_, _, _ = i.vtbl.RemoveStateChanged.Call(uintptr(unsafe.Pointer(i)), uintptr(token.value))
}

// EventToken identifies a registered event handler (EventRegistrationToken).
type EventToken struct {
	value int64
}

// Event handlers implemented in Go. The handler objects are handed to WebView2 as raw pointers, so the
// caller must keep them referenced until they are removed from the event they were added to.

func comQueryInterface(this unsafe.Pointer, _ uintptr, object *unsafe.Pointer) uintptr {
	*object = this
	return 0
}

func comAddRefRelease(unsafe.Pointer) uintptr {
	return 1
}

// DownloadStartingEventHandler (ICoreWebView2DownloadStartingEventHandler)

type _downloadStartingEventHandlerVtbl struct {
	_IUnknownVtbl
	Invoke ComProc
}

type DownloadStartingEventHandler struct {
	vtbl *_downloadStartingEventHandlerVtbl
	fn   func(args *ICoreWebView2DownloadStartingEventArgs)
}

var _downloadStartingEventHandlerFn = _downloadStartingEventHandlerVtbl{
	_IUnknownVtbl{
		NewComProc(comQueryInterface),
		NewComProc(comAddRefRelease),
		NewComProc(comAddRefRelease),
	},
	NewComProc(func(this *DownloadStartingEventHandler, _ *ICoreWebView2, args *ICoreWebView2DownloadStartingEventArgs) uintptr {
		if this.fn != nil {
			this.fn(args)
		}
		return 0
	}),
}

func NewDownloadStartingEventHandler(fn func(args *ICoreWebView2DownloadStartingEventArgs)) *DownloadStartingEventHandler {
	return &DownloadStartingEventHandler{vtbl: &_downloadStartingEventHandlerFn, fn: fn}
}

// DownloadOperationEventHandler is used for both BytesReceivedChanged and StateChanged, which share
// the same Invoke signature.

type _downloadOperationEventHandlerVtbl struct {
	_IUnknownVtbl
	Invoke ComProc
}

type DownloadOperationEventHandler struct {
	vtbl *_downloadOperationEventHandlerVtbl
	fn   func(op *ICoreWebView2DownloadOperation)
}

var _downloadOperationEventHandlerFn = _downloadOperationEventHandlerVtbl{
	_IUnknownVtbl{
		NewComProc(comQueryInterface),
		NewComProc(comAddRefRelease),
		NewComProc(comAddRefRelease),
	},
	NewComProc(func(this *DownloadOperationEventHandler, sender *ICoreWebView2DownloadOperation, _ uintptr) uintptr {
		if this.fn != nil {
			this.fn(sender)
		}
		return 0
	}),
}

func NewDownloadOperationEventHandler(fn func(op *ICoreWebView2DownloadOperation)) *DownloadOperationEventHandler {
	return &DownloadOperationEventHandler{vtbl: &_downloadOperationEventHandlerFn, fn: fn}
}

// AddDownloadStarting registers h on the webview. It returns false when the installed WebView2
// runtime is too old for downloads events (before ICoreWebView2_4).
func (e *Chromium) AddDownloadStarting(h *DownloadStartingEventHandler) bool {
	if e.webview == nil {
		return false
	}
	wv4 := e.webview.GetICoreWebView2_4()
	if wv4 == nil {
		return false
	}
	defer wv4.Release()
	var token EventToken
	hr, _, _ := wv4.vtbl.AddDownloadStarting.Call(uintptr(unsafe.Pointer(wv4)), uintptr(unsafe.Pointer(h)), uintptr(unsafe.Pointer(&token)))
	return int32(hr) >= 0
}

// Zoom (ICoreWebView2Controller.ZoomFactor). Unlike CSS zoom, this is the browser's own zoom: the page
// lays itself out again for the smaller or larger viewport. Added for whatsapp-desktop.

func (i *ICoreWebView2Controller) PutZoomFactor(factor float64) error {
	// Windows x64 passes the double in XMM1; Go's stdcall bridge loads the argument registers into XMM too.
	hr, _, _ := i.vtbl.PutZoomFactor.Call(uintptr(unsafe.Pointer(i)), uintptr(math.Float64bits(factor)))
	if int32(hr) < 0 {
		return windows.Errno(hr)
	}
	return nil
}

func (i *ICoreWebView2Controller) GetZoomFactor() (float64, error) {
	var factor float64
	hr, _, _ := i.vtbl.GetZoomFactor.Call(uintptr(unsafe.Pointer(i)), uintptr(unsafe.Pointer(&factor)))
	if int32(hr) < 0 {
		return 0, windows.Errno(hr)
	}
	return factor, nil
}
