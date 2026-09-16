# flow-plan-inits-spectre — design

**Change:** `flow-plan-inits-spectre`
**Jira:** none — run unlinked, by the operator's choice
**Date:** 2026-09-16

## Context

`/flow`'s resume-at-`STARTED` logic (**Resuming at `STARTED`**, `skills/flow/brainstorm.md`)
already reads where a creating run stopped: worktree present, plan present, decision recorded,
then straight to implement. Everything `/flow-plan` needs to produce is therefore already
defined by cited sections of `/flow` — A, kickoff, C, D, Decide, the plan review gate. The change
makes `/flow-plan` run those sections at capture instead of writing a note for `/flow` to
re-render, and deletes the note mechanism.

## 1. `/flow-plan` after capture

- The stance, the fixed section structure, the no-open-items rule and the gate are unchanged.
- **Read vantage:** the main checkout, read-only. Research writes nothing until capture, so the
  `_plan-<stem>` worktree and the `plan-<stem>` branch go, and with them the direct push onto
  `<default-branch>` and its git-boundaries exception.
- **Capture with no existing change**, after the convergence check closes: run, in order and
  citing rather than restating, `/flow`'s **A. Resolve the change and write `STARTED`** and the
  kickoff steps (`skills/flow/brainstorm.md`), **C. Create the change and its artifacts**,
  **D. Basic Workflow #3 — Writing plans**, **Decide** and **Plan review gate**
  (`skills/flow/brainstorm-planner.md`). The note's thread sections and its step-by-step
  breakdown become `design.md`'s body, `## Decisions` is filled from the session's answers,
  `## Open questions` is present and empty, `proposal.md` carries `## Why` and `## What changes`.
  The decision is recorded with `flow record decision -change <name> -session-token
  fp-<literal-token>`. The gate keeps its **Push artifacts?** wording.
- **On Yes:** commit `<project>/spectre/changes/<name>/` on `spectre/<name>` with subject
  `chore(spectre): plan` and push. Integrate's reshape (`reset --soft` to the merge base, then
  two commits) re-splits everything since the merge base, so an early planning commit changes
  nothing downstream.
- **Marks:** `plan.session` alone, begun against the Jira key as today and closed `started`;
  the `flow.*` keys the session executes stay `/flow`'s. The store attaches the open plan
  session to the change when `flow state set` creates it (`stats/internal/store/changes.go`).
- **Bare topic:** the Jira Task is still created first so the key names the change; with Jira
  skipped, the name is the slug alone, as `/flow` names an unlinked change.
- **Existing change in scope:** the `design.md` addition path is unchanged.
- **Nothing captured:** nothing is created; init happens only after the gate answers Yes.

## 2. `/flow`

- **Removed:** the research-seed line and the seeded startup block (`skills/flow/SKILL.md`); the
  seed section and **One shared mechanism, not two copies**, C's delete step, D's seeded-plan
  paragraph, Decide's seeded-decision paragraph and the gate's fully-seeded sentence
  (`skills/flow/brainstorm-planner.md`); the fully-seeded bypass (`skills/flow/brainstorm.md`);
  the second bounded exception and the `docs/research/` half of the planning-path rule
  (`skills/flow-contracts/pipeline.md`); the `/flow-plan` rows, the research-branch exception and
  the `docs/research/` pathspec (`skills/flow-contracts/git-boundaries.md`); the adopted-note
  clause (`skills/flow-contracts/finish-contract-run1.md`); the `docs/research/` mentions in
  `skills/flow/implement.md`, `skills/flow/verify-and-handoff.md`, `skills/flow-fast/SKILL.md`;
  the seed rationale entries (`skills/flow/SKILL-rationale.md`).
- **Name lookup by key:** section A, with a linked issue, first checks the candidate set
  (**Change name resolution**, `skills/flow-contracts/pipeline.md`) for exactly one name
  starting with `<lowercased-key>-` and resumes it; only with no match is a slug derived.
- **Kickoff step 3:** when `origin/spectre/<name>` exists after the fetch, `git worktree add
  <project>/.worktrees/<name> spectre/<name>` tracking it, instead of `-b … origin/<default-branch>`;
  step 5's push is then a plain `git push -u origin spectre/<name>` that is a no-op.

## 3. Scripts, docs, deletions

- `scripts/commit-split.sh` resets and excludes `spectre/changes/` alone; `test-commit-split.sh`
  moves case 2's file under `spectre/changes/` and adds a case proving a `docs/research/` file
  lands in the implementation commit.
- `scripts/recover-guard-incident.sh` defaults to `spectre/changes` alone; its test's default
  fixture and case 10's assertion follow.
- `scripts/gather-self-review-context.sh` keeps `docs/research/` in its planning-only-parent
  exclusion as a retired path beside `docs/superpowers/`, since it classifies commits already in
  history; only its comments change to say so.
- `README.md`, `skills/README.md`, `CLAUDE.md`, `AGENTS.md`, `commands/flow-plan.md`,
  `commands-claude/flow-plan.md` describe `/flow-plan` as ending at `STARTED`.
- `docs/research/` is deleted: `kan-492`, `kan-512`, `kan-516` with their sibling directories,
  `flow-speedup.md`, `flow-gymie-implementation-speedup.md`.
- No Go change: `plan.session` is unchanged and the store accepts any outcome string.
- Historical files under `docs/superpowers/` and the `kan-516` worktree are not touched.

## Decisions

### `/flow-plan` ends at `STARTED`, not at a note

**ID:** `flow-plan-ends-at-started`
**Status:** active
**Chosen:** the session runs `/flow`'s cited planning sections at capture and leaves a `STARTED`
change whose artifacts, plan and decision are its own; `/flow <KEY>` is a plain resume.
**Considered:** keeping the note and only tightening the adoption — rejected because the
re-rendering and the seed machinery are the cost, not their tuning; landing the planning
artifacts on the default branch — rejected because an unstarted change under `spectre/changes/`
on main is listed by `spectre list` and `/flow-status` as open.

### Init happens after the gate, reading from the main checkout

**ID:** `init-after-gate-main-checkout-read`
**Status:** superseded by init-after-capture-offer
**Chosen:** the session reads the main checkout read-only and creates nothing until the gate
answers Yes, so an abandoned session leaves no change and no worktree.
**Considered:** writing `STARTED` and creating the worktree at session start — rejected because
an abandoned session would leave a `STARTED` change to clean up; keeping the research worktree
as a read vantage — rejected as a worktree whose only purpose is reading.

### Only `plan.session` is marked

**ID:** `plan-session-only-mark`
**Status:** active
**Chosen:** `/flow-plan` marks `plan.session` alone, closed `started`; the `flow.*` keys stay
`/flow`'s in the Level 1 table.
**Considered:** marking `flow.kickoff`, `flow.create-artifacts`, `flow.writing-plans` and
`flow.decide` under `/flow-plan` — rejected as a table change and a second command per row for
stages a plan session already spans as one run.

### Existing notes are deleted, not migrated

**ID:** `delete-existing-notes`
**Status:** active
**Chosen:** the five notes and three sibling directories are deleted in this change.
**Considered:** leaving them as inert docs — rejected as files no command reads; converting the
keyed notes into `STARTED` changes here — rejected as three more branches in one change, and
`kan-516` already has a worktree.

### The self-review context guard keeps `docs/research/` as a retired path

**ID:** `gather-keeps-retired-research-path`
**Status:** active
**Chosen:** `gather-self-review-context.sh` still excludes `docs/research/` when deciding
whether a plan commit's parent touched implementation, beside the already-retired
`docs/superpowers/`.
**Considered:** dropping it — rejected because archived changes whose plan commits deleted an
adopted note are already in history, and the guard classifies history.

### Kickoff adopts an existing remote branch

**ID:** `kickoff-adopts-remote-branch`
**Status:** active
**Chosen:** step 3 checks out `origin/spectre/<name>` when it exists; step 5's push stays.
**Considered:** leaving it to a separate change — rejected because `/flow-plan` now hands off a
pushed branch, so the latent resume defect becomes the ordinary path.

### Init happens after the capture offer is accepted, and the gate gates the commit

**ID:** `init-after-capture-offer`
**Status:** active
**Chosen:** the `STARTED` write, the worktree and the artifacts follow the operator's **Yes** to
the capture offer, asked as one **AskUserQuestion**; the plan review gate then gates only the
commit and push, and its **No** revises the plan. Section 1's "Nothing captured" bullet reads
"after the capture offer is accepted" in place of "after the gate answers Yes". The main-checkout
read vantage is unchanged.
**Considered:** keeping "nothing before the gate" — rejected because the gate needs the plan and
the decision, which need the artifacts, which need the worktree (panel F2); writing the
artifacts somewhere other than the change worktree first — rejected as the scratch-directory
shape kan-488 already rejected.

## Open questions
