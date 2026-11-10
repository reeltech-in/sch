package main

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRentangSemester(t *testing.T) {
	kosong := sql.NullString{}
	if m, a := rentangSemester("2026/2027", kosong, kosong, kosong, 1); m != "2026-07-01" || a != "2026-12-31" {
		t.Fatalf("ganjil bawaan: %s–%s", m, a)
	}
	if m, a := rentangSemester("2026/2027", kosong, kosong, kosong, 2); m != "2027-01-01" || a != "2027-06-30" {
		t.Fatalf("genap bawaan: %s–%s", m, a)
	}
	ag := sql.NullString{String: "2026-12-20", Valid: true}
	mulai := sql.NullString{String: "2026-07-14", Valid: true}
	if m, a := rentangSemester("2026/2027", mulai, kosong, ag, 1); m != "2026-07-14" || a != "2026-12-20" {
		t.Fatalf("ganjil diatur: %s–%s", m, a)
	}
	if m, _ := rentangSemester("2026/2027", mulai, kosong, ag, 2); m != "2026-12-21" {
		t.Fatalf("genap mulai sehari setelah akhir ganjil, dapat %s", m)
	}
}

func TestDeskripsiKokurikuler(t *testing.T) {
	got := deskripsiKoku(map[string]string{"penalaran": "SB", "kolaborasi": "SB", "komunikasi": "MB"})
	want := "Ananda sangat berkembang dalam penalaran kritis dan kolaborasi; mulai berkembang dalam komunikasi."
	if got != want {
		t.Fatalf("dapat %q", got)
	}
}

// Alur lengkap: guru mapel mengisi nilai, wali mengisi kokurikuler & terbit, siswa melihat setelah terbit,
// isi terkunci setelah terbit, peran lain ditolak, Word bisa diunduh.
func TestRaporAlurLengkap(t *testing.T) {
	e := newEnv(t)
	g1 := e.login("guru001", "guru123") // Matematika 7A
	g2 := e.login("guru002", "guru123") // IPA 7A
	guru1 := e.id(`SELECT id FROM users WHERE username='guru001'`)
	e.app.db.Exec(`UPDATE classes SET wali_id=? WHERE id=1`, guru1) // guru001 wali 7A
	s1 := e.login("siswa001", "siswa123")
	q := "tahun=1&smt=1"

	var mp struct {
		Siswa []struct {
			ID       int64     `json:"id"`
			Otomatis skorMapel `json:"otomatis"`
		}
		Terbit bool
	}
	if code := g1.json("GET", "/api/rapor/mapel?course_id=1&"+q, nil, &mp); code != 200 || len(mp.Siswa) != 5 {
		t.Fatalf("rapor mapel: %d, %d siswa", code, len(mp.Siswa))
	}
	if code, _ := g2.do("GET", "/api/rapor/mapel?course_id=1&"+q, nil); code != 403 {
		t.Fatalf("guru lain tidak boleh membuka nilai rapor mapel orang lain, dapat %d", code)
	}
	isi := func(c *client, course int64, nilai float64) (int, []byte) {
		var items []map[string]any
		for _, s := range mp.Siswa {
			items = append(items, map[string]any{"siswa_id": s.ID, "nilai": nilai, "tinggi": "Menunjukkan penguasaan yang baik dalam bilangan bulat.", "perlu": "Perlu bimbingan dalam pecahan."})
		}
		return c.do("PUT", "/api/rapor/mapel", map[string]any{"course_id": course, "tahun": 1, "smt": 1, "items": items})
	}
	if code, _ := isi(g1, 1, 120); code != 400 {
		t.Fatalf("nilai di atas 100 harus ditolak")
	}
	if code, b := isi(g1, 1, 85); code != 200 {
		t.Fatalf("isi nilai mapel: %d %s", code, b)
	}
	// terbit ditolak selama mapel IPA belum diisi
	if code, b := g1.do("PUT", "/api/rapor/kelas/1/terbit", map[string]any{"tahun": 1, "smt": 1, "terbit": true}); code != 400 || !strings.Contains(string(b), "Ilmu Pengetahuan Alam") {
		t.Fatalf("terbit harus ditolak bila ada mapel belum lengkap: %d %s", code, b)
	}
	if code, b := isi(g2, 2, 78); code != 200 {
		t.Fatalf("isi nilai IPA: %d %s", code, b)
	}
	sid := e.id(`SELECT id FROM users WHERE username='siswa001'`)
	if code, b := g1.do("PUT", fmt.Sprintf("/api/rapor/siswa/%d", sid), map[string]any{"tahun": 1, "smt": 1,
		"kokurikuler": map[string]string{"penalaran": "SB", "kolaborasi": "BSH", "ngawur": "SB"}, "catatan": "Pertahankan semangat belajarmu.", "keputusan": "naik"}); code != 400 {
		t.Fatalf("keputusan naik kelas di semester ganjil harus ditolak: %d %s", code, b)
	}
	if code, b := g1.do("PUT", fmt.Sprintf("/api/rapor/siswa/%d", sid), map[string]any{"tahun": 1, "smt": 1,
		"kokurikuler": map[string]string{"penalaran": "SB", "kolaborasi": "BSH", "ngawur": "SB"}, "catatan": "Pertahankan semangat belajarmu.", "sakit": 2}); code != 200 {
		t.Fatalf("isi data wali: %d %s", code, b)
	}
	if code, _ := g2.do("PUT", fmt.Sprintf("/api/rapor/siswa/%d", sid), map[string]any{"tahun": 1, "smt": 1, "catatan": "x"}); code != 403 {
		t.Fatalf("bukan wali kelas harus ditolak")
	}
	// siswa belum bisa melihat sebelum terbit
	if code, _ := s1.do("GET", fmt.Sprintf("/api/rapor/siswa/%d?%s", sid, q), nil); code != 400 {
		t.Fatalf("rapor belum terbit tidak boleh terlihat siswa, dapat %d", code)
	}
	if code, b := g1.do("PUT", "/api/rapor/kelas/1/terbit", map[string]any{"tahun": 1, "smt": 1, "terbit": true}); code != 200 {
		t.Fatalf("terbit: %d %s", code, b)
	}
	var r raporLengkap
	if code := s1.json("GET", fmt.Sprintf("/api/rapor/siswa/%d?%s", sid, q), nil, &r); code != 200 {
		t.Fatalf("siswa melihat rapor: %d", code)
	}
	if len(r.Mapel) != 2 || r.Absen["sakit"] != 2 || r.Catatan == "" || !strings.Contains(r.KokuDesk, "sangat berkembang dalam penalaran kritis") || r.Terbit == "" {
		t.Fatalf("isi rapor tidak lengkap: %+v", r)
	}
	for _, k := range r.Koku {
		if k["dimensi"] == "ngawur" {
			t.Fatalf("dimensi tidak dikenal tidak boleh tersimpan")
		}
	}
	s2id := e.id(`SELECT id FROM users WHERE username='siswa002'`)
	if code, _ := s1.do("GET", fmt.Sprintf("/api/rapor/siswa/%d?%s", s2id, q), nil); code != 403 {
		t.Fatalf("siswa tidak boleh melihat rapor siswa lain")
	}
	// terkunci setelah terbit
	if code, _ := isi(g1, 1, 90); code != 400 {
		t.Fatalf("nilai tidak boleh berubah setelah terbit")
	}
	var me struct{ Terbit []map[string]any }
	if code := s1.json("GET", "/api/rapor/me", nil, &me); code != 200 || len(me.Terbit) != 1 {
		t.Fatalf("daftar rapor terbit siswa: %d %v", code, me.Terbit)
	}
	// Word
	code, b := s1.do("GET", fmt.Sprintf("/api/rapor/siswa/%d/docx?%s", sid, q), nil)
	if code != 200 {
		t.Fatalf("unduh Word: %d", code)
	}
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatalf("docx bukan zip: %v", err)
	}
	var doc string
	for _, f := range zr.File {
		if f.Name == "word/document.xml" {
			rc, _ := f.Open()
			var buf bytes.Buffer
			buf.ReadFrom(rc)
			rc.Close()
			doc = buf.String()
		}
	}
	for _, want := range []string{"LAPORAN HASIL BELAJAR", "Matematika", "Kokurikuler", "Ketidakhadiran", "Pertahankan semangat belajarmu."} {
		if !strings.Contains(doc, want) {
			t.Fatalf("Word rapor tidak memuat %q", want)
		}
	}
	// batal terbit membuka kunci
	if code, _ := g1.do("PUT", "/api/rapor/kelas/1/terbit", map[string]any{"tahun": 1, "smt": 1, "terbit": false}); code != 200 {
		t.Fatalf("batal terbit")
	}
	if code, _ := isi(g1, 1, 90); code != 200 {
		t.Fatalf("setelah batal terbit nilai bisa diubah lagi")
	}
}

// Ekstrakurikuler: admin membuat, pembina mengisi anggota & predikat, guru lain ditolak, masuk ke rapor.
func TestEkskulPembina(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	g1 := e.login("guru001", "guru123")
	g2 := e.login("guru002", "guru123")
	guru2 := e.id(`SELECT id FROM users WHERE username='guru002'`)
	if code, b := a.do("POST", "/api/ekskul", map[string]any{"nama": "Pramuka", "pembina_id": guru2}); code != 200 {
		t.Fatalf("buat ekskul: %d %s", code, b)
	}
	if code, _ := g1.do("POST", "/api/ekskul", map[string]any{"nama": "Futsal"}); code != 403 {
		t.Fatalf("guru biasa tidak boleh membuat ekskul")
	}
	ek := e.id(`SELECT id FROM ekskul WHERE nama='Pramuka'`)
	sid := e.id(`SELECT id FROM users WHERE username='siswa001'`)
	if code, _ := g1.do("PUT", fmt.Sprintf("/api/ekskul/%d/anggota", ek), map[string]any{"tahun": 1, "siswa_ids": []int64{sid}}); code != 403 {
		t.Fatalf("bukan pembina harus ditolak")
	}
	if code, b := g2.do("PUT", fmt.Sprintf("/api/ekskul/%d/anggota", ek), map[string]any{"tahun": 1, "siswa_ids": []int64{sid}}); code != 200 {
		t.Fatalf("pembina mengisi anggota: %d %s", code, b)
	}
	if code, _ := g2.do("PUT", fmt.Sprintf("/api/ekskul/%d/nilai", ek), map[string]any{"tahun": 1, "smt": 1, "items": []map[string]any{{"siswa_id": sid, "predikat": "Hebat"}}}); code != 400 {
		t.Fatalf("predikat tidak dikenal harus ditolak")
	}
	if code, b := g2.do("PUT", fmt.Sprintf("/api/ekskul/%d/nilai", ek), map[string]any{"tahun": 1, "smt": 1, "items": []map[string]any{{"siswa_id": sid, "predikat": "Sangat Baik", "keterangan": "Aktif dalam kegiatan perkemahan."}}}); code != 200 {
		t.Fatalf("pembina mengisi nilai: %d %s", code, b)
	}
	var r raporLengkap
	if code := a.json("GET", fmt.Sprintf("/api/rapor/siswa/%d?tahun=1&smt=1", sid), nil, &r); code != 200 || len(r.Ekskul) != 1 || r.Ekskul[0]["predikat"] != "Sangat Baik" {
		t.Fatalf("ekskul di rapor: %d %+v", code, r.Ekskul)
	}
	if code, _ := a.do("DELETE", fmt.Sprintf("/api/ekskul/%d", ek), nil); code != 409 {
		t.Fatalf("ekskul yang sudah dinilai tidak boleh dihapus")
	}
}

// Portofolio: siswa mengelola karyanya sendiri; guru mapel memberi catatan; siswa/guru lain ditolak.
func TestPortofolio(t *testing.T) {
	e := newEnv(t)
	s1 := e.login("siswa001", "siswa123")
	s2 := e.login("siswa002", "siswa123")
	g1 := e.login("guru001", "guru123") // Matematika 7A
	g2 := e.login("guru002", "guru123")
	if code, _ := s1.do("POST", "/api/portofolio", map[string]any{"judul": "Kosong"}); code != 400 {
		t.Fatalf("karya tanpa isi harus ditolak")
	}
	if code, b := s1.do("POST", "/api/portofolio", map[string]any{"judul": "Poster Bilangan Bulat", "course_id": 1, "deskripsi": "Poster garis bilangan", "file": pngDataURL(t)}); code != 200 {
		t.Fatalf("unggah karya: %d %s", code, b)
	}
	if code, _ := s1.do("POST", "/api/portofolio", map[string]any{"judul": "Mapel kelas lain", "course_id": 3, "deskripsi": "x"}); code != 400 {
		t.Fatalf("mapel kelas lain tidak boleh dipilih")
	}
	id := e.id(`SELECT id FROM portfolios WHERE judul='Poster Bilangan Bulat'`)
	var list2 []map[string]any
	s2.json("GET", "/api/portofolio", nil, &list2)
	if len(list2) != 0 {
		t.Fatalf("siswa lain tidak boleh melihat karya siswa001")
	}
	if code, _ := s2.do("DELETE", fmt.Sprintf("/api/portofolio/%d", id), nil); code != 403 {
		t.Fatalf("siswa lain tidak boleh menghapus")
	}
	if code, _ := g2.do("PUT", fmt.Sprintf("/api/portofolio/%d/catatan", id), map[string]any{"catatan": "x", "nilai": 90}); code != 403 {
		t.Fatalf("guru yang tidak mengampu mapel itu harus ditolak")
	}
	if code, b := g1.do("PUT", fmt.Sprintf("/api/portofolio/%d/catatan", id), map[string]any{"catatan": "Rapi dan jelas.", "nilai": 92}); code != 200 {
		t.Fatalf("guru mapel memberi catatan: %d %s", code, b)
	}
	if code, _ := s1.do("PUT", fmt.Sprintf("/api/portofolio/%d", id), map[string]any{"judul": "Ganti", "deskripsi": "x"}); code != 400 {
		t.Fatalf("karya yang sudah dinilai tidak boleh diubah siswa")
	}
	var mp struct {
		Siswa []struct {
			ID         int64 `json:"id"`
			Portofolio any   `json:"portofolio"`
		}
	}
	g1.json("GET", "/api/rapor/mapel?course_id=1&tahun=1&smt=1", nil, &mp)
	ok := false
	for _, s := range mp.Siswa {
		if s.ID == e.id(`SELECT id FROM users WHERE username='siswa001'`) && toF(s.Portofolio) == 1 {
			ok = true
		}
	}
	if !ok {
		t.Fatalf("jumlah portofolio siswa harus tampil di rapor mapel")
	}
}

// Saran deskripsi capaian dari AI: satu permintaan per kelas, nama siswa tidak dikirim, memakai kuota harian.
func TestRaporSaranAI(t *testing.T) {
	e := newEnv(t)
	var diterima string
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		diterima = string(b)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"choices":[{"message":{"content":"Berikut:\n[{\"no\":1,\"tinggi\":\"Menunjukkan penguasaan yang baik dalam bilangan bulat.\",\"perlu\":\"Perlu bimbingan dalam pecahan.\"}]"}}]}`)
	}))
	defer fake.Close()
	kvSet(e.app.db, "cfg_ai", `{"base_url":"`+fake.URL+`","model":"palsu","kuota_harian":5,"suhu":0.7}`)
	g := e.login("guru001", "guru123")
	var res struct {
		Saran []struct {
			SiswaID int64  `json:"siswa_id"`
			Tinggi  string `json:"tinggi"`
		}
	}
	if code := g.json("POST", "/api/rapor/mapel/ai", map[string]any{"course_id": 1, "tahun": 1, "smt": 1}, &res); code != 200 || len(res.Saran) != 1 || res.Saran[0].SiswaID == 0 {
		t.Fatalf("saran AI: %d %+v", code, res)
	}
	if strings.Contains(diterima, "Andi") || strings.Contains(diterima, "siswa001") {
		t.Fatalf("nama/username siswa tidak boleh dikirim ke penyedia AI")
	}
	if !strings.Contains(diterima, "Siswa 1:") {
		t.Fatalf("data nilai per materi harus dikirim dengan nomor urut")
	}
	if n := e.id(`SELECT COUNT(*) FROM kv WHERE k LIKE 'ai_pakai:%'`); n != 1 {
		t.Fatalf("permintaan AI harus dihitung ke kuota, dapat %d", n)
	}
}

func TestSaranCapaianDariNilai(t *testing.T) {
	asg := []map[string]any{{"id": int64(1), "judul": "Latihan Pecahan", "nilai_maks": float64(100)}, {"id": int64(2), "judul": "Tugas Persamaan Linear", "nilai_maks": float64(50)}}
	subs := []map[string]any{{"assignment_id": int64(1), "siswa_id": int64(7), "nilai": float64(95)}, {"assignment_id": int64(2), "siswa_id": int64(7), "nilai": float64(30)}}
	att := []map[string]any{{"quiz_id": int64(1), "judul": "Kuis Pecahan", "siswa_id": int64(7), "skor": float64(90)}}
	tinggi, perlu := saranCapaian("", asg, subs, att, 7)
	if tinggi != "Menunjukkan penguasaan yang sangat baik dalam materi Pecahan." || perlu != "Perlu bimbingan dalam materi Persamaan Linear." {
		t.Fatalf("saran: %q / %q", tinggi, perlu)
	}
	if a, b := saranCapaian("", asg, subs, att, 99); a != "" || b != "" {
		t.Fatalf("siswa tanpa nilai tidak diberi saran")
	}
}

func TestSaranCapaianMateriSama(t *testing.T) {
	asg := []map[string]any{{"id": int64(1), "judul": "Latihan Bilangan Bulat", "nilai_maks": float64(100)}}
	subs := []map[string]any{{"assignment_id": int64(1), "siswa_id": int64(7), "nilai": float64(85)}}
	att := []map[string]any{{"quiz_id": int64(1), "judul": "Kuis Bilangan Bulat", "siswa_id": int64(7), "skor": float64(66.7)}}
	tinggi, perlu := saranCapaian("", asg, subs, att, 7)
	if tinggi != "Menunjukkan penguasaan yang baik dalam materi Bilangan Bulat." || perlu != "Perlu meningkatkan ketelitian dalam mengerjakan soal materi Bilangan Bulat." {
		t.Fatalf("saran: %q / %q", tinggi, perlu)
	}
}
