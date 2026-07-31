# State file

**The state file lives outside the repo, under a user-scoped directory.** It is **never committed, never staged, and never archived** — it is machine-local metadata, not part of the change.

```text
/Users/tweety53/Agents/myflow/state/<project-key>/<name>.json
```

`<project-key>` = `<basename of main checkout>-<first 8 hex of sha1 of the main checkout's absolute path>` — e.g. `myrepo-3f9a1c02`. The basename keeps it readable; the hash makes two same-named repos in different directories unambiguous.

**Resolving the main checkout is load-bearing.** `git rev-parse --show-toplevel` returns the *worktree* root when run inside a worktree, which would give apply (in a worktree) and review (in the main checkout) two different state files for the same change. Always resolve via `--git-common-dir`, which points at the **main** repo's `.git` from anywhere, including inside a worktree:

```bash
MAIN_CHECKOUT="$(cd "$(dirname "$(git rev-parse --git-common-dir)")" && pwd)"
PROJECT_KEY="$(basename "$MAIN_CHECKOUT")-$(printf '%s' "$MAIN_CHECKOUT" | shasum | cut -c1-8)"
STATE_FILE="/Users/tweety53/Agents/myflow/state/$PROJECT_KEY/<name>.json"
mkdir -p "$(dirname "$STATE_FILE")"
```

Every command resolves the path this way, so the same change maps to the same file from the main checkout and from every worktree.

```json
{
  "state": "IN_PROGRESS",
  "branch": "openspec/<name>",
  "worktrees": {
    "/absolute/path/to/worktree": "<merge-base sha>"
  },
  "artifactUrl": null,
  "jiraIssue": null,
  "effort": null,
  "prUrl": null,
  "updatedAt": "2026-07-28T10:00:00Z",
  "updatedBy": "/myflow-do"
}
```

- `state` — one of the three values in **States** (`skills/myflow-contracts/pipeline.md`):
  `STARTED`, `IN_PROGRESS`, `FINISHED`.
- `branch` — the change's branch, `openspec/<name>`; `null` before one exists.
- `worktrees` — an object **keyed by the absolute path** of each affected worktree, whose value is
  that worktree's merge base. `{}` when none exist or all were removed. **A `FINISHED` change may
  legitimately carry a non-empty map:** `/myflow-finish` clears only the entries whose removal
  actually succeeded, so a worktree that could not be removed stays listed and remains findable.
  See **Multi-repo shape** below.
- `artifactUrl` — the published proposal artifact's URL; `null` until `/myflow-start` publishes one.
- `jiraIssue` — the key of the Jira issue driving this change (e.g. `"KAN-8"`), or `null` when no issue is linked. Written only by `/myflow-start`; every other command **carries it forward verbatim**. See **Jira integration** (`jira-integration.md`).
- `effort` — the reasoning effort chosen for this change's planning: `"low"`, `"medium"`, `"high"`, or
  `null` when none was chosen. Written only by `/myflow-start`, on the run that **creates** the
  change; every other command **carries it forward verbatim**. It governs `/myflow-start`'s own
  reasoning depth and nothing else — no command derives behaviour from it, and the review panel's
  breadth is never scaled from it. See **Effort** (`state-file.md`) below.

  **A state file that omits `effort` entirely is valid**, and is read as `null`. This is a
  deliberate exception to the closed-schema rule in
  **State self-heal** (`skills/myflow-contracts/state-self-heal.md`), which otherwise makes a file
  unparseable both for missing a documented field and for carrying an undocumented one. Without
  the exception every
  file written before this field existed would be routed through self-heal, which announces
  unrecovered fields and rewrites from artifact inference — a loud correction for a value nobody
  had the chance to set. `effort` is the first field added since the schema closed, so the
  carve-out is stated rather than inferred: `artifactUrl`, `jiraIssue` and `prUrl` are all
  *present and nullable*, which is a different thing from *absent*.
- `prUrl` — the pull request's URL once one is open; `null` otherwise. Its non-nullness is what
  records that a PR was opened, so no separate boolean exists. It is also what tells `/myflow-do`
  that a fix must be committed and pushed rather than merely staged.
- `updatedAt` — the ISO-8601 UTC instant of the last write. Read the actual current time
  (`date -u +%Y-%m-%dT%H:%M:%SZ`); never invent or placeholder it, since `/myflow-status` reports
  "last update" from this field and a fabricated value makes a stalled change look freshly touched.
- `updatedBy` — the command that last wrote the file, e.g. `"/myflow-do"`.

**This file records no human confirmation and no fix origin.** No command observes whether the
human ran the apps, so nothing could honestly confirm that a human reviewed the work. And a fix
never moves the state, so there is no origin state for a fix to return to.

**Multi-repo shape.** `worktrees` carries one entry per affected repository, so a change spanning
two repos records both:

```json
"worktrees": {
  "/Users/tweety53/Projects/agents-worktrees/openspec-<name>": "5ee4c9a…",
  "/Users/tweety53/Projects/other-worktrees/openspec-<name>": "b31f7c2…"
}
```

The **key set of `worktrees` is the authoritative list of affected worktrees** — it is what
`/myflow-finish` cleans up, and what resolves an app's root when a handoff needs an absolute
path. The scalar `branch` names the shared branch only. Never infer a worktree path from a
conventional layout; layout differs per repository.

Read it with:

```bash
jq -r '.state' "$STATE_FILE" 2>/dev/null || echo "MISSING"
```

Write it by rendering the full object above — always write every field, never a partial merge. Never `git add` it, never include it in a commit, and never move it into `openspec/changes/archive/`.

**Five fields from the twelve-stage predecessor are retired and never appear in a state file this
contract produces.** A file still carrying one is unparseable under
**State self-heal** (`skills/myflow-contracts/state-self-heal.md`):

- the field that recorded where a dynamic-target write originated — meaningless once targets
  stopped being dynamic.
- the boolean flag that recorded a change having taken the shortened single-session route — that
  route is gone, and its shape does not survive in three states. It did **not** mark a change as
  skipping review: the route it recorded ran a review panel of its own, just a smaller roster than
  the standard one. What actually skipped review was a separate *invocation flag* on the
  twelve-stage cycle, which was never a state-file field at all and so has no tombstone here.
- the cache key that pinned a stage-advance script's last-scanned commit — the script it belonged
  to no longer exists.
- the four-valued object that tracked pass/fail across four named checkpoints of the old
  twelve-stage pipeline — collapsed into the single `state` field above.
- the boolean recording whether a human had exercised the change — no command ever observes this
  honestly (see "This file records no human confirmation" above), so the field recorded a claim
  nobody could verify.

**State writes are monotonic.** No command may write a `state` earlier than the one it found. The
single carve-out is described under
**State self-heal** (`skills/myflow-contracts/state-self-heal.md`) and clears `prUrl` rather than
moving the state.

Because writes render the whole object, every command must first **read the existing file and
carry forward** every field it does not itself own — `artifactUrl`, `jiraIssue`, `prUrl` and
`worktrees` among them. Re-emit each as read (`null` only if it was already `null`). Dropping one
erases it permanently: the published proposal link, the link to the Jira issue, the PR (which also
silently downgrades the next fix from commit-and-push to staged-only), or the authoritative list of
worktrees for a multi-repo change.

## Effort

**Which file to change first.** The normative requirement — that three levels exist, that `medium`
is the default offered, and that no level may switch a gate off — is stated in
`openspec/specs/myflow-effort/spec.md`, under *Effort scales the reasoning spent inside the gates,
never the gates themselves*. That spec is the requirement; the table below is the **operational form
the commands read**, and it exists here so `/myflow-start` has one place to look rather than a
requirements document to interpret. Change the spec first and bring this table with it: a table that
contradicts the requirement is this file's defect, not the spec's.

Three levels, offered by `/myflow-start` on the run that creates a change, with `medium` the default:

| Level | What it changes |
|-------|-----------------|
| `low` | Questions batched rather than asked one at a time; the design presented once; `tasks.md` grouped more coarsely |
| `medium` | The checklist followed with related questions grouped |
| `high` | Each checklist item worked separately, alternatives enumerated per open question, each design section approved on its own |

**No level may switch a gate off.** Brainstorming runs, the design approval gate holds,
writing-plans runs, and `tasks.md` is never left a thin scaffold — at every level. A lower level
means fewer rounds and coarser grouping, never a gate that does not run. An effort level able to
skip a gate would be a way to skip review rather than a way to size the thinking inside it.
