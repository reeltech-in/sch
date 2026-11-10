# SCH — Sistem Informasi Sekolah (Go + SQLite)

Satu aplikasi untuk seluruh kebutuhan sekolah, berjalan di **PC server sekolah sendiri**, tanpa
internet dan tanpa biaya hosting. Satu file program, data tersimpan di satu file database.
Satu login dan satu data siswa/guru/kelas dipakai bersama oleh semua modul.

| Modul | Isi |
|---|---|
| **Pembelajaran (LMS)** | materi (teks, video, tautan, gambar), tugas & penilaian, kuis, absensi per mapel, rekap nilai |
| **Buku Induk** | biodata lengkap siswa & orang tua, foto, cetak per siswa |
| **Absensi Harian** | scan kartu QR di gerbang (scanner USB / kamera HP), ceklist per kelas, absen selfie + lokasi, rekap bulanan cetak/CSV |
| **Kartu Siswa** | kartu QR bertanda tangan digital (tidak bisa dipalsukan), cetak 10 kartu per A4 |
| **Kalender Akademik** | agenda sekolah; hari libur otomatis ditandai di rekap absensi |
| **Ujian CBT** | soal bergambar, acak soal & pilihan, token, jadwal buka–tutup, simpan otomatis, impor soal dari teks |
| **Manajemen Ujian** | ruang, bagi peserta otomatis (silang kelas), cetak kartu peserta, denah, daftar hadir, berita acara |
| **Kelulusan** | pengumuman online dengan NISN + tanggal lahir, terbuka otomatis pada waktu yang ditentukan |
| **Inventaris Sarpras** | KIB A–F lengkap, foto, kondisi, nilai aset, cetak KIB & label QR |
| **Perpustakaan** | katalog, pinjam/kembali dengan scan kartu & label buku, denda otomatis, rak **E-book**: unggah PDF sendiri atau impor gratis dari Buku Kemendikbud (±1.600 judul; PDF disimpan di server saat pertama dibaca) |
| **Asisten AI** | generator RPM, LKPD, soal (langsung masuk CBT/Kuis), surat, tanya bebas; penyedia apa pun yang kompatibel OpenAI |
| **PPDB Online** | formulir publik + unggah berkas, nomor & PIN cek status, verifikasi, "jadikan siswa" sekali klik |
| **Buku Tamu** | mode kiosk untuk tablet di meja piket, foto opsional, rekap & CSV |
| **Surat Masuk/Keluar** | agenda, arsip scan, disposisi, nomor surat keluar otomatis, cetak agenda & lembar disposisi |
| **E-Voting OSIS** | suara rahasia, memilih dari HP atau bilik suara dengan scan kartu, hasil langsung |
| **Kasir / Koperasi** | scan barcode, struk 58 mm, stok otomatis, laporan omzet & laba |

Admin bisa menyalakan/mematikan modul di **Pengaturan → Modul aplikasi**.
**Tugas tambahan guru/TU:** di form akun guru, centang modul yang boleh dikelola (misalnya Perpustakaan
untuk pustakawan, Surat & PPDB untuk TU, Kasir untuk pengelola koperasi). Mereka mendapat menu dan hak kelola
modul itu saja, tanpa menjadi admin penuh.

Kebutuhan server: PC/laptop apa saja (Windows 7+ atau Linux), RAM 1 GB sudah cukup.
Server memakai sekitar 10–50 MB RAM. Raspberry Pi juga bisa.

---

## 1. Menjalankan di PC server

### Windows
1. Salin ke satu folder, misalnya `C:\SCH`:
   - `sch-server-windows-amd64.exe` (Windows 64-bit) atau `sch-server-windows-386.exe` (Windows 32-bit lama)
   - `jalankan.bat`
2. Klik dua kali `jalankan.bat`.
3. Saat pertama kali dijalankan muncul **username & password admin**. Password ini juga disimpan di
   `data\ADMIN-PASSWORD.txt`. Login, ganti password di menu **Akun Saya**, lalu hapus file itu.
4. Jika Windows Firewall bertanya, pilih **Allow / Izinkan** untuk jaringan *Private*.

### Server Ubuntu / Debian (otomatis, disarankan)
Satu perintah memasang dependensi, Go, build, user `sch`, dan layanan systemd:
```bash
curl -fsSL https://raw.githubusercontent.com/reeltech-in/sch/main/install.sh | sudo bash
```
Repo privat, pengaturan port/proxy, update, dan uninstall: lihat [docs/INSTALL-UBUNTU.md](docs/INSTALL-UBUNTU.md).

### Linux / Raspberry Pi (manual)
```bash
chmod +x sch-server-linux-amd64        # Raspberry Pi 64-bit: sch-server-linux-arm64, 32-bit: sch-server-linux-arm
./sch-server-linux-amd64 -addr :3220
```

## 2. Membuka dari HP / laptop
Saat server berjalan, jendela hitam menampilkan alamatnya, contoh:
```
http://192.168.1.10:3220     (pemakaian biasa)
https://192.168.1.10:3221    (untuk kamera HP: scan QR & absen selfie)
```
Sambungkan HP ke WiFi sekolah yang sama, lalu buka alamat itu di browser.

**Kamera HP butuh alamat `https://`.** Saat pertama membuka alamat https, browser memberi peringatan
"koneksi tidak pribadi" karena sertifikatnya dibuat sendiri oleh server sekolah. Pilih
**Lanjutkan / Advanced → Proceed**. Ini aman selama di jaringan sekolah sendiri.

**Tanpa kamera:** scanner barcode USB (± Rp150 ribu) bisa dipakai di laptop petugas untuk absensi dan
perpustakaan. Cukup klik kotak scan, lalu arahkan scanner ke kartu, tidak perlu https.

**Tips:** atur IP server agar tetap (*DHCP reservation* di router) supaya alamatnya tidak berubah.

## 3. Online lewat Cloudflare Zero Trust Tunnel
Tunnel membuat aplikasi bisa dibuka dari internet (mis. `https://lms.sekolahanda.sch.id`) **tanpa IP publik
dan tanpa membuka port di router**. HTTPS diurus Cloudflare, jadi kamera HP langsung bisa dipakai.

1. Jalankan server hanya untuk tunnel, dengan pilihan `-proxy` (Windows: klik `jalankan-online.bat`):
   ```bash
   ./sch-server-linux-amd64 -addr 127.0.0.1:3220 -https "" -proxy
   ```
   `-proxy` **wajib** di belakang tunnel: server membaca IP asli pengunjung dari header Cloudflare.
   Tanpa ini semua pengunjung terlihat dari 127.0.0.1, sehingga satu orang yang salah password bisa membuat
   semua orang terblokir login. `-proxy` saja hanya mempercayai proxy di komputer yang sama; bila cloudflared/nginx
   berjalan di mesin lain atau di Docker, pakai `-proxy-ip <IP mesin itu>` (lihat di bawah).
2. Di Cloudflare: **Zero Trust → Networks → Tunnels → Create tunnel (Cloudflared)**, pasang `cloudflared` di
   PC server sesuai perintah yang ditampilkan.
3. Tambahkan **Public hostname**: subdomain pilihan Anda → Service `HTTP` → `localhost:3220`.
4. (Opsional, disarankan) **Zero Trust → Access → Applications**: lindungi path `/api/backup` atau seluruh
   aplikasi untuk admin dengan login email. Halaman publik (`/#/ppdb`, `/#/kelulusan`, `/#/bukutamu`) tetap
   terbuka bila diperlukan.

**Bila cloudflared berjalan di komputer/VM lain** (mis. VM tunnel terpisah): jalankan SCH di semua alamat dan
izinkan hanya IP VM tunnel sebagai proxy, lalu arahkan Public hostname ke `http://IP-SERVER-SCH:3220`:
```bash
./sch-server-linux-amd64 -addr :3220 -https :3221 -proxy-ip 10.32.10.5     # ganti dengan IP VM cloudflared
```
`-proxy-ip` membuat header Cloudflare (IP asli, HTTPS) hanya dipercaya bila datang dari VM tunnel, sehingga
pengguna LAN yang mengakses langsung tidak bisa memalsukan IP untuk menghindari pembatas login.

Di jaringan sekolah, aplikasi tetap bisa dibuka langsung lewat IP server bila server dijalankan dengan
`-addr :3220` (tanpa `127.0.0.1`).

## 4. Asisten AI
Menu **Asisten AI → Pengaturan** (admin):
1. Isi **URL endpoint** yang kompatibel OpenAI (ada contoh: OpenAI, Gemini, Anthropic, OpenRouter, Ollama lokal).
2. Tempel **API key**. Key disimpan di server sekolah dan tidak pernah dikirim ke browser pengguna.
3. Klik **Refresh model**: daftar model diambil langsung dari endpoint `/models` penyedia, pilih dari dropdown.
4. Atur **kuota per guru per hari** untuk mengendalikan biaya.

Hasil AI selalu perlu diperiksa guru sebelum dipakai. Biaya API ditagih oleh penyedia yang Anda pilih.

## 5. Urutan persiapan (admin)
1. **Pengaturan** — nama sekolah, logo, kepala sekolah, tahun ajaran, modul yang dipakai, jam masuk.
2. **Kelas** → **Guru** → **Siswa** (tambah satu per satu atau **Impor CSV**: `nama, nis, password, kelas`).
   Kolom password boleh dikosongkan: sistem membuat password acak dan daftar akunnya bisa diunduh untuk
   dicetak sebagai slip login. Akun baru, hasil impor, atau yang direset admin **wajib ganti password** saat
   pertama login.
3. **Mata Pelajaran** — per kelas, pilih guru pengampu.
4. **Buku Induk** — lengkapi NISN & tanggal lahir (dibutuhkan untuk cek kelulusan).
5. **Kartu Siswa** — cetak, potong, laminating.
6. **Absen selfie** (opsional) — Pengaturan → Absensi harian → centang selfie, isi lokasi sekolah
   (tombol "Isi dengan lokasi perangkat ini" saat berada di sekolah) dan radius.

### Tahun ajaran & kenaikan kelas
Menu **Tahun Ajaran** (admin). Setiap kelas, mapel, nilai, dan absensi tersimpan per tahun ajaran. Di akhir tahun:
**Kenaikan kelas** → atur kelas tujuan (otomatis disarankan, mis. 7A → 8A), tandai siswa yang tinggal kelas,
lulus, atau keluar → jalankan. Tahun baru langsung aktif; data tahun lalu tetap bisa dibuka lewat pilihan tahun di
halaman Kelas, Mapel, dan Mapel Saya (guru).

### Pemeliharaan otomatis
Setiap hari (setelah pukul 01.00) server membuat backup database, menghapus foto selfie/buku tamu yang melewati
masa simpan, dan membersihkan file upload yang tidak dipakai. Atur di **Pengaturan → Penyimpanan & backup**, yang
juga menampilkan pemakaian disk dan daftar backup untuk diunduh.

## 6. Pilihan saat menjalankan
| Pilihan | Arti | Bawaan |
|---|---|---|
| `-addr :3220` | port HTTP (pakai `:80` agar cukup ketik IP tanpa `:3220`) | `:3220` |
| `-https :3221` | port HTTPS untuk kamera HP; isi kosong `-https ""` untuk mematikan | `:3221` |
| `-proxy` | wajib bila online lewat Cloudflare Tunnel / reverse proxy **di komputer yang sama** (hanya 127.0.0.1 dipercaya) | mati |
| `-upload-max-gb 5` | batas total ukuran file upload (foto, berkas PPDB, e-book) | `5` |
| `-proxy-ip 10.0.0.5` | proxy/cloudflared di **mesin lain** (VM, Docker 172.17.0.0/16): percayai header proxy dari IP/CIDR ini (otomatis menyalakan `-proxy`) | – |
| `-db data/sch.db` | lokasi file database (folder ini juga menyimpan upload & sertifikat) | `data/sch.db` |
| `-reset-admin` | buat password admin baru lalu keluar (jika lupa password) | – |
| `-demo` | isi data contoh untuk mencoba (hanya jika database kosong) | – |

Mencoba dengan data contoh, memakai database terpisah:
```bash
sch-server-windows-amd64.exe -db demo/sch.db -demo
```
Akun contoh: guru `guru001` / `guru123`, siswa `siswa001` / `siswa123`. Token ujian contoh: `MTK26`.

## 7. Backup & pemulihan
Semua data ada di folder `data/`: `sch.db` (database) dan `uploads/` (foto, gambar soal, e-book).
- **Backup database**: menu **Pengaturan → Unduh backup**, minimal seminggu sekali.
- **Backup lengkap**: salin seluruh folder `data/` ke flashdisk saat server dimatikan.
- **Pulihkan**: matikan server, kembalikan folder `data/` (atau ganti `data/sch.db` dengan file backup
  dan hapus `sch.db-wal` serta `sch.db-shm`), lalu jalankan lagi.

## 8. Keamanan
- Password disimpan sebagai hash bcrypt; login dibatasi 10 percobaan gagal per 10 menit per perangkat.
- Semua aturan akses diperiksa di server: siswa hanya melihat kelasnya, guru hanya mapel/ujian miliknya.
  Modul yang dimatikan juga ditolak di server.
- Kunci jawaban kuis/CBT tidak pernah dikirim ke siswa; nilai dihitung di server; waktu ujian dihitung server.
- Kartu QR siswa ditandatangani (HMAC) sehingga tidak bisa dibuat sendiri dari NIS teman.
- File upload diperiksa jenis aslinya (JPG/PNG/WEBP/GIF/PDF) dan hanya bisa dibuka pengguna yang login.
- Formulir publik (PPDB, buku tamu, cek kelulusan) dibatasi per IP, memakai honeypot anti-bot, dan berkas
  dibatasi jenis & ukurannya.
- E-Voting: siapa yang sudah memilih dicatat terpisah dari pilihannya, sehingga pilihan tidak bisa dilacak.
- Hasil audit keamanan sudah diterapkan: batas waktu CBT/kuis ditegakkan server, pengaturan AI khusus admin
  (key tidak bisa dikirim ke endpoint lain), perolehan suara e-voting dirahasiakan sampai selesai, guru hanya
  melihat data siswa di kelas yang diajar, penghapusan yang merusak arsip diblokir (gunakan status Nonaktif),
  pembatas login ramah NAT sekolah, CSP, file data hanya bisa dibaca pemilik proses, dan ekspor CSV aman dari
  sisipan rumus.
- Untuk online, gunakan Cloudflare Tunnel (bagian 3) dengan `-proxy`. Jangan membuka port server langsung ke internet.

## 9. Database: kenapa SQLite?
Untuk satu sekolah (ratusan sampai beberapa ribu siswa, CBT ratusan siswa bersamaan) SQLite dalam mode WAL
sudah lebih dari cukup, tidak perlu memasang server database, dan backup cukup menyalin file.
PostgreSQL/MySQL baru diperlukan bila satu server melayani banyak sekolah (tingkat dinas) atau dipindah ke
cloud; kode SQL-nya standar sehingga bisa dipindahkan.

## 10. Rilis & versi
Setiap rilis punya nomor versi (`MAYOR.MINOR.PERBAIKAN`) yang tampil di menu kiri aplikasi, di
`/api/version`, di header `X-SCH-Version`, dan lewat `sch-server -version`. Riwayatnya ada di `CHANGELOG.md`.

Alur rilis ke server:
1. Naikkan angka di file `VERSION` (mis. `1.4.0` → `1.4.1` untuk perbaikan, `1.5.0` untuk fitur baru).
2. Tulis catatannya di `CHANGELOG.md` dengan judul `## v1.4.1 — tanggal`.
3. Jalankan `./deploy.sh`. Skrip akan: build dengan versi tertanam → backup database ke
   `/opt/sch/data/backups/sebelum-v<versi>-<waktu>.db` → simpan rilis di `/opt/sch/releases/` → restart →
   cek `/api/version`. Bila versi baru tidak merespons, otomatis kembali ke rilis sebelumnya.
4. Kembali ke rilis lama secara manual: `./rollback.sh` (lihat daftar) lalu `./rollback.sh 1.4.0`.
   Database tidak ikut dikembalikan; gunakan file backup bila perlu.

`./build.sh` membuat semua file program (Windows/Linux/Raspberry Pi) dengan nomor versi di folder `dist/`.

## 11. Untuk pengembang
```bash
go run . -db demo-data/sch.db -demo -addr :8090 -https :8493     # versi pengembangan
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/sch-server-windows-amd64.exe .
```
Struktur (satu file per fitur; setiap modul mendaftarkan skema & rutenya sendiri di `init()`):

**Inti server**
- `main.go` — titik mulai: opsi, membuka database, menjalankan server & berhenti rapi
- `app.go` — `App`, `User`, pekerja latar (tahan panic, berhenti saat shutdown)
- `handler.go` — kerangka handler: sesi/peran/CSRF, log aktivitas, galat JSON · `auth.go` — login, logout, password
- `limiter.go` — pembatas percobaan login · `static.go` — file tampilan, data awal di index.html, header keamanan
- `db.go`, `migrate.go`, `indexes.go` — skema inti, migrasi bernomor versi, indeks
- `modules.go` — registri modul, konfigurasi (kv), upload, QR, kartu · `ops.go` — zona waktu, bcrypt, log aktivitas
- `maintenance.go` — backup harian, masa simpan, file yatim (ke folder sampah) · `guards.go` — penjaga data ganda/tidak cocok
- `tls.go` — sertifikat HTTPS lokal · `demo.go` — data contoh

**Data sekolah & LMS**
- `api.go` (pengaturan, dasbor, kelas) · `api_users.go` (akun & impor) · `mod_tahun.go` (tahun ajaran, kenaikan kelas)
- `api_lms.go` (mapel, materi, tugas) · `api_kuis.go` (kuis) · `api_absensi_mapel.go` (absensi mapel, rekap nilai)
- `mapel.go` (pasangan kode–nama mapel) · `questions.go` (impor soal sekaligus) · `saran.go` (saran isian)

**Modul**
- Kesiswaan: `mod_bukuinduk.go` (Buku Induk, kartu QR), `mod_absensi.go` (absensi harian), `mod_kalender.go`
- Ujian: `mod_ujian.go` (CBT), `mod_ujian_event.go` (manajemen ujian), `mod_kelulusan.go`, `mod_ljk.go` (LJK kertas)
- Sarana: `mod_sarpras.go` (inventaris KIB & perpustakaan)
- Layanan: `mod_ppdb.go`, `mod_bukutamu.go`, `mod_surat.go`, `mod_evoting.go`, `mod_kasir.go`
  (`mod_layanan.go` & `mod_osis_kasir.go` hanya berisi skema & pendaftaran rutenya)
- `mod_admguru.go` (jurnal, catatan siswa, wali kelas) · `mod_ai.go` (Asisten AI) · `mod_wa.go` (notifikasi WhatsApp)

**Tampilan (`web/`, tanpa library luar, tertanam di program)**
- `app.js` — inti: API & cache, formulir, navigasi, pemuat modul, helper bersama · `theme.js` — tema sebelum halaman digambar
- `pages-admin.js`, `pages-guru.js`, `pages-siswa.js` — halaman per peran
- `mod-*.js` — satu file per modul; hanya dimuat bila perannya cocok & modulnya aktif (lihat `MOD_FILES` di `app.js`)
- `docs/design-system.md` — design system tampilan

Tes: `go test ./...` (dijalankan otomatis oleh `deploy.sh` dan `build.sh` sebelum rilis).

Menambah modul baru: buat `mod_xxx.go` (daftarkan skema & rute di `init()`, kunci modul di `moduleKeys`), lalu buat
`web/mod-xxx.js` yang memanggil `addNav()` dan mengisi `PAGES`, dan daftarkan file itu di `MOD_FILES` (`web/app.js`).
