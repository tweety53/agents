# self-review-kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the nine non-`big` findings of the self-review pass of gymie's
`kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer` — 1.1 through 1.7, 2.1 and 8.1.

**Architecture:** four Go gates and fixes in `stats/` (JSON escaping, close-task's suite-zip
check, the verifier-row admission, the visual-verify stage-end trailer gate), the verify-report
guard's per-frame matrix line, five sentences in the verifier contract and one in **The loop**,
and in gymie a log rotation on dev-stack start and a trace flag on the `verify` row.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-27-self-review-kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` in either repository — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Tasks 1–7 work in this worktree `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer`, branch `self-review-kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer`; their `**Files:**` paths are relative to its root.
- Tasks 8–9 work in the gymie worktree `/Users/tweety53/Projects/gymie-worktrees/kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer`, branch `spectre/kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer`; their `**Files:**` paths are relative to that root. Each commit there carries `-m "Flow-Key: flow.self-review"`. Never start, stop or restart that worktree's running dev stack, and never run its Playwright suite. Every Gradle call sets the Bash tool's maximum timeout and runs in the foreground; no two Gradle calls run at once in that worktree.
- Never touch a main checkout, and never create a worktree.
- A task whose fix step gives a unified diff applies it with `git apply` from the worktree's root; each was applied in order and checked in this worktree at `1be0e995`. A diff that does not apply is reported, never hand-merged by guesswork. Each prose sentence lands character for character: `scripts/check-verbatim-moves.sh` matches it against `verbatim-moves.txt` exactly.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.
- Every new test fails with its task's fix reverted: run it once against the reverted source before committing and state the failure in the report.

## Review Focus

- Task 2: the check rebuilds from the worktree and writes the committed zip bytes back, so a refused close leaves the tree as the commit left it; a commit touching no suite zip is untouched.
- Task 3: only `-role verifier` skips the dirty-tree refusal; an inline implementer begin over tracked modifications is still refused.
- Task 4: the gate runs only on `/flow` `flow.visual-verify` `-outcome completed`, over every worktree of the state map, exempting exactly what `check-untasked-commits` exempts.
- Task 8: every reader of `<service>.log` still reads the running service's output.

- [x] 1. Escape the strings the JSON writer prints
  - [x] **Step 1: Write the failing test.** Add `TestPyJSONEscapesStrings` to `stats/internal/guard/measure_visual_properties_test.go`: `pyDump(pyObj{{"unresolved", "region corner (0,9) is not background\nregion \"b\" \\ \t\x01 — end"}})` passes `json.Valid`, unmarshals back to the same string, and its bytes carry `\n`, `\"`, `\\`, `\t` and `\u0001` escapes with the `—` raw. Run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestPyJSON' -count=1`: the new test fails (an invalid control character).
  - [x] **Step 2: Fix.** In `pyWrite`'s `case string` (`stats/internal/guard/pyjson.go`) write the string through one helper that escapes `"` and `\` with a backslash, `\n` `\r` `\t` `\b` `\f` by their short forms, every other rune below U+0020 as `\u00XX` (lower-case hex, as Python's `json.dumps`), and every other rune raw (decision `pyjson-escape`); route object keys (`kv.k`) through the same helper. Replace the comment claiming nothing needs escaping with one naming the two-sided `unresolved` row (`measurevisualproperties.go`) as the string that does.
  - [x] **Step 3: Verify.** `cd stats && gofmt -l . && make -s web-dist && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestPyJSON|TestMeasureVisualProperties' -count=1` — green, and `TestMeasureVisualProperties`' golden outputs unchanged.

**Build:** green
**Files:** `stats/internal/guard/pyjson.go`, `stats/internal/guard/measure_visual_properties_test.go`
**Tests:** `TestPyJSONEscapesStrings`
**Regression:** `TestPyJSONEscapesStrings` — reverting the commit writes the raw newline and quote, and `json.Valid` fails.
**Baseline:** before=339 after=340
<!-- measured: cat stats/internal/guard/*_test.go | grep -c -E '^func Test' @ branch self-review-kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer -->
**Commit:** fix(guard): JSON-escape every string the matrix writer prints
**After:** none

- [x] 2. Refuse a hand-built full app suite zip at close-task
  - [x] **Step 1: Write the failing test.** Add `TestCloseTaskRefusesAHandBuiltSuiteZip` to `stats/internal/guard/close_task_test.go`, built on the fixtures `stats/internal/guard/rebuild_full_app_suite_zip_test.go` uses for a screenshots root the resolver reads: a task commit whose `full-app-suite.zip` was written with `archive/zip` (not the script) is refused with exit 1 and the `is not rebuild-full-app-suite-zip.sh's output` line, the file's bytes after the run equal the committed ones, and the same commit carrying the zip the script builds closes. Run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCloseTaskRefusesAHandBuiltSuiteZip' -count=1`: it fails.
  - [x] **Step 2: Fix.** In `closeTask` step 2 (`stats/internal/guard/closetask.go`), per task commit and per map pair, when `git diff-tree --no-commit-id --name-only -r <sha>` lists a path whose base name is `full-app-suite.zip`, run `Registry["rebuild-full-app-suite-zip"]` in-process on that worktree with the zip's absolute directory as `<screenshots root>` (with `FLOW_GUARD_SCRIPT_DIR` set as the shim sets it), compare the rebuilt file's bytes with `git show <sha>:<path>`, write the committed bytes back, and fold the result into `worst`: a difference is exit 1, printed as `close-task: task <id>: <path> is not rebuild-full-app-suite-zip.sh's output — rebuild it with that script and re-commit`; a dirty screenshots root (`git status --porcelain -- <root>` non-empty) or a rebuild exit other than 0 is exit 2 (decision `suite-zip-at-close-task`). Add the step to `scripts/close-task.sh`'s header, step 2.
  - [x] **Step 3: Verify.** `cd stats && gofmt -l . && make -s web-dist && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestCloseTask|TestRebuildFullAppSuiteZip' -count=1`, then `scripts/check-guard-symlinks.sh` — green.

**Build:** green
**Files:** `stats/internal/guard/closetask.go`, `stats/internal/guard/close_task_test.go`, `scripts/close-task.sh`
**Tests:** `TestCloseTaskRefusesAHandBuiltSuiteZip`
**Regression:** `TestCloseTaskRefusesAHandBuiltSuiteZip` — reverting the commit closes the task over the hand-built zip with exit 0.
**Baseline:** before=340 after=341
<!-- measured: cat stats/internal/guard/*_test.go | grep -c -E '^func Test' @ branch self-review-kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer (before Task 1, plus Task 1's one test) -->
**Commit:** fix(guard): refuse a full app suite zip the rebuild script did not build
**After:** Task 1

- [x] 3. Admit an inline verifier row over a dirty tree
  - [x] **Step 1: Write the failing test.** Add `TestRecordDispatchBeginInlineDirtyTree` to `stats/cmd/flow/record_test.go`: in a scratch repository whose `spectre/changes/x/visual-verification.md` is committed then modified, `flow record dispatch begin -change x -role verifier -agent-id inline -key verify-x …` is not refused for the dirty tree, and the same begin with `-role implementer` (a key its role admits) still is, its stderr naming the path. Run `cd stats && make -s web-dist && go test ./cmd/flow/ -run 'TestRecordDispatchBeginInlineDirtyTree' -count=1`: the verifier case fails.
  - [x] **Step 2: Fix.** In `stats/cmd/flow/record.go`, guard the `inlineDirtyPaths` block with `*role != "verifier"` as well, its comment naming why: a verify row records a run, not an edit, and `run-verify-list` already writes no green sha over a dirty tree (decision `verifier-skips-inline-dirty`).
  - [x] **Step 3: Verify.** `cd stats && gofmt -l . && make -s web-dist && go vet ./cmd/flow/ && go test ./cmd/flow/ -run 'TestRecordDispatch' -count=1` — green.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordDispatchBeginInlineDirtyTree`
**Regression:** `TestRecordDispatchBeginInlineDirtyTree` — reverting the commit refuses the verifier begin over the modified record.
**Baseline:** before=336 after=337
<!-- measured: cat stats/cmd/flow/*_test.go | grep -c -E '^func Test' @ branch self-review-kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer -->
**Commit:** fix(flow): admit an inline verifier row over tracked modifications
**After:** none

- [x] 4. Refuse the visual-verify stage end over a commit without its trailer
  - [x] **Step 1: Write the failing test.** Add `TestStageEndVisualVerifyRefusesAnUntaskedCommit` to `stats/cmd/flow/stage_test.go`, in the shape of the existing `gateLiveVerification` tests: a change whose state `worktrees` map names a second repository carrying a commit with neither `Task-Id:` nor `Flow-Key:` trailer is refused by `flow stage end -command '/flow' -stage flow.visual-verify -outcome completed <name>` with exit 1 naming the commit; after the commit is amended with `Flow-Key: flow.visual-verify` the gate passes; `-outcome stopped` is never gated. Run `cd stats && make -s web-dist && go test ./cmd/flow/ -run 'TestStageEndVisualVerifyRefusesAnUntaskedCommit' -count=1`: it fails.
  - [x] **Step 2: Fix.** Add `gateVisualVerifyTrailers(f, outcome, stderr)` to `stats/cmd/flow/stage.go`, called in `runStageEnd` beside the other two gates, for `/flow` `flow.visual-verify` `completed` only: resolve the change's checkout as `gatePanelFindingsClosed` does with no `-C`, read the `worktrees` map from the change's state, and for each worktree run `guard.Registry["check-untasked-commits"]` in-process from `git merge-base HEAD origin/<base>`, `<base>` resolved as integrate's unfinished-work gate resolves it for that worktree (`stats/internal/guard/integratechange.go`). Any listing or a guard exit 2 refuses with exit 1 and `flow: stage end refused: commits outside every task lack their Flow-Key: trailer in <worktree> — amend each with -m "Flow-Key: flow.visual-verify" before anything is pushed` (decision `visual-verify-end-trailer-gate`). unverified:confirm which helper in `stats/cmd/flow/state.go` reads the `worktrees` map and how integratechange.go resolves `<base>` per worktree; reuse both rather than re-deriving.
  - [x] **Step 3: Verify.** `cd stats && gofmt -l . && make -s web-dist && go vet ./cmd/flow/ && go test ./cmd/flow/ -run 'TestStageEnd|TestStage' -count=1` — green.

**Build:** green
**Files:** `stats/cmd/flow/stage.go`, `stats/cmd/flow/stage_test.go`, `stats/internal/guard/integratechange.go`
**Tests:** `TestStageEndVisualVerifyRefusesAnUntaskedCommit`
**Regression:** `TestStageEndVisualVerifyRefusesAnUntaskedCommit` — reverting the commit ends the stage over the trailerless commit with exit 0.
**Baseline:** before=337 after=338
<!-- measured: cat stats/cmd/flow/*_test.go | grep -c -E '^func Test' @ branch self-review-kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer (before Task 3, plus Task 3's one test) -->
**Commit:** fix(flow): refuse the visual-verify stage end over a commit lacking its Flow-Key trailer
**After:** Task 3

- [x] 5. Close the failing visual-verify stage mark in the loop
  - [x] **Step 1: Apply the sentence** (decision `loop-closes-stage-mark`):

````diff verified:applied with git apply to this worktree at 1be0e995; check-verbatim-moves, check-references, check-markdown-integrity, check-stage-mark-calls green
diff --git a/skills/flow/verify-fix-loop.md b/skills/flow/verify-fix-loop.md
index 3079286a..12deb336 100644
--- a/skills/flow/verify-fix-loop.md
+++ b/skills/flow/verify-fix-loop.md
@@ -58,7 +58,7 @@ One round runs these steps in order:
    **One round appends one task**, carrying every defect the report names — never a task or a
    round per defect. Before the append, the stage's failing row is closed with `flow record dispatch
    end` at `-outcome blocked -cause test-failure`, so no row of the stage stays open while this
-   round's rows open.
+   round's rows open. A failing `flow.visual-verify` also has its stage mark closed beside that row, `flow stage end -command '/flow' -stage flow.visual-verify -outcome stopped <name>`, and a run that re-enters this loop while `flow.write-in-progress` is open closes that mark the same way first, so the round's next `flow stage begin` supersedes no open run (kan-1016: `flow.visual-verify` attempts 1–9 and 11 closed `superseded`).
 2. **Implement it.** **4. Execute (SDD + TDD)** (`skills/flow/implement.md`) executes the appended
    task exactly as a plan-time task: the decision's implementer pair, `close-task.sh` with its
    guard and its review gate. The task closes through `close-task.sh` — an inline-committed one included — before step 3 opens any `panel-<round>-*` dispatch; `flow record dispatch begin` refuses a panel key while `tasks.md` holds an unticked task (kan-1007: panel round 4 began 16 seconds after task 8's inline commit, before its close). The implementer's dispatch also carries the MUTATION PROOF
````
  - [x] **Step 2: Verify.** `scripts/check-verbatim-moves.sh && scripts/check-references.sh && scripts/check-markdown-integrity.sh && scripts/check-stage-mark-calls.sh` — green.

**Build:** green
**Files:** `skills/flow/verify-fix-loop.md`
**Tests:** **none** — a prose rule; `check-verbatim-moves.sh` covers its acknowledged sentence.
**Regression:** none — no test is declared.
**Commit:** docs(flow): close the failing visual-verify stage mark in the fix loop
**After:** none

- [x] 6. Name the one matrix file per frame
  - [x] **Step 1: Apply the sentence** (decision `matrix-one-file`):

````diff verified:applied with git apply to this worktree at 1be0e995; check-verbatim-moves green
diff --git a/skills/flow/visual-verify-verifier.md b/skills/flow/visual-verify-verifier.md
index c634edf1..120231e8 100644
--- a/skills/flow/visual-verify-verifier.md
+++ b/skills/flow/visual-verify-verifier.md
@@ -414,7 +414,7 @@ never an exit code the command did not return.
       (kan-1015: the band read took each scrim-dimmed frame as one band, and every matrix was
       measured with a hand-written script). `measure-visual-properties.sh <frame> <capture> --scale <f> --props matrix` then
       fills every column below for every paired element in one call, each region clipped to its
-      image; a row it does not pair takes the per-element calls.
+      image; a row it does not pair takes the per-element calls. Its output is saved once, to `<changeRoot>/visual-verification/matrix/<frame id>.json`, overwritten every round — never a second copy under another name (kan-1016: each frame's matrix was committed three times under three names).
    2. **Fill every column for every row from the script's own output**, one call per element:
 
       | column | text run (`--props ink`, the crop spanning the run's container's full width) | control, row, card (`--props box,radius,border,fill,shadow,content,gap`) | rule, divider, fill band (`--props runs`, the region crossing it) |
````
  - [x] **Step 2: Verify.** `scripts/check-verbatim-moves.sh && scripts/check-references.sh && scripts/check-markdown-integrity.sh` — green.

**Build:** green
**Files:** `skills/flow/visual-verify-verifier.md`
**Tests:** **none** — a prose rule; `check-verbatim-moves.sh` covers its acknowledged sentence.
**Regression:** none — no test is declared.
**Commit:** docs(flow): name the one file a verifier saves a frame's matrix to
**After:** none

- [x] 7. Keep every edited spec's frames on the list and one matrix line per frame
  - [x] **Step 1: Write the failing test.** Add `TestCheckVerifyReportMatrixLinePerFrame` to `stats/internal/guard/check_verify_report_test.go`: a report whose `- steps:` line has `10 done`, with `- templates-W1: <path> diff=0.1` and `- templates-A4b: <path> diff=0.2` lines and only `- templates-W1 matrix: …`, exits 1 naming `templates-A4b`; adding `- templates-A4b matrix: …` exits 0; a report with step 10 not `done` is not judged on it. Run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckVerifyReportMatrixLinePerFrame' -count=1`: it fails.
  - [x] **Step 2: Fix the guard.** In `stats/internal/guard/verifyreport.go`, when step 10's status is `done`, collect every frame id from lines matching `^\s*- (\S+): \S+ diff=` and refuse (exit 1, one line per id: `<frame id> has no matrix: line — step 10 is done only with every frame's matrix`) each with no `^\s*- <frame id> matrix:` line. Add the rule to `scripts/check-verify-report.sh`'s header.
  - [x] **Step 3: Apply the four sentences** (decision `verifier-frame-list-and-matrix-lines`), on top of Task 6's:

````diff verified:applied with git apply on top of Task 6's diff at 1be0e995 (one line of context); check-verbatim-moves, check-references, check-markdown-integrity green
diff --git a/skills/flow/visual-verify-verifier.md b/skills/flow/visual-verify-verifier.md
index 120231e8..388dd605 100644
--- a/skills/flow/visual-verify-verifier.md
+++ b/skills/flow/visual-verify-verifier.md
@@ -173,3 +173,3 @@ never an exit code the command did not return.
    **The difference panel cannot show a structural departure, so the first read of every pair
-   is the script's band pairing, never the panel.** The panel is white wherever any channel
+   is the script's band pairing, never the panel.** The composite's panels are the capture left, the frame middle and the difference right, unlabelled; a radius, a stroke width, a glyph size or a colour is read from `<stem>.frame.png` and the capture PNG themselves, never off the composite (kan-1016: a stepper radius read off the composite was raised in two rounds and withdrawn). The panel is white wherever any channel
    differs, and on a real pair 20–50% of it is white from seed data, font rasterisation and the
@@ -221,3 +221,3 @@ never an exit code the command did not return.
    merge base carries for that spec, read out with `git show <merge-base>:<baseline png path>`
-   into a temporary file and measured like any other image. The comparison decides which of the
+   into a temporary file and measured like any other image. A recorded decision is a named cause only for the deltas it names, by element and property: any other delta on the same band or element is unexplained until measured (kan-1016: a band's `colour` delta was put down to decision `warmup-sublines-w4b` for four rounds, and was the checkbox's size). The comparison decides which of the
    three the departure is, every time: **present at the merge base** — pre-existing, blocking and
@@ -295,3 +295,3 @@ never an exit code the command did not return.
    covered by the composite diff, "shipped" and pre-existing, blocked by a real environment limit)
-   still gets named, with the reason. **The list is the change's, never the round's**: a fix run
+   still gets named, with the reason. The list also carries every frame the `<spec>.mockups` sidecar of each capture spec this change edits maps, whether or not its baseline has moved yet, and two frame ids are never read as one (kan-1016: the warm-up sheet's frame `templates-A4b`, mapped by the edited `templates.spec.ts`'s sidecar, joined the list only in round 7, once its baseline moved). **The list is the change's, never the round's**: a fix run
    re-verifies the same list, not the views its fix touched, and the specs `capture` runs are
@@ -621,2 +621,3 @@ sweep` line `done` or `n/a — <why it does not apply>` — a sweep or a cell le
 or a measurement the pixels could not give is step 10 undone.
+Each composed frame carries its own `- <frame id> matrix:` line, never one line for a range of frames: `check-verify-report.sh` refuses a `done` step 10 while a frame with a `- <frame id>: <composite path> diff=` line has none (kan-1016: one line covered nine frames, and the warm-up sheet's frame had no matrix until round 8).
 
````
  - [x] **Step 4: Verify.** `cd stats && gofmt -l . && make -s web-dist && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestCheckVerifyReport' -count=1`, then `scripts/check-verbatim-moves.sh && scripts/check-references.sh && scripts/check-markdown-integrity.sh && scripts/check-guard-symlinks.sh` — green.

**Build:** green
**Files:** `stats/internal/guard/verifyreport.go`, `stats/internal/guard/check_verify_report_test.go`, `scripts/check-verify-report.sh`, `skills/flow/visual-verify-verifier.md`
**Tests:** `TestCheckVerifyReportMatrixLinePerFrame`
**Regression:** `TestCheckVerifyReportMatrixLinePerFrame` — reverting the commit admits a `done` step 10 with a composed frame no `matrix:` line names.
**Baseline:** before=341 after=342
<!-- measured: cat stats/internal/guard/*_test.go | grep -c -E '^func Test' @ branch self-review-kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer (before Task 1, plus Tasks 1 and 2's one test each) -->
**Commit:** fix(flow): keep an edited spec's frames on the verifier's list, one matrix line each
**After:** Task 2, 6

- [x] 8. Rotate a dev-stack service log on start (gymie)
  - [x] **Step 1: Write the failing test.** Create `buildSrc/src/test/kotlin/com/gymie/gradle/DevLogsTest.kt` with `rotatesAnExistingLogToDotOne`: in a temp directory, `rotateDevLog(File(dir, "backend.log"))` over a non-empty `backend.log` and an older `backend.log.1` leaves `backend.log.1` holding the former `backend.log`'s bytes and no `backend.log` (or an empty one); over an absent or empty log it changes nothing. Run `./gradlew -p buildSrc test --tests 'com.gymie.gradle.DevLogsTest'`: it fails to compile.
  - [x] **Step 2: Fix.** Add `fun rotateDevLog(file: File)` to `buildSrc/src/main/kotlin/com/gymie/gradle/DevLogs.kt` (move a non-empty file to `<name>.1`, replacing an older one, with `Files.move(…, REPLACE_EXISTING)`), and call it on `logFile` in `gradle/dev-lifecycle.gradle.kts` just before the `ProcessBuilder` whose output is `Redirect.appendTo(logFile)` (decision `dev-log-rotate-on-start`). Every reader named in that decision keeps reading `<service>.log`.
  - [x] **Step 3: Verify.** `./gradlew -p buildSrc test --tests 'com.gymie.gradle.DevLogsTest'`, then `./gradlew ktlintFormat`, then `./gradlew verifyChange` — green. Commit with `-m "Flow-Key: flow.self-review"`.

**Build:** green
**Files:** `buildSrc/src/main/kotlin/com/gymie/gradle/DevLogs.kt`, `buildSrc/src/test/kotlin/com/gymie/gradle/DevLogsTest.kt`, `gradle/dev-lifecycle.gradle.kts`
**Tests:** `rotatesAnExistingLogToDotOne`
**Regression:** `rotatesAnExistingLogToDotOne` — reverting the commit removes `rotateDevLog`, and the test no longer compiles.
**Baseline:** before=253 after=254
<!-- measured: cat buildSrc/src/test/kotlin/com/gymie/gradle/*.kt | grep -c -E '^\s*@Test\b' in the gymie worktree @ branch spectre/kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer -->
**Commit:** fix(dev-lifecycle): rotate a service's log on each start
**After:** none

- [x] 9. Keep a Playwright trace for every failing verify-run test (gymie)
  - [x] **Step 1: Fix.** In gymie `.flow/project.md`'s `## visual verification` command table, the `verify` row's `npx playwright test <specs> --output test-results/$(date -u +%Y%m%dT%H%M%SZ)` becomes `npx playwright test <specs> --trace retain-on-failure --output test-results/$(date -u +%Y%m%dT%H%M%SZ)`; nothing else in the row or the `capture` row changes (decision `verify-row-trace-on-failure`).
  - [x] **Step 2: Verify.** `git diff --word-diff` shows the one inserted flag; `/Users/tweety53/Projects/agents/scripts/project-get.sh /Users/tweety53/Projects/gymie-worktrees/kan-1016-templates-step-4-fe-plan-a-week-empty-day-offer 'visual verification'` still prints the table with the flag in the `verify` row. Commit with `-m "Flow-Key: flow.self-review"`.

**Build:** green
**Files:** `.flow/project.md`
**Tests:** **none** — a configuration row; a load-only boot timeout's next occurrence carries its trace.
**Regression:** none — no test is declared.
**Commit:** chore(flow): keep a Playwright trace for every failing verify-run test
**After:** none
