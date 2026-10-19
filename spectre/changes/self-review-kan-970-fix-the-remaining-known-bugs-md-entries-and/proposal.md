# self-review-kan-970-fix-the-remaining-known-bugs-md-entries-and

## Why

The self-review pass of `kan-970-fix-the-remaining-known-bugs-md-entries-and` found nine fixable
defects in the pipeline and in gymie's dev tooling:

- `check-visual-verify-dispatched` read a satellite's archive move of its own `link.md` as UI work
  after the verified HEAD and forced a second, scoped visual verifier over nothing (F1).
- `check-unfinished-work` cleared a change whose implementer groups and review panel had recorded
  no row at all: nothing requires a change with code commits to carry any pass row (F2).
- The visual verifier wrote `visual-verification.md` outside the change root, citing session
  scratchpad captures, and the visual guard passed on the dispatch row alone (F3).
- `protect-main-checkout.py` denied a write into a sibling worktree because an unquoted heredoc
  body holding a backtick span kept its whole text scanned, reading a literal `> b` as a redirect
  into the main checkout (F4). Planning found the opposite gap beside it: a backtick span, or a
  `$(…)` inside a double-quoted word, was never scanned as a command.
- Two guards accepted a zsh-unsplit list as one argument and failed with a message naming git or
  the list's shape rather than the cause (F5).
- `brainstorm-planner.md` never states the `**After:**` list separator, and a planner wrote
  `Task 5, Task 6`, which the grammar refuses (F6).
- gymie's `devRestart` and a bare `devStart` silently gave an isolated stack's OAuth ports back
  (F7).
- The kan-970 plan designed fixes for entries already fixed upstream and for a remedy two earlier
  changes had rejected (F8).
- The inline-verify recording example puts `begin` and `end` on adjacent lines, the shape that was
  run as one call (F9).

## What changes

- The visual guard drops `<spec root>/changes/` paths before judging, and counts a found verifier
  row only with its record in the change root, live or archived.
- `check-unfinished-work` reports a change with a code commit and no pass row.
- The main-checkout hook scans only the substitution spans of an unquoted data heredoc, and scans
  every backtick span and every quoted `$(…)` as a command.
- `check-unfinished-work` and `gather-dispatch-context` name zsh's unsplit `$VAR` when an argument
  that cannot hold whitespace does.
- `brainstorm-planner.md` states the `**After:**` separator, and its section B checks a reported
  defect against history before its fix is designed.
- The inline-verify recording example separates `begin` from `end`.
- gymie: a run without `-PisolateOAuthPorts` keeps the isolation a running moved frontend or
  gateway was started with; kan-970's visual record moves into its archived change root with its
  evidence.
