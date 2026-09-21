# Self-review context bundle for kan-618-flow-stats-app-flowd-could-warn-when-a-stage

found: 3 of 7 sources; skipped: 4 of 7 sources
skipped: change summary (absent)
skipped: spectre/changes/archive/kan-618-flow-stats-app-flowd-could-warn-when-a-stage/tasks.md (absent)
skipped: spectre/changes/archive/kan-618-flow-stats-app-flowd-could-warn-when-a-stage/design.md (absent)
skipped: spectre/changes/archive/kan-618-flow-stats-app-flowd-could-warn-when-a-stage/narrative.md (absent)

## .superpowers/sdd/ledgers/kan-618-flow-stats-app-flowd-could-warn-when-a-stage.md

# SDD ledger — kan-618-flow-stats-app-flowd-could-warn-when-a-stage

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: glm-5.3-flash effort=high
- Commit: no commit
- Outcome: completed
- Started: 2026-09-21T18:00:06Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-618-flow-stats-app-flowd-could-warn-when-a-stage-panel.md

# Review panel — kan-618-flow-stats-app-flowd-could-warn-when-a-stage

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Minor | stats/internal/api/changes_test.go:50 | the diff touches changes_test.go (fakeStore gains beginStageSuperseded), but the file appears in no task Files field, breaking the diff-to-plan bijection; the edit itself is necessary and correct |   |
| P1 | principles | Minor | stats/internal/store/stageruns.go:139 | StageRun.SupersededRuns is set only by BeginStage and nil on every read path, so a StageRun value meaning depends on which method produced it and a reader cannot distinguish nothing-superseded from not-reported |   |
| F2 | primary | Minor | stats/internal/store/stageruns.go:465 | the plan-session begin path (insertPlanStageRun, reachable via flow stage begin -jira-key K -stage plan.session) neither supersedes nor reports an open same-session-token run, so KAN-618 named overlap stays silent there; pre-existing no-supersede design this diff inherits, not a regression |   |
| P2 | principles | Minor | stats/internal/store/stageruns.go:178 | BeginStage doc promises it closes every still-open run sharing the call session token, unscoped, but the JiraKey branch never supersedes, so the new write-time warning silently does not exist on that branch |   |
| F3 | primary | Minor | stats/cmd/flow/stage.go:411 | the multi-run warning possessive "(its end mark never landed)" is singular and dangles when "runs" is plural |   |

findings-total: 5
finding-status: F1 fixed
finding-status: P1 deferred the write-only field is documented at the type and its one consumer; reshaping BeginStage return trades an interface break for no reader benefit
finding-status: F2 deferred plan sessions share their token with nothing by design, so no same-session overlap exists there — pre-existing no-supersede decision, outside this plan tasks
finding-status: P2 fixed
finding-status: F3 fixed

reproducers-total: 5
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh
finding-reproducer: P1 none — reproducer run live by the slot against the real store, verified output in .superpowers/sdd/panel-report-0-principles.md; script not retained
finding-reproducer: F2 none — reproducer run live by the slot against the real store, verified output in .superpowers/sdd/panel-report-0-primary.md; script not retained
finding-reproducer: P2 none — reproducer run live by the slot against the real store, verified output in .superpowers/sdd/panel-report-0-principles.md; script not retained
finding-reproducer: F3 .superpowers/sdd/reproducers/0-primary-3.sh

## Pass log

### Round 0

- roster: compact — 66
- diff size: 404 changed lines, under cap — automatic proceed (exit 0)
- docs-only: no — stats/cmd/flow/stage.go; resolved roster runs unchanged
- no addition this round — the resolved list ran alone
- round raised five Minors only — no slot re-runs; F1/F3/P2 fixed inline, F2/P1 deferred
## git log --stat

commit 03b169a3b733b49e9b0d1d9db86e335b4b6bc04d
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 21:25:55 2026 +0300

    fix(stats): scope the supersede promise, fix warning copy

 stats/cmd/flow/stage.go           | 2 +-
 stats/internal/store/stageruns.go | 4 ++++
 2 files changed, 5 insertions(+), 1 deletion(-)

commit 8eab5c64eba38b174fa7f0bf69f5746023df9aeb
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 20:57:58 2026 +0300

    feat(flow): warn when a stage begin supersedes an open run

 stats/cmd/flow/stage.go              | 23 ++++++++++++++++++++++-
 stats/cmd/flow/stage_test.go         | 34 ++++++++++++++++++++++++++++++++++
 stats/internal/client/client.go      | 21 ++++++++++++++++++++-
 stats/internal/client/client_test.go | 21 +++++++++++++++++++++
 4 files changed, 97 insertions(+), 2 deletions(-)

commit a651b2cb84960e4acc6371151d958c821b8a992b
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 20:56:12 2026 +0300

    feat(api): return superseded stage runs on begin marks

 stats/internal/api/changes_test.go |  12 ++--
 stats/internal/api/stages.go       |  57 ++++++++++++++++++-
 stats/internal/api/stages_test.go  | 110 +++++++++++++++++++++++++++++++++----
 3 files changed, 163 insertions(+), 16 deletions(-)

commit 7f644e526aa879f8b3a6c2302208bd121d22c002
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Mon Sep 21 20:51:39 2026 +0300

    fix(store): report the runs a stage-begin supersede closes

 stats/internal/store/stageruns.go      | 42 ++++++++++++++++-
 stats/internal/store/stageruns_test.go | 84 ++++++++++++++++++++++++++++++++++
 2 files changed, 124 insertions(+), 2 deletions(-)

## Session narrative

This run implemented KAN-618 inline under a decided compact panel (primary+principles, one dispatch): flowd now reports, on the begin path, exactly which still-open runs a supersede closed — store RETURNING, API response field and daemon warn log, CLI stderr warning. It struggled most with ceremony fidelity rather than the code: the plan-shape guard demanded column-0 fields (one rewrite), the context-bundle shape argument wanted single-repo, the pass-note and close-guard calls each needed one retry for flags or cwd; and the one unresolved thread is an intermittent internal/store -race failure observed once here and once by the panel reviewer, never with a captured failing-test name, and not reproduced in five subsequent full runs — disclosed here rather than guessed at.
