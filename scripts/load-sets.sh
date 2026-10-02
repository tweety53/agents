#!/usr/bin/env bash
# load-sets.sh [<repo>] [<rendered-global-CLAUDE.md>] — the static instruction bytes each /flow
# session loads from this repository (KAN-852). A report, never a verdict: it always exits 0.
#
# Sums the bytes of each session's load set as the flow files' load directives define it —
# planning, implementation and finish, each its own session (/clear between them) — plus the
# files cited at the point of use and the conditional loads. Tokens are approximated as
# bytes/4. It excludes the superpowers skills, MEMORY.md and the project's own code.
#
# <repo> defaults to this script's repository. The second argument defaults to
# ~/.claude/CLAUDE.md, the always-on block setup.sh installs; render a sandboxed one with
#   SB="$(mktemp -d)"; HOME="$SB" ./setup.sh global   # then pass "$SB/.claude/CLAUDE.md"
#
# The arrays below are the load sets, by hand: edit them when a change moves what a session
# loads. A file that does not exist is printed as MISSING and counted as zero bytes, so a stale
# array shows up in the report rather than aborting it.
#
# Ported from the prototype loadsets.sh in docs/prompt-audit-2026-09-29/tools.md.
set -uo pipefail
R="${1:-$(cd "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)}"
GLOBAL="${2:-$HOME/.claude/CLAUDE.md}"
F="$R/skills/flow"; C="$R/skills/flow-contracts"

bytes() { # total bytes of the files named; a missing one is reported on stderr and counts 0
  local t=0 f
  for f in "$@"; do
    if [ -f "$f" ]; then t=$((t + $(wc -c < "$f"))); else echo "load-sets: MISSING ${f#"$R"/}" >&2; fi
  done
  echo "$t"
}
row() { local name="$1" b; shift; b=$(bytes "$@"); printf '%-50s %7d bytes  ~%6d tok\n' "$name" "$b" $((b / 4)); }

ALWAYS=("$GLOBAL" "$R/CLAUDE.md")
ROUTER=("$R/commands-claude/flow.md" "$F/SKILL.md" "$C/pipeline.md")
row "always-on (global block + project CLAUDE.md)" "${ALWAYS[@]}"
row "router (command + SKILL.md + pipeline.md)" "${ROUTER[@]}"

echo "--- planning session (creating run) ---"
PD=("$F/brainstorm.md" "$F/brainstorm-planner.md" "$C/jira-integration.md" "$C/plan-provenance.md" "$C/build-green.md")
PC=("$C/worktree-resolution.md" "$C/git-boundaries.md" "$C/operator-prompts.md" "$C/operator-prompts-auto-resolution.md")
PX=("$F/withdrawal.md" "$F/seeded-note.md" "$F/resume.md")
row "  phase files + load directives" "${PD[@]}"
row "  + cited at point of use" "${PC[@]}"
row "  TOTAL definite" "${ALWAYS[@]}" "${ROUTER[@]}" "${PD[@]}"
row "  + conditional (flow-fix/flow-cost/flow-speed, seeded note, resume)" "${PX[@]}"
row "  TOTAL incl. cited" "${ALWAYS[@]}" "${ROUTER[@]}" "${PD[@]}" "${PC[@]}"

echo "--- implementation session (enters via Resuming at STARTED) ---"
ID=("$F/resume.md" "$F/implement.md" "$C/worktree-resolution.md" "$F/review-panel.md" "$F/verify-and-handoff.md" "$C/session-records.md" "$C/git-boundaries.md")
IC=("$C/operator-prompts.md" "$C/operator-prompts-auto-resolution.md" "$C/project-configuration-standards.md" "$C/model-policy.md" "$C/known-bugs.md" "$F/primary-reviewer-prompt.md" "$F/principles-reviewer-prompt.md" "$F/failure-modes-reviewer-prompt.md" "$C/plan-amendment.md")
IX=("$F/document-fix.md" "$F/cross-repo-worktrees.md" "$F/sdd-dispatch.md" "$F/gated-review-fix.md" "$F/review-panel-late-fix.md" "$F/review-panel-fix-round.md" "$F/review-panel-optional-slots.md" "$F/review-panel-experimental-slot.md" "$C/workspace-isolation.md" "$F/visual-verify.md" "$F/visual-verify-tooling-analysis.md" "$C/git-boundaries-commit-chain.md" "$C/project-configuration-visual.md" "$C/guard-verdict-verification.md")
row "  phase files + load directives" "${ID[@]}"
row "  + cited / reviewer templates" "${IC[@]}"
row "  + conditional (fix run, cross-repo, sdd, gated fix, fix round, optional slots, isolation, UI)" "${IX[@]}"
row "  TOTAL definite" "${ALWAYS[@]}" "${ROUTER[@]}" "${ID[@]}"
row "  TOTAL incl. cited" "${ALWAYS[@]}" "${ROUTER[@]}" "${ID[@]}" "${IC[@]}"
row "  TOTAL worst case" "${ALWAYS[@]}" "${ROUTER[@]}" "${ID[@]}" "${IC[@]}" "${IX[@]}"

echo "--- finish session (merge-and-push: run 1 chained into run 2) ---"
FD=("$F/integrate.md" "$C/worktree-resolution.md" "$C/finish-contract-run1.md" "$C/git-boundaries.md" "$C/git-boundaries-commit-chain.md" "$C/session-records.md" "$C/jira-integration.md" "$C/jira-integration-finish.md" "$F/archive.md" "$C/artifacts-registry.md" "$C/finish-contract-run2.md")
FC=("$C/operator-prompts.md" "$C/model-policy.md")
FX=("$F/unfinished-work-gate.md" "$F/sync-onto-base.md" "$C/finish-hand-fallbacks.md" "$C/jira-followups.md" "$C/jira-followups-join.md" "$C/state-file.md" "$C/project-configuration.md" "$C/guard-verdict-verification.md")
row "  phase files + load directives" "${FD[@]}"
row "  + cited" "${FC[@]}"
row "  + conditional (gate, base moved, hand fallbacks, follow-ups, state file, config, verdicts)" "${FX[@]}"
row "  TOTAL definite" "${ALWAYS[@]}" "${ROUTER[@]}" "${FD[@]}"
row "  TOTAL incl. cited" "${ALWAYS[@]}" "${ROUTER[@]}" "${FD[@]}" "${FC[@]}"

echo "--- subagents (each also inherits whatever always-on context the harness gives it) ---"
row "  reviewer: baseline + primary template" "$R/rules/agent-baseline.md" "$F/primary-reviewer-prompt.md"
row "  reviewer: baseline + principles + EP" "$R/rules/agent-baseline.md" "$F/principles-reviewer-prompt.md" "$F/engineering-principles.md"
row "  reviewer: baseline + failure modes" "$R/rules/agent-baseline.md" "$F/failure-modes-reviewer-prompt.md"
row "  verifier: baseline + verifier steps (UI runs)" "$R/rules/agent-baseline.md" "$F/visual-verify-verifier.md"
row "  project-configuration.md (cited everywhere)" "$C/project-configuration.md"
row "  state-file.md (load before any state r/w)" "$C/state-file.md"
exit 0
