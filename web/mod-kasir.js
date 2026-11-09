'use strict';
/* Kasir / Koperasi: penjualan dengan barcode, stok, struk, laporan laba */

addNav('admin', ['kasir', 'Kasir / Koperasi', 'cart', 'kasir', 'Layanan']);

// =====================================================================
// KASIR / KOPERASI
// =====================================================================
async function pageKasir() {
  const tab = ['jual', 'barang', 'laporan'].includes(S.params.tab) ? S.params.tab : 'jual';
  const hd = head('Kasir / Koperasi', 'Penjualan dengan scan barcode, struk, stok otomatis, dan laporan laba.');
  const t = tabs([['jual', 'Kasir'], ['barang', 'Barang & stok'], ['laporan', 'Laporan']], tab);
  const ps = await GET('/produk');
  S.prod = Object.fromEntries(ps.map(p => [p.id, p])); S.prodList = ps;
  if (S.cart) S.cart = S.cart.filter(x => S.prod[x.id]).map(x => ({ ...x, qty: Math.max(1, Math.min(x.qty, S.prod[x.id].stok || x.qty)) }));
  if (tab === 'barang') return hd + t + kasirBarang(ps);
  if (tab === 'laporan') return hd + t + await kasirLaporan();
  S.cart = S.cart || [];
  S.after = () => {
    const inp = $('#posIn'); inp?.focus();
    inp?.addEventListener('keydown', e => { if (e.key === 'Enter') { e.preventDefault(); const v = inp.value.trim(); inp.value = ''; if (v) posAdd(v); } });
    renderCart();
  };
  return hd + t + `<div class="cols cols-21"><section class="card">
      <input class="scan-input" id="posIn" list="posList" placeholder="Scan barcode / ketik nama barang lalu Enter" autocomplete="off" aria-label="Barang">
      <datalist id="posList">${ps.filter(p => p.aktif).map(p => `<option value="${esc(p.kode)}">${esc(p.nama)} · ${rupiah(p.harga)}</option>`).join('')}</datalist>
      <div id="cart" style="margin-top:14px"></div></section>
    <section class="card"><div class="card-head"><h3>Pembayaran</h3></div><div id="payBox"></div></section></div>`;
}
function posAdd(q) {
  const ql = q.toLowerCase();
  const p = S.prodList.find(x => x.kode.toLowerCase() === ql) || S.prodList.find(x => x.nama.toLowerCase().includes(ql));
  if (!p) { beep(false); return toast('Barang tidak ditemukan: ' + q, 'bad'); }
  if (!p.aktif) return toast(p.nama + ' sedang tidak dijual.', 'bad');
  const it = S.cart.find(x => x.id === p.id);
  if ((it ? it.qty : 0) + 1 > p.stok) { beep(false); return toast(`Stok ${p.nama} tinggal ${p.stok}.`, 'bad'); }
  it ? it.qty++ : S.cart.push({ id: p.id, qty: 1 });
  beep(); renderCart();
}
const cartTotal = () => S.cart.reduce((s, x) => s + S.prod[x.id].harga * x.qty, 0);
function renderCart() {
  const box = $('#cart'); if (!box) return;
  box.innerHTML = S.cart.length ? `<div class="tw"><table><thead><tr><th>Barang</th><th class="n">Harga</th><th class="n">Jumlah</th><th class="n">Subtotal</th><th></th></tr></thead><tbody>
    ${S.cart.map((x, i) => { const p = S.prod[x.id]; return `<tr><td><b>${esc(p.nama)}</b><small class="muted">${esc(p.kode)}</small></td><td class="n">${rupiah(p.harga)}</td>
      <td class="n"><div class="acts">${btn('−', 'cartQty', { i, d: -1 }, 'ghost sm icon')}<b style="min-width:26px;text-align:center;align-self:center">${x.qty}</b>${btn('+', 'cartQty', { i, d: 1 }, 'ghost sm icon')}</div></td>
      <td class="n">${rupiah(p.harga * x.qty)}</td><td>${iconBtn('x', 'cartDel', { i }, 'Hapus')}</td></tr>`; }).join('')}</tbody></table></div>` : empty('Keranjang kosong. Scan barang untuk mulai.', 'cart');
  const tot = cartTotal();
  $('#payBox').innerHTML = `<div style="font-size:13px" class="muted">Total</div><div style="font-size:34px;font-weight:800;color:var(--brand);margin-bottom:12px">${rupiah(tot)}</div>
    <label class="f"><span>Uang dibayar</span><input type="number" id="payIn" min="0" step="500" value="${tot || ''}" class="scan-input" style="text-align:right"></label>
    <div class="acts" style="justify-content:flex-start;margin:8px 0 12px">${[tot, 5000, 10000, 20000, 50000, 100000].filter((v, i, a) => v && a.indexOf(v) === i).map(v => btn(v === tot ? 'Uang pas' : rupiah(v), 'payQuick', { v }, 'ghost sm')).join('')}</div>
    <div id="kembali" class="muted" style="margin-bottom:12px"></div>
    ${btn('Bayar & cetak struk', 'paySubmit', {}, 'block', 'check')}${S.cart.length ? `<div style="margin-top:8px">${btn('Kosongkan', 'cartClear', {}, 'ghost sm block')}</div>` : ''}`;
  const pin = $('#payIn');
  const upd = () => { const k = Number(pin.value || 0) - tot; $('#kembali').innerHTML = tot ? (k >= 0 ? `Kembalian: <b>${rupiah(k)}</b>` : `<span style="color:var(--bad)">Kurang ${rupiah(-k)}</span>`) : ''; };
  pin.addEventListener('input', upd); upd();
}
ACT.cartQty = el => { const x = S.cart[el.dataset.i]; const p = S.prod[x.id]; const n = x.qty + Number(el.dataset.d); if (n > p.stok) return toast(`Stok ${p.nama} tinggal ${p.stok}.`, 'bad'); if (n < 1) S.cart.splice(el.dataset.i, 1); else x.qty = n; renderCart(); };
ACT.cartDel = el => { S.cart.splice(el.dataset.i, 1); renderCart(); };
ACT.cartClear = () => { S.cart = []; renderCart(); $('#posIn')?.focus(); };
ACT.payQuick = el => { $('#payIn').value = el.dataset.v; $('#payIn').dispatchEvent(new Event('input')); };
ACT.paySubmit = async el => {
  if (!S.cart.length) return toast('Keranjang masih kosong.', 'bad');
  el.disabled = true;
  try {
    const r = await POST('/kasir/jual', { items: S.cart.map(x => ({ product_id: x.id, qty: x.qty })), bayar: Number($('#payIn').value || 0) });
    S.cart = [];
    showStruk(r);
    await refresh(true);
  } finally { el.disabled = false; }
};
function strukHtml(r) {
  return `<div class="struk"><div style="text-align:center"><b>${esc(S.settings.nama_sekolah || '')}</b><br>Koperasi Sekolah<br>${fmtDT(r.created_at)}<br>${esc(r.no)}</div><hr>
    ${r.items.map(i => `<div>${esc(i.nama)}</div><div style="display:flex;justify-content:space-between"><span>${i.qty} x ${rupiah(i.harga)}</span><span>${rupiah(i.subtotal)}</span></div>`).join('')}<hr>
    <div style="display:flex;justify-content:space-between"><b>Total</b><b>${rupiah(r.total)}</b></div><div style="display:flex;justify-content:space-between"><span>Bayar</span><span>${rupiah(r.bayar)}</span></div>
    <div style="display:flex;justify-content:space-between"><span>Kembali</span><span>${rupiah(r.kembali)}</span></div><hr><div style="text-align:center">Kasir: ${esc(r.kasir || '')}<br>Terima kasih</div></div>`;
}
function showStruk(r) {
  S.struk = r;
  openModal({ title: 'Transaksi berhasil', body: `<div class="scan-result ok" style="margin-bottom:12px"><div><b>Kembalian ${rupiah(r.kembali)}</b>${esc(r.no)}</div></div>${strukHtml(r)}
    <div class="sheet-foot" style="margin-top:14px">${btn('Cetak struk', 'strukPrint', {}, 'ghost', 'print')}<button class="btn" data-act="closeModal">Transaksi baru</button></div>` });
}
ACT.strukPrint = () => printDoc(`<div style="width:58mm;font-size:9pt">${strukHtml(S.struk)}</div>`);

function kasirBarang(ps) {
  const nilai = ps.reduce((s, p) => s + p.harga_beli * p.stok, 0);
  return `<div class="card-head"><p class="muted">${ps.length} barang · nilai stok (harga beli) ${rupiah(nilai)}</p><div class="acts">${btn('Cetak label', 'prodLabels', {}, 'ghost sm', 'print')}${btn('Tambah barang', 'prodForm', {}, 'sm', 'plus')}</div></div>
    <div class="bar"><input class="search" type="search" placeholder="Cari barang…" data-filter="tblPr" aria-label="Cari"></div>` +
    (ps.length ? `<div class="tw"><table id="tblPr"><thead><tr><th>Barang</th><th>Kategori</th><th class="n">Harga jual</th><th class="n">Harga beli</th><th class="n">Stok</th><th></th></tr></thead><tbody>
    ${ps.map(p => `<tr><td><b>${esc(p.nama)}</b><small class="muted"><code>${esc(p.kode)}</code>${p.aktif ? '' : ' · tidak dijual'}</small></td><td>${esc(p.kategori || '–')}</td><td class="n">${rupiah(p.harga)}</td><td class="n">${rupiah(p.harga_beli)}</td>
      <td class="n">${p.stok <= 5 ? tag(p.stok + ' ' + (p.satuan || ''), p.stok ? 'warn' : 'bad') : p.stok + ' ' + esc(p.satuan || '')}</td>
      <td><div class="acts">${btn('Stok', 'prodStock', { id: p.id }, 'ghost sm')}${iconBtn('edit', 'prodForm', { id: p.id }, 'Ubah')}${iconBtn('trash', 'prodDel', { id: p.id, nama: p.nama }, 'Hapus', 'danger')}</div></td></tr>`).join('')}
    </tbody></table></div>` : `<div class="card">${empty('Belum ada barang.', 'cart')}</div>`);
}
ACT.prodForm = async el => {
  const [kat, sat] = await Promise.all([saran('produk.kategori', ['ATK', 'Makanan', 'Minuman', 'Seragam', 'Buku', 'Perlengkapan sekolah']),
    saran('produk.satuan', ['pcs', 'buah', 'pak', 'lusin', 'rim', 'botol', 'bungkus', 'set', 'lembar'])]);
  const p = el.dataset.id ? { ...S.prod[el.dataset.id], aktif: !!S.prod[el.dataset.id].aktif } : { aktif: true, satuan: 'pcs', stok: 0 };
  const fields = [[{ n: 'kode', l: 'Kode / barcode', req: true, hint: 'Scan barcode kemasan, atau buat kode sendiri.' }, { n: 'nama', l: 'Nama barang', req: true }],
    [{ n: 'harga', l: 'Harga jual (Rp)', t: 'number', min: 0, req: true }, { n: 'harga_beli', l: 'Harga beli (Rp)', t: 'number', min: 0 }],
    [{ n: 'kategori', l: 'Kategori', t: 'combo', o: kat }, { n: 'satuan', l: 'Satuan', t: 'combo', o: sat }]];
  if (!p.id) fields.push({ n: 'stok', l: 'Stok awal', t: 'number', min: 0 });
  fields.push({ n: 'aktif', l: 'Dijual', t: 'check' });
  openForm({ title: p.id ? 'Ubah barang' : 'Tambah barang', values: p, fields,
    onSubmit: v => { const b = { ...v, harga: v.harga || 0, harga_beli: v.harga_beli || 0, stok: v.stok || 0 }; return p.id ? PUT('/produk/' + p.id, b) : POST('/produk', b); } });
};
ACT.prodDel = el => del('/produk/' + el.dataset.id, `Hapus barang <b>${esc(el.dataset.nama)}</b>? Riwayat penjualan tetap tersimpan.`);
ACT.prodStock = el => {
  const p = S.prod[el.dataset.id];
  openForm({ title: 'Stok: ' + p.nama, intro: `<p class="muted" style="margin-bottom:10px">Stok sekarang <b>${p.stok}</b>. Isi angka positif untuk barang masuk, negatif untuk koreksi/rusak.</p>`,
    fields: [{ n: 'tambah', l: 'Perubahan stok', t: 'number', req: true }], success: 'Stok diperbarui.', onSubmit: v => POST(`/produk/${p.id}/stok`, { tambah: v.tambah }) });
};
ACT.prodLabels = () => {
  if (!S.prodList.length) return toast('Belum ada barang.', 'bad');
  const lab = p => `<div class="label">${qrImg(p.kode, 84)}<div><b>${esc(p.nama)}</b><br>${rupiah(p.harga)}<br><small>${esc(p.kode)}</small></div></div>`;
  const pages = [];
  for (let i = 0; i < S.prodList.length; i += 24) pages.push(`<div class="pg"><div class="labels">${S.prodList.slice(i, i + 24).map(lab).join('')}</div></div>`);
  printDoc(pages.join(''));
};

async function kasirLaporan() {
  const mode = S.params.bulan ? 'bulan' : 'hari';
  // nilai dari URL divalidasi ketat (cegah XSS lewat link berisi kode)
  let val = mode === 'bulan' ? S.params.bulan : (S.params.tanggal || today());
  if (!(mode === 'bulan' ? /^\d{4}-\d{2}$/ : /^\d{4}-\d{2}-\d{2}$/).test(val || '')) val = mode === 'bulan' ? monthNow() : today();
  const r = await GET('/kasir/penjualan?' + (mode === 'bulan' ? 'bulan=' + val : 'tanggal=' + val));
  S.lapKasir = r;
  const omzet = r.penjualan.reduce((s, x) => s + x.total, 0), laba = r.barang.reduce((s, x) => s + x.laba, 0);
  return `<div class="bar"><select data-change="lapMode" aria-label="Periode"><option value="hari" ${mode === 'hari' ? 'selected' : ''}>Harian</option><option value="bulan" ${mode === 'bulan' ? 'selected' : ''}>Bulanan</option></select>
      <input type="${mode === 'bulan' ? 'month' : 'date'}" value="${esc(val)}" data-change="lapVal" aria-label="Tanggal" style="min-width:0">${btn('Cetak', 'lapPrint', {}, 'ghost sm', 'print')}</div>
    <div class="stats">${statRp('chart', omzet, 'Omzet')}${statRp('check', laba, 'Laba kotor')}${stat('cart', r.penjualan.length, 'Transaksi')}</div>
    <div class="cols"><section class="card"><div class="card-head"><h3>Barang terlaris</h3></div>${r.barang.length ? r.barang.slice(0, 15).map(b => `<div class="li"><div><b>${esc(b.nama)}</b><small class="muted">${b.qty} terjual · laba ${rupiah(b.laba)}</small></div><b>${rupiah(b.omzet)}</b></div>`).join('') : empty('Belum ada penjualan.', 'cart')}</section>
    <section class="card"><div class="card-head"><h3>Transaksi</h3></div>${r.penjualan.length ? r.penjualan.map(s => `<div class="li"><div><b>${esc(s.no)}</b><small class="muted">${fmtDT(s.created_at)} · ${esc(s.kasir || '')}</small></div>
      <div class="acts"><b style="align-self:center">${rupiah(s.total)}</b>${iconBtn('eye', 'saleView', { id: s.id }, 'Lihat struk')}${S.user.role === 'admin' ? iconBtn('trash', 'saleVoid', { id: s.id, no: s.no }, 'Batalkan', 'danger') : ''}</div></div>`).join('') : empty('Belum ada transaksi.', 'inbox')}</section></div>`;
}
CHG.lapMode = el => go('kasir', el.value === 'bulan' ? { tab: 'laporan', bulan: monthNow() } : { tab: 'laporan' });
CHG.lapVal = el => go('kasir', el.type === 'month' ? { tab: 'laporan', bulan: el.value } : { tab: 'laporan', tanggal: el.value });
ACT.saleView = async el => showStruk(await GET('/kasir/penjualan/' + el.dataset.id));
ACT.saleVoid = el => del('/kasir/penjualan/' + el.dataset.id, `Batalkan transaksi <b>${esc(el.dataset.no)}</b>? Stok barang dikembalikan.`, 'Transaksi dibatalkan.');
ACT.lapPrint = () => {
  const r = S.lapKasir;
  printDoc(schoolHead() + `<h3 style="text-align:center">LAPORAN PENJUALAN KOPERASI · ${r.dari === r.sampai ? fmtDate(r.dari) : fmtDate(r.dari) + ' – ' + fmtDate(r.sampai)}</h3>
    <table><thead><tr><th>Barang</th><th>Terjual</th><th>Omzet</th><th>Laba</th></tr></thead><tbody>${r.barang.map(b => `<tr><td>${esc(b.nama)}</td><td>${b.qty}</td><td>${rupiah(b.omzet)}</td><td>${rupiah(b.laba)}</td></tr>`).join('')}
    <tr><td><b>Jumlah</b></td><td></td><td><b>${rupiah(r.barang.reduce((s, b) => s + b.omzet, 0))}</b></td><td><b>${rupiah(r.barang.reduce((s, b) => s + b.laba, 0))}</b></td></tr></tbody></table>`);
};

Object.assign(PAGES.admin, { kasir: pageKasir });
