# Self-review context bundle for kan-622-flow-fix-the-exit-contract-guard-s-exit-1

found: 4 of 7 sources; skipped: 3 of 7 sources
skipped: spectre/changes/archive/kan-622-flow-fix-the-exit-contract-guard-s-exit-1/tasks.md (absent)
skipped: spectre/changes/archive/kan-622-flow-fix-the-exit-contract-guard-s-exit-1/design.md (absent)
skipped: spectre/changes/archive/kan-622-flow-fix-the-exit-contract-guard-s-exit-1/narrative.md (absent)

## change summary

# kan-622-flow-fix-the-exit-contract-guard-s-exit-1 — change summary

KAN-622 (panel finding F2 of kan-568's deferred self-review): the exit-contract
guard's not-demonstrated message and its header contract paragraph told authors of
declared mutation reproducers the opposite of the truth — "must exit non-zero here" —
while the runner reads a mutation reproducer inverted (exit 0 = defect present,
the build succeeding with the mutation landed).

## What changed

- `scripts/check-panel-reproducer-exit-contract.sh` — the runner-exit-1 violation
  message (d57d56f) and the header's THE CLAIM paragraph (822ac97) now state the
  contract verdict-based and name both conventions: a generic reproducer
  demonstrates with a non-zero exit, a declared mutation-reproducer with exit 0.
- `scripts/test-check-panel-reproducer-exit-contract.sh` — case 30 pins the
  mutation-convention wording, case 31 pins the generic clause; each proven
  load-bearing by mutation (deleting its clause fails exactly that case).

## How it was verified

- Red-first TDD for case 30 against the merge-base guard; full harness 31/31 green.
- Real guard + real runner exercised in both directions (mutation-declared exit 0
  → demonstrated; non-zero → not demonstrated).
- Fix round: both slots' reproducers re-run on pinned shas, both flipped to
  not-demonstrated; full `## lint` list (25 guard commands, gofmt, go vet, tsc -b)
  exit 0 in the worktree; targeted harness re-run at verify.

## Panel outcome

Pass 1 (primary+principles) raised three findings: F1 Important (header contract
disagreed with the reworded message — fixed, 822ac97), F2 Minor (generic clause
unpinned — fixed, b117c98), F3 Minor deferred (commit scope `harness` vs the
`guards` convention on these files — the subject sits on pushed d57d56f and the
fix-round contract takes fixes as new commits, never a pushed-history rewrite).
F2's slot reproducer was re-authored after the runner's 20s bound could not
verdict its full-suite run (recorded in the pass log); its replacement
demonstrates pre-fix and flips post-fix on a pinned sha. Both round-1 re-runs
came back clean.

## Deliberately left out

- F3 only (deferred, recorded with reason and category). No other scope.
## .superpowers/sdd/ledgers/kan-622-flow-fix-the-exit-contract-guard-s-exit-1.md

# SDD ledger — kan-622-flow-fix-the-exit-contract-guard-s-exit-1

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-21T19:28:59Z
- Tokens: not measured

## Dispatch 2 — panel-fix

- Task: no task
- Role: panel-fix
- Key: panel-fix-1
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-21T19:48:34Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-21T19:52:36Z
- Tokens: not measured

## Dispatch 4 — reviewer

- Task: no task
- Role: reviewer
- Slot: principles
- Key: panel-1-principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-21T19:52:36Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-622-flow-fix-the-exit-contract-guard-s-exit-1-panel.md

# Review panel — kan-622-flow-fix-the-exit-contract-guard-s-exit-1

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary+principles | important | scripts/check-panel-reproducer-exit-contract.sh:20 | The guard's own header contract paragraph still inverts the instruction this change exists to fix — the file now states its contract two ways that disagree. |   |
| F2 | primary+principles | minor | scripts/test-check-panel-reproducer-exit-contract.sh:586 | Case 30 pins only the mutation half of the reworded message; the generic half is deletable with the whole harness green. |   |
| F3 | primary | minor | commit d57d56f | Commit scope 'harness' does not name the module the commit's substance moved. |   |

findings-total: 3
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 deferred the subject sits on pushed commit d57d56f and the fix-round contract takes fixes as new commits, never a pushed-history rewrite

reproducers-total: 3
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh

## Pass log

### Round 0

- roster: compact — 59
- diff size 16 lines, under cap — proceed
- docs-only: no — first non-documentation path scripts/check-panel-reproducer-exit-contract.sh; resolved roster primary+principles runs
- no addition this round — the resolved list ran alone
- F2 reproducer unverifiable — runner timeout at the 20s bound (it ran the full harness); re-authored in place as a bound-fitting instrument for the same defect identity; fresh pre-fix verdict demonstrated, sha b7155fe51628ab7

### Round 1

- inline panel-fix (execution inline, parent applies the fix): F1 header contract reworded verdict-based (822ac97), F2 generic clause pinned by case 31 (b117c98); F3 deferred before the fix; fix diff at .superpowers/sdd/fix-round-1.diff
- reproducer re-runs: F1 and F2 both flipped to not-demonstrated on pinned shas (13d37036, b7155fe5); fix diffs touch each named path — F1 hunk is comment-only because the defect site is the header comment itself
- delta re-runs: primary and principles each re-ran alone on fix-round-1.diff at their own finding sites — both verdicts fixed, no new defects, harness 31/31 green
fix-mutation: scripts/check-panel-reproducer-exit-contract.sh — none — header hunk is comment-only — no executable behaviour changed
fix-mutation: scripts/check-panel-reproducer-exit-contract.sh — deleted the message generic clause after "it declares:" — case 31 measured both sides: suite exit 1 with clause gone, exit 0 restored
fix-mutations-total: 2
## git log --stat

commit b117c981e2392ceb572cf5a9bca3c1aaa1651226
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 22:52:01 2026 +0300

    test(guards): pin the generic convention clause of the exit-contract guard's message

 scripts/test-check-panel-reproducer-exit-contract.sh | 7 +++++++
 1 file changed, 7 insertions(+)

commit 822ac97a4284b2848ca93a64c7b1e5a7149443c6
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 22:52:01 2026 +0300

    fix(guards): state the exit-code contract verdict-based in the exit-contract guard's header

 scripts/check-panel-reproducer-exit-contract.sh | 9 ++++++---
 1 file changed, 6 insertions(+), 3 deletions(-)

commit d57d56f574a03f5bf5912411ec8d9ebc285d4fed
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 22:26:31 2026 +0300

    fix(harness): name both exit-code conventions in the exit-contract guard's not-demonstrated message

 scripts/check-panel-reproducer-exit-contract.sh      |  2 +-
 scripts/test-check-panel-reproducer-exit-contract.sh | 14 ++++++++++++++
 2 files changed, 15 insertions(+), 1 deletion(-)

## Session narrative

A `/flow-fast` run resolved KAN-622 to `kan-622-flow-fix-the-exit-contract-guard-s-exit-1`,
planned one task, rolled a small class with a compact primary+principles panel, and
implemented the message reword red-first (case 30) in commit d57d56f. The panel's pass 1
confirmed the fix and caught the one thing the plan under-scoped — the guard's own header
still stated the generic-only contract — plus the unpinned generic clause and the
off-convention commit scope. The inline fix round reworded the header (822ac97), added the
load-bearing case 31 (b117c98), and both targeted re-runs came back clean. Two moments
struggled: F2's slot reproducer ran the full harness and could not be verdicted inside the
runner's fixed 20-second bound, so it was re-authored in place as a bound-fitting
instrument for the same defect identity (recorded in the pass log with its fresh sha); and
the first mutation probe of case 31 deleted the clause's first occurrence in the file —
the header copy — rather than the message copy, read green wrongly, and had to be retargeted
at the runtime message before the pin was proven load-bearing.
