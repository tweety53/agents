#!/usr/bin/env bash
# check-dispatch-paragraphs.sh — the guard that keeps a required dispatch
# paragraph from silently disappearing.
#
# Why this exists: KAN-289's root cause was that the REPRODUCE, DON'T READ
# instruction lived in no template and depended on the dispatcher
# remembering to type it. Part 1 of that change fixed it by adding the
# paragraph, verbatim, at dispatch sites. Nothing then stops a LATER prose
# edit from trimming a required paragraph away one line at a time — this
# guard is what makes that loud instead of silent, the same role
# check-guard-symlinks.sh and check-vocabulary.sh play for their own
# templates. KAN-217 added a second required paragraph — the VERBATIM
# REPORT — THE FACT blockquote that tells the fix subagent its slot's
# report file outranks the dispatcher's summary — and generalized this
# guard from one hard-coded paragraph to a table of them so a third would
# cost a table row, not a second copy of the machinery. KAN-263 added a
# third required paragraph — FOREGROUND BUILDS, which forbids a dispatched
# agent from ending its turn with a build/test/long-running command still
# running in the background — at all four dispatch sites that can run one:
# the implementer dispatch, the per-task reviewer dispatch, the review
# panel's slot dispatch, and the panel-fix subagent dispatch.
#
# Argument-free and self-scoped, exactly like check-guard-symlinks.sh: the
# scan root is this repository's own root, resolved from this script's own
# location, so no call site can narrow it. CHECK_DISPATCH_PARAGRAPHS_ROOT is
# an explicit, opt-in override honored only when set (mirrors
# CHECK_GUARD_SYMLINKS_ROOT), so the companion harness
# (test-check-dispatch-paragraphs.sh) can point this guard at a sandboxed
# fixture tree under TMPDIR without touching this repository — never set it
# for a normal invocation.
#
# THE PARAGRAPH TABLE. Each entry is a label, its shared phrases (required
# of every block carrying that label, held as short literals rather than a
# copy of the whole blockquote, per design.md's `label-plus-phrases`), and
# its variants — each variant a name plus its own load-bearing phrase, or no
# variants at all when every block of that label is equivalent. Every entry
# lists the sites that require it, each site naming its minimum block count
# and which variants it requires.
#
#   Paragraph                          Site                        Min  Variants
#   **REPRODUCE, DON'T READ:**         skills/flow/review-panel.md  1   reviewer
#   **REPRODUCE, DON'T READ:**         skills/flow/implement.md     2   reviewer AND implementer
#   **VERBATIM REPORT — THE FACT:**    skills/flow/review-panel.md  1   (none)
#   **FOREGROUND BUILDS:**             skills/flow/implement.md     2   (none)
#   **FOREGROUND BUILDS:**             skills/flow/review-panel.md  2   (none)
#
#   REPRODUCE, DON'T READ shared phrases: "crosses a boundary", "the store,
#   the filesystem, a guard, a real transcript", "exercise the real thing"
#     - reviewer variant's own phrase: "worth less than one you did"
#     - implementer variant's own phrase: "a fake or a hand-built value"
#
#   VERBATIM REPORT — THE FACT shared phrases (no variants — every block
#   carrying the label must carry all three): "the reviewer's own report",
#   "never a source of fact", "the report wins"
#
#   FOREGROUND BUILDS shared phrases (no variants — every block carrying the
#   label must carry all three): "still executing in the background", "Run
#   it in the foreground", "poll it to completion". Required twice in each
#   of implement.md (implementer dispatch, per-task reviewer dispatch) and
#   review-panel.md (panel slot dispatch, panel-fix subagent dispatch).
#
# A BLOCK is a line carrying a label, plus every immediately-following line
# that continues the same markdown blockquote (a line beginning with `>`) —
# i.e. the whole paragraph. A block counts toward a required variant only
# when it carries the label, every shared phrase, AND that variant's own
# phrase. Two blocks of the same variant do NOT satisfy a two-variant site —
# implement.md requires one reviewer block and one implementer block, not
# merely two blocks. An entry with no variants needs no variant match: every
# phrase it lists is simply required of the block.
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
# WHAT A GREEN RUN DOES NOT PROVE: only that each label and its phrases are
# present at each required site — never that a dispatcher actually wrote the
# report file described by VERBATIM REPORT — THE FACT, never that a fix
# agent read one, and never that a reviewer or an implementer actually
# obeyed REPRODUCE, DON'T READ. A paraphrase that drops one of the listed
# phrases fails loud; a paraphrase that keeps all of them while changing the
# surrounding prose passes clean.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ -n "${CHECK_DISPATCH_PARAGRAPHS_ROOT:-}" ]; then
  ROOT="$CHECK_DISPATCH_PARAGRAPHS_ROOT"
elif [ "${CHECK_DISPATCH_PARAGRAPHS_ROOT+set}" = "set" ]; then
  echo "check-dispatch-paragraphs: CHECK_DISPATCH_PARAGRAPHS_ROOT is set but empty" >&2
  exit 2
else
  ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
fi

if [ ! -d "$ROOT" ] || [ ! -r "$ROOT" ] || [ ! -x "$ROOT" ]; then
  echo "check-dispatch-paragraphs: $ROOT is not a readable directory — cannot scan" >&2
  exit 2
fi

# The unit separator, used to delimit phrases within one entry's
# shared-phrase string — phrases themselves contain ordinary spaces.
US=$'\x1f'

declare -A ENTRY_LABEL=(
  [reproduce]="**REPRODUCE, DON'T READ:**"
  [verbatim]="**VERBATIM REPORT — THE FACT:**"
  [foreground]="**FOREGROUND BUILDS:**"
)

declare -A ENTRY_SHARED_PHRASES=(
  [reproduce]="crosses a boundary${US}the store, the filesystem, a guard, a real transcript${US}exercise the real thing"
  [verbatim]="the reviewer's own report${US}never a source of fact${US}the report wins"
  [foreground]="still executing in the background${US}Run it in the foreground${US}poll it to completion"
)

# Variant names per entry, space-separated; empty means the entry has no
# variants — every block of that label is equivalent.
declare -A ENTRY_VARIANTS=(
  [reproduce]="reviewer implementer"
  [verbatim]=""
  [foreground]=""
)

# Each variant's own load-bearing phrase, keyed "<entry>:<variant>".
declare -A VARIANT_PHRASE=(
  [reproduce:reviewer]="worth less than one you did"
  [reproduce:implementer]="a fake or a hand-built value"
)

# Sites: parallel arrays. Each site names the entry (paragraph) it
# requires, the site's path relative to ROOT, its minimum block count, and
# the variants it requires (space-separated; empty means none required
# beyond the shared phrases).
SITE_ENTRY=(reproduce reproduce verbatim foreground foreground)
SITE_PATHS=("skills/flow/review-panel.md" "skills/flow/implement.md" "skills/flow/review-panel.md" "skills/flow/implement.md" "skills/flow/review-panel.md")
SITE_MIN_BLOCKS=(1 2 1 2 2)
SITE_VARIANTS=("reviewer" "reviewer implementer" "" "" "")

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
#
# The continuation walk stops at the first line that ITSELF carries any
# known label (from ENTRY_LABEL), even though that line still begins with
# '>'. Without this, two required blocks glued together by a missing blank
# line (a future edit's copy-paste slip) merge into one block, and a
# paraphrase in the first that quietly drops a required phrase can pass by
# absorbing the second block's own text — the exact failure class this
# guard exists to catch, defeated by its own continuation rule.
extract_block_text() {
  local -n _lines="$1"
  local n="$2" start="$3"
  local idx=$((start - 1))
  local first="${_lines[$idx]}"
  local is_bq=0
  case "$first" in
    '>'*) is_bq=1 ;;
  esac
  local text="" j="$idx" cur stripped next lbl starts_new_label
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
    next="${_lines[$j]}"
    case "$next" in
      '>'*) : ;;
      *) break ;;
    esac
    starts_new_label=0
    for lbl in "${ENTRY_LABEL[@]}"; do
      case "$next" in
        *"$lbl"*) starts_new_label=1; break ;;
      esac
    done
    [ "$starts_new_label" -eq 0 ] || break
  done
  BLOCK_TEXT="$text"
}

# Cache of scanned files, so a site sharing its path and label-hits with an
# earlier site in the table (review-panel.md appears twice) does not scan
# the file twice needlessly. Keyed by path; not required for correctness.
declare -A FILE_ACCESS_CHECKED=()

for site_idx in "${!SITE_PATHS[@]}"; do
  entry="${SITE_ENTRY[$site_idx]}"
  site_path="${SITE_PATHS[$site_idx]}"
  min_blocks="${SITE_MIN_BLOCKS[$site_idx]}"
  required_variants="${SITE_VARIANTS[$site_idx]}"
  label="${ENTRY_LABEL[$entry]}"
  IFS="$US" read -r -a shared_phrases <<<"${ENTRY_SHARED_PHRASES[$entry]}"
  entry_variants="${ENTRY_VARIANTS[$entry]}"
  full="$ROOT/$site_path"

  if [ -L "$full" ]; then
    die2 "$full" "is a symlink — a required dispatch-paragraph site must be a real file, never a symlink"
  fi
  if [ ! -e "$full" ]; then
    die2 "$full" "does not exist — this is a required dispatch-paragraph site"
  fi
  if [ ! -f "$full" ]; then
    die2 "$full" "exists but is not a regular file (neither file nor a symlink to check) — cannot scan"
  fi
  if [ ! -r "$full" ]; then
    die2 "$full" "is not readable — cannot scan"
  fi

  set +e
  label_hits="$(grep -anF -- "$label" "$full")"
  grep_rc=$?
  set -e
  if [ "$grep_rc" -ge 2 ]; then
    die2 "$full" "grep exited $grep_rc while scanning for the label \"$label\" — a failure to look, not an absence"
  fi

  mapfile -t LINES < "$full"
  n="${#LINES[@]}"

  block_count=0
  declare -A site_has_variant=()
  for v in $entry_variants; do site_has_variant[$v]=0; done

  if [ "$grep_rc" -eq 0 ]; then
    while IFS=: read -r lineno _rest; do
      [ -n "$lineno" ] || continue
      block_count=$((block_count + 1))

      extract_block_text LINES "$n" "$lineno"
      text="$BLOCK_TEXT"

      missing_shared=()
      for phrase in "${shared_phrases[@]}"; do
        case "$text" in
          *"$phrase"*) : ;;
          *) missing_shared+=("$phrase") ;;
        esac
      done
      for phrase in "${missing_shared[@]}"; do
        VIOLATIONS+=("$(report_line "$full" "$lineno" "block carrying \"$label\" is missing the required phrase: \"$phrase\"")")
      done

      if [ "${#missing_shared[@]}" -eq 0 ]; then
        for v in $entry_variants; do
          vphrase="${VARIANT_PHRASE[$entry:$v]}"
          case "$text" in
            *"$vphrase"*) site_has_variant[$v]=1 ;;
          esac
        done
      fi
    done <<<"$label_hits"
  fi

  if [ "$block_count" -lt "$min_blocks" ]; then
    VIOLATIONS+=("$(report_line "$full" 0 "requires at least $min_blocks block(s) carrying the label \"$label\", found $block_count")")
  fi

  for variant in $required_variants; do
    if [ "${site_has_variant[$variant]:-0}" -ne 1 ]; then
      vphrase="${VARIANT_PHRASE[$entry:$variant]}"
      VIOLATIONS+=("$(report_line "$full" 0 "missing a block that satisfies the $variant variant (the label \"$label\" plus every shared phrase plus \"$vphrase\")")")
    fi
  done
  unset site_has_variant
done

if [ "${#VIOLATIONS[@]}" -gt 0 ]; then
  printf '%s\n' "${VIOLATIONS[@]}"
  printf 'DISPATCH-PARAGRAPHS-INVALID: %s — %s violation(s)\n' "$ROOT" "${#VIOLATIONS[@]}"
  exit 1
fi

printf 'DISPATCH-PARAGRAPHS-OK: %s — %s site(s) validated\n' "$ROOT" "${#SITE_PATHS[@]}"
exit 0
