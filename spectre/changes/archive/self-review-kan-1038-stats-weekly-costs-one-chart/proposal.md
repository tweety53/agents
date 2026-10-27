# self-review-kan-1038-stats-weekly-costs-one-chart

## Why

The self-review pass of `/flow` change `kan-1038-stats-weekly-costs-one-chart`
(`.superpowers/sdd/self-review-pass.md` in this worktree) found thirteen non-`big` defects, the
operator's instruction included:

- F1 (1.1): `pipeline-fix.sh verify` never runs the `internal/guard` tests that audit every script,
  so a touched `scripts/*.sh` that breaks `TestNoPipeIntoEarlyExitGrep` lands main red.
- F2 (1.2): the docs say a dispatch launches first and its `begin` is recorded after, while the
  begin-row hook refuses a launch with no open row.
- F3 (1.3): `flow stage`, `flow state`, `flow settings` and `flow jira` follow the
  workspace-isolated `FLOW_ADDR`, so their writes fall to the journal in a worktree.
- F4 (1.4): the visual-verify landing gate counts a verifier row the parent recorded inline.
- F5 (1.5): one vitest test timed out and two screenshots failed under verify's concurrent load.
- F6 (4.1): `run-reproducer` notices a reproducer that leaves the tree modified only after a kill.
- F7 (4.2): the `capture` and `verify` rows run against a possibly stale UI-test stack.
- F8 (5.1): verify rows recorded by `run-verify-list.sh` are measured as one turn or not at all.
- F9 (6.1): the guard package's tests run twice in every verify.
- F10 (8.1): a fix-round re-run is never asked for defects the fix diff itself introduces.
- F11 (8.2): an approved HTML prototype never becomes a frame `flow.visual-verify` measures against.
- F12 (8.3): a visual-fix task's **Files:** omit the `visual-verification.md` its fix makes stale.
- F13 (operator): integrate and the self-review still ask scope and cost questions that carry a
  recommended option.

## What changes

- `pipeline-fix.sh verify` runs the script audits of `internal/guard` for a touched script.
- Every `begin` is recorded before its launch, in every passage that orders the two.
- The four verbs resolve the records address, as the record family does.
- An `-agent-id inline` verifier row no longer satisfies the visual-verify gate.
- The two load flakes are reproduced and pinned at their source.
- `run-reproducer` exits 5 when a run leaves the tree different from how it found it.
- The `capture` and `verify` rows restart a stale UI-test stack before they run.
- The harvester's settle pass re-measures verify rows begun by `run-verify-list.sh`; this change's
  rows are restamped.
- The guard package runs once per verify, under `-race`.
- The FIX-ROUND SCOPE paragraph asks for defects the fix introduces, and the guard pins it.
- `flow.visual-verify` renders an approved prototype as each capture's frame; the report guard
  refuses step 10 `n/a` while one is cited.
- The visual-fix task lists the visual record in **Files:**.
- No scope or cost question with a recommended option is asked during integrate or the self-review.

## Impact

`stats/internal/guard`, `stats/internal/harvest`, `stats/internal/repotest`, `stats/cmd/flow`,
`stats/web`, `scripts/`, `.flow/project.md`, `skills/flow/`, `skills/flow-contracts/`,
`skills/flow-self-review/SKILL.md`.
