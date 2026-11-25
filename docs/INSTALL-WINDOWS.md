# Instalasi SCH di Windows

Untuk PC atau laptop sekolah dengan **Windows 10 / 11** atau **Windows Server 2016** ke atas (Windows 7 dan 8 tidak
didukung). Tidak perlu memasang program lain: SCH adalah satu file `.exe`, dan datanya tersimpan di folder `data`.

Untuk server yang menyala terus, Ubuntu/Debian lebih disarankan ([INSTALL-UBUNTU.md](INSTALL-UBUNTU.md)), tetapi
Windows tetap bisa dipakai dengan baik.

---

## 1. Unduh dan jalankan

1. Buka halaman **[Releases](https://github.com/reeltech-in/sch/releases/latest)** dan unduh `SCH-v…-windows.zip`.
2. Klik kanan file zip → **Extract All** → pilih `C:\` sehingga hasilnya folder `C:\SCH-windows`. Ganti namanya
   menjadi `C:\SCH` agar mudah. Hindari folder Desktop/Downloads dan folder yang tersinkron OneDrive.
3. Klik dua kali **`jalankan.bat`**.
   - Bila muncul **"Windows protected your PC"**, klik **More info → Run anyway**. Peringatan ini muncul karena
     programnya belum ditandatangani sertifikat berbayar.
   - Bila **Windows Firewall** bertanya, centang **Private networks** lalu **Allow access**.
4. Jendela hitam menampilkan alamat server dan **username & password admin** pertama. Password itu juga tersimpan di
   `C:\SCH\data\ADMIN-PASSWORD.txt`.
5. Buka alamatnya di browser (misalnya `http://192.168.1.10:3220`), login sebagai admin, ganti password di
   **Akun Saya**, lalu hapus file `ADMIN-PASSWORD.txt`.

Biarkan jendela hitam itu terbuka; menutupnya berarti mematikan server.

**Komputer Windows 32-bit:** `jalankan.bat` otomatis memakai `sch-server-windows-386.exe` bila versi 64-bit tidak ada.
Hapus `sch-server-windows-amd64.exe` dari folder bila komputernya 32-bit.

---

## 2. Menyala otomatis saat komputer dihidupkan (disarankan)

Agar server tetap jalan walau tidak ada yang login ke Windows, daftarkan sebagai tugas terjadwal.
Klik **Start**, ketik `cmd`, klik kanan **Command Prompt** → **Run as administrator**, lalu jalankan:

```bat
schtasks /Create /TN "SCH" /SC ONSTARTUP /RU SYSTEM /RL HIGHEST /TR "\"C:\SCH\sch-server-windows-amd64.exe\" -addr :3220 -https :3221 -db C:\SCH\data\sch.db"
```

Buka port di firewall (karena tidak ada jendela yang bertanya lagi):

```bat
netsh advfirewall firewall add rule name="SCH" dir=in action=allow protocol=TCP localport=3220,3221 profile=private,domain
```

Mulai sekarang tanpa restart:

```bat
schtasks /Run /TN "SCH"
```

Jangan menjalankan `jalankan.bat` lagi selama tugas ini aktif, karena port-nya sudah dipakai.

| Keperluan | Perintah (Command Prompt administrator) |
|---|---|
| hentikan server | `schtasks /End /TN "SCH"` |
| jalankan server | `schtasks /Run /TN "SCH"` |
| hapus dari startup | `schtasks /Delete /TN "SCH" /F` |
| lupa password admin | hentikan server, lalu `C:\SCH\sch-server-windows-amd64.exe -db C:\SCH\data\sch.db -reset-admin`, lalu jalankan lagi |

Pastikan juga komputer tidak tidur: **Settings → System → Power** → *Sleep* = **Never** saat tersambung listrik.

---

## 3. Memperbarui ke versi terbaru

1. Backup dulu: salin folder `C:\SCH\data` ke flashdisk atau folder lain.
2. Hentikan server (tutup jendela hitam, atau `schtasks /End /TN "SCH"`).
3. Unduh zip versi terbaru dari Releases, lalu timpa file `.exe` dan `.bat` di `C:\SCH`.
   **Jangan menimpa atau menghapus folder `data`.**
4. Jalankan lagi. Database diperbarui otomatis saat program menyala.

Versi yang berjalan terlihat di menu kiri aplikasi. Perubahan tiap versi ada di `CHANGELOG.md`.

---

## 4. Online lewat Cloudflare Tunnel

Pakai **`jalankan-online.bat`** (server hanya mendengarkan dari komputer itu sendiri, dengan mode proxy), pasang
`cloudflared` di komputer yang sama, lalu arahkan Public hostname ke `http://localhost:3220`. Langkah lengkapnya ada di
[README bagian 3](../README.md#3-online-lewat-cloudflare-zero-trust-tunnel).

Untuk tugas terjadwal, ganti argumen di bagian 2 menjadi
`-addr 127.0.0.1:3220 -https "" -proxy -db C:\SCH\data\sch.db` (tanda petik kosong ditulis `\"\"` di dalam `/TR`).

---

## 5. Lokasi file

| Lokasi | Isi |
|---|---|
| `C:\SCH\sch-server-windows-amd64.exe` | program |
| `C:\SCH\data\sch.db` | database |
| `C:\SCH\data\uploads\` | foto, berkas, gambar soal, e-book |
| `C:\SCH\data\backups\` | backup otomatis harian |
| `C:\SCH\PANDUAN.md` | panduan pemakaian lengkap |

Backup lengkap: menu **Pengaturan → Penyimpanan & backup → Unduh backup lengkap (.zip)**, simpan di flashdisk.
Memulihkan: hentikan server, salin file backup ke `C:\SCH`, lalu di Command Prompt jalankan
`C:\SCH\sch-server-windows-amd64.exe -db C:\SCH\data\sch.db -restore C:\SCH\NAMA-FILE.zip`, lalu jalankan server lagi.
Penjelasan lengkap ada di
[README bagian 7](../README.md#7-backup--pemulihan).
