# Finish — one run, merge-and-push only

Loaded by `skills/flow-fast/SKILL.md` on a bare invocation at `IN_PROGRESS`. `/flow-fast` lands by
**merge-and-push only** — there is no landing question, no open-PR or manual route, and no second
run. All twelve finish-side stage keys mark this one procedure.

## Preflight

**Check for a `RUN1` verdict first.** `check-finish-preflight.sh <worktree> <base-ref>
<recorded-merge-base|->` decides whether this change's branch has already reached the base branch,
exactly as `skills/flow-contracts/finish-contract-run1.md`'s own **Finish contract** verdict table
defines — cited, never restated. A **`RUN2`** verdict here is the wrong-state handoff (**Wrong
state for this command**, `skills/flow-contracts/pipeline.md`): `/flow-fast` runs no separate
archive invocation, so an already-merged branch means this run already completed, or completed by
some other means. A **`REFUSE`** stops and asks, per that same contract.

```bash
flow stage begin -command '/flow-fast' -stage flow.preflight -harness <harness> -session-token ff-<literal-token> <name>
flow stage end   -command '/flow-fast' -stage flow.preflight -outcome completed <name>
```

**Then the unfinished-work gate** — `check-unfinished-work.sh`, the same three-course prompt
(`Stop`/`Continue`/`File or join a Jira follow-up`) on `OUTSTANDING:`, per that same contract's
**Run 1** section:

```bash
flow stage begin -command '/flow-fast' -stage flow.unfinished-work-gate -harness <harness> -session-token ff-<literal-token> <name>
# … the check, and the prompt on OUTSTANDING: …
flow stage end   -command '/flow-fast' -stage flow.unfinished-work-gate -outcome completed <name>
```

**Then check whether the base branch has moved** — `check-base-moved.sh`, exactly as that same
contract describes, asked only on an overlapping `MOVED`.

**These three guards run as one chained call** — `check-finish-preflight.sh`,
`check-unfinished-work.sh` and `check-base-moved.sh` in sequence, prompting only on
`OUTSTANDING:` or an overlapping `MOVED:`; a `REFUSE` from either of the first two stops the run
the same way.

**No landing question, and `## default landing route` is never read.** `/flow-fast` has exactly
one route — a change that needs a pull request or a manual handoff is a `/flow` run, not a
`/flow-fast` one.

```bash
flow stage begin -command '/flow-fast' -stage flow.landing-question -harness <harness> -session-token ff-<literal-token> <name>
flow stage end   -command '/flow-fast' -stage flow.landing-question -outcome completed <name>
```

`flow.landing-question` marks and closes with nothing asked — this run has no branch to take.

## Reshape and commit

```bash
flow stage begin -command '/flow-fast' -stage flow.preserve-sessions -harness <harness> -session-token ff-<literal-token> <name>
flow stage end   -command '/flow-fast' -stage flow.preserve-sessions -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.commit-two -harness <harness> -session-token ff-<literal-token> <name>
```

`git -C <abs-worktree> reset --soft <recorded-merge-base>`, then the guarded two-commit chain
(**Git boundaries**, `skills/flow-contracts/git-boundaries.md`) — cited, never restated: the
implementation commit, then the fixed `chore(spectre): plan and session records` commit carrying
`<project>/spectre/changes/` and the session records under `<project>/docs/superpowers/`.

```bash
flow stage end -command '/flow-fast' -stage flow.commit-two -outcome completed <name>
```

## Merge and archive on `<base>`

```bash
flow stage begin -command '/flow-fast' -stage flow.landing-routes -harness <harness> -session-token ff-<literal-token> <name>
```

Push the change branch is **skipped** — merge-and-push never pushes `spectre/<name>` itself, only
`<base>` at the end. Resolve `<base>` (`resolve-base-branch.sh`), then:

```bash
prepare-archive-branch.sh <project>/.worktrees/_landing-<name> <base> <base>
git -C <landing-worktree> merge --no-ff spectre/<name>
```

`<archive-branch>` equal to `<base>` itself positions the landing worktree directly on `<base>` —
`prepare-archive-branch.sh`'s own accepted shape, per **Finish contract**
(`skills/flow-contracts/finish-contract-run1.md`)'s merge-and-push route. A merge conflict stops
the run, reports the landing worktree path, and leaves it as `git merge` left it. The main checkout
is never checked out, staged or committed.

```bash
flow stage end -command '/flow-fast' -stage flow.landing-routes -outcome completed <name>
```

```bash
flow stage begin -command '/flow-fast' -stage flow.verify-merge -harness <harness> -session-token ff-<literal-token> <name>
flow stage end   -command '/flow-fast' -stage flow.verify-merge -outcome completed <name>
```

The merge just performed **is** the verification — `flow.verify-merge` marks and closes with
nothing further to check.

```bash
flow stage begin -command '/flow-fast' -stage flow.sync-archive -harness <harness> -session-token ff-<literal-token> <name>
```

`spectre archive <name>` in `<landing-worktree>`, **on `<base>`** — never a `chore/archive-<name>`
branch; there is no second branch on this route.

```bash
flow stage end   -command '/flow-fast' -stage flow.sync-archive -outcome completed <name>
flow stage begin -command '/flow-fast' -stage flow.commit-archive -harness <harness> -session-token ff-<literal-token> <name>
```

`git -C <landing-worktree> add -A`, `check-archive-scope.sh <landing-worktree> "spectre/changes/"`
between the add and the commit (**Run 2 — the branch is merged**,
`skills/flow-contracts/finish-contract-run2.md`, step 4 — cited, never restated), then commit on
`<base>` — asserting that branch rather than assuming it. A `SCOPE-VIOLATION` refuses the commit
and leaves the change at `IN_PROGRESS`.

```bash
flow stage end -command '/flow-fast' -stage flow.commit-archive -outcome completed <name>
```

## One push

```bash
flow stage begin -command '/flow-fast' -stage flow.push-archive -harness <harness> -session-token ff-<literal-token> <name>
git -C <landing-worktree> push origin <base>
```

One push, once — the merge and the archive commit both ride on it. A refused push stops the run and
reports the landing worktree path; nothing here retries.

```bash
flow stage end -command '/flow-fast' -stage flow.push-archive -outcome completed <name>
```

## Cleanup — stop and process scan only

```bash
flow stage begin -command '/flow-fast' -stage flow.cleanup -harness <harness> -session-token ff-<literal-token> <name>
```

Checks 1–4 of **Worktree cleanup** (`skills/flow-contracts/finish-contract-run2.md`) are **skipped
outright** — checks 1–3 are already proven by construction (commit-two left the tree clean, the
merge just performed makes `HEAD` an ancestor of `<base>`), and check 4's ignored-files disclosure
prompt is skipped with it, per that section's own `/flow-fast` override. Checks 5 (`## stop`,
60-second bound) and 6 (`check-worktree-processes.sh`, cwd outside every worktree first) stay as
gates exactly as that section defines them — cited, never restated. `HELD:` or a non-zero exit from
either stops the run at `IN_PROGRESS` with the pids.

Then, for the apply worktree:

```bash
git -C <repo> worktree remove --force <abs-worktree>
git -C <repo> branch -d spectre/<name>
git -C <repo> worktree prune
```

**No remote branch delete** — the change branch was never pushed on this route, so there is nothing
to delete. The project's workspace `remove` command still runs, via `flow workspace-id <name>`
(**Project configuration**, `skills/flow-contracts/project-configuration.md`).

`check-cleanup-complete.sh` is **skipped outright** — no verify-cleanup pass runs for `/flow-fast`.

```bash
flow stage end -command '/flow-fast' -stage flow.cleanup -outcome completed <name>
```

## Finish

```bash
flow stage begin -command '/flow-fast' -stage flow.write-finished -harness <harness> -session-token ff-<literal-token> <name>
```

Write `FINISHED`, clearing from `worktrees` only the entries whose removal actually succeeded.

```bash
flow stage end -command '/flow-fast' -stage flow.write-finished -outcome completed <name>
```

**Then, and only now that `FINISHED` is written, one Jira transition to Done** — no `In Review`
hop. This ordering matters: the push (above) and the cleanup gates (**Cleanup**, above) can each
still stop the run at `IN_PROGRESS`, and Jira's own **Transitions** contract
(`skills/flow-contracts/jira-integration.md`) places every terminal transition after its
corresponding state write for exactly that reason — a run that stopped before this point never
reports Done for a change that is not.

**No self-review** — no subagent is dispatched, per this command's own design. Remove the proposal
artifact source per the temporary-artifacts registry's condition, if one exists.

```bash
git -C <main-checkout> worktree remove --force <landing-worktree>
git -C <main-checkout> worktree prune
```

The change is `FINISHED`. This invocation is terminal — nothing further to run.

## Guardrails

- **Merge-and-push is the only route** — never ask the landing question, never read `## default
  landing route`, never push the change branch itself, never open a pull request.
- Never commit `<project>/spectre/changes/` or `<project>/docs/superpowers/` in a task commit —
  only in the planning-artifacts commit above.
- Never a `chore/archive-<name>` branch, a second merge, a second push, or a remote branch delete.
- Only a completed run writes `FINISHED` — a `LEFTOVER`-equivalent stop (check 5 or 6 above)
  leaves the change at `IN_PROGRESS`.
