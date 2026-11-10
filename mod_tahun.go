package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// Tahun ajaran & kenaikan kelas.
//
// Model: setiap kelas milik satu tahun ajaran (classes.tahun_id). Mapel, tugas, kuis, absensi, dan ujian
// melekat ke kelas, sehingga otomatis terarsip per tahun. users.kelas_id = kelas siswa saat ini.
// Saat kenaikan kelas, keanggotaan lama dicatat di class_members agar rekap tahun lalu tetap bisa dibuka.

func init() {
	moduleRoutes = append(moduleRoutes, func(a *App, m *http.ServeMux) {
		m.HandleFunc("GET /api/tahun", a.h("admin,guru", apiYears))
		m.HandleFunc("POST /api/tahun", a.h("admin", apiYearSave))
		m.HandleFunc("PUT /api/tahun/{id}", a.h("admin", apiYearSave))
		m.HandleFunc("PUT /api/tahun/{id}/aktif", a.h("admin", apiYearActivate))
		m.HandleFunc("POST /api/tahun/{id}/naik", a.h("admin", apiPromote))
	})
}

// activeYear: id tahun ajaran aktif.
func activeYear(q interface {
	QueryRow(string, ...any) *sql.Row
}) int64 {
	var id int64
	q.QueryRow(`SELECT id FROM academic_years WHERE aktif=1 ORDER BY id LIMIT 1`).Scan(&id)
	return id
}

// yearParam: tahun dari ?tahun=<id>, bawaan tahun aktif. 0 berarti "semua".
func (c *Ctx) yearParam() int64 {
	if v := c.R.URL.Query().Get("tahun"); v == "semua" {
		return 0
	} else if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
		return n
	}
	return activeYear(c.A.db)
}

// memberSQL: kondisi "u adalah siswa kelas X" yang juga berlaku untuk kelas tahun lalu (arsip).
// Kelas tahun berjalan: siswa aktif dengan kelas_id = X. Kelas arsip: tercatat di class_members.
const memberSQL = `((u.kelas_id = %[1]s AND u.status='aktif') OR EXISTS (SELECT 1 FROM class_members m WHERE m.siswa_id=u.id AND m.kelas_id=%[1]s))`

func memberOf(kelasExpr string) string { return fmt.Sprintf(memberSQL, kelasExpr) }

func apiYears(c *Ctx) error {
	rows, err := c.list(`SELECT y.*, (SELECT COUNT(*) FROM classes k WHERE k.tahun_id=y.id) AS jumlah_kelas,
		(SELECT COUNT(*) FROM courses co JOIN classes k ON k.id=co.kelas_id WHERE k.tahun_id=y.id) AS jumlah_mapel,
		(SELECT COUNT(*) FROM users u JOIN classes k ON k.id=u.kelas_id WHERE k.tahun_id=y.id AND u.role='siswa' AND u.status='aktif') AS jumlah_siswa
		FROM academic_years y ORDER BY y.nama DESC`)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

var yearRe = regexp.MustCompile(`^\d{4}/\d{4}$`)

func apiYearSave(c *Ctx) error {
	var in struct {
		Nama    string  `json:"nama"`
		Mulai   *string `json:"mulai"`
		Selesai *string `json:"selesai"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	nama := strings.TrimSpace(in.Nama)
	if !yearRe.MatchString(nama) {
		return bad("Format tahun ajaran harus seperti 2026/2027.")
	}
	dates := []any{nil, nil}
	for i, p := range []*string{in.Mulai, in.Selesai} {
		if v := str(p); v != nil {
			d, err := cleanDate(v.(string))
			if err != nil {
				return err
			}
			dates[i] = d
		}
	}
	if y1, y2 := nama[:4], nama[5:]; func() bool { a, _ := strconv.Atoi(y1); b, _ := strconv.Atoi(y2); return b != a+1 }() {
		return bad("Tahun ajaran harus dua tahun berurutan, mis. " + y1 + "/" + func() string { a, _ := strconv.Atoi(y1); return strconv.Itoa(a + 1) }() + ".")
	}
	if err := cekRentang(dates[0], dates[1], "Tanggal"); err != nil {
		return err
	}
	var err error
	if id := c.ID("id"); id > 0 {
		_, err = c.A.db.Exec(`UPDATE academic_years SET nama=?, mulai=?, selesai=? WHERE id=?`, nama, dates[0], dates[1], id)
		if err == nil {
			c.A.db.Exec(`UPDATE settings SET tahun_ajaran=? WHERE id=1 AND (SELECT aktif FROM academic_years WHERE id=?)=1`, nama, id)
			c.A.db.Exec(`UPDATE classes SET tahun_ajaran=? WHERE tahun_id=?`, nama, id)
		}
	} else {
		_, err = c.A.db.Exec(`INSERT INTO academic_years (nama, mulai, selesai, created_at) VALUES (?,?,?,?)`, nama, dates[0], dates[1], nowStr())
	}
	if err != nil {
		return err
	}
	return c.OK()
}

func apiYearActivate(c *Ctx) error {
	id := c.ID("id")
	var nama string
	if c.A.db.QueryRow(`SELECT nama FROM academic_years WHERE id=?`, id).Scan(&nama) != nil {
		return notFound()
	}
	tx, err := c.A.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, q := range [][]any{{`UPDATE academic_years SET aktif=(id=?)`, id}, {`UPDATE settings SET tahun_ajaran=? WHERE id=1`, nama}} {
		if _, err := tx.Exec(q[0].(string), q[1:]...); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.OK()
}

// apiPromote: kenaikan kelas dari tahun {id} (harus aktif) ke tahun tujuan, dalam satu transaksi.
// promoteReq: permintaan kenaikan kelas dari tahun ajaran aktif ke tahun tujuan.
type promoteReq struct {
	Tujuan struct {
		Nama    string  `json:"nama"`
		Mulai   *string `json:"mulai"`
		Selesai *string `json:"selesai"`
	} `json:"tujuan"`
	Kelas   []promoteKelas `json:"kelas"`
	Kecuali []struct {
		SiswaID int64  `json:"siswa_id"`
		Aksi    string `json:"aksi"` // naik | tinggal | lulus | keluar
	} `json:"kecuali"`
	SalinMapel bool `json:"salin_mapel"`
}

type promoteKelas struct {
	KelasID int64  `json:"kelas_id"`
	Aksi    string `json:"aksi"`    // naik | lulus
	KeNama  string `json:"ke_nama"` // nama kelas tujuan bila naik
}

// pengecualian per siswa (mis. tinggal kelas di kelas yang naik)
func (in *promoteReq) exceptions() (map[int64]string, error) {
	exc := map[int64]string{}
	for _, e := range in.Kecuali {
		switch e.Aksi {
		case "naik", "tinggal", "lulus", "keluar":
			exc[e.SiswaID] = e.Aksi
		default:
			return nil, bad("Aksi siswa tidak valid.")
		}
	}
	return exc, nil
}

// unlistedClasses: kelas tahun sumber yang masih berisi siswa aktif tetapi tidak diatur dalam permintaan.
// Semua wajib diproses agar tidak ada siswa tertinggal di tahun lama.
func unlistedClasses(c *Ctx, src int64, in *promoteReq) ([]string, error) {
	listed := map[int64]bool{}
	for _, k := range in.Kelas {
		listed[k.KelasID] = true
	}
	rows, err := c.A.reader().Query(`SELECT DISTINCT k.id, k.nama FROM classes k JOIN users u ON u.kelas_id=k.id
		WHERE k.tahun_id=? AND u.role='siswa' AND u.status='aktif'`, src)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var missing []string
	for rows.Next() {
		var id int64
		var nm string
		if err := rows.Scan(&id, &nm); err != nil {
			return nil, err
		}
		if !listed[id] {
			missing = append(missing, nm)
		}
	}
	return missing, rows.Err()
}

// promoter menjalankan kenaikan kelas di dalam SATU transaksi (semua atau tidak sama sekali).
type promoter struct {
	tx              *sql.Tx
	src, dst        int64
	srcNama, tujuan string
	kelasTujuan     map[string]int64 // nama kelas (huruf kecil) di tahun tujuan → id
	sum             map[string]int
}

// targetYear: tahun tujuan yang sudah ada (belum punya kelas) atau dibuat baru.
func (p *promoter) targetYear(in *promoteReq) error {
	if p.tx.QueryRow(`SELECT id FROM academic_years WHERE nama=?`, p.tujuan).Scan(&p.dst) == nil {
		return nil
	}
	tgl := func(s *string) any {
		if v := str(s); v != nil {
			if d, err := cleanDate(v.(string)); err == nil {
				return d
			}
		}
		return nil
	}
	res, err := p.tx.Exec(`INSERT INTO academic_years (nama, mulai, selesai, created_at) VALUES (?,?,?,?)`, p.tujuan, tgl(in.Tujuan.Mulai), tgl(in.Tujuan.Selesai), nowStr())
	if err != nil {
		return err
	}
	p.dst, _ = res.LastInsertId()
	return nil
}

// kelas: ambil atau buat kelas bernama nama di tahun tujuan.
func (p *promoter) kelas(nama, tingkat string) (int64, error) {
	key := strings.ToLower(strings.TrimSpace(nama))
	if id, ok := p.kelasTujuan[key]; ok {
		return id, nil
	}
	var id int64
	if p.tx.QueryRow(`SELECT id FROM classes WHERE tahun_id=? AND nama=? COLLATE NOCASE`, p.dst, strings.TrimSpace(nama)).Scan(&id) != nil {
		res, err := p.tx.Exec(`INSERT INTO classes (tahun_id, nama, tingkat, tahun_ajaran, created_at) VALUES (?,?,?,?,?)`,
			p.dst, strings.TrimSpace(nama), nullIfEmpty(tingkat), p.tujuan, nowStr())
		if err != nil {
			return 0, err
		}
		id, _ = res.LastInsertId()
	}
	p.kelasTujuan[key] = id
	return id, nil
}

// processClass: satu kelas sumber naik ke kelas tujuan (opsional menyalin mapel) atau lulus, lalu tiap siswanya dipindah.
func (p *promoter) processClass(k promoteKelas, exc map[int64]string, salinMapel bool) error {
	var nama string
	var tingkat sql.NullString
	if p.tx.QueryRow(`SELECT nama, tingkat FROM classes WHERE id=? AND tahun_id=?`, k.KelasID, p.src).Scan(&nama, &tingkat) != nil {
		return bad("Kelas sumber tidak valid.")
	}
	var naikID int64
	switch k.Aksi {
	case "naik":
		if strings.TrimSpace(k.KeNama) == "" {
			return bad("Isi nama kelas tujuan untuk " + nama + ".")
		}
		var err error
		if naikID, err = p.kelas(k.KeNama, nextTingkat(tingkat.String)); err != nil {
			return err
		}
		if salinMapel {
			if _, err := p.tx.Exec(`INSERT INTO courses (kode, nama, deskripsi, kelas_id, guru_id, created_at)
				SELECT kode, nama, deskripsi, ?, guru_id, ? FROM courses WHERE kelas_id=?
				AND lower(kode) NOT IN (SELECT lower(kode) FROM courses WHERE kelas_id=?)`, naikID, nowStr(), k.KelasID, naikID); err != nil {
				return err
			}
		}
	case "lulus":
	default:
		return bad("Aksi kelas harus naik atau lulus.")
	}
	ids, err := p.students(k.KelasID)
	if err != nil {
		return err
	}
	for _, sid := range ids {
		aksi := k.Aksi
		if e, ok := exc[sid]; ok {
			aksi = e
		}
		if err := p.moveStudent(sid, aksi, k.KelasID, naikID, nama, tingkat.String); err != nil {
			return err
		}
		p.sum[aksi]++
	}
	return nil
}

func (p *promoter) students(kelasID int64) ([]int64, error) {
	rows, err := p.tx.Query(`SELECT id FROM users WHERE role='siswa' AND status='aktif' AND kelas_id=?`, kelasID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// moveStudent: catat riwayat kelas lama, lalu naik / tinggal (kelas bernama sama di tahun baru) / lulus / keluar.
func (p *promoter) moveStudent(sid int64, aksi string, kelasID, naikID int64, nama, tingkat string) error {
	if _, err := p.tx.Exec(`INSERT OR REPLACE INTO class_members (siswa_id, kelas_id, tahun_id, status) VALUES (?,?,?,?)`, sid, kelasID, p.src, aksi); err != nil {
		return err
	}
	var err error
	switch aksi {
	case "naik":
		if naikID == 0 {
			return bad("Siswa di kelas " + nama + " ditandai naik, tetapi kelas ini diluluskan. Pilih tinggal/lulus/keluar.")
		}
		_, err = p.tx.Exec(`UPDATE users SET kelas_id=? WHERE id=?`, naikID, sid)
	case "tinggal":
		var tid int64
		if tid, err = p.kelas(nama, tingkat); err == nil {
			_, err = p.tx.Exec(`UPDATE users SET kelas_id=? WHERE id=?`, tid, sid)
		}
	case "lulus", "keluar":
		q := `UPDATE users SET kelas_id=NULL, status='nonaktif' WHERE id=?`
		args := []any{sid}
		if aksi == "lulus" {
			q = `UPDATE users SET kelas_id=NULL, status='nonaktif', angkatan_lulus=? WHERE id=?`
			args = []any{p.srcNama, sid}
		}
		if _, err = p.tx.Exec(q, args...); err == nil {
			_, err = p.tx.Exec(`DELETE FROM sessions WHERE user_id=?`, sid) // akun nonaktif keluar dari semua perangkat
		}
	}
	return err
}

// finish: tahun sumber ditandai sudah dinaikkan, tahun tujuan menjadi aktif.
func (p *promoter) finish() error {
	for _, q := range [][]any{{`UPDATE academic_years SET dinaikkan=1 WHERE id=?`, p.src}, {`UPDATE academic_years SET aktif=(id=?)`, p.dst},
		{`UPDATE settings SET tahun_ajaran=? WHERE id=1`, p.tujuan}} {
		if _, err := p.tx.Exec(q[0].(string), q[1:]...); err != nil {
			return err
		}
	}
	return nil
}

func apiPromote(c *Ctx) error {
	var in promoteReq
	if err := c.Bind(&in); err != nil {
		return err
	}
	src := c.ID("id")
	var srcNama string
	var aktif, dinaikkan int
	if c.A.db.QueryRow(`SELECT nama, aktif, dinaikkan FROM academic_years WHERE id=?`, src).Scan(&srcNama, &aktif, &dinaikkan) != nil {
		return notFound()
	}
	if aktif != 1 {
		return bad("Kenaikan kelas hanya dari tahun ajaran yang sedang aktif.")
	}
	if dinaikkan == 1 {
		return bad("Kenaikan kelas untuk tahun ajaran ini sudah pernah dijalankan.")
	}
	tujuan := strings.TrimSpace(in.Tujuan.Nama)
	if !yearRe.MatchString(tujuan) || tujuan == srcNama {
		return bad("Isi tahun ajaran tujuan, misalnya 2027/2028.")
	}
	exc, err := in.exceptions()
	if err != nil {
		return err
	}
	missing, err := unlistedClasses(c, src, &in)
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		return bad("Kelas berikut masih berisi siswa dan belum diatur: " + strings.Join(missing, ", "))
	}

	tx, err := c.A.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	p := &promoter{tx: tx, src: src, srcNama: srcNama, tujuan: tujuan, kelasTujuan: map[string]int64{}, sum: map[string]int{}}
	if err := p.targetYear(&in); err != nil {
		return err
	}
	for _, k := range in.Kelas {
		if err := p.processClass(k, exc, in.SalinMapel); err != nil {
			return err
		}
	}
	if err := p.finish(); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.JSON(map[string]any{"tahun_baru": tujuan, "naik": p.sum["naik"], "tinggal": p.sum["tinggal"], "lulus": p.sum["lulus"], "keluar": p.sum["keluar"], "kelas_baru": len(p.kelasTujuan)})
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return strings.TrimSpace(s)
}

var romanGrades = []string{"I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"}

// nextTingkat: "7" -> "8", "X" -> "XI". Selain itu dibiarkan.
func nextTingkat(t string) string {
	t = strings.TrimSpace(t)
	if n, err := strconv.Atoi(t); err == nil {
		return strconv.Itoa(n + 1)
	}
	for i, r := range romanGrades {
		if strings.EqualFold(r, t) && i+1 < len(romanGrades) {
			return romanGrades[i+1]
		}
	}
	return t
}
