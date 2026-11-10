# Graph Report - .  (2026-11-10)

## Corpus Check
- 80 files · ~129,818 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1139 nodes · 2160 edges · 63 communities detected
- Extraction: 100% EXTRACTED · 0% INFERRED · 0% AMBIGUOUS
- Token cost: 0 input · 0 output
- Edge kinds: contains: 885 · calls: 816 · MODIFIES: 229 · method: 143 · ON_BRANCH: 44 · PARENT_OF: 43


## Input Scope
- Requested: auto
- Resolved: committed (source: default-auto)
- Included files: 80 · Candidates: 94
- Excluded: 0 untracked · 389 ignored · 0 sensitive · 0 missing committed
- Recommendation: Use --scope all or graphify.yaml inputs.corpus for a knowledge-base folder.

## Graph Freshness
- Built from Git commit: `b729d3d`
- Compare this hash to `git rev-parse HEAD` before trusting freshness-sensitive graph output.
## God Nodes (most connected - your core abstractions)
1. `App` - 68 edges
2. `Ctx` - 47 edges
3. `newEnv()` - 32 edges
4. `esc()` - 29 edges
5. `expect()` - 25 edges
6. `ico()` - 15 edges
7. `loginLimiter` - 11 edges
8. `route()` - 11 edges
9. `api()` - 10 edges
10. `apiAIGenerate()` - 9 edges

## Surprising Connections (you probably didn't know these)
- `09479bb Perbarui peta kode graphify` --ON_BRANCH--> `main`  [EXTRACTED]
  git → git  _Bridges community 29 → community 9_
- `1bb9e1e v1.20.0: perbaikan hasil audit keamanan, ujian, query, dan tampilan` --ON_BRANCH--> `main`  [EXTRACTED]
  git → git  _Bridges community 2 → community 9_
- `1bb9e1e v1.20.0: perbaikan hasil audit keamanan, ujian, query, dan tampilan` --PARENT_OF--> `3a5fb3d Perbarui peta kode graphify`  [EXTRACTED]
  git → git  _Bridges community 2 → community 17_
- `1e19524 Tahap 1: peserta LJK ikut riwayat kelas, validasi jam absensi, dasbor guru per tahun aktif, cache & halaman publik lebih aman, muat paralel` --ON_BRANCH--> `main`  [EXTRACTED]
  git → git  _Bridges community 17 → community 9_
- `1e19524 Tahap 1: peserta LJK ikut riwayat kelas, validasi jam absensi, dasbor guru per tahun aktif, cache & halaman publik lebih aman, muat paralel` --PARENT_OF--> `bcd8537 Tahap 2: inti permintaan bersama + batal saat pindah menu, helper kamera/scan/password bersama, jendela dialog aksesibel`  [EXTRACTED]
  git → git  _Bridges community 17 → community 26_

## Communities

### Community 0 - "Community 0"
Cohesion: 0.03
Nodes (39): acctItems(), acctMenu(), ACT, BOOT, BUSY, CHG, collect(), COMBO (+31 more)

### Community 1 - "Community 1"
Cohesion: 0.05
Nodes (16): apiDeleteQuestion(), apiQuizAnswersAutosave(), apiQuizSave(), apiQuizStart(), apiQuizSubmit(), apiSaveQuestion(), quizClean(), quizDeadline() (+8 more)

### Community 2 - "Community 2"
Cohesion: 0.06
Nodes (38): 1bb9e1e v1.20.0: perbaikan hasil audit keamanan, ujian, query, dan tampilan, 334b045 v1.19.2: dropdown pilih-atau-ketik ramah HP, panah select seragam, Asisten AI satu kolom di HP, b86971e v1.19.1: mapel yang diampu guru diutamakan di Asisten AI dan CBT, AGAMA, AI_JENJANG, AI_KELAS, AI_MAPEL, AI_PRESETS (+30 more)

### Community 3 - "Community 3"
Cohesion: 0.20
Nodes (37): client, env, fakeWA, expect(), newEnv(), newFakeWA(), TestAdmGuruWaliDanCatatan(), TestCBTBatasWaktuDiServer() (+29 more)

### Community 4 - "Community 4"
Cohesion: 0.08
Nodes (26): apiCBTAttemptDelete(), apiCBTDelete(), apiCBTHasil(), apiCBTList(), apiCBTSave(), apiCBTSaveAnswers(), apiCBTSoal(), apiCBTSoalDelete() (+18 more)

### Community 5 - "Community 5"
Cohesion: 0.09
Nodes (36): applyTheme(), avatar(), brandHtml(), btn(), cycleTheme(), empty(), ensureThemeFloat(), esc() (+28 more)

### Community 6 - "Community 6"
Cohesion: 0.10
Nodes (13): apiElectionResult(), apiElections(), apiElectionsForMe(), apiVote(), castVote(), apiElectionResult(), apiElections(), apiElectionsForMe() (+5 more)

### Community 7 - "Community 7"
Cohesion: 0.09
Nodes (24): absenCeklist(), absenRekap(), absenScanHtml(), AKSI, autoTag(), canEditEv(), doScan(), evRow() (+16 more)

### Community 8 - "Community 8"
Cohesion: 0.13
Nodes (1): App

### Community 9 - "Community 9"
Cohesion: 0.13
Nodes (25): main, 01cf6c7 Perbarui peta kode graphify, 044f2aa v1.23.1: kerapian menu guru & siswa di HP: tombol absensi lebih besar + simpan menempel, kolom nama tetap terlihat di tabel rekap, kehadiran ditulis lengkap, 'Belum ada materi', 1119b6b Tahap 3: mesin ujian kuis & CBT memakai hitung mundur, kirim otomatis, dan kartu hasil bersama, 137792c Tahap 4: hapus 57 deklarasi CSS yang selalu ditimpa (tampilan terverifikasi identik di desktop & HP); peringatan keluar hanya saat ujian belum terkirim, 306dce2 Perbarui peta kode graphify, 3cd8909 Perbarui peta kode graphify, 3e27651 Perbarui peta kode graphify (+17 more)

### Community 10 - "Community 10"
Cohesion: 0.13
Nodes (16): aiAdminOnly(), aiCleanInputs(), aiKeyHint(), aiProviderError(), aiReadStream(), aiRequest(), aiUsageKey(), apiAIConfig() (+8 more)

### Community 11 - "Community 11"
Cohesion: 0.10
Nodes (15): adminDash(), adminPengaturan(), fmtBytes(), KEYMAP, ROMAWI, SET_PAGES, SET_SECTIONS, setCard() (+7 more)

### Community 12 - "Community 12"
Cohesion: 0.12
Nodes (25): afterSave(), closeModal(), errBox(), formHtml(), go(), head(), isAdm(), judulKeHeader() (+17 more)

### Community 13 - "Community 13"
Cohesion: 0.13
Nodes (19): bookCard(), bookCover(), bookRead(), circBorrow(), circLog(), circMember(), circUpdate(), hasEbook() (+11 more)

### Community 14 - "Community 14"
Cohesion: 0.10
Nodes (5): apiPPDBInfo(), ppdbOpen(), randDigits(), apiPPDBInfo(), ppdbOpen()

### Community 15 - "Community 15"
Cohesion: 0.13
Nodes (11): apiAbsenHarian(), apiAbsenHarianSave(), apiAbsenRekap(), haversine(), init(), itoa(), jamValid(), localClock() (+3 more)

### Community 16 - "Community 16"
Cohesion: 0.15
Nodes (8): activeYear(), apiPromote(), nextTingkat(), nullIfEmpty(), unlistedClasses(), promoteKelas, promoter, promoteReq

### Community 17 - "Community 17"
Cohesion: 0.10
Nodes (8): 1e19524 Tahap 1: peserta LJK ikut riwayat kelas, validasi jam absensi, dasbor guru per tahun aktif, cache & halaman publik lebih aman, muat paralel, 3550bcd v1.24.1: audit tahap 1 - kuis dinilai dari soal yang ditampilkan, batas unggahan publik, tugas dinilai tak tertimpa, buku berriwayat tak dihapus, putus sesi atomik, petugas bilik e-voting tercatat, 3a5fb3d Perbarui peta kode graphify, 450a912 v1.20.0: perbaiki sisa temuan audit dan risiko baru dari tinjauan ulang, ae5fa1f Perbarui peta kode graphify, cfgInt(), BERKAS, PPDB_ST

### Community 18 - "Community 18"
Cohesion: 0.18
Nodes (11): apiSibiImport(), apiSibiSearch(), cacheName(), ebookCacheMax(), nullStr(), safeFileName(), sibiFetch(), sibiJenjang() (+3 more)

### Community 19 - "Community 19"
Cohesion: 0.12
Nodes (20): clone(), comboHtml(), comboItems(), examTime(), fieldHtml(), GET(), jenjangDiatur(), kelasOpts() (+12 more)

### Community 20 - "Community 20"
Cohesion: 0.19
Nodes (5): clientIP(), isHTTPS(), limitKey(), viaProxy(), loginLimiter

### Community 21 - "Community 21"
Cohesion: 0.17
Nodes (8): apiJournal(), apiNoteDelete(), apiNotes(), apiNoteSave(), apiNoteStudents(), apiWaliClasses(), apiWaliDetail(), str2()

### Community 22 - "Community 22"
Cohesion: 0.18
Nodes (7): apiWAConfigSave(), cfgOn(), cfgStr(), waFill(), waHost(), waNumber(), waSend()

### Community 23 - "Community 23"
Cohesion: 0.18
Nodes (10): addColumn(), apiGetConfig(), apiPutConfig(), apiPutModules(), getConfig(), init(), kvForget(), kvGet() (+2 more)

### Community 24 - "Community 24"
Cohesion: 0.15
Nodes (4): kunciUjian(), TestCBTDinilaiOtomatisSetelahWaktuHabis(), TestCBTSoalTambahanTidakMenurunkanNilai(), TestKirimUlangSetelahDinilaiOtomatis()

### Community 25 - "Community 25"
Cohesion: 0.19
Nodes (5): hashToken(), apiVersion(), bad(), hashToken(), limitKey()

### Community 26 - "Community 26"
Cohesion: 0.15
Nodes (9): 6cf5813 Perbarui peta kode graphify, bcd8537 Tahap 2: inti permintaan bersama + batal saat pindah menu, helper kamera/scan/password bersama, jendela dialog aksesibel, ca7c373 Perbarui peta kode graphify, d578b62 v1.23.2: ujian CBT, kuis, dan tugas yang sudah ada nilai siswa tidak bisa dihapus guru; admin dengan konfirmasi kedua, f8135d2 v1.23.3: sudut kartu, tabel, jendela, tombol, dan isian tidak terlalu membulat, cbtTimeText(), pageCBT(), renderCbt() (+1 more)

### Community 27 - "Community 27"
Cohesion: 0.18
Nodes (2): onlyDigits(), Ctx

### Community 28 - "Community 28"
Cohesion: 0.19
Nodes (8): JAM_KE, JENIS_CAT, KATEGORI_CAT, monthInput(), pageCatatan(), pageJurnal(), pageWaliKelas(), validBulan()

### Community 29 - "Community 29"
Cohesion: 0.16
Nodes (6): 09479bb Perbarui peta kode graphify, 2c481ea v1.22.1: perangkat ajar AI per jenjang (SD/SMP/SMA/SMK): kelas & fase, saran mapel, validasi server, instruksi jenjang ke AI, 4f7bbc2 Perbarui peta kode graphify, 69d9920 Perbarui peta kode graphify, 792dd82 v1.23.0: jenjang sekolah diatur sekali di Pengaturan; tingkat kelas, saran mapel, kenaikan kelas, filter e-book, dan Asisten AI mengikuti jenjang, c3fe8fe v1.22.0: RPP Kurikulum Merdeka terbaru (RPM/Modul Ajar/RPP 1 lembar, fase & alokasi waktu otomatis, cek waktu), sunting, cetak resmi berkop & tanda tangan, unduh Word

### Community 30 - "Community 30"
Cohesion: 0.14
Nodes (2): 5720746 SCH v1.18.0: sistem informasi sekolah (Go + SQLite), GRAD

### Community 31 - "Community 31"
Cohesion: 0.23
Nodes (10): findMarkers(), homography(), LJK, ljkPos(), ljkProcess(), ljkRead(), ljkRedraw(), ljkSheet() (+2 more)

### Community 32 - "Community 32"
Cohesion: 0.29
Nodes (11): noKelas(), siswaCourse(), siswaDash(), siswaData(), siswaKuis(), siswaMapel(), siswaNilai(), siswaTugas() (+3 more)

### Community 33 - "Community 33"
Cohesion: 0.27
Nodes (8): apiGetSettings(), apiPutSettings(), apiSaveClass(), cleanDeadline(), cleanURL(), deref(), required(), str()

### Community 34 - "Community 34"
Cohesion: 0.37
Nodes (11): apiAIDocDocx(), buildDocx(), docxBody(), docxKop(), docxPara(), docxRuns(), docxTable(), docxTTD() (+3 more)

### Community 35 - "Community 35"
Cohesion: 0.26
Nodes (13): api(), apiFetch(), apiJSON(), checkVersion(), confirmBox(), DEL(), examSubmit(), GETpage() (+5 more)

### Community 36 - "Community 36"
Cohesion: 0.38
Nodes (11): courseBar(), guruAbsensi(), guruCourses(), guruKuis(), guruMapel(), guruMateri(), guruRekap(), guruTugas() (+3 more)

### Community 37 - "Community 37"
Cohesion: 0.33
Nodes (11): clientIP(), isHTTPS(), main(), newLoginLimiter(), newStaticHandler(), notFound(), printAddresses(), securityHeaders() (+3 more)

### Community 38 - "Community 38"
Cohesion: 0.18
Nodes (2): apiReturn(), perpusCfg()

### Community 39 - "Community 39"
Cohesion: 0.23
Nodes (8): apiAudit(), applyTimezone(), checkPassword(), dayRangeUTC(), init(), pageParams(), withBcrypt(), statusRecorder

### Community 40 - "Community 40"
Cohesion: 0.24
Nodes (5): apiDeleteCourse(), apiDeleteUser(), apiSaveUser(), b2i(), userIn

### Community 41 - "Community 41"
Cohesion: 0.24
Nodes (7): bad(), forbidden(), notFound(), writeErr(), forbidden(), apiErr, bisaDipaksa

### Community 42 - "Community 42"
Cohesion: 0.27
Nodes (6): aiJenjang(), rppFase(), rppMenitJP(), rppNum(), rppPrompt(), rppWaktu()

### Community 43 - "Community 43"
Cohesion: 0.24
Nodes (3): pngDataURL(), TestPPDBDitolakTidakMenyimpanBerkas(), TestUnggahanPublikDibatasiHarian()

### Community 44 - "Community 44"
Cohesion: 0.33
Nodes (8): bilikHtml(), bilikOpen(), bilikScan(), elStatus(), evoteVoter(), hasilHtml(), hasilUtama(), pageEvoting()

### Community 45 - "Community 45"
Cohesion: 0.31
Nodes (4): apiKalender(), jamLokal(), localDay(), tglJamLokal()

### Community 46 - "Community 46"
Cohesion: 0.31
Nodes (4): siswaID(), TestAbsensiMapelAlurLengkap(), TestPinjamKembaliBuku(), TestTugasBuatKumpulNilai()

### Community 47 - "Community 47"
Cohesion: 0.36
Nodes (8): cartTotal(), kasirBarang(), kasirLaporan(), pageKasir(), posAdd(), renderCart(), showStruk(), strukHtml()

### Community 48 - "Community 48"
Cohesion: 0.29
Nodes (2): apiRekapMe(), nilaiSiswa()

### Community 49 - "Community 49"
Cohesion: 0.36
Nodes (4): ensureAdmin(), hashPassword(), nowStr(), randomPassword()

### Community 50 - "Community 50"
Cohesion: 0.29
Nodes (2): apiSale(), apiSell()

### Community 51 - "Community 51"
Cohesion: 0.29
Nodes (4): nextYearName(), NK_AKSI, pageNaik(), ROMAN

### Community 52 - "Community 52"
Cohesion: 0.43
Nodes (6): nilaiAkhir(), rata(), skorDariRekap(), toF(), skorMapel, tugasNilai

### Community 53 - "Community 53"
Cohesion: 0.43
Nodes (5): apiCBTSoalBatch(), apiQuizQuestionBatch(), parseBatch(), parseMCQ(), mcq

### Community 55 - "Community 55"
Cohesion: 0.38
Nodes (4): AUDIT_LABELS, GETpage(), moreBtn(), pageAudit()

### Community 57 - "Community 57"
Cohesion: 0.67
Nodes (5): f(), fmtSkor(), sama(), TestNilaiAkhirDiRekapDanSiswaSama(), TestRumusNilaiAkhir()

### Community 58 - "Community 58"
Cohesion: 0.40
Nodes (2): apiDelete(), apiDeleteChild()

### Community 59 - "Community 59"
Cohesion: 0.40
Nodes (1): tingkatDariNama()

### Community 60 - "Community 60"
Cohesion: 0.50
Nodes (3): defaultTahunAjaran(), migTahunAjaran(), migration

### Community 61 - "Community 61"
Cohesion: 0.50
Nodes (2): manUjianDetail(), pageManUjian()

### Community 62 - "Community 62"
Cohesion: 0.50
Nodes (1): User

### Community 66 - "Community 66"
Cohesion: 0.50
Nodes (1): staticFile

### Community 67 - "Community 67"
Cohesion: 0.50
Nodes (2): KODE_SURAT, SURAT_ST

## Knowledge Gaps
- **94 isolated node(s):** `lampiran`, `userIn`, `docxOpts`, `migration`, `absenResult` (+89 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `Community 8`** (1 nodes): `App`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 27`** (2 nodes): `onlyDigits()`, `Ctx`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 30`** (2 nodes): `5720746 SCH v1.18.0: sistem informasi sekolah (Go + SQLite)`, `GRAD`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 38`** (2 nodes): `apiReturn()`, `perpusCfg()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 48`** (2 nodes): `apiRekapMe()`, `nilaiSiswa()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 50`** (2 nodes): `apiSale()`, `apiSell()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 58`** (2 nodes): `apiDelete()`, `apiDeleteChild()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 59`** (1 nodes): `tingkatDariNama()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 61`** (2 nodes): `manUjianDetail()`, `pageManUjian()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 62`** (1 nodes): `User`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 66`** (1 nodes): `staticFile`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 67`** (2 nodes): `KODE_SURAT`, `SURAT_ST`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App` connect `Community 8` to `Community 62`, `Community 15`, `Community 68`, `Community 18`, `Community 38`, `Community 14`, `Community 25`, `Community 6`, `Community 22`, `Community 20`, `Community 4`, `Community 1`, `Community 41`, `Community 34`, `Community 23`, `Community 58`, `Community 33`, `Community 37`?**
  _High betweenness centrality (0.094) - this node is a cross-community bridge._
- **Why does `Ctx` connect `Community 27` to `Community 41`, `Community 10`, `Community 45`, `Community 25`, `Community 14`, `Community 23`, `Community 6`, `Community 21`, `Community 59`, `Community 40`, `Community 58`, `Community 20`, `Community 15`, `Community 4`, `Community 1`, `Community 16`, `Community 37`?**
  _High betweenness centrality (0.059) - this node is a cross-community bridge._
- **Why does `loginLimiter` connect `Community 20` to `Community 25`, `Community 37`?**
  _High betweenness centrality (0.014) - this node is a cross-community bridge._
- **What connects `lampiran`, `userIn`, `docxOpts` to the rest of the system?**
  _94 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Community 0` be split into smaller, more focused modules?**
  _Cohesion score 0.02846846846846847 - nodes in this community are weakly interconnected._
- **Should `Community 1` be split into smaller, more focused modules?**
  _Cohesion score 0.052244897959183675 - nodes in this community are weakly interconnected._
- **Should `Community 2` be split into smaller, more focused modules?**
  _Cohesion score 0.06262626262626263 - nodes in this community are weakly interconnected._