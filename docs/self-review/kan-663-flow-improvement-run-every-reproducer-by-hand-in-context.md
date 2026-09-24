# Self-review context bundle for kan-663-flow-improvement-run-every-reproducer-by-hand-in

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-663-flow-improvement-run-every-reproducer-by-hand-in/tasks.md (absent)
skipped: spectre/changes/archive/kan-663-flow-improvement-run-every-reproducer-by-hand-in/design.md (absent)
skipped: spectre/changes/archive/kan-663-flow-improvement-run-every-reproducer-by-hand-in/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-663-flow-improvement-run-every-reproducer-by-hand-in.md

# SDD ledger — kan-663-flow-improvement-run-every-reproducer-by-hand-in

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: not recorded
- Started: 2026-09-24T22:14:35Z
- Tokens: not measured

## Dispatch 2 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-24T22:52:38Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: principles
- Key: panel-1-principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-24T22:52:38Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-663-flow-improvement-run-every-reproducer-by-hand-in-panel.md

# Review panel — kan-663-flow-improvement-run-every-reproducer-by-hand-in

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary+principles | Critical | skills/flow/review-panel.md:414 | the prose invokes prove-reproducer.sh but the tracked symlink skills/flow/scripts/prove-reproducer.sh was never added, so check-guard-symlinks.sh exits 1 and run-guard-tests.sh fails its own-tree case — the plan Build-green is unmet |   |
| F2 | primary+principles | Important | scripts/prove-reproducer.sh:112 | mkdir -p and cp -p run unguarded under set -e, so a plumbing failure exits 1 (the proof-did-not-hold verdict code, silently, no leg run) instead of the documented cannot-answer 2 |   |
| F3 | primary | Minor | scripts/test-prove-reproducer.sh | no case drives the leg-refused mapping (runner 2/3/4 to prove 2); cases 4/5/7 all stop at the door |   |
| F4 | principles | Minor | scripts/prove-reproducer.sh:87 | DRY: third verbatim carrier of the worktree-relative path-shape rule; judged trade noted, extraction precedent is reproducer-metachars.sh |   |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F3 none — a missing test case is not itself executable
finding-reproducer: F4 .superpowers/sdd/reproducers/0-principles-2.sh

## Pass log

### Round 0

- roster: compact — 87
- no addition this round — the resolved list ran alone
- diff size: 449 lines, under cap — proceed
- docs-only: exit 1 — scripts/prove-reproducer.sh; resolved roster unchanged

### Round 1

- parent fixed inline — F1 symlink (a535fb6), F3 runner exec-failure mapping (2eee855), F2+F4 guarded copy + shared refusals (5cdf10a); three mutation flips red→green; three pinned reproducer re-runs read not-demonstrated
- delta re-runs: primary and principles, one dispatch per role — all findings verified fixed, no new defects; clean round
fix-mutation: skills/flow/scripts/prove-reproducer.sh — tracked symlink added — 1→0 check-guard-symlinks violations
fix-mutation: scripts/prove-reproducer.sh — mkdir guard removed — F2 reproducer: demonstrated→not-demonstrated
fix-mutation: scripts/prove-reproducer.sh — exit-4 mapping opened — case_8 red→green
fix-mutation: scripts/run-reproducer.sh — exec-failure sentinel write dropped — runner case 34 red→green
fix-mutation: scripts/lib/reproducer-path.sh — path-shape rule extracted — 3→2 verbatim carriers (F4 reproducer flip)
fix-mutations-total: 5
## git log --stat

commit 5cdf10a01fc3331be375a4e99fc914e3ffc01470
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 25 01:50:26 2026 +0300

    fix(scripts): guard the scratch copy and single-source the recorded-form refusals

 scripts/lib/reproducer-path.sh   | 37 +++++++++++++++++++++++++++++++++++++
 scripts/prove-reproducer.sh      | 33 +++++++++++++++++++--------------
 scripts/test-prove-reproducer.sh | 33 +++++++++++++++++++++++++++++++++
 3 files changed, 89 insertions(+), 14 deletions(-)

commit 2eee855beb868e18246577ec219aebe6141c0070
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 25 01:50:23 2026 +0300

    fix(scripts): read a shim exec failure as cannot-answer, never a defect verdict

 scripts/run-reproducer.sh      | 22 +++++++++++++++++++++-
 scripts/test-run-reproducer.sh | 17 +++++++++++++++++
 2 files changed, 38 insertions(+), 1 deletion(-)

commit a535fb6e0bbec1628d08013a6a951d310125a41e
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 25 01:50:21 2026 +0300

    fix(flow): track the skills-side symlink for the reproducer prover

 skills/flow/scripts/prove-reproducer.sh | 1 +
 1 file changed, 1 insertion(+)

commit 97adf488d29089bce9d6c892cee452af9ad9ad0c
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 25 01:12:01 2026 +0300

    docs(flow): require a reproducer proved in both directions before it is recorded

 skills/flow/review-panel.md | 19 +++++++++++++++++--
 1 file changed, 17 insertions(+), 2 deletions(-)

commit a8acefee810004bcee1351aba1ed4ddcd24c244f
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 25 01:09:51 2026 +0300

    feat(scripts): prove a reproducer in both directions from a scratch worktree

 scripts/prove-reproducer.sh      | 151 +++++++++++++++++++++
 scripts/test-prove-reproducer.sh | 279 +++++++++++++++++++++++++++++++++++++++
 2 files changed, 430 insertions(+)

## Session narrative

This `/flow-fast` run implemented KAN-663 in the flow machinery itself: `scripts/prove-reproducer.sh` mechanizes the two-direction reproducer proof (a detached scratch worktree at the caller-named pre-fix commit, the reproducer copied to the same worktree-relative path inside it, one `run-reproducer.sh` leg per side, both exits printed), `scripts/test-prove-reproducer.sh` asserts the contract across eight cases, and `skills/flow/review-panel.md`'s reproducer rule now requires both recorded exits at authoring and at repair, naming the script as the mechanism. The run's own review panel then earned its keep: pass 1 caught a missing tracked skills-side symlink (Critical — `check-guard-symlinks` red on the change's own tree), an unguarded copy step that mis-reported plumbing failures as the proof-failed verdict (Important), and two Minors; the inline fix round fixed all four and, while adding the requested mapping test, discovered a real pre-existing runner defect — `run-reproducer.sh` read a shim exec failure as the reproducer exiting 1, a false *defect demonstrated* on a script that never ran — and fixed it with a second sentinel byte from the shim's except path, mapped to the documented cannot-answer 4. Every fix was mutation-proved red-then-green and closed on pinned reproducer re-runs; both delta re-runs came back clean. Where it struggled: the zsh session shell could not source `scripts/lib/sha256-hex.sh` (`local path` collides with zsh's special `$path`), so the pinned re-runs computed digests with `shasum` directly — a bash-only assumption in a shipped lib worth a look on a future flow-fix; and one stage mark (`flow.sdd-tdd` end) landed late, after the review-panel mark had already superseded it, so the stage set is complete but its timings are not clean.
