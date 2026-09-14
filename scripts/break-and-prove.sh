#!/usr/bin/env bash
# break-and-prove.sh [--clean <command>] <file> (--sed <expr> | --patch <patchfile>) \
#   -- <test-command> [<args>...]
#
# Mechanizes KAN-329's break-it/prove-it loop for config guards, so a fix
# round's evidence is uniform and quotable instead of prose each agent
# composes by hand: apply the mutation to <file>, force a clean re-run of
# <test-command>, assert it exits non-zero, restore <file>, force another
# clean re-run, assert it exits zero, and print both runs' observed output.
# A fix round that skips the proof is visible because the script produced
# no output.
#
# THE RESTORE IS A PRE-MUTATION SNAPSHOT, NEVER `git checkout --`. One of
# the observed failures this script exists to prevent is an agent whose
# `git checkout --` restore also reverted the uncommitted edits it had
# made on the same file. The file's bytes are copied to a temp snapshot
# before anything mutates, and the restore writes those bytes back — so
# uncommitted edits on <file> itself survive the loop, and the script
# deliberately does NOT refuse a dirty <file> the way mutate-and-verify.sh
# does. No git checkout call appears in this script, structurally.
#
# THE CLEAN RE-RUN IS THE SCRIPT'S JOB, not the caller's memory. Gradle
# silently skips a re-run when only a compose or YAML file changed — those
# are not declared task inputs — and an agent that forgot `:app:cleanTest`
# recorded a stale green as evidence. `--clean <command>` runs before EACH
# test invocation, from the repository root, through `sh -c`; the command
# is the operator's own (it is printed verbatim before it runs) and is the
# one place this script evaluates a string rather than an argv vector.
# Omit it for runners with no skip-to-green trap.
#
# THE TEST COMMAND IS AN ARGV VECTOR, NEVER A STRING THROUGH A SHELL, and
# it runs from the repository root with its arguments passed verbatim —
# give it paths relative to that root, the same way run-reproducer.sh
# passes its reproducer argv through untouched.
#
# WHERE THE TARGET MAY LIVE IS A PROPERTY OF THE MUTATION KIND, and the
# asymmetry is deliberate: a `--sed` target may sit outside the repository
# (the snapshot restore needs no git), while a `--patch` target must not
# (git apply is repository-bound).
#
# Exit codes (the same philosophy as mutate-and-verify.sh and
# run-reproducer.sh: the code reports this script's own mechanics and the
# proof's legs, never a richer verdict):
#   0  proof held — the mutation was applied, run 1 exited non-zero, the
#      file was restored byte-exact, and run 2 exited zero. Both runs'
#      output is printed above the verdict in labeled blocks.
#   1  the proof did not hold — run 1 exited 0 (the mutation did not break
#      the test: a guard gap or a stale-cache skip; its output is printed),
#      or run 2 exited non-zero after a byte-exact restore (the test fails
#      independently of the mutation). Which leg failed is named in the
#      report.
#   2  refused before mutating anything — <file> is missing or not
#      readable, the patch does not apply cleanly, the patch touches files
#      other than <file>, a --patch target lies outside the repository, or
#      the sed expression fails or changes nothing — OR post-restore drift:
#      the tree no longer matches its pre-mutation snapshot (a new stash
#      entry or an unexpected status line, each named in the report, per
#      scripts/lib/post-mutation-check.sh).
#   3  could not restore — <file> is not byte-identical to its pre-mutation
#      snapshot after the restore; the file may still be mutated. Takes
#      precedence over exit 2 when drift also fired.
#   4  cannot answer — bad usage, not inside a git worktree, an unreadable
#      patch file, the test command exited 126 or 127 (an exec failure or
#      the command's own such exit — either way it is not usable as
#      evidence), or the --clean command itself failed.
#
# Bash 3.2 is the floor (this repository's, per run-guard-tests.sh's own
# header): indexed arrays only, no associative arrays, no `wait -n`.
set -euo pipefail

usage_fail() {
  echo "break-and-prove: usage: break-and-prove.sh [--clean <command>] <file> (--sed <expr> | --patch <patchfile>) -- <test-command> [<args>...]" >&2
  exit 4
}

cannot_answer() {
  echo "break-and-prove: cannot answer — $1" >&2
  exit 4
}

refuse() {
  echo "break-and-prove: refused — $1 — nothing was mutated" >&2
  exit 2
}

CLEAN_CMD=""
FILE_ARG=""
MUTATION_KIND=""
MUTATION_VALUE=""
TEST_CMD=()

while [ "$#" -gt 0 ]; do
  case "$1" in
    --clean)
      [ "$#" -ge 2 ] || usage_fail
      [ -z "$CLEAN_CMD" ] || usage_fail
      CLEAN_CMD="$2"
      shift 2
      ;;
    --sed|--patch)
      [ "$#" -ge 2 ] || usage_fail
      [ -z "$MUTATION_KIND" ] || usage_fail
      MUTATION_KIND="${1#--}"
      MUTATION_VALUE="$2"
      shift 2
      ;;
    --)
      shift
      TEST_CMD=("$@")
      break
      ;;
    *)
      [ -z "$FILE_ARG" ] || usage_fail
      FILE_ARG="$1"
      shift
      ;;
  esac
done

[ -n "$FILE_ARG" ] || usage_fail
[ -n "$MUTATION_KIND" ] || usage_fail
[ "${#TEST_CMD[@]}" -ge 1 ] || usage_fail

ORIG_PWD="$(pwd)"

# Resolve the file and patch arguments to absolute paths against the
# caller's own cwd BEFORE this script changes directory to the repo root —
# a relative argument must keep meaning what the caller meant by it, not be
# reinterpreted against a different directory. to_abs and the
# usage_fail/cannot_answer/refuse trio above deliberately repeat
# mutate-and-verify.sh's rather than moving into scripts/lib/: tiny,
# program-name-parameterized helpers, and an extraction would be the wrong
# abstraction for helpers this small.
to_abs() {
  case "$1" in
    /*) printf '%s\n' "$1" ;;
    *) printf '%s\n' "$ORIG_PWD/$1" ;;
  esac
}

FILE_ABS="$(to_abs "$FILE_ARG")"
PATCH_FILE=""
if [ "$MUTATION_KIND" = "patch" ]; then
  PATCH_FILE="$(to_abs "$MUTATION_VALUE")"
fi

# Step 1: resolve the repo root; exit 4 if this is not a git worktree.
REPO_ROOT="$(cd -- "$ORIG_PWD" && git rev-parse --show-toplevel 2>/dev/null)" \
  || cannot_answer "not inside a git worktree: $ORIG_PWD"
cd -- "$REPO_ROOT"

# Step 2: the target file must exist and be readable — the loop mutates an
# existing config file; creating one is not part of the proof.
[ -f "$FILE_ABS" ] && [ -r "$FILE_ABS" ] \
  || refuse "target file does not exist or is not readable: $FILE_ARG"

# Step 3: the patch, when given, must exist and be readable.
if [ -n "$PATCH_FILE" ]; then
  [ -f "$PATCH_FILE" ] && [ -r "$PATCH_FILE" ] \
    || cannot_answer "cannot read patch file: $MUTATION_VALUE"
fi

# Step 4: git apply --check — refuse (exit 2) without touching anything.
if [ -n "$PATCH_FILE" ]; then
  APPLY_CHECK_ERR=""
  set +e
  APPLY_CHECK_ERR="$(git apply --check "$PATCH_FILE" 2>&1)"
  APPLY_CHECK_RC=$?
  set -e
  if [ "$APPLY_CHECK_RC" -ne 0 ]; then
    echo "break-and-prove: git apply --check failed:" >&2
    printf '%s\n' "$APPLY_CHECK_ERR" >&2
    refuse "patch does not apply cleanly: $MUTATION_VALUE"
  fi
fi

# Step 5: the patch must touch exactly <file> — the operator named the one
# file being proven, and the snapshot/restore contract covers exactly that
# file, so a patch reaching further is refused rather than silently
# narrowed.
if [ -n "$PATCH_FILE" ]; then
  TOUCHED_COUNT=0
  TOUCHED_FILE=""
  while IFS= read -r line; do
    [ -n "$line" ] || continue
    TOUCHED_COUNT=$((TOUCHED_COUNT + 1))
    TOUCHED_FILE="$(printf '%s\n' "$line" | cut -f3)"
  done < <(git apply --numstat "$PATCH_FILE")
  [ "$TOUCHED_COUNT" -eq 1 ] || refuse "patch touches $TOUCHED_COUNT files — it must touch exactly $FILE_ARG"
  # Compare canonicalized paths: a caller cwd reached through a symlink
  # (macOS's /tmp → /private/tmp) yields an absolute path that shares no
  # string prefix with git's physical toplevel.
  FILE_CANON="$(cd -- "$(dirname -- "$FILE_ABS")" && pwd -P)/$(basename -- "$FILE_ABS")"
  REPO_CANON="$(cd -- "$REPO_ROOT" && pwd -P)"
  case "$FILE_CANON" in
    "$REPO_CANON"/*) FILE_REL="${FILE_CANON#"$REPO_CANON"/}" ;;
    *) refuse "target file is outside the repository: $FILE_ARG" ;;
  esac
  [ "$TOUCHED_FILE" = "$FILE_REL" ] \
    || refuse "patch touches $TOUCHED_FILE, not $FILE_REL"
fi

# Step 6: temp workspace and the pre-mutation byte snapshot — the restore
# source, never `git checkout --`.
WORK="$(mktemp -d "${TMPDIR:-/tmp}/break-and-prove.XXXXXX")"
SNAP="$WORK/snapshot"
cp "$FILE_ABS" "$SNAP"

# Step 7: snapshot the tree the mutation starts from — status and stash
# list — so the EXIT trap can tell post-restore residue from the state the
# run actually left.
. "$(dirname "$0")/lib/post-mutation-check.sh"
TREE_SNAPSHOT="$(snapshot_tree_state "$REPO_ROOT")"

# Step 8: install the EXIT trap now, before anything mutates. If the
# mutation was applied and the main flow has not yet verified the restore,
# the trap restores from the snapshot and holds exit 3 until the file is
# byte-identical again; any tree drift forces exit 2 unless exit 3 already
# took precedence — the same precedence mutate-and-verify.sh applies.
APPLIED=0
RESTORED=0
on_exit() {
  local rc=$? drift
  if [ "$APPLIED" -eq 1 ]; then
    if [ "$RESTORED" -eq 0 ]; then
      cat "$SNAP" > "$FILE_ABS" 2>/dev/null || true
    fi
    if ! cmp -s "$SNAP" "$FILE_ABS"; then
      echo "break-and-prove: could not restore $FILE_ARG byte-exact — residual content differs" >&2
      rc=3
    elif [ "$RESTORED" -eq 0 ]; then
      echo "break-and-prove: $FILE_ARG restored byte-exact (restored by the exit trap)"
    fi
    drift="$(check_tree_restored "$REPO_ROOT" "$TREE_SNAPSHOT" || true)"
    if [ -n "$drift" ]; then
      echo "break-and-prove: post-restore drift detected:" >&2
      printf '%s\n' "$drift" >&2
      if [ "$rc" -ne 3 ]; then rc=2; fi
    fi
  fi
  rm -rf "$WORK" 2>/dev/null || true
  exit "$rc"
}
trap on_exit EXIT

# Step 9: apply the mutation.
echo "break-and-prove: proving $FILE_ARG"
case "$MUTATION_KIND" in
  sed)
    MUTATED="$WORK/mutated"
    set +e
    sed -e "$MUTATION_VALUE" "$FILE_ABS" > "$MUTATED"
    SED_RC=$?
    set -e
    if [ "$SED_RC" -ne 0 ]; then
      refuse "sed failed with exit $SED_RC: $MUTATION_VALUE"
    fi
    if cmp -s "$MUTATED" "$FILE_ABS"; then
      refuse "sed expression changes nothing — the mutation must alter the file: $MUTATION_VALUE"
    fi
    echo "break-and-prove: mutating via sed: $MUTATION_VALUE"
    cat "$MUTATED" > "$FILE_ABS"
    ;;
  patch)
    echo "break-and-prove: mutating via patch: $MUTATION_VALUE"
    git apply "$PATCH_FILE"
    ;;
  *)
    usage_fail
    ;;
esac
APPLIED=1

# run_clean_before_leg — the --clean command runs before EACH test
# invocation; its own failure means neither leg is evidence, so it answers
# exit 4 rather than being read as either leg's result.
run_clean_before_leg() {
  [ -z "$CLEAN_CMD" ] && return 0
  echo "break-and-prove: clean — sh -c $(printf '%q' "$CLEAN_CMD")"
  set +e
  sh -c "$CLEAN_CMD"
  CLEAN_RC=$?
  set -e
  if [ "$CLEAN_RC" -ne 0 ]; then
    cannot_answer "clean command exited $CLEAN_RC — the proof environment is broken"
  fi
}

# run_leg <label> — forces the clean re-run, then runs the test command's
# argv vector from the repository root, capturing its combined output.
# Sets LEG_RC and prints the output in a labeled, quotable block.
run_leg() {
  local label="$1"
  run_clean_before_leg
  echo "break-and-prove: --- begin $label output ---"
  set +e
  LEG_OUT="$("${TEST_CMD[@]}" 2>&1)"
  LEG_RC=$?
  set -e
  if [ -n "$LEG_OUT" ]; then
    printf '%s\n' "$LEG_OUT"
  else
    echo "(no output)"
  fi
  echo "break-and-prove: --- end $label output ---"
}

# Step 10: run 1, against the mutated file — it must fail.
run_leg "run 1 (mutated)"
echo "break-and-prove: run 1 exited $LEG_RC"
# The 126/127 guard below appears twice, once per leg, on purpose (WET):
# whichever leg the bad exit lands on, the answer is the same
# cannot-answer, and a shared helper would bury that behind indirection.
if [ "$LEG_RC" -eq 126 ] || [ "$LEG_RC" -eq 127 ]; then
  cannot_answer "test command exited $LEG_RC (126/127) — an exec failure or its own such exit, not usable as evidence"
fi
if [ "$LEG_RC" -eq 0 ]; then
  echo "break-and-prove: run 1 exited 0 — the proof did not hold: the mutation did not break the test" >&2
  exit 1
fi

# Step 11: restore from the pre-mutation snapshot, verified byte-exact
# before run 2 is trusted.
cat "$SNAP" > "$FILE_ABS" 2>/dev/null || true
cmp -s "$SNAP" "$FILE_ABS" || exit 3
RESTORED=1
echo "break-and-prove: $FILE_ARG restored byte-exact"

# Step 12: run 2, against the restored file — it must pass.
run_leg "run 2 (restored)"
echo "break-and-prove: run 2 exited $LEG_RC"
if [ "$LEG_RC" -eq 126 ] || [ "$LEG_RC" -eq 127 ]; then
  cannot_answer "test command exited $LEG_RC (126/127) — an exec failure or its own such exit, not usable as evidence"
fi
if [ "$LEG_RC" -ne 0 ]; then
  echo "break-and-prove: run 2 exited $LEG_RC — the proof did not hold: the test fails on the restored file" >&2
  exit 1
fi

echo "break-and-prove: proof held — mutated run failed (exit was non-zero), restored run passed, $FILE_ARG restored byte-exact"
exit 0
