'use strict';
/* Halaman admin + halaman Akun (semua peran) */

// =====================================================================
// HALAMAN ADMIN
// =====================================================================
async function adminDash() {
  const d = await GET('/dashboard');
  const warn = [];
  if (d.tanpa_kelas) warn.push(`<div class="li"><div><b>${d.tanpa_kelas} siswa belum punya kelas</b><small>Siswa tanpa kelas tidak bisa melihat mapel.</small></div>${btn('Atur', 'nav', { page: 'siswa' }, 'ghost sm')}</div>`);
  if (d.tanpa_guru) warn.push(`<div class="li"><div><b>${d.tanpa_guru} mapel belum punya guru</b><small>Mapel tanpa guru tidak bisa diisi materi.</small></div>${btn('Atur', 'nav', { page: 'mapel' }, 'ghost sm')}</div>`);
  if (!S.settings.jenjang_diatur) warn.push(`<div class="li"><div><b>Jenjang sekolah belum diatur</b><small>Atur SD/SMP/SMA/SMK agar pilihan kelas, mapel, kenaikan kelas, e-book, dan Asisten AI sesuai jenjang.</small></div>${btn('Atur', 'go', { page: 'pengaturan', params: JSON.stringify({ tab: 'sekolah' }) }, 'ghost sm')}</div>`);
  const pk = d.penyimpanan || {};
  if (Number(pk.persen) >= 80) warn.push(`<div class="li"><div><b>Penyimpanan file ${esc(pk.persen)}% terpakai</b><small>${fmtBytes(pk.terpakai)} dari ${fmtBytes(pk.batas)}. Bila penuh, upload (foto, berkas PPDB) gagal.</small></div>${btn('Atur', 'nav', { page: 'pengaturan' }, 'ghost sm')}</div>`);
  if (!pk.backup_terakhir) warn.push(`<div class="li"><div><b>Belum ada backup otomatis</b><small>Backup harian dibuat setiap hari setelah pukul 01.00.</small></div>${btn('Lihat', 'nav', { page: 'pengaturan' }, 'ghost sm')}</div>`);
  const steps = [['Isi identitas sekolah', 'pengaturan'], ['Buat kelas', 'kelas'], ['Tambah akun guru', 'guru'], ['Tambah / impor siswa', 'siswa'], ['Buat mata pelajaran & pilih guru', 'mapel']];
  return head('Beranda', `Selamat datang, ${esc(S.user.nama_lengkap)} · tahun ajaran <b>${esc(d.tahun || '–')}</b>`) +
    `<div class="stats">${stat('users', d.siswa, 'Siswa aktif')}${stat('user', d.guru, 'Guru aktif')}${stat('school', d.kelas, 'Kelas')}${stat('book', d.mapel, 'Mata pelajaran')}</div>
    <div class="cols">
      <section class="card"><div class="card-head"><h3>Langkah persiapan</h3></div>
        ${steps.map(([t, p], i) => `<div class="li"><div><b>${i + 1}. ${esc(t)}</b></div>${btn('Buka', 'nav', { page: p }, 'ghost sm')}</div>`).join('')}</section>
      <section class="card"><div class="card-head"><h3>Perlu perhatian</h3></div>
        ${warn.length ? warn.join('') : empty('Semua beres.', 'check')}
        <div class="hint" style="margin-top:12px">Rutin unduh <b>backup database</b> di menu Pengaturan, simpan di flashdisk atau Google Drive.</div></section>
    </div>`;
}

async function adminKelas() {
  const ys = await GET('/tahun');
  const ty = yearPick(ys, S.params.tahun);
  const ks = await GET('/classes?tahun=' + ty.id);
  return head('Kelas', `Rombongan belajar tahun ajaran <b>${esc(ty.nama)}</b>.`, ty.aktif ? btn('Tambah kelas', 'kelasForm', {}, '', 'plus') : '') +
    yearBar(ys, ty, 'kelasTahun') +
    (ks.length ? `<div class="tw"><table><thead><tr><th>Nama kelas</th><th>Tingkat</th><th>Wali kelas</th><th class="n">Siswa</th><th class="n">Mapel</th><th></th></tr></thead><tbody>
    ${ks.map(k => `<tr><td><b>${esc(k.nama)}</b></td><td>${esc(k.tingkat || '–')}</td><td>${k.wali_nama ? esc(k.wali_nama) : '<small class="muted">–</small>'}</td><td class="n">${k.jumlah_siswa}</td><td class="n">${k.jumlah_mapel}</td>
      <td><div class="acts">${iconBtn('users', 'go', { page: 'siswa', params: JSON.stringify({ kelas: k.id }) }, 'Lihat siswa')}${iconBtn('edit', 'kelasForm', { id: k.id, k: JSON.stringify(k) }, 'Ubah')}${iconBtn('trash', 'kelasDel', { id: k.id, nama: k.nama }, 'Hapus', 'danger')}</div></td></tr>`).join('')}
    </tbody></table></div>` : `<div class="card">${empty('Belum ada kelas. Klik "Tambah kelas" untuk memulai.', 'school')}</div>`);
}
CHG.kelasTahun = el => go('kelas', { tahun: el.value });
ACT.kelasForm = async el => {
  const k = el.dataset.k ? JSON.parse(el.dataset.k) : {};
  const gs = (await GET('/users?role=guru')).filter(g => g.status === 'aktif');
  openForm({
    title: k.id ? 'Ubah kelas' : 'Tambah kelas', values: k,
    fields: [{ n: 'nama', l: 'Nama kelas', req: true, ph: 'Contoh: 7A, X IPA 1' }, { n: 'tingkat', l: 'Tingkat', t: 'select', o: tingkatOpts(), blank: '—', hint: 'Kelas baru otomatis masuk tahun ajaran aktif (' + (S.settings.tahun_ajaran || '') + ').' },
      { n: 'wali_id', l: 'Wali kelas', t: 'select', o: gs.map(g => [g.id, g.nama_lengkap]), blank: '— Belum ditentukan —', hint: 'Wali kelas bisa membuka halaman Wali Kelas: rekap kehadiran & poin siswa kelasnya.' }],
    onSubmit: v => k.id ? PUT('/classes/' + k.id, v) : POST('/classes', v),
  });
  // tingkat mengikuti nama kelas (7A → 7, XI IPA 1 → 11); server juga menolak pasangan yang tidak cocok
  const f = $$('.modal form').pop();
  f.nama.addEventListener('input', () => { const t = tingkatDariNama(f.nama.value); if (t) f.tingkat.value = String(t); });
};
const ROMAWI = { I: 1, II: 2, III: 3, IV: 4, V: 5, VI: 6, VII: 7, VIII: 8, IX: 9, X: 10, XI: 11, XII: 12 };
function tingkatDariNama(nama) {
  const m = String(nama).trim().toUpperCase().match(/^(\d{1,2}|XII|XI|X|IX|VIII|VII|VI|V|IV|III|II|I)(?=[^A-Z0-9]|[A-Z](?:\b|\d)|$)/);
  if (!m) return 0;
  const n = /^\d/.test(m[1]) ? Number(m[1]) : ROMAWI[m[1]];
  return n >= 1 && n <= 12 ? n : 0;
}
ACT.kelasDel = el => del('/classes/' + el.dataset.id, `Hapus kelas <b>${esc(el.dataset.nama)}</b>? Semua mapel di kelas ini beserta materi, tugas, kuis, dan absensinya ikut terhapus. Akun siswa tetap ada tetapi tanpa kelas.`);

async function adminUsers(role) {
  const [us, ks] = await Promise.all([GET('/users?role=' + role), GET('/classes')]);
  const fk = S.params.kelas || '';
  // semua baris dirender sekali; saringan kelas berjalan di browser (tanpa memuat ulang daftar dari server)
  const rows = us;
  S.usersById = Object.fromEntries(us.map(u => [u.id, u]));
  if (role === 'siswa') S.after = () => applyKelasFilter(fk);
  const title = role === 'siswa' ? 'Siswa' : 'Guru';
  const filter = role === 'siswa' ? `<select data-change="fKelas" aria-label="Saring kelas"><option value="">Semua kelas</option>${opts(ks.map(k => [k.id, k.nama]), fk)}<option value="none" ${fk === 'none' ? 'selected' : ''}>Belum punya kelas</option></select>` : '';
  const cols = role === 'siswa'
    ? ['Nama', 'Username / NIS', 'Kelas', 'Status']
    : ['Nama', 'Username', 'NIP', 'Mapel', 'Status'];
  return head(title, `${us.length} akun ${role}.`, btn('Impor CSV', 'importForm', { role }, 'ghost', 'upload') + btn('Tambah ' + role, 'userForm', { role }, '', 'plus')) +
    `<div class="bar"><input class="search" type="search" placeholder="Cari nama atau username…" data-filter="tblU" aria-label="Cari">${filter}</div>` +
    (rows.length ? `<div class="tw"><table id="tblU"><thead><tr>${cols.map(c => `<th${c === 'Mapel' ? ' class="n"' : ''}>${c}</th>`).join('')}<th></th></tr></thead><tbody>
    ${rows.map(u => `<tr data-kelas="${u.kelas_id || 'none'}"><td><b>${esc(u.nama_lengkap)}</b>${u.email ? `<small class="muted">${esc(u.email)}</small>` : ''}</td><td>${esc(u.username)}</td>
      ${role === 'siswa' ? `<td>${u.kelas_nama ? esc(u.kelas_nama) : tag('Belum ada', 'warn')}</td>` : `<td>${esc(u.nip || '–')}${(u.akses || []).length ? `<small class="muted">Kelola: ${esc(u.akses.map(a => (STAFF_MODS.find(m => m[0] === a) || [a, a])[1]).join(', '))}</small>` : ''}</td><td class="n">${u.jumlah_mapel}</td>`}
      <td>${u.status === 'aktif' ? tag('Aktif', 'ok') : tag('Nonaktif')}</td>
      <td><div class="acts">${iconBtn('edit', 'userForm', { role, id: u.id }, 'Ubah')}${iconBtn('trash', 'userDel', { id: u.id, nama: u.nama_lengkap }, 'Hapus', 'danger')}</div></td></tr>`).join('')}
    </tbody></table></div><div class="card" id="uKosong" hidden>${empty('Tidak ada siswa di kelas ini.', 'users')}</div>`
      : `<div class="card">${empty(`Belum ada ${role}. Tambah satu per satu atau impor dari file CSV.`, 'users')}</div>`);
}
// saring kelas di browser: baris kelas lain disembunyikan (kelas CSS terpisah dari [hidden] milik kotak cari,
// jadi keduanya bisa dipakai bersamaan)
function applyKelasFilter(fk) {
  let n = 0;
  $$('#tblU tbody tr').forEach(tr => { const off = !!fk && tr.dataset.kelas !== fk; tr.classList.toggle('f-off', off); if (!off) n++; });
  const k = $('#uKosong'); if (k) k.hidden = n > 0;
}
CHG.fKelas = el => {
  S.params.kelas = el.value; // dipakai juga sebagai kelas bawaan saat menambah siswa
  history.replaceState(null, '', '#/siswa' + (el.value ? '?kelas=' + encodeURIComponent(el.value) : '')); // tanpa memuat ulang
  applyKelasFilter(el.value);
};

ACT.userForm = async el => {
  const role = el.dataset.role;
  const u = el.dataset.id ? { ...S.usersById[el.dataset.id] } : { status: 'aktif', kelas_id: S.params.kelas && S.params.kelas !== 'none' ? S.params.kelas : '' };
  const ks = role === 'siswa' ? await GET('/classes') : [];
  const fields = [
    { n: 'nama_lengkap', l: 'Nama lengkap', req: true },
    [{ n: 'username', l: role === 'siswa' ? 'Username / NIS' : 'Username', req: true, hint: 'Dipakai untuk login. Tanpa spasi.' },
      { n: 'password', l: u.id ? 'Password baru' : 'Password', t: 'password', req: !u.id, minlength: 6, ac: 'new-password', hint: u.id ? 'Kosongkan jika tidak diganti.' : 'Minimal 6 karakter.' }],
    role === 'siswa' ? { n: 'kelas_id', l: 'Kelas', t: 'select', o: ks.map(k => [k.id, k.nama]), blank: '— Pilih kelas —' } : { n: 'nip', l: 'NIP' },
    [{ n: 'email', l: 'Email', t: 'email' }, { n: 'status', l: 'Status', t: 'select', o: [['aktif', 'Aktif'], ['nonaktif', 'Nonaktif (tidak bisa login)']] }],
  ];
  if (role === 'guru') fields.push({ n: 'akses', l: 'Tugas tambahan: boleh mengelola modul (seperti admin, hanya modul ini)', t: 'checks', o: STAFF_MODS.filter(([k]) => modOn(k)) });
  openForm({
    title: (u.id ? 'Ubah ' : 'Tambah ') + role, values: u, fields,
    onSubmit: v => {
      const body = { ...v, role, kelas_id: v.kelas_id ? Number(v.kelas_id) : null };
      return u.id ? PUT('/users/' + u.id, body) : POST('/users', body);
    },
  });
};
const STAFF_MODS = [['bukuinduk', 'Buku Induk'], ['absensi', 'Absensi & Kartu'], ['kalender', 'Kalender'], ['ujian', 'Manajemen Ujian'], ['kelulusan', 'Kelulusan'],
  ['sarpras', 'Sarpras'], ['perpus', 'Perpustakaan'], ['ppdb', 'PPDB'], ['surat', 'Surat'], ['bukutamu', 'Buku Tamu'], ['evoting', 'E-Voting'], ['kasir', 'Kasir/Koperasi'], ['admguru', 'BK / Kesiswaan (semua jurnal & catatan siswa)'], ['rapor', 'e-Rapor (semua kelas, mapel & ekstrakurikuler)']];
ACT.userDel = el => del('/users/' + el.dataset.id, `Hapus akun <b>${esc(el.dataset.nama)}</b>? Semua tugas, nilai, dan absensi miliknya ikut terhapus. Untuk sekadar memblokir login, ubah status menjadi Nonaktif.`);

// --- Impor CSV ---
function parseCSV(text) {
  text = text.replace(/^﻿/, '');
  const first = text.split(/\r?\n/)[0] || '';
  const d = (first.match(/;/g) || []).length > (first.match(/,/g) || []).length ? ';' : (first.includes('\t') ? '\t' : ',');
  const rows = []; let row = [], cell = '', q = false;
  for (let i = 0; i < text.length; i++) {
    const c = text[i];
    if (q) { if (c === '"') { if (text[i + 1] === '"') { cell += '"'; i++; } else q = false; } else cell += c; }
    else if (c === '"') q = true;
    else if (c === d) { row.push(cell); cell = ''; }
    else if (c === '\n' || c === '\r') { if (c === '\r' && text[i + 1] === '\n') i++; row.push(cell); rows.push(row); row = []; cell = ''; }
    else cell += c;
  }
  if (cell || row.length) { row.push(cell); rows.push(row); }
  return rows.filter(r => r.some(x => x.trim()));
}
const KEYMAP = { nama: 'nama', namalengkap: 'nama', namasiswa: 'nama', namaguru: 'nama', username: 'username', nis: 'username', nisn: 'username', user: 'username',
  password: 'password', sandi: 'password', katasandi: 'password', kelas: 'kelas', rombel: 'kelas', email: 'email', nip: 'nip' };

ACT.importForm = el => {
  const role = el.dataset.role;
  const cols = role === 'siswa' ? 'nama;nis;password;kelas' : 'nama;username;password;nip';
  const sample = role === 'siswa' ? 'Andi Pratama;1001;siswa1001;7A\nBunga Lestari;1002;siswa1002;7A' : 'Sari Wulandari;sari;guru12345;198701012010012001';
  let parsed = [];
  const id = registerForm([], async () => {
    if (!parsed.length) throw new Error('Pilih file CSV terlebih dahulu.');
    const r = await POST('/users/import', { role, rows: parsed });
    closeModal();
    await afterSave();
    S.importAkun = r.akun || [];
    openModal({ title: 'Hasil impor', body: `<p style="margin-bottom:10px"><b>${r.added}</b> akun berhasil ditambahkan. Pengguna wajib mengganti password saat pertama login.</p>` +
      (S.importAkun.length ? `<div class="hint" style="margin-bottom:10px">Unduh daftar akun untuk dicetak sebagai slip login. Password hanya tampil sekali ini.
        <div style="margin-top:8px">${btn('Unduh daftar akun (CSV)', 'importAkunCSV', {}, 'sm', 'download')}</div></div>` : '') +
      (r.skipped.length ? `<p class="muted" style="margin-bottom:6px">${r.skipped.length} baris dilewati:</p><div class="tw" style="max-height:260px;overflow:auto"><table><tbody>${r.skipped.map(s => `<tr><td>${esc(s)}</td></tr>`).join('')}</tbody></table></div>` : '') +
      `<div class="sheet-foot" style="margin-top:14px"><button class="btn" data-act="closeModal">Selesai</button></div>` });
  });
  openModal({ title: 'Impor ' + role + ' dari CSV', body: `<div class="form" style="gap:12px">
    <div class="hint">Buat tabel di Excel/Google Sheets dengan kolom: <code>${cols.replaceAll(';', ', ')}</code>, lalu simpan sebagai <b>CSV</b>.
      Jika password dikosongkan, dibuatkan password acak (bisa diunduh setelah impor).${role === 'siswa' ? ' Kelas yang belum ada akan dibuat otomatis.' : ''}</div>
    <div>${btn('Unduh contoh file', 'csvSample', { name: 'contoh-' + role + '.csv', text: cols + '\n' + sample }, 'ghost sm', 'download')}</div>
    <label class="f"><span>File CSV</span><input type="file" accept=".csv,text/csv,.txt" data-change="csvPick"></label>
    <div id="csvInfo" class="muted"></div>
    <form class="form" data-form="${id}"><div class="form-err" role="alert"></div><div class="sheet-foot">
      <button type="button" class="btn ghost" data-act="closeModal">Batal</button><button class="btn" type="submit">${ico('upload')} Impor</button></div></form></div>` });
  CHG.csvPick = async inp => {
    const f = inp.files[0];
    if (!f) return;
    const rows = parseCSV(await f.text());
    const hdr = (rows.shift() || []).map(h => KEYMAP[h.toLowerCase().replace(/[^a-z]/g, '')] || '');
    if (!hdr.includes('nama') || !hdr.includes('username')) {
      parsed = [];
      $('#csvInfo').innerHTML = `<span style="color:var(--bad)">Kolom <b>nama</b> dan <b>${role === 'siswa' ? 'nis' : 'username'}</b> tidak ditemukan di baris pertama.</span>`;
      return;
    }
    parsed = rows.map(r => Object.fromEntries(hdr.map((h, i) => [h, (r[i] || '').trim()]).filter(([h]) => h)));
    $('#csvInfo').innerHTML = `${parsed.length} baris siap diimpor. Contoh: <b>${esc(parsed[0]?.nama)}</b> (${esc(parsed[0]?.username)})${parsed[0]?.kelas ? ', kelas ' + esc(parsed[0].kelas) : ''}.`;
  };
};
ACT.csvSample = el => downloadText(el.dataset.name, el.dataset.text);
ACT.importAkunCSV = () => downloadText('akun-baru.csv', toCSV([['Nama', 'Username', 'Password awal', 'Kelas'], ...S.importAkun.map(a => [a.nama, a.username, a.password, a.kelas])]));

async function adminMapel() {
  const ys = await GET('/tahun');
  const ty = yearPick(ys, S.params.tahun);
  const [cs, ks, gs] = await Promise.all([GET('/courses?tahun=' + ty.id), GET('/classes?tahun=' + ty.id), GET('/users?role=guru')]);
  S.mapelRef = { ks, gs };
  const fk = S.params.kelas || '';
  const rows = fk ? cs.filter(c => String(c.kelas_id) === fk) : cs;
  return head('Mata Pelajaran', `Mapel tahun ajaran <b>${esc(ty.nama)}</b>. Setiap mapel terhubung ke satu kelas dan satu guru pengampu.`, ks.length && ty.aktif ? btn('Tambah mapel', 'mapelForm', {}, '', 'plus') : '') +
    yearBar(ys, ty, 'mapelTahun') +
    (!ks.length ? `<div class="card">${empty('Buat kelas terlebih dahulu sebelum menambah mapel.', 'school')}<div style="text-align:center">${btn('Buka menu Kelas', 'nav', { page: 'kelas' }, 'ghost')}</div></div>` :
    `<div class="bar"><input class="search" type="search" placeholder="Cari mapel atau guru…" data-filter="tblM" aria-label="Cari"><select data-change="fKelasM" aria-label="Saring kelas"><option value="">Semua kelas</option>${opts(ks.map(k => [k.id, k.nama]), fk)}</select></div>` +
    (rows.length ? `<div class="tw"><table id="tblM"><thead><tr><th>Kode</th><th>Mata pelajaran</th><th>Kelas</th><th>Guru pengampu</th><th class="n">Siswa</th><th></th></tr></thead><tbody>
    ${rows.map(c => `<tr><td>${tag(c.kode, 'brand')}</td><td><b>${esc(c.nama)}</b></td><td>${esc(c.kelas_nama)}</td><td>${c.guru_nama ? esc(c.guru_nama) : tag('Belum ada', 'warn')}</td><td class="n">${c.jumlah_siswa}</td>
      <td><div class="acts">${iconBtn('edit', 'mapelForm', { c: JSON.stringify(c) }, 'Ubah')}${iconBtn('trash', 'mapelDel', { id: c.id, nama: c.nama + ' ' + c.kelas_nama }, 'Hapus', 'danger')}</div></td></tr>`).join('')}
    </tbody></table></div>` : `<div class="card">${empty('Belum ada mapel.', 'book')}</div>`));
}
CHG.fKelasM = el => go('mapel', { kelas: el.value, tahun: S.params.tahun });
CHG.mapelTahun = el => go('mapel', { tahun: el.value });
ACT.mapelForm = async el => {
  const c = el.dataset.c ? JSON.parse(el.dataset.c) : { kelas_id: S.params.kelas || '' };
  const { ks, gs } = S.mapelRef;
  // pasangan kode–nama dari server (yang sudah dipakai sekolah + daftar umum); server juga memeriksanya saat simpan
  const pairs = (await GET('/mapel/pilihan')).map(p => [p.kode, p.nama]);
  openForm({
    title: c.id ? 'Ubah mapel' : 'Tambah mapel', values: c,
    fields: [[{ n: 'nama', l: 'Nama mapel', req: true, t: 'select', o: [...pairs.map(([, n]) => [n, n]), ['__lain', '+ Mapel lain (ketik baru)…']], blank: '— Pilih mapel —', hint: 'Kode terisi otomatis.' },
      { n: 'kode', l: 'Kode', req: true, ph: 'Otomatis', hint: 'Mengikuti nama mapel.' }],
      { n: 'nama_baru', l: 'Nama mapel baru', ph: 'Mis. Muatan Lokal Batik' },
      [{ n: 'kelas_id', l: 'Kelas', t: 'select', req: true, o: ks.map(k => [k.id, k.nama]), blank: '— Pilih kelas —' },
        { n: 'guru_id', l: 'Guru pengampu', t: 'select', o: gs.filter(g => g.status === 'aktif').map(g => [g.id, g.nama_lengkap]), blank: '— Belum ditentukan —' }],
      { n: 'deskripsi', l: 'Deskripsi', t: 'textarea', rows: 3 }],
    onSubmit: v => {
      const nama = v.nama === '__lain' ? v.nama_baru : v.nama;
      if (!nama) throw new Error('Tulis nama mapel baru.');
      const body = { ...v, nama, kelas_id: Number(v.kelas_id), guru_id: v.guru_id ? Number(v.guru_id) : null };
      delete body.nama_baru;
      return c.id ? PUT('/courses/' + c.id, body) : POST('/courses', body);
    },
  });
  // kode terisi otomatis sesuai nama yang dipilih (selama kode belum diketik manual)
  const f = $$('.modal form').pop();
  const baru = f.nama_baru.closest('.f');
  const sync = () => {
    const lain = f.nama.value === '__lain';
    baru.hidden = !lain; f.nama_baru.required = lain;
    const p = pairs.find(([, n]) => n === f.nama.value);
    f.kode.readOnly = !lain; // kode hanya bisa diketik untuk mapel baru di luar daftar
    if (p) f.kode.value = p[0]; else if (lain && f.kode.dataset.auto !== '0') f.kode.value = '';
    f.kode.closest('.f').querySelector('small').textContent = lain ? 'Tulis kode singkat, mis. MULOK.' : 'Mengikuti nama mapel.';
  };
  f.kode.addEventListener('input', () => { f.kode.dataset.auto = '0'; f.kode.value = f.kode.value.toUpperCase(); });
  f.nama.addEventListener('change', sync); sync();
};
ACT.mapelDel = el => del('/courses/' + el.dataset.id, `Hapus mapel <b>${esc(el.dataset.nama)}</b>? Semua materi, tugas, kuis, nilai, dan absensinya ikut terhapus.`);

// Pengaturan dibagi per bagian (submenu); hanya bagian yang dibuka yang dimuat.
const SET_SECTIONS = [
  ['sekolah', 'Identitas sekolah', 'school', 'Nama, alamat, kepala sekolah, logo'],
  ['akses', 'Hak akses guru', 'key', 'Fitur yang boleh dikelola guru'],
  ['modul', 'Modul aplikasi', 'layers', 'Nyalakan / matikan modul'],
  ['tampilan', 'Warna tema', 'sun', 'Warna sesuai jenjang sekolah'],
  ['absensi', 'Absensi harian', 'cal', 'Jam masuk, selfie & lokasi', 'absensi'],
  ['waktu', 'Zona waktu', 'clock', 'WIB / WITA / WIT'],
  ['ai', 'Asisten AI', 'spark', 'Endpoint, API key, model', 'ai'],
  ['whatsapp', 'Notifikasi WhatsApp', 'inbox', 'Gateway, templat, riwayat kirim'],
  ['penyimpanan', 'Penyimpanan & backup', 'db', 'Masa simpan foto, backup harian'],
];
async function adminPengaturan() {
  const secs = SET_SECTIONS.filter(x => modOn(x[4]));
  const cur = secs.find(x => x[0] === S.params.tab) || secs[0];
  const body = await (SET_PAGES[cur[0]] || SET_PAGES.sekolah)();
  // di HP submenu berupa baris geser: pastikan bagian yang aktif terlihat
  const inner = S.after; // fungsi "after" milik bagian (mis. WhatsApp) tetap dijalankan
  S.after = () => { inner && inner(); const b = $('.subnav .on'); if (b && b.parentElement.scrollWidth > b.parentElement.clientWidth) b.parentElement.scrollLeft = b.offsetLeft - 16; };
  return head('Pengaturan', 'Atur identitas sekolah, modul, tampilan, dan sistem.') +
    `<div class="set-layout"><nav class="subnav" aria-label="Bagian pengaturan">${secs.map(([k, l, ic, d]) =>
      `<button type="button" class="${k === cur[0] ? 'on' : ''}" data-act="go" data-page="pengaturan" data-params='${esc(JSON.stringify({ tab: k }))}' ${k === cur[0] ? 'aria-current="page"' : ''}>
        ${ico(ic)}<span><b>${esc(l)}</b><small>${esc(d)}</small></span></button>`).join('')}</nav>
    <div class="set-body">${body}</div></div>`;
}
const setCard = (title, sub, inner) => `<section class="card"><div class="card-head"><h3>${esc(title)}</h3></div>${sub ? `<p class="muted" style="margin:-4px 0 12px">${sub}</p>` : ''}${inner}</section>`;
// simpan identitas/hak akses: cukup kirim kolom yang berubah; server membalas dengan pengaturan terbaru
async function saveSettings(patch) {
  S.settings = await PUT('/settings', patch); // server mengubah kolom yang dikirim saja & mengembalikan pengaturan terbaru
}
const SET_PAGES = {
  async sekolah() {
    const s = S.settings;
    const fields = [
      { n: 'nama_sekolah', l: 'Nama sekolah', req: true },
      { n: 'alamat', l: 'Alamat', t: 'textarea', rows: 2 },
      [{ n: 'telepon', l: 'Telepon' }, { t: 'html', html: `<div class="f"><span>Tahun ajaran aktif</span><div class="ro-val"><b>${esc(s.tahun_ajaran || '–')}</b>
        ${btn('Ubah di Tahun Ajaran', 'nav', { page: 'tahun' }, 'ghost sm')}</div><small class="muted">Satu sumber untuk seluruh aplikasi: aktifkan tahun di menu Tahun Ajaran.</small></div>` }],
      [{ n: 'kepala_sekolah', l: 'Nama kepala sekolah' }, { n: 'nip_kepala', l: 'NIP kepala sekolah' }],
      { n: 'jenjang', l: 'Jenjang sekolah', t: 'select', req: true, o: [['SD', 'SD / MI'], ['SMP', 'SMP / MTs'], ['SMA', 'SMA / MA'], ['SMK', 'SMK / MAK']],
        hint: 'Semua pilihan yang terkait jenjang mengikuti ini: tingkat kelas, saran mapel, kenaikan kelas, filter e-book, dan Asisten AI (kelas, fase, mapel, lama 1 JP).' },
      { n: 'logo_url', l: 'Logo sekolah', t: 'file', accept: 'image/*', hint: 'Tanpa logo, huruf depan nama sekolah yang tampil.' }];
    const id = registerForm(fields, async v => { await saveSettings(v); toast('Identitas sekolah tersimpan.'); renderShell(); route(); });
    return setCard('Identitas sekolah', 'Tampil di halaman login, kop dokumen cetak, dan tanda tangan laporan.', formHtml(id, fields, s, { cancel: false }));
  },
  async akses() {
    const ak = S.settings.akses_guru || {};
    const keys = [['materi', 'Guru boleh mengelola materi'], ['tugas', 'Guru boleh mengelola tugas'], ['kuis', 'Guru boleh mengelola kuis & soal'],
      ['absensi', 'Guru boleh mengisi absensi mapel'], ['hapus_absen', 'Guru boleh menghapus absensi mapel']];
    const fields = keys.map(([n, l]) => ({ n, l, t: 'check' }));
    const id = registerForm(fields, async v => { await saveSettings({ akses_guru: v }); toast('Hak akses guru tersimpan.'); });
    return setCard('Hak akses guru', 'Matikan fitur yang tidak boleh dikelola guru. Tugas tambahan per guru (BK, perpustakaan, kasir, dll.) diatur di menu <b>Guru</b> → Ubah.',
      formHtml(id, fields, Object.fromEntries(keys.map(([k]) => [k, ak[k] !== false])), { cancel: false }));
  },
  async modul() {
    const MODS = [['lms', 'Pembelajaran (materi, tugas, kuis, rekap nilai)'], ['bukuinduk', 'Buku Induk Siswa'], ['absensi', 'Absensi Harian & Kartu QR'], ['kalender', 'Kalender Akademik'],
      ['ujian', 'Ujian CBT & Manajemen Ujian'], ['kelulusan', 'Pengumuman Kelulusan'], ['sarpras', 'Inventaris Sarpras (KIB)'], ['perpus', 'Perpustakaan'],
      ['ai', 'Asisten AI'], ['ppdb', 'PPDB Online'], ['surat', 'Surat Masuk/Keluar'], ['bukutamu', 'Buku Tamu'], ['evoting', 'E-Voting OSIS'], ['kasir', 'Kasir / Koperasi'], ['admguru', 'Administrasi Guru & Wali Kelas (jurnal, catatan siswa)'], ['rapor', 'e-Rapor & Portofolio Siswa']];
    const fields = MODS.map(([k, l]) => ({ n: k, l, t: 'check' }));
    const id = registerForm(fields, async v => { await PUT('/modules', v); S.settings = await GET('/settings'); await loadModules(); toast('Modul diperbarui.'); renderShell(); route(); });
    return setCard('Modul aplikasi', 'Matikan modul yang tidak dipakai sekolah. Datanya tetap tersimpan dan muncul lagi bila dinyalakan.', formHtml(id, fields, S.settings.modul || {}, { cancel: false }));
  },
  async tampilan() { return warnaCard(); },
  async absensi() {
    const ab = await GET('/config/absen');
    const fields = [[{ n: 'jam_masuk', l: 'Batas jam masuk', t: 'time', hint: 'Lewat jam ini dicatat terlambat.' }, { n: 'jam_pulang', l: 'Jam pulang', t: 'time' }],
      { n: 'selfie', l: 'Izinkan siswa absen selfie + lokasi dari HP', t: 'check' },
      [{ n: 'lat', l: 'Lintang sekolah', t: 'number', step: 'any' }, { n: 'lng', l: 'Bujur sekolah', t: 'number', step: 'any' }],
      { n: 'radius_m', l: 'Radius area sekolah (meter)', t: 'number', min: 20 },
      { t: 'html', html: `<div>${btn('Isi dengan lokasi perangkat ini', 'geoFill', {}, 'ghost sm', 'eye')}</div>` }];
    const id = registerForm(fields, async v => {
      for (const k of ['jam_masuk', 'jam_pulang']) if (v[k] && !/^\d{2}:\d{2}$/.test(v[k])) throw new Error('Format jam harus JJ:MM, contoh 07:15.');
      await PUT('/config/absen', v); S.settings = await GET('/settings'); toast('Pengaturan absensi tersimpan.');
    });
    return setCard('Absensi harian', 'Berlaku untuk absen QR, ceklist, dan selfie.', formHtml(id, fields, ab, { cancel: false }));
  },
  async waktu() { return tzCard(); },
  async ai() {
    return setCard('Asisten AI', 'Isi URL endpoint (kompatibel OpenAI), API key, dan pilih model untuk generator RPM, LKPD, soal, surat, Prota/Promes, dan KKTP. Hanya admin yang bisa melihat pengaturan ini.',
      await aiAtur());
  },
  async penyimpanan() { return storageCard(); },
  whatsapp: waSettings,
};

// ---- Notifikasi WhatsApp (gateway pihak ketiga; token tidak pernah dikirim ke browser)
const WA_EVENTS = [['kirim_masuk', 'Siswa tiba di sekolah (scan QR / selfie)'], ['kirim_pulang', 'Siswa pulang (scan QR)'],
  ['kirim_alfa', 'Siswa alfa (ceklist absensi hari ini)'], ['kirim_ppdb', 'Pendaftar PPDB: nomor pendaftaran & PIN'], ['kirim_catatan', 'Catatan prestasi / pelanggaran siswa']];
const WA_TPL = [['tpl_masuk', 'Tiba di sekolah', '{nama} {kelas} {jam} {status} {sekolah}'], ['tpl_pulang', 'Pulang', '{nama} {kelas} {jam} {sekolah}'],
  ['tpl_alfa', 'Alfa', '{nama} {kelas} {tanggal} {sekolah}'], ['tpl_ppdb', 'PPDB', '{nama} {no_daftar} {pin} {sekolah}'],
  ['tpl_catatan', 'Catatan siswa', '{nama} {kelas} {jenis} {tanggal} {keterangan} {sekolah}']];
async function waSettings() {
  const [c, log] = await Promise.all([GET('/wa/config'), GET('/wa/outbox' + (S.params.st ? '?status=' + encodeURIComponent(S.params.st) : ''))]);
  const gw = [{ n: 'aktif', l: 'Aktifkan notifikasi WhatsApp', t: 'check' },
    [{ n: 'provider', l: 'Gateway', t: 'select', o: [['fonnte', 'Fonnte (api.fonnte.com)'], ['kustom', 'Kustom (Wablas, dll.)']] },
      { n: 'jeda_detik', l: 'Jeda antar pesan (detik)', t: 'number', min: 1, max: 60, hint: 'Jeda mencegah nomor WA sekolah dianggap spam.' }],
    { n: 'url', l: 'URL kirim pesan (gateway kustom)', ph: 'https://<server>.wablas.com/api/send-message' },
    [{ n: 'field_tujuan', l: 'Nama field nomor tujuan', ph: 'phone' }, { n: 'field_pesan', l: 'Nama field pesan', ph: 'message' }],
    { n: 'token', l: 'Token gateway', t: 'password', ac: 'new-password', ph: c.ada_token ? 'Tersimpan — kosongkan bila tidak diubah' : 'Tempel token dari dashboard gateway',
      hint: 'Token hanya disimpan di server dan tidak pernah ditampilkan lagi.' },
    ...(c.ada_token ? [{ n: 'hapus_token', l: 'Hapus token tersimpan', t: 'check' }] : []),
    { t: 'html', html: '<h3 style="margin-top:8px">Kirim notifikasi untuk</h3><p class="muted" style="margin-top:-6px">Nomor tujuan siswa diambil dari <b>No. HP orang tua/wali</b> di Buku Induk.</p>' },
    ...WA_EVENTS.map(([n, l]) => ({ n, l, t: 'check' })),
    { t: 'html', html: '<h3 style="margin-top:8px">Templat pesan</h3><p class="muted" style="margin-top:-6px">Kata dalam kurung kurawal diganti otomatis.</p>' },
    ...WA_TPL.map(([n, l, ph]) => ({ n, l, t: 'textarea', rows: 2, req: true, hint: 'Bisa memakai: ' + ph }))];
  const id = registerForm(gw, async v => { await PUT('/wa/config', v); toast('Pengaturan WhatsApp tersimpan.'); await refresh(true); });
  const tes = [{ n: 'hp', l: 'Nomor WhatsApp tujuan tes', req: true, ph: '08…' }];
  const tid = registerForm(tes, async v => { const r = await POST('/wa/tes', v); toast('Pesan tes terkirim ke ' + r.tujuan + '.'); });
  S.after = () => { // kolom gateway kustom hanya tampil bila dipilih
    const f = $(`form[data-form="${id}"]`); if (!f) return;
    const sync = () => ['url', 'field_tujuan', 'field_pesan'].forEach(n => { const row = f[n]?.closest('.f'); if (row) row.hidden = f.provider.value !== 'kustom'; });
    f.provider.addEventListener('change', sync); sync();
  };
  const ST = { antre: ['Antre', 'warn'], terkirim: ['Terkirim', 'ok'], gagal: ['Gagal', 'bad'] };
  return setCard('Notifikasi WhatsApp', 'Pesan otomatis ke orang tua dan pendaftar lewat gateway WhatsApp pihak ketiga (biaya dari penyedia gateway). Pesan dikirim berurutan dan dicoba ulang otomatis bila gagal.',
      formHtml(id, gw, c, { cancel: false })) +
    setCard('Kirim pesan tes', 'Pastikan token dan gateway benar sebelum diaktifkan.', formHtml(tid, tes, {}, { submit: 'Kirim tes', cancel: false })) +
    setCard('Riwayat pengiriman', `${c.antre} antre · ${c.gagal ? `<b>${c.gagal} gagal</b>` : '0 gagal'} · 200 pesan terakhir`,
      `<div class="bar"><label class="sel">Status <select data-change="waSt"><option value="">Semua</option>${opts(Object.entries(ST).map(([k, [l]]) => [k, l]), S.params.st || '')}</select></label></div>` +
      (log.length ? `<div class="tw" style="max-height:420px;overflow:auto"><table><thead><tr><th>Waktu</th><th>Tujuan</th><th>Pesan</th><th>Status</th><th></th></tr></thead><tbody>
      ${log.map(r => `<tr><td><small>${fmtDT(r.sent_at || r.created_at)}</small></td><td><small>${esc(r.tujuan)}</small><small class="muted">${esc(r.jenis)}</small></td>
        <td><small>${esc(r.pesan)}</small>${r.galat ? `<small style="color:var(--bad)">${esc(r.galat)}</small>` : ''}</td>
        <td>${tag((ST[r.status] || [r.status])[0], (ST[r.status] || [])[1])}${r.percobaan > 1 ? `<small class="muted">${r.percobaan}× coba</small>` : ''}</td>
        <td>${r.status === 'gagal' ? btn('Ulang', 'waUlang', { id: r.id }, 'ghost sm') : ''}</td></tr>`).join('')}</tbody></table></div>` : empty('Belum ada pesan.', 'inbox')));
}
CHG.waSt = el => go('pengaturan', { tab: 'whatsapp', st: el.value });
ACT.waUlang = async el => { await POST(`/wa/outbox/${el.dataset.id}/ulang`); toast('Pesan masuk antrean lagi.'); await refresh(true); };

// ---- Warna tema sekolah: mengikuti warna seragam jenjang
function warnaCard() {
  const cur = WARNA[S.settings.warna] ? S.settings.warna : 'biru';
  return `<section class="card"><div class="card-head"><h3>Warna tema</h3></div>
    <p class="muted" style="margin:-4px 0 10px">Sesuaikan dengan jenjang sekolah. Berlaku untuk semua pengguna, termasuk halaman login.</p>
    <div class="warna-pick">${Object.entries(WARNA).map(([k, [n, j, c]]) => `<label><input type="radio" name="warna" value="${k}" ${k === cur ? 'checked' : ''} data-change="warnaSet">
      <span class="sw" style="background:${c}"></span>${esc(n)} <small class="muted">${esc(j)}</small></label>`).join('')}</div></section>`;
}
CHG.warnaSet = async el => {
  applyWarna(el.value); // pratinjau langsung
  try { await PUT('/config/tampilan', { warna: el.value }); S.settings.warna = el.value; toast('Warna tema tersimpan.'); }
  catch (e) { applyWarna(S.settings.warna); toast(e.message, 'bad'); }
};

// ---- Zona waktu sekolah (dipakai untuk jam terlambat, tanggal absensi, laporan harian)
async function tzCard() {
  const cfg = await GET('/config/waktu');
  const fields = [{ n: 'zona', l: 'Zona waktu', t: 'select', o: [['', 'Ikut server'], ['Asia/Jakarta', 'WIB — Asia/Jakarta'], ['Asia/Makassar', 'WITA — Asia/Makassar'], ['Asia/Jayapura', 'WIT — Asia/Jayapura']],
    hint: 'Batas jam masuk, tanggal absensi, dan laporan harian mengikuti zona ini, walau server berada di zona lain. Perubahan berlaku setelah server dimulai ulang.' }];
  const id = registerForm(fields, async v => { await PUT('/config/waktu', v); toast('Zona waktu tersimpan. Berlaku setelah server dimulai ulang (otomatis saat rilis berikutnya).'); });
  return `<section class="card"><div class="card-head"><h3>Zona waktu sekolah</h3></div>${formHtml(id, fields, cfg, { cancel: false })}</section>`;
}

// ---- Penyimpanan, masa simpan foto, dan backup otomatis
const fmtBytes = n => { n = Number(n || 0); return n >= 1 << 30 ? (n / (1 << 30)).toFixed(2) + ' GB' : n >= 1 << 20 ? (n / (1 << 20)).toFixed(1) + ' MB' : Math.round(n / 1024) + ' KB'; };
async function storageCard() {
  const [d, cfg, bs] = await Promise.all([GET('/dashboard'), GET('/config/penyimpanan'), GET('/backups')]);
  const pk = d.penyimpanan || {};
  const fields = [[{ n: 'selfie_hari', l: 'Simpan foto selfie (hari)', t: 'number', min: 0 }, { n: 'tamu_hari', l: 'Simpan foto buku tamu (hari)', t: 'number', min: 0 }],
    [{ n: 'backup_simpan', l: 'Jumlah backup harian yang disimpan', t: 'number', min: 1 }, { n: 'log_hari', l: 'Simpan log aktivitas (hari)', t: 'number', min: 0 }],
    [{ n: 'ai_hari', l: 'Simpan dokumen Asisten AI (hari)', t: 'number', min: 0 }, { t: 'html', html: '<div></div>' }],
    { t: 'html', html: '<small class="muted">0 hari = tidak pernah dihapus otomatis.</small>' }];
  const id = registerForm(fields, async v => { await PUT('/config/penyimpanan', v); toast('Pengaturan penyimpanan tersimpan.'); });
  return `<section class="card"><div class="card-head"><h3>Penyimpanan & backup</h3></div>
    <p class="muted" style="margin-bottom:6px">File upload: <b>${fmtBytes(pk.terpakai)}</b> dari ${fmtBytes(pk.batas)} (${esc(pk.persen)}%)</p>${prog(Number(pk.persen) || 0)}
    <div style="margin-top:14px">${formHtml(id, fields, cfg, { cancel: false })}</div>
    <h3 style="margin:16px 0 8px">Backup otomatis harian</h3>
    ${bs.length ? `<div class="tw" style="max-height:220px;overflow:auto"><table><tbody>${bs.map(b => `<tr><td>${esc(b.nama)}<small class="muted">${fmtDT(b.waktu)} · ${fmtBytes(b.ukuran)}</small></td>
      <td><a class="btn ghost sm" href="/api/backups/${encodeURIComponent(b.nama)}" download>${ico('download')} Unduh</a></td></tr>`).join('')}</tbody></table></div>`
      : '<p class="muted">Belum ada. Backup dibuat otomatis setiap hari setelah pukul 01.00.</p>'}
    <div class="acts" style="justify-content:flex-start;margin-top:12px"><a class="btn" href="/api/backup" download>${ico('db')} Unduh backup sekarang</a>${btn('Jalankan pemeliharaan', 'maintNow', {}, 'ghost')}</div>
    <div class="hint" style="margin-top:14px">Simpan salinan backup di luar server (flashdisk/Drive) minimal seminggu sekali. Untuk memulihkan: matikan server, ganti
      <code>data/sch.db</code> dengan file backup (ganti nama menjadi <code>sch.db</code>), lalu nyalakan lagi.</div></section>`;
}
ACT.maintNow = async el => {
  if (!(await confirmBox('Jalankan pemeliharaan sekarang? Backup dibuat, foto kedaluwarsa dan file yatim dihapus.', 'Jalankan', ''))) return;
  el.disabled = true;
  try { const r = await POST('/pemeliharaan'); toast('Selesai: ' + r.hasil); await refresh(true); } finally { el.disabled = false; }
};
ACT.geoFill = el => {
  if (!navigator.geolocation) return toast('Perangkat ini tidak mendukung lokasi.', 'bad');
  navigator.geolocation.getCurrentPosition(p => {
    const f = el.closest('form');
    f.lat.value = p.coords.latitude.toFixed(6); f.lng.value = p.coords.longitude.toFixed(6);
    toast(`Lokasi terisi (akurasi ±${Math.round(p.coords.accuracy)} m). Klik Simpan.`);
  }, () => toast('Lokasi tidak terbaca. Izinkan akses lokasi (butuh HTTPS bila bukan di komputer server).', 'bad'), { enableHighAccuracy: true, timeout: 15000 });
};

Object.assign(PAGES.admin, { dash: adminDash, kelas: adminKelas, siswa: () => adminUsers('siswa'), guru: () => adminUsers('guru'), mapel: adminMapel, pengaturan: adminPengaturan });
