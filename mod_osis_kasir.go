package main

import (
	"net/http"
	"sync"
	"time"
)

// Modul: E-Voting OSIS (suara rahasia, mode HP & mode bilik scan kartu) dan Kasir/Koperasi (barcode, struk, laporan).

func init() {
	moduleSchemas = append(moduleSchemas, `
CREATE TABLE IF NOT EXISTS elections (
  id           INTEGER PRIMARY KEY,
  judul        TEXT NOT NULL,
  deskripsi    TEXT,
  mulai        TEXT NOT NULL,
  selesai      TEXT NOT NULL,
  pemilih      TEXT NOT NULL DEFAULT 'siswa' CHECK (pemilih IN ('siswa','semua')),
  kelas_ids    TEXT NOT NULL DEFAULT '[]',
  tampil_hasil INTEGER NOT NULL DEFAULT 1,
  created_at   TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS candidates (
  id          INTEGER PRIMARY KEY,
  election_id INTEGER NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
  no          INTEGER NOT NULL,
  nama        TEXT NOT NULL,
  wakil       TEXT,
  visi        TEXT,
  misi        TEXT,
  foto        TEXT,
  suara       INTEGER NOT NULL DEFAULT 0
);
-- siapa yang SUDAH memilih (tanpa pilihannya) -> suara rahasia
CREATE TABLE IF NOT EXISTS election_voters (
  election_id INTEGER NOT NULL REFERENCES elections(id) ON DELETE CASCADE,
  user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  voted_at    TEXT NOT NULL,
  metode      TEXT,
  PRIMARY KEY (election_id, user_id)
);

CREATE TABLE IF NOT EXISTS products (
  id         INTEGER PRIMARY KEY,
  kode       TEXT NOT NULL UNIQUE COLLATE NOCASE,
  nama       TEXT NOT NULL,
  kategori   TEXT,
  harga      REAL NOT NULL DEFAULT 0,
  harga_beli REAL NOT NULL DEFAULT 0,
  stok       INTEGER NOT NULL DEFAULT 0,
  satuan     TEXT,
  aktif      INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS sales (
  id         INTEGER PRIMARY KEY,
  no         TEXT NOT NULL UNIQUE,
  total      REAL NOT NULL,
  bayar      REAL NOT NULL,
  kembali    REAL NOT NULL,
  kasir_id   INTEGER REFERENCES users(id) ON DELETE SET NULL,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS sale_items (
  sale_id    INTEGER NOT NULL REFERENCES sales(id) ON DELETE CASCADE,
  product_id INTEGER REFERENCES products(id) ON DELETE SET NULL,
  nama       TEXT NOT NULL,
  harga      REAL NOT NULL,
  harga_beli REAL NOT NULL DEFAULT 0,
  qty        INTEGER NOT NULL,
  subtotal   REAL NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sales_at ON sales(created_at);
`)
	moduleRoutes = append(moduleRoutes, func(a *App, m *http.ServeMux) {
		e := func(roles string, fn func(*Ctx) error) http.HandlerFunc { return a.mod("evoting", roles, fn) }
		m.HandleFunc("GET /api/evote", e("admin", apiElections))
		m.HandleFunc("POST /api/evote", e("admin", apiElectionSave))
		m.HandleFunc("PUT /api/evote/{id}", e("admin", apiElectionSave))
		m.HandleFunc("DELETE /api/evote/{id}", e("admin", apiDelete("elections")))
		m.HandleFunc("POST /api/evote/{id}/kandidat", e("admin", apiCandidateSave))
		m.HandleFunc("PUT /api/evote-kandidat/{id}", e("admin", apiCandidateSave))
		m.HandleFunc("DELETE /api/evote-kandidat/{id}", e("admin", apiCandidateDelete))
		m.HandleFunc("GET /api/evote/{id}/hasil", e("any", apiElectionResult))
		m.HandleFunc("GET /api/evote-aktif", e("any", apiElectionsForMe))
		m.HandleFunc("POST /api/evote/{id}/pilih", e("siswa,guru", apiVote))
		m.HandleFunc("POST /api/evote/{id}/bilik", e("admin", a.apiVoteBooth))

		k := func(roles string, fn func(*Ctx) error) http.HandlerFunc { return a.mod("kasir", roles, fn) }
		m.HandleFunc("GET /api/produk", k("admin", apiProducts))
		m.HandleFunc("POST /api/produk", k("admin", apiProductSave))
		m.HandleFunc("PUT /api/produk/{id}", k("admin", apiProductSave))
		m.HandleFunc("DELETE /api/produk/{id}", k("admin", apiDelete("products")))
		m.HandleFunc("POST /api/produk/{id}/stok", k("admin", apiProductStock))
		m.HandleFunc("POST /api/kasir/jual", k("admin", apiSell))
		m.HandleFunc("GET /api/kasir/penjualan", k("admin", apiSales))
		m.HandleFunc("GET /api/kasir/penjualan/{id}", k("admin", apiSale))
		m.HandleFunc("DELETE /api/kasir/penjualan/{id}", k("admin", apiSaleVoid))
	})
}

var (
	boothMu   sync.Mutex
	boothLast = map[int64]time.Time{}
)
