# Finish contract — run 1 (the branch is not merged)

**This file is canonical for `/flow`'s integrate run** — the preflight-signal decision, run 1's
procedure, and resolving a change's worktrees.

bare `/flow` is the only command that loads this file whole; `/flow-status` cites single sections of it.

**Load `skills/flow-contracts/finish-hand-fallbacks.md` only when** the guard presence check named
one of the scripts this file calls missing, or a call finds the script absent — every by-hand
procedure for them is there.

## Finish contract

bare `/flow` is a **two-run** command. Which run happens is decided by one thing: whether the
change's branch has already reached the base branch. No field records "integration started" — the
branch's merge status is the only source of truth, and a field could disagree with it.

**The merge status is decided by three signals, in this order, and by a script — not by prose.**
`check-finish-preflight.sh <worktree> <base-ref> <recorded-merge-base|->` prints one verdict
line and exits 0 whenever it reached a verdict. It exits 2 with no verdict when it cannot read the
worktree at all — an unreadable tree is never a licence to proceed.

**`<base-ref>` is composed as `origin/$BASE`** — the remote-tracking ref, `$BASE` being the bare
name `resolve-base-branch.sh` printed — never the bare local name on its own.

| Verdict | Meaning |
|---------|---------|
| `RUN1` | integrate — the branch has not reached the base branch |
| `RUN2` | archive — merged, and nothing is outstanding |
| `REFUSE` | stop and ask the operator before anything is archived |

On a `REFUSE`, stop before touching anything, report `HEAD`, the base branch and the uncommitted
count, and ask the operator explicitly. On a multi-repo change, run the script once per worktree in
the set found by **Resolving a change's worktrees** below, and proceed to run 2 only when **every**
worktree returns `RUN2`. A resolved set that comes back empty is not "every worktree" — it stops the
run exactly as **Resolving a change's worktrees** requires, never a vacuous `RUN2`.

### Surface foreign staged work before the preflight

Before `check-finish-preflight.sh` runs for any worktree, the run surfaces the foreign staged work
the affected main checkouts carry.

The affected repositories are resolved from the worktree set: for each worktree, the main checkout
`git rev-parse --git-common-dir` resolves, made absolute and physical, deduplicated — the same
resolution the preflight's own main-checkout assertion performs. `check-foreign-staged.sh` runs
once per distinct main checkout, and its header is canonical for the verdict grammar it prints. On
`STAGED-CLEAN` from every repository the run continues into the preflight with nothing more said.
On any `STAGED-FOREIGN`, every repository's listing is shown together and the run stops to ask,
exactly once, shape per **Operator prompts** (`skills/flow-contracts/operator-prompts.md`). **An
exit 2 with no verdict** — a main checkout whose status cannot be read — stops and asks the same
question a `STAGED-FOREIGN` asks, with nothing to list; an inability is never read as
`STAGED-CLEAN`.

**Stop** leaves the change where its state has it with nothing staged, committed, pushed, reset or
stashed by the run; the operator commits, stashes or resets the residue themselves and re-runs.
**Continue** carries the listing into the handoff and proceeds — and relaxes nothing: every later
gate keeps exactly the behavior it already had, the preflight's main-checkout assertion included.
The relay includes the guard's own hand-verification procedure per **Hand-verifying a guard
verdict** (`skills/flow-contracts/pipeline.md`).

In the same pre-run position, over the same distinct main checkouts, the run also surfaces each
checkout's drift from its expected post-merge state — on the repository's default branch with
nothing tracked modified, staged or unmerged. `check-main-checkout-drift.sh` runs once per
distinct main checkout, and its header is canonical for the verdict grammar it prints: a
`DRIFT-BRANCH` line names a checkout on any other branch than the one `refs/remotes/origin/HEAD`
points at, a `DRIFT-DIRTY` line counts tracked entries. On `DRIFT-CLEAN` from every
repository the run continues with nothing more said. On any `DRIFT-BRANCH` or `DRIFT-DIRTY`
line, every repository's findings — this guard's and `check-foreign-staged.sh`'s alike — are
shown together and the run stops to ask the question below exactly once; **an exit 2 with no
verdict** — a checkout whose branch, default branch or status cannot be read — stops and asks the
same way, an inability never reading as `DRIFT-CLEAN`. A checkout clean but behind its default
branch is ordinary not-pulled state and never a finding here; the archive's own refresh step owns
bringing it forward.

> **Main checkouts carry foreign staged work or drift — how should the run proceed?**
> - **Stop — I'll clean it up and re-run** *(default, recommended)*
> - **Continue — leave it in place**

### Run 1 — the branch is not merged

**Check for unfinished work first — before the landing question and before any git action.**
`check-unfinished-work.sh <worktree> <change-name> [canonical-worktree]` prints one verdict line and
exits 0 whenever it reached a verdict. It exits 2 with **no** verdict line when it cannot read the
worktree. Run it once per worktree in the set found by **Resolving a change's worktrees** below —
never a raw read of the state file's `worktrees` map, for the same reason the preflight verdict
above does not read it raw. Pass `[canonical-worktree]` on every call in the run: the one member of
the resolved set whose own `<project>/<spec-root>/changes/<change-name>/tasks.md` exists.

| Verdict | Meaning |
|---------|---------|
| `CLEAR` | nothing outstanding — go straight to the landing question, with no extra prompt |
| `OUTSTANDING` | show the breakdown and offer the three courses of **The three courses** (`skills/flow/unfinished-work-gate.md`) |

A missing verdict line is not a verdict. Treat it exactly as the preflight script's fourth outcome
above: stop and ask the operator. The exit code is checked as well as the line, because a caller
that greps for `CLEAR` in empty output finds nothing.

**Then run `check-visual-verify-dispatched.sh <worktree> <change-name> <recorded-merge-base>`**,
once per worktree in the same resolved set, `<recorded-merge-base>` being the same state-file value
signal 1 above already reads. A `VISUAL-VERIFY-OK` line joins `CLEAR` and folds no further
breakdown in; a `VISUAL-VERIFY-MISSING` line is treated exactly as `OUTSTANDING` — it feeds the same
breakdown and the same three courses below, not a second prompt. Exit 2 (cannot answer) is stop-and-ask,
the same as a missing `check-unfinished-work.sh` verdict line above.

**Load `skills/flow/unfinished-work-gate.md` only when** a worktree reported `OUTSTANDING` or
`VISUAL-VERIFY-MISSING` — the three courses, the filing and the relay are there.

**Check whether the base branch has moved — after the unfinished-work gate, before the landing
question.** `check-base-moved.sh <worktree> <base-ref> <recorded-merge-base|->` prints one verdict
line and exits 0 whenever it reached a verdict; it exits 2 with no verdict line when it cannot read
the worktree. `<base-ref>` is composed as `origin/$BASE`, exactly as the preflight's own is above,
and `<recorded-merge-base>` is the merge base recorded in the state file's `worktrees` map for that
worktree. The three verdicts — `CLEAR`, `MOVED` and `REFUSE` — and the exit contract are the
script's own; see `<agents repo>/scripts/check-base-moved.sh`'s header rather than a copy of them
here. A `MOVED` verdict is relayed with the guard's hand-verification procedure alongside it, per
**Hand-verifying a guard verdict** (`skills/flow-contracts/pipeline.md`).

Run it once per worktree in the set found by **Resolving a change's worktrees** below — never a raw
read of the state file's `worktrees` map, for the same reason the preflight verdict and the
unfinished-work check above do not read it raw. Every worktree's verdict is reported, and none is
prompted: a `MOVED` verdict, overlapping or not, is what **Sync the branch onto the base**
(`skills/flow/sync-onto-base.md`) consumes. A `REFUSE`, an exit 2, or a resolved set that comes back empty stops and asks, exactly
as the preflight verdict above does.

Only then decide, **before any git action**, how the branch should land — the default, the prompt
and its parse are step 2 of `skills/flow/integrate.md`. The answer is never remembered between runs.

**The reshape-commit-route sequence below runs once per worktree in the resolved set, in the order
given by the canonical `link.md`'s `## Merge order`, stopping on the first failure with that
repository's own output.** docs/links.md in the `spectre` repository is canonical for that
section's grammar. A multi-repository change always carries the record: the implement phase
writes it (`skills/flow/implement.md`), naming every repository of the resolved set whether or
not a link created it. A change with no `link.md` has one repository and one trivially-ordered
route, so nothing about the single-repository path changes.

**Before any route commits, reshape the branch.** Run
`reshape-branch.sh <abs-worktree> <name> <recorded-merge-base>`, where `<recorded-merge-base>` is the
merge base recorded in the state file's `worktrees` map for this worktree — the same merge base
**Resolving a change's worktrees** and the finish-preflight verdict above both reference — **or
`<rebased-merge-base>`, for a worktree **Sync the branch onto the base** (`skills/flow/sync-onto-base.md`) rebased**. Every
planning commit on the branch — each commit touching `<project>/spectre/changes/`: the plan-gate,
link, reviewer-dispatch, fix-run and write-in-progress commits of **Planning commits**
(`git-boundaries.md`) and `/flow-plan`'s capture commit — is kept as its own commit, in order, with
its message and author, rebuilt on the merge base. Every per-task and fixup commit `/flow`'s
implement phase made is collapsed back into the working tree, uncommitted, so the two-commit chain
below commits it as one implementation commit on top of the kept planning commits. **Planning
commits are never squashed into one**, at integrate or on any landing route. The script runs
`check-planning-commit-location.sh` first and stops on its verdict. **When the script cannot be
located**, stop and report it; never substitute `reset --soft`, which folds the planning commits
away.

All three routes then commit the work, in **two** commits and never one: the implementation,
subject `<type>(<module>): <what the implementation does>` with `<module>` naming the area the
reshaped diff carries, then whatever `<project>/spectre/changes/` planning delta the kept planning
commits left — this run's narrative and anything the operator edited — subject the fixed literal
`chore(spectre): plan`. The landed branch is therefore: the kept planning commits, the
implementation commit, then that last planning commit.

**Nothing under `<abs-worktree>/.superpowers/sdd/` is committed** — not the rendered ledger and
panel record, not the brainstorm design document. They are worktree-lifetime files, removed with
the worktree at run 2; the store is the terminal record, and run 2 renders from it what the
self-review bundle needs (**Temporary artifacts registry**, `artifacts-registry.md`). The proposal
artifact source stays in the state directory until run 2 removes it; run 1 copies nothing.

The planning path, `<project>/spectre/changes/`, is cleared from the index before the first `add` and excluded from it by
pathspec — the same clearing pass **Git boundaries** (`git-boundaries.md`) gives `/flow`'s implement phase, and
for the same reason: an exclusion cannot retract what an earlier step staged, and at this gate that
step may have been the operator's own `git add`. The second `add` carries no pathspec, which is what
picks that path up; `<abs-worktree>/.superpowers/sdd/` is gitignored, so it never does. The sequence itself — the guarded commits, the skipped-empty rule, the
failure rule and the symlink case — is the chain **The guarded two-commit chain** (`git-boundaries-commit-chain.md`) gives;
`<agents repo>/scripts/commit-split.sh` is what runs it, at both this
call site and the implement phase's PR-exception path.

| Route | Then |
|-------|------|
| **Open a pull request** | push `--force-with-lease` (**Branch backup**, `skills/flow-contracts/git-boundaries.md`); open a PR via `gh` when usable for the host, else print the forge's create-PR URL and ask whether it was opened; record `prUrl` |
| **Merge and push** | push `--force-with-lease`; `classify-untracked.sh <project>/.worktrees/_landing-<name>` first when the landing worktree already exists — the archive pre-flight, **Run 2 — the branch is merged** (`skills/flow-contracts/finish-contract-run2.md`) step 2's class rule, settling its assets through the operator — then `prepare-archive-branch.sh <project>/.worktrees/_landing-<name> <base> <base>`; `git -C <landing-worktree> merge --no-ff spectre/<name>`. **`<base>` is not pushed here, and the landing worktree stays** — run 2's same-invocation continuation archives on top of this local merge and pushes `<base>` once, carrying the merge, the archive and the self-review output together (**Run 2 — the branch is merged**, `skills/flow-contracts/finish-contract-run2.md`, step 10). A merge conflict here means the base moved after the sync: `git -C <landing-worktree> merge --abort`, remove the landing worktree, and re-run **Sync the branch onto the base** (`skills/flow/sync-onto-base.md`) and this route once |
| **Handle it manually** | push the branch `--force-with-lease` only; say plainly what is left to do |

`<archive-branch>` equal to `<base>` itself — the merge-and-push route's own use above — means
"position `<landing-worktree>` on `<base>` itself", and `prepare-archive-branch.sh` accepts it; see
that script's own header. The main checkout is never checked out, staged or committed by this
route — the landing worktree carries the merge in its place, and run 2 the push.

Run 1 ends at `IN_PROGRESS`, and its handoff's last line is `/flow <name>` again.

**Resolve the base branch; never assume it, and never derive it from the current branch.**

```bash
BASE="$(resolve-base-branch.sh "<abs-worktree>")"
```

`<abs-worktree>` is **the apply worktree** — the same one **Reshape the branch** above just
committed from — where `HEAD` is the change's own branch, `spectre/<name>`, by construction; never
the main checkout. The exit contract: `0` resolved, with the name on stdout; `1` a named refusal —
detached `HEAD`, no base resolved, the base equal to the current branch, or a base name that fails
validation; `2` the tree cannot be read — `<abs-worktree>` is missing, unreadable, or not a git
worktree — or `HEAD`'s own ref cannot be read (corrupt or permission-denied); `3` the repository has
no `origin` remote at all.

**Never fall back to `HEAD@{upstream}`.**

If no base branch resolves, **stop and ask**.

**A repository with no remote at all cannot be integrated by this command.** Every route needs a
push, and base resolution needs `origin` — `resolve-base-branch.sh` is where that check lives,
and its exit `3` is how a caller recognises this case. Say exactly that — *"this repository has no
remote, so there is nothing to push to or merge into"* — rather than reporting a base-branch
failure, which sends the operator debugging the wrong thing. Offer to leave the change at
`IN_PROGRESS` with the work staged; there is nothing to lose, because nothing was pushed.

**No verification gate runs before integration.** No tests, no linters, no spec-coverage check. Two exceptions exist, both under **Sync the branch onto the base** (`skills/flow/sync-onto-base.md`): a rebase
that needed conflict resolution runs the whole `## lint` and `## test` lists, and a clean rebase
that changed a file this change also touches runs the scoped re-verification that file is
canonical for.

### Resolving a change's worktrees

The scan that finds the worktrees carrying a change's branch. The preflight verdict, the unfinished-work gate, and run
2's removal all resolve the set through this same procedure.

The set of worktrees is the **keys of the state file's `worktrees` map**. When that is absent or
empty, scan each affected repository:

```bash
git -C "$REPO" worktree list --porcelain \
  | awk '/^worktree /{w=substr($0, 10)} /^branch /{if ($2=="refs/heads/spectre/<name>") print w}'
```

**Never guess a path.** The kickoff creates every worktree at `<project>/.worktrees/<name>`
(**Git boundaries**, `skills/flow-contracts/git-boundaries.md`), but a worktree made by hand or by
an older run can sit anywhere, so the set is read, never composed.

**Here, a resolved set that is still empty means the map was absent or empty *and* the scan found no
worktree on the change's branch in any affected repository.** Per **Resolving a change's worktrees**
(`skills/flow-contracts/worktree-resolution.md`), that is a state the pipeline cannot explain: stop and
report it to the operator, exactly as a `REFUSE`
verdict would, rather than letting a zero-iteration loop read as "every worktree returned `RUN2`" or
"`CLEAR` from every worktree." This applies wherever this procedure is used — the preflight verdict,
the unfinished-work gate and run 2's removal alike.

**The path is taken with `substr`, never `$2`.**

