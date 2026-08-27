#!/usr/bin/env bash
# resolve-visual-screenshots.sh — resolve a captured spec's PNGs, in a guard
# instead of prose.
#
# Usage: resolve-visual-screenshots.sh <project root> <capture spec basename>
#
# Replaces flow.visual-verify's step 7 resolution prose
# (skills/flow/verify-and-handoff.md): "search recursively beneath
# `screenshots` for every PNG whose path contains the capture spec's own
# basename." Reading, still what a script cannot do, stays the agent's job —
# this guard only resolves WHICH paths to read.
#
# Prints one ABSOLUTE PNG path per line, nothing else, on a match — the
# stage feeds this straight into "read every captured PNG", so stdout here
# is data, not a report.
#
# THREE EXIT CODES:
#   0  at least one PNG matched; its path(s) are on stdout
#   1  zero matches — the case the stage must block on, per design.md. A
#      `screenshots` directory that does not exist YET is the identical
#      answer: `capture` creates it on its first run, so its absence is
#      `capture`'s own success path, not a guard failure.
#   2  cannot answer: usage error, the project root is not a directory,
#      `.flow/project.md` is missing/not a regular file/unreadable, the
#      heading scan failed to look, the section is declared twice, the
#      `screenshots` setting is absent or empty, or a declared `regression
#      checkout` does not resolve to an existing directory (so there is no
#      base to search relative to at all).
#
# `screenshots` RESOLVES RELATIVE TO THE REGRESSION CHECKOUT WHEN ONE IS
# DECLARED, OTHERWISE THE PROJECT ROOT — design.md's "The contract" table,
# cited rather than restated: this guard enforces it, it does not explain
# it. This guard does not otherwise validate the section's shape (that a
# declared `regression checkout`/`regression repo` pair is well formed, or
# that `regression repo` matches the checkout's real `origin`) —
# check-visual-verification.sh already does, and this guard's job is
# narrower: resolve one search root or say it cannot.
#
# THE FILTER READS THE WHOLE PATH, WITH NO SPECIAL-CASED EXCLUSION OF
# ANYTHING — a `screenshots` root wide enough to sweep in a `node_modules`
# tree still filters correctly, because "some path segment starts with the
# spec basename" is exactly as narrow as the stage's own step 7 says it is.
# Pinned rather than assumed: task 10's harness plants a decoy PNG inside
# `node_modules` whose name deliberately DOES start with the spec basename
# and confirms it is found — the filter is not a `node_modules` carve-out
# that could hide a real match sitting there too (a project genuinely
# emitting PNGs into a vendored path is not this guard's problem to solve).
#
# THE MATCH IS ANCHORED AT A PATH-SEGMENT BOUNDARY, not a bare substring
# test — task 16's own panel-demonstrated defect. `visual-baseline.spec.ts`
# CONTAINS `baseline.spec.ts` as a substring, so a naive `case "$png" in
# *"$SPEC_BASENAME"*)` filter resolving `baseline.spec.ts` also returned
# `visual-baseline.spec.ts-snapshots/`'s PNGs — a different spec's own
# baseline, sitting under a directory that merely happens to end with the
# queried name. Splitting the path on `/` and testing whether any SEGMENT
# starts with `$SPEC_BASENAME` closes that: `visual-baseline.spec.ts-…`
# starts with `visual-`, not `baseline.spec.ts`, so it is correctly
# rejected, while `baseline.spec.ts-snapshots` and `baseline.spec.ts.png`
# (the node_modules decoy above) both still match, because each IS the
# start of its own segment. Only the START is anchored — the string after
# the spec basename within a segment is unconstrained on purpose, which is
# what lets both the real `-snapshots` suffix and a literal `.png` extension
# still match.
#
# THE INPUT IS ATTACKER-INFLUENCED, exactly as check-visual-verification.sh's
# own header states: `.flow/project.md` is tracked and editable in any pull
# request. Every interpolated cell that reaches an exit-2 message passes
# through sanitize_display first, for the identical forged-verdict reason
# that file's own copy of the function documents. The PNG paths this guard
# prints on success are filesystem paths this guard discovered by walking a
# real directory, never text lifted out of the config file, so they carry no
# such risk and are printed unescaped.
#
# THE HEADING SCAN AND TABLE PARSER ARE check-visual-verification.sh's OWN
# CONVENTIONS, reused rather than reinvented — see check-visual-trigger.sh's
# identical header note for why the parser's split_cells/trimcell/foldcell
# trio is a sourced helper here rather than a copy.
#
# A LEADING UTF-8 BOM IS STRIPPED BEFORE EITHER HEADING READ OR THE TABLE
# PARSE, via the sourced strip_bom_cat (scripts/lib/strip-bom.sh) — see that
# file's header for why a BOM would otherwise make this guard read a present
# `## visual verification` heading as absent and exit 2.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ "$#" -ne 2 ]; then
  echo "resolve-visual-screenshots: usage: resolve-visual-screenshots.sh <project root> <capture spec basename>" >&2
  exit 2
fi
ROOT="$1"
SPEC_BASENAME="$2"

if [ -z "$SPEC_BASENAME" ]; then
  echo "resolve-visual-screenshots: the capture spec basename argument is empty" >&2
  exit 2
fi

if [ ! -d "$ROOT" ]; then
  echo "resolve-visual-screenshots: $ROOT is not a directory — cannot resolve what it declares" >&2
  exit 2
fi
ROOT_ABS="$(cd "$ROOT" && pwd)"

CFG="$ROOT_ABS/.flow/project.md"

# sanitize_display — sourced from lib/sanitize-display.sh; see that file's
# header for why this guard, unlike a hand-copied one, may source a sibling
# instead of carrying its own copy.
source "$SCRIPT_DIR/lib/sanitize-display.sh"

# strip_bom_cat — sourced from lib/strip-bom.sh; see that file's header for
# why a leading UTF-8 BOM would otherwise make a present `## visual
# verification` heading read as absent, and why this guard, unlike a
# hand-copied one, may source a sibling instead of carrying its own copy.
source "$SCRIPT_DIR/lib/strip-bom.sh"

if [ ! -e "$CFG" ] && [ ! -L "$CFG" ]; then
  echo "resolve-visual-screenshots: $ROOT_ABS has no .flow/project.md — cannot resolve where screenshots live" >&2
  exit 2
fi
if [ ! -f "$CFG" ]; then
  echo "resolve-visual-screenshots: $CFG is not a regular file — cannot resolve what it declares" >&2
  exit 2
fi
if [ ! -r "$CFG" ]; then
  echo "resolve-visual-screenshots: $CFG exists but is not readable — cannot resolve what it declares" >&2
  exit 2
fi

VV_HEADING='^##[[:space:]]+visual verification[[:space:]]*$'

set +e
grep -qiE "$VV_HEADING" <(strip_bom_cat "$CFG")
VV_GREP_RC=$?
set -e
if [ "$VV_GREP_RC" -ge 2 ]; then
  echo "resolve-visual-screenshots: grep exited $VV_GREP_RC while looking for the '## visual verification' heading in $CFG — that is a failure to look, not an absence" >&2
  exit 2
fi
if [ "$VV_GREP_RC" -ne 0 ]; then
  echo "resolve-visual-screenshots: $CFG declares no '## visual verification' section — cannot resolve where screenshots live" >&2
  exit 2
fi

set +e
VV_COUNT="$(grep -ciE "$VV_HEADING" <(strip_bom_cat "$CFG"))"
VV_COUNT_RC=$?
set -e
case "$VV_COUNT_RC:$VV_COUNT" in
  0:[0-9]*) ;;
  *)
    echo "resolve-visual-screenshots: could not count the '## visual verification' headings in $CFG (grep exited $VV_COUNT_RC)" >&2
    exit 2
    ;;
esac
if [ "$VV_COUNT" -gt 1 ]; then
  echo "resolve-visual-screenshots: $CFG declares $VV_COUNT '## visual verification' sections — a second declaration is ambiguous, so neither was read" >&2
  exit 2
fi

# Extracts `screenshots` and `regression checkout` (whichever are present) in
# one pass. NO APOSTROPHE MAY APPEAR ANYWHERE BELOW, comments included —
# check-visual-verification.sh's own header records the same trap.
SECTION_REPORT="$(awk -v heading_re="$VV_HEADING" \
  -f "$SCRIPT_DIR/lib/visual-table-cells.awk" \
  -f <(cat <<'AWK_PROG'
  BEGIN { in_sec = 0; SEC_LEVEL = 2 }

  /^#+[[:space:]]/ {
    match($0, /^#+/)
    hlevel = RLENGTH
    is_own = (tolower($0) ~ heading_re)
    if (in_sec && !is_own && hlevel > SEC_LEVEL) { next }
    in_sec = is_own
    next
  }

  !in_sec { next }

  {
    n = split_cells($0, cells)
    if (n != 2) next
    key = foldcell(cells[1])
    if (key == "screenshots" && !("screenshots" in seen)) {
      seen["screenshots"] = 1
      printf "SCREENSHOTS\t%s\n", trimcell(cells[2])
    }
    if (key == "regression checkout" && !("checkout" in seen)) {
      seen["checkout"] = 1
      printf "CHECKOUT\t%s\n", trimcell(cells[2])
    }
  }
AWK_PROG
  ) <(strip_bom_cat "$CFG"))"

SCREENSHOTS_VAL=""
CHECKOUT_VAL=""
CHECKOUT_FOUND=0
while IFS=$'\t' read -r tag val; do
  [ "$tag" = "SCREENSHOTS" ] && SCREENSHOTS_VAL="$val"
  if [ "$tag" = "CHECKOUT" ]; then
    CHECKOUT_VAL="$val"
    CHECKOUT_FOUND=1
  fi
done <<< "$SECTION_REPORT"

if [ -z "$SCREENSHOTS_VAL" ]; then
  printf 'resolve-visual-screenshots: `screenshots` is absent or empty in %s — cannot resolve where captures land\n' "$CFG" | sanitize_display >&2
  exit 2
fi

BASE="$ROOT_ABS"
if [ "$CHECKOUT_FOUND" -eq 1 ]; then
  if [ ! -d "$CHECKOUT_VAL" ]; then
    printf 'resolve-visual-screenshots: `regression checkout` names `%s`, which is not an existing directory — cannot resolve `screenshots` relative to it\n' "$CHECKOUT_VAL" | sanitize_display >&2
    exit 2
  fi
  BASE="$(cd "$CHECKOUT_VAL" && pwd)"
fi

SEARCH_ROOT="$BASE/$SCREENSHOTS_VAL"

MATCHES_FILE="$(mktemp "${TMPDIR:-/tmp}/resolve-visual-screenshots.XXXXXX")"
trap 'rm -f "$MATCHES_FILE"' EXIT
: > "$MATCHES_FILE"

if [ -d "$SEARCH_ROOT" ]; then
  SEARCH_ROOT_ABS="$(cd "$SEARCH_ROOT" && pwd)"
  while IFS= read -r -d '' png; do
    IFS='/' read -r -a png_segments <<< "$png"
    for seg in "${png_segments[@]}"; do
      case "$seg" in
        "$SPEC_BASENAME"*)
          printf '%s\n' "$png" >> "$MATCHES_FILE"
          break
          ;;
      esac
    done
  done < <(find "$SEARCH_ROOT_ABS" -type f -name '*.png' -print0)
fi

sort -o "$MATCHES_FILE" "$MATCHES_FILE"

if [ ! -s "$MATCHES_FILE" ]; then
  echo "resolve-visual-screenshots: zero PNGs under $SEARCH_ROOT matched \`$SPEC_BASENAME\`" >&2
  exit 1
fi

cat "$MATCHES_FILE"
exit 0
