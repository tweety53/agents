# pipeline-fix-kan-940-step-5-frontend-cardio-row-on-the-kan-928-18

## Why

gymie change kan-940 was created with `--base feature/cardio`. Its gymie and gymie-frontend
worktrees record `branch.spectre/<name>.flowBase=feature/cardio`, but its gymie-playwright worktree
— the project's `regression checkout`, added during the review-panel fix round, after kickoff —
was cut by a hand-typed `git worktree add -b spectre/<name> … origin/main` and recorded no
flowBase, although `origin/feature/cardio` exists in gymie-playwright. At integrate,
`resolve-base-branch.sh` resolved `main` for that worktree, and `prepare` rebased the playwright
branch onto `main` and nearly pushed it there (the operator: "why do the playwright code/images
pushed to main and not to a feature branch?").

The recipe that inherits the base exists only as prose in `skills/flow/cross-repo-worktrees.md`,
bound to `flow.isolate-workspace`; `skills/flow/implement.md` forbids improvising a worktree later
but names no recipe for a stage that needs one, and nothing at integrate notices a worktree set
whose bases disagree.

## What changes

- `kickoff-worktree.sh` takes `--from <change worktree>`: the base defaults to that worktree's
  recorded flowBase, and the merge base is recorded in that worktree's project's state record. A
  missing `origin/<base>` stays exit 1, nothing added, its message naming the operator's two
  remedies.
- `skills/flow/cross-repo-worktrees.md`'s `## apps` recipe is that one call, and
  `skills/flow/implement.md` binds any later stage needing a worktree to the same recipe.
- `integrate-change.sh prepare` stops at `STOP: base-mismatch` when a base one worktree records is
  on another worktree's origin but that worktree resolves a different base;
  `skills/flow/integrate.md` says what the parent does at it.
- The saved gymie memory `cardio-feature-branch.md`, which says the cardio epic lives in gymie and
  gymie-frontend only, names gymie-playwright too.

Blast radius: 34 files — the 32 in `<agents repo>` citing `kickoff-worktree`,
`cross-repo-worktrees` or `integrate-change`, plus the two memory files.
<!-- measured: grep -rl --exclude-dir=.git --exclude-dir=archive -e kickoff-worktree -e cross-repo-worktrees -e integrate-change . | sort -u | wc -l → 32, plus 2 memory files @ 2d9dc724 -->
