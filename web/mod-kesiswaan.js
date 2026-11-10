'use strict';
// pilihan pekerjaan orang tua (mengikuti Dapodik)
const PEKERJAAN = ['Tidak bekerja', 'Nelayan', 'Petani', 'Peternak', 'PNS/TNI/Polri', 'Karyawan swasta', 'Pedagang kecil', 'Pedagang besar',
  'Wiraswasta', 'Wirausaha', 'Buruh', 'Pensiunan', 'Tenaga Kerja Indonesia', 'Sudah meninggal', 'Lainnya'].map(x => [x, x]);
/* Modul Kesiswaan: Buku Induk, Absensi Harian (scan QR / ceklist / selfie), Kartu Siswa, Kalender Akademik */

addNav('admin', ['bukuinduk', 'Buku Induk', 'book', 'bukuinduk', 'Kesiswaan']);
addNav('admin', ['absenharian', 'Absensi Harian', 'cal', 'absensi', 'Kesiswaan']);
addNav('admin', ['kartu', 'Kartu Siswa', 'user', 'absensi', 'Kesiswaan']);
addNav('admin', ['kalender', 'Kalender Akademik', 'cal', 'kalender', 'Kesiswaan']);
addNav('guru', ['absenharian', 'Absensi Harian', 'cal', 'absensi', 'Kesiswaan']);
addNav('guru', ['bukuinduk', 'Buku Induk', 'book', 'bukuinduk', 'Kesiswaan']);
addNav('guru', ['kalender', 'Kalender Akademik', 'cal', 'kalender', 'Kesiswaan']);
addNav('siswa', ['absenku', 'Absensi & Kartu', 'cal', 'absensi', 'Kesiswaan']);
addNav('siswa', ['kalender', 'Kalender Akademik', 'cal', 'kalender', 'Kesiswaan']);

const kelasSelect = (ks, val, change, blank = 'Pilih kelas') =>
  `<select data-change="${change}" aria-label="Kelas">${blank ? `<option value="">${esc(blank)}</option>` : ''}${opts(ks.map(k => [k.id, k.nama]), val)}</select>`;
const firstKelas = (ks, p) => (p && ks.some(k => String(k.id) === String(p))) ? Number(p) : (ks[0] ? ks[0].id : null);

// =====================================================================
// BUKU INDUK
// =====================================================================
const PROFILE_FIELDS = [
  ['nisn', 'NISN'], ['nik', 'NIK'], ['jk', 'Jenis kelamin'], ['agama', 'Agama'], ['tempat_lahir', 'Tempat lahir'], ['tgl_lahir', 'Tanggal lahir'],
  ['alamat', 'Alamat'], ['no_hp', 'No. HP siswa'], ['nama_ayah', 'Nama ayah'], ['pekerjaan_ayah', 'Pekerjaan ayah'], ['nama_ibu', 'Nama ibu'],
  ['pekerjaan_ibu', 'Pekerjaan ibu'], ['nama_wali', 'Nama wali'], ['hp_ortu', 'No. HP orang tua/wali'], ['asal_sekolah', 'Asal sekolah'],
  ['tgl_masuk', 'Tanggal masuk'], ['catatan', 'Catatan'],
];
const fmtProfile = (k, v) => !v ? '–' : (k === 'tgl_lahir' || k === 'tgl_masuk') ? fmtDate(v) : k === 'jk' ? (v === 'L' ? 'Laki-laki' : 'Perempuan') : v;

// Guru tanpa akses kelola modul hanya melihat kelas yang ia ajar (sesuai pembatasan di server).
async function myClasses(mod) {
  if (isAdm(mod)) return GET('/classes');
  const [ks, cs] = await Promise.all([GET('/classes'), GET('/courses')]);
  const mine = new Set(cs.map(c => c.kelas_id));
  return ks.filter(k => mine.has(k.id));
}

async function pageBukuInduk() {
  const kid = S.params.kelas || '';
  const [ks, rows] = await Promise.all([myClasses('bukuinduk'), GET('/bukuinduk' + (kid ? '?kelas_id=' + encodeURIComponent(kid) : ''))]);
  const terisi = rows.filter(r => r.terisi).length;
  return head('Buku Induk Siswa', `${rows.length} siswa · ${terisi} biodata terisi${rows.length - terisi ? ` · <b>${rows.length - terisi} belum</b>` : ''}`) +
    `<div class="bar"><input class="search" type="search" placeholder="Cari nama, NIS, NISN, orang tua…" data-filter="tblBI" aria-label="Cari">${kelasSelect(ks, kid, 'biKelas', 'Semua kelas')}</div>` +
    (rows.length ? `<div class="tw"><table id="tblBI"><thead><tr><th>Siswa</th><th>NIS / NISN</th><th>Kelas</th><th>L/P</th><th>Tempat, tgl lahir</th><th>Orang tua</th><th></th></tr></thead><tbody>
    ${rows.map(r => `<tr><td><div class="cell-user">${avatar(r.foto, r.nama_lengkap)}<div><b>${esc(r.nama_lengkap)}</b>${r.terisi ? '' : tag('Belum diisi', 'warn')}</div></div></td>
      <td>${esc(r.username)}<small class="muted">${esc(r.nisn || '–')}</small></td><td>${esc(r.kelas_nama || '–')}</td><td>${esc(r.jk || '–')}</td>
      <td>${r.tempat_lahir || r.tgl_lahir ? esc(r.tempat_lahir || '') + (r.tgl_lahir ? ', ' + fmtDate(r.tgl_lahir) : '') : '–'}</td>
      <td>${esc(r.nama_ayah || r.nama_ibu || '–')}<small class="muted">${esc(r.hp_ortu || '')}</small></td>
      <td><div class="acts">${iconBtn('eye', 'biView', { id: r.id }, 'Lihat & cetak')}${isAdm('bukuinduk') ? iconBtn('edit', 'biEdit', { id: r.id }, 'Ubah biodata') : ''}</div></td></tr>`).join('')}
    </tbody></table></div>` : `<div class="card">${empty('Belum ada siswa.', 'users')}</div>`);
}
CHG.biKelas = el => go('bukuinduk', { kelas: el.value });

ACT.biEdit = async el => {
  const p = await GET('/bukuinduk/' + el.dataset.id);
  openForm({
    title: 'Biodata: ' + p.nama_lengkap, wide: true, values: p,
    intro: `<p class="muted" style="margin-bottom:12px">${esc(p.username)} · ${esc(p.kelas_nama || 'tanpa kelas')}. NISN dan tanggal lahir dipakai untuk cek pengumuman kelulusan.</p>`,
    fields: [{ n: 'foto', l: 'Foto siswa', t: 'file', accept: 'image/*' },
      [{ n: 'nisn', l: 'NISN' }, { n: 'nik', l: 'NIK' }],
      [{ n: 'jk', l: 'Jenis kelamin', t: 'select', o: [['L', 'Laki-laki'], ['P', 'Perempuan']], blank: '—' }, { n: 'agama', l: 'Agama', t: 'select', o: ['Islam', 'Kristen', 'Katolik', 'Hindu', 'Buddha', 'Konghucu'].map(a => [a, a]), blank: '—' }],
      [{ n: 'tempat_lahir', l: 'Tempat lahir' }, { n: 'tgl_lahir', l: 'Tanggal lahir', t: 'date' }],
      { n: 'alamat', l: 'Alamat', t: 'textarea', rows: 2 },
      [{ n: 'no_hp', l: 'No. HP siswa' }, { n: 'hp_ortu', l: 'No. HP orang tua/wali' }],
      [{ n: 'nama_ayah', l: 'Nama ayah' }, { n: 'pekerjaan_ayah', l: 'Pekerjaan ayah', t: 'select', o: PEKERJAAN, blank: '—' }],
      [{ n: 'nama_ibu', l: 'Nama ibu' }, { n: 'pekerjaan_ibu', l: 'Pekerjaan ibu', t: 'select', o: PEKERJAAN, blank: '—' }],
      [{ n: 'nama_wali', l: 'Nama wali' }, { n: 'asal_sekolah', l: 'Asal sekolah' }],
      [{ n: 'tgl_masuk', l: 'Tanggal masuk', t: 'date' }, { t: 'html', html: '<div></div>' }],
      { n: 'catatan', l: 'Catatan', t: 'textarea', rows: 2 }],
    onSubmit: v => PUT('/bukuinduk/' + p.id, v),
  });
};

function profileHtml(p) {
  return `<div style="display:flex;gap:18px;align-items:flex-start;flex-wrap:wrap">${p.foto ? `<img src="${esc(p.foto)}" alt="" style="width:110px;height:140px;object-fit:cover;border-radius:12px;border:1px solid var(--border)">` : ''}
    <table style="flex:1;min-width:260px"><tbody>
      <tr><th style="width:38%">Nama lengkap</th><td><b>${esc(p.nama_lengkap)}</b></td></tr>
      <tr><th>NIS</th><td>${esc(p.username)}</td></tr><tr><th>Kelas</th><td>${esc(p.kelas_nama || '–')}</td></tr>
      ${PROFILE_FIELDS.map(([k, l]) => `<tr><th>${esc(l)}</th><td>${esc(fmtProfile(k, p[k]))}</td></tr>`).join('')}
    </tbody></table></div>`;
}
ACT.biView = async el => {
  const p = await GET('/bukuinduk/' + el.dataset.id);
  S.biPrint = p;
  openModal({ title: 'Buku Induk: ' + p.nama_lengkap, wide: true, body: `<div class="tw" style="padding:14px">${profileHtml(p)}</div>
    <div class="sheet-foot" style="margin-top:14px">${btn('Cetak', 'biPrint', {}, '', 'print')}</div>` });
};
ACT.biPrint = () => printDoc(schoolHead() + `<h2 style="text-align:center;margin:6px 0 14px">BUKU INDUK SISWA</h2>` + profileHtml(S.biPrint));

// =====================================================================
// ABSENSI HARIAN (admin & guru)
// =====================================================================
S.scanLog = [];
async function pageAbsenHarian() {
  const tab = S.params.tab || 'scan';
  const t = tabs([['scan', 'Scan QR'], ['ceklist', 'Ceklist Kelas'], ['rekap', 'Rekap Bulanan']], tab);
  if (tab === 'scan') return head('Absensi Harian', 'Scan kartu siswa di gerbang atau kelas. Bisa memakai scanner barcode USB atau kamera.') + t + absenScanHtml();
  const ks = await myClasses('absensi');
  if (!ks.length) return head('Absensi Harian') + t + `<div class="card">${empty(isAdm('absensi') ? 'Belum ada kelas.' : 'Anda belum mengajar di kelas mana pun.', 'school')}</div>`;
  const kid = firstKelas(ks, S.params.kelas);
  return head('Absensi Harian', tab === 'rekap' ? 'Rekap kehadiran harian per kelas.' : 'Tandai kehadiran siswa satu kelas sekaligus.') + t +
    (tab === 'ceklist' ? await absenCeklist(ks, kid) : await absenRekap(ks, kid));
}

function absenScanHtml() {
  S.after = () => {
    const inp = $('#scanIn');
    inp?.focus();
    onScan(inp, doScan);
  };
  return `<div class="cols"><section class="card scan-box">
      <label class="f"><span>Mode</span><select id="scanMode"><option value="auto">Otomatis (pertama = masuk, berikutnya = pulang)</option><option value="masuk">Masuk saja</option><option value="pulang">Pulang saja</option></select></label>
      <input class="scan-input" id="scanIn" placeholder="Arahkan scanner ke kartu…" autocomplete="off" aria-label="Isi QR kartu">
      <div class="acts" style="justify-content:flex-start">${btn('Pakai kamera', 'camStart', {}, 'ghost', 'eye')}${btn('Matikan kamera', 'camStop', {}, 'ghost')}</div>
      <video class="scan-video" id="scanVideo" playsinline muted hidden></video>
      <p class="muted" style="font-size:13px">Scanner USB bekerja seperti keyboard: klik kotak di atas, lalu scan. Batas jam masuk diatur di Pengaturan.</p>
    </section>
    <section class="card"><div class="card-head"><h3>Hasil scan</h3><small class="muted" id="scanCount">${S.scanLog.filter(x => !x.error).length} tercatat sesi ini</small></div>
      <div id="scanLast">${S.scanLog[0] ? scanCard(S.scanLog[0]) : empty('Belum ada scan.', 'eye')}</div>
      <div id="scanLog" style="margin-top:12px">${S.scanLog.slice(1, 12).map(scanRow).join('')}</div></section></div>`;
}
const AKSI = { masuk: 'Masuk', pulang: 'Pulang', sudah: 'Sudah tercatat' };
function scanCard(r) {
  if (r.error) return `<div class="scan-result bad">${avatar(null, '!', 64)}<div><b>Gagal</b>${esc(r.error)}</div></div>`;
  const kind = r.aksi === 'sudah' ? 'warn' : r.status === 'terlambat' && r.aksi === 'masuk' ? 'warn' : 'ok';
  return `<div class="scan-result ${kind}">${avatar(r.foto, r.nama, 64)}<div><b>${esc(r.nama)}</b>${esc(r.kelas)} · ${AKSI[r.aksi]} ${r.aksi === 'pulang' ? r.jam_pulang : r.jam_masuk || ''}
    ${r.status === 'terlambat' ? tag('Terlambat', 'warn') : ''}</div></div>`;
}
const scanRow = r => r.error ? `<div class="li"><div><b style="color:var(--bad)">Gagal</b><small class="muted">${esc(r.error)}</small></div><small>${r.at}</small></div>`
  : `<div class="li"><div><b>${esc(r.nama)}</b><small class="muted">${esc(r.kelas)} · ${AKSI[r.aksi]}${r.status === 'terlambat' ? ' · terlambat' : ''}</small></div><small>${r.at}</small></div>`;
async function doScan(token) {
  const mode = $('#scanMode')?.value || 'auto';
  let r;
  try { r = await POST('/absen/scan', { token, mode }); beep(r.aksi !== 'sudah'); }
  catch (e) { r = { error: e.message }; beep(false); }
  r.at = new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  S.scanLog.unshift(r);
  S.scanLog = S.scanLog.slice(0, 200);
  if ($('#scanLast')) { $('#scanLast').innerHTML = scanCard(r); $('#scanLog').innerHTML = S.scanLog.slice(1, 12).map(scanRow).join(''); $('#scanCount').textContent = S.scanLog.filter(x => !x.error).length + ' tercatat sesi ini'; }
  $('#scanIn')?.focus();
}
ACT.camStart = async () => {
  const v = $('#scanVideo');
  v.hidden = false;
  try { await Scanner.start(v, code => doScan(code)); } catch (e) { v.hidden = true; toast(e.message, 'bad'); }
};

const ST5 = [['hadir', 'Hadir'], ['terlambat', 'Telat'], ['izin', 'Izin'], ['sakit', 'Sakit'], ['alfa', 'Alfa']];
async function absenCeklist(ks, kid) {
  const tgl = /^\d{4}-\d{2}-\d{2}$/.test(S.params.tanggal || '') ? S.params.tanggal : today();
  const rows = await GET(`/absen/harian?kelas_id=${kid}&tanggal=${tgl}`);
  S.ck = { kid, tgl, rows };
  const n = rows.filter(r => r.status).length;
  return `<div class="bar">${kelasSelect(ks, kid, 'ckKelas', '')}<input type="date" value="${tgl}" data-change="ckTgl" aria-label="Tanggal" style="min-width:0"></div>
    <div class="card-head"><div>${tag(`${n}/${rows.length} tercatat`, n === rows.length && n ? 'ok' : 'warn')}</div>
      <div class="acts">${btn('Semua hadir', 'ckAll', { st: 'hadir' }, 'ghost sm')}${btn('Sisanya alfa', 'ckRest', {}, 'ghost sm')}${btn('Simpan', 'ckSave', {}, 'sm', 'check')}</div></div>
    ${rows.length ? `<div class="tw" id="ckBox"><table><thead><tr><th>Siswa</th><th>Jam</th><th>Kehadiran</th><th>Keterangan</th></tr></thead><tbody>
    ${rows.map(r => `<tr data-sid="${r.id}"><td><b>${esc(r.nama_lengkap)}</b><small class="muted">${esc(r.username)}</small></td>
      <td><small>${r.jam_masuk ? esc(r.jam_masuk) + (r.jam_pulang ? '–' + esc(r.jam_pulang) : '') : '–'}</small>${r.metode && r.metode !== 'manual' ? tag(r.metode === 'qr' ? 'QR' : 'Selfie', 'info') : ''}
        ${r.foto_absen ? `<a href="${esc(r.foto_absen)}" target="_blank" rel="noopener" style="font-size:12px">foto</a>` : ''}</td>
      <td><div class="seg">${ST5.map(([s, l]) => `<label><input type="radio" name="ck_${r.id}" value="${s}" ${r.status === s ? 'checked' : ''}><span class="${s === 'terlambat' ? 'sakit' : s}">${l}</span></label>`).join('')}</div></td>
      <td><input name="ket" value="${esc(r.keterangan || '')}" placeholder="Opsional" aria-label="Keterangan"></td></tr>`).join('')}
    </tbody></table></div>` : `<div class="card">${empty('Belum ada siswa di kelas ini.', 'users')}</div>`}`;
}
CHG.ckKelas = el => go('absenharian', { tab: 'ceklist', kelas: el.value, tanggal: S.ck.tgl });
CHG.ckTgl = el => go('absenharian', { tab: 'ceklist', kelas: S.ck.kid, tanggal: el.value });
ACT.ckAll = () => $$('#ckBox input[value=hadir]').forEach(i => { i.checked = true; });
ACT.ckRest = () => $$('#ckBox tr[data-sid]').forEach(tr => { if (!$('input:checked', tr)) $('input[value=alfa]', tr).checked = true; });
ACT.ckSave = async el => {
  const items = $$('#ckBox tr[data-sid]').map(tr => ({ siswa_id: Number(tr.dataset.sid), status: ($('input[type=radio]:checked', tr) || {}).value || '', keterangan: $('input[name=ket]', tr).value }));
  el.disabled = true;
  try { await PUT('/absen/harian', { kelas_id: S.ck.kid, tanggal: S.ck.tgl, items }); toast('Absensi tersimpan.'); await refresh(true); } finally { el.disabled = false; }
};

const STL = { hadir: 'H', terlambat: 'T', izin: 'I', sakit: 'S', alfa: 'A' };
async function absenRekap(ks, kid) {
  const bulan = /^\d{4}-\d{2}$/.test(S.params.bulan || '') ? S.params.bulan : monthNow();
  const r = await GET(`/absen/rekap?kelas_id=${kid}&bulan=${bulan}`);
  const [y, m] = bulan.split('-').map(Number);
  const days = new Date(y, m, 0).getDate();
  const libur = d => { const s = `${bulan}-${pad(d)}`; return r.libur.some(e => s >= e.mulai && s <= e.selesai); };
  const by = {};
  r.records.forEach(x => { (by[x.siswa_id] = by[x.siswa_id] || {})[Number(x.tanggal.slice(8))] = x.status; });
  const kelas = ks.find(k => k.id === kid);
  const rows = r.students.map(s => {
    const rec = by[s.id] || {};
    const cnt = { hadir: 0, terlambat: 0, izin: 0, sakit: 0, alfa: 0 };
    Object.values(rec).forEach(v => cnt[v]++);
    const tot = Object.values(cnt).reduce((a, b) => a + b, 0);
    return { s, rec, cnt, pct: tot ? (cnt.hadir + cnt.terlambat) / tot * 100 : null };
  });
  const table = `<table class="mgrid"><thead><tr><th>Nama</th>${Array.from({ length: days }, (_, i) => `<th>${i + 1}</th>`).join('')}<th>H</th><th>T</th><th>I</th><th>S</th><th>A</th><th>%</th></tr></thead><tbody>
    ${rows.map(x => `<tr><td>${esc(x.s.nama_lengkap)}</td>${Array.from({ length: days }, (_, i) => {
      const d = i + 1, st = x.rec[d], wd = new Date(y, m - 1, d).getDay();
      return `<td class="${st ? 'st-' + STL[st] : ''} ${!st && (wd === 0 || libur(d)) ? 'st-L' : ''}">${st ? STL[st] : ''}</td>`; }).join('')}
      <td>${x.cnt.hadir}</td><td>${x.cnt.terlambat}</td><td>${x.cnt.izin}</td><td>${x.cnt.sakit}</td><td>${x.cnt.alfa}</td><td>${x.pct == null ? '–' : fmtN(x.pct)}</td></tr>`).join('')}
    </tbody></table>`;
  S.rekapAbsen = { html: schoolHead() + `<h3 style="text-align:center">REKAP ABSENSI HARIAN · KELAS ${esc(kelas?.nama || '')} · ${esc(monthLabel(bulan).toUpperCase())}</h3>` + table +
    `<p style="font-size:9pt">H hadir · T terlambat · I izin · S sakit · A alfa · kolom merah = Minggu/libur</p>`,
    csv: toCSV([['Nama', ...Array.from({ length: days }, (_, i) => i + 1), 'H', 'T', 'I', 'S', 'A'], ...rows.map(x => [x.s.nama_lengkap, ...Array.from({ length: days }, (_, i) => STL[x.rec[i + 1]] || ''), x.cnt.hadir, x.cnt.terlambat, x.cnt.izin, x.cnt.sakit, x.cnt.alfa])]),
    name: `absensi-${kelas?.nama || ''}-${bulan}.csv` };
  return `<div class="bar">${kelasSelect(ks, kid, 'rkKelas', '')}<input type="month" value="${bulan}" data-change="rkBulan" aria-label="Bulan" style="min-width:0">
      <div class="acts" style="margin-left:auto">${btn('Unduh CSV', 'rkCSV', {}, 'ghost sm', 'download')}${btn('Cetak', 'rkPrint', {}, 'sm', 'print')}</div></div>
    ${rows.length ? `<div class="tw no-stack">${table}</div><p class="muted" style="font-size:12px;margin-top:8px">H hadir · T terlambat · I izin · S sakit · A alfa · sel merah muda = Minggu/hari libur (dari Kalender)</p>` : `<div class="card">${empty('Belum ada siswa di kelas ini.', 'users')}</div>`}`;
}
CHG.rkKelas = el => go('absenharian', { tab: 'rekap', kelas: el.value, bulan: S.params.bulan });
CHG.rkBulan = el => go('absenharian', { tab: 'rekap', kelas: S.params.kelas, bulan: el.value });
ACT.rkPrint = () => printDoc(S.rekapAbsen.html);
ACT.rkCSV = () => downloadText(S.rekapAbsen.name, S.rekapAbsen.csv);

// =====================================================================
// KARTU SISWA (cetak)
// =====================================================================
function idCard(s) {
  return `<div class="idcard">${qrImg(s.token, 128)}<div class="t"><div class="sch">${esc(S.settings.nama_sekolah || '')}</div>
    <b>${esc(s.nama_lengkap)}</b>NIS ${esc(s.username)}${s.nisn ? '<br>NISN ' + esc(s.nisn) : ''}<br>Kelas ${esc(s.kelas_nama || '–')}</div></div>`;
}
async function pageKartu() {
  const ks = await GET('/classes');
  const kid = firstKelas(ks, S.params.kelas);
  const rows = kid ? await GET('/kartu?kelas_id=' + kid) : [];
  S.kartu = rows;
  return head('Kartu Siswa', 'Kartu berisi QR bertanda tangan digital untuk absensi dan perpustakaan. Cetak di kertas A4, potong, lalu laminating.',
      rows.length ? btn('Cetak kartu', 'kartuPrint', {}, '', 'print') : '') +
    `<div class="bar">${kelasSelect(ks, kid, 'kartuKelas', '')}</div>` +
    (rows.length ? `<div class="cards screen">${rows.map(idCard).join('')}</div>` : `<div class="card">${empty('Pilih kelas yang memiliki siswa.', 'users')}</div>`);
}
CHG.kartuKelas = el => go('kartu', { kelas: el.value });
ACT.kartuPrint = () => {
  const pages = [];
  for (let i = 0; i < S.kartu.length; i += 10) pages.push(`<div class="pg"><div class="cards">${S.kartu.slice(i, i + 10).map(idCard).join('')}</div></div>`);
  printDoc(pages.join(''));
};

// =====================================================================
// KALENDER AKADEMIK
// =====================================================================
const KAT = [['kegiatan', 'Kegiatan'], ['libur', 'Libur'], ['ujian', 'Ujian'], ['rapat', 'Rapat'], ['lainnya', 'Lainnya']];
const KAT_TAMPIL = [...KAT, ['tugas', 'Tenggat tugas']];
const katLabel = k => (KAT_TAMPIL.find(x => x[0] === k) || [0, k])[1];
const katTag = e => tag(katLabel(e.kategori), e.kategori === 'libur' ? 'bad' : e.kategori === 'ujian' ? 'warn' : e.kategori === 'tugas' ? 'info' : 'brand');
const autoTag = e => e.otomatis ? `<small class="muted" title="Diambil otomatis dari modul ${esc(e.halaman || '')}">otomatis</small>` : '';
// pengelola boleh mengubah kegiatan kalender biasa; kegiatan otomatis diubah di modul asalnya
const canEditEv = e => isAdm('kalender') && !e.otomatis;
const evRow = e => `<div class="li"><div><b>${esc(e.judul)}</b><small class="muted">${fmtDate(e.mulai)}${e.selesai !== e.mulai ? ' – ' + fmtDate(e.selesai) : ''}${e.keterangan ? ' · ' + esc(e.keterangan) : ''}</small></div>
  <div class="acts">${katTag(e)}${autoTag(e)}${canEditEv(e) ? iconBtn('edit', 'evForm', { id: e.id }, 'Ubah') : e.otomatis && e.halaman && pageFn(S.user.role, e.halaman) ? btn('Buka', 'go', { page: e.halaman, params: '{}' }, 'ghost sm') : ''}</div></div>`;
async function pageKalender() {
  const ym = /^\d{4}-\d{2}$/.test(S.params.bulan || '') ? S.params.bulan : monthNow();
  const [y, m] = ym.split('-').map(Number);
  const first = new Date(y, m - 1, 1);
  const start = new Date(first); start.setDate(1 - ((first.getDay() + 6) % 7)); // mulai Senin
  const end = new Date(start); end.setDate(start.getDate() + 41);
  while (new Date(end.getFullYear(), end.getMonth(), end.getDate() - 6).getMonth() !== m - 1) end.setDate(end.getDate() - 7);
  const iso = d => `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
  const evs = await GET(`/kalender?dari=${iso(start)}&sampai=${iso(end)}`);
  S.events = Object.fromEntries(evs.map(e => [e.id, e]));
  const prev = new Date(y, m - 2, 1), next = new Date(y, m, 1);
  const ymOf = d => `${d.getFullYear()}-${pad(d.getMonth() + 1)}`;
  const cells = [];
  for (let d = new Date(start); d <= end; d.setDate(d.getDate() + 1)) {
    const s = iso(d);
    const list = evs.filter(e => s >= e.mulai && s <= e.selesai);
    // seluruh kotak tanggal bisa diketuk (daftar kegiatan hari itu); di HP kegiatan tampil sebagai titik berwarna
    cells.push(`<div class="day ${d.getMonth() !== m - 1 ? 'out' : ''} ${s === today() ? 'today' : ''} ${list.length ? 'has' : ''}" data-act="calDay" data-tgl="${s}" role="button" tabindex="0"
      aria-label="${esc(fmtDate(s))}${list.length ? `, ${list.length} kegiatan` : ''}"><div class="n" style="${d.getDay() === 0 ? 'color:var(--bad)' : ''}">${d.getDate()}</div>
      ${list.slice(0, 3).map(e => `<span class="ev ${esc(e.kategori)} ${e.otomatis ? 'auto' : ''}" data-act="evOpen" data-id="${esc(e.id)}" title="${esc(e.judul)}${e.otomatis ? ' (otomatis)' : ''}">${esc(e.judul)}</span>`).join('')}${list.length > 3 ? `<small class="ev-more">+${list.length - 3} lainnya</small>` : ''}
      ${list.length ? `<div class="dots">${list.slice(0, 3).map(e => `<i class="dot ${esc(e.kategori)}"></i>`).join('')}${list.length > 3 ? `<small>+${list.length - 3}</small>` : ''}</div>` : ''}</div>`);
  }
  const month = evs.filter(e => e.selesai >= `${ym}-01` && e.mulai <= `${ym}-31`);
  return head('Kalender Akademik', monthLabel(ym), btn('Cetak', 'calPrint', {}, 'ghost', 'print') +
      (isAdm('kalender') ? btn('Libur nasional', 'calLibur', {}, 'ghost', 'cal') + btn('Tambah kegiatan', 'evForm', {}, '', 'plus') : '')) +
    `<div class="bar">${btn('', 'go', { page: 'kalender', params: JSON.stringify({ bulan: ymOf(prev) }) }, 'ghost icon', 'back').replace('<button', '<button aria-label="Bulan sebelumnya"')}
      <input type="month" value="${ym}" data-change="calMonth" aria-label="Bulan" style="min-width:0">
      ${btn('', 'go', { page: 'kalender', params: JSON.stringify({ bulan: ymOf(next) }) }, 'ghost icon', 'back').replace('<svg', '<svg style="transform:rotate(180deg)"').replace('<button', '<button aria-label="Bulan berikutnya"')}
      ${btn('Bulan ini', 'go', { page: 'kalender', params: '{}' }, 'ghost sm')}
      <div class="cal-legend">${(evs.some(e => e.kategori === 'tugas') ? KAT_TAMPIL : KAT).map(([k, l]) => `<span><i class="dot ${k}"></i>${l}</span>`).join('')}</div></div>
    <div class="cal">${['Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab', 'Min'].map(d => `<div class="dow">${d}</div>`).join('')}${cells.join('')}</div>
    <section class="card" style="margin-top:16px"><div class="card-head"><h3>Agenda ${esc(monthLabel(ym))}</h3></div>
      ${month.length ? month.map(evRow).join('') : empty('Tidak ada agenda bulan ini.', 'cal')}</section>`;
}
CHG.calMonth = el => go('kalender', { bulan: el.value });
// ketuk tanggal: daftar kegiatan hari itu (+ tambah kegiatan untuk pengelola)
ACT.calDay = el => {
  const tgl = el.dataset.tgl;
  const list = Object.values(S.events).filter(e => tgl >= e.mulai && tgl <= e.selesai);
  const hari = new Date(tgl + 'T00:00:00').toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' });
  openModal({ title: hari, body: (list.length ? list.map(evRow).join('')
    : `<p class="muted">Tidak ada kegiatan pada tanggal ini.</p>`) +
    (isAdm('kalender') ? `<div class="acts" style="justify-content:flex-start;margin-top:14px">${btn('Tambah kegiatan pada tanggal ini', 'evForm', { tgl }, '', 'plus')}</div>` : '') });
};
// ---- Libur nasional tanggal tetap (17 Agustus, 25 Desember, 1 Januari, 1 Mei, 1 Juni)
ACT.calLibur = async () => {
  const ta = S.settings.tahun_ajaran || '';
  if (!(await confirmBox(`Tambahkan libur nasional bertanggal tetap untuk tahun ajaran <b>${esc(ta)}</b>?<br><small class="muted">17 Agustus, 25 Desember, 1 Januari, 1 Mei, 1 Juni. Libur yang tanggalnya ditetapkan pemerintah tiap tahun (Idulfitri, Iduladha, Nyepi, Imlek, Waisak, dll.) tetap diisi manual sesuai SKB.</small>`, 'Tambahkan', ''))) return;
  const r = await POST('/kalender/libur-nasional', { tahun_ajaran: ta });
  toast(r.ditambahkan ? `${r.ditambahkan} libur nasional ditambahkan untuk ${r.tahun_ajaran}.` : 'Semua libur nasional tanggal tetap sudah ada.');
  await refresh(true);
};

// ---- Cetak kalender per semester / satu tahun ajaran (kisi bulan + daftar kegiatan)
ACT.calPrint = () => {
  const ta = /^\d{4}\/\d{4}$/.test(S.settings.tahun_ajaran || '') ? S.settings.tahun_ajaran : (() => { const y = new Date().getFullYear(), mo = new Date().getMonth() + 1; return mo >= 7 ? `${y}/${y + 1}` : `${y - 1}/${y}`; })();
  openForm({
    title: 'Cetak Kalender Akademik', submit: 'Cetak', success: 'Siap dicetak.',
    values: { periode: new Date().getMonth() + 1 >= 7 ? 'ganjil' : 'genap', otomatis: true, ta },
    fields: [{ n: 'ta', l: 'Tahun ajaran', t: 'select', o: [[ta, ta]] },
      { n: 'periode', l: 'Periode', t: 'select', o: [['ganjil', 'Semester Ganjil (Juli–Desember)'], ['genap', 'Semester Genap (Januari–Juni)'], ['tahun', 'Satu tahun ajaran (Juli–Juni)']] },
      { n: 'otomatis', l: 'Sertakan ujian, PPDB, e-voting, dan pengumuman dari modul lain', t: 'check' }],
    onSubmit: async v => {
      const y1 = Number(v.ta.slice(0, 4));
      const bulan = v.periode === 'ganjil' ? [[y1, 7], [y1, 8], [y1, 9], [y1, 10], [y1, 11], [y1, 12]]
        : v.periode === 'genap' ? [[y1 + 1, 1], [y1 + 1, 2], [y1 + 1, 3], [y1 + 1, 4], [y1 + 1, 5], [y1 + 1, 6]]
        : [7, 8, 9, 10, 11, 12].map(m => [y1, m]).concat([1, 2, 3, 4, 5, 6].map(m => [y1 + 1, m]));
      const dari = `${bulan[0][0]}-${pad(bulan[0][1])}-01`, sampai = `${bulan[bulan.length - 1][0]}-${pad(bulan[bulan.length - 1][1])}-31`;
      const evs = (await GET(`/kalender?dari=${dari}&sampai=${sampai}&otomatis=${v.otomatis ? 1 : 0}`)).filter(e => e.kategori !== 'tugas');
      const judul = v.periode === 'ganjil' ? 'Semester Ganjil' : v.periode === 'genap' ? 'Semester Genap' : 'Satu Tahun Ajaran';
      setTimeout(() => printDoc(calPrintHtml(bulan, evs, v.ta, judul)), 250);
    },
  });
};
const PRIO = ['libur', 'ujian', 'kegiatan', 'rapat', 'lainnya'];
function calPrintHtml(bulan, evs, ta, judul) {
  const s = S.settings;
  const iso = (y, m, d) => `${y}-${pad(m)}-${pad(d)}`;
  const mini = ([y, m]) => {
    const first = new Date(y, m - 1, 1), days = new Date(y, m, 0).getDate(), lead = (first.getDay() + 6) % 7;
    let cells = '';
    for (let i = 0; i < lead; i++) cells += '<td></td>';
    for (let d = 1; d <= days; d++) {
      const t = iso(y, m, d), on = evs.filter(e => t >= e.mulai && t <= e.selesai);
      const k = PRIO.find(p => on.some(e => e.kategori === p));
      const sun = (lead + d - 1) % 7 === 6;
      cells += `<td class="${k ? 'k-' + k : ''}${sun ? ' sun' : ''}">${d}</td>` + ((lead + d) % 7 === 0 && d < days ? '</tr><tr>' : '');
    }
    return `<div class="cp-m"><b>${esc(MONTHS[m - 1])} ${y}</b><table><thead><tr>${['S', 'S', 'R', 'K', 'J', 'S', 'M'].map(x => `<th>${x}</th>`).join('')}</tr></thead><tbody><tr>${cells}</tr></tbody></table></div>`;
  };
  return `<style>
    #printArea .cp-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 10px 14px; margin: 10px 0 14px; }
    #printArea .cp-m b { display: block; text-align: center; font-size: 10.5pt; margin-bottom: 3px; }
    #printArea .cp-m table { width: 100%; border-collapse: collapse; table-layout: fixed; }
    #printArea .cp-m th, #printArea .cp-m td { border: 1px solid #bbb; padding: 2px 0; text-align: center; font-size: 8.5pt; }
    #printArea .cp-m th { background: #eee; }
    #printArea .cp-m td.sun { color: #c00; font-weight: 700; }
    #printArea .k-libur { background: #fecaca !important; } #printArea .k-ujian { background: #fde68a !important; }
    #printArea .k-kegiatan { background: #bfdbfe !important; } #printArea .k-rapat { background: #ddd6fe !important; } #printArea .k-lainnya { background: #e5e7eb !important; }
    #printArea .cp-leg span { display: inline-block; margin-right: 14px; font-size: 9pt; } #printArea .cp-leg i { display: inline-block; width: 11px; height: 11px; border: 1px solid #999; vertical-align: -1px; margin-right: 4px; }
    #printArea * { -webkit-print-color-adjust: exact; print-color-adjust: exact; }
  </style>${schoolHead()}
  <h3 style="text-align:center;margin:0">KALENDER AKADEMIK TAHUN AJARAN ${esc(ta)}</h3><p style="text-align:center;margin:2px 0 0">${esc(judul)}</p>
  <div class="cp-grid">${bulan.map(mini).join('')}</div>
  <div class="cp-leg">${KAT.map(([k, l]) => `<span><i class="k-${k}"></i>${l}</span>`).join('')}<span style="color:#c00;font-weight:700">Merah: hari Minggu</span></div>
  <table style="margin-top:10px"><thead><tr><th style="width:28%">Tanggal</th><th>Kegiatan</th><th style="width:16%">Kategori</th></tr></thead><tbody>
  ${evs.length ? evs.map(e => `<tr><td>${fmtDate(e.mulai)}${e.selesai !== e.mulai ? ' – ' + fmtDate(e.selesai) : ''}</td><td>${esc(e.judul)}${e.keterangan ? `<br><small>${esc(e.keterangan)}</small>` : ''}</td><td>${esc(katLabel(e.kategori))}</td></tr>`).join('')
    : '<tr><td colspan="3" style="text-align:center">Belum ada kegiatan.</td></tr>'}</tbody></table>
  <table style="width:100%;margin-top:22px;border:0"><tr><td style="border:0;width:55%"></td><td style="border:0;text-align:center">${esc(fmtDate(today()))}<br>Kepala Sekolah<br><br><br><br><b>${esc(s.kepala_sekolah || '........................')}</b><br>${s.nip_kepala ? 'NIP ' + esc(s.nip_kepala) : ''}</td></tr></table>`;
}

ACT.evOpen = el => {
  const e = S.events[el.dataset.id];
  if (!e) return;
  if (canEditEv(e)) return ACT.evForm(el);
  const link = e.otomatis && e.halaman && pageFn(S.user.role, e.halaman);
  openModal({ title: e.judul, body: `<p>${fmtDate(e.mulai)}${e.selesai !== e.mulai ? ' – ' + fmtDate(e.selesai) : ''} ${katTag(e)}</p>${e.keterangan ? `<p class="muted" style="margin-top:8px">${esc(e.keterangan)}</p>` : ''}
    ${e.otomatis ? `<p class="muted" style="margin-top:10px;font-size:13px">Diambil otomatis dari modul terkait; ubah di modul tersebut.</p>` : ''}
    ${link ? `<div class="acts" style="justify-content:flex-start;margin-top:12px">${btn('Buka halaman', 'go', { page: e.halaman, params: '{}' }, '', 'eye')}</div>` : ''}` });
};
ACT.evForm = el => {
  const e = el.dataset.id ? S.events[el.dataset.id] : { kategori: 'kegiatan', mulai: el.dataset.tgl || today(), selesai: el.dataset.tgl || '' };
  while (MODALS.length) closeModal(); // dari daftar kegiatan sehari: ganti dengan formulir
  openForm({
    title: e.id ? 'Ubah kegiatan' : 'Tambah kegiatan', values: e,
    fields: [{ n: 'judul', l: 'Nama kegiatan', req: true }, [{ n: 'mulai', l: 'Mulai', t: 'date', req: true }, { n: 'selesai', l: 'Selesai', t: 'date' }],
      { n: 'kategori', l: 'Kategori', t: 'select', o: KAT, hint: 'Kategori "Libur" otomatis ditandai di rekap absensi.' }, { n: 'keterangan', l: 'Keterangan', t: 'textarea', rows: 2 }],
    onSubmit: v => e.id ? PUT('/kalender/' + e.id, v) : POST('/kalender', v),
  });
};
ACT.evDel = el => del('/kalender/' + el.dataset.id, 'Hapus kegiatan ini dari kalender?');

// =====================================================================
// SISWA: Absensi & kartu saya, absen selfie
// =====================================================================
async function pageAbsenku() {
  const bulan = /^\d{4}-\d{2}$/.test(S.params.bulan || '') ? S.params.bulan : monthNow();
  const d = await GET('/absen/me?bulan=' + bulan);
  const t = d.today;
  const cnt = { hadir: 0, terlambat: 0, izin: 0, sakit: 0, alfa: 0 };
  d.records.forEach(r => cnt[r.status]++);
  const selfie = d.config.selfie;
  return head('Absensi & Kartu', 'Tunjukkan QR ini ke petugas di gerbang, atau absen selfie jika diizinkan sekolah.') +
    `<div class="cols"><section class="card" style="text-align:center">
      <div class="big-qr">${qrImg(d.token, 240)}</div>
      <h2 style="margin-top:12px">${esc(S.user.nama_lengkap)}</h2><p class="muted">${esc(S.user.username)}</p>
      <p class="muted" style="font-size:12px;margin-top:6px">Perbesar kecerahan layar saat di-scan.</p></section>
    <section class="card"><div class="card-head"><h3>Hari ini</h3></div>
      ${t ? `<div class="scan-result ${t.status === 'hadir' ? 'ok' : 'warn'}"><div><b>${esc(t.status[0].toUpperCase() + t.status.slice(1))}</b>Masuk ${esc(t.jam_masuk || '–')}${t.jam_pulang ? ' · Pulang ' + esc(t.jam_pulang) : ''}</div></div>`
        : `<div class="scan-result warn"><div><b>Belum absen</b>Batas jam masuk ${esc(d.config.jam_masuk || '–')}</div></div>`}
      ${selfie ? `<div style="margin-top:14px">${btn(t && !t.jam_pulang ? 'Absen pulang (selfie)' : 'Absen selfie', 'selfieOpen', {}, '', 'eye')}</div>
        <p class="muted" style="font-size:12px;margin-top:8px">Butuh kamera & lokasi. Harus berada di area sekolah.</p>` : ''}
      <div class="stats" style="margin-top:16px;grid-template-columns:repeat(auto-fill,minmax(90px,1fr))">${[['hadir', 'Hadir'], ['terlambat', 'Telat'], ['izin', 'Izin'], ['sakit', 'Sakit'], ['alfa', 'Alfa']].map(([k, l]) =>
        `<div class="stat" style="padding:10px;flex-direction:column;gap:2px;text-align:center"><b>${cnt[k]}</b><small>${l}</small></div>`).join('')}</div></section></div>
    <section class="card" style="margin-top:16px"><div class="card-head"><h3>Riwayat</h3><input type="month" value="${bulan}" data-change="akBulan" aria-label="Bulan" style="width:auto"></div>
      ${d.records.length ? `<div class="tw"><table><thead><tr><th>Tanggal</th><th>Status</th><th>Masuk</th><th>Pulang</th><th>Keterangan</th></tr></thead><tbody>
      ${d.records.map(r => `<tr><td>${fmtDate(r.tanggal)}</td><td>${tag(r.status, r.status === 'hadir' ? 'ok' : r.status === 'alfa' ? 'bad' : 'warn')}</td><td>${esc(r.jam_masuk || '–')}</td><td>${esc(r.jam_pulang || '–')}</td><td>${esc(r.keterangan || '')}</td></tr>`).join('')}
      </tbody></table></div>` : empty('Belum ada catatan bulan ini.', 'cal')}</section>`;
}
CHG.akBulan = el => go('absenku', { bulan: el.value });

ACT.selfieOpen = () => {
  if (!window.isSecureContext) { toast('Absen selfie butuh HTTPS. Buka aplikasi lewat alamat https:// dari sekolah.', 'bad'); return; }
  const m = openModal({ title: 'Absen selfie', body: `<div class="scan-box"><video class="scan-video" id="selfieV" playsinline muted autoplay style="transform:scaleX(-1)"></video>
    <p class="muted" id="selfieLoc">Mengambil lokasi…</p><div class="sheet-foot">${btn('Batal', 'selfieClose', {}, 'ghost')}${btn('Ambil foto & absen', 'selfieShot', {}, '', 'check')}</div></div>` });
  S.selfie = { lat: null, lng: null, stream: null };
  openCamera($('#selfieV'), { facingMode: 'user', width: 640 }).then(st => { if (st) S.selfie.stream = st; })
    .catch(() => { const el = $('#selfieLoc'); if (el) el.textContent = 'Kamera tidak bisa dibuka. Izinkan akses kamera.'; });
  navigator.geolocation.getCurrentPosition(p => { S.selfie.lat = p.coords.latitude; S.selfie.lng = p.coords.longitude; S.selfie.akurasi = p.coords.accuracy; const el = $('#selfieLoc'); if (el) el.textContent = `Lokasi terbaca (akurasi ±${Math.round(p.coords.accuracy)} m).`; },
    () => { const el = $('#selfieLoc'); if (el) el.textContent = 'Lokasi tidak terbaca. Izinkan akses lokasi.'; }, { enableHighAccuracy: true, timeout: 15000 });
  return m;
};
const stopSelfie = () => { stopStream(S.selfie?.stream, $('#selfieV')); if (S.selfie) S.selfie.stream = null; };
ACT.selfieClose = () => { stopSelfie(); closeModal(); };
ACT.selfieShot = async el => {
  const v = $('#selfieV');
  const foto = snapshot(v);
  if (!foto) return toast('Kamera belum siap.', 'bad');
  if (S.selfie.lat == null) return toast('Tunggu lokasi terbaca terlebih dahulu.', 'bad');
  el.disabled = true;
  try {
    const r = await POST('/absen/selfie', { foto, lat: S.selfie.lat, lng: S.selfie.lng, akurasi: S.selfie.akurasi });
    stopSelfie(); closeModal();
    toast(r.aksi === 'sudah' ? 'Anda sudah absen.' : `Absen ${r.aksi} tercatat pukul ${r.aksi === 'pulang' ? r.jam_pulang : r.jam_masuk}.`);
    refresh(true);
  } catch (e) { toast(e.message, 'bad'); } finally { el.disabled = false; }
};

Object.assign(PAGES.admin, { bukuinduk: pageBukuInduk, absenharian: pageAbsenHarian, kartu: pageKartu, kalender: pageKalender });
Object.assign(PAGES.guru, { bukuinduk: pageBukuInduk, absenharian: pageAbsenHarian, kalender: pageKalender });
Object.assign(PAGES.siswa, { absenku: pageAbsenku, kalender: pageKalender });
