#!/usr/bin/env bash
# run-guard-tests.sh — discovers scripts/test-*.sh and runs them concurrently
# through scripts/lib/parallel.sh (KAN-362 task 1's spawn/capture/replay
# primitive), per design.md's glob-discovers-harnesses and
# replay-failures-only decisions.
#
# OUTPUT SHAPE: one `ok`/`FAIL` line per harness with its wall time, then a
# summary naming the total, the passed count, the failed count and the wall
# clock; then, only when at least one harness failed, each failing harness's
# captured stdout+stderr replayed in full and contiguously (never
# interleaved with another harness's), each block introduced by the
# harness's own name so a reader can tell which is which. A clean run never
# dumps a harness's output — see design.md's replay-failures-only decision,
# rejecting a thousand lines of noise on every green run.
#
# DISCOVERY IS BY GLOB, scripts/test-*.sh, flat — never recursive, and never
# a hand-maintained list — so a new harness nobody wires into .flow/project.md
# is never silently skipped. Verified byte-identical to today's 40 lines by
# design.md's own measurement.
#
# RUN_GUARD_TESTS_ROOT (design.md's runner-root-override): an explicit,
# opt-in override honoured only when set — copied verbatim from
# check-references.sh's own CHECK_REFERENCES_ROOT idiom rather than writing
# a third variant of it. This is load-bearing, not decoration:
# scripts/test-run-guard-tests.sh is itself matched by this runner's own
# glob, so without a fixture override its own harness would re-enter the
# real suite.
#
# Bash 3.2 is the floor: indexed arrays only, no associative arrays, no
# `wait -n`. scripts/lib/parallel.sh installs its own process-wide
# EXIT/INT/TERM traps at source time (see its header) — this script installs
# no competing trap.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LIB="$SCRIPT_DIR/lib/parallel.sh"

if [ -n "${RUN_GUARD_TESTS_ROOT:-}" ]; then
  TEST_ROOT="$RUN_GUARD_TESTS_ROOT"
elif [ "${RUN_GUARD_TESTS_ROOT+set}" = "set" ]; then
  printf 'RUN_GUARD_TESTS_ROOT is set but empty\n' >&2
  exit 2
else
  TEST_ROOT="$SCRIPT_DIR"
fi

if [ ! -d "$TEST_ROOT" ]; then
  printf 'run-guard-tests: not a directory: %s\n' "$TEST_ROOT" >&2
  exit 2
fi

# shellcheck source=lib/parallel.sh
source "$LIB"

HARNESSES=()
for f in "$TEST_ROOT"/test-*.sh; do
  [ -e "$f" ] || continue
  HARNESSES+=("$f")
done

if [ "${#HARNESSES[@]}" -eq 0 ]; then
  printf 'run-guard-tests: no test-*.sh harnesses found under %s\n' "$TEST_ROOT" >&2
  exit 2
fi

# TIME_DIR holds one <index>.time file per job, written by the job's own
# command string (never by parallel.sh, which owns only stdout+stderr
# capture) so per-harness wall time is available without adding a single
# extra byte to a harness's own captured output — that output is replayed
# verbatim on failure, and an injected timing line would break that.
# Cleaned up explicitly at the end of a normal run; a mktemp -d leaking on
# SIGINT here is a trivial cost next to parallel.sh's own cleanup of its
# (much larger) capture directory, which its process-wide trap already owns.
TIME_DIR="$(mktemp -d "${TMPDIR:-/tmp}/run-guard-tests-time.XXXXXX")"

SECONDS=0

CMDS=()
i=0
for h in "${HARNESSES[@]}"; do
  printf -v qh '%q' "$h"
  printf -v qtime '%q' "$TIME_DIR/$i.time"
  CMDS+=("S=\$(date +%s); bash $qh; RC=\$?; E=\$(date +%s); printf '%s\n' \$((E - S)) > $qtime; exit \$RC")
  i=$((i + 1))
done

set +e
parallel_run "${CMDS[@]}"
RUN_RC=$?
set -e

if [ "$RUN_RC" -eq 2 ]; then
  rm -rf "$TIME_DIR"
  printf 'run-guard-tests: could not run the harnesses (see above)\n' >&2
  exit 2
fi

PASSED=0
FAILED=0
FAILED_NAMES=()
NAMES=()
i=0
for h in "${HARNESSES[@]}"; do
  name="$(basename "$h")"
  NAMES+=("$name")
  rc="${PARALLEL_RC[$i]}"
  elapsed="$(cat "$TIME_DIR/$i.time" 2>/dev/null || printf '?')"
  if [ "$rc" -eq 0 ]; then
    printf 'ok:   %s (%ss)\n' "$name" "$elapsed"
    PASSED=$((PASSED + 1))
  else
    printf 'FAIL: %s (%ss)\n' "$name" "$elapsed"
    FAILED=$((FAILED + 1))
    FAILED_NAMES+=("$name")
  fi
  i=$((i + 1))
done

rm -rf "$TIME_DIR"

TOTAL=${#HARNESSES[@]}
WALL=$SECONDS

printf '\n%s harnesses, %s passed, %s failed, %ss wall\n' "$TOTAL" "$PASSED" "$FAILED" "$WALL"

if [ "$FAILED" -eq 0 ]; then
  exit 0
fi

printf '\nFAILED:\n'
for name in "${FAILED_NAMES[@]}"; do
  printf '  %s\n' "$name"
done

printf '\n'
parallel_replay_failures "${NAMES[@]}"

exit 1
