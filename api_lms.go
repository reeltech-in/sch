package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// LMS: mapel, materi, tugas, kuis, absensi mapel, rekap nilai.

// ---------------------------------------------------------------------
// Mapel
// ---------------------------------------------------------------------
func apiCourses(c *Ctx) error {
	q := `SELECT co.id, co.kode, co.nama, co.deskripsi, co.kelas_id, co.guru_id, k.nama AS kelas_nama, g.nama_lengkap AS guru_nama,
		` + memberCount("co.kelas_id") + ` AS jumlah_siswa, k.tahun_id
		FROM courses co JOIN classes k ON k.id=co.kelas_id LEFT JOIN users g ON g.id=co.guru_id`
	var args []any
	switch c.U.Role {
	case "guru":
		q += ` WHERE co.guru_id=? AND (?=0 OR k.tahun_id=?)`
		y := c.yearParam()
		args = append(args, c.U.ID, y, y)
	case "siswa":
		q += ` WHERE co.kelas_id=?`
		args = append(args, c.U.KelasID.Int64)
	default:
		q += ` WHERE (?=0 OR k.tahun_id=?)`
		y := c.yearParam()
		args = append(args, y, y)
	}
	rows, err := c.list(q+` ORDER BY k.nama, co.nama`, args...)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

func apiSaveCourse(c *Ctx) error {
	var in struct {
		Kode      string  `json:"kode"`
		Nama      string  `json:"nama"`
		Deskripsi *string `json:"deskripsi"`
		KelasID   int64   `json:"kelas_id"`
		GuruID    *int64  `json:"guru_id"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	kode, err := required(in.Kode, "Kode mapel")
	if err != nil {
		return err
	}
	nama, err := required(in.Nama, "Nama mapel")
	if err != nil {
		return err
	}
	if in.KelasID == 0 {
		return bad("Kelas wajib dipilih.")
	}
	var guru any
	if in.GuruID != nil && *in.GuruID > 0 {
		var role string
		if c.A.db.QueryRow(`SELECT role FROM users WHERE id=?`, *in.GuruID).Scan(&role) != nil || role != "guru" {
			return bad("Guru pengampu tidak valid.")
		}
		guru = *in.GuruID
	}
	// kode mengikuti nama: sama dengan yang sudah dipakai sekolah / kode standar; kode milik mapel lain ditolak
	if kode, err = c.mapelKode(nama, kode, c.ID("id")); err != nil {
		return err
	}
	// satu mapel tidak boleh ganda di kelas yang sama (nama atau kode)
	var dup string
	if c.A.db.QueryRow(`SELECT nama FROM courses WHERE kelas_id=? AND id<>? AND (nama=? COLLATE NOCASE OR kode=? COLLATE NOCASE) LIMIT 1`,
		in.KelasID, c.ID("id"), nama, kode).Scan(&dup) == nil {
		return bad("Mapel " + dup + " (atau kode " + kode + ") sudah ada di kelas ini. Ubah mapel yang ada untuk mengganti guru pengampu.")
	}
	if id := c.ID("id"); id > 0 {
		_, err = c.A.db.Exec(`UPDATE courses SET kode=?, nama=?, deskripsi=?, kelas_id=?, guru_id=? WHERE id=?`, kode, nama, str(in.Deskripsi), in.KelasID, guru, id)
	} else {
		_, err = c.A.db.Exec(`INSERT INTO courses (kode, nama, deskripsi, kelas_id, guru_id, created_at) VALUES (?,?,?,?,?,?)`, kode, nama, str(in.Deskripsi), in.KelasID, guru, nowStr())
	}
	if err != nil {
		return err
	}
	return c.OK()
}

func (c *Ctx) students(courseID int64) ([]map[string]any, error) {
	var kelas int64
	c.A.db.QueryRow(`SELECT kelas_id FROM courses WHERE id=?`, courseID).Scan(&kelas)
	return c.list(`SELECT u.id, u.username, u.nama_lengkap FROM users u
		WHERE u.role='siswa' AND `+memberOf("?")+` ORDER BY u.nama_lengkap`, kelas, kelas)
}

func apiCourseStudents(c *Ctx) error {
	id := c.ID("id")
	if !c.canManage(id) {
		return forbidden()
	}
	rows, err := c.students(id)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

// ---------------------------------------------------------------------
// Materi
// ---------------------------------------------------------------------
func apiMaterials(c *Ctx) error {
	cid := c.QInt("course_id")
	if !c.canView(cid) {
		return forbidden()
	}
	rows, err := c.list(`SELECT id, course_id, judul, konten, lampiran, urutan, created_at,
		(SELECT COUNT(*) FROM material_progress p WHERE p.material_id=m.id) AS jumlah_selesai
		FROM materials m WHERE course_id=? ORDER BY urutan, id`, cid)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

type lampiran struct {
	Tipe  string `json:"tipe"`
	Judul string `json:"judul"`
	Isi   string `json:"isi"`
}

func apiSaveMaterial(c *Ctx) error {
	var in struct {
		CourseID int64      `json:"course_id"`
		Judul    string     `json:"judul"`
		Konten   *string    `json:"konten"`
		Lampiran []lampiran `json:"lampiran"`
		Urutan   int        `json:"urutan"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	id := c.ID("id")
	if id > 0 {
		in.CourseID = c.courseOf("materials", id)
	}
	if err := c.requireManage(in.CourseID, "materi"); err != nil {
		return err
	}
	judul, err := required(in.Judul, "Judul materi")
	if err != nil {
		return err
	}
	clean := []lampiran{}
	for _, l := range in.Lampiran {
		l.Isi = strings.TrimSpace(l.Isi)
		if l.Isi == "" {
			continue
		}
		if !urlRe.MatchString(l.Isi) {
			return bad("Tautan lampiran harus diawali http:// atau https:// (" + l.Isi + ")")
		}
		switch l.Tipe {
		case "youtube", "dokumen", "gambar", "link":
		default:
			l.Tipe = "link"
		}
		clean = append(clean, lampiran{l.Tipe, strings.TrimSpace(l.Judul), l.Isi})
	}
	lb, _ := json.Marshal(clean)
	if in.Urutan <= 0 {
		in.Urutan = 1
	}
	if id > 0 {
		_, err = c.A.db.Exec(`UPDATE materials SET judul=?, konten=?, lampiran=?, urutan=? WHERE id=?`, judul, str(in.Konten), string(lb), in.Urutan, id)
	} else {
		_, err = c.A.db.Exec(`INSERT INTO materials (course_id, judul, konten, lampiran, urutan, created_at) VALUES (?,?,?,?,?,?)`,
			in.CourseID, judul, str(in.Konten), string(lb), in.Urutan, nowStr())
	}
	if err != nil {
		return err
	}
	return c.OK()
}

func apiMaterialDone(c *Ctx) error {
	id := c.ID("id")
	var in struct {
		Selesai bool `json:"selesai"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	if !c.canView(c.courseOf("materials", id)) {
		return forbidden()
	}
	var err error
	if in.Selesai {
		_, err = c.A.db.Exec(`INSERT OR IGNORE INTO material_progress (material_id, siswa_id, completed_at) VALUES (?,?,?)`, id, c.U.ID, nowStr())
	} else {
		_, err = c.A.db.Exec(`DELETE FROM material_progress WHERE material_id=? AND siswa_id=?`, id, c.U.ID)
	}
	if err != nil {
		return err
	}
	return c.OK()
}

func apiProgress(c *Ctx) error {
	cid := c.QInt("course_id")
	if !c.canView(cid) {
		return forbidden()
	}
	q := `SELECT p.material_id, p.siswa_id, p.completed_at FROM material_progress p JOIN materials m ON m.id=p.material_id WHERE m.course_id=?`
	args := []any{cid}
	if c.U.Role == "siswa" {
		q += ` AND p.siswa_id=?`
		args = append(args, c.U.ID)
	}
	rows, err := c.list(q, args...)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

// ---------------------------------------------------------------------
// Tugas
// ---------------------------------------------------------------------
func apiAssignments(c *Ctx) error {
	cid := c.QInt("course_id")
	if !c.canView(cid) {
		return forbidden()
	}
	rows, err := c.list(`SELECT a.id, a.course_id, a.judul, a.deskripsi, a.deadline, a.nilai_maks, a.created_at,
		(SELECT COUNT(*) FROM submissions s WHERE s.assignment_id=a.id) AS jumlah_kumpul,
		(SELECT COUNT(*) FROM submissions s WHERE s.assignment_id=a.id AND s.nilai IS NOT NULL) AS jumlah_dinilai
		FROM assignments a WHERE a.course_id=? ORDER BY COALESCE(a.deadline, a.created_at) DESC`, cid)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

func apiSaveAssignment(c *Ctx) error {
	var in struct {
		CourseID  int64   `json:"course_id"`
		Judul     string  `json:"judul"`
		Deskripsi *string `json:"deskripsi"`
		Deadline  *string `json:"deadline"`
		NilaiMaks float64 `json:"nilai_maks"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	id := c.ID("id")
	if id > 0 {
		in.CourseID = c.courseOf("assignments", id)
	}
	if err := c.requireManage(in.CourseID, "tugas"); err != nil {
		return err
	}
	judul, err := required(in.Judul, "Judul tugas")
	if err != nil {
		return err
	}
	dl, err := cleanDeadline(in.Deadline)
	if err != nil {
		return err
	}
	if in.NilaiMaks <= 0 {
		in.NilaiMaks = 100
	}
	if id > 0 {
		_, err = c.A.db.Exec(`UPDATE assignments SET judul=?, deskripsi=?, deadline=?, nilai_maks=? WHERE id=?`, judul, str(in.Deskripsi), dl, in.NilaiMaks, id)
	} else {
		_, err = c.A.db.Exec(`INSERT INTO assignments (course_id, judul, deskripsi, deadline, nilai_maks, created_at) VALUES (?,?,?,?,?,?)`,
			in.CourseID, judul, str(in.Deskripsi), dl, in.NilaiMaks, nowStr())
	}
	if err != nil {
		return err
	}
	return c.OK()
}

func apiSubmissions(c *Ctx) error {
	id := c.ID("id")
	cid := c.courseOf("assignments", id)
	if !c.canManage(cid) {
		return forbidden()
	}
	as, err := c.list(`SELECT id, course_id, judul, deadline, nilai_maks FROM assignments WHERE id=?`, id)
	if err != nil || len(as) == 0 {
		return notFound()
	}
	st, err := c.students(cid)
	if err != nil {
		return err
	}
	rows, err := c.list(`SELECT s.id, s.siswa_id, s.jawaban, s.link_file, s.nilai, s.feedback, s.submitted_at
		FROM submissions s WHERE s.assignment_id=?`, id)
	if err != nil {
		return err
	}
	return c.JSON(map[string]any{"assignment": as[0], "students": st, "submissions": rows})
}

func apiSubmit(c *Ctx) error {
	id := c.ID("id")
	var in struct {
		Jawaban  *string `json:"jawaban"`
		LinkFile *string `json:"link_file"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	if !c.canView(c.courseOf("assignments", id)) {
		return forbidden()
	}
	link, err := cleanURL(in.LinkFile)
	if err != nil {
		return err
	}
	if str(in.Jawaban) == nil && link == nil {
		return bad("Isi jawaban atau tautan file terlebih dahulu.")
	}
	// satu pernyataan atomik: jawaban yang sudah dinilai tidak pernah tertimpa, walau guru menilai
	// tepat saat siswa mengirim ulang (pemeriksaan terpisah lalu simpan bisa kalah balapan)
	res, err := c.A.db.Exec(`INSERT INTO submissions (assignment_id, siswa_id, jawaban, link_file, submitted_at) VALUES (?,?,?,?,?)
		ON CONFLICT (assignment_id, siswa_id) DO UPDATE SET jawaban=excluded.jawaban, link_file=excluded.link_file, submitted_at=excluded.submitted_at
		WHERE submissions.nilai IS NULL`,
		id, c.U.ID, str(in.Jawaban), link, nowStr())
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return bad("Tugas sudah dinilai dan tidak bisa diubah.")
	}
	return c.OK()
}

// apiGradeSubmission: (guru) beri nilai & umpan balik untuk satu kumpulan tugas siswa.
func apiGradeSubmission(c *Ctx) error {
	id := c.ID("id")
	var in struct {
		Nilai    *float64 `json:"nilai"`
		Feedback *string  `json:"feedback"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	var aid int64
	var maks float64
	if err := c.A.db.QueryRow(`SELECT a.id, a.nilai_maks FROM submissions s JOIN assignments a ON a.id=s.assignment_id WHERE s.id=?`, id).Scan(&aid, &maks); err != nil {
		return notFound()
	}
	if !c.canManage(c.courseOf("assignments", aid)) {
		return forbidden()
	}
	var nilai any
	if in.Nilai != nil {
		if *in.Nilai < 0 || *in.Nilai > maks {
			return bad("Nilai harus antara 0 dan " + strconv.FormatFloat(maks, 'f', -1, 64) + ".")
		}
		nilai = *in.Nilai
	}
	c.Note = fmt.Sprintf("nilai tugas #%d = %v", id, nilai)
	if _, err := c.A.db.Exec(`UPDATE submissions SET nilai=?, feedback=? WHERE id=?`, nilai, str(in.Feedback), id); err != nil {
		return err
	}
	return c.OK()
}
