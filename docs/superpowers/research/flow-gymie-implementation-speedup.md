# flow implementation speed-up in the gymie repos — research notes

Source: none

Explored via `/flow-research` on 2026-09-06. The operator's topic: "how can I speed up tasks
implementation in `flow` — what takes the most time?" The session began against this repository's
own runs, then the operator redirected the target to the three gymie checkouts
(`/Users/tweety53/Projects/gymie`, `gymie-frontend`, `gymie-admin-frontend`); every figure below
is from those unless it says otherwise. Two threads were taken to convergence — the serial
implementer loop inside `flow.sdd-tdd` (section 1) and the review panel's fix-round cost
(section 2) — plus one follow-up round on an item the first thread had left open (section 3). The earlier note `docs/superpowers/research/flow-speedup.md` (2026-09-04) covers the
pipeline's choreography, cost and operator-attention levers against this repository; most of its
levers have since landed (kan-431, kan-432, kan-433, kan-435, kan-436) and this note does not
restate them. Its still-unlanded levers — wave-parallel bundles, the verifier alongside panel pass
1, Minor-only fixes via the last implementer, the rationale split, the tail relay — remain open
there and are not re-analysed here.

Sources: the dev stats store at `http://127.0.0.1:4173` — `/api/v1/records/gymie-7c1f238a/<change>`
for dispatch rows and findings (timestamps are `+03:00`), `/api/v1/stage-runs` paged with
`offset` — and every dispatch's sidechain transcript under
`~/.claude/projects/-Users-tweety53-Projects-gymie/<session>/subagents/agent-<agentId>.jsonl`
(timestamps are UTC; the three-hour offset misled two early window reads, corrected below). The
eight changes read: kan-420, kan-421, kan-422, kan-423, kan-424, kan-425, kan-454, kan-455 — all
`/flow` runs between 2026-09-04 and 2026-09-06, 6–19 tasks each.

## 1. The serial implementer loop inside `flow.sdd-tdd`

### Where the wall clock goes

78 implementer and panel-fix dispatches across the eight changes, 1548 minutes of implementer wall
clock measured from each transcript's first to last timestamp:

- **Tool execution is 37% of implementer wall clock; Gradle is 12%.** A single Gradle invocation is
  3–40 s targeted (`--tests`), 50–110 s for a whole-module `:shared:desktopTest`, 100 s for the
  backend's `./gradlew test`. kan-441's premise that a frontend iteration waits on a 5–10-minute
  Gradle run does not hold in these transcripts.
- **63% is model latency times turn count.** Median API round trip 3–5 s, p90 10–17 s. Ordinary
  tasks are 30–90 API calls and 5–13 minutes; wall clock correlates with call count (r = 0.47),
  not with test time. Backend tasks are cheap — kan-454's six tasks took 22 minutes, 2–7 each;
  frontend tasks run 2–3× longer per task.
- **The long tail is three implementers of 50–65 minutes**, each 285–354 API calls on a context
  that reached 450–585k tokens, and together 30% of their changes' implementer time. They are
  three different shapes:
  - kan-425 task 6 (56 min, 300 calls, 44 Gradle runs, 31 failing): **one stubborn Compose UI
    test**, `aRevokedGrantRemovesTheLoggingButtonWithoutNavigating` in `GroupSessionFloorUiTest`,
    failed about twenty consecutive targeted runs; the implementer invented a `ScratchDebugTest`
    to probe it. The task bundled three source files with the `desktopTest` file.
  - kan-423 task 18 (65 min, 354 calls): **a detekt loop, not a test loop.** Three targeted
    `desktopTest` runs (green on the third), one whole-module run (it was the last code task, so
    FULL SUITE applied), and 25 lint runs — about sixteen consecutive failing
    `ktlintFormat`/`detekt` runs while the implementer shaved `AuthenticatedShell` in
    `AppScreens.kt` from 98 lines to under detekt's `LongMethod` 90 and `CyclomaticComplexMethod`
    26, one edit per run — plus `devStart`/`dbSeed`/`devStop` to "verify end-to-end in the
    running app", which nothing asked of it. Splitting the UI tests would not have changed it;
    section 3 corrects an earlier reading of this task as TARGETED TESTS non-compliance.
  - kan-455 task 6 (51 min, 285 calls, 14 Gradle runs, 4 failing): **a 15-file deletion** of
    List mode — 75 reads and 78 edits at 585k context, not a test problem.
- **Stalls, not work:** kan-421 task-4-fix 139 min on 87 calls (96 s per call), kan-424 task 6
  85 min on 78 calls, kan-454's first task boundary 18 min of which 17 were one
  `git diff --name-only` Bash call (14:42:38 → 14:59:20 UTC). These are the API-retry and
  laptop-sleep class `flow-speedup.md` section 8 documented; nothing in `/flow` shortens them.
- **Boundaries are solved.** Post kan-431, implementer-end to next-launch is 16–70 s. The larger
  first-boundary gaps are a conductor death and re-dispatch (kan-423, `conductor-2`) and the
  kan-454 stall above, preceded by 2 minutes of the conductor hand-editing `tasks.md` because the
  planner had abbreviated `**Files:**` paths (`profile/…`) that `check-task-commit-fields.sh`
  rejected.

### How the planner decides task granularity today

- `skills/flow/brainstorm-planner.md` hands granularity entirely to `superpowers:writing-plans`:
  "a task is the smallest unit that carries its own test cycle and is worth a fresh reviewer's
  gate" and "each step is one action (2–5 minutes)". Nothing bounds a task by file count, test
  count or test kind. `scripts/plan-dispatch-bundles.py` only merges tasks that share a
  `**Files:**` path or a `**Squash-with:**`; it never splits.
- The de-facto frontend pattern is "feature plus its `desktopTest` file in one task": kan-423 has
  eleven of twelve frontend tasks carrying exactly one `desktopTest` file beside 3–8 source files
  and 5–15 tests; kan-425 and kan-455 follow the same shape. UI tests were never meant to be split
  out; the plans consistently fuse them.

### Decision

The operator chose one lever: **the planner puts a task's `desktopTest` UI tests into their own
follow-on task.** The implementation task keeps its `commonTest` unit tests and its source files;
the UI-test task declares only the `desktopTest` file(s) under `**Files:**`, so the two tasks are
file-disjoint and `plan-dispatch-bundles.py` keeps them as separate bundles with no change to the
bundling rule. The UI-test implementer starts at a small context instead of iterating on the
450–585k one the feature work built up. A full `## test` run before the panel stays — the
operator confirmed that gate; what is not wanted is the whole suite mid-task, which the TARGETED
TESTS paragraph already forbids.

Declined: a hard cap on files per task (would have split kan-455 task 6), and an implementer
failure budget (same test failing N consecutive runs ends the turn with a hand-back). Both are
recorded in the Open list rather than dropped.

## 2. The review panel's fix-round cost

### Anatomy across seven changes

kan-420 is excluded as an overnight run (746 minutes on a sleeping laptop). The other seven ran
16–83 minutes of panel, median 32, about 1000 panel-minutes in total. The shape is the same every
time:

- **Pass 1 slots, 5–12 minutes.** `mutation` is the long pole in every change: 6–11 min, 45–66
  calls, 7–16 Gradle runs. `primary`, `principles` and `code-review-low` finish in 2–4 min.
- **A fix round in all seven.** 3–7 findings each; no change closed on pass 1. `panel-fix` runs
  5–18 minutes (kan-454's 65-minute one is a stall — 30 calls).
- **A delta re-run of all four slots, 1–6 minutes**, because every round raised something above
  Minor, which `skills/flow/review-panel.md`'s **Panel re-runs** reads as "every diff-reading
  slot re-runs on its delta".
- **Conductor glue, 5–25 minutes per panel.** kan-455's 56-minute panel: 14 minutes between the
  slots finishing and the fix dispatch — one minute reading the four reports, two minutes on
  `flow record finding --help`, then a `## Question` to the operator about three refused
  reproducers that waited ten minutes for an answer, after which the conductor wrote three
  reproducer scripts itself — and 7 minutes after the fix returned, in which the conductor ran the
  mutation-proof itself: copied files, ran `:shared:desktopTest` four times, fixed an import in a
  test it had edited. The contract's **The fix round mutation-proves what it changed** says
  "*you* then mutate each one", so that Gradle work sits on the 200k+ conductor context by design.

### What the rounds find

- 33 findings in the seven changes. The high-value ones — kan-422's Critical and three Majors,
  kan-454's Critical and High, kan-424's High, kan-455's Major — all came from pass 1, and nine of
  the 33 from `mutation`.
- The re-run round's findings were four stale plan fields and three nits: kan-455 F5–F7 and
  kan-454 F4 are `primary` flagging `Tests:`/`Baseline:` fields in `tasks.md` left stale by the
  fix round (kan-455's conductor then verified they were already satisfied and closed them);
  kan-425's three Minors are "structural nit", "cosmetic token-duplication nit", "naming nit" in
  the reviewer's own words.
- **Reproducers: 11 of 33 runnable.** The rest are `none —` because `check-panel-reproducers.sh`
  refuses shell metacharacters, so reviewers stop at a `grep` pipeline they cannot submit. Where a
  runnable one exists it is a script the conductor wrote afterwards —
  `.superpowers/sdd/reproducers/f1-panel-rise-dry.sh` (kan-455),
  `.superpowers/sdd/repro-f1-scientific-notation.sh` (kan-454), `.superpowers/sdd/reproduce-f1.sh`
  (kan-424) — and the guard accepted every one of those paths: a relative, worktree-contained
  script path is already a legal reproducer. What is missing is the slot being told so.

### Decisions

The operator chose all three levers:

- **(a) Delta-only re-run of the slots whose findings were fixed.** After a fix round, re-run on
  the delta only the slot(s) that raised a finding the round fixed, instead of every diff-reading
  slot. A slot that raised nothing keeps its pass-1 result on the unchanged code; the fix
  subagent's own mutation-proof (below) covers what the fix changed. This narrows **Panel
  re-runs**' "every diff-reading slot … re-runs on its delta" rule; the Bugbot/Mutation/Security
  clause ("re-run only when that slot raised a finding in the previous round or the previous
  round raised a new Critical") is already this shape and becomes the rule for every slot.
- **(b) The fix subagent carries the mutation-proof and the `tasks.md` field updates.** Its
  dispatch already carries the PLAN FIELDS paragraph, which it did not honour in kan-454 and
  kan-455; the mutation-proof moves from the conductor's own turns into the fix subagent's
  report, which names each mutated mechanism, the test that failed, and the restore — the
  conductor checks the report against the fix diff (the existing hunk walk) and runs no Gradle
  itself. `scripts/mutate-and-verify.sh` is the fix subagent's tool rather than the conductor's.
- **(c) A reviewer may name a script it wrote as its reproducer.** No guard change:
  `check-panel-reproducers.sh` and `run-reproducer.sh` already accept a relative,
  worktree-contained path with no metacharacters. The slot's dispatch prompt gains one sentence
  — when the demonstrating command needs a pipe, a quote or a glob, write it to
  `<abs-worktree>/.superpowers/sdd/reproducers/<ref>.sh` and record that path — so the
  "unverifiable → ask the operator" wait disappears.

## 3. Whole-module test runs inside a task — the planner, not the implementer

Opened as a follow-up on the Open-list item "the TARGETED TESTS paragraph is not always honoured",
which section 1 originally pinned on kan-423 task 18. One investigate-then-ask round, then closed.

- **The instruction** is the `**TARGETED TESTS:**` paragraph of `skills/flow/implement.md` §4,
  landed by kan-441 on 2026-09-05 and required on every implementer and panel-fix dispatch by
  `scripts/check-dispatch-paragraphs.sh`: "Run only the tests this task's `**Tests:**` field
  names, through the build tool's own selector — `--tests '<class>'` for Gradle … — once for RED,
  once for GREEN, and again only after a source edit. Never run the module or repository suite
  mid-task: the full `## test` list runs once per worktree at the last bundle, and again in
  `flow.verify`." The text is clear.
- **The pattern across 77 implementer dispatches:** 62 ran at least one whole-module test
  invocation, 132 in total. Thirteen ran more than two despite carrying named tests — kan-421
  task 3 (11), kan-420 task 2 (6), kan-423 task 17 (5) among them — and all but two of those
  thirteen predate kan-441 landing. In the four changes run after it (kan-424, kan-425, kan-454,
  kan-455) no named-tests task exceeded two. Implementer non-compliance is already fixed.
- **The root cause is the plan.** 24 task-step lines across kan-421, kan-422, kan-423, kan-425
  and kan-455 end each task with `./gradlew ktlintFormat`, `./gradlew ktlintCheck detekt`,
  `./gradlew :shared:desktopTest` — the planner writes the bare module suite into the task's own
  verify step, and the implementer obeys the plan over the prompt. kan-454's planner wrote
  `--tests` in five steps and its tasks ran 0–2 whole-module runs; kan-425's wrote the bare module
  in all eight. Nothing in `skills/flow/brainstorm-planner.md`'s writing-plans paragraph tells the
  planner otherwise. The cost is one 50–110 s `:shared:desktopTest` per task — 8–15 minutes on a
  7–9-task frontend change — and it is the run the FULL SUITE paragraph and `flow.verify` already
  make at the end.

### Decision

**One sentence in `skills/flow/brainstorm-planner.md`'s writing-plans paragraph:** a task's
verify step names its lint commands and `--tests '<class>'` for each `**Tests:**` entry, never the
bare module or repository suite — that run belongs to the last bundle's FULL SUITE and to
`flow.verify`. No `check-plan-shape.sh` rule (offered, declined): the post-kan-441 data shows the
implementer follows whatever the plan says, so fixing the plan template is the whole fix. The
detekt-shaving loop and the unprompted dev-stack start in kan-423 task 18 were not taken up as
threads.

## Step-by-step breakdown

### The implementer loop (`flow.sdd-tdd`, section 4 of `skills/flow/implement.md`)

**What:** One implementer subagent per file-disjoint bundle, serial per worktree, running
RED-GREEN-REFACTOR with targeted Gradle runs, committing per task; the conductor guards and ticks
each commit and launches the next bundle.
**Why:** The largest wall-clock block of an implementation run; in the gymie repos its cost is
turn count on a growing context (63% model latency), not Gradle (12%).
**Uses:** `skills/flow/implement.md` §4, `scripts/plan-dispatch-bundles.py`,
`scripts/gather-dispatch-context.sh`, `scripts/check-task-commit-fields.sh`, `flow tasks tick`,
`flow record dispatch`, `./gradlew :shared:desktopTest --tests`, `./gradlew test`.

### Split `desktopTest` UI tests into their own follow-on task (thread 1 lever)

**What:** The planner's plan template writes a frontend feature as two tasks — the source and
`commonTest` task, then a `desktopTest`-only task whose `**Files:**` names only the UI-test
file(s) — so the UI tests run in a fresh implementer at a small context.
**Why:** kan-425 task 6 spent 56 minutes and 300 calls at 447k context looping on one Compose UI
test inside the feature task; file-disjoint tasks stay separate bundles with no change to
bundling, and the pre-panel full-suite run still covers the pair.
**Uses:** `skills/flow/brainstorm-planner.md` (the writing-plans enrichment and the task-shape
paragraph), `superpowers:writing-plans` **Task Right-Sizing**, `scripts/plan-dispatch-bundles.py`
(unchanged), `scripts/check-plan-shape.sh`.

### A task's verify step names targeted tests, never the module suite (thread 3 lever)

**What:** The planner's writing-plans enrichment writes each task's verify step as its lint
commands plus `--tests '<class>'` per `**Tests:**` entry; the bare `:shared:desktopTest`,
`./gradlew test`, `go test ./...` or `npm test` appears in no task step.
**Why:** 24 step lines in five gymie plans instruct the whole-module run the implementer prompt
forbids, and post-kan-441 implementers follow the plan; one 50–110 s run per task, 8–15 minutes
per frontend change, duplicated by FULL SUITE and `flow.verify`.
**Uses:** `skills/flow/brainstorm-planner.md` (the writing-plans paragraph that tells the skill the
task shape), `skills/flow/implement.md` §4 TARGETED TESTS and FULL SUITE paragraphs (unchanged),
`scripts/check-plan-shape.sh` (unchanged).

### The review panel (`skills/flow/review-panel.md`)

**What:** Pass-1 slots from the settings-store roster (`primary`, `principles`,
`code-review-low`, `mutation` in these runs), findings as store rows, one fix subagent per round,
reproducer re-runs, the conductor's mutation-proof, delta re-runs, close on zero open findings.
**Why:** The second-largest block, 16–83 minutes per gymie change; every change ran a fix round,
and the re-run round plus the conductor's own glue are where the minutes go without findings.
**Uses:** `skills/flow/review-panel.md`, `scripts/check-panel-diff-size.sh`,
`scripts/check-panel-docs-only.sh`, `scripts/check-panel-reproducers.sh`,
`scripts/run-reproducer.sh`, `scripts/mutate-and-verify.sh`,
`scripts/check-panel-findings-closed.sh`, `flow record finding`, `flow record status`,
`flow record render -kind panel`.

### Delta-only re-run of the slots whose findings were fixed (thread 2 lever a)

**What:** After a fix round, dispatch on the delta only the slot(s) whose finding the round
fixed; every other slot keeps its pass-1 result as current.
**Why:** Seven re-run rounds produced four stale plan fields and three nits while costing 1–6
minutes of slots plus the conductor's dispatch and record turns each time.
**Uses:** `skills/flow/review-panel.md` **Panel re-runs** (the "every diff-reading slot" sentence
and the stale-result definition), `flow record dispatch`, `check-panel-diff-size.sh`.

### The fix subagent carries the mutation-proof and the plan-field updates (thread 2 lever b)

**What:** The fix subagent's dispatch and report gain the mutation-proof — mechanism mutated,
failing test, restore, one `fix-mutation:` line each — and its existing PLAN FIELDS obligation is
checked by the conductor against the diff before the round closes; the conductor runs no Gradle
and edits no test itself.
**Why:** kan-455's conductor spent 7 minutes running `:shared:desktopTest` four times on a 200k+
context to prove one fix, and `primary`'s re-run findings in kan-454 and kan-455 were plan fields
the fix subagent had been told to update and had not.
**Uses:** `skills/flow/review-panel.md` **The fix round mutation-proves what it changed** and the
fix-subagent dispatch paragraphs (PLAN FIELDS, REPORT FILE), `scripts/mutate-and-verify.sh`,
`<abs-worktree>/.superpowers/sdd/panel-fix-report-<round>.md`.

### A reviewer names a script path as its reproducer (thread 2 lever c)

**What:** Every slot's dispatch prompt states that a reproducer needing shell metacharacters is
written to `<abs-worktree>/.superpowers/sdd/reproducers/<ref>.sh` and recorded as that relative
path; the guards are unchanged.
**Why:** 22 of 33 findings carried `none —` for want of a submittable command, and the conductor
wrote such scripts itself in three changes — once after a ten-minute operator wait — and the
guard accepted every one.
**Uses:** `skills/flow/review-panel.md` (the "Every slot must supply, per finding, a reproducer"
paragraph), `scripts/check-panel-reproducers.sh` and `scripts/run-reproducer.sh` (unchanged,
already accept the shape), `scripts/reproducer-metachars.sh`.

## Open / undesigned

- A files-per-task cap in the plan template (kan-455 task 6, 15 files, 51 minutes) — declined this
  session.
- An implementer failure budget — the same test failing N consecutive runs ends the turn with a
  hand-back (kan-425 task 6, about twenty runs on one test) — declined this session.
- A task whose `**Tests:**` is `none` has nothing to target (kan-455 task 6, kan-454 task 6,
  kan-455 task 2 — 2–4 whole-module runs each); neither the TARGETED TESTS paragraph nor the
  section 3 planner sentence says what such a task's verify step should run.
- Lint runs are outside the TARGETED TESTS instruction; kan-423 task 18's 25 `ktlintFormat`/
  `detekt` invocations were a rule-by-rule shaving loop the lint-fix policy requires and nothing
  bounds.
- The planner abbreviated `**Files:**` paths in kan-454 (`profile/…`) and the conductor
  hand-edited `tasks.md` to satisfy `check-task-commit-fields.sh`; `check-plan-shape.sh` does not
  catch the shape.
- `flow.visual-verify` on kan-454 dispatched an unrecorded screenshot-baseline agent (no
  `flow record dispatch` row) — the same hidden-work class kan-433 closed for `flow.verify`.
- Stalls of 15–17 minutes on a single tool call (three sightings in eight changes) and a
  65-minute `panel-fix` on 30 calls — outside `/flow`'s control, but they dominate any change's
  mean.
