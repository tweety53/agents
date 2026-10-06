#!/usr/bin/env bash
# check-ask-silence.sh — fail when a run-loaded ask site states none of its
# own silence outcome, a delegation of the ask, or a citation of the pipeline
# contract's **Unanswered mid-run asks** section.
#
# Why this exists. KAN-880, deferred from KAN-772's self-review: the
# corpus-wide survey of AskUserQuestion ask sites' silence outcomes was
# hand-run once and nothing re-ran it, so a future ask site added without its
# own silence outcome and without citing **Unanswered mid-run asks**
# (`skills/flow-contracts/pipeline.md`) drifted silently. This guard is that
# survey, re-run on every lint pass.
#
# THE CONTRACT lives in one home: the Go port's doc comment,
# stats/internal/guard/asksilence.go — the ask-site unit, the pass reasons
# and their exact vocabulary, the heuristic's limits, the corpus definition
# and the exit codes. This header states only the shape: a Markdown section
# mentioning AskUserQuestion outside a code fence must cite the contract,
# state a silence outcome, or delegate the ask; every violation names
# `file:line` and the section heading, and the remedy is to state the
# outcome, delegate the ask, or cite the section — never to widen the
# vocabulary to hide one. The doc comment's table tests
# (stats/internal/guard/check_ask_silence_test.go) stand in the
# test-check-*.sh companion's place.
#
# Verdicts:
#
#   ASK-SILENCE-OK:   <root> — N ask section(s), each citing, delegating, or stating its silence outcome
#   ASK-SILENCE-OK:   <root> — no AskUserQuestion site in the corpus
#   ASK-SILENCE-FAIL: <root> — N ask section(s) of M state no silence outcome; state the outcome, delegate the ask, or cite Unanswered mid-run asks
#
# The guard runs as the Go port in stats/internal/guard/asksilence.go. The
# binary does not live in this checkout, so this shim exports
# FLOW_GUARD_REPO_ROOT — this script's parent directory — as the checkout
# whose corpus is walked. flow-guard is built from this checkout, never
# taken from PATH: scripts/lib/flow-guard.sh derives it, and exits 2 (this
# guard's cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
[ -r "$SCRIPT_DIR/lib/flow-guard.sh" ] && . "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "check-ask-silence: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
FLOW_GUARD_REPO_ROOT="$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
export FLOW_GUARD_REPO_ROOT
flow_guard_exec check-ask-silence 2 "check-ask-silence:" "$@"
