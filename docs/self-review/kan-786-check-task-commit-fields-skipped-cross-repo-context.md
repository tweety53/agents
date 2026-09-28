# Self-review context bundle for kan-786-check-task-commit-fields-skipped-cross-repo

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-786-check-task-commit-fields-skipped-cross-repo/tasks.md (absent)
skipped: spectre/changes/archive/kan-786-check-task-commit-fields-skipped-cross-repo/design.md (absent)
skipped: spectre/changes/archive/kan-786-check-task-commit-fields-skipped-cross-repo/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-786-check-task-commit-fields-skipped-cross-repo.md

# SDD ledger — kan-786-check-task-commit-fields-skipped-cross-repo

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-28T19:55:06Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-786-check-task-commit-fields-skipped-cross-repo-panel.md

# Review panel — kan-786-check-task-commit-fields-skipped-cross-repo

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary+principles | Minor | stats/internal/guard/taskcommitfields.go:138-143 | the commit map's malformed-entry and duplicate-worktree refusal branches have no test case (cases 151/152 cover the other two refusals) — behavior verified correct via the real shim; the gap is coverage only |   |
| F2 | primary | Minor | stats/internal/guard/check_task_commit_fields_test.go:1359 | tcfMapRepos deviates from the plan's stated fixture model (case 72's fx.repo + link.md): bare repos resolving via the absent-dir branch instead — flagged for confirmation |   |

findings-total: 2
finding-status: F1 deferred a round that raised nothing above Minor fixes no Minor — behavior verified correct by the round's live shim run; the table cases ride the next change
finding-status: F2 deferred the fixture models the absent-dir shape a real cross-repo run presents, deliberately; a link-route map case is beyond the ask

reproducers-total: 2
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 none — fixture-model deviation flagged for confirmation, not a runnable defect; fix adds a link.md map case

## Pass log

### Round 0

- roster: compact — 41
- diff-size: 432 under cap — proceed automatic
- docs-only: no — first non-doc path scripts/check-task-commit-fields.sh — resolved roster primary+principles runs
- no addition this round — the resolved list ran alone
- base moved at entry: 2 commits, no overlap — rebased clean onto 157e4bd0, re-checked CLEAR
## git log --stat

commit 853635a27ce3f63e93e1783359abbb169c773e3e
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 28 23:15:23 2026 +0300

    docs(flow): root the commit-map citation

 skills/flow/implement.md | 4 ++--
 1 file changed, 2 insertions(+), 2 deletions(-)

commit c96f1ad6e293ed0194464fb1e143607ab202c8bf
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 28 23:13:25 2026 +0300

    docs(known-bugs): fold the new entries under the existing heading

 KNOWN-BUGS.md | 10 ++--------
 1 file changed, 2 insertions(+), 8 deletions(-)

commit 2f5158b9fdd9d02a8d9cadb0a32189eac2176d28
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 28 23:12:24 2026 +0300

    docs(known-bugs): record the panel's two deferred minors

 KNOWN-BUGS.md | 8 ++++++++
 1 file changed, 8 insertions(+)

commit f1e30d293b9eaf417c8924c223b37da9aefdc9e2
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 28 22:51:34 2026 +0300

    docs(flow): check task fields across every repository on cross-repo changes

 skills/flow/implement.md | 12 ++++++++++++
 1 file changed, 12 insertions(+)

commit 995b015cef0443582d76543ea3e62c61a6277d10
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 28 22:50:57 2026 +0300

    docs(scripts): document the commit-map calling convention

 scripts/check-task-commit-fields.sh | 19 +++++++++++++++++++
 1 file changed, 19 insertions(+)

commit b1dfedb2f257821765124fc2a045da6c99719bc1
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 28 22:50:26 2026 +0300

    feat(guard): accept a per-repository commit map

 .../guard/check_task_commit_fields_test.go         | 193 +++++++++++++++++++
 stats/internal/guard/taskcommitfields.go           | 208 ++++++++++++++++-----
 2 files changed, 356 insertions(+), 45 deletions(-)

## Session narrative

This run verified KAN-786's defect against the tree before planning — a task whose `**Files:**`
span two repositories cannot reach exit 0 through any single-repository invocation of
`check-task-commit-fields.sh` (the other repository's declared paths read as
declared-but-untouched), which is why cross-repo runs skipped the guard by hand — and
implemented the issue's first named fix: the third argument now also accepts a per-repository
commit map (`<worktree>=<sha>[,<worktree>=<sha>…]`), with the verdict merged across the listed
repositories (union of changed paths, concatenated diffs against `Tests:`, summed `@Test`
counts against `Baseline:`, the declared subject required on every listed commit, the tree
check satisfied by any listed tree; the recorded measured-command check skips beyond one pair,
never a verdict). Cases 147–155 were written first and captured failing against the unmodified
guard before the implementation turned them green; the shim header and `implement.md`'s
task-close step document and mandate the form. The review panel ran one bundled
primary+principles dispatch, raised no Critical or Important, and its two Minors (missing
refusal-branch test cases; the fixture modelling the absent-dir rather than link.md satellite
shape) were deferred to `KNOWN-BUGS.md` under the round rule that a round raising nothing above
Minor fixes no Minor. Verification: the full `## lint` list clean (one citation-prefix hit in
the new paragraph, fixed, not suppressed), `run-guard-tests.sh` 52/52, the guard package green
under `-race`, and the normative inventory unchanged byte for byte. Where it struggled: the
first tasks.md draft indented the field lines under the step checkboxes — exactly the shape
`build-green.md` warns about — and the shape and build-green guards caught it before the decide
step; the base moved twice commits mid-run and the panel entry rebased clean; case 151's
fixture initially `rev-parse`d `HEAD~1` on a single-commit repo and was rewritten to pass a
non-empty fourth argument directly; and the tree-marker bracket the panel contract names around
a dispatch was not snapshotted for this round — the plan-tree guard (`check-plan-unchanged.sh`)
was, and it verified clean — a narrowing this narrative records rather than hides.
