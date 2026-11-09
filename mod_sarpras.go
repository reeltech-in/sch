package main

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Modul: Sarana Prasarana (inventaris KIB A–F + label QR) dan Perpustakaan (katalog, e-book, sirkulasi).

func init() {
	configDefaults["perpus"] = `{"lama_pinjam":7,"denda_per_hari":500,"maks_pinjam":3}`
	moduleSchemas = append(moduleSchemas, `
CREATE TABLE IF NOT EXISTS assets (
  id          INTEGER PRIMARY KEY,
  kode        TEXT NOT NULL UNIQUE COLLATE NOCASE,
  nama        TEXT NOT NULL,
  kib         TEXT NOT NULL CHECK (kib IN ('A','B','C','D','E','F')),
  kategori    TEXT,
  merk        TEXT,
  tahun       TEXT,
  sumber_dana TEXT,
  harga       REAL NOT NULL DEFAULT 0,
  jumlah      INTEGER NOT NULL DEFAULT 1,
  satuan      TEXT,
  kondisi     TEXT NOT NULL DEFAULT 'baik' CHECK (kondisi IN ('baik','rusak_ringan','rusak_berat')),
  lokasi      TEXT,
  foto        TEXT,
  detail      TEXT NOT NULL DEFAULT '{}',
  keterangan  TEXT,
  created_at  TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_assets_kib ON assets(kib);

CREATE TABLE IF NOT EXISTS books (
  id         INTEGER PRIMARY KEY,
  kode       TEXT NOT NULL UNIQUE COLLATE NOCASE,
  isbn       TEXT,
  judul      TEXT NOT NULL,
  pengarang  TEXT,
  penerbit   TEXT,
  tahun      TEXT,
  kategori   TEXT,
  rak        TEXT,
  stok       INTEGER NOT NULL DEFAULT 1,
  cover      TEXT,
  ebook      TEXT,
  sinopsis   TEXT,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS loans (
  id          INTEGER PRIMARY KEY,
  book_id     INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
  user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  tgl_pinjam  TEXT NOT NULL,
  jatuh_tempo TEXT NOT NULL,
  tgl_kembali TEXT,
  denda       REAL NOT NULL DEFAULT 0,
  petugas_id  INTEGER REFERENCES users(id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS idx_loans_user ON loans(user_id);
CREATE INDEX IF NOT EXISTS idx_loans_book ON loans(book_id);
`)
	moduleRoutes = append(moduleRoutes, func(a *App, m *http.ServeMux) {
		s := func(roles string, fn func(*Ctx) error) http.HandlerFunc { return a.mod("sarpras", roles, fn) }
		m.HandleFunc("GET /api/aset", s("admin", apiAssets))
		m.HandleFunc("POST /api/aset", s("admin", apiAssetSave))
		m.HandleFunc("PUT /api/aset/{id}", s("admin", apiAssetSave))
		m.HandleFunc("DELETE /api/aset/{id}", s("admin", apiDelete("assets")))

		p := func(roles string, fn func(*Ctx) error) http.HandlerFunc { return a.mod("perpus", roles, fn) }
		m.HandleFunc("GET /api/buku", p("any", apiBooks))
		m.HandleFunc("POST /api/buku", p("admin", apiBookSave))
		m.HandleFunc("PUT /api/buku/{id}", p("admin", apiBookSave))
		m.HandleFunc("DELETE /api/buku/{id}", p("admin", apiDelete("books")))
		m.HandleFunc("GET /api/perpus/anggota", p("admin", a.apiMember))
		m.HandleFunc("POST /api/perpus/pinjam", p("admin", a.apiBorrow))
		m.HandleFunc("POST /api/perpus/kembali/{id}", p("admin", apiReturn))
		m.HandleFunc("GET /api/perpus/pinjaman", p("admin", apiLoans))
		m.HandleFunc("GET /api/perpus/me", p("any", apiMyLoans))
	})
}

// ---------------------------------------------------------------------
// Sarpras
// ---------------------------------------------------------------------
func apiAssets(c *Ctx) error {
	where := `1=1`
	args := []any{}
	if k := c.R.URL.Query().Get("kib"); k != "" {
		where += ` AND kib=?`
		args = append(args, k)
	}
	if s := strings.TrimSpace(c.R.URL.Query().Get("q")); s != "" {
		where += ` AND (kode LIKE ? OR nama LIKE ? OR lokasi LIKE ? OR merk LIKE ?)`
		like := "%" + s + "%"
		args = append(args, like, like, like, like)
	}
	limit, offset := pageParams(c, 500)
	var total int
	var nilai float64
	c.A.db.QueryRow(`SELECT COUNT(*), COALESCE(SUM(harga*jumlah),0) FROM assets WHERE `+where, args...).Scan(&total, &nilai)
	c.W.Header().Set("X-Total-Count", strconv.Itoa(total))
	c.W.Header().Set("X-Total-Nilai", strconv.FormatFloat(nilai, 'f', 0, 64))
	rows, err := c.list(`SELECT * FROM assets WHERE `+where+` ORDER BY kib, kode LIMIT ? OFFSET ?`, append(args, limit, offset)...)
	if err != nil {
		return err
	}
	for _, r := range rows {
		if s, ok := r["detail"].(string); ok && json.Valid([]byte(s)) {
			r["detail"] = json.RawMessage(s)
		}
	}
	return c.JSON(rows)
}

func apiAssetSave(c *Ctx) error {
	var in struct {
		Kode       string            `json:"kode"`
		Nama       string            `json:"nama"`
		KIB        string            `json:"kib"`
		Kategori   *string           `json:"kategori"`
		Merk       *string           `json:"merk"`
		Tahun      *string           `json:"tahun"`
		SumberDana *string           `json:"sumber_dana"`
		Harga      float64           `json:"harga"`
		Jumlah     int               `json:"jumlah"`
		Satuan     *string           `json:"satuan"`
		Kondisi    string            `json:"kondisi"`
		Lokasi     *string           `json:"lokasi"`
		Foto       *string           `json:"foto"`
		Detail     map[string]string `json:"detail"`
		Keterangan *string           `json:"keterangan"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	kode, err := required(in.Kode, "Kode barang")
	if err != nil {
		return err
	}
	nama, err := required(in.Nama, "Nama barang")
	if err != nil {
		return err
	}
	if !strings.Contains("ABCDEF", in.KIB) || len(in.KIB) != 1 {
		return bad("Pilih jenis KIB A sampai F.")
	}
	switch in.Kondisi {
	case "baik", "rusak_ringan", "rusak_berat":
	default:
		in.Kondisi = "baik"
	}
	if in.Jumlah < 1 {
		in.Jumlah = 1
	}
	if in.Harga < 0 {
		return bad("Harga tidak boleh negatif.")
	}
	foto := str(in.Foto)
	if foto != nil && !fileNameRe.MatchString(strings.TrimPrefix(foto.(string), "/files/")) {
		return bad("Foto tidak valid.")
	}
	det := map[string]string{}
	for k, v := range in.Detail {
		if len(k) <= 40 && strings.TrimSpace(v) != "" {
			det[k] = strings.TrimSpace(v)
		}
	}
	db, _ := json.Marshal(det)
	in.Kategori, in.SumberDana = c.canon("assets", "kategori", in.Kategori), c.canon("assets", "sumber_dana", in.SumberDana)
	in.Satuan, in.Lokasi = c.canon("assets", "satuan", in.Satuan), c.canon("assets", "lokasi", in.Lokasi)
	args := []any{kode, nama, in.KIB, str(in.Kategori), str(in.Merk), str(in.Tahun), str(in.SumberDana), in.Harga, in.Jumlah, str(in.Satuan), in.Kondisi, str(in.Lokasi), foto, string(db), str(in.Keterangan)}
	if id := c.ID("id"); id > 0 {
		_, err = c.A.db.Exec(`UPDATE assets SET kode=?, nama=?, kib=?, kategori=?, merk=?, tahun=?, sumber_dana=?, harga=?, jumlah=?, satuan=?, kondisi=?, lokasi=?, foto=?, detail=?, keterangan=? WHERE id=?`, append(args, id)...)
	} else {
		_, err = c.A.db.Exec(`INSERT INTO assets (kode, nama, kib, kategori, merk, tahun, sumber_dana, harga, jumlah, satuan, kondisi, lokasi, foto, detail, keterangan, created_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, append(args, nowStr())...)
	}
	if err != nil {
		return err
	}
	return c.OK()
}

// ---------------------------------------------------------------------
// Perpustakaan
// ---------------------------------------------------------------------
func apiBooks(c *Ctx) error {
	// pencarian & halaman di server: katalog bisa berisi puluhan ribu judul
	where, args := `1=1`, []any{}
	if s := strings.TrimSpace(c.R.URL.Query().Get("q")); s != "" {
		where = `(b.judul LIKE ? OR b.pengarang LIKE ? OR b.kode LIKE ? OR b.isbn LIKE ? OR b.kategori LIKE ?)`
		like := "%" + s + "%"
		args = append(args, like, like, like, like, like)
	}
	if c.R.URL.Query().Get("ebook") == "1" {
		where += ` AND (b.ebook IS NOT NULL OR b.ebook_url IS NOT NULL)`
	}
	if j := c.R.URL.Query().Get("jenjang"); j != "" {
		where += ` AND b.jenjang=?`
		args = append(args, j)
	}
	limit, offset := pageParams(c, 300)
	var total int
	c.A.db.QueryRow(`SELECT COUNT(*) FROM books b WHERE `+where, args...).Scan(&total)
	rows, err := c.list(`SELECT b.*, b.stok - (SELECT COUNT(*) FROM loans l WHERE l.book_id=b.id AND l.tgl_kembali IS NULL) AS tersedia
		FROM books b WHERE `+where+` ORDER BY b.judul LIMIT ? OFFSET ?`, append(args, limit, offset)...)
	if err != nil {
		return err
	}
	c.W.Header().Set("X-Total-Count", strconv.Itoa(total))
	return c.JSON(rows)
}

func apiBookSave(c *Ctx) error {
	var in struct {
		Kode      string  `json:"kode"`
		ISBN      *string `json:"isbn"`
		Judul     string  `json:"judul"`
		Pengarang *string `json:"pengarang"`
		Penerbit  *string `json:"penerbit"`
		Tahun     *string `json:"tahun"`
		Kategori  *string `json:"kategori"`
		Rak       *string `json:"rak"`
		Stok      int     `json:"stok"`
		Cover     *string `json:"cover"`
		Ebook     *string `json:"ebook"`
		Sinopsis  *string `json:"sinopsis"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	kode, err := required(in.Kode, "Kode buku")
	if err != nil {
		return err
	}
	judul, err := required(in.Judul, "Judul buku")
	if err != nil {
		return err
	}
	if in.Stok < 0 {
		return bad("Stok tidak boleh negatif.")
	}
	for _, p := range []*string{in.Cover, in.Ebook} {
		if v := str(p); v != nil && !fileNameRe.MatchString(strings.TrimPrefix(v.(string), "/files/")) {
			return bad("File cover/e-book tidak valid.")
		}
	}
	if v := str(in.ISBN); v != nil {
		isbn := strings.ToUpper(strings.NewReplacer("-", "", " ", "").Replace(v.(string)))
		if len(isbn) != 10 && len(isbn) != 13 {
			return bad("ISBN harus 10 atau 13 digit.")
		}
		in.ISBN = &isbn
		if j, ok := c.exists(`SELECT judul FROM books WHERE isbn=? AND id<>?`, isbn, c.ID("id")); ok {
			return bad("ISBN ini sudah dipakai buku " + j + ". Tambahkan jumlah eksemplarnya saja.")
		}
	}
	in.Kategori, in.Rak, in.Penerbit = c.canon("books", "kategori", in.Kategori), c.canon("books", "rak", in.Rak), c.canon("books", "penerbit", in.Penerbit)
	args := []any{kode, str(in.ISBN), judul, str(in.Pengarang), str(in.Penerbit), str(in.Tahun), str(in.Kategori), str(in.Rak), in.Stok, str(in.Cover), str(in.Ebook), str(in.Sinopsis)}
	if id := c.ID("id"); id > 0 {
		_, err = c.A.db.Exec(`UPDATE books SET kode=?, isbn=?, judul=?, pengarang=?, penerbit=?, tahun=?, kategori=?, rak=?, stok=?, cover=?, ebook=?, sinopsis=? WHERE id=?`, append(args, id)...)
	} else {
		_, err = c.A.db.Exec(`INSERT INTO books (kode, isbn, judul, pengarang, penerbit, tahun, kategori, rak, stok, cover, ebook, sinopsis, created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`, append(args, nowStr())...)
	}
	if err != nil {
		return err
	}
	return c.OK()
}

func perpusCfg(db *sql.DB) (lama int, denda float64, maks int) {
	cfg := getConfig(db, "perpus")
	l, _ := cfg["lama_pinjam"].(float64)
	d, _ := cfg["denda_per_hari"].(float64)
	m, _ := cfg["maks_pinjam"].(float64)
	if l < 1 {
		l = 7
	}
	if m < 1 {
		m = 3
	}
	return int(l), d, int(m)
}

// findMember mencari anggota dari isi QR kartu atau username.
func (a *App) findMember(q string) (int64, error) {
	q = strings.TrimSpace(q)
	if id := a.cardUser(q); id > 0 {
		return id, nil
	}
	var id int64
	if err := a.db.QueryRow(`SELECT id FROM users WHERE username=? COLLATE NOCASE AND role IN ('siswa','guru') AND status='aktif'`, q).Scan(&id); err != nil {
		return 0, bad("Anggota tidak ditemukan. Scan kartu siswa atau ketik username/NIS.")
	}
	return id, nil
}

func (a *App) loansOf(c *Ctx, userID int64, activeOnly bool) ([]map[string]any, error) {
	q := `SELECT l.id, l.tgl_pinjam, l.jatuh_tempo, l.tgl_kembali, l.denda, b.kode, b.judul, b.pengarang FROM loans l JOIN books b ON b.id=l.book_id WHERE l.user_id=?`
	if activeOnly {
		q += ` AND l.tgl_kembali IS NULL`
	}
	return c.list(q+` ORDER BY l.tgl_pinjam DESC`, userID)
}

func (a *App) apiMember(c *Ctx) error {
	id, err := a.findMember(c.R.URL.Query().Get("q"))
	if err != nil {
		return err
	}
	u, err := c.list(`SELECT u.id, u.username, u.nama_lengkap, u.role, k.nama AS kelas_nama, sp.foto FROM users u LEFT JOIN classes k ON k.id=u.kelas_id
		LEFT JOIN student_profiles sp ON sp.siswa_id=u.id WHERE u.id=?`, id)
	if err != nil || len(u) == 0 {
		return notFound()
	}
	loans, err := a.loansOf(c, id, true)
	if err != nil {
		return err
	}
	return c.JSON(map[string]any{"anggota": u[0], "pinjaman": loans})
}

func (a *App) apiBorrow(c *Ctx) error {
	var in struct {
		UserID int64  `json:"user_id"`
		Buku   string `json:"buku"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	lama, _, maks := perpusCfg(a.db)
	var bookID int64
	var judul string
	var stok int
	if err := a.db.QueryRow(`SELECT id, judul, stok FROM books WHERE kode=? COLLATE NOCASE OR (isbn IS NOT NULL AND isbn=?)`, strings.TrimSpace(in.Buku), strings.TrimSpace(in.Buku)).Scan(&bookID, &judul, &stok); err != nil {
		return bad("Buku dengan kode itu tidak ditemukan.")
	}
	var dipinjam, aktif, sama int
	a.db.QueryRow(`SELECT COUNT(*) FROM loans WHERE book_id=? AND tgl_kembali IS NULL`, bookID).Scan(&dipinjam)
	a.db.QueryRow(`SELECT COUNT(*) FROM loans WHERE user_id=? AND tgl_kembali IS NULL`, in.UserID).Scan(&aktif)
	a.db.QueryRow(`SELECT COUNT(*) FROM loans WHERE user_id=? AND book_id=? AND tgl_kembali IS NULL`, in.UserID, bookID).Scan(&sama)
	switch {
	case dipinjam >= stok:
		return bad("Stok \"" + judul + "\" sedang habis dipinjam.")
	case sama > 0:
		return bad("Anggota ini masih meminjam buku yang sama.")
	case aktif >= maks:
		return bad("Anggota sudah meminjam " + itoa(aktif) + " buku (batas " + itoa(maks) + ").")
	}
	var ok int
	a.db.QueryRow(`SELECT COUNT(*) FROM users WHERE id=? AND status='aktif' AND role IN ('siswa','guru')`, in.UserID).Scan(&ok)
	if ok == 0 {
		return bad("Anggota tidak valid.")
	}
	due := time.Now().AddDate(0, 0, lama).Format("2006-01-02")
	if _, err := a.db.Exec(`INSERT INTO loans (book_id, user_id, tgl_pinjam, jatuh_tempo, petugas_id) VALUES (?,?,?,?,?)`, bookID, in.UserID, localDate(), due, c.U.ID); err != nil {
		return err
	}
	return c.JSON(map[string]any{"judul": judul, "jatuh_tempo": due})
}

func apiReturn(c *Ctx) error {
	id := c.ID("id")
	var due string
	var back sql.NullString
	if err := c.A.db.QueryRow(`SELECT jatuh_tempo, tgl_kembali FROM loans WHERE id=?`, id).Scan(&due, &back); err != nil {
		return notFound()
	}
	if back.Valid {
		return bad("Buku ini sudah dikembalikan.")
	}
	_, perHari, _ := perpusCfg(c.A.db)
	d, _ := time.ParseInLocation("2006-01-02", due, time.Local)
	today, _ := time.ParseInLocation("2006-01-02", localDate(), time.Local)
	late := int(math.Max(0, today.Sub(d).Hours()/24))
	denda := float64(late) * perHari
	res, err := c.A.db.Exec(`UPDATE loans SET tgl_kembali=?, denda=? WHERE id=? AND tgl_kembali IS NULL`, localDate(), denda, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return bad("Buku ini sudah dikembalikan.")
	}
	return c.JSON(map[string]any{"terlambat_hari": late, "denda": denda})
}

func apiLoans(c *Ctx) error {
	q := `SELECT l.id, l.tgl_pinjam, l.jatuh_tempo, l.tgl_kembali, l.denda, b.kode, b.judul, u.nama_lengkap, u.username, k.nama AS kelas_nama
		FROM loans l JOIN books b ON b.id=l.book_id JOIN users u ON u.id=l.user_id LEFT JOIN classes k ON k.id=u.kelas_id`
	switch c.R.URL.Query().Get("status") {
	case "semua":
	case "terlambat":
		q += ` WHERE l.tgl_kembali IS NULL AND l.jatuh_tempo < '` + localDate() + `'`
	default:
		q += ` WHERE l.tgl_kembali IS NULL`
	}
	rows, err := c.list(q + ` ORDER BY l.jatuh_tempo LIMIT 500`)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

func apiMyLoans(c *Ctx) error {
	rows, err := c.list(`SELECT l.id, l.tgl_pinjam, l.jatuh_tempo, l.tgl_kembali, l.denda, b.kode, b.judul, b.pengarang
		FROM loans l JOIN books b ON b.id=l.book_id WHERE l.user_id=? ORDER BY l.tgl_pinjam DESC LIMIT 100`, c.U.ID)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}
