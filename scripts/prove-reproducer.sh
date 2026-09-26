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
# EACH LEG IS A run-reproducer.sh CALL, NOT A RE-IMPLEMENTATION. Containment,
# the argv-exec barrier, the wall-clock bound, the kill sequence, the
# verdict vocabulary and the mutation-reproducer convention (read from the
# script itself, in whichever tree the leg runs in) are all run-reproducer's
# own contract, asserted by its Go port's tests (TestRunReproducer,
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
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
RUNNER="$SCRIPT_DIR/run-reproducer.sh"

usage() {
  printf 'usage: prove-reproducer.sh <worktree> <pre-fix-ref> <reproducer-path>\n' >&2
  printf '       <reproducer-path> is relative to <worktree>; the pre-fix leg runs\n' >&2
  printf '       it in a detached scratch worktree at <pre-fix-ref>, the post-fix\n' >&2
  printf '       leg against <worktree> itself\n' >&2
}

die() {
  printf 'prove-reproducer: %s\n' "$1" >&2
  exit 2
}

# Readability is checked before anything is sourced or run, the same rule
# run-reproducer.sh applies to its own dependencies: a missing dependency is
# "cannot answer", never a verdict on a reproducer that was never executed.
[ -r "$RUNNER" ] || die "cannot read $RUNNER — the per-leg verdicts have no runner"
[ -r "$SCRIPT_DIR/lib/reproducer-path.sh" ] || die "cannot read $SCRIPT_DIR/lib/reproducer-path.sh — the recorded-form refusals have no source"
# shellcheck source=lib/reproducer-path.sh
source "$SCRIPT_DIR/lib/reproducer-path.sh"

[ "$#" -eq 3 ] || { usage; exit 2; }

WT="$1"
REF="$2"
REL="$3"

[ -d "$WT" ] || die "not a directory: $WT"
git -C "$WT" rev-parse --is-inside-work-tree >/dev/null 2>&1 \
  || die "not a git worktree: $WT"

PRE_SHA="$(git -C "$WT" rev-parse --verify --quiet "${REF}^{commit}")" \
  || die "pre-fix ref does not resolve to a commit: $REF"
[ -n "$PRE_SHA" ] || die "pre-fix ref does not resolve to a commit: $REF"

# The reproducer path follows the recorded-form refusals at the door, single-
# sourced in lib/reproducer-path.sh: relative, no leading dash, no `..`
# segment. The copy step resolves it against the scratch root, so a `..` here
# would escape the scratch the same way F17's absolute ROOT escaped the live
# tree.
if ! REASON="$(reproducer_path_refusal "$REL")"; then
  die "$REASON"
fi
[ -f "$WT/$REL" ] || die "no reproducer file at $REL in $WT"
[ -x "$WT/$REL" ] || die "reproducer is not executable: $WT/$REL"

BASE="$(mktemp -d "${TMPDIR:-/tmp}/prove-reproducer.XXXXXX")" \
  || die "no writable temporary directory"
SCRATCH="$BASE/scratch"
LEG1_OUT="$BASE/leg1.out"
LEG2_OUT="$BASE/leg2.out"

cleanup() {
  # `|| true` throughout: cleanup runs on every path, including those where
  # the scratch was never created or git is mid-failure, and a cleanup
  # failure must never mask the verdict the script already printed.
  git -C "$WT" worktree remove --force "$SCRATCH" >/dev/null 2>&1 || true
  rm -rf "$BASE"
}
trap cleanup EXIT

if ! git -C "$WT" worktree add --detach --quiet "$SCRATCH" "$PRE_SHA" >"$BASE/add.out" 2>&1; then
  cat "$BASE/add.out" >&2
  die "could not materialize the scratch worktree at $PRE_SHA"
fi

# The copy steps are guarded explicitly: under `set -e` an unguarded
# failure here would exit 1 — the "proof did not hold" verdict code, with
# no leg run and no PROOF FAILED line printed. A plumbing failure is the
# documented cannot-answer 2, never a verdict.
mkdir -p "$SCRATCH/$(dirname "$REL")" || die "cannot create $SCRATCH/$(dirname "$REL") in the scratch worktree"
cp -p "$WT/$REL" "$SCRATCH/$REL" || die "cannot copy $WT/$REL into the scratch worktree"
[ -x "$SCRATCH/$REL" ] || die "the scratch copy lost its executable bit: $SCRATCH/$REL"

set +e
"$RUNNER" "$SCRATCH" "$REL" >"$LEG1_OUT" 2>&1
LEG1=$?
"$RUNNER" "$WT" "$REL" >"$LEG2_OUT" 2>&1
LEG2=$?
set -e

printf '== pre-fix leg — scratch worktree at %s\n' "$PRE_SHA"
cat "$LEG1_OUT"
printf '== pre-fix leg exit: %s\n' "$LEG1"
printf '== post-fix leg — %s\n' "$WT"
cat "$LEG2_OUT"
printf '== post-fix leg exit: %s\n' "$LEG2"

for leg in "$LEG1" "$LEG2"; do
  case "$leg" in
    0|1) ;;
    *)
      printf 'prove-reproducer: cannot spend the proof — run-reproducer exited %s on a leg (2 refused, 3 unverifiable, 4 cannot answer)\n' "$leg" >&2
      exit 2
      ;;
  esac
done

if [ "$LEG1" -eq 0 ] && [ "$LEG2" -eq 1 ]; then
  printf 'PROOF HELD — pre-fix: defect demonstrated (exit 0); post-fix: defect not demonstrated (exit 1)\n'
  exit 0
fi

if [ "$LEG1" -ne 0 ]; then
  printf 'PROOF FAILED — pre-fix leg read defect not demonstrated: the reproducer does not demonstrate the defect at %s\n' "$PRE_SHA"
fi
if [ "$LEG2" -ne 1 ]; then
  printf 'PROOF FAILED — post-fix leg still reads defect demonstrated: the fix does not stop this reproducer\n'
fi
exit 1
