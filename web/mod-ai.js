'use strict';
/* Modul Asisten AI: generator RPM, LKPD, soal, surat, tanya bebas. Penyedia: API apa pun yang kompatibel OpenAI. */

addNav('admin', ['ai', 'Asisten AI', 'spark', 'ai', 'Layanan']);
addNav('guru', ['ai', 'Asisten AI', 'spark', 'ai', 'Pembelajaran']);

// ---------- Markdown sederhana (aman: teks di-escape dulu) ----------
function mdInline(s) {
  return s.replace(/`([^`]+)`/g, '<code>$1</code>').replace(/\*\*([^*]+)\*\*/g, '<b>$1</b>').replace(/(^|[^*])\*([^*\n]+)\*/g, '$1<i>$2</i>');
}
function mdToHtml(src) {
  const raw = String(src ?? '').split('\n'); // esc() tidak menambah/mengurangi baris: indeks sama dengan `lines`
  const lines = esc(src).split('\n');
  const out = [];
  let i = 0;
  while (i < lines.length) {
    const l = lines[i];
    let m;
    // gambar soal dari AI: SVG ditampilkan lewat <img> (skrip di dalamnya tidak jalan), PNG tersimpan lewat /files/
    if (/^\s*```\s*svg/i.test(l)) {
      const buf = []; i++; while (i < lines.length && !/^\s*```/.test(lines[i])) buf.push(raw[i++]);
      out.push(i < lines.length ? `<p><img class="ai-img" alt="Gambar soal" src="data:image/svg+xml;charset=utf-8,${encodeURIComponent(buf.join('\n'))}"></p>`
        : '<p class="muted">Menggambar…</p>');
      i++; continue;
    }
    if ((m = l.trim().match(/^!\[([^\]]*)\]\((\/files\/[a-f0-9]{24}\.(?:png|jpg))\)$/))) { out.push(`<p><img class="ai-img" alt="${m[1] || 'Gambar'}" src="${m[2]}"></p>`); i++; continue; }
    if (/^```/.test(l)) { const buf = []; i++; while (i < lines.length && !/^```/.test(lines[i])) buf.push(lines[i++]); i++; out.push(`<pre>${buf.join('\n')}</pre>`); continue; }
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
      const start = ordered ? parseInt(l, 10) : 1; // daftar bernomor yang terpotong (mis. soal + pilihan) tetap lanjut nomornya
      while (i < lines.length && /^\s*([-*•]|\d+[.)])\s+/.test(lines[i])) items.push(mdInline(lines[i++].replace(/^\s*([-*•]|\d+[.)])\s+/, '')));
      out.push(`<${ordered ? 'ol' : 'ul'}${start > 1 ? ` start="${start}"` : ''}>${items.map(x => `<li>${x}</li>`).join('')}</${ordered ? 'ol' : 'ul'}>`);
      continue;
    }
    if (/^\s*(-{3,}|_{3,})\s*$/.test(l)) { out.push('<hr>'); i++; continue; }
    if (!l.trim()) { i++; continue; }
    const buf = [];
    while (i < lines.length && lines[i].trim() && !/^(#{1,4}\s|\s*```|\s*!\[|\s*([-*•]|\d+[.)])\s+|\s*\|.*\|\s*$)/.test(lines[i])) buf.push(mdInline(lines[i++]));
    if (buf.length) out.push(`<p>${buf.join('<br>')}</p>`); else { out.push(`<p>${mdInline(l)}</p>`); i++; }
  }
  return out.join('');
}

// ---------- Pilihan isian ----------
// Format perencanaan (Kurikulum Merdeka – Pembelajaran Mendalam, Permendikdasmen 13/2025)
const RPP_FORMAT = [['rpm', 'RPM – Rencana Pembelajaran Mendalam (terbaru)'], ['modul', 'Modul Ajar (lengkap)'], ['rpp1', 'RPP satu lembar (ringkas)']];
const MENIT_JP = [['', 'Otomatis sesuai jenjang (SD 35 · SMP 40 · SMA/SMK 45)'], ['35', '35 menit'], ['40', '40 menit'], ['45', '45 menit']];
const MODEL_BELAJAR = ['Problem Based Learning (PBL)', 'Project Based Learning (PjBL)', 'Discovery Learning', 'Inquiry Learning', 'Cooperative Learning',
  'Contextual Teaching and Learning (CTL)', 'Problem Solving', 'Pembelajaran Berdiferensiasi', 'Direct Instruction'].map(x => [x, x]);
const PROFIL_LULUSAN = ['Keimanan dan ketakwaan terhadap Tuhan YME', 'Kewargaan', 'Penalaran kritis', 'Kreativitas', 'Kolaborasi', 'Kemandirian', 'Kesehatan', 'Komunikasi'].map(x => [x, x]);
const KEGIATAN_LKPD = ['Eksperimen / praktikum', 'Diskusi kelompok', 'Studi kasus', 'Observasi', 'Proyek', 'Latihan soal', 'Permainan edukatif', 'Analisis teks/gambar'].map(x => [x, x]);
const JENIS_SURAT = ['Undangan rapat orang tua/wali', 'Undangan rapat guru', 'Pemberitahuan libur', 'Pemberitahuan kegiatan', 'Surat tugas', 'Surat izin kegiatan',
  'Surat keterangan aktif sekolah', 'Surat keterangan pindah', 'Surat rekomendasi', 'Surat pernyataan', 'Surat edaran'];
// ---- Jenjang → kelas → fase → mapel: pilihan dibatasi per jenjang agar AI tidak "liar"
// (mis. materi SMA untuk kelas SD, atau mapel IPAS untuk SMA)
const faseKelas = k => (k <= 2 ? 'A' : k <= 4 ? 'B' : k <= 6 ? 'C' : k <= 9 ? 'D' : k === 10 ? 'E' : 'F');
const kelasJenjang = j => (JENJANG_KELAS[j] || JENJANG_KELAS.SMP).map(k => [`Kelas ${k} (Fase ${faseKelas(k)})`, `Kelas ${k} · Fase ${faseKelas(k)}`]);
// struktur mapel Kurikulum Merdeka per jenjang (saran; mapel lain tetap bisa diketik)
const AGAMA = ['Pendidikan Agama Islam dan Budi Pekerti', 'Pendidikan Agama Kristen dan Budi Pekerti', 'Pendidikan Agama Katolik dan Budi Pekerti',
  'Pendidikan Agama Hindu dan Budi Pekerti', 'Pendidikan Agama Buddha dan Budi Pekerti', 'Pendidikan Agama Khonghucu dan Budi Pekerti'];
const MAPEL_JENJANG = {
  SD: [...AGAMA, 'Pendidikan Pancasila', 'Bahasa Indonesia', 'Matematika', 'Ilmu Pengetahuan Alam dan Sosial', 'Pendidikan Jasmani, Olahraga, dan Kesehatan',
    'Seni Rupa', 'Seni Musik', 'Seni Tari', 'Seni Teater', 'Bahasa Inggris', 'Bahasa Daerah', 'Koding dan Kecerdasan Artifisial'],
  SMP: [...AGAMA, 'Pendidikan Pancasila', 'Bahasa Indonesia', 'Matematika', 'Ilmu Pengetahuan Alam', 'Ilmu Pengetahuan Sosial', 'Bahasa Inggris',
    'Pendidikan Jasmani, Olahraga, dan Kesehatan', 'Informatika', 'Seni Rupa', 'Seni Musik', 'Seni Tari', 'Seni Teater', 'Prakarya', 'Bahasa Daerah', 'Koding dan Kecerdasan Artifisial'],
  SMA: [...AGAMA, 'Pendidikan Pancasila', 'Bahasa Indonesia', 'Matematika', 'Bahasa Inggris', 'Pendidikan Jasmani, Olahraga, dan Kesehatan', 'Sejarah',
    'Ilmu Pengetahuan Alam', 'Ilmu Pengetahuan Sosial', 'Fisika', 'Kimia', 'Biologi', 'Ekonomi', 'Geografi', 'Sosiologi', 'Antropologi', 'Informatika',
    'Matematika Tingkat Lanjut', 'Bahasa Indonesia Tingkat Lanjut', 'Bahasa Inggris Tingkat Lanjut', 'Seni Rupa', 'Seni Musik', 'Prakarya dan Kewirausahaan',
    'Bahasa Daerah', 'Koding dan Kecerdasan Artifisial'],
  SMK: [...AGAMA, 'Pendidikan Pancasila', 'Bahasa Indonesia', 'Matematika', 'Bahasa Inggris', 'Pendidikan Jasmani, Olahraga, dan Kesehatan', 'Sejarah',
    'Seni Budaya', 'Informatika', 'Projek Ilmu Pengetahuan Alam dan Sosial', 'Dasar-dasar Program Keahlian', 'Konsentrasi Keahlian',
    'Projek Kreatif dan Kewirausahaan', 'Mata Pelajaran Pilihan', 'Bahasa Daerah'],
};
// jenjang: satu sumber, Pengaturan → Identitas sekolah (server juga selalu memakai nilai ini)
const jenjangAwal = () => (JENJANG_KELAS[S.settings.jenjang] ? S.settings.jenjang : 'SMP');
// mapel: yang diampu guru di atas (berlabel), lalu saran sesuai jenjang
const mapelJenjang = j => {
  const ampu = (S.aiMapel || []).filter(Array.isArray);
  return [...ampu, ...(MAPEL_JENJANG[j] || MAPEL_UMUM).filter(n => !ampu.some(([a]) => a.toLowerCase() === n.toLowerCase()))];
};
const AI_MAPEL = { n: 'mapel', l: 'Mata pelajaran', req: true, t: 'combo', o: () => mapelJenjang(S.aiJenjang) };
const AI_KELAS = { n: 'kelas', l: 'Kelas', req: true, t: 'select', o: () => kelasJenjang(S.aiJenjang),
  hint: 'Sesuai jenjang sekolah (diatur admin di Pengaturan → Identitas sekolah).' };

// ---------- Template (field untuk tiap jenis dokumen; prompt ada di server) ----------
const AI_TPL = {
  rpm: { judul: 'RPP / RPM / Modul Ajar', desc: 'Perencanaan Kurikulum Merdeka terbaru (Pembelajaran Mendalam): RPM, Modul Ajar, atau RPP satu lembar, lengkap dengan asesmen.', icon: 'layers',
    fields: [{ n: 'format', l: 'Format', t: 'select', o: RPP_FORMAT },
      [AI_KELAS, AI_MAPEL],
      { n: 'materi', l: 'Materi pokok', req: true },
      [{ n: 'jp', l: 'JP per pertemuan', t: 'number', min: 1, max: 6, v: 2 }, { n: 'pertemuan', l: 'Jumlah pertemuan', t: 'number', min: 1, max: 20, v: 1 }],
      { n: 'menit_jp', l: 'Lama 1 JP', t: 'select', o: MENIT_JP, hint: 'Otomatis: SD 35 · SMP 40 · SMA/SMK 45 menit.' },
      { n: 'model', l: 'Model pembelajaran', t: 'select', o: MODEL_BELAJAR, blank: 'Pilih yang paling sesuai (biar AI memilih)' },
      { n: 'profil', l: 'Dimensi profil lulusan (pilih 1–3; kosongkan = AI memilih)', t: 'checks', o: PROFIL_LULUSAN },
      { n: 'tujuan', l: 'Tujuan pembelajaran', t: 'textarea', rows: 2 },
      { n: 'cp', l: 'Capaian Pembelajaran (opsional)', t: 'textarea', rows: 3, ph: 'Tempel CP resmi dari Kepka BSKAP 046/H/KR/2025 untuk fase & mapel ini', hint: 'Bila kosong, AI menulis ringkasan CP yang ditandai "cocokkan dengan dokumen CP resmi".' },
      { n: 'catatan', l: 'Catatan tambahan', t: 'textarea', rows: 2, ph: 'Karakteristik siswa, sarana yang ada…' }] },
  lkpd: { judul: 'LKPD', desc: 'Lembar Kerja Peserta Didik siap cetak dengan ruang jawaban.', icon: 'file',
    fields: [[AI_KELAS, AI_MAPEL], { n: 'materi', l: 'Materi', req: true },
      { n: 'tujuan', l: 'Tujuan pembelajaran', t: 'textarea', rows: 2 }, [{ n: 'kegiatan', l: 'Jenis kegiatan', t: 'select', o: KEGIATAN_LKPD, blank: 'Pilih yang paling sesuai' }, { n: 'jumlah', l: 'Jumlah soal/aktivitas', t: 'number', min: 1, max: 30, v: 5 }],
      { n: 'catatan', l: 'Catatan', t: 'textarea', rows: 2 }] },
  soal: { judul: 'Generator Soal', desc: 'Soal pilihan ganda + kunci + pembahasan. Bisa langsung dimasukkan ke CBT/Kuis.', icon: 'quiz',
    fields: [[AI_KELAS, AI_MAPEL], { n: 'materi', l: 'Materi / indikator', req: true, t: 'textarea', rows: 2 },
      [{ n: 'jumlah', l: 'Jumlah soal', t: 'number', min: 1, max: 50, v: 10 }, { n: 'opsi', l: 'Pilihan jawaban', t: 'select', o: [['4', '4 (A–D)'], ['5', '5 (A–E)'], ['3', '3 (A–C)']] }],
      [{ n: 'tingkat', l: 'Tingkat kesulitan', t: 'select', o: [['mudah', 'Mudah'], ['sedang', 'Sedang'], ['sulit (HOTS)', 'Sulit (HOTS)'], ['campuran', 'Campuran']] },
        { n: 'gambar', l: 'Gambar / grafik soal', t: 'select', o: [['perlu', 'Bila diperlukan materi'], ['banyak', 'Sebanyak mungkin'], ['tidak', 'Tanpa gambar']] }],
      { n: 'catatan', l: 'Catatan', t: 'textarea', rows: 2 }] },
  surat: { judul: 'Surat Sekolah', desc: 'Undangan, pemberitahuan, surat tugas, keterangan, dll. sesuai tata naskah.', icon: 'inbox',
    fields: [{ n: 'jenis', l: 'Jenis surat', req: true, t: 'combo', o: JENIS_SURAT }, { n: 'tujuan', l: 'Ditujukan kepada', ph: 'Bapak/Ibu Orang Tua/Wali Murid Kelas 7' },
      { n: 'isi', l: 'Isi pokok / keperluan', t: 'textarea', rows: 3, req: true }, { n: 'waktu', l: 'Waktu & tempat (jika ada)', ph: 'Sabtu, 15 Nov 2026, 09.00, Aula' }, { n: 'ttd', l: 'Penanda tangan', ph: 'Kosongkan = kepala sekolah' }] },
  prota: { judul: 'Prota & Promes', desc: 'Program tahunan dan program semester lengkap dengan minggu efektif.', icon: 'cal',
    fields: [[AI_KELAS, AI_MAPEL], [{ n: 'tahun', l: 'Tahun ajaran', t: 'combo', o: () => S.aiTahun || [] }, { n: 'jp', l: 'JP per minggu', t: 'number', min: 1, max: 20, v: 4 }],
      { n: 'materi', l: 'Daftar materi / bab (opsional)', t: 'textarea', rows: 3, ph: 'Kosongkan agar disusun dari Capaian Pembelajaran' },
      { n: 'catatan', l: 'Catatan (libur, kegiatan, ujian)', t: 'textarea', rows: 2 }] },
  kktp: { judul: 'KKTP', desc: 'Kriteria Ketercapaian Tujuan Pembelajaran: deskripsi, rubrik, atau interval nilai.', icon: 'check',
    fields: [[AI_KELAS, AI_MAPEL], { n: 'materi', l: 'Materi / lingkup materi', req: true },
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
  const [st, pilihan] = await Promise.all([GET('/ai/status'), tpl ? Promise.all([mapelOpts(), [], tahunAjaranOpts()]) : null]);
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
  S.aiJenjang = jenjangAwal();
  const id = registerForm(x.fields, async v => aiRun(tpl, v));
  return hd + t + `<div style="margin-bottom:10px">${btn('Semua jenis', 'go', { page: 'ai', params: '{}' }, 'ghost sm', 'back')}</div>` + info +
    `<div class="cols cols-ai"><section class="card"><div class="card-head"><h3>${esc(x.judul)}</h3></div>
      ${formHtml(id, x.fields, { mapel: mapelAwal() }, { submit: 'Buat dengan AI', cancel: false })}</section>
    <section class="card"><div class="card-head"><h3>Hasil</h3><div class="acts" id="aiActs"></div></div><div id="aiOut" class="md">${empty('Isi formulir lalu klik "Buat dengan AI".', 'spark')}</div></section></div>`;
}

// ---- Pemeriksa alokasi waktu RPP: tiap pertemuan wajib punya baris
// "Rincian waktu: awal A menit + inti B menit + penutup C menit = T menit" (diminta di prompt server, rpp.go).
function rppTarget(inp) {
  const kelas = String(inp.kelas || '');
  let t = 0, fase = (kelas.match(/fase\s+([a-f])/i) || [])[1];
  if (!fase) {
    // sama dengan server (tingkatRe di guards.go): "7A", "VIII B", "10 IPA"; selain itu angka pertama ("Kelas 10")
    const m = kelas.toUpperCase().match(/^\s*(\d{1,2}|XII|XI|X|IX|VIII|VII|VI|V|IV|III|II|I)(?:[^A-Za-z0-9]|[A-Z]\b|$|[A-Z]\d|\s)/);
    if (m) t = /^\d/.test(m[1]) ? Number(m[1]) : ({ I: 1, II: 2, III: 3, IV: 4, V: 5, VI: 6, VII: 7, VIII: 8, IX: 9, X: 10, XI: 11, XII: 12 })[m[1]];
    else t = Number((kelas.match(/\d+/) || [0])[0]);
  }
  fase = fase ? fase.toUpperCase() : t ? (t <= 2 ? 'A' : t <= 4 ? 'B' : t <= 6 ? 'C' : t <= 9 ? 'D' : t === 10 ? 'E' : 'F') : '';
  const clamp = (v, d, lo, hi) => { const n = parseInt(v, 10); return n >= lo && n <= hi ? n : d; };
  const menit = clamp(inp.menit_jp, 0, 20, 60) || ('ABC'.includes(fase) && fase ? 35 : fase === 'D' ? 40 : 45);
  return { total: clamp(inp.jp, 2, 1, 6) * menit, pertemuan: clamp(inp.pertemuan, 1, 1, 20) };
}
function rppCekHtml(text, inp) {
  if (!inp) return '';
  const { total, pertemuan } = rppTarget(inp);
  const rows = [...text.matchAll(/Rincian waktu:\s*awal\s*(\d+)\s*menit\s*\+\s*inti\s*(\d+)\s*menit\s*\+\s*penutup\s*(\d+)\s*menit\s*=\s*(\d+)\s*menit/gi)]
    .map(m => ({ a: +m[1], b: +m[2], c: +m[3], t: +m[4] }));
  const salah = rows.map((r, i) => ({ ...r, i: i + 1 })).filter(r => r.a + r.b + r.c !== total || r.t !== total);
  if (rows.length >= pertemuan && !salah.length)
    return `<div class="hint ok" role="status">${ico('check')} Alokasi waktu sesuai: ${pertemuan} pertemuan × ${total} menit.</div>`;
  const msg = !rows.length ? 'Rincian waktu per pertemuan tidak ditemukan.'
    : [rows.length < pertemuan ? `Rincian waktu hanya untuk ${rows.length} dari ${pertemuan} pertemuan.` : '',
      ...salah.map(r => `Pertemuan ${r.i}: ${r.a} + ${r.b} + ${r.c} = ${r.a + r.b + r.c} menit, seharusnya ${total} menit.`)].filter(Boolean).join(' ');
  return `<div class="hint warn" role="status"><b>Periksa alokasi waktu.</b> ${esc(msg)} Perbaiki lewat tombol <b>Sunting</b> sebelum dicetak.</div>`;
}

async function aiRun(tpl, raw) {
  if (tpl === 'rpm' && Array.isArray(raw.profil) && raw.profil.length > 3) throw new Error('Pilih paling banyak 3 dimensi profil lulusan.');
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
  S.aiDoc = { id: Number(docId) || null, template: tpl, output: body, judul: AI_TPL[tpl].judul, input: inputs };
  out.innerHTML = (tpl === 'rpm' ? rppCekHtml(body, inputs) : '') + mdToHtml(body);
  if (S.aiStatus) S.aiStatus.terpakai++;
  if (aiAdaSvg(body)) { acts.innerHTML = '<span class="muted">Menyimpan gambar soal…</span>'; await aiGambarkan(); }
  acts.innerHTML = aiActions(tpl);
}
const AI_SVG_RE = /^\s*```\s*svg/i;
const aiAdaSvg = t => /^\s*```\s*svg/im.test(t || '');
// Gambar SVG dari AI → PNG di server, lalu dokumen disimpan ulang dengan baris ![Gambar soal](/files/…).
// Sesudahnya cetak, Word, dan impor ke CBT/Kuis memakai gambar yang sama (tidak diunggah berulang).
async function aiGambarkan() {
  const d = S.aiDoc;
  if (!d || !aiAdaSvg(d.output)) return;
  const lines = d.output.split('\n'), out = [];
  let ok = 0, gagal = 0;
  for (let i = 0; i < lines.length; i++) {
    if (!AI_SVG_RE.test(lines[i])) { out.push(lines[i]); continue; }
    let j = i + 1;
    while (j < lines.length && !/^\s*```/.test(lines[j])) j++;
    if (j >= lines.length) { out.push(...lines.slice(i)); break; } // blok tidak tertutup (jawaban AI terpotong): biarkan
    try { out.push(`![Gambar soal](${await uploadFile(await svgToPng(lines.slice(i + 1, j).join('\n')))})`); ok++; }
    catch { out.push(...lines.slice(i, j + 1)); gagal++; }
    i = j;
  }
  if (gagal) toast(`${gagal} gambar dari AI tidak bisa diproses. Periksa lewat Sunting atau tambahkan gambar manual setelah impor.`, 'bad');
  if (!ok || S.aiDoc !== d) return;
  d.output = out.join('\n');
  if (d.id) await PUT('/ai/docs/' + d.id, { judul: d.judul, output: d.output }).catch(e => toast('Gambar belum tersimpan di dokumen: ' + e.message, 'bad'));
  aiRefreshView();
}
// tombol hasil: sunting & unduh Word hanya untuk dokumen yang sudah tersimpan
const aiActions = tpl => btn('Salin', 'aiCopy', {}, 'ghost sm') +
  (S.aiDoc?.id ? btn('Sunting', 'aiEdit', {}, 'ghost sm', 'edit') : '') +
  btn('Cetak', 'aiPrint', {}, 'ghost sm', 'print') +
  (S.aiDoc?.id ? btn('Word', 'aiDocx', {}, 'ghost sm', 'download') : '') +
  (tpl === 'soal' ? btn('Masukkan ke CBT/Kuis', 'aiImport', {}, 'sm', 'upload') : '');
// tampilkan ulang dokumen di mana pun sedang terbuka (halaman hasil atau jendela Riwayat)
function aiRefreshView() {
  const v = $('#aiDocView'); if (v) v.innerHTML = aiDocHtml();
  const o = $('#aiOut'); if (o && !v) o.innerHTML = aiDocHtml();
}

// ---- Sunting: teks Markdown + pratinjau langsung, lalu simpan ulang ke server
ACT.aiEdit = () => {
  const d = S.aiDoc;
  openModal({ title: 'Sunting dokumen', wide: true, body: `<form class="form" id="aiEdForm">
    <label class="f"><span>Judul</span><input name="judul" value="${esc(d.judul)}" required maxlength="200"></label>
    <div class="ai-ed">
      <label class="f"><span>Isi (format Markdown: # judul, - daftar, | tabel |, **tebal**)</span>
        <textarea name="output" id="aiEdTxt" rows="22" spellcheck="true" required>${esc(d.output)}</textarea></label>
      <div class="ai-ed-prev"><span class="muted" style="font-size:13px;font-weight:600">Pratinjau</span><div id="aiEdPrev"></div></div>
    </div>
    <div class="form-err" role="alert"></div>
    <div class="sheet-foot">${btn('Batal', 'closeModal', {}, 'ghost')}<button class="btn" type="button" data-act="aiEdSave">${ico('check')} Simpan</button></div></form>` });
  aiEdPreview();
};
function aiEdPreview() {
  const t = $('#aiEdTxt'), p = $('#aiEdPrev');
  if (t && p) p.innerHTML = (S.aiDoc.template === 'rpm' && S.aiDoc.input?.jp != null ? rppCekHtml(t.value, S.aiDoc.input) : '') + `<div class="md">${mdToHtml(t.value)}</div>`;
}
let aiEdT = null;
document.addEventListener('input', e => { if (e.target.id === 'aiEdTxt') { clearTimeout(aiEdT); aiEdT = setTimeout(aiEdPreview, 250); } });
ACT.aiEdSave = async () => {
  const f = $('#aiEdForm'), err = $('.form-err', f);
  const judul = f.judul.value.trim(), output = f.output.value.trim();
  if (!judul || !output) { err.textContent = 'Judul dan isi wajib diisi.'; return; }
  try { await PUT('/ai/docs/' + S.aiDoc.id, { judul, output }); }
  catch (e) { err.textContent = e.message; return; }
  Object.assign(S.aiDoc, { judul, output });
  closeModal();
  aiRefreshView();
  toast('Perubahan tersimpan.');
};
ACT.aiCopy = async () => { try { await navigator.clipboard.writeText(S.aiDoc.output); toast('Teks disalin.'); } catch { toast('Browser menolak menyalin. Blok teks lalu salin manual.', 'bad'); } };
// ---- Cetak resmi: kop sekolah + isi + tanda tangan (Mengetahui Kepala Sekolah & Guru Mata Pelajaran)
const AI_TTD = ['rpm', 'lkpd', 'soal', 'prota', 'kktp']; // dokumen perangkat ajar; surat sudah punya tanda tangan sendiri
const tglPanjang = d => new Date(d + 'T00:00:00').toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' });
function aiPrintOpts() {
  let tempat = '';
  try { tempat = localStorage.getItem('ttd_tempat') || ''; } catch { /* abaikan */ }
  const s = S.settings, u = S.user;
  return { kop: true, ttd: AI_TTD.includes(S.aiDoc.template), tempat, tanggal: today(),
    guru: u.nama_lengkap || '', nip_guru: u.nip || '', kepsek: s.kepala_sekolah || '', nip_kepsek: s.nip_kepala || '' };
}
ACT.aiPrint = () => aiOutDialog('print');
ACT.aiDocx = () => aiOutDialog('docx');
// pilihan kop & tanda tangan — dipakai untuk cetak maupun unduh Word
function aiOutDialog(mode) {
  const o = aiPrintOpts();
  const fields = [[{ n: 'kop', l: 'Kop sekolah (logo, nama, alamat)', t: 'check' }, { n: 'ttd', l: 'Kolom tanda tangan', t: 'check' }],
    [{ n: 'tempat', l: 'Tempat', ph: 'Nama kota / kecamatan' }, { n: 'tanggal', l: 'Tanggal', t: 'date' }],
    [{ n: 'kepsek', l: 'Kepala sekolah' }, { n: 'nip_kepsek', l: 'NIP kepala sekolah' }],
    [{ n: 'guru', l: 'Guru mata pelajaran' }, { n: 'nip_guru', l: 'NIP guru' }]];
  // bukan openForm: itu memuat ulang halaman setelah simpan (hasil AI yang belum disunting bisa hilang dari layar)
  const id = registerForm(fields, async v => {
    try { localStorage.setItem('ttd_tempat', v.tempat || ''); } catch { /* abaikan */ }
    closeModal();
    if (mode === 'docx') {
      // unduh lewat tautan biasa (cookie sesi ikut), nama file diberikan server
      const q = qs({ kop: v.kop ? 1 : 0, ttd: v.ttd ? 1 : 0, tempat: v.tempat, tanggal: v.tanggal, guru: v.guru, nip_guru: v.nip_guru, kepsek: v.kepsek, nip_kepsek: v.nip_kepsek });
      const a = Object.assign(document.createElement('a'), { href: `/api/ai/docs/${S.aiDoc.id}/docx?${q}`, download: '' });
      document.body.append(a); a.click(); a.remove();
      return;
    }
    setTimeout(() => printDoc(aiPrintHtml(v)), 50); // setelah jendela pilihan tertutup
  });
  openModal({ title: mode === 'docx' ? 'Unduh Word (.docx)' : 'Cetak dokumen', body: formHtml(id, fields, o, { submit: mode === 'docx' ? 'Unduh Word' : 'Cetak' }) });
}
function aiPrintHtml(v) {
  const s = S.settings;
  const nip = n => n ? `NIP. ${esc(n)}` : 'NIP. ……………………';
  const nama = n => `<b><u>${esc(n || '……………………………')}</u></b>`;
  const ttd = v.ttd ? `<table class="ttd"><tr>
      <td>Mengetahui,<br>Kepala ${esc(s.nama_sekolah || 'Sekolah')}<div class="ttd-ruang"></div>${nama(v.kepsek)}<br>${nip(v.nip_kepsek)}</td>
      <td>${esc(v.tempat || '……………')}, ${v.tanggal ? esc(tglPanjang(v.tanggal)) : '……………'}<br>Guru Mata Pelajaran<div class="ttd-ruang"></div>${nama(v.guru)}<br>${nip(v.nip_guru)}</td>
    </tr></table>` : '';
  return (v.kop ? schoolHead() : '') + `<div class="md">${mdToHtml(S.aiDoc.output)}</div>` + ttd;
}

async function aiRiwayat() {
  const ds = await GET('/ai/docs');
  return ds.length ? `<div class="tw"><table><thead><tr><th>Dokumen</th><th>Jenis</th>${S.user.role === 'admin' ? '<th>Model</th>' : ''}<th>Dibuat</th><th></th></tr></thead><tbody>
    ${ds.map(d => `<tr><td><b>${esc(d.judul)}</b></td><td>${tag(AI_TPL[d.template]?.judul || d.template, 'brand')}</td>${S.user.role === 'admin' ? `<td><small>${esc(d.model || '')}</small></td>` : ''}<td>${fmtDT(d.created_at)}</td>
      <td><div class="acts">${btn('Buka', 'aiOpen', { id: d.id }, 'sm')}${iconBtn('trash', 'aiDel', { id: d.id, judul: d.judul }, 'Hapus', 'danger')}</div></td></tr>`).join('')}
    </tbody></table></div>` : `<div class="card">${empty('Belum ada dokumen.', 'spark')}</div>`;
}
ACT.aiOpen = async el => {
  const d = await GET('/ai/docs/' + el.dataset.id);
  let input = {};
  try { input = JSON.parse(d.input || '{}'); } catch { /* dokumen lama */ }
  S.aiDoc = { id: d.id, template: d.template, output: d.output, judul: d.judul, input };
  openModal({ title: d.judul, wide: true, body: `<div class="acts" style="justify-content:flex-start;margin-bottom:12px">${aiActions(d.template)}</div>
    <div id="aiDocView">${aiDocHtml()}</div>` });
  aiGambarkan(); // dokumen lama yang gambarnya belum sempat disimpan
};
// isi dokumen + hasil pemeriksaan alokasi waktu (khusus RPP)
const aiDocHtml = () => (S.aiDoc.template === 'rpm' && S.aiDoc.input?.jp != null ? rppCekHtml(S.aiDoc.output, S.aiDoc.input) : '') + `<div class="md">${mdToHtml(S.aiDoc.output)}</div>`;
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
      const soal = await soalPayload(qs, L);
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
