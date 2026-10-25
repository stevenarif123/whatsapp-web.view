package main

import (
	"os"
	"strings"
	"testing"
)

func usePinDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	return getPinFilePath()
}

func TestPinDefaultAndChange(t *testing.T) {
	usePinDir(t)
	if st := getAppPinStatus(); !st.IsDefault || st.Length != 4 {
		t.Fatalf("fresh install should use the default PIN: %+v", st)
	}
	if st := verifyAppPin(defaultAppPin); !st.OK {
		t.Fatalf("default PIN rejected: %+v", st)
	}
	if st := changeAppPin("", defaultAppPin); st.OK || st.Error != "default" {
		t.Fatalf("setting the default PIN again must fail: %+v", st)
	}
	if st := changeAppPin("", "12a4"); st.OK || st.Error != "format" {
		t.Fatalf("non-digit PIN must fail: %+v", st)
	}
	if st := changeAppPin("", "482915"); !st.OK || st.IsDefault || st.Length != 6 {
		t.Fatalf("change from default failed: %+v", st)
	}
	if st := verifyAppPin(defaultAppPin); st.OK {
		t.Fatal("old default PIN still accepted")
	}
	if st := changeAppPin("000000", "5555"); st.OK || st.Error != "wrong" {
		t.Fatalf("change with a wrong old PIN must fail: %+v", st)
	}
	if st := changeAppPin("482915", "5555"); !st.OK {
		t.Fatalf("change with the right old PIN failed: %+v", st)
	}
}

func TestPinFileHasNoPlaintext(t *testing.T) {
	p := usePinDir(t)
	changeAppPin("", "731904")
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "731904") {
		t.Fatalf("PIN stored in plaintext: %s", data)
	}
}

func TestPinLegacyMigration(t *testing.T) {
	p := usePinDir(t)
	if err := os.WriteFile(p, []byte("2468\r\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if st := verifyAppPin("2468"); !st.OK || st.IsDefault {
		t.Fatalf("legacy plaintext PIN not migrated: %+v", st)
	}
	data, _ := os.ReadFile(p)
	if strings.Contains(string(data), "2468") {
		t.Fatal("legacy file not rewritten as a hash")
	}
}

func TestPinLockout(t *testing.T) {
	usePinDir(t)
	changeAppPin("", "9999")
	for i := 1; i < pinFreeAttempts; i++ {
		st := verifyAppPin("0000")
		if st.OK || st.Wait != 0 || st.Remaining != pinFreeAttempts-i {
			t.Fatalf("attempt %d: %+v", i, st)
		}
	}
	st := verifyAppPin("0000")
	if st.Wait < 29 {
		t.Fatalf("expected a lockout after %d wrong PINs: %+v", pinFreeAttempts, st)
	}
	// Even the right PIN is refused during the lockout.
	if st := verifyAppPin("9999"); st.OK || st.Error != "locked" {
		t.Fatalf("right PIN accepted during lockout: %+v", st)
	}
	// The lockout survives a reload from disk (an app restart).
	if st := getAppPinStatus(); st.Wait == 0 {
		t.Fatal("lockout lost after reload")
	}
}
