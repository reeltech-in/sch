package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// Kerangka handler: Ctx, pemeriksaan sesi/peran/CSRF, log aktivitas, galat JSON, daftar baris.

// ---------------------------------------------------------------------
// Kerangka handler
// ---------------------------------------------------------------------
type apiErr struct {
	code int
	msg  string
}

func (e apiErr) Error() string { return e.msg }

func bad(msg string) error { return apiErr{400, msg} }

func forbidden() error { return apiErr{403, "Anda tidak punya akses untuk tindakan ini."} }

func notFound() error { return apiErr{404, "Data tidak ditemukan."} }

type Ctx struct {
	W    http.ResponseWriter
	R    *http.Request
	U    *User
	A    *App
	Note string // keterangan tambahan untuk log aktivitas
}

func (c *Ctx) JSON(v any) error {
	c.W.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.W.Header().Set("Cache-Control", "no-store")
	return json.NewEncoder(c.W).Encode(v)
}

func (c *Ctx) OK() error { return c.JSON(map[string]bool{"ok": true}) }

func (c *Ctx) Bind(v any) error {
	dec := json.NewDecoder(http.MaxBytesReader(c.W, c.R.Body, 2<<20))
	if err := dec.Decode(v); err != nil {
		return bad("Data yang dikirim tidak valid.")
	}
	return nil
}

func (c *Ctx) ID(name string) int64 {
	n, _ := strconv.ParseInt(c.R.PathValue(name), 10, 64)
	return n
}

func (c *Ctx) QInt(name string) int64 {
	n, _ := strconv.ParseInt(c.R.URL.Query().Get(name), 10, 64)
	return n
}

// list menjalankan SELECT. Permintaan GET memakai pool baca sehingga tidak antre di belakang
// penulisan (mis. simpan otomatis CBT); permintaan lain memakai koneksi tulis agar konsisten.
func (c *Ctx) list(q string, args ...any) ([]map[string]any, error) {
	db := c.A.db
	if c.R != nil && c.R.Method == http.MethodGet && c.A.rdb != nil {
		db = c.A.rdb
	}
	rows, err := db.QueryContext(c.R.Context(), q, args...)
	if err != nil {
		return nil, err
	}
	return rowsToMaps(rows)
}

// h membungkus handler dengan cek login & peran.
// roles: "" = publik, "any" = sudah login, atau daftar dipisah koma: "admin,guru".
func (a *App) h(roles string, fn func(*Ctx) error) http.HandlerFunc {
	return func(w0 http.ResponseWriter, r *http.Request) {
		// catat otomatis setiap permintaan yang mengubah data (log aktivitas)
		rec := &statusRecorder{ResponseWriter: w0, status: 200}
		w := http.ResponseWriter(rec)
		c := &Ctx{W: w, R: r, A: a}
		if r.Method != http.MethodGet && !auditSkip.MatchString(r.URL.Path) {
			defer func() { a.audit(r, c.U, r.Method+" "+r.URL.Path, rec.status, c.Note) }()
		}
		// Perlindungan CSRF: permintaan yang mengubah data wajib membawa header khusus,
		// yang tidak bisa dikirim situs lain tanpa izin CORS.
		if r.Method != http.MethodGet && r.Header.Get("X-SCH") != "1" {
			writeErr(w, apiErr{403, "Permintaan ditolak."})
			return
		}
		if roles != "" {
			u := a.currentUser(r)
			if u == nil {
				writeErr(w, apiErr{401, "Sesi berakhir. Silakan login lagi."})
				return
			}
			if roles != "any" && !strings.Contains(","+roles+",", ","+u.Role+",") {
				writeErr(w, forbidden())
				return
			}
			// Akun yang wajib ganti password (hasil impor/reset admin) hanya boleh mengganti password dulu.
			if u.GantiPW && r.URL.Path != "/api/password" && r.URL.Path != "/api/me" && r.URL.Path != "/api/logout" {
				writeErr(w, apiErr{403, "Silakan ganti password Anda terlebih dahulu."})
				return
			}
			c.U = u
		}
		if err := fn(c); err != nil {
			writeErr(w, err)
		}
	}
}

// bisaDipaksa: penolakan hapus (409) yang boleh dilewati admin dengan ?paksa=1 (dikirim sebagai "paksa": true)
type bisaDipaksa struct{ apiErr }

func (e bisaDipaksa) Unwrap() error { return e.apiErr }

func writeErr(w http.ResponseWriter, err error) {
	var bp bisaDipaksa
	paksa := errors.As(err, &bp)
	var ae apiErr
	if !errors.As(err, &ae) {
		msg := err.Error()
		switch {
		case strings.Contains(msg, "UNIQUE constraint failed"):
			ae = apiErr{400, "Data sudah ada (duplikat). Periksa nama, kode, atau username."}
		case strings.Contains(msg, "FOREIGN KEY constraint failed"):
			ae = apiErr{400, "Data terkait tidak ditemukan."}
		case strings.Contains(msg, "CHECK constraint failed"):
			ae = apiErr{400, "Ada isian yang nilainya tidak diperbolehkan."}
		case errors.Is(err, sql.ErrNoRows):
			ae = notFound().(apiErr)
		default:
			log.Println("error:", err)
			ae = apiErr{500, "Terjadi kesalahan di server."}
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(ae.code)
	out := map[string]any{"error": ae.msg}
	if paksa {
		out["paksa"] = true
	}
	json.NewEncoder(w).Encode(out)
}
