# Graph Report - .  (2026-11-10)

## Corpus Check
- 83 files · ~143,868 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1227 nodes · 2365 edges · 62 communities detected
- Extraction: 100% EXTRACTED · 0% INFERRED · 0% AMBIGUOUS
- Token cost: 0 input · 0 output
- Edge kinds: contains: 945 · calls: 904 · MODIFIES: 253 · method: 160 · ON_BRANCH: 52 · PARENT_OF: 51


## Input Scope
- Requested: auto
- Resolved: committed (source: default-auto)
- Included files: 83 · Candidates: 97
- Excluded: 0 untracked · 409 ignored · 0 sensitive · 0 missing committed
- Recommendation: Use --scope all or graphify.yaml inputs.corpus for a knowledge-base folder.

## Graph Freshness
- Built from Git commit: `ec20e2b`
- Compare this hash to `git rev-parse HEAD` before trusting freshness-sensitive graph output.
## God Nodes (most connected - your core abstractions)
1. `App` - 69 edges
2. `Ctx` - 62 edges
3. `newEnv()` - 32 edges
4. `esc()` - 29 edges
5. `expect()` - 25 edges
6. `ico()` - 15 edges
7. `loginLimiter` - 11 edges
8. `route()` - 11 edges
9. `apiRaporMapel()` - 10 edges
10. `api()` - 10 edges

## Surprising Connections (you probably didn't know these)
- `09479bb Perbarui peta kode graphify` --ON_BRANCH--> `main`  [EXTRACTED]
  git → git  _Bridges community 31 → community 8_
- `1119b6b Tahap 3: mesin ujian kuis & CBT memakai hitung mundur, kirim otomatis, dan kartu hasil bersama` --ON_BRANCH--> `main`  [EXTRACTED]
  git → git  _Bridges community 0 → community 8_
- `1bb9e1e v1.20.0: perbaikan hasil audit keamanan, ujian, query, dan tampilan` --ON_BRANCH--> `main`  [EXTRACTED]
  git → git  _Bridges community 2 → community 8_
- `1bb9e1e v1.20.0: perbaikan hasil audit keamanan, ujian, query, dan tampilan` --PARENT_OF--> `3a5fb3d Perbarui peta kode graphify`  [EXTRACTED]
  git → git  _Bridges community 2 → community 42_
- `3a5fb3d Perbarui peta kode graphify` --ON_BRANCH--> `main`  [EXTRACTED]
  git → git  _Bridges community 42 → community 8_

## Communities

### Community 0 - "Community 0"
Cohesion: 0.03
Nodes (42): 1119b6b Tahap 3: mesin ujian kuis & CBT memakai hitung mundur, kirim otomatis, dan kartu hasil bersama, 1e19524 Tahap 1: peserta LJK ikut riwayat kelas, validasi jam absensi, dasbor guru per tahun aktif, cache & halaman publik lebih aman, muat paralel, bcd8537 Tahap 2: inti permintaan bersama + batal saat pindah menu, helper kamera/scan/password bersama, jendela dialog aksesibel, acctItems(), acctMenu(), ACT, BOOT, BUSY (+34 more)

### Community 1 - "Community 1"
Cohesion: 0.06
Nodes (30): hashToken(), bad(), forbidden(), notFound(), writeErr(), clientIP(), isHTTPS(), limitKey() (+22 more)

### Community 2 - "Community 2"
Cohesion: 0.05
Nodes (42): 1bb9e1e v1.20.0: perbaikan hasil audit keamanan, ujian, query, dan tampilan, 334b045 v1.19.2: dropdown pilih-atau-ketik ramah HP, panah select seragam, Asisten AI satu kolom di HP, b86971e v1.19.1: mapel yang diampu guru diutamakan di Asisten AI dan CBT, apiKalender(), jamLokal(), localDay(), tglJamLokal(), AGAMA (+34 more)

### Community 3 - "Community 3"
Cohesion: 0.05
Nodes (16): apiDeleteQuestion(), apiQuizAnswersAutosave(), apiQuizSave(), apiQuizStart(), apiQuizSubmit(), apiSaveQuestion(), quizClean(), quizDeadline() (+8 more)

### Community 4 - "Community 4"
Cohesion: 0.20
Nodes (37): client, env, fakeWA, expect(), newEnv(), newFakeWA(), TestAdmGuruWaliDanCatatan(), TestCBTBatasWaktuDiServer() (+29 more)

### Community 5 - "Community 5"
Cohesion: 0.08
Nodes (26): apiCBTAttemptDelete(), apiCBTDelete(), apiCBTHasil(), apiCBTList(), apiCBTSave(), apiCBTSaveAnswers(), apiCBTSoal(), apiCBTSoalDelete() (+18 more)

### Community 6 - "Community 6"
Cohesion: 0.09
Nodes (36): applyTheme(), avatar(), brandHtml(), btn(), cycleTheme(), empty(), ensureThemeFloat(), esc() (+28 more)

### Community 7 - "Community 7"
Cohesion: 0.10
Nodes (13): apiElectionResult(), apiElections(), apiElectionsForMe(), apiVote(), castVote(), apiElectionResult(), apiElections(), apiElectionsForMe() (+5 more)

### Community 8 - "Community 8"
Cohesion: 0.11
Nodes (31): main, 01cf6c7 Perbarui peta kode graphify, 044f2aa v1.23.1: kerapian menu guru & siswa di HP: tombol absensi lebih besar + simpan menempel, kolom nama tetap terlihat di tabel rekap, kehadiran ditulis lengkap, 'Belum ada materi', 137792c Tahap 4: hapus 57 deklarasi CSS yang selalu ditimpa (tampilan terverifikasi identik di desktop & HP); peringatan keluar hanya saat ujian belum terkirim, 306dce2 Perbarui peta kode graphify, 3550bcd v1.24.1: audit tahap 1 - kuis dinilai dari soal yang ditampilkan, batas unggahan publik, tugas dinilai tak tertimpa, buku berriwayat tak dihapus, putus sesi atomik, petugas bilik e-voting tercatat, 3cd8909 Perbarui peta kode graphify, 3e27651 Perbarui peta kode graphify (+23 more)

### Community 9 - "Community 9"
Cohesion: 0.09
Nodes (24): absenCeklist(), absenRekap(), absenScanHtml(), AKSI, autoTag(), canEditEv(), doScan(), evRow() (+16 more)

### Community 10 - "Community 10"
Cohesion: 0.11
Nodes (2): cfgInt(), App

### Community 11 - "Community 11"
Cohesion: 0.13
Nodes (16): aiAdminOnly(), aiCleanInputs(), aiKeyHint(), aiProviderError(), aiReadStream(), aiRequest(), aiUsageKey(), apiAIConfig() (+8 more)

### Community 12 - "Community 12"
Cohesion: 0.13
Nodes (16): apiEkskulNilaiSave(), apiEkskulSave(), apiPortoCatatan(), apiPortoDelete(), apiPortoList(), apiRaporDocx(), apiRaporMapel(), apiRaporMapelAI() (+8 more)

### Community 13 - "Community 13"
Cohesion: 0.11
Nodes (5): apiDelete(), apiDeleteChild(), onlyDigits(), tingkatDariNama(), Ctx

### Community 14 - "Community 14"
Cohesion: 0.10
Nodes (15): adminDash(), adminPengaturan(), fmtBytes(), KEYMAP, ROMAWI, SET_PAGES, SET_SECTIONS, setCard() (+7 more)

### Community 15 - "Community 15"
Cohesion: 0.12
Nodes (25): afterSave(), closeModal(), errBox(), formHtml(), go(), head(), isAdm(), judulKeHeader() (+17 more)

### Community 16 - "Community 16"
Cohesion: 0.13
Nodes (19): bookCard(), bookCover(), bookRead(), circBorrow(), circLog(), circMember(), circUpdate(), hasEbook() (+11 more)

### Community 17 - "Community 17"
Cohesion: 0.13
Nodes (11): apiAbsenHarian(), apiAbsenHarianSave(), apiAbsenRekap(), haversine(), init(), itoa(), jamValid(), localClock() (+3 more)

### Community 18 - "Community 18"
Cohesion: 0.15
Nodes (8): activeYear(), apiPromote(), nextTingkat(), nullIfEmpty(), unlistedClasses(), promoteKelas, promoter, promoteReq

### Community 19 - "Community 19"
Cohesion: 0.18
Nodes (11): apiSibiImport(), apiSibiSearch(), cacheName(), ebookCacheMax(), nullStr(), safeFileName(), sibiFetch(), sibiJenjang() (+3 more)

### Community 20 - "Community 20"
Cohesion: 0.12
Nodes (20): clone(), comboHtml(), comboItems(), examTime(), fieldHtml(), GET(), jenjangDiatur(), kelasOpts() (+12 more)

### Community 21 - "Community 21"
Cohesion: 0.11
Nodes (2): 5720746 SCH v1.18.0: sistem informasi sekolah (Go + SQLite), GRAD

### Community 22 - "Community 22"
Cohesion: 0.17
Nodes (8): apiJournal(), apiNoteDelete(), apiNotes(), apiNoteSave(), apiNoteStudents(), apiWaliClasses(), apiWaliDetail(), str2()

### Community 23 - "Community 23"
Cohesion: 0.14
Nodes (3): apiPPDBInfo(), ppdbOpen(), randDigits()

### Community 24 - "Community 24"
Cohesion: 0.18
Nodes (7): apiWAConfigSave(), cfgOn(), cfgStr(), waFill(), waHost(), waNumber(), waSend()

### Community 25 - "Community 25"
Cohesion: 0.18
Nodes (10): addColumn(), apiGetConfig(), apiPutConfig(), apiPutModules(), getConfig(), init(), kvForget(), kvGet() (+2 more)

### Community 26 - "Community 26"
Cohesion: 0.18
Nodes (12): pageRapor(), raporHtml(), RP_EKSKUL, RP_KOKU, rpAtur(), rpEkskul(), rpGetRapor(), rpMapel() (+4 more)

### Community 27 - "Community 27"
Cohesion: 0.15
Nodes (4): kunciUjian(), TestCBTDinilaiOtomatisSetelahWaktuHabis(), TestCBTSoalTambahanTidakMenurunkanNilai(), TestKirimUlangSetelahDinilaiOtomatis()

### Community 28 - "Community 28"
Cohesion: 0.17
Nodes (8): b86e256 Perbarui peta kode graphify, bfca767 Perbarui peta kode graphify, c05aceb Tahun ajaran satu sumber: Pengaturan hanya menampilkan tahun aktif; server selalu memakai tahun aktif dari menu Tahun Ajaran, d53112a Kelas: tahun ajaran tidak lagi terhapus saat kelas diubah; data lama diperbaiki lewat migrasi; kolom tahun ajaran dihapus dari tabel Kelas, siswaID(), TestAbsensiMapelAlurLengkap(), TestPinjamKembaliBuku(), TestTugasBuatKumpulNilai()

### Community 29 - "Community 29"
Cohesion: 0.19
Nodes (8): JAM_KE, JENIS_CAT, KATEGORI_CAT, monthInput(), pageCatatan(), pageJurnal(), pageWaliKelas(), validBulan()

### Community 30 - "Community 30"
Cohesion: 0.14
Nodes (2): 8c2e389 v1.21.1: kuis mapel tersimpan otomatis & dinilai bila terlambat, pencarian log 90 hari berindeks, simpan pengaturan sekali, saring kelas di browser, ec20e2b v1.25.1: statistik query SQLite selalu segar (PRAGMA optimize=0x10002 saat menyala & tiap jam), rekap lebih ringan, permintaan e-Rapor digabung

### Community 31 - "Community 31"
Cohesion: 0.16
Nodes (6): 09479bb Perbarui peta kode graphify, 2c481ea v1.22.1: perangkat ajar AI per jenjang (SD/SMP/SMA/SMK): kelas & fase, saran mapel, validasi server, instruksi jenjang ke AI, 4f7bbc2 Perbarui peta kode graphify, 69d9920 Perbarui peta kode graphify, 792dd82 v1.23.0: jenjang sekolah diatur sekali di Pengaturan; tingkat kelas, saran mapel, kenaikan kelas, filter e-book, dan Asisten AI mengikuti jenjang, c3fe8fe v1.22.0: RPP Kurikulum Merdeka terbaru (RPM/Modul Ajar/RPP 1 lembar, fase & alokasi waktu otomatis, cek waktu), sunting, cetak resmi berkop & tanda tangan, unduh Word

### Community 32 - "Community 32"
Cohesion: 0.19
Nodes (8): apiEkskulAnggota(), apiEkskulAnggotaSave(), apiEkskulCalon(), apiEkskulList(), apiRaporKelas(), apiRaporTugasku(), deskripsiKoku(), periode

### Community 33 - "Community 33"
Cohesion: 0.23
Nodes (10): findMarkers(), homography(), LJK, ljkPos(), ljkProcess(), ljkRead(), ljkRedraw(), ljkSheet() (+2 more)

### Community 34 - "Community 34"
Cohesion: 0.29
Nodes (11): noKelas(), siswaCourse(), siswaDash(), siswaData(), siswaKuis(), siswaMapel(), siswaNilai(), siswaTugas() (+3 more)

### Community 35 - "Community 35"
Cohesion: 0.27
Nodes (8): apiGetSettings(), apiPutSettings(), apiSaveClass(), cleanDeadline(), cleanURL(), deref(), required(), str()

### Community 36 - "Community 36"
Cohesion: 0.37
Nodes (11): apiAIDocDocx(), buildDocx(), docxBody(), docxKop(), docxPara(), docxRuns(), docxTable(), docxTTD() (+3 more)

### Community 37 - "Community 37"
Cohesion: 0.26
Nodes (13): api(), apiFetch(), apiJSON(), checkVersion(), confirmBox(), DEL(), examSubmit(), GETpage() (+5 more)

### Community 38 - "Community 38"
Cohesion: 0.38
Nodes (11): courseBar(), guruAbsensi(), guruCourses(), guruKuis(), guruMapel(), guruMateri(), guruRekap(), guruTugas() (+3 more)

### Community 39 - "Community 39"
Cohesion: 0.18
Nodes (2): apiReturn(), perpusCfg()

### Community 40 - "Community 40"
Cohesion: 0.23
Nodes (8): apiAudit(), applyTimezone(), checkPassword(), dayRangeUTC(), init(), pageParams(), withBcrypt(), statusRecorder

### Community 41 - "Community 41"
Cohesion: 0.24
Nodes (5): apiDeleteCourse(), apiDeleteUser(), apiSaveUser(), b2i(), userIn

### Community 42 - "Community 42"
Cohesion: 0.18
Nodes (3): 3a5fb3d Perbarui peta kode graphify, 450a912 v1.20.0: perbaiki sisa temuan audit dan risiko baru dari tinjauan ulang, ae5fa1f Perbarui peta kode graphify

### Community 43 - "Community 43"
Cohesion: 0.27
Nodes (6): apiRaporMe(), apiRaporPeriode(), apiRaporSiswa(), apiRaporSiswaSave(), apiRaporTerbit(), rentangSemester()

### Community 44 - "Community 44"
Cohesion: 0.27
Nodes (6): aiJenjang(), rppFase(), rppMenitJP(), rppNum(), rppPrompt(), rppWaktu()

### Community 45 - "Community 45"
Cohesion: 0.24
Nodes (3): pngDataURL(), TestPPDBDitolakTidakMenyimpanBerkas(), TestUnggahanPublikDibatasiHarian()

### Community 46 - "Community 46"
Cohesion: 0.33
Nodes (8): bilikHtml(), bilikOpen(), bilikScan(), elStatus(), evoteVoter(), hasilHtml(), hasilUtama(), pageEvoting()

### Community 47 - "Community 47"
Cohesion: 0.28
Nodes (4): cbtTimeText(), pageCBT(), renderCbt(), siswaCBT()

### Community 48 - "Community 48"
Cohesion: 0.36
Nodes (8): cartTotal(), kasirBarang(), kasirLaporan(), pageKasir(), posAdd(), renderCart(), showStruk(), strukHtml()

### Community 49 - "Community 49"
Cohesion: 0.29
Nodes (2): apiRekapMe(), nilaiSiswa()

### Community 50 - "Community 50"
Cohesion: 0.36
Nodes (4): ensureAdmin(), hashPassword(), nowStr(), randomPassword()

### Community 51 - "Community 51"
Cohesion: 0.29
Nodes (2): apiSale(), apiSell()

### Community 52 - "Community 52"
Cohesion: 0.29
Nodes (4): nextYearName(), NK_AKSI, pageNaik(), ROMAN

### Community 53 - "Community 53"
Cohesion: 0.33
Nodes (2): apiPPDBInfo(), ppdbOpen()

### Community 54 - "Community 54"
Cohesion: 0.43
Nodes (6): nilaiAkhir(), rata(), skorDariRekap(), toF(), skorMapel, tugasNilai

### Community 55 - "Community 55"
Cohesion: 0.43
Nodes (5): apiCBTSoalBatch(), apiQuizQuestionBatch(), parseBatch(), parseMCQ(), mcq

### Community 57 - "Community 57"
Cohesion: 0.38
Nodes (4): AUDIT_LABELS, GETpage(), moreBtn(), pageAudit()

### Community 59 - "Community 59"
Cohesion: 0.67
Nodes (5): f(), fmtSkor(), sama(), TestNilaiAkhirDiRekapDanSiswaSama(), TestRumusNilaiAkhir()

### Community 60 - "Community 60"
Cohesion: 0.50
Nodes (3): defaultTahunAjaran(), migTahunAjaran(), migration

### Community 61 - "Community 61"
Cohesion: 0.50
Nodes (2): manUjianDetail(), pageManUjian()

### Community 62 - "Community 62"
Cohesion: 0.40
Nodes (2): BERKAS, PPDB_ST

### Community 65 - "Community 65"
Cohesion: 0.50
Nodes (2): KODE_SURAT, SURAT_ST

## Knowledge Gaps
- **97 isolated node(s):** `lampiran`, `userIn`, `docxOpts`, `migration`, `absenResult` (+92 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `Community 10`** (2 nodes): `cfgInt()`, `App`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 21`** (2 nodes): `5720746 SCH v1.18.0: sistem informasi sekolah (Go + SQLite)`, `GRAD`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 30`** (2 nodes): `8c2e389 v1.21.1: kuis mapel tersimpan otomatis & dinilai bila terlambat, pencarian log 90 hari berindeks, simpan pengaturan sekali, saring kelas di browser`, `ec20e2b v1.25.1: statistik query SQLite selalu segar (PRAGMA optimize=0x10002 saat menyala & tiap jam), rekap lebih ringan, permintaan e-Rapor digabung`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 39`** (2 nodes): `apiReturn()`, `perpusCfg()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 49`** (2 nodes): `apiRekapMe()`, `nilaiSiswa()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 51`** (2 nodes): `apiSale()`, `apiSell()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 53`** (2 nodes): `apiPPDBInfo()`, `ppdbOpen()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 61`** (2 nodes): `manUjianDetail()`, `pageManUjian()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 62`** (2 nodes): `BERKAS`, `PPDB_ST`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 65`** (2 nodes): `KODE_SURAT`, `SURAT_ST`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App` connect `Community 10` to `Community 30`, `Community 17`, `Community 19`, `Community 39`, `Community 23`, `Community 1`, `Community 12`, `Community 7`, `Community 24`, `Community 5`, `Community 3`, `Community 36`, `Community 25`, `Community 13`, `Community 35`?**
  _High betweenness centrality (0.092) - this node is a cross-community bridge._
- **Why does `Ctx` connect `Community 13` to `Community 1`, `Community 32`, `Community 11`, `Community 12`, `Community 2`, `Community 23`, `Community 43`, `Community 25`, `Community 7`, `Community 22`, `Community 41`, `Community 17`, `Community 5`, `Community 3`, `Community 18`?**
  _High betweenness centrality (0.083) - this node is a cross-community bridge._
- **What connects `lampiran`, `userIn`, `docxOpts` to the rest of the system?**
  _97 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Community 0` be split into smaller, more focused modules?**
  _Cohesion score 0.027972027972027972 - nodes in this community are weakly interconnected._
- **Should `Community 1` be split into smaller, more focused modules?**
  _Cohesion score 0.05555555555555555 - nodes in this community are weakly interconnected._
- **Should `Community 2` be split into smaller, more focused modules?**
  _Cohesion score 0.050505050505050504 - nodes in this community are weakly interconnected._
- **Should `Community 3` be split into smaller, more focused modules?**
  _Cohesion score 0.052244897959183675 - nodes in this community are weakly interconnected._