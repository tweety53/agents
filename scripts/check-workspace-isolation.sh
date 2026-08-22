#!/usr/bin/env bash
# check-workspace-isolation.sh — validate a project's `## workspace isolation`
# declaration against the contract that defines it.
#
# Usage: check-workspace-isolation.sh [<project root> ...]
#
# With no arguments it checks the repository this script ships in. Each argument
# is a PROJECT ROOT — the directory holding `.myflow/project.md` — never the
# configuration file itself, because that is the path every caller already has.
#
# WHO RUNS IT, AND WHY THAT IS TWO CALLERS RATHER THAN ONE. `/myflow-do` runs it
# against each apply worktree before it resolves or exports a single row —
# section 7 of skills/myflow-do/SKILL.md — and that is the call that reaches
# every project myflow is installed into, because the read happens in the
# project rather than here. This repository ALSO lists it in its own `## lint`,
# which is a self-check on this repository's own `## workspace isolation`
# section and nothing more: a green lint run says nothing about any other
# project. Reading the lint entry as the enforcement was exactly the shape of
# false pass every branch below is written against, one level up.
#
# Prints one violation line per finding, then ONE verdict line per project:
#   ISOLATION-OK:      <path> — <what was checked>
#   ISOLATION-INVALID: <path> — <n> violation(s)
#
# Exit 0 when every project checked is well formed, 1 when any violation was
# found, 2 when it cannot answer at all — a project root that is not a
# directory, a `.myflow/project.md` that is not a regular file (a directory, a
# fifo, a dangling symlink) or cannot be read, or a scan of it that failed rather
# than found nothing. The violations are on stdout and the refusals on stderr, so
# a caller reading stdout is reading findings and never an excuse for their
# absence.
#
# WHY THE VERDICT WORDS DIFFER FROM check-cleanup-complete.sh's. That guard says
# COMPLETE/LEFTOVER about the same section of the same file, and the two
# vocabularies are deliberately not shared. It asks whether the removal
# HAPPENED; this one asks whether the declaration is WELL FORMED. A reader who
# carried COMPLETE across to OK would be carrying an answer about resources into
# an answer about text.
#
# NEVER FAIL OPEN — the rule every branch below is written to. A project that
# declares no section passes SILENTLY, and that is the overwhelmingly common
# case, so every OTHER way of ending up with nothing to check has to be told
# apart from it: an unreadable configuration refuses, a duplicated heading is a
# violation, a table whose header this guard does not recognise is a violation,
# and a row whose column count does not match its header is a violation. Reading
# "declares nothing" out of any of those would report every malformed project
# valid, which is the failure this guard exists to prevent.
#
# A COMMAND'S FAILURE IS NEVER READ AS ITS NEGATIVE ANSWER, which is the same
# invariant one level down. `-r` is true of a directory and `grep -q` then fails
# with the same exit status it uses for "no match", so a `.myflow/project.md`
# that is a directory once answered "declares no section" and exited 0. Every
# test below therefore says which question it is answering: the file type is
# tested before its readability, and each `grep` exit status is read as three
# outcomes — matched, did not match, could not look — rather than as two.
#
# WHAT IS CANONICAL, AND WHAT THIS FILE IS. Every rule enforced here is stated
# by **Project configuration** (`skills/myflow-contracts/project-configuration.md`), which wins on
# any disagreement — under its "How a `## workspace isolation` section is
# written" paragraph and the two bullets that follow "An isolation row resolves
# under the same rules this file applies to everything else it consumes". This
# file restates none of the reasoning behind a rule; it cites the paragraph and
# enforces the rule. Which rules are mechanical and which remain the agent's to
# apply is recorded in that same file, under "Which of these rules a script
# checks, and which are left to the agent", so the contract does not imply that
# all of them are checked.
#
# THE PARAGRAPH NAMES ABOVE ARE PROSE, NOT CITATIONS, AND THAT IS DELIBERATE.
# check-references.sh resolves a bold token against the REAL HEADINGS of the
# file it names, and project-configuration.md is structured almost entirely as
# bold lead-in paragraphs rather than headings. A bold token naming one of those
# would be a stale reference the moment anything checked it. The paragraph names
# here are unbolded for that reason, and the only bold tokens are ones that
# resolve to a heading of that file. No COUNT of its headings is written here:
# one was — "exactly one, its title" — and it went stale the first time a
# heading was added to that file.
#
# THE INPUT IS ATTACKER-INFLUENCED. `.myflow/project.md` is tracked in the
# repository and editable in any pull request. NOTHING read here is executed:
# this guard never runs a project's `create`, `remove` or `survivors` command,
# never resolves a path out of the file, and never interpolates a cell into a
# shell. The whole of the parsing happens inside one awk program whose only
# input is the file's text — which is also why the cell splitter below is a
# character walk rather than a field split.
#
# THE PARSER IS THE CLEANUP GUARD'S, WIDENED FROM TWO COLUMNS TO FOUR. Its
# hazards are the same because its input is: a `|` inside a cell, written `\|`,
# must not split the row; a row may omit its trailing `|`; a file may arrive
# with CRLF line endings; and a row may be indented. THE PARSER ITSELF stays a
# copy, for the reason recorded in check-cleanup-complete.sh's containment
# comment — these guards are single-file by design and are copied into
# projects one at a time, so a sourced helper would make a guard that is
# present but unrunnable.
#
# `resolve_file` IS THE ONE EXCEPTION TO THAT RULE ON THIS FILE, and the
# exception is deliberate rather than accidental. Unlike the containment
# parser above — which a project can and does copy into its OWN tooling,
# standalone, with nothing else from this repository — this guard's own
# distribution is the KAN-73 symlink farm: it is never hand-copied anywhere,
# only ever reached at `<repo>/scripts/check-workspace-isolation.sh` or
# through a `skills/<name>/scripts/check-workspace-isolation.sh` symlink that
# KAN-73's install carries, and both directories carry a `lib` symlink
# beside it as part of that same farm (KAN-73's design.md, "The guard-to-skill
# map"). "A sourced helper would make a guard present but unrunnable" is
# exactly the concern that fails to apply here: the helper travels WITH the
# guard, unconditionally, wherever the guard is reachable at all — never
# absent, so never a way to be present-but-unrunnable. This is the identical
# argument KAN-153's review (F7) accepted for check-unfinished-work.sh sourcing
# scripts/lib/panel-record.sh, on the evidence that `setup.sh` installs
# skills, not `scripts/`, so a guard and the library it sources always travel
# together in this repository — see that review's Disposition section for the
# full reasoning, carried forward here rather than re-litigated. Left
# genuinely single-file, out of this change's scope: gather-self-review-context.sh,
# which is copied into OTHER projects' own tooling standalone and cannot assume
# any sibling travels with it.
#
# A FENCED EXAMPLE INSIDE THE SECTION IS PARSED AS A REAL TABLE, and that is a
# known limitation rather than an oversight. This guard carries no fence
# tracker, exactly as the cleanup guard's parser carries none, so a project that
# writes a specimen table inside the section will have it validated. The failure
# direction is the safe one — a specimen is reported rather than silently
# skipped — and the remedy is to put the example outside the section, where
# nothing reads it.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# resolve_file — sourced from lib/resolve-file.sh; see that file's header for
# why this guard, unlike gather-self-review-context.sh, may source a sibling
# instead of carrying its own copy. "One level above $SCRIPT_DIR" is NOT enough to derive the
# repository root, because this script is now reachable from more than one
# directory — its real home at <repo>/scripts/, and a skills/<name>/scripts/
# symlink a command skill carries it under (prepare-workspace.sh's own call
# site). Going up one level from the SECOND of those lands on the skill
# directory, which exists, so a fixed-depth guard would proceed against it
# and return a confident wrong answer (silently "declares nothing") rather
# than an error. Resolving this script's own symlinks first, and deriving the
# root from ITS real physical location, gives the same answer regardless of
# which path invoked it.
source "$SCRIPT_DIR/lib/resolve-file.sh"

# CHECK_WORKSPACE_ISOLATION_PRINT_ROWS — an internal, undocumented-to-operators
# switch that prepare-workspace.sh sets when it invokes this guard, so it can
# derive the workspace variables from the SAME parse this guard already did
# rather than opening `.myflow/project.md` a second time and re-implementing
# the cell splitter. Unset (the default), it changes nothing: the awk program
# below never emits a `#ROW` line, so stdout is byte-for-byte what it always
# was. Set, each validated resource row is appended to this project's report
# as a `#ROW\t<resource>\t<variable>\t<default>\t<in a workspace>` line, after
# the verdict line, and only for a project whose report reached its verdict —
# a project this guard refused on (exit 2) never reaches the printf that emits
# them. A caller that does not know this switch exists reads exactly the
# output documented above; only prepare-workspace.sh reads the `#ROW` lines.
PRINT_ROWS="${CHECK_WORKSPACE_ISOLATION_PRINT_ROWS:-0}"

# The self-resolution above is deferred to exactly here, inside the branch
# that is the only reader of its answer. Every OTHER caller passes an
# explicit project root — `/myflow-do` against each apply worktree, and this
# repository's own `## lint` entry against `$REPO_ROOT` computed the same
# way one level up — and resolving unconditionally at the top of the file
# meant a failure to resolve THIS script's own location could abort a run
# that never needed the answer at all: the production path always supplies
# an argument. `plan-dispatch-bundles.sh` already resolves this lazily, with
# its own comment on why; this makes the two consistent.
if [ "$#" -eq 0 ]; then
  SELF_REAL="$(resolve_file "${BASH_SOURCE[0]}")" || {
    echo "check-workspace-isolation: cannot resolve this script's own location" >&2
    exit 2
  }
  REPO_ROOT="$(cd "$(dirname "$SELF_REAL")/.." && pwd)"
  set -- "$REPO_ROOT"
fi

# The heading rule, written once and used by the presence test, the duplicate
# count and the extraction alike. Two spellings of it would be two answers to
# "is this project isolated?", and the extraction's would win silently. It is
# check-cleanup-complete.sh's rule character for character, so a section that
# guard reads is a section this one validates.
ISO_HEADING='^##[[:space:]]+workspace isolation[[:space:]]*$'

TOTAL_VIOLATIONS=0

# sanitize_display — copy stdin to stdout with every C0 control byte, DEL and
# backslash rendered as visible text. Every violation line below quotes a cell
# STRAIGHT OUT of `.myflow/project.md`, which is tracked in the repository and
# editable in any pull request, so a cell holding an escape sequence writes that
# sequence to the operator's terminal. `\033[2K` erases the line it is on and
# `\033[1;32m` turns the next one bold green: a cell can therefore leave a report
# whose BYTES say ISOLATION-INVALID and whose RENDERING shows a convincing
# ISOLATION-OK. The findings are the whole of what an operator acts on here, so
# they have to be incapable of controlling the terminal that shows them.
#
# ESCAPED, NOT STRIPPED AND NOT REFUSED. Stripping changes the text the operator
# is being sent to go and fix, so they would search their configuration for a
# string that is not in it. Refusing hands whoever edited the file a way to
# suppress a real finding. Escaping keeps the finding, keeps it countable, and
# makes the offending byte visible as itself — a cell that carries an escape is
# something the author wants to see.
#
# THE ENCODING IS INJECTIVE, which is why `\` is escaped alongside the controls.
# Without that, a cell containing the four literal characters `\x1b` and a cell
# containing a real ESC byte produce the same report, and the operator cannot
# tell which one is in their file.
#
# C0 AND DEL, AND NOTHING ABOVE 0x7F. Bytes 0x80-0xFF are the continuation and
# lead bytes of ordinary UTF-8, so a project whose bucket is named `café-bücket`
# reports its real name; escaping them would mangle every non-ASCII declaration
# in exchange for nothing, since a UTF-8 terminal acts on none of them. `LC_ALL=C`
# is a prefix on this one command rather than an export, exactly as
# check-cleanup-complete.sh keeps it off its own environment: it makes `length`
# and `substr` count bytes here without changing the locale anything else sees.
sanitize_display() {
  LC_ALL=C awk '
    BEGIN {
      for (j = 1; j < 32; j++) esc[sprintf("%c", j)] = sprintf("\\x%02x", j)
      esc[sprintf("%c", 127)] = "\\x7f"
      esc["\\"] = "\\\\"
    }
    {
      out = ""
      n = length($0)
      for (i = 1; i <= n; i++) {
        c = substr($0, i, 1)
        out = out ((c in esc) ? esc[c] : c)
      }
      print out
    }
  '
}

for ROOT in "$@"; do
  if [ ! -d "$ROOT" ]; then
    echo "check-workspace-isolation: $ROOT is not a directory — cannot tell whether it declares a workspace isolation section" >&2
    exit 2
  fi

  CFG="$ROOT/.myflow/project.md"

  # THE ABSENT CASE IS TESTED WITH `-e` AND `-L` TOGETHER, not with `-e` alone.
  # `-e` follows symlinks, so a `.myflow/project.md` pointing at a path that does
  # not exist answers `-e` false and would be reported as a project that declares
  # nothing — a link anyone able to edit the repository can create, resolving to
  # the one verdict this guard must never give away. `-L` is true for the link
  # itself, so the pair says "there is nothing here at all" and a dangling link
  # falls through to the refusal below instead.
  if [ ! -e "$CFG" ] && [ ! -L "$CFG" ]; then
    # The file is optional. A project with no `.myflow/project.md` is a
    # supported, ordinary case, not an error.
    printf 'ISOLATION-OK: %s — no .myflow/project.md, so nothing is declared\n' "$ROOT"
    continue
  fi

  # A REGULAR FILE, TESTED BEFORE READABILITY, because `-r` ANSWERS ABOUT A
  # DIRECTORY TOO. `-r` on a directory is true, and `grep` then fails with "Is a
  # directory" while its exit status is indistinguishable from "no match" — so
  # `ln -s somedir .myflow/project.md` produced `ISOLATION-OK: … declares no
  # `## workspace isolation` section` and exit 0, which is this file's
  # NEVER FAIL OPEN invariant broken in the one script written to hold it. `-f`
  # follows symlinks, so a configuration that IS a symlink to a real file is
  # still read; what it excludes is a directory, a fifo (which `grep` would block
  # on forever) and a device.
  if [ ! -f "$CFG" ]; then
    echo "check-workspace-isolation: $CFG is not a regular file — cannot validate what it declares" >&2
    exit 2
  fi

  if [ ! -r "$CFG" ]; then
    # A file that exists but cannot be read is NOT "declares no isolation".
    # Reading absence out of a path that was never readable is the false pass
    # this guard exists to prevent, so it refuses instead of assuming.
    echo "check-workspace-isolation: $CFG exists but is not readable — cannot validate what it declares" >&2
    exit 2
  fi

  # GREP HAS THREE ANSWERS AND THIS READS ALL THREE. `if ! grep -q` collapses
  # them to two: exit 1 is "no match" and exit 2 or more is "I could not look",
  # and conflating the second with the first is the same fail-open the `-f` test
  # above closes, reached from an I/O error rather than from a file type. The
  # status is captured rather than branched on directly because `set -e` would
  # otherwise abort the run on the ordinary no-match answer.
  set +e
  grep -qiE "$ISO_HEADING" "$CFG"
  ISO_GREP_RC=$?
  set -e
  if [ "$ISO_GREP_RC" -ge 2 ]; then
    echo "check-workspace-isolation: grep exited $ISO_GREP_RC while looking for the '## workspace isolation' heading in $CFG — that is a failure to look rather than an absence, so nothing was validated" >&2
    exit 2
  fi
  if [ "$ISO_GREP_RC" -ne 0 ]; then
    printf 'ISOLATION-OK: %s — declares no `## workspace isolation` section\n' "$CFG"
    continue
  fi

  # TWO DECLARATIONS ARE NOT A DECLARATION. The file is hand-written, so a
  # heading duplicated by a bad merge or a copied block is a realistic mistake,
  # and the two sections can name two different commands against two different
  # services. check-cleanup-complete.sh reports the same shape as a SKIP,
  # because its question is whether it may run one of them; the question here is
  # whether the declaration is well formed, and an ambiguous one is not. The
  # count is reported because "you have two" is what the operator has to fix,
  # and a message naming the heading without saying it appears more than once
  # reads as the heading being wrong rather than repeated.
  #
  # Neither section is validated. Validating both would report each row twice,
  # and validating the first would be this guard picking the winner the contract
  # declines to pick.
  #
  # `|| true` IS NOT USED HERE, for the reason the readability tests above are
  # written the way they are: it turns every failure of the count into the number
  # zero, and zero is "declared once" — the answer that lets the run continue. A
  # heading this guard has already SEEN cannot legitimately count as fewer than
  # one, so anything but a clean count is a failure to count.
  set +e
  ISO_COUNT="$(grep -ciE "$ISO_HEADING" "$CFG")"
  ISO_COUNT_RC=$?
  set -e
  case "$ISO_COUNT_RC:$ISO_COUNT" in
    0:[0-9]*) ;;
    *)
      echo "check-workspace-isolation: could not count the '## workspace isolation' headings in $CFG (grep exited $ISO_COUNT_RC) — nothing was validated" >&2
      exit 2
      ;;
  esac
  if [ "$ISO_COUNT" -gt 1 ]; then
    printf '%s:0: the file declares %s `## workspace isolation` sections — a second declaration is an ambiguous declaration rather than a merge, so neither was validated and neither of their commands would be run\n' "$CFG" "$ISO_COUNT"
    TOTAL_VIOLATIONS=$((TOTAL_VIOLATIONS + 1))
    printf 'ISOLATION-INVALID: %s — 1 violation(s) in the `## workspace isolation` section\n' "$CFG"
    continue
  fi

  # The whole of the validation. awk prints one violation line per finding and
  # one `#SUMMARY <resource rows> <command rows>` line last; the shell below
  # turns those into the verdict. Keeping the counting in awk and the verdict in
  # the shell is what lets a caller pass several project roots and get one
  # verdict per project.
  #
  # NO APOSTROPHE MAY APPEAR ANYWHERE BELOW, comments included: the program is
  # delimited by single quotes, so one ends it mid-word and the shell then reads
  # the rest of this file as a string. The failure is not local to the line that
  # introduced it — bash reports an unmatched quote at the END of the file — so
  # the next editor to write "the contract's" here will spend the debugging time
  # this sentence exists to save them. Write "the contract states" instead.
  REPORT="$(awk -v cfg="$CFG" -v heading_re="$ISO_HEADING" -v print_rows="$PRINT_ROWS" '
    # ---------------------------------------------------------------------
    # Cell splitting. A character walk, not a field split: a realistic cell
    # contains a `|` written `\|`, and a split on `|` would cut the row there
    # and leave a shorter row that still looks well formed.
    # ---------------------------------------------------------------------
    function split_cells(line, out,   i, ch, nxt, n, cur, len, k) {
      for (k in out) delete out[k]
      gsub(/\r/, "", line)
      sub(/^[[:space:]]+/, "", line)
      sub(/[[:space:]]+$/, "", line)
      if (substr(line, 1, 1) != "|") return 0
      line = substr(line, 2)
      len = length(line)
      n = 0
      cur = ""
      for (i = 1; i <= len; i++) {
        ch = substr(line, i, 1)
        if (ch == "\\" && i < len) {
          nxt = substr(line, i + 1, 1)
          if (nxt == "|" || nxt == "\\") { cur = cur nxt; i++; continue }
          cur = cur ch
          continue
        }
        if (ch == "|") { n++; out[n] = cur; cur = ""; continue }
        cur = cur ch
      }
      n++
      out[n] = cur
      # A trailing `|` leaves one empty trailing field; a row that omits it
      # leaves a real one. Dropping only the empty one accepts both shapes.
      if (n > 1 && cur == "") { delete out[n]; n-- }
      return n
    }

    # The cell as the contract reads it: without its surrounding whitespace and
    # without the backticks a project writes around a literal.
    function trimcell(c) {
      gsub(/^[[:space:]`]+|[[:space:]`]+$/, "", c)
      return c
    }

    # A header cell, folded for comparison: lowercased, with internal runs of
    # whitespace collapsed, so `In  a workspace` and `In a workspace` are one
    # heading and `In a workspace` and `Workspace` are two.
    function foldcell(c) {
      c = tolower(trimcell(c))
      gsub(/[[:space:]]+/, " ", c)
      return c
    }

    function violation(lineno, msg) {
      printf "%s:%d: %s\n", cfg, lineno, msg
      nviol++
    }

    # A markdown delimiter row (`|----|----|`), which carries no data.
    function is_delimiter(n, cells,   i, c) {
      for (i = 1; i <= n; i++) {
        c = trimcell(cells[i])
        if (c !~ /^:?-+:?$/) return 0
      }
      return n > 0
    }

    BEGIN { in_sec = 0; in_blk = 0; nviol = 0; nres = 0; ncmd = 0; ntable = 0; iso_line = 0 }

    # A heading of any level ends the section, which is the same shape the
    # cleanup guard scans with: a project configuration is a human-written
    # Markdown file of sections, not a nested document.
    /^#+[[:space:]]/ {
      if (in_blk) { flush_block(); in_blk = 0 }
      in_sec = (tolower($0) ~ heading_re)
      if (in_sec) iso_line = NR
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
      resolve_refs()

      # A HEADING WITH NO TABLE UNDER IT DECLARED NOTHING. That is how a
      # mistyped header actually presents — the tables are there and neither was
      # recognised — and it is also what a section holding only prose is.
      # Passing it would report a project isolated that declared nothing, which
      # is the one answer that must be reserved for a project carrying no
      # heading at all.
      if (ntable == 0 && nviol == 0) {
        violation(iso_line, "the `## workspace isolation` section carries neither the resource table nor the command table — a heading with nothing under it declares nothing, and a project that is not isolated says so by carrying no heading")
      }

      printf "#SUMMARY %d %d\n", nres, ncmd
      exit (nviol > 0)
    }

    # -----------------------------------------------------------------------
    # Every `<value:...>` reference, resolved once the whole table has been
    # read. Resolving as each row was read would make a reference legal or
    # illegal according to the order the author happened to write the rows in,
    # and one level of reference resolves in a single pass precisely so there is
    # no resolution order to declare rows in.
    #
    # These findings are therefore reported AFTER the row findings rather than
    # in line order, which is what "a single pass" costs and is cheap: each line
    # names its own file and line number.
    # -----------------------------------------------------------------------
    function resolve_refs(   i, nm) {
      for (i = 1; i <= nref; i++) {
        nm = ref_name[i]
        if (!(nm in var_kind)) {
          violation(ref_line[i], "row `" ref_row[i] "`: " ref_tok[i] " names no row in this table — a reference names the row whose Variable column holds that name, so the row is dropped")
          continue
        }
        if (var_kind[nm] == "url") {
          violation(ref_line[i], "row `" ref_row[i] "`: " ref_tok[i] " names the `url` row at line " var_line[nm] ", and a reference may name a `database`, `bucket` or `port` row and never another `url` row — the row is dropped")
          continue
        }
        if (var_kind[nm] == "cache index") {
          violation(ref_line[i], "row `" ref_row[i] "`: " ref_tok[i] " names the `cache index` row at line " var_line[nm] ", whose value is claimed at run time rather than derived from the id and so cannot be referenced — the row is dropped")
          continue
        }
      }
    }

    # -----------------------------------------------------------------------
    # One contiguous run of table rows. Its first row is the header, and the
    # header is what says which of the two tables this is.
    # -----------------------------------------------------------------------
    function flush_block(   i, j, n, cells, hdr, kind, want, other) {
      n = split_cells(blk_line[1], cells)
      hdr = ""
      for (i = 1; i <= n; i++) hdr = hdr (i > 1 ? "|" : "") foldcell(cells[i])
      kind = ""
      want = 0
      if (hdr == "resource|variable|default|in a workspace") { kind = "resource"; want = 4 }
      else if (hdr == "command|runs") { kind = "command"; want = 2 }

      # A HEADER THIS GUARD DOES NOT RECOGNISE IS A VIOLATION, never a table it
      # skips. Skipping is the fail-open this guard exists to prevent: a
      # reordered header leaves every cell present and every one meaning
      # something else, and a section whose only table was skipped would declare
      # nothing and pass. The two column lists are stated in the message because
      # what the author has to see is the order they were meant to be in.
      if (kind == "") {
        violation(blk_no[1], "a table under `## workspace isolation` whose header is `" hdr "` — the section holds one resource table (`Resource | Variable | Default | In a workspace`, in that order) and one command table (`Command | Runs`), and nothing else in it is read")
        return
      }
      ntable++

      for (i = 2; i <= blk_n; i++) {
        n = split_cells(blk_line[i], cells)
        if (is_delimiter(n, cells)) continue

        # A SECOND HEADER INSIDE ONE BLOCK IS TWO TABLES THAT MERGED. Markdown
        # separates tables with a blank line, and two written back to back are
        # one contiguous run of rows to the scanner above — so this parser read
        # the FIRST header and would go on to evaluate the second table rows
        # against the first table columns. It did exactly that before this
        # branch existed: a resource table with a command table under it emitted
        # four "cell count mismatch" violations, one per command row, and named
        # nothing that would lead the author to the missing blank line. It failed
        # closed, which was right, and pointed at four lines that were each
        # individually correct, which was not.
        #
        # The block is ABANDONED here rather than split into two. Splitting would
        # be this guard repairing input the contract says to report, and every
        # row below this line belongs to a table whose header this parser never
        # started reading — so evaluating them is exactly the wrong-columns
        # reading the cell-count rule exists to refuse.
        other = ""
        for (j = 1; j <= n; j++) other = other (j > 1 ? "|" : "") foldcell(cells[j])
        if (other == "resource|variable|default|in a workspace" || other == "command|runs") {
          violation(blk_no[i], "a second table header (`" other "`) begins here, inside the table that started at line " blk_no[1] " — two tables written with no blank line or prose between them are one block to this parser, so it read only the first header and stopped here rather than reading the rows below against the wrong columns; separate the two tables with a blank line")
          return
        }

        # A ROW THAT DOES NOT LINE UP WITH ITS HEADER CANNOT BE READ AGAINST IT.
        # Reading a short row would report the wrong rule broken — a row missing
        # its `Default` presents as a row whose `In a workspace` cell is empty,
        # and the operator would be sent to fill in a cell that is missing
        # rather than blank.
        if (n != want) {
          violation(blk_no[i], "a " kind " row with " n " cell(s) where its header has " want " (expected " want ") — the row does not line up with its header, so it is dropped rather than read against the wrong columns")
          continue
        }

        if (kind == "resource") { nres++; check_resource_row(blk_no[i], cells) }
        else { ncmd++; check_command_row(blk_no[i], cells) }
      }
    }

    # The row named the way the operator has to find it: by the variable it
    # carries. A row reported only by its line number is one the operator has to
    # go and look up before they can act on it, and the contract asks for the
    # row to be reported BY NAME.
    function rowname(cells,   v) {
      v = trimcell(cells[2])
      return (v == "" ? "(no Variable cell)" : v)
    }

    # Every token in a cell, one per call, walked left to right. The shape is
    # check-cleanup-complete.sh, character for character: `<[A-Za-z0-9_:.-]+>`
    # admits `<id>`, `<id_underscored>` and `<value:NAME>` and excludes `>`, so
    # an ordinary shell redirection in a command cell carries no match. A wider
    # shape would report `cmd < in.txt > out.txt` as naming a token.
    function next_token(s, from, found,   m, start) {
      s = substr(s, from)
      m = match(s, /<[A-Za-z0-9_:.-]+>/)
      if (m == 0) { found[1] = ""; return 0 }
      found[1] = substr(s, m, RLENGTH)
      return from + m + RLENGTH - 1
    }

    # -----------------------------------------------------------------------
    # A TOKEN WITH A SPACE IN IT IS INVISIBLE TO next_token, AND THAT IS THE
    # LIKELIEST TYPO AT THIS KEYBOARD. `<value: API_PORT>` and `< id_underscored>`
    # match the shape above nowhere, so the walk finds no token, every rule keyed
    # on tokens passes, and the cell is read as ordinary literal text — a row
    # reported valid whose value would reach a run with the brackets still in it.
    #
    # THE SHAPE CANNOT SIMPLY BE WIDENED TO ADMIT WHITESPACE, which is what the
    # narrow one is there to prevent: `cmd < in.txt > out.txt` is an ordinary
    # redirection whose bracketed run holds only token characters and a space, so
    # a widened next_token would report every command written that way. A guard
    # that cried wolf on real commands would be worked around rather than fixed.
    #
    # SO THE TEST IS "IS THIS A NEAR MISS OF A TOKEN THIS CONTRACT DEFINES", not
    # "does this look bracketed". The vocabulary is closed and tiny, so a
    # bracketed run whose whitespace-squeezed content IS one of those names was
    # meant to be that token and was mistyped; one whose content is anything else
    # is not something this contract defines. `< in.txt >` squeezes to `in.txt` and is
    # left alone; `<value: API_PORT>` squeezes to `value:API_PORT` and is
    # reported.
    #
    # WHAT IT DELIBERATELY DOES NOT CATCH is a bracketed run that is neither a
    # near miss nor a redirection — `<my thing>` is read as literal text, exactly
    # as it is today. Reporting it would mean deciding that every `<`…`>` in a
    # cell is a token attempt, which is the wolf-crying above.
    function check_spaced_tokens(lineno, who, cell, ctx,   s, m, run, body) {
      s = cell
      while ((m = match(s, /<[^<>]*>/)) > 0) {
        run = substr(s, m, RLENGTH)
        s = substr(s, m + RLENGTH)
        body = substr(run, 2, length(run) - 2)
        if (body !~ /[[:space:]]/) continue
        gsub(/[[:space:]]+/, "", body)
        if (body == "id" || body == "id_underscored" || body == "offset" \
            || body ~ /^value:[A-Za-z_][A-Za-z0-9_]*$/) {
          violation(lineno, ctx " `" who "`: `" run "` carries whitespace inside its brackets, so it is not the token `<" body ">` and nothing substitutes it — a bracketed run with a space in it is read as literal text, and the row is dropped")
        }
      }
    }

    # -----------------------------------------------------------------------
    # One row of the resource table.
    # -----------------------------------------------------------------------
    function check_resource_row(lineno, cells,   res, who, var, def, ws, pos, tok, ref, tokbuf) {
      # Emitted BEFORE any of the checks below run, and unconditionally for
      # every row this function is called on — a row already known, from the
      # n == want test in flush_block, to line up with the resource header.
      # A row this function goes on to reject still gets its `#ROW` line:
      # prepare-workspace.sh never reads it, because a rejected row means
      # this guard exits non-zero and prepare-workspace.sh exits on that
      # before it ever looks at a `#ROW` line.
      if (print_rows) {
        printf "#ROW\t%s\t%s\t%s\t%s\n", trimcell(cells[1]), trimcell(cells[2]), trimcell(cells[3]), trimcell(cells[4])
      }
      who = rowname(cells)

      # The `Resource` word, folded for case and internal whitespace and for
      # nothing else. Case is not the mistake anyone makes here, and `cache
      # index` is two words whose space is part of the word — so a fold that
      # REMOVED whitespace would accept `cacheindex` and one that split on it
      # would reject the real spelling.
      res = foldcell(cells[1])
      if (res != "database" && res != "bucket" && res != "cache index" \
          && res != "port" && res != "url") {
        violation(lineno, "row `" who "`: Resource `" trimcell(cells[1]) "` is not one of `database`, `bucket`, `cache index`, `port` or `url` — the vocabulary is closed, so the row is dropped and nothing removes it")
        return
      }

      # The `Variable`, which a run EXPORTS. The shape is the one the contract
      # states, and the ranges are written out rather than as `[[:alpha:]]`
      # because a character class is exactly the construct that starts meaning
      # something else in another locale — the same reasoning the Protection 1
      # comment in check-cleanup-complete.sh records for its enumerated set.
      var = trimcell(cells[2])
      if (var == "") {
        violation(lineno, "resource row: the Variable cell is empty — a row with no variable carries nothing, so the row is dropped")
        return
      }
      if (var !~ /^[A-Za-z_][A-Za-z0-9_]*$/) {
        violation(lineno, "row `" who "`: Variable `" var "` is not a legal environment-variable name (`[A-Za-z_][A-Za-z0-9_]*`) — a run exports these, so the row is dropped")
        return
      }

      # How many times this `Resource` word may appear. `port` and `url` repeat
      # by design — once per port the project publishes, once per composed value
      # it declares — while the other three name one thing each, and a second
      # `database` row is two databases one workspace id cannot name.
      if (res == "database" || res == "bucket" || res == "cache index") {
        if (res in res_seen) {
          violation(lineno, "row `" who "`: a second `" res "` row, the first being at line " res_seen[res] " — `database`, `bucket` and `cache index` each appear at most once, so the row is dropped")
        } else {
          res_seen[res] = lineno
        }
      }

      # The variable index, which is what a `<value:...>` reference resolves
      # against. A reference names THE row whose Variable column holds that
      # name — singular — so two rows holding one name resolve to no single row,
      # and a guard that took the first would resolve to whichever the author
      # happened to write first.
      if (var in var_kind) {
        violation(lineno, "row `" who "`: a second resource row holds the Variable `" var "`, the first being at line " var_line[var] " — a `<value:" var ">` reference names one row and cannot name two, and a run would export one value over the other")
      } else {
        var_kind[var] = res
        var_line[var] = lineno
      }

      # The bare-integer `Default`, which binds `port` and `cache index` and no
      # other resource. A `database` default is a connection string and a `url`
      # default is a URL, so a rule applied to every row would reject every real
      # project. `+` and `-` are excluded along with everything else that is not
      # a digit: a signed default is not a port number, and reading one would
      # make the offset arithmetic mean something nobody wrote.
      def = trimcell(cells[3])
      if ((res == "port" || res == "cache index") && def !~ /^[0-9]+$/) {
        violation(lineno, "row `" who "`: Default `" def "` is not a bare integer, which a `" res "` row requires — the workspace value is arithmetic on this number, so the row is dropped")
        return
      }

      # The `In a workspace` cell, whose form the ROW Resource selects. The
      # check is on the FORM and not only on the tokens, which is what catches
      # the cell that carries none: a `port` row whose cell reads `9090` names
      # no token at all, so a token-only check passes it while it matches no
      # form and would be added to its own `Default`.
      ws = trimcell(cells[4])

      if (res == "port") {
        if (ws != "+<offset>") {
          violation(lineno, "row `" who "`: the In a workspace cell is `" ws "`, and a `port` row writes the literal `+<offset>` with nothing before or after it — the row is dropped")
        }
        return
      }

      if (res == "cache index") {
        if (ws != "probed") {
          violation(lineno, "row `" who "`: the In a workspace cell is `" ws "`, and a `cache index` row writes the literal `probed` alone, because the index is claimed at run time rather than derived — the row is dropped")
        }
        return
      }

      # `database`, `bucket` and `url` all write the value out in full, so what
      # separates them is which tokens they may carry: `<value:...>` is legal in
      # a `url` cell and nowhere else, because a `database` or `bucket` cell is
      # what the removal commands target and a removal target assembled out of
      # other rows would depend on those rows having resolved first.
      if (ws == "") {
        violation(lineno, "row `" who "`: the In a workspace cell is empty, and a `" res "` row writes its value out in full — the row is dropped")
        return
      }

      check_spaced_tokens(lineno, who, cells[4], "row")

      pos = 1
      while ((pos = next_token(cells[4], pos, tokbuf)) > 0) {
        tok = tokbuf[1]
        if (tok == "<id>" || tok == "<id_underscored>") continue
        if (tok ~ /^<value:/) {
          if (res != "url") {
            violation(lineno, "row `" who "`: the In a workspace cell names " tok ", and `<value:...>` is legal in a `url` cell and nowhere else — the row is dropped")
            continue
          }
          # The reference is RESOLVED at END, once every row has been read.
          # Doing it here would make a reference legal or illegal according to
          # the order the author happened to write the rows in.
          ref = substr(tok, 8, length(tok) - 8)
          nref++
          ref_name[nref] = ref
          ref_row[nref] = who
          ref_line[nref] = lineno
          ref_tok[nref] = tok
          continue
        }
        violation(lineno, "row `" who "`: the In a workspace cell names " tok ", which is not a token this contract substitutes (`<id>`, `<id_underscored>`" (res == "url" ? ", `<value:VARIABLE>`" : "") ") — the row is dropped")
      }
    }

    # -----------------------------------------------------------------------
    # One row of the command table. A command row is EXECUTED, at the same trust
    # level as the `## run`, `## test` and `## lint` commands, which carry no
    # containment rule either — so the one thing checked in its text is its
    # tokens. No path is isolated inside a command and none is looked for:
    # finding the path in an arbitrary command means parsing a shell command,
    # and a parser that is wrong about one command reports the wrong thing about
    # all of them.
    # -----------------------------------------------------------------------
    function check_command_row(lineno, cells,   verb, cmd, pos, tok, tokbuf) {
      verb = foldcell(cells[1])
      if (verb != "create" && verb != "remove" && verb != "survivors") {
        violation(lineno, "command row `" trimcell(cells[1]) "`: the command table has three verbs — `create`, `remove` and `survivors` — and nothing calls anything else, so the row is dropped and never runs")
        return
      }
      if (verb in cmd_seen) {
        violation(lineno, "command row `" verb "`: a second `" verb "` row, the first being at line " cmd_seen[verb] " — two commands under one verb resolve to whichever is read first, so the row is dropped")
        return
      }
      cmd_seen[verb] = lineno

      cmd = trimcell(cells[2])
      if (cmd == "") {
        violation(lineno, "command row `" verb "`: the Runs cell is empty — a verb declared with no command is a step that would run as the empty string, so the row is dropped")
        return
      }

      # `<id>` and `<id_underscored>` are substituted and nothing else is.
      # `<value:...>` is deliberately among the things that are not: a command
      # needing a derived value reads the exported variable, which is already in
      # its environment by the time it runs.
      check_spaced_tokens(lineno, verb, cells[2], "command row")

      pos = 1
      while ((pos = next_token(cells[2], pos, tokbuf)) > 0) {
        tok = tokbuf[1]
        if (tok == "<id>" || tok == "<id_underscored>") continue
        violation(lineno, "command row `" verb "`: the command names " tok ", and only `<id>` and `<id_underscored>` are substituted in a command — the pipeline would hand the shell a literal nobody intended, so the row is dropped")
      }
    }
  ' "$CFG")" || true

  # THE SENTINEL IS WHAT TELLS A CLEAN FILE FROM A VALIDATOR THAT NEVER RAN.
  # awk exits non-zero when it found violations, so the exit status cannot carry
  # this — and an empty report looks identical either way from out here. The
  # `#SUMMARY` line is printed unconditionally at the end of the program, so its
  # ABSENCE means the program did not reach its end: no awk on this machine, a
  # syntax error introduced by an edit, a signal. Every one of those has checked
  # nothing, and reporting ISOLATION-OK over them would pass every malformed
  # project on a machine where the parser is broken — the exact fail-open this
  # guard exists to close, reached from inside the guard rather than from its
  # input. awk stderr is deliberately not captured: it is inherited, so whatever
  # awk said about why is already in front of the operator.
  case "$REPORT" in
    *'#SUMMARY '*) ;;
    *)
      echo "check-workspace-isolation: the validator did not run to completion on $CFG — it produced no summary line, so nothing in the section was checked" >&2
      exit 2
      ;;
  esac

  SUMMARY="$(printf '%s\n' "$REPORT" | awk '/^#SUMMARY /{ print; exit }')"
  # `#ROW` lines are neither a finding nor the summary — carved out here so
  # they are never sanitized-and-printed as a violation and never counted
  # toward N_FOUND below. They exist in $REPORT only when PRINT_ROWS is set,
  # which no caller but prepare-workspace.sh sets.
  FINDINGS="$(printf '%s\n' "$REPORT" | awk '!/^#SUMMARY / && !/^#ROW\t/')"
  ROWS="$(printf '%s\n' "$REPORT" | awk '/^#ROW\t/')"
  N_RES="$(printf '%s' "$SUMMARY" | awk '{ print $2 + 0 }')"
  N_CMD="$(printf '%s' "$SUMMARY" | awk '{ print $3 + 0 }')"

  if [ -n "$FINDINGS" ]; then
    # Sanitized at the ONE point they reach stdout, rather than at each of the
    # twenty places a cell is interpolated into a message: a chokepoint covers
    # the next message somebody adds, and twenty call sites are twenty chances to
    # forget. The count is taken from the unsanitized text because the escaping
    # never adds or removes a line.
    printf '%s\n' "$FINDINGS" | sanitize_display
    N_FOUND="$(printf '%s\n' "$FINDINGS" | wc -l | tr -d ' ')"
    TOTAL_VIOLATIONS=$((TOTAL_VIOLATIONS + N_FOUND))
    printf 'ISOLATION-INVALID: %s — %s violation(s) in the `## workspace isolation` section\n' "$CFG" "$N_FOUND"
  else
    printf 'ISOLATION-OK: %s — %s resource row(s) and %s command row(s) validated\n' "$CFG" "$N_RES" "$N_CMD"
  fi

  # `#ROW` lines, if any, always come last — after the verdict line, so a
  # caller not reading them (every caller but prepare-workspace.sh, and this
  # branch is a no-op whenever PRINT_ROWS is unset) sees only the verdict.
  if [ "$PRINT_ROWS" = "1" ] && [ -n "$ROWS" ]; then
    printf '%s\n' "$ROWS"
  fi
done

if [ "$TOTAL_VIOLATIONS" -ne 0 ]; then
  exit 1
fi
exit 0
