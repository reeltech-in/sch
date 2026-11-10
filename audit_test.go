package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
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
