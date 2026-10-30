# self-review-kan-1069-interactive-session

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode
> **Baseline convention:** every `**Baseline:** before=N after=M` is chain-position absolutes —
> `before` is the count at this task's commit's parent and `after` at the commit, each task
> measured at its own position, never a plan-wide pair — split by source set, counted by a
> word-boundary-safe counter (`@Test` never matches `@Testcontainers`), with any `measured:`
> command naming no ref of its own.

**Goal:** close the five non-`big` findings of `kan-1069-interactive-session`'s self-review pass.

**Architecture:** two more checks in `gateLandedRequest`; an array guard in `DecisionsPanel`; one
error line on the state board; gated reviewer findings recorded under slot `gated-reviewer`, which
every panel-scoped reader skips through one constant.

**Spec:** `design.md` beside this file; the brainstorm design is
`.superpowers/sdd/2026-10-30-self-review-kan-1069-interactive-session-design.md` in this worktree.

## Global constraints

- Self-review fixes: one commit per task with its `**Commit:**` subject and no `Task-Id:` trailer. Never write `spectre/changes/` — this change's `verbatim-moves.txt` is already committed with the plan. Never merge or push.
- Work in the worktree `/Users/tweety53/Projects/agents-worktrees/kan-1069-interactive-session`, branch `kan-1069-interactive-session`; every `**Files:**` path is relative to its root. Never touch a main checkout, and never create a worktree.
- Every prose edit lands character for character through the script below, saved to a scratch file outside the worktree and run as `python3 -I <saved script> "$PWD" <task number>` from the worktree's root; it asserts each anchor matches exactly once. `scripts/check-verbatim-moves.sh` matches the result against `verbatim-moves.txt` exactly — a hand edit that drifts by one character fails it. A run that asserts is reported, never hand-merged.
- Run `gofmt -w` on every Go file you change before committing; `gofmt -l` and `go vet` stay clean. Never start, stop or restart the protected `flowd` on `127.0.0.1:4173`, nor the `flow-postgres` container.

```python verified:applied with all three task numbers to this worktree at 3deff9fa; scripts/check-verbatim-moves.sh reported 0 violation(s) with this change's verbatim-moves.txt, then reverted
import sys
root = sys.argv[1]
def sub(path, old, new):
    p = f"{root}/{path}"
    s = open(p).read()
    assert s.count(old) == 1, (path, old[:60], s.count(old))
    open(p, "w").write(s.replace(old, new))
which = sys.argv[2:]
if "1" in which:
    sub("skills/flow-interactive/SKILL.md",
        "   `-base` an ancestor of `-commit`.\n",
        "   `-base` an ancestor of `-commit`. The request's latest `/flow-interactive` `flow.verify` stage\n   run carrying `-request <n>`, by start, must have ended `completed`, whatever the diff touches.\n")
    sub("skills/flow-interactive/SKILL.md",
        "the stage-run read is skipped\n   with one `⚠` line,",
        "the stage-run reads are skipped\n   with one `⚠` line,")
    sub("skills/flow-interactive/SKILL.md",
        "land the commit, run the UI check and end it `completed` — and the end re-run",
        "land the commit, run the verify and the UI check and end each `completed` — and the end re-run")
if "2" in which:
    sub("skills/flow-interactive/SKILL.md",
        "   `VISUAL-TRIGGER-CANNOT-ANSWER` refuses, quoting its cause.\n",
        "   `VISUAL-TRIGGER-CANNOT-ANSWER` refuses, quoting its cause. A `completed` UI check that ended\n   less than one second after it began is a chained mark (step 7) and is refused.\n")
if "5" in which:
    sub("skills/flow/implement.md",
        "its report file's `## Verdict`, so per-task review yield stays measurable against the gate.\n",
        "its report file's `## Verdict`, so per-task review yield stays measurable against the gate.\n"
        "The parent records every finding a pass raises as a findings row in one Bash call, `flow record\n"
        "finding -change <name> -from <abs-worktree>/.superpowers/sdd/findings-gated-<n+n+n>.jsonl`, each\n"
        "row under `-slot gated-reviewer -round 0 -task <n> -dispatch-seq <the bundle's seq> -status open\n"
        "-reproducer 'none — gated per-task review; its report file holds the evidence'`, its ref from\n"
        "`flow record next-ref`, so the run's Findings count holds what the gate caught; the panel's\n"
        "guards and its rendered record skip `gated-reviewer` rows.\n")
    sub("skills/flow/implement.md",
        "pushed, no dispatch, no dispatch record and no `flow record status` — and only while no",
        "pushed, no dispatch and no dispatch record, then `flow record status -ref F<n> -status fixed` for\neach — and only while no")
    sub("skills/flow/implement.md",
        "A Minor no change to the tree can resolve is named,\none clause, in the parent's output.",
        "A Minor no change to the tree can resolve is named,\none clause, in the parent's output, and recorded `withdrawn <that clause>`.")
    sub("skills/flow/gated-review-fix.md",
        "The fix the parent applies is mutation-proved",
        "A clean re-review pass closes each finding its task's earlier pass raised, `flow record status\n-change <name> -ref F<n> -status fixed`; a finding a re-review pass raises is recorded as **The\ngated per-task reviewer** (`skills/flow/implement.md`) records a first pass's.\n\nThe fix the parent applies is mutation-proved")
```

## Review Focus

- Task 1: the verify read runs on every `landed` end, the non-UI range included, and the store-down path still prints exactly one `⚠` line.
- Task 2: the chained check fires on a sub-second `completed` run only; a run with no `endedAt` is not passed as honest.
- Task 5: a gated row never trips a panel guard, yet an open gated row still stops `check-unfinished-work`; `next-round` still prints 0 for a change holding only gated rows.

- [ ] 1. Require a completed verify on every landed request
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** In `stats/cmd/flow/request_test.go`, make `requestDaemon`'s GET answer filter its `runs` to the rows whose `"stage"` equals the request's `stage` query parameter, then add `TestRecordRequestEndLandedNeedsVerify`: a non-UI range (`end(docs, base)`) with no `flow.verify` run for request 3 is refused naming `flow.verify` with no write; with a `failed` latest verify run it is refused; with a `completed` latest verify run (and, for a UI range, a completed UI check) it writes once. Run `cd stats && make -s web-dist && go test ./cmd/flow/ -run 'TestRecordRequestEndLandedNeedsVerify' -count=1`: it fails on the untouched `request.go` (exit 0 on the non-UI range).
  - [ ] **Step 2: Fix `gateLandedRequest`** (`stats/cmd/flow/request.go`): after the `git diff --name-only` read and before the `check-visual-trigger` switch, read `flow.verify` runs with the same `callRecord` / `ListStageRuns` call and pick the request's latest by start — factor the existing latest-run loop into one helper both reads use. Refuse `request %d has no verify (flow.verify run carrying -request %d)` when none, and `request %d's latest verify (flow.verify) is %s, not completed` when not `completed`. With the store unavailable (`client.ErrUnavailable` without `ErrDaemonAnswered`) print one line, `⚠ flow: store unreachable — skipped request %d's verify and UI check reads; the git checks passed`, and return 0 without the UI read. Update the doc comment. Update `TestRecordRequestEndLandedGuard`'s fixtures so each passing case also carries a completed `flow.verify` row for request 3, and its "a range touching no ui path" case expects one GET (the verify read) instead of none.
  - [ ] **Step 3: Prose** — run the Global constraints script with task number `1`.
  - [ ] **Step 4: Verify** — `gofmt -w` the Go files; `cd stats && make -s web-dist && go test ./cmd/flow/ -run 'TestRecordRequestEnd' -count=1` passes; `scripts/run-verify-list.sh /Users/tweety53/Projects/agents-worktrees/kan-1069-interactive-session lint` is green.

**Build:** green
**Files:** `stats/cmd/flow/request.go`, `stats/cmd/flow/request_test.go`, `skills/flow-interactive/SKILL.md`
**Tests:** `TestRecordRequestEndLandedNeedsVerify`
**Regression:** `TestRecordRequestEndLandedNeedsVerify` — reverting the commit lets a non-UI landed end with no `flow.verify` run exit 0 and write.
**Baseline:** before=2 after=3
<!-- measured: grep -cE '^func Test' stats/cmd/flow/request_test.go @ 3deff9fa (2 before this task) -->
**Commit:** fix(stats): require a completed verify before a request ends landed
**After:** none

**Decision:** verify-required-on-landed
**Decision:** no-row-repair-for-landed-requests

- [ ] 2. Refuse a UI check whose end was chained to its begin
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Add `TestRecordRequestEndRefusesChainedUICheck` to `stats/cmd/flow/request_test.go`: a UI range whose request carries a completed verify run and a latest `flow.visual-verify` row `completed` with `"endedAt"` 31 ms after its `"startedAt"` is refused naming the chained mark, with no write; the same row ending 11 s after its start writes once. Run `cd stats && make -s web-dist && go test ./cmd/flow/ -run 'TestRecordRequestEndRefusesChainedUICheck' -count=1`: it fails before the fix (exit 0 on the 31 ms row).
    <!-- predicted: the fixture values the test writes, 31 ms and 11 s, mirror the measured chained and shortest honest runs in design.md's Context -->
  - [ ] **Step 2: Fix** — add `EndedAt *string \`json:"endedAt,omitempty"\`` to `client.StageRunRow` (`stats/internal/client/client.go`). In `gateLandedRequest`, after the UI check's `completed` test, parse `EndedAt` (RFC3339Nano) and refuse `request %d's latest UI check (flow.visual-verify) ended %s after its begin — a chained mark (step 7); end it after the screenshot is read` when it is under one second after `StartedAt`; an unreadable `endedAt` is refused like an unreadable `startedAt`. Give the existing UI-check fixtures' completed rows an `endedAt` at least one second after their start.
  - [ ] **Step 3: Prose** — run the Global constraints script with task number `2`.
  - [ ] **Step 4: Verify** — `gofmt -w` the Go files; `cd stats && make -s web-dist && go test ./cmd/flow/ ./internal/client/ -run 'TestRecordRequestEnd|TestListStageRuns' -count=1` passes; `scripts/run-verify-list.sh /Users/tweety53/Projects/agents-worktrees/kan-1069-interactive-session lint` is green.

**Build:** green
**Files:** `stats/cmd/flow/request.go`, `stats/cmd/flow/request_test.go`, `stats/internal/client/client.go`, `skills/flow-interactive/SKILL.md`
**Tests:** `TestRecordRequestEndRefusesChainedUICheck`
**Regression:** `TestRecordRequestEndRefusesChainedUICheck` — reverting the commit lets a `completed` UI check ended 31 ms after its begin pass the guard.
**Baseline:** before=3 after=4
<!-- measured: grep -cE '^func Test' stats/cmd/flow/request_test.go @ 3deff9fa (2, plus Task 1's one) -->
**Commit:** fix(stats): refuse a UI check whose end was chained to its begin
**After:** Task 1

**Decision:** chained-mark-under-one-second

- [ ] 3. Keep the run page up when a decision's overrides is an object
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Add to `stats/web/src/components/DecisionsPanel.test.tsx` the case `it("ignores an overrides object", …)` rendering `<DecisionsPanel decision={{ implementer: { model: "opus", effort: "high", reason: "seam" }, overrides: {} }} />` and expecting the rows `["Implementeropus · high"]`. Run `cd stats && make -s web-dist && cd web && npx vitest run src/components/DecisionsPanel.test.tsx -t 'ignores an overrides object'`: it fails with `flatMap is not a function`.
  - [ ] **Step 2: Fix** — in `stats/web/src/components/DecisionsPanel.tsx`, type `overrides` as `unknown` on `Decision` and read it as `(Array.isArray(d.overrides) ? (d.overrides as Override[]) : [])`, `Override` the existing element type named once.
  - [ ] **Step 3: Verify** — the vitest command above passes; `cd stats && make -s web-dist && cd web && npx tsc -b` is clean.

**Build:** green
**Files:** `stats/web/src/components/DecisionsPanel.tsx`, `stats/web/src/components/DecisionsPanel.test.tsx`
**Tests:** `ignores an overrides object`
**Regression:** `ignores an overrides object` — reverting the commit makes `lines()` throw on `overrides: {}` and the render fails.
**Baseline:** before=2 after=3
<!-- measured: grep -cE '^ *it\(' stats/web/src/components/DecisionsPanel.test.tsx @ 3deff9fa (2 before this task) -->
**Commit:** fix(stats): keep the run page up when a decision's overrides is an object
**After:** none

**Decision:** overrides-array-only

- [ ] 4. Show a failed board read once the board has rows
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Add to `stats/web/src/views/views.test.tsx`, beside the other state-board cases, `it("state board shows a failed finished read", …)`: the mock resolves the open poll (`from === to`) with `fixtures["state-board"]` and rejects every other read with `new Error("boom")`; render `<StateBoard />`, open the Finished group, and expect `Could not load the board: boom` to be shown. Run `cd stats && make -s web-dist && cd web && npx vitest run src/views/views.test.tsx -t 'state board shows a failed finished read'`: it fails, no such text.
  - [ ] **Step 2: Fix** — in `stats/web/src/views/StateBoard.tsx`, after `{heading}` in the rendered-rows return, render `{fetched.status === "error" && <p className="lb-sub" role="alert">Could not load the board: {fetched.message}</p>}`, so whichever read errored — the open poll first, as `fetched` already prefers it — is named.
  - [ ] **Step 3: Verify** — the vitest command above passes; `cd stats && make -s web-dist && cd web && npx tsc -b` is clean.

**Build:** green
**Files:** `stats/web/src/views/StateBoard.tsx`, `stats/web/src/views/views.test.tsx`
**Tests:** `state board shows a failed finished read`
**Regression:** `state board shows a failed finished read` — reverting the commit leaves a failed finished read with no message once the open poll answered.
**Baseline:** before=55 after=56
<!-- measured: grep -cE '^ *it\(' stats/web/src/views/views.test.tsx @ 3deff9fa (55 before this task) -->
**Commit:** fix(stats): show a failed state-board read once the board has rows
**After:** none

**Decision:** board-error-beside-heading

- [ ] 5. Record gated reviewer findings and keep the panel's guards off them
  - [ ] **Step 1: Reproduce before touching anything; do not guess.** Add `TestCheckPanelFindingsClosedSkipsGatedRows` to `stats/internal/guard/check_panel_findings_closed_test.go`: through `Env.Findings` / `Env.Dispatches`, a change whose only findings are a `fixed` Important row and a `fixed` Critical row under slot `gated-reviewer`, round 0, reproducer `none — gated per-task review; its report file holds the evidence`, and whose only dispatch is a `task-1-reviewer` row closed `clean`, exits 0 `FINDINGS-CLOSED`. Run `cd stats && make -s web-dist && go test ./internal/guard/ -run 'TestCheckPanelFindingsClosedSkipsGatedRows' -count=1`: it fails, the guard refusing the two rows as fixed with no clean re-run.
  - [ ] **Step 2: The constant** — add `const GatedReviewerSlot = "gated-reviewer"` with a one-paragraph doc comment (the per-task gated reviewer's findings slot; panel-scoped readers skip it) to `stats/internal/records/` beside the findings types.
  - [ ] **Step 3: Skip gated rows** in each panel-scoped reader, comparing the row's slot to `records.GatedReviewerSlot`: `checkPanelFindingsClosed` drops them right after `cfcParse`, so every check including `cfcUncounted` and the open check sees none (`stats/internal/guard/panelfindingsclosed.go`, and a sentence in `scripts/check-panel-findings-closed.sh`'s header comment); `uwUnprovedFixRounds` ignores them (`stats/internal/guard/unfinishedwork.go`), while `check-unfinished-work`'s open-finding check still counts them; `check-panel-reproducers` skips them (`stats/internal/guard/panelreproducers.go`); `flow record render -kind panel` leaves them out of its findings table and both totals (`stats/internal/records/render.go`); `highestRecordedRound` and `runRecordNextRound` skip them (`stats/cmd/flow/record.go`).
  - [ ] **Step 4: Tests for each reader** — `TestUnfinishedWorkIgnoresGatedRowsForFixRoundProof` (`check_unfinished_work_test.go`: a `fixed` Important gated row with no mutation row is no unproved round, while an `open` gated row is still reported open), `TestPanelReproducersSkipsGatedRows` (`check_panel_reproducers_test.go`: an Important gated row with the `none — …` reproducer is not refused), `TestRenderPanelOmitsGatedRows` (`render_test.go`: a panel row and a gated row render `findings-total: 1`), `TestRecordNextRoundSkipsGatedRows` (`record_test.go`: a change holding only a round-0 gated row prints `0`). Each fails with Step 3's hunk for its reader reverted.
  - [ ] **Step 5: Prose** — run the Global constraints script with task number `5`.
  - [ ] **Step 6: The KPI** — remove `hint="review panel only"` from the Findings `<Kpi>` in `stats/web/src/views/RunDetail.tsx`.
  - [ ] **Step 7: Verify** — `gofmt -w` the Go files; `cd stats && make -s web-dist && go test ./internal/guard/ ./internal/records/ ./cmd/flow/ -run 'TestCheckPanelFindingsClosed|TestUnfinishedWork|TestPanelReproducers|TestCheckPanelReproducers|TestRenderPanel|TestRecordNextRound' -count=1` passes; `cd stats && make -s web-dist && cd web && npx tsc -b` is clean; `scripts/run-verify-list.sh /Users/tweety53/Projects/agents-worktrees/kan-1069-interactive-session lint` is green.

**Build:** green
**Files:** `skills/flow/implement.md`, `skills/flow/gated-review-fix.md`, `scripts/check-panel-findings-closed.sh`, `stats/internal/records/render.go`, `stats/internal/records/render_test.go`, `stats/internal/guard/panelfindingsclosed.go`, `stats/internal/guard/check_panel_findings_closed_test.go`, `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go`, `stats/internal/guard/panelreproducers.go`, `stats/internal/guard/check_panel_reproducers_test.go`, `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `stats/web/src/views/RunDetail.tsx`
**Allowed-collateral:** `stats/internal/records/*.go`
**Tests:** `TestCheckPanelFindingsClosedSkipsGatedRows`, `TestUnfinishedWorkIgnoresGatedRowsForFixRoundProof`, `TestPanelReproducersSkipsGatedRows`, `TestRenderPanelOmitsGatedRows`, `TestRecordNextRoundSkipsGatedRows`
**Regression:** `TestCheckPanelFindingsClosedSkipsGatedRows` — reverting the commit refuses fixed gated Criticals and Importants with no panel re-run; `TestUnfinishedWorkIgnoresGatedRowsForFixRoundProof` — reverting reports round 0 unproved; `TestPanelReproducersSkipsGatedRows` — reverting refuses the Important's `none —` exemption; `TestRenderPanelOmitsGatedRows` — reverting renders `findings-total: 2`; `TestRecordNextRoundSkipsGatedRows` — reverting prints `1`.
**Baseline:** before=7 after=8 (`check_panel_findings_closed_test.go`); before=15 after=16 (`check_unfinished_work_test.go`); before=3 after=4 (`check_panel_reproducers_test.go`); before=40 after=41 (`render_test.go`); before=163 after=164 (`record_test.go`)
<!-- measured: grep -cE '^func Test' stats/internal/guard/check_panel_findings_closed_test.go @ 3deff9fa (7; the other four files' counts by the same grep at the same ref) -->
**Commit:** fix(stats): record gated reviewer findings and keep the panel's guards off them
**After:** none

**Decision:** gated-findings-recorded
**Decision:** gated-rows-skipped-by-panel-guards
**Decision:** no-gated-history-backfill
**Decision:** findings-kpi-no-hint
