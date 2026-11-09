package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Migrasi bernomor versi (PRAGMA user_version).
//
// Aturan menambah migrasi:
//   - TAMBAHKAN di akhir daftar `migrations`, jangan pernah mengubah/menghapus yang sudah dirilis.
//   - Satu migrasi = satu fungsi; dijalankan sekali, berurutan, di dalam transaksi.
//   - Bila perlu membangun ulang tabel (mengubah CHECK/UNIQUE/FOREIGN KEY), set rebuild: true agar
//     foreign key dimatikan sementara sesuai prosedur resmi SQLite.
//
// Sebelum migrasi pada database yang sudah berisi data, salinan otomatis dibuat di
// <folder data>/backups/sebelum-migrasi-v<N>-<waktu>.db.
type migration struct {
	name    string
	rebuild bool
	up      func(tx *sql.Tx) error
}

var migrations = []migration{
	{name: "tahun ajaran, kelas per tahun, riwayat anggota kelas", rebuild: true, up: migTahunAjaran},
}

func runVersioned(db *sql.DB, dbPath string) error {
	var cur int
	if err := db.QueryRow(`PRAGMA user_version`).Scan(&cur); err != nil {
		return err
	}
	if cur > len(migrations) {
		return fmt.Errorf("database versi %d lebih baru dari program ini (mengenal sampai versi %d). Pakai rilis program yang lebih baru; jangan menjalankan program lama pada database yang sudah dimigrasi", cur, len(migrations))
	}
	if cur >= len(migrations) {
		return nil
	}
	// backup bila database sudah berisi data (ada pengguna selain admin awal)
	var users int
	db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&users)
	if users > 1 {
		dir := filepath.Join(filepath.Dir(dbPath), "backups")
		os.MkdirAll(dir, 0o700)
		f := filepath.Join(dir, fmt.Sprintf("sebelum-migrasi-v%d-%s.db", cur+1, time.Now().Format("20060102-150405")))
		if _, err := db.Exec(`VACUUM INTO ?`, f); err != nil {
			return fmt.Errorf("backup sebelum migrasi gagal: %w", err)
		}
		os.Chmod(f, 0o600)
		log.Println("backup sebelum migrasi:", f)
	}
	for v := cur; v < len(migrations); v++ {
		m := migrations[v]
		log.Printf("migrasi database v%d: %s", v+1, m.name)
		if m.rebuild {
			if _, err := db.Exec(`PRAGMA foreign_keys=OFF`); err != nil {
				return err
			}
		}
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if err := m.up(tx); err != nil {
			tx.Rollback()
			db.Exec(`PRAGMA foreign_keys=ON`)
			return fmt.Errorf("migrasi v%d (%s): %w", v+1, m.name, err)
		}
		if m.rebuild {
			// pastikan tidak ada referensi yang putus sebelum disimpan
			rows, err := tx.Query(`PRAGMA foreign_key_check`)
			if err != nil {
				tx.Rollback()
				return err
			}
			bad := rows.Next()
			rows.Close()
			if bad {
				tx.Rollback()
				db.Exec(`PRAGMA foreign_keys=ON`)
				return fmt.Errorf("migrasi v%d: pemeriksaan foreign key gagal", v+1)
			}
		}
		if _, err := tx.Exec(fmt.Sprintf(`PRAGMA user_version=%d`, v+1)); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
		if m.rebuild {
			db.Exec(`PRAGMA foreign_keys=ON`)
		}
	}
	return nil
}

// defaultTahunAjaran: Juli–Desember = tahun ini/tahun depan, Januari–Juni = tahun lalu/tahun ini.
func defaultTahunAjaran(t time.Time) string {
	y := t.Year()
	if t.Month() < time.July {
		y--
	}
	return fmt.Sprintf("%d/%d", y, y+1)
}

// v1: tabel tahun ajaran; kelas terikat ke tahun ajaran (nama kelas unik per tahun); riwayat anggota kelas.
func migTahunAjaran(tx *sql.Tx) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS academic_years (
			id         INTEGER PRIMARY KEY,
			nama       TEXT NOT NULL UNIQUE,
			mulai      TEXT,
			selesai    TEXT,
			aktif      INTEGER NOT NULL DEFAULT 0,
			dinaikkan  INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL
		)`,
	}
	for _, s := range stmts {
		if _, err := tx.Exec(s); err != nil {
			return err
		}
	}
	var n int
	tx.QueryRow(`SELECT COUNT(*) FROM academic_years`).Scan(&n)
	if n == 0 {
		var nama sql.NullString
		tx.QueryRow(`SELECT NULLIF(trim(tahun_ajaran),'') FROM settings WHERE id=1`).Scan(&nama)
		if !nama.Valid {
			tx.QueryRow(`SELECT NULLIF(trim(tahun_ajaran),'') FROM classes WHERE NULLIF(trim(tahun_ajaran),'') IS NOT NULL LIMIT 1`).Scan(&nama)
		}
		tn := nama.String
		if tn == "" {
			tn = defaultTahunAjaran(time.Now())
		}
		if _, err := tx.Exec(`INSERT INTO academic_years (nama, aktif, created_at) VALUES (?,1,?)`, tn, nowStr()); err != nil {
			return err
		}
		tx.Exec(`UPDATE settings SET tahun_ajaran=? WHERE id=1`, tn)
	}
	var aktif int64
	if err := tx.QueryRow(`SELECT id FROM academic_years WHERE aktif=1 ORDER BY id LIMIT 1`).Scan(&aktif); err != nil {
		return err
	}
	rebuild := []string{
		`CREATE TABLE classes_new (
			id           INTEGER PRIMARY KEY,
			tahun_id     INTEGER NOT NULL REFERENCES academic_years(id) ON DELETE RESTRICT,
			nama         TEXT NOT NULL COLLATE NOCASE,
			tingkat      TEXT,
			tahun_ajaran TEXT,
			created_at   TEXT NOT NULL,
			UNIQUE (tahun_id, nama)
		)`,
		fmt.Sprintf(`INSERT INTO classes_new (id, tahun_id, nama, tingkat, tahun_ajaran, created_at)
			SELECT id, %d, nama, tingkat, tahun_ajaran, created_at FROM classes`, aktif),
		`DROP TABLE classes`,
		`ALTER TABLE classes_new RENAME TO classes`,
		`CREATE INDEX IF NOT EXISTS idx_classes_tahun ON classes(tahun_id)`,
		`CREATE TABLE IF NOT EXISTS class_members (
			siswa_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			kelas_id INTEGER NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
			tahun_id INTEGER NOT NULL REFERENCES academic_years(id),
			status   TEXT NOT NULL CHECK (status IN ('naik','tinggal','lulus','keluar')),
			PRIMARY KEY (siswa_id, kelas_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_members_kelas ON class_members(kelas_id)`,
	}
	for _, s := range rebuild {
		if _, err := tx.Exec(s); err != nil {
			return err
		}
	}
	// kolom tambahan pengguna (aman bila sudah ada)
	var has int
	tx.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='angkatan_lulus'`).Scan(&has)
	if has == 0 {
		if _, err := tx.Exec(`ALTER TABLE users ADD COLUMN angkatan_lulus TEXT`); err != nil {
			return err
		}
	}
	return nil
}
