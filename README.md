# WhatsApp Webview Desktop

Lightweight Windows desktop wrapper for [WhatsApp Web](https://web.whatsapp.com), built with Go and Microsoft Edge WebView2.

## Features

- **🕵️ Mode Siluman (Ghost Mode) & Anti-Kepo**:
  - **🗑️ Anti-Tarik Pesan (*Anti-Revoke / Anti-Delete*)**: Deteksi dan pulihkan pesan yang telah dihapus atau ditarik oleh pengirim secara otomatis. Menampilkan teks asli dengan lencana khusus `[🗑️ Pesan Ditarik (Waktu)]` sehingga pesan tidak pernah hilang dari layar Anda.
  - **👀 Anti-Centang Biru (*Stealth Read Receipts*)**: Baca semua pesan chat hingga tuntas tanpa mengirimkan tanda centang biru ke pengirim (di HP pengirim tetap centang dua abu-abu). Dilengkapi tombol manual **`✓ Tandai Dibaca`** di header obrolan jika Anda ingin mengirim centang biru sewaktu-waktu.
  - **👻 Intip Status / Story Teman (*Ghost Status Viewer*)**: Tonton semua foto dan video status teman sesering mungkin tanpa meninggalkan jejak (nama Anda tidak akan pernah tercantum di daftar pemirsa/viewers mereka).
  - **🤫 Sembunyikan Status "Sedang Mengetik..." (*Hide Typing Indicator*)**: Ketik draf pesan sepanjang apa pun tanpa lawan bicara melihat status *"sedang mengetik..."* di bawah nama Anda.
  - **🎵 Dengarkan Voice Note Tanpa Ketahuan (*Stealth Voice Note Playback*)**: Putar rekaman suara/VN teman tanpa mengubah ikon mikrofon menjadi warna biru di HP pengirim (tetap hijau/abu-abu).
  - **❄️ Bekukan / Sembunyikan Status "Online" Pribadi (*Freeze / Hide Online Presence*)**: Menahan pengiriman sinyal kehadiran aktif sehingga kontak tidak dapat melihat label *"Online"* saat Anda sedang membuka aplikasi.
  - **🔔 Radar Kontak Online Real-Time (*Online Tracker / Radar*)**: Notifikasi pop-up otomatis dan Windows Toast saat kontak yang sedang dipantau baru saja membuka WhatsApp (*"🔔 Budi sedang Online!"*) atau mulai mengetik.
- **📌 Pin Chat Tanpa Batas (*Unlimited Pinned Chats*)**: Lewati batas bawaan WhatsApp (maksimal 3 obrolan). Sematkan chat sebanyak yang Anda mau di bagian atas daftar obrolan dengan indikator pin `📌` dan 1-klik pin/unpin.
- **✉️ Filter Obrolan Belum Dibaca (*Quick Unread Filter*)**: Tombol cepat satu klik di sebelah bilah pencarian untuk menyaring dan hanya menampilkan obrolan yang belum dibaca atau memerlukan balasan.
- **👥 Dual WhatsApp / Profil Ganda (*Multi-Account Profile*)**: Jalankan 2 akun WhatsApp berbeda secara bersamaan di satu komputer dalam jendela terpisah dengan data sesi independen (`UserData_Profile2`). Dapat dibuka melalui menu Control Center, klik kanan System Tray, atau parameter `--profile 2`.
- **Native WebView2**: Ultra-lightweight Windows desktop wrapper using Edge WebView2 instead of Electron (~50-80MB RAM vs 400MB+ in Electron).
- **⚡ Menu Fitur In-App (Control Center)**: Akses langsung semua fitur tambahan melalui tombol icon `⚡` di sidebar kiri WhatsApp (sejajar icon Meta AI/Pengaturan) atau tombol dock mengambang di pojok kiri bawah. Dapat juga dibuka via shortcut **`Alt + M`** atau klik kanan tray icon Windows.
- **🤫 Boss Key Global (`Ctrl + Alt + W`)**: Sembunyikan atau munculkan WhatsApp seketika dari aplikasi apa pun di Windows tanpa perlu klik mouse.
- **🚀 Mulai Otomatis saat Windows Boot (*Autostart to Tray*)**: Opsi untuk menjalankan aplikasi otomatis saat Windows menyala langsung dalam keadaan tersembunyi di System Tray.
- **⚡ Manajer Balasan Cepat Kustom (*Custom Quick Replies*)**: Buat, edit, dan hapus template pesan shortcut Anda sendiri (misal `/toko`, `/promo`, `/norek`) langsung dari Control Center. Ketik shortcut di chat lalu tekan `Space` atau `Tab` untuk memunculkan teks lengkap secara instan.
- **👁️ Kenyamanan Mata & Tema (*Eye Comfort*)**:
  - **Layar Hangat (Sepia / Anti-Silau)**: Filter lembut penahan radiasi cahaya biru agar mata tidak cepat lelah atau pusing saat menatap layar berjam-jam.
  - **OLED Pure Black**: Tema latar belakang hitam pekat murni (`#000000`) yang sangat empuk di mata dan hemat daya.
  - **Tampilan Chat Rapat (*Compact List*)**: Memadatkan baris obrolan hingga 35% lebih efisien agar dapat melihat lebih banyak kontak tanpa sering *scroll*.
- **📝 Papan Catatan Tempel (*In-App Scratchpad*) (`Alt + N`)**: Panel catatan samping di dalam WA untuk menyimpan draft pesan, nomor resi pengiriman, to-do list, atau nomor rekening sementara dengan autosave otomatis.
- **📥 Pengunduh Status / Story WhatsApp**: Tombol unduh otomatis (**`⬇️ Unduh Status`**) yang muncul saat Anda membuka status foto atau video teman.
- **🔒 Mode Tangkapan Layar Anonim (*Anonymized Screenshot*)**: Samarkan seluruh nama kontak dan foto profil dalam satu klik agar aman saat mengambil tangkapan layar bukti transfer atau komplain (`Win + Shift + S`).
- **🔵 Telegram Web & Tab Switcher (`Ctrl + 2`)**: Buka Telegram Web bersamaan dengan profil sesi terisolasi (`UserData_Telegram`). Dilengkapi dock tab mengambang di pojok kiri bawah untuk beralih instan antara WhatsApp dan Telegram.
- **◫ Mode Berdampingan / Split Screen 50:50 (`Ctrl + 3`)**: Tata jendela WhatsApp di separuh layar kiri dan Telegram di separuh layar kanan secara otomatis sesuai resolusi monitor Anda.
- **🕵️ Modular "Anti-Pusing" Privacy Mode**: Pengaturan privasi modular dengan intensitas blur yang dapat dipilih (Lembut 3px, Sedang 5px, Kuat 8px) serta pilihan granular untuk menyamarkan Nama Kontak, Preview Pesan, Balon Chat, Media, atau Foto Profil. Dilengkapi efek *hover* untuk mengintip.
- **🔐 App Lock & Ganti PIN (`Ctrl + L`)**: Kunci layar aplikasi dengan PIN (PIN bawaan: `1234`). Kunci otomatis setelah 5 menit tidak aktif. PIN dapat diganti langsung kapan saja melalui tautan **"⚙️ Ganti / Ubah PIN"** di layar kunci, dari menu Control Center, atau dari System Tray menu.
- **💾 Status & Manajemen Sesi Login**: Sesi login (cookies, IndexedDB, token) tersimpan permanen di `%APPDATA%\WhatsAppDesktopLight\UserData` sehingga tidak perlu scan QR berulang kali. Tersedia indikator status sesi dan tombol aman untuk **Reset / Logout Sesi** jika ingin login ulang dari awal.
- **🚀 Chat Langsung Tanpa Simpan Nomor (`Ctrl + N`)**: Kirim pesan WhatsApp ke nomor baru tanpa perlu menyimpan ke kontak HP. Format nomor otomatis dinormalisasi (misal `0812...` otomatis menjadi `62812...`).
- **🎵 Pengatur Kecepatan Voice Note (`[` / `]`)**: Atur kecepatan pemutaran audio/voice note dari 0.5x, 1.0x, 1.5x, 2.0x hingga 3.0x dengan indikator visual.
- **📺 Picture-in-Picture untuk Video (`Alt + V`)**: Tonton video yang dikirim di chat dalam jendela mengambang (*floating window*) sembari tetap membaca dan membalas pesan.
- **🔍 Pengatur Skala Tampilan / Zoom (`Ctrl +` / `Ctrl -` / `Ctrl 0`)**: Sesuaikan ukuran font dan elemen antarmuka dari 75% hingga 150% dengan memori otomatis.
- **System Tray & Notifikasi Pintar**: Tombol Close (`X`) meminimalkan aplikasi ke Tray latar belakang. Dilengkapi kedipan taskbar oranye saat ada pesan masuk dan penghitung pesan belum dibaca di tooltip.
- **Keamanan & Isolasi Link**: Link eksternal otomatis dibuka di browser default Windows (Edge/Chrome/Firefox) agar sesi akun tetap aman.

## Keyboard Shortcuts

| Shortcut | Action |
| :--- | :--- |
| `Ctrl + Alt + W` | **Boss Key Global (Sembunyikan / Munculkan WA & Telegram seketika dari mana saja)** |
| `Alt + M` | **Buka Menu Fitur & Control Center In-App** |
| `Ctrl + 1` | **Fokus ke Jendela WhatsApp** |
| `Ctrl + 2` | **Buka / Fokus ke Jendela Telegram Web** |
| `Ctrl + 3` | **Tata Berdampingan (Split View 50:50 WhatsApp & Telegram)** |
| `Alt + N` | **Buka / Tutup Papan Catatan Tempel (Scratchpad)** |
| `Alt + P` | Toggle Cepat Mode Privasi (On / Off) |
| `Ctrl + N` | Dialog Chat ke nomor baru tanpa simpan kontak |
| `Ctrl + L` | Kunci WhatsApp dengan PIN (Default: `1234`) |
| `]` | Percepat pemutaran voice note (+0.25x hingga 3.0x) |
| `[` | Perlambat pemutaran voice note (-0.25x hingga 0.5x) |
| `Alt + V` | Toggle Picture-in-Picture (PiP) untuk video |
| `Ctrl + +` / `Ctrl + =` | Perbesar tampilan WhatsApp (Zoom in hingga 150%) |
| `Ctrl + -` | Perkecil tampilan WhatsApp (Zoom out hingga 75%) |
| `Ctrl + 0` | Reset zoom tampilan ke 100% |
| `/shortcut` + Space | Canned reply: Ekspansi template balasan cepat kustom |


## Requirements

- Windows 10 or newer.
- Microsoft Edge WebView2 Runtime. The app can download it automatically when missing.
- WhatsApp account paired with WhatsApp Web.

## Download

Download `WhatsApp.exe` from the latest [GitHub Release](https://github.com/Adytm404/whatsapp-web.view/releases).

Run the executable, scan the QR code, allow camera and microphone access when prompted by Windows, and keep using WhatsApp normally. The session is saved automatically.

## Session Data

Profile data is stored at:

```text
%APPDATA%\WhatsAppDesktopLight\UserData
```

Do not delete this folder if the existing login session must remain available. Closing the app does not clear session data.

## Build From Source

Install Go and a Windows C compiler, then run:

```powershell
go mod download
go build -ldflags="-H windowsgui -s -w" -o WhatsApp.exe .
```

The repository includes the Windows manifest and embedded icon resource used by the build.

## Project Files

- `main.go`: WebView2 window, persistent profile, dark frame, and notifications.
- `app.manifest`: Windows DPI and application manifest.
- `resource.rc`: Windows icon and manifest resource definitions.
- `icon.ico`: Application icon.

## Privacy

This app loads WhatsApp Web directly. Chat data and authentication state are handled by WhatsApp Web and stored locally in the WebView2 profile above. This project is not affiliated with WhatsApp or Meta.

## License

No license has been declared yet.
