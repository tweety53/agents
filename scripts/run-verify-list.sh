#!/usr/bin/env bash
# run-verify-list.sh — run one of a project's verify lists the way **Inline
# verify** (skills/flow/verify-and-handoff.md) prescribes, and print its
# `## Report` lines.
#
# Usage: run-verify-list.sh <worktree> lint|test [<change> <key> <session-token> <model> <effort> [<canonical-worktree>]]
#
# Exit codes:
#   0  every command exited 0
#   1  at least one command exited non-zero
#   2  usage (not `<worktree> lint|test` plus none, all five, or all six recording
#      arguments, no flow binary on PATH for those, or <worktree> is not a
#      directory), a list line holding a `<word>` placeholder — its stderr
#      line names it — or project-get.sh exit 2 — its stderr line says why
#   3  the list is not declared (project-get.sh exit 1): the caller
#      auto-detects, as **Inline verify** says
#   4  the store refused the recording begin (stderr names why): no command ran
#
# With the recording arguments the list ends the row it began with one of
# three outcomes: blocked with cause test-failure when a command failed,
# skipped when every command of a test list was skipped, and completed
# otherwise — a lint list always runs check-spec-reach.sh, so it never ends
# skipped.
#
# A `# suite: <name>` line in the `test` list names the one command below it
# as that suite: it runs under `flow suite record -C <worktree> -suite <name>
# --`, so its runtime reaches the store `flow suite list` reads; with no flow
# on PATH it runs bare.
#
# A `test` list ends with the `## visual verification` `verify` row, run once
# after every other command, `<specs>` empty, when the worktree's diff from its
# base branch touches a `ui paths` glob and decision.json does not record
# `visual.verify` `skipped`; a base that does not resolve runs it, and it fails
# the list like any other command (kan-1047).
#
# The mechanism, and its full contract — the list, the green sha, the output
# — is stats/internal/guard/runverifylist.go; its test is
# stats/internal/guard/run_verify_list_test.go.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "run-verify-list: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
: "$SCRIPT_DIR/project-get.sh" "$SCRIPT_DIR/check-spec-reach.sh"
FLOW_GUARD_SCRIPT_DIR="$SCRIPT_DIR"
export FLOW_GUARD_SCRIPT_DIR
flow_guard_exec run-verify-list 2 "run-verify-list:" "$@"
