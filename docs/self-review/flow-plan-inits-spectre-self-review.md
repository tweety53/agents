# flow-plan-inits-spectre — self-review

**Deferred:** reasoning pass run on builtin:zai-coding-plan/GLM-5.3-Flash from docs/self-review/flow-plan-inits-spectre-context.md
**Rating:** not recorded — the filing ask went unanswered; dispositions below follow the session's earlier selections

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** `check-task-commit-fields.sh` clobbers its own positional arguments when invoked in a loop, reporting every task as not found until rerun one call per task; fix the argument handling and pin it with a multi-task harness case — filed: KAN-528
- **[flow-fix]** the mutation slot's throwaway worktree was created without its `.superpowers/sdd` bundle and had to be re-briefed, and the round-2 mutation report was destroyed with that worktree, surviving only through the parent session's earlier read; scaffold slot worktrees and persist reviewer reports store-first — filed: KAN-529
- **[flow-fix]** `check-guard-symlinks.sh` reads an unmatchable guard basename in an invoking paragraph as prose, so `GUARD-SYMLINKS-OK` still prints and the per-skill count drops silently (panel F6, deferred in-run) — filed: KAN-530
- **[flow-fix]** dropping `-F` from `check-stage-mark-calls.sh`'s `grep -qxF` membership test survives until a metacharacter-adjacent key exists, and message-text-only mutants change no exit code, so no fixture pins either (panel F11+F12, deferred in-run) — filed: KAN-531
- **[flow-fix]** the round-1 mutation reproducers shipped with an inverted exit-code convention and needed the parent to invert their sense to the `run-reproducer.sh` contract, the third occurrence of the class; evidence added to the existing issue rather than a new one — filed: KAN-519

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** 21 of 23 dispatches read `Tokens: not measured`, leaving the run's cost picture to the final panel round alone; already covered by the inline-token-accounting issue from the kan-512 pass — filed: KAN-525

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** the mutation slot delivered 5 of the 12 findings (F4, F5, F8, F9, F12), each a survivor that passed the guard's own harness until mutated; the slot is already in the standard roster this run used, so reproducing it needs no change — declined

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** the six-basename guard list in `skills/flow-plan/SKILL.md`'s invoking paragraph is enforced by nothing, so dropping a basename only moves an informational per-skill count (panel F10, deferred in-run) — filed: KAN-532

## What could move to the Go app or its persistent storage — `flow-stats-app`

- **[flow-stats-app]** the stage-key vocabulary lives in three places (README's Level 1 table, the Go `stageKeyRE`, and the Bash guard that now reads README) and panel F7 caught the Bash and Go copies already diverging; flowd or the `flow` CLI should serve it once — filed: KAN-533
