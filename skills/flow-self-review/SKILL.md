---
name: flow-self-review
description: Run a change's self-review pass, inline on this session's model, from the context bundle `/flow` or `/flow-fast` saved; file, rate, write the report, delete the bundle. Standalone, not a pipeline stage. Use for /flow-self-review.
allowed-tools: Bash(git:*), Bash(scripts/check-self-review-report.sh:*), Bash(land-self-review-report.sh:*)
license: MIT
compatibility: Requires the change's default branch to be checked out and a saved context bundle at docs/self-review/<name>-context.md.
---

Run a change's self-review reasoning pass — the only one the pipeline has — from the context
bundle run 2 step 9 saved (`skills/flow-contracts/finish-contract-run2.md` step 9 is canonical
for that bundle's shape) or **5. Verify** (`skills/flow-fast/SKILL.md`) saved on the change
branch before landing. This file is canonical for the six angles, what may be filed, the
filing-and-rating prompt and the report. **The pass runs inline, in this session, on whatever
model it is already on** — no subagent, no dispatch. The model is picked by picking the model
this session runs on (`/model`) before invoking this command, not by anything this skill itself
resolves.

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

**A finding is filed only from the six angles, and only by the operator's choice.** A finding
about the pipeline itself is offered under its angle. A finding about the project's own product
code is offered only when it is Important or worse — something a user or the data would
suffer; a Minor one (naming, doc-comment drift, an unused parameter, a duplicated fixture, a
missing test over already-correct code) is left out of the prompt. The report carries no
section beyond the six angles and the rating. The filing prompt is never waived: a pass with
no operator to answer it files nothing and records every finding `declined`.

**One combined pass** — never six separate reads. The pass covers what the bundle holds and
nothing beyond it — say so in the report's `**Deferred:**` line.

### 3. Explain, then ask

Every finding is explained in the message body first, before any prompt fires — what was
observed, what breaks, and what the fix would be. A prompt's option text cannot carry that
explanation, so the prompt records the decision only: a filed issue is durable, and an
explanation arriving afterward describes something the operator did not agree to. The filing ask
and the rating are **one `AskUserQuestion` call**, shape per **Operator prompts**
(`skills/flow-contracts/operator-prompts.md`): up to three multi-select questions of
three findings each, every option prefixed with its angle's label, plus **None — file nothing**
as the default; the rating last, `5 — excellent` / `4 — good` / `3 — fine` / `2 — rough`, a `1`
typed through the tool's free-text "Other". More than nine findings roll the overflow into one
further call of the same shape, without the rating.

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

### 4. File chosen findings

File each chosen finding as a Jira issue per **Labels on issues the pipeline creates** (`skills/flow-contracts/jira-integration-finish.md`), carrying its angle's label on top of that set.
`## jira` absent or `none` in `<project>/.flow/project.md`, or no Atlassian tooling available in
this session: print `⚠ Jira: skipped — <reason>` and record that finding `declined` instead.

### 5. Write the report, delete the bundle, land it

Write `<project>/docs/self-review/<name>-self-review.md` in the shape
`<project>/scripts/check-self-review-report.sh` checks — one section per angle, all six present; each
finding one line naming its angle's label, the finding, and its disposition (`filed: <KEY>` or
`declined`); an angle with no findings carrying the explicit none-marker — plus one line under the
title:

```
**Deferred:** reasoning pass run on <model named in this session's own system prompt> from <bundle path>
```

Delete `<bundle>`. Run `<project>/scripts/check-self-review-report.sh` when the project declares
it in `<project>/.flow/project.md`'s `## lint` section; fix any violation before committing.
**Exit 2 — the guard cannot answer at all (a missing or unreadable target directory, an
unreadable report inside it, an internal coverage.sh call failing) — is not a violation**: report it and stop before the
commit, never commit a report the guard could not read.
Commit both paths in one commit through the landing chain's one script — the same invocation
`skills/flow/archive.md` step 9 lands the context bundle with, differing in the asserted branch,
the report path, the removed context-bundle path and the `--push` — since the round-trip through the six-angle
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

### 6. Report

End naming the report path, the rating, and the Jira keys filed (or `none`).

## Guardrails

- **Never** push directly to a protected branch under any name but the project's own default
  branch, and never force-push.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Run a deferred self-review pass for `<name>` | `/flow-self-review <name>` |
