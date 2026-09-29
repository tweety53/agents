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
PC=("$C/worktree-resolution.md" "$C/git-boundaries.md" "$C/handoff-blocks.md" "$C/operator-prompts.md" "$C/operator-prompts-auto-resolution.md")
row "  phase files + load directives" "${PD[@]}"
row "  + cited at point of use" "${PC[@]}"
row "  TOTAL definite" "${ALWAYS[@]}" "${ROUTER[@]}" "${PD[@]}"
row "  TOTAL incl. cited" "${ALWAYS[@]}" "${ROUTER[@]}" "${PD[@]}" "${PC[@]}"

echo "--- implementation session (enters via Resuming at STARTED) ---"
ID=("$F/brainstorm.md" "$F/implement.md" "$C/artifacts-registry.md" "$C/worktree-resolution.md" "$F/review-panel.md" "$F/verify-and-handoff.md" "$C/session-records.md" "$C/git-boundaries.md")
IC=("$C/operator-prompts.md" "$C/operator-prompts-auto-resolution.md" "$C/project-configuration-standards.md" "$C/model-policy.md" "$C/known-bugs.md" "$F/primary-reviewer-prompt.md" "$F/principles-reviewer-prompt.md" "$F/failure-modes-reviewer-prompt.md")
IX=("$F/review-panel-optional-slots.md" "$C/workspace-isolation.md" "$F/visual-verify.md" "$C/project-configuration-visual.md" "$C/guard-verdict-verification.md")
row "  phase files + load directives" "${ID[@]}"
row "  + cited / reviewer templates" "${IC[@]}"
row "  + conditional (optional slots, isolation, UI)" "${IX[@]}"
row "  TOTAL definite" "${ALWAYS[@]}" "${ROUTER[@]}" "${ID[@]}"
row "  TOTAL incl. cited" "${ALWAYS[@]}" "${ROUTER[@]}" "${ID[@]}" "${IC[@]}"
row "  TOTAL worst case" "${ALWAYS[@]}" "${ROUTER[@]}" "${ID[@]}" "${IC[@]}" "${IX[@]}"

echo "--- finish session (merge-and-push: run 1 chained into run 2) ---"
FD=("$F/integrate.md" "$C/worktree-resolution.md" "$C/finish-contract-run1.md" "$C/git-boundaries.md" "$C/session-records.md" "$C/jira-integration.md" "$F/archive.md" "$C/artifacts-registry.md" "$C/finish-contract-run2.md")
FC=("$C/operator-prompts.md" "$C/model-policy.md")
FX=("$C/jira-followups.md" "$C/state-file.md" "$C/project-configuration.md" "$C/guard-verdict-verification.md")
row "  phase files + load directives" "${FD[@]}"
row "  + cited" "${FC[@]}"
row "  + conditional (follow-ups, state file, config, verdicts)" "${FX[@]}"
row "  TOTAL definite" "${ALWAYS[@]}" "${ROUTER[@]}" "${FD[@]}"
row "  TOTAL incl. cited" "${ALWAYS[@]}" "${ROUTER[@]}" "${FD[@]}" "${FC[@]}"

echo "--- subagents (each also inherits whatever always-on context the harness gives it) ---"
row "  reviewer: baseline + primary template" "$R/rules/agent-baseline.md" "$F/primary-reviewer-prompt.md"
row "  reviewer: baseline + principles + EP" "$R/rules/agent-baseline.md" "$F/principles-reviewer-prompt.md" "$F/engineering-principles.md"
row "  reviewer: baseline + failure modes" "$R/rules/agent-baseline.md" "$F/failure-modes-reviewer-prompt.md"
row "  project-configuration.md (cited everywhere)" "$C/project-configuration.md"
row "  state-file.md (load before any state r/w)" "$C/state-file.md"
exit 0
