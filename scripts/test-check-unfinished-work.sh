#!/usr/bin/env bash
# Assertion harness for check-unfinished-work.sh. Builds throwaway worktree
# fixtures under a sandboxed TMPDIR and asserts the guard's verdict and exit
# status. Never touches the real repository tree.
#
# THE PANEL RECORD LIVES AT ITS RENDERED PATH, `docs/superpowers/reviews/
# <YYYY-MM-DD>-<change>-panel.md`, and group 13 below is what holds the guard to
# it. The record used to be hand-written into `.superpowers/sdd/
# final-review-panel.md`, which is now the PASS LOG and carries no marker block
# at all — so a guard that still read it, or that fell back to it when the
# rendered record was absent, would report every change OUTSTANDING at finish
# run 1. Group 13 asserts both halves: the rendered record IS read, and the sdd
# path is NOT, in either direction.
#
# READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the stated
# contract in this repository's myflow-finish-cleanup spec — the requirement
# "Run 1 refuses to integrate silently over unfinished work", which lives in
# this change's delta spec until it is synced there. Never assert against
# observed output. test-check-plan-provenance.sh's header records that suite
# encoding the guard's own defects as its specification more than once, which
# then made each defect look verified.
#
# WHY SOME CASES ASSERT ON THE REASON TEXT. The requirement asks for a
# per-signal BREAKDOWN, not merely the word OUTSTANDING. A prefix assertion
# alone cannot tell "this signal fired" from "some signal fired", so a guard
# that collapsed both signals into one generic sentence would pass every
# prefix case while leaving the operator nothing to act on. The needles below
# are therefore short and behavioural — evidence that the named signal reached
# the breakdown, not a transcript of its prose.
#
# THE FINDINGS CASES (group 4) ARE WRITTEN FROM SIX DEMONSTRATED UNDER-COUNTS,
# not from the code that answers them. Three review passes hid an open Critical
# from the hand-rolled table parser six distinct ways — a capitalised `Open`, a
# status with commentary after it, an unescaped `|` in an earlier cell, a
# reordered column, a row missing its leading pipe that ended table tracking and
# hid every row below it, and a row both detached and malformed that each half
# of the parser left to the other — plus a table inside a blockquote and a
# collation-dependent status comparison. Signal two was then redesigned so
# that none of those questions is asked at all: the panel writes an anchored
# marker line per finding and the guard counts markers. Each case below re-runs
# one demonstration against the new format, so the suite still fails if the
# under-count returns by another route.
#
# Group 4 also asserts the OTHER direction throughout. The record's roster and
# convergence tables, pipes anywhere in a row, sparse and out-of-order
# identifiers, and a marker block written on consecutive lines must all still
# reach `CLEAR:`, and a renamed or reordered header must not change the count —
# a guard that answered OUTSTANDING to everything would satisfy every fail-safe
# case and be just as useless.
#
# THE FINDINGS TABLE CARRIES NO STATUS COLUMN. A finding's state is written once,
# on its marker line. Fixtures follow that format; the two that do carry a stale
# `Status` cell are deliberately malformed records, and they are there to show
# that such a cell reaches no count.
#
# COUNTS, NOT PRESENCE. A needle like "open finding" with no number in front of
# it is satisfied by a guard that stops counting after its first hit, which is
# the under-count the whole signal exists to prevent. Assert the number.
#
# Bash 3.2 is the floor, as test-check-finish-preflight.sh's header records:
# indexed arrays only, no associative arrays.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-unfinished-work.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# An indexed array, not a space-separated string: sandbox paths come from
# mktemp under TMPDIR, which may contain spaces, and word-splitting a string
# would then rm -rf the fragments.
SANDBOXES=()
cleanup() {
  # ${SANDBOXES[@]} is unset-expansion-unsafe under `set -u` on bash 3.2 when empty.
  [ "${#SANDBOXES[@]}" -eq 0 ] && return 0
  for s in "${SANDBOXES[@]}"; do rm -rf "$s"; done
}
trap cleanup EXIT

WORK="$(mktemp -d "${TMPDIR:-/tmp}/unfinished-work-test.XXXXXX")"
SANDBOXES+=("$WORK")
ERRFILE="$WORK/stderr"

# run_guard <worktree> <change-name> -> sets OUT (stdout only), ERR, RC.
# The two streams are captured SEPARATELY rather than merged with 2>&1,
# because the contract distinguishes them: a refusal puts its message on
# stderr and must leave stdout empty, and a merged capture cannot tell an
# empty stdout from a stdout carrying the message.
run_guard() {
  set +e
  OUT="$("$GUARD" "$@" 2>"$ERRFILE")"
  RC=$?
  set -e
  ERR="$(cat "$ERRFILE")"
}

# run_guard_in <locale> <worktree> <change-name> — the same, with LC_ALL set in
# the guard's environment. The guard's verdict must not depend on the collating
# sequence of the session that invoked it: comparing a status by collation made
# a zero-width character in it report OUTSTANDING under a UTF-8 locale and
# vanish under LC_ALL=C, which is one of the six under-counts this suite covers.
run_guard_in() {
  local loc="$1"
  shift
  set +e
  OUT="$(LC_ALL="$loc" "$GUARD" "$@" 2>"$ERRFILE")"
  RC=$?
  set -e
  ERR="$(cat "$ERRFILE")"
}

# assert_verdict <expected-prefix> <label> — exit 0, exactly one stdout line,
# and that line begins with the expected verdict token.
assert_verdict() {
  local want="$1" label="$2" lines
  if [ "$RC" -ne 0 ]; then
    fail "$label: expected exit 0, got rc=$RC out=$OUT err=$ERR"
    return 0
  fi
  if [ -z "$OUT" ]; then
    fail "$label: expected a verdict line, got empty stdout (err=$ERR)"
    return 0
  fi
  lines="$(printf '%s\n' "$OUT" | wc -l | tr -d ' ')"
  if [ "$lines" != "1" ]; then
    fail "$label: expected exactly one stdout line, got $lines: $OUT"
    return 0
  fi
  case "$OUT" in
    "$want"*) pass "$label" ;;
    *) fail "$label: expected a line beginning $want, got: $OUT" ;;
  esac
}

# assert_reason <needle> <label> — the breakdown names this signal.
assert_reason() {
  case "$OUT" in
    *"$1"*) pass "$2" ;;
    *) fail "$2: the breakdown does not name the signal ($1): $OUT" ;;
  esac
}

# REVIEWS is the directory the panel record is RENDERED into, relative to a
# worktree, and SDD_PANEL is the pass log the guard must NOT read. Both are
# written out here rather than derived from the guard's own variables, so that a
# guard that started reading the wrong one fails these cases instead of moving
# the fixtures with it.
REVIEWS='docs/superpowers/reviews'
SDD_PANEL='.superpowers/sdd/final-review-panel.md'

# new_fixture -> sets WT to a worktree holding a fully finished change named
# "demo": every plan item checked, no open finding. PANEL is that fixture's
# rendered panel record, at the dated path the renderer writes and the guard
# resolves.
#
# THE SDD PATH IS STILL CREATED, EMPTY. It is the worktree-lifetime pass log,
# and it exists in every real worktree the guard runs against; leaving it out of
# the fixture would make "the guard does not read it" an untested claim in a
# tree where the file was absent anyway. The cases below that plant content
# there are what turn it into an assertion.
new_fixture() {
  WT="$(mktemp -d "${TMPDIR:-/tmp}/unfinished-work-test.XXXXXX")"
  SANDBOXES+=("$WT")
  mkdir -p "$WT/spectre/changes/demo" "$WT/.superpowers/sdd" "$WT/$REVIEWS"
  PANEL="$WT/$REVIEWS/2026-01-01-demo-panel.md"
  printf -- '- [x] 1.1 done\n' > "$WT/spectre/changes/demo/tasks.md"
  write_panel 1 fixed
}

# write_panel <declared-total> [status ...] — rewrite the fixture's panel record
# with one human table row and one marker line per <status>, numbered F1..Fn.
#
# THE TABLE CARRIES NO STATUS CELL, matching the format: a finding's state is
# written once, on its marker line. A status column beside the marker is a second
# surface that can silently disagree with the line that governs.
#
# <declared-total> is a SEPARATE argument rather than "the number of statuses" so
# that a case can make the checksum disagree on purpose; every other case passes
# the matching number and the helper stays honest about what it wrote.
write_panel() {
  local total="$1" i status
  shift
  {
    printf '| ID | Slot | Severity | Location | Note |\n'
    printf '|---|---|---|---|---|\n'
    # Iterated over "$@" rather than `seq 1 $#`: BSD seq prints "1 0" for an
    # empty list, so `seq` would emit two rows for a panel with no findings.
    i=0
    for status in "$@"; do
      i=$((i + 1))
      printf '| F%d | Lens B | Minor | a.sh:%d | a note |\n' "$i" "$i"
    done
    printf '\nfindings-total: %s\n' "$total"
    i=0
    for status in "$@"; do
      i=$((i + 1))
      printf 'finding-status: F%d %s\n' "$i" "$status"
    done
  } > "$PANEL"
}

# 1. The whole point of the CLEAR verdict: a finished change is not interrupted.
new_fixture
run_guard "$WT" demo
assert_verdict "CLEAR:" "a finished change is CLEAR"

# 1b. The panel's other closed status is closed, not open. Without this case a
#     guard that counted every marker line — or matched `open` anywhere on the
#     line rather than as the status token — still passes case 1, whose fixture
#     holds a single `fixed` marker and nothing resembling the word.
new_fixture
write_panel 3 fixed "withdrawn retracted, the guard already covers it" fixed
run_guard "$WT" demo
assert_verdict "CLEAR:" "fixed and withdrawn findings are closed, not open"

# 2. docs/manual-test/ is not a signal at all, even when a leftover guide is
#    still sitting in the worktree at the moment the guard runs. The earlier
#    version of this case created the directory and then `rm -rf`'d it before
#    calling the guard, so the tree the guard actually saw was identical to
#    case 1's and the assertion proved nothing about the directory at all.
#    This plants the guide and leaves it in place, so the guard's silence on
#    it is observed directly rather than inferred from an already-erased
#    tree.
new_fixture
mkdir -p "$WT/docs/manual-test"
printf -- 'a leftover guide\n' > "$WT/docs/manual-test/guide.md"
run_guard "$WT" demo
assert_verdict "CLEAR:" "a leftover docs/manual-test/ present at invocation time is not a signal"

# 3. Signal one — the plan.
new_fixture
printf -- '- [ ] 1.1 not done\n' > "$WT/spectre/changes/demo/tasks.md"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "an unchecked plan item is OUTSTANDING"
assert_reason "unchecked plan item" "an unchecked plan item names its signal"

# 3b. The same signal in a nested fix sub-change, whose plan is a second file.
#     Case 3 passes with the sub-change lookup deleted outright, so without this
#     case a fix round's unfinished plan reads as CLEAR — the exact silence this
#     guard exists to break. Both layouts the contract leaves open are covered;
#     see the guard's comment on why it reads both.
new_fixture
mkdir -p "$WT/spectre/changes/demo-fix-1"
printf -- '- [ ] 1.1 the fix is not done\n' > "$WT/spectre/changes/demo-fix-1/tasks.md"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "an unchecked item in a sibling fix sub-change is OUTSTANDING"

new_fixture
mkdir -p "$WT/spectre/changes/demo/demo-fix-1"
printf -- '- [ ] 1.1 the fix is not done\n' > "$WT/spectre/changes/demo/demo-fix-1/tasks.md"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "an unchecked item in a nested fix sub-change is OUTSTANDING"

# 3c. A fix OF a fix, at a depth the original three glob shapes did not reach.
#     Nothing in the repository validates that a fix run lands in one of those
#     shapes, so a plan one level deeper contributed no signal at all and the
#     gate reported CLEAR over it. Both directions of the layout are covered:
#     nested inside the change, and beside it.
new_fixture
mkdir -p "$WT/spectre/changes/demo/demo-fix-1/demo-fix-1-fix-2"
printf -- '- [ ] 1.1 the fix of the fix is not done\n' \
  > "$WT/spectre/changes/demo/demo-fix-1/demo-fix-1-fix-2/tasks.md"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a plan two levels deep under the change is OUTSTANDING"
assert_reason "unchecked plan item" "a plan two levels deep names its signal"

new_fixture
mkdir -p "$WT/spectre/changes/demo-fix-1-fix-2"
printf -- '- [ ] 1.1 the sibling fix of the fix is not done\n' \
  > "$WT/spectre/changes/demo-fix-1-fix-2/tasks.md"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a sibling fix-of-a-fix plan is OUTSTANDING"

# 3d. Another change's plan is not this change's signal. Reading every tasks.md
#     under spectre/changes/ would report OUTSTANDING for every unrelated
#     change in flight, which is a guard the operator learns to click past.
new_fixture
mkdir -p "$WT/spectre/changes/other-change"
printf -- '- [ ] 1.1 not our problem\n' > "$WT/spectre/changes/other-change/tasks.md"
run_guard "$WT" demo
assert_verdict "CLEAR:" "another change's unchecked plan is not this change's signal"

# 3e. The PRIMARY plan is not optional, and its absence is not silence. It was
#     read through the same `[ -f ] || continue` as the genuinely-optional fix
#     sub-changes, so a change with no plan directory at all reported
#     "CLEAR: … every plan item is checked" — a claim about a plan that does not
#     exist, and the exact silence-as-clearance this guard exists to break.
new_fixture
rm -rf "$WT/spectre/changes/demo"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a missing primary plan is OUTSTANDING, not CLEAR"
assert_reason "no plan at" "a missing primary plan names its signal"

new_fixture
rm -rf "$WT/spectre"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "no spectre tree at all is OUTSTANDING, not CLEAR"
assert_reason "no plan at" "a missing spectre tree names the plan signal"

# 4. Signal two — findings whose recorded status is not closed.
#
# READ THE GUARD'S HEADER FIRST. Signal two does not parse the human findings
# table any more. The panel writes one anchored `finding-status: F<n> <status>`
# marker line per finding plus one `findings-total:` checksum, and the guard
# counts those. So the cases below come in three kinds:
#
#   - the SIX ways the old table parser under-counted, each re-run against the
#     marker format, where they are now either irrelevant or loud;
#   - the ways the MARKER format itself can be got wrong, every one of which must
#     reach OUTSTANDING;
#   - the opposite direction, so that a guard which answered OUTSTANDING to
#     everything — just as useless — cannot pass the suite.
#
# COUNTS ARE ASSERTED, NOT PRESENCE. A needle like "open finding" with no number
# is satisfied by a guard that stops counting after the first hit, which is the
# under-count this signal exists to prevent. Every countable case names its
# number.

# 4a. One open finding.
new_fixture
write_panel 1 open
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "an open finding is OUTSTANDING"
assert_reason "1 open finding(s)" "an open finding names its signal, with its count"

# 4b. Three open findings among four are three, not one and not four.
new_fixture
write_panel 4 open fixed open open
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "several open findings are OUTSTANDING"
assert_reason "3 open finding(s)" "every open finding is counted, not just the first"

# 4c. Under-count shape 1 — a capitalised `Open`. The marker's status is compared
#     BYTE FOR BYTE against the three legal values, so this is not "not open", it
#     is not a legal status at all, and anything unrecognised is outstanding.
new_fixture
write_panel 1 Open
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a capitalised Open is OUTSTANDING"
assert_reason "1 finding marker(s) whose status is none of" "a capitalised Open names the unrecognised-status signal"

# 4d. Under-count shape 2 — a status with commentary after it.
new_fixture
write_panel 1 "open (needs discussion)"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a status with commentary is OUTSTANDING"
assert_reason "none of open, fixed or withdrawn" "an unrecognised status names its signal"

# 4e. Under-count shape 3 — unescaped `|` characters in the human row. This shifted
#     what the old parser read as the status cell and made the row vanish. A marker
#     line has no cells, so pipes anywhere in the table are now simply text: the
#     open finding is still counted, and the escaping rule is gone.
new_fixture
{
  printf '| ID | Slot | Severity | Location | Note |\n'
  printf '|---|---|---|---|---|\n'
  printf '| F1 | Adversarial | Critical | grep -c foo | bar | hidden behind a pipe | and | another |\n'
  printf '\nfindings-total: 1\nfinding-status: F1 open\n'
} > "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "unescaped pipes in a row cannot hide its finding"
assert_reason "1 open finding(s)" "a row full of pipes is still one counted finding"

# 4f. Under-count shape 4 — a reordered, renamed or entirely absent header row.
#     The old parser found the table by its header and counted nothing when it did
#     not recognise it. The marker block does not depend on the header at all: the
#     columns here are renamed and reordered, and there is a retired `Status`
#     column carrying a stale value, none of which reaches the count.
new_fixture
{
  printf '| Ref | Where | How bad | Raised by | Status | Remark |\n'
  printf '|---|---|---|---|---|---|\n'
  printf '| F1 | app.py:12 | Critical | Security | fixed | a stale cell from the retired format |\n'
  printf '\nfindings-total: 1\nfinding-status: F1 open\n'
} > "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a reordered header cannot hide a finding"
assert_reason "1 open finding(s)" "column order no longer decides whether a finding counts"

# 4g. Under-count shape 5 — a row missing its LEADING pipe, which is valid GFM and
#     once ended table tracking, hiding that row and every row below it. Both rows
#     must still be counted, from their markers, and the count is asserted.
new_fixture
{
  printf '| ID | Slot | Severity | Location | Note |\n'
  printf '|---|---|---|---|---|\n'
  printf 'F1 | Security | Critical | app.py:12 | no leading pipe |\n'
  printf '| F2 | Lens C | Important | b.sh:3 | and everything below it |\n'
  printf '\nfindings-total: 2\nfinding-status: F1 open\nfinding-status: F2 open\n'
} > "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a row missing its leading pipe is OUTSTANDING"
assert_reason "2 open finding(s)" "no finding below a pipe-less row is lost"

# 4h. Under-count shape 6 — a row detached from the table by a blank line and a
#     heading, and a stray prose line inside the table. The old parser needed an
#     orphan net for the first and an unparseable count for the second; neither
#     concept exists now, and both records read exactly as they should.
new_fixture
{
  printf '| ID | Slot | Severity | Location | Note |\n'
  printf '|---|---|---|---|---|\n'
  printf 'this prose sneaked into the table\n'
  printf '| F1 | Lens C | Important | b.sh:3 | below the prose |\n'
  printf '\n## Appendix\n\n'
  printf '| F2 | Security | Critical | app.py:12 | parked under a heading |\n'
  printf '\nfindings-total: 2\nfinding-status: F1 fixed\nfinding-status: F2 open\n'
} > "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a detached row and a stray line change nothing"
assert_reason "1 open finding(s)" "a detached row is counted once, from its marker"

# 4i. A findings table written inside a GFM blockquote. Once a legitimate table
#     existed elsewhere the old parser could not see this one at all. The rows no
#     longer read as findings rows, so the table and the marker block disagree and
#     the record is outstanding — loud where it used to be invisible.
new_fixture
{
  printf '> | ID | Slot | Severity | Location | Note |\n'
  printf '> |---|---|---|---|---|\n'
  printf '> | F1 | Security | Critical | app.py:12 | quoted away |\n'
  printf '\nfindings-total: 1\nfinding-status: F1 open\n'
} > "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a blockquoted findings table is OUTSTANDING"
assert_reason "do not name the same findings" "a blockquoted table names the correspondence signal"

# 4j. A zero-width character inside the status. `==` in awk compares by the
#     locale's collating sequence, so this reported OUTSTANDING under a UTF-8
#     locale and vanished under LC_ALL=C. The answer must be the same under both,
#     and it must be OUTSTANDING — the guard compares bytes.
for loc in C en_US.UTF-8; do
  new_fixture
  {
    printf '| ID | Slot | Severity | Location | Note |\n'
    printf '|---|---|---|---|---|\n'
    printf '| F1 | Security | Critical | app.py:12 | zero width inside |\n'
    printf '\nfindings-total: 1\n'
    printf 'finding-status: F1 fi\xe2\x80\x8bxed\n'
  } > "$PANEL"
  run_guard_in "$loc" "$WT" demo
  assert_verdict "OUTSTANDING:" "a zero-width character in the status is OUTSTANDING (LC_ALL=$loc)"
  assert_reason "none of open, fixed or withdrawn" "a zero-width status names its signal (LC_ALL=$loc)"
done

# 4k. A marker that does not begin its line — indented, or inside a blockquote, or
#     written without the `F<n> ` identifier. Each of these is a line that NAMES
#     `finding-status:` without being one, and each is reported rather than
#     silently not counted. The count is asserted.
new_fixture
{
  printf '| ID | Slot | Severity | Location | Note |\n'
  printf '|---|---|---|---|---|\n'
  printf '| F1 | Security | Critical | app.py:12 | one |\n'
  printf '| F2 | Lens C | Important | b.sh:3 | two |\n'
  printf '| F3 | Bugbot | Minor | c.sh:7 | three |\n'
  printf '\nfindings-total: 3\n'
  printf '  finding-status: F1 open\n'
  printf '> finding-status: F2 open\n'
  printf 'finding-status: open\n'
} > "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "markers that do not begin their line are OUTSTANDING"
assert_reason "3 line(s) naming finding-status:" "every malformed marker line is counted"

# 4l. A withdrawal states its reason on the marker line, because the reason is part
#     of the state the guard checks — and one with no reason is outstanding, which
#     is the pass-1 rule re-expressed for markers.
new_fixture
write_panel 1 "withdrawn the operator retracted it: the guard already covers this"
run_guard "$WT" demo
assert_verdict "CLEAR:" "a withdrawal with a reason is closed"

new_fixture
write_panel 1 withdrawn
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a withdrawal with no stated reason is OUTSTANDING"
assert_reason "1 withdrawn finding(s) with no reason" "an unreasoned withdrawal names its signal"

# 4m. Case is not folded. `WITHDRAWN` is not one of the three legal values, and
#     unlike the old table — where an unrecognised status could fall through to
#     "not open" — anything unrecognised here is outstanding.
new_fixture
write_panel 1 "WITHDRAWN retracted with a reason"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "an upper-case status is not a legal status"
assert_reason "none of open, fixed or withdrawn" "an upper-case status names its signal"

# 4n. A record that never speaks cannot show zero findings. A panel record with no
#     marker block and no declared total is outstanding however clean its prose.
new_fixture
printf '# Panel\n\nEverything was fine, honestly.\n' \
  > "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a panel record with no marker block is OUTSTANDING"
assert_reason "does not declare a findings total" "a silent panel record names its signal"

# 4o. A record that DOES speak, and says zero. This is the one way a panel that
#     raised nothing clears the gate, and it is a positive declaration rather than
#     an absence — which is exactly the distinction case 4n turns on.
new_fixture
printf '# Panel\n\nNo slot raised a finding.\n\nfindings-total: 0\n' \
  > "$PANEL"
run_guard "$WT" demo
assert_verdict "CLEAR:" "a record that declares zero findings is CLEAR"

# 4p. Markers with no declared total, a total declared twice, and a total that is
#     not a plain count. Each leaves the checksum unusable, and an unusable
#     checksum is outstanding rather than ignored.
new_fixture
write_panel 1 fixed
grep -v '^findings-total:' "$PANEL" > "$WT/panel.tmp"
mv "$WT/panel.tmp" "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "markers with no declared total are OUTSTANDING"
assert_reason "does not declare a findings total" "a missing total names its signal"

new_fixture
write_panel 1 fixed
printf 'findings-total: 1\n' >> "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a total declared twice is OUTSTANDING"
assert_reason "more than once" "a duplicated total names its signal"

new_fixture
write_panel "one" fixed
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a total that is not a plain count is OUTSTANDING"
assert_reason "not a plain count" "a malformed total names its signal"

# 4p-i. A findings-total: digit run long enough to overflow shell arithmetic
#     (26 digits) is bounded out at the pattern, the same bound
#     check-panel-reproducers.sh already applies to reproducers-total. It must
#     be reported as malformed — "not a plain count" — never read as a total
#     and held against the marker count, which is the disagreement message a
#     shorter, well-formed mismatch gets.
new_fixture
write_panel "12345678901234567890123456" fixed
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "an oversized total is OUTSTANDING at exit 0, not a refusal at exit 2"
assert_reason "not a plain count" "an oversized total is reported malformed, not compared as a total"

# 4q. THE DRIFT THIS DESIGN MUST NOT HIDE: a row written with no marker beside it.
#     The declared total still counts the finding, so the checksum disagrees with
#     the marker block and the record is outstanding. Both numbers are named, so
#     the operator is told which one to fix.
new_fixture
write_panel 3 open fixed
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a declared total above the marker count is OUTSTANDING"
assert_reason "declares findings-total: 3" "the disagreement names the declared total"
assert_reason "2 finding-status: marker line(s)" "the disagreement names the marker count"

new_fixture
write_panel 1 open fixed
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a declared total below the marker count is OUTSTANDING"
assert_reason "declares findings-total: 1" "the low disagreement names the declared total"

# 4r. The same drift caught from the other side: the human table and the marker
#     block must name the SAME findings. A row with no marker, a marker with no
#     row, and a duplicated identifier are all correspondence failures — the last
#     one matters because counts alone would balance.
new_fixture
{
  printf '| ID | Slot | Severity | Location | Note |\n'
  printf '|---|---|---|---|---|\n'
  printf '| F1 | Lens B | Minor | a.sh:1 | one |\n'
  printf '| F2 | Lens C | Critical | b.sh:2 | two, and nobody wrote its marker |\n'
  printf '\nfindings-total: 2\n'
  printf 'finding-status: F1 fixed\n'
  printf 'finding-status: F1 fixed\n'
} > "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a duplicated marker identifier is OUTSTANDING"
assert_reason "do not name the same findings" "a duplicated identifier names the correspondence signal"

new_fixture
write_panel 2 fixed fixed
printf '| F9 | Lens C | Minor | z.sh:1 | fixed | a row nobody marked |\n' \
  >> "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a findings row with no marker is OUTSTANDING"
assert_reason "do not name the same findings" "an unmarked row names the correspondence signal"

# 4r-i. AN IDENTIFIER REUSED SYMMETRICALLY, which the correspondence check alone
#     cannot see. It compares the two identifier lists as sorted MULTISETS, so two
#     genuinely distinct findings both labelled `F1` — against the spec's own
#     "unique within the record" — with `F1` also written twice in the marker
#     block leave the duplicate matched on both sides and the check passing
#     vacuously. Demonstrated: an open Critical read as `CLEAR:`, and the word
#     `open` never appears in a marker, so the open count cannot help either.
#     Uniqueness is therefore asserted on each side independently.
new_fixture
{
  printf '| ID | Slot | Severity | Location | Note |\n'
  printf '|---|---|---|---|---|\n'
  printf '| F1 | Lens B | Minor | a.sh:1 | genuinely fixed finding |\n'
  printf '| F1 | Security | Critical | app.py:12 | genuinely OPEN finding, same id |\n'
  printf '\nfindings-total: 2\n'
  printf 'finding-status: F1 fixed\n'
  printf 'finding-status: F1 fixed\n'
} > "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "an identifier reused on both sides is OUTSTANDING"
assert_reason "reuses identifier(s) F1" "a reused identifier is named"

# 4r-ii. Each side is checked on its own, so neither can be excused by the other
#     carrying the same duplicate. The marker block alone, then the table alone.
new_fixture
write_panel 3 fixed fixed fixed
grep -v '^finding-status: F3 ' "$PANEL" > "$WT/panel.tmp"
{ cat "$WT/panel.tmp"; printf 'finding-status: F1 fixed\n'; } \
  > "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "an identifier reused in the marker block alone is OUTSTANDING"
assert_reason "marker block reuses identifier(s) F1" "the marker block's duplicate is named"

new_fixture
write_panel 2 fixed fixed
printf '| F1 | Lens C | Minor | z.sh:1 | a second row reusing F1 |\n' \
  >> "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "an identifier reused in the table alone is OUTSTANDING"
assert_reason "findings table reuses identifier(s) F1" "the table's duplicate is named"

# 4r-iii. The opposite direction: a record whose identifiers are all distinct must
#     not be reported, including one whose numbering is sparse and out of order —
#     the rule is uniqueness, not a contiguous sequence starting at one.
new_fixture
{
  printf '| ID | Slot | Severity | Location | Note |\n'
  printf '|---|---|---|---|---|\n'
  printf '| F7 | Lens B | Minor | a.sh:1 | seven |\n'
  printf '| F2 | Lens C | Minor | b.sh:2 | two |\n'
  printf '| F13 | Bugbot | Minor | c.sh:3 | thirteen |\n'
  printf '\nfindings-total: 3\n'
  printf 'finding-status: F13 fixed\n'
  printf 'finding-status: F7 fixed\n'
  printf 'finding-status: F2 fixed\n'
} > "$PANEL"
run_guard "$WT" demo
assert_verdict "CLEAR:" "sparse, out-of-order but distinct identifiers are CLEAR"

# 4s. A marker quoted inside a fenced block in the record is still counted — this
#     signal has no notion of a block, which is the point. The count then exceeds
#     the declared total and the record reports outstanding, which is the safe
#     direction.
new_fixture
write_panel 1 fixed
printf '\nWrite them like this:\n\n```\nfinding-status: F7 fixed\n```\n' \
  >> "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a marker quoted inside a fence is counted, not ignored"

# 4s-0. THE ONE ROUTE THAT STILL UNDER-COUNTED, found by attacking the redesign
#     rather than by reviewing it. Because a fenced marker is counted like any
#     other, a fenced `finding-status: F2 fixed` STOOD IN FOR F2's real marker
#     when that marker was never written: the identifiers matched, the checksum
#     matched, and an open finding read as `CLEAR:`. The marker lines must be one
#     unbroken block, which the fence's own delimiter line breaks — closing the
#     route without teaching this signal about fences, since block tracking is
#     exactly what the redesign deleted.
new_fixture
{
  printf '| ID | Slot | Severity | Location | Note |\n'
  printf '|---|---|---|---|---|\n'
  printf '| F1 | Lens B | Minor | a.sh:1 | one |\n'
  printf '| F2 | Security | Critical | app.py:12 | its marker was never written |\n'
  printf '\nfindings-total: 2\n'
  printf 'finding-status: F1 fixed\n'
  printf '```\nfinding-status: F2 fixed\n```\n'
} > "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a fenced marker cannot stand in for a missing one"
assert_reason "must be one unbroken block" "a broken marker block names its signal"

# 4s-0b. The opposite direction, so the fix is not "any second block is fatal": a
#     correct record's markers sit on consecutive lines and reach CLEAR, and it is
#     specifically a marker parked away from the block that is reported.
new_fixture
write_panel 3 fixed fixed fixed
run_guard "$WT" demo
assert_verdict "CLEAR:" "three markers on consecutive lines are CLEAR"

new_fixture
write_panel 3 fixed fixed fixed
grep -v '^finding-status: F2 ' "$PANEL" > "$WT/panel.tmp"
{ cat "$WT/panel.tmp"; printf '\nfinding-status: F2 fixed\n'; } \
  > "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a marker parked away from the block is OUTSTANDING"
assert_reason "must be one unbroken block" "a detached marker names the contiguity signal"

# 4s-i. FOUND BY ATTACKING THE REDESIGN, not by reading it. A record written with
#     CRLF line endings carried the declared total as `2<CR>`, which `read -r`
#     does not split off; `[ … -ne … ]` then exited 2 with "integer expression
#     expected", and because that sits inside an `if`, errexit never saw it and
#     THE CHECKSUM SILENTLY STOPPED RUNNING. Both directions are asserted, so a
#     fix that simply ignores the total cannot pass.
new_fixture
printf '| F1 |\r\n| F2 |\r\n\r\nfindings-total: 2\r\nfinding-status: F1 fixed\r\nfinding-status: F2 fixed\r\n' \
  > "$PANEL"
run_guard "$WT" demo
assert_verdict "CLEAR:" "a CRLF record with a correct checksum is CLEAR"

new_fixture
printf '| F1 |\r\n| F2 |\r\n\r\nfindings-total: 3\r\nfinding-status: F1 fixed\r\nfinding-status: F2 fixed\r\n' \
  > "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "the checksum still runs on a CRLF record"
assert_reason "declares findings-total: 3" "the CRLF checksum names the declared total"

# 4s-ii. The same class by another route: one stray NUL byte puts grep into
#     binary mode, where the two checks that read TEXT — the declared total and
#     the identifier lists — got `Binary file … matches` instead of the record's
#     content and compared it against itself. An open finding in such a record
#     must still be reported.
new_fixture
{
  printf '| F1 |\n| F2 |\n\nfindings-total: 2\n'
  printf 'finding-status: F1 open\n'
  printf '\x00stray byte from a bad merge\n'
  printf 'finding-status: F2 open\n'
} > "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a record with a NUL byte is still counted"
assert_reason "2 open finding(s)" "a NUL byte does not stop the open findings being counted"

# 4s-iii. The binary-mode gap in its OTHER half, added because mutation-testing
#     the suite found it: without `-a`, `grep -o` on a record carrying a NUL byte
#     reports "no match" for BOTH identifier lists, they compare equal, and the
#     table-against-markers check passes vacuously. Case 4s-ii could not see that
#     — it asserts the open count, which comes from `grep -c` and survives binary
#     mode — so this fixture makes the two lists genuinely disagree.
new_fixture
{
  printf '| ID | Slot | Severity | Location | Note |\n'
  printf '|---|---|---|---|---|\n'
  printf '| F1 | Lens B | Minor | a.sh:1 | one |\n'
  printf '| F2 | Security | Critical | app.py:12 | its marker was never written |\n'
  printf '\x00stray byte from a bad merge\n'
  printf '\nfindings-total: 1\n'
  printf 'finding-status: F1 fixed\n'
} > "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "the identifier check still runs on a record with a NUL byte"
assert_reason "do not name the same findings" "a NUL byte does not void the identifier check"

# 4t. The opposite direction, so a guard that answered OUTSTANDING to everything
#     cannot pass: the record's OTHER tables — the roster, the convergence table —
#     carry rows and pipes and even the word `open`, and none of it is a finding.
new_fixture
{
  printf '\n## Roster\n\n'
  printf '| # | Slot | Required | Included | Trigger |\n'
  printf '|---|---|---|---|---|\n'
  printf '| 0 | Primary | always | yes | — |\n'
  printf '| 3 | Security | conditional | yes | path handling, still open questions |\n'
  printf '\n## Convergence\n\n'
  printf '| Defect | Slots |\n|---|---|\n| Parser fails open | Security, Adversarial |\n'
} >> "$PANEL"
run_guard "$WT" demo
assert_verdict "CLEAR:" "the record's other tables are not read as findings"


# 5. Both surviving signals firing at once must be counted independently and
#    reported together on the one line — the guard's own header says so, and
#    this is the direct demonstration: an unchecked plan item and an open
#    finding, both named in the single OUTSTANDING line.
new_fixture
printf -- '- [ ] 1.1 not done\n' > "$WT/spectre/changes/demo/tasks.md"
write_panel 1 open
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "both signals firing at once is OUTSTANDING"
assert_reason "unchecked plan item" "the combined line names the plan signal"
assert_reason "1 open finding(s)" "the combined line also names the findings signal"

# 5b. A file that EXISTS but cannot be READ is neither signal — it is an honest
#     unknown, and the guard refuses rather than guessing. Demonstrated against
#     the primary plan, read through count_unticked -> count_matching and
#     nothing else: chmod 000 makes `grep` fail with rc > 1, which unreadable()
#     and count_matching's `rc > 1` discipline turn into a refusal (exit 2,
#     nothing on stdout, the reason named on stderr) rather than a silent "zero
#     matches" that would read as clean. Targeting tasks.md rather than the
#     panel record keeps this case specific to count_matching's own discipline:
#     the panel record is also re-read by ids_of, whose separate `rc > 1` guard
#     would mask a swallowed count_matching failure. Skipped when running as
#     root, where chmod 000 does not block reads.
if [ "$(id -u)" -ne 0 ]; then
  new_fixture
  chmod 000 "$WT/spectre/changes/demo/tasks.md"
  run_guard "$WT" demo
  chmod 644 "$WT/spectre/changes/demo/tasks.md"
  [ "$RC" -ne 0 ] && pass "an unreadable tasks.md exits non-zero" \
    || fail "unreadable tasks.md: expected a non-zero exit, got rc=$RC out=$OUT"
  [ -z "$OUT" ] && pass "an unreadable tasks.md emits no verdict line" \
    || fail "unreadable tasks.md: emitted a verdict line: $OUT"
  case "$ERR" in
    *"check-unfinished-work: cannot read"*) \
      pass "an unreadable tasks.md names the failure on stderr" ;;
    *) fail "unreadable tasks.md: no named message on stderr: $ERR" ;;
  esac
fi

# 6. A missing file is outstanding, never clear. This is the failure the whole
#    requirement exists to prevent: silence is not clearance.
new_fixture
rm "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a missing panel record is OUTSTANDING, not CLEAR"
assert_reason "no review panel record" "a missing panel record names its signal"

# 8e. The change name is PR-controlled — it reaches here from a state file
#     anyone able to open a pull request can edit — and it is concatenated into
#     every path above. The allowlist is records.Destination's Protection 1
#     (stats/internal/records/render.go), and these cases are the same shapes
#     test-check-cleanup-complete.sh rejects, asserted here so the two copies
#     cannot drift apart in silence. Without it `../../../planted/clear` reads a plan outside the
#     worktree entirely and reports CLEAR for a change that has none.
for bad_name in "../../../planted/clear" "demo*" "demo/../demo" ".hidden" "demo?x"; do
  new_fixture
  run_guard "$WT" "$bad_name"
  [ "$RC" -eq 2 ] && pass "a change name outside the allowlist ($bad_name) -> exit 2" \
    || fail "change name $bad_name: expected exit 2, got rc=$RC out=$OUT"
  [ -z "$OUT" ] && pass "a change name outside the allowlist ($bad_name) emits no verdict line" \
    || fail "change name $bad_name: emitted a verdict line: $OUT"
  case "$ERR" in
    *"check-unfinished-work: change name"*) pass "a rejected change name ($bad_name) names the failure" ;;
    *) fail "change name $bad_name: no named message on stderr: $ERR" ;;
  esac
done

# 8e-i. THE ALLOWLIST MUST NOT DEPEND ON THE CALLER'S LOCALE. `case`'s `A-Za-z0-9`
#     ranges are collating ranges, not byte ranges: measured on bash 3.2, the name
#     `démo` is REJECTED under LC_ALL=C and ACCEPTED under en_US.UTF-8. A guard
#     whose accepted input set changes with the operator's environment is the same
#     defect class as a status compared by collation, and this copy of the
#     allowlist would then diverge from check-cleanup-complete.sh's under one
#     locale and not the other. The guard pins the locale for its whole run; this
#     case is what proves the pin is load-bearing rather than decorative.
for loc in C en_US.UTF-8; do
  new_fixture
  run_guard_in "$loc" "$WT" "démo"
  [ "$RC" -eq 2 ] && pass "a non-ASCII change name is rejected under LC_ALL=$loc" \
    || fail "change name démo under LC_ALL=$loc: expected exit 2, got rc=$RC out=$OUT"
  [ -z "$OUT" ] && pass "a non-ASCII change name emits no verdict line under LC_ALL=$loc" \
    || fail "change name démo under LC_ALL=$loc: emitted a verdict line: $OUT"
done

# 8f. Traversal, demonstrated end to end rather than asserted from the message.
#     A planted "finished change" outside the worktree must not be readable
#     through the name, whatever the guard's error text says. The allowlist at
#     check-unfinished-work.sh:112-118 rejects any name containing "/", so it
#     is the WHOLE containment mechanism: no name that passes it can traverse
#     at all. That means the name here must be BOTH relative (an absolute
#     "$PLANTED/..." name fails for the wrong reason — the allowlist rejects
#     it on "/" the same as anything else, so a case built that way proves
#     nothing about traversal specifically) and shaped to genuinely resolve
#     onto the planted tree once concatenated, so that containment failing
#     here would produce a real CLEAR rather than a path-concatenation
#     OUTSTANDING. NAME is therefore relative: $WT and $PLANTED are both
#     mktemp -d siblings directly under the same TMPDIR, so
#     "../../../$(basename "$PLANTED")/spectre/changes/clear", concatenated
#     onto "$WT/spectre/changes/" by the guard, lands exactly on
#     "$PLANTED/spectre/changes/clear" — a change with a fully-ticked
#     tasks.md. The panel record is read from $WT's own fixture (already
#     CLEAR via new_fixture's write_panel), never from $PLANTED: the change
#     name reaches the panel path only as a filename component matched
#     literally against the entries of $WORKTREE's own reviews directory
#     (panel_record_path), and no name containing "/" can equal a directory
#     entry's basename — so a second panel record planted under $PLANTED would
#     be unreachable by construction and is not planted here.
#
#     Proved by mutation, not merely asserted: temporarily deleting the
#     allowlist `case` block from check-unfinished-work.sh and re-running this
#     suite turns this case's exit 2 into an exit 0 carrying a verdict line
#     built from a plan read OUTSIDE the worktree — confirming the allowlist
#     is what stops it, not the path shape. Both assertions below fail on that
#     mutation: the plan signal is the half the traversal reaches, so the
#     verdict is OUTSTANDING rather than CLEAR once the panel path stops
#     resolving too, and a verdict at all is the failure.
new_fixture
PLANTED="$(mktemp -d "${TMPDIR:-/tmp}/unfinished-work-planted.XXXXXX")"
SANDBOXES+=("$PLANTED")
mkdir -p "$PLANTED/spectre/changes/clear"
printf -- '- [x] 1.1 done\n' > "$PLANTED/spectre/changes/clear/tasks.md"
run_guard "$WT" "../../../$(basename "$PLANTED")/spectre/changes/clear"
[ "$RC" -eq 2 ] && pass "a traversal-shaped name is rejected by the allowlist before any path is built" \
  || fail "traversal: expected exit 2 from the allowlist, got rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "a traversal-shaped name reads nothing outside the worktree" \
  || fail "traversal: the guard produced a verdict from outside the worktree: $OUT"

# 9. A worktree that cannot be read is not a verdict: exit non-zero, nothing on
#    stdout, and the reason on stderr — so a caller grepping stdout for CLEAR
#    cannot read clearance out of a failure.
run_guard "/nonexistent/worktree" demo
[ "$RC" -ne 0 ] && pass "an unreadable worktree exits non-zero" \
  || fail "unreadable worktree: expected a non-zero exit, got rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "an unreadable worktree writes nothing to stdout" \
  || fail "unreadable worktree: emitted a verdict line: $OUT"
# The needle carries the colon deliberately. Without it the shell's own
# "…/check-unfinished-work.sh: No such file or directory" satisfies the case,
# so it passes while the guard does not exist — a vacuous assertion that
# proves nothing about the guard's own reporting.
case "$ERR" in
  *"check-unfinished-work: "*) pass "an unreadable worktree names the failure on stderr" ;;
  *) fail "unreadable worktree: no named message on stderr: $ERR" ;;
esac

# 10. A missing argument is programmer error, reported rather than guessed at.
run_guard "" ""
[ "$RC" -eq 2 ] && pass "missing arguments -> exit 2" \
  || fail "missing arguments: expected exit 2, got rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "missing arguments: emits no verdict line" \
  || fail "missing arguments: emitted a verdict line: $OUT"

# 11. A missing library is a refusal (spec scenario, F4): a copy of this guard
#    with no lib/panel-record.sh beside it cannot source the marker helpers
#    and must exit 2, naming the problem, rather than continuing with a
#    reduced set of checks. The copy is made in a throwaway sandbox — never
#    the real scripts/lib/panel-record.sh, which is not touched — so a
#    failure partway through this case cannot corrupt the working tree.
NOLIB_DIR="$(mktemp -d "${TMPDIR:-/tmp}/unfinished-work-nolib-test.XXXXXX")"
SANDBOXES+=("$NOLIB_DIR")
cp "$GUARD" "$NOLIB_DIR/check-unfinished-work.sh"
chmod +x "$NOLIB_DIR/check-unfinished-work.sh"
set +e
OUT="$("$NOLIB_DIR/check-unfinished-work.sh" "$WORK" demo 2>"$ERRFILE")"
RC=$?
set -e
ERR="$(cat "$ERRFILE")"
[ "$RC" -eq 2 ] && pass "a guard copy with no lib/panel-record.sh beside it exits 2" \
  || fail "missing library: expected exit 2, got rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "a guard copy with no lib/panel-record.sh beside it emits no verdict line" \
  || fail "missing library: emitted a verdict line: $OUT"
case "$ERR" in
  *"panel-record.sh"*) pass "a guard copy with no lib/panel-record.sh beside it names the missing library on stderr" ;;
  *) fail "missing library: stderr does not name panel-record.sh: $ERR" ;;
esac

# 12. A fix round's fix-mutation: lines are inert to this guard (KAN-209).
#     A fix round records what it mutated in the pass log entry, which is now a
#     SEPARATE file this guard does not read at all — case 13b covers that
#     separation directly. These cases keep asserting the harder direction: the
#     lines are inert even written into the record itself, where they are placed
#     OUTSIDE the marker block on purpose, because one between two markers would
#     split the unbroken run cases 4s-0 and 4s-0b assert on, and one shaped like
#     a table row would enter the row-identifier set the 4q/4r subgroup compares
#     against the markers. Asserting them here rather than only in the pass log
#     keeps the guard's own inertness a checked property rather than a
#     consequence of which file the lines happen to be in.
#
#     The CLEAR fixture below carries TWO closed markers, not one: with only
#     one marker there is no "between two markers" for a misplaced line to
#     occupy, and the contiguity mechanism could never be exercised. Two
#     markers make both mutations structurally possible against the same
#     CLEAR: assertion — one splits the unbroken run, the other unbalances
#     the row/marker identifier comparison.
add_pass_log() {
  cat >> "$PANEL" <<'PASSLOG'

### Pass 2 — targeted

fix-mutation: scripts/check-thing.sh — reverted the unit-separator delimiter to a tab — scripts/test-check-thing.sh case 4 failed
fix-mutation: skills/myflow-do/SKILL.md — none — prose only
fix-mutations-total: 2
PASSLOG
}

new_fixture
write_panel 2 fixed fixed
add_pass_log
run_guard "$WT" demo
assert_verdict "CLEAR:" "fix-mutation: lines leave a clear record clear"

new_fixture
write_panel 2 open fixed
add_pass_log
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "fix-mutation: lines do not hide an open finding"

# 12a. THE PLACEMENT RULE ITSELF, PINNED (KAN-209 fix round, F4): a
#     fix-mutation: line written BETWEEN two finding-status: markers — rather
#     than in the pass log entry where the format requires it — splits the
#     unbroken run signal two names, so this must read OUTSTANDING naming that
#     signal specifically, not just any breakdown line.
new_fixture
{
  printf '| ID | Slot | Severity | Location | Note |\n'
  printf '|---|---|---|---|---|\n'
  printf '| F1 | Lens B | Minor | a.sh:1 | one |\n'
  printf '| F2 | Lens B | Minor | a.sh:2 | two |\n'
  printf '\nfindings-total: 2\n'
  printf 'finding-status: F1 fixed\n'
  printf 'fix-mutation: scripts/check-thing.sh — reverted the delimiter — case 4 failed\n'
  printf 'finding-status: F2 fixed\n'
} > "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a fix-mutation: line between two markers is OUTSTANDING"
assert_reason "must be one unbroken block" "a fix-mutation: line splitting the marker block names the contiguity signal"

# 12b. THE OTHER HALF OF THE SAME PLACEMENT RULE, PINNED: a fix-mutation: line
#     shaped like a findings-table row — matching the row-identifier pattern
#     `^\|?[[:space:]]*F[0-9]+[[:space:]]*\|` — enters the row-identifier set
#     the 4q/4r subgroup compares against the markers, with no marker of its
#     own to balance it, so this must read OUTSTANDING naming the
#     row/marker correspondence signal specifically.
new_fixture
write_panel 2 fixed fixed
{
  printf '\n### Pass 2 — targeted\n\n'
  printf '| F9 | scripts/check-thing.sh | reverted the delimiter | case 4 failed |\n'
  printf 'fix-mutations-total: 1\n'
} >> "$PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a fix-mutation: line shaped like a table row is OUTSTANDING"
assert_reason "do not name the same findings" "a row-shaped fix-mutation: line names the row/marker correspondence signal"

# --- 13. THE RECORD IS READ WHERE IT IS RENDERED, AND ONLY THERE ---
#
# Every case above already proves the rendered record is read: new_fixture
# writes nothing anywhere else. This group proves the other half — that the
# guard does not ALSO read, or fall back to, the pass log at
# .superpowers/sdd/final-review-panel.md — and that the rendered record is
# resolved by an ANCHORED match rather than by a loose glob.
#
# Without 13a and 13b a guard that silently fell back to the sdd path would pass
# this whole suite while reading a file nothing writes: every change would read
# OUTSTANDING at finish run 1, and no case here would say so.

# 13a. A perfectly clear record at the OLD path, and nothing at the new one, is
#     no record at all. This is the fallback, asserted directly.
new_fixture
mv "$PANEL" "$WT/$SDD_PANEL"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "a record at the old sdd path only is OUTSTANDING"
assert_reason "no review panel record" "a record at the old sdd path only names the missing-record signal"

# 13b. The other direction, which a fallback alone would not catch: a guard that
#     read BOTH files and unioned them would pass 13a and still let the pass log
#     contribute findings. An OPEN finding written into the sdd path, beside a
#     CLEAR rendered record, must count for nothing.
new_fixture
write_panel 1 fixed
{
  printf '| ID | Slot | Severity | Location | Note |\n'
  printf '|---|---|---|---|---|\n'
  printf '| F7 | Lens B | Critical | a.sh:7 | still open in the pass log |\n'
  printf '\nfindings-total: 1\n'
  printf 'finding-status: F7 open\n'
} > "$WT/$SDD_PANEL"
run_guard "$WT" demo
assert_verdict "CLEAR:" "an open finding in the pass log does not reach the verdict"

# 13c. THE MATCH IS ANCHORED ON THE CHANGE NAME. `*-demo-panel.md` also matches
#     `2026-01-01-other-demo-panel.md`, another change's record sitting in the
#     same directory — reading it would answer this change's question with that
#     change's findings. The renderer's own comment records the write-side half
#     of this incident, where a loose match overwrote a different change's
#     record.
new_fixture
mv "$PANEL" "$WT/$REVIEWS/2026-01-01-other-demo-panel.md"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "another change's dated record is not this change's record"
assert_reason "no review panel record" "another change's dated record names the missing-record signal"

# 13d. AND ANCHORED ON THE DATE. A file named `demo-panel.md` with no date
#     prefix is not a rendered record — the renderer never writes one — and a
#     match that accepted it would accept whatever else a pull request put in
#     that directory under that name.
new_fixture
mv "$PANEL" "$WT/$REVIEWS/demo-panel.md"
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "an undated demo-panel.md is not a rendered record"

# 13e. WHICH DATED FILE IS READ IS DETERMINISTIC, and it is the same one
#     records.existingDatedFile returns: the earliest, its `sort.Strings` on the
#     matching names followed by `found[0]`. A change has one rendered record —
#     the date is fixed at the first render and a fix round overwrites in place
#     — so two of them means something already went wrong, and the guard reading
#     a DIFFERENT one from the renderer would hide an open finding behind a file
#     nothing updates. The open finding is written to the earlier date, so a
#     guard that took the later one reads CLEAR.
new_fixture
write_panel 1 open
mv "$PANEL" "$WT/$REVIEWS/2026-01-01-demo-panel.md"
PANEL="$WT/$REVIEWS/2026-02-02-demo-panel.md"
write_panel 1 fixed
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "the earliest dated record is the one read"
assert_reason "1 open finding" "the earliest dated record's open finding reaches the verdict"

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'check-unfinished-work: all cases pass\n'
