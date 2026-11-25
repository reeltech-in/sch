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

// guessAllowed memeriksa batas cek publik (kelulusan, status PPDB) dan memesan satu slot pada kunci DATA yang dicek
// (NISN / nomor pendaftaran) dalam satu kunci mutex: tebakan serentak ke satu data tidak bisa melewati batas
// guessIPWindow. Kunci IP TIDAK dipesan di depan, karena ratusan siswa satu sekolah berbagi satu IP (NAT) dan
// mengecek bersamaan saat pengumuman dibuka (regresi v1.28.1: hanya 73 dari 200 cek yang benar berhasil).
// Setelah diperiksa, panggil guessOK (isian benar) atau guessFailed (isian salah).
func (c *Ctx) guessAllowed(scope, target string) error {
	ip := limitKey(clientIP(c.R))
	l := c.A.limit
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.blockedLocked(scope+"ip:"+ip, guessIPWindow) {
		return apiErr{429, "Terlalu banyak percobaan dari perangkat ini. Coba lagi 10 menit lagi."}
	}
	d := l.dailyLocked()
	if d[scope+"ipd:"+ip] >= guessIPDay {
		return apiErr{429, "Terlalu banyak percobaan salah dari perangkat ini hari ini. Coba lagi besok atau hubungi sekolah."}
	}
	if (d[scope+"t:"+target] >= guessTarget && d[scope+"ipd:"+ip] > 0) || l.blockedLocked(scope+"tw:"+target, guessIPWindow) {
		return apiErr{429, "Data ini terlalu sering dicoba dengan isian salah hari ini. Periksa kembali isian Anda, coba dari HP lain, atau hubungi sekolah."}
	}
	l.failLocked(scope + "tw:" + target)
	return nil
}

// guessOK: isian benar, slot yang dipesan guessAllowed dikembalikan.
func (c *Ctx) guessOK(scope, target string) {
	c.A.limit.unfail(scope + "tw:" + target)
}

// guessFailed: isian salah, dicatat pada kunci IP (10 menit & harian) dan hitungan harian data itu.
// Slot pada kunci data tetap terpakai sebagai percobaan gagal.
func (c *Ctx) guessFailed(scope, target string) {
	ip := limitKey(clientIP(c.R))
	l := c.A.limit
	l.mu.Lock()
	defer l.mu.Unlock()
	l.failLocked(scope + "ip:" + ip)
	d := l.dailyLocked()
	d[scope+"ipd:"+ip]++
	d[scope+"t:"+target]++
}

// blockedN: true bila key gagal >= n kali dalam 10 menit terakhir (dipakai per-username & formulir publik).
func (l *loginLimiter) blockedN(key string, n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.blockedLocked(key, n)
}

// tryAdd: periksa batas lalu catat satu percobaan dalam satu kunci (permintaan serentak tidak bisa lolos bersama).
func (l *loginLimiter) tryAdd(key string, n int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.blockedLocked(key, n) {
		return false
	}
	l.failLocked(key)
	return true
}

func (l *loginLimiter) blockedLocked(key string, n int) bool {
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
	l.failLocked(ip)
}

func (l *loginLimiter) failLocked(ip string) {
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

// loginReserve memeriksa batas login lalu langsung mencatat percobaan ini pada kunci PASANGAN (IP+username) dalam
// satu kunci mutex, sehingga tebakan serentak ke satu akun tidak bisa melewati batas 5.
// Kunci IP (60) dan username (15) TIDAK dipesan di depan: ratusan siswa satu sekolah berbagi satu IP (NAT) dan
// login bersamaan saat ujian dimulai; memesan kunci IP membuat login yang benar ikut ditolak setelah 60 orang
// (regresi v1.28.1). Keduanya dicatat setelah password terbukti salah, lewat loginFailed.
// ok=false: diblokir. slow=true: username ini sedang sering dicoba (diperlambat, tidak dikunci).
func (l *loginLimiter) loginReserve(pair, ip, user string) (ok, slow bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.blockedLocked(pair, 5) || l.blockedLocked(ip, 60) {
		return false, false
	}
	slow = l.blockedLocked(user, 15)
	l.failLocked(pair)
	return true, slow
}

// loginFailed mencatat password yang terbukti salah pada kunci IP dan username (kunci pasangan sudah dipesan).
func (l *loginLimiter) loginFailed(ip, user string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.failLocked(ip)
	l.failLocked(user)
}

// unfail menghapus satu catatan percobaan terakhir pada tiap kunci (percobaan yang ternyata berhasil).
func (l *loginLimiter) unfail(keys ...string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, k := range keys {
		if ts := l.fails[k]; len(ts) > 1 {
			l.fails[k] = ts[:len(ts)-1]
		} else {
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

// proxyHeader: "cf" (Cloudflare Tunnel, bawaan) atau "xff" (nginx/proxy lain; CF-Connecting-IP diabaikan).
var proxyHeader = "cf"

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
		// CF-Connecting-IP hanya dipercaya di mode Cloudflare: Cloudflare selalu menimpanya, sedangkan nginx/proxy
		// lain meneruskan header kiriman pengunjung apa adanya (pengunjung bisa berganti "IP" tiap permintaan).
		if proxyHeader == "cf" {
			if ip := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); net.ParseIP(ip) != nil {
				return ip
			}
		}
		// Elemen PERTAMA X-Forwarded-For bisa diisi pengunjung sendiri; proxy menambahkan IP
		// yang ia lihat di UJUNG daftar, jadi ambil elemen terakhir (dok. Cloudflare HTTP headers).
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			if ip := strings.TrimSpace(parts[len(parts)-1]); net.ParseIP(ip) != nil {
				return ip
			}
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
