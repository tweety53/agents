## Context

Self-review fix change for `kan-1046-review-dynamics-dashboard`, planned on that change's own
branch `spectre/kan-1046-review-dynamics-dashboard` at agents `6c2ee105` (the fix-base). Nine
non-`big` findings, as `proposal.md` names them, each carried by one task in `tasks.md`: 1a
(Task 1), 1b (Task 2), 1c (Task 3), 1d (Task 4), 1e (Task 5), 1f (Task 6), 4a (Task 7), 8c
(Task 8), 8d (Task 9). Every finding was traced before its fix was designed; 8d's noise was
reproduced and pinned by a live spike at plan time. Every question was auto-resolved to its
recommended option (**Auto-resolution**, `skills/flow-contracts/operator-prompts-auto-resolution.md`).
The brainstorm design, with the traces and the spike, is
`.superpowers/sdd/2026-10-28-self-review-kan-1046-review-dynamics-dashboard-design.md` in this
worktree.

A fix repairs what its defect already wrote: the store rows 8371, 8415 and 8419 and the
panel-mutation rows 1076, 1077 and 1079 of the dev `flow` database are repaired through `flowd`
subcommands in the task that fixes their cause. `flowd` on 4173, the `flow-postgres` container and
the `flow` database are never stopped or dropped.

## Decisions

### A re-end of a closed dispatch row keeps its end

**ID:** closed-row-keeps-end
**Status:** active
**Chosen:** `Store.EndDispatch`'s `UPDATE` sets `ended_at = COALESCE(d.ended_at, $7)`, so a second `end` on a closed row re-points its commit and outcome only; moving a closed row's end stays `flowd retime-dispatch`'s job. Rows 8415 and 8419 are retimed to end at their commits' author times, 14:37:26Z and 14:41:35Z, keeping their recorded commits.
**Considered:** refusing a re-end outright — the run's re-point after an autosquash is legitimate and needs the commit update; ending at the original `close-task` instant — nothing recorded it before the overwrite, while the commit times are on record.

### An inline back-fill floors at the change's visual-verify verdicts

**ID:** inline-floor-visual-verify
**Status:** active
**Chosen:** `inlineFloor` (`stats/cmd/flow/record.go`) also takes, as a floor candidate, the end of any row of the change with role `verifier` and a key starting `visual-verify`, under the same not-after-`endT` bound. Row 8415's start is retimed to 14:34:31Z, row 8409's end.
**Considered:** flooring at every non-inline row — a subagent running beside the parent's inline work would wrongly clip it; only a verifier verdict is a precondition of the fix that answers it.

### An inline panel-fix back-fill waits for its round's reviewers to close

**ID:** panel-fix-needs-closed-round
**Status:** active
**Chosen:** in the inline back-fill path of `flow record dispatch backfill`, a `panel-fix-<r>[-<chunk>]` key is refused, exit 2 before any store call, while any reviewer row of panel round `<r>-1` has no `ended_at`; the message reads `close panel round <r-1>'s reviewer rows first`. A failed run-record read refuses nothing. Row 8419's start is retimed to 14:40:55Z, round 3's last reviewer end.
**Considered:** flooring at the open rows' starts — a fix cannot begin before the reports it fixes exist, and an open row has no report yet; `panelFixKeyedARoundLow` already reads starts for the opposite question.

### A mutation proof precedes the re-review it feeds

**ID:** mutation-before-re-review
**Status:** active
**Chosen:** `flow record mutation -round <r>` is refused, exit 2 before the store call, once any reviewer row of panel round `<r>` (`records.PanelDispatchRound`) has a `started_at` not after now; the message names the round and says the proof is recorded before the re-review. A failed run-record read refuses nothing. One sentence beside **Record each one** in `skills/flow/review-panel-fix-round.md` states the order. A new `flowd drop-mutation <id>` over a new `Store.DropMutation`, shaped as `drop-pass`/`DropPass`, deletes rows 1076, 1077 and 1079 — the round-1 rows panel-fix-1 never ran; 1078, the one it did run, stays.
**Considered:** making `check-unfinished-work` compare row times — the gate runs at landing, after the re-review the rows were meant to precede; a SQL `DELETE` by hand — the project repairs rows through its own commands.

### The brainstorm stage cannot end over an open dispatch row

**ID:** brainstorm-end-needs-closed-rows
**Status:** active
**Chosen:** `flow stage end -command '/flow' -stage flow.brainstorm` reads the change's run record and is refused, exit 1, while any of its dispatch rows has no `ended_at`, naming each open key, beside the existing `gate*` checks in `stats/cmd/flow/stage.go`; a failed read refuses nothing. One sentence in **Explore and design agents** (`skills/flow/brainstorm-planner.md`) says each planning row is ended when its agent hands back, before the stage ends. Row 8371 is retimed to end at 12:36:38Z, the agent's last transcript line.
**Considered:** matching on the session token — `stage end` takes no token, and during `flow.brainstorm` the change's only rows are the planning session's; ending every stage this way — only brainstorm hands over to a different session, which is where the open row was lost.

### close-task runs the plan guard on appended tasks

**ID:** close-task-plan-shape
**Status:** active
**Chosen:** `closeTask` (`stats/internal/guard/closetask.go`) runs `Registry["check-plan-shape"]` on `<canonical>/spectre/changes/<name>/tasks.md` whenever a task it closes is appended (`**Tasks appended:**`) or titled `In-run fix <k> — `, before anything is ticked, and refuses on its exit 1. **The loop** step 1 (`skills/flow/verify-fix-loop.md`) adds that the visual record is untracked until that task's commit, which `git add`s it. The `scripts/close-task.sh` header names the new step.
**Considered:** a prose reminder only — `document-fix.md` already states the rule, and the run missed it.

### The live-check paragraph names a declared stand-in tool

**ID:** live-check-names-declared-tool
**Status:** active
**Chosen:** one sentence in **Write a live check** (`skills/flow/brainstorm-planner.md`): when `<project>/.flow/project.md` names a tool for the check's stand-in, the section invokes it by name and never re-specifies it.
**Considered:** nothing narrower states it at plan time.

### A prototype-specified capture measures before it commits

**ID:** prototype-capture-measures
**Status:** active
**Chosen:** **Write a capture task that baselines against a mockup to measure before it commits** (`skills/flow/brainstorm-planner.md`) is extended: an approved prototype, rendered as each capture's frame per step 10 of `skills/flow/visual-verify-verifier.md`, gets the same compose-and-measure steps in the capture task. The missed state captures are planner variance under an existing rule and change nothing.
**Considered:** a new paragraph — the measure rule is one rule, and two copies drift.

### The baseline noise is pinned at its source

**ID:** pin-chromium-raster
**Status:** active
**Chosen:** `stats/web/playwright.config.ts` sets `use.launchOptions.args` to `--disable-gpu-rasterization`, `--disable-partial-raster` and `--disable-gpu`, replaces the comment claiming no rendering pin is needed, and every baseline under `stats/web/tests/visual/*-snapshots/` is recaptured once under the pin, with `full-app-suite.zip` rebuilt by `scripts/rebuild-full-app-suite-zip.sh`. The plan-time spike: three unpinned captures differ pixelwise at the select corners, four pinned captures are byte-identical. `visual-verify-verifier.md`'s `=all` rule and `verify-fix-loop.md:22` are unchanged.
**Considered:** scoping `=all` to the specs the change's diff touches, the pass's proposal — with the noise pinned, `=all` rewrites an unrelated baseline byte-identically, so the scope adds a rule with no effect, and it would hide the next nondeterminism instead of surfacing it (`rules/fix-determinism-at-the-source.mdc`); a `maxDiffPixels` or threshold — a widened tolerance.

### The noise stash is the parent's

**ID:** stash-drop-by-parent
**Status:** active
**Chosen:** the parent session drops the stash holding the noise, `36807042` "aa-noise full-app-suite recapture" (`stash@{0}` at plan time), confirming the sha by `git stash list --format='%h %gs'` before `git stash drop`; no task does.
**Considered:** a task step — fixers never touch shared refs outside their commits.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Tasks 3, 4 and 5 add refusals to commands every `/flow` run calls, Task 6 changes what `close-task` refuses, and Tasks 1–5 write the live store.
**Considered:** `flow-low` — the diffs are small, but a refusal wrong by one condition stops every run's panel-fix back-fill, mutation record or brainstorm end.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** three groups in series: `-fix-g1` Tasks 1–6, the Go store, CLI and guard work with its store repairs (Tasks 2–4 share `stats/cmd/flow/record.go`; Tasks 1 and 4 share `stats/internal/store/records.go`); `-fix-g2` Tasks 7–8, the two `brainstorm-planner.md` prose edits, after Task 5's sentence in the same file; `-fix-g3` Task 9, the Playwright pin and recapture against the UI-test stack.
**Considered:** one fixer for all nine — one context carrying Go, prose and a full visual recapture; one per task — nine dispatches, five sharing files.

## Open questions

## Live check

Each step runs before landing, with every task committed, in the worktree. Store reads use
`docker exec flow-postgres psql -U flow -d flow -tAc "<query>"`. Refusal checks run against a copy:
`scripts/live-store-copy.sh up k1046sr` from the worktree, its printed `FLOW_ADDR`/
`FLOW_RECORDS_ADDR` exported, then `cd stats && go run ./cmd/flow …`; `scripts/live-store-copy.sh
down k1046sr` at the end.

1. **Repaired rows.** `select id, started_at, ended_at from dispatches where id in (8371,8415,8419)`
   — before: 8371 ends 12:57:35, 8415 runs 14:26:41–14:54:33.63, 8419 runs 14:37:39–14:54:33.75;
   after: 8371 ends 12:36:38, 8415 runs 14:34:31–14:37:26, 8419 runs 14:40:55–14:41:35. Not
   working looks like any before value.
2. **Mutation rows.** `select id from panel_mutations where id in (1076,1077,1078,1079)` — before
   four ids, after `1078` alone.
3. **Re-end keeps the end (copy).** `go run ./cmd/flow record dispatch end -C <worktree> -change
   kan-1046-review-dynamics-dashboard -key panel-fix-4 -session-token mf-k1046-20261028T125815Z
   -commit 0acfc59094f51a01bd1b59232f6cee0a3d4eb9e4 -outcome completed`, then read row 8419's
   `ended_at` from `flow_k1046sr_live` — still 14:41:35. Not working looks like now.
4. **Mutation refusal (copy).** `go run ./cmd/flow record mutation -C <worktree> -change
   kan-1046-review-dynamics-dashboard -round 1 -path stats/cmd/flow/prompttuning.go -mutated x
   -test y` — exit 2 naming panel round 1; `panel_mutations` count in the copy unchanged.
5. **Plan guard at close (scratch).** The Task 6 test is the check; `scripts/check-plan-shape.sh`
   on this change's `tasks.md` exits 0.
6. **Pin holds.** Two `full-app-suite.spec.ts` captures through a scratch config pointing
   `snapshotPathTemplate` outside the tree, otherwise the committed config, come out
   byte-identical (`cmp`); `npm run test:visual` against the restarted UI-test stack passes.
   Not working looks like a `cmp` difference.

## Recommended approach

Fix each defect where it starts: the store keeps a closed row's end; the back-fill floors read the
verifier and open reviewer rows the run record already holds; the mutation record and the
brainstorm end read the same record before writing; `close-task` calls the plan guard it already
links; two planner paragraphs gain one sentence each; the SPA's Playwright config pins the
rasteriser. Store rows the defects wrote are repaired by `flowd` in the same task.

## Critical files

- `stats/internal/store/records.go`, `records_test.go` — `EndDispatch`, new `DropMutation`.
- `stats/cmd/flow/record.go`, `record_test.go` — `inlineFloor`, the inline back-fill, `runRecordMutation`.
- `stats/cmd/flowd/main.go` — `drop-mutation`.
- `stats/cmd/flow/stage.go`, `stage_test.go` — the brainstorm-end gate.
- `stats/internal/guard/closetask.go`, `close_task_test.go`, `scripts/close-task.sh`.
- `skills/flow/review-panel-fix-round.md`, `skills/flow/brainstorm-planner.md`, `skills/flow/verify-fix-loop.md`.
- `stats/web/playwright.config.ts`, `stats/web/tests/visual/*-snapshots/`, `stats/web/tests/visual/full-app-suite.zip`.

## Existing code to reuse

- `RetimeDispatch` / `flowd retime-dispatch`; `DropPass` / `flowd drop-pass` as the shape for `drop-mutation`.
- `records.PanelDispatchRound`, `callRecord` + `GetRunRecord` as `panelFixFloor` uses them.
- `gateLiveVerification` and its siblings in `stage.go` as the gate shape.
- `ctUndispatched`'s appended/`In-run fix` detection (`uwAppended`) in `closetask.go`; `Registry["check-plan-shape"]`.
- `scripts/rebuild-full-app-suite-zip.sh`; the `capture` row of `.flow/project.md` `## visual verification`.

## Verification

Beyond each task's verify step: `scripts/run-guard-tests.sh` and `cd stats && make -s web-dist &&
go test -race -count=1 $(go list ./... | grep -v '/internal/guard$')` pass; `cd stats && make -s
web-dist && cd web && npm test` and `npm run test:visual` pass; the agents `## lint` list passes;
the **Live check** above.
