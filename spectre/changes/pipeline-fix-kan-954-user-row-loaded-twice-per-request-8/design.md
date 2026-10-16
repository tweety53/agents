## Context

`lpFromKey` (`stats/internal/guard/loop.go`) maps a loop key to the change its flowd rows are filed
under. A `self-review-<name>-…` key with no `self-review-<name>` branch in the agents repository is
a self-review on the change's own branch and files under `<name>`. Every flow call and git-config
read for the loop then runs with `-C <agents main checkout>` (`lpMain(repo)`): `lpRows`,
`flow record dispatch begin`, `lpEscalated`, `lpFirstReviewType` in `checkLoopDispatch`, and the
same in `pipelinefix.go`'s `pfEnd` and `pfEscalate`. `flow`'s `requireChangeCheckout`
(`stats/cmd/flow/record.go`) refuses a row for a change no worktree of `-C`'s repository carries, so
for a change of any other project the loop cannot record a single dispatch. Any `-C` inside the
change's own repository works: `requireChangeCheckout` searches all its worktrees.

`check-loop-landing.sh` is not affected: `hooks/protect-main-checkout.py` passes it the repository
being pushed, and `pfLand` only lands agents loop worktrees. `pipeline-fix.sh verify` resolves its
diff base from the agents checkout's `origin/HEAD`; for a gymie worktree that is `origin/main`
instead of gymie's `origin/develop`.

Reproduced at plan time at b803ec00: the two Go tests and hook case 12 below fail on the untouched
tree and pass with the fix, and `cd stats && go test ./internal/guard/ -count=1` stays green.

Blast radius — the defective sources, their tests, the skill, and every other file `grep -rl`
finds for `lpLoop`, `lpFromKey`, `pfBase`, `enforce-loop-dispatch`, `check-loop-dispatch` and the
changed paragraph — not `big`: 16 files.
<!-- measured: grep -rl -e lpLoop -e lpFromKey -e pfBase -e enforce-loop-dispatch -e check-loop-dispatch -e 'Every loop dispatch is admitted' . excluding .git and spectre archive (15 files) plus check_loop_dispatch_test.go @ b803ec00 -->

## Decisions

### Resolve an own-branch self-review's project from the caller's working directory

**ID:** caller-dir-resolves-own-branch
**Status:** active
**Chosen:** `lpLoop` gains `main`, the checkout every flow call and git-config read for the loop
runs with `-C`. It is the agents main checkout for a pipeline-fix loop and a `self-review-<name>`
loop branch; for an own-branch self-review it is the main checkout of `env.Dir` (flow-guard's
working directory) when that is a git checkout, else the agents checkout. No new argument: the
working directory already reaches flow-guard as `env.Dir`.
**Considered:** a `-C <dir>` option on `check-loop-dispatch.sh` and `pipeline-fix.sh` — every
caller in the skills would have to type it, and a parent that forgets it repeats this defect; an
env override — same, invisible in the command text; searching every known project for the change
— no registry of projects exists to search.

### The hook runs the checker in the payload's cwd

**ID:** hook-runs-in-payload-cwd
**Status:** active
**Chosen:** `hooks/enforce-loop-dispatch.py` passes the PreToolUse payload's `cwd` as the
checker's working directory when it is an existing directory, else inherits its own.
**Considered:** a fifth checker argument — changes the shim's usage line and every caller for what
the working directory already carries.

### verify diffs against the worktree's own origin/HEAD

**ID:** verify-base-from-worktree
**Status:** active
**Chosen:** `pfBase` takes the directory to read `origin/HEAD` from; `verify` passes the worktree,
`start` and `land` keep the agents checkout (identical for an agents loop worktree).
**Considered:** leaving verify as is — a gymie worktree would be diffed against `origin/main`,
which gymie's default branch is not.

### The no-stop sentence lives in step 3's mechanical-admission paragraph

**ID:** no-stop-sentence-in-step-3
**Status:** active
**Chosen:** one sentence appended to `skills/flow-self-review/SKILL.md`'s **Every loop dispatch is
admitted and recorded mechanically.** paragraph — the paragraph that introduces the tools, read by
every self-review and in-run fix parent (pipeline.md's section cites step 3 as canonical for the
loop) — citing **Pipeline defects found mid-run** for how the fix runs.
**Considered:** a sentence in `skills/flow-contracts/pipeline.md` as well — restates the same rule
in a second file; pipeline.md alone — a self-review parent running step 3 reads the skill, not that
section.

### Memories are untouched

**ID:** no-memory-update
**Status:** active
**Chosen:** no saved memory changes: none encodes that a loop tool resolves from the agents
checkout (checked `MEMORY.md`'s index for loop, pipeline-fix and self-review entries).
**Considered:** adding a memory for the no-stop rule — the skill now states it.

### First review on medium

**ID:** first-review-effort
**Status:** active
**Chosen:** the first review, `pipeline-fix-8-review-1`, runs on `flow-medium` (effort `medium`) on
`opus`: the change reroutes where every own-branch loop's flowd rows and escalation markers go,
and a gate hook's working directory.
**Considered:** `low` — too little for a change to which flowd project a gate records.

### One fixer group

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group — Tasks 1, 2 and 3 in one `pipeline-fix-8-fix` dispatch: three small,
file-disjoint tasks whose diffs are given verbatim.
**Considered:** one group per task — three dispatches for edits that apply as given.

## Open questions

## Live check

No runtime state changes: the guard and the hook are one-shot processes per call. After landing,
the gymie run's next own-branch self-review dispatch (`Key: self-review-kan-954-…-<suffix>`) is
admitted and `flow record dispatches -change kan-954-user-row-loaded-twice-per-request -C
/Users/tweety53/Projects/gymie` lists its row; failure looks like the same `no checkout of the
repository holding …/agents carries change` deny.
