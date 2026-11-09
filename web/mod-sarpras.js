'use strict';
/* Modul Sarana: Inventaris Sarpras (KIB A–F, foto, label QR) dan Perpustakaan (katalog, e-book, sirkulasi) */

addNav('admin', ['aset', 'Inventaris (KIB)', 'db', 'sarpras', 'Sarana']);
addNav('admin', ['perpus', 'Perpustakaan', 'book', 'perpus', 'Sarana']);
addNav('guru', ['perpus', 'Perpustakaan', 'book', 'perpus', 'Sarana']);
addNav('siswa', ['perpus', 'Perpustakaan', 'book', 'perpus', 'Sarana']);


// =====================================================================
// INVENTARIS SARPRAS (KIB)
// =====================================================================
const KIB = {
  A: ['Tanah', [['luas_m2', 'Luas (m²)'], ['status_hak', 'Status hak'], ['no_sertifikat', 'No. sertifikat'], ['tgl_sertifikat', 'Tgl. sertifikat'], ['penggunaan', 'Penggunaan']]],
  B: ['Peralatan dan Mesin', [['ukuran', 'Ukuran/CC'], ['bahan', 'Bahan'], ['no_pabrik', 'No. pabrik/seri'], ['no_rangka', 'No. rangka'], ['no_mesin', 'No. mesin'], ['no_polisi', 'No. polisi'], ['no_bpkb', 'No. BPKB']]],
  C: ['Gedung dan Bangunan', [['bertingkat', 'Bertingkat/tidak'], ['beton', 'Beton/tidak'], ['luas_lantai', 'Luas lantai (m²)'], ['no_dokumen', 'No. dokumen gedung'], ['luas_tanah', 'Luas tanah (m²)'], ['status_tanah', 'Status tanah']]],
  D: ['Jalan, Irigasi, dan Jaringan', [['konstruksi', 'Konstruksi'], ['panjang_m', 'Panjang (m)'], ['lebar_m', 'Lebar (m)'], ['luas_m2', 'Luas (m²)'], ['no_dokumen', 'No. dokumen'], ['status_tanah', 'Status tanah']]],
  E: ['Aset Tetap Lainnya', [['judul_pencipta', 'Judul/pencipta (buku)'], ['spesifikasi', 'Spesifikasi'], ['asal_daerah', 'Asal daerah (kesenian)'], ['bahan', 'Bahan'], ['jenis_hewan', 'Jenis hewan/tumbuhan'], ['ukuran', 'Ukuran']]],
  F: ['Konstruksi Dalam Pengerjaan', [['bangunan', 'Bangunan (P/SP/D)'], ['bertingkat', 'Bertingkat/tidak'], ['beton', 'Beton/tidak'], ['luas_m2', 'Luas (m²)'], ['tgl_mulai', 'Tgl. mulai'], ['status_tanah', 'Status tanah'], ['nilai_kontrak', 'Nilai kontrak']]],
};
const KONDISI = [['baik', 'Baik'], ['rusak_ringan', 'Rusak ringan'], ['rusak_berat', 'Rusak berat']];
const kondisiTag = k => tag(KONDISI.find(x => x[0] === k)?.[1] || k, k === 'baik' ? 'ok' : k === 'rusak_ringan' ? 'warn' : 'bad');

async function pageAset() {
  const kib = 'ABCDEF'.includes(S.params.tab || '-') && S.params.tab ? S.params.tab : 'all';
  const pg = await GETpage('/aset?' + (kib !== 'all' ? 'kib=' + kib + '&' : '') + new URLSearchParams(Object.entries({ q: S.params.q || '', limit: S.params.lim || 300 }).filter(([, v]) => v !== '')).toString());
  const rows = pg.rows, total = pg.total;
  S.aset = rows;
  S.asetMap = Object.fromEntries(rows.map(r => [r.id, r]));
  const nilai = Number(pg.headers.get('X-Total-Nilai') || rows.reduce((s, r) => s + r.harga * r.jumlah, 0));
  const rusak = rows.filter(r => r.kondisi !== 'baik').length;
  return head('Inventaris Sarpras', 'Kartu Inventaris Barang (KIB) A–F, lengkap dengan foto dan label QR.',
      btn('Cetak label QR', 'asetLabels', {}, 'ghost', 'print') + btn('Cetak KIB', 'asetKIB', {}, 'ghost', 'print') + btn('Tambah barang', 'asetForm', { kib: kib === 'all' ? 'B' : kib }, '', 'plus')) +
    tabs([['all', 'Semua'], ...Object.entries(KIB).map(([k, [n]]) => [k, `KIB ${k}`])], kib) +
    `<div class="stats">${stat('db', total, kib === 'all' ? 'Jenis barang' : 'KIB ' + kib + ' · ' + KIB[kib][0])}${statRp('chart', nilai, 'Nilai aset')}${stat('trash', rusak, 'Perlu perbaikan')}</div>` +
    (rows.length ? `${pagerBar('aset', total, rows.length, 'Cari kode, nama, lokasi, merk… (Enter)')}
    <div class="tw"><table id="tblA"><thead><tr><th>Barang</th><th>KIB</th><th>Lokasi</th><th>Tahun</th><th class="n">Jumlah</th><th>Kondisi</th><th class="n">Harga satuan</th><th></th></tr></thead><tbody>
    ${rows.map(r => `<tr><td><div class="cell-user">${r.foto ? `<img class="avatar" src="${esc(r.foto)}" alt="">` : `<span class="avatar">${ico('db')}</span>`}<div><b>${esc(r.nama)}</b><small class="muted"><code>${esc(r.kode)}</code> ${esc(r.merk || '')}</small></div></div></td>
      <td>${tag(r.kib, 'brand')}</td><td>${esc(r.lokasi || '–')}</td><td>${esc(r.tahun || '–')}</td><td class="n">${r.jumlah} ${esc(r.satuan || '')}</td><td>${kondisiTag(r.kondisi)}</td><td class="n">${rupiah(r.harga)}</td>
      <td><div class="acts">${iconBtn('edit', 'asetForm', { id: r.id }, 'Ubah')}${iconBtn('trash', 'asetDel', { id: r.id, nama: r.nama }, 'Hapus', 'danger')}</div></td></tr>`).join('')}
    </tbody></table></div>${moreBtn('aset', total, rows.length)}` : `<div class="card">${empty('Belum ada barang di daftar ini.', 'db')}</div>`);
}
const kibDetailHtml = (kib, det = {}) => `<div class="form" id="kibDetail" style="gap:12px;background:var(--surface-2);padding:14px;border-radius:14px">
  <b style="font-size:13px">Rincian KIB ${kib} · ${esc(KIB[kib][0])}</b><div class="row2">${KIB[kib][1].map(([k, l]) =>
    `<label class="f"><span>${esc(l)}</span><input name="det_${k}" value="${esc(det[k] || '')}" style="background:var(--surface)"></label>`).join('')}</div></div>`;
ACT.asetForm = async el => {
  const [kat, sat, dana, lok] = await Promise.all([saran('aset.kategori', ['Elektronik', 'Mebel', 'Alat peraga', 'Alat olahraga', 'Alat kantor', 'Kendaraan', 'Alat laboratorium']),
    saran('aset.satuan', ['unit', 'buah', 'set', 'paket', 'lembar', 'meter', 'm²']), saran('aset.sumber_dana', ['BOSP', 'APBD', 'APBN', 'DAK', 'Komite', 'Hibah', 'Swadaya']), saran('aset.lokasi')]);
  const a = el.dataset.id ? S.asetMap[el.dataset.id] : { kib: el.dataset.kib || 'B', kondisi: 'baik', jumlah: 1, satuan: 'unit', tahun: String(new Date().getFullYear()), detail: {} };
  openForm({
    title: a.id ? 'Ubah barang' : 'Tambah barang', values: a, wide: true,
    fields: [[{ n: 'kode', l: 'Kode barang', req: true, ph: 'INV-B-001', hint: 'Dicetak sebagai QR di label.' }, { n: 'nama', l: 'Nama barang', req: true }],
      [{ n: 'kib', l: 'Jenis KIB', t: 'select', o: Object.entries(KIB).map(([k, [n]]) => [k, `${k} · ${n}`]) }, { n: 'kategori', l: 'Kategori', t: 'combo', o: kat }],
      [{ n: 'merk', l: 'Merk/tipe' }, { n: 'tahun', l: 'Tahun perolehan', t: 'select', o: yearOpts(60).map(y => [y, y]), blank: '—' }],
      [{ n: 'jumlah', l: 'Jumlah', t: 'number', min: 1, req: true }, { n: 'satuan', l: 'Satuan', t: 'combo', o: sat }],
      [{ n: 'harga', l: 'Harga satuan (Rp)', t: 'number', min: 0, step: 'any' }, { n: 'sumber_dana', l: 'Sumber dana', t: 'combo', o: dana }],
      [{ n: 'kondisi', l: 'Kondisi', t: 'select', o: KONDISI }, { n: 'lokasi', l: 'Lokasi/ruang', t: 'combo', o: lok, ph: 'Pilih atau ketik ruang baru…' }],
      { n: 'foto', l: 'Foto barang', t: 'file', accept: 'image/*' },
      { t: 'html', html: kibDetailHtml(a.kib, a.detail || {}) },
      { n: 'keterangan', l: 'Keterangan', t: 'textarea', rows: 2 }],
    onSubmit: (v, form) => {
      const detail = Object.fromEntries($$('input[name^=det_]', form).map(i => [i.name.slice(4), i.value.trim()]));
      const body = { ...v, harga: v.harga || 0, jumlah: v.jumlah || 1, detail };
      return a.id ? PUT('/aset/' + a.id, body) : POST('/aset', body);
    },
  });
  // ganti rincian saat jenis KIB diubah
  const sel = $$('.modal select[name=kib]').pop();
  sel?.addEventListener('change', () => { $$('.modal #kibDetail').pop().outerHTML = kibDetailHtml(sel.value, a.detail || {}); });
};
ACT.asetDel = el => del('/aset/' + el.dataset.id, `Hapus <b>${esc(el.dataset.nama)}</b> dari inventaris?`);
ACT.asetLabels = () => {
  if (!S.aset.length) return toast('Tidak ada barang untuk dicetak.', 'bad');
  const lab = r => `<div class="label">${qrImg(r.kode, 84)}<div><b>${esc(r.kode)}</b><br>${esc(r.nama)}<br>${esc(r.lokasi || '')}<br><small>${esc(S.settings.nama_sekolah || '')}</small></div></div>`;
  const pages = [];
  for (let i = 0; i < S.aset.length; i += 24) pages.push(`<div class="pg"><div class="labels">${S.aset.slice(i, i + 24).map(lab).join('')}</div></div>`);
  printDoc(pages.join(''));
};
ACT.asetKIB = () => {
  const groups = {};
  S.aset.forEach(r => (groups[r.kib] = groups[r.kib] || []).push(r));
  if (!Object.keys(groups).length) return toast('Tidak ada barang untuk dicetak.', 'bad');
  const st = S.settings;
  printDoc(Object.keys(groups).sort().map(k => {
    const cols = KIB[k][1];
    const rows = groups[k];
    const total = rows.reduce((s, r) => s + r.harga * r.jumlah, 0);
    return `<div class="pg">${schoolHead()}<h3 style="text-align:center;margin:0">KARTU INVENTARIS BARANG (KIB) ${k}</h3><p style="text-align:center;margin:2px 0 10px">${esc(KIB[k][0].toUpperCase())}</p>
      <table><thead><tr><th>No</th><th>Kode</th><th>Nama barang</th><th>Merk/tipe</th>${cols.map(([, l]) => `<th>${esc(l)}</th>`).join('')}<th>Tahun</th><th>Jml</th><th>Sumber dana</th><th>Harga (Rp)</th><th>Kondisi</th><th>Ket.</th></tr></thead><tbody>
      ${rows.map((r, i) => `<tr><td>${i + 1}</td><td>${esc(r.kode)}</td><td>${esc(r.nama)}</td><td>${esc(r.merk || '')}</td>${cols.map(([c]) => `<td>${esc((r.detail || {})[c] || '')}</td>`).join('')}
        <td>${esc(r.tahun || '')}</td><td>${r.jumlah} ${esc(r.satuan || '')}</td><td>${esc(r.sumber_dana || '')}</td><td style="text-align:right">${Math.round(r.harga * r.jumlah).toLocaleString('id-ID')}</td>
        <td>${esc(KONDISI.find(x => x[0] === r.kondisi)?.[1] || '')}</td><td>${esc(r.keterangan || '')}</td></tr>`).join('')}
      <tr><td colspan="${8 + cols.length}" style="text-align:right"><b>Jumlah</b></td><td style="text-align:right"><b>${Math.round(total).toLocaleString('id-ID')}</b></td><td colspan="2"></td></tr></tbody></table>
      <div class="sign"><div>Mengetahui,<br>Kepala Sekolah<br><br><br><br><b><u>${esc(st.kepala_sekolah || '..............................')}</u></b><br>NIP. ${esc(st.nip_kepala || '')}</div>
      <div>${fmtDate(today())}<br>Pengurus Barang<br><br><br><br>..............................</div></div></div>`;
  }).join(''));
};

// =====================================================================
// PERPUSTAKAAN — admin
// =====================================================================
async function pagePerpus() {
  if (!isAdm('perpus')) return perpusAnggota();
  const tab = S.params.tab || 'sirkulasi';
  const t = tabs([['sirkulasi', 'Sirkulasi'], ['katalog', 'Katalog Buku'], ['ebook', 'E-book'], ['pinjaman', 'Peminjaman'], ['atur', 'Pengaturan']], tab);
  const hd = head('Perpustakaan', 'Pinjam & kembalikan dengan scan kartu siswa dan label buku (scanner USB atau kamera).');
  if (tab === 'katalog') return hd + t + await perpusKatalog();
  if (tab === 'ebook') return hd + t + await perpusRak(true);
  if (tab === 'pinjaman') return hd + t + await perpusPinjaman();
  if (tab === 'atur') return hd + t + await perpusAtur();
  return hd + t + perpusSirkulasi();
}

S.circ = { member: null, log: [] };
function perpusSirkulasi() {
  S.after = () => {
    const m = $('#circMember'), b = $('#circBook');
    (S.circ.member ? b : m)?.focus();
    m?.addEventListener('keydown', e => { if (e.key === 'Enter') { e.preventDefault(); const v = m.value.trim(); m.value = ''; if (v) circMember(v); } });
    b?.addEventListener('keydown', e => { if (e.key === 'Enter') { e.preventDefault(); const v = b.value.trim(); b.value = ''; if (v) circBorrow(v); } });
  };
  const mb = S.circ.member;
  return `<div class="cols"><section class="card scan-box">
      <label class="f"><span>1. Anggota</span><input class="scan-input" id="circMember" placeholder="Scan kartu siswa / ketik NIS" autocomplete="off"></label>
      <div id="circMemberBox">${mb ? memberBox(mb) : '<p class="muted">Belum ada anggota dipilih.</p>'}</div>
      <label class="f"><span>2. Buku yang dipinjam</span><input class="scan-input" id="circBook" placeholder="Scan label buku / ketik kode" autocomplete="off" ${mb ? '' : 'disabled'}></label>
      <div class="acts" style="justify-content:flex-start">${btn('Pakai kamera', 'circCam', {}, 'ghost sm', 'eye')}${btn('Matikan kamera', 'camStop', {}, 'ghost sm')}${mb ? btn('Ganti anggota', 'circReset', {}, 'ghost sm') : ''}</div>
      <video class="scan-video" id="scanVideo" playsinline muted hidden></video></section>
    <section class="card"><div class="card-head"><h3>Riwayat sesi ini</h3></div>${S.circ.log.length ? S.circ.log.slice(0, 15).map(l => `<div class="li"><div><b>${esc(l.t)}</b><small class="muted">${esc(l.s)}</small></div><small>${l.at}</small></div>`).join('') : empty('Belum ada transaksi.', 'book')}</section></div>`;
}
function memberBox(d) {
  const a = d.anggota;
  return `<div class="scan-result ok">${avatar(a.foto, a.nama_lengkap, 56)}<div style="flex:1"><b>${esc(a.nama_lengkap)}</b>${esc(a.kelas_nama || (a.role === 'guru' ? 'Guru' : ''))} · ${esc(a.username)}</div></div>
    ${d.pinjaman.length ? `<div style="margin-top:8px">${d.pinjaman.map(l => { const late = l.jatuh_tempo < today();
      return `<div class="li"><div><b>${esc(l.judul)}</b><small class="muted">${esc(l.kode)} · kembali ${fmtDate(l.jatuh_tempo)} ${late ? tag('Terlambat', 'bad') : ''}</small></div>${btn('Kembalikan', 'circReturn', { id: l.id, judul: l.judul }, 'sm')}</div>`; }).join('')}</div>`
      : '<p class="muted" style="margin-top:8px">Tidak ada pinjaman aktif.</p>'}`;
}
const circLog = (t, s) => S.circ.log.unshift({ t, s, at: new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' }) });
// Saat kamera aktif, perbarui bagian halaman saja (refresh penuh akan mematikan kamera).
function circUpdate() {
  if (!Scanner.stream || !$('#circMemberBox')) return refresh(true);
  const mb = S.circ.member;
  $('#circMemberBox').innerHTML = mb ? memberBox(mb) : '<p class="muted">Belum ada anggota dipilih.</p>';
  const b = $('#circBook'); if (b) b.disabled = !mb;
  const log = $('#main .cols > section:last-child');
  if (log) log.innerHTML = `<div class="card-head"><h3>Riwayat sesi ini</h3></div>${S.circ.log.length ? S.circ.log.slice(0, 15).map(l => `<div class="li"><div><b>${esc(l.t)}</b><small class="muted">${esc(l.s)}</small></div><small>${l.at}</small></div>`).join('') : empty('Belum ada transaksi.', 'book')}`;
}
async function circMember(q) {
  try { S.circ.member = await GET('/perpus/anggota?q=' + encodeURIComponent(q)); beep(); }
  catch (e) { beep(false); toast(e.message, 'bad'); return; }
  circUpdate();
}
async function circBorrow(kode) {
  const m = S.circ.member;
  if (!m) return;
  try {
    const r = await POST('/perpus/pinjam', { user_id: m.anggota.id, buku: kode });
    beep(); circLog('Pinjam: ' + r.judul, `${m.anggota.nama_lengkap} · kembali ${fmtDate(r.jatuh_tempo)}`);
    toast(`Dipinjam: ${r.judul}`);
    S.circ.member = await GET('/perpus/anggota?q=' + encodeURIComponent(m.anggota.username));
  } catch (e) { beep(false); toast(e.message, 'bad'); }
  circUpdate();
}
ACT.circReturn = async el => {
  const r = await POST('/perpus/kembali/' + el.dataset.id);
  circLog('Kembali: ' + el.dataset.judul, r.denda ? `terlambat ${r.terlambat_hari} hari · denda ${rupiah(r.denda)}` : 'tepat waktu');
  toast(r.denda ? `Dikembalikan. Terlambat ${r.terlambat_hari} hari, denda ${rupiah(r.denda)}.` : 'Buku dikembalikan.');
  if (S.circ.member) S.circ.member = await GET('/perpus/anggota?q=' + encodeURIComponent(S.circ.member.anggota.username));
  if (S.page === 'perpus' && (S.params.tab || 'sirkulasi') === 'sirkulasi') circUpdate(); else refresh(true);
};
ACT.circReset = () => { S.circ.member = null; refresh(true); };
ACT.circCam = async () => {
  const v = $('#scanVideo');
  v.hidden = false;
  try { await Scanner.start(v, code => (S.circ.member && !code.startsWith('S1.') ? circBorrow(code) : circMember(code))); }
  catch (e) { v.hidden = true; toast(e.message, 'bad'); }
};

async function perpusKatalog() {
  const pg = await GETpage('/buku?' + new URLSearchParams(Object.entries({ q: S.params.q || '', limit: S.params.lim || 300 }).filter(([, v]) => v !== '')).toString());
  const bs = pg.rows;
  S.buku = Object.fromEntries(bs.map(b => [b.id, b]));
  S.bukuList = bs;
  return `<div class="card-head"><p class="muted">${bs.length} judul · ${bs.reduce((s, b) => s + b.stok, 0)} eksemplar</p>
    <div class="acts">${btn('Impor e-book Kemendikbud', 'sibiOpen', {}, 'ghost sm', 'download')}${btn('Cetak label buku', 'bukuLabels', {}, 'ghost sm', 'print')}${btn('Tambah buku', 'bukuForm', {}, 'sm', 'plus')}</div></div>` +
    (bs.length ? `${pagerBar('perpus', pg.total, bs.length, 'Cari judul, pengarang, kode, ISBN… (Enter)')}
    <div class="tw"><table id="tblB"><thead><tr><th>Buku</th><th>Kategori</th><th>Rak</th><th class="n">Tersedia</th><th></th></tr></thead><tbody>
    ${bs.map(b => `<tr><td><div class="cell-user">${bookCover(b) ? `<img class="avatar" src="${esc(bookCover(b))}" alt="" loading="lazy" style="height:52px;width:40px;object-fit:cover">` : `<span class="avatar">${ico('book')}</span>`}
      <div><b>${esc(b.judul)}</b><small class="muted"><code>${esc(b.kode)}</code> ${esc(b.pengarang || '')}${b.tahun ? ' · ' + esc(b.tahun) : ''}</small>${hasEbook(b) ? tag(b.ebook_url ? 'E-book Kemendikbud' : 'E-book', 'info') : ''}</div></div></td>
      <td>${esc(b.kategori || '–')}</td><td>${esc(b.rak || '–')}</td><td class="n">${!b.stok && hasEbook(b) ? 'Digital' : `${b.tersedia}/${b.stok}`}</td>
      <td><div class="acts">${iconBtn('edit', 'bukuForm', { id: b.id }, 'Ubah')}${iconBtn('trash', 'bukuDel', { id: b.id, judul: b.judul }, 'Hapus', 'danger')}</div></td></tr>`).join('')}
    </tbody></table></div>${moreBtn('perpus', pg.total, bs.length)}` : `<div class="card">${empty('Belum ada buku.', 'book')}</div>`);
}
ACT.bukuForm = async el => {
  const [kat, rak, pen] = await Promise.all([saran('buku.kategori', ['Pelajaran', 'Fiksi', 'Nonfiksi', 'Referensi', 'Sains', 'Agama', 'Sejarah', 'Biografi', 'Kamus/Ensiklopedia']),
    saran('buku.rak'), saran('buku.penerbit')]);
  const b = el.dataset.id ? S.buku[el.dataset.id] : { stok: 1 };
  openForm({
    title: b.id ? 'Ubah buku' : 'Tambah buku', values: b, wide: true,
    fields: [[{ n: 'kode', l: 'Kode buku', req: true, ph: 'BK-0001', hint: 'Dicetak sebagai QR pada label buku.' }, { n: 'isbn', l: 'ISBN' }],
      { n: 'judul', l: 'Judul', req: true }, [{ n: 'pengarang', l: 'Pengarang' }, { n: 'penerbit', l: 'Penerbit', t: 'combo', o: pen }],
      [{ n: 'tahun', l: 'Tahun terbit', t: 'select', o: yearOpts(80).map(y => [y, y]), blank: '—' }, { n: 'kategori', l: 'Kategori', t: 'combo', o: kat }],
      [{ n: 'rak', l: 'Rak', t: 'combo', o: rak }, { n: 'stok', l: 'Jumlah eksemplar', t: 'number', min: 0, req: true }],
      [{ n: 'cover', l: 'Sampul', t: 'file', accept: 'image/*' }, { n: 'ebook', l: 'File e-book (PDF)', t: 'file', accept: 'application/pdf', hint: 'Maks 30 MB. Siswa bisa membaca langsung.' }],
      { n: 'sinopsis', l: 'Sinopsis', t: 'textarea', rows: 3 }],
    onSubmit: v => { const body = { ...v, stok: v.stok ?? 0 }; return b.id ? PUT('/buku/' + b.id, body) : POST('/buku', body); },
  });
};
ACT.bukuDel = el => del('/buku/' + el.dataset.id, `Hapus buku <b>${esc(el.dataset.judul)}</b> beserta riwayat peminjamannya?`);
ACT.bukuLabels = () => {
  if (!S.bukuList.length) return toast('Belum ada buku.', 'bad');
  const lab = b => `<div class="label">${qrImg(b.kode, 84)}<div><b>${esc(b.kode)}</b><br>${esc(b.judul.slice(0, 50))}<br>${esc(b.rak ? 'Rak ' + b.rak : '')}<br><small>Perpustakaan ${esc(S.settings.nama_sekolah || '')}</small></div></div>`;
  const items = S.bukuList.filter(b => b.stok > 0 || !hasEbook(b)).flatMap(b => Array.from({ length: Math.max(1, b.stok) }, () => b));
  const pages = [];
  for (let i = 0; i < items.length; i += 24) pages.push(`<div class="pg"><div class="labels">${items.slice(i, i + 24).map(lab).join('')}</div></div>`);
  printDoc(pages.join(''));
};

async function perpusPinjaman() {
  const f = ['aktif', 'terlambat', 'semua'].includes(S.params.status) ? S.params.status : 'aktif';
  const ls = await GET('/perpus/pinjaman?status=' + f);
  return `<div class="bar"><select data-change="loanFilter" aria-label="Saring">${opts([['aktif', 'Sedang dipinjam'], ['terlambat', 'Terlambat'], ['semua', 'Semua (500 terakhir)']], f)}</select>
    <input class="search" type="search" placeholder="Cari peminjam atau judul…" data-filter="tblL" aria-label="Cari"></div>` +
    (ls.length ? `<div class="tw"><table id="tblL"><thead><tr><th>Peminjam</th><th>Buku</th><th>Pinjam</th><th>Jatuh tempo</th><th>Status</th><th></th></tr></thead><tbody>
    ${ls.map(l => { const late = !l.tgl_kembali && l.jatuh_tempo < today();
      return `<tr><td><b>${esc(l.nama_lengkap)}</b><small class="muted">${esc(l.kelas_nama || l.username)}</small></td><td>${esc(l.judul)}<small class="muted">${esc(l.kode)}</small></td>
      <td>${fmtDate(l.tgl_pinjam)}</td><td>${fmtDate(l.jatuh_tempo)}</td>
      <td>${l.tgl_kembali ? tag('Kembali ' + fmtDate(l.tgl_kembali), 'ok') + (l.denda ? ' ' + tag(rupiah(l.denda), 'warn') : '') : late ? tag('Terlambat', 'bad') : tag('Dipinjam', 'info')}</td>
      <td>${l.tgl_kembali ? '' : btn('Kembalikan', 'circReturn', { id: l.id, judul: l.judul }, 'ghost sm')}</td></tr>`; }).join('')}
    </tbody></table></div>` : `<div class="card">${empty('Tidak ada data.', 'book')}</div>`);
}
CHG.loanFilter = el => go('perpus', { tab: 'pinjaman', status: el.value });

async function perpusAtur() {
  const c = await GET('/config/perpus');
  const fields = [{ n: 'lama_pinjam', l: 'Lama peminjaman (hari)', t: 'number', min: 1, req: true }, { n: 'denda_per_hari', l: 'Denda keterlambatan per hari (Rp)', t: 'number', min: 0, req: true },
    { n: 'maks_pinjam', l: 'Maksimal buku dipinjam per anggota', t: 'number', min: 1, req: true }];
  const id = registerForm(fields, async v => { await PUT('/config/perpus', v); toast('Pengaturan perpustakaan tersimpan.'); });
  return `<section class="card" style="max-width:520px">${formHtml(id, fields, c, { cancel: false })}</section>`;
}

// =====================================================================
// PERPUSTAKAAN — siswa & guru
// =====================================================================
// sampul & e-book: lewat server sekolah (file unggahan sendiri atau salinan dari Buku Kemendikbud)
const hasEbook = b => !!(b.ebook || b.ebook_url);
const bookCover = b => b.cover || b.cover_url ? `/api/buku/${b.id}/sampul` : '';
const bookRead = b => `/api/buku/${b.id}/baca`;
const bookCard = b => `<section class="card book-card">
      ${bookCover(b) ? `<img src="${esc(bookCover(b))}" alt="" loading="lazy" data-ph="book">` : `<span class="avatar book-ph">${ico('book')}</span>`}
      <div class="book-info"><b>${esc(b.judul)}</b><small class="muted">${esc(b.pengarang || '')}${b.tahun ? ' · ' + esc(b.tahun) : ''}</small>
        <div>${!b.stok && hasEbook(b) ? tag('Digital', 'info') : b.tersedia > 0 ? tag(`Tersedia ${b.tersedia}`, 'ok') : tag('Sedang dipinjam', 'warn')}
          ${b.rak ? tag('Rak ' + b.rak) : ''}${b.jenjang ? tag(b.jenjang) : ''}${b.kelas ? tag('Kelas ' + b.kelas) : ''}</div>
        ${hasEbook(b) ? `<a class="btn sm" href="${esc(bookRead(b))}" target="_blank" rel="noopener" style="align-self:flex-start;margin-top:4px">${ico('file')} Baca e-book</a>` : ''}</div></section>`;
// sampul gagal dimuat (sumber tidak bisa dihubungi): ganti ikon buku. CSP melarang onerror inline, jadi lewat listener.
document.addEventListener('error', e => {
  const img = e.target;
  if (img.tagName === 'IMG' && img.dataset.ph === 'book') { const s = document.createElement('span'); s.className = 'avatar book-ph'; s.innerHTML = ico('book'); img.replaceWith(s); }
}, true);

// rak buku: siswa/guru (semua buku atau e-book saja) dan tab E-book pengelola
async function perpusRak(adminEbook = false) {
  const ebook = adminEbook || S.params.rak === 'ebook';
  const qs = new URLSearchParams(Object.entries({ q: S.params.q || '', limit: S.params.lim || 300, ebook: ebook ? 1 : '' }).filter(([, v]) => v !== '')).toString();
  const pg = await GETpage('/buku?' + qs);
  const bs = pg.rows;
  const chips = adminEbook ? '' : `<div class="acts" style="justify-content:flex-start;margin-bottom:10px">${[['', 'Semua buku'], ['ebook', 'E-book']].map(([k, l]) =>
    btn(l, 'go', { page: 'perpus', params: JSON.stringify({ rak: k }) }, (S.params.rak || '') === k ? 'sm' : 'ghost sm')).join('')}</div>`;
  const tools = adminEbook ? `<div class="card-head"><p class="muted">E-book dibaca lewat server sekolah. Buku dari Kemendikbud diunduh sekali saat pertama dibuka, lalu tersimpan untuk pembaca berikutnya.</p>
    <div class="acts">${btn('Impor e-book Kemendikbud', 'sibiOpen', {}, 'sm', 'download')}</div></div>` : '';
  return tools + chips + `${pagerBar('perpus', pg.total, bs.length, 'Cari judul, pengarang, kategori… (Enter)')}
    ${bs.length ? `<div class="grid" id="bookGrid">${bs.map(bookCard).join('')}</div>${moreBtn('perpus', pg.total, bs.length)}`
      : `<div class="card">${empty(ebook ? 'Belum ada e-book.' + (isAdm('perpus') ? ' Klik "Impor e-book Kemendikbud".' : '') : 'Katalog perpustakaan masih kosong.', 'book')}</div>`}`;
}
async function perpusAnggota() {
  const ls = await GET('/perpus/me');
  const aktif = ls.filter(l => !l.tgl_kembali);
  return head('Perpustakaan', 'Cari buku, baca e-book, dan lihat pinjaman Anda.') +
    (aktif.length ? `<section class="card" style="margin-bottom:16px"><div class="card-head"><h3>Sedang Anda pinjam</h3></div>${aktif.map(l => { const late = l.jatuh_tempo < today();
      return `<div class="li"><div><b>${esc(l.judul)}</b><small class="muted">${esc(l.pengarang || '')} · kembalikan ${fmtDate(l.jatuh_tempo)}</small></div>${late ? tag('Terlambat', 'bad') : tag('Dipinjam', 'info')}</div>`; }).join('')}</section>` : '') +
    await perpusRak();
}

// ---- Impor e-book dari Buku Kemendikbud (SIBI). Yang disimpan hanya data & alamat; PDF diunduh saat pertama dibaca.
const SIBI_JENIS = [['kurmer', 'Kurikulum Merdeka'], ['k13', 'Kurikulum 2013'], ['nonteks', 'Buku bacaan (nonteks)']];
S.sibi = { jenis: 'kurmer', jenjang: '', kelas: '', tipe: '', q: '', sel: new Set() };
ACT.sibiOpen = () => { S.sibi.sel = new Set(); S.sibiModal = openModal({ title: 'Impor e-book dari Buku Kemendikbud', wide: true, builder: sibiBody }); };
async function sibiBody() {
  const f = S.sibi;
  const r = await GET('/perpus/sibi?' + new URLSearchParams(Object.entries({ jenis: f.jenis, jenjang: f.jenjang, kelas: f.kelas, tipe: f.tipe, q: f.q }).filter(([, v]) => v)).toString());
  const kelasOpt = [['', 'Semua kelas'], ...Array.from({ length: 12 }, (_, i) => [String(i + 1), 'Kelas ' + (i + 1)])];
  const tipeOpt = [['', 'Siswa & guru'], ['Buku siswa', 'Buku siswa'], ['Buku guru', 'Buku guru']];
  const sel = (k, label, o, v) => `<select data-change="sibiSet" data-k="${k}" aria-label="${label}">${opts(o, v)}</select>`;
  return `<p class="muted" style="margin:0 0 12px">Sumber: <b>buku.kemendikdasmen.go.id</b> (gratis, lisensi terbuka). Buku masuk katalog sebagai e-book digital; file PDF diunduh server sekali saat pertama dibuka lalu disimpan.</p>
    <div class="sibi-filter">${sel('jenis', 'Jenis buku', SIBI_JENIS, f.jenis)}${sel('jenjang', 'Jenjang', [['', 'Semua jenjang'], ...r.jenjang.map(j => [j, j])], f.jenjang)}
      ${f.jenis === 'nonteks' ? '' : sel('kelas', 'Kelas', kelasOpt, f.kelas) + sel('tipe', 'Buku siswa atau guru', tipeOpt, f.tipe)}
      <input type="search" id="sibiQ" placeholder="Cari judul, mapel, penulis… (Enter)" value="${esc(f.q)}" aria-label="Cari buku Kemendikbud"></div>
    <div class="sibi-bar"><small class="muted">${r.total > r.rows.length ? `Menampilkan ${r.rows.length} dari ${r.total} — persempit dengan filter/pencarian` : `${r.total} buku`}</small>
      <span class="acts">${r.rows.some(b => !b.sudah_impor) ? btn('Pilih semua di daftar', 'sibiAll', {}, 'ghost sm') : ''}
      <button class="btn sm" type="button" data-act="sibiImport" id="sibiGo" ${f.sel.size ? '' : 'disabled'}>${ico('download')} Impor terpilih (<span id="sibiN">${f.sel.size}</span>)</button></span></div>
    ${r.rows.length ? `<div class="sibi-list">${r.rows.map(b => `<label class="sibi-item ${b.sudah_impor ? 'done' : ''}">
      <input type="checkbox" data-change="sibiPick" value="${esc(b.id)}" ${b.sudah_impor ? 'disabled checked' : f.sel.has(b.id) ? 'checked' : ''}>
      <span><b>${esc(b.judul)}</b><small class="muted">${[b.tipe, b.jenjang, b.kelas && 'Kelas ' + b.kelas, b.mapel, b.tahun].filter(Boolean).map(esc).join(' · ')}</small></span>
      ${b.sudah_impor ? tag('Sudah ada', 'ok') : ''}</label>`).join('')}</div>` : empty('Tidak ada buku yang cocok.', 'book')}`;
}
CHG.sibiSet = el => {
  S.sibi[el.dataset.k] = el.value;
  if (el.dataset.k === 'jenis') Object.assign(S.sibi, { jenjang: '', kelas: '', tipe: '', sel: new Set() }); // pilihan berlaku per jenis
  refreshModal(S.sibiModal);
};
CHG.sibiPick = el => { el.checked ? S.sibi.sel.add(el.value) : S.sibi.sel.delete(el.value); sibiCount(); };
function sibiCount() { const n = S.sibi.sel.size; $('#sibiN').textContent = n; $('#sibiGo').disabled = !n; }
ACT.sibiAll = () => { $$('.sibi-item input:not(:disabled)').forEach(i => { i.checked = true; S.sibi.sel.add(i.value); }); sibiCount(); };
ACT.sibiImport = async el => {
  el.disabled = true;
  try {
    const r = await POST('/perpus/sibi/impor', { jenis: S.sibi.jenis, ids: [...S.sibi.sel] });
    toast(`${r.baru} e-book masuk katalog${r.dilewati ? `, ${r.dilewati} dilewati (sudah ada)` : ''}.`);
    S.sibi.sel = new Set();
    closeModal(); refresh(true);
  } finally { el.disabled = false; }
};
document.addEventListener('keydown', e => {
  if (e.key === 'Enter' && e.target.id === 'sibiQ') { e.preventDefault(); S.sibi.q = e.target.value.trim(); refreshModal(S.sibiModal); }
});
document.addEventListener('input', e => {
  const el = e.target.closest('[data-filter2]');
  if (!el) return;
  const q = el.value.trim().toLowerCase();
  $$('#' + el.dataset.filter2 + ' > [data-text]').forEach(c => { c.hidden = q && !c.dataset.text.includes(q); });
});

Object.assign(PAGES.admin, { aset: pageAset, perpus: pagePerpus });
Object.assign(PAGES.guru, { perpus: pagePerpus });
Object.assign(PAGES.siswa, { perpus: pagePerpus });
