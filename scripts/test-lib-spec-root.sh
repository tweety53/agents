#!/usr/bin/env bash
# test-lib-spec-root.sh — assertion harness for scripts/lib/spec-root.sh,
# per this repository's own convention (see test-lib-change-plan.sh's
# header): the thing under test is the library's own contract, sourced
# directly, not any caller's use of it.
#
# Every fixture is a real directory tree built with mktemp -d, and every
# case runs spec_root_leaf against it for real, capturing stdout (the
# selection) and stderr (the dual-tree warning) separately.
#
# The dedup cases each point TMPDIR at a private directory — the shared
# temporary namespace is the exact failure mode .flow/project.md's ## test
# section records for concurrent harnesses — and the marker that dedup
# writes lives under that private TMPDIR, so cases never suppress one
# another's warnings.
#
# Six cases: spectre-only selects spectre silently; openspec-only selects
# openspec silently; neither selects spectre silently; both warns once and
# selects spectre; a second call on the same dir through the same TMPDIR is
# silent; a distinct dir through the same TMPDIR still warns.
#
# Bash 3.2 is the floor, as test-check-finish-preflight.sh's header records:
# indexed arrays only, no associative arrays.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LIB="$SCRIPT_DIR/lib/spec-root.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

if [ ! -r "$LIB" ]; then
  echo "test-lib-spec-root: cannot read $LIB" >&2
  exit 2
fi

# run_leaf <tmpdir> <dir> — call spec_root_leaf with TMPDIR pinned to
# <tmpdir>, in a subshell so the sourced library cannot leak state; prints
# the selection on stdout, the warning on stderr, both captured.
run_leaf() {
  ( export TMPDIR="$1"
    # shellcheck source=lib/spec-root.sh
    source "$LIB"
    spec_root_leaf "$2"
  )
}

# mk_project <spectre:0|1> <openspec:0|1> — print a fresh project dir.
mk_project() {
  local d
  d="$(mktemp -d)"
  [ "$1" -eq 1 ] && mkdir -p "$d/spectre/changes"
  [ "$2" -eq 1 ] && mkdir -p "$d/openspec/changes"
  printf '%s' "$d"
}

TMP_A="$(mktemp -d)"
TMP_B="$(mktemp -d)"

# 1. spectre-only: selects spectre, no warning.
D="$(mk_project 1 0)"
OUT="$(run_leaf "$TMP_A" "$D" 2>"$TMP_A/err")"
[ "$OUT" = "spectre" ] && pass "spectre-only selects spectre" || fail "spectre-only — expected [spectre], got [$OUT]"
[ ! -s "$TMP_A/err" ] && pass "spectre-only is silent" || fail "spectre-only — unexpected stderr: $(cat "$TMP_A/err")"

# 2. openspec-only: selects openspec, no warning.
D="$(mk_project 0 1)"
OUT="$(run_leaf "$TMP_A" "$D" 2>"$TMP_A/err")"
[ "$OUT" = "openspec" ] && pass "openspec-only selects openspec" || fail "openspec-only — expected [openspec], got [$OUT]"
[ ! -s "$TMP_A/err" ] && pass "openspec-only is silent" || fail "openspec-only — unexpected stderr: $(cat "$TMP_A/err")"

# 3. neither: selects spectre (the locate-not-a-preference default), no warning.
D="$(mk_project 0 0)"
OUT="$(run_leaf "$TMP_A" "$D" 2>"$TMP_A/err")"
[ "$OUT" = "spectre" ] && pass "neither selects spectre" || fail "neither — expected [spectre], got [$OUT]"
[ ! -s "$TMP_A/err" ] && pass "neither is silent" || fail "neither — unexpected stderr: $(cat "$TMP_A/err")"

# 4. both, fresh TMPDIR: warns once, selects spectre.
D="$(mk_project 1 1)"
OUT="$(run_leaf "$TMP_B" "$D" 2>"$TMP_B/err")"
[ "$OUT" = "spectre" ] && pass "both selects spectre" || fail "both — expected [spectre], got [$OUT]"
grep -q "carries both" "$TMP_B/err" && pass "both warns" || fail "both — expected the dual-tree warning, got stderr: $(cat "$TMP_B/err")"

# 5. both, same dir, same TMPDIR: the second call is silent (the dedup).
OUT="$(run_leaf "$TMP_B" "$D" 2>"$TMP_B/err")"
[ "$OUT" = "spectre" ] && pass "second call still selects spectre" || fail "second call — expected [spectre], got [$OUT]"
[ ! -s "$TMP_B/err" ] && pass "second call is silent" || fail "second call — expected silence, got stderr: $(cat "$TMP_B/err")"

# 6. both, a distinct dir through the same TMPDIR: still warns.
D2="$(mk_project 1 1)"
OUT="$(run_leaf "$TMP_B" "$D2" 2>"$TMP_B/err")"
grep -q "carries both" "$TMP_B/err" && pass "a distinct dir still warns" || fail "distinct dir — expected the dual-tree warning, got stderr: $(cat "$TMP_B/err")"
[ "$OUT" = "spectre" ] || fail "distinct dir — expected [spectre], got [$OUT]"

if [ "$FAILURES" -eq 0 ]; then
  printf 'ok: test-lib-spec-root — 6 cases, 0 failures\n'
  exit 0
fi
printf 'FAIL: test-lib-spec-root — %d failure(s)\n' "$FAILURES" >&2
exit 1
