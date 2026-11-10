package main

// util.go: fungsi bantu yang dipakai lintas modul (sebelumnya tersebar di file modul tertentu).

import (
	"encoding/json"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// ---- tanggal, angka, dan daftar kelas

// localDate: tanggal hari ini (zona waktu server), format 2006-01-02
func localDate() string { return time.Now().Format("2006-01-02") }

func monthRange(s string) (string, string, error) {
	t, err := time.Parse("2006-01", strings.TrimSpace(s))
	if err != nil {
		return "", "", bad("Bulan tidak valid.")
	}
	return t.Format("2006-01-02"), t.AddDate(0, 1, -1).Format("2006-01-02"), nil
}

func itoa(n int) string { return strconv.Itoa(n) }

// persen: nilai 0–100 dengan dua desimal (dipakai CBT, kuis, dan LJK)
func persen(benar, total float64) float64 {
	if total <= 0 {
		return 0
	}
	return math.Round(benar/total*10000) / 100
}

func kelasIn(list string, kelas int64) bool {
	var ids []int64
	json.Unmarshal([]byte(list), &ids)
	for _, k := range ids {
		if k == kelas {
			return true
		}
	}
	return false
}

// fmtTanggal: 2026-11-09 → 9/11/2026
func fmtTanggal(s string) string {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.Format("2/1/2006")
	}
	return s
}

// ---- halaman publik (tanpa login): batas kiriman per perangkat & badan permintaan besar

// publicGuard membatasi formulir publik: per-IP dan honeypot anti-bot.
// publicBlocked: hanya memeriksa (tidak menghitung). Dipanggil SEBELUM membaca isi kiriman yang besar, agar
// perangkat yang sudah diblokir tidak menghabiskan memori; kiriman dihitung lewat publicGuard setelah terbaca.
func (c *Ctx) publicBlocked(key string, max int) error {
	if c.A.limit.blockedN(key+limitKey(clientIP(c.R)), max) {
		return apiErr{429, "Terlalu banyak pengiriman dari perangkat ini. Coba lagi 10 menit lagi."}
	}
	return nil
}

func (c *Ctx) publicGuard(key string, max int, honeypot string) error {
	ip := limitKey(clientIP(c.R))
	if c.A.limit.blockedN(key+ip, max) {
		return apiErr{429, "Terlalu banyak pengiriman dari perangkat ini. Coba lagi 10 menit lagi."}
	}
	c.A.limit.fail(key + ip)
	if strings.TrimSpace(honeypot) != "" {
		return bad("Pengiriman ditolak.")
	}
	return nil
}

// bindLarge seperti Bind tetapi dengan batas ukuran sendiri (untuk formulir berisi foto).
func (c *Ctx) bindLarge(v any, limit int64) error {
	b, err := io.ReadAll(http.MaxBytesReader(c.W, c.R.Body, limit))
	if err != nil {
		return bad("Data terlalu besar. Kecilkan ukuran file.")
	}
	if json.Unmarshal(b, v) != nil {
		return bad("Data yang dikirim tidak valid.")
	}
	return nil
}
