package main

import (
	"archive/zip"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Uji ketahanan hasil audit 25 Nov 2026: backup lengkap & pemulihan, dan alur inti yang sebelumnya bercakupan rendah
// (materi & tugas LMS, kasir, PPDB sampai menjadi siswa).

func unduhBackupLengkap(t *testing.T, e *env) []byte {
	t.Helper()
	a := e.login("admin", testAdminPW)
	req, _ := http.NewRequest("GET", e.srv.URL+"/api/backup/lengkap", nil)
	res, err := a.hc.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 || res.Header.Get("Content-Type") != "application/zip" {
		t.Fatalf("unduh backup lengkap: %d %s", res.StatusCode, res.Header.Get("Content-Type"))
	}
	return b
}

// Backup lengkap berisi database + berkas unggahan, dan bisa dipulihkan ke folder data baru (server pengganti).
func TestBackupLengkapDanPulihkan(t *testing.T) {
	e := newEnv(t)
	s1 := e.login("siswa001", "siswa123")
	if code, b := s1.do("POST", "/api/portofolio", map[string]any{"judul": "Poster Backup", "course_id": 1, "file": pngDataURL(t)}); code != 200 {
		t.Fatalf("simpan portofolio: %d %s", code, b)
	}
	var url string
	e.app.db.QueryRow(`SELECT file_url FROM portfolios ORDER BY id DESC LIMIT 1`).Scan(&url)
	berkas := strings.TrimPrefix(url, "/files/")

	data := unduhBackupLengkap(t, e)
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("backup lengkap bukan zip yang valid: %v", err)
	}
	isi := map[string]bool{}
	for _, f := range zr.File {
		isi[f.Name] = true
	}
	if !isi["sch.db"] || !isi["uploads/"+berkas] {
		t.Fatalf("backup lengkap harus berisi sch.db dan uploads/%s, isinya: %v", berkas, isi)
	}

	// pulihkan ke server pengganti (folder data kosong)
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "backup.zip")
	os.WriteFile(zipPath, data, 0o600)
	baru := filepath.Join(dir, "data", "sch.db")
	os.MkdirAll(filepath.Dir(baru), 0o700)
	if _, err := pulihkanBackup(baru, zipPath); err != nil {
		t.Fatalf("pulihkan: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "data", "uploads", berkas)); err != nil {
		t.Errorf("berkas unggahan tidak ikut dipulihkan: %v", err)
	}
	db, err := sql.Open("sqlite", "file:"+baru+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	var judul string
	db.QueryRow(`SELECT judul FROM portfolios WHERE file_url=?`, url).Scan(&judul)
	if judul != "Poster Backup" {
		t.Errorf("isi database tidak ikut dipulihkan (judul=%q)", judul)
	}

	// pulihkan sekali lagi di atas data yang ada: data lama disimpan, bukan dihapus
	if _, err := pulihkanBackup(baru, zipPath); err != nil {
		t.Fatalf("pulihkan kedua: %v", err)
	}
	lama, _ := filepath.Glob(filepath.Join(dir, "data", "backups", "sebelum-pulih-*.db"))
	upLama, _ := filepath.Glob(filepath.Join(dir, "data", "uploads-sebelum-pulih-*"))
	if len(lama) != 1 || len(upLama) != 1 {
		t.Errorf("data yang terpasang harus disimpan sebelum ditimpa (db %v, uploads %v)", lama, upLama)
	}
}

// Pemulihan ditolak saat server masih berjalan, dan file yang bukan backup SCH tidak pernah dipasang.
func TestPulihkanMenolakKondisiBerbahaya(t *testing.T) {
	e := newEnv(t)
	// server tes sedang berjalan (WAL berisi): harus ditolak
	e.app.db.Exec(`UPDATE settings SET nama_sekolah='Sedang Jalan' WHERE id=1`)
	dir := t.TempDir()
	cadangan := filepath.Join(dir, "cadangan.db")
	if _, err := e.app.rdb.Exec(`VACUUM INTO ?`, cadangan); err != nil {
		t.Fatal(err)
	}
	if _, err := pulihkanBackup(e.app.dbPath, cadangan); err == nil || !strings.Contains(err.Error(), "masih berjalan") {
		t.Fatalf("pemulihan saat server berjalan harus ditolak, dapat: %v", err)
	}

	target := filepath.Join(dir, "data", "sch.db")
	os.MkdirAll(filepath.Dir(target), 0o700)
	if _, err := pulihkanBackup(target, cadangan); err != nil {
		t.Fatal(err)
	}
	asli, _ := os.ReadFile(target)
	for nama, isi := range map[string][]byte{
		"bukan-db.db":  []byte("ini bukan database"),
		"kosong.db":    {},
		"bukan-sch.db": nil, // database SQLite valid tanpa tabel users
	} {
		p := filepath.Join(dir, nama)
		if isi != nil {
			os.WriteFile(p, isi, 0o600)
		} else {
			db, _ := sql.Open("sqlite", "file:"+p)
			db.Exec(`CREATE TABLE x (a)`)
			db.Close()
		}
		if _, err := pulihkanBackup(target, p); err == nil {
			t.Errorf("%s: file yang bukan backup SCH diterima", nama)
		}
		if now, _ := os.ReadFile(target); !bytes.Equal(now, asli) {
			t.Fatalf("%s: database terpasang berubah walau pemulihan ditolak", nama)
		}
	}
}

// Isi .zip tidak bisa menulis ke luar folder data (nama berkas "../" dan sejenisnya diabaikan).
func TestPulihkanZipTidakBisaMenulisKeLuar(t *testing.T) {
	e := newEnv(t)
	dir := t.TempDir()
	cadangan := filepath.Join(dir, "c.db")
	e.app.rdb.Exec(`VACUUM INTO ?`, cadangan)
	dbData, _ := os.ReadFile(cadangan)
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for nama, isi := range map[string][]byte{
		"sch.db": dbData, "../jahat.txt": []byte("x"), "uploads/../../jahat2.txt": []byte("x"),
		"uploads/bukan-nama-sah.sh": []byte("x"), "/etc/jahat3": []byte("x"),
	} {
		w, _ := zw.Create(nama)
		w.Write(isi)
	}
	zw.Close()
	zp := filepath.Join(dir, "jahat.zip")
	os.WriteFile(zp, buf.Bytes(), 0o600)
	target := filepath.Join(dir, "data", "sch.db")
	os.MkdirAll(filepath.Dir(target), 0o700)
	if _, err := pulihkanBackup(target, zp); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{filepath.Join(dir, "jahat.txt"), filepath.Join(dir, "jahat2.txt"), filepath.Join(dir, "data", "uploads", "bukan-nama-sah.sh")} {
		if _, err := os.Stat(p); err == nil {
			t.Errorf("isi zip tertulis di %s", p)
		}
	}
}

// Backup harian bisa diarahkan ke disk lain lewat -backup-dir.
func TestFolderBackupDiskLain(t *testing.T) {
	e := newEnv(t)
	lain := t.TempDir()
	backupDirFlag = lain
	defer func() { backupDirFlag = "" }()
	if e.app.backupDir() != lain {
		t.Fatalf("backupDir = %s, harusnya %s", e.app.backupDir(), lain)
	}
}

// Alur inti LMS: guru membuat materi & tugas di mapelnya, siswa kelas itu melihat & mengumpulkan, guru menilai.
// Siswa kelas lain dan guru lain tidak bisa ikut campur.
func TestAlurMateriTugasNilai(t *testing.T) {
	e := newEnv(t)
	g1 := e.login("guru001", "guru123")
	mtk7a := e.id(`SELECT c.id FROM courses c JOIN classes k ON k.id=c.kelas_id WHERE c.kode='MTK' AND k.nama='7A'`)
	ipa7a := e.id(`SELECT id FROM courses WHERE kode='IPA'`)

	code, b := g1.do("POST", "/api/materials", map[string]any{"course_id": mtk7a, "judul": "Pecahan Senilai", "konten": "Isi materi"})
	expect(t, "guru membuat materi di mapelnya", code, 200, b)
	code, b = g1.do("POST", "/api/materials", map[string]any{"course_id": ipa7a, "judul": "Bukan mapelnya", "konten": "x"})
	if code == 200 {
		t.Errorf("guru bisa membuat materi di mapel guru lain: %s", b)
	}

	s1 := e.login("siswa001", "siswa123") // 7A
	var mats []map[string]any
	s1.json("GET", fmt.Sprintf("/api/materials?course_id=%d", mtk7a), nil, &mats)
	ada := false
	for _, m := range mats {
		ada = ada || m["judul"] == "Pecahan Senilai"
	}
	if !ada {
		t.Errorf("siswa 7A tidak melihat materi mapelnya: %v", mats)
	}
	s6 := e.login("siswa006", "siswa123") // 7B
	if code, _ := s6.do("GET", fmt.Sprintf("/api/materials?course_id=%d", mtk7a), nil); code == 200 {
		var lain []map[string]any
		s6.json("GET", fmt.Sprintf("/api/materials?course_id=%d", mtk7a), nil, &lain)
		if len(lain) > 0 {
			t.Errorf("siswa 7B melihat materi mapel 7A")
		}
	}

	var tugas map[string]any
	code = g1.json("POST", "/api/assignments", map[string]any{"course_id": mtk7a, "judul": "Latihan Pecahan", "nilai_maks": 100}, &tugas)
	expect(t, "guru membuat tugas", code, 200, nil)
	asg := e.id(`SELECT id FROM assignments WHERE judul='Latihan Pecahan'`)

	code, b = s1.do("POST", fmt.Sprintf("/api/assignments/%d/submit", asg), map[string]any{"jawaban": "1/2 = 2/4"})
	expect(t, "siswa mengumpulkan", code, 200, b)
	code, b = s6.do("POST", fmt.Sprintf("/api/assignments/%d/submit", asg), map[string]any{"jawaban": "titip"})
	expect(t, "siswa kelas lain mengumpulkan", code, 403, b)

	sub := e.id(`SELECT id FROM submissions WHERE assignment_id=?`, asg)
	g2 := e.login("guru002", "guru123")
	code, b = g2.do("PUT", fmt.Sprintf("/api/submissions/%d/grade", sub), map[string]any{"nilai": 100})
	if code == 200 {
		t.Errorf("guru lain bisa menilai tugas mapel ini: %s", b)
	}
	code, b = g1.do("PUT", fmt.Sprintf("/api/submissions/%d/grade", sub), map[string]any{"nilai": 87, "feedback": "Bagus"})
	expect(t, "guru menilai", code, 200, b)
	var nilai float64
	e.app.db.QueryRow(`SELECT nilai FROM submissions WHERE id=?`, sub).Scan(&nilai)
	if nilai != 87 {
		t.Errorf("nilai tersimpan %v, harusnya 87", nilai)
	}
	code, b = s1.do("POST", fmt.Sprintf("/api/assignments/%d/submit", asg), map[string]any{"jawaban": "ubah setelah dinilai"})
	expect(t, "ubah jawaban setelah dinilai", code, 400, b)
}

// Kasir: harga & total dihitung server (bukan dari browser), stok berkurang, stok kurang ditolak, batal
// mengembalikan stok.
func TestKasirStokDanTotalDariServer(t *testing.T) {
	e := newEnv(t)
	a := e.login("admin", testAdminPW)
	code, b := a.do("POST", "/api/produk", map[string]any{"kode": "UJI01", "nama": "Pensil Uji", "harga": 2500, "harga_beli": 2000, "stok": 10, "aktif": true})
	expect(t, "tambah produk", code, 200, b)
	pid := e.id(`SELECT id FROM products WHERE kode='UJI01'`)

	var res map[string]any
	code = a.json("POST", "/api/kasir/jual", map[string]any{"items": []map[string]any{{"product_id": pid, "qty": 3}, {"product_id": pid, "qty": 1}, {"product_id": pid, "harga": 1}}, "bayar": 20000}, &res)
	expect(t, "jual", code, 200, nil)
	var total float64
	var stok int
	e.app.db.QueryRow(`SELECT total FROM sales ORDER BY id DESC LIMIT 1`).Scan(&total)
	e.app.db.QueryRow(`SELECT stok FROM products WHERE id=?`, pid).Scan(&stok)
	if total != 10000 || stok != 6 {
		t.Errorf("total %v (harus 10000), stok %d (harus 6)", total, stok)
	}
	code, b = a.do("POST", "/api/kasir/jual", map[string]any{"items": []map[string]any{{"product_id": pid, "qty": 4}, {"product_id": pid, "qty": 3}}, "bayar": 100000})
	expect(t, "stok kurang (baris ganda digabung)", code, 400, b)
	code, b = a.do("POST", "/api/kasir/jual", map[string]any{"items": []map[string]any{{"product_id": pid, "qty": 1}}, "bayar": 100})
	expect(t, "uang bayar kurang", code, 400, b)

	sid := e.id(`SELECT id FROM sales ORDER BY id DESC LIMIT 1`)
	code, b = a.do("DELETE", fmt.Sprintf("/api/kasir/penjualan/%d", sid), nil)
	expect(t, "batalkan penjualan", code, 200, b)
	e.app.db.QueryRow(`SELECT stok FROM products WHERE id=?`, pid).Scan(&stok)
	if stok != 10 {
		t.Errorf("stok setelah batal %d, harusnya kembali 10", stok)
	}
}

// PPDB: pendaftar diterima lalu "jadikan siswa" membuat akun siswa + buku induk, dan tidak bisa dibuat dua kali.
func TestPPDBSampaiJadiSiswa(t *testing.T) {
	e := newEnv(t)
	bukaPPDB(e)
	var daftar map[string]any
	code := e.anon().json("POST", "/api/ppdb/daftar", map[string]any{"data": map[string]string{
		"nama": "Calon Siswa Baru", "tgl_lahir": "2014-03-03", "hp": "081233334444", "jk": "L", "nisn": "0099887766"}}, &daftar)
	expect(t, "daftar", code, 200, nil)
	id := e.id(`SELECT id FROM ppdb_applicants WHERE nisn='0099887766'`)
	a := e.login("admin", testAdminPW)
	kelas := e.id(`SELECT id FROM classes WHERE nama='7B'`)

	code, b := a.do("POST", fmt.Sprintf("/api/ppdb/%d/siswa", id), map[string]any{"kelas_id": kelas})
	expect(t, "jadikan siswa sebelum diterima", code, 400, b)
	code, b = a.do("PUT", fmt.Sprintf("/api/ppdb/%d", id), map[string]any{"status": "diterima"})
	expect(t, "terima pendaftar", code, 200, b)
	var hasil map[string]any
	code = a.json("POST", fmt.Sprintf("/api/ppdb/%d/siswa", id), map[string]any{"kelas_id": kelas}, &hasil)
	expect(t, "jadikan siswa", code, 200, nil)
	var nama, nisn string
	e.app.db.QueryRow(`SELECT u.nama_lengkap, sp.nisn FROM users u JOIN student_profiles sp ON sp.siswa_id=u.id
		WHERE u.id=(SELECT user_id FROM ppdb_applicants WHERE id=?)`, id).Scan(&nama, &nisn)
	if nama != "Calon Siswa Baru" || nisn != "0099887766" {
		t.Errorf("akun/buku induk siswa tidak terbentuk: nama=%q nisn=%q (%v)", nama, nisn, hasil)
	}
	code, b = a.do("POST", fmt.Sprintf("/api/ppdb/%d/siswa", id), map[string]any{"kelas_id": kelas})
	expect(t, "jadikan siswa dua kali", code, 400, b)
	_ = json.Valid
}
