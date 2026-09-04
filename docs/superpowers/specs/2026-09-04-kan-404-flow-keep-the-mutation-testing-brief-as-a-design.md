# KAN-404 — keep the mutation-testing brief as a standing panel slot

## Problem

The review panel dispatches Bugbot by `subagent_type: bugbot` carrying a mutation-testing brief.
Where the harness offers no such agent type, a general-purpose subagent runs the same brief in
Bugbot's place. KAN-29's panel ran that substitute and it produced the panel's highest-severity
findings. Nothing lets an operator choose that reviewer deliberately: it exists only as a fallback
that fires when Bugbot is absent.

## Design

**A sixth reviewer id, `mutation`.** `ValidReviewers` in `stats/internal/store/settings.go` gains
`"mutation"`. `DefaultReviewers` is unchanged, so no project's panel grows without an operator
choosing it through `/flow-settings`.

**Dispatch.** A general-purpose subagent on `DEFAULT_MODEL` (recorded as that model, never
`unknown (agent-defined)`), carrying the mutation-testing brief, dispatched once per round into its
own throwaway worktree copy per repository, exactly as Bugbot is. `-slot` records `Mutation`.
Re-runs follow the rule Bugbot and Security already follow: it reads no diff file, so it re-runs
only when it raised a finding in the previous round or the previous round raised a new Critical.

**Shared brief and worktree treatment.** The roster sections "Bugbot's mutation-testing brief" and
"Bugbot's throwaway worktree" become the brief and the worktree treatment of both `bugbot` and
`mutation`; the copy is named `<worktree>-<slot>-<round>` so both slots in one roster produce
distinct copies in one round. `artifacts-registry.md`'s row for the copy is widened to that name.

**The brief grows two shapes**, alongside flip a condition / drop a guard / move a boundary / remove
a branch: move an interaction off its target; overlay an earlier commit's tree and run the tests.

**`bugbot` is unchanged**: own agent type where offered, substituted where not.

## Files

- `stats/internal/store/settings.go`, `stats/internal/store/settings_test.go`
- `stats/cmd/flow/settings.go` — `-reviewers` help string
- `skills/flow/review-panel.md` — roster row, "five entries, never a sixth", the two shared
  sections, the re-run rule, dispatch recording
- `skills/flow/SKILL.md` needs no edit — its by-`subagent_type` parenthetical stays true
- `skills/flow-contracts/artifacts-registry.md` — copy name
