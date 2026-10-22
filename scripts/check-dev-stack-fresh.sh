#!/usr/bin/env bash
# check-dev-stack-fresh.sh — verify, before a handoff prints a running dev
# stack's URL, that the stack is serving the worktree's own build rather
# than one committed before HEAD.
#
# Usage: check-dev-stack-fresh.sh <worktree>
#
# ONE ARGUMENT, THE WORKTREE — the directory holding `.flow/project.md` and
# the build the stack ought to be serving, matching check-visual-verification.sh's
# own project-root convention.
#
# KAN-334's failure, measured on kan-333: a handoff printed `localhost:3000`
# while the dev server behind it was serving a bundle eleven commits stale —
# the watcher had stopped recompiling, the log had stopped growing, and
# nothing in the pipeline looked. The start rule in the run-instructions
# stage rebuilds and restarts what it can, but a stack it only skips or
# merely starts (a start command that does not rebuild, a watcher that has
# died) is trusted rather than checked. This guard is the check.
#
# THE CHECK IS THE PROJECT'S OWN DECLARED FINGERPRINT. A project tells the
# pipeline once, in `.flow/project.md`'s `## visual verification` section,
# how to prove a served bundle is the worktree's build: the `fingerprint`
# command row, which rebuilds first and then compares the served document
# against the rebuild (this repository's own row is the worked example).
# Reusing that row is the whole design: no second per-project mechanism is
# invented, no new project-configuration vocabulary is added, and a project
# that already declared the row for visual verification needs to write
# nothing. The row's own exit code is the verdict.
#
# THREE EXIT CODES, project-get.sh's convention:
#   0  fresh — a `fingerprint` row is declared and its command exited 0.
#   1  stale (or not serving) — the command is declared and exited non-zero.
#      The restart action is the same for both, so the caller words one line.
#   2  cannot answer — usage; `<worktree>` is not a directory; no
#      `.flow/project.md`; no `## visual verification` section; the section
#      carries no readable `fingerprint` row. The caller prints its
#      freshness-unverified gap line from this exit, so the reason goes to
#      stderr and the exit is never a silent pass.
#
# Exit 1 carries the fingerprint command's own output, streamed verbatim —
# it is the evidence, and a diff or a curl failure names itself.
#
# THE DECLARED COMMAND IS EXECUTED — that is this guard's purpose, and the
# one thing that separates it from check-visual-verification.sh, which
# validates the same section and deliberately runs nothing. The command
# comes from the project's own tracked `.flow/project.md`, the same trust
# every `## run` command the pipeline executes already carries.
#
# The row is read by the same table helpers every visual guard uses —
# vtSplitCells, vtTrimCell and vtFoldCell, after stripBOM — and code fences
# are skipped so a worked example inside the section can never declare a
# row. The mechanism is stats/internal/guard/devstackfresh.go; its test is
# stats/internal/guard/check_dev_stack_fresh_test.go.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-dev-stack-fresh: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_cwd_checkout "${BASH_SOURCE[0]}" "$@"
flow_guard_exec check-dev-stack-fresh 2 "check-dev-stack-fresh:" "$@"
