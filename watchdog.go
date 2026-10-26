package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/jchv/go-webview2"
	"golang.org/x/sys/windows"
)

// Watchdog: brings WhatsApp Web back when it gets stuck, for an app that runs all day from startup.
// The page reports its state every 30 seconds (assets/watchdog.js): "chats", "qr", "loading" or
// "offline". The page is reloaded when the reports stop (the page hung or its process crashed) or
// when it stays on the loading screen too long. Reloads back off (2, 4, 8 ... 30 minutes) so an
// offline computer does not reload in a loop, and the back-off resets once the chat list was fine
// for a while. Hidden windows get their timers slowed down by the browser, hence the generous limits.

const (
	watchdogTick        = 15 * time.Second
	watchdogSilence     = 3 * time.Minute  // no report for this long: the page hung or crashed
	watchdogStartGrace  = 2 * time.Minute  // time the page gets to send its first report
	watchdogLoadingMax  = 3 * time.Minute  // stuck on the loading screen
	watchdogBackoffMin  = 2 * time.Minute  // minimum time between automatic reloads
	watchdogBackoffMax  = 30 * time.Minute //
	watchdogHealthyTime = 10 * time.Minute // chat list shown this long resets the back-off
)

type watchdog struct {
	mu          sync.Mutex
	started     time.Time
	lastBeat    time.Time
	state       string
	stateSince  time.Time
	lastReload  time.Time
	backoff     time.Duration
	pendingNote string // reason of the last reload, shown once the page reports again
}

func newWatchdog(now time.Time) *watchdog {
	return &watchdog{started: now, backoff: watchdogBackoffMin}
}

// beat records a report from the page and returns a note to show (after an automatic reload).
func (d *watchdog) beat(state string, now time.Time) string {
	d.mu.Lock()
	defer d.mu.Unlock()
	if state != d.state {
		d.state = state
		d.stateSince = now
	}
	d.lastBeat = now
	if state == "chats" && !d.stateSince.IsZero() && now.Sub(d.stateSince) >= watchdogHealthyTime {
		d.backoff = watchdogBackoffMin
	}
	if state == "chats" && d.pendingNote != "" {
		note := d.pendingNote
		d.pendingNote = ""
		return note
	}
	return ""
}

// check decides whether the page needs a reload now; it returns the reason, or "" when not.
func (d *watchdog) check(now time.Time) string {
	d.mu.Lock()
	defer d.mu.Unlock()
	var reason string
	switch {
	case d.lastBeat.IsZero():
		if now.Sub(d.started) >= watchdogStartGrace+watchdogSilence {
			reason = "WhatsApp Web tidak termuat"
		}
	case now.Sub(d.lastBeat) >= watchdogSilence:
		reason = "halaman tidak merespons"
	case d.state == "loading" && now.Sub(d.stateSince) >= watchdogLoadingMax:
		reason = "tersangkut di layar memuat"
	}
	if reason == "" {
		return ""
	}
	if !d.lastReload.IsZero() && now.Sub(d.lastReload) < d.backoff {
		return ""
	}
	if !d.lastReload.IsZero() {
		d.backoff *= 2
		if d.backoff > watchdogBackoffMax {
			d.backoff = watchdogBackoffMax
		}
	}
	d.lastReload = now
	// The reloaded page gets the start grace again before it counts as silent.
	d.lastBeat = time.Time{}
	d.started = now
	d.state = ""
	d.pendingNote = reason
	return reason
}

var procNtQueryInformationProcess = windows.NewLazySystemDLL("ntdll.dll").NewProc("NtQueryInformationProcess")

const processCommandLineInformation = 60

// processCommandLine returns the command line of a process, or "" when it cannot be read.
func processCommandLine(pid uint32) string {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(h)
	buf := make([]byte, 32*1024)
	var size uint32
	r, _, _ := procNtQueryInformationProcess.Call(uintptr(h), processCommandLineInformation,
		uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), uintptr(unsafe.Pointer(&size)))
	if r != 0 {
		return ""
	}
	// The buffer starts with a UNICODE_STRING pointing into the rest of the buffer.
	us := (*struct {
		Length, MaximumLength uint16
		_                     uint32
		Buffer                *uint16
	})(unsafe.Pointer(&buf[0]))
	if us.Buffer == nil || us.Length == 0 {
		return ""
	}
	return windows.UTF16ToString(unsafe.Slice(us.Buffer, us.Length/2))
}

// isRendererOf reports whether a WebView2 command line belongs to a page renderer of the profile in
// userDataDir (WebView2 keeps its data in <userDataDir>\EBWebView). UserData_Telegram does not match
// UserData.
func isRendererOf(cmdLine, userDataDir string) bool {
	cmd := strings.ToLower(cmdLine)
	profile := strings.ToLower(filepath.Join(userDataDir, "EBWebView"))
	return strings.Contains(cmd, "--type=renderer") &&
		(strings.Contains(cmd, profile+`"`) || strings.Contains(cmd, profile+" ") || strings.HasSuffix(cmd, profile))
}

// killRenderers ends the page renderer processes of one profile. A hung page cannot be navigated away
// from; once its renderer is gone, the next navigation starts a fresh one.
func killRenderers(userDataDir string) int {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0
	}
	defer windows.CloseHandle(snap)
	killed := 0
	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	for err := windows.Process32First(snap, &pe); err == nil; err = windows.Process32Next(snap, &pe) {
		if !strings.EqualFold(windows.UTF16ToString(pe.ExeFile[:]), "msedgewebview2.exe") {
			continue
		}
		if !isRendererOf(processCommandLine(pe.ProcessID), userDataDir) {
			continue
		}
		if h, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, pe.ProcessID); err == nil {
			if windows.TerminateProcess(h, 1) == nil {
				killed++
			}
			windows.CloseHandle(h)
		}
	}
	return killed
}

// startWatchdog binds the page's reports and reloads the page when check says so. userDataDir is the
// WebView2 profile folder, used to find the page's renderer when it hangs.
func startWatchdog(w webview2.WebView, url, userDataDir string) {
	d := newWatchdog(time.Now())
	_ = w.Bind("waHeartbeat", func(state string) {
		if note := d.beat(state, time.Now()); note != "" {
			w.Eval(fmt.Sprintf("window.waOnAutoRecovered && window.waOnAutoRecovered(%s)", jsString(note)))
		}
	})
	go func() {
		for {
			time.Sleep(watchdogTick)
			reason := d.check(time.Now())
			if reason == "" {
				continue
			}
			if reason != "tersangkut di layar memuat" {
				// The page does not answer: navigation would wait for it forever.
				killRenderers(userDataDir)
				time.Sleep(time.Second)
			}
			w.Dispatch(func() { w.Navigate(url) })
		}
	}()
}
