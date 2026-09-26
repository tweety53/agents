#!/usr/bin/env bash
# check-mutation-reproducer-pin.sh — pin the mutation-reproducer marker
# literal and its line window mechanically across the two files that state
# them (KAN-624, panel F5 of kan-568's deferred self-review).
#
# Why this exists: the exact literal `# mutation-reproducer` and the
# 10-line window are each stated in skills/flow/review-panel.md (the
# reproducer-convention paragraph and the mutation-testing brief),
# scripts/run-reproducer.sh (the header comment) and the runner's Go port
# stats/internal/guard/runreproducer.go, with no mechanical pin
# keeping them together — check-dispatch-paragraphs.sh covers the required
# dispatch paragraphs but not this pair, so the two statements could drift
# silently, one prose edit or one code edit at a time. This guard is the
# pin.
#
# The source of truth is the runner's own canonical declaration — the code
# that decides the convention at run time. Since KAN-760 the runner is Go
# (scripts/run-reproducer.sh is a shim exec'ing flow-guard), so the
# declaration is two lines of stats/internal/guard/runreproducer.go: the
# marker constant `rrMarker = "<MARKER>"` and declaresMutation's
# `range headLines(path, <N>)`. The guard extracts the marker and the
# window from those lines and then requires the stated sites to agree, two
# ways:
#
#   - PRESENCE — skills/flow/review-panel.md carries the marker backticked
#     and the window stated as "first <N> lines"; scripts/run-reproducer.sh
#     carries the marker backticked (its header comment) and the same
#     window statement.
#   - NEAR-MISS ABSENCE — presence alone cannot see a single site drift
#     while another copy of the same fact still reads correctly, which is
#     the one-edit-at-a-time drift this guard exists to catch (KAN-624
#     F1). So every "first <M> lines" statement and every backticked
#     `# mutation-...` span in any of the three files must carry exactly the
#     canonical value, on a single line or wrapped across two adjacent
#     lines (the mutation-testing brief's own window statement is
#     wrapped), and each offending occurrence is reported at its own
#     line.
#
# A change to any one site without the others is a drift: exit 1, named
# per file and line, until all sites move together. Reshaping a canonical
# declaration line itself — so that no line of that shape can be found, or
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
#   0  clean — the marker and the window agree across all three files
#   1  drift — a stated site disagrees with the canonical declaration
#   2  cannot answer — a pinned file missing, a symlink, not a regular
#      file or unreadable; a canonical declaration line absent or
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
SOURCE="$ROOT/stats/internal/guard/runreproducer.go"

for pinned in "$SOURCE" "$RUNNER" "$PANEL"; do
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

# decl_line <ERE> <what> — the one line of $SOURCE carrying half of the
# canonical declaration: exactly one is answerable. Zero means the runner
# stopped declaring its convention in this shape (moved, renamed or
# rewritten); more than one means the reading is ambiguous. Either way this
# guard refuses to guess.
decl_line() {
  local ere="$1" what="$2" lines rc count
  set +e
  lines="$(grep -nE -- "$ere" "$SOURCE")"
  rc=$?
  set -e
  if [ "$rc" -ge 2 ]; then
    die2 "$SOURCE" "grep exited $rc while scanning for the canonical declaration line — a failure to look, not an absence"
  fi
  count="$(printf '%s' "$lines" | (grep -c . || true))"
  if [ "$rc" -eq 1 ] || [ "$count" -eq 0 ]; then
    die2 "$SOURCE" "the canonical declaration line ($what) is absent — the pin cannot answer until it is re-pointed at the declaration's new shape"
  fi
  if [ "$count" -gt 1 ]; then
    die2 "$SOURCE" "carries $count lines matching the canonical declaration shape ($what) — the reading is ambiguous; keep exactly one"
  fi
  printf '%s\n' "$lines" | head -n 1 | cut -d: -f2-
}

MARKER_DECL="$(decl_line '(^|[[:space:]])rrMarker[[:space:]]*=' 'rrMarker = "<marker>"')"
WINDOW_DECL="$(decl_line 'range headLines\(path, ' 'range headLines(path, <N>)')"
# Extraction is grep -oE, never a bare sed substitution: sed's no-match
# answer is the input line itself, so an unextractable marker or window
# (the constant re-quoted as a raw string, or the window made a named
# constant) would sail past an empty-check and be misreported as drift with
# the whole line pasted in as the value (KAN-624 F2). No match here is the
# pin's re-point signal: exit 2, never a guess.
set +e
WINDOW_TOK="$(printf '%s' "$WINDOW_DECL" | grep -oE 'headLines\(path, [0-9]+\)')"
MARKER_TOK="$(printf '%s' "$MARKER_DECL" | grep -oE '= "[^"\\]*"[[:space:]]*$')"
set -e
if [ -z "$WINDOW_TOK" ] || [ -z "$MARKER_TOK" ]; then
  die2 "$SOURCE" "the canonical declaration lines matched but the marker or window could not be extracted from them — their shape moved; re-point the pin: $MARKER_DECL / $WINDOW_DECL"
fi
WINDOW="$(printf '%s' "$WINDOW_TOK" | sed -E 's/^headLines\(path, //; s/\)$//')"
MARKER="$(printf '%s' "$MARKER_TOK" | sed -E 's/^= "//; s/"[[:space:]]*$//')"

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

PANEL_MARKER_MSG="no longer states the marker literal backticked — it drifted from stats/internal/guard/runreproducer.go's canonical declaration (marker: '$MARKER'); move every site together"
PANEL_WINDOW_MSG="no longer states the window as \"first $WINDOW lines\" — it drifted from stats/internal/guard/runreproducer.go's canonical declaration (window: $WINDOW); move every site together"
RUNNER_MARKER_MSG="its header comment no longer states the marker literal backticked — it drifted from stats/internal/guard/runreproducer.go's canonical declaration (marker: '$MARKER'); move every site together"
RUNNER_WINDOW_MSG="its header comment no longer states the window as \"first $WINDOW lines\" — it drifted from stats/internal/guard/runreproducer.go's canonical declaration (window: $WINDOW); move every site together"

check_contains "$PANEL" "\`$MARKER\`" "$PANEL_MARKER_MSG"   '# mutation-[a-z0-9][a-z0-9-]*'
check_contains "$PANEL" "first $WINDOW lines" "$PANEL_WINDOW_MSG"  'first [0-9]+ lines'
check_contains "$RUNNER" "\`$MARKER\`" "$RUNNER_MARKER_MSG" '# mutation-[a-z0-9][a-z0-9-]*'
check_contains "$RUNNER" "first $WINDOW lines" "$RUNNER_WINDOW_MSG" 'first [0-9]+ lines'

# Near-miss absence — the presence checks above cannot see a single site
# drifting while another copy of the same fact still reads correctly, which
# is the one-edit-at-a-time drift this guard exists to catch (KAN-624 F1:
# the real runner states its window more than once, the panel carries the marker
# three times, and the brief's window statement is wrapped across two
# lines where a single-line needle never looks). So every numeric window
# statement and every backticked mutation-marker span in any pinned
# file must carry exactly the canonical value — on a single line or
# wrapped across two adjacent lines — and each offending occurrence is
# reported at its own line (KAN-624 F4: a drift row names an editable
# line, never :0).
scan_near_miss() {
  local file="$1"
  awk -v win="$WINDOW" -v marker="$MARKER" '
    function scan(s, ln, b,    cross, n, rest, span) {
      rest = s
      while (match(rest, /first [0-9]+ lines/)) {
        # b is the join boundary: 0 on a single-line scan, else the index
        # of the inserted space in "prev line current line". A joined scan
        # reports only a match that SPANS the boundary — one found wholly
        # inside either half belongs to that half s own single-line scan,
        # and reporting it here would duplicate it at the previous line s
        # number, a phantom row naming a line that carries no statement
        # (KAN-624 F5).
        cross = (b == 0) || (RSTART < b && RSTART + RLENGTH - 1 > b)
        if (cross) {
          n = substr(rest, RSTART + 6, RLENGTH - 12) + 0
          if (n != win + 0)
            printf "%s:%d: states the window as \"first %d lines\" — a near-miss of the canonical window %s; every window statement must read \"first %s lines\"\n", FILENAME, ln, n, win, win
        }
        rest = substr(rest, RSTART + RLENGTH)
      }
      rest = s
      while (match(rest, /`# mutation-[a-z0-9][a-z0-9-]*`/)) {
        cross = (b == 0) || (RSTART < b && RSTART + RLENGTH - 1 > b)
        if (cross) {
          span = substr(rest, RSTART, RLENGTH)
          if (span != "`" marker "`")
            printf "%s:%d: carries the backticked mutation-marker span %s, which is not the canonical marker `%s`\n", FILENAME, ln, span, marker
        }
        rest = substr(rest, RSTART + RLENGTH)
      }
    }
    { scan($0, NR, 0); if (NR > 1) scan(prev " " $0, NR - 1, length(prev) + 1); prev = $0 }
  ' "$file"
}

for pinned in "$SOURCE" "$RUNNER" "$PANEL"; do
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

printf 'MUTATION-REPRODUCER-PIN-OK: %s — marker %s window %s pinned across stats/internal/guard/runreproducer.go, scripts/run-reproducer.sh and skills/flow/review-panel.md\n' "$ROOT" "'$MARKER'" "$WINDOW"
exit 0
