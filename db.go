package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

const schemaSQL = `
CREATE TABLE IF NOT EXISTS settings (
  id             INTEGER PRIMARY KEY CHECK (id = 1),
  nama_sekolah   TEXT NOT NULL DEFAULT 'Sekolah Saya',
  logo_url       TEXT,
  alamat         TEXT,
  telepon        TEXT,
  tahun_ajaran   TEXT,
  kepala_sekolah TEXT,
  nip_kepala     TEXT,
  akses_guru     TEXT NOT NULL DEFAULT '{"materi":true,"tugas":true,"kuis":true,"absensi":true,"hapus_absen":true}'
);
INSERT OR IGNORE INTO settings (id) VALUES (1);

CREATE TABLE IF NOT EXISTS classes (
  id           INTEGER PRIMARY KEY,
  nama         TEXT NOT NULL UNIQUE COLLATE NOCASE,
  tingkat      TEXT,
  tahun_ajaran TEXT,
  created_at   TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
  id            INTEGER PRIMARY KEY,
  username      TEXT NOT NULL UNIQUE COLLATE NOCASE,
  password_hash TEXT NOT NULL,
  nama_lengkap  TEXT NOT NULL,
  role          TEXT NOT NULL CHECK (role IN ('admin','guru','siswa')),
  email         TEXT,
  status        TEXT NOT NULL DEFAULT 'aktif' CHECK (status IN ('aktif','nonaktif')),
  kelas_id      INTEGER REFERENCES classes(id) ON DELETE SET NULL,
  nip           TEXT,
  created_at    TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_users_kelas ON users(kelas_id);

CREATE TABLE IF NOT EXISTS sessions (
  token_hash TEXT PRIMARY KEY,
  user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS courses (
  id         INTEGER PRIMARY KEY,
  kode       TEXT NOT NULL,
  nama       TEXT NOT NULL,
  deskripsi  TEXT,
  kelas_id   INTEGER NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
  guru_id    INTEGER REFERENCES users(id) ON DELETE SET NULL,
  created_at TEXT NOT NULL,
  UNIQUE (kelas_id, kode COLLATE NOCASE)
);
CREATE INDEX IF NOT EXISTS idx_courses_guru ON courses(guru_id);

CREATE TABLE IF NOT EXISTS materials (
  id         INTEGER PRIMARY KEY,
  course_id  INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
  judul      TEXT NOT NULL,
  konten     TEXT,
  lampiran   TEXT NOT NULL DEFAULT '[]',
  urutan     INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_materials_course ON materials(course_id);

CREATE TABLE IF NOT EXISTS material_progress (
  material_id  INTEGER NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
  siswa_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  completed_at TEXT NOT NULL,
  PRIMARY KEY (material_id, siswa_id)
);

CREATE TABLE IF NOT EXISTS assignments (
  id         INTEGER PRIMARY KEY,
  course_id  INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
  judul      TEXT NOT NULL,
  deskripsi  TEXT,
  deadline   TEXT,
  nilai_maks REAL NOT NULL DEFAULT 100,
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_assign_course ON assignments(course_id);

CREATE TABLE IF NOT EXISTS submissions (
  id            INTEGER PRIMARY KEY,
  assignment_id INTEGER NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
  siswa_id      INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  jawaban       TEXT,
  link_file     TEXT,
  nilai         REAL,
  feedback      TEXT,
  submitted_at  TEXT NOT NULL,
  UNIQUE (assignment_id, siswa_id)
);
CREATE INDEX IF NOT EXISTS idx_subs_siswa ON submissions(siswa_id);

CREATE TABLE IF NOT EXISTS quizzes (
  id           INTEGER PRIMARY KEY,
  course_id    INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
  judul        TEXT NOT NULL,
  durasi_menit INTEGER NOT NULL DEFAULT 15,
  jumlah_opsi  INTEGER NOT NULL DEFAULT 4 CHECK (jumlah_opsi BETWEEN 2 AND 5),
  created_at   TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_quiz_course ON quizzes(course_id);

CREATE TABLE IF NOT EXISTS quiz_questions (
  id            INTEGER PRIMARY KEY,
  quiz_id       INTEGER NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
  pertanyaan    TEXT NOT NULL,
  opsi_a        TEXT NOT NULL,
  opsi_b        TEXT NOT NULL,
  opsi_c        TEXT,
  opsi_d        TEXT,
  opsi_e        TEXT,
  jawaban_benar TEXT NOT NULL CHECK (jawaban_benar IN ('a','b','c','d','e')),
  bobot         REAL NOT NULL DEFAULT 1,
  created_at    TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_quizq_quiz ON quiz_questions(quiz_id);

CREATE TABLE IF NOT EXISTS quiz_attempts (
  id           INTEGER PRIMARY KEY,
  quiz_id      INTEGER NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
  siswa_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  skor         REAL NOT NULL DEFAULT 0,
  jawaban      TEXT,
  started_at   TEXT NOT NULL,
  submitted_at TEXT,
  selesai      INTEGER NOT NULL DEFAULT 0,
  UNIQUE (quiz_id, siswa_id)
);

CREATE TABLE IF NOT EXISTS attendance (
  course_id INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
  siswa_id  INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  tanggal   TEXT NOT NULL,
  status    TEXT NOT NULL CHECK (status IN ('hadir','izin','sakit','alfa')),
  PRIMARY KEY (course_id, siswa_id, tanggal)
);
CREATE INDEX IF NOT EXISTS idx_attend_siswa ON attendance(siswa_id);
`

func openDB(path string) (*sql.DB, error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	dsn := "file:" + path + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// Satu koneksi: SQLite menulis satu per satu; ini paling stabil untuk PC sederhana.
	db.SetMaxOpenConns(1)
	// data sensitif (hash password, kunci AI, rahasia kartu): hanya pemilik proses yang boleh membaca
	os.Chmod(filepath.Dir(path), 0o700)
	defer func() {
		for _, f := range []string{path, path + "-wal", path + "-shm"} {
			os.Chmod(f, 0o600)
		}
	}()
	if _, err := db.Exec(schemaSQL); err != nil {
		return nil, fmt.Errorf("membuat tabel: %w", err)
	}
	if err := runModuleSchemas(db); err != nil {
		return nil, fmt.Errorf("membuat tabel modul: %w", err)
	}
	if err := runVersioned(db, path); err != nil {
		return nil, err
	}
	return db, nil
}

// openReadDB membuka pool khusus baca (WAL mengizinkan banyak pembaca bersamaan dengan satu penulis).
func openReadDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro&_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(4)
	return db, db.Ping()
}

func nowStr() string { return time.Now().UTC().Format(time.RFC3339) }

func hashPassword(pw string) (string, error) {
	var b []byte
	err := withBcrypt(func() (e error) { b, e = bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost); return })
	return string(b), err
}

func randomPassword(n int) string {
	const chars = "abcdefghjkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var sb strings.Builder
	for i := 0; i < n; i++ {
		k, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		sb.WriteByte(chars[k.Int64()])
	}
	return sb.String()
}

// ensureAdmin membuat akun admin pertama bila belum ada, atau mereset password-nya bila reset=true.
func ensureAdmin(db *sql.DB, dbPath string, reset bool) error {
	var id int64
	err := db.QueryRow(`SELECT id FROM users WHERE role='admin' ORDER BY id LIMIT 1`).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil && !reset {
		return nil
	}
	pw := randomPassword(10)
	h, err := hashPassword(pw)
	if err != nil {
		return err
	}
	if id == 0 {
		_, err = db.Exec(`INSERT INTO users (username, password_hash, nama_lengkap, role, created_at) VALUES ('admin', ?, 'Administrator', 'admin', ?)`, h, nowStr())
	} else {
		_, err = db.Exec(`UPDATE users SET password_hash=?, status='aktif' WHERE id=?`, h, id)
		if err == nil {
			_, err = db.Exec(`DELETE FROM sessions WHERE user_id=?`, id)
		}
	}
	if err != nil {
		return err
	}
	var uname string
	db.QueryRow(`SELECT username FROM users WHERE role='admin' ORDER BY id LIMIT 1`).Scan(&uname)
	note := filepath.Join(filepath.Dir(dbPath), "ADMIN-PASSWORD.txt")
	msg := fmt.Sprintf("Akun admin SCH\nUsername : %s\nPassword : %s\n\nSegera login, ganti password di menu Akun, lalu HAPUS file ini.\n", uname, pw)
	_ = os.WriteFile(note, []byte(msg), 0o600)
	fmt.Println("==================================================")
	fmt.Println(" AKUN ADMIN")
	fmt.Println("   Username :", uname)
	fmt.Println("   Password :", pw)
	fmt.Println(" (juga disimpan di", note+")")
	fmt.Println("==================================================")
	return nil
}

// rowsToMaps mengubah hasil query menjadi slice map untuk dikirim sebagai JSON.
func rowsToMaps(rows *sql.Rows) ([]map[string]any, error) {
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	out := []map[string]any{}
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		m := make(map[string]any, len(cols))
		for i, c := range cols {
			v := vals[i]
			if b, ok := v.([]byte); ok {
				v = string(b)
			}
			if s, ok := v.(string); ok && (c == "lampiran" || c == "akses_guru" || c == "jawaban_json" || c == "akses") {
				if json.Valid([]byte(s)) {
					v = json.RawMessage(s)
				}
			}
			m[c] = v
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
