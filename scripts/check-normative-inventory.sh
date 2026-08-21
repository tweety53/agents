#!/usr/bin/env bash
# check-normative-inventory.sh — print every normative sentence in this
# repository's owned Markdown, one per line, sorted.
#
# It exists because nothing else in `## lint` would notice a deleted
# requirement. check-references.sh and check-installed-citations.sh catch a
# citation that stops resolving; check-markdown-integrity.py catches structural
# damage; the harnesses under `## test` exercise script behaviour. A sentence
# carrying a SHALL can be deleted with every one of them green — which is the
# failure mode a corpus-wide prose trim risks, and the reason this guard is a
# precondition of one. A change that edits prose in bulk captures this output
# before its first edit and after its last, and requires the two to be
# byte-identical; a difference is resolved by restoring the sentence, never by
# accepting the new inventory.
#
# THIS GUARD REPORTS A SET. It has no violation exit code, because comparing two
# sets is the caller's act — `diff <(scripts/check-normative-inventory.sh)
# normative-baseline.txt` is the whole check. Exit codes: 0 the inventory
# printed, 2 it cannot answer at all (a scope root missing or unreadable, a file
# it cannot read, an override set but empty).
#
# A NORMATIVE SENTENCE is one containing `SHALL`, `SHALL NOT`, `MUST` or
# `MUST NOT` as a whole word. Nothing else is a keyword: `SHOULD` and `MAY`
# state no obligation whose loss changes behaviour, and including them would
# make the inventory churn on ordinary editing. The match is case-sensitive and
# word-bounded, so `MUST_NOT_TOUCH` and `shall` are not keywords.
#
# THE SENTENCE BOUNDARY RULE, in full, because the guard's whole value rests on
# it being stable:
#
#   1. BLOCKS. A block is a heading line, one table cell, one list item, one
#      line inside a fenced code block, or a run of consecutive non-blank prose
#      lines. Each source line is stripped of leading and trailing whitespace and
#      of any blockquote (`>`) markers before it is classified, and the lines of
#      a prose run are joined with a single space. Joining is what makes the
#      output REFLOW-INVARIANT: the same wording rewrapped across different line
#      boundaries, at any indentation, yields the same block and therefore the
#      same output. A blank line, a heading, a list marker, a table row, a
#      thematic break or a fence ends the block being accumulated.
#   2. SENTENCES. A block is split at `.`, `!` or `?` followed by whitespace,
#      with any closing quote, bracket, backtick or emphasis characters riding
#      with the terminator (`... X.**` and `... X.)` end there too). The block's
#      tail, terminated or not, is the last sentence.
#   3. OUTPUT. A sentence is printed when it carries a keyword. Its internal
#      whitespace is collapsed to single spaces and its ends stripped; NOTHING
#      ELSE is normalised, because case, punctuation and wording are the content
#      being protected. Lines are sorted with LC_ALL=C, so the output is
#      `diff`-comparable between runs and independent of file order. Duplicates
#      are KEPT: the same sentence restated in two files must be printed twice,
#      or deleting one of the two copies would leave the inventory unchanged.
#
# ABBREVIATIONS are exempt from rule 2. A period closing `e.g.`, `i.e.`, `cf.`,
# `vs.`, `etc.`, `al.`, `Dr.` or `No.` — matched case-insensitively on the token
# before it — is not a terminator, and the sentence carries on past it. Without
# that exemption the splitter truncates a requirement at its first abbreviation
# and inventories the stub alone, which is worse than missing the sentence
# outright: a stub is a STABLE line that keeps matching while the obligation
# behind it is replaced wholesale, and counting keyword occurrences cannot see
# the swap either, because the keyword never moves.
#
# WHAT REMAINS, stated as what it is rather than as a coverage gap:
#
#   * An abbreviation OUTSIDE the list above ("Fig.", "approx.", "Prof.") still
#     ends a sentence early, and a substantive change to the text after it can
#     hide behind the stub — the defect described above, unfixed for exactly
#     those spellings. The corpus carries none of them today; the fix for a new
#     one is a word added to the list, in this file, visible in its diff.
#   * A sentence that genuinely ENDS in a listed abbreviation is merged with the
#     sentence after it. That direction is safe: both texts remain in the line,
#     so no change to either can hide. The cost is a coarser diff, never a
#     missed edit.
#
# Fenced content is scanned line by line rather than joined, because code is not
# reflowable prose; a worked example carrying a SHALL is therefore inventoried,
# which is correct, since a worked example is one of the things a trim may never
# cut.
#
# The corpus is resolved by scripts/lib/owned-corpus.sh, which is also what
# check-contract-budget.sh calls once its own widening lands — one implementation,
# so the two guards cannot disagree about which files this repository owns.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# REPO_ROOT is resolved from this script's own location, exactly like
# check-vocabulary.sh and check-contract-budget.sh. That is what makes the guard
# argument-free and self-scoped from any cwd. This is repo lint, not a shipped
# guard: it is symlinked into no skill's scripts/ directory, so
# check-guard-symlinks.sh's rule 4 — which forbids a SHIPPED guard from deriving
# a root as $SCRIPT_DIR/.. — does not reach it, by the same reading that leaves
# check-vocabulary.sh alone.
#
# CHECK_NORMATIVE_INVENTORY_ROOT is an explicit, opt-in override honored only
# when set. It exists solely so the companion harness can point the guard at a
# sandboxed fixture tree — never set it for a normal invocation. Set-but-empty is
# refused rather than treated as unset: falling back there would make the harness
# scan the real repository while believing it scanned a fixture.
if [ -n "${CHECK_NORMATIVE_INVENTORY_ROOT:-}" ]; then
  REPO_ROOT="$CHECK_NORMATIVE_INVENTORY_ROOT"
elif [ "${CHECK_NORMATIVE_INVENTORY_ROOT+set}" = "set" ]; then
  printf 'CHECK_NORMATIVE_INVENTORY_ROOT is set but empty\n' >&2
  exit 2
else
  REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
fi

[ -d "$REPO_ROOT" ] || { printf 'not a directory: %s\n' "$REPO_ROOT" >&2; exit 2; }

# shellcheck source=lib/owned-corpus.sh
source "$SCRIPT_DIR/lib/owned-corpus.sh"

WORK="$(mktemp -d)" || { printf 'cannot create a temporary directory\n' >&2; exit 2; }
trap 'rm -rf "$WORK"' EXIT

owned_corpus_files "$REPO_ROOT" > "$WORK/files" || exit 2

files=()
while IFS= read -r -d '' path; do
  files+=("$path")
done < "$WORK/files"

# An empty corpus is refused, not reported as an empty inventory. The two are
# indistinguishable on stdout, and one of them is "every normative sentence in
# this repository has just been deleted" — the exact comparison this guard feeds.
if [ "${#files[@]}" -eq 0 ]; then
  printf 'no .md or .mdc file found under %s — refusing to report an empty inventory\n' \
    "$REPO_ROOT" >&2
  exit 2
fi

read -r -d '' AWK_INVENTORY <<'AWK' || true
function has_keyword(s) {
  return (s ~ /(^|[^A-Za-z0-9_])(SHALL|MUST)([^A-Za-z0-9_]|$)/)
}

# ABBREVIATIONS whose trailing period does NOT end a sentence. Matched on the
# last whitespace-separated token before the period, lower-cased, with any
# leading punctuation ("(e.g.") stripped first.
#
# Without this the splitter truncates a requirement at its first abbreviation
# and inventories only the stub. That is worse than missing the sentence
# outright: the stub is a STABLE line that keeps matching while the obligation
# behind it is replaced wholesale — "The tool SHALL validate e.g. input for XSS
# and SQL injection before use." and "The tool SHALL validate e.g. nothing at
# all here whatsoever." both reduce to "The tool SHALL validate e.g." and the
# inventory does not move. A keyword-occurrence reconciliation cannot see it
# either, because the keyword never moves.
function is_abbrev(text,   w, n, parts, last) {
  w = text
  sub(/ +$/, "", w)
  if (w !~ /\.$/) return 0
  n = split(w, parts, " ")
  last = parts[n]
  sub(/^[^A-Za-z0-9]+/, "", last)
  return (index(" e.g. i.e. cf. vs. etc. al. dr. no. ", " " tolower(last) " ") > 0)
}

# emit(block) — normalise one block, split it into sentences, print the
# normative ones.
function emit(s,   rest, cut, candidate, sentence, pending) {
  gsub(/[ \t]+/, " ", s)
  sub(/^ +/, "", s)
  sub(/ +$/, "", s)
  if (s == "") return
  rest = s
  pending = ""
  while (match(rest, /[.!?][]"'`*_)]* +/)) {
    cut = RSTART + RLENGTH - 1
    candidate = pending substr(rest, 1, cut)
    rest = substr(rest, cut + 1)
    # An abbreviation's period is not a boundary: carry the text forward and
    # keep looking for the real terminator.
    if (is_abbrev(candidate)) {
      pending = candidate
      continue
    }
    sentence = candidate
    sub(/ +$/, "", sentence)
    if (has_keyword(sentence)) print sentence
    pending = ""
  }
  rest = pending rest
  if (has_keyword(rest)) print rest
}

function flush() {
  if (block != "") { emit(block); block = "" }
}

FNR == 1 { flush(); in_fence = 0 }

{
  line = $0
  sub(/\r$/, "", line)
  sub(/^[ \t]+/, "", line)
  sub(/[ \t]+$/, "", line)

  # A fence line ends the block it interrupts and toggles fenced mode. Inside a
  # fence every line is its own block: code is not reflowable prose, so joining
  # its lines would be a fiction, and dropping them would lose a worked example.
  if (line ~ /^(```|~~~)/) { flush(); in_fence = (in_fence ? 0 : 1); next }
  if (in_fence) { emit(line); next }

  # Blockquote markers are stripped before classification, so a quoted heading,
  # bullet or paragraph is read exactly as an unquoted one. rules/*.mdc carry
  # normative text inside block quotes; without this they would be one
  # unreflowable line each.
  while (line ~ /^>/) {
    sub(/^>[ \t]?/, "", line)
    sub(/^[ \t]+/, "", line)
  }
  sub(/[ \t]+$/, "", line)

  if (line == "") { flush(); next }
  if (line ~ /^(---+|___+|\*\*\*+|===+)$/) { flush(); next }
  if (line ~ /^#+ /) { flush(); emit(line); next }

  # A table row is split on the cell separator: each cell is an independent
  # unit, so editing one cell cannot perturb the inventory line another cell
  # produces. agents-baseline.md's rule table is one requirement per cell.
  if (line ~ /^\|/) {
    flush()
    n = split(line, cells, "|")
    for (i = 1; i <= n; i++) emit(cells[i])
    next
  }

  if (line ~ /^([-*+]|[0-9]+[.)])[ \t]+/) {
    flush()
    sub(/^([-*+]|[0-9]+[.)])[ \t]+/, "", line)
    block = line
    next
  }

  block = (block == "" ? line : block " " line)
}

END { flush() }
AWK

: > "$WORK/sentences"
for path in "${files[@]}"; do
  if [ ! -r "$path" ]; then
    printf 'cannot read %s\n' "$path" >&2
    exit 2
  fi
  if ! awk "$AWK_INVENTORY" < "$path" >> "$WORK/sentences"; then
    printf 'cannot scan %s\n' "$path" >&2
    exit 2
  fi
done

# Every remaining step is guarded so its own status can never escape as this
# guard's. The contract allows 0 and 2 and nothing else, and an unguarded `sort`
# under `set -e` would exit with sort's status — 2 by coincidence today, whatever
# a future sort chooses tomorrow — which is an exit code this guard never decided
# to return.
if ! count="$(wc -l < "$WORK/sentences")"; then
  printf 'cannot count the inventory\n' >&2
  exit 2
fi
if ! count="$(printf '%s' "$count" | tr -d '[:space:]')"; then
  printf 'cannot normalise the inventory count\n' >&2
  exit 2
fi

if ! LC_ALL=C sort "$WORK/sentences"; then
  printf 'cannot sort the inventory\n' >&2
  exit 2
fi

# The counts go to STDERR, never stdout: stdout is the set the caller diffs, and
# a verdict line in it would be a difference of its own every time a sentence was
# legitimately added elsewhere. They are printed at all for the reason
# check-contract-budget.sh prints its count — a silent run and a run that scanned
# nothing look identical otherwise.
if ! printf 'check-normative-inventory: %s normative sentence(s) from %s file(s) under %s\n' \
  "$count" "${#files[@]}" "$REPO_ROOT" >&2; then
  exit 2
fi
