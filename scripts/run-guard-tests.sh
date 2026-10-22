#!/usr/bin/env bash
# run-guard-tests.sh — discovers scripts/test-*.sh and runs them concurrently
# in flow-guard (stats/internal/guard/runguardtests.go, KAN-921), per
# design.md's glob-discovers-harnesses and replay-failures-only decisions.
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
# list), which is also what lets this runner's own Go test exercise the
# check through a RUN_GUARD_TESTS_ROOT fixture. A guard ported to Go
# (kan-760) is covered by its Go test instead: when the guard's file contains
# `flow_guard_exec <name> ` and
# TEST_ROOT/../stats/internal/guard/<name, - replaced by _>_test.go exists,
# it needs no test-check-*.sh — scripts/test-go-guards.sh runs those tests.
#
# FLOW-GUARD CACHE (kan-760): a ported guard's shim builds flow-guard from
# its own checkout into a cache (scripts/lib/flow-guard.sh). The runner
# points that cache at its own temp directory through FLOW_GUARD_CACHE_DIR,
# so no harness writes into the operator's real cache; harnesses also run
# with FLOW_GUARD_TELEMETRY=off. The temp directory is removed on exit,
# SIGINT and SIGTERM included.
#
# JOBS: the harnesses run at most JOBS at a time, the core count when unset;
# a JOBS set but not a positive integer is refused, exit 2.
#
# RUN_GUARD_TESTS_ROOT (design.md's runner-root-override): an explicit,
# opt-in override honoured only when set — copied verbatim from
# check-references.sh's own CHECK_REFERENCES_ROOT idiom rather than writing
# a third variant of it. This is load-bearing, not decoration: the
# runner's own Go test points it at fixtures, so no case re-enters the real
# suite.
#
# GUARD_TESTS_REPO_ROOT (kan-584's watched-repo override): the git work tree
# whose cleanliness gates the suite — the repository that owns this runner
# when unset. Same opt-in-override idiom as RUN_GUARD_TESTS_ROOT, and
# load-bearing for the same reason: the runner's Go test points the gate at
# a fixture repository, never at this repository's own tree.
# Set but empty is refused, exit 2, exactly like its sibling. A watched
# root that is not a git work tree cannot answer the question, so the gate
# is skipped with one stderr line — never a silent pass.
#
# THE TREE GATE (kan-584): before the harnesses launch, the runner snapshots
# `git status --porcelain` over the watched root; after the summary it
# re-reads it. A NET change fails the runner with exit 1, naming the paths
# present after the suite and absent before — a harness case that mutates a
# real shared file is a suite-contract violation, the same code a failing
# harness gets. The gate compares STATUS LINES, not content, so its green
# is narrower than "nothing touched the tree": edits made and restored
# inside the run net to zero, and mutating an already-modified or
# already-untracked file in place nets the same porcelain line it had
# before — neither is caught. The gate's contract is structural change to
# the tree the suite leaves behind, not a filesystem watch or a content
# hash.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "run-guard-tests: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec run-guard-tests 2 "run-guard-tests:" "$@"
