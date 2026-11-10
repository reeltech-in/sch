package main

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"strings"
)

// Modul Layanan & Administrasi: PPDB online, Buku Tamu, Surat Masuk/Keluar.

func init() {
	configDefaults["ppdb"] = `{"buka":null,"tutup":null,"tahun_ajaran":"","kuota":0,"jalur":"Zonasi, Afirmasi, Prestasi, Mutasi","info":"Siapkan scan/foto Kartu Keluarga, Akta Kelahiran, dan pas foto (JPG/PNG/PDF, maks 2 MB per file)."}`
	configDefaults["surat"] = `{"singkatan":"SKL","format":"{no}/{kode}/{singkatan}/{bulan}/{tahun}"}`
	moduleSchemas = append(moduleSchemas, `
CREATE TABLE IF NOT EXISTS ppdb_applicants (
  id           INTEGER PRIMARY KEY,
  no_daftar    TEXT NOT NULL UNIQUE,
  pin_hash     TEXT NOT NULL,
  nama         TEXT NOT NULL,
  nisn TEXT, nik TEXT, jk TEXT, tempat_lahir TEXT, tgl_lahir TEXT, agama TEXT, alamat TEXT,
  asal_sekolah TEXT, nama_ayah TEXT, nama_ibu TEXT, hp TEXT, email TEXT,
  jalur        TEXT,
  berkas       TEXT NOT NULL DEFAULT '{}',
  status       TEXT NOT NULL DEFAULT 'baru' CHECK (status IN ('baru','verifikasi','diterima','cadangan','ditolak')),
  catatan      TEXT,
  user_id      INTEGER REFERENCES users(id) ON DELETE SET NULL,
  created_at   TEXT NOT NULL,
  updated_at   TEXT
);
CREATE TABLE IF NOT EXISTS guestbook (
  id         INTEGER PRIMARY KEY,
  nama       TEXT NOT NULL,
  instansi   TEXT,
  hp         TEXT,
  keperluan  TEXT NOT NULL,
  bertemu    TEXT,
  foto       TEXT,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS letters (
  id              INTEGER PRIMARY KEY,
  jenis           TEXT NOT NULL CHECK (jenis IN ('masuk','keluar')),
  nomor           TEXT,
  tanggal         TEXT NOT NULL,
  pihak           TEXT NOT NULL,
  perihal         TEXT NOT NULL,
  kode            TEXT,
  ringkasan       TEXT,
  file            TEXT,
  disposisi_ke    TEXT,
  disposisi_isi   TEXT,
  status          TEXT NOT NULL DEFAULT 'baru' CHECK (status IN ('baru','didisposisi','selesai')),
  created_by      INTEGER REFERENCES users(id) ON DELETE SET NULL,
  created_at      TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_letters ON letters(jenis, tanggal);
`)
	moduleRoutes = append(moduleRoutes, func(a *App, m *http.ServeMux) {
		p := func(roles string, fn func(*Ctx) error) http.HandlerFunc { return a.mod("ppdb", roles, fn) }
		m.HandleFunc("GET /api/ppdb/info", p("", apiPPDBInfo))
		m.HandleFunc("POST /api/ppdb/daftar", p("", a.apiPPDBDaftar))
		m.HandleFunc("POST /api/ppdb/cek", p("", apiPPDBCek))
		m.HandleFunc("GET /api/ppdb", p("admin", apiPPDBList))
		m.HandleFunc("PUT /api/ppdb/{id}", p("admin", apiPPDBUpdate))
		m.HandleFunc("DELETE /api/ppdb/{id}", p("admin", apiDelete("ppdb_applicants")))
		m.HandleFunc("POST /api/ppdb/{id}/siswa", p("admin", apiPPDBToStudent))

		g := func(roles string, fn func(*Ctx) error) http.HandlerFunc { return a.mod("bukutamu", roles, fn) }
		m.HandleFunc("POST /api/bukutamu", g("", a.apiGuestAdd))
		m.HandleFunc("GET /api/bukutamu", g("admin", apiGuestList))
		m.HandleFunc("DELETE /api/bukutamu/{id}", g("admin", apiDelete("guestbook")))

		s := func(roles string, fn func(*Ctx) error) http.HandlerFunc { return a.mod("surat", roles, fn) }
		m.HandleFunc("GET /api/surat", s("admin", apiLetters))
		m.HandleFunc("GET /api/surat/nomor", s("admin", apiLetterNumber))
		m.HandleFunc("POST /api/surat", s("admin", apiLetterSave))
		m.HandleFunc("PUT /api/surat/{id}", s("admin", apiLetterSave))
		m.HandleFunc("DELETE /api/surat/{id}", s("admin", apiDelete("letters")))
	})
}

func randDigits(n int) string {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		k, _ := rand.Int(rand.Reader, big.NewInt(10))
		sb.WriteByte(byte('0' + k.Int64()))
	}
	return sb.String()
}
