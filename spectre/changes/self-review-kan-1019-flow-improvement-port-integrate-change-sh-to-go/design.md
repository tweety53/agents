## Context

The six findings of `proposal.md` are planned as one self-review change on kan-1019's own branch
`spectre/kan-1019-flow-improvement-port-integrate-change-sh-to-go`, in its worktree
`/Users/tweety53/Projects/agents-worktrees/kan-1019-flow-improvement-port-integrate-change-sh-to-go`,
fix base `1420a779`. The base `origin/main` is still `a03f3c4b`, kan-1019's merge base, so no path a
finding names moved upstream; no archived change records a rejection of any remedy below. The
self-review pass's report, with its evidence, is the dispatch's input; D2(b) is `big` and stays
with the pass.

Evidence at `1420a779`:

- A: kan-1019's archived `design.md` carries a `## Live check` naming two `prepare` runs; its change
  directory holds no `live-verification.md`. `check-live-verification` is called by no step of
  `integratechange_prepare.go`. A survey of every tracked `design.md` with a `## Live check`: each
  no-runtime section opens with `None`, `No ` or `Not applicable` — several wrap over two or more
  lines, so "one line" is not a usable test — and every section naming something to exercise opens
  otherwise.
- B: `scripts/generate-relocation-comparison.sh` on a clone of this branch at `1420a779`, plan
  kan-1019's archived `tasks.md`, merge base `a03f3c4b`, exits 2: "scoped file missing from
  worktree: …/scripts/test-integrate-change.sh". `grcGitShow` already reads a path absent at the
  merge base as empty; the worktree side has no such case.
- D2: `grep -n "CALLER_DIR\|padded\|indented" stats/internal/guard/integrate_change*_test.go` finds
  nothing. The reproducers `.superpowers/sdd/reproducers/0-primary-{1,2,3}.sh` (untracked) carry
  each defect's shape and are the cases' source.
- C: `scripts/check-plan-shape.sh` on `git show 8f1f3eb4:spectre/changes/kan-1019-flow-improvement-port-integrate-change-sh-to-go/tasks.md`
  exits 0 although tasks 3 and 4 share one `measured:` command and both declare `before=1 after=2`.
- E: kan-1019's rows `panel-0-primary+principles` (id 7101) and `panel-0-test-sanity+call-efficiency`
  (id 7102) both end at 20:51:54+03:00; the slot's report files were written at 20:51:26
  (`panel-report-0-principles.md`, the later of the bundle's two) and 20:39:20
  (`panel-report-0-test-sanity.md` and `-call-efficiency.md`).
  <!-- measured: flow record dispatches -change kan-1019-flow-improvement-port-integrate-change-sh-to-go and ls -lT .superpowers/sdd/panel-report-0-*.md @ branch spectre/kan-1019-flow-improvement-port-integrate-change-sh-to-go, machine-local -->
  `flow record dispatch end` already takes `-ended-at` and the store keeps it (`ApplyDispatchEnd`
  stamps now only when it is zero).
- D1: the self-review pass measured `TestIntegrateChange` on this worktree; `narrative_format`
  (six fixtures in one bash session) was its slowest subtest and `rerun_pin` (two) the next.
  <!-- measured: go test ./internal/guard -run TestIntegrateChange -v by the self-review pass @ 1420a779, machine-local, not re-run here -->

Repair:

- A left kan-1019 without its live check. It is run here, at plan time, on the stand-in decision
  `repair-kan-1019-live-check` names, and its record committed with this plan, so a re-run of
  kan-1019's `prepare` after Task 1 lands passes the new gate.
- E left two wrong `ended_at` values in the dev store; Task 5 re-times them from the report files.
  Every earlier change's parallel panel rows carry the same defect, but their report files left
  with their worktrees, so no evidence remains to re-time them from.
- B, C, D1 and D2 persisted nothing wrong: kan-1019's panel round is closed, its plan's Baseline
  was corrected in place, and the tests and cost were in-run.

## Live check

None — the change edits one-shot guard processes (`flow-guard` subcommands and their shims), their
tests and skill prose; Task 5's store repair is a one-time correction verified by its own read-back.

## Decisions

### The live-check gate lives in integrate-change prepare

**ID:** live-check-gate-in-prepare
**Status:** active
**Chosen:** `prepare`, after its archive decision and before the unfinished-work signals, reads the
change's `design.md` (the live directory, else the archived one). A `## Live check` section whose
first non-blank line opens with neither `None`, `No ` nor `Not applicable` names a check; then
`<that directory>/live-verification.md` must exist and pass `check-live-verification`, else
`STOP: live-check`, which `--accept-outstanding` never passes. `check-live-verification` joins the
shim's sibling list. The no-runtime test is the section's opening word because a survey of every
tracked `## Live check` found justifications wrapping over several lines.
**Considered:** gating at `flow.verify`'s close — the stage close is a mark a run can skip, where
`prepare` is the one step no landing skips (the pass's recommendation); counting the section's
lines — wrapped `None — …` justifications would trip it; a new marker line — every existing design
would need rewriting.

### A live check is runnable at verify through a stated stand-in

**ID:** live-check-stand-in
**Status:** active
**Chosen:** `brainstorm-planner.md` requires every step of a live check to run at `flow.verify`,
before landing; a step needing the landed change, a restarted shared service or a write to the
change's own integrate state is written as the stand-in that runs in its place, named in the
section. `verify-and-handoff.md` states the no-runtime form and that the check is never deferred
past verify; `integrate.md` handles `STOP: live-check`; `finish-contract-run1.md` lists the gate
beside the visual-verify one.
**Considered:** letting a design mark a check "post-landing" — the gate would then pass exactly
the deferral kan-1019 made.

### kan-1019's live check runs on a scratch fixture at plan time

**ID:** repair-kan-1019-live-check
**Status:** active
**Chosen:** the two `prepare` runs kan-1019's design names, made on a scratch two-repository fixture
of the integrate-change harness's shape at one fixed path (bash at `a03f3c4b` with this tree's
other scripts beside it, then the shim at `1420a779`), from inside the fixture's canonical
worktree, compared on stdout, stderr, exit code and stage marks; a clean fixture and a
staged-foreign one both matched. The record is committed with this plan as
`spectre/changes/archive/kan-1019-flow-improvement-port-integrate-change-sh-to-go/live-verification.md`.
**Considered:** running `prepare` on kan-1019's real record — writes its integrate stage marks and
state mid-run; leaving it to a fixer task — the record lies under the planning paths, which a task
commit may not touch (check-plan-shape F14).

### A branch-deleted scoped file reads as empty

**ID:** relocation-deleted-file-empty
**Status:** active
**Chosen:** when a scoped path is absent on disk and `git -C <worktree> ls-files --error-unmatch --
<path>` exits 1 (not in the index), its worktree text is empty, so its passages classify as moved or
removed. Any other read failure — a path still in the index, an `ls-files` exit other than 1 — stays
exit 2. Case g keeps its exit-2 pin, relabelled as the still-tracked case.
**Considered:** reading every missing file as empty — hides a worktree damaged on disk; resolving
the branch's tip with `git show HEAD:<path>` — misses an uncommitted deletion the plan's tasks made.

### Parity fixes get suite cases from their reproducers

**ID:** parity-fix-suite-cases
**Status:** active
**Chosen:** F2 and F3 become cases of `TestIntegrateChangeLandCleanup` in the harness's own shape
(reproducers 2 and 3); F1 becomes a case of a new `TestIntegrateChangeShim` in
`integrate_change_test.go` that runs the real shim from inside the fixture's worktree with a
relative main-checkout argument, since the in-process `sut` cannot see the process's cwd. Each is
proven by reverting its fix in place and watching the case fail.
**Considered:** tracking the reproducer scripts — they compare against an oracle at `a03f3c4b` and
build a module copy; a suite case is cheaper and runs in every `go test`.

### Baseline chains are checked in plan order

**ID:** plan-shape-baseline-chain
**Status:** active
**Chosen:** check-plan-shape F17: tasks whose Baseline carries the same `measured:` command, in plan
order, each compared with the previous such task — a `before` below that task's `after` is
reported at the Baseline line; the later task of a `**Squash-with:**` pair is exempt, since the pair
shares one commit. Measured against every tracked `tasks.md`: the rule flags none but kan-499's
squash pair (exempt), and flags kan-1019's original plan at task 4.
<!-- measured: python3 chain scan over git ls-files '*tasks.md' by the planner @ 1420a779, machine-local script; strict == form: 3 breaks, < form: 1 (a Squash-with pair) -->
**Considered:** `before` equal to the previous `after`, the pass's wording — flags kan-916 task 9 and
kan-939's self-review task 8, both legitimate: a task between them added tests to the same source
set under another `measured:` command; changing the convention line — reaches its many citers and
makes the fix `big`.

### Each panel slot ends at its report's mtime

**ID:** panel-end-report-mtime
**Status:** active
**Chosen:** the one `end` call per round keeps its shape; each `end` passes
`-ended-at "$(date -u -r <report-file> +%Y-%m-%dT%H:%M:%SZ)"`, the slot's own report file (a bundled
dispatch's, the later of its files). kan-1019's rows 7101 and 7102 are re-timed with
`flowd retime-dispatch`, keeping their start instants.
**Considered:** one `end` call per report as each appears — the round's wait is one call by the turn
discipline of `skills/flow/implement.md`; a `flowd` change deriving the end from the report —
the daemon never sees the report path.

### Multi-fixture cases split into parallel entries

**ID:** split-multi-fixture-cases
**Status:** active
**Chosen:** `narrative_format`'s five passing fixtures and `rerun_pin`'s reclone fixture become
their own `icCaseDef` entries; the shared bash helper (`narrative_passes`, `rerun_pin_case`) is one
Go string constant prepended to each entry's script; every check label is unchanged.
**Considered:** a Go loop over a fixture table — the same entries with less readable labels; leaving
`rerun_pin` as is — it is the next critical path once `narrative_format` splits.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-medium — three guard verdicts move (a new `prepare` stop, a new plan-shape finding,
the relocation generator's exit on a deleted file) beside tests, prose and a store repair.
**Considered:** flow-low — fits prose-only or mechanical changes, not a changed guard verdict.

### Fixer groups

**ID:** fixer-groups
**Status:** active
**Chosen:** two fixer dispatches in series — g1: Tasks 1, 3, 6 (integrate-change: the gate, the
parity cases, the split; one test harness, slow suites), then g2: Tasks 2, 4, 5 (relocation
generator, plan-shape F17, the panel prose and store repair). Task 4's `**After:** Task 1` and Task
6's are met by the order.
**Considered:** one dispatch for all six — one context carries both slow integrate-change suite runs
and three unrelated guards; one per task — six cold starts in series.

## Open questions
