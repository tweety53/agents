# Self-review report for kan-1067-flow-allowed-pairs-experiments-stored-in-flowd

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle
**Rating:** not asked — the operator suspended the rating ask

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** `check-task-records` check 2 judged only the commit carrying a task's subject while `check-task-commit-fields` joins `Task-Id:` fix commits, so declared files landed by review fixes were refused at verify; it now joins a task's `Task-Id:` fix commits, scoped to the plan that declares their files — fixed: 6b77d840
- **[flow-fix]** `close-task` ran `check-plan-shape` only for appended tasks, so a widened planned task (a test path under `Tests: none`) surfaced only at verify; it now runs on every close — fixed: ab9f1149
- **[flow-fix]** A verify re-run keyed outside the contract (`verify-fix-1b`) formed its own `check-verify-green` group and stayed NOT-GREEN; `run-verify-list` now refuses such a key before recording or running anything — fixed: 2686b9a7
- **[flow-fix]** The mutation throwaway copy could not compile the `internal/api` tests, which imported the make-copied embed; the tests read the policy table from its source — fixed: ede25ac8
- **[flow-fix]** `live-store-copy.sh` refused a hyphenated name such as `kan-1067`; it admits `-`, mapping it to `_` for the database and state directory — fixed: 266cdd56
- **[flow-fix]** Task verify steps hand-picked a subset of `## lint` guards, so `check-ask-silence` failures first surfaced at verify; a task editing `skills/` or `rules/` Markdown now verifies with the whole lint list, and `verify-and-handoff.md`'s **Recording.** block names the key by repository, as its prose does — fixed: 93784a21
- **[flow-fix]** The design's live-check step 4 (non-zero before figures right after opening) contradicted its own equal-length-window decision; a one-off plan slip the live check measured on a stated proxy — noted
- **[flow-fix]** `main`'s Playwright suite was already red before this change: SPA commits landed outside `/flow` without updating their baselines, and this run repaired them in its full-suite fix — noted
- **[flow-fix]** An implementer committed the visual evidence directory in its task commit; the existing `check-task-commit-planning-paths` sweep caught it and the commit was split — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** Dispatch 10 (tasks 6–10 as one group) ran 30m39s and 31.7M cache-read tokens, green first time; no defect behind the size — noted
- **[flow-cost]** The full-suite fix spent 17.3M cache-read tokens repairing `main`'s pre-existing visual drift, the noted `flow-fix` item above — noted
- **[flow-cost]** The extra verify rounds trace to the plan-shape and lint-subset defects, fixed under `flow-fix` — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** Task 4's implementer measured the embed/build break and corrected the plan with a guard test instead of shipping the planned shape — noted
- **[flow-improvement]** The live check measured a stated proxy for an impossible step rather than skipping it — noted
- **[flow-improvement]** Late-fix panel rounds 3–6 each closed in under two minutes — noted
- **[flow-improvement]** A verifier brief sweeping every before/after combination closed the visual work in one round — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** Hand-copying `model-policy.md` into mutation copies is no longer needed once the `internal/api` tests read the source, fixed under `flow-fix` — noted

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

- **[flow-speed]** Visual verification took four fix rounds (~63 min); departures A and B were both in the first full capture, so two rounds were avoidable — the instruction gap is the `flow-prompt-tuning` case below — noted
- **[flow-speed]** The 21-minute verify chain traces to the plan-shape, verify-key and lint-subset defects, fixed under `flow-fix` — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** impl — Task 3 "Overlay open overrides on every allowed-pairs reader" — not mined (mining ask suspended) — noted
  - code: `stats/internal/guard/modelpolicy.go:36-62` (`policyTable`), landed in dc63c17c
  - why: a fail-open overlay with testable fallbacks and a mutation proof
- **[flow-eval]** review — F1, Important, primary, round 0 — not mined (mining ask suspended) — noted
  - location: `stats/migrations/0055_task_count_corrections.sql:5`; "Every task-count row recorded before 0055 … reads as `corrections = 0`"
  - bad: `INT NOT NULL DEFAULT 0`; fixed: nullable column, the aggregate skipping NULL — f523ea04
  - why: a default the plan dictated fabricates metric data
- **[flow-eval]** review — F4, Important, mutation, round 1 — not mined (mining ask suspended) — noted
  - location: `stats/internal/store/taskcounts.go:58`; "no test lists a legacy NULL-corrections row"
  - fixed: `TestListTaskCountsLegacyNullCorrections` — 0a476dc9
  - why: a fix introduced nullability without a test for the legacy row
- **[flow-eval]** visual — Experiments view, closed experiment — not mined (mining ask suspended) — noted
  - capture: `docs/prompt-tuning/kan-1067-flow-allowed-pairs-experiments-stored-in-flowd-1/capture.png`; mockup: `docs/prompt-tuning/kan-1067-flow-allowed-pairs-experiments-stored-in-flowd-1/mockup.png`
  - findings: an equal value drawn as an improvement (departure A); a missing before drawn as worse (departure B)
  - why: the defect is a false claim about the data, which a pixel comparison against the frame cannot see

## Instruction gaps that cost review-panel or visual-verify fix rounds — `flow-prompt-tuning`

- **[flow-prompt-tuning]** A comparison indicator is judged only on the branch the frame draws — reviewer (miss) — docs/prompt-tuning/kan-1067-flow-allowed-pairs-experiments-stored-in-flowd-1/ — noted
