package main

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRPPFaseDanWaktu(t *testing.T) {
	cases := []struct {
		kelas      string
		fase       string
		menit, tot int
	}{
		{"7A", "Fase D", 40, 80}, {"VIII B", "Fase D", 40, 80}, {"Kelas 10", "Fase E", 45, 90}, {"XI IPA", "Fase F", 45, 90},
		{"2", "Fase A", 35, 70}, {"Fase C (Kelas 5–6)", "Fase C", 35, 70}, {"Fase F (Kelas 11–12)", "Fase F", 45, 90},
	}
	for _, c := range cases {
		_, menit, _, total, fase := rppWaktu(map[string]string{"kelas": c.kelas, "jp": "2"})
		if fase != c.fase || menit != c.menit || total != c.tot {
			t.Errorf("%q: fase %q menit %d total %d, harusnya %q %d %d", c.kelas, fase, menit, total, c.fase, c.menit, c.tot)
		}
	}
	// lama JP diisi guru menimpa bawaan jenjang
	if _, m, p, tot, _ := rppWaktu(map[string]string{"kelas": "7A", "jp": "3", "menit_jp": "35", "pertemuan": "2"}); m != 35 || p != 2 || tot != 105 {
		t.Errorf("menit_jp guru: %d %d %d", m, p, tot)
	}
}

func TestRPPPromptFormat(t *testing.T) {
	in := map[string]string{"mapel": "Matematika", "kelas": "7A", "materi": "Aljabar", "jp": "2", "pertemuan": "2"}
	for format, kunci := range map[string]string{"rpm": "RENCANA PEMBELAJARAN MENDALAM", "modul": "MODUL AJAR", "rpp1": "RPP SATU LEMBAR", "": "RENCANA PEMBELAJARAN MENDALAM"} {
		in["format"] = format
		p := rppPrompt(in, "SMP Uji")
		for _, w := range []string{kunci, "Fase D", "tepat 80 menit", "Rincian waktu: awal A menit", "Permendikdasmen No. 13 Tahun 2025", "Kepka BSKAP 046/H/KR/2025", "Kewargaan"} {
			if !strings.Contains(p, w) {
				t.Errorf("format %q: prompt tidak memuat %q", format, w)
			}
		}
	}
	in["cp"] = "Peserta didik dapat ..."
	if p := rppPrompt(in, "SMP Uji"); strings.Contains(p, "Ringkasan CP – cocokkan") {
		t.Error("CP ditempel guru tetapi AI tetap diminta merangkum CP")
	}
	if j := rppJudul(map[string]string{"format": "modul", "mapel": "IPA", "materi": "Energi"}); j != "Modul Ajar IPA – Energi" {
		t.Errorf("judul: %q", j)
	}
}

// file .docx: zip berisi XML yang valid, memuat kop, logo, tabel, daftar, dan tanda tangan
func TestDocxValid(t *testing.T) {
	dir := t.TempDir()
	logo := filepath.Join(dir, "logo.png")
	f, _ := os.Create(logo)
	img := image.NewRGBA(image.Rect(0, 0, 80, 40))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+3] = 200, 255 // merah, terlihat di pratinjau
	}
	png.Encode(f, img)
	f.Close()
	md := "# RPM Matematika\n\n## 1. Identitas\n| Komponen | Isi |\n|---|---|\n| Mapel | **Matematika** & <IPA> |\n\n- satu\n- dua *miring*\n1. langkah `x+1`\n\n---\nParagraf biasa\nbaris dua"
	b, err := buildDocx(md, docxOpts{Judul: "RPM", Kop: true, NamaSek: "SMP Uji", AlamatSek: "Jl. Merdeka 1", LogoPath: logo,
		TTD: true, Tempat: "Bandung", Tanggal: tanggalPanjang("2026-11-10"), Guru: "Bu Rina", NIPGuru: "1990", Kepsek: "Pak Budi"})
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		t.Fatalf("bukan zip: %v", err)
	}
	got := map[string]string{}
	for _, zf := range zr.File {
		rc, _ := zf.Open()
		data, _ := io.ReadAll(rc)
		rc.Close()
		got[zf.Name] = string(data)
		if strings.HasSuffix(zf.Name, ".xml") || strings.HasSuffix(zf.Name, ".rels") {
			d := xml.NewDecoder(bytes.NewReader(data))
			for {
				if _, err := d.Token(); err == io.EOF {
					break
				} else if err != nil {
					t.Fatalf("%s: XML tidak valid: %v", zf.Name, err)
				}
			}
		}
	}
	for _, name := range []string{"[Content_Types].xml", "_rels/.rels", "word/document.xml", "word/styles.xml", "word/_rels/document.xml.rels", "word/media/logo.png"} {
		if _, ok := got[name]; !ok {
			t.Errorf("bagian %s tidak ada", name)
		}
	}
	doc := got["word/document.xml"]
	for _, w := range []string{"SMP Uji", "Heading1", "<w:tbl>", "&amp; &lt;IPA&gt;", "<w:b/>", "<w:i/>", "Consolas", "Mengetahui,", "Bandung, 10 November 2026", "Bu Rina", "NIP. 1990", "Pak Budi", "NIP. ……", "rIdLogo"} {
		if !strings.Contains(doc, w) {
			t.Errorf("document.xml tidak memuat %q", w)
		}
	}
	if out := os.Getenv("SCH_DOCX_OUT"); out != "" { // untuk diperiksa manual (LibreOffice / Word)
		if src := os.Getenv("SCH_DOCX_MD"); src != "" {
			mdb, _ := os.ReadFile(src)
			b, _ = buildDocx(string(mdb), docxOpts{Judul: "RPM", Kop: true, NamaSek: "SMP Nusantara Contoh", AlamatSek: "Jl. Pendidikan No. 1",
				TTD: true, Tempat: "Bandung", Tanggal: tanggalPanjang("2026-11-10"), Guru: "Budi Santoso, S.Pd.", Kepsek: "Drs. Ahmad Fauzi", NIPKepsek: "196805121994031002"})
		}
		os.WriteFile(out, b, 0o644)
	}
}

// sunting & unduh Word: hanya pemilik dokumen
func TestDokumenAISuntingDanWord(t *testing.T) {
	e := newEnv(t)
	g1 := e.id(`SELECT id FROM users WHERE username='guru001'`)
	e.app.db.Exec(`INSERT INTO ai_documents (user_id, template, judul, input, output, model, created_at) VALUES (?,?,?,?,?,?,?)`,
		g1, "rpm", "RPM Uji", `{"jp":"2","kelas":"7A"}`, "# RPM\nisi awal", "m", nowStr())
	doc := e.id(`SELECT id FROM ai_documents WHERE judul='RPM Uji'`)
	a := e.login("guru001", "guru123")
	b := e.login("guru002", "guru123")
	if code, body := a.do("PUT", fmt.Sprintf("/api/ai/docs/%d", doc), map[string]string{"judul": "RPM Revisi", "output": "# RPM\nisi baru"}); code != 200 {
		t.Fatalf("pemilik menyunting: %d %s", code, body)
	}
	if code, _ := b.do("PUT", fmt.Sprintf("/api/ai/docs/%d", doc), map[string]string{"judul": "x", "output": "y"}); code != 404 {
		t.Errorf("guru lain bisa menyunting dokumen orang: %d", code)
	}
	var judul, out string
	e.app.db.QueryRow(`SELECT judul, output FROM ai_documents WHERE id=?`, doc).Scan(&judul, &out)
	if judul != "RPM Revisi" || out != "# RPM\nisi baru" {
		t.Errorf("tersimpan: %q %q", judul, out)
	}
	code, body := a.raw("GET", fmt.Sprintf("/api/ai/docs/%d/docx?ttd=1&tempat=Bandung&tanggal=2026-11-10", doc), nil, true)
	if code != 200 || !bytes.HasPrefix(body, []byte("PK")) || !bytes.Contains(body, []byte("word/document.xml")) {
		t.Errorf("unduh Word: %d, %d byte", code, len(body))
	}
	if code, _ := b.raw("GET", fmt.Sprintf("/api/ai/docs/%d/docx", doc), nil, true); code != 404 {
		t.Errorf("guru lain bisa mengunduh dokumen orang: %d", code)
	}
}

func TestJenjangKelasDibatasi(t *testing.T) {
	cases := []struct {
		jenjang, kelas string
		ok             bool
		fase           string
	}{
		{"SMP", "Kelas 7 (Fase D)", true, "Fase D"}, {"SD", "Kelas 4 (Fase B)", true, "Fase B"}, {"SMA", "Kelas 11 (Fase F)", true, "Fase F"},
		{"SMK", "Kelas 10 (Fase E)", true, "Fase E"}, {"SD", "Kelas 7 (Fase D)", false, ""}, {"SMP", "Kelas 10 (Fase E)", false, ""},
		{"SMA", "", false, ""}, {"TK", "Kelas 1", false, ""},
	}
	for _, c := range cases {
		in := map[string]string{"jenjang": c.jenjang, "kelas": c.kelas}
		kon, err := aiJenjang(in)
		if (err == nil) != c.ok {
			t.Errorf("%s %q: err=%v, harusnya ok=%v", c.jenjang, c.kelas, err, c.ok)
			continue
		}
		if c.ok && (!strings.Contains(kon, c.fase) || !strings.Contains(kon, jenjangNama[c.jenjang])) {
			t.Errorf("%s %q: konteks %q", c.jenjang, c.kelas, kon)
		}
	}
	if kon, err := aiJenjang(map[string]string{"kelas": "7A"}); err != nil || kon != "" {
		t.Errorf("kiriman lama tanpa jenjang harus tetap diterima tanpa konteks: %q %v", kon, err)
	}
}

// jenjang & kelas benar-benar sampai ke AI; kelas di luar jenjang ditolak sebelum menghubungi AI
func TestJenjangMasukPrompt(t *testing.T) {
	e := newEnv(t)
	var diterima string
	fake := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		diterima = string(b)
		w.Header().Set("Content-Type", "text/event-stream")
		io.WriteString(w, "data: {\"choices\":[{\"delta\":{\"content\":\"# LKPD\"}}]}\n\ndata: [DONE]\n\n")
	}))
	defer fake.Close()
	kvSet(e.app.db, "cfg_ai", `{"base_url":"`+fake.URL+`","model":"palsu","kuota_harian":0,"suhu":0.7}`)
	g := e.login("guru001", "guru123")
	code, b := g.do("POST", "/api/ai/generate", map[string]any{"template": "lkpd", "inputs": map[string]string{"jenjang": "SD", "kelas": "Kelas 7 (Fase D)", "mapel": "Matematika", "materi": "Pecahan"}})
	if code != 400 || diterima != "" {
		t.Errorf("kelas di luar jenjang tidak ditolak: %d %s (AI dihubungi: %v)", code, b, diterima != "")
	}
	code, b = g.do("POST", "/api/ai/generate", map[string]any{"template": "lkpd", "inputs": map[string]string{"jenjang": "SD", "kelas": "Kelas 4 (Fase B)", "mapel": "Matematika", "materi": "Pecahan"}})
	if code != 200 || !strings.Contains(diterima, "SD/MI, kelas 4, Fase B") || !strings.Contains(diterima, "jangan memakai materi") {
		t.Errorf("konteks jenjang tidak sampai ke AI: %d %s", code, diterima[:min(300, len(diterima))])
	}
}
