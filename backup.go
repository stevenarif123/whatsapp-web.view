package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"

	"github.com/jchv/go-webview2"
	"golang.org/x/sys/windows"
)

// Settings backup: one JSON file with the add-on settings of the WhatsApp page, the shared notes and
// quick replies, and the notification settings. The PIN, the lock state and the login session are never
// included. The files are chosen with the Windows Save/Open dialogs, shown from the dispatch queue so
// they do not run inside a WebView2 event handler; results go back to the page via window.waSettingsResult.

const (
	backupApp     = "WhatsAppDesktopLight"
	backupVersion = 1
	maxBackupSize = 16 << 20
)

// backupPageKeys are the WhatsApp page's localStorage entries that belong to the add-on.
var backupPageKeys = []string{
	"wa_privacy_config", "wa_eye_comfort_config", "wa_zoom_level", "wa_audio_speed",
	"wa_pins_enabled", "wa_pinned_chats", "wa_blur_chats", "wa_radar_config",
	"wa_compact_narrow", "wa_lock_on_hide",
	"wa_label_defs", "wa_chat_labels", "wa_contact_notes", "wa_followup_cfg",
}

var procGetOpenFile = comdlg32.NewProc("GetOpenFileNameW")

const OFN_FILEMUSTEXIST = 0x00001000

type settingsBackup struct {
	App        string                     `json:"app"`
	Version    int                        `json:"version"`
	AppVersion string                     `json:"appVersion"`
	Created    string                     `json:"created"`
	Page       map[string]string          `json:"page"`
	Shared     map[string]json.RawMessage `json:"shared"`
	Notify     *notifySettings            `json:"notify,omitempty"`
}

func isBackupPageKey(k string) bool {
	for _, x := range backupPageKeys {
		if x == k {
			return true
		}
	}
	return false
}

// buildBackup collects everything into one file body. page holds the page's localStorage entries.
func buildBackup(page map[string]string, now time.Time) ([]byte, error) {
	b := settingsBackup{
		App:        backupApp,
		Version:    backupVersion,
		AppVersion: appVersion,
		Created:    now.Format(time.RFC3339),
		Page:       map[string]string{},
		Shared:     map[string]json.RawMessage{},
	}
	for k, v := range page {
		if isBackupPageKey(k) {
			b.Page[k] = v
		}
	}
	for name := range sharedFiles {
		if raw := getShared(name); raw != "" {
			b.Shared[name] = json.RawMessage(raw)
		}
	}
	ns := getNotifySettings()
	b.Notify = &ns
	return json.MarshalIndent(b, "", "  ")
}

// applyBackup restores the Go-side parts of a backup and returns the page entries to restore.
func applyBackup(data []byte) (map[string]string, error) {
	var b settingsBackup
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("bukan file cadangan yang valid")
	}
	if b.App != backupApp || b.Version < 1 || b.Version > backupVersion {
		return nil, fmt.Errorf("bukan file cadangan aplikasi ini")
	}
	for name, raw := range b.Shared {
		if _, ok := sharedFiles[name]; ok {
			saveShared(name, string(raw))
		}
	}
	if b.Notify != nil {
		setNotifySettings(*b.Notify)
	}
	page := map[string]string{}
	for k, v := range b.Page {
		if isBackupPageKey(k) {
			page[k] = v
		}
	}
	return page, nil
}

func pickOpenPath(owner uintptr) string {
	fileBuffer := make([]uint16, windows.MAX_PATH*4)
	filter, _ := syscall.UTF16PtrFromString("Cadangan pengaturan (*.json)\x00*.json\x00Semua file\x00*.*\x00\x00")
	title, _ := syscall.UTF16PtrFromString("Pilih File Cadangan Pengaturan")
	name := openFileName{
		StructSize:  uint32(unsafe.Sizeof(openFileName{})),
		Owner:       owner,
		Filter:      filter,
		FilterIndex: 1,
		File:        &fileBuffer[0],
		MaxFile:     uint32(len(fileBuffer)),
		Title:       title,
		Flags:       OFN_FILEMUSTEXIST | OFN_PATHMUSTEXIST | OFN_NOCHANGEDIR,
	}
	if result, _, _ := procGetOpenFile.Call(uintptr(unsafe.Pointer(&name))); result == 0 {
		return ""
	}
	return windows.UTF16ToString(fileBuffer)
}

func defaultBackupPath(now time.Time) string {
	dir, err := os.UserHomeDir()
	if err == nil {
		dir = filepath.Join(dir, "Documents")
	}
	return filepath.Join(dir, "WhatsAppDesktop-pengaturan-"+now.Format("2006-01-02")+".json")
}

// bindBackup exposes export and import to the WhatsApp page.
func bindBackup(w webview2.WebView, owner uintptr) {
	result := func(v interface{}) {
		data, _ := json.Marshal(v)
		w.Eval(fmt.Sprintf("window.waSettingsResult && window.waSettingsResult(%s)", data))
	}

	_ = w.Bind("getBackupKeys", func() []string { return backupPageKeys })

	_ = w.Bind("exportSettings", func(page map[string]string) {
		w.Dispatch(func() {
			now := time.Now()
			data, err := buildBackup(page, now)
			if err != nil {
				result(map[string]interface{}{"action": "export", "ok": false, "error": err.Error()})
				return
			}
			path := pickSavePath(owner, defaultBackupPath(now))
			if path == "" {
				result(map[string]interface{}{"action": "export", "ok": false, "cancelled": true})
				return
			}
			if err := os.WriteFile(path, data, 0600); err != nil {
				result(map[string]interface{}{"action": "export", "ok": false, "error": err.Error()})
				return
			}
			result(map[string]interface{}{"action": "export", "ok": true, "name": filepath.Base(path)})
		})
	})

	_ = w.Bind("importSettings", func() {
		w.Dispatch(func() {
			path := pickOpenPath(owner)
			if path == "" {
				result(map[string]interface{}{"action": "import", "ok": false, "cancelled": true})
				return
			}
			st, err := os.Stat(path)
			if err != nil || st.Size() > maxBackupSize {
				result(map[string]interface{}{"action": "import", "ok": false, "error": "file tidak bisa dibaca atau terlalu besar"})
				return
			}
			data, err := os.ReadFile(path)
			if err == nil {
				var page map[string]string
				if page, err = applyBackup(data); err == nil {
					result(map[string]interface{}{"action": "import", "ok": true, "page": page})
					return
				}
			}
			result(map[string]interface{}{"action": "import", "ok": false, "error": err.Error()})
		})
	})
}
