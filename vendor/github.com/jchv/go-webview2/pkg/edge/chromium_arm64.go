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

	words := (*[2]uintptr)(unsafe.Pointer(&bounds))
	e.controller.vtbl.PutBounds.Call(
		uintptr(unsafe.Pointer(e.controller)),
		words[0],
		words[1],
	)
}
