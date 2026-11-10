package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Modul: CBT (ujian berbasis komputer), Manajemen Ujian (ruang, peserta, jadwal), Pengumuman Kelulusan.

func init() {
	moduleSchemas = append(moduleSchemas, `
CREATE TABLE IF NOT EXISTS exams (
  id           INTEGER PRIMARY KEY,
  judul        TEXT NOT NULL,
  mapel        TEXT,
  kelas_ids    TEXT NOT NULL DEFAULT '[]',
  mulai        TEXT NOT NULL,
  selesai      TEXT NOT NULL,
  durasi_menit INTEGER NOT NULL DEFAULT 60,
  jumlah_opsi  INTEGER NOT NULL DEFAULT 4 CHECK (jumlah_opsi BETWEEN 2 AND 5),
  acak_soal    INTEGER NOT NULL DEFAULT 1,
  acak_opsi    INTEGER NOT NULL DEFAULT 1,
  token        TEXT,
  tampil_nilai INTEGER NOT NULL DEFAULT 1,
  created_by   INTEGER REFERENCES users(id) ON DELETE SET NULL,
  created_at   TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS exam_questions (
  id            INTEGER PRIMARY KEY,
  exam_id       INTEGER NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
  pertanyaan    TEXT NOT NULL,
  gambar        TEXT,
  opsi_a TEXT NOT NULL, opsi_b TEXT NOT NULL, opsi_c TEXT, opsi_d TEXT, opsi_e TEXT,
  jawaban_benar TEXT NOT NULL CHECK (jawaban_benar IN ('a','b','c','d','e')),
  bobot         REAL NOT NULL DEFAULT 1,
  created_at    TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_examq_exam ON exam_questions(exam_id);
CREATE TABLE IF NOT EXISTS exam_attempts (
  id           INTEGER PRIMARY KEY,
  exam_id      INTEGER NOT NULL REFERENCES exams(id) ON DELETE CASCADE,
  siswa_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  started_at   TEXT NOT NULL,
  submitted_at TEXT,
  selesai      INTEGER NOT NULL DEFAULT 0,
  skor         REAL,
  jawaban      TEXT NOT NULL DEFAULT '{}',
  urutan       TEXT NOT NULL DEFAULT '[]',
  opsi_map     TEXT NOT NULL DEFAULT '{}',
  UNIQUE (exam_id, siswa_id)
);

CREATE TABLE IF NOT EXISTS exam_events (
  id           INTEGER PRIMARY KEY,
  nama         TEXT NOT NULL,
  tahun_ajaran TEXT,
  semester     TEXT,
  mulai        TEXT,
  selesai      TEXT,
  created_at   TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS exam_rooms (
  id        INTEGER PRIMARY KEY,
  event_id  INTEGER NOT NULL REFERENCES exam_events(id) ON DELETE CASCADE,
  nama      TEXT NOT NULL,
  kapasitas INTEGER NOT NULL DEFAULT 20,
  kolom     INTEGER NOT NULL DEFAULT 4,
  pengawas  TEXT
);
CREATE TABLE IF NOT EXISTS exam_participants (
  event_id   INTEGER NOT NULL REFERENCES exam_events(id) ON DELETE CASCADE,
  siswa_id   INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  room_id    INTEGER REFERENCES exam_rooms(id) ON DELETE SET NULL,
  no_peserta TEXT,
  kursi      INTEGER,
  PRIMARY KEY (event_id, siswa_id)
);
CREATE TABLE IF NOT EXISTS exam_schedule (
  id          INTEGER PRIMARY KEY,
  event_id    INTEGER NOT NULL REFERENCES exam_events(id) ON DELETE CASCADE,
  tanggal     TEXT NOT NULL,
  jam_mulai   TEXT NOT NULL,
  jam_selesai TEXT NOT NULL,
  mapel       TEXT NOT NULL,
  kelas       TEXT
);

CREATE TABLE IF NOT EXISTS graduation (
  siswa_id   INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  status     TEXT NOT NULL CHECK (status IN ('lulus','tidak_lulus','ditunda')),
  catatan    TEXT,
  updated_at TEXT NOT NULL
);
`)
	moduleRoutes = append(moduleRoutes, func(a *App, m *http.ServeMux) {
		u := func(roles string, fn func(*Ctx) error) http.HandlerFunc { return a.mod("ujian", roles, fn) }
		m.HandleFunc("GET /api/cbt", u("any", apiCBTList))
		m.HandleFunc("POST /api/cbt", u("admin,guru", apiCBTSave))
		m.HandleFunc("PUT /api/cbt/{id}", u("admin,guru", apiCBTSave))
		m.HandleFunc("DELETE /api/cbt/{id}", u("admin,guru", apiCBTDelete))
		m.HandleFunc("GET /api/cbt/{id}/soal", u("admin,guru", apiCBTSoal))
		m.HandleFunc("POST /api/cbt/{id}/soal", u("admin,guru", apiCBTSoalSave))
		m.HandleFunc("PUT /api/cbt-soal/{id}", u("admin,guru", apiCBTSoalSave))
		m.HandleFunc("DELETE /api/cbt-soal/{id}", u("admin,guru", apiCBTSoalDelete))
		m.HandleFunc("GET /api/cbt/{id}/hasil", u("admin,guru", apiCBTHasil))
		m.HandleFunc("DELETE /api/cbt-attempt/{id}", u("admin,guru", apiCBTAttemptDelete))
		m.HandleFunc("POST /api/cbt/{id}/start", u("siswa", apiCBTStart))
		m.HandleFunc("POST /api/cbt/{id}/save", u("siswa", apiCBTSaveAnswers))
		m.HandleFunc("POST /api/cbt/{id}/submit", u("siswa", apiCBTSubmit))

		m.HandleFunc("GET /api/ujian-event", u("admin", apiEvents))
		m.HandleFunc("POST /api/ujian-event", u("admin", apiEventSave))
		m.HandleFunc("PUT /api/ujian-event/{id}", u("admin", apiEventSave))
		m.HandleFunc("DELETE /api/ujian-event/{id}", u("admin", apiDelete("exam_events")))
		m.HandleFunc("GET /api/ujian-event/{id}", u("admin", a.apiEventDetail))
		m.HandleFunc("POST /api/ujian-event/{id}/ruang", u("admin", apiRoomSave))
		m.HandleFunc("PUT /api/ujian-ruang/{id}", u("admin", apiRoomSave))
		m.HandleFunc("DELETE /api/ujian-ruang/{id}", u("admin", apiDelete("exam_rooms")))
		m.HandleFunc("POST /api/ujian-event/{id}/jadwal", u("admin", apiScheduleSave))
		m.HandleFunc("DELETE /api/ujian-jadwal/{id}", u("admin", apiDelete("exam_schedule")))
		m.HandleFunc("POST /api/ujian-event/{id}/bagi", u("admin", apiEventAssign))

		k := func(roles string, fn func(*Ctx) error) http.HandlerFunc { return a.mod("kelulusan", roles, fn) }
		m.HandleFunc("GET /api/kelulusan", k("admin", apiGradList))
		m.HandleFunc("PUT /api/kelulusan", k("admin", apiGradSave))
		m.HandleFunc("GET /api/kelulusan/info", k("", apiGradInfo))
		m.HandleFunc("POST /api/kelulusan/cek", k("", apiGradCheck))
	})
}

// ---------------------------------------------------------------------
// CBT
// ---------------------------------------------------------------------
func (c *Ctx) ownsExam(id int64) bool {
	var by sql.NullInt64
	if c.A.reader().QueryRow(`SELECT created_by FROM exams WHERE id=?`, id).Scan(&by) != nil {
		return false
	}
	return c.U.Role == "admin" || (by.Valid && by.Int64 == c.U.ID)
}

func kelasIn(list string, kelas int64) bool {
	var ids []int64
	json.Unmarshal([]byte(list), &ids)
	for _, k := range ids {
		if k == kelas {
			return true
		}
	}
	return false
}

func apiCBTList(c *Ctx) error {
	base := `SELECT e.id, e.judul, e.mapel, e.kelas_ids, e.mulai, e.selesai, e.durasi_menit, e.jumlah_opsi, e.acak_soal, e.acak_opsi,
		e.tampil_nilai, e.created_by, g.nama_lengkap AS pembuat,
		(SELECT COUNT(*) FROM exam_questions q WHERE q.exam_id=e.id) AS jumlah_soal`
	switch c.U.Role {
	case "siswa":
		rows, err := c.list(base+`, t.selesai AS sudah, t.skor, t.started_at FROM exams e LEFT JOIN users g ON g.id=e.created_by
			LEFT JOIN exam_attempts t ON t.exam_id=e.id AND t.siswa_id=?
			WHERE EXISTS (SELECT 1 FROM json_each(e.kelas_ids) WHERE value=?) ORDER BY e.mulai DESC LIMIT 200`, c.U.ID, c.U.KelasID.Int64)
		if err != nil {
			return err
		}
		out := []map[string]any{}
		for _, r := range rows {
			if c.U.KelasID.Valid && kelasIn(fmt.Sprint(r["kelas_ids"]), c.U.KelasID.Int64) {
				if r["tampil_nilai"] != int64(1) {
					r["skor"] = nil
				}
				out = append(out, r)
			}
		}
		return c.JSON(out)
	case "guru":
		rows, err := c.list(base+`, (SELECT COUNT(*) FROM exam_attempts t WHERE t.exam_id=e.id AND t.selesai=1) AS jumlah_selesai, e.token
			FROM exams e LEFT JOIN users g ON g.id=e.created_by WHERE e.created_by=? ORDER BY e.mulai DESC`, c.U.ID)
		if err != nil {
			return err
		}
		return c.JSON(rows)
	}
	rows, err := c.list(base + `, (SELECT COUNT(*) FROM exam_attempts t WHERE t.exam_id=e.id AND t.selesai=1) AS jumlah_selesai, e.token
		FROM exams e LEFT JOIN users g ON g.id=e.created_by ORDER BY e.mulai DESC`)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

func apiCBTSave(c *Ctx) error {
	var in struct {
		Judul       string  `json:"judul"`
		Mapel       *string `json:"mapel"`
		KelasIDs    []int64 `json:"kelas_ids"`
		Mulai       *string `json:"mulai"`
		Selesai     *string `json:"selesai"`
		DurasiMenit int     `json:"durasi_menit"`
		JumlahOpsi  int     `json:"jumlah_opsi"`
		AcakSoal    bool    `json:"acak_soal"`
		AcakOpsi    bool    `json:"acak_opsi"`
		Token       *string `json:"token"`
		TampilNilai bool    `json:"tampil_nilai"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	id := c.ID("id")
	if id > 0 && !c.ownsExam(id) {
		return forbidden()
	}
	judul, err := required(in.Judul, "Nama ujian")
	if err != nil {
		return err
	}
	if len(in.KelasIDs) == 0 {
		return bad("Pilih minimal satu kelas peserta.")
	}
	mulai, err := cleanDeadline(in.Mulai)
	if err != nil || mulai == nil {
		return bad("Waktu mulai wajib diisi.")
	}
	selesai, err := cleanDeadline(in.Selesai)
	if err != nil || selesai == nil {
		return bad("Waktu selesai wajib diisi.")
	}
	if selesai.(string) <= mulai.(string) {
		return bad("Waktu selesai harus setelah waktu mulai.")
	}
	if in.DurasiMenit < 1 || in.DurasiMenit > 600 {
		return bad("Durasi harus 1 sampai 600 menit.")
	}
	if in.JumlahOpsi < 2 || in.JumlahOpsi > 5 {
		in.JumlahOpsi = 4
	}
	kb, _ := json.Marshal(in.KelasIDs)
	tok := str(in.Token)
	if tok != nil {
		tok = strings.ToUpper(tok.(string))
	}
	if id > 0 {
		if err := opsiCukup(c, "exam_questions", "exam_id", id, in.JumlahOpsi); err != nil {
			return err
		}
		_, err = c.A.db.Exec(`UPDATE exams SET judul=?, mapel=?, kelas_ids=?, mulai=?, selesai=?, durasi_menit=?, jumlah_opsi=?, acak_soal=?, acak_opsi=?, token=?, tampil_nilai=? WHERE id=?`,
			judul, str(in.Mapel), string(kb), mulai, selesai, in.DurasiMenit, in.JumlahOpsi, b2i(in.AcakSoal), b2i(in.AcakOpsi), tok, b2i(in.TampilNilai), id)
	} else {
		_, err = c.A.db.Exec(`INSERT INTO exams (judul, mapel, kelas_ids, mulai, selesai, durasi_menit, jumlah_opsi, acak_soal, acak_opsi, token, tampil_nilai, created_by, created_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`, judul, str(in.Mapel), string(kb), mulai, selesai, in.DurasiMenit, in.JumlahOpsi, b2i(in.AcakSoal), b2i(in.AcakOpsi), tok, b2i(in.TampilNilai), c.U.ID, nowStr())
	}
	if err != nil {
		return err
	}
	return c.OK()
}

func apiCBTDelete(c *Ctx) error {
	if !c.ownsExam(c.ID("id")) {
		return forbidden()
	}
	return apiDelete("exams")(c)
}

func apiCBTSoal(c *Ctx) error {
	id := c.ID("id")
	if !c.ownsExam(id) {
		return forbidden()
	}
	rows, err := c.list(`SELECT * FROM exam_questions WHERE exam_id=? ORDER BY id`, id)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

func apiCBTSoalSave(c *Ctx) error {
	in := map[string]*string{}
	var raw map[string]json.RawMessage
	if err := c.Bind(&raw); err != nil {
		return err
	}
	for k, v := range raw {
		var s *string
		if json.Unmarshal(v, &s) == nil {
			in[k] = s
		}
	}
	var bobot float64 = 1
	json.Unmarshal(raw["bobot"], &bobot)
	isUpdate := c.R.Method == http.MethodPut
	id := c.ID("id")
	examID := id
	if isUpdate {
		c.A.db.QueryRow(`SELECT exam_id FROM exam_questions WHERE id=?`, id).Scan(&examID)
	}
	if !c.ownsExam(examID) {
		return forbidden()
	}
	var jopsi int
	c.A.db.QueryRow(`SELECT jumlah_opsi FROM exams WHERE id=?`, examID).Scan(&jopsi)
	get := func(k string) string {
		if in[k] == nil {
			return ""
		}
		return strings.TrimSpace(*in[k])
	}
	p := get("pertanyaan")
	if p == "" {
		return bad("Pertanyaan wajib diisi.")
	}
	opts := make([]any, 5)
	for i := 0; i < 5; i++ {
		v := get("opsi_" + string(rune('a'+i)))
		if i < jopsi {
			if v == "" {
				return bad("Pilihan " + string(rune('A'+i)) + " wajib diisi.")
			}
			opts[i] = v
		}
	}
	kunci := strings.ToLower(get("jawaban_benar"))
	if idx := strings.Index("abcde", kunci); len(kunci) != 1 || idx < 0 || idx >= jopsi {
		return bad("Pilih kunci jawaban yang valid.")
	}
	var gambar any
	if g := get("gambar"); g != "" {
		if !fileNameRe.MatchString(strings.TrimPrefix(g, "/files/")) {
			return bad("Gambar tidak valid.")
		}
		gambar = g
	}
	if bobot <= 0 {
		bobot = 1
	}
	var err error
	if isUpdate {
		var res sql.Result
		res, err = c.A.db.Exec(`UPDATE exam_questions SET pertanyaan=?, gambar=?, opsi_a=?, opsi_b=?, opsi_c=?, opsi_d=?, opsi_e=?, jawaban_benar=?, bobot=? WHERE id=?`,
			p, gambar, opts[0], opts[1], opts[2], opts[3], opts[4], kunci, bobot, id)
		if err == nil {
			if n, _ := res.RowsAffected(); n == 0 {
				return notFound() // soal sudah dihapus (mis. di tab lain): jangan laporkan "tersimpan"
			}
		}
	} else {
		_, err = c.A.db.Exec(`INSERT INTO exam_questions (exam_id, pertanyaan, gambar, opsi_a, opsi_b, opsi_c, opsi_d, opsi_e, jawaban_benar, bobot, created_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?)`, examID, p, gambar, opts[0], opts[1], opts[2], opts[3], opts[4], kunci, bobot, nowStr())
	}
	if err != nil {
		return err
	}
	return c.OK()
}

func apiCBTSoalDelete(c *Ctx) error {
	var examID int64
	c.A.db.QueryRow(`SELECT exam_id FROM exam_questions WHERE id=?`, c.ID("id")).Scan(&examID)
	if !c.ownsExam(examID) {
		return forbidden()
	}
	return apiDelete("exam_questions")(c)
}

func apiCBTHasil(c *Ctx) error {
	id := c.ID("id")
	if !c.ownsExam(id) {
		return forbidden()
	}
	c.A.finalizeCBT(id) // nilai langsung terlihat untuk siswa yang tidak sempat mengumpulkan
	var kelas string
	c.A.db.QueryRow(`SELECT kelas_ids FROM exams WHERE id=?`, id).Scan(&kelas)
	var ids []int64
	json.Unmarshal([]byte(kelas), &ids)
	ph := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
	args := []any{id}
	for _, k := range ids {
		args = append(args, k)
	}
	if ph == "" {
		ph = "NULL"
	}
	// peserta = siswa kelas tersebut, termasuk yang sudah naik kelas (riwayat), atau siapa pun yang punya attempt
	args = append(args, args[1:]...)
	args = append(args, id)
	rows, err := c.list(`SELECT u.id AS siswa_id, u.username, u.nama_lengkap, k.nama AS kelas_nama, t.id AS attempt_id, t.selesai, t.skor, t.started_at, t.submitted_at
		FROM users u LEFT JOIN classes k ON k.id=u.kelas_id LEFT JOIN exam_attempts t ON t.siswa_id=u.id AND t.exam_id=?
		WHERE u.role='siswa' AND ((u.status='aktif' AND u.kelas_id IN (`+ph+`))
		   OR EXISTS (SELECT 1 FROM class_members m WHERE m.siswa_id=u.id AND m.kelas_id IN (`+ph+`))
		   OR EXISTS (SELECT 1 FROM exam_attempts x WHERE x.siswa_id=u.id AND x.exam_id=?))
		ORDER BY k.nama, u.nama_lengkap`, args...)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

func apiCBTAttemptDelete(c *Ctx) error {
	var examID int64
	c.A.db.QueryRow(`SELECT exam_id FROM exam_attempts WHERE id=?`, c.ID("id")).Scan(&examID)
	if !c.ownsExam(examID) {
		return forbidden()
	}
	return apiDelete("exam_attempts")(c)
}

func shuffleInts(a []int64) {
	for i := len(a) - 1; i > 0; i-- {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		a[i], a[j.Int64()] = a[j.Int64()], a[i]
	}
}
func shuffleStr(s string) string {
	r := []rune(s)
	for i := len(r) - 1; i > 0; i-- {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		r[i], r[j.Int64()] = r[j.Int64()], r[i]
	}
	return string(r)
}

type examRow struct {
	judul, kelasIDs, mulai, selesai string
	durasi, jopsi                   int
	acakSoal, acakOpsi, tampil      bool
	token                           sql.NullString
}

func (c *Ctx) loadExam(id int64) (*examRow, error) {
	e := &examRow{}
	var as, ao, tn int
	err := c.A.db.QueryRow(`SELECT judul, kelas_ids, mulai, selesai, durasi_menit, jumlah_opsi, acak_soal, acak_opsi, tampil_nilai, token FROM exams WHERE id=?`, id).
		Scan(&e.judul, &e.kelasIDs, &e.mulai, &e.selesai, &e.durasi, &e.jopsi, &as, &ao, &tn, &e.token)
	if err != nil {
		return nil, notFound()
	}
	e.acakSoal, e.acakOpsi, e.tampil = as == 1, ao == 1, tn == 1
	if !c.U.KelasID.Valid || !kelasIn(e.kelasIDs, c.U.KelasID.Int64) {
		return nil, notFound()
	}
	return e, nil
}

func apiCBTStart(c *Ctx) error {
	id := c.ID("id")
	var in struct {
		Token string `json:"token"`
	}
	c.Bind(&in)
	e, err := c.loadExam(id)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	mulai, _ := time.Parse(time.RFC3339, e.mulai)
	selesai, _ := time.Parse(time.RFC3339, e.selesai)
	var started, urutan, omap string
	var done int
	err = c.A.db.QueryRow(`SELECT started_at, selesai, urutan, opsi_map FROM exam_attempts WHERE exam_id=? AND siswa_id=?`, id, c.U.ID).Scan(&started, &done, &urutan, &omap)
	if err == nil && done == 1 {
		return bad("Ujian ini sudah Anda kerjakan.")
	}
	if err == sql.ErrNoRows {
		if now.Before(mulai) {
			return bad("Ujian belum dimulai. Dibuka " + mulai.Local().Format("02/01/2006 15:04") + ".")
		}
		if now.After(selesai) {
			return bad("Waktu ujian sudah ditutup.")
		}
		if e.token.Valid && e.token.String != "" && !strings.EqualFold(strings.TrimSpace(in.Token), e.token.String) {
			key := fmt.Sprintf("tok:%d:%d", id, c.U.ID)
			if c.A.limit.blockedN(key, 10) {
				return bad("Terlalu banyak token salah. Minta pengawas memeriksa, coba lagi 10 menit lagi.")
			}
			c.A.limit.fail(key)
			return apiErr{400, "TOKEN"}
		}
		var qids []int64
		rows, err := c.A.db.Query(`SELECT id FROM exam_questions WHERE exam_id=? ORDER BY id`, id)
		if err != nil {
			return err
		}
		for rows.Next() {
			var q int64
			rows.Scan(&q)
			qids = append(qids, q)
		}
		rows.Close()
		if len(qids) == 0 {
			return bad("Ujian ini belum memiliki soal.")
		}
		if e.acakSoal {
			shuffleInts(qids)
		}
		om := map[string]string{}
		letters := "abcde"[:e.jopsi]
		for _, q := range qids {
			if e.acakOpsi {
				om[strconv.FormatInt(q, 10)] = shuffleStr(letters)
			} else {
				om[strconv.FormatInt(q, 10)] = letters
			}
		}
		ub, _ := json.Marshal(qids)
		ob, _ := json.Marshal(om)
		started, urutan, omap = nowStr(), string(ub), string(ob)
		if _, err := c.A.db.Exec(`INSERT INTO exam_attempts (exam_id, siswa_id, started_at, urutan, opsi_map) VALUES (?,?,?,?,?)
			ON CONFLICT (exam_id, siswa_id) DO NOTHING`, id, c.U.ID, started, urutan, omap); err != nil {
			return err
		}
		// klik ganda: pakai attempt yang sudah tersimpan lebih dulu
		c.A.db.QueryRow(`SELECT started_at, urutan, opsi_map FROM exam_attempts WHERE exam_id=? AND siswa_id=?`, id, c.U.ID).Scan(&started, &urutan, &omap)
	} else if err != nil {
		return err
	}
	// susun soal sesuai urutan & acakan milik siswa ini, TANPA kunci jawaban
	var qids []int64
	json.Unmarshal([]byte(urutan), &qids)
	om := map[string]string{}
	json.Unmarshal([]byte(omap), &om)
	all, err := c.list(`SELECT id, pertanyaan, gambar, opsi_a, opsi_b, opsi_c, opsi_d, opsi_e FROM exam_questions WHERE exam_id=?`, id)
	if err != nil {
		return err
	}
	byID := map[int64]map[string]any{}
	for _, q := range all {
		byID[q["id"].(int64)] = q
	}
	var saved map[string]string
	var jaw string
	c.A.db.QueryRow(`SELECT jawaban FROM exam_attempts WHERE exam_id=? AND siswa_id=?`, id, c.U.ID).Scan(&jaw)
	json.Unmarshal([]byte(jaw), &saved)
	qs := []map[string]any{}
	answers := map[string]string{}
	for _, qid := range qids {
		q, ok := byID[qid]
		if !ok {
			continue
		}
		k := strconv.FormatInt(qid, 10)
		order := om[k]
		opsi := []any{}
		for i, orig := range order {
			opsi = append(opsi, q["opsi_"+string(orig)])
			if saved[k] == string(orig) {
				answers[k] = string(rune('a' + i))
			}
		}
		qs = append(qs, map[string]any{"id": qid, "pertanyaan": q["pertanyaan"], "gambar": q["gambar"], "opsi": opsi})
	}
	st, _ := time.Parse(time.RFC3339, started)
	end := st.Add(time.Duration(e.durasi) * time.Minute)
	if end.After(selesai) {
		end = selesai
	}
	return c.JSON(map[string]any{
		"started_at": started,
		"exam":       map[string]any{"id": id, "judul": e.judul, "durasi_menit": e.durasi},
		"sisa_detik": math.Max(0, math.Floor(time.Until(end).Seconds())),
		"questions":  qs,
		"answers":    answers,
	})
}

// examDeadline: batas waktu attempt = min(mulai + durasi, jadwal ditutup) + toleransi 60 detik.
func examDeadline(started, durasiMenit, selesai string) time.Time {
	st, err := time.Parse(time.RFC3339, started)
	if err != nil {
		return time.Now().Add(time.Hour)
	}
	d, _ := strconv.Atoi(durasiMenit)
	end := st.Add(time.Duration(d) * time.Minute)
	if cl, err := time.Parse(time.RFC3339, selesai); err == nil && cl.Before(end) {
		end = cl
	}
	return end.Add(60 * time.Second)
}

// toOriginal mengubah jawaban (huruf tampilan) menjadi huruf asli sesuai acakan opsi.
func toOriginal(answers map[string]string, omap string) map[string]string {
	om := map[string]string{}
	json.Unmarshal([]byte(omap), &om)
	out := map[string]string{}
	for k, v := range answers {
		order, ok := om[k]
		v = strings.ToLower(strings.TrimSpace(v))
		if !ok || len(v) != 1 {
			continue
		}
		i := int(v[0] - 'a')
		if i >= 0 && i < len(order) {
			out[k] = string(order[i])
		}
	}
	return out
}

func apiCBTSaveAnswers(c *Ctx) error {
	id := c.ID("id")
	var in struct {
		Answers map[string]string `json:"answers"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	var omap, started, durasi, selesai string
	var done int
	if err := c.A.db.QueryRow(`SELECT t.opsi_map, t.selesai, t.started_at, e.durasi_menit, e.selesai FROM exam_attempts t JOIN exams e ON e.id=t.exam_id
		WHERE t.exam_id=? AND t.siswa_id=?`, id, c.U.ID).Scan(&omap, &done, &started, &durasi, &selesai); err != nil || done == 1 {
		return bad("Ujian tidak sedang berlangsung.")
	}
	if time.Now().After(examDeadline(started, durasi, selesai)) {
		return bad("Waktu ujian sudah habis. Jawaban terakhir yang tersimpan akan dinilai.")
	}
	jb, _ := json.Marshal(toOriginal(in.Answers, omap))
	if _, err := c.A.db.Exec(`UPDATE exam_attempts SET jawaban=? WHERE exam_id=? AND siswa_id=? AND selesai=0`, string(jb), id, c.U.ID); err != nil {
		return err
	}
	return c.OK()
}

func apiCBTSubmit(c *Ctx) error {
	id := c.ID("id")
	var in struct {
		Answers map[string]string `json:"answers"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	tx, err := c.A.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var aid int64
	var omap, saved, started, durasi, selesai, urutan string
	if err := tx.QueryRow(`SELECT t.id, t.opsi_map, t.jawaban, t.started_at, e.durasi_menit, e.selesai, t.urutan FROM exam_attempts t JOIN exams e ON e.id=t.exam_id
		WHERE t.exam_id=? AND t.siswa_id=? AND t.selesai=0`, id, c.U.ID).Scan(&aid, &omap, &saved, &started, &durasi, &selesai, &urutan); err != nil {
		// sudah ditutup (dinilai otomatis karena waktu habis saat HP offline, atau terkirim dari tab lain):
		// kirim ulang dari HP tidak boleh gagal terus — tampilkan nilai yang tersimpan
		var skor sql.NullFloat64
		var tampil int
		if tx.QueryRow(`SELECT t.skor, e.tampil_nilai FROM exam_attempts t JOIN exams e ON e.id=t.exam_id
			WHERE t.exam_id=? AND t.siswa_id=? AND t.selesai=1 AND COALESCE(t.metode,'')<>'ljk'`, id, c.U.ID).Scan(&skor, &tampil) == nil {
			if tampil != 1 || !skor.Valid {
				return c.JSON(map[string]any{"skor": nil})
			}
			return c.JSON(map[string]any{"skor": skor.Float64})
		}
		return bad("Ujian belum dimulai atau sudah dikirim.")
	}
	ans := toOriginal(in.Answers, omap)
	if time.Now().After(examDeadline(started, durasi, selesai)) {
		// terlambat: nilai hanya jawaban yang sudah tersimpan sebelum waktu habis
		ans = map[string]string{}
		json.Unmarshal([]byte(saved), &ans)
	}
	skor, err := scoreAttempt(tx, id, urutan, ans)
	if err != nil {
		return err
	}
	jb, _ := json.Marshal(ans)
	if _, err := tx.Exec(`UPDATE exam_attempts SET jawaban=?, skor=?, selesai=1, submitted_at=? WHERE id=?`, string(jb), skor, nowStr(), aid); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	var tampil int
	c.A.db.QueryRow(`SELECT tampil_nilai FROM exams WHERE id=?`, id).Scan(&tampil)
	if tampil != 1 {
		return c.JSON(map[string]any{"skor": nil})
	}
	return c.JSON(map[string]any{"skor": skor})
}

// scoreAttempt menilai jawaban (huruf asli) hanya atas soal yang memang diterima siswa (urutan miliknya):
// soal yang ditambahkan guru setelah siswa mulai tidak menurunkan nilai.
func scoreAttempt(tx *sql.Tx, examID int64, urutan string, ans map[string]string) (float64, error) {
	var got []int64
	json.Unmarshal([]byte(urutan), &got)
	dapat := map[int64]bool{}
	for _, q := range got {
		dapat[q] = true
	}
	rows, err := tx.Query(`SELECT id, jawaban_benar, bobot FROM exam_questions WHERE exam_id=?`, examID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var total, benar float64
	for rows.Next() {
		var qid int64
		var kunci string
		var bobot float64
		if err := rows.Scan(&qid, &kunci, &bobot); err != nil {
			return 0, err
		}
		if len(dapat) > 0 && !dapat[qid] {
			continue
		}
		total += bobot
		if ans[strconv.FormatInt(qid, 10)] == kunci {
			benar += bobot
		}
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	return persen(benar, total), nil
}

// persen: nilai 0–100 dengan dua desimal (dipakai CBT, kuis, dan LJK)
func persen(benar, total float64) float64 {
	if total <= 0 {
		return 0
	}
	return math.Round(benar/total*10000) / 100
}

// finalizeCBT menilai ujian yang waktunya sudah habis tetapi tidak pernah dikumpulkan (HP mati, browser
// tertutup): jawaban yang sudah tersimpan otomatis dinilai, sehingga tidak tertinggal "belum selesai" selamanya.
// examID 0 = semua ujian.
func (a *App) finalizeCBT(examID int64) {
	q := `SELECT t.id, t.exam_id, t.started_at, e.durasi_menit, e.selesai FROM exam_attempts t
		JOIN exams e ON e.id=t.exam_id WHERE t.selesai=0`
	args := []any{}
	if examID > 0 {
		q += ` AND t.exam_id=?`
		args = append(args, examID)
	}
	type att struct {
		id, exam          int64
		start, dur, tutup string
	}
	rows, err := a.reader().Query(q, args...)
	if err != nil {
		return
	}
	var due []att
	for rows.Next() {
		var x att
		if rows.Scan(&x.id, &x.exam, &x.start, &x.dur, &x.tutup) == nil && time.Now().After(examDeadline(x.start, x.dur, x.tutup)) {
			due = append(due, x)
		}
	}
	rows.Close()
	for _, x := range due {
		func() {
			tx, err := a.db.Begin()
			if err != nil {
				return
			}
			defer tx.Rollback()
			var jawaban, urutan string
			if tx.QueryRow(`SELECT jawaban, urutan FROM exam_attempts WHERE id=? AND selesai=0`, x.id).Scan(&jawaban, &urutan) != nil {
				return
			}
			ans := map[string]string{}
			json.Unmarshal([]byte(jawaban), &ans)
			skor, err := scoreAttempt(tx, x.exam, urutan, ans)
			if err != nil {
				return
			}
			if res, err := tx.Exec(`UPDATE exam_attempts SET skor=?, selesai=1, submitted_at=? WHERE id=? AND selesai=0`, skor, nowStr(), x.id); err == nil {
				if n, _ := res.RowsAffected(); n == 1 {
					tx.Commit()
				}
			}
		}()
	}
}
