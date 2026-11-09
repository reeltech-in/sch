package main

import (
	"database/sql"
	"fmt"
	"time"
)

// seedDemo mengisi data contoh (hanya jika belum ada kelas). Untuk mencoba aplikasi.
func seedDemo(db *sql.DB) error {
	var n int
	db.QueryRow(`SELECT COUNT(*) FROM classes`).Scan(&n)
	if n > 0 {
		fmt.Println("Data contoh dilewati: database sudah berisi kelas.")
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := nowStr()
	ex := func(q string, args ...any) int64 {
		if err != nil {
			return 0
		}
		var res sql.Result
		res, err = tx.Exec(q, args...)
		if err != nil {
			return 0
		}
		id, _ := res.LastInsertId()
		return id
	}
	hGuru, _ := hashPassword("guru123")
	hSiswa, _ := hashPassword("siswa123")

	ex(`UPDATE settings SET nama_sekolah='SMP Nusantara Contoh', alamat='Jl. Pendidikan No. 1', tahun_ajaran='2026/2027', kepala_sekolah='Drs. Ahmad Fauzi', nip_kepala='196805121994031002' WHERE id=1`)
	ty := activeYear(tx)
	k7a := ex(`INSERT INTO classes (tahun_id, nama, tingkat, tahun_ajaran, created_at) VALUES (?,'7A','7','2026/2027',?)`, ty, now)
	k7b := ex(`INSERT INTO classes (tahun_id, nama, tingkat, tahun_ajaran, created_at) VALUES (?,'7B','7','2026/2027',?)`, ty, now)
	g1 := ex(`INSERT INTO users (username, password_hash, nama_lengkap, role, nip, created_at) VALUES ('guru001',?,'Sari Wulandari, S.Pd.','guru','198701012010012001',?)`, hGuru, now)
	g2 := ex(`INSERT INTO users (username, password_hash, nama_lengkap, role, created_at) VALUES ('guru002',?,'Budi Santoso, S.Pd.','guru',?)`, hGuru, now)
	names := []string{"Andi Pratama", "Bunga Lestari", "Citra Maharani", "Dimas Saputra", "Eka Putri"}
	var siswa []int64
	for i, nm := range names {
		siswa = append(siswa, ex(`INSERT INTO users (username, password_hash, nama_lengkap, role, kelas_id, created_at) VALUES (?,?,?,'siswa',?,?)`,
			fmt.Sprintf("siswa%03d", i+1), hSiswa, nm, k7a, now))
	}
	ex(`INSERT INTO users (username, password_hash, nama_lengkap, role, kelas_id, created_at) VALUES ('siswa006',?,'Fajar Nugroho','siswa',?,?)`, hSiswa, k7b, now)

	mtk := ex(`INSERT INTO courses (kode, nama, deskripsi, kelas_id, guru_id, created_at) VALUES ('MTK','Matematika','Bilangan, aljabar, dan geometri dasar',?,?,?)`, k7a, g1, now)
	ex(`INSERT INTO courses (kode, nama, kelas_id, guru_id, created_at) VALUES ('IPA','Ilmu Pengetahuan Alam',?,?,?)`, k7a, g2, now)
	ex(`INSERT INTO courses (kode, nama, kelas_id, guru_id, created_at) VALUES ('MTK','Matematika',?,?,?)`, k7b, g1, now)

	m1 := ex(`INSERT INTO materials (course_id, judul, konten, lampiran, urutan, created_at) VALUES (?,?,?,?,1,?)`, mtk, "Bilangan Bulat",
		"Bilangan bulat terdiri dari bilangan negatif, nol, dan bilangan positif.\n\nContoh: ..., -3, -2, -1, 0, 1, 2, 3, ...\n\nPelajari video dan ringkasan berikut, lalu kerjakan latihan di bagian Tugas.",
		`[{"tipe":"link","judul":"Bilangan bulat (Wikipedia)","isi":"https://id.wikipedia.org/wiki/Bilangan_bulat"}]`, now)
	ex(`INSERT INTO materials (course_id, judul, konten, lampiran, urutan, created_at) VALUES (?,?,?,'[]',2,?)`, mtk, "Operasi Hitung Bilangan Bulat",
		"Penjumlahan: 5 + (-3) = 2\nPengurangan: 4 - (-2) = 6\nPerkalian: (-3) × 4 = -12\nPembagian: (-12) : (-3) = 4", now)
	ex(`INSERT INTO material_progress (material_id, siswa_id, completed_at) VALUES (?,?,?)`, m1, siswa[0], now)

	dl := time.Now().Add(72 * time.Hour).UTC().Format(time.RFC3339)
	a1 := ex(`INSERT INTO assignments (course_id, judul, deskripsi, deadline, nilai_maks, created_at) VALUES (?,?,?,?,100,?)`, mtk,
		"Latihan Bilangan Bulat", "Kerjakan soal latihan halaman 12 nomor 1-10. Tulis jawaban beserta caranya.", dl, now)
	ex(`INSERT INTO submissions (assignment_id, siswa_id, jawaban, nilai, feedback, submitted_at) VALUES (?,?,?,85,'Bagus, perhatikan tanda negatif.',?)`, a1, siswa[0], "1. -2\n2. 7\n3. -12", now)
	ex(`INSERT INTO submissions (assignment_id, siswa_id, jawaban, submitted_at) VALUES (?,?,?,?)`, a1, siswa[1], "1. -2\n2. 6\n3. -12", now)

	q1 := ex(`INSERT INTO quizzes (course_id, judul, durasi_menit, jumlah_opsi, created_at) VALUES (?,?,15,4,?)`, mtk, "Kuis Bilangan Bulat", now)
	qs := [][]string{
		{"Hasil dari -5 + 8 adalah …", "-13", "-3", "3", "13", "c"},
		{"Hasil dari (-4) × (-6) adalah …", "-24", "-10", "10", "24", "d"},
		{"Bilangan yang lebih kecil dari -2 adalah …", "-3", "-1", "0", "2", "a"},
	}
	for _, q := range qs {
		ex(`INSERT INTO quiz_questions (quiz_id, pertanyaan, opsi_a, opsi_b, opsi_c, opsi_d, jawaban_benar, bobot, created_at) VALUES (?,?,?,?,?,?,?,1,?)`,
			q1, q[0], q[1], q[2], q[3], q[4], q[5], now)
	}
	ex(`INSERT INTO quiz_attempts (quiz_id, siswa_id, skor, jawaban, started_at, submitted_at, selesai) VALUES (?,?,66.67,'{}',?,?,1)`, q1, siswa[1], now, now)

	for d := 3; d >= 1; d-- {
		tgl := time.Now().AddDate(0, 0, -d).Format("2006-01-02")
		for i, s := range siswa {
			st := "hadir"
			if i == 3 && d == 2 {
				st = "sakit"
			}
			if i == 4 && d == 1 {
				st = "alfa"
			}
			ex(`INSERT INTO attendance (course_id, siswa_id, tanggal, status) VALUES (?,?,?,?)`, mtk, s, tgl, st)
		}
	}
	// --- Data contoh modul ---
	tl := []string{"2013-03-14", "2013-07-02", "2012-11-20", "2013-01-09", "2013-05-25"}
	for i, s := range siswa {
		jk := "L"
		if i == 1 || i == 2 || i == 4 {
			jk = "P"
		}
		ex(`INSERT INTO student_profiles (siswa_id, nisn, jk, tempat_lahir, tgl_lahir, agama, alamat, nama_ayah, nama_ibu, hp_ortu, updated_at) VALUES (?,?,?,?,?,'Islam','Jl. Melati No. 5',?,?,'081234567890',?)`,
			s, fmt.Sprintf("00%08d", 12345670+i), jk, "Bandung", tl[i], "Ayah "+names[i], "Ibu "+names[i], now)
	}
	ym := time.Now().Format("2006-01")
	ex(`INSERT INTO events (judul, mulai, selesai, kategori, created_at) VALUES ('Upacara Hari Pahlawan', ?, ?, 'kegiatan', ?)`, ym+"-10", ym+"-10", now)
	ex(`INSERT INTO events (judul, mulai, selesai, kategori, created_at) VALUES ('Penilaian Tengah Semester', ?, ?, 'ujian', ?)`, ym+"-17", ym+"-21", now)
	ex(`INSERT INTO events (judul, mulai, selesai, kategori, created_at) VALUES ('Libur Semester', ?, ?, 'libur', ?)`, ym+"-28", ym+"-28", now)

	mulai := time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)
	selesai := time.Now().Add(7 * 24 * time.Hour).UTC().Format(time.RFC3339)
	e1 := ex(`INSERT INTO exams (judul, mapel, kelas_ids, mulai, selesai, durasi_menit, jumlah_opsi, token, created_by, created_at) VALUES ('PTS Matematika','Matematika',?,?,?,60,4,'MTK26',?,?)`,
		fmt.Sprintf("[%d]", k7a), mulai, selesai, g1, now)
	cq := [][]string{
		{"Hasil dari 15 + (-7) adalah …", "-22", "-8", "8", "22", "c"},
		{"Bentuk sederhana dari 3a + 5a - 2a adalah …", "6a", "8a", "10a", "4a", "a"},
		{"Keliling persegi dengan sisi 9 cm adalah …", "18 cm", "27 cm", "36 cm", "81 cm", "c"},
		{"Nilai dari 2³ adalah …", "6", "8", "9", "5", "b"},
	}
	for _, q := range cq {
		ex(`INSERT INTO exam_questions (exam_id, pertanyaan, opsi_a, opsi_b, opsi_c, opsi_d, jawaban_benar, created_at) VALUES (?,?,?,?,?,?,?,?)`, e1, q[0], q[1], q[2], q[3], q[4], q[5], now)
	}
	ev := ex(`INSERT INTO exam_events (nama, tahun_ajaran, semester, mulai, selesai, created_at) VALUES ('Penilaian Tengah Semester Ganjil','2026/2027','Ganjil',?,?,?)`, ym+"-17", ym+"-21", now)
	ex(`INSERT INTO exam_rooms (event_id, nama, kapasitas, kolom, pengawas) VALUES (?, 'Ruang 01', 4, 2, 'Budi Santoso, S.Pd.')`, ev)
	ex(`INSERT INTO exam_rooms (event_id, nama, kapasitas, kolom, pengawas) VALUES (?, 'Ruang 02', 4, 2, 'Sari Wulandari, S.Pd.')`, ev)
	ex(`INSERT INTO exam_schedule (event_id, tanggal, jam_mulai, jam_selesai, mapel, kelas) VALUES (?, ?, '07:30', '09:30', 'Matematika', 'Kelas 7')`, ev, ym+"-17")

	books := [][]string{{"BK-0001", "Laskar Pelangi", "Andrea Hirata", "Fiksi", "A1", "3"}, {"BK-0002", "Matematika Kelas VII", "Kemdikbud", "Pelajaran", "B2", "10"},
		{"BK-0003", "Ensiklopedia Sains Anak", "Tim Penulis", "Sains", "C1", "2"}}
	for _, b := range books {
		ex(`INSERT INTO books (kode, judul, pengarang, kategori, rak, stok, created_at) VALUES (?,?,?,?,?,?,?)`, b[0], b[1], b[2], b[3], b[4], b[5], now)
	}
	assets := [][]any{{"INV-B-001", "Laptop Guru", "B", "Elektronik", "Lenovo ThinkPad", "2024", 7500000.0, 2, "unit", "baik", "Ruang Guru", `{"no_pabrik":"LNV-2024-01"}`},
		{"INV-B-002", "Proyektor", "B", "Elektronik", "Epson EB-X51", "2023", 5200000.0, 3, "unit", "rusak_ringan", "Ruang Kelas 7A", `{}`},
		{"INV-C-001", "Gedung Kelas Utama", "C", "Bangunan", "", "2010", 850000000.0, 1, "unit", "baik", "Kompleks sekolah", `{"bertingkat":"Bertingkat","beton":"Beton","luas_lantai":"600"}`},
		{"INV-A-001", "Tanah Sekolah", "A", "Tanah", "", "2005", 1200000000.0, 1, "bidang", "baik", "Jl. Pendidikan No. 1", `{"luas_m2":"2500","status_hak":"Hak Pakai"}`}}
	for _, a := range assets {
		ex(`INSERT INTO assets (kode, nama, kib, kategori, merk, tahun, harga, jumlah, satuan, kondisi, lokasi, detail, sumber_dana, created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,'BOS',?)`,
			append(a, now)...)
	}

	// Kasir & e-voting
	for _, p := range [][]any{{"8991001", "Pensil 2B", "ATK", 3000.0, 2000.0, 50}, {"8991002", "Buku Tulis 38 lbr", "ATK", 5000.0, 3500.0, 40}, {"8991003", "Air Mineral 600ml", "Minuman", 4000.0, 2500.0, 24}} {
		ex(`INSERT INTO products (kode, nama, kategori, harga, harga_beli, stok, satuan, created_at) VALUES (?,?,?,?,?,?,'pcs',?)`, append(p, now)...)
	}
	el := ex(`INSERT INTO elections (judul, deskripsi, mulai, selesai, pemilih, kelas_ids, created_at) VALUES ('Pemilihan Ketua OSIS 2026/2027','Gunakan hak pilihmu!',?,?,'siswa','[]',?)`, mulai, selesai, now)
	ex(`INSERT INTO candidates (election_id, no, nama, wakil, visi, misi) VALUES (?,1,'Citra Maharani','Dimas Saputra','Sekolah hijau dan berprestasi','1. Gerakan kelas bersih
2. Klub sains')`, el)
	ex(`INSERT INTO candidates (election_id, no, nama, wakil, visi, misi) VALUES (?,2,'Eka Putri','Andi Pratama','OSIS yang mendengar','1. Kotak aspirasi digital
2. Festival seni')`, el)

	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	fmt.Println("Data contoh dibuat. Akun guru: guru001 / guru123  ·  Akun siswa: siswa001 / siswa123")
	fmt.Println("PENTING: data contoh hanya untuk mencoba. Gunakan database baru untuk data asli.")
	return nil
}
