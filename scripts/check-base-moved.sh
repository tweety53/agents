#!/usr/bin/env bash
# check-base-moved.sh — report how far the base branch has moved under a
# change, and whether the movement overlaps the change's own touched paths.
#
# Usage: check-base-moved.sh <worktree> <base-ref> <recorded-merge-base|->
#
# Prints ONE verdict line to stdout:
#   CLEAR:  <worktree> — <ref> has not moved since the recorded merge base
#   MOVED:  <worktree> — <n> commits on <ref> since the recorded merge base;
#           no overlap with this change's paths
#   MOVED:  <worktree> — <n> commits on <ref> since the recorded merge base;
#           overlaps: <paths>
#   REFUSE: <reason>
#
# Exit 0 whenever a verdict was reached; exit 2 when the tree cannot be read.
# The VERDICT carries the answer, not the exit status — see
# check-finish-preflight.sh's header for why this repository separates them
# (design.md: verdict-protocol-matches-siblings). This guard performs no
# fetch of its own: resolve-base-branch.sh fetched the worktree when the
# caller resolved the base ref.
#
# Base-ref resolution is shared with check-finish-preflight.sh via
# scripts/lib/resolve-remote-base.sh (design.md: base-moved-is-a-guard), so
# the two guards can never disagree about which ref answers a question about
# the base.
#
# THE CHANGE'S OWN PATHS INCLUDE THE INDEX AND THE WORKING TREE (design.md:
# touched-paths-include-index-and-worktree). Run 1 — the only run this guard
# serves — is reached with work staged and uncommitted by design, so a
# comparison limited to committed history would report "no overlap" for a
# change that does conflict.
#
# Every git invocation whose failure would otherwise be read as an answer is
# captured into a variable and checked on its own line, never piped straight
# into `sort`/`comm`/`wc` — the reasoning check-finish-preflight.sh's signal
# (d) comment already records, cited rather than restated here. A failing
# invocation is exit 2 with a named message, never a CLEAR.
set -euo pipefail
export LC_ALL=C

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/resolve-remote-base.sh
. "$SCRIPT_DIR/lib/resolve-remote-base.sh"

WORKTREE="${1:-}"
BASE_REF="${2:-}"
RECORDED="${3:-}"

if [ -z "$WORKTREE" ] || [ -z "$BASE_REF" ] || [ -z "$RECORDED" ]; then
  cat >&2 <<'EOF'
usage: check-base-moved.sh <worktree> <base-ref> <recorded-merge-base|->
  <base-ref>  the base branch name, bare (main) or remote-tracking
              (origin/main). The guard prefers refs/remotes/origin/<base-ref>
              when it resolves, so a bare name is never tested against a
              stale local branch.
EOF
  exit 2
fi

if [ ! -d "$WORKTREE" ]; then
  echo "check-base-moved: $WORKTREE is not a directory — cannot determine anything" >&2
  exit 2
fi

if ! git -C "$WORKTREE" rev-parse --git-dir >/dev/null 2>&1; then
  echo "check-base-moved: $WORKTREE is not a git worktree — cannot determine anything" >&2
  exit 2
fi

# No recorded merge base: an honest unknown, never an inferred verdict.
if [ "$RECORDED" = "-" ]; then
  echo "REFUSE: no merge base recorded for $WORKTREE — cannot tell whether the base has moved"
  exit 0
fi

# Every ref this script did not choose itself is passed after
# --end-of-options, so a value beginning with `-` is read as a ref and
# rejected rather than parsed as a git option.
RECORDED_SHA="$(git -C "$WORKTREE" rev-parse --verify --end-of-options "${RECORDED}^{commit}" 2>/dev/null)" || {
  echo "REFUSE: recorded merge base '$RECORDED' does not resolve in $WORKTREE"
  exit 0
}

EFFECTIVE_REF="$(resolve_remote_base "$WORKTREE" "$BASE_REF")"

if ! git -C "$WORKTREE" rev-parse --verify --end-of-options "${EFFECTIVE_REF}^{commit}" >/dev/null 2>&1; then
  echo "REFUSE: base ref '$EFFECTIVE_REF' does not resolve in $WORKTREE — cannot tell whether the base has moved"
  exit 0
fi

COUNT="$(git -C "$WORKTREE" rev-list --count --end-of-options "${RECORDED_SHA}..${EFFECTIVE_REF}" 2>/dev/null)" || {
  echo "check-base-moved: cannot count commits between $RECORDED_SHA and $EFFECTIVE_REF in $WORKTREE" >&2
  exit 2
}

if [ "$COUNT" = "0" ]; then
  echo "CLEAR: $WORKTREE — $EFFECTIVE_REF has not moved since the recorded merge base"
  exit 0
fi

MOVED_RAW="$(git -C "$WORKTREE" diff --name-only --end-of-options "${RECORDED_SHA}..${EFFECTIVE_REF}" 2>/dev/null)" || {
  echo "check-base-moved: cannot list paths changed on $EFFECTIVE_REF in $WORKTREE" >&2
  exit 2
}

# The change's own paths (design.md: touched-paths-include-index-and-worktree):
# the union of what HEAD carries since the recorded merge base, what is
# staged, and what is unstaged. Three separate captures, each checked before
# any is used, rather than one shell pipeline whose partial failure could
# read as an empty — and therefore clean — set.
COMMITTED_RAW="$(git -C "$WORKTREE" diff --name-only --end-of-options "${RECORDED_SHA}..HEAD" 2>/dev/null)" || {
  echo "check-base-moved: cannot list this change's committed paths in $WORKTREE" >&2
  exit 2
}

STAGED_RAW="$(git -C "$WORKTREE" diff --name-only --cached 2>/dev/null)" || {
  echo "check-base-moved: cannot list this change's staged paths in $WORKTREE" >&2
  exit 2
}

UNSTAGED_RAW="$(git -C "$WORKTREE" diff --name-only 2>/dev/null)" || {
  echo "check-base-moved: cannot list this change's unstaged paths in $WORKTREE" >&2
  exit 2
}

MOVED_SORTED="$(printf '%s\n' "$MOVED_RAW" | sort -u)"
CHANGE_SORTED="$(printf '%s\n%s\n%s\n' "$COMMITTED_RAW" "$STAGED_RAW" "$UNSTAGED_RAW" | sort -u)"

# The intersection, computed only from the already-captured, already-checked
# variables above — never by re-running git inside the pipe. `comm -12`
# needs sorted input, which is what MOVED_SORTED and CHANGE_SORTED are.
OVERLAP=()
while IFS= read -r path; do
  [ -n "$path" ] && OVERLAP+=("$path")
done < <(comm -12 <(printf '%s\n' "$MOVED_SORTED") <(printf '%s\n' "$CHANGE_SORTED"))

if [ "${#OVERLAP[@]}" -eq 0 ]; then
  echo "MOVED: $WORKTREE — $COUNT commits on $EFFECTIVE_REF since the recorded merge base; no overlap with this change's paths"
  exit 0
fi

TOTAL="${#OVERLAP[@]}"
if [ "$TOTAL" -gt 10 ]; then
  SHOWN="$(printf '%s, ' "${OVERLAP[@]:0:10}")"
  SHOWN="${SHOWN%, } (+$((TOTAL - 10)) more)"
else
  SHOWN="$(printf '%s, ' "${OVERLAP[@]}")"
  SHOWN="${SHOWN%, }"
fi

echo "MOVED: $WORKTREE — $COUNT commits on $EFFECTIVE_REF since the recorded merge base; overlaps: $SHOWN"
exit 0
