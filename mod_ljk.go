package main

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
)

// LJK kertas untuk ujian CBT: guru mencetak lembar jawaban, siswa menghitamkan bulatan, lalu guru memotret/memindai
// lembar itu. Pembacaan gambar dilakukan di browser (web/mod-ljk.js); server hanya menerima jawaban per nomor soal,
// memeriksa hak akses, menilai dengan kunci & bobot yang sama seperti CBT daring, lalu menyimpannya sebagai hasil ujian.

const ljkMaxSoal = 60 // kapasitas satu lembar (3 kolom × 20 nomor)

func init() {
	moduleMigrations = append(moduleMigrations, func(db *sql.DB) error {
		return addColumn(db, "exam_attempts", "metode", "TEXT") // NULL/daring = CBT, 'ljk' = lembar kertas
	})
	moduleRoutes = append(moduleRoutes, func(a *App, m *http.ServeMux) {
		m.HandleFunc("GET /api/cbt/{id}/ljk", a.mod("ujian", "admin,guru", apiLJKInfo))
		m.HandleFunc("POST /api/cbt/{id}/ljk", a.mod("ujian", "admin,guru", apiLJKSave))
	})
}

// apiLJKInfo: data ujian + peserta untuk mencetak dan memindai LJK.
func apiLJKInfo(c *Ctx) error {
	id := c.ID("id")
	if !c.ownsExam(id) {
		return forbidden()
	}
	ex, err := c.list(`SELECT e.id, e.judul, e.mapel, e.jumlah_opsi, e.kelas_ids,
		(SELECT COUNT(*) FROM exam_questions q WHERE q.exam_id=e.id) AS jumlah_soal FROM exams e WHERE e.id=?`, id)
	if err != nil || len(ex) == 0 {
		return notFound()
	}
	ids, _ := ex[0]["kelas_ids"].(string)
	rows, err := c.list(`SELECT u.id, u.nama_lengkap, u.username, k.nama AS kelas_nama, t.selesai, t.skor, t.metode
		FROM users u JOIN classes k ON k.id=u.kelas_id
		LEFT JOIN exam_attempts t ON t.exam_id=? AND t.siswa_id=u.id
		WHERE u.role='siswa' AND u.status='aktif' AND u.kelas_id IN (SELECT value FROM json_each(?))
		ORDER BY k.nama, u.nama_lengkap`, id, ids)
	if err != nil {
		return err
	}
	delete(ex[0], "kelas_ids")
	ex[0]["maks_soal"] = ljkMaxSoal
	return c.JSON(map[string]any{"ujian": ex[0], "peserta": rows})
}

// apiLJKSave menyimpan jawaban hasil pindaian satu siswa. jawaban[i] = huruf untuk soal nomor i+1 ("" = kosong).
func apiLJKSave(c *Ctx) error {
	id := c.ID("id")
	if !c.ownsExam(id) {
		return forbidden()
	}
	var in struct {
		SiswaID int64    `json:"siswa_id"`
		Jawaban []string `json:"jawaban"`
		Timpa   bool     `json:"timpa"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	var jopsi int
	var kelasIDs string
	if c.A.db.QueryRow(`SELECT jumlah_opsi, kelas_ids FROM exams WHERE id=?`, id).Scan(&jopsi, &kelasIDs) != nil {
		return notFound()
	}
	var kelas sql.NullInt64
	var role, status string
	if c.A.db.QueryRow(`SELECT role, status, kelas_id FROM users WHERE id=?`, in.SiswaID).Scan(&role, &status, &kelas) != nil ||
		role != "siswa" || status != "aktif" || !kelas.Valid || !kelasIn(kelasIDs, kelas.Int64) {
		return bad("Siswa ini bukan peserta ujian.")
	}
	// soal berurutan sesuai nomor di lembar (urutan id, sama seperti CBT tanpa acak)
	type q struct {
		id    int64
		kunci string
		bobot float64
	}
	var qs []q
	rows, err := c.A.db.Query(`SELECT id, jawaban_benar, bobot FROM exam_questions WHERE exam_id=? ORDER BY id`, id)
	if err != nil {
		return err
	}
	for rows.Next() {
		var x q
		if err := rows.Scan(&x.id, &x.kunci, &x.bobot); err != nil {
			rows.Close()
			return err
		}
		qs = append(qs, x)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	if len(qs) == 0 {
		return bad("Ujian ini belum memiliki soal.")
	}
	if len(qs) > ljkMaxSoal {
		return bad("LJK kertas mendukung maksimal " + strconv.Itoa(ljkMaxSoal) + " soal.")
	}
	if len(in.Jawaban) > len(qs) {
		return bad("Jumlah jawaban melebihi jumlah soal.")
	}
	letters := "abcde"[:jopsi]
	ans := map[string]string{}
	urutan := make([]int64, len(qs))
	omap := map[string]string{}
	var total, benar float64
	nBenar := 0
	for i, x := range qs {
		key := strconv.FormatInt(x.id, 10)
		urutan[i], omap[key] = x.id, letters
		total += x.bobot
		if i >= len(in.Jawaban) {
			continue
		}
		v := strings.ToLower(strings.TrimSpace(in.Jawaban[i]))
		if v == "" {
			continue
		}
		if len(v) != 1 || !strings.Contains(letters, v) {
			return bad("Jawaban nomor " + strconv.Itoa(i+1) + " tidak valid.")
		}
		ans[key] = v
		if v == x.kunci {
			benar += x.bobot
			nBenar++
		}
	}
	skor := 0.0
	if total > 0 {
		skor = math.Round(benar/total*10000) / 100
	}
	jb, _ := json.Marshal(ans)
	ub, _ := json.Marshal(urutan)
	ob, _ := json.Marshal(omap)
	tx, err := c.A.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var done sql.NullInt64
	err = tx.QueryRow(`SELECT selesai FROM exam_attempts WHERE exam_id=? AND siswa_id=?`, id, in.SiswaID).Scan(&done)
	switch {
	case err == sql.ErrNoRows:
		_, err = tx.Exec(`INSERT INTO exam_attempts (exam_id, siswa_id, started_at, submitted_at, selesai, skor, jawaban, urutan, opsi_map, metode)
			VALUES (?,?,?,?,1,?,?,?,?,'ljk')`, id, in.SiswaID, nowStr(), nowStr(), skor, string(jb), string(ub), string(ob))
	case err != nil:
		return err
	case done.Int64 == 1 && !in.Timpa:
		return apiErr{409, "Siswa ini sudah punya nilai untuk ujian ini."}
	default:
		_, err = tx.Exec(`UPDATE exam_attempts SET submitted_at=?, selesai=1, skor=?, jawaban=?, urutan=?, opsi_map=?, metode='ljk' WHERE exam_id=? AND siswa_id=?`,
			nowStr(), skor, string(jb), string(ub), string(ob), id, in.SiswaID)
	}
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	c.Note = "LJK skor " + strconv.FormatFloat(skor, 'f', -1, 64)
	return c.JSON(map[string]any{"skor": skor, "benar": nBenar, "jumlah_soal": len(qs), "dijawab": len(ans)})
}
