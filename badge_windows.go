package main

import (
	"strconv"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Taskbar overlay badge: draws a red circle with the unread count and sets it as the
// overlay icon of the main window's taskbar button (ITaskbarList3::SetOverlayIcon).

var (
	ole32  = windows.NewLazySystemDLL("ole32.dll")
	gdi32b = windows.NewLazySystemDLL("gdi32.dll")

	procCoInitializeEx     = ole32.NewProc("CoInitializeEx")
	procCoCreateInstance   = ole32.NewProc("CoCreateInstance")
	procGetDC              = user32.NewProc("GetDC")
	procReleaseDC          = user32.NewProc("ReleaseDC")
	procDrawText           = user32.NewProc("DrawTextW")
	procCreateIconIndirect = user32.NewProc("CreateIconIndirect")
	procDestroyIconB       = user32.NewProc("DestroyIcon")
	procCreateCompatDC     = gdi32b.NewProc("CreateCompatibleDC")
	procCreateCompatBmp    = gdi32b.NewProc("CreateCompatibleBitmap")
	procCreateBitmap       = gdi32b.NewProc("CreateBitmap")
	procSelectObject       = gdi32b.NewProc("SelectObject")
	procDeleteObject       = gdi32b.NewProc("DeleteObject")
	procDeleteDC           = gdi32b.NewProc("DeleteDC")
	procCreateSolidBrush   = gdi32b.NewProc("CreateSolidBrush")
	procEllipse            = gdi32b.NewProc("Ellipse")
	procPatBlt             = gdi32b.NewProc("PatBlt")
	procSetBkMode          = gdi32b.NewProc("SetBkMode")
	procSetTextColor       = gdi32b.NewProc("SetTextColor")
	procCreateFont         = gdi32b.NewProc("CreateFontW")
	procGetStockObject     = gdi32b.NewProc("GetStockObject")
)

type iconInfo struct {
	FIcon    uint32
	XHotspot uint32
	YHotspot uint32
	HbmMask  uintptr
	HbmColor uintptr
}

type rect struct{ Left, Top, Right, Bottom int32 }

var (
	clsidTaskbarList = windows.GUID{Data1: 0x56FDF344, Data2: 0xFD6D, Data3: 0x11d0, Data4: [8]byte{0x95, 0x8A, 0x00, 0x60, 0x97, 0xC9, 0xA0, 0x90}}
	iidTaskbarList3  = windows.GUID{Data1: 0xEA1AFB91, Data2: 0x9E28, Data3: 0x4B86, Data4: [8]byte{0x90, 0xE9, 0x9E, 0x9F, 0x8A, 0x5E, 0xEF, 0xAF}}

	taskbarList uintptr // *ITaskbarList3
	badgeIcon   uintptr // HICON currently set as overlay
	lastBadge   = -1
)

const (
	vtHrInit         = 3
	vtSetOverlayIcon = 18
	badgeSize        = 32
)

// ptrAt reads a pointer-sized value from a raw COM address.
func ptrAt(addr uintptr) uintptr {
	p := addr
	return *(*uintptr)(*(*unsafe.Pointer)(unsafe.Pointer(&p)))
}

func comCall(obj uintptr, index uintptr, args ...uintptr) uintptr {
	vtbl := ptrAt(obj)
	fn := ptrAt(vtbl + index*unsafe.Sizeof(uintptr(0)))
	r, _, _ := syscall.SyscallN(fn, append([]uintptr{obj}, args...)...)
	return r
}

func ensureTaskbarList() bool {
	if taskbarList != 0 {
		return true
	}
	procCoInitializeEx.Call(0, 2) // COINIT_APARTMENTTHREADED; S_FALSE/RPC_E_CHANGED_MODE are fine
	var obj uintptr
	hr, _, _ := procCoCreateInstance.Call(
		uintptr(unsafe.Pointer(&clsidTaskbarList)), 0, 1, // CLSCTX_INPROC_SERVER
		uintptr(unsafe.Pointer(&iidTaskbarList3)), uintptr(unsafe.Pointer(&obj)))
	if hr != 0 || obj == 0 {
		return false
	}
	if comCall(obj, vtHrInit) != 0 {
		return false
	}
	taskbarList = obj
	return true
}

func rgb(r, g, b byte) uintptr { return uintptr(uint32(r) | uint32(g)<<8 | uint32(b)<<16) }

// makeBadgeIcon renders a red circle with white text into a 32x32 icon.
func makeBadgeIcon(text string) uintptr {
	hdcScreen, _, _ := procGetDC.Call(0)
	defer procReleaseDC.Call(0, hdcScreen)

	// Colour bitmap
	dcColor, _, _ := procCreateCompatDC.Call(hdcScreen)
	defer procDeleteDC.Call(dcColor)
	bmColor, _, _ := procCreateCompatBmp.Call(hdcScreen, badgeSize, badgeSize)
	oldC, _, _ := procSelectObject.Call(dcColor, bmColor)
	blackBrush, _, _ := procGetStockObject.Call(4) // BLACK_BRUSH
	procSelectObject.Call(dcColor, blackBrush)
	procPatBlt.Call(dcColor, 0, 0, badgeSize, badgeSize, 0x00000042) // BLACKNESS
	red, _, _ := procCreateSolidBrush.Call(rgb(220, 38, 38))
	oldBrush, _, _ := procSelectObject.Call(dcColor, red)
	nullPen, _, _ := procGetStockObject.Call(8) // NULL_PEN
	oldPen, _, _ := procSelectObject.Call(dcColor, nullPen)
	procEllipse.Call(dcColor, 0, 0, badgeSize+1, badgeSize+1)
	procSelectObject.Call(dcColor, oldPen)
	procSelectObject.Call(dcColor, oldBrush)
	procDeleteObject.Call(red)

	height := -22
	if len(text) == 2 {
		height = -18
	} else if len(text) >= 3 {
		height = -14
	}
	face, _ := windows.UTF16PtrFromString("Segoe UI")
	font, _, _ := procCreateFont.Call(uintptr(height), 0, 0, 0, 700, 0, 0, 0, 1, 0, 0, 4, 0, uintptr(unsafe.Pointer(face)))
	oldFont, _, _ := procSelectObject.Call(dcColor, font)
	procSetBkMode.Call(dcColor, 1) // TRANSPARENT
	procSetTextColor.Call(dcColor, rgb(255, 255, 255))
	txt, _ := windows.UTF16FromString(text)
	rc := rect{0, 0, badgeSize, badgeSize}
	procDrawText.Call(dcColor, uintptr(unsafe.Pointer(&txt[0])), uintptr(len(txt)-1), uintptr(unsafe.Pointer(&rc)), 0x25) // DT_CENTER|DT_VCENTER|DT_SINGLELINE
	procSelectObject.Call(dcColor, oldFont)
	procDeleteObject.Call(font)
	procSelectObject.Call(dcColor, oldC)

	// Monochrome mask: white = transparent, black circle = opaque
	dcMask, _, _ := procCreateCompatDC.Call(hdcScreen)
	defer procDeleteDC.Call(dcMask)
	bmMask, _, _ := procCreateBitmap.Call(badgeSize, badgeSize, 1, 1, 0)
	oldM, _, _ := procSelectObject.Call(dcMask, bmMask)
	procPatBlt.Call(dcMask, 0, 0, badgeSize, badgeSize, 0x00FF0062) // WHITENESS
	procSelectObject.Call(dcMask, blackBrush)
	procSelectObject.Call(dcMask, nullPen)
	procEllipse.Call(dcMask, 0, 0, badgeSize+1, badgeSize+1)
	procSelectObject.Call(dcMask, oldM)

	ii := iconInfo{FIcon: 1, HbmMask: bmMask, HbmColor: bmColor}
	icon, _, _ := procCreateIconIndirect.Call(uintptr(unsafe.Pointer(&ii)))
	procDeleteObject.Call(bmColor)
	procDeleteObject.Call(bmMask)
	return icon
}

// setTaskbarBadge shows count on the taskbar button; count <= 0 clears the badge.
func setTaskbarBadge(count int) {
	if globalHwnd == 0 || count == lastBadge || !ensureTaskbarList() {
		return
	}
	lastBadge = count

	var icon uintptr
	var desc *uint16
	if count > 0 {
		label := strconv.Itoa(count)
		if count > 99 {
			label = "99+"
		}
		icon = makeBadgeIcon(label)
		desc, _ = windows.UTF16PtrFromString(strconv.Itoa(count) + " pesan belum dibaca")
	}
	comCall(taskbarList, vtSetOverlayIcon, globalHwnd, icon, uintptr(unsafe.Pointer(desc)))
	if badgeIcon != 0 {
		procDestroyIconB.Call(badgeIcon)
	}
	badgeIcon = icon
}
