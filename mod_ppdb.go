package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// PPDB online: formulir publik, cek status (nomor + PIN), verifikasi, jadikan siswa.

// ---------------------------------------------------------------------
// PPDB
// ---------------------------------------------------------------------
func ppdbOpen(cfg map[string]any) (bool, string) {
	now := time.Now()
	if s, _ := cfg["buka"].(string); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil && now.Before(t) {
			return false, "Pendaftaran dibuka " + t.Local().Format("02/01/2006 15:04") + "."
		}
	} else {
		return false, "Pendaftaran belum dibuka."
	}
	if s, _ := cfg["tutup"].(string); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil && now.After(t) {
			return false, "Pendaftaran sudah ditutup."
		}
	}
	return true, ""
}

func apiPPDBInfo(c *Ctx) error {
	cfg := getConfig(c.A.db, "ppdb")
	open, msg := ppdbOpen(cfg)
	var n int
	c.A.db.QueryRow(`SELECT COUNT(*) FROM ppdb_applicants`).Scan(&n)
	var nama string
	c.A.db.QueryRow(`SELECT nama_sekolah FROM settings WHERE id=1`).Scan(&nama)
	return c.JSON(map[string]any{"buka": open, "pesan": msg, "jadwal_buka": cfg["buka"], "jadwal_tutup": cfg["tutup"], "tahun_ajaran": cfg["tahun_ajaran"],
		"kuota": cfg["kuota"], "jalur": cfg["jalur"], "info": cfg["info"], "pendaftar": n, "nama_sekolah": nama})
}

var ppdbFields = []string{"nama", "nisn", "nik", "jk", "tempat_lahir", "tgl_lahir", "agama", "alamat", "asal_sekolah", "nama_ayah", "nama_ibu", "hp", "email", "jalur"}

func (a *App) apiPPDBDaftar(c *Ctx) error {
	var in struct {
		Data    map[string]string `json:"data"`
		Berkas  map[string]string `json:"berkas"`  // {kk|akta|foto|lain: dataURL}
		Website string            `json:"website"` // honeypot
	}
	// batas per perangkat & masa pendaftaran dicek SEBELUM membaca isi (hingga 10 MB), agar pengirim
	// yang sudah diblokir tidak bisa menghabiskan memori server
	if err := c.publicGuard("ppdb:", 3, ""); err != nil {
		return err
	}
	cfg := getConfig(a.db, "ppdb")
	if open, msg := ppdbOpen(cfg); !open {
		return bad(msg)
	}
	c.longBody()
	if err := c.bindLarge(&in, 10<<20); err != nil {
		return err
	}
	if strings.TrimSpace(in.Website) != "" { // honeypot
		return bad("Pengiriman ditolak.")
	}
	d := map[string]any{}
	for _, f := range ppdbFields {
		r := []rune(strings.TrimSpace(in.Data[f]))
		if len(r) > 300 {
			r = r[:300]
		}
		if len(r) == 0 {
			d[f] = nil
		} else {
			d[f] = string(r)
		}
	}
	for _, f := range []string{"nama", "tgl_lahir", "hp"} {
		if d[f] == nil {
			return bad("Nama, tanggal lahir, dan nomor HP wajib diisi.")
		}
	}
	tgl, err := cleanDate(d["tgl_lahir"].(string))
	if err != nil {
		return bad("Tanggal lahir tidak valid.")
	}
	d["tgl_lahir"] = tgl
	if jk, _ := d["jk"].(string); jk != "" && jk != "L" && jk != "P" {
		d["jk"] = nil
	}
	berkas := map[string]string{}
	for _, k := range []string{"kk", "akta", "foto", "lain"} {
		s := in.Berkas[k]
		if s == "" {
			continue
		}
		data, err := decodeDataURL(s)
		if err != nil || len(data) > 2<<20 {
			return bad("Berkas " + k + " tidak valid atau lebih dari 2 MB.")
		}
		url, err := a.saveUpload(data, false)
		if err != nil {
			return err
		}
		berkas[k] = url
	}
	bb, _ := json.Marshal(berkas)
	pin := randDigits(8)
	phs, err := hashPassword(pin)
	if err != nil {
		return err
	}
	ph := []byte(phs)
	year := time.Now().Format("2006")
	if b, _ := cfg["buka"].(string); len(b) >= 4 {
		if t, err := time.Parse(time.RFC3339, b); err == nil {
			year = t.Local().Format("2006")
		} else {
			year = b[:4]
		}
	}
	// NISN/NIK: format benar & belum terdaftar (kecuali pendaftaran lama yang ditolak)
	for _, f := range []struct {
		key, label string
		n          int
	}{{"nisn", "NISN", 10}, {"nik", "NIK", 16}} {
		v, _ := d[f.key].(string)
		if v == "" {
			continue
		}
		v = onlyDigits(v)
		if len(v) != f.n {
			return bad(f.label + " harus " + fmt.Sprint(f.n) + " digit angka.")
		}
		d[f.key] = v
		var nd string
		if a.db.QueryRow(`SELECT no_daftar FROM ppdb_applicants WHERE `+f.key+`=? AND status<>'ditolak' LIMIT 1`, v).Scan(&nd) == nil {
			return bad(f.label + " ini sudah terdaftar dengan nomor " + nd + ". Cek status pendaftaran dengan nomor tersebut.")
		}
	}
	var no string
	for i := 0; i < 5; i++ {
		var n int // urutan terbesar + 1 (aman walau ada pendaftar yang dihapus)
		a.db.QueryRow(`SELECT COALESCE(MAX(CAST(substr(no_daftar, 11) AS INTEGER)), 0) FROM ppdb_applicants WHERE no_daftar >= ? AND no_daftar < ?`, "PPDB-"+year+"-", "PPDB-"+year+"-~").Scan(&n)
		no = fmt.Sprintf("PPDB-%s-%04d", year, n+1)
		_, err = a.db.Exec(`INSERT INTO ppdb_applicants (no_daftar, pin_hash, nama, nisn, nik, jk, tempat_lahir, tgl_lahir, agama, alamat, asal_sekolah, nama_ayah, nama_ibu, hp, email, jalur, berkas, created_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, no, string(ph), d["nama"], d["nisn"], d["nik"], d["jk"], d["tempat_lahir"], d["tgl_lahir"], d["agama"], d["alamat"],
			d["asal_sekolah"], d["nama_ayah"], d["nama_ibu"], d["hp"], d["email"], d["jalur"], string(bb), nowStr())
		if err == nil || !strings.Contains(err.Error(), "UNIQUE") {
			break
		}
	}
	if err != nil {
		return err
	}
	if hp, _ := d["hp"].(string); hp != "" {
		a.waQueue("ppdb", "ppdb:"+no, hp, map[string]string{"nama": d["nama"].(string), "no_daftar": no, "pin": pin}, true)
	}
	return c.JSON(map[string]string{"no_daftar": no, "pin": pin, "nama": d["nama"].(string)})
}

func apiPPDBCek(c *Ctx) error {
	var in struct {
		NoDaftar string `json:"no_daftar"`
		PIN      string `json:"pin"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	ipKey := "pc:" + limitKey(clientIP(c.R))
	noKey := "pn:" + strings.ToUpper(strings.TrimSpace(in.NoDaftar))
	if c.A.limit.blockedN(ipKey, 20) || c.A.limit.blockedN(noKey, 5) {
		return apiErr{429, "Terlalu banyak percobaan. Coba lagi 10 menit lagi."}
	}
	var nama, status, hash, jalur string
	var catatan sql.NullString
	err := c.A.db.QueryRow(`SELECT nama, status, pin_hash, COALESCE(jalur,''), catatan FROM ppdb_applicants WHERE no_daftar=?`, strings.ToUpper(strings.TrimSpace(in.NoDaftar))).
		Scan(&nama, &status, &hash, &jalur, &catatan)
	if err != nil || checkPassword(hash, strings.TrimSpace(in.PIN)) != nil {
		c.A.limit.fail(ipKey)
		c.A.limit.fail(noKey)
		return bad("Nomor pendaftaran atau PIN salah.")
	}
	return c.JSON(map[string]any{"nama": nama, "status": status, "jalur": jalur, "catatan": catatan.String})
}

func apiPPDBList(c *Ctx) error {
	q := `SELECT id, no_daftar, nama, nisn, nik, jk, tempat_lahir, tgl_lahir, agama, alamat, asal_sekolah, nama_ayah, nama_ibu, hp, email, jalur, berkas, status, catatan, user_id, created_at
		FROM ppdb_applicants`
	where, args := ` WHERE 1=1`, []any{}
	if s := c.R.URL.Query().Get("status"); s != "" {
		where += ` AND status=?`
		args = append(args, s)
	}
	if s := strings.TrimSpace(c.R.URL.Query().Get("q")); s != "" {
		where += ` AND (nama LIKE ? OR no_daftar LIKE ? OR nisn LIKE ? OR asal_sekolah LIKE ? OR hp LIKE ?)`
		like := "%" + s + "%"
		args = append(args, like, like, like, like, like)
	}
	limit, offset := pageParams(c, 300)
	var total int
	c.A.db.QueryRow(`SELECT COUNT(*) FROM ppdb_applicants`+where, args...).Scan(&total)
	c.W.Header().Set("X-Total-Count", strconv.Itoa(total))
	rows, err := c.list(q+where+` ORDER BY created_at DESC LIMIT ? OFFSET ?`, append(args, limit, offset)...)
	if err != nil {
		return err
	}
	for _, r := range rows {
		if s, ok := r["berkas"].(string); ok && json.Valid([]byte(s)) {
			r["berkas"] = json.RawMessage(s)
		}
	}
	return c.JSON(rows)
}

func apiPPDBUpdate(c *Ctx) error {
	var in struct {
		Status  string  `json:"status"`
		Catatan *string `json:"catatan"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	switch in.Status {
	case "baru", "verifikasi", "diterima", "cadangan", "ditolak":
	default:
		return bad("Status tidak valid.")
	}
	if _, err := c.A.db.Exec(`UPDATE ppdb_applicants SET status=?, catatan=?, updated_at=? WHERE id=?`, in.Status, str(in.Catatan), nowStr(), c.ID("id")); err != nil {
		return err
	}
	return c.OK()
}

// apiPPDBToStudent membuat akun siswa + biodata Buku Induk dari data pendaftar yang diterima.
func apiPPDBToStudent(c *Ctx) error {
	var in struct {
		KelasID int64 `json:"kelas_id"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	id := c.ID("id")
	var r struct {
		no, nama, status                                                  string
		nisn, nik, jk, tl, tgl, agama, alamat, asal, ayah, ibu, hp, email sql.NullString
		userID                                                            sql.NullInt64
	}
	err := c.A.db.QueryRow(`SELECT no_daftar, nama, status, nisn, nik, jk, tempat_lahir, tgl_lahir, agama, alamat, asal_sekolah, nama_ayah, nama_ibu, hp, email, user_id FROM ppdb_applicants WHERE id=?`, id).
		Scan(&r.no, &r.nama, &r.status, &r.nisn, &r.nik, &r.jk, &r.tl, &r.tgl, &r.agama, &r.alamat, &r.asal, &r.ayah, &r.ibu, &r.hp, &r.email, &r.userID)
	if err != nil {
		return notFound()
	}
	if r.userID.Valid {
		return bad("Pendaftar ini sudah dibuatkan akun siswa.")
	}
	if r.status != "diterima" {
		return bad("Ubah status menjadi Diterima terlebih dahulu.")
	}
	username := r.no
	if r.nisn.Valid && r.nisn.String != "" {
		username = r.nisn.String
	}
	pw := randomPassword(8)
	h, err := hashPassword(pw)
	if err != nil {
		return err
	}
	var kelas any
	if in.KelasID > 0 {
		kelas = in.KelasID
	}
	tx, err := c.A.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`INSERT INTO users (username, password_hash, nama_lengkap, role, email, kelas_id, created_at) VALUES (?,?,?,'siswa',?,?,?)`,
		username, h, r.nama, r.email, kelas, nowStr())
	if err != nil {
		return err
	}
	uid, _ := res.LastInsertId()
	if _, err := tx.Exec(`INSERT INTO student_profiles (siswa_id, nisn, nik, jk, tempat_lahir, tgl_lahir, agama, alamat, nama_ayah, nama_ibu, hp_ortu, asal_sekolah, tgl_masuk, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, uid, r.nisn, r.nik, r.jk, r.tl, r.tgl, r.agama, r.alamat, r.ayah, r.ibu, r.hp, r.asal, localDate(), nowStr()); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE ppdb_applicants SET user_id=?, updated_at=? WHERE id=?`, uid, nowStr(), id); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.JSON(map[string]string{"username": username, "password": pw})
}
