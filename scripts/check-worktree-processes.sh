#!/usr/bin/env bash
# check-worktree-processes.sh — report every process whose working directory is
# at or under a worktree, so /myflow-finish run 2 can verify the project's stack
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
# argument is resolved the same way (`cd … && pwd -P`) before anything is
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
# `-e` IS OMITTED ON PURPOSE, AND THE SIBLING GUARDS' `-euo` DOES NOT APPLY
# HERE. Every other /myflow-finish guard opens `set -euo pipefail`; this one
# opens `set -uo pipefail`, and adding the `-e` would break the contract stated
# at the top of this header. The lsof call below assigns its output in a command
# substitution and reads `$?` on the very next line, which is what lets a failed
# lsof be reported by name. Under `set -e` the shell aborts at the assignment
# itself, before that status is ever read, so the operator gets bash's bare
# unexplained exit in place of the named reason — an inability reported without
# its reason, which is precisely what this guard promises never to do. Nothing
# here relies on `-e` to stop: every failure path calls die() explicitly, so the
# exit-2 refusals are deliberate rather than inherited.

set -uo pipefail

SELF="check-worktree-processes"

die() {
  printf '%s: %s\n' "$SELF" "$*" >&2
  exit 2
}

[ "$#" -eq 1 ] || die "usage: check-worktree-processes.sh <worktree>"

WT="$1"

# lsof is this guard's only source of an answer, so its absence is an inability
# and never an empty result. The message names the tool, because "the check
# failed" leaves the operator nothing to install.
command -v lsof >/dev/null 2>&1 \
  || die "lsof is not on PATH — this guard has no other way to read the process table"

[ -d "$WT" ] || die "not a directory, so it cannot be scanned: $WT"

# `cd` fails on a directory whose contents cannot be read, which `-d` above does
# not catch — a mode the operator set, or one inherited from a parent.
PHYS="$(cd -- "$WT" 2>/dev/null && pwd -P)" || PHYS=""
[ -n "$PHYS" ] || die "cannot resolve the worktree to a physical path — unreadable directory: $WT"

# One pass over every process's working directory. -Fp emits `p<pid>`, -Fn emits
# `n<path>`, and lsof always emits the `f<fd>` line that opens each file record;
# with -d cwd the only file record a process has is its working directory, so
# every `n` line belongs to the `p` line above it.
LSOF_OUT="$(lsof -d cwd -Fpn 2>/dev/null)"
LSOF_RC=$?
[ "$LSOF_RC" -eq 0 ] \
  || die "lsof -d cwd -Fpn exited $LSOF_RC — the process table could not be read, which is not the same as finding nothing"

# is_at_or_under <path> — the guard's one decision. Both patterns are needed:
# the first for a process sitting in the worktree root, the second for one in
# any descendant of it. The quoted "$PHYS" is matched literally, so a worktree
# path containing a glob character cannot widen the match.
is_at_or_under() {
  case "$1" in
    "$PHYS"|"$PHYS"/*) return 0 ;;
    *) return 1 ;;
  esac
}

held=""
pid=""
while IFS= read -r line; do
  case "$line" in
    p*)
      pid="${line#p}"
      ;;
    n*)
      path="${line#n}"
      if is_at_or_under "$path"; then
        held="$held${held:+; }$pid $path"
      fi
      ;;
  esac
done <<EOF
$LSOF_OUT
EOF

if [ -n "$held" ]; then
  printf 'HELD: %s — %s\n' "$PHYS" "$held"
else
  printf 'CLEAR: %s\n' "$PHYS"
fi
exit 0
