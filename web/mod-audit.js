'use strict';
/* Log aktivitas (admin) */

addNav('admin', ['audit', 'Log Aktivitas', 'clip', '', 'Sistem']);

// Label aksi yang mudah dibaca
const AUDIT_LABELS = [
  [/^LOGIN$/, 'Masuk'], [/^LOGIN GAGAL$/, 'Gagal masuk'], [/^POST \/api\/logout$/, 'Keluar'], [/^POST \/api\/password$/, 'Ganti password'],
  [/^PUT \/api\/submissions\/\d+\/grade$/, 'Memberi nilai tugas'], [/^POST \/api\/assignments\/\d+\/submit$/, 'Mengumpulkan tugas'],
  [/^POST \/api\/users\/import$/, 'Impor akun'], [/^POST \/api\/users$/, 'Tambah akun'], [/^PUT \/api\/users\/\d+$/, 'Ubah akun'], [/^DELETE \/api\/users\/\d+$/, 'Hapus akun'],
  [/^(POST|PUT) \/api\/classes/, 'Ubah kelas'], [/^DELETE \/api\/classes/, 'Hapus kelas'], [/^(POST|PUT) \/api\/courses/, 'Ubah mapel'], [/^DELETE \/api\/courses/, 'Hapus mapel'],
  [/^(POST|PUT|DELETE) \/api\/materials/, 'Ubah materi'], [/^(POST|PUT|DELETE) \/api\/assignments/, 'Ubah tugas'], [/^(POST|PUT|DELETE) \/api\/(quizzes|questions)/, 'Ubah kuis/soal'],
  [/^POST \/api\/quizzes\/\d+\/(start|submit)$/, 'Mengerjakan kuis'], [/^DELETE \/api\/attempts\//, 'Reset hasil kuis'],
  [/^(PUT|DELETE) \/api\/attendance/, 'Absensi mapel'], [/^(PUT) \/api\/absen\/harian/, 'Ceklist absensi harian'], [/^POST \/api\/absen\/scan/, 'Scan kartu absensi'], [/^POST \/api\/absen\/selfie/, 'Absen selfie'],
  [/^POST \/api\/cbt\/\d+\/(start|submit)$/, 'Mengerjakan CBT'], [/^(POST|PUT|DELETE) \/api\/(cbt|cbt-soal)/, 'Ubah ujian CBT'], [/^DELETE \/api\/cbt-attempt/, 'Reset pekerjaan CBT'],
  [/^PUT \/api\/kelulusan$/, 'Status kelulusan'], [/^POST \/api\/tahun\/\d+\/naik$/, 'Kenaikan kelas'], [/^(POST|PUT) \/api\/tahun/, 'Tahun ajaran'],
  [/^PUT \/api\/(settings|modules|config)/, 'Ubah pengaturan'], [/^PUT \/api\/ai\/config$/, 'Ubah pengaturan AI'], [/^POST \/api\/ai\/generate$/, 'Buat dokumen AI'],
  [/^POST \/api\/ppdb\/daftar$/, 'Pendaftaran PPDB'], [/^(PUT|DELETE|POST) \/api\/ppdb\//, 'Kelola PPDB'], [/^POST \/api\/bukutamu$/, 'Isi buku tamu'],
  [/^POST \/api\/kasir\/jual$/, 'Penjualan kasir'], [/^DELETE \/api\/kasir\/penjualan/, 'Batalkan transaksi'], [/^POST \/api\/evote\/\d+\/(pilih|bilik)$/, 'Memberikan suara'],
  [/^POST \/api\/perpus\/(pinjam|kembali)/, 'Sirkulasi perpustakaan'], [/^POST \/api\/upload$/, 'Unggah file'], [/^POST \/api\/pemeliharaan$/, 'Jalankan pemeliharaan'],
];
const auditLabel = a => (AUDIT_LABELS.find(([re]) => re.test(a)) || [null, null])[1];

async function pageAudit() {
  const p = S.params;
  const lim = Number(p.lim || 100);
  const qs = new URLSearchParams({ q: p.q || '', dari: p.dari || '', sampai: p.sampai || '', gagal: p.gagal || '', limit: lim }).toString();
  const { rows, total } = await GETpage('/audit?' + qs);
  return head('Log Aktivitas', 'Catatan otomatis setiap perubahan data: siapa, kapan, dari mana. Disimpan sesuai pengaturan (bawaan 365 hari).') +
    `<div class="bar"><input class="search" type="search" placeholder="Cari nama, aksi, IP, keterangan… (Enter)" value="${esc(p.q || '')}" data-srv="audit" aria-label="Cari">
      <input type="date" value="${esc(p.dari || '')}" data-change="auDari" aria-label="Dari tanggal" style="min-width:0">
      <input type="date" value="${esc(p.sampai || '')}" data-change="auSampai" aria-label="Sampai tanggal" style="min-width:0">
      <label class="chk"><input type="checkbox" data-change="auGagal" ${p.gagal === '1' ? 'checked' : ''}> Hanya gagal</label>
      <small class="muted">${rows.length < total ? `Menampilkan ${rows.length} dari ${total}` : `${total} catatan`}</small></div>` +
    (rows.length ? `<div class="tw"><table><thead><tr><th>Waktu</th><th>Pengguna</th><th>Aksi</th><th>Status</th><th>IP</th><th>Keterangan</th></tr></thead><tbody>
    ${rows.map(r => { const lbl = auditLabel(r.aksi); return `<tr><td><small>${fmtDT(r.waktu)}</small></td><td>${r.nama ? `<b>${esc(r.nama)}</b><small class="muted">${esc(r.role || '')}</small>` : '<small class="muted">publik</small>'}</td>
      <td>${lbl ? `<b>${esc(lbl)}</b>` : ''}<small class="muted"><code>${esc(r.aksi)}</code></small></td>
      <td>${r.status >= 400 ? tag(String(r.status), 'bad') : tag('OK', 'ok')}</td><td><small>${esc(r.ip || '')}</small></td><td><small>${esc(r.detail || '')}</small></td></tr>`; }).join('')}
    </tbody></table></div>${moreBtn('audit', total, rows.length)}` : `<div class="card">${empty('Belum ada catatan.', 'clip')}</div>`);
}
CHG.auDari = el => go('audit', { ...S.params, dari: el.value, lim: '' });
CHG.auSampai = el => go('audit', { ...S.params, sampai: el.value, lim: '' });
CHG.auGagal = el => go('audit', { ...S.params, gagal: el.checked ? '1' : '', lim: '' });

Object.assign(PAGES.admin, { audit: pageAudit });
