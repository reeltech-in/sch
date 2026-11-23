package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// Uji regresi untuk temuan audit keamanan run-1 (lihat CHANGELOG v1.28.1).

// Berkas portofolio tidak boleh ikut dibersihkan sebagai file yatim, dan berkas lama dibuang saat diganti.
func TestPortofolioBerkasAmanDanTidakMenumpuk(t *testing.T) {
	e := newEnv(t)
	s1 := e.login("siswa001", "siswa123")
	if code, b := s1.do("POST", "/api/portofolio", map[string]any{"judul": "Poster", "course_id": 1, "file": pngDataURL(t)}); code != 200 {
		t.Fatalf("simpan portofolio: %d %s", code, b)
	}
	var id int64
	var url string
	e.app.db.QueryRow(`SELECT id, file_url FROM portfolios ORDER BY id DESC LIMIT 1`).Scan(&id, &url)
	path := filepath.Join(e.app.uploadDir(), strings.TrimPrefix(url, "/files/"))
	if _, err := e.app.removeOrphans(0); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("berkas portofolio ikut dibersihkan sebagai file yatim: %v", err)
	}
	// ganti berkas: berkas lama langsung hilang, yang baru ada
	if code, b := s1.do("PUT", fmt.Sprintf("/api/portofolio/%d", id), map[string]any{"judul": "Poster", "course_id": 1, "file": pngDataURL(t)}); code != 200 {
		t.Fatalf("ubah portofolio: %d %s", code, b)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("berkas lama tidak dibuang saat diganti (err=%v)", err)
	}
	var baru string
	e.app.db.QueryRow(`SELECT file_url FROM portfolios WHERE id=?`, id).Scan(&baru)
	if _, err := os.Stat(filepath.Join(e.app.uploadDir(), strings.TrimPrefix(baru, "/files/"))); err != nil {
		t.Errorf("berkas baru tidak ada: %v", err)
	}
}

func bukaPPDB(e *env) {
	past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)
	kvSet(e.app.db, "cfg_ppdb", fmt.Sprintf(`{"buka":%q,"tutup":"","kuota":0,"jalur":"Zonasi","tahun_ajaran":"","info":""}`, past))
}

// Pendaftaran PPDB publik tidak boleh dipakai mengirim teks bebas lewat WA sekolah, dan dibatasi per nomor.
func TestPPDBWhatsAppTidakBisaDisalahgunakan(t *testing.T) {
	e := newEnv(t)
	bukaPPDB(e)
	a := e.login("admin", testAdminPW)
	if code, b := a.do("PUT", "/api/wa/config", map[string]any{"aktif": true, "provider": "fonnte", "token": "T", "kirim_ppdb": true}); code != 200 {
		t.Fatalf("simpan WA: %d %s", code, b)
	}
	daftar := func(nama, hp string) (int, []byte) {
		e.app.limit = newLoginLimiter()
		return e.anon().do("POST", "/api/ppdb/daftar", map[string]any{"data": map[string]string{"nama": nama, "tgl_lahir": "2014-05-05", "hp": hp, "jk": "L"}})
	}
	code, b := daftar("Budi\n\nPENTING: transfer biaya ke rek 0012345 http://contoh.invalid/bayar", "081200000001")
	expect(t, "daftar", code, 200, b)
	var pesan string
	e.app.db.QueryRow(`SELECT pesan FROM wa_outbox WHERE jenis='ppdb' ORDER BY id DESC LIMIT 1`).Scan(&pesan)
	if !strings.Contains(pesan, "atas nama Budi P. T.") {
		t.Errorf("nama di pesan WA tidak disingkat: %q", pesan)
	}
	for _, bad := range []string{"\n", "http", "://", "0012345", "/bayar", "transfer", "PENTING"} {
		if strings.Contains(pesan, bad) {
			t.Errorf("isian nama tembus ke pesan WA (%q): %q", bad, pesan)
		}
	}
	// nomor yang sama: maksimal waPPDBPerNomor pesan per hari
	for i := 0; i < 3; i++ {
		daftar(fmt.Sprintf("Calon %c", 'A'+i), "081200000001")
	}
	if n := e.id(`SELECT COUNT(*) FROM wa_outbox WHERE jenis='ppdb' AND tujuan='6281200000001'`); n != waPPDBPerNomor {
		t.Errorf("pesan PPDB ke satu nomor: %d, harusnya %d", n, waPPDBPerNomor)
	}
	// pesan kehadiran didahulukan dari antrean PPDB
	s1 := e.id(`SELECT id FROM users WHERE username='siswa001'`)
	e.app.db.Exec(`INSERT INTO student_profiles (siswa_id, hp_ortu) VALUES (?, '081233334444') ON CONFLICT (siswa_id) DO UPDATE SET hp_ortu=excluded.hp_ortu`, s1)
	a.do("POST", "/api/absen/scan", map[string]string{"token": e.app.cardToken(s1), "mode": "masuk"})
	gw := newFakeWA(t)
	old := waFonnteURL
	waFonnteURL = gw.srv.URL
	t.Cleanup(func() { waFonnteURL = old })
	e.app.waSendNext()
	if len(gw.got) != 1 || gw.got[0]["target"] != "6281233334444" {
		t.Errorf("pesan kehadiran tidak dikirim lebih dulu: %v", gw.got)
	}
}

// Pesan "sudah terdaftar" tidak menyebut nomor pendaftaran orang lain.
func TestPPDBDuplikatTidakMembocorkanNomor(t *testing.T) {
	e := newEnv(t)
	bukaPPDB(e)
	daftar := func() (int, []byte) {
		e.app.limit = newLoginLimiter()
		return e.anon().do("POST", "/api/ppdb/daftar", map[string]any{"data": map[string]string{"nama": "Calon", "tgl_lahir": "2014-05-05", "hp": "081299998888", "nik": "3201010101010001"}})
	}
	code, b := daftar()
	expect(t, "daftar pertama", code, 200, b)
	code, b = daftar()
	expect(t, "NIK yang sama", code, 400, b)
	if strings.Contains(string(b), "PPDB-") {
		t.Errorf("nomor pendaftaran bocor ke pengirim anonim: %s", b)
	}
}

// Permintaan serentak tidak boleh menembus batas 10 tebakan salah per perangkat.
func TestCekKelulusanSerentakTetapDibatasi(t *testing.T) {
	e := newEnv(t)
	b, _ := json.Marshal(map[string]any{"buka": time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)})
	kvSet(e.app.db, "cfg_kelulusan", string(b))
	var wg sync.WaitGroup
	var mu sync.Mutex
	dinilai := 0
	start := make(chan struct{})
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			body := strings.NewReader(fmt.Sprintf(`{"nisn":"0012345678","tgl_lahir":"2010-01-%02d"}`, i%28+1))
			req, _ := http.NewRequest("POST", e.srv.URL+"/api/kelulusan/cek", body)
			req.Header.Set("X-SCH", "1")
			req.Header.Set("Content-Type", "application/json")
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return
			}
			res.Body.Close()
			if res.StatusCode != 429 {
				mu.Lock()
				dinilai++
				mu.Unlock()
			}
		}(i)
	}
	close(start)
	wg.Wait()
	if dinilai > guessIPWindow {
		t.Errorf("%d tebakan serentak diperiksa, batasnya %d", dinilai, guessIPWindow)
	}
}

// Isian yang benar mengembalikan jatah (siswa asli tidak kehilangan jatah karena cek yang berhasil).
func TestCekKelulusanBenarTidakMemakanJatah(t *testing.T) {
	e := newEnv(t)
	b, _ := json.Marshal(map[string]any{"buka": time.Now().Add(-time.Hour).UTC().Format(time.RFC3339)})
	kvSet(e.app.db, "cfg_kelulusan", string(b))
	s1 := e.id(`SELECT id FROM users WHERE username='siswa001'`)
	e.app.db.Exec(`INSERT INTO student_profiles (siswa_id, nisn, tgl_lahir) VALUES (?, '0055555555', '2011-02-03') ON CONFLICT (siswa_id) DO UPDATE SET nisn=excluded.nisn, tgl_lahir=excluded.tgl_lahir`, s1)
	e.app.db.Exec(`INSERT INTO graduation (siswa_id, status, updated_at) VALUES (?, 'lulus', ?) ON CONFLICT (siswa_id) DO UPDATE SET status='lulus'`, s1, nowStr())
	c := e.anon()
	for i := 0; i < 15; i++ {
		if code, b := c.do("POST", "/api/kelulusan/cek", map[string]string{"nisn": "0055555555", "tgl_lahir": "2011-02-03"}); code != 200 {
			t.Fatalf("cek benar ke-%d: %d %s", i+1, code, b)
		}
	}
}

// Slot baca kiriman publik: per perangkat dibatasi (kiriman besar yang sedang dibaca belum terhitung batas 10 menit).
func TestKirimanPublikSerentakDibatasi(t *testing.T) {
	e := newEnv(t)
	ctx := func() *Ctx {
		r := httptest.NewRequest("POST", "/api/bukutamu", nil)
		r.RemoteAddr = "203.0.113.7:5555"
		return &Ctx{R: r, A: e.app}
	}
	var lepas []func()
	for i := 0; i < publicInflightPerIP; i++ {
		done, err := ctx().publicBlocked("tamu:", 6)
		if err != nil {
			t.Fatalf("slot %d ditolak: %v", i+1, err)
		}
		lepas = append(lepas, done)
	}
	if _, err := ctx().publicBlocked("tamu:", 6); err == nil {
		t.Error("kiriman serentak melebihi batas per perangkat tetap diterima")
	}
	for _, f := range lepas {
		f()
	}
	if done, err := ctx().publicBlocked("tamu:", 6); err != nil {
		t.Errorf("slot tidak dikembalikan: %v", err)
	} else {
		done()
	}
}

// Satu perangkat tidak bisa menghabiskan jatah unggahan publik harian milik semua pendaftar.
func TestJatahUnggahPublikPerPerangkat(t *testing.T) {
	e := newEnv(t)
	png := []byte("\x89PNG\r\n\x1a\n" + strings.Repeat("x", 2000))
	publikUpl.Lock()
	publikUpl.hari, publikUpl.bytes, publikUpl.per = time.Now().Format("2006-01-02"), 0, map[string]int64{"198.51.100.1": publikPerPerangkat - 100}
	publikUpl.Unlock()
	if _, err := e.app.savePublicUpload(png, true, "198.51.100.1"); err == nil {
		t.Error("perangkat yang jatahnya habis masih bisa mengunggah")
	}
	if _, err := e.app.savePublicUpload(png, true, "198.51.100.2"); err != nil {
		t.Errorf("perangkat lain ikut terblokir: %v", err)
	}
}

// Di mode nginx (-proxy-header xff) header CF-Connecting-IP kiriman pengunjung diabaikan; nilai non-IP diabaikan.
func TestIPAsliDiBelakangProxy(t *testing.T) {
	oldT, oldP, oldH := trustProxy, trustedProxies, proxyHeader
	t.Cleanup(func() { trustProxy, trustedProxies, proxyHeader = oldT, oldP, oldH })
	trustProxy, trustedProxies = true, nil
	req := func(cf, xff string) *http.Request {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = "127.0.0.1:9999"
		if cf != "" {
			r.Header.Set("CF-Connecting-IP", cf)
		}
		if xff != "" {
			r.Header.Set("X-Forwarded-For", xff)
		}
		return r
	}
	proxyHeader = "cf"
	if got := clientIP(req("198.51.100.9", "")); got != "198.51.100.9" {
		t.Errorf("mode cf: %s", got)
	}
	if got := clientIP(req("bukan-ip-acak-123", "")); got != "127.0.0.1" {
		t.Errorf("nilai CF-Connecting-IP bukan IP ikut dipakai: %s", got)
	}
	proxyHeader = "xff"
	if got := clientIP(req("198.51.100.9", "1.2.3.4, 203.0.113.5")); got != "203.0.113.5" {
		t.Errorf("mode xff harus memakai elemen terakhir X-Forwarded-For dan mengabaikan CF-Connecting-IP: %s", got)
	}
}

// E-voting: setelah ada suara, pemilihan tidak bisa ditutup-buka untuk mengintip hasil, dan identitas kandidat dikunci.
func TestEvotingTerkunciSetelahAdaSuara(t *testing.T) {
	e := newEnv(t)
	el := e.id(`SELECT id FROM elections LIMIT 1`)
	cand := e.id(`SELECT id FROM candidates WHERE election_id=? ORDER BY no LIMIT 1`, el)
	s := e.login("siswa001", "siswa123")
	if code, b := s.do("POST", fmt.Sprintf("/api/evote/%d/pilih", el), map[string]any{"candidate_id": cand}); code != 200 {
		t.Fatalf("memilih: %d %s", code, b)
	}
	a := e.login("admin", testAdminPW)
	var judul, mulai, selesai string
	e.app.db.QueryRow(`SELECT judul, mulai, selesai FROM elections WHERE id=?`, el).Scan(&judul, &mulai, &selesai)
	simpan := func(m, sl string) (int, []byte) {
		return a.do("PUT", fmt.Sprintf("/api/evote/%d", el), map[string]any{"judul": judul, "mulai": m, "selesai": sl, "pemilih": "siswa", "kelas_ids": []int64{}})
	}
	tutup := time.Now().Add(-time.Second).UTC().Format(time.RFC3339)
	code, b := simpan(mulai, tutup)
	expect(t, "menutup lebih awal", code, 200, b)
	code, b = simpan(mulai, time.Now().Add(time.Hour).UTC().Format(time.RFC3339))
	expect(t, "membuka lagi pemilihan yang sudah selesai", code, 400, b)
	code, b = simpan(time.Now().Add(-2*time.Hour).UTC().Format(time.RFC3339), tutup)
	expect(t, "mengubah waktu mulai setelah ada suara", code, 400, b)
	// kandidat: nama/nomor terkunci, visi boleh diubah
	var no int
	var nama, wakil string
	e.app.db.QueryRow(`SELECT no, nama, COALESCE(wakil,'') FROM candidates WHERE id=?`, cand).Scan(&no, &nama, &wakil)
	code, b = a.do("PUT", fmt.Sprintf("/api/evote-kandidat/%d", cand), map[string]any{"no": no, "nama": "Nama Lain", "wakil": wakil})
	expect(t, "mengganti nama kandidat yang sudah punya suara", code, 400, b)
	code, b = a.do("PUT", fmt.Sprintf("/api/evote-kandidat/%d", cand), map[string]any{"no": no, "nama": nama, "wakil": wakil, "visi": "Visi baru"})
	expect(t, "memperbaiki visi", code, 200, b)
}

// Mematikan modul Pembelajaran harus menolak API LMS di server; beranda siswa tetap terbuka tanpa isi LMS.
func TestModulLMSDimatikanDitolakServer(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	if code, b := a.do("PUT", "/api/modules", map[string]any{"lms": false}); code != 200 {
		t.Fatalf("matikan lms: %d %s", code, b)
	}
	// cache kv berlaku untuk seluruh proses tes: nyalakan lagi agar tes lain tidak ikut terdampak
	t.Cleanup(func() { kvSet(e.app.db, "modul", "{}"); kvForget("modul") })
	quiz := e.id(`SELECT id FROM quizzes LIMIT 1`)
	s := e.login("siswa001", "siswa123")
	code, b := s.do("POST", fmt.Sprintf("/api/quizzes/%d/start", quiz), nil)
	expect(t, "mulai kuis saat modul mati", code, 403, b)
	g := e.login("guru001", "guru123")
	code, b = g.do("POST", "/api/materials", map[string]any{"course_id": 1, "judul": "x"})
	expect(t, "guru menambah materi saat modul mati", code, 403, b)
	var me map[string]any
	if code := s.json("GET", "/api/rekap/me", nil, &me); code != 200 {
		t.Fatalf("beranda siswa: %d", code)
	}
	if q, _ := me["quizzes"].([]any); len(q) != 0 {
		t.Errorf("isi LMS tetap muncul saat modul mati: %v", me["quizzes"])
	}
}

// Login serentak dengan password salah tidak boleh menembus batas 5 percobaan per akun per perangkat
// (sebelumnya semua permintaan lolos pemeriksaan selama bcrypt berjalan).
func TestLoginSerentakTetapDibatasi(t *testing.T) {
	e := newEnv(t)
	var wg sync.WaitGroup
	var mu sync.Mutex
	diperiksa := 0
	start := make(chan struct{})
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			body := strings.NewReader(fmt.Sprintf(`{"username":"guru001","password":"tebakan%d"}`, i))
			req, _ := http.NewRequest("POST", e.srv.URL+"/api/login", body)
			req.Header.Set("X-SCH", "1")
			req.Header.Set("Content-Type", "application/json")
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				return
			}
			res.Body.Close()
			if res.StatusCode != 429 {
				mu.Lock()
				diperiksa++
				mu.Unlock()
			}
		}(i)
	}
	close(start)
	wg.Wait()
	if diperiksa > 5 {
		t.Errorf("%d tebakan password serentak diperiksa, batasnya 5", diperiksa)
	}
}

// Login yang berhasil tidak memakan jatah: banyak siswa di balik satu IP sekolah (NAT) tetap bisa masuk,
// dan salah ketik sebelum berhasil tidak menumpuk.
func TestLoginBerhasilTidakMemakanJatah(t *testing.T) {
	e := newEnv(t)
	c := e.anon()
	for i := 0; i < 65; i++ {
		code, b := c.do("POST", "/api/login", map[string]string{"username": "siswa001", "password": "siswa123"})
		expect(t, fmt.Sprintf("login benar ke-%d", i+1), code, 200, b)
	}
	for ronde := 0; ronde < 3; ronde++ {
		for i := 0; i < 4; i++ {
			code, b := c.do("POST", "/api/login", map[string]string{"username": "guru001", "password": "salah"})
			expect(t, "salah ketik", code, 401, b)
		}
		code, b := c.do("POST", "/api/login", map[string]string{"username": "guru001", "password": "guru123"})
		expect(t, fmt.Sprintf("benar setelah salah ketik (ronde %d)", ronde+1), code, 200, b)
	}
}

// Aturan password per peran: admin 10, guru 8, siswa 6; tidak boleh memuat username atau password umum.
func TestAturanPasswordPerPeran(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	akun := func(role, uname, pw string) (int, []byte) {
		return a.do("POST", "/api/users", map[string]any{"nama_lengkap": "Uji " + uname, "username": uname, "password": pw, "role": role})
	}
	for _, k := range []struct {
		role, uname, pw string
		want            int
	}{
		{"guru", "guruuji1", "Kebun7mangga", 200},
		{"guru", "guruuji2", "abc1234", 400},     // 7 karakter: kurang untuk guru
		{"guru", "guruuji3", "12345678", 400},    // password umum
		{"guru", "guruuji4", "guruuji4abc", 400}, // memuat username
		{"admin", "adminuji1", "Kebun7mangga", 200},
		{"admin", "adminuji2", "Kebun7man", 400}, // 9 karakter: kurang untuk admin
		{"siswa", "siswauji1", "mawar7", 200},    // siswa tetap 6
		{"siswa", "siswauji2", "mawar", 400},     // 5 karakter
		{"guru", "guruuji5", "aaaaaaaaaa", 400},  // satu huruf berulang
	} {
		code, b := akun(k.role, k.uname, k.pw)
		expect(t, k.role+" "+k.pw, code, k.want, b)
	}
	// ganti password sendiri juga memakai aturan peran
	code, b := a.do("POST", "/api/password", map[string]string{"old": testAdminPW, "new": "pendek123"})
	expect(t, "admin ganti ke 9 karakter", code, 400, b)
	code, b = a.do("POST", "/api/password", map[string]string{"old": testAdminPW, "new": "Kebun7mangga-baru"})
	expect(t, "admin ganti ke password kuat", code, 200, b)
}

// Berkas pendaftar PPDB (KK, akta, foto) hanya bisa dibuka admin & guru yang diberi tugas PPDB, bukan siswa atau
// guru lain yang mengetahui nama filenya. Berkas pendaftar lama ikut terlindungi setelah pembaruan.
func TestBerkasPPDBHanyaUntukPengelola(t *testing.T) {
	e := newEnv(t)
	bukaPPDB(e)
	code, b := e.anon().do("POST", "/api/ppdb/daftar", map[string]any{
		"data":   map[string]string{"nama": "Calon Berkas", "tgl_lahir": "2014-05-05", "hp": "081277776666", "jk": "P"},
		"berkas": map[string]string{"kk": pngDataURL(t)}})
	expect(t, "daftar PPDB dengan berkas", code, 200, b)
	var berkas string
	e.app.db.QueryRow(`SELECT berkas FROM ppdb_applicants ORDER BY id DESC LIMIT 1`).Scan(&berkas)
	var m map[string]string
	json.Unmarshal([]byte(berkas), &m)
	url := m["kk"]
	if !strings.HasPrefix(url, "/files/") {
		t.Fatalf("berkas KK tidak tersimpan: %q", berkas)
	}
	expect(t, "admin membuka KK", func() int { c, _ := e.login("admin", testAdminPW).do("GET", url, nil); return c }(), 200, nil)
	expect(t, "siswa membuka KK", func() int { c, _ := e.login("siswa001", "siswa123").do("GET", url, nil); return c }(), 404, nil)
	expect(t, "guru tanpa tugas PPDB", func() int { c, _ := e.login("guru001", "guru123").do("GET", url, nil); return c }(), 404, nil)
	e.app.db.Exec(`UPDATE users SET akses='["ppdb"]' WHERE username='guru001'`)
	expect(t, "guru bertugas PPDB", func() int { c, _ := e.login("guru001", "guru123").do("GET", url, nil); return c }(), 200, nil)

	// pendaftar dari versi lama (belum tercatat di upload_akses) ikut terlindungi saat server dinyalakan ulang
	e.app.db.Exec(`DELETE FROM upload_akses`)
	expect(t, "sebelum migrasi", func() int { c, _ := e.login("siswa001", "siswa123").do("GET", url, nil); return c }(), 200, nil)
	if err := runModuleSchemas(e.app.db); err != nil {
		t.Fatal(err)
	}
	expect(t, "siswa setelah migrasi", func() int { c, _ := e.login("siswa001", "siswa123").do("GET", url, nil); return c }(), 404, nil)
}
