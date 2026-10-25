package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/jchv/go-webview2"
)

// Notes (Alt+N) are stored here instead of in each page's localStorage, so the WhatsApp and Telegram
// windows (and a second WhatsApp profile) show the same notes. The page owns the format; Go only checks
// that it is JSON of a sane size.

const maxNotesSize = 4 << 20

var notesMu sync.Mutex

func notesPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.Getenv("APPDATA")
		if configDir == "" {
			configDir = "."
		}
	}
	base := filepath.Join(configDir, "WhatsAppDesktopLight")
	_ = os.MkdirAll(base, 0755)
	return filepath.Join(base, "notes.json")
}

// getSharedNotes returns the stored notes JSON, or "" when there are none yet.
func getSharedNotes() string {
	notesMu.Lock()
	defer notesMu.Unlock()
	data, err := os.ReadFile(notesPath())
	if err != nil || !json.Valid(data) {
		return ""
	}
	return string(data)
}

func saveSharedNotes(raw string) bool {
	if len(raw) > maxNotesSize || !json.Valid([]byte(raw)) {
		return false
	}
	notesMu.Lock()
	defer notesMu.Unlock()
	p := notesPath()
	tmp := p + ".tmp"
	if os.WriteFile(tmp, []byte(raw), 0600) != nil {
		return false
	}
	return os.Rename(tmp, p) == nil
}

func bindNotes(w webview2.WebView) {
	_ = w.Bind("getSharedNotes", getSharedNotes)
	_ = w.Bind("saveSharedNotes", saveSharedNotes)
}
