package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// Akun pengguna: daftar, simpan, hapus (dengan penjaga referensi), impor massal.

// ---------------------------------------------------------------------
// Pengguna
// ---------------------------------------------------------------------
func apiUsers(c *Ctx) error {
	qp := c.R.URL.Query()
	where, args := `1=1`, []any{}
	if r := qp.Get("role"); r != "" {
		where += ` AND u.role=?`
		args = append(args, r)
	}
	switch st := qp.Get("status"); st {
	case "aktif", "nonaktif":
		where += ` AND u.status=?`
		args = append(args, st)
	}
	switch k := qp.Get("kelas_id"); {
	case k == "none":
		where += ` AND u.kelas_id IS NULL`
	case k != "":
		where += ` AND u.kelas_id=?`
		args = append(args, k)
	}
	if s := strings.TrimSpace(qp.Get("q")); s != "" {
		like := "%" + s + "%"
		where += ` AND (u.nama_lengkap LIKE ? OR u.username LIKE ? OR u.email LIKE ? OR u.nip LIKE ?)`
		args = append(args, like, like, like, like)
	}
	q := `SELECT u.id, u.username, u.nama_lengkap, u.role, u.email, u.status, u.kelas_id, u.nip, u.akses, k.nama AS kelas_nama,
		(SELECT COUNT(*) FROM courses co WHERE co.guru_id=u.id) AS jumlah_mapel
		FROM users u LEFT JOIN classes k ON k.id=u.kelas_id WHERE ` + where + ` ORDER BY u.nama_lengkap, u.id`
	if qp.Get("limit") != "" { // halaman: LIMIT/OFFSET (data ribuan baris; indeks role+nama)
		var total int
		c.A.reader().QueryRow(`SELECT COUNT(*) FROM users u WHERE `+where, args...).Scan(&total)
		c.W.Header().Set("X-Total-Count", strconv.Itoa(total))
		limit, offset := pageParams(c, 50)
		q += ` LIMIT ? OFFSET ?`
		args = append(args, limit, offset)
	}
	rows, err := c.list(q, args...)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

type userIn struct {
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	NamaLengkap string   `json:"nama_lengkap"`
	Role        string   `json:"role"`
	Email       *string  `json:"email"`
	Status      string   `json:"status"`
	KelasID     *int64   `json:"kelas_id"`
	NIP         *string  `json:"nip"`
	Akses       []string `json:"akses"`
}

func apiSaveUser(c *Ctx) error {
	var in userIn
	if err := c.Bind(&in); err != nil {
		return err
	}
	id := c.ID("id")
	uname, err := required(in.Username, "Username")
	if err != nil {
		return err
	}
	if strings.ContainsAny(uname, " \t") {
		return bad("Username tidak boleh mengandung spasi.")
	}
	nama, err := required(in.NamaLengkap, "Nama lengkap")
	if err != nil {
		return err
	}
	if in.Role != "admin" && in.Role != "guru" && in.Role != "siswa" {
		return bad("Peran tidak valid.")
	}
	if in.Status != "nonaktif" {
		in.Status = "aktif"
	}
	if id == c.U.ID && (in.Role != "admin" || in.Status != "aktif") {
		return bad("Anda tidak bisa menonaktifkan atau mengubah peran akun Anda sendiri.")
	}
	var kelas any
	if in.Role == "siswa" && in.KelasID != nil && *in.KelasID > 0 {
		kelas = *in.KelasID
	}
	if id == 0 || in.Password != "" {
		if len(in.Password) < 6 {
			return bad("Password minimal 6 karakter.")
		}
	}
	nip, err := c.cekNIP(id, in.NIP)
	if err != nil {
		return err
	}
	in.NIP = nip
	var akses any
	if in.Role == "guru" {
		list := []string{}
		for _, m := range in.Akses {
			for _, k := range moduleKeys {
				if m == k && m != "lms" {
					list = append(list, m)
				}
			}
		}
		b, _ := json.Marshal(list)
		akses = string(b)
	}
	if id > 0 {
		err = c.updateUser(id, uname, nama, in.Role, str(in.Email), in.Status, kelas, str(in.NIP), akses, in.Password)
	} else {
		var h string
		if h, err = hashPassword(in.Password); err == nil {
			_, err = c.A.db.Exec(`INSERT INTO users (username, password_hash, nama_lengkap, role, email, status, kelas_id, nip, akses, ganti_pw, created_at)
				VALUES (?,?,?,?,?,?,?,?,?,1,?)`, uname, h, nama, in.Role, str(in.Email), in.Status, kelas, str(in.NIP), akses, nowStr())
		}
	}
	if err != nil {
		return err
	}
	return c.OK()
}

// updateUser: ubah data akun; bila password direset atau akun dinonaktifkan, semua sesinya diputus di transaksi
// yang sama, jadi tidak mungkin password sudah berganti tetapi orang lain masih login dengan sesi lama.
func (c *Ctx) updateUser(id int64, uname, nama, role string, email any, status string, kelas any, nip any, akses any, password string) error {
	var h string
	if password != "" {
		var err error
		if h, err = hashPassword(password); err != nil { // bcrypt di luar transaksi: tidak menahan koneksi tulis
			return err
		}
	}
	tx, err := c.A.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE users SET username=?, nama_lengkap=?, role=?, email=?, status=?, kelas_id=?, nip=?, akses=? WHERE id=?`,
		uname, nama, role, email, status, kelas, nip, akses, id); err != nil {
		return err
	}
	if h != "" {
		// password direset admin: pemilik akun wajib menggantinya saat login berikutnya
		if _, err := tx.Exec(`UPDATE users SET password_hash=?, ganti_pw=? WHERE id=?`, h, b2i(id != c.U.ID), id); err != nil {
			return err
		}
	}
	if h != "" || status == "nonaktif" {
		if _, err := tx.Exec(`DELETE FROM sessions WHERE user_id=?`, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// countRefs menjumlahkan baris riwayat yang terkait (dipakai untuk mencegah penghapusan yang merusak arsip).
func (c *Ctx) countRefs(queries []string, id int64) int {
	total := 0
	for _, q := range queries {
		var n int
		if c.A.db.QueryRow(q, id).Scan(&n) == nil {
			total += n
		}
	}
	return total
}

// Akun yang sudah punya riwayat (nilai, absensi, ujian, pinjaman, suara) tidak boleh dihapus:
// menghapusnya ikut menghapus arsip. Sarankan menonaktifkan.
func apiDeleteUser(c *Ctx) error {
	id := c.ID("id")
	if id == c.U.ID {
		return bad("Anda tidak bisa menghapus akun Anda sendiri.")
	}
	n := c.countRefs([]string{
		`SELECT COUNT(*) FROM submissions WHERE siswa_id=?`, `SELECT COUNT(*) FROM quiz_attempts WHERE siswa_id=?`,
		`SELECT COUNT(*) FROM attendance WHERE siswa_id=?`, `SELECT COUNT(*) FROM daily_attendance WHERE siswa_id=?`,
		`SELECT COUNT(*) FROM exam_attempts WHERE siswa_id=?`, `SELECT COUNT(*) FROM loans WHERE user_id=?`,
		`SELECT COUNT(*) FROM graduation WHERE siswa_id=?`, `SELECT COUNT(*) FROM election_voters WHERE user_id=?`,
		`SELECT COUNT(*) FROM sales WHERE kasir_id=?`,
		// guru: mapel yang diampu & ujian buatannya akan kehilangan pemilik
		`SELECT COUNT(*) FROM courses WHERE guru_id=?`, `SELECT COUNT(*) FROM exams WHERE created_by=?`,
		`SELECT COUNT(*) FROM teaching_journal WHERE guru_id=?`, `SELECT COUNT(*) FROM student_notes WHERE siswa_id=?`,
	}, id)
	if n > 0 {
		return bad(fmt.Sprintf("Akun ini sudah memiliki %d catatan riwayat (nilai, absensi, ujian, pinjaman, dll). Ubah statusnya menjadi Nonaktif agar arsip tetap utuh.", n))
	}
	return apiDelete("users")(c)
}

func apiDeleteClass(c *Ctx) error {
	id := c.ID("id")
	var siswa, mapel int
	c.A.db.QueryRow(`SELECT COUNT(*) FROM users WHERE kelas_id=?`, id).Scan(&siswa)
	c.A.db.QueryRow(`SELECT COUNT(*) FROM courses WHERE kelas_id=?`, id).Scan(&mapel)
	if siswa > 0 || mapel > 0 {
		return bad(fmt.Sprintf("Kelas masih memiliki %d siswa dan %d mapel. Pindahkan siswa dan hapus mapelnya terlebih dahulu.", siswa, mapel))
	}
	return apiDelete("classes")(c)
}

func apiDeleteCourse(c *Ctx) error {
	id := c.ID("id")
	n := c.countRefs([]string{
		`SELECT COUNT(*) FROM submissions s JOIN assignments a ON a.id=s.assignment_id WHERE a.course_id=?`,
		`SELECT COUNT(*) FROM quiz_attempts t JOIN quizzes q ON q.id=t.quiz_id WHERE q.course_id=?`,
		`SELECT COUNT(*) FROM attendance WHERE course_id=?`,
	}, id)
	if n > 0 {
		return bad(fmt.Sprintf("Mapel ini sudah memiliki %d nilai/absensi siswa. Menghapusnya akan menghilangkan arsip; pindahkan guru atau biarkan mapel tetap ada.", n))
	}
	return apiDelete("courses")(c)
}

// Impor siswa/guru massal dari Excel/CSV (sudah diurai di browser).
func apiImportUsers(c *Ctx) error {
	var in struct {
		Role string `json:"role"`
		Rows []struct {
			Nama     string `json:"nama"`
			Username string `json:"username"`
			Password string `json:"password"`
			Kelas    string `json:"kelas"`
			Email    string `json:"email"`
			NIP      string `json:"nip"`
		} `json:"rows"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	if in.Role != "siswa" && in.Role != "guru" {
		return bad("Peran tidak valid.")
	}
	if len(in.Rows) == 0 || len(in.Rows) > 3000 {
		return bad("Jumlah baris harus 1 sampai 3000.")
	}
	type row struct {
		line, nama, uname, pw, kelas, email, nip, hash string
		acak                                           bool
	}
	var rows []row
	skipped := []string{}
	seen := map[string]bool{}
	for i, r := range in.Rows {
		line := "Baris " + strconv.Itoa(i+2) + ": "
		x := row{line: line, nama: strings.TrimSpace(r.Nama), uname: strings.TrimSpace(r.Username), pw: strings.TrimSpace(r.Password),
			kelas: strings.TrimSpace(r.Kelas), email: strings.TrimSpace(r.Email), nip: strings.TrimSpace(r.NIP)}
		if x.nama == "" || x.uname == "" || strings.ContainsAny(x.uname, " \t") {
			skipped = append(skipped, line+"nama/username kosong atau mengandung spasi")
			continue
		}
		if seen[strings.ToLower(x.uname)] {
			skipped = append(skipped, line+"username "+x.uname+" ganda di file")
			continue
		}
		seen[strings.ToLower(x.uname)] = true
		if x.pw == "" { // tanpa password: buat acak (bukan NIS, yang mudah ditebak)
			x.pw, x.acak = randomPassword(8), true
		}
		if len(x.pw) < 6 || len(x.pw) > 72 {
			skipped = append(skipped, line+"password harus 6–72 karakter ("+x.uname+")")
			continue
		}
		rows = append(rows, x)
	}
	// Hash password secara paralel DI LUAR transaksi: bcrypt lambat (±80 ms) dan koneksi tulis
	// hanya satu, jadi jangan menahan seluruh aplikasi (login, simpan jawaban CBT) selama impor.
	work := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < max(2, runtime.NumCPU()); w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range work {
				rows[i].hash, _ = hashPassword(rows[i].pw)
			}
		}()
	}
	for i := range rows {
		work <- i
	}
	close(work)
	wg.Wait()

	tx, err := c.A.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	kelasIDs := map[string]int64{}
	type akun struct {
		Nama     string `json:"nama"`
		Username string `json:"username"`
		Password string `json:"password"`
		Kelas    string `json:"kelas"`
	}
	akunBaru := []akun{}
	for _, x := range rows {
		if x.hash == "" {
			skipped = append(skipped, x.line+"gagal membuat password")
			continue
		}
		var exists int
		tx.QueryRow(`SELECT COUNT(*) FROM users WHERE username=? COLLATE NOCASE`, x.uname).Scan(&exists)
		if exists > 0 {
			skipped = append(skipped, x.line+"username "+x.uname+" sudah ada")
			continue
		}
		var kelas any
		if in.Role == "siswa" && x.kelas != "" {
			key := strings.ToLower(x.kelas)
			kid, ok := kelasIDs[key]
			if !ok {
				ty := activeYear(tx)
				if tx.QueryRow(`SELECT id FROM classes WHERE nama=? COLLATE NOCASE AND tahun_id=?`, x.kelas, ty).Scan(&kid) != nil {
					res, err := tx.Exec(`INSERT INTO classes (tahun_id, nama, created_at) VALUES (?,?,?)`, ty, x.kelas, nowStr())
					if err != nil {
						return err
					}
					kid, _ = res.LastInsertId()
				}
				kelasIDs[key] = kid
			}
			kelas = kid
		}
		if _, err := tx.Exec(`INSERT INTO users (username, password_hash, nama_lengkap, role, email, kelas_id, nip, ganti_pw, created_at) VALUES (?,?,?,?,?,?,?,1,?)`,
			x.uname, x.hash, x.nama, in.Role, str(&x.email), kelas, str(&x.nip), nowStr()); err != nil {
			skipped = append(skipped, x.line+err.Error())
			continue
		}
		akunBaru = append(akunBaru, akun{x.nama, x.uname, x.pw, x.kelas})
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.JSON(map[string]any{"added": len(akunBaru), "skipped": skipped, "akun": akunBaru})
}
