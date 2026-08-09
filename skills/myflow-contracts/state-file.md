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
  "planningEffort": null,
  "models": {
    "implementation": null,
    "reviewPanel": null,
    "panelFix": null
  },
  "reviewPanelRoster": null,
  "prUrl": null,
  "updatedAt": "2026-07-28T10:00:00Z",
  "updatedBy": "/myflow-do"
}
```

**A state file is unparseable when it is not valid JSON, omits a documented field other than
`planningEffort`, `models` or `reviewPanelRoster`, or carries an undocumented one.** An unparseable file is reported and
skipped, never rebuilt from inference — no command infers a state file's contents. JSON that parses
but is missing one or more of the fields this contract requires is unparseable in full on that
account alone, not partially recovered.

- `state` — one of the three values in **States** (`skills/myflow-contracts/pipeline.md`):
  `STARTED`, `IN_PROGRESS`, `FINISHED`.
- `branch` — the change's branch, `openspec/<name>`; `null` before one exists.
- `worktrees` — an object **keyed by the absolute path** of each affected worktree, whose value is
  that worktree's merge base. `{}` when none exist or all were removed. **A `FINISHED` change may
  legitimately carry a non-empty map:** `/myflow-finish` clears only the entries whose removal
  actually succeeded, so a worktree that could not be removed stays listed and remains findable.
  See **Multi-repo shape** below.

  **A `null` value is legal and means *no merge base recorded* for that path** — it can occur in a
  hand-edited or out-of-band-modified file. Every rule this contract and the pipeline already state
  for a **missing** recorded merge base applies to a `null` one unchanged: the preflight is handed
  `-` and refuses, the merge status is `inconclusive`, and the review command falls back to the
  staged diff. A `null` value is therefore never a licence to infer a merge base — it is the refusal
  to.
- `artifactUrl` — the published proposal artifact's URL; `null` until `/myflow-start` publishes one.
- `jiraIssue` — the key of the Jira issue driving this change (e.g. `"KAN-8"`), or `null` when no issue is linked. Written only by `/myflow-start`; every other command **carries it forward verbatim**. See **Jira integration** (`jira-integration.md`).
- `planningEffort` — the level chosen for this change's planning, or `null` when none was chosen.
  Written only by `/myflow-start`, on the run that **creates** the change; every other command
  **carries it forward verbatim**. It governs `/myflow-start`'s own reasoning depth and nothing
  else — no command derives behaviour from it, and the review panel's breadth is never scaled from
  it. The levels, and which of them is offered as the recommendation, are stated once under
  **Planning effort** (`state-file.md`) below and are deliberately not repeated here.
- `models` — an object carrying `implementation`, `reviewPanel` and `panelFix`, each naming the
  model chosen for that role, or `null` where none was chosen. Written only by `/myflow-start`, on
  the run that **creates** the change; every other command **carries it forward verbatim**. Its
  live consumer is `/myflow-do`, which dispatches on those values. The roles, their defaults and
  how an operator override applies are stated once under
  **Model policy** (`skills/myflow-contracts/pipeline.md`), which is canonical for them; a second
  copy here is what this repository's reference guard exists to prevent. These fields record what
  was *chosen* — the SDD ledger remains the only record of what a dispatch actually ran on.

  **A state file that omits `planningEffort`, `models` or `reviewPanelRoster` entirely is valid**,
  and each absent key is read as *not recorded*. This is a deliberate exception to the closed-schema
  rule stated above, which otherwise makes a file unparseable both for missing a documented field and
  for carrying an undocumented one. Without the exception, every file written before these fields
  existed would be unparseable and so reported and skipped — a spurious report against a value nobody
  had the chance to set. The carve-out is stated rather than inferred, and it covers a key that is
  **absent**: `artifactUrl`, `jiraIssue` and `prUrl` are all *present and nullable*, which is a
  different thing from *absent*. For `reviewPanelRoster`, *not recorded* resolves to the default
  preset rather than leaving the panel unconfigured, so a command never has to ask which roster to
  use at panel time.

  **A file carrying the retired `effort` key is read as recording the equivalent level** — `medium`
  as `default`, `high` as `detailed`, `low` as `low` — and is rewritten under `planningEffort` on
  the next write that file receives. It is not unparseable, and the rewrite is **not announced as a
  correction**: the value was written correctly under the contract in force when it was set. **No
  migration pass is run** — no command sweeps existing state files, and a file nothing writes to
  keeps the old key indefinitely without that being a fault.

  **The compatibility read is the fallback, not a preference, and every consumer performs it.** A
  command that reads only `planningEffort` reports a file recording a real level as having recorded
  none — the exact outcome this exception exists to prevent, and a promise made in prose and
  implemented nowhere is not a promise. The mechanical form is
  `(.planningEffort // .effort)`, as `/myflow-status` reads it.

  **When a file carries both keys, `planningEffort` wins.** Both are individually excepted from the
  closed-schema rule, so such a file parses, and it needs a stated answer rather than an implied
  one. The current key is the one this contract writes and the retired one is read only for a file
  that never had it, so a file carrying both is one already migrated whose old key was never
  cleared — the new value is the later of the two. The next write drops `effort` rather than
  re-emitting it; a write renders the whole object, so leaving it out is what removes it.

  **A value outside those three reads as *not recorded*, and never makes the file unparseable.**
  The retired key never held anything else under the contract in force when it was written, so a
  file carrying, say, `urgent` under it is not one this pipeline produced: it maps to no level, and
  surfacing the raw value would put a level this pipeline does not have in front of the operator.
  *Not recorded* is what remains once an unrecognisable value is discarded, and it is the honest
  answer for a value with no defined target.

  **Reading an unmapped value as *unparseable* was specified first and withdrawn, and the reason is
  recorded here so it is not reinstated.** *Unparseable* is only worth saying if some command detects
  the key and treats the file specially on account of it, and none does: no command infers or
  rebuilds a state file's contents, and `/myflow-status` reads the file through a literal `jq`
  projection that silently ignores every key it does not name. A rule whose only effect is an
  announcement no command emits is worse than no rule, because it reads as a guarantee. *Not
  recorded* needs no detection to be true, loses nothing the mapping had not already discarded, and
  is what the commands actually do.
- `reviewPanelRoster` — carries the review panel roster preset chosen for the change, one of
  `light`, `standard` or `full`. Written only by `/myflow-start`, on the run that **creates** the
  change; every other command **carries it forward verbatim**. Its live consumer is `/myflow-do`,
  which selects the panel's required slots and the per-task review's shape from it.
  `skills/myflow-do/SKILL.md` is canonical for what each preset means, and this file does not
  restate it. The field is top-level rather than nested under `models` because a roster is not a
  model.
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

The **key set of `worktrees` is the authoritative recorded list of affected worktrees** — it is
what `/myflow-finish` cleans up, and what resolves an app's root when a handoff needs an absolute
path. It is the record, not the iteration set: a step that needs "the worktrees" resolves that set
first, per **Resolving a change's worktrees** (`skills/myflow-contracts/pipeline.md`), rather than
looping over this map directly. The scalar `branch` names the shared branch only. Never infer a
worktree path from a conventional layout; layout differs per repository.

Read it with:

```bash
jq -r '.state' "$STATE_FILE" 2>/dev/null || echo "MISSING"
```

Write it by rendering the full object above — always write every field, never a partial merge. Never `git add` it, never include it in a commit, and never move it into `openspec/changes/archive/`.

**Five fields from the twelve-stage predecessor are retired and never appear in a state file this
contract produces.** A file still carrying one is unparseable under the closed-schema rule stated
above:

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

**State writes are monotonic.** No command may write a `state` earlier than the one it found.

Because writes render the whole object, every command must first **read the existing file and
carry forward** every field it does not itself own — `artifactUrl`, `jiraIssue`, `prUrl`,
`reviewPanelRoster` and `worktrees` among them. Re-emit each as read (`null` only if it was already
`null`). For `reviewPanelRoster` there is no mapping to perform, so re-emit as read is the whole
rule. Dropping one erases it permanently: the published proposal link, the link to the Jira issue,
the PR (which also silently downgrades the next fix from commit-and-push to staged-only), the
chosen review panel roster preset, or the authoritative list of worktrees for a multi-repo change.

**Carrying the planning effort forward is what performs the rewrite**, and it is the one field where
*verbatim* needs saying precisely. A file read through the retired-key fallback above is carried
forward as its **mapped level under `planningEffort`** — that is the rewrite the exception promises,
and the only write that ever performs it. Re-emitting the old key instead leaves the file
permanently unmigrated, and emitting `planningEffort: null` because the new key was absent from the
read erases a level the operator chose. Every command that writes a state file it did not create
does this, so no single command is responsible for a migration none of them run.

## Planning effort

**Which file to change first.** The normative requirement — that three levels exist, that `default`
is the level offered as the recommendation, and that no level may switch a gate off — is
**Requirement: Planning effort scales the reasoning spent inside the gates, never the gates themselves** (`openspec/specs/myflow-planning-effort/spec.md`).
Naming the requirement in full, rather than giving the path alone, is what makes
`scripts/check-references.sh` check this pointer — an OpenSpec `### Requirement: …` heading is a
heading like any other. The guard skips a path that does not resolve, so this one is checked only
once the capability lands in `openspec/specs/` at finish run 2; until then it is a reference nobody
verifies, which is said here rather than left to look otherwise. That spec is the requirement; the
table below is the
**operational form the commands read**, and it exists here so `/myflow-start` has one place to look
rather than a requirements document to interpret. Change the spec first and bring this table with
it: a table that contradicts the requirement is this file's defect, not the spec's.

Three levels, offered by `/myflow-start` on the run that creates a change, with `default` the level
offered as the recommendation:

| Level | What it changes |
|-------|-----------------|
| `low` | Questions batched rather than asked one at a time; the design presented once; `tasks.md` grouped more coarsely |
| `default` | The checklist followed with related questions grouped |
| `detailed` | Each checklist item worked separately, alternatives enumerated per open question, each design section approved on its own |

**The retired key belongs to the field, not to this table.** `effort`, its mapping onto these three
levels, which key wins when a file carries both, and what an unmapped value reads as are all stated
with the `planningEffort` field above — one statement, beside the field it governs. This section is
the levels themselves; it names where that rule lives rather than carrying a second copy of it.

**No level may switch a gate off.** Brainstorming runs, the design approval gate holds,
writing-plans runs, and `tasks.md` is never left a thin scaffold — at every level. A lower level
means fewer rounds and coarser grouping, never a gate that does not run. A planning effort level
able to skip a gate would be a way to skip review rather than a way to size the thinking inside it.
