'use strict';
/* Buku tamu: kiosk publik dengan foto & daftar kunjungan (admin) */

addNav('admin', ['bukutamu', 'Buku Tamu', 'book', 'bukutamu', 'Layanan']);

// =====================================================================
// BUKU TAMU
// =====================================================================
async function pageBukuTamu() {
  const bulan = /^\d{4}-\d{2}$/.test(S.params.bulan || '') ? S.params.bulan : monthNow();
  const rows = await GET('/bukutamu?bulan=' + bulan);
  S.tamu = rows;
  return head('Buku Tamu', 'Tamu mengisi sendiri lewat tablet di meja piket (mode kiosk).') +
    `<div class="cols cols-21"><div><div class="bar"><input type="month" value="${bulan}" data-change="tamuBulan" aria-label="Bulan" style="min-width:0">
      <input class="search" type="search" placeholder="Cari nama, instansi, keperluan…" data-filter="tblT" aria-label="Cari">${btn('Unduh CSV', 'tamuCSV', {}, 'ghost sm', 'download')}</div>
    ${rows.length ? `<div class="tw"><table id="tblT"><thead><tr><th>Waktu</th><th>Tamu</th><th>Keperluan</th><th>Bertemu</th><th></th></tr></thead><tbody>
    ${rows.map(r => `<tr><td>${fmtDT(r.created_at)}</td><td><div class="cell-user">${r.foto ? `<a href="${esc(r.foto)}" target="_blank" rel="noopener"><img class="avatar" src="${esc(r.foto)}" alt=""></a>` : ''}<div><b>${esc(r.nama)}</b><small class="muted">${esc(r.instansi || '')} ${esc(r.hp || '')}</small></div></div></td>
      <td>${esc(r.keperluan)}</td><td>${esc(r.bertemu || '–')}</td><td>${iconBtn('trash', 'tamuDel', { id: r.id }, 'Hapus', 'danger')}</td></tr>`).join('')}</tbody></table></div>`
      : `<div class="card">${empty('Belum ada tamu bulan ini.', 'book')}</div>`}</div>
    ${linkCard(location.origin + '/#/bukutamu', 'Alamat kiosk buku tamu')}</div>`;
}
CHG.tamuBulan = el => go('bukutamu', { bulan: el.value });
ACT.tamuDel = el => del('/bukutamu/' + el.dataset.id, 'Hapus catatan tamu ini?');
ACT.tamuCSV = () => downloadText('buku-tamu.csv', toCSV([['Waktu', 'Nama', 'Instansi', 'HP', 'Keperluan', 'Bertemu'], ...S.tamu.map(r => [fmtDT(r.created_at), r.nama, r.instansi, r.hp, r.keperluan, r.bertemu])]));

PUBLIC_PAGES.bukutamu = () => {
  let foto = '';
  const id = registerForm([], async (v, form) => {
    const d = Object.fromEntries([...new FormData(form)].map(([k, x]) => [k, String(x).trim()]));
    await POST('/bukutamu', { ...d, foto });
    stopCam();
    $('#tamuBox').innerHTML = `<div class="grad-result"><div class="stat" style="justify-content:center;border:0;box-shadow:none;background:none"><div class="ic">${ico('check')}</div></div>
      <h2>Terima kasih, ${esc(d.nama)}!</h2><p class="muted">Silakan menunggu, petugas akan segera membantu.</p></div>`;
    setTimeout(() => { if (!S.user && parseHash().page === 'bukutamu') PUBLIC_PAGES.bukutamu(); }, 6000);
  });
  let stream = null;
  const stopCam = () => { stopStream(stream, $('#tamuV')); stream = null; };
  ACT.tamuCam = async () => {
    try {
      stopCam();
      const vd = $('#tamuV');
      if (vd) vd.hidden = false;
      stream = await openCamera(vd, { facingMode: 'user', width: 640 });
      if (stream) $('#tamuShot').hidden = false;
    } catch (e) { toast(e.message.startsWith('Kamera') || e.message.startsWith('Browser') ? e.message : 'Kamera tidak bisa dibuka.', 'bad'); }
  };
  ACT.tamuShot = () => {
    const vd = $('#tamuV');
    const f = snapshot(vd); if (!f) return;
    foto = f;
    stopCam(); vd.hidden = true; $('#tamuShot').hidden = true; $('#tamuPrev').innerHTML = `<img src="${foto}" alt="" style="width:120px;border-radius:8px">`;
  };
  $('#app').innerHTML = publicShell('Buku Tamu', `<h1 style="font-size:26px;margin-bottom:4px">Selamat datang</h1><p class="muted" style="margin-bottom:16px">Mohon isi buku tamu berikut.</p>
    <div id="tamuBox"><form class="form" data-form="${id}">
      <label class="f"><span>Nama</span><input name="nama" required></label>
      <div class="row2"><label class="f"><span>Instansi / asal</span><input name="instansi"></label><label class="f"><span>No. HP</span><input name="hp" inputmode="tel"></label></div>
      <label class="f"><span>Keperluan</span><textarea name="keperluan" rows="2" required></textarea></label>
      <label class="f"><span>Ingin bertemu dengan</span><input name="bertemu" placeholder="Nama guru / bagian"></label>
      <div class="upl">${btn('Ambil foto (opsional)', 'tamuCam', {}, 'ghost sm', 'image')}<button type="button" class="btn sm" id="tamuShot" data-act="tamuShot" hidden>Jepret</button><div id="tamuPrev"></div></div>
      <video id="tamuV" class="scan-video" playsinline muted autoplay hidden style="transform:scaleX(-1)"></video>
      <input name="website" tabindex="-1" autocomplete="off" style="position:absolute;left:-9999px" aria-hidden="true">
      <div class="form-err" role="alert"></div><button class="btn block" type="submit">Kirim</button></form></div>`);
  document.title = 'Buku Tamu · ' + (S.settings.nama_sekolah || '');
};

Object.assign(PAGES.admin, { bukutamu: pageBukuTamu });
