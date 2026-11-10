'use strict';
/* Administrasi Guru & Wali Kelas: jurnal mengajar, catatan siswa (prestasi/pelanggaran/BK), halaman wali kelas */

addNav('admin', ['jurnal', 'Jurnal Mengajar', 'clip', 'admguru', 'Administrasi Guru']);
addNav('admin', ['catatan', 'Catatan Siswa', 'check', 'admguru', 'Administrasi Guru']);
addNav('admin', ['walikelas', 'Wali Kelas', 'users', 'admguru', 'Administrasi Guru']);
addNav('guru', ['jurnal', 'Jurnal Mengajar', 'clip', 'admguru', 'Administrasi Guru']);
addNav('guru', ['catatan', 'Catatan Siswa', 'check', 'admguru', 'Administrasi Guru']);
addNav('guru', ['walikelas', 'Wali Kelas', 'users', 'admguru', 'Administrasi Guru']);
addNav('siswa', ['catatanku', 'Prestasi & Poin', 'check', 'admguru', 'Kesiswaan']);

const JENIS_CAT = { prestasi: ['Prestasi', 'ok'], pelanggaran: ['Pelanggaran', 'bad'], catatan: ['Catatan wali', ''], bk: ['Catatan BK', 'warn'] };
const KATEGORI_CAT = {
  prestasi: ['Juara lomba', 'Aktif di kelas', 'Membantu teman', 'Pengurus kegiatan', 'Hafalan / literasi', 'Kebersihan kelas'],
  pelanggaran: ['Terlambat', 'Tidak berseragam lengkap', 'Tidak mengerjakan tugas', 'Membolos', 'Bermain HP saat pelajaran', 'Berkata kasar', 'Berkelahi', 'Merokok', 'Merusak fasilitas'],
  catatan: ['Perkembangan belajar', 'Kesehatan', 'Komunikasi orang tua', 'Kehadiran'],
  bk: ['Konseling individu', 'Konseling kelompok', 'Panggilan orang tua', 'Home visit', 'Rujukan'],
};
const JAM_KE = ['1', '2', '3', '4', '5', '6', '7', '8', '1-2', '3-4', '5-6', '7-8'];
const monthInput = (val, change) => `<input type="month" value="${esc(val)}" data-change="${change}" aria-label="Bulan" style="min-width:0">`;
const canEditRow = r => S.user.role === 'admin' || r.created_by === S.user.id || r.guru_id === S.user.id;

// ---------------------------------------------------------------------
// Jurnal mengajar
// ---------------------------------------------------------------------
const validBulan = v => /^\d{4}-(0[1-9]|1[0-2])$/.test(v || '') ? v : '';
async function pageJurnal() {
  const bulan = validBulan(S.params.bulan) || monthNow();
  let rows = await GET('/jurnal?bulan=' + encodeURIComponent(bulan));
  const gurus = uniqList(rows.map(r => r.guru_nama));
  const semua = gurus.length > 1 || (rows.length && rows[0].guru_id !== S.user.id);
  if (S.params.guru) rows = rows.filter(r => r.guru_nama === S.params.guru);
  S.jurnal = Object.fromEntries(rows.map(r => [r.id, r]));
  const kh = r => r.course_id && (r.hadir || r.tidak_hadir) ? `${r.hadir} hadir${r.tidak_hadir ? `, <b>${r.tidak_hadir} tidak</b>` : ''}` : '';
  return head('Jurnal Mengajar', `Catatan kegiatan mengajar ${monthLabel(bulan)}. Kehadiran terisi otomatis bila absensi mapel diisi.`,
    btn('Isi jurnal', 'jurForm', {}, '', 'plus')) +
    `<div class="bar no-print">${monthInput(bulan, 'jurBulan')}
      ${semua ? `<label class="sel">Guru <select data-change="jurGuru"><option value="">Semua guru</option>${opts(gurus.map(g => [g, g]), S.params.guru || '')}</select></label>` : ''}
      <small class="muted">${rows.length} entri</small><span style="flex:1"></span>
      ${rows.length ? btn('Cetak', 'jurPrint', {}, 'ghost sm', 'print') + btn('CSV', 'jurCSV', {}, 'ghost sm', 'download') : ''}</div>` +
    (rows.length ? `<div class="tw"><table><thead><tr><th>Tanggal</th><th>Jam</th>${semua ? '<th>Guru</th>' : ''}<th>Mapel / kelas</th><th>Materi & kegiatan</th><th>Kehadiran</th><th></th></tr></thead><tbody>
    ${rows.map(r => `<tr><td>${fmtDate(r.tanggal)}</td><td>${esc(r.jam_ke || '–')}</td>${semua ? `<td>${esc(r.guru_nama)}</td>` : ''}
      <td>${r.mapel ? `<b>${esc(r.mapel)}</b><small class="muted">${esc(r.kelas_nama || '')}</small>` : '<small class="muted">–</small>'}</td>
      <td><b>${esc(r.materi)}</b>${r.kegiatan ? `<small class="muted">${esc(r.kegiatan)}</small>` : ''}${r.catatan ? `<small>Catatan: ${esc(r.catatan)}</small>` : ''}</td>
      <td><small>${kh(r)}${r.kehadiran ? (kh(r) ? '<br>' : '') + esc(r.kehadiran) : ''}${!kh(r) && !r.kehadiran ? '–' : ''}</small></td>
      <td><div class="acts">${canEditRow(r) ? iconBtn('edit', 'jurForm', { id: r.id }, 'Ubah') + iconBtn('trash', 'jurDel', { id: r.id }, 'Hapus', 'danger') : ''}</div></td></tr>`).join('')}
    </tbody></table></div>` : `<div class="card">${empty('Belum ada jurnal bulan ini. Klik "Isi jurnal" setelah mengajar.', 'clip')}</div>`);
}
CHG.jurBulan = el => go('jurnal', { ...S.params, bulan: el.value });
CHG.jurGuru = el => go('jurnal', { ...S.params, guru: el.value });
ACT.jurForm = async el => {
  const r = el.dataset.id ? S.jurnal[el.dataset.id] : { tanggal: today() };
  const cs = (await GET('/courses')).filter(c => S.user.role === 'admin' || c.guru_id === S.user.id || c.guru_id == null);
  openForm({
    title: r.id ? 'Ubah jurnal' : 'Isi jurnal mengajar', values: r, wide: true,
    fields: [[{ n: 'tanggal', l: 'Tanggal', t: 'date', req: true }, { n: 'jam_ke', l: 'Jam ke', t: 'combo', o: JAM_KE }],
      { n: 'course_id', l: 'Mapel / kelas', t: 'select', o: cs.map(c => [c.id, `${c.nama} · ${c.kelas_nama || ''}`]), blank: '— Tanpa mapel (kegiatan lain) —' },
      { n: 'materi', l: 'Materi / topik', req: true },
      { n: 'kegiatan', l: 'Kegiatan pembelajaran', t: 'textarea', rows: 2, ph: 'Mis. diskusi kelompok, presentasi, latihan soal' },
      { n: 'kehadiran', l: 'Catatan kehadiran', ph: 'Otomatis dari absensi mapel; atau tulis mis. Andi (S), Budi (I)' },
      { n: 'catatan', l: 'Catatan / kendala / tindak lanjut', t: 'textarea', rows: 2 }],
    onSubmit: v => { const b = { ...v, course_id: v.course_id ? Number(v.course_id) : null }; return r.id ? PUT('/jurnal/' + r.id, b) : POST('/jurnal', b); },
  });
};
ACT.jurDel = el => del('/jurnal/' + el.dataset.id, 'Hapus entri jurnal ini?');
const jurRows = () => Object.values(S.jurnal).sort((a, b) => a.tanggal.localeCompare(b.tanggal));
ACT.jurCSV = () => downloadText(`jurnal-${S.params.bulan || monthNow()}.csv`, toCSV([['Tanggal', 'Jam ke', 'Guru', 'Mapel', 'Kelas', 'Materi', 'Kegiatan', 'Hadir', 'Tidak hadir', 'Catatan kehadiran', 'Catatan'],
  ...jurRows().map(r => [r.tanggal, r.jam_ke, r.guru_nama, r.mapel, r.kelas_nama, r.materi, r.kegiatan, r.hadir, r.tidak_hadir, r.kehadiran, r.catatan])]));
ACT.jurPrint = () => {
  const rows = jurRows(), s = S.settings, bulan = S.params.bulan || monthNow();
  const guru = uniqList(rows.map(r => r.guru_nama));
  printDoc(`<h2 style="text-align:center;margin:0">JURNAL MENGAJAR GURU</h2><p style="text-align:center;margin:4px 0 14px">${esc(s.nama_sekolah || '')} · ${esc(monthLabel(bulan))}</p>
    ${guru.length === 1 ? `<p>Nama guru: <b>${esc(guru[0])}</b></p>` : ''}
    <table><thead><tr><th>No</th><th>Tanggal</th><th>Jam</th>${guru.length > 1 ? '<th>Guru</th>' : ''}<th>Mapel/Kelas</th><th>Materi</th><th>Kegiatan</th><th>Kehadiran</th><th>Catatan</th></tr></thead><tbody>
    ${rows.map((r, i) => `<tr><td>${i + 1}</td><td>${fmtDate(r.tanggal)}</td><td>${esc(r.jam_ke || '')}</td>${guru.length > 1 ? `<td>${esc(r.guru_nama)}</td>` : ''}<td>${esc((r.mapel || '') + (r.kelas_nama ? ' / ' + r.kelas_nama : ''))}</td>
      <td>${esc(r.materi)}</td><td>${esc(r.kegiatan || '')}</td><td>${r.course_id && (r.hadir || r.tidak_hadir) ? `${r.hadir} hadir, ${r.tidak_hadir} tidak` : ''} ${esc(r.kehadiran || '')}</td><td>${esc(r.catatan || '')}</td></tr>`).join('')}</tbody></table>
    ${ttdBlock(guru.length === 1 ? guru[0] : '')}`);
};
// blok tanda tangan: kepala sekolah (kiri) & guru/wali (kanan)
const ttdBlock = (nama, nip = '', jabatan = 'Guru') => `<table style="width:100%;margin-top:28px;border:0"><tr>
  <td style="border:0;width:50%;text-align:center">Mengetahui,<br>Kepala Sekolah<br><br><br><br><b>${esc(S.settings.kepala_sekolah || '........................')}</b><br>NIP ${esc(S.settings.nip_kepala || '')}</td>
  <td style="border:0;width:50%;text-align:center">${esc(fmtDate(today()))}<br>${esc(jabatan)}<br><br><br><br><b>${esc(nama || '........................')}</b><br>${nip ? 'NIP ' + esc(nip) : ''}</td></tr></table>`;

// ---------------------------------------------------------------------
// Catatan siswa
// ---------------------------------------------------------------------
async function pageCatatan() {
  const p = { ...S.params, bulan: validBulan(S.params.bulan) };
  const query = qs({ kelas_id: p.kelas || '', jenis: p.jenis || '', bulan: p.bulan || '', siswa_id: p.siswa || '' });
  S.catQ = '/catatan' + (query ? '?' + query : '');
  const [pg, ref] = await Promise.all([GETpage(S.catQ + (query ? '&' : '?') + qs(pageQ(50))), GET('/catatan/siswa')]);
  const rows = pg.rows;
  if (pageOverflow(pg.total, rows.length)) return '';
  S.catRef = ref; S.catatan = Object.fromEntries(rows.map(r => [r.id, r]));
  const kelas = uniqList(ref.siswa.map(s => s.kelas_id + '|' + s.kelas_nama)).map(x => x.split('|'));
  // jumlah & poin atas seluruh hasil saringan (header dari server), bukan hanya halaman ini
  const pel = Number(pg.headers.get('X-Poin-Pelanggaran') || 0), pre = Number(pg.headers.get('X-Poin-Prestasi') || 0);
  const sis = p.siswa ? ref.siswa.find(s => String(s.id) === String(p.siswa)) : null;
  return head('Catatan Siswa', sis ? `Riwayat <b>${esc(sis.nama_lengkap)}</b> (${esc(sis.kelas_nama)})` : 'Prestasi, pelanggaran (poin), catatan wali kelas, dan catatan BK.',
    ref.siswa.length ? btn('Tambah catatan', 'catForm', p.siswa ? { siswa: p.siswa } : {}, '', 'plus') : '') +
    `<div class="stats">${stat('check', angka(pg.total), 'Catatan')}${stat('chart', pre, 'Poin prestasi')}${stat('x', pel, 'Poin pelanggaran')}</div>
    <div class="bar no-print">
      ${sis ? btn('Semua siswa', 'go', { page: 'catatan', params: '{}' }, 'ghost sm', 'back') : `<label class="sel">Kelas <select data-change="catKelas"><option value="">Semua</option>${opts(kelas, p.kelas || '')}</select></label>`}
      <label class="sel">Jenis <select data-change="catJenis"><option value="">Semua</option>${opts(Object.entries(JENIS_CAT).map(([k, [l]]) => [k, l]), p.jenis || '')}</select></label>
      ${monthInput(p.bulan || '', 'catBulan')}${p.bulan ? btn('Semua bulan', 'catBulanAll', {}, 'ghost sm') : ''}
      <small class="muted">${pagerInfo(pg.total, rows.length, 'catatan')}</small>
      <span style="flex:1"></span>${rows.length ? btn('Cetak', 'catPrint', {}, 'ghost sm', 'print') + btn('CSV', 'catCSV', {}, 'ghost sm', 'download') : ''}</div>` +
    (rows.length ? `<div class="tw"><table><thead><tr><th>Tanggal</th>${sis ? '' : '<th>Siswa</th>'}<th>Jenis</th><th>Keterangan</th><th class="n">Poin</th><th>Penulis</th><th></th></tr></thead><tbody>
    ${rows.map(r => `<tr><td>${fmtDate(r.tanggal)}</td>${sis ? '' : `<td><a href="#" data-act="go" data-page="catatan" data-params='${esc(JSON.stringify({ siswa: r.siswa_id }))}'><b>${esc(r.siswa_nama)}</b></a><small class="muted">${esc(r.kelas_nama || '')}</small></td>`}
      <td>${tag(JENIS_CAT[r.jenis][0], JENIS_CAT[r.jenis][1])}${r.kategori ? `<small class="muted">${esc(r.kategori)}</small>` : ''}</td>
      <td>${esc(r.keterangan)}${r.tindak_lanjut ? `<small>Tindak lanjut: ${esc(r.tindak_lanjut)}</small>` : ''}</td>
      <td class="n">${r.poin ? (r.jenis === 'pelanggaran' ? '−' : '+') + r.poin : '–'}</td><td><small>${esc(r.penulis || '–')}</small></td>
      <td><div class="acts">${S.catRef.staf || r.created_by === S.user.id ? iconBtn('edit', 'catForm', { id: r.id }, 'Ubah') + iconBtn('trash', 'catDel', { id: r.id }, 'Hapus', 'danger') : ''}</div></td></tr>`).join('')}
    </tbody></table></div>${pagerNav(pg.total)}` : `<div class="card">${empty(ref.siswa.length ? 'Belum ada catatan.' : 'Anda belum mengajar atau menjadi wali kelas mana pun.', 'check')}</div>`);
}
CHG.catKelas = el => go('catatan', { ...S.params, kelas: el.value });
CHG.catJenis = el => go('catatan', { ...S.params, jenis: el.value });
CHG.catBulan = el => go('catatan', { ...S.params, bulan: el.value });
ACT.catBulanAll = () => go('catatan', { ...S.params, bulan: '' });
ACT.catDel = el => del('/catatan/' + el.dataset.id, 'Hapus catatan ini?');
ACT.catForm = async el => {
  const ref = S.catRef || await GET('/catatan/siswa');
  const r = el.dataset.id ? S.catatan[el.dataset.id] : { tanggal: today(), jenis: 'pelanggaran', siswa_id: el.dataset.siswa ? Number(el.dataset.siswa) : '' };
  const boleBK = ref.staf || ref.siswa.some(s => s.wali);
  const jenis = Object.entries(JENIS_CAT).filter(([k]) => k !== 'bk' || boleBK).map(([k, [l]]) => [k, l]);
  const fields = [[{ n: 'siswa_id', l: 'Siswa', t: 'select', req: true, o: ref.siswa.map(s => [s.id, `${s.nama_lengkap} · ${s.kelas_nama}`]), blank: '— Pilih siswa —' },
    { n: 'tanggal', l: 'Tanggal', t: 'date', req: true }],
    [{ n: 'jenis', l: 'Jenis', t: 'select', o: jenis }, { n: 'kategori', l: 'Kategori', t: 'combo', o: KATEGORI_CAT[r.jenis] || [] }],
    { n: 'poin', l: 'Poin (prestasi/pelanggaran)', t: 'number', min: 0, max: 1000, hint: 'Poin dijumlahkan di halaman Wali Kelas. Kosongkan untuk catatan biasa.' },
    { n: 'keterangan', l: 'Keterangan', t: 'textarea', rows: 3, req: true },
    { n: 'tindak_lanjut', l: 'Tindak lanjut', t: 'textarea', rows: 2, ph: 'Mis. pembinaan, panggilan orang tua, pemberian penghargaan' }];
  openForm({
    title: r.id ? 'Ubah catatan' : 'Tambah catatan siswa', values: r, wide: true, fields,
    onSubmit: v => { const b = { ...v, siswa_id: Number(v.siswa_id), poin: v.poin || 0 }; return r.id ? PUT('/catatan/' + r.id, b) : POST('/catatan', b); },
  });
  // kategori menyesuaikan jenis yang dipilih
  const f = $$('.modal form').pop();
  f.jenis.addEventListener('change', () => { setComboOptions(f.kategori, KATEGORI_CAT[f.jenis.value] || []); });
};
// cetak & CSV: seluruh catatan hasil saringan (semua halaman), urut tanggal
const catRows = async () => (await semuaHalaman(S.catQ)).sort((a, b) => a.tanggal.localeCompare(b.tanggal));
ACT.catCSV = async () => downloadText('catatan-siswa.csv', toCSV([['Tanggal', 'Siswa', 'NIS', 'Kelas', 'Jenis', 'Kategori', 'Poin', 'Keterangan', 'Tindak lanjut', 'Penulis'],
  ...(await catRows()).map(r => [r.tanggal, r.siswa_nama, r.nis, r.kelas_nama, JENIS_CAT[r.jenis][0], r.kategori, r.poin, r.keterangan, r.tindak_lanjut, r.penulis])]));
ACT.catPrint = async () => { const semua = await catRows(); printDoc(`<h2 style="text-align:center;margin:0">CATATAN SISWA</h2><p style="text-align:center;margin:4px 0 14px">${esc(S.settings.nama_sekolah || '')}</p>
  <table><thead><tr><th>No</th><th>Tanggal</th><th>Siswa</th><th>Kelas</th><th>Jenis</th><th>Keterangan</th><th>Poin</th><th>Tindak lanjut</th></tr></thead><tbody>
  ${semua.map((r, i) => `<tr><td>${i + 1}</td><td>${fmtDate(r.tanggal)}</td><td>${esc(r.siswa_nama)}</td><td>${esc(r.kelas_nama || '')}</td><td>${esc(JENIS_CAT[r.jenis][0])}${r.kategori ? ' – ' + esc(r.kategori) : ''}</td>
    <td>${esc(r.keterangan)}</td><td>${r.poin || ''}</td><td>${esc(r.tindak_lanjut || '')}</td></tr>`).join('')}</tbody></table>${ttdBlock(S.user.nama_lengkap, S.user.nip || '', S.user.role === 'admin' ? 'Petugas' : 'Guru')}`); };

// ---------------------------------------------------------------------
// Wali kelas
// ---------------------------------------------------------------------
async function pageWaliKelas() {
  const ks = await GET('/walikelas');
  if (!ks.length) {
    return head('Wali Kelas') + `<div class="card">${empty(S.user.role === 'admin' ? 'Belum ada kelas di tahun ajaran aktif.' : 'Anda belum ditetapkan sebagai wali kelas. Admin mengaturnya di menu Kelas (Ubah kelas → Wali kelas).', 'users')}</div>`;
  }
  const kid = S.params.kelas && ks.some(k => String(k.id) === String(S.params.kelas)) ? S.params.kelas : ks[0].id;
  const bulan = validBulan(S.params.bulan) || monthNow();
  const d = await GET(`/walikelas/${encodeURIComponent(kid)}?bulan=${bulan}`);
  S.wali = d;
  const k = d.kelas, rows = d.siswa;
  const sum = f => rows.reduce((s, r) => s + (r[f] || 0), 0);
  return head('Wali Kelas', `Kelas <b>${esc(k.nama)}</b> · wali: ${esc(k.wali_nama || 'belum ditetapkan')} · ${rows.length} siswa`) +
    `<div class="bar no-print">${ks.length > 1 ? `<label class="sel">Kelas <select data-change="waliKelas">${opts(ks.map(x => [x.id, x.nama + (x.wali_nama ? ' · ' + x.wali_nama : '')]), kid)}</select></label>` : ''}
      ${monthInput(bulan, 'waliBulan')}<span style="flex:1"></span>
      ${btn('Catatan kelas', 'go', { page: 'catatan', params: JSON.stringify({ kelas: kid }) }, 'ghost sm', 'check')}${btn('Cetak rekap', 'waliPrint', {}, 'ghost sm', 'print')}</div>
    <div class="stats">${stat('check', sum('hadir'), 'Kehadiran ' + monthLabel(bulan))}${stat('clock', sum('terlambat'), 'Terlambat')}${stat('x', sum('alfa'), 'Alfa')}${stat('chart', sum('poin_pelanggaran'), 'Poin pelanggaran')}</div>` +
    (rows.length ? `<div class="tw no-stack"><table><thead><tr><th>No</th><th class="fz">Siswa</th><th>L/P</th><th class="n">Hadir</th><th class="n">Telat</th><th class="n">S</th><th class="n">I</th><th class="n">A</th><th class="n">Poin −</th><th class="n">Poin +</th><th>HP orang tua</th><th></th></tr></thead><tbody>
    ${rows.map((r, i) => `<tr><td>${i + 1}</td><td class="fz"><b>${esc(r.nama_lengkap)}</b><small class="muted">${esc(r.username)}${r.nisn ? ' · ' + esc(r.nisn) : ''}</small></td><td>${esc(r.jk || '–')}</td>
      <td class="n">${r.hadir}</td><td class="n">${r.terlambat || '–'}</td><td class="n">${r.sakit || '–'}</td><td class="n">${r.izin || '–'}</td><td class="n">${r.alfa ? tag(String(r.alfa), 'bad') : '–'}</td>
      <td class="n">${r.poin_pelanggaran ? tag(String(r.poin_pelanggaran), r.poin_pelanggaran >= 50 ? 'bad' : 'warn') : '–'}</td><td class="n">${r.poin_prestasi ? tag(String(r.poin_prestasi), 'ok') : '–'}</td>
      <td>${r.hp_ortu ? `<a href="${esc(waLink(r.hp_ortu))}" target="_blank" rel="noopener">${esc(r.hp_ortu)}</a>` : '<small class="muted">–</small>'}</td>
      <td><div class="acts">${iconBtn('plus', 'catForm', { siswa: r.id }, 'Tambah catatan')}${iconBtn('eye', 'go', { page: 'catatan', params: JSON.stringify({ siswa: r.id }) }, 'Riwayat catatan')}</div></td></tr>`).join('')}
    </tbody></table></div>` : `<div class="card">${empty('Belum ada siswa di kelas ini.', 'users')}</div>`) +
    (modOn('absensi') ? '' : '<p class="muted" style="margin-top:10px">Modul Absensi Harian nonaktif, jadi kolom kehadiran kosong.</p>');
}
CHG.waliKelas = el => go('walikelas', { ...S.params, kelas: el.value });
CHG.waliBulan = el => go('walikelas', { ...S.params, bulan: el.value });
ACT.waliPrint = () => {
  const { kelas: k, siswa: rows, bulan } = S.wali;
  printDoc(`<h2 style="text-align:center;margin:0">REKAP WALI KELAS</h2><p style="text-align:center;margin:4px 0 14px">${esc(S.settings.nama_sekolah || '')} · Kelas ${esc(k.nama)} · ${esc(monthLabel(bulan))}</p>
    <table><thead><tr><th>No</th><th>Nama</th><th>NIS</th><th>L/P</th><th>Hadir</th><th>Terlambat</th><th>Sakit</th><th>Izin</th><th>Alfa</th><th>Poin pelanggaran</th><th>Poin prestasi</th></tr></thead><tbody>
    ${rows.map((r, i) => `<tr><td>${i + 1}</td><td>${esc(r.nama_lengkap)}</td><td>${esc(r.username)}</td><td>${esc(r.jk || '')}</td><td>${r.hadir}</td><td>${r.terlambat}</td><td>${r.sakit}</td><td>${r.izin}</td><td>${r.alfa}</td><td>${r.poin_pelanggaran}</td><td>${r.poin_prestasi}</td></tr>`).join('')}
    </tbody></table>${ttdBlock(k.wali_nama, k.wali_nip || '', 'Wali Kelas ' + k.nama)}`);
};

// ---------------------------------------------------------------------
// Siswa: prestasi & poin miliknya
// ---------------------------------------------------------------------
async function pageCatatanku() {
  const rows = await GET('/catatan/me');
  const pel = rows.filter(r => r.jenis === 'pelanggaran').reduce((s, r) => s + r.poin, 0);
  const pre = rows.filter(r => r.jenis === 'prestasi').reduce((s, r) => s + r.poin, 0);
  return head('Prestasi & Poin', 'Catatan prestasi dan pelanggaran dari guru dan wali kelas.') +
    `<div class="stats">${stat('chart', pre, 'Poin prestasi')}${stat('x', pel, 'Poin pelanggaran')}</div>` +
    (rows.length ? `<div class="tw"><table><thead><tr><th>Tanggal</th><th>Jenis</th><th>Keterangan</th><th class="n">Poin</th></tr></thead><tbody>
    ${rows.map(r => `<tr><td>${fmtDate(r.tanggal)}</td><td>${tag(JENIS_CAT[r.jenis][0], JENIS_CAT[r.jenis][1])}${r.kategori ? `<small class="muted">${esc(r.kategori)}</small>` : ''}</td><td>${esc(r.keterangan)}</td>
      <td class="n">${r.poin ? (r.jenis === 'pelanggaran' ? '−' : '+') + r.poin : '–'}</td></tr>`).join('')}</tbody></table></div>`
      : `<div class="card">${empty('Belum ada catatan. Pertahankan!', 'check')}</div>`);
}

Object.assign(PAGES.admin, { jurnal: pageJurnal, catatan: pageCatatan, walikelas: pageWaliKelas });
Object.assign(PAGES.guru, { jurnal: pageJurnal, catatan: pageCatatan, walikelas: pageWaliKelas });
Object.assign(PAGES.siswa, { catatanku: pageCatatanku });
