package main

import "time"

// Nilai akhir mapel = 40 % rata-rata tugas + 40 % rata-rata kuis + 20 % persentase hadir, dari komponen yang ada.
// Dihitung di server (satu sumber untuk rekap guru, halaman nilai siswa, dan cetak) dengan jam server, sehingga
// jam perangkat yang salah tidak mengubah apakah tugas lewat tenggat dihitung 0.

type tugasNilai struct {
	nilaiMaks float64
	deadline  string   // RFC3339, kosong = tanpa tenggat
	nilai     *float64 // nil = belum dinilai
	kumpul    bool     // siswa sudah mengumpulkan
}

type skorMapel struct {
	Tugas *float64 `json:"tugas"`
	Kuis  *float64 `json:"kuis"`
	Hadir *float64 `json:"hadir"`
	Akhir *float64 `json:"akhir"`
}

func rata(v []float64) *float64 {
	if len(v) == 0 {
		return nil
	}
	s := 0.0
	for _, x := range v {
		s += x
	}
	r := s / float64(len(v))
	return &r
}

// nilaiAkhir: tugas yang dinilai → nilai/maks×100; tidak dikumpulkan & lewat tenggat → 0; lainnya tidak dihitung.
func nilaiAkhir(tugas []tugasNilai, skorKuis []float64, statusHadir []string, now time.Time) skorMapel {
	var t []float64
	for _, a := range tugas {
		maks := a.nilaiMaks
		if maks == 0 {
			maks = 100
		}
		switch {
		case a.nilai != nil:
			t = append(t, *a.nilai/maks*100)
		case !a.kumpul && a.deadline != "":
			if dl, err := time.Parse(time.RFC3339, a.deadline); err == nil && dl.Before(now) {
				t = append(t, 0)
			}
		}
	}
	out := skorMapel{Tugas: rata(t), Kuis: rata(skorKuis)}
	if len(statusHadir) > 0 {
		n := 0
		for _, s := range statusHadir {
			if s == "hadir" {
				n++
			}
		}
		h := float64(n) / float64(len(statusHadir)) * 100
		out.Hadir = &h
	}
	if out.Tugas == nil && out.Kuis == nil {
		return out // kehadiran saja belum cukup untuk nilai akhir
	}
	var jumlah, bobot float64
	for _, p := range []struct {
		v *float64
		b float64
	}{{out.Tugas, 40}, {out.Kuis, 40}, {out.Hadir, 20}} {
		if p.v != nil {
			jumlah += *p.v * p.b
			bobot += p.b
		}
	}
	akhir := jumlah / bobot
	out.Akhir = &akhir
	return out
}

// skorDariRekap menghitung nilai per siswa dari baris yang sudah dimuat untuk satu mapel.
// asg: tugas mapel; subs/att/abs: baris semua siswa (kolom siswa_id) atau satu siswa (tanpa siswa_id → semua milik dia).
func skorDariRekap(asg, subs, attempts, absen []map[string]any, siswa int64, now time.Time) skorMapel {
	milik := func(r map[string]any) bool {
		v, ok := r["siswa_id"]
		return !ok || v == siswa
	}
	sub := map[any]map[string]any{}
	for _, s := range subs {
		if milik(s) {
			sub[s["assignment_id"]] = s
		}
	}
	var tg []tugasNilai
	for _, a := range asg {
		x := tugasNilai{nilaiMaks: toF(a["nilai_maks"])}
		x.deadline, _ = a["deadline"].(string)
		if s, ok := sub[a["id"]]; ok {
			x.kumpul = true
			if s["nilai"] != nil {
				n := toF(s["nilai"])
				x.nilai = &n
			}
		}
		tg = append(tg, x)
	}
	var kuis []float64
	for _, t := range attempts {
		if milik(t) {
			kuis = append(kuis, toF(t["skor"]))
		}
	}
	var hadir []string
	for _, h := range absen {
		if milik(h) {
			st, _ := h["status"].(string)
			hadir = append(hadir, st)
		}
	}
	return nilaiAkhir(tg, kuis, hadir, now)
}

func toF(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int64:
		return float64(x)
	case int:
		return float64(x)
	}
	return 0
}
