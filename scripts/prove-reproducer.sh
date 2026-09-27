#!/usr/bin/env bash
# prove-reproducer.sh <worktree> <pre-fix-ref> <reproducer-path>
#
# Mechanizes KAN-663's two-direction reproducer proof: after a reproducer is
# authored or repaired it is executed twice — once against a scratch worktree
# at the pre-fix commit, where it must read defect demonstrated, and once
# against the fix, where it must read defect not demonstrated — and both
# exits are recorded. KAN-580's three authoring defects (a grep window too
# narrow to see the fixed call sites, an inverted exit-code convention, a
# hardcoded absolute ROOT path that made every "pre-fix" check silently run
# against the already-fixed live worktree) were each invisible to every
# self-report and caught only by this loop run by hand.
#
# THE PRE-FIX TREE IS A DETACHED SCRATCH WORKTREE, NEVER A CHECKOUT OR A
# MUTATION OF THE LIVE TREE. `git worktree add --detach` materializes the
# pre-fix commit in a throwaway directory under TMPDIR, the reproducer is
# copied to the SAME worktree-relative path inside it, and the leg runs with
# its working directory set to that scratch — so the cwd contract holds and
# no absolute path is ever needed. F17's ROOT class is then structural: a
# script that hardcodes its own tree cannot reach past the scratch, because
# the scratch is the tree the leg runs in. The scratch is removed on every
# path out of this script, failed proof or not.
#
# EACH LEG IS A run-reproducer CALL, NOT A RE-IMPLEMENTATION -- in-process,
# the Go run-reproducer (stats/internal/guard/runreproducer.go) called with
# the argv run-reproducer.sh takes and its exit code read as that script's
# exit. Containment, the argv-exec barrier, the wall-clock bound, the kill
# sequence, the verdict vocabulary and the mutation-reproducer convention
# (read from the script itself, in whichever tree the leg runs in) are all
# run-reproducer's own contract, asserted by its Go port's tests (TestRunReproducer,
# stats/internal/guard/runreproducer_test.go); this script composes two
# calls of it and judges the pair. A refused or unverifiable leg (exit 2 or
# 3) or a leg that cannot answer at all (4) spends the proof: nothing is
# claimed about the defect either way.
#
# Exit codes:
#   0  proof held — the pre-fix leg read defect demonstrated and the
#      post-fix leg read defect not demonstrated; both runs' labeled output
#      and exits are printed above the verdict.
#   1  the proof did not hold — a leg read the wrong demonstration; the
#      verdict names which leg and what it read.
#   2  cannot answer — usage, an unresolvable <pre-fix-ref>, an unusable
#      <reproducer-path>, a git plumbing failure, or a leg run-reproducer
#      refused (2), could not verdict (3) or could not answer (4). Never a
#      verdict on the defect.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this script's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
. "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "prove-reproducer: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec prove-reproducer 2 "prove-reproducer:" "$@"
