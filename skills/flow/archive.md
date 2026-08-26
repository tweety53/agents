# Archive and clean up (run 2)

Loaded either by `skills/flow/integrate.md`'s merge-and-push route, in the same invocation, or by a
fresh bare `/flow <name>` invocation once `check-finish-preflight.sh` returns `RUN2` from every
worktree. Unchanged in procedure from the retired `/myflow-finish` run 2, except the stage keys.

**Load `skills/flow-contracts/artifacts-registry.md`** — every removal below is a row in it.

**`skills/flow-contracts/finish-contract.md` is canonical for the full procedure.** In outline,
each numbered step below is bracketed by its own mark, with three exceptions that run inside the
mark of the step before: step 2 (positioning) inside step 3's `flow.sync-archive`; step 6 (remove
the proposal artifact source) inside step 5's `flow.cleanup`; step 11 (restore the checkout) inside
step 10's `flow.push-archive`. **Eleven steps, eight marks.**

**Generate this run's own session token here, before this first mark** — a separate invocation from
the integrate phase above (unless chained straight through from merge-and-push, in which case reuse
that same run's token — see **Model resolution** and the token note in `skills/flow/SKILL.md`).

```bash
flow stage begin -command '/flow' -stage flow.verify-merge -harness <harness> -session-token mf-<literal-token> <name>
```

1. **Verify the merge** — a PR CLI when usable, otherwise `git merge-base --is-ancestor`. Fetch
   first. Not merged → this is not run 2; fall back to `skills/flow/integrate.md` and **archive
   nothing** — end this mark `-outcome not-run-2` and stop.

```bash
flow stage end   -command '/flow' -stage flow.verify-merge -outcome completed <name>
flow stage begin -command '/flow' -stage flow.sync-archive -harness <harness> -session-token mf-<literal-token> <name>
```

2. **Position the main checkout on the archive branch**, before anything else touches it. Resolve
   `<base>` with `resolve-base-branch.sh` against the apply worktree, then invoke
   `prepare-archive-branch.sh <main-checkout> <base> chore/archive-<name>`. Exit `0` → the checkout
   is on `chore/archive-<name>`, cut from a fast-forwarded `<base>`; continue to step 3. Anything
   else stops run 2 here, with nothing staged, committed, pushed or removed. The four exit codes are
   **Run 2 — the branch is merged** (`skills/flow-contracts/finish-contract.md`), step 2.

   **When the guard is absent**, perform the same positioning by hand, in the same order.

3. **Archive the change** — under the mark `flow.sync-archive`. Run `spectre archive "<name>"` in
   the main checkout. It `git mv`s `<project>/spectre/changes/<name>/` into
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

4. **Commit the archive** on `chore/archive-<name>` in the main checkout — no push here; step 10
   carries it.

   ```bash
   [ "$(git -C <main-checkout> branch --show-current)" = "chore/archive-<name>" ] \
     && git -C <main-checkout> add -A \
     && { git -C <main-checkout> diff --cached --quiet \
          || git -C <main-checkout> commit -m "chore(spectre): archive <name>"; }
   ```

   A branch mismatch is reported, naming the branch found, and stops the commit, leaving the change
   at `IN_PROGRESS`. The subject is the fixed literal shown, per **Commit scopes name the module**
   (`<agents repo>/rules/commit-scope-is-the-module.mdc`). `gather-self-review-context.sh` resolves
   this commit at step 9 by matching that subject line whole — **reproduce it exactly.**

```bash
flow stage end   -command '/flow' -stage flow.commit-archive -outcome completed <name>
flow stage begin -command '/flow' -stage flow.cleanup -harness <harness> -session-token mf-<literal-token> <name>
```

5. **Clean up the worktrees, the local branch and the remote branch, then remove the workspace's
   database and bucket.** The workspace half runs the project's `remove` command from **Project
   configuration** (`skills/flow-contracts/project-configuration.md`), run from the **main
   checkout**, with the workspace id re-derived here from the change name — never handed to this
   run. A project declaring no `## workspace isolation` section, or no `remove` command, has this
   half **skipped, not failed**. A removal that fails is **the one exception to the
   stop-at-the-first-failure rule**: report it and carry on to step 7, which decides the verdict
   from the project's survivor report, never from this command's exit code.
6. **Remove the proposal artifact source** — always the literal `null`/absent under `/flow`, since
   `publish-proposal-removed` means no `/flow` change ever publishes one, per **Temporary artifacts
   registry** (`skills/flow-contracts/artifacts-registry.md`)'s row for it. This step is
   unconditionally a no-op skip for a change that ran only under `/flow`; it still runs, and still
   reports the skip, because a change resumed from before this rework (one whose state file was
   written by the retired `/myflow-start`) may carry a real preserved copy to remove.

Steps 5 and 6 together are the one `flow.cleanup` stage:

```bash
flow stage end   -command '/flow' -stage flow.cleanup -outcome completed <name>
flow stage begin -command '/flow' -stage flow.verify-cleanup -harness <harness> -session-token mf-<literal-token> <name>
```

7. **Verify the cleanup.** Run `check-cleanup-complete.sh <repo> <name> <state-dir>` once per
   repository, after every removal above. `COMPLETE:` → report the cleanup as verified, **relay
   every clause the line carries after ` — ` word for word**, and go on to step 8 — a `SKIPPED:`
   clause there is never a pass. `LEFTOVER:` → name what remains and **stop without writing
   `FINISHED`**, leaving the change at `IN_PROGRESS`. **No verdict line at all, and a non-zero
   exit** → report it, leave the affected `worktrees` entries in the state file, and treat it as
   `LEFTOVER`.

```bash
flow stage end -command '/flow' -stage flow.verify-cleanup -outcome completed <name>
```

(`completed` on `COMPLETE:`, `leftover` on `LEFTOVER:` or on a missing verdict line — either way the
run stops here, at `IN_PROGRESS`, and nothing below runs.)

```bash
flow stage begin -command '/flow' -stage flow.write-finished -harness <harness> -session-token mf-<literal-token> <name>
```

8. **Write `FINISHED`** — reached only on `COMPLETE:` — clearing from `worktrees` **only the
   entries whose removal actually succeeded**. Carry `artifactUrl` (`null`), `jiraIssue`,
   `planningEffort` (`null`), `models.default` (`null`) and `prUrl` forward. This is the terminal
   write.

```bash
flow stage end -command '/flow' -stage flow.write-finished -outcome completed <name>
flow stage begin -command '/flow' -stage flow.self-review -harness <harness> -session-token mf-<literal-token> <name>
```

**Transition the issue to Done** after the state write, per **Jira integration**
(`skills/flow-contracts/jira-integration.md`). A run that stopped at step 7 transitions nothing.

9. **Run self-review.** The procedure — skippable per run with running it the default, gathering
   input via a script rather than an inline re-read, one combined reasoning pass across all five
   angles plus the rating, the per-angle filing ask, and the report path — is **Run 2 — the branch
   is merged** (`skills/flow-contracts/finish-contract.md`), step 9, canonical for it. What is
   specific to *executing* it here: the script invocation `gather-self-review-context.sh
   <archived-change-path> <name> <state-dir> <main-checkout>`, resolving `<archived-change-path>` as
   `<project>/spectre/changes/archive/<name>/`, and passing `<main-checkout>` as the trust anchor.

   The skip prompt fires first:

   > **Run self-review for this change?**
   > - **Yes — run it** *(default, recommended)*
   > - **No — skip**

   An explicit **No** stops step 9 here; the handoff's `Self-review` line reads `skipped`. A session
   with no interactive channel to present this prompt still runs self-review, exactly as an explicit
   **Yes** would.

   | # | Angle | Label |
   |---|-------|-------|
   | 1 | Problems encountered, and what pipeline change would avoid them | `myflow-fix` |
   | 2 | Token/time cost, and what would reduce it without quality loss | `myflow-cost` |
   | 3 | What went well, and how to reproduce it | `myflow-improvement` |
   | 4 | What could be automated or moved to a script | `myflow-automation` |
   | 5 | What could move to the Go app or its persistent storage | `myflow-stats-app` |

   **One combined pass** — never five separate dispatches. Every finding is explained in the message
   body first, before any prompt fires. The filing ask is **one multi-select prompt per angle**,
   defaulting to filing none:

   > **File any of this angle's findings as Jira issues?**
   > - **<finding 1>**
   > - **<finding 2>**
   > - **None — file nothing** *(default, recommended)*

   Then ask the operator to rate the run: **Rate this flow run, 1 (rough) to 5 (excellent):**

   Write `<project>/docs/self-review/<name>-self-review.md` — one section per angle, all five
   present; each finding one line naming its angle's label, the finding, and its disposition; an
   angle with no findings carrying an explicit none-marker — plus the rating — and commit it on
   `chore/archive-<name>` **in the main checkout**, not pushing here:

   ```bash
   [ "$(git -C <main-checkout> branch --show-current)" = "chore/archive-<name>" ] \
     && git -C <main-checkout> add -- docs/self-review/<name>-self-review.md \
     && { git -C <main-checkout> diff --cached --quiet \
          || git -C <main-checkout> commit -m "docs(self-review): <name> self-review report"; }
   ```

   A branch mismatch or a commit that FAILS is reported and stops this commit. The change stays
   `FINISHED` regardless.

```bash
flow stage end -command '/flow' -stage flow.self-review -outcome completed <name>
flow stage begin -command '/flow' -stage flow.push-archive -harness <harness> -session-token mf-<literal-token> <name>
```

10. **Push the archive branch and open its pull request.** Push `chore/archive-<name>`; open a pull
    request against `<base>` via a PR CLI when usable, else print the forge's create-PR URL and ask
    whether it was opened. This push carries both the archive commit and the self-review report, so
    there is no window in which the archive PR merges while the report is still unwritten. Never
    pushes to `<base>`.

    A failed push, or a failed PR creation, is reported and never moves the change off `FINISHED`.
11. **Restore the main checkout to `<base>`.** Successful or not — never leave the checkout on the
    archive branch. Runs inside step 10's mark.

```bash
flow stage end -command '/flow' -stage flow.push-archive -outcome completed <name>
```

```
## Finished

**Change:** <name>
**Archived:** spectre/changes/archive/<name>/ (committed on chore/archive-<name>)
**Archive PR:** <prUrl> | not pushed — <reason>; land it with: git -C <main-checkout> push -u origin chore/archive-<name>, then open a PR against <base>
**Worktrees:** removed | left alone — <reason>
**Remote branch:** deleted | already gone | not deleted — <reason>
**Cleanup:** verified
**Self-review:** <path> (rating: <n>/5) | skipped
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

For each worktree, run **every** check below before removing anything — the six-check sequence is
**Worktree cleanup** (`skills/flow-contracts/finish-contract.md`), canonical for it, and is not
restated in full here beyond the one override `/flow` carries forward from the retired
`/myflow-fast`:

**Never ask check 4's ignored-files confirmation before removing a worktree.** Report what
`--force` will destroy — how many ignored files, which are build output, and which are irreplaceable
together with whether they were already preserved — and proceed. This is a scoped override of the
disclosure ask in **Worktree cleanup** (`skills/flow-contracts/finish-contract.md`); it is safe
here for the same reason it was safe under `/myflow-fast`: the records worth keeping are already out
of the worktree by this point, committed at `flow.preserve-sessions`
(`skills/flow/integrate.md`). **Checks 1, 2, 3, 5 and 6 remain gates.** Check 6, the live-process
check, is named explicitly because it is the one this override could plausibly be read as reaching:
a live process is not a preserved record, so `HELD:` and the guard's exit 2 both stop `/flow` exactly
as they stop the base contract. Check 4 turning up something genuinely irreplaceable and
*unpreserved* is not this override's case: stop and ask.

## Guardrails

- **Never** merge the change branch in run 2; step 1 already proved it.
- **Never** state a cleanup rule here — **Temporary artifacts registry** (`skills/flow-contracts/artifacts-registry.md`)
  and **Worktree cleanup** (`skills/flow-contracts/finish-contract.md`) are canonical.
- **Never** report a cleanup as done without the verdict that says so, and **never write
  `FINISHED` over a leftover or an unverified cleanup**.
- **Never** `git add` the state file, and never move it into the archive.
- **Never** let a Jira call block the archive — one skipped-with-reason line.
- **Never** let self-review block, delay, or undo the `FINISHED` write.
- **Never** ask the self-review skip prompt, a per-angle filing ask, or the rating question before
  `FINISHED` has been written.
