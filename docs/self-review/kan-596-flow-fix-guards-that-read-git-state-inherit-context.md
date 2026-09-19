# Self-review context bundle for kan-596-flow-fix-guards-that-read-git-state-inherit

found: 2 of 6 sources; skipped: 4 of 6 sources
skipped: spectre/changes/archive/kan-596-flow-fix-guards-that-read-git-state-inherit/tasks.md (absent)
skipped: spectre/changes/archive/kan-596-flow-fix-guards-that-read-git-state-inherit/design.md (absent)
skipped: spectre/changes/archive/kan-596-flow-fix-guards-that-read-git-state-inherit/narrative.md (absent)
skipped: git log --stat (absent)

## .superpowers/sdd/ledgers/kan-596-flow-fix-guards-that-read-git-state-inherit.md

# SDD ledger — kan-596-flow-fix-guards-that-read-git-state-inherit

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-19T20:27:26Z
- Tokens: not measured

## Dispatch 2 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-19T21:12:08Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: principles
- Key: panel-1-principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-19T21:12:08Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-596-flow-fix-guards-that-read-git-state-inherit-panel.md

# Review panel — kan-596-flow-fix-guards-that-read-git-state-inherit

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary+principles | critical | scripts/check-base-moved.sh:133 | adding --no-renames shifted the guard's git args, so test-check-base-moved.sh's positional failure-injection shims never fire — canonical suite red, 8 cases fail; causality proven by merge-base archive run exiting 0 (principles P1: the guarded call shape changed and its coverage was silently defused instead of updated) |   |
| F2 | primary+principles | critical | scripts/test-git-config-pins.sh:131 | R1 auditor matches only the literal token git, so "$GIT_BIN"-invoked diffs are never audited — live unpinned verdict-bearing reads exist while case 5 reports the corpus pinned (principles P2: the auditor returns no-hits for invocation forms it cannot parse, indistinguishable from compliant) |   |
| F3 | primary | critical | scripts/lib/panel-touched-paths.sh:89 | four verdict-bearing diff reads remain unpinned in panel-touched-paths.sh:89,93,97 and plan-class.sh:150 — with diff.renames=true panel_touched_paths elides a rename's source path and unpinned numstat counts 0 vs 10 changed lines, flipping guard verdicts by machine config |   |
| P3 | principles | minor | scripts/test-git-config-pins.sh:66 | p.name != test-git-config-pins.sh is dead — not p.name.startswith(test-) already excludes the auditor |   |
| F4 | primary | minor | scripts/test-git-config-pins.sh:106 | R2 accepts any --whitespace=<value>; the plan's rule demands --whitespace=nowarn, so a future --whitespace=error audits clean |   |
| P4 | principles | minor | scripts/test-git-config-pins.sh:25 | header still states R2 as any explicit --whitespace= while the auditor now demands literal nowarn, and omits the new $GIT_BIN coverage it documents for g() |   |
| F5 | primary | minor | scripts/test-git-config-pins.sh:120 | line-oriented scanner misses backslash-continued git invocations (no live instance today) |   |
| F6 | primary | minor | .superpowers/sdd/kan-596-flow-fix-guards-that-read-git-state-inherit/tasks.md:13 | task 1 Baseline before=0 after=4 vs the delivered harness's 5 cases |   |

findings-total: 8
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: P3 fixed
finding-status: F4 fixed
finding-status: P4 fixed
finding-status: F5 deferred continuation-joining is new scanner logic needing its own fixture; no live instance today
finding-status: F6 fixed

reproducers-total: 8
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: P3 none — a dead boolean provable only by reading; exempt as Minor
finding-reproducer: F4 .superpowers/sdd/reproducers/0-primary-4.sh
finding-reproducer: P4 none — stale header prose, zero corpus impact
finding-reproducer: F5 .superpowers/sdd/reproducers/0-primary-5.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-primary-6.sh

## Pass log

### Round 0

- roster: compact — 73
- diff-size: 349 lines, under cap — proceeding automatically
- docs-only: exit 1 — runs the resolved roster; first non-documentation path scripts/break-and-prove.sh
- no addition this round — the resolved list ran alone

### Round 1

- inline fix (parent itself, execution inline) — no panel-fix dispatch; diff .superpowers/sdd/fix-round-1.diff; F1 shims re-aimed, F2 auditor GIT_BIN forms, F3 four reads pinned, F4 R2 tightened, F6 baseline corrected, P3 dead boolean; F3 reproducer re-authored per the refused-re-run rule (fresh sha 5a2c1bdb), all five runnable findings re-run clean; F5 deferred coverage-gap
- delta re-run on the rerun pair: primary clean (F1-F4,F6 fixed, F5 deferral stands), principles fixed with one new Minor P4 (stale auditor header) — fixed inline as 3d7fedd; round raised only Minors, no further re-run
fix-mutation: scripts/test-check-base-moved.sh — 9c/9f shims re-aimed at the pinned 7-arg/5-arg diff shape — 8→0 failing cases in test-check-base-moved.sh; mutating 9c back to $6 fails 4 cases again
fix-mutation: scripts/test-git-config-pins.sh — R1 auditor extended to $GIT_BIN command forms; R2 tightened to --whitespace=nowarn — 0→4 R1 hits named with a GIT_BIN pin reverted; --whitespace=error fixture 0→1 R2 hits
fix-mutation: scripts/lib/panel-touched-paths.sh — --no-renames pinned on 4 verdict-bearing reads (3 touched-path diffs, 1 numstat) — reverting the COMMITTED pin: audit 0→1 R1 hits naming the site; restored: exit 0
fix-mutation: scripts/test-git-config-pins.sh — dead auditor-name boolean removed (P3) — none — audit output byte-identical pre/post, which is the definition of dead
fix-mutation: .superpowers/sdd/kan-596-flow-fix-guards-that-read-git-state-inherit/tasks.md — none — Baseline count corrected in the planning file (F6) — none — no executable behaviour
fix-mutations-total: 5

## Session narrative

The run pinned the guard corpus's git reads against machine config (rename detection, whitespace action, untracked visibility) and added `scripts/test-git-config-pins.sh` as the audit that keeps them pinned, TDD-style: the audit harness landed first red with 27 live hits, then four tasks turned it green across two shell-guard commits, one Python-guard commit, and one mutation-harness commit. Where it struggled: the brainstorm inventory itself missed the `"$GIT_BIN"` indirection form and the `lib/panel-touched-paths.sh` reads, and the pass-1 panel caught both (F2, F3, both Critical) plus a real regression of the run's own making — the `--no-renames` pins shifted argument positions so `test-check-base-moved.sh`'s positional failure-injection shims silently stopped firing (F1), which the fix round repaired by re-aiming the shims and, on the same class, a citation-trigger xtrace anchor the panel had not named. The round's one reproducer that could never flip (it compared a raw unpinned call against a pinned reference — a divergence that exists by design forever) was re-authored per the refused-re-run rule before the finding closed. Origin/main moved twice mid-run; the entry rebase was clean and the landing rebase is expected to repeat that.
