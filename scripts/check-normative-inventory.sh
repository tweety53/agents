#!/usr/bin/env bash
# check-normative-inventory.sh — print every normative sentence in this
# repository's owned Markdown, one per line, sorted.
#
# It exists because nothing else in `## lint` would notice a deleted
# requirement. check-references.sh and check-installed-citations.sh catch a
# citation that stops resolving; check-markdown-integrity.sh catches structural
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
# The corpus is ownedCorpus (stats/internal/guard/ownedcorpus.go), the one
# definition check-ask-silence walks too; the mechanism is
# stats/internal/guard/normativeinventory.go.
#
# The repository root is this shim's own parent, exported as
# FLOW_GUARD_REPO_ROOT — that is what makes the guard argument-free and
# self-scoped from any cwd. CHECK_NORMATIVE_INVENTORY_ROOT is an explicit,
# opt-in override honored only when set, for a sandboxed fixture tree — never
# set it for a normal invocation; set-but-empty is refused.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-normative-inventory: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec check-normative-inventory 2 "check-normative-inventory:" "$@"
