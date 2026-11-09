package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Modul Asisten AI: generator RPM, LKPD, soal, surat, dan tanya bebas.
// Memakai API apa pun yang kompatibel OpenAI (OpenAI, Gemini, Anthropic, OpenRouter, Ollama, dll).
// API key disimpan di server dan tidak pernah dikirim ke browser.

func init() {
	configDefaults["ai"] = `{"base_url":"","model":"","kuota_harian":30,"suhu":0.7}`
	moduleSchemas = append(moduleSchemas, `
CREATE TABLE IF NOT EXISTS ai_documents (
  id         INTEGER PRIMARY KEY,
  user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  template   TEXT NOT NULL,
  judul      TEXT NOT NULL,
  input      TEXT NOT NULL DEFAULT '{}',
  output     TEXT NOT NULL,
  model      TEXT,
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_aidocs_user ON ai_documents(user_id, created_at);
`)
	moduleRoutes = append(moduleRoutes, func(a *App, m *http.ServeMux) {
		x := func(roles string, fn func(*Ctx) error) http.HandlerFunc { return a.mod("ai", roles, fn) }
		m.HandleFunc("GET /api/ai/config", x("admin", apiAIConfig))
		m.HandleFunc("PUT /api/ai/config", x("admin", apiAIConfigSave))
		m.HandleFunc("POST /api/ai/models", x("admin", apiAIModels))
		m.HandleFunc("GET /api/ai/status", x("admin,guru", apiAIStatus))
		m.HandleFunc("POST /api/ai/generate", x("admin,guru", apiAIGenerate))
		m.HandleFunc("GET /api/ai/docs", x("admin,guru", apiAIDocs))
		m.HandleFunc("GET /api/ai/docs/{id}", x("admin,guru", apiAIDoc))
		m.HandleFunc("DELETE /api/ai/docs/{id}", x("admin,guru", apiAIDocDelete))
	})
}

// ---------------------------------------------------------------------
// Konfigurasi
// ---------------------------------------------------------------------
func aiKeyHint(k string) string {
	if k == "" {
		return ""
	}
	if len(k) <= 4 {
		return "••••"
	}
	return "••••" + k[len(k)-4:]
}

// Pengaturan AI hanya untuk admin sungguhan, bukan guru yang diberi akses modul AI:
// pengaturan ini memegang API key berbayar milik sekolah.
func aiAdminOnly(c *Ctx) error {
	if c.U.Role != "admin" {
		return apiErr{403, "Pengaturan Asisten AI hanya untuk admin."}
	}
	return nil
}

func apiAIConfig(c *Ctx) error {
	if err := aiAdminOnly(c); err != nil {
		return err
	}
	cfg := getConfig(c.A.db, "ai")
	cfg["key_hint"] = aiKeyHint(kvGet(c.A.db, "ai_key"))
	return c.JSON(cfg)
}

func normBase(u string) string { return strings.TrimRight(strings.TrimSpace(u), "/") }

func apiAIConfigSave(c *Ctx) error {
	var in struct {
		BaseURL     string   `json:"base_url"`
		APIKey      *string  `json:"api_key"` // kosong = tidak diubah
		HapusKey    bool     `json:"hapus_key"`
		Model       string   `json:"model"`
		KuotaHarian *float64 `json:"kuota_harian"`
		Suhu        *float64 `json:"suhu"`
	}
	if err := aiAdminOnly(c); err != nil {
		return err
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	base := normBase(in.BaseURL)
	if base != "" && !urlRe.MatchString(base) {
		return bad("URL endpoint harus diawali http:// atau https://")
	}
	cfg := getConfig(c.A.db, "ai")
	oldBase, _ := cfg["base_url"].(string)
	newKey := in.APIKey != nil && strings.TrimSpace(*in.APIKey) != ""
	if !newKey && !sameHost(oldBase, base) && kvGet(c.A.db, "ai_key") != "" {
		// endpoint berganti tanpa key baru: hapus key lama agar tidak terkirim ke host lain
		in.HapusKey = true
	}
	cfg["base_url"], cfg["model"] = base, strings.TrimSpace(in.Model)
	if in.KuotaHarian != nil && *in.KuotaHarian >= 0 {
		cfg["kuota_harian"] = *in.KuotaHarian
	}
	if in.Suhu != nil && *in.Suhu >= 0 && *in.Suhu <= 2 {
		cfg["suhu"] = *in.Suhu
	}
	b, _ := json.Marshal(cfg)
	if err := kvSet(c.A.db, "cfg_ai", string(b)); err != nil {
		return err
	}
	if in.HapusKey {
		kvSet(c.A.db, "ai_key", "")
	} else if in.APIKey != nil && strings.TrimSpace(*in.APIKey) != "" {
		kvSet(c.A.db, "ai_key", strings.TrimSpace(*in.APIKey))
	}
	return c.OK()
}

// Jangan ikuti redirect: header API key bisa ikut terkirim ke host tujuan redirect.
var aiHTTP = &http.Client{Timeout: 0, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}

// sameHost: apakah dua base URL menunjuk skema+host+port yang sama.
func sameHost(a, b string) bool {
	ua, e1 := url.Parse(a)
	ub, e2 := url.Parse(b)
	return e1 == nil && e2 == nil && ua.Host != "" && strings.EqualFold(ua.Scheme, ub.Scheme) && strings.EqualFold(ua.Host, ub.Host)
}

// aiInflight: generate yang sedang berjalan per pengguna (agar kuota tidak bisa diakali dengan permintaan paralel).
var aiInflight = struct {
	sync.Mutex
	m map[int64]int
}{m: map[int64]int{}}

func aiRequest(ctx context.Context, method, url, key string, body any) (*http.Response, error) {
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, rd)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
		req.Header.Set("x-api-key", key) // beberapa penyedia (Anthropic) memakai header ini
	}
	return aiHTTP.Do(req)
}

func upstreamErr(res *http.Response) error {
	b, _ := io.ReadAll(io.LimitReader(res.Body, 2000))
	msg := strings.TrimSpace(string(b))
	var e struct {
		Error any `json:"error"`
	}
	if json.Unmarshal(b, &e) == nil && e.Error != nil {
		switch v := e.Error.(type) {
		case string:
			msg = v
		case map[string]any:
			if m, ok := v["message"].(string); ok {
				msg = m
			}
		}
	}
	if len(msg) > 300 {
		msg = msg[:300] + "…"
	}
	return bad(fmt.Sprintf("Penyedia AI menolak (%d): %s", res.StatusCode, msg))
}

// apiAIModels mengambil daftar model dari {base_url}/models. Bisa memakai URL/key yang belum disimpan.
func apiAIModels(c *Ctx) error {
	var in struct {
		BaseURL string `json:"base_url"`
		APIKey  string `json:"api_key"`
	}
	if err := aiAdminOnly(c); err != nil {
		return err
	}
	c.Bind(&in)
	saved, _ := getConfig(c.A.db, "ai")["base_url"].(string)
	base := normBase(in.BaseURL)
	if base == "" {
		base = saved
	}
	key := strings.TrimSpace(in.APIKey)
	if key == "" {
		// key tersimpan hanya boleh dikirim ke endpoint yang tersimpan (cegah pencurian key)
		if !sameHost(saved, base) {
			return bad("Endpoint berbeda dari yang tersimpan. Tempel API key untuk endpoint ini.")
		}
		key = kvGet(c.A.db, "ai_key")
	}
	if base == "" {
		return bad("Isi URL endpoint terlebih dahulu.")
	}
	ctx, cancel := context.WithTimeout(c.R.Context(), 20*time.Second)
	defer cancel()
	res, err := aiRequest(ctx, "GET", base+"/models", key, nil)
	if err != nil {
		log.Println("ai: model list:", err)
		return bad("Tidak bisa menghubungi endpoint. Periksa URL, koneksi internet server, dan API key.")
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return upstreamErr(res)
	}
	var out struct {
		Data []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 8<<20)).Decode(&out); err != nil {
		return bad("Jawaban endpoint /models tidak dikenali.")
	}
	ids := []string{}
	for _, m := range out.Data {
		if m.ID != "" {
			ids = append(ids, m.ID)
		}
	}
	for _, m := range out.Models {
		ids = append(ids, strings.TrimPrefix(m.Name, "models/"))
	}
	sort.Strings(ids)
	if len(ids) == 0 {
		return bad("Endpoint tidak mengembalikan daftar model.")
	}
	return c.JSON(map[string]any{"models": ids})
}

// Pemakaian harian dihitung di tabel kv (ai_pakai:<tanggal sekolah>:<user>), bukan dari jumlah dokumen,
// agar kuota tidak bisa diakali dengan menghapus dokumen. Tidak lewat kvCache (ditulis dengan SQL langsung).
func aiUsageKey(uid int64) string {
	return "ai_pakai:" + localDate() + ":" + strconv.FormatInt(uid, 10)
}

func (c *Ctx) aiUsedToday() int {
	var n int
	c.A.db.QueryRow(`SELECT CAST(v AS INTEGER) FROM kv WHERE k=?`, aiUsageKey(c.U.ID)).Scan(&n)
	return n
}

func (c *Ctx) aiCountUse() {
	c.A.db.Exec(`INSERT INTO kv (k, v) VALUES (?, '1') ON CONFLICT (k) DO UPDATE SET v=CAST(v AS INTEGER)+1`, aiUsageKey(c.U.ID))
}

func apiAIStatus(c *Ctx) error {
	cfg := getConfig(c.A.db, "ai")
	kuota, _ := cfg["kuota_harian"].(float64)
	ready := cfg["base_url"] != "" && cfg["model"] != ""
	out := map[string]any{"siap": ready, "admin": c.U.Role == "admin", "terpakai": c.aiUsedToday(), "kuota": kuota}
	if c.U.Role == "admin" { // nama model hanya untuk admin
		out["model"] = cfg["model"]
	}
	return c.JSON(out)
}

// ---------------------------------------------------------------------
// Template prompt
// ---------------------------------------------------------------------
type aiTemplate struct {
	Judul  func(in map[string]string) string
	Prompt func(in map[string]string, sekolah string) string
}

func v(in map[string]string, k, def string) string {
	if s := strings.TrimSpace(in[k]); s != "" {
		return s
	}
	return def
}

var aiTemplates = map[string]aiTemplate{
	"rpm": {
		Judul: func(in map[string]string) string { return "RPM " + v(in, "mapel", "") + " – " + v(in, "materi", "") },
		Prompt: func(in map[string]string, sekolah string) string {
			return fmt.Sprintf(`Susun Rencana Pembelajaran Mendalam (RPM) Kurikulum Merdeka dengan pendekatan Pembelajaran Mendalam (berkesadaran/mindful, bermakna/meaningful, menggembirakan/joyful).

Data:
- Satuan pendidikan: %s
- Mata pelajaran: %s
- Fase/Kelas: %s
- Materi pokok: %s
- Alokasi waktu: %s, %s pertemuan
- Model/metode yang diinginkan: %s
- Dimensi profil lulusan yang ditekankan: %s
- Tujuan pembelajaran (jika diisi guru): %s
- Catatan guru: %s

Struktur wajib (gunakan heading Markdown):
1. Identitas
2. Identifikasi (kesiapan peserta didik, karakteristik materi, dimensi profil lulusan)
3. Desain Pembelajaran (capaian pembelajaran, lintas disiplin ilmu, tujuan pembelajaran, topik, praktik pedagogis, kemitraan, lingkungan belajar, pemanfaatan digital)
4. Pengalaman Belajar per pertemuan: Awal (berkesadaran), Inti (memahami, mengaplikasi, merefleksi), Penutup – sertakan estimasi menit
5. Asesmen (awal, proses, akhir) beserta contoh instrumen dan rubrik singkat dalam tabel
6. Lampiran singkat (LKPD ringkas / bahan bacaan / glosarium)`,
				sekolah, v(in, "mapel", "-"), v(in, "kelas", "-"), v(in, "materi", "-"), v(in, "waktu", "2 x 40 menit"), v(in, "pertemuan", "1"),
				v(in, "model", "pilih yang paling sesuai"), v(in, "profil", "pilih yang paling sesuai"), v(in, "tujuan", "rumuskan sendiri"), v(in, "catatan", "-"))
		},
	},
	"lkpd": {
		Judul: func(in map[string]string) string { return "LKPD " + v(in, "mapel", "") + " – " + v(in, "materi", "") },
		Prompt: func(in map[string]string, sekolah string) string {
			return fmt.Sprintf(`Buat Lembar Kerja Peserta Didik (LKPD) yang siap dicetak.
- Sekolah: %s
- Mata pelajaran: %s
- Kelas: %s
- Materi: %s
- Tujuan pembelajaran: %s
- Jenis kegiatan: %s
- Jumlah soal/aktivitas: %s
- Catatan guru: %s

Struktur: judul, identitas (nama, kelas, kelompok – berupa garis isian), tujuan, petunjuk kerja, ringkasan materi singkat, kegiatan bertahap dengan ruang jawaban (garis titik-titik), soal refleksi, dan kolom penilaian. Bahasa sesuai usia siswa.`,
				sekolah, v(in, "mapel", "-"), v(in, "kelas", "-"), v(in, "materi", "-"), v(in, "tujuan", "rumuskan sendiri"), v(in, "kegiatan", "diskusi kelompok dan latihan"), v(in, "jumlah", "5"), v(in, "catatan", "-"))
		},
	},
	"soal": {
		Judul: func(in map[string]string) string { return "Soal " + v(in, "mapel", "") + " – " + v(in, "materi", "") },
		Prompt: func(in map[string]string, sekolah string) string {
			n := v(in, "opsi", "4")
			letters := map[string]string{"3": "A–C", "4": "A–D", "5": "A–E"}[n]
			if letters == "" {
				letters = "A–D"
			}
			return fmt.Sprintf(`Buat %s soal pilihan ganda.
- Mata pelajaran: %s
- Kelas: %s
- Materi: %s
- Tingkat kesulitan: %s (gunakan variasi level kognitif C1–C6 sesuai tingkat)
- Pilihan jawaban: %s
- Catatan guru: %s

FORMAT WAJIB (agar bisa diimpor otomatis). Jangan menambah teks pembuka/penutup, jangan memakai tabel atau huruf tebal:
1. Teks pertanyaan
A. pilihan
B. pilihan
C. pilihan
D. pilihan
Kunci: B
Pembahasan: penjelasan singkat

Ulangi untuk setiap nomor. Kunci jawaban harus tersebar merata.`,
				v(in, "jumlah", "10"), v(in, "mapel", "-"), v(in, "kelas", "-"), v(in, "materi", "-"), v(in, "tingkat", "sedang"), letters, v(in, "catatan", "-"))
		},
	},
	"surat": {
		Judul: func(in map[string]string) string { return "Surat " + v(in, "jenis", "") },
		Prompt: func(in map[string]string, sekolah string) string {
			return fmt.Sprintf(`Buat surat resmi sekolah dalam Bahasa Indonesia baku sesuai tata naskah dinas.
- Kop: %s
- Jenis surat: %s
- Ditujukan kepada: %s
- Isi pokok / keperluan: %s
- Waktu & tempat kegiatan (jika ada): %s
- Penanda tangan: %s

Sertakan: kop sekolah (nama & alamat), nomor surat (beri placeholder ......), lampiran, perihal, tujuan, salam pembuka, isi, penutup, tempat & tanggal, jabatan, nama, dan NIP penanda tangan. Jangan menambah informasi yang tidak diberikan; pakai placeholder [....] bila perlu.`,
				sekolah, v(in, "jenis", "pemberitahuan"), v(in, "tujuan", "Bapak/Ibu Orang Tua/Wali Murid"), v(in, "isi", "-"), v(in, "waktu", "-"), v(in, "ttd", "Kepala Sekolah"))
		},
	},
	"prota": {
		Judul: func(in map[string]string) string {
			return "Prota & Promes " + v(in, "mapel", "") + " " + v(in, "kelas", "")
		},
		Prompt: func(in map[string]string, sekolah string) string {
			return fmt.Sprintf(`Susun Program Tahunan (Prota) dan Program Semester (Promes) Kurikulum Merdeka.
- Sekolah: %s
- Mata pelajaran: %s
- Fase/Kelas: %s
- Tahun ajaran: %s
- Jam pelajaran per minggu: %s
- Daftar materi/bab (jika diisi guru): %s
- Catatan guru (kalender, libur, ujian): %s

Struktur (heading Markdown):
1. Identitas
2. Program Tahunan: tabel (No, Semester, Tujuan Pembelajaran / Materi, Alokasi JP), total JP per semester
3. Program Semester Ganjil: tabel per bulan Juli–Desember dengan kolom minggu ke-1 s.d. 5, tandai minggu efektif dan alokasi JP; sisakan minggu untuk asesmen sumatif tengah & akhir semester
4. Program Semester Genap: tabel per bulan Januari–Juni dengan format sama
5. Rincian minggu efektif per semester (tabel bulan, jumlah minggu, minggu tidak efektif, minggu efektif)
Bila daftar materi kosong, susun berdasarkan Capaian Pembelajaran fase tersebut. Jangan mengarang tanggal libur; gunakan placeholder [....] bila perlu.`,
				sekolah, v(in, "mapel", "-"), v(in, "kelas", "-"), v(in, "tahun", "-"), v(in, "jp", "4"), v(in, "materi", "susun sendiri dari CP"), v(in, "catatan", "-"))
		},
	},
	"kktp": {
		Judul: func(in map[string]string) string { return "KKTP " + v(in, "mapel", "") + " – " + v(in, "materi", "") },
		Prompt: func(in map[string]string, sekolah string) string {
			return fmt.Sprintf(`Susun Kriteria Ketercapaian Tujuan Pembelajaran (KKTP) Kurikulum Merdeka.
- Sekolah: %s
- Mata pelajaran: %s
- Fase/Kelas: %s
- Materi / lingkup materi: %s
- Tujuan pembelajaran (jika diisi guru): %s
- Pendekatan KKTP: %s
- Catatan guru: %s

Isi wajib (heading Markdown):
1. Identitas dan tujuan pembelajaran (rumuskan bila kosong, 2–4 TP)
2. KKTP dengan pendekatan yang dipilih, dalam tabel:
   - Deskripsi kriteria: kriteria per TP dan indikator bukti
   - Rubrik: level Perlu bimbingan / Cukup / Baik / Sangat baik dengan deskriptor jelas
   - Interval nilai: rentang 0–40, 41–65, 66–85, 86–100 beserta tindak lanjut tiap rentang
3. Contoh instrumen asesmen singkat untuk mengukur kriteria
4. Tindak lanjut (remedial dan pengayaan)`,
				sekolah, v(in, "mapel", "-"), v(in, "kelas", "-"), v(in, "materi", "-"), v(in, "tujuan", "rumuskan sendiri"), v(in, "pendekatan", "rubrik"), v(in, "catatan", "-"))
		},
	},
	"bebas": {
		Judul: func(in map[string]string) string {
			t := v(in, "pertanyaan", "Tanya bebas")
			if len([]rune(t)) > 60 {
				t = string([]rune(t)[:60]) + "…"
			}
			return t
		},
		Prompt: func(in map[string]string, sekolah string) string { return v(in, "pertanyaan", "") },
	},
}

// ---------------------------------------------------------------------
// Generate (streaming)
// ---------------------------------------------------------------------
func apiAIGenerate(c *Ctx) error {
	var in struct {
		Template string         `json:"template"`
		Inputs   map[string]any `json:"inputs"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	tpl, ok := aiTemplates[in.Template]
	if !ok {
		return bad("Jenis dokumen tidak dikenal.")
	}
	clean := map[string]string{}
	for k, raw := range in.Inputs {
		if len(k) > 30 || raw == nil {
			continue
		}
		r := []rune(strings.TrimSpace(fmt.Sprint(raw)))
		if len(r) > 3000 {
			r = r[:3000]
		}
		clean[k] = string(r)
	}
	if in.Template == "bebas" && clean["pertanyaan"] == "" {
		return bad("Tulis pertanyaan terlebih dahulu.")
	}
	cfg := getConfig(c.A.db, "ai")
	base, _ := cfg["base_url"].(string)
	model, _ := cfg["model"].(string)
	if base == "" || model == "" {
		return bad("Asisten AI belum diatur. Admin perlu mengisi endpoint dan memilih model di menu Asisten AI → Pengaturan.")
	}
	aiInflight.Lock()
	running := aiInflight.m[c.U.ID]
	if running >= 2 {
		aiInflight.Unlock()
		return bad("Tunggu dokumen sebelumnya selesai dibuat.")
	}
	if kuota, _ := cfg["kuota_harian"].(float64); c.U.Role != "admin" && kuota > 0 && float64(c.aiUsedToday()+running) >= kuota {
		aiInflight.Unlock()
		return bad(fmt.Sprintf("Kuota harian Anda (%d dokumen) sudah habis. Coba lagi besok.", int(kuota)))
	}
	aiInflight.m[c.U.ID]++
	aiInflight.Unlock()
	defer func() {
		aiInflight.Lock()
		if aiInflight.m[c.U.ID]--; aiInflight.m[c.U.ID] <= 0 {
			delete(aiInflight.m, c.U.ID)
		}
		aiInflight.Unlock()
	}()
	suhu, _ := cfg["suhu"].(float64)
	var sekolah, alamat, kepsek, nip string
	c.A.db.QueryRow(`SELECT nama_sekolah, COALESCE(alamat,''), COALESCE(kepala_sekolah,''), COALESCE(nip_kepala,'') FROM settings WHERE id=1`).Scan(&sekolah, &alamat, &kepsek, &nip)
	sekolahInfo := sekolah
	if alamat != "" {
		sekolahInfo += ", " + alamat
	}
	if in.Template == "surat" && clean["ttd"] == "" && kepsek != "" {
		clean["ttd"] = "Kepala Sekolah, " + kepsek + " (NIP " + nip + ")"
	}
	system := "Anda adalah asisten administrasi dan perencanaan pembelajaran untuk guru di " + sekolah +
		", Indonesia. Jawab dalam Bahasa Indonesia yang baku, jelas, dan sesuai Kurikulum Merdeka bila relevan. " +
		"Gunakan format Markdown (heading, daftar, tabel bila perlu). Jangan mengarang data pribadi siswa."

	body := map[string]any{
		"model":       model,
		"stream":      true,
		"temperature": suhu,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": tpl.Prompt(clean, sekolahInfo)},
		},
	}
	ctx, cancel := context.WithTimeout(c.R.Context(), 10*time.Minute)
	defer cancel()
	res, err := aiRequest(ctx, "POST", base+"/chat/completions", kvGet(c.A.db, "ai_key"), body)
	if err != nil {
		log.Println("ai: generate:", err)
		return bad("Tidak bisa menghubungi penyedia AI. Periksa koneksi internet server, lalu coba lagi.")
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return upstreamErr(res)
	}

	// Kirim teks bertahap ke browser (mencegah putus oleh batas waktu proxy saat jawaban panjang)
	rc := http.NewResponseController(c.W)
	rc.SetWriteDeadline(time.Now().Add(11 * time.Minute))
	c.W.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.W.Header().Set("Cache-Control", "no-store")
	c.W.Header().Set("X-Accel-Buffering", "no")
	c.W.WriteHeader(200)
	var out strings.Builder
	write := func(s string) {
		if s == "" {
			return
		}
		out.WriteString(s)
		io.WriteString(c.W, s)
		rc.Flush()
	}
	if strings.Contains(res.Header.Get("Content-Type"), "text/event-stream") {
		sc := bufio.NewScanner(res.Body)
		sc.Buffer(make([]byte, 64*1024), 4<<20)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if !strings.HasPrefix(line, "data:") {
				continue
			}
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data == "[DONE]" {
				break
			}
			var ev struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
				Error any `json:"error"`
			}
			if json.Unmarshal([]byte(data), &ev) != nil {
				continue
			}
			if ev.Error != nil {
				log.Printf("ai: penyedia mengirim galat: %v", ev.Error)
				msg := "penyedia AI menolak permintaan (kuota habis, model tidak tersedia, atau isi diblokir)"
				if m, ok := ev.Error.(map[string]any); ok {
					if s, ok := m["message"].(string); ok && s != "" {
						msg = s
					}
				} else if s, ok := ev.Error.(string); ok && s != "" {
					msg = s
				}
				write("\n\n[Gagal: " + msg + "]")
				break
			}
			if len(ev.Choices) > 0 {
				write(ev.Choices[0].Delta.Content)
			}
		}
		if err := sc.Err(); err != nil {
			// sambungan ke penyedia putus di tengah jalan: tandai agar dokumen tidak tersimpan seolah lengkap
			write("\n\n[Jawaban terpotong: sambungan ke penyedia AI terputus. Coba buat ulang.]")
		}
	} else {
		// penyedia yang tidak mendukung streaming
		var full struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}
		if json.NewDecoder(res.Body).Decode(&full) == nil && len(full.Choices) > 0 {
			write(full.Choices[0].Message.Content)
		}
	}
	text := strings.TrimSpace(out.String())
	if text == "" {
		io.WriteString(c.W, "[Tidak ada jawaban dari model. Coba model lain.]")
		return nil
	}
	ib, _ := json.Marshal(clean)
	c.aiCountUse() // dokumen berhasil dibuat: dihitung ke kuota (walau nanti dihapus)
	res2, err := c.A.db.Exec(`INSERT INTO ai_documents (user_id, template, judul, input, output, model, created_at) VALUES (?,?,?,?,?,?,?)`,
		c.U.ID, in.Template, tpl.Judul(clean), string(ib), text, model, nowStr())
	if err == nil {
		id, _ := res2.LastInsertId()
		io.WriteString(c.W, fmt.Sprintf("\n\u0000DOC:%d", id)) // penanda akhir: id dokumen tersimpan
		rc.Flush()
	}
	return nil
}

func apiAIDocs(c *Ctx) error {
	rows, err := c.list(`SELECT id, template, judul, model, created_at FROM ai_documents WHERE user_id=? ORDER BY created_at DESC LIMIT 200`, c.U.ID)
	if err != nil {
		return err
	}
	if c.U.Role != "admin" {
		for _, r := range rows {
			delete(r, "model")
		}
	}
	return c.JSON(rows)
}

func apiAIDoc(c *Ctx) error {
	rows, err := c.list(`SELECT * FROM ai_documents WHERE id=? AND user_id=?`, c.ID("id"), c.U.ID)
	if err != nil || len(rows) == 0 {
		return notFound()
	}
	if c.U.Role != "admin" {
		delete(rows[0], "model")
	}
	return c.JSON(rows[0])
}

func apiAIDocDelete(c *Ctx) error {
	res, err := c.A.db.Exec(`DELETE FROM ai_documents WHERE id=? AND user_id=?`, c.ID("id"), c.U.ID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return notFound()
	}
	return c.OK()
}
