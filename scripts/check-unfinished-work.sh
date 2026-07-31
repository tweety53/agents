#!/usr/bin/env bash
# check-unfinished-work.sh — report whether a change carries unfinished work,
# for /myflow-finish's run-1 gate.
#
# Usage: check-unfinished-work.sh <worktree> <change-name>
#
# Prints ONE verdict line to stdout:
#   CLEAR: <reason>          nothing outstanding
#   OUTSTANDING: <breakdown> one or more signals fired
#
# Exit 0 whenever a verdict was reached; exit 2 when it cannot answer at all —
# an unreadable worktree, a change name outside the allowlist, or a file that
# exists but cannot be read. The VERDICT carries the answer, not the exit
# status — see check-finish-preflight.sh's header for why this repository
# separates them. Finish runs this once per worktree, so each verdict names the
# worktree it judged; a bare CLEAR would be unattributable across several.
#
# WHY THE VERDICT WORDS DIFFER FROM check-cleanup-complete.sh's. That guard says
# COMPLETE/LEFTOVER, this one says CLEAR/OUTSTANDING, and the two vocabularies
# are deliberately not shared. They answer OPPOSITE questions about absence:
# here a file that is missing counts AGAINST the change, because a missing
# record proves nothing about the work; there a missing artifact IS the answer,
# because the question is "is it gone?". A reader who carried the meaning of
# CLEAR across to COMPLETE would carry that inversion with it. Distinct words
# are what stop the two verdicts reading as one. check-finish-preflight.sh's
# third vocabulary (RUN1/RUN2/REFUSE) differs for a further reason again: it
# selects a procedure rather than reporting a state.
#
# A MISSING FILE COUNTS AS OUTSTANDING, NOT CLEAR. Treating an absent guide, an
# absent plan, an absent panel record or an absent "## Known incomplete" section
# as clearance is the failure this guard exists to prevent: a branch merged over
# unfinished work because nothing was written down. The same reasoning covers a
# "## Known incomplete" heading with nothing under it — the contract asks for
# the word `None.` or a list of items, and a bare heading is neither, so it
# clears nothing.
#
# All four signals are counted independently and reported together on the one
# line, so the operator sees the whole picture in one prompt rather than being
# sent back around the loop one signal at a time.
#
# WHY THE CHECKLIST PATTERN IS ANCHORED. `- [ ]` is matched only at the start of
# a line (leading whitespace allowed, for a nested item). Unanchored, it also
# matches a plan that merely QUOTES a checklist inside a fenced example, which
# this repository's own plans do — measured: 44 unanchored hits against 42
# anchored ones in openspec/changes/kan-17-finish-gate-jira-and-commit-hygiene/
# tasks.md, the two extra being `printf` lines in fenced shell. A guard that
# fires on every plan that documents a checkbox is a guard the operator learns
# to click past. Fence tracking was rejected as the fuller fix: it is real
# complexity, and what it would still catch — a fenced example whose line BEGINS
# with `- [ ]` — errs toward OUTSTANDING, which prompts the operator rather than
# clearing the gate silently.
#
# WHERE FENCE TRACKING IS PAID FOR, AND THE RULE THAT DECIDES IT. Exactly one
# scan below tracks fences: the `## Known incomplete` section finder. The rule
# is not "fences matter" but "fences are tracked wherever ignoring them fails
# OPEN". A fenced example containing that heading and the word `None.` would
# otherwise be read as the guide's own section and CLEAR the signal for a guide
# that never wrote one — silent clearance, the failure this guard exists to
# break. The checklist pattern above fails toward OUTSTANDING when a fenced
# example is misread, so it does not pay for the tracking; signal three below
# reads anchored marker lines and has no notion of a block at all.
#
# WHAT THE FENCE RULE ACTUALLY IS, stated because an earlier version of this
# comment claimed the tracking closed the hole while the code implemented a
# boolean toggle on "the line contains ``` or ~~~" — which closes it only for
# examples whose markers happen to pair up. Two shapes got through: a closing
# marker indented four spaces inside an unclosed fence, and a ```` fence
# documenting a ``` one. Either desynchronises the toggle, after which every
# line is read with the fence state inverted and the fenced example becomes the
# guide's own section.
#
# The rule below is CommonMark's, and opening and closing are separate tests. An
# opening marker is at most three spaces of indent then three or more of ` or ~;
# a backtick opener may not carry a backtick in its info string. A CLOSING marker
# must use the SAME character, be AT LEAST AS LONG as the opener, and carry
# nothing but whitespace after it. An unclosed fence runs to end of file. What
# this still does NOT model, deliberately: fences nested inside list items or
# blockquotes, where the container's indentation shifts the limit. Both of those
# fail toward OUTSTANDING here — the section is not seen — so they are the safe
# direction, which is why they are not modelled.
set -euo pipefail

# NO ANSWER THIS GUARD GIVES MAY DEPEND ON THE CALLER'S LOCALE. Be precise about
# what this line does and does not buy, because a comment claiming more than its
# code delivers is itself one of the defects this file was rewritten to remove.
#
# WHAT FIXED THE STATUS COMPARISON WAS THE REDESIGN, NOT THIS LINE. An earlier
# version compared a status with awk's `==`, which goes through the locale's
# collating sequence, so a zero-width character in it reported OUTSTANDING under
# a UTF-8 locale and vanished under `LC_ALL=C`. Signal three no longer compares
# strings at all: it matches anchored patterns, and a pattern match is not
# collation-sensitive in the way `==` is.
#
# WHAT THIS LINE DOES BUY, measured rather than assumed: `case`'s `A-Za-z0-9`
# are COLLATING ranges. On bash 3.2 the change name `démo` is rejected by the
# allowlist below under `LC_ALL=C` and ACCEPTED under `en_US.UTF-8` — so without
# this line the guard's accepted input set changes with the operator's
# environment, and this copy of preserve-session-records.sh's Protection 1
# diverges from that one under some locales and not others. It also pins
# `[[:space:]]`, `sort`'s ordering and `tr`'s ranges to bytes, which costs
# nothing and removes the question. test-check-unfinished-work.sh case 8e-i
# asserts the allowlist behaves identically under both locales.
export LC_ALL=C

WORKTREE="${1:-}"
NAME="${2:-}"

if [ -z "$WORKTREE" ] || [ -z "$NAME" ]; then
  echo "usage: check-unfinished-work.sh <worktree> <change-name>" >&2
  exit 2
fi

# CONTAINMENT: the change name arrives from a pull-request-editable state file
# and is concatenated into every path below. Without this check
# `../../../planted/clear` makes the gate read a file outside the worktree
# entirely and report CLEAR for a change that has no guide at all, and a glob
# metacharacter reaches the `case` patterns and the `find` filter further down.
#
# The rule is preserve-session-records.sh's Protection 1, character for
# character, and its comment there is canonical for why each hazard is in it —
# including the incident where `*` in a name "matched and overwrote a DIFFERENT
# change's preserved record". THE COPY IS DELIBERATE. These guards are
# single-file by design and are copied into projects one at a time (the finish
# contract has an explicit "when the script is absent" path), so a sourced
# helper would make a guard that is present but unrunnable — a worse failure
# than a duplicated six-line `case`. Both harnesses assert the same rejected
# shapes, which is what keeps the copies from drifting apart silently.
#
# It also closes the symlink question at these paths: `-f` and `-d` follow
# symlinks, so a name that cannot leave the change's own directory is what makes
# the paths this guard reads the ones it was asked about.
case "$NAME" in
  [!A-Za-z0-9]* | *[!A-Za-z0-9._-]*)
    echo "check-unfinished-work: change name '$NAME' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'" >&2
    exit 2
    ;;
esac

if [ ! -d "$WORKTREE" ]; then
  echo "check-unfinished-work: $WORKTREE is not a directory — cannot determine anything" >&2
  exit 2
fi

GUIDE="$WORKTREE/docs/manual-test/$NAME.md"
PANEL="$WORKTREE/.superpowers/sdd/final-review-panel.md"
CHANGES="$WORKTREE/openspec/changes"
PRIMARY_PLAN="$CHANGES/$NAME/tasks.md"

REASONS=""
add() { REASONS="${REASONS:+$REASONS; }$1"; }

# unreadable <path> — a file that exists but cannot be read is neither signal:
# it is an honest unknown, and this guard refuses rather than guessing. Reading
# it as zero unticked boxes would fail toward the reassuring answer, which is
# the one failure direction this whole guard exists to remove.
unreadable() {
  echo "check-unfinished-work: cannot read $1 — cannot determine anything" >&2
  exit 2
}

# count_matching <file> <ere> — how many lines of <file> match <ere>.
#
# `-a` ON EVERY GREP IN THIS FILE, and here is exactly what it is load-bearing
# for. A record carrying one stray NUL byte — a bad merge, a truncated write —
# puts grep into binary mode, where it suppresses output and reports "no match"
# whatever the file contains.
#
# For a COUNT that fails safe on its own: every count drops to zero, which makes
# the record look like one with no declared total, and a missing total is
# outstanding. For `ids_of` it does NOT: both identifier lists come back empty,
# empty equals empty, and the check that holds the findings table against the
# marker block passes vacuously. That is the reassuring direction, so it is the
# one this flag exists for; test-check-unfinished-work.sh case 4s-iii pins it
# with a record whose two lists genuinely disagree. The flag is on the other
# greps as defence in depth — the same input should not be read two ways by two
# lines of the same guard.
#
# `grep -c` exits 1 on "no match", which is an answer (`0`) and not a failure,
# and >= 2 on a real error — an unreadable file, a permission anomaly. A
# blanket `|| true` cannot tell those apart and turns the error into `0`, i.e.
# "nothing outstanding here". They are separated here: 0 and 1 return the count,
# anything else returns non-zero for the caller to refuse on. Every caller below
# refuses; that discipline is what keeps an unreadable file from reading as a
# clean one.
count_matching() {
  local out rc=0
  out="$(grep -acE "$2" "$1")" || rc=$?
  if [ "$rc" -gt 1 ]; then
    return "$rc"
  fi
  printf '%s\n' "${out:-0}"
}

# ids_of <file> <ere> — the digits of every match of <ere>, one per line,
# sorted, so two such lists can be compared for equality with `=`. Used to hold
# the findings table's row identifiers against the marker block's; see signal
# three. `grep -o` is fed the same exit-code discipline as count_matching.
ids_of() {
  local raw rc=0
  raw="$(grep -aoE "$2" "$1")" || rc=$?
  if [ "$rc" -gt 1 ]; then
    return "$rc"
  fi
  [ -n "$raw" ] || return 0
  printf '%s' "$raw" | tr -cd '0-9\n' | sort
}

# repeated_ids <sorted-id-list> — the identifiers appearing more than once in a
# list produced by ids_of, rendered `F<n>` and comma separated; empty when every
# identifier is unique. `uniq -d` is exact on a sorted list and names each repeat
# once however many times it occurs.
repeated_ids() {
  local out="" n
  [ -n "$1" ] || return 0
  while IFS= read -r n; do
    [ -n "$n" ] || continue
    out="${out:+$out, }F$n"
  done <<< "$(printf '%s\n' "$1" | uniq -d)"
  printf '%s' "$out"
}

# line_numbers <file> <ere> — the line numbers of the matching lines, one per
# line, in file order. The number is everything before grep's first colon, and a
# line number never contains one, so `cut -d:` is exact here.
line_numbers() {
  local raw rc=0
  raw="$(grep -anE "$2" "$1")" || rc=$?
  if [ "$rc" -gt 1 ]; then
    return "$rc"
  fi
  [ -n "$raw" ] || return 0
  printf '%s\n' "$raw" | cut -d: -f1
}

# count_unticked <file> — unticked checklist items in a markdown file. The
# pattern lives here alone: the guide and the plans are the same knowledge
# about what an unticked box looks like, and two copies would drift.
count_unticked() {
  count_matching "$1" '^[[:space:]]*- \[ \]'
}

# Signals one and four — the manual test guide's boxes, and the work it already
# knows is undone.
if [ ! -f "$GUIDE" ]; then
  add "no manual test guide at $GUIDE"
else
  BOXES="$(count_unticked "$GUIDE")" || unreadable "$GUIDE"
  if [ "${BOXES:-0}" -gt 0 ]; then
    add "$BOXES unticked box(es) in the manual test guide"
  fi

  # One pass answers all three questions about the section: is it there, does it
  # hold anything, and is anything in it other than `None.`. The section runs to
  # the next `## ` heading, so a later section's content is never read as
  # incomplete work; blank lines are dropped as they are collected.
  #
  # Fence-aware, per the header's rule. Content inside a fence that is inside the
  # section still counts as content — a fenced snippet under the heading is the
  # author saying something — but neither the heading nor the section terminator
  # is recognised inside one.
  SECTION_STATE="$(awk '
    # scan_fence(line) — 1 when <line> has the SHAPE of a fenced-code-block
    # marker, with FCHAR/FLEN/FREST set to its character, its length and the rest
    # of the line. CommonMark: at most three spaces of indent, then three or more
    # of ` or ~. Only literal spaces are counted as indent, so a leading tab —
    # four columns, past the limit — correctly reads as content, not a marker.
    function scan_fence(line,   i, n, ch, cnt) {
      i = 1; n = 0
      while (substr(line, i, 1) == " ") { n++; i++ }
      if (n > 3) return 0
      ch = substr(line, i, 1)
      if (ch != "`" && ch != "~") return 0
      cnt = 0
      while (substr(line, i, 1) == ch) { cnt++; i++ }
      if (cnt < 3) return 0
      FCHAR = ch; FLEN = cnt; FREST = substr(line, i)
      return 1
    }
    {
      # OPENING and CLOSING ARE DIFFERENT TESTS, which is the whole correction.
      # A boolean toggle over "this line looks like a fence" desynchronises on
      # the first marker that is not a real closer, and every line after it is
      # then read with the state inverted — which is how a guide whose only
      # `## Known incomplete` sits inside a fenced example reported CLEAR.
      #
      # A closer must use the SAME character, be AT LEAST AS LONG as the opener,
      # and carry nothing but whitespace after it. An opening backtick fence may
      # not carry a backtick in its info string; such a line opens nothing.
      # An unclosed fence runs to end of file, as CommonMark says.
      if (scan_fence($0)) {
        if (!in_fence) {
          if (FCHAR != "`" || index(FREST, "`") == 0) {
            in_fence = 1; fence_char = FCHAR; fence_len = FLEN
          }
        } else if (FCHAR == fence_char && FLEN >= fence_len && FREST ~ /^[[:space:]]*$/) {
          in_fence = 0
        }
      }
      if (!in_fence) {
        if ($0 ~ /^## Known incomplete/) { found = 1; in_section = 1; next }
        if (in_section && $0 ~ /^## /) { in_section = 0 }
      }
      if (in_section && $0 ~ /[^[:space:]]/) {
        lines++
        if ($0 !~ /^[[:space:]]*None\.[[:space:]]*$/) { remainder++ }
      }
    }
    END { printf "%d %d %d\n", found + 0, lines + 0, remainder + 0 }
  ' "$GUIDE")" || unreadable "$GUIDE"
  read -r KNOWN_FOUND KNOWN_LINES KNOWN_REMAINDER <<< "$SECTION_STATE"

  if [ "$KNOWN_FOUND" -eq 0 ]; then
    add "the manual test guide has no '## Known incomplete' section"
  elif [ "$KNOWN_LINES" -eq 0 ]; then
    add "the '## Known incomplete' section is empty — it must state None. or list what is undone"
  elif [ "$KNOWN_REMAINDER" -gt 0 ]; then
    # `None.` on its own line is the contract's way of saying "nothing".
    # Anything surviving that filter is a signal, including `None.` written
    # alongside items that contradict it.
    add "the guide records work known to be incomplete"
  fi
fi

# Signal two — the plan, including any fix sub-change.
#
# THE PRIMARY PLAN IS TRACKED SEPARATELY FROM THE REST, because their absences
# mean opposite things. `openspec/changes/<name>/tasks.md` is the change's plan
# and every myflow change has one: missing, it is outstanding like any other
# missing record, and reporting "every plan item is checked" over a change with
# no plan at all is the same silent clearance the header rejects. A fix
# sub-change's plan is genuinely optional — most changes have none — so its
# absence is nothing.
#
# THE THREE GLOB SHAPES THIS USED TO READ ARE GONE. They fixed the layouts a fix
# run may write (`<name>-fix-N` beside the change, or nested inside it) at one
# level of nesting each, and nothing in the repository validates that a fix run
# lands in one of them — so a fix-of-a-fix, or a third layout nobody has written
# yet, silently contributed no signal. Every `tasks.md` at any depth under the
# change's own directory, and under any sibling directory whose name starts
# `<name>-fix-`, is read instead: the discovery no longer depends on guessing
# the layout. What remains undiscoverable is a sub-change whose name has no
# relation to its parent's, which the contract does not permit.
PLAN_OPEN=0
if [ ! -f "$PRIMARY_PLAN" ]; then
  add "no plan at $PRIMARY_PLAN"
else
  N="$(count_unticked "$PRIMARY_PLAN")" || unreadable "$PRIMARY_PLAN"
  PLAN_OPEN=$((PLAN_OPEN + ${N:-0}))
fi

if [ -d "$CHANGES" ]; then
  while IFS= read -r -d '' plan; do
    [ "$plan" != "$PRIMARY_PLAN" ] || continue
    case "$plan" in
      "$CHANGES/$NAME/"* | "$CHANGES/$NAME"-fix-*) : ;;
      *) continue ;;
    esac
    N="$(count_unticked "$plan")" || unreadable "$plan"
    PLAN_OPEN=$((PLAN_OPEN + ${N:-0}))
  done < <(find "$CHANGES" -type f -name tasks.md -print0 2>/dev/null || true)
fi

if [ "$PLAN_OPEN" -gt 0 ]; then
  add "$PLAN_OPEN unchecked plan item(s)"
fi

# Signal three — findings whose recorded status is not closed.
#
# THIS SIGNAL DOES NOT PARSE THE FINDINGS TABLE, and the deletion is the point.
# Three review passes hid an open Critical from a hand-rolled GFM table parser
# six distinct ways: a capitalised `Open`; `open (needs discussion)`; an
# unescaped `|` in an earlier cell shifting the columns; a reordered header; a
# row missing its LEADING pipe silently ending table tracking and hiding every
# row below it; and a row both detached from the table and malformed, which each
# half of the parser left to the other. A table written inside a blockquote was
# invisible once a legitimate one existed elsewhere, and the status comparison
# went through the locale's collating sequence. Every fix closed the shape it
# was shown and left the class alive, because the guard was recovering one fact
# — is any finding still open — by re-implementing a document grammar defined in
# another file's prose.
#
# This repository has paid for that lesson once already: `.myflow/project.md`
# records check-plan-provenance's Bash classifier taking five review panel
# passes and seven fix waves before it was replaced with a real parser. And
# check-cleanup-complete.sh solves the analogous problem the right way in this
# same change — the emitting side writes machine-readable markers
# (`registry-row-checked:`) and the guard counts them.
#
# SO THE EMITTING SIDE WRITES THE FACT DOWN. `/myflow-do`'s panel writes, beside
# every human table row, one anchored line:
#
#   finding-status: F<n> open
#   finding-status: F<n> fixed
#   finding-status: F<n> withdrawn <reason>
#
# and one checksum line, `findings-total: <n>`. The format is a SHALL in
# openspec/specs/myflow-review-panel-economics/spec.md and is documented where
# the table is, in skills/myflow-do/SKILL.md. Because a marker is a whole line
# there are no cells to split, no escaping rule, no table boundary to track and
# no orphan net: none of the six defects has anywhere to happen.
#
# WHAT "CANNOT CONFIDENTLY ANSWER" MEANS HERE, exhaustively. Each is
# outstanding, because the one direction this guard may never fail in is the
# reassuring one:
#
#   1. NO DECLARED TOTAL. A record that never states how many findings it
#      carries has not spoken, and zero findings cannot be read out of silence.
#      A panel that genuinely raised none says so — `findings-total: 0` — which
#      is a positive declaration and clears.
#   2. A TOTAL DECLARED TWICE, or one that is not a plain count. An ambiguous
#      checksum is no checksum.
#   3. THE TOTAL DISAGREEING WITH THE MARKER COUNT. This is the drift that
#      matters most: a fix round that adds a row and forgets its marker leaves
#      the count short, and the guard says so with both numbers.
#   4. THE TABLE AND THE MARKER BLOCK NAMING DIFFERENT FINDINGS. The row
#      identifiers and the marker identifiers are compared as sorted lists, so a
#      row with no marker, a marker with no row, and a duplicated identifier —
#      which counts alone would balance — are all reported.
#   5. A MARKER WHOSE STATUS IS NOT ONE OF THE THREE LEGAL VALUES. Compared
#      byte for byte: `Open`, `WITHDRAWN`, `open (needs discussion)` and a status
#      carrying a zero-width character are none of them, and none of them falls
#      through to "not open" the way the old table's status cell could.
#   6. A LINE THAT NAMES `finding-status:` WITHOUT BEING A MARKER — indented,
#      inside a blockquote, or missing its identifier. Counting only anchored
#      lines would drop it silently; the difference between the two counts is
#      reported instead.
#   7. A `withdrawn` MARKER WITH NO REASON. Pass 1's rule, re-expressed: the
#      reason is part of the state, so it lives on the marker line with it.
#   8. THE FILE UNREADABLE, which refuses outright rather than answering.
#
# ROW IDENTIFIERS ARE READ WITH AN ANCHORED PREFIX, NOT A PARSE. The row test is
# "does this line begin with an optional pipe and then `F<digits> |`" — it never
# looks at a second cell, a header, or where a table starts or ends, so none of
# the six defects reaches it. Every way of getting a row wrong that the prefix
# does not match — a blockquoted table, an indented one — makes the two lists
# differ, which is OUTSTANDING. The coupling to the table's shape is one
# anchored pattern, and it fails in the safe direction.
if [ ! -f "$PANEL" ]; then
  add "no review panel record at $PANEL"
else
  M_ANCHORED="$(count_matching "$PANEL" '^finding-status: F[0-9]+ ')" || unreadable "$PANEL"
  M_NAMED="$(count_matching "$PANEL" 'finding-status:')" || unreadable "$PANEL"
  M_OPEN="$(count_matching "$PANEL" '^finding-status: F[0-9]+ open[[:space:]]*$')" || unreadable "$PANEL"
  M_FIXED="$(count_matching "$PANEL" '^finding-status: F[0-9]+ fixed[[:space:]]*$')" || unreadable "$PANEL"
  M_WITHDRAWN="$(count_matching "$PANEL" '^finding-status: F[0-9]+ withdrawn[[:space:]]+[^[:space:]]')" || unreadable "$PANEL"
  M_NOREASON="$(count_matching "$PANEL" '^finding-status: F[0-9]+ withdrawn[[:space:]]*$')" || unreadable "$PANEL"
  T_NAMED="$(count_matching "$PANEL" 'findings-total:')" || unreadable "$PANEL"
  T_WELLFORMED="$(count_matching "$PANEL" '^findings-total: (0|[1-9][0-9]*)[[:space:]]*$')" || unreadable "$PANEL"

  # The four status patterns are mutually exclusive by construction — a marker
  # ends in `open`, ends in `fixed`, has a non-blank tail after `withdrawn`, or
  # has a blank one — so what is left over is exactly the unrecognised markers.
  # Leading zeros are excluded from the total so that `010` cannot be read as
  # eight by the shell's arithmetic below.
  M_UNRECOGNISED=$((M_ANCHORED - M_OPEN - M_FIXED - M_WITHDRAWN - M_NOREASON))
  M_MALFORMED=$((M_NAMED - M_ANCHORED))

  if [ "$T_NAMED" -eq 0 ]; then
    add "the review panel record does not declare a findings total — it must carry one 'findings-total: <n>' line, and a record that never says how many findings it has cannot show zero"
  elif [ "$T_NAMED" -gt 1 ]; then
    add "the review panel record declares a findings total more than once"
  elif [ "$T_WELLFORMED" -ne 1 ]; then
    add "the review panel record's findings total is not a plain count — it must read 'findings-total: <n>'"
  else
    TOTAL_LINE="$(grep -am1 -E '^findings-total: (0|[1-9][0-9]*)[[:space:]]*$' "$PANEL")" || unreadable "$PANEL"
    # THE COUNT IS TAKEN WITH PARAMETER EXPANSION, and that is the fix. `read -r
    # _ n` splits on IFS, which does not include `\r`, so on a CRLF record the
    # count arrived as `1<CR>` and `[ … -ne … ]` exited 2 with "integer
    # expression expected" — inside an `if`, which errexit does not see, so the
    # branch was skipped and THE CHECKSUM SILENTLY STOPPED RUNNING. Stripping
    # every trailing non-digit removes that whole class, whatever the record's
    # line endings are.
    #
    # The comparison is then `!=` rather than `-ne` as defence in depth, not as
    # the fix: both counts are plain digit strings with no leading zero (the
    # pattern above excludes one, `grep -c` never writes one), so the two are
    # equivalent here — but `!=` cannot fail in a way a conditional swallows,
    # which is the failure mode that let the skip go unnoticed.
    DECLARED_TOTAL="${TOTAL_LINE#findings-total: }"
    DECLARED_TOTAL="${DECLARED_TOTAL%%[![:digit:]]*}"
    case "$DECLARED_TOTAL" in
      '')
        # Unreachable through the pattern above, and reported rather than
        # assumed away: a checksum that could not be read is not a checksum
        # that agreed, and "it cannot happen" is how the last silent skip got in.
        add "the review panel record's findings total could not be read"
        ;;
      *)
        if [ "$DECLARED_TOTAL" != "$M_ANCHORED" ]; then
          add "the review panel record declares findings-total: $DECLARED_TOTAL but carries $M_ANCHORED finding-status: marker line(s)"
        fi
        ;;
    esac
  fi

  # THE MARKER LINES MUST BE ONE UNBROKEN BLOCK, and this rule is here because
  # attacking the redesign found the one route that still under-counted. A marker
  # line quoted inside a fenced example is counted like any other — signal three
  # has no notion of a block, which is the whole point — so a fenced
  # `finding-status: F2 fixed` sitting elsewhere in the record STOOD IN FOR F2's
  # real marker when that marker was never written: identifiers matched, the
  # checksum matched, and an open finding read as clear.
  #
  # Requiring the markers to occupy consecutive lines closes it without teaching
  # this signal about fences, which is the block tracking the redesign deleted:
  # the fence's own delimiter line falls between the quoted marker and the real
  # block and breaks the run. The rule also matches the format as documented —
  # one marker block, one line per row — so it costs a correct record nothing.
  if [ "$M_ANCHORED" -gt 0 ]; then
    M_AT="$(line_numbers "$PANEL" '^finding-status: F[0-9]+ ')" || unreadable "$PANEL"
    M_SPAN=$(( ${M_AT##*$'\n'} - ${M_AT%%$'\n'*} + 1 ))
    if [ "$M_SPAN" -ne "$M_ANCHORED" ]; then
      add "the review panel record's $M_ANCHORED marker line(s) are spread over $M_SPAN lines — they must be one unbroken block, so that a marker written anywhere else cannot stand in for a missing one"
    fi
  fi

  # The row identifiers against the marker identifiers, as sorted lists rather
  # than as counts, so that a duplicated identifier cannot balance a missing one.
  ROW_IDS="$(ids_of "$PANEL" '^\|?[[:space:]]*F[0-9]+[[:space:]]*\|')" || unreadable "$PANEL"
  MARKER_IDS="$(ids_of "$PANEL" '^finding-status: F[0-9]+ ')" || unreadable "$PANEL"
  if [ "$ROW_IDS" != "$MARKER_IDS" ]; then
    add "the findings table and the marker block do not name the same findings — every row's F<n> needs exactly one 'finding-status: F<n> …' line and the reverse"
  fi

  # UNIQUENESS IS ASSERTED ON EACH SIDE SEPARATELY, because the comparison above
  # holds two sorted MULTISETS against each other and a duplicate that appears on
  # both sides cancels out. Two genuinely distinct findings both labelled `F1`,
  # with `F1` written twice in the marker block, matched perfectly and cleared —
  # an open Critical hidden behind a reused identifier, with the word `open`
  # never appearing in a marker, so the open count could not help either. The
  # spec already says an identifier is unique within the record; this is that
  # SHALL made mechanical, on each side independently so that neither can be
  # excused by the other carrying the same repeat.
  ROW_DUPES="$(repeated_ids "$ROW_IDS")"
  MARKER_DUPES="$(repeated_ids "$MARKER_IDS")"
  if [ -n "$ROW_DUPES" ]; then
    add "the findings table reuses identifier(s) $ROW_DUPES — each F<n> must name exactly one finding"
  fi
  if [ -n "$MARKER_DUPES" ]; then
    add "the marker block reuses identifier(s) $MARKER_DUPES — each F<n> must have exactly one marker line"
  fi

  if [ "$M_OPEN" -gt 0 ]; then
    add "$M_OPEN open finding(s) in the review panel record"
  fi
  if [ "$M_UNRECOGNISED" -gt 0 ]; then
    add "$M_UNRECOGNISED finding marker(s) whose status is none of open, fixed or withdrawn — counted as outstanding"
  fi
  if [ "$M_NOREASON" -gt 0 ]; then
    add "$M_NOREASON withdrawn finding(s) with no reason after the status"
  fi
  if [ "$M_MALFORMED" -gt 0 ]; then
    add "$M_MALFORMED line(s) naming finding-status: that are not marker lines — a marker must begin its line, as 'finding-status: F<n> <status>'"
  fi
fi

if [ -z "$REASONS" ]; then
  echo "CLEAR: $WORKTREE — every checklist box is ticked, every plan item is checked, no finding is open, and nothing is recorded as incomplete"
else
  echo "OUTSTANDING: $WORKTREE — $REASONS"
fi
exit 0
