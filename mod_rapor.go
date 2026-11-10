package main

// e-Rapor Kurikulum Merdeka (Panduan Pembelajaran & Asesmen 2025, Permendikdasmen 13/2025) dan portofolio siswa.
//
// Isi rapor per semester: identitas, nilai akhir & capaian kompetensi per mapel (guru mapel), kokurikuler
// (8 dimensi profil lulusan, wali kelas), ekstrakurikuler (pembina), ketidakhadiran (absensi harian, bisa
// dikoreksi wali), catatan wali kelas, dan keputusan kenaikan kelas (semester genap).
//
// Peran:
//   - guru mapel  : nilai & deskripsi capaian untuk mapel yang diampu (nilai awal otomatis dari rumus 40/40/20)
//   - wali kelas  : kokurikuler, catatan, ketidakhadiran, keputusan, terbitkan rapor kelasnya
//   - pembina     : predikat & keterangan ekstrakurikuler yang dibinanya
//   - admin / staf modul "rapor": semuanya
//   - siswa       : melihat rapornya sendiri setelah diterbitkan; mengelola portofolio karyanya
//
// Setelah rapor kelas diterbitkan, isinya dikunci (batalkan terbit untuk mengubah).

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

func init() {
	moduleKeys = append(moduleKeys, "rapor")
	moduleSchemas = append(moduleSchemas, `
CREATE TABLE IF NOT EXISTS rapor_nilai (
  tahun_id    INTEGER NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
  semester    INTEGER NOT NULL CHECK (semester IN (1,2)),
  course_id   INTEGER NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
  siswa_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  nilai       REAL CHECK (nilai IS NULL OR (nilai >= 0 AND nilai <= 100)),
  tinggi      TEXT,
  perlu       TEXT,
  updated_by  INTEGER REFERENCES users(id) ON DELETE SET NULL,
  updated_at  TEXT NOT NULL,
  PRIMARY KEY (tahun_id, semester, course_id, siswa_id)
);
CREATE TABLE IF NOT EXISTS rapor_siswa (
  tahun_id     INTEGER NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
  semester     INTEGER NOT NULL CHECK (semester IN (1,2)),
  siswa_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  kokurikuler  TEXT NOT NULL DEFAULT '{}',
  koku_desk    TEXT,
  catatan      TEXT,
  keputusan    TEXT CHECK (keputusan IS NULL OR keputusan IN ('naik','tinggal','lulus')),
  sakit        INTEGER,
  izin         INTEGER,
  alfa         INTEGER,
  updated_by   INTEGER REFERENCES users(id) ON DELETE SET NULL,
  updated_at   TEXT NOT NULL,
  PRIMARY KEY (tahun_id, semester, siswa_id)
);
CREATE TABLE IF NOT EXISTS rapor_terbit (
  tahun_id   INTEGER NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
  semester   INTEGER NOT NULL CHECK (semester IN (1,2)),
  kelas_id   INTEGER NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
  tanggal    TEXT NOT NULL,
  oleh       INTEGER REFERENCES users(id) ON DELETE SET NULL,
  PRIMARY KEY (tahun_id, semester, kelas_id)
);
CREATE TABLE IF NOT EXISTS ekskul (
  id          INTEGER PRIMARY KEY,
  nama        TEXT NOT NULL UNIQUE COLLATE NOCASE,
  pembina_id  INTEGER REFERENCES users(id) ON DELETE SET NULL,
  created_at  TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS ekskul_anggota (
  ekskul_id  INTEGER NOT NULL REFERENCES ekskul(id) ON DELETE CASCADE,
  tahun_id   INTEGER NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
  siswa_id   INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  PRIMARY KEY (ekskul_id, tahun_id, siswa_id)
);
CREATE TABLE IF NOT EXISTS rapor_ekskul (
  tahun_id    INTEGER NOT NULL REFERENCES academic_years(id) ON DELETE CASCADE,
  semester    INTEGER NOT NULL CHECK (semester IN (1,2)),
  ekskul_id   INTEGER NOT NULL REFERENCES ekskul(id) ON DELETE CASCADE,
  siswa_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  predikat    TEXT CHECK (predikat IS NULL OR predikat IN ('Sangat Baik','Baik','Cukup','Kurang')),
  keterangan  TEXT,
  updated_at  TEXT NOT NULL,
  PRIMARY KEY (tahun_id, semester, ekskul_id, siswa_id)
);
CREATE TABLE IF NOT EXISTS portfolios (
  id           INTEGER PRIMARY KEY,
  siswa_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  course_id    INTEGER REFERENCES courses(id) ON DELETE SET NULL,
  judul        TEXT NOT NULL,
  deskripsi    TEXT,
  file_url     TEXT,
  link         TEXT,
  catatan_guru TEXT,
  nilai        REAL CHECK (nilai IS NULL OR (nilai >= 0 AND nilai <= 100)),
  dinilai_oleh INTEGER REFERENCES users(id) ON DELETE SET NULL,
  dinilai_at   TEXT,
  created_at   TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_rapor_nilai_siswa ON rapor_nilai(siswa_id, tahun_id, semester);
CREATE INDEX IF NOT EXISTS idx_ekskul_anggota_siswa ON ekskul_anggota(siswa_id, tahun_id);
CREATE INDEX IF NOT EXISTS idx_portfolio_siswa ON portfolios(siswa_id, created_at);
CREATE INDEX IF NOT EXISTS idx_portfolio_course ON portfolios(course_id);
`)
	// batas semester ganjil (tanggal terakhir semester ganjil); kosong = 31 Desember tahun pertama.
	// Migrasi bernomor: dijalankan setelah tabel academic_years dibuat.
	migrations = append(migrations, migration{name: "semester ganjil/genap", up: func(tx *sql.Tx) error {
		var n int
		tx.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('academic_years') WHERE name='akhir_ganjil'`).Scan(&n)
		if n > 0 {
			return nil
		}
		_, err := tx.Exec(`ALTER TABLE academic_years ADD COLUMN akhir_ganjil TEXT`)
		return err
	}})
	moduleRoutes = append(moduleRoutes, func(a *App, m *http.ServeMux) {
		r := func(roles string, fn func(*Ctx) error) http.HandlerFunc { return a.mod("rapor", roles, fn) }
		m.HandleFunc("GET /api/rapor/periode", r("any", apiRaporPeriode))
		m.HandleFunc("PUT /api/rapor/periode/{id}", r("admin", apiRaporPeriodeSave))
		m.HandleFunc("GET /api/rapor/tugas", r("admin,guru", apiRaporTugasku))
		m.HandleFunc("GET /api/rapor/mapel", r("admin,guru", apiRaporMapel))
		m.HandleFunc("PUT /api/rapor/mapel", r("admin,guru", apiRaporMapelSave))
		m.HandleFunc("POST /api/rapor/mapel/ai", r("admin,guru", apiRaporMapelAI))
		m.HandleFunc("GET /api/rapor/kelas", r("admin,guru", apiRaporKelas))
		m.HandleFunc("PUT /api/rapor/kelas/{id}/terbit", r("admin,guru", apiRaporTerbit))
		m.HandleFunc("PUT /api/rapor/siswa/{id}", r("admin,guru", apiRaporSiswaSave))
		m.HandleFunc("GET /api/rapor/siswa/{id}", r("any", apiRaporSiswa))
		m.HandleFunc("GET /api/rapor/siswa/{id}/docx", r("any", apiRaporDocx))
		m.HandleFunc("GET /api/rapor/me", r("siswa", apiRaporMe))
		m.HandleFunc("GET /api/ekskul", r("admin,guru", apiEkskulList))
		m.HandleFunc("POST /api/ekskul", r("admin", apiEkskulSave))
		m.HandleFunc("PUT /api/ekskul/{id}", r("admin", apiEkskulSave))
		m.HandleFunc("DELETE /api/ekskul/{id}", r("admin", apiEkskulDelete))
		m.HandleFunc("GET /api/ekskul/{id}/anggota", r("admin,guru", apiEkskulAnggota))
		m.HandleFunc("PUT /api/ekskul/{id}/anggota", r("admin,guru", apiEkskulAnggotaSave))
		m.HandleFunc("PUT /api/ekskul/{id}/nilai", r("admin,guru", apiEkskulNilaiSave))
		m.HandleFunc("GET /api/ekskul/{id}/calon", r("admin,guru", apiEkskulCalon))
		m.HandleFunc("GET /api/rapor/guru", r("admin", apiRaporGuru))
		m.HandleFunc("GET /api/portofolio", r("any", apiPortoList))
		m.HandleFunc("POST /api/portofolio", r("siswa", a.apiPortoSave))
		m.HandleFunc("PUT /api/portofolio/{id}", r("siswa", a.apiPortoSave))
		m.HandleFunc("DELETE /api/portofolio/{id}", r("any", apiPortoDelete))
		m.HandleFunc("PUT /api/portofolio/{id}/catatan", r("admin,guru", apiPortoCatatan))
	})
}

// ---------------------------------------------------------------------
// Periode (tahun ajaran + semester)
// ---------------------------------------------------------------------

type periode struct {
	tahun        int64
	nama         string
	smt          int
	mulai, akhir string // 2006-01-02, inklusif
}

// rentangSemester: semester ganjil = mulai tahun ajaran (bawaan 1 Juli) s.d. akhir ganjil (bawaan 31 Desember);
// genap = sehari setelah akhir ganjil s.d. selesai tahun ajaran (bawaan 30 Juni).
func rentangSemester(nama string, mulai, selesai, akhirGanjil sql.NullString, smt int) (string, string) {
	y1, _ := strconv.Atoi(strings.SplitN(nama, "/", 2)[0])
	if y1 == 0 {
		y1 = time.Now().Year()
	}
	tgl := func(v sql.NullString, def string) string {
		if len(v.String) >= 10 {
			if _, err := time.Parse("2006-01-02", v.String[:10]); err == nil {
				return v.String[:10]
			}
		}
		return def
	}
	ag := tgl(akhirGanjil, fmt.Sprintf("%d-12-31", y1))
	if smt == 1 {
		return tgl(mulai, fmt.Sprintf("%d-07-01", y1)), ag
	}
	t, _ := time.Parse("2006-01-02", ag)
	return t.AddDate(0, 0, 1).Format("2006-01-02"), tgl(selesai, fmt.Sprintf("%d-06-30", y1+1))
}

// periodeDari: tahun & semester dari query (?tahun=&smt=); kosong = tahun aktif & semester berjalan.
func (c *Ctx) periodeDari(tahun int64, smt int) (periode, error) {
	if tahun == 0 {
		tahun = activeYear(c.A.reader())
	}
	var p periode
	var mulai, selesai, ag sql.NullString
	if err := c.A.reader().QueryRow(`SELECT id, nama, mulai, selesai, akhir_ganjil FROM academic_years WHERE id=?`, tahun).Scan(&p.tahun, &p.nama, &mulai, &selesai, &ag); err != nil {
		return p, bad("Tahun ajaran tidak ditemukan. Atur di menu Tahun Ajaran.")
	}
	if smt != 1 && smt != 2 {
		smt = 1
		if _, akhir := rentangSemester(p.nama, mulai, selesai, ag, 1); localDate() > akhir {
			smt = 2
		}
	}
	p.smt = smt
	p.mulai, p.akhir = rentangSemester(p.nama, mulai, selesai, ag, smt)
	return p, nil
}

func (c *Ctx) periodeQuery() (periode, error) {
	smt, _ := strconv.Atoi(c.R.URL.Query().Get("smt"))
	return c.periodeDari(c.QInt("tahun"), smt)
}

func (p periode) label() string {
	if p.smt == 1 {
		return "Ganjil"
	}
	return "Genap"
}

func apiRaporPeriode(c *Ctx) error {
	rows, err := c.list(`SELECT id, nama, mulai, selesai, akhir_ganjil, aktif FROM academic_years ORDER BY nama DESC`)
	if err != nil {
		return err
	}
	for _, r := range rows {
		ns := func(k string) sql.NullString { s, ok := r[k].(string); return sql.NullString{String: s, Valid: ok} }
		nama, _ := r["nama"].(string)
		m1, a1 := rentangSemester(nama, ns("mulai"), ns("selesai"), ns("akhir_ganjil"), 1)
		m2, a2 := rentangSemester(nama, ns("mulai"), ns("selesai"), ns("akhir_ganjil"), 2)
		r["semester"] = []map[string]string{{"smt": "1", "nama": "Ganjil", "mulai": m1, "akhir": a1}, {"smt": "2", "nama": "Genap", "mulai": m2, "akhir": a2}}
	}
	p, _ := c.periodeDari(0, 0)
	return c.JSON(map[string]any{"tahun": rows, "aktif": map[string]any{"tahun": p.tahun, "smt": p.smt}})
}

func apiRaporPeriodeSave(c *Ctx) error {
	var in struct {
		AkhirGanjil *string `json:"akhir_ganjil"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	var v any
	if s := str(in.AkhirGanjil); s != nil {
		d, err := cleanDate(s.(string))
		if err != nil {
			return err
		}
		v = d
	}
	res, err := c.A.db.Exec(`UPDATE academic_years SET akhir_ganjil=? WHERE id=?`, v, c.ID("id"))
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return notFound()
	}
	return c.OK()
}

// ---------------------------------------------------------------------
// Hak akses
// ---------------------------------------------------------------------

func (c *Ctx) staffRapor() bool { return c.U.Role == "admin" || c.U.Has("rapor") }

// kelasMapel: kelas & tahun ajaran sebuah mapel.
func (c *Ctx) kelasMapel(courseID int64) (kelas, tahun int64, nama string) {
	c.A.reader().QueryRow(`SELECT k.id, COALESCE(k.tahun_id,0), k.nama FROM courses co JOIN classes k ON k.id=co.kelas_id WHERE co.id=?`, courseID).Scan(&kelas, &tahun, &nama)
	return
}

// kelasSiswaTahun: kelas siswa pada tahun ajaran tertentu (riwayat anggota kelas, atau kelas sekarang).
func (c *Ctx) kelasSiswaTahun(siswa, tahun int64) int64 {
	var k int64
	if c.A.reader().QueryRow(`SELECT m.kelas_id FROM class_members m JOIN classes k ON k.id=m.kelas_id WHERE m.siswa_id=? AND k.tahun_id=? LIMIT 1`, siswa, tahun).Scan(&k) == nil {
		return k
	}
	c.A.reader().QueryRow(`SELECT u.kelas_id FROM users u JOIN classes k ON k.id=u.kelas_id WHERE u.id=? AND k.tahun_id=?`, siswa, tahun).Scan(&k)
	return k
}

func (c *Ctx) terbit(p periode, kelas int64) bool {
	var n int
	c.A.reader().QueryRow(`SELECT COUNT(*) FROM rapor_terbit WHERE tahun_id=? AND semester=? AND kelas_id=?`, p.tahun, p.smt, kelas).Scan(&n)
	return n > 0
}

func terkunci() error {
	return bad("Rapor kelas ini sudah diterbitkan. Minta wali kelas membatalkan terbit terlebih dahulu untuk mengubah isinya.")
}

func (c *Ctx) siswaKelas(kelas int64) ([]map[string]any, error) {
	return c.list(`SELECT u.id, u.username, u.nama_lengkap, sp.nisn FROM users u LEFT JOIN student_profiles sp ON sp.siswa_id=u.id
		WHERE u.role='siswa' AND `+memberOf("?")+` ORDER BY u.nama_lengkap`, kelas, kelas)
}

// apiRaporTugasku: ringkasan tugas rapor milik guru (mapel yang diampu, kelas perwalian, ekskul binaan).
func apiRaporTugasku(c *Ctx) error {
	p, err := c.periodeQuery()
	if err != nil {
		return err
	}
	out := map[string]any{"tahun": p.tahun, "smt": p.smt, "periode": p.nama + " · " + p.label(), "mulai": p.mulai, "akhir": p.akhir, "staf": c.staffRapor()}
	set := func(k, q string, args ...any) {
		if err != nil {
			return
		}
		var r []map[string]any
		r, err = c.list(q, args...)
		out[k] = r
	}
	if c.staffRapor() {
		set("mapel", `SELECT co.id, co.nama, k.nama AS kelas_nama, k.id AS kelas_id FROM courses co JOIN classes k ON k.id=co.kelas_id WHERE k.tahun_id=? ORDER BY k.nama, co.nama`, p.tahun)
		set("kelas", `SELECT k.id, k.nama FROM classes k WHERE k.tahun_id=? ORDER BY k.nama`, p.tahun)
		set("ekskul", `SELECT e.id, e.nama, u.nama_lengkap AS pembina FROM ekskul e LEFT JOIN users u ON u.id=e.pembina_id ORDER BY e.nama`)
	} else {
		set("mapel", `SELECT co.id, co.nama, k.nama AS kelas_nama, k.id AS kelas_id FROM courses co JOIN classes k ON k.id=co.kelas_id WHERE k.tahun_id=? AND co.guru_id=? ORDER BY k.nama, co.nama`, p.tahun, c.U.ID)
		set("kelas", `SELECT k.id, k.nama FROM classes k WHERE k.tahun_id=? AND k.wali_id=? ORDER BY k.nama`, p.tahun, c.U.ID)
		set("ekskul", `SELECT e.id, e.nama, u.nama_lengkap AS pembina FROM ekskul e LEFT JOIN users u ON u.id=e.pembina_id WHERE e.pembina_id=? ORDER BY e.nama`, c.U.ID)
	}
	if err != nil {
		return err
	}
	return c.JSON(out)
}

// ---------------------------------------------------------------------
// Nilai & capaian per mapel (guru mapel)
// ---------------------------------------------------------------------

// dataMapelSemester: komponen nilai satu mapel dalam rentang semester (untuk nilai otomatis & saran deskripsi).
func (c *Ctx) dataMapelSemester(courseID int64, p periode) (asg, subs, att, abs []map[string]any, err error) {
	lo, hi := p.mulai, p.akhir+"T23:59:59Z"
	if asg, err = c.list(`SELECT id, judul, nilai_maks, deadline FROM assignments WHERE course_id=? AND COALESCE(deadline, created_at) BETWEEN ? AND ? ORDER BY created_at`, courseID, lo, hi); err != nil {
		return
	}
	if subs, err = c.list(`SELECT s.assignment_id, s.siswa_id, s.nilai FROM submissions s JOIN assignments a ON a.id=s.assignment_id
		WHERE a.course_id=? AND COALESCE(a.deadline, a.created_at) BETWEEN ? AND ?`, courseID, lo, hi); err != nil {
		return
	}
	if att, err = c.list(`SELECT t.quiz_id, q.judul, t.siswa_id, t.skor FROM quiz_attempts t JOIN quizzes q ON q.id=t.quiz_id
		WHERE q.course_id=? AND t.selesai=1 AND COALESCE(t.submitted_at, t.started_at) BETWEEN ? AND ?`, courseID, lo, hi); err != nil {
		return
	}
	abs, err = c.list(`SELECT siswa_id, tanggal, status FROM attendance WHERE course_id=? AND tanggal BETWEEN ? AND ?`, courseID, p.mulai, p.akhir)
	return
}

// saranCapaian: kalimat capaian dari data nilai siswa (tanpa AI): materi dengan nilai tertinggi dan terendah.
func saranCapaian(nama string, asg, subs, att []map[string]any, siswa int64) (tinggi, perlu string) {
	type item struct {
		judul string
		skor  float64
	}
	var it []item
	sub := map[any]map[string]any{}
	for _, s := range subs {
		if s["siswa_id"] == siswa {
			sub[s["assignment_id"]] = s
		}
	}
	for _, a := range asg {
		if s, ok := sub[a["id"]]; ok && s["nilai"] != nil {
			maks := toF(a["nilai_maks"])
			if maks == 0 {
				maks = 100
			}
			j, _ := a["judul"].(string)
			it = append(it, item{j, toF(s["nilai"]) / maks * 100})
		}
	}
	for _, t := range att {
		if t["siswa_id"] == siswa {
			j, _ := t["judul"].(string)
			it = append(it, item{j, toF(t["skor"])})
		}
	}
	if len(it) == 0 {
		return "", ""
	}
	sort.SliceStable(it, func(i, j int) bool { return it[i].skor > it[j].skor })
	// judul tugas/kuis → nama materi ("Latihan Bilangan Bulat" → "Bilangan Bulat")
	materi := func(j string) string {
		j = strings.TrimSpace(j)
		for _, p := range []string{"latihan ", "kuis ", "quiz ", "tugas ", "ulangan harian ", "ulangan ", "ujian ", "pr ", "penilaian harian ", "ph "} {
			if len(j) > len(p) && strings.EqualFold(j[:len(p)], p) {
				return strings.TrimSpace(j[len(p):])
			}
		}
		return j
	}
	ambil := func(xs []item) string {
		var n []string
		seen := map[string]bool{}
		for _, x := range xs {
			if m := materi(x.judul); m != "" && !seen[strings.ToLower(m)] {
				seen[strings.ToLower(m)] = true
				n = append(n, m)
			}
		}
		if len(n) > 2 {
			n = n[:2]
		}
		return "materi " + strings.Join(n, " dan ")
	}
	var baik, kurang []item
	for _, x := range it {
		if x.skor >= 80 {
			baik = append(baik, x)
		} else if x.skor < 70 {
			kurang = append(kurang, x)
		}
	}
	if len(baik) == 0 {
		baik = it[:1]
	}
	tinggi = "Menunjukkan penguasaan yang baik dalam " + ambil(baik) + "."
	if baik[0].skor >= 90 {
		tinggi = "Menunjukkan penguasaan yang sangat baik dalam " + ambil(baik) + "."
	}
	if len(kurang) > 0 {
		for i, j := 0, len(kurang)-1; i < j; i, j = i+1, j-1 {
			kurang[i], kurang[j] = kurang[j], kurang[i]
		}
		// materi yang sama sudah disebut dikuasai (mis. tugas bagus, kuis kurang): yang perlu adalah ketelitiannya
		if ambil(kurang[:1]) == ambil(baik[:1]) {
			perlu = "Perlu meningkatkan ketelitian dalam mengerjakan soal " + ambil(kurang) + "."
		} else {
			perlu = "Perlu bimbingan dalam " + ambil(kurang) + "."
		}
	} else {
		perlu = "Perlu terus meningkatkan konsistensi belajar " + nama + "."
	}
	return
}

func apiRaporMapel(c *Ctx) error {
	cid := c.QInt("course_id")
	if !c.canManage(cid) && !c.staffRapor() {
		return forbidden()
	}
	p, err := c.periodeQuery()
	if err != nil {
		return err
	}
	kelas, tahunKelas, kelasNama := c.kelasMapel(cid)
	if kelas == 0 {
		return notFound()
	}
	if tahunKelas != 0 && tahunKelas != p.tahun {
		return bad("Mapel ini milik tahun ajaran lain. Pilih tahun ajaran yang sesuai.")
	}
	var mapel string
	c.A.reader().QueryRow(`SELECT nama FROM courses WHERE id=?`, cid).Scan(&mapel)
	st, err := c.siswaKelas(kelas)
	if err != nil {
		return err
	}
	asg, subs, att, abs, err := c.dataMapelSemester(cid, p)
	if err != nil {
		return err
	}
	saved, err := c.list(`SELECT siswa_id, nilai, tinggi, perlu, updated_at FROM rapor_nilai WHERE tahun_id=? AND semester=? AND course_id=?`, p.tahun, p.smt, cid)
	if err != nil {
		return err
	}
	porto, err := c.list(`SELECT siswa_id, COUNT(*) AS n FROM portfolios WHERE course_id=? AND created_at BETWEEN ? AND ? GROUP BY siswa_id`, cid, p.mulai, p.akhir+"T23:59:59Z")
	if err != nil {
		return err
	}
	byS := map[any]map[string]any{}
	for _, r := range saved {
		byS[r["siswa_id"]] = r
	}
	nPorto := map[any]any{}
	for _, r := range porto {
		nPorto[r["siswa_id"]] = r["n"]
	}
	now := time.Now()
	for _, s := range st {
		id, _ := s["id"].(int64)
		s["otomatis"] = skorDariRekap(asg, subs, att, abs, id, now)
		s["saran_tinggi"], s["saran_perlu"] = saranCapaian("", asg, subs, att, id)
		s["portofolio"] = nPorto[id]
		if r, ok := byS[id]; ok {
			s["nilai"], s["tinggi"], s["perlu"], s["tersimpan"] = r["nilai"], r["tinggi"], r["perlu"], r["updated_at"]
		}
	}
	return c.JSON(map[string]any{"mapel": mapel, "kelas": kelasNama, "kelas_id": kelas, "periode": p.nama + " · " + p.label(),
		"tahun": p.tahun, "smt": p.smt, "mulai": p.mulai, "akhir": p.akhir, "terbit": c.terbit(p, kelas), "siswa": st,
		"jumlah_tugas": len(asg), "jumlah_kuis": jumlahKuis(att)})
}

func cutRunes(s string, n int) string {
	s = strings.TrimSpace(s)
	if r := []rune(s); len(r) > n {
		return string(r[:n])
	}
	return s
}

func apiRaporMapelSave(c *Ctx) error {
	var in struct {
		CourseID int64 `json:"course_id"`
		Tahun    int64 `json:"tahun"`
		Smt      int   `json:"smt"`
		Items    []struct {
			SiswaID int64    `json:"siswa_id"`
			Nilai   *float64 `json:"nilai"`
			Tinggi  string   `json:"tinggi"`
			Perlu   string   `json:"perlu"`
		} `json:"items"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	if !c.canManage(in.CourseID) && !c.staffRapor() {
		return forbidden()
	}
	p, err := c.periodeDari(in.Tahun, in.Smt)
	if err != nil {
		return err
	}
	kelas, tahunKelas, _ := c.kelasMapel(in.CourseID)
	if kelas == 0 {
		return notFound()
	}
	if tahunKelas != 0 && tahunKelas != p.tahun {
		return bad("Mapel ini milik tahun ajaran lain.")
	}
	if c.terbit(p, kelas) {
		return terkunci()
	}
	tx, err := c.A.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, it := range in.Items {
		if it.Nilai != nil && (*it.Nilai < 0 || *it.Nilai > 100) {
			return bad("Nilai rapor harus 0 sampai 100.")
		}
		var ok int
		tx.QueryRow(`SELECT COUNT(*) FROM users u WHERE u.id=? AND u.role='siswa' AND `+memberOf("?"), it.SiswaID, kelas, kelas).Scan(&ok)
		if ok == 0 {
			continue // bukan siswa kelas mapel ini
		}
		var nilai any
		if it.Nilai != nil {
			nilai = *it.Nilai
		}
		if _, err := tx.Exec(`INSERT INTO rapor_nilai (tahun_id, semester, course_id, siswa_id, nilai, tinggi, perlu, updated_by, updated_at) VALUES (?,?,?,?,?,?,?,?,?)
			ON CONFLICT (tahun_id, semester, course_id, siswa_id) DO UPDATE SET nilai=excluded.nilai, tinggi=excluded.tinggi, perlu=excluded.perlu,
			updated_by=excluded.updated_by, updated_at=excluded.updated_at`,
			p.tahun, p.smt, in.CourseID, it.SiswaID, nilai, cutRunes(it.Tinggi, 500), cutRunes(it.Perlu, 500), c.U.ID, nowStr()); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.OK()
}

// apiRaporMapelAI: saran deskripsi capaian untuk satu kelas sekaligus (satu kali kuota AI). Nama siswa tidak
// dikirim ke penyedia AI; hanya nomor urut dan nilai per materi.
func apiRaporMapelAI(c *Ctx) error {
	var in struct {
		CourseID int64 `json:"course_id"`
		Tahun    int64 `json:"tahun"`
		Smt      int   `json:"smt"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	if !c.canManage(in.CourseID) && !c.staffRapor() {
		return forbidden()
	}
	p, err := c.periodeDari(in.Tahun, in.Smt)
	if err != nil {
		return err
	}
	kelas, _, kelasNama := c.kelasMapel(in.CourseID)
	if kelas == 0 {
		return notFound()
	}
	var mapel string
	c.A.reader().QueryRow(`SELECT nama FROM courses WHERE id=?`, in.CourseID).Scan(&mapel)
	st, err := c.siswaKelas(kelas)
	if err != nil {
		return err
	}
	asg, subs, att, _, err := c.dataMapelSemester(in.CourseID, p)
	if err != nil {
		return err
	}
	if len(asg) == 0 && len(att) == 0 {
		return bad("Belum ada tugas atau kuis pada semester ini, jadi belum ada bahan untuk deskripsi capaian.")
	}
	var b strings.Builder
	ids := map[int]int64{}
	for i, s := range st {
		id, _ := s["id"].(int64)
		ids[i+1] = id
		fmt.Fprintf(&b, "Siswa %d:", i+1)
		sub := map[any]map[string]any{}
		for _, x := range subs {
			if x["siswa_id"] == id {
				sub[x["assignment_id"]] = x
			}
		}
		n := 0
		for _, a := range asg {
			if x, ok := sub[a["id"]]; ok && x["nilai"] != nil {
				maks := toF(a["nilai_maks"])
				if maks == 0 {
					maks = 100
				}
				fmt.Fprintf(&b, " [%s: %.0f]", a["judul"], toF(x["nilai"])/maks*100)
				n++
			}
		}
		for _, t := range att {
			if t["siswa_id"] == id {
				fmt.Fprintf(&b, " [%s: %.0f]", t["judul"], toF(t["skor"]))
				n++
			}
		}
		if n == 0 {
			b.WriteString(" (belum ada nilai)")
		}
		b.WriteString("\n")
	}
	fase, _ := rppFase(kelasNama)
	prompt := "Tulis deskripsi capaian kompetensi rapor Kurikulum Merdeka untuk mata pelajaran " + mapel + " kelas " + kelasNama +
		" (" + fase + "), semester " + p.label() + " " + p.nama + ".\n" +
		"Untuk setiap siswa buat dua kalimat singkat (maksimal 25 kata per kalimat), bahasa positif dan spesifik materi:\n" +
		"- tinggi: capaian yang paling dikuasai, diawali \"Menunjukkan penguasaan yang baik/sangat baik dalam ...\"\n" +
		"- perlu: hal yang perlu ditingkatkan, diawali \"Perlu bimbingan dalam ...\" atau \"Perlu meningkatkan ...\"\n" +
		"Gunakan hanya materi yang tercantum. Jangan menyebut angka nilai. Siswa tanpa nilai: kosongkan kedua kalimat.\n" +
		"Jawab HANYA dengan JSON array, tanpa teks lain: [{\"no\":1,\"tinggi\":\"...\",\"perlu\":\"...\"}, ...]\n\nData nilai per materi (0–100):\n" + b.String()
	text, err := c.aiSelesai("Anda membantu guru menulis deskripsi capaian kompetensi rapor sesuai Panduan Pembelajaran dan Asesmen Kurikulum Merdeka. Jawab dalam Bahasa Indonesia baku.", prompt)
	if err != nil {
		return err
	}
	if i, j := strings.Index(text, "["), strings.LastIndex(text, "]"); i >= 0 && j > i {
		text = text[i : j+1]
	}
	var hasil []struct {
		No     int    `json:"no"`
		Tinggi string `json:"tinggi"`
		Perlu  string `json:"perlu"`
	}
	if err := json.Unmarshal([]byte(text), &hasil); err != nil {
		return bad("Jawaban AI tidak bisa dibaca. Coba lagi, atau pakai saran otomatis dari nilai.")
	}
	out := []map[string]any{}
	for _, h := range hasil {
		if id, ok := ids[h.No]; ok {
			out = append(out, map[string]any{"siswa_id": id, "tinggi": cutRunes(h.Tinggi, 500), "perlu": cutRunes(h.Perlu, 500)})
		}
	}
	return c.JSON(map[string]any{"saran": out})
}

// aiSelesai: satu permintaan AI tanpa streaming (dipakai e-Rapor). Memakai kuota harian yang sama.
func (c *Ctx) aiSelesai(system, prompt string) (string, error) {
	if !modulesEnabled(c.A.db)["ai"] {
		return "", bad("Modul Asisten AI sedang dinonaktifkan.")
	}
	cfg := getConfig(c.A.db, "ai")
	base, _ := cfg["base_url"].(string)
	model, _ := cfg["model"].(string)
	if base == "" || model == "" {
		return "", bad("Asisten AI belum diatur. Admin perlu mengisi endpoint dan model di Pengaturan → Asisten AI.")
	}
	release, err := c.aiReserve(cfg)
	if err != nil {
		return "", err
	}
	defer release()
	body := map[string]any{"model": model, "stream": false, "temperature": 0.3,
		"messages": []map[string]string{{"role": "system", "content": system}, {"role": "user", "content": prompt}}}
	ctx, cancel := context.WithTimeout(c.R.Context(), 3*time.Minute)
	defer cancel()
	res, err := aiRequest(ctx, "POST", base+"/chat/completions", kvGet(c.A.db, "ai_key"), body)
	if err != nil {
		return "", bad("Tidak bisa menghubungi penyedia AI. Periksa koneksi internet server, lalu coba lagi.")
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return "", upstreamErr(res)
	}
	var out strings.Builder
	aiReadStream(res, func(s string) { out.WriteString(s) })
	text := strings.TrimSpace(out.String())
	if text == "" {
		return "", bad("Tidak ada jawaban dari model AI. Coba lagi atau pilih model lain.")
	}
	c.aiCountUse()
	return text, nil
}

// ---------------------------------------------------------------------
// Kokurikuler, ketidakhadiran, catatan, terbit (wali kelas)
// ---------------------------------------------------------------------

// 8 dimensi profil lulusan (Permendikdasmen 13/2025)
var dimensiProfil = []struct{ Key, Nama string }{
	{"keimanan", "Keimanan dan Ketakwaan terhadap Tuhan YME"}, {"kewargaan", "Kewargaan"}, {"penalaran", "Penalaran Kritis"},
	{"kreativitas", "Kreativitas"}, {"kolaborasi", "Kolaborasi"}, {"kemandirian", "Kemandirian"}, {"kesehatan", "Kesehatan"}, {"komunikasi", "Komunikasi"},
}
var predikatKoku = map[string]string{"SB": "sangat berkembang", "BSH": "berkembang sesuai harapan", "MB": "mulai berkembang", "BB": "belum berkembang"}

// deskripsiKoku: kalimat kokurikuler dari predikat per dimensi (bila wali tidak menulis sendiri).
func deskripsiKoku(k map[string]string) string {
	grup := map[string][]string{}
	for _, d := range dimensiProfil {
		if p := k[d.Key]; predikatKoku[p] != "" {
			grup[p] = append(grup[p], strings.ToLower(d.Nama))
		}
	}
	var parts []string
	for _, p := range []string{"SB", "BSH", "MB", "BB"} {
		if len(grup[p]) > 0 {
			n := grup[p]
			s := strings.Join(n[:len(n)-1], ", ")
			if s != "" {
				s += " dan "
			}
			parts = append(parts, predikatKoku[p]+" dalam "+s+n[len(n)-1])
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "Ananda " + strings.Join(parts, "; ") + "."
}

func (c *Ctx) bolehWali(kelas int64) bool { return c.staffRapor() || c.isWali(kelas) }

// absenSemester: jumlah sakit/izin/alfa dari absensi harian dalam rentang semester.
func (c *Ctx) absenSemester(p periode, kelas int64) (map[int64][3]int, error) {
	rows, err := c.A.reader().Query(`SELECT d.siswa_id, d.status, COUNT(*) FROM daily_attendance d JOIN users u ON u.id=d.siswa_id
		WHERE d.tanggal BETWEEN ? AND ? AND d.status IN ('sakit','izin','alfa') AND `+memberOf("?")+` GROUP BY d.siswa_id, d.status`, p.mulai, p.akhir, kelas, kelas)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int64][3]int{}
	for rows.Next() {
		var id int64
		var st string
		var n int
		if err := rows.Scan(&id, &st, &n); err != nil {
			return nil, err
		}
		v := out[id]
		switch st {
		case "sakit":
			v[0] = n
		case "izin":
			v[1] = n
		case "alfa":
			v[2] = n
		}
		out[id] = v
	}
	return out, rows.Err()
}

func apiRaporKelas(c *Ctx) error {
	kelas := c.QInt("kelas_id")
	if !c.bolehWali(kelas) {
		return forbidden()
	}
	p, err := c.periodeQuery()
	if err != nil {
		return err
	}
	var nama string
	var wali sql.NullString
	if c.A.reader().QueryRow(`SELECT k.nama, u.nama_lengkap FROM classes k LEFT JOIN users u ON u.id=k.wali_id WHERE k.id=?`, kelas).Scan(&nama, &wali) != nil {
		return notFound()
	}
	st, err := c.siswaKelas(kelas)
	if err != nil {
		return err
	}
	mapel, err := c.list(`SELECT co.id, co.nama, g.nama_lengkap AS guru FROM courses co LEFT JOIN users g ON g.id=co.guru_id WHERE co.kelas_id=? ORDER BY co.nama`, kelas)
	if err != nil {
		return err
	}
	isi, err := c.list(`SELECT course_id, COUNT(*) AS n FROM rapor_nilai WHERE tahun_id=? AND semester=? AND nilai IS NOT NULL
		AND course_id IN (SELECT id FROM courses WHERE kelas_id=?) GROUP BY course_id`, p.tahun, p.smt, kelas)
	if err != nil {
		return err
	}
	nIsi := map[any]any{}
	for _, r := range isi {
		nIsi[r["course_id"]] = r["n"]
	}
	for _, m := range mapel {
		m["terisi"] = nIsi[m["id"]]
	}
	data, err := c.list(`SELECT siswa_id, kokurikuler, koku_desk, catatan, keputusan, sakit, izin, alfa FROM rapor_siswa WHERE tahun_id=? AND semester=?`, p.tahun, p.smt)
	if err != nil {
		return err
	}
	byS := map[any]map[string]any{}
	for _, r := range data {
		byS[r["siswa_id"]] = r
	}
	absen, err := c.absenSemester(p, kelas)
	if err != nil {
		return err
	}
	for _, s := range st {
		id, _ := s["id"].(int64)
		a := absen[id]
		s["absen_otomatis"] = map[string]int{"sakit": a[0], "izin": a[1], "alfa": a[2]}
		if r, ok := byS[id]; ok {
			k := map[string]string{}
			ks, _ := r["kokurikuler"].(string)
			json.Unmarshal([]byte(ks), &k)
			r["kokurikuler"] = k
			delete(r, "siswa_id")
			s["rapor"] = r
		}
	}
	var tglTerbit sql.NullString
	c.A.reader().QueryRow(`SELECT tanggal FROM rapor_terbit WHERE tahun_id=? AND semester=? AND kelas_id=?`, p.tahun, p.smt, kelas).Scan(&tglTerbit)
	return c.JSON(map[string]any{"kelas": nama, "kelas_id": kelas, "wali": wali.String, "periode": p.nama + " · " + p.label(), "tahun": p.tahun, "smt": p.smt,
		"mulai": p.mulai, "akhir": p.akhir, "siswa": st, "mapel": mapel, "terbit": tglTerbit.String, "dimensi": dimensiProfil})
}

func apiRaporSiswaSave(c *Ctx) error {
	siswa := c.ID("id")
	var in struct {
		Tahun       int64             `json:"tahun"`
		Smt         int               `json:"smt"`
		Kokurikuler map[string]string `json:"kokurikuler"`
		KokuDesk    string            `json:"koku_desk"`
		Catatan     string            `json:"catatan"`
		Keputusan   string            `json:"keputusan"`
		Sakit       *int              `json:"sakit"`
		Izin        *int              `json:"izin"`
		Alfa        *int              `json:"alfa"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	p, err := c.periodeDari(in.Tahun, in.Smt)
	if err != nil {
		return err
	}
	kelas := c.kelasSiswaTahun(siswa, p.tahun)
	if kelas == 0 || !c.bolehWali(kelas) {
		return forbidden()
	}
	if c.terbit(p, kelas) {
		return terkunci()
	}
	k := map[string]string{}
	for _, d := range dimensiProfil {
		if v := in.Kokurikuler[d.Key]; predikatKoku[v] != "" {
			k[d.Key] = v
		}
	}
	kb, _ := json.Marshal(k)
	var kep any
	switch in.Keputusan {
	case "":
	case "naik", "tinggal", "lulus":
		if p.smt != 2 {
			return bad("Keputusan kenaikan kelas hanya diisi di semester genap.")
		}
		kep = in.Keputusan
	default:
		return bad("Keputusan tidak valid.")
	}
	n := func(v *int) any {
		if v == nil || *v < 0 {
			return nil
		}
		if *v > 300 {
			return 300
		}
		return *v
	}
	_, err = c.A.db.Exec(`INSERT INTO rapor_siswa (tahun_id, semester, siswa_id, kokurikuler, koku_desk, catatan, keputusan, sakit, izin, alfa, updated_by, updated_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT (tahun_id, semester, siswa_id) DO UPDATE SET kokurikuler=excluded.kokurikuler, koku_desk=excluded.koku_desk, catatan=excluded.catatan,
		keputusan=excluded.keputusan, sakit=excluded.sakit, izin=excluded.izin, alfa=excluded.alfa, updated_by=excluded.updated_by, updated_at=excluded.updated_at`,
		p.tahun, p.smt, siswa, string(kb), cutRunes(in.KokuDesk, 800), cutRunes(in.Catatan, 800), kep, n(in.Sakit), n(in.Izin), n(in.Alfa), c.U.ID, nowStr())
	if err != nil {
		return err
	}
	return c.OK()
}

func apiRaporTerbit(c *Ctx) error {
	kelas := c.ID("id")
	var in struct {
		Tahun  int64 `json:"tahun"`
		Smt    int   `json:"smt"`
		Terbit bool  `json:"terbit"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	if !c.bolehWali(kelas) {
		return forbidden()
	}
	p, err := c.periodeDari(in.Tahun, in.Smt)
	if err != nil {
		return err
	}
	if !in.Terbit {
		_, err := c.A.db.Exec(`DELETE FROM rapor_terbit WHERE tahun_id=? AND semester=? AND kelas_id=?`, p.tahun, p.smt, kelas)
		if err != nil {
			return err
		}
		c.Note = "batal terbit"
		return c.OK()
	}
	// semua mapel kelas harus sudah punya nilai untuk setiap siswa
	var kosong []string
	rows, err := c.A.reader().Query(`SELECT co.nama, (SELECT COUNT(*) FROM users u WHERE u.role='siswa' AND `+memberOf("co.kelas_id")+`) -
		(SELECT COUNT(*) FROM rapor_nilai r WHERE r.tahun_id=? AND r.semester=? AND r.course_id=co.id AND r.nilai IS NOT NULL) FROM courses co WHERE co.kelas_id=? ORDER BY co.nama`, p.tahun, p.smt, kelas)
	if err != nil {
		return err
	}
	for rows.Next() {
		var nama string
		var sisa int
		if rows.Scan(&nama, &sisa) == nil && sisa > 0 {
			kosong = append(kosong, fmt.Sprintf("%s (%d siswa)", nama, sisa))
		}
	}
	rows.Close()
	if len(kosong) > 0 {
		return bad("Nilai rapor belum lengkap: " + strings.Join(kosong, ", ") + ". Minta guru mapel mengisinya dulu.")
	}
	if _, err := c.A.db.Exec(`INSERT INTO rapor_terbit (tahun_id, semester, kelas_id, tanggal, oleh) VALUES (?,?,?,?,?) ON CONFLICT DO NOTHING`,
		p.tahun, p.smt, kelas, localDate(), c.U.ID); err != nil {
		return err
	}
	c.Note = "terbit"
	return c.OK()
}

// ---------------------------------------------------------------------
// Rapor lengkap (cetak, Word, siswa)
// ---------------------------------------------------------------------

type raporLengkap struct {
	Sekolah    map[string]any   `json:"sekolah"`
	Siswa      map[string]any   `json:"siswa"`
	Kelas      string           `json:"kelas"`
	Fase       string           `json:"fase"`
	Periode    string           `json:"periode"`
	Semester   string           `json:"semester"`
	TahunNama  string           `json:"tahun_ajaran"`
	Wali       map[string]any   `json:"wali"`
	Mapel      []map[string]any `json:"mapel"`
	Koku       []map[string]any `json:"kokurikuler"`
	KokuDesk   string           `json:"koku_desk"`
	Ekskul     []map[string]any `json:"ekskul"`
	Absen      map[string]int   `json:"ketidakhadiran"`
	Catatan    string           `json:"catatan"`
	Keputusan  string           `json:"keputusan"`
	Terbit     string           `json:"terbit"`
	Genap      bool             `json:"genap"`
	TingkatBrk string           `json:"tingkat_berikut"`
}

func (c *Ctx) raporLengkap(siswa int64, p periode) (*raporLengkap, error) {
	kelas := c.kelasSiswaTahun(siswa, p.tahun)
	if kelas == 0 {
		return nil, bad("Siswa tidak terdaftar di kelas mana pun pada tahun ajaran ini.")
	}
	r := &raporLengkap{Periode: p.nama + " · " + p.label(), Semester: p.label(), TahunNama: p.nama, Genap: p.smt == 2}
	s, err := c.A.settingsData()
	if err != nil {
		return nil, err
	}
	r.Sekolah = map[string]any{"nama": s["nama_sekolah"], "alamat": s["alamat"], "telepon": s["telepon"], "logo_url": s["logo_url"],
		"kepala": s["kepala_sekolah"], "nip_kepala": s["nip_kepala"], "jenjang": s["jenjang"]}
	ss, err := c.list(`SELECT u.id, u.username, u.nama_lengkap, sp.nisn FROM users u LEFT JOIN student_profiles sp ON sp.siswa_id=u.id WHERE u.id=?`, siswa)
	if err != nil || len(ss) == 0 {
		return nil, notFound()
	}
	r.Siswa = ss[0]
	var waliNama, waliNIP sql.NullString
	c.A.reader().QueryRow(`SELECT k.nama, u.nama_lengkap, u.nip FROM classes k LEFT JOIN users u ON u.id=k.wali_id WHERE k.id=?`, kelas).Scan(&r.Kelas, &waliNama, &waliNIP)
	r.Wali = map[string]any{"nama": waliNama.String, "nip": waliNIP.String}
	r.Fase, _ = rppFase(r.Kelas)
	if t := tingkatDariNama(r.Kelas); t > 0 {
		r.TingkatBrk = strconv.Itoa(t + 1)
	}
	r.Mapel, err = c.list(`SELECT co.nama AS mapel, r.nilai, r.tinggi, r.perlu FROM courses co
		LEFT JOIN rapor_nilai r ON r.course_id=co.id AND r.siswa_id=? AND r.tahun_id=? AND r.semester=? WHERE co.kelas_id=? ORDER BY co.nama`, siswa, p.tahun, p.smt, kelas)
	if err != nil {
		return nil, err
	}
	var koku, desk, catatan, kep sql.NullString
	var sakit, izin, alfa sql.NullInt64
	c.A.reader().QueryRow(`SELECT kokurikuler, koku_desk, catatan, keputusan, sakit, izin, alfa FROM rapor_siswa WHERE tahun_id=? AND semester=? AND siswa_id=?`,
		p.tahun, p.smt, siswa).Scan(&koku, &desk, &catatan, &kep, &sakit, &izin, &alfa)
	k := map[string]string{}
	json.Unmarshal([]byte(koku.String), &k)
	for _, d := range dimensiProfil {
		r.Koku = append(r.Koku, map[string]any{"dimensi": d.Nama, "predikat": k[d.Key], "keterangan": predikatKoku[k[d.Key]]})
	}
	r.KokuDesk = desk.String
	if r.KokuDesk == "" {
		r.KokuDesk = deskripsiKoku(k)
	}
	r.Catatan, r.Keputusan = catatan.String, kep.String
	r.Ekskul, err = c.list(`SELECT e.nama, re.predikat, re.keterangan FROM ekskul_anggota a JOIN ekskul e ON e.id=a.ekskul_id
		LEFT JOIN rapor_ekskul re ON re.ekskul_id=a.ekskul_id AND re.siswa_id=a.siswa_id AND re.tahun_id=? AND re.semester=?
		WHERE a.siswa_id=? AND a.tahun_id=? ORDER BY e.nama`, p.tahun, p.smt, siswa, p.tahun)
	if err != nil {
		return nil, err
	}
	auto, err := c.absenSemester(p, kelas)
	if err != nil {
		return nil, err
	}
	a := auto[siswa]
	pilih := func(v sql.NullInt64, def int) int {
		if v.Valid {
			return int(v.Int64)
		}
		return def
	}
	r.Absen = map[string]int{"sakit": pilih(sakit, a[0]), "izin": pilih(izin, a[1]), "alfa": pilih(alfa, a[2])}
	var tgl sql.NullString
	c.A.reader().QueryRow(`SELECT tanggal FROM rapor_terbit WHERE tahun_id=? AND semester=? AND kelas_id=?`, p.tahun, p.smt, kelas).Scan(&tgl)
	r.Terbit = tgl.String
	return r, nil
}

// bolehLihatRapor: wali/staf/guru mapel kelas itu; siswa sendiri hanya setelah terbit.
func (c *Ctx) bolehLihatRapor(siswa int64, p periode) error {
	kelas := c.kelasSiswaTahun(siswa, p.tahun)
	switch {
	case c.U.Role == "siswa":
		if siswa != c.U.ID {
			return forbidden()
		}
		if !c.terbit(p, kelas) {
			return bad("Rapor semester ini belum diterbitkan wali kelas.")
		}
		return nil
	case c.bolehWali(kelas), c.U.Role == "guru" && c.teachesClass(kelas):
		return nil
	}
	return forbidden()
}

func apiRaporSiswa(c *Ctx) error {
	p, err := c.periodeQuery()
	if err != nil {
		return err
	}
	id := c.ID("id")
	if err := c.bolehLihatRapor(id, p); err != nil {
		return err
	}
	r, err := c.raporLengkap(id, p)
	if err != nil {
		return err
	}
	return c.JSON(r)
}

func apiRaporMe(c *Ctx) error {
	// daftar semester yang rapornya sudah terbit untuk siswa ini
	rows, err := c.list(`SELECT t.tahun_id, y.nama, t.semester, t.tanggal FROM rapor_terbit t JOIN academic_years y ON y.id=t.tahun_id
		WHERE t.kelas_id IN (SELECT kelas_id FROM class_members WHERE siswa_id=? UNION SELECT kelas_id FROM users WHERE id=? AND kelas_id IS NOT NULL)
		ORDER BY y.nama DESC, t.semester DESC`, c.U.ID, c.U.ID)
	if err != nil {
		return err
	}
	return c.JSON(map[string]any{"terbit": rows})
}

func fmtNilai(v any) string {
	if v == nil {
		return "–"
	}
	f := toF(v)
	if f == float64(int64(f)) {
		return strconv.FormatInt(int64(f), 10)
	}
	return strings.Replace(strconv.FormatFloat(f, 'f', 1, 64), ".", ",", 1)
}

// raporMarkdown: isi rapor sebagai Markdown untuk berkas Word (memakai pembuat .docx yang sama dengan Asisten AI).
func raporMarkdown(r *raporLengkap) string {
	s := func(m map[string]any, k string) string { v, _ := m[k].(string); return v }
	md := func(t string) string { return strings.NewReplacer("|", "/", "\n", " ").Replace(t) }
	var b strings.Builder
	b.WriteString("# LAPORAN HASIL BELAJAR\n\n")
	fmt.Fprintf(&b, "| | | | |\n|---|---|---|---|\n| Nama | %s | Kelas | %s |\n| NIS / NISN | %s / %s | Fase | %s |\n| Sekolah | %s | Semester | %s |\n| Alamat | %s | Tahun ajaran | %s |\n\n",
		md(s(r.Siswa, "nama_lengkap")), md(r.Kelas), md(s(r.Siswa, "username")), md(s(r.Siswa, "nisn")), md(r.Fase),
		md(s(r.Sekolah, "nama")), r.Semester, md(s(r.Sekolah, "alamat")), r.TahunNama)
	b.WriteString("## A. Nilai dan Capaian Kompetensi\n\n| No | Mata Pelajaran | Nilai Akhir | Capaian Kompetensi |\n|---|---|---|---|\n")
	for i, m := range r.Mapel {
		cap := strings.TrimSpace(s(m, "tinggi") + " " + s(m, "perlu"))
		fmt.Fprintf(&b, "| %d | %s | %s | %s |\n", i+1, md(s(m, "mapel")), fmtNilai(m["nilai"]), md(cap))
	}
	b.WriteString("\n## B. Kokurikuler\n\n")
	if r.KokuDesk != "" {
		b.WriteString(md(r.KokuDesk) + "\n\n")
	} else {
		b.WriteString("–\n\n")
	}
	b.WriteString("## C. Ekstrakurikuler\n\n")
	if len(r.Ekskul) == 0 {
		b.WriteString("Tidak mengikuti ekstrakurikuler.\n")
	} else {
		b.WriteString("| No | Kegiatan | Predikat | Keterangan |\n|---|---|---|---|\n")
	}
	for i, e := range r.Ekskul {
		fmt.Fprintf(&b, "| %d | %s | %s | %s |\n", i+1, md(s(e, "nama")), md(s(e, "predikat")), md(s(e, "keterangan")))
	}
	fmt.Fprintf(&b, "\n## D. Ketidakhadiran\n\n| | |\n|---|---|\n| Sakit | %d hari |\n| Izin | %d hari |\n| Tanpa keterangan | %d hari |\n\n", r.Absen["sakit"], r.Absen["izin"], r.Absen["alfa"])
	b.WriteString("## E. Catatan Wali Kelas\n\n")
	if r.Catatan != "" {
		b.WriteString(md(r.Catatan) + "\n\n")
	} else {
		b.WriteString("–\n\n")
	}
	if r.Genap && r.Keputusan != "" {
		b.WriteString("## F. Keputusan\n\n")
		switch r.Keputusan {
		case "naik":
			b.WriteString("Berdasarkan pencapaian seluruh kompetensi, peserta didik dinyatakan **naik ke kelas " + r.TingkatBrk + "**.\n\n")
		case "tinggal":
			b.WriteString("Berdasarkan pencapaian seluruh kompetensi, peserta didik dinyatakan **tinggal di kelas " + md(r.Kelas) + "**.\n\n")
		case "lulus":
			b.WriteString("Peserta didik dinyatakan **LULUS**.\n\n")
		}
	}
	b.WriteString("Orang Tua/Wali Murid: ................................\n")
	return b.String()
}

func apiRaporDocx(c *Ctx) error {
	p, err := c.periodeQuery()
	if err != nil {
		return err
	}
	id := c.ID("id")
	if err := c.bolehLihatRapor(id, p); err != nil {
		return err
	}
	r, err := c.raporLengkap(id, p)
	if err != nil {
		return err
	}
	sv := func(m map[string]any, k string) string { v, _ := m[k].(string); return v }
	tgl := r.Terbit
	if tgl == "" {
		tgl = localDate()
	}
	o := docxOpts{Judul: "Rapor " + sv(r.Siswa, "nama_lengkap"), Kop: true, NamaSek: sv(r.Sekolah, "nama"), AlamatSek: sv(r.Sekolah, "alamat"),
		TeleponSek: sv(r.Sekolah, "telepon"), LogoPath: c.A.localLogo(sv(r.Sekolah, "logo_url")), TTD: true, Tanggal: tanggalPanjang(tgl),
		Guru: sv(r.Wali, "nama"), NIPGuru: sv(r.Wali, "nip"), Kepsek: sv(r.Sekolah, "kepala"), NIPKepsek: sv(r.Sekolah, "nip_kepala")}
	b, err := buildDocx(raporMarkdown(r), o)
	if err != nil {
		return err
	}
	name := "Rapor " + sv(r.Siswa, "nama_lengkap") + " " + r.Kelas + " " + strings.ReplaceAll(r.TahunNama, "/", "-") + " " + r.Semester
	name = strings.Map(func(x rune) rune {
		if strings.ContainsRune(`/\:*?"<>|`, x) || x < 32 {
			return '-'
		}
		return x
	}, name)
	h := c.W.Header()
	h.Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	h.Set("Content-Disposition", `attachment; filename="rapor.docx"; filename*=UTF-8''`+url.PathEscape(name+".docx"))
	h.Set("Cache-Control", "no-store")
	_, err = c.W.Write(b)
	return err
}

// ---------------------------------------------------------------------
// Ekstrakurikuler
// ---------------------------------------------------------------------

func (c *Ctx) pembinaEkskul(id int64) bool {
	if c.staffRapor() {
		return true
	}
	var p sql.NullInt64
	c.A.reader().QueryRow(`SELECT pembina_id FROM ekskul WHERE id=?`, id).Scan(&p)
	return p.Valid && p.Int64 == c.U.ID
}

func apiEkskulList(c *Ctx) error {
	p, err := c.periodeQuery()
	if err != nil {
		return err
	}
	rows, err := c.list(`SELECT e.id, e.nama, e.pembina_id, u.nama_lengkap AS pembina,
		(SELECT COUNT(*) FROM ekskul_anggota a WHERE a.ekskul_id=e.id AND a.tahun_id=?) AS anggota FROM ekskul e LEFT JOIN users u ON u.id=e.pembina_id ORDER BY e.nama`, p.tahun)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

func apiEkskulSave(c *Ctx) error {
	var in struct {
		Nama      string `json:"nama"`
		PembinaID *int64 `json:"pembina_id"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	nama := cutRunes(in.Nama, 80)
	if nama == "" {
		return bad("Nama ekstrakurikuler wajib diisi.")
	}
	var pembina any
	if in.PembinaID != nil && *in.PembinaID > 0 {
		var n int
		c.A.db.QueryRow(`SELECT COUNT(*) FROM users WHERE id=? AND role IN ('guru','admin') AND status='aktif'`, *in.PembinaID).Scan(&n)
		if n == 0 {
			return bad("Pembina harus akun guru yang aktif.")
		}
		pembina = *in.PembinaID
	}
	var err error
	if id := c.ID("id"); id > 0 {
		_, err = c.A.db.Exec(`UPDATE ekskul SET nama=?, pembina_id=? WHERE id=?`, nama, pembina, id)
	} else {
		_, err = c.A.db.Exec(`INSERT INTO ekskul (nama, pembina_id, created_at) VALUES (?,?,?)`, nama, pembina, nowStr())
	}
	if err != nil {
		return err
	}
	return c.OK()
}

func apiEkskulDelete(c *Ctx) error {
	var n int
	c.A.db.QueryRow(`SELECT COUNT(*) FROM rapor_ekskul WHERE ekskul_id=? AND predikat IS NOT NULL`, c.ID("id")).Scan(&n)
	if n > 0 {
		return apiErr{409, fmt.Sprintf("Ekstrakurikuler ini sudah punya %d nilai rapor. Agar rapor lama tetap utuh, ekstrakurikuler tidak dihapus.", n)}
	}
	return apiDelete("ekskul")(c)
}

func apiEkskulAnggota(c *Ctx) error {
	id := c.ID("id")
	if !c.pembinaEkskul(id) {
		return forbidden()
	}
	p, err := c.periodeQuery()
	if err != nil {
		return err
	}
	rows, err := c.list(`SELECT u.id, u.nama_lengkap, u.username, k.nama AS kelas, re.predikat, re.keterangan FROM ekskul_anggota a
		JOIN users u ON u.id=a.siswa_id LEFT JOIN classes k ON k.id=u.kelas_id
		LEFT JOIN rapor_ekskul re ON re.ekskul_id=a.ekskul_id AND re.siswa_id=a.siswa_id AND re.tahun_id=? AND re.semester=?
		WHERE a.ekskul_id=? AND a.tahun_id=? ORDER BY k.nama, u.nama_lengkap`, p.tahun, p.smt, id, p.tahun)
	if err != nil {
		return err
	}
	return c.JSON(map[string]any{"anggota": rows, "tahun": p.tahun, "smt": p.smt, "periode": p.nama + " · " + p.label()})
}

func apiEkskulAnggotaSave(c *Ctx) error {
	id := c.ID("id")
	if !c.pembinaEkskul(id) {
		return forbidden()
	}
	var in struct {
		Tahun int64   `json:"tahun"`
		Siswa []int64 `json:"siswa_ids"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	p, err := c.periodeDari(in.Tahun, 1)
	if err != nil {
		return err
	}
	tx, err := c.A.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM ekskul_anggota WHERE ekskul_id=? AND tahun_id=?`, id, p.tahun); err != nil {
		return err
	}
	for _, s := range in.Siswa {
		if _, err := tx.Exec(`INSERT INTO ekskul_anggota (ekskul_id, tahun_id, siswa_id) SELECT ?, ?, id FROM users WHERE id=? AND role='siswa'
			ON CONFLICT DO NOTHING`, id, p.tahun, s); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.OK()
}

func apiEkskulNilaiSave(c *Ctx) error {
	id := c.ID("id")
	if !c.pembinaEkskul(id) {
		return forbidden()
	}
	var in struct {
		Tahun int64 `json:"tahun"`
		Smt   int   `json:"smt"`
		Items []struct {
			SiswaID    int64  `json:"siswa_id"`
			Predikat   string `json:"predikat"`
			Keterangan string `json:"keterangan"`
		} `json:"items"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	p, err := c.periodeDari(in.Tahun, in.Smt)
	if err != nil {
		return err
	}
	tx, err := c.A.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, it := range in.Items {
		var pred any
		switch it.Predikat {
		case "":
		case "Sangat Baik", "Baik", "Cukup", "Kurang":
			pred = it.Predikat
		default:
			return bad("Predikat ekstrakurikuler: Sangat Baik, Baik, Cukup, atau Kurang.")
		}
		var anggota int
		tx.QueryRow(`SELECT COUNT(*) FROM ekskul_anggota WHERE ekskul_id=? AND tahun_id=? AND siswa_id=?`, id, p.tahun, it.SiswaID).Scan(&anggota)
		if anggota == 0 {
			continue
		}
		var kunci int
		tx.QueryRow(`SELECT COUNT(*) FROM rapor_terbit t JOIN users u ON u.id=? WHERE t.tahun_id=? AND t.semester=? AND `+memberOf("t.kelas_id"), it.SiswaID, p.tahun, p.smt).Scan(&kunci)
		if kunci > 0 {
			return terkunci()
		}
		if _, err := tx.Exec(`INSERT INTO rapor_ekskul (tahun_id, semester, ekskul_id, siswa_id, predikat, keterangan, updated_at) VALUES (?,?,?,?,?,?,?)
			ON CONFLICT (tahun_id, semester, ekskul_id, siswa_id) DO UPDATE SET predikat=excluded.predikat, keterangan=excluded.keterangan, updated_at=excluded.updated_at`,
			p.tahun, p.smt, id, it.SiswaID, pred, cutRunes(it.Keterangan, 300), nowStr()); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.OK()
}

// ---------------------------------------------------------------------
// Portofolio (internal: siswa pemilik, guru mapel terkait, wali kelas, staf)
// ---------------------------------------------------------------------

func (c *Ctx) bolehLihatPorto(siswa int64, course sql.NullInt64) bool {
	switch {
	case c.U.Role == "siswa":
		return siswa == c.U.ID
	case c.staffRapor():
		return true
	case course.Valid && c.canManage(course.Int64):
		return true
	}
	var kelas sql.NullInt64
	c.A.reader().QueryRow(`SELECT kelas_id FROM users WHERE id=?`, siswa).Scan(&kelas)
	return kelas.Valid && c.isWali(kelas.Int64)
}

func apiPortoList(c *Ctx) error {
	q := `SELECT p.id, p.siswa_id, u.nama_lengkap AS siswa, k.nama AS kelas, p.course_id, co.nama AS mapel, p.judul, p.deskripsi, p.file_url, p.link,
		p.catatan_guru, p.nilai, g.nama_lengkap AS dinilai_oleh, p.dinilai_at, p.created_at
		FROM portfolios p JOIN users u ON u.id=p.siswa_id LEFT JOIN classes k ON k.id=u.kelas_id LEFT JOIN courses co ON co.id=p.course_id
		LEFT JOIN users g ON g.id=p.dinilai_oleh WHERE `
	var args []any
	switch {
	case c.U.Role == "siswa":
		q += `p.siswa_id=?`
		args = append(args, c.U.ID)
	case c.QInt("course_id") > 0:
		cid := c.QInt("course_id")
		if !c.canManage(cid) && !c.staffRapor() {
			return forbidden()
		}
		q += `p.course_id=?`
		args = append(args, cid)
	case c.QInt("kelas_id") > 0:
		kid := c.QInt("kelas_id")
		if !c.staffRapor() && !c.isWali(kid) {
			return forbidden()
		}
		q += `u.kelas_id=?`
		args = append(args, kid)
	case c.staffRapor():
		q += `1=1`
	default: // guru: karya di mapel yang diampu + siswa kelas perwalian
		q += `(co.guru_id=? OR k.wali_id=?)`
		args = append(args, c.U.ID, c.U.ID)
	}
	q += ` ORDER BY p.created_at DESC LIMIT 500`
	rows, err := c.list(q, args...)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

func (a *App) apiPortoSave(c *Ctx) error {
	c.longBody()
	var in struct {
		Judul     string `json:"judul"`
		Deskripsi string `json:"deskripsi"`
		CourseID  *int64 `json:"course_id"`
		Link      string `json:"link"`
		File      string `json:"file"`       // data URL berkas baru (gambar/PDF, maks 5 MB)
		HapusFile bool   `json:"hapus_file"` // saat mengubah: buang berkas lama
	}
	if err := c.bindLarge(&in, 8<<20); err != nil {
		return err
	}
	judul := cutRunes(in.Judul, 150)
	if judul == "" {
		return bad("Judul karya wajib diisi.")
	}
	var course any
	if in.CourseID != nil && *in.CourseID > 0 {
		if !c.canView(*in.CourseID) {
			return bad("Pilih mata pelajaran di kelas Anda.")
		}
		course = *in.CourseID
	}
	link, err := cleanURL(&in.Link)
	if err != nil {
		return err
	}
	id := c.ID("id")
	var lama sql.NullString
	if id > 0 {
		var nilai sql.NullFloat64
		if c.A.db.QueryRow(`SELECT file_url, nilai FROM portfolios WHERE id=? AND siswa_id=?`, id, c.U.ID).Scan(&lama, &nilai) != nil {
			return notFound()
		}
		if nilai.Valid {
			return bad("Karya ini sudah dinilai guru dan tidak bisa diubah.")
		}
	} else {
		var n int
		c.A.db.QueryRow(`SELECT COUNT(*) FROM portfolios WHERE siswa_id=? AND created_at >= ?`, c.U.ID, time.Now().Add(-24*time.Hour).UTC().Format(time.RFC3339)).Scan(&n)
		if n >= 20 {
			return bad("Batas 20 karya per hari tercapai. Coba lagi besok.")
		}
	}
	file := lama.String
	if in.HapusFile {
		file = ""
	}
	if in.File != "" {
		data, err := decodeDataURL(in.File)
		if err != nil || len(data) > 5<<20 {
			return bad("Berkas tidak valid atau lebih dari 5 MB.")
		}
		if file, err = a.saveUpload(data, false); err != nil {
			return err
		}
	}
	if file == "" && link == nil && strings.TrimSpace(in.Deskripsi) == "" {
		return bad("Lampirkan berkas, tautan, atau tulis deskripsi karya.")
	}
	var fv any
	if file != "" {
		fv = file
	}
	if id > 0 {
		_, err = c.A.db.Exec(`UPDATE portfolios SET judul=?, deskripsi=?, course_id=?, link=?, file_url=? WHERE id=? AND siswa_id=?`,
			judul, cutRunes(in.Deskripsi, 2000), course, link, fv, id, c.U.ID)
	} else {
		_, err = c.A.db.Exec(`INSERT INTO portfolios (siswa_id, course_id, judul, deskripsi, file_url, link, created_at) VALUES (?,?,?,?,?,?,?)`,
			c.U.ID, course, judul, cutRunes(in.Deskripsi, 2000), fv, link, nowStr())
	}
	if err != nil {
		return err
	}
	return c.OK()
}

func apiPortoDelete(c *Ctx) error {
	id := c.ID("id")
	var siswa int64
	var nilai sql.NullFloat64
	if c.A.db.QueryRow(`SELECT siswa_id, nilai FROM portfolios WHERE id=?`, id).Scan(&siswa, &nilai) != nil {
		return notFound()
	}
	switch {
	case c.staffRapor():
	case c.U.Role == "siswa" && siswa == c.U.ID:
		if nilai.Valid {
			return bad("Karya ini sudah dinilai guru dan tidak bisa dihapus.")
		}
	default:
		return forbidden()
	}
	return apiDelete("portfolios")(c)
}

func apiPortoCatatan(c *Ctx) error {
	id := c.ID("id")
	var in struct {
		Catatan string   `json:"catatan"`
		Nilai   *float64 `json:"nilai"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	var siswa int64
	var course sql.NullInt64
	if c.A.db.QueryRow(`SELECT siswa_id, course_id FROM portfolios WHERE id=?`, id).Scan(&siswa, &course) != nil {
		return notFound()
	}
	if !c.bolehLihatPorto(siswa, course) {
		return forbidden()
	}
	if in.Nilai != nil && (*in.Nilai < 0 || *in.Nilai > 100) {
		return bad("Nilai karya harus 0 sampai 100.")
	}
	var nilai any
	if in.Nilai != nil {
		nilai = *in.Nilai
	}
	_, err := c.A.db.Exec(`UPDATE portfolios SET catatan_guru=?, nilai=?, dinilai_oleh=?, dinilai_at=? WHERE id=?`, cutRunes(in.Catatan, 1000), nilai, c.U.ID, nowStr(), id)
	if err != nil {
		return err
	}
	return c.OK()
}

// apiEkskulCalon: siswa aktif di kelas tahun ajaran terpilih (untuk memilih anggota ekstrakurikuler).
func apiEkskulCalon(c *Ctx) error {
	if !c.pembinaEkskul(c.ID("id")) {
		return forbidden()
	}
	p, err := c.periodeQuery()
	if err != nil {
		return err
	}
	rows, err := c.list(`SELECT u.id, u.nama_lengkap, k.nama AS kelas FROM users u JOIN classes k ON k.id=u.kelas_id
		WHERE u.role='siswa' AND u.status='aktif' AND k.tahun_id=? ORDER BY k.nama, u.nama_lengkap`, p.tahun)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

// apiRaporGuru: guru aktif (pilihan pembina ekstrakurikuler) untuk admin & staf e-Rapor.
func apiRaporGuru(c *Ctx) error {
	rows, err := c.list(`SELECT id, nama_lengkap FROM users WHERE role='guru' AND status='aktif' ORDER BY nama_lengkap`)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

// jumlahKuis: banyaknya kuis berbeda (bukan banyaknya pengerjaan) dalam data semester.
func jumlahKuis(att []map[string]any) int {
	m := map[any]bool{}
	for _, t := range att {
		m[t["quiz_id"]] = true
	}
	return len(m)
}
