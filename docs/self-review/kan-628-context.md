# Self-review context bundle for kan-628

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-628/tasks.md (absent)
skipped: spectre/changes/archive/kan-628/design.md (absent)
skipped: spectre/changes/archive/kan-628/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-628.md

# SDD ledger — kan-628

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-22T18:00:10Z
- Tokens: not measured

## Dispatch 2 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: 6773d5277d0ec298bc2e7037924702a4a619daab
- Outcome: completed
- Started: 2026-09-22T18:34:36Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: principles
- Key: panel-1-principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: 6773d5277d0ec298bc2e7037924702a4a619daab
- Outcome: completed
- Started: 2026-09-22T18:34:36Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-628-panel.md

# Review panel — kan-628

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary+principles | important | scripts/aside-planning-artifacts.sh:106 | aside's CLEAN decision uses an unpinned git status --porcelain; under status.showUntrackedFiles=no an untracked planning file is invisible and the helper answers a false CLEAN — the repository's git-config-pin rule (test-git-config-pins.sh R3) states the pin as absolute, and the canonical suite is red at HEAD naming this line |   |
| F2 | primary | minor | scripts/test-aside-planning-artifacts.sh:208 | the harness never exercises the helper's wrong-argument-count guard: case 7a passes an empty string refused by the -d check, so removing the count check entirely leaves all 14 cases green |   |
| F3 | primary | minor | scripts/aside-planning-artifacts.sh:145 | any failed stash pop is labeled PLANNING-ARTIFACTS-CONFLICT, including local-modification refusals with zero conflict markers; exit semantics, kept stash and recovery path are unaffected |   |
| F4 | primary | minor | skills/flow/integrate.md:111 | integrate.md wraps each MOVED worktree's rebase while the plan's letter says aside before the loop; the per-worktree reading is the better behavior and is flagged for confirmation as intentional |   |
| F5 | principles | minor | scripts/aside-planning-artifacts.sh:17 | the planning-path set is now defined in two cross-naming scripts; a set change must be written in both or the commit-sweep guard and the stash helper silently disagree — a defensible WET tradeoff today, annotated at both ends |   |
| F6 | principles | minor | skills/flow/integrate.md:110 | the helper's contract is restated at three citation sites, multiplying the drift surface; dispatched agents read the skill prose and never the helper, so self-contained paragraphs are a defensible tradeoff |   |

findings-total: 6
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 deferred — any failed pop reports the same exit 1, keeps the stash and names git stash list; telling a local-modification refusal from a conflicted apply costs a new verdict protocol across the header contract, the harness case and two prose sites, for a label-only gain
finding-status: F4 fixed
finding-status: F5 deferred — the set membership is annotated at both ends and the two callers change for different reasons; extracting a shared planning-pathspecs helper is the move when the set actually grows a third member, not before
finding-status: F6 deferred — dispatched implementers read the skill prose and never the helper, so self-contained citation paragraphs are the deliberate trade; thinning them swaps drift surface for dispatch clarity

reproducers-total: 6
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 none — prose deviation between plan and citation
finding-reproducer: F5 none — structural duplication
finding-reproducer: F6 none — prose duplication

## Pass log

### Round 0

- base moved with path overlap on skills/flow/implement.md; git merge-tree --write-tree proved the rebase conflict-free, so the rebase-now action ran unasked (unattended run, handoff none) — rebased onto 396dc08, re-check CLEAR; slot shas cleared
- diff size 427 under cap; docs-only exit 1 (first non-doc path scripts/aside-planning-artifacts.sh) — resolved roster primary+principles runs unchanged; no operator addition this round — the resolved list ran alone
- probe
- context bundle rebuilt (single-repo); relocation comparison generated; dispatch recorded on agents project key

### Round 1

- fix round 1: F1 (important, primary+principles) fixed by pinning --untracked-files=normal on the dirt check; F2 fixed by harness case 7d; F4 resolved by a dated Correction paragraph in tasks.md; F3/F5/F6 deferred with reasons before the fix went out; fix commit aa1cc92 pushed; fix diff .superpowers/sdd/fix-round-1.diff; re-running primary and principles (both raised findings), each alone on the rerun pair
- both re-run slots clean: F1/F2 verified fixed (reproducer flip + named paths touched), F4 resolved by the Correction paragraph, F3/F5/F6 deferrals confirmed holding, no new defects in the fix diff; check-panel-findings-closed.sh: FINDINGS-CLOSED; panel closed, no render per /flow-fast (finding rows are the record)
fix-mutation: scripts/aside-planning-artifacts.sh — removed --untracked-files=normal from the dirt check (sandbox copy) — reproducer 0-primary-1.sh demonstrates again (exit 1); test-git-config-pins R3 naming the line 1→0 pre→post
fix-mutation: scripts/test-aside-planning-artifacts.sh — deleted the argument-count guard (sandbox copy) — case 7d fails (rc 1 ≠ 2) — verified by both re-run slots
fix-mutations-total: 2
## git log --stat

commit 9fd4d87cf185ce7371c6ab6f092c5d78133f3afa
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Sep 22 21:49:11 2026 +0300

    docs(flow): prefix the planning-path citations with their roots

 skills/flow/implement.md | 5 +++--
 skills/flow/integrate.md | 5 +++--
 2 files changed, 6 insertions(+), 4 deletions(-)

commit aa1cc921ee82a2f0d29c58253b4514c932829db0
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Sep 22 21:34:13 2026 +0300

    fix(scripts): pin untracked-files on the aside dirt check and exercise the argument-count guard

 scripts/aside-planning-artifacts.sh      |  2 +-
 scripts/test-aside-planning-artifacts.sh | 13 +++++++++++++
 2 files changed, 14 insertions(+), 1 deletion(-)

commit 6773d5277d0ec298bc2e7037924702a4a619daab
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Sep 22 20:54:41 2026 +0300

    docs(flow): cite aside-planning-artifacts.sh at both pipeline rebase sites

 skills/flow-contracts/finish-contract-run1.md   |  7 +++++++
 skills/flow/implement.md                        |  6 ++++++
 skills/flow/integrate.md                        | 10 +++++++++-
 skills/flow/scripts/aside-planning-artifacts.sh |  1 +
 4 files changed, 23 insertions(+), 1 deletion(-)

commit 377b31475bd74a18862aa359f6ffd98e5b14e8da
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Tue Sep 22 20:50:01 2026 +0300

    feat(scripts): add aside-planning-artifacts.sh, the pre-rebase planning-artifact stash

 scripts/aside-planning-artifacts.sh      | 157 ++++++++++++++++++++
 scripts/test-aside-planning-artifacts.sh | 246 +++++++++++++++++++++++++++++++
 2 files changed, 403 insertions(+)

## Session narrative

Implemented KAN-628 inline in one session: a new `scripts/aside-planning-artifacts.sh` (aside/restore of the planning paths around any pipeline rebase), its 16-assertion sandbox harness, the symlink-farm entry, and citations at both rebase sites — the unpushed-history autosquash fold in `skills/flow/implement.md` and the pre-landing sync in `skills/flow/integrate.md` plus one prose sentence in `finish-contract-run1.md`. TDD ran red-then-green; the compact panel (primary+principles, one bundled dispatch) returned verdict `fix` on pass 1 — the Important finding (the dirt check inherited the invoking machine's `status.showUntrackedFiles`, a stated-absolute git-config-pin violation that had the canonical suite red at HEAD) was fixed with one flag, a Minor added the missing argument-count harness case, and the plan deviation (per-worktree wrap vs before-the-loop) was recorded as a dated Correction paragraph; three Minors were deferred with reasons. Both targeted re-runs came back clean and the suite went 79/79. The session struggled twice: the plan-shape guard refused indented field lines (its column-0 grammar), and the installed-citations guard refused unprefixed path fragments twice — both taught the house conventions and were fixed by rewording, never by suppression. Jira was unavailable throughout (flowd has no FLOWD_JIRA_* credentials), so every transition degraded to the contract's one-line skip.
