## Context

Self-review fix change for `kan-1045-flow-one-allowed-model-effort-list-per-subtask`, planned on that
change's own branch `spectre/kan-1045-flow-one-allowed-model-effort-list-per-subtask` at `8ebfe849`,
in `/Users/tweety53/Projects/agents-worktrees/kan-1045-flow-one-allowed-model-effort-list-per-subtask`.
Six non-`big` findings, as `proposal.md` names them, each carried by one task in `tasks.md`: F-A
(Task 1), F-B (Task 2), F-C (Task 3), F-D (Task 4), F-E (Task 5), F-F (Task 6). Each was traced on
that HEAD before its fix was designed; every question was auto-resolved to its recommended option
(**Auto-resolution**, `skills/flow-contracts/operator-prompts-auto-resolution.md`). The evidence is
the pass's `.superpowers/sdd/self-review-pass.md` in this worktree; the traces are in
`.superpowers/sdd/2026-10-29-self-review-kan-1045-flow-one-allowed-model-effort-list-per-subtask-design.md`.
No archived change records a rejection of any remedy planned here.

`flowd` on 4173, the `flow-postgres` container and the `flow` database are never stopped, restarted
or dropped; the live check below runs on a scratch copy.

## Decisions

### Narrowing a Files field is a `remove` verb

**ID:** tasks-files-remove
**Status:** active
**Chosen:** `flow tasks files [-C dir] <change> <task-id> remove <path>...` drops each backticked
`<path>` entry from that one task's `**Files:**` field, rejoining the rest with `, `; a field left
empty becomes `none`; a `<path>` the field does not name is refused, exit 1, the plan left unchanged
— a typo then fails loudly instead of reporting a narrowing that did not happen. `add` and `remove`
share one locate loop in `stats/cmd/flow/tick.go`, now that there are two callers.
**Considered:** a `set` verb replacing the whole field — rewrites what no refusal named, so the
transcription stops being the disclosed correction; a silent no-op for an absent path, as `add`
treats a present one — a mistyped narrowing would print success and leave the guard red.

### The findings merger is enforced at the begin that follows it

**ID:** merger-enforced-at-begin
**Status:** active
**Chosen:** the recommended option, mechanical enforcement; the rule itself stands — a slot's own
handback is not the merged summary with its `duplicate:` and `missing-reproducer:` lines, and the
merger is what keeps every report out of the parent's context. `flow record dispatch begin`
(`stats/cmd/flow/record.go`) refuses, exit 2, (a) a `panel-fix-<r>[-<n>][-retry]` key, live **or**
back-filled (`-started-at`), and (b) a live reviewer row of panel round `<r>` with `r ≥ 1`
(`records.PanelDispatchRound`), while the run record holds no dispatch keyed `panel-<r-1>-merger`
with role `merger`, an end and outcome `completed`. The message names the missing key and
`skills/flow/review-panel.md`. Check (a) runs before `panelFixBeginRefusal`'s checks; check (b) opens
`reReviewBlocked`. A failed record read refuses nothing. Back-fills are included for (a) because
kan-1045's own `panel-fix-1` was back-filled — a live-only check would not have caught the run that
raised the finding.
**Considered:** a sanctioned skip in `skills/flow/review-panel.md` when every slot already hands back
a summary — the rule is not wrong (above), and the skip would make the text agree with the run by
weakening the text; enforcing at `flow stage end` of the review or in `check-unfinished-work` —
after the fix already read the raw reports; enforcing on a clean last round too — no begin follows a
clean round to hang the check on, and the finding's scope is the fix and re-run begins.

### The hook tests share one flow-guard cache

**ID:** hook-tests-share-guard-cache
**Status:** active
**Chosen:** `runHook` (`stats/internal/repotest/helpers_test.go`) appends
`FLOW_GUARD_CACHE_DIR=<dir>` after `HOME` and before the caller's `env`, `<dir>` one temp directory
the package builds flow-guard into once per run through a `sync.OnceValues`, exactly as
`guardCacheBuilt` (`stats/internal/guard/helpers_test.go`) does — `flow_guard_exec spec-root` from
`scripts/lib/flow-guard.sh` — and a `TestMain` removes it after the run. A caller's own
`FLOW_GUARD_CACHE_DIR` still wins.
**Considered:** exporting the operator's real cache (`os.UserCacheDir()/flow-guard`) — writes into
it, which `scripts/run-guard-tests.sh` deliberately avoids; a shared directory without the one-time
build — parallel first calls still race N builds; a retry or longer timeout on the hook tests —
widens a tolerance over a nondeterminism whose source is known.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 3 adds a refusal to `flow record dispatch begin` on every panel-fix
and re-review of every run; a predicate wrong by one round stops every review loop.
**Considered:** `flow-low` — the diffs are small, but the refusal sits on every run's path.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, `-fix`, Tasks 1–6 in series in this worktree.
**Considered:** one fixer per task — Tasks 2 and 4 both edit `skills/flow/implement.md`, and the six
diffs are small; two groups split Go from prose — the prose tasks carry their own Go-adjacent edits
(Task 4's test comments), so no split is file-disjoint enough to pay for a second context.

## Open questions

## Live check

Each step runs before landing, with every task committed, from this worktree.

1. **Narrowing.** Copy this change's `tasks.md` to a scratch `spectre/changes/lc/tasks.md` under a
   temp directory, then `cd stats && go run ./cmd/flow tasks files -C <tmp> lc 1 remove .flow/project.md`:
   before, exit 2 "expected <change> <task-id> add <path>..."; after, exit 0 and Task 1's
   `**Files:**` reads `none`. Not working looks like the exit 2 or an unchanged field.
2. **Merger refusal on real records.** `scripts/live-store-copy.sh up self-review-kan-1045 4175`, then,
   with the printed `FLOW_ADDR` exported, from `stats/`:
   `go run ./cmd/flow record dispatch begin -change kan-1045-flow-one-allowed-model-effort-list-per-subtask -role panel-fix -model opus -effort low -key panel-fix-2 -session-token mf-live-check-1045 -C ..`
   — kan-1045's records carry no merger row in either round. Before: exit 2 "fixes review round 1,
   which holds no Critical or Important finding"; after: exit 2 naming `panel-1-merger`. Then
   `... -role reviewer -key panel-1-primary ...`: before, admitted into the copy; after, exit 2
   naming `panel-0-merger`. Finish with `scripts/live-store-copy.sh down self-review-kan-1045`. Not
   working looks like the before messages, or an admitted row.

## Recommended approach

Fix each defect where it starts: the project text the plan copied (F-A); the CLI verb the guard's
refusal needs (F-B, `tasks-files-remove`); the begin that follows a round, where a skipped merger is
still repairable (F-C, `merger-enforced-at-begin`); the four citations (F-D); the test harness's
cache, after the trace (F-E, `hook-tests-share-guard-cache`); the flow-fast citation (F-F). Each
finding has one obvious fix, so no Explore or Plan agents were dispatched: every file the evidence
names was read directly.

## Critical files

- `.flow/project.md` — F-A companion-test name.
- `stats/cmd/flow/tick.go`, `stats/cmd/flow/tasks.go`, `stats/cmd/flow/tasks_test.go` — F-B verb.
- `skills/flow/implement.md` — F-B narrowing sentence, F-D citation.
- `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `skills/flow/review-panel.md` — F-C.
- `scripts/pipeline-fix.sh`, `stats/internal/guard/pipeline_fix_test.go`,
  `stats/internal/guard/check_loop_dispatch_test.go` — F-D.
- `stats/internal/repotest/helpers_test.go` — F-E.
- `skills/flow-fast/SKILL.md` — F-F.

## Existing code to reuse

- `filesAdd`, `taskBody`, `fencePattern`, `filesLinePrefix`, `ErrTaskIDInvalid` in `stats/cmd/flow/tick.go`;
  `TestTasksFilesAddWidensOneTask`'s fixture and `writePlan` in `stats/cmd/flow/tasks_test.go`.
- `panelFixBeginRefusal`, `panelFixRoundRE`, `panelFixKeyRE`, `reReviewBlocked`, `callRecord` in
  `stats/cmd/flow/record.go`; `records.PanelDispatchRound`; the `panelFixBegin` fakes in
  `stats/cmd/flow/record_test.go`.
- `guardCacheBuilt` / `guardCache` in `stats/internal/guard/helpers_test.go`; `flow_guard_exec` in
  `scripts/lib/flow-guard.sh`.
- `scripts/live-store-copy.sh`, `scripts/run-capped.sh`, `scripts/break-and-prove.sh`.

## Verification

Beyond each task's verify step: `scripts/run-guard-tests.sh` and `cd stats && make -s web-dist &&
go test -race -count=1 ./...` pass; the agents `## lint` list passes; the **Live check** above.
