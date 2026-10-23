# self-review-kan-1026-stored-data-first-on-every-open

## Why

The self-review pass of `kan-1026-stored-data-first-on-every-open` found seven fixable pipeline
defects:

- Ledger rows left without their agent or stamped with another dispatch's agent: launch
  descriptions written in prose, so no launch was stamped by its key; begin keys built from a
  `$R` substitution; a lone begin timed from the start of the Bash call it closed; and a flowd
  restart between a round's begins and its launches.
- A panel round's begins block is hand-written each round, which is where the substituted keys
  came from.
- A panel slot's end instant is derived in Bash from its report file's mtime.
- `skills/flow/visual-verify.md` prints a hand push for a commit in the change's own worktree,
  which Branch backup already pushes.
- `hooks/protect-main-checkout.py` denies commands that write only into worktrees: a python
  heredoc's `d>0` read as a redirect, and a report heredoc whose command names variables.
- The `-tokens` refusal on `flow record dispatch end` is checked on `CLAUDE_CODE_SESSION_ID`
  alone, so unsetting it bypasses the guard.
- The self-review bundle carries the verification record's raw bands, seams and ink summaries,
  which no angle reads.

## What changes

- In a flow run, `hooks/enforce-agent-baseline.py` denies a dispatch whose description does not
  open with its row; a lone begin pairs at its row's own `started_at`; the kan-1026 rows are
  restamped onto their true agents.
- `panel-round-begins.sh` records a round's begins under literal `panel-<round>-<slot>` keys.
- flowd closes a stamped row whose end names no instant at its agent transcript's last write.
- A visual-verify commit in a resolved worktree rides Branch backup; only a main-checkout commit
  prints a hand push.
- A heredoc body no shell runs is judged by its git invocations only, and the hook reads a
  command's variables and `$(...)` spans before deciding a report body may run.
- The store refuses a token report on any session begun on a harvested harness, naming the
  backfill route.
- The bundle replaces each bands, seams or ink section of `visual-verification.md` with one line
  naming its size, heading and the record's path.
