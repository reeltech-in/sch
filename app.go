package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"sync"
	"time"
)

// App (koneksi database, pembatas login, pekerja latar) dan User (akun yang sedang masuk).

type App struct {
	db       *sql.DB // koneksi tulis (tunggal)
	rdb      *sql.DB // pool baca (beberapa koneksi, WAL) untuk permintaan GET
	dbPath   string
	limit    *loginLimiter
	secretMu sync.Mutex
	secretV  []byte
	quit     chan struct{}  // ditutup saat server berhenti
	bg       sync.WaitGroup // pekerja latar yang sedang berjalan
}

// background menjalankan pekerja latar: panic dicatat lalu pekerja dijalankan ulang 30 detik kemudian
// (server tetap hidup), dan pekerja berhenti saat a.quit ditutup.
func (a *App) background(name string, fn func()) {
	a.bg.Add(1)
	go func() {
		defer a.bg.Done()
		for {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("pekerja %s panic: %v (dijalankan ulang)", name, r)
					}
				}()
				fn()
			}()
			if a.sleep(30 * time.Second) {
				return
			}
		}
	}()
}

// sleep menunggu d; true bila server sedang berhenti.
func (a *App) sleep(d time.Duration) bool {
	select {
	case <-a.quit:
		return true
	case <-time.After(d):
		return false
	}
}

type User struct {
	ID          int64          `json:"id"`
	Username    string         `json:"username"`
	NamaLengkap string         `json:"nama_lengkap"`
	Role        string         `json:"role"`
	Email       sql.NullString `json:"-"`
	KelasID     sql.NullInt64  `json:"-"`
	NIP         sql.NullString `json:"-"`
	AksesRaw    sql.NullString `json:"-"`
	GantiPW     bool           `json:"-"`
}

// Has: guru/staf yang diberi akses kelola modul tertentu oleh admin.
func (u *User) Has(mod string) bool {
	if u.Role == "admin" {
		return true
	}
	var list []string
	json.Unmarshal([]byte(u.AksesRaw.String), &list)
	for _, m := range list {
		if m == mod {
			return true
		}
	}
	return false
}

func (u *User) MarshalJSON() ([]byte, error) {
	m := map[string]any{"id": u.ID, "username": u.Username, "nama_lengkap": u.NamaLengkap, "role": u.Role,
		"email": nil, "kelas_id": nil, "nip": nil, "akses": []string{}}
	var ak []string
	if json.Unmarshal([]byte(u.AksesRaw.String), &ak) == nil && ak != nil {
		m["akses"] = ak
	}
	if u.Email.Valid {
		m["email"] = u.Email.String
	}
	if u.KelasID.Valid {
		m["kelas_id"] = u.KelasID.Int64
	}
	if u.NIP.Valid {
		m["nip"] = u.NIP.String
	}
	m["ganti_pw"] = u.GantiPW
	return json.Marshal(m)
}

func (a *App) cleanupLoop() {
	for i := 0; ; i++ {
		if i%60 == 0 {
			a.db.Exec(`DELETE FROM sessions WHERE expires_at < ?`, nowStr())
			a.db.Exec(`PRAGMA wal_checkpoint(TRUNCATE)`)
		}
		a.limit.prune()
		if i%10 == 0 {
			a.maybeMaintain()
		}
		if i%5 == 0 {
			a.finalizeCBT(0) // ujian yang waktunya habis tapi tidak dikumpulkan
		}
		if a.sleep(time.Minute) {
			return
		}
	}
}
