# SCH — Sistem Informasi Sekolah

**Aplikasi sekolah lengkap yang gratis, terbuka, dan berjalan di server sekolah sendiri.**
LMS, CBT, e-Rapor Kurikulum Merdeka, absensi QR, perpustakaan, PPDB, sampai kasir koperasi, dalam satu program.

- **Tanpa internet & tanpa biaya hosting**: cukup satu PC/laptop di sekolah; guru dan siswa membuka lewat WiFi.
- **Satu file program, satu file database**: tidak perlu memasang web server, PHP, atau MySQL. Backup cukup salin folder.
- **Ringan**: RAM 10–50 MB, jalan di Windows 10/11, Linux, dan Raspberry Pi.
- **Satu login untuk semua modul**: data siswa, guru, dan kelas diisi sekali, dipakai di semua modul.
- **Data milik sekolah**: tidak dikirim ke pihak mana pun (kecuali Asisten AI bila sekolah mengaktifkannya).

**[Unduh](https://github.com/reeltech-in/sch/releases/latest)** ·
**[Pasang di Ubuntu/Debian](docs/INSTALL-UBUNTU.md)** ·
**[Pasang di Windows](docs/INSTALL-WINDOWS.md)** ·
**[Catatan rilis](CHANGELOG.md)** ·
Lisensi **[AGPL-3.0](LICENSE)**

### Coba dalam satu menit
Server Ubuntu/Debian dengan data contoh:
```bash
curl -fsSL https://raw.githubusercontent.com/reeltech-in/sch/main/install.sh | sudo SCH_DEMO=1 bash
```
Windows: unduh `SCH-v…-windows.zip` dari [Releases](https://github.com/reeltech-in/sch/releases/latest), ekstrak, lalu
jalankan di Command Prompt dari folder itu: `sch-server-windows-amd64.exe -db demo\sch.db -demo`.
Buka `http://localhost:3220`. Akun contoh: guru `guru001` / `guru123`, siswa `siswa001` / `siswa123`
(password admin tampil di layar). **Jangan pakai `-demo` untuk data sungguhan.**

## Fitur

| Modul | Isi |
|---|---|
| **Pembelajaran (LMS)** | materi (teks, video, tautan, gambar), tugas & penilaian, kuis, absensi per mapel, rekap nilai |
| **e-Rapor** | rapor Kurikulum Merdeka per semester: nilai akhir otomatis (bisa disesuaikan) & capaian kompetensi per mapel dengan saran dari nilai atau Asisten AI, kokurikuler 8 dimensi profil lulusan, ekstrakurikuler (pembina), ketidakhadiran dari absensi harian, catatan wali, keputusan naik kelas; terbitkan per kelas, cetak A4 & unduh Word, ekspor Excel (leger + nilai per mapel) untuk dipindahkan ke e-Rapor resmi |
| **Portofolio Siswa** | siswa mengunggah karya (gambar/PDF/tautan) per mapel; guru mapel & wali kelas memberi catatan dan nilai; tampil di rapor mapel |
| **Buku Induk** | biodata lengkap siswa & orang tua, foto, cetak per siswa |
| **Absensi Harian** | scan kartu QR di gerbang (scanner USB / kamera HP), ceklist per kelas, absen selfie + lokasi, rekap bulanan cetak/CSV |
| **Kartu Siswa** | kartu QR bertanda tangan digital (tidak bisa dipalsukan), cetak 10 kartu per A4 |
| **Kalender Akademik** | agenda sekolah; hari libur otomatis ditandai di rekap absensi |
| **Ujian CBT** | soal bergambar, acak soal & pilihan, token, jadwal buka–tutup, simpan otomatis, impor soal dari teks |
| **Manajemen Ujian** | ruang, bagi peserta otomatis (silang kelas), cetak kartu peserta, denah, daftar hadir, berita acara |
| **Kelulusan** | pengumuman online dengan NISN + tanggal lahir, terbuka otomatis pada waktu yang ditentukan |
| **Inventaris Sarpras** | KIB A–F lengkap, foto, kondisi, nilai aset, cetak KIB & label QR |
| **Perpustakaan** | katalog, pinjam/kembali dengan scan kartu & label buku, denda otomatis, rak **E-book**: unggah PDF sendiri atau impor gratis dari Buku Kemendikbud (±1.600 judul; PDF disimpan di server saat pertama dibaca) |
| **Asisten AI** | RPP/RPM/Modul Ajar Kurikulum Merdeka terbaru (cek alokasi waktu, sunting, cetak resmi berkop & tanda tangan, unduh Word), LKPD, soal (langsung masuk CBT/Kuis), surat, tanya bebas; penyedia apa pun yang kompatibel OpenAI |
| **PPDB Online** | formulir publik + unggah berkas, nomor & PIN cek status, verifikasi, "jadikan siswa" sekali klik |
| **Buku Tamu** | mode kiosk untuk tablet di meja piket, foto opsional, rekap & CSV |
| **Surat Masuk/Keluar** | agenda, arsip scan, disposisi, nomor surat keluar otomatis, cetak agenda & lembar disposisi |
| **E-Voting OSIS** | suara rahasia, memilih dari HP atau bilik suara dengan scan kartu, hasil langsung |
| **Kasir / Koperasi** | scan barcode, struk 58 mm, stok otomatis, laporan omzet & laba |

Admin bisa menyalakan/mematikan modul di **Pengaturan → Modul aplikasi**.
**Tugas tambahan guru/TU:** di form akun guru, centang modul yang boleh dikelola (misalnya Perpustakaan
untuk pustakawan, Surat & PPDB untuk TU, Kasir untuk pengelola koperasi). Mereka mendapat menu dan hak kelola
modul itu saja, tanpa menjadi admin penuh.

Kebutuhan server: PC/laptop apa saja (Windows 10/11, Windows Server 2016+, atau Linux), RAM 1 GB sudah cukup.
Server memakai sekitar 10–50 MB RAM. Raspberry Pi juga bisa.

---

## 1. Menjalankan di PC server

Semua file program ada di halaman **[Releases](https://github.com/reeltech-in/sch/releases/latest)**.

### Server Ubuntu / Debian (disarankan)
Satu perintah memasang dependensi, Go, build, user `sch`, dan layanan systemd yang menyala otomatis:
```bash
curl -fsSL https://raw.githubusercontent.com/reeltech-in/sch/main/install.sh | sudo bash
```
Pengaturan port/proxy, memperbarui, dan menghapus: lihat [docs/INSTALL-UBUNTU.md](docs/INSTALL-UBUNTU.md).

### Windows
1. Unduh `SCH-v…-windows.zip` dari Releases dan ekstrak ke satu folder, misalnya `C:\SCH`. Isinya:
   `sch-server-windows-amd64.exe` (Windows 64-bit), `sch-server-windows-386.exe` (Windows 10 32-bit),
   `jalankan.bat`, dan `jalankan-online.bat`.
2. Klik dua kali `jalankan.bat`.
3. Saat pertama kali dijalankan muncul **username & password admin**. Password ini juga disimpan di
   `data\ADMIN-PASSWORD.txt`. Login, ganti password di menu **Akun Saya**, lalu hapus file itu.
4. Jika Windows Firewall bertanya, pilih **Allow / Izinkan** untuk jaringan *Private*.

Panduan lengkap, termasuk agar server menyala otomatis saat PC dihidupkan:
[docs/INSTALL-WINDOWS.md](docs/INSTALL-WINDOWS.md).

### Linux lain / Raspberry Pi (manual)
Unduh program yang sesuai dari Releases (`linux-amd64` untuk PC, `linux-arm64` untuk Raspberry Pi 64-bit,
`linux-arm` untuk 32-bit), lalu:
```bash
chmod +x sch-server-v*-linux-amd64
```
```bash
./sch-server-v*-linux-amd64 -addr :3220
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
Menu **Pengaturan → Asisten AI** (admin):
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

### e-Rapor (per semester)
Menu **e-Rapor**. Atur dulu tanggal akhir semester ganjil di menu **Tahun Ajaran** (bawaan 31 Desember), dan daftar
ekstrakurikuler beserta pembinanya (minimal Pramuka) di tab **Pengaturan** e-Rapor. Lalu:
1. **Guru mapel** (tab Nilai mapel): nilai awal terisi otomatis dari rumus 40% tugas + 40% kuis + 20% kehadiran
   semester itu; sesuaikan bila perlu, isi capaian tertinggi & yang perlu ditingkatkan (tombol "Isi dari nilai
   otomatis" atau "Saran deskripsi AI"), lalu **Simpan semua**.
2. **Pembina ekstrakurikuler** (tab Ekstrakurikuler): atur anggota, isi predikat & keterangan.
3. **Wali kelas** (tab Wali kelas): per siswa isi kokurikuler (8 dimensi), catatan, koreksi ketidakhadiran bila
   perlu, dan keputusan naik/tinggal/lulus di semester genap. Setelah semua mapel lengkap → **Terbitkan rapor**:
   siswa bisa melihat & mengunduh rapornya, dan isinya terkunci (batalkan terbit untuk mengubah).
4. Cetak satu siswa, **Cetak semua** satu kelas (A4, satu rapor per halaman), atau **Unduh Word**.
5. **Ekspor Excel**: di tab Nilai mapel (satu mapel) atau Wali kelas (lembar *Leger* semua mapel + ketidakhadiran, lalu
   satu lembar per mapel berisi NISN, NIS, nama, nilai akhir, dan deskripsi capaian). Urutan siswa menurut nama, sama
   dengan Dapodik, dan NISN tetap teks (nol di depan tidak hilang). Nilai yang belum disimpan guru dibiarkan kosong
   dengan keterangan "Belum diisi".

**SMA/SMK:** nilai rapor untuk seleksi masuk PTN (SNBP) tetap dikirim lewat aplikasi **e-Rapor Kemendikdasmen**
(tersambung Dapodik). Pakai SCH untuk mengolah nilai, lalu salin kolom dari berkas Ekspor Excel ke template impor
e-Rapor resmi. SD/SMP dapat memakai rapor SCH langsung, kecuali Dinas setempat mewajibkan aplikasi resmi.

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
| `-proxy-header xff` | **wajib** bila di belakang nginx/proxy lain (bukan Cloudflare): IP asli dibaca dari elemen terakhir `X-Forwarded-For`, header `CF-Connecting-IP` diabaikan (bisa dipalsukan pengunjung lewat nginx). Bawaan `cf` untuk Cloudflare Tunnel | `cf` |
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
- Password disimpan sebagai hash bcrypt. Login dibatasi 5 percobaan gagal per akun per perangkat dalam 10 menit,
  termasuk bila dikirim serentak. Password admin minimal 10 karakter, guru 8, siswa 6; password admin & guru
  tidak boleh memuat username atau berupa password umum.
- Semua aturan akses diperiksa di server: siswa hanya melihat kelasnya, guru hanya mapel/ujian miliknya.
  Modul yang dimatikan juga ditolak di server.
- Kunci jawaban kuis/CBT tidak pernah dikirim ke siswa; nilai dihitung di server; waktu ujian dihitung server.
- Kartu QR siswa ditandatangani (HMAC) sehingga tidak bisa dibuat sendiri dari NIS teman.
- File upload diperiksa jenis aslinya (JPG/PNG/WEBP/GIF/PDF) dan hanya bisa dibuka pengguna yang login. Berkas
  pendaftar PPDB (KK, akta) hanya bisa dibuka admin dan guru yang diberi tugas PPDB.
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

Alur rilis:
1. Naikkan angka di file `VERSION` (mis. `1.4.0` → `1.4.1` untuk perbaikan, `1.5.0` untuk fitur baru).
2. Tulis catatannya di `CHANGELOG.md` dengan judul `## v1.4.1 — tanggal`, lalu commit.
3. **Rilis publik**: `git tag v1.4.1 && git push origin v1.4.1`. GitHub Actions menjalankan tes, membuat paket
   Windows/Linux/Raspberry Pi, dan menerbitkannya di halaman Releases beserta catatan dari CHANGELOG.
4. **Server sendiri lewat SSH**: salin `deploy.env.example` menjadi `deploy.env` (tidak ikut di-commit), isi alamat
   server & kunci SSH, lalu jalankan `./deploy.sh`. Skrip akan: build dengan versi tertanam → backup database ke
   `/opt/sch/data/backups/sebelum-v<versi>-<waktu>.db` → simpan rilis di `/opt/sch/releases/` → restart →
   cek `/api/version`. Bila versi baru tidak merespons, otomatis kembali ke rilis sebelumnya.
5. Kembali ke rilis lama secara manual: `./rollback.sh` (lihat daftar) lalu `./rollback.sh 1.4.0`.
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

## 12. Kontribusi & laporan
- Bug dan usul fitur: buka **Issue**. Panduan kontribusi kode ada di [CONTRIBUTING.md](CONTRIBUTING.md).
- Celah keamanan: laporkan secara pribadi, **jangan** lewat issue publik. Lihat [SECURITY.md](SECURITY.md).

## 13. Lisensi
SCH adalah perangkat lunak bebas di bawah **[GNU Affero General Public License v3.0](LICENSE)**.

- Sekolah, dinas, dan siapa pun boleh memakai, menyalin, dan memasangnya **gratis**, termasuk untuk banyak sekolah.
- Boleh diubah. Bila versi ubahan dibagikan **atau dijalankan sebagai layanan untuk orang lain** (misalnya hosting
  yang dijual ke sekolah), kode sumber ubahannya wajib dibuka dengan lisensi yang sama.
- Tanpa jaminan apa pun; sekolah bertanggung jawab atas data dan backup-nya sendiri.
