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

Follow **rules/myflow-manual-review.mdc** — section **Pipeline stages** — and **State file** (`skills/myflow-contracts/state-file.md`) and **State self-heal** (`skills/myflow-contracts/state-self-heal.md`).

## Workflow

### 1. List open changes

```bash
openspec list --json
```

With a `<name>` argument, restrict to that change and include the detail view (step 4). With no argument, report every non-archived change.

Zero open changes → say so and suggest `/myflow-start`. Stop.

### 2. Resolve each change's stage

For each change, resolve its user-scoped state file path per **State file** in `skills/myflow-contracts/state-file.md`, then read it:

```bash
MAIN_CHECKOUT="$(cd "$(dirname "$(git rev-parse --git-common-dir)")" && pwd)"
PROJECT_KEY="$(basename "$MAIN_CHECKOUT")-$(printf '%s' "$MAIN_CHECKOUT" | shasum | cut -c1-8)"
STATE_FILE="/Users/tweety53/Agents/myflow/state/$PROJECT_KEY/<name>.json"

jq -r '.stage, .gates.reviewed, .gates.tested, .gates.prOpened, .gates.prMerged, .worktree, .branch, .artifactUrl, .jiraIssue, .originStage, .updatedAt, .updatedBy' \
  "$STATE_FILE" 2>/dev/null
```

Then validate against artifacts per **State self-heal** (`skills/myflow-contracts/state-self-heal.md`) — including its "read artifacts from the apply worktree when one exists" rule. Resolve the worktree **first**, then root every subsequent artifact check there:

1. worktree resolution — find the apply worktree for branch `openspec/<name>` in `git worktree list`; if found, treat its path as the artifact root for the rest of this step, otherwise use the main checkout
2. `tasks.md` (at the resolved root) — count `- [x]` vs `- [ ]` items
3. `docs/manual-test/<name>.md` (at the resolved root) — exists? contains `**Manual test status:** SKIPPED`? any unchecked boxes?
4. PR — `gh pr list --head <branch> --state all --json number,state,url`, **only when `gh` is installed and the remote is a GitHub host**. Otherwise (Bitbucket, no `gh`, no network, branch never pushed) PR state is **unknown** — report it as unknown and treat it as inconclusive, never as "no PR exists".

If the file is missing, unparseable, or contradicted, infer the stage from artifacts, **rewrite the state file**, and mark the row `⚠`.

**Self-heal is stage-only and monotonic** (per **State file** in `skills/myflow-contracts/state-file.md` → gate monotonicity):

- Infer and correct **`stage` only**. Preserve every existing gate value exactly as read.
- Fill `null` gates **conservatively** — `false`, never `true`.
- **Never** infer `gates.tested: true`; only `/myflow-review` writes that. Never demote `true` or `"skipped"`.
- **Never write a stage earlier than the one recorded.** A check that cannot be performed is not a contradiction: an undeterminable PR state leaves `awaiting-pr-review` exactly as recorded.

### 3. Render the table

Sort by stage order per **Pipeline stages** in the rule file (`awaiting-proposal-review`,
`proposal-done`, `awaiting-do-review`, `do-review-started`, `do-done`, `awaiting-fix-review`,
`fix-review-started`, `awaiting-manual-test`, `manual-test-done`, `awaiting-pr-review`,
`review-done`, `finished`), then by `updatedAt` descending. Omit `finished` changes — they are
archived.

```
## myflow status

| Change | Jira | Stage | Gates | Next | Worktree / branch | Updated |
|--------|------|-------|-------|------|------|-----------|
| kan-31-user-workout-core | KAN-31 | awaiting-manual-test | review ✓ · test ☐ · PR — | run the guide, then `/myflow-manual-test-done` | `.worktrees/openspec-kan-31-user-workout-core` @ `openspec/kan-31-user-workout-core` | 22h ago (/myflow-manual-test) |
| active-workout-session-editing | — | proposal-done | — | `/myflow-do` | none | 19h ago (/myflow-start-done) |

⚠ = state file was stale and has been corrected from artifacts.
```

Gate glyphs: `✓` passed · `☐` pending · `⊘` skipped · `—` not yet reached.

The **Jira** column shows `jiraIssue` verbatim, or `—` when the change has no linked issue. This
is a **read-only** report: never call Jira, never transition an issue, never infer a key from the
change name.

Surface, when present:
- `artifactUrl` — link to the published proposal artifact
- `originStage` — the stage a fix in flight (`awaiting-fix-review` / `fix-review-started`) will return to

Next-command mapping:

| Stage | Next |
|-------|------|
| `awaiting-proposal-review` | read the artifact, then `/myflow-start-done` (or `/myflow-start-fix`) |
| `proposal-done` | `/myflow-do <name>` |
| `awaiting-do-review` | `/myflow-do-manual-review`, then `/myflow-do-done` (or `/myflow-do-fix`) |
| `do-review-started` | `/myflow-do-done` (or `/myflow-do-fix`) |
| `do-done` | `/myflow-manual-test <name>` (or `/myflow-do-fix`) |
| `awaiting-fix-review` | `/myflow-do-fix-manual-review`, then `/myflow-do-fix-done` |
| `fix-review-started` | `/myflow-do-fix-done` |
| `awaiting-manual-test` | run the guide, then `/myflow-manual-test-done` (or `/myflow-do-fix`) |
| `manual-test-done` | `/myflow-review <name>` (or `/myflow-do-fix`) |
| `awaiting-pr-review` | review + merge the PR, then `/myflow-review-done` (or `/myflow-do-fix`) |

`/myflow-do-fix` is offered only at its six accepted origins (`awaiting-do-review`,
`do-review-started`, `do-done`, `awaiting-manual-test`, `manual-test-done`, `awaiting-pr-review`)
— never at `awaiting-fix-review` or `fix-review-started`, where the next step is
`/myflow-do-fix-done`.
| `review-done` | `/myflow-finish <name>` |
| `finished` | — |

### 4. Detail view (only when `<name>` was given)

Add below the table:

- Linked Jira issue key, or "none linked"
- Task progress (`N/M` checked from `tasks.md`)
- Nested `<name>-fix-N` sub-changes, if any
- Manual test guide path + checked/total box count, or `SKIPPED`
- PR number, state, and URL when one exists
- Every `⚠` correction made, with the reason

## Guardrails

- **Never** commit, stage, push, merge, or archive.
- **Never** advance a stage — only correct a stale cache to match artifacts.
- **Never** rewind a stage, lower a gate value, or infer `gates.tested: true` — self-heal corrects `stage` only and fills `null` gates as `false`.
- **Never** create a worktree or branch.
- Report `gh` being unavailable, or a non-GitHub forge, as "PR state unknown" rather than guessing — and never rewind a stage on that unknown.
- Never guess a stage when artifacts are ambiguous — show `?` and say which check was inconclusive.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Status of all open changes | `/myflow-status` |
| Detail for one change | `/myflow-status <name>` |
| Pipeline reference | `/myflow-info` |
