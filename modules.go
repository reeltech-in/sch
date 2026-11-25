package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

// ---------------------------------------------------------------------
// Pendaftaran skema & rute modul
// ---------------------------------------------------------------------
var moduleSchemas []string
var moduleMigrations []func(db *sql.DB) error
var moduleRoutes []func(a *App, m *http.ServeMux)

// Daftar modul yang bisa dinyalakan/dimatikan admin.
var moduleKeys = []string{"lms", "bukuinduk", "absensi", "kalender", "ujian", "kelulusan", "sarpras", "perpus", "ai", "ppdb", "surat", "bukutamu", "evoting", "kasir"}

func init() {
	moduleSchemas = append(moduleSchemas, `
CREATE TABLE IF NOT EXISTS kv (k TEXT PRIMARY KEY, v TEXT NOT NULL);
`)
	// index tambahan hasil audit (dijalankan setelah semua tabel modul dibuat)
	moduleMigrations = append(moduleMigrations, func(db *sql.DB) error {
		_, err := db.Exec(`
CREATE INDEX IF NOT EXISTS idx_sale_items_sale ON sale_items(sale_id);
CREATE INDEX IF NOT EXISTS idx_sessions_exp ON sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_exam_att_siswa ON exam_attempts(siswa_id);
CREATE INDEX IF NOT EXISTS idx_quiz_att_siswa ON quiz_attempts(siswa_id);
CREATE INDEX IF NOT EXISTS idx_participants_room ON exam_participants(room_id);
CREATE INDEX IF NOT EXISTS idx_loans_open ON loans(book_id) WHERE tgl_kembali IS NULL;`)
		return err
	})
	moduleMigrations = append(moduleMigrations, func(db *sql.DB) error {
		if err := addColumn(db, "users", "akses", "TEXT"); err != nil { // daftar modul yang boleh dikelola guru/staf (JSON)
			return err
		}
		return addColumn(db, "users", "ganti_pw", "INTEGER NOT NULL DEFAULT 0") // wajib ganti password saat login
	})
	moduleRoutes = append(moduleRoutes, func(a *App, m *http.ServeMux) {
		m.HandleFunc("PUT /api/modules", a.h("admin", apiPutModules))
		m.HandleFunc("GET /api/config/{key}", a.h("admin,guru", apiGetConfig))
		m.HandleFunc("PUT /api/config/{key}", a.h("admin,guru", apiPutConfig))
		m.HandleFunc("POST /api/upload", a.h("admin,guru", a.apiUpload))
		m.HandleFunc("GET /files/{name}", a.h("any", a.apiFile))
		m.HandleFunc("GET /api/logo", a.h("", a.apiLogo))
		m.HandleFunc("GET /api/qr", a.h("any", apiQR))
	})
}

// mod membungkus rute modul: menolak jika modul dimatikan, dan mengizinkan guru/staf
// yang diberi akses kelola modul ini memakai rute yang ditandai "admin".
func (a *App) mod(key string, roles string, fn func(*Ctx) error) http.HandlerFunc {
	check := func(c *Ctx) error {
		if !modulesEnabled(c.A.db)[key] {
			return apiErr{403, "Modul ini sedang dinonaktifkan admin."}
		}
		return fn(c)
	}
	if roles == "" {
		return a.h("", check)
	}
	return a.h("any", func(c *Ctx) error {
		ok := roles == "any" || strings.Contains(","+roles+",", ","+c.U.Role+",")
		if !ok && strings.Contains(","+roles+",", ",admin,") && c.U.Role == "guru" && c.U.Has(key) {
			ok = true
		}
		if !ok {
			return forbidden()
		}
		return check(c)
	})
}

func runModuleSchemas(db *sql.DB) error {
	for _, s := range moduleSchemas {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	for _, m := range moduleMigrations {
		if err := m(db); err != nil {
			return err
		}
	}
	return nil
}

// addColumn menambah kolom jika belum ada (SQLite tidak punya ADD COLUMN IF NOT EXISTS).
func addColumn(db *sql.DB, table, col, def string) error {
	rows, err := db.Query(`SELECT name FROM pragma_table_info(?)`, table)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var n string
		rows.Scan(&n)
		if n == col {
			return nil
		}
	}
	rows.Close()
	_, err = db.Exec(`ALTER TABLE ` + table + ` ADD COLUMN ` + col + ` ` + def)
	return err
}

// ---------------------------------------------------------------------
// Penyimpanan konfigurasi (kv)
// ---------------------------------------------------------------------
// kvCache: pengaturan (modul aktif, konfigurasi) dibaca hampir di setiap permintaan, jadi disimpan di
// memori. Semua penulisan kv WAJIB lewat kvSet (atau kvForget setelah SQL langsung) agar cache tetap benar.
var kvCache sync.Map

func kvGet(db *sql.DB, k string) string {
	if v, ok := kvCache.Load(k); ok {
		return v.(string)
	}
	var v string
	err := db.QueryRow(`SELECT v FROM kv WHERE k=?`, k).Scan(&v)
	if err == nil || err == sql.ErrNoRows { // galat lain (sibuk/IO) tidak disimpan, dicoba lagi berikutnya
		kvCache.Store(k, v)
	}
	return v
}
func kvSet(db *sql.DB, k, v string) error {
	_, err := db.Exec(`INSERT INTO kv (k, v) VALUES (?,?) ON CONFLICT (k) DO UPDATE SET v=excluded.v`, k, v)
	if err == nil {
		kvCache.Store(k, v)
	} else {
		kvCache.Delete(k)
	}
	return err
}
func kvForget(k string) { kvCache.Delete(k) }

func modulesEnabled(db *sql.DB) map[string]bool {
	m := map[string]bool{}
	json.Unmarshal([]byte(kvGet(db, "modul")), &m)
	out := map[string]bool{}
	for _, k := range moduleKeys {
		v, ok := m[k]
		out[k] = !ok || v
	}
	return out
}

func apiPutModules(c *Ctx) error {
	in := map[string]bool{}
	if err := c.Bind(&in); err != nil {
		return err
	}
	out := map[string]bool{}
	for _, k := range moduleKeys {
		v, ok := in[k]
		out[k] = !ok || v
	}
	b, _ := json.Marshal(out)
	if err := kvSet(c.A.db, "modul", string(b)); err != nil {
		return err
	}
	return c.OK()
}

// configHooks: dijalankan setelah konfigurasi disimpan (validasi lanjut / terapkan langsung).
var configHooks = map[string]func(db *sql.DB, cfg map[string]any) error{}

// Konfigurasi modul yang diizinkan beserta nilai bawaannya.
var configDefaults = map[string]string{
	"absen":     `{"jam_masuk":"07:15","jam_pulang":"13:00","selfie":false,"lat":null,"lng":null,"radius_m":150}`,
	"kelulusan": `{"buka":null,"judul":"Pengumuman Kelulusan","pesan":"Selamat atas kelulusan Anda. Tetap jaga nama baik sekolah.","pesan_tidak":"Silakan hubungi wali kelas atau bagian kurikulum."}`,
}

func getConfig(db *sql.DB, key string) map[string]any {
	m := map[string]any{}
	json.Unmarshal([]byte(configDefaults[key]), &m)
	saved := map[string]any{}
	json.Unmarshal([]byte(kvGet(db, "cfg_"+key)), &saved)
	for k, v := range saved {
		if _, ok := m[k]; ok {
			m[k] = v
		}
	}
	return m
}

// configModule: modul pemilik tiap konfigurasi (untuk izin staf).
var configModule = map[string]string{"absen": "absensi", "kelulusan": "kelulusan", "perpus": "perpus", "ppdb": "ppdb", "surat": "surat", "ai": "ai"}

func (c *Ctx) canConfig(key string) error {
	// ai & wa punya endpoint sendiri (token rahasia, validasi gateway), jadi tidak lewat /api/config
	if _, ok := configDefaults[key]; !ok || key == "ai" || key == "wa" {
		return notFound()
	}
	if !c.U.Has(configModule[key]) {
		return forbidden()
	}
	return nil
}

func apiGetConfig(c *Ctx) error {
	key := c.R.PathValue("key")
	if err := c.canConfig(key); err != nil {
		return err
	}
	return c.JSON(getConfig(c.A.db, key))
}

func apiPutConfig(c *Ctx) error {
	key := c.R.PathValue("key")
	if err := c.canConfig(key); err != nil {
		return err
	}
	in := map[string]any{}
	if err := c.Bind(&in); err != nil {
		return err
	}
	cur := getConfig(c.A.db, key)
	for k := range cur {
		if v, ok := in[k]; ok {
			cur[k] = v
		}
	}
	if h := configHooks[key]; h != nil {
		if err := h(c.A.db, cur); err != nil {
			return err
		}
	}
	b, _ := json.Marshal(cur)
	if err := kvSet(c.A.db, "cfg_"+key, string(b)); err != nil {
		return err
	}
	return c.OK()
}

// ---------------------------------------------------------------------
// Tanda tangan kartu QR siswa (mencegah QR palsu)
// ---------------------------------------------------------------------
// secret: kunci HMAC kartu, dibaca sekali lalu disimpan di memori.
// Dibuat dengan INSERT OR IGNORE lalu dibaca ulang, sehingga dua permintaan bersamaan
// tidak menghasilkan dua rahasia berbeda (kartu yang sudah dicetak tetap berlaku).
func (a *App) secret() []byte {
	a.secretMu.Lock()
	defer a.secretMu.Unlock()
	if a.secretV != nil {
		return a.secretV
	}
	b := make([]byte, 32)
	rand.Read(b)
	a.db.Exec(`INSERT OR IGNORE INTO kv (k, v) VALUES ('secret', ?)`, hex.EncodeToString(b))
	kvForget("secret")
	if s := kvGet(a.db, "secret"); s != "" {
		a.secretV = []byte(s)
	}
	return a.secretV
}

func (a *App) cardToken(userID int64) string {
	key := a.secret()
	if len(key) == 0 {
		return "" // kunci belum terbaca: jangan pernah membuat/menerima token dengan kunci kosong
	}
	id := strconv.FormatInt(userID, 10)
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte("kartu:" + id))
	return "S1." + id + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))[:12]
}

// cardUser mengembalikan id pengguna dari isi QR kartu, atau 0 jika palsu.
func (a *App) cardUser(token string) int64 {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 || parts[0] != "S1" {
		return 0
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	want := ""
	if err == nil {
		want = a.cardToken(id)
	}
	if want == "" || !hmac.Equal([]byte(want), []byte(strings.TrimSpace(token))) {
		return 0
	}
	return id
}

// ---------------------------------------------------------------------
// QR code (PNG)
// ---------------------------------------------------------------------
func apiQR(c *Ctx) error {
	d := c.R.URL.Query().Get("d")
	if d == "" || len(d) > 300 {
		return bad("Isi QR tidak valid.")
	}
	size, _ := strconv.Atoi(c.R.URL.Query().Get("s"))
	if size < 64 || size > 1024 {
		size = 256
	}
	png, err := qrcode.Encode(d, qrcode.Medium, size)
	if err != nil {
		return err
	}
	c.W.Header().Set("Content-Type", "image/png")
	c.W.Header().Set("Cache-Control", "private, max-age=86400")
	_, err = c.W.Write(png)
	return err
}

// ---------------------------------------------------------------------
// Upload & unduh file
// ---------------------------------------------------------------------
func (a *App) uploadDir() string { return filepath.Join(filepath.Dir(a.dbPath), "uploads") }

var allowedUploads = map[string]string{
	"image/jpeg": ".jpg", "image/png": ".png", "image/gif": ".gif", "image/webp": ".webp", "application/pdf": ".pdf",
}

// uploadMaxBytes: batas total folder upload (atur dengan -upload-max-gb).
var uploadMaxBytes int64 = 5 << 30
var uploadSize struct {
	sync.Mutex
	bytes int64
	at    time.Time
}

func (a *App) uploadUsed() int64 {
	uploadSize.Lock()
	defer uploadSize.Unlock()
	if time.Since(uploadSize.at) > time.Minute {
		var n int64
		filepath.WalkDir(a.uploadDir(), func(_ string, d os.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				if fi, e := d.Info(); e == nil {
					n += fi.Size()
				}
			}
			return nil
		})
		uploadSize.bytes, uploadSize.at = n, time.Now()
	}
	return uploadSize.bytes
}

// saveUpload menyimpan data byte setelah memeriksa jenis aslinya. Mengembalikan alamat /files/...
func (a *App) saveUpload(data []byte, onlyImage bool) (string, error) {
	if a.uploadUsed()+int64(len(data)) > uploadMaxBytes {
		return "", apiErr{507, "Penyimpanan file sekolah penuh. Hubungi admin."}
	}
	ct := http.DetectContentType(data)
	if i := strings.Index(ct, ";"); i > 0 {
		ct = ct[:i]
	}
	ext, ok := allowedUploads[ct]
	if !ok || (onlyImage && ct == "application/pdf") {
		return "", bad("Jenis file tidak didukung. Gunakan JPG, PNG, WEBP, GIF, atau PDF.")
	}
	if err := os.MkdirAll(a.uploadDir(), 0o700); err != nil {
		return "", err
	}
	b := make([]byte, 12)
	rand.Read(b)
	name := hex.EncodeToString(b) + ext
	if err := os.WriteFile(filepath.Join(a.uploadDir(), name), data, 0o600); err != nil {
		return "", err
	}
	uploadSize.Lock()
	uploadSize.bytes += int64(len(data))
	uploadSize.Unlock()
	return "/files/" + name, nil
}

// Unggahan dari halaman publik (buku tamu, PPDB) dibatasi tersendiri: jatah harian bersama, dan tidak boleh
// memakai cadangan ruang terakhir, agar pengunjung tanpa login tidak bisa memenuhi penyimpanan sekolah
// (selfie absensi, gambar soal, logo tetap bisa diunggah).
const publikHarianMaks = 300 << 20 // 300 MB per hari untuk semua unggahan publik

// publikPerPerangkat: jatah harian satu perangkat (IP, IPv6 per /64) dari kolam publik di atas, agar satu
// pengirim tidak bisa menghabiskan jatah semua pendaftar lain. Cukup untuk beberapa pendaftaran lengkap.
const publikPerPerangkat = 60 << 20

var publikUpl struct {
	sync.Mutex
	hari  string
	bytes int64
	per   map[string]int64
}

// savePublicUpload: who = kunci perangkat pengirim (limitKey(clientIP)).
func (a *App) savePublicUpload(data []byte, onlyImage bool, who string) (string, error) {
	cadangan := uploadMaxBytes / 5 // 20% terakhir hanya untuk pengguna yang login
	if a.uploadUsed()+int64(len(data)) > uploadMaxBytes-cadangan {
		return "", apiErr{507, "Penyimpanan unggahan sedang penuh. Silakan hubungi sekolah."}
	}
	hari := time.Now().Format("2006-01-02")
	publikUpl.Lock()
	if publikUpl.hari != hari || publikUpl.per == nil {
		publikUpl.hari, publikUpl.bytes, publikUpl.per = hari, 0, map[string]int64{}
	}
	if publikUpl.per[who]+int64(len(data)) > publikPerPerangkat {
		publikUpl.Unlock()
		return "", apiErr{429, "Batas unggahan dari perangkat ini hari ini sudah tercapai. Silakan coba lagi besok atau hubungi sekolah."}
	}
	if publikUpl.bytes+int64(len(data)) > publikHarianMaks {
		publikUpl.Unlock()
		return "", apiErr{429, "Batas unggahan hari ini sudah tercapai. Silakan coba lagi besok atau hubungi sekolah."}
	}
	publikUpl.bytes += int64(len(data))
	publikUpl.per[who] += int64(len(data))
	publikUpl.Unlock()
	url, err := a.saveUpload(data, onlyImage)
	if err != nil {
		publikUpl.Lock()
		if publikUpl.hari == hari {
			publikUpl.bytes -= int64(len(data))
			publikUpl.per[who] -= int64(len(data))
		}
		publikUpl.Unlock()
	}
	return url, err
}

// removeUploads menghapus berkas yang sudah tersimpan bila penyimpanan datanya gagal (tidak meninggalkan file yatim).
func (a *App) removeUploads(urls []string) {
	for _, u := range urls {
		name := strings.TrimPrefix(u, "/files/")
		if name == "" || strings.ContainsAny(name, "/\\") {
			continue
		}
		if st, err := os.Stat(filepath.Join(a.uploadDir(), name)); err == nil && os.Remove(filepath.Join(a.uploadDir(), name)) == nil {
			uploadSize.Lock()
			uploadSize.bytes -= st.Size()
			uploadSize.Unlock()
		}
	}
}

func (a *App) apiUpload(c *Ctx) error {
	c.longBody()
	c.R.Body = http.MaxBytesReader(c.W, c.R.Body, 32<<20)
	f, _, err := c.R.FormFile("file")
	if err != nil {
		return bad("File terlalu besar (maks 30 MB) atau tidak terbaca.")
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return bad("File tidak terbaca.")
	}
	ct := http.DetectContentType(data)
	if strings.HasPrefix(ct, "image/") && len(data) > 5<<20 {
		return bad("Gambar maksimal 5 MB.")
	}
	url, err := c.A.saveUpload(data, false)
	if err != nil {
		return err
	}
	return c.JSON(map[string]string{"url": url})
}

var fileNameRe = regexp.MustCompile(`^[a-f0-9]{24}\.(jpg|png|gif|webp|pdf)$`)

func (a *App) apiFile(c *Ctx) error {
	name := c.R.PathValue("name")
	if !fileNameRe.MatchString(name) {
		return notFound()
	}
	// berkas khusus modul (mis. dokumen pendaftar PPDB): hanya admin & guru yang diberi tugas modul itu
	var modul string
	if c.A.reader().QueryRow(`SELECT modul FROM upload_akses WHERE nama=?`, name).Scan(&modul) == nil {
		if !(c.U.Role == "admin" || (c.U.Role == "guru" && c.U.Has(modul))) {
			return notFound()
		}
	}
	p := filepath.Join(a.uploadDir(), name)
	if _, err := os.Stat(p); err != nil {
		return notFound()
	}
	// lampiran PDF / e-book unggahan bisa puluhan MB: di WiFi sekolah yang padat unduhannya lebih lama dari batas
	// tulis server (60 detik) dan terputus di tengah. Perpanjang khusus untuk respons ini.
	http.NewResponseController(c.W).SetWriteDeadline(time.Now().Add(30 * time.Minute))
	c.W.Header().Set("Cache-Control", "private, max-age=604800")
	c.W.Header().Set("Content-Security-Policy", "default-src 'none'; img-src 'self'; style-src 'unsafe-inline'; object-src 'self'; plugin-types application/pdf")
	http.ServeFile(c.W, c.R, p)
	return nil
}

// apiLogo menyajikan logo sekolah tanpa login (dipakai di halaman login & publik).
func (a *App) apiLogo(c *Ctx) error {
	var logo sql.NullString
	a.db.QueryRow(`SELECT logo_url FROM settings WHERE id=1`).Scan(&logo)
	name := strings.TrimPrefix(logo.String, "/files/")
	if !logo.Valid || !fileNameRe.MatchString(name) || strings.HasSuffix(name, ".pdf") {
		return notFound()
	}
	c.W.Header().Set("Cache-Control", "public, max-age=3600")
	http.ServeFile(c.W, c.R, filepath.Join(a.uploadDir(), name))
	return nil
}

// decodeDataURL mengubah "data:image/jpeg;base64,..." menjadi byte.
func decodeDataURL(s string) ([]byte, error) {
	i := strings.Index(s, ",")
	if !strings.HasPrefix(s, "data:") || i < 0 {
		return nil, fmt.Errorf("format foto tidak valid")
	}
	return base64.StdEncoding.DecodeString(s[i+1:])
}
