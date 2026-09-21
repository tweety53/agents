#!/usr/bin/env bash
# Assertion harness for check-mutation-reproducer-pin.sh.
#
# Builds throwaway fixture trees under TMPDIR and points the guard at each
# with CHECK_MUTATION_REPRODUCER_PIN_ROOT, invoking the REAL
# scripts/check-mutation-reproducer-pin.sh as a subprocess against REAL
# fixture files on disk — never a copy of its logic. Never edits this
# repository's own skills/flow/review-panel.md or scripts/run-reproducer.sh.
#
# KAN-624 (panel F5 of kan-568's deferred self-review): the marker literal
# `# mutation-reproducer` and the 10-line window are stated in both
# skills/flow/review-panel.md and scripts/run-reproducer.sh with no
# mechanical pin keeping them together. The cases:
#
#   pin_ok                        — agreeing fixtures exit 0 with the OK verdict
#   marker_drift_in_prose_fails   — prose marker changed, runner unchanged: exit 1
#   window_drift_in_prose_fails   — prose window changed, runner unchanged: exit 1
#   comment_drift_fails           — runner comment window changed, code unchanged: exit 1
#   missing_declaration_is_exit_2 — the canonical declaration line absent: exit 2
#
# plus two cannot-answer cases: the root env var set but empty, and a
# pinned site file missing entirely.
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
CONVENTION="generic"
if head -n 10 -- "$RESOLVED_PATH" | grep -qx '# mutation-reproducer'; then
  CONVENTION="mutation"
fi
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
  TMP_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/mutation-reproducer-pin.XXXXXX")"
  mkdir -p "$TMP_ROOT/scripts" "$TMP_ROOT/skills/flow"
  write_runner
  write_panel
}

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
CONVENTION="generic"
if head -n 10 -- "$RESOLVED_PATH" | grep -qx '# mutation-reproducer'; then
  CONVENTION="mutation"
fi
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
  cat > "$TMP_ROOT/scripts/run-reproducer.sh" <<'EOF'
#!/usr/bin/env bash
# THE MUTATION-REPRODUCER CONVENTION (KAN-568), declared elsewhere now.
CONVENTION="generic"
case "$(head -n 10 -- "$RESOLVED_PATH")" in
  *'# mutation-reproducer'*) CONVENTION="mutation" ;;
esac
EOF
  run_guard
  [ "$RC" -eq 2 ] && pass "missing_declaration_is_exit_2: exits 2" || fail "missing_declaration_is_exit_2: expected exit 2, got rc=$RC out=$OUT"
  case "$OUT" in
    *"declaration line"*) pass "missing_declaration_is_exit_2: names the declaration line" ;;
    *) fail "missing_declaration_is_exit_2: expected the declaration line named, got: $OUT" ;;
  esac
}

empty_root_env_is_exit_2() {
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

pin_ok
marker_drift_in_prose_fails
window_drift_in_prose_fails
comment_drift_fails
missing_declaration_is_exit_2
empty_root_env_is_exit_2
missing_site_file_is_exit_2

if [ "$FAIL" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAIL" >&2
  exit 1
fi
printf 'all cases passed\n'
