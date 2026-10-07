# WhatsApp Webview Desktop

Pembungkus desktop ringan untuk [WhatsApp Web](https://web.whatsapp.com) di Windows, dibuat dengan Go dan Microsoft Edge WebView2 (sekitar 50-80 MB RAM, bukan 400 MB+ seperti Electron).

Semua fitur tambahan hanya mengubah tampilan di layar Anda. Tidak ada yang memodifikasi koneksi atau enkripsi WhatsApp, sehingga akun tidak berisiko diblokir karena fitur-fitur ini.

## Fitur

### Pusat Kontrol (`Alt + M`)
Panel pengaturan dengan lima bagian: **Privasi**, **Tampilan**, **Chat & Alat**, **Sistem**, dan **Diagnostik**. Bisa dibuka dari ikon di sidebar WhatsApp, dock mengambang di kiri bawah, atau klik kanan ikon tray.

### Privasi
- **Mode Privasi / blur layar** (`Alt + P`): buramkan nama kontak, pratinjau pesan, balon obrolan, foto & video, dan foto profil. Intensitas Halus / Sedang / Kuat. Arahkan kursor untuk mengintip.
- **Screenshot Anonim**: samarkan nama dan foto profil dengan satu klik sebelum mengambil tangkapan layar.
- **Kunci Aplikasi** (`Ctrl + L`): kunci dengan PIN (awal: `1234`), terkunci otomatis setelah 5 menit tidak aktif. PIN bisa diganti dari Pusat Kontrol, layar kunci, atau menu tray.
- **Boss Key global** (`Ctrl + Alt + W`): sembunyikan atau munculkan jendela dari aplikasi mana pun.

### Chat & Alat
- **Pin chat tanpa batas**: arahkan kursor ke chat lalu klik ikon pin. Chat yang dipin muncul sebagai chip di atas daftar chat. Pin hanya tersimpan di komputer ini dan tidak mengubah pin bawaan WhatsApp.
- **Filter chat belum dibaca**.
- **Radar online**: toast dan notifikasi Windows saat kontak di chat yang sedang terbuka menjadi online atau mulai mengetik. Bisa dibatasi ke nama tertentu. *Catatan: WhatsApp hanya mengirim status online untuk chat yang sedang Anda buka, dan hanya jika kontak mengizinkannya.*
- **Balasan cepat**: buat pintasan seperti `/toko`, ketik di chat, lalu tekan Tab, Enter, atau Spasi untuk mengembangkannya.
- **Chat ke nomor baru** (`Ctrl + N`): tanpa menyimpan kontak. `0812…` otomatis menjadi `62812…`.
- **Catatan tempel** (`Alt + N`): panel memo samping dengan simpan otomatis.
- **Unduh status**: tombol unduh muncul saat Anda membuka status foto atau video.

### Tampilan
- Tema **Standar gelap**, **OLED hitam**, dan **Layar hangat**.
- **Daftar chat rapat**.
- **Zoom** 75-150% (`Ctrl +` / `Ctrl -` / `Ctrl 0`).
- **Kecepatan voice note** 0.5×-3.0× (`[` dan `]`).
- **Picture-in-Picture** untuk video (`Alt + V`).

### Sistem
- **Dua akun WhatsApp** dalam jendela terpisah (`--profile 2`).
- **Telegram Web** berdampingan dengan WhatsApp (`Ctrl + 2`), serta **tampilan berdampingan 50:50** (`Ctrl + 3`).
- **System Tray**: tombol Close menyembunyikan aplikasi ke tray, taskbar berkedip saat ada pesan, tooltip menampilkan jumlah belum dibaca.
- **Mulai otomatis** saat Windows menyala, langsung ke tray.
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

## Struktur proyek

| File | Isi |
| :--- | :--- |
| `main.go` | Jendela WebView2, tray, hotkey, profil, dan script yang di-inject |
| `assets/addon-core.js` | Registry selektor, diagnostik, pin chat, radar online |
| `assets/addon.css` | Design system: token warna, Pusat Kontrol, dock, toast |
| `assets/control-center.html` | Markup Pusat Kontrol |
| `assets/preview.html` | Pratinjau Pusat Kontrol di browser tanpa menjalankan aplikasi |
| `app.manifest`, `resource.rc`, `icon.ico` | Manifest, resource, dan ikon Windows |

Untuk melihat perubahan tampilan Pusat Kontrol tanpa membangun aplikasi, jalankan `python -m http.server` di folder `assets` lalu buka `preview.html`.

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
