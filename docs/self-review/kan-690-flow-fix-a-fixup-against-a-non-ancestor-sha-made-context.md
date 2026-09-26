# Self-review context bundle for kan-690-flow-fix-a-fixup-against-a-non-ancestor-sha-made

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-690-flow-fix-a-fixup-against-a-non-ancestor-sha-made/tasks.md (absent)
skipped: spectre/changes/archive/kan-690-flow-fix-a-fixup-against-a-non-ancestor-sha-made/design.md (absent)
skipped: spectre/changes/archive/kan-690-flow-fix-a-fixup-against-a-non-ancestor-sha-made/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-690-flow-fix-a-fixup-against-a-non-ancestor-sha-made.md

# SDD ledger — kan-690-flow-fix-a-fixup-against-a-non-ancestor-sha-made

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: timed-out
- Started: 2026-09-25T20:45:39Z
- Tokens: not measured

## Dispatch 2 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles-2
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: timed-out
- Started: 2026-09-25T21:28:15Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-690-flow-fix-a-fixup-against-a-non-ancestor-sha-made-panel.md

# Review panel — kan-690-flow-fix-a-fixup-against-a-non-ancestor-sha-made

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary+principles | important | scripts/guard-autosquash.sh:69 | after fails open — a tasks-md that is a directory passes the -r gate, the sweep sweeps nothing, and the guard prints after ok where its contract promises exit 2 |   |
| F2 | primary+principles | important | skills/flow/implement.md:909 | the new guard citations name no root and check-installed-citations.sh exits 1 on exactly those lines |   |
| F3 | primary | important | skills/flow/review-panel.md:783 | task record wired the post-fold after on FIX_BASE while the diff wires <task-sha>^; deviation proven forced — after a real fold the plan wiring refuses every fold |   |
| F4 | primary+principles | minor | scripts/guard-autosquash.sh:75 | hex runs over 40 chars are silently skipped — a stale sha glued into a longer run passes as after ok |   |
| F5 | primary+principles | minor | scripts/guard-autosquash.sh:63 | git plumbing errors (exit 128, a non-commit target) reported as is-not-an-ancestor exit 1 — a fact git never determined |   |
| F6 | primary | minor | scripts/test-guard-autosquash.sh:29 | the harness SANDBOX root leaks every run; the house style registers its own sandbox for removal |   |
| F7 | principles | minor | skills/flow/implement.md:905 | incident rationale near-verbatim in the guard header and the implement.md paragraph — drift-prone, judged defensible WET |   |

findings-total: 7
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 deferred defensible WET — plan asked for the sentences in both places; drift accepted

reproducers-total: 7
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-primary-4.sh
finding-reproducer: F5 .superpowers/sdd/reproducers/0-primary-5.sh
finding-reproducer: F6 .superpowers/sdd/reproducers/0-primary-6.sh
finding-reproducer: F7 none — wording-drift observation, no runnable demonstration

## Pass log

### Round 0

- diff size: 254 lines, under cap
- not docs-only — scripts/guard-autosquash.sh
- roster: compact — 98
- no addition this round — the resolved list ran alone
- wall-clock breach: primary+principles elapsed 38m33s against the 15-minute ceiling — re-dispatching once
- wall-clock breach (2nd, same slot): primary+principles elapsed 23m21s against the 15-minute ceiling
- auto-resolved: slot breached the wall-clock ceiling a second time — how should this proceed? → Stop the run
- operator continuation: after 2x wall-clock breach the two completed pass-1 dispatches stand as the review evidence; no third dispatch — findings recorded from them and closed below
## git log --stat

commit 8c3f1cb6cc414716c33188fadf08a01d2aa7bd83
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Sep 26 03:10:47 2026 +0300

    docs(flow): root the ancestry guard citations

 skills/flow/implement.md    | 3 ++-
 skills/flow/review-panel.md | 2 +-
 2 files changed, 3 insertions(+), 2 deletions(-)

commit 6b16125b6cd75cfdb2cb295f51a1bc0e4f4aac18
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Sep 26 03:10:44 2026 +0300

    test(scripts): cover the hardened ancestry guard

 scripts/test-guard-autosquash.sh | 29 +++++++++++++++++++++++++++--
 1 file changed, 27 insertions(+), 2 deletions(-)

commit 39b75a6415aaa7532cc4bd0da585b35bcae1d637
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Sep 26 03:10:41 2026 +0300

    fix(scripts): close the ancestry guard's fail-open sweep

 scripts/guard-autosquash.sh | 35 ++++++++++++++++++++++-------------
 1 file changed, 22 insertions(+), 13 deletions(-)

commit 06beac1ae64ff6e2481f601c157a7a9f417f745c
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 25 23:38:49 2026 +0300

    docs(flow): assert fixup ancestry around the panel fix fold

 skills/flow/review-panel.md | 7 ++++++-
 1 file changed, 6 insertions(+), 1 deletion(-)

commit ac60271043c72ec4e1ef75b379467afd1e3498cc
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 25 23:38:46 2026 +0300

    docs(flow): assert fixup ancestry around the task-fix fold

 skills/flow/implement.md | 9 ++++++++-
 1 file changed, 8 insertions(+), 1 deletion(-)

commit 24b43435e14fbfddf72ce619dc082aceb4a2f169
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 25 23:32:41 2026 +0300

    test(scripts): cover the autosquash ancestry guard

 scripts/test-guard-autosquash.sh | 148 +++++++++++++++++++++++++++++++++++++++
 1 file changed, 148 insertions(+)

commit 7c2bf6b3ecaa3eb88dece3dc47f0d4286a20c7d8
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Fri Sep 25 23:32:38 2026 +0300

    feat(scripts): add the autosquash ancestry guard

 scripts/guard-autosquash.sh | 90 +++++++++++++++++++++++++++++++++++++++++++++
 1 file changed, 90 insertions(+)

## Session narrative

A /flow-fast run created the change from KAN-690: a mechanical ancestry guard
(scripts/guard-autosquash.sh, targets/after, exit 0/1/2) for the fixup-to-autosquash
sequence, its house-style test harness, and two paragraph wirings into
skills/flow/implement.md and skills/flow/review-panel.md. The run's review panel — one
bundled primary+principles dispatch — breached its 15-minute wall-clock ceiling twice
(38m33s, then 23m21s on the mandated re-dispatch), auto-resolved to Stop the run, and the
operator continued with the two completed dispatches standing as the review evidence; their
findings were recorded and closed in the store. This session struggled twice on harness
mechanics: the background SPA build silently ran from the main checkout (background calls
do not inherit the session's cd), producing false go vet/tsc failures until re-run with an
explicit -C path; and the guard-test suite's first parallel run starved
test-check-cleanup-complete.sh's 5-second wall-clock bound (3 failures that pass standalone
and on re-run). The fixes: the after sweep refuses a directory tasks-md (-f, exit 2),
targets resolves a sha before judging ancestry and says so, the sweep dropped its 40-char
ceiling so a glued sha is reported, the harness cleans its sandbox, and both citations are
<agents repo>/-rooted. Full lint battery green; guard suite 87/87.
