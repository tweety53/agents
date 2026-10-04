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
| `RUN2` | clean up — merged, and nothing is outstanding |
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
resolution the preflight's own stray-worktree assertion performs. `check-foreign-staged.sh` runs
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
gate keeps exactly the behavior it already had.
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
branch is ordinary not-pulled state and never a finding here; run 2's refresh step owns
bringing it forward.

> **Main checkouts carry foreign staged work or drift — how should the run proceed?**
> - **Stop — I'll clean it up and re-run** *(default, recommended)*
> - **Continue — leave it in place**

### Run 1 — the branch is not merged

**An archive a stopped run 1 left uncommitted is undone first**, in the canonical worktree, before
anything below: when `<project>/spectre/changes/archive/<name>/` exists there and `<merge-base>..HEAD`
holds no `chore(spectre): archive <name>` commit — a run 1 that stopped between `spectre archive` and
`commit-archive.sh`, or between the parent's call and a `<name>-fix-N` sibling's — `git mv` that
directory, and every `<project>/spectre/changes/archive/<name>-fix-N/` beside it, back under
`<project>/spectre/changes/`. **When `<project>/spectre/changes/<name>/` also exists, stop and ask** — `git mv` would nest the archive inside it, hiding the fix run's tasks from the gate; the operator removes the stale live copy and re-runs. The change is then not archived, and this run archives it as a first
run 1 does: any fix run's work is gated, reshaped and committed by the two-commit chain, and the
move lands after it as its own archive commit, never riding `chore(spectre): plan`.

**Whether the change is already archived is decided once per change, in the canonical
repository, and holds for every worktree in the set.** It is archived when the canonical
worktree's `<project>/spectre/changes/archive/<name>/` exists and its
`<project>/spectre/changes/<name>/` does not — an earlier run 1 ran `spectre archive` and stopped
somewhere after it, and any fix run since wrote into the archived directory (**A change's
directory**, `skills/flow-contracts/pipeline.md`).

**An archived re-run's base** is the newest commit in `<merge-base>..HEAD` whose subject is
`docs(self-review): <name> self-review context bundle`, else the newest whose subject is
`chore(spectre): archive <name>`, else `<merge-base>` itself — a satellite worktree's range holds
neither — `<merge-base>` being the merge base the reshape below would otherwise take. Everything at
or below it is kept untouched — a reshape from that merge base itself would fold the archive and
bundle commits into the implementation commit. The re-run **carries new work** when a commit sits
above the canonical worktree's base — every fix run commits its plan there, whichever repositories
its tasks touched — or any worktree in the set holds an uncommitted tracked change, judged at the
unfinished-work gate, before this run's own narrative append.

A re-run of an archived change skips `spectre archive`. **Without new work** it also skips, before
it, the unfinished-work gate, the reshape and the two-commit chain — `spectre archive` already
refused any unchecked task, and nothing above the base needs committing. **With new work** it
skips none of them: the unfinished-work gate reads the archived `tasks.md`, and the reshape takes
the archived re-run's base as its `<merge-base>` argument, so the planning commits above the base
are kept, the task and fixup commits above it collapse into a new implementation commit, and
`chore(spectre): plan` carries the archived directory's planning delta. Everything else runs
either way: the base-moved check and **Sync the branch onto the base**
(`skills/flow/sync-onto-base.md`), `commit-archive.sh`, which prints
`ARCHIVE-NOTHING-STAGED` once the archive commit exists and nothing it copies has changed, the
bundle step, and the route.

**Check for unfinished work first — before the landing question and before any git action but the undo of an uncommitted archive above.**
`check-unfinished-work.sh <worktree> <change-name> [canonical-worktree]` prints one verdict line and
exits 0 whenever it reached a verdict. It exits 2 with **no** verdict line when it cannot read the
worktree. Run it once per worktree in the set found by **Resolving a change's worktrees** below —
never a raw read of the state file's `worktrees` map, for the same reason the preflight verdict
above does not read it raw. Pass `[canonical-worktree]` on every call in the run: the one member of
the resolved set whose own change directory (**A change's directory**,
`skills/flow-contracts/pipeline.md`) holds `tasks.md`.

| Verdict | Meaning |
|---------|---------|
| `CLEAR` | nothing outstanding — go straight to the landing question, with no extra prompt |
| `OUTSTANDING` | show the breakdown and offer the three courses of **The three courses** (`skills/flow/unfinished-work-gate.md`) |

A missing verdict line is not a verdict. Treat it exactly as the preflight script's fourth outcome
above: stop and ask the operator. The exit code is checked as well as the line, because a caller
that greps for `CLEAR` in empty output finds nothing.

**Then run `check-visual-verify-dispatched.sh <worktree> <change-name> <recorded-merge-base>`**,
once per worktree in the same resolved set, `<recorded-merge-base>` being the same state-file value
the preflight above already reads. A `VISUAL-VERIFY-OK` line joins `CLEAR` and folds no further
breakdown in; a `VISUAL-VERIFY-MISSING` line is treated exactly as `OUTSTANDING` — it feeds the same
breakdown and the same three courses (**The three courses**, `skills/flow/unfinished-work-gate.md`), not a second prompt. Exit 2 (cannot answer) is stop-and-ask,
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

Only then decide, **before any git action but the undo and the sync above**, how the branch should land — the default, the prompt
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
the worktree at run 2; the store is the terminal record, and **Save the self-review context
bundle** below renders from it what the bundle needs (**Temporary artifacts registry**,
`artifacts-registry.md`). The proposal
artifact source stays in the state directory until run 2 removes it; run 1 copies nothing.

The planning path, `<project>/spectre/changes/`, is cleared from the index before the first `add` and excluded from it by
pathspec — the same clearing pass **Git boundaries** (`git-boundaries.md`) gives `/flow`'s implement phase, and
for the same reason: an exclusion cannot retract what an earlier step staged, and at this gate that
step may have been the operator's own `git add`. The second `add` carries no pathspec, which is what
picks that path up; `<abs-worktree>/.superpowers/sdd/` is gitignored, so it never does. The sequence itself — the guarded commits, the skipped-empty rule, the
failure rule and the symlink case — is the chain **The guarded two-commit chain** (`git-boundaries-commit-chain.md`) gives;
`<agents repo>/scripts/commit-split.sh` is what runs it, at both this
call site and the implement phase's PR-exception path.

### Archive on the change branch

After the two-commit chain and before any route pushes, in the canonical apply worktree on
`spectre/<name>` — the planning, archive and self-review commits land on the same branch, and so
the same pull request, as the code:

1. **Archive the change** — `spectre archive <name>` moves it into
   `<project>/spectre/changes/archive/<name>/`. **The archived leaf carries no date prefix**.
   **One call per change, parent and sub-change alike.** A `<name>-fix-N` sub-change is a flat
   sibling under `<project>/spectre/changes/`, never a directory inside its parent — `spectre new`
   refuses an id that is not a single flat directory name — so the parent's call cannot reach it and
   each sub-change is archived by its own call in this same step. Never left behind, never archived
   alone. **There is nothing to sync into
   `<project>/spectre/specs/` first**.
2. **Commit the archive on `spectre/<name>`**, as its own commit after the two-commit chain. A
   finished change never leaves the archive move uncommitted in the working tree.

   **The staging is `git add -A`, and this commit's diff is verified scoped to
   `<project>/spectre/changes/`
   before it is made** — `check-archive-scope.sh <canonical-worktree> "spectre/changes/"`, run between
   the add and the commit. A `SCOPE-VIOLATION` refuses the commit and leaves the change at `IN_PROGRESS`
   rather than let a stray path land on `spectre/<name>` unremarked. **An exit 2 with
   nothing on stdout** — a worktree that is not a readable git worktree, or no
   allowed-prefix given — refuses the commit the same way: the guard's header is
   explicit that an inability to answer is never reported as a verdict, so an unreadable tree is
   never read as `SCOPE-OK`.

   **Before the `git add -A`, the rendered ledger and panel record are preserved into this
   commit.** The canonical apply worktree's `<abs-worktree>/.superpowers/sdd/ledgers/<name>.md`
   and `<abs-worktree>/.superpowers/sdd/reviews/<name>-panel.md` are copied into
   `<project>/spectre/changes/archive/<name>/` as `ledger.md` and `panel.md` — each when
   present, an absent file copying nothing — where they ride the archive commit under the scope
   the check above verifies.

   **Before any of that, the Done-when cross-check refuses the commit** — `check-done-when-paths`
   runs in-process on the worktree, printing one `DONE-WHEN-PATH: <path> — <file>` line
   per path a `## Done when` section of the worktree's tracked markdown names that the index does
   not track, and its exit 1 refuses the commit exactly as a `SCOPE-VIOLATION` does, leaving the
   change at `IN_PROGRESS` — the archive's answer to a done criterion satisfied by uncommitted
   files, whose authoring side is the design rule of **The checklist**
   (`skills/flow/brainstorm-planner.md`), and an exit 2 refuses the commit too, never read as a
   pass.

### Save the self-review context bundle

Then, still on `spectre/<name>` in the canonical apply worktree, before any route pushes. Self-review
is always deferred: no reasoning pass and no prompt run here. The session fetches the
bundle with
`flow self-review bundle -change <name>`: flowd assembles the whole bundle — the ledger and the
panel record rendered from the store, or, when the store yields no render for that record, read
from the copies the archive commit carries, the archived `tasks.md`,
`design.md` and `narrative.md`
read out of `spectre/<name>` in the repository the command resolves from its
own location (the main checkout its working directory sits in — the store carries no repository
roots for the pipeline's changes), and the `git log --stat` of every implementation, planning and
archive commit of the change, oldest first — and prints it as one Markdown document. A source that is absent is reported
`skipped: <source> (absent)` inside the bundle, never fatal.
The session appends `## Session narrative` (one paragraph it writes for run 1 itself; the
archived `narrative.md` is already a bundle section, and a change predating the narrative rule
has the bundle report it skipped), writes the whole to
`<project>/docs/self-review/<name>-context.md` in the worktree, and
commits it on `spectre/<name>` with subject `docs(self-review): <name> self-review context
bundle`. An already-committed bundle is skipped, not rewritten — save on an archived re-run with
new work (**Run 1 — the branch is not merged** above), which regenerates it as a new commit. The pass then runs in `/flow-self-review <name>`
(`skills/flow-self-review/SKILL.md`), canonical for the six angles, what is fixed and what may
be filed, the filing-and-rating prompt and the report, which deletes the bundle in its report commit. A
deferred pass covers what the bundle holds and nothing beyond it — the report's
`**Deferred:**` line states that.

### The routes

Every route runs from the apply worktree; none touches the main checkout or any other worktree.

| Route | Then |
|-------|------|
| **Open a pull request** | push `--force-with-lease` (**Branch backup**, `skills/flow-contracts/git-boundaries.md`); open a PR via `gh` when usable for the host, else print the forge's create-PR URL and ask whether it was opened; record `prUrl`. The PR carries the code, planning, archive and bundle commits. A `prUrl` the state file already records is that PR: the push updates it, and none is opened |
| **Merge and push** | push `--force-with-lease` (**Branch backup**, `skills/flow-contracts/git-boundaries.md`), then `git -C <apply-worktree> push origin HEAD:<base>`, `HEAD` being `spectre/<name>` there — a fast-forward, since the branch already sits on `origin/<base>`: the base had not moved, or **Sync the branch onto the base** rebased it there. A rejected push means the base moved after the sync: re-resolve the base with `resolve-base-branch.sh`, which fetches, re-run **Sync the branch onto the base** (`skills/flow/sync-onto-base.md`) — the archive and the bundle, already committed, ride the rebase — and push both once more; a second rejection stops and reports, leaving the change at `IN_PROGRESS`. Run 2 then continues in the same invocation |
| **Handle it manually** | push the branch `--force-with-lease` only; say plainly what is left to do |

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

