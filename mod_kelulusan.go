package main

import (
	"database/sql"
	"strings"
	"time"
)

// Pengumuman kelulusan: data status per siswa (admin) dan halaman cek publik (NISN + tanggal lahir).

// ---------------------------------------------------------------------
// Pengumuman kelulusan
// ---------------------------------------------------------------------
func apiGradList(c *Ctx) error {
	q := `SELECT u.id, u.username, u.nama_lengkap, k.nama AS kelas_nama, u.kelas_id, sp.nisn, sp.tgl_lahir, g.status, g.catatan
		FROM users u LEFT JOIN classes k ON k.id=u.kelas_id LEFT JOIN student_profiles sp ON sp.siswa_id=u.id LEFT JOIN graduation g ON g.siswa_id=u.id
		WHERE u.role='siswa'`
	args := []any{}
	if kid := c.QInt("kelas_id"); kid > 0 {
		q += ` AND u.kelas_id=?`
		args = append(args, kid)
	}
	rows, err := c.list(q+` ORDER BY k.nama, u.nama_lengkap`, args...)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

func apiGradSave(c *Ctx) error {
	var in struct {
		Items []struct {
			SiswaID int64   `json:"siswa_id"`
			Status  string  `json:"status"`
			Catatan *string `json:"catatan"`
		} `json:"items"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	tx, err := c.A.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, it := range in.Items {
		switch it.Status {
		case "":
			tx.Exec(`DELETE FROM graduation WHERE siswa_id=?`, it.SiswaID)
			continue
		case "lulus", "tidak_lulus", "ditunda":
		default:
			return bad("Status kelulusan tidak valid.")
		}
		var role string
		if tx.QueryRow(`SELECT role FROM users WHERE id=?`, it.SiswaID).Scan(&role); role != "siswa" {
			continue
		}
		if _, err := tx.Exec(`INSERT INTO graduation (siswa_id, status, catatan, updated_at) VALUES (?,?,?,?)
			ON CONFLICT (siswa_id) DO UPDATE SET status=excluded.status, catatan=excluded.catatan, updated_at=excluded.updated_at`,
			it.SiswaID, it.Status, str(it.Catatan), nowStr()); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.OK()
}

func apiGradInfo(c *Ctx) error {
	cfg := getConfig(c.A.db, "kelulusan")
	var nama string
	c.A.db.QueryRow(`SELECT nama_sekolah FROM settings WHERE id=1`).Scan(&nama)
	return c.JSON(map[string]any{"judul": cfg["judul"], "buka": cfg["buka"], "sekarang": nowStr(), "nama_sekolah": nama})
}

func apiGradCheck(c *Ctx) error {
	ipKey := "kl:" + limitKey(clientIP(c.R))
	if c.A.limit.blockedN(ipKey, 30) {
		return apiErr{429, "Terlalu banyak percobaan. Coba lagi 10 menit lagi."}
	}
	cfg := getConfig(c.A.db, "kelulusan")
	buka, _ := cfg["buka"].(string)
	t, err := time.Parse(time.RFC3339, buka)
	if buka == "" || err != nil {
		return bad("Pengumuman kelulusan belum dijadwalkan.")
	}
	if time.Now().Before(t) {
		return bad("Pengumuman dibuka " + t.Local().Format("02/01/2006 pukul 15:04") + ".")
	}
	var in struct {
		NISN     string `json:"nisn"`
		TglLahir string `json:"tgl_lahir"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	tgl, err := cleanDate(in.TglLahir)
	if err != nil || strings.TrimSpace(in.NISN) == "" {
		return bad("Isi NISN dan tanggal lahir.")
	}
	// kunci per NISN: mencegah menebak tanggal lahir satu siswa dari banyak IP
	nKey := "kn:" + strings.TrimSpace(in.NISN)
	if c.A.limit.blockedN(nKey, 5) {
		return apiErr{429, "Terlalu banyak percobaan untuk NISN ini. Coba lagi 10 menit lagi."}
	}
	var nama, status string
	var kelas, catatan sql.NullString
	err = c.A.db.QueryRow(`SELECT u.nama_lengkap, k.nama, g.status, g.catatan FROM student_profiles sp JOIN users u ON u.id=sp.siswa_id
		LEFT JOIN classes k ON k.id=u.kelas_id JOIN graduation g ON g.siswa_id=u.id WHERE sp.nisn=? AND sp.tgl_lahir=?`,
		strings.TrimSpace(in.NISN), tgl).Scan(&nama, &kelas, &status, &catatan)
	if err != nil {
		c.A.limit.fail(ipKey)
		c.A.limit.fail(nKey)
		return bad("Data tidak ditemukan. Periksa kembali NISN dan tanggal lahir.")
	}
	pesan := cfg["pesan"]
	if status != "lulus" {
		pesan = cfg["pesan_tidak"]
	}
	return c.JSON(map[string]any{"nama": nama, "kelas": kelas.String, "status": status, "catatan": catatan.String, "pesan": pesan})
}
