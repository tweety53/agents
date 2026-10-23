## Context

Defect verified by the self-review pass of `kan-1021-flow-improvement-move-the-bash-hook-and-setup`:
the change's branch reflog shows `32908088` (`chore(spectre): plan self-review-<name>`) reset to
`f85971db`, then `5bbe4b50` "chore(spectre): plan kan-1021-…" carrying the self-review plan files
under the parent change's planning subject.

On `<agents repo>` a change's branch gathers, in order: the implementation commit,
`chore(spectre): plan <name>`, `chore(spectre): archive <name>`, then the self-review loop —
`chore(spectre): plan self-review-<name>`, its fix commits, `chore(spectre): archive
self-review-<name>` — and last `docs(self-review): <name> self-review report` (origin/main at
`e77f951b`: `88385198`, `853ac4a5`, `43daaca3`, `0774bfd8`, `e77f951b` for kan-1020).
<!-- measured: git log --oneline -25 origin/main @ e77f951b -->

`archBase` (`stats/internal/guard/integratechange.go`, called from `integratechange_prepare.go` for
the new-work test and the unfinished-work gate's ranges) recognises only the report or bundle, then
`archive <name>`. Reproduced at `e77f951b`: case `archived_rerun_self_review` (Task 1) — a run 1
resumed above `plan self-review-demo` and one fix commit — prints `ARCHIVED: demo — with new work`
and stops at `STOP: unfinished-work` listing the fix commit as `UNTASKED`; with the fix it prints
`ARCHIVED: demo — without new work`, and `go test ./internal/guard -run TestIntegrateChange` passes.
<!-- measured: go test ./internal/guard -run 'TestIntegrateChangePrepareCommit/archived_rerun_self_review' in stats/ @ e77f951b → FAIL, "ARCHIVED: demo — with new work"; with Task 1's diff → ok, and -run TestIntegrateChange → ok -->

The contract sentence lives in `skills/flow-contracts/finish-contract-run1.md`, its reasoning in
`skills/flow-contracts/finish-contract-rationale.md`; `git grep "re-run's base"` outside archives
finds no other citer. No saved memory encodes the defect.
Blast radius: 4 files — the function, its test, the contract sentence and its rationale; not `big`.
<!-- measured: git grep -n "re-run's base\|archBase" -- skills rules scripts stats @ e77f951b → the four files above -->

Change kan-1021's misplaced commit is not repaired here: that branch is the running change's own,
reshaped by its own run 1, and is not this loop's to rewrite (see `no-history-repair`).

## Decisions

### Where the reshape base stops

**ID:** self-review-loop-base
**Status:** active
**Chosen:** walk `<merge-base>..HEAD` oldest-first: `chore(spectre): plan self-review-<name>` sets the base and opens the loop; every later commit extends the base while the loop is open; a fix run's `chore(spectre): plan <name>` (bare or with its ` — <stage>` suffix) closes it; the report, the bundle and `chore(spectre): archive self-review-<name>` set the base and close it. The last base set wins. No loop commit falls back to `archive <name>`, else the merge base, as today.
**Considered:** adding the two self-review subjects to the existing newest-match list — leaves the loop's fix commits, which carry no marker subject, above the base, so they still fold into the implementation commit; HEAD whenever a self-review plan commit exists — hides a fix run's work committed after an interrupted loop, which would then skip the gate and the reshape; reading the loop's `tasks.md` `**Commit:**` subjects to recognise fix commits — a plan parse in a git-log walk, for what the fix-run planning commit already bounds.

### The test

**ID:** prepare-case-test
**Status:** active
**Chosen:** one case, `archived_rerun_self_review`, in `stats/internal/guard/integrate_change_prepare_test.go` beside `archived_rerun`, asserting the `ARCHIVED:` line before and after a fix-run plan commit.
**Considered:** `integrate_change_test.go`, as the dispatch named — it holds the happy path; the archived re-run cases live in the prepare file.

### Kan-1021's history

**ID:** no-history-repair
**Status:** active
**Chosen:** no repair task — the misfolded commits sit on the running change's own branch, which its own run 1 owns and rewrites; nothing in the store or on `<agents-base>` is wrong.
**Considered:** rewriting kan-1021's branch from this loop — a force-rewrite of another run's branch mid-run.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — the function decides what integrate run 1 reshapes and so whether commits are lost; the walk's state needs a careful read.
**Considered:** `flow-low` — the diff is small and the case pins it, but a wrong base silently squashes commits.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — the plan holds one task.
**Considered:** none — a single task has no split.

## Open questions
