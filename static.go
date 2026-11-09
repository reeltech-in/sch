package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

// Penyajian file tampilan tertanam (versi hash, gzip, data awal di index.html) dan header keamanan.

//go:embed web
var webFS embed.FS

// staticHandler menyajikan file tampilan dengan versi otomatis (hash isi file).
// index.html merujuk app.js?v=<hash>, sehingga setelah update browser dan cache Cloudflare
// langsung memakai file baru (tanpa ini Cloudflare menyimpan JS/CSS lama sampai 4 jam).
type staticFile struct {
	data  []byte
	gz    []byte // versi gzip (dibuat sekali saat server menyala)
	etag  string
	ctype string
}

// newStaticHandler melayani file tampilan (tertanam). index.html dibuat per permintaan: berisi data awal
// (pengaturan sekolah, akun yang sedang masuk, daftar file berversi) agar aplikasi tidak perlu menunggu
// satu perjalanan lagi ke server sebelum tampil, dan modul dimuat sesuai kebutuhan.
func newStaticHandler(a *App) http.Handler {
	sub, _ := fs.Sub(webFS, "web")
	files := map[string]*staticFile{}
	fs.WalkDir(sub, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		b, _ := fs.ReadFile(sub, p)
		h := sha256.Sum256(b)
		ct := mime.TypeByExtension(filepath.Ext(p))
		if ct == "" {
			ct = http.DetectContentType(b)
		}
		files["/"+p] = &staticFile{data: b, etag: `"` + hex.EncodeToString(h[:8]) + `"`, ctype: ct}
		return nil
	})
	// tulis ulang index.html: tambahkan ?v=<hash> pada setiap file lokal yang dirujuk
	if idx := files["/index.html"]; idx != nil {
		html := string(idx.data)
		for name, f := range files {
			if name == "/index.html" {
				continue
			}
			ref := strings.TrimPrefix(name, "/")
			html = strings.ReplaceAll(html, `"`+ref+`"`, `"`+ref+"?v="+strings.Trim(f.etag, `"`)+`"`)
		}
		h := sha256.Sum256([]byte(html))
		idx.data, idx.etag = []byte(html), `"`+hex.EncodeToString(h[:8])+`"`
	}
	versioned := map[string]string{}
	for name, f := range files {
		if strings.HasSuffix(name, ".js") {
			versioned[strings.TrimPrefix(name, "/")] = strings.TrimPrefix(name, "/") + "?v=" + strings.Trim(f.etag, `"`)
		}
	}
	for _, f := range files {
		if strings.HasPrefix(f.ctype, "text/") || strings.Contains(f.ctype, "javascript") || strings.Contains(f.ctype, "svg") {
			var buf bytes.Buffer
			zw, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
			zw.Write(f.data)
			zw.Close()
			if buf.Len() < len(f.data) {
				f.gz = buf.Bytes()
			}
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if p == "/" {
			p = "/index.html"
		}
		f := files[p]
		if f == nil {
			http.NotFound(w, r)
			return
		}
		if p == "/index.html" && a != nil {
			boot := map[string]any{"files": versioned}
			if s, err := a.settingsData(); err == nil {
				boot["settings"] = s
			}
			if u := a.currentUser(r); u != nil {
				boot["me"] = u
			}
			b, _ := json.Marshal(boot) // json.Marshal meng-escape < > & sehingga aman di dalam <script>
			html := strings.Replace(string(f.data), "<!--BOOT-->", `<script type="application/json" id="boot">`+string(b)+`</script>`, 1)
			w.Header().Set("Content-Type", f.ctype)
			w.Header().Set("Cache-Control", "no-store, private") // berisi data akun: jangan disimpan cache/proxy
			w.Write([]byte(html))
			return
		}
		w.Header().Set("Content-Type", f.ctype)
		w.Header().Set("ETag", f.etag)
		if p == "/index.html" || r.URL.Query().Get("v") == "" {
			w.Header().Set("Cache-Control", "no-cache") // selalu periksa versi terbaru
		} else {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable") // URL berversi tidak pernah berubah
		}
		w.Header().Set("Vary", "Accept-Encoding")
		if r.Header.Get("If-None-Match") == f.etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		if f.gz != nil && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			w.Write(f.gz)
			return
		}
		w.Write(f.data)
	})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-SCH-Version", version)
		h.Set("X-Frame-Options", "SAMEORIGIN")
		h.Set("Referrer-Policy", "same-origin")
		h.Set("Permissions-Policy", "camera=(self), geolocation=(self), microphone=()")
		if !strings.HasPrefix(r.URL.Path, "/files/") {
			h.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob: https:; media-src 'self' blob:; connect-src 'self'; frame-src https://www.youtube-nocookie.com; object-src 'none'; base-uri 'none'; form-action 'self'; frame-ancestors 'self'")
		}
		if viaProxy(r) && isHTTPS(r) {
			h.Set("Strict-Transport-Security", "max-age=15552000")
		}
		next.ServeHTTP(w, r)
	})
}
