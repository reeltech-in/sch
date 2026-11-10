'use strict';
/* PPDB online: formulir publik, cek status, verifikasi pendaftar, jadikan siswa */

addNav('admin', ['ppdb', 'PPDB Online', 'users', 'ppdb', 'Layanan']);

// =====================================================================
// PPDB
// =====================================================================
const PPDB_ST = [['baru', 'Baru'], ['verifikasi', 'Diverifikasi'], ['diterima', 'Diterima'], ['cadangan', 'Cadangan'], ['ditolak', 'Ditolak']];
const ppdbTag = s => tag(PPDB_ST.find(x => x[0] === s)?.[1] || s, s === 'diterima' ? 'ok' : s === 'ditolak' ? 'bad' : s === 'cadangan' ? 'warn' : s === 'verifikasi' ? 'info' : '');
async function pagePPDB() {
  const tab = S.params.tab || 'daftar';
  const hd = head('PPDB Online', 'Penerimaan peserta didik baru: formulir publik, verifikasi berkas, lalu jadikan siswa sekali klik.');
  const t = tabs([['daftar', 'Pendaftar'], ['atur', 'Pengaturan']], tab);
  if (tab === 'atur') {
    const [c, tas] = await Promise.all([GET('/config/ppdb'), tahunAjaranOpts()]);
    const fields = [[{ n: 'buka', l: 'Dibuka', t: 'datetime', req: true }, { n: 'tutup', l: 'Ditutup', t: 'datetime' }], [{ n: 'tahun_ajaran', l: 'Tahun ajaran', t: 'select', o: tas.map(x => [x, x]), blank: '—' }, { n: 'kuota', l: 'Kuota', t: 'number', min: 0 }],
      { n: 'jalur', l: 'Jalur pendaftaran', hint: 'Pisahkan dengan koma.' }, { n: 'info', l: 'Informasi untuk pendaftar', t: 'textarea', rows: 3 }];
    const id = registerForm(fields, async v => { await PUT('/config/ppdb', v); toast('Pengaturan PPDB tersimpan.'); });
    return hd + t + `<div class="cols"><section class="card">${formHtml(id, fields, c, { cancel: false })}</section>${linkCard(location.origin + '/#/ppdb', 'Alamat formulir pendaftaran')}</div>`;
  }
  const f = S.params.status || '';
  const pg = await GETpage('/ppdb?' + qs({ status: f, q: S.params.q || '', ...pageQ(50) }));
  const rows = pg.rows;
  if (pageOverflow(pg.total, rows.length)) return '';
  S.ppdbQ = '/ppdb?' + qs({ status: f, q: S.params.q || '' });
  S.ppdb = Object.fromEntries(rows.map(r => [r.id, r]));
  S.ppdbRows = rows;
  return hd + t + `<div class="bar"><input class="search" type="search" placeholder="Cari nama, nomor, NISN, asal sekolah… (Enter)" value="${esc(S.params.q || '')}" data-srv="ppdb" aria-label="Cari"><small class="muted">${pagerInfo(pg.total, rows.length, 'pendaftar')}</small>
    <select data-change="ppdbF" aria-label="Status"><option value="">Semua status</option>${opts(PPDB_ST, f)}</select>${btn('Unduh CSV', 'ppdbCSV', {}, 'ghost sm', 'download')}</div>` +
    (rows.length ? `<div class="tw"><table id="tblPP"><thead><tr><th>Pendaftar</th><th>Asal sekolah</th><th>Jalur</th><th>Berkas</th><th>Daftar</th><th>Status</th><th></th></tr></thead><tbody>
    ${rows.map(r => `<tr><td><b>${esc(r.nama)}</b><small class="muted"><code>${esc(r.no_daftar)}</code> · ${esc(r.hp || '')}</small></td><td>${esc(r.asal_sekolah || '–')}</td><td>${esc(r.jalur || '–')}</td>
      <td>${Object.keys(r.berkas || {}).length ? tag(Object.keys(r.berkas).length + ' file', 'info') : tag('Kosong', 'warn')}</td><td>${fmtDate(r.created_at)}</td><td>${ppdbTag(r.status)}${r.user_id ? ' ' + tag('Akun dibuat', 'ok') : ''}</td>
      <td><div class="acts">${btn('Periksa', 'ppdbOpen', { id: r.id }, 'sm')}${iconBtn('trash', 'ppdbDel', { id: r.id, nama: r.nama }, 'Hapus', 'danger')}</div></td></tr>`).join('')}
    </tbody></table></div>${pagerNav(pg.total)}` : `<div class="card">${empty('Belum ada pendaftar.', 'users')}</div>`);
}
CHG.ppdbF = el => go('ppdb', { status: el.value, q: S.params.q || '', per: S.params.per || '' });
ACT.ppdbDel = el => del('/ppdb/' + el.dataset.id, `Hapus data pendaftar <b>${esc(el.dataset.nama)}</b>?`);
ACT.ppdbCSV = async () => downloadText('pendaftar-ppdb.csv', toCSV([['No daftar', 'Nama', 'NISN', 'JK', 'Tempat lahir', 'Tgl lahir', 'Asal sekolah', 'Ayah', 'Ibu', 'HP', 'Jalur', 'Status'],
  ...(await semuaHalaman(S.ppdbQ)).map(r => [r.no_daftar, r.nama, r.nisn, r.jk, r.tempat_lahir, r.tgl_lahir, r.asal_sekolah, r.nama_ayah, r.nama_ibu, r.hp, r.jalur, r.status])]));
const BERKAS = { kk: 'Kartu Keluarga', akta: 'Akta Kelahiran', foto: 'Pas foto', lain: 'Lainnya (rapor/sertifikat)' };
ACT.ppdbOpen = async el => {
  const r = S.ppdb[el.dataset.id];
  const ks = await GET('/classes');
  const fields = [['nisn', 'NISN'], ['nik', 'NIK'], ['jk', 'L/P'], ['tempat_lahir', 'Tempat lahir'], ['tgl_lahir', 'Tanggal lahir'], ['agama', 'Agama'], ['alamat', 'Alamat'],
    ['asal_sekolah', 'Asal sekolah'], ['nama_ayah', 'Ayah'], ['nama_ibu', 'Ibu'], ['hp', 'HP'], ['email', 'Email'], ['jalur', 'Jalur']];
  const id = registerForm([{ n: 'status' }, { n: 'catatan' }], async v => { await PUT('/ppdb/' + r.id, v); closeModal(); toast('Status disimpan.'); await afterSave(); });
  const wa = waLink(r.hp, `Yth. Orang tua/wali ${r.nama}, status pendaftaran ${r.no_daftar} di ${S.settings.nama_sekolah}: `);
  openModal({ title: r.nama + ' · ' + r.no_daftar, wide: true, body: `<div class="cols"><div><div class="tw"><table><tbody>${fields.map(([k, l]) => `<tr><th style="width:38%">${l}</th><td>${esc(k === 'tgl_lahir' && r[k] ? fmtDate(r[k]) : r[k] || '–')}</td></tr>`).join('')}</tbody></table></div></div>
    <div><h3 style="margin-bottom:8px">Berkas</h3>${Object.keys(BERKAS).filter(k => r.berkas?.[k]).map(k => `<a class="lamp" href="${esc(r.berkas[k])}" target="_blank" rel="noopener" style="margin-bottom:8px">${ico(r.berkas[k].endsWith('.pdf') ? 'file' : 'image')}<span>${BERKAS[k]}</span></a>`).join('') || '<p class="muted">Tidak ada berkas.</p>'}
      <form class="form" data-form="${id}" style="margin-top:14px"><label class="f"><span>Status</span><select name="status">${opts(PPDB_ST, r.status)}</select></label>
        <label class="f"><span>Catatan untuk pendaftar</span><textarea name="catatan" rows="2">${esc(r.catatan || '')}</textarea></label><div class="form-err" role="alert"></div>
        <div class="sheet-foot">${wa ? `<a class="btn ghost" href="${esc(wa)}" target="_blank" rel="noopener">WhatsApp</a>` : ''}<button class="btn" type="submit">Simpan status</button></div></form>
      ${r.status === 'diterima' && !r.user_id ? `<div class="card" style="margin-top:14px;background:var(--surface-2);box-shadow:none"><h3 style="margin-bottom:8px">Jadikan siswa</h3>
        <div style="display:flex;gap:8px"><select id="ppdbKelas" style="flex:1"><option value="">— Kelas (bisa diatur nanti) —</option>${opts(ks.map(k => [k.id, k.nama]), '')}</select>${btn('Buat akun', 'ppdbSiswa', { id: r.id }, 'sm')}</div>
        <small class="muted">Membuat akun siswa (username = NISN) dan biodata Buku Induk.</small></div>` : ''}</div></div>` });
};
ACT.ppdbSiswa = async el => {
  const r = await POST(`/ppdb/${el.dataset.id}/siswa`, { kelas_id: Number($('#ppdbKelas').value) || 0 });
  closeModal(); await afterSave();
  openModal({ title: 'Akun siswa dibuat', body: `<p>Berikan data login ini kepada siswa:</p><div class="tw" style="margin:12px 0"><table><tbody><tr><th>Username</th><td><code>${esc(r.username)}</code></td></tr><tr><th>Password</th><td><code>${esc(r.password)}</code></td></tr></tbody></table></div>
    <p class="muted" style="font-size:13px">Password hanya tampil sekali. Siswa dapat menggantinya di menu Akun Saya.</p><div class="sheet-foot"><button class="btn" data-act="closeModal">Selesai</button></div>` });
};

PUBLIC_PAGES.ppdb = async () => {
  const info = await GET('/ppdb/info').catch(e => ({ error: e.message }));
  if (S.user || parseHash().page !== 'ppdb') return; // sudah pindah halaman selama menunggu (WiFi lambat)
  const jalur = String(info.jalur || '').split(',').map(s => s.trim()).filter(Boolean);
  const fid = registerForm([], async (v, form) => {
    const d = Object.fromEntries([...new FormData(form)].filter(([k]) => !k.startsWith('f_')).map(([k, x]) => [k, String(x).trim()]));
    const berkas = {};
    for (const k of Object.keys(BERKAS)) { const f = form['f_' + k]?.files[0]; if (f) berkas[k] = await fileToDataURL(f); }
    const website = d.website; delete d.website;
    const r = await POST('/ppdb/daftar', { data: d, berkas, website });
    $('#ppdbBox').innerHTML = `<div class="grad-result"><div class="stat" style="justify-content:center;border:0;box-shadow:none;background:none"><div class="ic">${ico('check')}</div></div>
      <h2>Pendaftaran berhasil</h2><p class="muted" style="margin:6px 0 14px">Simpan atau foto data berikut untuk mengecek status pendaftaran.</p>
      <div class="tw" style="text-align:left"><table><tbody><tr><th>Nama</th><td><b>${esc(r.nama)}</b></td></tr><tr><th>No. pendaftaran</th><td><b style="font-size:18px">${esc(r.no_daftar)}</b></td></tr>
      <tr><th>PIN</th><td><b style="font-size:18px">${esc(r.pin)}</b></td></tr></tbody></table></div>
      <div style="margin-top:14px">${btn('Cetak bukti', 'print', {}, 'ghost', 'print')}</div></div>`;
    window.scrollTo(0, 0);
  });
  const cid = registerForm([{ n: 'no_daftar' }, { n: 'pin' }], async v => {
    const r = await POST('/ppdb/cek', v);
    $('#ppdbCek').innerHTML = `<div class="scan-result ${r.status === 'diterima' ? 'ok' : r.status === 'ditolak' ? 'bad' : 'warn'}" style="margin-top:12px"><div><b>${esc(r.nama)}</b>Status: ${ppdbTag(r.status)}${r.catatan ? `<br>${esc(r.catatan)}` : ''}</div></div>`;
  });
  const fld = (n, l, extra = '') => `<label class="f"><span>${l}</span><input name="${n}" ${extra}></label>`;
  $('#app').innerHTML = publicShell('Penerimaan Peserta Didik Baru', `<h1 style="font-size:24px;margin-bottom:4px">PPDB ${esc(info.tahun_ajaran || '')}</h1>
    <p class="muted" style="margin-bottom:14px">${esc(info.nama_sekolah || '')}${info.kuota ? ` · kuota ${info.kuota}` : ''} · ${info.pendaftar || 0} pendaftar</p>
    ${info.info ? `<div class="hint" style="margin-bottom:16px">${esc(info.info)}</div>` : ''}
    <div id="ppdbBox">${info.error ? errBox({ message: info.error }) : !info.buka ? `<p><b>${esc(info.pesan || 'Pendaftaran belum dibuka.')}</b></p>` : `<form class="form" data-form="${fid}">
      ${fld('nama', 'Nama lengkap calon siswa', 'required')}
      <div class="row2">${fld('nisn', 'NISN', 'inputmode="numeric"')}${fld('nik', 'NIK', 'inputmode="numeric"')}</div>
      <div class="row2"><label class="f"><span>Jenis kelamin</span><select name="jk"><option value="">—</option><option value="L">Laki-laki</option><option value="P">Perempuan</option></select></label>
        <label class="f"><span>Agama</span><select name="agama"><option value="">—</option>${['Islam', 'Kristen', 'Katolik', 'Hindu', 'Buddha', 'Konghucu'].map(a => `<option>${a}</option>`).join('')}</select></label></div>
      <div class="row2">${fld('tempat_lahir', 'Tempat lahir')}${fld('tgl_lahir', 'Tanggal lahir', 'type="date" required')}</div>
      <label class="f"><span>Alamat</span><textarea name="alamat" rows="2"></textarea></label>
      ${fld('asal_sekolah', 'Asal sekolah')}
      <div class="row2">${fld('nama_ayah', 'Nama ayah')}${fld('nama_ibu', 'Nama ibu')}</div>
      <div class="row2">${fld('hp', 'No. HP/WhatsApp orang tua', 'required inputmode="tel"')}${fld('email', 'Email', 'type="email"')}</div>
      ${jalur.length ? `<label class="f"><span>Jalur pendaftaran</span><select name="jalur">${jalur.map(j => `<option>${esc(j)}</option>`).join('')}</select></label>` : ''}
      <h3 style="margin-top:6px">Unggah berkas <small class="muted">(JPG/PNG/PDF, foto otomatis diperkecil)</small></h3>
      ${Object.entries(BERKAS).map(([k, l]) => `<label class="f"><span>${l}</span><input type="file" name="f_${k}" accept="image/*,application/pdf"></label>`).join('')}
      <input name="website" tabindex="-1" autocomplete="off" style="position:absolute;left:-9999px" aria-hidden="true">
      <div class="form-err" role="alert"></div><button class="btn block" type="submit">Kirim pendaftaran</button></form>`}</div>
    <hr style="border:0;border-top:1px solid var(--border);margin:22px 0">
    <h3 style="margin-bottom:10px">Cek status pendaftaran</h3>
    <form class="form" data-form="${cid}"><div class="row2">${fld('no_daftar', 'No. pendaftaran', 'required placeholder="PPDB-2026-0001"')}${fld('pin', 'PIN', 'required inputmode="numeric"')}</div>
      <div class="form-err" role="alert"></div><button class="btn ghost" type="submit">Cek status</button></form><div id="ppdbCek"></div>`);
  document.title = 'PPDB · ' + (info.nama_sekolah || '');
};

Object.assign(PAGES.admin, { ppdb: pagePPDB });
