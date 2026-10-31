## Context

Self-review fix change for `kan-1072-audit-whether-archived-spectre-change-artifacts`, planned on
that change's own branch `spectre/kan-1072-audit-whether-archived-spectre-change-artifacts` at
`45aac320` (the pass's `<fix-base>`), in
`/Users/tweety53/Projects/agents-worktrees/kan-1072-audit-whether-archived-spectre-change-artifacts`.
Four non-`big` findings, as `proposal.md` names them: F1 (Task 3, its history repair Task 1), F2
(Task 4), F3 (Task 5), F4 (Task 6) — and Task 2, a defect planning found that keeps the loop's verify
red (decision `acks-of-removed-change-dirs`). Each was traced on that HEAD before its fix was designed; every
question was auto-resolved to its recommended option (**Auto-resolution**,
`skills/flow-contracts/operator-prompts-auto-resolution.md`). The traces and the F4 measurement are in
`.superpowers/sdd/2026-10-31-self-review-kan-1072-audit-whether-archived-spectre-change-artifacts-design.md`
in this worktree; the F4 prototype is `.superpowers/sdd/plan-step-files-prototype.py` beside it.

The repair rewrites already-pushed history in the kan-1072 worktrees,
`/Users/tweety53/Projects/<repo>-worktrees/kan-1072-audit-whether-archived-spectre-change-artifacts`
for `<repo>` in agents, spectre, gymie, gymie-playwright, flow-eval and gymie-frontend, each on
`spectre/kan-1072-audit-whether-archived-spectre-change-artifacts`. Nothing is pushed by a fixer.

`flowd` on 4173, the `flow-postgres` container and the `flow` database are never stopped, restarted
or dropped; the live check only reads from them.

## Decisions

### A removed change directory's acknowledgements still count on its branch

**ID:** acks-of-removed-change-dirs
**Status:** active
**Chosen:** `check-verbatim-moves` also reads every `<leaf>/changes/<dir>/verbatim-moves.txt` the
branch deleted since the base and the base lacks, from the commit before the deletion
(`git log --no-renames --diff-filter=D <base>..HEAD`, then `git show <sha>^:<path>`). On this branch,
the archive commit moved kan-1072's list out of the live tree, so the guard printed a FAIL line for
every sentence kan-1072 acknowledged and `pipeline-fix.sh verify` would stay red whatever this loop
fixes; every one of those lines is in the archived list.
<!-- measured: cd <wt> && scripts/check-verbatim-moves.sh — 178 FAIL lines, 0 missing from 45aac320^:spectre/changes/kan-1072-audit-whether-archived-spectre-change-artifacts/verbatim-moves.txt @ 45aac320 -->
A list the base already held — an archive entry a tip cleanup deletes — stays unread: it
acknowledges a diff that already landed.
**Considered:** reading `archive/*/verbatim-moves.txt` from the tree again — kan-1072 removed that
read on purpose, and under run 1's drop no archive entry exists; copying the list into
`.superpowers/sdd/<change>/` by hand — a gitignored local file nothing records, and the next run hits
the same red.

### The planning commit is scoped to the change folder

**ID:** split-scopes-change-folder
**Status:** active
**Chosen:** `commit-split` keeps its `reset -q -- <leaf>/changes/` step, then stages the
implementation commit with `add -A -- . ':(exclude)<leaf>/changes/<name>'` plus one
`':(exclude)<leaf>/changes/<dir>'` per directory entry of `<worktree>/<leaf>/changes/` that
`isFixOf(<dir>, <name>)` accepts; the `link.md` re-add names `<leaf>/changes/<name>/link.md` alone
and runs when that file exists (`os.Lstat`), replacing `hasChangeLinkMD`. Every other path under
`changes/` — an archive deletion, another change's directory or `link.md` — lands in the
implementation commit; the final `add -A` leaves the plan commit only the change's own folders.
A tracked symlink at `spectre/changes` still makes the exclude-add exit 128 (spiked on git 2.50.1:
`git add -A -- . ':(exclude)spectre/changes/demo'` fails "beyond a symbolic link").
**Considered:** a third commit for foreign `changes/` paths — the reshape exists to leave two
commits; a `:(glob)<name>-fix-[0-9]*` exclude — admits `-fix-1x`, which `isFixOf` refuses; keeping
`changes/` excluded and adding foreign paths back afterwards — two more steps for the same result.

### The kan-1072 branches are re-split, not left as pushed

**ID:** repair-resplits-six-branches
**Status:** active
**Chosen:** Task 1 rewrites each branch locally to: implementation commit(s), then one
`chore(spectre): remove the change archive` commit carrying every `spectre/changes/archive/`
deletion, then (agents only) `chore(spectre): plan kan-1072-…` carrying only
`spectre/changes/kan-1072-…/`, then every commit above it replayed. spectre's implementation subject
returns to its pre-reshape `docs(spectre-planning): re-copy the planning skill — the rejection search
reads history` (its reflog). A repository whose plan commit held only archive deletions (spectre,
gymie-playwright, flow-eval, gymie-frontend) ends with no plan commit — there is no change folder to
carry. gymie is already in the target shape and is left untouched. Each tip's tree is asserted
identical to the tip before the rewrite.
**Considered:** leaving the history and fixing forward — the self-review rule requires a fix to
repair what its defect already wrote; re-running the fixed `commit-split` — it puts the deletion in
the implementation commit, not its own commit, which the repair asks for.

### Every archive-removal commit carries `Flow-Key: flow.commit-two`

**ID:** archive-removal-commit-trailer
**Status:** active
**Chosen:** each rewritten `chore(spectre): remove the change archive` commit carries
`-m "Flow-Key: flow.commit-two"`, as gymie's untouched commit and the squashed gymie-playwright and
gymie-frontend commits (reusing their existing removal commits' messages) already do — a commit
outside every task carries the stage key that made it (**Commits outside every task**,
`skills/flow-contracts/git-boundaries.md`).
**Considered:** no trailer, as the pre-reshape tip commits had — leaves the repositories
inconsistent with each other.

### The repair runs first and moves `<fix-base>`

**ID:** repair-first-new-fix-base
**Status:** active
**Chosen:** Task 1 runs before any fix commit exists, replaying every commit above the old plan
commit — the archive commit and this plan's own commit — onto the re-split history. Its report
carries `NEW-FIX-BASE: <sha>`, the rewritten image of `45aac320`, tree-identical to it; the session
uses that sha as `<fix-base>` for every later step (review diff, autosquash). The fixer never pushes:
the report lists one `PUSH:` line per rewritten repository,
`git -C <W> push --force-with-lease=spectre/kan-1072-audit-whether-archived-spectre-change-artifacts:<old-sha> origin spectre/kan-1072-audit-whether-archived-spectre-change-artifacts`,
which the session runs before the landing push (**Branch backup**, `skills/flow-contracts/git-boundaries.md`).
**Considered:** running the repair last — every fix and review fixup would then be rebased across a
rewritten base, and `git rebase --autosquash <fix-base>` would replay the whole change; a plain
`--force` push — `--force-with-lease` only, so a remote moved since is never overwritten.

### Group implementer keys are `+`-joined ids

**ID:** implementer-key-ids
**Status:** active
**Chosen:** `implement.md`'s per-group implementer key is `task-<ids>-implementer` and its record
`-task <ids>`, `<ids>` the group's task ids `+`-joined in plan order — the roles table row, the two
record calls, the key sentence after them and the prompt stub. The per-task inline row and the
`-recommit-<k>` keys stay `<n>`; the Go comments naming `task-<n>-implementer` describe a valid
one-task key and stay.

Correction (2026-10-31): review-1 (I1) showed `-task <ids>` writes a `+`-joined string into `dispatches.task_id`; the record call keeps `-task <n>`, passed only on a one-task group and omitted otherwise. The key stays `task-<ids>-implementer`.
**Considered:** teaching the hook to accept a single id for a group — the hook is right; the doc is
what is wrong.

### The bundle has its own default deadline

**ID:** bundle-default-deadline
**Status:** active
**Chosen:** `runSelfReviewBundle` sets its `-timeout` default to a `selfReviewBundleTimeout` of one
minute after `registerRecordConnFlags`, the flag's `DefValue` updated to match; `defaultTimeout`
stays as it is for every store write with a fallback.
**Considered:** passing `-timeout` at each caller (integrate's NEXT line, finish-contract-run1) — two
sites to keep in step and the bare command still fails; raising `defaultTimeout` — every `/flow`
command would wait longer on a down store.

### Step-edited paths are checked by a sub-guard

**ID:** plan-step-files-subguard
**Status:** active
**Chosen:** a new sub-guard `plan-step-files` (`stats/internal/guard/planstepfiles.go`), run by
`psCheckOne` after `plan-task-id-rule`, on the rule the brainstorm design file states (edit-verb step
labels, path-shaped tokens that exist, edit positions, `tcfCheckFiles` as the cover test). Measured
on the archived plans at `84e5e25b`: 3 findings in 2 plans, one a false positive (a fixture written
into a test's scratch repository); on the kan-1072 plan, exactly its four known misses.
<!-- measured: python3 -I .superpowers/sdd/plan-step-files-prototype.py over every spectre/changes/archive/*/tasks.md of 84e5e25b, and over the kan-1072 tasks.md of 45aac320 @ 84e5e25b -->
The finding tells the planner to add the path to `**Files:**` when the task's commit changes it,
else to reword the step — never to list a read-only path in `**Files:**`, which
`check_declared_files` would then refuse.
**Considered:** a new F-rule inside `planshape.go` — that file is already the largest guard, and
the step-scanning sub-guards live beside it as `plan-script-paths` and `plan-files-paths`; flagging
every backticked path in an edit step — over a thousand hits across most plans; edit steps with read
cues but no position rule — still hundreds.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 3 changes the commit split every integrate runs, Task 1 rewrites
pushed history in several repositories, Task 2 changes what a prose guard accepts, and Task 6 adds a
refusal to every plan's shape check.
**Considered:** `flow-low` — Tasks 4 and 5 are small, but the other four sit on every run's path.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, `-fix`, Tasks 1–6 in series in this worktree, on `opus`; Task 1 first.
**Considered:** a separate fixer for Task 1 — its commands are spelled out and it must finish before
any other commit anyway, so a second context buys nothing; one fixer per task — six small diffs.

## Open questions

## Recommended approach

Repair the branches first (decision `repair-first-new-fix-base`), make the removed change's
acknowledgements count (`acks-of-removed-change-dirs`), then scope the split
(`split-scopes-change-folder`), fix the doc key (`implementer-key-ids`), give the bundle its own
deadline (`bundle-default-deadline`) and add the step-files sub-guard (`plan-step-files-subguard`).
Each rejected alternative is the `**Considered:**` line of the named decision.

## Critical files

- `stats/internal/guard/verbatimmoves.go`, `stats/internal/guard/check_verbatim_moves_test.go`,
  `scripts/check-verbatim-moves.sh` — acknowledgement reading (Task 2).
- `stats/internal/guard/commitsplit.go` — the two-commit split (F1).
- `stats/internal/guard/commit_split_test.go` — its tests.
- `scripts/commit-split.sh` — shim whose header states the planning path.
- `skills/flow-contracts/git-boundaries-commit-chain.md` — canonical chain text, names the pathspecs.
- `skills/flow/implement.md` — implementer key rows (F2).
- `stats/cmd/flow/selfreview.go`, `stats/cmd/flow/selfreview_test.go` — the bundle CLI (F3).
- `stats/internal/guard/planstepfiles.go`, `stats/internal/guard/plan_step_files_test.go` — new
  sub-guard (F4).
- `stats/internal/guard/planshape.go` — `psCheckOne` and its header list of sub-guards.
- The kan-1072 worktrees — the repair (Task 1).

## Existing code to reuse

- `isFixOf` (`stats/internal/guard/commitarchive.go`) — the `<name>-fix-<digits>` test.
- `specRootLeaf`, `pcAbs`, `envGit` — already used by `commitSplit`.
- `tcfIterTasks`, `tcfParseTask`, `tcfUnfenced`, `tcfPlanLines`, `tcfCheckFiles`
  (`stats/internal/guard/planparse.go`, `taskcommitfields.go`) — task bodies, fields, fences and the
  commit guard's own cover test.
- `planScriptPathsCmd` (`stats/internal/guard/planscriptpaths.go`) — the sub-guard shape: argument
  handling, root from the plan path, `file:line:` output.
- `genuineDaemon`, `gitRepo`, `isolatedStateRoot` (`stats/cmd/flow` tests) — the bundle test fixture.
- `scripts/break-and-prove.sh` — mutation proofs.

## Verification

- Task 1's assertions: each rewritten tip tree-identical to its old tip; per-commit path checks.
- `cd stats && go test ./internal/guard -run 'TestCommitSplit|TestPlanStepFiles|TestCheckPlanShape' -count=1`
  and `cd stats && go test ./cmd/flow -run 'TestRunSelfReviewBundle' -count=1` pass.
- `scripts/check-plan-shape.sh` over the archived plans of `84e5e25b`, extracted to a scratch tree,
  prints exactly the measured `plan-step-files` lines (Task 6's measurement step).
- `scripts/run-verify-list.sh <worktree> lint` exits 0 after Tasks 3 and 4.
- `scripts/check-verbatim-moves.sh` exits 0 in the worktree after Task 2.

## Live check

From this worktree, after every task is committed, read-only against the dev `flowd` on 4173:
`cd stats && go build -o "$TMPDIR/flow-sr1072" ./cmd/flow`, then
`time "$TMPDIR/flow-sr1072" self-review bundle -C /Users/tweety53/Projects/agents-worktrees/kan-1072-audit-whether-archived-spectre-change-artifacts -change kan-1072-audit-whether-archived-spectre-change-artifacts > "$TMPDIR/bundle-sr1072.md"; echo "exit=$?"`.
Record the elapsed time and the exit code. Working: exit 0 and a non-empty bundle file. Not
working: exit 1 with `context deadline exceeded` after the old store deadline — the old default
still applied. Before figure: the installed `flow` with no `-timeout` on the same command (the pass
recorded a timeout on its first call).
