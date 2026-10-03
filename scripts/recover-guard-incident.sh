#!/usr/bin/env bash
# recover-guard-incident.sh — reproduce KAN-423's incident recovery on
# demand: abort a stuck `git revert`, then restore the untracked planning
# files the incident stash holds, unstaged, via `git show "stash@{0}^3:<f>" >
# <f>` redirects.
#
# THE ORDER IS THE POINT. `git checkout stash@{...} -- <path>` stages the
# restore, and any later `revert --abort` discards it — the incident's first
# restore attempt was lost exactly that way. Here every abort precedes every
# restore, and the redirects leave the files unstaged, so nothing after the
# abort point can discard them.
#
# Usage: recover-guard-incident.sh [--apply] [repo-dir] [path...]
#
#   --apply    execute; without it, print what would run and change nothing
#   repo-dir   defaults to the caller's cwd; must be a git repository
#   path...    planning paths to restore, relative to the repo root;
#              defaults to `spectre/changes`
#
# Exit 0 on a printed dry-run plan or a completed apply; 1 on a failed
# precondition (cause on stderr, stdout empty); 2 on a usage error — an
# unknown option, a repo-dir that is not a directory, or a repo-dir that is
# not a git repository.
#
# PRECONDITIONS ARE CHECKED IN ORDER AND THE FIRST FAILURE NAMES ITS CAUSE
# ON STDERR WITH NOTHING ON STDOUT: no revert in progress (REVERT_HEAD
# missing); no stash at all; a stash created without
# `-u` (no untracked third parent — the remedy is re-stashing with -u BEFORE
# any abort, so the two causes are named differently); a restore target
# tracked in the index or HEAD, which refuses the whole run at the first
# offending file.
# A target already present untracked is restored over, named as an
# overwrite in the plan — untracked state is never clobbered silently, and
# tracked state is never clobbered at all.
#
# THE REFLOG BLOCK PRINTS IN EVERY MODE, dry-run and apply alike — the
# incident was diagnosed from `git reflog -g HEAD`'s alternating reset
# pattern, so the diagnosis is visible at the point of decision, not only
# in the mode that acts on it.
#
# Ported to Go (KAN-842): the body's reasoning is in
# stats/internal/guard/recoverguardincident.go.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this script's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "recover-guard-incident: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec recover-guard-incident 2 "recover-guard-incident:" "$@"
