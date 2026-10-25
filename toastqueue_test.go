package main

import (
	"strings"
	"testing"
	"time"

	"github.com/go-toast/toast"
)

type sentToast struct {
	title, message, launch, tag string
	silent                      bool
	actions                     int
}

func captureToasts(t *testing.T) *[]sentToast {
	t.Helper()
	var got []sentToast
	old := sendToast
	sendToast = func(app, appID, title, message, icon, launch string, actions []toast.Action, tag, group string, silent bool) {
		got = append(got, sentToast{title, message, launch, tag, silent, len(actions)})
	}
	t.Cleanup(func() { sendToast = old; resumedAt.Store(0) })
	resumedAt.Store(0)
	return &got
}

func req(title string, id int) toastReq {
	return toastReq{app: "wa", appID: waToastAppID, title: title, message: "isi", id: id}
}

func TestSingleToastHasChatTagAndClose(t *testing.T) {
	got := captureToasts(t)
	b := &toastBatch{items: []toastReq{req("Budi", 7)}}
	now := time.Now()
	flushBatch(b, now)
	if len(*got) != 1 {
		t.Fatalf("got %d toasts", len(*got))
	}
	s := (*got)[0]
	if s.title != "Budi" || s.tag != chatTag("budi ") || s.silent || s.actions != 2 {
		t.Fatalf("single toast: %+v", s)
	}
	if !strings.Contains(s.launch, "id=7") {
		t.Fatalf("click should open the chat: %s", s.launch)
	}
}

func TestBurstBecomesOneSummary(t *testing.T) {
	got := captureToasts(t)
	b := &toastBatch{}
	for i, c := range []string{"Budi", "Grup Kantor", "Grup Kantor", "Siti", "Grup Kantor", "Budi", "Andi", "Rina"} {
		b.items = append(b.items, req(c, i+1))
	}
	now := time.Now()
	flushBatch(b, now)
	if len(*got) != 1 {
		t.Fatalf("burst gave %d toasts, want 1", len(*got))
	}
	s := (*got)[0]
	if s.title != "8 pesan baru dari 5 chat" || s.message != "Grup Kantor (3), Budi (2), Siti, +2 chat lain" {
		t.Fatalf("summary: %q / %q", s.title, s.message)
	}
	if s.tag != "summary" || s.silent || strings.Contains(s.launch, "chat=Budi") {
		t.Fatalf("summary toast: %+v", s)
	}

	// A second wave soon after is merged into the same summary, without popping up again.
	b.items = []toastReq{req("Budi", 20), req("Tono", 21)}
	flushBatch(b, now.Add(10*time.Second))
	s = (*got)[1]
	if s.title != "10 pesan baru dari 6 chat" || !s.silent || s.tag != "summary" {
		t.Fatalf("second wave: %+v", s)
	}

	// A lone message during the cooldown still pops up (updated summary).
	b.items = []toastReq{req("Budi", 22)}
	flushBatch(b, now.Add(20*time.Second))
	if s = (*got)[2]; s.silent {
		t.Fatal("a single new message must pop up")
	}

	// After the cooldown, single messages are single toasts again.
	b.items = []toastReq{req("Siti", 30)}
	flushBatch(b, now.Add(2*time.Minute))
	if s = (*got)[3]; s.title != "Siti" || s.tag == "summary" {
		t.Fatalf("after cooldown: %+v", s)
	}
}

func TestSummaryForOneChatOpensIt(t *testing.T) {
	got := captureToasts(t)
	b := &toastBatch{items: []toastReq{req("Budi", 3), req("Budi", 4), req("Budi", 5)}}
	flushBatch(b, time.Now())
	s := (*got)[0]
	if s.title != "Budi" || s.message != "3 pesan baru" || !strings.Contains(s.launch, "id=5") {
		t.Fatalf("one-chat summary: %+v", s)
	}
}

func TestAfterResumeEverythingWaits(t *testing.T) {
	captureToasts(t)
	now := time.Now()
	b := &toastBatch{items: []toastReq{req("Budi", 1)}, first: now, last: now}
	if !batchReady(b, now.Add(2*time.Second)) {
		t.Fatal("a lone toast should be shown after the short hold")
	}
	markResumed()
	if batchReady(b, now.Add(5*time.Second)) {
		t.Fatal("right after waking up the replay must be collected first")
	}
	if !batchReady(b, now.Add(16*time.Second)) {
		t.Fatal("the replay should be shown once it went quiet")
	}
}

func TestTelegramSummaryIsLabelled(t *testing.T) {
	title, _ := summaryText("tg", map[string]int{"A": 1, "B": 1}, []string{"A", "B"}, 2)
	if title != "Telegram: 2 pesan baru dari 2 chat" {
		t.Fatal(title)
	}
}
