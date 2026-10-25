package main

import (
	"net/url"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

func at(hhmm string) time.Time {
	t, _ := time.Parse("15:04", hhmm)
	return time.Date(2026, 10, 25, t.Hour(), t.Minute(), 0, 0, time.Local)
}

func TestQuietHours(t *testing.T) {
	overnight := notifySettings{DNDEnabled: true, DNDStart: "22:00", DNDEnd: "07:00"}
	daytime := notifySettings{DNDEnabled: true, DNDStart: "13:00", DNDEnd: "14:30"}
	cases := []struct {
		s    notifySettings
		now  string
		want bool
	}{
		{overnight, "21:59", false},
		{overnight, "22:00", true},
		{overnight, "02:15", true},
		{overnight, "06:59", true},
		{overnight, "07:00", false},
		{daytime, "12:59", false},
		{daytime, "13:30", true},
		{daytime, "14:30", false},
		{notifySettings{DNDEnabled: false, DNDStart: "00:00", DNDEnd: "23:59"}, "12:00", false},
		{notifySettings{DNDEnabled: true, DNDStart: "bad", DNDEnd: "07:00"}, "03:00", false},
	}
	for _, c := range cases {
		if got := inQuietHours(c.s, at(c.now)); got != c.want {
			t.Errorf("%s-%s at %s: got %v, want %v", c.s.DNDStart, c.s.DNDEnd, c.now, got, c.want)
		}
	}
}

func TestMuteSettings(t *testing.T) {
	t.Setenv("APPDATA", t.TempDir())
	muteChat("Grup Kantor")
	muteChat("grup kantor ") // same chat, different case and spacing
	muteChat("Budi")
	s := getNotifySettings()
	if len(s.Muted) != 2 {
		t.Fatalf("expected 2 muted chats, got %v", s.Muted)
	}
	if !isMuted(s, "GRUP KANTOR") || isMuted(s, "Grup") {
		t.Fatal("mute matching should be exact, ignoring case")
	}
	s.Muted = []string{"Budi"}
	s = setNotifySettings(s)
	if isMuted(s, "Grup Kantor") {
		t.Fatal("unmute not saved")
	}
}

func TestActivationURI(t *testing.T) {
	uri := activationURI("open", "wa", 42, "Ibu & Ayah ❤")
	u, err := url.Parse(uri)
	if err != nil || u.Scheme != notifyProtocol || u.Opaque != "open" {
		t.Fatalf("bad uri %q: %v", uri, err)
	}
	q := u.Query()
	if q.Get("app") != "wa" || q.Get("id") != "42" || q.Get("chat") != "Ibu & Ayah ❤" {
		t.Fatalf("query lost data: %v", q)
	}
}

func TestReadCopyData(t *testing.T) {
	const uri = "wadesktoplight:open?app=wa&id=3"
	data, _ := syscall.UTF16FromString(uri)
	cds := copyDataStruct{Magic: copyDataMagic, Size: uint32(len(data) * 2), Data: &data[0]}
	if got := readCopyData(uintptr(unsafe.Pointer(&cds))); got != uri {
		t.Fatalf("got %q", got)
	}
	cds.Magic = 1
	if got := readCopyData(uintptr(unsafe.Pointer(&cds))); got != "" {
		t.Fatal("foreign WM_COPYDATA must be ignored")
	}
}
