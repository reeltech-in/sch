package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Notifikasi WhatsApp lewat gateway pihak ketiga (Fonnte, atau gateway lain dengan mode "kustom", mis. Wablas).
//
//   - Token gateway hanya tersimpan di server (kv "wa_token"); tidak pernah dikirim ke browser.
//   - Pesan masuk antrean (wa_outbox) lalu dikirim satu per satu dengan jeda, agar nomor WA sekolah tidak dianggap spam.
//   - Kunci unik per kejadian (mis. "masuk:12:2026-11-09") mencegah pesan ganda.
//   - Gagal kirim dicoba ulang otomatis (maks. 5 kali, jeda makin lama); setelah itu ditandai "gagal" dan bisa diulang manual.
//   - Pesan yang memuat rahasia (PIN PPDB) disembunyikan dari log setelah terkirim.

const waMaxTry = 5

var waFonnteURL = "https://api.fonnte.com/send" // variabel agar bisa diarahkan ke server tiruan saat tes

func init() {
	configDefaults["wa"] = `{"aktif":false,"provider":"fonnte","url":"","field_tujuan":"phone","field_pesan":"message","jeda_detik":3,` +
		`"kirim_masuk":true,"kirim_pulang":true,"kirim_alfa":true,"kirim_ppdb":true,"kirim_catatan":false,` +
		`"tpl_masuk":"Yth. Bapak/Ibu, {nama} ({kelas}) sudah tiba di sekolah pukul {jam} ({status}). — {sekolah}",` +
		`"tpl_pulang":"Yth. Bapak/Ibu, {nama} ({kelas}) pulang dari sekolah pukul {jam}. — {sekolah}",` +
		`"tpl_alfa":"Yth. Bapak/Ibu, {nama} ({kelas}) tercatat TIDAK HADIR tanpa keterangan pada {tanggal}. Mohon konfirmasi ke wali kelas. — {sekolah}",` +
		`"tpl_ppdb":"Pendaftaran PPDB {sekolah} atas nama {nama} berhasil. No. pendaftaran: {no_daftar}, PIN: {pin}. Simpan untuk cek status pendaftaran.",` +
		`"tpl_catatan":"Yth. Bapak/Ibu, informasi dari wali kelas: {nama} ({kelas}) mendapat catatan {jenis} pada {tanggal}: {keterangan}. — {sekolah}"}`
	moduleSchemas = append(moduleSchemas, `
CREATE TABLE IF NOT EXISTS wa_outbox (
  id         INTEGER PRIMARY KEY,
  kunci      TEXT NOT NULL UNIQUE,
  jenis      TEXT NOT NULL,
  tujuan     TEXT NOT NULL,
  pesan      TEXT NOT NULL,
  rahasia    INTEGER NOT NULL DEFAULT 0,
  status     TEXT NOT NULL DEFAULT 'antre' CHECK (status IN ('antre','terkirim','gagal')),
  percobaan  INTEGER NOT NULL DEFAULT 0,
  galat      TEXT,
  next_at    TEXT NOT NULL,
  created_at TEXT NOT NULL,
  sent_at    TEXT
);
CREATE INDEX IF NOT EXISTS idx_wa_antre ON wa_outbox(status, next_at);
`)
	moduleRoutes = append(moduleRoutes, func(a *App, m *http.ServeMux) {
		m.HandleFunc("GET /api/wa/config", a.h("admin", apiWAConfig))
		m.HandleFunc("PUT /api/wa/config", a.h("admin", apiWAConfigSave))
		m.HandleFunc("POST /api/wa/tes", a.h("admin", a.apiWATest))
		m.HandleFunc("GET /api/wa/outbox", a.h("admin", apiWAOutbox))
		m.HandleFunc("POST /api/wa/outbox/{id}/ulang", a.h("admin", apiWARetry))
	})
}

// ---------------------------------------------------------------------
// Nomor & templat
// ---------------------------------------------------------------------

var nonDigit = regexp.MustCompile(`\D`)

// waNumber: 0812… / +62 812… / 62812… → 62812…; kosong bila tidak tampak seperti nomor HP.
func waNumber(hp string) string {
	d := nonDigit.ReplaceAllString(hp, "")
	switch {
	case strings.HasPrefix(d, "0"):
		d = "62" + d[1:]
	case strings.HasPrefix(d, "8"):
		d = "62" + d
	}
	if len(d) < 10 || len(d) > 15 {
		return ""
	}
	return d
}

func waFill(tpl string, vals map[string]string) string {
	pairs := make([]string, 0, len(vals)*2)
	for k, v := range vals {
		pairs = append(pairs, "{"+k+"}", v)
	}
	return strings.NewReplacer(pairs...).Replace(tpl)
}

// fmtTanggal: 2026-11-09 → 9/11/2026
func fmtTanggal(s string) string {
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t.Format("2/1/2006")
	}
	return s
}

func cfgStr(c map[string]any, k string) string { s, _ := c[k].(string); return strings.TrimSpace(s) }
func cfgOn(c map[string]any, k string) bool    { b, _ := c[k].(bool); return b }

// ---------------------------------------------------------------------
// Antrean
// ---------------------------------------------------------------------

// pengganti isi pesan yang memuat PIN PPDB (log & backup tidak menyimpan PIN)
const waHidden = "(isi disembunyikan karena memuat PIN)"

// waQueue memasukkan pesan ke antrean bila notifikasi aktif dan jenis kejadian ini dinyalakan.
// Tidak pernah menggagalkan aksi utama (absensi/pendaftaran tetap tersimpan walau WA bermasalah).
func (a *App) waQueue(event, kunci, hp string, vals map[string]string, rahasia bool) {
	cfg := getConfig(a.db, "wa")
	if !cfgOn(cfg, "aktif") || !cfgOn(cfg, "kirim_"+event) {
		return
	}
	to := waNumber(hp)
	if to == "" {
		return
	}
	if vals["sekolah"] == "" {
		var s string
		a.db.QueryRow(`SELECT nama_sekolah FROM settings WHERE id=1`).Scan(&s)
		vals["sekolah"] = s
	}
	msg := waFill(cfgStr(cfg, "tpl_"+event), vals)
	if msg == "" {
		return
	}
	now := nowStr()
	a.db.Exec(`INSERT OR IGNORE INTO wa_outbox (kunci, jenis, tujuan, pesan, rahasia, next_at, created_at) VALUES (?,?,?,?,?,?,?)`,
		kunci, event, to, msg, b2i(rahasia), now, now)
}

// siswaWA: nama, kelas, dan nomor HP orang tua (Buku Induk) seorang siswa.
func (a *App) siswaWA(id int64) (nama, kelas, hp string) {
	var k, h sql.NullString
	a.db.QueryRow(`SELECT u.nama_lengkap, k.nama, p.hp_ortu FROM users u LEFT JOIN classes k ON k.id=u.kelas_id
		LEFT JOIN student_profiles p ON p.siswa_id=u.id WHERE u.id=?`, id).Scan(&nama, &k, &h)
	return nama, k.String, h.String
}

// ---------------------------------------------------------------------
// Pengirim
// ---------------------------------------------------------------------

var waHTTP = &http.Client{Timeout: 20 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}

// waSend mengirim satu pesan lewat gateway sesuai pengaturan.
func waSend(cfg map[string]any, token, to, msg string) error {
	form := url.Values{}
	endpoint := waFonnteURL
	if cfgStr(cfg, "provider") == "kustom" {
		endpoint = cfgStr(cfg, "url")
		ft, fp := cfgStr(cfg, "field_tujuan"), cfgStr(cfg, "field_pesan")
		if ft == "" {
			ft = "phone"
		}
		if fp == "" {
			fp = "message"
		}
		form.Set(ft, to)
		form.Set(fp, msg)
	} else {
		form.Set("target", to)
		form.Set("message", msg)
		form.Set("countryCode", "62")
	}
	if !urlRe.MatchString(endpoint) {
		return fmt.Errorf("URL gateway belum diisi")
	}
	req, err := http.NewRequest("POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	res, err := waHTTP.Do(req)
	if err != nil {
		return fmt.Errorf("gateway tidak bisa dihubungi")
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 64<<10))
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return fmt.Errorf("gateway menjawab HTTP %d", res.StatusCode)
	}
	// Fonnte/Wablas: {"status": true/false, "reason"/"message": "..."}
	var r map[string]any
	if json.Unmarshal(body, &r) == nil {
		if st, ok := r["status"].(bool); ok && !st {
			reason := cfgStr(r, "reason")
			if reason == "" {
				reason = cfgStr(r, "message")
			}
			if reason == "" {
				reason = "ditolak gateway"
			}
			if len(reason) > 200 {
				reason = reason[:200]
			}
			return fmt.Errorf("%s", reason)
		}
	}
	return nil
}

// waSendNext mengirim satu pesan antrean yang sudah waktunya. Mengembalikan true bila ada yang diproses.
func (a *App) waSendNext() bool {
	cfg := getConfig(a.db, "wa")
	token := kvGet(a.db, "wa_token")
	if !cfgOn(cfg, "aktif") {
		return false
	}
	// pesan kehadiran yang tertunda >6 jam (gateway mati / notifikasi sempat dimatikan) tidak dikirim lagi:
	// "sudah tiba pukul 07.00" yang datang sore/esok hari menyesatkan. Pesan lain kedaluwarsa setelah 3 hari.
	a.db.Exec(`UPDATE wa_outbox SET status='gagal', galat='kedaluwarsa: tertunda terlalu lama',
		pesan=CASE WHEN rahasia=1 THEN '`+waHidden+`' ELSE pesan END WHERE status='antre'
		AND ((jenis IN ('masuk','pulang','alfa') AND created_at < ?) OR created_at < ?)`,
		time.Now().UTC().Add(-6*time.Hour).Format(time.RFC3339), time.Now().UTC().Add(-72*time.Hour).Format(time.RFC3339))
	var id int64
	var to, msg string
	var try, secret int
	now := nowStr()
	if a.db.QueryRow(`SELECT id, tujuan, pesan, percobaan, rahasia FROM wa_outbox WHERE status='antre' AND next_at<=? ORDER BY id LIMIT 1`, now).
		Scan(&id, &to, &msg, &try, &secret) != nil {
		return false
	}
	// "sewa" 10 menit sebelum mengirim: bila status gagal diperbarui setelah terkirim (DB sibuk/server mati),
	// pesan tidak dikirim ulang tiap beberapa detik.
	if _, err := a.db.Exec(`UPDATE wa_outbox SET next_at=? WHERE id=? AND status='antre'`, time.Now().UTC().Add(10*time.Minute).Format(time.RFC3339), id); err != nil {
		return false
	}
	if err := waSend(cfg, token, to, msg); err != nil {
		try++
		status, next := "antre", time.Now().UTC().Add(time.Duration(30<<try)*time.Second).Format(time.RFC3339) // 1, 2, 4, 8 menit…
		if try >= waMaxTry {
			status = "gagal"
		}
		// pesan berisi PIN yang akhirnya gagal tidak disimpan terbuka (log & backup)
		a.db.Exec(`UPDATE wa_outbox SET status=?, percobaan=?, galat=?, next_at=?,
			pesan=CASE WHEN rahasia=1 AND ?='gagal' THEN '`+waHidden+`' ELSE pesan END WHERE id=?`, status, try, err.Error(), next, status, id)
		return true
	}
	if secret == 1 {
		msg = waHidden
	}
	if _, err := a.db.Exec(`UPDATE wa_outbox SET status='terkirim', percobaan=?, galat=NULL, sent_at=?, pesan=? WHERE id=?`, try+1, nowStr(), msg, id); err != nil {
		log.Println("wa: status terkirim gagal disimpan (pesan tidak akan diulang sebelum 10 menit):", err)
	}
	return true
}

// waLoop: pekerja latar yang mengosongkan antrean dengan jeda antar pesan.
func (a *App) waLoop() {
	for {
		if a.waSendNext() {
			jeda := 3.0
			if v, ok := getConfig(a.db, "wa")["jeda_detik"].(float64); ok && v >= 1 && v <= 60 {
				jeda = v
			}
			if a.sleep(time.Duration(jeda * float64(time.Second))) {
				return
			}
			continue
		}
		if a.sleep(5 * time.Second) {
			return
		}
	}
}

// ---------------------------------------------------------------------
// API admin
// ---------------------------------------------------------------------

func apiWAConfig(c *Ctx) error {
	cfg := getConfig(c.A.db, "wa")
	cfg["ada_token"] = kvGet(c.A.db, "wa_token") != ""
	var antre, gagal int
	c.A.db.QueryRow(`SELECT COUNT(*) FROM wa_outbox WHERE status='antre'`).Scan(&antre)
	c.A.db.QueryRow(`SELECT COUNT(*) FROM wa_outbox WHERE status='gagal'`).Scan(&gagal)
	cfg["antre"], cfg["gagal"] = antre, gagal
	return c.JSON(cfg)
}

// waHost: tujuan token (dipakai untuk menghapus token bila gateway berganti host).
func waHost(cfg map[string]any) string {
	if cfgStr(cfg, "provider") == "kustom" {
		return cfgStr(cfg, "url")
	}
	return waFonnteURL
}

func apiWAConfigSave(c *Ctx) error {
	var in map[string]any
	if err := c.Bind(&in); err != nil {
		return err
	}
	cfg := getConfig(c.A.db, "wa")
	oldHost := waHost(cfg)
	for k := range cfg {
		v, ok := in[k]
		if !ok {
			continue
		}
		switch cfg[k].(type) {
		case bool:
			b, _ := v.(bool)
			cfg[k] = b
		case float64:
			if n, ok := v.(float64); ok {
				cfg[k] = n
			}
		case string:
			s, _ := v.(string)
			if len(s) > 1000 {
				return bad("Isian terlalu panjang.")
			}
			cfg[k] = strings.TrimSpace(s)
		}
	}
	if p := cfgStr(cfg, "provider"); p != "fonnte" && p != "kustom" {
		return bad("Pilih gateway.")
	}
	if cfgStr(cfg, "provider") == "kustom" && cfgStr(cfg, "url") != "" && !urlRe.MatchString(cfgStr(cfg, "url")) {
		return bad("URL gateway harus diawali http:// atau https://")
	}
	if j, _ := cfg["jeda_detik"].(float64); j < 1 || j > 60 {
		return bad("Jeda antar pesan 1–60 detik.")
	}
	tok, _ := in["token"].(string)
	tok = strings.TrimSpace(tok)
	hapus, _ := in["hapus_token"].(bool)
	if tok == "" && !sameHost(oldHost, waHost(cfg)) && kvGet(c.A.db, "wa_token") != "" {
		hapus = true // gateway berganti tanpa token baru: jangan kirim token lama ke host lain
	}
	b, _ := json.Marshal(cfg)
	if err := kvSet(c.A.db, "cfg_wa", string(b)); err != nil {
		return err
	}
	if hapus {
		kvSet(c.A.db, "wa_token", "")
	} else if tok != "" {
		kvSet(c.A.db, "wa_token", tok)
	}
	return c.OK()
}

// apiWATest mengirim satu pesan uji langsung (tanpa antrean) agar admin segera tahu hasilnya.
func (a *App) apiWATest(c *Ctx) error {
	var in struct {
		HP string `json:"hp"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	to := waNumber(in.HP)
	if to == "" {
		return bad("Nomor WhatsApp tidak valid.")
	}
	var s string
	a.db.QueryRow(`SELECT nama_sekolah FROM settings WHERE id=1`).Scan(&s)
	if err := waSend(getConfig(a.db, "wa"), kvGet(a.db, "wa_token"), to, "Tes notifikasi WhatsApp dari "+s+" berhasil. ✅"); err != nil {
		return bad("Gagal mengirim: " + err.Error())
	}
	return c.JSON(map[string]string{"tujuan": to})
}

func apiWAOutbox(c *Ctx) error {
	// pesan antre yang memuat PIN tidak ditampilkan ke admin
	q := `SELECT id, jenis, tujuan, CASE WHEN rahasia=1 THEN '` + waHidden + `' ELSE pesan END AS pesan, status, percobaan, galat, created_at, sent_at FROM wa_outbox`
	var args []any
	if st := c.R.URL.Query().Get("status"); st == "antre" || st == "terkirim" || st == "gagal" {
		q += ` WHERE status=?`
		args = append(args, st)
	}
	rows, err := c.list(q+` ORDER BY id DESC LIMIT 200`, args...)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

func apiWARetry(c *Ctx) error {
	// created_at diperbarui: tanpa ini pesan lama langsung ditandai kedaluwarsa lagi oleh penyapu antrean
	res, err := c.A.db.Exec(`UPDATE wa_outbox SET status='antre', percobaan=0, galat=NULL, next_at=?, created_at=? WHERE id=? AND status='gagal'
		AND pesan NOT LIKE '(isi disembunyikan%'`, nowStr(), nowStr(), c.ID("id"))
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return bad("Hanya pesan gagal yang bisa dikirim ulang (pesan berisi PIN yang sudah disembunyikan tidak bisa).")
	}
	return c.OK()
}
