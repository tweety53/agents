# Self-review context bundle for kan-805-flow-fix-process-lessons-live-only-in-per-run

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-805-flow-fix-process-lessons-live-only-in-per-run/tasks.md (absent)
skipped: spectre/changes/archive/kan-805-flow-fix-process-lessons-live-only-in-per-run/design.md (absent)
skipped: spectre/changes/archive/kan-805-flow-fix-process-lessons-live-only-in-per-run/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-805-flow-fix-process-lessons-live-only-in-per-run.md

# SDD ledger — kan-805-flow-fix-process-lessons-live-only-in-per-run

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: 5c8a0f955b1ece687e2b3eb57fb96f8c619baee2
- Outcome: timed-out
- Started: 2026-10-02T22:11:15Z
- Tokens: not measured

## Dispatch 2 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles-redisp1
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: 5c8a0f955b1ece687e2b3eb57fb96f8c619baee2
- Outcome: timed-out
- Started: 2026-10-02T22:33:21Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles-redisp2
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Diff base: 5c8a0f955b1ece687e2b3eb57fb96f8c619baee2
- Outcome: completed
- Started: 2026-10-02T23:17:08Z
- Tokens: not measured

## Dispatch 4 — panel-fix

- Task: no task
- Role: panel-fix
- Key: panel-fix-1
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-10-02T23:42:45Z
- Tokens: not measured

## Dispatch 5 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary
- Key: panel-1-primary
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-10-03T00:08:58Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-805-flow-fix-process-lessons-live-only-in-per-run-panel.md

# Review panel — kan-805-flow-fix-process-lessons-live-only-in-per-run

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | important | stats/cmd/flow/lesson.go:83 | a whitespace-only -topic (or "-") passes the CLI's plain-emptiness check, hits the daemon's normalize-empty 400, and the client maps 400 to ErrUnavailable — a caller mistake exits 1 with 'store unavailable: unexpected status 400', against the command's own usage contract; the CLI should check lessons.Normalize(topic) itself |   |
| F2 | primary | minor | stats/internal/lessons/lessons.go:146-168 | an unreadable brief/narrative file inside a readable root silently reads as found: 0 with no note line, against the package's own Result contract; the plan scoped the note line to whole roots only |   |
| F3 | primary | minor | 7634f2fc | commit 7634f2fc (the four no-op RecordStore fakes) is named by no task's **Commit:** field — compile-forced by the RecordStore widening; confirm intentional |   |
| F4 | principles | minor | stats/internal/lessons/lessons.go:28 | lessons.Root.ProjectKey is declared, plumbed by the handler, and read nowhere — a speculative field; drop it or give it a reader |   |

findings-total: 4
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed

reproducers-total: 4
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: F2 .superpowers/sdd/reproducers/0-primary-2.sh
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh
finding-reproducer: F4 .superpowers/sdd/reproducers/0-principles-1.sh

## Pass log

### Round 0

- roster: compact — 39
- diff-size: 1118 lines, under cap — proceeding
- docs-only: exit 1 — resolved roster runs; first non-doc path: spectre/changes/kan-805-flow-fix-process-lessons-live-only-in-per-run/verbatim-moves.txt
- no addition this round — the resolved list ran alone.
- ceiling breach: primary+principles — 18.9 min elapsed against the 15-minute ceiling (reports landed +16.7/+17.4 min); timed-out slot raises no finding; re-dispatching the slot once
- second ceiling breach: primary+principles — 23.2 min elapsed (reports landed +21.9/+22.7 min); second breach of the same slot; auto-resolution takes the recommended option — stop the run
- decided under: panel slot ceilings this run are 30 minutes — pass-1 and fix-round re-runs alike — by operator instruction 'raise the limit to 30 minutes and continue', replacing the 15-minute and 5-minute ceilings for this run; the two prior round-0 breaches closed timed-out under the old wording and stand as closed

### Round 1

- fix: one dispatch (panel-fix-1), 4 findings, 1 chunk; diff path: .superpowers/sdd/fix-round-1.diff pending write; fix commit a06cbcb4
- re-run disposition: primary re-runs (raised F1 Important); principles not re-run — raised only a Minor (F4); ceiling this run 30 min (operator instruction)
- rebased at round boundary — origin/main moved; branch rewritten, fix commit now 26520c61; merge base eb2cc688; held shas cleared
- re-run: primary clean — F1-F3 all fixed, no new defects at the sites; report .superpowers/sdd/panel-report-1-primary.md
- stage bookkeeping: flow.review-panel re-opened here — the stage row had ended stopped at the two ceiling breaches; every dispatch, finding and pass row of the raised-ceiling round 0 and fix round 1 landed beneath it as recorded
fix-mutation: stats/cmd/flow/lesson.go — normalization dropped from the topic check (back to plain topic == "") — TestRunLessonResolveRefusesWhitespaceTopic
fix-mutation: stats/internal/lessons/lessons.go — both per-file Unreadable appends flipped to the pre-fix silent skip — TestResolveUnreadableFileIsReported
fix-mutation: stats/internal/lessons/lessons.go — none — dead-field removal (F4): no executable behaviour to flip
fix-mutations-total: 3
## git log --stat

commit 26520c61fc4d632548011a380bb687573fea3286
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Oct 3 02:55:27 2026 +0300

    fix(stats): review findings from round 0

 stats/cmd/flow/lesson.go               |  6 +++-
 stats/cmd/flow/lesson_test.go          | 21 +++++++++++++
 stats/internal/api/lessons.go          |  2 +-
 stats/internal/lessons/lessons.go      | 56 +++++++++++++++++++++-------------
 stats/internal/lessons/lessons_test.go | 54 ++++++++++++++++++++++++++++----
 stats/internal/lessons/render.go       |  6 ++--
 6 files changed, 113 insertions(+), 32 deletions(-)

commit ef300cc355a918473150350a0005f0ea5200065e
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Oct 3 01:04:26 2026 +0300

    test(stats): teach the remaining RecordStore fakes the project-roots read

 stats/internal/client/client_test.go    | 4 ++++
 stats/internal/reconcile/record_test.go | 8 ++++++++
 stats/internal/web/embed_test.go        | 4 ++++
 3 files changed, 16 insertions(+)

commit fe4a5e2ad4929d9cd4b6a8d37b0c20beb1febf46
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Oct 3 01:01:57 2026 +0300

    docs(flow-contracts): the process-lessons home and its resolve command

 skills/flow-contracts/SKILL.md                     |  1 +
 skills/flow-contracts/lessons.md                   | 44 ++++++++++++++++++++++
 skills/flow-fast/SKILL.md                          |  7 +++-
 skills/flow/brainstorm-planner.md                  |  4 ++
 skills/flow/verify-and-handoff.md                  |  6 ++-
 .../verbatim-moves.txt                             | 24 ++++++++++++
 stats/internal/guard/references.go                 |  1 +
 7 files changed, 83 insertions(+), 4 deletions(-)

commit 79c2f0257dae558205c0d10f1d9a7b6dda290903
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Oct 3 00:53:40 2026 +0300

    feat(stats): flow lesson resolve over the daemon

 stats/cmd/flow/lesson.go           |  98 ++++++++++++++++++++++++++++++++++
 stats/cmd/flow/lesson_test.go      | 106 +++++++++++++++++++++++++++++++++++++
 stats/cmd/flow/main.go             |   3 ++
 stats/internal/api/changes_test.go |   5 ++
 stats/internal/api/lessons.go      |  63 ++++++++++++++++++++++
 stats/internal/api/lessons_test.go | 105 ++++++++++++++++++++++++++++++++++++
 stats/internal/api/records.go      |   7 +++
 stats/internal/api/records_test.go |   7 +++
 stats/internal/api/server.go       |   2 +
 stats/internal/client/lessons.go   |  47 ++++++++++++++++
 10 files changed, 443 insertions(+)

commit c7d19525507ba33dcd39af6748a41595a2562ad5
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Oct 3 00:45:43 2026 +0300

    feat(stats): scan briefs and archived narratives for a lesson

 stats/internal/lessons/lessons.go      | 209 +++++++++++++++++++++++++++++++++
 stats/internal/lessons/lessons_test.go | 144 +++++++++++++++++++++++
 stats/internal/lessons/render.go       |  41 +++++++
 stats/internal/lessons/render_test.go  |  54 +++++++++
 stats/internal/store/projects.go       |  45 +++++++
 stats/internal/store/projects_test.go  |  79 +++++++++++++
 6 files changed, 572 insertions(+)

## Session narrative

Created the process-lessons home and its resolver: `docs/briefs/` as the canonical briefs location (contract `skills/flow-contracts/lessons.md`, cited from both run shapes' narrative steps and both ticket-reading moments), and `flow lesson resolve -topic <words>` — a read endpoint that scans every registered project's briefs and archived narratives fresh on every call, ranked briefs-first, with per-file unreadables reported in `Result.Unreadable` note lines rather than silently skipped. Three implementation commits plus the docs commit, then the panel. The panel took three dispatches to get a verdict: two ceiling breaches at the default 15-minute slot bound (both substantive, both voided by the breach course), resolved when the operator raised this run's ceiling to 30 minutes and the third dispatch returned in 18.7. The struggle worth recording: the resolved findings (empty-topic surfacing as store-unavailable; per-file unreadable silent; a write-only struct field; an undeclared fourth commit) were fixed in one fixer dispatch, but all four original reproducers then refused as ambiguous — each had to be re-authored against the fixed tree before the pinned re-runs would accept the flips, and F3's subject being the untracked plan record meant its two-leg proof ran in place rather than in a scratch worktree. The panel re-run confirmed all three primary findings fixed with no new defects. Two rebases mid-run (origin/main moved twice) were resolved with no path overlap; one concurrency test in `internal/reconcile` flaked once under full-suite load and passed on three subsequent runs — left unrepaired, no introducing commit nameable.
