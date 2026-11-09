package main

import (
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
}

func newLoginLimiter() *loginLimiter { return &loginLimiter{fails: map[string][]time.Time{}} }

func (l *loginLimiter) blocked(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	cut := time.Now().Add(-10 * time.Minute)
	keep := l.fails[ip][:0]
	for _, t := range l.fails[ip] {
		if t.After(cut) {
			keep = append(keep, t)
		}
	}
	l.fails[ip] = keep
	return len(keep) >= 10
}

// blockedN: seperti blocked tetapi dengan batas sendiri (dipakai per-username & formulir publik).
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

func clientIP(r *http.Request) string {
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
