'use strict';
/* Manajemen ujian: kegiatan, ruang, pembagian peserta, jadwal, cetak kartu/denah/daftar hadir/berita acara */

addNav('admin', ['manujian', 'Manajemen Ujian', 'clip', 'ujian', 'Ujian']);

// =====================================================================
// MANAJEMEN UJIAN (admin)
// =====================================================================
async function pageManUjian(p) {
  if (p.id) return manUjianDetail(Number(p.id));
  const evs = await GET('/ujian-event');
  S.evList = Object.fromEntries(evs.map(e => [e.id, e]));
  return head('Manajemen Ujian', 'Atur ruang, bagi peserta otomatis, lalu cetak kartu peserta, denah, daftar hadir, dan berita acara.', btn('Kegiatan ujian baru', 'evuForm', {}, '', 'plus')) +
    (evs.length ? `<div class="grid">${evs.map(e => `<section class="card course-card" data-act="go" data-page="manujian" data-params='${esc(JSON.stringify({ id: e.id }))}' role="button" tabindex="0">
      <div class="top">${tag(e.semester || 'Ujian', 'brand')}${tag(e.jumlah_peserta + ' peserta')}</div><h3>${esc(e.nama)}</h3>
      <small class="muted">${esc(e.tahun_ajaran || '')}${e.mulai ? ' · ' + fmtDate(e.mulai) + (e.selesai ? ' – ' + fmtDate(e.selesai) : '') : ''} · ${e.jumlah_ruang} ruang</small></section>`).join('')}</div>`
      : `<div class="card">${empty('Belum ada kegiatan ujian. Contoh: PTS Ganjil, PAS, Ujian Sekolah.', 'clip')}</div>`);
}
ACT.evuForm = async el => {
  const tas = await tahunAjaranOpts();
  const e = el.dataset.id ? (S.evDetail?.event || S.evList[el.dataset.id]) : { tahun_ajaran: S.settings.tahun_ajaran || '', semester: 'Ganjil' };
  openForm({
    title: e.id ? 'Ubah kegiatan ujian' : 'Kegiatan ujian baru', values: e,
    fields: [{ n: 'nama', l: 'Nama kegiatan', req: true, ph: 'Penilaian Akhir Semester Ganjil' },
      [{ n: 'tahun_ajaran', l: 'Tahun ajaran', t: 'select', o: tas.map(x => [x, x]) }, { n: 'semester', l: 'Semester', t: 'select', o: [['Ganjil', 'Ganjil'], ['Genap', 'Genap']] }],
      [{ n: 'mulai', l: 'Tanggal mulai', t: 'date' }, { n: 'selesai', l: 'Tanggal selesai', t: 'date' }]],
    onSubmit: v => e.id ? PUT('/ujian-event/' + e.id, v) : POST('/ujian-event', v),
  });
};

async function manUjianDetail(id) {
  const [d, ks] = await Promise.all([GET('/ujian-event/' + id), GET('/classes')]);
  S.evDetail = d; S.evKelas = ks;
  const e = d.event;
  const tab = S.params.tab || 'ruang';
  const roomName = Object.fromEntries(d.rooms.map(r => [r.id, r.nama]));
  let body = '';
  if (tab === 'ruang') {
    body = `<div class="card-head"><p class="muted">Total kapasitas ${d.rooms.reduce((s, r) => s + r.kapasitas, 0)} kursi</p>${btn('Tambah ruang', 'roomForm', {}, 'sm', 'plus')}</div>` +
      (d.rooms.length ? `<div class="tw"><table><thead><tr><th>Ruang</th><th class="n">Kapasitas</th><th class="n">Terisi</th><th class="n">Kolom denah</th><th>Pengawas</th><th></th></tr></thead><tbody>
      ${d.rooms.map(r => `<tr><td><b>${esc(r.nama)}</b></td><td class="n">${r.kapasitas}</td><td class="n">${r.terisi}</td><td class="n">${r.kolom}</td><td>${esc(r.pengawas || '–')}</td>
        <td><div class="acts">${iconBtn('edit', 'roomForm', { id: r.id }, 'Ubah')}${iconBtn('trash', 'roomDel', { id: r.id, nama: r.nama }, 'Hapus', 'danger')}</div></td></tr>`).join('')}</tbody></table></div>`
        : `<div class="card">${empty('Belum ada ruang. Tambahkan ruang ujian terlebih dahulu.', 'school')}</div>`);
  } else if (tab === 'peserta') {
    body = `<div class="card-head"><p class="muted">${d.participants.length} peserta</p>${btn('Bagi peserta otomatis', 'assignForm', {}, 'sm', 'users')}</div>` +
      (d.participants.length ? `<div class="bar"><input class="search" type="search" placeholder="Cari nama, nomor, ruang…" data-filter="tblP" aria-label="Cari"></div>
      <div class="tw"><table id="tblP"><thead><tr><th>No. peserta</th><th>Nama</th><th>Kelas</th><th>Ruang</th><th class="n">Kursi</th></tr></thead><tbody>
      ${d.participants.map(x => `<tr><td><code>${esc(x.no_peserta)}</code></td><td><b>${esc(x.nama_lengkap)}</b></td><td>${esc(x.kelas_nama || '')}</td><td>${esc(roomName[x.room_id] || '–')}</td><td class="n">${x.kursi || ''}</td></tr>`).join('')}
      </tbody></table></div>` : `<div class="card">${empty('Belum ada peserta. Klik "Bagi peserta otomatis".', 'users')}</div>`);
  } else if (tab === 'jadwal') {
    body = `<div class="card-head"><p class="muted">Jadwal dipakai untuk daftar hadir dan berita acara.</p>${btn('Tambah jadwal', 'schedForm', {}, 'sm', 'plus')}</div>` +
      (d.schedule.length ? `<div class="tw"><table><thead><tr><th>Hari, tanggal</th><th>Jam</th><th>Mata pelajaran</th><th>Kelas</th><th></th></tr></thead><tbody>
      ${d.schedule.map(s => `<tr><td>${new Date(s.tanggal + 'T00:00').toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}</td><td>${esc(s.jam_mulai)}–${esc(s.jam_selesai)}</td><td><b>${esc(s.mapel)}</b></td><td>${esc(s.kelas || 'Semua')}</td>
        <td><div class="acts">${iconBtn('trash', 'schedDel', { id: s.id }, 'Hapus', 'danger')}</div></td></tr>`).join('')}</tbody></table></div>` : `<div class="card">${empty('Belum ada jadwal.', 'cal')}</div>`);
  } else {
    const can = d.participants.length > 0;
    body = !can ? `<div class="card">${empty('Bagi peserta terlebih dahulu sebelum mencetak.', 'print')}</div>` : `<div class="grid">
      ${[['Kartu peserta', 'Nomor peserta, ruang, kursi, dan QR. 8 kartu per halaman.', 'prKartu'], ['Denah tempat duduk', 'Satu halaman per ruang, untuk ditempel di pintu.', 'prDenah'],
        ['Daftar hadir', 'Per ruang per jadwal, dengan kolom tanda tangan.', 'prHadir'], ['Berita acara', 'Per ruang per jadwal, diisi pengawas.', 'prBA']].map(([t, s, a]) =>
        `<section class="card" style="display:flex;flex-direction:column;gap:8px"><h3>${t}</h3><small class="muted">${s}</small><div>${btn('Cetak', a, {}, 'sm', 'print')}</div></section>`).join('')}</div>
      ${d.schedule.length ? `<label class="f" style="margin-top:16px;max-width:420px"><span>Jadwal untuk daftar hadir & berita acara</span><select id="prSched">${d.schedule.map(s => `<option value="${s.id}">${fmtDate(s.tanggal)} · ${esc(s.jam_mulai)} · ${esc(s.mapel)}</option>`).join('')}</select></label>`
        : '<p class="hint" style="margin-top:16px">Tambahkan jadwal agar daftar hadir & berita acara terisi mapel dan jam.</p>'}`;
  }
  return `<div style="margin-bottom:10px">${btn('Semua kegiatan', 'nav', { page: 'manujian' }, 'ghost sm', 'back')}</div>` +
    head(e.nama, `${esc(e.tahun_ajaran || '')} ${esc(e.semester || '')}${e.mulai ? ' · ' + fmtDate(e.mulai) + (e.selesai ? ' – ' + fmtDate(e.selesai) : '') : ''}`,
      btn('Ubah', 'evuForm', { id: e.id }, 'ghost', 'edit') + btn('Hapus', 'evuDel', { id: e.id }, 'danger', 'trash')) +
    tabs([['ruang', `Ruang (${d.rooms.length})`], ['peserta', `Peserta (${d.participants.length})`], ['jadwal', `Jadwal (${d.schedule.length})`], ['cetak', 'Cetak']], tab) + body;
}
ACT.evuDel = async el => {
  if (!(await confirmBox('Hapus kegiatan ujian ini beserta ruang, peserta, dan jadwalnya?'))) return;
  await DEL('/ujian-event/' + el.dataset.id); toast('Terhapus.'); go('manujian');
};
ACT.roomForm = async el => {
  const pengawas = uniqList([...await saran('ujian.guru'), ...await saran('ujian.pengawas')]);
  const r = el.dataset.id ? S.evDetail.rooms.find(x => x.id === Number(el.dataset.id)) : { kapasitas: 20, kolom: 4 };
  openForm({
    title: r.id ? 'Ubah ruang' : 'Tambah ruang', values: r,
    fields: [{ n: 'nama', l: 'Nama ruang', req: true, ph: 'Ruang 01' }, [{ n: 'kapasitas', l: 'Kapasitas', t: 'number', min: 1, req: true }, { n: 'kolom', l: 'Jumlah kolom meja (denah)', t: 'number', min: 1, max: 12, req: true }],
      { n: 'pengawas', l: 'Pengawas', t: 'combo', o: pengawas, ph: 'Pilih guru atau ketik nama' }],
    onSubmit: v => r.id ? PUT('/ujian-ruang/' + r.id, v) : POST(`/ujian-event/${S.evDetail.event.id}/ruang`, v),
  });
};
ACT.roomDel = el => del('/ujian-ruang/' + el.dataset.id, `Hapus ruang <b>${esc(el.dataset.nama)}</b>? Peserta di ruang ini perlu dibagi ulang.`);
ACT.assignForm = () => openForm({
  title: 'Bagi peserta otomatis', success: 'Peserta dibagi.',
  intro: '<div class="hint" style="margin-bottom:12px">Pembagian ulang akan mengganti nomor peserta dan kursi yang sudah ada.</div>',
  values: { silang: true, prefix: (S.evDetail.event.semester || 'P').slice(0, 1).toUpperCase() + new Date().getFullYear().toString().slice(2) },
  fields: [{ n: 'kelas_ids', l: 'Kelas peserta', t: 'checks', o: S.evKelas.map(k => [k.id, k.nama]) },
    { n: 'silang', l: 'Silang kelas (siswa dari kelas berbeda duduk berselang-seling)', t: 'check' },
    { n: 'prefix', l: 'Awalan nomor peserta', req: true, hint: 'Contoh: G26 → G26-001, G26-002, …' }],
  onSubmit: v => POST(`/ujian-event/${S.evDetail.event.id}/bagi`, { kelas_ids: v.kelas_ids.map(Number), silang: v.silang, prefix: v.prefix }),
});
ACT.schedForm = async () => {
  const [mapel, kelas] = await Promise.all([mapelOpts(), kelasOpts()]);
  const tingkat = uniqList(S.evKelas.map(k => k.tingkat).filter(Boolean)).map(t => 'Kelas ' + t);
  openForm({
    title: 'Tambah jadwal', values: { tanggal: S.evDetail.event.mulai || today(), jam_mulai: '07:30', jam_selesai: '09:30' },
    fields: [{ n: 'mapel', l: 'Mata pelajaran', req: true, t: 'combo', o: mapel }, { n: 'tanggal', l: 'Tanggal', t: 'date', req: true },
      [{ n: 'jam_mulai', l: 'Jam mulai', req: true, t: 'time' }, { n: 'jam_selesai', l: 'Jam selesai', req: true, t: 'time' }],
      { n: 'kelas', l: 'Kelas/tingkat', t: 'combo', o: ['Semua kelas', ...tingkat, ...kelas] }],
    onSubmit: v => POST(`/ujian-event/${S.evDetail.event.id}/jadwal`, v),
  });
};
ACT.schedDel = el => del('/ujian-jadwal/' + el.dataset.id, 'Hapus jadwal ini?');

// --- Cetak ---
const evRooms = () => S.evDetail.rooms.map(r => ({ ...r, peserta: S.evDetail.participants.filter(p => p.room_id === r.id).sort((a, b) => a.kursi - b.kursi) })).filter(r => r.peserta.length);
const evSched = () => S.evDetail.schedule.find(s => String(s.id) === $('#prSched')?.value);
ACT.prKartu = () => {
  const e = S.evDetail.event, st = S.settings, rn = Object.fromEntries(S.evDetail.rooms.map(r => [r.id, r.nama]));
  const card = p => `<div class="idcard" style="height:6.2cm;align-items:flex-start">${qrImg(p.token, 110)}<div class="t" style="flex:1"><div class="sch">KARTU PESERTA · ${esc(e.nama)}</div>
    <table style="font-size:9pt"><tbody><tr><td style="border:0;padding:1px 4px 1px 0">No.</td><td style="border:0;padding:1px"><b>${esc(p.no_peserta)}</b></td></tr>
    <tr><td style="border:0;padding:1px 4px 1px 0">Nama</td><td style="border:0;padding:1px"><b>${esc(p.nama_lengkap)}</b></td></tr>
    <tr><td style="border:0;padding:1px 4px 1px 0">Kelas</td><td style="border:0;padding:1px">${esc(p.kelas_nama || '')}</td></tr>
    <tr><td style="border:0;padding:1px 4px 1px 0">Ruang</td><td style="border:0;padding:1px">${esc(rn[p.room_id] || '')} · kursi ${p.kursi}</td></tr></tbody></table>
    <div style="font-size:8pt;margin-top:4px">${esc(st.nama_sekolah)}<br>Kepala Sekolah: ${esc(st.kepala_sekolah || '………………')}</div></div></div>`;
  const ps = S.evDetail.participants, pages = [];
  for (let i = 0; i < ps.length; i += 8) pages.push(`<div class="pg"><div class="cards">${ps.slice(i, i + 8).map(card).join('')}</div></div>`);
  printDoc(pages.join(''));
};
ACT.prDenah = () => {
  const e = S.evDetail.event;
  printDoc(evRooms().map(r => `<div class="pg">${schoolHead()}<h2 style="text-align:center;margin:0">DENAH TEMPAT DUDUK · ${esc(r.nama.toUpperCase())}</h2>
    <p style="text-align:center">${esc(e.nama)} · ${esc(e.tahun_ajaran || '')}</p><p style="text-align:center;border:1px solid #333;padding:4px;margin:10px auto;width:40%">MEJA PENGAWAS</p>
    <div class="seat-grid" style="grid-template-columns:repeat(${r.kolom},1fr)">${Array.from({ length: r.kapasitas }, (_, i) => { const p = r.peserta.find(x => x.kursi === i + 1);
      return `<div class="seat"><b>${i + 1}</b>${p ? `<br>${esc(p.no_peserta)}<br>${esc(p.nama_lengkap)}<br><small>${esc(p.kelas_nama || '')}</small>` : '<br>–'}</div>`; }).join('')}</div></div>`).join(''));
};
ACT.prHadir = () => {
  const e = S.evDetail.event, s = evSched();
  printDoc(evRooms().map(r => `<div class="pg">${schoolHead()}<h3 style="text-align:center;margin:0">DAFTAR HADIR PESERTA</h3>
    <p style="text-align:center">${esc(e.nama)} · ${esc(e.tahun_ajaran || '')}</p>
    <p>Ruang: <b>${esc(r.nama)}</b>${s ? ` · Mapel: <b>${esc(s.mapel)}</b> · ${fmtDate(s.tanggal)} · ${esc(s.jam_mulai)}–${esc(s.jam_selesai)}` : ' · Mapel: ……………… · Tanggal: ………………'}</p>
    <table><thead><tr><th>No</th><th>No. peserta</th><th>Nama</th><th>Kelas</th><th style="width:30%">Tanda tangan</th></tr></thead><tbody>
    ${r.peserta.map((p, i) => `<tr><td>${i + 1}</td><td>${esc(p.no_peserta)}</td><td>${esc(p.nama_lengkap)}</td><td>${esc(p.kelas_nama || '')}</td><td style="text-align:${i % 2 ? 'right' : 'left'}">${i + 1}.</td></tr>`).join('')}</tbody></table>
    <div class="sign"><div></div><div>Pengawas,<br><br><br><br>${esc(r.pengawas || '..............................')}</div></div></div>`).join(''));
};
ACT.prBA = () => {
  const e = S.evDetail.event, s = evSched(), st = S.settings;
  printDoc(evRooms().map(r => `<div class="pg">${schoolHead()}<h3 style="text-align:center;margin:0">BERITA ACARA PELAKSANAAN UJIAN</h3><p style="text-align:center">${esc(e.nama)} · ${esc(e.tahun_ajaran || '')}</p>
    <p style="line-height:2">Pada hari ini ${s ? `<b>${new Date(s.tanggal + 'T00:00').toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}</b>` : '……………… tanggal ……………………'},
    telah diselenggarakan ujian mata pelajaran <b>${s ? esc(s.mapel) : '……………………'}</b> pukul ${s ? esc(s.jam_mulai) + ' – ' + esc(s.jam_selesai) : '……… – ………'}
    di <b>${esc(r.nama)}</b>, ${esc(st.nama_sekolah)}.</p>
    <table><tbody><tr><td style="width:45%">Jumlah peserta seharusnya</td><td>${r.peserta.length} orang</td></tr><tr><td>Jumlah peserta hadir</td><td>……… orang</td></tr>
    <tr><td>Jumlah peserta tidak hadir</td><td>……… orang</td></tr><tr><td>Nomor peserta tidak hadir</td><td style="height:40px"></td></tr></tbody></table>
    <p style="margin-top:12px">Catatan selama pelaksanaan ujian:</p><div style="border:1px solid #555;height:110px"></div>
    <div class="sign"><div>Mengetahui,<br>Kepala Sekolah<br><br><br><br><b><u>${esc(st.kepala_sekolah || '..............................')}</u></b><br>NIP. ${esc(st.nip_kepala || '')}</div>
    <div>Pengawas,<br><br><br><br><br><b><u>${esc(r.pengawas || '..............................')}</u></b></div></div></div>`).join(''));
};

Object.assign(PAGES.admin, { manujian: pageManUjian });
