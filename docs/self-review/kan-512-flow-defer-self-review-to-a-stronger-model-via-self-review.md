# kan-512-flow-defer-self-review-to-a-stronger-model-via — self-review

**Deferred:** reasoning pass run on builtin:zai-coding-plan/GLM-5.3-Flash from docs/self-review/kan-512-flow-defer-self-review-to-a-stronger-model-via-context.md
**Rating:** 3/5

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** `spectre link` resolves a peer's relative path against `git rev-parse --git-common-dir`, which for a worktree call always resolves to the main checkout, so linking the gymie satellite worktree to the canonical agents change could not resolve and needed a throwaway peers entry, `spectre link --force`, a hand-edited `link.md`, and a revert; peer resolution should account for the invoking worktree — filed: KAN-518
- **[flow-fix]** every round-0 reproducer was authored with the inverted exit-code convention (exit 0 on the defect's own diagnostic success rather than 0 meaning fixed) and was hand-corrected before `run-reproducer.sh`'s pre-fix/post-fix semantics worked; the reviewer prompt should state the convention — filed: KAN-519
- **[flow-fix]** run 2 found the main checkout's working tree carrying a stale staged `.flow/project.md` reverting `## self review` from `defer` to `skip` while HEAD said `defer`, and the collision took an operator ruling mid-archive; keys should resolve from HEAD or warn on working-tree divergence — filed: KAN-520

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** the round-0 combined reviewer pass (primary+simple-reviewer+principles) read 5,059,600 cached tokens plus 187,883 cache-creation, an order of magnitude above rounds 1 and 2 (759,353 and 103,139); a named entry context (diff, touched files, named contracts) would scope panel exploration — filed: KAN-521

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** the panel's delta re-runs caught a fix-of-a-fix: round 0 found four findings, round 1 found task 10's own fix incomplete (F4, fixed as task 11), round 2 confirmed clean; re-run the panel after every fix task and close only on a clean round — filed: KAN-522

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** the report-landing shell chain (branch assert, add, rm, commit, pull --rebase, push) exists as prose in `/flow-self-review` step 5 and archive.md step 9 and took two panel rounds to become correct (F3: missing branch re-assert; F4: pull/push outside the guard); extract it into one tested script both call sites invoke — filed: KAN-523
- **[flow-automation]** correcting the inverted reproducer convention was a per-reproducer hand edit with nothing mechanical catching it; `run-reproducer.sh` should refuse a reproducer whose exit status is identical pre-fix and post-fix, naming the expected convention — filed: KAN-524

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** the SDD ledger records tokens for the three reviewer dispatches only and the implementer and verifier rows read `Tokens: not measured`, the two heaviest roles; flowd could accept token reports for same-session dispatches so per-dispatch cost covers the whole run — filed: KAN-525
- **[flow-stats-app]** the bundle's four sources are gathered by a Bash script reading files that are themselves store renders, one via an absolute landing-worktree path baked into the bundle; flowd could serve the bundle directly (`flow self-review bundle -change <name>`), removing the Bash gather and its worktree-path coupling — filed: KAN-526
