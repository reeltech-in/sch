# Catatan Rilis SCH

Format versi: `MAYOR.MINOR.PERBAIKAN`
- **MAYOR**: perubahan besar yang bisa membutuhkan migrasi data/manual.
- **MINOR**: fitur baru, kompatibel dengan data lama.
- **PERBAIKAN**: perbaikan bug/keamanan tanpa fitur baru.

Versi yang sedang berjalan terlihat di menu kiri aplikasi (bawah nama pengguna), di `GET /api/version`,
di header respons `X-SCH-Version`, dan dengan perintah `sch-server -version`.

## v1.29.0 — 2026-11-25 (Backup lengkap & pemulihan, hasil audit ketahanan)
- **Backup lengkap**: tombol baru di Pengaturan → Penyimpanan & backup mengunduh satu file **.zip** berisi database
  dan semua berkas (foto, berkas PPDB, gambar soal, e-book). Sebelumnya backup hanya berisi database, sehingga
  pemulihan meninggalkan foto & berkas yang hilang. Backup berjalan tanpa mengganggu pemakaian (login tetap normal).
- **Pemulihan yang aman**: `sch-server -db data/sch.db -restore <file.zip atau .db>`. Backup diperiksa keutuhannya
  dulu, data yang sedang terpasang disimpan (tidak dihapus), dan pemulihan ditolak bila server masih berjalan.
- **Backup harian di disk lain**: pilihan `-backup-dir` (install.sh: `SCH_BACKUP_DIR`) untuk flashdisk/NAS.
- Halaman yang mengalami galat tak terduga kini menampilkan pesan "muat ulang halaman", bukan diam membeku.
- Penyimpanan yang gagal kini tidak lagi diam-diam setengah tersimpan: simpan kelulusan dan ubah nama tahun ajaran
  melaporkan galatnya; pembersihan foto lama memperbarui database dulu sebelum menghapus file (tidak ada lagi tautan
  ke foto yang sudah hilang); kegagalan catatan aktivitas, antrean WhatsApp, dan pembersihan berkala tercatat di log.
- Pengembang: tes otomatis berjalan di GitHub untuk setiap push (format, vet, tes, sintaks JS); tes baru untuk backup &
  pemulihan, alur materi–tugas–nilai, kasir, dan PPDB sampai menjadi siswa.

## v1.28.4 — 2026-11-25 (Perbaikan login & cek kelulusan massal, unduhan lambat)
Sangat disarankan bagi yang sudah memasang v1.28.1–v1.28.3.
- **Perbaikan regresi v1.28.1: login massal.** Ratusan siswa yang login bersamaan dari jaringan sekolah (satu IP)
  sebagian besar ditolak "Terlalu banyak percobaan login" walau password benar (hanya 60 dari 300 berhasil),
  misalnya saat ujian dimulai. Kini semua berhasil; tebakan serentak ke satu akun tetap dibatasi 5.
- **Perbaikan regresi v1.28.1: cek kelulusan & status PPDB massal.** Saat pengumuman dibuka, cek dengan data benar dari
  WiFi sekolah sebagian ditolak (73 dari 200 berhasil). Kini semua berhasil; tebakan serentak ke satu NISN/nomor
  pendaftaran tetap dibatasi.
- **Unduhan berkas besar tidak lagi terputus.** Lampiran PDF materi / e-book unggahan berukuran besar terputus setelah
  ±60 detik di jaringan lambat (file 10 MB di 120 KB/detik hanya terunduh 9,2 MB). Kini batas waktunya 30 menit.

## v1.28.3 — 2026-11-23
- Tampilan hanya mencantumkan **© ReeLTech** (tautan "Kode sumber" dihapus).

## v1.28.2 — 2026-11-23 (Hak cipta)
- Tampilan kini mencantumkan **© ReeLTech** beserta tautan **Kode sumber** (repo GitHub, lisensi AGPL-3.0) di bawah
  menu kiri dan di halaman login.

## v1.28.1 — 2026-11-23 (Perbaikan keamanan)
Disarankan untuk semua sekolah. Perbarui dengan perintah pemasangan yang sama; data tidak berubah.

**Login & tebakan**
- Batas percobaan login kini tidak bisa ditembus dengan mengirim banyak percobaan sekaligus. Sebelumnya puluhan
  tebakan password serentak semuanya diperiksa (batasnya 5). Hal yang sama diperbaiki untuk cek kelulusan dan
  cek status PPDB. Login yang berhasil tidak memakan jatah, jadi siswa satu sekolah di balik satu IP tetap lancar.
- **Aturan password baru**: admin minimal 10 karakter, guru minimal 8, siswa tetap 6. Password admin & guru tidak
  boleh memuat username atau berupa password umum (12345678, password, admin123, dsb.). Berlaku saat membuat akun,
  mengganti, atau mengimpor; password yang sudah ada tetap bisa dipakai.
- Pilihan baru `-proxy-header xff` (install.sh: `SCH_PROXY_HEADER=xff`) untuk SCH di belakang **nginx/proxy lain**:
  tanpa ini pengunjung bisa memalsukan header IP untuk mengakali pembatas login. Cloudflare Tunnel tidak perlu diubah.

**Data pribadi**
- **Berkas pendaftar PPDB** (KK, akta, foto) kini hanya bisa dibuka admin dan guru yang diberi tugas PPDB.
  Sebelumnya siapa pun yang sudah login bisa membukanya bila mengetahui nama filenya. Berkas pendaftar lama ikut
  terlindungi otomatis saat server diperbarui.
- Formulir PPDB tidak lagi menyebut nomor pendaftaran bila NISN/NIK sudah terdaftar (orang lain yang tahu NISN
  seorang anak tidak bisa mengetahui anak itu mendaftar).

**Penyalahgunaan formulir publik**
- WhatsApp dari formulir PPDB tidak bisa lagi dipakai mengirim teks bebas lewat nomor sekolah: nama disingkat,
  maksimal 2 pesan per nomor tujuan dan 300 per hari. Pesan absensi didahulukan dari pesan PPDB.
- Unggahan publik (PPDB, buku tamu) dibatasi 60 MB per perangkat per hari, dan kiriman besar yang sedang diproses
  dibatasi per perangkat, agar satu orang tidak bisa menghabiskan jatah atau memori server.

**Lain-lain**
- E-Voting: setelah ada suara masuk, pemilih, kelas, waktu mulai, serta nomor/nama kandidat terkunci, dan
  pemilihan yang sudah selesai tidak bisa dibuka lagi (mencegah mengintip perolehan suara di tengah jalan).
- Mematikan modul **Pembelajaran** kini benar-benar menolak aksesnya di server, bukan hanya menyembunyikan menu.
- Perbaikan: berkas **portofolio siswa** tidak lagi ikut terhapus oleh pembersihan file harian. Berkas lama yang
  diganti, dan foto selfie yang tidak dipakai, langsung dibuang sehingga tidak memenuhi penyimpanan.

## v1.28.0 — 2026-11-22 (Soal AI bergambar & berstimulus)
- **Generator Soal (Asisten AI)** kini bisa membuat soal bergambar seperti di LKS/buku paket: segitiga trigonometri,
  bangun datar/ruang, grafik fungsi, diagram batang, tabel data, garis bilangan, dsb. Pilihan baru **Gambar / grafik
  soal**: *Bila diperlukan materi* (bawaan), *Sebanyak mungkin*, atau *Tanpa gambar*.
  - AI menggambar dalam format SVG; browser guru mengubahnya menjadi PNG lalu menyimpannya di server. Kode SVG tidak
    pernah disajikan apa adanya (skrip di dalamnya tidak bisa jalan).
  - Soal boleh diawali wacana/data dengan pengantar wajar ("Bacalah teks berikut!", "Perhatikan gambar berikut!"),
    tanpa label "Stimulus". Setiap soal berdiri sendiri karena urutan soal diacak saat ujian.
  - Rumus ditulis dengan simbol biasa (x², √2, π, 30°), bukan LaTeX.
  - Gambar ikut tampil di pratinjau, cetak, unduhan **Word**, dan saat **Masukkan ke CBT/Kuis**.
- **Kuis mapel** kini mendukung gambar soal (sebelumnya hanya CBT): unggah di formulir soal, tampil di layar siswa.
- **Impor soal dari teks** (CBT) membaca baris gambar `![..](/files/..)` dan blok ```` ```svg ````.
- Perbaikan: nomor soal di pratinjau dokumen AI tidak lagi selalu "1." untuk setiap soal.
- Pembersihan file yatim harian tidak memindahkan gambar soal kuis maupun gambar di dokumen soal AI yang belum diimpor.

## v1.27.1 — 2026-11-11 (Scan absen ganda lebih rapi)
- **Scan QR absensi**: kartu yang di-scan ulang tetap diterima tanpa mencatat ulang (seperti sebelumnya, satu catatan per
  siswa per hari), tapi tampilannya kini jelas:
  - kartu hasil biru "Sudah absen masuk pukul 12:29, tidak perlu scan lagi" (bukan oranye seperti peringatan); di mode
    Otomatis ditambah "Absen pulang bisa mulai pukul 13:29";
  - riwayat tidak lagi berisi baris berulang: scan ulang siswa yang sama menaikkan barisnya ke atas dengan tanda
    "scan ulang 2×";
  - penghitung menghitung **siswa** yang tercatat, bukan jumlah scan;
  - kode yang sama dalam 3 detik (scanner USB membaca dua kali) diabaikan tanpa mengirim ke server.

## v1.27.0 — 2026-11-11 (Ekspor Excel e-Rapor)
- **Ekspor Excel** di e-Rapor untuk memindahkan nilai ke e-Rapor resmi Kemendikdasmen tanpa mengetik ulang:
  - tab **Nilai mapel**: satu lembar untuk mapel itu (guru pengampu atau staf rapor);
  - tab **Wali kelas**: lembar **Leger** (nilai akhir semua mapel, rata-rata, sakit/izin/tanpa keterangan) lalu satu
    lembar per mapel (wali kelas atau staf rapor).
  - Kolom: No, NISN, NIS, Nama, Nilai Akhir, Deskripsi Capaian (gabungan seperti di rapor), Capaian Tertinggi, Perlu
    Ditingkatkan, Keterangan. Siswa urut nama (seperti Dapodik), NISN disimpan sebagai teks, judul kolom dibekukan.
  - Nilai yang belum disimpan guru dikosongkan dengan keterangan "Belum diisi" (tidak diisi nilai otomatis).
  - Dibuat tanpa pustaka luar; diuji terbuka di LibreOffice dan pembaca Excel (openpyxl). Satu kelas 36 siswa × 10 mapel:
    6,6 ms; satu mapel 1,7 ms (data uji 1.080 siswa, 300 mapel).
- Sekolah berjenjang **SMA/SMK** melihat keterangan di e-Rapor bahwa nilai untuk SNBP tetap dikirim lewat e-Rapor resmi.
- README: tanggal semester diatur di menu Tahun Ajaran (bukan di Pengaturan e-Rapor).

## v1.26.1 — 2026-11-10 (Pindah tab tanpa kedip)
- **Pindah tab, submenu, saringan, dan halaman** di dalam halaman yang sama (mis. Pengaturan, e-Rapor, Perpustakaan,
  Inventaris KIB, Surat, Kasir, Absensi harian, pilihan mapel/kelas) tidak lagi mengganti layar dengan kerangka abu-abu
  dan tidak melompat ke atas: isi lama tetap tampil (sedikit redup bila > 120 ms) sampai isi baru siap.
- Tab/submenu yang diklik **langsung tersorot**, sebelum datanya selesai dimuat.
- Data pengaturan (absensi, zona waktu, Asisten AI, WhatsApp, dll.) disimpan sementara 2 menit di browser dan dikosongkan
  setiap ada perubahan; data siswa (Beranda, Mapel, Tugas, Kuis, Nilai) 30 detik. Uji dengan jeda jaringan 300 ms:
  tab Pengaturan 330 → 16 ms, pindah menu siswa 310 → 11 ms; 40 tab/saringan admin, guru, siswa tanpa kedip.

## v1.26.0 — 2026-11-10 (Pagination bernomor di semua daftar besar)
- **Pagination bernomor** seragam di 7 daftar: Siswa/Guru, Kelulusan, Catatan siswa, Log aktivitas, PPDB,
  Inventaris, dan Perpustakaan (katalog & rak). Info "Menampilkan 51–100 dari 2.886 siswa", tombol ‹ Sebelumnya ·
  1 … 4 5 6 … 58 · Berikutnya ›, pilihan 25/50/100/200 per halaman; halaman tersimpan di alamat (kembali ke halaman
  yang sama), halaman yang melewati batas otomatis ke halaman terakhir. Di HP cukup ‹ › agar muat satu baris.
- **Siswa**: saringan kelas, status (bawaan **aktif**; alumni/nonaktif lewat saringan), dan pencarian dijalankan di
  server. Uji 2.886 siswa (5 tahun alumni): 516 KB → 9 KB per halaman, tampil ±300 ms → ±9 ms.
- **Kelulusan**: bawaan kelas tingkat akhir, hanya siswa aktif, cari di server; jumlah "belum lengkap datanya" untuk
  seluruh hasil. **Catatan siswa**: jumlah & poin prestasi/pelanggaran dihitung atas seluruh hasil saringan.
  **Inventaris**: jumlah "perlu perbaikan" atas seluruh hasil.
- **Cetak & CSV** (catatan, PPDB, label/KIB inventaris, label buku) tetap mencakup **semua** hasil saringan, bukan hanya
  halaman yang tampil (diambil bertahap 1.000 baris).
- Perbaikan: peringatan Beranda "siswa belum punya kelas" tidak lagi menghitung alumni/siswa nonaktif.

## v1.25.3 — 2026-11-10 (Header & sidebar rapi sejajar)
- **Header** kini dua baris: label kecil berwarna (grup menu, mis. "Pembelajaran"; Beranda: "Ringkasan hari ini") di
  atas judul halaman yang tebal; tinggi header 72 px.
- **Sidebar desktop**: blok nama sekolah setinggi header dengan garis bawah yang menyambung lurus ke garis header,
  dan garis tegak sidebar berupa border 1 px yang tidak tertutup header; header & blok nama satu warna. Garis
  presisi di tema terang & gelap (diukur: garis bawah y=72 px keduanya, garis tegak tepat di tepi kiri header).
- Keterangan di bawah nama sekolah: jenjang sekolah (bila diatur) atau "Sistem Informasi Sekolah"; peran pengguna
  tetap di menu akun.

## v1.25.2 — 2026-11-10 (Semester diatur di menu Tahun Ajaran)
- **Semester satu tempat:** isian **Akhir semester ganjil** pindah ke form Tahun Ajaran (bersama tanggal mulai &
  selesai), dan daftar Tahun Ajaran menampilkan rentang Ganjil & Genap. Server menolak tanggal di luar tahun ajaran.
  Tab Pengaturan e-Rapor kini hanya berisi ekstrakurikuler dengan keterangan semester aktif dan tautan ke menu
  Tahun Ajaran; pengaturan semester terpisah di e-Rapor dihapus.

## v1.25.1 — 2026-11-10 (Pembelajaran, e-Rapor, Portofolio lebih cepat)
- **Statistik perencana query SQLite selalu segar.** Server menjalankan `PRAGMA optimize=0x10002` saat menyala dan
  tiap jam (rekomendasi resmi SQLite untuk koneksi yang hidup lama). Sebelumnya hanya `PRAGMA optimize` harian di
  koneksi tulis, padahal query baca lewat koneksi lain, sehingga statistik dari saat database masih kecil membuat
  SQLite membaca seluruh tabel walau indeksnya ada. Uji data satu sekolah penuh (30 kelas, 1.080 siswa, 300 mapel,
  ±700 ribu baris nilai & absensi): Rekap Nilai 87 → 6 ms, Nilai mapel e-Rapor 81 → 5 ms, daftar Kuis 41 → 1 ms,
  Beranda guru 28 → 1 ms; semua halaman ketiga modul ≤ 6 ms di server.
- **Rekap Nilai** tidak lagi mengirim seluruh baris absensi/nilai mentah (123 KB → ±5 KB); nilai sudah dihitung server.
- **e-Rapor**: periode & daftar tugas rapor diminta bersamaan dan disimpan sementara di browser; **Rapor siswa**
  memuat daftar semester dan rapornya dalam satu permintaan (hemat satu putaran ±100–200 ms lewat internet).

## v1.25.0 — 2026-11-10 (e-Rapor Kurikulum Merdeka & Portofolio Siswa)
- **e-Rapor per semester** (modul baru "e-Rapor", menu e-Rapor untuk admin, guru, dan siswa) sesuai komponen
  laporan hasil belajar Panduan Pembelajaran & Asesmen 2025 / Permendikdasmen 13/2025:
  - **Semester ganjil/genap** pada tiap tahun ajaran (atur tanggal akhir ganjil; bawaan 31 Desember).
  - **Nilai & capaian kompetensi per mapel** oleh guru pengampu: nilai awal otomatis dari rumus 40/40/20 untuk
    rentang semester, bisa disesuaikan; deskripsi capaian tertinggi & perlu ditingkatkan dengan saran dari nilai
    per materi, atau **saran Asisten AI** untuk satu kelas sekaligus (1 kuota; nama siswa tidak dikirim ke AI).
  - **Kokurikuler** 8 dimensi profil lulusan (SB/BSH/MB/BB) dengan deskripsi otomatis, **catatan wali kelas**,
    **ketidakhadiran** dari absensi harian (bisa dikoreksi), **keputusan naik/tinggal/lulus** di semester genap.
  - **Ekstrakurikuler**: daftar kegiatan & pembina, anggota per tahun ajaran, predikat & keterangan oleh pembina.
  - **Terbitkan per kelas** (hanya bila semua mapel lengkap): siswa melihat rapornya, isi rapor terkunci sampai
    terbit dibatalkan. **Cetak** satu siswa atau satu kelas (A4, satu rapor per halaman) dan **unduh Word**.
- **Portofolio siswa** (internal): siswa mengunggah karya (gambar/PDF/tautan, per mapel), guru mapel & wali kelas
  memberi catatan dan nilai; jumlah karya tampil di tab Nilai mapel e-Rapor. Karya yang sudah dinilai tidak bisa
  diubah siswa.
- Tes otomatis untuk alur rapor, ekstrakurikuler, portofolio, saran deskripsi, dan saran AI.
- **Perbaikan Kelas:** mengubah kelas (mis. mengisi wali kelas) dulu mengosongkan kolom tahun ajarannya karena
  formulir tidak memuat isian itu. Kini tahun ajaran kelas selalu mengikuti tahun kelasnya, data yang sudah
  terlanjur kosong diperbaiki otomatis saat pembaruan, dan ganti nama tahun ajaran ikut memperbarui kelas. Kolom
  "Tahun ajaran" di tabel Kelas dihapus karena halaman itu sudah menampilkan satu tahun ajaran terpilih.
- **Tahun ajaran satu sumber:** isian "Tahun ajaran" di Pengaturan → Identitas sekolah kini hanya menampilkan tahun
  aktif (tombol "Ubah di Tahun Ajaran"); server mengabaikan tahun ajaran dari Pengaturan dan selalu memakai tahun
  yang aktif di menu Tahun Ajaran, sehingga beranda, kop cetakan, rekap, dan kelas tidak bisa berbeda tahun.

## v1.24.3 — 2026-11-10 (Judul halaman di header)
- Judul halaman (Beranda, Tugas, Rekap Nilai, …) pindah ke header di kiri, sejajar dengan jam dan menu akun; di HP
  menggantikan nama sekolah. Keterangan dan tombol halaman tetap di bawahnya. Judul juga dipakai sebagai judul tab
  browser ("Tugas · Nama Sekolah"). Saat dicetak, judul asli tetap tercetak.

## v1.24.2 — 2026-11-10 (Audit tahap 3: perapian kode)
- **Nilai akhir dihitung di server** (nilai.go) dengan jam server: rekap guru, halaman nilai siswa, dan CSV memakai satu
  rumus yang sama (40 % tugas + 40 % kuis + 20 % kehadiran), kini dijaga tes. Jam HP/PC yang salah tidak lagi
  memengaruhi apakah tugas lewat tenggat dihitung 0. Angka tidak berubah.
- Penilaian pilihan ganda CBT & kuis mapel memakai satu fungsi bersama, begitu pula penilaian otomatis attempt yang
  waktunya habis (setiap attempt dalam transaksinya sendiri).
- Fungsi bantu lintas modul (tanggal, persen, daftar kelas, batas halaman publik) dikumpulkan di util.go.
- Nama yang membingungkan diganti: simpan otomatis jawaban kuis siswa, nilai tugas, dan simpan kelulusan.

## v1.24.1 — 2026-11-10 (Audit tahap 1: keamanan & integritas data)
- **Kuis mapel dinilai dari soal yang ditampilkan.** Daftar soal dicatat saat siswa mulai; soal yang ditambahkan guru
  sesudahnya tidak ikut dihitung salah, dan saat melanjutkan siswa melihat soal yang sama.
- **Unggahan publik dibatasi.** Buku tamu & PPDB punya jatah bersama 300 MB per hari dan tidak boleh memakai 20 % ruang
  terakhir, sehingga selfie absensi, gambar soal, dan logo tetap bisa diunggah. PPDB memeriksa semua isian (termasuk
  NISN/NIK ganda) sebelum menyimpan berkas; berkas dihapus lagi bila penyimpanan gagal.
- **Tugas yang sudah dinilai tidak bisa ditimpa** walau siswa mengirim ulang tepat saat guru menilai (satu pernyataan atomik).
- **Buku yang pernah dipinjam tidak bisa dihapus**, agar riwayat pinjaman & dendanya tetap utuh (ubah stok jadi 0).
- **Reset password / nonaktifkan akun oleh admin** dan **ganti password sendiri** kini mengubah password dan memutus sesi
  lama dalam satu transaksi.
- **E-voting bilik:** petugas yang memindai kartu dicatat pada daftar hadir pemilih (pilihan tetap rahasia), pengelola
  melihat rekap "Suara lewat bilik per petugas", dan jeda antar-suara per petugas menjadi 10 detik (kartu yang ditolak
  tidak memicu jeda).
- Penolakan hapus yang boleh dilewati admin kini ditandai jelas oleh server; tes baru untuk semua perbaikan di atas.
- **Audit tahap 2 (tes):** alur yang sebelumnya tidak pernah dites kini dijaga tes otomatis — absensi mapel (simpan,
  ubah, rekap, hapus, siswa luar kelas, guru lain), tugas (buat, kumpul, ubah, nilai dalam batas, guru lain ditolak),
  impor pengguna, aktivasi tahun ajaran, PPDB → akun siswa, pinjam/kembali buku, serta penolakan per peran untuk
  kasir, sarpras, PPDB, acara ujian, perpustakaan, bilik e-voting, impor, dan cadangan. Cakupan tes 52 % → 58 %.

## v1.24.0 — 2026-11-10 (Header baru & sudut kartu lebih tegas)
- **Header di bagian atas** (desktop): di kanan ada hari, tanggal, dan jam menurut waktu server (WIB/WITA/WIT), lalu **menu akun**
  berisi nama & peran, Akun Saya (profil & kata sandi), Pengaturan (admin), pilihan **tema Sistem / Terang / Gelap**,
  dan Keluar. Di HP header berisi tombol menu, nama sekolah, dan tombol akun yang sama. Menu akun bisa dipakai dengan
  keyboard (panah, Home/End, Esc) sesuai pola Menu Button WAI-ARIA. Nama pengguna, tombol tema, dan Keluar tidak lagi
  di bawah menu samping.
- Sudut kartu, tabel, dan jendela tidak terlalu membulat lagi (kartu 20 → 12 px, jendela 24 → 16 px); tombol, isian,
  tab, dan elemen kecil ikut disesuaikan (8–10 px) agar seragam. Label status dan bilah progres tetap berbentuk pil.

## v1.23.2 — 2026-11-10 (Nilai siswa terlindung dari salah hapus; pesan jelas saat ujian CBT tidak bisa dibuka)
- **Ujian CBT, Kuis, dan Tugas yang sudah ada hasil siswa tidak bisa dihapus guru.** Menghapusnya dulu ikut menghapus
  nilai siswa secara permanen. Sekarang guru mendapat penjelasan ("Kuis ini sudah dikerjakan 12 siswa…"); yang belum
  ada hasilnya tetap bisa langsung dihapus. Ujian CBT otomatis tertutup setelah jadwal selesai, atau pekerjaan siswa
  bisa dihapus satu per satu di Hasil ujian.
- **Admin** tetap bisa menghapus setelah konfirmasi kedua yang menyebut jumlah nilai yang ikut hilang; tercatat di
  Log aktivitas ("beserta hasil N siswa"). Penghapusan berantai tetap bersih, tidak meninggalkan data yatim (ada tesnya).
- Hapus mapel kini juga menghitung kuis yang sedang dikerjakan siswa.
- Bila siswa membuka halaman ujian yang sudah dihapus guru (mis. tab lama masih terbuka), ditutup, sudah dikerjakan,
  atau belum dibuka, tampil kartu "Ujian tidak bisa dibuka" dengan alasannya dan tombol **Kembali ke daftar ujian**,
  bukan "Terjadi kendala · Data tidak ditemukan."

## v1.23.1 — 2026-11-10 (Kerapian menu guru & siswa di HP)
- **Absensi mapel (guru, HP)**: tombol Hadir/Izin/Sakit/Alfa selebar kartu dan lebih tinggi (±43 px) agar tidak salah sentuh;
  tombol **Simpan absensi** menempel di bawah layar sehingga kelas panjang tidak perlu digulir ke atas.
- **Rekap nilai, rekap absensi harian, rekap wali kelas**: kolom nama siswa tetap terlihat saat tabel digeser ke samping.
- **Nilai & Kehadiran (siswa)**: kehadiran ditulis lengkap ("Hadir 3 · Izin 0 · Sakit 0 · Alfa 0"), bukan singkatan H/I/S/A.
- **Beranda & Mata Pelajaran (siswa)**: mapel yang belum punya materi menampilkan "Belum ada materi", bukan "0/0" dengan bilah kosong.

## v1.23.0 — 2026-11-10 (Jenjang sekolah di satu tempat)
- **Jenjang sekolah** (SD/MI, SMP/MTs, SMA/MA, SMK/MAK) diatur sekali di **Pengaturan → Identitas sekolah**. Setelah
  diatur, semua yang terkait jenjang mengikuti:
  - **Kelas**: pilihan tingkat hanya tingkat jenjang itu (SMP: 7–9), dan server menolak tingkat di luar jenjang.
  - **Mata pelajaran**: saran mapel standar sesuai struktur Kurikulum Merdeka jenjang itu (mis. SMP tanpa IPAS/Fisika);
    mapel yang sudah dipakai sekolah tetap tampil.
  - **Kenaikan kelas**: saran "Lulus" untuk tingkat akhir jenjang (6, 9, atau 12).
  - **Impor e-book Kemendikbud**: langsung tersaring ke jenjang & kelas sekolah (tetap bisa diubah).
  - **Asisten AI** (RPP/RPM/Modul Ajar, LKPD, soal, Prota/Promes, KKTP): tanpa pilihan jenjang lagi — kelas hanya
    kelas jenjang sekolah lengkap dengan fase ("Kelas 7 · Fase D"), saran mapel per jenjang, lama 1 JP otomatis, dan
    server selalu memakai jenjang dari Pengaturan serta menambahkan instruksi tegas ke AI agar materi, istilah, contoh,
    dan tingkat kesulitan sesuai jenjang & fase.
- Selama belum diatur, aplikasi bekerja seperti sebelumnya (tidak membatasi) dan dasbor admin menampilkan pengingat
  "Jenjang sekolah belum diatur".

## v1.22.0 — 2026-11-10 (RPP Kurikulum Merdeka terbaru: format, sunting, cetak resmi, Word)
- **Asisten AI → RPP / RPM / Modul Ajar**, mengikuti Kurikulum Merdeka dengan Pembelajaran Mendalam
  (Permendikdasmen No. 13 Tahun 2025, 8 dimensi profil lulusan; CP Kepka BSKAP No. 046/H/KR/2025):
  - **Pilihan format**: RPM – Rencana Pembelajaran Mendalam (bawaan), Modul Ajar lengkap, atau RPP satu lembar.
  - **Fase otomatis** dari kelas (7A → Fase D) dan **lama 1 JP** sesuai jenjang (SD 35, SMP 40, SMA/SMK 45 menit,
    bisa diubah); isian JP per pertemuan & jumlah pertemuan.
  - **Pemeriksa alokasi waktu**: tiap pertemuan wajib berjumlah tepat (JP × menit); bila hasil AI meleset, tampil
    peringatan per pertemuan.
  - **Capaian Pembelajaran**: guru bisa menempel CP resmi; bila kosong, ringkasan CP dari AI diberi tanda "cocokkan
    dengan dokumen CP resmi".
  - Dimensi profil lulusan dibatasi 1–3.
- **Sunting** hasil AI (judul & isi, dengan pratinjau langsung) lalu simpan ulang — hanya dokumen milik sendiri.
- **Cetak resmi**: kop sekolah (logo, nama, alamat), kolom tanda tangan "Mengetahui Kepala Sekolah" & "Guru Mata
  Pelajaran" lengkap dengan NIP, tempat dan tanggal (nama & NIP terisi otomatis, bisa diubah).
- **Unduh Word (.docx)** dengan kop, tabel, daftar, dan tanda tangan yang sama — siap dirapikan atau dikirim.

## v1.21.1 — 2026-11-10 (Kuis tersimpan otomatis, menu lebih ringan)
- **Kuis mapel tersimpan ke server selama dikerjakan** (seperti CBT): jawaban tidak hilang bila HP mati, bisa dilanjutkan
  dari perangkat lain, kiriman terlambat menilai jawaban yang sudah tersimpan (sebelumnya nilainya 0), kuis yang tidak
  dikirim dinilai otomatis setelah waktunya habis, dan kirim ulang setelah koneksi putus menampilkan nilai tersimpan.
- **Log Aktivitas**: pencarian teks tanpa tanggal "Dari" dibatasi 90 hari terakhir dan memakai indeks waktu
  (pada 300 ribu catatan: dari ±57 ms + 57 ms menjadi ±0,1 ms). Isi tanggal "Dari" untuk mencari lebih lama.
- **Pengaturan**: menyimpan cukup satu permintaan (server hanya mengubah kolom yang dikirim dan membalas dengan
  pengaturan terbaru); sebelumnya ambil–simpan–ambil–ambil.
- **Menu Siswa**: saring kelas berjalan langsung di browser tanpa memuat ulang daftar dari server; bisa digabung
  dengan kotak cari.

## v1.21.0 — 2026-11-10 (Perbaikan bertahap: konsistensi, kerapian kode, aksesibilitas)
**Perbaikan untuk pengguna**
- **Input LJK** kini juga menampilkan siswa yang sudah naik kelas/lulus, sama seperti daftar hasil CBT
  (nilai ujian tahun lalu tetap bisa diinput setelah kenaikan kelas).
- **Pengaturan absensi**: jam masuk/pulang divalidasi & dirapikan ("7:15" → "07:15"; sebelumnya format salah membuat
  tidak ada siswa tercatat terlambat), jam pulang harus setelah jam masuk, radius & koordinat diperiksa.
- **Dasbor guru** menghitung materi/tugas/kuis/tugas belum dinilai untuk tahun ajaran aktif saja (lebih cepat, tidak
  tercampur tahun lalu).
- **Pindah menu lebih ringan**: permintaan halaman lama dibatalkan saat pindah menu, halaman publik (kelulusan/PPDB)
  tidak saling menimpa di WiFi lambat, data rujukan yang berangkat sebelum ada perubahan tidak disimpan basi.
- **Jendela dialog aksesibel**: halaman di belakang tidak bisa difokus selama jendela terbuka; fokus kembali ke tombol
  pembuka saat ditutup.
- Peringatan "ujian sedang berjalan" saat Keluar hanya muncul bila jawaban belum terkirim.
- Daftar CBT guru/admin dibatasi 500 terbaru; e-voting & Asisten AI memuat datanya bersamaan.

**Kerapian kode (tanpa perubahan tampilan)**
- Mesin ujian kuis mapel & CBT memakai hitung mundur, kirim otomatis (dengan coba ulang), dan kartu hasil yang sama.
- Satu inti permintaan ke server untuk semua fitur (versi aplikasi, koreksi jam, pesan jaringan, sesi habis).
- Satu cara membuka/menutup kamera & mengambil foto (scan QR, selfie, buku tamu, LJK), satu penangan scanner USB,
  satu formulir ganti password.
- 57 deklarasi CSS yang selalu ditimpa dihapus — tampilan diverifikasi identik di desktop & HP.
- Fungsi server yang terlalu panjang dipecah: kenaikan kelas, mulai ujian CBT (kini membaca lewat koneksi baca,
  tidak antre di koneksi tulis saat ratusan siswa mulai bersamaan), dan pembuatan dokumen Asisten AI.
- Tes baru: semua jalur kenaikan kelas, melanjutkan ujian CBT, Asisten AI dengan penyedia palsu (streaming, galat, kuota).

## v1.20.0 — 2026-11-10 (Hasil audit: keamanan, ujian, query, tampilan)
**Keamanan**
- Mode proxy (`-proxy` tanpa `-proxy-ip`) kini hanya mempercayai proxy di komputer yang sama (127.0.0.1). Sebelumnya
  pengguna LAN bisa memalsukan header IP dan lolos dari pembatas login/PPDB/kelulusan. Server live tidak terdampak
  (memakai `-proxy-ip`).
- Guru tidak bisa lagi menghapus absensi harian siswa kelas lain dengan mengganti `siswa_id`.
- Formulir publik PPDB & buku tamu: batas per perangkat dicek **sebelum** membaca isi (hingga 10 MB).
- E-book: unduhan dari sumber tidak mengikuti pengalihan ke alamat lain, total salinan dibatasi (= batas upload),
  dan satu akun dibatasi 15 e-book baru per 10 menit.
- PIN PPDB pada pesan WhatsApp yang gagal/kedaluwarsa/antre tidak lagi tersimpan atau tampil terbuka.

- **Cek kelulusan & status PPDB**: tebakan tanggal lahir/PIN kini dibatasi per perangkat (10 salah/10 menit, 40/hari)
  dan per NISN/nomor (setelah 20 salah sehari hanya perangkat yang belum pernah salah boleh mencoba). Orang iseng
  tidak bisa lagi mengunci siswa asli dengan sengaja salah berkali-kali.
- Absen selfie: lokasi yang kurang akurat ditolak; absen mencurigakan (titik persis di titik sekolah, koordinat
  sama dengan siswa lain, akurasi tidak dilaporkan) diberi keterangan "Selfie perlu dicek" untuk petugas.
- Galat penyedia AI hanya menampilkan pesan galat resmi penyedia, bukan isi halaman dari alamat lain.
- Server mencatat peringatan bila header proxy datang dari alamat yang tidak dipercaya (memudahkan menemukan salah setel).

> **Catatan pembaruan:** `-proxy` saja kini hanya mempercayai proxy di komputer yang sama. Bila cloudflared/nginx
> berjalan di mesin lain atau Docker, tambahkan `-proxy-ip <IP>` (install.sh: `SCH_PROXY_IP=<IP> SCH_UNIT=rewrite`).

**Ujian & nilai**
- Ujian CBT yang waktunya habis tetapi tidak dikumpulkan (HP mati, browser tertutup) kini dinilai otomatis dari
  jawaban yang tersimpan, tidak lagi "belum selesai" selamanya.
- Nilai CBT hanya dihitung dari soal yang memang diterima siswa (soal tambahan susulan tidak menurunkan nilai).
- Jumlah pilihan ujian/kuis tidak bisa dikurangi bila masih ada soal berkunci di luar pilihan baru.
- Kuota harian Asisten AI tidak bisa lagi diakali dengan menghapus dokumen.
- Siswa yang HP-nya offline saat waktu habis: ujian dinilai otomatis, dan kiriman ulang dari HP menampilkan nilainya
  (tidak terjebak pesan "sudah dikirim").

**Perbaikan lain**
- Kalender: ujian/tenggat pagi hari di tanggal pertama rentang tidak lagi hilang.
- Pinjam buku: stok terakhir tidak bisa dipinjamkan dua kali bersamaan; buku khusus e-book ditolak dengan pesan jelas.
- Nomor surat keluar tidak lagi berulang setelah ada surat yang dihapus. "Kirim ulang" WA yang kedaluwarsa kini berfungsi.
- Pemeliharaan harian tidak bisa berjalan ganda; galat penyimpanan pada pergantian tahun ajaran tidak lagi diabaikan.
- Pesan galat Asisten AI dalam bahasa Indonesia (rincian teknis ke log server).
- Katalog Buku Kemendikbud: bila sumber sedang bermasalah, daftar lama tetap dipakai dan permintaan lain tidak ikut menunggu.

- "Jalankan pemeliharaan sekarang" saat jadwal harian sedang berjalan kini memberi pesan sibuk yang jelas.
- Berkas PPDB yang kebesaran tidak lagi memakan jatah 3 kiriman orang tua per 10 menit.
- Jatah 15 e-book baru per 10 menit hanya terpakai bila unduhan berhasil.

**Kerapian kode**
- Rumus nilai, rentang tanggal sekolah→UTC, dan pembuat query string disatukan menjadi satu fungsi bersama;
  aturan CSS yang bertentangan/terpecah dirapikan.

**Kecepatan (query)**
- Indeks baru untuk kunci tamu (menghapus guru tidak lagi memindai ratusan ribu baris absensi), ISBN buku, dan
  jumlah peserta selesai CBT. Pemeriksaan baca-saja (hak akses, dasbor) tidak lagi antre di koneksi tulis.

**Tampilan**
- Sesi habis di tengah kuis/ujian tidak lagi menghapus jawaban yang belum terkirim.
- Klik ganda cepat tidak menjalankan aksi dua kali. Kamera tidak tetap menyala bila halaman ditinggal saat izin kamera ditanya.
- Isian saring daftar tetap setelah menyimpan data. Hasil saringan lama tidak menimpa yang baru di jendela impor e-book.
- iPhone: jendela & menu memakai tinggi layar yang terlihat (tombol Simpan tidak tertutup toolbar); isian 16px di semua layar sentuh.
- Dropdown di HP tidak lagi menggulir layar setiap ketikan. Tab & tombol bulan kalender berlabel untuk pembaca layar.
- Tabel kartu HP hanya diproses ulang bila ada tabel baru (lebih ringan saat timer ujian berjalan).

## v1.19.2 — 2026-11-10 (Dropdown ramah HP)
- **Dropdown "pilih atau ketik"** (mapel, kelas, kategori, penerbit, rak, dll.) diganti komponen sendiri: diketuk langsung
  menampilkan semua pilihan, mengetik menyaring daftar, baris pilihan besar untuk jari, isian baru tetap bisa diketik.
  Sebelumnya memakai daftar bawaan browser yang di HP tidak tampil (iPhone) atau hanya menampilkan yang cocok (Android).
  Bisa dipakai dengan keyboard (panah, Enter, Esc) dan pembaca layar.
- Semua dropdown biasa memakai panah yang seragam dan tidak menempel di tepi.
- Halaman **Asisten AI** di HP kini satu kolom (sebelumnya formulir terjepit setengah layar).

## v1.19.1 — 2026-11-09 (Mapel yang diampu diutamakan)
- Kolom **Mata pelajaran** di Asisten AI (RPM, LKPD, soal, Prota/Promes, KKTP) dan formulir **Ujian CBT**: mapel yang
  diampu guru tampil paling atas dengan keterangan *Mapel yang Anda ampu*. Mapel lain tetap bisa dipilih atau diketik.
- Guru yang hanya mengampu satu mapel: kolom mapel langsung terisi.

## v1.19.0 — 2026-11-09 (E-book dari Buku Kemendikbud)
- **Impor e-book gratis dari Buku Kemendikbud** (buku.kemendikdasmen.go.id): Perpustakaan → Katalog/E-book →
  *Impor e-book Kemendikbud*. Pilih dari ±1.600 judul (Kurikulum Merdeka, Kurikulum 2013, buku bacaan), saring per
  jenjang, kelas, buku siswa/guru, atau cari judul/mapel, lalu impor sekaligus. Buku yang sudah ada ditandai.
- **PDF disimpan saat pertama dibuka**: impor hanya menyimpan data buku. File PDF diunduh server satu kali saat ada
  yang pertama membacanya, lalu disajikan dari server sekolah untuk pembaca berikutnya (cepat, hemat kuota internet).
  Sampul juga disimpan. Alamat sumber tidak pernah dibuka langsung oleh browser siswa; hanya pengguna yang login bisa membaca.
- **Rak E-book**: tab *E-book* untuk pengelola, tombol *Semua buku / E-book* untuk siswa & guru. E-book tanpa buku
  fisik berlabel **Digital** (bukan "Sedang dipinjam") dan tidak ikut dicetak labelnya.
- Pemeliharaan harian membersihkan salinan e-book yang bukunya sudah dihapus dari katalog.

 (Pengaturan AI pindah ke menu Pengaturan)
- **Pengaturan Asisten AI** (URL endpoint, API key, model, kuota, kreativitas) kini diisi langsung di
  **Pengaturan → Asisten AI**, tidak lagi lewat tab terpisah di halaman Asisten AI. Tautan lama otomatis diarahkan.
- Halaman Asisten AI tinggal *Buat dokumen* dan *Riwayat*; bila AI belum diatur, admin mendapat tombol langsung ke pengaturannya.
- Skrip **install.sh** dan panduan **docs/INSTALL-UBUNTU.md**: pasang/perbarui di server Ubuntu/Debian dengan satu perintah.

## v1.18.0 — 2026-11-09 (Kalender Akademik: otomatis, libur nasional, cetak)
- **Kegiatan otomatis dari modul lain** tampil di kalender tanpa input ganda (bergaris putus-putus, "otomatis"):
  ujian CBT, kegiatan & jadwal Manajemen Ujian, masa pendaftaran PPDB, masa E-Voting, pengumuman kelulusan, dan
  tenggat tugas. Hanya yang boleh dilihat pengguna itu (siswa: ujian & tugas kelasnya, e-voting bila ia pemilih;
  guru: tugas mapel ampuannya). Diubah di modul asalnya (tombol "Buka halaman"); tidak memengaruhi hari libur absensi.
- **Libur nasional tanggal tetap sekali klik** untuk tahun ajaran aktif: 17 Agustus, 25 Desember, 1 Januari, 1 Mei,
  1 Juni (tidak digandakan). Libur yang tanggalnya ditetapkan SKB tiap tahun tetap diisi manual agar tidak salah.
- **Cetak Kalender Akademik** per semester (Ganjil/Genap) atau satu tahun ajaran: kop sekolah, kisi bulan berwarna per
  kategori, hari Minggu merah, daftar kegiatan, tanda tangan kepala sekolah.
- Desktop: maksimal 3 kegiatan per tanggal + "+n lainnya" (ketuk tanggal untuk semuanya).

## v1.17.1 — 2026-11-09 (Kalender Akademik)
- **Setiap tanggal bisa diketuk**: muncul daftar kegiatan hari itu; pengelola bisa langsung "Tambah kegiatan pada tanggal
  ini" (tanggal terisi otomatis).
- **HP**: penanda kegiatan berupa titik berwarna tegas (maks. 3 + jumlah sisanya), menggantikan garis 6 px yang redup
  dan nyaris tidak bisa disentuh. Desktop tetap menampilkan nama kegiatan.
- Keterangan warna (legend) lebih jelas; minggu terakhir yang seluruhnya bulan berikutnya tidak ditampilkan.
- Bisa dipakai dengan papan ketik (Tab + Enter) dan dibacakan pembaca layar ("17 Nov 2026, 1 kegiatan").

## v1.17.0 — 2026-11-09 (tampilan presisi di HP & desktop)
Diaudit dengan pengukuran otomatis di setiap halaman & formulir (admin 30, guru 19, siswa 15) pada 375, 768, dan 1366 px.
- **Halaman tidak lagi bergeser ke samping di HP** (sebelumnya 12 halaman): kolom konten tidak ikut melebar mengikuti
  tabel (grid minmax(0, 1fr)).
- **Tabel jadi kartu di HP**: tiap baris menjadi kartu (judul di atas, label kolom di kiri, nilai di kanan, tombol di
  bawah), label otomatis dari judul kolom; sel kosong disembunyikan. Tabel matriks (rekap nilai, rekap absensi, wali
  kelas) tetap tabel yang bisa digeser. Tampilan desktop tidak berubah.
- **Target sentuh minimal 40 px** di HP (sebelumnya 34 px pada 88 tombol); isian **16 px** agar iPhone tidak
  memperbesar layar saat mengetik; teks bantuan formulir 12 px.
- Statistik dua kolom di HP; nilai uang panjang diringkas (Rp 2,08 M) dengan nilai penuh di tooltip.
- Bilah filter menyesuaikan lebar HP; ketebalan teks isian seragam.

## v1.16.0 — 2026-11-09 (kode dirapikan: modul dipecah per fitur)
Tanpa perubahan perilaku: 9.629 baris kode Go terbukti identik, hanya berpindah file; semua tes & uji halaman lulus.
- **Backend dipecah per fitur**: `main.go` (869 → 191 baris) menjadi app, handler, auth, limiter, static;
  modul ujian → CBT, manajemen ujian, kelulusan; layanan → PPDB, buku tamu, surat; OSIS/kasir → e-voting, kasir;
  LMS → mapel/materi/tugas, kuis, absensi & rekap; kesiswaan → buku induk, absensi harian, kalender.
  Tidak ada lagi file Go di atas ±700 baris.
- **Tampilan dipecah per modul**: mod-layanan.js & mod-ujian.js menjadi 8 file (PPDB, buku tamu, surat, e-voting,
  kasir, CBT, manajemen ujian, kelulusan). Setiap file hanya dimuat bila modulnya aktif; modul khusus pengelola
  hanya untuk admin/guru bertugas tambahan. Guru biasa kini memuat 10 file (sebelumnya 16), siswa 8.
- Helper bersama (unggah foto, tautan WhatsApp, kerangka halaman publik) dipindah ke inti; helper ganda disatukan.
- README: peta struktur kode untuk pengembang.

## v1.15.1 — 2026-11-09
- Pengaturan → Penyimpanan & backup: isian masa simpan dokumen Asisten AI (0 = simpan selamanya).

## v1.15.0 — 2026-11-09 (audit kualitas & bom waktu)
Audit menyeluruh (query/database, backend Go, frontend) dan pengecekan dokumentasi SQLite, modernc/sqlite, dan Go net/http.
- **File upload tidak bisa terhapus tanpa sengaja**: pembersihan file yatim dibatalkan total bila satu query rujukan gagal;
  file yatim kini dipindah ke folder sampah dan baru dihapus setelah 30 hari.
- **Backup lebih aman**: ditulis ke file sementara lalu diganti nama (backup setengah jadi tidak dianggap sah); hari itu
  hanya dianggap selesai bila backup berhasil (gagal → dicoba lagi); backup sebelum rilis/migrasi dibatasi 10 terakhir.
- **WhatsApp**: pesan disewa sebelum dikirim (tidak terkirim berulang bila status gagal disimpan); pesan kehadiran yang
  tertunda >6 jam (lainnya >3 hari) tidak dikirim lagi; notifikasi alfa hanya untuk siswa yang benar-benar tercatat.
- **Kunci kartu QR**: pengaturan tidak lagi disimpan kosong di cache saat database sibuk; kunci kosong = semua kartu ditolak.
- **Pembatas login** tetap bekerja saat diserang dengan banyak kunci acak (entri paling lama diam yang dibuang).
- **Pekerja latar** (pemeliharaan, WhatsApp) tahan panic dan berhenti rapi sebelum database ditutup.
- **Zona waktu** tidak diubah saat server berjalan (balapan data); berlaku setelah server dimulai ulang.
- **Program lama menolak database yang lebih baru**; deploy menunggu hingga 60 detik (migrasi + backup).
- Asisten AI menandai jawaban yang terpotong; nomor PPDB mengikuti tahun pembukaan pendaftaran.
- **Database siap tumbuh bertahun-tahun**: 9 indeks baru; filter bulanan memakai rentang tanggal (bukan LIKE, yang tidak
  memakai indeks); jumlah siswa per kelas/mapel tanpa memindai seluruh pengguna; log aktivitas dihitung di koneksi baca;
  dasbor siswa & daftar CBT siswa hanya memuat data kelasnya; nomor kasir/PPDB/surat memakai indeks; PRAGMA optimize harian.
  Pengaturan baru: masa simpan dokumen Asisten AI (bawaan: selamanya).
- **Frontend**:
  - status ujian/pengumuman memakai jam server (HP/PC lab dengan jam salah tetap bisa mulai ujian);
  - kirim otomatis saat waktu habis dicoba ulang bila jaringan putus;
  - kamera LJK & kiosk buku tamu dimatikan saat pindah halaman;
  - pemindai QR tidak menumpuk di HP lambat;
  - kasir tidak macet bila barang di keranjang dihapus;
  - tautan berisi kode HTML di menu jurnal/catatan/wali ditolak;
  - pesan jelas untuk browser yang terlalu lama;
  - foto LJK tetap terbaca di browser lama;
  - beberapa perbaikan kecil (kalender, kotak konfirmasi, kartu di layar 320 px).
- Tes otomatis baru: pembersihan file aman, kesetaraan hitung siswa, dan rencana query memakai indeks. Lulus dengan -race.

## v1.14.0 — 2026-11-09 (penjaga keutuhan data)
- **Mapel**: kode selalu mengikuti nama (kode yang sudah dipakai sekolah atau kode standar Kurikulum Merdeka); kode
  milik mapel lain ditolak. Nama mapel berupa dropdown dengan pilihan *Mapel lain*; kode terkunci kecuali untuk mapel baru.
- **Data tidak bisa ganda / tidak cocok** (diperiksa di server):
  - NISN 10 digit & NIK 16 digit, unik antar siswa; PPDB menolak NISN/NIK yang sudah terdaftar.
  - Tingkat kelas harus sesuai nama kelas (terisi otomatis, termasuk angka romawi); satu guru wali satu kelas per tahun.
  - Tahun ajaran dua tahun berurutan; tanggal/jam selesai tidak boleh sebelum mulai (tahun ajaran, kegiatan & jadwal
    ujian); jadwal ujian harus di dalam rentang kegiatan.
  - Unik: nomor urut & nama kandidat e-voting, nama barang kasir, ISBN buku, NIP guru (18 digit / NUPTK 16 digit),
    nomor surat keluar, nama ruang ujian per kegiatan, jadwal ujian, dan jurnal mengajar (mapel+tanggal+jam).
  - Ejaan kategori, satuan, lokasi, rak, penerbit, dan disposisi disamakan dengan yang sudah ada.
- Isian *pilih atau ketik* diberi panah dropdown; atribut tersembunyi kini selalu benar-benar menyembunyikan isian.
- Tab yang lama terbuka diberi tahu bila ada versi baru (tombol *Muat ulang*).
- Aturan berlaku saat menyimpan; data lama tidak diubah otomatis.

## v1.13.0 — 2026-11-09 (LJK kertas & pemindai)
- **LJK kertas untuk ujian CBT** (tombol *LJK* di daftar Ujian CBT): untuk sekolah yang perangkatnya belum cukup.
  - **Cetak LJK** per peserta (nama, NIS, kelas, dan QR identitas) atau lembar kosong; A4, maksimal 60 soal, 2–5 pilihan.
  - **Pindai** dengan foto HP/kamera atau pilih banyak foto sekaligus. Pembacaan dilakukan di browser tanpa pustaka luar
    (tetap jalan offline): 4 penanda sudut → lembar diluruskan (miring, perspektif, terbalik, atau foto mendatar),
    lalu kegelapan tiap bulatan diukur. Siswa dikenali otomatis dari QR (browser yang mendukung pembaca QR).
  - Jawaban **kosong** dan **ganda** ditandai; guru bisa mengoreksi tiap nomor sebelum menyimpan. Pratinjau menandai
    bacaan di foto.
  - Nilai dihitung server dengan kunci & bobot yang sama seperti CBT daring dan tampil di *Hasil* ujian (ditandai LJK).
    Menimpa nilai yang sudah ada perlu konfirmasi.
  - Uji akurasi pada lembar sintetis: 39/39 lembar terbaca sempurna (miring hingga 20°, terbalik, buram, berderau,
    cahaya tidak rata, pensil tipis, foto jauh/resolusi rendah, 60 soal × 5 pilihan) dan lembar cetak asli 30/30.

## v1.12.1 — 2026-11-09
- **Mapel tidak bisa ganda**: nama atau kode mapel yang sama ditolak di kelas yang sama, dengan pesan yang jelas
  (sebelumnya nama ganda lolos dan kode ganda menghasilkan galat server).
- Formulir mapel: **Nama mapel** berupa pilihan (daftar umum Kurikulum Merdeka + mapel yang sudah dipakai sekolah,
  tetap boleh mengetik baru) dan **kode terisi otomatis**, sehingga penulisan seragam di semua kelas.

## v1.12.0 — 2026-11-09 (Notifikasi WhatsApp)
- **Notifikasi WhatsApp otomatis** (Pengaturan → Notifikasi WhatsApp, khusus admin), lewat gateway **Fonnte** atau
  gateway **kustom** (Wablas, dll.: URL, nama field nomor & pesan bisa diatur):
  - orang tua: siswa tiba (scan QR / selfie, dengan status terlambat), pulang, alfa di ceklist hari ini,
    serta catatan prestasi/pelanggaran (opsional, bawaan mati);
  - pendaftar PPDB: nomor pendaftaran & PIN (PIN disembunyikan dari log setelah terkirim).
  - Nomor orang tua diambil dari Buku Induk (No. HP orang tua/wali); format 08…/+62… dinormalkan otomatis.
- Templat pesan bisa diubah (kata {nama}, {kelas}, {jam}, {tanggal}, {status}, {sekolah}, dll.).
- Pesan masuk **antrean** dan dikirim berurutan dengan jeda (bawaan 3 detik) agar nomor sekolah tidak dianggap spam;
  satu kejadian = satu pesan (scan ganda tidak mengirim dua kali); gagal kirim dicoba ulang otomatis hingga 5 kali,
  lalu bisa diulang manual dari *Riwayat pengiriman*. Tombol *Kirim tes* untuk memastikan gateway benar.
- Token gateway hanya disimpan di server, tidak pernah dikirim ke browser, dan otomatis dihapus bila alamat gateway
  diganti tanpa token baru. Riwayat pesan ikut dibersihkan sesuai masa simpan log.
- Absensi dan pendaftaran tetap tersimpan walau gateway WA sedang bermasalah.

## v1.11.1 — 2026-11-09
- **Pengaturan berupa submenu**: Identitas sekolah, Hak akses guru, Modul aplikasi, Warna tema, Absensi harian,
  Zona waktu, Asisten AI, Penyimpanan & backup. Satu bagian tampil per halaman (submenu di kiri; di HP berupa baris
  tab yang bisa digeser). Hanya data bagian yang dibuka yang dimuat. Menyimpan satu bagian tidak mengubah bagian lain.
- Ikon tambahan dipindah ke inti aplikasi agar tetap tampil walau modulnya belum dimuat.

## v1.11.0 — 2026-11-09 (Administrasi Guru & Wali Kelas)
- **Modul baru: Administrasi Guru & Wali Kelas** (bisa dimatikan di Pengaturan → Modul).
  - **Wali kelas**: admin menetapkan wali di menu Kelas (Ubah kelas → Wali kelas); kolom wali tampil di daftar kelas.
  - **Jurnal Mengajar**: guru mengisi tanggal, jam ke, mapel/kelas, materi, kegiatan, catatan. Kehadiran terhitung
    otomatis dari absensi mapel. Filter bulan, cetak (dengan tanda tangan kepala sekolah) dan CSV. Admin/BK melihat
    jurnal semua guru.
  - **Catatan Siswa**: prestasi, pelanggaran (dengan poin), catatan wali, catatan BK; kategori siap pilih; cetak & CSV.
    Hak lihat: admin & staf BK/Kesiswaan semua; wali kelas siswa kelasnya (BK milik orang lain tersembunyi);
    guru pengajar hanya catatannya sendiri; siswa hanya prestasi & pelanggaran miliknya (menu *Prestasi & Poin*).
  - **Halaman Wali Kelas**: daftar siswa dengan rekap absensi harian per bulan (hadir, terlambat, S, I, A), poin
    pelanggaran/prestasi, tautan WhatsApp orang tua, tambah catatan cepat, dan cetak rekap.
  - Tugas tambahan baru untuk guru: *BK / Kesiswaan* (melihat semua jurnal & catatan siswa).
- **Asisten AI**: template baru **Prota & Promes** dan **KKTP**.
- Menu: grup *Layanan* dan grup baru tidak lagi muncul di paling atas menu.
- Migrasi database v2 (kolom wali kelas), dengan backup otomatis sebelum migrasi.
- Tes otomatis baru untuk hak akses jurnal, catatan BK, wali kelas, dan siswa.

## v1.10.0 — 2026-11-09 (muat seperlunya)
- **Hanya yang dibutuhkan yang dimuat.** Halaman login cukup memuat inti aplikasi. Setelah masuk, browser hanya
  mengunduh file untuk peran pengguna dan modul yang aktif (mis. guru tidak memuat halaman admin, log aktivitas,
  dan tahun ajaran; modul yang dimatikan tidak dimuat sama sekali dan langsung dimuat bila dinyalakan lagi).
  Halaman publik (PPDB, kelulusan, buku tamu) memuat modulnya saat dibuka.
- **Satu perjalanan lebih sedikit saat membuka aplikasi**: pengaturan sekolah dan data akun ikut di halaman awal
  (tidak di-cache, khusus pengguna itu), sehingga aplikasi tidak menunggu /api/settings dan /api/me dulu.
- Kode bersama (halaman Akun, ganti password, pemilih tahun, daftar berhalaman, unduh CSV) dipindah ke inti.

## v1.9.2 — 2026-11-09 (pindah menu lebih cepat)
- Daftar mapel, kelas, dan tahun ajaran disimpan sementara di browser (2 menit, langsung dikosongkan setiap ada
  penyimpanan/perubahan data), jadi pindah menu tidak lagi menunggu data yang sama diambil ulang lewat internet.
- Permintaan yang sama yang sedang berjalan dipakai bersama; Buku Induk mengambil datanya bersamaan, bukan berurutan.
- Kerangka loading hanya muncul bila halaman butuh lebih dari 150 ms, sehingga perpindahan menu tidak berkedip.

## v1.9.1 — 2026-11-09
- **Kartu sejajar**: kartu di dalam grid/kolom (mis. pilihan jenis dokumen Asisten AI, kartu mapel) tidak lagi turun
  14px mulai kartu kedua; semua kartu dalam satu baris rata atas dan sama tinggi.
- Kartu Asisten AI memakai tata letak sendiri (ikon, judul, keterangan) yang rapi.

## v1.9.0 — 2026-11-09 (warna tema per jenjang)
- **Tidak lagi ungu.** Warna tema mengikuti warna seragam jenjang: **Biru (SMP)** bawaan, **Abu-abu (SMA/SMK)**,
  **Merah (SD)**. Admin memilihnya di *Pengaturan → Warna tema*; berlaku untuk semua pengguna, termasuk halaman login,
  ilustrasi, ikon tab browser, dan warna bilah browser di HP.
- Mode gelap memakai latar netral (sebelumnya keunguan) dan cocok untuk ketiga warna.
- Warna tersimpan di browser sehingga halaman langsung tampil dengan warna sekolah tanpa berkedip.

## v1.8.0 — 2026-11-09 (isian pilih, bukan ketik)
- **Dropdown untuk pilihan yang sudah pasti**: tingkat kelas, tahun ajaran (Pengaturan, PPDB, kegiatan ujian, tambah
  tahun ajaran), tahun perolehan aset & tahun terbit buku, pekerjaan ayah/ibu (pilihan Dapodik), model pembelajaran,
  dimensi profil lulusan (centang), jenis kegiatan LKPD.
- **Dropdown + boleh isi baru** untuk daftar yang bisa bertambah; pilihannya diambil dari data yang sudah ada ditambah
  pilihan umum: mapel (CBT, jadwal ujian, Asisten AI), kelas/fase (Asisten AI), kategori/satuan/sumber dana/lokasi aset,
  kategori/rak/penerbit buku, kategori/satuan barang kasir, kode klasifikasi surat (dengan keterangan), disposisi surat,
  pengawas ruang ujian (daftar guru), jenis surat & alokasi waktu (Asisten AI).
- **Pemilih jam** untuk jam masuk/pulang absensi dan jam jadwal ujian (selalu tersimpan 24 jam, mis. 07:30).
- Nilai lama yang tidak ada di daftar tetap tampil saat formulir diubah, jadi tidak hilang ketika disimpan ulang.
- Endpoint saran `GET /api/saran/<kolom>`: hanya kolom tertentu, dijaga hak akses modul (dengan tes otomatis).

## v1.7.1 — 2026-11-09
- **Asisten AI (guru)**: nama model tidak lagi ditampilkan ke guru (baris info dan kolom Riwayat), dan server tidak
  mengirimkannya ke selain admin. Guru tetap melihat kuota: "Terpakai hari ini: x/y dokumen".

## v1.7.0 — 2026-11-09 (mitigasi prioritas 3)
- **Impor soal dalam satu transaksi** (impor teks CBT & "Masukkan ke…" dari Asisten AI, untuk CBT maupun kuis mapel):
  semua soal divalidasi dulu; bila satu soal salah, tidak ada yang tersimpan dan pesan menyebut nomor soalnya.
  Mengulang impor yang gagal tidak lagi membuat soal ganda.
- **Lebih ringan**: pengaturan modul & hak akses guru disimpan di memori (tidak dibaca ulang dari database di setiap
  permintaan); file tampilan dikirim terkompresi gzip untuk akses LAN langsung; daftar e-voting tanpa query per pemilihan.
- **Hapus akun guru dijaga**: guru yang masih mengampu mapel atau memiliki ujian CBT tidak bisa dihapus (jadikan
  Nonaktif atau pindahkan dulu), agar mapel/ujian tidak kehilangan pemilik.
- **Kode dirapikan**: `api.go` dipecah (inti, akun, LMS) dan `app.js` dipecah (inti + halaman admin/guru/siswa).
- **Tes otomatis** (`go test ./...`, 16 skenario): header CSRF, kunci login, pembatasan peran, kunci jawaban tidak
  bocor ke siswa, token & kelas CBT, batas waktu CBT di server, isolasi antar guru, impor soal atomik, kartu QR palsu,
  nomor kasir setelah batal, penjaga hapus, wajib ganti password, kerahasiaan & sekali-pilih e-voting, batas cek
  kelulusan per NISN, kenaikan kelas. `deploy.sh` dan `build.sh` membatalkan rilis bila ada tes yang gagal.

## v1.6.0 — 2026-11-09 (mitigasi prioritas 2)
- **Lonjakan login**: pengecekan password (bcrypt) dibatasi sesuai jumlah CPU; permintaan lain tetap lancar.
  Uji: 300 login serentak + 100 siswa CBT → simpan jawaban p99 ≤ 60 ms, 0 gagal.
- **Zona waktu sekolah** (WIB/WITA/WIT) di Pengaturan; jam terlambat, tanggal absensi, laporan harian, dan jadwal
  pemeliharaan tidak lagi bergantung zona waktu server. Data zona waktu tertanam di program.
- **Upload besar** (e-book, berkas PPDB, selfie, buku tamu) diberi batas waktu baca 10 menit, tidak putus di 30 detik.
- **Log aktivitas**: setiap perubahan data dicatat otomatis (pengguna, waktu, IP, aksi, status), termasuk login
  berhasil/gagal dan perubahan nilai beserta angkanya. Halaman *Log Aktivitas* (admin) dengan pencarian, filter
  tanggal, dan "hanya gagal". Masa simpan bawaan 365 hari (atur di Pengaturan).
- **Paginasi & pencarian di server** untuk katalog buku, inventaris aset (nilai total dihitung server), dan
  pendaftar PPDB, dengan tombol "Muat lebih banyak".

## v1.5.0 — 2026-11-09 (mitigasi "bom waktu" sebelum live)
- **Tahun ajaran & kenaikan kelas**: setiap kelas milik satu tahun ajaran; menu *Tahun Ajaran* dengan wizard
  kenaikan kelas (naik / tinggal kelas / lulus / keluar per siswa, salin mapel & guru), satu transaksi, tidak bisa
  dijalankan dua kali. Riwayat anggota kelas disimpan sehingga mapel, rekap nilai, absensi, dan hasil CBT tahun
  lalu tetap bisa dibuka (pilihan tahun di Kelas, Mapel, Mapel Saya). Siswa lulus dinonaktifkan dengan tanda angkatan.
- **Migrasi database bernomor versi** (`PRAGMA user_version`) dengan backup otomatis sebelum migrasi dan
  pemeriksaan foreign key; migrasi v1 membangun ulang tabel kelas (nama unik per tahun).
- **Pemeliharaan harian otomatis**: backup harian (`data/backups/harian-*.db`, simpan 14), hapus foto selfie
  (60 hari) & foto buku tamu (365 hari) yang kedaluwarsa, hapus file upload yatim; dapat diatur di Pengaturan.
  Daftar backup bisa diunduh dari Pengaturan; dashboard memperingatkan bila penyimpanan ≥ 80% atau belum ada backup.
- **Shutdown rapi**: saat restart/deploy server menunggu permintaan berjalan (mis. kirim jawaban CBT) hingga 25 detik
  dan menutup database dengan bersih.
- Perbaikan: saran "Lulus" di wizard hanya untuk tingkat akhir jenjang (6/9/12) atau tingkat tertinggi bila ada
  beberapa tingkat; jumlah siswa kelas arsip memakai riwayat anggota kelas.

## v1.4.0 — 2026-11-09
- Tombol tema **ikut sistem / terang / gelap** (menu kiri, bar atas HP, halaman login & publik).
- Nomor versi rilis tertanam di program; skrip `deploy.sh` (rilis tersimpan per versi, backup database
  otomatis, cek kesehatan, rollback otomatis bila gagal) dan `rollback.sh`.
- File tampilan diberi versi otomatis (`app.js?v=<hash>`) agar cache Cloudflare/browser langsung memakai
  versi baru setelah deploy (sebelumnya JS/CSS lama tertahan sampai 4 jam).
- Pintasan "Atur endpoint & API key AI" di halaman Pengaturan.
- Pilihan `-backup <file>` untuk salinan database yang konsisten.

## v1.3.0 — 2026-11-09 (perbaikan hasil audit keamanan & database)
- Batas waktu CBT/kuis ditegakkan server; tebakan token ujian dibatasi.
- Pengaturan AI khusus admin; API key tidak bisa dikirim ke endpoint lain; redirect tidak diikuti; kuota
  tidak bisa diakali permintaan paralel.
- E-voting: perolehan suara dirahasiakan sampai selesai; jeda bilik suara; token kartu khusus staf absensi.
- Guru hanya melihat Buku Induk & absensi kelas yang diajar.
- Penghapusan kelas/mapel/akun yang memiliki riwayat diblokir (gunakan Nonaktif).
- Impor: password acak + wajib ganti password saat login pertama; hashing paralel di luar transaksi.
- Pembatas login ramah NAT (per IP+username), IPv6 per /64, pembersihan berkala; kunci per NISN/no. PPDB;
  PIN PPDB 8 digit; kuota total upload (`-upload-max-gb`).
- Nomor transaksi kasir & PPDB berbasis urutan terbesar; race klik ganda; stok keranjang digabung; index baru.
- Pool baca SQLite terpisah; backup tidak menahan koneksi tulis; file data 0600/0700; CSP; HSTS.
- Tampilan: XSS laporan kasir; ekspor CSV aman dari rumus; logout membersihkan data lokal; kamera pasti mati;
  balapan pindah halaman; hasil AI tidak hilang saat impor soal.
- `-proxy-ip`: header Cloudflare hanya dipercaya dari VM tunnel.

## v1.2.0 — 2026-11-09
- Modul Asisten AI (endpoint kompatibel OpenAI, pilih model dari daftar, streaming, kuota).
- Modul PPDB online, Buku Tamu (kiosk), Surat Masuk/Keluar, E-Voting OSIS, Kasir/Koperasi.
- Akses staf per modul untuk guru/TU; siap online lewat Cloudflare Tunnel (`-proxy`).

## v1.1.0 — 2026-11-09
- Satu sistem banyak modul: Buku Induk, Absensi Harian (QR/ceklist/selfie), Kartu QR bertanda tangan,
  Kalender, CBT, Manajemen Ujian, Kelulusan, Inventaris KIB, Perpustakaan.
- HTTPS lokal untuk kamera; tampilan baru gaya "Split Card Violet".

## v1.0.0 — 2026-11-09
- LMS sekolah Go + SQLite: admin, guru, siswa; materi, tugas, kuis, absensi mapel, rekap nilai.
