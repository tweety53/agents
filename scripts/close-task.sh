#!/usr/bin/env bash
# close-task.sh — run the task-close sequence of skills/flow/implement.md
# (**The next implementer overlaps the guard.**, step 2) for every task a
# returning implementer committed.
#
# Usage: close-task.sh [-end-key <key> -session-token <token> [-end-commit <sha>]]
#          [-undeclared <task-id>=<path>[,<path>…]]…
#          <canonical-worktree> <name> <merge-base> <task-id>:<sha|worktree=sha[,worktree=sha…]>…
#
# In this order:
#   1. With -end-key: `flow record dispatch end -change <name> -key <key>
#      -session-token <token> [-commit <sha>] -outcome completed`. A record
#      never blocks: a failed call prints a warning and the sequence goes on.
#   2. check-task-commit-fields.sh <canonical> <id> <sha|map> "" <canonical>
#      <name> for every task.
#   3. check-task-commit-planning-paths.sh <canonical> <merge-base>.
#   4. Any guard exit 2 → exit 2; else any exit 1 → exit 1. Nothing is ticked
#      or pushed either way, so a refused commit is never pushed.
#   5. check-review-gate.sh <canonical> <id> <sha|map> <canonical> <name>
#      [<the task's -undeclared paths>…] for every task; its FIRE:/QUIET:
#      line is printed. A gate exit 2 → exit 2, nothing ticked or pushed.
#   6. `flow tasks tick -C <canonical> <name> <id>` for every task the gate
#      left QUIET.
#   7. `git -C <worktree> push origin spectre/<name>` once per distinct
#      worktree: the canonical one for a bare sha, each map pair's otherwise.
#   A failed tick or push → exit 2; every tick and push is still attempted.
#
# The guards run in-process in flow-guard; their own lines are printed as
# they print them.
#
# Exit 0 closed (read each FIRE: line: that task's tick waits for its
# reviewer); 1 a commit guard refused — re-commit, then re-run without the
# -end-* flags; 2 stop: a guard could not judge, a tick or push failed, or
# a usage error.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
. "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "close-task: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec close-task 2 "close-task:" "$@"
