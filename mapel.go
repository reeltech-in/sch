package main

import (
	"net/http"
	"strings"
)

// Pasangan kode–nama mata pelajaran. Satu sumber kebenaran untuk formulir mapel dan pemeriksaan server:
// nama yang sama selalu memakai kode yang sama di semua kelas, dan satu kode hanya untuk satu nama.

var mapelStd = [][2]string{
	{"PAI", "Pendidikan Agama Islam dan Budi Pekerti"}, {"PAK", "Pendidikan Agama Kristen dan Budi Pekerti"},
	{"PKAT", "Pendidikan Agama Katolik dan Budi Pekerti"}, {"PAH", "Pendidikan Agama Hindu dan Budi Pekerti"},
	{"PAB", "Pendidikan Agama Buddha dan Budi Pekerti"}, {"PAKH", "Pendidikan Agama Khonghucu dan Budi Pekerti"},
	{"PPKN", "Pendidikan Pancasila"}, {"BIN", "Bahasa Indonesia"}, {"MTK", "Matematika"},
	{"IPA", "Ilmu Pengetahuan Alam"}, {"IPS", "Ilmu Pengetahuan Sosial"}, {"IPAS", "Ilmu Pengetahuan Alam dan Sosial"},
	{"BIG", "Bahasa Inggris"}, {"PJOK", "Pendidikan Jasmani, Olahraga, dan Kesehatan"}, {"SBD", "Seni Budaya"},
	{"SR", "Seni Rupa"}, {"SM", "Seni Musik"}, {"INF", "Informatika"}, {"PKY", "Prakarya"}, {"BDR", "Bahasa Daerah"},
	{"BAR", "Bahasa Arab"}, {"FIS", "Fisika"}, {"KIM", "Kimia"}, {"BIO", "Biologi"}, {"EKO", "Ekonomi"},
	{"GEO", "Geografi"}, {"SOS", "Sosiologi"}, {"SEJ", "Sejarah"}, {"BK", "Bimbingan Konseling"},
}

func init() {
	moduleRoutes = append(moduleRoutes, func(a *App, m *http.ServeMux) {
		m.HandleFunc("GET /api/mapel/pilihan", a.h("admin", apiMapelPilihan))
	})
}

// apiMapelPilihan: pasangan yang sudah dipakai sekolah lebih dulu, lalu daftar umum yang belum ada.
func apiMapelPilihan(c *Ctx) error {
	rows, err := c.list(`SELECT kode, nama FROM courses GROUP BY nama COLLATE NOCASE ORDER BY nama`)
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	out := []map[string]string{}
	for _, r := range rows {
		k, _ := r["kode"].(string)
		n, _ := r["nama"].(string)
		seen[strings.ToLower(n)] = true
		out = append(out, map[string]string{"kode": k, "nama": n, "sumber": "sekolah"})
	}
	// saran mapel standar mengikuti jenjang sekolah (bila sudah diatur); mapel yang sudah dipakai sekolah tetap tampil
	j, diatur := jenjangDiatur(c.A.reader())
	for _, p := range mapelStd {
		if diatur && !mapelJenjang[j][p[0]] {
			continue
		}
		if !seen[strings.ToLower(p[1])] {
			out = append(out, map[string]string{"kode": p[0], "nama": p[1], "sumber": "umum"})
		}
	}
	return c.JSON(out)
}

// mapelKode menentukan kode yang sah untuk sebuah nama mapel (mengabaikan mapel `selfID` sendiri):
// kode yang sudah dipakai sekolah untuk nama itu, atau kode standar, atau kode masukan untuk mapel baru.
// Gagal bila kode itu sudah dipakai mapel dengan nama lain.
func (c *Ctx) mapelKode(nama, kodeIn string, selfID int64) (string, error) {
	kode := strings.ToUpper(strings.TrimSpace(kodeIn))
	var ada string
	if c.A.db.QueryRow(`SELECT kode FROM courses WHERE nama=? COLLATE NOCASE AND id<>? LIMIT 1`, nama, selfID).Scan(&ada) == nil {
		kode = strings.ToUpper(ada)
	} else {
		for _, p := range mapelStd {
			if strings.EqualFold(p[1], nama) {
				kode = p[0]
				break
			}
		}
	}
	if kode == "" {
		return "", bad("Kode mapel wajib diisi.")
	}
	var lain string
	if c.A.db.QueryRow(`SELECT nama FROM courses WHERE kode=? COLLATE NOCASE AND nama<>? COLLATE NOCASE AND id<>? LIMIT 1`, kode, nama, selfID).Scan(&lain) == nil {
		return "", bad("Kode " + kode + " sudah dipakai untuk mapel " + lain + ". Pilih nama mapel yang sesuai atau pakai kode lain.")
	}
	for _, p := range mapelStd { // kode standar milik mapel lain juga tidak boleh dipinjam
		if p[0] == kode && !strings.EqualFold(p[1], nama) {
			return "", bad("Kode " + kode + " adalah kode standar untuk " + p[1] + ". Pakai kode lain.")
		}
	}
	return kode, nil
}
