'use strict';
/* SCH LMS — aplikasi web (tanpa library eksternal, jalan offline). */

// =====================================================================
// Utilitas
// =====================================================================
const $ = (s, r = document) => r.querySelector(s);
const $$ = (s, r = document) => [...r.querySelectorAll(s)];
const esc = v => String(v ?? '').replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
const logoSrc = s => String(s.logo_url || '').startsWith('/files/') ? '/api/logo?v=' + encodeURIComponent(s.logo_url.slice(7, 15)) : safeUrl(s.logo_url);
const safeUrl = u => /^https?:\/\//i.test(String(u || '').trim()) ? String(u).trim() : '#';
const num = v => (v === '' || v == null ? null : Number(v));
const fmtN = n => (n == null || isNaN(n) ? '–' : (Math.round(n * 10) / 10).toLocaleString('id-ID'));
const fmtDate = s => s ? new Date(s.length === 10 ? s + 'T00:00:00' : s).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' }) : '–';
const fmtDT = s => s ? new Date(s).toLocaleString('id-ID', { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' }) : '–';
const pad = n => String(n).padStart(2, '0');
const today = () => { const d = new Date(); return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`; };
const toLocalInput = s => { if (!s) return ''; const d = new Date(s); return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`; };
const store = {
  get(k) { try { return JSON.parse(localStorage.getItem(k)); } catch { return null; } },
  set(k, v) { try { localStorage.setItem(k, JSON.stringify(v)); } catch { /* abaikan */ } },
  del(k) { try { localStorage.removeItem(k); } catch { /* abaikan */ } },
};

const ICONS = {
  home: '<path d="M3 10.5 12 3l9 7.5"/><path d="M5 9.5V21h14V9.5"/><path d="M10 21v-6h4v6"/>',
  school: '<path d="M3 21h18"/><path d="M5 21V10l7-5 7 5v11"/><path d="M10 21v-5h4v5"/>',
  users: '<circle cx="9" cy="8" r="3.5"/><path d="M2.5 20a6.5 6.5 0 0 1 13 0"/><path d="M16 4.5a3.5 3.5 0 0 1 0 7"/><path d="M18 14a6 6 0 0 1 3.5 6"/>',
  user: '<circle cx="12" cy="8" r="4"/><path d="M4 21a8 8 0 0 1 16 0"/>',
  book: '<path d="M4 4.5A2.5 2.5 0 0 1 6.5 2H20v17H6.5A2.5 2.5 0 0 0 4 21.5z"/><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/>',
  layers: '<path d="m12 3 9 5-9 5-9-5z"/><path d="m3 13 9 5 9-5"/>',
  clip: '<rect x="5" y="4" width="14" height="17" rx="2"/><path d="M9 4V3h6v1"/><path d="M9 11h6M9 15h4"/>',
  quiz: '<circle cx="12" cy="12" r="9"/><path d="M9.5 9a2.5 2.5 0 1 1 3.5 2.3c-.7.3-1 .9-1 1.7"/><path d="M12 17h.01"/>',
  cal: '<rect x="3" y="5" width="18" height="16" rx="2"/><path d="M3 10h18M8 3v4M16 3v4"/>',
  chart: '<path d="M4 20V10M10 20V4M16 20v-7M22 20H2"/>',
  gear: '<circle cx="12" cy="12" r="3"/><path d="M12 2v3M12 19v3M4.9 4.9l2.1 2.1M17 17l2.1 2.1M2 12h3M19 12h3M4.9 19.1 7 17M17 7l2.1-2.1"/>',
  out: '<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><path d="m16 17 5-5-5-5M21 12H9"/>',
  plus: '<path d="M12 5v14M5 12h14"/>',
  edit: '<path d="M12 20h9"/><path d="M16.5 3.5a2.1 2.1 0 1 1 3 3L7 19l-4 1 1-4z"/>',
  trash: '<path d="M3 6h18M8 6V4h8v2M19 6l-1 14H6L5 6"/>',
  menu: '<path d="M4 6h16M4 12h16M4 18h16"/>',
  x: '<path d="M18 6 6 18M6 6l12 12"/>',
  check: '<path d="m5 12 5 5L20 7"/>',
  print: '<path d="M6 9V2h12v7"/><rect x="6" y="14" width="12" height="8"/><path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2"/>',
  upload: '<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><path d="m17 8-5-5-5 5M12 3v12"/>',
  down: '<path d="m6 9 6 6 6-6"/>',
  download: '<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><path d="m7 10 5 5 5-5M12 15V3"/>',
  link: '<path d="M10 13a5 5 0 0 0 7.5.5l3-3a5 5 0 0 0-7-7l-1.7 1.7"/><path d="M14 11a5 5 0 0 0-7.5-.5l-3 3a5 5 0 0 0 7 7l1.7-1.7"/>',
  file: '<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><path d="M14 2v6h6"/>',
  play: '<circle cx="12" cy="12" r="9"/><path d="m10 8 6 4-6 4z"/>',
  image: '<rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="9" cy="9" r="2"/><path d="m21 15-5-5L5 21"/>',
  eye: '<path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12z"/><circle cx="12" cy="12" r="3"/>',
  clock: '<circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/>',
  inbox: '<path d="M22 12h-6l-2 3h-4l-2-3H2"/><path d="M5.5 5h13L22 12v7a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2v-7z"/>',
  key: '<circle cx="7.5" cy="15.5" r="4.5"/><path d="m10.7 12.3 9.8-9.8M17 6l3 3M14.5 8.5l2 2"/>',
  db: '<ellipse cx="12" cy="5" rx="8" ry="3"/><path d="M4 5v14c0 1.7 3.6 3 8 3s8-1.3 8-3V5"/><path d="M4 12c0 1.7 3.6 3 8 3s8-1.3 8-3"/>',
  back: '<path d="m15 18-6-6 6-6"/>',
  sun: '<circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4"/>',
  moon: '<path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z"/>',
  monitor: '<rect x="2" y="4" width="20" height="13" rx="2"/><path d="M8 21h8M12 17v4"/>',
};
// Ilustrasi panel login (SVG tertanam, tanpa unduhan gambar)
const LOGIN_ART = `<svg viewBox="0 0 280 350" xmlns="http://www.w3.org/2000/svg">
  <defs><linearGradient id="lg1" x1="0" y1="0" x2="0" y2="1"><stop offset="0" style="stop-color:var(--art-1)"/><stop offset="1" style="stop-color:var(--art-2)"/></linearGradient>
    <radialGradient id="lg2" cx=".5" cy=".5" r=".5"><stop offset="0" stop-color="#fde68a" stop-opacity=".55"/><stop offset="1" stop-color="#fde68a" stop-opacity="0"/></radialGradient></defs>
  <rect width="280" height="350" fill="url(#lg1)"/>
  <g style="fill:var(--art-3)" opacity=".55"><circle cx="30" cy="30" r="1.5"/><circle cx="250" cy="22" r="1.2"/><circle cx="210" cy="60" r="1"/><circle cx="60" cy="70" r="1"/></g>
  <rect x="36" y="44" width="208" height="120" rx="10" fill="#1f2a44" style="stroke:var(--art-4)" stroke-width="4"/>
  <g style="stroke:var(--art-3)" stroke-width="2.5" stroke-linecap="round" fill="none" opacity=".9">
    <path d="M56 74h40M56 92h62M56 110h30"/><path d="M140 70l14 14M154 70l-14 14"/><path d="M176 77h22M187 66v22"/>
    <path d="M140 116c10-18 20-18 30 0s20 18 30 0"/><circle cx="215" cy="135" r="9"/></g>
  <circle cx="226" cy="214" r="46" fill="url(#lg2)"/>
  <path d="M214 184l12 26h-24z" fill="#fbbf24"/><rect x="224" y="210" width="4" height="40" style="fill:var(--art-4)"/>
  <rect x="0" y="250" width="280" height="100" style="fill:var(--art-2)"/>
  <rect x="20" y="246" width="240" height="10" rx="5" style="fill:var(--brand-400)"/>
  <g><rect x="34" y="214" width="56" height="10" rx="2" fill="#f472b6"/><rect x="38" y="224" width="52" height="10" rx="2" fill="#60a5fa"/><rect x="32" y="234" width="60" height="12" rx="2" fill="#fbbf24"/></g>
  <rect x="108" y="190" width="92" height="56" rx="6" fill="#e5e9f0"/><rect x="113" y="195" width="82" height="44" rx="3" style="fill:var(--art-1)"/>
  <g stroke="#a5f3fc" stroke-width="2" stroke-linecap="round"><path d="M120 205h30M120 213h48M120 221h22M120 229h38"/></g>
  <path d="M100 246h108l-8-6h-92z" style="fill:var(--art-3)"/>
  <circle cx="154" cy="166" r="0" fill="none"/>
</svg>`;
// ikon tambahan (dulu di file modul; di inti agar tersedia walau modulnya belum dimuat)
ICONS.refresh = '<path d="M21 12a9 9 0 1 1-2.6-6.4L21 8"/><path d="M21 3v5h-5"/>';
ICONS.spark = '<path d="M12 3v4M12 17v4M3 12h4M17 12h4M5.6 5.6l2.8 2.8M15.6 15.6l2.8 2.8M5.6 18.4l2.8-2.8M15.6 8.4l2.8-2.8"/>';
ICONS.vote = '<path d="M9 12l2 2 4-4"/><rect x="3" y="4" width="18" height="16" rx="2"/>';
ICONS.cart = '<circle cx="9" cy="20" r="1.5"/><circle cx="18" cy="20" r="1.5"/><path d="M2 3h3l2.6 12.4a2 2 0 0 0 2 1.6h8.8a2 2 0 0 0 2-1.6L22 7H6"/>';
const ico = n =>`<svg class="i" viewBox="0 0 24 24" aria-hidden="true">${ICONS[n] || ''}</svg>`;

// =====================================================================
// API
// =====================================================================
// Inti semua permintaan ke server (api, GETpage, Asisten AI): header CSRF, pesan jaringan yang ramah,
// pemeriksaan versi aplikasi, dan koreksi jam perangkat. Permintaan GET untuk halaman dibatalkan
// (AbortController) begitu pengguna pindah menu, agar halaman baru tidak antre di belakang yang lama.
async function apiFetch(method, path, body) {
  const opt = { method, headers: { 'X-SCH': '1' }, credentials: 'same-origin' };
  if (body !== undefined) { opt.headers['Content-Type'] = 'application/json'; opt.body = JSON.stringify(body); }
  if (method === 'GET' && S.navCtl) opt.signal = S.navCtl.signal; // simpan/ubah data tidak pernah dibatalkan
  let res;
  try { res = await fetch('/api' + path, opt); } catch (e) {
    if (e.name === 'AbortError') throw e;
    throw new Error('Tidak terhubung ke server sekolah. Periksa jaringan WiFi/LAN.');
  }
  checkVersion(res.headers.get('X-SCH-Version'));
  const sd = Date.parse(res.headers.get('Date') || '');
  if (sd) S.skew = sd - Date.now();
  return res;
}
// baca jawaban JSON; sesi habis (401) → kembali ke halaman masuk tanpa menghapus jawaban ujian/kuis
async function apiJSON(res, path) {
  let data = null;
  try { data = await res.json(); } catch { /* bukan JSON */ }
  if (res.status === 401 && S.user && path !== '/login') { resetClient(true); }
  if (!res.ok) throw new Error((data && data.error) || `Gagal (${res.status})`);
  return data;
}
async function api(method, path, body) { return apiJSON(await apiFetch(method, path, body), path); }
const isAbort = e => e && e.name === 'AbortError';
// Data rujukan yang dipakai hampir semua halaman (mapel, kelas, tahun ajaran) disimpan sebentar di memori,
// supaya pindah menu tidak menunggu permintaan yang sama berulang kali lewat internet. Setiap perubahan data
// (POST/PUT/DELETE) mengosongkan simpanan ini, jadi tampilan tidak pernah memakai data basi setelah menyimpan.
// Waktu menurut server (jam perangkat siswa/PC lab bisa salah): dipakai untuk status ujian/pengumuman.
const nowMs = () => Date.now() + (Math.abs(S.skew || 0) > 60000 ? S.skew : 0);
// Versi aplikasi yang sedang terbuka; bila server sudah diperbarui, tampilkan ajakan memuat ulang (sekali).
let verNotified = false;
function checkVersion(v) {
  const cur = String((S.settings && S.settings.versi) || '').replace(/^v/, '');
  if (!v || !cur || v === cur || verNotified) return;
  verNotified = true;
  const bar = document.createElement('div');
  bar.className = 'ver-bar'; bar.setAttribute('role', 'status');
  bar.innerHTML = `<span>Versi baru aplikasi tersedia (v${esc(v)}).</span><button type="button" class="btn sm">Muat ulang</button>`;
  bar.querySelector('button').onclick = () => location.reload();
  document.body.append(bar);
}
const GET_CACHE = new Map(), GET_INFLIGHT = new Map();
const CACHE_TTL = 120000;
const cacheable = p => /^\/(courses|classes|tahun)(\?|$)/.test(p);
const clone = v => (v == null ? v : JSON.parse(JSON.stringify(v)));
function GET(p) {
  const hit = GET_CACHE.get(p);
  if (hit && Date.now() - hit.t < CACHE_TTL) return Promise.resolve(clone(hit.v));
  // permintaan yang sama yang sedang berjalan dipakai bersama (mis. dua bagian halaman butuh /courses)
  if (GET_INFLIGHT.has(p)) return GET_INFLIGHT.get(p).then(clone);
  // generasi: jawaban yang berangkat SEBELUM ada perubahan data tidak disimpan (bisa sudah basi)
  const gen = CACHE_GEN;
  const pr = api('GET', p).then(v => { if (cacheable(p) && gen === CACHE_GEN) GET_CACHE.set(p, { t: Date.now(), v }); return v; })
    .finally(() => { if (GET_INFLIGHT.get(p) === pr) GET_INFLIGHT.delete(p); });
  GET_INFLIGHT.set(p, pr);
  return pr.then(clone);
}
let CACHE_GEN = 0;
const mutate = (m, p, b) => { CACHE_GEN++; return api(m, p, b).finally(() => { CACHE_GEN++; GET_CACHE.clear(); GET_INFLIGHT.clear(); }); };
const POST = (p, b = {}) => mutate('POST', p, b);
const PUT = (p, b = {}) => mutate('PUT', p, b);
const DEL = p => mutate('DELETE', p);

// =====================================================================
// Status aplikasi
// =====================================================================
const S = { user: null, settings: {}, page: 'dash', params: {}, course: null, timer: null, quizzes: {} };
const ROLE_LABEL = { admin: 'Administrator', guru: 'Guru', siswa: 'Siswa' };

// Menu: [halaman, label, ikon, modul, grup]. Modul lain menambah menu lewat addNav().
const GROUP_ORDER = ['', 'Data Sekolah', 'Pembelajaran', 'Administrasi Guru', 'Kesiswaan', 'Ujian', 'Sarana', 'Layanan', 'Sistem'];
// grup yang tidak terdaftar ditaruh paling bawah (bukan paling atas)
const groupRank = g => { const i = GROUP_ORDER.indexOf(g || ''); return i < 0 ? GROUP_ORDER.length : i; };
const NAV = {
  admin: [['dash', 'Beranda', 'home', '', ''], ['kelas', 'Kelas', 'school', '', 'Data Sekolah'], ['siswa', 'Siswa', 'users', '', 'Data Sekolah'],
    ['guru', 'Guru', 'user', '', 'Data Sekolah'], ['mapel', 'Mata Pelajaran', 'book', '', 'Data Sekolah'],
    ['pengaturan', 'Pengaturan', 'gear', '', 'Sistem'], ['akun', 'Akun Saya', 'key', '', 'Sistem']],
  guru: [['dash', 'Beranda', 'home', '', ''], ['mapel', 'Mapel Saya', 'book', 'lms', 'Pembelajaran'], ['materi', 'Materi', 'layers', 'lms', 'Pembelajaran'],
    ['tugas', 'Tugas', 'clip', 'lms', 'Pembelajaran'], ['kuis', 'Kuis', 'quiz', 'lms', 'Pembelajaran'], ['absensi', 'Absensi Mapel', 'cal', 'lms', 'Pembelajaran'],
    ['rekap', 'Rekap Nilai', 'chart', 'lms', 'Pembelajaran'], ['akun', 'Akun Saya', 'key', '', 'Sistem']],
  siswa: [['dash', 'Beranda', 'home', '', ''], ['mapel', 'Mata Pelajaran', 'book', 'lms', 'Pembelajaran'], ['tugas', 'Tugas', 'clip', 'lms', 'Pembelajaran'],
    ['kuis', 'Kuis', 'quiz', 'lms', 'Pembelajaran'], ['nilai', 'Nilai & Kehadiran', 'chart', 'lms', 'Pembelajaran'], ['akun', 'Akun Saya', 'key', '', 'Sistem']],
};
function addNav(role, item) { NAV[role].push(item); }
const modOn = k => !k || !S.settings.modul || S.settings.modul[k] !== false;
// Guru/staf yang diberi akses kelola modul (Pengaturan akun guru) memakai halaman admin modul itu.
const isAdm = m => S.user && (S.user.role === 'admin' || (S.user.akses || []).includes(m));
function navFor(role) {
  let items = NAV[role].filter(n => modOn(n[3]));
  if (role === 'guru') {
    const extra = NAV.admin.filter(n => n[3] && isAdm(n[3]) && modOn(n[3]) && !items.some(x => x[0] === n[0]));
    items = items.concat(extra);
  }
  return items.sort((a, b) => groupRank(a[4]) - groupRank(b[4]));
}
function pageFn(role, page) {
  if (PAGES[role][page]) return PAGES[role][page];
  const parent = PAGE_PARENT[page] || page;
  const item = NAV.admin.find(n => n[0] === parent);
  if (role === 'guru' && item && item[3] && isAdm(item[3])) return PAGES.admin[page];
  return null;
}

// =====================================================================
// Komponen kecil
// =====================================================================
const head = (title, sub = '', actions = '') =>
  `<div class="page-head"><div><h1>${esc(title)}</h1>${sub ? `<p>${sub}</p>` : ''}</div>${actions ? `<div class="acts no-print">${actions}</div>` : ''}</div>`;
const tag = (t, kind = '') => `<span class="tag ${kind}">${esc(t)}</span>`;
const empty = (msg, icon = 'inbox') => `<div class="empty">${ico(icon)}<div>${esc(msg)}</div></div>`;
const errBox = e => `<div class="form-err">${esc(e.message || e)}</div>`;
const btn = (label, act, data = {}, kind = '', icon = '') =>
  `<button type="button" class="btn ${kind}" data-act="${act}" ${Object.entries(data).map(([k, v]) => `data-${k}="${esc(v)}"`).join(' ')}>${icon ? ico(icon) : ''}${label ? esc(label) : ''}</button>`;
const iconBtn = (icon, act, data, label, kind = 'ghost') =>
  `<button type="button" class="btn ${kind} sm icon" data-act="${act}" ${Object.entries(data).map(([k, v]) => `data-${k}="${esc(v)}"`).join(' ')} aria-label="${esc(label)}" title="${esc(label)}">${ico(icon)}</button>`;
const stat = (icon, n, label) => `<div class="stat"><div class="ic">${ico(icon)}</div><div><b>${esc(n)}</b><small>${esc(label)}</small></div></div>`;
// statistik uang: nilai penuh di layar lebar, ringkas (Rp 2,08 M / Rp 15,3 jt) di HP; nilai penuh tetap di tooltip
const rupiahRingkas = n => { const a = Math.abs(n || 0), f = (v, s) => 'Rp ' + v.toLocaleString('id-ID', { maximumFractionDigits: 2 }) + ' ' + s;
  return a >= 1e9 ? f(n / 1e9, 'M') : a >= 1e6 ? f(n / 1e6, 'jt') : 'Rp ' + Math.round(n || 0).toLocaleString('id-ID'); };
const statRp = (icon, n, label) => `<div class="stat" title="${esc('Rp ' + Math.round(n || 0).toLocaleString('id-ID'))}"><div class="ic">${ico(icon)}</div><div>
  <b><span class="v-full">${esc('Rp ' + Math.round(n || 0).toLocaleString('id-ID'))}</span><span class="v-short">${esc(rupiahRingkas(n))}</span></b><small>${esc(label)}</small></div></div>`;
const prog = pct => `<div class="prog" role="progressbar" aria-valuenow="${Math.round(pct)}" aria-valuemin="0" aria-valuemax="100"><i style="width:${Math.max(0, Math.min(100, pct))}%"></i></div>`;
const opts = (rows, val, blank) => (blank != null ? `<option value="">${esc(blank)}</option>` : '') +
  rows.map(([v, l]) => `<option value="${esc(v)}" ${String(v) === String(val) ? 'selected' : ''}>${esc(l)}</option>`).join('');

function toast(msg, kind = '') {
  const t = document.createElement('div');
  t.className = 'toast ' + kind;
  t.textContent = msg;
  $('#toasts').append(t);
  setTimeout(() => t.remove(), kind === 'bad' ? 5000 : 2600);
}

// =====================================================================
// Modal & formulir
// =====================================================================
const MODALS = [];
function openModal({ title, body = '', wide = false, builder = null }) {
  const el = document.createElement('div');
  el.className = 'modal';
  el.innerHTML = `<div class="sheet ${wide ? 'wide' : ''}" role="dialog" aria-modal="true" aria-label="${esc(title)}">
    <div class="sheet-head"><h2>${esc(title)}</h2>${iconBtn('x', 'closeModal', {}, 'Tutup')}</div>
    <div class="sheet-body">${body}</div></div>`;
  el.addEventListener('mousedown', e => { if (e.target === el) ACT.closeModal(); });
  $('#modals').append(el);
  // aksesibilitas (MDN aria-modal/inert): halaman di belakang tidak bisa difokus/diklik selama jendela terbuka,
  // dan fokus kembali ke tombol pembuka saat ditutup
  const m = { el, builder, body: $('.sheet-body', el), opener: document.activeElement };
  MODALS.push(m);
  $('#app').inert = true;
  if (builder) refreshModal(m);
  setTimeout(() => $('input:not([type=hidden]),select,textarea', el)?.focus(), 40);
  return m;
}
async function refreshModal(m) {
  if (!m.body.innerHTML.trim()) m.body.innerHTML = '<div class="skel"></div><div class="skel"></div><div class="skel"></div>';
  const seq = m.seq = (m.seq || 0) + 1; // filter diganti cepat: hasil permintaan lama tidak menimpa yang baru
  let html;
  try { html = await m.builder(); } catch (e) { html = errBox(e); }
  if (seq === m.seq) m.body.innerHTML = html;
}
function closeModal() {
  const m = MODALS.pop();
  if (!m) return;
  m.el.remove(); stopAllMedia();
  if (!MODALS.length) {
    $('#app').inert = false;
    if (m.opener?.isConnected) m.opener.focus({ preventScroll: true });
  }
}

// ---- Kamera: semua stream dicatat agar pasti dimatikan saat jendela ditutup atau pindah halaman
const MEDIA = new Set();
function trackStream(st) { MEDIA.add(st); return st; }
// Scanner barcode/QR USB bekerja seperti keyboard: mengetik kode lalu Enter. onScan(input, fn) memanggil fn(kode)
// lalu mengosongkan isian, siap untuk pindaian berikutnya (absensi gerbang, kasir, perpustakaan, bilik suara).
function onScan(inp, fn) {
  inp?.addEventListener('keydown', e => {
    if (e.key !== 'Enter') return;
    e.preventDefault();
    const v = inp.value.trim();
    inp.value = '';
    if (v) fn(v);
  });
}
// Satu cara membuka/menutup kamera untuk semua fitur (scan QR, selfie, buku tamu, LJK), mengikuti pola MDN:
// semua track dihentikan dan srcObject dilepas. Bila elemen video sudah hilang ketika izin kamera akhirnya
// diberikan (pengguna pindah halaman), kamera langsung dimatikan lagi.
async function openCamera(video, constraints) {
  if (!window.isSecureContext) throw new Error('Kamera butuh alamat https:// (di jaringan sekolah: port 3221).');
  if (!navigator.mediaDevices?.getUserMedia) throw new Error('Browser ini tidak mendukung kamera.');
  const st = trackStream(await navigator.mediaDevices.getUserMedia({ video: constraints, audio: false }));
  if (!video || !video.isConnected) { stopStream(st); return null; }
  video.srcObject = st;
  try { await video.play(); } catch { /* autoplay ditolak: video tetap tampil saat diketuk */ }
  return st;
}
function stopStream(st, video) {
  if (st) { st.getTracks().forEach(t => t.stop()); MEDIA.delete(st); }
  if (video) video.srcObject = null;
}
// gambar diam dari video (JPEG); null bila kamera belum menghasilkan gambar
function snapshot(video, w = 480, q = 0.75) {
  if (!video || !video.videoWidth) return null;
  const c = document.createElement('canvas');
  c.width = w; c.height = Math.round(video.videoHeight * w / video.videoWidth);
  c.getContext('2d').drawImage(video, 0, 0, c.width, c.height);
  return c.toDataURL('image/jpeg', q);
}
function stopAllMedia() {
  if (typeof Scanner !== 'undefined') Scanner.stop();
  MEDIA.forEach(st => st.getTracks().forEach(t => t.stop()));
  MEDIA.clear();
}
// Bersihkan semua data lokal milik pengguna (penting di PC lab bersama), lalu muat ulang aplikasi.
// keepAnswers: sesi habis (401) di tengah kuis/ujian → jawaban yang belum terkirim TIDAK dihapus; setelah login
// lagi siswa melanjutkan dari jawaban itu (kuncinya memuat id pengguna, jadi tidak terbaca akun lain).
function resetClient(keepAnswers = false) {
  stopAllMedia();
  const keep = k => keepAnswers && (k.startsWith('sch_quiz_') || k.startsWith('sch_cbt_'));
  try { Object.keys(localStorage).filter(k => k.startsWith('sch_') && !keep(k)).forEach(k => localStorage.removeItem(k)); } catch { /* abaikan */ }
  location.hash = '';
  location.reload();
}
// Satu fungsi CSV untuk semua ekspor: kolom selalu dikutip, dan nilai yang diawali = + - @ diberi
// tanda petik agar tidak dijalankan sebagai rumus oleh Excel (CSV injection dari data isian publik).
function toCSV(rows) {
  return rows.map(r => r.map(v => {
    let s = String(v ?? '');
    if (/^[=+\-@\t\r]/.test(s)) s = "'" + s;
    return '"' + s.replace(/"/g, '""') + '"';
  }).join(';')).join('\n');
}
async function afterSave() {
  for (const m of MODALS) if (m.builder) await refreshModal(m);
  await refresh(true);
}

const FORMS = {};
let formSeq = 0;

function fieldHtml(f, vals) {
  const v = vals[f.n] ?? f.v ?? '';
  if (typeof f.o === 'function') f = { ...f, o: f.o() };
  if (f.t === 'check') return `<label class="chk"><input type="checkbox" name="${f.n}" ${v ? 'checked' : ''}> ${esc(f.l)}</label>`;
  if (f.t === 'html') return f.html;
  if (f.t === 'checks') {
    const sel = (Array.isArray(v) ? v : []).map(String);
    return `<div class="f"><span style="font-size:13px;font-weight:600;color:var(--text-2)">${esc(f.l)}</span><div class="checks">${(f.o || []).map(([ov, ol]) =>
      `<label class="chk"><input type="checkbox" name="${f.n}" value="${esc(ov)}" ${sel.includes(String(ov)) ? 'checked' : ''}> ${esc(ol)}</label>`).join('')}</div></div>`;
  }
  if (f.t === 'file') {
    const isImg = f.accept !== 'application/pdf';
    return `<div class="f"><span style="font-size:13px;font-weight:600;color:var(--text-2)">${esc(f.l)} <small>(opsional)</small></span>
      <div class="upl"><input type="hidden" name="${f.n}" value="${esc(v)}">
        <div class="upl-prev">${v ? (isImg ? `<img src="${esc(v)}" alt="">` : `<a href="${esc(v)}" target="_blank" rel="noopener">${ico('file')} Lihat file</a>`) : ''}</div>
        <label class="btn ghost sm">${ico('upload')} Pilih file<input type="file" accept="${esc(f.accept || 'image/*')}" data-change="fileUp" hidden></label>
        ${v ? `<button type="button" class="btn ghost sm" data-act="fileClear">Hapus</button>` : ''}</div>${f.hint ? `<small class="muted">${esc(f.hint)}</small>` : ''}</div>`;
  }
  let inp;
  const req = f.req ? 'required' : '';
  if (f.t === 'textarea') inp = `<textarea name="${f.n}" ${req} rows="${f.rows || 4}" placeholder="${esc(f.ph || '')}">${esc(v)}</textarea>`;
  else if (f.t === 'select') {
    // nilai lama yang tidak ada di daftar tetap ditampilkan agar tidak hilang saat disimpan ulang
    const o = f.o || [];
    const keep = v !== '' && v != null && !o.some(([ov]) => String(ov) === String(v)) ? [[v, v]] : [];
    inp = `<select name="${f.n}" ${req}>${opts([...keep, ...o], v, f.blank)}</select>`;
  } else if (f.t === 'combo') {
    // dropdown pilihan + tetap boleh mengetik nilai baru (lihat Combo di bawah)
    inp = comboHtml(f.n, v, f.o || [], { req, ph: f.ph });
  }
  else {
    const type = { number: 'number', date: 'date', datetime: 'datetime-local', time: 'time', password: 'password', email: 'email', url: 'url' }[f.t] || 'text';
    const val = f.t === 'datetime' ? toLocalInput(v) : v;
    const extra = [f.min != null && `min="${f.min}"`, f.max != null && `max="${f.max}"`, f.step && `step="${f.step}"`, f.minlength && `minlength="${f.minlength}"`].filter(Boolean).join(' ');
    inp = `<input type="${type}" name="${f.n}" value="${esc(val)}" ${req} ${extra} placeholder="${esc(f.ph || '')}" autocomplete="${f.ac || 'off'}">`;
  }
  return `<label class="f"><span>${esc(f.l)}${f.req || f.t === 'select' ? '' : ' <small>(opsional)</small>'}</span>${inp}${f.hint ? `<small>${esc(f.hint)}</small>` : ''}</label>`;
}

// fields: daftar field; elemen berupa array = dua field berdampingan
function formHtml(id, fields, vals = {}, { submit = 'Simpan', cancel = true } = {}) {
  const rows = fields.map(f => Array.isArray(f) ? `<div class="row2">${f.map(x => fieldHtml(x, vals)).join('')}</div>` : fieldHtml(f, vals)).join('');
  return `<form class="form" data-form="${id}">${rows}<div class="form-err" role="alert"></div>
    <div class="sheet-foot">${cancel ? '<button type="button" class="btn ghost" data-act="closeModal">Batal</button>' : ''}<button class="btn" type="submit">${esc(submit)}</button></div></form>`;
}

// =====================================================================
// Combo: dropdown "pilih atau ketik". Pengganti <datalist> yang di HP tidak menampilkan daftar
// (iPhone hanya saran di atas keyboard, Android hanya yang cocok dengan isian). Pola ARIA combobox:
// ketuk/fokus = buka semua pilihan, ketik = saring, panah atas/bawah + Enter = pilih, Esc = tutup.
// =====================================================================
const COMBO = {};
let comboSeq = 0;
function comboItems(list) {
  const lbl = new Map(list.filter(Array.isArray).map(([ov, ol]) => [String(ov), ol]));
  return uniqList(list).map(v => ({ v, l: lbl.get(v) && lbl.get(v) !== v ? lbl.get(v) : '' }));
}
function comboHtml(name, val, list, { req = '', ph = '' } = {}) {
  const id = 'cb' + (++comboSeq);
  COMBO[id] = comboItems(list);
  return `<div class="cb" data-cb="${id}"><input name="${name}" value="${esc(val ?? '')}" ${req} placeholder="${esc(ph || 'Pilih atau ketik…')}" autocomplete="off"
      role="combobox" aria-autocomplete="list" aria-expanded="false" aria-controls="${id}-l">
    <button type="button" class="cb-btn" tabindex="-1" aria-label="Tampilkan pilihan">${ico('down')}</button>
    <ul class="cb-list" id="${id}-l" role="listbox" hidden></ul></div>`;
}
// ganti daftar pilihan combo yang sudah tampil (mis. kategori bergantung jenis)
function setComboOptions(input, list) { const w = input.closest('.cb'); if (w) COMBO[w.dataset.cb] = comboItems(list); }
const Combo = {
  open(w, filter) {
    const inp = $('input', w), ul = $('.cb-list', w), items = COMBO[w.dataset.cb] || [];
    const q = (filter ? inp.value : '').trim().toLowerCase();
    const cur = inp.value.trim().toLowerCase();
    const shown = q ? items.filter(x => (x.v + ' ' + x.l).toLowerCase().includes(q)) : items;
    ul.innerHTML = shown.length ? shown.map((x, i) => `<li role="option" id="${w.dataset.cb}-o${i}" data-v="${esc(x.v)}" aria-selected="${x.v.toLowerCase() === cur}">
        <span>${esc(x.v)}</span>${x.l ? `<small>${esc(x.l)}</small>` : ''}</li>`).join('')
      : `<li class="cb-none" aria-disabled="true">${q ? `Tidak ada di daftar — "${esc(inp.value.trim())}" akan disimpan sebagai isian baru` : 'Belum ada pilihan, ketik isian'}</li>`;
    const wasHidden = ul.hidden;
    ul.hidden = false; w.classList.add('open'); inp.setAttribute('aria-expanded', 'true');
    const sel = $('[aria-selected=true]', ul);
    if (sel) { Combo.hi(w, sel); sel.scrollIntoView({ block: 'nearest' }); } else inp.removeAttribute('aria-activedescendant');
    // di HP: pastikan daftar terlihat di atas keyboard — hanya saat baru dibuka, tidak setiap ketikan (layar bergetar)
    if (wasHidden && matchMedia('(max-width: 640px)').matches) setTimeout(() => inp.scrollIntoView({ block: 'start', behavior: 'smooth' }), 60);
  },
  close(w) {
    const ul = $('.cb-list', w);
    if (ul.hidden) return;
    ul.hidden = true; w.classList.remove('open');
    const inp = $('input', w); inp.setAttribute('aria-expanded', 'false'); inp.removeAttribute('aria-activedescendant');
  },
  hi(w, li) {
    $$('.cb-list li.hi', w).forEach(x => x.classList.remove('hi'));
    if (!li) return;
    li.classList.add('hi'); li.scrollIntoView({ block: 'nearest' });
    $('input', w).setAttribute('aria-activedescendant', li.id);
  },
  pick(w, li) {
    const inp = $('input', w);
    inp.value = li.dataset.v;
    Combo.close(w);
    inp.dispatchEvent(new Event('input', { bubbles: true }));
    inp.dispatchEvent(new Event('change', { bubbles: true }));
  },
};
// dibuka saat diketuk/diklik, diketik, atau panah bawah; tidak saat fokus otomatis (formulir baru dibuka)
document.addEventListener('click', e => { if (e.target.matches && e.target.matches('.cb input') && $('.cb-list', e.target.closest('.cb')).hidden) Combo.open(e.target.closest('.cb'), false); });
document.addEventListener('focusout', e => {
  const w = e.target.closest && e.target.closest('.cb');
  if (w) setTimeout(() => { if (!w.contains(document.activeElement)) Combo.close(w); }, 120);
});
document.addEventListener('input', e => { if (e.isTrusted !== false && e.target.matches && e.target.matches('.cb input')) Combo.open(e.target.closest('.cb'), true); });
// pilih dengan ketuk/klik; mousedown dicegah agar input tetap fokus (keyboard HP tidak naik-turun)
document.addEventListener('mousedown', e => { if (e.target.closest && e.target.closest('.cb-list, .cb-btn')) e.preventDefault(); });
document.addEventListener('click', e => {
  // preventDefault: combo ada di dalam <label>; tanpa ini klik diteruskan ke input dan daftar terbuka lagi
  const li = e.target.closest && e.target.closest('.cb-list li');
  if (li) { e.preventDefault(); if (li.matches('[role=option]')) Combo.pick(li.closest('.cb'), li); return; }
  const b = e.target.closest && e.target.closest('.cb-btn');
  if (b) { e.preventDefault(); const w = b.closest('.cb'); const inp = $('input', w); if ($('.cb-list', w).hidden) { inp.focus(); Combo.open(w, false); } else Combo.close(w); }
});
document.addEventListener('keydown', e => {
  if (!e.target.matches || !e.target.matches('.cb input')) return;
  const w = e.target.closest('.cb'), ul = $('.cb-list', w);
  const opts = $$('li[role=option]', ul), cur = $('li.hi', ul), i = opts.indexOf(cur);
  if (e.key === 'ArrowDown' || e.key === 'ArrowUp') {
    e.preventDefault();
    if (ul.hidden) return Combo.open(w, false);
    Combo.hi(w, opts[Math.min(opts.length - 1, Math.max(0, i + (e.key === 'ArrowDown' ? 1 : -1)))] || opts[0]);
  } else if (e.key === 'Enter' && !ul.hidden && cur) { e.preventDefault(); Combo.pick(w, cur); }
  else if (e.key === 'Escape' && !ul.hidden) { e.preventDefault(); e.stopPropagation(); Combo.close(w); }
  else if (e.key === 'Tab') Combo.close(w);
}, true);
// daftar unik tanpa kosong (urutan dipertahankan, tanpa beda huruf besar/kecil)
function uniqList(list) {
  const seen = new Set(), out = [];
  for (const x of list.map(x => Array.isArray(x) ? x[0] : x)) {
    const s = String(x ?? '').trim(), k = s.toLowerCase();
    if (s && !seen.has(k)) { seen.add(k); out.push(s); }
  }
  return out;
}
// saran isian dari data yang sudah ada (GET /api/saran/<kunci>) digabung dengan pilihan bawaan
async function saran(key, bawaan = []) {
  let ada = [];
  try { ada = await GET('/saran/' + key); } catch { /* tanpa saran tetap bisa mengetik */ }
  return uniqList([...bawaan, ...ada]);
}
// mata pelajaran: yang sudah ada di data + daftar umum Kurikulum Merdeka
// [kode, nama] mata pelajaran umum Kurikulum Merdeka
const MAPEL_STD = [['PAI', 'Pendidikan Agama Islam dan Budi Pekerti'], ['PAK', 'Pendidikan Agama Kristen dan Budi Pekerti'], ['PKAT', 'Pendidikan Agama Katolik dan Budi Pekerti'],
  ['PAH', 'Pendidikan Agama Hindu dan Budi Pekerti'], ['PAB', 'Pendidikan Agama Buddha dan Budi Pekerti'], ['PPKN', 'Pendidikan Pancasila'], ['BIN', 'Bahasa Indonesia'],
  ['MTK', 'Matematika'], ['IPA', 'Ilmu Pengetahuan Alam'], ['IPS', 'Ilmu Pengetahuan Sosial'], ['IPAS', 'Ilmu Pengetahuan Alam dan Sosial'], ['BIG', 'Bahasa Inggris'],
  ['PJOK', 'Pendidikan Jasmani, Olahraga, dan Kesehatan'], ['SBD', 'Seni Budaya'], ['SR', 'Seni Rupa'], ['SM', 'Seni Musik'], ['INF', 'Informatika'], ['PKY', 'Prakarya'],
  ['BDR', 'Bahasa Daerah'], ['BAR', 'Bahasa Arab'], ['FIS', 'Fisika'], ['KIM', 'Kimia'], ['BIO', 'Biologi'], ['EKO', 'Ekonomi'], ['GEO', 'Geografi'], ['SOS', 'Sosiologi'],
  ['SEJ', 'Sejarah'], ['BK', 'Bimbingan Konseling']];
const MAPEL_UMUM = MAPEL_STD.map(([, n]) => n);
// daftar mapel: untuk guru, mapel yang diampunya di urutan atas (berlabel); mapel lain tetap bisa dipilih/diketik
async function mapelOpts() {
  let cs = [];
  try { cs = uniqList((await GET('/courses')).map(c => c.nama)); } catch { /* tetap pakai daftar umum */ }
  const guru = S.user && S.user.role === 'guru';
  S.mapelAmpu = guru ? cs : [];
  return [...(guru ? cs.map(n => [n, 'Mapel yang Anda ampu']) : cs), ...MAPEL_UMUM.filter(n => !cs.some(c => c.toLowerCase() === n.toLowerCase()))];
}
// guru yang hanya mengampu satu mapel: kolom mapel langsung terisi
const mapelAwal = () => (S.mapelAmpu || []).length === 1 ? S.mapelAmpu[0] : '';
async function kelasOpts() {
  try { return uniqList((await GET('/classes')).map(k => k.nama)); } catch { return []; }
}
const TINGKAT_OPTS = Array.from({ length: 12 }, (_, i) => [String(i + 1), 'Tingkat ' + (i + 1)]);
// pilihan tahun (mis. tahun perolehan/terbit): tahun ini mundur n tahun
const yearOpts = (n = 40) => Array.from({ length: n }, (_, i) => String(new Date().getFullYear() - i));
// tahun ajaran: dari daftar tahun ajaran + tahun depan
async function tahunAjaranOpts() {
  let ys = [];
  try { ys = (await GET('/tahun')).map(y => y.nama); } catch { /* guru tanpa akses */ }
  const y = new Date().getFullYear();
  return uniqList([...ys, `${y - 1}/${y}`, `${y}/${y + 1}`, `${y + 1}/${y + 2}`]).sort().reverse();
}

function registerForm(fields, onSubmit) {
  const id = 'f' + (++formSeq);
  FORMS[id] = { fields: fields.flat(), onSubmit };
  return id;
}

function openForm({ title, fields, values = {}, onSubmit, submit, intro = '', wide = false, success = 'Tersimpan.' }) {
  const id = registerForm(fields, async (vals, form) => { await onSubmit(vals, form); closeModal(); toast(success); await afterSave(); });
  openModal({ title, wide, body: intro + formHtml(id, fields, values, { submit }) });
}

function collect(form, fields) {
  const out = {};
  for (const f of fields) {
    if (!f.n) continue;
    const el = form.elements[f.n];
    if (!el) continue;
    if (f.t === 'checks') out[f.n] = $$(`input[name="${f.n}"]:checked`, form).map(i => i.value);
    else if (f.t === 'check') out[f.n] = el.checked;
    else if (f.t === 'number') out[f.n] = num(el.value);
    else if (f.t === 'datetime') out[f.n] = el.value ? new Date(el.value).toISOString() : null;
    else if (f.t === 'password') out[f.n] = el.value;
    else out[f.n] = el.value.trim();
  }
  return out;
}

document.addEventListener('submit', async e => {
  const form = e.target.closest('form[data-form]');
  if (!form) return;
  e.preventDefault();
  const def = FORMS[form.dataset.form];
  if (!def) return;
  const err = $('.form-err', form);
  err.textContent = '';
  const bad = $$('input,select,textarea', form).find(el => !el.checkValidity());
  if (bad) { err.textContent = bad.validationMessage || 'Periksa kembali isian.'; bad.focus(); return; }
  const b = $('button[type=submit]', form);
  b.disabled = true;
  try { await def.onSubmit(collect(form, def.fields), form); }
  catch (ex) { err.textContent = ex.message; }
  finally { b.disabled = false; }
});

let confirmResolve = null;
function confirmBox(msg, ok = 'Hapus', kind = 'danger') {
  return new Promise(res => {
    confirmResolve = res;
    openModal({ title: 'Konfirmasi', body: `<p style="margin-bottom:18px">${msg}</p><div class="sheet-foot">
      <button type="button" class="btn ghost" data-act="confirmNo">Batal</button>
      <button type="button" class="btn ${kind}" data-act="confirmYes">${esc(ok)}</button></div>` });
  });
}

async function del(path, msg, success = 'Terhapus.') {
  if (!(await confirmBox(msg))) return;
  try { await DEL(path); toast(success); await afterSave(); } catch (e) { toast(e.message, 'bad'); }
}

// =====================================================================
// Navigasi
// =====================================================================
// query string dari objek; nilai kosong/null dilewati
const qs = o => new URLSearchParams(Object.entries(o).filter(([, v]) => v != null && v !== '')).toString();
function go(page, params = {}) {
  const q = qs(params);
  const h = '#/' + page + (q ? '?' + q : '');
  if (location.hash === h) route(); else location.hash = h;
}
function parseHash() {
  const m = location.hash.match(/^#\/([\w-]+)(?:\?(.*))?$/);
  if (!m) return { page: 'dash', params: {} };
  return { page: m[1], params: Object.fromEntries(new URLSearchParams(m[2] || '')) };
}
window.addEventListener('hashchange', () => { if (S.user) route(); });

async function route() {
  if (S.timer) { clearInterval(S.timer); S.timer = null; }
  if (S.navCtl) S.navCtl.abort();
  S.navCtl = new AbortController();
  GET_INFLIGHT.clear(); // halaman baru tidak ikut menunggu permintaan lama yang barusan dibatalkan
  if (confirmResolve) { const r = confirmResolve; confirmResolve = null; r(false); } // dialog konfirmasi ditinggal = "batal"
  while (MODALS.length) closeModal();
  stopAllMedia();
  for (const id of Object.keys(FORMS)) if (!document.querySelector(`[data-form="${id}"]`)) delete FORMS[id];
  for (const id of Object.keys(COMBO)) if (!document.querySelector(`[data-cb="${id}"]`)) delete COMBO[id];
  if (S.user && S.user.ganti_pw) return renderGantiPW();
  const { page, params } = parseHash();
  const navItem = navFor(S.user.role).find(n => n[0] === (PAGE_PARENT[page] || page)) || NAV[S.user.role].find(n => n[0] === (PAGE_PARENT[page] || page));
  S.page = pageFn(S.user.role, page) && (!navItem || modOn(navItem[3])) ? page : 'dash';
  S.params = params;
  if (!$('#shell')) renderShell();
  const active = PAGE_PARENT[S.page] || S.page;
  $$('.nav button').forEach(b => b.classList.toggle('on', b.dataset.page === active));
  $('#shell').classList.remove('menu-open');
  window.scrollTo(0, 0);
  await refresh(false);
}

async function refresh(quiet) {
  const main = $('#main');
  if (!main) return;
  const page = S.page;
  const seq = S.seq = (S.seq || 0) + 1;
  // kerangka loading hanya muncul bila halaman butuh >150 ms; yang cepat langsung berganti tanpa kedip
  if (!quiet) setTimeout(() => { if (seq === S.seq && main.dataset.seq !== String(seq)) main.innerHTML = '<div class="skel" style="width:40%;height:26px"></div><div class="skel"></div><div class="skel"></div><div class="skel" style="width:70%"></div>'; }, 150);
  S.after = null;
  try {
    // ambil S.after tepat saat halaman ini selesai dibuat, supaya halaman lama yang selesai belakangan
    // tidak menimpa/menjalankan "after" milik halaman baru
    const [html, after] = await pageFn(S.user.role, page)(S.params).then(h => { const a = S.after; S.after = null; return [h, a]; });
    if (seq !== S.seq || page !== S.page) return;
    // isian saring cepat ([data-filter]) dipertahankan saat halaman yang sama dimuat ulang (mis. setelah simpan)
    // hanya untuk halaman + tab yang sama (Surat masuk/keluar memakai id saringan yang sama)
    const view = page + '|' + (S.params.tab || '');
    const keepFilter = main.dataset.view === view ? $$('[data-filter]', main).map(i => [i.dataset.filter, i.value]).filter(([, v]) => v) : [];
    main.dataset.seq = seq;
    main.dataset.view = view;
    main.innerHTML = html;
    for (const [id, v] of keepFilter) {
      const i = $(`[data-filter="${id}"]`, main);
      if (i) { i.value = v; i.dispatchEvent(new Event('input', { bubbles: true })); }
    }
    if (after) after();
  } catch (e) {
    if (seq !== S.seq) return;
    main.dataset.seq = seq;
    main.innerHTML = head('Terjadi kendala') + errBox(e);
  }
}

function brandHtml(sub) {
  const s = S.settings;
  const initial = (s.nama_sekolah || 'S').trim().charAt(0).toUpperCase();
  return `<div class="brand"><div class="logo">${s.logo_url ? `<img src="${esc(logoSrc(s))}" alt="">` : esc(initial)}</div>
    <div><b>${esc(s.nama_sekolah || 'Sekolah')}</b><small>${esc(sub)}</small></div></div>`;
}

function renderShell() {
  const u = S.user;
  $('#app').innerHTML = `<div class="shell" id="shell">
    <aside class="side" aria-label="Menu utama">
      ${brandHtml(ROLE_LABEL[u.role])}
      <nav class="nav">${(() => { let g = null; return navFor(u.role).map(([p, l, i, , grp]) => {
        const h = grp !== g && grp ? `<div class="nav-group">${esc(grp)}</div>` : ''; g = grp;
        return h + `<button type="button" data-act="nav" data-page="${p}">${ico(i)}${esc(l)}</button>`; }).join(''); })()}</nav>
      <div class="who"><div class="who-row"><div><b>${esc(u.nama_lengkap)}</b><small>${esc(u.username)}</small></div>${themeBtn()}</div>
        <button type="button" class="btn ghost sm" data-act="logout">${ico('out')} Keluar</button>
        <small class="muted" style="font-size:11px">SCH ${esc(S.settings.versi || '')}</small></div>
    </aside>
    <div class="scrim" data-act="closeMenu"></div>
    <div><div class="topbar">${iconBtn('menu', 'openMenu', {}, 'Buka menu')}<b>${esc(S.settings.nama_sekolah || 'Sekolah')}</b>${themeBtn()}</div>
      <main class="main" id="main"></main></div>
  </div>`;
  document.title = (S.settings.nama_sekolah || 'Sekolah') + ' · LMS';
  document.body.classList.add('has-shell');
}

// ---- Tema: ikut sistem -> terang -> gelap. Disimpan di perangkat (bukan kunci sch_, jadi tidak hilang saat logout).
const THEMES = { system: ['monitor', 'Tema: ikut sistem'], light: ['sun', 'Tema: terang'], dark: ['moon', 'Tema: gelap'] };
// Warna tema sekolah (per jenjang). Disimpan juga di browser agar halaman berikutnya langsung berwarna benar.
const WARNA = { biru: ['Biru', 'SMP', '#1d4ed8'], abu: ['Abu-abu', 'SMA / SMK', '#475569'], merah: ['Merah', 'SD', '#dc2626'] };
function applyWarna(w) {
  if (!WARNA[w]) w = 'biru';
  if (w === 'biru') document.documentElement.removeAttribute('data-warna'); else document.documentElement.setAttribute('data-warna', w);
  try { localStorage.setItem('theme_warna', w); } catch { /* abaikan */ }
  const c = WARNA[w][2];
  document.querySelector('meta[name=theme-color]')?.setAttribute('content', c);
  const ic = document.querySelector('link[rel=icon]');
  if (ic) ic.href = ic.href.replace(/fill='%23[0-9a-fA-F]{6}'/, `fill='%23${c.slice(1)}'`);
}
function themePref() { try { return localStorage.getItem('theme_pref') || 'system'; } catch { return 'system'; } }
function applyTheme(t) {
  if (t === 'light' || t === 'dark') document.documentElement.setAttribute('data-theme', t);
  else document.documentElement.removeAttribute('data-theme');
  try { if (t === 'system') localStorage.removeItem('theme_pref'); else localStorage.setItem('theme_pref', t); } catch { /* abaikan */ }
  $$('.theme-btn').forEach(b => { b.innerHTML = ico(THEMES[t][0]); b.title = THEMES[t][1]; b.setAttribute('aria-label', THEMES[t][1] + '. Klik untuk mengganti.'); });
}
const themeBtn = (extra = '') => { const [i, l] = THEMES[themePref()] || THEMES.system; return `<button type="button" class="theme-btn ${extra}" data-act="theme" title="${l}" aria-label="${l}. Klik untuk mengganti.">${ico(i)}</button>`; };
function cycleTheme() {
  const order = ['system', 'light', 'dark'];
  const next = order[(order.indexOf(themePref()) + 1) % order.length];
  applyTheme(next);
  toast(THEMES[next][1].replace('Tema: ', 'Tema ') + '.');
};
// tombol tema melayang untuk halaman tanpa menu (login & halaman publik)
function ensureThemeFloat() {
  if (!$('#themeFloat')) document.body.insertAdjacentHTML('beforeend', themeBtn('theme-float').replace('<button', '<button id="themeFloat"'));
}

function renderLogin() {
  document.body.classList.remove('has-shell');
  if (S.timer) { clearInterval(S.timer); S.timer = null; }
  while (MODALS.length) closeModal();
  const id = registerForm([{ n: 'username' }, { n: 'password', t: 'password' }], async v => {
    const u = await POST('/login', { username: v.username, password: v.password });
    S.user = u;
    await loadModules();
    route();
  });
  $('#app').innerHTML = `<div class="login-wrap"><div style="width:100%;display:grid;place-items:center">
    <div class="login-card">
      <div class="login-form">
        ${brandHtml('Sistem Pembelajaran Sekolah')}
        <h1>Selamat Datang</h1><p class="sub">Silakan masuk dengan akun dari sekolah Anda.</p>
        <form class="form" data-form="${id}">
          <label class="inp-ic"><span class="sr" style="position:absolute;left:-9999px">Username atau NIS</span>${ico('user')}<input name="username" required autocomplete="username" autocapitalize="off" placeholder="Username / NIS"></label>
          <label class="inp-ic"><span class="sr" style="position:absolute;left:-9999px">Password</span>${ico('key')}<input name="password" type="password" required autocomplete="current-password" placeholder="Password"></label>
          <div class="form-err" role="alert"></div>
          <button class="btn" type="submit">Masuk</button>
        </form>
      </div>
      <div class="login-art" aria-hidden="true"><div class="art-frame">${LOGIN_ART}</div></div>
      <div class="badge-bolt" aria-hidden="true"><svg viewBox="0 0 24 24"><path d="M13 2 4 14h7l-1 8 9-12h-7z"/></svg></div>
    </div>
    <p class="login-note">Lupa password? Hubungi admin sekolah.</p>
    <p class="login-note" style="margin-top:6px">${[['ppdb', '#/ppdb', 'Daftar PPDB'], ['kelulusan', '#/kelulusan', 'Cek kelulusan'], ['bukutamu', '#/bukutamu', 'Buku tamu']].filter(([m]) => modOn(m)).map(([, h, l]) => `<a href="${h}">${l}</a>`).join(' · ')}</p>
  </div></div>`;
  document.title = (S.settings.nama_sekolah || 'Sekolah') + ' · Masuk';
  setTimeout(() => $('input[name=username]')?.focus(), 30);
}

// =====================================================================
// Aksi (event delegation)
// =====================================================================
const ACT = {
  nav: el => go(el.dataset.page),
  go: el => go(el.dataset.page, JSON.parse(el.dataset.params || '{}')),
  openMenu: () => $('#shell').classList.add('menu-open'),
  closeMenu: () => $('#shell').classList.remove('menu-open'),
  closeModal: () => { if (confirmResolve) { confirmResolve(false); confirmResolve = null; } closeModal(); },
  confirmYes: () => { const r = confirmResolve; confirmResolve = null; closeModal(); r && r(true); },
  confirmNo: () => { const r = confirmResolve; confirmResolve = null; closeModal(); r && r(false); },
  async logout() {
    if ((S.page === 'ujian' || S.page === 'cbtujian') && !(await confirmBox('Ujian/kuis sedang berjalan. Jawaban yang belum dikirim akan hilang dari perangkat ini. Tetap keluar?', 'Keluar'))) return;
    try { await POST('/logout'); } catch { /* abaikan */ }
    resetClient();
  },
  print: () => window.print(),
  theme: () => cycleTheme(),
};
const CHG = {};

document.addEventListener('click', e => {
  const el = e.target.closest('[data-act]');
  if (!el) return;
  const fn = ACT[el.dataset.act];
  if (!fn) return;
  e.preventDefault();
  // klik ganda cepat tidak menjalankan aksi dua kali (pinjam/kembali buku, kirim ulang WA, buka jendela, dll.)
  if (BUSY.has(el)) return;
  BUSY.add(el);
  Promise.resolve().then(() => fn(el, e)).catch(ex => { if (!isAbort(ex)) toast(ex.message, 'bad'); }).finally(() => BUSY.delete(el));
});
const BUSY = new WeakSet();
document.addEventListener('change', e => {
  const el = e.target.closest('[data-change]');
  if (el && CHG[el.dataset.change]) Promise.resolve(CHG[el.dataset.change](el)).catch(ex => { if (!isAbort(ex)) toast(ex.message, 'bad'); });
});
document.addEventListener('input', e => {
  const el = e.target.closest('[data-filter]');
  if (!el) return;
  const q = el.value.trim().toLowerCase();
  $$('#' + el.dataset.filter + ' tbody tr').forEach(tr => { tr.hidden = q && !tr.textContent.toLowerCase().includes(q); });
});
document.addEventListener('keydown', e => {
  if (e.key === 'Escape' && MODALS.length) ACT.closeModal();
  // kartu/daftar yang bisa diklik juga bisa dipilih dengan Enter/Spasi (aksesibilitas keyboard)
  if ((e.key === 'Enter' || e.key === ' ') && e.target.matches('[data-act][tabindex], [data-act][role=button]') && e.target.tagName !== 'BUTTON') {
    e.preventDefault(); e.target.click();
  }
});

// =====================================================================
// Perhitungan nilai
// =====================================================================
// tugas & kuis dalam skala 0-100; nilai akhir = 40% tugas + 40% kuis + 20% kehadiran (dari komponen yang ada).
// Tugas yang tidak dikumpulkan sampai lewat tenggat dihitung 0. Nilai akhir hanya muncul jika ada nilai tugas/kuis.
const RUMUS = 'Nilai akhir = 40% tugas + 40% kuis + 20% kehadiran. Tugas yang tidak dikumpulkan sampai lewat tenggat dihitung 0; komponen yang belum ada tidak dihitung.';
function calcScore(assignments, subByAssignment, attempts, attendance) {
  const now = new Date();
  const t = assignments.map(a => {
    const s = subByAssignment[a.id];
    if (s && s.nilai != null) return (s.nilai / (a.nilai_maks || 100)) * 100;
    if (!s && a.deadline && new Date(a.deadline) < now) return 0;
    return null;
  }).filter(v => v != null);
  const tugas = t.length ? t.reduce((x, y) => x + y, 0) / t.length : null;
  const k = attempts.map(a => a.skor);
  const kuis = k.length ? k.reduce((x, y) => x + y, 0) / k.length : null;
  const hadir = attendance.length ? attendance.filter(a => a.status === 'hadir').length / attendance.length * 100 : null;
  const parts = [[tugas, 40], [kuis, 40], [hadir, 20]].filter(([v]) => v != null);
  const w = parts.reduce((x, [, b]) => x + b, 0);
  const akhir = w && (tugas != null || kuis != null) ? parts.reduce((x, [v, b]) => x + v * b, 0) / w : null;
  return { tugas, kuis, hadir, akhir };
}
const predikat = n => n == null ? ['–', ''] : n >= 90 ? ['A', 'ok'] : n >= 80 ? ['B', 'ok'] : n >= 70 ? ['C', 'info'] : n >= 60 ? ['D', 'warn'] : ['E', 'bad'];

// Lampiran materi: satu per baris "Judul | URL" atau hanya URL
function lampType(url) {
  if (/(youtube\.com\/|youtu\.be\/)/i.test(url)) return 'youtube';
  if (/\.(png|jpe?g|gif|webp|svg)(\?|$)/i.test(url)) return 'gambar';
  if (/\.(pdf|docx?|pptx?|xlsx?)(\?|$)|drive\.google|docs\.google/i.test(url)) return 'dokumen';
  return 'link';
}
function parseLamp(text) {
  return String(text || '').split('\n').map(l => l.trim()).filter(Boolean).map(l => {
    const i = l.lastIndexOf('|');
    const judul = i >= 0 ? l.slice(0, i).trim() : '';
    const isi = (i >= 0 ? l.slice(i + 1) : l).trim();
    return { tipe: lampType(isi), judul, isi };
  });
}
const lampText = arr => (arr || []).map(l => l.judul ? `${l.judul} | ${l.isi}` : l.isi).join('\n');
function ytId(u) { const m = String(u).match(/(?:youtu\.be\/|v=|embed\/|shorts\/)([\w-]{11})/); return m ? m[1] : null; }

function lampHtml(l) {
  const url = safeUrl(l.isi);
  if (l.tipe === 'youtube' && ytId(l.isi)) {
    return `<div><div class="embed"><iframe src="https://www.youtube-nocookie.com/embed/${ytId(l.isi)}" title="${esc(l.judul || 'Video')}" loading="lazy" allowfullscreen allow="encrypted-media; picture-in-picture"></iframe></div>
      <small><a href="${esc(url)}" target="_blank" rel="noopener">Buka di YouTube</a> · Video butuh internet</small></div>`;
  }
  if (l.tipe === 'gambar') return `<a href="${esc(url)}" target="_blank" rel="noopener"><img class="pic" src="${esc(url)}" alt="${esc(l.judul || 'Gambar materi')}" loading="lazy"></a>`;
  const i = l.tipe === 'dokumen' ? 'file' : 'link';
  return `<a class="lamp" href="${esc(url)}" target="_blank" rel="noopener">${ico(i)}<span>${esc(l.judul || l.isi)}</span></a>`;
}

function matHtml(m, i, { done = false, toggle = false, actions = '' } = {}) {
  const lamp = m.lampiran || [];
  return `<details class="mat"><summary><span class="num ${done ? 'done' : ''}">${done ? ico('check') : i + 1}</span>
      <div style="flex:1;min-width:0"><b>${esc(m.judul)}</b><small class="muted">${lamp.length ? lamp.length + ' lampiran · ' : ''}${fmtDate(m.created_at)}</small></div>
      ${actions}</summary>
    <div class="mat-body">${m.konten ? `<div class="mat-text">${esc(m.konten)}</div>` : ''}${lamp.map(lampHtml).join('')}
      ${toggle ? `<div>${done ? btn('Tandai belum selesai', 'matDone', { id: m.id, v: 0 }, 'ghost sm') : btn('Tandai selesai dipelajari', 'matDone', { id: m.id, v: 1 }, 'sm', 'check')}</div>` : ''}
    </div></details>`;
}

// =====================================================================
// Daftar halaman per peran
// =====================================================================
// =====================================================================
// Helper untuk modul: upload, cetak, pemindai QR
// =====================================================================
async function uploadFile(file) {
  const fd = new FormData();
  fd.append('file', file);
  let res;
  try { res = await fetch('/api/upload', { method: 'POST', body: fd, headers: { 'X-SCH': '1' }, credentials: 'same-origin' }); }
  catch { throw new Error('Gagal mengunggah. Periksa koneksi.'); }
  const d = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(d.error || 'Gagal mengunggah.');
  return d.url;
}
CHG.fileUp = async inp => {
  const f = inp.files[0];
  if (!f) return;
  const box = inp.closest('.upl');
  const prev = $('.upl-prev', box);
  prev.innerHTML = '<span class="muted">Mengunggah…</span>';
  try {
    const url = await uploadFile(f);
    $('input[type=hidden]', box).value = url;
    prev.innerHTML = f.type.startsWith('image/') ? `<img src="${esc(url)}" alt="">` : `<a href="${esc(url)}" target="_blank" rel="noopener">${ico('file')} ${esc(f.name)}</a>`;
  } catch (e) { prev.innerHTML = ''; toast(e.message, 'bad'); }
};
ACT.fileClear = el => { const box = el.closest('.upl'); $('input[type=hidden]', box).value = ''; $('.upl-prev', box).innerHTML = ''; el.remove(); };

// Helper bersama antarmodul (dipindah ke inti agar tidak bergantung urutan file)
const avatar = (foto, nama, size = 40) => foto
  ? `<img class="avatar" src="${esc(foto)}" alt="" style="width:${size}px;height:${size}px">`
  : `<span class="avatar" style="width:${size}px;height:${size}px">${esc((nama || '?').trim().charAt(0).toUpperCase())}</span>`;
const rupiah = n => 'Rp ' + Math.round(n || 0).toLocaleString('id-ID');
function parseSoalText(text) {
  const out = []; let cur = null;
  for (const raw of text.split(/\r?\n/)) {
    const l = raw.trim();
    if (!l) continue;
    let m;
    if ((m = l.match(/^(\d+)[.)]\s*(.+)$/))) { cur = { pertanyaan: m[2], opsi: {}, kunci: '' }; out.push(cur); }
    else if (cur && (m = l.match(/^([A-Ea-e])[.)]\s*(.+)$/))) cur.opsi[m[1].toLowerCase()] = m[2];
    else if (cur && (m = l.match(/^kunci\s*(?:jawaban)?\s*[:=]\s*([A-Ea-e])/i))) cur.kunci = m[1].toLowerCase();
    else if (cur && !Object.keys(cur.opsi).length) cur.pertanyaan += '\n' + l;
  }
  return out;
}

// Cetak potongan HTML (kartu, label, denah, laporan) tanpa elemen aplikasi lain
function printDoc(html) {
  let area = $('#printArea');
  if (!area) { area = document.createElement('div'); area.id = 'printArea'; document.body.append(area); }
  area.innerHTML = html;
  document.body.classList.add('printing');
  const done = () => { document.body.classList.remove('printing'); area.innerHTML = ''; window.removeEventListener('afterprint', done); };
  window.addEventListener('afterprint', done);
  // tunggu gambar (QR/foto) termuat sebelum mencetak
  const imgs = $$('img', area).filter(i => !i.complete);
  Promise.all(imgs.map(i => new Promise(r => { i.onload = i.onerror = r; }))).then(() => setTimeout(() => window.print(), 50));
}
const qrImg = (data, size = 160) => `<img class="qr" src="/api/qr?s=${size * 2}&d=${encodeURIComponent(data)}" width="${size}" height="${size}" alt="QR">`;
const schoolHead = () => { const s = S.settings; return `<div class="print-head">${s.logo_url ? `<img src="${esc(logoSrc(s))}" alt="">` : ''}<div><b style="font-size:16px">${esc(s.nama_sekolah)}</b><br>${esc(s.alamat || '')}${s.telepon ? ' · Telp. ' + esc(s.telepon) : ''}</div></div>`; };

// Pemindai QR kamera (BarcodeDetector bawaan browser; butuh HTTPS di jaringan)
const Scanner = {
  stream: null, timer: null,
  supported: () => 'BarcodeDetector' in window && !!navigator.mediaDevices,
  async start(video, onCode) {
    this.stop();
    if (!window.isSecureContext) throw new Error('Kamera butuh HTTPS. Buka aplikasi lewat alamat https://… (port 3221) yang tampil di layar server.');
    if (!this.supported()) throw new Error('Browser ini belum mendukung pemindai kamera. Gunakan Chrome di Android, atau scanner barcode USB.');
    const det = new BarcodeDetector({ formats: ['qr_code', 'code_128', 'ean_13', 'code_39'] });
    if (this.stream) stopStream(this.stream);
    this.stream = await openCamera(video, { facingMode: 'environment' });
    if (!this.stream) return; // sudah pindah halaman
    let last = '', lastAt = 0;
    const run = this.run = Symbol('scan'); // putaran lama berhenti bila start() dipanggil lagi
    const loop = async () => {
      if (this.run !== run) return;
      if (!document.body.contains(video)) return this.stop();
      try {
        const codes = await det.detect(video);
        if (codes[0]) { const v = codes[0].rawValue; if (v !== last || Date.now() - lastAt > 3000) { last = v; lastAt = Date.now(); onCode(v); } }
      } catch { /* abaikan frame gagal */ }
      if (this.run === run) this.timer = setTimeout(loop, 300);
    };
    this.timer = setTimeout(loop, 300);
  },
  stop() { this.run = null; clearTimeout(this.timer); this.timer = null; if (this.stream) this.stream.getTracks().forEach(t => t.stop()); this.stream = null; },
};
function beep(ok = true) {
  try {
    const ac = new (window.AudioContext || window.webkitAudioContext)(); const o = ac.createOscillator(); const g = ac.createGain();
    o.frequency.value = ok ? 880 : 220; o.connect(g); g.connect(ac.destination); g.gain.value = .08; o.start(); o.stop(ac.currentTime + (ok ? .12 : .35));
  } catch { /* tanpa suara */ }
}
const monthNow = () => today().slice(0, 7);
const MONTHS = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'];
const monthLabel = ym => { const [y, m] = ym.split('-'); return MONTHS[Number(m) - 1] + ' ' + y; };
const tabs = (items, cur, act = 'tab') => `<div class="tabs no-print" role="tablist">${items.map(([k, l]) => `<button type="button" role="tab" class="${k === cur ? 'on' : ''}" aria-selected="${k === cur}" data-act="${act}" data-tab="${k}">${esc(l)}</button>`).join('')}</div>`;
ACT.tab = el => go(S.page, { ...S.params, tab: el.dataset.tab });

// =====================================================================
// Bagian bersama (dipakai beberapa modul; dulu tersebar di file halaman)
// =====================================================================
function downloadText(name, text, type = 'text/csv') {
  const a = document.createElement('a');
  a.href = URL.createObjectURL(new Blob(['﻿' + text], { type: type + ';charset=utf-8' }));
  a.download = name;
  document.body.append(a); a.click(); a.remove();
  setTimeout(() => URL.revokeObjectURL(a.href), 1000);
}

// ---- Tahun ajaran: pilihan tahun untuk membuka arsip
function yearPick(ys, want) {
  return ys.find(y => String(y.id) === String(want)) || ys.find(y => y.aktif) || ys[0] || { id: '', nama: '–', aktif: 1 };
}
function yearBar(ys, ty, change, note = '') {
  if (ys.length < 2 && ty.aktif) return '';
  return `<div class="bar no-print"><label class="sel">Tahun ajaran <select data-change="${change}" aria-label="Tahun ajaran">${opts(ys.map(y => [y.id, y.nama + (y.aktif ? ' (aktif)' : '')]), ty.id)}</select></label>
    ${ty.aktif ? '' : tag('Arsip', 'warn')}${note ? `<small class="muted">${esc(note)}</small>` : ''}</div>`;
}

// GET dengan informasi jumlah total (header X-Total-Count) untuk daftar berhalaman
async function GETpage(path) {
  const res = await apiFetch('GET', path);
  const data = await apiJSON(res, path);
  return { rows: data || [], total: Number(res.headers.get('X-Total-Count') || (data || []).length), headers: res.headers };
}
// Bilah cari (server) + info jumlah + tombol muat lebih banyak
function pagerBar(page, total, shown, ph) {
  return `<div class="bar"><input class="search" type="search" placeholder="${esc(ph)}" value="${esc(S.params.q || '')}" data-srv="${page}" aria-label="Cari">
    <small class="muted">${shown < total ? `Menampilkan ${shown} dari ${total}` : `${total} data`}</small></div>`;
}
const moreBtn = (page, total, shown) => shown < total ? `<div style="text-align:center;margin-top:12px">${btn('Muat lebih banyak', 'pageMore', { page }, 'ghost')}</div>` : '';
ACT.pageMore = el => go(el.dataset.page, { ...S.params, lim: Number(S.params.lim || 300) + 300 });
document.addEventListener('keydown', e => {
  const el = e.target.closest && e.target.closest('[data-srv]');
  if (el && e.key === 'Enter') { e.preventDefault(); go(el.dataset.srv, { ...S.params, q: el.value.trim(), lim: '' }); }
});
document.addEventListener('search', e => { // tombol x pada input pencarian
  const el = e.target.closest && e.target.closest('[data-srv]');
  if (el && !el.value) go(el.dataset.srv, { ...S.params, q: '', lim: '' });
}, true);
ACT.camStop = () => { Scanner.stop(); const v = $('#scanVideo'); if (v) v.hidden = true; };

// Layar wajib ganti password (akun hasil impor atau direset admin)
// Formulir ganti password (layar wajib ganti & halaman Akun Saya memakai yang sama)
function pwForm(after) {
  const fields = [{ n: 'old', l: 'Password saat ini', t: 'password', req: true, ac: 'current-password' },
    { n: 'new', l: 'Password baru', t: 'password', req: true, minlength: 6, ac: 'new-password' },
    { n: 'new2', l: 'Ulangi password baru', t: 'password', req: true, minlength: 6, ac: 'new-password' }];
  const id = registerForm(fields, async (v, form) => {
    if (v.new !== v.new2) throw new Error('Ulangan password baru tidak sama.');
    await POST('/password', { old: v.old, new: v.new });
    toast('Password berhasil diganti.');
    after(form);
  });
  return { id, fields };
}
function renderGantiPW() {
  const { id, fields } = pwForm(() => { S.user.ganti_pw = false; route(); });
  $('#app').innerHTML = `<div class="login-wrap"><div style="width:min(440px,100%)">${brandHtml('Keamanan akun')}
    <div class="card" style="padding:26px"><h1 style="font-size:24px;margin-bottom:6px">Ganti password</h1>
    <p class="muted" style="margin-bottom:16px">Halo ${esc(S.user.nama_lengkap)}, demi keamanan silakan buat password baru sebelum melanjutkan.</p>
    ${formHtml(id, fields, {}, { submit: 'Simpan & lanjutkan', cancel: false })}
    <p class="login-note"><a href="#" data-act="logout">Keluar</a></p></div></div></div>`;
}

// =====================================================================
// HALAMAN AKUN (semua peran)
// =====================================================================
async function pageAkun() {
  const u = S.user;
  const { id, fields } = pwForm(form => form.reset());
  return head('Akun Saya') + `<div class="cols">
    <section class="card"><div class="card-head"><h3>Data akun</h3></div>
      <div class="li"><small>Nama</small><b>${esc(u.nama_lengkap)}</b></div>
      <div class="li"><small>Username</small><b>${esc(u.username)}</b></div>
      <div class="li"><small>Peran</small><b>${esc(ROLE_LABEL[u.role])}</b></div>
      ${u.nip ? `<div class="li"><small>NIP</small><b>${esc(u.nip)}</b></div>` : ''}
      <p class="muted" style="margin-top:12px;font-size:13px">Perubahan nama atau username dilakukan oleh admin sekolah.</p></section>
    <section class="card"><div class="card-head"><h3>Ganti password</h3></div>${formHtml(id, fields, {}, { submit: 'Ganti password', cancel: false })}</section></div>`;
}

// ---- Helper bersama modul layanan (PPDB, buku tamu, dll.)
// Perkecil foto sebelum dikirim (hemat kuota & batas 2 MB). PDF dikirim apa adanya.
function fileToDataURL(file, maxW = 1600) {
  return new Promise((resolve, reject) => {
    if (file.size > 8 << 20) return reject(new Error('File terlalu besar (maks 8 MB sebelum diperkecil).'));
    const rd = new FileReader();
    rd.onerror = () => reject(new Error('File tidak terbaca.'));
    rd.onload = () => {
      if (!file.type.startsWith('image/')) { if (file.size > 2 << 20) return reject(new Error('PDF maksimal 2 MB.')); return resolve(rd.result); }
      const img = new Image();
      img.onload = () => {
        const s = Math.min(1, maxW / Math.max(img.width, img.height));
        const c = document.createElement('canvas'); c.width = Math.round(img.width * s); c.height = Math.round(img.height * s);
        c.getContext('2d').drawImage(img, 0, 0, c.width, c.height);
        resolve(c.toDataURL('image/jpeg', .82));
      };
      img.onerror = () => reject(new Error('Gambar tidak terbaca.'));
      img.src = rd.result;
    };
    rd.readAsDataURL(file);
  });
}
// tautan WhatsApp untuk nomor Indonesia (08… → 628…); pesan opsional
const waLink = (hp, msg = '') => { let n = String(hp || '').replace(/\D/g, ''); if (n.startsWith('0')) n = '62' + n.slice(1); return n ? `https://wa.me/${n}${msg ? '?text=' + encodeURIComponent(msg) : ''}` : ''; };
const publicShell = (title, inner) => `<div class="login-wrap"><div style="width:min(620px,100%)">${brandHtml(title)}<div class="card" style="padding:26px">${inner}</div>
  <p class="login-note"><a href="#/">Masuk ke aplikasi</a></p></div></div>`;
const linkCard = (url, label) => `<section class="card"><div class="card-head"><h3>${esc(label)}</h3></div><p><code style="word-break:break-all">${esc(url)}</code></p>
  <div style="text-align:center;margin-top:10px">${qrImg(url, 150)}</div><div class="acts" style="justify-content:center;margin-top:10px">${btn('Buka', 'openUrl', { url }, 'ghost sm', 'eye')}</div></section>`;
ACT.openUrl = el => window.open(el.dataset.url, '_blank', 'noopener');

// ---- Tabel responsif: tiap sel diberi label dari judul kolomnya (dipakai CSS .tw.stack di HP: baris → kartu).
// Berlaku otomatis untuk semua tabel di dalam .tw (termasuk yang digambar ulang), kecuali tabel lebar (> 7 kolom).
function stackTables(root = document) {
  for (const t of root.querySelectorAll('.tw > table')) {
    const head = t.tHead && t.tHead.rows[0];
    const wrap = t.parentElement;
    if (!head || wrap.classList.contains('no-stack')) continue;
    const labels = [...head.cells].map(c => c.textContent.trim());
    if (labels.length > 9) continue; // tabel matriks lebar tetap tabel (geser di dalam kotaknya)
    wrap.classList.add('stack');
    for (const tr of t.tBodies[0] ? t.tBodies[0].rows : []) {
      [...tr.cells].forEach((td, i) => {
        if (td.colSpan === 1 && !td.hasAttribute('data-label')) td.setAttribute('data-label', labels[i] || '');
        // sel tanpa isi (bukan tombol/gambar) disembunyikan di tampilan kartu
        const kosong = !td.textContent.replace(/[–-]/g, '').trim() && !td.querySelector('img,button,a,input,select,svg');
        td.toggleAttribute('data-kosong', kosong && i > 0);
      });
    }
  }
}
let stackPending = false;
// hanya bereaksi bila ada TABEL baru (bukan setiap detik timer ujian, toast, atau ketikan di dropdown)
const addsTable = recs => recs.some(r => [...r.addedNodes].some(n => n.nodeType === 1 && (n.tagName === 'TABLE' || n.tagName === 'TR' || n.querySelector?.('table, tr'))));
new MutationObserver(recs => {
  if (stackPending || !addsTable(recs)) return;
  stackPending = true;
  requestAnimationFrame(() => { stackPending = false; stackTables(); });
}).observe(document.documentElement, { childList: true, subtree: true });

// Halaman turunan → menu yang disorot
const PAGE_PARENT = { course: 'mapel', ujian: 'kuis' };
// Registri halaman: diisi oleh pages-*.js dan mod-*.js lewat Object.assign(PAGES.<peran>, {...})
const PAGES = { admin: { akun: pageAkun }, guru: { akun: pageAkun }, siswa: { akun: pageAkun } };

// =====================================================================
// Mulai
// =====================================================================
// =====================================================================
// Pemuatan modul sesuai kebutuhan: halaman login hanya memuat inti; setelah masuk, hanya file untuk peran
// pengguna dan modul yang aktif yang diunduh (file berversi, tersimpan di cache browser). Modul khusus pengelola
// (PPDB, surat, kasir, dll.) hanya dimuat untuk admin atau guru yang diberi tugas tambahan modul itu (isAdm).
// =====================================================================
const BOOT = (() => { try { return JSON.parse(document.getElementById('boot')?.textContent || 'null'); } catch { return null; } })();
const MOD_FILES = [
  ['pages-admin.js', u => u.role === 'admin'],
  ['pages-guru.js', u => u.role === 'guru'],
  ['pages-siswa.js', u => u.role === 'siswa'],
  ['mod-kesiswaan.js', () => ['bukuinduk', 'absensi', 'kalender'].some(modOn)],
  ['mod-cbt.js', () => modOn('ujian')],
  ['mod-manujian.js', () => isAdm('ujian') && modOn('ujian')],
  ['mod-kelulusan.js', () => isAdm('kelulusan') && modOn('kelulusan')],
  ['mod-sarpras.js', () => ['sarpras', 'perpus'].some(modOn)],
  ['mod-ai.js', u => u.role !== 'siswa' && modOn('ai')],
  ['mod-ppdb.js', () => isAdm('ppdb') && modOn('ppdb')],
  ['mod-surat.js', () => isAdm('surat') && modOn('surat')],
  ['mod-bukutamu.js', () => isAdm('bukutamu') && modOn('bukutamu')],
  ['mod-evoting.js', () => modOn('evoting')],
  ['mod-kasir.js', () => isAdm('kasir') && modOn('kasir')],
  ['mod-admguru.js', () => modOn('admguru')],
  ['mod-ljk.js', u => u.role !== 'siswa' && modOn('ujian')],
  ['mod-tahun.js', u => u.role === 'admin'],
  ['mod-audit.js', u => u.role === 'admin'],
];
// halaman publik (tanpa login) dan file yang menyediakannya
const PUBLIC_FILES = { kelulusan: 'mod-kelulusan.js', ppdb: 'mod-ppdb.js', bukutamu: 'mod-bukutamu.js' };
const LOADED = new Map();
function loadScript(name) {
  if (!LOADED.has(name)) {
    const src = (BOOT && BOOT.files && BOOT.files[name]) || name;
    LOADED.set(name, new Promise((ok, fail) => {
      const s = document.createElement('script');
      s.src = src; s.async = false; // urutan eksekusi tetap sesuai urutan daftar
      s.onload = ok; s.onerror = () => { LOADED.delete(name); fail(new Error('Gagal memuat ' + name + '. Periksa koneksi lalu muat ulang halaman.')); };
      document.head.append(s);
    }));
  }
  return LOADED.get(name);
}
async function loadModules() {
  if (!S.user) return;
  await Promise.all(MOD_FILES.filter(([, need]) => need(S.user)).map(([f]) => loadScript(f)));
}

// Halaman publik (tanpa login), diisi modul. Contoh: #/kelulusan
const PUBLIC_PAGES = {};
async function renderPublicOrLogin() {
  document.body.classList.remove('has-shell');
  stopAllMedia();
  for (const id of Object.keys(FORMS)) if (!document.querySelector(`[data-form="${id}"]`)) delete FORMS[id];
  const { page } = parseHash();
  if (PUBLIC_FILES[page]) { try { await loadScript(PUBLIC_FILES[page]); } catch (e) { toast(e.message, 'bad'); } }
  if (S.user) return; // sudah masuk selagi memuat
  if (PUBLIC_PAGES[page] && parseHash().page === page) PUBLIC_PAGES[page](); else renderLogin();
}
window.addEventListener('hashchange', () => { if (!S.user) renderPublicOrLogin(); });

// Dijalankan setelah semua file modul dimuat
window.addEventListener('DOMContentLoaded', async function boot() {
  try {
    // data awal biasanya sudah ikut di halaman (tanpa perjalanan tambahan ke server)
    const [settings, me] = BOOT && BOOT.settings ? [BOOT.settings, BOOT.me || null] : await Promise.all([GET('/settings'), GET('/me')]);
    S.settings = settings || {};
    S.user = me;
    applyWarna(S.settings.warna);
    await loadModules();
  } catch (e) {
    $('#app').innerHTML = `<div class="login-wrap"><div class="card login">${errBox(e)}</div></div>`;
    return;
  }
  ensureThemeFloat();
  if (S.user) route(); else renderPublicOrLogin();
});
window.SCH_READY = 1;
