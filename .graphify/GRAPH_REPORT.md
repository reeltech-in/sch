# Graph Report - .  (2026-11-23)

## Corpus Check
- 91 files · ~158,334 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1499 nodes · 3159 edges · 70 communities detected
- Extraction: 100% EXTRACTED · 0% INFERRED · 0% AMBIGUOUS
- Token cost: 0 input · 0 output
- Edge kinds: calls: 1289 · contains: 1029 · method: 325 · MODIFIES: 312 · ON_BRANCH: 135 · PARENT_OF: 69


## Input Scope
- Requested: auto
- Resolved: committed (source: default-auto)
- Included files: 91 · Candidates: 108
- Excluded: 0 untracked · 127 ignored · 0 sensitive · 0 missing committed
- Recommendation: Use --scope all or graphify.yaml inputs.corpus for a knowledge-base folder.

## Graph Freshness
- Built from Git commit: `cea249f`
- Compare this hash to `git rev-parse HEAD` before trusting freshness-sensitive graph output.
## God Nodes (most connected - your core abstractions)
1. `App` - 69 edges
2. `App` - 68 edges
3. `Ctx` - 62 edges
4. `Ctx` - 62 edges
5. `newEnv()` - 32 edges
6. `esc()` - 29 edges
7. `expect()` - 25 edges
8. `apiRaporMapel()` - 17 edges
9. `apiRaporMapelAI()` - 16 edges
10. `ico()` - 15 edges

## Surprising Connections (you probably didn't know these)
- `apiPutSettings()` --calls--> `apiGetSettings()`  [EXTRACTED]
  api.go → api.go  _Bridges community 46 → community 73_
- `09479bb Perbarui peta kode graphify` --ON_BRANCH--> `claude/hello-230aaf`  [EXTRACTED]
  git → git  _Bridges community 27 → community 3_
- `1bb9e1e v1.20.0: perbaikan hasil audit keamanan, ujian, query, dan tampilan` --ON_BRANCH--> `claude/hello-230aaf`  [EXTRACTED]
  git → git  _Bridges community 31 → community 3_
- `1bb9e1e v1.20.0: perbaikan hasil audit keamanan, ujian, query, dan tampilan` --PARENT_OF--> `3a5fb3d Perbarui peta kode graphify`  [EXTRACTED]
  git → git  _Bridges community 31 → community 26_
- `1e19524 Tahap 1: peserta LJK ikut riwayat kelas, validasi jam absensi, dasbor guru per tahun aktif, cache & halaman publik lebih aman, muat paralel` --ON_BRANCH--> `claude/hello-230aaf`  [EXTRACTED]
  git → git  _Bridges community 26 → community 3_

## Communities

### Community 0 - "Community 0"
Cohesion: 0.03
Nodes (47): acctItems(), acctMenu(), ACT, angka(), BOOT, BUSY, CHG, collect() (+39 more)

### Community 1 - "Community 1"
Cohesion: 0.05
Nodes (17): apiDeleteQuestion(), apiQuizAnswersAutosave(), apiQuizSave(), apiQuizStart(), apiQuizSubmit(), apiSaveQuestion(), quizClean(), quizDeadline() (+9 more)

### Community 2 - "Community 2"
Cohesion: 0.06
Nodes (29): cekPassword(), hashToken(), bad(), forbidden(), notFound(), writeErr(), apiErr, bisaDipaksa (+21 more)

### Community 3 - "Community 3"
Cohesion: 0.10
Nodes (52): claude/hello-230aaf, main, 01cf6c7 Perbarui peta kode graphify, 044f2aa v1.23.1: kerapian menu guru & siswa di HP: tombol absensi lebih besar + simpan menempel, kolom nama tetap terlihat di tabel rekap, kehadiran ditulis lengkap, 'Belum ada materi', 1119b6b Tahap 3: mesin ujian kuis & CBT memakai hitung mundur, kirim otomatis, dan kartu hasil bersama, 1202eb2 v1.28.0: soal AI bergambar & berstimulus seperti LKS; gambar di soal kuis dan file Word, 137792c Tahap 4: hapus 57 deklarasi CSS yang selalu ditimpa (tampilan terverifikasi identik di desktop & HP); peringatan keluar hanya saat ujian belum terkirim, 2f0cb47 Perbarui peta kode graphify (+44 more)

### Community 4 - "Community 4"
Cohesion: 0.21
Nodes (40): client, env, fakeWA, client, env, fakeWA, expect(), newEnv() (+32 more)

### Community 5 - "Community 5"
Cohesion: 0.07
Nodes (28): cbtAttempt, examRow, apiCBTAttemptDelete(), apiCBTDelete(), apiCBTHasil(), apiCBTList(), apiCBTSave(), apiCBTSaveAnswers() (+20 more)

### Community 6 - "Community 6"
Cohesion: 0.09
Nodes (14): election, apiElectionResult(), apiElections(), apiElectionsForMe(), apiVote(), castVote(), apiElectionResult(), apiElections() (+6 more)

### Community 7 - "Community 7"
Cohesion: 0.07
Nodes (37): AGAMA, AI_JENJANG, AI_KELAS, AI_MAPEL, AI_PRESETS, AI_TPL, AI_TTD, aiActions() (+29 more)

### Community 8 - "Community 8"
Cohesion: 0.09
Nodes (1): App

### Community 9 - "Community 9"
Cohesion: 0.09
Nodes (4): cfgInt(), apiReturn(), perpusCfg(), App

### Community 10 - "Community 10"
Cohesion: 0.12
Nodes (19): raporLengkap, apiEkskulAnggotaSave(), apiEkskulCalon(), apiEkskulList(), apiEkskulNilaiSave(), apiEkskulSave(), apiRaporDocx(), apiRaporMapelSave() (+11 more)

### Community 11 - "Community 11"
Cohesion: 0.08
Nodes (30): absenCeklist(), absenRekap(), absenScanHtml(), AKSI, autoTag(), canEditEv(), doScan(), evRow() (+22 more)

### Community 12 - "Community 12"
Cohesion: 0.11
Nodes (6): loginLimiter, clientIP(), isHTTPS(), limitKey(), viaProxy(), loginLimiter

### Community 13 - "Community 13"
Cohesion: 0.10
Nodes (17): aiTemplate, aiAdminOnly(), aiCleanInputs(), aiKeyHint(), aiProviderError(), aiReadStream(), aiRequest(), aiUsageKey() (+9 more)

### Community 14 - "Community 14"
Cohesion: 0.12
Nodes (32): avatar(), brandHtml(), btn(), copyrightHtml(), empty(), ensureThemeFloat(), esc(), examResultHtml() (+24 more)

### Community 15 - "Community 15"
Cohesion: 0.12
Nodes (10): promoteKelas, promoter, promoteReq, apiPromote(), nextTingkat(), nullIfEmpty(), unlistedClasses(), promoteKelas (+2 more)

### Community 16 - "Community 16"
Cohesion: 0.08
Nodes (6): 3550bcd v1.24.1: audit tahap 1 - kuis dinilai dari soal yang ditampilkan, batas unggahan publik, tugas dinilai tak tertimpa, buku berriwayat tak dihapus, putus sesi atomik, petugas bilik e-voting tercatat, 5720746 SCH v1.18.0: sistem informasi sekolah (Go + SQLite), be1dae1 v1.28.1: perbaikan keamanan (login serentak, aturan password, berkas PPDB, formulir publik, e-voting), cea249f Perbarui peta kode graphify, userIn, userIn

### Community 17 - "Community 17"
Cohesion: 0.11
Nodes (12): absenResult, apiAbsenHarian(), apiAbsenHarianSave(), apiAbsenRekap(), haversine(), init(), itoa(), jamValid() (+4 more)

### Community 18 - "Community 18"
Cohesion: 0.14
Nodes (12): sibiBook, apiSibiImport(), apiSibiSearch(), cacheName(), ebookCacheMax(), nullStr(), safeFileName(), sibiFetch() (+4 more)

### Community 19 - "Community 19"
Cohesion: 0.10
Nodes (15): adminDash(), adminPengaturan(), fmtBytes(), KEYMAP, ROMAWI, SET_PAGES, SET_SECTIONS, setCard() (+7 more)

### Community 20 - "Community 20"
Cohesion: 0.11
Nodes (4): apiDelete(), apiDeleteChild(), onlyDigits(), Ctx

### Community 21 - "Community 21"
Cohesion: 0.13
Nodes (19): bookCard(), bookCover(), bookRead(), circBorrow(), circLog(), circMember(), circUpdate(), hasEbook() (+11 more)

### Community 22 - "Community 22"
Cohesion: 0.11
Nodes (3): tingkatDariNama(), Ctx, activeYear()

### Community 23 - "Community 23"
Cohesion: 0.15
Nodes (9): apiWAConfigSave(), cfgOn(), cfgStr(), waFill(), waHost(), waNama(), waNumber(), waOneLine() (+1 more)

### Community 24 - "Community 24"
Cohesion: 0.14
Nodes (9): periode, apiEkskulAnggota(), apiRaporKelas(), apiRaporMapel(), apiRaporMapelAI(), deskripsiKoku(), jumlahKuis(), saranCapaian() (+1 more)

### Community 25 - "Community 25"
Cohesion: 0.15
Nodes (8): apiJournal(), apiNoteDelete(), apiNotes(), apiNoteSave(), apiNoteStudents(), apiWaliClasses(), apiWaliDetail(), str2()

### Community 26 - "Community 26"
Cohesion: 0.11
Nodes (8): 1e19524 Tahap 1: peserta LJK ikut riwayat kelas, validasi jam absensi, dasbor guru per tahun aktif, cache & halaman publik lebih aman, muat paralel, 3a5fb3d Perbarui peta kode graphify, 450a912 v1.20.0: perbaiki sisa temuan audit dan risiko baru dari tinjauan ulang, 79219a2 v1.26.0: pagination bernomor seragam di 7 daftar besar (Siswa/Guru, Kelulusan, Catatan, Log, PPDB, Inventaris, Perpustakaan); saringan & cari di server; cetak/CSV tetap semua data, ae5fa1f Perbarui peta kode graphify, GRAD, BERKAS, PPDB_ST

### Community 27 - "Community 27"
Cohesion: 0.11
Nodes (6): 09479bb Perbarui peta kode graphify, 2c481ea v1.22.1: perangkat ajar AI per jenjang (SD/SMP/SMA/SMK): kelas & fase, saran mapel, validasi server, instruksi jenjang ke AI, 4f7bbc2 Perbarui peta kode graphify, 69d9920 Perbarui peta kode graphify, 792dd82 v1.23.0: jenjang sekolah diatur sekali di Pengaturan; tingkat kelas, saran mapel, kenaikan kelas, filter e-book, dan Asisten AI mengikuti jenjang, c3fe8fe v1.22.0: RPP Kurikulum Merdeka terbaru (RPM/Modul Ajar/RPP 1 lembar, fase & alokasi waktu otomatis, cek waktu), sunting, cetak resmi berkop & tanda tangan, unduh Word

### Community 28 - "Community 28"
Cohesion: 0.12
Nodes (5): apiPPDBInfo(), ppdbOpen(), randDigits(), apiPPDBInfo(), ppdbOpen()

### Community 29 - "Community 29"
Cohesion: 0.16
Nodes (10): addColumn(), apiGetConfig(), apiPutConfig(), apiPutModules(), getConfig(), init(), kvForget(), kvGet() (+2 more)

### Community 30 - "Community 30"
Cohesion: 0.13
Nodes (19): afterSave(), closeModal(), errBox(), go(), head(), isAdm(), judulKeHeader(), loadScript() (+11 more)

### Community 31 - "Community 31"
Cohesion: 0.12
Nodes (3): 1bb9e1e v1.20.0: perbaikan hasil audit keamanan, ujian, query, dan tampilan, ec20e2b v1.25.1: statistik query SQLite selalu segar (PRAGMA optimize=0x10002 saat menyala & tiap jam), rekap lebih ringan, permintaan e-Rapor digabung, User

### Community 32 - "Community 32"
Cohesion: 0.18
Nodes (17): api(), apiFetch(), apiJSON(), applyTheme(), checkVersion(), confirmBox(), cycleTheme(), DEL() (+9 more)

### Community 33 - "Community 33"
Cohesion: 0.18
Nodes (12): pageRapor(), raporHtml(), RP_EKSKUL, RP_KOKU, rpAtur(), rpEkskul(), rpGetRapor(), rpMapel() (+4 more)

### Community 34 - "Community 34"
Cohesion: 0.15
Nodes (4): kunciUjian(), TestCBTDinilaiOtomatisSetelahWaktuHabis(), TestCBTSoalTambahanTidakMenurunkanNilai(), TestKirimUlangSetelahDinilaiOtomatis()

### Community 35 - "Community 35"
Cohesion: 0.16
Nodes (9): statusRecorder, apiAudit(), applyTimezone(), checkPassword(), dayRangeUTC(), init(), pageParams(), withBcrypt() (+1 more)

### Community 36 - "Community 36"
Cohesion: 0.15
Nodes (4): bukaPPDB(), TestBerkasPPDBHanyaUntukPengelola(), TestPPDBDuplikatTidakMembocorkanNomor(), TestPPDBWhatsAppTidakBisaDisalahgunakan()

### Community 37 - "Community 37"
Cohesion: 0.30
Nodes (12): apiAIDocDocx(), buildDocx(), docxBody(), docxKop(), docxPara(), docxRuns(), docxTable(), docxTTD() (+4 more)

### Community 38 - "Community 38"
Cohesion: 0.19
Nodes (8): JAM_KE, JENIS_CAT, KATEGORI_CAT, monthInput(), pageCatatan(), pageJurnal(), pageWaliKelas(), validBulan()

### Community 39 - "Community 39"
Cohesion: 0.23
Nodes (10): findMarkers(), homography(), LJK, ljkPos(), ljkProcess(), ljkRead(), ljkRedraw(), ljkSheet() (+2 more)

### Community 40 - "Community 40"
Cohesion: 0.29
Nodes (11): noKelas(), siswaCourse(), siswaDash(), siswaData(), siswaKuis(), siswaMapel(), siswaNilai(), siswaTugas() (+3 more)

### Community 41 - "Community 41"
Cohesion: 0.19
Nodes (4): siswaID(), TestAbsensiMapelAlurLengkap(), TestPinjamKembaliBuku(), TestTugasBuatKumpulNilai()

### Community 42 - "Community 42"
Cohesion: 0.15
Nodes (13): comboHtml(), comboItems(), examTime(), fieldHtml(), monthNow(), opts(), pad(), pagerNav() (+5 more)

### Community 43 - "Community 43"
Cohesion: 0.38
Nodes (11): courseBar(), guruAbsensi(), guruCourses(), guruKuis(), guruMapel(), guruMateri(), guruRekap(), guruTugas() (+3 more)

### Community 44 - "Community 44"
Cohesion: 0.17
Nodes (1): c9b9c9d v1.19.0: impor e-book dari Buku Kemendikbud, PDF disimpan saat pertama dibuka

### Community 45 - "Community 45"
Cohesion: 0.21
Nodes (12): cacheTTL(), clone(), GET(), jenjangDiatur(), kelasOpts(), mapelOpts(), mapelUmumJenjang(), saran() (+4 more)

### Community 46 - "Community 46"
Cohesion: 0.33
Nodes (7): apiPutSettings(), apiSaveClass(), cleanDeadline(), cleanURL(), deref(), required(), str()

### Community 47 - "Community 47"
Cohesion: 0.22
Nodes (6): 334b045 v1.19.2: dropdown pilih-atau-ketik ramah HP, panah select seragam, Asisten AI satu kolom di HP, b86971e v1.19.1: mapel yang diampu guru diutamakan di Asisten AI dan CBT, cbtTimeText(), pageCBT(), renderCbt(), siswaCBT()

### Community 49 - "Community 49"
Cohesion: 0.27
Nodes (6): aiJenjang(), rppFase(), rppMenitJP(), rppNum(), rppPrompt(), rppWaktu()

### Community 50 - "Community 50"
Cohesion: 0.33
Nodes (4): apiKalender(), jamLokal(), localDay(), tglJamLokal()

### Community 51 - "Community 51"
Cohesion: 0.24
Nodes (3): pngDataURL(), TestPPDBDitolakTidakMenyimpanBerkas(), TestUnggahanPublikDibatasiHarian()

### Community 52 - "Community 52"
Cohesion: 0.33
Nodes (8): bilikHtml(), bilikOpen(), bilikScan(), elStatus(), evoteVoter(), hasilHtml(), hasilUtama(), pageEvoting()

### Community 53 - "Community 53"
Cohesion: 0.31
Nodes (8): skorMapel, tugasNilai, nilaiAkhir(), rata(), skorDariRekap(), toF(), skorMapel, tugasNilai

### Community 54 - "Community 54"
Cohesion: 0.33
Nodes (7): xlsxSheet, xlsxSheet, angkaXlsx(), buildXlsx(), kolomXlsx(), namaSheetXlsx(), sheetXML()

### Community 55 - "Community 55"
Cohesion: 0.36
Nodes (8): cartTotal(), kasirBarang(), kasirLaporan(), pageKasir(), posAdd(), renderCart(), showStruk(), strukHtml()

### Community 56 - "Community 56"
Cohesion: 0.29
Nodes (2): apiRekapMe(), nilaiSiswa()

### Community 57 - "Community 57"
Cohesion: 0.36
Nodes (4): ensureAdmin(), hashPassword(), nowStr(), randomPassword()

### Community 58 - "Community 58"
Cohesion: 0.36
Nodes (4): apiPortoCatatan(), apiPortoDelete(), apiPortoList(), apiRaporTugasku()

### Community 59 - "Community 59"
Cohesion: 0.36
Nodes (6): mcq, apiCBTSoalBatch(), apiQuizQuestionBatch(), parseBatch(), parseMCQ(), mcq

### Community 60 - "Community 60"
Cohesion: 0.29
Nodes (2): apiSale(), apiSell()

### Community 61 - "Community 61"
Cohesion: 0.29
Nodes (4): nextYearName(), NK_AKSI, pageNaik(), ROMAN

### Community 63 - "Community 63"
Cohesion: 0.38
Nodes (4): AUDIT_LABELS, GETpage(), moreBtn(), pageAudit()

### Community 64 - "Community 64"
Cohesion: 0.40
Nodes (4): migration, defaultTahunAjaran(), migTahunAjaran(), migration

### Community 66 - "Community 66"
Cohesion: 0.67
Nodes (5): f(), fmtSkor(), sama(), TestNilaiAkhirDiRekapDanSiswaSama(), TestRumusNilaiAkhir()

### Community 67 - "Community 67"
Cohesion: 0.40
Nodes (5): lampHtml(), logoSrc(), safeUrl(), schoolHead(), ytId()

### Community 68 - "Community 68"
Cohesion: 0.50
Nodes (2): manUjianDetail(), pageManUjian()

### Community 69 - "Community 69"
Cohesion: 0.67
Nodes (2): apiDeleteCourse(), apiDeleteUser()

### Community 70 - "Community 70"
Cohesion: 0.83
Nodes (2): apiSaveUser(), b2i()

### Community 72 - "Community 72"
Cohesion: 0.50
Nodes (2): KODE_SURAT, SURAT_ST

### Community 73 - "Community 73"
Cohesion: 0.67
Nodes (1): apiGetSettings()

## Knowledge Gaps
- **119 isolated node(s):** `lampiran`, `userIn`, `docxOpts`, `migration`, `absenResult` (+114 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `Community 8`** (1 nodes): `App`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 44`** (1 nodes): `c9b9c9d v1.19.0: impor e-book dari Buku Kemendikbud, PDF disimpan saat pertama dibuka`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 56`** (2 nodes): `apiRekapMe()`, `nilaiSiswa()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 60`** (2 nodes): `apiSale()`, `apiSell()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 68`** (2 nodes): `manUjianDetail()`, `pageManUjian()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 69`** (2 nodes): `apiDeleteCourse()`, `apiDeleteUser()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 70`** (2 nodes): `apiSaveUser()`, `b2i()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 72`** (2 nodes): `KODE_SURAT`, `SURAT_ST`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 73`** (1 nodes): `apiGetSettings()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App` connect `Community 9` to `Community 31`, `Community 17`, `Community 18`, `Community 28`, `Community 2`, `Community 6`, `Community 23`, `Community 12`, `Community 5`, `Community 1`, `Community 37`, `Community 29`, `Community 20`, `Community 73`?**
  _High betweenness centrality (0.074) - this node is a cross-community bridge._
- **Why does `App` connect `Community 8` to `Community 31`, `Community 17`, `Community 18`, `Community 9`, `Community 2`, `Community 28`, `Community 6`, `Community 23`, `Community 5`, `Community 1`, `Community 37`, `Community 29`, `Community 20`, `Community 73`?**
  _High betweenness centrality (0.071) - this node is a cross-community bridge._
- **Why does `Ctx` connect `Community 20` to `Community 2`, `Community 24`, `Community 13`, `Community 50`, `Community 28`, `Community 58`, `Community 10`, `Community 29`, `Community 6`, `Community 25`, `Community 22`, `Community 69`, `Community 12`, `Community 17`, `Community 5`, `Community 1`, `Community 70`?**
  _High betweenness centrality (0.062) - this node is a cross-community bridge._
- **What connects `lampiran`, `userIn`, `docxOpts` to the rest of the system?**
  _119 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Community 0` be split into smaller, more focused modules?**
  _Cohesion score 0.02667814113597246 - nodes in this community are weakly interconnected._
- **Should `Community 1` be split into smaller, more focused modules?**
  _Cohesion score 0.050314465408805034 - nodes in this community are weakly interconnected._
- **Should `Community 2` be split into smaller, more focused modules?**
  _Cohesion score 0.061495457721872815 - nodes in this community are weakly interconnected._