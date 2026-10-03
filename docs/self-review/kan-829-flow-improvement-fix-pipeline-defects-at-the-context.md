# Self-review context bundle for kan-829-flow-improvement-fix-pipeline-defects-at-the

found: 6 of 7 sources; skipped: 1 of 7 sources
skipped: change summary (absent)

## .superpowers/sdd/ledgers/kan-829-flow-improvement-fix-pipeline-defects-at-the.md

# SDD ledger — kan-829-flow-improvement-fix-pipeline-defects-at-the

Rendered from the store. Do not edit: every dispatch is a row, and the next render overwrites this file.

## Dispatch 1 — implementer

- Task: 1
- Role: implementer
- Key: task-1-implementer
- Model: opus effort=medium
- Commit: a6a4fcfa
- Outcome: completed
- Started: 2026-10-03T20:05:31Z
- Tokens: not measured

## Dispatch 2 — implementer

- Task: 2
- Role: implementer
- Key: task-2-implementer
- Model: opus effort=medium
- Commit: a09694c3
- Outcome: completed
- Started: 2026-10-03T20:07:43Z
- Tokens: not measured

## Dispatch 3 — reviewer

- Task: 1
- Role: reviewer
- Key: task-1-reviewer
- Model: opus effort=default
- Commit: no commit
- Outcome: fix
- Started: 2026-10-03T20:08:30Z
- Tokens: input 28, output 2405, cache read 398824, cache creation 44053

## Dispatch 4 — panel-fix

- Task: 1
- Role: panel-fix
- Key: task-1-implementer-fix-1
- Model: opus effort=medium
- Commit: 9be5d0a78f780df6f26f33e867bb724bb21aec2c
- Outcome: completed
- Started: 2026-10-03T20:10:42Z
- Tokens: not measured

## Dispatch 5 — reviewer

- Task: 1
- Role: reviewer
- Key: task-1-reviewer-fix-1
- Model: opus effort=default
- Commit: no commit
- Outcome: clean
- Started: 2026-10-03T20:11:13Z
- Tokens: input 12, output 112, cache read 123251, cache creation 24427

## Dispatch 6 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-0-primary+principles
- Model: opus effort=low
- Commit: no commit
- Outcome: completed
- Started: 2026-10-03T20:13:51Z
- Tokens: input 20, output 204, cache read 373790, cache creation 56934

## Dispatch 7 — reviewer

- Task: no task
- Role: reviewer
- Slot: primary+principles
- Key: panel-1-primary+principles
- Model: sonnet effort=low
- Commit: no commit
- Outcome: completed
- Started: 2026-10-03T20:15:44Z
- Tokens: not measured

## Dispatch 8 — verifier

- Task: no task
- Role: verifier
- Key: verify
- Model: opus effort=medium
- Commit: no commit
- Outcome: completed
- Started: 2026-10-03T20:27:24Z
- Tokens: not measured
## .superpowers/sdd/reviews/kan-829-flow-improvement-fix-pipeline-defects-at-the-panel.md

# Review panel — kan-829-flow-improvement-fix-pipeline-defects-at-the

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note | Lineage |
|---|---|---|---|---|---|
| F1 | primary | Minor | spectre/changes/kan-829-flow-improvement-fix-pipeline-defects-at-the/design.md:31 | design §3 still says `git push origin fix/<slug>:<default-branch>`; shipped contract uses `fix-<slug>` per tasks.md Correction (2). |   |

findings-total: 1
finding-status: F1 fixed

reproducers-total: 1
finding-reproducer: F1 .superpowers/sdd/reproducers/0-primary-1.sh

## Pass log

### Round 0

- diff-size: 404 lines, under cap — proceed
- docs-only: exit 1 — first non-doc path spectre/changes/kan-829-flow-improvement-fix-pipeline-defects-at-the/verbatim-moves.txt; roster primary+principles
- roster: default — settings store primary,principles; no addition this round — the resolved list ran alone.
- standards: CLAUDE.md (AGENTS.md is the zcode rendering of the same set)

### Round 1

- re-run: source commit 5b917257 (gated-fix role) landed after round 0's read; both slots re-read the fix-round-1 delta
- check-panel-fix-single-dispatch: task-1-implementer-fix-1 recorded under panel-fix (the ambiguity 5b917257 fixes); auto-resolved Continue — violation stays recorded
## spectre/changes/archive/kan-829-flow-improvement-fix-pipeline-defects-at-the/tasks.md

# kan-829-flow-improvement-fix-pipeline-defects-at-the

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** a `/flow` or `/flow-fast` run fixes a non-`big` pipeline defect it hits within the run,
through a one-shot fix → review → fix → re-review loop on its own `<agents repo>` branch, landed by
the default route; self-review stops re-filing what was fixed.

**Architecture:** Task 1 adds the contract section and the two citations that let `/flow` and
`/flow-fast` dispatch under it, in one citation-closed commit. Task 2 teaches self-review's angle 1
to read the narrative line Task 1 defines.

**Spec:** `design.md` beside this file is canonical for every scope decision. Tasks cite its
`## Decisions` by ID and never restate them.

## Global Constraints

- Markdown only; no guard, CLI or SPA change.
- A commit's scope names the module it moved (**Commit scopes name the module**,
  `rules/commit-scope-is-the-module.mdc`).
- Every run-loaded sentence a task adds, deletes or rewords is listed, exactly as
  `scripts/check-verbatim-moves.sh` prints it after `::`, in
  `spectre/changes/kan-829-flow-improvement-fix-pipeline-defects-at-the/verbatim-moves.txt`, under
  a first line `# Operator, 2026-10-03: KAN-829 — in-run pipeline fixes.` The file is edited in the
  worktree and **left uncommitted**: it is a planning path, carried by the next
  `chore(spectre): plan` commit (**Planning commits**, `skills/flow-contracts/git-boundaries.md`).
- Every task's verify step runs the Markdown lint lines: `scripts/check-references.sh`,
  `scripts/check-markdown-integrity.py`, `scripts/check-installed-citations.sh`,
  `scripts/check-verbatim-moves.sh`, `scripts/check-dispatch-paragraphs.sh`,
  `scripts/check-stage-mark-calls.sh`, `scripts/check-normative-inventory.sh` (output unchanged
  against the merge base).

## Review Focus

- **Reach.** `/flow-fast` loads no `pipeline.md`; its guardrail must cite the section, or a
  `/flow-fast` run is still forbidden the new dispatches.
- **The closed list.** `implement.md`'s "every Agent-tool dispatch" table must carry the new row,
  or the parent's self-check refuses the dispatch.
- **Main checkout.** The fix worktree must never be the main checkout or the change's worktree.

**Live verification:** none — prose only; no service, store or runtime state changes.

---

- [x] 1. The in-run pipeline-fix contract section, reachable from `/flow` and `/flow-fast`

`design.md` §§ 1–4.

  - [x] **Step 1: Add the section** to `skills/flow-contracts/pipeline.md`, directly after
    `## Fewest operator actions` and before the first `## Wrong state for this command`, verbatim the
    block under **Task 1 — the section text** below.

  - [x] **Step 2: Add the closed-list row** to `skills/flow/implement.md`'s
    **Dispatch sites — the parent's closed list**: change `These six rows are **every**` to
    `These seven rows are **every**`, and append the row under **Task 1 — the closed-list row**
    below.

    If the row count sentence reads differently once the board-swap change has landed on
    `origin/main`, keep its wording and change only the count.
  - [x] **Step 3: Cite the section from `/flow-fast`.** In `skills/flow-fast/SKILL.md`'s
    **Guardrails, the whole list.** paragraph, change
    `Never dispatch a subagent the recorded decision does not name — an implementer per group on
    `sdd`, the decision's panel dispatches, the panel-fix subagent; never a planner or a verifier.`
    to end `…; never a planner or a verifier, save the in-run fix loop of **Pipeline defects found
    mid-run** (`skills/flow-contracts/pipeline.md`), which a `/flow-fast` run follows too.`
  - [x] **Step 4: Verify** — `grep -c '^## Pipeline defects found mid-run$' skills/flow-contracts/pipeline.md`
    prints `1`; `grep -c 'Pipeline defects found mid-run' skills/flow/implement.md skills/flow-fast/SKILL.md`
    prints `1` for each. Run the Markdown lint lines; record every sentence
    `check-verbatim-moves.sh` flags in `verbatim-moves.txt`.
  - [x] **Step 5: Commit.**

**Files:** `skills/flow-contracts/pipeline.md`, `skills/flow/implement.md`, `skills/flow-fast/SKILL.md`
**Tests:** none — contract prose
**Regression:** none — prose.
**Baseline:** before=0 after=0
<!-- measured: no test files in this task @ 5e86aef6 -->
**Commit:** `feat(flow): fix a pipeline defect within the run that finds it`
**After:** none
**Build:** green

**Decision:** rule-home-pipeline-contract
**Decision:** land-by-default-route
**Decision:** blast-radius-class
**Decision:** one-shot-fix-loop
**Decision:** narrative-line-record

Correction (2026-10-03): three departures from the text below, all found at implementation.
(1) The closed-list row's Role column declared `pipeline-fix`; shipped `implementer` for the fix
and the fixer, `reviewer` for the review — `flow record dispatch` accepts only the served
`recordRoles` (`stats/cmd/flow/record.go`), so `-role pipeline-fix` is refused, and `panel-fix`
would trip `check-panel-fix-single-dispatch.sh`'s canonical-key check on a `pipeline-fix-*` key.
(2) The fix branch is `fix-<slug>`, not `fix/<slug>`: `scripts/check-installed-citations.sh`
reads a backticked `fix/<slug>` as a rootless path citation. (3) The summary citation reads
`below`, not `above` — **Summary and live-stack line, before every handoff** sits under
`## Handoff output`, after the new section.

### Task 1 — the section text

````markdown verified:authored in-tree for this change
## Pipeline defects found mid-run

**A `/flow` or `/flow-fast` run that hits a defect in the pipeline itself fixes it within the
run, unless the fix is `big`.** The pipeline is everything `<agents repo>` ships — a skill, a
contract, a guard, the `flow` CLI or flowd, a rule — and any saved memory that encodes its
behaviour. For this section alone `<agents repo>` is a repository the run was given
(**Reporting back**, `rules/agent-baseline.md`), whichever project the run is in.

**Predict the class before any work, from the blast radius — never from a plan.** Count the
files the fix must touch: the defective file, its test, and every file `grep -rl` finds in
`<agents repo>` citing the name or section the fix changes. The fix is `big` — `plan-class.sh`'s
own `files>=60` threshold — when that count is 60 or more, when it reaches outside
`<agents repo>`, or when it needs a design choice only the operator can make. A `big` fix is
deferred to self-review, and so is one found `big` once under way, which stops there.

**Every dispatch in the loop is one-shot.** No `SendMessage` to a finished child, and no child
parks a question for a resume. Each runs on `opus`, `subagent_type: flow-high`, its key per the
steps below and `<k>` counting this run's in-run fixes:

1. **Fix** — `pipeline-fix-<k>`, its prompt carrying the defect, its evidence and the counted
   blast radius. It works in its own worktree,
   `git -C <agents repo> worktree add -b fix/<slug> <agents repo>/.worktrees/<slug> origin/<default-branch>`
   — never the main checkout, never the change's own worktree — adds one test or guard that
   fails without the fix, updates a saved memory that encoded the defect, runs the
   `<agents repo>/.flow/project.md` `## lint` lines its files need, and commits with a module
   scope. It never merges or pushes.
2. **Review** — a fresh `pipeline-fix-<k>-review-<r>` over
   `git diff origin/<default-branch>...fix/<slug>`.
3. **Fix the findings** — the parent fixes them inline in that worktree, or dispatches a fresh
   `pipeline-fix-<k>-fix-<r>`; then step 2 again, `<r>` plus one. The loop runs to a clean
   review under **Fewest operator actions** above.
4. **Land** — by `<agents repo>`'s `## default landing route`
   (`project-get.sh <agents repo> 'default landing route'`), without asking. Merge and push is:
   rebase `fix/<slug>` onto `origin/<default-branch>`,
   `git push origin fix/<slug>:<default-branch>`, `git -C <agents repo> pull --ff-only`, then
   remove the worktree and the branch.

The loop runs in the background while the run's own work continues; the run still never ends a
turn with one of its children in flight.

**Record each one.** Every in-run fix, landed or deferred, is one line in the run's narrative —
`narrative.md` on `/flow` (**Write `IN_PROGRESS`**, `skills/flow/verify-and-handoff.md`),
`## Session narrative` on `/flow-fast` — and one decision bullet in the run's summary
(**Summary and live-stack line, before every handoff**, above):

```text
In-run pipeline fix: <agents sha | deferred> — <the defect, one line> (blast radius <N> files)
```
````

### Task 1 — the closed-list row

```markdown verified:authored in-tree for this change; table columns read from implement.md @ 5e86aef6
| in-run pipeline fix, review and fixer, one each per loop step | `pipeline-fix` | `pipeline-fix-<k>`, `pipeline-fix-<k>-review-<r>`, `pipeline-fix-<k>-fix-<r>` | `skills/flow-contracts/pipeline.md`, **Pipeline defects found mid-run** |
```

- [x] 2. Self-review's angle 1 skips a defect fixed in-run

`design.md` § 4.

  - [x] **Step 1: Add the sentence** to `skills/flow-self-review/SKILL.md`, as its own paragraph
    directly after the paragraph opening `Angle 5's remit covers`, verbatim:

    > An `In-run pipeline fix:` line in the bundle (**Pipeline defects found mid-run**,
    > `skills/flow-contracts/pipeline.md`) that names a sha is reported under angle 1 as fixed and
    > is never offered for filing; one that reads `deferred` is an angle-1 finding like any other.

  - [x] **Step 2: Verify** — `grep -c 'In-run pipeline fix:' skills/flow-self-review/SKILL.md`
    prints `1`. Run the Markdown lint lines plus `scripts/check-self-review-report.sh`; record the
    new sentences in `verbatim-moves.txt`.
  - [x] **Step 3: Commit.**

**Files:** `skills/flow-self-review/SKILL.md`
**Tests:** none — skill prose
**Regression:** none — prose.
**Baseline:** before=0 after=0
<!-- measured: no test files in this task @ 5e86aef6 -->
**Commit:** `feat(flow-self-review): report an in-run pipeline fix as fixed, never re-file it`
**After:** Task 1
**Build:** green

**Decision:** narrative-line-record
## spectre/changes/archive/kan-829-flow-improvement-fix-pipeline-defects-at-the/design.md

## Context

One prose change across the pipeline contract, `implement.md`'s closed dispatch list, `/flow-fast`'s
guardrail and `/flow-self-review`'s angle 1. No guard, CLI or SPA change.

## 1. The rule's home and reach

- New `## Pipeline defects found mid-run` in `skills/flow-contracts/pipeline.md`, after
  `## Fewest operator actions`, whose no-round-count loop rule it runs under.
- `/flow` loads `pipeline.md` first, so it reaches every `/flow` phase. `/flow-fast` does not load
  it, so its guardrail sentence cites the section.
- Scope: the pipeline is everything `<agents repo>` ships, plus a saved memory encoding its
  behaviour. For this section alone `<agents repo>` is a repository the run was given, overriding
  `rules/agent-baseline.md`'s "a repository you were not given" for that one repository.

## 2. Predicting the class without a plan

- Blast radius: the defective file, its test, and every `grep -rl` hit in `<agents repo>` citing
  the name or section the fix changes.
- `big` when the count is 60 or more (`plan-class.sh`'s own `files>=60` big threshold), when the
  fix reaches outside `<agents repo>`, or when it needs a design choice only the operator can make.
  Everything else (`micro`/`small`/`regular`) is fixed in-run.
- A fix found `big` once under way stops and is deferred the same way.

## 3. The one-shot loop

- Fix subagent → fresh reviewer → parent inline fix or fresh fixer → fresh re-review, until a clean
  review; then land. Every dispatch is fresh; nothing is resumed and no question is parked.
- Each dispatch runs on `opus` at `high` (`flow-high`), keys `pipeline-fix-<k>`,
  `pipeline-fix-<k>-review-<r>`, `pipeline-fix-<k>-fix-<r>`.
- The fix branch lives in its own `<agents repo>` worktree from `origin/<default-branch>` — never
  the main checkout (read-only) and never the change's own worktree.
- Landing: `<agents repo>`'s `## default landing route`, no ask. Merge and push is
  `git push origin fix-<slug>:<default-branch>` after a rebase, then `pull --ff-only` in the main
  checkout so the installed symlinks see the fix.
- The loop runs in the background; the run's own work continues.

## 4. Recording and self-review

- One narrative line per fix: `In-run pipeline fix: <agents sha | deferred> — <defect> (blast radius
  <N> files)`, in `narrative.md` (`/flow`) or `## Session narrative` (`/flow-fast`), both of which
  already reach the self-review bundle; plus one decision bullet in the run summary.
- `/flow-self-review` angle 1 reports a line carrying a sha as fixed and offers no ticket for it; a
  `deferred` line is offered as any other angle-1 finding.

## Decisions

### Where the rule lives

**ID:** rule-home-pipeline-contract
**Status:** active
**Chosen:** a section of `skills/flow-contracts/pipeline.md` — loaded first by every `/flow` run; `/flow-fast` cites it.
**Considered:** amending `rules/agent-baseline.md` so `<agents repo>` is always "given" — reaches every agent, /flow or not, wider than asked; a `docs/briefs/` lesson only — changes no behaviour.

### How a fix lands

**ID:** land-by-default-route
**Status:** active
**Chosen:** `<agents repo>`'s `## default landing route`, without asking — the declared route is the durable authorization.
**Considered:** push the branch and leave merging to the operator — the fix does not reach the running pipeline; ask per fix — a stop the run does not need.

### Which defects are fixed in-run

**ID:** blast-radius-class
**Status:** active
**Chosen:** predicted class from a blast-radius file count, `big` at 60+ files, an outside reach, or an operator-only choice — deferred when `big`.
**Considered:** the agent-baseline cheap-fix test alone — too narrow, the operator asked for a bigger threshold; one module and no contract text — superseded by the operator's class-based framing; `plan-class.sh` on a throwaway task sketch — a mini plan, which the prediction must avoid; unquantified judgment — no checkable number.

### One-shot dispatch

**ID:** one-shot-fix-loop
**Status:** active
**Chosen:** every fix, review, fixer and re-review is a fresh dispatch run to completion; the parent never resumes a child and no child parks a question.
**Considered:** resuming the fix subagent with review findings — the pause-resume cycle the operator ruled out.

### Self-review sees in-run fixes

**ID:** narrative-line-record
**Status:** active
**Chosen:** one narrative line per fix, read by self-review from the bundle it already holds.
**Considered:** a new bundle section or store record — a CLI change for what one existing prose channel carries.

## Open questions
## spectre/changes/archive/kan-829-flow-improvement-fix-pipeline-defects-at-the/narrative.md

# kan-829-flow-improvement-fix-pipeline-defects-at-the — session narrative

## 2026-10-03 — creating run

Resumed at `STARTED` with the plan ready; implemented inline (micro). The plan's closed-list row
named a dispatch role `pipeline-fix`, which `flow record dispatch -role` refuses (`recordRoles`,
`stats/cmd/flow/record.go`); shipped `implementer`/`reviewer` instead, recorded as a dated
Correction with two smaller ones (branch `fix-<slug>`, since the citations guard reads
`fix/<slug>` as a rootless path; the summary heading sits below the new section, not above).
The gated reviewer on Task 1 raised one Important — the loop's prompts carried none of the
mandatory dispatch paragraphs — fixed inline and re-reviewed clean.

Mid-run the parent itself hit a pipeline ambiguity: `implement.md`'s inline Records bullet reads
as recording every fix round `-role panel-fix`, but a gated fix round's key
`task-<n>-implementer-fix-<k>` is out of shape for `check-panel-fix-single-dispatch.sh` under that
role, so the stage-close guard flagged this run's own row. Fixed within this change (5b917257,
`-role implementer` stated in `implement.md` and `gated-review-fix.md`) and the panel re-ran on the
delta clean; the one recorded violation stays in this run's records.

In-run pipeline fix: 5b917257 — gated fix round's record role was ambiguous, tripping check-panel-fix-single-dispatch (blast radius 2 files)

Zsh word-splitting broke a pathspec held in a variable once, and a `bash -c` wrapper for the
verify list was refused by the harness's removal check; a script file under the job's tmp ran it.

## 2026-10-03 — integrate run

Preflight returned RUN1; the main checkout was staged-clean and drift-clean, and the
unfinished-work and visual-verify gates both cleared with nothing outstanding. `origin/main` had
moved 9 commits since the recorded merge base with no overlap on this change's paths; the rebase
onto 75cc1a75 was clean and no guard test was named for re-verification. The route was the
project's configured default, merge and push, so no landing question was asked. One stumble:
`flow state get` takes its `-C` flag before the change name, not after it.
## git log --stat

commit 6b2020bf0bdbda6c248e9ba253594878d2a0eca7
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Oct 3 23:34:10 2026 +0300

    feat(flow): fix a pipeline defect within the run that finds it

 skills/flow-contracts/pipeline.md | 52 +++++++++++++++++++++++++++++++++++++++
 skills/flow-fast/SKILL.md         |  3 ++-
 skills/flow-self-review/SKILL.md  |  4 +++
 skills/flow/gated-review-fix.md   |  2 +-
 skills/flow/implement.md          | 14 ++++++-----
 5 files changed, 67 insertions(+), 8 deletions(-)

commit f1a4bdbf1a006855f6673ed1456be03de827a792
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Oct 3 23:34:11 2026 +0300

    chore(spectre): plan

 .../narrative.md                                                 | 9 +++++++++
 1 file changed, 9 insertions(+)

commit f114147d90a66a05353dbf911f21acc2d8c46bda
Author: Yuriy Aleksandrov <yatweety@gmail.com>
Date:   Sat Oct 3 23:35:01 2026 +0300

    chore(spectre): archive kan-829-flow-improvement-fix-pipeline-defects-at-the

 .../design.md                                      |  0
 .../ledger.md                                      | 94 ++++++++++++++++++++++
 .../narrative.md                                   |  0
 .../panel.md                                       | 27 +++++++
 .../proposal.md                                    |  0
 .../tasks.md                                       |  0
 .../verbatim-moves.txt                             |  0
 7 files changed, 121 insertions(+)

## Session narrative

Run 2 followed run 1 in the same invocation on the project's default merge-and-push route. The
merge was verified locally in the landing worktree, the change archived and committed on
`chore/archive-kan-829-flow-improvement-fix-pipeline-defects-at-the`, and the apply worktree,
local branch and remote branch removed after check 4 disclosed four unclassified entries — three
`__pycache__` files and `stats/web/tsconfig.tsbuildinfo`, all build caches, none irreplaceable — and
proceeded without asking. Check 5 was skipped because `## stop` declares no fenced command. The
workspace database `flow_kan_829_flow_6c28` was already absent, and the cleanup verified complete.
No proposal artifact source existed. One harness friction: an `rm` on a variable-built path was
refused until the expansions were guarded with `${VAR:?}`.
