# kan-760-agents-port-the-flow-guard-scripts-and-their — self-review

**Deferred:** reasoning pass run on account:zai-individual-coding-plan/GLM-5.3-Flash from docs/self-review/kan-760-agents-port-the-flow-guard-scripts-and-their-context.md
**Rating:** not collected — the pass filed without the operator prompt, at the operator's instruction

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** the design's `follow-ups-at-integrate` decision never fired — no follow-up slice issue exists in Jira — so the three slowest unported harnesses the change names as the suite's new wall have no owner; file the named slices per `jira-followups.md` — filed: KAN-776
- **[flow-fix]** the bundle's held observations record the same zsh footguns (`path` clobbered twice, unquoted-variable loops silently emptied) that three earlier bundles in this batch hit, plus the compound-command refusals the hook's deny-not-rewrite design imposes; no contract states the defensive-shell discipline, so every run re-derives it — filed: KAN-777

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** the unported slow harnesses (`test-check-installed-citations.sh` 104–167s under load, `test-check-plan-provenance.sh`, `test-check-references.sh`) now bound every full-suite run; the porting slices named in the change's design own this cost, and their absence is filed under `flow-fix` this pass — declined

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** the deferred-flake class was handled by proving at the source: task 14 measured the survivor flake (17/240 under stress), diagnosed two causes, fixed both, and showed 0/480 — reproduce by never re-running past a flake; stress, diagnose, fix, re-stress — declined

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** the shim's build-from-checkout cache (`scripts/lib/flow-guard.sh`) removed the entire install-and-drift class the panel's round 0 found; reproduce by keying a derived artifact on a hash of its exact sources and building atomically into that cache — declined

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._
