package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Diisi saat build: go build -ldflags "-X main.version=1.4.0 -X main.buildTime=202611091300"
var (
	version   = "dev"
	buildTime = ""
)

const sessionCookie = "sch_session"
const sessionTTL = 7 * 24 * time.Hour

func main() {
	addr := flag.String("addr", ":3220", "alamat server, contoh :3220 atau :80")
	dbPath := flag.String("db", "data/sch.db", "lokasi file database SQLite")
	resetAdmin := flag.Bool("reset-admin", false, "buat ulang password admin lalu keluar")
	demo := flag.Bool("demo", false, "isi data contoh untuk mencoba (hanya jika database masih kosong)")
	tlsAddr := flag.String("https", ":3221", "alamat HTTPS lokal (untuk kamera HP: scan QR & selfie). Kosongkan untuk mematikan")
	flag.BoolVar(&trustProxy, "proxy", false, "aktifkan jika online lewat Cloudflare Tunnel / reverse proxy (baca IP asli & HTTPS dari header)")
	maxGB := flag.Float64("upload-max-gb", 5, "batas total ukuran file upload (GB)")
	showVer := flag.Bool("version", false, "tampilkan versi lalu keluar")
	backupTo := flag.String("backup", "", "buat salinan database yang konsisten ke file ini lalu keluar (dipakai skrip deploy)")
	restoreFrom := flag.String("restore", "", "pulihkan dari file backup (.db atau .zip backup lengkap) lalu keluar; server harus dimatikan dulu")
	flag.StringVar(&backupDirFlag, "backup-dir", "", "folder backup harian otomatis, sebaiknya di disk lain (USB/NAS). Kosong = <folder data>/backups")
	proxyIP := flag.String("proxy-ip", "", "IP/CIDR proxy yang dipercaya, pisahkan koma (mis. IP VM cloudflared). Kosong = hanya proxy di komputer ini (127.0.0.1)")
	flag.StringVar(&proxyHeader, "proxy-header", "cf", "sumber IP asli di mode proxy: cf = Cloudflare Tunnel (CF-Connecting-IP), xff = nginx/proxy lain (elemen terakhir X-Forwarded-For)")
	flag.Parse()
	if proxyHeader != "cf" && proxyHeader != "xff" {
		log.Fatal("-proxy-header harus cf atau xff")
	}
	for _, p := range strings.Split(*proxyIP, ",") {
		if p = strings.TrimSpace(p); p == "" {
			continue
		}
		if !strings.Contains(p, "/") {
			if strings.Contains(p, ":") {
				p += "/128"
			} else {
				p += "/32"
			}
		}
		_, n, err := net.ParseCIDR(p)
		if err != nil {
			log.Fatal("-proxy-ip tidak valid: ", p)
		}
		trustedProxies = append(trustedProxies, n)
	}
	if len(trustedProxies) > 0 {
		trustProxy = true
	}
	if trustProxy && proxyHeader == "cf" {
		log.Println("mode proxy: IP asli dibaca dari header Cloudflare (CF-Connecting-IP). Di belakang nginx/proxy lain " +
			"(bukan Cloudflare), jalankan dengan -proxy-header xff agar header itu tidak bisa dipalsukan pengunjung.")
	}
	if len(trustedProxies) == 0 && trustProxy {
		// -proxy tanpa -proxy-ip: hanya percaya proxy di mesin ini (cloudflared/nginx lokal). Sebelumnya semua
		// alamat dipercaya, sehingga pengguna LAN bisa memalsukan CF-Connecting-IP dan lolos pembatas login.
		for _, c := range []string{"127.0.0.0/8", "::1/128"} {
			_, n, _ := net.ParseCIDR(c)
			trustedProxies = append(trustedProxies, n)
		}
		log.Println("mode proxy: hanya proxy di komputer ini yang dipercaya. Proxy di mesin lain? tambahkan -proxy-ip <IP>")
	}
	uploadMaxBytes = int64(*maxGB * (1 << 30))
	if *showVer {
		fmt.Println("SCH", versionString())
		return
	}
	if *restoreFrom != "" {
		ringkas, err := pulihkanBackup(*dbPath, *restoreFrom)
		if err != nil {
			log.Fatal("pemulihan dibatalkan: ", err)
		}
		fmt.Println(ringkas)
		fmt.Println("Selesai. Jalankan server lagi; database diperbarui otomatis bila backup berasal dari versi lama.")
		return
	}
	if *backupTo != "" {
		rdb, err := openReadDB(*dbPath)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := rdb.Exec(`VACUUM INTO ?`, *backupTo); err != nil {
			log.Fatal("backup gagal: ", err)
		}
		os.Chmod(*backupTo, 0o600)
		fmt.Println("Backup tersimpan:", *backupTo)
		return
	}
	log.Println("SCH", versionString(), "mulai")

	db, err := openDB(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := ensureAdmin(db, *dbPath, *resetAdmin); err != nil {
		log.Fatal(err)
	}
	if *resetAdmin {
		return
	}
	if *demo {
		if err := seedDemo(db); err != nil {
			log.Fatal("data contoh: ", err)
		}
	}

	if z, _ := getConfig(db, "waktu")["zona"].(string); z != "" {
		applyTimezone(z)
	}
	rdb, err := openReadDB(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	app := &App{db: db, rdb: rdb, dbPath: *dbPath, limit: newLoginLimiter(), quit: make(chan struct{})}
	app.background("pemeliharaan", app.cleanupLoop)
	app.background("whatsapp", app.waLoop)

	mux := http.NewServeMux()
	app.routes(mux)
	mux.Handle("GET /", newStaticHandler(app))

	srv := &http.Server{
		Addr:              *addr,
		Handler:           securityHeaders(mux),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
	}
	printAddresses(*addr, "http")
	var tsrv *http.Server
	if *tlsAddr != "" {
		cert, err := selfSignedCert(filepath.Dir(*dbPath))
		if err != nil {
			log.Println("HTTPS tidak aktif:", err)
		} else {
			tsrv = &http.Server{Addr: *tlsAddr, Handler: srv.Handler, TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12},
				ReadHeaderTimeout: 10 * time.Second, ReadTimeout: 60 * time.Second, WriteTimeout: 120 * time.Second}
			fmt.Println("HTTPS (untuk kamera HP: scan QR & absen selfie). Browser akan memberi peringatan, pilih Lanjutkan/Advanced:")
			printAddresses(*tlsAddr, "https")
			go func() { log.Println("HTTPS berhenti:", tsrv.ListenAndServeTLS("", "")) }()
		}
	}
	fmt.Println("Tekan Ctrl+C untuk berhenti.")

	// Berhenti dengan rapi (Ctrl+C / systemctl restart / deploy): tolak permintaan baru, tunggu yang
	// sedang berjalan (mis. kirim jawaban CBT) sampai 25 detik, lalu tutup database dengan bersih.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	log.Println("berhenti: menunggu permintaan yang sedang berjalan selesai…")
	sctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	srv.Shutdown(sctx)
	if tsrv != nil {
		tsrv.Shutdown(sctx)
	}
	close(app.quit) // hentikan pekerja latar (WA, pemeliharaan) sebelum database ditutup
	done := make(chan struct{})
	go func() { app.bg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(20 * time.Second):
		log.Println("pekerja latar belum selesai; tetap menutup database")
	}
	db.Exec(`PRAGMA wal_checkpoint(TRUNCATE)`)
	rdb.Close()
	db.Close()
	log.Println("SCH berhenti dengan rapi.")
}

func versionString() string {
	if buildTime != "" {
		return "v" + version + " (build " + buildTime + ")"
	}
	return "v" + version
}

func apiVersion(c *Ctx) error {
	return c.JSON(map[string]string{"version": version, "build": buildTime})
}

func printAddresses(addr, scheme string) {
	_, port, _ := net.SplitHostPort(addr)
	if port == "" {
		port = "80"
	}
	if scheme == "http" {
		fmt.Println("SCH berjalan. Buka di browser:")
	}
	fmt.Printf("   %s://localhost:%s   (di komputer server ini)\n", scheme, port)
	ifaces, _ := net.InterfaceAddrs()
	for _, a := range ifaces {
		if ipn, ok := a.(*net.IPNet); ok && !ipn.IP.IsLoopback() && ipn.IP.To4() != nil {
			fmt.Printf("   %s://%s:%s   (dari HP/laptop di jaringan sekolah)\n", scheme, ipn.IP, port)
		}
	}
}
