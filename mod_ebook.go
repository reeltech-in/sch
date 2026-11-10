package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// E-book dari Sistem Informasi Perbukuan Indonesia (SIBI, buku.kemendikdasmen.go.id).
//
// Admin/pustakawan memilih buku dari katalog SIBI lalu "impor": yang disimpan hanya data buku dan alamat
// PDF/sampulnya. File PDF baru diunduh server saat ada yang PERTAMA kali membukanya, lalu disimpan di
// <data>/ebook-cache sehingga pembaca berikutnya tidak mengunduh ulang dari internet (hemat kuota, cepat,
// dan tetap bisa dibaca bila situs sumber sedang lambat). Alamat sumber tidak pernah dipakai langsung oleh browser.

const sibiAPI = "https://api.buku.cloudapp.web.id/api/catalogue/"

// host file SIBI yang dipercaya; host lama kemdikbud dialihkan ke cermin yang aktif
var sibiHosts = map[string]bool{"static-sc.cloudapp.web.id": true, "static.sc.cloudapp.web.id": true, "files.cloudapp.web.id": true}

var sibiJenis = map[string]struct{ ep, label string }{
	"kurmer":  {"getPenggerakTextBooks", "Kurikulum Merdeka"},
	"k13":     {"getTextBooks", "Kurikulum 2013"},
	"nonteks": {"getNonTextBooks", "Buku bacaan (nonteks)"},
}

func init() {
	moduleMigrations = append(moduleMigrations, func(db *sql.DB) error {
		for _, c := range []string{"ebook_url", "cover_url", "sumber_id", "jenjang", "kelas"} {
			if err := addColumn(db, "books", c, "TEXT"); err != nil {
				return err
			}
		}
		_, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_books_sumber ON books(sumber_id) WHERE sumber_id IS NOT NULL`)
		return err
	})
	moduleRoutes = append(moduleRoutes, func(a *App, m *http.ServeMux) {
		p := func(roles string, fn func(*Ctx) error) http.HandlerFunc { return a.mod("perpus", roles, fn) }
		m.HandleFunc("GET /api/perpus/sibi", p("admin", apiSibiSearch))
		m.HandleFunc("POST /api/perpus/sibi/impor", p("admin", apiSibiImport))
		m.HandleFunc("GET /api/buku/{id}/baca", p("any", a.apiBookRead))
		m.HandleFunc("GET /api/buku/{id}/sampul", p("any", a.apiBookCover))
	})
}

// ---------------------------------------------------------------------
// Katalog SIBI (disimpan di memori 6 jam; ±1.700 judul)
// ---------------------------------------------------------------------
type sibiBook struct {
	ID       string `json:"id"`
	Judul    string `json:"judul"`
	Penulis  string `json:"penulis"`
	Penerbit string `json:"penerbit"`
	Tahun    string `json:"tahun"`
	ISBN     string `json:"isbn"`
	Jenjang  string `json:"jenjang"`
	Kelas    string `json:"kelas"`
	Mapel    string `json:"mapel"`
	Tipe     string `json:"tipe"` // buku siswa / buku guru / bacaan
	Jenis    string `json:"jenis"`
	PDF      string `json:"-"`
	Cover    string `json:"-"`
	Sinopsis string `json:"-"`
	Impor    bool   `json:"sudah_impor"`
}

var sibiCache struct {
	sync.Mutex
	at   map[string]time.Time
	fail map[string]time.Time
	data map[string][]sibiBook
}

// pengalihan (redirect) dari sumber hanya diikuti bila tujuannya tetap host cermin resmi (cegah SSRF ke jaringan dalam)
var httpLong = &http.Client{Timeout: 15 * time.Minute, CheckRedirect: func(req *http.Request, via []*http.Request) error {
	if len(via) >= 5 || sibiURL(req.URL.String()) == "" {
		return http.ErrUseLastResponse
	}
	return nil
}}

// katalog: pengalihan hanya ke host API yang sama
var httpCatalog = &http.Client{Timeout: 90 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
	if len(via) >= 5 || req.URL.Scheme != "https" || req.URL.Host != via[0].URL.Host {
		return http.ErrUseLastResponse
	}
	return nil
}}

// sibiURL memeriksa & menormalkan alamat file SIBI (hanya https ke host cermin resmi).
func sibiURL(raw string) string {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil || raw == "" {
		return ""
	}
	if u.Host == "static.buku.kemdikbud.go.id" || u.Host == "static.buku.kemendikdasmen.go.id" {
		u.Host = "static-sc.cloudapp.web.id"
	}
	if u.Scheme != "https" || !sibiHosts[u.Host] {
		return ""
	}
	return u.String()
}

func sibiJenjang(level, jenis string) string {
	l := strings.ToUpper(level)
	switch {
	case jenis == "nonteks" && l != "" && l != "PAUD":
		return "Bacaan jenjang " + level
	case strings.HasPrefix(l, "PAUD"):
		return "PAUD"
	case strings.HasPrefix(l, "SD"):
		return "SD"
	case strings.HasPrefix(l, "SMP"):
		return "SMP"
	case strings.HasPrefix(l, "SMK"):
		return "SMK"
	case strings.HasPrefix(l, "SMA"):
		return "SMA"
	}
	return level
}

func sibiLoad(jenis string) ([]sibiBook, error) {
	j, ok := sibiJenis[jenis]
	if !ok {
		return nil, bad("Jenis buku tidak dikenal.")
	}
	get := func() (d []sibiBook, fresh bool, failedRecently bool) {
		sibiCache.Lock()
		defer sibiCache.Unlock()
		if sibiCache.data == nil {
			sibiCache.at, sibiCache.data, sibiCache.fail = map[string]time.Time{}, map[string][]sibiBook{}, map[string]time.Time{}
		}
		d = sibiCache.data[jenis]
		return d, d != nil && time.Since(sibiCache.at[jenis]) < 6*time.Hour, time.Since(sibiCache.fail[jenis]) < time.Minute
	}
	if d, fresh, failed := get(); fresh || (failed && d != nil) {
		return d, nil
	} else if failed {
		return nil, apiErr{502, "Katalog Buku Kemendikdasmen tidak bisa dihubungi. Coba lagi sebentar lagi."}
	}
	// satu pengambilan per jenis; kunci cache TIDAK dipegang selama unduhan (bisa 90 detik)
	mu, _ := sibiFetchMu.LoadOrStore(jenis, &sync.Mutex{})
	mu.(*sync.Mutex).Lock()
	defer mu.(*sync.Mutex).Unlock()
	if d, fresh, failed := get(); fresh || (failed && d != nil) {
		return d, nil
	} else if failed { // yang menunggu antrean tidak mengulang unduhan 90 detik setelah pengambilan barusan gagal
		return nil, apiErr{502, "Katalog Buku Kemendikdasmen tidak bisa dihubungi. Coba lagi sebentar lagi."}
	}
	out, err := sibiFetch(j.ep, jenis)
	sibiCache.Lock()
	defer sibiCache.Unlock()
	if err != nil {
		sibiCache.fail[jenis] = time.Now()
		if old := sibiCache.data[jenis]; old != nil {
			return old, nil // sumber sedang bermasalah: pakai daftar lama
		}
		return nil, err
	}
	sibiCache.data[jenis], sibiCache.at[jenis] = out, time.Now()
	return out, nil
}

var sibiFetchMu sync.Map

func sibiFetch(ep, jenis string) ([]sibiBook, error) {
	resp, err := httpCatalog.Get(sibiAPI + ep + "?title=&limit=3000&offset=0")
	if err != nil {
		return nil, apiErr{502, "Katalog Buku Kemendikdasmen tidak bisa dihubungi. Pastikan server terhubung internet."}
	}
	defer resp.Body.Close()
	var raw struct {
		Results []map[string]any `json:"results"`
	}
	if resp.StatusCode != 200 || json.NewDecoder(io.LimitReader(resp.Body, 64<<20)).Decode(&raw) != nil {
		return nil, apiErr{502, "Katalog Buku Kemendikdasmen sedang bermasalah. Coba lagi nanti."}
	}
	s := func(m map[string]any, k string) string {
		v, _ := m[k].(string)
		return strings.TrimSpace(v)
	}
	out := make([]sibiBook, 0, len(raw.Results))
	for _, r := range raw.Results {
		pdf := sibiURL(s(r, "attachment"))
		if s(r, "type") != "pdf" || pdf == "" || s(r, "title") == "" || s(r, "id") == "" {
			continue
		}
		tipe := map[string]string{"buku_siswa": "Buku siswa", "buku_guru": "Buku guru"}[s(r, "book_type")]
		if jenis == "nonteks" {
			tipe = "Bacaan"
		}
		kelas := s(r, "class")
		if n, err := strconv.Atoi(kelas); err != nil || n < 1 || n > 12 {
			kelas = ""
		}
		// data sumber kadang keliru: panduan guru ditandai buku siswa
		if lt := strings.ToLower(s(r, "title")); jenis != "nonteks" && (strings.Contains(lt, "panduan guru") || strings.Contains(lt, "buku guru")) {
			tipe = "Buku guru"
		}
		tahun := s(r, "published_year")
		if len(tahun) > 4 {
			tahun = tahun[:4]
		}
		if y, err := strconv.Atoi(tahun); err != nil || y < 1950 {
			tahun = ""
		}
		out = append(out, sibiBook{ID: s(r, "id"), Judul: s(r, "title"), Penulis: s(r, "writer"), Penerbit: s(r, "publisher"), Tahun: tahun,
			ISBN: s(r, "isbn"), Jenjang: sibiJenjang(s(r, "level"), jenis), Kelas: kelas, Tipe: tipe, Jenis: jenis,
			Mapel: strings.ReplaceAll(s(r, "subject"), "_", " "), PDF: pdf, Cover: sibiURL(s(r, "image")), Sinopsis: s(r, "description")})
	}
	sort.SliceStable(out, func(i, k int) bool { return out[i].Judul < out[k].Judul })
	return out, nil
}

func apiSibiSearch(c *Ctx) error {
	q := c.R.URL.Query()
	jenis := q.Get("jenis")
	if jenis == "" {
		jenis = "kurmer"
	}
	all, err := sibiLoad(jenis)
	if err != nil {
		return err
	}
	words := strings.Fields(strings.ToLower(q.Get("q")))
	jenjang, kelas, tipe := q.Get("jenjang"), q.Get("kelas"), q.Get("tipe")
	have := map[string]bool{}
	if rows, err := c.A.reader().Query(`SELECT sumber_id FROM books WHERE sumber_id LIKE 'sibi:%'`); err == nil {
		for rows.Next() {
			var s string
			if rows.Scan(&s) == nil {
				have[s] = true
			}
		}
		rows.Close()
	}
	var hasil []sibiBook
	jenjangSet := map[string]bool{}
	for _, b := range all {
		jenjangSet[b.Jenjang] = true
		if (jenjang != "" && b.Jenjang != jenjang) || (kelas != "" && b.Kelas != kelas) || (tipe != "" && b.Tipe != tipe) {
			continue
		}
		hay := strings.ToLower(b.Judul + " " + b.Penulis + " " + b.Mapel + " " + b.ISBN)
		match := true
		for _, w := range words {
			if !strings.Contains(hay, w) {
				match = false
				break
			}
		}
		if match {
			b.Impor = have["sibi:"+b.ID]
			hasil = append(hasil, b)
		}
	}
	jj := make([]string, 0, len(jenjangSet))
	for k := range jenjangSet {
		if k != "" {
			jj = append(jj, k)
		}
	}
	sort.Strings(jj)
	total := len(hasil)
	if len(hasil) > 100 {
		hasil = hasil[:100]
	}
	return c.JSON(map[string]any{"total": total, "rows": hasil, "jenjang": jj})
}

// apiSibiImport memasukkan buku SIBI terpilih ke katalog sebagai e-book digital (stok fisik 0).
// Data diambil dari katalog server sendiri berdasarkan id, bukan dari isi kiriman browser.
func apiSibiImport(c *Ctx) error {
	var in struct {
		Jenis    string   `json:"jenis"`
		IDs      []string `json:"ids"`
		Kategori string   `json:"kategori"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	if len(in.IDs) == 0 {
		return bad("Pilih minimal satu buku.")
	}
	if len(in.IDs) > 200 {
		return bad("Maksimal 200 buku sekali impor.")
	}
	all, err := sibiLoad(in.Jenis)
	if err != nil {
		return err
	}
	byID := map[string]sibiBook{}
	for _, b := range all {
		byID[b.ID] = b
	}
	kat := strings.TrimSpace(in.Kategori)
	if kat == "" {
		kat = "E-book " + sibiJenis[in.Jenis].label
	}
	katp := c.canon("books", "kategori", &kat)
	tx, err := c.A.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	baru, lewat := 0, 0
	for _, id := range in.IDs {
		b, ok := byID[id]
		if !ok {
			lewat++
			continue
		}
		var isbn any
		if v := strings.ToUpper(strings.NewReplacer("-", "", " ", "").Replace(b.ISBN)); len(v) == 10 || len(v) == 13 {
			isbn = v
		}
		var dup int
		tx.QueryRow(`SELECT COUNT(*) FROM books WHERE sumber_id=? OR (? IS NOT NULL AND isbn=? AND ebook_url IS NOT NULL)`, "sibi:"+b.ID, isbn, isbn).Scan(&dup)
		if dup > 0 {
			lewat++
			continue
		}
		judul := b.Judul
		if lj := strings.ToLower(judul); b.Tipe != "" && b.Tipe != "Bacaan" && !strings.Contains(lj, "guru") && !strings.Contains(lj, "siswa") {
			judul += " (" + b.Tipe + ")"
		}
		_, err := tx.Exec(`INSERT INTO books (kode, isbn, judul, pengarang, penerbit, tahun, kategori, stok, sinopsis, ebook_url, cover_url, sumber_id, jenjang, kelas, created_at)
			VALUES (?,?,?,?,?,?,?,0,?,?,?,?,?,?,?)`,
			"SIBI-"+b.ID, isbn, judul, nullStr(b.Penulis), nullStr(b.Penerbit), nullStr(b.Tahun), str(katp), nullStr(b.Sinopsis), b.PDF, nullStr(b.Cover),
			"sibi:"+b.ID, nullStr(b.Jenjang), nullStr(b.Kelas), nowStr())
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				lewat++
				continue
			}
			return err
		}
		baru++
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.JSON(map[string]int{"baru": baru, "dilewati": lewat})
}

func nullStr(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// ---------------------------------------------------------------------
// Membaca: unduh sekali dari sumber, simpan, lalu sajikan dari server sekolah
// ---------------------------------------------------------------------
func (a *App) ebookCacheDir() string { return filepath.Join(filepath.Dir(a.dbPath), "ebook-cache") }

func cacheName(src, ext string) string {
	h := sha256.Sum256([]byte(src))
	return hex.EncodeToString(h[:16]) + ext
}

var fetchLocks sync.Map                 // nama file -> *sync.Mutex (satu unduhan per file walau dibuka banyak siswa bersamaan)
var fetchSlots = make(chan struct{}, 3) // maksimal 3 unduhan dari sumber sekaligus

const ebookMaxBytes = 500 << 20

// batas total salinan e-book: sama dengan batas folder upload (-upload-max-gb, bawaan 5 GB)
func ebookCacheMax() int64 { return uploadMaxBytes }

// cachedFile mengembalikan path file lokal untuk src; mengunduhnya dulu bila belum ada.
func (a *App) cachedFile(src, ext string, isPDF bool) (string, error) {
	dir := a.ebookCacheDir()
	p := filepath.Join(dir, cacheName(src, ext))
	if _, err := os.Stat(p); err == nil {
		now := time.Now()
		os.Chtimes(p, now, now) // tanda terakhir dibaca (untuk pembersihan file yang tidak dirujuk)
		return p, nil
	}
	mu, _ := fetchLocks.LoadOrStore(p, &sync.Mutex{})
	mu.(*sync.Mutex).Lock()
	defer func() { mu.(*sync.Mutex).Unlock(); fetchLocks.CompareAndDelete(p, mu) }()
	if _, err := os.Stat(p); err == nil { // sudah diunduh permintaan lain saat menunggu
		return p, nil
	}
	select {
	case fetchSlots <- struct{}{}:
		defer func() { <-fetchSlots }()
	case <-time.After(2 * time.Minute):
		return "", apiErr{503, "Server sedang mengunduh beberapa e-book. Coba buka lagi sebentar lagi."}
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	resp, err := httpLong.Get(src)
	if err != nil {
		return "", apiErr{502, "E-book belum tersimpan di server dan situs sumber tidak bisa dihubungi. Pastikan server terhubung internet."}
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", apiErr{502, fmt.Sprintf("Situs sumber menolak (%d). File e-book mungkin sudah dipindah.", resp.StatusCode)}
	}
	tmp, err := os.CreateTemp(dir, ".unduh-*")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name())
	n, err := io.Copy(tmp, io.LimitReader(resp.Body, ebookMaxBytes+1))
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return "", apiErr{502, "Unduhan e-book dari sumber terputus. Coba buka lagi."}
	}
	if n > ebookMaxBytes {
		return "", bad("File e-book terlalu besar (lebih dari 500 MB).")
	}
	head := make([]byte, 512)
	f, err := os.Open(tmp.Name())
	if err != nil {
		return "", err
	}
	k, _ := io.ReadFull(f, head)
	f.Close()
	ct := http.DetectContentType(head[:k])
	if (isPDF && !strings.HasPrefix(string(head[:k]), "%PDF")) || (!isPDF && !strings.HasPrefix(ct, "image/")) {
		return "", apiErr{502, "Situs sumber tidak mengirim file yang benar."}
	}
	if err := os.Rename(tmp.Name(), p); err != nil {
		return "", err
	}
	return p, nil
}

func (a *App) bookSource(c *Ctx, col string) (local, remote, judul string, err error) {
	var l, r, j sql.NullString
	q := `SELECT ebook, ebook_url, judul FROM books WHERE id=?`
	if col == "cover" {
		q = `SELECT cover, cover_url, judul FROM books WHERE id=?`
	}
	if e := a.reader().QueryRow(q, c.ID("id")).Scan(&l, &r, &j); e != nil {
		return "", "", "", notFound()
	}
	return l.String, r.String, j.String, nil
}

func (a *App) apiBookRead(c *Ctx) error {
	// file besar ke HP bersinyal lemah: jangan diputus batas waktu tulis server
	http.NewResponseController(c.W).SetWriteDeadline(time.Now().Add(60 * time.Minute))
	local, remote, judul, err := a.bookSource(c, "ebook")
	if err != nil {
		return err
	}
	var p string
	switch name := strings.TrimPrefix(local, "/files/"); {
	case local != "" && fileNameRe.MatchString(name):
		p = filepath.Join(a.uploadDir(), name)
	case sibiURL(remote) != "":
		src := sibiURL(remote)
		if _, e := os.Stat(filepath.Join(a.ebookCacheDir(), cacheName(src, ".pdf"))); e != nil {
			// buku yang belum tersimpan: batasi per pengguna agar tidak ada yang memicu unduhan seluruh katalog
			key := "ebook:" + strconv.FormatInt(c.U.ID, 10)
			if c.U.Role != "admin" && c.A.limit.blockedN(key, 15) {
				return apiErr{429, "Terlalu banyak e-book baru dibuka. Coba lagi 10 menit lagi, atau buka e-book yang sudah pernah dibaca."}
			}
			defer func() {
				if p != "" { // hanya unduhan yang berhasil dihitung (sumber mati / penyimpanan penuh tidak memakan jatah)
					c.A.limit.fail(key)
				}
			}()
			if _, size := a.ebookCacheSize(); size > ebookCacheMax() {
				return apiErr{507, "Penyimpanan salinan e-book penuh. Hubungi admin (hapus e-book yang tidak dipakai dari katalog)."}
			}
		}
		if p, err = a.cachedFile(src, ".pdf", true); err != nil {
			return err
		}
	default:
		return apiErr{404, "Buku ini tidak punya e-book."}
	}
	f, err := os.Open(p)
	if err != nil {
		return notFound()
	}
	defer f.Close()
	st, _ := f.Stat()
	h := c.W.Header()
	h.Set("Content-Type", "application/pdf")
	h.Set("Content-Disposition", "inline; filename*=UTF-8''"+url.PathEscape(safeFileName(judul)+".pdf"))
	h.Set("Cache-Control", "private, max-age=86400")
	h.Set("Content-Security-Policy", "default-src 'none'; object-src 'self'; plugin-types application/pdf")
	http.ServeContent(c.W, c.R, "", st.ModTime(), f)
	return nil
}

func (a *App) apiBookCover(c *Ctx) error {
	local, remote, _, err := a.bookSource(c, "cover")
	if err != nil {
		return err
	}
	if name := strings.TrimPrefix(local, "/files/"); local != "" && fileNameRe.MatchString(name) {
		c.W.Header().Set("Cache-Control", "private, max-age=604800")
		http.ServeFile(c.W, c.R, filepath.Join(a.uploadDir(), name))
		return nil
	}
	src := sibiURL(remote)
	if src == "" {
		return notFound()
	}
	p, err := a.cachedFile(src, ".img", false)
	if err != nil {
		return notFound() // sampul gagal diambil: tampilan memakai ikon buku
	}
	c.W.Header().Set("Cache-Control", "private, max-age=604800")
	c.W.Header().Set("Content-Security-Policy", "default-src 'none'")
	http.ServeFile(c.W, c.R, p)
	return nil
}

func safeFileName(s string) string {
	s = strings.Map(func(r rune) rune {
		if strings.ContainsRune(`/\:*?"<>|`, r) || r < 32 {
			return ' '
		}
		return r
	}, s)
	if len([]rune(s)) > 120 {
		s = string([]rune(s)[:120])
	}
	return strings.TrimSpace(s)
}

// pruneEbookCache menghapus salinan e-book/sampul yang bukunya sudah dihapus dari katalog (dijalankan pemeliharaan harian).
func (a *App) pruneEbookCache() (int, error) {
	dir := a.ebookCacheDir()
	ents, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	keep := map[string]bool{}
	rows, err := a.reader().Query(`SELECT ebook_url, cover_url FROM books WHERE ebook_url IS NOT NULL OR cover_url IS NOT NULL`)
	if err != nil {
		return 0, err
	}
	for rows.Next() {
		var e, cv sql.NullString
		if err := rows.Scan(&e, &cv); err != nil {
			rows.Close()
			return 0, err
		}
		if u := sibiURL(e.String); u != "" {
			keep[cacheName(u, ".pdf")] = true
		}
		if u := sibiURL(cv.String); u != "" {
			keep[cacheName(u, ".img")] = true
		}
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, err
	}
	rows.Close()
	n := 0
	for _, e := range ents {
		info, err := e.Info()
		if err != nil || e.IsDir() || keep[e.Name()] || time.Since(info.ModTime()) < 24*time.Hour {
			continue
		}
		if os.Remove(filepath.Join(dir, e.Name())) == nil {
			n++
		}
	}
	return n, nil
}

// ebookCacheSize: ukuran total salinan e-book (ditampilkan di info penyimpanan).
func (a *App) ebookCacheSize() (files int, bytes int64) {
	ents, _ := os.ReadDir(a.ebookCacheDir())
	for _, e := range ents {
		if info, err := e.Info(); err == nil && !e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			files++
			bytes += info.Size()
		}
	}
	return
}
