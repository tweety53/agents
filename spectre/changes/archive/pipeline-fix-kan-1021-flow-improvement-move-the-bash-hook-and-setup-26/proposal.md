# pipeline-fix-kan-1021-flow-improvement-move-the-bash-hook-and-setup-26

## Why

`archBase` in `stats/internal/guard/integratechange.go` picks the commit integrate run 1's reshape
of an archived change starts from. It recognises the self-review report (or context bundle) commit,
else `chore(spectre): archive <name>`, and nothing of the self-review loop that runs before the
report: `chore(spectre): plan self-review-<name>`, the loop's fix commits and
`chore(spectre): archive self-review-<name>`. A run 1 resumed mid-loop takes the archive commit as
its base, folds the loop's fix commits into a new implementation commit — losing their shas — and
carries the loop's plan files under `chore(spectre): plan <name>`. Change kan-1021 hit it: its
branch's `32908088` (the self-review plan commit) was reset to `f85971db` and `5bbe4b50`
"chore(spectre): plan kan-1021-…" carried the self-review plan files.

## What changes

- An archived re-run's base is the newest commit of the self-review loop: its report or bundle
  commit, its `archive self-review-<name>` commit, or, from its `plan self-review-<name>` commit
  on, every commit above that plan up to a fix run's first `chore(spectre): plan <name>` planning
  commit. Without any loop commit the base is unchanged: `archive <name>`, else the merge base.
- `skills/flow-contracts/finish-contract-run1.md` states the new base; its rationale appendix
  records why.
- `stats/internal/guard/integrate_change_prepare_test.go` gains case
  `archived_rerun_self_review`: a resumed run 1 above the loop's plan and fix commits holds no new
  work, and a fix run's plan commit above them is new work.
