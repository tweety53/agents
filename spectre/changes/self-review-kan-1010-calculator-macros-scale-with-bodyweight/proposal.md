# self-review-kan-1010-calculator-macros-scale-with-bodyweight

## Why

kan-1010's self-review pass found eight pipeline defects, none `big`:

- **1 (flow-fix).** A fix run appended Task 3 to the archived plan already ticked (`- [x] 3.`, gymie
  `76e18959`) and skipped implement, panel and handoff; no guard refuses a planning commit whose
  `tasks.md` adds an already-ticked task.
- **2 (flow-fix).** `rebase_lint` in `scripts/integrate-change.sh` stops whenever the tree is dirty
  after the `## lint` list, without comparing to before. The rebase sets the planning paths aside
  and restores them, so an uncommitted `narrative.md` edit stopped kan-1010's integrate at
  `STOP: rebase-lint` with the list blamed.
- **3 (flow-fix).** The visual verifier reported `fingerprint: not declared` though gymie declares
  the row and the parent ran step 6 (`FRESH=0`): the dispatch prompt's list in
  `skills/flow/visual-verify.md` never passes steps 5 and 6's results, while the report template in
  `skills/flow/visual-verify-verifier.md` asks for them.
- **4 (flow-fix).** The unfinished-work gate re-flagged `screenshots.spec.ts:52` on every run: a
  false-positive verdict attaches only to that run's newest row (store verdicts 662, 666, 681).
- **5 (project tooling).** gymie's `verify` and `capture` rows run Playwright without
  `devStackFresh`, which the `## run` block requires before any Playwright run; kan-1010's tests
  ran on a stale stack.
- **6 (project tooling).** gymie-playwright's `scripts/rebaseline.mjs` runs
  `--update-snapshots=changed`, which leaves a move the change meant un-rebaselined while it stays
  inside the comparison tolerance.
- **8 (test-harness).** `screenshots.spec.ts:52` times out in `waitForStableCanvas` under
  full-suite load and passes alone; not yet traced.
- **9 (flow-stats-app).** Dispatch rows `panel-0-primary+principles` and
  `panel-0-test-sanity+call-efficiency` (ids 6922, 6923) of kan-1010 were never stamped with an
  agent id, silently.

## What changes

- `check-planning-commit-location.sh` refuses a planning commit appending an already-ticked task
  (`PLANNING-COMMIT-TICKED-APPEND`); `skills/flow/document-fix.md` cites it.
- `rebase_lint` stops only when the tracked diff changed across the lint list.
- The verifier's prompt carries steps 5 and 6's results, and `check-verify-report.sh` refuses a
  bare `fingerprint: not declared`.
- `check-unfinished-work.sh` clears a verify failure an earlier false-positive verdict of the change
  names by `<spec>:<line>`; `skills/flow/unfinished-work-gate.md` says so.
- gymie's `verify` and `capture` rows open on the backend worktree's `./gradlew -q devStackFresh`,
  pinned by `scripts/test-playwright-rows-check-stack.sh`.
- `rebaseline.mjs` runs `--update-snapshots=all`; gymie's `.flow/project.md` sentence follows.
- `screenshots.spec.ts`'s sweep settles under load by a deterministic gate found by its trace.
- The harvester keeps a pair whose stamp found no row pending for the next batch; rows 6922 and
  6923 are restamped with `flowd restamp-dispatch`.
