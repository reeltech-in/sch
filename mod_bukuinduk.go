package main

import (
	"strings"
)

// Buku Induk siswa dan kartu QR siswa.

// ---------------------------------------------------------------------
// Buku Induk
// ---------------------------------------------------------------------
var profileFields = []string{"nisn", "nik", "jk", "tempat_lahir", "tgl_lahir", "agama", "alamat", "no_hp", "nama_ayah",
	"pekerjaan_ayah", "nama_ibu", "pekerjaan_ibu", "nama_wali", "hp_ortu", "asal_sekolah", "tgl_masuk", "foto", "catatan"}

func apiBukuIndukList(c *Ctx) error {
	q := `SELECT u.id, u.username, u.nama_lengkap, u.status, u.kelas_id, k.nama AS kelas_nama,
		p.nisn, p.jk, p.tempat_lahir, p.tgl_lahir, p.nama_ayah, p.nama_ibu, p.hp_ortu, p.foto,
		(p.siswa_id IS NOT NULL) AS terisi
		FROM users u LEFT JOIN classes k ON k.id=u.kelas_id LEFT JOIN student_profiles p ON p.siswa_id=u.id
		WHERE u.role='siswa'`
	args := []any{}
	if kid := c.QInt("kelas_id"); kid > 0 {
		q += ` AND u.kelas_id=?`
		args = append(args, kid)
	}
	if !c.U.Has("bukuinduk") { // guru biasa: hanya siswa di kelas yang ia ajar (data pribadi, UU PDP)
		q += ` AND u.kelas_id IN (SELECT kelas_id FROM courses WHERE guru_id=?)`
		args = append(args, c.U.ID)
	}
	rows, err := c.list(q+` ORDER BY k.nama, u.nama_lengkap`, args...)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

func apiBukuIndukGet(c *Ctx) error {
	id := c.ID("id")
	if c.U.Role == "siswa" && id != c.U.ID {
		return forbidden()
	}
	if c.U.Role == "guru" && !c.U.Has("bukuinduk") {
		var n int
		c.A.db.QueryRow(`SELECT COUNT(*) FROM users u JOIN courses co ON co.kelas_id=u.kelas_id WHERE u.id=? AND co.guru_id=?`, id, c.U.ID).Scan(&n)
		if n == 0 {
			return forbidden()
		}
	}
	rows, err := c.list(`SELECT u.id, u.username, u.nama_lengkap, u.email, u.status, u.kelas_id, k.nama AS kelas_nama, p.*
		FROM users u LEFT JOIN classes k ON k.id=u.kelas_id LEFT JOIN student_profiles p ON p.siswa_id=u.id
		WHERE u.id=? AND u.role='siswa'`, id)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return notFound()
	}
	return c.JSON(rows[0])
}

func apiBukuIndukSave(c *Ctx) error {
	id := c.ID("id")
	var role string
	if c.A.db.QueryRow(`SELECT role FROM users WHERE id=?`, id).Scan(&role) != nil || role != "siswa" {
		return notFound()
	}
	in := map[string]*string{}
	if err := c.Bind(&in); err != nil {
		return err
	}
	if err := c.cekIdentitasSiswa(id, in); err != nil {
		return err
	}
	cols, vals := []string{"siswa_id"}, []any{id}
	for _, f := range profileFields {
		v := str(in[f])
		switch f {
		case "tgl_lahir", "tgl_masuk":
			if v != nil {
				d, err := cleanDate(v.(string))
				if err != nil {
					return bad("Format tanggal tidak valid.")
				}
				v = d
			}
		case "foto":
			if v != nil && !fileNameRe.MatchString(strings.TrimPrefix(v.(string), "/files/")) {
				return bad("Foto tidak valid.")
			}
		case "jk":
			if v != nil && v != "L" && v != "P" {
				return bad("Jenis kelamin harus L atau P.")
			}
		}
		cols = append(cols, f)
		vals = append(vals, v)
	}
	cols = append(cols, "updated_at")
	vals = append(vals, nowStr())
	set := []string{}
	for _, cn := range cols[1:] {
		set = append(set, cn+"=excluded."+cn)
	}
	q := `INSERT INTO student_profiles (` + strings.Join(cols, ",") + `) VALUES (?` + strings.Repeat(",?", len(cols)-1) + `)
		ON CONFLICT (siswa_id) DO UPDATE SET ` + strings.Join(set, ",")
	if _, err := c.A.db.Exec(q, vals...); err != nil {
		return err
	}
	return c.OK()
}

// ---------------------------------------------------------------------
// Kartu QR siswa
// ---------------------------------------------------------------------
func (a *App) apiKartu(c *Ctx) error {
	q := `SELECT u.id, u.username, u.nama_lengkap, k.nama AS kelas_nama, p.nisn, p.foto
		FROM users u LEFT JOIN classes k ON k.id=u.kelas_id LEFT JOIN student_profiles p ON p.siswa_id=u.id
		WHERE u.role='siswa' AND u.status='aktif'`
	args := []any{}
	if kid := c.QInt("kelas_id"); kid > 0 {
		q += ` AND u.kelas_id=?`
		args = append(args, kid)
	}
	rows, err := c.list(q+` ORDER BY k.nama, u.nama_lengkap`, args...)
	if err != nil {
		return err
	}
	for _, r := range rows {
		r["token"] = a.cardToken(r["id"].(int64))
	}
	return c.JSON(rows)
}
