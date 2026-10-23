# self-review-kan-1027-flow-learning-setting

## Why

The self-review pass of `kan-1027-flow-learning-setting` found five fixable pipeline defects:

- `skills/flow-contracts/learning.md` records the learning note with `flow learning add … -C
  <worktree>` as the final message prints; on `/flow-fast` the worktree is already removed by
  cleanup, so `fallback.ProjectKey` fails (CLI exit 1, `resolve project key`), the failure reads
  as an unreachable store, and the three-times cap never applies.
- The COMMIT-PER-TASK chain in `skills/flow/implement.md` runs unanchored `git reset`/`add`/
  `commit` with repo-relative paths; from a subdirectory such as `stats/` the `git add` fails
  with a pathspec error (exit 128).
- Nothing tells a planner that a task editing run-loaded prose in this repository needs
  `check-verbatim-moves.sh` in its verify step and a `verbatim-moves.txt` committed with the plan;
  kan-1027's plan named neither, and the panel raised it as F1/F8.
- An inline `flow record dispatch backfill -commit` derives its start from the commit before the
  fix commit; kan-1027's `panel-fix-1` row took the pre-review planning commit's instant
  (09:19:34Z), before panel-0's reviewers began (09:20:01Z) or reported (09:28:18Z), inflating its
  duration and leaving its token figures empty.
- The live check measures its merge-base before figures serially in `flow.verify`, after the
  panel, although they depend on nothing the panel changes.

## What changes

- `learning.md` records the note with `-C <project>`.
- The COMMIT-PER-TASK chain anchors every git call with `git -C <abs-worktree>`.
- `.flow/project.md`'s `## lint` prose tells a plan to name `check-verbatim-moves.sh` in a
  prose task's verify step and commit `verbatim-moves.txt` with the plan.
- An inline `panel-fix-<r>` back-fill's derived start is never earlier than the last end among
  round `r-1`'s reviewer rows; kan-1027's `panel-fix-1` row is retimed and its tokens restamped.
- Pass 1's panel window measures a live check's before figures into
  `.superpowers/sdd/live-before.md`, and **Live check** reuses them.
