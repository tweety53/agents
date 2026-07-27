#!/usr/bin/env bash
# check-vocabulary.sh — drift guard for retired myflow vocabulary.
#
# Usage: scripts/check-vocabulary.sh [path ...]               # dirs or files
#
# With no arguments it scans the project's own documentation and command trees, resolved
# relative to the repo root — see DEFAULT_TARGETS below. That default is the ONE place the
# invocation lives: the rule file and the review skill say "run the script", they do not
# each carry a copy of the path list that can drift from this one.
#
# Scans the given directories and files for retired pipeline-stage vocabulary and retired
# review-panel vocabulary. Every hit is reported as file:line. Exits non-zero if
# either check fails, so it can gate CI or run on demand after a rename.
#
# A path that exists as neither a file nor a directory is a hard error (exit 2), never a
# silent skip: a CI job pointed at a renamed directory would otherwise stay green forever
# while checking nothing. So is a grep that fails for any reason other than "no match" —
# an unreadable directory or a broken pattern reported as "✓ clean" is that same vacuous
# pass one layer down.
#
# WHAT THIS CAN AND CANNOT DO — read before trusting a green run. This is a regression
# test for a fixed list of literals we know were retired. It proves those exact strings
# are absent. It does NOT prove a rename is complete: any paraphrase outside the list
# ("the full panel", "every reviewer slot", a roster written in a new order) passes clean,
# and so does every literal retired by the NEXT rename until someone adds it here. A green
# run means "the strings we listed have not come back" — nothing wider. Reviewing the diff
# for the rename's actual completeness is still a human job.
#
# Previously this lived inside a project-side file-edit hook, which had to fail
# open and so could only warn. As a standalone script it can fail loudly.

set -uo pipefail

die() {
  echo "ERROR: $*" >&2
  exit 2
}

# Temp files are registered here and removed by a single EXIT trap, so they are cleaned up on
# `die` and on a signal too, not only on the happy path.
TMP_FILES=()
cleanup_tmp() { [[ ${#TMP_FILES[@]} -eq 0 ]] || rm -f "${TMP_FILES[@]}"; }
trap cleanup_tmp EXIT

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# The default scan set: everything in this repo that can carry the vocabulary. Kept here
# and nowhere else, so callers only ever have to say `scripts/check-vocabulary.sh`.
DEFAULT_TARGETS=(skills rules commands commands-claude scripts README.md AGENTS.md CLAUDE.md)

if [[ $# -gt 0 ]]; then
  TREES=("$@")
else
  cd "$REPO_ROOT" || die "cannot enter repo root $REPO_ROOT"
  TREES=("${DEFAULT_TARGETS[@]}")
fi

for target in "${TREES[@]}"; do
  [[ -e "$target" ]] || die "no such file or directory: $target"
done

# The opt-out marker.  Some lines must name retired vocabulary to do their job: the legacy-stage
# migration mapping has to spell out the old values it maps FROM, the paragraph describing this
# guard has to quote the tokens it bans, and this script has to list them to search for them.
# Those are documentation, not drift.  Such a line carries a trailing `<!-- vocab-guard:allow -->`
# (or `# vocab-guard:allow` in a script), which both checks honour.  The marker is deliberately
# per-line rather than per-file: exempting a whole file would let real drift hide behind one
# documentary sentence.  The filter below therefore matches the marker only in the CONTENT field
# of grep's output — a marker in a *path* must never exempt everything under it.
#
# Separating the two needs an unambiguous boundary between the path field and the content field,
# and a colon is not one: `grep -n` emits `path:line:content`, so a directory literally named
# `d:0:vocab-guard:allow/` supplies a `:<digits>:` AND the marker from the PATH alone and exempts
# its whole subtree, no matter where the pattern is anchored. `--null` makes grep terminate the
# filename with a NUL instead, which cannot occur inside a path or a text line; SEP below is that
# NUL after `tr` maps it to a byte the shell can carry through a variable. So the filter can
# require the marker to sit strictly after the separator — i.e. in the content.
SEP=$'\001'
MARKER_FILTER="^[^$SEP]*$SEP"'[0-9]+:.*vocab-guard:allow'

# abs_dir <dir> — the physical absolute path of a directory, links resolved. `pwd -P` is used
# because `realpath` is absent from stock macOS.
abs_dir() { (cd "$1" 2>/dev/null && pwd -P); }

# Files reached through a symlink are required to resolve under the scan root or under this
# repo — see the `-L` note in collect_hits below.
REPO_ROOT_REAL="$(abs_dir "$REPO_ROOT")" || die "cannot resolve repo root $REPO_ROOT"

# collect_hits <path> <grep-args...> — search one path, minus opted-out lines.
#
# The result is published in the global HITS_OUT rather than written to stdout, because callers
# would otherwise have to capture it with `$(...)` — and a `die` inside a command substitution
# exits only that subshell, so the run would carry on and print "✓ clean" after the very failure
# it was meant to abort on.
#
# Files are enumerated by `find -L` rather than by grep's own `-r`/`-R`, because neither flag
# traverses symlinked directories portably: macOS BSD grep 2.6.0-FreeBSD descends a symlinked
# subdirectory under NEITHER -r nor -R (verified), while GNU grep and ugrep do under -R. Every
# installed tree (~/.claude/skills, ~/.cursor/skills) is a farm of symlinked directories, so
# relying on grep's recursion means "✓ clean" over a tree that was never opened. `find -L`
# follows links on both platforms.
#
# What `-L` costs, and why the walk is bounded: following links means the walk is NOT confined
# to the named target. A single committed link — `skills/anything/link -> /`, or `-> $HOME` —
# turns a bare invocation into a filesystem-wide scan that reports hits at paths outside the
# repo entirely (verified: a link under skills/ reported a file in another tree as if it were
# part of the target). So the escape is bounded rather than the feature removed:
#   - `-xdev` keeps the walk off other filesystems (mounted volumes, network shares).
#   - Every enumerated file is resolved and required to live under the scan root or under this
#     repo — the second root is what keeps the installed-farm case working, since those links
#     resolve back into this checkout. Anything else aborts, naming the path. Silently dropping
#     it would be a scan reporting "✓ clean" over files it never opened, i.e. this guard's own
#     failure mode; the user is told to scan that root directly instead.
#
# A symlink LOOP is left to find, which is loud on both platforms rather than silent: BSD find
# reports the loop and exits 0, GNU find prints "File system loop detected" and exits 1, so the
# rc check below turns the GNU case into a hard exit 2. Worth naming because the trees this
# scans are symlink farms.
#
# `-H` is passed because GNU grep omits the filename when handed a single file, which would both
# strip the path from every reported hit and silently disable the marker filter above; BSD grep
# includes it. `-H` makes the output shape the same everywhere.
#
# `-a` is passed so a file containing a NUL is still reported as `file:line:content`. Without it
# grep collapses the whole file to "Binary file … matches", which breaks the header's promise
# that every hit carries a location. `-I` (skip binary files) was rejected as the other way to
# make the shape uniform: it makes a file that IS scanned today silently unscanned, which is the
# vacuous pass this guard exists to prevent — a garbled content line is the cheaper failure.
#
# Failures abort instead of being swallowed. find's status is checked (unreadable directory,
# broken symlink, symlink loop) and grep's status is checked — 0 (matched) and 1 (no match) are
# normal, >= 2 is a real error — for the marker filter as well as for the search itself. Neither
# stderr is redirected, so the reason stays visible. A swallowed error reported as "✓ clean" is
# the same vacuous pass this guard exists to prevent.
HITS_OUT=""
collect_hits() {
  local target="$1"; shift
  local list_tmp raw rc filtered file root dir last_dir="" last_real=""
  local -a files=()
  HITS_OUT=""

  # For a file target the root is its directory: the file was named explicitly, so it cannot
  # be an escape. For a directory target it is the directory itself.
  if [[ -d "$target" ]]; then
    root="$(abs_dir "$target")" || die "cannot resolve scan target '$target'"
  else
    root="$(abs_dir "$(dirname "$target")")" || die "cannot resolve scan target '$target'"
  fi

  list_tmp="$(mktemp)"
  TMP_FILES+=("$list_tmp")
  find -L "$target" -xdev -type f -print0 >"$list_tmp"
  rc=$?
  (( rc == 0 )) || die "find exited $rc while enumerating '$target' (reason above) — refusing to report a clean run"
  while IFS= read -r -d '' file; do
    # Resolve per directory, not per file: find walks depth-first, so consecutive entries
    # almost always share a parent and the subshell is paid once per directory.
    dir="${file%/*}"
    [[ "$dir" != "$file" ]] || dir="."
    if [[ "$dir" != "$last_dir" ]]; then
      last_dir="$dir"
      last_real="$(abs_dir "$dir")" || die "cannot resolve '$dir' while enumerating '$target'"
    fi
    if [[ "$last_real" != "$root" && "$last_real" != "$root"/* &&
          "$last_real" != "$REPO_ROOT_REAL" && "$last_real" != "$REPO_ROOT_REAL"/* ]]; then
      die "'$file' is reached through a symlink that leaves the scanned tree (it resolves to
  $last_real/$(basename "$file"), outside both '$root' and $REPO_ROOT_REAL).
  Following it would scan an unbounded amount of the filesystem and report hits at paths
  that are not part of the target. Remove the link, or scan that root directly."
    fi
    files+=("$file")
  done <"$list_tmp"
  [[ ${#files[@]} -gt 0 ]] || return 0

  raw="$(grep -aHn --null "$@" "${files[@]}" | tr '\0' "$SEP")"
  rc=$?
  (( rc < 2 )) || die "grep exited $rc while scanning '$target' (reason above) — refusing to report a clean run"
  [[ -n "$raw" ]] || return 0
  filtered="$(printf '%s\n' "$raw" | grep -vE "$MARKER_FILTER")"
  rc=$?
  (( rc < 2 )) || die "grep exited $rc while applying the vocab-guard:allow filter for '$target' (reason above) — refusing to report a clean run"
  [[ -n "$filtered" ]] || return 0
  HITS_OUT="$(printf '%s\n' "$filtered" | tr "$SEP" ':')"
}

# Guard: retired stage vocabulary must never reappear in any command tree.
#
# A past rename updated the stage vocabulary in the rule file and the skills but never swept the
# COMMAND files, leaving four commands gating on stages that no longer existed and contradicting
# the skills they delegate to. This check makes that class of drift loud instead of silent.
#
# The legal stages are owned by `rules/myflow-manual-review.mdc` — see its `## Stage transitions`
# section. They are deliberately not copied here; a second list would drift from the first.
#
# Legitimate exceptions, deliberately allowed:
#   - `requesting-code-review` — a real external Superpowers skill; never rename it. The pattern
#     already excludes it structurally (the `-` before `code-review` fails `[^-]`), so no  # vocab-guard:allow
#     whole-line filter is needed — one that dropped the line would also hide a genuine retired
#     token sitting on the same line.
#   - "(not `start`)" in myflow-start.md — names the retired legacy value on purpose, to contrast.
#   - Any line carrying the marker `vocab-guard:allow` — see above.
check_retired_stage_vocabulary() {
  local pattern='awaiting-review|awaiting-test|(^|[^-])\bcode-review\b' # vocab-guard:allow
  local hits="" tree
  for tree in "${TREES[@]}"; do
    collect_hits "$tree" -E "$pattern"
    [[ -z "$HITS_OUT" ]] || hits+="$HITS_OUT"$'\n'
  done

  if [[ -n "$hits" ]]; then
    printf '\n⚠ Retired myflow stage vocabulary found:\n%s\n' "$hits" >&2
    printf 'Stage gates must match the rule file'"'"'s `## Stage transitions` table.\n' >&2
    return 1
  fi

  echo "✓ Stage-vocabulary guard: clean"
  return 0
}

# Guard: retired review-panel vocabulary must never reappear.
#
# The review panel was reduced and its reviewer slots renamed. Any surviving mention of a removed
# reviewer, the old panel size, or a retired pass name means one layer of the docs/commands/skills
# was left behind by the rename. The terms cover the ways the old roster was actually written —
# spelled-out slot names, roster fragments ("… + Senior", ", Conventions"), and every spelling of  # vocab-guard:allow
# the old panel size — because a narrow list passes clean over real drift.
#
# Matching stays case-SENSITIVE, and both capitalisations of the reviewer slot are listed
# separately as a result. Case-insensitive matching was tried and rejected: it flags the
# adversarial reviewer's own persona line ("You are an adversarial senior engineer"), which is
# prose about a reviewer's attitude, not the retired roster slot. The retired slot was always
# written capitalised.
#
# This remains a blacklist: see the header's "WHAT THIS CAN AND CANNOT DO". A paraphrase nobody
# thought to list still passes.
check_retired_panel_vocabulary() {
  local terms=(
    'Conventions & hygiene'   # vocab-guard:allow
    'senior-engineer-reviewer' # vocab-guard:allow
    'conventions-reviewer'    # vocab-guard:allow
    'all six'                 # vocab-guard:allow
    '1 primary + 5'           # vocab-guard:allow
    '1+5 agents'              # vocab-guard:allow
    '+ Senior'                # vocab-guard:allow
    ', Conventions'           # vocab-guard:allow
    'Economic Senior'         # vocab-guard:allow
    'two-agent'               # vocab-guard:allow
    '2-agent'                 # vocab-guard:allow
    'Architect pass'          # vocab-guard:allow
    ', Senior'                # vocab-guard:allow
    'Senior engineer'         # vocab-guard:allow
    'Senior Engineer'         # vocab-guard:allow
    'six-agent'               # vocab-guard:allow
    '6-agent'                 # vocab-guard:allow
    '5 additional'            # vocab-guard:allow
    'five additional'         # vocab-guard:allow
    'all 6'                   # vocab-guard:allow
  )
  local args=() term hits="" tree
  for term in "${terms[@]}"; do args+=(-e "$term"); done

  for tree in "${TREES[@]}"; do
    collect_hits "$tree" -F "${args[@]}"
    [[ -z "$HITS_OUT" ]] || hits+="$HITS_OUT"$'\n'
  done

  if [[ -n "$hits" ]]; then
    printf '\n⚠ Retired myflow review-panel vocabulary found:\n%s\n' "$hits" >&2
    printf 'The panel roster must match the rule file'"'"'s review-panel section.\n' >&2
    return 1
  fi

  echo "✓ Panel-vocabulary guard: clean"
  return 0
}

status=0
check_retired_stage_vocabulary || status=1
check_retired_panel_vocabulary || status=1
exit "$status"
