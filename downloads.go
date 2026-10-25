package main

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"github.com/jchv/go-webview2"
	"golang.org/x/sys/windows"
)

// Downloads: every download asks for a location with the Windows Save As dialog, then WebView2 writes
// the file there itself (no copy through the page, so file size does not matter). Progress is sent to
// the page, which shows it in a small panel (assets/downloads.js).

type openFileName struct {
	StructSize       uint32
	Owner            uintptr
	Instance         uintptr
	Filter           *uint16
	CustomFilter     *uint16
	MaxCustomFilter  uint32
	FilterIndex      uint32
	File             *uint16
	MaxFile          uint32
	FileTitle        *uint16
	MaxFileTitle     uint32
	InitialDir       *uint16
	Title            *uint16
	Flags            uint32
	FileOffset       uint16
	FileExtension    uint16
	DefaultExtension *uint16
	CustData         uintptr
	Hook             uintptr
	TemplateName     *uint16
	Reserved         uintptr
	Reserved2        uint32
	FlagsEx          uint32
}

var (
	downloadsMu    sync.Mutex
	lastSaveDir    string
	completedFiles = map[int]string{} // download ID -> saved path, for "show in folder"
)

// pickSavePath shows the Save As dialog prefilled with suggested (a full path) and returns the chosen
// path, or "" when the user cancels. The folder of the previous download is offered again.
func pickSavePath(owner uintptr, suggested string) string {
	downloadsMu.Lock()
	if lastSaveDir != "" {
		suggested = filepath.Join(lastSaveDir, filepath.Base(suggested))
	}
	downloadsMu.Unlock()

	fileBuffer := make([]uint16, windows.MAX_PATH*4)
	copy(fileBuffer, windows.StringToUTF16(suggested))
	filter, _ := syscall.UTF16PtrFromString("Semua file\x00*.*\x00\x00")
	title, _ := syscall.UTF16PtrFromString("Simpan File")
	extension, _ := syscall.UTF16PtrFromString(strings.TrimPrefix(filepath.Ext(suggested), "."))
	name := openFileName{
		StructSize:       uint32(unsafe.Sizeof(openFileName{})),
		Owner:            owner,
		Filter:           filter,
		FilterIndex:      1,
		File:             &fileBuffer[0],
		MaxFile:          uint32(len(fileBuffer)),
		Title:            title,
		Flags:            OFN_OVERWRITEPROMPT | OFN_PATHMUSTEXIST | OFN_NOCHANGEDIR,
		DefaultExtension: extension,
	}
	if result, _, _ := procGetSaveFile.Call(uintptr(unsafe.Pointer(&name))); result == 0 {
		return "" // User cancelled.
	}
	path := windows.UTF16ToString(fileBuffer)
	downloadsMu.Lock()
	lastSaveDir = filepath.Dir(path)
	downloadsMu.Unlock()
	return path
}

// showInFolder opens Explorer with the file selected.
func showInFolder(path string) {
	verbPtr, _ := syscall.UTF16PtrFromString("open")
	exePtr, _ := syscall.UTF16PtrFromString("explorer.exe")
	paramPtr, _ := syscall.UTF16PtrFromString(`/select,"` + path + `"`)
	procShellExecute.Call(0, uintptr(unsafe.Pointer(verbPtr)), uintptr(unsafe.Pointer(exePtr)), uintptr(unsafe.Pointer(paramPtr)), 0, 1)
}

// enableDownloads routes the page's downloads through the Save As dialog and the download panel.
// owner is the top-level window the dialog belongs to.
func enableDownloads(w webview2.WebView, owner uintptr) {
	ok := w.SetDownloadHandler(func(d webview2.DownloadInfo) string {
		return pickSavePath(owner, d.SuggestedPath)
	}, func(p webview2.DownloadProgress) {
		if p.State == "completed" {
			downloadsMu.Lock()
			completedFiles[p.ID] = p.Path
			downloadsMu.Unlock()
		}
		// The page only gets the file name, not the full path.
		data, _ := json.Marshal(map[string]interface{}{
			"id":       p.ID,
			"name":     filepath.Base(p.Path),
			"received": p.Received,
			"total":    p.Total,
			"state":    p.State,
		})
		w.Eval(fmt.Sprintf("window.waDownloadProgress && window.waDownloadProgress(%s)", data))
	})
	if !ok {
		log.Println("WebView2 runtime has no download events; downloads use the default behaviour")
	}

	_ = w.Bind("cancelDownload", func(id int) {
		w.CancelDownload(id)
	})
	_ = w.Bind("showDownloadInFolder", func(id int) {
		downloadsMu.Lock()
		path, found := completedFiles[id]
		downloadsMu.Unlock()
		if found {
			showInFolder(path)
		}
	})
}
