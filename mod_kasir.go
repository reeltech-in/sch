package main

import (
	"fmt"
	"time"
)

// Kasir / Koperasi: barang & stok, penjualan dengan nomor harian, pembatalan, laporan laba.

// ---------------------------------------------------------------------
// Kasir / Koperasi
// ---------------------------------------------------------------------
func apiProducts(c *Ctx) error {
	rows, err := c.list(`SELECT * FROM products ORDER BY nama`)
	if err != nil {
		return err
	}
	return c.JSON(rows)
}

func apiProductSave(c *Ctx) error {
	var in struct {
		Kode      string  `json:"kode"`
		Nama      string  `json:"nama"`
		Kategori  *string `json:"kategori"`
		Harga     float64 `json:"harga"`
		HargaBeli float64 `json:"harga_beli"`
		Stok      int     `json:"stok"`
		Satuan    *string `json:"satuan"`
		Aktif     bool    `json:"aktif"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	kode, err := required(in.Kode, "Kode/barcode")
	if err != nil {
		return err
	}
	nama, err := required(in.Nama, "Nama barang")
	if err != nil {
		return err
	}
	if in.Harga < 0 || in.HargaBeli < 0 {
		return bad("Harga tidak boleh negatif.")
	}
	if k, ok := c.exists(`SELECT kode FROM products WHERE nama=? COLLATE NOCASE AND id<>?`, nama, c.ID("id")); ok {
		return bad("Barang " + nama + " sudah ada (kode " + k + "). Ubah barang itu atau tambah stoknya.")
	}
	in.Kategori, in.Satuan = c.canon("products", "kategori", in.Kategori), c.canon("products", "satuan", in.Satuan)
	ak := 0
	if in.Aktif {
		ak = 1
	}
	if id := c.ID("id"); id > 0 {
		_, err = c.A.db.Exec(`UPDATE products SET kode=?, nama=?, kategori=?, harga=?, harga_beli=?, satuan=?, aktif=? WHERE id=?`, kode, nama, str(in.Kategori), in.Harga, in.HargaBeli, str(in.Satuan), ak, id)
	} else {
		_, err = c.A.db.Exec(`INSERT INTO products (kode, nama, kategori, harga, harga_beli, stok, satuan, aktif, created_at) VALUES (?,?,?,?,?,?,?,?,?)`, kode, nama, str(in.Kategori), in.Harga, in.HargaBeli, in.Stok, str(in.Satuan), ak, nowStr())
	}
	if err != nil {
		return err
	}
	return c.OK()
}

func apiProductStock(c *Ctx) error {
	var in struct {
		Tambah int `json:"tambah"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	if in.Tambah == 0 {
		return bad("Isi jumlah perubahan stok.")
	}
	res, err := c.A.db.Exec(`UPDATE products SET stok=MAX(0, stok+?) WHERE id=?`, in.Tambah, c.ID("id"))
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return notFound()
	}
	return c.OK()
}

func apiSell(c *Ctx) error {
	var in struct {
		Items []struct {
			ProductID int64 `json:"product_id"`
			Qty       int   `json:"qty"`
		} `json:"items"`
		Bayar float64 `json:"bayar"`
	}
	if err := c.Bind(&in); err != nil {
		return err
	}
	if len(in.Items) == 0 {
		return bad("Keranjang masih kosong.")
	}
	tx, err := c.A.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	type line struct {
		id               int64
		nama             string
		harga, beli, sub float64
		qty              int
	}
	// gabungkan baris ganda untuk barang yang sama agar cek stok tidak bisa diakali
	qty := map[int64]int{}
	order := []int64{}
	for _, it := range in.Items {
		if it.Qty < 1 || it.Qty > 10000 {
			continue
		}
		if _, ok := qty[it.ProductID]; !ok {
			order = append(order, it.ProductID)
		}
		qty[it.ProductID] += it.Qty
	}
	in.Items = in.Items[:0]
	for _, pid := range order {
		in.Items = append(in.Items, struct {
			ProductID int64 `json:"product_id"`
			Qty       int   `json:"qty"`
		}{pid, qty[pid]})
	}
	var lines []line
	total := 0.0
	for _, it := range in.Items {
		if it.Qty < 1 {
			continue
		}
		var l line
		var stok, aktif int
		if err := tx.QueryRow(`SELECT id, nama, harga, harga_beli, stok, aktif FROM products WHERE id=?`, it.ProductID).Scan(&l.id, &l.nama, &l.harga, &l.beli, &stok, &aktif); err != nil {
			return bad("Barang tidak ditemukan.")
		}
		if aktif != 1 {
			return bad(l.nama + " sedang tidak dijual.")
		}
		if stok < it.Qty {
			return bad(fmt.Sprintf("Stok %s tinggal %d.", l.nama, stok))
		}
		l.qty, l.sub = it.Qty, l.harga*float64(it.Qty)
		total += l.sub
		lines = append(lines, l)
	}
	if total <= 0 {
		return bad("Keranjang masih kosong.")
	}
	if in.Bayar < total {
		return bad("Uang bayar kurang dari total.")
	}
	// nomor = urutan terbesar hari ini + 1 (bukan COUNT, yang bentrok setelah transaksi dibatalkan)
	var n int
	day := time.Now().Format("20060102")
	tx.QueryRow(`SELECT COALESCE(MAX(CAST(substr(no, 12) AS INTEGER)), 0) FROM sales WHERE no >= ? AND no < ?`, "TRX"+day, "TRX"+day+"~").Scan(&n)
	no := fmt.Sprintf("TRX%s%04d", day, n+1)
	res, err := tx.Exec(`INSERT INTO sales (no, total, bayar, kembali, kasir_id, created_at) VALUES (?,?,?,?,?,?)`, no, total, in.Bayar, in.Bayar-total, c.U.ID, nowStr())
	if err != nil {
		return err
	}
	sid, _ := res.LastInsertId()
	for _, l := range lines {
		if _, err := tx.Exec(`INSERT INTO sale_items (sale_id, product_id, nama, harga, harga_beli, qty, subtotal) VALUES (?,?,?,?,?,?,?)`, sid, l.id, l.nama, l.harga, l.beli, l.qty, l.sub); err != nil {
			return err
		}
		if _, err := tx.Exec(`UPDATE products SET stok=stok-? WHERE id=?`, l.qty, l.id); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	c.R.SetPathValue("id", fmt.Sprint(sid))
	return apiSale(c)
}

func apiSale(c *Ctx) error {
	s, err := c.list(`SELECT s.*, u.nama_lengkap AS kasir FROM sales s LEFT JOIN users u ON u.id=s.kasir_id WHERE s.id=?`, c.ID("id"))
	if err != nil || len(s) == 0 {
		return notFound()
	}
	items, err := c.list(`SELECT nama, harga, qty, subtotal FROM sale_items WHERE sale_id=?`, c.ID("id"))
	if err != nil {
		return err
	}
	s[0]["items"] = items
	return c.JSON(s[0])
}

func apiSales(c *Ctx) error {
	q := c.R.URL.Query()
	var from, to string
	if b := q.Get("bulan"); b != "" {
		f, t, err := monthRange(b)
		if err != nil {
			return err
		}
		from, to = f, t
	} else {
		d, err := cleanDate(q.Get("tanggal"))
		if err != nil {
			d = localDate()
		}
		from, to = d, d
	}
	t1, _ := time.ParseInLocation("2006-01-02", from, time.Local)
	t2, _ := time.ParseInLocation("2006-01-02", to, time.Local)
	a, b := t1.UTC().Format(time.RFC3339), t2.AddDate(0, 0, 1).UTC().Format(time.RFC3339)
	sales, err := c.list(`SELECT s.id, s.no, s.total, s.bayar, s.kembali, s.created_at, u.nama_lengkap AS kasir FROM sales s LEFT JOIN users u ON u.id=s.kasir_id
		WHERE s.created_at>=? AND s.created_at<? ORDER BY s.created_at DESC`, a, b)
	if err != nil {
		return err
	}
	top, err := c.list(`SELECT i.nama, SUM(i.qty) AS qty, SUM(i.subtotal) AS omzet, SUM(i.subtotal - i.harga_beli*i.qty) AS laba
		FROM sale_items i JOIN sales s ON s.id=i.sale_id WHERE s.created_at>=? AND s.created_at<? GROUP BY i.nama ORDER BY omzet DESC`, a, b)
	if err != nil {
		return err
	}
	return c.JSON(map[string]any{"dari": from, "sampai": to, "penjualan": sales, "barang": top})
}

func apiSaleVoid(c *Ctx) error {
	if c.U.Role != "admin" {
		return apiErr{403, "Pembatalan transaksi hanya oleh admin."}
	}
	tx, err := c.A.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`UPDATE products SET stok=stok+(SELECT COALESCE(SUM(qty),0) FROM sale_items WHERE sale_id=? AND product_id=products.id)
		WHERE id IN (SELECT product_id FROM sale_items WHERE sale_id=?)`, c.ID("id"), c.ID("id")); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM sales WHERE id=?`, c.ID("id")); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return c.OK()
}
