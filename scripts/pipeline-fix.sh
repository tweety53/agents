#!/usr/bin/env bash
# pipeline-fix.sh — the fixed mechanics of a pipeline-fix or self-review fix
# loop (skills/flow-contracts/pipeline.md **Pipeline defects found mid-run**,
# skills/flow-self-review/SKILL.md step 3), run by the parent, so no step is
# retyped or skipped. The loop is its own change in this repository's flowd
# project, named after its branch — pipeline-fix-<name>-<k> or
# self-review-<name> — and check-loop-dispatch.sh files every loop dispatch
# under it (a self-review on the change's own branch files under <name>, in
# the flowd project of the caller's working directory: end and escalate of
# such a key run from the run's own project).
#
# Usage:
#   pipeline-fix.sh start <name> [<k>]
#       git fetch origin; `git worktree add -b <branch> <repo>-worktrees/<branch>
#       origin/<default>`, <branch> pipeline-fix-<name>-<k> with <k>, else
#       self-review-<name>; refuses a <k> that is not one more than the highest
#       <n> of any pipeline-fix-*-<n> branch (local or remote) or change folder
#       (spectre/changes/ or its archive/), so <k> is repo-unique; creates <worktree>/.superpowers/sdd/<branch>/ and the
#       loop's STARTED state record (flow state set); runs ## worktree setup.
#       stdout: `worktree: <path>`.
#   pipeline-fix.sh end <key> <outcome>
#   pipeline-fix.sh end <review-key> <report-file>|aborted
#       closes the dispatch <key> (flow record dispatch end), `completed` or
#       `aborted`. A review takes its reviewer's report file, and closes
#       `clean` when its `Critical: <n>` and `Important: <n>` lines are both 0,
#       `fix` otherwise; a report missing either line exits 2. An `aborted`
#       key may be dispatched again.
#   pipeline-fix.sh escalate <key>
#       records <key>'s no-progress escalation (git config flow.escalation,
#       value `<loop branch>:<key>`; land and drop clear the loop's), which
#       check-loop-dispatch.sh requires before admitting <key> on its role's
#       marked `No-progress escalation` row of **Allowed pairs**
#       (skills/flow-contracts/model-policy.md); refused unless that table
#       gives <key>'s role such a row and an earlier round of its kind
#       completed.
#   pipeline-fix.sh verify <worktree>
#       runs the ## lint lines and tests the branch's touched paths need,
#       against <worktree>'s own origin/HEAD: every guard line, the `cd stats` lines for
#       touched Go, the `cd stats/web` lines for a touched SPA file; go test
#       of each touched stats package, scripts/test-<n>.sh or
#       scripts/test-<n>-hook.sh of a touched script or hook, the setuptest
#       package for setup.sh, the internal/guard tests that audit every
#       script (TestNoPipeIntoEarlyExitGrep and its kin) for a touched script
#       or setup.sh, vitest of a touched SPA test or the test beside a touched
#       source. One `green:` or `red:` line per command, a red one
#       followed by its last 40 output lines.
#   pipeline-fix.sh land <worktree>
#       the merge-and-push route: refuses any other `## default landing
#       route`; git fetch origin; a branch whose base moved is rebased onto
#       origin/<default> (a stopped rebase is aborted); verify, always, before
#       anything is pushed; refuses a spectre/changes/<branch>/ still in the
#       branch's tree (unarchived); check-loop-landing's refusals; `git push
#       origin <branch>:<default>`; the loop's FINISHED state record; `git
#       pull --ff-only` in the main checkout; removes the worktree and the
#       branch. stdout: `landed: <sha>`. When that pull succeeded and moved
#       a stats/ path (with stats/Makefile and stats/cmd/flowd present), `make -C <main>/stats
#       restart` runs after it: `FLOWD-RESTARTED: <main>/stats`, or
#       `FLOWD-RESTART-FAILED: make -C <main>/stats restart exited <n> — run
#       it by hand` with make's output on stderr, exit 1.
#   pipeline-fix.sh drop <worktree>
#       a deferred loop: the FINISHED state record with withdrawn: true, then
#       removes the worktree (--force) and the branch, landing nothing.
#
# Exit codes: 0 done; 1 refused, red, or a step failed (its output relayed);
# 2 usage or it cannot answer (no flowd answer, no git, no origin/HEAD).
#
# Go, in flow-guard (stats/internal/guard/pipelinefix.go). flow-guard is
# built from this checkout, never taken from PATH: scripts/lib/flow-guard.sh
# derives it, and exits 2 with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "pipeline-fix: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_REPO_ROOT="$(flow_guard_root)"
FLOW_GUARD_PROJECT_GET="$SCRIPT_DIR/project-get.sh"
export FLOW_GUARD_REPO_ROOT FLOW_GUARD_PROJECT_GET
flow_guard_exec pipeline-fix 2 "pipeline-fix:" "$@"
