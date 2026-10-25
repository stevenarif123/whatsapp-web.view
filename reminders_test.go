package main

import (
	"strings"
	"testing"
	"time"
)

func TestReminders(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	now := time.Date(2026, 10, 26, 10, 0, 0, 0, time.Local)
	old := currentProfileID
	t.Cleanup(func() { currentProfileID = old })
	currentProfileID = "1"

	if _, err := addReminder("", "x", now.Add(time.Hour).Unix(), now); err == nil {
		t.Fatal("reminder without chat accepted")
	}
	if _, err := addReminder("Budi", "", now.Add(-time.Minute).Unix(), now); err == nil {
		t.Fatal("reminder in the past accepted")
	}
	if _, err := addReminder("Budi", "", now.Add(400*24*time.Hour).Unix(), now); err == nil {
		t.Fatal("reminder more than a year ahead accepted")
	}

	list, err := addReminder("Budi", "kirim invoice", now.Add(2*time.Hour).Unix(), now)
	if err != nil {
		t.Fatal(err)
	}
	list, _ = addReminder("Siti", strings.Repeat("a", 500), now.Add(time.Hour).Unix(), now)
	if len(list) != 2 || list[0].Chat != "Siti" || len([]rune(list[0].Note)) != 300 {
		t.Fatalf("expected Siti first with a trimmed note: %+v", list)
	}

	// Another profile's reminders are kept apart.
	currentProfileID = "2"
	addReminder("Akun2", "", now.Add(time.Hour).Unix(), now)
	if l := listReminders(); len(l) != 1 || l[0].Chat != "Akun2" {
		t.Fatalf("profile 2 sees %+v", l)
	}
	currentProfileID = "1"

	if due := takeDueReminders(now.Add(30 * time.Minute)); len(due) != 0 {
		t.Fatal("nothing should be due yet")
	}
	due := takeDueReminders(now.Add(90 * time.Minute))
	if len(due) != 1 || due[0].Chat != "Siti" {
		t.Fatalf("due: %+v", due)
	}
	if l := listReminders(); len(l) != 1 || l[0].Chat != "Budi" {
		t.Fatalf("fired reminder not removed: %+v", l)
	}

	title, msg := reminderToast(list[1], time.Unix(list[1].Due, 0).Add(time.Minute))
	if !strings.Contains(title, "balas Budi") || strings.Contains(title, "terlewat") || msg != "kirim invoice" {
		t.Fatalf("on-time toast: %q / %q", title, msg)
	}
	title, _ = reminderToast(list[1], time.Unix(list[1].Due, 0).Add(3*time.Hour))
	if !strings.Contains(title, "terlewat") {
		t.Fatalf("missed reminder not marked: %q", title)
	}

	if l := deleteReminder(list[1].ID); len(l) != 0 {
		t.Fatalf("delete failed: %+v", l)
	}
}

func TestActivationChatURI(t *testing.T) {
	uri := activationChatURI("Ibu & Ayah")
	if !strings.HasPrefix(uri, notifyProtocol+":chat?") || !strings.Contains(uri, "chat=Ibu+%26+Ayah") {
		t.Fatalf("bad uri %s", uri)
	}
}
