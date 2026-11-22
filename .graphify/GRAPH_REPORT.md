# Graph Report - .  (2026-11-22)

## Corpus Check
- 85 files · ~151,258 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1463 nodes · 3067 edges · 72 communities detected
- Extraction: 100% EXTRACTED · 0% INFERRED · 0% AMBIGUOUS
- Token cost: 0 input · 0 output
- Edge kinds: calls: 1268 · contains: 1006 · method: 319 · MODIFIES: 280 · ON_BRANCH: 130 · PARENT_OF: 64


## Input Scope
- Requested: auto
- Resolved: committed (source: default-auto)
- Included files: 85 · Candidates: 99
- Excluded: 1 untracked · 11 ignored · 0 sensitive · 0 missing committed
- Recommendation: Use --scope all or graphify.yaml inputs.corpus for a knowledge-base folder.

## Graph Freshness
- Built from Git commit: `e7607dc`
- Compare this hash to `git rev-parse HEAD` before trusting freshness-sensitive graph output.
## God Nodes (most connected - your core abstractions)
1. `App` - 69 edges
2. `App` - 68 edges
3. `Ctx` - 62 edges
4. `Ctx` - 61 edges
5. `newEnv()` - 32 edges
6. `esc()` - 29 edges
7. `expect()` - 25 edges
8. `apiRaporMapel()` - 17 edges
9. `apiRaporMapelAI()` - 16 edges
10. `ico()` - 15 edges

## Surprising Connections (you probably didn't know these)
- `09479bb Perbarui peta kode graphify` --ON_BRANCH--> `claude/hello-230aaf`  [EXTRACTED]
  git → git  _Bridges community 28 → community 3_
- `1119b6b Tahap 3: mesin ujian kuis & CBT memakai hitung mundur, kirim otomatis, dan kartu hasil bersama` --ON_BRANCH--> `claude/hello-230aaf`  [EXTRACTED]
  git → git  _Bridges community 48 → community 3_
- `1bb9e1e v1.20.0: perbaikan hasil audit keamanan, ujian, query, dan tampilan` --ON_BRANCH--> `claude/hello-230aaf`  [EXTRACTED]
  git → git  _Bridges community 5 → community 3_
- `1bb9e1e v1.20.0: perbaikan hasil audit keamanan, ujian, query, dan tampilan` --PARENT_OF--> `3a5fb3d Perbarui peta kode graphify`  [EXTRACTED]
  git → git  _Bridges community 5 → community 23_
- `1e19524 Tahap 1: peserta LJK ikut riwayat kelas, validasi jam absensi, dasbor guru per tahun aktif, cache & halaman publik lebih aman, muat paralel` --ON_BRANCH--> `claude/hello-230aaf`  [EXTRACTED]
  git → git  _Bridges community 23 → community 3_

## Communities

### Community 0 - "Community 0"
Cohesion: 0.03
Nodes (46): acctItems(), acctMenu(), ACT, angka(), BOOT, BUSY, CHG, collect() (+38 more)

### Community 1 - "Community 1"
Cohesion: 0.05
Nodes (17): apiDeleteQuestion(), apiQuizAnswersAutosave(), apiQuizSave(), apiQuizStart(), apiQuizSubmit(), apiSaveQuestion(), quizClean(), quizDeadline() (+9 more)

### Community 2 - "Community 2"
Cohesion: 0.21
Nodes (40): client, env, fakeWA, client, env, fakeWA, expect(), newEnv() (+32 more)

### Community 3 - "Community 3"
Cohesion: 0.12
Nodes (47): claude/hello-230aaf, main, 01cf6c7 Perbarui peta kode graphify, 044f2aa v1.23.1: kerapian menu guru & siswa di HP: tombol absensi lebih besar + simpan menempel, kolom nama tetap terlihat di tabel rekap, kehadiran ditulis lengkap, 'Belum ada materi', 137792c Tahap 4: hapus 57 deklarasi CSS yang selalu ditimpa (tampilan terverifikasi identik di desktop & HP); peringatan keluar hanya saat ujian belum terkirim, 2f0cb47 Perbarui peta kode graphify, 306dce2 Perbarui peta kode graphify, 3234f5c v1.25.2: akhir semester ganjil diatur di form Tahun Ajaran; rentang semester tampil di daftar; pengaturan semester di e-Rapor dihapus (+39 more)

### Community 4 - "Community 4"
Cohesion: 0.07
Nodes (28): cbtAttempt, examRow, apiCBTAttemptDelete(), apiCBTDelete(), apiCBTHasil(), apiCBTList(), apiCBTSave(), apiCBTSaveAnswers() (+20 more)

### Community 5 - "Community 5"
Cohesion: 0.06
Nodes (5): 1bb9e1e v1.20.0: perbaikan hasil audit keamanan, ujian, query, dan tampilan, 3550bcd v1.24.1: audit tahap 1 - kuis dinilai dari soal yang ditampilkan, batas unggahan publik, tugas dinilai tak tertimpa, buku berriwayat tak dihapus, putus sesi atomik, petugas bilik e-voting tercatat, 5720746 SCH v1.18.0: sistem informasi sekolah (Go + SQLite), ec20e2b v1.25.1: statistik query SQLite selalu segar (PRAGMA optimize=0x10002 saat menyala & tiap jam), rekap lebih ringan, permintaan e-Rapor digabung, User

### Community 6 - "Community 6"
Cohesion: 0.08
Nodes (2): hashToken(), App

### Community 7 - "Community 7"
Cohesion: 0.09
Nodes (14): election, apiElectionResult(), apiElections(), apiElectionsForMe(), apiVote(), castVote(), apiElectionResult(), apiElections() (+6 more)

### Community 8 - "Community 8"
Cohesion: 0.07
Nodes (37): AGAMA, AI_JENJANG, AI_KELAS, AI_MAPEL, AI_PRESETS, AI_TPL, AI_TTD, aiActions() (+29 more)

### Community 9 - "Community 9"
Cohesion: 0.08
Nodes (30): absenCeklist(), absenRekap(), absenScanHtml(), AKSI, autoTag(), canEditEv(), doScan(), evRow() (+22 more)

### Community 10 - "Community 10"
Cohesion: 0.13
Nodes (19): raporLengkap, apiEkskulAnggotaSave(), apiEkskulCalon(), apiEkskulList(), apiEkskulNilaiSave(), apiEkskulSave(), apiRaporDocx(), apiRaporMapelSave() (+11 more)

### Community 11 - "Community 11"
Cohesion: 0.10
Nodes (35): applyTheme(), avatar(), brandHtml(), btn(), comboHtml(), comboItems(), cycleTheme(), empty() (+27 more)

### Community 12 - "Community 12"
Cohesion: 0.10
Nodes (2): cfgInt(), App

### Community 13 - "Community 13"
Cohesion: 0.10
Nodes (17): aiTemplate, aiAdminOnly(), aiCleanInputs(), aiKeyHint(), aiProviderError(), aiReadStream(), aiRequest(), aiUsageKey() (+9 more)

### Community 14 - "Community 14"
Cohesion: 0.12
Nodes (10): promoteKelas, promoter, promoteReq, apiPromote(), nextTingkat(), nullIfEmpty(), unlistedClasses(), promoteKelas (+2 more)

### Community 15 - "Community 15"
Cohesion: 0.11
Nodes (12): absenResult, apiAbsenHarian(), apiAbsenHarianSave(), apiAbsenRekap(), haversine(), init(), itoa(), jamValid() (+4 more)

### Community 16 - "Community 16"
Cohesion: 0.10
Nodes (5): apiDelete(), apiDeleteChild(), onlyDigits(), activeYear(), Ctx

### Community 17 - "Community 17"
Cohesion: 0.14
Nodes (12): sibiBook, apiSibiImport(), apiSibiSearch(), cacheName(), ebookCacheMax(), nullStr(), safeFileName(), sibiFetch() (+4 more)

### Community 18 - "Community 18"
Cohesion: 0.11
Nodes (2): tingkatDariNama(), Ctx

### Community 19 - "Community 19"
Cohesion: 0.11
Nodes (26): afterSave(), closeModal(), errBox(), formHtml(), go(), head(), isAdm(), judulKeHeader() (+18 more)

### Community 20 - "Community 20"
Cohesion: 0.10
Nodes (15): adminDash(), adminPengaturan(), fmtBytes(), KEYMAP, ROMAWI, SET_PAGES, SET_SECTIONS, setCard() (+7 more)

### Community 21 - "Community 21"
Cohesion: 0.13
Nodes (19): bookCard(), bookCover(), bookRead(), circBorrow(), circLog(), circMember(), circUpdate(), hasEbook() (+11 more)

### Community 22 - "Community 22"
Cohesion: 0.15
Nodes (8): apiJournal(), apiNoteDelete(), apiNotes(), apiNoteSave(), apiNoteStudents(), apiWaliClasses(), apiWaliDetail(), str2()

### Community 23 - "Community 23"
Cohesion: 0.11
Nodes (8): 1e19524 Tahap 1: peserta LJK ikut riwayat kelas, validasi jam absensi, dasbor guru per tahun aktif, cache & halaman publik lebih aman, muat paralel, 3a5fb3d Perbarui peta kode graphify, 450a912 v1.20.0: perbaiki sisa temuan audit dan risiko baru dari tinjauan ulang, 79219a2 v1.26.0: pagination bernomor seragam di 7 daftar besar (Siswa/Guru, Kelulusan, Catatan, Log, PPDB, Inventaris, Perpustakaan); saringan & cari di server; cetak/CSV tetap semua data, ae5fa1f Perbarui peta kode graphify, GRAD, BERKAS, PPDB_ST

### Community 24 - "Community 24"
Cohesion: 0.16
Nodes (7): apiWAConfigSave(), cfgOn(), cfgStr(), waFill(), waHost(), waNumber(), waSend()

### Community 25 - "Community 25"
Cohesion: 0.15
Nodes (9): periode, apiEkskulAnggota(), apiRaporKelas(), apiRaporMapel(), apiRaporMapelAI(), deskripsiKoku(), jumlahKuis(), saranCapaian() (+1 more)

### Community 26 - "Community 26"
Cohesion: 0.12
Nodes (5): apiPPDBInfo(), ppdbOpen(), randDigits(), apiPPDBInfo(), ppdbOpen()

### Community 27 - "Community 27"
Cohesion: 0.16
Nodes (10): addColumn(), apiGetConfig(), apiPutConfig(), apiPutModules(), getConfig(), init(), kvForget(), kvGet() (+2 more)

### Community 28 - "Community 28"
Cohesion: 0.13
Nodes (6): 09479bb Perbarui peta kode graphify, 2c481ea v1.22.1: perangkat ajar AI per jenjang (SD/SMP/SMA/SMK): kelas & fase, saran mapel, validasi server, instruksi jenjang ke AI, 4f7bbc2 Perbarui peta kode graphify, 69d9920 Perbarui peta kode graphify, 792dd82 v1.23.0: jenjang sekolah diatur sekali di Pengaturan; tingkat kelas, saran mapel, kenaikan kelas, filter e-book, dan Asisten AI mengikuti jenjang, c3fe8fe v1.22.0: RPP Kurikulum Merdeka terbaru (RPM/Modul Ajar/RPP 1 lembar, fase & alokasi waktu otomatis, cek waktu), sunting, cetak resmi berkop & tanda tangan, unduh Word

### Community 29 - "Community 29"
Cohesion: 0.21
Nodes (5): loginLimiter, clientIP(), isHTTPS(), limitKey(), viaProxy()

### Community 30 - "Community 30"
Cohesion: 0.18
Nodes (12): pageRapor(), raporHtml(), RP_EKSKUL, RP_KOKU, rpAtur(), rpEkskul(), rpGetRapor(), rpMapel() (+4 more)

### Community 31 - "Community 31"
Cohesion: 0.15
Nodes (4): kunciUjian(), TestCBTDinilaiOtomatisSetelahWaktuHabis(), TestCBTSoalTambahanTidakMenurunkanNilai(), TestKirimUlangSetelahDinilaiOtomatis()

### Community 32 - "Community 32"
Cohesion: 0.17
Nodes (5): bad(), apiVersion(), bad(), hashToken(), limitKey()

### Community 33 - "Community 33"
Cohesion: 0.16
Nodes (9): statusRecorder, apiAudit(), applyTimezone(), checkPassword(), dayRangeUTC(), init(), pageParams(), withBcrypt() (+1 more)

### Community 34 - "Community 34"
Cohesion: 0.30
Nodes (12): apiAIDocDocx(), buildDocx(), docxBody(), docxKop(), docxPara(), docxRuns(), docxTable(), docxTTD() (+4 more)

### Community 35 - "Community 35"
Cohesion: 0.18
Nodes (8): forbidden(), notFound(), writeErr(), apiErr, bisaDipaksa, forbidden(), apiErr, bisaDipaksa

### Community 36 - "Community 36"
Cohesion: 0.19
Nodes (8): JAM_KE, JENIS_CAT, KATEGORI_CAT, monthInput(), pageCatatan(), pageJurnal(), pageWaliKelas(), validBulan()

### Community 37 - "Community 37"
Cohesion: 0.20
Nodes (6): apiDeleteCourse(), apiDeleteUser(), apiSaveUser(), b2i(), userIn, userIn

### Community 38 - "Community 38"
Cohesion: 0.24
Nodes (8): apiGetSettings(), apiPutSettings(), apiSaveClass(), cleanDeadline(), cleanURL(), deref(), required(), str()

### Community 39 - "Community 39"
Cohesion: 0.14
Nodes (3): 334b045 v1.19.2: dropdown pilih-atau-ketik ramah HP, panah select seragam, Asisten AI satu kolom di HP, b86971e v1.19.1: mapel yang diampu guru diutamakan di Asisten AI dan CBT, c9b9c9d v1.19.0: impor e-book dari Buku Kemendikbud, PDF disimpan saat pertama dibuka

### Community 40 - "Community 40"
Cohesion: 0.23
Nodes (14): api(), apiFetch(), apiJSON(), checkVersion(), confirmBox(), DEL(), examSubmit(), GETpage() (+6 more)

### Community 41 - "Community 41"
Cohesion: 0.23
Nodes (10): findMarkers(), homography(), LJK, ljkPos(), ljkProcess(), ljkRead(), ljkRedraw(), ljkSheet() (+2 more)

### Community 42 - "Community 42"
Cohesion: 0.29
Nodes (11): noKelas(), siswaCourse(), siswaDash(), siswaData(), siswaKuis(), siswaMapel(), siswaNilai(), siswaTugas() (+3 more)

### Community 43 - "Community 43"
Cohesion: 0.19
Nodes (4): siswaID(), TestAbsensiMapelAlurLengkap(), TestPinjamKembaliBuku(), TestTugasBuatKumpulNilai()

### Community 44 - "Community 44"
Cohesion: 0.38
Nodes (11): courseBar(), guruAbsensi(), guruCourses(), guruKuis(), guruMapel(), guruMateri(), guruRekap(), guruTugas() (+3 more)

### Community 45 - "Community 45"
Cohesion: 0.33
Nodes (11): clientIP(), isHTTPS(), main(), newLoginLimiter(), newStaticHandler(), notFound(), printAddresses(), securityHeaders() (+3 more)

### Community 46 - "Community 46"
Cohesion: 0.26
Nodes (1): loginLimiter

### Community 47 - "Community 47"
Cohesion: 0.21
Nodes (12): cacheTTL(), clone(), GET(), jenjangDiatur(), kelasOpts(), mapelOpts(), mapelUmumJenjang(), saran() (+4 more)

### Community 48 - "Community 48"
Cohesion: 0.22
Nodes (6): 1119b6b Tahap 3: mesin ujian kuis & CBT memakai hitung mundur, kirim otomatis, dan kartu hasil bersama, bcd8537 Tahap 2: inti permintaan bersama + batal saat pindah menu, helper kamera/scan/password bersama, jendela dialog aksesibel, cbtTimeText(), pageCBT(), renderCbt(), siswaCBT()

### Community 50 - "Community 50"
Cohesion: 0.27
Nodes (6): aiJenjang(), rppFase(), rppMenitJP(), rppNum(), rppPrompt(), rppWaktu()

### Community 51 - "Community 51"
Cohesion: 0.33
Nodes (4): apiKalender(), jamLokal(), localDay(), tglJamLokal()

### Community 52 - "Community 52"
Cohesion: 0.24
Nodes (3): pngDataURL(), TestPPDBDitolakTidakMenyimpanBerkas(), TestUnggahanPublikDibatasiHarian()

### Community 53 - "Community 53"
Cohesion: 0.33
Nodes (8): bilikHtml(), bilikOpen(), bilikScan(), elStatus(), evoteVoter(), hasilHtml(), hasilUtama(), pageEvoting()

### Community 54 - "Community 54"
Cohesion: 0.31
Nodes (8): skorMapel, tugasNilai, nilaiAkhir(), rata(), skorDariRekap(), toF(), skorMapel, tugasNilai

### Community 55 - "Community 55"
Cohesion: 0.33
Nodes (7): xlsxSheet, xlsxSheet, angkaXlsx(), buildXlsx(), kolomXlsx(), namaSheetXlsx(), sheetXML()

### Community 56 - "Community 56"
Cohesion: 0.36
Nodes (8): cartTotal(), kasirBarang(), kasirLaporan(), pageKasir(), posAdd(), renderCart(), showStruk(), strukHtml()

### Community 57 - "Community 57"
Cohesion: 0.29
Nodes (2): apiRekapMe(), nilaiSiswa()

### Community 58 - "Community 58"
Cohesion: 0.36
Nodes (4): ensureAdmin(), hashPassword(), nowStr(), randomPassword()

### Community 59 - "Community 59"
Cohesion: 0.36
Nodes (4): apiPortoCatatan(), apiPortoDelete(), apiPortoList(), apiRaporTugasku()

### Community 60 - "Community 60"
Cohesion: 0.36
Nodes (6): mcq, apiCBTSoalBatch(), apiQuizQuestionBatch(), parseBatch(), parseMCQ(), mcq

### Community 61 - "Community 61"
Cohesion: 0.29
Nodes (2): apiSale(), apiSell()

### Community 62 - "Community 62"
Cohesion: 0.29
Nodes (4): nextYearName(), NK_AKSI, pageNaik(), ROMAN

### Community 64 - "Community 64"
Cohesion: 0.38
Nodes (4): AUDIT_LABELS, GETpage(), moreBtn(), pageAudit()

### Community 65 - "Community 65"
Cohesion: 0.40
Nodes (4): migration, defaultTahunAjaran(), migTahunAjaran(), migration

### Community 67 - "Community 67"
Cohesion: 0.67
Nodes (5): f(), fmtSkor(), sama(), TestNilaiAkhirDiRekapDanSiswaSama(), TestRumusNilaiAkhir()

### Community 68 - "Community 68"
Cohesion: 0.40
Nodes (2): staticFile, staticFile

### Community 69 - "Community 69"
Cohesion: 0.40
Nodes (5): examTime(), monthNow(), pad(), today(), toLocalInput()

### Community 70 - "Community 70"
Cohesion: 0.40
Nodes (5): lampHtml(), logoSrc(), safeUrl(), schoolHead(), ytId()

### Community 71 - "Community 71"
Cohesion: 0.50
Nodes (2): manUjianDetail(), pageManUjian()

### Community 72 - "Community 72"
Cohesion: 0.50
Nodes (2): apiReturn(), perpusCfg()

### Community 73 - "Community 73"
Cohesion: 0.50
Nodes (2): KODE_SURAT, SURAT_ST

### Community 74 - "Community 74"
Cohesion: 0.67
Nodes (1): User

## Knowledge Gaps
- **118 isolated node(s):** `lampiran`, `userIn`, `docxOpts`, `migration`, `absenResult` (+113 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `Community 6`** (2 nodes): `hashToken()`, `App`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 12`** (2 nodes): `cfgInt()`, `App`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 18`** (2 nodes): `tingkatDariNama()`, `Ctx`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 46`** (1 nodes): `loginLimiter`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 57`** (2 nodes): `apiRekapMe()`, `nilaiSiswa()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 61`** (2 nodes): `apiSale()`, `apiSell()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 68`** (2 nodes): `staticFile`, `staticFile`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 71`** (2 nodes): `manUjianDetail()`, `pageManUjian()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 72`** (2 nodes): `apiReturn()`, `perpusCfg()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 73`** (2 nodes): `KODE_SURAT`, `SURAT_ST`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 74`** (1 nodes): `User`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App` connect `Community 12` to `Community 5`, `Community 15`, `Community 17`, `Community 72`, `Community 26`, `Community 32`, `Community 7`, `Community 24`, `Community 46`, `Community 4`, `Community 1`, `Community 35`, `Community 34`, `Community 27`, `Community 16`, `Community 38`, `Community 45`?**
  _High betweenness centrality (0.077) - this node is a cross-community bridge._
- **Why does `App` connect `Community 6` to `Community 5`, `Community 15`, `Community 17`, `Community 72`, `Community 26`, `Community 7`, `Community 24`, `Community 4`, `Community 1`, `Community 35`, `Community 34`, `Community 27`, `Community 16`, `Community 38`?**
  _High betweenness centrality (0.074) - this node is a cross-community bridge._
- **Why does `Ctx` connect `Community 16` to `Community 35`, `Community 25`, `Community 13`, `Community 51`, `Community 32`, `Community 26`, `Community 59`, `Community 10`, `Community 27`, `Community 7`, `Community 22`, `Community 18`, `Community 37`, `Community 46`, `Community 15`, `Community 4`, `Community 1`, `Community 45`?**
  _High betweenness centrality (0.066) - this node is a cross-community bridge._
- **What connects `lampiran`, `userIn`, `docxOpts` to the rest of the system?**
  _118 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Community 0` be split into smaller, more focused modules?**
  _Cohesion score 0.027034969144872172 - nodes in this community are weakly interconnected._
- **Should `Community 1` be split into smaller, more focused modules?**
  _Cohesion score 0.050314465408805034 - nodes in this community are weakly interconnected._
- **Should `Community 3` be split into smaller, more focused modules?**
  _Cohesion score 0.11933395004625347 - nodes in this community are weakly interconnected._