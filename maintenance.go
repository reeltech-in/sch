package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Pemeliharaan harian otomatis (berjalan sekali sehari, mulai pukul 01.00 waktu server, atau saat
// server menyala bila hari itu belum dijalankan):
//   1. backup database harian -> <data>/backups/harian-YYYY-MM-DD.db, simpan N terakhir
//   2. hapus foto selfie absensi & foto buku tamu yang melewati masa simpan
//   3. hapus file upload yatim (tidak dirujuk data mana pun) yang berumur > 24 jam

func init() {
	configDefaults["penyimpanan"] = `{"selfie_hari":60,"tamu_hari":365,"backup_simpan":14,"log_hari":365,"ai_hari":0}`
	configModule["penyimpanan"] = "__admin__" // hanya admin
	moduleRoutes = append(moduleRoutes, func(a *App, m *http.ServeMux) {
		m.HandleFunc("GET /api/backups", a.h("admin", a.apiBackups))
		m.HandleFunc("GET /api/backups/{name}", a.h("admin", a.apiBackupFile))
		m.HandleFunc("POST /api/pemeliharaan", a.h("admin", a.apiMaintainNow))
	})
}

func (a *App) backupDir() string { return filepath.Join(filepath.Dir(a.dbPath), "backups") }

// maybeMaintain menjalankan pemeliharaan bila hari ini belum dan sudah lewat pukul 01.00.
func (a *App) maybeMaintain() {
	today := time.Now().Format("2006-01-02")
	if kvGet(a.db, "maint_last") == today || time.Now().Hour() < 1 {
		return
	}
	rep, ok := a.maintain()
	if ok { // backup gagal → dicoba lagi pada putaran berikutnya (10 menit), tidak ditandai selesai
		kvSet(a.db, "maint_last", today)
	}
	log.Println("pemeliharaan harian:", rep)
}

func cfgInt(m map[string]any, k string, def int) int {
	if v, ok := m[k].(float64); ok && v >= 0 {
		return int(v)
	}
	return def
}

func (a *App) reader() *sql.DB {
	if a.rdb != nil {
		return a.rdb
	}
	return a.db
}

// maintain menjalankan pemeliharaan harian. ok=false bila backup gagal (agar dicoba ulang).
var maintMu sync.Mutex

func (a *App) maintain() (string, bool) {
	// tombol "Jalankan sekarang" dan jadwal harian tidak boleh berjalan bersamaan (file backup .tmp yang sama)
	if !maintMu.TryLock() {
		return "pemeliharaan sedang berjalan", false
	}
	defer maintMu.Unlock()
	cfg := getConfig(a.db, "penyimpanan")
	var notes []string
	ok := true
	// 1. backup harian: tulis ke file sementara lalu ganti nama, agar backup setengah jadi tidak dianggap sah
	if err := os.MkdirAll(a.backupDir(), 0o700); err != nil {
		ok = false
		notes = append(notes, "backup GAGAL: "+err.Error())
	} else {
		name := "harian-" + time.Now().Format("2006-01-02") + ".db"
		f := filepath.Join(a.backupDir(), name)
		if _, err := os.Stat(f); err != nil {
			tmp := f + ".tmp"
			os.Remove(tmp)
			if _, err := a.reader().Exec(`VACUUM INTO ?`, tmp); err != nil {
				ok = false
				os.Remove(tmp)
				notes = append(notes, "backup GAGAL: "+err.Error())
			} else if err := os.Rename(tmp, f); err != nil {
				ok = false
				notes = append(notes, "backup GAGAL: "+err.Error())
			} else {
				os.Chmod(f, 0o600)
				kvSet(a.db, "backup_terakhir", name)
				notes = append(notes, "backup "+name)
			}
		}
		keep := cfgInt(cfg, "backup_simpan", 14)
		if keep < 1 {
			keep = 1
		}
		prune := func(pattern string, keep int) {
			files, _ := filepath.Glob(filepath.Join(a.backupDir(), pattern))
			sort.Strings(files)
			for len(files) > keep {
				os.Remove(files[0])
				files = files[1:]
			}
		}
		prune("harian-*.db", keep)
		// backup otomatis sebelum rilis/migrasi juga dibatasi agar disk tidak penuh perlahan
		prune("sebelum-v*.db", 10)
		prune("sebelum-migrasi-*.db", 10)
	}
	// 2. masa simpan foto
	removed := 0
	purge := func(q, upd string, args ...any) {
		rows, err := a.reader().Query(q, args...)
		if err != nil {
			return
		}
		var urls []string
		for rows.Next() {
			var u string
			if rows.Scan(&u) == nil {
				urls = append(urls, u)
			}
		}
		if rows.Close() != nil || rows.Err() != nil {
			return
		}
		for _, u := range urls {
			if name := strings.TrimPrefix(u, "/files/"); fileNameRe.MatchString(name) {
				if os.Remove(filepath.Join(a.uploadDir(), name)) == nil {
					removed++
				}
			}
		}
		a.db.Exec(upd, args...)
	}
	if d := cfgInt(cfg, "selfie_hari", 60); d > 0 {
		cut := time.Now().AddDate(0, 0, -d).Format("2006-01-02")
		purge(`SELECT foto FROM daily_attendance WHERE foto IS NOT NULL AND tanggal < ?`,
			`UPDATE daily_attendance SET foto=NULL WHERE foto IS NOT NULL AND tanggal < ?`, cut)
	}
	if d := cfgInt(cfg, "tamu_hari", 365); d > 0 {
		cut := time.Now().AddDate(0, 0, -d).UTC().Format(time.RFC3339)
		purge(`SELECT foto FROM guestbook WHERE foto IS NOT NULL AND created_at < ?`,
			`UPDATE guestbook SET foto=NULL WHERE foto IS NOT NULL AND created_at < ?`, cut)
	}
	// log aktivitas lama
	if d := cfgInt(cfg, "log_hari", 365); d > 0 {
		a.db.Exec(`DELETE FROM audit_log WHERE waktu < ?`, time.Now().AddDate(0, 0, -d).UTC().Format(time.RFC3339))
		a.db.Exec(`DELETE FROM wa_outbox WHERE status<>'antre' AND created_at < ?`, time.Now().AddDate(0, 0, -d).UTC().Format(time.RFC3339))
	}
	// dokumen Asisten AI lama (0 = simpan selamanya)
	if d := cfgInt(cfg, "ai_hari", 0); d > 0 {
		a.db.Exec(`DELETE FROM ai_documents WHERE created_at < ?`, time.Now().AddDate(0, 0, -d).UTC().Format(time.RFC3339))
	}
	// penghitung kuota Asisten AI harian (ai_pakai:<tanggal>:<user>) yang lebih dari 7 hari
	a.db.Exec(`DELETE FROM kv WHERE k LIKE 'ai_pakai:%' AND substr(k, 10, 10) < ?`, time.Now().AddDate(0, 0, -7).Format("2006-01-02"))
	// 3. file yatim
	orphans, err := a.removeOrphans(24 * time.Hour)
	trash := a.emptyTrash(30 * 24 * time.Hour)
	if n, err := a.pruneEbookCache(); err != nil {
		notes = append(notes, "pembersihan salinan e-book dilewati: "+err.Error())
	} else if n > 0 {
		notes = append(notes, fmt.Sprintf("%d salinan e-book yang bukunya sudah dihapus dibersihkan", n))
	}
	uploadSize.Lock()
	uploadSize.at = time.Time{} // hitung ulang pemakaian
	uploadSize.Unlock()
	if err != nil {
		notes = append(notes, fmt.Sprintf("%d foto kedaluwarsa dihapus; pembersihan file yatim DILEWATI: %v", removed, err))
	} else {
		notes = append(notes, fmt.Sprintf("%d foto kedaluwarsa dihapus, %d file yatim dipindah ke sampah, %d file sampah >30 hari dihapus", removed, orphans, trash))
	}
	// 4. statistik perencana query (murah; membantu SQLite memilih indeks seiring data bertambah)
	a.db.Exec(`PRAGMA optimize`)
	return strings.Join(notes, "; "), ok
}

var fileRefRe = regexp.MustCompile(`[a-f0-9]{24}\.(?:jpg|png|gif|webp|pdf)`)

// removeOrphans memindahkan file upload yang tidak dirujuk kolom mana pun (dan lebih tua dari minAge) ke folder
// sampah. Bila SATU saja query gagal, seluruh pembersihan dibatalkan: lebih baik ada file yatim daripada menghapus
// file yang masih dipakai (upload tidak ikut di backup database).
func (a *App) removeOrphans(minAge time.Duration) (int, error) {
	refs := map[string]bool{}
	queries := []string{
		`SELECT foto FROM student_profiles`, `SELECT foto FROM daily_attendance WHERE foto IS NOT NULL`, `SELECT gambar FROM exam_questions`,
		`SELECT foto FROM assets`, `SELECT cover FROM books`, `SELECT ebook FROM books`, `SELECT file FROM letters`,
		`SELECT foto FROM guestbook`, `SELECT foto FROM candidates`, `SELECT logo_url FROM settings`,
		`SELECT berkas FROM ppdb_applicants`, `SELECT lampiran FROM materials`, `SELECT konten FROM materials`,
		`SELECT link_file FROM submissions`, `SELECT v FROM kv`,
	}
	for _, q := range queries {
		rows, err := a.reader().Query(q)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", q, err)
		}
		for rows.Next() {
			var s *string
			if err := rows.Scan(&s); err != nil {
				rows.Close()
				return 0, err
			}
			if s != nil {
				for _, m := range fileRefRe.FindAllString(*s, -1) {
					refs[m] = true
				}
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return 0, err
		}
		rows.Close()
	}
	entries, err := os.ReadDir(a.uploadDir())
	if err != nil {
		return 0, err
	}
	trash := filepath.Join(a.uploadDir(), ".sampah")
	if err := os.MkdirAll(trash, 0o700); err != nil {
		return 0, err
	}
	n := 0
	for _, e := range entries {
		if e.IsDir() || !fileNameRe.MatchString(e.Name()) || refs[e.Name()] {
			continue
		}
		if fi, err := e.Info(); err == nil && time.Since(fi.ModTime()) > minAge {
			dst := filepath.Join(trash, e.Name())
			if os.Rename(filepath.Join(a.uploadDir(), e.Name()), dst) == nil {
				now := time.Now()
				os.Chtimes(dst, now, now) // umur di sampah dihitung sejak dipindah
				n++
			}
		}
	}
	return n, nil
}

// emptyTrash menghapus file di folder sampah upload yang sudah lebih lama dari maxAge.
func (a *App) emptyTrash(maxAge time.Duration) int {
	dir := filepath.Join(a.uploadDir(), ".sampah")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		if fi, err := e.Info(); err == nil && !e.IsDir() && time.Since(fi.ModTime()) > maxAge {
			if os.Remove(filepath.Join(dir, e.Name())) == nil {
				n++
			}
		}
	}
	return n
}

var backupNameRe = regexp.MustCompile(`^(harian|sebelum)-[A-Za-z0-9.\-]+\.db$`)

func (a *App) apiBackups(c *Ctx) error {
	entries, _ := os.ReadDir(a.backupDir())
	out := []map[string]any{}
	for _, e := range entries {
		if fi, err := e.Info(); err == nil && backupNameRe.MatchString(e.Name()) {
			out = append(out, map[string]any{"nama": e.Name(), "ukuran": fi.Size(), "waktu": fi.ModTime().UTC().Format(time.RFC3339)})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i]["waktu"].(string) > out[j]["waktu"].(string) })
	return c.JSON(out)
}

func (a *App) apiBackupFile(c *Ctx) error {
	name := c.R.PathValue("name")
	if !backupNameRe.MatchString(name) {
		return notFound()
	}
	p := filepath.Join(a.backupDir(), name)
	if _, err := os.Stat(p); err != nil {
		return notFound()
	}
	http.NewResponseController(c.W).SetWriteDeadline(time.Now().Add(10 * time.Minute))
	c.W.Header().Set("Content-Type", "application/octet-stream")
	c.W.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	http.ServeFile(c.W, c.R, p)
	return nil
}

func (a *App) apiMaintainNow(c *Ctx) error {
	rep, ok := a.maintain()
	if ok {
		kvSet(a.db, "maint_last", time.Now().Format("2006-01-02"))
	}
	return c.JSON(map[string]string{"hasil": rep})
}

// storageInfo untuk dashboard admin.
func (a *App) storageInfo() map[string]any {
	used := a.uploadUsed()
	pct := 0.0
	if uploadMaxBytes > 0 {
		pct = float64(used) / float64(uploadMaxBytes) * 100
	}
	return map[string]any{"terpakai": used, "batas": uploadMaxBytes, "persen": strconv.FormatFloat(pct, 'f', 1, 64),
		"backup_terakhir": kvGet(a.db, "backup_terakhir")}
}
