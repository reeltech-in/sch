package main

// Tes otomatis alur kritis & hak akses. Setiap tes memakai database demo baru di folder sementara
// dan server httptest dengan rute asli. Jalankan: go test ./...

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

const testAdminPW = "admin-uji-123"

type env struct {
	t   *testing.T
	app *App
	srv *httptest.Server
}

func newEnv(t *testing.T) *env {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "sch.db")
	db, err := openDB(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := ensureAdmin(db, path, false); err != nil {
		t.Fatal(err)
	}
	h, _ := hashPassword(testAdminPW)
	db.Exec(`UPDATE users SET password_hash=? WHERE role='admin'`, h)
	if err := seedDemo(db); err != nil {
		t.Fatal(err)
	}
	rdb, err := openReadDB(path)
	if err != nil {
		t.Fatal(err)
	}
	app := &App{db: db, rdb: rdb, dbPath: path, limit: newLoginLimiter()}
	mux := http.NewServeMux()
	app.routes(mux)
	mux.Handle("GET /", newStaticHandler(app))
	srv := httptest.NewServer(securityHeaders(mux))
	t.Cleanup(func() { srv.Close(); rdb.Close(); db.Close() })
	return &env{t: t, app: app, srv: srv}
}

type client struct {
	e  *env
	hc *http.Client
}

func (e *env) anon() *client {
	jar, _ := cookiejar.New(nil)
	return &client{e: e, hc: &http.Client{Jar: jar}}
}

// login masuk dan gagal-kan tes bila tidak berhasil.
func (e *env) login(user, pw string) *client {
	c := e.anon()
	if code, body := c.do("POST", "/api/login", map[string]string{"username": user, "password": pw}); code != 200 {
		e.t.Fatalf("login %s: %d %s", user, code, body)
	}
	return c
}

func (c *client) raw(method, path string, body any, csrf bool) (int, []byte) {
	c.e.t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, c.e.srv.URL+path, rd)
	req.Header.Set("Content-Type", "application/json")
	if csrf {
		req.Header.Set("X-SCH", "1")
	}
	res, err := c.hc.Do(req)
	if err != nil {
		c.e.t.Fatal(err)
	}
	defer res.Body.Close()
	out, _ := io.ReadAll(res.Body)
	return res.StatusCode, out
}

func (c *client) do(method, path string, body any) (int, []byte) {
	c.e.t.Helper()
	return c.raw(method, path, body, true)
}

func (c *client) json(method, path string, body any, v any) int {
	c.e.t.Helper()
	code, b := c.do(method, path, body)
	if v != nil && code < 300 {
		if err := json.Unmarshal(b, v); err != nil {
			c.e.t.Fatalf("%s %s: json: %v (%s)", method, path, err, b)
		}
	}
	return code
}

func expect(t *testing.T, what string, got, want int, body []byte) {
	t.Helper()
	if got != want {
		t.Errorf("%s: status %d, harusnya %d (%s)", what, got, want, strings.TrimSpace(string(body)))
	}
}

func (e *env) id(q string, args ...any) int64 {
	e.t.Helper()
	var id int64
	if err := e.app.db.QueryRow(q, args...).Scan(&id); err != nil {
		e.t.Fatalf("query %q: %v", q, err)
	}
	return id
}

// ---------------------------------------------------------------------

func TestCSRFHeaderWajib(t *testing.T) {
	e := newEnv(t)
	c := e.anon()
	code, b := c.raw("POST", "/api/login", map[string]string{"username": "admin", "password": testAdminPW}, false)
	expect(t, "login tanpa X-SCH", code, 403, b)
	a := e.login("admin", testAdminPW)
	code, b = a.raw("POST", "/api/classes", map[string]string{"nama": "9Z"}, false)
	expect(t, "POST tanpa X-SCH", code, 403, b)
}

func TestLoginTerkunciSetelahGagalBeruntun(t *testing.T) {
	e := newEnv(t)
	c := e.anon()
	for i := 0; i < 5; i++ {
		code, b := c.do("POST", "/api/login", map[string]string{"username": "guru001", "password": "salah"})
		expect(t, fmt.Sprintf("gagal ke-%d", i+1), code, 401, b)
	}
	code, b := c.do("POST", "/api/login", map[string]string{"username": "guru001", "password": "guru123"})
	expect(t, "password benar saat terkunci", code, 429, b)
	// akun lain dari IP yang sama tetap bisa
	code, b = c.do("POST", "/api/login", map[string]string{"username": "siswa001", "password": "siswa123"})
	expect(t, "akun lain", code, 200, b)
}

func TestTanpaLoginDitolak(t *testing.T) {
	e := newEnv(t)
	c := e.anon()
	for _, p := range []string{"/api/users", "/api/courses", "/api/cbt", "/api/audit"} {
		code, b := c.do("GET", p, nil)
		expect(t, p, code, 401, b)
	}
}

func TestPeranDibatasi(t *testing.T) {
	e := newEnv(t)
	s := e.login("siswa001", "siswa123")
	g := e.login("guru001", "guru123")
	cases := []struct {
		c    *client
		m, p string
	}{
		{s, "GET", "/api/users"},
		{s, "POST", "/api/classes"},
		{s, "GET", "/api/kartu"},
		{s, "GET", "/api/ai/config"},
		{s, "GET", "/api/audit"},
		{g, "GET", "/api/users"},
		{g, "PUT", "/api/settings"},
		{g, "GET", "/api/ai/config"},
		{g, "GET", "/api/kartu"},
		{g, "POST", "/api/tahun/1/naik"},
		{g, "DELETE", "/api/kasir/penjualan/1"},
	}
	for _, tc := range cases {
		code, b := tc.c.do(tc.m, tc.p, map[string]any{})
		if code != 403 {
			t.Errorf("%s %s: status %d, harusnya 403 (%s)", tc.m, tc.p, code, b)
		}
	}
}

func TestSiswaTidakBisaMelihatKunciJawaban(t *testing.T) {
	e := newEnv(t)
	s := e.login("siswa001", "siswa123")
	exam := e.id(`SELECT id FROM exams LIMIT 1`)
	quiz := e.id(`SELECT id FROM quizzes LIMIT 1`)
	for _, p := range []string{fmt.Sprintf("/api/cbt/%d/soal", exam), fmt.Sprintf("/api/quizzes/%d/questions", quiz)} {
		code, b := s.do("GET", p, nil)
		expect(t, p, code, 403, b)
	}
	code, b := s.do("POST", fmt.Sprintf("/api/cbt/%d/start", exam), map[string]string{"token": "MTK26"})
	expect(t, "mulai CBT", code, 200, b)
	if bytes.Contains(b, []byte("jawaban_benar")) {
		t.Errorf("respons mulai CBT memuat kunci jawaban: %s", b)
	}
	code, b = s.do("POST", fmt.Sprintf("/api/quizzes/%d/start", quiz), nil)
	expect(t, "mulai kuis", code, 200, b)
	if bytes.Contains(b, []byte("jawaban_benar")) {
		t.Errorf("respons mulai kuis memuat kunci jawaban: %s", b)
	}
}

func TestCBTTokenSalah(t *testing.T) {
	e := newEnv(t)
	s := e.login("siswa001", "siswa123")
	exam := e.id(`SELECT id FROM exams LIMIT 1`)
	code, b := s.do("POST", fmt.Sprintf("/api/cbt/%d/start", exam), map[string]string{"token": "SALAH"})
	if code == 200 {
		t.Errorf("token salah diterima: %s", b)
	}
	// siswa kelas lain (7B) tidak terdaftar di ujian 7A
	s6 := e.login("siswa006", "siswa123")
	code, b = s6.do("POST", fmt.Sprintf("/api/cbt/%d/start", exam), map[string]string{"token": "MTK26"})
	if code == 200 {
		t.Errorf("siswa kelas lain bisa mulai ujian: %s", b)
	}
}

func TestCBTBatasWaktuDiServer(t *testing.T) {
	e := newEnv(t)
	s := e.login("siswa001", "siswa123")
	exam := e.id(`SELECT id FROM exams LIMIT 1`)
	if code, b := s.do("POST", fmt.Sprintf("/api/cbt/%d/start", exam), map[string]string{"token": "MTK26"}); code != 200 {
		t.Fatalf("mulai: %d %s", code, b)
	}
	// mundurkan waktu mulai melewati durasi (60 menit) + toleransi
	old := time.Now().Add(-3 * time.Hour).UTC().Format(time.RFC3339)
	e.app.db.Exec(`UPDATE exam_attempts SET started_at=? WHERE exam_id=?`, old, exam)
	code, b := s.do("POST", fmt.Sprintf("/api/cbt/%d/save", exam), map[string]any{"answers": map[string]string{}})
	expect(t, "simpan setelah waktu habis", code, 400, b)
	// kumpul tetap diterima, tetapi hanya jawaban tersimpan yang dinilai
	code, b = s.do("POST", fmt.Sprintf("/api/cbt/%d/submit", exam), map[string]any{"answers": map[string]string{"1": "a", "2": "a", "3": "a", "4": "a"}})
	expect(t, "kumpul terlambat", code, 200, b)
	var nilai float64
	e.app.db.QueryRow(`SELECT COALESCE(nilai,0) FROM exam_attempts WHERE exam_id=?`, exam).Scan(&nilai)
	if nilai != 0 {
		t.Errorf("jawaban yang dikirim setelah waktu habis ikut dinilai: nilai=%v", nilai)
	}
}

func TestGuruHanyaMengelolaMilikSendiri(t *testing.T) {
	e := newEnv(t)
	g2 := e.login("guru002", "guru123")
	exam := e.id(`SELECT id FROM exams WHERE created_by=(SELECT id FROM users WHERE username='guru001')`)
	mtk := e.id(`SELECT q.id FROM quizzes q JOIN courses c ON c.id=q.course_id WHERE c.guru_id=(SELECT id FROM users WHERE username='guru001') LIMIT 1`)
	soal := map[string]any{"soal": []map[string]any{{"pertanyaan": "X", "opsi_a": "1", "opsi_b": "2", "opsi_c": "3", "opsi_d": "4", "jawaban_benar": "a"}}}
	code, b := g2.do("GET", fmt.Sprintf("/api/cbt/%d/soal", exam), nil)
	expect(t, "guru lain melihat soal CBT", code, 403, b)
	code, b = g2.do("POST", fmt.Sprintf("/api/cbt/%d/soal/batch", exam), soal)
	expect(t, "guru lain menambah soal CBT", code, 403, b)
	code, b = g2.do("POST", fmt.Sprintf("/api/quizzes/%d/questions/batch", mtk), soal)
	expect(t, "guru lain menambah soal kuis", code, 403, b)
	code, b = g2.do("DELETE", fmt.Sprintf("/api/cbt/%d", exam), nil)
	expect(t, "guru lain menghapus CBT", code, 403, b)
}

func TestImporSoalAtomik(t *testing.T) {
	e := newEnv(t)
	g := e.login("guru001", "guru123")
	exam := e.id(`SELECT id FROM exams LIMIT 1`)
	count := func() int64 { return e.id(`SELECT COUNT(*) FROM exam_questions WHERE exam_id=?`, exam) }
	before := count()
	ok := map[string]any{"pertanyaan": "A", "opsi_a": "1", "opsi_b": "2", "opsi_c": "3", "opsi_d": "4", "jawaban_benar": "b"}
	bad := map[string]any{"pertanyaan": "B", "opsi_a": "1", "opsi_b": "2", "opsi_c": "3", "opsi_d": "4", "jawaban_benar": "e"} // ujian 4 pilihan
	code, b := g.do("POST", fmt.Sprintf("/api/cbt/%d/soal/batch", exam), map[string]any{"soal": []any{ok, bad}})
	expect(t, "batch berisi soal salah", code, 400, b)
	if n := count(); n != before {
		t.Errorf("batch ditolak tetapi %d soal masuk", n-before)
	}
	code, b = g.do("POST", fmt.Sprintf("/api/cbt/%d/soal/batch", exam), map[string]any{"soal": []any{ok, ok}})
	expect(t, "batch valid", code, 200, b)
	if n := count(); n != before+2 {
		t.Errorf("harusnya +2 soal, dapat %+d", n-before)
	}
}

func TestKartuQRPalsuDitolak(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	sid := e.id(`SELECT id FROM users WHERE username='siswa001'`)
	tok := e.app.cardToken(sid)
	if e.app.cardUser(tok) != sid {
		t.Fatal("token kartu asli tidak dikenali")
	}
	forged := []string{
		fmt.Sprintf("S1.%d.000000000000", sid),
		strings.Replace(tok, fmt.Sprintf(".%d.", sid), fmt.Sprintf(".%d.", sid+1), 1),
		fmt.Sprint(sid),
		"",
	}
	for _, f := range forged {
		code, b := a.do("POST", "/api/absen/scan", map[string]string{"token": f})
		expect(t, "token palsu "+f, code, 400, b)
	}
	code, b := a.do("POST", "/api/absen/scan", map[string]string{"token": tok})
	expect(t, "token asli", code, 200, b)
}

func TestKasirNomorTidakBentrokSetelahBatal(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	pid := e.id(`SELECT id FROM products LIMIT 1`)
	sell := func() (int64, string) {
		var r struct {
			ID int64  `json:"id"`
			No string `json:"no"`
		}
		if code := a.json("POST", "/api/kasir/jual", map[string]any{"items": []map[string]any{{"product_id": pid, "qty": 1}}, "bayar": 100000}, &r); code != 200 {
			t.Fatalf("jual: %d", code)
		}
		return r.ID, r.No
	}
	stok := func() int64 { return e.id(`SELECT stok FROM products WHERE id=?`, pid) }
	s0 := stok()
	id1, no1 := sell()
	_, no2 := sell()
	if code, b := a.do("DELETE", fmt.Sprintf("/api/kasir/penjualan/%d", id1), nil); code != 200 {
		t.Fatalf("batal: %d %s", code, b)
	}
	_, no3 := sell()
	if no3 == no1 || no3 == no2 {
		t.Errorf("nomor transaksi bentrok: %s %s %s", no1, no2, no3)
	}
	if got := stok(); got != s0-2 {
		t.Errorf("stok %d, harusnya %d (batal mengembalikan stok)", got, s0-2)
	}
}

func TestHapusDijagaReferensi(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	k7a := e.id(`SELECT id FROM classes WHERE nama='7A'`)
	g1 := e.id(`SELECT id FROM users WHERE username='guru001'`)
	code, b := a.do("DELETE", fmt.Sprintf("/api/classes/%d", k7a), nil)
	if code < 400 {
		t.Errorf("kelas berisi siswa bisa dihapus: %d %s", code, b)
	}
	code, b = a.do("DELETE", fmt.Sprintf("/api/users/%d", g1), nil)
	if code < 400 {
		t.Errorf("guru yang punya mapel/ujian bisa dihapus: %d %s", code, b)
	}
	admin := e.id(`SELECT id FROM users WHERE role='admin'`)
	code, b = a.do("DELETE", fmt.Sprintf("/api/users/%d", admin), nil)
	if code < 400 {
		t.Errorf("admin bisa menghapus dirinya sendiri: %d %s", code, b)
	}
}

func TestGantiPasswordWajibMemblokirAPI(t *testing.T) {
	e := newEnv(t)
	e.app.db.Exec(`UPDATE users SET ganti_pw=1 WHERE username='siswa001'`)
	s := e.login("siswa001", "siswa123")
	code, b := s.do("GET", "/api/courses", nil)
	expect(t, "API lain sebelum ganti password", code, 403, b)
	code, b = s.do("GET", "/api/me", nil)
	expect(t, "/api/me", code, 200, b)
	code, b = s.do("POST", "/api/password", map[string]string{"old": "siswa123", "new": "rahasia-baru-1"})
	expect(t, "ganti password", code, 200, b)
	code, b = s.do("GET", "/api/courses", nil)
	expect(t, "API setelah ganti password", code, 200, b)
}

func TestEvotingRahasiaDanSekaliPilih(t *testing.T) {
	e := newEnv(t)
	el := e.id(`SELECT id FROM elections LIMIT 1`)
	cand := e.id(`SELECT id FROM candidates WHERE election_id=? ORDER BY no LIMIT 1`, el)
	s := e.login("siswa001", "siswa123")
	code, b := s.do("POST", fmt.Sprintf("/api/evote/%d/pilih", el), map[string]any{"candidate_id": cand})
	expect(t, "memilih", code, 200, b)
	code, b = s.do("POST", fmt.Sprintf("/api/evote/%d/pilih", el), map[string]any{"candidate_id": cand})
	if code == 200 {
		t.Errorf("bisa memilih dua kali: %s", b)
	}
	code, b = s.do("GET", fmt.Sprintf("/api/evote/%d/hasil", el), nil)
	expect(t, "siswa melihat hasil sebelum selesai", code, 403, b)
	// admin pun tidak melihat angka sebelum pemilihan selesai
	a := e.login("admin", testAdminPW)
	var res struct {
		Kandidat []map[string]any `json:"kandidat"`
	}
	if code := a.json("GET", fmt.Sprintf("/api/evote/%d/hasil", el), nil, &res); code != 200 {
		t.Fatalf("admin hasil: %d", code)
	}
	for _, k := range res.Kandidat {
		if k["suara"] != nil {
			t.Errorf("perolehan suara terlihat sebelum selesai: %v", k)
		}
	}
	// pilihan tidak tersimpan bersama identitas pemilih
	var cols string
	e.app.db.QueryRow(`SELECT group_concat(name) FROM pragma_table_info('election_voters')`).Scan(&cols)
	if strings.Contains(cols, "candidate") || strings.Contains(cols, "kandidat") {
		t.Errorf("tabel pemilih menyimpan pilihan: %s", cols)
	}
}

func TestKelulusanDibatasiPerNISN(t *testing.T) {
	e := newEnv(t)
	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	b, _ := json.Marshal(map[string]any{"buka": past})
	kvSet(e.app.db, "cfg_kelulusan", string(b))
	c := e.anon()
	var last int
	for i := 0; i < 7; i++ {
		last, _ = c.do("POST", "/api/kelulusan/cek", map[string]string{"nisn": "0012345678", "tgl_lahir": "2010-01-01"})
	}
	if last != 429 {
		t.Errorf("percobaan ke-7 untuk NISN yang sama: status %d, harusnya 429", last)
	}
}

func TestKenaikanKelas(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	ty := e.id(`SELECT id FROM academic_years WHERE aktif=1`)
	k7a := e.id(`SELECT id FROM classes WHERE nama='7A'`)
	k7b := e.id(`SELECT id FROM classes WHERE nama='7B'`)
	// tidak menyebut semua kelas berisi siswa -> ditolak
	code, b := a.do("POST", fmt.Sprintf("/api/tahun/%d/naik", ty), map[string]any{
		"tujuan": map[string]any{"nama": "2027/2028"}, "kelas": []map[string]any{{"kelas_id": k7a, "aksi": "naik", "ke_nama": "8A"}},
	})
	if code < 400 {
		t.Errorf("kenaikan tanpa semua kelas diterima: %d %s", code, b)
	}
	code, b = a.do("POST", fmt.Sprintf("/api/tahun/%d/naik", ty), map[string]any{
		"tujuan": map[string]any{"nama": "2027/2028"}, "kelas": []map[string]any{{"kelas_id": k7a, "aksi": "naik", "ke_nama": "8A"}, {"kelas_id": k7b, "aksi": "naik", "ke_nama": "8B"}},
	})
	expect(t, "kenaikan kelas", code, 200, b)
	if code != 200 {
		return
	}
	var kelas string
	e.app.db.QueryRow(`SELECT c.nama FROM users u JOIN classes c ON c.id=u.kelas_id WHERE u.username='siswa001'`).Scan(&kelas)
	if kelas != "8A" {
		t.Errorf("siswa001 di kelas %q, harusnya 8A", kelas)
	}
	if n := e.id(`SELECT COUNT(*) FROM class_members WHERE kelas_id=?`, k7a); n != 5 {
		t.Errorf("riwayat anggota 7A = %d, harusnya 5", n)
	}
	if e.id(`SELECT aktif FROM academic_years WHERE id=?`, ty) != 0 {
		t.Error("tahun lama masih aktif")
	}
}

func TestModelAITersembunyiDariGuru(t *testing.T) {
	e := newEnv(t)
	kvSet(e.app.db, "cfg_ai", `{"base_url":"http://contoh.local/v1","model":"model-rahasia","kuota_harian":30,"suhu":0.7}`)
	g := e.login("guru001", "guru123")
	code, b := g.do("GET", "/api/ai/status", nil)
	expect(t, "status AI guru", code, 200, b)
	if bytes.Contains(b, []byte("model-rahasia")) {
		t.Errorf("guru melihat nama model: %s", b)
	}
	if !bytes.Contains(b, []byte(`"kuota":30`)) {
		t.Errorf("guru tidak melihat kuota: %s", b)
	}
	a := e.login("admin", testAdminPW)
	if _, b := a.do("GET", "/api/ai/status", nil); !bytes.Contains(b, []byte("model-rahasia")) {
		t.Errorf("admin tidak melihat model: %s", b)
	}
}

func TestSaranIsianDijagaHakAkses(t *testing.T) {
	e := newEnv(t)
	g := e.login("guru001", "guru123")
	a := e.login("admin", testAdminPW)
	for _, k := range []string{"aset.lokasi", "buku.rak", "produk.kategori", "surat.disposisi", "ujian.guru"} {
		code, b := g.do("GET", "/api/saran/"+k, nil)
		expect(t, "guru tanpa akses "+k, code, 403, b)
		code, b = a.do("GET", "/api/saran/"+k, nil)
		expect(t, "admin "+k, code, 200, b)
	}
	code, b := a.do("GET", "/api/saran/users.password_hash", nil)
	if code == 200 {
		t.Errorf("kolom di luar daftar bisa dibaca: %s", b)
	}
}

func TestWarnaTema(t *testing.T) {
	e := newEnv(t)
	var s map[string]any
	e.anon().json("GET", "/api/settings", nil, &s)
	if s["warna"] != "biru" {
		t.Errorf("warna bawaan %v, harusnya biru", s["warna"])
	}
	g := e.login("guru001", "guru123")
	code, b := g.do("PUT", "/api/config/tampilan", map[string]string{"warna": "abu"})
	expect(t, "guru mengubah warna", code, 403, b)
	a := e.login("admin", testAdminPW)
	code, b = a.do("PUT", "/api/config/tampilan", map[string]string{"warna": "ungu\"><script>"})
	expect(t, "warna tidak dikenal", code, 400, b)
	code, b = a.do("PUT", "/api/config/tampilan", map[string]string{"warna": "abu"})
	expect(t, "admin mengubah warna", code, 200, b)
	e.anon().json("GET", "/api/settings", nil, &s)
	if s["warna"] != "abu" {
		t.Errorf("warna setelah diubah %v, harusnya abu", s["warna"])
	}
}

func TestHalamanAwalBerisiDataAwal(t *testing.T) {
	e := newEnv(t)
	get := func(c *client) (string, http.Header) {
		res, err := c.hc.Get(e.srv.URL + "/")
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		b, _ := io.ReadAll(res.Body)
		return string(b), res.Header
	}
	boot := func(html string) map[string]any {
		i := strings.Index(html, `id="boot">`)
		if i < 0 {
			t.Fatal("data awal tidak ada di halaman")
		}
		s := html[i+len(`id="boot">`):]
		s = s[:strings.Index(s, "</script>")]
		var m map[string]any
		if err := json.Unmarshal([]byte(s), &m); err != nil {
			t.Fatalf("data awal bukan JSON: %v", err)
		}
		return m
	}
	html, h := get(e.anon())
	if cc := h.Get("Cache-Control"); !strings.Contains(cc, "no-store") {
		t.Errorf("halaman awal boleh di-cache: %q", cc)
	}
	b := boot(html)
	if b["me"] != nil {
		t.Errorf("pengunjung tanpa login mendapat data akun: %v", b["me"])
	}
	if b["settings"] == nil || b["files"] == nil {
		t.Errorf("pengaturan/daftar file tidak ada: %v", b)
	}
	if strings.Contains(html, "pages-admin.js") && !strings.Contains(html, `"pages-admin.js":`) {
		t.Error("file halaman admin masih dimuat langsung di halaman login")
	}
	html, _ = get(e.login("guru001", "guru123"))
	me, _ := boot(html)["me"].(map[string]any)
	if me["username"] != "guru001" {
		t.Errorf("data akun di halaman awal salah: %v", me)
	}
	if strings.Contains(html, "password") && strings.Contains(html, "$2a$") {
		t.Error("hash password bocor ke halaman")
	}
}

func TestAdmGuruWaliDanCatatan(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	g1 := e.id(`SELECT id FROM users WHERE username='guru001'`) // mengajar 7A & 7B
	g2 := e.id(`SELECT id FROM users WHERE username='guru002'`) // mengajar IPA 7A
	k7a := e.id(`SELECT id FROM classes WHERE nama='7A'`)
	k7b := e.id(`SELECT id FROM classes WHERE nama='7B'`)
	s1 := e.id(`SELECT id FROM users WHERE username='siswa001'`) // 7A
	s6 := e.id(`SELECT id FROM users WHERE username='siswa006'`) // 7B

	// admin menetapkan wali kelas 7A = guru001; siswa tidak boleh jadi wali
	code, b := a.do("PUT", fmt.Sprintf("/api/classes/%d", k7a), map[string]any{"nama": "7A", "tingkat": "7", "wali_id": g1})
	expect(t, "set wali", code, 200, b)
	code, b = a.do("PUT", fmt.Sprintf("/api/classes/%d", k7b), map[string]any{"nama": "7B", "tingkat": "7", "wali_id": s1})
	expect(t, "siswa sebagai wali", code, 400, b)

	w := e.login("guru001", "guru123")
	o := e.login("guru002", "guru123")
	s := e.login("siswa001", "siswa123")

	// halaman wali kelas: hanya wali & admin
	code, b = w.do("GET", fmt.Sprintf("/api/walikelas/%d", k7a), nil)
	expect(t, "wali membuka kelasnya", code, 200, b)
	code, b = o.do("GET", fmt.Sprintf("/api/walikelas/%d", k7a), nil)
	expect(t, "bukan wali membuka kelas", code, 403, b)
	var wk []map[string]any
	o.json("GET", "/api/walikelas", nil, &wk)
	if len(wk) != 0 {
		t.Errorf("guru bukan wali melihat %d kelas wali", len(wk))
	}

	note := func(c *client, siswa int64, jenis string, poin int) (int, []byte) {
		return c.do("POST", "/api/catatan", map[string]any{"siswa_id": siswa, "tanggal": "2026-11-09", "jenis": jenis, "poin": poin, "keterangan": "uji " + jenis})
	}
	// guru002 mengajar 7A -> boleh mencatat pelanggaran siswa 7A, tidak boleh 7B, tidak boleh BK
	code, b = note(o, s1, "pelanggaran", 10)
	expect(t, "guru pengajar mencatat", code, 200, b)
	code, b = note(o, s6, "pelanggaran", 10)
	expect(t, "guru mencatat siswa kelas lain", code, 403, b)
	code, b = note(o, s1, "bk", 0)
	expect(t, "guru biasa menulis catatan BK", code, 403, b)
	// wali menulis catatan BK & prestasi
	code, b = note(w, s1, "bk", 0)
	expect(t, "wali menulis BK", code, 200, b)
	code, b = note(w, s1, "prestasi", 5)
	expect(t, "wali menulis prestasi", code, 200, b)
	// admin menulis BK yang bukan milik wali
	code, b = note(a, s1, "bk", 0)
	expect(t, "admin menulis BK", code, 200, b)

	var rows []map[string]any
	o.json("GET", "/api/catatan", nil, &rows)
	if len(rows) != 1 {
		t.Errorf("guru pengajar harusnya hanya melihat 1 catatan miliknya, dapat %d", len(rows))
	}
	w.json("GET", "/api/catatan", nil, &rows)
	bk := 0
	for _, r := range rows {
		if r["jenis"] == "bk" {
			bk++
		}
	}
	if len(rows) != 3 || bk != 1 {
		t.Errorf("wali harusnya melihat 3 catatan (BK milik admin tersembunyi), dapat %d dengan %d BK", len(rows), bk)
	}
	// guru002 tidak boleh mengubah/menghapus catatan wali
	var mine int64
	for _, r := range rows {
		if r["jenis"] == "prestasi" {
			mine = int64(r["id"].(float64))
		}
	}
	code, b = o.do("DELETE", fmt.Sprintf("/api/catatan/%d", mine), nil)
	expect(t, "hapus catatan orang lain", code, 403, b)

	// siswa: hanya prestasi & pelanggaran miliknya, tanpa BK
	code, b = s.do("GET", "/api/catatan", nil)
	expect(t, "siswa membuka semua catatan", code, 403, b)
	s.json("GET", "/api/catatan/me", nil, &rows)
	for _, r := range rows {
		if r["jenis"] == "bk" || r["jenis"] == "catatan" {
			t.Errorf("siswa melihat catatan %v", r["jenis"])
		}
	}
	if len(rows) != 2 {
		t.Errorf("siswa harusnya melihat 2 catatan (prestasi+pelanggaran), dapat %d", len(rows))
	}
	// rekap poin di halaman wali
	var d struct {
		Siswa []map[string]any `json:"siswa"`
	}
	w.json("GET", fmt.Sprintf("/api/walikelas/%d", k7a), nil, &d)
	for _, r := range d.Siswa {
		if int64(r["id"].(float64)) == s1 && (r["poin_pelanggaran"].(float64) != 10 || r["poin_prestasi"].(float64) != 5) {
			t.Errorf("rekap poin salah: %v", r)
		}
	}
	_ = g2
}

func TestJurnalMengajar(t *testing.T) {
	e := newEnv(t)
	w := e.login("guru001", "guru123")
	o := e.login("guru002", "guru123")
	a := e.login("admin", testAdminPW)
	mtk := e.id(`SELECT id FROM courses WHERE guru_id=(SELECT id FROM users WHERE username='guru001') LIMIT 1`)
	code, b := w.do("POST", "/api/jurnal", map[string]any{"course_id": mtk, "tanggal": "2026-11-09", "jam_ke": "1-2", "materi": "Bilangan bulat"})
	expect(t, "isi jurnal", code, 200, b)
	code, b = o.do("POST", "/api/jurnal", map[string]any{"course_id": mtk, "tanggal": "2026-11-09", "materi": "Bukan mapel saya"})
	expect(t, "jurnal untuk mapel guru lain", code, 400, b)
	var rows []map[string]any
	o.json("GET", "/api/jurnal?bulan=2026-11", nil, &rows)
	if len(rows) != 0 {
		t.Errorf("guru lain melihat %d jurnal", len(rows))
	}
	a.json("GET", "/api/jurnal?bulan=2026-11", nil, &rows)
	if len(rows) != 1 {
		t.Fatalf("admin harusnya melihat 1 jurnal, dapat %d", len(rows))
	}
	id := int64(rows[0]["id"].(float64))
	code, b = o.do("DELETE", fmt.Sprintf("/api/jurnal/%d", id), nil)
	expect(t, "guru lain menghapus jurnal", code, 403, b)
	code, b = w.do("PUT", fmt.Sprintf("/api/jurnal/%d", id), map[string]any{"course_id": mtk, "tanggal": "2026-11-10", "materi": "Bilangan bulat (lanjutan)"})
	expect(t, "ubah jurnal sendiri", code, 200, b)
	// guru yang punya jurnal tidak bisa dihapus (arsip tetap utuh)
	code, b = a.do("DELETE", fmt.Sprintf("/api/users/%d", e.id(`SELECT id FROM users WHERE username='guru001'`)), nil)
	if code < 400 {
		t.Errorf("guru dengan jurnal terhapus: %s", b)
	}
}

// gateway WA tiruan: mencatat permintaan; gagal bila fail > 0
type fakeWA struct {
	srv  *httptest.Server
	got  []map[string]string
	auth []string
	fail int
}

func newFakeWA(t *testing.T) *fakeWA {
	f := &fakeWA{}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		m := map[string]string{}
		for k := range r.PostForm {
			m[k] = r.PostForm.Get(k)
		}
		f.got = append(f.got, m)
		f.auth = append(f.auth, r.Header.Get("Authorization"))
		if f.fail > 0 {
			f.fail--
			w.Write([]byte(`{"status":false,"reason":"device disconnected"}`))
			return
		}
		w.Write([]byte(`{"status":true,"detail":"success! message in queue"}`))
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func TestNotifikasiWhatsApp(t *testing.T) {
	e := newEnv(t)
	gw := newFakeWA(t)
	old := waFonnteURL
	waFonnteURL = gw.srv.URL
	t.Cleanup(func() { waFonnteURL = old })

	a := e.login("admin", testAdminPW)
	g := e.login("guru001", "guru123")
	// hanya admin
	code, b := g.do("GET", "/api/wa/config", nil)
	expect(t, "guru membuka pengaturan WA", code, 403, b)
	code, b = a.do("PUT", "/api/config/wa", map[string]any{"provider": "kustom", "url": "http://jahat.example"})
	expect(t, "lewat /api/config umum", code, 404, b)

	code, b = a.do("PUT", "/api/wa/config", map[string]any{"aktif": true, "provider": "fonnte", "token": "TOKEN-RAHASIA", "kirim_masuk": true, "kirim_ppdb": true})
	expect(t, "simpan pengaturan WA", code, 200, b)
	_, b = a.do("GET", "/api/wa/config", nil)
	if bytes.Contains(b, []byte("TOKEN-RAHASIA")) || !bytes.Contains(b, []byte(`"ada_token":true`)) {
		t.Errorf("token bocor atau tidak tersimpan: %s", b)
	}

	// nomor orang tua siswa001
	s1 := e.id(`SELECT id FROM users WHERE username='siswa001'`)
	e.app.db.Exec(`INSERT INTO student_profiles (siswa_id, hp_ortu) VALUES (?, '0812-3456-7890') ON CONFLICT (siswa_id) DO UPDATE SET hp_ortu=excluded.hp_ortu`, s1)
	// scan kartu dua kali: hanya satu pesan "masuk"
	tok := e.app.cardToken(s1)
	a.do("POST", "/api/absen/scan", map[string]string{"token": tok, "mode": "masuk"})
	a.do("POST", "/api/absen/scan", map[string]string{"token": tok, "mode": "masuk"})
	if n := e.id(`SELECT COUNT(*) FROM wa_outbox WHERE jenis='masuk'`); n != 1 {
		t.Fatalf("pesan masuk di antrean: %d, harusnya 1", n)
	}
	// gateway gagal sekali -> dicoba ulang, lalu terkirim
	gw.fail = 1
	e.app.waSendNext()
	var st string
	var try int
	e.app.db.QueryRow(`SELECT status, percobaan FROM wa_outbox WHERE jenis='masuk'`).Scan(&st, &try)
	if st != "antre" || try != 1 {
		t.Errorf("setelah gagal: status=%s percobaan=%d", st, try)
	}
	e.app.db.Exec(`UPDATE wa_outbox SET next_at='2000-01-01T00:00:00Z'`)
	e.app.waSendNext()
	e.app.db.QueryRow(`SELECT status FROM wa_outbox WHERE jenis='masuk'`).Scan(&st)
	if st != "terkirim" {
		t.Errorf("pesan tidak terkirim setelah coba ulang: %s", st)
	}
	last := gw.got[len(gw.got)-1]
	if last["target"] != "6281234567890" || !strings.Contains(last["message"], "Andi Pratama") || gw.auth[len(gw.auth)-1] != "TOKEN-RAHASIA" {
		t.Errorf("isi kiriman salah: %v auth=%q", last, gw.auth[len(gw.auth)-1])
	}

	// PPDB: PIN dikirim, lalu disembunyikan dari log
	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	kvSet(e.app.db, "cfg_ppdb", fmt.Sprintf(`{"buka":%q,"tutup":"","kuota":0,"jalur":"Zonasi","tahun_ajaran":"","info":""}`, past))
	var reg struct {
		PIN string `json:"pin"`
	}
	if code := e.anon().json("POST", "/api/ppdb/daftar", map[string]any{"data": map[string]string{"nama": "Calon Siswa", "tgl_lahir": "2014-05-05", "hp": "081299998888", "jalur": "Zonasi", "jk": "L"}}, &reg); code != 200 || reg.PIN == "" {
		t.Fatalf("daftar PPDB gagal: %d", code)
	}
	e.app.waSendNext()
	if m := gw.got[len(gw.got)-1]["message"]; !strings.Contains(m, reg.PIN) {
		t.Errorf("PIN tidak ada di pesan PPDB: %q", m)
	}
	var pesan string
	e.app.db.QueryRow(`SELECT pesan FROM wa_outbox WHERE jenis='ppdb'`).Scan(&pesan)
	if strings.Contains(pesan, reg.PIN) {
		t.Error("PIN masih tersimpan di log setelah terkirim")
	}

	// ganti gateway ke host lain tanpa token baru -> token lama dihapus
	a.do("PUT", "/api/wa/config", map[string]any{"provider": "kustom", "url": "https://gateway-lain.example/send"})
	if kvGet(e.app.db, "wa_token") != "" {
		t.Error("token lama tetap ada setelah gateway berganti host")
	}
	// nonaktif -> tidak ada antrean baru
	a.do("PUT", "/api/wa/config", map[string]any{"aktif": false})
	a.do("POST", "/api/absen/scan", map[string]string{"token": tok, "mode": "pulang"})
	if n := e.id(`SELECT COUNT(*) FROM wa_outbox WHERE jenis='pulang'`); n != 0 {
		t.Errorf("notifikasi nonaktif tetapi %d pesan pulang diantrekan", n)
	}
}

func TestMapelTidakGandaDalamKelas(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	k7a := e.id(`SELECT id FROM classes WHERE nama='7A'`)
	k7b := e.id(`SELECT id FROM classes WHERE nama='7B'`)
	code, b := a.do("POST", "/api/courses", map[string]any{"kode": "MTK2", "nama": "matematika", "kelas_id": k7a})
	expect(t, "nama sama (beda huruf besar/kecil) di kelas yang sama", code, 400, b)
	code, b = a.do("POST", "/api/courses", map[string]any{"kode": "mtk", "nama": "Matematika Lanjut", "kelas_id": k7a})
	expect(t, "kode sama di kelas yang sama", code, 400, b)
	code, b = a.do("POST", "/api/courses", map[string]any{"kode": "BIN", "nama": "Bahasa Indonesia", "kelas_id": k7b})
	expect(t, "mapel baru di kelas lain", code, 200, b)
	// mengubah mapel sendiri (nama tetap) tidak dianggap ganda
	mtk := e.id(`SELECT id FROM courses WHERE kelas_id=? AND kode='MTK'`, k7a)
	code, b = a.do("PUT", fmt.Sprintf("/api/courses/%d", mtk), map[string]any{"kode": "MTK", "nama": "Matematika", "kelas_id": k7a})
	expect(t, "simpan ulang mapel sendiri", code, 200, b)
}

func TestLJKSimpanNilai(t *testing.T) {
	e := newEnv(t)
	g := e.login("guru001", "guru123") // pembuat ujian demo
	o := e.login("guru002", "guru123")
	exam := e.id(`SELECT id FROM exams LIMIT 1`)                 // 4 soal, kunci c,a,c,b
	s1 := e.id(`SELECT id FROM users WHERE username='siswa001'`) // 7A (peserta)
	s6 := e.id(`SELECT id FROM users WHERE username='siswa006'`) // 7B (bukan peserta)
	p := fmt.Sprintf("/api/cbt/%d/ljk", exam)

	code, b := o.do("GET", p, nil)
	expect(t, "guru lain membuka LJK", code, 403, b)
	code, b = o.do("POST", p, map[string]any{"siswa_id": s1, "jawaban": []string{"c"}})
	expect(t, "guru lain menyimpan LJK", code, 403, b)
	var info struct {
		Ujian   map[string]any   `json:"ujian"`
		Peserta []map[string]any `json:"peserta"`
	}
	g.json("GET", p, nil, &info)
	if len(info.Peserta) != 5 || info.Ujian["jumlah_soal"].(float64) != 4 {
		t.Errorf("info LJK salah: %d peserta, %v soal", len(info.Peserta), info.Ujian["jumlah_soal"])
	}
	code, b = g.do("POST", p, map[string]any{"siswa_id": s6, "jawaban": []string{"c"}})
	expect(t, "bukan peserta", code, 400, b)
	code, b = g.do("POST", p, map[string]any{"siswa_id": s1, "jawaban": []string{"c", "e"}}) // ujian 4 pilihan
	expect(t, "huruf di luar pilihan", code, 400, b)
	code, b = g.do("POST", p, map[string]any{"siswa_id": s1, "jawaban": []string{"c", "a", "a", "b", "a"}})
	expect(t, "jawaban melebihi jumlah soal", code, 400, b)

	var r map[string]any
	if code := g.json("POST", p, map[string]any{"siswa_id": s1, "jawaban": []string{"c", "a", "", "b"}}, &r); code != 200 {
		t.Fatalf("simpan LJK: %d", code)
	}
	if r["skor"].(float64) != 75 || r["benar"].(float64) != 3 {
		t.Errorf("nilai LJK salah: %v", r)
	}
	var metode string
	e.app.db.QueryRow(`SELECT metode FROM exam_attempts WHERE exam_id=? AND siswa_id=?`, exam, s1).Scan(&metode)
	if metode != "ljk" {
		t.Errorf("metode %q, harusnya ljk", metode)
	}
	// sudah bernilai: perlu konfirmasi timpa
	code, b = g.do("POST", p, map[string]any{"siswa_id": s1, "jawaban": []string{"c", "a", "c", "b"}})
	expect(t, "menimpa tanpa konfirmasi", code, 409, b)
	g.json("POST", p, map[string]any{"siswa_id": s1, "jawaban": []string{"c", "a", "c", "b"}, "timpa": true}, &r)
	if r["skor"].(float64) != 100 {
		t.Errorf("nilai setelah ditimpa %v, harusnya 100", r["skor"])
	}
	// hasil muncul di daftar nilai CBT
	var hasil []map[string]any
	g.json("GET", fmt.Sprintf("/api/cbt/%d/hasil", exam), nil, &hasil)
	found := false
	for _, h := range hasil {
		if h["siswa_id"] != nil && int64(h["siswa_id"].(float64)) == s1 {
			found = h["skor"] != nil && h["skor"].(float64) == 100
		}
	}
	if !found {
		t.Error("nilai LJK tidak tampil di hasil CBT")
	}
}

func TestKodeMapelMengikutiNama(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	k7a := e.id(`SELECT id FROM classes WHERE nama='7A'`)
	k7b := e.id(`SELECT id FROM classes WHERE nama='7B'`)
	kode := func(nama string, kelas int64) string {
		var k string
		e.app.db.QueryRow(`SELECT kode FROM courses WHERE nama=? AND kelas_id=?`, nama, kelas).Scan(&k)
		return k
	}
	// nama standar + kode keliru -> kode dibetulkan server
	code, b := a.do("POST", "/api/courses", map[string]any{"kode": "PAK", "nama": "Pendidikan Agama Buddha dan Budi Pekerti", "kelas_id": k7a})
	expect(t, "agama Buddha dengan kode PAK", code, 200, b)
	if k := kode("Pendidikan Agama Buddha dan Budi Pekerti", k7a); k != "PAB" {
		t.Errorf("kode tersimpan %q, harusnya PAB", k)
	}
	// nama yang sudah dipakai sekolah -> kode mengikuti yang sudah ada, walau dikirim lain
	code, b = a.do("POST", "/api/courses", map[string]any{"kode": "XYZ", "nama": "ilmu pengetahuan alam", "kelas_id": k7b})
	expect(t, "IPA di kelas lain", code, 200, b)
	if k := kode("ilmu pengetahuan alam", k7b); k != "IPA" {
		t.Errorf("kode IPA kelas lain %q, harusnya IPA", k)
	}
	// mapel baru meminjam kode mapel lain -> ditolak
	code, b = a.do("POST", "/api/courses", map[string]any{"kode": "mtk", "nama": "Muatan Lokal Batik", "kelas_id": k7b})
	expect(t, "mapel baru memakai kode MTK", code, 400, b)
	code, b = a.do("POST", "/api/courses", map[string]any{"kode": "BIG", "nama": "Muatan Lokal Batik", "kelas_id": k7b})
	expect(t, "mapel baru memakai kode standar milik mapel lain", code, 400, b)
	code, b = a.do("POST", "/api/courses", map[string]any{"kode": "MULOK", "nama": "Muatan Lokal Batik", "kelas_id": k7b})
	expect(t, "mapel baru dengan kode sendiri", code, 200, b)
	// mengubah nama mapel: kode ikut berubah
	ipa := e.id(`SELECT id FROM courses WHERE kelas_id=? AND kode='IPA'`, k7b)
	code, b = a.do("PUT", fmt.Sprintf("/api/courses/%d", ipa), map[string]any{"kode": "IPA", "nama": "Bahasa Inggris", "kelas_id": k7b})
	expect(t, "ganti nama mapel", code, 200, b)
	if k := kode("Bahasa Inggris", k7b); k != "BIG" {
		t.Errorf("kode setelah ganti nama %q, harusnya BIG", k)
	}
	var pil []map[string]string
	a.json("GET", "/api/mapel/pilihan", nil, &pil)
	seen := map[string]int{}
	for _, p := range pil {
		seen[strings.ToLower(p["nama"])]++
	}
	for n, c := range seen {
		if c > 1 {
			t.Errorf("pilihan mapel ganda: %s ×%d", n, c)
		}
	}
}

func TestPenjagaKeutuhanData(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	g := e.login("guru001", "guru123")
	req := func(want int, what string, c *client, method, path string, body any) {
		t.Helper()
		code, b := c.do(method, path, body)
		expect(t, what, code, want, b)
	}
	val := func(q string, args ...any) string { v, _ := (&Ctx{A: e.app}).exists(q, args...); return v }
	s1 := e.id(`SELECT id FROM users WHERE username='siswa001'`)
	s2 := e.id(`SELECT id FROM users WHERE username='siswa002'`)
	g1 := e.id(`SELECT id FROM users WHERE username='guru001'`)
	g2 := e.id(`SELECT id FROM users WHERE username='guru002'`)
	k7a := e.id(`SELECT id FROM classes WHERE nama='7A'`)
	k7b := e.id(`SELECT id FROM classes WHERE nama='7B'`)

	// NISN / NIK
	req(400, "NISN bukan 10 digit", a, "PUT", fmt.Sprintf("/api/bukuinduk/%d", s1), map[string]any{"nisn": "123"})
	req(200, "NISN dengan spasi (dirapikan)", a, "PUT", fmt.Sprintf("/api/bukuinduk/%d", s1), map[string]any{"nisn": "0011 2233 44", "nik": "3201010101010001"})
	if v := val(`SELECT nisn FROM student_profiles WHERE siswa_id=?`, s1); v != "0011223344" {
		t.Errorf("NISN tersimpan %q", v)
	}
	req(400, "NISN dipakai siswa lain", a, "PUT", fmt.Sprintf("/api/bukuinduk/%d", s2), map[string]any{"nisn": "0011223344"})
	req(400, "NIK dipakai siswa lain", a, "PUT", fmt.Sprintf("/api/bukuinduk/%d", s2), map[string]any{"nik": "3201010101010001"})

	// kelas: tingkat sesuai nama; satu guru satu kelas wali
	req(400, "8A bertingkat 7", a, "POST", "/api/classes", map[string]any{"nama": "8A", "tingkat": "7"})
	req(200, "XI IPA 1 tingkat 11", a, "POST", "/api/classes", map[string]any{"nama": "XI IPA 1", "tingkat": "11"})
	req(200, "9B tanpa tingkat", a, "POST", "/api/classes", map[string]any{"nama": "9B"})
	if v := val(`SELECT tingkat FROM classes WHERE nama='9B'`); v != "9" {
		t.Errorf("tingkat 9B terisi %q, harusnya 9", v)
	}
	req(200, "wali 7A", a, "PUT", fmt.Sprintf("/api/classes/%d", k7a), map[string]any{"nama": "7A", "tingkat": "7", "wali_id": g1})
	req(400, "guru yang sama wali 7B", a, "PUT", fmt.Sprintf("/api/classes/%d", k7b), map[string]any{"nama": "7B", "tingkat": "7", "wali_id": g1})
	req(200, "simpan ulang 7A (wali sama)", a, "PUT", fmt.Sprintf("/api/classes/%d", k7a), map[string]any{"nama": "7A", "tingkat": "7", "wali_id": g1})

	// tahun ajaran
	req(400, "tahun tidak berurutan", a, "POST", "/api/tahun", map[string]any{"nama": "2027/2029"})
	req(400, "tanggal selesai sebelum mulai", a, "POST", "/api/tahun", map[string]any{"nama": "2027/2028", "mulai": "2027-07-15", "selesai": "2027-01-01"})
	req(200, "tahun ajaran sah", a, "POST", "/api/tahun", map[string]any{"nama": "2027/2028", "mulai": "2027-07-15", "selesai": "2028-06-30"})

	// e-voting: nomor urut & nama kandidat
	el := e.id(`SELECT id FROM elections LIMIT 1`)
	req(400, "nomor urut ganda", a, "POST", fmt.Sprintf("/api/evote/%d/kandidat", el), map[string]any{"no": 1, "nama": "Calon Baru"})
	req(400, "nama kandidat ganda", a, "POST", fmt.Sprintf("/api/evote/%d/kandidat", el), map[string]any{"no": 3, "nama": "citra maharani"})
	req(200, "kandidat nomor 3", a, "POST", fmt.Sprintf("/api/evote/%d/kandidat", el), map[string]any{"no": 3, "nama": "Calon Baru"})

	// kasir: nama barang ganda; ejaan kategori disamakan
	req(400, "nama barang ganda", a, "POST", "/api/produk", map[string]any{"kode": "999", "nama": "pensil 2b", "harga": 1000})
	req(200, "barang baru kategori huruf kecil", a, "POST", "/api/produk", map[string]any{"kode": "998", "nama": "Penghapus", "kategori": "atk", "harga": 1000})
	if v := val(`SELECT kategori FROM products WHERE kode='998'`); v != "ATK" {
		t.Errorf("kategori tersimpan %q, harusnya ATK", v)
	}

	// buku: ISBN
	req(200, "buku ber-ISBN", a, "POST", "/api/buku", map[string]any{"kode": "BK-1", "judul": "Laskar Pelangi", "isbn": "978-979-3062-79-1", "stok": 2})
	req(400, "ISBN sama, judul lain", a, "POST", "/api/buku", map[string]any{"kode": "BK-2", "judul": "Laskar Pelangi (2)", "isbn": "9789793062791", "stok": 1})
	req(400, "ISBN salah panjang", a, "POST", "/api/buku", map[string]any{"kode": "BK-3", "judul": "X", "isbn": "123", "stok": 1})

	// NIP guru
	req(400, "NIP dipakai guru lain", a, "PUT", fmt.Sprintf("/api/users/%d", g2), map[string]any{"username": "guru002", "nama_lengkap": "Budi", "role": "guru", "nip": "19870101 201001 2 001"})
	req(400, "NIP salah panjang", a, "PUT", fmt.Sprintf("/api/users/%d", g2), map[string]any{"username": "guru002", "nama_lengkap": "Budi", "role": "guru", "nip": "12345"})
	req(200, "NIP unik", a, "PUT", fmt.Sprintf("/api/users/%d", g2), map[string]any{"username": "guru002", "nama_lengkap": "Budi", "role": "guru", "nip": "199001012015011002"})

	// surat keluar
	req(200, "surat keluar 1", a, "POST", "/api/surat", map[string]any{"jenis": "keluar", "nomor": "421/001/SMP/2026", "tanggal": "2026-11-09", "pihak": "Dinas", "perihal": "Laporan"})
	req(400, "nomor surat keluar ganda", a, "POST", "/api/surat", map[string]any{"jenis": "keluar", "nomor": "421/001/SMP/2026", "tanggal": "2026-11-10", "pihak": "Komite", "perihal": "Undangan"})
	req(200, "surat masuk bernomor sama (dari instansi lain) tetap boleh", a, "POST", "/api/surat", map[string]any{"jenis": "masuk", "nomor": "421/001/SMP/2026", "tanggal": "2026-11-10", "pihak": "SMP Lain", "perihal": "Undangan"})

	// kegiatan ujian, ruang, jadwal
	req(200, "kegiatan ujian", a, "POST", "/api/ujian-event", map[string]any{"nama": "PAS Ganjil", "mulai": "2026-12-01", "selesai": "2026-12-05"})
	ev := e.id(`SELECT id FROM exam_events WHERE nama='PAS Ganjil'`)
	req(400, "kegiatan selesai sebelum mulai", a, "POST", "/api/ujian-event", map[string]any{"nama": "PAS Genap", "mulai": "2027-06-10", "selesai": "2027-06-01"})
	req(200, "ruang 01", a, "POST", fmt.Sprintf("/api/ujian-event/%d/ruang", ev), map[string]any{"nama": "Ruang 01", "kapasitas": 20})
	req(400, "ruang 01 ganda", a, "POST", fmt.Sprintf("/api/ujian-event/%d/ruang", ev), map[string]any{"nama": "ruang 01", "kapasitas": 20})
	req(400, "jam selesai sebelum mulai", a, "POST", fmt.Sprintf("/api/ujian-event/%d/jadwal", ev), map[string]any{"tanggal": "2026-12-02", "jam_mulai": "10:00", "jam_selesai": "08:00", "mapel": "Matematika"})
	req(400, "jadwal di luar rentang kegiatan", a, "POST", fmt.Sprintf("/api/ujian-event/%d/jadwal", ev), map[string]any{"tanggal": "2026-12-20", "jam_mulai": "07:30", "jam_selesai": "09:30", "mapel": "Matematika"})
	req(200, "jadwal sah", a, "POST", fmt.Sprintf("/api/ujian-event/%d/jadwal", ev), map[string]any{"tanggal": "2026-12-02", "jam_mulai": "07:30", "jam_selesai": "09:30", "mapel": "Matematika", "kelas": "Kelas 7"})
	req(400, "jadwal ganda", a, "POST", fmt.Sprintf("/api/ujian-event/%d/jadwal", ev), map[string]any{"tanggal": "2026-12-02", "jam_mulai": "10:00", "jam_selesai": "11:00", "mapel": "matematika", "kelas": "Kelas 7"})

	// jurnal mengajar ganda
	mtk := e.id(`SELECT id FROM courses WHERE guru_id=? LIMIT 1`, g1)
	req(200, "jurnal 1", g, "POST", "/api/jurnal", map[string]any{"course_id": mtk, "tanggal": "2026-11-09", "jam_ke": "1-2", "materi": "A"})
	req(400, "jurnal ganda", g, "POST", "/api/jurnal", map[string]any{"course_id": mtk, "tanggal": "2026-11-09", "jam_ke": "1-2", "materi": "B"})
	req(200, "jurnal jam lain", g, "POST", "/api/jurnal", map[string]any{"course_id": mtk, "tanggal": "2026-11-09", "jam_ke": "3-4", "materi": "C"})
}

func TestPPDBTidakGanda(t *testing.T) {
	e := newEnv(t)
	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	kvSet(e.app.db, "cfg_ppdb", fmt.Sprintf(`{"buka":%q,"tutup":"","kuota":0,"jalur":"Zonasi","tahun_ajaran":"","info":""}`, past))
	daftar := func(nisn string) (int, []byte) {
		e.app.limit = newLoginLimiter() // batas formulir publik tidak ikut diuji di sini
		return e.anon().do("POST", "/api/ppdb/daftar", map[string]any{"data": map[string]string{"nama": "Calon", "tgl_lahir": "2014-05-05", "hp": "081299998888", "jalur": "Zonasi", "jk": "L", "nisn": nisn}})
	}
	code, b := daftar("12345")
	expect(t, "NISN pendek", code, 400, b)
	code, b = daftar("0099887766")
	expect(t, "daftar pertama", code, 200, b)
	code, b = daftar("0099 8877 66")
	expect(t, "NISN yang sama mendaftar lagi", code, 400, b)
}

func TestTingkatDariNama(t *testing.T) {
	for nama, want := range map[string]int{"7A": 7, "8 B": 8, "XI IPA 1": 11, "XII-2": 12, "X": 10, "IX A": 9, "VII": 7, "VIII C": 8,
		"10 TKJ 1": 10, "Kelas Unggulan": 0, "IPA X": 0, "1A": 1} {
		if got := tingkatDariNama(nama); got != want {
			t.Errorf("tingkatDariNama(%q) = %d, harusnya %d", nama, got, want)
		}
	}
}

func TestHitungSiswaSetaraDanPakaiIndeks(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	// buat arsip: naik kelas 7A→8A, 7B→8B, lalu satu siswa keluar & satu dipindah manual
	ty := e.id(`SELECT id FROM academic_years WHERE aktif=1`)
	k7a, k7b := e.id(`SELECT id FROM classes WHERE nama='7A'`), e.id(`SELECT id FROM classes WHERE nama='7B'`)
	code, b := a.do("POST", fmt.Sprintf("/api/tahun/%d/naik", ty), map[string]any{"tujuan": map[string]any{"nama": "2027/2028"},
		"kelas": []map[string]any{{"kelas_id": k7a, "aksi": "naik", "ke_nama": "8A"}, {"kelas_id": k7b, "aksi": "naik", "ke_nama": "8B"}}})
	expect(t, "kenaikan", code, 200, b)
	e.app.db.Exec(`UPDATE users SET status='nonaktif' WHERE username='siswa002'`)
	rows, _ := e.app.db.Query(`SELECT id FROM classes`)
	var ids []int64
	for rows.Next() {
		var id int64
		rows.Scan(&id)
		ids = append(ids, id)
	}
	rows.Close()
	for _, id := range ids {
		var lama, baru int
		e.app.db.QueryRow(`SELECT COUNT(*) FROM users u WHERE u.role='siswa' AND `+memberOf("?"), id, id).Scan(&lama)
		e.app.db.QueryRow(`SELECT ` + strings.ReplaceAll(memberCount("?"), "?", fmt.Sprint(id))).Scan(&baru)
		if lama != baru {
			t.Errorf("kelas %d: memberOf=%d, memberCount=%d", id, lama, baru)
		}
	}
	plan := func(q string, args ...any) string {
		rows, err := e.app.db.Query(`EXPLAIN QUERY PLAN `+q, args...)
		if err != nil {
			t.Fatalf("plan %q: %v", q, err)
		}
		defer rows.Close()
		var out []string
		for rows.Next() {
			var id, parent, notused int
			var detail string
			rows.Scan(&id, &parent, &notused, &detail)
			out = append(out, detail)
		}
		return strings.Join(out, " | ")
	}
	for name, q := range map[string][]any{
		"nomor kasir":          {`SELECT MAX(no) FROM sales WHERE no >= ? AND no < ?`, "TRX20261109", "TRX20261109~"},
		"jurnal bulanan admin": {`SELECT * FROM teaching_journal WHERE tanggal >= ? AND tanggal <= ?`, "2026-11-01", "2026-11-30"},
		"kehadiran jurnal":     {`SELECT COUNT(*) FROM attendance WHERE course_id=? AND tanggal=? AND status='hadir'`, 1, "2026-11-09"},
		"progres siswa":        {`SELECT material_id FROM material_progress WHERE siswa_id=?`, 1},
		"rekap wali":           {`SELECT siswa_id FROM daily_attendance WHERE tanggal >= ? AND tanggal <= ?`, "2026-11-01", "2026-11-30"},
	} {
		if p := plan(q[0].(string), q[1:]...); !strings.Contains(p, "USING") || strings.Contains(p, "SCAN "+strings.Fields(strings.SplitN(q[0].(string), "FROM ", 2)[1])[0]+" ") && !strings.Contains(p, "INDEX") {
			t.Errorf("%s tidak memakai indeks: %s", name, p)
		}
	}
}

func TestPembersihanFileYatimAman(t *testing.T) {
	e := newEnv(t)
	dir := e.app.uploadDir()
	os.MkdirAll(dir, 0o700)
	old := time.Now().Add(-48 * time.Hour)
	mk := func(name string) string {
		p := filepath.Join(dir, name)
		os.WriteFile(p, []byte("x"), 0o600)
		os.Chtimes(p, old, old)
		return p
	}
	dipakai := mk("aaaaaaaaaaaaaaaaaaaaaaaa.jpg")
	yatim := mk("bbbbbbbbbbbbbbbbbbbbbbbb.jpg")
	baru := filepath.Join(dir, "cccccccccccccccccccccccc.jpg")
	os.WriteFile(baru, []byte("x"), 0o600) // < 24 jam: belum boleh disentuh
	s1 := e.id(`SELECT id FROM users WHERE username='siswa001'`)
	e.app.db.Exec(`INSERT INTO student_profiles (siswa_id, foto) VALUES (?, '/files/aaaaaaaaaaaaaaaaaaaaaaaa.jpg') ON CONFLICT (siswa_id) DO UPDATE SET foto=excluded.foto`, s1)

	// query rujukan gagal (kolom hilang) -> tidak ada file yang dipindah
	e.app.db.Exec(`ALTER TABLE letters RENAME COLUMN file TO file_lama`)
	if n, err := e.app.removeOrphans(24 * time.Hour); err == nil || n != 0 {
		t.Fatalf("pembersihan harus dibatalkan saat query gagal: n=%d err=%v", n, err)
	}
	for _, p := range []string{dipakai, yatim, baru} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("file %s tersentuh padahal pembersihan dibatalkan", filepath.Base(p))
		}
	}
	e.app.db.Exec(`ALTER TABLE letters RENAME COLUMN file_lama TO file`)

	n, err := e.app.removeOrphans(24 * time.Hour)
	if err != nil || n != 1 {
		t.Fatalf("pembersihan: n=%d err=%v", n, err)
	}
	if _, err := os.Stat(dipakai); err != nil {
		t.Error("file yang masih dipakai ikut dipindah")
	}
	if _, err := os.Stat(baru); err != nil {
		t.Error("file baru (<24 jam) ikut dipindah")
	}
	if _, err := os.Stat(filepath.Join(dir, ".sampah", filepath.Base(yatim))); err != nil {
		t.Error("file yatim tidak ada di folder sampah (harus dipindah, bukan dihapus)")
	}
	if e.app.emptyTrash(30*24*time.Hour) != 0 {
		t.Error("file sampah yang baru dipindah ikut dihapus")
	}
}

func TestKalenderOtomatisDanLiburNasional(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	s1 := e.login("siswa001", "siswa123") // 7A
	s6 := e.login("siswa006", "siswa123") // 7B
	g2 := e.login("guru002", "guru123")
	ambil := func(c *client) map[string]map[string]any {
		var rows []map[string]any
		c.json("GET", "/api/kalender?dari=2000-01-01&sampai=2100-12-31", nil, &rows)
		out := map[string]map[string]any{}
		for _, r := range rows {
			if id, ok := r["id"].(string); ok {
				out[id] = r
			}
		}
		return out
	}
	exam := e.id(`SELECT id FROM exams LIMIT 1`) // ujian demo untuk kelas 7A
	cbt := fmt.Sprintf("auto-cbt-%d", exam)
	if _, ok := ambil(s1)[cbt]; !ok {
		t.Error("siswa 7A tidak melihat ujian CBT kelasnya di kalender")
	}
	if _, ok := ambil(s6)[cbt]; ok {
		t.Error("siswa 7B melihat ujian CBT kelas lain")
	}
	if _, ok := ambil(a)[cbt]; !ok {
		t.Error("admin tidak melihat ujian CBT")
	}
	// tenggat tugas: hanya siswa kelasnya & guru pengampu
	mtk := e.id(`SELECT id FROM courses WHERE kode='MTK' AND kelas_id=(SELECT id FROM classes WHERE nama='7A')`)
	e.app.db.Exec(`INSERT INTO assignments (course_id, judul, deadline, nilai_maks, created_at) VALUES (?, 'PR Pecahan', ?, 100, ?)`, mtk, time.Now().Add(48*time.Hour).UTC().Format(time.RFC3339), nowStr())
	tid := fmt.Sprintf("auto-tugas-%d", e.id(`SELECT id FROM assignments WHERE judul='PR Pecahan'`))
	if r, ok := ambil(s1)[tid]; !ok || r["kategori"] != "tugas" {
		t.Errorf("siswa tidak melihat tenggat tugas: %v", r)
	}
	if _, ok := ambil(s6)[tid]; ok {
		t.Error("siswa kelas lain melihat tenggat tugas")
	}
	if _, ok := ambil(g2)[tid]; ok {
		t.Error("guru lain melihat tenggat tugas mapel yang bukan ampuannya")
	}
	// e-voting: pemilih siswa → siswa melihat, guru biasa tidak
	el := e.id(`SELECT id FROM elections LIMIT 1`)
	vid := fmt.Sprintf("auto-vote-%d", el)
	if _, ok := ambil(s1)[vid]; !ok {
		t.Error("siswa pemilih tidak melihat masa e-voting")
	}
	if _, ok := ambil(g2)[vid]; ok {
		t.Error("guru yang bukan pemilih melihat masa e-voting")
	}
	// tanpa kegiatan otomatis
	var plain []map[string]any
	s1.json("GET", "/api/kalender?otomatis=0", nil, &plain)
	for _, r := range plain {
		if r["otomatis"] == true {
			t.Fatal("otomatis=0 masih memuat kegiatan otomatis")
		}
	}

	// libur nasional tanggal tetap untuk tahun ajaran aktif
	var ta string
	e.app.db.QueryRow(`SELECT nama FROM academic_years WHERE aktif=1`).Scan(&ta)
	var r map[string]any
	if code := a.json("POST", "/api/kalender/libur-nasional", map[string]string{"tahun_ajaran": ta}, &r); code != 200 || r["ditambahkan"].(float64) != 5 {
		t.Fatalf("libur nasional: %d %v", code, r)
	}
	y1, _ := strconv.Atoi(ta[:4])
	for _, d := range []string{fmt.Sprintf("%d-08-17", y1), fmt.Sprintf("%d-12-25", y1), fmt.Sprintf("%d-01-01", y1+1), fmt.Sprintf("%d-05-01", y1+1), fmt.Sprintf("%d-06-01", y1+1)} {
		if e.id(`SELECT COUNT(*) FROM events WHERE mulai=? AND kategori='libur'`, d) != 1 {
			t.Errorf("libur %s tidak tercatat", d)
		}
	}
	a.json("POST", "/api/kalender/libur-nasional", map[string]string{"tahun_ajaran": ta}, &r)
	if r["ditambahkan"].(float64) != 0 {
		t.Errorf("libur nasional tergandakan: %v", r)
	}
	code, b := g2.do("POST", "/api/kalender/libur-nasional", map[string]string{"tahun_ajaran": ta})
	expect(t, "guru biasa menambah libur nasional", code, 403, b)
}
