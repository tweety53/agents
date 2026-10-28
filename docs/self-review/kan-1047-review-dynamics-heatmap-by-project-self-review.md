# Self-review report for kan-1047-review-dynamics-heatmap-by-project

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** A missed fix-mutation row could never be recorded once its round's re-review began, deadlocking the panel close — fixed in-run as pipeline-fix 40 — fixed: 69938fe7
- **[flow-fix]** The whole-suite visual `verify` before handoff was unenforced, so a base spec asserting the old 8-week default surfaced only at a visual-verify dispatch. `run-verify-list.sh`'s `test` list now ends with the visual `verify` on a UI change, run from the change's own worktree of a regression checkout — fixed: a4f3e724
- **[flow-fix]** The plan had no capture task for the new bar states a prototype drew, which cost visual rounds 1 and 2. `check-plan-shape.sh` F22 now refuses a prototype-specified UI plan with no capture task — fixed: 316e3594
- **[flow-fix]** The prototype file was a caption over a JPEG of the live page, so the frame comparison never paired. The prototype rule now says the file draws the page alone — fixed: 36913447
- **[flow-fix]** Panel reproducers were broken as written (an EXIT trap restoring a relative path after `cd`, nested quotes in `sed`). The mutation-testing brief now forbids `cd` and nested quoting — fixed: 38b88362
- **[flow-fix]** Round 6's mutation reproducer left a live mutant in the canonical worktree: `runReproducer` restored only paths already dirty before the run, so a reproducer whose own restore failed left a clean file mutated. Every tracked path a run changed is now restored from HEAD — fixed: 70a6ab39
- **[flow-fix]** A re-review slot cited absolute paths and was sent back once with `-cite` — the guard worked — noted
- **[flow-fix]** `live-store-copy.sh` refused while an earlier copy's directory existed; its refusal names the cure (`down` first) — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** Rounds 3 and 4 changed only capture specs, PNGs and the UI-test seed yet ran both always-on extras (~0.5M cache-read tokens a round, 0 findings); deliberate per the panel contract — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** Pipeline-fix 40 was found, landed on main and the run resumed within the run — noted
- **[flow-improvement]** The live check compared before and after on a store copy with a concrete failure look (gymie 67 → 30 under Big) — noted
- **[flow-improvement]** Three mid-run operator requests became their own tasks and were re-decided — noted
- **[flow-improvement]** The pass caught a product defect (y-axis labels naming values other than their tick heights) and the run fixed it as fix run 1 before landing — noted

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

- **[flow-speed]** The three serial visual rounds (~37 min) came from the capture-task and full-suite gaps, fixed under `flow-fix` above — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** impl case — Task 1, the class filter threaded through store, API and SPA; not mined (operator paused mining) — noted
- **[flow-eval]** review case — F1, a one-project fixture that cannot tell the Projects filter from its absence; not mined — noted
- **[flow-eval]** visual case — the picked capture's `1, 1, 0` axis and pointer left on Big; not mined — noted

## Instruction gaps that cost review-panel or visual-verify fix rounds — `flow-prompt-tuning`

- **[flow-prompt-tuning]** A capture written by an implementer is taken before its state or with the pointer on a control — implementer — docs/prompt-tuning/kan-1047-review-dynamics-heatmap-by-project-1/ — noted
- **[flow-prompt-tuning]** The verifier's late miss of the unparked pointer on the filters capture was reviewer variance, and parking it left that baseline unchanged — disproved as a visible defect — noted
- **[flow-prompt-tuning]** The F1–F3 class (fixture data that cannot separate a dropped clause) is already covered by pending cases kan-1032 and kan-1046 — noted
