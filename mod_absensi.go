package main

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Modul: Buku Induk Siswa, Absensi Harian (QR / ceklist / selfie), Kalender Akademik.

func init() {
	moduleSchemas = append(moduleSchemas, `
CREATE TABLE IF NOT EXISTS student_profiles (
  siswa_id       INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  nisn TEXT, nik TEXT, jk TEXT, tempat_lahir TEXT, tgl_lahir TEXT, agama TEXT,
  alamat TEXT, no_hp TEXT, nama_ayah TEXT, pekerjaan_ayah TEXT, nama_ibu TEXT, pekerjaan_ibu TEXT,
  nama_wali TEXT, hp_ortu TEXT, asal_sekolah TEXT, tgl_masuk TEXT, foto TEXT, catatan TEXT,
  updated_at TEXT
);
CREATE INDEX IF NOT EXISTS idx_profiles_nisn ON student_profiles(nisn);

CREATE TABLE IF NOT EXISTS daily_attendance (
  siswa_id   INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  tanggal    TEXT NOT NULL,
  status     TEXT NOT NULL CHECK (status IN ('hadir','terlambat','izin','sakit','alfa')),
  jam_masuk  TEXT,
  jam_pulang TEXT,
  metode     TEXT,
  lat REAL, lng REAL, foto TEXT,
  keterangan TEXT,
  petugas_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
  PRIMARY KEY (siswa_id, tanggal)
);
CREATE INDEX IF NOT EXISTS idx_daily_tgl ON daily_attendance(tanggal);

CREATE TABLE IF NOT EXISTS events (
  id         INTEGER PRIMARY KEY,
  judul      TEXT NOT NULL,
  mulai      TEXT NOT NULL,
  selesai    TEXT NOT NULL,
  kategori   TEXT NOT NULL DEFAULT 'kegiatan' CHECK (kategori IN ('kegiatan','libur','ujian','rapat','lainnya')),
  keterangan TEXT,
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_events_mulai ON events(mulai);
`)
	moduleRoutes = append(moduleRoutes, func(a *App, m *http.ServeMux) {
		m.HandleFunc("GET /api/bukuinduk", a.mod("bukuinduk", "admin,guru", apiBukuIndukList))
		m.HandleFunc("GET /api/bukuinduk/{id}", a.mod("bukuinduk", "any", apiBukuIndukGet))
		m.HandleFunc("PUT /api/bukuinduk/{id}", a.mod("bukuinduk", "admin", apiBukuIndukSave))

		m.HandleFunc("GET /api/kartu", a.mod("absensi", "admin", a.apiKartu)) // token kartu: admin & staf absensi saja
		m.HandleFunc("POST /api/absen/scan", a.mod("absensi", "admin,guru", a.apiAbsenScan))
		m.HandleFunc("GET /api/absen/harian", a.mod("absensi", "admin,guru", apiAbsenHarian))
		m.HandleFunc("PUT /api/absen/harian", a.mod("absensi", "admin,guru", apiAbsenHarianSave))
		m.HandleFunc("GET /api/absen/rekap", a.mod("absensi", "admin,guru", apiAbsenRekap))
		m.HandleFunc("GET /api/absen/me", a.mod("absensi", "siswa", a.apiAbsenMe))
		m.HandleFunc("POST /api/absen/selfie", a.mod("absensi", "siswa", a.apiAbsenSelfie))

		m.HandleFunc("GET /api/kalender", a.mod("kalender", "any", apiKalender))
		m.HandleFunc("POST /api/kalender", a.mod("kalender", "admin", apiSaveEvent))
		m.HandleFunc("PUT /api/kalender/{id}", a.mod("kalender", "admin", apiSaveEvent))
		m.HandleFunc("DELETE /api/kalender/{id}", a.mod("kalender", "admin", apiDelete("events")))
	})
}

// ---------------------------------------------------------------------
// Absensi harian
// ---------------------------------------------------------------------
func localDate() string  { return time.Now().Format("2006-01-02") }
func localClock() string { return time.Now().Format("15:04") }

func (a *App) statusMasuk() string {
	cfg := getConfig(a.db, "absen")
	batas, _ := cfg["jam_masuk"].(string)
	if batas != "" && localClock() > batas {
		return "terlambat"
	}
	return "hadir"
}

type absenResult struct {
	SiswaID   int64  `json:"siswa_id"`
	Nama      string `json:"nama"`
	Kelas     string `json:"kelas"`
	Foto      any    `json:"foto"`
	Status    string `json:"status"`
	JamMasuk  any    `json:"jam_masuk"`
	JamPulang any    `json:"jam_pulang"`
	Aksi      string `json:"aksi"` // masuk | pulang | sudah
}

// recordPresence mencatat kedatangan/kepulangan. mode: "masuk", "pulang", atau "auto".
func (a *App) recordPresence(siswaID int64, mode, metode string, petugas any, lat, lng any, foto any) (*absenResult, error) {
	r := &absenResult{SiswaID: siswaID}
	var kelas sql.NullString
	var fotoProfil sql.NullString
	var role, status string
	if err := a.db.QueryRow(`SELECT u.nama_lengkap, u.role, u.status, k.nama, p.foto FROM users u LEFT JOIN classes k ON k.id=u.kelas_id
		LEFT JOIN student_profiles p ON p.siswa_id=u.id WHERE u.id=?`, siswaID).Scan(&r.Nama, &role, &status, &kelas, &fotoProfil); err != nil || role != "siswa" {
		return nil, bad("Kartu tidak dikenali.")
	}
	if status != "aktif" {
		return nil, bad("Akun siswa ini nonaktif.")
	}
	r.Kelas = kelas.String
	if fotoProfil.Valid {
		r.Foto = fotoProfil.String
	}
	tgl, jam := localDate(), localClock()
	var st string
	var jm, jp sql.NullString
	err := a.db.QueryRow(`SELECT status, jam_masuk, jam_pulang FROM daily_attendance WHERE siswa_id=? AND tanggal=?`, siswaID, tgl).Scan(&st, &jm, &jp)
	switch {
	case err == sql.ErrNoRows && mode != "pulang":
		st = a.statusMasuk()
		res, err := a.db.Exec(`INSERT INTO daily_attendance (siswa_id, tanggal, status, jam_masuk, metode, lat, lng, foto, petugas_id) VALUES (?,?,?,?,?,?,?,?,?)
			ON CONFLICT (siswa_id, tanggal) DO NOTHING`, siswaID, tgl, st, jam, metode, lat, lng, foto, petugas)
		if err != nil {
			return nil, err
		}
		if n, _ := res.RowsAffected(); n == 0 { // scan ganda bersamaan: catatan sudah dibuat permintaan lain
			r.Aksi, r.Status, r.JamMasuk = "sudah", st, jam
			return r, nil
		}
		r.Aksi, r.Status, r.JamMasuk = "masuk", st, jam
	case err == sql.ErrNoRows && mode == "pulang":
		return nil, bad(r.Nama + " belum tercatat masuk hari ini.")
	case err != nil:
		return nil, err
	case (mode == "pulang" || mode == "auto") && !jp.Valid && (st == "hadir" || st == "terlambat"):
		if mode == "auto" && jm.Valid {
			// mode otomatis: catat pulang hanya jika minimal 60 menit setelah masuk
			t1, _ := time.Parse("15:04", jm.String)
			t2, _ := time.Parse("15:04", jam)
			if t2.Sub(t1) < time.Hour {
				r.Aksi, r.Status, r.JamMasuk = "sudah", st, jm.String
				return r, nil
			}
		}
		if _, err := a.db.Exec(`UPDATE daily_attendance SET jam_pulang=? WHERE siswa_id=? AND tanggal=?`, jam, siswaID, tgl); err != nil {
			return nil, err
		}
		r.Aksi, r.Status, r.JamMasuk, r.JamPulang = "pulang", st, jm.String, jam
	default:
		r.Aksi, r.Status = "sudah", st
		if jm.Valid {
			r.JamMasuk = jm.String
		}
		if jp.Valid {
			r.JamPulang = jp.String
		}
	}
	if r.Aksi == "masuk" || r.Aksi == "pulang" { // notifikasi WhatsApp ke orang tua (bila diaktifkan)
		nama, kelas, hp := a.siswaWA(siswaID)
		a.waQueue(r.Aksi, fmt.Sprintf("%s:%d:%s", r.Aksi, siswaID, tgl), hp,
			map[string]string{"nama": nama, "kelas": kelas, "jam": jam, "status": r.Status, "tanggal": tgl}, false)
	}
	return r, nil
}

func (a *App) apiAbsenScan(c *Ctx) error {
	var in struct {
		Token string `json:"token"`
		Mode  string `json:"mode"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	id := a.cardUser(in.Token)
	if id == 0 {
		return bad("QR tidak valid atau bukan kartu siswa sekolah ini.")
	}
	if in.Mode != "masuk" && in.Mode != "pulang" {
		in.Mode = "auto"
	}
	r, err := a.recordPresence(id, in.Mode, "qr", c.U.ID, nil, nil, nil)
	if err != nil {
		return err
	}
	return c.JSON(r)
}

// guruKelas: guru tanpa akses kelola modul hanya boleh mengelola kelas yang ia ajar.
func (c *Ctx) guruKelas(mod string, kelasID int64) error {
	if c.U.Has(mod) {
		return nil
	}
	var n int
	c.A.db.QueryRow(`SELECT COUNT(*) FROM courses WHERE guru_id=? AND kelas_id=?`, c.U.ID, kelasID).Scan(&n)
	if n == 0 {
		return apiErr{403, "Anda hanya bisa mengelola kelas yang Anda ajar."}
	}
	return nil
}

func apiAbsenHarian(c *Ctx) error {
	kid := c.QInt("kelas_id")
	if err := c.guruKelas("absensi", kid); err != nil {
		return err
	}
	tgl, err := cleanDate(c.R.URL.Query().Get("tanggal"))
	if err != nil {
		return err
	}
	rows, err := c.list(`SELECT u.id, u.username, u.nama_lengkap, d.status, d.jam_masuk, d.jam_pulang, d.metode, d.keterangan, d.foto AS foto_absen
		FROM users u LEFT JOIN daily_attendance d ON d.siswa_id=u.id AND d.tanggal=?
		WHERE u.role='siswa' AND u.status='aktif' AND u.kelas_id=? ORDER BY u.nama_lengkap`, tgl, kid)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

func apiAbsenHarianSave(c *Ctx) error {
	var in struct {
		KelasID int64  `json:"kelas_id"`
		Tanggal string `json:"tanggal"`
		Items   []struct {
			SiswaID    int64   `json:"siswa_id"`
			Status     string  `json:"status"`
			Keterangan *string `json:"keterangan"`
		} `json:"items"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	tgl, err := cleanDate(in.Tanggal)
	if err != nil {
		return err
	}
	if err := c.guruKelas("absensi", in.KelasID); err != nil {
		return err
	}
	tx, err := c.A.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var alfa []int64 // siswa yang benar-benar dicatat alfa (hanya mereka yang dikabari)
	for _, it := range in.Items {
		switch it.Status {
		case "hadir", "terlambat", "izin", "sakit", "alfa":
		case "":
			tx.Exec(`DELETE FROM daily_attendance WHERE siswa_id=? AND tanggal=? AND metode='manual'`, it.SiswaID, tgl)
			continue
		default:
			return bad("Status tidak valid.")
		}
		var ok int
		tx.QueryRow(`SELECT COUNT(*) FROM users WHERE id=? AND role='siswa' AND kelas_id=?`, it.SiswaID, in.KelasID).Scan(&ok)
		if ok == 0 {
			continue
		}
		if it.Status == "alfa" {
			alfa = append(alfa, it.SiswaID)
		}
		if _, err := tx.Exec(`INSERT INTO daily_attendance (siswa_id, tanggal, status, metode, keterangan, petugas_id) VALUES (?,?,?,'manual',?,?)
			ON CONFLICT (siswa_id, tanggal) DO UPDATE SET status=excluded.status, keterangan=excluded.keterangan, petugas_id=excluded.petugas_id`,
			it.SiswaID, tgl, it.Status, str(it.Keterangan), c.U.ID); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	// ketidakhadiran tanpa keterangan hari ini → kabari orang tua (bila diaktifkan); tanggal lampau tidak dikirim
	if tgl == localDate() {
		for _, sid := range alfa {
			{
				nama, kelas, hp := c.A.siswaWA(sid)
				c.A.waQueue("alfa", fmt.Sprintf("alfa:%d:%s", sid, tgl), hp,
					map[string]string{"nama": nama, "kelas": kelas, "tanggal": fmtTanggal(tgl), "status": "alfa"}, false)
			}
		}
	}
	return c.OK()
}

func monthRange(s string) (string, string, error) {
	t, err := time.Parse("2006-01", strings.TrimSpace(s))
	if err != nil {
		return "", "", bad("Bulan tidak valid.")
	}
	return t.Format("2006-01-02"), t.AddDate(0, 1, -1).Format("2006-01-02"), nil
}

func apiAbsenRekap(c *Ctx) error {
	kid := c.QInt("kelas_id")
	if err := c.guruKelas("absensi", kid); err != nil {
		return err
	}
	from, to, err := monthRange(c.R.URL.Query().Get("bulan"))
	if err != nil {
		return err
	}
	st, err := c.list(`SELECT u.id, u.username, u.nama_lengkap FROM users u WHERE u.role='siswa' AND `+memberOf("?")+` ORDER BY u.nama_lengkap`, kid, kid)
	if err != nil {
		return err
	}
	rec, err := c.list(`SELECT d.siswa_id, d.tanggal, d.status FROM daily_attendance d JOIN users u ON u.id=d.siswa_id
		WHERE `+memberOf("?")+` AND d.tanggal BETWEEN ? AND ?`, kid, kid, from, to)
	if err != nil {
		return err
	}
	ev, err := c.list(`SELECT mulai, selesai, judul FROM events WHERE kategori='libur' AND selesai>=? AND mulai<=?`, from, to)
	if err != nil {
		return err
	}
	return c.JSON(map[string]any{"students": st, "records": rec, "libur": ev, "dari": from, "sampai": to})
}

func (a *App) apiAbsenMe(c *Ctx) error {
	from, to, err := monthRange(c.R.URL.Query().Get("bulan"))
	if err != nil {
		return err
	}
	rec, err := c.list(`SELECT tanggal, status, jam_masuk, jam_pulang, metode, keterangan FROM daily_attendance
		WHERE siswa_id=? AND tanggal BETWEEN ? AND ? ORDER BY tanggal DESC`, c.U.ID, from, to)
	if err != nil {
		return err
	}
	var today map[string]any
	t, _ := c.list(`SELECT status, jam_masuk, jam_pulang FROM daily_attendance WHERE siswa_id=? AND tanggal=?`, c.U.ID, localDate())
	if len(t) > 0 {
		today = t[0]
	}
	return c.JSON(map[string]any{"records": rec, "today": today, "token": a.cardToken(c.U.ID), "config": getConfig(a.db, "absen")})
}

func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const r = 6371000.0
	p1, p2 := lat1*math.Pi/180, lat2*math.Pi/180
	dp, dl := (lat2-lat1)*math.Pi/180, (lng2-lng1)*math.Pi/180
	h := math.Sin(dp/2)*math.Sin(dp/2) + math.Cos(p1)*math.Cos(p2)*math.Sin(dl/2)*math.Sin(dl/2)
	return 2 * r * math.Asin(math.Sqrt(h))
}

func (a *App) apiAbsenSelfie(c *Ctx) error {
	cfg := getConfig(a.db, "absen")
	if cfg["selfie"] != true {
		return bad("Absen selfie belum diaktifkan admin.")
	}
	c.longBody()
	var in struct {
		Foto string   `json:"foto"`
		Lat  *float64 `json:"lat"`
		Lng  *float64 `json:"lng"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	if in.Lat == nil || in.Lng == nil {
		return bad("Lokasi tidak terbaca. Izinkan akses lokasi di browser.")
	}
	if slat, ok := cfg["lat"].(float64); ok {
		slng, _ := cfg["lng"].(float64)
		rad, _ := cfg["radius_m"].(float64)
		if rad <= 0 {
			rad = 150
		}
		if d := haversine(slat, slng, *in.Lat, *in.Lng); d > rad {
			return bad("Anda berada sekitar " + itoa(int(d)) + " m dari sekolah (batas " + itoa(int(rad)) + " m). Absen hanya bisa di area sekolah.")
		}
	}
	var jp sql.NullString
	if c.A.db.QueryRow(`SELECT jam_pulang FROM daily_attendance WHERE siswa_id=? AND tanggal=?`, c.U.ID, localDate()).Scan(&jp) == nil && jp.Valid {
		return bad("Absen masuk dan pulang hari ini sudah tercatat.")
	}
	data, err := decodeDataURL(in.Foto)
	if err != nil || len(data) < 1000 {
		return bad("Foto selfie tidak terbaca.")
	}
	if len(data) > 2<<20 {
		return bad("Foto terlalu besar.")
	}
	url, err := a.saveUpload(data, true)
	if err != nil {
		return err
	}
	r, err := a.recordPresence(c.U.ID, "auto", "selfie", nil, *in.Lat, *in.Lng, url)
	if err != nil {
		return err
	}
	return c.JSON(r)
}

func itoa(n int) string { return strconv.Itoa(n) }
