package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Sesi & akun: login, logout, /api/me, ganti password, backup unduh.

// ---------------------------------------------------------------------
// Sesi
// ---------------------------------------------------------------------
func hashToken(t string) string {
	s := sha256.Sum256([]byte(t))
	return hex.EncodeToString(s[:])
}

func (a *App) currentUser(r *http.Request) *User {
	// Browser bisa mengirim lebih dari satu cookie bernama sama (mis. sesi lama); coba semuanya.
	for _, ck := range r.Cookies() {
		if ck.Name != sessionCookie || ck.Value == "" {
			continue
		}
		u := &User{}
		db := a.db
		if a.rdb != nil {
			db = a.rdb
		}
		err := db.QueryRow(`SELECT u.id, u.username, u.nama_lengkap, u.role, u.email, u.kelas_id, u.nip, u.akses, COALESCE(u.ganti_pw,0)
			FROM sessions s JOIN users u ON u.id = s.user_id
			WHERE s.token_hash = ? AND s.expires_at > ? AND u.status = 'aktif'`, hashToken(ck.Value), nowStr()).
			Scan(&u.ID, &u.Username, &u.NamaLengkap, &u.Role, &u.Email, &u.KelasID, &u.NIP, &u.AksesRaw, &u.GantiPW)
		if err == nil {
			return u
		}
	}
	return nil
}

func (a *App) userByID(id int64) *User {
	u := &User{}
	if a.db.QueryRow(`SELECT id, username, nama_lengkap, role, email, kelas_id, nip, akses, COALESCE(ganti_pw,0) FROM users WHERE id=?`, id).
		Scan(&u.ID, &u.Username, &u.NamaLengkap, &u.Role, &u.Email, &u.KelasID, &u.NIP, &u.AksesRaw, &u.GantiPW) != nil {
		return nil
	}
	return u
}

func (a *App) apiLogin(c *Ctx) error {
	var in struct{ Username, Password string }
	if err := c.Bind(&in); err != nil {
		return err
	}
	// Pembatas berlapis, ramah NAT sekolah/operator seluler (banyak orang berbagi satu IP):
	//  - pasangan IP+username: 5 gagal  -> diblokir 10 menit (hanya akun itu dari IP itu)
	//  - satu IP: 60 gagal              -> diblokir (serangan massal)
	//  - satu username: >=15 gagal      -> diperlambat 2 detik, TIDAK dikunci (cegah penguncian admin)
	ip := limitKey(clientIP(c.R))
	uname := strings.ToLower(strings.TrimSpace(in.Username))
	pairKey, ipKey, ukey := "lp:"+ip+"|"+uname, "li:"+ip, "lu:"+uname
	if c.A.limit.blockedN(pairKey, 5) || c.A.limit.blockedN(ipKey, 60) {
		return apiErr{429, "Terlalu banyak percobaan login. Coba lagi 10 menit lagi."}
	}
	if c.A.limit.blockedN(ukey, 15) {
		time.Sleep(2 * time.Second)
	}
	var id int64
	var hash string
	err := c.A.db.QueryRow(`SELECT id, password_hash FROM users WHERE username = ? COLLATE NOCASE AND status='aktif'`,
		strings.TrimSpace(in.Username)).Scan(&id, &hash)
	if err != nil {
		hash = "$2a$10$7EqJtq98hPqEX7fNZaFWoOhi5BWX4Z3hP8Ts8DHjEoS6p9Lf7pq6W" // samakan waktu respon
	}
	if err := checkPassword(hash, in.Password); err != nil || id == 0 {
		if ae, ok := err.(apiErr); ok && ae.code == 503 {
			return ae // server sibuk: bukan kesalahan password
		}
		c.A.audit(c.R, nil, "LOGIN GAGAL", 401, "username: "+strings.TrimSpace(in.Username))
		c.A.limit.fail(pairKey)
		c.A.limit.fail(ipKey)
		c.A.limit.fail(ukey)
		return apiErr{401, "Username atau password salah."}
	}
	c.A.limit.reset(pairKey)
	defer func() { c.A.audit(c.R, c.A.userByID(id), "LOGIN", 200, "") }()
	b := make([]byte, 32)
	rand.Read(b)
	tok := hex.EncodeToString(b)
	exp := time.Now().Add(sessionTTL)
	if _, err := c.A.db.Exec(`INSERT INTO sessions (token_hash, user_id, expires_at) VALUES (?,?,?)`,
		hashToken(tok), id, exp.UTC().Format(time.RFC3339)); err != nil {
		return err
	}
	http.SetCookie(c.W, &http.Cookie{Name: sessionCookie, Value: tok, Path: "/", Expires: exp,
		HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: isHTTPS(c.R)})
	return c.JSON(c.A.userByID(id))
}

func (a *App) apiLogout(c *Ctx) error {
	for _, ck := range c.R.Cookies() {
		if ck.Name == sessionCookie {
			c.A.db.Exec(`DELETE FROM sessions WHERE token_hash = ?`, hashToken(ck.Value))
		}
	}
	http.SetCookie(c.W, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true})
	return c.OK()
}

func (a *App) apiMe(c *Ctx) error {
	u := c.A.currentUser(c.R)
	if u == nil {
		return c.JSON(nil)
	}
	return c.JSON(u)
}

func (a *App) apiPassword(c *Ctx) error {
	var in struct{ Old, New string }
	if err := c.Bind(&in); err != nil {
		return err
	}
	if len(in.New) < 6 || len(in.New) > 72 {
		return bad("Password baru harus 6 sampai 72 karakter.")
	}
	if in.New == in.Old {
		return bad("Password baru harus berbeda dari password lama.")
	}
	var hash string
	c.A.db.QueryRow(`SELECT password_hash FROM users WHERE id=?`, c.U.ID).Scan(&hash)
	if checkPassword(hash, in.Old) != nil {
		return bad("Password lama salah.")
	}
	h, err := hashPassword(in.New)
	if err != nil {
		return err
	}
	tx, err := c.A.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE users SET password_hash=?, ganti_pw=0 WHERE id=?`, h, c.U.ID); err != nil {
		return err
	}
	// keluarkan sesi lain milik pengguna ini (perangkat yang mungkin dipakai orang lain)
	keep := []any{c.U.ID}
	ph := ""
	for _, ck := range c.R.Cookies() {
		if ck.Name == sessionCookie {
			keep = append(keep, hashToken(ck.Value))
			ph += ",?"
		}
	}
	if _, err := tx.Exec(`DELETE FROM sessions WHERE user_id=? AND token_hash NOT IN (''`+ph+`)`, keep...); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.OK()
}

// Unduh salinan database (admin)
func (a *App) apiBackup(c *Ctx) error {
	// Folder sementara 0700 (tidak terbaca user lain); VACUUM INTO dijalankan di pool baca
	// sehingga koneksi tulis (simpan jawaban CBT, scan absensi) tidak tertahan.
	dir, err := os.MkdirTemp("", "sch-backup-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	tmp := filepath.Join(dir, "backup.db")
	db := c.A.rdb
	if db == nil {
		db = c.A.db
	}
	if _, err := db.ExecContext(context.Background(), `VACUUM INTO ?`, tmp); err != nil {
		return err
	}
	http.NewResponseController(c.W).SetWriteDeadline(time.Now().Add(10 * time.Minute))
	name := "sch-backup-" + time.Now().Format("2006-01-02-1504") + ".db"
	c.W.Header().Set("Content-Type", "application/octet-stream")
	c.W.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	http.ServeFile(c.W, c.R, tmp)
	return nil
}
