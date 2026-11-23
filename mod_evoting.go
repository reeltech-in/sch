package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// E-Voting OSIS: pemilihan, kandidat, suara rahasia (siswa/guru atau bilik suara), hasil setelah selesai.

// ---------------------------------------------------------------------
// E-Voting
// ---------------------------------------------------------------------
func apiElections(c *Ctx) error {
	rows, err := c.list(`SELECT e.*, (SELECT COUNT(*) FROM election_voters v WHERE v.election_id=e.id) AS jumlah_memilih,
		(SELECT COUNT(*) FROM candidates k WHERE k.election_id=e.id) AS jumlah_kandidat FROM elections e ORDER BY e.mulai DESC`)
	if err != nil {
		return err
	}
	kand := c.candidatesByElection()
	for _, r := range rows {
		ids, _ := r["kelas_ids"].(string)
		r["kelas_ids"] = json.RawMessage(ids)
		r["kandidat"] = kand[r["id"].(int64)]
	}
	return c.JSON(rows)
}

func apiElectionSave(c *Ctx) error {
	var in struct {
		Judul       string  `json:"judul"`
		Deskripsi   *string `json:"deskripsi"`
		Mulai       *string `json:"mulai"`
		Selesai     *string `json:"selesai"`
		Pemilih     string  `json:"pemilih"`
		KelasIDs    []int64 `json:"kelas_ids"`
		TampilHasil bool    `json:"tampil_hasil"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	judul, err := required(in.Judul, "Nama pemilihan")
	if err != nil {
		return err
	}
	mulai, err := cleanDeadline(in.Mulai)
	if err != nil || mulai == nil {
		return bad("Waktu mulai wajib diisi.")
	}
	selesai, err := cleanDeadline(in.Selesai)
	if err != nil || selesai == nil || selesai.(string) <= mulai.(string) {
		return bad("Waktu selesai harus setelah waktu mulai.")
	}
	if in.Pemilih != "semua" {
		in.Pemilih = "siswa"
	}
	if in.KelasIDs == nil {
		in.KelasIDs = []int64{}
	}
	kb, _ := json.Marshal(in.KelasIDs)
	th := 0
	if in.TampilHasil {
		th = 1
	}
	if id := c.ID("id"); id > 0 {
		// Setelah ada suara masuk: pemilih, kelas, dan waktu mulai dikunci; pemilihan yang sudah selesai tidak bisa
		// dibuka lagi. Tanpa ini, staf bisa memundurkan waktu selesai untuk melihat perolehan suara di tengah
		// pemungutan (hasil disembunyikan sampai selesai) lalu membukanya kembali.
		var curMulai, curSelesai, curPemilih, curKelas string
		var suara int
		if c.A.db.QueryRow(`SELECT mulai, selesai, pemilih, COALESCE(kelas_ids,'[]'), (SELECT COUNT(*) FROM election_voters v WHERE v.election_id=e.id)
			FROM elections e WHERE id=?`, id).Scan(&curMulai, &curSelesai, &curPemilih, &curKelas, &suara) != nil {
			return notFound()
		}
		if suara > 0 {
			if mulai.(string) != curMulai || in.Pemilih != curPemilih || string(kb) != curKelas {
				return bad("Pemungutan suara sudah berjalan; pemilih, kelas, dan waktu mulai tidak bisa diubah.")
			}
			if selesai.(string) != curSelesai && nowStr() > curSelesai {
				return bad("Pemilihan sudah selesai; waktunya tidak bisa diubah lagi.")
			}
		}
		_, err = c.A.db.Exec(`UPDATE elections SET judul=?, deskripsi=?, mulai=?, selesai=?, pemilih=?, kelas_ids=?, tampil_hasil=? WHERE id=?`, judul, str(in.Deskripsi), mulai, selesai, in.Pemilih, string(kb), th, id)
	} else {
		_, err = c.A.db.Exec(`INSERT INTO elections (judul, deskripsi, mulai, selesai, pemilih, kelas_ids, tampil_hasil, created_at) VALUES (?,?,?,?,?,?,?,?)`, judul, str(in.Deskripsi), mulai, selesai, in.Pemilih, string(kb), th, nowStr())
	}
	if err != nil {
		return err
	}
	return c.OK()
}

func apiCandidateSave(c *Ctx) error {
	var in struct {
		No    int     `json:"no"`
		Nama  string  `json:"nama"`
		Wakil *string `json:"wakil"`
		Visi  *string `json:"visi"`
		Misi  *string `json:"misi"`
		Foto  *string `json:"foto"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	nama, err := required(in.Nama, "Nama kandidat")
	if err != nil {
		return err
	}
	if in.No < 1 {
		return bad("Nomor urut wajib diisi.")
	}
	if v := str(in.Foto); v != nil && !fileNameRe.MatchString(strings.TrimPrefix(v.(string), "/files/")) {
		return bad("Foto tidak valid.")
	}
	elID, selfID := c.ID("id"), int64(0)
	if c.R.Method == http.MethodPut {
		selfID = elID
		c.A.db.QueryRow(`SELECT election_id FROM candidates WHERE id=?`, selfID).Scan(&elID)
	}
	if n, ok := c.exists(`SELECT nama FROM candidates WHERE election_id=? AND no=? AND id<>?`, elID, in.No, selfID); ok {
		return bad(fmt.Sprintf("Nomor urut %d sudah dipakai %s.", in.No, n))
	}
	if _, ok := c.exists(`SELECT nama FROM candidates WHERE election_id=? AND nama=? COLLATE NOCASE AND id<>?`, elID, nama, selfID); ok {
		return bad(nama + " sudah terdaftar sebagai kandidat.")
	}
	if c.R.Method == http.MethodPut {
		// setelah ada suara masuk: nomor, nama, dan wakil dikunci (suara tidak boleh "pindah" ke identitas lain);
		// visi, misi, dan foto masih boleh diperbaiki
		var started int
		c.A.db.QueryRow(`SELECT COUNT(*) FROM election_voters WHERE election_id=?`, elID).Scan(&started)
		if started > 0 {
			var curNo int
			var curNama string
			var curWakil sql.NullString
			if c.A.db.QueryRow(`SELECT no, nama, wakil FROM candidates WHERE id=?`, selfID).Scan(&curNo, &curNama, &curWakil) != nil {
				return notFound()
			}
			wakil, _ := str(in.Wakil).(string)
			if in.No != curNo || nama != curNama || wakil != curWakil.String {
				return bad("Pemungutan suara sudah berjalan; nomor, nama, dan wakil kandidat tidak bisa diubah.")
			}
		}
		_, err = c.A.db.Exec(`UPDATE candidates SET no=?, nama=?, wakil=?, visi=?, misi=?, foto=? WHERE id=?`, in.No, nama, str(in.Wakil), str(in.Visi), str(in.Misi), str(in.Foto), c.ID("id"))
	} else {
		var started int
		c.A.db.QueryRow(`SELECT COUNT(*) FROM election_voters WHERE election_id=?`, c.ID("id")).Scan(&started)
		if started > 0 {
			return bad("Pemungutan suara sudah berjalan; kandidat tidak bisa ditambah.")
		}
		_, err = c.A.db.Exec(`INSERT INTO candidates (election_id, no, nama, wakil, visi, misi, foto) VALUES (?,?,?,?,?,?,?)`, c.ID("id"), in.No, nama, str(in.Wakil), str(in.Visi), str(in.Misi), str(in.Foto))
	}
	if err != nil {
		return err
	}
	return c.OK()
}

func apiCandidateDelete(c *Ctx) error {
	var suara int
	c.A.db.QueryRow(`SELECT suara FROM candidates WHERE id=?`, c.ID("id")).Scan(&suara)
	if suara > 0 {
		return bad("Kandidat sudah memperoleh suara; tidak bisa dihapus.")
	}
	return apiDelete("candidates")(c)
}

type election struct {
	id                    int64
	judul, mulai, selesai string
	pemilih, kelasIDs     string
	tampil                int
}

func (c *Ctx) loadElection(id int64) (*election, error) {
	e := &election{id: id}
	err := c.A.db.QueryRow(`SELECT judul, mulai, selesai, pemilih, kelas_ids, tampil_hasil FROM elections WHERE id=?`, id).
		Scan(&e.judul, &e.mulai, &e.selesai, &e.pemilih, &e.kelasIDs, &e.tampil)
	if err != nil {
		return nil, notFound()
	}
	return e, nil
}

// candidatesByElection: semua kandidat dikelompokkan per pemilihan (satu query, bukan satu per pemilihan).
func (c *Ctx) candidatesByElection() map[int64][]map[string]any {
	out := map[int64][]map[string]any{}
	rows, err := c.list(`SELECT election_id, id, no, nama, wakil, visi, misi, foto FROM candidates ORDER BY election_id, no`)
	if err != nil {
		return out
	}
	for _, r := range rows {
		eid := r["election_id"].(int64)
		delete(r, "election_id")
		out[eid] = append(out[eid], r)
	}
	return out
}

// eligible: apakah pengguna boleh memilih pada pemilihan ini.
func (e *election) eligible(db *sql.DB, userID int64) bool {
	var role string
	var kelas sql.NullInt64
	if db.QueryRow(`SELECT role, kelas_id FROM users WHERE id=? AND status='aktif'`, userID).Scan(&role, &kelas) != nil {
		return false
	}
	return e.eligibleFor(role, kelas)
}

func (e *election) eligibleFor(role string, kelas sql.NullInt64) bool {
	if role == "guru" {
		return e.pemilih == "semua"
	}
	if role != "siswa" {
		return false
	}
	var ids []int64
	json.Unmarshal([]byte(e.kelasIDs), &ids)
	return len(ids) == 0 || (kelas.Valid && kelasIn(e.kelasIDs, kelas.Int64))
}

func (e *election) state() string {
	now := nowStr()
	switch {
	case now < e.mulai:
		return "belum"
	case now > e.selesai:
		return "selesai"
	}
	return "berlangsung"
}

func apiElectionsForMe(c *Ctx) error {
	rows, err := c.list(`SELECT id, judul, deskripsi, mulai, selesai, tampil_hasil, pemilih, kelas_ids FROM elections ORDER BY mulai DESC`)
	if err != nil {
		return err
	}
	// sekali baca: peran & kelas pengguna, pemilihan yang sudah diikuti, dan semua kandidat
	out := []map[string]any{}
	voted := map[int64]bool{}
	if vr, err := c.list(`SELECT election_id FROM election_voters WHERE user_id=?`, c.U.ID); err == nil {
		for _, v := range vr {
			voted[v["election_id"].(int64)] = true
		}
	}
	kand := c.candidatesByElection()
	for _, r := range rows {
		e := &election{id: r["id"].(int64), mulai: r["mulai"].(string), selesai: r["selesai"].(string), pemilih: r["pemilih"].(string), kelasIDs: r["kelas_ids"].(string)}
		if !e.eligibleFor(c.U.Role, c.U.KelasID) {
			continue
		}
		r["sudah"] = voted[e.id]
		r["status"] = e.state()
		delete(r, "kelas_ids")
		r["kandidat"] = kand[e.id]
		out = append(out, r)
	}
	return c.JSON(out)
}

// castVote mencatat bahwa pemilih sudah memilih dan menambah suara kandidat, dalam satu transaksi.
func init() {
	moduleMigrations = append(moduleMigrations, func(db *sql.DB) error { return addColumn(db, "election_voters", "petugas_id", "INTEGER") })
}

// castVote mencatat satu suara. petugasID = petugas bilik yang memindai kartu (0 bila pemilih memilih sendiri);
// dicatat di daftar hadir pemilih (bukan pada pilihannya), agar pencoblosan massal oleh petugas bisa dilacak.
func castVote(db *sql.DB, e *election, userID, candID int64, metode string, petugasID int64) error {
	if e.state() != "berlangsung" {
		return bad("Pemungutan suara tidak sedang berlangsung.")
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var ok int
	tx.QueryRow(`SELECT COUNT(*) FROM candidates WHERE id=? AND election_id=?`, candID, e.id).Scan(&ok)
	if ok == 0 {
		return bad("Kandidat tidak valid.")
	}
	var petugas any
	if petugasID > 0 {
		petugas = petugasID
	}
	if _, err := tx.Exec(`INSERT INTO election_voters (election_id, user_id, voted_at, metode, petugas_id) VALUES (?,?,?,?,?)`, e.id, userID, nowStr(), metode, petugas); err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return bad("Pemilih ini sudah menggunakan hak suaranya.")
		}
		return err
	}
	if _, err := tx.Exec(`UPDATE candidates SET suara=suara+1 WHERE id=?`, candID); err != nil {
		return err
	}
	return tx.Commit()
}

func apiVote(c *Ctx) error {
	var in struct {
		CandidateID int64 `json:"candidate_id"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	e, err := c.loadElection(c.ID("id"))
	if err != nil {
		return err
	}
	if !e.eligible(c.A.db, c.U.ID) {
		return forbidden()
	}
	if err := castVote(c.A.db, e, c.U.ID, in.CandidateID, "online", 0); err != nil {
		return err
	}
	return c.OK()
}

// jeda minimum antar-suara per petugas bilik (pemilih sungguhan butuh waktu memindai kartu & memilih)
var boothJeda = 10 * time.Second

// apiVoteBooth: mode bilik suara. Petugas scan kartu, lalu pemilih memilih di layar yang sama.
func (a *App) apiVoteBooth(c *Ctx) error {
	var in struct {
		Token       string `json:"token"`
		CandidateID int64  `json:"candidate_id"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	e, err := c.loadElection(c.ID("id"))
	if err != nil {
		return err
	}
	uid := a.cardUser(in.Token)
	if uid == 0 {
		return bad("Kartu tidak valid.")
	}
	tercatat := false
	if in.CandidateID != 0 {
		// jeda antar-suara per petugas bilik: mencegah pencoblosan massal dengan skrip
		boothMu.Lock()
		last := boothLast[c.U.ID]
		if time.Since(last) < boothJeda {
			boothMu.Unlock()
			return bad("Tunggu sebentar sebelum pemilih berikutnya.")
		}
		boothLast[c.U.ID] = time.Now() // dipesan dulu (dua kiriman bersamaan tidak lolos); dibatalkan bila suara gagal
		boothMu.Unlock()
		defer func() {
			if !tercatat { // kartu ditolak / sudah memilih: pemilih berikutnya tidak perlu menunggu
				boothMu.Lock()
				boothLast[c.U.ID] = last // kembali ke waktu suara sah terakhir
				boothMu.Unlock()
			}
		}()
	}
	var nama string
	var kelas sql.NullString
	a.db.QueryRow(`SELECT u.nama_lengkap, k.nama FROM users u LEFT JOIN classes k ON k.id=u.kelas_id WHERE u.id=?`, uid).Scan(&nama, &kelas)
	if !e.eligible(a.db, uid) {
		return bad(nama + " tidak terdaftar sebagai pemilih pada pemilihan ini.")
	}
	var voted int
	a.db.QueryRow(`SELECT COUNT(*) FROM election_voters WHERE election_id=? AND user_id=?`, e.id, uid).Scan(&voted)
	if voted > 0 {
		return bad(nama + " sudah menggunakan hak suaranya.")
	}
	if in.CandidateID == 0 {
		return c.JSON(map[string]any{"nama": nama, "kelas": kelas.String, "siap": true})
	}
	if err := castVote(a.db, e, uid, in.CandidateID, "bilik", c.U.ID); err != nil {
		return err
	}
	tercatat = true
	return c.JSON(map[string]any{"nama": nama, "tercatat": true})
}

func apiElectionResult(c *Ctx) error {
	e, err := c.loadElection(c.ID("id"))
	if err != nil {
		return err
	}
	if !c.U.Has("evoting") && !(e.tampil == 1 && e.state() == "selesai" && e.eligible(c.A.db, c.U.ID)) {
		return apiErr{403, "Hasil belum bisa dilihat."}
	}
	kand, err := c.list(`SELECT id, no, nama, wakil, foto, suara FROM candidates WHERE election_id=? ORDER BY no`, e.id)
	if err != nil {
		return err
	}
	// Perolehan suara disembunyikan dari SEMUA orang sampai pemilihan selesai. Di mode bilik suara
	// masuk satu per satu, sehingga melihat angka bergerak akan membuka pilihan tiap pemilih.
	if e.state() != "selesai" {
		for _, k := range kand {
			k["suara"] = nil
		}
	}
	var voted int
	c.A.db.QueryRow(`SELECT COUNT(*) FROM election_voters WHERE election_id=?`, e.id).Scan(&voted)
	// jumlah pemilih terdaftar
	var total int
	if e.pemilih == "semua" {
		c.A.db.QueryRow(`SELECT COUNT(*) FROM users WHERE status='aktif' AND role IN ('siswa','guru')`).Scan(&total)
	} else {
		var ids []int64
		json.Unmarshal([]byte(e.kelasIDs), &ids)
		if len(ids) == 0 {
			c.A.db.QueryRow(`SELECT COUNT(*) FROM users WHERE status='aktif' AND role='siswa'`).Scan(&total)
		} else {
			ph := strings.TrimSuffix(strings.Repeat("?,", len(ids)), ",")
			args := []any{}
			for _, k := range ids {
				args = append(args, k)
			}
			c.A.db.QueryRow(`SELECT COUNT(*) FROM users WHERE status='aktif' AND role='siswa' AND kelas_id IN (`+ph+`)`, args...).Scan(&total)
		}
	}
	out := map[string]any{"judul": e.judul, "status": e.state(), "kandidat": kand, "memilih": voted, "pemilih": total}
	if c.U.Has("evoting") {
		bilik, err := c.list(`SELECT COALESCE(u.nama_lengkap, '(tidak tercatat)') AS petugas, COUNT(*) AS jumlah, MIN(v.voted_at) AS pertama, MAX(v.voted_at) AS terakhir
			FROM election_voters v LEFT JOIN users u ON u.id=v.petugas_id WHERE v.election_id=? AND v.metode='bilik' GROUP BY v.petugas_id ORDER BY jumlah DESC`, e.id)
		if err != nil {
			return err
		}
		out["bilik"] = bilik
	}
	return c.JSON(out)
}
