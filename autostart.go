package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Start with Windows: a value in HKCU\...\Run starts the app hidden in the tray ("--minimized").
//
// - On the very first start autostart is switched on; after that the user's choice is kept.
// - Each start repoints the entry at the exe that is running, so moving or replacing the exe does not
//   leave a dead entry. Copies in a temp folder (dev builds, downloads opened from an archive) never
//   take the entry over.
// - Windows can switch an entry off on its own list (Task Manager > Startup apps, Settings > Apps >
//   Startup); that is stored under StartupApproved. It counts as off, and switching autostart on in
//   the app clears it.

var (
	runKeyPath             = `Software\Microsoft\Windows\CurrentVersion\Run`
	startupApprovedKeyPath = `Software\Microsoft\Windows\CurrentVersion\Explorer\StartupApproved\Run`
)

const runKeyName = `WhatsAppDesktopLight`

func autostartCommand(exe string) string {
	return fmt.Sprintf(`"%s" --minimized`, exe)
}

// commandExe returns the exe path from a Run command line.
func commandExe(cmd string) string {
	cmd = strings.TrimSpace(cmd)
	if strings.HasPrefix(cmd, `"`) {
		if end := strings.Index(cmd[1:], `"`); end >= 0 {
			return cmd[1 : end+1]
		}
	}
	if i := strings.Index(cmd, " "); i >= 0 {
		return cmd[:i]
	}
	return cmd
}

func readRegValue(keyPath, name string) ([]byte, bool) {
	var hKey windows.Handle
	subPath, _ := windows.UTF16PtrFromString(keyPath)
	if windows.RegOpenKeyEx(windows.Handle(HKEY_CURRENT_USER), subPath, 0, KEY_READ, &hKey) != nil {
		return nil, false
	}
	defer windows.RegCloseKey(hKey)
	valName, _ := windows.UTF16PtrFromString(name)
	var valType, size uint32
	if windows.RegQueryValueEx(hKey, valName, nil, &valType, nil, &size) != nil || size == 0 {
		return nil, false
	}
	buf := make([]byte, size)
	if windows.RegQueryValueEx(hKey, valName, nil, &valType, &buf[0], &size) != nil {
		return nil, false
	}
	return buf[:size], true
}

func deleteRegValue(keyPath, name string) {
	var hKey windows.Handle
	subPath, _ := windows.UTF16PtrFromString(keyPath)
	if windows.RegOpenKeyEx(windows.Handle(HKEY_CURRENT_USER), subPath, 0, KEY_WRITE, &hKey) != nil {
		return
	}
	defer windows.RegCloseKey(hKey)
	valName, _ := windows.UTF16PtrFromString(name)
	procRegDeleteVal.Call(uintptr(hKey), uintptr(unsafe.Pointer(valName)))
}

// readRunCommand returns the app's Run entry, or "" when there is none.
func readRunCommand() string {
	data, ok := readRegValue(runKeyPath, runKeyName)
	if !ok || len(data) < 2 {
		return ""
	}
	u16 := unsafe.Slice((*uint16)(unsafe.Pointer(&data[0])), len(data)/2)
	return windows.UTF16ToString(u16)
}

// disabledInStartupApps reports whether the entry was switched off in Windows' own startup list.
// The first byte of the StartupApproved value is even when enabled and odd when disabled.
func disabledInStartupApps() bool {
	data, ok := readRegValue(startupApprovedKeyPath, runKeyName)
	return ok && len(data) > 0 && data[0]&1 == 1
}

func isAppAutoStartEnabled() bool {
	return readRunCommand() != "" && !disabledInStartupApps()
}

func writeRunCommand(cmd string) bool {
	var hKey windows.Handle
	subPath, _ := windows.UTF16PtrFromString(runKeyPath)
	r, _, _ := procRegCreateKeyEx.Call(HKEY_CURRENT_USER, uintptr(unsafe.Pointer(subPath)), 0, 0, 0, KEY_WRITE, 0, uintptr(unsafe.Pointer(&hKey)), 0)
	if r != 0 {
		return false
	}
	defer windows.RegCloseKey(hKey)
	valName, _ := windows.UTF16PtrFromString(runKeyName)
	val, err := syscall.UTF16FromString(cmd)
	if err != nil {
		return false
	}
	r, _, _ = procRegSetValueEx.Call(uintptr(hKey), uintptr(unsafe.Pointer(valName)), 0, REG_SZ,
		uintptr(unsafe.Pointer(&val[0])), uintptr(len(val)*2))
	return r == 0
}

func setAppAutoStartEnabled(enabled bool) bool {
	if !enabled {
		deleteRegValue(runKeyPath, runKeyName)
		return readRunCommand() == ""
	}
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	if !writeRunCommand(autostartCommand(exe)) {
		return false
	}
	// A missing StartupApproved value counts as enabled
	deleteRegValue(startupApprovedKeyPath, runKeyName)
	return isAppAutoStartEnabled()
}

// inTempDir reports whether path lies in the user's temp folder.
func inTempDir(path string) bool {
	tmp, err := filepath.Abs(os.TempDir())
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(tmp, path)
	return err == nil && !strings.HasPrefix(rel, "..")
}

func autostartMarkerPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.Getenv("APPDATA")
	}
	base := filepath.Join(configDir, "WhatsAppDesktopLight")
	_ = os.MkdirAll(base, 0755)
	return filepath.Join(base, "autostart_initialized")
}

// ensureAutoStart runs at every start of the main WhatsApp window (profile 1).
func ensureAutoStart() {
	exe, err := os.Executable()
	if err != nil || inTempDir(exe) {
		return
	}
	marker := autostartMarkerPath()
	if _, err := os.Stat(marker); os.IsNotExist(err) {
		// First start: on by default. The marker keeps a later "off" from being undone.
		if setAppAutoStartEnabled(true) {
			_ = os.WriteFile(marker, []byte("1"), 0644)
		}
		return
	}
	cmd := readRunCommand()
	if cmd == "" {
		return // the user switched it off
	}
	if !strings.EqualFold(filepath.Clean(commandExe(cmd)), filepath.Clean(exe)) {
		writeRunCommand(autostartCommand(exe))
	}
}
