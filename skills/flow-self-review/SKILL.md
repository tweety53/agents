---
name: flow-self-review
description: Run the self-review pass a `/flow` run deferred, inline on this session's model, from the saved context bundle; file, rate, write the report, delete the bundle. Standalone, not a pipeline stage. Use for /flow-self-review.
allowed-tools: Bash(git:*), Bash(scripts/check-self-review-report.sh:*)
license: MIT
compatibility: Requires the change's default branch to be checked out and a saved context bundle at docs/self-review/<name>-context.md.
metadata:
  author: gymie
  version: "1.0"
---

Run the self-review reasoning pass a `/flow` run deferred (`## self review: defer`, per
`skills/flow-contracts/project-configuration.md`), from the context bundle run 2 step 9 saved —
`skills/flow-contracts/finish-contract-run2.md` step 9 is canonical for that bundle's shape and
the five-angle table below. **The pass runs inline, in this session, on whatever model it is
already on** — no subagent, no dispatch, no `Model:` handshake. Picking a stronger model than the
same-run pass would have used is done by picking the model this session runs on (`/model`) before
invoking this command, not by anything this skill itself resolves.

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

### 2. Read the bundle and run the five angles

Read `<bundle>` in full. Run the five-angle table over it, inline, every angle explicit —
present-but-empty for an angle with no findings, never omitted:

| # | Angle | Label |
|---|-------|-------|
| 1 | Problems encountered, and what pipeline change would avoid them | `flow-fix` |
| 2 | Token/time cost, and what would reduce it without quality loss | `flow-cost` |
| 3 | What went well, and how to reproduce it | `flow-improvement` |
| 4 | What could be automated or moved to a script | `flow-automation` |
| 5 | What could move to the Go app or its persistent storage | `flow-stats-app` |

**One combined pass** — never five separate reads. A deferred pass covers what the bundle holds
and nothing a same-run session could still remember beyond it — **Run 2 — the branch is merged** (`skills/flow-contracts/finish-contract-run2.md`), step 9 — say so in the report, never imply
parity with a same-run pass.

### 3. Explain, then ask

Every finding is explained in the message body first, before any prompt fires — what was
observed, what breaks, and what the fix would be. The filing ask and the rating are **one
`AskUserQuestion` call**, shaped exactly as run 2 step 9's own — **Run 2 — the branch is merged** (`skills/flow-contracts/finish-contract-run2.md`), step 9: up to three multi-select questions of
three findings each, every option prefixed with its angle's label, plus **None — file nothing**
as the default; the rating last, `5 — excellent` / `4 — good` / `3 — fine` / `2 — rough`, a `1`
typed through the tool's free-text "Other". More than nine findings roll the overflow into one
further call of the same shape, without the rating.

### 4. File chosen findings

File each chosen finding as a Jira issue per **Labels on issues the pipeline creates** (`skills/flow-contracts/jira-integration.md`), carrying its angle's label on top of that set.
`## jira` absent or `none` in `<project>/.flow/project.md`, or no Atlassian tooling available in
this session: print `⚠ Jira: skipped — <reason>` and record that finding `declined` instead.

### 5. Write the report, delete the bundle, land it

Write `<project>/docs/self-review/<name>-self-review.md` in the shape
`<project>/scripts/check-self-review-report.sh` checks — one section per angle, all five present; each
finding one line naming its angle's label, the finding, and its disposition (`filed: <KEY>` or
`declined`); an angle with no findings carrying the explicit none-marker — plus one line under the
title:

```
**Deferred:** reasoning pass run on <model named in this session's own system prompt> from <bundle path>
```

Delete `<bundle>`. Run `<project>/scripts/check-self-review-report.sh` when the project declares
it in `<project>/.flow/project.md`'s `## lint` section; fix any violation before committing.
Commit both paths in one commit, re-asserting the branch immediately before committing — the same
shape `skills/flow/archive.md` step 9's own commit shells use, since the round-trip through the
five-angle pass and the filing-and-rating prompt above is long enough that the branch is worth
re-checking rather than trusted from step 1 alone:

```bash
[ "$(git branch --show-current)" = "<default-branch>" ] \
  && git add -- "docs/self-review/<name>-self-review.md" \
  && git rm -- "docs/self-review/<name>-context.md" \
  && git commit -m "docs(self-review): <name> self-review report" \
  && git pull --rebase origin <default-branch> \
  && git push origin <default-branch>
```

A branch mismatch stops here — nothing is committed, pulled, or pushed. A rejected push leaves the
commit local and this run names it — never retried around.

### 6. Report

End naming the report path, the rating, and the Jira keys filed (or `none`).

## Guardrails

- **Never** dispatch a subagent for the reasoning pass — it runs inline, in this session, always.
- **Never** touch a per-change state file or any `flow state`/`flow stage` call.
- **Never** push directly to a protected branch under any name but the project's own default
  branch, and never force-push.
- **No flags** — the only input is the change name.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Run a deferred self-review pass for `<name>` | `/flow-self-review <name>` |
