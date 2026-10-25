package main

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jchv/go-webview2"
)

// App lock PIN. The PIN is stored only as a salted PBKDF2 hash in app_pin.dat and is never handed to the
// page: the lock screens ask Go to verify it. Failed attempts and the lockout are kept in the same file,
// so restarting the app does not reset them. All windows and processes share this one file.

const (
	defaultAppPin    = "1234"
	pinIterations    = 100000
	pinFreeAttempts  = 5 // wrong PINs allowed before the first lockout
	pinBaseLockout   = 30 * time.Second
	pinMaxLockout    = 15 * time.Minute
	pinRecordVersion = 2
)

var (
	pinMu      sync.Mutex
	pinPattern = regexp.MustCompile(`^\d{4,6}$`)
)

type pinRecord struct {
	Version     int    `json:"v"`
	Salt        string `json:"salt"`
	Hash        string `json:"hash"`
	Iter        int    `json:"iter"`
	Length      int    `json:"len"`
	IsDefault   bool   `json:"default"`
	Failed      int    `json:"failed"`
	LockedUntil int64  `json:"lockedUntil"` // unix seconds
}

// pinStatus is what the lock screens receive. It never contains the PIN or its hash.
type pinStatus struct {
	OK        bool   `json:"ok"`
	IsDefault bool   `json:"isDefault"`
	Length    int    `json:"length"`
	Wait      int    `json:"wait"`      // seconds until the next attempt is allowed
	Remaining int    `json:"remaining"` // wrong PINs left before the next lockout
	Error     string `json:"error,omitempty"`
}

func hashPin(pin string, salt []byte, iter int) string {
	key, err := pbkdf2.Key(sha256.New, pin, salt, iter, 32)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(key)
}

func newPinRecord(pin string, isDefault bool) pinRecord {
	salt := make([]byte, 16)
	_, _ = rand.Read(salt)
	return pinRecord{
		Version:   pinRecordVersion,
		Salt:      hex.EncodeToString(salt),
		Hash:      hashPin(pin, salt, pinIterations),
		Iter:      pinIterations,
		Length:    len(pin),
		IsDefault: isDefault,
	}
}

// loadPinRecord reads app_pin.dat. A missing or unreadable file means the default PIN; an old plaintext
// file (from before hashing) is converted and rewritten.
func loadPinRecord() pinRecord {
	data, err := os.ReadFile(getPinFilePath())
	if err != nil {
		return newPinRecord(defaultAppPin, true)
	}
	var rec pinRecord
	if json.Unmarshal(data, &rec) == nil && rec.Version == pinRecordVersion && rec.Hash != "" {
		return rec
	}
	legacy := strings.TrimSpace(string(data))
	if !pinPattern.MatchString(legacy) {
		return newPinRecord(defaultAppPin, true)
	}
	rec = newPinRecord(legacy, legacy == defaultAppPin)
	_ = savePinRecord(rec)
	return rec
}

func savePinRecord(rec pinRecord) error {
	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	p := getPinFilePath()
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func (rec pinRecord) status() pinStatus {
	st := pinStatus{IsDefault: rec.IsDefault, Length: rec.Length}
	if wait := time.Until(time.Unix(rec.LockedUntil, 0)); wait > 0 {
		st.Wait = int(wait.Round(time.Second) / time.Second)
	}
	st.Remaining = pinFreeAttempts - rec.Failed
	if st.Remaining < 1 {
		st.Remaining = 1 // past the free attempts every wrong PIN starts a new lockout
	}
	return st
}

func (rec *pinRecord) matches(pin string) bool {
	salt, err := hex.DecodeString(rec.Salt)
	if err != nil {
		return false
	}
	got := hashPin(pin, salt, rec.Iter)
	return got != "" && subtle.ConstantTimeCompare([]byte(got), []byte(rec.Hash)) == 1
}

// checkPin verifies pin against rec, updating the failure counter and lockout. The caller saves rec.
func (rec *pinRecord) checkPin(pin string) pinStatus {
	if st := rec.status(); st.Wait > 0 {
		st.Error = "locked"
		return st
	}
	if rec.matches(strings.TrimSpace(pin)) {
		rec.Failed = 0
		rec.LockedUntil = 0
		st := rec.status()
		st.OK = true
		return st
	}
	rec.Failed++
	if over := rec.Failed - pinFreeAttempts; over >= 0 {
		lockout := pinBaseLockout << min(over, 5)
		if lockout > pinMaxLockout {
			lockout = pinMaxLockout
		}
		rec.LockedUntil = time.Now().Add(lockout).Unix()
	}
	st := rec.status()
	st.Error = "wrong"
	return st
}

func getAppPinStatus() pinStatus {
	pinMu.Lock()
	defer pinMu.Unlock()
	return loadPinRecord().status()
}

func verifyAppPin(pin string) pinStatus {
	pinMu.Lock()
	defer pinMu.Unlock()
	rec := loadPinRecord()
	st := rec.checkPin(pin)
	_ = savePinRecord(rec)
	return st
}

// changeAppPin sets a new PIN. While the default PIN is still in use the old PIN is not asked for.
func changeAppPin(oldPin, newPin string) pinStatus {
	pinMu.Lock()
	defer pinMu.Unlock()
	rec := loadPinRecord()
	if !rec.IsDefault {
		st := rec.checkPin(oldPin)
		_ = savePinRecord(rec)
		if !st.OK {
			return st
		}
	}
	newPin = strings.TrimSpace(newPin)
	if !pinPattern.MatchString(newPin) {
		st := rec.status()
		st.Error = "format"
		return st
	}
	if newPin == defaultAppPin {
		st := rec.status()
		st.Error = "default"
		return st
	}
	rec = newPinRecord(newPin, false)
	if err := savePinRecord(rec); err != nil {
		st := rec.status()
		st.Error = "save"
		return st
	}
	st := rec.status()
	st.OK = true
	return st
}

// broadcastPinChanged lets the lock screens in this process refresh their PIN status.
func broadcastPinChanged() {
	const js = "window.onAppPinChanged && window.onAppPinChanged()"
	if globalWebView != nil {
		globalWebView.Eval(js)
	}
	if globalTelegramWebView != nil {
		globalTelegramWebView.Eval(js)
	}
}

// bindPinLock exposes the PIN functions to a page.
func bindPinLock(w webview2.WebView) {
	_ = w.Bind("getAppPinStatus", getAppPinStatus)
	_ = w.Bind("verifyAppPin", verifyAppPin)
	_ = w.Bind("changeAppPin", func(oldPin, newPin string) pinStatus {
		st := changeAppPin(oldPin, newPin)
		if st.OK {
			broadcastPinChanged()
		}
		return st
	})
}

func getPinFilePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.Getenv("APPDATA")
		if configDir == "" {
			configDir = "."
		}
	}
	base := filepath.Join(configDir, "WhatsAppDesktopLight")
	_ = os.MkdirAll(base, 0755)
	return filepath.Join(base, "app_pin.dat")
}
