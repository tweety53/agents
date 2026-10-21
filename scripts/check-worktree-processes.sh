#!/usr/bin/env bash
# check-worktree-processes.sh — report every process whose working directory is
# at or under a worktree, so /flow's cleanup run can verify the project's stack
# actually stopped before it removes anything.
#
# Usage: check-worktree-processes.sh <worktree>
#
# Prints ONE verdict line to stdout:
#   CLEAR: <worktree>                       no process has a cwd at or under it
#   HELD: <worktree> — <pid> <cwd>[; ...]   one or more processes do
#
# Exit 0 whenever a verdict was reached; exit 2 when it cannot answer at all —
# lsof absent, the wrong number of arguments, a worktree path that is not a
# readable directory, or an lsof invocation that itself failed. The VERDICT
# carries the answer, not the exit status; check-cleanup-complete.sh's header
# records why this repository separates the two. A refusal writes its reason to
# stderr and NOTHING to stdout, so a caller grepping for CLEAR in empty output
# can never read an inability as a pass.
#
# WHY A WORKING DIRECTORY AND NOT A PORT. A port proves nothing: a live
# workspace legitimately holds one, and a stack that has genuinely stopped can
# still leave a port in TIME_WAIT. What a process launched from a worktree has,
# and what ties it to the directory about to be removed, is its working
# directory — which is also why removing the worktree is what makes the process
# unreachable. gymie's own OrphanScan.kt reaches the same conclusion
# independently.
#
# WHY THIS ASKS THE OPERATING SYSTEM AND NOT THE PROJECT. The records a project
# keeps about what it started live inside the worktree being removed, so they
# are exactly the evidence a removal destroys. Reading them to decide whether
# the removal is safe is circular. This guard reads the process table instead.
#
# PROJECT-AGNOSTIC BY CONSTRUCTION. It holds no port list, invokes no build
# tool, and requires no project to declare anything, so it protects every
# project flow is installed into rather than only those that opted in. That
# is deliberate: a guard needing an opt-in would not have protected the project
# this incident happened in.
#
# NO sudo, AND THE LIMIT THAT FOLLOWS FROM IT. lsof without privileges reports
# the invoking user's own processes, which is the correct set — a worktree's
# stack is started by the operator, in the operator's own session. The limit is
# recorded here rather than left to be discovered: a process owned by another
# user, or a root-owned one, is outside what this guard can see, and a CLEAR
# verdict says nothing about those.
#
# THE CALLER'S OWN SHELL COUNTS. A shell whose cwd is inside the worktree is a
# process with a working directory at or under it, and this guard reports it
# like any other — it cannot tell the operator's own `cd` from a stray daemon.
# Run it from outside the worktree.
#
# PATHS ARE COMPARED IN THEIR PHYSICAL FORM. lsof reports resolved paths, so on
# macOS a worktree under /tmp reaches it as /private/tmp/... The worktree
# argument is resolved the same way (the `pwd -P` form) before anything is
# compared, and the verdict line echoes that resolved form rather than the
# argument, so the path in the report is the path the comparison used. Skipping
# this makes every symlinked worktree report CLEAR while holding a process.
#
# "AT OR UNDER" IS NOT "STARTS WITH". A path matches when it equals the worktree
# or begins with the worktree plus a slash. A bare string-prefix test reports
# .../spectre-kan-1 as held by a process running in .../spectre-kan-12, which
# is a false block on a worktree the operator cannot clear — and this pipeline's
# own branch naming produces exactly those pairs.
#
# THE lsof EXIT STATUS IS READ, NEVER THE EMPTY OUTPUT. An lsof that failed and
# an lsof that found nothing both print nothing, and treating them alike would
# turn every failure into a CLEAR verdict — the inversion this guard exists to
# prevent, one layer down.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-worktree-processes: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec check-worktree-processes 2 "check-worktree-processes:" "$@"
