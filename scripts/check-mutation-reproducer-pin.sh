#!/usr/bin/env bash
# check-mutation-reproducer-pin.sh — pin the mutation-reproducer marker
# literal and its line window mechanically across the two files that state
# them (KAN-624, panel F5 of kan-568's deferred self-review).
#
# Why this exists: the exact literal `# mutation-reproducer` and the
# 10-line window are each stated in skills/flow/review-panel.md (the
# reproducer-convention paragraph and the mutation-testing brief) and
# scripts/run-reproducer.sh (the header comment) with no mechanical pin
# keeping them together — check-dispatch-paragraphs.sh covers the required
# dispatch paragraphs but not this pair, so the two statements could drift
# silently, one prose edit or one code edit at a time. This guard is the
# pin.
#
# The source of truth is the runner's own canonical declaration line — the
# `head -n <N> -- "$RESOLVED_PATH" | grep -qx '<MARKER>'` conditional that
# decides the convention at run time. The guard extracts the marker and
# the window from that one line and then requires the stated sites to
# agree, two ways:
#
#   - PRESENCE — skills/flow/review-panel.md carries the marker backticked
#     and the window stated as "first <N> lines"; scripts/run-reproducer.sh
#     carries the marker backticked (its header comment) and the same
#     window statement.
#   - NEAR-MISS ABSENCE — presence alone cannot see a single site drift
#     while another copy of the same fact still reads correctly, which is
#     the one-edit-at-a-time drift this guard exists to catch (KAN-624
#     F1). So every "first <M> lines" statement and every backticked
#     `# mutation-...` span in either file must carry exactly the
#     canonical value, on a single line or wrapped across two adjacent
#     lines (the mutation-testing brief's own window statement is
#     wrapped), and each offending occurrence is reported at its own
#     line.
#
# A change to any one site without the others is a drift: exit 1, named
# per file and line, until all sites move together. Reshaping the
# canonical line itself — so that no line of that shape can be found, or
# more than one can, or its marker or window is not extractable from it —
# is not a drift this guard interprets: it cannot tell what the new
# reading would be, so it exits 2 and the pin waits to be re-pointed
# deliberately, the same not-a-verdict course check-plan-shape.sh's exit 2
# takes.
#
# Argument-free and self-scoped exactly like check-dispatch-paragraphs.sh:
# the scan root is this repository's own root, resolved from this script's
# own location. CHECK_MUTATION_REPRODUCER_PIN_ROOT is an explicit, opt-in
# override honored only when set (mirrors CHECK_DISPATCH_PARAGRAPHS_ROOT),
# so the companion harness (test-check-mutation-reproducer-pin.sh) can
# point this guard at a sandboxed fixture tree under TMPDIR without
# touching this repository — never set it for a normal invocation.
#
# Exit codes:
#   0  clean — the marker and the window agree across both files
#   1  drift — a stated site disagrees with the canonical declaration
#   2  cannot answer — a pinned file missing, a symlink, not a regular
#      file or unreadable; the canonical declaration line absent or
#      ambiguous; the marker or window not extractable; or
#      CHECK_MUTATION_REPRODUCER_PIN_ROOT set but empty
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ -n "${CHECK_MUTATION_REPRODUCER_PIN_ROOT:-}" ]; then
  ROOT="$CHECK_MUTATION_REPRODUCER_PIN_ROOT"
elif [ "${CHECK_MUTATION_REPRODUCER_PIN_ROOT+set}" = "set" ]; then
  echo "check-mutation-reproducer-pin: CHECK_MUTATION_REPRODUCER_PIN_ROOT is set but empty" >&2
  exit 2
else
  ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
fi

# report_line <path> <line> <message> -> prints one "path:line: message" row.
report_line() {
  printf '%s:%s: %s\n' "$1" "$2" "$3"
}

# die2 <path> <message> -> a hard "cannot answer at all" refusal, printed
# in the same file:line shape as every other finding this guard reports,
# then exits 2 immediately.
die2() {
  report_line "$1" 0 "$2" >&2
  exit 2
}

[ -d "$ROOT" ] || die2 "$ROOT" "is not a directory — cannot scan"

RUNNER="$ROOT/scripts/run-reproducer.sh"
PANEL="$ROOT/skills/flow/review-panel.md"

for pinned in "$RUNNER" "$PANEL"; do
  if [ -L "$pinned" ]; then
    die2 "$pinned" "is a symlink — a pinned site must be a real file, never a symlink"
  fi
  if [ ! -e "$pinned" ]; then
    die2 "$pinned" "does not exist — this is a pinned mutation-reproducer site"
  fi
  if [ ! -f "$pinned" ]; then
    die2 "$pinned" "exists but is not a regular file — cannot scan"
  fi
  if [ ! -r "$pinned" ]; then
    die2 "$pinned" "is not readable — cannot scan"
  fi
done

# The canonical declaration line: exactly one is answerable. Zero means the
# runner stopped declaring its convention in this shape (moved, renamed or
# rewritten); more than one means the reading is ambiguous. Either way this
# guard refuses to guess.
set +e
DECL_LINES="$(grep -nE 'head -n [0-9]+ -- "\$RESOLVED_PATH" \| grep -qx .' "$RUNNER")"
grep_rc=$?
set -e
if [ "$grep_rc" -ge 2 ]; then
  die2 "$RUNNER" "grep exited $grep_rc while scanning for the canonical declaration line — a failure to look, not an absence"
fi
decl_count="$(printf '%s' "$DECL_LINES" | (grep -c . || true))"
if [ "$grep_rc" -eq 1 ] || [ "$decl_count" -eq 0 ]; then
  die2 "$RUNNER" "the canonical declaration line (head -n <N> -- \"\$RESOLVED_PATH\" | grep -qx '<marker>') is absent — the pin cannot answer until it is re-pointed at the declaration's new shape"
fi
if [ "$decl_count" -gt 1 ]; then
  die2 "$RUNNER" "carries $decl_count lines matching the canonical declaration shape — the reading is ambiguous; keep exactly one"
fi

decl_line="$(printf '%s\n' "$DECL_LINES" | head -n 1 | cut -d: -f2-)"
# Extraction is grep -oE, never a bare sed substitution: sed's no-match
# answer is the input line itself, so an unextractable marker or window
# (the canonical line re-quoted, or `head` replaced) would sail past an
# empty-check and be misreported as drift with the whole line pasted in
# as the value (KAN-624 F2). No match here is the pin's re-point signal:
# exit 2, never a guess.
set +e
WINDOW_TOK="$(printf '%s' "$decl_line" | grep -oE 'head -n [0-9]+ --')"
MARKER_TOK="$(printf '%s' "$decl_line" | grep -oE "grep -qx '[^']*'")"
set -e
if [ -z "$WINDOW_TOK" ] || [ -z "$MARKER_TOK" ]; then
  die2 "$RUNNER" "the canonical declaration line matched but its marker or window could not be extracted from it — the line's shape moved; re-point the pin: $decl_line"
fi
WINDOW="$(printf '%s' "$WINDOW_TOK" | sed -E 's/^head -n //; s/ --$//')"
MARKER="$(printf '%s' "$MARKER_TOK" | sed -E "s/^grep -qx '//; s/'\$//")"

VIOLATIONS=()

# anchor_line <file> <ERE> — the first line carrying any statement of the
# needle's own fact, so a drift row names an editable line even when the
# needle itself has already drifted away (KAN-624 F4); 0 when the file
# carries no such statement at all — the total-absence case, where no
# line exists to name.
anchor_line() {
  local file="$1" ere="$2" hit
  set +e
  hit="$(grep -nE -- "$ere" "$file" 2>/dev/null | head -n 1 | cut -d: -f1)"
  set -e
  printf '%s' "${hit:-0}"
}

# check_contains <file> <needle> <message> <anchor-ere> — a pinned
# statement must be present verbatim in <file>; its absence is the drift
# this guard exists to catch, reported at the anchor line. A grep failure
# other than "no match" is a failure to look, not an absence.
check_contains() {
  local file="$1" needle="$2" message="$3" ere="$4" rc line
  set +e
  grep -qF -- "$needle" "$file" >/dev/null 2>&1
  rc=$?
  set -e
  if [ "$rc" -ge 2 ]; then
    die2 "$file" "grep exited $rc while scanning for the pinned statement \"$needle\" — a failure to look, not an absence"
  fi
  if [ "$rc" -eq 1 ]; then
    line="$(anchor_line "$file" "$ere")"
    VIOLATIONS+=("$(report_line "$file" "$line" "$message")")
  fi
}

PANEL_MARKER_MSG="no longer states the marker literal backticked — it drifted from scripts/run-reproducer.sh's canonical declaration (marker: '$MARKER'); move both sites together"
PANEL_WINDOW_MSG="no longer states the window as \"first $WINDOW lines\" — it drifted from scripts/run-reproducer.sh's canonical declaration (window: $WINDOW); move both sites together"
RUNNER_MARKER_MSG="its own comment no longer states the marker literal backticked — it drifted from the runner's canonical declaration line (marker: '$MARKER'); move both sites together"
RUNNER_WINDOW_MSG="its own comment no longer states the window as \"first $WINDOW lines\" — it drifted from the runner's canonical declaration line (window: $WINDOW); move both sites together"

check_contains "$PANEL" "\`$MARKER\`" "$PANEL_MARKER_MSG"   '# mutation-[a-z0-9][a-z0-9-]*'
check_contains "$PANEL" "first $WINDOW lines" "$PANEL_WINDOW_MSG"  'first [0-9]+ lines'
check_contains "$RUNNER" "\`$MARKER\`" "$RUNNER_MARKER_MSG" '# mutation-[a-z0-9][a-z0-9-]*'
check_contains "$RUNNER" "first $WINDOW lines" "$RUNNER_WINDOW_MSG" 'first [0-9]+ lines'

# Near-miss absence — the presence checks above cannot see a single site
# drifting while another copy of the same fact still reads correctly, which
# is the one-edit-at-a-time drift this guard exists to catch (KAN-624 F1:
# the real runner states its window twice, the panel carries the marker
# three times, and the brief's window statement is wrapped across two
# lines where a single-line needle never looks). So every numeric window
# statement and every backticked mutation-marker span in either pinned
# file must carry exactly the canonical value — on a single line or
# wrapped across two adjacent lines — and each offending occurrence is
# reported at its own line (KAN-624 F4: a drift row names an editable
# line, never :0).
scan_near_miss() {
  local file="$1"
  awk -v win="$WINDOW" -v marker="$MARKER" '
    function scan(s, ln,    n, rest, span) {
      rest = s
      while (match(rest, /first [0-9]+ lines/)) {
        n = substr(rest, RSTART + 6, RLENGTH - 12) + 0
        if (n != win + 0)
          printf "%s:%d: states the window as \"first %d lines\" — a near-miss of the canonical window %s; every window statement must read \"first %s lines\"\n", FILENAME, ln, n, win, win
        rest = substr(rest, RSTART + RLENGTH)
      }
      rest = s
      while (match(rest, /`# mutation-[a-z0-9][a-z0-9-]*`/)) {
        span = substr(rest, RSTART, RLENGTH)
        if (span != "`" marker "`")
          printf "%s:%d: carries the backticked mutation-marker span %s, which is not the canonical marker `%s`\n", FILENAME, ln, span, marker
        rest = substr(rest, RSTART + RLENGTH)
      }
    }
    { scan($0, NR); if (NR > 1) scan(prev " " $0, NR - 1); prev = $0 }
  ' "$file"
}

for pinned in "$RUNNER" "$PANEL"; do
  while IFS= read -r violation; do
    [ -n "$violation" ] || continue
    VIOLATIONS+=("$violation")
  done < <(scan_near_miss "$pinned" | sort -u)
done

if [ "${#VIOLATIONS[@]}" -gt 0 ]; then
  printf '%s\n' "${VIOLATIONS[@]}"
  printf 'MUTATION-REPRODUCER-PIN-INVALID: %s — %s violation(s)\n' "$ROOT" "${#VIOLATIONS[@]}"
  exit 1
fi

printf 'MUTATION-REPRODUCER-PIN-OK: %s — marker %s window %s pinned across scripts/run-reproducer.sh and skills/flow/review-panel.md\n' "$ROOT" "'$MARKER'" "$WINDOW"
exit 0
