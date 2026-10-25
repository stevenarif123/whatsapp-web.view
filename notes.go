package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/jchv/go-webview2"
)

// Shared data: notes (Alt+N) and quick replies are stored here instead of in each page's localStorage,
// so the WhatsApp and Telegram windows (and a second WhatsApp profile) use the same data. The pages own
// the format; Go only checks that it is JSON of a sane size.

const maxSharedSize = 4 << 20

// sharedFiles maps the names pages may use to their files in the config dir.
var sharedFiles = map[string]string{
	"notes":        "notes.json",
	"quickReplies": "quick_replies.json",
}

var sharedMu sync.Mutex

func sharedPath(name string) string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.Getenv("APPDATA")
		if configDir == "" {
			configDir = "."
		}
	}
	base := filepath.Join(configDir, "WhatsAppDesktopLight")
	_ = os.MkdirAll(base, 0755)
	return filepath.Join(base, sharedFiles[name])
}

// getShared returns the stored JSON for name, or "" when there is none yet.
func getShared(name string) string {
	if _, ok := sharedFiles[name]; !ok {
		return ""
	}
	sharedMu.Lock()
	defer sharedMu.Unlock()
	data, err := os.ReadFile(sharedPath(name))
	if err != nil || !json.Valid(data) {
		return ""
	}
	return string(data)
}

func saveShared(name, raw string) bool {
	if _, ok := sharedFiles[name]; !ok || len(raw) > maxSharedSize || !json.Valid([]byte(raw)) {
		return false
	}
	sharedMu.Lock()
	defer sharedMu.Unlock()
	p := sharedPath(name)
	tmp := p + ".tmp"
	if os.WriteFile(tmp, []byte(raw), 0600) != nil {
		return false
	}
	return os.Rename(tmp, p) == nil
}

func getSharedNotes() string                 { return getShared("notes") }
func saveSharedNotes(raw string) bool        { return saveShared("notes", raw) }
func getSharedQuickReplies() string          { return getShared("quickReplies") }
func saveSharedQuickReplies(raw string) bool { return saveShared("quickReplies", raw) }

func bindNotes(w webview2.WebView) {
	_ = w.Bind("getSharedNotes", getSharedNotes)
	_ = w.Bind("saveSharedNotes", saveSharedNotes)
	_ = w.Bind("getSharedQuickReplies", getSharedQuickReplies)
	_ = w.Bind("saveSharedQuickReplies", saveSharedQuickReplies)
}
