package main

import (
	"fmt"
	"strings"
	"time"
)

// Surat masuk / keluar: arsip, nomor otomatis, disposisi.

// ---------------------------------------------------------------------
// Surat masuk / keluar
// ---------------------------------------------------------------------
func apiLetters(c *Ctx) error {
	q := `SELECT * FROM letters WHERE 1=1`
	args := []any{}
	if j := c.R.URL.Query().Get("jenis"); j == "masuk" || j == "keluar" {
		q += ` AND jenis=?`
		args = append(args, j)
	}
	if y := c.R.URL.Query().Get("tahun"); len(y) == 4 {
		q += ` AND tanggal >= ? AND tanggal < ?`
		args = append(args, y+"-01-01", y+"-12-32")
	}
	rows, err := c.list(q+` ORDER BY tanggal DESC, id DESC`, args...)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

var romawi = []string{"", "I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"}

func apiLetterNumber(c *Ctx) error {
	cfg := getConfig(c.A.db, "surat")
	kode := strings.TrimSpace(c.R.URL.Query().Get("kode"))
	if kode == "" {
		kode = "421"
	}
	now := time.Now()
	var n int
	c.A.db.QueryRow(`SELECT COUNT(*) FROM letters WHERE jenis='keluar' AND tanggal >= ? AND tanggal < ?`, now.Format("2006")+"-01-01", now.Format("2006")+"-12-32").Scan(&n)
	f, _ := cfg["format"].(string)
	sing, _ := cfg["singkatan"].(string)
	// urut = jumlah surat + 1; bila nomor itu sudah terpakai (ada surat yang dihapus), naikkan sampai belum ada
	var no string
	for i := n + 1; i <= n+1000; i++ {
		no = strings.NewReplacer("{no}", fmt.Sprintf("%03d", i), "{kode}", kode, "{singkatan}", sing,
			"{bulan}", romawi[now.Month()], "{tahun}", now.Format("2006")).Replace(f)
		if _, dup := c.exists(`SELECT nomor FROM letters WHERE jenis='keluar' AND nomor=?`, no); !dup || !strings.Contains(f, "{no}") {
			break
		}
	}
	return c.JSON(map[string]string{"nomor": no})
}

func apiLetterSave(c *Ctx) error {
	var in struct {
		Jenis        string  `json:"jenis"`
		Nomor        *string `json:"nomor"`
		Tanggal      string  `json:"tanggal"`
		Pihak        string  `json:"pihak"`
		Perihal      string  `json:"perihal"`
		Kode         *string `json:"kode"`
		Ringkasan    *string `json:"ringkasan"`
		File         *string `json:"file"`
		DisposisiKe  *string `json:"disposisi_ke"`
		DisposisiIsi *string `json:"disposisi_isi"`
		Status       string  `json:"status"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	if in.Jenis != "masuk" && in.Jenis != "keluar" {
		return bad("Jenis surat tidak valid.")
	}
	tgl, err := cleanDate(in.Tanggal)
	if err != nil {
		return err
	}
	pihak, err := required(in.Pihak, map[string]string{"masuk": "Pengirim", "keluar": "Tujuan"}[in.Jenis])
	if err != nil {
		return err
	}
	perihal, err := required(in.Perihal, "Perihal")
	if err != nil {
		return err
	}
	if v := str(in.File); v != nil && !fileNameRe.MatchString(strings.TrimPrefix(v.(string), "/files/")) {
		return bad("File tidak valid.")
	}
	switch in.Status {
	case "baru", "didisposisi", "selesai":
	default:
		in.Status = "baru"
	}
	if v := str(in.Nomor); v != nil && in.Jenis == "keluar" {
		if p, ok := c.exists(`SELECT perihal FROM letters WHERE jenis='keluar' AND nomor=? COLLATE NOCASE AND id<>?`, v, c.ID("id")); ok {
			return bad("Nomor surat keluar ini sudah dipakai (perihal: " + p + ").")
		}
	}
	in.DisposisiKe = c.canon("letters", "disposisi_ke", in.DisposisiKe)
	args := []any{in.Jenis, str(in.Nomor), tgl, pihak, perihal, str(in.Kode), str(in.Ringkasan), str(in.File), str(in.DisposisiKe), str(in.DisposisiIsi), in.Status}
	if id := c.ID("id"); id > 0 {
		_, err = c.A.db.Exec(`UPDATE letters SET jenis=?, nomor=?, tanggal=?, pihak=?, perihal=?, kode=?, ringkasan=?, file=?, disposisi_ke=?, disposisi_isi=?, status=? WHERE id=?`, append(args, id)...)
	} else {
		_, err = c.A.db.Exec(`INSERT INTO letters (jenis, nomor, tanggal, pihak, perihal, kode, ringkasan, file, disposisi_ke, disposisi_isi, status, created_by, created_at) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			append(args, c.U.ID, nowStr())...)
	}
	if err != nil {
		return err
	}
	return c.OK()
}
