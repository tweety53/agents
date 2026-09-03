#!/usr/bin/env bash
# check-panel-docs-only.sh — decide whether this change's own paths are
# entirely documentation, in a guard instead of prose.
#
# Usage: check-panel-docs-only.sh <worktree> <merge-base>
#
# Exit codes:
#   0  every touched path ends .md or .mdc
#   1  at least one touched path does not — that first path is printed to
#      stdout — including an empty touched-path set
#   2  cannot answer: a missing argument, <worktree> not a directory or not
#      a git worktree, <merge-base> not resolving, or a git invocation
#      failed
#
# THE CHANGE'S OWN PATHS are the union of committed-since-merge-base,
# staged and unstaged paths, exactly as check-panel-citation-trigger.sh
# collects them.
#
# WHY THIS GUARD EXISTS (KAN-312): the whole-branch panel re-reads a
# docs-only branch every per-task reviewer already read.
set -euo pipefail

PROG="check-panel-docs-only"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# GIT_BIN resolution, argument/worktree/merge-base validation and the
# COMMITTED/STAGED/UNSTAGED path collection are shared with
# check-panel-citation-trigger.sh — see lib/panel-touched-paths.sh for the
# full rationale (KAN-304 review round 6, F8/F11/F13 on GIT_BIN; KAN-312
# review round 0, F1 on why this preamble is a sourced library rather than
# two copies).
source "$SCRIPT_DIR/lib/panel-touched-paths.sh"

WORKTREE="${1:-}"
MERGEBASE="${2:-}"

GIT_BIN="$(panel_resolve_git "$PROG")" || exit 2
panel_validate_worktree "$PROG" "$WORKTREE" "$MERGEBASE" "$GIT_BIN" || exit 2
PATHS="$(panel_touched_paths "$PROG" "$WORKTREE" "$MERGEBASE" "$GIT_BIN")" || exit 2

if [ -z "$PATHS" ]; then
  exit 1
fi

NONDOC="$(printf '%s\n' "$PATHS" | awk '$0 !~ /\.mdc?$/ { print; exit }')"

if [ -n "$NONDOC" ]; then
  printf '%s\n' "$NONDOC"
  exit 1
fi
exit 0
