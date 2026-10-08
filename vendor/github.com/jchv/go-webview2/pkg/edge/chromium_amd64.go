//go:build windows
// +build windows

package edge

import (
	"unsafe"

)

func (e *Chromium) Resize() {
	if e.controller == nil {
		return
	}
	bounds := e.currentBounds()
	_, _, _ = e.controller.vtbl.PutBounds.Call(
		uintptr(unsafe.Pointer(e.controller)),
		uintptr(unsafe.Pointer(&bounds)),
	)
}
