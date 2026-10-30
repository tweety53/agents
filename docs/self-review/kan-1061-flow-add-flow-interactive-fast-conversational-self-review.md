# Self-review report for kan-1061-flow-add-flow-interactive-fast-conversational

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** `check-plan-provenance.sh` missing from the /flow skill scripts and its shim resolving the repo root from its own directory (in-run pipeline fix) — fixed: 6226d3ed
- **[flow-fix]** A pre-existing red vitest on main read the Experiments chart before the stats fetch resolved; it now awaits `findAllByRole` — fixed: 9e893356
- **[flow-fix]** Panel rows of joined slots (`panel-N-test-sanity+call-efficiency`, `panel-N-mutation+failure-modes`) recorded a single `-slot`; `begin` now refuses a slot its key does not name, and `flowd reslot-dispatch` repaired the five kan-1061 rows — fixed: 079984de
- **[flow-fix]** The full-page/viewport capture rule and the no-duplicate-state rule reached only the verifier, so Task 9 shipped element shots and Task 15's switch made two captures byte-identical; the planner's capture-task paragraph now states both — fixed: f45bc7f0
- **[flow-fix]** `check-task-build-green.sh` not installed under the skill scripts: it is the project-configured guard, exempt in `gsExempt`, and running it from `scripts/` is correct — noted
- **[flow-fix]** Appended tasks' plan fields were corrected only after implementation; `document-fix.md` already runs the plan guards at append — noted
- **[flow-fix]** zsh does not word-split unquoted variables in ad hoc commands; no pipeline text prescribes such a command — noted
- **[flow-fix]** Task 15's three surviving mutants: MUTATION PROOF and the new-branch-test rule already cover them, and the round-3 mutation slot caught all three — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** Planning took five opus xhigh dispatches (~457k output tokens, 63M cache read), and dispatch 15 carried nine tasks over 45m50s with 51M cache read — an observation of where the cost sat — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** Before/after live check against store copies at the merge base and HEAD — noted
- **[flow-improvement]** F1 caught in round 0, all six of its fix mutations killed — noted
- **[flow-improvement]** The verifier's `identical:` check surfaced the byte-identical waiting/running captures — noted
- **[flow-improvement]** Parallel implementer wave in throwaway worktrees, cherry-picked clean — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** Dispatch hooks refused launches for a missing description opening or `Role:` line; `flow record dispatch begin` now prints both and `implement.md` says to paste them — fixed: 43fe7706

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** Hook denials left no store record; each dispatch-hook deny now writes a `flow record incident` row, failing open — fixed: 3fe2d2a1
- **[flow-stats-app]** Zero-token inline verify rows are a real measurement per `render.go`, not a missing one — noted

## What can be sped up — `flow-speed`

- **[flow-speed]** Four verifier passes trace to the capture-shape gap fixed under `flow-fix`; the hook-retry gaps are fixed under `flow-automation` — noted
- **[flow-speed]** The round-0 mutation slot ran 12m39s — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** review case — F1, Important, `stats/internal/api/request.go:222`; mining not asked this run — noted
  - Bad: `a9539509^`; fixed: `a9539509`; test `TestRecordRequestEndLandedGuard`
  - Why: a fallback for an unreachable store that also fired on a reachable store's error answer
- **[flow-eval]** impl case — Task 4, `gateLandedRequest` at `stats/internal/api/request.go:166`; mining not asked this run — noted
  - Why: a landed-end gate with several store and git edge cases under test
- **[flow-eval]** visual case — run-view capture at `f6c8cf26` against `.superpowers/sdd/interactive-frames/prototype-run-view.png`, verdict mismatch; mining not asked this run — noted
  - Why: a layout departure (column split, panel placement) a pixel diff alone under-weights

## Instruction gaps that cost fix rounds, and the generic prompt update that would have avoided them — `flow-prompt-tuning`

- **[flow-prompt-tuning]** a departure the change itself introduced is handed back as a leftover with no cause — implementer — docs/prompt-tuning/kan-1061-flow-add-flow-interactive-fast-conversational-1/ — noted
- **[flow-prompt-tuning]** a fallback for an unreachable dependency also fires when the reachable dependency answers with an error — implementer — docs/prompt-tuning/kan-1061-flow-add-flow-interactive-fast-conversational-2/ — noted
