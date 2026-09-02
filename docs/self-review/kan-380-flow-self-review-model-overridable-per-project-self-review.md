# Self-review — kan-380-flow-self-review-model-overridable-per-project

Model: Fable 5.1 (`SELF_REVIEW_MODEL` resolved to `fable` — `.flow/project.md`'s `## self review
model` key, this change's own new project-key tier)

**Rating:** 3/5

## Problems encountered, and what pipeline change would avoid them (`myflow-fix`)

- Near-miss: the archive-time dispatcher read the settings store's raw `selfReviewModel` value
  directly instead of re-resolving `SELF_REVIEW_MODEL` through this change's own new three-tier
  order, nearly dispatching self-review on `opus` instead of the correctly-resolved `fable` —
  caught only because the operator asked "why opus?" mid-run. Filed: KAN-381.
- All 6 reproducers this run's review slots supplied were rejected by `check-panel-reproducers.sh`
  (shell metacharacters, or a nonstandard `none` form), forcing manual verification of every
  finding. Filed: KAN-382.
- A compound heredoc-piped `flow state set` command was blocked by the harness's permission
  classifier mid-run. Filed: KAN-383.

## Token/time cost, and what would reduce it without quality loss (`myflow-cost`)

- 17 of 18 ledger rows read "not measured"/"session never bound", despite KAN-172 and KAN-212
  (both Done) supposedly covering this. Filed: KAN-384 (shared with the stats-app angle below).
- The Full re-run escalation trigger fired on operator-approved new files, costing a second fix
  round and a third panel round. Filed: KAN-385.

## What went well, and how to reproduce it (`myflow-improvement`)

None — no actionable findings this round; the relay-contract planner dispatch, the mutation-testing
brief on the Bugbot-substitute slot (which caught F3/F4 that no static-reading slot found), and the
operator's mid-review scope decision folding cleanly into the existing fix-round mechanism all
worked as designed, with nothing to change.

## What could be automated or moved to a script (`myflow-automation`)

- `skills/flow/archive.md` still describes hand-deriving the workspace id instead of naming the
  existing `flow workspace-id` command (KAN-224, Done). Filed: KAN-386.
- No mechanical resolver exists at any model-bearing dispatch site — the root cause of the near-miss
  above. Folded into KAN-381 (added `myflow-automation` label rather than duplicating).
- `scripts/run-guard-tests.sh` doesn't flag a `check-*.sh` guard with no `test-check-*.sh`
  companion, so KAN-380's own round-1 fix shipped two guards with none until a reviewer caught it.
  Filed: KAN-387.

## What could move to the Go app or its persistent storage (`myflow-stats-app`)

- `scripts/check-model-keys.sh` regexes `ValidModels` out of Go source instead of asking `flowd`'s
  API, which doesn't yet expose the valid set. Filed: KAN-388.
- Token attribution/session binding is still broken for this harness's dispatch path despite two
  closed issues claiming to have fixed it. Folded into KAN-384 (added `myflow-stats-app` label
  rather than duplicating).

## Filed

KAN-381, KAN-382, KAN-383, KAN-384, KAN-385, KAN-386, KAN-387, KAN-388.
