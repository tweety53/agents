#!/usr/bin/env bash
# check-base-moved.sh — report how far the base branch has moved under a
# change, and whether the movement overlaps the change's own touched paths.
#
# Usage: check-base-moved.sh <worktree> <base-ref> <recorded-merge-base|->
#
# Prints ONE verdict line to stdout:
#   CLEAR:  <worktree> — <ref> has not moved since the recorded merge base
#   CLEAR:  <worktree> — the <n> commits <ref> gained since the recorded
#           merge base are all already carried by this branch — nothing to
#           rebase
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
# stats/internal/guard/resolveremotebase.go (design.md: base-moved-is-a-guard), so
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
# (d) comment recorded at d71a2327 (now stats/internal/guard/finishpreflight.go's (d)),
# cited rather than restated here. A failing
# invocation is exit 2 with a named message, never a CLEAR.
#
# HOW TO HAND-VERIFY A MOVED VERDICT (KAN-446). Recount the commits yourself
# with `git rev-list <recorded-merge-base>..<ref>`, `<ref>` being the one the
# verdict line names (EFFECTIVE_REF — not always the bare argument), and
# intersect the verdict's `overlaps:` paths with the change's own touched
# paths — which include the index and the working tree, exactly as this guard
# counts them. A stale recorded merge base — one recorded before a rebase
# this pipeline performed — is the known structural cause of a movement that
# is not real. A CLEAR naming carried movement recounts instead with
# `git rev-list <recorded-merge-base>..<ref> ^HEAD` — zero means every commit
# the base gained is already reachable from the branch, the benign cause
# KAN-535 records.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-base-moved: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-base-moved 2 "check-base-moved:" "$@"
