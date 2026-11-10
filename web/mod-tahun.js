'use strict';
/* Tahun ajaran & kenaikan kelas (admin) */

addNav('admin', ['tahun', 'Tahun Ajaran', 'cal', '', 'Data Sekolah']);
PAGE_PARENT.naik = 'tahun';

const ROMAN = ['I', 'II', 'III', 'IV', 'V', 'VI', 'VII', 'VIII', 'IX', 'X', 'XI', 'XII'];
// 7A -> 8A, X IPA 1 -> XI IPA 1, 6 -> 7. Bila tidak dikenali, nama dibiarkan.
function nextClassName(nama) {
  const s = String(nama || '').trim();
  let m = s.match(/^(\d+)(.*)$/);
  if (m) return String(Number(m[1]) + 1) + m[2];
  m = s.match(/^(XII|XI|X|IX|VIII|VII|VI|V|IV|III|II|I)\b(.*)$/i);
  if (m) { const i = ROMAN.indexOf(m[1].toUpperCase()); if (i >= 0 && i + 1 < ROMAN.length) return ROMAN[i + 1] + m[2]; }
  return s;
}
const gradeNum = (k) => { const t = String(k.tingkat || k.nama || '').trim(); const n = parseInt(t, 10); if (!isNaN(n)) return n; const r = ROMAN.indexOf((t.match(/^(XII|XI|X|IX|VIII|VII|VI|V|IV|III|II|I)\b/i) || [''])[0].toUpperCase()); return r >= 0 ? r + 1 : null; };
const nextYearName = n => { const m = String(n || '').match(/^(\d{4})\/(\d{4})$/); return m ? `${Number(m[1]) + 1}/${Number(m[2]) + 1}` : ''; };

async function pageTahun() {
  const ys = await GET('/tahun');
  S.years = Object.fromEntries(ys.map(y => [y.id, y]));
  const aktif = ys.find(y => y.aktif);
  const r = S.naikHasil; S.naikHasil = null;
  const hasil = r ? `<section class="card" style="margin-bottom:14px"><div class="card-head"><h3>Kenaikan kelas selesai · ${esc(r.tahun_baru)} sekarang aktif</h3></div>
    <div class="stats">${stat('users', r.naik, 'Naik kelas')}${stat('school', r.tinggal, 'Tinggal kelas')}${stat('check', r.lulus, 'Lulus')}${stat('out', r.keluar, 'Keluar')}</div>
    <p class="muted">${r.kelas_baru} kelas dibuat/dipakai di tahun baru. Periksa guru pengampu di menu <b>Mata Pelajaran</b>, lalu tambahkan siswa baru (PPDB) ke kelas awal.</p></section>` : '';
  return hasil + head('Tahun Ajaran', 'Setiap kelas, mapel, nilai, dan absensi tersimpan per tahun ajaran. Tahun lalu tetap bisa dibuka sebagai arsip.',
      btn('Tambah tahun ajaran', 'yearForm', {}, 'ghost', 'plus') + (aktif && !aktif.dinaikkan ? btn('Kenaikan kelas', 'go', { page: 'naik', params: '{}' }, '', 'users') : '')) +
    `<div class="tw"><table><thead><tr><th>Tahun ajaran</th><th>Periode</th><th class="n">Kelas</th><th class="n">Mapel</th><th class="n">Siswa</th><th>Status</th><th></th></tr></thead><tbody>
    ${ys.map(y => `<tr><td><b>${esc(y.nama)}</b></td><td>${y.mulai ? fmtDate(y.mulai) : '–'}${y.selesai ? ' – ' + fmtDate(y.selesai) : ''}</td>
      <td class="n">${y.jumlah_kelas}</td><td class="n">${y.jumlah_mapel}</td><td class="n">${y.jumlah_siswa}</td>
      <td>${y.aktif ? tag('Aktif', 'ok') : tag('Arsip')}${y.dinaikkan ? ' ' + tag('Sudah kenaikan', 'info') : ''}</td>
      <td><div class="acts">${btn('Kelas', 'go', { page: 'kelas', params: JSON.stringify({ tahun: y.id }) }, 'ghost sm')}${iconBtn('edit', 'yearForm', { id: y.id }, 'Ubah')}
        ${y.aktif ? '' : btn('Jadikan aktif', 'yearActivate', { id: y.id, nama: y.nama }, 'ghost sm')}</div></td></tr>`).join('')}
    </tbody></table></div>
    <div class="hint" style="margin-top:14px"><b>Alur akhir tahun:</b> selesaikan nilai & rapor → buka <b>Kenaikan kelas</b> → atur kelas tujuan dan siswa yang tinggal kelas/lulus →
      jalankan. Tahun ajaran baru otomatis aktif; data tahun lama tetap tersimpan dan bisa dibuka lewat pilihan tahun di halaman Kelas, Mapel, dan Mapel Saya.</div>`;
}
ACT.yearForm = el => {
  const y = el.dataset.id ? S.years[el.dataset.id] : { nama: nextYearName(Object.values(S.years).find(x => x.aktif)?.nama) };
  openForm({
    title: y.id ? 'Ubah tahun ajaran' : 'Tambah tahun ajaran', values: y,
    fields: [{ n: 'nama', l: 'Tahun ajaran', req: true, t: 'select', o: (() => { const n = new Date().getFullYear(); return Array.from({ length: 7 }, (_, i) => `${n - 3 + i}/${n - 2 + i}`).map(x => [x, x]); })() }, [{ n: 'mulai', l: 'Mulai', t: 'date' }, { n: 'selesai', l: 'Selesai', t: 'date' }]],
    onSubmit: v => y.id ? PUT('/tahun/' + y.id, v) : POST('/tahun', v),
  });
};
ACT.yearActivate = async el => {
  if (!(await confirmBox(`Jadikan <b>${esc(el.dataset.nama)}</b> tahun ajaran aktif? Halaman guru dan siswa akan menampilkan kelas tahun ini. Untuk akhir tahun, gunakan <b>Kenaikan kelas</b> agar siswa ikut pindah kelas.`, 'Jadikan aktif', ''))) return;
  await PUT(`/tahun/${el.dataset.id}/aktif`);
  S.settings = await GET('/settings');
  toast('Tahun ajaran aktif diganti.');
  refresh(true);
};

// ---------- Wizard kenaikan kelas ----------
async function pageNaik() {
  const ys = await GET('/tahun');
  const src = ys.find(y => y.aktif);
  if (!src) throw new Error('Belum ada tahun ajaran aktif.');
  if (src.dinaikkan) return head('Kenaikan Kelas') + `<div class="card">${empty('Kenaikan kelas tahun ini sudah dijalankan.', 'check')}</div>`;
  const [ks, siswa] = await Promise.all([GET('/classes?tahun=' + src.id), GET('/users?role=siswa')]);
  const aktifSiswa = siswa.filter(u => u.status === 'aktif');
  // Saran "Lulus" hanya untuk tingkat akhir jenjang (6 SD, 9 SMP, 12 SMA/SMK), atau tingkat tertinggi
  // bila memang ada beberapa tingkat berbeda. Jangan meluluskan semua kelas hanya karena tingkatnya sama.
  const grades = [...new Set(ks.map(gradeNum).filter(n => n != null))];
  const maxGrade = Math.max(...grades);
  const akhir = jenjangDiatur() ? Math.max(...JENJANG_KELAS[jenjangDiatur()]) : null; // tingkat akhir jenjang sekolah
  const isTop = g => g != null && (akhir ? g === akhir : ([6, 9, 12].includes(g) || (grades.length > 1 && g === maxGrade)));
  S.naik = { src, ks, siswa: aktifSiswa, kec: {} };
  const rows = ks.map(k => {
    const n = aktifSiswa.filter(u => u.kelas_id === k.id).length;
    const top = isTop(gradeNum(k));
    return `<tr data-kid="${k.id}"><td><b>${esc(k.nama)}</b><small class="muted">${n} siswa</small></td>
      <td><select name="aksi" data-change="naikAksi" aria-label="Aksi ${esc(k.nama)}"><option value="naik" ${top ? '' : 'selected'}>Naik kelas</option><option value="lulus" ${top ? 'selected' : ''}>Lulus</option></select></td>
      <td><input name="ke" value="${esc(top ? '' : nextClassName(k.nama))}" placeholder="Nama kelas tujuan" ${top ? 'disabled' : ''} aria-label="Kelas tujuan ${esc(k.nama)}"></td>
      <td>${n ? btn('Atur siswa', 'naikSiswa', { kid: k.id }, 'ghost sm') : ''}<small class="muted" id="kec-${k.id}"></small></td></tr>`;
  }).join('');
  return `<div style="margin-bottom:10px">${btn('Tahun ajaran', 'nav', { page: 'tahun' }, 'ghost sm', 'back')}</div>` +
    head('Kenaikan Kelas', `Dari tahun ajaran <b>${esc(src.nama)}</b>. Semua kelas yang berisi siswa wajib diatur.`) +
    `<section class="card" style="margin-bottom:14px"><div class="row2" style="display:grid;grid-template-columns:repeat(auto-fit,minmax(180px,1fr));gap:12px">
      <label class="f"><span>Tahun ajaran tujuan</span><input id="nkNama" value="${esc(nextYearName(src.nama))}" placeholder="2027/2028"></label>
      <label class="f"><span>Mulai</span><input id="nkMulai" type="date"></label><label class="f"><span>Selesai</span><input id="nkSelesai" type="date"></label></div>
      <label class="chk" style="margin-top:10px"><input type="checkbox" id="nkSalin" checked> Salin mata pelajaran & guru pengampu ke kelas tujuan (tanpa materi/tugas)</label></section>
    ${ks.length ? `<div class="tw" id="nkTbl"><table><thead><tr><th>Kelas ${esc(src.nama)}</th><th>Aksi</th><th>Kelas tujuan</th><th>Pengecualian</th></tr></thead><tbody>${rows}</tbody></table></div>` : `<div class="card">${empty('Belum ada kelas di tahun ini.', 'school')}</div>`}
    <div class="hint" style="margin-top:12px">Siswa yang <b>tinggal kelas</b> masuk kelas bernama sama di tahun baru. Siswa <b>lulus/keluar</b> dinonaktifkan (data & nilainya tetap tersimpan).
      Proses dijalankan dalam satu langkah dan tidak bisa dibatalkan otomatis; backup database dibuat setiap hari dan sebelum setiap rilis.</div>
    <div class="sheet-foot" style="margin-top:14px">${btn('Jalankan kenaikan kelas', 'naikRun', {}, '', 'check')}</div>`;
}
CHG.naikAksi = el => { const tr = el.closest('tr'); const ke = $('input[name=ke]', tr); const k = S.naik.ks.find(x => x.id === Number(tr.dataset.kid)); ke.disabled = el.value === 'lulus'; ke.value = el.value === 'lulus' ? '' : (ke.value || nextClassName(k.nama)); };
const NK_AKSI = [['', 'Ikut kelas'], ['tinggal', 'Tinggal kelas'], ['naik', 'Naik'], ['lulus', 'Lulus'], ['keluar', 'Keluar/pindah']];
ACT.naikSiswa = el => {
  const kid = Number(el.dataset.kid);
  const k = S.naik.ks.find(x => x.id === kid);
  const list = S.naik.siswa.filter(u => u.kelas_id === kid);
  const id = registerForm([], async (v, form) => {
    $$('select[data-sid]', form).forEach(s => { if (s.value) S.naik.kec[s.dataset.sid] = s.value; else delete S.naik.kec[s.dataset.sid]; });
    const n = list.filter(u => S.naik.kec[u.id]).length;
    $('#kec-' + kid).textContent = n ? ` ${n} pengecualian` : '';
    closeModal();
  });
  openModal({ title: 'Siswa ' + k.nama, wide: true, body: `<form class="form" data-form="${id}"><div class="tw"><table><thead><tr><th>Siswa</th><th>Aksi</th></tr></thead><tbody>
    ${list.map(u => `<tr><td><b>${esc(u.nama_lengkap)}</b><small class="muted">${esc(u.username)}</small></td><td><select data-sid="${u.id}" aria-label="Aksi ${esc(u.nama_lengkap)}">${opts(NK_AKSI, S.naik.kec[u.id] || '')}</select></td></tr>`).join('')}
    </tbody></table></div><div class="form-err" role="alert"></div><div class="sheet-foot"><button type="button" class="btn ghost" data-act="closeModal">Batal</button><button class="btn" type="submit">Simpan</button></div></form>` });
};
ACT.naikRun = async el => {
  const kelas = $$('#nkTbl tr[data-kid]').map(tr => ({ kelas_id: Number(tr.dataset.kid), aksi: $('select[name=aksi]', tr).value, ke_nama: $('input[name=ke]', tr).value.trim() }));
  const tujuan = { nama: $('#nkNama').value.trim(), mulai: $('#nkMulai').value, selesai: $('#nkSelesai').value };
  if (!/^\d{4}\/\d{4}$/.test(tujuan.nama)) return toast('Isi tahun ajaran tujuan, misalnya 2027/2028.', 'bad');
  const kosong = kelas.find(k => k.aksi === 'naik' && !k.ke_nama);
  if (kosong) return toast('Isi nama kelas tujuan untuk semua kelas yang naik.', 'bad');
  const kecuali = Object.entries(S.naik.kec).map(([sid, aksi]) => ({ siswa_id: Number(sid), aksi }));
  const naik = kelas.filter(k => k.aksi === 'naik').length, lulus = kelas.length - naik;
  if (!(await confirmBox(`Jalankan kenaikan kelas ke <b>${esc(tujuan.nama)}</b>? ${naik} kelas naik, ${lulus} kelas lulus, ${kecuali.length} pengecualian siswa. Tahun ajaran baru langsung aktif.`, 'Jalankan', ''))) return;
  el.disabled = true;
  try {
    const r = await POST(`/tahun/${S.naik.src.id}/naik`, { tujuan, kelas, kecuali, salin_mapel: $('#nkSalin').checked });
    S.settings = await GET('/settings');
    S.naikHasil = r; // ditampilkan sebagai kartu di halaman Tahun Ajaran
    renderShell();
    go('tahun');
  } finally { el.disabled = false; }
};

Object.assign(PAGES.admin, { tahun: pageTahun, naik: pageNaik });
