package main

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// LMS: kuis pilihan ganda (soal, pengerjaan, nilai).

// ---------------------------------------------------------------------
// Kuis
// ---------------------------------------------------------------------
func apiQuizzes(c *Ctx) error {
	cid := c.QInt("course_id")
	if !c.canView(cid) {
		return forbidden()
	}
	rows, err := c.list(`SELECT q.id, q.course_id, q.judul, q.durasi_menit, q.jumlah_opsi, q.created_at,
		(SELECT COUNT(*) FROM quiz_questions x WHERE x.quiz_id=q.id) AS jumlah_soal,
		(SELECT COUNT(*) FROM quiz_attempts t WHERE t.quiz_id=q.id AND t.selesai=1) AS jumlah_dikerjakan
		FROM quizzes q WHERE q.course_id=? ORDER BY q.created_at DESC`, cid)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

func apiSaveQuiz(c *Ctx) error {
	var in struct {
		CourseID    int64  `json:"course_id"`
		Judul       string `json:"judul"`
		DurasiMenit int    `json:"durasi_menit"`
		JumlahOpsi  int    `json:"jumlah_opsi"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	id := c.ID("id")
	if id > 0 {
		in.CourseID = c.courseOf("quizzes", id)
	}
	if err := c.requireManage(in.CourseID, "kuis"); err != nil {
		return err
	}
	judul, err := required(in.Judul, "Judul kuis")
	if err != nil {
		return err
	}
	if in.DurasiMenit < 1 || in.DurasiMenit > 600 {
		return bad("Durasi harus 1 sampai 600 menit.")
	}
	if in.JumlahOpsi < 2 || in.JumlahOpsi > 5 {
		in.JumlahOpsi = 4
	}
	if id > 0 {
		if err := opsiCukup(c, "quiz_questions", "quiz_id", id, in.JumlahOpsi); err != nil {
			return err
		}
		_, err = c.A.db.Exec(`UPDATE quizzes SET judul=?, durasi_menit=?, jumlah_opsi=? WHERE id=?`, judul, in.DurasiMenit, in.JumlahOpsi, id)
	} else {
		_, err = c.A.db.Exec(`INSERT INTO quizzes (course_id, judul, durasi_menit, jumlah_opsi, created_at) VALUES (?,?,?,?,?)`,
			in.CourseID, judul, in.DurasiMenit, in.JumlahOpsi, nowStr())
	}
	if err != nil {
		return err
	}
	return c.OK()
}

func apiQuestions(c *Ctx) error {
	id := c.ID("id")
	if !c.canManage(c.courseOf("quizzes", id)) {
		return forbidden()
	}
	rows, err := c.list(`SELECT id, quiz_id, pertanyaan, opsi_a, opsi_b, opsi_c, opsi_d, opsi_e, jawaban_benar, bobot
		FROM quiz_questions WHERE quiz_id=? ORDER BY id`, id)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

func (c *Ctx) quizOfQuestion(qid int64) int64 {
	var id int64
	c.A.db.QueryRow(`SELECT quiz_id FROM quiz_questions WHERE id=?`, qid).Scan(&id)
	return id
}

func apiSaveQuestion(c *Ctx) error {
	var in struct {
		Pertanyaan          string `json:"pertanyaan"`
		OpsiA, OpsiB        string
		OpsiC, OpsiD, OpsiE *string
		JawabanBenar        string  `json:"jawaban_benar"`
		Bobot               float64 `json:"bobot"`
	}
	var raw map[string]json.RawMessage
	if err := c.Bind(&raw); err != nil {
		return err
	}
	gs := func(k string) *string {
		var s *string
		json.Unmarshal(raw[k], &s)
		return s
	}
	if s := gs("pertanyaan"); s != nil {
		in.Pertanyaan = *s
	}
	if s := gs("opsi_a"); s != nil {
		in.OpsiA = *s
	}
	if s := gs("opsi_b"); s != nil {
		in.OpsiB = *s
	}
	in.OpsiC, in.OpsiD, in.OpsiE = gs("opsi_c"), gs("opsi_d"), gs("opsi_e")
	if s := gs("jawaban_benar"); s != nil {
		in.JawabanBenar = strings.ToLower(strings.TrimSpace(*s))
	}
	json.Unmarshal(raw["bobot"], &in.Bobot)

	qid := c.ID("id") // id soal (PUT /api/questions/{id}) atau id kuis (POST /api/quizzes/{id}/questions)
	isUpdate := c.R.Method == http.MethodPut
	quizID := qid
	if isUpdate {
		quizID = c.quizOfQuestion(qid)
	}
	if err := c.requireManage(c.courseOf("quizzes", quizID), "kuis"); err != nil {
		return err
	}
	var jopsi int
	c.A.db.QueryRow(`SELECT jumlah_opsi FROM quizzes WHERE id=?`, quizID).Scan(&jopsi)
	p, err := required(in.Pertanyaan, "Pertanyaan")
	if err != nil {
		return err
	}
	oa, err := required(in.OpsiA, "Opsi A")
	if err != nil {
		return err
	}
	ob, err := required(in.OpsiB, "Opsi B")
	if err != nil {
		return err
	}
	opts := []any{oa, ob, str(in.OpsiC), str(in.OpsiD), str(in.OpsiE)}
	for i := 2; i < 5; i++ {
		if i >= jopsi {
			opts[i] = nil
		} else if opts[i] == nil {
			return bad("Opsi " + string(rune('A'+i)) + " wajib diisi.")
		}
	}
	idx := strings.Index("abcde", in.JawabanBenar)
	if len(in.JawabanBenar) != 1 || idx < 0 || idx >= jopsi {
		return bad("Pilih kunci jawaban yang valid.")
	}
	if in.Bobot <= 0 {
		in.Bobot = 1
	}
	if isUpdate {
		_, err = c.A.db.Exec(`UPDATE quiz_questions SET pertanyaan=?, opsi_a=?, opsi_b=?, opsi_c=?, opsi_d=?, opsi_e=?, jawaban_benar=?, bobot=? WHERE id=?`,
			p, opts[0], opts[1], opts[2], opts[3], opts[4], in.JawabanBenar, in.Bobot, qid)
	} else {
		_, err = c.A.db.Exec(`INSERT INTO quiz_questions (quiz_id, pertanyaan, opsi_a, opsi_b, opsi_c, opsi_d, opsi_e, jawaban_benar, bobot, created_at)
			VALUES (?,?,?,?,?,?,?,?,?,?)`, quizID, p, opts[0], opts[1], opts[2], opts[3], opts[4], in.JawabanBenar, in.Bobot, nowStr())
	}
	if err != nil {
		return err
	}
	return c.OK()
}

func apiDeleteQuestion(c *Ctx) error {
	id := c.ID("id")
	if err := c.requireManage(c.courseOf("quizzes", c.quizOfQuestion(id)), "kuis"); err != nil {
		return err
	}
	return apiDelete("quiz_questions")(c)
}

func apiAttempts(c *Ctx) error {
	id := c.ID("id")
	if !c.canManage(c.courseOf("quizzes", id)) {
		return forbidden()
	}
	rows, err := c.list(`SELECT t.id, t.siswa_id, t.skor, t.started_at, t.submitted_at, t.selesai, u.nama_lengkap, u.username
		FROM quiz_attempts t JOIN users u ON u.id=t.siswa_id WHERE t.quiz_id=? ORDER BY u.nama_lengkap`, id)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

func apiDeleteAttempt(c *Ctx) error {
	id := c.ID("id")
	var quiz int64
	if c.A.db.QueryRow(`SELECT quiz_id FROM quiz_attempts WHERE id=?`, id).Scan(&quiz) != nil {
		return notFound()
	}
	if !c.canManage(c.courseOf("quizzes", quiz)) {
		return forbidden()
	}
	return apiDelete("quiz_attempts")(c)
}

func apiQuizStart(c *Ctx) error {
	id := c.ID("id")
	if !c.canView(c.courseOf("quizzes", id)) {
		return notFound()
	}
	var judul string
	var durasi, jopsi int
	if err := c.A.db.QueryRow(`SELECT judul, durasi_menit, jumlah_opsi FROM quizzes WHERE id=?`, id).Scan(&judul, &durasi, &jopsi); err != nil {
		return notFound()
	}
	var started string
	var selesai int
	err := c.A.db.QueryRow(`SELECT started_at, selesai FROM quiz_attempts WHERE quiz_id=? AND siswa_id=?`, id, c.U.ID).Scan(&started, &selesai)
	if err == nil && selesai == 1 {
		return bad("Kuis ini sudah Anda kerjakan.")
	}
	var nsoal int
	c.A.db.QueryRow(`SELECT COUNT(*) FROM quiz_questions WHERE quiz_id=?`, id).Scan(&nsoal)
	if nsoal == 0 {
		return bad("Kuis ini belum memiliki soal.")
	}
	if err == sql.ErrNoRows {
		started = nowStr()
		if _, err := c.A.db.Exec(`INSERT INTO quiz_attempts (quiz_id, siswa_id, started_at) VALUES (?,?,?)`, id, c.U.ID, started); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	// Soal dikirim TANPA kunci jawaban
	qs, err := c.list(`SELECT id, pertanyaan, opsi_a, opsi_b, opsi_c, opsi_d, opsi_e FROM quiz_questions WHERE quiz_id=? ORDER BY id`, id)
	if err != nil {
		return err
	}
	st, _ := time.Parse(time.RFC3339, started)
	remaining := time.Duration(durasi)*time.Minute - time.Since(st)
	// jawaban yang sudah tersimpan di server (melanjutkan dari HP lain / setelah HP mati)
	var jaw sql.NullString
	c.A.db.QueryRow(`SELECT jawaban FROM quiz_attempts WHERE quiz_id=? AND siswa_id=?`, id, c.U.ID).Scan(&jaw)
	answers := map[string]string{}
	json.Unmarshal([]byte(jaw.String), &answers)
	return c.JSON(map[string]any{
		"started_at": started,
		"quiz":       map[string]any{"id": id, "judul": judul, "durasi_menit": durasi, "jumlah_opsi": jopsi},
		"sisa_detik": math.Max(0, math.Floor(remaining.Seconds())),
		"questions":  qs,
		"answers":    answers,
	})
}

// batas kuis: mulai + durasi + toleransi 60 detik (jaringan lambat)
func quizDeadline(started string, durasi int) time.Time {
	st, err := time.Parse(time.RFC3339, started)
	if err != nil {
		return time.Now().Add(time.Hour)
	}
	return st.Add(time.Duration(durasi)*time.Minute + 60*time.Second)
}

// quizClean: jawaban siswa → huruf kecil a–e saja, untuk soal yang ada.
func quizClean(in map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range in {
		v = strings.ToLower(strings.TrimSpace(v))
		if len(k) <= 20 && len(v) == 1 && v[0] >= 'a' && v[0] <= 'e' {
			out[k] = v
		}
	}
	return out
}

// scoreQuiz menilai jawaban atas semua soal kuis (kuis mapel tidak mengacak pilihan).
func scoreQuiz(tx *sql.Tx, quizID int64, ans map[string]string) (float64, error) {
	rows, err := tx.Query(`SELECT id, jawaban_benar, bobot FROM quiz_questions WHERE quiz_id=?`, quizID)
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

// apiQuizSave: simpan jawaban sementara selama kuis berjalan (seperti CBT), agar tidak hilang bila HP mati.
func apiQuizSave(c *Ctx) error {
	id := c.ID("id")
	var in struct {
		Answers map[string]string `json:"answers"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	var started string
	var durasi int
	if err := c.A.db.QueryRow(`SELECT t.started_at, q.durasi_menit FROM quiz_attempts t JOIN quizzes q ON q.id=t.quiz_id
		WHERE t.quiz_id=? AND t.siswa_id=? AND t.selesai=0`, id, c.U.ID).Scan(&started, &durasi); err != nil {
		return bad("Kuis belum dimulai atau sudah dikirim.")
	}
	if time.Now().After(quizDeadline(started, durasi)) {
		return bad("Waktu kuis sudah habis.")
	}
	jb, _ := json.Marshal(quizClean(in.Answers))
	if _, err := c.A.db.Exec(`UPDATE quiz_attempts SET jawaban=? WHERE quiz_id=? AND siswa_id=? AND selesai=0`, string(jb), id, c.U.ID); err != nil {
		return err
	}
	return c.OK()
}

func apiQuizSubmit(c *Ctx) error {
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
	var attempt int64
	var started string
	var saved sql.NullString
	var durasi int
	if err := tx.QueryRow(`SELECT t.id, t.started_at, t.jawaban, q.durasi_menit FROM quiz_attempts t JOIN quizzes q ON q.id=t.quiz_id
		WHERE t.quiz_id=? AND t.siswa_id=? AND t.selesai=0`, id, c.U.ID).Scan(&attempt, &started, &saved, &durasi); err != nil {
		// sudah ditutup (terkirim sebelumnya tapi balasan hilang, atau dinilai otomatis): tampilkan nilai tersimpan
		var skor float64
		if tx.QueryRow(`SELECT skor FROM quiz_attempts WHERE quiz_id=? AND siswa_id=? AND selesai=1`, id, c.U.ID).Scan(&skor) == nil {
			return c.JSON(map[string]any{"skor": skor})
		}
		return bad("Kuis belum dimulai atau sudah dikirim.")
	}
	ans := quizClean(in.Answers)
	if time.Now().After(quizDeadline(started, durasi)) {
		// terlambat: yang dinilai hanya jawaban yang sudah tersimpan sebelum waktu habis (sama seperti CBT)
		ans = map[string]string{}
		json.Unmarshal([]byte(saved.String), &ans)
		ans = quizClean(ans)
	}
	skor, err := scoreQuiz(tx, id, ans)
	if err != nil {
		return err
	}
	jb, _ := json.Marshal(ans)
	if _, err := tx.Exec(`UPDATE quiz_attempts SET skor=?, jawaban=?, selesai=1, submitted_at=? WHERE id=?`, skor, string(jb), nowStr(), attempt); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.JSON(map[string]any{"skor": skor})
}

// finalizeQuiz menilai kuis yang waktunya habis tetapi tidak pernah dikirim, dari jawaban yang tersimpan.
func (a *App) finalizeQuiz() {
	rows, err := a.reader().Query(`SELECT t.id, t.quiz_id, t.started_at, q.durasi_menit FROM quiz_attempts t JOIN quizzes q ON q.id=t.quiz_id WHERE t.selesai=0`)
	if err != nil {
		return
	}
	type att struct {
		id, quiz int64
		start    string
		dur      int
	}
	var due []att
	for rows.Next() {
		var x att
		if rows.Scan(&x.id, &x.quiz, &x.start, &x.dur) == nil && time.Now().After(quizDeadline(x.start, x.dur)) {
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
			var saved sql.NullString
			if tx.QueryRow(`SELECT jawaban FROM quiz_attempts WHERE id=? AND selesai=0`, x.id).Scan(&saved) != nil {
				return
			}
			ans := map[string]string{}
			json.Unmarshal([]byte(saved.String), &ans)
			ans = quizClean(ans)
			skor, err := scoreQuiz(tx, x.quiz, ans)
			if err != nil {
				return
			}
			jb, _ := json.Marshal(ans)
			if res, err := tx.Exec(`UPDATE quiz_attempts SET skor=?, jawaban=?, selesai=1, submitted_at=? WHERE id=? AND selesai=0`, skor, string(jb), nowStr(), x.id); err == nil {
				if n, _ := res.RowsAffected(); n == 1 {
					tx.Commit()
				}
			}
		}()
	}
}
