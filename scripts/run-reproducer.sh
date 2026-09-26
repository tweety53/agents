#!/usr/bin/env bash
# run-reproducer.sh <worktree> <reproducer-command-line> [--pre-fix-verdict <demonstrated|not-demonstrated>] [--reproducer-sha <sha>]
#
# Runs one finding's reproducer — the command text that follows
# `finding-reproducer: F<n> ` in a panel record, already validated
# lexically by check-panel-reproducers.sh — and reports whether it
# demonstrates the defect it names.
#
# WHY THIS SCRIPT EXISTS. Every rule about *running* a reproducer used to be
# prose in skills/flow/review-panel.md, with no code behind it: the
# containment of each token, the ban on shell metacharacters, the direct
# exec, the bound, and the timed-out and surviving-process dispositions.
# Eight review findings across three panel passes were defects in that
# prose, because a rule nothing implements can only be re-read rather than
# tested — most plainly F44, whose fix for an earlier finding cited
# `ps -o sid=`, a keyword this platform's `ps` does not have:
#
#   $ ps -o sid= -p $$
#   ps: sid: keyword not found
#
# This script turns those rules into something with its own tests
# (stats/internal/guard/runreproducer_test.go), so the constraints are
# enforced where they were only described.
#
# THE COMMAND IS DISPATCHED AS AN ARGUMENT VECTOR, NEVER A STRING THROUGH A
# SHELL. The path token is resolved to a real file on disk and executed
# directly, with its arguments passed as an argument vector. The exec itself
# starts the reproducer in a session and process group of its own (D4,
# SysProcAttr.Setsid in the Go port — Darwin ships no `setsid` binary, and
# the bash body needed a python3 shim for it), and the argv vector reaches
# the kernel untouched: the resolved path and its arguments exactly as
# given, so nothing in the reproducer's own text is ever handed to a shell
# for interpretation. That is the guarantee the prose this script replaces
# claimed and had nothing behind; the shell-metacharacter checks in
# stats/internal/guard/runreproducer.go still run first, but they are now
# the second layer of a real argv-exec barrier rather than the only one.
#
# CONTAINMENT IS RESOLVED, NOT MERELY LEXICAL. check-panel-reproducers.sh
# checks a panel record's reproducer lines lexically only — no leading `-`
# on the path token, no absolute token, no `..` path segment, no shell
# metacharacter, no URL — because that guard's record may be read in a
# context where the worktree that wrote it does not exist to resolve
# against. This script always has a real worktree, so it repeats those
# lexical checks (a record can be edited after the guard last ran) and then
# goes further: every token is resolved to its physical path — following
# `..`, `.` and symlinks — and required to stay inside the worktree. A
# relative path with no `..` segment can still point outside the worktree
# through a symlink, which is exactly the shape the record-format guard's
# own header says it deliberately cannot decide; this script is where that
# shape is decided.
#
# THE MUTATION-REPRODUCER CONVENTION (KAN-568). A surviving-mutant
# reproducer — one that lands a mutation and runs the test suite — carries
# the opposite answer in its own exit code from every other reproducer:
# BUILD SUCCESSFUL with the mutation landed IS the bug present, the tests
# having failed to catch it. Such a reproducer declares itself with the
# exact line `# mutation-reproducer` within its first 10 lines, and this
# script then reads it under that convention — exit 0 is "defect
# demonstrated", any non-zero exit "not demonstrated" — instead of the
# generic mapping in the exit-code list below. Everything else about this
# script is identical under both conventions: containment, the bound, the
# kill sequence and the exit-code vocabulary itself never move; only the
# reading of the reproducer's own exit status flips, and because the
# ambiguity refusal (KAN-524) compares verdicts rather than raw exit codes,
# it works unchanged under either.
#
# THE COMPARISON IS VALID ONLY BETWEEN TWO RUNS OF THE SAME FILE. The
# verdict a `--pre-fix-verdict` re-run compares against is meaningful only when
# the file is byte-for-byte what the dispatch-time run read — a re-authored
# reproducer, its mutation-convention declaration included, is a different
# reproducer answering in a vocabulary the pre-fix verdict never carried.
# The runner therefore prints `reproducer sha <hex>` with every verdict, and
# `--reproducer-sha <sha>` pins the re-run to the dispatch-time file: a
# mismatch is refused (exit 2, the never-executed shape class) and the
# reproducer is re-authored instead — re-run it against the defect-present
# code and carry its fresh verdict and sha (KAN-568 review, F1).
#
# Exit codes:
#   0  defect demonstrated — the command ran to completion inside the
#      bound, as a direct exec, and exited non-zero (exit 0 under the
#      mutation-reproducer convention)
#   1  defect not demonstrated — the command ran to completion inside the
#      bound and exited 0 (any non-zero exit under the mutation-reproducer
#      convention); the instruction built on this reproducer cannot
#      be verified as a fix
#   2  refused — one of two classes. The shape class failed a lexical or
#      resolved containment/shape check — or a --reproducer-sha pin — before
#      execution and was never run at all. The ambiguity class (KAN-524) ran
#      to a verdict that is
#      IDENTICAL to the pre-fix verdict the caller passed in
#      --pre-fix-verdict: a reproducer that answers the same way before and
#      after the fix demonstrates nothing under either exit-code
#      convention, and the expected convention is named on stderr. Only
#      the shape class never executes; the ambiguity class is refused at
#      its verdict, after a real run.
#   3  unverifiable — the command was still running at the bound (or a
#      detached child of it survived the kill sequence) and was killed;
#      its exit status is read as neither a pass nor a fail. A surviving
#      child is additionally named on stderr with its pid.
#   4  cannot answer at all — bad usage, no such worktree, an
#      environment failure (no writable temporary directory, and similar), or
#      a plumbing failure in the exec subshell itself (the worktree vanished
#      before it could be entered, or exec could not run the resolved path), or
#      a process-table read that failed, which may have missed a detached
#      child — never the reproducer's own exit status

# Header corrected for the Go port (KAN-760): the tests, the setsid mechanism
# and the metacharacter checks are named as they now are. The body's
# reasoning for every branch is in stats/internal/guard/runreproducer.go.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 4 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
. "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "run-reproducer: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 4
}
flow_guard_exec run-reproducer 4 "run-reproducer:" "$@"
