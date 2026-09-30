#!/usr/bin/env bash
# render-slot-prompt.sh — render one review-panel dispatch's brief
# (skills/flow/review-panel.md, **Bundled dispatch**).
#
# Usage: render-slot-prompt.sh <skill-dir> <round> <canonical-wt> <plan-dir> <slot>[+<slot>...] \
#          -diff final|late-fix|delta|fix-round [-no-bundle] [-standard <path>...] \
#          [-fix-report <path>...] -- <worktree>...
#
# <skill-dir> is the directory review-panel.md was read from; <worktree>...
# is the change's resolved set, the canonical worktree first. Every path is
# absolute: the subagent resolves what it is given from its own directory.
#
# Writes <canonical-wt>/.superpowers/sdd/slot-prompt-<round>-<slot>[+<slot>...].md,
# atomically, and prints its path. In order:
#   1. the shared blocks, each extracted verbatim by its `> **<LABEL>:**`
#      blockquote from <skill-dir>/review-panel.md (the first one — the
#      slot's): TOOLS, FOREGROUND BUILDS, NO DELEGATION, REPRODUCE, DON'T
#      READ and WORKTREES (the resolved set filled in); CITATION CHECK once
#      per worktree whose .superpowers/sdd/citation-check.md exists; ENTRY
#      CONTEXT followed by the diff's .touched list (write-panel-diff.sh) —
#      each pass's own on a delta bundle; INDEPENDENT PASSES when more than
#      one slot; FIX-ROUND SCOPE from review-panel-fix-round.md, naming the
#      -fix-report path(s), only when -fix-report is given;
#   2. one `## PASS <id>` section per slot, in the order given: the fenced
#      body of <skill-dir>/<id>-reviewer-prompt.md, dedented, every
#      placeholder filled per review-panel.md's placeholder table (**The
#      roster**), then that slot's REPORT FILE block.
#
# The diff each pass reads, in <canonical-wt>/.superpowers/sdd/:
#   final      final-review.diff
#   late-fix   late-fix.diff
#   fix-round  fix-round-<round>.diff
#   delta      slot-delta-<round>-<id>.diff
# -no-bundle is the CONTEXT BUNDLE FAILURE continue path: [CONTEXT_BUNDLE_PATHS]
# becomes `none — the bundle was not built`, [GLOBAL_CONSTRAINTS] the plan's
# design.md (or `none`). -standard paths are [STANDARDS_PATHS]; empty when
# none are given.
#
# `mutation`'s pass is never rendered: its brief and its throwaway copy stay
# typed by the parent. A bundle naming it renders every other role, the
# shared INDEPENDENT PASSES paragraph and a file name carrying the whole
# bundle; mutation alone is refused. MODEL HANDSHAKE, CONTEXT BUNDLE, the relocation-comparison
# pointer and the reproducer rule stay typed in the Agent call too.
#
# Exit codes:
#   0  written; its path on stdout
#   1  a placeholder is left unfilled — each named on stderr; nothing written
#   2  cannot answer — usage, a relative path, mutation alone or an unknown slot, a
#      missing template, block, touched list, standards, principles or
#      calibration file
#
# flow-guard is built from this checkout, never taken from PATH:
# scripts/lib/flow-guard.sh derives it, and exits 2 (this guard's
# cannot-answer code) with the cause when it cannot.
set -euo pipefail
# $SCRIPT_DIR/ spells each sibling this shim needs where check-guard-symlinks rule 2 reads it.
SCRIPT_DIR="$(cd "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
. "$SCRIPT_DIR/lib/flow-guard.sh" || {
  echo "render-slot-prompt: cannot load lib/flow-guard.sh beside ${BASH_SOURCE[0]}" >&2
  exit 2
}
flow_guard_exec render-slot-prompt 2 "render-slot-prompt:" "$@"
