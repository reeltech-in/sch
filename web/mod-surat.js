'use strict';
/* Surat masuk/keluar: arsip, nomor otomatis, disposisi */

addNav('admin', ['surat', 'Surat Masuk/Keluar', 'inbox', 'surat', 'Layanan']);

// =====================================================================
// SURAT MASUK / KELUAR
// =====================================================================
const SURAT_ST = [['baru', 'Baru'], ['didisposisi', 'Didisposisi'], ['selesai', 'Selesai']];
async function pageSurat() {
  const tab = ['masuk', 'keluar', 'atur'].includes(S.params.tab) ? S.params.tab : 'masuk';
  const tahun = /^\d{4}$/.test(S.params.tahun || '') ? S.params.tahun : String(new Date().getFullYear());
  const hd = head('Surat Masuk & Keluar', 'Agenda surat, arsip scan, disposisi, dan penomoran surat keluar otomatis.', tab !== 'atur' ? btn('Catat surat ' + tab, 'suratForm', { jenis: tab }, '', 'plus') : '');
  const t = tabs([['masuk', 'Surat masuk'], ['keluar', 'Surat keluar'], ['atur', 'Format nomor']], tab);
  if (tab === 'atur') {
    const c = await GET('/config/surat');
    const fields = [{ n: 'singkatan', l: 'Singkatan sekolah', req: true }, { n: 'format', l: 'Format nomor surat keluar', req: true, hint: 'Penanda: {no} {kode} {singkatan} {bulan} (romawi) {tahun}' }];
    const id = registerForm(fields, async v => { await PUT('/config/surat', v); toast('Format nomor tersimpan.'); });
    return hd + t + `<section class="card" style="max-width:560px">${formHtml(id, fields, c, { cancel: false })}</section>`;
  }
  const rows = await GET(`/surat?jenis=${tab}&tahun=${tahun}`);
  S.surat = Object.fromEntries(rows.map(r => [r.id, r]));
  S.suratRows = { rows, tab, tahun };
  return hd + t + `<div class="bar"><input type="number" value="${tahun}" min="2000" max="2100" data-change="suratTahun" aria-label="Tahun" style="width:110px;min-width:0">
    <input class="search" type="search" placeholder="Cari nomor, ${tab === 'masuk' ? 'pengirim' : 'tujuan'}, perihal…" data-filter="tblS2" aria-label="Cari">${btn('Cetak agenda', 'suratAgenda', {}, 'ghost sm', 'print')}</div>` +
    (rows.length ? `<div class="tw"><table id="tblS2"><thead><tr><th>Tanggal</th><th>Nomor</th><th>${tab === 'masuk' ? 'Pengirim' : 'Tujuan'}</th><th>Perihal</th><th>Status</th><th></th></tr></thead><tbody>
    ${rows.map(r => `<tr><td>${fmtDate(r.tanggal)}</td><td><code>${esc(r.nomor || '–')}</code></td><td>${esc(r.pihak)}</td><td><b>${esc(r.perihal)}</b>${r.disposisi_ke ? `<small class="muted">Disposisi: ${esc(r.disposisi_ke)}</small>` : ''}</td>
      <td>${tag(SURAT_ST.find(s => s[0] === r.status)?.[1] || r.status, r.status === 'selesai' ? 'ok' : r.status === 'didisposisi' ? 'info' : 'warn')}</td>
      <td><div class="acts">${r.file ? iconBtn('file', 'openUrl', { url: r.file }, 'Lihat file') : ''}${tab === 'masuk' ? iconBtn('print', 'suratDispo', { id: r.id }, 'Cetak lembar disposisi') : ''}${iconBtn('edit', 'suratForm', { id: r.id, jenis: tab }, 'Ubah')}${iconBtn('trash', 'suratDel', { id: r.id }, 'Hapus', 'danger')}</div></td></tr>`).join('')}
    </tbody></table></div>` : `<div class="card">${empty('Belum ada surat ' + tab + ' tahun ini.', 'inbox')}</div>`);
}
// kode klasifikasi arsip yang umum dipakai sekolah (boleh mengetik kode lain)
const KODE_SURAT = [['005', 'Undangan'], ['045', 'Kearsipan'], ['090', 'Perjalanan dinas'], ['400', 'Kesejahteraan rakyat'], ['420', 'Pendidikan'], ['421', 'Sekolah'],
  ['421.2', 'Sekolah menengah pertama'], ['421.3', 'Sekolah menengah atas'], ['422', 'Administrasi sekolah'], ['423', 'Metode belajar'], ['424', 'Tenaga pengajar'],
  ['425', 'Sarana pendidikan'], ['426', 'Keolahragaan'], ['800', 'Kepegawaian'], ['850', 'Cuti'], ['900', 'Keuangan']];
CHG.suratTahun = el => go('surat', { tab: S.params.tab || 'masuk', tahun: el.value });
ACT.suratDel = el => del('/surat/' + el.dataset.id, 'Hapus catatan surat ini?');
ACT.suratForm = async el => {
  const disp = el.dataset.jenis === 'masuk' ? await saran('surat.disposisi', ['Kepala Sekolah', 'Wakasek Kurikulum', 'Wakasek Kesiswaan', 'Wakasek Sarpras', 'Wakasek Humas', 'Tata Usaha', 'Bendahara', 'Guru BK', 'Wali Kelas']) : [];
  const jenis = el.dataset.jenis;
  const r = el.dataset.id ? S.surat[el.dataset.id] : { jenis, tanggal: today(), status: jenis === 'keluar' ? 'selesai' : 'baru' };
  const fields = [[{ n: 'tanggal', l: 'Tanggal surat', t: 'date', req: true }, { n: 'kode', l: 'Kode klasifikasi', t: 'combo', o: KODE_SURAT }],
    { t: 'html', html: `<label class="f"><span>Nomor surat</span><div style="display:flex;gap:8px"><input name="nomor" value="${esc(r.nomor || '')}" style="flex:1">${jenis === 'keluar' ? btn('Nomor otomatis', 'suratNomor', {}, 'ghost sm') : ''}</div></label>` },
    { n: 'pihak', l: jenis === 'masuk' ? 'Pengirim' : 'Tujuan', req: true }, { n: 'perihal', l: 'Perihal', req: true }, { n: 'ringkasan', l: 'Ringkasan isi', t: 'textarea', rows: 2 },
    { n: 'file', l: 'Scan / file surat', t: 'file', accept: 'image/*,application/pdf' }];
  if (jenis === 'masuk') fields.push([{ n: 'disposisi_ke', l: 'Disposisi kepada', t: 'combo', o: disp }, { n: 'status', l: 'Status', t: 'select', o: SURAT_ST }], { n: 'disposisi_isi', l: 'Isi disposisi', t: 'textarea', rows: 2 });
  openForm({
    title: (r.id ? 'Ubah' : 'Catat') + ' surat ' + jenis, values: r, wide: true, fields,
    onSubmit: (v, form) => { const body = { ...v, jenis, nomor: form.nomor.value.trim(), status: v.status || r.status }; return r.id ? PUT('/surat/' + r.id, body) : POST('/surat', body); },
  });
};
ACT.suratNomor = async el => { const f = el.closest('form'); const r = await GET('/surat/nomor?kode=' + encodeURIComponent(f.kode.value || '')); f.nomor.value = r.nomor; };
ACT.suratAgenda = () => {
  const { rows, tab, tahun } = S.suratRows;
  printDoc(schoolHead() + `<h3 style="text-align:center">AGENDA SURAT ${tab.toUpperCase()} TAHUN ${esc(tahun)}</h3><table><thead><tr><th>No</th><th>Tanggal</th><th>Nomor surat</th><th>${tab === 'masuk' ? 'Pengirim' : 'Tujuan'}</th><th>Perihal</th><th>Keterangan</th></tr></thead><tbody>
    ${rows.slice().reverse().map((r, i) => `<tr><td>${i + 1}</td><td>${fmtDate(r.tanggal)}</td><td>${esc(r.nomor || '')}</td><td>${esc(r.pihak)}</td><td>${esc(r.perihal)}</td><td>${esc(r.disposisi_ke ? 'Disposisi: ' + r.disposisi_ke : '')}</td></tr>`).join('')}</tbody></table>`);
};
ACT.suratDispo = el => {
  const r = S.surat[el.dataset.id], st = S.settings;
  printDoc(schoolHead() + `<h3 style="text-align:center">LEMBAR DISPOSISI</h3><table><tbody>
    <tr><td style="width:30%">Surat dari</td><td>${esc(r.pihak)}</td></tr><tr><td>Nomor surat</td><td>${esc(r.nomor || '')}</td></tr><tr><td>Tanggal surat</td><td>${fmtDate(r.tanggal)}</td></tr>
    <tr><td>Perihal</td><td>${esc(r.perihal)}</td></tr><tr><td>Diteruskan kepada</td><td style="height:40px">${esc(r.disposisi_ke || '')}</td></tr><tr><td>Isi disposisi</td><td style="height:120px;vertical-align:top">${esc(r.disposisi_isi || '')}</td></tr></tbody></table>
    <div class="sign"><div></div><div>Kepala Sekolah,<br><br><br><br><b><u>${esc(st.kepala_sekolah || '..............................')}</u></b><br>NIP. ${esc(st.nip_kepala || '')}</div></div>`);
};

Object.assign(PAGES.admin, { surat: pageSurat });
