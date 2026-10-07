# Clean up (run 2)

Loaded either by `skills/flow/integrate.md`'s merge-and-push route, in the same invocation, or by a
fresh bare `/flow <name>` invocation once `check-finish-preflight.sh` returns `RUN2` from every
worktree.

**Load `skills/flow-contracts/artifacts-registry.md`** — every removal below is a row in it.

**`skills/flow-contracts/finish-contract-run2.md` is canonical for the full procedure.** In outline,
each numbered step below is bracketed by its own mark, with one exception that runs inside the
mark of the step before: step 3 (remove the proposal artifact source) inside step 2's
`flow.cleanup`. **Six steps, five marks.** The archive and the self-review pass are run
1's (`skills/flow/integrate.md`).

```bash
flow stage begin -command '/flow' -stage flow.verify-merge -harness <harness> -session-token mf-<literal-token> <name>
```

1. **Verify the merge** — a PR CLI when usable, otherwise `git merge-base --is-ancestor` against
   `origin/<base>`. Fetch first. Not merged → this is not run 2; fall back to `skills/flow/integrate.md` and **remove
   nothing** — end this mark `-outcome not-run-2` and stop.

```bash
flow stage end   -command '/flow' -stage flow.verify-merge -outcome completed <name>
flow stage begin -command '/flow' -stage flow.cleanup -harness <harness> -session-token mf-<literal-token> <name>
```

2. **Clean up the worktrees, the local branch and the remote branch, then remove the workspace's
   database and bucket.** The workspace half runs the `remove` row of the command table
   `project-get.sh <main-checkout> "workspace isolation"` prints (exit 1: the skipped-not-failed
   case below), from the **main checkout**, with `<id>` from `flow workspace-id <name>` — never
   handed to this run. A project declaring no `## workspace isolation` section, or no `remove`
   command, has this half **skipped, not failed**. A removal that fails is **the one exception to the
   stop-at-the-first-failure rule**: report it and carry on to step 4, which decides the verdict
   from the project's survivor report, never from this command's exit code.
3. **Remove the proposal artifact source** — delete `<state-dir>/<name>-proposal-artifact.html`
   when present (`<state-dir>` the path `flow state dir` prints for this repository), per **Temporary artifacts registry**
   (`skills/flow-contracts/artifacts-registry.md`)'s row for it. `/flow` has written none since
   `publish-proposal-removed`, but a change created before it may still hold one, and step 4's
   guard reports a survivor as a leftover. Absent, there is nothing to remove, and the step says so.

Steps 2 and 3 together are the one `flow.cleanup` stage:

```bash
flow stage end   -command '/flow' -stage flow.cleanup -outcome completed <name>
flow stage begin -command '/flow' -stage flow.verify-cleanup -harness <harness> -session-token mf-<literal-token> <name>
```

4. **Verify the cleanup.** Run `check-cleanup-complete.sh <repo> <name> <state-dir>` once per
   repository, after every removal above; its verdicts are step 4 of **Run 2 — the branch is merged**
   (`skills/flow-contracts/finish-contract-run2.md`).

```bash
flow stage end -command '/flow' -stage flow.verify-cleanup -outcome completed <name>
```

(`completed` on `COMPLETE:`, `leftover` on `LEFTOVER:` or on a missing verdict line — either way the
run stops here, at `IN_PROGRESS`, and nothing below runs.)

```bash
flow stage begin -command '/flow' -stage flow.write-finished -harness <harness> -session-token mf-<literal-token> <name>
```

5. **Write `FINISHED`** — reached only on `COMPLETE:` — clearing from `worktrees` **only the
   entries whose removal actually succeeded**. Carry `artifactUrl`, `jiraIssue`,
   `planningEffort`, `models.default` and `prUrl` forward as recorded. This is the terminal
   write.

```bash
flow stage end -command '/flow' -stage flow.write-finished -outcome completed <name>
```

**Load `skills/flow-contracts/jira-integration.md`.** **Transition the issue to Done** after the state write, per **Jira integration**
(`skills/flow-contracts/jira-integration.md`). A run that stopped at step 4 transitions nothing.

```bash
flow stage begin -command '/flow' -stage flow.refresh-main-checkout -harness <harness> -session-token mf-<literal-token> <name>
```

6. **Bring the main checkout forward:**

   ```bash
   refresh-main-checkout.sh <main-checkout> <base>
   ```

   Step 6 of **Run 2 — the branch is merged** (`skills/flow-contracts/finish-contract-run2.md`);
   a `REFRESH-REFUSED` line goes into the handoff verbatim.

```bash
flow stage end -command '/flow' -stage flow.refresh-main-checkout -outcome completed <name>
```

```
## Finished

**Change:** <name>
**Archived:** spectre/changes/archive/<name>/ (landed on <base> with the change)
**Worktrees:** removed | left alone — <reason>
**Remote branch:** deleted | already gone | not deleted — <reason>
**Cleanup:** verified
**Self-review:** docs/self-review/<name>-self-review.md
**Main checkout:** fast-forwarded | already current | <the REFRESH-REFUSED line>
**Guards:** all present | N missing — those checks were performed by hand (see the guard presence check above)
**Jira:** <KEY> → Done | none linked | ⚠ Jira: skipped — <reason>
```

A run 2 that **completes step 5** is terminal and names **no** next command.

On a leftover — or on no verdict at all — the run stops at step 4 instead:

```
## Cleanup incomplete — not finished

**Change:** <name>
**Remaining:** <what the guard named> | unverified — <what the guard reported on stderr>
**State:** IN_PROGRESS — FINISHED is not written while anything remains

<what the operator must clear>

Next:
/flow <name>
```

## Worktree cleanup

Run `remove-change-worktrees.sh` once per repository per **Worktree cleanup**
(`skills/flow-contracts/finish-contract-run2.md`), canonical for its checks, exit codes and relay:

**Check 4 asks only about an irreplaceable, unpreserved entry**, per check 4 of **Worktree cleanup**
(`skills/flow-contracts/finish-contract-run2.md`), canonical for the buckets and that one ask;
everything else it finds is reported and the removal proceeds. **Checks 1, 2, 3, 5 and 6 remain
gates.** Check 6, the live-process check, is named explicitly because it is the one check 4's
proceed-without-asking could plausibly be read as reaching: a live process is not a preserved
record, so `HELD:` and the guard's exit 2 both stop `/flow`.

## Guardrails

- **Never** `git add` the state file, and never move it into the archive.
