# scripts/lib/panel-record.sh — the panel-record marker helpers, defined once.
#
# Sourced by scripts/check-unfinished-work.sh and scripts/check-panel-reproducers.sh,
# which used to carry their own near-identical copies of these three functions.
# The copies had already drifted: check-unfinished-work.sh's `findings-total`
# pattern had no bound on the digit run, while check-panel-reproducers.sh's
# equivalent `reproducers-total` pattern was bounded at `{0,14}` after the
# crash that bound was written to fix. One definition, sourced by both, is what
# stops that drift from happening a second time.
#
# Not meant to be executed directly — a caller sources it and calls its
# functions; it sets no `set -euo pipefail` of its own and relies on the
# sourcing script's.
#
# THREE DISCIPLINES ARE STATED HERE ONCE, rather than in every guard that used
# to define these functions on its own:
#
#   `-a` ON EVERY GREP. A record carrying one stray NUL byte — a bad merge, a
#   truncated write — puts grep into binary mode, where it suppresses output
#   and reports "no match" whatever the file contains. Without `-a` a
#   corrupted record reads as a clean one.
#
#   THE `rc > 1` SPLIT. `grep`'s exit 1 means "no match", which is an answer
#   (an empty result), not a failure. An exit of 2 or more is a real error —
#   an unreadable file, a permission anomaly — and every function below
#   propagates that status to its caller rather than folding it into an empty
#   or zero result. A blanket `|| true` cannot tell the two apart and turns a
#   real error into "nothing matched here", which is the reassuring direction
#   a guard must never fail toward.
#
#   `--` BEFORE EVERY PATH. A record path beginning with `-` must never be
#   read by `grep` as an option.
#
# PANEL_RECORD_TOTAL_DIGITS — the bounded-integer sub-pattern both guards'
# declared-total lines share: a plain count with no leading zero, capped at 15
# digits (comfortably inside every shell's 64-bit arithmetic). Centralized
# here because it used to be written out inline, separately, at four call
# sites — check-unfinished-work.sh's `findings-total` line and
# check-panel-reproducers.sh's `reproducers-total` line, twice each — and had
# already drifted once between the two guards before this library existed
# (see the file header above). A caller wraps it in its own surrounding
# pattern text (the field name, the trailing `[[:space:]]*$`), which stays at
# the call site since that part is not shared knowledge.
PANEL_RECORD_TOTAL_DIGITS='(0|[1-9][0-9]{0,14})'

# PANEL_RECORD_DIR — the directory, relative to a worktree, that the review
# panel record is RENDERED into. It is `records.renderKinds`'s `panel` entry
# (stats/internal/records/render.go) expressed for the shell, and the two are
# the same fact: `myflow record render -kind panel` writes the file, and the
# two guards below read it.
#
# THE GUARDS USED TO READ `.superpowers/sdd/final-review-panel.md` INSTEAD, and
# that is now the wrong file. Once a change's findings became rows in the store,
# nothing writes a findings table or a marker block into the sdd path any more:
# it survives as the pass log alone — the mode, the slots, the diff path, the
# `fix-mutation:` lines and the bounces — and it carries no `findings-total:`
# line at all. A guard left reading it would have reported every change
# OUTSTANDING at finish run 1 and failed every fix round.
#
# THE RECORD IS IN ONE PLACE, NOT TWO. Rendering to both paths was considered
# and rejected: it would put the same record where two copies can disagree,
# which is the defect the move into the store exists to remove, and
# docs/superpowers/reviews/ is where finish run 1 commits it from in any case.
# The pass log is NOT a second copy of it — it is a different record, of a
# different thing, and stays where it is.
PANEL_RECORD_DIR='docs/superpowers/reviews'

# panel_record_path <worktree> <change> — the change's rendered review panel
# record inside <worktree>, or empty when the change has none. Never fails on a
# missing directory: "no record" is an answer both callers act on themselves.
#
# THE MATCH IS ANCHORED AT BOTH ENDS, and that is the whole of this function.
# `records.existingDatedFile` resolves the same file with the regexp
# `^[0-9]{4}-[0-9]{2}-[0-9]{2}-` + QuoteMeta(change + "-panel.md") + `$`, and
# the `case` pattern below is that expression character for character: four
# digits, two, two, then the change name and the suffix LITERALLY — `"$change"`
# is quoted inside the pattern, so a metacharacter in it could not be a
# wildcard even if the caller's own allowlist had not already rejected one. A
# looser `*-<change>-panel.md` would also match a DIFFERENT change whose name
# ends in this one — `2020-01-01-other-demo-panel.md` for the change `demo` —
# and that is not hypothetical: the renderer's own comment records `*` in a
# name matching and overwriting a different change's record. A guard reading
# the wrong change's findings is the same defect on the read side.
#
# THE EARLIEST DATE WINS, matching existingDatedFile's `sort.Strings` on the
# matching names followed by `found[0]`. The date is fixed at a change's FIRST
# render and a fix round overwrites in place, so more than one dated file for
# one change means something went wrong; the two sides agreeing on WHICH one to
# read is what keeps the guard reading the file the renderer writes. Bash
# expands a glob in sorted order and this guard's `LC_ALL=C` makes that order
# bytewise, so the first match encountered is that file.
panel_record_path() {
  local worktree="$1" change="$2" dir f base
  dir="$worktree/$PANEL_RECORD_DIR"
  [ -d "$dir" ] || return 0
  for f in "$dir"/*-panel.md; do
    # A glob that matched nothing expands to itself, and a directory sitting
    # where a record should be is not a record — `existingDatedFile` skips an
    # entry whose IsDir() is true for the same reason.
    [ -f "$f" ] || continue
    base="${f##*/}"
    case "$base" in
      [0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]-"$change"-panel.md)
        printf '%s\n' "$f"
        return 0
        ;;
    esac
  done
  return 0
}

# count_matching <file> <ere> — the number of lines of <file> matching <ere>.
count_matching() {
  local file="$1" ere="$2" out rc=0
  out="$(grep -acE -- "$ere" "$file")" || rc=$?
  if [ "$rc" -gt 1 ]; then
    return "$rc"
  fi
  printf '%s\n' "${out:-0}"
}

# grep_lines_of <file> <mode> <ere> — every match of <ere> in <file>, one per
# line, raw and not deduplicated.
#   mode "match" — `grep -o`: only the substring the pattern matched, for a
#     caller that needs to see repeats or extract identifiers.
#   mode "line"  — the whole matching line, for a caller that needs the full
#     command text.
# Empty, with a clean 0 return, when nothing matches — the same exit-code
# discipline as count_matching.
grep_lines_of() {
  local file="$1" mode="$2" ere="$3" raw rc=0 opt
  case "$mode" in
    match) opt='-aoE' ;;
    line) opt='-aE' ;;
    *)
      echo "grep_lines_of: unknown mode '$mode'" >&2
      return 2
      ;;
  esac
  raw="$(grep "$opt" -- "$ere" "$file")" || rc=$?
  if [ "$rc" -gt 1 ]; then
    return "$rc"
  fi
  [ -n "$raw" ] && printf '%s\n' "$raw"
  return 0
}

# ids_of <file> <ere> <shape> — the finding identifiers of every matching line
# in <file>, in one of two shapes. Both SHALL exist because the two callers
# ask different questions: a duplicate-identifier check needs a list that
# retains duplicates, and a set-comparison check needs one that does not. A
# library exporting only one shape would silently change a caller's answer.
#
#   shape "digits"     — bare digits, sorted, DUPLICATES RETAINED.
#                        check-unfinished-work.sh's repeated_ids reads this
#                        list and finds repeats with `uniq -d`; removing
#                        duplicates here would make that check pass vacuously.
#   shape "ids-unique" — F<n>, sorted and de-duplicated, for set comparison
#                        with `=` or `comm`.
ids_of() {
  local file="$1" ere="$2" shape="$3" raw
  raw="$(grep_lines_of "$file" match "$ere")" || return "$?"
  [ -n "$raw" ] || return 0
  case "$shape" in
    digits)     printf '%s' "$raw" | tr -cd '0-9\n' | sort ;;
    ids-unique) printf '%s\n' "$raw" | grep -aoE 'F[0-9]+' | sort -u ;;
    *) echo "ids_of: unknown shape '$shape'" >&2; return 2 ;;
  esac
}
