# kan-312-myflow-per-task-review-and-the-panel-duplicate — design

## Context

KAN-302 (4 tasks, two Markdown files, 114 changed lines) ran 9 reviewer dispatches: 4 per-task
combined reviewers, 2 per-task re-reviews, 3 whole-branch panel slots. Every finding came from the
per-task layer; all three panel slots returned clean. KAN-287, the counter-case the panel exists
for, was 24 tasks / 17 commits touching Go store and harvester code, and its six Major defects
lived at commit seams in that code. The signal separating the two is diff composition: KAN-302's
branch touched only `.md` paths; KAN-287's did not.

Today `skills/flow/review-panel.md` dispatches the settings-store roster on every pass 1, and its
own text names "a documentation-, prompt-, or test-only diff" as a case that "runs the resolved
list alone" — the exact case this change narrows. The store's roster decision
(`roster-from-settings`, archived `panel-roster-follows-the-settings-list`) forbids any automatic
*addition*; nothing in it addresses an automatic reduction, and `skills/flow/SKILL.md`'s
empty-list row already defines a `primary`-alone roster shape.

No staged research note existed (`docs/superpowers/research/kan-312.md`, `kan-312-*.md` absent).

## 1. `scripts/check-panel-docs-only.sh <worktree> <merge-base>`

A shipped guard, symlinked as `skills/flow/scripts/check-panel-docs-only.sh`. Collects the change's
own paths the way `check-panel-citation-trigger.sh` does — committed since the merge base, staged,
and unstaged, unioned — with the same `GIT_BIN="$(type -P git)"` idiom and the same usage checks.

Exit codes:

- **0** — every collected path ends `.md` or `.mdc` (the citation trigger's own `\.mdc?$`
  pattern: one definition of "documentation" in this repository). Prints nothing.
- **1** — at least one path does not. Prints that first non-documentation path to stdout, so the
  panel record can quote why the roster ran in full. An empty path set also exits 1: a branch with
  nothing to read is not "docs-only".
- **2** — cannot answer: a missing argument, `<worktree>` not a directory or not a git worktree,
  `<merge-base>` not resolving, or a git invocation failing. Reason on stderr.

`scripts/test-check-panel-docs-only.sh` is its harness, in `test-check-panel-citation-trigger.sh`'s
shape: throwaway repositories, cases for committed / staged / unstaged non-doc paths, a pure `.md`
and `.mdc` branch, an empty branch, and each exit-2 shape; one case is mutation-proved by swapping
the pattern for one that never matches and confirming the case flips. `run-guard-tests.sh`
discovers it by glob.

## 2. The panel dispatches `primary` alone on a docs-only branch

In `skills/flow/review-panel.md`, immediately after `check-panel-diff-size.sh`:

```bash
check-panel-docs-only.sh <worktree> <merge-base>
```

- **Exit 0:** pass 1 dispatches **`primary` alone** plus any slot the operator's existing per-run
  instruction names (checked at stage start, per **The roster**). Every other resolved slot is
  recorded `not dispatched — docs-only reduction`. Where the resolved roster does not carry
  `primary`, `primary` is still the reduced roster — the same shape `skills/flow/SKILL.md`'s
  empty-list row already defines.
- **Exit 1:** the resolved roster, unchanged.
- **Exit 2:** report stderr and dispatch the full resolved roster. An unanswered question never
  reduces a panel.

The verdict, the printed non-doc path on exit 1, and the roster actually dispatched are recorded in
`<abs-worktree>/.superpowers/sdd/final-review-panel.md` on every run, beside the diff-size fields.

**Fix rounds re-classify.** The guard runs again wherever the re-run cap check runs. If a fix made
the branch no longer docs-only, every resolved slot not yet dispatched this run is dispatched in
that round reading the whole `final-review.diff` — the existing rule for a slot with no held
last-reviewed sha — and the record says so. A branch that stays docs-only keeps the reduced roster;
`primary` re-runs on its delta exactly as **Panel re-runs** already states.

Nothing below the roster changes: findings, reproducers, fix rounds, mutation-proof,
`check-panel-findings-closed.sh`, and the zero-open-findings handoff bar are untouched.

Two existing sentences are replaced, not paraphrased: the clause "a documentation-, prompt-, or
test-only diff with no operator addition runs the resolved list alone, and that is a correct
outcome" under **The roster**, and "Pass 1 always runs the resolved roster plus every slot the
operator named" under **Panel re-runs**.

## 3. `skills/flow/SKILL.md` and `skills/flow/verify-and-handoff.md`

- `SKILL.md`'s "Never add a slot beyond the resolved roster automatically" bullet gains its mirror:
  the one automatic *reduction* is `check-panel-docs-only.sh`'s, and it only ever removes slots.
  The guard-presence list gains `check-panel-docs-only.sh`.
- `verify-and-handoff.md`'s `Panel:` handoff line gains a field: `reduced: docs-only — <slots not
  dispatched> | no`, and its prose describing the line names the reduction alongside substitution
  and addition.

## 4. Records and scope

- `design.md` decision `docs-only-reduces-to-primary`, narrowing `roster-from-settings`: an
  automatic reduction on one deterministic signal, never an addition.
- No `spectre/specs/` capability exists in this repository, so no spec delta.
- Out of scope: KAN-287 (the diff cap scaling with task count) — a separate signal on the same
  guard site, unchanged here. KAN-277's delta re-runs are already shipped and are what section 2's
  fix-round rule builds on.

## Testing

- `scripts/test-check-panel-docs-only.sh` green; `scripts/run-guard-tests.sh` green.
- `.flow/project.md`'s `## lint` list green — `check-guard-symlinks.sh` for the new symlink,
  `check-references.sh` for every new cite, `check-contract-budget.sh` for the three edited
  contracts, `check-normative-inventory.sh` diffed before and after the contract edits with the
  only differences being the sentences section 2 replaces.
