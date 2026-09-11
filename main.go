package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unicode"
	"unsafe"

	"github.com/go-toast/toast"
	"github.com/jchv/go-webview2"
	"golang.org/x/sys/windows"
)

var (
	kernel32             = windows.NewLazySystemDLL("kernel32.dll")
	user32               = windows.NewLazySystemDLL("user32.dll")
	dwmapi               = windows.NewLazySystemDLL("dwmapi.dll")
	shell32              = windows.NewLazySystemDLL("shell32.dll")
	advapi32             = windows.NewLazySystemDLL("advapi32.dll")
	procCreateMutex      = kernel32.NewProc("CreateMutexW")
	procFindWindow       = user32.NewProc("FindWindowW")
	procSetFgWindow      = user32.NewProc("SetForegroundWindow")
	procGetFgWindow      = user32.NewProc("GetForegroundWindow")
	procShowNormal       = user32.NewProc("ShowWindow")
	procShowWindow       = user32.NewProc("ShowWindow")
	procDestroyWindow    = user32.NewProc("DestroyWindow")
	procDwmSetAttr       = dwmapi.NewProc("DwmSetWindowAttribute")
	procShellExecute     = shell32.NewProc("ShellExecuteW")
	procShellNotifyIcon  = shell32.NewProc("Shell_NotifyIconW")
	procFlashWindowEx    = user32.NewProc("FlashWindowEx")
	procCreatePopupMenu  = user32.NewProc("CreatePopupMenu")
	procAppendMenu       = user32.NewProc("AppendMenuW")
	procTrackPopupMenu   = user32.NewProc("TrackPopupMenu")
	procDestroyMenu      = user32.NewProc("DestroyMenu")
	procGetCursorPos     = user32.NewProc("GetCursorPos")
	procSetWindowLongPtr = user32.NewProc("SetWindowLongPtrW")
	procCallWindowProc   = user32.NewProc("CallWindowProcW")
	procLoadImage        = user32.NewProc("LoadImageW")
	procSendMessage      = user32.NewProc("SendMessageW")
	procPostMessage      = user32.NewProc("PostMessageW")
	procRegisterHotKey   = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey = user32.NewProc("UnregisterHotKey")
	procIsWindowVisible  = user32.NewProc("IsWindowVisible")
	procRegSetValueEx    = advapi32.NewProc("RegSetValueExW")
	procRegDeleteVal     = advapi32.NewProc("RegDeleteValueW")
	procMoveWindow       = user32.NewProc("MoveWindow")
	procSystemParamsInfo = user32.NewProc("SystemParametersInfoW")
	procSetParent        = user32.NewProc("SetParent")
	procGetWindowLongPtr = user32.NewProc("GetWindowLongPtrW")
	procSetWindowPos     = user32.NewProc("SetWindowPos")
	procGetClientRect    = user32.NewProc("GetClientRect")
	procFindWindowEx     = user32.NewProc("FindWindowExW")
	procSetFocus         = user32.NewProc("SetFocus")
)

const (
	windowTitle = "WhatsApp Desktop"
	appURL      = "https://web.whatsapp.com"
	mutexName   = "WhatsAppDesktopSingleInstanceMutex"
	userAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"

	telegramWindowTitle = "Telegram Web Desktop"
	telegramAppURL      = "https://web.telegram.org/k/"
	telegramMutexName   = "TelegramDesktopSingleInstanceMutex"

	SPI_GETWORKAREA = 0x0030

	// Registry constants for Windows Startup
	HKEY_CURRENT_USER = 0x80000001
	KEY_READ          = 0x20019
	KEY_WRITE         = 0x20006
	REG_SZ            = 1
	runKeyPath        = `Software\Microsoft\Windows\CurrentVersion\Run`
	runKeyName        = `WhatsAppDesktopLight`

	// Boss Key Global Hotkey (Ctrl + Alt + W)
	HOTKEY_BOSS  = 1001
	MOD_ALT      = 0x0001
	MOD_CONTROL  = 0x0002
	MOD_NOREPEAT = 0x4000
	WM_HOTKEY    = 0x0312

	// DWM Window Attributes for Dark Theme
	DWMWA_USE_IMMERSIVE_DARK_MODE_BEFORE_20H1 = 19
	DWMWA_USE_IMMERSIVE_DARK_MODE             = 20
	DWMWA_CAPTION_COLOR                      = 35
	DWMWA_TEXT_COLOR                         = 36

	// Win32 Window Messages & Constants
	WM_DESTROY       = 0x0002
	WM_SIZE          = 0x0005
	WM_CLOSE         = 0x0010
	WM_COMMAND       = 0x0111
	WM_GETICON       = 0x007F
	WM_LBUTTONUP     = 0x0202
	WM_LBUTTONDBLCLK = 0x0203
	WM_RBUTTONUP     = 0x0205
	WM_APP           = 0x8000
	WM_TRAYICON      = WM_APP + 101

	GWLP_WNDPROC = -4

	GWL_STYLE       = -16
	GWL_EXSTYLE     = -20
	WS_CHILD        = 0x40000000
	WS_POPUP        = 0x80000000
	WS_CAPTION      = 0x00C00000
	WS_THICKFRAME   = 0x00040000
	WS_MINIMIZEBOX  = 0x00020000
	WS_MAXIMIZEBOX  = 0x00010000
	WS_SYSMENU      = 0x00080000
	WS_EX_APPWINDOW = 0x00040000

	SWP_NOSIZE       = 0x0001
	SWP_NOMOVE       = 0x0002
	SWP_NOZORDER     = 0x0004
	SWP_FRAMECHANGED = 0x0020
	SWP_SHOWWINDOW   = 0x0040

	SW_HIDE    = 0
	SW_SHOW    = 5
	SW_RESTORE = 9

	NIM_ADD    = 0x00000000
	NIM_MODIFY = 0x0001
	NIM_DELETE = 0x0002

	NIF_MESSAGE = 0x0001
	NIF_ICON    = 0x0002
	NIF_TIP     = 0x0004

	FLASHW_STOP      = 0
	FLASHW_ALL       = 3
	FLASHW_TIMERNOFG = 12

	MF_STRING    = 0x0000
	MF_CHECKED   = 0x0008
	MF_SEPARATOR = 0x0800

	TPM_RETURNCMD = 0x0100
	TPM_NONOTIFY  = 0x0080

	ID_TRAY_CONTROL_CENTER = 2000
	ID_TRAY_SHOW           = 2001
	ID_TRAY_DIRECT_CHAT    = 2002
	ID_TRAY_PRIVACY        = 2003
	ID_TRAY_LOCK           = 2004
	ID_TRAY_CHANGE_PIN     = 2005
	ID_TRAY_AUTOSTART      = 2006
	ID_TRAY_SCRATCHPAD     = 2007
	ID_TRAY_DUAL_ACCOUNT   = 2008
	ID_TRAY_TELEGRAM       = 2010
	ID_TRAY_SPLIT_VIEW     = 2011
	ID_TRAY_EXIT           = 2009

	ICON_SMALL      = 0
	IMAGE_ICON      = 1
	LR_LOADFROMFILE = 0x0010
)

type RECT struct {
	Left   int32
	Top    int32
	Right  int32
	Bottom int32
}

type NOTIFYICONDATAW struct {
	CbSize           uint32
	_                uint32
	HWnd             windows.Handle
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	_                uint32
	HIcon            windows.Handle
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	TimeoutOrVersion uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         windows.GUID
	HBalloonIcon     windows.Handle
}

type FLASHWINFO struct {
	CbSize    uint32
	_         uint32
	HWnd      windows.Handle
	DwFlags   uint32
	UCount    uint32
	DwTimeout uint32
	_         uint32
}

type POINT struct {
	X int32
	Y int32
}

var (
	globalWebView         webview2.WebView
	globalHwnd            uintptr
	globalTelegramWebView webview2.WebView
	globalTelegramHwnd    uintptr
	activeMessengerMode   = "wa" // "wa", "tg", or "split"
	globalTrayData        NOTIFYICONDATAW
	origWndProc           uintptr
	forceQuit             bool
	privacyModeActive     bool
	unreadMessagesCount   int
	currentProfileID      = "1"
	currentWindowTitle    = windowTitle
	currentMutexName      = mutexName
)

func getWindowLongPtr(hwnd uintptr, index int) uintptr {
	if procSetWindowLongPtr.Find() == nil {
		procGetWindowLong := user32.NewProc("GetWindowLongW")
		r, _, _ := procGetWindowLong.Call(hwnd, uintptr(index))
		return r
	}
	r, _, _ := procGetWindowLongPtr.Call(hwnd, uintptr(index))
	return r
}

func findDirectChildByClass(parent uintptr, className string) uintptr {
	pClass, _ := windows.UTF16PtrFromString(className)
	h, _, _ := procFindWindowEx.Call(parent, 0, uintptr(unsafe.Pointer(pClass)), 0)
	return h
}

func setDarkWindowFrame(hwnd uintptr) {
	darkMode := int32(1)
	// Try standard DWMWA_USE_IMMERSIVE_DARK_MODE (Win10 20H1+ & Win11)
	procDwmSetAttr.Call(
		hwnd,
		uintptr(DWMWA_USE_IMMERSIVE_DARK_MODE),
		uintptr(unsafe.Pointer(&darkMode)),
		unsafe.Sizeof(darkMode),
	)
	// Try older Win10 build
	procDwmSetAttr.Call(
		hwnd,
		uintptr(DWMWA_USE_IMMERSIVE_DARK_MODE_BEFORE_20H1),
		uintptr(unsafe.Pointer(&darkMode)),
		unsafe.Sizeof(darkMode),
	)

	// Set dark caption color (COLORREF: 0x00111B21 WhatsApp Dark Header: RGB 17, 27, 33)
	captionColor := uint32(0x00211B11) // 0x00BBGGRR
	procDwmSetAttr.Call(
		hwnd,
		uintptr(DWMWA_CAPTION_COLOR),
		uintptr(unsafe.Pointer(&captionColor)),
		unsafe.Sizeof(captionColor),
	)

	// Set white caption text (RGB 255, 255, 255)
	textColor := uint32(0x00FFFFFF)
	procDwmSetAttr.Call(
		hwnd,
		uintptr(DWMWA_TEXT_COLOR),
		uintptr(unsafe.Pointer(&textColor)),
		unsafe.Sizeof(textColor),
	)
}

func checkSingleInstance(mName, wTitle string) (uintptr, bool) {
	namePtr, _ := syscall.UTF16PtrFromString(mName)
	handle, _, err := procCreateMutex.Call(0, 1, uintptr(unsafe.Pointer(namePtr)))
	if err == windows.ERROR_ALREADY_EXISTS {
		titlePtr, _ := syscall.UTF16PtrFromString(wTitle)
		hwnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titlePtr)))
		if hwnd != 0 {
			procShowWindow.Call(hwnd, SW_SHOW)
			procShowNormal.Call(hwnd, SW_RESTORE)
			procSetFgWindow.Call(hwnd)
		}
		return handle, false
	}
	return handle, true
}

func showMainWindow() {
	if globalHwnd != 0 {
		procShowWindow.Call(globalHwnd, SW_SHOW)
		procShowWindow.Call(globalHwnd, SW_RESTORE)
		procSetFgWindow.Call(globalHwnd)
		applyMessengerLayout()
	}
}

func updateTrayTooltip(text string) {
	if globalHwnd == 0 {
		return
	}
	utf16Text, _ := windows.UTF16FromString(text)
	copy(globalTrayData.SzTip[:], utf16Text)
	globalTrayData.UFlags = NIF_TIP
	procShellNotifyIcon.Call(NIM_MODIFY, uintptr(unsafe.Pointer(&globalTrayData)))
}

func removeTrayIcon() {
	if globalHwnd != 0 {
		procShellNotifyIcon.Call(NIM_DELETE, uintptr(unsafe.Pointer(&globalTrayData)))
	}
}

func isAppAutoStartEnabled() bool {
	var hKey windows.Handle
	subPath, _ := windows.UTF16PtrFromString(runKeyPath)
	err := windows.RegOpenKeyEx(windows.Handle(HKEY_CURRENT_USER), subPath, 0, KEY_READ, &hKey)
	if err != nil {
		return false
	}
	defer windows.RegCloseKey(hKey)

	valName, _ := windows.UTF16PtrFromString(runKeyName)
	var valType uint32
	var bufLen uint32
	err = windows.RegQueryValueEx(hKey, valName, nil, &valType, nil, &bufLen)
	return err == nil && bufLen > 0
}

func setAppAutoStartEnabled(enabled bool) bool {
	exePath, err := os.Executable()
	if err != nil {
		return false
	}
	subPath, _ := windows.UTF16PtrFromString(runKeyPath)
	valName, _ := windows.UTF16PtrFromString(runKeyName)

	if enabled {
		var hKey windows.Handle
		err := windows.RegOpenKeyEx(windows.Handle(HKEY_CURRENT_USER), subPath, 0, KEY_WRITE, &hKey)
		if err != nil {
			return false
		}
		defer windows.RegCloseKey(hKey)

		cmdValue := fmt.Sprintf("\"%s\" --minimized", exePath)
		utf16Val, err := syscall.UTF16FromString(cmdValue)
		if err != nil {
			return false
		}
		byteSize := uint32(len(utf16Val) * 2)

		r, _, _ := procRegSetValueEx.Call(
			uintptr(hKey),
			uintptr(unsafe.Pointer(valName)),
			0,
			REG_SZ,
			uintptr(unsafe.Pointer(&utf16Val[0])),
			uintptr(byteSize),
		)
		return r == 0
	} else {
		var hKey windows.Handle
		err := windows.RegOpenKeyEx(windows.Handle(HKEY_CURRENT_USER), subPath, 0, KEY_WRITE, &hKey)
		if err != nil {
			return false
		}
		defer windows.RegCloseKey(hKey)

		r, _, _ := procRegDeleteVal.Call(uintptr(hKey), uintptr(unsafe.Pointer(valName)))
		return r == 0 || r == 2
	}
}

func showTrayContextMenu() {
	hMenu, _, _ := procCreatePopupMenu.Call()
	if hMenu == 0 {
		return
	}
	procAppendMenu.Call(hMenu, MF_STRING, ID_TRAY_CONTROL_CENTER, uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("⚡ Menu Add-on & Pengaturan..."))))
	procAppendMenu.Call(hMenu, MF_SEPARATOR, 0, 0)
	procAppendMenu.Call(hMenu, MF_STRING, ID_TRAY_SHOW, uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("Buka WhatsApp"))))
	procAppendMenu.Call(hMenu, MF_STRING, ID_TRAY_DIRECT_CHAT, uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("Chat ke Nomor Baru (Ctrl+N)"))))

	privacyFlags := uintptr(MF_STRING)
	if privacyModeActive {
		privacyFlags |= MF_CHECKED
	}
	procAppendMenu.Call(hMenu, privacyFlags, ID_TRAY_PRIVACY, uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("Mode Privasi (Blur) (Alt+P)"))))
	procAppendMenu.Call(hMenu, MF_STRING, ID_TRAY_LOCK, uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("Kunci WhatsApp (Ctrl+L)"))))
	procAppendMenu.Call(hMenu, MF_STRING, ID_TRAY_CHANGE_PIN, uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("Ubah PIN Kunci..."))))

	autostartFlags := uintptr(MF_STRING)
	if isAppAutoStartEnabled() {
		autostartFlags |= MF_CHECKED
	}
	procAppendMenu.Call(hMenu, autostartFlags, ID_TRAY_AUTOSTART, uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("Mulai Otomatis saat Boot"))))
	procAppendMenu.Call(hMenu, MF_STRING, ID_TRAY_SCRATCHPAD, uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("Catatan Cepat (Scratchpad) (Alt+N)"))))
	procAppendMenu.Call(hMenu, MF_STRING, ID_TRAY_DUAL_ACCOUNT, uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("Buka Akun WhatsApp Ke-2..."))))
	procAppendMenu.Call(hMenu, MF_STRING, ID_TRAY_TELEGRAM, uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("🔵 Buka Telegram Web (Ctrl+2)"))))
	procAppendMenu.Call(hMenu, MF_STRING, ID_TRAY_SPLIT_VIEW, uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("◫ Buka Keduanya Berdampingan (Ctrl+3)"))))
	procAppendMenu.Call(hMenu, MF_SEPARATOR, 0, 0)
	procAppendMenu.Call(hMenu, MF_STRING, ID_TRAY_EXIT, uintptr(unsafe.Pointer(windows.StringToUTF16Ptr("Keluar"))))

	var pt POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	procSetFgWindow.Call(globalHwnd)
	cmd, _, _ := procTrackPopupMenu.Call(hMenu, TPM_RETURNCMD|TPM_NONOTIFY, uintptr(pt.X), uintptr(pt.Y), 0, globalHwnd, 0)
	procDestroyMenu.Call(hMenu)

	switch cmd {
	case ID_TRAY_CONTROL_CENTER:
		showMainWindow()
		if globalWebView != nil {
			globalWebView.Eval("window.openAddonControlCenter && window.openAddonControlCenter()")
		}
	case ID_TRAY_SHOW:
		showMainWindow()
		switchToWhatsApp()
	case ID_TRAY_DIRECT_CHAT:
		showMainWindow()
		if globalWebView != nil {
			globalWebView.Eval("window.openDirectChatModal && window.openDirectChatModal()")
		}
	case ID_TRAY_PRIVACY:
		if globalWebView != nil {
			globalWebView.Eval("window.togglePrivacyMode && window.togglePrivacyMode()")
		}
	case ID_TRAY_LOCK:
		showMainWindow()
		if globalWebView != nil {
			globalWebView.Eval("window.lockWhatsApp && window.lockWhatsApp()")
		}
		if globalTelegramWebView != nil {
			globalTelegramWebView.Eval("window.lockTelegram && window.lockTelegram()")
		}
	case ID_TRAY_CHANGE_PIN:
		showMainWindow()
		if activeMessengerMode == "tg" && globalTelegramWebView != nil {
			globalTelegramWebView.Eval("window.openChangePinModal && window.openChangePinModal()")
		} else if globalWebView != nil {
			globalWebView.Eval("window.openChangePinModal && window.openChangePinModal()")
		}
	case ID_TRAY_AUTOSTART:
		current := isAppAutoStartEnabled()
		setAppAutoStartEnabled(!current)
		if globalWebView != nil {
			globalWebView.Eval("window.syncAutoStartUI && window.syncAutoStartUI()")
		}
	case ID_TRAY_SCRATCHPAD:
		showMainWindow()
		if globalWebView != nil {
			globalWebView.Eval("window.toggleScratchpad && window.toggleScratchpad()")
		}
	case ID_TRAY_DUAL_ACCOUNT:
		launchDualAccount()
	case ID_TRAY_TELEGRAM:
		showMainWindow()
		switchToTelegram()
	case ID_TRAY_SPLIT_VIEW:
		showMainWindow()
		switchToSplitView()
	case ID_TRAY_EXIT:
		forceQuit = true
		removeTrayIcon()
		procPostMessage.Call(globalHwnd, WM_CLOSE, 0, 0)
	}
}

func setWindowLongPtr(hwnd uintptr, index int, newLong uintptr) uintptr {
	if procSetWindowLongPtr.Find() == nil {
		r, _, _ := procSetWindowLongPtr.Call(hwnd, uintptr(index), newLong)
		return r
	}
	procSetWindowLong := user32.NewProc("SetWindowLongW")
	r, _, _ := procSetWindowLong.Call(hwnd, uintptr(index), newLong)
	return r
}

func customWndProc(hWnd, msg, wParam, lParam uintptr) uintptr {
	switch uint32(msg) {
	case WM_CLOSE:
		if !forceQuit {
			// Minimize to tray instead of quitting
			procShowWindow.Call(hWnd, SW_HIDE)
			return 0
		}
	case WM_SIZE:
		r, _, _ := procCallWindowProc.Call(origWndProc, hWnd, msg, wParam, lParam)
		applyMessengerLayout()
		return r
	case WM_HOTKEY:
		if wParam == HOTKEY_BOSS {
			fg, _, _ := procGetFgWindow.Call()
			visible, _, _ := procIsWindowVisible.Call(hWnd)
			if visible != 0 && fg == hWnd {
				procShowWindow.Call(hWnd, SW_HIDE)
			} else {
				showMainWindow()
				procSetFgWindow.Call(hWnd)
			}
			return 0
		}
	case WM_COMMAND:
		switch uint32(wParam) {
		case ID_TRAY_SHOW:
			showMainWindow()
			switchToWhatsApp()
			return 0
		case ID_TRAY_TELEGRAM:
			showMainWindow()
			switchToTelegram()
			return 0
		case ID_TRAY_SPLIT_VIEW:
			showMainWindow()
			switchToSplitView()
			return 0
		}
	case WM_TRAYICON:
		switch uint32(lParam) {
		case WM_LBUTTONUP, WM_LBUTTONDBLCLK:
			showMainWindow()
			return 0
		case WM_RBUTTONUP:
			showTrayContextMenu()
			return 0
		}
	}
	r, _, _ := procCallWindowProc.Call(origWndProc, hWnd, msg, wParam, lParam)
	return r
}

func updateUnreadCount(count int) {
	unreadMessagesCount = count
	if count > 0 {
		tip := fmt.Sprintf("%s (%d pesan baru)", currentWindowTitle, count)
		updateTrayTooltip(tip)

		fg, _, _ := procGetFgWindow.Call()
		if fg != globalHwnd {
			fInfo := FLASHWINFO{
				CbSize:    uint32(unsafe.Sizeof(FLASHWINFO{})),
				HWnd:      windows.Handle(globalHwnd),
				DwFlags:   FLASHW_ALL | FLASHW_TIMERNOFG,
				UCount:    0,
				DwTimeout: 0,
			}
			procFlashWindowEx.Call(uintptr(unsafe.Pointer(&fInfo)))
		}
	} else {
		updateTrayTooltip(currentWindowTitle)
		fInfo := FLASHWINFO{
			CbSize:    uint32(unsafe.Sizeof(FLASHWINFO{})),
			HWnd:      windows.Handle(globalHwnd),
			DwFlags:   FLASHW_STOP,
			UCount:    0,
			DwTimeout: 0,
		}
		procFlashWindowEx.Call(uintptr(unsafe.Pointer(&fInfo)))
	}
}


var (
	cachedAppPin = "1234"
)

func getPinFilePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.Getenv("APPDATA")
		if configDir == "" {
			configDir = "."
		}
	}
	base := filepath.Join(configDir, "WhatsAppDesktopLight")
	_ = os.MkdirAll(base, 0755)
	return filepath.Join(base, "app_pin.dat")
}

func loadSharedPin() string {
	p := getPinFilePath()
	if data, err := os.ReadFile(p); err == nil {
		val := strings.TrimSpace(string(data))
		if len(val) >= 4 && len(val) <= 6 {
			cachedAppPin = val
		}
	}
	return cachedAppPin
}

func saveSharedPin(pin string) {
	pin = strings.TrimSpace(pin)
	if len(pin) >= 4 && len(pin) <= 6 {
		cachedAppPin = pin
		p := getPinFilePath()
		_ = os.WriteFile(p, []byte(pin), 0600)
	}
}

func getUserDataDir(profile string) string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.Getenv("APPDATA")
		if configDir == "" {
			configDir = "."
		}
	}
	var dirName string
	if profile == "1" || profile == "" {
		dirName = "UserData"
	} else {
		dirName = fmt.Sprintf("UserData_Profile%s", profile)
	}
	dir := filepath.Join(configDir, "WhatsAppDesktopLight", dirName)
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func launchDualAccount() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}
	exeDir := filepath.Dir(exePath)
	verbPtr, _ := syscall.UTF16PtrFromString("open")
	filePtr, _ := syscall.UTF16PtrFromString(exePath)
	paramPtr, _ := syscall.UTF16PtrFromString("--profile 2")
	dirPtr, _ := syscall.UTF16PtrFromString(exeDir)
	procShellExecute.Call(0, uintptr(unsafe.Pointer(verbPtr)), uintptr(unsafe.Pointer(filePtr)), uintptr(unsafe.Pointer(paramPtr)), uintptr(unsafe.Pointer(dirPtr)), 1)
}

func initTelegramChild() {
	if globalTelegramWebView != nil {
		return
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.Getenv("APPDATA")
		if configDir == "" {
			configDir = "."
		}
	}
	userDataDir := filepath.Join(configDir, "WhatsAppDesktopLight", "UserData_Telegram")
	_ = os.MkdirAll(userDataDir, 0755)

	opts := webview2.WebViewOptions{
		Window:    nil,
		Debug:     false,
		DataPath:  userDataDir,
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title:  telegramWindowTitle,
			Width:  1050,
			Height: 720,
		},
	}

	w := webview2.NewWithOptions(opts)
	if w == nil {
		log.Println("Gagal inisialisasi WebView2 untuk Telegram")
		return
	}
	globalTelegramWebView = w
	hTG := uintptr(w.Window())
	globalTelegramHwnd = hTG

	// Hide initially and reparent into main window
	procShowWindow.Call(hTG, SW_HIDE)
	procSetParent.Call(hTG, globalHwnd)

	// Convert to WS_CHILD without borders/caption
	s := getWindowLongPtr(hTG, GWL_STYLE)
	newStyle := (s &^ uintptr(WS_POPUP|WS_CAPTION|WS_THICKFRAME|WS_MINIMIZEBOX|WS_MAXIMIZEBOX|WS_SYSMENU)) | uintptr(WS_CHILD)
	setWindowLongPtr(hTG, GWL_STYLE, newStyle)

	ex := getWindowLongPtr(hTG, GWL_EXSTYLE)
	newEx := ex &^ uintptr(WS_EX_APPWINDOW)
	setWindowLongPtr(hTG, GWL_EXSTYLE, newEx)

	procSetWindowPos.Call(hTG, 0, 0, 0, 0, 0, SWP_NOMOVE|SWP_NOSIZE|SWP_NOZORDER|SWP_FRAMECHANGED)

	executablePath, _ := os.Executable()
	iconFullPath := filepath.Join(filepath.Dir(executablePath), "icon.ico")

	_ = w.Bind("sendNativeNotification", func(title, body string) {
		go showNativeNotification(title, body, iconFullPath)
	})
	_ = w.Bind("openExternalLink", func(rawURL string) {
		openExternalLink(rawURL)
	})
	_ = w.Bind("openWhatsAppWindow", func() {
		switchToWhatsApp()
	})
	_ = w.Bind("switchToWhatsApp", func() {
		switchToWhatsApp()
	})
	_ = w.Bind("openTelegramWindow", func() {
		switchToTelegram()
	})
	_ = w.Bind("switchToTelegram", func() {
		switchToTelegram()
	})
	_ = w.Bind("openSideBySideView", func() {
		switchToSplitView()
	})
	_ = w.Bind("switchToSplitView", func() {
		switchToSplitView()
	})
	_ = w.Bind("getAppPin", func() string {
		return loadSharedPin()
	})
	_ = w.Bind("syncAppPin", func(pin string) {
		saveSharedPin(pin)
		if globalWebView != nil {
			globalWebView.Eval(fmt.Sprintf("if(window.updateStoredPin) window.updateStoredPin(%q);", pin))
		}
		if globalTelegramWebView != nil {
			globalTelegramWebView.Eval(fmt.Sprintf("if(window.updateStoredPin) window.updateStoredPin(%q);", pin))
		}
	})

	w.Init(telegramInitScript)
	w.Navigate(telegramAppURL)
}

func applyMessengerLayout() {
	if globalHwnd == 0 {
		return
	}
	var rc RECT
	procGetClientRect.Call(globalHwnd, uintptr(unsafe.Pointer(&rc)))
	totalW := rc.Right - rc.Left
	totalH := rc.Bottom - rc.Top
	if totalW <= 0 || totalH <= 0 {
		return
	}

	hWAWidget := findDirectChildByClass(globalHwnd, "Chrome_WidgetWin_0")

	switch activeMessengerMode {
	case "tg":
		if globalTelegramHwnd != 0 {
			procMoveWindow.Call(globalTelegramHwnd, 0, 0, uintptr(totalW), uintptr(totalH), 1)
			procShowWindow.Call(globalTelegramHwnd, SW_SHOW)
			procSetFocus.Call(globalTelegramHwnd)
		}
	case "split":
		halfW := totalW / 2
		if hWAWidget != 0 {
			procMoveWindow.Call(hWAWidget, 0, 0, uintptr(halfW), uintptr(totalH), 1)
		}
		if globalTelegramHwnd != 0 {
			procMoveWindow.Call(globalTelegramHwnd, uintptr(halfW), 0, uintptr(totalW-halfW), uintptr(totalH), 1)
			procShowWindow.Call(globalTelegramHwnd, SW_SHOW)
		}
	default: // "wa"
		if globalTelegramHwnd != 0 {
			procShowWindow.Call(globalTelegramHwnd, SW_HIDE)
		}
		if hWAWidget != 0 {
			procMoveWindow.Call(hWAWidget, 0, 0, uintptr(totalW), uintptr(totalH), 1)
		}
		procSetFocus.Call(globalHwnd)
	}
}

func switchToWhatsApp() {
	activeMessengerMode = "wa"
	applyMessengerLayout()
	if globalWebView != nil {
		globalWebView.SetTitle(currentWindowTitle)
		globalWebView.Eval("window.ensureDock && window.ensureDock()")
		globalWebView.Eval("window.syncDockActiveTab && window.syncDockActiveTab('wa')")
	}
	if globalTelegramWebView != nil {
		globalTelegramWebView.Eval("window.syncDockActiveTab && window.syncDockActiveTab('wa')")
	}
}

func switchToTelegram() {
	if globalTelegramWebView == nil {
		initTelegramChild()
	}
	activeMessengerMode = "tg"
	applyMessengerLayout()
	if globalWebView != nil {
		globalWebView.SetTitle(telegramWindowTitle)
		globalWebView.Eval("window.syncDockActiveTab && window.syncDockActiveTab('tg')")
	}
	if globalTelegramWebView != nil {
		globalTelegramWebView.Eval("window.ensureDock && window.ensureDock()")
		globalTelegramWebView.Eval("window.syncDockActiveTab && window.syncDockActiveTab('tg')")
	}
}

func switchToSplitView() {
	if globalTelegramWebView == nil {
		initTelegramChild()
	}
	activeMessengerMode = "split"
	applyMessengerLayout()
	if globalWebView != nil {
		globalWebView.SetTitle("WhatsApp & Telegram")
		globalWebView.Eval("window.syncDockActiveTab && window.syncDockActiveTab('split')")
	}
	if globalTelegramWebView != nil {
		globalTelegramWebView.Eval("window.ensureDock && window.ensureDock()")
		globalTelegramWebView.Eval("window.syncDockActiveTab && window.syncDockActiveTab('split')")
	}
}

func launchTelegram() {
	showMainWindow()
	switchToTelegram()
}

func focusWhatsApp() {
	showMainWindow()
	switchToWhatsApp()
}

func arrangeSideBySide() {
	showMainWindow()
	switchToSplitView()
}

func setTelegramDarkWindowFrame(hwnd uintptr) {
	darkMode := int32(1)
	procDwmSetAttr.Call(
		hwnd,
		uintptr(DWMWA_USE_IMMERSIVE_DARK_MODE),
		uintptr(unsafe.Pointer(&darkMode)),
		unsafe.Sizeof(darkMode),
	)
	captionColor := uint32(0x002B2117) // Telegram dark: #17212b -> BGR 0x002B2117
	procDwmSetAttr.Call(
		hwnd,
		uintptr(DWMWA_CAPTION_COLOR),
		uintptr(unsafe.Pointer(&captionColor)),
		unsafe.Sizeof(captionColor),
	)
	textColor := uint32(0x00FFFFFF)
	procDwmSetAttr.Call(
		hwnd,
		uintptr(DWMWA_TEXT_COLOR),
		uintptr(unsafe.Pointer(&textColor)),
		unsafe.Sizeof(textColor),
	)
}

func telegramWndProc(hWnd, msg, wParam, lParam uintptr) uintptr {
	switch uint32(msg) {
	case WM_CLOSE:
		if !forceQuit {
			procShowWindow.Call(hWnd, SW_HIDE)
			return 0
		}
	case WM_HOTKEY:
		if wParam == HOTKEY_BOSS {
			wTitleWA, _ := windows.UTF16PtrFromString(windowTitle)
			hWA, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(wTitleWA)))

			fg, _, _ := procGetFgWindow.Call()
			visible, _, _ := procIsWindowVisible.Call(hWnd)
			if visible != 0 && (fg == hWnd || fg == hWA) {
				procShowWindow.Call(hWnd, SW_HIDE)
				if hWA != 0 {
					procShowWindow.Call(hWA, SW_HIDE)
				}
			} else {
				procShowWindow.Call(hWnd, SW_RESTORE)
				procSetFgWindow.Call(hWnd)
				if hWA != 0 {
					procShowWindow.Call(hWA, SW_RESTORE)
				}
			}
			return 0
		}
	case WM_TRAYICON:
		switch uint32(lParam) {
		case WM_LBUTTONUP, WM_LBUTTONDBLCLK:
			procShowWindow.Call(hWnd, SW_RESTORE)
			procSetFgWindow.Call(hWnd)
			return 0
		case WM_RBUTTONUP:
			showTrayContextMenu()
			return 0
		}
	}
	r, _, _ := procCallWindowProc.Call(origWndProc, hWnd, msg, wParam, lParam)
	return r
}

func runTelegramWindow(userDataDir string, startMinimized bool) {
	executablePath, _ := os.Executable()
	iconFullPath := filepath.Join(filepath.Dir(executablePath), "icon.ico")

	opts := webview2.WebViewOptions{
		Window:    nil,
		Debug:     false,
		DataPath:  userDataDir,
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title:  telegramWindowTitle,
			Width:  1050,
			Height: 720,
			IconId: 2,
			Center: true,
		},
	}

	w := webview2.NewWithOptions(opts)
	if w == nil {
		log.Fatalln("Gagal inisialisasi WebView2 untuk Telegram")
	}
	defer w.Destroy()

	hwnd := uintptr(w.Window())
	globalHwnd = hwnd
	globalWebView = w

	setTelegramDarkWindowFrame(hwnd)
	w.SetTitle(telegramWindowTitle)

	if startMinimized {
		procShowWindow.Call(hwnd, SW_HIDE)
	} else {
		w.SetSize(1050, 720, webview2.HintNone)
	}

	procRegisterHotKey.Call(hwnd, HOTKEY_BOSS, MOD_CONTROL|MOD_ALT|MOD_NOREPEAT, uintptr('W'))
	defer procUnregisterHotKey.Call(hwnd, HOTKEY_BOSS)

	origWndProc = setWindowLongPtr(hwnd, GWLP_WNDPROC, windows.NewCallback(telegramWndProc))

	_ = w.Bind("sendNativeNotification", func(title, body string) {
		go showNativeNotification(title, body, iconFullPath)
	})
	_ = w.Bind("openExternalLink", func(rawURL string) {
		openExternalLink(rawURL)
	})
	_ = w.Bind("openWhatsAppWindow", func() {
		focusWhatsApp()
	})
	_ = w.Bind("openTelegramWindow", func() {
		launchTelegram()
	})
	_ = w.Bind("openSideBySideView", func() {
		arrangeSideBySide()
	})
	_ = w.Bind("getAppPin", func() string {
		return loadSharedPin()
	})
	_ = w.Bind("syncAppPin", func(pin string) {
		saveSharedPin(pin)
	})

	w.Init(telegramInitScript)
	w.Navigate(telegramAppURL)
	w.Run()
}

const telegramInitScript = `
	// UserAgent override
	Object.defineProperty(navigator, 'userAgent', {
		get: () => '` + userAgent + `'
	});
	Object.defineProperty(navigator, 'appVersion', {
		get: () => '` + userAgent + `'
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
		};
		window.Notification.permission = 'granted';
		window.Notification.requestPermission = function(callback) {
			var p = Promise.resolve('granted');
			if (typeof callback === 'function') callback('granted');
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
					if (!u.hostname.includes('telegram.org') && (u.protocol === 'http:' || u.protocol === 'https:')) {
						if (window.openExternalLink) window.openExternalLink(u.href);
						return null;
					}
				} catch(e) {}
			}
			return origOpen.apply(this, arguments);
		};
	})();

	// In-App Tab Switcher (Floating Glassmorphic Dock for Telegram)
	(function() {
		function injectTabs() {
			var root = document.documentElement || document.body;
			if (!root) return;

			if (!document.getElementById('tg-dock-style')) {
				var style = document.createElement('style');
				style.id = 'tg-dock-style';
				style.textContent = [
					'.wa-dock { position:fixed !important;bottom:16px !important;left:16px !important;z-index:2147483647 !important;display:flex !important;align-items:center !important;gap:4px !important;background:rgba(23,33,43,0.96) !important;backdrop-filter:blur(24px) !important;-webkit-backdrop-filter:blur(24px) !important;border:1px solid rgba(255,255,255,0.16) !important;padding:4px 8px !important;border-radius:24px !important;box-shadow:0 8px 32px rgba(0,0,0,0.7) !important;user-select:none !important;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,Helvetica,sans-serif !important;pointer-events:auto !important;visibility:visible !important;opacity:1 !important; }',
					'.wa-dock-item { display:flex !important;align-items:center !important;gap:6px !important;padding:6px 13px !important;border-radius:18px !important;font-size:12px !important;font-weight:500 !important;color:#9db2c6 !important;background:transparent !important;border:none !important;cursor:pointer !important;transition:all 0.15s cubic-bezier(0.16,1,0.3,1) !important;outline:none !important;white-space:nowrap !important; }',
					'.wa-dock-item:hover { color:#ffffff !important;background:rgba(255,255,255,0.08) !important; }',
					'.wa-dock-item.active { background:rgba(36,161,222,0.22) !important;color:#24A1DE !important;font-weight:600 !important;cursor:default !important; }',
					'.wa-dock-sep { width:1px !important;height:14px !important;background:rgba(255,255,255,0.12) !important;margin:0 2px !important; }',
					'.wa-dot-wa { width:7px !important;height:7px !important;border-radius:50% !important;background:#00a884 !important;display:inline-block !important;box-shadow:0 0 6px rgba(0,168,132,0.6) !important; }',
					'.wa-dot-tg { width:7px !important;height:7px !important;border-radius:50% !important;background:#24A1DE !important;display:inline-block !important; }',
					'.tg-qr-popup { position:fixed !important; z-index:2147483647 !important; background:rgba(23,33,43,0.98) !important; backdrop-filter:blur(20px) !important; -webkit-backdrop-filter:blur(20px) !important; border:1px solid rgba(255,255,255,0.16) !important; border-radius:10px !important; box-shadow:0 16px 40px rgba(0,0,0,0.75) !important; width:380px !important; max-width:90vw !important; max-height:260px !important; display:none; flex-direction:column !important; overflow:hidden !important; font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,Helvetica,sans-serif !important; user-select:none !important; pointer-events:auto !important; }',
					'.tg-qr-header { padding:8px 12px !important; background:rgba(15,22,29,0.95) !important; border-bottom:1px solid rgba(255,255,255,0.08) !important; display:flex !important; align-items:center !important; justify-content:space-between !important; font-size:11px !important; color:#9db2c6 !important; }',
					'.tg-qr-badge { font-size:10px !important; padding:2px 6px !important; border-radius:4px !important; background:rgba(36,161,222,0.18) !important; color:#24A1DE !important; font-weight:600 !important; }',
					'.tg-qr-list { overflow-y:auto !important; padding:4px !important; display:flex !important; flex-direction:column !important; gap:2px !important; max-height:210px !important; }',
					'.tg-qr-item { display:flex !important; flex-direction:column !important; padding:8px 10px !important; border-radius:6px !important; cursor:pointer !important; transition:all 0.12s ease !important; border:1px solid transparent !important; }',
					'.tg-qr-item:hover, .tg-qr-item.active { background:rgba(255,255,255,0.08) !important; border-color:rgba(255,255,255,0.12) !important; }',
					'.tg-qr-item.active { background:rgba(36,161,222,0.22) !important; border-color:rgba(36,161,222,0.4) !important; }',
					'.tg-qr-item-key { font-size:12.5px !important; font-weight:600 !important; color:#24A1DE !important; display:flex !important; align-items:center !important; gap:6px !important; }',
					'.tg-qr-item-text { font-size:11.5px !important; color:#9db2c6 !important; white-space:nowrap !important; overflow:hidden !important; text-overflow:ellipsis !important; margin-top:2px !important; }'
				].join('\n');
				(document.head || root).appendChild(style);
			}

			var bar = document.getElementById('tg-messenger-tabs');
			if (!bar) {
				bar = document.createElement('div');
				bar.id = 'tg-messenger-tabs';
				bar.className = 'wa-dock';
				bar.innerHTML = [
					'<button id="tg-tab-wa" class="wa-dock-item" title="Kembali ke WhatsApp (Ctrl+1)">' +
						'<span class="wa-dot-wa"></span><span>WhatsApp</span>' +
					'</button>',
					'<div class="wa-dock-sep"></div>',
					'<button id="tg-tab-tg" class="wa-dock-item active" title="Telegram Aktif (Ctrl+2)">' +
						'<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/></svg>' +
						'<span>Telegram</span>' +
					'</button>',
					'<div class="wa-dock-sep"></div>',
					'<button id="tg-tab-split" class="wa-dock-item" title="Mode Berdampingan 50:50 (Ctrl+3)">' +
						'<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="12" y1="3" x2="12" y2="21"/></svg>' +
						'<span>Berdampingan</span>' +
					'</button>',
					'<div class="wa-dock-sep"></div>',
					'<button id="tg-tab-lock" class="wa-dock-item" title="Kunci Telegram (Ctrl+L)">' +
						'<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>' +
						'<span>Kunci</span>' +
					'</button>'
				].join('');

				var btnWA = bar.querySelector('#tg-tab-wa');
				var btnTG = bar.querySelector('#tg-tab-tg');
				var btnSplit = bar.querySelector('#tg-tab-split');
				var btnLock = bar.querySelector('#tg-tab-lock');

				if (btnLock) {
					btnLock.onclick = function(e) {
						e.preventDefault();
						e.stopPropagation();
						if (window.lockTelegram) window.lockTelegram();
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

				root.appendChild(bar);
			} else if (!bar.isConnected || bar.parentElement !== root) {
				root.appendChild(bar);
			}
		}

		window.ensureDock = injectTabs;

		window.syncDockActiveTab = function(mode) {
			var bWA = document.getElementById('tg-tab-wa');
			var bTG = document.getElementById('tg-tab-tg');
			var bSp = document.getElementById('tg-tab-split');
			if (bWA) bWA.classList.toggle('active', mode === 'wa');
			if (bTG) bTG.classList.toggle('active', mode === 'tg');
			if (bSp) bSp.classList.toggle('active', mode === 'split');
		};

		// Shortcuts: Ctrl+1 (WhatsApp), Ctrl+2 (Telegram), Ctrl+3 (Split View)
		var handleShortcuts = function(e) {
			if (e.ctrlKey || e.metaKey) {
				if (e.key === '1' || e.code === 'Digit1') {
					if (e.preventDefault) e.preventDefault();
					if (e.stopPropagation) e.stopPropagation();
					if (window.switchToWhatsApp) window.switchToWhatsApp();
					else if (window.openWhatsAppWindow) window.openWhatsAppWindow();
				} else if (e.key === '2' || e.code === 'Digit2') {
					if (e.preventDefault) e.preventDefault();
					if (e.stopPropagation) e.stopPropagation();
					if (window.switchToTelegram) window.switchToTelegram();
					else if (window.openTelegramWindow) window.openTelegramWindow();
				} else if (e.key === '3' || e.code === 'Digit3') {
					if (e.preventDefault) e.preventDefault();
					if (e.stopPropagation) e.stopPropagation();
					if (window.switchToSplitView) window.switchToSplitView();
					else if (window.openSideBySideView) window.openSideBySideView();
				}
			}
		};
		window.addEventListener('keydown', handleShortcuts, true);
		document.addEventListener('keydown', handleShortcuts, true);

		// Initial injection attempts
		if (document.body || document.documentElement) {
			injectTabs();
		}
		document.addEventListener('DOMContentLoaded', injectTabs, { once: true });
		setTimeout(injectTabs, 500);
		setTimeout(injectTabs, 1500);
		setTimeout(injectTabs, 3000);

		// Continuous health-check interval to survive dynamic SPA DOM re-renders
		setInterval(injectTabs, 1000);

		// MutationObserver to immediately re-inject if Telegram replaces root DOM elements
		try {
			var observer = new MutationObserver(function() {
				if (!document.getElementById('tg-messenger-tabs')) {
					injectTabs();
				}
			});
			observer.observe(document.documentElement, { childList: true, subtree: true });
		} catch(e) {}

		function showTgToast(msg) {
			var root = document.documentElement || document.body;
			if (!root) return;
			var toast = document.getElementById('tg-toast-msg');
			if (!toast) {
				toast = document.createElement('div');
				toast.id = 'tg-toast-msg';
				toast.style.cssText = 'position:fixed;bottom:70px;left:50%;transform:translateX(-50%);background:rgba(23,33,43,0.96);color:#ffffff;padding:8px 18px;border-radius:20px;font-size:12.5px;font-weight:500;box-shadow:0 8px 24px rgba(0,0,0,0.6);border:1px solid rgba(255,255,255,0.12);z-index:2147483647;pointer-events:none;transition:opacity 0.2s ease,transform 0.2s ease;opacity:0;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;';
				root.appendChild(toast);
			}
			toast.textContent = msg;
			toast.style.opacity = '1';
			toast.style.transform = 'translateX(-50%) translateY(0)';
			clearTimeout(toast._timer);
			toast._timer = setTimeout(function() {
				toast.style.opacity = '0';
				toast.style.transform = 'translateX(-50%) translateY(6px)';
			}, 2500);
		}

		// Feature: Quick Replies in Telegram
		(function() {
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

			var tgQRTypedWord = '';

			function getTgEditorUserText(editable) {
				if (!editable) return '';
				return cleanZeroWidth(editable.innerText || editable.textContent || '').trim();
			}

			function placeTgCaretAtEnd(editable) {
				editable.focus();
				var sel = window.getSelection();
				var range = document.createRange();
				var tn = (editable.lastChild && editable.lastChild.nodeType === Node.TEXT_NODE) ? editable.lastChild :
				         (editable.firstChild && editable.firstChild.nodeType === Node.TEXT_NODE) ? editable.firstChild : null;
				if (tn) {
					range.setStart(tn, tn.textContent.length);
					range.collapse(true);
				} else {
					range.selectNodeContents(editable);
					range.collapse(false);
				}
				sel.removeAllRanges();
				sel.addRange(range);
			}

			function applyTgQuickReplyToEditable(editable, matchKey, textToInsert, typedWord) {
				if (!editable) return false;
				var userText = getTgEditorUserText(editable);
				placeTgCaretAtEnd(editable);

				var m = userText.match(/(?:^|\s)(\/[\w-]*)$/);
				var slashCmd = m ? m[1] : (typedWord || matchKey);

				var delCount = slashCmd.length;
				for (var k = 0; k < delCount; k++) {
					document.execCommand('delete');
				}

				document.execCommand('insertText', false, textToInsert);
				editable.dispatchEvent(new InputEvent('input', { bubbles: true, cancelable: true, inputType: 'insertText', data: textToInsert }));
				if (typeof showTgToast === 'function') {
					showTgToast('⚡ Template: ' + matchKey + ' diterapkan');
				}
				return true;
			}

			function applyTgQuickReply(item, target) {
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
					var slashCmd = m ? m[1] : (tgQRTypedWord || item.key);
					var kIdx = before.lastIndexOf(slashCmd);
					if (kIdx !== -1) {
						target.setRangeText(textToInsert, kIdx, selEnd, 'end');
					} else {
						target.setRangeText(textToInsert, selEnd - slashCmd.length >= 0 ? selEnd - slashCmd.length : 0, selEnd, 'end');
					}
					target.dispatchEvent(new Event('input', { bubbles: true }));
					if (typeof showTgToast === 'function') {
						showTgToast('⚡ Template: ' + item.key + ' diterapkan');
					}
					return true;
				}

				var editable = (target.closest && target.closest('[contenteditable="true"]')) ||
				               (target.isContentEditable ? target : null) ||
				               document.querySelector('.input-message-input[contenteditable="true"]') ||
				               document.querySelector('div[contenteditable="true"]');
				if (!editable) return false;

				return applyTgQuickReplyToEditable(editable, item.key, textToInsert, tgQRTypedWord);
			}

			var tgQRPopup = null;
			var tgQRMatches = [];
			var tgQRSelectedIndex = 0;
			var tgQRTarget = null;

			function ensureTgQRPopup() {
				if (tgQRPopup && tgQRPopup.parentNode) return tgQRPopup;
				tgQRPopup = document.createElement('div');
				tgQRPopup.id = 'tg-qr-popup';
				tgQRPopup.className = 'tg-qr-popup';
				tgQRPopup.style.display = 'none';
				(document.body || document.documentElement).appendChild(tgQRPopup);
				return tgQRPopup;
			}

			function hideTgQRPopup() {
				if (tgQRPopup) {
					tgQRPopup.style.display = 'none';
					tgQRMatches = [];
					tgQRSelectedIndex = 0;
					tgQRTarget = null;
					tgQRTypedWord = '';
				}
			}

			function showTgQRPopup(target, matches, queryWord) {
				if (!matches || matches.length === 0) {
					hideTgQRPopup();
					return;
				}
				tgQRTarget = target;
				tgQRMatches = matches;
				tgQRSelectedIndex = 0;
				tgQRTypedWord = queryWord || '';

				var popup = ensureTgQRPopup();
				var rect = target.getBoundingClientRect();
				var bottomPos = (window.innerHeight - rect.top + 10);
				if (bottomPos < 50) bottomPos = 90;

				var leftPos = rect.left;
				if (leftPos + 390 > window.innerWidth) {
					leftPos = window.innerWidth - 400;
				}
				if (leftPos < 16) leftPos = 16;

				popup.style.bottom = bottomPos + 'px';
				popup.style.left = leftPos + 'px';

				renderTgQRPopup();
				popup.style.display = 'flex';
			}

			function renderTgQRPopup() {
				if (!tgQRPopup) return;
				var html = '<div class="tg-qr-header">' +
					'<div style="display:flex;align-items:center;gap:6px;"><span style="width:6px;height:6px;border-radius:50%;background:#24A1DE;display:inline-block;"></span><b style="color:#ffffff;">Balas Cepat Telegram</b></div>' +
					'<span class="tg-qr-badge">Tekan [Tab] / [Enter] / [Klik]</span>' +
					'</div>' +
					'<div class="tg-qr-list">';

				for (var i = 0; i < tgQRMatches.length; i++) {
					var it = tgQRMatches[i];
					var isAct = (i === tgQRSelectedIndex);
					var cleanSnippet = it.text.replace(/\n/g, ' ').substring(0, 56);
					if (it.text.length > 56) cleanSnippet += '...';
					html += '<div class="tg-qr-item ' + (isAct ? 'active' : '') + '" data-idx="' + i + '">' +
						'<div class="tg-qr-item-key"><span>' + it.key + '</span></div>' +
						'<div class="tg-qr-item-text">' + cleanSnippet + '</div>' +
						'</div>';
				}
				html += '</div>';
				tgQRPopup.innerHTML = html;

				var items = tgQRPopup.querySelectorAll('.tg-qr-item');
				for (var j = 0; j < items.length; j++) {
					(function(idx) {
						var el = items[idx];
						el.addEventListener('mousedown', function(ev) {
							if (ev.preventDefault) ev.preventDefault();
							if (ev.stopPropagation) ev.stopPropagation();
							var chosen = tgQRMatches[idx];
							var tgt = tgQRTarget;
							if (chosen) {
								applyTgQuickReply(chosen, tgt);
							}
							hideTgQRPopup();
						});
					})(j);
				}
			}

			function tryExpandTgQuickReply(e) {
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
								if (typeof showTgToast === 'function') {
									showTgToast('⚡ Template: ' + item.key + ' diterapkan');
								}
								return true;
							}
						}
					}
					return false;
				}

				var editable = (target.closest && target.closest('[contenteditable="true"]')) ||
				               (target.isContentEditable ? target : null) ||
				               document.querySelector('.input-message-input[contenteditable="true"]') ||
				               document.querySelector('div[contenteditable="true"]');
				if (!editable) return false;

				var fullText = getTgEditorUserText(editable);
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

				applyTgQuickReplyToEditable(editable, matchItem.key, textToInsert, matchItem.key);
				return true;
			}

			// Input listener for slash typing in Telegram
			document.addEventListener('input', function(e) {
				var target = e.target;
				if (!target) return;

				var isInput = (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA');
				var isCE = target.isContentEditable || (target.closest && target.closest('[contenteditable="true"]'));
				if (!isInput && !isCE) {
					hideTgQRPopup();
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
						textBefore = getTgEditorUserText(host);
					}
				}

				var m = textBefore.match(/(?:^|\s)(\/[\w-]*)$/);
				if (!m) {
					hideTgQRPopup();
					return;
				}

				var query = m[1].toLowerCase();
				var all = loadQuickReplies();
				var matches = all.filter(function(it) {
					return it.key.toLowerCase().startsWith(query);
				});

				if (matches.length > 0) {
					showTgQRPopup(target, matches, m[1]);
				} else {
					hideTgQRPopup();
				}
			}, true);

			// Keydown listener for Telegram keyboard navigation & quick replies
			document.addEventListener('keydown', function(e) {
				var key = e.key;
				var code = e.code;
				var isTab = (key === 'Tab' || code === 'Tab' || e.keyCode === 9);
				var isEnter = (key === 'Enter' || code === 'Enter' || e.keyCode === 13);
				var isEscape = (key === 'Escape' || code === 'Escape' || e.keyCode === 27);
				var isUp = (key === 'ArrowUp' || code === 'ArrowUp' || e.keyCode === 38);
				var isDown = (key === 'ArrowDown' || code === 'ArrowDown' || e.keyCode === 40);
				var isSpace = (key === ' ' || code === 'Space' || e.keyCode === 32);

				if (tgQRPopup && tgQRPopup.style.display !== 'none' && tgQRMatches.length > 0) {
					if (isDown) {
						if (e.preventDefault) e.preventDefault();
						if (e.stopPropagation) e.stopPropagation();
						tgQRSelectedIndex = (tgQRSelectedIndex + 1) % tgQRMatches.length;
						renderTgQRPopup();
						return;
					}
					if (isUp) {
						if (e.preventDefault) e.preventDefault();
						if (e.stopPropagation) e.stopPropagation();
						tgQRSelectedIndex = (tgQRSelectedIndex - 1 + tgQRMatches.length) % tgQRMatches.length;
						renderTgQRPopup();
						return;
					}
					if (isEscape) {
						if (e.preventDefault) e.preventDefault();
						if (e.stopPropagation) e.stopPropagation();
						hideTgQRPopup();
						return;
					}
					if (isTab || isEnter) {
						if (e.preventDefault) e.preventDefault();
						if (e.stopPropagation) e.stopPropagation();
						var selected = tgQRMatches[tgQRSelectedIndex];
						var tgt = tgQRTarget || e.target;
						if (selected) {
							applyTgQuickReply(selected, tgt);
						}
						hideTgQRPopup();
						return;
					}
				}

				if (isTab || isEnter || isSpace) {
					var expanded = tryExpandTgQuickReply(e);
					if (expanded) {
						hideTgQRPopup();
					}
				}
			}, true);

			document.addEventListener('click', function(e) {
				if (tgQRPopup && tgQRPopup.style.display !== 'none') {
					if (!tgQRPopup.contains(e.target)) {
						hideTgQRPopup();
					}
				}
			}, true);
		})();
	})();

	// Feature: Telegram App Lock with PIN (Ctrl + L & 5-minute Inactivity Auto-Lock)
	(function() {
		var storedPin = '1234';
		var isLocked = false;
		try {
			storedPin = localStorage.getItem('tg_app_pin') || '1234';
			isLocked = localStorage.getItem('tg_is_locked') === 'true';
		} catch(e) {}

		window.updateStoredPin = function(pin) {
			if (pin && pin.length >= 4) {
				storedPin = pin;
				try { localStorage.setItem('tg_app_pin', pin); } catch(e) {}
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
				window.lockTelegram();
			}, INACTIVITY_TIMEOUT);
		}
		['mousedown', 'keydown', 'touchstart'].forEach(function(evt) {
			window.addEventListener(evt, resetInactivityTimer, { passive: true });
		});
		resetInactivityTimer();

		function showTgToast(msg) {
			var root = document.documentElement || document.body;
			if (!root) return;
			var toast = document.getElementById('tg-toast-msg');
			if (!toast) {
				toast = document.createElement('div');
				toast.id = 'tg-toast-msg';
				toast.style.cssText = 'position:fixed;bottom:70px;left:50%;transform:translateX(-50%);background:rgba(23,33,43,0.96);color:#ffffff;padding:8px 18px;border-radius:20px;font-size:12.5px;font-weight:500;box-shadow:0 8px 24px rgba(0,0,0,0.6);border:1px solid rgba(255,255,255,0.12);z-index:2147483647;pointer-events:none;transition:opacity 0.2s ease,transform 0.2s ease;opacity:0;font-family:-apple-system,BlinkMacSystemFont,\"Segoe UI\",Roboto,sans-serif;';
				root.appendChild(toast);
			}
			toast.textContent = msg;
			toast.style.opacity = '1';
			toast.style.transform = 'translateX(-50%) translateY(0)';
			clearTimeout(toast._timer);
			toast._timer = setTimeout(function() {
				toast.style.opacity = '0';
				toast.style.transform = 'translateX(-50%) translateY(6px)';
			}, 2500);
		}

		function createLockOverlay() {
			var existing = document.getElementById('tg-lock-overlay');
			if (existing) return existing;
			var root = document.documentElement || document.body;
			if (!root) return null;

			var overlay = document.createElement('div');
			overlay.id = 'tg-lock-overlay';
			overlay.style.cssText = 'display:none;position:fixed;inset:0;width:100vw;height:100vh;background:rgba(15,20,28,0.95);backdrop-filter:blur(18px);-webkit-backdrop-filter:blur(18px);z-index:2147483646;align-items:center;justify-content:center;user-select:none;font-family:-apple-system,BlinkMacSystemFont,\"Segoe UI\",Roboto,sans-serif;';
			overlay.innerHTML = '<div style="background:#17212b;border:1px solid rgba(255,255,255,0.1);border-radius:14px;width:330px;padding:30px 26px;text-align:center;box-shadow:0 24px 60px rgba(0,0,0,0.8);display:flex;flex-direction:column;align-items:center;box-sizing:border-box;">' +
				'<div style="width:48px;height:48px;border-radius:50%;background:rgba(36,161,222,0.15);border:1px solid rgba(36,161,222,0.3);display:flex;align-items:center;justify-content:center;color:#24A1DE;margin-bottom:14px;">' +
				'  <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>' +
				'</div>' +
				'<h2 style="margin:0 0 4px 0;font-size:18px;font-weight:600;color:#ffffff;">Telegram Terkunci</h2>' +
				'<p style="margin:0 0 18px 0;font-size:12.5px;color:#708499;">Masukkan PIN untuk membuka akses</p>' +
				'<input id="tg-pin-input" type="password" maxlength="6" style="width:180px;height:42px;text-align:center;font-size:22px;letter-spacing:8px;margin-bottom:12px;background:#0e1621;border:1px solid rgba(255,255,255,0.16);color:#ffffff;border-radius:8px;outline:none;box-sizing:border-box;font-family:inherit;">' +
				'<div id="tg-pin-error" style="color:#ef4444;font-size:12px;min-height:18px;margin-bottom:10px;"></div>' +
				'<button id="tg-pin-unlock-btn" style="width:100%;height:38px;font-size:13px;margin-bottom:12px;background:#24A1DE;color:#ffffff;font-weight:600;border:none;border-radius:8px;cursor:pointer;outline:none;transition:background 0.15s ease;font-family:inherit;">Buka Kunci</button>' +
				'<div style="font-size:11px;color:#6c7883;">Default: 1234 • Ctrl+L untuk mengunci</div>' +
				'<div style="margin-top:10px;"><a id="tg-pin-change-link" href="javascript:void(0)" style="color:#24A1DE;font-size:12px;text-decoration:none;font-weight:500;">Ganti / Ubah PIN</a></div>' +
				'<div style="margin-top:14px;padding-top:12px;border-top:1px solid rgba(255,255,255,0.08);width:100%;display:flex;gap:6px;justify-content:center;">' +
				'  <button id="tg-lock-btn-wa" style="background:#242f3d;color:#ffffff;border:1px solid rgba(0,168,132,0.35);padding:5px 12px;border-radius:14px;font-size:11px;cursor:pointer;outline:none;font-family:inherit;">' +
				'    <span style="color:#00a884;font-weight:bold;">WhatsApp</span> (Ctrl+1)' +
				'  </button>' +
				'</div>' +
				'</div>';
			root.appendChild(overlay);

			var btnLockWA = overlay.querySelector('#tg-lock-btn-wa');
			if (btnLockWA) {
				btnLockWA.onclick = function(e) {
					if (e.preventDefault) e.preventDefault();
					if (e.stopPropagation) e.stopPropagation();
					if (window.switchToWhatsApp) window.switchToWhatsApp();
					else if (window.openWhatsAppWindow) window.openWhatsAppWindow();
				};
			}

			var pinInput = document.getElementById('tg-pin-input');
			var unlockBtn = document.getElementById('tg-pin-unlock-btn');
			var errEl = document.getElementById('tg-pin-error');
			var changeLink = document.getElementById('tg-pin-change-link');

			if (changeLink) {
				changeLink.onclick = function(e) {
					if (e.preventDefault) e.preventDefault();
					window.openChangePinModal();
				};
			}

			function unlock() {
				var val = (pinInput ? pinInput.value : '').trim();
				if (val === storedPin) {
					overlay.style.display = 'none';
					pinInput.value = '';
					errEl.textContent = '';
					try { localStorage.setItem('tg_is_locked', 'false'); } catch(e) {}
					showTgToast('🔓 Telegram terbuka');
					resetInactivityTimer();
				} else {
					errEl.textContent = 'PIN salah! Coba lagi.';
					if (pinInput) {
						pinInput.value = '';
						pinInput.focus();
					}
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
			var existing = document.getElementById('tg-change-pin-modal');
			if (existing) return existing;
			var root = document.documentElement || document.body;
			if (!root) return null;

			var modal = document.createElement('div');
			modal.id = 'tg-change-pin-modal';
			modal.style.cssText = 'display:none;position:fixed;inset:0;width:100vw;height:100vh;background:rgba(0,0,0,0.7);backdrop-filter:blur(10px);-webkit-backdrop-filter:blur(10px);z-index:2147483647;align-items:center;justify-content:center;user-select:none;font-family:-apple-system,BlinkMacSystemFont,\"Segoe UI\",Roboto,sans-serif;';
			modal.innerHTML = '<div style="background:#17212b;border:1px solid rgba(255,255,255,0.12);border-radius:14px;width:340px;padding:26px 24px;text-align:center;box-shadow:0 24px 60px rgba(0,0,0,0.85);box-sizing:border-box;">' +
				'<div style="width:40px;height:40px;border-radius:50%;background:rgba(36,161,222,0.15);border:1px solid rgba(36,161,222,0.3);display:flex;align-items:center;justify-content:center;color:#24A1DE;margin:0 auto 12px auto;">' +
				'  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 2l-2 2m-1.5 1.5L10 13l-4 1 1-4 7.5-7.5m1.5-1.5l2-2"/><circle cx="7.5" cy="16.5" r="3.5"/></svg>' +
				'</div>' +
				'<h3 style="margin:0 0 4px 0;font-size:16px;font-weight:600;color:#ffffff;">Ubah PIN Telegram</h3>' +
				'<p style="margin:0 0 16px 0;font-size:12px;color:#708499;">Tentukan 4-6 angka PIN keamanan Anda</p>' +
				'<div style="text-align:left;margin-bottom:10px;">' +
				'<label style="font-size:11px;color:#8293a4;display:block;margin-bottom:4px;">PIN Saat Ini (Default: 1234):</label>' +
				'<input id="tg-cp-old" type="password" maxlength="6" placeholder="PIN Lama" style="width:100%;background:#0e1621;border:1px solid rgba(255,255,255,0.14);color:#ffffff;padding:8px 12px;border-radius:6px;font-size:12.5px;outline:none;box-sizing:border-box;">' +
				'</div>' +
				'<div style="text-align:left;margin-bottom:10px;">' +
				'<label style="font-size:11px;color:#8293a4;display:block;margin-bottom:4px;">PIN Baru (4-6 angka):</label>' +
				'<input id="tg-cp-new" type="password" maxlength="6" placeholder="PIN Baru" style="width:100%;background:#0e1621;border:1px solid rgba(255,255,255,0.14);color:#ffffff;padding:8px 12px;border-radius:6px;font-size:12.5px;outline:none;box-sizing:border-box;">' +
				'</div>' +
				'<div style="text-align:left;margin-bottom:12px;">' +
				'<label style="font-size:11px;color:#8293a4;display:block;margin-bottom:4px;">Konfirmasi PIN Baru:</label>' +
				'<input id="tg-cp-confirm" type="password" maxlength="6" placeholder="Ulangi PIN Baru" style="width:100%;background:#0e1621;border:1px solid rgba(255,255,255,0.14);color:#ffffff;padding:8px 12px;border-radius:6px;font-size:12.5px;outline:none;box-sizing:border-box;">' +
				'</div>' +
				'<div id="tg-cp-error" style="color:#ef4444;font-size:12px;min-height:16px;margin-bottom:12px;"></div>' +
				'<div style="display:flex;gap:8px;">' +
				'<button id="tg-cp-cancel-btn" style="flex:1;height:34px;background:#242f3d;color:#ffffff;border:1px solid rgba(255,255,255,0.1);border-radius:6px;font-size:12px;font-weight:500;cursor:pointer;outline:none;">Batal</button>' +
				'<button id="tg-cp-save-btn" style="flex:1;height:34px;background:#24A1DE;color:#ffffff;border:none;border-radius:6px;font-size:12px;font-weight:600;cursor:pointer;outline:none;">Simpan PIN</button>' +
				'</div></div>';
			root.appendChild(modal);

			var oldInput = document.getElementById('tg-cp-old');
			var newInput = document.getElementById('tg-cp-new');
			var confirmInput = document.getElementById('tg-cp-confirm');
			var errEl = document.getElementById('tg-cp-error');
			var cancelBtn = document.getElementById('tg-cp-cancel-btn');
			var saveBtn = document.getElementById('tg-cp-save-btn');

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
				try { localStorage.setItem('tg_app_pin', newVal); } catch(e) {}
				if (window.syncAppPin) {
					window.syncAppPin(newVal);
				}
				modal.style.display = 'none';
				oldInput.value = '';
				newInput.value = '';
				confirmInput.value = '';
				errEl.textContent = '';
				showTgToast('✅ PIN berhasil diubah!');
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
			var modal = createChangePinModal();
			if (modal) {
				modal.style.display = 'flex';
				var oldInput = document.getElementById('tg-cp-old');
				var errEl = document.getElementById('tg-cp-error');
				if (errEl) errEl.textContent = '';
				if (oldInput) {
					oldInput.value = '';
					setTimeout(function() { oldInput.focus(); }, 50);
				}
			}
		};

		window.lockTelegram = function() {
			var overlay = createLockOverlay();
			if (overlay) {
				if (overlay.style.display === 'flex') {
					return;
				}
				overlay.style.display = 'flex';
				try { localStorage.setItem('tg_is_locked', 'true'); } catch(e) {}
				var input = document.getElementById('tg-pin-input');
				if (input) {
					input.value = '';
					setTimeout(function() {
						if (document.activeElement !== input) {
							input.focus();
						}
					}, 50);
				}
			}
		};

		window.addEventListener('keydown', function(e) {
			if ((e.ctrlKey || e.metaKey) && (e.code === 'KeyL' || e.key === 'l' || e.key === 'L')) {
				if (e.preventDefault) e.preventDefault();
				if (e.stopPropagation) e.stopPropagation();
				window.lockTelegram();
			}
		}, true);

		document.addEventListener('keydown', function(e) {
			if ((e.ctrlKey || e.metaKey) && (e.code === 'KeyL' || e.key === 'l' || e.key === 'L')) {
				if (e.preventDefault) e.preventDefault();
				if (e.stopPropagation) e.stopPropagation();
				window.lockTelegram();
			}
		}, true);

		if (isLocked) {
			if (document.body || document.documentElement) {
				window.lockTelegram();
			} else {
				document.addEventListener('DOMContentLoaded', function() {
					window.lockTelegram();
				}, { once: true });
			}
		}
	})();

`


func sanitizeNotificationText(s string, maxLen int) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsPrint(r) || r == '\n' || r == '\r' || r == '\t' {
			b.WriteRune(r)
		}
	}
	res := strings.TrimSpace(b.String())
	if len(res) > maxLen {
		res = res[:maxLen] + "..."
	}
	return res
}

func showNativeNotification(title, message, iconPath string) {
	title = sanitizeNotificationText(title, 128)
	message = sanitizeNotificationText(message, 512)
	if title == "" && message == "" {
		return
	}

	notification := toast.Notification{
		AppID:   "WhatsApp Desktop",
		Title:   title,
		Message: message,
		Icon:    iconPath,
	}
	_ = notification.Push()
}

func openExternalLink(rawURL string) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return
	}
	// Strictly only allow http and https schemes to prevent protocol handler exploits
	if u.Scheme != "http" && u.Scheme != "https" {
		return
	}

	verbPtr, _ := syscall.UTF16PtrFromString("open")
	urlPtr, _ := syscall.UTF16PtrFromString(rawURL)
	procShellExecute.Call(0, uintptr(unsafe.Pointer(verbPtr)), uintptr(unsafe.Pointer(urlPtr)), 0, 0, 1)
}

func main() {
	isTelegram := false
	profileID := "1"
	startMinimized := false
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "--telegram" || arg == "-telegram" {
			isTelegram = true
		} else if arg == "--profile" && i+1 < len(os.Args) {
			profileID = os.Args[i+1]
			i++
		} else if strings.HasPrefix(arg, "--profile=") {
			profileID = strings.TrimPrefix(arg, "--profile=")
		} else if arg == "--minimized" || arg == "--tray" || arg == "-hidden" {
			startMinimized = true
		}
	}

	if isTelegram {
		currentWindowTitle = telegramWindowTitle
		currentMutexName = telegramMutexName
		_, isSingle := checkSingleInstance(telegramMutexName, telegramWindowTitle)
		if !isSingle {
			os.Exit(0)
		}
		configDir, err := os.UserConfigDir()
		if err != nil {
			configDir = os.Getenv("APPDATA")
			if configDir == "" {
				configDir = "."
			}
		}
		userDataDir := filepath.Join(configDir, "WhatsAppDesktopLight", "UserData_Telegram")
		_ = os.MkdirAll(userDataDir, 0755)
		runTelegramWindow(userDataDir, startMinimized)
		return
	}
	currentProfileID = profileID
	wTitle := windowTitle
	mName := mutexName
	if profileID != "1" {
		wTitle = fmt.Sprintf("WhatsApp Desktop (Akun %s)", profileID)
		mName = fmt.Sprintf("%s_%s", mutexName, profileID)
	}
	currentWindowTitle = wTitle
	currentMutexName = mName

	_, isSingle := checkSingleInstance(mName, wTitle)
	if !isSingle {
		os.Exit(0)
	}
	userDataDir := getUserDataDir(profileID)
	executablePath, _ := os.Executable()
	iconFullPath := filepath.Join(filepath.Dir(executablePath), "icon.ico")

	opts := webview2.WebViewOptions{
		Window:    nil,
		Debug:     false,
		DataPath:  userDataDir,
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title:  wTitle,
			Width:  1100,
			Height: 750,
			IconId: 2,
			Center: true,
		},
	}

	w := webview2.NewWithOptions(opts)
	if w == nil {
		log.Fatalln("Gagal inisialisasi WebView2")
	}
	defer w.Destroy()

	hwnd := uintptr(w.Window())
	globalHwnd = hwnd
	globalWebView = w

	setDarkWindowFrame(hwnd)
	w.SetTitle(wTitle)

	if startMinimized {
		procShowWindow.Call(hwnd, SW_HIDE)
	} else {
		w.SetSize(1100, 750, webview2.HintNone)
	}

	// Register Boss Key (Ctrl + Alt + W)
	procRegisterHotKey.Call(hwnd, HOTKEY_BOSS, MOD_CONTROL|MOD_ALT|MOD_NOREPEAT, uintptr('W'))
	defer procUnregisterHotKey.Call(hwnd, HOTKEY_BOSS)

	// Load Icon for System Tray
	hIcon, _, _ := procLoadImage.Call(0, uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(iconFullPath))), IMAGE_ICON, 16, 16, LR_LOADFROMFILE)
	if hIcon == 0 {
		r, _, _ := procSendMessage.Call(hwnd, WM_GETICON, ICON_SMALL, 0)
		hIcon = r
	}

	globalTrayData = NOTIFYICONDATAW{
		CbSize:           uint32(unsafe.Sizeof(NOTIFYICONDATAW{})),
		HWnd:             windows.Handle(hwnd),
		UID:              1,
		UFlags:           NIF_MESSAGE | NIF_ICON | NIF_TIP,
		UCallbackMessage: WM_TRAYICON,
		HIcon:            windows.Handle(hIcon),
	}
	utf16Title, _ := windows.UTF16FromString(wTitle)
	copy(globalTrayData.SzTip[:], utf16Title)
	procShellNotifyIcon.Call(NIM_ADD, uintptr(unsafe.Pointer(&globalTrayData)))
	defer removeTrayIcon()

	// Subclass window procedure for System Tray events and Close-to-Tray
	origWndProc = setWindowLongPtr(hwnd, GWLP_WNDPROC, windows.NewCallback(customWndProc))

	// Bind native notification bridge
	_ = w.Bind("sendNativeNotification", func(title, body string) {
		go showNativeNotification(title, body, iconFullPath)
	})

	// Bind external link handler (opens in system default browser)
	_ = w.Bind("openExternalLink", func(rawURL string) {
		openExternalLink(rawURL)
	})

	// Bind unread message counter
	_ = w.Bind("onUnreadCountChanged", func(count int) {
		updateUnreadCount(count)
	})

	// Bind privacy mode status sync
	_ = w.Bind("onPrivacyModeToggled", func(active bool) {
		privacyModeActive = active
	})

	// Bind Windows Autostart status sync
	_ = w.Bind("getAutoStartStatus", func() bool {
		return isAppAutoStartEnabled()
	})
	_ = w.Bind("setAutoStartStatus", func(enabled bool) bool {
		return setAppAutoStartEnabled(enabled)
	})

	// Bind dual account launcher
	_ = w.Bind("openDualAccount", func() {
		launchDualAccount()
	})

	// Bind unified tab switcher and side-by-side mode
	_ = w.Bind("openTelegramWindow", func() {
		switchToTelegram()
	})
	_ = w.Bind("switchToTelegram", func() {
		switchToTelegram()
	})
	_ = w.Bind("openWhatsAppWindow", func() {
		switchToWhatsApp()
	})
	_ = w.Bind("switchToWhatsApp", func() {
		switchToWhatsApp()
	})
	_ = w.Bind("openSideBySideView", func() {
		switchToSplitView()
	})
	_ = w.Bind("switchToSplitView", func() {
		switchToSplitView()
	})
	_ = w.Bind("getAppPin", func() string {
		return loadSharedPin()
	})
	_ = w.Bind("syncAppPin", func(pin string) {
		saveSharedPin(pin)
		if globalWebView != nil {
			globalWebView.Eval(fmt.Sprintf("if(window.updateStoredPin) window.updateStoredPin(%q);", pin))
		}
		if globalTelegramWebView != nil {
			globalTelegramWebView.Eval(fmt.Sprintf("if(window.updateStoredPin) window.updateStoredPin(%q);", pin))
		}
	})

	// Injected JavaScript: User-Agent, Notification Polyfill, Link Isolation, Privacy Mode, and Unread Message Observer
	initScript := `
		// UserAgent override
		Object.defineProperty(navigator, 'userAgent', {
			get: () => '` + userAgent + `'
		});
		Object.defineProperty(navigator, 'appVersion', {
			get: () => '` + userAgent + `'
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
			blurIntensity: 4
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
			style.textContent = [
				':root {',
				'  --wa-bg: #111b21;',
				'  --wa-bg-elevated: #182229;',
				'  --wa-bg-card: #1f2c34;',
				'  --wa-bg-hover: #222e35;',
				'  --wa-border: rgba(134, 150, 160, 0.14);',
				'  --wa-border-strong: rgba(134, 150, 160, 0.28);',
				'  --wa-primary: #00a884;',
				'  --wa-primary-hover: #02906f;',
				'  --wa-primary-subtle: rgba(0, 168, 132, 0.12);',
				'  --wa-tg: #24A1DE;',
				'  --wa-tg-subtle: rgba(36, 161, 222, 0.12);',
				'  --wa-danger: #ef4444;',
				'  --wa-danger-subtle: rgba(239, 68, 68, 0.12);',
				'  --wa-text: #e9edef;',
				'  --wa-text-muted: #8696a0;',
				'  --wa-text-dim: #667781;',
				'  --wa-font: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;',
				'  --wa-shadow-modal: 0 24px 60px rgba(0, 0, 0, 0.75), 0 0 0 1px rgba(255, 255, 255, 0.08);',
				'  --wa-shadow-dock: 0 10px 30px rgba(0, 0, 0, 0.55), 0 0 0 1px rgba(255, 255, 255, 0.08);',
				'  --wa-radius-sm: 6px;',
				'  --wa-radius-md: 10px;',
				'  --wa-radius-lg: 14px;',
				'  --wa-ease: cubic-bezier(0.16, 1, 0.3, 1);',
				'}',
				'#wa-addon-modal ::-webkit-scrollbar, #wa-scratchpad-drawer ::-webkit-scrollbar, #wa-qr-list ::-webkit-scrollbar { width: 5px; height: 5px; }',
				'#wa-addon-modal ::-webkit-scrollbar-track, #wa-scratchpad-drawer ::-webkit-scrollbar-track, #wa-qr-list ::-webkit-scrollbar-track { background: transparent; }',
				'#wa-addon-modal ::-webkit-scrollbar-thumb, #wa-scratchpad-drawer ::-webkit-scrollbar-thumb, #wa-qr-list ::-webkit-scrollbar-thumb { background: rgba(134, 150, 160, 0.22); border-radius: 4px; }',
				'#wa-addon-modal ::-webkit-scrollbar-thumb:hover, #wa-scratchpad-drawer ::-webkit-scrollbar-thumb:hover, #wa-qr-list ::-webkit-scrollbar-thumb:hover { background: rgba(134, 150, 160, 0.38); }',
				'.wa-modal-backdrop { display:none; position:fixed; inset:0; width:100vw; height:100vh; background:rgba(0,0,0,0.65); backdrop-filter:blur(10px); -webkit-backdrop-filter:blur(10px); z-index:99999999; align-items:center; justify-content:center; font-family:var(--wa-font); user-select:none; animation:waFadeIn 0.18s var(--wa-ease); }',
				'@keyframes waFadeIn { from { opacity: 0; } to { opacity: 1; } }',
				'.wa-modal-box { background:var(--wa-bg); border:1px solid var(--wa-border); color:var(--wa-text); width:600px; max-height:86vh; border-radius:var(--wa-radius-lg); box-shadow:var(--wa-shadow-modal); display:flex; flex-direction:column; overflow:hidden; box-sizing:border-box; animation:waSlideUp 0.2s var(--wa-ease); }',
				'@keyframes waSlideUp { from { opacity:0; transform:translateY(8px) scale(0.98); } to { opacity:1; transform:translateY(0) scale(1); } }',
				'.wa-modal-header { padding:16px 22px; border-bottom:1px solid var(--wa-border); display:flex; align-items:center; justify-content:space-between; background:var(--wa-bg-elevated); }',
				'.wa-header-left { display:flex; align-items:center; gap:12px; }',
				'.wa-header-icon { width:34px; height:34px; border-radius:9px; background:var(--wa-primary-subtle); border:1px solid rgba(0,168,132,0.25); display:flex; align-items:center; justify-content:center; color:var(--wa-primary); flex-shrink:0; }',
				'.wa-header-title { margin:0; font-size:15px; font-weight:600; color:var(--wa-text); letter-spacing:-0.01em; }',
				'.wa-header-subtitle { margin:2px 0 0 0; font-size:11.5px; color:var(--wa-text-muted); }',
				'.wa-close-btn { background:none; border:none; color:var(--wa-text-muted); cursor:pointer; width:30px; height:30px; border-radius:50%; display:flex; align-items:center; justify-content:center; transition:all 0.15s ease; outline:none; }',
				'.wa-close-btn:hover { color:var(--wa-text); background:rgba(255,255,255,0.06); }',
				'.wa-tab-bar { display:flex; border-bottom:1px solid var(--wa-border); background:var(--wa-bg); padding:0 16px; gap:4px; }',
				'.wa-tab-btn { background:none; border:none; color:var(--wa-text-muted); padding:11px 14px; font-size:12px; font-weight:500; cursor:pointer; border-bottom:2px solid transparent; transition:all 0.18s var(--wa-ease); font-family:var(--wa-font); outline:none; user-select:none; display:flex; align-items:center; gap:7px; }',
				'.wa-tab-btn:hover { color:var(--wa-text); }',
				'.wa-tab-btn.active { color:var(--wa-primary); font-weight:600; border-bottom-color:var(--wa-primary); }',
				'.wa-modal-body { flex:1; overflow-y:auto; padding:18px 22px; display:flex; flex-direction:column; gap:14px; }',
				'.wa-card { background:var(--wa-bg-elevated); border-radius:var(--wa-radius-md); border:1px solid var(--wa-border); padding:15px 18px; display:flex; flex-direction:column; gap:12px; }',
				'.wa-card-header { display:flex; align-items:center; justify-content:space-between; margin-bottom:2px; }',
				'.wa-card-title { font-size:12px; font-weight:600; color:var(--wa-primary); display:flex; align-items:center; gap:7px; text-transform:uppercase; letter-spacing:0.04em; }',
				'.wa-row { display:flex; align-items:center; justify-content:space-between; gap:16px; }',
				'.wa-row + .wa-row { border-top:1px solid var(--wa-border); padding-top:11px; }',
				'.wa-row-title { font-size:13px; font-weight:500; color:var(--wa-text); }',
				'.wa-row-desc { font-size:11.5px; color:var(--wa-text-muted); line-height:1.4; margin-top:1px; }',
				'.wa-switch { position:relative; display:inline-block; width:36px; height:20px; flex-shrink:0; cursor:pointer; }',
				'.wa-switch input { opacity:0; width:0; height:0; position:absolute; }',
				'.wa-slider { position:absolute; cursor:pointer; inset:0; background-color:#2a3942; transition:all 0.22s var(--wa-ease); border-radius:20px; border:1px solid rgba(255,255,255,0.06); }',
				'.wa-slider:before { position:absolute; content:""; height:14px; width:14px; left:2px; bottom:2px; background-color:#e9edef; transition:transform 0.22s var(--wa-ease); border-radius:50%; box-shadow:0 1px 3px rgba(0,0,0,0.4); }',
				'.wa-switch input:checked + .wa-slider { background-color:var(--wa-primary); border-color:var(--wa-primary); }',
				'.wa-switch input:checked + .wa-slider:before { transform:translateX(16px); background-color:#ffffff; }',
				'.wa-btn { font-family:var(--wa-font); display:inline-flex; align-items:center; justify-content:center; gap:6px; font-size:12px; font-weight:500; padding:7px 14px; border-radius:var(--wa-radius-sm); cursor:pointer; outline:none; transition:all 0.15s var(--wa-ease); border:1px solid transparent; user-select:none; white-space:nowrap; }',
				'.wa-btn-primary { background:var(--wa-primary); color:#111b21; font-weight:600; border-color:var(--wa-primary); }',
				'.wa-btn-primary:hover { background:var(--wa-primary-hover); border-color:var(--wa-primary-hover); }',
				'.wa-btn-secondary { background:var(--wa-bg); color:var(--wa-text); border-color:var(--wa-border); }',
				'.wa-btn-secondary:hover { background:var(--wa-bg-hover); border-color:var(--wa-border-strong); }',
				'.wa-btn-tg { background:var(--wa-tg); color:#ffffff; font-weight:600; border-color:var(--wa-tg); }',
				'.wa-btn-tg:hover { background:#1d88be; }',
				'.wa-btn-danger { background:var(--wa-danger-subtle); color:var(--wa-danger); border-color:rgba(239,68,68,0.3); font-weight:600; }',
				'.wa-btn-danger:hover { background:rgba(239,68,68,0.2); }',
				'.wa-input { font-family:var(--wa-font); background:var(--wa-bg); border:1px solid var(--wa-border); color:var(--wa-text); padding:8px 12px; border-radius:var(--wa-radius-sm); font-size:12.5px; outline:none; transition:all 0.15s ease; box-sizing:border-box; }',
				'.wa-input:focus { border-color:var(--wa-primary); box-shadow:0 0 0 1px var(--wa-primary); }',
				'.wa-input::placeholder { color:var(--wa-text-dim); }',
				'.wa-theme-grid { display:grid; grid-template-columns:1fr 1fr 1fr; gap:10px; }',
				'.wa-theme-card { background:var(--wa-bg); border:1px solid var(--wa-border); border-radius:var(--wa-radius-md); padding:12px; display:flex; flex-direction:column; align-items:center; gap:8px; cursor:pointer; transition:all 0.15s ease; color:var(--wa-text); font-size:12px; font-weight:500; }',
				'.wa-theme-card:hover { background:var(--wa-bg-hover); border-color:var(--wa-border-strong); }',
				'.wa-theme-card.active { border-color:var(--wa-primary); background:var(--wa-primary-subtle); color:var(--wa-primary); font-weight:600; }',
				'.wa-theme-swatch { width:100%; height:24px; border-radius:4px; border:1px solid rgba(255,255,255,0.08); }',
				'.wa-segmented { display:inline-flex; background:var(--wa-bg); border:1px solid var(--wa-border); border-radius:var(--wa-radius-sm); padding:2px; gap:2px; }',
				'.wa-segmented-btn { background:none; border:none; color:var(--wa-text-muted); font-size:11.5px; font-weight:500; padding:5px 12px; border-radius:4px; cursor:pointer; transition:all 0.15s ease; font-family:var(--wa-font); outline:none; }',
				'.wa-segmented-btn:hover { color:var(--wa-text); }',
				'.wa-segmented-btn.active { background:var(--wa-primary); color:#111b21; font-weight:600; }',
				'.wa-dock { position:fixed !important; bottom:14px !important; left:14px !important; z-index:2147483647 !important; display:flex !important; align-items:center !important; gap:3px !important; background:rgba(17,27,33,0.96) !important; backdrop-filter:blur(20px) !important; -webkit-backdrop-filter:blur(20px) !important; border:1px solid rgba(255,255,255,0.12) !important; padding:3px 5px !important; border-radius:24px !important; box-shadow:var(--wa-shadow-dock) !important; user-select:none !important; font-family:var(--wa-font) !important; pointer-events:auto !important; visibility:visible !important; }',
				'.wa-dock-item { display:flex; align-items:center; gap:6px; padding:5px 11px; border-radius:18px; font-size:12px; font-weight:500; color:var(--wa-text-muted); background:transparent; border:none; cursor:pointer; transition:all 0.15s var(--wa-ease); outline:none; white-space:nowrap; }',
				'.wa-dock-item:hover { color:var(--wa-text); background:rgba(255,255,255,0.06); }',
				'.wa-dock-item.active { background:rgba(0,168,132,0.18); color:var(--wa-primary); font-weight:600; cursor:default; }',
				'.wa-dock-sep { width:1px; height:14px; background:var(--wa-border); margin:0 1px; }',
				'.wa-status-dot { width:6px; height:6px; border-radius:50%; background:var(--wa-primary); display:inline-block; }',
				'.wa-toast { position:fixed; bottom:24px; right:24px; background:rgba(17,27,33,0.94); backdrop-filter:blur(16px); -webkit-backdrop-filter:blur(16px); color:var(--wa-text); border:1px solid var(--wa-border-strong); padding:10px 18px; border-radius:var(--wa-radius-md); font-size:12.5px; font-weight:500; z-index:999999999; font-family:var(--wa-font); box-shadow:0 12px 32px rgba(0,0,0,0.5); pointer-events:none; opacity:0; transform:translateY(6px); transition:all 0.22s var(--wa-ease); display:flex; align-items:center; gap:8px; }',
				'.wa-toast.show { opacity:1; transform:translateY(0); }',
				'.wa-scratchpad { position:fixed; top:0; right:-360px; width:340px; height:100vh; background:var(--wa-bg); border-left:1px solid var(--wa-border); z-index:9999998; box-shadow:-8px 0 32px rgba(0,0,0,0.7); display:flex; flex-direction:column; font-family:var(--wa-font); transition:right 0.25s var(--wa-ease); box-sizing:border-box; user-select:none; }',
				'.wa-qr-popup { position:fixed !important; z-index:2147483647 !important; background:var(--wa-bg-elevated) !important; border:1px solid var(--wa-border-strong) !important; border-radius:var(--wa-radius-md) !important; box-shadow:0 16px 40px rgba(0,0,0,0.65), 0 0 0 1px rgba(255,255,255,0.06) !important; width:380px !important; max-width:90vw !important; max-height:260px !important; display:none; flex-direction:column !important; overflow:hidden !important; font-family:var(--wa-font) !important; user-select:none !important; animation:waSlideUp 0.15s var(--wa-ease) !important; pointer-events:auto !important; }',
				'.wa-qr-header { padding:8px 12px !important; background:var(--wa-bg) !important; border-bottom:1px solid var(--wa-border) !important; display:flex !important; align-items:center !important; justify-content:space-between !important; font-size:11px !important; color:var(--wa-text-muted) !important; }',
				'.wa-qr-badge { font-size:10px !important; padding:2px 6px !important; border-radius:4px !important; background:var(--wa-primary-subtle) !important; color:var(--wa-primary) !important; font-weight:600 !important; }',
				'.wa-qr-list { overflow-y:auto !important; padding:4px !important; display:flex !important; flex-direction:column !important; gap:2px !important; max-height:210px !important; }',
				'.wa-qr-item { display:flex !important; flex-direction:column !important; padding:8px 10px !important; border-radius:var(--wa-radius-sm) !important; cursor:pointer !important; transition:all 0.12s ease !important; border:1px solid transparent !important; }',
				'.wa-qr-item:hover, .wa-qr-item.active { background:var(--wa-bg-hover) !important; border-color:var(--wa-border) !important; }',
				'.wa-qr-item.active { background:var(--wa-primary-subtle) !important; border-color:rgba(0, 168, 132, 0.35) !important; }',
				'.wa-qr-item-key { font-size:12.5px !important; font-weight:600 !important; color:var(--wa-primary) !important; display:flex !important; align-items:center !important; gap:6px !important; }',
				'.wa-qr-item-text { font-size:11.5px !important; color:var(--wa-text-muted) !important; white-space:nowrap !important; overflow:hidden !important; text-overflow:ellipsis !important; margin-top:2px !important; }'
			].join('\n');
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
					'body.wa-privacy-active #pane-side [role="listitem"] [title]',
					'body.wa-privacy-active #pane-side [role="row"] span[dir]',
					'body.wa-privacy-active #pane-side span[title]'
				);
			}
			if (privacyConfig.blurPreview) {
				selectors.push(
					'body.wa-privacy-active #pane-side [role="gridcell"]',
					'body.wa-privacy-active #pane-side [role="listitem"] div[dir]'
				);
			}
			if (privacyConfig.blurMessages) {
				selectors.push(
					'body.wa-privacy-active .message-in',
					'body.wa-privacy-active .message-out',
					'body.wa-privacy-active div[data-testid="msg-container"]',
					'body.wa-privacy-active div[class*="message-in"]',
					'body.wa-privacy-active div[class*="message-out"]'
				);
			}
			if (privacyConfig.blurMedia) {
				selectors.push(
					'body.wa-privacy-active div[data-testid="cell-frame-container"] img',
					'body.wa-privacy-active div[data-testid="image-thumb"]',
					'body.wa-privacy-active div[data-testid="media-canvas"]',
					'body.wa-privacy-active div[data-testid="msg-container"] img',
					'body.wa-privacy-active div[data-testid="msg-container"] video'
				);
			}
			if (privacyConfig.blurAvatars) {
				selectors.push(
					'body.wa-privacy-active #pane-side img',
					'body.wa-privacy-active header img'
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

		function applyQuickReplyToEditable(editable, matchKey, textToInsert, typedWord) {
			if (!editable) return false;
			var userText = getEditorUserText(editable);
			placeCaretAtEnd(editable);

			var m = userText.match(/(?:^|\s)(\/[\w-]*)$/);
			var slashCmd = m ? m[1] : (typedWord || matchKey);

			var delCount = slashCmd.length;
			for (var k = 0; k < delCount; k++) {
				document.execCommand('delete');
			}

			document.execCommand('insertText', false, textToInsert);
			editable.dispatchEvent(new InputEvent('input', { bubbles: true, cancelable: true, inputType: 'insertText', data: textToInsert }));
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

		// Feature: In-App Scratchpad / Quick Notes
		function createScratchpadDrawer() {
			var existing = document.getElementById('wa-scratchpad-drawer');
			if (existing) return existing;
			if (!document.body) return null;

			var drawer = document.createElement('div');
			drawer.id = 'wa-scratchpad-drawer';
			drawer.className = 'wa-scratchpad';

			drawer.innerHTML = '<div class="wa-modal-header" style="padding:14px 18px;">' +
				'<div class="wa-header-left">' +
				'  <div class="wa-header-icon" style="width:30px;height:30px;">' +
				'    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>' +
				'  </div>' +
				'  <h3 class="wa-header-title" style="font-size:14px;">Catatan Tempel</h3>' +
				'</div>' +
				'<button id="wa-sp-close" class="wa-close-btn" title="Tutup">' +
				'  <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>' +
				'</button>' +
				'</div>' +
				'<div style="flex:1;padding:14px 18px;display:flex;flex-direction:column;">' +
				'<textarea id="wa-sp-textarea" class="wa-input" placeholder="Tulis draf pesan, nomor resi, to-do list, catatan telepon di sini...\\n\\nTersimpan otomatis." style="flex:1;width:100%;resize:none;line-height:1.5;"></textarea>' +
				'<div style="display:flex;align-items:center;justify-content:space-between;margin-top:10px;font-size:11px;color:var(--wa-text-muted);">' +
				'<span id="wa-sp-count">0 karakter</span>' +
				'<div style="display:flex;gap:6px;">' +
				'<button id="wa-sp-copy" class="wa-btn wa-btn-secondary" style="font-size:11px;padding:4px 10px;">Salin</button>' +
				'<button id="wa-sp-clear" class="wa-btn wa-btn-danger" style="font-size:11px;padding:4px 10px;">Hapus</button>' +
				'</div></div></div>' +
				'<div style="padding:10px 18px;border-top:1px solid var(--wa-border);font-size:11px;color:var(--wa-text-dim);text-align:center;">' +
				'Pintasan <b style="color:var(--wa-text-muted);">Alt + N</b> untuk membuka / menutup catatan' +
				'</div>';

			document.body.appendChild(drawer);

			var ta = document.getElementById('wa-sp-textarea');
			var count = document.getElementById('wa-sp-count');
			var closeBtn = document.getElementById('wa-sp-close');
			var copyBtn = document.getElementById('wa-sp-copy');
			var clearBtn = document.getElementById('wa-sp-clear');

			try {
				var savedNotes = localStorage.getItem('wa_scratchpad_notes') || '';
				if (ta) {
					ta.value = savedNotes;
					if (count) count.textContent = savedNotes.length + ' karakter';
				}
			} catch(e) {}

			if (ta) {
				ta.oninput = function() {
					try { localStorage.setItem('wa_scratchpad_notes', ta.value); } catch(e) {}
					if (count) count.textContent = ta.value.length + ' karakter';
				};
			}

			if (closeBtn) closeBtn.onclick = function() { drawer.style.right = '-340px'; };

			if (copyBtn) {
				copyBtn.onclick = function() {
					if (ta && ta.value) {
						navigator.clipboard.writeText(ta.value).then(function() {
							showAddonToast('📋 Catatan disalin ke clipboard');
							copyBtn.textContent = '✓ Tersalin!';
							copyBtn.style.background = '#00a884';
							copyBtn.style.color = '#111b21';
							setTimeout(function() {
								copyBtn.textContent = '📋 Salin';
								copyBtn.style.background = '#111b21';
								copyBtn.style.color = '#00a884';
							}, 1800);
						});
					}
				};
			}

			if (clearBtn) {
				clearBtn.onclick = function() {
					if (confirm('Bersihkan isi catatan?')) {
						if (ta) {
							ta.value = '';
							try { localStorage.setItem('wa_scratchpad_notes', ''); } catch(e) {}
							if (count) count.textContent = '0 karakter';
						}
					}
				};
			}

			return drawer;
		}

		window.toggleScratchpad = function() {
			whenDOMReady(function() {
				var drawer = createScratchpadDrawer();
				if (!drawer) return;
				var isOpen = drawer.style.right === '0px';
				drawer.style.right = isOpen ? '-340px' : '0px';
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
					'#pane-side [role="listitem"] [title], #pane-side span[title], header span[dir="auto"], div[role="listitem"] span[dir="auto"] {',
					'  color: transparent !important;',
					'  text-shadow: 0 0 10px rgba(255,255,255,0.85) !important;',
					'}',
					'#pane-side img, header img, div[role="listitem"] img {',
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
				modal.innerHTML = '<div class="wa-modal-box">' +
					'    <!-- Top Header -->' +
					'    <div class="wa-modal-header">' +
					'      <div class="wa-header-left">' +
					'        <div class="wa-header-icon">' +
					'          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="4" y1="21" x2="4" y2="14"/><line x1="4" y1="10" x2="4" y2="3"/><line x1="12" y1="21" x2="12" y2="12"/><line x1="12" y1="8" x2="12" y2="3"/><line x1="20" y1="21" x2="20" y2="16"/><line x1="20" y1="12" x2="20" y2="3"/><line x1="1" y1="14" x2="7" y2="14"/><line x1="9" y1="8" x2="15" y2="8"/><line x1="17" y1="16" x2="23" y2="16"/></svg>' +
					'        </div>' +
					'        <div>' +
					'          <h3 class="wa-header-title">Pusat Kontrol & Fitur Tambahan</h3>' +
					'          <p class="wa-header-subtitle">Kustomisasi mode privasi, tampilan, dan alat produktivitas</p>' +
					'        </div>' +
					'      </div>' +
					'      <button id="wa-cc-close" class="wa-close-btn" title="Tutup (Esc)">' +
					'        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>' +
					'      </button>' +
					'    </div>' +
					'' +
					'    <!-- Navigation Tabs -->' +
					'    <div class="wa-tab-bar">' +
					'      <button class="wa-tab-btn active" data-tab="stealth">' +
					'        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>' +
					'        <span>Privasi Layar</span>' +
					'      </button>' +
					'      <button class="wa-tab-btn" data-tab="theme">' +
					'        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20 10 10 0 0 0 0-20"/></svg>' +
					'        <span>Tampilan & Tema</span>' +
					'      </button>' +
					'      <button class="wa-tab-btn" data-tab="productivity">' +
					'        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>' +
					'        <span>Alat & Percakapan</span>' +
					'      </button>' +
					'      <button class="wa-tab-btn" data-tab="system">' +
					'        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>' +
					'        <span>Sistem & Keamanan</span>' +
					'      </button>' +
					'    </div>' +
					'' +
					'    <!-- Tab Contents Container -->' +
					'    <div class="wa-modal-body">' +
					'' +
					'      <!-- TAB 1: PRIVASI LAYAR -->' +
					'      <div id="wa-tab-content-stealth" class="wa-tab-pane" style="display:flex;flex-direction:column;gap:14px;">' +
					'        <div class="wa-card">' +
					'          <div class="wa-card-header">' +
					'            <div>' +
					'              <div class="wa-card-title">' +
					'                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/></svg>' +
					'                <span>Mode Privasi (Blur Layar - Alt + P)</span>' +
					'              </div>' +
					'              <div class="wa-row-desc" style="margin-top:2px;">Buramkan obrolan agar aman dari pandangan rekan di sekitar (Bisa diintip saat hover kursor)</div>' +
					'            </div>' +
					'            <button id="wa-cc-privacy-btn" class="wa-btn wa-btn-secondary" style="font-size:11.5px;padding:5px 12px;">Toggle Privasi</button>' +
					'          </div>' +
					'          <div style="font-size:11.5px;color:var(--wa-text-dim);background:rgba(255,255,255,0.03);padding:6px 10px;border-radius:var(--wa-radius-sm);border:1px solid var(--wa-border);">' +
					'            💡 <b>Hover to Peek:</b> Dekatkan kursor mouse ke pesan/kontak yang diburamkan untuk melihat isinya sementara.' +
					'          </div>' +
					'          <div style="display:grid;grid-template-columns:1fr 1fr;gap:10px;font-size:12.5px;color:var(--wa-text);margin-top:2px;">' +
					'            <label style="display:flex;align-items:center;gap:8px;cursor:pointer;"><input type="checkbox" id="wa-cc-blur-contacts"> Nama Kontak</label>' +
					'            <label style="display:flex;align-items:center;gap:8px;cursor:pointer;"><input type="checkbox" id="wa-cc-blur-preview"> Preview Pesan</label>' +
					'            <label style="display:flex;align-items:center;gap:8px;cursor:pointer;"><input type="checkbox" id="wa-cc-blur-messages"> Balon Obrolan</label>' +
					'            <label style="display:flex;align-items:center;gap:8px;cursor:pointer;"><input type="checkbox" id="wa-cc-blur-media"> Foto & Video</label>' +
					'            <label style="display:flex;align-items:center;gap:8px;cursor:pointer;"><input type="checkbox" id="wa-cc-blur-avatars"> Foto Profil</label>' +
					'          </div>' +
					'          <div class="wa-row" style="margin-top:4px;">' +
					'            <span style="font-size:12px;color:var(--wa-text-muted);">Intensitas Efek Blur:</span>' +
					'            <div class="wa-segmented">' +
					'              <button id="wa-blur-btn-3" class="wa-segmented-btn">Halus (3px)</button>' +
					'              <button id="wa-blur-btn-5" class="wa-segmented-btn">Sedang (5px)</button>' +
					'              <button id="wa-blur-btn-8" class="wa-segmented-btn">Kuat (8px)</button>' +
					'            </div>' +
					'          </div>' +
					'        </div>' +
					'' +
					'        <div class="wa-card">' +
					'          <div class="wa-card-header">' +
					'            <div>' +
					'              <div class="wa-card-title">' +
					'                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>' +
					'                <span>Mode Screenshot Anonim</span>' +
					'              </div>' +
					'              <div class="wa-row-desc" style="margin-top:2px;">Samarkan nama dan nomor kontak sebelum mengambil tangkapan layar chat</div>' +
					'            </div>' +
					'            <button id="wa-cc-anon-btn" class="wa-btn wa-btn-secondary" style="font-size:11.5px;padding:5px 12px;">Aktifkan</button>' +
					'          </div>' +
					'        </div>' +
					'' +
					'        <div class="wa-card">' +
					'          <div class="wa-card-header">' +
					'            <div>' +
					'              <div class="wa-card-title">' +
					'                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>' +
					'                <span>Filter Chat Belum Dibaca</span>' +
					'              </div>' +
					'              <div class="wa-row-desc" style="margin-top:2px;">Hanya tampilkan obrolan yang belum dibaca di panel samping</div>' +
					'            </div>' +
					'            <button id="wa-cc-unread-toggle" class="wa-btn wa-btn-secondary" style="font-size:11.5px;padding:5px 12px;">Aktifkan Filter</button>' +
					'          </div>' +
					'        </div>' +
					'' +
					'        <div class="wa-card" style="border-left:3px solid var(--wa-primary);">' +
					'          <div style="display:flex;align-items:flex-start;gap:10px;">' +
					'            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="var(--wa-primary)" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="flex-shrink:0;margin-top:2px;"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>' +
					'            <div>' +
					'              <div style="font-size:12px;font-weight:600;color:var(--wa-text);">Privasi Layar 100% Aman & Anti-Banned</div>' +
					'              <div class="wa-row-desc" style="margin-top:2px;line-height:1.45;">Aplikasi ini tidak memodifikasi transmisi enkripsi end-to-end (Noise Protocol) WhatsApp resmi sehingga akun Anda dijamin 100% aman dan tidak berisiko diblokir oleh Meta.</div>' +
					'            </div>' +
					'          </div>' +
					'        </div>' +
					'      </div>' +
					'' +
					'      <!-- TAB 2: TAMPILAN & TEMA -->' +
					'      <div id="wa-tab-content-theme" class="wa-tab-pane" style="display:none;flex-direction:column;gap:14px;">' +
					'        <div class="wa-card">' +
					'          <div class="wa-card-header">' +
					'            <div class="wa-card-title">' +
					'              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M12 2a14.5 14.5 0 0 0 0 20 10 10 0 0 0 0-20"/></svg>' +
					'              <span>Tema & Kenyamanan Visual</span>' +
					'            </div>' +
					'          </div>' +
					'          <div class="wa-theme-grid">' +
					'            <div id="wa-theme-default" class="wa-theme-card active">' +
					'              <div class="wa-theme-swatch" style="background:#111b21;"></div>' +
					'              <span>Standar Gelap</span>' +
					'            </div>' +
					'            <div id="wa-theme-oled" class="wa-theme-card">' +
					'              <div class="wa-theme-swatch" style="background:#000000;"></div>' +
					'              <span>OLED Black</span>' +
					'            </div>' +
					'            <div id="wa-theme-warm" class="wa-theme-card">' +
					'              <div class="wa-theme-swatch" style="background:#1a1e21;border-color:rgba(217,119,6,0.3);"></div>' +
					'              <span>Layar Hangat</span>' +
					'            </div>' +
					'          </div>' +
					'          <div class="wa-row">' +
					'            <div>' +
					'              <div class="wa-row-title">Daftar Chat Rapat (Compact Mode)</div>' +
					'              <div class="wa-row-desc">Memadatkan baris daftar obrolan agar memuat lebih banyak kontak</div>' +
					'            </div>' +
					'            <label class="wa-switch"><input type="checkbox" id="wa-compact-cb"><span class="wa-slider"></span></label>' +
					'          </div>' +
					'        </div>' +
					'' +
					'        <div class="wa-card">' +
					'          <div class="wa-card-header">' +
					'            <div class="wa-card-title">' +
					'              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>' +
					'              <span>Skala Tampilan & Kecepatan Suara</span>' +
					'            </div>' +
					'          </div>' +
					'          <div class="wa-row">' +
					'            <div>' +
					'              <div class="wa-row-title">Zoom Antarmuka</div>' +
					'              <div class="wa-row-desc">Shortcut: Ctrl + / Ctrl - / Ctrl 0</div>' +
					'            </div>' +
					'            <div class="wa-segmented">' +
					'              <button id="wa-cc-zoom-out" class="wa-segmented-btn" title="Perkecil">−</button>' +
					'              <button id="wa-cc-zoom-reset" class="wa-segmented-btn" title="Reset (100%)">Reset</button>' +
					'              <button id="wa-cc-zoom-in" class="wa-segmented-btn" title="Perbesar">+</button>' +
					'            </div>' +
					'          </div>' +
					'          <div class="wa-row">' +
					'            <div>' +
					'              <div class="wa-row-title">Kecepatan Default Voice Note</div>' +
					'              <div class="wa-row-desc">Atur laju pemutaran rekaman suara WhatsApp</div>' +
					'            </div>' +
					'            <div style="display:flex;gap:4px;">' +
					'              <button class="wa-audio-btn wa-btn wa-btn-secondary" data-spd="1.0" style="padding:4px 8px;font-size:11px;">1.0x</button>' +
					'              <button class="wa-audio-btn wa-btn wa-btn-secondary" data-spd="1.25" style="padding:4px 8px;font-size:11px;">1.25x</button>' +
					'              <button class="wa-audio-btn wa-btn wa-btn-secondary" data-spd="1.5" style="padding:4px 8px;font-size:11px;">1.5x</button>' +
					'              <button class="wa-audio-btn wa-btn wa-btn-secondary" data-spd="2.0" style="padding:4px 8px;font-size:11px;">2.0x</button>' +
					'            </div>' +
					'          </div>' +
					'        </div>' +
					'      </div>' +
					'' +
					'      <!-- TAB 3: ALAT & CHAT -->' +
					'      <div id="wa-tab-content-productivity" class="wa-tab-pane" style="display:none;flex-direction:column;gap:14px;">' +
					'        <div style="display:grid;grid-template-columns:1fr 1fr;gap:10px;">' +
					'          <div class="wa-card" style="justify-content:space-between;">' +
					'            <div>' +
					'              <div class="wa-row-title" style="display:flex;align-items:center;gap:6px;">' +
					'                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>' +
					'                <span>Catatan Tempel</span>' +
					'              </div>' +
					'              <div class="wa-row-desc" style="margin-top:4px;">Panel memo samping dengan autosave otomatis (Alt + N)</div>' +
					'            </div>' +
					'            <button id="wa-cc-scratchpad-btn" class="wa-btn wa-btn-secondary" style="margin-top:10px;width:100%;">Buka Catatan</button>' +
					'          </div>' +
					'          <div class="wa-card" style="justify-content:space-between;">' +
					'            <div>' +
					'              <div class="wa-row-title" style="display:flex;align-items:center;gap:6px;">' +
					'                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>' +
					'                <span>Screenshot Anonim</span>' +
					'              </div>' +
					'              <div class="wa-row-desc" style="margin-top:4px;">Samarkan nama dan avatar sebelum mengambil tangkapan layar</div>' +
					'            </div>' +
					'            <button id="wa-cc-anon-btn" class="wa-btn wa-btn-secondary" style="margin-top:10px;width:100%;">Aktifkan Mode</button>' +
					'          </div>' +
					'        </div>' +
					'' +
					'        <div class="wa-card">' +
					'          <div class="wa-card-header">' +
					'            <div class="wa-card-title">' +
					'              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"/></svg>' +
					'              <span>Chat ke Nomor Baru (Ctrl + N)</span>' +
					'            </div>' +
					'          </div>' +
					'          <div style="display:flex;gap:8px;">' +
					'            <input id="wa-cc-phone" class="wa-input" type="text" placeholder="Contoh: 08123456789 atau 628..." style="flex:1;">' +
					'            <button id="wa-cc-phone-btn" class="wa-btn wa-btn-primary">Buka Chat</button>' +
					'          </div>' +
					'        </div>' +
					'' +
					'        <div class="wa-card">' +
					'          <div class="wa-card-header">' +
					'            <div>' +
					'              <div class="wa-card-title">' +
					'                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>' +
					'                <span>Template Balasan Cepat (Quick Replies)</span>' +
					'              </div>' +
					'              <div class="wa-row-desc" style="margin-top:2px;">Ketik /shortcut lalu tekan Tab, Enter, Spasi, atau klik popup otomatis</div>' +
					'            </div>' +
					'          </div>' +
					'          <div id="wa-qr-list" style="max-height:120px;overflow-y:auto;display:flex;flex-direction:column;gap:6px;"></div>' +
					'          <div style="display:flex;gap:6px;border-top:1px solid var(--wa-border);padding-top:10px;">' +
					'            <input id="wa-qr-new-key" class="wa-input" type="text" placeholder="/shortcut" style="width:110px;">' +
					'            <input id="wa-qr-new-text" class="wa-input" type="text" placeholder="Teks balasan otomatis..." style="flex:1;">' +
					'            <button id="wa-qr-add-btn" class="wa-btn wa-btn-primary">+ Tambah</button>' +
					'          </div>' +
					'        </div>' +
					'      </div>' +
					'' +
					'      <!-- TAB 4: SISTEM & KEAMANAN -->' +
					'      <div id="wa-tab-content-system" class="wa-tab-pane" style="display:none;flex-direction:column;gap:14px;">' +
					'        <div class="wa-card">' +
					'          <div class="wa-card-header">' +
					'            <div>' +
					'              <div class="wa-card-title">' +
					'                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>' +
					'                <span>Kunci Aplikasi (Ctrl + L)</span>' +
					'              </div>' +
					'              <div class="wa-row-desc" style="margin-top:2px;">Kunci otomatis setelah 5 menit tidak aktif. PIN default: 1234</div>' +
					'            </div>' +
					'          </div>' +
					'          <div style="display:flex;gap:10px;">' +
					'            <button id="wa-cc-lock-btn" class="wa-btn wa-btn-secondary" style="flex:1;">Kunci Sekarang</button>' +
					'            <button id="wa-cc-changepin-btn" class="wa-btn wa-btn-primary" style="flex:1;">Ubah / Ganti PIN</button>' +
					'          </div>' +
					'        </div>' +
					'' +
					'        <div class="wa-card">' +
					'          <div class="wa-card-header">' +
					'            <div>' +
					'              <div class="wa-card-title">' +
					'                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="12" y1="3" x2="12" y2="21"/></svg>' +
					'                <span>Multi-Messenger & Mode Berdampingan</span>' +
					'              </div>' +
					'              <div class="wa-row-desc" style="margin-top:2px;">Jalankan dua akun WhatsApp atau buka WhatsApp & Telegram berdampingan 50:50</div>' +
					'            </div>' +
					'          </div>' +
					'          <div style="display:grid;grid-template-columns:1fr 1fr;gap:8px;">' +
					'            <button id="wa-cc-dual-acc-btn" class="wa-btn wa-btn-secondary">Buka Akun WhatsApp Ke-2</button>' +
					'            <button id="wa-cc-filter-unread-btn" class="wa-btn wa-btn-secondary">Filter Belum Dibaca</button>' +
					'            <button id="wa-cc-tg-btn" class="wa-btn wa-btn-tg">Buka Telegram Web (Ctrl+2)</button>' +
					'            <button id="wa-cc-split-btn" class="wa-btn wa-btn-secondary" style="border-color:rgba(36,161,222,0.4);color:var(--wa-tg);">Berdampingan 50:50 (Ctrl+3)</button>' +
					'          </div>' +
					'        </div>' +
					'' +
					'        <div class="wa-card">' +
					'          <div class="wa-card-header">' +
					'            <div class="wa-card-title">' +
					'              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>' +
					'              <span>Startup & Pemeliharaan</span>' +
					'            </div>' +
					'          </div>' +
					'          <div class="wa-row">' +
					'            <div>' +
					'              <div class="wa-row-title">Mulai Otomatis saat Boot Windows</div>' +
					'              <div class="wa-row-desc">Aplikasi otomatis berjalan di System Tray saat komputer dinyalakan</div>' +
					'            </div>' +
					'            <label class="wa-switch"><input type="checkbox" id="wa-cc-autostart-cb"><span class="wa-slider"></span></label>' +
					'          </div>' +
					'          <div style="font-size:11.5px;color:var(--wa-text-muted);background:var(--wa-bg);padding:10px 14px;border-radius:var(--wa-radius-sm);border:1px solid var(--wa-border);line-height:1.4;">' +
					'            Pintasan Boss Key Global: Tekan <b style="color:var(--wa-text);">Ctrl + Alt + W</b> kapan saja untuk menyembunyikan atau memunculkan jendela seketika.' +
					'          </div>' +
					'          <div style="display:flex;gap:10px;border-top:1px solid var(--wa-border);padding-top:12px;">' +
					'            <button id="wa-cc-reload-btn" class="wa-btn wa-btn-secondary" style="flex:1;">Muat Ulang Halaman</button>' +
					'            <button id="wa-cc-logout-btn" class="wa-btn wa-btn-danger" style="flex:1;">Reset Sesi Login</button>' +
					'          </div>' +
					'        </div>' +
					'      </div>' +
					'    </div>' +
					'  </div>';

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
						if (window.setAppAutoStart) {
							window.setAppAutoStart(autoCb.checked);
							showAddonToast(autoCb.checked ? 'Mulai Otomatis: Aktif' : 'Mulai Otomatis: Nonaktif');
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
					logBtn.onclick = function() {
						if (confirm('Apakah Anda yakin ingin menghapus data sesi login di PC ini? Anda perlu scan QR ulang untuk masuk kembali.')) {
							try {
								localStorage.clear();
								sessionStorage.clear();
							} catch(e) {}
							window.location.href = 'https://web.whatsapp.com';
						}
					};
				}

				modal._syncUI = syncPrivacyUI;
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
					var navContainers = document.querySelectorAll('header, nav, [role="navigation"], [aria-label*="Navigation"], [aria-label*="navigation"]');
					for (var i = 0; i < navContainers.length; i++) {
						var container = navContainers[i];
						var buttons = container.querySelectorAll('[role="button"], button');
						if (buttons.length >= 3) {
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
	`

	w.Init(initScript)

	// Initialize Telegram child webview before navigation begins
	initTelegramChild()

	w.Navigate(appURL)
	w.Run()
}
