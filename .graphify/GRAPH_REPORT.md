# Graph Report - .  (2026-11-09)

## Corpus Check
- 69 files · ~114,005 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 883 nodes · 1478 edges · 50 communities detected
- Extraction: 100% EXTRACTED · 0% INFERRED · 0% AMBIGUOUS
- Token cost: 0 input · 0 output
- Edge kinds: contains: 716 · calls: 627 · method: 111 · references: 15 · reads_from: 8 · triggers: 1


## Input Scope
- Requested: auto
- Resolved: all (source: default-auto)
- Included files: 69 · Candidates: recursive
- Excluded: 0 untracked · 0 ignored · 0 sensitive · 0 missing committed
## God Nodes (most connected - your core abstractions)
1. `App` - 55 edges
2. `Ctx` - 39 edges
3. `newEnv()` - 32 edges
4. `expect()` - 25 edges
5. `esc()` - 25 edges
6. `ico()` - 12 edges
7. `public.users` - 11 edges
8. `route()` - 11 edges
9. `TestNotifikasiWhatsApp()` - 9 edges
10. `loginLimiter` - 8 edges

## Surprising Connections (you probably didn't know these)
- `api()` --calls--> `resetClient()`  [EXTRACTED]
  web/app.js → web/app.js  _Bridges community 28 → community 18_
- `brandHtml()` --calls--> `esc()`  [EXTRACTED]
  web/app.js → web/app.js  _Bridges community 32 → community 12_
- `brandHtml()` --calls--> `logoSrc()`  [EXTRACTED]
  web/app.js → web/app.js  _Bridges community 32 → community 46_
- `checkVersion()` --calls--> `esc()`  [EXTRACTED]
  web/app.js → web/app.js  _Bridges community 28 → community 12_
- `errBox()` --calls--> `esc()`  [EXTRACTED]
  web/app.js → web/app.js  _Bridges community 18 → community 12_

## Communities

### Community 0 - "Community 0"
Cohesion: 0.04
Nodes (29): ACT, BOOT, CHG, collect(), FORMS, GET_CACHE, GET_INFLIGHT, GROUP_ORDER (+21 more)

### Community 1 - "Community 1"
Cohesion: 0.06
Nodes (28): hashToken(), bad(), forbidden(), notFound(), writeErr(), clientIP(), isHTTPS(), viaProxy() (+20 more)

### Community 2 - "Community 2"
Cohesion: 0.20
Nodes (37): client, env, fakeWA, expect(), newEnv(), newFakeWA(), TestAdmGuruWaliDanCatatan(), TestCBTBatasWaktuDiServer() (+29 more)

### Community 3 - "Community 3"
Cohesion: 0.05
Nodes (8): apiDeleteQuestion(), apiSaveQuestion(), apiCourseStudents(), apiDeleteQuestion(), apiRekap(), apiSaveQuestion(), apiSubmissions(), lampiran

### Community 4 - "Community 4"
Cohesion: 0.09
Nodes (11): onlyDigits(), tingkatDariNama(), apiJournal(), apiNoteDelete(), apiNotes(), apiNoteSave(), apiNoteStudents(), apiWaliClasses() (+3 more)

### Community 5 - "Community 5"
Cohesion: 0.11
Nodes (13): apiElectionResult(), apiElections(), apiElectionsForMe(), apiVote(), castVote(), apiElectionResult(), apiElections(), apiElectionsForMe() (+5 more)

### Community 6 - "Community 6"
Cohesion: 0.09
Nodes (24): absenCeklist(), absenRekap(), absenScanHtml(), AKSI, autoTag(), canEditEv(), doScan(), evRow() (+16 more)

### Community 7 - "Community 7"
Cohesion: 0.10
Nodes (17): apiCBTAttemptDelete(), apiCBTDelete(), apiCBTHasil(), apiCBTList(), apiCBTSave(), apiCBTSaveAnswers(), apiCBTSoal(), apiCBTSoalDelete() (+9 more)

### Community 8 - "Community 8"
Cohesion: 0.10
Nodes (15): adminDash(), adminPengaturan(), fmtBytes(), KEYMAP, ROMAWI, SET_PAGES, SET_SECTIONS, setCard() (+7 more)

### Community 9 - "Community 9"
Cohesion: 0.10
Nodes (5): apiPPDBInfo(), ppdbOpen(), randDigits(), apiPPDBInfo(), ppdbOpen()

### Community 10 - "Community 10"
Cohesion: 0.14
Nodes (1): App

### Community 11 - "Community 11"
Cohesion: 0.18
Nodes (20): jsonb_to_recordset, old.password, public.assignments, public.attendance, public.change_password(), public.classes, public.courses, public.hash_user_password() (+12 more)

### Community 12 - "Community 12"
Cohesion: 0.15
Nodes (23): avatar(), btn(), empty(), ensureThemeFloat(), esc(), fieldHtml(), fmtDate(), ico() (+15 more)

### Community 13 - "Community 13"
Cohesion: 0.18
Nodes (12): aiAdminOnly(), aiKeyHint(), aiRequest(), apiAIConfig(), apiAIConfigSave(), apiAIGenerate(), apiAIModels(), apiAIStatus() (+4 more)

### Community 14 - "Community 14"
Cohesion: 0.16
Nodes (8): apiAbsenHarian(), apiAbsenHarianSave(), apiAbsenRekap(), haversine(), itoa(), localDate(), monthRange(), absenResult

### Community 15 - "Community 15"
Cohesion: 0.14
Nodes (17): AI_KELAS, AI_MAPEL, AI_PRESETS, AI_TPL, aiActions(), aiAtur(), aiRiwayat(), aiRun() (+9 more)

### Community 16 - "Community 16"
Cohesion: 0.18
Nodes (7): apiWAConfigSave(), cfgOn(), cfgStr(), waFill(), waHost(), waNumber(), waSend()

### Community 17 - "Community 17"
Cohesion: 0.18
Nodes (10): addColumn(), apiGetConfig(), apiPutConfig(), apiPutModules(), getConfig(), init(), kvForget(), kvGet() (+2 more)

### Community 18 - "Community 18"
Cohesion: 0.15
Nodes (17): afterSave(), closeModal(), errBox(), GETpage(), go(), isAdm(), loadScript(), modOn() (+9 more)

### Community 19 - "Community 19"
Cohesion: 0.19
Nodes (13): circBorrow(), circLog(), circMember(), circUpdate(), KIB, KONDISI, memberBox(), pagePerpus() (+5 more)

### Community 20 - "Community 20"
Cohesion: 0.19
Nodes (9): apiDelete(), apiDeleteChild(), apiGetSettings(), apiPutSettings(), apiSaveClass(), cleanDeadline(), cleanURL(), required() (+1 more)

### Community 21 - "Community 21"
Cohesion: 0.19
Nodes (8): JAM_KE, JENIS_CAT, KATEGORI_CAT, monthInput(), pageCatatan(), pageJurnal(), pageWaliKelas(), validBulan()

### Community 22 - "Community 22"
Cohesion: 0.23
Nodes (10): findMarkers(), homography(), LJK, ljkPos(), ljkProcess(), ljkRead(), ljkRedraw(), ljkSheet() (+2 more)

### Community 23 - "Community 23"
Cohesion: 0.29
Nodes (11): noKelas(), siswaCourse(), siswaDash(), siswaData(), siswaKuis(), siswaMapel(), siswaNilai(), siswaTugas() (+3 more)

### Community 24 - "Community 24"
Cohesion: 0.38
Nodes (11): courseBar(), guruAbsensi(), guruCourses(), guruKuis(), guruMapel(), guruMateri(), guruRekap(), guruTugas() (+3 more)

### Community 25 - "Community 25"
Cohesion: 0.20
Nodes (2): apiReturn(), perpusCfg()

### Community 26 - "Community 26"
Cohesion: 0.22
Nodes (4): activeYear(), apiPromote(), nextTingkat(), nullIfEmpty()

### Community 27 - "Community 27"
Cohesion: 0.24
Nodes (7): apiAudit(), applyTimezone(), checkPassword(), init(), pageParams(), withBcrypt(), statusRecorder

### Community 28 - "Community 28"
Cohesion: 0.24
Nodes (11): api(), applyTheme(), checkVersion(), confirmBox(), cycleTheme(), DEL(), mutate(), POST() (+3 more)

### Community 29 - "Community 29"
Cohesion: 0.24
Nodes (5): apiDeleteCourse(), apiDeleteUser(), apiSaveUser(), b2i(), userIn

### Community 30 - "Community 30"
Cohesion: 0.25
Nodes (1): cfgInt()

### Community 31 - "Community 31"
Cohesion: 0.31
Nodes (4): apiKalender(), jamLokal(), localDay(), tglJamLokal()

### Community 32 - "Community 32"
Cohesion: 0.31
Nodes (9): brandHtml(), formHtml(), head(), openForm(), pageAkun(), publicShell(), registerForm(), renderGantiPW() (+1 more)

### Community 33 - "Community 33"
Cohesion: 0.28
Nodes (4): cbtTimeText(), pageCBT(), renderCbt(), siswaCBT()

### Community 34 - "Community 34"
Cohesion: 0.36
Nodes (8): cartTotal(), kasirBarang(), kasirLaporan(), pageKasir(), posAdd(), renderCart(), showStruk(), strukHtml()

### Community 35 - "Community 35"
Cohesion: 0.36
Nodes (4): ensureAdmin(), hashPassword(), nowStr(), randomPassword()

### Community 36 - "Community 36"
Cohesion: 0.29
Nodes (2): apiSale(), apiSell()

### Community 37 - "Community 37"
Cohesion: 0.39
Nodes (6): bilikOpen(), bilikScan(), elStatus(), evoteVoter(), hasilHtml(), pageEvoting()

### Community 38 - "Community 38"
Cohesion: 0.29
Nodes (4): nextYearName(), NK_AKSI, pageNaik(), ROMAN

### Community 40 - "Community 40"
Cohesion: 0.43
Nodes (5): apiCBTSoalBatch(), apiQuizQuestionBatch(), parseBatch(), parseMCQ(), mcq

### Community 41 - "Community 41"
Cohesion: 0.43
Nodes (7): clone(), GET(), kelasOpts(), mapelOpts(), saran(), tahunAjaranOpts(), uniqList()

### Community 42 - "Community 42"
Cohesion: 0.38
Nodes (4): AUDIT_LABELS, GETpage(), moreBtn(), pageAudit()

### Community 44 - "Community 44"
Cohesion: 0.50
Nodes (3): defaultTahunAjaran(), migTahunAjaran(), migration

### Community 46 - "Community 46"
Cohesion: 0.40
Nodes (5): lampHtml(), logoSrc(), safeUrl(), schoolHead(), ytId()

### Community 47 - "Community 47"
Cohesion: 0.50
Nodes (2): manUjianDetail(), pageManUjian()

### Community 48 - "Community 48"
Cohesion: 0.40
Nodes (2): BERKAS, PPDB_ST

### Community 49 - "Community 49"
Cohesion: 0.67
Nodes (1): localClock()

### Community 53 - "Community 53"
Cohesion: 0.50
Nodes (4): monthNow(), pad(), today(), toLocalInput()

### Community 54 - "Community 54"
Cohesion: 0.50
Nodes (2): KODE_SURAT, SURAT_ST

### Community 57 - "Community 57"
Cohesion: 0.67
Nodes (1): GRAD

## Knowledge Gaps
- **75 isolated node(s):** `lampiran`, `userIn`, `migration`, `absenResult`, `aiTemplate` (+70 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **Thin community `Community 10`** (1 nodes): `App`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 25`** (2 nodes): `apiReturn()`, `perpusCfg()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 30`** (1 nodes): `cfgInt()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 36`** (2 nodes): `apiSale()`, `apiSell()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 47`** (2 nodes): `manUjianDetail()`, `pageManUjian()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 48`** (2 nodes): `BERKAS`, `PPDB_ST`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 49`** (1 nodes): `localClock()`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 54`** (2 nodes): `KODE_SURAT`, `SURAT_ST`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.
- **Thin community `Community 57`** (1 nodes): `GRAD`
  Too small to be a meaningful cluster - may be noise or needs more connections extracted.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `Ctx` connect `Community 4` to `Community 1`, `Community 13`, `Community 31`, `Community 9`, `Community 17`, `Community 5`, `Community 29`, `Community 20`, `Community 14`, `Community 7`, `Community 3`, `Community 26`?**
  _High betweenness centrality (0.134) - this node is a cross-community bridge._
- **Why does `App` connect `Community 10` to `Community 1`, `Community 14`, `Community 49`, `Community 25`, `Community 9`, `Community 30`, `Community 5`, `Community 16`, `Community 17`, `Community 20`?**
  _High betweenness centrality (0.080) - this node is a cross-community bridge._
- **Why does `apiDeleteChild()` connect `Community 20` to `Community 4`?**
  _High betweenness centrality (0.014) - this node is a cross-community bridge._
- **What connects `lampiran`, `userIn`, `migration` to the rest of the system?**
  _75 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Community 0` be split into smaller, more focused modules?**
  _Cohesion score 0.03506721215663355 - nodes in this community are weakly interconnected._
- **Should `Community 1` be split into smaller, more focused modules?**
  _Cohesion score 0.05928614640048397 - nodes in this community are weakly interconnected._
- **Should `Community 3` be split into smaller, more focused modules?**
  _Cohesion score 0.05365853658536585 - nodes in this community are weakly interconnected._