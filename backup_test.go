package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestBackupRoundTrip(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	saveShared("notes", `{"list":[{"id":"n1","text":"resi 123"}],"active":"n1"}`)
	saveShared("quickReplies", `[{"key":"/toko","text":"Buka 08-21"}]`)
	muteChat("Grup Kantor")
	changeAppPin("", "482915") // must never end up in a backup

	page := map[string]string{
		"wa_pinned_chats":   `["Budi"]`,
		"wa_privacy_config": `{"active":true}`,
		"wa_is_locked":      "true", // lock state is not a setting
		"WACachedPage":      "x",    // WhatsApp's own data
	}
	data, err := buildBackup(page, time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, bad := range []string{"wa_is_locked", "WACachedPage", "482915", "app_pin"} {
		if strings.Contains(text, bad) {
			t.Fatalf("backup contains %q", bad)
		}
	}

	// Restore into an empty profile.
	t.Setenv("APPDATA", t.TempDir())
	restored, err := applyBackup(data)
	if err != nil {
		t.Fatal(err)
	}
	if restored["wa_pinned_chats"] != `["Budi"]` || len(restored) != 2 {
		t.Fatalf("page entries: %v", restored)
	}
	if !strings.Contains(getShared("quickReplies"), "/toko") || !strings.Contains(getShared("notes"), "resi 123") {
		t.Fatal("shared data not restored")
	}
	if !isMuted(getNotifySettings(), "Grup Kantor") {
		t.Fatal("notification settings not restored")
	}
	if getAppPinStatus().IsDefault != true {
		t.Fatal("restoring settings must not touch the PIN")
	}
}

func TestBackupRejectsForeignFiles(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	for _, body := range []string{`not json`, `{"app":"Other","version":1}`, `{"app":"WhatsAppDesktopLight","version":99}`} {
		if _, err := applyBackup([]byte(body)); err == nil {
			t.Errorf("accepted %s", body)
		}
	}
	// Unknown page keys and shared names are dropped on import.
	b, _ := json.Marshal(settingsBackup{App: backupApp, Version: 1,
		Page:   map[string]string{"wa_is_locked": "false", "wa_zoom_level": "110"},
		Shared: map[string]json.RawMessage{"secrets": json.RawMessage(`{}`)}})
	page, err := applyBackup(b)
	if err != nil || len(page) != 1 || page["wa_zoom_level"] != "110" {
		t.Fatalf("page=%v err=%v", page, err)
	}
}
