package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// Impor soal pilihan ganda sekaligus (dari teks atau hasil Asisten AI) dalam SATU transaksi:
// semua soal divalidasi dulu; bila satu salah, tidak ada yang tersimpan (mengulang impor tidak menggandakan soal).

func init() {
	moduleRoutes = append(moduleRoutes, func(a *App, m *http.ServeMux) {
		m.HandleFunc("POST /api/cbt/{id}/soal/batch", a.mod("ujian", "admin,guru", apiCBTSoalBatch))
		m.HandleFunc("POST /api/quizzes/{id}/questions/batch", a.h("admin,guru", apiQuizQuestionBatch))
	})
}

type mcq struct {
	pertanyaan string
	gambar     any
	opsi       [5]any
	kunci      string
	bobot      float64
}

// parseMCQ memvalidasi satu soal pilihan ganda sesuai jumlah pilihan (2–5).
func parseMCQ(raw map[string]json.RawMessage, jopsi int, allowImage bool) (mcq, error) {
	get := func(k string) string {
		var s *string
		if json.Unmarshal(raw[k], &s) != nil || s == nil {
			return ""
		}
		return strings.TrimSpace(*s)
	}
	q := mcq{pertanyaan: get("pertanyaan"), bobot: 1}
	if q.pertanyaan == "" {
		return q, bad("Pertanyaan wajib diisi.")
	}
	for i := 0; i < 5; i++ {
		v := get("opsi_" + string(rune('a'+i)))
		if i < jopsi {
			if v == "" {
				return q, bad("Pilihan " + string(rune('A'+i)) + " wajib diisi.")
			}
			q.opsi[i] = v
		}
	}
	q.kunci = strings.ToLower(get("jawaban_benar"))
	if idx := strings.Index("abcde", q.kunci); len(q.kunci) != 1 || idx < 0 || idx >= jopsi {
		return q, bad("Pilih kunci jawaban yang valid.")
	}
	if g := get("gambar"); g != "" && allowImage {
		if !fileNameRe.MatchString(strings.TrimPrefix(g, "/files/")) {
			return q, bad("Gambar tidak valid.")
		}
		q.gambar = g
	}
	var b float64
	if json.Unmarshal(raw["bobot"], &b) == nil && b > 0 {
		q.bobot = b
	}
	return q, nil
}

func parseBatch(c *Ctx, jopsi int, allowImage bool) ([]mcq, error) {
	var in struct {
		Soal []map[string]json.RawMessage `json:"soal"`
	}
	if err := c.Bind(&in); err != nil {
		return nil, err
	}
	if len(in.Soal) == 0 || len(in.Soal) > 300 {
		return nil, bad("Jumlah soal harus 1 sampai 300.")
	}
	out := make([]mcq, 0, len(in.Soal))
	for i, raw := range in.Soal {
		q, err := parseMCQ(raw, jopsi, allowImage)
		if err != nil {
			return nil, bad("Soal nomor " + strconv.Itoa(i+1) + ": " + err.Error())
		}
		out = append(out, q)
	}
	return out, nil
}

func apiCBTSoalBatch(c *Ctx) error {
	examID := c.ID("id")
	if !c.ownsExam(examID) {
		return forbidden()
	}
	var jopsi int
	c.A.db.QueryRow(`SELECT jumlah_opsi FROM exams WHERE id=?`, examID).Scan(&jopsi)
	qs, err := parseBatch(c, jopsi, true)
	if err != nil {
		return err
	}
	tx, err := c.A.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, q := range qs {
		if _, err := tx.Exec(`INSERT INTO exam_questions (exam_id, pertanyaan, gambar, opsi_a, opsi_b, opsi_c, opsi_d, opsi_e, jawaban_benar, bobot, created_at)
			VALUES (?,?,?,?,?,?,?,?,?,?,?)`, examID, q.pertanyaan, q.gambar, q.opsi[0], q.opsi[1], q.opsi[2], q.opsi[3], q.opsi[4], q.kunci, q.bobot, nowStr()); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	c.Note = strconv.Itoa(len(qs)) + " soal"
	return c.JSON(map[string]int{"ditambahkan": len(qs)})
}

func apiQuizQuestionBatch(c *Ctx) error {
	quizID := c.ID("id")
	if err := c.requireManage(c.courseOf("quizzes", quizID), "kuis"); err != nil {
		return err
	}
	var jopsi int
	c.A.db.QueryRow(`SELECT jumlah_opsi FROM quizzes WHERE id=?`, quizID).Scan(&jopsi)
	qs, err := parseBatch(c, jopsi, false)
	if err != nil {
		return err
	}
	tx, err := c.A.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, q := range qs {
		if _, err := tx.Exec(`INSERT INTO quiz_questions (quiz_id, pertanyaan, opsi_a, opsi_b, opsi_c, opsi_d, opsi_e, jawaban_benar, bobot, created_at)
			VALUES (?,?,?,?,?,?,?,?,?,?)`, quizID, q.pertanyaan, q.opsi[0], q.opsi[1], q.opsi[2], q.opsi[3], q.opsi[4], q.kunci, q.bobot, nowStr()); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	c.Note = strconv.Itoa(len(qs)) + " soal"
	return c.JSON(map[string]int{"ditambahkan": len(qs)})
}
