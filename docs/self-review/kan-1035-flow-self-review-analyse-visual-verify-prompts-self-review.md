# Self-review report for kan-1035-flow-self-review-analyse-visual-verify-prompts

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

**Rating:** skipped — paused by the operator

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** The new angle-8 text judged visual-verify findings by a severity the verifier never assigns, while case.json requires one; a visual-verify departure is now graded Important with slot `visual-verify` — fixed: 58057ff0
- **[flow-fix]** A reviewer case targeting the verifier could never reproduce under the diff-based replay; a verifier case now carries its capture and mockup and is replayed by a verifier over them — fixed: 0bebc9b4
- **[flow-fix]** A stage mark was refused for a session token without its UTC suffix; the refusal names the required shape, so no defect — noted
- **[flow-fix]** A draft edited outside any run was caught by check-verbatim-moves, as designed — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** The two always-on extra bundles spent about 68k cache-creation tokens for zero findings on a Markdown change; by design, and the design-text gaps above are outside those slots' remit — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** The verbatim-moves guard turned an out-of-run edit into a recorded change; keep it in `## lint` — noted

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** `/flow-fast` ran its `## lint` list by hand and picked up the code-fence lines; it now runs the list through `run-verify-list.sh` as `/flow` does — fixed: 1375c91a

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

_none — this angle produced no findings._

## What would make a good flow-eval case — `flow-eval`

_none — this angle produced no findings._

## Instruction gaps that cost review-panel or visual-verify fix rounds — `flow-prompt-tuning`

- **[flow-prompt-tuning]** Suspected the bundle lacks the per-round visual-verify history angle 8 needs; disproved — tasks.md and the archived visual-verification.md carry it — noted
