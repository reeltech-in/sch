# Bedah UI: Referensi Halaman Login (gaya "Split Card Violet")

> **Pembaruan v1.9.0 — warna mengikuti jenjang sekolah.** Warna ungu di dokumen asli ini sudah diganti. Token `--brand*`
> sekarang ditentukan oleh pilihan **Pengaturan → Warna tema** (atribut `data-warna` pada `<html>`):
> **Biru / SMP** (bawaan) `#1D4ED8` · hover `#1E40AF` · terang `#3B82F6` · lembut `#EAF1FF`;
> **Abu-abu / SMA–SMK** `#475569` · `#334155` · `#64748B` · `#EEF2F6`;
> **Merah / SD** `#DC2626` · `#B91C1C` · `#EF4444` · `#FDEEEE`.
> Mode gelap memakai permukaan netral (`#0D1117` / `#151B24`) dengan versi brand yang lebih terang. Tata letak, radius,
> dan komponen di bawah tetap berlaku; baca "violet" sebagai "warna brand".


Sumber: screenshot halaman login LMS (kartu putih + panel ungu bergambar).
Warna diambil langsung dari piksel gambar (sampling), lalu dipetakan ke Tailwind terdekat.

## 1. Karakter & Gaya Desain (UI Style)
- **Gaya:** Modern Minimalist + *Soft Split Card*. Satu kartu besar membulat dibelah dua: kiri putih bersih (form), kanan panel violet solid dengan ilustrasi berbingkai kaca (*glass frame*) dan lencana bulat melayang di garis belahan.
- **Kesan:** ramah, profesional, "edukasi digital", fokus pada satu aksi utama.
- **Target:** platform belajar / dashboard untuk pengguna awam (guru, siswa). Kontras tinggi, elemen sedikit, tombol besar.

## 2. Palet Warna & Token (HEX & Tailwind)
| Token | HEX (sampel) | Tailwind terdekat | Dipakai untuk |
|---|---|---|---|
| Page BG | `#EEF2F7` | `bg-slate-100` (#F1F5F9) | latar halaman |
| Surface | `#FFFFFF` | `bg-white` | kartu, modal, sidebar |
| Surface-2 / Input BG | `#F9FAFC` | `bg-slate-50` (#F8FAFC) | isian form, header tabel |
| Primary | `#7D3AED` | `bg-violet-600` (#7C3AED) | tombol utama, panel, menu aktif |
| Primary hover | — | `bg-violet-700` (#6D28D9) | hover tombol |
| Primary light | `#9461F0` | `violet-500` (#8B5CF6) | ujung gradien panel |
| Primary soft | — | `bg-violet-50/100` (#F5F3FF / #EDE9FE) | latar ikon, chip aktif |
| Accent | `#FFB234` | `amber-400` (#FBBF24) | ikon petir pada lencana |
| Heading | — | `text-slate-800` (#1E293B) | judul |
| Body | `#6F7985` | `text-slate-500` (#64748B) | subjudul, teks isi |
| Muted / Placeholder | — | `text-slate-400` (#94A3B8) | placeholder, ikon input |
| Border | — | `border-slate-200` (#E2E8F0) | garis pemisah (sangat tipis/jarang) |
| Success / Warning / Danger | — | `emerald-600` / `amber-600` / `rose-600` | status (tidak ada di referensi, ditambahkan senada) |

## 3. Tipografi & Hierarki
- **Font:** sans geometris tebal (mirip *Inter* / *Plus Jakarta Sans*). Untuk server offline dipakai *system font stack* (Segoe UI / SF Pro / Roboto) yang karakternya sangat dekat.
- **Heading (H1):** 30–32px, `font-extrabold` (800), `tracking-tight` (-0.02em), `leading-tight`.
- **Subjudul:** 13–14px, `font-medium` (500), slate-500.
- **Input/placeholder:** 14px, 400, slate-400.
- **Tombol:** 15px, `font-bold` (700), putih.
- Rata kiri di kolom form; ruang kosong besar di atas judul.

## 4. Struktur Layout & Navigasi
- **Halaman login:** kartu tengah layar, lebar ±750–900px, 2 kolom 50/50. Kiri: padding ±52px, form vertikal. Kanan: panel violet berisi ilustrasi dalam bingkai membulat; dekorasi bentuk lengkung transparan di latar panel.
- **Gap:** antar input 16px; input → tombol 24px; judul → subjudul 6px; subjudul → form 36px.
- **Navigasi aplikasi (turunan gaya):** sidebar putih kiri; item aktif = pil violet solid dengan teks putih + bayangan violet lembut (menggemakan tombol Login).
- **Responsif:** < 860px panel ilustrasi disembunyikan, kartu menjadi satu kolom penuh; sidebar menjadi laci (*drawer*).

## 5. Bedah Komponen Detail
- **Kartu utama:** `rounded-[28px]`, tanpa border, bayangan sangat lembut (`shadow-xl` dengan opasitas rendah), sisi kanan ikut membulat bersama panel.
- **Panel violet:** solid `#7C3AED` dengan gradien halus ke `#8B5CF6` di bawah; bentuk lengkung putih transparan 10–15%.
- **Bingkai ilustrasi:** `rounded-3xl` (24px), border 8px putih transparan ±20% (efek kaca), bayangan dalam.
- **Lencana melayang:** lingkaran putih 44px, `rounded-full`, `shadow-lg`, ikon petir amber, diletakkan tepat di garis belahan kartu.
- **Input:** *filled*, tinggi 44px, `rounded-xl` (12px), latar slate-50, **tanpa border**; ikon kiri slate-400 (user, gembok); fokus = ring violet 3px.
- **Tombol utama:** tinggi 44px, `rounded-2xl` (14px), violet-600, teks putih tebal, **bayangan berwarna** `0 10px 20px -6px rgba(124,58,237,.5)`; lebar mengikuti teks (min 140px).

## 6. Detail Visual & Interaksi (Shadows, Radius, Icons)
- **Radius:** kartu 28px · bingkai/sheet 24px · tombol 14px · input 12px · chip `rounded-full`.
- **Bayangan:** netral sangat tipis untuk kartu; bayangan *violet* hanya untuk tombol utama & menu aktif.
- **Ikon:** garis (outline) 2px, ujung membulat → **Lucide Icons** (sudah dipakai, digambar inline agar offline).
- **Interaksi:** hover tombol → violet-700 + naik 1px; transisi 150ms; fokus ring violet; loading = skeleton slate.

## 7. Design System Blueprint (Untuk Webapp Builder)
```
STYLE      : Modern minimalist, soft split-card, violet primary, putih bersih.
COLORS     : page #EEF2F7 · surface #FFFFFF · input/surface-2 #F8FAFC · border #E2E8F0
             primary #7C3AED (hover #6D28D9, light #8B5CF6, soft #F3EEFF) · accent #FBBF24
             heading #1E293B · body #64748B · muted #94A3B8
             success #059669 · warning #D97706 · danger #E11D48
TYPE       : system sans; H1 30px/800/-0.02em; H2 18px/700; body 14–15px/400; label 13px/600.
RADIUS     : card 28px · sheet 24px · button 14px · input 12px · chip full.
SHADOW     : card 0 1px 2px rgb(15 23 42/.04), 0 12px 32px -12px rgb(15 23 42/.12)
             primary 0 10px 20px -6px rgb(124 58 237/.5)
INPUT      : filled #F8FAFC, tanpa border, h-44, ikon kiri muted, focus ring 3px violet/20.
BUTTON     : primary violet solid + shadow violet; ghost = putih + border tipis.
LOGIN      : kartu split 2 kolom; kanan panel violet + ilustrasi berbingkai kaca + lencana bulat amber
             di garis belahan; < 860px panel disembunyikan.
NAV        : sidebar putih; item aktif pil violet solid teks putih + shadow violet.
STATS      : ikon dalam kotak violet-soft 44px rounded-xl.
ICONS      : Lucide (outline 2px), inline SVG.
```
