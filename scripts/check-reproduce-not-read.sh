#!/usr/bin/env bash
# check-reproduce-not-read.sh — the guard that keeps KAN-289's REPRODUCE,
# DON'T READ dispatch paragraph from silently disappearing.
#
# Why this exists: KAN-289's root cause was that the instruction lived in no
# template and depended on the dispatcher remembering to type it. Part 1 of
# that change fixes it by adding the paragraph, verbatim, at three dispatch
# sites in two files. Nothing then stops a LATER prose edit from trimming
# that paragraph away one line at a time — this guard is what makes that
# loud instead of silent, the same role check-guard-symlinks.sh and
# check-vocabulary.sh play for their own templates.
#
# Argument-free and self-scoped, exactly like check-guard-symlinks.sh: the
# scan root is this repository's own root, resolved from this script's own
# location, so no call site can narrow it. CHECK_REPRODUCE_NOT_READ_ROOT is
# an explicit, opt-in override honored only when set (mirrors
# CHECK_GUARD_SYMLINKS_ROOT), so the companion harness
# (test-check-reproduce-not-read.sh) can point this guard at a sandboxed
# fixture tree under TMPDIR without touching this repository — never set it
# for a normal invocation.
#
# THE REQUIRED-SITE TABLE, keyed by path relative to the scan root:
#
#   # site                              min blocks   required variant(s)
#   skills/flow/review-panel.md              1        reviewer
#   skills/flow/implement.md                 2        reviewer AND implementer
#
# A BLOCK is a line carrying the literal label `**REPRODUCE, DON'T READ:**`,
# plus every immediately-following line that continues the same markdown
# blockquote (a line beginning with `>`) — i.e. the whole paragraph. A
# VARIANT is identified by its own load-bearing phrases, held here as short
# literals rather than a copy of the whole paragraph (design.md's
# `label-plus-phrases`):
#
#   - shared, both variants: "crosses a boundary", "the store, the
#     filesystem, a guard, a real transcript", "exercise the real thing"
#   - reviewer only: "worth less than one you did"
#   - implementer only: "a fake or a hand-built value"
#
# A block counts toward a required variant only when it carries the label,
# every shared phrase, AND that variant's own phrase. Two blocks of the same
# variant do NOT satisfy a two-variant site — implement.md requires one
# reviewer block and one implementer block, not merely two blocks.
#
# Every failure is reported as `file:line` naming the missing element — the
# label, the specific missing phrase, or the missing variant. Exit `0`
# clean, `1` a required site missing its block or one of its phrases, `2` it
# cannot answer at all: a scoped file missing, unreadable, a symlink, or a
# grep failing for any reason other than "no match". A scoped path that
# exists as neither a file nor a directory is folded into the same hard `2`
# ("not a regular file") — never a silent skip, per check-vocabulary.sh's
# own header warning about a vacuous "✓ clean".
#
# WHAT A GREEN RUN DOES NOT PROVE: only that the label and its variant's
# phrases are present at each required site — never that a reviewer or an
# implementer actually obeyed the paragraph, and never that the wording
# reads well. A paraphrase that drops one of the listed phrases fails loud;
# a paraphrase that keeps all of them while changing the surrounding prose
# passes clean.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ -n "${CHECK_REPRODUCE_NOT_READ_ROOT:-}" ]; then
  ROOT="$CHECK_REPRODUCE_NOT_READ_ROOT"
elif [ "${CHECK_REPRODUCE_NOT_READ_ROOT+set}" = "set" ]; then
  echo "check-reproduce-not-read: CHECK_REPRODUCE_NOT_READ_ROOT is set but empty" >&2
  exit 2
else
  ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
fi

if [ ! -d "$ROOT" ] || [ ! -r "$ROOT" ] || [ ! -x "$ROOT" ]; then
  echo "check-reproduce-not-read: $ROOT is not a readable directory — cannot scan" >&2
  exit 2
fi

LABEL="**REPRODUCE, DON'T READ:**"
SHARED_PHRASES=(
  "crosses a boundary"
  "the store, the filesystem, a guard, a real transcript"
  "exercise the real thing"
)
REVIEWER_PHRASE="worth less than one you did"
IMPLEMENTER_PHRASE="a fake or a hand-built value"

SITE_PATHS=("skills/flow/review-panel.md" "skills/flow/implement.md")
SITE_MIN_BLOCKS=(1 2)
SITE_VARIANTS=("reviewer" "reviewer implementer")

# report_line <path> <line> <message> -> prints one "path:line: message" row.
report_line() {
  printf '%s:%s: %s\n' "$1" "$2" "$3"
}

# die2 <path> <message> -> a hard "cannot answer at all" refusal. Printed in
# the same file:line shape as every other finding this guard reports, then
# exits 2 immediately: an access failure at one required site means the
# whole run cannot answer, not that this one site failed.
die2() {
  report_line "$1" 0 "$2" >&2
  exit 2
}

VIOLATIONS=()

BLOCK_TEXT=""
# extract_block_text <lines-array-name> <n> <start_line_1based> -> sets
# BLOCK_TEXT to the label line's text, plus every immediately-following
# blockquote-continuation line, joined with single spaces (matching how a
# wrapped markdown paragraph reads as prose) so a phrase split across two
# source lines by the wrap is still found as one substring.
extract_block_text() {
  local -n _lines="$1"
  local n="$2" start="$3"
  local idx=$((start - 1))
  local first="${_lines[$idx]}"
  local is_bq=0
  case "$first" in
    '>'*) is_bq=1 ;;
  esac
  local text="" j="$idx" cur stripped
  while :; do
    cur="${_lines[$j]}"
    stripped="$cur"
    case "$cur" in
      '>'*) stripped="${cur#>}"; stripped="${stripped# }" ;;
    esac
    if [ -n "$text" ]; then text="$text $stripped"; else text="$stripped"; fi
    j=$((j + 1))
    [ "$is_bq" -eq 1 ] || break
    [ "$j" -lt "$n" ] || break
    case "${_lines[$j]}" in
      '>'*) : ;;
      *) break ;;
    esac
  done
  BLOCK_TEXT="$text"
}

for site_idx in "${!SITE_PATHS[@]}"; do
  site_path="${SITE_PATHS[$site_idx]}"
  min_blocks="${SITE_MIN_BLOCKS[$site_idx]}"
  required_variants="${SITE_VARIANTS[$site_idx]}"
  full="$ROOT/$site_path"

  if [ -L "$full" ]; then
    die2 "$full" "is a symlink — a required REPRODUCE, DON'T READ site must be a real file, never a symlink"
  fi
  if [ ! -e "$full" ]; then
    die2 "$full" "does not exist — this is a required REPRODUCE, DON'T READ site"
  fi
  if [ ! -f "$full" ]; then
    die2 "$full" "exists but is not a regular file (neither file nor a symlink to check) — cannot scan"
  fi
  if [ ! -r "$full" ]; then
    die2 "$full" "is not readable — cannot scan"
  fi

  set +e
  label_hits="$(grep -anF -- "$LABEL" "$full")"
  grep_rc=$?
  set -e
  if [ "$grep_rc" -ge 2 ]; then
    die2 "$full" "grep exited $grep_rc while scanning for the label — a failure to look, not an absence"
  fi

  mapfile -t LINES < "$full"
  n="${#LINES[@]}"

  block_count=0
  site_has_reviewer=0
  site_has_implementer=0

  if [ "$grep_rc" -eq 0 ]; then
    while IFS=: read -r lineno _rest; do
      [ -n "$lineno" ] || continue
      block_count=$((block_count + 1))

      extract_block_text LINES "$n" "$lineno"
      text="$BLOCK_TEXT"

      missing_shared=()
      for phrase in "${SHARED_PHRASES[@]}"; do
        case "$text" in
          *"$phrase"*) : ;;
          *) missing_shared+=("$phrase") ;;
        esac
      done
      for phrase in "${missing_shared[@]}"; do
        VIOLATIONS+=("$(report_line "$full" "$lineno" "block is missing the required phrase: \"$phrase\"")")
      done

      has_reviewer_phrase=0
      has_implementer_phrase=0
      case "$text" in
        *"$REVIEWER_PHRASE"*) has_reviewer_phrase=1 ;;
      esac
      case "$text" in
        *"$IMPLEMENTER_PHRASE"*) has_implementer_phrase=1 ;;
      esac

      if [ "${#missing_shared[@]}" -eq 0 ]; then
        [ "$has_reviewer_phrase" -eq 1 ] && site_has_reviewer=1
        [ "$has_implementer_phrase" -eq 1 ] && site_has_implementer=1
      fi
    done <<<"$label_hits"
  fi

  if [ "$block_count" -lt "$min_blocks" ]; then
    VIOLATIONS+=("$(report_line "$full" 0 "requires at least $min_blocks block(s) carrying the label \"$LABEL\", found $block_count")")
  fi

  for variant in $required_variants; do
    case "$variant" in
      reviewer)
        [ "$site_has_reviewer" -eq 1 ] || VIOLATIONS+=("$(report_line "$full" 0 "missing a block that satisfies the reviewer variant (the label plus every shared phrase plus \"$REVIEWER_PHRASE\")")")
        ;;
      implementer)
        [ "$site_has_implementer" -eq 1 ] || VIOLATIONS+=("$(report_line "$full" 0 "missing a block that satisfies the implementer variant (the label plus every shared phrase plus \"$IMPLEMENTER_PHRASE\")")")
        ;;
    esac
  done
done

if [ "${#VIOLATIONS[@]}" -gt 0 ]; then
  printf '%s\n' "${VIOLATIONS[@]}"
  printf 'REPRODUCE-NOT-READ-INVALID: %s — %s violation(s)\n' "$ROOT" "${#VIOLATIONS[@]}"
  exit 1
fi

printf 'REPRODUCE-NOT-READ-OK: %s — %s site(s) validated\n' "$ROOT" "${#SITE_PATHS[@]}"
exit 0
