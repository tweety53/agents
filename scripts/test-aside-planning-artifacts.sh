#!/usr/bin/env bash
# Assertion harness for aside-planning-artifacts.sh. Builds throwaway git
# repositories under a sandboxed TMPDIR and asserts the helper's verdict and
# exit status. Never touches the real repository tree. Same shape as
# test-check-base-moved.sh: an indexed REPOS array (bash 3.2 has no
# associative arrays), removed by an EXIT trap, real git repositories rather
# than fixture trees.
#
# READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the stated
# contract in tasks.md's task 1, never against observed output.
#
# The fixture repos carry a `spectre/changes/demo/` tree so
# scripts/lib/spec-root.sh's leaf probe answers `spectre` the way it does in
# a real project, plus `docs/superpowers/` — the helper's two planning
# paths.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HELPER="$SCRIPT_DIR/aside-planning-artifacts.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

REPOS=()
cleanup() {
  [ "${#REPOS[@]}" -eq 0 ] && return 0
  for repo in "${REPOS[@]}"; do
    rm -rf "$repo"
  done
}
trap cleanup EXIT

# run_helper <action> [worktree] — captures stdout and stderr separately, so
# the exit-2 cases can assert NOTHING on stdout. The worktree is passed
# through even when absent (case 7a): the helper's own argument check is
# what must refuse it, never this harness's `set -u`.
run_helper() {
  set +e
  OUT="$(bash "$HELPER" "$1" "${2-}" 2>"$SANDBOX/stderr")"
  RC=$?
  set -e
}

# new_repo — one commit carrying a tracked planning file under the spec
# tree's changes directory, checked out on branch `demo`.
new_repo() {
  REPO="$(mktemp -d "${TMPDIR:-/tmp}/aside-artifacts-test.XXXXXX")"
  REPOS+=("$REPO")
  git -C "$REPO" init -q -b demo
  git -C "$REPO" config user.email test@example.invalid
  git -C "$REPO" config user.name "Test"
  mkdir -p "$REPO/spectre/changes/demo" "$REPO/docs/superpowers"
  echo plan > "$REPO/spectre/changes/demo/tasks.md"
  echo ledger > "$REPO/docs/superpowers/ledger.md"
  echo code > "$REPO/src.txt"
  git -C "$REPO" add spectre docs src.txt
  git -C "$REPO" commit -qm "base"
}

# planning_clean <repo> — zero porcelain lines under the planning paths.
planning_clean() {
  [ -z "$(git -C "$1" status --porcelain -- spectre/changes docs/superpowers)" ]
}

SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/aside-artifacts-sandbox.XXXXXX")"
REPOS+=("$SANDBOX")

# --- Case 1: tracked modification + untracked file under the spec tree's
# changes/ — aside clears both (planning paths porcelain-empty), restore
# brings both back byte-identical.
new_repo
echo plan-edited > "$REPO/spectre/changes/demo/tasks.md"
echo scratch > "$REPO/spectre/changes/demo/notes.md"
run_helper aside "$REPO"
if [ "$RC" -eq 0 ] && printf '%s' "$OUT" | grep -q '^PLANNING-ARTIFACTS-ASIDE:'; then
  if planning_clean "$REPO"; then
    pass "case 1: aside clears the planning paths"
  else
    fail "case 1: planning paths still dirty after aside: $(git -C "$REPO" status --porcelain -- spectre/changes docs/superpowers)"
  fi
else
  fail "case 1: aside did not report ASIDE (rc=$RC out=$OUT)"
fi
run_helper restore "$REPO"
restored_tasks="$(cat "$REPO/spectre/changes/demo/tasks.md")"
restored_notes="$(cat "$REPO/spectre/changes/demo/notes.md" 2>/dev/null || true)"
if [ "$RC" -eq 0 ] && printf '%s' "$OUT" | grep -q '^PLANNING-ARTIFACTS-RESTORED:' \
  && [ "$restored_tasks" = "plan-edited" ] && [ "$restored_notes" = "scratch" ]; then
  pass "case 1: restore brings the tracked edit and the untracked file back"
else
  fail "case 1: restore incomplete (rc=$RC out=$OUT tasks=$restored_tasks notes=$restored_notes)"
fi

# --- Case 2: state under docs/superpowers/ alone is set aside and restored.
new_repo
echo ledger-edited > "$REPO/docs/superpowers/ledger.md"
run_helper aside "$REPO"
if [ "$RC" -eq 0 ] && planning_clean "$REPO" \
  && [ "$(cat "$REPO/docs/superpowers/ledger.md")" = "ledger" ]; then
  pass "case 2: aside clears docs/superpowers"
else
  fail "case 2: aside did not clear docs/superpowers (rc=$RC)"
fi
run_helper restore "$REPO"
if [ "$RC" -eq 0 ] && [ "$(cat "$REPO/docs/superpowers/ledger.md")" = "ledger-edited" ]; then
  pass "case 2: restore brings the docs/superpowers edit back"
else
  fail "case 2: restore incomplete (rc=$RC out=$OUT)"
fi

# --- Case 3: a dirty path outside the planning paths survives aside
# untouched — the helper stashes planning state only, never implementation
# WIP.
new_repo
echo plan-edited > "$REPO/spectre/changes/demo/tasks.md"
echo wip > "$REPO/src.txt"
run_helper aside "$REPO"
# src.txt must still carry its unstaged WIP: the content unchanged on disk
# and the diff still present (diff --quiet exits 1 on a present diff).
if [ "$RC" -eq 0 ] && [ "$(cat "$REPO/src.txt")" = "wip" ] \
  && ! git -C "$REPO" diff --quiet -- src.txt; then
  pass "case 3: implementation WIP outside the planning paths is untouched"
else
  fail "case 3: src.txt was stashed or modified by aside (rc=$RC)"
fi
run_helper restore "$REPO"

# --- Case 4: restore with no aside stash on top — no stash at all, and an
# operator stash shadowing the helper's entry — reports NONE, exits 0, and
# never pops the operator's entry.
new_repo
run_helper restore "$REPO"
if [ "$RC" -eq 0 ] && printf '%s' "$OUT" | grep -q '^PLANNING-ARTIFACTS-NONE:'; then
  pass "case 4a: restore with no stash reports NONE"
else
  fail "case 4a: restore with no stash rc=$RC out=$OUT"
fi
echo operator-wip > "$REPO/src.txt"
git -C "$REPO" stash push -m "operator wip" -- src.txt >/dev/null
echo plan-edited > "$REPO/spectre/changes/demo/tasks.md"
run_helper aside "$REPO"
echo operator-later > "$REPO/src.txt"
git -C "$REPO" stash push -m "operator second" -- src.txt >/dev/null
run_helper restore "$REPO"
# Three entries at this point: the operator's first stash, the helper's
# aside (now shadowed beneath the operator's second), both untouched.
stash_count="$(git -C "$REPO" stash list | wc -l | tr -d ' ')"
if [ "$RC" -eq 0 ] && printf '%s' "$OUT" | grep -q '^PLANNING-ARTIFACTS-NONE:' \
  && [ "$stash_count" -eq 3 ]; then
  pass "case 4b: an operator stash on top is never popped"
else
  fail "case 4b: rc=$RC out=$OUT stashes=$stash_count (expected 3)"
fi

# --- Case 5: a change committed to the same planning file between aside and
# restore (what a rebase that moved the file does) makes restore exit 1 with
# CONFLICT and the stash still listed.
new_repo
echo line-one-aside > "$REPO/spectre/changes/demo/tasks.md"
run_helper aside "$REPO"
echo line-one-upstream > "$REPO/spectre/changes/demo/tasks.md"
git -C "$REPO" add spectre/changes/demo/tasks.md
git -C "$REPO" commit -qm "upstream moves the planning file"
run_helper restore "$REPO"
stash_count="$(git -C "$REPO" stash list | wc -l | tr -d ' ')"
if [ "$RC" -eq 1 ] && printf '%s' "$OUT" | grep -q '^PLANNING-ARTIFACTS-CONFLICT:' \
  && [ "$stash_count" -eq 1 ]; then
  pass "case 5: conflicted restore exits 1 and keeps the stash"
else
  fail "case 5: rc=$RC out=$OUT stashes=$stash_count (expected 1)"
fi

# --- Case 6: restore is refused while a rebase is still in progress —
# exit 2, nothing on stdout.
new_repo
git -C "$REPO" checkout -q -b upstream
echo upstream-line > "$REPO/src.txt"
git -C "$REPO" add src.txt
git -C "$REPO" commit -qm "upstream rewrites src"
git -C "$REPO" checkout -q demo
echo demo-line > "$REPO/src.txt"
git -C "$REPO" add src.txt
git -C "$REPO" commit -qm "demo rewrites src"
echo plan-edited > "$REPO/spectre/changes/demo/tasks.md"
run_helper aside "$REPO"
set +e
git -C "$REPO" rebase upstream >/dev/null 2>&1
set -e
run_helper restore "$REPO"
if [ "$RC" -eq 2 ] && [ -z "$OUT" ]; then
  pass "case 6: restore refused mid-rebase with nothing on stdout"
else
  fail "case 6: rc=$RC out=$OUT (expected 2 and empty)"
fi
set +e
git -C "$REPO" rebase --abort >/dev/null 2>&1
set -e
run_helper restore "$REPO"
if [ "$RC" -eq 0 ] && printf '%s' "$OUT" | grep -q '^PLANNING-ARTIFACTS-RESTORED:'; then
  pass "case 6: restore succeeds once the rebase has aborted"
else
  fail "case 6: post-abort restore rc=$RC out=$OUT"
fi

# --- Case 7: wrong argument count, an unknown subcommand, and a non-repo
# path each exit 2 with nothing on stdout.
run_helper aside
if [ "$RC" -eq 2 ] && [ -z "$OUT" ]; then
  pass "case 7a: missing worktree argument exits 2 silently"
else
  fail "case 7a: rc=$RC out=$OUT"
fi
run_helper frobnicate "$REPO"
if [ "$RC" -eq 2 ] && [ -z "$OUT" ]; then
  pass "case 7b: unknown subcommand exits 2 silently"
else
  fail "case 7b: rc=$RC out=$OUT"
fi
# 7d — a genuine one-argument call, refused by the count guard itself and not
# by the -d directory check 7a's empty string hits: with the count check
# removed, "$2" is unset under the helper's set -u and the helper dies rc=1,
# which this case reads as the guard gone.
set +e
OUT="$(bash "$HELPER" aside 2>"$SANDBOX/stderr")"
RC=$?
set -e
if [ "$RC" -eq 2 ] && [ -z "$OUT" ]; then
  pass "case 7d: a one-argument call exits 2 silently via the count guard"
else
  fail "case 7d: rc=$RC out=$OUT (expected 2 and empty)"
fi
NOT_A_REPO="$(mktemp -d "${TMPDIR:-/tmp}/aside-artifacts-notrepo.XXXXXX")"
REPOS+=("$NOT_A_REPO")
run_helper aside "$NOT_A_REPO"
if [ "$RC" -eq 2 ] && [ -z "$OUT" ]; then
  pass "case 7c: non-repository path exits 2 silently"
else
  fail "case 7c: rc=$RC out=$OUT"
fi

# --- Case 8: clean planning paths — aside reports CLEAN and creates no
# stash.
new_repo
run_helper aside "$REPO"
stash_count="$(git -C "$REPO" stash list | wc -l | tr -d ' ')"
if [ "$RC" -eq 0 ] && printf '%s' "$OUT" | grep -q '^PLANNING-ARTIFACTS-CLEAN:' \
  && [ "$stash_count" -eq 0 ]; then
  pass "case 8: clean planning paths report CLEAN and stash nothing"
else
  fail "case 8: rc=$RC out=$OUT stashes=$stash_count"
fi

# --- Case 9: a stash list long enough that git writes it in several pipe
# chunks — restore reads the top entry and must never read git's SIGPIPE
# (head exits after the first line) as a refusal. The reflog file is
# appended directly: `update-ref` skips no-op writes, `stash store` demands
# stash-like commits, and 250 real stashes would tax the suite's measured
# runtime for no extra coverage. `git stash list` reads this file.
new_repo
HEAD_SHA="$(git -C "$REPO" rev-parse HEAD)"
# one real entry creates the ref and its reflog; stash list reads a ref
# that exists, never a reflog alone
git -C "$REPO" update-ref -m "operator wip 0" --create-reflog refs/stash HEAD
for n in $(seq 1 2000); do
  printf '%s %s Test <test@example.invalid> %s +0000\toperator wip %s\n' \
    "$HEAD_SHA" "$HEAD_SHA" "$((1700000000 + n))" "$n" >> "$REPO/.git/logs/refs/stash"
done
run_helper restore "$REPO"
if [ "$RC" -eq 0 ] && printf '%s' "$OUT" | grep -q '^PLANNING-ARTIFACTS-NONE:'; then
  pass "case 9: a long stash list is read without a SIGPIPE refusal"
else
  fail "case 9: rc=$RC out=$OUT"
fi

if [ "$FAILURES" -eq 0 ]; then
  printf 'aside-planning-artifacts: all cases pass\n'
  exit 0
fi
printf 'aside-planning-artifacts: %d failure(s)\n' "$FAILURES" >&2
exit 1
