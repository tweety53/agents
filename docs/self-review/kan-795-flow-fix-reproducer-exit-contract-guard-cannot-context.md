# Self-review context bundle for kan-795-flow-fix-reproducer-exit-contract-guard-cannot

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-795-flow-fix-reproducer-exit-contract-guard-cannot/tasks.md (absent)
skipped: spectre/changes/archive/kan-795-flow-fix-reproducer-exit-contract-guard-cannot/design.md (absent)
skipped: spectre/changes/archive/kan-795-flow-fix-reproducer-exit-contract-guard-cannot/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-795-flow-fix-reproducer-exit-contract-guard-cannot.md

# SDD ledger — kan-795-flow-fix-reproducer-exit-contract-guard-cannot

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-10-02T21:45:04Z
- Tokens: not measured

## Dispatch 2 — panel-fix

- Task: no task
- Role: panel-fix
- Key: panel-fix-1
- Model: glm-5.3-flash effort=high
- Commit: 6b1a40d1
- Outcome: completed
- Started: 2026-10-02T22:06:21Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: 62c42495
- Outcome: completed
- Started: 2026-10-02T22:18:09Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-795-flow-fix-reproducer-exit-contract-guard-cannot-panel.md

# Review panel — kan-795-flow-fix-reproducer-exit-contract-guard-cannot

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | important | stats/internal/guard/panelexitcontract.go:177 | Test gap: the prefix index's canonical-tree arm is uncovered — deleting the arm leaves the entire suite green (proven), so a refactor that drops or misorders it lands silently and the first canonical-prefixed citation bounces as a malformed declaration at exit 1. |   |
| F2 | primary | minor | stats/internal/guard/panelexitcontract.go:459-460 | The ambiguous-prefix branch emits one fact as two strings, one of which can never be printed — cannot always sets cannot-answer and exit-2 precedence prints only the cannot message, so the duplicated sentence is unreachable output that will drift. |   |
| F3 | principles | minor | stats/internal/guard/panelexitcontract.go:167-185 | DRY — the recorded-worktree to physical-tree resolution policy (skip vanished, EvalSymlinks, alias dedupe) now lives twice, and the F4 dedupe rule already had to be carried consciously from one site to the other. |   |
| F4 | principles | minor | stats/internal/guard/panelexitcontract.go:459-460 | DRY/KISS — the prefix-ambiguity fact is encoded in two strings, one of which can never be emitted; the second formatting exists only to give the audit a non-empty violation to skip on. |   |
| F5 | primary | minor | .superpowers/sdd/reproducers/0-primary-2.sh:2-4 | F2's reproducer pins panelexitcontract.go:459/460/373 with pre-fix content; the fix rewrote those lines (now 441/348/350), so a guard-mediated re-run bounces at the instrument audit instead of running the script — the refused-re-run re-authoring route that re-authored F1's reproducer should have covered it too; bounded: F2 is fixed and retired, provenance not behavior. |   |

findings-total: 5
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed

reproducers-total: 5
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-principles-2.sh
finding-reproducer: F5 none — the defect is in the reproducer file's own declaration lines, which no runnable command demonstrates; the re-run slot's direct run (exit 0, defect absent) carries the verdict

## Pass log

### Round 0

- roster: compact — 15
- diff size: 179 lines measured, under cap — proceed automatic
- docs-only: no — first non-doc path scripts/check-panel-reproducer-exit-contract.sh; resolved roster dispatched
- no addition this round — the resolved list ran alone
- planning commit: none — flow-fast planning tree is untracked (.superpowers git-excluded), nothing to commit
- wall clock: dispatch returned complete at 936.8s against the 15-minute ceiling — blocking call, the overrun paid in full before it could be recorded; this harness tracks no in-flight elapsed time

### Round 1

- fix: inline by the parent (decision fixer skipped — inline), diff 62c42495..6b1a40d1 at .superpowers/sdd/fix-round-1.diff; no auto-decisions — all four findings went to the fix; no fixup fold, so check-task-commit-fields has no task to re-check
- re-run cap: 252 lines from merge base (no held sha — the round-boundary rebase cleared it), under cap; docs-only: no — Go sources; primary re-runs alone (raised the round's only Important), delta re-runs not applicable (reads fix-round-1.diff)
- re-run wall clock: dispatch returned complete at 423.8s against the 5-minute re-run ceiling — blocking call, the overrun paid in full before it could be recorded
- round-1 close: no Critical or Important raised by the re-run — F5 fixed by the parent inline at the round close (re-pinned the reproducer citations); no commit — .superpowers is git-excluded, nothing tracked changed
fix-mutation: stats/internal/guard/panelexitcontract.go — canonical-tree arm removed from the prefix index (append(..., worktree) -> resolvedRecordedTrees(env, recorded)) — TestCheckPanelReproducerExitContract/case_42 fails under the flip; reproducer 0-primary-1.sh pre-fix demonstrated (exit 9) -> post-fix not demonstrated (exit 0)
fix-mutation: stats/internal/guard/panelexitcontract.go — alias dedupe in the basename index removed — TestCheckPanelReproducerExitContract/case_43 fails under the flip (canonical basename reads ambiguous)
fix-mutation: stats/internal/guard/panelexitcontract.go — ambiguous-prefix dead violation wording replaced by a never-printed sentinel — pre: 1 dead format string in source, 0 printed; post: 1 sentinel, 0 printed — reproducers 0-primary-2.sh and 0-principles-2.sh flipped demonstrated (exit 9) -> not demonstrated (runner exit 1); closes F2 and F4 (one defect identity)
fix-mutation: stats/internal/guard/panelexitcontract.go — recorded-tree resolution extracted to resolvedRecordedTrees — resolution idiom sites 2->1 (pre: EvalSymlinks(pcAbs(env, rw)) at lines 182+314; post: helper only) — reproducer 0-principles-1.sh flipped demonstrated (exit 9) -> not demonstrated (runner exit 1)
fix-mutations-total: 4
## git log --stat

commit 6b1a40d17cd1ac645969455261dae4a818e5f9f9
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Oct 3 01:12:18 2026 +0300

    fix(guard): cover the prefix index's canonical arm and single-source the recorded-tree resolution

 .../check_panel_reproducer_exit_contract_test.go   | 36 +++++++++++
 stats/internal/guard/panelexitcontract.go          | 75 ++++++++++++----------
 2 files changed, 77 insertions(+), 34 deletions(-)

commit 62c424959afd9edd484536739105bc2239ae84fb
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Oct 3 00:40:52 2026 +0300

    fix(guard): resolve a prefixed reproducer citation against the worktree its basename names

 scripts/check-panel-reproducer-exit-contract.sh    | 21 +++++-
 .../check_panel_reproducer_exit_contract_test.go   | 87 ++++++++++++++++++++++
 stats/internal/guard/panelexitcontract.go          | 71 ++++++++++++++++--
 3 files changed, 170 insertions(+), 9 deletions(-)

## Session narrative

A `/flow-fast` run resolved KAN-795 to change `kan-795-flow-fix-reproducer-exit-contract-guard-cannot`, created the worktree off `origin/main`, and rolled the one-task plan to class `small` — execution inline, a compact panel whose single bundled dispatch carried primary+principles. The fix itself was small: the reproducer exit-contract guard's instrument audit had resolved every `# demonstrates:`/`# premise:` citation against the finding's own tree only, so the cross-repo prefixed form the panel records already use (`gymie-frontend:src/Foo.tsx:42`) bounced as a malformed declaration; the change indexes the change's trees by resolved basename (canonical plus the state record's `worktrees` map) and redirects the whole citation resolution to the tree the prefix names, tests first (five new cases), with the bash contract header updated in step. The panel came back with one Important (the new index's canonical-tree arm shipped with zero coverage — a one-line deletion survived the whole suite, proven by mutation) and three Minors (a dead violation string that exit-2 precedence can never print, raised twice; the recorded-tree resolution policy now living twice). The fix round ran inline per the decision — coverage cases with their own proven flips, the sentinel, and one shared `resolvedRecordedTrees` helper — and the primary re-run confirmed both its findings fixed, raising one new Minor (the fixed finding's reproducer still pinned pre-fix line numbers, so a guard-mediated re-run would bounce at the audit instead of running; re-pinned inline). Where the run struggled: the first mutation flip was restored with `git checkout --` before the fix was committed, which silently discarded the uncommitted refactor and made one flip "pass" against unmutated code — caught by re-inspecting the file, fixed by committing before any flip and asserting each mutation landed before trusting its run. The round-boundary base movement also rebased the branch, so the push needed `--force-with-lease`. Both wall-clock ceilings (15-minute panel, 5-minute re-run) were overrun by complete dispatches the harness runs as one blocking call; recorded, not re-dispatched.
