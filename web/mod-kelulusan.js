'use strict';
/* Pengumuman kelulusan: status per siswa (admin) dan halaman cek publik */

addNav('admin', ['kelulusan', 'Kelulusan', 'check', 'kelulusan', 'Ujian']);

// =====================================================================
// KELULUSAN (admin) + halaman publik
// =====================================================================
const GRAD = [['lulus', 'Lulus'], ['tidak_lulus', 'Tidak lulus'], ['ditunda', 'Ditunda']];
async function pageKelulusan() {
  const [ks, cfg] = await Promise.all([GET('/classes'), GET('/config/kelulusan')]);
  // bawaan: kelas tingkat akhir (mis. 9A untuk SMP); "Semua kelas" tetap bisa dipilih
  const akhir = Math.max(0, ...ks.map(k => Number(k.tingkat) || 0));
  const kid = S.params.kelas ?? (ks.find(k => Number(k.tingkat) === akhir)?.id || '');
  const pg = await GETpage('/kelulusan?' + qs({ kelas_id: kid === 'semua' ? '' : kid, q: S.params.q || '', ...pageQ(50) }));
  const rows = pg.rows;
  if (pageOverflow(pg.total, rows.length)) return '';
  const kurang = Number(pg.headers.get('X-Kurang-Data') || 0); // seluruh hasil saringan
  const cid = registerForm([{ n: 'judul' }, { n: 'buka', t: 'datetime' }, { n: 'pesan' }, { n: 'pesan_tidak' }], async v => { await PUT('/config/kelulusan', v); toast('Pengaturan pengumuman tersimpan.'); refresh(true); });
  const url = location.origin + '/#/kelulusan';
  const open = cfg.buka && new Date(cfg.buka) <= new Date(nowMs());
  return head('Pengumuman Kelulusan', 'Siswa/orang tua mengecek hasil dengan NISN dan tanggal lahir (dari Buku Induk), tanpa login.') +
    `<div class="cols"><section class="card"><div class="card-head"><h3>Pengaturan pengumuman</h3>${cfg.buka ? tag(open ? 'Sudah dibuka' : 'Terjadwal', open ? 'ok' : 'info') : tag('Belum dijadwalkan', 'warn')}</div>
      <form class="form" data-form="${cid}">
        <label class="f"><span>Judul</span><input name="judul" value="${esc(cfg.judul || '')}"></label>
        <label class="f"><span>Dibuka pada</span><input type="datetime-local" name="buka" value="${toLocalInput(cfg.buka)}"></label>
        <label class="f"><span>Pesan untuk yang lulus</span><textarea name="pesan" rows="2">${esc(cfg.pesan || '')}</textarea></label>
        <label class="f"><span>Pesan untuk yang tidak lulus / ditunda</span><textarea name="pesan_tidak" rows="2">${esc(cfg.pesan_tidak || '')}</textarea></label>
        <div class="form-err" role="alert"></div><div class="sheet-foot"><button class="btn" type="submit">Simpan</button></div></form></section>
    <section class="card"><div class="card-head"><h3>Alamat halaman pengumuman</h3></div>
      <p><code style="word-break:break-all">${esc(url)}</code></p>
      <div class="acts" style="justify-content:flex-start;margin:10px 0">${btn('Buka halaman', 'gradOpenPublic', {}, 'ghost sm', 'eye')}</div>
      <div style="text-align:center">${qrImg(url, 150)}</div>
      ${kurang ? `<div class="hint" style="margin-top:12px"><b>${kurang} siswa</b> belum punya NISN atau tanggal lahir di Buku Induk, sehingga tidak bisa mengecek.</div>` : ''}</section></div>
    <div class="card-head" style="margin-top:18px"><h2>Status siswa</h2><div class="acts">${btn(pg.total > rows.length ? 'Semua lulus (halaman ini)' : 'Semua lulus', 'gradAll', {}, 'ghost sm')}${btn('Simpan status', 'gradSave', {}, 'sm', 'check')}</div></div>
    <div class="bar"><input class="search" type="search" placeholder="Cari nama, NIS, NISN… (Enter)" value="${esc(S.params.q || '')}" data-srv="kelulusan" aria-label="Cari">
      <select data-change="gradKelas" aria-label="Kelas"><option value="semua">Semua kelas</option>${opts(ks.map(k => [k.id, k.nama]), kid)}</select>
      <small class="muted">${pagerInfo(pg.total, rows.length, 'siswa')}</small></div>
    ${rows.length ? `<div class="tw"><table id="tblG2"><thead><tr><th>Siswa</th><th>Kelas</th><th>NISN / tgl lahir</th><th>Status</th><th>Catatan</th></tr></thead><tbody>
    ${rows.map(r => `<tr data-sid="${r.id}"><td><b>${esc(r.nama_lengkap)}</b><small class="muted">${esc(r.username)}</small></td><td>${esc(r.kelas_nama || '–')}</td>
      <td>${r.nisn && r.tgl_lahir ? `${esc(r.nisn)}<small class="muted">${fmtDate(r.tgl_lahir)}</small>` : tag('Lengkapi Buku Induk', 'warn')}</td>
      <td><select name="st" aria-label="Status">${opts(GRAD, r.status || '', '—')}</select></td><td><input name="cat" value="${esc(r.catatan || '')}" placeholder="Opsional" aria-label="Catatan"></td></tr>`).join('')}
    </tbody></table></div>${pagerNav(pg.total)}` : `<div class="card">${empty('Belum ada siswa.', 'users')}</div>`}`;
}
CHG.gradKelas = el => go('kelulusan', { kelas: el.value, per: S.params.per || '' });
ACT.gradAll = () => $$('#tblG2 select[name=st]').forEach(s => { s.value = 'lulus'; });
ACT.gradSave = async el => {
  const items = $$('#tblG2 tr[data-sid]').map(tr => ({ siswa_id: Number(tr.dataset.sid), status: $('select', tr).value, catatan: $('input[name=cat]', tr).value }));
  el.disabled = true;
  try { await PUT('/kelulusan', { items }); toast('Status kelulusan tersimpan.'); } finally { el.disabled = false; }
};
ACT.gradOpenPublic = () => window.open('/#/kelulusan', '_blank', 'noopener');

PUBLIC_PAGES.kelulusan = async () => {
  let info;
  try { info = await GET('/kelulusan/info'); } catch (e) { info = { error: e.message }; }
  if (S.user || parseHash().page !== 'kelulusan') return; // sudah pindah halaman selama menunggu (WiFi lambat)
  const id = registerForm([{ n: 'nisn' }, { n: 'tgl_lahir' }], async v => {
    const r = await POST('/kelulusan/cek', v);
    const ok = r.status === 'lulus';
    $('#gradOut').innerHTML = `<div class="grad-result"><p class="muted">${esc(r.kelas)}</p><h2>${esc(r.nama)}</h2>
      <div class="badge ${ok ? 'lulus' : 'tidak'}">${ok ? 'LULUS' : r.status === 'ditunda' ? 'DITUNDA' : 'TIDAK LULUS'}</div>
      <p style="color:var(--text-2)">${esc(r.pesan || '')}</p>${r.catatan ? `<p class="muted" style="margin-top:6px">${esc(r.catatan)}</p>` : ''}</div>`;
  });
  const buka = info.buka ? new Date(info.buka) : null;
  const belum = buka && buka > new Date(nowMs());
  $('#app').innerHTML = `<div class="login-wrap"><div style="width:min(460px,100%)">
    ${brandHtml(info.judul || 'Pengumuman Kelulusan')}
    <div class="card" style="padding:26px">
      <h1 style="font-size:24px;margin-bottom:6px">${esc(info.judul || 'Pengumuman Kelulusan')}</h1>
      ${info.error ? errBox({ message: info.error }) : !buka ? '<p class="muted">Pengumuman belum dijadwalkan.</p>' : belum
        ? `<p class="muted" style="margin-bottom:12px">Pengumuman dibuka ${buka.toLocaleString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric', hour: '2-digit', minute: '2-digit' })}</p>
           <div class="timer" id="gradCd" style="font-size:30px;text-align:center"></div>`
        : `<p class="muted" style="margin-bottom:16px">Masukkan NISN dan tanggal lahir sesuai data sekolah.</p>
        <form class="form" data-form="${id}"><label class="f"><span>NISN</span><input name="nisn" required inputmode="numeric" autocomplete="off"></label>
        <label class="f"><span>Tanggal lahir</span><input name="tgl_lahir" type="date" required></label><div class="form-err" role="alert"></div>
        <button class="btn block" type="submit">Lihat hasil</button></form><div id="gradOut"></div>`}
    </div><p class="login-note"><a href="#/">Masuk ke aplikasi</a></p></div></div>`;
  document.title = (info.judul || 'Pengumuman Kelulusan') + ' · ' + (info.nama_sekolah || S.settings.nama_sekolah || '');
  if (belum) {
    const t = setInterval(() => {
      const el = $('#gradCd');
      if (!el) return clearInterval(t);
      const s = Math.max(0, Math.round((buka - nowMs()) / 1000));
      el.textContent = `${Math.floor(s / 86400)} hari ${pad(Math.floor(s / 3600) % 24)}:${pad(Math.floor(s / 60) % 60)}:${pad(s % 60)}`;
      if (!s) { clearInterval(t); PUBLIC_PAGES.kelulusan(); }
    }, 1000);
  }
};

Object.assign(PAGES.admin, { kelulusan: pageKelulusan });
