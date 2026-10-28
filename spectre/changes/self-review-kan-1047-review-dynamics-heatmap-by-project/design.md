## Context

Self-review fixes for kan-1047, planned on the change's own branch
`spectre/kan-1047-review-dynamics-heatmap-by-project` in its worktree
`/Users/tweety53/Projects/agents-worktrees/kan-1047-review-dynamics-heatmap-by-project`. Five
findings, each carried by one task; none is `big`. Every question was auto-resolved to its
recommended option.

**Finding 5 traced (the live mutant).** The round-6 `mutation+failure-modes` brief, rendered by
`scripts/render-slot-prompt.sh … -diff late-fix`, did name a throwaway copy:
"Mutate and run the tests only in your throwaway copies, one per repository —
`/Users/tweety53/Projects/agents-worktrees/kan-1047-review-dynamics-heatmap-by-project-mutation-6`".
The parent creates that copy with `throwaway-worktree.sh create` (**The throwaway worktree**,
`skills/flow/review-panel-optional-slots.md`); nothing refuses a mutation dispatch without it, but
the copy did exist — its directory still holds a residue dated before the round's reproducers. The
mutant left on the canonical tree is exactly `R-6-mutation-3.sh`'s edit, and that reproducer's tree
is `$PWD`, the worktree `run-reproducer.sh` passes — the canonical one (**The mutation-testing
brief**, `skills/flow/review-panel.md`). Every reproducer run — the parent's, the exit-contract
guard's (`stats/internal/guard/panelexitcontract.go`) and `prove-reproducer`'s — goes through
`runReproducer` (`stats/internal/guard/runreproducer.go`), which snapshots and restores only the
tracked paths that differed from HEAD **before** the run (`rrSnapshotDirty`/`rrRestoreDirty`). A
clean source file a reproducer mutates and fails to put back — its EXIT trap restoring a relative
path after a `cd` (finding 4), or the trap never firing because the bound SIGKILLs the shell — is
left mutated: exit 5 or exit 3 report the changed tree and restore nothing. So the root cause is
`runReproducer`, not the brief or the throwaway copy, and one fix there covers every caller.

**Finding 3 confirmed.** The head of the prototype file under `.superpowers/sdd/` is a `<p>`
caption followed by a base64 JPEG `<img>` of the live page.

**Finding 2 against kan-1047's history.** The plan as committed held three tasks touching
`stats/web/src/**` and none naming a file under `stats/web/tests/visual`; the capture tasks were
appended later as in-run fixes. F22 runs at plan time, so it would have refused that plan.

## Decisions

### Restore every tracked path a reproducer run changed

**ID:** reproducer-restores-clean-paths
**Status:** active
**Chosen:** `rrRestoreDirty` also puts back, from HEAD, every tracked path that differs from HEAD
after the run and was not in the before-snapshot, naming each `run-reproducer: restored <path>`;
the verdict paths stay as they are (a restored path still makes a completed run exit 5, a timed
out one still exits 3) — the tree is back whatever the verdict.
**Considered:** render-slot-prompt refusing when the throwaway copy is missing — the copy existed,
so it would not have caught this; a brief sentence alone — the parent and the exit-contract guard
run reproducers too, and a SIGKILLed trap can never restore; removing new untracked files as well —
a leftover untracked file never changes what a test reads from tracked source, and the exit-5 report
still names it.

### Reproducers never cd and never nest shell quoting

**ID:** reproducer-no-cd-no-nested-quotes
**Status:** active
**Chosen:** one sentence in **The mutation-testing brief** (`skills/flow/review-panel.md`): no `cd`
(`go -C`, `make -C`, `npm --prefix`), the edit through a patch file or a fixed-string replacement,
never nested shell quoting. The brief is rendered verbatim by `render-slot-prompt.sh`, so no
renderer change.
**Considered:** a reproducer lint guard — a new guard for what one sentence plus
`reproducer-restores-clean-paths` already contains.

### The prototype file draws the page alone

**ID:** prototype-draws-page-alone
**Status:** active
**Chosen:** one sentence in the prototype rule of `skills/flow/brainstorm-planner.md`: the file
holds the proposed page alone; notes to the operator go in the message, since the verifier renders
the whole file as the capture's frame.
**Considered:** cropping the caption in the verifier — the verifier cannot tell a caption from the
page.

### Plan-shape F22 — a prototype-specified UI change plans its capture

**ID:** plan-shape-prototype-capture
**Status:** active
**Chosen:** `check-plan-shape` F22: when the sibling `design.md` cites a `.superpowers/sdd/*.html`
file, the project root (`git rev-parse --show-toplevel` of the plan's directory) declares a
`## visual verification` section with no `regression checkout`, some task's `**Files:**` path
matches a `ui paths` glob (`visualTrigger`, `stats/internal/guard/visualtrigger.go`), and no task's
`**Files:**` path lies under `screenshots`, the plan is refused at the first such task's title line.
A plan with any of those conditions unmet is not checked; a declared `regression checkout` is not
checked, since its captures live in another repository.
**Considered:** requiring one capture per interactive state — the guard cannot read states from the
prototype; a prose-only rule — rule `skills/flow/brainstorm-planner.md` already said it and kan-1047
missed it.

### run-verify-list ends a test list with the visual verify

**ID:** verify-list-runs-visual
**Status:** active
**Chosen:** `run-verify-list.sh <worktree> test`, after every other command has exited, runs the
`## visual verification` `verify` row once in the foreground, `<specs>` replaced by nothing, from
the declared `regression checkout` or else `<worktree>`, when all hold: the row is declared;
`<worktree>/.superpowers/sdd/decision.json` does not record `visual.verify` `skipped`; and
`git diff --name-only <merge-base>..HEAD` touches a `ui paths` glob (`visualTrigger`), the merge base
being `git merge-base HEAD <resolveRemoteBase(resolve-base-branch)>` — when the base does not
resolve, it runs (running more is the safe side). It reports like any other command, counts toward
the exit code and the green sha, and is skipped with the rest when HEAD is the green sha. The
parent's prose in `skills/flow/implement.md` keeps the `--repo` case of another worktree, the stack
freshness check and the `setup` row, and drops the rest of the hand-run.
**Considered:** a separate guard checking the run happened — a second mechanism for what running it
makes impossible to skip; running it concurrently with the list — `make -s web-dist` and the
verify's `npm run build` write the same `dist/`; a new merge-base argument — every caller would
change, where the base branch already resolves in-process.

### First review effort and fixer grouping

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` for the first review — Tasks 1 and 5 change `run-reproducer.sh` and
`run-verify-list.sh`, which every run's reproducers and verify lists go through. One fixer group
(key `self-review-kan-1047-review-dynamics-heatmap-by-project-fix`) carrying Tasks 1–5 in order:
two are one-sentence doc edits and the three Go tasks are independent and small.
**Considered:** `flow-low` — too light for two guards on every run's path; one fixer per task — five
dispatches where one context serves all.

## Open questions

## Recommended approach

Fix each defect where every caller routes through it: `runReproducer` for the mutant
(`reproducer-restores-clean-paths`), `runVerifyList` for the unrun visual verify
(`verify-list-runs-visual`), `check-plan-shape` for the missing capture task
(`plan-shape-prototype-capture`), and one sentence each where the authoring rule lives
(`reproducer-no-cd-no-nested-quotes`, `prototype-draws-page-alone`). Each Go fix carries one test
that fails without it.

## Critical files

- `stats/internal/guard/runreproducer.go`, `stats/internal/guard/runreproducer_test.go`,
  `scripts/run-reproducer.sh` — reproducer runner, its test, its shim header.
- `skills/flow/review-panel.md` — the mutation-testing brief.
- `skills/flow/brainstorm-planner.md` — the prototype rule.
- `stats/internal/guard/planshape.go`, `stats/internal/guard/check_plan_shape_test.go` — plan-shape
  guard and test.
- `stats/internal/guard/runverifylist.go`, `stats/internal/guard/run_verify_list_test.go`,
  `scripts/run-verify-list.sh` — verify-list runner, its test, its shim header.
- `skills/flow/implement.md` — the full-suite run's visual paragraph.

## Existing code to reuse

- `rrSnapshotDirty`, `rrRestoreDirty`, `rrGitStatus` (`stats/internal/guard/runreproducer.go`).
- `visualTrigger` (`stats/internal/guard/visualtrigger.go`) — the `ui paths` match.
- `vvSectionLines`, `vtSplitCells`, `vtFoldCell` (`stats/internal/guard/visualsection.go`,
  `stats/internal/guard/specreach.go`'s `regression checkout` read) — the section's rows.
- `Registry["resolve-base-branch"]` and `resolveRemoteBase`
  (`stats/internal/guard/resolvebasebranch.go`, `stats/internal/guard/resolveremotebase.go`).
- `runVerifyList`'s own `run`, `report` and `unchanged` closures.
- `tcfParseTask`'s `files` field as F21 reads it (`stats/internal/guard/planshape.go`).

## Verification

- `cd stats && go test ./internal/guard/ -count=1 -run 'RunReproducer|RunVerifyList|CheckPlanShape'`
  passes, and each new test fails with its task's production change reverted.
- `scripts/check-plan-shape.sh` (no argument) exits 0 over every live plan in this worktree.
- The lint list of `.flow/project.md` exits clean.

## Live check

None — the fixes are guards and prose run from this checkout; no daemon, store or scheduler state
changes. Step 4's `pipeline-fix.sh verify` runs the changed `run-verify-list.sh` over this branch,
whose diff touches `stats/web/src/**`, so the visual `verify` must appear in its output.
