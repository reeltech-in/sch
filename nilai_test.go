package main

import (
	"fmt"
	"math"
	"testing"
	"time"
)

func f(v float64) *float64 { return &v }

func sama(a, b *float64) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return math.Abs(*a-*b) < 1e-9
}

// Rumus nilai akhir (dulu hanya di browser, kini satu sumber di server).
func TestRumusNilaiAkhir(t *testing.T) {
	now := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	lewat := "2026-11-01T00:00:00Z"
	nanti := "2026-12-01T00:00:00Z"
	cases := []struct {
		nama  string
		tugas []tugasNilai
		kuis  []float64
		hadir []string
		want  skorMapel
	}{
		{"belum ada apa-apa", nil, nil, nil, skorMapel{}},
		{"hanya kehadiran: belum ada nilai akhir", nil, nil, []string{"hadir", "alfa"}, skorMapel{Hadir: f(50)}},
		{"lengkap 40/40/20",
			[]tugasNilai{{nilaiMaks: 100, nilai: f(80), kumpul: true}},
			[]float64{90},
			[]string{"hadir", "hadir", "sakit", "hadir"},
			skorMapel{Tugas: f(80), Kuis: f(90), Hadir: f(75), Akhir: f((80*40 + 90*40 + 75*20) / 100.0)}},
		{"nilai maks 50 diskalakan ke 100",
			[]tugasNilai{{nilaiMaks: 50, nilai: f(45), kumpul: true}}, nil, nil,
			skorMapel{Tugas: f(90), Akhir: f(90)}},
		{"nilai maks 0 dianggap 100",
			[]tugasNilai{{nilaiMaks: 0, nilai: f(70), kumpul: true}}, nil, nil,
			skorMapel{Tugas: f(70), Akhir: f(70)}},
		{"tidak dikumpulkan & lewat tenggat = 0; belum tenggat & belum dinilai tidak dihitung",
			[]tugasNilai{{nilaiMaks: 100, nilai: f(100), kumpul: true}, {nilaiMaks: 100, deadline: lewat}, {nilaiMaks: 100, deadline: nanti}, {nilaiMaks: 100, deadline: lewat, kumpul: true}},
			nil, nil,
			skorMapel{Tugas: f(50), Akhir: f(50)}},
		{"tanpa tugas: kuis 80% bobot + hadir 20%",
			nil, []float64{60, 80}, []string{"hadir"},
			skorMapel{Kuis: f(70), Hadir: f(100), Akhir: f((70*40 + 100*20) / 60.0)}},
	}
	for _, c := range cases {
		got := nilaiAkhir(c.tugas, c.kuis, c.hadir, now)
		if !sama(got.Tugas, c.want.Tugas) || !sama(got.Kuis, c.want.Kuis) || !sama(got.Hadir, c.want.Hadir) || !sama(got.Akhir, c.want.Akhir) {
			t.Errorf("%s: dapat %+v", c.nama, fmtSkor(got))
		}
	}
}

func fmtSkor(s skorMapel) string {
	p := func(v *float64) string {
		if v == nil {
			return "nil"
		}
		return fmt.Sprintf("%.4f", *v)
	}
	return fmt.Sprintf("tugas=%s kuis=%s hadir=%s akhir=%s", p(s.Tugas), p(s.Kuis), p(s.Hadir), p(s.Akhir))
}

// Rekap guru & halaman nilai siswa memakai nilai yang sama dari server.
func TestNilaiAkhirDiRekapDanSiswaSama(t *testing.T) {
	e := newEnv(t)
	g := e.login("guru001", "guru123")
	s := e.login("siswa001", "siswa123")
	var rk struct{ Nilai map[string]skorMapel }
	if code := g.json("GET", "/api/rekap?course_id=1", nil, &rk); code != 200 {
		t.Fatalf("rekap: %d", code)
	}
	var me struct{ Nilai map[string]skorMapel }
	if code := s.json("GET", "/api/rekap/me", nil, &me); code != 200 {
		t.Fatalf("rekap/me: %d", code)
	}
	id := fmt.Sprint(e.id(`SELECT id FROM users WHERE username='siswa001'`))
	a, b := rk.Nilai[id], me.Nilai["1"]
	if a.Akhir == nil || !sama(a.Akhir, b.Akhir) || !sama(a.Tugas, b.Tugas) || !sama(a.Kuis, b.Kuis) || !sama(a.Hadir, b.Hadir) {
		t.Fatalf("nilai rekap guru (%s) ≠ halaman siswa (%s)", fmtSkor(a), fmtSkor(b))
	}
}
