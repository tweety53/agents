#!/usr/bin/env bash
# mutate-and-verify.sh <patch-file> <harness>...
#
# Mechanizes the review panel's fix-round mutation-proof loop
# (skills/flow/review-panel.md, "The fix round mutation-proves what it
# changed"): apply a mutation, run one or more scripts/test-*.sh-shaped
# harnesses before and after, report the new-failure set per harness,
# flag a suspiciously broad blast radius rather than reporting it as a
# strong pass, then unconditionally restore and verify the touched files
# are clean. WHICH mechanism to mutate and whether a survivor is real or
# equivalent stay the caller's own judgment calls — this script only does
# the mechanics. Commit the fix before the first flip: a touched file
# carrying uncommitted changes is refused outright (exit 2, nothing
# mutated), and the unconditional restore is `git checkout --`, which
# reverts the file to its index state — commit-first is what makes that
# restore well-defined. A harness's after-run is proof only once the
# mutation is asserted landed, the mutated bytes read back from disk the
# ones the patch wrote.
#
# Exit codes (mirrors run-reproducer.sh's own philosophy: the code reports
# this script's own mechanics, never a verdict on the mutation):
#   0  ran clean — the patch applied, every harness ran before and after,
#      the mutation was reverted, and the touched files were verified clean
#      afterward. The per-harness verdict (surviving mutant / caught /
#      suspicious blast radius) is in the report body, never the exit code.
#   2  refused before mutating anything — the patch does not apply cleanly,
#      a file it touches already has uncommitted changes, or a named
#      harness does not exist or is not executable — OR post-restore
#      drift: the touched files restored clean but the tree no longer
#      matches its pre-mutation snapshot (a new stash entry or an
#      unexpected status line, each named in the report).
#   3  could not fully restore — after `git checkout --` the touched files
#      are still not clean; the residual `git status --porcelain` output is
#      printed and the files may still be mutated. Takes precedence over
#      exit 2 when both drift kinds fire.
#   4  cannot answer — bad usage, not inside a git worktree, or a harness
#      produced no readable `ok:`/`FAIL:` line at all on either run.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 4 (this script's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "mutate-and-verify: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 4
}
flow_guard_exec mutate-and-verify 4 "mutate-and-verify:" "$@"
