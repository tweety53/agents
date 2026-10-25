# self-review-kan-1033-plan-mode-planning

## Why

kan-1033's self-review pass
(`/Users/tweety53/Projects/agents-worktrees/kan-1033-plan-mode-planning/.superpowers/sdd/self-review-pass.md`)
raised six non-`big` findings:

- **1.2** — the run's in-run pipeline fix (`ba68637c`, `Flow-Key: pipeline-fix-mid-run-question`)
  was committed straight on `spectre/kan-1033-plan-mode-planning` with no loop rows, and nothing
  refused it.
- **1.3** — the planning texts pass `plan-agent-pair.sh`'s effort as the Agent tool's `effort`,
  while `skills/flow/SKILL-rationale.md` says the Agent tool has no dispatch-time effort parameter.
- **1.4** — `check-guard-symlinks` drops a backticked citation whose first argument is a
  literal-prefixed placeholder (`fp-<literal-token>`).
- **4.1** — the inline verify row is recorded by hand; kan-1033's was back-filled twice and its
  `verify` row spans from before its fix to the end of the run, enclosing `verify-fix-1` and
  `verify-2`.
- **4.2** — a bare `cat > /tmp/x.go` with no stdin hung a Bash call until its timeout.
- **5.1** — the pass rows of every re-run round were recorded only after the unfinished-work
  guard flagged them.

## What changes

- **1.2:** `check-untasked-commits` lists a `Flow-Key: pipeline-fix-…` commit on any branch but a
  `pipeline-fix-<name>-<k>` loop's own, so the push hook, `close-task.sh` and run 1's
  unfinished-work gate refuse it (Task 1).
- **1.3:** the rationale's stale clause is corrected; the planning texts stand (Task 2).
- **1.4:** `gsPlaceholder` accepts a word prefix before the placeholder's `<` (Task 3).
- **4.1:** `run-verify-list.sh` records the verify row itself when given the row's identity, and
  kan-1033's `verify-fix-1`, `verify` and `verify-2` rows are retimed to the instants the parent
  transcript holds (Task 4).
- **4.2:** `protect-main-checkout.py` denies a Bash command whose `cat` writes a file with no
  input (Task 5).
- **5.1:** `panel-round-begins.sh` records the round's pass row with its begins (Task 6).
