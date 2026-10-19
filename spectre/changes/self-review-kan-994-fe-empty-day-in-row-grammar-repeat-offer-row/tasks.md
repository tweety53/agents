# self-review-kan-994-fe-empty-day-in-row-grammar-repeat-offer-row

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the seven non-`big` findings of `kan-994-fe-empty-day-in-row-grammar-repeat-offer-row`'s
self-review pass — two dispatch-recording gaps, an unchecked narrative append, an unchecked mutation
path, a verify failure a later green run hid, and two gymie dev-tooling lessons paid twice.

**Architecture:** refusals in `flow record dispatch begin` and `flow record mutation`
(`stats/cmd/flow/record.go`); a new `flow stage sessions` read and a `check-narrative-runs.sh` guard
wired into `flow.write-in-progress` and `integrate-change.sh commit`; signal nine of
`check-unfinished-work` (`stats/internal/guard/unfinishedwork.go`); a Node script in
gymie-playwright with its gymie note; a gymie brief.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-19-self-review-kan-994-fe-empty-day-in-row-grammar-repeat-offer-row-design.md`
in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer — Task 6 makes one commit in each of two repositories, as its steps say. Never write `spectre/changes/` in this agents worktree: this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push. No `Co-Authored-By` or other attribution trailer.
- Each task names the repository it runs in by its worktree; every `**Files:**` path is relative to that worktree's root. Tasks 1–5 run in this agents worktree `/Users/tweety53/Projects/agents-worktrees/self-review-kan-994-fe-empty-day-in-row-grammar-repeat-offer-row` (branch `self-review-kan-994-fe-empty-day-in-row-grammar-repeat-offer-row`). Task 6 runs in the gymie-playwright worktree `/Users/tweety53/Projects/gymie-playwright-worktrees/kan-994-fe-empty-day-in-row-grammar-repeat-offer-row` and, for its note, the gymie worktree `/Users/tweety53/Projects/gymie-worktrees/kan-994-fe-empty-day-in-row-grammar-repeat-offer-row`; Task 7 in that gymie worktree. Both of those are on branch `spectre/kan-994-fe-empty-day-in-row-grammar-repeat-offer-row`. Never touch a main checkout, and never create a worktree.
- Go: run `cd stats && gofmt -w .` before committing; `gofmt -l .` prints nothing and `go vet` stays clean. Open every Go test run with `make -s web-dist` (`stats/Makefile`). Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container, and never write to the live store: every test runs against a fake daemon.
- No new dependency in any repository.

## Review Focus

- Task 1: a non-panel reviewer key (`self-review-…-review-1`, `pipeline-fix-…`) and every non-reviewer role still record with no `-slot`.
- Task 2: a store that is unreachable or answers not-found still lets `begin` journal or write — the round check never turns an outage into a refusal; a round equal to the highest recorded is admitted.
- Task 3: a session stopped at the unfinished-work gate, and a run 2 cleanup, owe no section; a narrative file that does not exist reads as zero sections, not as exit 2.
- Task 4: the exemption form (`-mutated none`) is held to the same path rule; a path in a peer repository's worktree named by the state record is admitted.
- Task 5: a capture touched only before the report that failed it is still reported; a passing report (`- verify: exit 0`) contributes nothing; a failure cited in the narrative with a sha outside every range is still reported.

- [ ] 1. Refuse a panel reviewer dispatch without its slot
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In this agents worktree, write the test below, then run `cd stats && make -s web-dist && go test ./cmd/flow/ -run 'TestRecordDispatchBeginRefusesAPanelReviewerWithoutSlot' -count=1`: it fails, the call exiting 0 and reaching the fake daemon.
  - [ ] **Step 2: Write `TestRecordDispatchBeginRefusesAPanelReviewerWithoutSlot`** in `stats/cmd/flow/record_test.go`: through `dispatchBeginKey`'s fake daemon, `-role reviewer -key panel-1-primary` with no `-slot` exits 2, its stderr names `-slot`, and the daemon is never contacted; `-role reviewer -key self-review-demo-review-1` and `-role implementer -key task-1-implementer`, both with no `-slot`, exit 0.
  - [ ] **Step 3: Implement** the refusal in `runRecordDispatchBegin` (`stats/cmd/flow/record.go`), beside the existing panel-key shape check: `-role reviewer`, a `panel-` key and an empty `-slot` print `flow: -key %q is a review-panel reviewer key: pass -slot <slot|slot+slot+slot> (skills/flow/review-panel.md)` and return 2 (decision `panel-reviewer-slot-required`). Give `dispatchBeginKey` a `-slot primary` whenever its role is `reviewer` and its key starts `panel-`, and add `-slot` to the two tests that record a `panel-` reviewer row without one: `TestRecordDispatchSendsAgentIDOnlyWhenGiven` (`-slot primary`) and the diff-base test keyed `panel-1-principles` (`-slot principles`).
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l . && go vet ./cmd/flow/ && go test ./cmd/flow/ -run 'TestRecordDispatch' -count=1` passes — every dispatch test, the edited ones included.
  - [ ] **Step 5: Commit** `stats/cmd/flow/record.go` and `stats/cmd/flow/record_test.go` with the subject below.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordDispatchBeginRefusesAPanelReviewerWithoutSlot`
**Regression:** `TestRecordDispatchBeginRefusesAPanelReviewerWithoutSlot` — reverting the commit records a panel reviewer row with no slot, and the test fails on exit 0 and a contacted daemon.
**Baseline:** before=116 after=117
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ branch self-review-kan-994-fe-empty-day-in-row-grammar-repeat-offer-row (116 at bc701950) -->
**Commit:** fix(flow-cli): refuse a panel reviewer dispatch without its slot
**After:** none

**Decision:** panel-reviewer-slot-required

- [ ] 2. Refuse a panel round below the rounds already recorded
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** With Task 1 committed, write the test below and run `cd stats && make -s web-dist && go test ./cmd/flow/ -run 'TestRecordDispatchBeginRefusesAPanelRoundBelowTheRecorded' -count=1`: the below-floor subtest fails, `begin` exiting 0 and posting the row.
  - [ ] **Step 2: Write `TestRecordDispatchBeginRefusesAPanelRoundBelowTheRecorded`** in `stats/cmd/flow/record_test.go`, one fake daemon per subtest that answers `GET …/records/<project>/<change>` and counts `POST`s:
    - the record holds a pass row of round 3 and a reviewer dispatch keyed `panel-2-primary`: `-key panel-1-primary -slot primary` exits 2, nothing is posted, stderr names `3` and `flow record next-round`;
    - the same record: `panel-3-primary` and `panel-4-primary` each exit 0 and post one row;
    - the record read answers 404: `panel-0-primary` exits 0 and posts;
    - the record holds only a mutation row of round 2: `panel-1-primary` exits 2.
  - [ ] **Step 3: Implement** in `runRecordDispatchBegin`, after the flag checks and before the write: for a key `records.PanelDispatchRound` reads as round `n`, read the change's run record with `callRecord(…, cl.GetRunRecord)`; on success take the highest round over its `Passes`, `Findings`, `Mutations` and the `PanelDispatchRound` of its `Dispatches`, and when `n` is below it print `flow: -key %q opens panel round %d, below round %d the change's records already carry — a run's first round is what `flow record next-round` prints (skills/flow/review-panel.md)` and return 2. A read error of any kind skips the check (decision `panel-round-floor`). No other role or key reads the record.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l . && go vet ./cmd/flow/ && go test ./cmd/flow/ -run 'TestRecordDispatch' -count=1` passes — every dispatch test, since each panel reviewer begin now reads the record first.
  - [ ] **Step 5: Commit** `stats/cmd/flow/record.go` and `stats/cmd/flow/record_test.go` with the subject below.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordDispatchBeginRefusesAPanelRoundBelowTheRecorded`
**Regression:** `TestRecordDispatchBeginRefusesAPanelRoundBelowTheRecorded` — reverting the commit posts a `panel-1-…` row over recorded round 3, and the below-floor subtests fail on exit 0.
**Baseline:** before=117 after=118
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ branch self-review-kan-994-fe-empty-day-in-row-grammar-repeat-offer-row (116 at bc701950, plus Task 1's one) -->
**Commit:** fix(flow-cli): refuse a panel round below the rounds already recorded
**After:** Task 1

**Decision:** panel-round-floor

- [ ] 3. Check the narrative's run sections against the store's stage marks
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Confirm the gap the check closes: `grep -c -- '— fix run' /Users/tweety53/Projects/gymie-worktrees/kan-994-fe-empty-day-in-row-grammar-repeat-offer-row/spectre/changes/archive/kan-994-fe-empty-day-in-row-grammar-repeat-offer-row/narrative.md` prints `0`, while `curl -s 'http://127.0.0.1:4173/api/v1/stage-runs?project_key=gymie-7c1f238a&name=kan-994-fe-empty-day-in-row-grammar-repeat-offer-row&command=/flow&stage=flow.document-fix&limit=1000'` (a read) answers one row; and `go run ./cmd/flow stage sessions` (from `stats/`) is refused as an unknown stage command.
    <!-- measured: the grep and the GET above @ the kan-994 gymie worktree and live flowd, 2026-10-19 → 0 headings, 1 session (0cdee296) -->
  - [ ] **Step 2: Write `TestRunStageSessions`** in `stats/cmd/flow/stage_test.go`: a fake daemon asserts the request is `GET /api/v1/stage-runs` carrying `project_key` (the key `-C` resolves), `name=<change>`, `command=/flow`, `stage=flow.document-fix` and `limit=1000`, and answers `{"total":4,"stageRuns":[{"sessionId":"a"},{"sessionId":"a"},{"sessionId":"b"},{}]}`; `flow stage sessions -C <repo> -stage flow.document-fix <change>` prints `3` and exits 0. A second subtest against `deadPortAddr(t)` exits 1 with nothing on stdout; a third, with an undocumented `-stage`, exits 2.
  - [ ] **Step 3: Implement `flow stage sessions`.** In `stats/internal/client/client.go`, `func (c *Client) ListStageRuns(ctx context.Context, project, change, command, stage string) ([]StageRunRow, error)` with `type StageRunRow struct { SessionID *string `json:"sessionId,omitempty"` }`, a `GET` built like `FindState`'s (daemon header required, status 200, `stageRuns` decoded). In `stats/cmd/flow/stage.go`, a `sessions` case: flags `-addr`, `-timeout`, `-C`, `-stage` (validated against the stage-key table, exit 2 on an unknown key) and the change as the one positional argument; it prints the number of distinct non-empty `sessionId`s plus one per row with none, and exits 1 with the error on stderr when the read fails. Add the usage line to `stageUsage`. (Decision `narrative-runs-from-stage-marks`.)
  - [ ] **Step 4: Write the guard and its harness.** `scripts/check-narrative-runs.sh <worktree> <change> <narrative.md>` (bash, `set -euo pipefail`, header comment stating usage, exit codes and decision `narrative-runs-from-stage-marks`): `f` and `i` from `flow stage sessions -C <worktree> -stage flow.document-fix <change>` and `-stage flow.preserve-sessions`, each required to print one integer (else exit 2 naming the read); `hf`/`hi` the counts of lines matching `^## .* — fix run` and `^## .* — integrate run` in the narrative, 0 when the file does not exist, exit 2 when it exists and cannot be read. When `hf < f` or `hi < i` it prints one line, `NARRATIVE-RUNS-MISSING: <narrative> holds <hf> '— fix run' section(s) for <f> fix run(s) and <hi> '— integrate run' section(s) for <i> integrate run(s) the store marked`, and exits 1; otherwise `NARRATIVE-RUNS-OK: <f> fix run(s), <i> integrate run(s), each with its section` and exit 0. Add the symlink `skills/flow/scripts/check-narrative-runs.sh -> ../../../scripts/check-narrative-runs.sh`. `scripts/test-check-narrative-runs.sh` runs the guard against a stub `flow` on `PATH` answering `stage sessions` per stage from fixture files, with the checks `a fix run with no section exits 1`, `an integrate run with no section exits 1`, `every run with its section exits 0`, `an absent narrative with no runs exits 0` and `a failed store read exits 2`.
  - [ ] **Step 5: Wire it.** In `scripts/integrate-change.sh` `cmd_commit`, after the in-run fix line loop: the narrative is the archived `narrative.md` when only it exists, else the live one; `run "$SCRIPT_DIR/check-narrative-runs.sh" "$CANON" "$NAME" "$nf"`, and a non-zero `RC` stops `narrative` with the guard's last line and `— append each missing run's section (skills/flow/verify-and-handoff.md, Write IN_PROGRESS), then re-run commit`. Extend the header's `commit` paragraph by one sentence naming this stop. In `scripts/test-integrate-change.sh`, the stub `flow` answers `stage sessions` with the contents of `$WORK/fix-sessions` for `flow.document-fix` (`0` when absent) and `0` otherwise, and a new case `narrative_runs: a fix run with no section stops commit at STOP: narrative` writes `1` there, runs `prepare` and `commit`, and asserts exit 1 with a last line opening `STOP: narrative — NARRATIVE-RUNS-MISSING`; appending `## 2026-10-19 — fix run` and re-running `commit` no longer stops at `narrative`. Then insert two paragraphs, exactly: the first below as its own paragraph in `skills/flow/verify-and-handoff.md`, right after the one ending `made once the append lands; nothing else stages it.`; the second as its own paragraph in `skills/flow/integrate.md`, right after the existing `**`STOP: narrative`**` paragraph.

  ```markdown verified:applied at plan time; check-verbatim-moves.sh printed its sentences, listed in verbatim-moves.txt
  **Then check the narrative against the runs the store marked:** `check-narrative-runs.sh <abs-worktree> <name> <narrative.md>`. Exit 1 names a fix or integrate run whose `## <YYYY-MM-DD> — fix run` or `— integrate run` section the file lacks: append that run's section from the change's ledger and history, and run the check again. Exit 2 is reported and stops the write.
  ```

  ```markdown verified:applied at plan time; check-verbatim-moves.sh printed its sentence, listed in verbatim-moves.txt
  **`STOP: narrative`** also stops on a fix or integrate run whose section `narrative.md` lacks (`check-narrative-runs.sh`, the check `flow.write-in-progress` runs): append the missing section and re-run `commit`.
  ```

  - [ ] **Step 6: Verify.** `cd stats && gofmt -l . && go vet ./... && go test ./cmd/flow/ -run 'TestRunStageSessions' -count=1` passes; `bash scripts/test-check-narrative-runs.sh` and `bash scripts/test-integrate-change.sh` end with no failure; `scripts/check-verbatim-moves.sh`, `scripts/check-guard-symlinks.sh`, `scripts/check-references.sh` and `scripts/check-markdown-integrity.py` exit 0.
  - [ ] **Step 7: Commit** the ten files below with the subject below.

**Build:** green
**Files:** `stats/internal/client/client.go`, `stats/cmd/flow/stage.go`, `stats/cmd/flow/stage_test.go`, `scripts/check-narrative-runs.sh`, `scripts/test-check-narrative-runs.sh`, `skills/flow/scripts/check-narrative-runs.sh`, `scripts/integrate-change.sh`, `scripts/test-integrate-change.sh`, `skills/flow/verify-and-handoff.md`, `skills/flow/integrate.md`
**Tests:** `TestRunStageSessions`, `a fix run with no section exits 1`, `an integrate run with no section exits 1`, `every run with its section exits 0`, `an absent narrative with no runs exits 0`, `a failed store read exits 2`, `narrative_runs: a fix run with no section stops commit at STOP: narrative`
**Regression:** `TestRunStageSessions` — reverting removes `flow stage sessions` and the test fails on an unknown stage command; `a fix run with no section exits 1` — without the guard the harness has nothing to run; `narrative_runs: a fix run with no section stops commit at STOP: narrative` — reverting lets `commit` reshape over a fix run with no section.
**Baseline:** before=35 after=36
<!-- measured: grep -c '^func Test' stats/cmd/flow/stage_test.go @ branch self-review-kan-994-fe-empty-day-in-row-grammar-repeat-offer-row (35 at bc701950) -->
**Commit:** feat(scripts): stop a run whose narrative lacks a fix or integrate run's section
**After:** none

**Decision:** narrative-runs-from-stage-marks
**Decision:** narrative-check-wiring

- [ ] 4. Refuse a mutation path no worktree of the change holds
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** With Task 2 committed, write the test below and run `cd stats && make -s web-dist && go test ./cmd/flow/ -run 'TestRecordMutationRefusesAPathNoWorktreeHolds' -count=1`: the missing-path subtest fails, the row posted.
  - [ ] **Step 2: Write `TestRecordMutationRefusesAPathNoWorktreeHolds`** in `stats/cmd/flow/record_test.go`, a fake daemon that answers `GET …/changes/<project>/<change>` with `{"worktrees":{"<peer dir>":"<sha>"}}` and counts mutation `POST`s; `<peer dir>` is a second temp directory holding `peer/Fixed.kt`, and the `-C` repo holds `local/Fixed.kt`:
    - `-path src/Missing.kt` exits 2, posts nothing, and stderr names `-path`, the `-C` repo and `<peer dir>`;
    - `-path local/Fixed.kt`, `-path peer/Fixed.kt` and `-path <abs path of local/Fixed.kt>` each exit 0 and post;
    - with the change read answering 404, `-path peer/Fixed.kt` exits 2 and `-path local/Fixed.kt` exits 0.
    Make `TestRecordMutationCommand` create `render.go` in its repo, so its `-path render.go` names a file.
  - [ ] **Step 3: Implement** in `runRecordMutation`, after the required-flag check and before the write: candidates are `git -C <dir> rev-parse --show-toplevel` and every key of the `worktrees` map in the change's state record (`getChange` with the project key `-C` resolves; a failed read adds none); an absolute `-path` is checked as given, a relative one joined to each candidate; no regular file found prints `flow: record mutation: -path %q is no file in %s — name the mutated file relative to its worktree` (the roots joined with `, `) and returns 2 (decision `mutation-path-exists`).
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l . && go vet ./cmd/flow/ && go test ./cmd/flow/ -run 'TestRecordMutation|TestRecordPassAndMutation' -count=1` passes.
  - [ ] **Step 5: Commit** `stats/cmd/flow/record.go` and `stats/cmd/flow/record_test.go` with the subject below.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordMutationRefusesAPathNoWorktreeHolds`
**Regression:** `TestRecordMutationRefusesAPathNoWorktreeHolds` — reverting the commit posts a mutation row naming a file no worktree holds, and the missing-path subtest fails on exit 0.
**Baseline:** before=118 after=119
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go @ branch self-review-kan-994-fe-empty-day-in-row-grammar-repeat-offer-row (116 at bc701950, plus Tasks 1 and 2) -->
**Commit:** fix(flow-cli): refuse a mutation path no worktree of the change holds
**After:** Task 1, 2

**Decision:** mutation-path-exists

- [ ] 5. Report a verify-report failure no later commit addressed
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Write the test below and run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckUnfinishedWorkVerifyFailureNotAddressed' -count=1`: the unaddressed-failure subtest fails, the gate answering with no such reason.
  - [ ] **Step 2: Write `TestCheckUnfinishedWorkVerifyFailureNotAddressed`** in `stats/internal/guard/check_unfinished_work_test.go`, on the file's existing fixture of a worktree holding the plan and a range. Its report is `.superpowers/sdd/verify-report-visual-verify.md` holding `- verify: exit 1` and the continuation line `  1 failed — fidelity/full-app-suite.spec.ts:110 register-from-invite.png, 24px differ`; report mtime and commit times are set explicitly (`os.Chtimes`, `GIT_COMMITTER_DATE`). Subtests:
    - no commit touches the capture: the verdict names `verify failure not addressed: fidelity/full-app-suite.spec.ts:110 register-from-invite.png (verify-report-visual-verify.md)`;
    - a commit after the report touches `fidelity/full-app-suite.spec.ts-snapshots/register-from-invite-darwin.png`: no such reason;
    - that commit dated before the report: the reason stands;
    - no commit, and the canonical `narrative.md` holds `Verify failure addressed: fidelity/full-app-suite.spec.ts:110 — <sha of a commit in the range>`: no such reason; the same line with a sha outside the range: the reason stands;
    - a failure naming no capture (`  1 failed — tests/auth.spec.ts:42 times out`) and a later commit touching `tests/auth.spec.ts`: no such reason;
    - the report's verify item reads `- verify: exit 0` above the same continuation line: no such reason.
  - [ ] **Step 3: Implement signal nine** in `stats/internal/guard/unfinishedwork.go`, where signal three reads (the worktree holding the plan), whenever a range exists, archived plan or not, per decision `verify-failure-signal`: glob `<worktree>/.superpowers/sdd/verify-report-*.md`; in each, the `- verify: exit <n>` item with `<n>` not 0 and its following lines indented by whitespace, up to the next `- ` item; every `<path>.spec.ts:<line>` there a failure, its capture the first `<name>.png` after it before the next `.spec.ts:`; a range commit's paths read with `git -C <wt> log --no-merges --format=%x00%H %ct --name-only <base>..HEAD`; a failure is addressed per the decision; each unaddressed one appends the reason `verify failure not addressed: <spec>:<line>[ <name>.png] (<report base name>) — fix it in this run, or cite its fixing commit as a narrative line "Verify failure addressed: <spec>:<line> — <sha>"`. An unreadable report is exit 2 like any unreadable file here. Add a `SIGNAL NINE` paragraph to `scripts/check-unfinished-work.sh`'s header after signal eight's, and a matching comment in the Go.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l . && go vet ./internal/guard/ && go test ./internal/guard/ -run 'TestCheckUnfinishedWork' -count=1` passes; `bash scripts/test-integrate-change.sh` ends with no failure.
  - [ ] **Step 5: Commit** the three files below with the subject below.

**Build:** green
**Files:** `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`, `scripts/check-unfinished-work.sh`
**Tests:** `TestCheckUnfinishedWorkVerifyFailureNotAddressed`
**Regression:** `TestCheckUnfinishedWorkVerifyFailureNotAddressed` — reverting the commit clears a change whose verify report failed a capture nothing touched since, and the unaddressed subtest fails.
**Baseline:** before=4 after=5
<!-- measured: grep -c '^func Test' stats/internal/guard/check_unfinished_work_test.go @ branch self-review-kan-994-fe-empty-day-in-row-grammar-repeat-offer-row (4 at bc701950) -->
**Commit:** fix(guard): report a verify-report failure no later commit addressed
**After:** none

**Decision:** verify-failure-signal

- [ ] 6. Re-baseline changed captures in one run and box each diff
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In the gymie-playwright worktree, `ls scripts/` lists no re-baseline script, and `npx playwright test --help` lists `--update-snapshots [mode]` with the choice `changed`.
    <!-- measured: ls scripts; npx playwright test --help @ branch spectre/kan-994-fe-empty-day-in-row-grammar-repeat-offer-row (gymie-playwright) → specs-for-diff only; choices all, changed, missing, none -->
  - [ ] **Step 2: Write `scripts/rebaseline.test.mjs`** (`node:test`, `node:assert/strict`, like the `specs-for-diff` test beside it) over `diffBox(a, b)` with synthetic `{ width, height, data }` RGBA images: `identical images have no diff box` (returns `null`); `the box bounds every differing pixel` (a 4x4 pair differing at (1,2) and (3,3) returns `{ x: 1, y: 2, w: 3, h: 2, pixels: 2 }`); `a size change is reported, not boxed` (a 4x4 against a 4x5 returns `{ resized: true }`).
  - [ ] **Step 3: Write `scripts/rebaseline.mjs`** per decision `rebaseline-script`: `export function diffBox(a, b)`; run as `node scripts/rebaseline.mjs <spec>...` (guarded by `process.argv[1] === fileURLToPath(import.meta.url)`, as `specs-for-diff.mjs` is) it refuses an empty spec list with a usage line and exit 2, spawns `npx playwright test <spec>... --update-snapshots=changed` with `stdio: 'inherit'` and the caller's environment, then reads `git status --porcelain=v1 -z -- '*.png'` and prints one line per modified or new PNG — `<path> — <x>,<y> <w>x<h>, <n> px`, `<path> — size <w>x<h> → <w>x<h>`, or `<path> — new` — decoding `git show HEAD:<path>` and the working file with `createRequire(import.meta.url)('playwright-core/lib/utilsBundle').PNG.sync.read`, and exits with Playwright's status. A header comment names the capture row's `FRONTEND_URL`/`BACKEND_URL` prefix the caller sets, and says specs are arguments so no shell splits a variable.
    <!-- measured: node -e with PNG.sync.read over fidelity/full-app-suite.spec.ts-snapshots/register-from-invite-darwin.png at fbef27b and HEAD @ branch spectre/kan-994-fe-empty-day-in-row-grammar-repeat-offer-row (gymie-playwright) → 1200x1164, 1524 px, box 460,424 282x22 -->
  - [ ] **Step 4: Verify.** `node --test scripts/rebaseline.test.mjs` passes; `npm run test:scripts` passes.
  - [ ] **Step 5: Commit in gymie-playwright** `scripts/rebaseline.mjs` and `scripts/rebaseline.test.mjs` with the subject below.
  - [ ] **Step 6: Note it in gymie.** In the gymie worktree's `.flow/project.md` `## visual verification`, insert as its own paragraph right after the paragraph opening `` `specs` prints the specs reaching a tag ``:

  ```markdown verified:authored for this change; the script and its flag were checked at plan time (Step 1, Step 3)
  **A change that moves captures re-baselines them with `node scripts/rebaseline.mjs <spec>...`** in the regression checkout's worktree, under the `capture` row's `FRONTEND_URL`/`BACKEND_URL` prefix: one Playwright run with `--update-snapshots=changed`, then one line per changed PNG with its diff's bounding box against `HEAD`. A box outside the region the change moved names a capture to restore with `git checkout -- <png>`, not to commit. The specs are arguments — never one unquoted variable, which zsh does not split (kan-994).
  ```
  - [ ] **Step 7: Commit in gymie** `.flow/project.md` alone, with the subject `docs(flow): name the re-baseline script under visual verification` (decision `rebaseline-note-own-commit`).

**Build:** green
**Files:** `scripts/rebaseline.mjs`, `scripts/rebaseline.test.mjs` (the gymie-playwright worktree Global constraints names), `.flow/project.md` (the gymie worktree Global constraints names)
**Tests:** `identical images have no diff box`, `the box bounds every differing pixel`, `a size change is reported, not boxed`
**Regression:** `the box bounds every differing pixel` — reverting the commit removes `diffBox`, and the import fails.
**Baseline:** before=0 after=3
<!-- measured: ls scripts/rebaseline.test.mjs (absent) @ branch spectre/kan-994-fe-empty-day-in-row-grammar-repeat-offer-row (gymie-playwright); after = the three test( calls Step 2 names -->
**Commit:** feat(scripts): re-baseline changed captures in one run and box each diff
**After:** none

**Decision:** rebaseline-script
**Decision:** rebaseline-note-own-commit

- [ ] 7. Brief Compose text on a CSS line box
  - [ ] **Step 1: Confirm the copies.** In the gymie-frontend worktree `/Users/tweety53/Projects/gymie-frontend-worktrees/kan-994-fe-empty-day-in-row-grammar-repeat-offer-row`, `git grep -n 'LineHeightStyle(' -- shared/src/commonMain` lists the declarations `design.md`'s Context names; a path that moved is cited where it now is.
  - [ ] **Step 2: Write `docs/briefs/compose-text-on-a-css-line-box.md`** in the gymie worktree, exactly:

  ```markdown verified:authored from kan-933's and kan-994's narratives and the gymie-frontend declarations read at plan time
  # Compose text on a CSS line box

  Gymie's frames are HTML and CSS. A browser lays a `line-height: L` line out as a box exactly `L`
  tall, splitting the leading evenly above and below the glyphs. Compose's default
  `LineHeightStyle` trims a lone line's outer leading to the font's own height, and Source Serif 4
  is taller than most of the frames' line heights — so a Compose text drawn to a frame's numbers
  stands taller than drawn and its glyphs sit lower.

  ## The rule

  Give a text drawn to a frame's line height
  `LineHeightStyle(LineHeightStyle.Alignment.Center, LineHeightStyle.Trim.None)`: the whole line
  height kept, the glyphs centred in it, as the browser centres the half-leading. Reuse
  `FRAME_LINE_BOX` (`shared/src/commonMain/kotlin/com/gymie/shared/ui/exercise/ExerciseInfoSheet.kt`,
  `internal`) rather than declaring another copy; the existing private copies are `CSS_LINE_BOX`
  (`ui/cardio/CardioPlanParts.kt`), `FORM_TITLE_LINE_HEIGHT_STYLE` (`ui/components/FormSheet.kt`)
  and an inline one on the empty day's heading (`ui/HomeScreen.kt`).

  ## Measured exceptions

  `Trim.None` always; the alignment is measured against the frame, never assumed:

  - `Alignment.Bottom` for the 11px caps label (`ui/theme/BroadsheetType.kt`): Blink rounds the
    font's ascent and descent to whole pixels and floors a negative half-leading, so the browser's
    baseline sits higher than a centred one (KAN-941).
  - `Alignment.Bottom` for LAST TIME's kind line (`ui/cardio/CardioPlanParts.kt`) and
    `Alignment.Top` for the editor bar's Save (`ui/workout/EditorBar.kt`), each measured on its
    frame.

  ## History

  - kan-933: a confirm sheet stood 6px off its frame — Compose's trimmed box made two 24sp lines
    46dp tall.
  - kan-994: *Nothing planned* measured 41dp with its padding against the frame's 38;
    `Trim.None` restored the 30px line, and `Alignment.Center` closed the last gap where `Bottom`
    left the glyphs 2px high.

  ## Verifying

  Compare the capture's text bands with the frame's (`flow.visual-verify`'s band diff), not the
  look. Read the build output before trusting a test result: in kan-994 a missing
  `LineHeightStyle` import broke the compile, Gradle kept the previous run's results, and the fix
  looked as if it had not worked.
  ```
  - [ ] **Step 3: Verify.** From this agents worktree, `go run ./cmd/flow lesson resolve -topic 'css line box'` (from `stats/`) prints the brief first once the commit exists in the gymie worktree — a read of the live daemon; if the daemon scans only the main checkout, record that and move on.
    <!-- unverified:whether flow lesson resolve scans worktrees or only registered main checkouts -->
  - [ ] **Step 4: Commit in gymie** `docs/briefs/compose-text-on-a-css-line-box.md` alone, with the subject below.

**Build:** green
**Files:** `docs/briefs/compose-text-on-a-css-line-box.md` (the gymie worktree Global constraints names)
**Tests:** **none** — a brief; Step 3's resolve is its check.
**Regression:** none — prose.
**Commit:** docs(briefs): brief Compose text on a CSS line box
**After:** none

**Decision:** line-box-brief
