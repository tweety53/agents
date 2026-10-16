---
name: flow-self-review
description: Run a change's self-review pass, a one-shot opus dispatch — inside `/flow`'s integrate and `/flow-fast`'s verify, or standalone from the context bundle an older run saved; plan in full, fix and land every finding that is not big, file the big ones, mine the flow-eval cases the operator approves, record each in the flow store, rate, write the report, delete the bundle. Standalone, not a pipeline stage. Use for /flow-self-review.
allowed-tools: Bash(git:*), Bash(flow:*), Bash(go:*), Bash(scripts/check-self-review-report.sh:*), Bash(land-self-review-report.sh:*)
license: MIT
compatibility: Standalone, requires the change's default branch to be checked out and a saved context bundle at docs/self-review/<name>-context.md.
---

Run a change's self-review reasoning pass — the only one the pipeline has — over the bundle
`/flow`'s run 1 (**Run the self-review pass**,
`skills/flow-contracts/finish-contract-run1.md`, canonical for that bundle's shape) or **5. Verify**
(`skills/flow-fast/SKILL.md`) assembles: inside that run, or, as this standalone command, from the
bundle an older run saved on the change branch before landing. This file is canonical for the seven angles, what is fixed, what may be filed and what may be mined, the
filing-and-rating prompt, the store record and the report. **Neither the pass nor its fixes run
in this session**: step 2 dispatches the pass, and step 3 plans, fixes and reviews, each a one-shot
dispatch on the pair this file names — never the model this session runs on.

**This is a standalone command, not a pipeline stage.** It takes one change name, writes no
per-change state file, and marks no `flow stage` call.

**No flags.** The only argument is the change name.

**Announce at start:** "Using flow-self-review."

## Workflow

### 1. Resolve and refuse

`<bundle>` is `<project>/docs/self-review/<name>-context.md`. Absent: print every
`docs/self-review/*-context.md` present (or `none pending` when there are none) and stop.

`git branch --show-current` must be `<default-branch>` (the project's own default branch). A
checkout on any other branch: print the branch and stop.

**Both refusals bind the standalone command only.** A pass run inside `/flow`'s integrate or
`/flow-fast`'s verify holds its bundle in memory, runs on the change branch, and starts at step 2.

### 2. Read the bundle and run the seven angles

Dispatch the pass, in-run and standalone alike: key `self-review-<name>-pass`, on the `pass` row
of the pairs table (step 3), its prompt carrying the bundle in full (`<bundle>`
standalone, the bundle held in memory in-run), this step's rules and step 3's classification, the
dispatch paragraphs step 3 names and READ-ONLY REVIEW. It returns every finding with its angle, its
evidence and its classification — `noted` or fixable, in or out of scope, and for a fixable one its
counted blast radius and whether it is `big`. This session keeps the filing-and-rating prompt, the
store rows and the report.

The pass runs the seven-angle table over the bundle, every angle explicit —
present-but-empty for an angle with no findings, never omitted:

| # | Angle | Label |
|---|-------|-------|
| 1 | Problems encountered, and what pipeline change would avoid them | `flow-fix` |
| 2 | Token/time cost, and what would reduce it without quality loss | `flow-cost` |
| 3 | What went well, and how to reproduce it | `flow-improvement` |
| 4 | What could be automated or moved to a script | `flow-automation` |
| 5 | What could move to the Go app or its persistent storage | `flow-stats-app` |
| 6 | What can be sped up — wall-clock time the run spent waiting or working that a pipeline change would cut | `flow-speed` |
| 7 | What would make a good flow-eval case — a task, a review finding or a visual pair from this run | `flow-eval` |

Angle 5's remit covers the records the pipeline writes to files today and the derivation work
now done in Bash or by the agent — **not** what the SPA should display. Angle 6 is elapsed time, where angle 2 is what a step spends: serial steps that could run in parallel, slow guards, builds or test runs, waits, and redundant re-runs. A silent angle and a
skipped angle are indistinguishable to a reader, which is why an empty angle says so.

Angle 7 lists the run's tasks, review findings and visual-verification pairs that would make good
flow-eval cases — what `case add impl|review|visual` mines, per `<agents repo>/../flow-eval/README.md`,
the `flow-eval` checkout beside `<agents repo>` — one candidate per case, in its kind's template,
every field present:

- **impl** — the task, as `tasks.md` records it; the resulting code, an excerpt under its
  worktree-resolved `path:line`; why it is a good candidate.
- **review** — a Critical or Important finding only, never a Minor one, a Critical preferred
  over an Important; each finding with its severity, its location and the reviewer's own sentence; the
  bad code; the fixed code, with the fixing commit's sha; why it is a good candidate.
- **visual** — the actual app screenshot and the mockup screenshot, each by absolute path; the
  findings; why it is a good candidate.

Every path is absolute. A file that exists only in git history, or only inside a worktree that
cleanup removes, is extracted by this session before step 4 — `git -C <repo> show <sha>:<path>`
into `<state-dir>/flow-eval/<name>/`, `<state-dir>` being the path `flow state dir` prints for
that repository — and cited there. A candidate names no defect: it is recorded `noted`, never fixed or filed — the one `noted`
item offered, for mining in step 4. Angle 7 reads its candidates from the run's own work in any repository the
run touched, product code included, and reads beyond the bundle only for the commits and files
the bundle names.

An `In-run pipeline fix:` line in the bundle (**Pipeline defects found mid-run**,
`skills/flow-contracts/pipeline.md`) that names a sha is reported under angle 1 as fixed and
is never offered for filing; one that reads `deferred` is an angle-1 finding like any other.

**A finding is fixed or filed only from the seven angles.** A finding about the pipeline itself is
fixed in step 3 unless it is `big`, and offered under its angle when it is. A finding that names no defect — what
went well, under angle 3, or an observation with nothing to fix — is reported and recorded `noted`, never offered: **only a `big`
finding, or one whose fix step 3 dropped, is ever offered for filing**. **The pass covers
the pipeline and the project's own dev tooling — build scripts, dev-stack and test-harness
config, guards — and nothing else.** A finding about the project's own product code is out of
scope: it is never offered or filed, and goes to a fix run instead — a pass inside a run stops
here, at the end of this step (**the sweep**, `skills/flow-contracts/known-bugs.md`). A landing HEAD
whose UI no visual verifier saw is such a finding, never a risk line. The report carries no
section beyond the seven angles and the rating. The filing prompt is never waived: a pass with
no operator to answer it files and mines nothing and records every offered finding `declined`.

**One combined pass** — never seven separate reads. The pass covers what the bundle holds and
nothing beyond it — say so in the report's `**Deferred:**` line.

### 3. Fix every finding that is not `big`

**Every finding is classified before any work**, by the pass, per the blast-radius rule of **Pipeline defects found
mid-run** (`skills/flow-contracts/pipeline.md`): `big` when its fix touches 60 or more files — a
design choice never is, since the planner takes the recommended option, and neither is a
reach outside `<agents repo>`, whose fix goes where that section places it. An
`In-run pipeline fix:` line naming a sha is already fixed and is not classified again.

**Every finding that is not `big` is fixed and landed without asking**, on one branch for the
whole pass. A pass with none creates no worktree and dispatches nothing. **The planner never drops a
non-`big` finding because its first design would misfire** — a guard that would flag the
pipeline's own commits, a check that would refuse legitimate history. That objection is a design
constraint: the planner designs within it, measures the result against the history it named, and
plans the fix. A finding leaves step 3 unfixed only when it is `big`, or disproved — the planner
shows, with evidence, that the defect it names does not exist. A planner report that drops a
finding on any other ground is sent back to a fresh planner dispatch carrying the objection as a
constraint, never offered for filing in step 4 (kan-982: the `Task-Id:` guard and the scope check
were dropped as imprecise and offered for filing; the operator answered "fix them now"). A finding hard
to solve escalates the planner rather than leaving the loop: the re-plan, key
`self-review-<name>-plan-<n>`, may run on `flow-high` (effort `high`) in place of the planner
pair's `flow-medium`. Where `high` still cannot produce a
design that holds, the next step is effort `xhigh` — and an `xhigh` dispatch is never made
without the operator's approval, asked first with the finding, the objection and what `high`
returned; a refusal leaves the finding offered for filing in step 4. **This session plans,
changes and reviews nothing itself** (`design.md`, `one-shot-subagent-fix-loop`), save item 4's
ticks and archive commit: each numbered step below that does is a fresh dispatch on the pair it names, its prompt carrying the MODEL HANDSHAKE,
TOOLS, FOREGROUND BUILDS and NO DELEGATION paragraphs of **4. Execute (SDD + TDD)**
(`skills/flow/implement.md`) verbatim — a review's READ-ONLY REVIEW too — with **The
handshake** there applying to each reply. Every dispatch is one-shot: no `SendMessage` to it once
it returns. The no-progress escalation of **Fewest operator actions**
(`skills/flow-contracts/pipeline.md`) still runs on `flow-high`.

**The pairs are this file's own**, and **Pipeline defects found mid-run** cites them — one row per
key suffix, the table `check-loop-dispatch.sh` reads and refuses a loop dispatch off:

| Key suffix | `subagent_type` | Model | On a recorded escalation |
|---|---|---|---|
| `pass` | `flow-medium` | `opus` | — |
| `plan`, `plan-<n>` | `flow-medium` | `opus` | `flow-high` |
| `fix`, `fix-<r>`, `fix-g<g>`, `fix-v<v>` | `flow-low` | `opus` | `flow-high` |
| `review-1` | `flow-medium`, `flow-low` | `opus` | — |
| `review-<r>` | `flow-low` | `sonnet` | — |

**The planner pair** is the `plan` row, **the fixer pair** the `fix` row and **the review pair**
the two `review` rows: the first review on `flow-medium` or `flow-low` as the planner recorded —
`flow-medium` when it recorded none — and every re-review after a fix on `flow-low` and `sonnet`.

**Every loop dispatch is admitted and recorded mechanically.**
`<agents repo>/hooks/enforce-loop-dispatch.py` hands each Agent call whose prompt carries a
`Key: pipeline-fix-<k>-…` or `Key: self-review-<name>-…` line to
`<agents repo>/scripts/check-loop-dispatch.sh`, which refuses it off its row of the table above,
while an earlier dispatch of the loop is still open, or — any dispatch but the plan and the pass —
before a completed plan, and records the admitted one in flowd. The parent closes each dispatch
once it returns with `<agents repo>/scripts/pipeline-fix.sh end <key> <outcome>`: `completed`, or
`aborted` for one that did not finish or never ran — a dispatch denied after it was recorded is
closed `aborted` and may be dispatched again under the same key. A review is closed with
`<agents repo>/scripts/pipeline-fix.sh end <key> <report-file>`, its reviewer's report, whose `Critical: <n>` and
`Important: <n>` lines decide the outcome: `clean` when both are 0, `fix` otherwise. A dispatch on its row's escalation column is admitted only after
`<agents repo>/scripts/pipeline-fix.sh escalate <key>`, and every loop branch's landing push only
once `<agents repo>/scripts/check-loop-landing.sh` finds its plan, its fix and a last review
closed `clean`. For an own-branch loop in another project, `pipeline-fix.sh end` and `escalate`
run from that project's working directory, never from `<agents repo>`. A refusal or failure of any of these tools on a step this file prescribes is a
pipeline defect, fixed in-run per **Pipeline defects found mid-run**
(`skills/flow-contracts/pipeline.md`) and the loop resumed once the fix lands — never a stop or a
question to the operator.

`<agents-base>` is the one **Pipeline defects found mid-run** binds. `<fix-branch>` and
`<fix-base>` are where the fixes go (`design.md`, `fixes-ride-the-change`): inside a run whose
change's canonical repository is `<agents repo>`, the change's own branch in its own
worktree — `spectre/<name>` on `/flow`, `<name>` on `/flow-fast` — based at the commit the branch
holds when the pass starts; standalone, or inside
any other project's run, `self-review-<name>` based at `origin/<agents-base>`.

1. **Plan, then fix** — on `self-review-<name>`, this session creates the worktree with
   `<agents repo>/scripts/pipeline-fix.sh start <name>` — never the main checkout; on the change's
   own branch its worktree is already set up. It then dispatches key `self-review-<name>-plan` on the planner pair, its
   prompt carrying every non-`big` finding with its evidence and counted blast radius. The planner
   plans them all in full as one spectre change, `self-review-<name>`, in that worktree, per
   `skills/flow/brainstorm-planner.md`: brainstorming (section B), the artifacts (section C —
   `proposal.md`, and `design.md` with `## Context`, `## Decisions` and `## Open questions`) and
   writing plans (section D — a `tasks.md` in which each finding is carried by exactly one task, a
   `Build: red` task counting with the partner it folds into). `flow-medium` carries no Skill
   tool, so the prompt names the path of the installed superpowers plugin's
   `skills/brainstorming/SKILL.md` and `skills/writing-plans/SKILL.md` — on Claude Code the
   newest `~/.claude/plugins/cache/*/superpowers/*/skills/<skill>/SKILL.md` — and the planner reads and
   follows each file where those sections invoke its skill. **Every question it meets is
   auto-resolved to its recommended option** per **Auto-resolution**
   (`skills/flow-contracts/operator-prompts-auto-resolution.md`) and recorded as a decision: it
   never asks the operator. Section B's design file goes under
   `<agents repo>-worktrees/self-review-<name>/.superpowers/sdd/`, or
   `<agents repo>-worktrees/<name>/.superpowers/sdd/` on the change's own branch. Section D's
   guards — `spectre validate "self-review-<name>"`, `check-plan-shape.sh`,
   `check-plan-provenance.sh` and `check-task-build-green.sh` — pass before it commits. It runs none of what records or gates a pipeline run: no `flow stage` call, no
   `flow tasks count`, no **Decide** — no `decision.json`, no `## Decision` block — and no **Plan
   review gate**. `design.md`'s decisions record instead this pass's dynamic choices: the first
   review's effort, `medium` or `low`, and the grouping of tasks into fixer dispatches. It commits
   the artifacts as `chore(spectre): plan self-review-<name>`, and never merges or pushes.

   Then one fixer per recorded group, in series — key `self-review-<name>-fix`, `-fix-g<g>` when
   the planner recorded more than one group — on the fixer pair, its prompt naming its tasks. It
   implements them in that worktree, one commit per task with its declared `**Commit:**` subject
   and no `Task-Id:` trailer, never writing `<agents repo>/spectre/changes/` — the ticks are item
   4's — adding one test or guard that fails without the fix wherever the fix changes behaviour, and running
   the `<agents repo>/.flow/project.md` `## lint` lines its files need, save
   `check-task-records.sh`, which judges those ticks. It reports each task's
   commit, or a finding it found `big` once under way, left uncommitted. It never merges or
   pushes.
2. **Review** — key `self-review-<name>-review-<r>` on the review pair, over
   `git diff <fix-base>...<fix-branch>`, its prompt naming each finding beside its task's
   commit, judging the code against `tasks.md` and `design.md`'s decisions, and asking each review
   finding graded Critical, Important or Minor, and its report to open with one `Critical: <n>`
   and one `Important: <n>` line counting them.
3. **Fix the review's findings** — key `self-review-<name>-fix-<r>` on the fixer pair, its prompt carrying the
   review's report verbatim. Each fix is folded into the commit of the task it fixes
   (`git commit --fixup <that commit>`, then `git rebase --autosquash <fix-base>`), so
   every task stays one commit; a commit the review judges not to fix its finding, or to make
   things worse, is dropped from the branch. **A review whose every finding is Minor is closed by
   that one dispatch, with no re-review** — the no-re-review half of **Panel re-runs**
   (`skills/flow/review-panel.md`); unlike the panel's parent, this session fixes nothing inline,
   so the fix stays this dispatch — and step 4 is its verification. After a review that raised a
   Critical or Important, step 2 runs again, `<r>` plus one, until a review comes back clean or
   raises only Minors, under **Fewest operator actions** (`skills/flow-contracts/pipeline.md`). A
   dropped commit's finding is offered in step 4 as a `big` one is; so is a fix found `big` once
   under way.
4. **Verify** — once a review comes back clean, or a Minor-only review's fix is committed, this session ticks each task whose commit is on `<fix-branch>`
   (`flow tasks tick -C <fix worktree> self-review-<name> <n>`), runs
   `spectre archive self-review-<name>` in the fix worktree — `--force` only when a task stays
   unticked, its finding found `big` or its commit dropped — and commits both as
   `chore(spectre): archive self-review-<name>`, on either branch, before any landing push —
   `<agents repo>`'s push hook refuses an unarchived change on its default branch. That commit
   alone carries the ticks, and is made once. This session then runs
   `<agents repo>/scripts/pipeline-fix.sh verify <fix worktree>`. A red line is fixed by a fresh
   dispatch of step 3, key
   `self-review-<name>-fix-v<v>`, its prompt carrying the failing command and its output
   verbatim; then step 2, then this step's lint lines and tests again, `<v>` plus one. The
   autosquash keeps the archive commit last; a task commit dropped after it has its task unticked
   in the archive commit by this session (`git commit --amend`). A red branch never lands, and a
   red line is never waived — **Fewest operator actions** governs this loop as it does the review.
5. **Land once.** On the change's own branch nothing lands here: the fix commits ride the run's
   own route, whose rebase can rewrite their shas, so the committed report is their record and
   step 6 records no `fixed` row for them. The report's `fixed:` lines carry each commit's sha on
   the branch when written. **The landing push** — the one that puts the commits on `<base>` on
   merge and push, the branch push on a pull-request or manual route — each attempt of it, a
   retry after a rejected push included, is preceded, in whichever invocation makes it, by a
   re-read of each fix commit's sha by its subject, in the worktree of the change's set whose
   repository holds that sha (`git log --format='%h %s' HEAD --not origin/<base>`); a changed sha
   rewrites its `fixed:` line, committed in the canonical worktree, which carries the report, as a
   new report commit through `land-self-review-report.sh` with the report's own subject; a rewrite
   after the canonical worktree's landing push pushes its branch again, and on merge and push its
   `<base>`, a fast-forward. Once that
   push succeeds, the same invocation records one `fixed` row per `fixed:` line whose sha a
   repository of the change's worktree set carries and `flow self-review findings -change <name>`
   does not list yet, its angle and note
   read off the line and no blast radius; a rewrite after rows were already recorded adds a
   second row under the new sha. A pull request merged by squash or rebase rewrites the
   shas once more, and no commit then carries a fix's own sha. On `self-review-<name>`,
   land by `<agents repo>`'s `## default landing route`
   (`project-get.sh <agents repo> 'default landing route'`), without asking — merge and push as
   step 4 of **Pipeline defects found mid-run** states it, with `self-review-<name>` as the
   branch. The pass is not done until every fixed finding's commit is on `origin/<agents-base>`. Each fixed
   finding's sha is read off `<agents-base>` after the landing, never from
   the branch before its rebase; on a route that opens a pull request, once that pull request
   merges.

### 4. Explain, then ask

Every offered finding is explained before the operator answers — what was observed, what breaks,
and what the fix would be — as the turn's final chat message, per **The explanation is the turn's final message**
(`skills/flow-contracts/operator-prompts.md`): a filed issue is durable, and an explanation
arriving afterward describes something the operator did not agree to. Every fixed finding is one
line in the turn's final message, naming its landed sha; it is not offered. The filing ask,
the mining ask and the rating are **one `AskUserQuestion` call**, shape per **Operator prompts**
(`skills/flow-contracts/operator-prompts.md`): up to three multi-select questions of
three offered items each — findings to file or angle-7 candidates to mine, never both in one
question — every option prefixed with its angle's label, plus **None — file nothing** or
**None — mine nothing** as the default; the rating last, `5 — excellent` / `4 — good` / `3 — fine` / `2 — rough`, a `1`
typed through the tool's free-text "Other". More than nine offered items roll the overflow into one
further call of the same shape, without the rating. With nothing to offer, the call
carries the rating alone. A candidate is mined only when the operator selects it: silence, a
stalled prompt and **Auto-resolution** (`skills/flow-contracts/operator-prompts-auto-resolution.md`)
mine nothing. An angle-7 candidate's explanation in the message body is its template.

That message renders in Markdown; step 2's template stays the field list, this is its layout:

- **Findings** — one table, every finding a row: `#`, observed, what breaks, fix (sha).
- **Each flow-eval candidate** — its own `### <n> · <kind> case — <title>` heading, then a
  two-column field table: severity, location `path:line`, the reviewer's sentence verbatim, fix
  commit sha + subject, test names; an impl case adds task title, commit subject, files, tests
  before→after, code `path:line`; a visual case adds actual and mockup absolute paths, diff score,
  verdict.
- **Code** — every excerpt a fenced block with its language tag (` ```kotlin `, ` ```go ` …), real
  multi-line code taken from the commit with `git show <sha>`, never a one-line paraphrase. A review
  case shows **Bad** and **Fixed** as two separate blocks, each labelled above it, the fixed block
  the diff's post-image (or a ` ```diff ` block of the relevant hunk), plus the relevant test's key
  assertion. An impl case excerpts its core function(s) and lists the plan's worked figures as
  bullets. A visual case lists its departures in a measured app-vs-mockup table.
- Each candidate closes with a `**Why it's a good case:**` line; the message's last line is the
  decision: "reply with which to mine: any of 1–N, or none".

#### The multi-select variant

Some prompts ask the operator to choose any subset of several options, not exactly one. This
variant states:

- the question, with each option listed separately
- that the operator may select any subset of the listed options — none, one, or several
- one explicitly stated default — named by the call site — for what happens if the operator is
  silent

This contract fixes the shape, not the default's polarity: the call site chooses it, and states it
plainly. A safe default may resolve silence to the empty set (an explicit "None" option, marked
recommended) or to the full set (every listed option, silence needing no option of its own to name
it) — whichever matches what the options actually control. The live multi-select call sites,
the self-review filing and mining asks, choose the empty set: silence selects
**None — file nothing** and **None — mine nothing**.

### 5. File chosen findings

File each chosen finding as a Jira issue per **Labels on issues the pipeline creates** (`skills/flow-contracts/jira-integration-finish.md`), carrying its angle's label on top of that set.
`## jira` absent or `none` in `<project>/.flow/project.md`, or no Atlassian tooling available in
this session: print `⚠ Jira: skipped — <reason>` and record that finding `declined` instead.

Mine each chosen candidate into the `flow-eval` checkout once the change's commits are on its
repository's default branch — standalone, at once; in a run, after the landing push of item 5 of
step 3 on merge and push. On any other route nothing is mined: each chosen candidate's
`case add` command is named in step 8 for after the merge. The mining runs in a worktree, never
the `flow-eval` main checkout, one branch per pass —
`git -C <agents repo>/../flow-eval worktree add -b flow-eval-cases-<name> <agents repo>/../flow-eval-worktrees/flow-eval-cases-<name> origin/main`
— from that worktree's root, one `case add` per candidate, per `<agents repo>/../flow-eval/docs/cases.md`:

```bash
go run ./cmd/flow-eval case add impl --repo <code checkout> [--plan-repo <plan checkout>] --change <name> --task <n>
go run ./cmd/flow-eval case add review --repo <code checkout> [--plan-repo <plan checkout>] --change <name>
go run ./cmd/flow-eval case add visual --id <id> --mockup <mockup path> --capture <capture path> --verdict match|mismatch [--defect '<text>' [--region x,y,w,h]]…
```

`--repo` and `--plan-repo` are the repositories' main checkouts, standalone, or the change's
worktrees in a run; `--mockup` is the mockup's path inside its checkout — the main checkout, or
the worktree in a run — never the copy extracted into `<state-dir>`, which sits outside any checkout
and is refused. A case drafted from a worktree, by `--repo`, `--plan-repo` or `--mockup`, has its
`case.json` `source.repo` then set to the main checkout's directory name. Edit each draft as that file's "What to check and edit" asks, from what the
candidate's evidence answers; run `case validate` the same way; and commit each case
alone as `feat(cases): add <role> case <id>`. The branch is never merged or pushed: the
operator's review of its commits is the label approval the README names. No `flow-eval`
checkout there: print `⚠ flow-eval: skipped — <reason>` and mine nothing. A `case add` or
`case validate` that fails leaves its candidate unmined, its output named in the final report.

### 6. Record every finding in the flow store

One call per finding, every disposition alike, before the report is written — save a `fixed`
finding on the change's own branch, whose row step 5's landing push records:

```bash
flow self-review finding -change <name> -angle <label> -disposition fixed|filed|declined|noted [-ref <sha|KEY>] [-blast-radius <N>] -note '<the finding, one line>'
```

`-ref` is the landed sha for `fixed`, the issue key for `filed`, and omitted for `declined` and `noted`;
`-blast-radius` is step 3's count. An `In-run pipeline fix:`
line naming a sha gets no second row when `flow self-review findings -change <name>` already
lists that sha (**Pipeline defects found mid-run**, `skills/flow-contracts/pipeline.md`). A store failure is one
warning line and the pass continues — the report below is the durable record.

### 7. Write the report, delete the bundle, land it

Write `<project>/docs/self-review/<name>-self-review.md` in the shape
`<project>/scripts/check-self-review-report.sh` checks — one section per angle, all seven present; each
finding one line naming its angle's label, the finding, and its disposition (`fixed: <sha>`, `filed: <KEY>`,
`declined` or `noted`); an angle with no findings carrying the explicit none-marker — plus one line under the
title:

```
**Deferred:** reasoning pass run on <the model the pass dispatch's handshake named> from <bundle path>
```

An angle-7 candidate is a `noted` line naming its kind and whether the operator chose to mine
it, its template's fields beneath it as indented sub-bullets.

Delete `<bundle>`; a pass inside a run has no bundle file and commits the report alone, per its
run's own section. Run `<project>/scripts/check-self-review-report.sh` when the project declares
it in `<project>/.flow/project.md`'s `## lint` section; fix any violation before committing.
**Exit 2 — the guard cannot answer at all (a missing or unreadable target directory, an
unreadable report inside it, an internal coverage.sh call failing) — is not a violation**: report it and stop before the
commit, never commit a report the guard could not read.
Commit both paths in one commit through the landing chain's one script — the same invocation
`skills/flow/integrate.md` step 4 lands the report with, differing in the asserted branch,
the removed context-bundle path and the `--push` — since the round-trip through the seven-angle
pass and the filing-and-rating prompt above is long enough that the branch is worth re-checking
rather than trusted from step 1 alone:

```bash
land-self-review-report.sh "<project>" "<default-branch>" \
  "docs(self-review): <name> self-review report" \
  "docs/self-review/<name>-self-review.md" \
  "docs/self-review/<name>-context.md" \
  --push "<default-branch>"
```

A branch mismatch stops here — nothing is committed, pulled, or pushed. Staged work in the
checkout beyond the chain's own two paths refuses the commit (`LAND-FOREIGN-STAGED`) — clear
the staging or land from a clean checkout, never around the refusal. A rejected push leaves the
commit local and this run names it — never retried around.

### 8. Report

End naming the report path, the rating, the shas landed, the Jira keys filed and the flow-eval cases mined, by branch and sha (each `none` when empty), each chosen candidate left unmined with its `case add` command, and each product defect the pass met (**the sweep**, `skills/flow-contracts/known-bugs.md`).

## Guardrails

- **Never** push directly to a protected branch under any name but the project's own default
  branch, and never force-push.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Run a deferred self-review pass for `<name>` | `/flow-self-review <name>` |
