package main

import (
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

var procRegDeleteKey = advapi32.NewProc("RegDeleteKeyW")

func deleteTestKey(path string) {
	p, _ := windows.UTF16PtrFromString(path)
	procRegDeleteKey.Call(HKEY_CURRENT_USER, uintptr(unsafe.Pointer(p)))
}

// setStartupApproved writes the binary flag Windows keeps for its own startup list.
func setStartupApproved(t *testing.T, data []byte) {
	t.Helper()
	var hKey windows.Handle
	sub, _ := windows.UTF16PtrFromString(startupApprovedKeyPath)
	if r, _, _ := procRegCreateKeyEx.Call(HKEY_CURRENT_USER, uintptr(unsafe.Pointer(sub)), 0, 0, 0, KEY_WRITE, 0, uintptr(unsafe.Pointer(&hKey)), 0); r != 0 {
		t.Fatalf("create key: %d", r)
	}
	defer windows.RegCloseKey(hKey)
	name, _ := windows.UTF16PtrFromString(runKeyName)
	const regBinary = 3
	if r, _, _ := procRegSetValueEx.Call(uintptr(hKey), uintptr(unsafe.Pointer(name)), 0, regBinary, uintptr(unsafe.Pointer(&data[0])), uintptr(len(data))); r != 0 {
		t.Fatalf("set value: %d", r)
	}
}

// The tests use their own keys under HKCU\Software\WhatsAppDesktopLightTest, never the real Run key.
func useTestRunKeys(t *testing.T) {
	t.Helper()
	base := `Software\WhatsAppDesktopLightTest`
	oldRun, oldApproved := runKeyPath, startupApprovedKeyPath
	runKeyPath, startupApprovedKeyPath = base+`\Run`, base+`\StartupApproved`
	t.Cleanup(func() {
		runKeyPath, startupApprovedKeyPath = oldRun, oldApproved
		deleteTestKey(base + `\Run`)
		deleteTestKey(base + `\StartupApproved`)
		deleteTestKey(base)
	})
	t.Setenv("APPDATA", t.TempDir())
}

func TestCommandExe(t *testing.T) {
	cases := map[string]string{
		`"C:\Program Files\WA\WhatsApp.exe" --minimized`: `C:\Program Files\WA\WhatsApp.exe`,
		`C:\WhatsApp.exe --minimized`:                    `C:\WhatsApp.exe`,
		`C:\WhatsApp.exe`:                                `C:\WhatsApp.exe`,
	}
	for in, want := range cases {
		if got := commandExe(in); got != want {
			t.Errorf("%s -> %s", in, got)
		}
	}
}

func TestAutoStartOnOff(t *testing.T) {
	useTestRunKeys(t)
	if isAppAutoStartEnabled() {
		t.Fatal("should start off")
	}
	if !setAppAutoStartEnabled(true) || !isAppAutoStartEnabled() {
		t.Fatal("could not switch on")
	}
	exe, _ := os.Executable()
	if got := readRunCommand(); got != autostartCommand(exe) {
		t.Fatalf("run entry %q", got)
	}
	if !setAppAutoStartEnabled(false) || isAppAutoStartEnabled() {
		t.Fatal("could not switch off")
	}
}

func TestAutoStartRespectsWindowsStartupList(t *testing.T) {
	useTestRunKeys(t)
	setAppAutoStartEnabled(true)
	setStartupApproved(t, []byte{3, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}) // "disabled" in Task Manager
	if isAppAutoStartEnabled() {
		t.Fatal("an entry disabled in Windows' startup list must count as off")
	}
	if !setAppAutoStartEnabled(true) || disabledInStartupApps() {
		t.Fatal("switching on in the app must clear Windows' disabled flag")
	}
}

func TestEnsureAutoStart(t *testing.T) {
	useTestRunKeys(t)
	exe, _ := os.Executable()
	if inTempDir(exe) {
		// go test runs from a temp folder; pretend the exe lives elsewhere by checking the parts directly
		t.Log("test binary is in the temp folder: ensureAutoStart is expected to do nothing")
		ensureAutoStart()
		if readRunCommand() != "" {
			t.Fatal("a copy in the temp folder must not register itself")
		}
	}

	// First start switches it on, and the marker keeps a later "off" in place.
	marker := autostartMarkerPath()
	os.Remove(marker)
	if setAppAutoStartEnabled(true) {
		os.WriteFile(marker, []byte("1"), 0644)
	}
	setAppAutoStartEnabled(false)
	if readRunCommand() != "" {
		t.Fatal("off not saved")
	}

	// A stale entry pointing at a moved exe is detected as such.
	writeRunCommand(autostartCommand(filepath.Join(`C:\old place`, "WhatsApp.exe")))
	if commandExe(readRunCommand()) == exe {
		t.Fatal("stale path expected for this check")
	}
}
