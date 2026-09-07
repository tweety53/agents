#!/usr/bin/env bash
# Assertion harness for scripts/lib/post-mutation-check.sh. Builds
# throwaway git repositories under a sandboxed TMPDIR and asserts the
# library's snapshot/check roundtrip: a clean tree matches its snapshot,
# a new stash entry is named, an unexpected status line is named, and a
# stash the snapshot recorded is not drift. Never touches the real
# repository tree.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/post-mutation-check.sh
. "$SCRIPT_DIR/lib/post-mutation-check.sh"

FAILURES=0
fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

REPOS=()
cleanup() {
  [ "${#REPOS[@]}" -eq 0 ] && return 0
  for p in "${REPOS[@]}"; do rm -rf "$p"; done
}
trap cleanup EXIT

# new_repo — print a fresh throwaway repository path (one empty baseline
# commit), registered for cleanup.
new_repo() {
  local repo
  repo="$(mktemp -d "${TMPDIR:-/tmp}/pmc-lib.XXXXXX")"
  REPOS+=("$repo")
  git -C "$repo" init -q
  git -C "$repo" -c user.name=t -c user.email=t@example.com commit -q --allow-empty -m base
  printf '%s\n' "$repo"
}

# 1. A tree that changed not at all matches its snapshot.
repo="$(new_repo)"
snap="$(snapshot_tree_state "$repo")"
if check_tree_restored "$repo" "$snap"; then
  pass "clean tree matches its snapshot"
else
  fail "clean tree matches its snapshot"
fi

# 2. A stash entry the snapshot does not hold is named and fails.
repo="$(new_repo)"
snap="$(snapshot_tree_state "$repo")"
printf 'dirty\n' > "$repo/file.txt"
git -C "$repo" stash push -q --include-untracked -m kan-448-test-residue
set +e
out="$(check_tree_restored "$repo" "$snap")"
rc=$?
set -e
if [ "$rc" -ne 0 ] && printf '%s' "$out" | grep -q 'new stash entry: .*kan-448-test-residue'; then
  pass "new stash entry is named and fails the check"
else
  fail "new stash entry is named and fails the check (rc=$rc out=$out)"
fi

# 3. A status line the snapshot does not hold is named and fails.
repo="$(new_repo)"
snap="$(snapshot_tree_state "$repo")"
printf 'stray\n' > "$repo/stray.txt"
set +e
out="$(check_tree_restored "$repo" "$snap")"
rc=$?
set -e
if [ "$rc" -ne 0 ] && printf '%s' "$out" | grep -q 'unexpected status line: .*stray.txt'; then
  pass "unexpected status line is named and fails the check"
else
  fail "unexpected status line is named and fails the check (rc=$rc out=$out)"
fi

# 4. A stash the snapshot recorded is expected, not drift.
repo="$(new_repo)"
printf 'kept\n' > "$repo/file.txt"
git -C "$repo" stash push -q --include-untracked -m kan-448-kept
snap="$(snapshot_tree_state "$repo")"
if check_tree_restored "$repo" "$snap"; then
  pass "snapshot-recorded stash still present is not drift"
else
  fail "snapshot-recorded stash still present is not drift"
fi

if [ "$FAILURES" -gt 0 ]; then
  printf 'FAIL: %d case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf '%s: all cases passed\n' "$(basename "$0")"
