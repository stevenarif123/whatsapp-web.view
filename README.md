# WhatsApp Webview Desktop

Pembungkus desktop ringan untuk [WhatsApp Web](https://web.whatsapp.com) di Windows, dibuat dengan Go dan Microsoft Edge WebView2 (sekitar 50-80 MB RAM, bukan 400 MB+ seperti Electron).

Semua fitur tambahan hanya mengubah tampilan di layar Anda. Tidak ada yang memodifikasi koneksi atau enkripsi WhatsApp, sehingga akun tidak berisiko diblokir karena fitur-fitur ini.

## Fitur

### Pusat Kontrol (`Alt + M`)
Panel pengaturan dengan lima bagian: **Privasi**, **Tampilan**, **Chat & Alat**, **Sistem**, dan **Diagnostik**. Bisa dibuka dari ikon di sidebar WhatsApp, dock mengambang di kiri bawah, atau klik kanan ikon tray.

### Privasi
- **Mode Privasi / blur layar** (`Alt + P`): buramkan nama kontak, pratinjau pesan, balon obrolan, foto & video, dan foto profil. Intensitas Halus / Sedang / Kuat. Arahkan kursor untuk mengintip.
- **Screenshot Anonim**: samarkan nama dan foto profil dengan satu klik sebelum mengambil tangkapan layar.
- **Kunci Aplikasi** (`Ctrl + L`): kunci dengan PIN, terkunci otomatis setelah 5 menit tidak aktif. Saat pertama mengunci Anda diminta membuat PIN sendiri (PIN bawaan `1234` tidak bisa dipakai untuk mengunci). PIN disimpan sebagai hash PBKDF2 dan tidak pernah dikirim ke halaman; setelah 5 PIN salah percobaan dijeda (30 detik, makin lama hingga 15 menit), juga setelah aplikasi dibuka ulang. PIN bisa diganti dari Pusat Kontrol, layar kunci, atau menu tray.
- **Boss Key global** (`Ctrl + Alt + W`): sembunyikan atau munculkan jendela dari aplikasi mana pun.

### Chat & Alat
- **Pin chat tanpa batas**: arahkan kursor ke chat lalu klik ikon pin. Chat yang dipin muncul sebagai chip di atas daftar chat. Pin hanya tersimpan di komputer ini dan tidak mengubah pin bawaan WhatsApp.
- **Blur per chat**: klik ikon mata di baris chat agar chat itu selalu buram (arahkan kursor untuk mengintip), tanpa menyalakan Mode Privasi untuk semuanya.
- **Filter chat belum dibaca**: memakai chip "Belum dibaca" bawaan WhatsApp.
- **Radar online**: toast dan notifikasi Windows saat kontak di chat yang sedang terbuka menjadi online atau mulai mengetik. Bisa dibatasi ke nama tertentu. *Catatan: WhatsApp hanya mengirim status online untuk chat yang sedang Anda buka, dan hanya jika kontak mengizinkannya.*
- **Balasan cepat**: buat pintasan seperti `/toko`, ketik di chat, lalu tekan Tab, Enter, atau Spasi untuk mengembangkannya. Teks bisa memakai variabel `{nama}` (nama chat), `{depan}`, `{tanggal}`, `{hari}`, dan `{jam}`. Daftarnya sama di WhatsApp dan Telegram.
- **Pengingat balas** (`Alt + R` atau tombol Ingatkan di dock): pilih 30 menit, 1 jam, 3 jam, besok 09.00, atau waktu sendiri, plus catatan opsional. Pada waktunya muncul notifikasi; klik untuk langsung membuka chatnya. Pengingat yang terlewat saat aplikasi tertutup muncul begitu aplikasi dibuka. Semuanya tersimpan di komputer ini, tidak ada yang dikirim ke WhatsApp.
- **Ekspor chat**: simpan pesan yang sudah termuat di chat terbuka sebagai berkas `.txt`.
- **Chat ke nomor baru** (`Ctrl + N`): tanpa menyimpan kontak. `0812…` otomatis menjadi `62812…`.
- **Catatan** (`Alt + N`): panel samping dengan banyak catatan, pencarian, dan simpan otomatis. Catatan yang sama muncul di WhatsApp dan Telegram.
- **Unduh status**: tombol unduh muncul saat Anda membuka status foto atau video.
- **Simpan Sebagai & panel unduhan**: setiap unduhan (WhatsApp dan Telegram) membuka dialog *Save As* Windows, lalu WebView2 menulis file langsung ke sana, jadi ukuran file tidak masalah. Progres tampil di panel kecil kanan bawah dengan tombol Batal dan Buka folder.
- **Pintasan clipboard standar**: `Ctrl C`, `Ctrl X`, `Ctrl V`, dan `Ctrl A` berfungsi normal.

### Tampilan
- Tema **Standar gelap**, **OLED hitam**, dan **Layar hangat**.
- **Daftar chat rapat**.
- **Mode ringkas di jendela sempit**: di bawah ±900 px (misalnya dibagi dua dengan aplikasi lain) WhatsApp Web biasanya terpotong karena lebar minimumnya 748 px dan di bawah ±900 px panel chat masih terlalu sempit sehingga bubble dan gambar terpotong. Aplikasi ini menampilkan satu panel saja: daftar chat, atau chat yang dibuka dengan tombol kembali (`‹`), seperti aplikasi native. Di jendela sedang, daftar chat dibatasi 300–420 px (sekitar 30%, bukan 40% seperti WhatsApp Web) supaya panel chat lebih lega.
- **Zoom** 75-150% (`Ctrl +` / `Ctrl -` / `Ctrl 0` atau `Ctrl` + scroll). Memakai zoom bawaan browser (WebView2), sehingga WhatsApp menyusun ulang tata letaknya di setiap ukuran dan tidak ada yang terpotong.
- **Kecepatan voice note** 0.5×-3.0× (`[` dan `]`).
- **Picture-in-Picture** untuk video (`Alt + V`).

### Sistem
- **Dua akun WhatsApp** dalam jendela terpisah (`--profile 2`).
- **Telegram Web** berdampingan dengan WhatsApp (`Ctrl + 2`), serta **tampilan berdampingan 50:50** (`Ctrl + 3`).
- **Telegram**: kunci PIN (`Ctrl + L`) yang juga membuat Telegram tidak lagi terlihat online selama terkunci, **Mode Privasi** blur (`Alt + P`, tombol Privasi di dock, Shift+klik untuk ganti intensitas), notifikasi dengan ikon serta nama aplikasi berbeda dari WhatsApp, **Catatan** (`Alt + N` atau tombol di dock), **chat ke nomor baru** (`Ctrl + N`, hanya untuk akun yang mengizinkan dicari lewat nomor), dan jumlah belum dibaca yang ikut di badge taskbar dan tooltip tray.
- **Notifikasi pintar**: klik notifikasi untuk langsung membuka chatnya, tombol "Bisukan chat ini" di notifikasi, daftar chat yang dibisukan, dan **jam tenang** (misalnya 22:00–07:00). Diatur di Pusat Kontrol > Sistem dan berlaku untuk WhatsApp dan Telegram. Aplikasi mendaftarkan protokol `wadesktoplight:` (per pengguna) untuk klik notifikasi.
  Pesan yang datang bersamaan (misalnya setelah komputer bangun dari sleep) digabung menjadi satu ringkasan seperti "12 pesan baru dari 4 chat" alih-alih puluhan notifikasi beruntun; pesan baru dari chat yang sama menggantikan notifikasi lamanya, dan setiap notifikasi punya tombol **Tutup**.
- **System Tray**: tombol Close menyembunyikan aplikasi ke tray, taskbar berkedip saat ada pesan, tooltip menampilkan jumlah belum dibaca, dan tombol taskbar diberi badge angka merah.
- **Kunci saat disembunyikan** (opsional): kunci PIN otomatis saat jendela disembunyikan ke tray atau lewat Boss Key.
- **Mulai otomatis** saat Windows menyala, langsung ke tray.
- **Pembaruan otomatis**: sekali sehari aplikasi memeriksa rilis terbaru di GitHub dan menampilkan banner jika ada versi baru. **Perbarui sekarang** mengunduh `WhatsApp.exe` dari rilis repositori ini, memasangnya hanya jika SHA-256-nya cocok dengan `WhatsApp.exe.sha256` di rilis yang sama dan file itu program Windows 64-bit, lalu membuka ulang aplikasi. Versi lama disimpan sebagai `WhatsApp.exe.old` dan dikembalikan bila pemasangan gagal. Sesi login, PIN, catatan, dan pengaturan tidak tersentuh. Ada juga tombol Unduh (halaman rilis) dan Lewati versi ini; pengecekan bisa dimatikan di Pusat Kontrol > Sistem.
- **Cadangkan & pulihkan pengaturan**: satu file `.json` berisi balasan cepat, catatan, pin & blur chat, radar, tampilan, dan pengaturan notifikasi. PIN dan sesi login tidak ikut. Pusat Kontrol > Sistem.
- Link eksternal dibuka di browser default Windows.

### Diagnostik
WhatsApp Web sering mengubah struktur halamannya sehingga fitur yang bergantung pada tampilan bisa berhenti bekerja. Tab **Diagnostik** memeriksa elemen WhatsApp yang dipakai aplikasi ini (daftar chat, header, pesan, kotak ketik, dan lainnya) dan menunjukkan mana yang masih dikenali, mana yang memakai selektor cadangan, dan mana yang hilang. Semua selektor ada di satu tempat: [assets/addon-core.js](assets/addon-core.js).

## Yang tidak tersedia

Fitur berikut **tidak ada** di aplikasi ini: menyembunyikan centang biru, anti-hapus pesan (anti-revoke), menyembunyikan status mengetik atau online, dan melihat status tanpa meninggalkan jejak. Semuanya mengharuskan mengubah komunikasi dengan server WhatsApp. Itu melanggar ketentuan layanan WhatsApp dan berisiko membuat akun dibatasi atau diblokir.

## Pintasan keyboard

| Pintasan | Fungsi |
| :--- | :--- |
| `Ctrl + Alt + W` | Boss Key global (sembunyikan / munculkan) |
| `Alt + M` | Buka Pusat Kontrol |
| `Alt + P` | Mode Privasi on / off |
| `Alt + N` | Catatan tempel |
| `Alt + R` | Pengingat balas untuk chat yang terbuka |
| `Alt + V` | Picture-in-Picture video |
| `Ctrl + N` | Chat ke nomor baru |
| `Ctrl + L` | Kunci aplikasi |
| `Ctrl + 1` / `2` / `3` | Fokus WhatsApp / Telegram / tampilan berdampingan |
| `Ctrl +` / `Ctrl -` / `Ctrl 0` | Zoom in / out / reset |
| `[` / `]` | Perlambat / percepat voice note |
| `/pintasan` + Spasi | Kembangkan balasan cepat |

## Kebutuhan

- Windows 10 atau lebih baru.
- Microsoft Edge WebView2 Runtime (diunduh otomatis jika belum ada).
- Akun WhatsApp yang sudah dipasangkan ke WhatsApp Web.

## Build dari source

Pasang Go dan compiler C untuk Windows, lalu:

```powershell
go mod download
go build -ldflags="-H windowsgui -s -w" -o WhatsApp.exe .
```

### Merilis versi baru

```powershell
powershell -ExecutionPolicy Bypass -File tools\release.ps1 -Version 1.2.0
```

Skrip ini menjalankan test, membangun `WhatsApp.exe` dengan versi tertanam, dan menulis `WhatsApp.exe.sha256`. Lalu buat rilis di GitHub dengan tag `v1.2.0` (bukan draft atau pre-release) dan unggah **kedua** file. Aplikasi yang sedang dipakai akan menawarkan **Perbarui sekarang**; tanpa file `.sha256`, aplikasi hanya menawarkan tombol Unduh. Tag rilis harus sama dengan versi di exe, kalau tidak aplikasi akan terus menawarkan pembaruan yang sama.

## Struktur proyek

| File | Isi |
| :--- | :--- |
| `main.go` | Jendela WebView2, tray, hotkey, profil, tata letak WhatsApp/Telegram |
| `badge_windows.go` | Badge angka di tombol taskbar (ITaskbarList3) |
| `pinlock.go` | PIN kunci aplikasi: hash PBKDF2, hitungan PIN salah, dan jeda |
| `downloads.go` | Dialog Save As untuk unduhan dan laporan progres ke halaman |
| `notify.go` | Toast, klik notifikasi (protokol `wadesktoplight:` + `WM_COPYDATA`), jam tenang, chat dibisukan |
| `notes.go` | Penyimpanan bersama WhatsApp & Telegram: catatan dan balasan cepat |
| `updater.go`, `selfupdate.go` | Cek pembaruan harian ke GitHub Releases; unduh, verifikasi checksum, ganti exe, dan buka ulang |
| `tools/release.ps1` | Membangun rilis: `WhatsApp.exe` dengan versi tertanam dan `WhatsApp.exe.sha256` |
| `backup.go` | Ekspor & impor pengaturan |
| `reminders.go` | Pengingat balas |
| `assets/addon-main.js` | Skrip utama yang di-inject ke WhatsApp Web: privasi, balasan cepat, kunci, catatan, dock |
| `assets/addon-core.js` | Registry selektor, diagnostik, pin, blur per chat, radar, mode ringkas, ekspor |
| `assets/telegram-main.js` | Skrip yang di-inject ke Telegram Web |
| `assets/notify.js`, `assets/notes.js`, `assets/quickreplies.js`, `assets/downloads.js` | Skrip bersama WhatsApp & Telegram: polyfill notifikasi, catatan, sinkron balasan cepat, panel unduhan |
| `assets/addon.css` | Design system: token warna, Pusat Kontrol, dock, toast, mode ringkas |
| `assets/control-center.html` | Markup Pusat Kontrol |
| `assets/preview.html` | Pratinjau Pusat Kontrol di browser tanpa menjalankan aplikasi |
| `tools/dev.ps1`, `tools/smoke.js` | Menjalankan salinan dev dengan DevTools dan uji asap end-to-end |
| `vendor/github.com/jchv/go-webview2` | Library WebView2 (ditambah `SetContentBounds` untuk ukuran viewport, event unduhan `ICoreWebView2_4`, dan antrean `Dispatch` bersama agar webview Telegram yang tertanam juga menerima hasil binding) |
| `app.manifest`, `resource.rc`, `icon.ico` | Manifest, resource, dan ikon Windows |
| `WhatsApp.svg.webp`, `assets/telegram-logo.png` | Logo resmi WhatsApp dan Telegram (sumber semua ikon) |
| `tools/make_icons.py` | Membuat `icon.ico` (16–256 px) dan ikon notifikasi `assets/notif-*.png` dari logo resmi |

`icon.ico` adalah ikon aplikasi, jendela, dan tray. Semua ikon dibuat dari logo resmi dengan `python tools/make_icons.py` (butuh Pillow). Ikon ditanam ke exe lewat `rsrc.syso`; setelah mengganti `icon.ico`, buat ulang dengan `windres -O coff -i resource.rc -o rsrc.syso` (manifest ID 1, grup ikon ID 2).

Untuk melihat perubahan tampilan Pusat Kontrol tanpa membangun aplikasi, jalankan `python -m http.server` di folder `assets` lalu buka `preview.html`.

## Menguji perubahan

WhatsApp Web sering berubah, jadi ada uji asap yang menjalankan fitur-fiturnya di WhatsApp Web sungguhan:

```powershell
# 1. Keluar dari aplikasi lewat tray (salinan dev memakai profil yang sama)
powershell -ExecutionPolicy Bypass -File tools\dev.ps1
node tools\smoke.js
```

Uji ini membuka satu chat tanpa pesan belum dibaca (agar tidak mengirim centang biru), tidak pernah menekan Enter, dan hanya mengetik teks uji di chat dengan diri sendiri. Hasilnya `PASS`/`FAIL` per fitur.

## Data sesi

Sesi login (cookie, IndexedDB, token) tersimpan di:

```text
%APPDATA%\WhatsAppDesktopLight\UserData
```

Jangan hapus folder ini jika sesi login ingin dipertahankan. Menutup aplikasi tidak menghapus sesi.

## Privasi

Aplikasi ini memuat WhatsApp Web secara langsung. Data obrolan dan status autentikasi ditangani oleh WhatsApp Web dan disimpan lokal di profil WebView2 di atas. Pin, balasan cepat, dan pengaturan disimpan di penyimpanan lokal aplikasi. Proyek ini tidak berafiliasi dengan WhatsApp atau Meta.

## Lisensi

Belum ada lisensi yang dinyatakan.
