'use strict';
/* Halaman guru */

// =====================================================================
// HALAMAN GURU
// =====================================================================
async function guruCourses() {
  const cs = await GET('/courses' + (S.tahunGuru ? '?tahun=' + S.tahunGuru : ''));
  const want = Number(S.params.course) || S.course;
  S.course = cs.some(c => c.id === want) ? want : cs[0]?.id ?? null;
  return cs;
}
CHG.guruTahun = el => { S.tahunGuru = el.value; S.course = null; refresh(false); };

const courseBar = (cs, extra = '') => `<div class="bar no-print"><select data-change="pickCourse" aria-label="Pilih mapel" style="min-width:240px">${opts(cs.map(c => [c.id, `${c.nama} · ${c.kelas_nama}`]), S.course)}</select>${extra}</div>`;
CHG.pickCourse = el => go(S.page, { ...S.params, course: el.value });
const noCourse = () => `<div class="card">${empty('Anda belum ditugaskan mengampu mapel. Hubungi admin sekolah.', 'book')}</div>`;

async function guruDash() {
  const d = await GET('/dashboard');
  return head('Beranda', `Selamat datang, ${esc(S.user.nama_lengkap)}`) +
    `<div class="stats">${stat('book', d.mapel, 'Mapel diampu')}${stat('users', d.siswa, 'Siswa')}${stat('layers', d.materi, 'Materi')}${stat('clip', d.tugas, 'Tugas')}${stat('quiz', d.kuis, 'Kuis')}</div>
    <section class="card"><div class="card-head"><h3>Tugas menunggu penilaian</h3>${d.pending.length ? tag(d.pending.length + (d.pending.length >= 20 ? '+' : ''), 'warn') : ''}</div>
      ${d.pending.length ? d.pending.map(p => `<div class="li"><div><b>${esc(p.nama_lengkap)}</b><small class="muted">${esc(p.judul)} · ${esc(p.mapel)} ${esc(p.kelas_nama)} · ${fmtDT(p.submitted_at)}</small></div>
        ${btn('Nilai', 'gradeOpen', { id: p.assignment_id, judul: p.judul }, 'sm')}</div>`).join('') : empty('Tidak ada tugas yang menunggu penilaian.', 'check')}</section>`;
}

async function guruMapel() {
  const ys = await GET('/tahun');
  const ty = yearPick(ys, S.tahunGuru);
  const cs = await GET('/courses?tahun=' + ty.id);
  return head('Mapel Saya', `Mata pelajaran yang Anda ampu, tahun ajaran <b>${esc(ty.nama)}</b>.`) +
    yearBar(ys, ty, 'guruTahun', 'Pilih tahun lalu untuk membuka arsip materi, tugas, dan rekap nilai.') + (cs.length ? `<div class="grid">${cs.map(c => `<section class="card" style="display:flex;flex-direction:column;gap:10px">
    <div style="display:flex;justify-content:space-between;gap:8px">${tag(c.kode, 'brand')}${tag(c.kelas_nama)}</div>
    <h3>${esc(c.nama)}</h3><small class="muted">${c.jumlah_siswa} siswa${c.deskripsi ? ' · ' + esc(c.deskripsi) : ''}</small>
    <div class="acts" style="justify-content:flex-start">${[['materi', 'Materi'], ['tugas', 'Tugas'], ['kuis', 'Kuis'], ['absensi', 'Absensi'], ['rekap', 'Rekap']].map(([p, l]) =>
      btn(l, 'go', { page: p, params: JSON.stringify({ course: c.id }) }, 'ghost sm')).join('')}</div></section>`).join('')}</div>` : noCourse());
}

async function guruMateri() {
  const cs = await guruCourses();
  if (!cs.length) return head('Materi') + noCourse();
  const c = cs.find(x => x.id === S.course);
  const ms = await GET('/materials?course_id=' + S.course);
  S.matNext = (ms.reduce((m, x) => Math.max(m, x.urutan), 0) || 0) + 1;
  return head('Materi', 'Materi tampil ke siswa sesuai urutan.', btn('Tambah materi', 'matForm', {}, '', 'plus')) + courseBar(cs) +
    (ms.length ? ms.map((m, i) => matHtml(m, i, { actions: `<span class="tag ${m.jumlah_selesai >= c.jumlah_siswa && c.jumlah_siswa ? 'ok' : ''}" title="Siswa yang menandai selesai">${m.jumlah_selesai}/${c.jumlah_siswa} selesai</span>
      ${iconBtn('edit', 'matForm', { m: JSON.stringify(m) }, 'Ubah')}${iconBtn('trash', 'matDel', { id: m.id, judul: m.judul }, 'Hapus', 'danger')}` })).join('')
      : `<div class="card">${empty('Belum ada materi untuk mapel ini.', 'layers')}</div>`);
}
ACT.matForm = el => {
  const m = el.dataset.m ? JSON.parse(el.dataset.m) : { urutan: S.matNext };
  openForm({
    title: m.id ? 'Ubah materi' : 'Tambah materi', wide: true,
    values: { ...m, lampiran: lampText(m.lampiran) },
    fields: [[{ n: 'judul', l: 'Judul materi', req: true }, { n: 'urutan', l: 'Urutan', t: 'number', min: 1, req: true }],
      { n: 'konten', l: 'Isi materi', t: 'textarea', rows: 8, ph: 'Tulis ringkasan, penjelasan, atau instruksi belajar…' },
      { n: 'lampiran', l: 'Lampiran', t: 'textarea', rows: 4, ph: 'Video pengantar | https://youtu.be/xxxx\nModul PDF | https://drive.google.com/…', hint: 'Satu tautan per baris, format "Judul | tautan". Tautan YouTube tampil sebagai video, tautan gambar tampil sebagai gambar.' }],
    onSubmit: v => {
      const body = { course_id: S.course, judul: v.judul, konten: v.konten, urutan: v.urutan || 1, lampiran: parseLamp(v.lampiran) };
      return m.id ? PUT('/materials/' + m.id, body) : POST('/materials', body);
    },
  });
};
ACT.matDel = el => del('/materials/' + el.dataset.id, `Hapus materi <b>${esc(el.dataset.judul)}</b>?`);

async function guruTugas() {
  const cs = await guruCourses();
  if (!cs.length) return head('Tugas') + noCourse();
  const c = cs.find(x => x.id === S.course);
  const as = await GET('/assignments?course_id=' + S.course);
  return head('Tugas', 'Buat tugas, lalu nilai hasil kerja siswa.', btn('Buat tugas', 'tugasForm', {}, '', 'plus')) + courseBar(cs) +
    (as.length ? `<div class="tw"><table><thead><tr><th>Tugas</th><th>Tenggat</th><th class="n">Terkumpul</th><th class="n">Dinilai</th><th></th></tr></thead><tbody>
    ${as.map(a => `<tr><td><b>${esc(a.judul)}</b><small class="muted">Nilai maks ${fmtN(a.nilai_maks)}</small></td>
      <td>${a.deadline ? fmtDT(a.deadline) + (new Date(a.deadline) < new Date() ? ' ' + tag('Lewat', 'mute') : '') : '–'}</td>
      <td class="n">${a.jumlah_kumpul}/${c.jumlah_siswa}</td><td class="n">${a.jumlah_dinilai < a.jumlah_kumpul ? tag(a.jumlah_dinilai + '/' + a.jumlah_kumpul, 'warn') : a.jumlah_dinilai}</td>
      <td><div class="acts">${btn('Nilai', 'gradeOpen', { id: a.id, judul: a.judul }, 'sm')}${iconBtn('edit', 'tugasForm', { a: JSON.stringify(a) }, 'Ubah')}${iconBtn('trash', 'tugasDel', { id: a.id, judul: a.judul }, 'Hapus', 'danger')}</div></td></tr>`).join('')}
    </tbody></table></div>` : `<div class="card">${empty('Belum ada tugas untuk mapel ini.', 'clip')}</div>`);
}
ACT.tugasForm = el => {
  const a = el.dataset.a ? JSON.parse(el.dataset.a) : { nilai_maks: 100 };
  openForm({
    title: a.id ? 'Ubah tugas' : 'Buat tugas', values: a,
    fields: [{ n: 'judul', l: 'Judul tugas', req: true }, { n: 'deskripsi', l: 'Instruksi', t: 'textarea', rows: 5 },
      [{ n: 'deadline', l: 'Tenggat', t: 'datetime' }, { n: 'nilai_maks', l: 'Nilai maksimal', t: 'number', min: 1, req: true }]],
    onSubmit: v => {
      const body = { ...v, course_id: S.course };
      return a.id ? PUT('/assignments/' + a.id, body) : POST('/assignments', body);
    },
  });
};
ACT.tugasDel = el => del('/assignments/' + el.dataset.id, `Hapus tugas <b>${esc(el.dataset.judul)}</b>? Tugas yang sudah dikumpulkan siswa tidak bisa dihapus agar nilainya tidak hilang.`);

ACT.gradeOpen = el => {
  const aid = Number(el.dataset.id);
  openModal({
    title: 'Penilaian: ' + el.dataset.judul, wide: true,
    builder: async () => {
      const { assignment: a, students: st, submissions: subs } = await GET(`/assignments/${aid}/submissions`);
      const bySiswa = Object.fromEntries(subs.map(s => [s.siswa_id, s]));
      return `<p class="muted" style="margin-bottom:12px">${subs.length} dari ${st.length} siswa mengumpulkan · nilai maksimal ${fmtN(a.nilai_maks)}${a.deadline ? ' · tenggat ' + fmtDT(a.deadline) : ''}</p>
      <div class="tw"><table><thead><tr><th>Siswa</th><th>Jawaban</th><th style="width:110px">Nilai</th><th>Catatan</th><th></th></tr></thead><tbody>
      ${st.map(s => {
        const x = bySiswa[s.id];
        if (!x) return `<tr><td><b>${esc(s.nama_lengkap)}</b><small class="muted">${esc(s.username)}</small></td><td colspan="4">${tag('Belum mengumpulkan')}</td></tr>`;
        const late = a.deadline && new Date(x.submitted_at) > new Date(a.deadline);
        return `<tr data-sub="${x.id}"><td><b>${esc(s.nama_lengkap)}</b><small class="muted">${fmtDT(x.submitted_at)} ${late ? tag('Terlambat', 'warn') : ''}</small></td>
          <td style="max-width:320px">${x.jawaban ? `<div style="white-space:pre-wrap;max-height:120px;overflow:auto">${esc(x.jawaban)}</div>` : ''}${x.link_file ? `<a href="${esc(safeUrl(x.link_file))}" target="_blank" rel="noopener">${ico('link')} Buka file</a>` : ''}</td>
          <td><input type="number" name="nilai" min="0" max="${a.nilai_maks}" step="any" value="${x.nilai ?? ''}" aria-label="Nilai ${esc(s.nama_lengkap)}"></td>
          <td><input name="feedback" value="${esc(x.feedback || '')}" placeholder="Opsional" aria-label="Catatan untuk ${esc(s.nama_lengkap)}"></td>
          <td>${btn('Simpan', 'saveGrade', { id: x.id }, x.nilai == null ? 'sm' : 'ghost sm')}</td></tr>`;
      }).join('')}</tbody></table></div>`;
    },
  });
};
ACT.saveGrade = async el => {
  const tr = el.closest('tr');
  const n = $('input[name=nilai]', tr).value;
  el.disabled = true;
  try {
    await PUT(`/submissions/${el.dataset.id}/grade`, { nilai: num(n), feedback: $('input[name=feedback]', tr).value });
    toast('Nilai tersimpan.');
    el.classList.add('ghost');
    refresh(true);
  } finally { el.disabled = false; }
};

async function guruKuis() {
  const cs = await guruCourses();
  if (!cs.length) return head('Kuis') + noCourse();
  const c = cs.find(x => x.id === S.course);
  const qs = await GET('/quizzes?course_id=' + S.course);
  qs.forEach(q => { S.quizzes[q.id] = q; });
  return head('Kuis', 'Soal pilihan ganda. Skor dihitung otomatis, kunci jawaban tidak pernah dikirim ke siswa.', btn('Buat kuis', 'kuisForm', {}, '', 'plus')) + courseBar(cs) +
    (qs.length ? `<div class="tw"><table><thead><tr><th>Kuis</th><th class="n">Durasi</th><th class="n">Soal</th><th class="n">Dikerjakan</th><th></th></tr></thead><tbody>
    ${qs.map(q => `<tr><td><b>${esc(q.judul)}</b><small class="muted">${q.jumlah_opsi} pilihan jawaban (A–${'ABCDE'[q.jumlah_opsi - 1]})</small></td><td class="n">${q.durasi_menit} mnt</td>
      <td class="n">${q.jumlah_soal ? q.jumlah_soal : tag('0', 'warn')}</td><td class="n">${q.jumlah_dikerjakan}/${c.jumlah_siswa}</td>
      <td><div class="acts">${btn('Soal', 'soalOpen', { id: q.id }, 'sm')}${btn('Hasil', 'hasilOpen', { id: q.id }, 'ghost sm')}${iconBtn('edit', 'kuisForm', { id: q.id }, 'Ubah')}${iconBtn('trash', 'kuisDel', { id: q.id, judul: q.judul }, 'Hapus', 'danger')}</div></td></tr>`).join('')}
    </tbody></table></div>` : `<div class="card">${empty('Belum ada kuis untuk mapel ini.', 'quiz')}</div>`);
}
ACT.kuisForm = el => {
  const q = el.dataset.id ? S.quizzes[el.dataset.id] : { durasi_menit: 30, jumlah_opsi: 4 };
  openForm({
    title: q.id ? 'Ubah kuis' : 'Buat kuis', values: q,
    fields: [{ n: 'judul', l: 'Judul kuis', req: true, ph: 'Ulangan Harian Bab 1' },
      [{ n: 'durasi_menit', l: 'Durasi (menit)', t: 'number', min: 1, max: 600, req: true },
        { n: 'jumlah_opsi', l: 'Jumlah pilihan', t: 'select', o: [[2, '2 (A–B)'], [3, '3 (A–C)'], [4, '4 (A–D)'], [5, '5 (A–E)']] }]],
    onSubmit: v => {
      const body = { ...v, jumlah_opsi: Number(v.jumlah_opsi), course_id: S.course };
      return q.id ? PUT('/quizzes/' + q.id, body) : POST('/quizzes', body);
    },
  });
};
ACT.kuisDel = el => del('/quizzes/' + el.dataset.id, `Hapus kuis <b>${esc(el.dataset.judul)}</b> beserta soalnya? Kuis yang sudah dikerjakan siswa tidak bisa dihapus agar nilainya tidak hilang.`);

ACT.soalOpen = el => {
  const q = S.quizzes[el.dataset.id];
  openModal({
    title: 'Soal: ' + q.judul, wide: true,
    builder: async () => {
      const xs = await GET(`/quizzes/${q.id}/questions`);
      S.soal = Object.fromEntries(xs.map(x => [x.id, x]));
      const letters = 'abcde'.slice(0, q.jumlah_opsi).split('');
      return `<div class="card-head"><p class="muted">${xs.length} soal · total bobot ${fmtN(xs.reduce((s, x) => s + x.bobot, 0))}</p>${btn('Tambah soal', 'soalForm', { quiz: q.id }, 'sm', 'plus')}</div>` +
        (xs.length ? xs.map((x, i) => `<div class="card" style="margin-top:10px;padding:14px">
          <div style="display:flex;gap:10px;justify-content:space-between;align-items:flex-start">
            <p style="white-space:pre-wrap;font-weight:600"><span class="muted">${i + 1}.</span> ${esc(x.pertanyaan)}</p>
            <div class="acts" style="flex:none">${iconBtn('edit', 'soalForm', { quiz: q.id, id: x.id }, 'Ubah soal')}${iconBtn('trash', 'soalDel', { id: x.id, no: i + 1 }, 'Hapus soal', 'danger')}</div></div>
          <div style="display:grid;gap:4px;margin-top:8px">${letters.map(l => `<div style="${x.jawaban_benar === l ? 'color:var(--ok);font-weight:700' : 'color:var(--text-2)'}">${l.toUpperCase()}. ${esc(x['opsi_' + l])} ${x.jawaban_benar === l ? ico('check') : ''}</div>`).join('')}</div>
          ${x.bobot !== 1 ? `<small class="muted">Bobot ${fmtN(x.bobot)}</small>` : ''}</div>`).join('') : empty('Belum ada soal. Siswa belum bisa mengerjakan kuis ini.', 'quiz'));
    },
  });
};
ACT.soalForm = el => {
  const q = S.quizzes[el.dataset.quiz];
  const x = el.dataset.id ? S.soal[el.dataset.id] : { bobot: 1, jawaban_benar: 'a' };
  const letters = 'abcde'.slice(0, q.jumlah_opsi).split('');
  openForm({
    title: x.id ? 'Ubah soal' : 'Tambah soal', values: x, wide: true,
    fields: [{ n: 'pertanyaan', l: 'Pertanyaan', t: 'textarea', req: true, rows: 3 },
      ...letters.map(l => ({ n: 'opsi_' + l, l: 'Pilihan ' + l.toUpperCase(), req: true })),
      [{ n: 'jawaban_benar', l: 'Kunci jawaban', t: 'select', o: letters.map(l => [l, l.toUpperCase()]) }, { n: 'bobot', l: 'Bobot', t: 'number', min: 0.1, step: 'any', req: true }]],
    onSubmit: v => x.id ? PUT('/questions/' + x.id, v) : POST(`/quizzes/${q.id}/questions`, v),
    success: 'Soal tersimpan.',
  });
};
ACT.soalDel = el => del('/questions/' + el.dataset.id, `Hapus soal nomor ${esc(el.dataset.no)}?`);

ACT.hasilOpen = el => {
  const q = S.quizzes[el.dataset.id];
  openModal({
    title: 'Hasil: ' + q.judul, wide: true,
    builder: async () => {
      const [st, at] = await Promise.all([GET(`/courses/${q.course_id}/students`), GET(`/quizzes/${q.id}/attempts`)]);
      const by = Object.fromEntries(at.map(a => [a.siswa_id, a]));
      const done = at.filter(a => a.selesai);
      const avg = done.length ? done.reduce((s, a) => s + a.skor, 0) / done.length : null;
      return `<div class="stats">${stat('users', done.length + '/' + st.length, 'Sudah mengerjakan')}${stat('chart', fmtN(avg), 'Rata-rata skor')}
        ${stat('check', done.length ? fmtN(Math.max(...done.map(a => a.skor))) : '–', 'Skor tertinggi')}</div>
      <div class="tw"><table><thead><tr><th>Siswa</th><th>Status</th><th class="n">Skor</th><th></th></tr></thead><tbody>
      ${st.map(s => { const a = by[s.id];
        return `<tr><td><b>${esc(s.nama_lengkap)}</b><small class="muted">${esc(s.username)}</small></td>
        <td>${!a ? tag('Belum') : a.selesai ? tag('Selesai · ' + fmtDT(a.submitted_at), 'ok') : tag('Sedang mengerjakan', 'info')}</td>
        <td class="n"><b>${a && a.selesai ? fmtN(a.skor) : '–'}</b></td>
        <td>${a ? `<div class="acts">${btn('Ulang', 'attemptReset', { id: a.id, nama: s.nama_lengkap }, 'ghost sm')}</div>` : ''}</td></tr>`; }).join('')}
      </tbody></table></div><p class="muted" style="margin-top:10px;font-size:13px">"Ulang" menghapus hasil siswa agar ia bisa mengerjakan kembali.</p>`;
    },
  });
};
ACT.attemptReset = el => del('/attempts/' + el.dataset.id, `Hapus hasil kuis <b>${esc(el.dataset.nama)}</b> agar bisa mengerjakan ulang?`, 'Hasil dihapus.');

async function guruAbsensi() {
  const cs = await guruCourses();
  if (!cs.length) return head('Absensi') + noCourse();
  const tgl = /^\d{4}-\d{2}-\d{2}$/.test(S.params.tanggal || '') ? S.params.tanggal : today();
  const [st, at] = await Promise.all([GET(`/courses/${S.course}/students`), GET(`/attendance?course_id=${S.course}&tanggal=${tgl}`)]);
  const rec = Object.fromEntries(at.records.map(r => [r.siswa_id, r.status]));
  const saved = at.records.length > 0;
  const counts = { hadir: 0, izin: 0, sakit: 0, alfa: 0 };
  at.records.forEach(r => counts[r.status]++);
  S.absen = { tgl, st };
  const seg = (sid, cur) => `<div class="seg" role="radiogroup">${['hadir', 'izin', 'sakit', 'alfa'].map(s =>
    `<label><input type="radio" name="st_${sid}" value="${s}" ${cur === s ? 'checked' : ''}><span class="${s}">${s[0].toUpperCase() + s.slice(1)}</span></label>`).join('')}</div>`;
  return head('Absensi', 'Pilih tanggal, tandai kehadiran, lalu simpan.') +
    courseBar(cs, `<input type="date" value="${tgl}" data-change="pickDate" aria-label="Tanggal" style="min-width:0">`) +
    `<div class="cols cols-21">
    <section>${st.length ? `<div class="card-head"><div>${saved ? tag('Tersimpan', 'ok') + ' ' + Object.entries(counts).filter(([, n]) => n).map(([k, n]) => tag(`${k} ${n}`)).join(' ') : tag('Belum diisi', 'warn')}</div>
      <div class="acts">${btn('Semua hadir', 'absenAll', {}, 'ghost sm')}${saved ? btn('Hapus', 'absenDel', {}, 'danger sm') : ''}${btn('Simpan absensi', 'absenSave', {}, 'sm', 'check')}</div></div>
      <div class="tw" id="absBox"><table><thead><tr><th>Siswa</th><th>Kehadiran</th></tr></thead><tbody>
      ${st.map(s => `<tr><td><b>${esc(s.nama_lengkap)}</b><small class="muted">${esc(s.username)}</small></td><td>${seg(s.id, rec[s.id] || 'hadir')}</td></tr>`).join('')}
      </tbody></table></div>
      <div class="save-bar">${btn('Simpan absensi (' + st.length + ' siswa)', 'absenSave', {}, '', 'check')}</div>` : `<div class="card">${empty('Belum ada siswa di kelas ini.', 'users')}</div>`}</section>
    <section class="card"><div class="card-head"><h3>Riwayat</h3></div>
      ${at.dates.length ? at.dates.map(d => `<div class="li"><div><b>${fmtDate(d.tanggal)}</b><small class="muted">${d.hadir}/${d.jumlah} hadir</small></div>
        ${d.tanggal === tgl ? tag('Dibuka', 'brand') : btn('Buka', 'go', { page: 'absensi', params: JSON.stringify({ course: S.course, tanggal: d.tanggal }) }, 'ghost sm')}</div>`).join('') : empty('Belum ada riwayat absensi.', 'cal')}</section></div>`;
}
CHG.pickDate = el => go('absensi', { course: S.course, tanggal: el.value });
ACT.absenAll = () => $$('#absBox input[value=hadir]').forEach(i => { i.checked = true; });
ACT.absenSave = async el => {
  const items = S.absen.st.map(s => ({ siswa_id: s.id, status: ($(`input[name=st_${s.id}]:checked`) || {}).value || 'hadir' }));
  el.disabled = true;
  try { await PUT('/attendance', { course_id: S.course, tanggal: S.absen.tgl, items }); toast('Absensi tersimpan.'); await refresh(true); }
  finally { el.disabled = false; }
};
ACT.absenDel = () => del(`/attendance?course_id=${S.course}&tanggal=${S.absen.tgl}`, `Hapus seluruh absensi tanggal <b>${fmtDate(S.absen.tgl)}</b>?`);

async function guruRekap() {
  const cs = await guruCourses();
  if (!cs.length) return head('Rekap Nilai') + noCourse();
  const c = cs.find(x => x.id === S.course);
  const r = await GET('/rekap?course_id=' + S.course);
  const rows = r.students.map(s => {
    const sc = r.nilai[s.id] || {}; // nilai akhir dihitung server (nilai.go)
    return { s, ...sc };
  });
  S.rekapCSV = toCSV([['No', 'Nama', 'NIS', 'Rata Tugas', 'Rata Kuis', 'Kehadiran (%)', 'Nilai Akhir', 'Predikat'],
    ...rows.map((x, i) => [i + 1, x.s.nama_lengkap, x.s.username, fmtN(x.tugas), fmtN(x.kuis), fmtN(x.hadir), fmtN(x.akhir), predikat(x.akhir)[0]])]);
  S.rekapName = `rekap-${c.nama}-${c.kelas_nama}.csv`.replace(/\s+/g, '-');
  const st = S.settings;
  return head('Rekap Nilai', RUMUS,
      btn('Unduh CSV', 'rekapCSV', {}, 'ghost', 'download') + btn('Cetak', 'print', {}, '', 'print')) + courseBar(cs) +
    `<div class="print-only"><div class="print-head">${st.logo_url ? `<img src="${esc(logoSrc(st))}" alt="">` : ''}<div><b style="font-size:16px">${esc(st.nama_sekolah)}</b><br>${esc(st.alamat || '')}${st.telepon ? ' · Telp. ' + esc(st.telepon) : ''}</div></div>
      <h2 style="text-align:center;margin-bottom:6px">REKAP NILAI</h2>
      <p>Mata pelajaran: <b>${esc(c.nama)}</b> · Kelas: <b>${esc(c.kelas_nama)}</b> · Tahun ajaran: <b>${esc(st.tahun_ajaran || '–')}</b> · Guru: <b>${esc(S.user.nama_lengkap)}</b></p><br></div>
    <p class="muted no-print" style="margin-bottom:12px">${r.assignments.length} tugas · ${r.quizzes.length} kuis · ${r.pertemuan} pertemuan tercatat</p>
    ${rows.length ? `<div class="tw no-stack"><table><thead><tr><th class="n">No</th><th class="fz">Nama</th><th>NIS</th><th class="n">Tugas</th><th class="n">Kuis</th><th class="n">Hadir</th><th class="n">Akhir</th><th>Predikat</th></tr></thead><tbody>
      ${rows.map((x, i) => { const [p, k] = predikat(x.akhir); return `<tr><td class="n">${i + 1}</td><td class="fz"><b>${esc(x.s.nama_lengkap)}</b></td><td>${esc(x.s.username)}</td>
        <td class="n">${fmtN(x.tugas)}</td><td class="n">${fmtN(x.kuis)}</td><td class="n">${x.hadir == null ? '–' : fmtN(x.hadir) + '%'}</td><td class="n"><b>${fmtN(x.akhir)}</b></td><td>${tag(p, k)}</td></tr>`; }).join('')}
    </tbody></table></div>` : `<div class="card">${empty('Belum ada siswa di kelas ini.', 'users')}</div>`}
    <div class="print-only"><div class="print-sign">
      <div>Mengetahui,<br>Kepala Sekolah<br><br><br><br><b><u>${esc(st.kepala_sekolah || '.............................')}</u></b><br>NIP. ${esc(st.nip_kepala || '')}</div>
      <div>${fmtDate(today())}<br>Guru Mata Pelajaran<br><br><br><br><b><u>${esc(S.user.nama_lengkap)}</u></b><br>NIP. ${esc(S.user.nip || '')}</div></div></div>`;
}
ACT.rekapCSV = () => downloadText(S.rekapName, S.rekapCSV);

Object.assign(PAGES.guru, { dash: guruDash, mapel: guruMapel, materi: guruMateri, tugas: guruTugas, kuis: guruKuis, absensi: guruAbsensi, rekap: guruRekap });
