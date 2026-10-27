package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogFileName(t *testing.T) {
	for profile, want := range map[string]string{"": "app.log", "1": "app.log", "2": "app-profile2.log", "telegram": "telegram.log", `..\x`: "app-profilex.log"} {
		if got := logFileName(profile); got != want {
			t.Errorf("logFileName(%q) = %q, want %q", profile, got, want)
		}
	}
}

func TestRollingLogRollsOver(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	l, err := openRollingLog(path)
	if err != nil {
		t.Fatal(err)
	}
	line := strings.Repeat("x", 1000) + "\n"
	for written := 0; written <= maxLogSize; written += len(line) {
		if _, err := l.Write([]byte(line)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := l.Write([]byte("last\n")); err != nil {
		t.Fatal(err)
	}
	l.f.Close()
	old, err := os.Stat(path + ".1")
	if err != nil {
		t.Fatalf("no rolled-over file: %v", err)
	}
	if old.Size() > maxLogSize {
		t.Errorf("rolled-over file is %d bytes, over the %d limit", old.Size(), maxLogSize)
	}
	cur, _ := os.ReadFile(path)
	if len(cur) > maxLogSize || !strings.HasSuffix(string(cur), "last\n") {
		t.Errorf("current file has %d bytes, ending %q", len(cur), string(cur[max(0, len(cur)-5):]))
	}
}
