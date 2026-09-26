#!/usr/bin/env bash
# Assertion harness for check-mutation-reproducer-pin.sh.
#
# Builds throwaway fixture trees under TMPDIR and points the guard at each
# with CHECK_MUTATION_REPRODUCER_PIN_ROOT, invoking the REAL
# scripts/check-mutation-reproducer-pin.sh as a subprocess against REAL
# fixture files on disk — never a copy of its logic. Never edits this
# repository's own skills/flow/review-panel.md, scripts/run-reproducer.sh
# or stats/internal/guard/runreproducer.go.
#
# KAN-624 (panel F5 of kan-568's deferred self-review): the marker literal
# `# mutation-reproducer` and the 10-line window are stated in
# skills/flow/review-panel.md and scripts/run-reproducer.sh's header, and
# declared in code by stats/internal/guard/runreproducer.go (KAN-760: the
# runner's Go port), with no mechanical pin keeping them together. The
# cases:
#
#   pin_ok                        — agreeing fixtures exit 0 with the OK verdict
#   marker_drift_in_prose_fails   — prose marker changed, runner unchanged: exit 1
#   window_drift_in_prose_fails   — prose window changed, runner unchanged: exit 1
#   comment_drift_fails           — runner header window changed, code unchanged: exit 1
#   missing_declaration_is_exit_2 — a canonical declaration line absent: exit 2
#
# plus, from the round-0 panel findings:
#
#   multi_site_drift_fails          — F1: one of two runner window statements
#                                     drifted while the other reads 10: exit 1
#   wrapped_window_drift_fails      — F1: a near-miss window wrapped across two
#                                     lines, the brief's own shape: exit 1
#   near_miss_marker_in_prose_fails — F1: a backticked `# mutation-...` span
#                                     that is not the canonical literal: exit 1
#   raw_string_marker_is_exit_2     — F2: the marker constant re-quoted as a Go
#                                     raw string — not extractable: exit 2
#
# and two cannot-answer cases: the root env var set but empty, and a
# pinned site file missing entirely. Every fixture tree is removed on
# replacement and at exit (F3).
set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-mutation-reproducer-pin.sh"

PASS=0
FAIL=0
TMP_ROOT=""
OUT=""
RC=0

pass() { PASS=$((PASS + 1)); printf 'ok - %s\n' "$1"; }
fail() { FAIL=$((FAIL + 1)); printf 'FAIL - %s\n' "$1" >&2; }

write_runner() {
  cat > "$TMP_ROOT/scripts/run-reproducer.sh" <<'EOF'
#!/usr/bin/env bash
# THE MUTATION-REPRODUCER CONVENTION (KAN-568). Such a reproducer declares
# itself with the exact line `# mutation-reproducer` within its first 10 lines,
# and this script then reads it under that convention.
exec flow-guard run-reproducer "$@"
EOF
}

write_source() {
  cat > "$TMP_ROOT/stats/internal/guard/runreproducer.go" <<'EOF'
package guard

const (
	rrMarker     = "# mutation-reproducer"
)

func declaresMutation(path string) bool {
	for _, l := range headLines(path, 10) {
		if l == rrMarker {
			return true
		}
	}
	return false
}
EOF
}

write_panel() {
  cat > "$TMP_ROOT/skills/flow/review-panel.md" <<'EOF'
A surviving mutant's reproducer carries the exact line
`# mutation-reproducer` within its first 10 lines — the declaration
`run-reproducer.sh` reads as the mutation convention.
EOF
}

new_root() {
  [ -n "$TMP_ROOT" ] && rm -rf "$TMP_ROOT"
  TMP_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/mutation-reproducer-pin.XXXXXX")"
  mkdir -p "$TMP_ROOT/scripts" "$TMP_ROOT/skills/flow" "$TMP_ROOT/stats/internal/guard"
  write_source
  write_runner
  write_panel
}

# every fixture tree this harness creates is removed — on replacement by
# the next new_root, and the last one at exit (KAN-624 F3).
trap '[ -n "$TMP_ROOT" ] && rm -rf "$TMP_ROOT"' EXIT

run_guard() {
  OUT="$(CHECK_MUTATION_REPRODUCER_PIN_ROOT="$TMP_ROOT" bash "$GUARD" 2>&1)"
  RC=$?
}

# case <name> — every case below is a named function so the suite reports
# per-case and a reader can run one case by calling it.
pin_ok() {
  new_root
  run_guard
  [ "$RC" -eq 0 ] && pass "pin_ok: exits 0" || fail "pin_ok: expected exit 0, got rc=$RC out=$OUT"
  case "$OUT" in
    *"MUTATION-REPRODUCER-PIN-OK"*) pass "pin_ok: prints the OK verdict" ;;
    *) fail "pin_ok: expected MUTATION-REPRODUCER-PIN-OK in output, got: $OUT" ;;
  esac
}

marker_drift_in_prose_fails() {
  new_root
  cat > "$TMP_ROOT/skills/flow/review-panel.md" <<'EOF'
A surviving mutant's reproducer carries the exact line
`# mutation-check` within its first 10 lines — the declaration
`run-reproducer.sh` reads as the mutation convention.
EOF
  run_guard
  [ "$RC" -eq 1 ] && pass "marker_drift_in_prose_fails: exits 1" || fail "marker_drift_in_prose_fails: expected exit 1, got rc=$RC out=$OUT"
  case "$OUT" in
    *"review-panel.md"*) pass "marker_drift_in_prose_fails: names review-panel.md" ;;
    *) fail "marker_drift_in_prose_fails: expected review-panel.md named, got: $OUT" ;;
  esac
}

window_drift_in_prose_fails() {
  new_root
  cat > "$TMP_ROOT/skills/flow/review-panel.md" <<'EOF'
A surviving mutant's reproducer carries the exact line
`# mutation-reproducer` within its first 12 lines — the declaration
`run-reproducer.sh` reads as the mutation convention.
EOF
  run_guard
  [ "$RC" -eq 1 ] && pass "window_drift_in_prose_fails: exits 1" || fail "window_drift_in_prose_fails: expected exit 1, got rc=$RC out=$OUT"
  case "$OUT" in
    *"no longer states the window"*) pass "window_drift_in_prose_fails: names the drifted window statement" ;;
    *) fail "window_drift_in_prose_fails: expected the window drift named, got: $OUT" ;;
  esac
}

comment_drift_fails() {
  new_root
  cat > "$TMP_ROOT/scripts/run-reproducer.sh" <<'EOF'
#!/usr/bin/env bash
# THE MUTATION-REPRODUCER CONVENTION (KAN-568). Such a reproducer declares
# itself with the exact line `# mutation-reproducer` within its first 12 lines,
# and this script then reads it under that convention.
exec flow-guard run-reproducer "$@"
EOF
  run_guard
  [ "$RC" -eq 1 ] && pass "comment_drift_fails: exits 1" || fail "comment_drift_fails: expected exit 1, got rc=$RC out=$OUT"
  case "$OUT" in
    *"run-reproducer.sh"*) pass "comment_drift_fails: names run-reproducer.sh" ;;
    *) fail "comment_drift_fails: expected run-reproducer.sh named, got: $OUT" ;;
  esac
}

missing_declaration_is_exit_2() {
  new_root
  cat > "$TMP_ROOT/stats/internal/guard/runreproducer.go" <<'EOF'
package guard

// The window is declared elsewhere now.
const rrMarker = "# mutation-reproducer"

func declaresMutation(lines []string) bool {
	for _, l := range lines {
		if l == rrMarker {
			return true
		}
	}
	return false
}
EOF
  run_guard
  [ "$RC" -eq 2 ] && pass "missing_declaration_is_exit_2: exits 2" || fail "missing_declaration_is_exit_2: expected exit 2, got rc=$RC out=$OUT"
  case "$OUT" in
    *"declaration line"*) pass "missing_declaration_is_exit_2: names the declaration line" ;;
    *) fail "missing_declaration_is_exit_2: expected the declaration line named, got: $OUT" ;;
  esac
}

empty_root_env_is_exit_2() {
  [ -n "$TMP_ROOT" ] && rm -rf "$TMP_ROOT"
  TMP_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/mutation-reproducer-pin.XXXXXX")"
  OUT="$(CHECK_MUTATION_REPRODUCER_PIN_ROOT= bash "$GUARD" 2>&1)"
  RC=$?
  [ "$RC" -eq 2 ] && pass "empty_root_env_is_exit_2: exits 2" || fail "empty_root_env_is_exit_2: expected exit 2, got rc=$RC out=$OUT"
}

missing_site_file_is_exit_2() {
  new_root
  rm "$TMP_ROOT/skills/flow/review-panel.md"
  run_guard
  [ "$RC" -eq 2 ] && pass "missing_site_file_is_exit_2: exits 2" || fail "missing_site_file_is_exit_2: expected exit 2, got rc=$RC out=$OUT"
}

# KAN-624 F1: a runner site can state its window more than once, so a drift of one alone must fail even though
# the other still reads correctly — presence-per-file cannot see it, the
# near-miss scan must.
multi_site_drift_fails() {
  new_root
  cat > "$TMP_ROOT/scripts/run-reproducer.sh" <<'EOF'
#!/usr/bin/env bash
# THE MUTATION-REPRODUCER CONVENTION (KAN-568). Such a reproducer declares
# itself with the exact line `# mutation-reproducer` within its first 12 lines,
# and this script then reads it under that convention.
# Exact line, within the first 10 lines: a line that merely contains the
# marker is prose, not a declaration.
exec flow-guard run-reproducer "$@"
EOF
  run_guard
  [ "$RC" -eq 1 ] && pass "multi_site_drift_fails: exits 1" || fail "multi_site_drift_fails: expected exit 1, got rc=$RC out=$OUT"
  case "$OUT" in
    *'states the window as "first 12 lines"'*) pass "multi_site_drift_fails: names the drifted statement at its line" ;;
    *) fail "multi_site_drift_fails: expected the near-miss named, got: $OUT" ;;
  esac
}

# The real mutation-testing brief wraps its window statement across two
# lines; a near-miss value wrapped the same way must still be caught.
wrapped_window_drift_fails() {
  new_root
  cat > "$TMP_ROOT/skills/flow/review-panel.md" <<'EOF'
A surviving mutant's reproducer carries the exact line
`# mutation-reproducer` within its first 10 lines — the declaration
`run-reproducer.sh` reads as the mutation convention.

One `# demonstrates:` line per cited location, within the script's first 12
lines — the same window the `# mutation-reproducer` declaration reads.
EOF
  run_guard
  [ "$RC" -eq 1 ] && pass "wrapped_window_drift_fails: exits 1" || fail "wrapped_window_drift_fails: expected exit 1, got rc=$RC out=$OUT"
  case "$OUT" in
    *'states the window as "first 12 lines"'*) pass "wrapped_window_drift_fails: names the wrapped near-miss" ;;
    *) fail "wrapped_window_drift_fails: expected the wrapped near-miss named, got: $OUT" ;;
  esac
}

# A backticked mutation-marker span that is not the canonical literal is a
# renamed site drifting alone — caught even while correct spans remain.
near_miss_marker_in_prose_fails() {
  new_root
  cat > "$TMP_ROOT/skills/flow/review-panel.md" <<'EOF'
A surviving mutant's reproducer carries the exact line
`# mutation-reproducer` within its first 10 lines — the declaration
`run-reproducer.sh` reads as the mutation convention.

A reproducer declares it with the exact `# mutation-check` line.
EOF
  run_guard
  [ "$RC" -eq 1 ] && pass "near_miss_marker_in_prose_fails: exits 1" || fail "near_miss_marker_in_prose_fails: expected exit 1, got rc=$RC out=$OUT"
  case "$OUT" in
    *"`# mutation-check`"*) pass "near_miss_marker_in_prose_fails: names the near-miss span" ;;
    *) fail "near_miss_marker_in_prose_fails: expected the near-miss span named, got: $OUT" ;;
  esac
}

# KAN-624 F2: a marker constant that is not an interpreted string literal
# (a re-quote refactor to a Go raw string) is not extractable — exit 2,
# never a bogus exit-1 drift row pasting the whole line as the marker.
raw_string_marker_is_exit_2() {
  new_root
  cat > "$TMP_ROOT/stats/internal/guard/runreproducer.go" <<'EOF'
package guard

const rrMarker = `# mutation-reproducer`

func declaresMutation(path string) bool {
	for _, l := range headLines(path, 10) {
		if l == rrMarker {
			return true
		}
	}
	return false
}
EOF
  run_guard
  [ "$RC" -eq 2 ] && pass "raw_string_marker_is_exit_2: exits 2" || fail "raw_string_marker_is_exit_2: expected exit 2, got rc=$RC out=$OUT"
  case "$OUT" in
    *"could not be extracted"*) pass "raw_string_marker_is_exit_2: names the extraction failure" ;;
    *) fail "raw_string_marker_is_exit_2: expected the extraction failure named, got: $OUT" ;;
  esac
}

pin_ok
marker_drift_in_prose_fails
window_drift_in_prose_fails
comment_drift_fails
missing_declaration_is_exit_2
empty_root_env_is_exit_2
missing_site_file_is_exit_2
multi_site_drift_fails
wrapped_window_drift_fails
near_miss_marker_in_prose_fails
raw_string_marker_is_exit_2

if [ "$FAIL" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAIL" >&2
  exit 1
fi
printf 'all cases passed\n'
