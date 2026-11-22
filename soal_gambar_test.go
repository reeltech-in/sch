package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Soal bergambar (mis. hasil Asisten AI) bisa masuk ke kuis mapel, sama seperti CBT, dan gambar terkirim ke siswa.
func TestKuisSoalBergambar(t *testing.T) {
	e := newEnv(t)
	g := e.login("guru001", "guru123")
	course := e.id(`SELECT id FROM courses WHERE guru_id=(SELECT id FROM users WHERE username='guru001') LIMIT 1`)
	if code, b := g.do("POST", "/api/quizzes", map[string]any{"course_id": course, "judul": "Kuis Gambar", "durasi_menit": 10, "jumlah_opsi": 4}); code != 200 {
		t.Fatalf("buat kuis: %d %s", code, b)
	}
	quiz := e.id(`SELECT id FROM quizzes WHERE judul='Kuis Gambar'`)
	gb := "/files/0123456789abcdef01234567.png"
	soal := map[string]any{"pertanyaan": "Perhatikan gambar berikut! Nilai sin A adalah ....", "gambar": gb,
		"opsi_a": "3/5", "opsi_b": "4/5", "opsi_c": "3/4", "opsi_d": "4/3", "jawaban_benar": "a"}
	if code, b := g.do("POST", fmt.Sprintf("/api/quizzes/%d/questions/batch", quiz), map[string]any{"soal": []any{soal}}); code != 200 {
		t.Fatalf("impor: %d %s", code, b)
	}
	// alamat gambar di luar folder unggahan ditolak
	jahat := map[string]any{"pertanyaan": "x", "gambar": "/files/../sch.db", "opsi_a": "a", "opsi_b": "b", "opsi_c": "c", "opsi_d": "d", "jawaban_benar": "a"}
	if code, _ := g.do("POST", fmt.Sprintf("/api/quizzes/%d/questions/batch", quiz), map[string]any{"soal": []any{jahat}}); code != 400 {
		t.Errorf("gambar tidak valid diterima: %d", code)
	}
	if code, _ := g.do("POST", fmt.Sprintf("/api/quizzes/%d/questions", quiz), jahat); code != 400 {
		t.Errorf("gambar tidak valid diterima (satu soal): %d", code)
	}
	// ubah satu soal: gambar boleh dihapus
	qid := e.id(`SELECT id FROM quiz_questions WHERE quiz_id=?`, quiz)
	soal["gambar"] = ""
	if code, b := g.do("PUT", fmt.Sprintf("/api/questions/%d", qid), soal); code != 200 {
		t.Fatalf("ubah: %d %s", code, b)
	}
	if n := e.id(`SELECT COUNT(*) FROM quiz_questions WHERE id=? AND gambar IS NULL`, qid); n != 1 {
		t.Error("gambar tidak terhapus")
	}
	soal["gambar"] = gb
	g.do("PUT", fmt.Sprintf("/api/questions/%d", qid), soal)

	s := e.login("siswa001", "siswa123")
	var mulai struct {
		Questions []struct {
			Gambar string `json:"gambar"`
		} `json:"questions"`
	}
	if code := s.json("POST", fmt.Sprintf("/api/quizzes/%d/start", quiz), nil, &mulai); code != 200 || len(mulai.Questions) != 1 || mulai.Questions[0].Gambar != gb {
		t.Fatalf("siswa tidak menerima gambar soal: %d %+v", code, mulai.Questions)
	}
}

// Word: baris ![..](/files/x.png) disisipkan sebagai gambar; blok ```svg yang belum diproses tidak tercetak sebagai kode.
func TestDocxGambarSoal(t *testing.T) {
	dir := t.TempDir()
	name := "abcdef0123456789abcdef01.png"
	f, _ := os.Create(filepath.Join(dir, name))
	png.Encode(f, image.NewRGBA(image.Rect(0, 0, 600, 300)))
	f.Close()
	md := "1. Perhatikan gambar berikut!\n![Gambar soal](/files/" + name + ")\nA. 3/5\nB. 4/5\n\n2. Soal kedua\n```svg\n<svg xmlns=\"http://www.w3.org/2000/svg\"><text>rahasia</text></svg>\n```\nA. x\n" +
		"3. Gambar hilang\n![Gambar soal](/files/ffffffffffffffffffffffff.png)\nA. y"
	b, err := buildDocx(md, docxOpts{Judul: "Soal", ImgDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	zr, _ := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	got := map[string]string{}
	for _, zf := range zr.File {
		rc, _ := zf.Open()
		data, _ := io.ReadAll(rc)
		rc.Close()
		got[zf.Name] = string(data)
	}
	if _, ok := got["word/media/ridimg1.png"]; !ok {
		t.Error("gambar soal tidak ikut di file Word")
	}
	if len(zr.File) != 7 {
		t.Errorf("jumlah bagian %d, harusnya 7 (gambar yang tidak ada di server dilewati)", len(zr.File))
	}
	doc := got["word/document.xml"]
	// 600 px dibatasi 400 px → 400×200 px
	for _, w := range []string{`r:embed="rIdImg1"`, fmt.Sprintf(`cx="%d" cy="%d"`, 400*emuPerPx, 200*emuPerPx), "[gambar]"} {
		if !strings.Contains(doc, w) {
			t.Errorf("document.xml tidak memuat %q", w)
		}
	}
	if strings.Contains(doc, "rahasia") || strings.Contains(doc, "/files/") {
		t.Error("kode SVG / alamat file tercetak sebagai teks")
	}
	if !strings.Contains(got["word/_rels/document.xml.rels"], `Target="media/ridimg1.png"`) {
		t.Error("relasi gambar tidak ada")
	}
}

// Gambar soal kuis dan gambar di dokumen soal Asisten AI (belum diimpor) tidak dianggap file yatim.
func TestGambarSoalBukanFileYatim(t *testing.T) {
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
	kuis, ai, yatim := mk("dddddddddddddddddddddddd.png"), mk("eeeeeeeeeeeeeeeeeeeeeeee.png"), mk("ffffffffffffffffffffffff.png")
	quiz := e.id(`SELECT id FROM quizzes LIMIT 1`)
	if _, err := e.app.db.Exec(`INSERT INTO quiz_questions (quiz_id, pertanyaan, gambar, opsi_a, opsi_b, jawaban_benar, created_at) VALUES (?, 'x', '/files/dddddddddddddddddddddddd.png', 'a', 'b', 'a', ?)`, quiz, nowStr()); err != nil {
		t.Fatal(err)
	}
	guru := e.id(`SELECT id FROM users WHERE username='guru001'`)
	if _, err := e.app.db.Exec(`INSERT INTO ai_documents (user_id, template, judul, input, output, model, created_at) VALUES (?, 'soal', 'Soal', '{}', ?, 'm', ?)`,
		guru, "1. Perhatikan gambar berikut!\n![Gambar soal](/files/eeeeeeeeeeeeeeeeeeeeeeee.png)\nA. a", nowStr()); err != nil {
		t.Fatal(err)
	}
	if n, err := e.app.removeOrphans(24 * time.Hour); err != nil || n != 1 {
		t.Fatalf("pembersihan: n=%d err=%v", n, err)
	}
	for _, p := range []string{kuis, ai} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("gambar soal %s ikut dipindah ke sampah", filepath.Base(p))
		}
	}
	if _, err := os.Stat(yatim); err == nil {
		t.Error("file yatim tidak dipindah")
	}
}
