#!/usr/bin/env bash
# check-visual-verification.sh — validate a project's `## visual
# verification` declaration against the contract that defines it.
#
# Usage: check-visual-verification.sh <project root>
#
# ONE ARGUMENT, THE PROJECT ROOT — never the configuration file itself,
# matching check-workspace-isolation.sh's own convention: every caller
# already has the root, not the path underneath it.
#
# A project with no `.flow/project.md`, or one that carries it but declares
# no `## visual verification` section, is CLEAN, not a failure — the section
# is optional per design.md's "The contract", and a guard that failed on its
# absence would make every project in the world declare one.
#
# Prints one violation line per finding as `file:line: <message>`, then ONE
# verdict line:
#   VISUAL-OK:      <path> — <what was checked, or why nothing was>
#   VISUAL-INVALID: <path> — <n> violation(s) in the `## visual verification`
#                    section
#
# Exit 0 when clean, 1 when any violation was found, 2 when the guard cannot
# answer at all: the project root is not a directory, `.flow/project.md`
# exists but cannot be read or is not a regular file, the heading scan itself
# failed to look, or `git remote get-url origin` against a declared
# `regression checkout` failed for a reason other than "no such remote" (that
# one IS a finding — see below). A `2` is never a silent skip: a guard
# pointed at a moved file that reports `VISUAL-OK` is the vacuous pass
# scripts/check-vocabulary.sh's own header warns about.
#
# THE TWO ON-DISK TABLES, AND WHERE THEIR SHAPE COMES FROM. design.md's "The
# contract" section documents the settings and commands as
# `Setting | Required | Meaning` and `Command | Required | Meaning` — that is
# documentation prose, describing what each row MEANS, not the literal
# heading text of the table a project writes. The literal on-disk shape is
# established from two other places instead: task 8 of tasks.md's worked
# example for Gymie writes the settings table as `| Setting | Value |`, and
# `## workspace isolation` (skills/flow-contracts/project-configuration.md,
# `check-workspace-isolation.sh`) already establishes `| Command | Runs |` as
# this repository's one convention for a table of commands a project
# declares — reused here rather than invented a second time, so a project
# author who has already written one `## workspace isolation` section is not
# asked to learn a second table shape for `## capture`. Task 2, not yet
# written when this guard was, is canonical for the section's prose once it
# lands; this guard's own test harness pins the two headers it actually
# parses, so a disagreement between that landing and this file fails loudly
# rather than silently.
#
# WHAT IS CANONICAL, AND WHAT THIS FILE IS, for both checks below — matching
# check-workspace-isolation.sh's own citation of the identical shape for its
# `Resource` word. **Project configuration**
# (`skills/flow-contracts/project-configuration.md`), under its "## visual
# verification" heading, states the `Setting` and `Command` vocabularies as
# closed, and states that `regression repo` is an identity assertion, not an
# authorisation, checked whenever `regression checkout` and `regression
# repo` are both declared — design.md's `no-automatic-push` decision, which
# superseded an earlier design where the same check ran only behind a
# `push to default branch: allowed` row this contract no longer carries: a
# panel slot demonstrated that both fields of that row lived in this same
# pull-request-editable file and so authorised nothing. This file restates
# neither decision's reasoning; it cites the paragraph and enforces the rule.
#
# `git remote get-url origin` FAILING FOR "NO SUCH REMOTE" IS A FINDING, NOT
# A `2`. A checkout that is a real, valid git repository but carries no
# `origin` remote at all cannot honour the identity assertion either — that
# is a fact about the declaration, not a failure to look, so it is reported
# and the run continues. Any OTHER failure of that command (measured against
# git 2.50.1: none reproduces once `git rev-parse --git-dir` has already
# accepted the checkout as valid, which is why this guard's own harness
# reaches this branch through a stubbed `git` rather than a real one) is read
# as "cannot determine whether the identity assertion holds" and escalates
# the whole run to exit 2, matching this guard's own "never fail open" rule
# for every other unreadable input below.
#
# THE INPUT IS ATTACKER-INFLUENCED, the same fact check-workspace-isolation.sh
# is written against: `.flow/project.md` is tracked and editable in any pull
# request. NOTHING read here is executed — this guard never runs `setup`,
# `verify`, `capture` or `fingerprint`, and never interpolates a cell into a
# shell. All of
# the table parsing happens inside one awk program whose only input is the
# file's text, and every violation line the awk program prints, plus every
# cell this guard interpolates into a message afterward, passes through
# sanitize_display before it reaches stdout — see that function below for
# why escaping and not stripping or refusing.
#
# THE PARSER'S split_cells/trimcell/foldcell TRIO IS A SOURCED HELPER —
# scripts/lib/visual-table-cells.awk, not check-workspace-isolation.sh's own
# (wider, four-column) copy of the same shape. Its hazards are the same
# because its input is: a `|` inside a cell, written `\|`, must not split
# the row; a row may omit its trailing `|`; a file may arrive with CRLF line
# endings. See that lib file's header for why this guard, reached only
# through the skills/*/scripts/ symlink farm, may source a sibling instead
# of carrying its own copy — check-workspace-isolation.sh's own copy stays
# put, for the reason recorded there.
#
# A LEADING UTF-8 BOM IS STRIPPED BEFORE ANY OF THE ABOVE READS THE FILE, via
# the sourced strip_bom_cat (scripts/lib/strip-bom.sh) — see that file's
# header for why a BOM would otherwise make a present section read as
# absent, and for why this fix lives in scripts/lib/ rather than inline
# here: it arrived after check-visual-trigger.sh and
# resolve-visual-screenshots.sh already existed, sharing this same
# heading-parse convention, and stayed a single guard's own logic for one
# commit before moving here.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ "$#" -ne 1 ]; then
  echo "check-visual-verification: usage: check-visual-verification.sh <project root>" >&2
  exit 2
fi
ROOT="$1"

if [ ! -d "$ROOT" ]; then
  echo "check-visual-verification: $ROOT is not a directory — cannot tell whether it declares a visual verification section" >&2
  exit 2
fi

CFG="$ROOT/.flow/project.md"

# sanitize_display — sourced from lib/sanitize-display.sh; see that file's
# header for why this guard, unlike a hand-copied one, may source a sibling
# instead of carrying its own copy.
source "$SCRIPT_DIR/lib/sanitize-display.sh"

# git_clean — sourced from lib/git-clean.sh; see that file's header for the
# `GIT_DIR`-over-`-C` hazard it neutralizes and why this guard may source a
# sibling instead of carrying its own copy.
source "$SCRIPT_DIR/lib/git-clean.sh"

# strip_bom_cat — sourced from lib/strip-bom.sh; see that file's header for
# why a leading UTF-8 BOM would otherwise make a present `## visual
# verification` heading read as absent, and why this guard, unlike a
# hand-copied one, may source a sibling instead of carrying its own copy.
source "$SCRIPT_DIR/lib/strip-bom.sh"

# trim_glob_element — sourced from lib/trim-glob-element.sh; see that file's
# header for the split-then-strip parse it applies (THE SAME PARSE
# check-visual-trigger.sh applies to `ui paths` after splitting on comma),
# the single-pass strip it performs instead of a copy-per-character loop,
# and why this guard, unlike a hand-copied one, may source a sibling instead
# of carrying its own copy.
source "$SCRIPT_DIR/lib/trim-glob-element.sh"

# The file is optional. No `.flow/project.md` at all is a supported,
# ordinary case, not an error — matching check-workspace-isolation.sh's own
# treatment of the identical absence. `-L` catches a dangling symlink, which
# is not "nothing here" either.
if [ ! -e "$CFG" ] && [ ! -L "$CFG" ]; then
  printf 'VISUAL-OK: %s — no .flow/project.md, so nothing is declared\n' "$ROOT"
  exit 0
fi

# A REGULAR FILE, TESTED BEFORE READABILITY, because `-r` answers about a
# DIRECTORY too, and `grep` then fails with "Is a directory" while its exit
# status is indistinguishable from "no match" — the false pass
# check-workspace-isolation.sh's own comment documents at the same test.
if [ ! -f "$CFG" ]; then
  echo "check-visual-verification: $CFG is not a regular file — cannot validate what it declares" >&2
  exit 2
fi
if [ ! -r "$CFG" ]; then
  echo "check-visual-verification: $CFG exists but is not readable — cannot validate what it declares" >&2
  exit 2
fi

VV_HEADING='^##[[:space:]]+visual verification[[:space:]]*$'

# GREP HAS THREE ANSWERS, READ AS THREE: 0 matched, 1 did not match, 2 or
# more could not look. `if ! grep -q` would collapse the last two into one
# branch, the same fail-open the `-f` test above closes from a different
# angle. Every read below runs against strip_bom_cat's output, never the raw
# file, so a leading UTF-8 BOM never makes a present heading read as absent —
# see lib/strip-bom.sh's header for why.
set +e
grep -qiE "$VV_HEADING" <(strip_bom_cat "$CFG")
VV_GREP_RC=$?
set -e
if [ "$VV_GREP_RC" -ge 2 ]; then
  echo "check-visual-verification: grep exited $VV_GREP_RC while looking for the '## visual verification' heading in $CFG — that is a failure to look rather than an absence, so nothing was validated" >&2
  exit 2
fi
if [ "$VV_GREP_RC" -ne 0 ]; then
  printf 'VISUAL-OK: %s — declares no `## visual verification` section\n' "$CFG"
  exit 0
fi

# TWO DECLARATIONS ARE NOT A DECLARATION. This file is hand-written, so a
# heading duplicated by a bad merge or a copied block is a realistic
# mistake, and the two sections could name two different regression repos.
# Neither is validated; validating the first would be this guard picking the
# winner the contract declines to pick.
set +e
VV_COUNT="$(grep -ciE "$VV_HEADING" <(strip_bom_cat "$CFG"))"
VV_COUNT_RC=$?
set -e
case "$VV_COUNT_RC:$VV_COUNT" in
  0:[0-9]*) ;;
  *)
    echo "check-visual-verification: could not count the '## visual verification' headings in $CFG (grep exited $VV_COUNT_RC) — nothing was validated" >&2
    exit 2
    ;;
esac
if [ "$VV_COUNT" -gt 1 ]; then
  printf '%s:0: the file declares %s `## visual verification` sections — a second declaration is an ambiguous declaration rather than a merge, so neither was validated\n' "$CFG" "$VV_COUNT" | sanitize_display
  printf 'VISUAL-INVALID: %s — 1 violation(s) in the `## visual verification` section\n' "$CFG"
  exit 1
fi

# The whole of the table parsing. awk prints:
#   #VIOL\t<line>\t<message>   one per structural or vocabulary finding
#   #SET\t<line>\t<key>\t<value>   one per recognised, non-duplicate setting
#   #CMD\t<line>\t<key>\t<value>   one per recognised, non-duplicate command
#   #SUMMARY\t<heading line>   printed unconditionally last
#
# AN APOSTROPHE MAY NOW APPEAR BELOW — the program is a heredoc with a
# quoted terminator (`<<'AWK_PROG'`), not a single-quoted shell string, so an
# apostrophe inside no longer ends it mid-word the way it would in
# check-workspace-isolation.sh's own inline program. Kept quoted rather than
# unquoted regardless, so `$heading_re`-shaped text inside a comment is never
# read as shell expansion.
REPORT="$(awk -v cfg="$CFG" -v heading_re="$VV_HEADING" \
  -f "$SCRIPT_DIR/lib/visual-table-cells.awk" \
  -f <(cat <<'AWK_PROG'
  function is_delimiter(n, cells,   i, c) {
    for (i = 1; i <= n; i++) {
      c = trimcell(cells[i])
      if (c !~ /^:?-+:?$/) return 0
    }
    return n > 0
  }

  function violation(lineno, msg) {
    printf "#VIOL\t%d\t%s\n", lineno, msg
    nviol++
  }

  BEGIN { in_sec = 0; in_blk = 0; nviol = 0; heading_line = 0; SEC_LEVEL = 2 }

  # A HEADING DEEPER THAN THE LEVEL THIS SECTION IS WRITTEN AT (`##`, level
  # 2) IS A SUBHEADING NESTED INSIDE IT, NOT A BOUNDARY. project-
  # configuration.md allows prose beside the two tables, and a `###`
  # subheading is exactly that — plausible prose an author reaches for.
  # Only a heading AT OR ABOVE that level (a sibling `##` section, or a `#`
  # document heading) ends the scan; a deeper one is read as ordinary text
  # within the section and falls through to the table scan below, same as
  # a paragraph of prose would.
  /^#+[[:space:]]/ {
    match($0, /^#+/)
    hlevel = RLENGTH
    is_own = (tolower($0) ~ heading_re)
    if (in_sec && !is_own && hlevel > SEC_LEVEL) {
      if (in_blk) { flush_block(); in_blk = 0 }
      next
    }
    if (in_blk) { flush_block(); in_blk = 0 }
    in_sec = is_own
    if (in_sec) heading_line = NR
    next
  }

  !in_sec { next }

  {
    line = $0
    probe = line
    gsub(/\r/, "", probe)
    if (probe ~ /^[[:space:]]*\|/) {
      if (!in_blk) { in_blk = 1; blk_n = 0 }
      blk_n++
      blk_line[blk_n] = line
      blk_no[blk_n] = NR
      next
    }
    if (in_blk) { flush_block(); in_blk = 0 }
  }

  END {
    if (in_blk) flush_block()
    printf "#SUMMARY\t%d\n", heading_line
  }

  # One contiguous run of table rows. Its first row is the header, and the
  # header is what says which of the two tables this is.
  function flush_block(   i, j, n, cells, hdr, kind, want, other) {
    n = split_cells(blk_line[1], cells)
    hdr = ""
    for (i = 1; i <= n; i++) hdr = hdr (i > 1 ? "|" : "") foldcell(cells[i])
    kind = ""
    want = 0
    if (hdr == "setting|value") { kind = "setting"; want = 2 }
    else if (hdr == "command|runs") { kind = "command"; want = 2 }

    # A HEADER THIS GUARD DOES NOT RECOGNISE IS A VIOLATION, never a table it
    # skips — skipping would let a reordered or misspelled header leave every
    # cell present and every one meaning nothing, and a section whose only
    # table was skipped would declare nothing and pass.
    if (kind == "") {
      violation(blk_no[1], "a table under `## visual verification` whose header is `" hdr "` — the section holds one settings table (`Setting | Value`) and one commands table (`Command | Runs`), and nothing else in it is read")
      return
    }

    for (i = 2; i <= blk_n; i++) {
      n = split_cells(blk_line[i], cells)
      if (is_delimiter(n, cells)) continue

      # A SECOND HEADER INSIDE ONE BLOCK IS TWO TABLES THAT MERGED — the same
      # hazard check-workspace-isolation.sh flush_block guards against, and
      # the same fix: abandon the block rather than read its later rows
      # against the wrong columns.
      other = ""
      for (j = 1; j <= n; j++) other = other (j > 1 ? "|" : "") foldcell(cells[j])
      if (other == "setting|value" || other == "command|runs") {
        violation(blk_no[i], "a second table header (`" other "`) begins here, inside the table that started at line " blk_no[1] " — two tables written with no blank line or prose between them are one block to this parser, so it read only the first header and stopped here rather than reading the rows below against the wrong columns; separate the two tables with a blank line")
        return
      }

      if (n != want) {
        violation(blk_no[i], "a " kind " row with " n " cell(s) where its header has " want " (expected " want ") — the row does not line up with its header, so it is dropped rather than read against the wrong columns")
        continue
      }

      if (kind == "setting") check_setting_row(blk_no[i], cells)
      else check_command_row(blk_no[i], cells)
    }
  }

  # One row of the settings table. The `Setting` vocabulary is closed —
  # `ui paths`, `screenshots`, `regression checkout` and `regression repo` —
  # matching this guard header above on why an unrecognised name is reported
  # rather than silently skipped. `push to default branch` is deliberately
  # not in this list: task 14 removed it from the contract, so a row naming
  # it is now an unrecognised Setting like any other.
  function check_setting_row(lineno, cells,   key, disp) {
    disp = trimcell(cells[1])
    key = foldcell(cells[1])
    if (key != "ui paths" && key != "screenshots" && key != "regression checkout" \
        && key != "regression repo") {
      violation(lineno, "Setting `" disp "` is not one of `ui paths`, `screenshots`, `regression checkout` or `regression repo` — the vocabulary is closed, so the row is dropped")
      return
    }
    if (key in set_seen) {
      violation(lineno, "a second `" disp "` row, the first being at line " set_seen[key] " — the row is dropped")
      return
    }
    set_seen[key] = lineno
    printf "#SET\t%d\t%s\t%s\n", lineno, key, trimcell(cells[2])
  }

  # One row of the commands table. The `Command` vocabulary is closed —
  # `setup`, `verify`, `capture` and `fingerprint` — for the identical reason.
  # `fingerprint` is optional and nothing below requires it (KAN-395): a
  # project with no served bundle declares none, and its absence is silent
  # here exactly as `setup`'s is.
  function check_command_row(lineno, cells,   key, disp) {
    disp = trimcell(cells[1])
    key = foldcell(cells[1])
    if (key != "setup" && key != "verify" && key != "capture" && key != "fingerprint") {
      violation(lineno, "Command `" disp "` is not one of `setup`, `verify`, `capture` or `fingerprint` — the vocabulary is closed, so the row is dropped")
      return
    }
    if (key in cmd_seen) {
      violation(lineno, "a second `" disp "` row, the first being at line " cmd_seen[key] " — the row is dropped")
      return
    }
    cmd_seen[key] = lineno
    printf "#CMD\t%d\t%s\t%s\n", lineno, key, trimcell(cells[2])
  }
AWK_PROG
  ) <(strip_bom_cat "$CFG"))"

# THE SENTINEL IS WHAT TELLS A CLEAN FILE FROM A VALIDATOR THAT NEVER RAN —
# check-workspace-isolation.sh's identical rationale: awk stderr is
# deliberately not captured, so whatever awk said about why is already in
# front of the operator.
case "$REPORT" in
  *$'#SUMMARY\t'*) ;;
  *)
    echo "check-visual-verification: the validator did not run to completion on $CFG — it produced no summary line, so nothing in the section was checked" >&2
    exit 2
    ;;
esac

HEADING_LINE="$(printf '%s\n' "$REPORT" | awk -F'\t' '$1 == "#SUMMARY" { print $2; exit }')"

# get_field <tag> <key> -> sets FOUND, F_LINE, F_VAL for the first matching
# #SET/#CMD row. A loop over the (small, closed-vocabulary) report rather
# than an associative array: this repository's guards avoid `declare -A` so
# they run unmodified under macOS's stock bash 3.2, and eight possible rows
# is nothing to loop over per lookup.
get_field() {
  local tag="$1" key="$2" rtag rline rkey rval
  FOUND=0
  F_LINE=""
  F_VAL=""
  while IFS=$'\t' read -r rtag rline rkey rval; do
    [ "$rtag" = "$tag" ] || continue
    [ "$rkey" = "$key" ] || continue
    FOUND=1
    F_LINE="$rline"
    F_VAL="$rval"
    return 0
  done <<< "$REPORT"
}

VIOLATIONS=""
NVIOLS=0
addv() {
  VIOLATIONS="${VIOLATIONS}${CFG}:$1: $2
"
  NVIOLS=$((NVIOLS + 1))
}

# The #VIOL lines from the table scan itself.
while IFS=$'\t' read -r rtag rline rmsg; do
  [ "$rtag" = "#VIOL" ] || continue
  addv "$rline" "$rmsg"
done <<< "$REPORT"

# require_nonempty <tag> <key> <label> — the shared shape behind the four
# "absent or empty" findings the plan lists for `ui paths`, `screenshots`,
# `verify` and `capture`.
require_nonempty() {
  local tag="$1" key="$2" label="$3"
  get_field "$tag" "$key"
  if [ "$FOUND" -eq 0 ]; then
    addv "$HEADING_LINE" "\`$label\` is absent — this setting is required"
  elif [ -z "$F_VAL" ]; then
    addv "$F_LINE" "\`$label\` is empty — this setting is required and non-empty"
  fi
}

require_nonempty "#SET" "ui paths" "ui paths"
require_nonempty "#SET" "screenshots" "screenshots"
require_nonempty "#CMD" "verify" "verify"
require_nonempty "#CMD" "capture" "capture"

# trim_glob_element (sourced above, lib/trim-glob-element.sh) applies the
# same parse check-visual-trigger.sh applies to `ui paths` after splitting on
# comma — both guards must agree on what counts as a usable glob, and now
# share the one definition per design.md's `shared-trim-glob-element`
# decision, which superseded `no-shared-parser-yet`.

# `ui paths` YIELDING NO USABLE GLOB IS THE WORSE HALF OF THIS DEFECT, per
# design.md's `validator-agrees-with-trigger`: require_nonempty above only
# rejects a wholly empty cell, but a cell that is non-empty as a whole —
# e.g. two empty backtick pairs joined by a comma, "``, ``" — resolves,
# after the SAME split-then-strip parse check-visual-trigger.sh applies, to
# zero usable elements. That is exactly the condition under which the
# trigger guard itself exits 2 ("resolved to no usable glob"), so a
# validator that stayed silent on it would keep blessing a declaration the
# trigger guard cannot use. Runs only when `ui paths` is present and
# non-empty — an absent or empty cell already has its own finding above, and
# reporting this one too would name the same row twice under two messages.
get_field "#SET" "ui paths"
if [ "$FOUND" -eq 1 ] && [ -n "$F_VAL" ]; then
  IFS=',' read -r -a UI_PATH_ELEMS <<< "$F_VAL"
  UI_PATH_USABLE=0
  for elem in "${UI_PATH_ELEMS[@]}"; do
    elem="$(trim_glob_element "$elem")"
    [ -n "$elem" ] && UI_PATH_USABLE=$((UI_PATH_USABLE + 1))
  done
  if [ "$UI_PATH_USABLE" -eq 0 ]; then
    addv "$F_LINE" "\`ui paths\` is \`$F_VAL\` — splitting on commas and stripping each element's backticks and whitespace yields no usable glob, so check-visual-trigger.sh cannot use it either"
  fi
fi

get_field "#SET" "regression checkout"
CHECKOUT_FOUND="$FOUND"
CHECKOUT_LINE="$F_LINE"
CHECKOUT_VAL="$F_VAL"

get_field "#SET" "regression repo"
REPO_FOUND="$FOUND"
REPO_LINE="$F_LINE"
REPO_VAL="$F_VAL"

# `regression repo` present with no `regression checkout`, or the reverse —
# design.md names no use for either half declared alone: a repo with nothing
# to check its origin against, or a checkout nothing names as the repository
# it must equal.
if [ "$REPO_FOUND" -eq 1 ] && [ "$CHECKOUT_FOUND" -eq 0 ]; then
  addv "$REPO_LINE" "\`regression repo\` is present with no \`regression checkout\` — a repo with nothing to check its origin against declares nothing checkable"
fi
if [ "$CHECKOUT_FOUND" -eq 1 ] && [ "$REPO_FOUND" -eq 0 ]; then
  addv "$CHECKOUT_LINE" "\`regression checkout\` is present with no \`regression repo\` — nothing names the repository its origin must equal"
fi

CHECKOUT_VALID=0
if [ "$CHECKOUT_FOUND" -eq 1 ]; then
  if [ ! -d "$CHECKOUT_VAL" ]; then
    addv "$CHECKOUT_LINE" "\`regression checkout\` names \`$CHECKOUT_VAL\`, which is not an existing directory"
  elif ! git_clean -C "$CHECKOUT_VAL" rev-parse --git-dir >/dev/null 2>&1; then
    addv "$CHECKOUT_LINE" "\`regression checkout\` names \`$CHECKOUT_VAL\`, which is not a git checkout"
  else
    CHECKOUT_VALID=1
  fi
fi

# THE SECURITY-RELEVANT CHECK. `regression repo` is an identity assertion,
# not an authorisation — nothing is granted by a match, per design.md's
# `no-automatic-push` decision, which dropped this contract's earlier
# `push to default branch` setting after a panel slot demonstrated that both
# sides of this same equality lived in this one pull-request-editable file
# and so authorised nothing. Run whenever there is something to verify: a
# `regression checkout` that validated as a real git checkout above, and a
# declared `regression repo` to compare its origin against — unconditional
# on any push setting, because none survives in the contract. Every other
# combination already has its own finding above (a missing pairing or an
# invalid checkout), so running this too would report the same
# misconfiguration twice under two different names.
if [ "$CHECKOUT_VALID" -eq 1 ] && [ "$REPO_FOUND" -eq 1 ]; then
  GIT_ERR_FILE="$(mktemp "${TMPDIR:-/tmp}/check-visual-verification.XXXXXX")"
  trap 'rm -f "$GIT_ERR_FILE"' EXIT
  set +e
  ORIGIN_URL="$(git_clean -C "$CHECKOUT_VAL" remote get-url origin 2>"$GIT_ERR_FILE")"
  GIT_RC=$?
  set -e
  GIT_ERR="$(cat "$GIT_ERR_FILE")"
  rm -f "$GIT_ERR_FILE"
  trap - EXIT

  if [ "$GIT_RC" -eq 0 ]; then
    if [ "$ORIGIN_URL" != "$REPO_VAL" ]; then
      addv "$CHECKOUT_LINE" "the checkout's real \`origin\` (\`$ORIGIN_URL\`) does not equal the declared \`regression repo\` (\`$REPO_VAL\`)"
    fi
  else
    case "$GIT_ERR" in
      *[Nn]o\ such\ remote*)
        # A real, valid checkout with no `origin` remote at all cannot
        # honour the identity assertion either — a fact about the
        # declaration, not a failure to look, so it is reported and the run
        # continues.
        addv "$CHECKOUT_LINE" "the checkout at \`$CHECKOUT_VAL\` has no \`origin\` remote configured, so it cannot match the declared \`regression repo\` (\`$REPO_VAL\`)"
        ;;
      *)
        # Any OTHER failure of that command is not a fact about the
        # declaration; it is this guard failing to look, and it escalates
        # the whole run rather than being folded into a "clean" verdict.
        # $CHECKOUT_VAL is a cell out of `.flow/project.md` and $GIT_ERR is
        # git's own stderr against it, so both go through sanitize_display
        # exactly as every other interpolated cell in this file does.
        printf 'check-visual-verification: git remote get-url origin failed unexpectedly in %s (exit %s): %s — cannot determine whether the identity assertion holds\n' "$CHECKOUT_VAL" "$GIT_RC" "$GIT_ERR" | sanitize_display >&2
        exit 2
        ;;
    esac
  fi
fi

if [ "$NVIOLS" -ne 0 ]; then
  printf '%s' "$VIOLATIONS" | sanitize_display
  printf 'VISUAL-INVALID: %s — %s violation(s) in the `## visual verification` section\n' "$CFG" "$NVIOLS"
  exit 1
fi

printf 'VISUAL-OK: %s — the `## visual verification` section validated\n' "$CFG"
exit 0
