#!/usr/bin/env bash
# check-verify-green.sh — whether this run's inline verify is green, run by
# `flow.visual-verify` step 3 before any verifier is dispatched
# (skills/flow/visual-verify.md): visual verification starts only after
# **Verify** (skills/flow/verify-and-handoff.md) has closed green, never
# beside a verify still running or blocked.
#
# Usage: check-verify-green.sh <worktree> <change-name> <session-token>
#
# Reads `flow record dispatches -change <change-name> -C <worktree>` and
# judges this session token's inline verify rows — role `verifier`, key
# `verify` or `verify-<worktree basename>`; the visual verifier's
# `visual-verify…` rows never count. Only the latest such row per worktree is
# judged: rows group by key with a `-fix-<k>` suffix removed
# (skills/flow/verify-fix-loop.md step 4), and the last row of each group in
# record order stands. Prints one verdict line:
#   VERIFY-GREEN:     <change> — <n> inline verify row(s) completed
#   VERIFY-NOT-GREEN: <change> — this run recorded no inline verify
#   VERIFY-NOT-GREEN: <change> — <key> (running|<outcome>)[, …]
# Exit 0 every judged row carries outcome `completed` and at least one exists;
# 1 VERIFY-NOT-GREEN; 2 cannot answer (usage, a missing worktree, a change
# name that is not plain, a failed or malformed store read) — the cause on
# stderr.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-verify-green: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-verify-green 2 "check-verify-green:" "$@"
