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
# the mechanics.
#
# Bash 3.2 is the floor (this repository's, per run-guard-tests.sh's own
# header): indexed arrays only, no associative arrays, no `wait -n`.
#
# Exit codes (mirrors run-reproducer.sh's own philosophy: the code reports
# this script's own mechanics, never a verdict on the mutation):
#   0  ran clean — the patch applied, every harness ran before and after,
#      the mutation was reverted, and the touched files were verified clean
#      afterward. The per-harness verdict (surviving mutant / caught /
#      suspicious blast radius) is in the report body, never the exit code.
#   2  refused before mutating anything — the patch does not apply cleanly,
#      a file it touches already has uncommitted changes, or a named
#      harness does not exist or is not executable.
#   3  could not fully restore — after `git checkout --` the touched files
#      are still not clean; the residual `git status --porcelain` output is
#      printed and the files may still be mutated.
#   4  cannot answer — bad usage, not inside a git worktree, or a harness
#      produced no readable `ok:`/`FAIL:` line at all on either run.
set -euo pipefail

usage_fail() {
  echo "mutate-and-verify: usage: mutate-and-verify.sh <patch-file> <harness>..." >&2
  exit 4
}

cannot_answer() {
  echo "mutate-and-verify: cannot answer — $1" >&2
  exit 4
}

refuse() {
  echo "mutate-and-verify: refused — $1 — nothing was mutated" >&2
  exit 2
}

[ "$#" -ge 2 ] || usage_fail

ORIG_PWD="$(pwd)"
PATCH_FILE_ARG="$1"
shift
HARNESS_ARGS=("$@")

# Resolve patch and harness arguments to absolute paths against the
# caller's own cwd BEFORE this script changes directory to the repo root
# below — a relative argument must keep meaning what the caller meant by
# it, not be reinterpreted against a different directory.
to_abs() {
  case "$1" in
    /*) printf '%s\n' "$1" ;;
    *) printf '%s\n' "$ORIG_PWD/$1" ;;
  esac
}

PATCH_FILE="$(to_abs "$PATCH_FILE_ARG")"

HARNESSES=()
for h in "${HARNESS_ARGS[@]}"; do
  HARNESSES+=("$(to_abs "$h")")
done

# Step 1: resolve the repo root; exit 4 if this is not a git worktree.
REPO_ROOT="$(cd -- "$ORIG_PWD" && git rev-parse --show-toplevel 2>/dev/null)" \
  || cannot_answer "not inside a git worktree: $ORIG_PWD"
cd -- "$REPO_ROOT"

# Step 2: the patch file must exist and be readable.
[ -f "$PATCH_FILE" ] && [ -r "$PATCH_FILE" ] \
  || cannot_answer "cannot read patch file: $PATCH_FILE_ARG"

# Step 3: git apply --check — refuse (exit 2) without touching anything.
APPLY_CHECK_ERR=""
set +e
APPLY_CHECK_ERR="$(git apply --check "$PATCH_FILE" 2>&1)"
APPLY_CHECK_RC=$?
set -e
if [ "$APPLY_CHECK_RC" -ne 0 ]; then
  echo "mutate-and-verify: git apply --check failed:" >&2
  printf '%s\n' "$APPLY_CHECK_ERR" >&2
  refuse "$PATCH_FILE_ARG does not apply cleanly"
fi

# Step 4: derive the touched file list from git apply --numstat's third
# column.
TOUCHED=()
while IFS= read -r path; do
  [ -n "$path" ] && TOUCHED+=("$path")
done < <(git apply --numstat "$PATCH_FILE" | cut -f3)

[ "${#TOUCHED[@]}" -gt 0 ] || cannot_answer "$PATCH_FILE_ARG names no touched file"

# Step 5: every named harness must exist and be executable.
for h in "${HARNESSES[@]}"; do
  [ -f "$h" ] || refuse "harness does not exist: $h"
  [ -x "$h" ] || refuse "harness is not executable: $h"
done

# Step 6: none of the touched files may already carry uncommitted changes —
# never conflate a pre-existing dirty file with the mutation's own diff at
# restore time.
DIRTY_STATUS="$(git status --porcelain -- "${TOUCHED[@]}")"
[ -z "$DIRTY_STATUS" ] || {
  echo "mutate-and-verify: touched files already have uncommitted changes:" >&2
  printf '%s\n' "$DIRTY_STATUS" >&2
  refuse "a file the patch touches is not clean"
}

# Step 7: install the EXIT trap now, before anything mutates. On any exit,
# if the patch was applied, best-effort restore the touched files and
# re-verify they are clean; if not, force this script's own exit code to 3
# — overriding whatever the main flow was already carrying — so a mutated
# file is never left behind on any exit path, this one included.
APPLIED=0
on_exit() {
  local rc=$?
  if [ "$APPLIED" -eq 1 ]; then
    git checkout -- "${TOUCHED[@]}" 2>/dev/null || true
    local residual
    residual="$(git status --porcelain -- "${TOUCHED[@]}" 2>/dev/null || true)"
    if [ -n "$residual" ]; then
      echo "mutate-and-verify: could not fully restore — residual status:" >&2
      printf '%s\n' "$residual" >&2
      rc=3
    elif [ "$rc" -eq 0 ]; then
      # Only claim success once the restore has actually been verified
      # clean — printing this before the restore ran would contradict a
      # residual-status failure reported moments later.
      echo "mutate-and-verify: ran clean — touched files restored"
    fi
  fi
  exit "$rc"
}
trap on_exit EXIT

# Step 8: helper that runs one harness, captures its stdout+stderr, and
# reports total `ok: ` count, total `FAIL: ` count, and the set of
# `FAIL:` case names. Sets HARNESS_OK_COUNT, HARNESS_FAIL_COUNT,
# HARNESS_FAIL_NAMES (indexed array); returns 1 (never exits) when the
# harness's output carries neither an `ok:` nor a `FAIL:` line, so the
# caller can name which pass and which harness before answering exit 4.
run_harness() {
  local harness="$1" out line
  set +e
  out="$("$harness" 2>&1)"
  set -e
  HARNESS_OK_COUNT=0
  HARNESS_FAIL_COUNT=0
  HARNESS_FAIL_NAMES=()
  while IFS= read -r line; do
    case "$line" in
      'ok: '*)
        HARNESS_OK_COUNT=$((HARNESS_OK_COUNT + 1))
        ;;
      'FAIL: '*)
        HARNESS_FAIL_COUNT=$((HARNESS_FAIL_COUNT + 1))
        HARNESS_FAIL_NAMES+=("${line#FAIL: }")
        ;;
    esac
  done <<< "$out"
  if [ "$HARNESS_OK_COUNT" -eq 0 ] && [ "$HARNESS_FAIL_COUNT" -eq 0 ]; then
    return 1
  fi
  return 0
}

# in_array <needle> <haystack-array-name-as-args...> — bash-3.2-safe
# membership test (no associative arrays).
in_array() {
  local needle="$1"
  shift
  local x
  for x in "$@"; do
    [ "$x" = "$needle" ] && return 0
  done
  return 1
}

echo "mutate-and-verify: patch $PATCH_FILE_ARG touches:"
for t in "${TOUCHED[@]}"; do
  echo "  $t"
done

# Step 9: baseline pass — every harness, before applying the patch.
BASE_OK=()
BASE_FAIL=()
# A per-harness FAIL-name set can't be an associative array on bash 3.2, so
# each harness's set is kept as one newline-joined string in a parallel
# indexed array (index i matches HARNESSES[i]), rebuilt into a real array
# again when needed. Newline, never space, is the join/split delimiter — a
# case name itself can contain spaces ("guard exits 0"), and splitting on
# IFS-default whitespace would shred a multi-word name into several bogus
# single-word ones.
BASE_FAIL_NAMES_JOINED=()

echo
echo "mutate-and-verify: baseline pass (before mutation)"
i=0
for h in "${HARNESSES[@]}"; do
  hname="$(basename "$h")"
  if ! run_harness "$h"; then
    cannot_answer "harness $hname produced no ok:/FAIL: line on the baseline pass"
  fi
  BASE_OK+=("$HARNESS_OK_COUNT")
  BASE_FAIL+=("$HARNESS_FAIL_COUNT")
  if [ "${#HARNESS_FAIL_NAMES[@]}" -gt 0 ]; then
    BASE_FAIL_NAMES_JOINED+=("$(printf '%s\n' "${HARNESS_FAIL_NAMES[@]}")")
  else
    BASE_FAIL_NAMES_JOINED+=("")
  fi
  echo "  $hname: baseline ok=$HARNESS_OK_COUNT fail=$HARNESS_FAIL_COUNT"
  if [ "$HARNESS_FAIL_COUNT" -gt 0 ]; then
    echo "  $hname: baseline NOT clean — ${HARNESS_FAIL_NAMES[*]}"
  fi
  i=$((i + 1))
done

# Step 10: apply the mutation.
git apply "$PATCH_FILE"
APPLIED=1

# Step 11: mutated pass — every harness, after applying the patch.
MUT_OK=()
MUT_FAIL=()
MUT_FAIL_NAMES_JOINED=()

echo
echo "mutate-and-verify: mutated pass (after mutation)"
i=0
for h in "${HARNESSES[@]}"; do
  hname="$(basename "$h")"
  if ! run_harness "$h"; then
    cannot_answer "harness $hname produced no ok:/FAIL: line on the mutated pass"
  fi
  MUT_OK+=("$HARNESS_OK_COUNT")
  MUT_FAIL+=("$HARNESS_FAIL_COUNT")
  if [ "${#HARNESS_FAIL_NAMES[@]}" -gt 0 ]; then
    MUT_FAIL_NAMES_JOINED+=("$(printf '%s\n' "${HARNESS_FAIL_NAMES[@]}")")
  else
    MUT_FAIL_NAMES_JOINED+=("")
  fi
  echo "  $hname: mutated ok=$HARNESS_OK_COUNT fail=$HARNESS_FAIL_COUNT"
  i=$((i + 1))
done

# Step 13: resolve the blast-radius bound.
case "${MUTATE_AND_VERIFY_MAX_NEW_FAILURES:-}" in
  '') BOUND=5 ;;
  *[!0-9]*) BOUND=5 ;;
  *) BOUND="$MUTATE_AND_VERIFY_MAX_NEW_FAILURES" ;;
esac

echo
echo "mutate-and-verify: verdict (blast-radius bound: $BOUND)"

i=0
for h in "${HARNESSES[@]}"; do
  hname="$(basename "$h")"

  # Rebuild each harness's baseline/mutated FAIL-name arrays from the
  # newline-joined strings kept above — split on newline only, one read
  # per line, never `read -ra`'s default whitespace splitting, which would
  # shred a multi-word case name into several bogus single-word ones.
  base_names=()
  while IFS= read -r line; do
    [ -n "$line" ] && base_names+=("$line")
  done <<< "${BASE_FAIL_NAMES_JOINED[$i]:-}"
  mut_names=()
  while IFS= read -r line; do
    [ -n "$line" ] && mut_names+=("$line")
  done <<< "${MUT_FAIL_NAMES_JOINED[$i]:-}"

  # Step 12: new failures = mutated FAIL-name set minus baseline FAIL-name
  # set, a bash-3.2-safe loop-based set difference.
  new_failures=()
  if [ "${#mut_names[@]}" -gt 0 ]; then
    for n in "${mut_names[@]}"; do
      if [ "${#base_names[@]}" -eq 0 ] || ! in_array "$n" "${base_names[@]}"; then
        new_failures+=("$n")
      fi
    done
  fi
  new_count="${#new_failures[@]}"

  # Step 14: per-harness verdict.
  if [ "$new_count" -eq 0 ]; then
    echo "  $hname: surviving mutant — 0 new failures"
  elif [ "$new_count" -le "$BOUND" ]; then
    echo "  $hname: caught — $new_count new failure(s): ${new_failures[*]}"
  else
    echo "  $hname: FLAG suspicious blast radius — $new_count new failures (bound $BOUND): ${new_failures[*]}"
  fi

  i=$((i + 1))
done

echo
echo "mutate-and-verify: restoring touched files"

# Step 15's restore-and-verify — and the "ran clean" success report, printed
# only once the restore is actually verified — is performed by the EXIT trap
# installed at step 7; this script's own main flow ends here.
exit 0
