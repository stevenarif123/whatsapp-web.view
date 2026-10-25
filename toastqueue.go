package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-toast/toast"
)

// Toast batching. When the computer wakes from sleep (or many messages arrive at once) WhatsApp and
// Telegram replay every held message as a notification. Shown one by one, each toast takes a PowerShell
// process and they keep popping up over whatever the user is doing. Instead, every toast waits briefly;
// when more arrive meanwhile they become one summary ("12 pesan baru: Budi (3), ..."). Further batches
// soon after update that summary in the Action Center without popping up again. Single toasts carry a
// tag per chat, so a newer message from the same chat replaces the older toast.

const (
	toastHold        = 1500 * time.Millisecond // a lone toast waits this long for company
	toastBatchQuiet  = 4 * time.Second         // a batch is shown once nothing new came for this long
	toastBatchMax    = 20 * time.Second        // ...or at the latest after this long
	toastResumeQuiet = 90 * time.Second        // after waking up, wait for the whole replay
	summaryCooldown  = 60 * time.Second        // later batches within this time update the summary silently
	toastTick        = 250 * time.Millisecond
	clockJump        = 20 * time.Second // wall clock moving this much more than the monotonic clock means sleep
)

type toastReq struct {
	app, appID, title, message, icon string
	id                               int
}

type toastBatch struct {
	items       []toastReq
	first, last time.Time
	lastSummary time.Time
	chats       map[string]int // chats counted in the current summary
	order       []string       // chats in first-seen order
	total       int
	lastID      map[string]int // newest page notification ID per chat
}

var (
	// sendToast shows a toast; tests replace it.
	sendToast = pushToast

	toastQueue    = make(chan toastReq, 512)
	toastLoopOnce sync.Once
	resumedAt     atomic.Int64 // unix nanoseconds of the last wake-up
)

// markResumed is called when Windows reports a wake-up from sleep.
func markResumed() { resumedAt.Store(time.Now().UnixNano()) }

func recentlyResumed(now time.Time) bool {
	t := resumedAt.Load()
	return t != 0 && now.Sub(time.Unix(0, t)) < toastResumeQuiet
}

// queueToast hands a toast to the batching loop.
func queueToast(r toastReq) {
	toastLoopOnce.Do(func() { go toastLoop() })
	select {
	case toastQueue <- r:
	default: // a flood beyond the buffer adds nothing to the summary
	}
}

func toastLoop() {
	batches := map[string]*toastBatch{}
	tick := time.NewTicker(toastTick)
	prev := time.Now()
	for {
		select {
		case r := <-toastQueue:
			now := time.Now()
			b := batches[r.app]
			if b == nil {
				b = &toastBatch{}
				batches[r.app] = b
			}
			if len(b.items) == 0 {
				b.first = now
			}
			b.items = append(b.items, r)
			b.last = now
		case now := <-tick.C:
			// Windows does not advance Go's monotonic clock during sleep; the wall clock does.
			if now.Round(0).Sub(prev.Round(0))-now.Sub(prev) > clockJump {
				markResumed()
			}
			prev = now
			for _, b := range batches {
				if len(b.items) > 0 && batchReady(b, now) {
					flushBatch(b, now)
				}
			}
		}
	}
}

func batchReady(b *toastBatch, now time.Time) bool {
	quiet := toastHold
	if len(b.items) > 1 {
		quiet = toastBatchQuiet
	}
	if recentlyResumed(now) {
		quiet = toastResumeQuiet / 6 // 15 s: the replay after sleep comes in waves
	}
	return now.Sub(b.last) >= quiet || now.Sub(b.first) >= toastBatchMax
}

func flushBatch(b *toastBatch, now time.Time) {
	items := b.items
	b.items = nil
	inCooldown := !b.lastSummary.IsZero() && now.Sub(b.lastSummary) < summaryCooldown
	if len(items) == 1 && !inCooldown && !recentlyResumed(now) {
		showSingleToast(items[0])
		return
	}
	if !inCooldown {
		b.chats, b.order, b.total, b.lastID = map[string]int{}, nil, 0, map[string]int{}
	}
	for _, r := range items {
		chat := strings.TrimSpace(r.title)
		if chat == "" {
			chat = "Pesan"
		}
		if _, seen := b.chats[chat]; !seen {
			b.order = append(b.order, chat)
		}
		b.chats[chat]++
		b.total++
		if r.id > 0 {
			b.lastID[chat] = r.id
		}
	}
	title, message := summaryText(items[0].app, b.chats, b.order, b.total)
	launch := activationURI("open", items[0].app, 0, "")
	if len(b.order) == 1 {
		chat := b.order[0]
		launch = activationURI("open", items[0].app, b.lastID[chat], chat)
	}
	last := items[len(items)-1]
	// Update the summary quietly while a replay is still coming in; a lone new message pops up again.
	quiet := inCooldown && (len(items) > 1 || recentlyResumed(now))
	sendToast(last.app, last.appID, title, message, last.icon, launch, []toast.Action{dismissAction()},
		"summary", last.app, quiet)
	b.lastSummary = now
}

// summaryText builds the summary toast, busiest chats first.
func summaryText(app string, chats map[string]int, order []string, total int) (title, message string) {
	names := append([]string(nil), order...)
	sort.SliceStable(names, func(i, j int) bool { return chats[names[i]] > chats[names[j]] })
	if len(names) == 1 {
		title = names[0]
		if app == "tg" {
			title = "Telegram: " + title
		}
		return title, fmt.Sprintf("%d pesan baru", total)
	}
	title = fmt.Sprintf("%d pesan baru dari %d chat", total, len(names))
	if app == "tg" {
		title = "Telegram: " + title
	}
	var parts []string
	for i, n := range names {
		if i == 3 {
			parts = append(parts, fmt.Sprintf("+%d chat lain", len(names)-3))
			break
		}
		if chats[n] > 1 {
			parts = append(parts, fmt.Sprintf("%s (%d)", n, chats[n]))
		} else {
			parts = append(parts, n)
		}
	}
	return title, strings.Join(parts, ", ")
}

func showSingleToast(r toastReq) {
	actions := []toast.Action{}
	if r.id > 0 && r.title != "" {
		actions = append(actions, toast.Action{Type: "protocol", Label: "Bisukan chat ini", Arguments: activationURI("mute", r.app, r.id, r.title)})
	}
	actions = append(actions, dismissAction())
	sendToast(r.app, r.appID, r.title, r.message, r.icon, activationURI("open", r.app, r.id, r.title), actions,
		chatTag(r.title), r.app, false)
}

// dismissAction is a visible "Tutup" button; Windows only shows its own close button on hover.
func dismissAction() toast.Action {
	return toast.Action{Type: "system", Label: "Tutup", Arguments: "dismiss"}
}

// chatTag gives every chat a short stable toast tag (Windows allows at most 64 characters).
func chatTag(chat string) string {
	sum := sha1.Sum([]byte(strings.ToLower(strings.TrimSpace(chat))))
	return "c" + hex.EncodeToString(sum[:8])
}
