'use strict';
/* Ujian CBT: kelola ujian & soal (guru/admin) dan mengerjakan ujian (siswa) */

addNav('admin', ['cbt', 'Ujian CBT', 'quiz', 'ujian', 'Ujian']);
addNav('guru', ['cbt', 'Ujian CBT', 'quiz', 'ujian', 'Ujian']);
addNav('siswa', ['cbt', 'Ujian CBT', 'quiz', 'ujian', 'Ujian']);
PAGE_PARENT.cbtujian = 'cbt';

// =====================================================================
// CBT — guru & admin
// =====================================================================
async function pageCBT() {
  if (S.user.role === 'siswa') return siswaCBT();
  const [xs, ks] = await Promise.all([GET('/cbt'), GET('/classes')]);
  S.cbtList = Object.fromEntries(xs.map(x => [x.id, x]));
  S.cbtKelas = ks;
  const kn = ids => JSON.parse(ids || '[]').map(i => ks.find(k => k.id === i)?.nama).filter(Boolean).join(', ');
  const now = new Date(nowMs());
  return head('Ujian CBT', 'Ujian berbasis komputer: soal bisa bergambar, urutan soal & pilihan diacak, ditutup otomatis sesuai jadwal.', btn('Buat ujian', 'cbtForm', {}, '', 'plus')) +
    (xs.length ? `<div class="tw"><table><thead><tr><th>Ujian</th><th>Kelas</th><th>Jadwal</th><th class="n">Soal</th><th class="n">Selesai</th><th>Token</th><th></th></tr></thead><tbody>
    ${xs.map(x => { const st = now < new Date(x.mulai) ? ['Terjadwal', 'info'] : now > new Date(x.selesai) ? ['Ditutup', ''] : ['Berlangsung', 'ok'];
      return `<tr><td><b>${esc(x.judul)}</b><small class="muted">${esc(x.mapel || '')} · ${x.durasi_menit} menit${S.user.role === 'admin' && x.pembuat ? ' · ' + esc(x.pembuat) : ''}</small></td>
      <td>${esc(kn(x.kelas_ids))}</td><td><small>${fmtDT(x.mulai)}<br>s/d ${fmtDT(x.selesai)}</small> ${tag(st[0], st[1])}</td>
      <td class="n">${x.jumlah_soal || tag('0', 'warn')}</td><td class="n">${x.jumlah_selesai}</td><td>${x.token ? `<code>${esc(x.token)}</code>` : '<small class="muted">tanpa</small>'}</td>
      <td><div class="acts">${btn('Soal', 'cbtSoal', { id: x.id }, 'sm')}${btn('Hasil', 'cbtHasil', { id: x.id }, 'ghost sm')}${btn('LJK', 'ljkOpen', { id: x.id }, 'ghost sm', 'print')}${iconBtn('edit', 'cbtForm', { id: x.id }, 'Ubah')}${iconBtn('trash', 'cbtDel', { id: x.id, judul: x.judul }, 'Hapus', 'danger')}</div></td></tr>`; }).join('')}
    </tbody></table></div>` : `<div class="card">${empty('Belum ada ujian CBT.', 'quiz')}</div>`);
}
const randToken = () => Array.from(crypto.getRandomValues(new Uint8Array(5)), b => 'ABCDEFGHJKLMNPQRSTUVWXYZ23456789'[b % 32]).join('');
ACT.cbtForm = async el => {
  const mapel = await mapelOpts();
  const x = el.dataset.id ? { ...S.cbtList[el.dataset.id], kelas_ids: JSON.parse(S.cbtList[el.dataset.id].kelas_ids) }
    : { durasi_menit: 90, jumlah_opsi: 4, acak_soal: 1, acak_opsi: 1, tampil_nilai: 1, token: randToken(), kelas_ids: [] };
  x.acak_soal = !!x.acak_soal; x.acak_opsi = !!x.acak_opsi; x.tampil_nilai = !!x.tampil_nilai;
  openForm({
    title: x.id ? 'Ubah ujian' : 'Buat ujian CBT', values: x, wide: true,
    fields: [[{ n: 'judul', l: 'Nama ujian', req: true, ph: 'PTS Ganjil Matematika' }, { n: 'mapel', l: 'Mata pelajaran', t: 'combo', o: mapel }],
      { n: 'kelas_ids', l: 'Kelas peserta', t: 'checks', o: S.cbtKelas.map(k => [k.id, k.nama]) },
      [{ n: 'mulai', l: 'Dibuka', t: 'datetime', req: true }, { n: 'selesai', l: 'Ditutup', t: 'datetime', req: true }],
      [{ n: 'durasi_menit', l: 'Durasi (menit)', t: 'number', min: 1, max: 600, req: true }, { n: 'jumlah_opsi', l: 'Jumlah pilihan', t: 'select', o: [[2, '2 (A–B)'], [3, '3 (A–C)'], [4, '4 (A–D)'], [5, '5 (A–E)']] }],
      { n: 'token', l: 'Token ujian', ph: 'Kosongkan jika tanpa token', hint: 'Token dibacakan pengawas saat ujian dimulai, agar siswa tidak mengerjakan dari luar ruang.' },
      { n: 'acak_soal', l: 'Acak urutan soal untuk tiap siswa', t: 'check' }, { n: 'acak_opsi', l: 'Acak urutan pilihan jawaban', t: 'check' },
      { n: 'tampil_nilai', l: 'Tampilkan nilai ke siswa setelah selesai', t: 'check' }],
    onSubmit: v => {
      const body = { ...v, kelas_ids: v.kelas_ids.map(Number), jumlah_opsi: Number(v.jumlah_opsi) };
      return x.id ? PUT('/cbt/' + x.id, body) : POST('/cbt', body);
    },
  });
};
ACT.cbtDel = el => del('/cbt/' + el.dataset.id, `Hapus ujian <b>${esc(el.dataset.judul)}</b> beserta semua soal dan hasilnya?`);

ACT.cbtSoal = el => {
  const x = S.cbtList[el.dataset.id];
  openModal({
    title: 'Soal: ' + x.judul, wide: true,
    builder: async () => {
      const qs = await GET(`/cbt/${x.id}/soal`);
      S.cbtSoal = Object.fromEntries(qs.map(q => [q.id, q]));
      const L = 'abcde'.slice(0, x.jumlah_opsi).split('');
      return `<div class="card-head"><p class="muted">${qs.length} soal · total bobot ${fmtN(qs.reduce((s, q) => s + q.bobot, 0))}</p>
        <div class="acts">${btn('Impor dari teks', 'cbtImport', { id: x.id }, 'ghost sm', 'upload')}${btn('Tambah soal', 'cbtSoalForm', { exam: x.id }, 'sm', 'plus')}</div></div>` +
        (qs.length ? qs.map((q, i) => `<div class="card" style="margin-top:10px;padding:14px"><div style="display:flex;gap:10px;justify-content:space-between">
          <div style="min-width:0"><p style="white-space:pre-wrap;font-weight:600"><span class="muted">${i + 1}.</span> ${esc(q.pertanyaan)}</p>
          ${q.gambar ? `<img src="${esc(q.gambar)}" alt="" style="max-height:140px;border-radius:10px;margin:6px 0">` : ''}
          <div style="display:grid;gap:3px;margin-top:6px">${L.map(l => `<div style="${q.jawaban_benar === l ? 'color:var(--ok);font-weight:700' : 'color:var(--text-2)'}">${l.toUpperCase()}. ${esc(q['opsi_' + l])} ${q.jawaban_benar === l ? ico('check') : ''}</div>`).join('')}</div></div>
          <div class="acts" style="flex:none">${iconBtn('edit', 'cbtSoalForm', { exam: x.id, id: q.id }, 'Ubah')}${iconBtn('trash', 'cbtSoalDel', { id: q.id, no: i + 1 }, 'Hapus', 'danger')}</div></div></div>`).join('')
          : empty('Belum ada soal. Tambah satu per satu atau impor dari teks.', 'quiz'));
    },
  });
};
ACT.cbtSoalForm = el => {
  const x = S.cbtList[el.dataset.exam];
  const q = el.dataset.id ? S.cbtSoal[el.dataset.id] : { bobot: 1, jawaban_benar: 'a' };
  const L = 'abcde'.slice(0, x.jumlah_opsi).split('');
  openForm({
    title: q.id ? 'Ubah soal' : 'Tambah soal', values: q, wide: true, success: 'Soal tersimpan.',
    fields: [{ n: 'pertanyaan', l: 'Pertanyaan', t: 'textarea', rows: 3, req: true }, { n: 'gambar', l: 'Gambar soal', t: 'file', accept: 'image/*' },
      ...L.map(l => ({ n: 'opsi_' + l, l: 'Pilihan ' + l.toUpperCase(), req: true })),
      [{ n: 'jawaban_benar', l: 'Kunci jawaban', t: 'select', o: L.map(l => [l, l.toUpperCase()]) }, { n: 'bobot', l: 'Bobot', t: 'number', min: 0.1, step: 'any', req: true }]],
    onSubmit: v => q.id ? PUT('/cbt-soal/' + q.id, v) : POST(`/cbt/${x.id}/soal`, v),
  });
};
ACT.cbtSoalDel = el => del('/cbt-soal/' + el.dataset.id, `Hapus soal nomor ${esc(el.dataset.no)}?`);

ACT.cbtImport = el => {
  const x = S.cbtList[el.dataset.id];
  openForm({
    title: 'Impor soal dari teks', wide: true, success: 'Soal diimpor.',
    intro: `<div class="hint" style="margin-bottom:12px">Tempel soal dari Word/Google Docs dengan format:<br><code>1. Ibu kota Indonesia adalah …</code><br><code>A. Bandung</code> <code>B. Jakarta</code> <code>C. Surabaya</code> <code>D. Medan</code><br><code>Kunci: B</code><br>Gambar bisa ditambahkan setelah impor lewat tombol ubah.</div>`,
    fields: [{ n: 'teks', l: 'Teks soal', t: 'textarea', rows: 14, req: true }],
    submit: 'Impor',
    onSubmit: async v => {
      const L = 'abcde'.slice(0, x.jumlah_opsi).split('');
      const qs = parseSoalText(v.teks);
      if (!qs.length) throw new Error('Tidak ada soal yang terbaca. Periksa formatnya.');
      const bad = qs.findIndex(q => !q.kunci || L.some(l => !q.opsi[l]));
      if (bad >= 0) throw new Error(`Soal nomor ${bad + 1}: pilihan ${L.map(l => l.toUpperCase()).join('/')} atau kunci belum lengkap.`);
      // satu permintaan, satu transaksi: semua masuk atau tidak sama sekali
      await POST(`/cbt/${x.id}/soal/batch`, { soal: qs.map(q => ({ pertanyaan: q.pertanyaan, jawaban_benar: q.kunci, bobot: 1, ...Object.fromEntries(L.map(l => ['opsi_' + l, q.opsi[l]])) })) });
    },
  });
};

ACT.cbtHasil = el => {
  const x = S.cbtList[el.dataset.id];
  openModal({
    title: 'Hasil: ' + x.judul, wide: true,
    builder: async () => {
      const rows = await GET(`/cbt/${x.id}/hasil`);
      const done = rows.filter(r => r.selesai);
      const avg = done.length ? done.reduce((s, r) => s + r.skor, 0) / done.length : null;
      S.cbtHasil = { x, rows };
      return `<div class="stats">${stat('users', done.length + '/' + rows.length, 'Selesai')}${stat('chart', fmtN(avg), 'Rata-rata')}${stat('check', done.length ? fmtN(Math.max(...done.map(r => r.skor))) : '–', 'Tertinggi')}</div>
        <div class="acts" style="margin-bottom:10px">${btn('Unduh CSV', 'cbtCSV', {}, 'ghost sm', 'download')}${btn('Cetak', 'cbtPrint', {}, 'ghost sm', 'print')}</div>
        <div class="tw"><table><thead><tr><th>Siswa</th><th>Kelas</th><th>Status</th><th class="n">Nilai</th><th></th></tr></thead><tbody>
        ${rows.map(r => `<tr><td><b>${esc(r.nama_lengkap)}</b><small class="muted">${esc(r.username)}</small></td><td>${esc(r.kelas_nama)}</td>
          <td>${!r.attempt_id ? tag('Belum') : r.selesai ? tag('Selesai · ' + fmtDT(r.submitted_at), 'ok') : tag('Mengerjakan', 'info')}</td>
          <td class="n"><b>${r.selesai ? fmtN(r.skor) : '–'}</b></td><td>${r.attempt_id ? btn('Ulang', 'cbtReset', { id: r.attempt_id, nama: r.nama_lengkap }, 'ghost sm') : ''}</td></tr>`).join('')}
        </tbody></table></div><p class="muted" style="font-size:12px;margin-top:8px">"Ulang" menghapus pekerjaan siswa agar bisa mengerjakan kembali (mis. komputer mati).</p>`;
    },
  });
};
ACT.cbtReset = el => del('/cbt-attempt/' + el.dataset.id, `Hapus pekerjaan <b>${esc(el.dataset.nama)}</b> agar bisa mengulang?`, 'Pekerjaan dihapus.');
ACT.cbtCSV = () => { const { x, rows } = S.cbtHasil; downloadText(`hasil-${x.judul}.csv`.replace(/\s+/g, '-'), toCSV([['No', 'Nama', 'NIS', 'Kelas', 'Nilai'], ...rows.map((r, i) => [i + 1, r.nama_lengkap, r.username, r.kelas_nama, r.selesai ? fmtN(r.skor) : ''])])); };
ACT.cbtPrint = () => { const { x, rows } = S.cbtHasil; printDoc(schoolHead() + `<h3 style="text-align:center">DAFTAR NILAI ${esc(x.judul.toUpperCase())}</h3><p>Mapel: ${esc(x.mapel || '–')}</p>
  <table><thead><tr><th>No</th><th>Nama</th><th>NIS</th><th>Kelas</th><th>Nilai</th></tr></thead><tbody>${rows.map((r, i) => `<tr><td>${i + 1}</td><td>${esc(r.nama_lengkap)}</td><td>${esc(r.username)}</td><td>${esc(r.kelas_nama)}</td><td>${r.selesai ? fmtN(r.skor) : ''}</td></tr>`).join('')}</tbody></table>`); };

// =====================================================================
// CBT — siswa
// =====================================================================
async function siswaCBT() {
  const xs = await GET('/cbt');
  const now = new Date();
  return head('Ujian CBT', 'Ujian dibuka sesuai jadwal. Siapkan token dari pengawas bila diminta.') +
    (xs.length ? `<div class="grid">${xs.map(x => {
      const open = now >= new Date(x.mulai) && now <= new Date(x.selesai);
      const st = x.sudah === 1 ? [x.skor != null ? 'Nilai ' + fmtN(x.skor) : 'Selesai', 'ok'] : x.started_at && open ? ['Sedang dikerjakan', 'info'] : now < new Date(x.mulai) ? ['Belum dibuka', ''] : now > new Date(x.selesai) ? ['Ditutup', 'bad'] : ['Tersedia', 'brand'];
      return `<section class="card" style="display:flex;flex-direction:column;gap:8px"><div style="display:flex;justify-content:space-between;gap:8px">${tag(x.mapel || 'Ujian', 'brand')}${tag(st[0], st[1])}</div>
        <h3>${esc(x.judul)}</h3><small class="muted">${x.jumlah_soal} soal · ${x.durasi_menit} menit<br>${fmtDT(x.mulai)} – ${fmtDT(x.selesai)}</small>
        ${x.sudah !== 1 && open ? `<div>${btn(x.started_at ? 'Lanjutkan' : 'Mulai ujian', 'cbtGo', { id: x.id, judul: x.judul, durasi: x.durasi_menit, lanjut: x.started_at ? 1 : 0 }, 'sm')}</div>` : ''}</section>`; }).join('')}</div>`
      : `<div class="card">${empty('Belum ada ujian untuk kelas Anda.', 'quiz')}</div>`);
}
ACT.cbtGo = async el => {
  if (el.dataset.lanjut !== '1' && !(await confirmBox(`Mulai <b>${esc(el.dataset.judul)}</b>? Waktu ${esc(el.dataset.durasi)} menit berjalan sejak Anda mulai dan tetap berjalan walau halaman ditutup.`, 'Mulai', ''))) return;
  go('cbtujian', { id: el.dataset.id });
};

S.cbtTok = {};
async function pageCBTUjian(p) {
  const id = Number(p.id);
  let r;
  try { r = await POST(`/cbt/${id}/start`, { token: S.cbtTok[id] || '' }); }
  catch (e) {
    if (e.message !== 'TOKEN') throw e;
    const fid = registerForm([{ n: 'token' }], async v => { S.cbtTok[id] = v.token; await refresh(false); });
    return `<div class="card" style="max-width:420px;margin:40px auto;text-align:center;padding:28px">
      <div class="stat" style="justify-content:center;border:0;box-shadow:none"><div class="ic">${ico('key')}</div></div>
      <h2>Masukkan token ujian</h2><p class="muted" style="margin:6px 0 16px">Token diberikan pengawas di ruang ujian.</p>
      <form class="form" data-form="${fid}"><input name="token" required autocomplete="off" autocapitalize="characters" class="scan-input" placeholder="TOKEN">
      <div class="form-err" role="alert">${S.cbtTok[id] ? 'Token salah. Coba lagi.' : ''}</div><button class="btn block" type="submit">Mulai</button></form></div>`;
  }
  // kunci lokal per attempt (berubah bila guru menekan "Ulang"); jawaban di server selalu menang
  const key = `sch_cbt_${S.user.id}_${id}_${r.started_at || ''}`;
  const local = store.get(key) || {};
  S.cbt = { id, key, qs: r.questions, ans: { ...(local.ans || {}), ...r.answers }, ragu: new Set(local.ragu || []), cur: 0, end: Date.now() + r.sisa_detik * 1000, sent: false, judul: r.exam.judul };
  S.after = () => {
    renderCbt();
    const tick = () => {
      const left = Math.max(0, Math.round((S.cbt.end - Date.now()) / 1000));
      const t = $('#cbtTimer');
      if (!t) return;
      t.textContent = cbtTimeText();
      t.classList.toggle('low', left <= 300);
      if (left <= 0 && !S.cbt.sent) {
        clearInterval(S.timer); S.timer = null; toast('Waktu habis. Jawaban dikirim otomatis.');
        const kirim = () => Promise.resolve(cbtSubmit(true)).catch(() => {}).then(() => {
          if (!S.cbt.sent && document.body.contains(t)) { toast('Gagal mengirim, mencoba lagi…', 'bad'); S.timer = setTimeout(kirim, 10000); }
        });
        kirim();
      }
    };
    tick();
    S.timer = setInterval(tick, 1000);
  };
  return `<div class="cbt"><section class="card cbt-q" id="cbtQ"></section><aside class="card cbt-nav" id="cbtNav"></aside></div>`;
}
function renderCbt() {
  const c = S.cbt, q = c.qs[c.cur];
  const L = 'ABCDE';
  $('#cbtQ').innerHTML = `<div class="card-head"><h3>Soal ${c.cur + 1} dari ${c.qs.length}</h3>${c.ragu.has(String(q.id)) ? tag('Ragu-ragu', 'warn') : ''}</div>
    <p class="tanya" style="white-space:pre-wrap;font-weight:600;margin-bottom:10px">${esc(q.pertanyaan)}</p>
    ${q.gambar ? `<img class="soal-img" src="${esc(q.gambar)}" alt="Gambar soal ${c.cur + 1}">` : ''}
    ${q.opsi.map((o, i) => `<label class="opt"><input type="radio" name="cbtopt" value="${'abcde'[i]}" data-change="cbtPick" ${c.ans[q.id] === 'abcde'[i] ? 'checked' : ''}><b>${L[i]}.</b><span>${esc(o)}</span></label>`).join('')}
    <div class="acts" style="justify-content:space-between;margin-top:14px">
      ${btn('Sebelumnya', 'cbtMove', { d: -1 }, 'ghost', 'back')}
      <label class="chk" style="font-weight:600"><input type="checkbox" data-change="cbtRagu" ${c.ragu.has(String(q.id)) ? 'checked' : ''}> Ragu-ragu</label>
      ${c.cur < c.qs.length - 1 ? btn('Berikutnya', 'cbtMove', { d: 1 }, '') : btn('Selesai', 'cbtFinish', {}, '', 'check')}</div>`;
  const answered = Object.keys(c.ans).filter(k => c.ans[k]).length;
  $('#cbtNav').innerHTML = `<b>${esc(c.judul)}</b><div class="timer" id="cbtTimer" style="font-size:28px;margin-top:6px">${cbtTimeText()}</div>
    <small class="muted">${answered} dari ${c.qs.length} dijawab</small>
    <div class="nums">${c.qs.map((x, i) => `<button type="button" data-act="cbtJump" data-i="${i}" class="${c.ragu.has(String(x.id)) ? 'ragu' : c.ans[x.id] ? 'ans' : ''} ${i === c.cur ? 'cur' : ''}" aria-label="Soal ${i + 1}">${i + 1}</button>`).join('')}</div>
    <div class="cbt-legend"><span><i style="background:var(--ok)"></i>Dijawab</span><span><i style="background:var(--accent)"></i>Ragu</span><span><i style="background:var(--surface-2);border:1px solid var(--border)"></i>Belum</span></div>
    <div style="margin-top:14px">${btn('Kirim jawaban', 'cbtFinish', {}, 'block', 'check')}</div>`;
}
function cbtTimeText() {
  const left = Math.max(0, Math.round((S.cbt.end - Date.now()) / 1000));
  const h = Math.floor(left / 3600);
  return (h ? h + ':' : '') + pad(Math.floor(left / 60) % 60) + ':' + pad(left % 60);
}
let cbtSaveT = null;
function cbtPersist() {
  const c = S.cbt;
  store.set(c.key, { ans: c.ans, ragu: [...c.ragu] });
  clearTimeout(cbtSaveT);
  cbtSaveT = setTimeout(() => POST(`/cbt/${c.id}/save`, { answers: c.ans }).catch(() => { /* coba lagi pada perubahan berikutnya */ }), 1200);
}
CHG.cbtPick = el => { const q = S.cbt.qs[S.cbt.cur]; S.cbt.ans[q.id] = el.value; cbtPersist(); renderCbt(); };
CHG.cbtRagu = el => { const id = String(S.cbt.qs[S.cbt.cur].id); el.checked ? S.cbt.ragu.add(id) : S.cbt.ragu.delete(id); cbtPersist(); renderCbt(); };
ACT.cbtMove = el => { S.cbt.cur = Math.max(0, Math.min(S.cbt.qs.length - 1, S.cbt.cur + Number(el.dataset.d))); renderCbt(); window.scrollTo(0, 0); };
ACT.cbtJump = el => { S.cbt.cur = Number(el.dataset.i); renderCbt(); window.scrollTo(0, 0); };
ACT.cbtFinish = async () => {
  const c = S.cbt;
  const kosong = c.qs.filter(q => !c.ans[q.id]).length;
  const msg = kosong ? `Masih ada <b>${kosong} soal</b> belum dijawab${c.ragu.size ? ` dan <b>${c.ragu.size}</b> ditandai ragu` : ''}. Tetap kirim?` : c.ragu.size ? `Ada <b>${c.ragu.size} soal</b> ditandai ragu-ragu. Tetap kirim?` : 'Kirim semua jawaban? Jawaban tidak bisa diubah lagi.';
  if (await confirmBox(msg, 'Kirim', '')) await cbtSubmit(false);
};
async function cbtSubmit(auto) {
  const c = S.cbt;
  if (c.sent) return;
  c.sent = true;
  clearTimeout(cbtSaveT);
  try {
    const r = await POST(`/cbt/${c.id}/submit`, { answers: c.ans });
    store.del(c.key);
    if (S.timer) { clearInterval(S.timer); S.timer = null; }
    $('#main').innerHTML = `<div class="card" style="max-width:480px;margin:40px auto;text-align:center;padding:32px">
      <div class="stat" style="justify-content:center;border:0;box-shadow:none"><div class="ic">${ico('check')}</div></div>
      <h2>${auto ? 'Waktu habis, jawaban terkirim' : 'Ujian selesai'}</h2>
      ${r.skor != null ? `<p class="muted" style="margin:6px 0 10px">Nilai Anda</p><div style="font-size:52px;font-weight:800;color:var(--brand)">${fmtN(r.skor)}</div>` : '<p class="muted" style="margin-top:8px">Nilai akan diumumkan oleh guru.</p>'}
      <div style="margin-top:20px">${btn('Kembali', 'nav', { page: 'cbt' }, 'ghost')}</div></div>`;
  } catch (e) { c.sent = false; toast(e.message + ' Jawaban tetap tersimpan. Coba kirim lagi.', 'bad'); }
}

Object.assign(PAGES.admin, { cbt: pageCBT });
Object.assign(PAGES.guru, { cbt: pageCBT });
Object.assign(PAGES.siswa, { cbt: pageCBT, cbtujian: pageCBTUjian });
