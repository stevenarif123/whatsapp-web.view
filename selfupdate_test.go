package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func testExeBytes(t *testing.T) []byte {
	t.Helper()
	exe, _ := os.Executable()
	data, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestParseSHA256File(t *testing.T) {
	sum := strings.Repeat("ab", 32)
	for _, in := range []string{sum, sum + "  WhatsApp.exe\n", "\uFEFF" + strings.ToUpper(sum) + " *WhatsApp.exe"} {
		if got, err := parseSHA256File([]byte(in)); err != nil || got != sum {
			t.Errorf("%q: %q %v", in, got, err)
		}
	}
	for _, in := range []string{"", "abc", strings.Repeat("zz", 32)} {
		if _, err := parseSHA256File([]byte(in)); err == nil {
			t.Errorf("%q accepted", in)
		}
	}
}

func TestValidWindowsExe(t *testing.T) {
	exe := testExeBytes(t)
	if err := validWindowsExe(exe); err != nil {
		t.Fatalf("the test binary itself should pass: %v", err)
	}
	if validWindowsExe([]byte("<html>not found</html>")) == nil || validWindowsExe(exe[:30]) == nil {
		t.Fatal("non-programs must be rejected")
	}
	x86 := append([]byte(nil), exe[:4096]...)
	off := int(binary.LittleEndian.Uint32(x86[0x3C:]))
	binary.LittleEndian.PutUint16(x86[off+4:], 0x14c)
	if validWindowsExe(x86) == nil {
		t.Fatal("32-bit program must be rejected")
	}
}

func TestRestartArgs(t *testing.T) {
	got := restartArgs([]string{"--profile", "2", "--minimized", "--after-update", "99"}, 1234)
	want := "--profile 2 --after-update 1234"
	if strings.Join(got, " ") != want {
		t.Fatalf("got %v", got)
	}
}

// fakeGitHub serves a release whose assets live on the test server, treated as this repo's downloads.
func fakeGitHub(t *testing.T, exe []byte, sumFile string, assetHost string) {
	t.Helper()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api":
			host := srv.URL
			if assetHost != "" {
				host = assetHost
			}
			fmt.Fprintf(w, `{"tag_name":"v9.9.9","html_url":"https://github.com/%s/releases/tag/v9.9.9","assets":[
				{"name":"WhatsApp.exe","browser_download_url":"%s/dl/WhatsApp.exe"},
				{"name":"WhatsApp.exe.sha256","browser_download_url":"%s/dl/WhatsApp.exe.sha256"}]}`, updateRepo, host, host)
		case "/dl/WhatsApp.exe":
			w.Write(exe)
		case "/dl/WhatsApp.exe.sha256":
			fmt.Fprint(w, sumFile)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	oldAPI, oldPrefix := updateAPIURL, releaseDownloadPrefix
	updateAPIURL, releaseDownloadPrefix = srv.URL+"/api", srv.URL+"/dl/"
	t.Cleanup(func() { updateAPIURL, releaseDownloadPrefix = oldAPI, oldPrefix })
}

func sumOf(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func TestDownloadVerifiedUpdate(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	exe := testExeBytes(t)

	fakeGitHub(t, exe, sumOf(exe)+"  WhatsApp.exe\n", "")
	if info := checkForUpdates(); !info.CanInstall {
		t.Fatalf("release with exe and checksum should be installable: %+v", info)
	}
	rel, err := fetchLatestRelease()
	if err != nil {
		t.Fatal(err)
	}
	got, err := downloadVerifiedUpdate(rel)
	if err != nil || len(got) != len(exe) {
		t.Fatalf("verified download failed: %v", err)
	}

	// A different file than the checksum says is refused.
	fakeGitHub(t, exe, strings.Repeat("0", 64), "")
	rel, _ = fetchLatestRelease()
	if _, err := downloadVerifiedUpdate(rel); err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("checksum mismatch not detected: %v", err)
	}

	// Assets hosted anywhere but this repository's releases are ignored.
	fakeGitHub(t, exe, sumOf(exe), "https://evil.example")
	rel, _ = fetchLatestRelease()
	if rel.ExeURL != "" || rel.SumURL != "" {
		t.Fatalf("foreign assets accepted: %+v", rel)
	}
	if info := checkForUpdates(); info.CanInstall {
		t.Fatal("a release without usable assets must not offer the install button")
	}
	if _, err := downloadLimited("https://evil.example/WhatsApp.exe", 10); err == nil {
		t.Fatal("download outside the repository allowed")
	}
}

func TestSwapExecutableRollsBack(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "app.exe")
	os.WriteFile(exe, []byte("old"), 0755)
	if err := swapExecutable(exe, []byte("new")); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(exe); string(b) != "new" {
		t.Fatal("new version not in place")
	}
	if b, _ := os.ReadFile(exe + ".old"); string(b) != "old" {
		t.Fatal("old version not kept")
	}
	// Missing target: nothing is lost and no temp file stays behind.
	if err := swapExecutable(filepath.Join(dir, "missing.exe"), []byte("x")); err == nil {
		t.Fatal("expected an error")
	}
	if _, err := os.Stat(filepath.Join(dir, "missing.exe.new")); !os.IsNotExist(err) {
		t.Fatal("temp file left behind")
	}
}

// TestSelfUpdateRestart runs a copy of this test binary that replaces its own (running) exe, starts the
// "new" version and exits; the new one must wait for the old process, then remove the .old file.
func TestSelfUpdateRestart(t *testing.T) {
	dir := t.TempDir()
	app := filepath.Join(dir, "app.exe")
	if err := os.WriteFile(app, testExeBytes(t), 0755); err != nil {
		t.Fatal(err)
	}
	// "--" keeps the test framework from parsing --after-update as one of its flags.
	cmd := exec.Command(app, "-test.run=^TestHelperUpdate$", "--")
	cmd.Env = append(os.Environ(), "WA_UPDATE_HELPER=old", "WA_UPDATE_DIR="+dir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("old version failed: %v\n%s", err, out)
	}
	marker := filepath.Join(dir, "new-done")
	for i := 0; i < 100; i++ {
		if _, err := os.Stat(marker); err == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	data, err := os.ReadFile(marker)
	if err != nil {
		t.Fatal("updated version never ran")
	}
	if !strings.Contains(string(data), "old-gone=true") || !strings.Contains(string(data), "old-file-removed=true") {
		t.Fatalf("restart handshake: %s", data)
	}
	if _, err := os.Stat(app); err != nil {
		t.Fatal("app.exe missing after update")
	}
}

func TestHelperUpdate(t *testing.T) {
	dir := os.Getenv("WA_UPDATE_DIR")
	switch os.Getenv("WA_UPDATE_HELPER") {
	case "old":
		exe, _ := os.Executable()
		data, _ := os.ReadFile(exe)
		if err := swapExecutable(exe, data); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(exe, restartArgs(os.Args[1:], os.Getpid())...)
		cmd.Env = append(os.Environ(), "WA_UPDATE_HELPER=new")
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		time.Sleep(1500 * time.Millisecond) // the old app takes a moment to close
	case "new":
		args := os.Args[1:]
		waitForUpdatedFrom(args)
		pid, _ := strconv.Atoi(args[len(args)-1])
		gone := true
		if p, err := os.FindProcess(pid); err == nil {
			// FindProcess opens a handle on Windows only if the process still exists
			gone = p.Signal(os.Kill) != nil
		}
		exe, _ := os.Executable()
		cleanupOldExe()
		_, statErr := os.Stat(exe + ".old")
		os.WriteFile(filepath.Join(dir, "new-done"), []byte(fmt.Sprintf("old-gone=%v old-file-removed=%v", gone, os.IsNotExist(statErr))), 0644)
	default:
		t.Skip("helper process only")
	}
}
