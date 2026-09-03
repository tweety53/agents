#!/usr/bin/env bash
# check-panel-citation-trigger.sh — decide whether the review panel's
# citation pre-check should run, in a guard instead of prose.
#
# Usage: check-panel-citation-trigger.sh <worktree> <merge-base>
#
# Prints nothing; the exit code is the whole answer.
#   0  the change's own paths include at least one path ending .md or .mdc
#   1  none do
#   2  usage error: a missing argument, <worktree> not a directory or not a
#      git worktree, <merge-base> not resolving, or a git invocation failed
#
# THE CHANGE'S OWN PATHS INCLUDE THE INDEX AND THE WORKING TREE (design.md:
# touched-paths-include-index-and-worktree, the same rule check-base-moved.sh
# already applies): the union of what HEAD carries since the merge base,
# what is staged, and what is unstaged.
set -euo pipefail

PROG="check-panel-citation-trigger"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# GIT_BIN resolution (KAN-304 review round 6, F8/F11/F13), argument/worktree
# /merge-base validation, and the COMMITTED/STAGED/UNSTAGED path collection
# are shared with check-panel-docs-only.sh — see lib/panel-touched-paths.sh
# for the full rationale, including why GIT_BIN is resolved to an absolute
# path with `type -P` rather than invoked as a bare `git` word (KAN-312
# review round 0, F1 on why this preamble is a sourced library rather than
# two copies).
source "$SCRIPT_DIR/lib/panel-touched-paths.sh"

WORKTREE="${1:-}"
MERGEBASE="${2:-}"

GIT_BIN="$(panel_resolve_git "$PROG")" || exit 2
panel_validate_worktree "$PROG" "$WORKTREE" "$MERGEBASE" "$GIT_BIN" || exit 2
PATHS="$(panel_touched_paths "$PROG" "$WORKTREE" "$MERGEBASE" "$GIT_BIN")" || exit 2

MATCH="$(printf '%s\n' "$PATHS" | awk 'NF && ($0 ~ /\.mdc?$/) { print; exit }')"

if [ -n "$MATCH" ]; then
  exit 0
fi
exit 1
