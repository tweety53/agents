# Finish — the same two contracts, run 2 minus two steps

Loaded by `skills/flow-fast/SKILL.md` on a bare invocation at `IN_PROGRESS`. Follows the same two
contracts `skills/flow/integrate.md` and `skills/flow/archive.md` follow — cited in full, never
restated — with every mark carrying `-command '/flow-fast'` in place of `-command '/flow'`, and run
2 skipping two steps outright.

## Deciding which run this is

Exactly `skills/flow/integrate.md`'s own **Deciding which run this is**: the branch's merge status
against its base, checked by `check-finish-preflight.sh`, decides run 1 vs. run 2 — never a stored
field.

## Run 1 — the branch is not merged

Follow `skills/flow-contracts/finish-contract-run1.md` and `skills/flow/integrate.md`'s own
sections exactly:

**On a `RUN1` verdict**, mark `flow.preflight` (closed immediately, since the verdict is already
in hand) then `flow.unfinished-work-gate`:

```bash
flow stage begin -command '/flow-fast' -stage flow.preflight -harness <harness> -session-token ff-<literal-token> <name>
flow stage end   -command '/flow-fast' -stage flow.preflight -outcome completed <name>
```

1. **Check for unfinished work** — `check-unfinished-work.sh`, the same three-course prompt
   (`Stop`/`Continue`/`File or join a Jira follow-up`) on `OUTSTANDING:`.

   ```bash
   flow stage begin -command '/flow-fast' -stage flow.unfinished-work-gate -harness <harness> -session-token ff-<literal-token> <name>
   # … the check, and the prompt on OUTSTANDING: …
   flow stage end   -command '/flow-fast' -stage flow.unfinished-work-gate -outcome completed <name>
   ```
2. **Ask how the branch should land** — open PR, merge and push, or manual — or read a project's
   `## default landing route` instead of asking, exactly as `skills/flow/integrate.md`'s second
   section does.

   ```bash
   flow stage begin -command '/flow-fast' -stage flow.landing-question -harness <harness> -session-token ff-<literal-token> <name>
   flow stage end   -command '/flow-fast' -stage flow.landing-question -outcome completed <name>
   ```
3. **Preserve sessions**, **commit the staged work as two commits** (the task commits already made
   by `skills/flow-fast/implement.md`, plus the planning-artifacts commit covering
   `<project>/spectre/changes/<name>/` and `<project>/docs/superpowers/`), and **take the chosen
   route** — exactly `skills/flow/integrate.md`'s third and fourth sections.

   ```bash
   flow stage begin -command '/flow-fast' -stage flow.preserve-sessions -harness <harness> -session-token ff-<literal-token> <name>
   flow stage end   -command '/flow-fast' -stage flow.preserve-sessions -outcome completed <name>
   flow stage begin -command '/flow-fast' -stage flow.commit-two -harness <harness> -session-token ff-<literal-token> <name>
   flow stage end   -command '/flow-fast' -stage flow.commit-two -outcome completed <name>
   flow stage begin -command '/flow-fast' -stage flow.landing-routes -harness <harness> -session-token ff-<literal-token> <name>
   flow stage end   -command '/flow-fast' -stage flow.landing-routes -outcome completed <name>
   ```

Open PR and manual routes stop and hand off here, at `IN_PROGRESS`, naming the next command as
`/flow-fast <name>` (bare) once the branch is merged. Merge-and-push continues into run 2 in the
same invocation.

## Run 2 — the branch is merged

Follow `skills/flow-contracts/finish-contract-run2.md` and `skills/flow/archive.md`'s own run-2
structure exactly, through its numbered steps:

1. Verify the merge.

   ```bash
   flow stage begin -command '/flow-fast' -stage flow.verify-merge -harness <harness> -session-token ff-<literal-token> <name>
   flow stage end   -command '/flow-fast' -stage flow.verify-merge -outcome completed <name>
   ```
2. Position the landing worktree on the archive branch — `prepare-archive-branch.sh`, per that
   contract's own exit contract.
3. Archive the change (`spectre archive <name>`) and 4. commit it on `chore/archive-<name>` — no
   push.

   ```bash
   flow stage begin -command '/flow-fast' -stage flow.sync-archive -harness <harness> -session-token ff-<literal-token> <name>
   flow stage end   -command '/flow-fast' -stage flow.sync-archive -outcome completed <name>
   flow stage begin -command '/flow-fast' -stage flow.commit-archive -harness <harness> -session-token ff-<literal-token> <name>
   flow stage end   -command '/flow-fast' -stage flow.commit-archive -outcome completed <name>
   ```
5. Clean up the worktrees, the local and remote branch, and the workspace's database/bucket.

   ```bash
   flow stage begin -command '/flow-fast' -stage flow.cleanup -harness <harness> -session-token ff-<literal-token> <name>
   flow stage end   -command '/flow-fast' -stage flow.cleanup -outcome completed <name>
   ```
6. Remove the proposal artifact source, per the temporary-artifacts registry's condition.
7. **Skipped outright** — `check-cleanup-complete.sh`'s verify-cleanup pass never runs for
   `/flow-fast`, per `skills/flow-contracts/finish-contract-run2.md`'s own `/flow-fast` exemption at
   this step. Cleanup itself (step 5) still ran; only its separate verification does not. Proceed
   to step 8 unconditionally — there is no `COMPLETE:`/`LEFTOVER:` verdict to gate on.
8. Write `FINISHED`, clearing from `worktrees` only the entries whose removal actually succeeded.

   ```bash
   flow stage begin -command '/flow-fast' -stage flow.write-finished -harness <harness> -session-token ff-<literal-token> <name>
   flow stage end   -command '/flow-fast' -stage flow.write-finished -outcome completed <name>
   ```
9. **Skipped outright** — no self-review subagent is dispatched, per that same contract's
   `/flow-fast` exemption at this step.
10. Push the archive branch and land it — the merge-and-push continuation pushes and merges into
    `<base>`; a standalone run 2 pushes and opens a pull request.

    ```bash
    flow stage begin -command '/flow-fast' -stage flow.push-archive -harness <harness> -session-token ff-<literal-token> <name>
    flow stage end   -command '/flow-fast' -stage flow.push-archive -outcome completed <name>
    ```
11. Remove the landing worktree, successful or not.

Because step 9 never runs, `/flow-fast`'s run 2 never touches Jira's In Review → Done transition
timing that a filed self-review finding would otherwise interact with — the base run 1 → In Review
and run 2 → Done transitions (`skills/flow-contracts/jira-integration.md`'s **Transitions** table)
are unaffected and still fire exactly as they do for `/flow`.

## Guardrails

Exactly `skills/flow/integrate.md`'s and `skills/flow/archive.md`'s own **Guardrails** — never push,
merge, or open a PR outside these routes; never commit `<project>/spectre/changes/` or
`<project>/docs/superpowers/` in a task commit (only in the planning-artifacts commit above); only
run 2 completing writes `FINISHED`.
