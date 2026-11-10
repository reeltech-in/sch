# Graph Report - .  (2026-11-10)

## Corpus Check
- 72 files · ~110,258 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 930 nodes · 1687 edges · 51 communities detected
- Extraction: 100% EXTRACTED · 0% INFERRED · 0% AMBIGUOUS
- Token cost: 0 input · 0 output
- Edge kinds: contains: 735 · calls: 684 · method: 126 · MODIFIES: 123 · ON_BRANCH: 10 · PARENT_OF: 9


## Input Scope
- Requested: auto
- Resolved: committed (source: default-auto)
- Included files: 72 · Candidates: 86
- Excluded: 0 untracked · 261 ignored · 0 sensitive · 0 missing committed
- Recommendation: Use --scope all or graphify.yaml inputs.corpus for a knowledge-base folder.

## Graph Freshness
- Built from Git commit: `450a912`
- Compare this hash to `git rev-parse HEAD` before trusting freshness-sensitive graph output.
## God Nodes (most connected - your core abstractions)
1. `App` - 63 edges
2. `Ctx` - 43 edges
3. `newEnv()` - 32 edges
4. `esc()` - 26 edges
5. `expect()` - 25 edges
6. `ico()` - 13 edges
7. `loginLimiter` - 11 edges
8. `route()` - 11 edges
9. `TestNotifikasiWhatsApp()` - 9 edges
10. `main()` - 8 edges

## Surprising Connections (you probably didn't know these)
- `1bb9e1e v1.20.0: perbaikan hasil audit keamanan, ujian, query, dan tampilan` --ON_BRANCH--> `main`  [EXTRACTED]
  git → git  _Bridges community 19 → community 31_
- `1bb9e1e v1.20.0: perbaikan hasil audit keamanan, ujian, query, dan tampilan` --PARENT_OF--> `3a5fb3d Perbarui peta kode graphify`  [EXTRACTED]
  git → git  _Bridges community 19 → community 26_
- `3a5fb3d Perbarui peta kode graphify` --ON_BRANCH--> `main`  [EXTRACTED]
  git → git  _Bridges community 26 → community 31_
- `5720746 SCH v1.18.0: sistem informasi sekolah (Go + SQLite)` --ON_BRANCH--> `main`  [EXTRACTED]
  git → git  _Bridges community 12 → community 31_
- `sibiLoad()` --calls--> `sibiFetch()`  [EXTRACTED]
  mod_ebook.go → mod_ebook.go  _Bridges community 42 → community 0_

## Communities

### Community 0 - "Community 0"
Cohesion: 0.05
Nodes (17): cfgInt(), apiAbsenHarian(), apiAbsenHarianSave(), apiAbsenRekap(), haversine(), itoa(), localClock(), localDate() (+9 more)

### Community 1 - "Community 1"
Cohesion: 0.03
Nodes (31): ACT, BOOT, BUSY, CHG, collect(), COMBO, FORMS, GET_CACHE (+23 more)

### Community 2 - "Community 2"
Cohesion: 0.06
Nodes (29): hashToken(), bad(), forbidden(), notFound(), writeErr(), clientIP(), isHTTPS(), limitKey() (+21 more)

### Community 3 - "Community 3"
Cohesion: 0.06
Nodes (15): onlyDigits(), tingkatDariNama(), apiJournal(), apiNoteDelete(), apiNotes(), apiNoteSave(), apiNoteStudents(), apiWaliClasses() (+7 more)

### Community 4 - "Community 4"
Cohesion: 0.20
Nodes (37): client, env, fakeWA, expect(), newEnv(), newFakeWA(), TestAdmGuruWaliDanCatatan(), TestCBTBatasWaktuDiServer() (+29 more)

### Community 5 - "Community 5"
Cohesion: 0.05
Nodes (8): apiDeleteQuestion(), apiSaveQuestion(), apiCourseStudents(), apiDeleteQuestion(), apiRekap(), apiSaveQuestion(), apiSubmissions(), lampiran

### Community 6 - "Community 6"
Cohesion: 0.11
Nodes (13): apiElectionResult(), apiElections(), apiElectionsForMe(), apiVote(), castVote(), apiElectionResult(), apiElections(), apiElectionsForMe() (+5 more)

### Community 7 - "Community 7"
Cohesion: 0.09
Nodes (19): apiCBTAttemptDelete(), apiCBTDelete(), apiCBTHasil(), apiCBTList(), apiCBTSave(), apiCBTSaveAnswers(), apiCBTSoal(), apiCBTSoalDelete() (+11 more)

### Community 8 - "Community 8"
Cohesion: 0.09
Nodes (24): absenCeklist(), absenRekap(), absenScanHtml(), AKSI, autoTag(), canEditEv(), doScan(), evRow() (+16 more)

### Community 9 - "Community 9"
Cohesion: 0.13
Nodes (19): bookCard(), bookCover(), bookRead(), circBorrow(), circLog(), circMember(), circUpdate(), hasEbook() (+11 more)

### Community 10 - "Community 10"
Cohesion: 0.10
Nodes (15): adminDash(), adminPengaturan(), fmtBytes(), KEYMAP, ROMAWI, SET_PAGES, SET_SECTIONS, setCard() (+7 more)

### Community 11 - "Community 11"
Cohesion: 0.14
Nodes (24): avatar(), btn(), comboHtml(), empty(), ensureThemeFloat(), esc(), fieldHtml(), fmtDate() (+16 more)

### Community 12 - "Community 12"
Cohesion: 0.09
Nodes (4): 5720746 SCH v1.18.0: sistem informasi sekolah (Go + SQLite), GRAD, KODE_SURAT, SURAT_ST

### Community 13 - "Community 13"
Cohesion: 0.17
Nodes (13): aiAdminOnly(), aiKeyHint(), aiRequest(), aiUsageKey(), apiAIConfig(), apiAIConfigSave(), apiAIGenerate(), apiAIModels() (+5 more)

### Community 14 - "Community 14"
Cohesion: 0.14
Nodes (17): AI_KELAS, AI_MAPEL, AI_PRESETS, AI_TPL, aiActions(), aiAtur(), aiRiwayat(), aiRun() (+9 more)

### Community 15 - "Community 15"
Cohesion: 0.14
Nodes (3): apiPPDBInfo(), ppdbOpen(), randDigits()

### Community 16 - "Community 16"
Cohesion: 0.18
Nodes (7): apiWAConfigSave(), cfgOn(), cfgStr(), waFill(), waHost(), waNumber(), waSend()

### Community 17 - "Community 17"
Cohesion: 0.18
Nodes (10): addColumn(), apiGetConfig(), apiPutConfig(), apiPutModules(), getConfig(), init(), kvForget(), kvGet() (+2 more)

### Community 18 - "Community 18"
Cohesion: 0.19
Nodes (9): apiDelete(), apiDeleteChild(), apiGetSettings(), apiPutSettings(), apiSaveClass(), cleanDeadline(), cleanURL(), required() (+1 more)

### Community 19 - "Community 19"
Cohesion: 0.13
Nodes (1): 1bb9e1e v1.20.0: perbaikan hasil audit keamanan, ujian, query, dan tampilan

### Community 20 - "Community 20"
Cohesion: 0.19
Nodes (8): JAM_KE, JENIS_CAT, KATEGORI_CAT, monthInput(), pageCatatan(), pageJurnal(), pageWaliKelas(), validBulan()

### Community 21 - "Community 21"
Cohesion: 0.23
Nodes (10): findMarkers(), homography(), LJK, ljkPos(), ljkProcess(), ljkRead(), ljkRedraw(), ljkSheet() (+2 more)

### Community 22 - "Community 22"
Cohesion: 0.29
Nodes (11): noKelas(), siswaCourse(), siswaDash(), siswaData(), siswaKuis(), siswaMapel(), siswaNilai(), siswaTugas() (+3 more)

### Community 23 - "Community 23"
Cohesion: 0.38
Nodes (11): courseBar(), guruAbsensi(), guruCourses(), guruKuis(), guruMapel(), guruMateri(), guruRekap(), guruTugas() (+3 more)

### Community 24 - "Community 24"
Cohesion: 0.23
Nodes (8): apiAudit(), applyTimezone(), checkPassword(), dayRangeUTC(), init(), pageParams(), withBcrypt(), statusRecorder

### Community 25 - "Community 25"
Cohesion: 0.20
Nodes (12): closeModal(), GETpage(), go(), loadScript(), modOn(), navFor(), parseHash(), qs() (+4 more)

### Community 26 - "Community 26"
Cohesion: 0.18
Nodes (2): 3a5fb3d Perbarui peta kode graphify, 450a912 v1.20.0: perbaiki sisa temuan audit dan risiko baru dari tinjauan ulang

### Community 27 - "Community 27"
Cohesion: 0.20
Nodes (2): apiReturn(), perpusCfg()

### Community 28 - "Community 28"
Cohesion: 0.22
Nodes (4): activeYear(), apiPromote(), nextTingkat(), nullIfEmpty()

### Community 29 - "Community 29"
Cohesion: 0.24
Nodes (11): api(), applyTheme(), checkVersion(), confirmBox(), cycleTheme(), DEL(), mutate(), POST() (+3 more)

### Community 30 - "Community 30"
Cohesion: 0.24
Nodes (5): apiDeleteCourse(), apiDeleteUser(), apiSaveUser(), b2i(), userIn

### Community 31 - "Community 31"
Cohesion: 0.36
Nodes (7): main, 334b045 v1.19.2: dropdown pilih-atau-ketik ramah HP, panah select seragam, Asisten AI satu kolom di HP, a82ab3f Skrip install.sh dan panduan instalasi Ubuntu/Debian, b479985 Pengaturan AI dipindah ke Pengaturan → Asisten AI, b86971e v1.19.1: mapel yang diampu guru diutamakan di Asisten AI dan CBT, c9b9c9d v1.19.0: impor e-book dari Buku Kemendikbud, PDF disimpan saat pertama dibuka, f48bfaa Rilis v1.18.1

### Community 32 - "Community 32"
Cohesion: 0.31
Nodes (9): brandHtml(), formHtml(), head(), openForm(), pageAkun(), publicShell(), registerForm(), renderGantiPW() (+1 more)

### Community 33 - "Community 33"
Cohesion: 0.31
Nodes (9): clone(), comboItems(), GET(), kelasOpts(), mapelOpts(), saran(), setComboOptions(), tahunAjaranOpts() (+1 more)

### Community 34 - "Community 34"
Cohesion: 0.28
Nodes (4): cbtTimeText(), pageCBT(), renderCbt(), siswaCBT()

### Community 35 - "Community 35"
Cohesion: 0.36
Nodes (8): cartTotal(), kasirBarang(), kasirLaporan(), pageKasir(), posAdd(), renderCart(), showStruk(), strukHtml()

### Community 36 - "Community 36"
Cohesion: 0.36
Nodes (4): kunciUjian(), TestCBTDinilaiOtomatisSetelahWaktuHabis(), TestCBTSoalTambahanTidakMenurunkanNilai(), TestKirimUlangSetelahDinilaiOtomatis()

### Community 37 - "Community 37"
Cohesion: 0.36
Nodes (4): ensureAdmin(), hashPassword(), nowStr(), randomPassword()

### Community 38 - "Community 38"
Cohesion: 0.29
Nodes (2): apiSale(), apiSell()

### Community 39 - "Community 39"
Cohesion: 0.39
Nodes (6): bilikOpen(), bilikScan(), elStatus(), evoteVoter(), hasilHtml(), pageEvoting()

### Community 40 - "Community 40"
Cohesion: 0.29
Nodes (4): nextYearName(), NK_AKSI, pageNaik(), ROMAN

### Community 42 - "Community 42"
Cohesion: 0.43
Nodes (5): apiSibiImport(), apiSibiSearch(), nullStr(), sibiLoad(), sibiBook

### Community 43 - "Community 43"
Cohesion: 0.33
Nodes (2): apiPPDBInfo(), ppdbOpen()

### Community 44 - "Community 44"
Cohesion: 0.43
Nodes (5): apiCBTSoalBatch(), apiQuizQuestionBatch(), parseBatch(), parseMCQ(), mcq

### Community 45 - "Community 45"
Cohesion: 0.38
Nodes (4): AUDIT_LABELS, GETpage(), moreBtn(), pageAudit()

### Community 47 - "Community 47"
Cohesion: 0.40
Nodes (6): afterSave(), errBox(), isAdm(), pageFn(), refresh(), refreshModal()

### Community 48 - "Community 48"
Cohesion: 0.50
Nodes (3): defaultTahunAjaran(), migTahunAjaran(), migration

### Community 49 - "Community 49"
Cohesion: 0.40
Nodes (5): lampHtml(), logoSrc(), safeUrl(), schoolHead(), ytId()

### Community 50 - "Community 50"
Cohesion: 0.50
Nodes (2): manUjianDetail(), pageManUjian()

### Community 51 - "Community 51"
Cohesion: 0.40
Nodes (2): BERKAS, PPDB_ST

### Community 52 - "Community 52"
Cohesion: 0.50
Nodes (4): monthNow(), pad(), today(), toLocalInput()

## Knowledge Gaps
- **78 isolated node(s):** `lampiran`, `userIn`, `migration`, `absenResult`, `aiTemplate` (+73 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `Community 19`** (1 nodes): `1bb9e1e v1.20.0: perbaikan hasil audit keamanan, ujian, query, dan tampilan`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 26`** (2 nodes): `3a5fb3d Perbarui peta kode graphify`, `450a912 v1.20.0: perbaiki sisa temuan audit dan risiko baru dari tinjauan ulang`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 27`** (2 nodes): `apiReturn()`, `perpusCfg()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 38`** (2 nodes): `apiSale()`, `apiSell()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 43`** (2 nodes): `apiPPDBInfo()`, `ppdbOpen()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 50`** (2 nodes): `manUjianDetail()`, `pageManUjian()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 51`** (2 nodes): `BERKAS`, `PPDB_ST`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `App` connect `Community 0` to `Community 19`, `Community 27`, `Community 15`, `Community 2`, `Community 6`, `Community 16`, `Community 7`, `Community 17`, `Community 18`?**
  _High betweenness centrality (0.106) - this node is a cross-community bridge._
- **Why does `Ctx` connect `Community 3` to `Community 2`, `Community 13`, `Community 15`, `Community 17`, `Community 6`, `Community 30`, `Community 18`, `Community 0`, `Community 7`, `Community 5`, `Community 28`?**
  _High betweenness centrality (0.070) - this node is a cross-community bridge._
- **What connects `lampiran`, `userIn`, `migration` to the rest of the system?**
  _78 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Community 0` be split into smaller, more focused modules?**
  _Cohesion score 0.05336538461538461 - nodes in this community are weakly interconnected._
- **Should `Community 1` be split into smaller, more focused modules?**
  _Cohesion score 0.03277009728622632 - nodes in this community are weakly interconnected._
- **Should `Community 2` be split into smaller, more focused modules?**
  _Cohesion score 0.05817028027498678 - nodes in this community are weakly interconnected._
- **Should `Community 3` be split into smaller, more focused modules?**
  _Cohesion score 0.06382978723404255 - nodes in this community are weakly interconnected._