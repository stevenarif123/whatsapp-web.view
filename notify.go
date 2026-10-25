package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/go-toast/toast"
	"github.com/jchv/go-webview2"
	"golang.org/x/sys/windows"
)

// Notification extras: clicking a toast opens its chat, quiet hours (do not disturb) and muted chats.
//
// Toasts are shown through PowerShell (go-toast) and cannot call back into this process. Instead they
// use the app's own URL protocol: Windows starts the exe with "--activate <uri>", which hands the URI to
// the running window with WM_COPYDATA (forwardActivation) and exits. The window then acts on it
// (handleActivation). The settings are shared by WhatsApp and Telegram.

const (
	notifyProtocol = "wadesktoplight"
	copyDataMagic  = 0x57414131 // "WAA1", marks our WM_COPYDATA messages
	WM_COPYDATA    = 0x004A
	ASFW_ANY       = 0xFFFFFFFF
)

var (
	procRegCreateKeyEx          = advapi32.NewProc("RegCreateKeyExW")
	procAllowSetForegroundWindow = user32.NewProc("AllowSetForegroundWindow")
)

type copyDataStruct struct {
	Magic uintptr
	Size  uint32
	Data  *uint16
}

type notifySettings struct {
	DNDEnabled bool     `json:"dndEnabled"`
	DNDStart   string   `json:"dndStart"` // "HH:MM"
	DNDEnd     string   `json:"dndEnd"`
	Muted      []string `json:"muted"`
}

var notifyMu sync.Mutex

func notifySettingsPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.Getenv("APPDATA")
		if configDir == "" {
			configDir = "."
		}
	}
	base := filepath.Join(configDir, "WhatsAppDesktopLight")
	_ = os.MkdirAll(base, 0755)
	return filepath.Join(base, "notify_settings.json")
}

func loadNotifySettings() notifySettings {
	s := notifySettings{DNDStart: "22:00", DNDEnd: "07:00", Muted: []string{}}
	if data, err := os.ReadFile(notifySettingsPath()); err == nil {
		_ = json.Unmarshal(data, &s)
	}
	if s.Muted == nil {
		s.Muted = []string{}
	}
	return s
}

func saveNotifySettings(s notifySettings) {
	clean := []string{}
	seen := map[string]bool{}
	for _, name := range s.Muted {
		name = strings.TrimSpace(name)
		if name != "" && !seen[strings.ToLower(name)] {
			seen[strings.ToLower(name)] = true
			clean = append(clean, name)
		}
	}
	s.Muted = clean
	data, err := json.Marshal(s)
	if err != nil {
		return
	}
	_ = os.WriteFile(notifySettingsPath(), data, 0600)
}

func getNotifySettings() notifySettings {
	notifyMu.Lock()
	defer notifyMu.Unlock()
	return loadNotifySettings()
}

func setNotifySettings(s notifySettings) notifySettings {
	notifyMu.Lock()
	defer notifyMu.Unlock()
	saveNotifySettings(s)
	return loadNotifySettings()
}

func muteChat(name string) {
	notifyMu.Lock()
	defer notifyMu.Unlock()
	s := loadNotifySettings()
	s.Muted = append(s.Muted, name)
	saveNotifySettings(s)
}

// parseClock turns "HH:MM" into minutes after midnight.
func parseClock(s string) (int, bool) {
	t, err := time.Parse("15:04", strings.TrimSpace(s))
	if err != nil {
		return 0, false
	}
	return t.Hour()*60 + t.Minute(), true
}

func inQuietHours(s notifySettings, now time.Time) bool {
	if !s.DNDEnabled {
		return false
	}
	start, ok1 := parseClock(s.DNDStart)
	end, ok2 := parseClock(s.DNDEnd)
	if !ok1 || !ok2 || start == end {
		return false
	}
	m := now.Hour()*60 + now.Minute()
	if start < end {
		return m >= start && m < end
	}
	return m >= start || m < end // the quiet period crosses midnight
}

func isMuted(s notifySettings, title string) bool {
	title = strings.TrimSpace(title)
	for _, name := range s.Muted {
		if strings.EqualFold(name, title) {
			return true
		}
	}
	return false
}

// activationURI builds the URI a toast opens. app is "wa" or "tg"; id is the page's notification ID.
func activationURI(action, app string, id int, chat string) string {
	q := url.Values{}
	q.Set("app", app)
	q.Set("profile", currentProfileID)
	q.Set("id", strconv.Itoa(id))
	q.Set("chat", chat)
	return notifyProtocol + ":" + action + "?" + q.Encode()
}

// showNativeNotification shows a Windows toast unless quiet hours are on or the chat is muted. id > 0
// links the toast to a notification object in the page so a click can open its chat.
func showNativeNotification(app, appID, title, message, iconPath string, id int) {
	s := getNotifySettings()
	if inQuietHours(s, time.Now()) || isMuted(s, title) {
		return
	}
	title = sanitizeNotificationText(title, 128)
	message = sanitizeNotificationText(message, 512)
	if title == "" && message == "" {
		return
	}

	notification := toast.Notification{
		AppID:               appID,
		Title:               title,
		Message:             message,
		Icon:                iconPath,
		ActivationType:      "protocol",
		ActivationArguments: activationURI("open", app, id, title),
	}
	if id > 0 && title != "" {
		notification.Actions = []toast.Action{
			{Type: "protocol", Label: "Bisukan chat ini", Arguments: activationURI("mute", app, id, title)},
		}
	}
	_ = notification.Push()
}

// bindNotifications exposes the toast bridge and the notification settings to a page.
func bindNotifications(w webview2.WebView, app, appID, iconPath string) {
	_ = w.Bind("sendNativeNotification", func(title, body string) {
		go showNativeNotification(app, appID, title, body, iconPath, 0)
	})
	_ = w.Bind("sendNativeNotificationEx", func(title, body string, id int) {
		go showNativeNotification(app, appID, title, body, iconPath, id)
	})
	_ = w.Bind("getNotifySettings", getNotifySettings)
	_ = w.Bind("setNotifySettings", setNotifySettings)
}

// registerNotifyProtocol points the wadesktoplight: URL protocol at this exe (per user, no admin).
func registerNotifyProtocol() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}
	setRegString := func(subKey, name, value string) {
		var hKey windows.Handle
		subPtr, _ := syscall.UTF16PtrFromString(subKey)
		r, _, _ := procRegCreateKeyEx.Call(HKEY_CURRENT_USER, uintptr(unsafe.Pointer(subPtr)), 0, 0, 0, KEY_WRITE, 0, uintptr(unsafe.Pointer(&hKey)), 0)
		if r != 0 {
			return
		}
		defer windows.RegCloseKey(hKey)
		var namePtr *uint16
		if name != "" {
			namePtr, _ = syscall.UTF16PtrFromString(name)
		}
		val, _ := syscall.UTF16FromString(value)
		procRegSetValueEx.Call(uintptr(hKey), uintptr(unsafe.Pointer(namePtr)), 0, REG_SZ, uintptr(unsafe.Pointer(&val[0])), uintptr(len(val)*2))
	}
	base := `Software\Classes\` + notifyProtocol
	setRegString(base, "", "URL:WhatsApp Desktop Light")
	setRegString(base, "URL Protocol", "")
	setRegString(base+`\shell\open\command`, "", fmt.Sprintf(`"%s" --activate "%%1"`, exePath))
}

// forwardActivation hands a toast URI to the running app window and reports whether one was found.
// It runs in the short-lived process Windows starts for the click.
func forwardActivation(uri string) bool {
	u, err := url.Parse(uri)
	if err != nil || u.Scheme != notifyProtocol {
		return false
	}
	q := u.Query()
	titles := []string{windowTitle}
	if p := q.Get("profile"); p != "" && p != "1" {
		titles = []string{fmt.Sprintf("WhatsApp Desktop (Akun %s)", p)}
	}
	if q.Get("app") == "tg" {
		// A standalone Telegram window first, else the WhatsApp window that embeds Telegram.
		titles = append([]string{telegramWindowTitle}, titles...)
	}
	for _, title := range titles {
		titlePtr, _ := syscall.UTF16PtrFromString(title)
		hwnd, _, _ := procFindWindow.Call(0, uintptr(unsafe.Pointer(titlePtr)))
		if hwnd == 0 {
			continue
		}
		data, _ := syscall.UTF16FromString(uri)
		cds := copyDataStruct{Magic: copyDataMagic, Size: uint32(len(data) * 2), Data: &data[0]}
		// Let the app window take the foreground; this process got that right from the click.
		procAllowSetForegroundWindow.Call(ASFW_ANY)
		procSendMessage.Call(hwnd, WM_COPYDATA, 0, uintptr(unsafe.Pointer(&cds)))
		return true
	}
	return false
}

// readCopyData returns the URI carried by one of our WM_COPYDATA messages, or "".
func readCopyData(lParam uintptr) string {
	cds := *(**copyDataStruct)(unsafe.Pointer(&lParam))
	if cds == nil || cds.Magic != copyDataMagic || cds.Data == nil || cds.Size < 2 || cds.Size > 64*1024 {
		return ""
	}
	return windows.UTF16ToString(unsafe.Slice(cds.Data, cds.Size/2))
}

// handleActivation acts on a toast URI inside the running app (UI thread).
func handleActivation(uri string) {
	u, err := url.Parse(uri)
	if err != nil || u.Scheme != notifyProtocol {
		return
	}
	q := u.Query()
	id, _ := strconv.Atoi(q.Get("id"))
	chat := q.Get("chat")
	isTG := q.Get("app") == "tg"

	target := globalWebView
	if isTG && globalTelegramWebView != nil {
		target = globalTelegramWebView
	}

	switch u.Opaque {
	case "open":
		if currentWindowTitle == telegramWindowTitle {
			// Standalone Telegram process: its own window is the main one.
			showMainWindow()
		} else if isTG {
			switchToTelegram()
		} else {
			switchToWhatsApp()
		}
		procSetFgWindow.Call(globalHwnd)
		if target != nil && id > 0 {
			target.Eval(fmt.Sprintf("window.waNotificationClicked && window.waNotificationClicked(%d, %s)", id, jsString(chat)))
		}
	case "mute":
		if chat == "" {
			return
		}
		muteChat(chat)
		if globalWebView != nil {
			globalWebView.Eval(fmt.Sprintf("window.waOnChatMuted && window.waOnChatMuted(%s)", jsString(chat)))
		}
	}
}
