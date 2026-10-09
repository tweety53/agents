---
name: flow-self-review
description: Run a change's self-review pass, inline on this session's model — inside `/flow`'s integrate and `/flow-fast`'s verify, or standalone from the context bundle an older run saved; fix and land every finding that is not big, file the big ones, record each in the flow store, rate, write the report, delete the bundle. Standalone, not a pipeline stage. Use for /flow-self-review.
allowed-tools: Bash(git:*), Bash(flow:*), Bash(scripts/check-self-review-report.sh:*), Bash(land-self-review-report.sh:*)
license: MIT
compatibility: Standalone, requires the change's default branch to be checked out and a saved context bundle at docs/self-review/<name>-context.md.
---

Run a change's self-review reasoning pass — the only one the pipeline has — over the bundle
`/flow`'s run 1 (**Run the self-review pass**,
`skills/flow-contracts/finish-contract-run1.md`, canonical for that bundle's shape) or **5. Verify**
(`skills/flow-fast/SKILL.md`) assembles: inside that run, or, as this standalone command, from the
bundle an older run saved on the change branch before landing. This file is canonical for the six angles, what is fixed and what may be filed, the
filing-and-rating prompt, the store record and the report. **The pass runs inline, in this session, on whatever model it is already on; its fixes never do**
— step 3 runs every fix, review and re-review as a one-shot `flow-medium` dispatch: the fix on
`opus`, or `sonnet` when every fix is a literal edit with nothing it can break; the first review
on `opus`; every re-review on `sonnet`. The model is
picked by picking the model this session runs on (`/model`) before invoking this command, not by
anything this skill itself resolves.

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

### 2. Read the bundle and run the six angles

Read `<bundle>` in full. Run the six-angle table over it, inline, every angle explicit —
present-but-empty for an angle with no findings, never omitted:

| # | Angle | Label |
|---|-------|-------|
| 1 | Problems encountered, and what pipeline change would avoid them | `flow-fix` |
| 2 | Token/time cost, and what would reduce it without quality loss | `flow-cost` |
| 3 | What went well, and how to reproduce it | `flow-improvement` |
| 4 | What could be automated or moved to a script | `flow-automation` |
| 5 | What could move to the Go app or its persistent storage | `flow-stats-app` |
| 6 | What can be sped up — wall-clock time the run spent waiting or working that a pipeline change would cut | `flow-speed` |

Angle 5's remit covers the records the pipeline writes to files today and the derivation work
now done in Bash or by the agent — **not** what the SPA should display. Angle 6 is elapsed time, where angle 2 is what a step spends: serial steps that could run in parallel, slow guards, builds or test runs, waits, and redundant re-runs. A silent angle and a
skipped angle are indistinguishable to a reader, which is why an empty angle says so.

An `In-run pipeline fix:` line in the bundle (**Pipeline defects found mid-run**,
`skills/flow-contracts/pipeline.md`) that names a sha is reported under angle 1 as fixed and
is never offered for filing; one that reads `deferred` is an angle-1 finding like any other.

**A finding is fixed or filed only from the six angles.** A finding about the pipeline itself is
fixed in step 3 unless it is `big`, and offered under its angle when it is. **The pass covers
the pipeline and the project's own dev tooling — build scripts, dev-stack and test-harness
config, guards — and nothing else.** A finding about the project's own product code is out of
scope: the pass neither offers, files nor records one. A product defect a run finds is fixed
inside that run, before integrate lands the change — by a fix run at `IN_PROGRESS` — never
deferred to self-review. The report carries no
section beyond the six angles and the rating. The filing prompt is never waived: a pass with
no operator to answer it files nothing and records every offered finding `declined`.

**One combined pass** — never six separate reads. The pass covers what the bundle holds and
nothing beyond it — say so in the report's `**Deferred:**` line.

### 3. Fix every finding that is not `big`

**Classify every finding before any work** by the blast-radius rule of **Pipeline defects found
mid-run** (`skills/flow-contracts/pipeline.md`): `big` when its fix touches 60 or more files,
reaches outside `<agents repo>`, or needs a design choice only the operator can make. An
`In-run pipeline fix:` line naming a sha is already fixed and is not classified again.

**Every finding that is not `big` is fixed and landed without asking**, on one branch for the
whole pass. A pass with none creates no worktree and dispatches nothing. **This session changes
and reviews nothing itself** (`design.md`, `one-shot-subagent-fix-loop`): each numbered step
below that does is a fresh dispatch on the pair it names, its prompt carrying the
paragraphs **Every dispatch in the loop is one-shot** (**Pipeline defects found mid-run**,
`skills/flow-contracts/pipeline.md`) names — a review's READ-ONLY REVIEW too — with **The
handshake** there applying to each reply. Every dispatch is one-shot: no `SendMessage` to it once
it returns. The no-progress escalation of **Fewest operator actions**
(`skills/flow-contracts/pipeline.md`) still runs on `flow-high`.

**The fixer pair** is `subagent_type: flow-medium` on `opus`, or on `sonnet` when every fix it
carries is a literal edit with nothing it can break, the one-line reason recorded with the
dispatch. **The review pair** is `flow-medium` on `opus` for the first review (`<r>` = 1) and
`flow-medium` on `sonnet` for every re-review after a fix.

`<agents-base>` is the one **Pipeline defects found mid-run** binds. `<fix-branch>` and
`<fix-base>` are where the fixes go (`design.md`, `fixes-ride-the-change`): inside a run whose
change's canonical repository is `<agents repo>`, the change's own branch in its own
worktree — `spectre/<name>` on `/flow`, `<name>` on `/flow-fast` — based at the commit the branch
holds when the pass starts; standalone, or inside
any other project's run, `self-review-<name>` based at `origin/<agents-base>`.

1. **Fix** — on `self-review-<name>`, this session creates the worktree,
   `git -C <agents repo> worktree add -b self-review-<name> <agents repo>-worktrees/self-review-<name> origin/<agents-base>`
   — never the main checkout — runs `<agents repo>`'s `## worktree setup` in it
   (`project-get.sh <agents repo> 'worktree setup'`); on the change's own branch its worktree is
   already set up. It then dispatches key `self-review-<name>-fix` on the fixer pair, its prompt carrying
   every non-`big` finding with its evidence and counted blast radius. The fixer works in that
   worktree and makes one commit per finding with a module scope, adding one test or guard that
   fails without the fix wherever the fix changes behaviour, and running the
   `<agents repo>/.flow/project.md` `## lint` lines its files need. It reports each finding's
   commit, or a finding it found `big` once under way, left uncommitted. It never merges or
   pushes.
2. **Review** — key `self-review-<name>-review-<r>` on the review pair, over
   `git diff <fix-base>...<fix-branch>`, its prompt naming each finding beside its
   commit.
3. **Fix the review's findings** — key `self-review-<name>-fix-<r>` on the fixer pair, its prompt carrying the
   review's report verbatim. Each fix is folded into the commit of the finding it fixes
   (`git commit --fixup <that commit>`, then `git rebase --autosquash <fix-base>`), so
   every finding stays one commit; a commit the review judges not to fix its finding, or to make
   things worse, is dropped from the branch. Then step 2 again, `<r>` plus one, until a review
   comes back clean, under **Fewest operator actions** (`skills/flow-contracts/pipeline.md`). A
   dropped commit's finding is offered in step 4 as a `big` one is; so is a fix found `big` once
   under way.
4. **Verify** — once a review comes back clean, this session runs, in the fix worktree, only the
   `<agents repo>/.flow/project.md` `## lint` lines the touched files need, plus the tests
   covering them: `go test` of the touched packages, the `scripts/test-*.sh` harness of a touched
   script, `npx vitest run <file>` of a touched SPA file — each output through `tail`, never the
   full `## lint` or `## test` list. A red line is fixed by a fresh dispatch of step 3, key
   `self-review-<name>-fix-v<v>`, its prompt carrying the failing command and its output
   verbatim; then step 2, then this step again, `<v>` plus one. A red branch never lands, and a
   red line is never waived — **Fewest operator actions** governs this loop as it does the review.
5. **Land once.** On the change's own branch nothing lands here: the fix commits ride the run's
   own route, whose rebase can rewrite their shas, so the committed report is their record and
   step 6 records no `fixed` row for them. The report's `fixed:` lines carry each commit's sha on
   the branch when written. **The landing push** — the one that puts the commits on `<base>` on
   merge and push, the branch push on a pull-request or manual route — each attempt of it, a
   retry after a rejected push included, is preceded, in whichever invocation makes it, by a
   re-read of each fix commit's sha by its subject
   (`git log --format='%h %s' HEAD --not origin/<base>`); a changed sha rewrites its `fixed:` line,
   committed as a new report commit through `land-self-review-report.sh` with the report's own
   subject. Once that push succeeds, the same invocation records one `fixed` row per `fixed:` line
   whose sha `flow self-review findings -change <name>` does not list yet, its angle and note
   read off the line and no blast radius; a rewrite after rows were already recorded adds a
   second row under the new sha. A pull request merged by squash or rebase rewrites the
   shas once more, and no commit then carries a fix's own sha. On `self-review-<name>`,
   land by `<agents repo>`'s `## default landing route`
   (`project-get.sh <agents repo> 'default landing route'`), without asking — merge and push as
   step 4 of **Pipeline defects found mid-run** states it, with `self-review-<name>` as the
   branch. A rebase onto a moved `origin/<agents-base>` runs item 4, **Verify**, again before the push. The
   pass is not done until every fixed finding's commit is on `origin/<agents-base>`. Each fixed
   finding's sha is read off `<agents-base>` after the landing, never from
   the branch before its rebase; on a route that opens a pull request, once that pull request
   merges.

### 4. Explain, then ask

Every finding is explained in the message body first, before any prompt fires — what was
observed, what breaks, and what the fix would be. Every fixed finding is one line in the same
body, naming its landed sha; it is not offered. A prompt's option text cannot carry that
explanation, so the prompt records the decision only: a filed issue is durable, and an
explanation arriving afterward describes something the operator did not agree to. The filing ask
and the rating are **one `AskUserQuestion` call**, shape per **Operator prompts**
(`skills/flow-contracts/operator-prompts.md`): up to three multi-select questions of
three offered findings each, every option prefixed with its angle's label, plus **None — file nothing**
as the default; the rating last, `5 — excellent` / `4 — good` / `3 — fine` / `2 — rough`, a `1`
typed through the tool's free-text "Other". More than nine findings roll the overflow into one
further call of the same shape, without the rating. With no finding to offer, the call
carries the rating alone.

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
it) — whichever matches what the options actually control. The one live multi-select call site,
the self-review filing ask, chooses the empty set: silence selects **None — file nothing**.

### 5. File chosen findings

File each chosen finding as a Jira issue per **Labels on issues the pipeline creates** (`skills/flow-contracts/jira-integration-finish.md`), carrying its angle's label on top of that set.
`## jira` absent or `none` in `<project>/.flow/project.md`, or no Atlassian tooling available in
this session: print `⚠ Jira: skipped — <reason>` and record that finding `declined` instead.

### 6. Record every finding in the flow store

One call per finding, every disposition alike, before the report is written — save a `fixed`
finding on the change's own branch, whose row step 5's landing push records:

```bash
flow self-review finding -change <name> -angle <label> -disposition fixed|filed|declined [-ref <sha|KEY>] [-blast-radius <N>] -note '<the finding, one line>'
```

`-ref` is the landed sha for `fixed`, the issue key for `filed`, and omitted for `declined`;
`-blast-radius` is step 3's count. An `In-run pipeline fix:`
line naming a sha gets no second row when `flow self-review findings -change <name>` already
lists that sha (**Pipeline defects found mid-run**, `skills/flow-contracts/pipeline.md`). A store failure is one
warning line and the pass continues — the report below is the durable record.

### 7. Write the report, delete the bundle, land it

Write `<project>/docs/self-review/<name>-self-review.md` in the shape
`<project>/scripts/check-self-review-report.sh` checks — one section per angle, all six present; each
finding one line naming its angle's label, the finding, and its disposition (`fixed: <sha>`, `filed: <KEY>` or
`declined`); an angle with no findings carrying the explicit none-marker — plus one line under the
title:

```
**Deferred:** reasoning pass run on <model named in this session's own system prompt> from <bundle path>
```

Delete `<bundle>`; a pass inside a run has no bundle file and commits the report alone, per its
run's own section. Run `<project>/scripts/check-self-review-report.sh` when the project declares
it in `<project>/.flow/project.md`'s `## lint` section; fix any violation before committing.
**Exit 2 — the guard cannot answer at all (a missing or unreadable target directory, an
unreadable report inside it, an internal coverage.sh call failing) — is not a violation**: report it and stop before the
commit, never commit a report the guard could not read.
Commit both paths in one commit through the landing chain's one script — the same invocation
`skills/flow/integrate.md` step 4 lands the report with, differing in the asserted branch,
the removed context-bundle path and the `--push` — since the round-trip through the six-angle
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

End naming the report path, the rating, the shas landed and the Jira keys filed (each `none` when empty).

## Guardrails

- **Never** push directly to a protected branch under any name but the project's own default
  branch, and never force-push.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Run a deferred self-review pass for `<name>` | `/flow-self-review <name>` |
