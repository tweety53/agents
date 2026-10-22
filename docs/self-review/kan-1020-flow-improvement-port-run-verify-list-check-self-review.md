# Self-review report for kan-1020-flow-improvement-port-run-verify-list-check

**Deferred:** reasoning pass run on Opus 5.5 from the in-run bundle
**Rating:** 4 — good

The pass covers what the bundle holds plus the commits and files it names, and nothing beyond them.

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** The Correction rule never said `check-plan-provenance.sh` reads a Correction; task 3's bare numbers were refused only at verify — fixed: 853ac4a5
- **[flow-fix]** `pipeline.md` exceeds the Bash output cap, so a `cat` load truncates the contract; SKILL.md now says load through the file-read tool — fixed: 43daaca3
- **[flow-fix]** Task 2's test course change (PNG mtime an hour back) had no Correction paragraph; record repaired — fixed: 88385198
- **[flow-fix]** zsh pathspec and `git add` after `git rm` friction came from the session's own improvised commands, recovered in-run — noted
- **[flow-fix]** The port traced against the bash at 8c54d6d3 (parsing, truncation, exit codes, skip order, PWD, missing PNG) shows no parity defect — noted

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** Inline task rows read about 2.2–2.4M cached tokens each, the planned inline route; the panel's cost was small — noted

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** Break-and-prove caught two plan mutations the old harness let survive; the byte-parity oracle surfaced Go exec dropping the logical `PWD` — noted

## What could be automated or moved to a script — `flow-automation`

_none — this angle produced no findings._

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** Reviewer rows reading `output not measured` is documented harvester behaviour — noted

## What can be sped up — `flow-speed`

- **[flow-speed]** Tasks ran serially as the Baseline chain requires; verify's one lint re-run is removed by 853ac4a5 — noted

## What would make a good flow-eval case — `flow-eval`

- **[flow-eval]** impl case: task 3, port run-verify-list.sh to flow-guard on rgtRun (the operator chose not to mine it) — noted
  - Task: tasks.md task 3
  - Code: `stats/internal/guard/runverifylist.go:245`, `:265`, commit 3b79a6e2
  - Why: byte-exact port with hidden traps — PWD loss under exec, list-order reporting, a mutation the given harness cannot kill
- **[flow-eval]** impl case: task 2, port rebuild-full-app-suite-zip.sh to flow-guard (the operator chose not to mine it) — noted
  - Task: tasks.md task 2
  - Code: `stats/internal/guard/rebuildfullappsuitezip.go:72`, `:89`, commit 3b79a6e2
  - Why: an `archive/zip` rewrite would change the bytes; the identity test needs the mtime outside zip's 2-second granularity
