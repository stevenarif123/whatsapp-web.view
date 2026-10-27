package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jchv/go-webview2"
)

// Update check: once a day the app asks GitHub for the latest release of this repository and, when it
// is newer than appVersion, tells the page to show a banner with a link to the release page. Nothing
// is downloaded or installed automatically.

// appVersion is the version of this build. Release builds may override it with
// -ldflags "-X main.appVersion=1.2.0".
var appVersion = "1.2.4"

const (
	updateRepo          = "stevenarif123/whatsapp-web.view"
	updateCheckInterval = 24 * time.Hour
	updateFirstDelay    = 45 * time.Second
)

var updateAPIURL = "https://api.github.com/repos/" + updateRepo + "/releases/latest"

type updateState struct {
	Enabled   *bool  `json:"enabled,omitempty"` // nil means on
	LastCheck int64  `json:"lastCheck"`
	Skipped   string `json:"skipped"` // a version the user chose to skip
}

// updateInfo is what the page receives.
type updateInfo struct {
	Current string `json:"current"`
	Latest  string `json:"latest"`
	Newer   bool   `json:"newer"`
	URL     string `json:"url"`
	Notes   string `json:"notes"`
	Skipped bool   `json:"skipped"`
	Enabled bool   `json:"enabled"`
	Error   string `json:"error,omitempty"`
	// CanInstall: the release carries WhatsApp.exe and its checksum, so it can be installed in place.
	CanInstall bool `json:"canInstall"`
}

// releaseInfo is the latest release as reported by GitHub.
type releaseInfo struct {
	Tag, URL, Notes string
	ExeURL, SumURL  string // download links of WhatsApp.exe and WhatsApp.exe.sha256, if published
}

var updateMu sync.Mutex

func updateStatePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.Getenv("APPDATA")
		if configDir == "" {
			configDir = "."
		}
	}
	base := filepath.Join(configDir, "WhatsAppDesktopLight")
	_ = os.MkdirAll(base, 0755)
	return filepath.Join(base, "update_check.json")
}

func loadUpdateState() updateState {
	var s updateState
	if data, err := os.ReadFile(updateStatePath()); err == nil {
		_ = json.Unmarshal(data, &s)
	}
	return s
}

func saveUpdateState(s updateState) {
	if data, err := json.Marshal(s); err == nil {
		_ = os.WriteFile(updateStatePath(), data, 0600)
	}
}

func (s updateState) enabled() bool { return s.Enabled == nil || *s.Enabled }

// parseVersion turns "v1.2.3" or "1.2" into numbers; ok is false for anything else.
func parseVersion(v string) ([3]int, bool) {
	var out [3]int
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if i := strings.IndexAny(v, "-+ "); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	if len(parts) == 0 || len(parts) > 3 {
		return out, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return out, false
		}
		out[i] = n
	}
	return out, true
}

// isNewerVersion reports whether latest is a higher version than current.
func isNewerVersion(latest, current string) bool {
	l, ok1 := parseVersion(latest)
	c, ok2 := parseVersion(current)
	if !ok1 || !ok2 {
		return false
	}
	for i := 0; i < 3; i++ {
		if l[i] != c[i] {
			return l[i] > c[i]
		}
	}
	return false
}

// fetchLatestRelease asks GitHub for the newest published release.
func fetchLatestRelease() (releaseInfo, error) {
	var out releaseInfo
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", updateAPIURL, nil)
	if err != nil {
		return out, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "WhatsAppDesktopLight/"+appVersion)
	resp, err := client.Do(req)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return out, fmt.Errorf("GitHub menjawab %s", resp.Status)
	}
	var rel struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
		Body    string `json:"body"`
		Draft   bool   `json:"draft"`
		Pre     bool   `json:"prerelease"`
		Assets  []struct {
			Name string `json:"name"`
			URL  string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&rel); err != nil {
		return out, err
	}
	if rel.Draft || rel.Pre || rel.TagName == "" {
		return out, fmt.Errorf("tidak ada rilis")
	}
	out.Tag = rel.TagName
	out.URL = rel.HTMLURL
	// Only ever link to this repository's release pages.
	if !strings.HasPrefix(out.URL, "https://github.com/"+updateRepo+"/releases/") {
		out.URL = "https://github.com/" + updateRepo + "/releases/latest"
	}
	out.Notes = rel.Body
	if len(out.Notes) > 600 {
		out.Notes = out.Notes[:600] + "…"
	}
	// Downloads are only taken from this repository's release assets.
	for _, a := range rel.Assets {
		if !strings.HasPrefix(a.URL, releaseDownloadPrefix) {
			continue
		}
		switch a.Name {
		case updateAssetName:
			out.ExeURL = a.URL
		case updateSumName:
			out.SumURL = a.URL
		}
	}
	return out, nil
}

// checkForUpdates queries GitHub now. It runs off the UI thread.
func checkForUpdates() updateInfo {
	updateMu.Lock()
	s := loadUpdateState()
	updateMu.Unlock()

	info := updateInfo{Current: appVersion, Enabled: s.enabled()}
	rel, err := fetchLatestRelease()
	if err != nil {
		info.Error = err.Error()
		return info
	}
	info.Latest = strings.TrimPrefix(rel.Tag, "v")
	info.URL = rel.URL
	info.Notes = rel.Notes
	info.Newer = isNewerVersion(rel.Tag, appVersion)
	info.CanInstall = info.Newer && rel.ExeURL != "" && rel.SumURL != ""
	info.Skipped = info.Newer && s.Skipped == info.Latest

	updateMu.Lock()
	s = loadUpdateState()
	s.LastCheck = time.Now().Unix()
	saveUpdateState(s)
	updateMu.Unlock()
	return info
}

func skipUpdateVersion(version string) {
	updateMu.Lock()
	defer updateMu.Unlock()
	s := loadUpdateState()
	s.Skipped = strings.TrimPrefix(version, "v")
	saveUpdateState(s)
}

func setUpdateCheckEnabled(on bool) bool {
	updateMu.Lock()
	defer updateMu.Unlock()
	s := loadUpdateState()
	s.Enabled = &on
	saveUpdateState(s)
	return on
}

func getUpdateSettings() updateInfo {
	updateMu.Lock()
	defer updateMu.Unlock()
	return updateInfo{Current: appVersion, Enabled: loadUpdateState().enabled()}
}

// startUpdateChecks checks shortly after start and then once a day, and shows the banner through
// show (called on the UI thread via dispatch) when a release newer than this build and not skipped exists.
func startUpdateChecks(dispatch func(func()), show func(updateInfo)) {
	go func() {
		time.Sleep(updateFirstDelay)
		for {
			updateMu.Lock()
			s := loadUpdateState()
			updateMu.Unlock()
			due := time.Since(time.Unix(s.LastCheck, 0)) >= updateCheckInterval
			if s.enabled() && due {
				if info := checkForUpdates(); info.Error == "" && info.Newer && !info.Skipped {
					dispatch(func() { show(info) })
				}
			}
			time.Sleep(time.Hour)
		}
	}()
}

// bindUpdates exposes the update check to the WhatsApp page and starts the daily check. The manual
// check runs in the background; its result arrives through window.waUpdateResult.
func bindUpdates(w webview2.WebView) {
	send := func(fn string, info updateInfo) {
		data, _ := json.Marshal(info)
		w.Eval(fmt.Sprintf("window.%s && window.%s(%s)", fn, fn, data))
	}
	_ = w.Bind("getUpdateSettings", getUpdateSettings)
	_ = w.Bind("setUpdateCheckEnabled", setUpdateCheckEnabled)
	_ = w.Bind("skipUpdateVersion", skipUpdateVersion)
	_ = w.Bind("checkForUpdatesNow", func() {
		go func() {
			info := checkForUpdates()
			w.Dispatch(func() { send("waUpdateResult", info) })
		}()
	})
	_ = w.Bind("installUpdateNow", func() {
		report := func(state, msg string) {
			log.Printf("pembaruan: %s %s", state, msg)
			data, _ := json.Marshal(map[string]string{"state": state, "message": msg})
			w.Dispatch(func() { w.Eval(fmt.Sprintf("window.waUpdateProgress && window.waUpdateProgress(%s)", data)) })
		}
		quit := func() {
			w.Dispatch(func() {
				forceQuit = true
				removeTrayIcon()
				procPostMessage.Call(globalHwnd, WM_CLOSE, 0, 0)
			})
		}
		go installUpdate(report, quit)
	})
	startUpdateChecks(w.Dispatch, func(info updateInfo) { send("waUpdateAvailable", info) })
}
