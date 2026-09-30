# Archive and clean up (run 2)

Loaded either by `skills/flow/integrate.md`'s merge-and-push route, in the same invocation, or by a
fresh bare `/flow <name>` invocation once `check-finish-preflight.sh` returns `RUN2` from every
worktree.

**Load `skills/flow-contracts/artifacts-registry.md`** — every removal below is a row in it.

**`skills/flow-contracts/finish-contract-run2.md` is canonical for the full procedure.** In outline,
each numbered step below is bracketed by its own mark, with three exceptions that run inside the
mark of the step before: step 2 (positioning) inside step 3's `flow.sync-archive`; step 6 (remove
the proposal artifact source) inside step 5's `flow.cleanup`; step 11 (remove the landing
worktree) inside step 10's `flow.push-archive`. **Eleven steps, eight marks.**

```bash
flow stage begin -command '/flow' -stage flow.verify-merge -harness <harness> -session-token mf-<literal-token> <name>
```

1. **Verify the merge** — a PR CLI when usable, otherwise `git merge-base --is-ancestor`. Fetch
   first. On the merge-and-push continuation the merge is still local, so the test is
   `git -C <landing-worktree> merge-base --is-ancestor spectre/<name> <base>`. Not merged → this is not run 2; fall back to `skills/flow/integrate.md` and **archive
   nothing** — end this mark `-outcome not-run-2` and stop.

```bash
flow stage end   -command '/flow' -stage flow.verify-merge -outcome completed <name>
flow stage begin -command '/flow' -stage flow.sync-archive -harness <harness> -session-token mf-<literal-token> <name>
```

2. **Position the landing worktree on the archive branch**, before anything else touches it.
   Run `resolve-base-branch.sh` against the apply worktree for `<base>`, run
   `classify-untracked.sh <project>/.worktrees/_landing-<name>` first when the worktree already
   exists — the archive pre-flight settling its untracked entries, **Run 2 — the branch is
   merged** (`skills/flow-contracts/finish-contract-run2.md`), step 2, canonical for the classes
   and for settling a reported asset through the operator — then invoke
   `prepare-archive-branch.sh <project>/.worktrees/_landing-<name> <base> chore/archive-<name>`.
   The four exit codes are **Run 2 — the branch is merged**
   (`skills/flow-contracts/finish-contract-run2.md`), step 2.

3. **Archive the change** — under the mark `flow.sync-archive`. Run `spectre archive "<name>"` in
   `<landing-worktree>`. It `git mv`s `<project>/spectre/changes/<name>/` into
   `<project>/spectre/changes/archive/<name>/` and leaves the rename staged; it does not commit —
   step 4 does.

   **One call per change, parent and each `<name>-fix-N` sibling alike** — run `spectre archive
   "<name>-fix-N"` for every sub-change in this same step.

   **`spectre archive` refuses four things; `--force` overrides three of them.** Unchecked tasks, a
   missing `tasks.md`, and a `tasks.md` carrying no task at all each exit `1` and say `(use --force
   to archive anyway)`. A destination that already exists also exits `1`, and `--force` does
   **not** override that one. **`/flow` never passes `--force`.** Report the refusal, stop, and
   leave the change at `IN_PROGRESS`.

```bash
flow stage end   -command '/flow' -stage flow.sync-archive -outcome completed <name>
flow stage begin -command '/flow' -stage flow.commit-archive -harness <harness> -session-token mf-<literal-token> <name>
```

4. **Commit the archive** on `chore/archive-<name>` in `<landing-worktree>` — no push here; step 10
   carries it. `<canonical-worktree>` resolves as run 1 resolves it (**Finish contract**,
   `skills/flow-contracts/finish-contract-run1.md`):

   ```bash
   commit-archive.sh <landing-worktree> <canonical-worktree> <name>
   ```

   `ARCHIVE-COMMITTED: <sha>` or `ARCHIVE-NOTHING-STAGED` (exit 0) continue.
   `ARCHIVE-WRONG-BRANCH: <found>` or the scope guard's `SCOPE-VIOLATION` lines (exit 1), or exit
   2, stop the commit and leave the change at `IN_PROGRESS`.

```bash
flow stage end   -command '/flow' -stage flow.commit-archive -outcome completed <name>
flow stage begin -command '/flow' -stage flow.cleanup -harness <harness> -session-token mf-<literal-token> <name>
```

5. **Clean up the worktrees, the local branch and the remote branch, then remove the workspace's
   database and bucket.** The workspace half runs the `remove` row of the command table
   `project-get.sh <main-checkout> "workspace isolation"` prints (exit 1: the skipped-not-failed
   case below), from the **main checkout**, with `<id>` from `flow workspace-id <name>` — never
   handed to this run. A project declaring no `## workspace isolation` section, or no `remove`
   command, has this half **skipped, not failed**. A removal that fails is **the one exception to the
   stop-at-the-first-failure rule**: report it and carry on to step 7, which decides the verdict
   from the project's survivor report, never from this command's exit code.
6. **Remove the proposal artifact source** — delete `<state-dir>/<name>-proposal-artifact.html`
   when present (`<state-dir>` the path `flow state dir` prints for this repository), per **Temporary artifacts registry**
   (`skills/flow-contracts/artifacts-registry.md`)'s row for it. `/flow` has written none since
   `publish-proposal-removed`, but a change created before it may still hold one, and step 7's
   guard reports a survivor as a leftover. Absent, there is nothing to remove, and the step says so.

Steps 5 and 6 together are the one `flow.cleanup` stage:

```bash
flow stage end   -command '/flow' -stage flow.cleanup -outcome completed <name>
flow stage begin -command '/flow' -stage flow.verify-cleanup -harness <harness> -session-token mf-<literal-token> <name>
```

7. **Verify the cleanup.** Run `check-cleanup-complete.sh <repo> <name> <state-dir>` once per
   repository, after every removal above; its verdicts are step 7 of **Run 2 — the branch is merged**
   (`skills/flow-contracts/finish-contract-run2.md`).

```bash
flow stage end -command '/flow' -stage flow.verify-cleanup -outcome completed <name>
```

(`completed` on `COMPLETE:`, `leftover` on `LEFTOVER:` or on a missing verdict line — either way the
run stops here, at `IN_PROGRESS`, and nothing below runs.)

```bash
flow stage begin -command '/flow' -stage flow.write-finished -harness <harness> -session-token mf-<literal-token> <name>
```

8. **Write `FINISHED`** — reached only on `COMPLETE:` — clearing from `worktrees` **only the
   entries whose removal actually succeeded**. Carry `artifactUrl`, `jiraIssue`,
   `planningEffort`, `models.default` and `prUrl` forward as recorded. This is the terminal
   write.

```bash
flow stage end -command '/flow' -stage flow.write-finished -outcome completed <name>
flow stage begin -command '/flow' -stage flow.self-review -harness <harness> -session-token mf-<literal-token> <name>
```

**Load `skills/flow-contracts/jira-integration.md`.** **Transition the issue to Done** after the state write, per **Jira integration**
(`skills/flow-contracts/jira-integration.md`). A run that stopped at step 7 transitions nothing.

9. **Save the self-review context bundle.** Self-review is always deferred: this step writes and
   lands the bundle, and `/flow-self-review` is the only reasoning pass — **Run 2 — the branch is
   merged** (`skills/flow-contracts/finish-contract-run2.md`), step 9, canonical for it.

   Write the bundle's stdout, then `## Session narrative` (one paragraph this session writes for
   run 2 itself — the archived `design.md` and `narrative.md` are bundle sections now, and a change
   predating the narrative rule has the bundle report it skipped), to
   `<project>/docs/self-review/<name>-context.md` physically under `<landing-worktree>`, and commit
   it on `chore/archive-<name>` **in `<landing-worktree>`**, not pushing here:

   ```bash
   land-self-review-report.sh "<landing-worktree>" "chore/archive-<name>" \
     "docs(self-review): <name> self-review context bundle" \
     docs/self-review/<name>-context.md
   ```

A branch mismatch or a commit that FAILS is reported and stops this commit. A staged index
holding anything beyond the chain's own path refuses the commit the same way
(`LAND-FOREIGN-STAGED`). The change stays `FINISHED` regardless.

```bash
flow stage end -command '/flow' -stage flow.self-review -outcome completed <name>
flow stage begin -command '/flow' -stage flow.push-archive -harness <harness> -session-token mf-<literal-token> <name>
```

10. **Push the archive branch and land it.** The procedure — the two-row route table keyed on how
    this run of `archive.md` was reached, the guarantee that the archive commit and step 9's
    self-review context bundle always land together — on the merge-and-push
    continuation, in the one push of `<base>` that also carries run 1's merge — and the
    failure-reporting rules —
    is **Run 2 — the branch is merged** (`skills/flow-contracts/finish-contract-run2.md`), step
    10, canonical for it.
11. **Remove the landing worktree.** Successful or not — never leave it behind for a later run to
    trip over:

    ```bash
    git -C <main-checkout> worktree remove --force <landing-worktree>
    git -C <main-checkout> worktree prune
    ```

    ```bash
    refresh-main-checkout.sh <main-checkout> <base>
    ```

    Runs inside step 10's mark. The refresh is step 11 of **Run 2 — the branch is merged**
    (`skills/flow-contracts/finish-contract-run2.md`); a `REFRESH-REFUSED` line goes into the
    handoff verbatim.

```bash
flow stage end -command '/flow' -stage flow.push-archive -outcome completed <name>
```

```
## Finished

**Change:** <name>
**Archived:** spectre/changes/archive/<name>/ (committed on chore/archive-<name>)
**Archive PR:** <prUrl> (merged) | none — merged directly into <base> | none — <base> not pushed: <reason>; land it with: git -C <main-checkout> push origin <base> | <prUrl> — open, not yet merged: <reason>; land it with: gh pr merge <prUrl> --merge | not pushed — <reason>; land it with: git -C <main-checkout> push -u origin chore/archive-<name>, then open and merge a PR against <base>
**Worktrees:** removed | left alone — <reason>
**Remote branch:** deleted | already gone | not deleted — <reason>
**Cleanup:** verified
**Self-review:** deferred — docs/self-review/<name>-context.md
**Guards:** all present | N missing — those checks were performed by hand (see the guard presence check above)
**Jira:** <KEY> → Done | none linked | ⚠ Jira: skipped — <reason>
```

A run 2 that **completes step 10** is terminal and names **no** next command.

On a leftover — or on no verdict at all — the run stops at step 7 instead:

```
## Cleanup incomplete — not finished

**Change:** <name>
**Archived:** spectre/changes/archive/<name>/ (committed on chore/archive-<name>)
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
- **Never** let self-review block, delay, or undo the `FINISHED` write.
- **Never** fetch or write the self-review context bundle before `FINISHED` has been written.
