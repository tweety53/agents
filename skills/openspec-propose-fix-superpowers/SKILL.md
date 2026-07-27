---
name: openspec-propose-fix-superpowers
description: Revise an OpenSpec change's proposal after the human has read it. Updates artifacts, republishes the proposal artifact to the same URL, stays at awaiting-proposal-review. Use for /myflow-start-fix.
allowed-tools: Bash(openspec:*), Bash(git:*), Bash(jq:*)
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: gymie
  version: "1.0"
---

Revise the proposal for a change that is sitting at the proposal-review gate. **No code is
written and no worktree is created** — implementation has not started.

**Announce at start:** "Using openspec-propose-fix-superpowers for change `<name>`."

Follow **rules/myflow-manual-review.mdc** — **Stage transitions**, **State file**, **Pipeline stages**.

## Stage gate

Requires stage **`awaiting-proposal-review`**. On mismatch, stop with the standard mismatch
handoff and AskUserQuestion override (default **No**). At `proposal-done` the proposal is already
accepted — recommend `/myflow-do`, or `/myflow-do-fix` if implementation has begun.

## Workflow

1. **Collect the feedback.** Ask what should change if the user has not already said.
2. **Revise the artifacts** — `proposal.md`, `design.md`, delta `specs/**`, `tasks.md` — so they
   stay coherent with each other. A change to scope must reach the specs and the task list, not
   just the prose.
3. **Re-open the design discussion only if the feedback reopens an architectural question.** If it
   does, put the reopened decision to the user, then update the `## Decisions` section using the
   **same ID-based supersede mechanism** `openspec-propose-superpowers` defines: locate the
   reopened entry **by its `**ID:**`, never by heading text** — headings may be reworded across
   rounds, but the ID assigned at creation is immutable and is the only reliable match key. Set
   that entry's `**Status:**` to `superseded by <new-id>`, then append a **new** entry with a
   fresh, immutable ID and `**Status:** active`. **Never** delete or rewrite the superseded entry
   in place — the history is the point. Each entry (old and new) carries both `**ID:**` and
   `**Status:**` lines, per the shape in `openspec-propose-superpowers/SKILL.md`.
4. **Republish the artifact to the same URL** by passing the same file path to the Artifact tool.
   `artifactUrl` must not change.
5. **Sync added scope to the Jira issue.** When this round adds scope the linked issue does not
   describe, append **one dated bullet** under `## Added during implementation`, per **Description
   sync** in **Jira integration** (`rules/myflow-manual-review.mdc`) — canonical, not restated
   here. Feedback that only sharpens wording adds no scope and writes nothing. Skip when
   `jiraIssue` is `null`; report the append (or one skipped-with-reason line) in the
   handoff. **Never transition the issue here** — the proposal gate is inside In Progress already.
6. **Write state:** stage stays `awaiting-proposal-review`; update only `updatedAt` and
   `updatedBy: "/myflow-start-fix"`. Carry every other field forward, gates included — and
   `artifactUrl`, `jiraIssue`, `fastPath`, `REVIEWED_TREE` and `MERGE_BASE` with them. Writes
   render the whole object, so an omitted field is erased permanently.

## Handoff

```
## Proposal Revised

**Change:** <name>
**Round:** <N>
**Artifact:** <artifactUrl> (same link, updated)
**Decisions:** <unchanged | 1 superseded, 1 added>
**Jira description:** appended 1 entry under `## Added during implementation` | unchanged (no added scope) | none linked

**Open in IntelliJ:**
open -na "IntelliJ IDEA" --args "<absolute main checkout path>"

**Next:** more changes → `/myflow-start-fix <name>` · looks right → `/myflow-start-done <name>`
```

## Guardrails

- **Never** write code, create a worktree, or create a branch.
- **Never** advance the stage — that is `/myflow-start-done`'s job.
- **Never** mint a new artifact URL; republish to the recorded one.
- **Never** transition the Jira issue, and never let a Jira call block the revision — one
  skipped-with-reason line and continue, per **Jira integration**.
- **Never** delete a superseded decision; mark it superseded.
