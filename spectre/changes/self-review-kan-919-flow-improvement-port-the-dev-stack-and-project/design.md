## Context

The five findings of `proposal.md` are planned as one self-review change on kan-919's own branch
`spectre/kan-919-flow-improvement-port-the-dev-stack-and-project`, in its worktree
`/Users/tweety53/Projects/agents-worktrees/kan-919-flow-improvement-port-the-dev-stack-and-project`,
fix-base `9468cc05`. No path a finding names changed between the branch and `origin/main`
(`0e0d8a42`), and no archived change records a rejection of any remedy below.

Every finding is reproduced at `9468cc05`:

- F-A: `go test ./internal/guard -race -run 'TestCheckDevStackFresh$' -count=1` → `FAIL`, the race
  reported between `check_dev_stack_fresh_test.go:71` (handler read) and `:75` (test write).
- F-B: `scripts/run-verify-list.sh:121-122` removes its output directory on exit; `:131-133` prints
  the last 20 lines. <!-- measured: sed -n '131,133p' scripts/run-verify-list.sh @ 9468cc05 → tail -n 20 -->
- F-C: `skills/flow/verify-and-handoff.md` line 115-116 grants the re-run with no exception.
- F-D: kan-919's `tasks.md` at `75ded465`, every backtick stripped from its `**Files:**` lines,
  passes `check-plan-shape.sh` with exit 0.
- F-E: case 28.b of `TestRunReproducer` pins today's behaviour — a marker line carrying extra text
  "declares nothing" and runs generic (`want: 1`).

The repair half of each fix: F-A's race lives in the test this change repairs; F-B's lost outputs
cannot be recovered; F-D's rule measured over every `tasks.md` in this repository (202 plans) flags
none, so no in-flight plan needs rewriting; no committed reproducer carries a near-miss marker line
(`git grep` over non-Markdown files at `9468cc05`), so F-E has nothing written wrong.

## Live check

None — the change edits guard scripts, a test and prose; no daemon, store or running service is
touched.

## Decisions

### Repair the race with an atomic value

**ID:** race-atomic-served
**Status:** active
**Chosen:** `served` becomes a `sync/atomic` `atomic.Value` the handler `Load`s and the test
`Store`s — the smallest change that gives the race detector an ordering edge.
**Considered:** a `sync.Mutex` around the string — two more lines for the same edge; a fresh
httptest server per subtest — restructures the case for no gain.

### Keep a failing command's full output in the git dir

**ID:** verify-output-kept-in-git-dir
**Status:** active
**Chosen:** on a non-zero exit, `run-verify-list.sh` copies that command's output file to
`<git-path flow-verify-failed>/<list>-<n>.log` (`<n>` the command's 0-based list position,
`lint-spec-reach.log` for the appended check) and prints `  full output: <path>` as the block's last
line; each run first removes that list's earlier `<list>-*.log` files, so a path printed always
holds this run's output. Outside a git worktree nothing is kept and no line printed.
**Considered:** keep the `mktemp` directory on failure — it sits in `TMPDIR`, outlives the worktree
and is never cleaned; print the whole output — a 1000-line `-race` log floods the `## Report`;
key the file by sha — a stale log from an earlier HEAD would read as this run's.

### A data race is never the environmental-flake case

**ID:** race-never-environmental
**Status:** active
**Chosen:** output carrying `WARNING: DATA RACE` or `race detected during execution of test` takes
**The loop** whatever the re-run exits; `run-verify-list.sh` prints `  data race — never
environmental` in that command's block whenever its full output carries either string — read from
the whole output, since the banner opens a `-race` log and the 20-line tail drops it. The prose is
`patches/task-3.diff`, listed in this change's `verbatim-moves.txt`.
**Considered:** prose alone — the parent reads only the tail and never sees the banner; skipping
the re-run for a race — the re-run is harmless once its result cannot clear the command.

### Refuse an unbackticked multi-path field

**ID:** plan-shape-f16
**Status:** active
**Chosen:** a new finding F16 in `scripts/check-plan-shape.py`: a `**Files:**` or
`**Allowed-collateral:**` value carrying no backticked token, not opening with `none`
(`NONE_OPEN_RE`), that contains a comma or whitespace. The raw values come from the parser itself:
`TaskFields` gains `files_value` and `collateral_value`, the joined strings `parse_task_fields`
already builds, exactly as it exposes `tests_value`. Measured over 202 plans in this repository:
0 hits; over gymie's 41: 3, each an `**Allowed-collateral:** none, unless …` sentence — which is
why `none`-opening values are exempt.
**Considered:** flag any token containing a comma or whitespace — fires on four backticked
`Allowed-collateral` phrases in archived plans (kan-271, kan-472) and needs no parser change only
by guessing which tokens were backticked; re-scan the raw field line in check-plan-shape — a
second grammar for continuation lines, the drift the guard's docstring forbids.

### Refuse a near-miss mutation marker

**ID:** mutation-marker-near-miss-refused
**Status:** active
**Chosen:** `run-reproducer.sh` exits 2, before execution (the shape class), when no line of the
reproducer's first 10 is exactly `# mutation-reproducer` but one, leading whitespace trimmed, opens
with it; the refusal names the line. A marker that appears mid-line stays generic. Case 28.b of
`TestRunReproducer` flips from "declares nothing" to refused — KAN-623 pinned the silent generic
reading without weighing that its author meant the inverted one, which is the defect.
**Considered:** accept the near-miss as a declaration — guesses intent the other way and widens
the marker `check-mutation-reproducer-pin.sh` pins; a warning on stderr — a reviewer's caller
reads the exit code, not stderr.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-medium — the change moves two guard verdicts (a new plan-shape finding, a new
run-reproducer refusal) and the verify runner's report, besides a concurrency repair in a test.
**Considered:** flow-low — fits prose-only or mechanical changes, not new refusals on the planning
and review paths.

### Fixer groups

**ID:** fixer-groups
**Status:** active
**Chosen:** one fixer dispatch, g1, Tasks 1, 2, 3, 4, 5 in order — five small tasks in one
repository, Tasks 2 and 3 sharing two files.
**Considered:** two groups (Tasks 1–3 on the verify path, Tasks 4–5 on the guards) — a second cold
start buys no parallelism, since fixer groups run in series; one dispatch per task — five cold
starts for five small tasks.

## Open questions
