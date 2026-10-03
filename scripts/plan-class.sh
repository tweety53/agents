#!/usr/bin/env bash
# plan-class.sh — the planner's mechanical classifier and rolls.
#
# Usage: plan-class.sh [-class <class>] <tasks.md> <repos> [<worktree> <merge-base>]
#
# Prints exactly seven lines to stdout:
#   inputs: tasks=N files=N repos=N migration=yes|no spec=yes|no red=yes|no unverified=yes|no
#   class: micro|small|regular|big
#   rolls: compact N · experimental N · bundle N · effort N
#   tree: class <c> · execution inline|sdd · implementer skipped — inline|chosen
#   panel: default|compact|full · roster <slot; …> · rerun delta
#   grouping: static|free · dispatches primary+principles[ · failure-modes+mutation[+exp-<name>]]
#   experimental: no slot|none available|exp-<name> · skills/flow/experimental/<name>.md · <description>[ · skipped — bundle cap]
#
# The first three lines are the mechanical answer. The four tree lines are
# brainstorm-planner.md's Decide tree for the effective class — the -class
# value when given, else class_mechanical — and the rolls: compact when
# compact < 90, static grouping when bundle < 30, an experimental slot when
# experimental < 30. The slot is sorted(skills/flow/experimental/*.md)
# [experimental mod count] under FLOW_GUARD_REPO_ROOT, with its line-1
# `description:` value; an absent directory or no *.md file is `none
# available`. primary+principles is the floor bundle and takes no third
# role, so the slot joins only a full regular/big roster's second dispatch
# and is otherwise `skipped — bundle cap`. On micro, panel is `default` and
# grouping/experimental read `not consulted — micro`.
#
# -class may equal class_mechanical or sit one step above it
# (micro→small→regular→big) — the planner's recorded raise; the class: line
# stays class_mechanical.
#
# Exit 0 on a printed answer, exit 2 on a missing <tasks.md>, a
# non-integer <repos>, an argument count that is neither 2 nor 4, a -class
# that is unknown, two or more steps up or any step down, an experimental
# candidate whose line 1 is not a `description:` line, or — with
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
# The planner may raise this class one step with a recorded override — the
# judgment and its reason live in brainstorm-planner.md; this script only
# bounds the raise (-class), so `override` never appears in this script's own
# output (stats/internal/guard/plan_class_test.go asserts exactly that).
#
# Rolls are reproducible per change name (basename of the directory holding
# <tasks.md>): compact_roll = sha256("<name>") mod 100, experimental_roll =
# sha256("<name>exp") mod 100, bundle_roll = sha256("<name>bundle") mod 100,
# effort_roll = sha256("<name>effort") mod 100 (added by
# medium-effort-default), each over the first 8 hex digits of the digest read as an integer —
# hashed by stats/internal/guard/sha256.go's sha256Hex.
#
# The logic is the Go port in stats/internal/guard/planclass.go. This shim
# exports FLOW_GUARD_REPO_ROOT — the checkout lib/ physically sits in, so a
# skill's scripts/ symlink resolves to the repository, not to the skill — as
# the root the experimental directory is read under. flow-guard
# is built from this checkout, never taken from PATH: scripts/lib/flow-guard.sh
# derives it, and exits 2 (this script's refusal code) with the cause when it
# cannot.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "plan-class: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_REPO_ROOT="$(flow_guard_root)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec plan-class 2 "plan-class:" "$@"
