package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestWatchdogHealthyPageIsLeftAlone(t *testing.T) {
	t0 := time.Now()
	d := newWatchdog(t0)
	for m := 0; m < 60; m++ {
		now := t0.Add(time.Duration(m) * time.Minute)
		d.beat("chats", now)
		if r := d.check(now.Add(30 * time.Second)); r != "" {
			t.Fatalf("reloaded a healthy page at minute %d: %s", m, r)
		}
	}
	// Not being logged in (QR screen) is not a fault either, as long as the page keeps reporting.
	for m := 61; m < 90; m++ {
		now := t0.Add(time.Duration(m) * time.Minute)
		d.beat("qr", now)
		if r := d.check(now.Add(30 * time.Second)); r != "" {
			t.Fatalf("reloaded the QR screen: %s", r)
		}
	}
}

func TestWatchdogSilenceReloads(t *testing.T) {
	t0 := time.Now()
	d := newWatchdog(t0)
	d.beat("chats", t0)
	if r := d.check(t0.Add(2 * time.Minute)); r != "" {
		t.Fatal("too early")
	}
	if r := d.check(t0.Add(3*time.Minute + time.Second)); r != "halaman tidak merespons" {
		t.Fatalf("hung page not reloaded: %q", r)
	}
	// After reloading, the page reports again and the note is shown once.
	if note := d.beat("chats", t0.Add(4*time.Minute)); note != "halaman tidak merespons" {
		t.Fatalf("note: %q", note)
	}
	if note := d.beat("chats", t0.Add(4*time.Minute+30*time.Second)); note != "" {
		t.Fatal("note shown twice")
	}
}

func TestWatchdogLoadingTooLong(t *testing.T) {
	t0 := time.Now()
	d := newWatchdog(t0)
	for s := 0; s <= 180; s += 30 {
		d.beat("loading", t0.Add(time.Duration(s)*time.Second))
	}
	if r := d.check(t0.Add(3*time.Minute + 5*time.Second)); r != "tersangkut di layar memuat" {
		t.Fatalf("stuck loading not reloaded: %q", r)
	}
}

func TestWatchdogNeverLoaded(t *testing.T) {
	t0 := time.Now()
	d := newWatchdog(t0)
	if r := d.check(t0.Add(4 * time.Minute)); r != "" {
		t.Fatal("the first report gets a grace period")
	}
	if r := d.check(t0.Add(5*time.Minute + time.Second)); r != "WhatsApp Web tidak termuat" {
		t.Fatalf("page that never loaded: %q", r)
	}
}

func TestWatchdogBacksOff(t *testing.T) {
	t0 := time.Now()
	d := newWatchdog(t0)
	now := t0
	var reloads []time.Duration
	// The computer is offline: the page never reports. Reloads must spread out.
	for i := 0; i < 400; i++ {
		now = now.Add(watchdogTick)
		if r := d.check(now); r != "" {
			reloads = append(reloads, now.Sub(t0))
		}
	}
	if len(reloads) < 3 {
		t.Fatalf("expected a few reloads, got %v", reloads)
	}
	for i := 2; i < len(reloads); i++ {
		gap, prev := reloads[i]-reloads[i-1], reloads[i-1]-reloads[i-2]
		if gap < prev {
			t.Fatalf("reloads are not backing off: %v", reloads)
		}
	}
	// A healthy period afterwards resets the back-off.
	for m := 0; m <= 11; m++ {
		d.beat("chats", now.Add(time.Duration(m)*time.Minute))
	}
	if d.backoff != watchdogBackoffMin {
		t.Fatalf("back-off not reset: %v", d.backoff)
	}
}

func TestIsRendererOf(t *testing.T) {
	wa := `C:\Users\A\AppData\Roaming\WhatsAppDesktopLight\UserData`
	cases := []struct {
		cmd  string
		want bool
	}{
		{`"msedgewebview2.exe" --type=renderer --user-data-dir="C:\Users\A\AppData\Roaming\WhatsAppDesktopLight\UserData\EBWebView" --lang=id`, true},
		{`msedgewebview2.exe --type=renderer --user-data-dir=C:\Users\A\AppData\Roaming\WhatsAppDesktopLight\UserData\EBWebView --x`, true},
		{`"msedgewebview2.exe" --type=renderer --user-data-dir="C:\Users\A\AppData\Roaming\WhatsAppDesktopLight\UserData_Telegram\EBWebView"`, false},
		{`"msedgewebview2.exe" --type=gpu-process --user-data-dir="C:\Users\A\AppData\Roaming\WhatsAppDesktopLight\UserData\EBWebView"`, false},
		{`"msedgewebview2.exe" --embedded-browser-webview=1 --user-data-dir="C:\Users\A\AppData\Roaming\WhatsAppDesktopLight\UserData\EBWebView"`, false},
	}
	for _, c := range cases {
		if got := isRendererOf(c.cmd, wa); got != c.want {
			t.Errorf("%v for %s", got, c.cmd)
		}
	}
}

func TestProcessCommandLine(t *testing.T) {
	cmd := processCommandLine(uint32(os.Getpid()))
	if !strings.Contains(strings.ToLower(cmd), ".test") && !strings.Contains(strings.ToLower(cmd), "test") {
		t.Fatalf("own command line not read: %q", cmd)
	}
}
