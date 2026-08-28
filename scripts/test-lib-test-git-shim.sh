#!/usr/bin/env bash
# Assertion harness for scripts/lib/test-git-shim.sh's assert_shim_fired
# (KAN-88 fix round 4, F17). Sources the library directly, in-process,
# rather than through a whole guard invocation — the same posture
# test-lib-coverage.sh takes toward scripts/lib/coverage.sh: the thing
# under test is the helper's own contract, not any guard's use of it.
#
# WHAT THIS DOES AND DOES NOT ESTABLISH. This pins exactly one contract:
# assert_shim_fired fails when the shim directory it is given has no
# `.fired` sentinel, and passes when it does. It does NOT close the
# meta-level regress the panel report named — pass/fail themselves, and
# assert_shim_fired's use of them, remain untested by anything in this
# repository, the same way coverage_verdict's own plumbing is untested by
# test-lib-coverage.sh. A future assert_shim_fired that always calls `pass`
# regardless of the sentinel would still slip past a reviewer who only reads
# this file's PASS/FAIL summary line, exactly as it slipped past every
# consumer guard before this file existed. Treat this as one rung of
# coverage, not proof the regress is closed.
#
# Bash 3.2 is the floor: indexed arrays only, per scripts/test-check-base-
# moved.sh's header and this repository's other test-lib-*.sh files.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LIB="$SCRIPT_DIR/lib/test-git-shim.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

if [ ! -r "$LIB" ]; then
  echo "test-lib-test-git-shim: cannot read $LIB" >&2
  exit 2
fi
# shellcheck source=lib/test-git-shim.sh
source "$LIB"

DIRS=()
cleanup() {
  [ "${#DIRS[@]}" -eq 0 ] && return 0
  for d in "${DIRS[@]}"; do
    rm -rf "$d"
  done
}
trap cleanup EXIT

# Each case below drives assert_shim_fired in a subshell with its own throwaway
# pass/fail, so the helper's own PASS/FAIL line is captured as plain text
# rather than mutating this file's FAILURES counter directly (assert_shim_fired
# calls whichever pass/fail the sourcing scope defines — see the library's own
# header) — this file's own pass/fail then judges that captured text.

# ---------------------------------------------------------------------------
# 1. An unprepared directory (no .fired sentinel) — assert_shim_fired must
#    itself report a failure, not pass vacuously.
# ---------------------------------------------------------------------------
UNPREPARED="$(mktemp -d "${TMPDIR:-/tmp}/shim-assert-test.XXXXXX")"
DIRS+=("$UNPREPARED")

CAPTURED="$(
  FAILURES=0
  fail() { printf 'FAIL: %s\n' "$1"; FAILURES=$((FAILURES + 1)); }
  pass() { printf 'ok: %s\n' "$1"; }
  assert_shim_fired "$UNPREPARED" "case 1"
  echo "RC=$FAILURES"
)"
if printf '%s\n' "$CAPTURED" | grep -q '^FAIL: case 1'; then
  pass "case 1: an unprepared dir (no .fired) makes assert_shim_fired fail"
else
  fail "case 1: an unprepared dir (no .fired) did not fail assert_shim_fired — got: $CAPTURED"
fi

# ---------------------------------------------------------------------------
# 2. A prepared directory (.fired present, as shim_failing_git writes it) —
#    assert_shim_fired must pass, not reject a real firing.
# ---------------------------------------------------------------------------
PREPARED="$(mktemp -d "${TMPDIR:-/tmp}/shim-assert-test.XXXXXX")"
DIRS+=("$PREPARED")
: > "$PREPARED/.fired"

CAPTURED="$(
  FAILURES=0
  fail() { printf 'FAIL: %s\n' "$1"; FAILURES=$((FAILURES + 1)); }
  pass() { printf 'ok: %s\n' "$1"; }
  assert_shim_fired "$PREPARED" "case 2"
  echo "RC=$FAILURES"
)"
if printf '%s\n' "$CAPTURED" | grep -q '^ok: case 2'; then
  pass "case 2: a prepared dir (.fired present) makes assert_shim_fired pass"
else
  fail "case 2: a prepared dir (.fired present) did not pass assert_shim_fired — got: $CAPTURED"
fi
if printf '%s\n' "$CAPTURED" | grep -q '^RC=0$'; then
  pass "case 2: a prepared dir records zero failures"
else
  fail "case 2: a prepared dir unexpectedly recorded a failure — got: $CAPTURED"
fi

# ---------------------------------------------------------------------------
# 3. Idempotent capture (KAN-88 fix round 4, F18) — sourcing the library a
#    second time while a shim sits earlier on PATH must not overwrite the
#    real-git path the first, clean source resolved. Runs in a subshell so
#    it does not disturb this file's own already-sourced copy of the library,
#    its PATH, or its TEST_GIT_SHIM_REAL_GIT. TEST_GIT_SHIM_REAL_GIT is
#    explicitly unset before the first source in there, simulating a fresh
#    process — without that, the first source would itself be a no-op
#    re-source and the case would pass under an unconditional assignment
#    too, testing nothing.
# ---------------------------------------------------------------------------
FAKE_BIN="$(mktemp -d "${TMPDIR:-/tmp}/shim-fake-git.XXXXXX")"
DIRS+=("$FAKE_BIN")
cat > "$FAKE_BIN/git" <<'EOF'
#!/usr/bin/env bash
echo "fake"
EOF
chmod +x "$FAKE_BIN/git"

CAPTURED="$(
  unset TEST_GIT_SHIM_REAL_GIT
  source "$LIB"
  FIRST="$TEST_GIT_SHIM_REAL_GIT"
  PATH="$FAKE_BIN:$PATH"
  source "$LIB"
  SECOND="$TEST_GIT_SHIM_REAL_GIT"
  echo "FIRST=$FIRST"
  echo "SECOND=$SECOND"
)"
FIRST_VAL="$(printf '%s\n' "$CAPTURED" | sed -n 's/^FIRST=//p')"
SECOND_VAL="$(printf '%s\n' "$CAPTURED" | sed -n 's/^SECOND=//p')"
if [ -n "$FIRST_VAL" ] && [ "$FIRST_VAL" = "$SECOND_VAL" ]; then
  pass "case 3: re-sourcing while a shim leads PATH keeps the first real-git path"
else
  fail "case 3: re-sourcing while a shim leads PATH changed the real-git path — got FIRST=$FIRST_VAL SECOND=$SECOND_VAL"
fi

# ---------------------------------------------------------------------------
if [ "$FAILURES" -eq 0 ]; then
  printf '\n✓ PASS\n'
  exit 0
fi
printf '\n✗ FAIL — %s failure(s)\n' "$FAILURES" >&2
exit 1
