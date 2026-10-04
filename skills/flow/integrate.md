# Integrate (run 1)

Loaded by `skills/flow/SKILL.md` on a **bare** invocation at `IN_PROGRESS` — no argument.

## Deciding which run this is

**`skills/flow-contracts/finish-contract-run1.md` is canonical for every procedure below** — the
base-branch resolution, the preflight checks, the removal sequence and their rationales live in
that file. This file carries only what is specific to *executing* it under `/flow`.

**Load `skills/flow-contracts/worktree-resolution.md`** too.

**Check guard presence** per **Guard presence check** (`skills/flow-contracts/pipeline.md`),
already run at the top of this invocation.

**First, surface the foreign staged work and the main checkouts' drift** — **Surface foreign
staged work before the preflight** (`skills/flow-contracts/finish-contract-run1.md`) is canonical
for it: run `check-foreign-staged.sh` and `check-main-checkout-drift.sh` once per distinct main
checkout.

Run `check-finish-preflight.sh` once per worktree in the set found by **Resolving a change's
worktrees** (`skills/flow-contracts/finish-contract-run1.md`) — never a raw read of the state file's
`worktrees` map. Its `<base-ref>` argument is `origin/$BASE`, `$BASE` being what running
`resolve-base-branch.sh <worktree>` prints — the same `<base>` every route below lands on.

- **`RUN1`** → this file (integrate)
- **`RUN2`** from every worktree → `skills/flow/cleanup.md` <!-- refs-guard:allow -->
- **`REFUSE`** → stop, report what the script reported, relay the guard's hand-verification
  procedure per **Hand-verifying a guard verdict** (`skills/flow-contracts/pipeline.md`), and ask
  the operator
- **A resolved set that comes back empty** → stop and ask, exactly as `REFUSE`
- **No verdict line at all, and exit 2** → treat exactly as `REFUSE`

**On a `RUN1` verdict**, mark `flow.preflight` (closed immediately, since the verdict is already in
hand) then `flow.unfinished-work-gate`:

```bash
flow stage begin -command '/flow' -stage flow.preflight -harness <harness> -session-token mf-<literal-token> <name>
flow stage end   -command '/flow' -stage flow.preflight -outcome completed <name>
flow stage begin -command '/flow' -stage flow.unfinished-work-gate -harness <harness> -session-token mf-<literal-token> <name>
```

## 1. Check for unfinished work

**An archive a stopped run 1 left uncommitted is undone first**, per **Run 1 — the branch is not
merged** (`skills/flow-contracts/finish-contract-run1.md`). **A change already archived** — decided
once, in the canonical repository, for every worktree, per the same section — skips
**4**'s `flow.sync-archive` mark with its `spectre archive` call, unmarked. **Without new work**, per
the same section, it also ends `flow.unfinished-work-gate` `completed` at once and runs **2**, then
**4** from `flow.commit-archive` on, then **5**; the rest of **1** and all of **3** are skipped,
unmarked. **With new work** every other step runs.

Run `check-unfinished-work.sh <worktree> <name> <canonical-worktree>` once per worktree in the
resolved set — before the landing question and before any git action but the undo of an uncommitted archive (**Run 1 — the branch is not merged**, `skills/flow-contracts/finish-contract-run1.md`).

Then run `check-visual-verify-dispatched.sh <worktree> <name> <recorded-merge-base>` once per
worktree in the same set, per **Run 1 — the branch is not merged**
(`skills/flow-contracts/finish-contract-run1.md`): `VISUAL-VERIFY-OK:` joins `CLEAR:`;
`VISUAL-VERIFY-MISSING:` and an exit 2 take the `OUTSTANDING:` and stop-and-ask rows below.

- **`CLEAR:` from every worktree** → continue to **2** with no extra prompt.
- **A resolved set that comes back empty** → stop and ask the operator.
- **`OUTSTANDING:`** → **Load `skills/flow/unfinished-work-gate.md`** — the breakdown, the
  three-course prompt and what each course does are there.
- **No verdict line at all, and a non-zero exit** → stop and ask.

```bash
flow stage end -command '/flow' -stage flow.unfinished-work-gate -outcome completed <name>
```

(`completed` on **Continue** or **File or join…**, `stopped` on **Stop**.)

## 2. Ask how the branch should land

```bash
flow stage begin -command '/flow' -stage flow.landing-question -harness <harness> -session-token mf-<literal-token> <name>
```

**Check whether the base branch has moved first, then sync onto it.** Run `check-base-moved.sh`
once per worktree in the resolved set and report every verdict, per **Finish contract**
(`skills/flow-contracts/finish-contract-run1.md`); a `REFUSE`, an exit 2 or an empty resolved set
stops and asks, closing the mark:

```bash
flow stage end -command '/flow' -stage flow.landing-question -outcome stopped <name>
```

Every `MOVED` worktree is then rebased — no prompt, conflicts resolved in place — per **Sync the
branch onto the base** (`skills/flow/sync-onto-base.md`), which is canonical for
the rebase, `<rebased-merge-base>`, the resolution rule, the stop-and-ask cases and the
after-resolution lint and test run; a stop there closes the mark `stopped` exactly as above. **Load `skills/flow/sync-onto-base.md` only when** a worktree's verdict is `MOVED`, or the merge-and-push route's push was rejected and the sync is re-run. No
`MOVED` verdict anywhere → report the counts and go straight to the landing question.

Run `project-get.sh <main-checkout> "default landing route" --enum "pull request" "merge and push" manual`:
exit 0 prints the resolved route; exit 1 means absent; exit 3 means the head matches none of the
three — report its stderr line by name and resolve as absent.

**A resolved default skips the question entirely** — take that route without asking, and say so
in the handoff (`Route: <route> — from this project's configured default, not asked`). Only an
absent or unresolved default falls back to asking:

> **How should this branch land?**
> - **Open a pull request** *(default, recommended)*
> - **Merge and push**
> - **Handle it manually**

Having asked once, run to completion without asking again.

**Report an existing PR before asking.** If a PR exists for this branch — from `prUrl`, or a PR CLI
when one is usable — say so, open or closed-unmerged, before the operator answers.

```bash
flow stage end -command '/flow' -stage flow.landing-question -outcome completed <name>
```

## 3. Commit the staged work

**Load `skills/flow-contracts/git-boundaries.md`** before either commit below.

**Load `skills/flow-contracts/git-boundaries-commit-chain.md`** before either commit below.

```bash
flow stage begin -command '/flow' -stage flow.preserve-sessions -harness <harness> -session-token mf-<literal-token> <name>
```

**Append this run's own narrative first**, the same append `flow.write-in-progress` <!-- refs-guard:allow -->
(`skills/flow/verify-and-handoff.md`) makes, heading `## <YYYY-MM-DD> — integrate run`, covering
this run's preflight, the unfinished-work gate and the rebase.

**Before any route commits, reshape the branch:**

```bash
reshape-branch.sh <worktree> <name> <recorded-merge-base>
```

`<recorded-merge-base>` is the merge base recorded in the state
file's `worktrees` map for this worktree — **or `<rebased-merge-base>` from step 2 above, for a
worktree this run rebased**, never the state file's now-stale pre-rebase value for that worktree.
**On an archived re-run with new work, pass the archived re-run's base instead**, found in the
range from that merge base (**Run 1 — the branch is not merged**,
`skills/flow-contracts/finish-contract-run1.md`).

**Load `skills/flow-contracts/session-records.md`** before rendering, below.

**Render the ledger first**, so a missing one is caught here — into the `<canonical-worktree>`
section 1 resolved, so a multi-worktree run writes one copy, not one per worktree:

```bash
flow record render -change <name> -kind ledger -repo <canonical-worktree>
```

**A change with no dispatch rows reports `MISSING: ledger` and exits 0** — never a failure. **A
non-zero exit means a destination was refused or could not be written:** report it and continue.

```bash
flow stage end   -command '/flow' -stage flow.preserve-sessions -outcome completed <name>
flow stage begin -command '/flow' -stage flow.commit-two -harness <harness> -session-token mf-<literal-token> <name>
```

Then stage and commit twice, in this order:

```bash
commit-split.sh <worktree> <name> \
  "<type>(<module>): <what the implementation does>" \
  "chore(spectre): plan"
```

`<type>`, `<module>` and `<what the implementation does>` are derived from the reshaped diff. The
planning message's subject is the fixed literal `chore(spectre): plan`; on **Continue** at **1** its
message also lists the outstanding work, as below.

**Run that as one command.** The guards, the skipped-empty rule, the stop-on-failure rule and the
symlinked-planning-path case are all under **The guarded two-commit chain**
(`skills/flow-contracts/git-boundaries-commit-chain.md`).

**Implementation first, planning delta second.** The second commit's message lists anything the
operator chose to integrate over at **1**. The state file is **not** committed.

```bash
flow stage end -command '/flow' -stage flow.commit-two -outcome completed <name>
```

## 4. Archive and save the self-review context bundle

In `<canonical-worktree>`, on `spectre/<name>`, per **Archive on the change branch** and **Save the
self-review context bundle** (`skills/flow-contracts/finish-contract-run1.md`), canonical for both.

```bash
flow stage begin -command '/flow' -stage flow.sync-archive -harness <harness> -session-token mf-<literal-token> <name>
```

Run `spectre archive "<name>"` in `<canonical-worktree>`. It `git mv`s `<project>/spectre/changes/<name>/` into
`<project>/spectre/changes/archive/<name>/` and leaves the rename staged; it does not commit —
the next mark does.

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

```bash
commit-archive.sh <canonical-worktree> <name>
```

`ARCHIVE-COMMITTED: <sha>` or `ARCHIVE-NOTHING-STAGED` (exit 0) continue.
`ARCHIVE-WRONG-BRANCH: <found>` or the scope guard's `SCOPE-VIOLATION` lines (exit 1), or exit
2, stop the commit and leave the change at `IN_PROGRESS`. The `DONE-WHEN-PATH:` lines of the
Done-when cross-check refuse the commit the same way, before anything is staged or copied.

```bash
flow stage end   -command '/flow' -stage flow.commit-archive -outcome completed <name>
flow stage begin -command '/flow' -stage flow.self-review -harness <harness> -session-token mf-<literal-token> <name>
```

Write the bundle's stdout, then `## Session narrative`, to
`<project>/docs/self-review/<name>-context.md` in `<canonical-worktree>`, and commit it there, not
pushing here:

```bash
land-self-review-report.sh "<canonical-worktree>" "spectre/<name>" \
  "docs(self-review): <name> self-review context bundle" \
  docs/self-review/<name>-context.md
```

A branch mismatch or a commit that FAILS is reported and stops the run at `IN_PROGRESS`. A staged
index holding anything beyond the chain's own path refuses the commit the same way
(`LAND-FOREIGN-STAGED`).

```bash
flow stage end -command '/flow' -stage flow.self-review -outcome completed <name>
```

## 5. Take the chosen route, write the state, and transition Jira

```bash
flow stage begin -command '/flow' -stage flow.landing-routes -harness <harness> -session-token mf-<literal-token> <name>
```

This stage carries three sub-steps under one mark: the git route, the state write, and the Jira transition.

Per **Finish contract** (`skills/flow-contracts/finish-contract-run1.md`) → run 1. Push with `-u` so
the branch has an upstream.

**Every git step here can fail, and none of them may fail silently.** A rejected push, a commit
blocked by a hook, or `gh pr create` erroring must be **reported with the
command's own output**, and the run must **stop** leaving the change at `IN_PROGRESS` — save the
merge-and-push route's one re-sync after a rejected push (**The routes**,
`skills/flow-contracts/finish-contract-run1.md`).

**Human confirmation is a legitimate substitute for an API probe** on a forge with no usable CLI. If
the answer is No, leave `prUrl` null and say what to do next.

**Sub-step: write the state.** Write `state` unchanged at `IN_PROGRESS`, `prUrl` set if a PR was
opened, and every other field carried forward.

**Load `skills/flow-contracts/jira-integration.md`** — it is canonical for the transition below.

**Load `skills/flow-contracts/jira-integration-finish.md`** too — its **In Review is not tied to a pull request** governs the transition below.

**Sub-step: transition the issue to In Review** per **Transitions**
(`skills/flow-contracts/jira-integration.md`): after the state write, never before, never
blocking.

```bash
flow stage end -command '/flow' -stage flow.landing-routes -outcome completed <name>
```

## Handoff

```
## Branch integrated — waiting on the merge | merged and waiting on run 2

**Change:** <name>
**Route:** pull request | merged and pushed | manual
**PR:** <prUrl> | none — merged directly | none — you are handling it
**Outstanding:** <what step 1 reported and the operator integrated over> | none
**Guards:** all present | N missing — those checks were performed by hand (see the guard presence check above)

<what the operator must do before the next run>

Next:
/clear
/flow <name>
```

| Route taken in step 5 | Heading |
|--------------------|---------|
| pull request | *waiting on the merge* |
| **merge and push** | *merged and waiting on run 2* — continue below |
| manual | *waiting on the merge* |

Where the route is not certain — a run resumed after a partial failure — take the answer from the
merge-status test in **The block each state renders**
(`skills/flow-contracts/handoff-blocks.md`) rather than assuming.

## After merge-and-push specifically

Continue, within the same invocation and without a further command from the operator, into
`skills/flow/cleanup.md` exactly as written. Nothing external blocks this route.

## After open PR or manual specifically

Stop after the route completes, printing the handoff above. Each of these two routes needs an
action outside this command's control before the branch merges. The next bare `/flow <name>`
call, once the branch is integrated, runs run 2's cleanup (`skills/flow/cleanup.md`).
