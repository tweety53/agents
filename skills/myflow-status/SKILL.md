---
name: myflow-status
description: Show every open myflow change with its pipeline state, PR, next command, worktree, and last update. Read-only. Use for /myflow-status.
allowed-tools: Bash(openspec:*), Bash(git:*), Bash(jq:*), Bash(gh:*)
license: MIT
compatibility: Requires openspec CLI and jq.
metadata:
  author: gymie
  version: "2.0"
---

Report the pipeline state of every open (non-archived) OpenSpec change. **Read-only** — never commits, never runs git write operations, never advances a state. The one exception is state self-heal (see below), which corrects a stale cache to match the artifacts already on disk.

**Announce at start:** "Using myflow-status."

Follow all three contracts:

- **States** (`skills/myflow-contracts/pipeline.md`)
- **State file** (`skills/myflow-contracts/state-file.md`)
- **State self-heal** (`skills/myflow-contracts/state-self-heal.md`)

## Workflow

### 1. List open changes

```bash
openspec list --json
```

With a `<name>` argument, restrict to that change and include the detail view (step 4). With no argument, report every non-archived change.

Zero open changes → say so and suggest `/myflow-start`. Stop.

### 2. Resolve each change's state

For each change, resolve its user-scoped state file path per **State file** in `skills/myflow-contracts/state-file.md`, then read it:

```bash
MAIN_CHECKOUT="$(cd "$(dirname "$(git rev-parse --git-common-dir)")" && pwd)"
PROJECT_KEY="$(basename "$MAIN_CHECKOUT")-$(printf '%s' "$MAIN_CHECKOUT" | shasum | cut -c1-8)"
STATE_FILE="/Users/tweety53/Agents/myflow/state/$PROJECT_KEY/<name>.json"

jq -r '.state, .branch, .prUrl, .artifactUrl, .jiraIssue, .effort, .updatedAt, .updatedBy, (.worktrees // {} | keys[])' \
  "$STATE_FILE" 2>/dev/null
```

Then validate against artifacts per **State self-heal** (`skills/myflow-contracts/state-self-heal.md`) — including its "read artifacts from the apply worktree when one exists" rule. Resolve the worktree **first**, then root every subsequent artifact check there:

1. worktree resolution — take the `worktrees` keys, or find the apply worktree for branch `openspec/<name>` in `git worktree list`; if found, treat its path as the artifact root for the rest of this step, otherwise use the main checkout
2. `tasks.md` (at the resolved root) — count `- [x]` vs `- [ ]` items
3. `docs/manual-test/<name>.md` (at the resolved root) — exists? any unchecked boxes?
4. merge status — `git merge-base --is-ancestor <branch> origin/<base>`, which decides whether the next `/myflow-finish` integrates or archives
5. PR — `gh pr list --head <branch> --state all --json number,state,url`, **only when `gh` is installed and the remote is a GitHub host**. Otherwise (Bitbucket, no `gh`, no network, branch never pushed) PR state is **unknown** — report it as unknown and treat it as inconclusive, never as "no PR exists".

If the file is missing, unparseable, or contradicted, infer the state from artifacts, **rewrite the state file** — carrying forward every field this command did not infer, per **State self-heal** (`skills/myflow-contracts/state-self-heal.md`) — and mark the row `⚠`.

**Self-heal is state-only and monotonic** (per **State file** in `skills/myflow-contracts/state-file.md`):

- Infer and correct **`state` only**, and **never fabricate a `prUrl`** — never invent one from a
  PR you happened to find, because a URL written from a guess is worse than a null.
- **Do apply the one permitted correction** the contract defines: when this command's own
  `gh pr list` probe (step 2, item 5) **conclusively** answers that no PR exists for the branch
  while the state file records a `prUrl`, clear `prUrl` to `null` and mark the row `⚠`. This
  command already gathers exactly that evidence for its own report, so it is the one place the
  correction can be made without an extra probe. An inconclusive answer — no `gh`, non-GitHub
  forge, no network — clears nothing.
- **Never write a state earlier than the one recorded.** A check that cannot be performed is not a contradiction: an undeterminable PR state leaves `prUrl` exactly as recorded.

### 3. Render the table

Sort by state order per **States** in `skills/myflow-contracts/pipeline.md` (`STARTED`,
`IN_PROGRESS`, `FINISHED`), then by `updatedAt` descending. Omit `FINISHED` changes — they are
archived.

```
## myflow status

| Change | Jira | State | PR | Next | Worktree / branch | Updated |
|--------|------|-------|----|------|-------------------|---------|
| kan-8-myflow-updates | KAN-8 | IN_PROGRESS | #42 open | review the diff + run the guide, then `/myflow-finish` | `/abs/path/openspec-kan-8-myflow-updates` @ `openspec/kan-8-myflow-updates` | 2h ago (/myflow-do) |
| active-workout-session-editing | — | STARTED | — | read the artifact, then `/myflow-do` | none | 19h ago (/myflow-start) |

⚠ = state file was stale and has been corrected from artifacts.
```

Worktree paths are **absolute**, taken from the `worktrees` keys. The **PR** column shows the
number and state when known, `—` when `prUrl` is null, and `?` when PR state could not be
determined.

The **Jira** column shows `jiraIssue` verbatim, or `—` when the change has no linked issue. This
is a **read-only** report: never call Jira, never transition an issue, never infer a key from the
change name.

Surface `artifactUrl` when present — the link to the published proposal artifact.

Surface `effort` the same way: the recorded level verbatim when there is one, and, when the field is
`null` or absent, `not recorded — planned at <default>`, filled in from the default level
under **Effort** in State file (`skills/myflow-contracts/state-file.md`) — that file is
canonical for which level is the default, and naming it here as well would be a second copy to keep
in step.
An absent `effort` is legal and is never a `⚠`:
per **State file** (`skills/myflow-contracts/state-file.md`), a file that omits it reads as `null`,
so it is neither a contradiction nor a field to infer. This report never writes `effort`; it carries
it forward untouched on a self-heal, like every other field it does not own.

Next-command mapping:

| State | Next |
|-------|------|
| `STARTED` | read the artifact, then `/myflow-do <name>` (or re-run `/myflow-start` to revise) |
| `IN_PROGRESS`, branch not merged | review the diff + run the guide, then `/myflow-finish <name>` (or re-run `/myflow-do` to fix) |
| `IN_PROGRESS`, branch merged | `/myflow-finish <name>` — it will archive |
| `FINISHED` | — |

The `IN_PROGRESS` row splits on merge status because `/myflow-finish` behaves differently either
side of it: it integrates before the merge and archives after. Say which run the operator is
about to get.

### 4. Detail view (only when `<name>` was given)

Add below the table:

- Linked Jira issue key, or "none linked"
- Task progress (`N/M` checked from `tasks.md`)
- Nested `<name>-fix-N` sub-changes, if any
- The manual test guide's **absolute** path + checked/total box count
- PR number, state, and URL when one exists
- Whether the branch has reached the base branch — i.e. which `/myflow-finish` run comes next
- Every `⚠` correction made, with the reason

## Guardrails

- **Never** commit, stage, push, merge, or archive.
- **Never** advance a state — only correct a stale cache to match artifacts.
- **Never** rewind a state, and never fabricate a `prUrl`. Clearing a `prUrl` that a conclusive
  probe disproved is the one permitted correction, per
  **State self-heal** (`skills/myflow-contracts/state-self-heal.md`); everything else about the
  file is read-only.
- **Never** create a worktree or branch.
- Report `gh` being unavailable, or a non-GitHub forge, as "PR state unknown" rather than guessing — and never clear `prUrl` on that unknown.
- Never guess a state when artifacts are ambiguous — show `?` and say which check was inconclusive.
- **No flags.** The only argument is the optional change name.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Status of all open changes | `/myflow-status` |
| Detail for one change | `/myflow-status <name>` |
| Pipeline reference | `/myflow-info` |
