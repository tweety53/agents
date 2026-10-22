## Context

The four findings of `proposal.md` are planned as one self-review change on kan-922's own branch
`spectre/kan-922-flow-improvement-port-the-bash-python-tooling`, in its worktree
`/Users/tweety53/Projects/agents-worktrees/kan-922-flow-improvement-port-the-bash-python-tooling`,
fix-base `15542943`. No path a finding names changed between `e111a883` and the fix-base outside
kan-922's own commits, and no archived change records a rejection of any remedy below.

Evidence at `15542943`:

- F-A: `cmd_prepare` (`scripts/integrate-change.sh`) sets `ask=1` unless the drift guard printed
  `DRIFT-CLEAN`, then stops at `STOP: foreign-staged` unless `--accept-foreign`. The guard
  (`stats/internal/guard/maincheckoutdrift.go`) only reports and stays as it is. Spiked: a
  `git switch main` whose `main` is checked out in another worktree exits 128 and leaves the branch
  unchanged; a worktree's `branch.spectre/<name>.flowBase` reads from its main checkout, branch
  config being the repository's shared config; a detached checkout's `git branch --show-current`
  prints nothing. `scripts/test-integrate-change.sh`'s `refresh_on_recorded_base` puts `a`'s main
  checkout on its recorded base `feature/x` and passes `--accept-foreign` past that drift (KAN-911).
- F-B: `landSelfReviewReport` (`stats/internal/guard/landselfreviewreport.go`) refuses only
  `LAND-UNTRACED-FINDING`/`LAND-DEFERRED-FINDING`; the self-review pass reproduced
  `check-self-review-report.sh` failing over e111a883's KAN-923 report (rc=1, two violations).
  `check-self-review-report.sh` takes a directory and reads its frozen lists from it. The chain's
  add-path is always the report; its optional rm-path is the bundle.
- F-C: `requireChangeCheckout` (`stats/cmd/flow/record.go`) scans every worktree for the change;
  `inlineDirtyPaths(f.dir, f.change)` then reads `f.dir`'s checkout alone.
- F-D: `breakandprove.go` sets `cmd.Dir` to the repository root, as `scripts/break-and-prove.sh`'s
  header states; the script is right, the brief is silent.

Repair: none of the four defects persisted a wrong record, row or file. F-A and F-C cost an
operator round-trip and a refused begin; F-B's corrupted KAN-923 report is already repaired in
kan-922 (c81799f5); F-D cost a re-run.

Run-loaded prose this change rewords or adds on purpose is acknowledged in `verbatim-moves.txt`
beside this file, each line as `check-verbatim-moves.sh` printed it over the drafted edits at
plan time; a task whose wording departs from its spelled text updates that file's line by a
dated Correction in `tasks.md`, never silently.

## Live check

None — the change edits a bash orchestrator acting on git checkouts (covered by the fixture
harness), a landing guard, a client-side refusal that runs before any daemon contact, and brief
prose; no daemon, store or running service changes behaviour.

## Decisions

### Prepare switches a clean drifted checkout back

**ID:** prepare-switch-clean-drift
**Status:** active
**Chosen:** a `switch_back <main-checkout>` function in `scripts/integrate-change.sh`, called from
`cmd_prepare`'s per-checkout loop only when `check-foreign-staged.sh` exited 0 with
`STAGED-CLEAN` and `check-main-checkout-drift.sh` exited 0 with `DRIFT-BRANCH` and no
`DRIFT-DIRTY`. It returns without switching on a detached HEAD (empty `branch --show-current`),
an unreadable `refs/remotes/origin/HEAD`, or a current branch equal to
`branch.spectre/<name>.flowBase` read in any of `MAINS`; otherwise it runs
`git -C <m> switch -q <default>`, printing
`DRIFT-SWITCHED: <m> — <branch> -> <default>, <branch> kept at <short sha> (<n> commits ahead of <default>)`
on success and `DRIFT-SWITCH-REFUSED: <m> — git switch <default> failed; left on <branch>` (git's
own error passing through on stderr) on failure. The drift guard then re-runs, and its verdict —
not the first — decides `ask`. The switch runs with or without `--accept-foreign`.
**Considered:** a Go `flow-guard` subcommand — four more files (Go, test, shim, the project's Go
list) for two git calls in the orchestrator that already sequences and stops; the memory "new
guards in Go" binds new guard logic, and this is an action, not a verdict. Switching in the drift
guard — the guard reports and its verdict grammar is relied on by run 2 and the hand fallback.
`git checkout -B`/reset — would move or rewrite the drift branch's commits. Excluding only the
recorded base read from the change's own worktrees — a worktree already removed would hide its
base; the shared config read from each main checkout carries it either way. Switching a detached
HEAD — its commits sit on no branch and would survive only in the reflog.

### Land chain runs the report check

**ID:** land-runs-report-check
**Status:** active
**Chosen:** after `lsrrUntraced`, and before anything is staged, `landSelfReviewReport` runs
`bash <repo>/scripts/check-self-review-report.sh <the report's directory>` (`cmd.Dir` the
repository, its stdout and stderr relayed) when the add-path's base name ends `-self-review.md`
and `<repo>/scripts/check-self-review-report.sh` is a regular file. Exit 1 prints
`LAND-REPORT-INVALID: <add-path> fails check-self-review-report — fix the lines it names; nothing added, committed, pulled or pushed`
and any other non-zero exit
`LAND-REPORT-UNCHECKED: <add-path> — check-self-review-report could not answer (exit <rc>); nothing added, committed, pulled or pushed`,
both on stderr and both exit 5.
**Considered:** the check in-process — it reads `FLOW_GUARD_REPO_ROOT` and the angle table of the
agents checkout, not the landing repository's; a temporary directory holding the report alone —
more code, and a sibling report's violation already fails the repository's `## lint`; a project
lookup through `.flow/project.md`'s `## lint` — the script's presence is the same answer in one
stat; exit 4 shared with the untraced refusal — a caller could not tell a malformed report from an
open finding.

### Inline begin reads the change's worktree

**ID:** inline-begin-change-worktree
**Status:** active
**Chosen:** `changeCheckouts(dir, change) (main string, carrying []string, err error)` holds
`requireChangeCheckout`'s worktree scan — `main` the first `worktree` entry of
`git -C dir worktree list --porcelain`, `carrying` every entry holding the change, in list order —
and `requireChangeCheckout` calls it. The inline begin passes `inlineDirtyPaths` the first
carrying entry that is not `main`, else `f.dir`.
**Considered:** `requireChangeCheckout` returning the path — five callers change for one that
reads it; always `f.dir`'s toplevel when it carries the change — the main checkout can carry an
archived copy after a landing and misfire the same way.

### Fixer brief names the test command's directory

**ID:** brief-break-and-prove-root
**Status:** active
**Chosen:** one sentence in the fixer brief's `break-and-prove.sh` paragraph naming the
repository root and the `go -C stats test ./internal/guard` form.
**Considered:** running `<test-command>` from the caller's cwd — the script's header and
`--patch` path rest on the root; `check-dispatch-paragraphs` pins no copy of this paragraph
(the drafted edit left it exit 0), so no mirror changes.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-medium — `prepare` now moves a main checkout's HEAD on its own and a landing
guard gains a refusal; both change what the pipeline does unasked.
**Considered:** flow-low — fits prose-only or mechanical changes, not new behaviour on a shared
checkout.

### Fixer groups

**ID:** fixer-groups
**Status:** active
**Chosen:** one fixer dispatch, g1, Tasks 1, 2, 3, 4 in order — four small tasks in one
repository; Task 2 follows Task 1 for their shared `skills/flow/integrate.md`.
**Considered:** one dispatch per task — four cold starts for small tasks; a separate group for the
Go tasks — fixer groups run in series, so a second cold start buys nothing.

## Open questions
