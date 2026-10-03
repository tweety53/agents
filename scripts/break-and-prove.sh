#!/usr/bin/env bash
# break-and-prove.sh [--clean <command>] <file> (--sed <expr> | --patch <patchfile>) \
#   -- <test-command> [<args>...]
#
# Mechanizes KAN-329's break-it/prove-it loop for config guards, so a fix
# round's evidence is uniform and quotable instead of prose each agent
# composes by hand: apply the mutation to <file>, force a clean re-run of
# <test-command>, assert it exits non-zero, restore <file>, force another
# clean re-run, assert it exits zero, and print both runs' observed output.
# A fix round that skips the proof is visible because the script produced
# no output.
#
# THE RESTORE IS A PRE-MUTATION SNAPSHOT, NEVER `git checkout --`. One of
# the observed failures this script exists to prevent is an agent whose
# `git checkout --` restore also reverted the uncommitted edits it had
# made on the same file. The file's bytes are copied to a temp snapshot
# before anything mutates, and the restore writes those bytes back — so
# uncommitted edits on <file> itself survive the loop, and the script
# deliberately does NOT refuse a dirty <file> the way mutate-and-verify.sh
# does. No git checkout call appears in this script, structurally.
#
# THE CLEAN RE-RUN IS THE SCRIPT'S JOB, not the caller's memory. Gradle
# silently skips a re-run when only a compose or YAML file changed — those
# are not declared task inputs — and an agent that forgot `:app:cleanTest`
# recorded a stale green as evidence. `--clean <command>` runs before EACH
# test invocation, from the repository root, through `sh -c`; the command
# is the operator's own (it is printed verbatim before it runs) and is the
# one place this script evaluates a string rather than an argv vector.
# Omit it for runners with no skip-to-green trap.
#
# THE TEST COMMAND IS AN ARGV VECTOR, NEVER A STRING THROUGH A SHELL, and
# it runs from the repository root with its arguments passed verbatim —
# give it paths relative to that root, the same way run-reproducer.sh
# passes its reproducer argv through untouched.
#
# WHERE THE TARGET MAY LIVE IS A PROPERTY OF THE MUTATION KIND, and the
# asymmetry is deliberate: a `--sed` target may sit outside the repository
# (the snapshot restore needs no git), while a `--patch` target must not
# (git apply is repository-bound).
#
# Exit codes (the same philosophy as mutate-and-verify.sh and
# run-reproducer.sh: the code reports this script's own mechanics and the
# proof's legs, never a richer verdict):
#   0  proof held — the mutation was applied, run 1 exited non-zero, the
#      file was restored byte-exact, and run 2 exited zero. Both runs'
#      output is printed above the verdict in labeled blocks.
#   1  the proof did not hold — run 1 exited 0 (the mutation did not break
#      the test: a guard gap or a stale-cache skip; its output is printed),
#      or run 2 exited non-zero after a byte-exact restore (the test fails
#      independently of the mutation). Which leg failed is named in the
#      report.
#   2  refused before mutating anything — <file> is missing or not
#      readable, the patch does not apply cleanly, the patch touches files
#      other than <file>, a --patch target lies outside the repository, or
#      the sed expression fails or changes nothing — OR post-restore drift:
#      the tree no longer matches its pre-mutation snapshot (a new stash
#      entry or an unexpected status line, each named in the report, per
#      stats/internal/guard/postmutationcheck.go).
#   3  could not restore — <file> is not byte-identical to its pre-mutation
#      snapshot after the restore; the file may still be mutated. Takes
#      precedence over exit 2 when drift also fired.
#   4  cannot answer — bad usage, not inside a git worktree, an unreadable
#      patch file, the test command exited 126 or 127 (an exec failure or
#      the command's own such exit — either way it is not usable as
#      evidence), or the --clean command itself failed.
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 4 (this script's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "break-and-prove: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 4
}
flow_guard_exec break-and-prove 4 "break-and-prove:" "$@"
