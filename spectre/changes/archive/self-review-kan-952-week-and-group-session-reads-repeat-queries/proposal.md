# self-review-kan-952-week-and-group-session-reads-repeat-queries

## Why

The self-review pass over kan-952 found three non-`big` defects, each held against the tree
(`design.md`, **Context**):

- A creating run never runs the unfinished-work guard on itself, so kan-952's panel round 1 —
  dispatched with no pass row — surfaced only at integrate's `prepare` as `STOP: unfinished-work`.
  The fix-run-only scope recorded by self-review-kan-937 rested on "a creating run's panel already
  closes its own findings", which stopped holding once the guard gained its pass-row signal.
- kan-952's `panel-fix-0` dispatch row recorded `2f0fe509`, the change's plan commit, as its
  `-commit`, while the fix landed as `50af0ddf`. `check-panel-fix-single-dispatch.sh` checks keys
  and counts only, so nothing caught it.
- gymie has no tool for counting the SQL statements and transactions an HTTP request runs; kan-924
  and kan-952 both did that audit by hand, and kan-952's first attempt sliced the Postgres log by
  time and got the wrong lines.

## What changes

- **Write `IN_PROGRESS`** (`skills/flow/verify-and-handoff.md`): every run writing `IN_PROGRESS`,
  a creating run as well as a fix run, runs `check-unfinished-work.sh` on itself before the state
  is written.
- `check-panel-fix-single-dispatch.sh` refuses (exit 1) a panel-fix row whose recorded commit
  carries neither a task's `**Commit:**` subject nor a `Flow-Key:` trailer naming the row's key;
  `skills/flow/review-panel.md` states both accepted commits and that the parent re-ends such a row
  with the commit the fix landed in.
- gymie gains `scripts/count-request-queries.sh`, printing statements and `BEGIN`s per request
  against a worktree database, with a shell harness for its log slicing and a pointer in gymie's
  `.flow/project.md`.
