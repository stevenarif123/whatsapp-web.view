package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/windows"
)

// Self-update: "Perbarui sekarang" downloads WhatsApp.exe from the latest release of this repository,
// checks it against the SHA-256 published next to it (WhatsApp.exe.sha256) and that it is a 64-bit
// Windows program, then puts it in place of the running exe and restarts. Windows lets a running exe be
// renamed but not overwritten, so the old one becomes <name>.old (removed on the next start) and is put
// back if anything fails. Login session, PIN, notes and settings live in %APPDATA% and are not touched.
//
// The checksum protects against broken or swapped downloads; the trust root is still the GitHub
// account that publishes the release.

const (
	updateAssetName = "WhatsApp.exe"
	updateSumName   = "WhatsApp.exe.sha256"
	maxUpdateSize   = 64 << 20
	afterUpdateFlag = "--after-update"
)

var releaseDownloadPrefix = "https://github.com/" + updateRepo + "/releases/download/"

// parseSHA256File reads the hash from a "<hex>  WhatsApp.exe" style checksum file.
func parseSHA256File(data []byte) (string, error) {
	fields := strings.Fields(strings.TrimPrefix(string(data), "\uFEFF"))
	if len(fields) == 0 {
		return "", fmt.Errorf("file checksum kosong")
	}
	sum := strings.ToLower(fields[0])
	if len(sum) != 64 {
		return "", fmt.Errorf("file checksum tidak valid")
	}
	if _, err := hex.DecodeString(sum); err != nil {
		return "", fmt.Errorf("file checksum tidak valid")
	}
	return sum, nil
}

// validWindowsExe checks the MZ and PE headers and that the program is for x64.
func validWindowsExe(data []byte) error {
	if len(data) < 0x40 || !bytes.HasPrefix(data, []byte("MZ")) {
		return fmt.Errorf("bukan program Windows")
	}
	off := int(binary.LittleEndian.Uint32(data[0x3C:]))
	if off <= 0 || off+6 > len(data) || !bytes.Equal(data[off:off+4], []byte("PE\x00\x00")) {
		return fmt.Errorf("bukan program Windows")
	}
	if machine := binary.LittleEndian.Uint16(data[off+4:]); machine != 0x8664 {
		return fmt.Errorf("program bukan untuk Windows 64-bit")
	}
	return nil
}

func downloadLimited(url string, max int64) ([]byte, error) {
	if !strings.HasPrefix(url, releaseDownloadPrefix) {
		return nil, fmt.Errorf("alamat unduhan bukan rilis %s", updateRepo)
	}
	client := &http.Client{Timeout: 5 * time.Minute}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "WhatsAppDesktopLight/"+appVersion)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unduhan gagal: %s", resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, max+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > max {
		return nil, fmt.Errorf("file terlalu besar")
	}
	return data, nil
}

// downloadVerifiedUpdate fetches the release's exe and checksum and verifies both.
func downloadVerifiedUpdate(rel releaseInfo) ([]byte, error) {
	if rel.ExeURL == "" || rel.SumURL == "" {
		return nil, fmt.Errorf("rilis ini tidak menyertakan %s dan %s", updateAssetName, updateSumName)
	}
	sumFile, err := downloadLimited(rel.SumURL, 4096)
	if err != nil {
		return nil, err
	}
	want, err := parseSHA256File(sumFile)
	if err != nil {
		return nil, err
	}
	exe, err := downloadLimited(rel.ExeURL, maxUpdateSize)
	if err != nil {
		return nil, err
	}
	got := sha256.Sum256(exe)
	if hex.EncodeToString(got[:]) != want {
		return nil, fmt.Errorf("checksum tidak cocok, pembaruan dibatalkan")
	}
	if err := validWindowsExe(exe); err != nil {
		return nil, err
	}
	return exe, nil
}

// swapExecutable puts data in place of exePath, keeping the old file as exePath+".old".
func swapExecutable(exePath string, data []byte) error {
	tmp := exePath + ".new"
	old := exePath + ".old"
	if err := os.WriteFile(tmp, data, 0755); err != nil {
		return fmt.Errorf("tidak bisa menulis di folder aplikasi: %w", err)
	}
	_ = os.Remove(old)
	if err := os.Rename(exePath, old); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("tidak bisa mengganti aplikasi: %w", err)
	}
	if err := os.Rename(tmp, exePath); err != nil {
		_ = os.Rename(old, exePath) // put the running version back
		os.Remove(tmp)
		return fmt.Errorf("tidak bisa mengganti aplikasi: %w", err)
	}
	return nil
}

// cleanupOldExe removes the previous version left by an update. It may still be running in another
// window (a second profile, standalone Telegram); then it is removed on a later start.
func cleanupOldExe() {
	if exe, err := os.Executable(); err == nil {
		_ = os.Remove(exe + ".old")
	}
}

// restartArgs are the arguments for the updated app: same profile, shown (not minimized), and told
// which process to wait for.
func restartArgs(args []string, pid int) []string {
	out := []string{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case afterUpdateFlag:
			i++ // skip its value
		case "--minimized", "--tray", "-hidden":
		default:
			out = append(out, args[i])
		}
	}
	return append(out, afterUpdateFlag, strconv.Itoa(pid))
}

// waitForUpdatedFrom lets a freshly updated app wait until the old process has exited, so the
// single-instance check does not mistake it for a second copy.
func waitForUpdatedFrom(args []string) {
	for i := 0; i+1 < len(args); i++ {
		if args[i] != afterUpdateFlag {
			continue
		}
		pid, err := strconv.Atoi(args[i+1])
		if err != nil || pid <= 0 {
			return
		}
		h, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
		if err != nil {
			return // already gone
		}
		defer windows.CloseHandle(h)
		_, _ = windows.WaitForSingleObject(h, 20000)
		return
	}
}

// installUpdate downloads, verifies and installs the latest release, then restarts the app.
// report is called with "downloading", "installing", "restarting" or "error" (plus a message).
func installUpdate(report func(state, msg string), quit func()) {
	rel, err := fetchLatestRelease()
	if err == nil && !isNewerVersion(rel.Tag, appVersion) {
		err = fmt.Errorf("sudah versi terbaru")
	}
	if err != nil {
		report("error", err.Error())
		return
	}
	report("downloading", strings.TrimPrefix(rel.Tag, "v"))
	data, err := downloadVerifiedUpdate(rel)
	if err != nil {
		report("error", err.Error())
		return
	}
	exe, err := os.Executable()
	if err == nil {
		exe, err = filepath.EvalSymlinks(exe)
	}
	if err != nil {
		report("error", err.Error())
		return
	}
	report("installing", "")
	if err := swapExecutable(exe, data); err != nil {
		report("error", err.Error())
		return
	}
	cmd := exec.Command(exe, restartArgs(os.Args[1:], os.Getpid())...)
	cmd.Dir = filepath.Dir(exe)
	if err := cmd.Start(); err != nil {
		// The new version is in place; it starts next time.
		report("error", "versi baru terpasang, buka ulang aplikasi untuk memakainya")
		return
	}
	report("restarting", "")
	quit()
}
