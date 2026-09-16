# kan-516-log-no-flow-kan-xxx-prefix-fixes-in-store-as-if — self-review

**Deferred:** reasoning pass run on builtin:zai-coding-plan/GLM-5.3-Flash from docs/self-review/kan-516-log-no-flow-kan-xxx-prefix-fixes-in-store-as-if-context.md
**Rating:** not recorded — the filing ask went unanswered; dispositions below follow the session's earlier selections

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** `hooks/*.py` has no linter in the project's `## lint` list, so the run's one Critical defect (a `# noqa: BLE001` suppression the change itself introduced, forbidden by the Lint Fix Priority rule) was catchable only by the panel and never mechanically; add a Python lint over `hooks/` or a suppression-marker guard — filed: KAN-534
- **[flow-fix]** `check-base-moved.sh` reported `MOVED` when origin/main's only new commit was the branch's own first commit landed upstream by an unrelated session, forcing a rebase and a stash dance before re-confirming `CLEAR`; movement satisfied entirely by commits the branch already carries should read `CLEAR` — filed: KAN-535
- **[flow-fix]** 2 of 3 reviewer reproducers shipped with inverted or dead exit-code logic, and the narrative names the remedy itself: a reproducer needs the same TDD-style proof its own finding does; evidence added to the existing issue rather than a new one — filed: KAN-519

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** the round-0 panel pass read 3,789,475 cached tokens plus 208,430 cache-creation, the run's dominant cost with no named entry context for reviewers; already covered by the reviewer-entry-context issue from the kan-512 pass — filed: KAN-521

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** the principles slot caught the run's own Critical violation of an always-on rule (the forbidden noqa marker) and it was fixed in-run; panels already check diffs against the always-on rules, so reproducing it needs no change — declined

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** `setup.sh` now carries the per-hook registration check-and-warn shape twice across `install_hooks` and `install_hooks_zcode`, worth a shared helper when a third hook arrives (panel F4, deferred in-run as the tracking reminder) — filed: KAN-536

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** the hook re-parses the session's JSONL transcript on every plain prompt to find the last `/flow` mark, a derivation flowd's transcript harvest already performs; the store could answer it and the hook becomes a thin query — filed: KAN-537
