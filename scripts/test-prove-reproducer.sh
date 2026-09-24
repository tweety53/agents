#!/usr/bin/env bash
# Assertion harness for prove-reproducer.sh.
#
# Builds a temporary real git repository per case, commits a defect state,
# writes a reproducer at a worktree-relative path, commits the fix, and runs
# the two-direction proof against the pair — asserting the exit code, the
# leg the verdict names, and (on refusals) that no scratch worktree was ever
# created, so nothing ran. Follows test-run-reproducer.sh's shape: a case_N
# section per case, a counter, and a non-zero exit when any case fails.
#
# `set -e` as well as `-u`/`pipefail`, matching this repository's sibling
# harnesses. `set +e`/`set -e` bracket every call expected to exit non-zero,
# because that is the whole point of most cases here and `-e` would
# otherwise abort the suite on the first one.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/prove-reproducer.sh"
FAILED=0

REPOS=()
cleanup() {
  [ "${#REPOS[@]}" -eq 0 ] && return 0
  for repo in "${REPOS[@]}"; do
    rm -rf "$repo"
  done
}
trap cleanup EXIT

# make_repo -> prints a fresh real git repository path with an initial
# commit already on it. Fails loudly rather than returning an empty path: a
# silent empty path from a failed mktemp is the same stale-sandbox hazard
# test-run-reproducer.sh's own make_worktree guards against.
make_repo() {
  local repo
  repo="$(mktemp -d "${TMPDIR:-/tmp}/prove-reproducer-test.XXXXXX")" || {
    printf 'make_repo: mktemp failed — aborting suite rather than continuing with a stale path\n' >&2
    exit 1
  }
  REPOS+=("$repo")
  git -C "$repo" init -q
  git -C "$repo" -c user.name=test -c user.email=test@test commit -q --allow-empty -m init
  printf '%s' "$repo"
}

# defect_fixture <repo> <body> -> writes check.txt carrying <body> and
# commits it as the pre-fix (defect-present) state; prints the commit sha.
defect_fixture() {
  local repo="$1" body="$2"
  printf '%s\n' "$body" > "$repo/check.txt"
  git -C "$repo" add check.txt
  git -C "$repo" -c user.name=test -c user.email=test@test commit -q -m defect
  git -C "$repo" rev-parse HEAD
}

# fix_commit <repo> <body> -> rewrites check.txt with <body> and commits the
# fix on top; prints the fix commit sha.
fix_commit() {
  local repo="$1" body="$2"
  printf '%s\n' "$body" > "$repo/check.txt"
  git -C "$repo" add check.txt
  git -C "$repo" -c user.name=test -c user.email=test@test commit -q -m fix
  git -C "$repo" rev-parse HEAD
}

# reproducer_fixture <repo> <relative-path> <body> -> writes an executable
# script at <relative-path> inside <repo>, with a shebang plus <body>.
reproducer_fixture() {
  local repo="$1" rel="$2" body="$3"
  mkdir -p "$repo/$(dirname "$rel")"
  {
    printf '#!/bin/sh\n'
    printf '%s\n' "$body"
  } > "$repo/$rel"
  chmod +x "$repo/$rel"
}

# scratch_count <repo> -> how many worktrees git registers beyond the main
# one. A refusal must create none: a second entry after a refused run would
# mean the scratch was built and the proof started anyway.
scratch_count() {
  git -C "$1" worktree list --porcelain | grep -c '^worktree '
}

# ---------------------------------------------------------------- case 1
# The proof holds: a reproducer that exits non-zero on the defect-present
# tree and zero on the fix reads demonstrated pre-fix and not-demonstrated
# post-fix, and the script exits 0 naming both legs. The scratch worktree is
# gone when the script returns.
case_1() {
  local repo pre sha out rc
  repo="$(make_repo)"
  pre="$(defect_fixture "$repo" "old-behaviour")"
  reproducer_fixture "$repo" ".superpowers/sdd/reproducers/0-primary-1.sh" \
    "exec sh -c '! grep -q old-behaviour check.txt'"
  fix_commit "$repo" "new-behaviour" >/dev/null
  set +e
  out="$("$GUARD" "$repo" "$pre" ".superpowers/sdd/reproducers/0-primary-1.sh" 2>&1)"
  rc=$?
  set -e
  if [ "$rc" -ne 0 ]; then
    printf 'case_1: expected exit 0, got %s\n%s\n' "$rc" "$out"
    FAILED=1
    return 0
  fi
  case "$out" in
    *pre-fix*post-fix*|*post-fix*pre-fix*) : ;;
    *) printf 'case_1: output does not name both legs:\n%s\n' "$out"; FAILED=1; return 0 ;;
  esac
  if [ "$(scratch_count "$repo")" -ne 1 ]; then
    printf 'case_1: scratch worktree survived the run:\n%s\n' "$(git -C "$repo" worktree list)"
    FAILED=1
  fi
}

# ---------------------------------------------------------------- case 2
# The pre-fix leg reads not demonstrated — the inverted-exit class (F17's
# first defect): a reproducer that exits 0 on the defect-present tree proves
# nothing, and exit 1 names the pre-fix leg.
case_2() {
  local repo pre out rc
  repo="$(make_repo)"
  pre="$(defect_fixture "$repo" "old-behaviour")"
  reproducer_fixture "$repo" ".superpowers/sdd/reproducers/0-primary-1.sh" \
    "exec grep -q old-behaviour check.txt"
  fix_commit "$repo" "new-behaviour" >/dev/null
  set +e
  out="$("$GUARD" "$repo" "$pre" ".superpowers/sdd/reproducers/0-primary-1.sh" 2>&1)"
  rc=$?
  set -e
  if [ "$rc" -ne 1 ]; then
    printf 'case_2: expected exit 1, got %s\n%s\n' "$rc" "$out"
    FAILED=1
    return 0
  fi
  case "$out" in
    *pre-fix*) : ;;
    *) printf 'case_2: verdict does not name the pre-fix leg:\n%s\n' "$out"; FAILED=1 ;;
  esac
}

# ---------------------------------------------------------------- case 3
# The post-fix leg still reads demonstrated — the too-narrow-window class
# (F4): a reproducer that exits non-zero on both trees cannot confirm the
# fix, and exit 1 names the post-fix leg.
case_3() {
  local repo pre out rc
  repo="$(make_repo)"
  pre="$(defect_fixture "$repo" "old-behaviour")"
  reproducer_fixture "$repo" ".superpowers/sdd/reproducers/0-primary-1.sh" \
    "exec sh -c 'exit 7'"
  fix_commit "$repo" "new-behaviour" >/dev/null
  set +e
  out="$("$GUARD" "$repo" "$pre" ".superpowers/sdd/reproducers/0-primary-1.sh" 2>&1)"
  rc=$?
  set -e
  if [ "$rc" -ne 1 ]; then
    printf 'case_3: expected exit 1, got %s\n%s\n' "$rc" "$out"
    FAILED=1
    return 0
  fi
  case "$out" in
    *post-fix*) : ;;
    *) printf 'case_3: verdict does not name the post-fix leg:\n%s\n' "$out"; FAILED=1 ;;
  esac
}

# ---------------------------------------------------------------- case 4
# An unresolvable pre-fix ref is refused (exit 2) and the reproducer never
# runs — no scratch worktree is created at all.
case_4() {
  local repo out rc
  repo="$(make_repo)"
  defect_fixture "$repo" "old-behaviour" >/dev/null
  reproducer_fixture "$repo" ".superpowers/sdd/reproducers/0-primary-1.sh" \
    "exec sh -c 'exit 1'"
  fix_commit "$repo" "new-behaviour" >/dev/null
  set +e
  out="$("$GUARD" "$repo" "0123456789abcdef0123456789abcdef01234567" ".superpowers/sdd/reproducers/0-primary-1.sh" 2>&1)"
  rc=$?
  set -e
  if [ "$rc" -ne 2 ]; then
    printf 'case_4: expected exit 2, got %s\n%s\n' "$rc" "$out"
    FAILED=1
    return 0
  fi
  if [ "$(scratch_count "$repo")" -ne 1 ]; then
    printf 'case_4: a scratch worktree was created for an unresolvable ref\n'
    FAILED=1
  fi
}

# ---------------------------------------------------------------- case 5
# An absolute reproducer path is refused (exit 2) before anything runs — the
# cwd contract holds at the door, not only inside run-reproducer.sh.
case_5() {
  local repo pre out rc
  repo="$(make_repo)"
  pre="$(defect_fixture "$repo" "old-behaviour")"
  reproducer_fixture "$repo" ".superpowers/sdd/reproducers/0-primary-1.sh" \
    "exec sh -c 'exit 1'"
  fix_commit "$repo" "new-behaviour" >/dev/null
  set +e
  out="$("$GUARD" "$repo" "$pre" "$repo/.superpowers/sdd/reproducers/0-primary-1.sh" 2>&1)"
  rc=$?
  set -e
  if [ "$rc" -ne 2 ]; then
    printf 'case_5: expected exit 2, got %s\n%s\n' "$rc" "$out"
    FAILED=1
    return 0
  fi
  if [ "$(scratch_count "$repo")" -ne 1 ]; then
    printf 'case_5: a scratch worktree was created for an absolute reproducer path\n'
    FAILED=1
  fi
}

# ---------------------------------------------------------------- case 6
# The mutation-reproducer convention is read from the script itself, in
# whichever tree the leg runs in: a reproducer declaring the convention and
# exiting 0 on the defect-present tree reads demonstrated pre-fix and —
# exiting non-zero on the fix — not-demonstrated post-fix, so the proof
# holds (exit 0) under the inverted reading too.
case_6() {
  local repo pre out rc
  repo="$(make_repo)"
  pre="$(defect_fixture "$repo" "old-behaviour")"
  reproducer_fixture "$repo" ".superpowers/sdd/reproducers/0-mutation-1.sh" \
    "# mutation-reproducer
exec grep -q old-behaviour check.txt"
  fix_commit "$repo" "new-behaviour" >/dev/null
  set +e
  out="$("$GUARD" "$repo" "$pre" ".superpowers/sdd/reproducers/0-mutation-1.sh" 2>&1)"
  rc=$?
  set -e
  if [ "$rc" -ne 0 ]; then
    printf 'case_6: expected exit 0 under the mutation-reproducer convention, got %s\n%s\n' "$rc" "$out"
    FAILED=1
  fi
}

# ---------------------------------------------------------------- case 7
# A `..` segment in the reproducer path is refused (exit 2) before anything
# runs — the copy step must never resolve outside the worktree either.
case_7() {
  local repo pre out rc
  repo="$(make_repo)"
  pre="$(defect_fixture "$repo" "old-behaviour")"
  reproducer_fixture "$repo" ".superpowers/sdd/reproducers/0-primary-1.sh" \
    "exec sh -c 'exit 1'"
  fix_commit "$repo" "new-behaviour" >/dev/null
  set +e
  out="$("$GUARD" "$repo" "$pre" "../escape.sh" 2>&1)"
  rc=$?
  set -e
  if [ "$rc" -ne 2 ]; then
    printf 'case_7: expected exit 2, got %s\n%s\n' "$rc" "$out"
    FAILED=1
    return 0
  fi
  if [ "$(scratch_count "$repo")" -ne 1 ]; then
    printf 'case_7: a scratch worktree was created for a .. path\n'
    FAILED=1
  fi
}

case_1
case_2
case_3
case_4
case_5
case_6
case_7

if [ "$FAILED" -ne 0 ]; then
  printf 'test-prove-reproducer: FAILED\n'
  exit 1
fi
printf 'test-prove-reproducer: all cases passed\n'
