#!/usr/bin/env bash
# test-lib-parallel.sh — assertion harness for scripts/lib/parallel.sh, the
# shared bounded-concurrency spawn/capture/replay primitive KAN-362 task 1
# adds. Sources the library directly rather than through either caller
# (scripts/run-guard-tests.sh, scripts/test-check-installed-citations.sh),
# per this repository's own convention (see test-lib-coverage.sh's header):
# the thing under test is the library's own contract, not any caller's use
# of it.
#
# The five bullets design.md's Testing section names for this file: a clean
# run replays nothing; a failing job's own captured stdout+stderr replays in
# full; replay follows LIST order under deliberately shuffled completion
# order (proven with differing sleep durations, not by hoping the scheduler
# varies); output that interleaves in real time still replays contiguously;
# JOBS=1 and JOBS=<n> produce identical output. A "robustness" section below
# those five exercises JOBS validation (bad value refused at exit 2, never
# coerced) and the mktemp -d cleanup, including on SIGINT/SIGTERM.
#
# Bash 3.2 is the floor, as scripts/test-check-finish-preflight.sh's header
# records: indexed arrays only, no associative arrays, no `wait -n`.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LIB="$SCRIPT_DIR/lib/parallel.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

if [ ! -r "$LIB" ]; then
  echo "test-lib-parallel: cannot read $LIB" >&2
  exit 2
fi
# shellcheck source=lib/parallel.sh
source "$LIB"

# assert_eq <label> <expected> <actual>
assert_eq() {
  if [ "$2" = "$3" ]; then
    pass "$1"
  else
    fail "$1 — expected [$2], got [$3]"
  fi
}

# assert_contains <label> <haystack> <needle>
assert_contains() {
  case "$2" in
    *"$3"*) pass "$1" ;;
    *) fail "$1 — [$2] does not contain [$3]" ;;
  esac
}

# assert_not_contains <label> <haystack> <needle>
assert_not_contains() {
  case "$2" in
    *"$3"*) fail "$1 — [$2] unexpectedly contains [$3]" ;;
    *) pass "$1" ;;
  esac
}

# assert_zero_rc / assert_nonzero_rc <label> <rc>
assert_zero_rc() {
  if [ "$2" -eq 0 ]; then pass "$1"; else fail "$1 — expected exit 0, got $2"; fi
}
assert_nonzero_rc() {
  if [ "$2" -ne 0 ]; then pass "$1"; else fail "$1 — expected a non-zero exit, got 0"; fi
}

# run_parallel <job...> -> sets RC (parallel_run's own return), REPLAY
# (parallel_replay_failures' stdout). Must not abort this harness under
# `set -e`, same set +e/set -e bracket test-check-panel-reproducers.sh's
# header describes for a command that is SUPPOSED to fail sometimes.
run_parallel() {
  set +e
  parallel_run "$@"
  RC=$?
  REPLAY="$(parallel_replay_failures)"
  set -e
}

# ---------------------------------------------------------------------------
# 1. A clean run exits 0 and replays nothing.
# ---------------------------------------------------------------------------
run_parallel "exit 0" "printf ok" "true"
assert_zero_rc "case 1: an all-passing run exits 0" "$RC"
assert_eq "case 1: an all-passing run replays nothing" "" "$REPLAY"

# MUTATION PROOF for case 1: a run with one failing job among passing ones
# must NOT satisfy the same assertions — proves assert_zero_rc above
# actually discriminates pass from fail rather than trivially passing.
run_parallel "exit 0" "exit 1" "true"
assert_nonzero_rc "case 1 mutation: one failing job among passing ones exits non-zero" "$RC"

# ---------------------------------------------------------------------------
# 2. One failing job exits non-zero and replays exactly that job's own
#    captured stdout and stderr — proven against a job that writes a
#    distinctive marker on BOTH streams, with a passing neighbour whose own
#    marker must NOT appear in the replay.
# ---------------------------------------------------------------------------
run_parallel \
  'printf "PASS-OUT-MARKER\n"; printf "PASS-ERR-MARKER\n" >&2; exit 0' \
  'printf "FAIL-OUT-MARKER\n"; printf "FAIL-ERR-MARKER\n" >&2; exit 7'
assert_nonzero_rc "case 2: a failing job among jobs exits non-zero" "$RC"
assert_contains "case 2: replay carries the failing job's stdout" "$REPLAY" "FAIL-OUT-MARKER"
assert_contains "case 2: replay carries the failing job's stderr" "$REPLAY" "FAIL-ERR-MARKER"
assert_not_contains "case 2: replay does not carry the passing job's stdout" "$REPLAY" "PASS-OUT-MARKER"
assert_not_contains "case 2: replay does not carry the passing job's stderr" "$REPLAY" "PASS-ERR-MARKER"
assert_eq "case 2: the failing job's own recorded exit code is preserved" "7" "${PARALLEL_RC[1]}"

# ---------------------------------------------------------------------------
# 3. Replay follows LIST order under deliberately shuffled completion order.
#    Job 0 sleeps longest and fails; job 1 finishes fastest and fails; job 2
#    finishes in the middle and fails. Completion order is 1, 2, 0 — replay
#    must read 0, 1, 2 (list order), never the completion order.
# ---------------------------------------------------------------------------
JOBS=3 run_parallel \
  'sleep 0.6; printf "MARK-0\n"; exit 1' \
  'sleep 0.1; printf "MARK-1\n"; exit 1' \
  'sleep 0.3; printf "MARK-2\n"; exit 1'
LIST_ORDER_IDX0=$(printf '%s' "$REPLAY" | grep -n 'MARK-0' | cut -d: -f1)
LIST_ORDER_IDX1=$(printf '%s' "$REPLAY" | grep -n 'MARK-1' | cut -d: -f1)
LIST_ORDER_IDX2=$(printf '%s' "$REPLAY" | grep -n 'MARK-2' | cut -d: -f1)
if [ "$LIST_ORDER_IDX0" -lt "$LIST_ORDER_IDX1" ] && [ "$LIST_ORDER_IDX1" -lt "$LIST_ORDER_IDX2" ]; then
  pass "case 3: replay follows list order under shuffled completion order"
else
  fail "case 3: replay followed completion order, not list order — got 0@$LIST_ORDER_IDX0 1@$LIST_ORDER_IDX1 2@$LIST_ORDER_IDX2"
fi

# MUTATION PROOF for case 3: completion order genuinely differs from list
# order in this scenario — if it did not, the assertion above would pass
# vacuously regardless of whether replay actually sorts by list order. Prove
# it by recording each job's own finish instant (via `date +%s%N`-free
# ordering: a shared counter file each job appends its index to) and
# checking the recorded completion order is NOT 0,1,2.
COUNTER_FILE="$(mktemp)"
trap 'rm -f "$COUNTER_FILE"' RETURN
JOBS=3 parallel_run \
  "sleep 0.6; printf '0\n' >> '$COUNTER_FILE'" \
  "sleep 0.1; printf '1\n' >> '$COUNTER_FILE'" \
  "sleep 0.3; printf '2\n' >> '$COUNTER_FILE'" \
  >/dev/null 2>&1 || true
COMPLETION_ORDER="$(tr -d '\n' < "$COUNTER_FILE")"
rm -f "$COUNTER_FILE"
trap - RETURN
if [ "$COMPLETION_ORDER" != "012" ]; then
  pass "case 3 mutation proof: completion order ($COMPLETION_ORDER) genuinely differs from list order (012), so case 3 is not vacuous"
else
  fail "case 3 mutation proof: completion order came back 012 — the shuffle did not actually shuffle, case 3 proves nothing"
fi

# ---------------------------------------------------------------------------
# 4. Jobs whose output interleaves in real time are still replayed
#    contiguously — two failing jobs each print an interleavable sequence of
#    lines with a tiny sleep between them; replay must show each job's lines
#    as one unbroken block, never interleaved with the other's.
# ---------------------------------------------------------------------------
run_parallel \
  'for i in 1 2 3 4 5; do printf "A-%s\n" "$i"; sleep 0.05; done; exit 1' \
  'for i in 1 2 3 4 5; do printf "B-%s\n" "$i"; sleep 0.05; done; exit 1'
A_BLOCK="$(printf '%s\n' "$REPLAY" | grep '^A-' | tr '\n' ',')"
B_BLOCK="$(printf '%s\n' "$REPLAY" | grep '^B-' | tr '\n' ',')"
assert_eq "case 4: job A's five lines are unbroken and in order" "A-1,A-2,A-3,A-4,A-5," "$A_BLOCK"
assert_eq "case 4: job B's five lines are unbroken and in order" "B-1,B-2,B-3,B-4,B-5," "$B_BLOCK"
# The interleaving lines are not contiguous in the replay text itself unless
# grouped by job — check the raw replay does NOT show A and B lines
# alternating (which would prove capture, not just filtering, is per-job).
RAW_SEQUENCE="$(printf '%s\n' "$REPLAY" | grep -E '^[AB]-' | cut -c1)"
FIRST_HALF="$(printf '%s' "$RAW_SEQUENCE" | head -n 5 | sort -u | tr -d '\n')"
SECOND_HALF="$(printf '%s' "$RAW_SEQUENCE" | tail -n 5 | sort -u | tr -d '\n')"
if [ "$FIRST_HALF" = "A" ] || [ "$FIRST_HALF" = "B" ]; then
  pass "case 4: the replay's first five lines belong to a single job (contiguous, not interleaved)"
else
  fail "case 4: the replay's first five lines mix jobs — got letters [$FIRST_HALF]"
fi
if [ "$SECOND_HALF" = "A" ] || [ "$SECOND_HALF" = "B" ]; then
  pass "case 4: the replay's last five lines belong to a single job (contiguous, not interleaved)"
else
  fail "case 4: the replay's last five lines mix jobs — got letters [$SECOND_HALF]"
fi

# ---------------------------------------------------------------------------
# 5. JOBS=1 and JOBS=<n> produce identical output.
# ---------------------------------------------------------------------------
JOBS=1 run_parallel \
  'printf "J1-out\n"; printf "J1-err\n" >&2; exit 3' \
  'printf "J2-out\n"; exit 0'
REPLAY_J1="$REPLAY"
RC_J1="$RC"
JOBS=4 run_parallel \
  'printf "J1-out\n"; printf "J1-err\n" >&2; exit 3' \
  'printf "J2-out\n"; exit 0'
REPLAY_J4="$REPLAY"
RC_J4="$RC"
assert_eq "case 5: JOBS=1 and JOBS=4 produce identical replay output" "$REPLAY_J1" "$REPLAY_J4"
assert_eq "case 5: JOBS=1 and JOBS=4 produce identical exit codes" "$RC_J1" "$RC_J4"

# ---------------------------------------------------------------------------
# 5b. parallel_replay_failures accepts optional per-job labels (F1: the
#    caller-provided name a failing job's output is introduced by, since
#    this library only ever sees opaque command strings). A passing job's
#    label must never appear — only failing jobs get a header at all.
# ---------------------------------------------------------------------------
set +e
parallel_run "exit 0" "exit 1" "exit 1"
RC=$?
REPLAY_LABELED="$(parallel_replay_failures alpha.sh beta.sh gamma.sh)"
set -e
assert_contains "case 5b: a failing job's label appears as a ---- header" "$REPLAY_LABELED" "---- beta.sh ----"
assert_contains "case 5b: the last failing job's label also appears" "$REPLAY_LABELED" "---- gamma.sh ----"
assert_not_contains "case 5b: a passing job's label is never printed" "$REPLAY_LABELED" "alpha.sh"
LABEL_IDX_BETA=$(printf '%s' "$REPLAY_LABELED" | grep -n '^---- beta.sh ----$' | cut -d: -f1)
LABEL_IDX_GAMMA=$(printf '%s' "$REPLAY_LABELED" | grep -n '^---- gamma.sh ----$' | cut -d: -f1)
if [ -n "$LABEL_IDX_BETA" ] && [ -n "$LABEL_IDX_GAMMA" ] && [ "$LABEL_IDX_BETA" -lt "$LABEL_IDX_GAMMA" ]; then
  pass "case 5b: labeled headers appear in list order"
else
  fail "case 5b: labeled headers did not appear in list order — beta@$LABEL_IDX_BETA gamma@$LABEL_IDX_GAMMA"
fi

# MUTATION PROOF for case 5b: calling parallel_replay_failures with NO
# labels must reproduce the exact pre-F1 header-less output — proves the
# labeled form is additive, not a silent change to every caller's existing
# behaviour (scripts/test-run-guard-tests.sh's own fixtures assert this
# unlabeled shape too).
REPLAY_UNLABELED="$(parallel_replay_failures)"
assert_not_contains "case 5b mutation: the unlabeled form prints no ---- header at all" "$REPLAY_UNLABELED" "----"

# ---------------------------------------------------------------------------
# 6. Robustness — JOBS validation is refused, never coerced (exit 2).
# ---------------------------------------------------------------------------
run_job_count() {
  set +e
  JC_OUT="$(JOBS="$1" parallel_job_count 2>/tmp/parallel-jc-err.$$)"
  JC_RC=$?
  JC_ERR="$(cat "/tmp/parallel-jc-err.$$")"
  rm -f "/tmp/parallel-jc-err.$$"
  set -e
}

run_job_count "abc"
assert_eq "case 6a: JOBS=abc is refused with exit 2" "2" "$JC_RC"

run_job_count "0"
assert_eq "case 6b: JOBS=0 is refused with exit 2" "2" "$JC_RC"

run_job_count "-1"
assert_eq "case 6c: JOBS=-1 is refused with exit 2" "2" "$JC_RC"

run_job_count "3.5"
assert_eq "case 6d: JOBS=3.5 is refused with exit 2" "2" "$JC_RC"

set +e
JOBS="" JC_OUT="$(parallel_job_count 2>/dev/null)"
JC_RC=$?
set -e
assert_eq "case 6e: JOBS= (set but empty) is refused with exit 2" "2" "$JC_RC"

run_job_count "5"
assert_eq "case 6f: JOBS=5 (valid) is accepted" "0" "$JC_RC"
assert_eq "case 6f: JOBS=5 (valid) is honoured verbatim" "5" "$JC_OUT"

set +e
unset JOBS
JC_OUT="$(parallel_job_count 2>/dev/null)"
JC_RC=$?
set -e
assert_zero_rc "case 6g: unset JOBS computes a fallback rather than erroring" "$JC_RC"
case "$JC_OUT" in
  ''|*[!0-9]*) fail "case 6g: unset JOBS fallback is not a positive integer — got [$JC_OUT]" ;;
  0) fail "case 6g: unset JOBS fallback must be at least 1 — got 0" ;;
  *) pass "case 6g: unset JOBS fallback is a positive integer ($JC_OUT)" ;;
esac

# MUTATION PROOF for case 6: parallel_run itself must refuse an invalid JOBS
# too, not just parallel_job_count in isolation — proves the validation is
# actually wired into the entry point a caller uses, not a dead function.
set +e
JOBS="not-a-number" parallel_run "true" >/dev/null 2>&1
RC=$?
set -e
assert_eq "case 6 mutation: parallel_run itself refuses a bad JOBS with exit 2" "2" "$RC"

# ---------------------------------------------------------------------------
# Both cases 7 and 8 run the library in a genuinely separate `bash` PROCESS
# (not a `(...)  &` subshell of this harness) so its `$$` is its own PID —
# `_parallel_install_traps`' `kill -s INT "$$"` targets the sourcing
# process's own PID, and inside a forked subshell bash's `$$` still resolves
# to the ORIGINATING process, which would misdirect the re-raised signal
# back at this harness itself. A real child process is also what
# REPRODUCE, DON'T READ calls for here: the real mktemp -d and the real
# signal delivery, not a stand-in for either. Case 8 additionally starts its
# child under `set -m` in its own subshell — see its own comment below for
# why a plain `cmd &` here is not enough.
#
# Each case gives its own child a PRIVATE TMPDIR (its own mktemp -d,
# distinct from the global $TMPDIR) rather than snapshotting the shared
# namespace: scripts/run-guard-tests.sh runs 40+ harnesses concurrently via
# scripts/lib/parallel.sh, and this file, scripts/test-run-guard-tests.sh
# and scripts/test-check-installed-citations.sh all create parallel-lib.*
# directories in that same shared $TMPDIR at the same time. A before/after
# diff against the shared namespace can pick up a SIBLING's live directory
# instead of this case's own child — proven by reproduction (competing
# parallel-lib.* dir created alongside the real child; the diff's `head
# -n1` returns whichever sorts first, sibling or own, nondeterministically).
# A private TMPDIR makes the namespace exclusive to this one child, so
# there is nothing left to disambiguate.
new_tmpdir_since() {
  local base="$2"
  comm -13 <(printf '%s\n' "$1") <(find "$base" -maxdepth 1 -name 'parallel-lib.*' 2>/dev/null | sort) | head -n1
}

# ---------------------------------------------------------------------------
# 7. Each job's output is captured under a mktemp -d that exists while the
#    run is in flight and is removed after a clean exit.
# ---------------------------------------------------------------------------
PRIVATE_TMP_7="$(mktemp -d "${TMPDIR:-/tmp}/test-lib-parallel-case7.XXXXXX")"
BEFORE_7="$(find "$PRIVATE_TMP_7" -maxdepth 1 -name 'parallel-lib.*' 2>/dev/null | sort)"
CHILD7="$(mktemp)"
cat > "$CHILD7" <<EOF
#!/usr/bin/env bash
set -euo pipefail
source "$LIB"
parallel_run "sleep 2; echo mid-flight" "true" >/dev/null 2>&1
EOF
TMPDIR="$PRIVATE_TMP_7" bash "$CHILD7" &
CHILD7_PID=$!
sleep 0.5
NEW_DIR_7="$(new_tmpdir_since "$BEFORE_7" "$PRIVATE_TMP_7")"
INFLIGHT_OK=1
[ -n "$NEW_DIR_7" ] && [ -d "$NEW_DIR_7" ] || INFLIGHT_OK=0
wait "$CHILD7_PID" 2>/dev/null || true
if [ "$INFLIGHT_OK" -eq 1 ]; then
  pass "case 7: a real mktemp -d exists while the run is in flight"
else
  fail "case 7: could not observe a real in-flight mktemp -d"
fi
if [ -n "$NEW_DIR_7" ] && [ ! -d "$NEW_DIR_7" ]; then
  pass "case 7: the mktemp -d is removed after the calling process exits cleanly"
else
  fail "case 7: the mktemp -d [$NEW_DIR_7] was not removed after a clean exit"
  rm -rf "$NEW_DIR_7"
fi
rm -f "$CHILD7"
rm -rf "$PRIVATE_TMP_7"

# ---------------------------------------------------------------------------
# 8. The mktemp -d is removed on interrupt (SIGINT) too, not only on a clean
#    exit — sent to a real child process running a real long-lived job.
#
#    Started under `set -m` inside its own subshell, per
#    scripts/test-check-installed-citations.sh's own trap-chain sub-case 2
#    (see its comment on `set -m` there for the reasoning this recipe is
#    copied from). Confirmed by reproduction, not assumed: a plain `cmd &`
#    here — no job control — leaves CHILD8 in THIS HARNESS's own process
#    group (`ps -o pid,pgid` on both prints the same pgid), and this harness
#    itself runs as one job inside scripts/run-guard-tests.sh's own
#    xargs -P pool of concurrent test-*.sh harnesses, which shares that same
#    process group too (no link in the chain up to the outer runner ever
#    enables job control either). `kill -INT "$CHILD8_PID"` targets that
#    literal pid, but CHILD8's own re-raise (`_parallel_install_traps`'
#    `kill -s INT "$$"`) and the surrounding process churn from 40+ sibling
#    harnesses forking heavily at the same time made that shared group the
#    live risk: reverting this fix and running the real 42-harness pool
#    under scripts/run-guard-tests.sh reproduced runs where the WHOLE outer
#    parallel_run's own capture directory vanished mid-flight ("42
#    harnesses, 0 passed, 42 failed", every sibling's replay a "No such
#    file" error) — the outer runner's own EXIT/INT trap (installed by
#    sourcing scripts/lib/parallel.sh, same as this harness) firing on a
#    signal that was never meant to reach it. `set -m` inside a dedicated
#    subshell gives CHILD8 its own process group, so nothing it is sent or
#    re-raises at its own $$ can reach anything outside itself. The
#    in-flight directory found and the kill/wait sequence both have to live
#    inside that subshell (job-control assignment happens at the
#    background job's own fork, not retroactively), so the found path is
#    handed back to this process via a plain file rather than a variable —
#    a subshell's own variable assignments never survive past its `)`.
# ---------------------------------------------------------------------------
PRIVATE_TMP_8="$(mktemp -d "${TMPDIR:-/tmp}/test-lib-parallel-case8.XXXXXX")"
BEFORE_8="$(find "$PRIVATE_TMP_8" -maxdepth 1 -name 'parallel-lib.*' 2>/dev/null | sort)"
CHILD8="$(mktemp)"
cat > "$CHILD8" <<EOF
#!/usr/bin/env bash
set -euo pipefail
source "$LIB"
parallel_run "sleep 5" >/dev/null 2>&1
EOF
NEW_DIR_MARKER_8="$(mktemp)"
(
  set -m
  TMPDIR="$PRIVATE_TMP_8" bash "$CHILD8" &
  CHILD8_PID=$!
  sleep 0.4
  NEW_DIR="$(new_tmpdir_since "$BEFORE_8" "$PRIVATE_TMP_8")"
  if [ -n "$NEW_DIR" ] && [ -d "$NEW_DIR" ]; then
    printf '%s\n' "$NEW_DIR" > "$NEW_DIR_MARKER_8"
    kill -INT "$CHILD8_PID" 2>/dev/null || true
    sleep 0.5
    wait "$CHILD8_PID" 2>/dev/null || true
  else
    kill "$CHILD8_PID" 2>/dev/null || true
    wait "$CHILD8_PID" 2>/dev/null || true
  fi
) 2>/dev/null
NEW_DIR_8="$(cat "$NEW_DIR_MARKER_8" 2>/dev/null || true)"
rm -f "$NEW_DIR_MARKER_8"
if [ -n "$NEW_DIR_8" ]; then
  if [ ! -d "$NEW_DIR_8" ]; then
    pass "case 8: SIGINT to the child process still removes its mktemp -d"
  else
    fail "case 8: SIGINT to the child process left [$NEW_DIR_8] behind"
    rm -rf "$NEW_DIR_8"
  fi
else
  fail "case 8: could not observe the in-flight tmp dir at all to test interrupt cleanup"
fi
rm -f "$CHILD8"
rm -rf "$PRIVATE_TMP_8"

# ---------------------------------------------------------------------------
# 9. F3: parallel_run neutralizes `set -e`/`pipefail` around its own
#    internal `xargs -P` pipeline so it is safe for a caller running under
#    `set -euo pipefail` to invoke it with NO `set +e` bracket of its own.
#
#    A caller that is `if`/`&&`/`||`-guarded, or that brackets the call in
#    its own `set +e`, was ALREADY safe before this fix — bash exempts a
#    function called in any of those contexts from `-e` for its own entire
#    body, so that shape cannot discriminate fixed from broken. What DOES
#    discriminate: a fully bare, unguarded call — the shape this finding's
#    own reproducer uses — still aborts the CALLING script at that point
#    (expected: parallel_run legitimately returns 1 for a failed job, and
#    any command returning non-zero as a bare statement aborts a `set -e`
#    script — proven generically below, not specific to this library). What
#    changes is WHERE inside parallel_run that abort happens: broken, the
#    untrapped `xargs` pipeline itself aborts execution before the
#    function ever reaches its own population loop, so PARALLEL_RC and
#    PARALLEL_OUTFILES are never set; fixed, execution always reaches that
#    loop and populates both arrays before parallel_run's own well-defined
#    `return "$rc_overall"` is what triggers the (still expected) abort.
#    Observed via an EXIT trap in a real child bash process, since the
#    aborting script's own variables cannot be inspected any other way
#    after it dies.
# ---------------------------------------------------------------------------

# Generic control: ANY function that returns non-zero, called bare under
# `set -e`, aborts the caller before a following statement runs — this is
# ordinary bash semantics, not a parallel.sh-specific defect, and is not
# what this case's own assertions below are about.
CONTROL_OUT="$(bash -c 'set -e; foo() { return 1; }; foo; echo unreached' 2>&1 || true)"
assert_not_contains "case 9 control: a bare non-zero return always aborts under set -e (generic bash, unrelated to this library)" "$CONTROL_OUT" "unreached"

F3_CHILD="$(mktemp)"
cat > "$F3_CHILD" <<EOF
#!/usr/bin/env bash
set -euo pipefail
source "$LIB"
trap 'printf "RC=%s OUTFILES=%s\n" "\${PARALLEL_RC[*]:-}" "\${#PARALLEL_OUTFILES[@]}"' EXIT
parallel_run "exit 0" "exit 1"
printf 'unreached\n'
EOF
F3_TRAP_OUTPUT="$(bash "$F3_CHILD" 2>&1 || true)"
rm -f "$F3_CHILD"
assert_contains "case 9: a bare unguarded parallel_run call under set -euo pipefail still populates PARALLEL_RC before aborting" "$F3_TRAP_OUTPUT" "RC=0 1"
assert_contains "case 9: ...and PARALLEL_OUTFILES too (both jobs' capture files recorded)" "$F3_TRAP_OUTPUT" "OUTFILES=2"
assert_not_contains "case 9: the bare call still does not reach code after it (parallel_run's own return is what aborts, as documented)" "$F3_TRAP_OUTPUT" "unreached"

# ---------------------------------------------------------------------------
if [ "$FAILURES" -eq 0 ]; then
  printf '\n✓ PASS\n'
  exit 0
fi
printf '\n✗ FAIL — %s failure(s)\n' "$FAILURES" >&2
exit 1
