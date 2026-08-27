#!/usr/bin/env bash
# check-visual-trigger.sh — decide whether a diff touched a project's
# declared `## visual verification` UI paths, in a guard instead of prose.
#
# Usage: check-visual-trigger.sh <project root>   (changed paths on stdin,
#                                                    one per line)
#
# Replaces flow.visual-verify's step 2 (skills/flow/verify-and-handoff.md):
# "run `git diff --name-only` and test every printed path as a glob against
# the comma-separated `ui paths`." Two runs reading the same globs
# differently was the defect this guard exists to make impossible — the
# glob semantics are the whole risk, per tasks.md task 9.
#
# THREE EXIT CODES, AND "NOT CONFIGURED" IS NEVER "NO MATCH":
#   0  at least one stdin path matched at least one `ui paths` glob
#   1  the section resolved cleanly and `ui paths` is non-empty, but no
#      stdin path (including an empty stdin — no diff at all) matched
#   2  cannot answer: the project root is not a directory, `.flow/project.md`
#      is missing, not a regular file, or unreadable, the heading scan
#      itself failed to look, the section is declared twice (ambiguous —
#      neither is read), or `ui paths` is absent or empty.
# "Not configured" (2) and "configured and unmatched" (1) are deliberately
# different answers — flow.visual-verify's step 1 and step 2 print two
# different messages and skip the stage for two different reasons.
#
# GLOB SEMANTICS, pinned because they already caused disagreement once:
#   - `**` matches any run of characters INCLUDING `/`, so it spans
#     multiple directory levels, not one segment.
#   - a bare `*` matches any run of characters EXCEPT `/` — one path
#     segment. (Every glob this repository declares today uses `**`; a
#     bare `*` is supported for completeness, not because a declaration
#     uses one.)
#   - `?` matches exactly one character except `/`.
#   - a leading `./` is stripped from BOTH the declared glob and the
#     candidate path before matching — a habit either side might write.
#   - a leading `/` on a DECLARED glob is stripped too: `.flow/project.md`
#     paths are inherently repository-root-relative, so `/stats/web/**` and
#     `stats/web/**` mean the same thing.
#   - an ABSOLUTE candidate path is never stripped and so never matches a
#     relative glob. `git diff --name-only` never emits one; a caller that
#     hands this guard one gets a defined refusal, not a silent grant.
#   - splitting `ui paths` on its separator comma never touches a SPACE
#     inside one glob — only leading/trailing whitespace around each
#     comma-separated element is trimmed.
#
# THE INPUT IS ATTACKER-INFLUENCED, exactly as check-visual-verification.sh's
# own header states: `.flow/project.md` is tracked and editable in any pull
# request. This guard never executes anything it reads — `ui paths` values
# are matched as glob patterns, never interpolated into a shell command —
# and every interpolated cell that reaches an exit-2 message passes through
# sanitize_display first, for the identical forged-verdict reason
# check-visual-verification.sh's own copy of that function documents.
#
# THE HEADING SCAN AND TABLE PARSER ARE check-visual-verification.sh's OWN
# CONVENTIONS, reused rather than reinvented: the same `## visual
# verification` heading regex, the same "a heading at or above this
# section's own level ends it, a deeper one is prose" rule, and the same
# `|`-row split with `\|`-escape awareness. THE PARSER'S split_cells/
# trimcell/foldcell TRIO IS A SOURCED HELPER, unlike
# check-workspace-isolation.sh's own (wider) copy of the same shape — see
# scripts/lib/visual-table-cells.awk's header for why this guard, reached
# only through the skills/*/scripts/ symlink farm, may source a sibling
# instead of carrying its own copy.
#
# A LEADING UTF-8 BOM IS STRIPPED BEFORE EITHER HEADING READ OR THE TABLE
# PARSE, via the sourced strip_bom_cat (scripts/lib/strip-bom.sh) — see that
# file's header for why a BOM would otherwise make this guard read a present
# `## visual verification` heading as absent and exit 2 ("not configured"),
# which is worse here than in the other two guards: this is the guard that
# decides whether flow.visual-verify runs at all, so that misread is a
# silent skip of the whole stage, not merely a misparse.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ "$#" -ne 1 ]; then
  echo "check-visual-trigger: usage: check-visual-trigger.sh <project root> (changed paths on stdin)" >&2
  exit 2
fi
ROOT="$1"

if [ ! -d "$ROOT" ]; then
  echo "check-visual-trigger: $ROOT is not a directory — cannot tell whether it declares a visual verification section" >&2
  exit 2
fi

CFG="$ROOT/.flow/project.md"

# sanitize_display — sourced from lib/sanitize-display.sh; see that file's
# header for why this guard, unlike a hand-copied one, may source a sibling
# instead of carrying its own copy.
source "$SCRIPT_DIR/lib/sanitize-display.sh"

# strip_bom_cat — sourced from lib/strip-bom.sh; see that file's header for
# why a leading UTF-8 BOM would otherwise make a present `## visual
# verification` heading read as absent, and why this guard, unlike a
# hand-copied one, may source a sibling instead of carrying its own copy.
source "$SCRIPT_DIR/lib/strip-bom.sh"

# "Not configured" (no `.flow/project.md` at all) is the SAME cannot-answer
# case as "file exists but declares no section" below — both are exit 2,
# never exit 1's "configured and unmatched".
if [ ! -e "$CFG" ] && [ ! -L "$CFG" ]; then
  echo "check-visual-trigger: $ROOT has no .flow/project.md — not configured, so nothing was matched" >&2
  exit 2
fi

if [ ! -f "$CFG" ]; then
  echo "check-visual-trigger: $CFG is not a regular file — cannot resolve what it declares" >&2
  exit 2
fi
if [ ! -r "$CFG" ]; then
  echo "check-visual-trigger: $CFG exists but is not readable — cannot resolve what it declares" >&2
  exit 2
fi

VV_HEADING='^##[[:space:]]+visual verification[[:space:]]*$'

set +e
grep -qiE "$VV_HEADING" <(strip_bom_cat "$CFG")
VV_GREP_RC=$?
set -e
if [ "$VV_GREP_RC" -ge 2 ]; then
  echo "check-visual-trigger: grep exited $VV_GREP_RC while looking for the '## visual verification' heading in $CFG — that is a failure to look, not an absence" >&2
  exit 2
fi
if [ "$VV_GREP_RC" -ne 0 ]; then
  echo "check-visual-trigger: $CFG declares no '## visual verification' section — not configured, so nothing was matched" >&2
  exit 2
fi

set +e
VV_COUNT="$(grep -ciE "$VV_HEADING" <(strip_bom_cat "$CFG"))"
VV_COUNT_RC=$?
set -e
case "$VV_COUNT_RC:$VV_COUNT" in
  0:[0-9]*) ;;
  *)
    echo "check-visual-trigger: could not count the '## visual verification' headings in $CFG (grep exited $VV_COUNT_RC) — cannot resolve what it declares" >&2
    exit 2
    ;;
esac
if [ "$VV_COUNT" -gt 1 ]; then
  echo "check-visual-trigger: $CFG declares $VV_COUNT '## visual verification' sections — a second declaration is ambiguous, so neither was read" >&2
  exit 2
fi

# The whole of the "ui paths" extraction. Prints exactly one line,
# `VAL\t<value>`, the first `ui paths` row found inside the section — or
# nothing at all if no such row exists. This guard does not re-validate the
# rest of the section's shape; check-visual-verification.sh already does
# that, and this guard's own job is narrower: resolve one field or say it
# cannot.
#
# NO APOSTROPHE MAY APPEAR ANYWHERE BELOW, comments included — the same trap
# check-visual-verification.sh's own header records: the program is
# delimited by single quotes, so one ends it mid-word.
UI_PATHS_LINE="$(awk -v heading_re="$VV_HEADING" \
  -f "$SCRIPT_DIR/lib/visual-table-cells.awk" \
  -f <(cat <<'AWK_PROG'
  BEGIN { in_sec = 0; SEC_LEVEL = 2; done = 0 }

  done { next }

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
    if (key == "ui paths") {
      printf "VAL\t%s\n", trimcell(cells[2])
      done = 1
    }
  }
AWK_PROG
  ) <(strip_bom_cat "$CFG"))"

UI_PATHS_VAL=""
if [ -n "$UI_PATHS_LINE" ]; then
  UI_PATHS_VAL="${UI_PATHS_LINE#VAL$'\t'}"
fi

if [ -z "$UI_PATHS_VAL" ]; then
  printf 'check-visual-trigger: `ui paths` is absent or empty in %s — cannot resolve what to match against\n' "$CFG" | sanitize_display >&2
  exit 2
fi

# Split on comma, trim leading/trailing whitespace per element, and NEVER
# trim an interior space — a glob containing one (a directory name with a
# space) survives the split intact.
IFS=',' read -r -a RAW_GLOBS <<< "$UI_PATHS_VAL"
GLOBS=()
for g in "${RAW_GLOBS[@]}"; do
  g="${g#"${g%%[![:space:]]*}"}"
  g="${g%"${g##*[![:space:]]}"}"
  [ -n "$g" ] && GLOBS+=("$g")
done

if [ "${#GLOBS[@]}" -eq 0 ]; then
  printf 'check-visual-trigger: `ui paths` in %s resolved to no usable glob\n' "$CFG" | sanitize_display >&2
  exit 2
fi

# glob_to_ere <glob> -> prints an ERE fragment on stdout. `**` becomes
# `.*` (spans `/`); a bare `*` becomes `[^/]*` (one segment); `?` becomes
# `[^/]`; every other character is escaped literally.
glob_to_ere() {
  local g="$1" out="" i=0 n=${#g} c nc
  while [ "$i" -lt "$n" ]; do
    c="${g:$i:1}"
    case "$c" in
      '*')
        nc="${g:$((i + 1)):1}"
        if [ "$nc" = '*' ]; then
          out+='.*'
          i=$((i + 2))
          continue
        fi
        out+='[^/]*'
        ;;
      '?')
        out+='[^/]'
        ;;
      '.' | '+' | '(' | ')' | '{' | '}' | '|' | '^' | '$' | '\' | '[' | ']')
        out+="\\$c"
        ;;
      *)
        out+="$c"
        ;;
    esac
    i=$((i + 1))
  done
  printf '%s' "$out"
}

# glob_match <path> <glob> -> true if <path> matches <glob> under the
# semantics documented above.
glob_match() {
  local path="$1" glob="$2" regex
  path="${path#./}"
  glob="${glob#./}"
  glob="${glob#/}"
  regex="$(glob_to_ere "$glob")"
  [[ "$path" =~ ^${regex}$ ]]
}

MATCHED=0
while IFS= read -r changed || [ -n "$changed" ]; do
  [ -n "$changed" ] || continue
  for glob in "${GLOBS[@]}"; do
    if glob_match "$changed" "$glob"; then
      printf 'MATCH: %s — matched `%s`\n' "$changed" "$glob" | sanitize_display
      MATCHED=1
    fi
  done
done

if [ "$MATCHED" -eq 1 ]; then
  printf 'VISUAL-TRIGGER-MATCH: %s — at least one changed path matched `ui paths`\n' "$ROOT"
  exit 0
fi

printf 'VISUAL-TRIGGER-NO-MATCH: %s — no changed path matched `ui paths`\n' "$ROOT"
exit 1
