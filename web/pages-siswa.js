'use strict';
/* Halaman siswa */

// =====================================================================
// HALAMAN SISWA
// =====================================================================
async function siswaData() {
  const d = await GET('/rekap/me');
  d.cm = Object.fromEntries(d.courses.map(c => [c.id, c]));
  d.sub = Object.fromEntries(d.submissions.map(s => [s.assignment_id, s]));
  d.att = Object.fromEntries(d.attempts.map(a => [a.quiz_id, a]));
  d.done = new Set(d.progress.map(p => p.material_id));
  d.prog = cid => { const ms = d.materials.filter(m => m.course_id === cid); return { total: ms.length, done: ms.filter(m => d.done.has(m.id)).length }; };
  return d;
}
const noKelas = () => `<div class="card">${empty('Akun Anda belum dimasukkan ke kelas. Hubungi admin sekolah.', 'school')}</div>`;

function tugasStatus(a, s) {
  if (s && s.nilai != null) return [`Nilai ${fmtN(s.nilai)}/${fmtN(a.nilai_maks)}`, 'ok'];
  if (s) return ['Terkumpul', 'info'];
  if (a.deadline && new Date(a.deadline) < new Date(nowMs())) return ['Terlambat', 'bad'];
  return ['Belum dikerjakan', 'warn'];
}

async function siswaDash() {
  if (!S.user.kelas_id) return head('Beranda') + noKelas();
  const d = await siswaData();
  const todo = d.assignments.filter(a => !d.sub[a.id]).sort((x, y) => (x.deadline || '9') < (y.deadline || '9') ? -1 : 1);
  const kuis = d.quizzes.filter(q => !(d.att[q.id] && d.att[q.id].selesai) && q.jumlah_soal > 0);
  return head('Beranda', `Halo, ${esc(S.user.nama_lengkap)}! Semangat belajar hari ini.`) +
    `<div class="stats">${stat('book', d.courses.length, 'Mata pelajaran')}${stat('clip', todo.length, 'Tugas belum dikumpulkan')}${stat('quiz', kuis.length, 'Kuis belum dikerjakan')}</div>
    <div class="cols"><section class="card"><div class="card-head"><h3>Tugas yang perlu dikerjakan</h3>${btn('Semua', 'nav', { page: 'tugas' }, 'ghost sm')}</div>
      ${todo.length ? todo.slice(0, 6).map(a => { const [t, k] = tugasStatus(a, null); return `<div class="li"><div><b>${esc(a.judul)}</b><small class="muted">${esc(d.cm[a.course_id]?.nama)}${a.deadline ? ' · tenggat ' + fmtDT(a.deadline) : ''}</small></div>
        ${k === 'bad' ? tag(t, k) : btn('Kerjakan', 'tugasOpen', { id: a.id }, 'sm')}</div>`; }).join('') : empty('Semua tugas sudah dikumpulkan.', 'check')}
      ${kuis.length ? `<div class="card-head" style="margin-top:16px"><h3>Kuis tersedia</h3></div>${kuis.slice(0, 4).map(q => `<div class="li"><div><b>${esc(q.judul)}</b><small class="muted">${esc(d.cm[q.course_id]?.nama)} · ${q.jumlah_soal} soal · ${q.durasi_menit} menit</small></div>${btn(d.att[q.id] ? 'Lanjutkan' : 'Kerjakan', 'kuisStart', { id: q.id, judul: q.judul, durasi: q.durasi_menit, lanjut: d.att[q.id] ? 1 : 0 }, 'sm')}</div>`).join('')}` : ''}</section>
    <section class="card"><div class="card-head"><h3>Progres materi</h3></div>
      ${d.courses.length ? d.courses.map(c => { const p = d.prog(c.id); return `<div class="li" style="cursor:pointer" data-act="go" data-page="course" data-params='${esc(JSON.stringify({ course: c.id }))}'><div style="flex:1"><b>${esc(c.nama)}</b>
        <small class="muted" style="margin-bottom:6px">${p.total ? `${p.done}/${p.total} materi selesai` : 'Belum ada materi'}</small>${p.total ? prog(p.done / p.total * 100) : ''}</div></div>`; }).join('') : empty('Belum ada mapel di kelas Anda.', 'book')}</section></div>`;
}

async function siswaMapel() {
  if (!S.user.kelas_id) return head('Mata Pelajaran') + noKelas();
  const d = await siswaData();
  return head('Mata Pelajaran', 'Pilih mapel untuk membuka materi.') + (d.courses.length ? `<div class="grid">${d.courses.map(c => { const p = d.prog(c.id);
    const tugas = d.assignments.filter(a => a.course_id === c.id && !d.sub[a.id]).length;
    return `<section class="card course-card" data-act="go" data-page="course" data-params='${esc(JSON.stringify({ course: c.id }))}' tabindex="0" role="button">
      <div class="top">${tag(c.kode, 'brand')}${tugas ? tag(tugas + ' tugas', 'warn') : ''}</div><h3>${esc(c.nama)}</h3>
      <small class="muted">${esc(c.guru_nama || 'Guru belum ditentukan')}</small>${p.total ? prog(p.done / p.total * 100) : ''}<small class="muted">${p.total ? `${p.done}/${p.total} materi selesai` : 'Belum ada materi'}</small></section>`; }).join('')}</div>`
    : `<div class="card">${empty('Belum ada mapel di kelas Anda.', 'book')}</div>`);
}

async function siswaCourse(p) {
  const cid = Number(p.course);
  const [d, ms] = await Promise.all([siswaData(), GET('/materials?course_id=' + cid)]);
  const c = d.cm[cid];
  if (!c) throw new Error('Mapel tidak ditemukan.');
  const as = d.assignments.filter(a => a.course_id === cid);
  const qs = d.quizzes.filter(q => q.course_id === cid);
  return `<div class="no-print" style="margin-bottom:10px">${btn('Semua mapel', 'nav', { page: 'mapel' }, 'ghost sm', 'back')}</div>` +
    head(c.nama, `${esc(c.guru_nama || '')}${c.deskripsi ? ' · ' + esc(c.deskripsi) : ''}`) +
    `<div class="cols cols-21"><section><h2 style="margin-bottom:12px">Materi</h2>
      ${ms.length ? ms.map((m, i) => matHtml(m, i, { done: d.done.has(m.id), toggle: true })).join('') : `<div class="card">${empty('Guru belum menambahkan materi.', 'layers')}</div>`}</section>
    <section><div class="card"><div class="card-head"><h3>Tugas</h3></div>
      ${as.length ? as.map(a => { const [t, k] = tugasStatus(a, d.sub[a.id]); return `<div class="li"><div><b>${esc(a.judul)}</b>${tag(t, k)}</div>${btn('Buka', 'tugasOpen', { id: a.id }, 'ghost sm')}</div>`; }).join('') : empty('Belum ada tugas.', 'clip')}</div>
    <div class="card"><div class="card-head"><h3>Kuis</h3></div>
      ${qs.length ? qs.map(q => { const a = d.att[q.id]; return `<div class="li"><div><b>${esc(q.judul)}</b><small class="muted">${q.jumlah_soal} soal · ${q.durasi_menit} menit</small></div>
        ${a && a.selesai ? tag('Skor ' + fmtN(a.skor), 'ok') : q.jumlah_soal ? btn(a ? 'Lanjutkan' : 'Kerjakan', 'kuisStart', { id: q.id, judul: q.judul, durasi: q.durasi_menit, lanjut: a ? 1 : 0 }, 'sm') : tag('Belum ada soal')}</div>`; }).join('') : empty('Belum ada kuis.', 'quiz')}</div></section></div>`;
}
ACT.matDone = async el => {
  await POST(`/materials/${el.dataset.id}/done`, { selesai: el.dataset.v === '1' });
  toast(el.dataset.v === '1' ? 'Materi ditandai selesai.' : 'Tanda selesai dihapus.');
  const open = el.closest('details');
  await refresh(true);
  // buka kembali materi yang sama setelah halaman diperbarui
  if (open) { const idx = [...$$('details.mat')].findIndex(x => x.querySelector(`[data-id="${el.dataset.id}"]`)); if (idx >= 0) $$('details.mat')[idx].open = true; }
};

async function siswaTugas() {
  if (!S.user.kelas_id) return head('Tugas') + noKelas();
  const d = await siswaData();
  S.sdata = d;
  return head('Tugas', 'Kumpulkan sebelum tenggat. Tugas yang sudah dinilai tidak bisa diubah.') +
    (d.assignments.length ? `<div class="tw"><table><thead><tr><th>Tugas</th><th>Mapel</th><th>Tenggat</th><th>Status</th><th></th></tr></thead><tbody>
    ${d.assignments.map(a => { const s = d.sub[a.id]; const [t, k] = tugasStatus(a, s);
      return `<tr><td><b>${esc(a.judul)}</b></td><td>${esc(d.cm[a.course_id]?.nama)}</td><td>${a.deadline ? fmtDT(a.deadline) : '–'}</td><td>${tag(t, k)}</td>
      <td><div class="acts">${btn(s ? 'Lihat' : 'Kerjakan', 'tugasOpen', { id: a.id }, s ? 'ghost sm' : 'sm')}</div></td></tr>`; }).join('')}
    </tbody></table></div>` : `<div class="card">${empty('Belum ada tugas.', 'clip')}</div>`);
}
ACT.tugasOpen = async el => {
  const d = await siswaData();
  const a = d.assignments.find(x => x.id === Number(el.dataset.id));
  if (!a) return;
  const s = d.sub[a.id];
  const info = `<div style="margin-bottom:14px"><p class="muted">${esc(d.cm[a.course_id]?.nama)} · nilai maks ${fmtN(a.nilai_maks)}${a.deadline ? ' · tenggat ' + fmtDT(a.deadline) : ''}</p>
    ${a.deskripsi ? `<div class="card" style="margin-top:10px;white-space:pre-wrap">${esc(a.deskripsi)}</div>` : ''}</div>`;
  if (s && s.nilai != null) {
    openModal({ title: a.judul, body: info + `<div class="stats" style="margin-bottom:12px">${stat('check', fmtN(s.nilai) + '/' + fmtN(a.nilai_maks), 'Nilai Anda')}</div>
      ${s.feedback ? `<div class="hint" style="margin-bottom:12px"><b>Catatan guru:</b> ${esc(s.feedback)}</div>` : ''}
      <h3 style="margin-bottom:6px">Jawaban Anda</h3><div style="white-space:pre-wrap;color:var(--text-2)">${esc(s.jawaban || '–')}</div>
      ${s.link_file ? `<p style="margin-top:8px"><a href="${esc(safeUrl(s.link_file))}" target="_blank" rel="noopener">${ico('link')} File yang dikumpulkan</a></p>` : ''}` });
    return;
  }
  openForm({
    title: a.judul, values: s || {}, intro: info + (s ? `<div class="hint" style="margin-bottom:12px">Terkumpul ${fmtDT(s.submitted_at)}. Anda masih bisa mengubahnya sebelum dinilai.</div>` : ''),
    submit: s ? 'Perbarui jawaban' : 'Kumpulkan', success: 'Tugas terkumpul.',
    fields: [{ n: 'jawaban', l: 'Jawaban', t: 'textarea', rows: 7 }, { n: 'link_file', l: 'Tautan file', t: 'url', ph: 'https://drive.google.com/…', hint: 'Unggah file ke Google Drive lalu tempel tautannya di sini.' }],
    onSubmit: v => {
      if (!v.jawaban && !v.link_file) throw new Error('Isi jawaban atau tautan file terlebih dahulu.');
      return POST(`/assignments/${a.id}/submit`, v);
    },
  });
};

async function siswaKuis() {
  if (!S.user.kelas_id) return head('Kuis') + noKelas();
  const d = await siswaData();
  return head('Kuis', 'Waktu mulai berjalan saat Anda menekan "Kerjakan" dan tetap berjalan walau halaman ditutup.') +
    (d.quizzes.length ? `<div class="tw"><table><thead><tr><th>Kuis</th><th>Mapel</th><th class="n">Soal</th><th class="n">Durasi</th><th>Status</th><th></th></tr></thead><tbody>
    ${d.quizzes.map(q => { const a = d.att[q.id];
      return `<tr><td><b>${esc(q.judul)}</b></td><td>${esc(d.cm[q.course_id]?.nama)}</td><td class="n">${q.jumlah_soal}</td><td class="n">${q.durasi_menit} mnt</td>
      <td>${a && a.selesai ? tag('Skor ' + fmtN(a.skor), 'ok') : a ? tag('Sedang dikerjakan', 'info') : tag('Belum')}</td>
      <td><div class="acts">${a && a.selesai ? '' : q.jumlah_soal ? btn(a ? 'Lanjutkan' : 'Kerjakan', 'kuisStart', { id: q.id, judul: q.judul, durasi: q.durasi_menit, lanjut: a ? 1 : 0 }, 'sm') : tag('Belum ada soal')}</div></td></tr>`; }).join('')}
    </tbody></table></div>` : `<div class="card">${empty('Belum ada kuis.', 'quiz')}</div>`);
}
ACT.kuisStart = async el => {
  if (el.dataset.lanjut !== '1' && !(await confirmBox(`Mulai <b>${esc(el.dataset.judul)}</b>? Waktu ${esc(el.dataset.durasi)} menit langsung berjalan dan kuis hanya bisa dikerjakan satu kali.`, 'Mulai', ''))) return;
  go('ujian', { id: el.dataset.id });
};

async function siswaUjian(p) {
  const qid = Number(p.id);
  const r = await POST(`/quizzes/${qid}/start`);
  const key = 'sch_quiz_' + S.user.id + '_' + qid + '_' + (r.started_at || '');
  const saved = { ...(store.get(key) || {}), ...(r.answers || {}) }; // jawaban di server menang (dikerjakan di HP lain)
  const letters = 'abcde'.slice(0, r.quiz.jumlah_opsi).split('');
  const end = Date.now() + r.sisa_detik * 1000;
  S.ujian = { qid, key, end, total: r.questions.length, sent: false };
  S.after = () => {
    examTimer(S.ujian, { timerSel: '#timer', lowAt: 60, submit: submitUjian });
    updateAnswered();
  };
  return `<div class="quiz-head"><div><b>${esc(r.quiz.judul)}</b><small class="muted" style="display:block" id="answered"></small></div>
      <div style="display:flex;align-items:center;gap:12px"><span class="timer" id="timer" aria-live="off">--:--</span>${btn('Kirim jawaban', 'ujianSubmit', {}, 'sm', 'check')}</div></div>
    <div id="ujian" style="max-width:820px">${r.questions.map((q, i) => `<fieldset class="card q" style="border:1px solid var(--border)"><legend class="sr" style="position:absolute;left:-9999px">Soal ${i + 1}</legend>
      <p class="tanya"><span class="muted">${i + 1}.</span> ${esc(q.pertanyaan)}</p>
      ${letters.map(l => q['opsi_' + l] != null ? `<label class="opt"><input type="radio" name="q_${q.id}" value="${l}" data-change="ujianPick" ${saved[q.id] === l ? 'checked' : ''}><b>${l.toUpperCase()}.</b><span>${esc(q['opsi_' + l])}</span></label>` : '').join('')}
    </fieldset>`).join('')}
    <div style="text-align:center;margin-top:18px">${btn('Kirim jawaban', 'ujianSubmit', {}, '', 'check')}</div></div>`;
}
function ujianAnswers() {
  const out = {};
  $$('#ujian input[type=radio]:checked').forEach(i => { out[i.name.slice(2)] = i.value; });
  return out;
}
function updateAnswered() {
  const n = Object.keys(ujianAnswers()).length;
  const el = $('#answered');
  if (el) el.textContent = `${n} dari ${S.ujian.total} soal dijawab`;
}
// tiap jawaban: simpan di perangkat sekaligus ke server (ditunda 1,2 detik agar klik beruntun cukup sekali kirim)
let kuisSaveT = null;
CHG.ujianPick = () => {
  const ans = ujianAnswers();
  store.set(S.ujian.key, ans);
  updateAnswered();
  clearTimeout(kuisSaveT);
  const qid = S.ujian.qid;
  kuisSaveT = setTimeout(() => POST(`/quizzes/${qid}/save`, { answers: ans }).catch(() => { /* dicoba lagi pada jawaban berikutnya */ }), 1200);
};
ACT.ujianSubmit = async () => {
  const n = Object.keys(ujianAnswers()).length;
  const msg = n < S.ujian.total ? `Masih ada <b>${S.ujian.total - n} soal</b> yang belum dijawab. Tetap kirim?` : 'Kirim semua jawaban sekarang? Jawaban tidak bisa diubah lagi.';
  if (await confirmBox(msg, 'Kirim', '')) await submitUjian(false);
};
function submitUjian(auto) {
  clearTimeout(kuisSaveT); // semua jawaban ikut terkirim
  return examSubmit(S.ujian, { url: `/quizzes/${S.ujian.qid}/submit`, answers: ujianAnswers(), auto, title: 'Jawaban terkirim', back: ['kuis', 'Kembali ke daftar kuis'] });
}

async function siswaNilai() {
  if (!S.user.kelas_id) return head('Nilai & Kehadiran') + noKelas();
  const d = await siswaData();
  const rows = d.courses.map(c => {
    const as = d.assignments.filter(a => a.course_id === c.id);
    const at = d.attempts.filter(a => a.selesai && d.quizzes.some(q => q.id === a.quiz_id && q.course_id === c.id));
    const ab = d.attendance.filter(a => a.course_id === c.id);
    const cnt = { hadir: 0, izin: 0, sakit: 0, alfa: 0 };
    ab.forEach(a => cnt[a.status]++);
    return { c, cnt, ...(d.nilai[c.id] || {}) }; // nilai akhir dihitung server (nilai.go)
  });
  const graded = d.assignments.filter(a => d.sub[a.id]?.nilai != null);
  return head('Nilai & Kehadiran', RUMUS) +
    (rows.length ? `<div class="tw"><table><thead><tr><th>Mapel</th><th class="n">Tugas</th><th class="n">Kuis</th><th class="n">Hadir</th><th class="n">Akhir</th><th>Kehadiran</th></tr></thead><tbody>
    ${rows.map(x => `<tr><td><b>${esc(x.c.nama)}</b></td><td class="n">${fmtN(x.tugas)}</td><td class="n">${fmtN(x.kuis)}</td><td class="n">${x.hadir == null ? '–' : fmtN(x.hadir) + '%'}</td>
      <td class="n"><b>${fmtN(x.akhir)}</b></td><td><small>Hadir ${x.cnt.hadir} · Izin ${x.cnt.izin} · Sakit ${x.cnt.sakit} · Alfa ${x.cnt.alfa}</small></td></tr>`).join('')}
    </tbody></table></div>` : `<div class="card">${empty('Belum ada mapel.', 'chart')}</div>`) +
    `<section class="card" style="margin-top:16px"><div class="card-head"><h3>Tugas yang sudah dinilai</h3></div>
      ${graded.length ? graded.map(a => { const s = d.sub[a.id]; return `<div class="li"><div><b>${esc(a.judul)}</b><small class="muted">${esc(d.cm[a.course_id]?.nama)}${s.feedback ? ' · “' + esc(s.feedback) + '”' : ''}</small></div>${tag(fmtN(s.nilai) + '/' + fmtN(a.nilai_maks), 'ok')}</div>`; }).join('') : empty('Belum ada tugas yang dinilai.', 'clip')}</section>`;
}

Object.assign(PAGES.siswa, { dash: siswaDash, mapel: siswaMapel, course: siswaCourse, tugas: siswaTugas, kuis: siswaKuis, ujian: siswaUjian, nilai: siswaNilai });
