# Finish contract — run 1 (the branch is not merged)

**This file is canonical for `/flow`'s integrate run** — the preflight-signal decision, run 1's
procedure, and resolving a change's worktrees.

bare `/flow` is the only command that loads this file whole.

## Finish contract

bare `/flow` is a **two-run** command. Which run happens is decided by one thing: whether the
change's branch has already reached the base branch. No field records "integration started" — the
branch's merge status is the only source of truth, and a field could disagree with it.

**The merge status is decided by three signals, in this order, and by a script — not by prose.**
`flow-guard check-finish-preflight <worktree> <base-ref> <recorded-merge-base|->` prints one verdict
line — and, after a `RUN1`, the base-moved `CLEAR` / `MOVED` / `REFUSE` line about the
same worktree — and exits 0 whenever it reached a verdict. It exits 2 with no verdict when it cannot read the
worktree at all — an unreadable tree is never a licence to proceed.

**`<base-ref>` is composed as `origin/$BASE`** — the remote-tracking ref, `$BASE` being the bare
name `flow-guard resolve-base-branch` printed — never the bare local name on its own.

| Verdict | Meaning |
|---------|---------|
| `RUN1` | integrate — the branch has not reached the base branch |
| `RUN2` | clean up — merged, and nothing is outstanding |
| `REFUSE` | stop and ask the operator before anything is dropped |

On a `REFUSE`, stop before touching anything, report `HEAD`, the base branch and the uncommitted
count, and ask the operator explicitly. On a multi-repo change, run the script once per worktree in
the set found by **Resolving a change's worktrees** below, and proceed to run 2 only when **every**
worktree returns `RUN2`. A resolved set that comes back empty is not "every worktree" — it stops the
run exactly as **Resolving a change's worktrees** requires, never a vacuous `RUN2`.

### Surface main-checkout drift before the preflight

**There must be no staged or dirty work in a main checkout — everything is done in worktrees.**
Before `flow-guard check-finish-preflight` runs for any worktree, the run surfaces each affected
main checkout's drift from its expected post-merge state — on the repository's default branch with
nothing tracked modified, staged or unmerged and no untracked file.

The affected repositories are resolved from the worktree set: for each worktree, the main checkout
`git rev-parse --git-common-dir` resolves, made absolute and physical, deduplicated.
`flow-guard integrate-change prepare` runs its main-checkout drift check (`stats/internal/guard/maincheckoutdrift.go`) once per distinct main checkout, and its header is
canonical for the verdict grammar it prints: a `DRIFT-BRANCH` line names a checkout on any other
branch than the one `refs/remotes/origin/HEAD` points at, a `DRIFT-DIRTY` line counts
entries, staged and untracked ones included. On `DRIFT-CLEAN` from every repository the run continues with
nothing more said. **A clean checkout on another branch is switched back, never asked about.** When
a repository's only finding is `DRIFT-BRANCH` — no `DRIFT-DIRTY` printed —
`flow-guard integrate-change prepare` runs `git -C <main-checkout> switch <default>`, prints one
`DRIFT-SWITCHED: <main-checkout> — <branch> -> <default>, <branch> kept at <sha> (<n> commits
ahead of <default>)` line, and re-runs that check, continuing only on
`DRIFT-CLEAN`. The switch moves no commit: `<branch>` keeps every commit it carried, and a live
session that had the checkout on it finds its HEAD moved, which the printed line names. A clean checkout on a base the change records (`branch.spectre/<name>.flowBase`) is neither
switched nor stopped: run 2's refresh brings it forward. Three shapes are never switched, and stop
as below: a detached HEAD, whose commits sit on no branch; a switch git refuses — the default branch checked out in
another worktree, an untracked file it would overwrite, no default branch to switch to, a locked
index — which leaves the checkout as it was and adds a `DRIFT-SWITCH-REFUSED` line to the
listing; and a re-run that does not print `DRIFT-CLEAN`. Each repository is judged on its own: one
repository's switch never waits on another's question. On any other `DRIFT-BRANCH` line, or any
`DRIFT-DIRTY` line, every repository's findings are shown together and the run stops at
`STOP: main-checkout`, leaving the change where its state has it with nothing staged, committed,
pushed, reset or stashed by the run; the operator commits, stashes or resets the residue
themselves and re-runs. The relay includes the guard's own hand-verification procedure per
**Hand-verifying a guard verdict** (`skills/flow-contracts/pipeline.md`). **An exit 2 with no
verdict** — a checkout whose branch, default branch or status cannot be read — stops the
same way, an inability never reading as `DRIFT-CLEAN`. A checkout clean but behind its default
branch is ordinary not-pulled state and never a finding here; run 2's refresh step owns
bringing it forward.

### Run 1 — the branch is not merged

**Check for unfinished work first — before the landing question and before any git action.**
`flow-guard check-unfinished-work <worktree> <change-name> [canonical-worktree] [<worktree>=<merge-base>]...` prints one verdict line and
exits 0 whenever it reached a verdict. It exits 2 with **no** verdict line when it cannot read the
worktree. Run it once per worktree in the set found by **Resolving a change's worktrees** below —
never a raw read of the state file's `worktrees` map, for the same reason the preflight verdict
above does not read it raw. Pass `[canonical-worktree]` on every call in the run: the one member of
the resolved set whose own change directory (**A change's directory**,
`skills/flow-contracts/pipeline.md`) holds `tasks.md`.
Pass every worktree of the resolved set as `<worktree>=<merge-base>`, its merge base `git -C <worktree> merge-base HEAD origin/<base>`, on every call.
A call passing none gets its own worktree's range derived the same way, from the base
`flow-guard resolve-base-branch` names; the guard's header (`<agents repo>/scripts/flow-guard check-unfinished-work`)
is canonical for the inline-verify signal it folds in.

| Verdict | Meaning |
|---------|---------|
| `CLEAR` | nothing outstanding — go straight to the landing question, with no extra prompt |
| `OUTSTANDING` | finish the outstanding work and re-run the guard, per **Finish it — never a prompt** (`skills/flow/unfinished-work-gate.md`) |

A missing verdict line is not a verdict. Treat it exactly as the preflight script's fourth outcome
above: stop and ask the operator. The exit code is checked as well as the line, because a caller
that greps for `CLEAR` in empty output finds nothing.

**Load `skills/flow/unfinished-work-gate.md` only when** a worktree reported `OUTSTANDING` — finishing the work is there.

**The base-moved verdict is the preflight's second line, not a call of its own.** After a `RUN1`
the preflight prints the base-moved `CLEAR`, `MOVED` or `REFUSE` line for the same worktree,
the three verdicts and the exit contract being the header's of `stats/internal/guard/basemoved.go`
rather than a copy of them here; `<base-ref>` is composed as `origin/$BASE`, exactly as
above, and `<recorded-merge-base>` is the merge base recorded in the state file's `worktrees` map
for that worktree. A `MOVED` verdict is relayed with the guard's hand-verification procedure
alongside it, per **Hand-verifying a guard verdict** (`skills/flow-contracts/pipeline.md`). Every
worktree's verdict is reported, and none is prompted: a `MOVED` verdict, overlapping or not, is
what **Sync the branch onto the base** (`skills/flow/sync-onto-base.md`) consumes, after the
unfinished-work gate. A `REFUSE`, an exit 2, or a resolved set that comes back empty stops and
asks, exactly as the preflight verdict above does.

Only then decide, **before any git action but the sync above**, how the branch should land — the default, the prompt
and its parse are step 2 of `skills/flow/integrate.md`. The answer is never remembered between runs.

**The reshape-commit-route sequence below runs once per worktree in the resolved set, in the order
given by the canonical `link.md`'s `## Merge order`, stopping on the first failure with that
repository's own output.** docs/links.md in the `spectre` repository is canonical for that
section's grammar. A multi-repository change always carries the record: the implement phase
writes it (`skills/flow/implement.md`), naming every repository of the resolved set whether or
not a link created it. A change with no `link.md` has one repository and one trivially-ordered
route, so nothing about the single-repository path changes.

**Before any route commits, reshape the branch.** `flow-guard integrate-change commit` runs
`git -C <abs-worktree> reset --soft <recorded-merge-base>`, where `<recorded-merge-base>` is the
merge base recorded in the state file's `worktrees` map for this worktree — the same merge base
**Resolving a change's worktrees** and the finish-preflight verdict above both reference — **or
`<rebased-merge-base>`, for a worktree **Sync the branch onto the base** (`skills/flow/sync-onto-base.md`) rebased
or whose base-moved `CLEAR` named `merge base now <sha>`**: every commit the run made on the branch — task and fixup commits alike — folds
back into the index, uncommitted.

**A re-run reshapes from past the work already landed in its fixed form**: `prepare` writes as the
reshape base the newest commit in the merge base's range carrying `Flow-Key: flow.commit-two`
(run 1's implementation commit), else the merge base.

All three routes then commit the work as the implementation commit, subject
`<type>(<module>): <what the implementation does>` with `<module>` naming the area the reshaped
diff carries. It carries, beside `Flow-Key: flow.commit-two`, the `Task-Id: <n>` trailer of every
commit the reshape squashed. `<project>/spectre/changes/` is gitignored, so the change directory lands in no
commit.

**Nothing under `<abs-worktree>/.superpowers/sdd/` is committed** — not the rendered ledger and
panel record, not the brainstorm design document. They are worktree-lifetime files, removed with
the worktree at run 2; the store is the terminal record, and **Run the self-review pass** below
renders from it what the bundle needs (**Temporary artifacts registry**,
`artifacts-registry.md`). The proposal
artifact source stays in the state directory until run 2 removes it; run 1 copies nothing.

The sequence itself — the guarded commits, the skipped-empty rule, the failure rule and the
symlink case — is the chain **The guarded commit chain** (`git-boundaries-commit-chain.md`) gives.

**Push every worktree's `spectre/<name>` `--force-with-lease`**, in `## Merge order`, once the
implementation commit lands — the reshape's push (**Branch backup**, `git-boundaries.md`).

### The routes

Every route runs from the apply worktree; none touches the main checkout or any other worktree.

| Route | Then |
|-------|------|
| **Open a pull request** | push `--force-with-lease` (**Branch backup**, `skills/flow-contracts/git-boundaries.md`); open a PR via `gh` when usable for the host, else print the forge's create-PR URL and ask whether it was opened; record `prUrl`. The PR carries the code. A `prUrl` the state file already records is that PR: the push updates it, and none is opened |
| **Merge and push** | push `--force-with-lease` (**Branch backup**, `skills/flow-contracts/git-boundaries.md`), then `git -C <apply-worktree> push origin HEAD:<base>`, `HEAD` being `spectre/<name>` there — a fast-forward, since the branch already sits on `origin/<base>`: the base had not moved, or **Sync the branch onto the base** rebased it there. A rejected push means the base moved after the sync: re-resolve the base with `flow-guard resolve-base-branch`, which fetches, re-run **Sync the branch onto the base** (`skills/flow/sync-onto-base.md`) and push both once more; a second rejection stops and reports, leaving the change at `IN_PROGRESS`. The self-review pass then runs, and run 2 continues in the same invocation |
| **Handle it manually** | push the branch `--force-with-lease` only; say plainly what is left to do |

Run 1 ends at `IN_PROGRESS`, and its handoff ends on `/flow <name>`
again.

### Run the self-review pass

After the route has landed, in the canonical apply worktree on `spectre/<name>`. The session fetches the bundle with
`flow self-review bundle -C <canonical-worktree> -change <name>`: flowd assembles the whole
bundle — the ledger and the panel record rendered from the store, the change directory's
`tasks.md`, `design.md`, `narrative.md`, `visual-verification.md` and `live-verification.md`
read from the worktree on disk, and the `git log --stat` of the branch's first-parent
`<flowLandBase>..HEAD` range when that key is recorded, else `merge-base..HEAD`, followed by each peer the `link.md`'s `## Merge order` names, its
`spectre/<name>` commits against its base under a `### <peer>` heading — and prints it as one
Markdown document. Run inside the worktree, the command also writes it to
`<abs-worktree>/.superpowers/sdd/<name>/self-review-context.md`, the file run 2's
`flow change save` keeps. Its `visual-verification.md` keeps every heading and
drops the table rows under each heading naming a matrix — any level, `matrix` in its text, any
case — replaces each section whose heading opens with bands, seams or ink, through the next heading of
its level or higher, with one line naming the lines it held, its heading and the record's path,
and its `git log --stat` lines leave out every `*.png` path. A source that is absent is reported
`skipped: <source> (absent)` inside the bundle, never fatal.
The session writes that document to `<sr-dir>/bundle.md` (**2. Read the bundle and run the seven angles**, `skills/flow-self-review/SKILL.md`), never printing it — it is never committed.

Over that bundle the session runs steps 2–5 of `/flow-self-review`
(`skills/flow-self-review/SKILL.md`), canonical for the seven angles, the fix and its
dispatch pairs, where the fixes land (always the `self-review-<name>` loop branch, landed by `flow-guard pipeline-fix land`, never onto the change's branch), the filing and the store rows — on every route, before cleanup removes the worktrees.
No report is written: the findings live in flowd. The session then ends the stage with
`flow stage end -command '/flow' -stage flow.self-review -outcome completed <name>`; `cleanup` checks that row
(`flow stage completed -C <main> -stage flow.self-review <name>`) and stops at
`STOP: self-review-report` until it holds. The run's final report names the branch and shas.

A completed `flow.self-review` stage skips the pass on a re-run, so no finding is recorded or
offered twice. So does a completed `self-review-<name>-pass` row while the loop branch `self-review-<name>` exists, the re-run resuming the fix from `<sr-dir>/findings.md` (step 2 of `/flow-self-review`).
A pass covers what the bundle holds and nothing beyond it.

**Resolve the base branch; never assume it, and never derive it from the current branch.**

```bash
BASE="$(flow-guard resolve-base-branch --fetch "<abs-worktree>")"
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
push, and base resolution needs `origin` — `flow-guard resolve-base-branch` is where that check lives,
and its exit `3` is how a caller recognises this case. Say exactly that — *"this repository has no
remote, so there is nothing to push to or merge into"* — rather than reporting a base-branch
failure, which sends the operator debugging the wrong thing. Offer to leave the change at
`IN_PROGRESS` with the work staged; there is nothing to lose, because nothing was pushed.

**No verification gate runs before integration.** No tests, no linters, no spec-coverage check. Two exceptions exist, all under **Sync the branch onto the base** (`skills/flow/sync-onto-base.md`): a rebase
that needed conflict resolution runs the whole `## lint` and `## test` lists, and every clean rebase runs the whole `## lint` list.

### Resolving a change's worktrees

The scan that finds the worktrees carrying a change's branch. The preflight verdict, the unfinished-work gate, and run
2's removal all resolve the set through this same procedure.

The set of worktrees is the **keys of the state file's `worktrees` map**. When that is absent or
empty, scan each affected repository:

```bash
git -C "$REPO" worktree list --porcelain \
  | awk '/^worktree /{w=substr($0, 10)} /^branch /{if ($2=="refs/heads/spectre/<name>") print w}'
```

**Never guess a path.** The kickoff creates every worktree at `<project>-worktrees/<name>`
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

