#!/usr/bin/env bash
# Assertion harness for check-fast-route-record.sh. Builds a fixture git
# repository under a sandboxed mktemp directory and asserts the shim's exit
# status and, where the case names one, the presence of the expected finding
# text. Never touches the real repository tree.
#
# The Go table tests (stats/internal/guard/fastrouterecord_test.go) own the
# walk's logic; this harness owns the shim's end-to-end plumbing — the
# flow_guard_exec build-from-checkout path, the usage refusal, the
# cannot-answer legs and one real finding run through the real binary.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-fast-route-record.sh"
FAILURES=0

FIXTURES=()
cleanup_fixtures() { [ "${#FIXTURES[@]}" -eq 0 ] || rm -rf "${FIXTURES[@]}"; }
trap cleanup_fixtures EXIT

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# fixture_repo <change-branch> [subject] [body] -> prints the fixture path;
# one base commit on main (mirrored to refs/remotes/origin/main the way a
# fetch would), then one commit on the change branch carrying subject/body.
fixture_repo() {
  local branch="$1" subject="${2:-}" body="${3:-}"
  local dir; dir="$(mktemp -d)"
  FIXTURES+=("$dir")
  git -C "$dir" init -q -b main
  git -C "$dir" -c user.email=t@t -c user.name=t commit --allow-empty -q -m "init: base"
  git -C "$dir" update-ref refs/remotes/origin/main HEAD
  git -C "$dir" checkout -q -b "$branch"
  if [ -n "$subject" ]; then
    if [ -n "$body" ]; then
      git -C "$dir" -c user.email=t@t -c user.name=t commit --allow-empty -q -m "$subject" -m "$body"
    else
      git -C "$dir" -c user.email=t@t -c user.name=t commit --allow-empty -q -m "$subject"
    fi
  fi
  printf '%s' "$dir"
}

run_guard() {
  set +e
  OUT="$("$GUARD" "$@" 2>&1)"
  RC=$?
  set -e
}

# 1. A clean conventional series exits 0 with no output.
d="$(fixture_repo kan-838-demo "feat(guard): fine")"
run_guard "$d" main
[ "$RC" -eq 0 ] && [ -z "$OUT" ] && pass "clean series exits 0 silently" \
  || fail "clean series: rc=$RC out=$OUT"

# 2. A bad commit is named, with its attribution finding, exit 1.
d="$(fixture_repo kan-838-demo "bad subject" "Co-Authored-By: Claude <n@example.com>")"
run_guard "$d" main
{ [ "$RC" -eq 1 ] && printf '%s' "$OUT" | grep -q "subject not in Conventional Commits form: bad subject" \
  && printf '%s' "$OUT" | grep -q "attribution trailer: Co-Authored-By:"; } \
  && pass "bad commit named with its finding" || fail "bad commit: rc=$RC out=$OUT"

# 3. No arguments is the usage refusal, exit 2.
run_guard
[ "$RC" -eq 2 ] && printf '%s' "$OUT" | grep -q "usage: check-fast-route-record.sh" \
  && pass "no args exits 2 with usage" || fail "no args: rc=$RC out=$OUT"

# 4. An unresolvable base is cannot-answer, exit 2 with the cause on stderr.
d="$(fixture_repo kan-838-demo "feat(guard): fine")"
run_guard "$d" nosuch
[ "$RC" -eq 2 ] && printf '%s' "$OUT" | grep -q "resolves neither as origin/nosuch" \
  && pass "unresolvable base exits 2 with its cause" || fail "bad base: rc=$RC out=$OUT"

# 5. A directory that is not a worktree is cannot-answer, exit 2 with the cause.
d="$(mktemp -d)"; FIXTURES+=("$d")
run_guard "$d" main
[ "$RC" -eq 2 ] && printf '%s' "$OUT" | grep -q "is not a git worktree" \
  && pass "non-worktree exits 2 with its cause" || fail "non-worktree: rc=$RC out=$OUT"

if [ "$FAILURES" -eq 0 ]; then
  printf 'test-check-fast-route-record: all cases pass\n'
else
  printf 'test-check-fast-route-record: %d failure(s)\n' "$FAILURES" >&2
  exit 1
fi
