#!/usr/bin/env bash
# check-model-statements.sh — fail when a model-statement line goes missing
# from a site this repository requires it at.
#
# WHY THIS EXISTS. KAN-381's change states the resolved model at every
# model-bearing dispatch site — the canonical rule in skills/flow/SKILL.md's
# Model resolution, plus one sentence at each dispatch site's file. Two
# mutation probes recorded against that change's review panel deleted the
# canonical paragraph and a site sentence and ran every lint guard: all
# green, including the normative inventory, whose wording deliberately
# avoids MUST/SHALL. This guard is the teeth: each row is a file and one
# load-bearing literal that must be present in it. A rewording that keeps
# the literal passes; a deletion fails loud.
#
# WHAT A GREEN RUN DOES NOT PROVE: only that each literal is present —
# never that a dispatcher actually printed the statement, and never that
# the printed value was the resolved one. The line's substance is the
# canonical rule in Model resolution; this guard keeps the sites from
# silently losing it, exactly as check-dispatch-paragraphs.sh keeps its
# required prompt paragraphs from silently disappearing.
#
# Usage: check-model-statements.sh
# Exit 0 every literal found at its site; 1 one or more missing; 2 it
# cannot answer at all (a scoped file missing, unreadable, or not a regular
# file). CHECK_MODEL_STATEMENTS_ROOT overrides the scan root, mirroring
# CHECK_DISPATCH_PARAGRAPHS_ROOT, so test-check-model-statements.sh can
# point it at a sandboxed fixture tree.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ -n "${CHECK_MODEL_STATEMENTS_ROOT:-}" ]; then
  ROOT="$CHECK_MODEL_STATEMENTS_ROOT"
elif [ "${CHECK_MODEL_STATEMENTS_ROOT+set}" = "set" ]; then
  echo "check-model-statements: CHECK_MODEL_STATEMENTS_ROOT is set but empty" >&2
  exit 2
else
  ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
fi

if [ ! -d "$ROOT" ] || [ ! -r "$ROOT" ]; then
  echo "check-model-statements: $ROOT is not a readable directory — cannot scan" >&2
  exit 2
fi

VIOLATIONS=0
require() {
  local rel="$1" literal="$2"
  local file="$ROOT/$rel"
  if [ ! -f "$file" ] || [ ! -r "$file" ]; then
    printf 'check-model-statements: %s: not a readable regular file\n' "$rel" >&2
    exit 2
  fi
  if ! grep -qF -- "$literal" "$file"; then
    printf '%s: required model-statement literal not found: %s\n' "$rel" "$literal"
    VIOLATIONS=1
  fi
}

require skills/flow/SKILL.md 'Every model-bearing dispatch states its resolved value'
require skills/flow/archive.md 'self-review model: <value> (<tier>)'
require skills/flow/brainstorm.md 'planner model: <PLANNING_MODEL> (<tier>)'
require skills/flow/implement.md 'conductor model: <value> (<tier>)'
require skills/flow/implement.md 'planner model: <PLANNING_MODEL> (<tier>)'
require skills/flow/review-panel.md '<slot> model: <value> (<tier>)'
require skills/flow/review-panel.md 'panel-fix model: <DEFAULT_MODEL> (<tier>)'
require skills/flow/verify-and-handoff.md 'verify model: sonnet (fixed literal)'
require skills/flow-research/SKILL.md 'research model: <value> (<tier>)'

if [ "$VIOLATIONS" -ne 0 ]; then
  echo "check-model-statements: required model-statement literal(s) missing" >&2
  exit 1
fi
echo "check-model-statements: all model statements present"
