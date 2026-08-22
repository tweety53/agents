#!/usr/bin/env bash
# Assertion harness for check-worktree-processes.sh. Builds throwaway worktree
# stand-ins under a sandboxed TMPDIR, starts real processes inside some of them,
# and asserts the guard's verdict line and its exit status. Never MODIFIES the
# real repository tree, and never inspects a process it did not start itself.
#
# READ THIS BEFORE ADDING OR "FIXING" A CASE. Assert against the stated contract
# — the verdict grammar under **Global constraints** in
# openspec/changes/kan-242-devstop-not-running-while-stack-alive/tasks.md and the
# requirement **Worktree cleanup verifies the stack actually stopped, before
# removing anything** in that change's specs/myflow-finish-cleanup/spec.md.
# Never assert against observed output.
# test-check-plan-provenance.sh's header records that suite encoding the guard's
# own defects as its specification more than once, which then made each defect
# look verified.
#
# THE GRAMMAR THIS FILE IS THE EXECUTABLE STATEMENT OF:
#
#   CLEAR: <worktree>                       no process has a cwd at or under it
#   HELD: <worktree> — <pid> <cwd>[; ...]   one or more processes do
#
# Exit 0 whenever a verdict was reached — the verdict carries the answer, not
# the exit status, exactly as check-cleanup-complete.sh splits the two. Exit 2
# when the guard cannot answer at all, with NOTHING on stdout: a caller grepping
# for CLEAR in empty output must not be able to read an inability as a pass.
#
# WHY EVERY PATH IN THIS FILE IS RESOLVED WITH `pwd -P`. lsof reports physical
# paths, so on macOS a fixture under TMPDIR is `/private/tmp/...` to lsof and
# `/tmp/...` to mktemp. A harness that compared the mktemp form would report
# CLEAR for every held fixture and prove nothing — which is precisely the defect
# case 2 exists to catch, so the harness must not carry it too.
#
# WHY THIS DUPLICATES test-check-workspace-isolation.sh's HELPERS instead of
# sharing them. That harness's shape — sandboxed TMPDIR fixtures, `pass`/`fail`
# counters, a `run_guard` that captures stdout and stderr separately — is
# followed here deliberately, and copied rather than extracted: the two suites
# test unrelated guards, and a shared harness library would mean a change to one
# guard's contract could only be made by editing a file the other one also runs.
# The duplication is the cheaper of the two, and it is recorded here rather than
# left unexplained.
#
# Bash 3.2 is the floor, as test-check-finish-preflight.sh's header records:
# indexed arrays only, no associative arrays.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# CHECK_WORKTREE_PROCESSES_GUARD points this suite at a guard other than its
# sibling. It exists for ONE caller — the mutation case at the end of this file,
# which re-runs this whole suite against a deliberately broken copy of the guard
# and requires it to go red. Never set it for a normal invocation.
GUARD="${CHECK_WORKTREE_PROCESSES_GUARD:-$SCRIPT_DIR/check-worktree-processes.sh}"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

# skip <label> <reason> — a case this environment cannot decide, reported as
# neither a pass nor a failure. Printing "ok" for an assertion that never ran is
# the vacuous pass every guard in this repository is written against.
skip() { printf 'skip: %s (%s)\n' "$1" "$2"; }

# An indexed array, not a space-separated string: sandbox paths come from mktemp
# under TMPDIR, which may contain spaces, and word-splitting a string would then
# rm -rf the fragments.
SANDBOXES=()
# Every pid this harness starts. The cleanup below reaps them ALL, on the happy
# path and on a signal alike — a suite about orphaned processes that leaked its
# own would be reproducing the incident it exists to prevent.
STARTED_PIDS=()
cleanup() {
  if [ "${#STARTED_PIDS[@]}" -ne 0 ]; then
    for p in "${STARTED_PIDS[@]}"; do
      kill "$p" 2>/dev/null || true
      wait "$p" 2>/dev/null || true
    done
  fi
  if [ "${#SANDBOXES[@]}" -ne 0 ]; then
    # chmod first: one case makes a fixture directory unreadable on purpose, and
    # rm -rf cannot descend into it until the bit is put back.
    for s in "${SANDBOXES[@]}"; do
      chmod -R u+rwX "$s" 2>/dev/null || true
      rm -rf "$s"
    done
  fi
}
trap cleanup EXIT

WORK="$(mktemp -d "${TMPDIR:-/tmp}/worktree-processes-test.XXXXXX")"
SANDBOXES+=("$WORK")
ERRFILE="$WORK/stderr"

# run_guard <arg ...> -> sets OUT (stdout only), ERR, RC.
# The two streams are captured SEPARATELY rather than merged with 2>&1, because
# the contract distinguishes them: a refusal puts its message on stderr and must
# leave stdout empty, and a merged capture cannot tell an empty stdout from a
# stdout carrying the message.
run_guard() {
  set +e
  OUT="$("$GUARD" "$@" 2>"$ERRFILE")"
  RC=$?
  set -e
  ERR="$(cat "$ERRFILE")"
}

# new_worktree [subdir] -> sets WT (as mktemp spelled it) and PHYS (as lsof will
# spell it). Both are needed: WT is what a caller hands the guard, PHYS is what
# every assertion about a path compares against.
new_worktree() {
  WT="$(mktemp -d "${TMPDIR:-/tmp}/worktree-processes-wt.XXXXXX")"
  SANDBOXES+=("$WT")
  PHYS="$(cd "$WT" && pwd -P)"
}

# start_in <dir> -> sets PID to a live process whose working directory is <dir>.
# `exec` inside the subshell replaces it, so $! is the sleep's own pid rather
# than a shell that merely spawned one.
start_in() {
  ( cd "$1" && exec sleep 300 ) &
  PID=$!
  STARTED_PIDS+=("$PID")
}

# await_cwd <pid> <physical-dir> -> 0 once lsof reports that pid's cwd as that
# directory. The subshell in start_in has not necessarily reached its `cd` by
# the time `&` returns, so a case that read the process table immediately would
# be racing. Polling for the state the case needs is deterministic where a fixed
# sleep is a guess; the bound is 5 seconds, after which the caller reports a
# failure rather than continuing on a premise that never became true.
await_cwd() {
  local i=0
  while [ "$i" -lt 100 ]; do
    if lsof -a -d cwd -p "$1" -Fn 2>/dev/null | grep -qx -- "n$2"; then
      return 0
    fi
    i=$((i + 1))
    sleep 0.05
  done
  return 1
}

# assert_clear <label> — exit 0 and a CLEAR verdict naming the worktree.
assert_clear() {
  if [ "$RC" -ne 0 ]; then
    fail "$1: expected exit 0, got rc=$RC out=$OUT err=$ERR"
    return 0
  fi
  case "$OUT" in
    "CLEAR: $PHYS"*) pass "$1" ;;
    *) fail "$1: expected 'CLEAR: $PHYS', got: $OUT" ;;
  esac
}

# assert_held <label> <pid> <cwd> — exit 0, a HELD verdict naming the worktree,
# and that pid paired with that working directory inside it. The pid and the
# path are asserted together rather than separately: a report naming a pid the
# operator cannot locate, or a directory with no pid to reach it, leaves the
# operator with nothing to act on, and the spec's first scenario requires both.
assert_held() {
  if [ "$RC" -ne 0 ]; then
    fail "$1: expected exit 0, got rc=$RC out=$OUT err=$ERR"
    return 0
  fi
  case "$OUT" in
    "HELD: $PHYS "*) pass "$1: reports HELD for the worktree" ;;
    *) fail "$1: expected a HELD verdict for $PHYS, got: $OUT"; return 0 ;;
  esac
  case "$OUT" in
    *"$2 $3"*) pass "$1: names pid $2 with its working directory $3" ;;
    *) fail "$1: the verdict does not pair pid $2 with $3: $OUT" ;;
  esac
}

# assert_cannot_answer <label> — exit 2, an empty stdout, and a named reason on
# stderr. All three matter: the exit code so a caller can branch, the empty
# stdout so no verdict token can be found in it, and the reason so the operator
# knows what to fix.
assert_cannot_answer() {
  [ "$RC" -eq 2 ] && pass "$1: exits 2" \
    || fail "$1: expected exit 2, got rc=$RC out=$OUT err=$ERR"
  [ -z "$OUT" ] && pass "$1: writes nothing to stdout" \
    || fail "$1: emitted something on stdout: $OUT"
  case "$OUT" in
    *CLEAR*|*HELD*) fail "$1: an inability printed a verdict token: $OUT" ;;
    *) pass "$1: prints no verdict token" ;;
  esac
  case "$ERR" in
    *"check-worktree-processes: "*) pass "$1: names the failure on stderr" ;;
    *) fail "$1: no named message on stderr: $ERR" ;;
  esac
}

# ---------------------------------------------------------------------------
# 1. A worktree nothing is running from.
# ---------------------------------------------------------------------------

# Case 1: the ordinary pre-removal state. Nothing was ever started here, so the
# guard must say so plainly and exit 0 — an empty answer and a refusal are
# different outcomes, and this case pins the first of them.
new_worktree
run_guard "$WT"
assert_clear "case 1: a worktree with no process in it is CLEAR"

# ---------------------------------------------------------------------------
# 2. A worktree a live process is running from. This is the incident.
# ---------------------------------------------------------------------------

# Case 2: a process whose cwd IS the worktree. Exit is 0 here on purpose: a
# verdict was reached, so the exit status is not where the answer lives.
new_worktree
start_in "$PHYS"
HELD_PID="$PID"
if await_cwd "$HELD_PID" "$PHYS"; then
  run_guard "$WT"
  assert_held "case 2: a process whose cwd is the worktree is HELD" "$HELD_PID" "$PHYS"
  [ "$RC" -eq 0 ] && pass "case 2: a reached verdict exits 0" \
    || fail "case 2: expected exit 0 for a reached verdict, got rc=$RC"
else
  fail "case 2: lsof never reported pid $HELD_PID's cwd as $PHYS — the harness cannot see its own process"
fi

# Case 3: a process in a SUBDIRECTORY of the worktree. This is the ordinary real
# shape — a backend started from its own module root — and a guard that compared
# only for equality with the worktree root would miss every real one.
new_worktree
mkdir -p "$WT/stats"
SUBDIR="$(cd "$WT/stats" && pwd -P)"
start_in "$SUBDIR"
SUB_PID="$PID"
if await_cwd "$SUB_PID" "$SUBDIR"; then
  run_guard "$WT"
  assert_held "case 3: a process under the worktree is HELD" "$SUB_PID" "$SUBDIR"
else
  fail "case 3: lsof never reported pid $SUB_PID's cwd as $SUBDIR — the harness cannot see its own process"
fi

# ---------------------------------------------------------------------------
# 3. "At or under" is not "starts with". The sibling-prefix defect.
# ---------------------------------------------------------------------------

# Case 4: two sibling worktrees whose names share a prefix, which is exactly what
# this pipeline's own naming produces — openspec-kan-1 beside openspec-kan-12. A
# bare string-prefix test reports the first as HELD by the second's process, and
# a run 2 blocked on a process in a different worktree is a false alarm the
# operator cannot clear.
SIBLINGS="$(mktemp -d "${TMPDIR:-/tmp}/worktree-processes-siblings.XXXXXX")"
SANDBOXES+=("$SIBLINGS")
mkdir -p "$SIBLINGS/openspec-kan-1" "$SIBLINGS/openspec-kan-12"
WT="$SIBLINGS/openspec-kan-1"
PHYS="$(cd "$WT" && pwd -P)"
NEIGHBOUR="$(cd "$SIBLINGS/openspec-kan-12" && pwd -P)"
start_in "$NEIGHBOUR"
SIB_PID="$PID"
if await_cwd "$SIB_PID" "$NEIGHBOUR"; then
  run_guard "$WT"
  assert_clear "case 4: a sibling sharing a name prefix is not HELD"
  case "$OUT" in
    *"$SIB_PID"*) fail "case 4: the sibling's pid $SIB_PID reached the verdict: $OUT" ;;
    *) pass "case 4: the sibling's pid is absent from the verdict" ;;
  esac
else
  fail "case 4: lsof never reported pid $SIB_PID's cwd as $NEIGHBOUR — the harness cannot see its own process"
fi

# ---------------------------------------------------------------------------
# 4. Inputs it cannot answer about. Never fail open: a path it could not read is
#    reported as an inability, never as a clean result.
# ---------------------------------------------------------------------------

# Case 5a: a path that does not exist.
WT="$WORK/no-such-worktree"
run_guard "$WT"
assert_cannot_answer "case 5a: a worktree path that does not exist"

# Case 5b: a path that exists but is not a directory. A guard that only tested
# for existence would try to scan it and report something.
printf 'not a directory\n' > "$WORK/regular-file"
run_guard "$WORK/regular-file"
assert_cannot_answer "case 5b: a worktree path that is a regular file"

# Case 5c: a directory whose contents cannot be read. Root ignores the mode bits,
# so this case reports itself as skipped there rather than passing vacuously.
UNREADABLE="$WORK/unreadable"
mkdir -p "$UNREADABLE"
chmod 000 "$UNREADABLE"
if [ "$(id -u)" = "0" ]; then
  skip "case 5c: an unreadable worktree directory" "running as root, which ignores the mode bits"
else
  run_guard "$UNREADABLE"
  assert_cannot_answer "case 5c: an unreadable worktree directory"
fi
chmod 755 "$UNREADABLE"

# Case 6: the scanning tool is absent. The guard's whole answer comes from lsof,
# so a PATH without it is an inability and not an empty result — and the message
# has to name the tool, or the operator is told only that something failed.
NOLSOF="$WORK/empty-path"
mkdir -p "$NOLSOF"
# The guard's `#!/usr/bin/env bash` line resolves `bash` through PATH, so an
# entirely empty PATH would stop the script from starting at all and would
# prove nothing about lsof. This directory therefore carries bash — and nothing
# else, lsof included.
ln -s "$(command -v bash)" "$NOLSOF/bash"
new_worktree
set +e
OUT="$(PATH="$NOLSOF" "$GUARD" "$WT" 2>"$ERRFILE")"
RC=$?
set -e
ERR="$(cat "$ERRFILE")"
assert_cannot_answer "case 6: lsof is not on PATH"
case "$ERR" in
  *lsof*) pass "case 6: the message names the missing tool" ;;
  *) fail "case 6: the message does not name lsof: $ERR" ;;
esac

# ---------------------------------------------------------------------------
# 5. KAN-197 mutation: this suite must detect the guard's own central defect.
#    A suite that cannot is not a suite, and every case above would then be
#    evidence of nothing.
# ---------------------------------------------------------------------------

# Skipped when this run IS the mutant run, which is how the recursion ends.
if [ -n "${CHECK_WORKTREE_PROCESSES_GUARD:-}" ]; then
  skip "case 7 (KAN-197 mutation)" "this run is itself the mutant run"
else
  MUTANT_DIR="$WORK/mutant"
  mkdir -p "$MUTANT_DIR"
  MUTANT="$MUTANT_DIR/check-worktree-processes.sh"
  # The mutation: the guard's one decision — whether a scanned working directory
  # is at or under the worktree — is forced false, so the accumulator stays empty
  # and every verdict becomes CLEAR. That is the guard's central defect expressed
  # as one edit.
  sed 's/if is_at_or_under "$path"; then/if false; then/' "$GUARD" > "$MUTANT"
  chmod +x "$MUTANT"

  # A mutation that stopped applying — because the guard was reworded — would
  # leave the copy identical to the original, which passes this suite, and the
  # case would then report a detection it never made. Asserting the edit landed
  # is what keeps this case from going quietly vacuous.
  if cmp -s "$GUARD" "$MUTANT"; then
    fail "case 7 (KAN-197 mutation): the mutation did not apply — the guard no longer carries the line this case edits, so nothing was mutated"
  else
    pass "case 7 (KAN-197 mutation): the mutation applied"
    set +e
    CHECK_WORKTREE_PROCESSES_GUARD="$MUTANT" "$SCRIPT_DIR/test-check-worktree-processes.sh" >/dev/null 2>&1
    MUTANT_RC=$?
    set -e
    [ "$MUTANT_RC" -ne 0 ] && pass "case 7 (KAN-197 mutation): a guard that always prints CLEAR fails this suite" \
      || fail "case 7 (KAN-197 mutation): a guard that always prints CLEAR passed this suite, which therefore detects nothing"
  fi
fi

if [ "$FAILURES" -ne 0 ]; then
  printf '%s case(s) failed\n' "$FAILURES" >&2
  exit 1
fi
printf 'check-worktree-processes: all cases pass\n'
