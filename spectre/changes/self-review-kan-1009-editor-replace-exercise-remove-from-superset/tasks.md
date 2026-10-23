# self-review-kan-1009-editor-replace-exercise-remove-from-superset

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the six pipeline defects the kan-1009 self-review found and repair the store rows
two of them wrote.

**Architecture:** two CLI refusals in `flow record dispatch`, one store repair command, two guard
fixes, one prompt paragraph, and gymie-playwright's spec mapper — each fix where its callers all
route through.

**Spec:** `design.md` of this change.

## Global constraints

- Tasks 1–5 run in `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1009-editor-replace-exercise-remove-from-superset`;
  Task 6 runs in `/Users/tweety53/Projects/gymie-playwright-worktrees/kan-1009-editor-replace-exercise-remove-from-superset`
  on branch `spectre/kan-1009-editor-replace-exercise-remove-from-superset`, and its **Files:**
  are relative to that repository (decision `fixer-groups`).
- No new inline suppressions; for Go, `cd stats && gofmt -l .` prints nothing and `go vet ./...`
  exits 0.
- Tests never call the live flowd or its database. Only the repair steps of Tasks 1 and 2 write
  the live store, through `flowd` subcommands, once each.
- Never touch the gymie main checkout; its stray `full-app-suite.zip` is the operator's to remove.

## Review Focus

- A key that is legitimate and merely ends in a digit or letter (`verify-gymie`,
  `visual-verify-fix-3-2-gymie-playwright`) still records (Task 1).
- `-agent-id inline` and `-agent-id none` keep their meanings on begin, end and backfill (Task 2).
- A range whose ends carry different prefixes (`LM-R1 … X3`) declares nothing extra (Task 4).
- A UI file referenced only from outside the UI tree still selects every spec (Task 6).
- A `//` inside a string such as `https://` does not swallow the rest of the line (Task 6).

- [ ] 1. Refuse a dispatch key ending in a dash and re-key the rows it wrote
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** From `stats/`, run
    `go run ./cmd/flow record dispatches -change kan-1009-editor-replace-exercise-remove-from-superset -C /Users/tweety53/Projects/gymie-worktrees/kan-1009-editor-replace-exercise-remove-from-superset`
    and confirm rows 7290, 7308, 7315, 7321 and 7334 carry the keys `design.md` **Context §1**
    lists.
  - [ ] **Step 2: Write the failing tests.**
    - `stats/cmd/flow/record_test.go`, `TestRecordDispatchBeginRefusesAKeyEndingInADash`:
      `dispatch begin` with `-key verify-` and with `-key verify-fix-1-` each exit 2, stderr names
      the key, and the fake daemon receives no request; `-key verify-gymie` exits 0.
    - `stats/internal/store/records_test.go`, `TestRekeyDispatchSetsTheKey`, shaped like
      `TestRetimeDispatchSetsInstantsAndCommit`: a recorded row re-keyed reads back under the new
      key; a re-key onto a key another row of the same run holds returns an error and leaves both
      rows as they were; an unknown id returns `store.ErrDispatchNotFound`.
    Run both; they fail.
  - [ ] **Step 3: Implement.** In `runRecordDispatchBegin` (`stats/cmd/flow/record.go`), beside
    the other key checks, refuse a `-key` with suffix `-`: exit 2 with
    `flow: -key %q ends in '-': a suffix that expanded to nothing (skills/flow/verify-and-handoff.md, verify[-<worktree basename>])`.
    Add `func (s *Store) RekeyDispatch(ctx context.Context, id int64, key string) error` to
    `stats/internal/store/records.go` (an `UPDATE dispatches SET dispatch_key` by id,
    `ErrDispatchNotFound` on no row), and `flowd rekey-dispatch <dispatch-id> <key>` to
    `stats/cmd/flowd/main.go`, wired and documented exactly as `retime-dispatch` is.
    Decisions `key-trailing-dash-refused-at-begin`, `rekey-dispatch-command`.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l . && go vet ./...` clean;
    `go test ./cmd/flow/ -run 'TestRecordDispatchBeginRefusesAKeyEndingInADash|TestRecordDispatchBeginAcceptsDiffBase' -count=1`
    and `go test ./internal/store/ -run 'TestRekeyDispatchSetsTheKey|TestRetimeDispatchSetsInstantsAndCommit' -count=1`
    pass. Commit the five files with the subject below.
  - [ ] **Step 5: Repair the store.** Build the binary once and re-key each row, each pair
    `<dispatch id>:<key>` from `design.md` **Context §1**:

```bash verified:row ids and target keys read from flow record dispatches and the parent transcript's Bash calls, 2026-10-23
cd /Users/tweety53/Projects/agents-worktrees/self-review-kan-1009-editor-replace-exercise-remove-from-superset/stats && go build -o /private/tmp/flowd-rekey ./cmd/flowd
for pair in 7290:verify-gymie 7308:verify-fix-1-2-gymie-frontend 7315:verify-fix-2-gymie-frontend \
  7321:verify-fix-3-gymie-frontend 7334:verify-fix-4-gymie; do
  /private/tmp/flowd-rekey rekey-dispatch "${pair%%:*}" "${pair#*:}" || echo "FAILED ${pair%%:*}"
done
```

    Then the Step 1 read shows each id under its new key and no key of the change ends in `-`.
    Report any `FAILED` line with its output.

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `stats/cmd/flowd/main.go`, `stats/internal/store/records.go`, `stats/internal/store/records_test.go`
**Tests:** `TestRecordDispatchBeginRefusesAKeyEndingInADash`, `TestRekeyDispatchSetsTheKey`
**Regression:** TestRecordDispatchBeginRefusesAKeyEndingInADash — without the refusal `-key verify-` records a row and exits 0; TestRekeyDispatchSetsTheKey — fails to compile without `RekeyDispatch`.
**Baseline:** before=227 after=229
<!-- measured: cat stats/cmd/flow/record_test.go stats/internal/store/records_test.go | grep -c '^func Test' -->
**Commit:** fix(flow): refuse a dispatch key ending in a dash
**After:** none

**Decision:** key-trailing-dash-refused-at-begin
**Decision:** rekey-dispatch-command

- [ ] 2. Refuse an invented agent id and restamp the rows it left unmeasured
  - [ ] **Step 1: Trace before touching anything; do not guess.** In the parent transcript
    `/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/9fe58c09-fded-4f35-bacc-6270f79bd1a9.jsonl`,
    confirm the `flow record dispatch begin` Bash calls from `2026-10-23T11:20:06Z` on carry
    `-agent-id` labels (`task-11`, `task-11-rev`, `panel-2-extra`, `visual-verify` …), and that
    the store rows of `design.md` **Context §4**'s table carry those labels with empty `metrics`.
    A row that does not fit the trace is reported, not restamped.
  - [ ] **Step 2: Write the failing test.** `stats/cmd/flow/record_test.go`,
    `TestRecordDispatchRefusesAnInventedAgentID`: `dispatch begin` with `-agent-id task-11`, and
    `dispatch end` with `-agent-id visual-verify`, each exit 2 with stderr naming the value and
    telling the caller to omit the flag, and the fake daemon receives no request; begin with
    `-agent-id a30f48a5e14b11392`, with `-agent-id agent_541a9582-890b-417e-a3c3-419047bb4ee6`, with
    `-agent-id inline` and with `-agent-id none` each exit 0. Change the two
    `agent-example0001` values of `TestRecordDispatchSendsAgentIDOnlyWhenGiven` and
    `TestDispatchEndAcceptsAgentID` to `a0123456789abcdef`. Run it; it fails.
  - [ ] **Step 3: Implement.** In `stats/cmd/flow/record.go`, one
    `func validAgentIDFlag(id string) bool` — `""`, `inline`, `none`, `^a[0-9a-f]{16}$`,
    `^agent_[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$` — called by
    `runRecordDispatchBegin` and `runRecordDispatchEnd` before any request, exit 2 with
    `flow: -agent-id %q is not a harness id -- omit -agent-id and the daemon stamps the launch (skills/flow/implement.md)`.
    Decision `agent-id-shape-at-cli`.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l . && go vet ./...` clean;
    `go test ./cmd/flow/ -run 'TestRecordDispatchRefusesAnInventedAgentID|TestRecordDispatchSendsAgentIDOnlyWhenGiven|TestDispatchEndAcceptsAgentID' -count=1`
    passes. Commit the two files with the subject below.
  - [ ] **Step 5: Repair the store.** Build the binary once; restamp each subagent row onto its
    agent transcript, each pair `<dispatch id>:<agent id>` from `design.md` **Context §4**, then
    each inline row from the parent transcript:

```bash verified:every agent file below exists under the session's subagents directory and its meta description matches the row's key (read 2026-10-23)
cd /Users/tweety53/Projects/agents-worktrees/self-review-kan-1009-editor-replace-exercise-remove-from-superset/stats && go build -o /private/tmp/flowd-restamp ./cmd/flowd
S=/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/9fe58c09-fded-4f35-bacc-6270f79bd1a9
for pair in 7294:a30f48a5e14b11392 7295:a721416b5c2a6fcf4 7297:ac48ed3052e52ccec \
  7298:a343f5bc00248dd18 7309:ae9f09c3949fc0d17 7310:ae00ec334d58fb969 \
  7311:af31af1511bec0a75 7312:aec2404d48dd3b55c 7313:a9f4ea45f2739a94a \
  7314:a395d3563d7f75f48 7316:af12aad0d787e4961 7317:aa4821d3bcd77b88a \
  7318:a14c5b9ed9017da4f 7319:a6c0c76621b4fb31c 7320:a594930fdc3fb33eb \
  7322:aa572b9ce0042dff8 7323:aef6a703059daa4e2 7324:a8dff3d66a557a7a3 \
  7331:ac4f440c2eb1cf217 7332:af4ae6359a567c2c2 7333:a165814e413169e5d; do
  /private/tmp/flowd-restamp restamp-dispatch "${pair%%:*}" "$S/subagents/agent-${pair#*:}.jsonl" || echo "FAILED ${pair%%:*}"
done
for id in 7307 7308 7334; do
  /private/tmp/flowd-restamp restamp-dispatch "$id" "$S.jsonl" || echo "NO-SHARE $id"
done
```

    Then `flow record cost-status -change kan-1009-editor-replace-exercise-remove-from-superset`
    (from the gymie worktree) no longer counts the subagent rows. Report any `FAILED` line with
    its output, and each `NO-SHARE` row as left unmeasured (decision `restamp-21-rows`).

**Build:** green
**Files:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`
**Tests:** `TestRecordDispatchRefusesAnInventedAgentID`
**Regression:** TestRecordDispatchRefusesAnInventedAgentID — without the shape check `-agent-id task-11` reaches the daemon and exits 0.
**Baseline:** before=130 after=131
<!-- measured: grep -c '^func Test' stats/cmd/flow/record_test.go -->
**Commit:** fix(flow): refuse an agent id no harness writes
**After:** Task 1

**Decision:** agent-id-shape-at-cli
**Decision:** restamp-21-rows

- [ ] 3. Refuse a relative screenshots root in the full-app-suite zip rebuild
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In a scratch directory
    holding a `.flow/project.md` whose `screenshots` is `shots`, run
    `scripts/rebuild-full-app-suite-zip.sh <that dir> shots` from a different working directory
    and see the zip land under the working directory, not under `<that dir>`.
  - [ ] **Step 2: Write the failing test.** In
    `stats/internal/guard/rebuild_full_app_suite_zip_test.go`, subtest
    `case 7 relative screenshots root`, run through `runGo` alone (the historical bash twin
    accepts it): `[]string{p, "shots"}` exits 2, stderr names `shots`, and no
    `full-app-suite.zip` exists under `p` or the test's working directory. Run it; it fails.
  - [ ] **Step 3: Implement.** In `rebuildFullAppSuiteZip`
    (`stats/internal/guard/rebuildfullappsuitezip.go`), after the usage check, a
    `!filepath.IsAbs(args[1])` exits 2 with
    `rebuild-full-app-suite-zip: <screenshots root> %q is relative — pass it resolved the way resolve-visual-screenshots.sh resolves it`;
    add that exit-2 case to the header of `scripts/rebuild-full-app-suite-zip.sh` and of the Go
    file. Decision `relative-screenshots-root-refused`.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l . && go vet ./...` clean;
    `go test ./internal/guard/ -run 'TestRebuildFullAppSuiteZip' -count=1` passes;
    `scripts/check-guard-symlinks.sh` exits 0. Commit the three files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/rebuildfullappsuitezip.go`, `stats/internal/guard/rebuild_full_app_suite_zip_test.go`, `scripts/rebuild-full-app-suite-zip.sh`
**Tests:** `case 7 relative screenshots root`
**Regression:** case 7 — without the refusal the zip is written under the caller's working directory and the exit is 0.
**Baseline:** before=6 after=7
<!-- measured: grep -c 't.Run(' stats/internal/guard/rebuild_full_app_suite_zip_test.go -->
**Commit:** fix(guard): refuse a relative screenshots root when rebuilding the suite zip
**After:** none

**Decision:** relative-screenshots-root-refused

- [ ] 4. Expand frame ranges and skip rejected alternatives in the frame-list guard
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Run
    `scripts/check-frame-list.sh` on the archived kan-1009 `design.md`
    (`/Users/tweety53/Projects/gymie-worktrees/kan-1009-editor-replace-exercise-remove-from-superset/spectre/changes/archive/kan-1009-editor-replace-exercise-remove-from-superset/design.md`),
    that worktree's `docs/design/mockups` and a report holding only `- frames: 0/0`; see the
    verdict declare `LM-R1`, `LM-R10`, `LM-R3` and `LM-R9` only.
  - [ ] **Step 2: Write the failing test.** `stats/internal/guard/check_frame_list_test.go`,
    `TestCheckFrameListExpandsRangesAndSkipsConsidered`: a mockups dir holding `LM-R1.png` to
    `LM-R4.png`, `R1.png` and `R2.png`; a design reading
    ``Frames `docs/design/mockups/LM-R1.png` … `docs/design/mockups/LM-R4.png`.`` and
    ``**Considered:** bare `R<n>.png` — `R1.png`/`R2.png` already exist.``. A sidecar and report
    covering `LM-R1`–`LM-R4` exit 0 with `4 declared`; the same report missing `LM-R2`'s line
    exits 1 with `no report line: LM-R2`; neither verdict names `R1` or `R2`. A third case,
    `LM-R1 … X3`, declares `LM-R1` alone. Run it; it fails.
  - [ ] **Step 3: Implement** decision `frame-list-ranges-and-considered` in `checkFrameList`
    (`stats/internal/guard/framelist.go`), and state both rules in the THE DECLARED LIST
    paragraph of `scripts/check-frame-list.sh`'s header.
  - [ ] **Step 4: Verify.** `cd stats && gofmt -l . && go vet ./...` clean;
    `go test ./internal/guard/ -run 'TestCheckFrameList' -count=1` passes, and the Step 1 run now
    declares the ten `LM-R` frames and no `R1`/`R2`. Commit the three files with the subject below.

**Build:** green
**Files:** `stats/internal/guard/framelist.go`, `stats/internal/guard/check_frame_list_test.go`, `scripts/check-frame-list.sh`
**Tests:** `TestCheckFrameListExpandsRangesAndSkipsConsidered`
**Regression:** TestCheckFrameListExpandsRangesAndSkipsConsidered — without range expansion `LM-R2` is never declared and the gap case exits 0; without the Considered skip `R1` and `R2` are declared.
**Baseline:** before=1 after=2
<!-- measured: grep -c '^func Test' stats/internal/guard/check_frame_list_test.go -->
**Commit:** fix(guard): expand frame ranges and skip rejected alternatives in the frame list
**After:** none

**Decision:** frame-list-ranges-and-considered

- [ ] 5. Tell the visual verifier to finish the steps its report guard names
  - [ ] **Step 1: Trace before touching anything; do not guess.** In
    `/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/9fe58c09-fded-4f35-bacc-6270f79bd1a9/subagents/agent-ae9f09c3949fc0d17.jsonl`,
    confirm the `check-verify-report.sh` call returning `VERIFY-REPORT-INCOMPLETE … 10 not done`
    and the handback that follows it ("left work, not a blocker; it needs a remainder dispatch");
    confirm the dispatch prompt (its first line) carries the REPORT FORMAT paragraph with "costs a
    remainder dispatch".
  - [ ] **Step 2: Rewrite the paragraph.** In `skills/flow/visual-verify.md`, replace the quoted
    REPORT FORMAT paragraph with:

```markdown verified:authored for this change from decision report-format-forbids-early-handback
> **REPORT FORMAT:** Before writing `## Report`, read the header of
> `<agents repo>/scripts/check-verify-report.sh`, and run it on your report before handing back:
> an exit 1 names the steps you still owe — do them in this dispatch and run it again. Hand back
> an incomplete report only for a blocker you cannot cure, named; time, context and a departure
> found are none. Report every departure you see in this one pass.
```

    and add after the quote: `kan-1009: the paragraph said a rejected report "costs a remainder
    dispatch", and a verifier handed back with step 10 undone in those words; its remainder ran
    9m34s.`
    <!-- measured: flow record dispatches row 7310 started 16:37:57 ended 16:47:30 @ store read 2026-10-23 -->
  - [ ] **Step 3: Verify.** `git grep -n -F -e 'costs a remainder dispatch' -- skills` prints
    nothing; `scripts/check-markdown-integrity.sh`, `scripts/check-references.sh`,
    `scripts/check-dispatch-paragraphs.sh` and `scripts/check-verbatim-moves.sh` exit 0. Commit
    with the subject below.

**Build:** green
**Files:** `skills/flow/visual-verify.md`
**Tests:** **none** — a prompt paragraph; the guard that judges the report is unchanged.
**Commit:** docs(flow): tell the visual verifier to finish the steps its report guard names
**After:** none

**Decision:** report-format-forbids-early-handback

- [ ] 6. Map a frontend diff to specs through declarations, not comments or whole files
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In the playwright worktree,
    `node scripts/specs-for-diff.mjs /Users/tweety53/Projects/gymie-frontend-worktrees/kan-1009-editor-replace-exercise-remove-from-superset a6b3e80e | wc -l`
    prints 94.
    <!-- measured: same command @ gymie-playwright 4c62dc9 → 94 -->
  - [ ] **Step 2: Write the failing tests** in `scripts/specs-for-diff.test.mjs`, in its
    existing `specsFor(touched, gated, helpers, ui, others)` style:
    - `a name a caller only mentions in a comment is no call`: the touched file declares
      `HomeRow`; a tagged UI file mentions `HomeRow` only in `// …` and `/** … */`, another calls
      it — only the caller's spec is selected.
    - `a suspend or generic-receiver declaration is found by its own name`: the touched file holds
      `suspend fun loadHome()` and `fun List<Row>.hairline()`; a UI file calling `loadHome()` and
      a UI file holding `List<Int>` with a different tag — only the caller of `loadHome` selects.
    - `a caller is followed only through the declarations that use the callee`: an untagged
      caller file holds `fun HomeSection() { HomeRow() }` and an unrelated `val RING = 2`; a
      tagged file using only `RING` is not selected, a tagged file calling `HomeSection` is.
    - `a UI file no source calls ends its branch with no specs`: the touched file is called only
      by an untagged UI file nothing references — the other branch's specs are selected, not every
      spec.
    Change `a caller branch that ends untagged selects every spec` to pass, as `others`, a non-UI
    source (`shared/src/commonMain/kotlin/com/gymie/App.kt`) calling `HomeSheet()`, so it still
    pins the every-spec fallback. Run `node --test scripts/specs-for-diff.test.mjs`; the new tests
    fail.
  - [ ] **Step 3: Implement** decision `specs-walk-declarations` in `scripts/specs-for-diff.mjs`:
    one `stripComments(text)` applied to every Kotlin text the CLI reads (touched, `ui`,
    `others`); `declared()` with `suspend` and a generic receiver; `reachedBy` following each
    caller through its using top-level declarations only, keyed in `seen` by path plus those
    declarations' names; a caller-less UI file with no reference in `others` ending its branch
    with no specs; `specsFor(touched, gated, helpers = [], ui = [], others = [])`, the CLI
    passing `git -C <frontend> ls-files '*.kt'` minus `UI_FILE` and `INERT` paths as `others`.
  - [ ] **Step 4: Verify.** `node --test scripts/specs-for-diff.test.mjs` passes, and the Step 1
    command prints fewer than 94 lines — 51 on the traced copy — and still lists
    <!-- measured: traced copy of specs-for-diff.mjs with all four rules, same diff @ gymie-playwright 4c62dc9 → 51 -->
    `fidelity/list-mode-editor.spec.ts` and `fidelity/kan-847-exercise-info.spec.ts`, the run's
    capture specs. Commit the two files
    on `spectre/kan-1009-editor-replace-exercise-remove-from-superset` with the subject below;
    never push.

**Build:** green
**Files:** `scripts/specs-for-diff.mjs`, `scripts/specs-for-diff.test.mjs`
**Tests:** `a name a caller only mentions in a comment is no call`, `a suspend or generic-receiver declaration is found by its own name`, `a caller is followed only through the declarations that use the callee`, `a UI file no source calls ends its branch with no specs`
**Regression:** each new test selects every spec, or the wrong spec, with its rule reverted; the kan-1009 diff maps back to 94.
**Baseline:** before=16 after=20
<!-- measured: grep -c '^test(' scripts/specs-for-diff.test.mjs -->
**Commit:** fix(scripts): map a frontend diff to specs through declarations, not comments or whole files
**After:** none

**Decision:** specs-walk-declarations
