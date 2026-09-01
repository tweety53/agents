# kan-371-myflow-rebase-onto-the-base-branch-before-the

## Why

`skills/flow/integrate.md`'s step 2 already detects when the base branch has moved and overlaps
this change's own touched paths (`check-base-moved.sh`), and stops to ask the operator. Today the
only choices are "stop, rebase by hand" or "continue anyway" — there is no in-pipeline rebase, so
every overlap costs the operator a manual detour outside the pipeline. Observed live integrating
KAN-304: `main` moved 6, then 9, commits ahead of the recorded merge base, overlapping
`skills/flow/review-panel.md` (touched by both changes), and the operator had to stop and rebase by
hand on every re-invocation.

## What changes

- The overlap prompt in `skills/flow/integrate.md`'s step 2 gains a third option: **Rebase onto
  `<base>` now, then continue**, alongside the existing Stop and Continue choices.
- On a clean rebase, the recorded merge base is updated to `origin/$BASE`'s resolved tip, and — only
  for the file(s) `check-base-moved.sh` reported as overlapping, and only where a discoverable test
  exists — a scoped re-verification runs before proceeding to the landing question. The project-wide
  `## lint`/`## test` list is not re-run; `integrate.md`'s "no verification gate" rule stays intact
  for the untouched-by-rebase case.
- On a conflicting rebase, the worktree is left mid-rebase (never auto-aborted) and the run stops,
  reporting the conflicting files and the manual `git rebase --continue`/`--abort` next step.
