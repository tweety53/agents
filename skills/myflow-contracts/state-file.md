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
  "stage": "awaiting-manual-test",
  "gates": {
    "reviewed": true,
    "tested": false,
    "prOpened": null,
    "prMerged": null
  },
  "worktree": "<main checkout of the primary repo>/.worktrees/openspec-<name>",
  "branch": "openspec/<name>",
  "originStage": null,
  "artifactUrl": null,
  "jiraIssue": null,
  "fastPath": null,
  "REVIEWED_TREE": null,
  "MERGE_BASE": null,
  "updatedAt": "2026-07-25T10:00:00Z",
  "updatedBy": "/myflow-manual-test"
}
```

- `stage` — one of the twelve values in **Pipeline stages** (`skills/myflow-contracts/pipeline.md`).
- `gates.tested` — `null` (stage not reached), `false` (guide written, not yet tested), `"skipped"` (Gate C intentionally bypassed — permanent, never overwritten), or `true` (testing completed and verified by `/myflow-review`, the only writer of `true`).
- `gates.reviewed` / `gates.prOpened` / `gates.prMerged` — boolean, or `null` before that stage is reached.
- `worktree` — absolute path, or `null` when no worktree exists.
- `originStage` — the stage `/myflow-do-fix` was invoked from; `null` when no fix is in flight.
- `artifactUrl` — the published proposal artifact's URL; `null` until `/myflow-start` publishes one.
- `jiraIssue` — the key of the Jira issue driving this change (e.g. `"KAN-7"`), or `null` when no issue is linked. Written only by `/myflow-start` (and `/myflow-fast-path` when it creates the change); every other command **carries it forward verbatim**. See **Jira integration** (jira-integration.md).
- `fastPath` — `true` when the change reached its current stage via `/myflow-fast-path`; `null`/absent otherwise. Written only by `/myflow-fast-path`; every other command **carries it forward verbatim**.
- `REVIEWED_TREE` — the `git write-tree` hash(es) recorded at a `/myflow-fast-path checkpoint` stop, proving the reviewed diff has not moved; `null` otherwise. Written only by `/myflow-fast-path`; every other command **carries it forward verbatim**.
- `MERGE_BASE` — the commit each affected worktree branched from, so a later session can still run `git diff MERGE_BASE`; `null` otherwise. Written by `/myflow-fast-path`; every other command **carries it forward verbatim**. (`/myflow-do` and `/myflow-do-fix` record their own `MERGE_BASE` in the SDD progress ledger at `.superpowers/sdd/progress-<name>.md` instead — the fast path uses the state file because it keeps no ledger.)
- `updatedAt` — the ISO-8601 UTC instant of the last write. Read the actual current time
  (`date -u +%Y-%m-%dT%H:%M:%SZ`); never invent or placeholder it, since `/myflow-status` reports
  "last update" from this field and a fabricated value makes a stalled change look freshly touched.
- `updatedBy` — the command that last wrote the file, e.g. `"/myflow-do"`.

**Multi-repo shape for `REVIEWED_TREE` and `MERGE_BASE`.** Both are objects **keyed by the absolute worktree path** of each affected repo, so a change spanning two repos records one entry per repo:

```json
"MERGE_BASE": {
  "<main checkout of the primary repo>/.worktrees/openspec-<name>": "5ee4c9a…",
  "<main checkout of the second repo>/.worktrees/openspec-<name>": "b31f7c2…"
}
```

The scalar `worktree` and `branch` fields name the **primary** repo only. When a change spans repos, the full set of affected worktrees is the **key set of `MERGE_BASE`** — that is the authoritative list, and `worktree` is one of its keys.

Read it with:

```bash
jq -r '.stage' "$STATE_FILE" 2>/dev/null || echo "MISSING"
```

Write it by rendering the full object above — always write every field, never a partial merge. Never `git add` it, never include it in a commit, and never move it into `openspec/changes/archive/`.

**Gate values are monotonic.** No command may lower a gate value. `gates.tested: true` and `gates.tested: "skipped"` are **sticky** — once set they are never overwritten or demoted. No command may write a `stage` earlier than the one it found, other than the single carve-out described under **State self-heal** (`skills/myflow-contracts/state-self-heal.md`). A self-heal may only **raise or fill** a value, never lower one, and must **never infer `gates.tested: true`** — only `/myflow-review` writes `true`.

Because writes render the whole object, every command must first **read the existing file and carry forward** every gate it does not itself own (e.g. `/myflow-manual-test` re-emits the `prOpened`/`prMerged` values it read, rather than resetting them to `null`). The same applies to
every non-gate field it does not own — including `artifactUrl`, `jiraIssue`, `fastPath`,
`REVIEWED_TREE` and `MERGE_BASE`: re-emit them as read. Dropping them would erase the published
proposal link, silently unlink the change from its Jira issue, strand a fast-path change
mid-resume, or destroy the authoritative list of affected worktrees for a multi-repo change.
