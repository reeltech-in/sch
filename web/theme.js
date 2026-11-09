// Terapkan tema pilihan pengguna sebelum halaman digambar (mencegah kedipan putih).
(function () {
  try {
    var t = localStorage.getItem('theme_pref');
    if (t === 'light' || t === 'dark') document.documentElement.setAttribute('data-theme', t);
    var w = localStorage.getItem('theme_warna');
    if (w === 'abu' || w === 'merah') document.documentElement.setAttribute('data-warna', w);
  } catch (e) { /* penyimpanan diblokir: ikut sistem */ }
  // app.js memakai JavaScript modern; browser yang terlalu lama berhenti di "Memuat aplikasi…" tanpa pesan
  setTimeout(function () {
    if (window.SCH_READY) return;
    var el = document.getElementById('app');
    if (el && /Memuat/.test(el.textContent)) {
      el.innerHTML = '<div style="padding:24px;font:16px sans-serif;max-width:480px;margin:auto">Aplikasi belum bisa dibuka di browser ini. ' +
        'Perbarui <b>Google Chrome</b> (atau gunakan Chrome/Firefox terbaru), lalu muat ulang halaman.</div>';
    }
  }, 8000);
})();
