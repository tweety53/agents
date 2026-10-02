#!/usr/bin/env bash
# Assertion harness for check-hand-notes-in-step.sh. Drives the shim against a
# sandboxed home through CHECK_HAND_NOTES_HOME; never reads the real one. The
# comparison logic is the Go port's and is pinned in
# stats/internal/guard/check_hand_notes_test.go — this harness asserts the shim
# reaches it and relays its four verdicts and exit codes.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT="$SCRIPT_DIR/check-hand-notes-in-step.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/check-hand-notes-in-step-test.XXXXXX")"
trap 'rm -rf "$SANDBOX"' EXIT
H="$SANDBOX/home"
mkdir -p "$H/.claude" "$H/.zcode"

# harness <file> <managed-block-body> <hand-section>
harness() {
  printf '%s\n<!-- flow:begin -->\n%s\n<!-- flow:end -->\n' "$3" "$2" > "$1"
}

# run <label> <expected-rc> <stdout-pattern>
run() {
  local out rc
  set +e
  out="$(CHECK_HAND_NOTES_HOME="$H" "$SCRIPT" 2>/dev/null)"
  rc=$?
  set -e
  if [ "$rc" -ne "$2" ]; then
    fail "$1: rc=$rc want $2 out=$out"
    return
  fi
  case "$out" in
    *"$3"*) pass "$1" ;;
    *) fail "$1: stdout lacks '$3': $out" ;;
  esac
}

run "no harness file is nothing to check" 0 "HAND-NOTES-NONE: $H"
harness "$H/.claude/CLAUDE.md" "claude pointers" "# notes"
run "one harness file has no pair" 0 "HAND-NOTES-SINGLE: $H"
harness "$H/.zcode/AGENTS.md" "zcode pointers" "# notes"
run "identical hand sections around differing blocks are in step" 0 "HAND-NOTES-OK: $H"
harness "$H/.zcode/AGENTS.md" "zcode pointers" "# notes, edited on one side"
run "a one-sided hand edit is drift" 1 "HAND-NOTES-DRIFT: $H"

if [ "$FAILURES" -gt 0 ]; then
  printf '%d failure(s)\n' "$FAILURES" >&2
  exit 1
fi
printf 'all check-hand-notes-in-step.sh cases passed\n'
