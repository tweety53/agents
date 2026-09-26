#!/usr/bin/env bash
# Assertion harness for guard-autosquash.sh. Builds throwaway git
# repositories under a sandboxed TMPDIR and asserts the guard's verdict and
# exit status. Never touches the real repository tree. Same shape as
# test-aside-planning-artifacts.sh: an indexed REPOS array (bash 3.2 has no
# associative arrays), removed by an EXIT trap, real git repositories rather
# than fixture trees.
#
# READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the stated
# contract in tasks.md's task 1, never against observed output.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HELPER="$SCRIPT_DIR/guard-autosquash.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

REPOS=()
cleanup() {
  [ -n "${SANDBOX:-}" ] && rm -rf "$SANDBOX"
  [ "${#REPOS[@]}" -eq 0 ] && return 0
  for repo in "${REPOS[@]}"; do
    rm -rf "$repo"
  done
}
trap cleanup EXIT

SANDBOX="$(mktemp -d)"

# run_guard <args...> — captures stdout and stderr separately, so the
# exit-code cases can assert on each stream alone.
run_guard() {
  set +e
  OUT="$(bash "$HELPER" "$@" 2>"$SANDBOX/stderr")"
  RC=$?
  set -e
  ERR="$(cat "$SANDBOX/stderr")"
}

# new_repo — one repository with two commits on its default branch, echoing
# the repo path. The root commit plays the fixup target's ancestor; HEAD the
# branch tip.
new_repo() {
  local repo
  repo="$(mktemp -d "$SANDBOX/repo.XXXXXX")"
  git -C "$repo" init -q
  git -C "$repo" config user.email guard@example.com
  git -C "$repo" config user.name 'guard test'
  printf 'one\n' >"$repo/one.txt"
  git -C "$repo" add one.txt
  git -C "$repo" commit -qm one
  printf 'two\n' >"$repo/two.txt"
  git -C "$repo" add two.txt
  git -C "$repo" commit -qm two
  REPOS+=("$repo")
  printf '%s' "$repo"
}

root_sha() { git -C "$1" rev-parse HEAD~1; }
head_sha() { git -C "$1" rev-parse HEAD; }

# orphan_commit <repo> — a commit on an orphan branch, so its sha is a real
# commit object that is nevertheless no ancestor of the default branch. This
# is the incident's shape: resolvable, real, and not on the branch. The
# branch is switched back by name: a fresh repository's HEAD has no checkout
# reflog, so `git checkout -` cannot resolve.
orphan_commit() {
  local back
  back="$(git -C "$1" rev-parse --abbrev-ref HEAD)"
  git -C "$1" checkout -q --orphan stray
  git -C "$1" rm -qrf .
  printf 'stray\n' >"$1/stray.txt"
  git -C "$1" add stray.txt
  git -C "$1" commit -qm stray
  git -C "$1" rev-parse HEAD
  git -C "$1" checkout -q "$back"
}

# targets-accepts-a-genuine-ancestor
repo="$(new_repo)"
run_guard targets "$repo" "$(root_sha "$repo")" "$(head_sha "$repo")"
if [ "$RC" -eq 0 ]; then pass "targets accepts a genuine ancestor"; else fail "targets rejects a genuine ancestor: rc=$RC err=$ERR"; fi

# targets-refuses-a-non-ancestor-sha
repo="$(new_repo)"
stray="$(orphan_commit "$repo")"
run_guard targets "$repo" "$stray"
if [ "$RC" -eq 1 ] && printf '%s' "$ERR" | grep -q "$stray"; then
  pass "targets refuses a non-ancestor sha and names it"
else
  fail "targets on a non-ancestor sha: rc=$RC err=$ERR"
fi
# targets-reports-a-non-resolving-sha-as-unresolved
run_guard targets "$repo" deadbeefcafe1234
if [ "$RC" -eq 1 ] && printf '%s' "$ERR" | grep -q 'does not resolve as a commit object'; then
  pass "targets reports a non-resolving sha as unresolved"
else
  fail "targets on a bogus sha: rc=$RC err=$ERR"
fi

# after-accepts-a-clean-branch
repo="$(new_repo)"
printf -- '# plan\n\n- task 1 commit %s\n<!-- baseline %s -->\n' "$(head_sha "$repo")" "$(root_sha "$repo")" >"$SANDBOX/clean-tasks.md"
run_guard after "$repo" "$(root_sha "$repo")" "$SANDBOX/clean-tasks.md"
if [ "$RC" -eq 0 ]; then pass "after accepts a clean branch"; else fail "after on a clean branch: rc=$RC err=$ERR"; fi

# after-refuses-a-moved-off-base
repo="$(new_repo)"
stray="$(orphan_commit "$repo")"
: >"$SANDBOX/empty-tasks.md"
run_guard after "$repo" "$stray" "$SANDBOX/empty-tasks.md"
if [ "$RC" -eq 1 ] && printf '%s' "$ERR" | grep -q "$stray"; then
  pass "after refuses a base that is no ancestor of HEAD"
else
  fail "after on a moved-off base: rc=$RC err=$ERR"
fi

# after-refuses-a-stale-tasks-md-sha — the amended sha still exists as an
# object (cat-file -e alone would pass it) but the branch no longer reaches
# it, which is the two-check design's reason.
repo="$(new_repo)"
stale="$(head_sha "$repo")"
git -C "$repo" commit -q --amend -m two-amended
printf -- '- task 5 baseline %s\n' "$stale" >"$SANDBOX/stale-tasks.md"
run_guard after "$repo" "$(root_sha "$repo")" "$SANDBOX/stale-tasks.md"
if [ "$RC" -eq 1 ] && printf '%s' "$ERR" | grep -q "$stale"; then
  pass "after refuses a stale tasks.md sha and names it"
else
  fail "after on a stale tasks.md sha: rc=$RC err=$ERR"
fi
printf 'recorded off the defaced token\n' >"$SANDBOX/prose-tasks.md"
run_guard after "$repo" "$(root_sha "$repo")" "$SANDBOX/prose-tasks.md"
if [ "$RC" -eq 1 ] && printf '%s' "$ERR" | grep -q 'defaced'; then
  pass "after reports a hex-shaped token that resolves to no commit"
else
  fail "after on a prose hex token: rc=$RC err=$ERR"
fi

# after-reports-a-long-hex-run — a sha glued into a longer hex run must be
# reported, never silently skipped for exceeding the sha-length window.
repo="$(new_repo)"
LONGHEX="a12345678901234567890123456789012345678901"
printf -- '- baseline %s\n' "$LONGHEX" >"$SANDBOX/long-tasks.md"
run_guard after "$repo" "$(root_sha "$repo")" "$SANDBOX/long-tasks.md"
if [ "$RC" -eq 1 ] && printf '%s' "$ERR" | grep -q "$LONGHEX"; then
  pass "after reports a long hex run"
else
  fail "after on a long hex run: rc=$RC err=$ERR"
fi

# after-exit-2-on-a-tasks-md-directory — a directory passes a readability
# test but not a regular-file test; the sweep must refuse, never fail open.
repo="$(new_repo)"
run_guard after "$repo" "$(root_sha "$repo")" "$SANDBOX"
if [ "$RC" -eq 2 ] && [ -z "$OUT" ]; then pass "after exits 2 on a tasks-md directory"; else fail "after on a tasks-md directory: rc=$RC out=$OUT"; fi

# after-exit-2-on-a-missing-tasks-md — cannot-answer exits, and stdout says
# nothing a caller could mistake for a verdict.
repo="$(new_repo)"
run_guard after "$repo" "$(root_sha "$repo")" "$SANDBOX/absent-tasks.md"
if [ "$RC" -eq 2 ] && [ -z "$OUT" ]; then pass "after exits 2 on a missing tasks.md"; else fail "after on a missing tasks.md: rc=$RC out=$OUT"; fi
run_guard targets "$SANDBOX" "$(root_sha "$repo")"
if [ "$RC" -eq 2 ]; then pass "targets exits 2 outside a git repository"; else fail "targets outside a repo: rc=$RC"; fi
run_guard
if [ "$RC" -eq 2 ]; then pass "no subcommand exits 2"; else fail "no subcommand: rc=$RC"; fi

if [ "$FAILURES" -gt 0 ]; then
  printf '%d failure(s)\n' "$FAILURES" >&2
  exit 1
fi
printf 'guard-autosquash: all cases passed\n'
