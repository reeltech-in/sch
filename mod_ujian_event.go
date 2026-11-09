package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Manajemen ujian: kegiatan ujian, ruang, pembagian peserta, jadwal, cetak kartu/denah/daftar hadir.

// ---------------------------------------------------------------------
// Manajemen ujian: kegiatan, ruang, peserta, jadwal
// ---------------------------------------------------------------------
func apiEvents(c *Ctx) error {
	rows, err := c.list(`SELECT e.*, (SELECT COUNT(*) FROM exam_participants p WHERE p.event_id=e.id) AS jumlah_peserta,
		(SELECT COUNT(*) FROM exam_rooms r WHERE r.event_id=e.id) AS jumlah_ruang FROM exam_events e ORDER BY e.created_at DESC`)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

func apiEventSave(c *Ctx) error {
	var in struct {
		Nama        string  `json:"nama"`
		TahunAjaran *string `json:"tahun_ajaran"`
		Semester    *string `json:"semester"`
		Mulai       *string `json:"mulai"`
		Selesai     *string `json:"selesai"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	nama, err := required(in.Nama, "Nama kegiatan ujian")
	if err != nil {
		return err
	}
	dates := []any{nil, nil}
	for i, p := range []*string{in.Mulai, in.Selesai} {
		if v := str(p); v != nil {
			d, err := cleanDate(v.(string))
			if err != nil {
				return err
			}
			dates[i] = d
		}
	}
	if err := cekRentang(dates[0], dates[1], "Tanggal"); err != nil {
		return err
	}
	if id := c.ID("id"); id > 0 {
		_, err = c.A.db.Exec(`UPDATE exam_events SET nama=?, tahun_ajaran=?, semester=?, mulai=?, selesai=? WHERE id=?`, nama, str(in.TahunAjaran), str(in.Semester), dates[0], dates[1], id)
	} else {
		_, err = c.A.db.Exec(`INSERT INTO exam_events (nama, tahun_ajaran, semester, mulai, selesai, created_at) VALUES (?,?,?,?,?,?)`, nama, str(in.TahunAjaran), str(in.Semester), dates[0], dates[1], nowStr())
	}
	if err != nil {
		return err
	}
	return c.OK()
}

func (a *App) apiEventDetail(c *Ctx) error {
	id := c.ID("id")
	ev, err := c.list(`SELECT * FROM exam_events WHERE id=?`, id)
	if err != nil || len(ev) == 0 {
		return notFound()
	}
	rooms, err := c.list(`SELECT r.*, (SELECT COUNT(*) FROM exam_participants p WHERE p.room_id=r.id) AS terisi FROM exam_rooms r WHERE r.event_id=? ORDER BY r.nama`, id)
	if err != nil {
		return err
	}
	parts, err := c.list(`SELECT p.siswa_id, p.room_id, p.no_peserta, p.kursi, u.username, u.nama_lengkap, k.nama AS kelas_nama, sp.nisn, sp.foto
		FROM exam_participants p JOIN users u ON u.id=p.siswa_id LEFT JOIN classes k ON k.id=u.kelas_id LEFT JOIN student_profiles sp ON sp.siswa_id=u.id
		WHERE p.event_id=? ORDER BY p.no_peserta`, id)
	if err != nil {
		return err
	}
	for _, p := range parts {
		p["token"] = a.cardToken(p["siswa_id"].(int64))
	}
	sched, err := c.list(`SELECT * FROM exam_schedule WHERE event_id=? ORDER BY tanggal, jam_mulai`, id)
	if err != nil {
		return err
	}
	return c.JSON(map[string]any{"event": ev[0], "rooms": rooms, "participants": parts, "schedule": sched})
}

func apiRoomSave(c *Ctx) error {
	var in struct {
		Nama      string  `json:"nama"`
		Kapasitas int     `json:"kapasitas"`
		Kolom     int     `json:"kolom"`
		Pengawas  *string `json:"pengawas"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	nama, err := required(in.Nama, "Nama ruang")
	if err != nil {
		return err
	}
	if in.Kapasitas < 1 || in.Kapasitas > 500 {
		return bad("Kapasitas harus 1 sampai 500.")
	}
	if in.Kolom < 1 || in.Kolom > 12 {
		in.Kolom = 4
	}
	evID, selfID := c.ID("id"), int64(0)
	if c.R.Method == http.MethodPut {
		selfID = evID
		c.A.db.QueryRow(`SELECT event_id FROM exam_rooms WHERE id=?`, selfID).Scan(&evID)
	}
	if _, ok := c.exists(`SELECT nama FROM exam_rooms WHERE event_id=? AND nama=? COLLATE NOCASE AND id<>?`, evID, nama, selfID); ok {
		return bad("Ruang " + nama + " sudah ada di kegiatan ujian ini.")
	}
	if c.R.Method == http.MethodPut {
		_, err = c.A.db.Exec(`UPDATE exam_rooms SET nama=?, kapasitas=?, kolom=?, pengawas=? WHERE id=?`, nama, in.Kapasitas, in.Kolom, str(in.Pengawas), c.ID("id"))
	} else {
		_, err = c.A.db.Exec(`INSERT INTO exam_rooms (event_id, nama, kapasitas, kolom, pengawas) VALUES (?,?,?,?,?)`, c.ID("id"), nama, in.Kapasitas, in.Kolom, str(in.Pengawas))
	}
	if err != nil {
		return err
	}
	return c.OK()
}

func apiScheduleSave(c *Ctx) error {
	var in struct {
		Tanggal    string  `json:"tanggal"`
		JamMulai   string  `json:"jam_mulai"`
		JamSelesai string  `json:"jam_selesai"`
		Mapel      string  `json:"mapel"`
		Kelas      *string `json:"kelas"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	tgl, err := cleanDate(in.Tanggal)
	if err != nil {
		return err
	}
	mapel, err := required(in.Mapel, "Mata pelajaran")
	if err != nil {
		return err
	}
	for _, j := range []string{in.JamMulai, in.JamSelesai} {
		if _, err := time.Parse("15:04", j); err != nil {
			return bad("Format jam harus JJ:MM.")
		}
	}
	if in.JamSelesai <= in.JamMulai {
		return bad("Jam selesai harus setelah jam mulai.")
	}
	var evMulai, evSelesai sql.NullString
	c.A.db.QueryRow(`SELECT mulai, selesai FROM exam_events WHERE id=?`, c.ID("id")).Scan(&evMulai, &evSelesai)
	if (evMulai.Valid && evMulai.String != "" && tgl < evMulai.String) || (evSelesai.Valid && evSelesai.String != "" && tgl > evSelesai.String) {
		return bad("Tanggal jadwal di luar rentang kegiatan ujian (" + evMulai.String + " s.d. " + evSelesai.String + ").")
	}
	if _, ok := c.exists(`SELECT mapel FROM exam_schedule WHERE event_id=? AND tanggal=? AND mapel=? COLLATE NOCASE AND COALESCE(kelas,'')=COALESCE(?, '')`, c.ID("id"), tgl, mapel, str(in.Kelas)); ok {
		return bad("Jadwal " + mapel + " pada tanggal itu sudah ada.")
	}
	_, err = c.A.db.Exec(`INSERT INTO exam_schedule (event_id, tanggal, jam_mulai, jam_selesai, mapel, kelas) VALUES (?,?,?,?,?,?)`, c.ID("id"), tgl, in.JamMulai, in.JamSelesai, mapel, str(in.Kelas))
	if err != nil {
		return err
	}
	return c.OK()
}

// apiEventAssign membagi siswa dari kelas terpilih ke ruang-ruang secara otomatis.
func apiEventAssign(c *Ctx) error {
	id := c.ID("id")
	var in struct {
		KelasIDs []int64 `json:"kelas_ids"`
		Silang   bool    `json:"silang"`
		Prefix   string  `json:"prefix"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	if len(in.KelasIDs) == 0 {
		return bad("Pilih minimal satu kelas.")
	}
	tx, err := c.A.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	type room struct {
		id  int64
		cap int
	}
	var rooms []room
	rr, err := tx.Query(`SELECT id, kapasitas FROM exam_rooms WHERE event_id=? ORDER BY nama`, id)
	if err != nil {
		return err
	}
	total := 0
	for rr.Next() {
		var r room
		rr.Scan(&r.id, &r.cap)
		rooms = append(rooms, r)
		total += r.cap
	}
	rr.Close()
	if len(rooms) == 0 {
		return bad("Tambahkan ruang ujian terlebih dahulu.")
	}
	// siswa per kelas
	var groups [][]int64
	count := 0
	for _, k := range in.KelasIDs {
		rows, err := tx.Query(`SELECT id FROM users WHERE role='siswa' AND status='aktif' AND kelas_id=? ORDER BY nama_lengkap`, k)
		if err != nil {
			return err
		}
		var g []int64
		for rows.Next() {
			var s int64
			rows.Scan(&s)
			g = append(g, s)
		}
		rows.Close()
		groups = append(groups, g)
		count += len(g)
	}
	if count > total {
		return bad(fmt.Sprintf("Kapasitas ruang (%d) kurang dari jumlah peserta (%d). Tambah ruang atau kapasitas.", total, count))
	}
	var order []int64
	if in.Silang {
		for i := 0; ; i++ {
			added := false
			for _, g := range groups {
				if i < len(g) {
					order = append(order, g[i])
					added = true
				}
			}
			if !added {
				break
			}
		}
	} else {
		for _, g := range groups {
			order = append(order, g...)
		}
	}
	if _, err := tx.Exec(`DELETE FROM exam_participants WHERE event_id=?`, id); err != nil {
		return err
	}
	prefix := strings.TrimSpace(in.Prefix)
	if prefix == "" {
		prefix = "P"
	}
	ri, seat := 0, 0
	for n, sid := range order {
		for seat >= rooms[ri].cap {
			ri, seat = ri+1, 0
		}
		seat++
		no := fmt.Sprintf("%s-%03d", prefix, n+1)
		if _, err := tx.Exec(`INSERT INTO exam_participants (event_id, siswa_id, room_id, no_peserta, kursi) VALUES (?,?,?,?,?)`, id, sid, rooms[ri].id, no, seat); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.JSON(map[string]any{"peserta": len(order)})
}
