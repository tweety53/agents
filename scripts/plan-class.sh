#!/usr/bin/env bash
# plan-class.sh — the planner's mechanical classifier and rolls.
#
# Usage: plan-class.sh <tasks.md> <repos> [<worktree> <merge-base>]
#
# Prints exactly three lines to stdout:
#   inputs: tasks=N files=N repos=N migration=yes|no spec=yes|no red=yes|no unverified=yes|no
#   class: micro|small|regular|big
#   rolls: compact N · experimental N · bundle N · effort N
#
# Exit 0 on a printed answer, exit 2 on a missing <tasks.md>, a
# non-integer <repos>, an argument count that is neither 2 nor 4, or — with
# the optional arguments — a <worktree>/<merge-base> that cannot be answered
# (not a directory, not a git worktree, an unresolving merge base, or a
# failed git invocation). Rules from design.md's "Inputs and the class" and
# "The rolls" (kan-472-flow-dynamic-review-panel-roster-repo-scoped); the
# small/big thresholds were raised by kan-490-flow-fast-a-reduced-ceremony-flow-variant-drop
# so more changes classify small/regular and roll compact, then lowered by
# roughly 25% so a plan is classed up one step sooner:
#
#   tasks     = count of column-0 `- [ ] <n>.` / `- [x] <n>.` lines
#   files     = size of the union of every task's `**Files:**` backticked
#               paths
#   repos     = the <repos> argument, verbatim — the number of distinct
#               repository roots the plan's **Files:** fall under, which the
#               planner counts (brainstorm-planner.md's Decide); this script
#               only classifies it
#   migration = any **Files:** path under stats/internal/store/migrations/
#               or ending .sql
#   spec      = any **Files:** path under spectre/specs/
#   red       = any task tagged `**Build:** red`
#   unverified= any `unverified:` provenance tag anywhere in the plan
#
#   small:   tasks<=8 and files<=19 and repos=1 and not migration and not spec
#   big:     tasks>=22 or files>=60 or (repos>1 and tasks>=11)
#            or (migration and tasks>=11)
#   regular: everything else
#
#   micro:   the small thresholds met with tasks<=2, every **Files:** path
#            documentation (.md/.mdc), no `**Build:** red` tag, and — when
#            <worktree> <merge-base> are passed — this change's own touched
#            paths (paneltouchedpaths.go's union of
#            committed-since-merge-base, staged and unstaged) entirely
#            documentation and at most planclass.go's pcMicroLineCap (20)
#            changed lines; an empty
#            touched surface passes vacuously. Without the optional arguments
#            micro never fires, so the two-argument form classifies exactly as
#            before it existed. Added by
#            kan-617-flow-cost-the-decide-record-ceremony-runs-full so a
#            two-line prose change stops paying the full decide ceremony
#            (brainstorm-planner.md's Decide collapses to a recorded default
#            decision on this class).
#
# The planner may raise this class one step with a recorded override — that
# judgment call lives in brainstorm-planner.md, never in this script, so
# `override` never appears in this script's own output
# (stats/internal/guard/plan_class_test.go asserts exactly that).
#
# Rolls are reproducible per change name (basename of the directory holding
# <tasks.md>): compact_roll = sha256("<name>") mod 100, experimental_roll =
# sha256("<name>exp") mod 100, bundle_roll = sha256("<name>bundle") mod 100,
# effort_roll = sha256("<name>effort") mod 100 (added by
# medium-effort-default), each over the first 8 hex digits of the digest read as an integer —
# hashed by stats/internal/guard/sha256.go's sha256Hex.
#
# The logic is the Go port in stats/internal/guard/planclass.go. flow-guard
# is built from this checkout, never taken from PATH: scripts/lib/flow-guard.sh
# derives it, and exits 2 (this script's refusal code) with the cause when it
# cannot.
set -euo pipefail
. "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "plan-class: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec plan-class 2 "plan-class:" "$@"
