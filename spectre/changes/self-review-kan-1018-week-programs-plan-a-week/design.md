## Context

Self-review fixes for `kan-1018-week-programs-plan-a-week`, on branch
`self-review-kan-1018-week-programs-plan-a-week` in
`/Users/tweety53/Projects/agents-worktrees/self-review-kan-1018-week-programs-plan-a-week`, based at
`d8471213`. Task 4 alone edits the gymie worktree
`/Users/tweety53/Projects/gymie-worktrees/kan-1018-week-programs-plan-a-week` on branch
`spectre/kan-1018-week-programs-plan-a-week`.

**P1.** `commit_lacks_trailer` (`hooks/protect-main-checkout.py`) returns False for any message the
command text does not state — the right answer for the commit-on-`spectre/*` refusal, which must
not guess, but the commit-then-push branch reuses it, so an unreadable message pushed in the same
call is never judged by anything: `untasked_push` runs before the command and still sees HEAD as
the previous commit.

**P2.** `insertDispatch` (`stats/internal/store/records.go`) keeps `dispatches.started_at` on a
repeated begin unless the row is an aborted re-dispatch (`reBegin`). Backfill is begin-then-end, so
on a row begun late and never ended the derived earlier instant is discarded and `EndDispatch`
refuses with `ErrInlineBegunAfterCommit` again. The `-tokens` refusal under
`CLAUDE_CODE_SESSION_ID` is correct (the harvester credits the row) and stays.

**P3.** `skills/flow/review-panel-fix-round.md` requires a `fix-mutation:` line per fixed finding
(a `none — <reason>` line for a fix with no executable behaviour), recorded as `records.Mutation`
rows with a `Round`. `check-unfinished-work` (`stats/internal/guard/unfinishedwork.go`) already
reads the change's findings from the store and refuses open ones; nothing reads mutations.

**P4.** Archived text in the gymie worktree; `WeekProgramController.kt:44` answers
`ResponseEntity.created(…)` (201), and the panel record holds 6 Important and 6 Minor round-0
findings.

**Repairs of what the defects wrote.** P4 is itself the repair. P1's commit `863f872a` is on
origin and is not rewritten (published history). P2's open kan-1018 row is repaired by re-running
its backfill once the fix is deployed to flowd (see Verification). P3 cannot fabricate mutation
proof for rounds that never measured one; kan-1018 has not landed (`IN_PROGRESS`), so its fix rounds are proved and recorded in its own worktree before it integrates.

## Decisions

### Unreadable commit message pushed in the same call

**ID:** p1-deny-unreadable-commit-push
**Status:** active
**Chosen:** in the commit-then-push branch, refuse when the same-target commit's message is unreadable (`commit_lacks_trailer`'s "never judged" cases) and the push sends to `spectre/*`; the remedy is to commit and push in two calls, where `untasked_push` sees the commit. `commit_lacks_trailer` is split so a small `commit_message_unread(args)` returns the unreadable verdict both callers share; the commit-only refusal keeps not judging such messages.
**Considered:** reading `-F <file>` contents — covers one of seven unreadable forms and not `$(…)` or `--amend`; a post-command hook — the push has already happened by then.

### Backfill moves an open row's begin back

**ID:** p2-open-row-earlier-begin
**Status:** active
**Chosen:** extend the ON CONFLICT `started_at` arm: an open row (`dispatches.ended_at IS NULL AND dispatches.outcome IS NULL`) takes `EXCLUDED.started_at` when it is earlier than its own. Moving only earlier keeps an in-order journal replay a no-op and never shortens a window.
**Considered:** a new `flowd` repair subcommand — a second path for what backfill already promises; overwriting `started_at` on any re-begin — a stale replay could move a live row's begin later.

### Fix-mutation gate in check-unfinished-work

**ID:** p3-mutation-row-per-fixed-round
**Status:** active
**Chosen:** a new signal in `check-unfinished-work`: for each round holding a finding whose status is exactly `fixed`, at least one `fix-mutation` row of the round that fix ran in must exist, else refuse with the round numbers. A `none — <reason>` row (doc-only fix) satisfies it; `withdrawn <reason>` findings count for nothing. The mutation rows are read through the same `flow record` call family the guard already uses for findings.
**Considered:** one row per fixed finding — mutation rows carry a path, not a finding ref, so the pairing is unreadable; checking the rendered panel.md text — the store is the record.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — a containment hook, the store's begin upsert and a landing gate; each misfire blocks or misattributes every later run.
**Considered:** `flow-low` — each fix is small and test-pinned, but the SQL arm and the round mapping are easy to get subtly wrong.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** `fix-g1` Tasks 1–3 (agents worktree, one fixer, sequential — one mutator per worktree); `fix-g2` Task 4 (gymie worktree, docs). Both on `flow-low`/opus.
**Considered:** one fixer per task — three agents mutating one worktree collide.

## Open questions

none

## Recommended approach

Three targeted fixes at the shared function each defect routes through (the hook's commit-then-push
branch, the dispatch upsert, the unfinished-work guard), plus a two-line doc correction in gymie.
Rejected alternatives: see `p1-deny-unreadable-commit-push`, `p2-open-row-earlier-begin`,
`p3-mutation-row-per-fixed-round`.

## Critical files

- `hooks/protect-main-checkout.py` — commit-then-push refusal.
- `stats/internal/repotest/protect_main_checkout_test.go` — hook tests.
- `stats/internal/store/records.go` — `insertDispatch` ON CONFLICT arm.
- `stats/internal/store/records_test.go` — store tests.
- `stats/internal/guard/unfinishedwork.go`, `stats/internal/guard/check_unfinished_work_test.go` — the gate.
- gymie: `spectre/changes/archive/kan-1018-week-programs-plan-a-week/live-verification.md`, `…/narrative.md`.

## Existing code to reuse

- `commit_lacks_trailer`, `push_refspecs` in `hooks/protect-main-checkout.py`.
- `reBegin` / `insertDispatch` in `stats/internal/store/records.go`.
- `pcFlow(env, …, "record", "findings", …)` and `uwOpenFindings` in `stats/internal/guard/unfinishedwork.go`; `records.Mutation` in `stats/internal/records/types.go`.

## Verification

- Each task's targeted tests, then `cd stats && go test ./... ` and `go vet ./...` once after the group.
- P3 measured against history: run the built guard over two recent archived changes that did record mutation rows (`flow record passes -change <name>`) and confirm no refusal, and over kan-1018 and confirm it refuses naming its fix rounds.

## Live check

The change touches the `flowd` store (P2) and the landing gate (P3). At `flow.verify`, against a
scratch fixture (the store tests' own database, never the protected `flowd` on `127.0.0.1:4173`):
begin an inline row, begin again with an instant 6 s earlier, end with a commit time between the two
— before: end refuses `ErrInlineBegunAfterCommit`; after: it closes with the earlier begin. Not
working looks like the refusal still printed. Repair after landing (operator-run, once flowd is
restarted on the landed build): re-run `flow record dispatch backfill` for kan-1018's open inline
row named in the pass report; its `started_at` reads the derived instant. Straight after landing,
run `make restart` in `agents/stats`: it reinstalls the `flow` CLI, without which the landed
`check-unfinished-work` calls `flow record mutations`, an unknown verb, and exits 2.
