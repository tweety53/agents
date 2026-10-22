## Context

The eight findings of `proposal.md` are planned as one self-review change on branch
`self-review-kan-1010-calculator-macros-scale-with-bodyweight`, based at origin/main `b1b0d601`.
Findings 5, 6 and 8 live outside this repository and are fixed on kan-1010's own change branch
`spectre/kan-1010-calculator-macros-scale-with-bodyweight`, in gymie's
`/Users/tweety53/Projects/gymie-worktrees/kan-1010-calculator-macros-scale-with-bodyweight` and
gymie-playwright's
`/Users/tweety53/Projects/gymie-playwright-worktrees/kan-1010-calculator-macros-scale-with-bodyweight`.
Every fix but finding 8's was drafted and checked in a scratch clone of this worktree at `b1b0d601`
(and of the two other worktrees at their HEADs, gymie `1e060db3`, gymie-playwright `bec3d0c`); each
task's diff is a file under `patches/` beside this file, and `tasks.md` carries the measured results.

**1's measurement.** gymie's `76e18959` (`chore(spectre): plan fix 1 of kan-1010-…`) adds
`- [x] 3.` to the archived plan. Over every planning commit outside commit-split's reshaped half
(subject other than exactly `chore(spectre): plan <name>`), 6 of 893 add a task already ticked —
gymie `76e18959`, `56cbf0ab`, `64bd76d4`, `c4b8c414`, `eff176f5` and agents `a3abf28c` — each
appending a task after its work, where `skills/flow/document-fix.md` makes the fix-run planning
commit before any implementer is dispatched. The reshaped half does carry ticked appends
legitimately: gymie `890410b5` (kan-924's second run, reset to a merge base already holding the
archived plan) adds tasks 18–30 ticked.

**2's trace.** kan-1010's integrate transcript
(`~/.claude/projects/-Users-tweety53-Projects-gymie/69093694-963d-4562-8c4c-6e728958a2b6.jsonl`)
stopped at `STOP: rebase-lint — the ## lint list rewrote tracked files on …/gymie-worktrees/kan-1010-…`;
`git status` there showed only ` M …/kan-1010-…/narrative.md`, an edit made before the run.
`rebaseOntoTip` (`stats/internal/guard/baserebase.go`) sets the planning paths aside before the
rebase and restores them after, so that edit is in the tree before the lint list runs.

**3's evidence.** Both kan-1010 verifier reports read `- fingerprint: not declared` and
`- stack: already running`; gymie declares `fingerprint` as `./gradlew -q devStackFresh`.

**4's evidence.** `flow record verdicts -guard check-unfinished-work -false-positive`, run from
gymie, lists verdicts 662, 666 and 681 for kan-1010, each flagged with a reason naming
`screenshots.spec.ts:52`; 666 and 681 sit on the playwright worktree's `CLEAR` rows, the newest
when each flag was written.

**9's trace.** The creating session
(`~/.claude/projects/-Users-tweety53-Projects-gymie/2a92ac3e-0c5c-4624-ab01-ea0146376e6b.jsonl`)
launched both panel-0 agents (results at 20:22:05.6Z and 20:22:07.2Z, agents
`a198d990a8e9305e5` and `ad9089f7f5a39abe2`, descriptions opening with each row's key), then ran one
Bash call whose line is stamped 20:22:13.2Z and whose begins wrote rows 6922 and 6923 at
20:22:17.4Z and 20:22:17.5Z. The watcher reads that line as soon as it is written: the keyed stamps
find no row, the time pairing finds the two begins within reach, `StampDispatchAgent` reports no
row, and `stampDispatchAgents` marks both sides used anyway — the pair is spent before the rows
exist. A replay of exactly that sequence (Task 9's test) leaves both rows unstamped on the base and
stamps both once the pair is kept pending.

## Live check

Against the dev store (`flowd` on 127.0.0.1:4173, database `flow`):

- `flow record dispatches -C /Users/tweety53/Projects/gymie -change kan-1010-calculator-macros-scale-with-bodyweight`
  — before: rows 6922 (`panel-0-primary+principles`) and 6923 (`panel-0-test-sanity+call-efficiency`)
  carry `"agentId": null`; after Task 9's repair: `a198d990a8e9305e5` and `ad9089f7f5a39abe2`, with
  non-zero token figures. A null agent id or zero figures means the repair never reached the store.
- After the landing and `make -C /Users/tweety53/Projects/agents/stats restart`, the next panel
  round any session runs (its launches before its begins, as review-panel.md orders them) leaves no
  row of that round with an empty agent id. An empty one means the live flowd still runs the old
  build.

## Decisions

### A planning commit appending an already-ticked task is refused

**ID:** ticked-append-refused
**Status:** active
**Chosen:** `check-planning-commit-location.sh`, run as the script, also diffs the working tree's
`<name>/tasks.md` (active or archived) against HEAD's and prints `PLANNING-COMMIT-TICKED-APPEND` at
exit 1 for each added `- [x] <n>.` line whose number HEAD's file did not carry; a tasks.md HEAD
lacks is a new plan and is not read. `reshape-branch.sh` and `commit-split.sh` keep calling the
location check alone. Measured to refuse 6 of 893 mid-run planning commits, each a breach of the
commit-before-dispatch rule (**Context**), and none of the reshaped halves.
**Considered:** refusing only when no commit carries the task's `Task-Id:` — a cross-repository
change carries that commit in another repository (gymie `64bd76d4`'s task 11 is a frontend task,
`c4b8c414`'s task 7 a playwright one), so the guard would need every peer's history; exempting a
tasks.md HEAD lacks and checking everywhere — commit-split's reset to a merge base already holding
the archived plan makes `890410b5`'s 13 ticked appends look new.

### rebase_lint compares the tracked diff across the list

**ID:** rebase-lint-diff-snapshot
**Status:** active
**Chosen:** `rebase_lint` hashes `git diff HEAD` (`git hash-object --stdin`) before
`prepare-workspace.sh` and stops `rebase-lint` only when the hash after a passing list differs.
**Considered:** `git diff` as the finding words it — worktree against index misses a formatter
rewriting a staged file, `git diff HEAD` sees both; stashing the tree around the list — the rebase
core already sets the planning paths aside, and a second stash would hide from the list the very
files it should format.

### The verifier is told steps 5 and 6's results, and a bare `not declared` is refused

**ID:** verifier-gets-stack-results
**Status:** active
**Chosen:** the prompt list in `skills/flow/visual-verify.md` names steps 5 and 6's results as the
report's `stack:` and `fingerprint:` lines state them; the report template's `not declared` reads
`not declared — <step 6's reason, from the prompt>`, and `check-verify-report.sh` counts a bare
`- fingerprint: not declared` as undone.
**Considered:** the guard comparing the report against the parent's step-6 exit — a new argument
every caller of `check-verify-report.sh` would pass; the prompt list alone — nothing fails when a
later edit drops the item again.

### An earlier false-positive verdict naming a verify failure clears it

**ID:** earlier-false-positive-clears
**Status:** active
**Chosen:** signal nine of `check-unfinished-work.sh` reads the change's false-positive verdicts
(`flow record verdicts -guard check-unfinished-work -false-positive`) once, when a failure is
otherwise unaddressed, and clears the failure whose `<spec>:<line>` a reason names as a whole token
(no path character before it, no digit after). A failed read clears nothing.
**Considered:** carrying a flag forward onto the next run's rows in the store — a store write path
for a question only this guard asks; a narrative line naming the failure with no sha — the existing
`Verify failure addressed:` line requires a range commit and means fixed, not judged a flake.

### gymie's Playwright rows open on devStackFresh

**ID:** rows-gate-on-fresh-stack
**Status:** active
**Chosen:** `verify` and `capture` run `(cd "$B" && ./gradlew -q devStackFresh) && { … }`, `$B` the
same-named backend worktree under `../../gymie-worktrees/`, else `../gymie` from the main regression
checkout; a stale stack fails the row before Playwright starts. Pinned by gymie's
`scripts/test-playwright-rows-check-stack.sh`, run from `ci.yml`'s shell-guard step.
**Considered:** restarting a stale stack from the row — the verifier starts and stops nothing
(`skills/flow/visual-verify.md`); a pipeline step running `fingerprint` before every `verify` — step 6
already does at the stage's entry, and the re-runs that skipped it go through the row.

### rebaseline.mjs rewrites every capture its specs take

**ID:** rebaseline-all
**Status:** active
**Chosen:** `--update-snapshots=all`, its argument list exported as `playwrightArgs` and pinned by
`rebaseline.test.mjs`; measured: an `all` run of `fidelity/body-calculator.spec.ts` on a fresh stack
left `git status` clean, so a capture whose pixels did not move is rewritten byte-identical and the
diff-box listing still names only the PNGs that moved.
**Considered:** `changed` plus a pixel-exact re-compare — reimplements Playwright's own update.

### Finding 6 is carried by two tasks

**ID:** finding-6-two-tasks
**Status:** active
**Chosen:** Task 6 changes `rebaseline.mjs` in gymie-playwright and Task 7 the sentence describing it
in gymie's `.flow/project.md`.
**Considered:** one task — one commit cannot span two repositories.

### The sweep's flake is fixed where its trace places it

**ID:** sweep-flake-after-trace
**Status:** active
**Chosen:** Task 8 opens by reproducing `screenshots.spec.ts:52`'s timeout under load and timing every
`waitForStableCanvas` call of the sweep, then fixes the nondeterminism that trace names by a
deterministic gate in the harness (a paused or pinned clock, a wait on the app's own drawn-and-idle
state), proven by removing the gate and watching the run fail under the same load.
**Considered:** a wider test timeout or stability window — forbidden
(`~/.claude/rules/fix-determinism-at-the-source.md`); splitting the sweep without a trace — a guess
that may move the flake rather than remove it.

### A stamp that finds no row keeps its pair pending

**ID:** stamp-retry-unwritten-row
**Status:** active
**Chosen:** in `stampDispatchAgents`, a time-paired stamp the store reports as not stamped returns
its begin and launch to the pending lists at the end of the pass; the next batch with an event of
its own tries again, its keyed stamp first. Rows 6922 and 6923 are repaired with the existing
`flowd restamp-dispatch <id> <agent transcript>`, which names the agent from the file and sets its
figures.
**Considered:** a new repair command — `restamp-dispatch` with an agent transcript already sets the
id; running the stamp pass on a batch with no event of its own — a store write per batch for
pending launches that may never get a row, where the session's next call comes soon enough.

### Patches live beside the plan

**ID:** patches-beside-plan
**Status:** active
**Chosen:** each drafted task's diff is `patches/task-<n>.diff` in this change folder, applied with
`git apply` from the task's worktree root.
**Considered:** inline diffs in `tasks.md` as kan-1003's plan did — 836 lines of diff would bury
the steps. <!-- measured: cat patches/*.diff | wc -l @ b1b0d601 → 836 -->

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-medium — the change moves the harvester's stamp pairing, two guards' verdicts and a
landing-path stop in `integrate-change.sh`, with a live store repair.
**Considered:** flow-low — fits prose-only changes, not new verdicts on landing-path guards.

### Fixer groups

**ID:** fixer-groups
**Status:** active
**Chosen:** four fixer dispatches in series: g1 Tasks 1, 2, 3, 4 (this repository's guards, script
and prose — no file shared with another group); g2 Task 9 (the harvester and the live repair);
g3 Tasks 5, 6, 7 (gymie's and gymie-playwright's tooling, Task 7 after Task 5 on the same file);
g4 Task 8 (the flake's trace, which needs the dev stack under load and a fresh context).
**Considered:** one dispatch for all nine — Task 8's load runs and Task 9's live repair each want a
fresh context; a dispatch per task — nine cold starts for four patch-sized tasks.

## Open questions
