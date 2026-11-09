package main

import ()

// LMS: absensi per mapel dan rekap nilai.

// ---------------------------------------------------------------------
// Absensi
// ---------------------------------------------------------------------
func apiAttendance(c *Ctx) error {
	cid := c.QInt("course_id")
	if !c.canManage(cid) {
		return forbidden()
	}
	tgl, err := cleanDate(c.R.URL.Query().Get("tanggal"))
	if err != nil {
		return err
	}
	rows, err := c.list(`SELECT siswa_id, status FROM attendance WHERE course_id=? AND tanggal=?`, cid, tgl)
	if err != nil {
		return err
	}
	dates, err := c.list(`SELECT tanggal, COUNT(*) AS jumlah, SUM(status='hadir') AS hadir FROM attendance WHERE course_id=?
		GROUP BY tanggal ORDER BY tanggal DESC LIMIT 30`, cid)
	if err != nil {
		return err
	}
	return c.JSON(map[string]any{"records": rows, "dates": dates})
}

func apiSaveAttendance(c *Ctx) error {
	var in struct {
		CourseID int64  `json:"course_id"`
		Tanggal  string `json:"tanggal"`
		Items    []struct {
			SiswaID int64  `json:"siswa_id"`
			Status  string `json:"status"`
		} `json:"items"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	if err := c.requireManage(in.CourseID, "absensi"); err != nil {
		return err
	}
	tgl, err := cleanDate(in.Tanggal)
	if err != nil {
		return err
	}
	tx, err := c.A.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, it := range in.Items {
		switch it.Status {
		case "hadir", "izin", "sakit", "alfa":
		default:
			return bad("Status kehadiran tidak valid.")
		}
		// pastikan siswa memang di kelas mapel ini
		var ok int
		tx.QueryRow(`SELECT COUNT(*) FROM users u JOIN courses co ON co.kelas_id=u.kelas_id WHERE co.id=? AND u.id=? AND u.role='siswa'`, in.CourseID, it.SiswaID).Scan(&ok)
		if ok == 0 {
			continue
		}
		if _, err := tx.Exec(`INSERT INTO attendance (course_id, siswa_id, tanggal, status) VALUES (?,?,?,?)
			ON CONFLICT (course_id, siswa_id, tanggal) DO UPDATE SET status=excluded.status`, in.CourseID, it.SiswaID, tgl, it.Status); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.OK()
}

func apiDeleteAttendance(c *Ctx) error {
	cid := c.QInt("course_id")
	if err := c.requireManage(cid, "hapus_absen"); err != nil {
		return err
	}
	tgl, err := cleanDate(c.R.URL.Query().Get("tanggal"))
	if err != nil {
		return err
	}
	if _, err := c.A.db.Exec(`DELETE FROM attendance WHERE course_id=? AND tanggal=?`, cid, tgl); err != nil {
		return err
	}
	return c.OK()
}

// ---------------------------------------------------------------------
// Rekap nilai
// ---------------------------------------------------------------------
func apiRekap(c *Ctx) error {
	cid := c.QInt("course_id")
	if !c.canManage(cid) {
		return forbidden()
	}
	out := map[string]any{}
	var err error
	set := func(k, q string, args ...any) {
		if err != nil {
			return
		}
		var r []map[string]any
		r, err = c.list(q, args...)
		out[k] = r
	}
	st, e := c.students(cid)
	if e != nil {
		return e
	}
	out["students"] = st
	set("assignments", `SELECT id, judul, nilai_maks, deadline FROM assignments WHERE course_id=? ORDER BY created_at`, cid)
	set("submissions", `SELECT s.assignment_id, s.siswa_id, s.nilai FROM submissions s JOIN assignments a ON a.id=s.assignment_id WHERE a.course_id=?`, cid)
	set("quizzes", `SELECT id, judul FROM quizzes WHERE course_id=? ORDER BY created_at`, cid)
	set("attempts", `SELECT t.quiz_id, t.siswa_id, t.skor FROM quiz_attempts t JOIN quizzes q ON q.id=t.quiz_id WHERE q.course_id=? AND t.selesai=1`, cid)
	set("attendance", `SELECT siswa_id, tanggal, status FROM attendance WHERE course_id=?`, cid)
	if err != nil {
		return err
	}
	return c.JSON(out)
}

func apiRekapMe(c *Ctx) error {
	kelas := c.U.KelasID.Int64
	out := map[string]any{}
	var err error
	set := func(k, q string, args ...any) {
		if err != nil {
			return
		}
		var r []map[string]any
		r, err = c.list(q, args...)
		out[k] = r
	}
	set("courses", `SELECT co.id, co.kode, co.nama, co.deskripsi, g.nama_lengkap AS guru_nama FROM courses co LEFT JOIN users g ON g.id=co.guru_id
		WHERE co.kelas_id=? ORDER BY co.nama`, kelas)
	set("materials", `SELECT m.id, m.course_id FROM materials m JOIN courses co ON co.id=m.course_id WHERE co.kelas_id=?`, kelas)
	set("progress", `SELECT p.material_id FROM material_progress p JOIN materials m ON m.id=p.material_id JOIN courses co ON co.id=m.course_id
		WHERE p.siswa_id=? AND co.kelas_id=?`, c.U.ID, kelas)
	set("assignments", `SELECT a.id, a.course_id, a.judul, a.deskripsi, a.deadline, a.nilai_maks, a.created_at FROM assignments a
		JOIN courses co ON co.id=a.course_id WHERE co.kelas_id=? ORDER BY COALESCE(a.deadline, a.created_at) DESC`, kelas)
	set("submissions", `SELECT s.id, s.assignment_id, s.jawaban, s.link_file, s.nilai, s.feedback, s.submitted_at FROM submissions s
		JOIN assignments a ON a.id=s.assignment_id JOIN courses co ON co.id=a.course_id WHERE s.siswa_id=? AND co.kelas_id=?`, c.U.ID, kelas)
	set("quizzes", `SELECT q.id, q.course_id, q.judul, q.durasi_menit, q.created_at, (SELECT COUNT(*) FROM quiz_questions x WHERE x.quiz_id=q.id) AS jumlah_soal
		FROM quizzes q JOIN courses co ON co.id=q.course_id WHERE co.kelas_id=? ORDER BY q.created_at DESC`, kelas)
	set("attempts", `SELECT t.quiz_id, t.skor, t.selesai, t.submitted_at FROM quiz_attempts t JOIN quizzes q ON q.id=t.quiz_id
		JOIN courses co ON co.id=q.course_id WHERE t.siswa_id=? AND co.kelas_id=?`, c.U.ID, kelas)
	set("attendance", `SELECT course_id, tanggal, status FROM attendance WHERE siswa_id=? AND course_id IN (SELECT id FROM courses WHERE kelas_id=?)
		ORDER BY tanggal DESC`, c.U.ID, kelas)
	if err != nil {
		return err
	}
	return c.JSON(out)
}

// kvTahun: nama tahun ajaran aktif (untuk dashboard).
func kvTahun(c *Ctx) string {
	var n string
	c.A.db.QueryRow(`SELECT nama FROM academic_years WHERE aktif=1`).Scan(&n)
	return n
}
