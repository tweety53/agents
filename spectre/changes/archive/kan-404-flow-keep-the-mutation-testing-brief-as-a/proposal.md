# kan-404-flow-keep-the-mutation-testing-brief-as-a

**Jira:** KAN-404

## Why

KAN-29's self-review, under "what went well", records that the panel's highest-severity findings
(F11, F24–F26) came from the reviewer that ran in Bugbot's place: the harness offered no `bugbot`
agent type, so `skills/flow/review-panel.md`'s **An unspawnable id is substituted, not skipped**
dispatched a general-purpose subagent carrying Bugbot's mutation-testing brief. Sabotage-proofing
— moving a click off its target and watching the test stay green, overlaying an earlier commit's
tree and running the suite — caught what green runs hid.

That reviewer exists only as a fallback. `ValidReviewers` in
`stats/internal/store/settings.go` carries five ids; the mutation brief reaches the panel only
through `bugbot`, and only runs general-purpose when `bugbot` cannot be spawned. An operator who
wants that reviewer on purpose — on a harness that does offer Bugbot, or beside it — has no id to
name in `/flow-settings`.

## What changes

- `stats/internal/store/settings.go` — `ValidReviewers` gains a sixth id, `mutation`.
  `DefaultReviewers` is unchanged: the slot is opt-in through `/flow-settings`.
- `stats/cmd/flow/settings.go` — the `-reviewers` help string names the six ids.
- `skills/flow/review-panel.md` — the roster gains a `mutation` row: a general-purpose subagent on
  `DEFAULT_MODEL`, carrying the mutation-testing brief, dispatched once per round into its own
  throwaway worktree copy per repository, re-run under the rule Bugbot and Security already follow.
  The brief and the throwaway-worktree sequence become sections shared by `bugbot` and `mutation`,
  the copy named `<worktree>-<slot>-<round>`. The brief gains two shapes: move an interaction off
  its target; overlay an earlier commit's tree and run the tests.
- `skills/flow-contracts/artifacts-registry.md` — the throwaway copy's row takes the per-slot name.
  `skills/flow/SKILL.md`'s "Slots dispatched by `subagent_type` (Bugbot, Security)" stays true:
  `mutation` is not one.
- `bugbot` is unchanged: its own agent type where offered, substituted where not.
