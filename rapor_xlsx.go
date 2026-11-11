package main

import (
	"math"
	"net/url"
	"strings"
)

// apiRaporXlsx: ekspor nilai rapor ke Excel untuk dipindahkan ke e-Rapor resmi (Kemendikdasmen/Dapodik).
//
//	?course_id=  satu mapel (guru pengampu/staf rapor): satu lembar
//	?kelas_id=   satu kelas (wali kelas/staf rapor): lembar Leger + satu lembar per mapel
//
// Siswa diurutkan menurut nama (sama dengan urutan di Dapodik), NISN/NIS disimpan sebagai teks.
// Nilai yang belum disimpan guru dibiarkan kosong (ditandai "Belum diisi"), bukan diisi nilai otomatis.
func apiRaporXlsx(c *Ctx) error {
	p, err := c.periodeQuery()
	if err != nil {
		return err
	}
	cid, kelas := c.QInt("course_id"), c.QInt("kelas_id")
	var mapel []map[string]any
	switch {
	case cid > 0:
		if !c.canManage(cid) && !c.staffRapor() {
			return forbidden()
		}
		k, tahunKelas, _ := c.kelasMapel(cid)
		if k == 0 {
			return notFound()
		}
		if tahunKelas != 0 && tahunKelas != p.tahun {
			return bad("Mapel ini milik tahun ajaran lain. Pilih tahun ajaran yang sesuai.")
		}
		kelas = k
		mapel, err = c.list(`SELECT id, nama FROM courses WHERE id=?`, cid)
	case kelas > 0:
		if !c.bolehWali(kelas) {
			return forbidden()
		}
		var tahunKelas int64
		if c.A.reader().QueryRow(`SELECT COALESCE(tahun_id,0) FROM classes WHERE id=?`, kelas).Scan(&tahunKelas) != nil {
			return notFound()
		}
		if tahunKelas != 0 && tahunKelas != p.tahun {
			return bad("Kelas ini milik tahun ajaran lain. Pilih tahun ajaran yang sesuai.")
		}
		mapel, err = c.list(`SELECT id, nama FROM courses WHERE kelas_id=? ORDER BY nama`, kelas)
	default:
		return bad("Pilih kelas atau mata pelajaran yang akan diekspor.")
	}
	if err != nil {
		return err
	}
	var kelasNama string
	c.A.reader().QueryRow(`SELECT nama FROM classes WHERE id=?`, kelas).Scan(&kelasNama)
	siswa, err := c.siswaKelas(kelas)
	if err != nil {
		return err
	}
	q, args := `SELECT course_id, siswa_id, nilai, tinggi, perlu FROM rapor_nilai WHERE tahun_id=? AND semester=? AND course_id IN (SELECT id FROM courses WHERE kelas_id=?)`, []any{p.tahun, p.smt, kelas}
	if cid > 0 {
		q, args = `SELECT course_id, siswa_id, nilai, tinggi, perlu FROM rapor_nilai WHERE tahun_id=? AND semester=? AND course_id=?`, []any{p.tahun, p.smt, cid}
	}
	isi, err := c.list(q, args...)
	if err != nil {
		return err
	}
	type kunci struct{ course, siswa int64 }
	nilai := map[kunci]map[string]any{}
	for _, r := range isi {
		co, _ := r["course_id"].(int64)
		si, _ := r["siswa_id"].(int64)
		nilai[kunci{co, si}] = r
	}
	str := func(m map[string]any, k string) string {
		if m == nil {
			return ""
		}
		v, _ := m[k].(string)
		return strings.TrimSpace(v)
	}
	identitas := func(i int, s map[string]any) []any {
		return []any{i + 1, str(s, "nisn"), str(s, "username"), str(s, "nama_lengkap")}
	}

	var sheets []xlsxSheet
	if cid == 0 { // Leger: semua nilai akhir satu kelas + ketidakhadiran (seperti di rapor)
		absen, err := c.absenSemester(p, kelas)
		if err != nil {
			return err
		}
		ubah, err := c.list(`SELECT siswa_id, sakit, izin, alfa FROM rapor_siswa WHERE tahun_id=? AND semester=? AND siswa_id IN
			(SELECT u.id FROM users u WHERE u.role='siswa' AND `+memberOf("?")+`)`, p.tahun, p.smt, kelas, kelas)
		if err != nil {
			return err
		}
		byS := map[int64]map[string]any{}
		for _, r := range ubah {
			id, _ := r["siswa_id"].(int64)
			byS[id] = r
		}
		judul := []any{"No", "NISN", "NIS", "Nama Siswa"}
		lebar := []float64{5, 14, 12, 30}
		for _, m := range mapel {
			judul = append(judul, str(m, "nama"))
			lebar = append(lebar, 12)
		}
		judul = append(judul, "Rata-rata", "Sakit", "Izin", "Tanpa Keterangan")
		lebar = append(lebar, 10, 8, 8, 12)
		rows := [][]any{judul}
		for i, s := range siswa {
			id, _ := s["id"].(int64)
			row := identitas(i, s)
			var jml float64
			var n int
			for _, m := range mapel {
				co, _ := m["id"].(int64)
				if v := nilai[kunci{co, id}]["nilai"]; v != nil {
					f := toF(v)
					row = append(row, f)
					jml += f
					n++
				} else {
					row = append(row, nil)
				}
			}
			if n > 0 {
				row = append(row, math.Round(jml/float64(n)*10)/10)
			} else {
				row = append(row, nil)
			}
			a, r := absen[id], byS[id]
			for j, k := range []string{"sakit", "izin", "alfa"} {
				if v, ok := r[k].(int64); ok {
					row = append(row, v)
				} else {
					row = append(row, a[j])
				}
			}
			rows = append(rows, row)
		}
		sheets = append(sheets, xlsxSheet{Nama: "Leger", Lebar: lebar, Baris: rows})
	}
	for _, m := range mapel { // satu lembar per mapel: kolom yang dibutuhkan template e-Rapor
		co, _ := m["id"].(int64)
		rows := [][]any{{"No", "NISN", "NIS", "Nama Siswa", "Nilai Akhir", "Deskripsi Capaian Kompetensi", "Capaian Tertinggi", "Perlu Ditingkatkan", "Keterangan"}}
		for i, s := range siswa {
			id, _ := s["id"].(int64)
			r := nilai[kunci{co, id}]
			row := identitas(i, s)
			var v any
			if r != nil && r["nilai"] != nil {
				v = toF(r["nilai"])
			}
			tinggi, perlu := str(r, "tinggi"), str(r, "perlu")
			ket := ""
			if v == nil {
				ket = "Belum diisi"
			} else if tinggi == "" && perlu == "" {
				ket = "Deskripsi belum diisi"
			}
			rows = append(rows, append(row, v, strings.TrimSpace(tinggi+" "+perlu), tinggi, perlu, ket))
		}
		sheets = append(sheets, xlsxSheet{Nama: str(m, "nama"), Lebar: []float64{5, 14, 12, 30, 11, 60, 40, 40, 20}, Baris: rows})
	}

	b, err := buildXlsx(sheets)
	if err != nil {
		return err
	}
	name := "Nilai Rapor " + kelasNama
	if cid > 0 && len(mapel) > 0 {
		name += " " + str(mapel[0], "nama")
	}
	name = namaBerkas(name + " " + strings.ReplaceAll(p.nama, "/", "-") + " " + p.label())
	h := c.W.Header()
	h.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	h.Set("Content-Disposition", `attachment; filename="nilai-rapor.xlsx"; filename*=UTF-8''`+url.PathEscape(name+".xlsx"))
	h.Set("Cache-Control", "no-store")
	_, err = c.W.Write(b)
	return err
}
