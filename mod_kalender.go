package main

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Kalender akademik: kegiatan yang diisi pengelola + kegiatan OTOMATIS dari modul lain (tanpa input ganda):
// ujian CBT, kegiatan & jadwal Manajemen Ujian, masa PPDB, masa E-Voting, pengumuman kelulusan, tenggat tugas.
// Kegiatan otomatis hanya ditampilkan (tidak disimpan di tabel events), jadi selalu mengikuti data aslinya, dan
// tidak memengaruhi hitungan hari libur absensi. Tiap pengguna hanya melihat yang memang boleh ia lihat.

func init() {
	moduleRoutes = append(moduleRoutes, func(a *App, m *http.ServeMux) {
		m.HandleFunc("POST /api/kalender/libur-nasional", a.mod("kalender", "admin", apiNationalHolidays))
	})
}

// ---------------------------------------------------------------------
// Kalender akademik
// ---------------------------------------------------------------------
func apiKalender(c *Ctx) error {
	from, to := c.R.URL.Query().Get("dari"), c.R.URL.Query().Get("sampai")
	if _, err := cleanDate(from); err != nil {
		from = "0000-01-01"
	}
	if _, err := cleanDate(to); err != nil {
		to = "9999-12-31"
	}
	rows, err := c.list(`SELECT id, judul, mulai, selesai, kategori, keterangan FROM events WHERE selesai>=? AND mulai<=? ORDER BY mulai`, from, to)
	if err != nil {
		return err
	}
	if c.R.URL.Query().Get("otomatis") != "0" {
		rows = append(rows, c.autoEvents(from, to)...)
		sort.SliceStable(rows, func(i, j int) bool { return rows[i]["mulai"].(string) < rows[j]["mulai"].(string) })
	}
	return c.JSON(rows)
}

// localDay: waktu RFC3339 (UTC) → tanggal lokal sekolah YYYY-MM-DD; tanggal biasa dikembalikan apa adanya.
func localDay(s string) string {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.Local().Format("2006-01-02")
	}
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

// tglJamLokal: "16/11/2026 11:23" dari waktu RFC3339 (untuk keterangan yang dibaca orang).
func tglJamLokal(s string) string {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.Local().Format("2/1/2006 15:04")
	}
	return s
}

// jamLokal: jam lokal HH:MM dari waktu RFC3339 (kosong bila bukan RFC3339).
func jamLokal(s string) string {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.Local().Format("15:04")
	}
	return ""
}

func (c *Ctx) autoEvents(from, to string) []map[string]any {
	on := modulesEnabled(c.A.db)
	out := []map[string]any{}
	add := func(id, judul, mulai, selesai, kategori, ket, halaman string) {
		if mulai == "" {
			return
		}
		if selesai == "" || selesai < mulai {
			selesai = mulai
		}
		if selesai < from || mulai > to {
			return
		}
		out = append(out, map[string]any{"id": "auto-" + id, "judul": judul, "mulai": mulai, "selesai": selesai, "kategori": kategori,
			"keterangan": ket, "otomatis": true, "halaman": halaman})
	}
	// batas kasar untuk kolom waktu RFC3339 (dibandingkan sebagai teks); ketepatan dicek ulang di add()
	// waktu disimpan UTC (WIB/WITA/WIT = UTC+7..9): ujian pukul 06.00 tanggal 1 tersimpan tanggal 31 UTC.
	// Saringan kasar diperlebar sehari; add() memeriksa tanggal lokal yang tepat.
	lo, hi := from, to+"~"
	if t, err := time.Parse("2006-01-02", from); err == nil {
		lo = t.AddDate(0, 0, -1).Format("2006-01-02")
	}
	if t, err := time.Parse("2006-01-02", to); err == nil {
		hi = t.AddDate(0, 0, 1).Format("2006-01-02") + "~"
	}
	staff := c.U.Role != "siswa"

	if on["ujian"] {
		if rows, err := c.list(`SELECT id, judul, mapel, mulai, selesai, kelas_ids FROM exams WHERE selesai>=? AND mulai<=?`, lo, hi); err == nil {
			for _, r := range rows {
				ids, _ := r["kelas_ids"].(string)
				if !staff && !(c.U.KelasID.Valid && kelasIn(ids, c.U.KelasID.Int64)) {
					continue
				}
				m, s := r["mulai"].(string), r["selesai"].(string)
				ket := "Dibuka " + jamLokal(m) + " · ditutup " + tglJamLokal(s)
				if mp, _ := r["mapel"].(string); mp != "" {
					ket = mp + " · " + ket
				}
				add("cbt-"+strconv.FormatInt(r["id"].(int64), 10), "Ujian CBT: "+r["judul"].(string), localDay(m), localDay(s), "ujian", ket, "cbt")
			}
		}
		if rows, err := c.list(`SELECT id, nama, mulai, selesai FROM exam_events WHERE mulai IS NOT NULL AND mulai<>'' AND COALESCE(NULLIF(selesai,''), mulai)>=? AND mulai<=?`, from, to); err == nil {
			for _, r := range rows {
				s, _ := r["selesai"].(string)
				add("ev-"+strconv.FormatInt(r["id"].(int64), 10), r["nama"].(string), r["mulai"].(string), s, "ujian", "Kegiatan ujian", "manujian")
			}
		}
		if rows, err := c.list(`SELECT s.id, s.tanggal, s.jam_mulai, s.jam_selesai, s.mapel, s.kelas, e.nama FROM exam_schedule s JOIN exam_events e ON e.id=s.event_id
			WHERE s.tanggal>=? AND s.tanggal<=?`, from, to); err == nil {
			for _, r := range rows {
				judul := "Jadwal ujian: " + r["mapel"].(string)
				if k, _ := r["kelas"].(string); k != "" {
					judul += " (" + k + ")"
				}
				add("jdw-"+strconv.FormatInt(r["id"].(int64), 10), judul, r["tanggal"].(string), "", "ujian",
					r["jam_mulai"].(string)+"–"+r["jam_selesai"].(string)+" · "+r["nama"].(string), "manujian")
			}
		}
	}
	if on["ppdb"] {
		cfg := getConfig(c.A.db, "ppdb")
		b, _ := cfg["buka"].(string)
		t, _ := cfg["tutup"].(string)
		if b != "" {
			ket := "Dibuka " + jamLokal(b)
			if t != "" {
				ket += " · ditutup " + tglJamLokal(t)
			}
			add("ppdb", "Pendaftaran PPDB", localDay(b), localDay(t), "lainnya", ket, "ppdb")
		}
	}
	if on["kelulusan"] {
		cfg := getConfig(c.A.db, "kelulusan")
		if b, _ := cfg["buka"].(string); b != "" {
			add("lulus", "Pengumuman kelulusan", localDay(b), "", "kegiatan", "Dibuka pukul "+jamLokal(b), "kelulusan")
		}
	}
	if on["evoting"] {
		if rows, err := c.list(`SELECT id, judul, mulai, selesai, pemilih, kelas_ids FROM elections WHERE selesai>=? AND mulai<=?`, lo, hi); err == nil {
			for _, r := range rows {
				e := &election{id: r["id"].(int64), judul: r["judul"].(string), mulai: r["mulai"].(string), selesai: r["selesai"].(string),
					pemilih: r["pemilih"].(string), kelasIDs: r["kelas_ids"].(string)}
				if !c.U.Has("evoting") && !e.eligibleFor(c.U.Role, c.U.KelasID) {
					continue
				}
				add("vote-"+strconv.FormatInt(e.id, 10), "E-Voting: "+e.judul, localDay(e.mulai), localDay(e.selesai), "kegiatan",
					"Pemungutan suara "+tglJamLokal(e.mulai)+" – "+tglJamLokal(e.selesai), "evoting")
			}
		}
	}
	// tenggat tugas: siswa → mapel kelasnya; guru → mapel yang ia ampu (admin tidak, terlalu banyak)
	if on["lms"] && c.U.Role != "admin" {
		q := `SELECT a.id, a.judul, a.deadline, co.nama AS mapel FROM assignments a JOIN courses co ON co.id=a.course_id
			WHERE a.deadline IS NOT NULL AND a.deadline<>'' AND a.deadline>=? AND a.deadline<=? AND `
		var args []any
		if c.U.Role == "siswa" {
			q += `co.kelas_id=?`
			args = []any{lo, hi, c.U.KelasID.Int64}
		} else {
			q += `co.guru_id=?`
			args = []any{lo, hi, c.U.ID}
		}
		if rows, err := c.list(q, args...); err == nil {
			for _, r := range rows {
				d := r["deadline"].(string)
				add("tugas-"+strconv.FormatInt(r["id"].(int64), 10), "Tenggat: "+r["judul"].(string), localDay(d), "", "tugas",
					r["mapel"].(string)+" · pukul "+jamLokal(d), "tugas")
			}
		}
	}
	return out
}

func apiSaveEvent(c *Ctx) error {
	var in struct {
		Judul      string  `json:"judul"`
		Mulai      string  `json:"mulai"`
		Selesai    string  `json:"selesai"`
		Kategori   string  `json:"kategori"`
		Keterangan *string `json:"keterangan"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	judul, err := required(in.Judul, "Nama kegiatan")
	if err != nil {
		return err
	}
	mulai, err := cleanDate(in.Mulai)
	if err != nil {
		return bad("Tanggal mulai tidak valid.")
	}
	selesai := mulai
	if strings.TrimSpace(in.Selesai) != "" {
		if selesai, err = cleanDate(in.Selesai); err != nil {
			return bad("Tanggal selesai tidak valid.")
		}
	}
	if selesai < mulai {
		return bad("Tanggal selesai tidak boleh sebelum tanggal mulai.")
	}
	switch in.Kategori {
	case "kegiatan", "libur", "ujian", "rapat", "lainnya":
	default:
		in.Kategori = "kegiatan"
	}
	if id := c.ID("id"); id > 0 {
		_, err = c.A.db.Exec(`UPDATE events SET judul=?, mulai=?, selesai=?, kategori=?, keterangan=? WHERE id=?`, judul, mulai, selesai, in.Kategori, str(in.Keterangan), id)
	} else {
		_, err = c.A.db.Exec(`INSERT INTO events (judul, mulai, selesai, kategori, keterangan, created_at) VALUES (?,?,?,?,?,?)`, judul, mulai, selesai, in.Kategori, str(in.Keterangan), nowStr())
	}
	if err != nil {
		return err
	}
	return c.OK()
}

// Libur nasional yang tanggalnya TETAP setiap tahun. Libur yang tanggalnya ditetapkan SKB tiap tahun (Idulfitri,
// Iduladha, Nyepi, Imlek, Waisak, dll.) sengaja tidak diisi otomatis agar tidak salah tanggal.
var liburTetap = []struct {
	bulan, hari int
	nama        string
}{
	{8, 17, "Hari Kemerdekaan Republik Indonesia"}, {12, 25, "Hari Raya Natal"},
	{1, 1, "Tahun Baru Masehi"}, {5, 1, "Hari Buruh Internasional"}, {6, 1, "Hari Lahir Pancasila"},
}

// apiNationalHolidays mengisi libur nasional tanggal tetap untuk satu tahun ajaran (Juli–Juni).
// Tahun ajaran diambil dari masukan "2026/2027" atau tahun ajaran aktif. Yang sudah ada tidak digandakan.
func apiNationalHolidays(c *Ctx) error {
	var in struct {
		TahunAjaran string `json:"tahun_ajaran"`
	}
	c.Bind(&in)
	ta := strings.TrimSpace(in.TahunAjaran)
	if ta == "" {
		c.A.db.QueryRow(`SELECT nama FROM academic_years WHERE aktif=1`).Scan(&ta)
	}
	if !yearRe.MatchString(ta) {
		return bad("Tahun ajaran harus seperti 2026/2027.")
	}
	y1, _ := strconv.Atoi(ta[:4])
	tx, err := c.A.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	added := 0
	var tanggal []string
	for _, h := range liburTetap {
		y := y1
		if h.bulan < 7 { // Januari–Juni jatuh di tahun kedua tahun ajaran
			y++
		}
		d := time.Date(y, time.Month(h.bulan), h.hari, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		var n int
		tx.QueryRow(`SELECT COUNT(*) FROM events WHERE mulai=? AND kategori='libur' AND judul=?`, d, h.nama).Scan(&n)
		if n > 0 {
			continue
		}
		if _, err := tx.Exec(`INSERT INTO events (judul, mulai, selesai, kategori, keterangan, created_at) VALUES (?,?,?,'libur','Libur nasional',?)`, h.nama, d, d, nowStr()); err != nil {
			return err
		}
		added++
		tanggal = append(tanggal, d)
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	c.Note = strconv.Itoa(added) + " libur nasional " + ta
	return c.JSON(map[string]any{"ditambahkan": added, "tanggal": tanggal, "tahun_ajaran": ta})
}
