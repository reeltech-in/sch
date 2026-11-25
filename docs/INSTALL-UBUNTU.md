# Instalasi SCH di Server Ubuntu / Debian

Panduan ini untuk memasang SCH di server baru cukup dengan **menempelkan satu perintah di terminal**.
Skrip `install.sh` otomatis:

1. memasang dependensi: `ca-certificates`, `curl`, `git`, `tar`, `sqlite3`
2. memasang **Go** versi yang dibutuhkan (dari go.dev, checksum diperiksa) ke `/usr/local/go`
3. mengunduh kode SCH dan mem-build programnya
4. membuat user sistem `sch` dan folder `/opt/sch` (data di `/opt/sch/data`, hanya bisa dibaca user `sch`)
5. memasang layanan **systemd** `sch` (jalan otomatis saat server menyala, restart sendiri bila berhenti)
6. mengecek program benar-benar berjalan, lalu menampilkan alamat dan **password admin pertama**

Untuk Ubuntu 22.04 / 24.04 dan Debian 12 / 13 (x86_64, arm64). Butuh akses `sudo` dan internet saat memasang;
setelah terpasang, SCH berjalan tanpa internet.

**Kebutuhan server:** RAM 1 GB, ruang disk 2 GB untuk program dan Go (ditambah foto, berkas, dan e-book sekolah),
dan IP tetap di jaringan sekolah (atur *DHCP reservation* di router). VM, PC bekas, dan mini PC sama-sama bisa.

Pakai Windows? Lihat [INSTALL-WINDOWS.md](INSTALL-WINDOWS.md).

---

## 1. Pasang baru

Masuk ke server lewat SSH (atau buka terminal di server), lalu tempel satu baris ini:

```bash
curl -fsSL https://raw.githubusercontent.com/reeltech-in/sch/main/install.sh | sudo bash
```

Proses pertama memakan waktu sekitar 2–5 menit, tergantung kecepatan internet (mengunduh Go ±70 MB).

> **Ingin membaca skripnya dulu?** Unduh, periksa, lalu jalankan:
> `curl -fsSLO https://raw.githubusercontent.com/reeltech-in/sch/main/install.sh`, lalu `less install.sh`,
> lalu `sudo bash install.sh`.

### Dari hasil git clone (fork atau kode yang Anda ubah)

```bash
sudo apt-get update && sudo apt-get install -y git
```

```bash
git clone https://github.com/reeltech-in/sch.git ~/sch && cd ~/sch && sudo ./install.sh
```

Skrip memakai kode dari folder tempat ia berada, termasuk perubahan Anda. Untuk fork, ganti alamat repo-nya.

Di akhir proses tampil seperti ini:

```
SCH v1.28.0 berjalan.
  Buka di browser : http://192.168.1.10:3220
  HTTPS lokal     : https://192.168.1.10:3221  (untuk kamera HP)
Login pertama (segera ganti password, lalu hapus file ini):
  username: admin   password: ........
```

Setelah login: ganti password di **Akun Saya**, lalu hapus file password:

```bash
sudo rm /opt/sch/data/ADMIN-PASSWORD.txt
```

---

## 2. Pengaturan saat memasang (opsional)

Tambahkan variabel sebelum `bash` (pemasangan lewat `curl`) atau sebelum `./install.sh` (dari git clone).

| Variabel | Bawaan | Fungsi |
|---|---|---|
| `SCH_PORT` | `3220` | port HTTP |
| `SCH_HTTPS_PORT` | `3221` | port HTTPS lokal untuk kamera HP; kosongkan (`SCH_HTTPS_PORT=`) untuk mematikan |
| `SCH_PROXY_IP` | – | IP mesin Cloudflare Tunnel / reverse proxy yang dipercaya (otomatis menyalakan mode proxy) |
| `SCH_PROXY` | `0` | `1` = mode proxy untuk cloudflared/nginx di server yang **sama** (hanya 127.0.0.1 dipercaya) |
| `SCH_PROXY_HEADER` | `cf` | `cf` = Cloudflare Tunnel; `xff` = **wajib** bila di belakang nginx/proxy lain (bukan Cloudflare), agar header `CF-Connecting-IP` kiriman pengunjung diabaikan. nginx harus mengisi `proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;` |
| `SCH_DEMO` | `0` | `1` = isi data contoh untuk mencoba (hanya bila database kosong) |
| `SCH_BACKUP_DIR` | – | folder backup harian di disk lain, mis. `/mnt/usb/sch-backup` (pasang disknya dulu, otomatis menyala saat boot) |
| `SCH_UFW` | `0` | `1` = buka port di firewall `ufw` |
| `SCH_TEST` | `0` | `1` = jalankan tes otomatis sebelum memasang |
| `SCH_DIR` | `/opt/sch` | folder program & data |
| `SCH_SERVICE` | `sch` | nama layanan systemd |
| `SCH_BRANCH` | `main` | cabang git yang dipasang |
| `SCH_UNIT` | `keep` | `rewrite` = tulis ulang file layanan memakai pengaturan baru |

Contoh — port 8080, online lewat Cloudflare Tunnel yang berjalan di mesin 10.0.0.5, buka firewall:

```bash
curl -fsSL https://raw.githubusercontent.com/reeltech-in/sch/main/install.sh | sudo SCH_PORT=8080 SCH_PROXY_IP=10.0.0.5 SCH_UFW=1 bash
```

Contoh — server percobaan dengan data contoh (jangan dipakai untuk data sungguhan):

```bash
curl -fsSL https://raw.githubusercontent.com/reeltech-in/sch/main/install.sh | sudo SCH_DEMO=1 bash
```

---

## 3. Memperbarui ke versi terbaru

Jalankan **perintah yang sama** lagi. Database otomatis dibackup ke
`/opt/sch/data/backups/sebelum-v<versi>-<waktu>.db`, lalu versi baru diaktifkan. Bila versi baru
tidak merespons dalam 60 detik, skrip **otomatis kembali** ke versi sebelumnya.

Dipasang lewat `curl`:

```bash
curl -fsSL https://raw.githubusercontent.com/reeltech-in/sch/main/install.sh | sudo bash
```

Dipasang dari git clone:

```bash
cd ~/sch && git pull && sudo ./install.sh
```

Apa saja yang berubah di tiap versi: [CHANGELOG.md](../CHANGELOG.md). Versi yang sedang berjalan terlihat di menu kiri
aplikasi.

Pengaturan layanan (port, proxy, dll.) dari pemasangan pertama **tetap dipakai**. Untuk mengubahnya,
tambahkan `SCH_UNIT=rewrite` beserta pengaturan barunya.

---

## 4. Perintah sehari-hari

| Keperluan | Perintah |
|---|---|
| status | `systemctl status sch` |
| log langsung | `journalctl -u sch -f` |
| restart | `sudo systemctl restart sch` |
| versi aktif | `curl -s http://127.0.0.1:3220/api/version` |
| backup manual | `sudo -u sch /opt/sch/sch-server -db /opt/sch/data/sch.db -backup /opt/sch/data/backups/manual-$(date +%F).db` |
| pulihkan backup | salin file backup ke `/opt/sch/data/`, lalu `sudo systemctl stop sch && sudo -u sch /opt/sch/sch-server -db /opt/sch/data/sch.db -restore /opt/sch/data/NAMA-FILE.zip && sudo systemctl start sch` |
| lupa password admin | `sudo systemctl stop sch && sudo -u sch /opt/sch/sch-server -db /opt/sch/data/sch.db -reset-admin && sudo systemctl start sch` |
| kembali ke versi lama | `ls /opt/sch/releases` lalu `sudo ln -sfn releases/sch-server-v1.17.1 /opt/sch/sch-server && sudo systemctl restart sch` |

Di Debian tanpa `sudo`, jalankan sebagai root dan ganti `sudo -u sch` dengan `su -s /bin/sh sch -c "..."`.

---

## 5. Online lewat Cloudflare Tunnel

1. Pasang SCH dengan `SCH_PROXY_IP=<IP mesin cloudflared>` (atau `SCH_PROXY=1` bila cloudflared di server yang sama).
2. Di Cloudflare Zero Trust → **Networks → Tunnels** → buat tunnel, pasang `cloudflared` sesuai perintah yang diberikan.
3. Tambahkan **Public Hostname**: `sekolah.domain.id` → `http://<IP-server>:3220`.

Detail lengkap ada di [README bagian 3](../README.md#3-online-lewat-cloudflare-zero-trust-tunnel).

---

## 6. Menghapus

Menghapus layanan dan program, **data tetap disimpan**:

```bash
curl -fsSL https://raw.githubusercontent.com/reeltech-in/sch/main/install.sh | sudo bash -s -- --uninstall
```

(Bila dipasang dari git clone: `cd ~/sch && sudo ./install.sh --uninstall`.)

Menghapus semua data secara permanen (tidak bisa dibatalkan — backup dulu bila perlu):

```bash
sudo rm -rf /opt/sch && sudo userdel sch
```

---

## 7. Lokasi file

| Lokasi | Isi |
|---|---|
| `/opt/sch/sch-server` | tautan ke rilis aktif |
| `/opt/sch/releases/` | 5 rilis terakhir |
| `/opt/sch/data/sch.db` | database (SQLite) |
| `/opt/sch/data/uploads/` | file unggahan |
| `/opt/sch/data/backups/` | backup otomatis harian & sebelum update |
| `/opt/sch/data/ebook-cache/` | salinan e-book Kemendikbud yang pernah dibuka (boleh dihapus; diunduh lagi saat dibuka) |
| `/opt/sch/src/` | kode sumber (bila dipasang lewat `curl`) |
| `/etc/systemd/system/sch.service` | layanan systemd |
| `/usr/local/go/` | Go (hanya untuk build; boleh dihapus, akan dipasang lagi saat update) |

---

## 8. Mengatasi masalah

| Gejala | Penyebab & solusi |
|---|---|
| `git clone gagal` | Server tidak bisa menjangkau github.com. Cek internet/DNS: `curl -I https://github.com` |
| `unduh Go gagal` / `checksum Go tidak cocok` | Unduhan dari go.dev terputus. Jalankan ulang perintah pemasangan |
| Halaman tidak terbuka dari HP/laptop lain | Firewall menutup port. Pasang ulang dengan `SCH_UFW=1`, atau `sudo ufw allow 3220/tcp` dan `sudo ufw allow 3221/tcp` |
| `tidak merespons di port 3220` | Port dipakai program lain. Cek `sudo ss -ltnp \| grep 3220`, lalu pasang dengan `SCH_PORT=8080 SCH_UNIT=rewrite` |
| Kamera HP tidak mau menyala | Kamera butuh `https://`. Buka alamat port 3221, pilih **Lanjutkan** pada peringatan sertifikat |
| Semua orang terblokir login setelah online | Server di belakang tunnel/proxy tanpa mode proxy. Pasang ulang dengan `SCH_PROXY=1` atau `SCH_PROXY_IP=<IP>` dan `SCH_UNIT=rewrite` |
| Lupa password admin | Lihat baris "lupa password admin" di [bagian 4](#4-perintah-sehari-hari) |

Log lengkap: `journalctl -u sch -n 100 --no-pager`. Bila masih bermasalah, buka Issue di GitHub dan sertakan versi
serta potongan log (hapus dulu password atau data pribadi).
