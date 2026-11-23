package main

import (
	"strings"
)

// Buku tamu: kiosk publik dengan foto, daftar kunjungan untuk admin.

// ---------------------------------------------------------------------
// Buku tamu
// ---------------------------------------------------------------------
func (a *App) apiGuestAdd(c *Ctx) error {
	var in struct {
		Nama      string `json:"nama"`
		Instansi  string `json:"instansi"`
		HP        string `json:"hp"`
		Keperluan string `json:"keperluan"`
		Bertemu   string `json:"bertemu"`
		Foto      string `json:"foto"`
		Website   string `json:"website"`
	}
	done, err := c.publicBlocked("tamu:", 6) // sebelum membaca isi (hingga 3 MB)
	if err != nil {
		return err
	}
	defer done()
	c.longBody()
	if err := c.bindLarge(&in, 3<<20); err != nil {
		return err
	}
	if err := c.publicGuard("tamu:", 6, in.Website); err != nil {
		return err
	}
	cut := func(s string, n int) any {
		r := []rune(strings.TrimSpace(s))
		if len(r) == 0 {
			return nil
		}
		if len(r) > n {
			r = r[:n]
		}
		return string(r)
	}
	if cut(in.Nama, 1) == nil || cut(in.Keperluan, 1) == nil {
		return bad("Nama dan keperluan wajib diisi.")
	}
	var foto any
	if in.Foto != "" {
		data, err := decodeDataURL(in.Foto)
		if err != nil || len(data) > 2<<20 {
			return bad("Foto tidak valid.")
		}
		if foto, err = a.savePublicUpload(data, true, limitKey(clientIP(c.R))); err != nil {
			return err
		}
	}
	if _, err := a.db.Exec(`INSERT INTO guestbook (nama, instansi, hp, keperluan, bertemu, foto, created_at) VALUES (?,?,?,?,?,?,?)`,
		cut(in.Nama, 100), cut(in.Instansi, 120), cut(in.HP, 30), cut(in.Keperluan, 500), cut(in.Bertemu, 100), foto, nowStr()); err != nil {
		if u, ok := foto.(string); ok {
			a.removeUploads([]string{u})
		}
		return err
	}
	return c.OK()
}

func apiGuestList(c *Ctx) error {
	from, to, err := monthRange(c.R.URL.Query().Get("bulan"))
	if err != nil {
		return err
	}
	lo, hi := dayRangeUTC(from, to)
	rows, err := c.list(`SELECT * FROM guestbook WHERE created_at >= ? AND created_at < ? ORDER BY created_at DESC`, lo, hi)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}
