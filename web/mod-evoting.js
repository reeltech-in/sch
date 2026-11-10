'use strict';
/* E-Voting OSIS: pemilihan, kandidat, bilik suara, hasil */

addNav('admin', ['evoting', 'E-Voting OSIS', 'vote', 'evoting', 'Layanan']);
addNav('guru', ['evoting', 'E-Voting', 'vote', 'evoting', 'Layanan']);
addNav('siswa', ['evoting', 'E-Voting OSIS', 'vote', 'evoting', 'Layanan']);
PAGE_PARENT.bilik = 'evoting';

// =====================================================================
// E-VOTING OSIS
// =====================================================================
const elStatus = e => { const n = new Date(nowMs()); return n < new Date(e.mulai) ? ['Belum dimulai', 'info'] : n > new Date(e.selesai) ? ['Selesai', ''] : ['Berlangsung', 'ok']; };
async function pageEvoting() {
  if (!isAdm('evoting')) return evoteVoter();
  const [els, ks] = await Promise.all([GET('/evote'), GET('/classes')]);
  S.els = Object.fromEntries(els.map(e => [e.id, e])); S.elKelas = ks;
  return head('E-Voting OSIS', 'Pemilihan ketua OSIS dengan suara rahasia. Siswa memilih dari HP, atau lewat bilik suara dengan scan kartu.', btn('Pemilihan baru', 'elForm', {}, '', 'plus')) +
    (els.length ? els.map(e => { const [s, k] = elStatus(e);
      return `<section class="card"><div class="card-head"><div><h2>${esc(e.judul)}</h2><small class="muted">${fmtDT(e.mulai)} – ${fmtDT(e.selesai)} · ${e.jumlah_memilih} suara masuk</small></div>
        <div class="acts">${tag(s, k)}${btn('Hasil', 'elHasil', { id: e.id }, 'ghost sm', 'chart')}${btn('Bilik suara', 'go', { page: 'bilik', params: JSON.stringify({ id: e.id }) }, 'ghost sm', 'vote')}${iconBtn('edit', 'elForm', { id: e.id }, 'Ubah')}${iconBtn('trash', 'elDel', { id: e.id }, 'Hapus', 'danger')}</div></div>
        <div class="grid">${e.kandidat.map(c => `<div class="card" style="box-shadow:none;background:var(--surface-2);display:flex;gap:12px;align-items:center">${avatar(c.foto, c.nama, 56)}<div style="flex:1"><b>${c.no}. ${esc(c.nama)}</b><small class="muted">${esc(c.wakil ? 'Wakil: ' + c.wakil : '')}</small></div>${iconBtn('edit', 'kandForm', { el: e.id, id: c.id }, 'Ubah')}${iconBtn('trash', 'kandDel', { id: c.id }, 'Hapus', 'danger')}</div>`).join('')}
          <button type="button" class="card" data-act="kandForm" data-el="${e.id}" style="box-shadow:none;border:2px dashed var(--border);background:none;cursor:pointer;color:var(--text-2);font:inherit;display:grid;place-items:center;min-height:80px">${ico('plus')} Tambah kandidat</button></div></section>`; }).join('')
      : `<div class="card">${empty('Belum ada pemilihan.', 'vote')}</div>`);
}
ACT.elForm = el => {
  const e = el.dataset.id ? { ...S.els[el.dataset.id] } : { pemilih: 'siswa', tampil_hasil: 1, kelas_ids: [] };
  e.tampil_hasil = !!e.tampil_hasil;
  openForm({
    title: e.id ? 'Ubah pemilihan' : 'Pemilihan baru', values: e, wide: true,
    fields: [{ n: 'judul', l: 'Nama pemilihan', req: true, ph: 'Pemilihan Ketua OSIS 2026/2027' }, { n: 'deskripsi', l: 'Keterangan', t: 'textarea', rows: 2 },
      [{ n: 'mulai', l: 'Mulai', t: 'datetime', req: true }, { n: 'selesai', l: 'Selesai', t: 'datetime', req: true }],
      { n: 'pemilih', l: 'Pemilih', t: 'select', o: [['siswa', 'Siswa saja'], ['semua', 'Siswa dan guru']] },
      { n: 'kelas_ids', l: 'Kelas pemilih (kosongkan = semua kelas)', t: 'checks', o: S.elKelas.map(k => [k.id, k.nama]) },
      { n: 'tampil_hasil', l: 'Tampilkan hasil ke pemilih setelah pemilihan selesai', t: 'check' }],
    onSubmit: v => { const b = { ...v, kelas_ids: v.kelas_ids.map(Number) }; return e.id ? PUT('/evote/' + e.id, b) : POST('/evote', b); },
  });
};
ACT.elDel = el => del('/evote/' + el.dataset.id, 'Hapus pemilihan ini beserta seluruh suaranya?');
ACT.kandForm = el => {
  const e = S.els[el.dataset.el];
  const c = el.dataset.id ? e.kandidat.find(x => x.id === Number(el.dataset.id)) : { no: e.kandidat.length + 1 };
  openForm({
    title: c.id ? 'Ubah kandidat' : 'Tambah kandidat', values: c, wide: true,
    fields: [[{ n: 'no', l: 'Nomor urut', t: 'number', min: 1, req: true }, { n: 'foto', l: 'Foto', t: 'file', accept: 'image/*' }], [{ n: 'nama', l: 'Calon ketua', req: true }, { n: 'wakil', l: 'Calon wakil' }],
      { n: 'visi', l: 'Visi', t: 'textarea', rows: 2 }, { n: 'misi', l: 'Misi', t: 'textarea', rows: 4 }],
    onSubmit: v => c.id ? PUT('/evote-kandidat/' + c.id, v) : POST(`/evote/${e.id}/kandidat`, v),
  });
};
ACT.kandDel = el => del('/evote-kandidat/' + el.dataset.id, 'Hapus kandidat ini?');
// rekap bilik per petugas: jumlah besar dalam waktu singkat patut diperiksa
const bilikHtml = r => r.bilik && r.bilik.length ? `<div class="card-head" style="margin-top:16px"><h3>Suara lewat bilik per petugas</h3></div>
  <div class="tw"><table><thead><tr><th>Petugas</th><th class="n">Suara</th><th>Pertama</th><th>Terakhir</th></tr></thead><tbody>
  ${r.bilik.map(b => `<tr><td>${esc(b.petugas)}</td><td class="n">${b.jumlah}</td><td>${fmtDT(b.pertama)}</td><td>${fmtDT(b.terakhir)}</td></tr>`).join('')}</tbody></table></div>
  <p class="muted" style="font-size:12px;margin-top:6px">Hanya terlihat oleh pengelola e-voting. Pilihan tiap pemilih tetap rahasia.</p>` : '';
const hasilHtml = r => bilikHtml(r) ? hasilUtama(r) + bilikHtml(r) : hasilUtama(r);
const hasilUtama = r => {
  if (r.kandidat.some(k => k.suara == null)) {
    return `<div class="stats">${stat('users', `${r.memilih}/${r.pemilih}`, 'Sudah memilih')}${stat('chart', r.pemilih ? fmtN(r.memilih / r.pemilih * 100) + '%' : '–', 'Partisipasi')}</div>
      <div class="hint">Perolehan suara dirahasiakan sampai pemungutan suara selesai, supaya pilihan tiap pemilih di bilik tidak bisa ditebak.</div>`;
  }
  const max = Math.max(1, ...r.kandidat.map(k => k.suara));
  const tot = r.kandidat.reduce((s, k) => s + k.suara, 0);
  return `<div class="stats">${stat('users', `${r.memilih}/${r.pemilih}`, 'Sudah memilih')}${stat('chart', r.pemilih ? fmtN(r.memilih / r.pemilih * 100) + '%' : '–', 'Partisipasi')}</div>
    ${r.kandidat.map(k => `<div style="margin-bottom:14px"><div style="display:flex;justify-content:space-between;gap:10px;margin-bottom:6px"><div class="cell-user">${avatar(k.foto, k.nama, 36)}<b>${k.no}. ${esc(k.nama)}${k.wakil ? ' & ' + esc(k.wakil) : ''}</b></div>
      <b>${k.suara} suara · ${tot ? fmtN(k.suara / tot * 100) : 0}%</b></div><div class="prog" style="height:14px"><i style="width:${k.suara / max * 100}%"></i></div></div>`).join('')}`;
};
ACT.elHasil = el => {
  const m = openModal({ title: 'Hasil: ' + S.els[el.dataset.id].judul, wide: true, builder: async () => hasilHtml(await GET(`/evote/${el.dataset.id}/hasil`)) + '<p class="muted" style="font-size:12px">Diperbarui otomatis tiap 5 detik.</p>' });
  const t = setInterval(() => { if (!document.body.contains(m.el)) return clearInterval(t); refreshModal(m); }, 5000);
};

async function evoteVoter() {
  const els = await GET('/evote-aktif');
  S.myEls = Object.fromEntries(els.map(e => [e.id, e]));
  if (!els.length) return head('E-Voting') + `<div class="card">${empty('Tidak ada pemilihan untuk Anda.', 'vote')}</div>`;
  // hasil semua pemilihan yang sudah boleh dilihat diambil bersamaan
  const lihatHasil = e => !(e.status === 'berlangsung' && !e.sudah) && !(e.sudah && e.status !== 'selesai') && e.status !== 'belum' && e.tampil_hasil;
  const hasil = await Promise.all(els.map(e => lihatHasil(e) ? GET(`/evote/${e.id}/hasil`).catch(() => ({ kandidat: [], memilih: 0, pemilih: 0 })) : null));
  const parts = [];
  for (const [i, e] of els.entries()) {
    let body;
    if (e.status === 'berlangsung' && !e.sudah) body = `<div class="grid">${e.kandidat.map(c => `<div class="card" style="box-shadow:none;background:var(--surface-2);text-align:center;display:flex;flex-direction:column;gap:8px;align-items:center">
        ${c.foto ? `<img src="${esc(c.foto)}" alt="" style="width:120px;height:150px;object-fit:cover;border-radius:12px">` : avatar(null, c.nama, 96)}
        <div style="font-size:28px;font-weight:800;color:var(--brand)">${c.no}</div><b>${esc(c.nama)}</b>${c.wakil ? `<small class="muted">& ${esc(c.wakil)}</small>` : ''}
        ${c.visi || c.misi ? `<details style="text-align:left;width:100%"><summary class="muted" style="cursor:pointer">Visi & misi</summary><p style="white-space:pre-wrap;margin-top:6px"><b>Visi:</b> ${esc(c.visi || '-')}<br><b>Misi:</b> ${esc(c.misi || '-')}</p></details>` : ''}
        ${btn('Pilih', 'votePick', { el: e.id, id: c.id }, 'block')}</div>`).join('')}</div>`;
    else if (e.sudah && e.status !== 'selesai') body = `<div class="scan-result ok"><div><b>Terima kasih</b>Suara Anda sudah tercatat. Pilihan Anda rahasia.</div></div>`;
    else if (e.status === 'belum') body = `<p class="muted">Pemungutan suara dibuka ${fmtDT(e.mulai)}.</p>`;
    else if (e.tampil_hasil) body = hasilHtml(hasil[i]);
    else body = '<p class="muted">Pemilihan selesai. Hasil diumumkan oleh sekolah.</p>';
    parts.push(`<section class="card"><div class="card-head"><div><h2>${esc(e.judul)}</h2><small class="muted">${fmtDT(e.mulai)} – ${fmtDT(e.selesai)}</small></div>${tag(...elStatus(e))}</div>${e.deskripsi ? `<p class="muted" style="margin-bottom:12px">${esc(e.deskripsi)}</p>` : ''}${body}</section>`);
  }
  return head('E-Voting', 'Satu orang satu suara. Pilihan Anda tidak bisa dilihat siapa pun.') + parts.join('');
}
ACT.votePick = async el => {
  const e = S.myEls[el.dataset.el], c = e.kandidat.find(x => x.id === Number(el.dataset.id));
  if (!(await confirmBox(`Pilih <b>${c.no}. ${esc(c.nama)}</b>? Suara tidak bisa diubah setelah dikirim.`, 'Ya, pilih', ''))) return;
  await POST(`/evote/${e.id}/pilih`, { candidate_id: c.id });
  toast('Suara Anda tercatat. Terima kasih!');
  refresh(true);
};

// Bilik suara (petugas): scan kartu → pemilih memilih di layar → kembali ke scan
async function pageBilik(p) {
  const els = await GET('/evote');
  const e = els.find(x => x.id === Number(p.id));
  if (!e) throw new Error('Pemilihan tidak ditemukan.');
  S.bilik = { e, voter: null };
  S.after = () => bilikScan();
  return `<div style="margin-bottom:10px">${btn('Kembali', 'nav', { page: 'evoting' }, 'ghost sm', 'back')}</div><div id="bilik" style="max-width:900px;margin:0 auto"></div>`;
}
function bilikScan() {
  const b = $('#bilik'); if (!b) return;
  S.bilik.voter = null;
  b.innerHTML = `<section class="card" style="text-align:center;padding:32px"><h1>${esc(S.bilik.e.judul)}</h1><p class="muted" style="margin:6px 0 18px">Scan kartu pemilih untuk membuka surat suara.</p>
    <input class="scan-input" id="bilikIn" placeholder="Scan kartu…" autocomplete="off" style="max-width:420px">
    <div class="acts" style="justify-content:center;margin-top:12px">${btn('Pakai kamera', 'bilikCam', {}, 'ghost sm', 'eye')}</div><video class="scan-video" id="scanVideo" playsinline muted hidden style="max-width:420px;margin:12px auto 0"></video></section>`;
  const inp = $('#bilikIn'); inp.focus();
  onScan(inp, bilikOpen);
}
ACT.bilikCam = async () => { const v = $('#scanVideo'); v.hidden = false; try { await Scanner.start(v, code => { Scanner.stop(); bilikOpen(code); }); } catch (e) { v.hidden = true; toast(e.message, 'bad'); } };
async function bilikOpen(token) {
  let r;
  try { r = await POST(`/evote/${S.bilik.e.id}/bilik`, { token }); } catch (e) { beep(false); toast(e.message, 'bad'); return bilikScan(); }
  beep();
  S.bilik.voter = { token, ...r };
  $('#bilik').innerHTML = `<section class="card" style="padding:26px"><div class="card-head"><div><h2>Surat suara</h2><small class="muted">Pemilih: <b>${esc(r.nama)}</b> ${esc(r.kelas || '')}</small></div>${btn('Batal', 'bilikCancel', {}, 'ghost sm')}</div>
    <div class="grid">${S.bilik.e.kandidat.map(c => `<button type="button" class="card" data-act="bilikVote" data-id="${c.id}" style="cursor:pointer;font:inherit;color:inherit;border:2px solid transparent;text-align:center;display:flex;flex-direction:column;gap:8px;align-items:center">
      ${c.foto ? `<img src="${esc(c.foto)}" alt="" style="width:130px;height:160px;object-fit:cover;border-radius:12px">` : avatar(null, c.nama, 110)}
      <div style="font-size:34px;font-weight:800;color:var(--brand)">${c.no}</div><b style="font-size:17px">${esc(c.nama)}</b>${c.wakil ? `<small class="muted">& ${esc(c.wakil)}</small>` : ''}</button>`).join('')}</div>
    <p class="muted" style="text-align:center;margin-top:14px">Sentuh foto kandidat pilihan Anda.</p></section>`;
}
ACT.bilikCancel = () => bilikScan();
ACT.bilikVote = async el => {
  const c = S.bilik.e.kandidat.find(x => x.id === Number(el.dataset.id));
  if (!(await confirmBox(`Pilih <b>${c.no}. ${esc(c.nama)}</b>?`, 'Ya, pilih', ''))) return;
  await POST(`/evote/${S.bilik.e.id}/bilik`, { token: S.bilik.voter.token, candidate_id: c.id });
  $('#bilik').innerHTML = `<section class="card grad-result" style="padding:40px"><div class="stat" style="justify-content:center;border:0;box-shadow:none;background:none"><div class="ic">${ico('check')}</div></div><h2>Terima kasih!</h2><p class="muted">Suara tercatat secara rahasia.</p></section>`;
  setTimeout(bilikScan, 3000);
};

Object.assign(PAGES.admin, { evoting: pageEvoting, bilik: pageBilik });
Object.assign(PAGES.guru, { evoting: pageEvoting });
Object.assign(PAGES.siswa, { evoting: pageEvoting });
