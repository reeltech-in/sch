'use strict';
/* LJK kertas: cetak lembar jawaban & baca hasil foto/kamera di browser (tanpa pustaka luar, bisa offline).
   Cara membaca: cari 4 penanda hitam di sudut → hitung homografi (meluruskan foto miring/perspektif) →
   tentukan arah lembar dari garis hitam di atas → ukur kegelapan tiap bulatan → pilih jawaban per nomor. */

PAGE_PARENT.ljk = 'cbt';

// ---- Tata letak lembar (milimeter, A4 tegak 210 × 297) ----
const LJK = {
  W: 210, H: 297, MARK: 8,
  corners: [[14, 14], [196, 14], [196, 283], [14, 283]], // pusat penanda: kiri-atas, kanan-atas, kanan-bawah, kiri-bawah
  bar: [60, 12, 150, 16],                                   // garis penunjuk arah (hanya di atas)
  cols: [16, 78, 140], rows: 20, y0: 84, dy: 9.4, r: 3, dx: 9.5, firstX: 13,
  max: 60,
};
const ljkPos = (q, o) => {
  const col = Math.floor(q / LJK.rows), row = q % LJK.rows;
  return [LJK.cols[col] + LJK.firstX + o * LJK.dx, LJK.y0 + row * LJK.dy];
};
const LETTERS = 'ABCDE';

// ---------------------------------------------------------------------
// Cetak
// ---------------------------------------------------------------------
function ljkSheet(u, n, nOpsi, x) {
  const s = S.settings;
  const mm = v => v + 'mm';
  const abs = (l, t, w, h, extra = '') => `position:absolute;left:${mm(l)};top:${mm(t)};width:${mm(w)};height:${mm(h)};${extra}`;
  let h = LJK.corners.map(([cx, cy]) => `<div style="${abs(cx - LJK.MARK / 2, cy - LJK.MARK / 2, LJK.MARK, LJK.MARK, 'background:#000')}"></div>`).join('');
  const [bx1, by1, bx2, by2] = LJK.bar;
  h += `<div style="${abs(bx1, by1, bx2 - bx1, by2 - by1, 'background:#000')}"></div>`;
  h += `<div style="${abs(22, 21, 140, 12, 'font:700 13pt sans-serif;line-height:1.25')}">LEMBAR JAWABAN · ${esc(s.nama_sekolah || '')}<br><span style="font-weight:400;font-size:10pt">${esc(x.judul)}${x.mapel ? ' · ' + esc(x.mapel) : ''}</span></div>`;
  const line = (t, lbl, val) => `<div style="${abs(22, t, 140, 7, 'font:10pt sans-serif;border-bottom:0.3mm solid #000;display:flex;align-items:flex-end;padding-bottom:0.8mm')}"><span style="width:32mm">${lbl}</span><b>${esc(val || '')}</b></div>`;
  h += line(38, 'Nama', u?.nama_lengkap) + line(47, 'No. induk / NIS', u?.username) + line(56, 'Kelas', u?.kelas_nama);
  if (u) h += `<img src="/api/qr?s=300&d=${encodeURIComponent(`LJK:${x.id}:${u.id}`)}" style="${abs(166, 22, 30, 30)}" alt="">`;
  h += `<div style="${abs(16, 68, 180, 8, 'font:8.5pt sans-serif;color:#333')}">Hitamkan penuh satu bulatan per nomor dengan pensil 2B/spidol hitam. Jangan melipat atau mencoret penanda hitam di sudut.</div>`;
  for (let q = 0; q < n; q++) {
    const [, y] = ljkPos(q, 0), col = Math.floor(q / LJK.rows);
    h += `<div style="${abs(LJK.cols[col], y - 2.5, 8, 5, 'font:700 9pt sans-serif;text-align:right')}">${q + 1}.</div>`;
    for (let o = 0; o < nOpsi; o++) {
      const [bx] = ljkPos(q, o);
      h += `<div style="${abs(bx - LJK.r, y - LJK.r, LJK.r * 2, LJK.r * 2, 'border:0.3mm solid #222;border-radius:50%;box-sizing:border-box;font:6.5pt sans-serif;color:#999;display:flex;align-items:center;justify-content:center')}">${LETTERS[o]}</div>`;
    }
  }
  return `<div class="ljk-sheet" style="position:relative;width:210mm;height:297mm;overflow:hidden;page-break-after:always;break-after:page;background:#fff">${h}</div>`;
}
function ljkPrint(peserta, x, kosong = 0) {
  const n = Math.min(x.jumlah_soal, LJK.max), nOpsi = x.jumlah_opsi;
  const sheets = kosong ? Array.from({ length: kosong }, () => ljkSheet(null, n, nOpsi, x)) : peserta.map(u => ljkSheet(u, n, nOpsi, x));
  printDoc(`<style>@page{size:A4 portrait;margin:0}#printArea{padding:0!important;margin:0!important}</style>${sheets.join('')}`);
}

// ---------------------------------------------------------------------
// Pembaca gambar
// ---------------------------------------------------------------------
// Homografi 3×3 dari 4 pasangan titik (src → dst), diselesaikan dengan eliminasi Gauss.
function homography(src, dst) {
  const A = [], b = [];
  for (let i = 0; i < 4; i++) {
    const [x, y] = src[i], [u, v] = dst[i];
    A.push([x, y, 1, 0, 0, 0, -u * x, -u * y]); b.push(u);
    A.push([0, 0, 0, x, y, 1, -v * x, -v * y]); b.push(v);
  }
  for (let c = 0; c < 8; c++) {
    let p = c;
    for (let r = c + 1; r < 8; r++) if (Math.abs(A[r][c]) > Math.abs(A[p][c])) p = r;
    [A[c], A[p]] = [A[p], A[c]]; [b[c], b[p]] = [b[p], b[c]];
    if (Math.abs(A[c][c]) < 1e-12) return null;
    for (let r = 0; r < 8; r++) {
      if (r === c) continue;
      const f = A[r][c] / A[c][c];
      for (let k = c; k < 8; k++) A[r][k] -= f * A[c][k];
      b[r] -= f * b[c];
    }
  }
  const h = b.map((v, i) => v / A[i][i]);
  return (x, y) => { const w = h[6] * x + h[7] * y + 1; return [(h[0] * x + h[1] * y + h[2]) / w, (h[3] * x + h[4] * y + h[5]) / w]; };
}

function otsu(gray) {
  const hist = new Float64Array(256);
  for (let i = 0; i < gray.length; i++) hist[gray[i]]++;
  let sum = 0; for (let i = 0; i < 256; i++) sum += i * hist[i];
  let sumB = 0, wB = 0, best = 0, t = 128;
  for (let i = 0; i < 256; i++) {
    wB += hist[i]; if (!wB) continue;
    const wF = gray.length - wB; if (!wF) break;
    sumB += i * hist[i];
    const mB = sumB / wB, mF = (sum - sumB) / wF, v = wB * wF * (mB - mF) ** 2;
    if (v > best) { best = v; t = i; }
  }
  return t;
}

// Cari kandidat penanda: komponen gelap berbentuk persegi padat.
function findMarkers(gray, W, H, t) {
  const lab = new Int32Array(W * H), stack = new Int32Array(W * H);
  const minSide = Math.max(W, H) * 0.012, maxSide = Math.max(W, H) * 0.09;
  const cands = [];
  let id = 0;
  for (let p0 = 0; p0 < W * H; p0++) {
    if (lab[p0] || gray[p0] > t) continue;
    id++;
    let sp = 0, area = 0, minx = W, maxx = 0, miny = H, maxy = 0, sx = 0, sy = 0, sxx = 0, syy = 0, sxy = 0;
    stack[sp++] = p0; lab[p0] = id;
    while (sp) {
      const p = stack[--sp], x = p % W, y = (p - x) / W;
      area++; sx += x; sy += y; sxx += x * x; syy += y * y; sxy += x * y;
      if (x < minx) minx = x; if (x > maxx) maxx = x; if (y < miny) miny = y; if (y > maxy) maxy = y;
      if (x > 0 && !lab[p - 1] && gray[p - 1] <= t) { lab[p - 1] = id; stack[sp++] = p - 1; }
      if (x < W - 1 && !lab[p + 1] && gray[p + 1] <= t) { lab[p + 1] = id; stack[sp++] = p + 1; }
      if (y > 0 && !lab[p - W] && gray[p - W] <= t) { lab[p - W] = id; stack[sp++] = p - W; }
      if (y < H - 1 && !lab[p + W] && gray[p + W] <= t) { lab[p + W] = id; stack[sp++] = p + W; }
    }
    const bw = maxx - minx + 1, bh = maxy - miny + 1;
    if (bw < minSide || bh < minSide || bw > maxSide || bh > maxSide) continue;
    // bentuk padat & tidak lonjong, dihitung dari momen sehingga tidak terpengaruh kemiringan:
    // persegi/bulatan penuh → kompak ≈ 1; huruf, cincin, atau gumpalan tak beraturan → jauh dari 1
    const mx = sx / area, my = sy / area;
    const vx = sxx / area - mx * mx, vy = syy / area - my * my, cxy = sxy / area - mx * my;
    const tr = vx + vy, disc = Math.sqrt(Math.max(0, tr * tr / 4 - (vx * vy - cxy * cxy)));
    const compact = area / (6 * tr), elong = (tr / 2 + disc) / Math.max(1e-9, tr / 2 - disc);
    if (compact < 0.8 || compact > 1.25 || elong > 1.8) continue;
    cands.push({ x: mx, y: my, area });
  }
  return cands;
}

// Pilih 4 penanda: kandidat paling ujung di tiap arah diagonal (kiri-atas, kanan-atas, kanan-bawah, kiri-bawah).
function pickCorners(c) {
  if (c.length < 4) return null;
  const dirs = [[-1, -1], [1, -1], [1, 1], [-1, 1]];
  const out = dirs.map(([a, b]) => c.reduce((best, k) => (a * k.x + b * k.y > a * best.x + b * best.y ? k : best)));
  if (new Set(out).size < 4) return null;
  // ukuran keempat penanda harus mirip
  const areas = out.map(k => k.area);
  if (Math.max(...areas) > 3.5 * Math.min(...areas)) return null;
  return out;
}

// Baca satu gambar LJK. Hasil: { jawaban: ['a'|''|'?', …], skor kegelapan, qr, kanvas pratinjau }
async function ljkRead(src, nSoal, nOpsi) {
  const sw = src.videoWidth || src.naturalWidth || src.width, sh = src.videoHeight || src.naturalHeight || src.height;
  const sc = Math.min(1, 1600 / Math.max(sw, sh));
  const W = Math.round(sw * sc), H = Math.round(sh * sc);
  const cv = document.createElement('canvas'); cv.width = W; cv.height = H;
  const ctx = cv.getContext('2d', { willReadFrequently: true });
  ctx.drawImage(src, 0, 0, W, H);
  const px = ctx.getImageData(0, 0, W, H).data, gray = new Uint8Array(W * H);
  for (let i = 0, j = 0; i < gray.length; i++, j += 4) gray[i] = (px[j] * 77 + px[j + 1] * 150 + px[j + 2] * 29) >> 8;
  const t = Math.min(otsu(gray), 150);
  const corners = pickCorners(findMarkers(gray, W, H, t));
  if (!corners) throw new Error('Penanda sudut tidak ditemukan. Foto seluruh lembar (keempat kotak hitam terlihat), tegak lurus, dengan cahaya cukup.');
  const at = (x, y) => { const xi = Math.round(x), yi = Math.round(y); return xi < 0 || yi < 0 || xi >= W || yi >= H ? 255 : gray[yi * W + xi]; };
  // kegelapan rata-rata di dalam lingkaran (mm) pada lembar
  const dark = (map, cx, cy, r) => {
    let s = 0, n = 0;
    for (let dy = -r; dy <= r + 1e-9; dy += r / 3) for (let dx = -r; dx <= r + 1e-9; dx += r / 3) {
      if (dx * dx + dy * dy > r * r) continue;
      const [x, y] = map(cx + dx, cy + dy); s += 255 - at(x, y); n++;
    }
    return s / n / 255;
  };
  // coba 4 kemungkinan arah lembar; yang benar memiliki garis hitam di atas
  const img = corners.map(k => [k.x, k.y]);
  let best = null;
  for (let r = 0; r < 4; r++) {
    const map = homography(LJK.corners, LJK.corners.map((_, i) => img[(i + r) % 4]));
    if (!map) continue;
    const [x1, , x2] = LJK.bar;
    let bar = 0, blank = 0;
    for (let x = x1 + 3; x <= x2 - 3; x += 6) { bar += dark(map, x, 14, 0.8); blank += dark(map, x, 19.5, 0.8); }
    const score = bar - blank;
    if (!best || score > best.score) best = { map, score };
  }
  if (!best || best.score < 0.25 * 15) throw new Error('Arah lembar tidak terbaca. Pastikan garis hitam di bagian atas lembar terlihat jelas.');
  const map = best.map;
  const jawaban = [], nilai = [];
  for (let q = 0; q < nSoal; q++) {
    const d = [];
    for (let o = 0; o < nOpsi; o++) { const [x, y] = ljkPos(q, o); d.push(dark(map, x, y, LJK.r * 0.62)); }
    const mn = Math.min(...d), ord = d.map((v, i) => [v - mn, i]).sort((a, b) => b[0] - a[0]);
    const [d1, i1] = ord[0], d2 = ord[1] ? ord[1][0] : 0;
    let ans = '';
    if (d1 >= 0.18) ans = d2 >= 0.18 && d2 >= 0.55 * d1 ? '?' : 'abcde'[i1];
    jawaban.push(ans); nilai.push(d);
  }
  // pratinjau: tandai penanda & bacaan
  ctx.lineWidth = Math.max(2, W / 500);
  corners.forEach(k => { ctx.strokeStyle = '#e11d48'; ctx.strokeRect(k.x - 8, k.y - 8, 16, 16); });
  for (let q = 0; q < nSoal; q++) for (let o = 0; o < nOpsi; o++) {
    const a = jawaban[q], hit = a === 'abcde'[o] || (a === '?' && nilai[q][o] - Math.min(...nilai[q]) >= 0.18);
    if (!hit) continue;
    const [x, y] = map(...ljkPos(q, o)), [x2] = map(ljkPos(q, o)[0] + LJK.r, ljkPos(q, o)[1]);
    ctx.strokeStyle = a === '?' ? '#f59e0b' : '#059669'; ctx.beginPath(); ctx.arc(x, y, Math.abs(x2 - x) * 1.3, 0, 7); ctx.stroke();
  }
  // QR identitas (bila browser mendukung pembaca bawaan)
  let qr = null;
  if ('BarcodeDetector' in window) {
    try { const codes = await new BarcodeDetector({ formats: ['qr_code'] }).detect(cv); qr = (codes.find(c => /^LJK:\d+:\d+$/.test(c.rawValue)) || {}).rawValue || null; } catch { /* abaikan */ }
  }
  return { jawaban, nilai, qr, canvas: cv };
}

// ---------------------------------------------------------------------
// Halaman LJK
// ---------------------------------------------------------------------
async function pageLJK() {
  const id = S.params.id;
  const [info, soal] = await Promise.all([GET(`/cbt/${id}/ljk`), GET(`/cbt/${id}/soal`)]);
  const x = info.ujian;
  S.ljk = { x, peserta: info.peserta, kunci: soal.map(q => q.jawaban_benar), hasil: S.ljk && S.ljk.x.id === x.id ? S.ljk.hasil : [] };
  const tooMany = x.jumlah_soal > x.maks_soal;
  const sudah = info.peserta.filter(p => p.selesai).length;
  return head('LJK Kertas', `${esc(x.judul)}${x.mapel ? ' · ' + esc(x.mapel) : ''} · ${x.jumlah_soal} soal, ${x.jumlah_opsi} pilihan`,
      btn('Kembali ke CBT', 'go', { page: 'cbt', params: '{}' }, 'ghost', 'back')) +
    (tooMany ? `<div class="hint" style="margin-bottom:14px">LJK kertas mendukung maksimal ${x.maks_soal} soal; ujian ini punya ${x.jumlah_soal} soal.</div>` : '') +
    (!x.jumlah_soal ? `<div class="card">${empty('Tambahkan soal & kunci jawaban dulu sebelum memakai LJK.', 'quiz')}</div>` : tooMany ? '' : `
    <div class="cols">
      <section class="card"><div class="card-head"><h3>1. Cetak LJK</h3></div>
        <p class="muted" style="margin-bottom:12px">Lembar per peserta memuat nama & QR, sehingga siswa dikenali otomatis saat dipindai. Cetak A4 tanpa skala (100%).</p>
        <div class="acts" style="justify-content:flex-start">${btn(`Cetak untuk ${info.peserta.length} peserta`, 'ljkCetak', {}, '', 'print')}${btn('Cetak lembar kosong', 'ljkCetakKosong', {}, 'ghost', 'print')}</div></section>
      <section class="card"><div class="card-head"><h3>2. Pindai jawaban</h3></div>
        <p class="muted" style="margin-bottom:12px">Foto seluruh lembar dari atas (keempat kotak hitam terlihat). Bisa pilih banyak foto sekaligus.</p>
        <div class="acts" style="justify-content:flex-start">
          <label class="btn">${ico('image')} Foto / pilih gambar<input type="file" accept="image/*" capture="environment" multiple data-change="ljkFiles" hidden></label>
          ${window.isSecureContext && navigator.mediaDevices ? btn('Kamera langsung', 'ljkCam', {}, 'ghost', 'eye') : ''}</div>
        <div id="ljkCamBox" hidden style="margin-top:12px"><div class="ljk-cam"><video id="ljkVideo" playsinline muted></video><div class="ljk-guide"></div></div>
          <div class="acts" style="justify-content:flex-start;margin-top:8px">${btn('Ambil & baca', 'ljkSnap', {}, '', 'check')}${btn('Tutup kamera', 'ljkCamStop', {}, 'ghost')}</div></div></section>
    </div>
    <div id="ljkHasil">${S.ljk.hasil.map(ljkCard).join('')}</div>
    <section class="card" style="margin-top:14px"><div class="card-head"><h3>Peserta</h3><small class="muted">${sudah} dari ${info.peserta.length} sudah bernilai</small></div>
      <div class="tw"><table><thead><tr><th>Nama</th><th>Kelas</th><th>Status</th><th class="n">Nilai</th></tr></thead><tbody>
      ${info.peserta.map(p => `<tr><td><b>${esc(p.nama_lengkap)}</b><small class="muted">${esc(p.username)}</small></td><td>${esc(p.kelas_nama)}</td>
        <td>${p.selesai ? tag(p.metode === 'ljk' ? 'LJK' : 'CBT daring', p.metode === 'ljk' ? 'brand' : 'ok') : '<small class="muted">Belum</small>'}</td><td class="n">${p.selesai ? fmtN(p.skor) : '–'}</td></tr>`).join('')}
      </tbody></table></div></section>`);
}

// kartu hasil satu lembar: pratinjau, pilihan siswa, jawaban yang bisa dikoreksi
function ljkCard(h, i) {
  const { x, peserta, kunci } = S.ljk;
  const benar = h.jawaban.filter((a, q) => a && a === kunci[q]).length;
  const kosong = h.jawaban.filter(a => !a).length, ganda = h.jawaban.filter(a => a === '?').length;
  const skor = kunci.length ? Math.round(benar / kunci.length * 10000) / 100 : 0;
  return `<section class="card ljk-card" style="margin-top:14px"><div class="card-head"><h3>Lembar ${i + 1}${h.saved ? ' ' + tag('Tersimpan', 'ok') : ''}</h3>
      <small class="muted">${benar} benar · ${kosong} kosong · ${ganda ? `<b style="color:var(--warn)">${ganda} ganda</b>` : '0 ganda'} · perkiraan nilai <b>${fmtN(skor)}</b> (tanpa bobot)</small></div>
    <div class="ljk-res"><img src="${h.preview}" alt="Pratinjau lembar ${i + 1}" class="ljk-prev">
      <div style="min-width:0">
        <label class="f"><span>Siswa${h.dariQR ? ' <small class="muted">(dari QR)</small>' : ''}</span><select data-change="ljkSiswa" data-i="${i}">${opts(peserta.map(p => [p.id, `${p.nama_lengkap} · ${p.kelas_nama}${p.selesai ? ' (sudah bernilai)' : ''}`]), h.siswa_id || '', '— Pilih siswa —')}</select></label>
        <div class="ljk-grid">${h.jawaban.map((a, q) => `<label class="${a === '?' ? 'ganda' : !a ? 'kosong' : a === kunci[q] ? 'ok' : 'salah'}"><span>${q + 1}</span>
          <select data-change="ljkJawab" data-i="${i}" data-q="${q}" aria-label="Jawaban nomor ${q + 1}">${opts([['', '–'], ['?', '??'], ...'abcde'.slice(0, x.jumlah_opsi).split('').map(l => [l, l.toUpperCase()])], a)}</select></label>`).join('')}</div>
        <p class="muted" style="margin:8px 0">Hijau = benar, merah = salah, kuning = ganda (pilih salah satu atau kosongkan), abu = kosong. Koreksi bila bacaan keliru.</p>
        <div class="acts" style="justify-content:flex-start">${btn(h.saved ? 'Simpan ulang' : 'Simpan nilai', 'ljkSave', { i }, '', 'check')}${btn('Buang', 'ljkBuang', { i }, 'ghost')}</div>
      </div></div></section>`;
}
const ljkRedraw = () => { const el = $('#ljkHasil'); if (el) el.innerHTML = S.ljk.hasil.map(ljkCard).join(''); };

async function ljkProcess(src, label) {
  const { x, peserta } = S.ljk;
  try {
    const r = await ljkRead(src, Math.min(x.jumlah_soal, LJK.max), x.jumlah_opsi);
    const m = r.qr && r.qr.match(/^LJK:(\d+):(\d+)$/);
    let siswa_id = '', dariQR = false;
    if (m && String(m[1]) === String(x.id) && peserta.some(p => String(p.id) === m[2])) { siswa_id = Number(m[2]); dariQR = true; }
    const pv = document.createElement('canvas'), k = 520 / Math.max(r.canvas.width, r.canvas.height);
    pv.width = Math.round(r.canvas.width * k); pv.height = Math.round(r.canvas.height * k);
    pv.getContext('2d').drawImage(r.canvas, 0, 0, pv.width, pv.height);
    S.ljk.hasil.unshift({ jawaban: r.jawaban, siswa_id, dariQR, preview: pv.toDataURL('image/jpeg', 0.75) });
    if (m && String(m[1]) !== String(x.id)) toast(`${label}: QR milik ujian lain. Periksa lembarnya.`, 'bad');
  } catch (e) { toast(`${label}: ${e.message}`, 'bad'); }
  ljkRedraw();
}
CHG.ljkFiles = async el => {
  const files = [...el.files]; el.value = '';
  for (const [n, f] of files.entries()) {
    toast(`Membaca foto ${n + 1} dari ${files.length}…`);
    let src = window.createImageBitmap ? await createImageBitmap(f, { imageOrientation: 'from-image' }).catch(() => createImageBitmap(f).catch(() => null)) : null;
    let url = null;
    if (!src) { // browser lama: muat lewat <img>
      url = URL.createObjectURL(f);
      src = await new Promise(ok => { const im = new Image(); im.onload = () => ok(im); im.onerror = () => ok(null); im.src = url; });
    }
    if (!src) { URL.revokeObjectURL(url); toast(`Foto ${n + 1}: format gambar tidak didukung.`, 'bad'); continue; }
    await ljkProcess(src, `Foto ${n + 1}`);
    if (src.close) src.close();
    if (url) URL.revokeObjectURL(url);
  }
};
ACT.ljkCam = async () => {
  const v = $('#ljkVideo');
  try {
    ACT.ljkCamStop(); // tombol diketuk dua kali: kamera lama dimatikan dulu
    S.ljkStream = await openCamera(v, { facingMode: 'environment', width: { ideal: 1920 }, height: { ideal: 1080 } });
    if (S.ljkStream) $('#ljkCamBox').hidden = false;
  } catch { toast('Kamera tidak bisa dibuka. Izinkan akses kamera atau pakai tombol Foto.', 'bad'); }
};
ACT.ljkCamStop = () => { stopStream(S.ljkStream, $('#ljkVideo')); S.ljkStream = null; const b = $('#ljkCamBox'); if (b) b.hidden = true; };
ACT.ljkSnap = () => ljkProcess($('#ljkVideo'), 'Kamera');
CHG.ljkSiswa = el => { S.ljk.hasil[el.dataset.i].siswa_id = el.value ? Number(el.value) : ''; };
CHG.ljkJawab = el => { // gambar ulang kartu itu saja (bukan semua kartu)
  const i = Number(el.dataset.i);
  S.ljk.hasil[i].jawaban[el.dataset.q] = el.value;
  const card = el.closest('.ljk-card');
  if (card) card.outerHTML = ljkCard(S.ljk.hasil[i], i); else ljkRedraw();
  const sel = $(`[data-change="ljkJawab"][data-i="${i}"][data-q="${el.dataset.q}"]`);
  if (sel) sel.focus();
};
ACT.ljkBuang = el => { S.ljk.hasil.splice(Number(el.dataset.i), 1); ljkRedraw(); };
ACT.ljkSave = async el => {
  const h = S.ljk.hasil[el.dataset.i];
  if (!h.siswa_id) return toast('Pilih siswa pemilik lembar ini dulu.', 'bad');
  const ganda = h.jawaban.map((a, q) => a === '?' ? q + 1 : 0).filter(Boolean);
  if (ganda.length) return toast(`Nomor ${ganda.join(', ')} terbaca ganda. Pilih jawabannya atau kosongkan dulu.`, 'bad');
  const body = { siswa_id: h.siswa_id, jawaban: h.jawaban };
  const p = S.ljk.peserta.find(x => x.id === h.siswa_id);
  if (p && p.selesai && !h.saved && !(await confirmBox(`${esc(p.nama_lengkap)} sudah punya nilai ${fmtN(p.skor)} (${p.metode === 'ljk' ? 'LJK' : 'CBT daring'}). Ganti dengan hasil lembar ini?`, 'Ganti nilai', 'danger'))) return;
  const r = await POST(`/cbt/${S.ljk.x.id}/ljk`, { ...body, timpa: true });
  h.saved = true;
  toast(`${p ? p.nama_lengkap : 'Siswa'}: nilai ${fmtN(r.skor)} (${r.benar}/${r.jumlah_soal} benar) tersimpan.`);
  await refresh(true);
};
ACT.ljkCetak = () => { if (!S.ljk.peserta.length) return toast('Belum ada peserta di kelas ujian ini.', 'bad'); ljkPrint(S.ljk.peserta, S.ljk.x); };
ACT.ljkCetakKosong = () => openForm({
  title: 'Cetak LJK kosong', submit: 'Cetak', success: 'Siap dicetak.',
  intro: '<p class="muted" style="margin-bottom:10px">Lembar tanpa nama & QR; siswa menulis namanya, lalu siswa dipilih manual saat memindai.</p>',
  fields: [{ n: 'jumlah', l: 'Jumlah lembar', t: 'number', min: 1, max: 200, v: 30, req: true }],
  onSubmit: async v => { setTimeout(() => ljkPrint([], S.ljk.x, Math.min(200, v.jumlah || 1)), 300); },
});
ACT.ljkOpen = el => go('ljk', { id: el.dataset.id });

Object.assign(PAGES.admin, { ljk: pageLJK });
Object.assign(PAGES.guru, { ljk: pageLJK });
