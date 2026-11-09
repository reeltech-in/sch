package main

import (
	"regexp"
	"strconv"
	"strings"
)

// Penjaga keutuhan data: mencegah data ganda atau pasangan yang tidak cocok (mis. kelas "8A" bertingkat 7).
// Semua pemeriksaan di server, jadi berlaku walau data dikirim tanpa lewat formulir.

var digitsRe = regexp.MustCompile(`\D`)

func onlyDigits(s string) string { return digitsRe.ReplaceAllString(s, "") }

// exists: nilai kolom pertama dari baris pertama yang cocok (mis. nama pemilik data yang bentrok).
func (c *Ctx) exists(q string, args ...any) (string, bool) {
	var v string
	err := c.A.db.QueryRow(q, args...).Scan(&v)
	return v, err == nil
}

// canon menyamakan ejaan dengan nilai yang sudah ada (tanpa beda huruf besar/kecil & spasi ganda),
// sehingga "elektronik" mengikuti "Elektronik" yang sudah dipakai. Tabel & kolom hanya dari kode, bukan masukan.
func (c *Ctx) canon(table, col string, p *string) *string {
	if p == nil {
		return nil
	}
	v := strings.Join(strings.Fields(*p), " ")
	if v == "" {
		return p
	}
	if ada, ok := c.exists(`SELECT `+col+` FROM `+table+` WHERE `+col+`=? COLLATE NOCASE LIMIT 1`, v); ok {
		v = ada
	}
	return &v
}

// tingkatDariNama: "7A" → 7, "X IPA 1" → 10, "XII-2" → 12; 0 bila nama tidak diawali angka/romawi.
var romawiTingkat = map[string]int{"I": 1, "II": 2, "III": 3, "IV": 4, "V": 5, "VI": 6, "VII": 7, "VIII": 8, "IX": 9, "X": 10, "XI": 11, "XII": 12}
var tingkatRe = regexp.MustCompile(`^\s*(\d{1,2}|XII|XI|X|IX|VIII|VII|VI|V|IV|III|II|I)(?:[^A-Za-z0-9]|[A-Z]\b|$|[A-Z]\d|\s)`)

func tingkatDariNama(nama string) int {
	m := tingkatRe.FindStringSubmatch(strings.ToUpper(nama))
	if m == nil {
		return 0
	}
	if n, err := strconv.Atoi(m[1]); err == nil {
		return n
	}
	return romawiTingkat[m[1]]
}

// cekKelas: tingkat harus sesuai nama kelas; satu guru hanya wali satu kelas per tahun ajaran.
func (c *Ctx) cekKelas(id int64, nama string, tingkat *string, wali any) error {
	if t := tingkatDariNama(nama); t > 0 {
		if tingkat == nil || strings.TrimSpace(*tingkat) == "" {
			s := strconv.Itoa(t)
			*tingkat = s
		} else if strings.TrimSpace(*tingkat) != strconv.Itoa(t) {
			return bad("Nama kelas " + nama + " menunjukkan tingkat " + strconv.Itoa(t) + ", tetapi tingkat yang dipilih " + strings.TrimSpace(*tingkat) + ".")
		}
	}
	if wali != nil {
		tahun := activeYear(c.A.db)
		if id > 0 {
			c.A.db.QueryRow(`SELECT tahun_id FROM classes WHERE id=?`, id).Scan(&tahun)
		}
		if k, ok := c.exists(`SELECT nama FROM classes WHERE wali_id=? AND tahun_id=? AND id<>?`, wali, tahun, id); ok {
			return bad("Guru ini sudah menjadi wali kelas " + k + " pada tahun ajaran yang sama.")
		}
	}
	return nil
}

// cekIdentitasSiswa: NISN 10 digit & NIK 16 digit, masing-masing unik di antara siswa.
func (c *Ctx) cekIdentitasSiswa(siswaID int64, in map[string]*string) error {
	for _, f := range []struct {
		key, label string
		n          int
	}{{"nisn", "NISN", 10}, {"nik", "NIK", 16}} {
		p := in[f.key]
		if p == nil || strings.TrimSpace(*p) == "" {
			continue
		}
		v := onlyDigits(*p)
		if len(v) != f.n {
			return bad(f.label + " harus " + strconv.Itoa(f.n) + " digit angka.")
		}
		in[f.key] = &v
		if nama, ok := c.exists(`SELECT u.nama_lengkap FROM student_profiles p JOIN users u ON u.id=p.siswa_id WHERE p.`+f.key+`=? AND p.siswa_id<>?`, v, siswaID); ok {
			return bad(f.label + " " + v + " sudah dipakai siswa " + nama + ".")
		}
	}
	return nil
}

// cekNIP: hanya angka (NIP 18 digit / NUPTK 16 digit), unik di antara pengguna.
func (c *Ctx) cekNIP(userID int64, p *string) (*string, error) {
	if p == nil || strings.TrimSpace(*p) == "" {
		return nil, nil
	}
	v := onlyDigits(*p)
	if len(v) != 18 && len(v) != 16 {
		return nil, bad("NIP harus 18 digit (atau NUPTK 16 digit) angka.")
	}
	if nama, ok := c.exists(`SELECT nama_lengkap FROM users WHERE nip=? AND id<>?`, v, userID); ok {
		return nil, bad("NIP " + v + " sudah dipakai " + nama + ".")
	}
	return &v, nil
}

// cekRentang: tanggal/jam selesai tidak boleh sebelum mulai (string ISO/JJ:MM bisa dibandingkan langsung).
func cekRentang(mulai, selesai any, label string) error {
	m, ok1 := mulai.(string)
	s, ok2 := selesai.(string)
	if ok1 && ok2 && m != "" && s != "" && s < m {
		return bad(label + " selesai tidak boleh sebelum " + strings.ToLower(label) + " mulai.")
	}
	return nil
}
