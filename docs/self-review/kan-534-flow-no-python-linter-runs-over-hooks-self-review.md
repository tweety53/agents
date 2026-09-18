# kan-534-flow-no-python-linter-runs-over-hooks — self-review

**Deferred:** reasoning pass run on builtin:zai-coding-plan/GLM-5.3-Flash from docs/self-review/kan-534-flow-no-python-linter-runs-over-hooks-context.md
**Rating:** not collected — the pass filed without the operator prompt, at the operator's instruction

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** `run-reproducer.sh`'s empty-array expansion crashes under macOS bash 3.2, the shell floor this repository's guards deliberately target, and the run worked around it with homebrew bash 5.3 and an explicit PATH; the script should run on the floor shell with the expansion guarded — filed: KAN-586
- **[flow-fix]** the session also fought a broken machine PATH (a trailing `:/`) that made previously-unhashed commands unresolvable mid-run; recorded by the run itself as an environment fact, not a branch fact, with no pipeline change that fixes it — declined

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** the fix round re-ran all three panel slots to catch two minor fix-introduced defects; the rerun-policy scope is kan-564's own pending change, so nothing filed here — declined

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** both fix-round fixes were verified by recorded mutation probes (temp-file census 2-orphaned-pre/0-post; scan-set census 31-pre/17-post from a foreign cwd), evidence the mechanism moved rather than the test; reproduce by landing every guard fix with a mutation reproducer that fails pre-fix and passes post-fix — filed: KAN-587
- **[flow-improvement]** the targeted round-1 re-runs caught the two defects the fix diff itself introduced; already filed from the kan-512 pass as the re-run-after-every-fix discipline — declined (duplicate of KAN-522)

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** all five dispatch rows read `Tokens: not measured`; duplicate of KAN-525 from the kan-512 pass — declined
