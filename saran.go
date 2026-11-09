package main

import "net/http"

// Saran isian (dropdown + boleh isi baru): nilai unik yang sudah pernah dipakai, mis. kategori aset atau rak buku.
// Hanya kolom di daftar ini yang bisa dibaca, dan tiap kolom dijaga hak akses modulnya.

var saranFields = []struct {
	key, mod, roles, sql string
}{
	{"aset.kategori", "sarpras", "admin", `SELECT DISTINCT kategori FROM assets WHERE kategori<>'' ORDER BY 1 LIMIT 200`},
	{"aset.satuan", "sarpras", "admin", `SELECT DISTINCT satuan FROM assets WHERE satuan<>'' ORDER BY 1 LIMIT 200`},
	{"aset.sumber_dana", "sarpras", "admin", `SELECT DISTINCT sumber_dana FROM assets WHERE sumber_dana<>'' ORDER BY 1 LIMIT 200`},
	{"aset.lokasi", "sarpras", "admin", `SELECT DISTINCT lokasi FROM assets WHERE lokasi<>'' ORDER BY 1 LIMIT 300`},
	{"buku.kategori", "perpus", "admin", `SELECT DISTINCT kategori FROM books WHERE kategori<>'' ORDER BY 1 LIMIT 200`},
	{"buku.rak", "perpus", "admin", `SELECT DISTINCT rak FROM books WHERE rak<>'' ORDER BY 1 LIMIT 200`},
	{"buku.penerbit", "perpus", "admin", `SELECT DISTINCT penerbit FROM books WHERE penerbit<>'' ORDER BY 1 LIMIT 300`},
	{"produk.kategori", "kasir", "admin", `SELECT DISTINCT kategori FROM products WHERE kategori<>'' ORDER BY 1 LIMIT 200`},
	{"produk.satuan", "kasir", "admin", `SELECT DISTINCT satuan FROM products WHERE satuan<>'' ORDER BY 1 LIMIT 200`},
	{"surat.disposisi", "surat", "admin", `SELECT DISTINCT disposisi_ke FROM letters WHERE disposisi_ke<>'' ORDER BY 1 LIMIT 200`},
	{"ujian.pengawas", "ujian", "admin", `SELECT DISTINCT pengawas FROM exam_rooms WHERE pengawas<>'' ORDER BY 1 LIMIT 200`},
	{"ujian.guru", "ujian", "admin", `SELECT nama_lengkap FROM users WHERE role='guru' AND status='aktif' ORDER BY 1 LIMIT 500`},
}

func init() {
	moduleRoutes = append(moduleRoutes, func(a *App, m *http.ServeMux) {
		for _, f := range saranFields {
			q := f.sql
			m.HandleFunc("GET /api/saran/"+f.key, a.mod(f.mod, f.roles, func(c *Ctx) error {
				out := []string{}
				rows, err := c.A.rdb.Query(q)
				if err != nil {
					return err
				}
				defer rows.Close()
				for rows.Next() {
					var s string
					if rows.Scan(&s) == nil {
						out = append(out, s)
					}
				}
				return c.JSON(out)
			}))
		}
	})
}
