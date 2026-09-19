#!/usr/bin/env bash
# Assertion harness for run-reproducer.sh.
#
# Builds a temporary worktree-shaped directory per case, runs the guard
# against it, and asserts the exit code (and, where the case cares, that the
# reported message names the defect and that a refused reproducer was never
# actually executed). Follows test-check-panel-reproducers.sh's shape: a
# case_N section per case, a counter, and a non-zero exit when any case
# fails.
#
# `-e` as well as `-u`/`pipefail`, matching this repository's sibling
# harnesses. `set +e`/`set -e` bracket every call expected to exit non-zero,
# because that is the whole point of most cases here and `-e` would
# otherwise abort the suite on the first one.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/run-reproducer.sh"
FAILED=0

# Every case leaves one sandbox worktree behind, removed on exit including
# on a failed assertion. An indexed array, not a space-separated string:
# mktemp paths under TMPDIR may contain spaces, and word-splitting a string
# would leak a sandbox whose path split and `rm -rf` the fragments.
WORKTREES=()
cleanup() {
  [ "${#WORKTREES[@]}" -eq 0 ] && return 0
  for wt in "${WORKTREES[@]}"; do
    rm -rf "$wt"
  done
}
trap cleanup EXIT

# make_worktree -> prints a fresh worktree path with a scripts/ directory
# already inside it. Fails loudly rather than returning an empty path: a
# silent empty path from a failed mktemp is the same stale-sandbox hazard
# test-check-panel-reproducers.sh's own make_worktree guards against.
make_worktree() {
  local wt
  wt="$(mktemp -d "${TMPDIR:-/tmp}/run-reproducer-test.XXXXXX")" || {
    printf 'make_worktree: mktemp failed — aborting suite rather than continuing with a stale path\n' >&2
    exit 1
  }
  WORKTREES+=("$wt")
  mkdir -p "$wt/scripts"
  printf '%s' "$wt"
}

# fixture <worktree> <relative-path> <body> -> writes an executable script
# at <relative-path> inside <worktree>, with a shebang plus <body>. Every
# fixture that is expected to actually run also drops a RAN marker file, so
# a case asserting a reproducer was refused can confirm it never executed —
# not merely that the exit code matched, which a guard broken to always
# refuse would also produce.
fixture() {
  local wt="$1" rel="$2" body="$3"
  mkdir -p "$(dirname "$wt/$rel")"
  {
    printf '#!/usr/bin/env bash\n'
    printf 'touch "%s/RAN"\n' "$wt"
    printf '%s\n' "$body"
  } > "$wt/$rel"
  chmod +x "$wt/$rel"
}

expect_exit() {
  # $1 = label, $2 = expected exit, $3... = command
  local label="$1" want="$2"; shift 2
  local out got
  set +e
  out="$("$@" 2>&1)"; got=$?
  set -e
  if [[ "$got" != "$want" ]]; then
    printf 'FAIL %s: expected exit %s, got %s\n%s\n' "$label" "$want" "$got" "$out"
    FAILED=1
  else
    printf 'ok: %s\n' "$label"
  fi
}

expect_exit_and_names() {
  # $1 = label, $2 = expected exit, $3 = substring the output must contain,
  # $4... = command
  local label="$1" want="$2" needle="$3"; shift 3
  local out got
  set +e
  out="$("$@" 2>&1)"; got=$?
  set -e
  if [[ "$got" != "$want" ]]; then
    printf 'FAIL %s: expected exit %s, got %s\n%s\n' "$label" "$want" "$got" "$out"
    FAILED=1
    return
  fi
  case "$out" in
    *"$needle"*) printf 'ok: %s\n' "$label" ;;
    *)
      printf 'FAIL %s: expected output to name %s, got:\n%s\n' "$label" "$needle" "$out"
      FAILED=1
      ;;
  esac
}

assert_not_ran() {
  # $1 = label, $2 = worktree. A refused reproducer must never have been
  # executed at all — asserted separately from the exit code, which a guard
  # broken to refuse-but-then-run-anyway would still satisfy.
  local label="$1" wt="$2"
  if [ -e "$wt/RAN" ]; then
    printf 'FAIL %s: the reproducer ran despite being refused (found %s/RAN)\n' "$label" "$wt"
    FAILED=1
  else
    printf 'ok: %s (never executed)\n' "$label"
  fi
}

# ===========================================================================
# 1. A contained relative path that exits non-zero — the defect is
#    demonstrated, exit 0 from the runner.
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/fails.sh" "exit 7"
expect_exit 'case 1: a failing contained reproducer demonstrates the defect' 0 "$GUARD" "$wt" "scripts/fails.sh"

# ===========================================================================
# 2. A contained relative path that exits 0 — the defect is not
#    demonstrated, distinct exit code from case 1's.
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/passes.sh" "exit 0"
expect_exit 'case 2: a passing contained reproducer is distinguished from a failing one' 1 "$GUARD" "$wt" "scripts/passes.sh"

# ===========================================================================
# 3. An absolute path token — refused, never executed.
# ===========================================================================
wt="$(make_worktree)"
expect_exit_and_names 'case 3: an absolute path token is refused' 2 'absolute' "$GUARD" "$wt" "/bin/echo hi"
assert_not_ran 'case 3' "$wt"

# ===========================================================================
# 4. A token containing a `..` segment — refused, never executed.
# ===========================================================================
wt="$(make_worktree)"
expect_exit_and_names 'case 4: a `..` path-token segment is refused' 2 "'..'" "$GUARD" "$wt" "../../../etc/passwd"
assert_not_ran 'case 4' "$wt"

# ===========================================================================
# 5. An argument containing a `..` segment — refused, argument checked as
#    well as path.
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/ok.sh" "exit 0"
expect_exit_and_names 'case 5: a `..` argument segment is refused' 2 "'..'" "$GUARD" "$wt" "scripts/ok.sh ../../../etc/passwd"
assert_not_ran 'case 5' "$wt"

# ===========================================================================
# 6. Every banned character, INDIVIDUALLY, is refused — a loop rather than
#    one near-identical case per character, mirroring
#    test-check-panel-reproducers.sh's own case 34/35 shape, including a
#    positive control so the loop cannot pass vacuously against a guard
#    broken to refuse everything. Both quote characters are in the set.
# ===========================================================================
banned_metachars='|;&$`<>(){}~*?[]#\'\''"'
mc_i=0
while [ "$mc_i" -lt "${#banned_metachars}" ]; do
  mc_c="${banned_metachars:$mc_i:1}"
  wt="$(make_worktree)"
  expect_exit_and_names "case 6.$mc_i: banned metacharacter '$mc_c' alone is refused" 2 'metacharacter' "$GUARD" "$wt" "scripts/x${mc_c}sh"
  assert_not_ran "case 6.$mc_i" "$wt"
  mc_i=$((mc_i + 1))
done

wt="$(make_worktree)"
fixture "$wt" "scripts/ok.sh" "exit 0"
expect_exit 'case 6 control: an ordinary command with no banned metacharacter is accepted' 1 "$GUARD" "$wt" "scripts/ok.sh"

# ===========================================================================
# 7. A path that does not exist inside the worktree — refused, never
#    executed.
# ===========================================================================
wt="$(make_worktree)"
expect_exit_and_names 'case 7: a nonexistent path is refused' 2 'does not exist' "$GUARD" "$wt" "scripts/nonexistent.sh"
assert_not_ran 'case 7' "$wt"

# ===========================================================================
# 8. A path escaping the worktree through a symlink — refused. This is the
#    case check-panel-reproducers.sh's own record-format guard deliberately
#    cannot decide (it never resolves a token against a filesystem), and the
#    runner is where it is decided, because the runner always has a real
#    worktree to resolve against.
# ===========================================================================
wt="$(make_worktree)"
outside="$(mktemp -d "${TMPDIR:-/tmp}/run-reproducer-test-outside.XXXXXX")"
WORKTREES+=("$outside")
fixture "$outside" "secret.sh" "exit 0"
ln -s "$outside/secret.sh" "$wt/scripts/escape.sh"
expect_exit_and_names 'case 8: a symlink escaping the worktree is refused' 2 'outside the worktree' "$GUARD" "$wt" "scripts/escape.sh"
if [ -e "$outside/RAN" ]; then
  printf 'FAIL case 8: the symlink target ran despite being refused\n'
  FAILED=1
else
  printf 'ok: case 8 (never executed)\n'
fi

# ===========================================================================
# 9. A command that runs longer than the bound — killed, reported
#    unverifiable, never read as either a pass or a fail. RUN_REPRODUCER_*
#    overrides shrink the bound and grace for this case alone, the same
#    test-economics idiom check-cleanup-complete.sh's own
#    CHECK_CLEANUP_SURVIVORS_TIMEOUT exists for: monotone in the safe
#    direction, so it can only ever produce MORE timeouts, never fewer.
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/hangs.sh" "sleep 30"
set +e
out9="$(RUN_REPRODUCER_BOUND_SECONDS=2 RUN_REPRODUCER_GRACE_SECONDS=1 "$GUARD" "$wt" "scripts/hangs.sh" 2>&1)"
got9=$?
set -e
if [ "$got9" -eq 3 ] && [[ "$out9" == *'unverifiable'* ]]; then
  printf 'ok: %s\n' 'case 9: a command past the bound is killed and reported unverifiable'
else
  printf 'FAIL %s: expected exit 3 naming unverifiable, got exit %s\n%s\n' 'case 9: a command past the bound is killed and reported unverifiable' "$got9" "$out9"
  FAILED=1
fi

# ===========================================================================
# 10. A command that detaches itself and survives the process-group kill —
#     reported as a surviving process with its pid, and the worktree
#     re-checked. The fixture double-forks via python3's os.fork/os.setsid,
#     since this platform ships neither a `setsid` binary nor a `ps -o sid=`
#     keyword (the citation task 7 exists to correct) and `ps -o sess=`
#     was measured to report 0 for every process tried, real session or
#     not — so the parent (tracked by this script's own process group) sleeps
#     past the bound and gets killed, while its detached child keeps
#     sleeping under a session and process group of its own.
#
#     The detached child also ignores SIGTERM (SIG_IGN), which is what
#     makes "still alive after the retry" deterministic rather than a race:
#     an ordinary child that lets SIGTERM kill it dies during the grace
#     window every time it was tried here, before the runner's SIGKILL is
#     ever sent, and the whole point of this case is exercising the SIGKILL
#     path and the immediate liveness check right after it. Measured on
#     Darwin 25.5.0: a SIGTERM-ignoring, launchd-reparented child, sent
#     SIGKILL and checked with `kill -0` immediately afterward — no
#     sleep — still answered "alive" 5 runs out of 5, because the process
#     table entry is not yet reclaimed by its new parent at that instant.
#     That is the real, narrow window run-reproducer.sh's own final check
#     relies on, and this fixture is what makes it observable on demand
#     rather than left to chance.
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/detach.sh" 'exec python3 -c "
import os, time, signal
if os.fork() == 0:
    os.setsid()
    signal.signal(signal.SIGTERM, signal.SIG_IGN)
    time.sleep(30)
    os._exit(0)
else:
    time.sleep(30)
"'
set +e
out10="$(RUN_REPRODUCER_BOUND_SECONDS=2 RUN_REPRODUCER_GRACE_SECONDS=1 "$GUARD" "$wt" "scripts/detach.sh" 2>&1)"
got10=$?
set -e
if [ "$got10" -eq 3 ] && [[ "$out10" == *'surviving process'* ]] && [[ "$out10" =~ pid\(s\):\ [0-9]+ ]]; then
  printf 'ok: %s\n' 'case 10: a detached survivor is named with its pid'
else
  printf 'FAIL %s: expected exit 3 naming a surviving process with a pid, got exit %s\n%s\n' 'case 10: a detached survivor is named with its pid' "$got10" "$out10"
  FAILED=1
fi
# Best-effort cleanup of whatever survived the case, so a failed assertion
# does not leave a sleeping process behind on the machine running the suite.
pkill -f "scripts/detach.sh" >/dev/null 2>&1 || true

# ===========================================================================
# 11. A passing reproducer's own diagnostic output is carried back by the
#     runner, delimited from the runner's own messages — the bounce this
#     guards against has nothing to carry otherwise, per finding F50.
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/passes-with-output.sh" 'echo "some diagnostic output here"
exit 0'
set +e
out11="$("$GUARD" "$wt" "scripts/passes-with-output.sh" 2>&1)"
got11=$?
set -e
if [ "$got11" -eq 1 ] \
  && [[ "$out11" == *'some diagnostic output here'* ]] \
  && [[ "$out11" == *'captured reproducer output begin'* ]] \
  && [[ "$out11" == *'captured reproducer output end'* ]]; then
  printf 'ok: %s\n' 'case 11: a passing reproducer'"'"'s output is captured and delimited'
else
  printf 'FAIL %s: expected exit 1 with delimited output naming the diagnostic text, got exit %s\n%s\n' 'case 11: a passing reproducer'"'"'s output is captured and delimited' "$got11" "$out11"
  FAILED=1
fi

# ===========================================================================
# 12. A tab, not a space, separates the path token from its argument — the
#     path token must still resolve to the fixture rather than to the whole
#     tab-joined string, per finding F49.
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/ok.sh" 'exit 0'
expect_exit $'case 12: a tab-separated command line still resolves the path token' 1 "$GUARD" "$wt" $'scripts/ok.sh\t--strict'

# ===========================================================================
# 13. A reproducer that forks a detached child and exits WELL WITHIN the
#     bound (no timeout at all) still leaks that child unless every poll,
#     not only the one at the deadline, watches for it — finding F51. Same
#     SIGTERM-ignoring, launchd-reparented child as case 10, but the parent
#     here exits quickly on its own rather than being killed at a bound.
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/detach-early.sh" 'exec python3 -c "
import os, time, signal
if os.fork() == 0:
    os.setsid()
    signal.signal(signal.SIGTERM, signal.SIG_IGN)
    time.sleep(30)
    os._exit(0)
else:
    time.sleep(2)
"'
set +e
out13="$(RUN_REPRODUCER_BOUND_SECONDS=8 RUN_REPRODUCER_GRACE_SECONDS=1 "$GUARD" "$wt" "scripts/detach-early.sh" 2>&1)"
got13=$?
set -e
if [ "$got13" -eq 3 ] && [[ "$out13" == *'surviving process'* ]] && [[ "$out13" == *'forked a detached child'* ]]; then
  printf 'ok: %s\n' 'case 13: a child detached before the parent exits normally is still caught'
else
  printf 'FAIL %s: expected exit 3 naming a surviving process from a non-timeout exit, got exit %s\n%s\n' 'case 13: a child detached before the parent exits normally is still caught' "$got13" "$out13"
  FAILED=1
fi
pkill -f "scripts/detach-early.sh" >/dev/null 2>&1 || true

# ===========================================================================
# 14. A REAL double fork — not a python os.fork()/os.setsid() single-level
#     detach as in cases 10 and 13, but a genuine two-level shell tree: the
#     fixture backgrounds a subshell, that subshell itself backgrounds a
#     sleep (the grandchild) and then exits, well within the bound, on its
#     own — no timeout, no process-group escape needed to demonstrate this.
#     A single-level `pgrep -P "$CMD_PID"` never sees the grandchild at
#     all: by the time any poll could run, the subshell in between is
#     already gone, and the grandchild was never $CMD_PID's own direct
#     child to begin with. Finding F54. The grandchild ignores SIGTERM
#     (`trap "" TERM`) so the retry-kill's SIGKILL path is exercised and
#     the survivor is caught in the same not-yet-reaped window cases 10 and
#     13 rely on, rather than dying quietly during the grace window.
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/doublefork.sh" '( bash -c "trap \"\" TERM; sleep 30" & sleep 0.6 ) &
sleep 1
exit 0'
set +e
out14="$(RUN_REPRODUCER_BOUND_SECONDS=5 RUN_REPRODUCER_GRACE_SECONDS=1 "$GUARD" "$wt" "scripts/doublefork.sh" 2>&1)"
got14=$?
set -e
grandchild_pid="$(printf '%s\n' "$out14" | grep -oE 'pid\(s\): [0-9]+' | grep -oE '[0-9]+' || true)"
if [ "$got14" -eq 3 ] && [[ "$out14" == *'surviving process'* ]] && [ -n "$grandchild_pid" ]; then
  printf 'ok: %s\n' 'case 14: a real double-forked grandchild is named with its pid'
else
  printf 'FAIL %s: expected exit 3 naming a surviving grandchild pid, got exit %s\n%s\n' 'case 14: a real double-forked grandchild is named with its pid' "$got14" "$out14"
  FAILED=1
fi
# Best-effort cleanup of whatever survived the case, so a failed assertion
# does not leave a sleeping process behind on the machine running the
# suite — mirroring cases 10 and 13's own pkill below.
[ -n "$grandchild_pid" ] && kill -KILL "$grandchild_pid" >/dev/null 2>&1 || true
pkill -f 'trap .* TERM; sleep 30' >/dev/null 2>&1 || true

# ===========================================================================
# 15. A missing scripts/reproducer-metachars.sh is reported as "cannot
#     answer" (exit 4), not "defect not demonstrated" (exit 1) — finding
#     F56. This script's own `source` of that file, with no readability
#     check ahead of it, let `set -e` abort the script at exit 1 on a
#     missing dependency: this script's own "ran to completion and exited
#     0" code, so a caller reading exit 1 as a verified pass-through would
#     be trusting a run that never happened. Exercised against a REAL copy
#     of this script, sitting in its own sandbox directory, deliberately
#     missing reproducer-metachars.sh beside it, rather than editing the
#     real scripts/ directory this suite runs from.
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/ok.sh" "exit 0"
MISSING_DEP_DIR="$(mktemp -d "${TMPDIR:-/tmp}/run-reproducer-test-missing-dep.XXXXXX")"
WORKTREES+=("$MISSING_DEP_DIR")
cp "$GUARD" "$MISSING_DEP_DIR/run-reproducer.sh"
expect_exit_and_names 'case 15: a missing reproducer-metachars.sh is cannot-answer, not a verdict' 4 'cannot read' \
  "$MISSING_DEP_DIR/run-reproducer.sh" "$wt" "scripts/ok.sh"
assert_not_ran 'case 15' "$wt"

# ===========================================================================
# 16. A FAST double fork — the intermediate exits before run-reproducer.sh's
#     very first poll can ever observe the grandchild as $CMD_PID's own
#     descendant, unlike case 14's slower two-level fork (whose intermediate
#     lives 0.6s, long enough for at least one 0.2s poll to see it via
#     collect_descendants). Here the intermediate backgrounds the
#     grandchild and exits with no delay at all: before task 6's fix, this
#     re-parents the grandchild to launchd before collect_descendants' very
#     first call can ever run, and a parentage walk can never see it again
#     afterwards — measured directly against the unpatched script, which
#     returned exit 1 "defect not demonstrated" here while leaving a real
#     `sleep 31` process running under launchd. Only the process group
#     task 6 establishes at launch (D4) — which survives re-parenting,
#     since re-parenting changes ppid, never pgid — can still find it. The
#     grandchild ignores SIGTERM so the SIGKILL path is exercised
#     deterministically, matching cases 10, 13 and 14's own fixtures.
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/doublefork-fast.sh" '( bash -c "trap \"\" TERM; sleep 31" & )
exit 0'
set +e
out16="$(RUN_REPRODUCER_GRACE_SECONDS=1 "$GUARD" "$wt" "scripts/doublefork-fast.sh" 2>&1)"
got16=$?
set -e
fastfork_pid="$(printf '%s\n' "$out16" | grep -oE 'pid\(s\): [0-9]+' | grep -oE '[0-9]+' || true)"
if [ "$got16" -eq 3 ] && [[ "$out16" == *'surviving process'* ]] && [ -n "$fastfork_pid" ]; then
  printf 'ok: %s\n' 'case 16: a fast double-forked grandchild that re-parents before the first poll is still named with its pid'
else
  printf 'FAIL %s: expected exit 3 naming a surviving grandchild pid, got exit %s\n%s\n' 'case 16: a fast double-forked grandchild that re-parents before the first poll is still named with its pid' "$got16" "$out16"
  FAILED=1
fi
# Best-effort cleanup mirroring cases 10, 13 and 14's own pkill below, plus a
# direct kill of the pid the guard itself named, so a failing assertion
# still leaves nothing running on the machine running the suite.
[ -n "$fastfork_pid" ] && kill -KILL "$fastfork_pid" >/dev/null 2>&1 || true
pkill -9 -f 'trap .* TERM; sleep 31' >/dev/null 2>&1 || true

# ===========================================================================
# 17. Control: an ordinary, non-forking reproducer still demonstrates the
#     defect exactly as case 1 does, and leaves no trace of itself running
#     afterward — proving the process-group wrapper task 6 adds at the exec
#     point did not change the normal, non-forking path any of cases 1-15
#     already exercise. The reproducer's own marker argument, unique per
#     run, is passed on its command line so a lingering copy of the exec'd
#     process (a broken shim leaving a zombie behind, for instance) would
#     still be found by `pgrep -f` even after this script's own $GUARD call
#     has returned and its own bookkeeping is done.
# ===========================================================================
wt="$(make_worktree)"
marker17="run-reproducer-test-case17-$$-$RANDOM"
fixture "$wt" "scripts/ordinary17.sh" "exit 5"
expect_exit 'case 17: an ordinary non-forking reproducer still demonstrates the defect' 0 "$GUARD" "$wt" "scripts/ordinary17.sh $marker17"
if pgrep -f "$marker17" >/dev/null 2>&1; then
  printf 'FAIL %s: a process carrying the reproducer'"'"'s own marker survived the run\n' 'case 17'
  FAILED=1
else
  printf 'ok: %s\n' 'case 17: no process is left behind'
fi


# ===========================================================================
# 18. F4: the SAME fast double fork as case 16, but this time the top-level
#     reproducer does NOT exit on its own — it sleeps well past the bound,
#     so run-reproducer.sh actually takes the TIMEOUT branch (unlike case
#     16, which pins the early-exit path). The grandchild ignores SIGTERM
#     and stays in $CMD_PID's own process group (no setsid of its own), and
#     its immediate parent exits with no delay, re-parenting it before the
#     very first poll — exactly the shape only the pre-kill `pgrep -g`
#     snapshot can see, per case 16's own comment.
#
#     Before task 6's restructuring, that snapshot was taken AFTER the
#     timeout branch's own blind `kill -TERM -PGID` / `kill -KILL -PGID`
#     had already run: by the time the snapshot's `pgrep -g` call fired,
#     the grandchild had already been signalled (and was often already
#     reaped), so it was silently absent from the snapshot, absent from
#     $KIDS, and absent from the operator report — an exit 3 that never
#     said "surviving process", even though a real process only died
#     because a blind, unnamed kill happened to reach it. This case pins
#     exactly that path and must fail against the unfixed script.
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/doublefork-timeout.sh" '( bash -c "trap \"\" TERM; sleep 30" & )
sleep 30'
set +e
out18="$(RUN_REPRODUCER_BOUND_SECONDS=2 RUN_REPRODUCER_GRACE_SECONDS=1 "$GUARD" "$wt" "scripts/doublefork-timeout.sh" 2>&1)"
got18=$?
set -e
timeoutfork_pid="$(printf '%s\n' "$out18" | grep -oE 'pid\(s\): [0-9]+' | grep -oE '[0-9]+' || true)"
if [ "$got18" -eq 3 ] && [[ "$out18" == *'surviving process'* ]] && [ -n "$timeoutfork_pid" ]; then
  printf 'ok: %s\n' 'case 18: a fast double-forked grandchild is still named on the timeout path'
else
  printf 'FAIL %s: expected exit 3 naming a surviving grandchild pid on the timeout path, got exit %s\n%s\n' 'case 18: a fast double-forked grandchild is still named on the timeout path' "$got18" "$out18"
  FAILED=1
fi
# Best-effort cleanup mirroring case 16's own, so a failing assertion still
# leaves nothing running on the machine running the suite.
[ -n "$timeoutfork_pid" ] && kill -KILL "$timeoutfork_pid" >/dev/null 2>&1 || true
pkill -9 -f 'trap .* TERM; sleep 30' >/dev/null 2>&1 || true

# ===========================================================================
# 19. KAN-524: the SAME reproducer verdict the caller observed pre-fix, seen
#     again post-fix, is refused as ambiguous — exit 2 naming the expected
#     convention. The pre-fix verdict is carried in the script's own verdict
#     vocabulary (0 = defect demonstrated, 1 = not demonstrated), which is
#     what the panel parent already holds from its `; echo "F<n>: exit $?"`
#     record of the dispatch-time run. This refusal fires AFTER the
#     reproducer ran — unlike every shape refusal above, the verdict here is
#     read from a real execution, so the fixture's RAN marker MUST exist.
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/fails-again.sh" "exit 7"
expect_exit_and_names 'case 19: a post-fix verdict identical to the pre-fix verdict (0/0) is refused' 2 'convention' \
  "$GUARD" "$wt" "scripts/fails-again.sh" --pre-fix-exit 0
if [ -e "$wt/RAN" ]; then
  printf 'ok: case 19 (the reproducer ran before the verdict was read)\n'
else
  printf 'FAIL case 19: the ambiguity refusal answered without running the reproducer\n'
  FAILED=1
fi

# ===========================================================================
# 20. The mirrored ambiguity: 1/1 — the reproducer exited 0 pre-fix and
#     exits 0 post-fix again. Identical under either convention, refused
#     the same way.
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/passes-again.sh" "exit 0"
expect_exit_and_names 'case 20: a post-fix verdict identical to the pre-fix verdict (1/1) is refused' 2 'convention' \
  "$GUARD" "$wt" "scripts/passes-again.sh" --pre-fix-exit 1

# ===========================================================================
# 21. The healthy flip: demonstrated pre-fix (verdict 0), not demonstrated
#     post-fix (verdict 1) — the fix verified, exit 1, no refusal.
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/now-passes.sh" "exit 0"
expect_exit 'case 21: a flipped verdict (0 pre-fix, 1 post-fix) is the fix verified' 1 \
  "$GUARD" "$wt" "scripts/now-passes.sh" --pre-fix-exit 0

# ===========================================================================
# 22. The opposite flip (1 pre-fix, 0 post-fix) is not identical either —
#     the verdict passes through unchanged, no refusal.
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/now-fails.sh" "exit 5"
expect_exit 'case 22: a flipped verdict (1 pre-fix, 0 post-fix) passes through' 0 \
  "$GUARD" "$wt" "scripts/now-fails.sh" --pre-fix-exit 1

# ===========================================================================
# 23. Anything but the script's own two verdict codes is a usage failure —
#     exit 4, the reproducer never run. A number the verdict vocabulary
#     cannot answer (2, 3, 4 are this script's own refusal/unverifiable/
#     cannot-answer codes, not the reproducer's verdict), a non-numeric
#     token, an empty value, a missing value and a stray third argument are
#     all reported, never ignored and never read as "no flag".
# ===========================================================================
for bad in 2 3 7 -1 zero '' ' '; do
  wt="$(make_worktree)"
  fixture "$wt" "scripts/never-runs.sh" "exit 0"
  if [ -z "$bad" ]; then
    # An empty value with the flag present means a missing value: the
    # argument vector ends after the flag name.
    expect_exit_and_names "case 23.$bad: a missing --pre-fix-exit value is a usage failure" 4 'usage' \
      "$GUARD" "$wt" "scripts/never-runs.sh" --pre-fix-exit
  else
    expect_exit_and_names "case 23.$bad: --pre-fix-exit '$bad' is a usage failure" 4 'usage' \
      "$GUARD" "$wt" "scripts/never-runs.sh" --pre-fix-exit "$bad"
  fi
  assert_not_ran "case 23.$bad" "$wt"
done
wt="$(make_worktree)"
fixture "$wt" "scripts/never-runs.sh" "exit 0"
expect_exit_and_names 'case 23.stray: a stray third argument is a usage failure' 4 'usage' \
  "$GUARD" "$wt" "scripts/never-runs.sh" extra-arg
assert_not_ran 'case 23.stray' "$wt"

# ===========================================================================
# 24. KAN-568: a reproducer declaring the mutation-reproducer convention —
#     the exact line `# mutation-reproducer` within its first 10 lines — is
#     read under that convention: the mutation lands and the build (test
#     suite) still SUCCEEDS, which is the surviving mutant, the bug present.
#     Exit 0 under the mutation convention is therefore "defect
#     demonstrated" (runner exit 0) — the exact inversion that made kan-485's
#     round-0 mutation reproducers hand-verified substitutions.
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/mutation-survives.sh" '# mutation-reproducer
exit 0'
expect_exit 'case 24: a mutation reproducer exiting 0 demonstrates the defect' 0 \
  "$GUARD" "$wt" "scripts/mutation-survives.sh"

# ===========================================================================
# 25. The mutation convention's not-demonstrated side: the mutation lands
#     and the build FAILS — a test caught it — so the reproducer exits
#     non-zero and the defect is NOT demonstrated (runner exit 1), the exact
#     reading the generic convention would have called "demonstrated".
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/mutation-caught.sh" '# mutation-reproducer
exit 3'
expect_exit 'case 25: a mutation reproducer exiting non-zero is not demonstrated' 1 \
  "$GUARD" "$wt" "scripts/mutation-caught.sh"

# ===========================================================================
# 26. The ambiguity refusal under the mutation convention: verdict 0 pre-fix
#     (the bug was present, the build succeeded) and verdict 0 again
#     post-fix from the SAME exit 0 — refused, exit 2, after a real run.
#     Pins that the refusal compares the convention-mapped VERDICT, not the
#     raw exit code: a raw-RC comparison against a mutation reproducer's
#     exit 0 would answer "not demonstrated" (1) and never refuse here.
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/mutation-still-survives.sh" '# mutation-reproducer
exit 0'
expect_exit_and_names 'case 26: a mutation reproducer with verdict 0 both sides is refused as ambiguous' 2 'convention' \
  "$GUARD" "$wt" "scripts/mutation-still-survives.sh" --pre-fix-exit 0
if [ -e "$wt/RAN" ]; then
  printf 'ok: case 26 (the reproducer ran before the verdict was read)\n'
else
  printf 'FAIL case 26: the ambiguity refusal answered without running the reproducer\n'
  FAILED=1
fi

# ===========================================================================
# 27. The healthy flip under the mutation convention: verdict 0 pre-fix, the
#     fix makes the build FAIL against the same mutation (exit 3), verdict 1
#     — the fix verified, exit 1, no refusal.
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/mutation-now-caught.sh" '# mutation-reproducer
exit 3'
expect_exit 'case 27: a mutation reproducer flipping 0 to 1 is the fix verified' 1 \
  "$GUARD" "$wt" "scripts/mutation-now-caught.sh" --pre-fix-exit 0

# ===========================================================================
# 28. The declaration is the exact line, within the window: a marker line
#     beyond the first 10 lines, or one carrying more text than the marker,
#     is NOT a declaration — the generic convention applies (exit 0 here is
#     "defect not demonstrated", runner exit 1). A grep over the whole file,
#     or a prefix match, would read this reproducer inverted.
# ===========================================================================
wt="$(make_worktree)"
filler=""
i=0
while [ "$i" -lt 12 ]; do filler="${filler}# filler line
"; i=$((i + 1)); done
fixture "$wt" "scripts/late-marker.sh" "${filler}# mutation-reproducer
exit 0"
expect_exit 'case 28.a: a marker beyond the first 10 lines declares nothing' 1 \
  "$GUARD" "$wt" "scripts/late-marker.sh"
wt="$(make_worktree)"
fixture "$wt" "scripts/marker-with-suffix.sh" '# mutation-reproducer: because the build succeeded
exit 0'
expect_exit 'case 28.b: a marker line carrying extra text declares nothing' 1 \
  "$GUARD" "$wt" "scripts/marker-with-suffix.sh"

# ===========================================================================
# 29. The reproducer sha (KAN-568 review, F1): every verdict carries a
#     `reproducer sha <hex>` line naming the file's SHA-256, and a
#     --reproducer-sha pin matching that sha is accepted and decides
#     normally.
# ===========================================================================
source "$SCRIPT_DIR/lib/sha256-hex.sh"
wt="$(make_worktree)"
fixture "$wt" "scripts/sha-ok.sh" "exit 5"
want_sha="$(sha256_hex_file "$wt/scripts/sha-ok.sh")"
set +e
out29="$("$GUARD" "$wt" "scripts/sha-ok.sh" --reproducer-sha "$want_sha" 2>&1)"
got29=$?
set -e
if [ "$got29" -eq 0 ] && [[ "$out29" == *"reproducer sha $want_sha"* ]]; then
  printf 'ok: %s\n' 'case 29: a matching --reproducer-sha pin is accepted and the sha is printed with the verdict'
else
  printf 'FAIL %s: expected exit 0 printing reproducer sha %s, got exit %s\n%s\n' 'case 29: a matching --reproducer-sha pin is accepted and the sha is printed with the verdict' "$want_sha" "$got29" "$out29"
  FAILED=1
fi

# ===========================================================================
# 30. A pin naming a DIFFERENT file is refused before execution — the
#     never-executed shape class, exit 2 — with both shas named. This is
#     the fix-round guard against a comment-only re-authoring of a
#     dispatched reproducer between its dispatch-time run and its re-run.
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/sha-pinned.sh" "exit 5"
expect_exit_and_names 'case 30: a mismatched --reproducer-sha pin is refused without running' 2 'pinned sha' \
  "$GUARD" "$wt" "scripts/sha-pinned.sh" --reproducer-sha 0000000000000000000000000000000000000000000000000000000000000000
assert_not_ran 'case 30' "$wt"

# ===========================================================================
# 31. A pin value that is not a hex string is a usage failure — exit 4, the
#     reproducer never run.
# ===========================================================================
wt="$(make_worktree)"
fixture "$wt" "scripts/sha-never-runs.sh" "exit 0"
expect_exit_and_names 'case 31: a non-hex --reproducer-sha value is a usage failure' 4 'usage' \
  "$GUARD" "$wt" "scripts/sha-never-runs.sh" --reproducer-sha zz-nothex
assert_not_ran 'case 31' "$wt"

# ===========================================================================
# 32-33. THE BASH 3.2 FLOOR (macOS's own /bin/bash — the floor
#     scripts/lib/coverage.sh's header states). The guard crashed there
#     twice: "${ARGS[@]}" on an EMPTY array is unbound-variable under its
#     own `set -u` (a single-token reproducer died at the argument loop,
#     exit 1 — a crash this script's exit-code vocabulary mis-reads as
#     "defect not demonstrated"), and `exec {fd}<>file` is bash-4.1+ syntax
#     3.2 parses as an exec of a command literally named `{fd}` (a
#     two-token reproducer died at the sentinel with a wrong verdict). Both
#     cases run the guard under /bin/bash itself, one per path.
# ===========================================================================
if [ -x /bin/bash ]; then
  wt="$(make_worktree)"
  fixture "$wt" "scripts/floor-single-token.sh" "exit 3"
  expect_exit_and_names 'case 32: a single-token reproducer decides under /bin/bash (empty ARGS)' 0 \
    'defect demonstrated' /bin/bash "$GUARD" "$wt" "scripts/floor-single-token.sh"
  wt="$(make_worktree)"
  fixture "$wt" "scripts/floor-two-token.sh" "exit 3"
  expect_exit_and_names 'case 33: a two-token reproducer decides under /bin/bash (sentinel fd)' 0 \
    'defect demonstrated' /bin/bash "$GUARD" "$wt" "scripts/floor-two-token.sh arg1"
else
  printf 'skip: cases 32-33 — no /bin/bash on this machine\n'
fi

if [ "$FAILED" -ne 0 ]; then
  printf 'test-run-reproducer: one or more cases failed\n' >&2
  exit 1
fi
printf 'test-run-reproducer: all cases plus the metacharacter loop and its control pass\n'
