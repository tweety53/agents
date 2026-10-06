#!/usr/bin/env bash
# check-ask-silence.sh — fail when a run-loaded ask site states neither its
# own silence outcome nor a citation of the pipeline contract's **Unanswered
# mid-run asks** section.
#
# Why this exists. KAN-880, deferred from KAN-772's self-review: the
# corpus-wide survey of AskUserQuestion ask sites' silence outcomes was
# hand-run once and nothing re-ran it, so a future ask site added without its
# own silence outcome and without citing **Unanswered mid-run asks**
# (`skills/flow-contracts/pipeline.md`) drifted silently. This guard is that
# survey, re-run on every lint pass.
#
# THE UNIT is the Markdown section — a heading to the next heading — of an
# owned-corpus file that mentions AskUserQuestion outside a code fence. A
# section passes when it cites the contract, states a silence outcome in the
# contract's own vocabulary (silence, silent, unanswered, default,
# recommended, explicit, without asking), or delegates the ask with one of
# the prepositions `per **` / `applies **` / `under **` / `exactly as **`.
# The detection is lexical, never semantic; the full contract, its limits
# and the corpus definition live in the Go port's doc comment,
# stats/internal/guard/asksilence.go, whose table tests
# (stats/internal/guard/check_ask_silence_test.go) stand in the
# test-check-*.sh companion's place.
#
# THE CORPUS mirrors scripts/lib/owned-corpus.sh — the scope roots skills/,
# rules/, spectre/specs/, commands-claude/ and .flow/ plus the .md/.mdc
# files directly at the root; a symlinked scope root is refused, and a
# nested symlinked directory is refused when it would hide Markdown. Change
# the two definitions together.
#
# Verdicts:
#
#   ASK-SILENCE-OK:   <root> — N ask section(s), each citing or stating its silence outcome
#   ASK-SILENCE-OK:   <root> — no AskUserQuestion site in the corpus
#   ASK-SILENCE-FAIL: <root> — N ask section(s) of M state no silence outcome; state the outcome, delegate the ask, or cite Unanswered mid-run asks
#
# Every violation names `file:line` and the section heading; the remedy is
# to state the outcome, delegate the ask, or cite the section — never to
# widen the vocabulary to hide one.
#
# Exit codes: 0 every ask section passes (and when the corpus holds none),
# 1 violation(s) found, 2 the corpus cannot be answered at all.
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
