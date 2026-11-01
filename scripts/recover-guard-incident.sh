#!/usr/bin/env bash
# recover-guard-incident.sh — KAN-423's incident recovery, reduced: abort a
# stuck `git revert`, after printing the reflog diagnosis. It once also
# restored the untracked planning files the incident stash held; the change
# directory is gitignored now (KAN-1073), so `git stash -u` never captures
# it and there is nothing to restore.
#
# Usage: recover-guard-incident.sh [--apply] [repo-dir]
#
#   --apply    execute; without it, print what would run and change nothing
#   repo-dir   defaults to the caller's cwd; must be a git repository
#
# Exit 0 on a printed dry-run plan or a completed apply; 1 when no revert is
# in progress (REVERT_HEAD missing; cause on stderr, stdout empty); 2 on a
# usage error — an unknown option, an argument after repo-dir, a repo-dir
# that is not a directory, or a repo-dir that is not a git repository.
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
