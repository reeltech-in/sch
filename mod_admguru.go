package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Modul: Administrasi Guru & Wali Kelas
//   - Wali kelas per kelas (classes.wali_id, migrasi v2)
//   - Jurnal mengajar harian (kehadiran dihitung otomatis dari absensi mapel bila diisi)
//   - Catatan siswa: prestasi, pelanggaran (poin), catatan wali, catatan BK
//
// Siapa melihat catatan siswa:
//   admin & staf "admguru" (mis. BK/Kesiswaan) : semua
//   wali kelas                                   : siswa di kelasnya, kecuali catatan BK milik orang lain
//   guru pengajar                                : hanya catatan yang ia tulis
//   siswa                                        : prestasi & pelanggaran miliknya sendiri

func init() {
	moduleKeys = append(moduleKeys, "admguru")
	migrations = append(migrations, migration{name: "wali kelas", up: func(tx *sql.Tx) error {
		var n int
		tx.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('classes') WHERE name='wali_id'`).Scan(&n)
		if n > 0 {
			return nil
		}
		_, err := tx.Exec(`ALTER TABLE classes ADD COLUMN wali_id INTEGER REFERENCES users(id) ON DELETE SET NULL`)
		return err
	}})
	moduleSchemas = append(moduleSchemas, `
CREATE TABLE IF NOT EXISTS teaching_journal (
  id          INTEGER PRIMARY KEY,
  guru_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  course_id   INTEGER REFERENCES courses(id) ON DELETE SET NULL,
  tanggal     TEXT NOT NULL,
  jam_ke      TEXT,
  materi      TEXT NOT NULL,
  kegiatan    TEXT,
  kehadiran   TEXT,
  catatan     TEXT,
  created_at  TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_journal_guru ON teaching_journal(guru_id, tanggal);
CREATE INDEX IF NOT EXISTS idx_journal_course ON teaching_journal(course_id, tanggal);

CREATE TABLE IF NOT EXISTS student_notes (
  id            INTEGER PRIMARY KEY,
  siswa_id      INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  tanggal       TEXT NOT NULL,
  jenis         TEXT NOT NULL CHECK (jenis IN ('prestasi','pelanggaran','catatan','bk')),
  kategori      TEXT,
  poin          INTEGER NOT NULL DEFAULT 0,
  keterangan    TEXT NOT NULL,
  tindak_lanjut TEXT,
  created_by    INTEGER REFERENCES users(id) ON DELETE SET NULL,
  created_at    TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_notes_siswa ON student_notes(siswa_id, tanggal);
`)
	moduleRoutes = append(moduleRoutes, func(a *App, m *http.ServeMux) {
		g := func(roles string, fn func(*Ctx) error) http.HandlerFunc { return a.mod("admguru", roles, fn) }
		m.HandleFunc("GET /api/jurnal", g("admin,guru", apiJournal))
		m.HandleFunc("POST /api/jurnal", g("admin,guru", apiJournalSave))
		m.HandleFunc("PUT /api/jurnal/{id}", g("admin,guru", apiJournalSave))
		m.HandleFunc("DELETE /api/jurnal/{id}", g("admin,guru", apiJournalDelete))

		m.HandleFunc("GET /api/catatan", g("admin,guru", apiNotes))
		m.HandleFunc("GET /api/catatan/me", g("siswa", apiNotesMe))
		m.HandleFunc("POST /api/catatan", g("admin,guru", apiNoteSave))
		m.HandleFunc("PUT /api/catatan/{id}", g("admin,guru", apiNoteSave))
		m.HandleFunc("DELETE /api/catatan/{id}", g("admin,guru", apiNoteDelete))

		m.HandleFunc("GET /api/walikelas", g("admin,guru", apiWaliClasses))
		m.HandleFunc("GET /api/walikelas/{id}", g("admin,guru", apiWaliDetail))
		m.HandleFunc("GET /api/catatan/siswa", g("admin,guru", apiNoteStudents))
	})
}

var monthRe = regexp.MustCompile(`^\d{4}-\d{2}$`)

// staffAdm: admin atau guru dengan tugas tambahan modul ini (BK/Kesiswaan).
func (c *Ctx) staffAdm() bool { return c.U.Has("admguru") }

func (c *Ctx) isWali(kelasID int64) bool {
	var w sql.NullInt64
	c.A.db.QueryRow(`SELECT wali_id FROM classes WHERE id=?`, kelasID).Scan(&w)
	return w.Valid && w.Int64 == c.U.ID
}

func (c *Ctx) teachesClass(kelasID int64) bool {
	var n int
	c.A.db.QueryRow(`SELECT COUNT(*) FROM courses WHERE kelas_id=? AND guru_id=?`, kelasID, c.U.ID).Scan(&n)
	return n > 0
}

// ---------------------------------------------------------------------
// Wali kelas
// ---------------------------------------------------------------------

// apiWaliClasses: kelas tahun aktif yang bisa dibuka di halaman Wali Kelas.
func apiWaliClasses(c *Ctx) error {
	q := `SELECT k.id, k.nama, k.tingkat, k.wali_id, u.nama_lengkap AS wali_nama,
		(SELECT COUNT(*) FROM users s WHERE s.role='siswa' AND s.status='aktif' AND s.kelas_id=k.id) AS jumlah_siswa
		FROM classes k LEFT JOIN users u ON u.id=k.wali_id WHERE k.tahun_id=?`
	args := []any{activeYear(c.A.db)}
	if !c.staffAdm() {
		q += ` AND k.wali_id=?`
		args = append(args, c.U.ID)
	}
	rows, err := c.list(q+` ORDER BY k.tingkat, k.nama`, args...)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

// apiWaliDetail: daftar siswa satu kelas + rekap absensi harian bulan ini + poin catatan.
func apiWaliDetail(c *Ctx) error {
	kid := c.ID("id")
	if !c.staffAdm() && !c.isWali(kid) {
		return forbidden()
	}
	bulan := c.R.URL.Query().Get("bulan")
	if !monthRe.MatchString(bulan) {
		bulan = time.Now().Format("2006-01")
	}
	m1, m2, _ := monthRange(bulan)
	k, err := c.list(`SELECT k.id, k.nama, k.tingkat, COALESCE(y.nama, k.tahun_ajaran) AS tahun_ajaran, u.nama_lengkap AS wali_nama, u.nip AS wali_nip
		FROM classes k LEFT JOIN users u ON u.id=k.wali_id LEFT JOIN academic_years y ON y.id=k.tahun_id WHERE k.id=?`, kid)
	if err != nil || len(k) == 0 {
		return notFound()
	}
	rows, err := c.list(`SELECT u.id, u.nama_lengkap, u.username, p.nisn, p.jk, p.hp_ortu,
		COALESCE(r.hadir,0) AS hadir, COALESCE(r.terlambat,0) AS terlambat, COALESCE(r.sakit,0) AS sakit,
		COALESCE(r.izin,0) AS izin, COALESCE(r.alfa,0) AS alfa,
		(SELECT COALESCE(SUM(poin),0) FROM student_notes n WHERE n.siswa_id=u.id AND n.jenis='pelanggaran') AS poin_pelanggaran,
		(SELECT COALESCE(SUM(poin),0) FROM student_notes n WHERE n.siswa_id=u.id AND n.jenis='prestasi') AS poin_prestasi,
		(SELECT COUNT(*) FROM student_notes n WHERE n.siswa_id=u.id AND n.jenis IN ('prestasi','pelanggaran','catatan')) AS jumlah_catatan
		FROM users u LEFT JOIN student_profiles p ON p.siswa_id=u.id
		LEFT JOIN (SELECT siswa_id, SUM(status IN ('hadir','terlambat')) AS hadir, SUM(status='terlambat') AS terlambat,
			SUM(status='sakit') AS sakit, SUM(status='izin') AS izin, SUM(status='alfa') AS alfa
			FROM daily_attendance WHERE tanggal >= ? AND tanggal <= ? GROUP BY siswa_id) r ON r.siswa_id=u.id
		WHERE u.role='siswa' AND `+memberOf("?")+` ORDER BY u.nama_lengkap`,
		m1, m2, kid, kid) // memberOf memakai id kelas dua kali
	if err != nil {
		return err
	}
	return c.JSON(map[string]any{"kelas": k[0], "bulan": bulan, "siswa": rows})
}

// ---------------------------------------------------------------------
// Jurnal mengajar
// ---------------------------------------------------------------------

func apiJournal(c *Ctx) error {
	q := c.R.URL.Query()
	bulan := q.Get("bulan")
	if !monthRe.MatchString(bulan) {
		bulan = time.Now().Format("2006-01")
	}
	sqlq := `SELECT j.*, u.nama_lengkap AS guru_nama, co.nama AS mapel, k.nama AS kelas_nama,
		(SELECT COUNT(*) FROM attendance a WHERE a.course_id=j.course_id AND a.tanggal=j.tanggal AND a.status='hadir') AS hadir,
		(SELECT COUNT(*) FROM attendance a WHERE a.course_id=j.course_id AND a.tanggal=j.tanggal AND a.status<>'hadir') AS tidak_hadir
		FROM teaching_journal j JOIN users u ON u.id=j.guru_id
		LEFT JOIN courses co ON co.id=j.course_id LEFT JOIN classes k ON k.id=co.kelas_id
		WHERE j.tanggal >= ? AND j.tanggal <= ?`
	m1, m2, _ := monthRange(bulan)
	args := []any{m1, m2}
	if !c.staffAdm() {
		sqlq += ` AND j.guru_id=?` // guru hanya melihat jurnalnya sendiri
		args = append(args, c.U.ID)
	} else if gid := q.Get("guru_id"); gid != "" {
		sqlq += ` AND j.guru_id=?`
		args = append(args, gid)
	}
	rows, err := c.list(sqlq+` ORDER BY j.tanggal DESC, j.jam_ke, j.id DESC LIMIT 1000`, args...)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

func apiJournalSave(c *Ctx) error {
	var in struct {
		CourseID  *int64  `json:"course_id"`
		Tanggal   string  `json:"tanggal"`
		JamKe     *string `json:"jam_ke"`
		Materi    string  `json:"materi"`
		Kegiatan  *string `json:"kegiatan"`
		Kehadiran *string `json:"kehadiran"`
		Catatan   *string `json:"catatan"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	tgl, err := cleanDate(in.Tanggal)
	if err != nil || tgl == "" {
		return bad("Tanggal wajib diisi.")
	}
	materi, err := required(in.Materi, "Materi / topik")
	if err != nil {
		return err
	}
	var course any
	if in.CourseID != nil && *in.CourseID > 0 {
		if !c.canManage(*in.CourseID) {
			return bad("Pilih mapel yang Anda ampu.")
		}
		course = *in.CourseID
	}
	if j := str(in.JamKe); j != nil && course != nil {
		if _, ok := c.exists(`SELECT materi FROM teaching_journal WHERE guru_id=? AND course_id=? AND tanggal=? AND jam_ke=? AND id<>?`, c.U.ID, course, tgl, j, c.ID("id")); ok {
			return bad("Jurnal untuk mapel, tanggal, dan jam itu sudah diisi. Ubah entri yang ada.")
		}
	}
	if id := c.ID("id"); id > 0 {
		res, err := c.A.db.Exec(`UPDATE teaching_journal SET course_id=?, tanggal=?, jam_ke=?, materi=?, kegiatan=?, kehadiran=?, catatan=? WHERE id=? AND (guru_id=? OR ?)`,
			course, tgl, str(in.JamKe), materi, str(in.Kegiatan), str(in.Kehadiran), str(in.Catatan), id, c.U.ID, c.U.Role == "admin")
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return forbidden()
		}
		return c.OK()
	}
	_, err = c.A.db.Exec(`INSERT INTO teaching_journal (guru_id, course_id, tanggal, jam_ke, materi, kegiatan, kehadiran, catatan, created_at) VALUES (?,?,?,?,?,?,?,?,?)`,
		c.U.ID, course, tgl, str(in.JamKe), materi, str(in.Kegiatan), str(in.Kehadiran), str(in.Catatan), nowStr())
	if err != nil {
		return err
	}
	return c.OK()
}

func apiJournalDelete(c *Ctx) error {
	res, err := c.A.db.Exec(`DELETE FROM teaching_journal WHERE id=? AND (guru_id=? OR ?)`, c.ID("id"), c.U.ID, c.U.Role == "admin")
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return forbidden()
	}
	return c.OK()
}

// ---------------------------------------------------------------------
// Catatan siswa
// ---------------------------------------------------------------------

var noteJenis = map[string]bool{"prestasi": true, "pelanggaran": true, "catatan": true, "bk": true}

// noteScope: kondisi SQL (alias n = student_notes, s = users siswa) catatan yang boleh dilihat pengguna ini.
func (c *Ctx) noteScope() (string, []any) {
	if c.staffAdm() {
		return "1=1", nil
	}
	return `(n.created_by=? OR (n.jenis<>'bk' AND EXISTS (SELECT 1 FROM classes w WHERE w.id=s.kelas_id AND w.wali_id=?)))`, []any{c.U.ID, c.U.ID}
}

func apiNotes(c *Ctx) error {
	q := c.R.URL.Query()
	scope, args := c.noteScope()
	sqlq := `SELECT n.*, s.nama_lengkap AS siswa_nama, s.username AS nis, k.nama AS kelas_nama, b.nama_lengkap AS penulis
		FROM student_notes n JOIN users s ON s.id=n.siswa_id LEFT JOIN classes k ON k.id=s.kelas_id
		LEFT JOIN users b ON b.id=n.created_by WHERE ` + scope
	if v := q.Get("siswa_id"); v != "" {
		sqlq += ` AND n.siswa_id=?`
		args = append(args, v)
	}
	if v := q.Get("kelas_id"); v != "" {
		sqlq += ` AND s.kelas_id=?`
		args = append(args, v)
	}
	if v := q.Get("jenis"); noteJenis[v] {
		sqlq += ` AND n.jenis=?`
		args = append(args, v)
	}
	if v := q.Get("bulan"); monthRe.MatchString(v) {
		m1, m2, _ := monthRange(v)
		sqlq += ` AND n.tanggal >= ? AND n.tanggal <= ?`
		args = append(args, m1, m2)
	}
	rows, err := c.list(sqlq+` ORDER BY n.tanggal DESC, n.id DESC LIMIT 1000`, args...)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

// apiNoteStudents: siswa aktif yang boleh dicatat pengguna ini (untuk pilihan di formulir).
func apiNoteStudents(c *Ctx) error {
	q := `SELECT s.id, s.nama_lengkap, s.username, k.id AS kelas_id, k.nama AS kelas_nama, (k.wali_id=?) AS wali
		FROM users s JOIN classes k ON k.id=s.kelas_id WHERE s.role='siswa' AND s.status='aktif'`
	args := []any{c.U.ID}
	if !c.staffAdm() {
		q += ` AND (k.wali_id=? OR EXISTS (SELECT 1 FROM courses co WHERE co.kelas_id=k.id AND co.guru_id=?))`
		args = append(args, c.U.ID, c.U.ID)
	}
	rows, err := c.list(q+` ORDER BY k.tingkat, k.nama, s.nama_lengkap`, args...)
	if err != nil {
		return err
	}
	return c.JSON(map[string]any{"siswa": rows, "staf": c.staffAdm()})
}

func apiNotesMe(c *Ctx) error {
	rows, err := c.list(`SELECT id, tanggal, jenis, kategori, poin, keterangan FROM student_notes
		WHERE siswa_id=? AND jenis IN ('prestasi','pelanggaran') ORDER BY tanggal DESC, id DESC`, c.U.ID)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

// canNote: boleh menulis catatan untuk siswa ini (staf, wali kelasnya, atau guru yang mengajar kelasnya).
func (c *Ctx) canNote(siswaID int64, jenis string) error {
	var role string
	var kelas sql.NullInt64
	if c.A.db.QueryRow(`SELECT role, kelas_id FROM users WHERE id=?`, siswaID).Scan(&role, &kelas) != nil || role != "siswa" {
		return bad("Pilih siswa yang valid.")
	}
	if c.staffAdm() {
		return nil
	}
	if jenis == "bk" && !(kelas.Valid && c.isWali(kelas.Int64)) {
		return apiErr{403, "Catatan BK hanya untuk wali kelas, BK/Kesiswaan, atau admin."}
	}
	if kelas.Valid && (c.isWali(kelas.Int64) || c.teachesClass(kelas.Int64)) {
		return nil
	}
	return apiErr{403, "Anda hanya bisa mencatat siswa di kelas yang Anda ajar atau walikan."}
}

func apiNoteSave(c *Ctx) error {
	var in struct {
		SiswaID      int64   `json:"siswa_id"`
		Tanggal      string  `json:"tanggal"`
		Jenis        string  `json:"jenis"`
		Kategori     *string `json:"kategori"`
		Poin         *int    `json:"poin"`
		Keterangan   string  `json:"keterangan"`
		TindakLanjut *string `json:"tindak_lanjut"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	if !noteJenis[in.Jenis] {
		return bad("Pilih jenis catatan.")
	}
	tgl, err := cleanDate(in.Tanggal)
	if err != nil || tgl == "" {
		return bad("Tanggal wajib diisi.")
	}
	ket, err := required(in.Keterangan, "Keterangan")
	if err != nil {
		return err
	}
	poin := 0
	if in.Poin != nil && (in.Jenis == "prestasi" || in.Jenis == "pelanggaran") {
		poin = *in.Poin
		if poin < 0 || poin > 1000 {
			return bad("Poin 0 sampai 1000.")
		}
	}
	if id := c.ID("id"); id > 0 {
		var owner sql.NullInt64
		var siswa int64
		if c.A.db.QueryRow(`SELECT created_by, siswa_id FROM student_notes WHERE id=?`, id).Scan(&owner, &siswa) != nil {
			return notFound()
		}
		if !c.staffAdm() && !(owner.Valid && owner.Int64 == c.U.ID) {
			return apiErr{403, "Catatan ini hanya bisa diubah penulisnya atau BK/admin."}
		}
		if err := c.canNote(siswa, in.Jenis); err != nil {
			return err
		}
		_, err := c.A.db.Exec(`UPDATE student_notes SET tanggal=?, jenis=?, kategori=?, poin=?, keterangan=?, tindak_lanjut=? WHERE id=?`,
			tgl, in.Jenis, str(in.Kategori), poin, ket, str(in.TindakLanjut), id)
		if err != nil {
			return err
		}
		c.Note = in.Jenis
		return c.OK()
	}
	if err := c.canNote(in.SiswaID, in.Jenis); err != nil {
		return err
	}
	res, err := c.A.db.Exec(`INSERT INTO student_notes (siswa_id, tanggal, jenis, kategori, poin, keterangan, tindak_lanjut, created_by, created_at) VALUES (?,?,?,?,?,?,?,?,?)`,
		in.SiswaID, tgl, in.Jenis, str(in.Kategori), poin, ket, str(in.TindakLanjut), c.U.ID, nowStr())
	if err != nil {
		return err
	}
	if in.Jenis == "prestasi" || in.Jenis == "pelanggaran" { // kabari orang tua (bila diaktifkan); BK & catatan wali tidak dikirim
		nid, _ := res.LastInsertId()
		nama, kelas, hp := c.A.siswaWA(in.SiswaID)
		c.A.waQueue("catatan", fmt.Sprintf("catatan:%d", nid), hp,
			map[string]string{"nama": nama, "kelas": kelas, "jenis": in.Jenis, "tanggal": fmtTanggal(tgl), "keterangan": ket}, false)
	}
	c.Note = in.Jenis + " " + strings.TrimSpace(str2(in.Kategori))
	return c.OK()
}

func str2(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func apiNoteDelete(c *Ctx) error {
	res, err := c.A.db.Exec(`DELETE FROM student_notes WHERE id=? AND (created_by=? OR ?)`, c.ID("id"), c.U.ID, c.staffAdm())
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return apiErr{403, "Catatan ini hanya bisa dihapus penulisnya atau BK/admin."}
	}
	return c.OK()
}
