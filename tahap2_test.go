package main

import (
	"fmt"
	"strings"
	"testing"
)

// Tahap 2 audit: tes alur penting yang sebelumnya tidak pernah dijalankan (absensi mapel, tugas, impor pengguna,
// aktivasi tahun ajaran, PPDB → siswa, pinjam/kembali buku) dan penolakan per peran untuk modul tambahan.

func siswaID(e *env, u string) int64 { return e.id(`SELECT id FROM users WHERE username=?`, u) }

// Absensi mapel: simpan, ubah (tidak ganda), rekap, hapus; guru lain & siswa luar kelas ditolak.
func TestAbsensiMapelAlurLengkap(t *testing.T) {
	e := newEnv(t)
	g := e.login("guru001", "guru123")
	const tgl = "2026-11-02"
	s1, s2 := siswaID(e, "siswa001"), siswaID(e, "siswa002")
	simpan := func(c *client, course int64, items []map[string]any) (int, []byte) {
		return c.do("PUT", "/api/attendance", map[string]any{"course_id": course, "tanggal": tgl, "items": items})
	}
	if code, b := simpan(g, 1, []map[string]any{{"siswa_id": s1, "status": "hadir"}, {"siswa_id": s2, "status": "sakit"}}); code != 200 {
		t.Fatalf("simpan absensi: %d %s", code, b)
	}
	// simpan ulang tanggal yang sama mengganti, tidak menggandakan
	if code, b := simpan(g, 1, []map[string]any{{"siswa_id": s1, "status": "hadir"}, {"siswa_id": s2, "status": "izin"}}); code != 200 {
		t.Fatalf("ubah absensi: %d %s", code, b)
	}
	if n := e.id(`SELECT COUNT(*) FROM attendance WHERE course_id=1 AND tanggal=?`, tgl); n != 2 {
		t.Fatalf("absensi tanggal yang sama harus 2 baris, ada %d", n)
	}
	var st string
	e.app.db.QueryRow(`SELECT status FROM attendance WHERE course_id=1 AND tanggal=? AND siswa_id=?`, tgl, s2).Scan(&st)
	if st != "izin" {
		t.Fatalf("status harus berubah menjadi izin, dapat %q", st)
	}
	if code, b := simpan(g, 1, []map[string]any{{"siswa_id": s1, "status": "bolos"}}); code != 400 {
		t.Fatalf("status tidak valid harus ditolak: %d %s", code, b)
	}
	// siswa di luar kelas mapel dilewati (tidak pernah tersimpan)
	lain := e.id(`SELECT id FROM users WHERE role='siswa' AND kelas_id<>1 LIMIT 1`)
	simpan(g, 1, []map[string]any{{"siswa_id": lain, "status": "hadir"}})
	if n := e.id(`SELECT COUNT(*) FROM attendance WHERE course_id=1 AND siswa_id=?`, lain); n != 0 {
		t.Fatalf("siswa dari kelas lain tersimpan di absensi mapel ini")
	}
	// guru lain tidak boleh mengisi absensi mapel milik guru001
	g2 := e.login("guru002", "guru123")
	if code, _ := simpan(g2, 1, []map[string]any{{"siswa_id": s1, "status": "alfa"}}); code != 403 {
		t.Fatalf("guru lain harus 403, dapat %d", code)
	}
	var rk struct {
		Attendance []struct {
			SiswaID int64 `json:"siswa_id"`
			Status  string
		}
	}
	if code := g.json("GET", "/api/rekap?course_id=1", nil, &rk); code != 200 || len(rk.Attendance) < 2 {
		t.Fatalf("rekap: %d, %d baris absensi", code, len(rk.Attendance))
	}
	if code, _ := g2.do("GET", "/api/rekap?course_id=1", nil); code != 403 {
		t.Fatalf("guru lain tidak boleh melihat rekap mapel orang lain, dapat %d", code)
	}
	if code, b := g.do("DELETE", "/api/attendance?course_id=1&tanggal="+tgl, nil); code != 200 {
		t.Fatalf("hapus absensi: %d %s", code, b)
	}
	if n := e.id(`SELECT COUNT(*) FROM attendance WHERE course_id=1 AND tanggal=?`, tgl); n != 0 {
		t.Fatalf("absensi tanggal itu harus terhapus, sisa %d", n)
	}
}

// Tugas: guru membuat, siswa mengumpulkan & mengubah sebelum dinilai, guru menilai dalam batas nilai maks,
// guru lain tidak bisa menilai, siswa hanya melihat nilainya sendiri.
func TestTugasBuatKumpulNilai(t *testing.T) {
	e := newEnv(t)
	g := e.login("guru001", "guru123")
	if code, b := g.do("POST", "/api/assignments", map[string]any{"course_id": 1, "judul": "Tugas Uji Alur", "nilai_maks": 50}); code != 200 {
		t.Fatalf("buat tugas: %d %s", code, b)
	}
	asg := e.id(`SELECT id FROM assignments WHERE judul='Tugas Uji Alur'`)
	g2 := e.login("guru002", "guru123")
	if code, _ := g2.do("POST", "/api/assignments", map[string]any{"course_id": 1, "judul": "Tugas Menyusup", "nilai_maks": 100}); code != 403 {
		t.Fatalf("guru lain tidak boleh membuat tugas di mapel orang lain, dapat %d", code)
	}
	s := e.login("siswa001", "siswa123")
	kumpul := func(c *client, teks string) (int, []byte) {
		return c.do("POST", fmt.Sprintf("/api/assignments/%d/submit", asg), map[string]any{"jawaban": teks})
	}
	if code, b := kumpul(s, "jawaban awal"); code != 200 {
		t.Fatalf("kumpul: %d %s", code, b)
	}
	if code, b := kumpul(s, "jawaban revisi"); code != 200 {
		t.Fatalf("ubah sebelum dinilai harus boleh: %d %s", code, b)
	}
	if code, _ := s.do("POST", fmt.Sprintf("/api/assignments/%d/submit", asg), map[string]any{}); code != 400 {
		t.Fatalf("kumpul kosong harus ditolak")
	}
	sub := e.id(`SELECT id FROM submissions WHERE assignment_id=? AND siswa_id=?`, asg, siswaID(e, "siswa001"))
	var jaw string
	e.app.db.QueryRow(`SELECT jawaban FROM submissions WHERE id=?`, sub).Scan(&jaw)
	if jaw != "jawaban revisi" {
		t.Fatalf("jawaban terakhir harus tersimpan, dapat %q", jaw)
	}
	nilai := func(c *client, n float64) (int, []byte) {
		return c.do("PUT", fmt.Sprintf("/api/submissions/%d/grade", sub), map[string]any{"nilai": n, "feedback": "baik"})
	}
	if code, _ := nilai(g, 60); code != 400 {
		t.Fatalf("nilai di atas nilai maks (50) harus ditolak")
	}
	if code, _ := nilai(g2, 40); code != 403 {
		t.Fatalf("guru lain tidak boleh menilai, dapat %d", code)
	}
	if code, b := nilai(g, 45); code != 200 {
		t.Fatalf("menilai: %d %s", code, b)
	}
	if code, _ := kumpul(s, "ganti setelah dinilai"); code != 400 {
		t.Fatalf("setelah dinilai jawaban tidak boleh berubah")
	}
	// siswa lain di kelas yang sama tidak melihat jawaban/nilai siswa001 lewat daftar kumpulan guru
	s2 := e.login("siswa002", "siswa123")
	if code, _ := s2.do("GET", fmt.Sprintf("/api/assignments/%d/submissions", asg), nil); code != 403 {
		t.Fatalf("siswa tidak boleh membuka daftar kumpulan, dapat %d", code)
	}
}

// Impor pengguna massal: baris sah masuk, username ganda & peran tak valid ditolak, guru tidak boleh mengimpor.
func TestImporPengguna(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	var res map[string]any
	code := a.json("POST", "/api/users/import", map[string]any{"role": "siswa", "rows": []map[string]string{
		{"nama": "Siswa Impor Satu", "username": "imp001", "password": "rahasia123", "kelas": "7A"},
		{"nama": "Siswa Impor Dua", "username": "imp002", "password": "rahasia123", "kelas": "7B"},
		{"nama": "Ganda", "username": "siswa001", "password": "rahasia123", "kelas": "7A"},
	}}, &res)
	if code != 200 {
		t.Fatalf("impor: %d %v", code, res)
	}
	if n := e.id(`SELECT COUNT(*) FROM users WHERE username IN ('imp001','imp002') AND role='siswa'`); n != 2 {
		t.Fatalf("2 siswa baru harus masuk, ada %d (%v)", n, res)
	}
	if k := e.id(`SELECT kelas_id FROM users WHERE username='imp001'`); k != e.id(`SELECT id FROM classes WHERE nama='7A'`) {
		t.Fatalf("kelas hasil impor salah")
	}
	if n := e.id(`SELECT COUNT(*) FROM users WHERE username='siswa001'`); n != 1 {
		t.Fatalf("username yang sudah ada tidak boleh tergandakan")
	}
	if code, _ := a.do("POST", "/api/users/import", map[string]any{"role": "admin", "rows": []map[string]string{{"nama": "X", "username": "x1", "password": "rahasia123"}}}); code != 400 {
		t.Fatalf("impor peran admin harus ditolak, dapat %d", code)
	}
	g := e.login("guru001", "guru123")
	if code, _ := g.do("POST", "/api/users/import", map[string]any{"role": "siswa", "rows": []map[string]string{{"nama": "X", "username": "x2", "password": "rahasia123"}}}); code != 403 {
		t.Fatalf("guru tidak boleh mengimpor pengguna, dapat %d", code)
	}
	// akun hasil impor wajib ganti password saat login pertama
	s := e.login("imp001", "rahasia123")
	if code, _ := s.do("GET", "/api/courses", nil); code != 403 {
		t.Fatalf("akun impor harus diminta ganti password dulu, dapat %d", code)
	}
}

// Tahun ajaran: buat tahun baru, aktifkan (tepat satu yang aktif & pengaturan ikut), format salah ditolak.
func TestAktivasiTahunAjaran(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	if code, _ := a.do("POST", "/api/tahun", map[string]any{"nama": "2027/2029"}); code != 400 {
		t.Fatalf("tahun tidak berurutan harus ditolak")
	}
	if code, b := a.do("POST", "/api/tahun", map[string]any{"nama": "2027/2028"}); code != 200 {
		t.Fatalf("buat tahun: %d %s", code, b)
	}
	baru := e.id(`SELECT id FROM academic_years WHERE nama='2027/2028'`)
	if code, b := a.do("PUT", fmt.Sprintf("/api/tahun/%d/aktif", baru), nil); code != 200 {
		t.Fatalf("aktifkan: %d %s", code, b)
	}
	if n := e.id(`SELECT COUNT(*) FROM academic_years WHERE aktif=1`); n != 1 {
		t.Fatalf("harus tepat satu tahun aktif, ada %d", n)
	}
	if aktif := e.id(`SELECT id FROM academic_years WHERE aktif=1`); aktif != baru {
		t.Fatalf("tahun yang diaktifkan bukan yang dipilih")
	}
	var ta string
	e.app.db.QueryRow(`SELECT tahun_ajaran FROM settings WHERE id=1`).Scan(&ta)
	if ta != "2027/2028" {
		t.Fatalf("pengaturan tahun ajaran harus ikut berganti, dapat %q", ta)
	}
	g := e.login("guru001", "guru123")
	if code, _ := g.do("PUT", fmt.Sprintf("/api/tahun/%d/aktif", baru), nil); code != 403 {
		t.Fatalf("guru tidak boleh mengaktifkan tahun ajaran")
	}
}

// PPDB → akun siswa: hanya pendaftar berstatus diterima, sekali saja, dengan profil siswa terisi.
func TestPPDBJadiSiswa(t *testing.T) {
	e := newEnv(t)
	kvSet(e.app.db, "cfg_ppdb", `{"buka":"2020-01-01T00:00:00Z","tutup":"","kuota":0,"jalur":"Zonasi","tahun_ajaran":"","info":""}`)
	e.app.limit = newLoginLimiter()
	if code, b := e.anon().do("POST", "/api/ppdb/daftar", map[string]any{"data": map[string]string{"nama": "Calon Diterima", "tgl_lahir": "2014-05-05", "hp": "081211112222", "jalur": "Zonasi", "jk": "P", "nisn": "0055667788"}}); code != 200 {
		t.Fatalf("daftar: %d %s", code, b)
	}
	id := e.id(`SELECT id FROM ppdb_applicants WHERE nisn='0055667788'`)
	a := e.login("admin", testAdminPW)
	kelas := e.id(`SELECT id FROM classes WHERE nama='7A'`)
	jadi := func() (int, []byte) {
		return a.do("POST", fmt.Sprintf("/api/ppdb/%d/siswa", id), map[string]any{"kelas_id": kelas})
	}
	if code, b := jadi(); code != 400 || !strings.Contains(string(b), "Diterima") {
		t.Fatalf("belum diterima harus ditolak: %d %s", code, b)
	}
	if code, b := a.do("PUT", fmt.Sprintf("/api/ppdb/%d", id), map[string]any{"status": "diterima"}); code != 200 {
		t.Fatalf("ubah status: %d %s", code, b)
	}
	if code, b := jadi(); code != 200 {
		t.Fatalf("buat akun siswa: %d %s", code, b)
	}
	uid := e.id(`SELECT user_id FROM ppdb_applicants WHERE id=?`, id)
	if uid == 0 {
		t.Fatalf("pendaftar harus tertaut ke akun siswa")
	}
	var role string
	var k int64
	e.app.db.QueryRow(`SELECT role, kelas_id FROM users WHERE id=?`, uid).Scan(&role, &k)
	if role != "siswa" || k != kelas {
		t.Fatalf("akun baru harus siswa di kelas 7A, dapat %s/%d", role, k)
	}
	if n := e.id(`SELECT COUNT(*) FROM student_profiles WHERE siswa_id=? AND nisn='0055667788'`, uid); n != 1 {
		t.Fatalf("profil siswa (NISN) harus terisi")
	}
	if code, _ := jadi(); code != 400 {
		t.Fatalf("akun tidak boleh dibuat dua kali")
	}
}

// Perpustakaan: pinjam mengurangi stok tersedia, stok habis ditolak, kembali sekali saja.
func TestPinjamKembaliBuku(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	e.app.db.Exec(`INSERT INTO books (kode, judul, stok, created_at) VALUES ('T-1EKS', 'Buku Satu Eksemplar', 1, ?)`, nowStr())
	pinjam := func(user string) (int, []byte) {
		return a.do("POST", "/api/perpus/pinjam", map[string]any{"user_id": siswaID(e, user), "buku": "T-1EKS"})
	}
	if code, b := pinjam("siswa001"); code != 200 {
		t.Fatalf("pinjam: %d %s", code, b)
	}
	if code, _ := pinjam("siswa002"); code != 400 {
		t.Fatalf("eksemplar terakhir sudah dipinjam: harus ditolak, dapat %d", code)
	}
	loan := e.id(`SELECT id FROM loans WHERE book_id=(SELECT id FROM books WHERE kode='T-1EKS')`)
	if code, b := a.do("POST", fmt.Sprintf("/api/perpus/kembali/%d", loan), nil); code != 200 {
		t.Fatalf("kembali: %d %s", code, b)
	}
	if code, _ := a.do("POST", fmt.Sprintf("/api/perpus/kembali/%d", loan), nil); code != 400 {
		t.Fatalf("mengembalikan dua kali harus ditolak")
	}
	if code, b := pinjam("siswa002"); code != 200 {
		t.Fatalf("setelah dikembalikan bisa dipinjam lagi: %d %s", code, b)
	}
	s := e.login("siswa001", "siswa123")
	if code, _ := s.do("POST", "/api/perpus/pinjam", map[string]any{"user_id": siswaID(e, "siswa001"), "buku": "T-1EKS"}); code != 403 {
		t.Fatalf("siswa tidak boleh mencatat pinjaman, dapat %d", code)
	}
}

// Penolakan per peran untuk modul tambahan (tanpa akses staf, guru & siswa ditolak).
func TestPeranDibatasiModul(t *testing.T) {
	e := newEnv(t)
	s := e.login("siswa001", "siswa123")
	g := e.login("guru001", "guru123")
	paths := []struct{ m, p string }{
		{"POST", "/api/kasir/jual"}, {"GET", "/api/kasir/penjualan"},
		{"GET", "/api/aset"}, {"POST", "/api/aset"},
		{"GET", "/api/ppdb"}, {"PUT", "/api/ppdb/1"}, {"POST", "/api/ppdb/1/siswa"},
		{"GET", "/api/ujian-event"}, {"POST", "/api/ujian-event"}, {"GET", "/api/ujian-event/1"},
		{"POST", "/api/perpus/pinjam"}, {"GET", "/api/perpus/pinjaman"},
		{"POST", "/api/evote/1/bilik"}, {"PUT", "/api/tahun/1/aktif"}, {"POST", "/api/users/import"},
		{"GET", "/api/backups"},
	}
	for _, c := range []*client{s, g} {
		for _, tc := range paths {
			code, b := c.do(tc.m, tc.p, map[string]any{})
			if code != 403 {
				t.Errorf("%s %s: status %d, harusnya 403 (%s)", tc.m, tc.p, code, b)
			}
		}
	}
}

// Mengubah kelas (mis. mengisi wali kelas) tidak boleh mengosongkan tahun ajarannya; daftar kelas selalu
// menampilkan nama tahun dari tahun kelas.
func TestUbahKelasTidakMengosongkanTahunAjaran(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	guru := e.id(`SELECT id FROM users WHERE username='guru001'`)
	if code, b := a.do("PUT", "/api/classes/1", map[string]any{"nama": "7A", "tingkat": "7", "wali_id": guru}); code != 200 {
		t.Fatalf("ubah kelas: %d %s", code, b)
	}
	var ta string
	e.app.db.QueryRow(`SELECT COALESCE(tahun_ajaran,'') FROM classes WHERE id=1`).Scan(&ta)
	if ta != "2026/2027" {
		t.Fatalf("tahun ajaran kelas harus tetap 2026/2027 setelah diubah, dapat %q", ta)
	}
	e.app.db.Exec(`UPDATE classes SET tahun_ajaran=NULL WHERE id=2`) // data lama yang sudah terlanjur kosong
	var ks []struct {
		ID          int64  `json:"id"`
		TahunAjaran string `json:"tahun_ajaran"`
	}
	if code := a.json("GET", "/api/classes", nil, &ks); code != 200 {
		t.Fatalf("daftar kelas: %d", code)
	}
	for _, k := range ks {
		if k.TahunAjaran != "2026/2027" {
			t.Fatalf("kelas %d: tahun ajaran %q", k.ID, k.TahunAjaran)
		}
	}
}
