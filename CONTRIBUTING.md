# Berkontribusi ke SCH

Terima kasih sudah ingin membantu. SCH dipakai sekolah dengan server sederhana, jadi setiap perubahan dijaga tetap
ringan, berjalan tanpa internet, dan mudah dipakai guru.

## Melaporkan bug atau usul fitur

Buka **Issue** dan tuliskan:
- versi SCH dan sistem operasi server (Windows/Ubuntu/Raspberry Pi);
- apa yang dilakukan, apa yang diharapkan, dan apa yang terjadi (tangkapan layar membantu);
- peran akun (admin/guru/siswa).

Jangan menempelkan data siswa sungguhan, password, atau API key. Celah keamanan dilaporkan secara pribadi
(lihat [SECURITY.md](SECURITY.md)).

## Menyiapkan lingkungan pengembangan

Butuh Go versi yang tertulis di `go.mod`. Tidak perlu Node.js maupun database terpisah.

```bash
git clone https://github.com/reeltech-in/sch.git && cd sch
```

```bash
go run . -db demo-data/sch.db -demo -addr :8090 -https :8493
```

Buka `http://localhost:8090`. Akun contoh: admin (password tampil di terminal), guru `guru001` / `guru123`,
siswa `siswa001` / `siswa123`. Tampilan (`web/`) tertanam di program, jadi **jalankan ulang server** setelah
mengubah file JS/CSS.

## Sebelum mengirim Pull Request

```bash
gofmt -l . && go vet ./... && go test ./...
```

- Satu PR untuk satu perubahan; jelaskan alasannya dan cara mencobanya.
- Tambahkan tes Go untuk aturan akses, nilai, atau perhitungan baru.
- Periksa akses di server: siswa hanya melihat kelasnya, guru hanya mapel miliknya, modul nonaktif ditolak.
- Teks tampilan berbahasa Indonesia yang wajar dan mudah dipahami guru.
- Jangan menambah library JavaScript luar atau layanan yang butuh internet untuk fitur inti.
- Tulis perubahan yang terlihat pengguna di `CHANGELOG.md` (bagian paling atas).

Struktur kode dijelaskan di [README bagian Untuk pengembang](README.md#11-untuk-pengembang).

## Lisensi

Dengan mengirim kontribusi, Anda setuju kontribusi itu dirilis di bawah lisensi
[GNU AGPL-3.0](LICENSE) yang sama dengan proyek ini.
