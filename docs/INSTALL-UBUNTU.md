# Instalasi SCH di Server Ubuntu / Debian

Panduan ini untuk memasang SCH di server baru cukup dengan **menempelkan satu perintah di terminal**.
Skrip `install.sh` otomatis:

1. memasang dependensi: `ca-certificates`, `curl`, `git`, `tar`, `sqlite3`
2. memasang **Go** versi yang dibutuhkan (dari go.dev, checksum diperiksa) ke `/usr/local/go`
3. mengunduh kode SCH dan mem-build programnya
4. membuat user sistem `sch` dan folder `/opt/sch` (data di `/opt/sch/data`, hanya bisa dibaca user `sch`)
5. memasang layanan **systemd** `sch` (jalan otomatis saat server menyala, restart sendiri bila berhenti)
6. mengecek program benar-benar berjalan, lalu menampilkan alamat dan **password admin pertama**

Untuk Ubuntu 22.04 / 24.04 dan Debian 12 / 13 (x86_64, arm64). Butuh akses `sudo` dan internet.

---

## 1. Pasang baru

### Repo publik — cukup satu baris

```bash
curl -fsSL https://raw.githubusercontent.com/reeltech-in/sch/main/install.sh | sudo bash
```

### Repo privat — clone dulu

Butuh akun GitHub yang punya akses ke repo. Cara paling mudah memakai GitHub CLI:

```bash
sudo apt-get update && sudo apt-get install -y git gh
```

```bash
gh auth login
```

```bash
gh repo clone reeltech-in/sch ~/sch && cd ~/sch && sudo ./install.sh
```

> Tanpa `gh`: `git clone https://github.com/reeltech-in/sch.git` lalu isi username GitHub dan
> **Personal Access Token** (bukan password) saat diminta.

Di akhir proses tampil seperti ini:

```
SCH v1.18.0 berjalan.
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

Tambahkan variabel sebelum `bash` (repo publik) atau sebelum `./install.sh` (repo privat).

| Variabel | Bawaan | Fungsi |
|---|---|---|
| `SCH_PORT` | `3220` | port HTTP |
| `SCH_HTTPS_PORT` | `3221` | port HTTPS lokal untuk kamera HP; kosongkan (`SCH_HTTPS_PORT=`) untuk mematikan |
| `SCH_PROXY_IP` | – | IP mesin Cloudflare Tunnel / reverse proxy yang dipercaya (otomatis menyalakan mode proxy) |
| `SCH_PROXY` | `0` | `1` = mode proxy tanpa membatasi IP |
| `SCH_DEMO` | `0` | `1` = isi data contoh untuk mencoba (hanya bila database kosong) |
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

Contoh — server percobaan dengan data contoh:

```bash
cd ~/sch && sudo SCH_DEMO=1 ./install.sh
```

---

## 3. Memperbarui ke versi terbaru

Jalankan **perintah yang sama** lagi. Database otomatis dibackup ke
`/opt/sch/data/backups/sebelum-v<versi>-<waktu>.db`, lalu versi baru diaktifkan. Bila versi baru
tidak merespons dalam 60 detik, skrip **otomatis kembali** ke versi sebelumnya.

Repo publik:

```bash
curl -fsSL https://raw.githubusercontent.com/reeltech-in/sch/main/install.sh | sudo bash
```

Repo privat:

```bash
cd ~/sch && git pull && sudo ./install.sh
```

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
cd ~/sch && sudo ./install.sh --uninstall
```

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
