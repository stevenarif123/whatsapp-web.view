package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// The diagnostic log (app.log, app-profileN.log or telegram.log in the app's config folder) records what
// the app does around its fragile parts: startup, updates, opening Telegram and a stalled dispatch queue.
// It never holds chat names or message text.
const maxLogSize = 1 << 20 // the current file rolls over to .1 beyond this

type rollingLog struct {
	mu   sync.Mutex
	path string
	f    *os.File
	size int64
}

func openRollingLog(path string) (*rollingLog, error) {
	l := &rollingLog{path: path}
	if err := l.open(); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *rollingLog) open() error {
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	st, err := f.Stat()
	if err != nil {
		f.Close()
		return err
	}
	l.f, l.size = f, st.Size()
	return nil
}

func (l *rollingLog) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.f != nil && l.size+int64(len(p)) > maxLogSize {
		l.f.Close()
		l.f = nil
		_ = os.Remove(l.path + ".1")
		_ = os.Rename(l.path, l.path+".1")
	}
	if l.f == nil {
		if err := l.open(); err != nil {
			return 0, err
		}
	}
	n, err := l.f.Write(p)
	l.size += int64(n)
	return n, err
}

func logFileName(profile string) string {
	switch profile {
	case "", "1":
		return "app.log"
	case "telegram":
		return "telegram.log"
	}
	return fmt.Sprintf("app-profile%s.log", filepath.Base(profile))
}

// setupLogging sends the standard logger (and so the WebView2 library's messages) to the diagnostic log.
func setupLogging(profile string) {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = os.Getenv("APPDATA")
	}
	dir = filepath.Join(dir, "WhatsAppDesktopLight")
	_ = os.MkdirAll(dir, 0755)
	l, err := openRollingLog(filepath.Join(dir, logFileName(profile)))
	if err != nil {
		log.SetOutput(io.Discard)
		return
	}
	log.SetOutput(l)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
}
