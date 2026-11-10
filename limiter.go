package main

import (
	"log"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// Pembatas percobaan login (per alamat IP, per pasangan IP+username, per username).

// ---------------------------------------------------------------------
// Pembatas percobaan login (per alamat IP)
// ---------------------------------------------------------------------
type loginLimiter struct {
	mu    sync.Mutex
	fails map[string][]time.Time
	// penghitung harian (direset tiap ganti tanggal) untuk cek publik kelulusan/PPDB
	day   string
	daily map[string]int
}

func newLoginLimiter() *loginLimiter {
	return &loginLimiter{fails: map[string][]time.Time{}, daily: map[string]int{}}
}

func (l *loginLimiter) dailyLocked() map[string]int {
	if d := time.Now().Format("2006-01-02"); d != l.day || l.daily == nil || len(l.daily) > 200000 {
		l.day, l.daily = d, map[string]int{}
	}
	return l.daily
}

// dayCount: berapa kali key gagal hari ini.
func (l *loginLimiter) dayCount(key string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.dailyLocked()[key]
}

func (l *loginLimiter) dayAdd(keys ...string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	m := l.dailyLocked()
	for _, k := range keys {
		m[k]++
	}
}

// Cek publik dengan rahasia yang bisa ditebak (NISN + tanggal lahir, nomor PPDB + PIN).
// Tujuannya dua: (1) menebak satu target butuh sangat banyak perangkat, (2) orang iseng TIDAK bisa
// mengunci siswa asli dengan sengaja salah berkali-kali. Aturan:
//   - satu perangkat: 10 salah per 10 menit, 40 salah per hari (semua target);
//   - satu target: setelah 20 salah dalam sehari, hanya perangkat yang hari ini BELUM pernah salah
//     yang masih boleh mencoba (siswa asli di HP-nya sendiri tetap bisa; penebak harus terus ganti perangkat).
const (
	guessIPWindow = 10
	guessIPDay    = 40
	guessTarget   = 20
)

func (c *Ctx) guessAllowed(scope, target string) error {
	ip := limitKey(clientIP(c.R))
	l := c.A.limit
	if l.blockedN(scope+"ip:"+ip, guessIPWindow) {
		return apiErr{429, "Terlalu banyak percobaan dari perangkat ini. Coba lagi 10 menit lagi."}
	}
	if l.dayCount(scope+"ipd:"+ip) >= guessIPDay {
		return apiErr{429, "Terlalu banyak percobaan salah dari perangkat ini hari ini. Coba lagi besok atau hubungi sekolah."}
	}
	if l.dayCount(scope+"t:"+target) >= guessTarget && l.dayCount(scope+"ipd:"+ip) > 0 {
		return apiErr{429, "Data ini terlalu sering dicoba dengan isian salah hari ini. Periksa kembali isian Anda, coba dari HP lain, atau hubungi sekolah."}
	}
	return nil
}

func (c *Ctx) guessFailed(scope, target string) {
	ip := limitKey(clientIP(c.R))
	c.A.limit.fail(scope + "ip:" + ip)
	c.A.limit.dayAdd(scope+"ipd:"+ip, scope+"t:"+target)
}

// blockedN: true bila key gagal >= n kali dalam 10 menit terakhir (dipakai per-username & formulir publik).
func (l *loginLimiter) blockedN(key string, n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	cut := time.Now().Add(-10 * time.Minute)
	keep := l.fails[key][:0]
	for _, t := range l.fails[key] {
		if t.After(cut) {
			keep = append(keep, t)
		}
	}
	l.fails[key] = keep
	return len(keep) >= n
}

func isHTTPS(r *http.Request) bool {
	return r.TLS != nil || (viaProxy(r) && strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https"))
}

func (l *loginLimiter) fail(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.fails) > 50000 {
		l.pruneLocked()
	}
	if _, ada := l.fails[ip]; !ada && len(l.fails) > 50000 {
		// masih penuh (serangan dengan banyak kunci acak): buang 10% entri yang paling lama tidak aktif.
		// Kunci yang sedang aktif (mis. akun yang ditebak) tetap tercatat, jadi perlindungan tidak mati.
		type kv struct {
			k string
			t time.Time
		}
		all := make([]kv, 0, len(l.fails))
		for k, ts := range l.fails {
			all = append(all, kv{k, ts[len(ts)-1]})
		}
		sort.Slice(all, func(i, j int) bool { return all[i].t.Before(all[j].t) })
		for _, x := range all[:len(all)/10] {
			delete(l.fails, x.k)
		}
	}
	l.fails[ip] = append(l.fails[ip], time.Now())
}

// prune membuang catatan yang sudah lewat 10 menit (bukan menghapus semuanya).
func (l *loginLimiter) prune() {
	l.mu.Lock()
	l.pruneLocked()
	l.mu.Unlock()
}

func (l *loginLimiter) pruneLocked() {
	cut := time.Now().Add(-10 * time.Minute)
	for k, ts := range l.fails {
		if len(ts) == 0 || ts[len(ts)-1].Before(cut) {
			delete(l.fails, k)
		}
	}
}

func (l *loginLimiter) reset(ip string) {
	l.mu.Lock()
	delete(l.fails, ip)
	l.mu.Unlock()
}

// trustProxy: aktif jika server berada di belakang Cloudflare Tunnel / reverse proxy.
// Tanpa ini, semua pengunjung terlihat dari 127.0.0.1 dan pembatas login memblokir semua orang.
var trustProxy bool

var trustedProxies []*net.IPNet

// viaProxy: header proxy hanya dipercaya bila permintaan datang dari proxy yang diizinkan.
// Mencegah pengguna LAN memalsukan header CF-Connecting-IP untuk menghindari pembatas login.
func viaProxy(r *http.Request) bool {
	if !trustProxy {
		return false
	}
	if len(trustedProxies) == 0 {
		return true
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	for _, n := range trustedProxies {
		if ip != nil && n.Contains(ip) {
			return true
		}
	}
	return false
}

// proxyWarnAt: kapan terakhir mencatat peringatan header proxy dari alamat tak dipercaya (maks 1x per 10 menit)
var proxyWarnAt struct {
	sync.Mutex
	t time.Time
}

func clientIP(r *http.Request) string {
	if trustProxy && !viaProxy(r) && (r.Header.Get("CF-Connecting-IP") != "" || r.Header.Get("X-Forwarded-For") != "") {
		proxyWarnAt.Lock()
		if time.Since(proxyWarnAt.t) > 10*time.Minute {
			proxyWarnAt.t = time.Now()
			log.Printf("PERINGATAN: header proxy datang dari %s yang tidak dipercaya; IP asli pengunjung diabaikan. "+
				"Bila ini proxy/cloudflared sekolah, jalankan dengan -proxy-ip %s", r.RemoteAddr, strings.Split(r.RemoteAddr, ":")[0])
		}
		proxyWarnAt.Unlock()
	}
	if viaProxy(r) {
		if ip := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); ip != "" {
			return ip
		}
		// Elemen PERTAMA X-Forwarded-For bisa diisi pengunjung sendiri; proxy menambahkan IP
		// yang ia lihat di UJUNG daftar, jadi ambil elemen terakhir (dok. Cloudflare HTTP headers).
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			return strings.TrimSpace(parts[len(parts)-1])
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// limitKey: IPv6 dikelompokkan per /64 agar penyerang tidak bisa berganti alamat tanpa batas.
func limitKey(ip string) string {
	p := net.ParseIP(ip)
	if p != nil && p.To4() == nil {
		return p.Mask(net.CIDRMask(64, 128)).String() + "/64"
	}
	return ip
}
