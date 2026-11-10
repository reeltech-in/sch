# Graph Report - .  (2026-11-10)

## Corpus Check
- 75 files · ~123,501 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 1072 nodes · 2027 edges · 50 communities detected
- Extraction: 100% EXTRACTED · 0% INFERRED · 0% AMBIGUOUS
- Token cost: 0 input · 0 output
- Edge kinds: contains: 836 · calls: 785 · MODIFIES: 197 · method: 138 · ON_BRANCH: 36 · PARENT_OF: 35


## Input Scope
- Requested: auto
- Resolved: committed (source: default-auto)
- Included files: 75 · Candidates: 89
- Excluded: 0 untracked · 357 ignored · 0 sensitive · 0 missing committed
- Recommendation: Use --scope all or graphify.yaml inputs.corpus for a knowledge-base folder.

## Graph Freshness
- Built from Git commit: `d06aeac`
- Compare this hash to `git rev-parse HEAD` before trusting freshness-sensitive graph output.
## God Nodes (most connected - your core abstractions)
1. `App` - 65 edges
2. `Ctx` - 46 edges
3. `newEnv()` - 32 edges
4. `esc()` - 29 edges
5. `expect()` - 25 edges
6. `ico()` - 15 edges
7. `loginLimiter` - 11 edges
8. `route()` - 11 edges
9. `api()` - 10 edges
10. `apiAIGenerate()` - 9 edges

## Surprising Connections (you probably didn't know these)
- `1119b6b Tahap 3: mesin ujian kuis & CBT memakai hitung mundur, kirim otomatis, dan kartu hasil bersama` --ON_BRANCH--> `main`  [EXTRACTED]
  git → git  _Bridges community 0 → community 15_
- `1bb9e1e v1.20.0: perbaikan hasil audit keamanan, ujian, query, dan tampilan` --ON_BRANCH--> `main`  [EXTRACTED]
  git → git  _Bridges community 3 → community 15_
- `1bb9e1e v1.20.0: perbaikan hasil audit keamanan, ujian, query, dan tampilan` --PARENT_OF--> `3a5fb3d Perbarui peta kode graphify`  [EXTRACTED]
  git → git  _Bridges community 3 → community 24_
- `3a5fb3d Perbarui peta kode graphify` --ON_BRANCH--> `main`  [EXTRACTED]
  git → git  _Bridges community 24 → community 15_
- `5720746 SCH v1.18.0: sistem informasi sekolah (Go + SQLite)` --ON_BRANCH--> `main`  [EXTRACTED]
  git → git  _Bridges community 13 → community 15_

## Communities

### Community 0 - "Community 0"
Cohesion: 0.03
Nodes (43): 1119b6b Tahap 3: mesin ujian kuis & CBT memakai hitung mundur, kirim otomatis, dan kartu hasil bersama, 137792c Tahap 4: hapus 57 deklarasi CSS yang selalu ditimpa (tampilan terverifikasi identik di desktop & HP); peringatan keluar hanya saat ujian belum terkirim, 1e19524 Tahap 1: peserta LJK ikut riwayat kelas, validasi jam absensi, dasbor guru per tahun aktif, cache & halaman publik lebih aman, muat paralel, bcd8537 Tahap 2: inti permintaan bersama + batal saat pindah menu, helper kamera/scan/password bersama, jendela dialog aksesibel, acctItems(), acctMenu(), ACT, BOOT (+35 more)

### Community 1 - "Community 1"
Cohesion: 0.05
Nodes (23): apiAbsenHarian(), apiAbsenHarianSave(), apiAbsenRekap(), haversine(), init(), itoa(), jamValid(), localClock() (+15 more)

### Community 2 - "Community 2"
Cohesion: 0.06
Nodes (29): hashToken(), bad(), forbidden(), notFound(), writeErr(), clientIP(), isHTTPS(), limitKey() (+21 more)

### Community 3 - "Community 3"
Cohesion: 0.05
Nodes (38): 1bb9e1e v1.20.0: perbaikan hasil audit keamanan, ujian, query, dan tampilan, 334b045 v1.19.2: dropdown pilih-atau-ketik ramah HP, panah select seragam, Asisten AI satu kolom di HP, b86971e v1.19.1: mapel yang diampu guru diutamakan di Asisten AI dan CBT, AGAMA, AI_JENJANG, AI_KELAS, AI_MAPEL, AI_PRESETS (+30 more)

### Community 4 - "Community 4"
Cohesion: 0.05
Nodes (13): apiDeleteQuestion(), apiQuizSave(), apiQuizSubmit(), apiSaveQuestion(), quizClean(), quizDeadline(), scoreQuiz(), apiCourseStudents() (+5 more)

### Community 5 - "Community 5"
Cohesion: 0.20
Nodes (37): client, env, fakeWA, expect(), newEnv(), newFakeWA(), TestAdmGuruWaliDanCatatan(), TestCBTBatasWaktuDiServer() (+29 more)

### Community 6 - "Community 6"
Cohesion: 0.09
Nodes (23): apiCBTAttemptDelete(), apiCBTDelete(), apiCBTHasil(), apiCBTList(), apiCBTSave(), apiCBTSaveAnswers(), apiCBTSoal(), apiCBTSoalDelete() (+15 more)

### Community 7 - "Community 7"
Cohesion: 0.09
Nodes (36): applyTheme(), avatar(), brandHtml(), btn(), cycleTheme(), empty(), ensureThemeFloat(), esc() (+28 more)

### Community 8 - "Community 8"
Cohesion: 0.11
Nodes (13): apiElectionResult(), apiElections(), apiElectionsForMe(), apiVote(), castVote(), apiElectionResult(), apiElections(), apiElectionsForMe() (+5 more)

### Community 9 - "Community 9"
Cohesion: 0.09
Nodes (24): absenCeklist(), absenRekap(), absenScanHtml(), AKSI, autoTag(), canEditEv(), doScan(), evRow() (+16 more)

### Community 10 - "Community 10"
Cohesion: 0.13
Nodes (16): aiAdminOnly(), aiCleanInputs(), aiKeyHint(), aiProviderError(), aiReadStream(), aiRequest(), aiUsageKey(), apiAIConfig() (+8 more)

### Community 11 - "Community 11"
Cohesion: 0.10
Nodes (15): adminDash(), adminPengaturan(), fmtBytes(), KEYMAP, ROMAWI, SET_PAGES, SET_SECTIONS, setCard() (+7 more)

### Community 12 - "Community 12"
Cohesion: 0.13
Nodes (19): bookCard(), bookCover(), bookRead(), circBorrow(), circLog(), circMember(), circUpdate(), hasEbook() (+11 more)

### Community 13 - "Community 13"
Cohesion: 0.08
Nodes (4): 5720746 SCH v1.18.0: sistem informasi sekolah (Go + SQLite), GRAD, KODE_SURAT, SURAT_ST

### Community 14 - "Community 14"
Cohesion: 0.10
Nodes (5): apiPPDBInfo(), ppdbOpen(), randDigits(), apiPPDBInfo(), ppdbOpen()

### Community 15 - "Community 15"
Cohesion: 0.18
Nodes (22): main, 01cf6c7 Perbarui peta kode graphify, 044f2aa v1.23.1: kerapian menu guru & siswa di HP: tombol absensi lebih besar + simpan menempel, kolom nama tetap terlihat di tabel rekap, kehadiran ditulis lengkap, 'Belum ada materi', 09479bb Perbarui peta kode graphify, 2c481ea v1.22.1: perangkat ajar AI per jenjang (SD/SMP/SMA/SMK): kelas & fase, saran mapel, validasi server, instruksi jenjang ke AI, 3cd8909 Perbarui peta kode graphify, 4f7bbc2 Perbarui peta kode graphify, 69d9920 Perbarui peta kode graphify (+14 more)

### Community 16 - "Community 16"
Cohesion: 0.15
Nodes (8): activeYear(), apiPromote(), nextTingkat(), nullIfEmpty(), unlistedClasses(), promoteKelas, promoter, promoteReq

### Community 17 - "Community 17"
Cohesion: 0.14
Nodes (3): onlyDigits(), tingkatDariNama(), Ctx

### Community 18 - "Community 18"
Cohesion: 0.12
Nodes (20): clone(), comboHtml(), comboItems(), examTime(), fieldHtml(), GET(), jenjangDiatur(), kelasOpts() (+12 more)

### Community 19 - "Community 19"
Cohesion: 0.17
Nodes (8): apiJournal(), apiNoteDelete(), apiNotes(), apiNoteSave(), apiNoteStudents(), apiWaliClasses(), apiWaliDetail(), str2()

### Community 20 - "Community 20"
Cohesion: 0.18
Nodes (7): apiWAConfigSave(), cfgOn(), cfgStr(), waFill(), waHost(), waNumber(), waSend()

### Community 21 - "Community 21"
Cohesion: 0.18
Nodes (10): addColumn(), apiGetConfig(), apiPutConfig(), apiPutModules(), getConfig(), init(), kvForget(), kvGet() (+2 more)

### Community 22 - "Community 22"
Cohesion: 0.18
Nodes (17): closeModal(), formHtml(), go(), loadScript(), modOn(), navFor(), openForm(), pageAkun() (+9 more)

### Community 23 - "Community 23"
Cohesion: 0.15
Nodes (4): kunciUjian(), TestCBTDinilaiOtomatisSetelahWaktuHabis(), TestCBTSoalTambahanTidakMenurunkanNilai(), TestKirimUlangSetelahDinilaiOtomatis()

### Community 24 - "Community 24"
Cohesion: 0.13
Nodes (5): 3a5fb3d Perbarui peta kode graphify, 450a912 v1.20.0: perbaiki sisa temuan audit dan risiko baru dari tinjauan ulang, ae5fa1f Perbarui peta kode graphify, BERKAS, PPDB_ST

### Community 25 - "Community 25"
Cohesion: 0.19
Nodes (8): JAM_KE, JENIS_CAT, KATEGORI_CAT, monthInput(), pageCatatan(), pageJurnal(), pageWaliKelas(), validBulan()

### Community 26 - "Community 26"
Cohesion: 0.23
Nodes (10): findMarkers(), homography(), LJK, ljkPos(), ljkProcess(), ljkRead(), ljkRedraw(), ljkSheet() (+2 more)

### Community 27 - "Community 27"
Cohesion: 0.29
Nodes (11): noKelas(), siswaCourse(), siswaDash(), siswaData(), siswaKuis(), siswaMapel(), siswaNilai(), siswaTugas() (+3 more)

### Community 28 - "Community 28"
Cohesion: 0.27
Nodes (8): apiGetSettings(), apiPutSettings(), apiSaveClass(), cleanDeadline(), cleanURL(), deref(), required(), str()

### Community 29 - "Community 29"
Cohesion: 0.37
Nodes (11): apiAIDocDocx(), buildDocx(), docxBody(), docxKop(), docxPara(), docxRuns(), docxTable(), docxTTD() (+3 more)

### Community 30 - "Community 30"
Cohesion: 0.26
Nodes (13): api(), apiFetch(), apiJSON(), checkVersion(), confirmBox(), DEL(), examSubmit(), GETpage() (+5 more)

### Community 31 - "Community 31"
Cohesion: 0.38
Nodes (11): courseBar(), guruAbsensi(), guruCourses(), guruKuis(), guruMapel(), guruMateri(), guruRekap(), guruTugas() (+3 more)

### Community 32 - "Community 32"
Cohesion: 0.23
Nodes (8): apiAudit(), applyTimezone(), checkPassword(), dayRangeUTC(), init(), pageParams(), withBcrypt(), statusRecorder

### Community 33 - "Community 33"
Cohesion: 0.20
Nodes (2): apiReturn(), perpusCfg()

### Community 34 - "Community 34"
Cohesion: 0.27
Nodes (6): aiJenjang(), rppFase(), rppMenitJP(), rppNum(), rppPrompt(), rppWaktu()

### Community 35 - "Community 35"
Cohesion: 0.24
Nodes (5): apiDeleteCourse(), apiDeleteUser(), apiSaveUser(), b2i(), userIn

### Community 36 - "Community 36"
Cohesion: 0.22
Nodes (5): a82ab3f Skrip install.sh dan panduan instalasi Ubuntu/Debian, b479985 Pengaturan AI dipindah ke Pengaturan → Asisten AI, c9b9c9d v1.19.0: impor e-book dari Buku Kemendikbud, PDF disimpan saat pertama dibuka, f48bfaa Rilis v1.18.1, cfgInt()

### Community 37 - "Community 37"
Cohesion: 0.31
Nodes (4): apiKalender(), jamLokal(), localDay(), tglJamLokal()

### Community 38 - "Community 38"
Cohesion: 0.28
Nodes (4): cbtTimeText(), pageCBT(), renderCbt(), siswaCBT()

### Community 39 - "Community 39"
Cohesion: 0.36
Nodes (8): cartTotal(), kasirBarang(), kasirLaporan(), pageKasir(), posAdd(), renderCart(), showStruk(), strukHtml()

### Community 40 - "Community 40"
Cohesion: 0.36
Nodes (4): ensureAdmin(), hashPassword(), nowStr(), randomPassword()

### Community 41 - "Community 41"
Cohesion: 0.29
Nodes (2): apiSale(), apiSell()

### Community 43 - "Community 43"
Cohesion: 0.39
Nodes (6): bilikOpen(), bilikScan(), elStatus(), evoteVoter(), hasilHtml(), pageEvoting()

### Community 44 - "Community 44"
Cohesion: 0.29
Nodes (4): nextYearName(), NK_AKSI, pageNaik(), ROMAN

### Community 46 - "Community 46"
Cohesion: 0.43
Nodes (5): apiCBTSoalBatch(), apiQuizQuestionBatch(), parseBatch(), parseMCQ(), mcq

### Community 47 - "Community 47"
Cohesion: 0.33
Nodes (7): afterSave(), errBox(), head(), isAdm(), pageFn(), refresh(), refreshModal()

### Community 48 - "Community 48"
Cohesion: 0.38
Nodes (4): AUDIT_LABELS, GETpage(), moreBtn(), pageAudit()

### Community 50 - "Community 50"
Cohesion: 0.40
Nodes (2): apiDelete(), apiDeleteChild()

### Community 51 - "Community 51"
Cohesion: 0.50
Nodes (3): defaultTahunAjaran(), migTahunAjaran(), migration

### Community 52 - "Community 52"
Cohesion: 0.50
Nodes (2): manUjianDetail(), pageManUjian()

## Knowledge Gaps
- **92 isolated node(s):** `lampiran`, `userIn`, `docxOpts`, `migration`, `absenResult` (+87 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `Community 33`** (2 nodes): `apiReturn()`, `perpusCfg()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 41`** (2 nodes): `apiSale()`, `apiSell()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 50`** (2 nodes): `apiDelete()`, `apiDeleteChild()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 52`** (2 nodes): `manUjianDetail()`, `pageManUjian()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App` connect `Community 1` to `Community 2`, `Community 33`, `Community 14`, `Community 8`, `Community 20`, `Community 6`, `Community 4`, `Community 29`, `Community 21`, `Community 50`, `Community 28`?**
  _High betweenness centrality (0.097) - this node is a cross-community bridge._
- **Why does `Ctx` connect `Community 17` to `Community 2`, `Community 10`, `Community 37`, `Community 14`, `Community 21`, `Community 8`, `Community 19`, `Community 35`, `Community 50`, `Community 1`, `Community 6`, `Community 4`, `Community 16`?**
  _High betweenness centrality (0.062) - this node is a cross-community bridge._
- **What connects `lampiran`, `userIn`, `docxOpts` to the rest of the system?**
  _92 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Community 0` be split into smaller, more focused modules?**
  _Cohesion score 0.027913015254787406 - nodes in this community are weakly interconnected._
- **Should `Community 1` be split into smaller, more focused modules?**
  _Cohesion score 0.050078247261345854 - nodes in this community are weakly interconnected._
- **Should `Community 2` be split into smaller, more focused modules?**
  _Cohesion score 0.05683563748079877 - nodes in this community are weakly interconnected._
- **Should `Community 3` be split into smaller, more focused modules?**
  _Cohesion score 0.05079825834542816 - nodes in this community are weakly interconnected._