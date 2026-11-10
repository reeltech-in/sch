package main

import (
	"database/sql"
	"log"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata" // data zona waktu tertanam (Windows/sistem minimal tidak selalu punya)

	"golang.org/x/crypto/bcrypt"
)

// =====================================================================
// Zona waktu sekolah
// =====================================================================
// Semua "jam lokal" (batas terlambat, tanggal absensi, nomor transaksi harian, jadwal pemeliharaan)
// memakai time.Local. Zona waktu sekolah diatur admin, tidak bergantung zona waktu server/cloud.

var zonaWaktu = map[string]string{"": "Ikut server", "Asia/Jakarta": "WIB", "Asia/Makassar": "WITA", "Asia/Jayapura": "WIT"}

func init() {
	configDefaults["tampilan"] = `{"warna":"biru"}`
	configModule["tampilan"] = "__admin__"
	configHooks["tampilan"] = func(db *sql.DB, cfg map[string]any) error {
		switch cfg["warna"] {
		case "biru", "abu", "merah":
			return nil
		}
		return bad("Pilihan warna tidak dikenal.")
	}
	configDefaults["waktu"] = `{"zona":""}`
	configModule["waktu"] = "__admin__"
	configHooks["waktu"] = func(db *sql.DB, cfg map[string]any) error {
		z, _ := cfg["zona"].(string)
		if _, ok := zonaWaktu[z]; !ok {
			return bad("Zona waktu tidak dikenal.")
		}
		// Tidak mengubah time.Local saat server berjalan (balapan data dengan semua goroutine yang membaca jam,
		// dan tanggal/nomor harian bisa bergeser di tengah hari). Zona dipasang sekali saat server mulai.
		return nil
	}
	moduleSchemas = append(moduleSchemas, `
CREATE TABLE IF NOT EXISTS audit_log (
  id      INTEGER PRIMARY KEY,
  waktu   TEXT NOT NULL,
  user_id INTEGER,
  nama    TEXT,
  role    TEXT,
  aksi    TEXT NOT NULL,
  status  INTEGER,
  ip      TEXT,
  detail  TEXT
);
CREATE INDEX IF NOT EXISTS idx_audit_waktu ON audit_log(waktu);
CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_log(user_id, waktu);
`)
	moduleRoutes = append(moduleRoutes, func(a *App, m *http.ServeMux) {
		m.HandleFunc("GET /api/audit", a.h("admin", apiAudit))
	})
}

// dayRangeUTC: rentang tanggal sekolah [from, to] (YYYY-MM-DD, zona sekolah) → batas waktu UTC RFC3339
// [awal from, awal hari setelah to), untuk kolom yang menyimpan waktu UTC (created_at, waktu).
func dayRangeUTC(from, to string) (string, string) {
	t1, _ := time.ParseInLocation("2006-01-02", from, time.Local)
	t2, _ := time.ParseInLocation("2006-01-02", to, time.Local)
	return t1.UTC().Format(time.RFC3339), t2.AddDate(0, 0, 1).UTC().Format(time.RFC3339)
}

func applyTimezone(name string) {
	if name == "" {
		return // ikut zona waktu server
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		log.Println("zona waktu tidak valid:", name)
		return
	}
	time.Local = loc
	log.Println("zona waktu sekolah:", name)
}

// =====================================================================
// Pembatas bcrypt: pengecekan password sengaja lambat (±80 ms, di PC lemah bisa 300 ms).
// Saat ratusan siswa login bersamaan, batasi agar CPU tetap tersisa untuk permintaan lain
// (simpan jawaban CBT, scan absensi). Permintaan yang menunggu terlalu lama diberi pesan sibuk.
// =====================================================================
var bcryptSem = make(chan struct{}, max(1, runtime.NumCPU()-1))

func withBcrypt(fn func() error) error {
	select {
	case bcryptSem <- struct{}{}:
	case <-time.After(20 * time.Second):
		return apiErr{503, "Server sedang melayani banyak login. Coba lagi beberapa detik lagi."}
	}
	defer func() { <-bcryptSem }()
	return fn()
}

func checkPassword(hash, pw string) error {
	return withBcrypt(func() error { return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) })
}

// =====================================================================
// Log aktivitas: setiap permintaan yang mengubah data dicatat otomatis (lihat a.h di main.go).
// =====================================================================
var auditSkip = regexp.MustCompile(`^/api/(cbt/\d+/save|kelulusan/cek|ppdb/cek|ai/models|login)$`)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int)        { s.status = code; s.ResponseWriter.WriteHeader(code) }
func (s *statusRecorder) Unwrap() http.ResponseWriter { return s.ResponseWriter }
func (s *statusRecorder) Flush() {
	if f, ok := s.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (a *App) audit(r *http.Request, u *User, aksi string, status int, detail string) {
	var uid any
	var nama, role any
	if u != nil {
		uid, nama, role = u.ID, u.NamaLengkap, u.Role
	}
	if len(detail) > 500 {
		detail = detail[:500]
	}
	a.db.Exec(`INSERT INTO audit_log (waktu, user_id, nama, role, aksi, status, ip, detail) VALUES (?,?,?,?,?,?,?,?)`,
		nowStr(), uid, nama, role, aksi, status, clientIP(r), nullIfEmpty(detail))
}

func apiAudit(c *Ctx) error {
	q := c.R.URL.Query()
	where, args := []string{"1=1"}, []any{}
	dari := q.Get("dari")
	if s := strings.TrimSpace(q.Get("q")); s != "" {
		where = append(where, "(aksi LIKE ? OR nama LIKE ? OR detail LIKE ? OR ip LIKE ?)")
		like := "%" + s + "%"
		args = append(args, like, like, like, like)
		// pencarian teks (LIKE '%x%') tidak bisa memakai indeks: tanpa tanggal "dari", batasi 90 hari terakhir
		// agar memakai indeks waktu dan tidak memindai seluruh log bertahun-tahun
		if _, err := cleanDate(dari); err != nil {
			dari = time.Now().AddDate(0, 0, -90).Format("2006-01-02")
			c.W.Header().Set("X-Audit-Dari", dari)
		}
	}
	if d, err := cleanDate(dari); err == nil {
		lo, _ := dayRangeUTC(d, d)
		where = append(where, "waktu >= ?")
		args = append(args, lo)
	}
	if d, err := cleanDate(q.Get("sampai")); err == nil {
		_, hi := dayRangeUTC(d, d)
		where = append(where, "waktu < ?")
		args = append(args, hi)
	}
	if q.Get("gagal") == "1" {
		where = append(where, "status >= 400")
	}
	limit, offset := pageParams(c, 100)
	w := strings.Join(where, " AND ")
	var total int
	c.A.reader().QueryRow(`SELECT COUNT(*) FROM (SELECT 1 FROM audit_log WHERE `+w+` LIMIT 10001)`, args...).Scan(&total)
	// urut waktu (bukan id): rencana query memakai idx_audit_waktu untuk rentang tanggal sekaligus urutan
	rows, err := c.list(`SELECT * FROM audit_log WHERE `+w+` ORDER BY waktu DESC, id DESC LIMIT ? OFFSET ?`, append(args, limit, offset)...)
	if err != nil {
		return err
	}
	c.W.Header().Set("X-Total-Count", strconv.Itoa(total))
	return c.JSON(rows)
}

// pageParams membaca ?limit= & ?offset= dengan batas aman.
func pageParams(c *Ctx, def int) (int, int) {
	limit, _ := strconv.Atoi(c.R.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(c.R.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = def
	}
	if limit > 1000 {
		limit = 1000
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// longBody: perpanjang batas waktu baca untuk upload besar lewat jaringan lambat (HP, sinyal lemah).
func (c *Ctx) longBody() {
	http.NewResponseController(c.W).SetReadDeadline(time.Now().Add(10 * time.Minute))
}
