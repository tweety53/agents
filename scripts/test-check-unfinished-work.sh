#!/usr/bin/env bash
# Assertion harness for check-unfinished-work.sh. Builds throwaway worktree
# fixtures under a sandboxed TMPDIR and asserts the guard's verdict and exit
# status. Never touches the real repository tree.
#
# SIGNAL TWO NOW READS THE STORE, NOT RENDERED MARKDOWN. Every signal-two case
# below builds a worktree-shaped sandbox and a stub `flow` binary placed
# ahead of the real one on PATH inside that sandbox: the stub answers
# `record findings -change <name> [-C <dir>]` with a canned JSON array (the
# shape `flow record findings` itself prints -- one object per finding with
# `ref`, `status`, `reproducer`), or exits non-zero to simulate a store the
# guard could not reach. No case writes a docs/superpowers/reviews/*-panel.md
# file any more; that path, and the marker-line grammar signal two used to
# parse out of it, are retired by this rewrite. SIGNAL ONE (the tasks.md
# checkbox count) AND ITS FIXTURES ARE COMPLETELY UNTOUCHED -- it never read
# the panel record at all.
#
# This task (task 7 of KAN-271) does NOT touch check-unfinished-work.sh
# itself -- that is task 8. Signal two today still resolves a rendered
# Markdown record via panel_record_path and finds none of these fixtures, so
# every signal-two case below is EXPECTED TO FAIL until task 8 lands.
#
# Dropped, relative to signal two's previous cases -- each is a
# Markdown-parsing failure mode or a document-grammar concept with no JSON
# equivalent, or a scenario the new architecture no longer produces at all:
#
#   - the checksum cases (a record with no declared total, a total declared
#     twice, one that is not a plain count, an oversized digit run, and the
#     declared total disagreeing with the marker count) -- the JSON array's
#     own length is the count now; there is no separate declaration left to
#     disagree with it.
#   - the marker-span / contiguity cases (a marker parked away from the
#     block, one quoted inside a fence) -- a JSON array has no notion of a
#     "block" that a stray line could interrupt.
#   - the table/marker-agreement and duplicate-identifier cases (a row with
#     no marker, a marker with no row, an identifier reused on one side or
#     both) -- a decoded finding is a single JSON object carrying its own
#     status together, so there is no second representation -- a findings
#     table -- whose identifier set could disagree with the marker block's,
#     and `findings_ref_key` is a store uniqueness constraint that makes a
#     duplicate ref unreachable in the first place.
#   - the human-table-parsing defects (unescaped pipes, a reordered or absent
#     header, a row missing its leading pipe, a row detached by a blank line
#     or heading, a table inside a blockquote, the record's other tables --
#     roster, convergence -- not being read as findings) -- none of these
#     scenarios exist once there is no table to parse at all.
#   - the malformed-status and case-folding cases (a capitalised `Open`, a
#     status with trailing commentary, a zero-width character inside the
#     status, a withdrawn marker with no reason, an upper-case `WITHDRAWN`,
#     a marker line that does not begin its own line) -- task 2's write-time
#     validation guarantees every stored status is `open`, `fixed`, or
#     `withdrawn <reason>`, so task 8 retires the branches that detected
#     these shapes entirely; none of them can reach the guard any more.
#   - the CRLF and NUL-byte robustness cases -- both exercised a hand-rolled
#     grep-based parser's byte-level failure modes (`read -r` not splitting
#     off a trailing `\r`, a stray NUL byte putting grep into binary mode). A
#     `jq` decode failure has one shape, not the several a text-based parser
#     had to distinguish.
#   - the fix-mutation placement cases (a `fix-mutation:` line splitting the
#     marker block, or shaped like a findings-table row) -- both describe a
#     line planted inside or beside a markdown structure that no longer
#     exists to be split or misread.
#   - the missing-panel-record case and the record-location group (the
#     record read from the old sdd path, another change's dated record, an
#     undated file, the earliest of two dated records) -- all exercise
#     `panel_record_path`'s glob-matching over `docs/superpowers/reviews`,
#     which task 8 deletes outright along with the function itself. A change
#     the store has never heard of now answers `[]` at exit 0 (task 4's
#     `ErrNotFound` branch) -- exactly the "zero findings" case below, not a
#     separate outstanding case.
#
# THE STORE-UNREACHABLE CASE IS NEW. It is the only "cannot answer at all"
# shape left for signal two: `flow record findings` itself failing.
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
# empty stdout from a stdout carrying the message. <worktree>/bin -- the stub
# flow's home, when the fixture carries one -- is placed ahead of the real
# PATH for this one invocation only, never leaking into the harness's own
# shell.
run_guard() {
  local wt="$1"
  set +e
  OUT="$(PATH="$wt/bin:$PATH" "$GUARD" "$@" 2>"$ERRFILE")"
  RC=$?
  set -e
  ERR="$(cat "$ERRFILE")"
}

# run_guard_in <locale> <worktree> <change-name> — the same, with LC_ALL set in
# the guard's environment. The guard's verdict must not depend on the collating
# sequence of the session that invoked it: comparing a status by collation made
# a zero-width character in it report OUTSTANDING under a UTF-8 locale and
# vanish under LC_ALL=C, which is one of signal one's own allowlist hazards
# (case 8e-i below).
run_guard_in() {
  local loc="$1"
  shift
  local wt="$1"
  set +e
  OUT="$(LC_ALL="$loc" PATH="$wt/bin:$PATH" "$GUARD" "$@" 2>"$ERRFILE")"
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

# findings_json <status> [<status> ...] -- a compact JSON array of finding
# objects, one per <status>, refs assigned F1..Fn in order. Built via
# `jq -nc --args` rather than hand-quoted string interpolation, so a status
# carrying a space or a quote (case 4l/1b's "withdrawn ... reason" text) is
# escaped correctly without this harness reimplementing JSON string escaping
# itself -- the same helper shape test-check-panel-reproducers.sh's
# findings_json uses. The reproducer field is a constant: signal two's jq
# query (task 8) only ever inspects status.
findings_json() {
  jq -nc '
    [$ARGS.positional as $a
     | range(0; $a | length)
     | {ref: ("F" + ((. + 1) | tostring)), status: $a[.], reproducer: "scripts/x.sh"}]
  ' --args -- "$@"
}

# set_findings [<status> ...] -- (re)writes the current fixture's stub data
# file so the next `flow record findings` call inside $WT answers with
# these statuses. Called with no arguments, it writes an empty array -- a
# store with no rows for this run. The stub script itself is written once, by
# new_fixture; this only ever rewrites the data it cats.
set_findings() {
  findings_json "$@" > "$WT/bin/findings.json"
}

# set_findings_unreachable -- replaces the current fixture's stub `flow`
# entirely with one that exits non-zero and prints nothing useful to stdout,
# simulating a store `record findings` could not reach.
set_findings_unreachable() {
  cat > "$WT/bin/flow" <<'STUB'
#!/usr/bin/env bash
echo "flow: connect: connection refused" >&2
exit 1
STUB
  chmod +x "$WT/bin/flow"
}

# new_fixture -> sets WT to a worktree holding a fully finished change named
# "demo": every plan item checked, one closed ("fixed") finding answered by a
# stub `flow` on WT/bin, ahead of the real one on PATH for any guard
# invocation run_guard/run_guard_in makes against WT.
new_fixture() {
  WT="$(mktemp -d "${TMPDIR:-/tmp}/unfinished-work-test.XXXXXX")"
  SANDBOXES+=("$WT")
  mkdir -p "$WT/spectre/changes/demo" "$WT/.superpowers/sdd" "$WT/bin"
  printf -- '- [x] 1.1 done\n' > "$WT/spectre/changes/demo/tasks.md"
  cat > "$WT/bin/flow" <<'STUB'
#!/usr/bin/env bash
cat "$(dirname -- "$0")/findings.json"
exit 0
STUB
  chmod +x "$WT/bin/flow"
  set_findings fixed
}

# 1. The whole point of the CLEAR verdict: a finished change is not interrupted.
new_fixture
run_guard "$WT" demo
assert_verdict "CLEAR:" "a finished change is CLEAR"

# 1b. The other closed statuses are closed, not open.
new_fixture
set_findings fixed "withdrawn retracted, the guard already covers it" fixed
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
# READ THE HEADER ABOVE FIRST for the full list of what this rewrite drops
# and why. What remains is the behavioural core, once the finding data comes
# from a decoded JSON array rather than a hand-parsed marker block: how many
# findings are open, that fixed/withdrawn-with-reason findings are not, that
# zero findings (from an empty array, whatever the reason) is clean, and that
# a store the guard cannot reach is a refusal, never a clean answer.

# 4a. One open finding.
new_fixture
set_findings open
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "an open finding is OUTSTANDING"
assert_reason "1 open finding(s)" "an open finding names its signal, with its count"

# 4b. Three open findings among four are three, not one and not four.
new_fixture
set_findings open fixed open open
run_guard "$WT" demo
assert_verdict "OUTSTANDING:" "several open findings are OUTSTANDING"
assert_reason "3 open finding(s)" "every open finding is counted, not just the first"

# 4c. A withdrawal that states its reason on the marker is closed.
new_fixture
set_findings "withdrawn the operator retracted it: the guard already covers this"
run_guard "$WT" demo
assert_verdict "CLEAR:" "a withdrawal with a reason is closed"

# 4d. A store with no findings at all for this run is clean, whether that is
#     because the run genuinely raised none or because the store has never
#     heard of it — `record findings` answers `[]` either way (task 4's
#     `ErrNotFound` branch), and there is no separate "never declared" case
#     left to distinguish from a genuine zero, unlike the retired checksum.
new_fixture
set_findings
run_guard "$WT" demo
assert_verdict "CLEAR:" "a store with no findings for this run is CLEAR"

# 4e. THE STORE IS UNREACHABLE -- stub flow exits non-zero for
#     `record findings`, and the guard must exit 2 ("cannot determine
#     anything"), never 0 (a signal it cannot evaluate must never read as
#     CLEAR) and never 1.
#
#     THIS CASE FAILS UNTIL TASK 8 LANDS, on purpose: today's guard still
#     resolves a rendered Markdown file, finds none (this sandbox writes no
#     docs/superpowers/reviews/*-panel.md), and reports "no review panel
#     record for ..." as one OUTSTANDING reason at exit 0 -- the wrong exit
#     code and the wrong wording, so it correctly fails now and will only
#     pass once task 8's guard actually calls flow and surfaces its
#     failure that way.
new_fixture
set_findings_unreachable
run_guard "$WT" demo
[ "$RC" -eq 2 ] && pass "an unreachable store exits 2, never 0 or 1" \
  || fail "unreachable store: expected exit 2, got rc=$RC out=$OUT"
[ -z "$OUT" ] && pass "an unreachable store emits no verdict line" \
  || fail "unreachable store: emitted a verdict line: $OUT"
case "$ERR" in
  *"cannot determine anything"*) pass "an unreachable store names the failure on stderr" ;;
  *) fail "unreachable store: no named message on stderr: $ERR" ;;
esac

# 5. Both surviving signals firing at once must be counted independently and
#    reported together on the one line — the guard's own header says so, and
#    this is the direct demonstration: an unchecked plan item and an open
#    finding, both named in the single OUTSTANDING line.
new_fixture
printf -- '- [ ] 1.1 not done\n' > "$WT/spectre/changes/demo/tasks.md"
set_findings open
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
#     findings signal keeps this case specific to count_matching's own
#     discipline on signal one, which this rewrite does not touch.
#     Skipped when running as root, where chmod 000 does not block reads.
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
#     tasks.md. Findings are read from $WT's own fixture (already CLEAR via
#     new_fixture's default), never from $PLANTED: the change name reaches
#     `flow record findings` only as the `-change` argument evaluated by
#     $WT's own stub, so no name containing "/" can make that stub answer for
#     a different sandbox — and no second stub is planted at $PLANTED here.
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

# 11. THE CLI'S DIAGNOSTICS MUST NOT REACH THE JSON PARSER.
#
# `flow` writes diagnostics to stderr while the findings JSON goes to stdout --
# most reliably the `flow: using FLOW_ADDR=...` line it prints whenever the
# address is overridden, which is precisely what this repository's own ui-test
# stack instructs an operator to do (`export FLOW_ADDR=http://127.0.0.1:4174`).
# The guard used to capture both streams together, so that one diagnostic line
# landed at the head of the payload and `jq` failed with a parse error on a run
# that had actually succeeded -- reporting "cannot determine anything" for a
# perfectly readable store. It broke only for the operator who followed the
# documented instructions, which is why it is pinned here rather than tolerated.
#
# The stub below reproduces exactly that shape: a diagnostic on stderr, valid
# JSON on stdout, exit 0. Reverting the guard to a `2>&1` capture makes this
# case fail.
new_fixture
cat > "$WT/bin/flow" <<'STUB'
#!/usr/bin/env bash
echo "flow: using FLOW_ADDR=http://127.0.0.1:4174" >&2
echo "[]"
exit 0
STUB
chmod +x "$WT/bin/flow"
run_guard "$WT" demo
[ "$RC" -eq 0 ] && pass "a diagnostic on stderr does not corrupt the findings JSON" \
  || fail "stderr diagnostic: expected exit 0, got rc=$RC out=$OUT err=$ERR"
case "$OUT$ERR" in
  *"jq failed"*) fail "stderr diagnostic: the diagnostic reached jq" ;;
  *) pass "a diagnostic on stderr never reaches jq" ;;
esac

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'check-unfinished-work: all cases pass\n'
