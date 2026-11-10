'use strict';
/* Modul Asisten AI: generator RPM, LKPD, soal, surat, tanya bebas. Penyedia: API apa pun yang kompatibel OpenAI. */

addNav('admin', ['ai', 'Asisten AI', 'spark', 'ai', 'Layanan']);
addNav('guru', ['ai', 'Asisten AI', 'spark', 'ai', 'Pembelajaran']);

// ---------- Markdown sederhana (aman: teks di-escape dulu) ----------
function mdInline(s) {
  return s.replace(/`([^`]+)`/g, '<code>$1</code>').replace(/\*\*([^*]+)\*\*/g, '<b>$1</b>').replace(/(^|[^*])\*([^*\n]+)\*/g, '$1<i>$2</i>');
}
function mdToHtml(src) {
  const lines = esc(src).split('\n');
  const out = [];
  let i = 0;
  while (i < lines.length) {
    const l = lines[i];
    if (/^```/.test(l)) { const buf = []; i++; while (i < lines.length && !/^```/.test(lines[i])) buf.push(lines[i++]); i++; out.push(`<pre>${buf.join('\n')}</pre>`); continue; }
    let m;
    if ((m = l.match(/^(#{1,4})\s+(.*)$/))) { const n = Math.min(4, m[1].length + 1); out.push(`<h${n}>${mdInline(m[2])}</h${n}>`); i++; continue; }
    if (/^\s*\|.*\|\s*$/.test(l) && i + 1 < lines.length && /^\s*\|?\s*:?-{2,}/.test(lines[i + 1])) {
      const row = r => r.trim().replace(/^\||\|$/g, '').split('|').map(c => mdInline(c.trim()));
      const hd = row(l); i += 2; const body = [];
      while (i < lines.length && /^\s*\|.*\|\s*$/.test(lines[i])) body.push(row(lines[i++]));
      out.push(`<div class="tw"><table><thead><tr>${hd.map(h => `<th>${h}</th>`).join('')}</tr></thead><tbody>${body.map(r => `<tr>${r.map(c => `<td>${c}</td>`).join('')}</tr>`).join('')}</tbody></table></div>`);
      continue;
    }
    if (/^\s*([-*•]|\d+[.)])\s+/.test(l)) {
      const ordered = /^\s*\d+[.)]/.test(l); const items = [];
      while (i < lines.length && /^\s*([-*•]|\d+[.)])\s+/.test(lines[i])) items.push(mdInline(lines[i++].replace(/^\s*([-*•]|\d+[.)])\s+/, '')));
      out.push(`<${ordered ? 'ol' : 'ul'}>${items.map(x => `<li>${x}</li>`).join('')}</${ordered ? 'ol' : 'ul'}>`);
      continue;
    }
    if (/^\s*(-{3,}|_{3,})\s*$/.test(l)) { out.push('<hr>'); i++; continue; }
    if (!l.trim()) { i++; continue; }
    const buf = [];
    while (i < lines.length && lines[i].trim() && !/^(#{1,4}\s|```|\s*([-*•]|\d+[.)])\s+|\s*\|.*\|\s*$)/.test(lines[i])) buf.push(mdInline(lines[i++]));
    if (buf.length) out.push(`<p>${buf.join('<br>')}</p>`); else { out.push(`<p>${mdInline(l)}</p>`); i++; }
  }
  return out.join('');
}

// ---------- Pilihan isian ----------
const FASE = [['Fase A (Kelas 1–2)'], ['Fase B (Kelas 3–4)'], ['Fase C (Kelas 5–6)'], ['Fase D (Kelas 7–9)'], ['Fase E (Kelas 10)'], ['Fase F (Kelas 11–12)']].map(([x]) => x);
const WAKTU = ['1 x 35 menit', '2 x 35 menit', '1 x 40 menit', '2 x 40 menit', '3 x 40 menit', '1 x 45 menit', '2 x 45 menit', '3 x 45 menit'];
const MODEL_BELAJAR = ['Problem Based Learning (PBL)', 'Project Based Learning (PjBL)', 'Discovery Learning', 'Inquiry Learning', 'Cooperative Learning',
  'Contextual Teaching and Learning (CTL)', 'Problem Solving', 'Pembelajaran Berdiferensiasi', 'Direct Instruction'].map(x => [x, x]);
const PROFIL_LULUSAN = ['Keimanan dan ketakwaan terhadap Tuhan YME', 'Kewargaan', 'Penalaran kritis', 'Kreativitas', 'Kolaborasi', 'Kemandirian', 'Kesehatan', 'Komunikasi'].map(x => [x, x]);
const KEGIATAN_LKPD = ['Eksperimen / praktikum', 'Diskusi kelompok', 'Studi kasus', 'Observasi', 'Proyek', 'Latihan soal', 'Permainan edukatif', 'Analisis teks/gambar'].map(x => [x, x]);
const JENIS_SURAT = ['Undangan rapat orang tua/wali', 'Undangan rapat guru', 'Pemberitahuan libur', 'Pemberitahuan kegiatan', 'Surat tugas', 'Surat izin kegiatan',
  'Surat keterangan aktif sekolah', 'Surat keterangan pindah', 'Surat rekomendasi', 'Surat pernyataan', 'Surat edaran'];
const AI_MAPEL = { n: 'mapel', l: 'Mata pelajaran', req: true, t: 'combo', o: () => S.aiMapel || MAPEL_UMUM };
const AI_KELAS = { n: 'kelas', l: 'Kelas', req: true, t: 'combo', o: () => [...(S.aiKelas || []), ...FASE] };

// ---------- Template (field untuk tiap jenis dokumen; prompt ada di server) ----------
const AI_TPL = {
  rpm: { judul: 'RPM / Modul Ajar', desc: 'Rencana Pembelajaran Mendalam Kurikulum Merdeka, lengkap dengan asesmen.', icon: 'layers',
    fields: [[AI_MAPEL, { n: 'kelas', l: 'Fase / kelas', req: true, t: 'combo', o: () => [...FASE, ...S.aiKelas] }], { n: 'materi', l: 'Materi pokok', req: true },
      [{ n: 'waktu', l: 'Alokasi waktu', t: 'combo', o: WAKTU }, { n: 'pertemuan', l: 'Jumlah pertemuan', t: 'number', min: 1, max: 20, v: 1 }],
      { n: 'model', l: 'Model pembelajaran', t: 'select', o: MODEL_BELAJAR, blank: 'Pilih yang paling sesuai (biar AI memilih)' },
      { n: 'profil', l: 'Dimensi profil lulusan (kosongkan = AI memilih)', t: 'checks', o: PROFIL_LULUSAN },
      { n: 'tujuan', l: 'Tujuan pembelajaran', t: 'textarea', rows: 2 }, { n: 'catatan', l: 'Catatan tambahan', t: 'textarea', rows: 2, ph: 'Karakteristik siswa, sarana yang ada…' }] },
  lkpd: { judul: 'LKPD', desc: 'Lembar Kerja Peserta Didik siap cetak dengan ruang jawaban.', icon: 'file',
    fields: [[AI_MAPEL, AI_KELAS], { n: 'materi', l: 'Materi', req: true },
      { n: 'tujuan', l: 'Tujuan pembelajaran', t: 'textarea', rows: 2 }, [{ n: 'kegiatan', l: 'Jenis kegiatan', t: 'select', o: KEGIATAN_LKPD, blank: 'Pilih yang paling sesuai' }, { n: 'jumlah', l: 'Jumlah soal/aktivitas', t: 'number', min: 1, max: 30, v: 5 }],
      { n: 'catatan', l: 'Catatan', t: 'textarea', rows: 2 }] },
  soal: { judul: 'Generator Soal', desc: 'Soal pilihan ganda + kunci + pembahasan. Bisa langsung dimasukkan ke CBT/Kuis.', icon: 'quiz',
    fields: [[AI_MAPEL, AI_KELAS], { n: 'materi', l: 'Materi / indikator', req: true, t: 'textarea', rows: 2 },
      [{ n: 'jumlah', l: 'Jumlah soal', t: 'number', min: 1, max: 50, v: 10 }, { n: 'opsi', l: 'Pilihan jawaban', t: 'select', o: [['4', '4 (A–D)'], ['5', '5 (A–E)'], ['3', '3 (A–C)']] }],
      { n: 'tingkat', l: 'Tingkat kesulitan', t: 'select', o: [['mudah', 'Mudah'], ['sedang', 'Sedang'], ['sulit (HOTS)', 'Sulit (HOTS)'], ['campuran', 'Campuran']] }, { n: 'catatan', l: 'Catatan', t: 'textarea', rows: 2 }] },
  surat: { judul: 'Surat Sekolah', desc: 'Undangan, pemberitahuan, surat tugas, keterangan, dll. sesuai tata naskah.', icon: 'inbox',
    fields: [{ n: 'jenis', l: 'Jenis surat', req: true, t: 'combo', o: JENIS_SURAT }, { n: 'tujuan', l: 'Ditujukan kepada', ph: 'Bapak/Ibu Orang Tua/Wali Murid Kelas 7' },
      { n: 'isi', l: 'Isi pokok / keperluan', t: 'textarea', rows: 3, req: true }, { n: 'waktu', l: 'Waktu & tempat (jika ada)', ph: 'Sabtu, 15 Nov 2026, 09.00, Aula' }, { n: 'ttd', l: 'Penanda tangan', ph: 'Kosongkan = kepala sekolah' }] },
  prota: { judul: 'Prota & Promes', desc: 'Program tahunan dan program semester lengkap dengan minggu efektif.', icon: 'cal',
    fields: [[AI_MAPEL, AI_KELAS], [{ n: 'tahun', l: 'Tahun ajaran', t: 'combo', o: () => S.aiTahun || [] }, { n: 'jp', l: 'JP per minggu', t: 'number', min: 1, max: 20, v: 4 }],
      { n: 'materi', l: 'Daftar materi / bab (opsional)', t: 'textarea', rows: 3, ph: 'Kosongkan agar disusun dari Capaian Pembelajaran' },
      { n: 'catatan', l: 'Catatan (libur, kegiatan, ujian)', t: 'textarea', rows: 2 }] },
  kktp: { judul: 'KKTP', desc: 'Kriteria Ketercapaian Tujuan Pembelajaran: deskripsi, rubrik, atau interval nilai.', icon: 'check',
    fields: [[AI_MAPEL, AI_KELAS], { n: 'materi', l: 'Materi / lingkup materi', req: true },
      { n: 'tujuan', l: 'Tujuan pembelajaran (opsional)', t: 'textarea', rows: 2 },
      { n: 'pendekatan', l: 'Pendekatan KKTP', t: 'select', o: [['rubrik', 'Rubrik'], ['deskripsi kriteria', 'Deskripsi kriteria'], ['interval nilai', 'Interval nilai'], ['ketiganya (deskripsi, rubrik, interval)', 'Ketiganya']] },
      { n: 'catatan', l: 'Catatan', t: 'textarea', rows: 2 }] },
  bebas: { judul: 'Tanya Bebas', desc: 'Pertanyaan atau permintaan apa saja seputar mengajar dan administrasi.', icon: 'spark',
    fields: [{ n: 'pertanyaan', l: 'Pertanyaan / permintaan', t: 'textarea', rows: 6, req: true, ph: 'Contoh: buatkan ice breaking 5 menit untuk kelas 7 tentang pecahan' }] },
};

async function pageAI() {
  const tab = S.params.tab || 'buat';
  // pengaturan penyedia AI kini ada di menu Pengaturan → Asisten AI (tautan lama tetap diarahkan ke sana)
  if (tab === 'atur' && S.user.role === 'admin') { go('pengaturan', { tab: 'ai' }); return ''; }
  const t = tabs([['buat', 'Buat dokumen'], ['riwayat', 'Riwayat']], tab);
  const hd = head('Asisten AI', 'Buat RPM, LKPD, soal, dan surat dengan AI. Selalu periksa ulang hasilnya sebelum dipakai.');
  if (tab === 'riwayat') return hd + t + await aiRiwayat();
  const tpl = AI_TPL[S.params.jenis] ? S.params.jenis : null;
  // status AI dan daftar pilihan formulir diambil bersamaan (sebelumnya berurutan)
  const [st, pilihan] = await Promise.all([GET('/ai/status'), tpl ? Promise.all([mapelOpts(), kelasOpts(), tahunAjaranOpts()]) : null]);
  S.aiStatus = st;
  // info model hanya untuk admin; guru cukup memakai asisten tanpa detail teknis
  const info = st.siap ? (st.admin ? `<p class="muted" style="margin-bottom:12px">Model: <b>${esc(st.model)}</b></p>`
      : st.kuota ? `<p class="muted" style="margin-bottom:12px">Terpakai hari ini: <b>${st.terpakai}/${st.kuota}</b> dokumen</p>` : '')
    : `<div class="hint" style="margin-bottom:14px">Asisten AI belum diatur. ${st.admin ? `Isi endpoint, API key, dan model di ${btn('Pengaturan → Asisten AI', 'go', { page: 'pengaturan', params: JSON.stringify({ tab: 'ai' }) }, 'ghost sm', 'spark')}` : 'Hubungi admin sekolah.'}</div>`;
  if (!tpl) {
    return hd + t + info + `<div class="grid">${Object.entries(AI_TPL).map(([k, x]) => `<section class="card course-card tpl-card" data-act="go" data-page="ai" data-params='${esc(JSON.stringify({ jenis: k }))}' role="button" tabindex="0">
      <div class="tpl-head"><div class="ic">${ico(x.icon)}</div><b>${esc(x.judul)}</b></div><small class="muted">${esc(x.desc)}</small></section>`).join('')}</div>`;
  }
  const x = AI_TPL[tpl];
  [S.aiMapel, S.aiKelas, S.aiTahun] = pilihan;
  const id = registerForm(x.fields, async v => aiRun(tpl, v));
  return hd + t + `<div style="margin-bottom:10px">${btn('Semua jenis', 'go', { page: 'ai', params: '{}' }, 'ghost sm', 'back')}</div>` + info +
    `<div class="cols cols-ai"><section class="card"><div class="card-head"><h3>${esc(x.judul)}</h3></div>
      ${formHtml(id, x.fields, { mapel: mapelAwal() }, { submit: 'Buat dengan AI', cancel: false })}</section>
    <section class="card"><div class="card-head"><h3>Hasil</h3><div class="acts" id="aiActs"></div></div><div id="aiOut" class="md">${empty('Isi formulir lalu klik "Buat dengan AI".', 'spark')}</div></section></div>`;
}

async function aiRun(tpl, raw) {
  const inputs = Object.fromEntries(Object.entries(raw).map(([k, v]) => [k, v == null ? '' : Array.isArray(v) ? v.join(', ') : String(v)]));
  const out = $('#aiOut'), acts = $('#aiActs');
  out.innerHTML = '<p class="muted">Menghubungi AI…</p>';
  acts.innerHTML = '';
  let res;
  res = await apiFetch('POST', '/ai/generate', { template: tpl, inputs });
  if (!res.ok) { try { await apiJSON(res, '/ai/generate'); } catch (e) { out.innerHTML = empty(e.message, 'x'); throw e; } }
  const reader = res.body.getReader();
  const dec = new TextDecoder();
  let text = '', last = 0;
  for (;;) {
    const { done, value } = await reader.read();
    if (done) break;
    text += dec.decode(value, { stream: true });
    if (Date.now() - last > 120 + text.length / 50) { out.innerHTML = mdToHtml(text.split('\n\u0000DOC:')[0]); last = Date.now(); out.scrollTop = out.scrollHeight; }
  }
  const [body, docId] = text.split('\n\u0000DOC:');
  out.innerHTML = mdToHtml(body);
  S.aiDoc = { id: Number(docId) || null, template: tpl, output: body, judul: AI_TPL[tpl].judul };
  acts.innerHTML = aiActions(tpl);
  if (S.aiStatus) S.aiStatus.terpakai++;
}
const aiActions = tpl => btn('Salin', 'aiCopy', {}, 'ghost sm') + btn('Cetak', 'aiPrint', {}, 'ghost sm', 'print') + (tpl === 'soal' ? btn('Masukkan ke CBT/Kuis', 'aiImport', {}, 'sm', 'upload') : '');
ACT.aiCopy = async () => { try { await navigator.clipboard.writeText(S.aiDoc.output); toast('Teks disalin.'); } catch { toast('Browser menolak menyalin. Blok teks lalu salin manual.', 'bad'); } };
ACT.aiPrint = () => printDoc(`<div class="md">${mdToHtml(S.aiDoc.output)}</div>`);

async function aiRiwayat() {
  const ds = await GET('/ai/docs');
  return ds.length ? `<div class="tw"><table><thead><tr><th>Dokumen</th><th>Jenis</th>${S.user.role === 'admin' ? '<th>Model</th>' : ''}<th>Dibuat</th><th></th></tr></thead><tbody>
    ${ds.map(d => `<tr><td><b>${esc(d.judul)}</b></td><td>${tag(AI_TPL[d.template]?.judul || d.template, 'brand')}</td>${S.user.role === 'admin' ? `<td><small>${esc(d.model || '')}</small></td>` : ''}<td>${fmtDT(d.created_at)}</td>
      <td><div class="acts">${btn('Buka', 'aiOpen', { id: d.id }, 'sm')}${iconBtn('trash', 'aiDel', { id: d.id, judul: d.judul }, 'Hapus', 'danger')}</div></td></tr>`).join('')}
    </tbody></table></div>` : `<div class="card">${empty('Belum ada dokumen.', 'spark')}</div>`;
}
ACT.aiOpen = async el => {
  const d = await GET('/ai/docs/' + el.dataset.id);
  S.aiDoc = { id: d.id, template: d.template, output: d.output, judul: d.judul };
  openModal({ title: d.judul, wide: true, body: `<div class="acts" style="justify-content:flex-start;margin-bottom:12px">${aiActions(d.template)}</div><div class="md">${mdToHtml(d.output)}</div>` });
};
ACT.aiDel = el => del('/ai/docs/' + el.dataset.id, `Hapus dokumen <b>${esc(el.dataset.judul)}</b>?`);

// Masukkan soal hasil AI ke ujian CBT atau kuis mapel
ACT.aiImport = async () => {
  const qs = parseSoalText(S.aiDoc.output);
  if (!qs.length) return toast('Format soal tidak terbaca.', 'bad');
  const targets = [];
  if (modOn('ujian')) (await GET('/cbt').catch(() => [])).forEach(x => targets.push([`cbt:${x.id}:${x.jumlah_opsi}`, `CBT · ${x.judul}`]));
  if (S.user.role === 'guru' && modOn('lms')) {
    // kuis semua mapel diambil bersamaan (sebelumnya satu per satu: guru dengan 10 mapel menunggu 10× lebih lama)
    const cs = await GET('/courses');
    const per = await Promise.all(cs.map(c => GET('/quizzes?course_id=' + c.id).catch(() => [])));
    cs.forEach((c, i) => per[i].forEach(q => targets.push([`kuis:${q.id}:${q.jumlah_opsi}`, `Kuis · ${q.judul} (${c.nama} ${c.kelas_nama})`])));
  }
  if (!targets.length) return toast('Buat dulu ujian CBT atau kuis mapel sebagai tujuan.', 'bad');
  const fields = [{ n: 'tujuan', l: 'Tujuan', t: 'select', o: targets }];
  const id = registerForm(fields, async v => {
      const [kind, tid, nopsi] = v.tujuan.split(':');
      const L = 'abcde'.slice(0, Number(nopsi)).split('');
      const bad = qs.findIndex(q => !q.kunci || !L.includes(q.kunci) || L.some(l => !q.opsi[l]));
      if (bad >= 0) throw new Error(`Soal nomor ${bad + 1} tidak cocok dengan ${L.length} pilihan jawaban tujuan.`);
      const soal = qs.map(q => ({ pertanyaan: q.pertanyaan, jawaban_benar: q.kunci, bobot: 1, ...Object.fromEntries(L.map(l => ['opsi_' + l, q.opsi[l]])) }));
      await POST(kind === 'cbt' ? `/cbt/${tid}/soal/batch` : `/quizzes/${tid}/questions/batch`, { soal });
      closeModal();
      toast(`${qs.length} soal dimasukkan.`); // hasil AI tetap tampil (halaman tidak dimuat ulang)
  });
  openModal({ title: `Masukkan ${qs.length} soal`, body: '<p class="muted" style="margin-bottom:10px">Soal ditambahkan di akhir daftar soal tujuan. Periksa kembali kunci jawabannya.</p>' + formHtml(id, fields, {}, { submit: 'Masukkan' }) });
};

// ---------- Pengaturan penyedia AI (admin; tampil di Pengaturan → Asisten AI) ----------
const AI_PRESETS = [['', '— Pilih contoh penyedia —'], ['https://api.openai.com/v1', 'OpenAI'], ['https://generativelanguage.googleapis.com/v1beta/openai', 'Google Gemini'],
  ['https://api.anthropic.com/v1', 'Anthropic (Claude)'], ['https://openrouter.ai/api/v1', 'OpenRouter'], ['http://localhost:11434/v1', 'Ollama (lokal)']];
async function aiAtur() {
  const c = await GET('/ai/config');
  S.aiModels = c.model ? [c.model] : [];
  const id = registerForm([{ n: 'base_url' }, { n: 'api_key', t: 'password' }, { n: 'hapus_key', t: 'check' }, { n: 'model' }, { n: 'kuota_harian', t: 'number' }, { n: 'suhu', t: 'number' }], async v => {
    if (!v.model) throw new Error('Pilih model dari daftar (klik Refresh model).');
    await PUT('/ai/config', v);
    toast('Pengaturan AI tersimpan.');
    refresh(true);
  });
  return `<form class="form" data-form="${id}" style="max-width:640px">
    <label class="f"><span>Contoh penyedia</span><select data-change="aiPreset">${opts(AI_PRESETS, '')}</select><small class="muted">Hanya mengisi URL. Model tetap dipilih dari daftar penyedia.</small></label>
    <label class="f"><span>URL endpoint (base URL, kompatibel OpenAI)</span><input name="base_url" type="url" value="${esc(c.base_url || '')}" placeholder="https://…/v1" required></label>
    <label class="f"><span>API key</span><input name="api_key" type="password" autocomplete="new-password" placeholder="${c.key_hint ? 'Tersimpan ' + esc(c.key_hint) + ' — kosongkan bila tidak diganti' : 'Tempel API key'}">
      <small class="muted">Disimpan di server sekolah, tidak pernah dikirim ke browser pengguna.</small></label>
    ${c.key_hint ? '<label class="chk"><input type="checkbox" name="hapus_key"> Hapus API key tersimpan</label>' : ''}
    <label class="f"><span>Model</span><div style="display:flex;gap:8px"><select name="model" id="aiModel" style="flex:1">${S.aiModels.length ? opts(S.aiModels.map(m => [m, m]), c.model) : '<option value="">— klik Refresh model —</option>'}</select>
      ${btn('Refresh model', 'aiRefresh', {}, 'ghost', 'refresh')}</div><small class="muted" id="aiModelInfo">Daftar diambil langsung dari endpoint /models penyedia.</small></label>
    <div class="row2"><label class="f"><span>Kuota per guru per hari</span><input name="kuota_harian" type="number" min="0" value="${esc(c.kuota_harian ?? 30)}"><small class="muted">0 = tanpa batas. Admin tidak dibatasi.</small></label>
      <label class="f"><span>Kreativitas (temperature)</span><input name="suhu" type="number" min="0" max="2" step="0.1" value="${esc(c.suhu ?? 0.7)}"></label></div>
    <div class="form-err" role="alert"></div><div class="sheet-foot"><button class="btn" type="submit">Simpan</button></div></form>`;
}
CHG.aiPreset = el => { if (el.value) el.closest('form').base_url.value = el.value; };
ACT.aiRefresh = async el => {
  const f = el.closest('form');
  el.disabled = true;
  $('#aiModelInfo').textContent = 'Mengambil daftar model…';
  try {
    const r = await POST('/ai/models', { base_url: f.base_url.value, api_key: f.api_key.value });
    const cur = f.model.value;
    f.model.innerHTML = opts(r.models.map(m => [m, m]), cur, '— pilih model —');
    $('#aiModelInfo').textContent = `${r.models.length} model tersedia.`;
  } catch (e) { $('#aiModelInfo').textContent = e.message; toast(e.message, 'bad'); }
  finally { el.disabled = false; }
};

Object.assign(PAGES.admin, { ai: pageAI });
Object.assign(PAGES.guru, { ai: pageAI });
