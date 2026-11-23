# Kebijakan Keamanan

SCH menyimpan data pribadi siswa, guru, dan orang tua. Laporan celah keamanan sangat kami hargai.

## Melaporkan celah

**Jangan membuka issue publik untuk celah keamanan.** Laporkan secara pribadi lewat
**Security → Report a vulnerability** di halaman repo ini (GitHub Private Vulnerability Reporting).

Sertakan bila bisa:
- versi SCH (menu kiri aplikasi, atau `sch-server -version`);
- langkah untuk mengulang, dan peran akun yang dipakai (admin, guru, siswa, atau tanpa login);
- dampaknya, misalnya data siapa yang bisa dibaca atau diubah.

Kami berusaha membalas dalam 7 hari dan merilis perbaikan secepatnya. Nama pelapor dicantumkan di
CHANGELOG bila berkenan.

## Versi yang didukung

Perbaikan keamanan hanya dirilis untuk versi terbaru. Perbarui server dengan perintah pemasangan yang
sama (lihat [docs/INSTALL-UBUNTU.md](docs/INSTALL-UBUNTU.md#3-memperbarui-ke-versi-terbaru)).

## Saran untuk sekolah yang memasang SCH

- Ganti password admin pertama dan hapus `data/ADMIN-PASSWORD.txt`.
- Jangan membuka port server langsung ke internet. Untuk akses online, pakai Cloudflare Tunnel dengan `-proxy`
  atau `-proxy-ip` (lihat README bagian 3).
- Jangan pakai `-demo` di server yang berisi data sungguhan; akun contohnya memakai password yang mudah ditebak.
- Simpan backup folder `data/` di tempat terpisah dan batasi siapa yang bisa mengaksesnya.
