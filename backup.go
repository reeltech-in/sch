package main

import (
	"archive/zip"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Backup lengkap (database + berkas unggahan) dan pemulihan dari backup.
//
// Backup harian otomatis (maintenance.go) hanya berisi database. Foto, berkas PPDB, gambar soal, dan e-book ada di
// folder uploads; tanpa itu, database hasil pemulihan berisi tautan ke berkas yang hilang. Backup lengkap berbentuk
// .zip agar bisa dibuka di Windows tanpa program tambahan.

// backupDirFlag: -backup-dir, folder backup harian di disk lain (USB/NAS). Kosong = <folder data>/backups.
var backupDirFlag string

// apiBackupLengkap mengirim .zip berisi salinan database yang konsisten (VACUUM INTO) dan semua berkas unggahan.
func (a *App) apiBackupLengkap(c *Ctx) error {
	dir, err := os.MkdirTemp("", "sch-backup-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	tmp := filepath.Join(dir, "sch.db")
	db := c.A.rdb
	if db == nil {
		db = c.A.db
	}
	if _, err := db.ExecContext(context.Background(), `VACUUM INTO ?`, tmp); err != nil {
		return err
	}
	// berkas sekolah bisa beberapa GB: beri waktu panjang untuk diunduh
	http.NewResponseController(c.W).SetWriteDeadline(time.Now().Add(3 * time.Hour))
	name := "sch-backup-lengkap-" + time.Now().Format("2006-01-02-1504") + ".zip"
	c.W.Header().Set("Content-Type", "application/zip")
	c.W.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	if err := tulisZipBackup(c.W, tmp, a.uploadDir()); err != nil {
		// status 200 sudah terkirim; unduhan di browser akan terlihat gagal/tidak lengkap
		log.Printf("backup lengkap terputus: %v", err)
	}
	return nil
}

// tulisZipBackup menulis sch.db (dikompresi) dan uploads/<nama> (disimpan apa adanya: gambar & PDF sudah terkompresi).
func tulisZipBackup(w io.Writer, dbFile, uploadDir string) error {
	zw := zip.NewWriter(w)
	tambah := func(nama, path string, metode uint16) error {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil {
			return err
		}
		h, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		h.Name, h.Method = nama, metode
		dst, err := zw.CreateHeader(h)
		if err != nil {
			return err
		}
		_, err = io.Copy(dst, f)
		return err
	}
	if err := tambah("sch.db", dbFile, zip.Deflate); err != nil {
		return err
	}
	entries, err := os.ReadDir(uploadDir)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, e := range entries {
		if !e.Type().IsRegular() || !fileNameRe.MatchString(e.Name()) {
			continue
		}
		if err := tambah("uploads/"+e.Name(), filepath.Join(uploadDir, e.Name()), zip.Store); err != nil {
			return fmt.Errorf("%s: %w", e.Name(), err)
		}
	}
	return zw.Close()
}

// pulihkanBackup mengganti database (dan berkas unggahan, bila sumbernya .zip lengkap) dengan isi backup.
// Dijalankan lewat `sch-server -db data/sch.db -restore <file>` saat server DIMATIKAN. Data yang sedang terpasang
// tidak dihapus: dipindahkan ke backups/sebelum-pulih-<waktu>.db dan uploads-sebelum-pulih-<waktu>.
func pulihkanBackup(dbPath, src string) (string, error) {
	if fi, err := os.Stat(dbPath + "-wal"); err == nil && fi.Size() > 0 {
		return "", fmt.Errorf("server tampaknya masih berjalan (file %s berisi data). Matikan server dulu "+
			"(Linux: sudo systemctl stop sch; Windows: tutup jendela SCH), lalu ulangi. Bila server sudah mati tetapi "+
			"pesan ini tetap muncul, jalankan server sekali lalu matikan dengan normal", filepath.Base(dbPath)+"-wal")
	}
	dataDir := filepath.Dir(dbPath)
	stamp := time.Now().Format("20060102-150405")
	tmpDB := dbPath + ".pulih-tmp"
	tmpUp := filepath.Join(dataDir, "uploads.pulih-tmp")
	os.Remove(tmpDB)
	os.RemoveAll(tmpUp)
	defer os.Remove(tmpDB)
	defer os.RemoveAll(tmpUp)

	adaUploads := false
	if strings.HasSuffix(strings.ToLower(src), ".zip") {
		n, err := bongkarZipBackup(src, tmpDB, tmpUp)
		if err != nil {
			return "", err
		}
		adaUploads = true
		log.Printf("pulihkan: %d berkas unggahan dari backup", n)
	} else if err := salinFile(src, tmpDB); err != nil {
		return "", err
	}
	if err := periksaDatabaseBackup(tmpDB); err != nil {
		return "", err
	}

	// simpan data yang sedang terpasang, lalu pasang backup
	ringkas := []string{}
	if _, err := os.Stat(dbPath); err == nil {
		dir := filepath.Join(dataDir, "backups")
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return "", err
		}
		lama := filepath.Join(dir, "sebelum-pulih-"+stamp+".db")
		if err := os.Rename(dbPath, lama); err != nil {
			return "", err
		}
		ringkas = append(ringkas, "database lama disimpan di "+lama)
	}
	os.Remove(dbPath + "-wal")
	os.Remove(dbPath + "-shm")
	if err := os.Rename(tmpDB, dbPath); err != nil {
		return "", err
	}
	os.Chmod(dbPath, 0o600)
	ringkas = append(ringkas, "database dipulihkan dari "+src)
	if adaUploads {
		up := filepath.Join(dataDir, "uploads")
		if _, err := os.Stat(up); err == nil {
			lama := filepath.Join(dataDir, "uploads-sebelum-pulih-"+stamp)
			if err := os.Rename(up, lama); err != nil {
				return "", err
			}
			ringkas = append(ringkas, "berkas unggahan lama disimpan di "+lama)
		}
		if err := os.Rename(tmpUp, up); err != nil {
			return "", err
		}
		ringkas = append(ringkas, "berkas unggahan dipulihkan")
	}
	return strings.Join(ringkas, "\n"), nil
}

// bongkarZipBackup mengekstrak sch.db dan uploads/<nama berkas sah> dari backup lengkap. Nama lain diabaikan
// (tidak bisa menulis ke luar folder data).
func bongkarZipBackup(src, dbOut, upOut string) (int, error) {
	zr, err := zip.OpenReader(src)
	if err != nil {
		return 0, fmt.Errorf("file backup .zip tidak bisa dibuka: %w", err)
	}
	defer zr.Close()
	if err := os.MkdirAll(upOut, 0o700); err != nil {
		return 0, err
	}
	adaDB, n := false, 0
	for _, f := range zr.File {
		var out string
		switch {
		case f.Name == "sch.db":
			out, adaDB = dbOut, true
		case strings.HasPrefix(f.Name, "uploads/") && fileNameRe.MatchString(strings.TrimPrefix(f.Name, "uploads/")):
			out = filepath.Join(upOut, strings.TrimPrefix(f.Name, "uploads/"))
			n++
		default:
			continue
		}
		if err := ekstrak(f, out); err != nil {
			return 0, fmt.Errorf("%s: %w", f.Name, err)
		}
	}
	if !adaDB {
		return 0, fmt.Errorf("file .zip ini bukan backup lengkap SCH (tidak ada sch.db di dalamnya)")
	}
	return n, nil
}

func ekstrak(f *zip.File, out string) error {
	r, err := f.Open()
	if err != nil {
		return err
	}
	defer r.Close()
	w, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, r); err != nil {
		w.Close()
		return err
	}
	return w.Close()
}

func salinFile(src, dst string) error {
	r, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("file backup tidak bisa dibuka: %w", err)
	}
	defer r.Close()
	w, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(w, r); err != nil {
		w.Close()
		return err
	}
	return w.Close()
}

// periksaDatabaseBackup memastikan file adalah database SCH yang utuh sebelum dipasang.
func periksaDatabaseBackup(path string) error {
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		return err
	}
	defer db.Close()
	var hasil string
	if err := db.QueryRow(`PRAGMA integrity_check`).Scan(&hasil); err != nil {
		return fmt.Errorf("file backup bukan database SQLite yang valid: %w", err)
	}
	if hasil != "ok" {
		return fmt.Errorf("file backup rusak (pemeriksaan integritas: %s)", hasil)
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE role='admin'`).Scan(&n); err != nil || n == 0 {
		return fmt.Errorf("file backup bukan database SCH (tidak ada akun admin)")
	}
	return nil
}
