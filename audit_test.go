package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Tes regresi hasil audit v1.20.0.

// kunci jawaban asli semua soal ujian: {"<id soal>": "<huruf>"}
func kunciUjian(t *testing.T, e *env, exam int64) map[string]string {
	rows, err := e.app.db.Query(`SELECT id, jawaban_benar FROM exam_questions WHERE exam_id=?`, exam)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var id int64
		var k string
		rows.Scan(&id, &k)
		out[fmt.Sprint(id)] = k
	}
	return out
}

// Siswa sudah menjawab (tersimpan otomatis) lalu HP mati dan tidak pernah menekan Kumpulkan:
// setelah waktu habis jawaban tetap dinilai.
func TestCBTDinilaiOtomatisSetelahWaktuHabis(t *testing.T) {
	e := newEnv(t)
	s := e.login("siswa001", "siswa123")
	exam := e.id(`SELECT id FROM exams LIMIT 1`)
	if code, b := s.do("POST", fmt.Sprintf("/api/cbt/%d/start", exam), map[string]string{"token": "MTK26"}); code != 200 {
		t.Fatalf("mulai: %d %s", code, b)
	}
	jb, _ := json.Marshal(kunciUjian(t, e, exam))
	old := time.Now().Add(-3 * time.Hour).UTC().Format(time.RFC3339)
	e.app.db.Exec(`UPDATE exam_attempts SET jawaban=?, started_at=? WHERE exam_id=?`, string(jb), old, exam)

	e.app.finalizeCBT(0)
	var selesai int
	var skor float64
	e.app.db.QueryRow(`SELECT selesai, COALESCE(skor,-1) FROM exam_attempts WHERE exam_id=?`, exam).Scan(&selesai, &skor)
	if selesai != 1 || skor != 100 {
		t.Errorf("ujian tidak dinilai otomatis: selesai=%d skor=%v (harusnya 1 & 100)", selesai, skor)
	}
}

// Soal yang ditambahkan guru setelah siswa mulai tidak ikut menurunkan nilai siswa itu.
func TestCBTSoalTambahanTidakMenurunkanNilai(t *testing.T) {
	e := newEnv(t)
	s := e.login("siswa001", "siswa123")
	exam := e.id(`SELECT id FROM exams LIMIT 1`)
	if code, b := s.do("POST", fmt.Sprintf("/api/cbt/%d/start", exam), map[string]string{"token": "MTK26"}); code != 200 {
		t.Fatalf("mulai: %d %s", code, b)
	}
	jb, _ := json.Marshal(kunciUjian(t, e, exam))
	old := time.Now().Add(-3 * time.Hour).UTC().Format(time.RFC3339)
	e.app.db.Exec(`UPDATE exam_attempts SET jawaban=?, started_at=? WHERE exam_id=?`, string(jb), old, exam)
	e.app.db.Exec(`INSERT INTO exam_questions (exam_id, pertanyaan, opsi_a, opsi_b, jawaban_benar, bobot, created_at) VALUES (?,?,?,?,?,?,?)`,
		exam, "Soal susulan", "x", "y", "a", 5, nowStr())
	code, b := s.do("POST", fmt.Sprintf("/api/cbt/%d/submit", exam), map[string]any{"answers": map[string]string{}})
	expect(t, "kumpul", code, 200, b)
	var skor float64
	e.app.db.QueryRow(`SELECT skor FROM exam_attempts WHERE exam_id=?`, exam).Scan(&skor)
	if skor != 100 {
		t.Errorf("soal yang tidak diterima siswa ikut dihitung: skor=%v, harusnya 100", skor)
	}
}

// Jumlah pilihan tidak boleh dikurangi bila masih ada soal berkunci di luar pilihan baru.
func TestJumlahPilihanTidakMemutusKunci(t *testing.T) {
	e := newEnv(t)
	exam := e.id(`SELECT id FROM exams LIMIT 1`)
	e.app.db.Exec(`UPDATE exam_questions SET jawaban_benar='e', opsi_e='lima' WHERE id=(SELECT MIN(id) FROM exam_questions WHERE exam_id=?)`, exam)
	c := &Ctx{A: e.app}
	if err := opsiCukup(c, "exam_questions", "exam_id", exam, 4); err == nil {
		t.Error("pengurangan ke 4 pilihan diterima padahal ada kunci E")
	}
	if err := opsiCukup(c, "exam_questions", "exam_id", exam, 5); err != nil {
		t.Errorf("5 pilihan harus diterima: %v", err)
	}
}

// Menghapus status absensi hanya berlaku untuk siswa kelas yang dikirim, bukan siswa kelas lain.
func TestHapusAbsensiHanyaKelasSendiri(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	lain := e.id(`SELECT id FROM users WHERE username='siswa006'`) // kelas 7B
	kelas7A := e.id(`SELECT kelas_id FROM users WHERE username='siswa001'`)
	tgl := time.Now().Format("2006-01-02")
	e.app.db.Exec(`INSERT INTO daily_attendance (siswa_id, tanggal, status, metode) VALUES (?,?,'sakit','manual')`, lain, tgl)
	code, b := a.do("PUT", "/api/absen/harian", map[string]any{"kelas_id": kelas7A, "tanggal": tgl,
		"items": []map[string]any{{"siswa_id": lain, "status": ""}}})
	expect(t, "simpan absensi", code, 200, b)
	var n int
	e.app.db.QueryRow(`SELECT COUNT(*) FROM daily_attendance WHERE siswa_id=? AND tanggal=?`, lain, tgl).Scan(&n)
	if n != 1 {
		t.Error("absensi siswa kelas lain terhapus lewat kelas yang berbeda")
	}
}

// Absen selfie: bacaan lokasi tidak akurat ditolak; titik persis di titik sekolah ditandai untuk diperiksa.
func TestSelfieLokasiDiperiksa(t *testing.T) {
	e := newEnv(t)
	kvSet(e.app.db, "cfg_absen", `{"selfie":true,"lat":-6.2,"lng":106.8,"radius_m":150,"jam_masuk":"07:00"}`)
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for i := range img.Pix {
		img.Pix[i] = byte(i * 7)
	}
	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	foto := "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())

	s2 := e.login("siswa002", "siswa123")
	code, b := s2.do("POST", "/api/absen/selfie", map[string]any{"foto": foto, "lat": -6.2005, "lng": 106.8, "akurasi": 5000})
	expect(t, "lokasi tidak akurat", code, 400, b)

	s1 := e.login("siswa001", "siswa123")
	code, b = s1.do("POST", "/api/absen/selfie", map[string]any{"foto": foto, "lat": -6.2, "lng": 106.8, "akurasi": 8})
	expect(t, "selfie di titik sekolah", code, 200, b)
	var ket string
	e.app.db.QueryRow(`SELECT COALESCE(keterangan,'') FROM daily_attendance WHERE siswa_id=(SELECT id FROM users WHERE username='siswa001')`).Scan(&ket)
	if !strings.Contains(ket, "persis sama") {
		t.Errorf("selfie mencurigakan tidak ditandai: keterangan=%q", ket)
	}
}

// HP offline saat waktu habis → ujian dinilai otomatis; kiriman ulang dari HP setelahnya tetap mendapat nilai
// (bukan galat "sudah dikirim" yang membuat siswa terjebak di layar ujian).
func TestKirimUlangSetelahDinilaiOtomatis(t *testing.T) {
	e := newEnv(t)
	s := e.login("siswa001", "siswa123")
	exam := e.id(`SELECT id FROM exams LIMIT 1`)
	if code, b := s.do("POST", fmt.Sprintf("/api/cbt/%d/start", exam), map[string]string{"token": "MTK26"}); code != 200 {
		t.Fatalf("mulai: %d %s", code, b)
	}
	jb, _ := json.Marshal(kunciUjian(t, e, exam))
	old := time.Now().Add(-3 * time.Hour).UTC().Format(time.RFC3339)
	e.app.db.Exec(`UPDATE exam_attempts SET jawaban=?, started_at=? WHERE exam_id=?`, string(jb), old, exam)
	e.app.db.Exec(`UPDATE exams SET tampil_nilai=1 WHERE id=?`, exam)
	e.app.finalizeCBT(0)
	var out struct {
		Skor *float64 `json:"skor"`
	}
	if code := s.json("POST", fmt.Sprintf("/api/cbt/%d/submit", exam), map[string]any{"answers": map[string]string{}}, &out); code != 200 {
		t.Fatalf("kirim ulang setelah dinilai otomatis: status %d, harusnya 200", code)
	}
	if out.Skor == nil || *out.Skor != 100 {
		t.Errorf("kirim ulang harus menampilkan nilai tersimpan 100, dapat %v", out.Skor)
	}
}

// Kenaikan kelas lengkap: naik + tinggal + keluar + kelas lulus + salin mapel; tidak bisa dijalankan dua kali.
func TestKenaikanKelasSemuaJalur(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	ty := e.id(`SELECT id FROM academic_years WHERE aktif=1`)
	k7a := e.id(`SELECT id FROM classes WHERE nama='7A'`)
	k7b := e.id(`SELECT id FROM classes WHERE nama='7B'`)
	s2 := e.id(`SELECT id FROM users WHERE username='siswa002'`)
	s3 := e.id(`SELECT id FROM users WHERE username='siswa003'`)
	mapel7a := e.id(`SELECT COUNT(*) FROM courses WHERE kelas_id=?`, k7a)
	body := map[string]any{
		"tujuan":      map[string]any{"nama": "2027/2028"},
		"kelas":       []map[string]any{{"kelas_id": k7a, "aksi": "naik", "ke_nama": "8A"}, {"kelas_id": k7b, "aksi": "lulus"}},
		"kecuali":     []map[string]any{{"siswa_id": s2, "aksi": "tinggal"}, {"siswa_id": s3, "aksi": "keluar"}},
		"salin_mapel": true,
	}
	code, b := a.do("POST", fmt.Sprintf("/api/tahun/%d/naik", ty), body)
	expect(t, "kenaikan kelas", code, 200, b)
	var r map[string]float64
	json.Unmarshal(b, &r)
	if r["naik"] != 3 || r["tinggal"] != 1 || r["keluar"] != 1 || r["lulus"] != 1 {
		t.Errorf("ringkasan salah: %v (harusnya naik 3, tinggal 1, keluar 1, lulus 1)", r)
	}
	kelasDari := func(user string) string {
		var k string
		e.app.db.QueryRow(`SELECT COALESCE(c.nama,'') FROM users u LEFT JOIN classes c ON c.id=u.kelas_id WHERE u.username=?`, user).Scan(&k)
		return k
	}
	dst := e.id(`SELECT id FROM academic_years WHERE nama='2027/2028'`)
	if k := kelasDari("siswa002"); k != "7A" || e.id(`SELECT tahun_id FROM classes c JOIN users u ON u.kelas_id=c.id WHERE u.username='siswa002'`) != dst {
		t.Errorf("siswa tinggal kelas: kelas %q, harusnya 7A tahun baru", k)
	}
	var st, lulusan string
	e.app.db.QueryRow(`SELECT status FROM users WHERE username='siswa003'`).Scan(&st)
	if st != "nonaktif" || kelasDari("siswa003") != "" {
		t.Errorf("siswa keluar masih aktif/berkelas: %s", st)
	}
	e.app.db.QueryRow(`SELECT status || '|' || COALESCE(angkatan_lulus,'') FROM users WHERE username='siswa006'`).Scan(&lulusan)
	if !strings.HasPrefix(lulusan, "nonaktif|") || strings.HasSuffix(lulusan, "|") {
		t.Errorf("siswa kelas lulus: %q, harusnya nonaktif dengan angkatan lulus", lulusan)
	}
	if n := e.id(`SELECT COUNT(*) FROM courses c JOIN classes k ON k.id=c.kelas_id WHERE k.nama='8A'`); n != mapel7a {
		t.Errorf("mapel tersalin ke 8A = %d, harusnya %d", n, mapel7a)
	}
	if n := e.id(`SELECT COUNT(*) FROM sessions WHERE user_id=?`, s3); n != 0 {
		t.Error("sesi siswa yang keluar tidak dihapus")
	}
	// tahun baru sekarang aktif; tahun lama tidak bisa dinaikkan lagi
	code, _ = a.do("POST", fmt.Sprintf("/api/tahun/%d/naik", ty), body)
	if code < 400 {
		t.Error("kenaikan kelas tahun yang sama bisa dijalankan dua kali")
	}
}

// Membuka ujian lagi (HP mati/pindah perangkat): soal & acakan sama, jawaban tersimpan kembali tampil.
func TestCBTLanjutkanMenampilkanJawaban(t *testing.T) {
	e := newEnv(t)
	s := e.login("siswa001", "siswa123")
	exam := e.id(`SELECT id FROM exams LIMIT 1`)
	type startRes struct {
		StartedAt string `json:"started_at"`
		Questions []struct {
			ID int64 `json:"id"`
		} `json:"questions"`
		Answers map[string]string `json:"answers"`
	}
	var a, b startRes
	if code := s.json("POST", fmt.Sprintf("/api/cbt/%d/start", exam), map[string]string{"token": "MTK26"}, &a); code != 200 || len(a.Questions) == 0 {
		t.Fatalf("mulai: %d", code)
	}
	q1 := fmt.Sprint(a.Questions[0].ID)
	if code, body := s.do("POST", fmt.Sprintf("/api/cbt/%d/save", exam), map[string]any{"answers": map[string]string{q1: "c"}}); code != 200 {
		t.Fatalf("simpan: %d %s", code, body)
	}
	if code := s.json("POST", fmt.Sprintf("/api/cbt/%d/start", exam), map[string]string{}, &b); code != 200 {
		t.Fatalf("lanjutkan: %d", code)
	}
	if b.StartedAt != a.StartedAt || len(b.Questions) != len(a.Questions) || b.Questions[0].ID != a.Questions[0].ID {
		t.Error("melanjutkan ujian mengubah waktu mulai atau urutan soal")
	}
	if b.Answers[q1] != "c" {
		t.Errorf("jawaban tersimpan tidak tampil lagi: %q, harusnya c", b.Answers[q1])
	}
}

// Asisten AI dengan penyedia palsu: jawaban streaming tersimpan sebagai dokumen, galat penyedia tampil sebagai
// pesan, dan kuota harian berlaku (dihitung dari pemakaian, bukan jumlah dokumen).
func TestAsistenAIStreamingDanKuota(t *testing.T) {
	e := newEnv(t)
	mode := "ok"
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		if mode == "galat" {
			io.WriteString(w, "data: {\"error\":{\"message\":\"model tidak tersedia\"}}\n\n")
			return
		}
		for _, p := range []string{"# LKPD", " Pecahan", "\\n1. Soal satu"} {
			fmt.Fprintf(w, "data: {\"choices\":[{\"delta\":{\"content\":\"%s\"}}]}\n\n", p)
		}
		io.WriteString(w, "data: [DONE]\n\n")
	}))
	defer fake.Close()
	kvSet(e.app.db, "cfg_ai", `{"base_url":"`+fake.URL+`","model":"palsu","kuota_harian":1,"suhu":0.7}`)
	g := e.login("guru001", "guru123")
	gen := func() (int, string) {
		code, b := g.do("POST", "/api/ai/generate", map[string]any{"template": "bebas", "inputs": map[string]string{"pertanyaan": "buat LKPD pecahan"}})
		return code, string(b)
	}
	code, out := gen()
	if code != 200 || !strings.Contains(out, "# LKPD Pecahan") || !strings.Contains(out, "\x00DOC:") {
		t.Fatalf("jawaban streaming: %d %q", code, out)
	}
	if n := e.id(`SELECT COUNT(*) FROM ai_documents`); n != 1 {
		t.Errorf("dokumen tersimpan = %d, harusnya 1", n)
	}
	// kuota 1 sudah terpakai; menghapus dokumen tidak mengembalikan kuota
	e.app.db.Exec(`DELETE FROM ai_documents`)
	if code, out = gen(); code != 400 || !strings.Contains(out, "Kuota") {
		t.Errorf("kuota harian tidak berlaku setelah dokumen dihapus: %d %s", code, out)
	}
	// galat dari penyedia tampil sebagai pesan yang bisa dibaca
	kvSet(e.app.db, "cfg_ai", `{"base_url":"`+fake.URL+`","model":"palsu","kuota_harian":0,"suhu":0.7}`)
	mode = "galat"
	if code, out = gen(); code != 200 || !strings.Contains(out, "[Gagal: model tidak tersedia]") {
		t.Errorf("galat penyedia: %d %q", code, out)
	}
}

// Kuis mapel: jawaban tersimpan ke server selama dikerjakan; kiriman terlambat menilai jawaban tersimpan (bukan 0);
// kuis yang tidak dikirim dinilai otomatis; kirim ulang setelah selesai mengembalikan nilai.
func TestKuisSimpanOtomatisDanTerlambat(t *testing.T) {
	e := newEnv(t)
	quiz := e.id(`SELECT id FROM quizzes LIMIT 1`)
	kunci := map[string]string{}
	rows, _ := e.app.db.Query(`SELECT id, jawaban_benar FROM quiz_questions WHERE quiz_id=?`, quiz)
	for rows.Next() {
		var id int64
		var k string
		rows.Scan(&id, &k)
		kunci[fmt.Sprint(id)] = k
	}
	rows.Close()
	var start struct {
		Answers map[string]string `json:"answers"`
	}
	run := func(user string) *client {
		s := e.login(user, "siswa123")
		if code := s.json("POST", fmt.Sprintf("/api/quizzes/%d/start", quiz), nil, &start); code != 200 {
			t.Fatalf("%s mulai kuis: %d", user, code)
		}
		if code, b := s.do("POST", fmt.Sprintf("/api/quizzes/%d/save", quiz), map[string]any{"answers": kunci}); code != 200 {
			t.Fatalf("%s simpan: %d %s", user, code, b)
		}
		return s
	}
	mundur := func(user string) {
		old := time.Now().Add(-3 * time.Hour).UTC().Format(time.RFC3339)
		e.app.db.Exec(`UPDATE quiz_attempts SET started_at=? WHERE quiz_id=? AND siswa_id=(SELECT id FROM users WHERE username=?)`, old, quiz, user)
	}
	skor := func(user string) (int, float64) {
		var sel int
		var sk float64
		e.app.db.QueryRow(`SELECT selesai, skor FROM quiz_attempts WHERE quiz_id=? AND siswa_id=(SELECT id FROM users WHERE username=?)`, quiz, user).Scan(&sel, &sk)
		return sel, sk
	}

	// 1. melanjutkan: jawaban tersimpan kembali dari server
	s1 := run("siswa001")
	s1.json("POST", fmt.Sprintf("/api/quizzes/%d/start", quiz), nil, &start)
	if len(start.Answers) != len(kunci) {
		t.Errorf("jawaban tersimpan tidak kembali saat melanjutkan: %v", start.Answers)
	}
	// 2. kirim terlambat (jawaban kosong dari HP) → yang dinilai jawaban tersimpan
	mundur("siswa001")
	var out struct {
		Skor float64 `json:"skor"`
	}
	if code := s1.json("POST", fmt.Sprintf("/api/quizzes/%d/submit", quiz), map[string]any{"answers": map[string]string{}}, &out); code != 200 || out.Skor != 100 {
		t.Errorf("kiriman terlambat: status %d skor %v, harusnya 200 & 100", code, out.Skor)
	}
	// 3. kirim ulang setelah selesai → nilai tersimpan, bukan galat
	if code := s1.json("POST", fmt.Sprintf("/api/quizzes/%d/submit", quiz), map[string]any{"answers": map[string]string{}}, &out); code != 200 || out.Skor != 100 {
		t.Errorf("kirim ulang: status %d skor %v", code, out.Skor)
	}
	// 4. tidak pernah dikirim → dinilai otomatis setelah waktu habis; simpan setelah waktu habis ditolak
	s2 := run("siswa003")
	mundur("siswa003")
	if code, _ := s2.do("POST", fmt.Sprintf("/api/quizzes/%d/save", quiz), map[string]any{"answers": kunci}); code != 400 {
		t.Errorf("simpan setelah waktu habis: %d, harusnya 400", code)
	}
	e.app.finalizeQuiz()
	if sel, sk := skor("siswa003"); sel != 1 || sk != 100 {
		t.Errorf("kuis tidak dinilai otomatis: selesai=%d skor=%v", sel, sk)
	}
}

// Pengaturan: hanya kolom yang dikirim yang berubah; balasan berisi pengaturan terbaru.
func TestPengaturanSimpanSebagian(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	var before map[string]any
	a.json("GET", "/api/settings", nil, &before)
	var after map[string]any
	if code := a.json("PUT", "/api/settings", map[string]any{"telepon": "0211234"}, &after); code != 200 {
		t.Fatalf("simpan sebagian: %d", code)
	}
	if after["telepon"] != "0211234" || after["nama_sekolah"] != before["nama_sekolah"] || after["kepala_sekolah"] != before["kepala_sekolah"] {
		t.Errorf("kolom lain ikut berubah: sebelum %v, sesudah %v", before, after)
	}
	if code, _ := a.do("PUT", "/api/settings", map[string]any{"nama_sekolah": "  "}); code != 400 {
		t.Error("nama sekolah kosong diterima")
	}
	var ak map[string]any
	a.json("PUT", "/api/settings", map[string]any{"akses_guru": map[string]bool{"tugas": false}}, &ak)
	if m, _ := ak["akses_guru"].(map[string]any); m == nil || m["tugas"] != false || m["materi"] != true {
		t.Errorf("hak akses guru: %v", ak["akses_guru"])
	}
}

// Log Aktivitas: pencarian teks tanpa tanggal dibatasi 90 hari (memakai indeks waktu), dengan tanggal mengikuti pilihan.
func TestLogAktivitasPencarianDibatasi(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	lama := time.Now().AddDate(0, 0, -200).UTC().Format(time.RFC3339)
	e.app.db.Exec(`INSERT INTO audit_log (waktu, aksi, status, nama, detail) VALUES (?,?,?,?,?)`, lama, "POST /api/uji", 200, "Penanda Lama", "catatan-lama-unik")
	code, b := a.raw("GET", "/api/audit?q=catatan-lama-unik", nil, true)
	if code != 200 || strings.Contains(string(b), "catatan-lama-unik") {
		t.Errorf("pencarian tanpa tanggal ikut memindai >90 hari: %d %s", code, b)
	}
	dari := time.Now().AddDate(0, 0, -365).Format("2006-01-02")
	code, b = a.raw("GET", "/api/audit?q=catatan-lama-unik&dari="+dari, nil, true)
	if code != 200 || !strings.Contains(string(b), "catatan-lama-unik") {
		t.Errorf("pencarian dengan tanggal dari: %d %s", code, b)
	}
}

// Halaman ujian lama masih terbuka setelah guru menghapus ujiannya: siswa mendapat alasan yang jelas.
func TestCBTDihapusPesanJelas(t *testing.T) {
	e := newEnv(t)
	s := e.login("siswa001", "siswa123")
	code, b := s.do("POST", "/api/cbt/999999/start", map[string]string{})
	if code != 404 || !strings.Contains(string(b), "sudah dihapus guru") {
		t.Fatalf("ujian dihapus: %d %s", code, b)
	}
}
