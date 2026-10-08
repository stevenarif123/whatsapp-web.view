package main

import (
	_ "embed"
	"encoding/json"
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

//go:embed assets/addon.css
var addonCSS string

//go:embed assets/control-center.html
var controlCenterHTML string

//go:embed assets/addon-core.js
var addonCoreJS string

//go:embed assets/addon-main.js
var addonMainJS string

//go:embed assets/telegram-main.js
var telegramMainJS string

// jsString returns s as a quoted JavaScript string literal.
func jsString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

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
		applyMessengerLayout(true)
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

// notifyHidden tells the page the window was hidden, so it can lock itself if the user enabled that.
func notifyHidden() {
	if globalWebView != nil {
		globalWebView.Eval("window.waOnHidden && window.waOnHidden()")
	}
}

func customWndProc(hWnd, msg, wParam, lParam uintptr) uintptr {
	switch uint32(msg) {
	case WM_CLOSE:
		if !forceQuit {
			// Minimize to tray instead of quitting
			procShowWindow.Call(hWnd, SW_HIDE)
			notifyHidden()
			return 0
		}
	case WM_SIZE:
		r, _, _ := procCallWindowProc.Call(origWndProc, hWnd, msg, wParam, lParam)
		applyMessengerLayout(false)
		return r
	case WM_HOTKEY:
		if wParam == HOTKEY_BOSS {
			fg, _, _ := procGetFgWindow.Call()
			visible, _, _ := procIsWindowVisible.Call(hWnd)
			if visible != 0 && fg == hWnd {
				procShowWindow.Call(hWnd, SW_HIDE)
				notifyHidden()
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
	setTaskbarBadge(count)
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

// applyMessengerLayout positions the WhatsApp content and the Telegram window for the active mode.
// refocus is true only for explicit mode switches; plain resizes must not steal keyboard focus.
func applyMessengerLayout(refocus bool) {
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

	// The WebView2 viewport must be resized through its controller bounds. Moving Chrome's internal
	// child window instead leaves the page laid out at the old size, so content gets cut off.
	switch activeMessengerMode {
	case "tg":
		if globalWebView != nil {
			globalWebView.ClearContentBounds()
		}
		if globalTelegramHwnd != 0 {
			procMoveWindow.Call(globalTelegramHwnd, 0, 0, uintptr(totalW), uintptr(totalH), 1)
			procShowWindow.Call(globalTelegramHwnd, SW_SHOW)
			if refocus {
				procSetFocus.Call(globalTelegramHwnd)
			}
		}
	case "split":
		halfW := totalW / 2
		if globalWebView != nil {
			globalWebView.SetContentBounds(0, 0, int32(halfW), int32(totalH))
		}
		if globalTelegramHwnd != 0 {
			procMoveWindow.Call(globalTelegramHwnd, uintptr(halfW), 0, uintptr(totalW-halfW), uintptr(totalH), 1)
			procShowWindow.Call(globalTelegramHwnd, SW_SHOW)
		}
	default: // "wa"
		if globalTelegramHwnd != 0 {
			procShowWindow.Call(globalTelegramHwnd, SW_HIDE)
		}
		if globalWebView != nil {
			globalWebView.ClearContentBounds()
		}
		if refocus {
			procSetFocus.Call(globalHwnd)
		}
	}
}

func switchToWhatsApp() {
	activeMessengerMode = "wa"
	applyMessengerLayout(true)
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
	applyMessengerLayout(true)
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
	applyMessengerLayout(true)
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
				notifyHidden()
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

var telegramInitScript = strings.ReplaceAll(telegramMainJS, "__WA_USER_AGENT__", userAgent)


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
	initScript := "var WA_ADDON_CSS = " + jsString(addonCSS) + ";\nvar WA_CC_HTML = " + jsString(controlCenterHTML) + ";\n" +
		strings.ReplaceAll(addonMainJS, "__WA_USER_AGENT__", userAgent)

	w.Init(initScript + "\n" + addonCoreJS)

	// Initialize Telegram child webview before navigation begins
	initTelegramChild()

	w.Navigate(appURL)
	w.Run()
}
