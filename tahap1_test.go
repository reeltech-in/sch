package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"
	"testing"
	"time"
)

// Kuis dinilai hanya dari soal yang ditampilkan saat siswa mulai: soal yang ditambahkan guru sesudahnya
// tidak membuat jawaban siswa dihitung salah, dan saat melanjutkan siswa melihat soal yang sama.
func TestKuisDinilaiDariSoalYangDitampilkan(t *testing.T) {
	e := newEnv(t)
	g := e.login("guru001", "guru123")
	course := e.id(`SELECT id FROM courses WHERE guru_id=(SELECT id FROM users WHERE username='guru001') LIMIT 1`)
	if code, b := g.do("POST", "/api/quizzes", map[string]any{"course_id": course, "judul": "Kuis Soal Tetap", "durasi_menit": 10, "jumlah_opsi": 4}); code != 200 {
		t.Fatalf("buat kuis: %d %s", code, b)
	}
	quiz := e.id(`SELECT id FROM quizzes WHERE judul='Kuis Soal Tetap'`)
	soal := func(teks string) {
		if code, b := g.do("POST", fmt.Sprintf("/api/quizzes/%d/questions", quiz), map[string]any{"pertanyaan": teks, "opsi_a": "benar", "opsi_b": "salah", "opsi_c": "x", "opsi_d": "y", "jawaban_benar": "a", "bobot": 1}); code != 200 {
			t.Fatalf("tambah soal %s: %d %s", teks, code, b)
		}
	}
	soal("Soal 1")
	soal("Soal 2")

	s := e.login("siswa001", "siswa123")
	var mulai struct {
		Questions []struct{ ID int64 } `json:"questions"`
	}
	if code := s.json("POST", fmt.Sprintf("/api/quizzes/%d/start", quiz), nil, &mulai); code != 200 || len(mulai.Questions) != 2 {
		t.Fatalf("mulai: %d, %d soal", code, len(mulai.Questions))
	}
	soal("Soal 3 (ditambahkan saat siswa mengerjakan)")
	var lanjut struct {
		Questions []struct{ ID int64 } `json:"questions"`
	}
	s.json("POST", fmt.Sprintf("/api/quizzes/%d/start", quiz), nil, &lanjut)
	if len(lanjut.Questions) != 2 {
		t.Fatalf("melanjutkan harus menampilkan 2 soal yang sama, dapat %d", len(lanjut.Questions))
	}
	ans := map[string]string{}
	for _, q := range mulai.Questions {
		ans[fmt.Sprint(q.ID)] = "a"
	}
	var hasil struct{ Skor float64 }
	if code := s.json("POST", fmt.Sprintf("/api/quizzes/%d/submit", quiz), map[string]any{"answers": ans}, &hasil); code != 200 || hasil.Skor != 100 {
		t.Fatalf("semua soal yang ditampilkan dijawab benar harus 100, dapat %d %v", code, hasil.Skor)
	}
}

func pngDataURL(t *testing.T) string {
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 4, 4))); err != nil {
		t.Fatal(err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// PPDB yang ditolak (NISN ganda) tidak menyimpan berkas; pendaftaran yang lolos menyimpannya.
func TestPPDBDitolakTidakMenyimpanBerkas(t *testing.T) {
	e := newEnv(t)
	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	kvSet(e.app.db, "cfg_ppdb", fmt.Sprintf(`{"buka":%q,"tutup":"","kuota":0,"jalur":"Zonasi","tahun_ajaran":"","info":""}`, past))
	files := func() int { es, _ := os.ReadDir(e.app.uploadDir()); return len(es) }
	daftar := func() (int, []byte) {
		e.app.limit = newLoginLimiter()
		return e.anon().do("POST", "/api/ppdb/daftar", map[string]any{
			"data":   map[string]string{"nama": "Calon", "tgl_lahir": "2014-05-05", "hp": "081299998888", "jalur": "Zonasi", "jk": "L", "nisn": "0011223344"},
			"berkas": map[string]string{"kk": pngDataURL(t), "foto": pngDataURL(t)}})
	}
	if code, b := daftar(); code != 200 {
		t.Fatalf("daftar pertama: %d %s", code, b)
	}
	before := files()
	if before != 2 {
		t.Fatalf("pendaftaran sah harus menyimpan 2 berkas, ada %d", before)
	}
	if code, b := daftar(); code != 400 {
		t.Fatalf("NISN ganda harus ditolak: %d %s", code, b)
	}
	if after := files(); after != before {
		t.Fatalf("pendaftaran yang ditolak tidak boleh menyimpan berkas: %d → %d", before, after)
	}
}

// Jatah unggahan publik harian: setelah habis, unggahan publik ditolak (429), unggahan pengguna login tetap jalan.
func TestUnggahanPublikDibatasiHarian(t *testing.T) {
	e := newEnv(t)
	publikUpl.Lock()
	publikUpl.hari, publikUpl.bytes = time.Now().Format("2006-01-02"), publikHarianMaks
	publikUpl.Unlock()
	t.Cleanup(func() { publikUpl.Lock(); publikUpl.hari, publikUpl.bytes = "", 0; publikUpl.Unlock() })
	e.app.limit = newLoginLimiter()
	code, b := e.anon().do("POST", "/api/bukutamu", map[string]any{"nama": "Tamu", "keperluan": "Rapat", "foto": pngDataURL(t)})
	if code != 429 {
		t.Fatalf("jatah publik habis harus 429: %d %s", code, b)
	}
	if _, err := e.app.saveUpload([]byte("\x89PNG\r\n\x1a\n0000"), false); err != nil && !strings.Contains(err.Error(), "Jenis file") {
		t.Fatalf("unggahan internal tidak boleh ikut terblokir: %v", err)
	}
}

// Jawaban yang sudah dinilai tidak bisa ditimpa dengan kirim ulang.
func TestTugasDinilaiTidakBisaDikirimUlang(t *testing.T) {
	e := newEnv(t)
	sub := e.id(`SELECT id FROM submissions WHERE siswa_id=(SELECT id FROM users WHERE username='siswa001') LIMIT 1`)
	asg := e.id(`SELECT assignment_id FROM submissions WHERE id=?`, sub)
	e.app.db.Exec(`UPDATE submissions SET nilai=80 WHERE id=?`, sub)
	var before string
	e.app.db.QueryRow(`SELECT COALESCE(jawaban,'') FROM submissions WHERE id=?`, sub).Scan(&before)
	s := e.login("siswa001", "siswa123")
	if code, b := s.do("POST", fmt.Sprintf("/api/assignments/%d/submit", asg), map[string]any{"jawaban": "jawaban baru"}); code != 400 {
		t.Fatalf("kirim ulang setelah dinilai harus ditolak: %d %s", code, b)
	}
	var after string
	e.app.db.QueryRow(`SELECT COALESCE(jawaban,'') FROM submissions WHERE id=?`, sub).Scan(&after)
	if after != before {
		t.Fatalf("jawaban yang sudah dinilai berubah: %q → %q", before, after)
	}
}

// Buku yang pernah dipinjam tidak bisa dihapus (riwayat & denda tetap utuh); buku tanpa riwayat bisa.
func TestBukuBerriwayatTidakDihapus(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	e.app.db.Exec(`INSERT INTO books (kode, judul, stok, created_at) VALUES ('T-PINJAM', 'Buku Dipinjam', 1, ?), ('T-BARU', 'Buku Baru', 1, ?)`, nowStr(), nowStr())
	dipinjam := e.id(`SELECT id FROM books WHERE judul='Buku Dipinjam'`)
	baru := e.id(`SELECT id FROM books WHERE judul='Buku Baru'`)
	siswa := e.id(`SELECT id FROM users WHERE username='siswa001'`)
	e.app.db.Exec(`INSERT INTO loans (book_id, user_id, tgl_pinjam, jatuh_tempo, tgl_kembali, denda) VALUES (?,?,?,?,?,2000)`, dipinjam, siswa, "2026-01-01", "2026-01-08", "2026-01-10")
	if code, b := a.do("DELETE", fmt.Sprintf("/api/buku/%d", dipinjam), nil); code != 409 || strings.Contains(string(b), `"paksa"`) {
		t.Fatalf("buku berriwayat harus ditolak tanpa opsi paksa: %d %s", code, b)
	}
	if code, b := a.do("DELETE", fmt.Sprintf("/api/buku/%d?paksa=1", dipinjam), nil); code != 409 {
		t.Fatalf("paksa tidak berlaku untuk buku: %d %s", code, b)
	}
	if n := e.id(`SELECT COUNT(*) FROM loans WHERE book_id=?`, dipinjam); n != 1 {
		t.Fatalf("riwayat pinjaman hilang")
	}
	if code, b := a.do("DELETE", fmt.Sprintf("/api/buku/%d", baru), nil); code != 200 {
		t.Fatalf("buku tanpa riwayat harus bisa dihapus: %d %s", code, b)
	}
}

// Reset password oleh admin langsung memutus sesi lama pemilik akun.
func TestResetPasswordMemutusSesi(t *testing.T) {
	e := newEnv(t)
	s := e.login("siswa003", "siswa123")
	if code, _ := s.do("GET", "/api/courses", nil); code != 200 {
		t.Fatalf("siswa harus login dulu")
	}
	a := e.login("admin", testAdminPW)
	id := e.id(`SELECT id FROM users WHERE username='siswa003'`)
	var kelas int64
	e.app.db.QueryRow(`SELECT kelas_id FROM users WHERE id=?`, id).Scan(&kelas)
	if code, b := a.do("PUT", fmt.Sprintf("/api/users/%d", id), map[string]any{"username": "siswa003", "nama_lengkap": "Citra Maharani", "role": "siswa", "status": "aktif", "kelas_id": kelas, "password": "BaruSekali123"}); code != 200 {
		t.Fatalf("reset password: %d %s", code, b)
	}
	if code, _ := s.do("GET", "/api/courses", nil); code != 401 {
		t.Fatalf("sesi lama harus terputus setelah reset password, dapat %d", code)
	}
	if n := e.id(`SELECT COUNT(*) FROM sessions WHERE user_id=?`, id); n != 0 {
		t.Fatalf("masih ada %d sesi", n)
	}
}

// Suara lewat bilik mencatat petugasnya; rekap per petugas hanya untuk pengelola e-voting.
func TestBilikMencatatPetugas(t *testing.T) {
	e := newEnv(t)
	boothJeda = 0
	t.Cleanup(func() { boothJeda = 10 * time.Second })
	el := e.id(`SELECT id FROM elections LIMIT 1`)
	cand := e.id(`SELECT id FROM candidates WHERE election_id=? ORDER BY no LIMIT 1`, el)
	a := e.login("admin", testAdminPW)
	for _, u := range []string{"siswa001", "siswa002"} {
		tok := e.app.cardToken(e.id(`SELECT id FROM users WHERE username=?`, u))
		if code, b := a.do("POST", fmt.Sprintf("/api/evote/%d/bilik", el), map[string]any{"token": tok, "candidate_id": cand}); code != 200 {
			t.Fatalf("bilik %s: %d %s", u, code, b)
		}
	}
	if n := e.id(`SELECT COUNT(*) FROM election_voters WHERE election_id=? AND metode='bilik' AND petugas_id=(SELECT id FROM users WHERE username='admin')`, el); n != 2 {
		t.Fatalf("petugas bilik harus tercatat di 2 suara, dapat %d", n)
	}
	var res struct {
		Bilik []struct {
			Petugas string
			Jumlah  int
		}
	}
	if code := a.json("GET", fmt.Sprintf("/api/evote/%d/hasil", el), nil, &res); code != 200 || len(res.Bilik) != 1 || res.Bilik[0].Jumlah != 2 {
		t.Fatalf("rekap bilik: %d %+v", code, res.Bilik)
	}
}

// Kartu yang ditolak (sudah memilih) tidak membuat petugas bilik harus menunggu jeda.
func TestBilikKartuDitolakTanpaJeda(t *testing.T) {
	e := newEnv(t)
	boothMu.Lock()
	boothLast = map[int64]time.Time{} // catatan jeda dari tes lain (id admin sama di tiap basis data uji)
	boothMu.Unlock()
	el := e.id(`SELECT id FROM elections LIMIT 1`)
	cand := e.id(`SELECT id FROM candidates WHERE election_id=? ORDER BY no LIMIT 1`, el)
	a := e.login("admin", testAdminPW)
	tok := func(u string) string { return e.app.cardToken(e.id(`SELECT id FROM users WHERE username=?`, u)) }
	if code, b := a.do("POST", fmt.Sprintf("/api/evote/%d/bilik", el), map[string]any{"token": tok("siswa001"), "candidate_id": cand}); code != 200 {
		t.Fatalf("suara pertama: %d %s", code, b)
	}
	boothMu.Lock()
	boothLast[e.id(`SELECT id FROM users WHERE username='admin'`)] = time.Time{} // anggap jeda sudah lewat
	boothMu.Unlock()
	if code, _ := a.do("POST", fmt.Sprintf("/api/evote/%d/bilik", el), map[string]any{"token": tok("siswa001"), "candidate_id": cand}); code != 400 {
		t.Fatalf("pemilih yang sama harus ditolak")
	}
	if code, b := a.do("POST", fmt.Sprintf("/api/evote/%d/bilik", el), map[string]any{"token": tok("siswa002"), "candidate_id": cand}); code != 200 {
		t.Fatalf("setelah kartu ditolak, pemilih berikutnya tidak boleh kena jeda: %d %s", code, b)
	}
}
