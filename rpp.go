package main

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Perencanaan pembelajaran (RPP) untuk Asisten AI, mengikuti Kurikulum Merdeka dengan Pembelajaran Mendalam:
//   - Permendikdasmen No. 13 Tahun 2025 (8 dimensi profil lulusan)
//   - Keputusan Kepala BSKAP No. 046/H/KR/2025 (Capaian Pembelajaran per fase)
//   - prinsip pembelajaran mendalam: berkesadaran, bermakna, menggembirakan;
//     pengalaman belajar: memahami – mengaplikasi – merefleksi.
// Guru tidak diwajibkan satu format; tersedia tiga format umum. Komponen inti selalu ada: tujuan, langkah, asesmen.

// rppFormats: kode → nama format yang tampil di dokumen
var rppFormats = map[string]string{
	"rpm":   "Rencana Pembelajaran Mendalam (RPM)",
	"rpp1":  "RPP Satu Lembar",
	"modul": "Modul Ajar",
}

var faseRe = regexp.MustCompile(`(?i)\bfase\s+([a-f])\b`)
var angkaRe = regexp.MustCompile(`\d+`)

// rppFase: fase Kurikulum Merdeka dari isian "Fase D (Kelas 7–9)" atau nama kelas "7A"/"VIII B"/"Kelas 10".
func rppFase(kelas string) (fase string, tingkat int) {
	if m := faseRe.FindStringSubmatch(kelas); m != nil {
		return "Fase " + strings.ToUpper(m[1]), 0
	}
	t := tingkatDariNama(kelas)
	if t == 0 {
		if n, err := strconv.Atoi(angkaRe.FindString(kelas)); err == nil {
			t = n
		}
	}
	switch {
	case t >= 1 && t <= 2:
		return "Fase A", t
	case t >= 3 && t <= 4:
		return "Fase B", t
	case t >= 5 && t <= 6:
		return "Fase C", t
	case t >= 7 && t <= 9:
		return "Fase D", t
	case t == 10:
		return "Fase E", t
	case t == 11 || t == 12:
		return "Fase F", t
	}
	return "", 0
}

// rppMenitJP: lama satu jam pelajaran sesuai jenjang (SD 35, SMP 40, SMA/SMK 45 menit) bila guru tidak mengisi.
func rppMenitJP(in map[string]string, fase string, tingkat int) int {
	if n, err := strconv.Atoi(strings.TrimSpace(in["menit_jp"])); err == nil && n >= 20 && n <= 60 {
		return n
	}
	switch {
	case tingkat >= 1 && tingkat <= 6, fase == "Fase A", fase == "Fase B", fase == "Fase C":
		return 35
	case tingkat >= 7 && tingkat <= 9, fase == "Fase D":
		return 40
	}
	return 45
}

func rppNum(in map[string]string, k string, def, lo, hi int) int {
	n, err := strconv.Atoi(strings.TrimSpace(in[k]))
	if err != nil || n < lo || n > hi {
		return def
	}
	return n
}

// rppWaktu: (JP per pertemuan, menit per JP, jumlah pertemuan, total menit per pertemuan)
func rppWaktu(in map[string]string) (jp, menit, pertemuan, total int, fase string) {
	fase, tingkat := rppFase(in["kelas"])
	jp = rppNum(in, "jp", 2, 1, 6)
	menit = rppMenitJP(in, fase, tingkat)
	pertemuan = rppNum(in, "pertemuan", 1, 1, 20)
	return jp, menit, pertemuan, jp * menit, fase
}

func rppJudul(in map[string]string) string {
	nama := map[string]string{"rpm": "RPM", "rpp1": "RPP", "modul": "Modul Ajar"}[in["format"]]
	if nama == "" {
		nama = "RPM"
	}
	return nama + " " + v(in, "mapel", "") + " – " + v(in, "materi", "")
}

// Rincian waktu wajib berbentuk satu baris baku per pertemuan; tampilan memeriksa jumlahnya (lihat rppCekWaktu di web).
const rppBarisWaktu = "Rincian waktu: awal A menit + inti B menit + penutup C menit = T menit"

func rppPrompt(in map[string]string, sekolah string) string {
	format := in["format"]
	if _, ok := rppFormats[format]; !ok {
		format = "rpm"
	}
	jp, menit, pertemuan, total, fase := rppWaktu(in)
	if fase == "" {
		fase = "tentukan dari kelas"
	}
	cp := v(in, "cp", "")
	cpAturan := `Capaian Pembelajaran (CP): guru menempelkan CP resmi di bawah ini — kutip apa adanya, jangan diubah.`
	if cp == "" {
		cpAturan = `Capaian Pembelajaran (CP): guru tidak menempelkan CP. Tulis ringkasan CP ` + fase + ` untuk mapel ini sesuai Keputusan Kepala BSKAP No. 046/H/KR/2025, lalu beri catatan miring tepat di bawahnya: "*Ringkasan CP – cocokkan dengan dokumen CP resmi (Kepka BSKAP 046/H/KR/2025) sebelum digunakan.*" Jangan mengarang nomor/kode CP.`
		cp = "-"
	}
	data := fmt.Sprintf(`Data:
- Satuan pendidikan: %s
- Jenjang: %s
- Mata pelajaran: %s
- Fase / kelas: %s / %s
- Materi pokok: %s
- Alokasi waktu: %d pertemuan × %d JP × %d menit = %d menit setiap pertemuan
- Model/metode yang diinginkan: %s
- Dimensi profil lulusan yang ditekankan (1–3): %s
- Tujuan pembelajaran (jika diisi guru): %s
- Capaian pembelajaran (jika ditempel guru): %s
- Catatan guru (karakteristik siswa, sarana): %s

%s`,
		sekolah, v(in, "jenjang", "-"), v(in, "mapel", "-"), fase, v(in, "kelas", "-"), v(in, "materi", "-"), pertemuan, jp, menit, total,
		v(in, "model", "pilih yang paling sesuai materi"), v(in, "profil", "pilih 1–3 yang paling relevan"), v(in, "tujuan", "rumuskan sendiri dari CP"),
		cp, v(in, "catatan", "-"), cpAturan)

	waktu := fmt.Sprintf(`Aturan waktu (WAJIB): setiap pertemuan berdurasi tepat %d menit. Tulis estimasi menit pada tiap kegiatan, dan akhiri rincian setiap pertemuan dengan SATU baris persis berformat:
%s
(ganti A, B, C dengan menit kegiatan awal, inti, penutup; A + B + C harus tepat %d, dan T = %d).`, total, rppBarisWaktu, total, total)

	umum := `Ketentuan umum:
- Kurikulum Merdeka dengan Pembelajaran Mendalam (Permendikdasmen No. 13 Tahun 2025): berkesadaran (mindful), bermakna (meaningful), menggembirakan (joyful); pengalaman belajar memahami – mengaplikasi – merefleksi.
- Dimensi profil lulusan hanya dari 8 dimensi resmi: Keimanan dan Ketakwaan terhadap Tuhan YME, Kewargaan, Penalaran Kritis, Kreativitas, Kolaborasi, Kemandirian, Kesehatan, Komunikasi. Pilih 1–3 saja dan tunjukkan di kegiatan mana dimensi itu dilatih.
- Tujuan pembelajaran memakai kata kerja operasional yang dapat diukur dan selaras dengan asesmen.
- Bahasa Indonesia baku, ringkas, siap dipakai guru. Format Markdown: heading, daftar, tabel. Jangan menulis kop sekolah atau kolom tanda tangan (ditambahkan aplikasi saat dicetak).`

	var struktur string
	switch format {
	case "rpp1":
		struktur = `Susun RPP SATU LEMBAR (ringkas, efisien, sesuai Surat Edaran Mendikbud No. 14 Tahun 2019: tiga komponen inti), diperkaya prinsip Pembelajaran Mendalam. Maksimal ±1 halaman per pertemuan.

Struktur wajib (heading Markdown):
1. Identitas (satu tabel ringkas: satuan pendidikan, mapel, fase/kelas, materi, alokasi waktu, dimensi profil lulusan)
2. Tujuan Pembelajaran (2–4 butir)
3. Langkah Pembelajaran per pertemuan: Awal – Inti (memahami, mengaplikasi, merefleksi) – Penutup, singkat, dengan menit
4. Asesmen (awal, formatif, sumatif – masing-masing satu kalimat teknik + instrumen)`
	case "modul":
		struktur = `Susun MODUL AJAR Kurikulum Merdeka yang lengkap, diperkaya prinsip Pembelajaran Mendalam.

Struktur wajib (heading Markdown):
A. Informasi Umum: identitas modul, kompetensi awal, dimensi profil lulusan, sarana dan prasarana, target peserta didik, model pembelajaran
B. Komponen Inti: capaian pembelajaran, tujuan pembelajaran, pemahaman bermakna, pertanyaan pemantik, kegiatan pembelajaran per pertemuan (pendahuluan – inti: memahami, mengaplikasi, merefleksi – penutup, dengan menit), asesmen (diagnostik, formatif, sumatif), pengayaan dan remedial, refleksi guru dan peserta didik
C. Lampiran: LKPD ringkas, bahan bacaan guru & peserta didik, glosarium, daftar pustaka`
	default:
		struktur = `Susun RENCANA PEMBELAJARAN MENDALAM (RPM).

Struktur wajib (heading Markdown):
1. Identitas (tabel: satuan pendidikan, mapel, fase/kelas, materi, alokasi waktu)
2. Identifikasi: kesiapan peserta didik (asesmen awal), karakteristik materi, dimensi profil lulusan yang disasar
3. Desain Pembelajaran: capaian pembelajaran, lintas disiplin ilmu, tujuan pembelajaran, topik pembelajaran, praktik pedagogis (model/metode), kemitraan pembelajaran, lingkungan pembelajaran, pemanfaatan digital
4. Pengalaman Belajar per pertemuan: Awal (berkesadaran) – Inti: Memahami, Mengaplikasi, Merefleksi (tandai prinsip bermakna/menggembirakan) – Penutup
5. Asesmen: awal (diagnostik), proses (formatif), akhir (sumatif) – teknik, instrumen, dan rubrik singkat dalam tabel
6. Lampiran singkat: LKPD ringkas / bahan bacaan / glosarium`
	}
	return strings.Join([]string{struktur, data, waktu, umum}, "\n\n")
}

func init() {
	// jenjang sekolah disimpan sekali di Pengaturan → Identitas sekolah
	moduleMigrations = append(moduleMigrations, func(db *sql.DB) error { return addColumn(db, "settings", "jenjang", "TEXT") })
}

// jenjangDiatur: jenjang yang benar-benar dipilih admin (bukan tebakan). Pembatasan tingkat kelas & daftar mapel
// hanya berlaku bila sudah diatur, agar sekolah yang belum mengisi tidak terhalang tebakan yang keliru.
func jenjangDiatur(db *sql.DB) (string, bool) {
	var j sql.NullString
	db.QueryRow(`SELECT jenjang FROM settings WHERE id=1`).Scan(&j)
	_, ok := jenjangNama[j.String]
	return j.String, ok
}

// mapel standar (kode) yang lazim per jenjang menurut struktur Kurikulum Merdeka
var mapelJenjang = func() map[string]map[string]bool {
	agama := "PAI PAK PKAT PAH PAB PAKH PPKN BIN MTK BIG PJOK BDR BAR"
	m := map[string]string{
		"SD":  agama + " IPAS SR SM",
		"SMP": agama + " IPA IPS INF PKY SR SM BK",
		"SMA": agama + " IPA IPS INF PKY SR SM SEJ FIS KIM BIO EKO GEO SOS BK",
		"SMK": agama + " IPAS INF SBD SEJ BK",
	}
	out := map[string]map[string]bool{}
	for j, list := range m {
		out[j] = map[string]bool{}
		for _, k := range strings.Fields(list) {
			out[j][k] = true
		}
	}
	return out
}()

// jenjangSekolah: jenjang di pengaturan; bila belum diisi, ditebak dari warna tema (merah SD, abu SMA, lainnya SMP).
func jenjangSekolah(db *sql.DB) string {
	var j sql.NullString
	db.QueryRow(`SELECT jenjang FROM settings WHERE id=1`).Scan(&j)
	if _, ok := jenjangNama[j.String]; ok {
		return j.String
	}
	switch getConfig(db, "tampilan")["warna"] {
	case "merah":
		return "SD"
	case "abu":
		return "SMA"
	}
	return "SMP"
}

// Jenjang pendidikan untuk perangkat ajar: membatasi kelas & kedalaman materi agar keluaran AI tidak keluar jenjang.
var jenjangNama = map[string]string{"SD": "SD/MI", "SMP": "SMP/MTs", "SMA": "SMA/MA", "SMK": "SMK/MAK"}
var jenjangKelas = map[string][2]int{"SD": {1, 6}, "SMP": {7, 9}, "SMA": {10, 12}, "SMK": {10, 12}}

// aiTplJenjang: templat yang memakai isian jenjang/kelas
var aiTplJenjang = map[string]bool{"rpm": true, "lkpd": true, "soal": true, "prota": true, "kktp": true}

// aiJenjang memeriksa jenjang & kelas yang dikirim dan menghasilkan baris konteks untuk prompt.
// Kiriman lama (tanpa jenjang) tetap diterima tanpa konteks tambahan.
func aiJenjang(in map[string]string) (string, error) {
	j := strings.ToUpper(strings.TrimSpace(in["jenjang"]))
	if j == "" {
		return "", nil
	}
	nama, ok := jenjangNama[j]
	if !ok {
		return "", bad("Jenjang tidak dikenal. Pilih SD, SMP, SMA, atau SMK.")
	}
	in["jenjang"] = nama
	kelas := 0
	if m := angkaRe.FindString(in["kelas"]); m != "" {
		kelas, _ = strconv.Atoi(m)
	}
	r := jenjangKelas[j]
	if kelas < r[0] || kelas > r[1] {
		return "", bad(fmt.Sprintf("Kelas tidak sesuai jenjang %s (kelas %d–%d).", nama, r[0], r[1]))
	}
	fase, _ := rppFase(fmt.Sprintf("%d", kelas))
	return fmt.Sprintf(`Jenjang & kelas (WAJIB dipatuhi): %s, kelas %d, %s. Sesuaikan kedalaman materi, istilah, contoh, panjang teks, dan tingkat kesulitan dengan peserta didik %s kelas %d; jangan memakai materi, istilah, atau capaian dari jenjang atau fase lain.`,
		nama, kelas, fase, nama, kelas), nil
}
