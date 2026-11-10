/* e-Rapor Kurikulum Merdeka (nilai & capaian, kokurikuler, ekstrakurikuler, ketidakhadiran, catatan wali, cetak/Word)
   dan portofolio karya siswa (internal: siswa, guru mapel, wali kelas, staf). */
addNav('admin', ['rapor', 'e-Rapor', 'file', 'rapor', 'Pembelajaran']);
addNav('admin', ['portofolio', 'Portofolio Siswa', 'image', 'rapor', 'Pembelajaran']);
addNav('guru', ['rapor', 'e-Rapor', 'file', 'rapor', 'Pembelajaran']);
addNav('guru', ['portofolio', 'Portofolio Siswa', 'image', 'rapor', 'Pembelajaran']);
addNav('siswa', ['rapor', 'Rapor', 'file', 'rapor', 'Pembelajaran']);
addNav('siswa', ['portofolio', 'Portofolio Saya', 'image', 'rapor', 'Pembelajaran']);

const RP_KOKU = [['SB', 'Sangat berkembang'], ['BSH', 'Berkembang sesuai harapan'], ['MB', 'Mulai berkembang'], ['BB', 'Belum berkembang']];
const RP_EKSKUL = ['Sangat Baik', 'Baik', 'Cukup', 'Kurang'];
const rpQ = () => qs({ tahun: S.rp?.tahun, smt: S.rp?.smt });

// ---- pilihan tahun ajaran & semester (dipakai semua tab)
async function rpPeriode() {
  if (!S.rpPer) S.rpPer = await GET('/rapor/periode');
  const tahun = Number(S.params.tahun) || S.rpPer.aktif.tahun;
  const smt = Number(S.params.smt) || S.rpPer.aktif.smt;
  S.rp = { tahun, smt };
  const opt = [];
  for (const y of S.rpPer.tahun) for (const s of y.semester) opt.push([`${y.id}:${s.smt}`, `${y.nama} · ${s.nama} (${fmtDate(s.mulai)} – ${fmtDate(s.akhir)})`]);
  return `<select data-change="rpPer" aria-label="Tahun ajaran & semester" style="min-width:0;flex:1 1 280px">${opts(opt, `${tahun}:${smt}`)}</select>`;
}
CHG.rpPer = el => { const [tahun, smt] = el.value.split(':'); go(S.page, { ...S.params, tahun, smt }); };

// =====================================================================
// Guru / admin: e-Rapor
// =====================================================================
async function pageRapor() {
  const per = await rpPeriode();
  const t = await GET('/rapor/tugas?' + rpQ());
  S.rpT = t;
  const items = [];
  if (t.mapel.length) items.push(['mapel', 'Nilai mapel']);
  if (t.kelas.length) items.push(['wali', 'Wali kelas']);
  if (t.ekskul.length || t.staf) items.push(['ekskul', 'Ekstrakurikuler']);
  if (S.user.role === 'admin' || t.staf) items.push(['atur', 'Pengaturan']);
  if (!items.length) {
    return head('e-Rapor', 'Rapor Kurikulum Merdeka per semester.') + `<div class="bar">${per}</div>` +
      `<div class="card">${empty('Anda belum mengampu mapel, menjadi wali kelas, atau membina ekstrakurikuler di tahun ajaran ini.', 'file')}</div>`;
  }
  const tab = items.some(([k]) => k === S.params.tab) ? S.params.tab : items[0][0];
  const body = await ({ mapel: rpMapel, wali: rpWali, ekskul: rpEkskul, atur: rpAtur })[tab]();
  return head('e-Rapor', `Rapor Kurikulum Merdeka · ${esc(t.periode)}`) + `<div class="bar">${per}</div>` + (items.length > 1 ? tabs(items, tab) : '') + body;
}

// ---- Nilai & capaian kompetensi per mapel
async function rpMapel() {
  const ms = S.rpT.mapel;
  const cid = ms.some(m => m.id === Number(S.params.course)) ? Number(S.params.course) : ms[0].id;
  const d = await GET(`/rapor/mapel?course_id=${cid}&` + rpQ());
  S.rpM = { cid, d };
  const kunci = d.terbit;
  const val = s => s.nilai != null ? s.nilai : (s.otomatis?.akhir != null ? Math.round(s.otomatis.akhir) : '');
  const rows = d.siswa.map((s, i) => `<tr data-sid="${s.id}"><td class="n">${i + 1}</td>
      <td><b>${esc(s.nama_lengkap)}</b><small class="muted">${esc(s.username)}${s.portofolio ? ` · ${s.portofolio} karya portofolio` : ''}</small></td>
      <td class="n" title="Tugas ${fmtN(s.otomatis?.tugas)} · Kuis ${fmtN(s.otomatis?.kuis)} · Hadir ${fmtN(s.otomatis?.hadir)}%">${fmtN(s.otomatis?.akhir)}</td>
      <td><input type="number" min="0" max="100" step="1" name="nilai" value="${esc(val(s))}" aria-label="Nilai rapor ${esc(s.nama_lengkap)}" style="width:80px" ${kunci ? 'disabled' : ''}>${s.tersimpan ? '' : ' <small class="muted">belum disimpan</small>'}</td>
      <td><textarea name="tinggi" rows="2" aria-label="Capaian tertinggi" ${kunci ? 'disabled' : ''} placeholder="${esc(s.saran_tinggi || 'Menunjukkan penguasaan yang baik dalam …')}">${esc(s.tinggi || '')}</textarea></td>
      <td><textarea name="perlu" rows="2" aria-label="Perlu ditingkatkan" ${kunci ? 'disabled' : ''} placeholder="${esc(s.saran_perlu || 'Perlu bimbingan dalam …')}">${esc(s.perlu || '')}</textarea></td></tr>`).join('');
  return `<div class="bar"><select data-change="rpCourse" aria-label="Mata pelajaran" style="min-width:0;flex:1 1 220px">${opts(ms.map(m => [m.id, `${m.nama} · ${m.kelas_nama}`]), cid)}</select>
      <div class="acts" style="margin-left:auto">${kunci ? '' : btn('Isi dari nilai otomatis', 'rpAuto', {}, 'ghost sm', 'check') + (modOn('ai') ? btn('Saran deskripsi AI', 'rpAI', {}, 'ghost sm', 'spark') : '') + btn('Simpan semua', 'rpMapelSave', {}, 'sm', 'check')}</div></div>
    ${kunci ? `<div class="hint warn" style="margin-bottom:12px">Rapor kelas ${esc(d.kelas)} sudah diterbitkan wali kelas, jadi isinya dikunci. Minta wali kelas membatalkan terbit untuk mengubah.</div>` : ''}
    <p class="muted" style="margin-bottom:10px;font-size:13px">Nilai otomatis = 40% tugas + 40% kuis + 20% kehadiran dari ${d.jumlah_tugas} tugas dan ${d.jumlah_kuis} kuis semester ini (${fmtDate(d.mulai)} – ${fmtDate(d.akhir)}). Nilai rapor boleh disesuaikan; deskripsi kosong memakai saran di kotak (abu-abu) setelah menekan "Isi dari nilai otomatis".</p>
    ${d.siswa.length ? `<div class="tw rp-tabel" id="rpMapelBox"><table><thead><tr><th class="n">No</th><th>Siswa</th><th class="n">Otomatis</th><th>Nilai rapor</th><th>Capaian tertinggi</th><th>Perlu ditingkatkan</th></tr></thead><tbody>${rows}</tbody></table></div>`
    : `<div class="card">${empty('Belum ada siswa di kelas ini.', 'users')}</div>`}`;
}
CHG.rpCourse = el => go('rapor', { ...S.params, tab: 'mapel', course: el.value });
ACT.rpAuto = () => {
  let n = 0;
  for (const s of S.rpM.d.siswa) {
    const tr = $(`#rpMapelBox tr[data-sid="${s.id}"]`);
    if (!tr) continue;
    const ni = $('[name=nilai]', tr), ti = $('[name=tinggi]', tr), pe = $('[name=perlu]', tr);
    if (ni.value === '' && s.otomatis?.akhir != null) { ni.value = Math.round(s.otomatis.akhir); n++; }
    if (!ti.value && s.saran_tinggi) { ti.value = s.saran_tinggi; n++; }
    if (!pe.value && s.saran_perlu) { pe.value = s.saran_perlu; n++; }
  }
  toast(n ? 'Isian kosong diisi dari nilai. Periksa, lalu tekan Simpan semua.' : 'Tidak ada isian kosong yang bisa diisi otomatis.');
};
ACT.rpAI = async el => {
  if (!(await confirmBox('Minta Asisten AI menulis saran deskripsi capaian untuk satu kelas? Memakai 1 kuota AI hari ini. Nama siswa tidak dikirim ke AI; saran hanya mengisi kotak yang masih kosong.', 'Minta saran', ''))) return;
  el.disabled = true; const label = el.innerHTML; el.innerHTML = 'Meminta saran…';
  try {
    const r = await POST('/rapor/mapel/ai', { course_id: S.rpM.cid, tahun: S.rp.tahun, smt: S.rp.smt });
    let n = 0;
    for (const x of r.saran) {
      const tr = $(`#rpMapelBox tr[data-sid="${x.siswa_id}"]`);
      if (!tr) continue;
      const ti = $('[name=tinggi]', tr), pe = $('[name=perlu]', tr);
      if (!ti.value && x.tinggi) { ti.value = x.tinggi; n++; }
      if (!pe.value && x.perlu) { pe.value = x.perlu; n++; }
    }
    toast(n ? `${n} deskripsi diisi dari saran AI. Periksa, lalu tekan Simpan semua.` : 'Tidak ada kotak kosong yang diisi.');
  } finally { el.disabled = false; el.innerHTML = label; }
};
ACT.rpMapelSave = async () => {
  const items = $$('#rpMapelBox tr[data-sid]').map(tr => {
    const v = $('[name=nilai]', tr).value;
    return { siswa_id: Number(tr.dataset.sid), nilai: v === '' ? null : Number(v), tinggi: $('[name=tinggi]', tr).value, perlu: $('[name=perlu]', tr).value };
  });
  if (items.some(x => x.nilai != null && (x.nilai < 0 || x.nilai > 100))) return toast('Nilai rapor harus 0 sampai 100.', 'bad');
  await PUT('/rapor/mapel', { course_id: S.rpM.cid, tahun: S.rp.tahun, smt: S.rp.smt, items });
  toast('Nilai rapor tersimpan.');
  await refresh(true);
};

// ---- Wali kelas: kokurikuler, ketidakhadiran, catatan, terbit, cetak
async function rpWali() {
  const ks = S.rpT.kelas;
  const kid = ks.some(k => k.id === Number(S.params.kelas)) ? Number(S.params.kelas) : ks[0].id;
  const d = await GET(`/rapor/kelas?kelas_id=${kid}&` + rpQ());
  S.rpW = { kid, d };
  const total = d.siswa.length;
  const kurang = d.mapel.filter(m => (m.terisi || 0) < total);
  const genap = d.smt === 2;
  const kokuRingkas = r => { const k = r?.kokurikuler || {}; const n = Object.keys(k).length; return n ? `${n}/8 dimensi` : '<span class="muted">belum</span>'; };
  const absen = s => { const a = s.rapor || {}, o = s.absen_otomatis; return `S ${a.sakit ?? o.sakit} · I ${a.izin ?? o.izin} · A ${a.alfa ?? o.alfa}`; };
  return `<div class="bar"><select data-change="rpKelas" aria-label="Kelas" style="min-width:0;flex:0 1 200px">${opts(ks.map(k => [k.id, 'Kelas ' + k.nama]), kid)}</select>
      <div class="acts" style="margin-left:auto">${btn('Cetak semua', 'rpPrintKelas', {}, 'ghost sm', 'print')}
      ${d.terbit ? btn('Batalkan terbit', 'rpTerbit', { on: 0 }, 'danger sm') : btn('Terbitkan rapor', 'rpTerbit', { on: 1 }, 'sm', 'check')}</div></div>
    <section class="card" style="margin-bottom:14px"><div class="card-head"><h3>Kelengkapan nilai mapel</h3>${d.terbit ? tag('Terbit ' + fmtDate(d.terbit), 'ok') : tag(kurang.length ? 'Belum lengkap' : 'Siap diterbitkan', kurang.length ? 'warn' : 'ok')}</div>
      <div style="display:flex;flex-wrap:wrap;gap:6px">${d.mapel.map(m => tag(`${m.nama}: ${m.terisi || 0}/${total}`, (m.terisi || 0) >= total ? 'ok' : 'warn')).join('') || '<span class="muted">Belum ada mapel di kelas ini.</span>'}</div>
      ${d.terbit ? '<p class="muted" style="margin-top:8px;font-size:13px">Rapor sudah bisa dilihat siswa. Isi rapor dikunci sampai terbit dibatalkan.</p>' : ''}</section>
    ${total ? `<div class="tw"><table><thead><tr><th class="n">No</th><th>Siswa</th><th>Kokurikuler</th><th>Ketidakhadiran</th><th>Catatan wali</th>${genap ? '<th>Keputusan</th>' : ''}<th></th></tr></thead><tbody>
      ${d.siswa.map((s, i) => `<tr><td class="n">${i + 1}</td><td><b>${esc(s.nama_lengkap)}</b><small class="muted">${esc(s.username)}${s.nisn ? ' · NISN ' + esc(s.nisn) : ''}</small></td>
        <td>${kokuRingkas(s.rapor)}</td><td>${absen(s)}</td><td>${s.rapor?.catatan ? esc(s.rapor.catatan.slice(0, 60)) + (s.rapor.catatan.length > 60 ? '…' : '') : '<span class="muted">belum</span>'}</td>
        ${genap ? `<td>${esc({ naik: 'Naik', tinggal: 'Tinggal', lulus: 'Lulus' }[s.rapor?.keputusan] || '–')}</td>` : ''}
        <td><div class="acts">${d.terbit ? '' : btn('Isi', 'rpSiswaForm', { id: s.id }, 'ghost sm', 'edit')}${btn('Lihat', 'rpLihat', { id: s.id }, 'ghost sm', 'eye')}</div></td></tr>`).join('')}
    </tbody></table></div>` : `<div class="card">${empty('Belum ada siswa di kelas ini.', 'users')}</div>`}`;
}
CHG.rpKelas = el => go('rapor', { ...S.params, tab: 'wali', kelas: el.value });

ACT.rpSiswaForm = el => {
  const d = S.rpW.d, s = d.siswa.find(x => x.id === Number(el.dataset.id));
  const r = s.rapor || {}, k = r.kokurikuler || {}, o = s.absen_otomatis;
  const fields = [
    { t: 'html', html: '<p class="muted" style="font-size:13px;margin-bottom:6px">Kokurikuler: capaian 8 dimensi profil lulusan selama semester ini.</p>' },
    ...d.dimensi.map(x => ({ n: 'k_' + x.Key, l: x.Nama, t: 'select', o: RP_KOKU, blank: '—' })),
    { n: 'koku_desk', l: 'Deskripsi kokurikuler', t: 'textarea', rows: 3, ph: 'Kosongkan untuk disusun otomatis dari pilihan di atas' },
    [{ n: 'sakit', l: 'Sakit (hari)', t: 'number', ph: String(o.sakit) }, { n: 'izin', l: 'Izin (hari)', t: 'number', ph: String(o.izin) }, { n: 'alfa', l: 'Tanpa keterangan', t: 'number', ph: String(o.alfa) }],
    { t: 'html', html: '<small class="muted" style="display:block;margin:-4px 0 8px">Kosongkan ketidakhadiran untuk memakai hitungan absensi harian (angka abu-abu).</small>' },
    { n: 'catatan', l: 'Catatan wali kelas', t: 'textarea', rows: 3 },
    ...(d.smt === 2 ? [{ n: 'keputusan', l: 'Keputusan akhir tahun', t: 'select', o: [['naik', 'Naik kelas'], ['tinggal', 'Tinggal di kelas'], ['lulus', 'Lulus']], blank: '—' }] : []),
  ];
  const values = { koku_desk: r.koku_desk || '', catatan: r.catatan || '', keputusan: r.keputusan || '', sakit: r.sakit ?? '', izin: r.izin ?? '', alfa: r.alfa ?? '' };
  for (const x of d.dimensi) values['k_' + x.Key] = k[x.Key] || '';
  openForm({ title: 'Rapor ' + s.nama_lengkap, fields, values, wide: true, onSubmit: v => {
    const koku = {};
    for (const x of d.dimensi) if (v['k_' + x.Key]) koku[x.Key] = v['k_' + x.Key];
    return PUT('/rapor/siswa/' + s.id, { tahun: d.tahun, smt: d.smt, kokurikuler: koku, koku_desk: v.koku_desk, catatan: v.catatan, keputusan: v.keputusan || '',
      sakit: v.sakit === '' || v.sakit == null ? null : v.sakit, izin: v.izin === '' || v.izin == null ? null : v.izin, alfa: v.alfa === '' || v.alfa == null ? null : v.alfa });
  } });
};

ACT.rpTerbit = async el => {
  const on = el.dataset.on === '1';
  if (!(await confirmBox(on ? 'Terbitkan rapor kelas ini? Siswa bisa melihat rapornya, dan isi rapor dikunci.' : 'Batalkan terbit? Rapor tidak terlihat siswa lagi dan isinya bisa diubah.', on ? 'Terbitkan' : 'Batalkan terbit', on ? '' : 'danger'))) return;
  await PUT(`/rapor/kelas/${S.rpW.kid}/terbit`, { tahun: S.rp.tahun, smt: S.rp.smt, terbit: on });
  toast(on ? 'Rapor diterbitkan.' : 'Terbit dibatalkan.');
  await refresh(true);
};

// ---- Lihat, cetak, Word
const rpNilai = v => v == null ? '–' : fmtN(v);
function raporHtml(r) {
  const s = r.siswa, k = r.sekolah;
  const kep = r.genap && r.keputusan ? `<h4>F. Keputusan</h4><p>${r.keputusan === 'naik' ? `Berdasarkan pencapaian seluruh kompetensi, peserta didik dinyatakan <b>naik ke kelas ${esc(r.tingkat_berikut || '')}</b>.`
    : r.keputusan === 'tinggal' ? `Berdasarkan pencapaian seluruh kompetensi, peserta didik dinyatakan <b>tinggal di kelas ${esc(r.kelas)}</b>.` : 'Peserta didik dinyatakan <b>LULUS</b>.'}</p>` : '';
  return `<div class="rapor">${schoolHead()}<h3 style="text-align:center;margin:6px 0 10px">LAPORAN HASIL BELAJAR</h3>
    <table class="rp-id"><tr><td>Nama</td><td><b>${esc(s.nama_lengkap)}</b></td><td>Kelas</td><td>${esc(r.kelas)}</td></tr>
      <tr><td>NIS / NISN</td><td>${esc(s.username)} / ${esc(s.nisn || '–')}</td><td>Fase</td><td>${esc(r.fase || '–')}</td></tr>
      <tr><td>Sekolah</td><td>${esc(k.nama || '')}</td><td>Semester</td><td>${esc(r.semester)}</td></tr>
      <tr><td>Alamat</td><td>${esc(k.alamat || '')}</td><td>Tahun ajaran</td><td>${esc(r.tahun_ajaran)}</td></tr></table>
    <h4>A. Nilai dan Capaian Kompetensi</h4>
    <table class="rp-m"><thead><tr><th class="n">No</th><th>Mata Pelajaran</th><th class="n">Nilai Akhir</th><th>Capaian Kompetensi</th></tr></thead><tbody>
      ${r.mapel.map((m, i) => `<tr><td class="n">${i + 1}</td><td>${esc(m.mapel)}</td><td class="n">${rpNilai(m.nilai)}</td><td>${esc([m.tinggi, m.perlu].filter(Boolean).join(' '))}</td></tr>`).join('')}</tbody></table>
    <h4>B. Kokurikuler</h4><p>${esc(r.koku_desk || '–')}</p>
    <h4>C. Ekstrakurikuler</h4>
    ${r.ekskul.length ? `<table class="rp-m rp-ek"><thead><tr><th class="n">No</th><th>Kegiatan</th><th>Predikat</th><th>Keterangan</th></tr></thead><tbody>
      ${r.ekskul.map((e, i) => `<tr><td class="n">${i + 1}</td><td>${esc(e.nama)}</td><td>${esc(e.predikat || '–')}</td><td>${esc(e.keterangan || '')}</td></tr>`).join('')}</tbody></table>`
      : '<p>Tidak mengikuti ekstrakurikuler.</p>'}
    <div class="rp-dua"><div><h4>D. Ketidakhadiran</h4><table><tr><td>Sakit</td><td class="n">${r.ketidakhadiran.sakit} hari</td></tr><tr><td>Izin</td><td class="n">${r.ketidakhadiran.izin} hari</td></tr><tr><td>Tanpa keterangan</td><td class="n">${r.ketidakhadiran.alfa} hari</td></tr></table></div>
      <div><h4>E. Catatan Wali Kelas</h4><p>${esc(r.catatan || '–')}</p></div></div>
    ${kep}
    <div class="rp-ttd"><div>Orang Tua/Wali Murid<br><br><br><br>.............................</div>
      <div>${fmtDate(r.terbit || today())}<br>Wali Kelas<br><br><br><br><b><u>${esc(r.wali.nama || '.............................')}</u></b><br>NIP. ${esc(r.wali.nip || '')}</div></div>
    <div class="rp-ttd satu"><div>Mengetahui,<br>Kepala Sekolah<br><br><br><br><b><u>${esc(k.kepala || '.............................')}</u></b><br>NIP. ${esc(k.nip_kepala || '')}</div></div></div>`;
}
const rpGetRapor = id => GET(`/rapor/siswa/${id}?` + rpQ());
ACT.rpLihat = async el => {
  const id = el.dataset.id;
  const r = await rpGetRapor(id);
  openModal({ title: 'Rapor ' + r.siswa.nama_lengkap, wide: true, body: `<div class="rp-prev">${raporHtml(r)}</div>
    <div class="sheet-foot">${btn('Unduh Word', 'rpWord', { id }, 'ghost', 'download')}${btn('Cetak', 'rpPrint', { id }, '', 'print')}</div>` });
};
ACT.rpPrint = async el => printDoc(raporHtml(await rpGetRapor(el.dataset.id)));
ACT.rpWord = el => { const a = Object.assign(document.createElement('a'), { href: `/api/rapor/siswa/${el.dataset.id}/docx?` + rpQ(), download: '' }); document.body.append(a); a.click(); a.remove(); };
ACT.rpPrintKelas = async el => {
  const ids = S.rpW.d.siswa.map(s => s.id);
  if (!ids.length) return;
  el.disabled = true;
  try {
    const all = [];
    for (const id of ids) all.push(await rpGetRapor(id)); // berurutan: tidak membebani server
    printDoc(all.map(r => `<div class="pg">${raporHtml(r)}</div>`).join(''));
  } finally { el.disabled = false; }
};

// ---- Ekstrakurikuler: anggota & predikat (pembina), daftar (admin)
async function rpEkskul() {
  const es = await GET('/ekskul?' + rpQ());
  const mine = S.rpT.staf ? es : es.filter(e => S.rpT.ekskul.some(x => x.id === e.id));
  if (!mine.length) return `<div class="card">${empty(S.rpT.staf ? 'Belum ada ekstrakurikuler. Tambahkan di tab Pengaturan.' : 'Anda belum menjadi pembina ekstrakurikuler.', 'users')}</div>`;
  const eid = mine.some(e => e.id === Number(S.params.ekskul)) ? Number(S.params.ekskul) : mine[0].id;
  const d = await GET(`/ekskul/${eid}/anggota?` + rpQ());
  S.rpE = { eid, d };
  return `<div class="bar"><select data-change="rpEk" aria-label="Ekstrakurikuler" style="min-width:0;flex:0 1 240px">${opts(mine.map(e => [e.id, `${e.nama}${e.pembina ? ' · ' + e.pembina : ''}`]), eid)}</select>
      <div class="acts" style="margin-left:auto">${btn('Atur anggota', 'rpAnggota', {}, 'ghost sm', 'users')}${d.anggota.length ? btn('Simpan predikat', 'rpEkSave', {}, 'sm', 'check') : ''}</div></div>
    ${d.anggota.length ? `<div class="tw" id="rpEkBox"><table><thead><tr><th class="n">No</th><th>Siswa</th><th>Kelas</th><th>Predikat</th><th>Keterangan</th></tr></thead><tbody>
      ${d.anggota.map((s, i) => `<tr data-sid="${s.id}"><td class="n">${i + 1}</td><td><b>${esc(s.nama_lengkap)}</b></td><td>${esc(s.kelas || '–')}</td>
        <td><select name="predikat" aria-label="Predikat">${opts(RP_EKSKUL.map(p => [p, p]), s.predikat || '', '—')}</select></td>
        <td><input name="keterangan" value="${esc(s.keterangan || '')}" placeholder="mis. Aktif dalam kegiatan perkemahan" aria-label="Keterangan"></td></tr>`).join('')}</tbody></table></div>`
    : `<div class="card">${empty('Belum ada anggota di tahun ajaran ini. Tekan "Atur anggota".', 'users')}</div>`}`;
}
CHG.rpEk = el => go('rapor', { ...S.params, tab: 'ekskul', ekskul: el.value });
ACT.rpEkSave = async () => {
  const items = $$('#rpEkBox tr[data-sid]').map(tr => ({ siswa_id: Number(tr.dataset.sid), predikat: $('[name=predikat]', tr).value, keterangan: $('[name=keterangan]', tr).value }));
  await PUT(`/ekskul/${S.rpE.eid}/nilai`, { tahun: S.rp.tahun, smt: S.rp.smt, items });
  toast('Predikat ekstrakurikuler tersimpan.');
  await refresh(true);
};
ACT.rpAnggota = async () => {
  const calon = await GET(`/ekskul/${S.rpE.eid}/calon?` + rpQ());
  const ada = new Set(S.rpE.d.anggota.map(a => String(a.id)));
  openForm({ title: 'Anggota ekstrakurikuler', wide: true,
    fields: [{ n: 'siswa', l: 'Pilih siswa anggota (tahun ajaran ini)', t: 'checks', o: calon.map(s => [s.id, `${s.nama_lengkap} · ${s.kelas}`]) }],
    values: { siswa: [...ada] },
    onSubmit: v => PUT(`/ekskul/${S.rpE.eid}/anggota`, { tahun: S.rp.tahun, siswa_ids: (v.siswa || []).map(Number) }) });
};

// ---- Pengaturan: batas semester & daftar ekstrakurikuler
async function rpAtur() {
  const [es, gs] = await Promise.all([GET('/ekskul?' + rpQ()), GET('/rapor/guru').catch(() => [])]);
  S.rpGuru = gs;
  S.rpEks = es;
  return `<div class="cols"><section class="card"><div class="card-head"><h3>Semester</h3></div>
      <p class="muted" style="font-size:13px;margin-bottom:10px">Semester ganjil berakhir pada tanggal di bawah; genap mulai keesokan harinya sampai akhir tahun ajaran. Nilai otomatis, absensi, dan portofolio rapor dihitung dari rentang ini.</p>
      ${S.rpPer.tahun.map(y => `<div class="li"><div><b>${esc(y.nama)}</b><small class="muted">Ganjil ${fmtDate(y.semester[0].mulai)} – ${fmtDate(y.semester[0].akhir)} · Genap ${fmtDate(y.semester[1].mulai)} – ${fmtDate(y.semester[1].akhir)}</small></div>
        ${btn('Ubah', 'rpSmtForm', { id: y.id }, 'ghost sm', 'edit')}</div>`).join('')}</section>
    <section class="card"><div class="card-head"><h3>Ekstrakurikuler</h3>${btn('Tambah', 'rpEkForm', {}, 'sm', 'plus')}</div>
      ${es.length ? es.map(e => `<div class="li"><div><b>${esc(e.nama)}</b><small class="muted">Pembina: ${esc(e.pembina || 'belum ditentukan')} · ${e.anggota} anggota</small></div>
        <div class="acts">${iconBtn('edit', 'rpEkForm', { id: e.id }, 'Ubah')}${iconBtn('trash', 'rpEkDel', { id: e.id, nama: e.nama }, 'Hapus', 'danger')}</div></div>`).join('') : empty('Belum ada ekstrakurikuler. Minimal Pramuka (wajib).', 'users')}</section></div>`;
}
ACT.rpSmtForm = el => {
  const y = S.rpPer.tahun.find(x => x.id === Number(el.dataset.id));
  openForm({ title: 'Semester ' + y.nama, fields: [{ n: 'akhir_ganjil', l: 'Tanggal terakhir semester ganjil', t: 'date', hint: 'Kosongkan untuk memakai 31 Desember.' }],
    values: { akhir_ganjil: y.akhir_ganjil || '' }, onSubmit: async v => { await PUT('/rapor/periode/' + y.id, { akhir_ganjil: v.akhir_ganjil || null }); S.rpPer = null; } });
};
ACT.rpEkForm = el => {
  const e = S.rpEks.find(x => x.id === Number(el.dataset.id)) || {};
  openForm({ title: e.id ? 'Ubah ekstrakurikuler' : 'Ekstrakurikuler baru', values: { nama: e.nama || '', pembina_id: e.pembina_id || '' },
    fields: [{ n: 'nama', l: 'Nama kegiatan', req: true, ph: 'Pramuka' }, { n: 'pembina_id', l: 'Pembina', t: 'select', o: S.rpGuru.map(g => [g.id, g.nama_lengkap]), blank: '—' }],
    onSubmit: v => { const body = { nama: v.nama, pembina_id: v.pembina_id ? Number(v.pembina_id) : null }; return e.id ? PUT('/ekskul/' + e.id, body) : POST('/ekskul', body); } });
};
ACT.rpEkDel = el => del('/ekskul/' + el.dataset.id, `Hapus ekstrakurikuler <b>${esc(el.dataset.nama)}</b> beserta daftar anggotanya?`);

// =====================================================================
// Siswa: rapor
// =====================================================================
async function siswaRapor() {
  const me = await GET('/rapor/me');
  if (!me.terbit.length) return head('Rapor', 'Rapor muncul di sini setelah diterbitkan wali kelas.') + `<div class="card">${empty('Belum ada rapor yang diterbitkan.', 'file')}</div>`;
  const key = S.params.p || `${me.terbit[0].tahun_id}:${me.terbit[0].semester}`;
  const [tahun, smt] = key.split(':').map(Number);
  S.rp = { tahun, smt };
  const r = await GET(`/rapor/siswa/${S.user.id}?` + rpQ());
  const opt = me.terbit.map(t => [`${t.tahun_id}:${t.semester}`, `${t.nama} · ${t.semester === 1 ? 'Ganjil' : 'Genap'}`]);
  return head('Rapor', 'Laporan hasil belajar per semester.', btn('Unduh Word', 'rpWord', { id: S.user.id }, 'ghost', 'download') + btn('Cetak', 'rpPrint', { id: S.user.id }, '', 'print')) +
    `<div class="bar"><select data-change="rpSiswaPer" aria-label="Semester" style="min-width:0;flex:0 1 280px">${opts(opt, key)}</select></div>
    <section class="card rp-prev">${raporHtml(r)}</section>`;
}
CHG.rpSiswaPer = el => go('rapor', { p: el.value });

// =====================================================================
// Portofolio
// =====================================================================
const portoCard = (p, guru) => `<section class="card porto">
  <div class="card-head"><div><h3>${esc(p.judul)}</h3><small class="muted">${guru ? `${esc(p.siswa)} · ${esc(p.kelas || '')} · ` : ''}${esc(p.mapel || 'Umum')} · ${fmtDate(p.created_at)}</small></div>
    ${p.nilai != null ? tag('Nilai ' + fmtN(p.nilai), 'ok') : p.catatan_guru ? tag('Dicatat guru', 'info') : tag('Belum dinilai')}</div>
  ${p.file_url ? (/\.(png|jpe?g|webp|gif)$/i.test(p.file_url) ? `<a href="${esc(p.file_url)}" target="_blank" rel="noopener"><img class="porto-img" src="${esc(p.file_url)}" alt="${esc(p.judul)}" loading="lazy"></a>`
    : `<p><a href="${esc(p.file_url)}" target="_blank" rel="noopener">${ico('file')} Buka berkas</a></p>`) : ''}
  ${p.deskripsi ? `<p style="white-space:pre-wrap">${esc(p.deskripsi)}</p>` : ''}
  ${p.link ? `<p><a href="${esc(p.link)}" target="_blank" rel="noopener noreferrer">${ico('link')} ${esc(p.link.slice(0, 60))}</a></p>` : ''}
  ${p.catatan_guru ? `<div class="hint" style="margin-top:8px"><b>Catatan ${esc(p.dinilai_oleh || 'guru')}:</b> ${esc(p.catatan_guru)}</div>` : ''}
  <div class="acts" style="margin-top:10px">${guru ? btn(p.dinilai_at ? 'Ubah catatan' : 'Beri catatan', 'portoNilai', { id: p.id }, 'ghost sm', 'edit')
    : p.nilai == null ? btn('Ubah', 'portoForm', { id: p.id }, 'ghost sm', 'edit') + iconBtn('trash', 'portoDel', { id: p.id, judul: p.judul }, 'Hapus', 'danger') : ''}</div></section>`;

async function siswaPorto() {
  const [list, courses] = await Promise.all([GET('/portofolio'), GET('/courses').catch(() => [])]);
  S.porto = list;
  S.portoMapel = courses;
  return head('Portofolio Saya', 'Kumpulkan karya terbaikmu: poster, laporan, foto proyek, atau tautan video. Guru mapel & wali kelas bisa memberi catatan.', btn('Tambah karya', 'portoForm', {}, '', 'plus')) +
    (list.length ? `<div class="grid">${list.map(p => portoCard(p, false)).join('')}</div>` : `<div class="card">${empty('Belum ada karya. Tekan "Tambah karya".', 'image')}</div>`);
}
ACT.portoForm = el => {
  const p = S.porto.find(x => x.id === Number(el.dataset.id)) || {};
  S.portoFile = '';
  const fileHtml = `<div class="f"><span style="font-size:13px;font-weight:600;color:var(--text-2)">Berkas karya <small>(gambar atau PDF, opsional)</small></span>
    <label class="btn ghost sm" style="width:fit-content">${ico('upload')} Pilih berkas<input type="file" accept="image/*,application/pdf" data-change="portoFile" hidden></label>
    <small class="muted" id="portoFileInfo">${p.file_url ? 'Berkas lama tetap dipakai bila tidak memilih yang baru.' : 'Foto diperkecil otomatis; PDF maks 2 MB.'}</small></div>`;
  openForm({ title: p.id ? 'Ubah karya' : 'Tambah karya', wide: true, values: { judul: p.judul || '', course_id: p.course_id || '', deskripsi: p.deskripsi || '', link: p.link || '' },
    fields: [{ n: 'judul', l: 'Judul karya', req: true, ph: 'Poster Garis Bilangan' },
      { n: 'course_id', l: 'Mata pelajaran', t: 'select', o: S.portoMapel.map(c => [c.id, c.nama]), blank: 'Umum (tanpa mapel)' },
      { n: 'deskripsi', l: 'Deskripsi', t: 'textarea', rows: 3, ph: 'Ceritakan karyamu: tujuan, proses, dan apa yang kamu pelajari.' },
      { n: 'link', l: 'Tautan (opsional)', ph: 'https://…' }, { t: 'html', html: fileHtml }],
    onSubmit: v => { const body = { judul: v.judul, course_id: v.course_id ? Number(v.course_id) : null, deskripsi: v.deskripsi, link: v.link, file: S.portoFile || '' };
      return p.id ? PUT('/portofolio/' + p.id, body) : POST('/portofolio', body); } });
};
CHG.portoFile = async inp => {
  const f = inp.files[0];
  if (!f) return;
  const info = $('#portoFileInfo');
  info.textContent = 'Menyiapkan berkas…';
  try { S.portoFile = await fileToDataURL(f, 1600); info.textContent = 'Siap diunggah: ' + f.name; } catch (e) { S.portoFile = ''; info.textContent = ''; toast(e.message, 'bad'); }
};
ACT.portoDel = el => del('/portofolio/' + el.dataset.id, `Hapus karya <b>${esc(el.dataset.judul)}</b>?`);

async function guruPorto() {
  const list = await GET('/portofolio' + (S.params.course ? '?course_id=' + Number(S.params.course) : ''));
  S.porto = list;
  const mapel = [...new Map(list.filter(p => p.course_id).map(p => [p.course_id, p.mapel])).entries()];
  return head('Portofolio Siswa', 'Karya siswa di mapel yang Anda ampu dan kelas perwalian Anda.') +
    (mapel.length > 1 || S.params.course ? `<div class="bar"><select data-change="portoCourse" aria-label="Mata pelajaran" style="min-width:0;flex:0 1 260px">${opts(mapel, S.params.course || '', 'Semua mapel')}</select></div>` : '') +
    (list.length ? `<div class="grid">${list.map(p => portoCard(p, true)).join('')}</div>` : `<div class="card">${empty('Belum ada karya siswa.', 'image')}</div>`);
}
CHG.portoCourse = el => go('portofolio', el.value ? { course: el.value } : {});
ACT.portoNilai = el => {
  const p = S.porto.find(x => x.id === Number(el.dataset.id));
  openForm({ title: 'Catatan untuk ' + p.siswa, values: { catatan: p.catatan_guru || '', nilai: p.nilai ?? '' },
    fields: [{ n: 'catatan', l: 'Catatan / umpan balik', t: 'textarea', rows: 4, req: true }, { n: 'nilai', l: 'Nilai (0–100, opsional)', t: 'number' }],
    onSubmit: v => PUT(`/portofolio/${p.id}/catatan`, { catatan: v.catatan, nilai: v.nilai === '' || v.nilai == null ? null : Number(v.nilai) }) });
};

Object.assign(PAGES.admin, { rapor: pageRapor, portofolio: guruPorto });
Object.assign(PAGES.guru, { rapor: pageRapor, portofolio: guruPorto });
Object.assign(PAGES.siswa, { rapor: siswaRapor, portofolio: siswaPorto });
