## Context

Found by the self-review pass of `/flow` change `kan-1015-templates-step-1-fe-templates-room-search-gym`
(`<name>` below), run in-integrate. The change's canonical repository is gymie, so the pipeline
fixes go on `self-review-<name>` in the agents repository, based at `origin/main` `0f456593`; the
repairs and the product fix go on `spectre/<name>` in the change's own gymie worktrees. Excluded:
the `big` finding — desktop UI tests do not run on the web composition setup, about 400 sites —
which the parent offers for filing.

The change's worktrees: `/Users/tweety53/Projects/gymie-worktrees/<name>` (backend, canonical,
holds `.superpowers/sdd/` with the panel and implementer reports),
`/Users/tweety53/Projects/gymie-frontend-worktrees/<name>`,
`/Users/tweety53/Projects/gymie-playwright-worktrees/<name>`; the run's session transcript is
`/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/c483d41b-ee52-4470-baed-b83e5f43d1f0.jsonl`.

**F1 — panel findings never recorded.** `flow record findings -change <name>` answers `[]`, and
`check-panel-findings-closed.sh` names five reports whose `Critical:`/`Important:` lines count
findings the store does not hold: round 0 failure-modes 3, mutation 1, primary 2, principles 6,
round 2 mutation 1 — 13 rows.
<!-- measured: check-panel-findings-closed.sh <backend wt> <name> and flow record findings -change <name> @ backend fe872cce -->
`check-unfinished-work.sh`, the integrate gate, counts only open rows (`uwOpenFindings`), so an
empty store reads as nothing open; the report-count comparison lives in
`stats/internal/guard/panelfindingsclosed.go` alone.

**F2 — dispatches with no ledger row.** The change's dispatch rows stop at
`task-16-implementer` (aborted); reports with no row: gymie-playwright
`implementer-report-9/11/13/15` (Task 16 and its re-captures), gymie-frontend
`implementer-report-10/12/14` (Tasks 17, 18, 19), gymie `implementer-report-16`–`21` (Task 20
twice, Tasks 21–24) and `fix-report-1`/`fix-report-2` (panel fix rounds 1 and 2).
<!-- measured: flow record dispatches -change <name> and the reports' headers @ backend fe872cce -->
`hooks/enforce-agent-baseline.py` checks a dispatch's `description` names a row, never that the row
was begun.

**F3 — appended tasks skipped the plan guards.** On the archived `tasks.md`,
`check-plan-shape.sh` exits 1 (F11 pairs over tasks 15–24, and `**Tests:**` declaring nothing on
tasks 21–23) and `check-task-build-green.sh` exits 1 (`**Build:** pending` on tasks 21–24).
<!-- measured: check-plan-shape.sh and check-task-build-green.sh on spectre/changes/archive/<name>/tasks.md @ backend fe872cce → both exit 1 -->
`check-planning-commit-location.sh`, run behind every planning commit, refuses a ticked append
but runs neither guard.

**F4 — F11 reads only direct After edges.** `fileOwnership` (`stats/internal/guard/planshape.go`)
flags a shared-file pair unless one task's `**After:**` names the other directly; task 17
(`After: Task 15, 16`) is flagged against task 3 though 15 names 3. With reachability, the only
After hit left on the archived plan is task 15 against task 9 (`TemplatesRoomUiTest.kt`), which no
path joins.
<!-- measured: the After fields of spectre/changes/archive/<name>/tasks.md traced by hand @ backend fe872cce -->

**F5 — PIXEL PROBE accepts geometry.** `skills/flow/review-panel.md`'s paragraph reads "against
the actual rendered pixels or geometry"; Task 24's bounds-only geometry test passed on a fill
only the render showed. The paragraph's own fallback sentence (pixels cannot be rendered → assert
the computed geometry and say why) stays.

**F6 — base staleness found only at verify.** `check-base-behind.sh` runs on the first
`flow.verify` alone (`skills/flow/verify-and-handoff.md`); feature/templates sat 21 commits behind
develop through implement, and the sync at verify exposed a Flyway V45 clash.

**F7 — pre-existing attribution with no merge-base capture.** `skills/flow/visual-verify.md`
tells the verifier to start, stop and restart nothing, while step 10 of
`skills/flow/visual-verify-verifier.md` requires a merge-base capture on its own stack; the
kan-466 weight-tab line of `verify-report-visual-verify-rest-2-gymie.md` reads "Attributed as
pre-existing from the code path only: the merge-base capture was not run because the dispatch
forbids starting a stack". `check-visual-verification.sh` validates `.flow/project.md` only; the
report guard is `check-verify-report.sh` (`stats/internal/guard/verifyreport.go`), so the refusal
goes there.

**F8 — an appended visual-fix task closed without its capture.** Tasks 21–24 each closed on
desktop UI tests; their capture ran only in the next verifier dispatch.

**F9 — measurement on a composite.** The implementer brief in **The loop**
(`skills/flow/verify-fix-loop.md`) names the measurements but not the images they run on.

**F10 — untrailered commits on a change branch.** `hooks/protect-main-checkout.py` judges a
commit's trailer only when the same command pushes to `spectre/`; the integrate gate later listed
nine untrailered commits on the change's branches.

**F11 — `-rest` verifier keys.** The run's rows include `visual-verify-rest-gymie` and
`visual-verify-rest-2-gymie`; `skills/flow/visual-verify.md` defines exactly that shape (`-rest`
before the repository suffix, `-rest-<n>` from 2), and the hook's `ROW_KEY` already admits it.
The closed list row in `skills/flow/implement.md` omits both suffixes: the list is the wrong side.
<!-- measured: python3 import of hooks/enforce-agent-baseline.py, ROW_DESCRIPTION.match on 'visual-verify-rest-2-gymie x' → True @ agents 0f456593 -->

**F12 — the band read on scrim frames.** The narrative records the band scanner reading
scrim-dimmed frames as one band, and step 10's matrices measured by hand with
`scratchpad/vv/elems.py`. That script finds the surface top by walking column x=3 up from the
bottom while the pixel's mean exceeds 200, segments ink rows against the surface's modal colour,
and pairs the runs in order; Task 12 ports it.

**F13 — pass-log verdict is free text.** `RenderPanel` (`stats/internal/records/render.go`) writes
each pass row's note verbatim; round 1's note says "verdict fix — 4 Important mutation survivors"
while the store held no finding row at all.

**F14 — no round base on a verifier row.** A scoped re-run and the loop's closing unscoped run
read alike in the store, so `check-visual-verify-dispatched.sh` cannot tell that the closing run
never happened.

**F15 — a scoped re-run ran the whole suite.** `node scripts/specs-for-diff.mjs <frontend wt>
c75529d2` prints two specs (`fidelity/full-app-suite.spec.ts`, `fidelity/templates.spec.ts`);
`verify-report-visual-verify-5-gymie.md` reads "95 mapped specs, 443 passed, 2 failed".
<!-- measured: node scripts/specs-for-diff.mjs <frontend wt> c75529d2 in the playwright worktree → 2 lines; grep '^- verify:' verify-report-visual-verify-5-gymie.md @ backend fe872cce -->

**Product — kan-466 weight-tab button.** The change's playwright commit `6554b4d` re-baselined
`kan-466-weigh-in-history.spec.ts-snapshots/kan466-weight-tab-darwin.png` with the "Log a weigh-in"
`BroadsheetAccentOutlineButton` (`BodyScreen.kt`, `BodyEmptyHistory`) drawn tinted, fill
`224,234,237`, ink `81,207,255`. No frontend file of the change is on the Body render path; the
frontend merge base is `4f8f8344`. The cause is not yet traced: Task 16 opens on the trace.

**Narrative.** The archived `narrative.md` says Material3's 48dp
`minimumInteractiveComponentSize` "is enforced on web but not in desktop tests". The pass found
`LocalMinimumInteractiveComponentSize` defaulting to 48dp in commonMain; the real miss was Task
24's bounds-only test.

## Decisions

### The report-count check is shared, not copied

**ID:** share-report-count
**Status:** active
**Chosen:** extract the report-count loop of `checkPanelFindingsClosed` into one function both guards call — a second caller now exists.
**Considered:** a copy in `unfinishedwork.go`, as `duplicate-the-predicate` did for the one-line open predicate — rejected: the loop is thirty lines with a glob, a regexp and a slot split, and two copies drift.

### The begin-row check asks the store through the CLI and fails open

**ID:** begin-row-via-cli
**Status:** active
**Chosen:** the hook resolves the session's change from its last `/flow` stage-begin (`/api/v1/stage-runs`, as `hooks/flow-active-change.py` does), runs `flow record dispatches -change <change> -C <payload cwd>`, and denies only when the store answered and no open row matches the description's row; any failure lets the dispatch through.
**Considered:** fail closed — rejected: this hook fails open by contract and the store lags its harvest by a cycle; reading the transcript for `dispatch begin` calls — rejected: the store is what the ledger renders from.

### Back-filled keys name their report

**ID:** backfill-keys
**Status:** active
**Chosen:** a re-dispatch of a task already holding a row is keyed `task-<n>-implementer-r<report n>`; panel fix rounds `panel-fix-<r>`.
**Considered:** `-fix-<r>` suffixes — rejected: those rounds were re-captures, not fix rounds, and the key would misstate them.

### The plan guards run on every append

**ID:** append-runs-plan-guards
**Status:** active
**Chosen:** `checkPlanningCommit` runs `checkPlanShape` and `checkTaskBuildGreen` on the working tree's `tasks.md` whenever the diff adds a task number HEAD lacks, ticked or not; either guard's exit 1 is the guard's exit 1, its exit 2 the guard's exit 2.
**Considered:** a separate call in `document-fix.md` — rejected: prose was what the appends skipped; the guard is already on every planning commit.

### F11 orders through reachability

**ID:** f11-reachability
**Status:** active
**Chosen:** a pair is ordered when either task reaches the other through declared gating After edges.
**Considered:** requiring direct edges — rejected: a chain already serialises the pair in `plan-dispatch-bundles.sh`.

### The verifier may start the merge-base stack

**ID:** merge-base-stack-allowed
**Status:** active
**Chosen:** the dispatch's "starting, stopping and restarting nothing" carves out the merge-base stack step 10 starts; the report names its evidence as `merge-base baseline:` or `merge-base capture:`, and `check-verify-report.sh` refuses a `pre-existing` or `present at the merge base` line that names neither.
**Considered:** the parent runs the merge-base capture — rejected: step 10 already assigns it to the verifier, and the measurement must follow it.

### The attribution refusal lives in the report guard

**ID:** attribution-in-verify-report
**Status:** active
**Chosen:** `check-verify-report.sh` (`verifyreport.go`).
**Considered:** `check-visual-verification.sh`, as the finding named — rejected: that guard validates `.flow/project.md`'s declaration and never reads a report.

### The closed list is the wrong side of F11

**ID:** rest-keys-on-closed-list
**Status:** active
**Chosen:** add `-rest` and `-rest-<n>` to the verifier row of the closed list in `skills/flow/implement.md`, and pin both keys admitted in the hook's test.
**Considered:** refusing the keys in the hook — rejected: `skills/flow/visual-verify.md` defines them; tightening `ROW_KEY` to the closed suffixes cannot tell a repository suffix from an off-list one, since repository names are free text.

### The trailer check covers every commit on a change branch

**ID:** spectre-branch-commit-trailer
**Status:** active
**Chosen:** a `git commit` the hook can read, in a worktree whose branch is `spectre/<name>`, is denied when `commit_lacks_trailer` holds — the same predicate the commit-then-push check uses.
**Considered:** leaving it to the push check — rejected: a commit made with no push in the command is judged only at integrate.

### `elements` is opt-in

**ID:** elements-opt-in
**Status:** active
**Chosen:** `--props elements` is accepted and is not in the default property set; the tool emits deltas and flags nothing, as for every other property.
**Considered:** adding it to the default — rejected: it changes every existing default-props output; emitting elems.py's `departure` verdicts — rejected: thresholds are the caller's.

### The verdict line is derived

**ID:** derived-verdict
**Status:** active
**Chosen:** each pass-log round renders `- verdict: clean` or `- verdict: fix (<c> Critical, <i> Important)` from that round's finding rows, any status, after the round's notes.
**Considered:** rewriting the notes — rejected: they are the parent's record and stay verbatim.

### The round base is a dispatch column

**ID:** round-base-column
**Status:** active
**Chosen:** a nullable `round_base` column on `dispatches`, set by `flow record dispatch begin -round-base <sha>`, answered as `roundBase`; `check-visual-verify-dispatched.sh` refuses a worktree whose newest completed qualifying verifier row carries one.
**Considered:** inferring scope from the key's `-fix-<k>` suffix — rejected: this run's re-runs were keyed `visual-verify-<n>-gymie`, and a key states no base.

### The scoped-spec count is the parent's

**ID:** scoped-specs-count
**Status:** active
**Chosen:** the parent runs `specs` with the round base, names the count in the prompt, and passes it as `check-verify-report.sh`'s optional third argument.
**Considered:** the guard runs `specs` itself — rejected: guards never execute `.flow/project.md` commands.

### The product fix opens on the trace

**ID:** product-trace-first
**Status:** active
**Chosen:** Task 16 captures the weight tab at the merge base and at the change's HEAD before any edit, fixes the cause where the trace puts it, and restores the baseline `6554b4d^` carries when the fixed render matches it.
**Considered:** restoring the old baseline alone — rejected: a baseline restored over an unfixed cause fails the next run.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — two hooks that gate every dispatch and every change-branch commit, a store migration, and four guards the integrate gate runs.
**Considered:** `flow-low` — every task is small and test-pinned, but a fail-open slip in a hook or a wrong exit in a gate reaches every later run.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** six fixer dispatches in series, the tasks numbered in that order — g1 Tasks 1–4 (prose, `dispatchparagraphs.go`); g2 Tasks 5–7 (plan and gate guards, with the archived-plan and store repairs); g3 Tasks 8–9 (records render and the round-base column); g4 Tasks 10–12 (hooks, the back-fill); g5 Tasks 13–15 (verifier report guard, `measure-visual-properties`, verifier prose); g6 Task 16 (gymie product fix). The narrative correction rides Task 6's archived-record repair commit: a task whose files all lie under `spectre/changes/` is refused by `check-plan-shape.sh` F14, and both edits repair the same archived change folder.
**Considered:** one fixer — rejected: sixteen tasks across Go, Python, prose and three gymie repositories outgrow one context; one per task — rejected: tasks sharing a file would serialise anyway.

## Open questions

## Live check

Run against the live flowd on `127.0.0.1:4173` and the change's real rows, before the loop lands:

- `check-panel-findings-closed.sh <backend wt> <name>` — before: exit 1 naming five reports;
  after Task 1's repair: exit 0. `flow record findings -change <name> | jq length` — before 0,
  after at least 13. Not working: still 0, or the guard still naming a report.
- `flow record dispatches -change <name>` — before: no row keyed `task-17-implementer` through
  `task-24-implementer`, `task-16-implementer-r<n>`, `task-20-implementer-r<n>`, `panel-fix-1` or
  `panel-fix-2`; after Task 2 all fifteen back-filled keys appear, each with a non-empty `outcome`
  and `agentId`. Not working: a key missing, or an `agentId` empty after the next harvest cycle.
- The worktree's own `scripts/check-unfinished-work.sh` on the backend worktree — before Task 1's
  repair it names the five reports; after, that reason is gone. Not working: the gate silent
  before the repair.
- The worktree's own `scripts/check-plan-shape.sh` and `scripts/check-task-build-green.sh` on the
  archived `tasks.md` — before both exit 1; after Tasks 3 and 4, both exit 0. Not working: any
  line printed.
- Task 14's migration runs against the test store (`go test ./internal/store/`), never the
  protected flowd; the live column appears only once the loop lands and flowd restarts.
