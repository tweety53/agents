# Self-review context bundle for kan-589-flow-automation-a-tested-helper-runs-an

found: 2 of 6 sources; skipped: 4 of 6 sources
skipped: spectre/changes/archive/kan-589-flow-automation-a-tested-helper-runs-an/tasks.md (absent)
skipped: spectre/changes/archive/kan-589-flow-automation-a-tested-helper-runs-an/design.md (absent)
skipped: spectre/changes/archive/kan-589-flow-automation-a-tested-helper-runs-an/narrative.md (absent)
skipped: git log --stat (absent)

## .superpowers/sdd/ledgers/kan-589-flow-automation-a-tested-helper-runs-an.md

# SDD ledger — kan-589-flow-automation-a-tested-helper-runs-an

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-18T21:24:11Z
- Tokens: not measured

## Dispatch 2 — panel-fix

- Task: no task
- Role: panel-fix
- Key: panel-fix-1
- Model: glm-5.3-flash effort=high
- Commit: 47ae414147fa759c741a4aef912651afbc010f6f
- Outcome: completed
- Started: 2026-09-18T21:50:53Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: haiku effort=low
- Commit: no commit
- Outcome: completed
- Started: 2026-09-18T21:51:48Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-589-flow-automation-a-tested-helper-runs-an-panel.md

# Review panel — kan-589-flow-automation-a-tested-helper-runs-an

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | important | scripts/installer-sandbox-diff.sh:90 | the symlink-target normalization expands $tree as a glob pattern, so a tree path containing a glob metacharacter silently defeats normalization and two identical trees are reported as differing |   |
| F2 | primary | minor | scripts/test-installer-sandbox-diff.sh:1 | the harness never tests the retention half of the exit-code contract — exit 0 removes the sandbox, exit 1/2 keep it and print its path — nor the bad-argument-count exit 2 |   |
| F3 | primary+principles | minor | scripts/installer-sandbox-diff.sh:35-37 | the header misdescribes the MANAGED_BLOCK_POSTPROCESS hazard — the zcode call is prefix-scoped inside the installer, and the real risk of a caller-exported value is masking a genuine content difference, the opposite direction |   |
| F4 | principles | minor | scripts/installer-sandbox-diff.sh:44-46 | an untrapped environmental failure after startup — mktemp unable to create the sandbox — dies under set -e with status 1, the documented differs code, instead of the contract exit 2 |   |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 none — coverage finding; its one-line grep probe carries a quote and is refused by the reproducer shape guard
finding-reproducer: F3 .superpowers/sdd/reproducers/0-principles-1.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-principles-2.sh

## Pass log

### Round 0

- no addition this round — the resolved list ran alone
- roster: compact — compact_roll 12 < 90
- diff size 256 under cap
- docs-only rc=1 (scripts/installer-sandbox-diff.sh); resolved roster runs
- experimental roll 18 < 30 but compact roster has no second dispatch — skipped: bundle cap

### Round 1

- panel-fix ran inline in the parent (execution inline) — F1 F3 F4 fixed, F2 fixed with the mutation-proof harness cases; diff .superpowers/sdd/fix-round-1.diff
- F3 post-fix reproducer refused ambiguous (exited 1 pre and post): the reproducer is behavior-shaped and the finding is documentation-accuracy — comment-only fix cannot flip it; closed on the path check, the named lines 35-40 rewritten in 47ae414, deviation recorded here and named in the summary
- primary re-ran targeted on F1 (rerun pair haiku/low recorded, glm-5.3-flash/high at dispatch per Harness mapping): FIXED, reproducer exit 0, harness 16/16, no new defects at the site; principles did not re-run — its findings were Minors fixed inline under the trivial bar
fix-mutation: scripts/installer-sandbox-diff.sh — unquoted the tree pattern in the symlink normalization () — scripts/test-installer-sandbox-diff.sh case 6 — harness exits 1, FAIL: case 6
fix-mutation: scripts/installer-sandbox-diff.sh — dropped the guarded mktemp assignment (|| die), leaving the untrapped set -e exit — scripts/test-installer-sandbox-diff.sh case 7 — harness exits 1, FAIL: case 7 expected exit 2 got 1
fix-mutations-total: 2

## Branch log

commit 47ae414147fa759c741a4aef912651afbc010f6f
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Sep 19 00:50:36 2026 +0300

    fix(scripts): close the review panel findings on the sandbox diff

 scripts/installer-sandbox-diff.sh      | 22 +++++++++++-----
 scripts/test-installer-sandbox-diff.sh | 48 ++++++++++++++++++++++++++++++++--
 2 files changed, 62 insertions(+), 8 deletions(-)

commit 580e5a1798991e5a80ed40298278c293d1ec7b68
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Sep 19 00:21:55 2026 +0300

    test(scripts): cover the installer old-vs-new sandbox diff

 scripts/test-installer-sandbox-diff.sh | 129 +++++++++++++++++++++++++++++++++
 1 file changed, 129 insertions(+)

commit fa17ea07f426b09dfcb4198754f9c696baa09e4a
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Sep 19 00:21:52 2026 +0300

    feat(scripts): add the installer old-vs-new sandbox diff

 scripts/installer-sandbox-diff.sh | 127 ++++++++++++++++++++++++++++++++++++++
 1 file changed, 127 insertions(+)

## Session narrative

This run added the installer old-vs-new sandbox diff the KAN-589 ask named: scripts/installer-sandbox-diff.sh runs the real `setup.sh global` from two trees under two fresh mktemp HOME sandboxes and diffs a normalized manifest (entry type, mode, path, symlink targets with the tree root factored out to <TREE>, regular-file checksums — which is where the managed blocks live), and scripts/test-installer-sandbox-diff.sh pins its contract with nine fixture-installer cases. It struggled in small ways: the first mutation-proof call chained its restore behind a grep -c whose zero count broke the chain and left the tool file mutated for the remainder of one Bash call (caught by re-checking state before proceeding; the restore was verified and the mutation re-run cleanly in its own call); the harness's run_diff forwarded $1/$2 positionally and crashed under set -u on the one-argument case it exists to test; and the build-green guard wanted **Build:** tags the flow-fast plan shape had not asked for, added without argument. The review panel (compact roster, bundled primary+principles on glm-5.3-flash/high) raised one Important — the tree path was expanded as a glob pattern in the symlink normalization, so a bracketed directory name silently produced a false differs verdict — plus three Minors, all fixed in one fix commit (47ae414) whose both executable behaviours were mutation-proved (unquoting and unguarding were each caught by the new cases) and whose F1 fix the primary slot re-verified on the rerun dispatch. One deviation recorded openly: F3's behavior-shaped reproducer refused as ambiguous on the post-fix re-run (a comment-accuracy fix cannot flip an exit code), so it closed on the path check with the refusal recorded in the pass log rather than being put to an operator mid-run.
