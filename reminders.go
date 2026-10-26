package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/go-toast/toast"
	"github.com/jchv/go-webview2"
)

// Reply reminders: "remind me to answer this chat" at a chosen time. They are kept in reminders.json and
// shown as a toast that opens the chat when clicked. Quiet hours and muted chats do not apply, since the
// user asked for the reminder. Everything stays on this computer; nothing is sent to WhatsApp.

type reminder struct {
	ID      string `json:"id"`
	Profile string `json:"profile"`
	Chat    string `json:"chat"`
	Note    string `json:"note"`
	Due     int64  `json:"due"` // unix seconds
}

const (
	reminderTick     = 20 * time.Second
	reminderMaxAhead = 366 * 24 * time.Hour
	reminderLateness = 2 * time.Minute // fired later than this counts as missed (app was closed)
	maxReminders     = 200
)

var remindersMu sync.Mutex

func remindersPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.Getenv("APPDATA")
		if configDir == "" {
			configDir = "."
		}
	}
	base := filepath.Join(configDir, "WhatsAppDesktopLight")
	_ = os.MkdirAll(base, 0755)
	return filepath.Join(base, "reminders.json")
}

func loadReminders() []reminder {
	var list []reminder
	if data, err := os.ReadFile(remindersPath()); err == nil {
		_ = json.Unmarshal(data, &list)
	}
	return list
}

func saveReminders(list []reminder) {
	if data, err := json.Marshal(list); err == nil {
		_ = os.WriteFile(remindersPath(), data, 0600)
	}
}

// remindersFor returns the reminders of one WhatsApp profile, soonest first.
func remindersFor(list []reminder, profile string) []reminder {
	out := []reminder{}
	for _, r := range list {
		if r.Profile == profile {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Due < out[j].Due })
	return out
}

func listReminders() []reminder {
	remindersMu.Lock()
	defer remindersMu.Unlock()
	return remindersFor(loadReminders(), currentProfileID)
}

func addReminder(chat, note string, due int64, now time.Time) ([]reminder, error) {
	chat = strings.TrimSpace(chat)
	note = strings.TrimSpace(note)
	if chat == "" {
		return nil, fmt.Errorf("buka sebuah chat dulu")
	}
	if utf8.RuneCountInString(note) > 300 {
		note = string([]rune(note)[:300])
	}
	when := time.Unix(due, 0)
	if !when.After(now) || when.Sub(now) > reminderMaxAhead {
		return nil, fmt.Errorf("waktu pengingat tidak valid")
	}
	remindersMu.Lock()
	defer remindersMu.Unlock()
	list := loadReminders()
	if len(list) >= maxReminders {
		return nil, fmt.Errorf("terlalu banyak pengingat")
	}
	idb := make([]byte, 6)
	_, _ = rand.Read(idb)
	list = append(list, reminder{ID: hex.EncodeToString(idb), Profile: currentProfileID, Chat: chat, Note: note, Due: due})
	saveReminders(list)
	return remindersFor(list, currentProfileID), nil
}

func deleteReminder(id string) []reminder {
	remindersMu.Lock()
	defer remindersMu.Unlock()
	list := loadReminders()
	kept := list[:0]
	for _, r := range list {
		if r.ID != id {
			kept = append(kept, r)
		}
	}
	saveReminders(kept)
	return remindersFor(kept, currentProfileID)
}

// takeDueReminders removes and returns this profile's reminders that are due at now.
func takeDueReminders(now time.Time) []reminder {
	remindersMu.Lock()
	defer remindersMu.Unlock()
	list := loadReminders()
	var due []reminder
	kept := list[:0]
	for _, r := range list {
		if r.Profile == currentProfileID && r.Due <= now.Unix() {
			due = append(due, r)
		} else {
			kept = append(kept, r)
		}
	}
	if len(due) > 0 {
		saveReminders(kept)
	}
	return due
}

func reminderToast(r reminder, now time.Time) (title, message string) {
	title = "⏰ Pengingat: balas " + r.Chat
	if now.Sub(time.Unix(r.Due, 0)) > reminderLateness {
		title = "⏰ Pengingat terlewat: balas " + r.Chat
	}
	message = r.Note
	if message == "" {
		message = "Waktunya membalas chat ini. Klik untuk membukanya."
	}
	return title, message
}

// bindReminders exposes reminders to the WhatsApp page and starts the timer that fires them.
func bindReminders(w webview2.WebView, iconPath string) {
	send := func(list []reminder) {
		data, _ := json.Marshal(list)
		w.Eval(fmt.Sprintf("window.waRemindersChanged && window.waRemindersChanged(%s)", data))
	}
	_ = w.Bind("listReminders", listReminders)
	_ = w.Bind("addReminder", func(chat, note string, due int64) map[string]interface{} {
		list, err := addReminder(chat, note, due, time.Now())
		if err != nil {
			return map[string]interface{}{"ok": false, "error": err.Error()}
		}
		return map[string]interface{}{"ok": true, "list": list}
	})
	_ = w.Bind("deleteReminder", deleteReminder)
	// "Belum dibalas" (assets/followup.js): a chat waited too long for an answer. Unlike reminders this
	// one follows quiet hours and muted chats; one toast per chat replaces the previous one.
	_ = w.Bind("notifyFollowUp", func(chat, message string) {
		s := getNotifySettings()
		if inQuietHours(s, time.Now()) || (chat != "" && isMuted(s, chat)) {
			return
		}
		if chat == "" {
			// Several chats at once: one summary that opens the app
			go pushToast("wa", waToastAppID, "⏳ Belum dibalas", message, iconPath, activationURI("open", "wa", 0, ""),
				[]toast.Action{dismissAction()}, "summary", "followup", false)
			return
		}
		go pushToast("wa", waToastAppID, "⏳ Belum dibalas: "+chat, message, iconPath, activationChatURI(chat),
			[]toast.Action{dismissAction()}, "u"+chatTag(chat), "followup", false)
	})

	go func() {
		for {
			now := time.Now()
			if due := takeDueReminders(now); len(due) > 0 {
				for _, r := range due {
					title, message := reminderToast(r, now)
					pushToast("wa", waToastAppID, title, message, iconPath, activationChatURI(r.Chat),
						[]toast.Action{dismissAction()}, "r"+r.ID, "reminder", false)
				}
				w.Dispatch(func() { send(listReminders()) })
			}
			time.Sleep(reminderTick)
		}
	}()
}
