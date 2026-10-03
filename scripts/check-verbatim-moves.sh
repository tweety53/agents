#!/usr/bin/env bash
# check-verbatim-moves.sh [<base-ref>] — prove a prompt trim moved or
# de-duplicated text and never reworded it (KAN-852).
#
# Compares the sentences of the run-loaded corpus — skills/**/*.md,
# commands-claude/*.md, rules/*.md and rules/*.mdc, symlinks skipped — at
# <base-ref> against the working tree. With no argument the base is the merge
# base of HEAD with the default branch, resolved offline (origin/HEAD, else
# origin/main, else main); on the default branch itself that is HEAD, so only
# uncommitted edits are judged.
#
#   FAIL   a sentence left the run-loaded files and is neither still stated in
#          one (a de-duplication) nor landed verbatim in a -rationale.md —
#          deleted or reworded.
#   FAIL   a new run-loaded sentence is neither a load directive ("**Load `…",
#          "Load `…` only when") nor a citation ("(`skills/….md`)") — a
#          paraphrase.
#   REVIEW a sentence moved verbatim into a -rationale.md carries an imperative
#          marker (never, always, must, only, before, unless, …): a rule a run
#          can no longer see. Reported, never failing — a human confirms it.
#
# A change that rewords or adds a rule on purpose lists each such sentence,
# exactly as the FAIL line prints it after "::", one per line, in its
# change's acknowledgement file — <spec-root>/changes/<change>/verbatim-moves.txt
# on a /flow run, <worktree>/.superpowers/sdd/<change>/verbatim-moves.txt on a
# /flow-fast run, whose guardrail forbids writing <project>/spectre/ — where
# `#` lines are comments and a leading `\` is dropped, so a heading is listed
# as `\## …`.
# Only in-flight changes count; an archived change's list is never read.
#
# It cannot judge whether a lazily loaded file's "Load X only when Y"
# condition is right — review each such directive by hand.
#
# Exit codes: 0 clean (REVIEW lines included), 1 at least one FAIL,
# 2 cannot answer (no git worktree, unresolvable base, empty corpus).
#
# Go, in flow-guard (stats/internal/guard/verbatimmoves.go), ported from the
# prototype verbatim_check.py in docs/prompt-audit-2026-09-29/tools.md.
# The binary does not live in this checkout, so this shim exports
# FLOW_GUARD_REPO_ROOT — this script's parent directory — as the root.
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
[ -r "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" ] && . "$(dirname -- "${BASH_SOURCE[0]}")/lib/flow-guard.sh" || {
  echo "check-verbatim-moves: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec check-verbatim-moves 2 "check-verbatim-moves:" "$@"
