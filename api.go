package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

func (a *App) routes(m *http.ServeMux) {
	// Akun & sesi
	m.HandleFunc("POST /api/login", a.h("", a.apiLogin))
	m.HandleFunc("POST /api/logout", a.h("", a.apiLogout))
	m.HandleFunc("GET /api/me", a.h("", a.apiMe))
	m.HandleFunc("GET /api/version", a.h("", apiVersion))
	m.HandleFunc("POST /api/password", a.h("any", a.apiPassword))

	// Pengaturan & admin
	m.HandleFunc("GET /api/settings", a.h("", apiGetSettings))
	m.HandleFunc("PUT /api/settings", a.h("admin", apiPutSettings))
	m.HandleFunc("GET /api/dashboard", a.h("admin,guru", apiDashboard))
	m.HandleFunc("GET /api/backup", a.h("admin", a.apiBackup))

	m.HandleFunc("GET /api/classes", a.h("any", apiClasses))
	m.HandleFunc("POST /api/classes", a.h("admin", apiSaveClass))
	m.HandleFunc("PUT /api/classes/{id}", a.h("admin", apiSaveClass))
	m.HandleFunc("DELETE /api/classes/{id}", a.h("admin", apiDeleteClass))

	m.HandleFunc("GET /api/users", a.h("admin", apiUsers))
	m.HandleFunc("POST /api/users", a.h("admin", apiSaveUser))
	m.HandleFunc("PUT /api/users/{id}", a.h("admin", apiSaveUser))
	m.HandleFunc("DELETE /api/users/{id}", a.h("admin", apiDeleteUser))
	m.HandleFunc("POST /api/users/import", a.h("admin", apiImportUsers))

	m.HandleFunc("GET /api/courses", a.h("any", apiCourses))
	m.HandleFunc("POST /api/courses", a.h("admin", apiSaveCourse))
	m.HandleFunc("PUT /api/courses/{id}", a.h("admin", apiSaveCourse))
	m.HandleFunc("DELETE /api/courses/{id}", a.h("admin", apiDeleteCourse))
	m.HandleFunc("GET /api/courses/{id}/students", a.h("admin,guru", apiCourseStudents))

	// Materi
	m.HandleFunc("GET /api/materials", a.h("any", apiMaterials))
	m.HandleFunc("POST /api/materials", a.h("admin,guru", apiSaveMaterial))
	m.HandleFunc("PUT /api/materials/{id}", a.h("admin,guru", apiSaveMaterial))
	m.HandleFunc("DELETE /api/materials/{id}", a.h("admin,guru", apiDeleteChild("materials", "materi")))
	m.HandleFunc("POST /api/materials/{id}/done", a.h("siswa", apiMaterialDone))
	m.HandleFunc("GET /api/progress", a.h("any", apiProgress))

	// Tugas
	m.HandleFunc("GET /api/assignments", a.h("any", apiAssignments))
	m.HandleFunc("POST /api/assignments", a.h("admin,guru", apiSaveAssignment))
	m.HandleFunc("PUT /api/assignments/{id}", a.h("admin,guru", apiSaveAssignment))
	m.HandleFunc("DELETE /api/assignments/{id}", a.h("admin,guru", apiDeleteChild("assignments", "tugas")))
	m.HandleFunc("GET /api/assignments/{id}/submissions", a.h("admin,guru", apiSubmissions))
	m.HandleFunc("POST /api/assignments/{id}/submit", a.h("siswa", apiSubmit))
	m.HandleFunc("PUT /api/submissions/{id}/grade", a.h("admin,guru", apiGrade))

	// Kuis
	m.HandleFunc("GET /api/quizzes", a.h("any", apiQuizzes))
	m.HandleFunc("POST /api/quizzes", a.h("admin,guru", apiSaveQuiz))
	m.HandleFunc("PUT /api/quizzes/{id}", a.h("admin,guru", apiSaveQuiz))
	m.HandleFunc("DELETE /api/quizzes/{id}", a.h("admin,guru", apiDeleteChild("quizzes", "kuis")))
	m.HandleFunc("GET /api/quizzes/{id}/questions", a.h("admin,guru", apiQuestions))
	m.HandleFunc("POST /api/quizzes/{id}/questions", a.h("admin,guru", apiSaveQuestion))
	m.HandleFunc("PUT /api/questions/{id}", a.h("admin,guru", apiSaveQuestion))
	m.HandleFunc("DELETE /api/questions/{id}", a.h("admin,guru", apiDeleteQuestion))
	m.HandleFunc("GET /api/quizzes/{id}/attempts", a.h("admin,guru", apiAttempts))
	m.HandleFunc("DELETE /api/attempts/{id}", a.h("admin,guru", apiDeleteAttempt))
	m.HandleFunc("POST /api/quizzes/{id}/start", a.h("siswa", apiQuizStart))
	m.HandleFunc("POST /api/quizzes/{id}/submit", a.h("siswa", apiQuizSubmit))
	m.HandleFunc("POST /api/quizzes/{id}/save", a.h("siswa", apiQuizSave))

	// Absensi & rekap
	m.HandleFunc("GET /api/attendance", a.h("admin,guru", apiAttendance))
	m.HandleFunc("PUT /api/attendance", a.h("admin,guru", apiSaveAttendance))
	m.HandleFunc("DELETE /api/attendance", a.h("admin,guru", apiDeleteAttendance))
	m.HandleFunc("GET /api/rekap", a.h("admin,guru", apiRekap))
	m.HandleFunc("GET /api/rekap/me", a.h("siswa", apiRekapMe))

	for _, r := range moduleRoutes {
		r(a, m)
	}
}

// ---------------------------------------------------------------------
// Hak akses
// ---------------------------------------------------------------------

// canManage: admin, atau guru pengampu mapel.
func (c *Ctx) canManage(courseID int64) bool {
	var guru sql.NullInt64
	if c.A.db.QueryRow(`SELECT guru_id FROM courses WHERE id=?`, courseID).Scan(&guru) != nil {
		return false
	}
	return c.U.Role == "admin" || (c.U.Role == "guru" && guru.Valid && guru.Int64 == c.U.ID)
}

// canView: pengelola mapel, atau siswa di kelas mapel tersebut.
func (c *Ctx) canView(courseID int64) bool {
	if c.U.Role == "siswa" {
		var kelas int64
		if c.A.db.QueryRow(`SELECT kelas_id FROM courses WHERE id=?`, courseID).Scan(&kelas) != nil {
			return false
		}
		return c.U.KelasID.Valid && c.U.KelasID.Int64 == kelas
	}
	return c.canManage(courseID)
}

func (c *Ctx) guruCan(key string) bool {
	if c.U.Role == "admin" {
		return true
	}
	m := aksesGuruCache.Load()
	if m == nil {
		var raw string
		c.A.db.QueryRow(`SELECT akses_guru FROM settings WHERE id=1`).Scan(&raw)
		mm := map[string]bool{}
		json.Unmarshal([]byte(raw), &mm)
		aksesGuruCache.Store(&mm)
		m = &mm
	}
	v, ok := (*m)[key]
	return !ok || v
}

// aksesGuruCache: hak akses guru (Pengaturan) di memori; dikosongkan saat pengaturan disimpan.
var aksesGuruCache atomic.Pointer[map[string]bool]

func (c *Ctx) requireManage(courseID int64, key string) error {
	if !c.canManage(courseID) {
		return forbidden()
	}
	if key != "" && !c.guruCan(key) {
		return apiErr{403, "Fitur ini dinonaktifkan admin untuk guru."}
	}
	return nil
}

// courseOf mencari course_id dari baris di tabel anak (materials/assignments/quizzes).
func (c *Ctx) courseOf(table string, id int64) int64 {
	var cid int64
	switch table {
	case "materials", "assignments", "quizzes":
		c.A.db.QueryRow(`SELECT course_id FROM `+table+` WHERE id=?`, id).Scan(&cid)
	}
	return cid
}

// ---------------------------------------------------------------------
// Helper validasi
// ---------------------------------------------------------------------
func str(p *string) any {
	if p == nil {
		return nil
	}
	s := strings.TrimSpace(*p)
	if s == "" {
		return nil
	}
	return s
}

func required(s, label string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", bad(label + " wajib diisi.")
	}
	return s, nil
}

var urlRe = regexp.MustCompile(`(?i)^https?://`)

func cleanURL(p *string) (any, error) {
	v := str(p)
	if v == nil {
		return nil, nil
	}
	if !urlRe.MatchString(v.(string)) {
		return nil, bad("Tautan harus diawali http:// atau https://")
	}
	return v, nil
}

func cleanDeadline(p *string) (any, error) {
	v := str(p)
	if v == nil {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, v.(string))
	if err != nil {
		return nil, bad("Format tenggat tidak valid.")
	}
	return t.UTC().Format(time.RFC3339), nil
}

func cleanDate(s string) (string, error) {
	t, err := time.Parse("2006-01-02", strings.TrimSpace(s))
	if err != nil {
		return "", bad("Tanggal tidak valid.")
	}
	return t.Format("2006-01-02"), nil
}

func apiDelete(table string) func(*Ctx) error {
	return func(c *Ctx) error {
		res, err := c.A.db.Exec(`DELETE FROM `+table+` WHERE id=?`, c.ID("id"))
		if err != nil {
			return err
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return notFound()
		}
		return c.OK()
	}
}

func apiDeleteChild(table, key string) func(*Ctx) error {
	return func(c *Ctx) error {
		id := c.ID("id")
		if err := c.requireManage(c.courseOf(table, id), key); err != nil {
			return err
		}
		return apiDelete(table)(c)
	}
}

// ---------------------------------------------------------------------
// Pengaturan & dashboard
// ---------------------------------------------------------------------
func apiGetSettings(c *Ctx) error {
	s, err := c.A.settingsData()
	if err != nil {
		return err
	}
	return c.JSON(s)
}

// settingsData: identitas sekolah + modul aktif (publik). Dipakai /api/settings dan disisipkan ke halaman awal.
func (a *App) settingsData() (map[string]any, error) {
	rows, err := a.rdb.Query(`SELECT nama_sekolah, logo_url, alamat, telepon, tahun_ajaran, kepala_sekolah, nip_kepala, akses_guru FROM settings WHERE id=1`)
	if err != nil {
		return nil, err
	}
	list, err := rowsToMaps(rows)
	if err != nil || len(list) == 0 {
		return map[string]any{}, err
	}
	s := list[0]
	s["modul"] = modulesEnabled(a.db)
	s["versi"] = "v" + version
	ab := getConfig(a.db, "absen")
	s["absen_selfie"] = ab["selfie"] == true
	s["warna"] = getConfig(a.db, "tampilan")["warna"]
	return s, nil
}

// apiPutSettings menyimpan identitas sekolah / hak akses guru. Hanya kolom yang DIKIRIM yang diubah (kolom lain
// tetap nilai tersimpan), lalu mengembalikan pengaturan terbaru — halaman tidak perlu mengambil ulang sebelum/sesudah
// menyimpan, dan perubahan admin lain pada kolom berbeda tidak tertimpa.
func apiPutSettings(c *Ctx) error {
	var raw map[string]json.RawMessage
	if err := c.Bind(&raw); err != nil {
		return err
	}
	var cur struct {
		nama                                                string
		logo, alamat, telepon, tahun, kepsek, nip, aksesRaw sql.NullString
	}
	if err := c.A.db.QueryRow(`SELECT nama_sekolah, logo_url, alamat, telepon, tahun_ajaran, kepala_sekolah, nip_kepala, akses_guru FROM settings WHERE id=1`).
		Scan(&cur.nama, &cur.logo, &cur.alamat, &cur.telepon, &cur.tahun, &cur.kepsek, &cur.nip, &cur.aksesRaw); err != nil {
		return err
	}
	// nilai kolom: dari kiriman bila ada (null/"" = kosongkan), selain itu nilai tersimpan
	val := func(k string, old sql.NullString) *string {
		if v, ok := raw[k]; ok {
			var s *string
			json.Unmarshal(v, &s)
			return s
		}
		if old.Valid {
			return &old.String
		}
		return nil
	}
	nama := cur.nama
	if v := val("nama_sekolah", sql.NullString{}); v != nil || raw["nama_sekolah"] != nil {
		n, err := required(deref(v), "Nama sekolah")
		if err != nil {
			return err
		}
		nama = n
	}
	logoIn := val("logo_url", cur.logo)
	var logo any
	var err error
	if v := str(logoIn); v != nil && strings.HasPrefix(v.(string), "/files/") {
		if !fileNameRe.MatchString(strings.TrimPrefix(v.(string), "/files/")) {
			return bad("Logo tidak valid.")
		}
		logo = v
	} else if logo, err = cleanURL(logoIn); err != nil {
		return err
	}
	akses := map[string]bool{}
	json.Unmarshal([]byte(cur.aksesRaw.String), &akses)
	if v, ok := raw["akses_guru"]; ok {
		in := map[string]bool{}
		json.Unmarshal(v, &in)
		akses = in
	}
	for _, k := range []string{"materi", "tugas", "kuis", "absensi", "hapus_absen"} {
		v, ok := akses[k]
		akses[k] = !ok || v
	}
	ab, _ := json.Marshal(akses)
	if _, err = c.A.db.Exec(`UPDATE settings SET nama_sekolah=?, logo_url=?, alamat=?, telepon=?, tahun_ajaran=?,
		kepala_sekolah=?, nip_kepala=?, akses_guru=? WHERE id=1`,
		nama, logo, str(val("alamat", cur.alamat)), str(val("telepon", cur.telepon)), str(val("tahun_ajaran", cur.tahun)),
		str(val("kepala_sekolah", cur.kepsek)), str(val("nip_kepala", cur.nip)), string(ab)); err != nil {
		return err
	}
	aksesGuruCache.Store(nil)
	return apiGetSettings(c)
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func apiDashboard(c *Ctx) error {
	count := func(q string, args ...any) int64 {
		var n int64
		c.A.reader().QueryRow(q, args...).Scan(&n)
		return n
	}
	if c.U.Role == "admin" {
		return c.JSON(map[string]any{
			"siswa":       count(`SELECT COUNT(*) FROM users WHERE role='siswa' AND status='aktif'`),
			"guru":        count(`SELECT COUNT(*) FROM users WHERE role='guru' AND status='aktif'`),
			"kelas":       count(`SELECT COUNT(*) FROM classes WHERE tahun_id=(SELECT id FROM academic_years WHERE aktif=1)`),
			"mapel":       count(`SELECT COUNT(*) FROM courses co JOIN classes k ON k.id=co.kelas_id WHERE k.tahun_id=(SELECT id FROM academic_years WHERE aktif=1)`),
			"tanpa_kelas": count(`SELECT COUNT(*) FROM users WHERE role='siswa' AND kelas_id IS NULL`),
			"tanpa_guru":  count(`SELECT COUNT(*) FROM courses co JOIN classes k ON k.id=co.kelas_id WHERE co.guru_id IS NULL AND k.tahun_id=(SELECT id FROM academic_years WHERE aktif=1)`),
			"tahun":       kvTahun(c),
			"penyimpanan": c.A.storageInfo(),
		})
	}
	pending, err := c.list(`SELECT s.id, s.submitted_at, a.id AS assignment_id, a.judul, co.id AS course_id, co.nama AS mapel,
			k.nama AS kelas_nama, u.nama_lengkap
		FROM submissions s JOIN assignments a ON a.id=s.assignment_id JOIN courses co ON co.id=a.course_id
		JOIN classes k ON k.id=co.kelas_id JOIN users u ON u.id=s.siswa_id
		WHERE co.guru_id=? AND s.nilai IS NULL AND k.tahun_id=(SELECT id FROM academic_years WHERE aktif=1) ORDER BY s.submitted_at LIMIT 20`, c.U.ID) // tahun ajaran aktif saja
	if err != nil {
		return err
	}
	return c.JSON(map[string]any{
		"mapel":   count(`SELECT COUNT(*) FROM courses co JOIN classes k ON k.id=co.kelas_id WHERE co.guru_id=? AND k.tahun_id=(SELECT id FROM academic_years WHERE aktif=1)`, c.U.ID),
		"materi":  count(`SELECT COUNT(*) FROM materials m JOIN courses co ON co.id=m.course_id JOIN classes k ON k.id=co.kelas_id WHERE co.guru_id=? AND k.tahun_id=(SELECT id FROM academic_years WHERE aktif=1)`, c.U.ID),
		"tugas":   count(`SELECT COUNT(*) FROM assignments a JOIN courses co ON co.id=a.course_id JOIN classes k ON k.id=co.kelas_id WHERE co.guru_id=? AND k.tahun_id=(SELECT id FROM academic_years WHERE aktif=1)`, c.U.ID),
		"kuis":    count(`SELECT COUNT(*) FROM quizzes q JOIN courses co ON co.id=q.course_id JOIN classes k ON k.id=co.kelas_id WHERE co.guru_id=? AND k.tahun_id=(SELECT id FROM academic_years WHERE aktif=1)`, c.U.ID),
		"siswa":   count(`SELECT COUNT(DISTINCT u.id) FROM users u JOIN courses co ON co.kelas_id=u.kelas_id WHERE co.guru_id=? AND u.role='siswa' AND u.status='aktif'`, c.U.ID),
		"pending": pending,
	})
}

// ---------------------------------------------------------------------
// Kelas
// ---------------------------------------------------------------------
func apiClasses(c *Ctx) error {
	rows, err := c.list(`SELECT k.id, k.nama, k.tingkat, k.tahun_ajaran, k.tahun_id, k.wali_id,
		(SELECT nama_lengkap FROM users w WHERE w.id=k.wali_id) AS wali_nama,
		`+memberCount("k.id")+` AS jumlah_siswa,
		(SELECT COUNT(*) FROM courses co WHERE co.kelas_id=k.id) AS jumlah_mapel
		FROM classes k WHERE (?=0 OR k.tahun_id=?) ORDER BY k.tingkat, k.nama`, c.yearParam(), c.yearParam())
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

func apiSaveClass(c *Ctx) error {
	var in struct {
		Nama        string  `json:"nama"`
		Tingkat     *string `json:"tingkat"`
		TahunAjaran *string `json:"tahun_ajaran"`
		WaliID      any     `json:"wali_id"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	nama, err := required(in.Nama, "Nama kelas")
	if err != nil {
		return err
	}
	var wali any // wali kelas: guru aktif, atau kosong
	if s := strings.TrimSpace(fmt.Sprint(in.WaliID)); in.WaliID != nil && s != "" && s != "0" {
		wid, _ := strconv.ParseInt(s, 10, 64)
		var role string
		if c.A.db.QueryRow(`SELECT role FROM users WHERE id=? AND status='aktif'`, wid).Scan(&role) != nil || role != "guru" {
			return bad("Wali kelas harus guru yang aktif.")
		}
		wali = wid
	}
	if in.Tingkat == nil {
		in.Tingkat = new(string)
	}
	if err := c.cekKelas(c.ID("id"), nama, in.Tingkat, wali); err != nil {
		return err
	}
	if id := c.ID("id"); id > 0 {
		_, err = c.A.db.Exec(`UPDATE classes SET nama=?, tingkat=?, tahun_ajaran=?, wali_id=? WHERE id=?`, nama, str(in.Tingkat), str(in.TahunAjaran), wali, id)
	} else {
		ty := activeYear(c.A.db)
		var tn string
		c.A.db.QueryRow(`SELECT nama FROM academic_years WHERE id=?`, ty).Scan(&tn)
		_, err = c.A.db.Exec(`INSERT INTO classes (tahun_id, nama, tingkat, tahun_ajaran, wali_id, created_at) VALUES (?,?,?,?,?,?)`, ty, nama, str(in.Tingkat), tn, wali, nowStr())
	}
	if err != nil {
		return err
	}
	return c.OK()
}
