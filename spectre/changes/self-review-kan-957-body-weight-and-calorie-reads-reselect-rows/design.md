## Context

- Three non-`big` findings from the kan-957 self-review pass, planned as one change per item 1 of
  **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`). Each is carried by
  exactly one task; none is dropped. All were checked against this worktree at `d1085fb4` and the
  gymie worktree `kan-957-body-weight-and-calorie-reads-reselect-rows` at `f3d2675d`:
  - F-a (Task 1, gymie): the repro
    `/private/tmp/claude-501/-Users-tweety53-Projects-gymie/19d20152-1ee3-4781-ba2d-176d4703d990/scratchpad/repro-leading-newline.sh`
    sources gymie's `scripts/count-request-queries.sh` and feeds `count_slice` one single-line
    select and one `execute <unnamed>: ` line followed by two tab-indented SQL lines; it prints
    `statements=1 begin=0 select:weigh_in=1`, where 2 statements and `select:weigh_in=2` are
    right. In `count_slice` the empty text fails `text != ""`, so nothing is counted and `held`
    stays empty, and the `/^\t/ && held != ""` rule then never takes the continuation.
  - F-b (Task 2): `grep -rn 'record summary' skills scripts hooks` finds the prescription alone
    (`skills/flow-contracts/pipeline.md`); no script or hook reads a summary row. The store has no
    summary read route of its own (`stats/internal/api/server.go` serves only
    `POST …/{change}/summary`); the one read is the self-review bundle
    (`GET /api/v1/self-review/{project}/{change}/bundle`, `flow self-review bundle`), whose header
    prints `skipped: change summary (absent)` when the store holds none
    (`stats/internal/selfreview/git.go`, `Bundle`). `check-unfinished-work.sh` already runs at both
    places a refusal belongs — **Write `IN_PROGRESS`**'s self-check and integrate's `prepare` gate
    (`scripts/integrate-change.sh`) — and already refuses (exit 2) a failed store read on signals
    three and five. **Write `IN_PROGRESS`** recorded the `**Summary:**` bullets only "before the
    block prints", after the guard, so the new signal would fire on every creating run's own
    self-check; the recording moves ahead of the guard. `scripts/test-integrate-change.sh` stubs
    `flow` and answers nothing for `self-review bundle`, so its stub gains a found-summary header.
    `/flow-fast` never runs this guard, so it is untouched.
  - F-c (Task 3): **Live check** (`skills/flow/verify-and-handoff.md`) says "write each
    before/after figure" and nothing about where the before figure is measured.
- Spiked at plan time, every edit reverted: Task 1's diff applied to a copy of gymie's two scripts —
  the repro prints `statements=2 begin=0 select:weigh_in=2`, the harness passes, and with
  `count-request-queries.sh` restored its new case fails (`statements=2 begin=0 select:weigh_in=1
  select:calorie_entry=1` where `statements=3 … select:weigh_in=2 …` is expected); `git apply
  --check` passes in the gymie worktree. Tasks 2 and 3 applied to a detached worktree of this
  checkout at `d1085fb4`: `gofmt -l`, `go vet ./internal/guard/` and `go test ./internal/guard/
  -count=1` pass; with `unfinishedwork.go` reverted the four new failing-side cases of
  `TestCheckUnfinishedWork` fail; `bash scripts/test-integrate-change.sh` prints `161 passed, 0
  failed`; `check-verbatim-moves.sh` (with this change's `verbatim-moves.txt`),
  `check-references.sh`, `check-markdown-integrity.py`, `check-dispatch-paragraphs.sh`,
  `check-installed-citations.sh`, `check-guard-symlinks.sh`, `check-stage-mark-calls.sh`,
  `check-ask-silence.sh` and `check-python-suppressions.sh` exit 0, and
  `check-normative-inventory.sh` prints the same inventory before and after. Both agents diffs pass
  `git apply --check` here.

## Decisions

### An empty statement line takes its tab-indented continuation as its text

**ID:** leading-newline-continuation
**Status:** active
**Chosen:** in `count_slice`, an empty statement text from one of the database's pids sets
`pending`; the next line, when tab-indented, is counted through one shared `count()` (trim,
`SET application_name` skip, `BEGIN` count, `held`), so the continuations after it attribute as
before; any other line clears `pending`, leaving an empty statement with no continuation the
uncounted validation query. The harness gains one case holding both shapes.
**Considered:** joining every continuation before counting — reshapes the whole slice for one
case; treating every empty statement as counted — counts the pool's validation query, which the
harness's first case pins as uncounted.

### No recorded summary is unfinished work, read through the self-review bundle

**ID:** summary-is-unfinished-work
**Status:** active
**Chosen:** signal six in `check-unfinished-work` (Go, `stats/internal/guard/unfinishedwork.go`):
on the worktree holding the plan, `flow self-review bundle -change <name> -C <anchor>`; a header —
the lines above the first `## ` section — carrying `skipped: change summary (absent)` adds the
breakdown reason naming `flow record summary -change <name> -C <anchor> -file -`; a failed read or
a header with no `found:` line is exit 2, as signals three and five refuse theirs. **Write
`IN_PROGRESS`** records the `**Summary:**` bullets before the guard runs, and integrate.md's
`STOP: unfinished-work` bullet says the summary line is closed by recording the summary and
re-running `prepare`.
**Considered:** a new `GET …/summary` route, client method and CLI read — four more files for a
read the bundle already serves; a separate stop in `integrate-change.sh prepare` — a second gate
beside the one that already sequences the change's store reads, and it would miss the
`IN_PROGRESS` self-check, where the run that printed the summary can still record it; a hook on
the handoff print — no tool call marks the print.

### A before figure comes from a detached worktree at the merge base

**ID:** live-check-before-from-detached-worktree
**Status:** active
**Chosen:** one sentence in **Live check**: `git -C <worktree> worktree add --detach
<project>-worktrees/<name>-before <merge-base>`, the base's stack run there, the worktree removed
once the figure is written — never base sources checked out over the change worktree.
**Considered:** `git stash` plus a base checkout in the change worktree — the hand-restore kan-957
paid for; measuring the before figure on the main checkout — another session's checkout, and the
protect-main-checkout hook's ground.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `medium` — Task 2 adds an `OUTSTANDING` signal and an exit-2 read to the gate every
integrate runs; a low-effort read could pass over a fixture or ordering gap there.
**Considered:** `low` — Tasks 1 and 3 are small, but Task 2 carries the review risk.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** one group, g1: Tasks 1–3 in plan order, one fixer — each task is a spiked diff applied
with `git apply`, Task 1 in the gymie worktree and Tasks 2–3 here.
**Considered:** splitting Task 1 off by repository — one more dispatch for a diff applied in a
named directory.

## Open questions

## Live check

None for this change's run: Task 1 changes a log parser covered by its shell harness over recorded
log text, Task 2 a guard covered by Go table tests and the integrate harness's stubbed store, and
Task 3 prose.
