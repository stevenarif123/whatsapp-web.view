package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsNewerVersion(t *testing.T) {
	cases := []struct {
		latest, current string
		want            bool
	}{
		{"v1.1.0", "1.0.0", true},
		{"v1.0.10", "1.0.9", true},
		{"v2.0", "1.9.9", true},
		{"v1.1.0", "1.1.0", false},
		{"v1.0.9", "1.1.0", false},
		{"v1.2.0-beta", "1.1.0", true},
		{"latest", "1.1.0", false},
		{"", "1.1.0", false},
	}
	for _, c := range cases {
		if got := isNewerVersion(c.latest, c.current); got != c.want {
			t.Errorf("isNewerVersion(%q, %q) = %v", c.latest, c.current, got)
		}
	}
}

func serveRelease(t *testing.T, status int, body string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}))
	t.Cleanup(srv.Close)
	old := updateAPIURL
	updateAPIURL = srv.URL
	t.Cleanup(func() { updateAPIURL = old })
}

func TestCheckForUpdates(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	serveRelease(t, 200, `{"tag_name":"v9.9.9","html_url":"https://github.com/`+updateRepo+`/releases/tag/v9.9.9","body":"catatan"}`)
	info := checkForUpdates()
	if info.Error != "" || !info.Newer || info.Latest != "9.9.9" || info.Notes != "catatan" {
		t.Fatalf("unexpected: %+v", info)
	}
	if loadUpdateState().LastCheck == 0 {
		t.Fatal("last check time not saved")
	}
	skipUpdateVersion("v9.9.9")
	if info := checkForUpdates(); !info.Skipped {
		t.Fatal("skipped version still offered")
	}
}

func TestUpdateLinkStaysOnRepo(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	serveRelease(t, 200, `{"tag_name":"v9.9.9","html_url":"https://evil.example/download"}`)
	info := checkForUpdates()
	if info.URL != "https://github.com/"+updateRepo+"/releases/latest" {
		t.Fatalf("foreign link not replaced: %s", info.URL)
	}
}

func TestUpdateIgnoresDraftsAndErrors(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	serveRelease(t, 200, `{"tag_name":"v9.9.9","prerelease":true}`)
	if info := checkForUpdates(); info.Error == "" || info.Newer {
		t.Fatalf("prerelease must not be offered: %+v", info)
	}
	serveRelease(t, 404, `{"message":"Not Found"}`)
	if info := checkForUpdates(); info.Error == "" {
		t.Fatal("HTTP error not reported")
	}
}

func TestUpdateCheckSetting(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	if !getUpdateSettings().Enabled {
		t.Fatal("update check should be on by default")
	}
	setUpdateCheckEnabled(false)
	if getUpdateSettings().Enabled {
		t.Fatal("setting not saved")
	}
}
