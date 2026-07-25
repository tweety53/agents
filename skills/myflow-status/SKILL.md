---
name: myflow-status
description: Show every open myflow change with its pipeline stage, gate flags, next command, worktree, and last update. Read-only. Use for /myflow-status.
allowed-tools: Bash(openspec:*), Bash(git:*), Bash(jq:*), Bash(gh:*)
license: MIT
compatibility: Requires openspec CLI and jq.
metadata:
  author: gymie
  version: "1.0"
---

Report the pipeline stage of every open (non-archived) OpenSpec change. **Read-only** — never commits, never runs git write operations, never advances a stage. The one exception is state self-heal (see below), which corrects a stale cache to match the artifacts already on disk.

**Announce at start:** "Using myflow-status."

Follow **rules/myflow-manual-review.mdc** — sections **State file**, **State self-heal**, **Pipeline stages**.

## Workflow

### 1. List open changes

```bash
openspec list --json
```

With a `<name>` argument, restrict to that change and include the detail view (step 4). With no argument, report every non-archived change.

Zero open changes → say so and suggest `/myflow-start`. Stop.

### 2. Resolve each change's stage

For each change, read the state file:

```bash
jq -r '.stage, .gates.reviewed, .gates.tested, .gates.prOpened, .gates.prMerged, .worktree, .branch, .updatedAt, .updatedBy' \
  openspec/changes/<name>/.myflow-state.json 2>/dev/null
```

Then validate against artifacts per **State self-heal**. Check, in order:

1. `tasks.md` — count `- [x]` vs `- [ ]` items
2. apply worktree — is `.worktree` present in `git worktree list`?
3. `docs/manual-test/<name>.md` — exists? contains `**Manual test status:** SKIPPED`? any unchecked boxes?
4. PR — `gh pr list --head <branch> --state all --json number,state,url` (skip if `gh` is unavailable or the branch was never pushed)

If the file is missing, unparseable, or contradicted, infer the stage from artifacts, **rewrite the state file**, and mark the row `⚠`.

### 3. Render the table

Sort by stage order (`start`, `awaiting-review`, `awaiting-test`, `awaiting-pr-review`), then by `updatedAt` descending. Omit `finished` changes — they are archived.

```
## myflow status

| Change | Stage | Gates | Next | Worktree / branch | Updated |
|--------|-------|-------|------|------|-----------|
| user-workout-core | awaiting-test | review ✓ · test ☐ · PR — | `/myflow-code-review` | `.worktrees/user-workout-core` @ `openspec/user-workout-core` | 22h ago (/myflow-manual-test) |
| active-workout-session-editing | start | — | `/myflow-do` | none | 19h ago (/myflow-start) |

⚠ = state file was stale and has been corrected from artifacts.
```

Gate glyphs: `✓` passed · `☐` pending · `⊘` skipped · `—` not yet reached.

Next-command mapping:

| Stage | Next |
|-------|------|
| `start` | `/myflow-do <name>` |
| `awaiting-review` | Gate B — review staged diff, then `/myflow-manual-test <name>` |
| `awaiting-test` | Gate C — run the guide, then `/myflow-code-review <name>` |
| `awaiting-pr-review` | Gate D — review + **merge the PR yourself**, then `/myflow-finish <name>` |

### 4. Detail view (only when `<name>` was given)

Add below the table:

- Task progress (`N/M` checked from `tasks.md`)
- Nested `<name>-fix-N` sub-changes, if any
- Manual test guide path + checked/total box count, or `SKIPPED`
- PR number, state, and URL when one exists
- Every `⚠` correction made, with the reason

## Guardrails

- **Never** commit, stage, push, merge, or archive.
- **Never** advance a stage — only correct a stale cache to match artifacts.
- **Never** create a worktree or branch.
- Report `gh` being unavailable as "PR state unknown" rather than guessing.
- Never guess a stage when artifacts are ambiguous — show `?` and say which check was inconclusive.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Status of all open changes | `/myflow-status` |
| Detail for one change | `/myflow-status <name>` |
| Pipeline reference | `/myflow-info` |
