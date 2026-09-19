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
# COMPANION PRESENCE (kan-387): every check-*.sh guard in TEST_ROOT gets a
# test-check-*.sh companion mutation harness — the repo's convention since
# KAN-197, until now enforced only by review-panel judgment. A guard whose
# companion is missing is refused before anything runs: one stderr line per
# gap naming the guard and the missing harness, then exit 1 — a detected
# violation of the suite's contract, the same code a failing harness gets,
# never exit 2's "cannot answer". The scan is the same flat TEST_ROOT glob
# the harness discovery uses (never recursive, never a hand-maintained
# list), which is also what lets this runner's own harness exercise the
# check through a RUN_GUARD_TESTS_ROOT fixture.
#
# RUN_GUARD_TESTS_ROOT (design.md's runner-root-override): an explicit,
# opt-in override honoured only when set — copied verbatim from
# check-references.sh's own CHECK_REFERENCES_ROOT idiom rather than writing
# a third variant of it. This is load-bearing, not decoration:
# scripts/test-run-guard-tests.sh is itself matched by this runner's own
# glob, so without a fixture override its own harness would re-enter the
# real suite.
#
# GUARD_TESTS_REPO_ROOT (kan-584's watched-repo override): the git work tree
# whose cleanliness gates the suite — the repository that owns this runner
# when unset. Same opt-in-override idiom as RUN_GUARD_TESTS_ROOT, and
# load-bearing for the same reason: scripts/test-run-guard-tests.sh points
# the gate at a fixture repository, never at this repository's own tree.
# Set but empty is refused, exit 2, exactly like its sibling. A watched
# root that is not a git work tree cannot answer the question, so the gate
# is skipped with one stderr line — never a silent pass.
#
# THE TREE GATE (kan-584): before the harnesses launch, the runner snapshots
# `git status --porcelain` over the watched root; after the summary it
# re-reads it. A NET change fails the runner with exit 1, naming the paths
# present after the suite and absent before — a harness case that mutates a
# real shared file is a suite-contract violation, the same code a failing
# harness gets, so "the whole guard suite ran and nothing touched the live
# tree" is a property every green run just proved. Edits made and restored
# inside the run net to zero and are not caught; the gate's contract is the
# tree the suite leaves behind, not a filesystem watch.
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

if [ -n "${GUARD_TESTS_REPO_ROOT:-}" ]; then
  WATCHED="$GUARD_TESTS_REPO_ROOT"
elif [ "${GUARD_TESTS_REPO_ROOT+set}" = "set" ]; then
  printf 'GUARD_TESTS_REPO_ROOT is set but empty\n' >&2
  exit 2
else
  WATCHED="$(git -C "$SCRIPT_DIR" rev-parse --show-toplevel 2>/dev/null || printf '')"
fi

if [ -n "$WATCHED" ] && ! git -C "$WATCHED" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  printf 'run-guard-tests: %s is not a git work tree — skipping the tree gate\n' "$WATCHED" >&2
  WATCHED=""
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

MISSING_COMPANIONS=()
for f in "$TEST_ROOT"/check-*.sh; do
  [ -e "$f" ] || continue
  name="$(basename "$f" .sh)"
  [ -e "$TEST_ROOT/test-$name.sh" ] || MISSING_COMPANIONS+=("$(basename "$f")")
done

if [ "${#MISSING_COMPANIONS[@]}" -gt 0 ]; then
  for guard in "${MISSING_COMPANIONS[@]}"; do
    printf 'run-guard-tests: %s has no companion test-%s.sh\n' "$guard" "${guard%.sh}" >&2
  done
  printf 'run-guard-tests: every check-*.sh guard gets a test-check-*.sh companion — add the missing harnesses\n' >&2
  exit 1
fi

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

if [ -n "$WATCHED" ]; then
  git -C "$WATCHED" status --porcelain > "$TIME_DIR/before.status"
fi

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

# TIME_DIR itself survives until an exit path below: the tree gate's
# after.status snapshot is still to be written into it.
TOTAL=${#HARNESSES[@]}
WALL=$SECONDS

printf '\n%s harnesses, %s passed, %s failed, %ss wall\n' "$TOTAL" "$PASSED" "$FAILED" "$WALL"

DIRTY=0
if [ -n "$WATCHED" ]; then
  git -C "$WATCHED" status --porcelain > "$TIME_DIR/after.status"
  if ! cmp -s "$TIME_DIR/before.status" "$TIME_DIR/after.status"; then
    DIRTY=1
  fi
fi

if [ "$FAILED" -eq 0 ] && [ "$DIRTY" -eq 0 ]; then
  rm -rf "$TIME_DIR"
  exit 0
fi

if [ "$FAILED" -gt 0 ]; then
  printf '\nFAILED:\n'
  for name in "${FAILED_NAMES[@]}"; do
    printf '  %s\n' "$name"
  done

  printf '\n'
  parallel_replay_failures "${NAMES[@]}"
fi

if [ "$DIRTY" -eq 1 ]; then
  printf '\nrun-guard-tests: the watched repository tree changed during the suite (%s):\n' "$WATCHED"
  # Name only the delta: entries present after the suite and absent before.
  # A pure-deletion delta leaves the grep empty, and an empty grep fails the
  # pipeline under pipefail — hence the || true and the explicit fallback.
  DELTA="$(diff "$TIME_DIR/before.status" "$TIME_DIR/after.status" | grep '^> ' | sed 's/^> //' || true)"
  if [ -n "$DELTA" ]; then
    printf '%s\n' "$DELTA"
  else
    printf '  (entries removed relative to the pre-suite snapshot)\n'
  fi
fi

rm -rf "$TIME_DIR"
exit 1
