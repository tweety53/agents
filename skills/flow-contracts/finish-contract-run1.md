# Finish contract — run 1 (the branch is not merged)

**This file is canonical for `/myflow-finish`'s run 1** — the preflight-signal decision, run 1's
procedure, and resolving a change's worktrees.

`/myflow-finish` is the only command that loads this file.

## Finish contract

`/myflow-finish` is a **two-run** command. Which run happens is decided by one thing: whether the
change's branch has already reached the base branch. No field records "integration started" — the
branch's merge status is the only source of truth, and a field could disagree with it.

**The merge status is decided by three signals, in this order, and by a script — not by prose.**
`check-finish-preflight.sh <worktree> <base-ref> <recorded-merge-base|->` prints one verdict
line and exits 0 whenever it reached a verdict. It exits 2 with no verdict when it cannot read the
worktree at all — an unreadable tree is never a licence to proceed.

**`<base-ref>` is composed as `origin/$BASE`** — the remote-tracking ref, `$BASE` being the bare
name `resolve-base-branch.sh` printed — never the bare local name on its own. A bare local branch is
wrong here because the ancestor test resolves whatever ref it is handed, and `resolve-base-branch.sh`'s
fetch refreshes only remote-tracking refs, never local branches; a stale local branch of the same
name would then silently feed the RUN1/RUN2/REFUSE decision, the gate in front of this command's most
destructive step.

| Verdict | Meaning |
|---------|---------|
| `RUN1` | integrate — the branch has not reached the base branch |
| `RUN2` | archive — merged, and nothing is outstanding |
| `REFUSE` | stop and ask the operator before anything is archived |

1. **`HEAD` against the merge base recorded in the state file's `worktrees` map.** No recorded
   value, or one that does not resolve → `REFUSE`; an honest unknown is never inferred. Equal to
   `HEAD` means the branch has no commits of its own → `RUN1`, whatever the ancestor test says.
   Both are answered before the base ref is resolved, so no environmental failure can hide them.
2. **The ancestor test** — `git merge-base --is-ancestor`, and only that. The script consults no PR
   CLI: it must give the same answer on every forge, and git alone already answers this question.
   The base ref is resolved first, so a base ref that does not resolve is its own `REFUSE` rather
   than an accidental `RUN1`.
3. **The worktree's cleanliness.** Merged by ancestry with uncommitted entries → `REFUSE`.
   Merged, distinct from the recorded merge base, and clean → `RUN2`.

**Never substitute a commit count.** `git rev-list --count <base>..HEAD` is zero both for a branch
with no commits and for a branch whose commits have joined the base branch, so it cannot separate the
dangerous state from the correct terminal one, and using it would refuse every legitimate archive.

**Signal 1 precedes signal 2, and that ordering is the point.** A branch with no commits of its own
is an ancestor of every branch, so the ancestor test alone reports *merged* on a branch whose work is
staged and never committed — after which run 2 archives the change and `--force`-removes the worktree
holding all of it.

On a `REFUSE`, stop before touching anything, report `HEAD`, the base branch and the uncommitted
count, and ask the operator explicitly. On a multi-repo change, run the script once per worktree in
the set found by **Resolving a change's worktrees** below, and proceed to run 2 only when **every**
worktree returns `RUN2`. A resolved set that comes back empty is not "every worktree" — it stops the
run exactly as **Resolving a change's worktrees** requires, never a vacuous `RUN2`.

**When the script is absent** — a harness whose repository does not carry it — perform the same three
signals by hand in the same order and say in the handoff that the check was run manually. The check is
never skipped for want of the script. Signal 2 may then be answered by a PR CLI when one is usable
for the host, as in run 2
(`skills/flow-contracts/finish-contract-run2.md`) — that option belongs to the human doing this by hand, never to the
script — but signals 1 and 3 still run, and still run in this order.

### Run 1 — the branch is not merged

**Check for unfinished work first — before the landing question and before any git action.**
`check-unfinished-work.sh <worktree> <change-name> [canonical-worktree]` prints one verdict line and
exits 0 whenever it reached a verdict. It exits 2 with **no** verdict line when it cannot read the
worktree. Run it once per worktree in the set found by **Resolving a change's worktrees** below —
never a raw read of the state file's `worktrees` map, for the same reason the preflight verdict
above does not read it raw. Pass `[canonical-worktree]` on every call in the run: the one member of
the resolved set whose own `<project>/<spec-root>/changes/<change-name>/tasks.md` exists. Without it, a
satellite worktree's call falls back to resolving the link's peer name through `peers`, which cannot
resolve from inside a worktree — see `design.md`'s `guards-take-the-canonical-worktree-path`. A
resolved set that comes back empty stops the run rather than passing as `CLEAR` from every worktree.

| Verdict | Meaning |
|---------|---------|
| `CLEAR` | nothing outstanding — go straight to the landing question, with no extra prompt |
| `OUTSTANDING` | show the breakdown and offer the three courses below |

A missing verdict line is not a verdict. Treat it exactly as the preflight script's fourth outcome
above: stop and ask the operator. The exit code is checked as well as the line, because a caller
that greps for `CLEAR` in empty output finds nothing.

On `OUTSTANDING` the operator is offered **exactly three** courses:

| Course | What run 1 then does |
|--------|----------------------|
| **Stop — I'll finish it first** *(recommended)* | stop, leaving the change at `IN_PROGRESS` with nothing staged, committed or pushed |
| **Continue — integrate anyway** | proceed to the landing question, carrying the outstanding list into the planning commit's message and the handoff |
| **File or join a Jira follow-up, then continue** | put the outstanding items on a follow-up issue — joining an open one where the operator confirms a candidate, otherwise filing a new one — then proceed |

**Stop is marked as the recommendation, and the reason is stated rather than left to be inferred.**
The gate only fires because something really is unfinished, and finishing it is the cheapest of the
three to recover from — Continue is the only course that reaches an irreversible step, and it exists
for work the operator deliberately deferred, which is a judgment only they hold. Marking a
recommendation is not a courtesy here: the planning-gate capability requires every choice a
`/flow*` command offers to name its recommended option, and this prompt is one of them.

There is no fourth course, and in particular none that hands back to `/myflow-do` inline. The filed
issue is labelled and linked per
**Labels on issues the pipeline creates** (`skills/flow-contracts/jira-integration.md`).

**A filing that fails is one skipped-with-reason line, and the run still proceeds** — the same
degradation every other Jira write in this pipeline has, per
**Never blocking** (`skills/flow-contracts/jira-integration.md`). Creation can fail for the usual
reasons (auth, permission, an unknown project key or label) and there may be no tracker configured
at all. None of them changes the operator's answer, which was *continue*: the outstanding list still
reaches the planning commit's message and the handoff, which is where this change requires the
durable record to be. A failed filing is never silently upgraded to **Stop**, and never passes
unmentioned.

**Three more outcomes of that course behave the same way**, and all three belong to
**Follow-up issues** (`skills/flow-contracts/jira-followups.md`) rather than here: the search
that finds a candidate asks the operator to confirm the join before writing to it, a declined
confirmation files a new follow-up instead, and a search that *fails* files nothing and says so. Each
is one line and none of them stops the run or changes the answer already given. That file is
canonical for all of it — including the ordering of a join's three writes and what a partial one
reports.

**What the operator integrated over is recorded where a transcript is not**: the outstanding list
goes into the message of the commit that carries the planning artifacts, and into run 1's handoff.
The signals that produce that list are the script's own.

**Check whether the base branch has moved — after the unfinished-work gate, before the landing
question.** `check-base-moved.sh <worktree> <base-ref> <recorded-merge-base|->` prints one verdict
line and exits 0 whenever it reached a verdict; it exits 2 with no verdict line when it cannot read
the worktree. `<base-ref>` is composed as `origin/$BASE`, exactly as the preflight's own is above,
and `<recorded-merge-base>` is the merge base recorded in the state file's `worktrees` map for that
worktree. The three verdicts — `CLEAR`, `MOVED` and `REFUSE` — and the exit contract are the
script's own; see `<agents repo>/scripts/check-base-moved.sh`'s header rather than a copy of them
here.

Run it once per worktree in the set found by **Resolving a change's worktrees** below — never a raw
read of the state file's `worktrees` map, for the same reason the preflight verdict and the
unfinished-work check above do not read it raw. Every worktree's verdict is reported. The operator
is asked **once**, for the whole change, and **only** when at least one worktree's verdict overlaps
this change's own paths (design.md: `ask-only-on-overlap`, `aggregate-the-multi-repo-ask`) — a
`MOVED` verdict with no overlap is reported and the run continues to the landing question with no
extra prompt. A `REFUSE`, an exit 2, or a resolved set that comes back empty stops and asks, exactly
as the preflight verdict above does.

**When the script is absent** — a harness whose repository does not carry it — reach the same three
verdicts by hand, in the same order, and say in the handoff that the check was run manually. The
check is never skipped for want of the script.

Only then decide, **before any git action**, how the branch should land. Read
`<project>/.flow/project.md`'s `## default landing route` (canonical in
**Project configuration**, `skills/flow-contracts/project-configuration.md`); a resolved default
is taken without asking, stated in the handoff as coming from the configured default rather than an
operator choice. Only an absent or unresolved default falls back to asking:

> **How should this branch land?**
> - **Open a pull request** *(default, recommended)*
> - **Merge and push**
> - **Handle it manually**

Then run to completion without asking again. The answer is never remembered between runs.

**The reshape-commit-route sequence below runs once per worktree in the resolved set, in the order
given by the canonical `link.md`'s `## Merge order`, stopping on the first failure with that
repository's own output.** docs/links.md in the `spectre` repository is canonical for that
section's grammar. A change with no `link.md` has one worktree and one trivially-ordered route, so
nothing about the single-repository path changes.

**Before any route commits, reshape the branch.** Run
`git -C <abs-worktree> reset --soft <recorded-merge-base>`, where `<recorded-merge-base>` is the
merge base recorded in the state file's `worktrees` map for this worktree — the same merge base
**Resolving a change's worktrees** and the finish-preflight verdict above both reference — **or the
this-run-only rebased merge base, for a worktree `skills/flow/integrate.md`'s own in-pipeline
rebase step rebased**; that file is canonical for the exception, not restated here. This
collapses every per-task and fixup commit `/myflow-do` made on the branch back into the working
tree, uncommitted, so the branch carries no history for the two-commit chain below to inherit —
that chain then commits from this reshaped state exactly as it always has.

All three routes first commit the work, in **two** commits and never one: the implementation,
subject `<type>(<module>): <what the implementation does>` with `<module>` naming the area the
reshaped diff carries, then the `<project>/spectre/changes/` planning artifacts and the session records
under `<project>/docs/superpowers/` (the SDD ledger and the review panel record), subject the fixed
literal `chore(spectre): plan and session records`. The ledger is rendered from the store before
staging, by:

```bash
flow record render -change <name> -kind ledger -repo <abs-repo-root>
```

`<abs-repo-root>` is the apply worktree's own root — the tree this run is committing from, never the
main checkout. The panel record is already under that path: `/myflow-do` renders it at panel close,
from the same rows. Act on the command's outcome word per the table under **Rendering the session
records** (`session-records.md`), which is canonical for it and is deliberately not restated here.

**The proposal artifact source is copied here, not rendered.** It is not a record: `/myflow-start`
wrote it to the state directory and it was never in the store, so `flow record render` — which
renders from the store's rows — has no business with it. Run 1 copies it directly, in this same step
and for the same reason the ledger is rendered here: the second `add` below carries no pathspec, so
a copy written under `<project>/docs/superpowers/` before staging lands in the planning commit and
reaches the base branch with it.

```bash
mkdir -p "<project>/docs/superpowers/artifacts"   # only after the checks below
cp "<state-dir>/<name>-proposal-artifact.html" \
   "<project>/docs/superpowers/artifacts/<YYYY-MM-DD>-<name>.html"
```

- **Validate `<name>` against `^[A-Za-z0-9][A-Za-z0-9._-]*$` before either path is built.** It is the
  rule `records.Destination` (`<agents repo>/stats/internal/records/render.go`) carries, for the
  reason its own comment states: a glob metacharacter in a name once matched and overwrote a
  *different* change's preserved copy. A name that fails is a caller mistake — report it and copy
  nothing.
- **The destination directory must resolve inside the repository root before anything is written
  through it.** `<project>/docs/superpowers/artifacts/` is a tracked, pull-request-editable path: a
  symlink there puts the copy outside the repository, and run 1 commits and pushes what it wrote.
  This is `records.Destination`'s second protection, and it applies to a `cp` for the same reason.
- **The source path must resolve inside the state directory it belongs to, checked before the copy
  runs.** Resolve `<state-dir>/<name>-proposal-artifact.html` following every symlink at every path
  component, and require the result to stay under `<state-dir>`. A source resolving anywhere else is
  **refused and reported**, never silently skipped — a refusal is not an absence, and reporting one
  as the other hides a planted path as a change that simply had no artifact to copy. This is the
  write-side half of the boundary pair `check_boundary`
  (`<agents repo>/scripts/gather-self-review-context.sh`) applies on the read side, and it is the
  one protection the retired copying script carried that an inline `cp` does not inherit: `cp`
  follows a symlinked source, so a stale or planted state directory puts arbitrary content under
  `<project>/docs/superpowers/artifacts/`, which run 1 then commits and pushes.
- **Reuse an existing `<project>/docs/superpowers/artifacts/<date>-<name>.html`** rather than writing
  a second dated file, so a re-run overwrites in place — the rule `records.Destination` applies to
  the rendered records, applied here to the copy.
- **A change with no artifact at that source is skipped and said so, never failed.** Every
  `/myflow-fast` run is such a change: it publishes no proposal artifact, so there is nothing to
  copy and nothing wrong.

This copy is what makes the `Proposal artifact source` row in **Temporary artifacts registry**
(`artifacts-registry.md`) reachable at all: its run-2 deletion is conditional on a preserved copy existing, and
this step is what produces one.

Those two planning paths are cleared from the index before the first `add` and excluded from it by
pathspec — the same clearing pass **Git boundaries** (`git-boundaries.md`) gives `/myflow-do`, and
for the same reason: an exclusion cannot retract what an earlier step staged, and at this gate that
step may have been the operator's own `git add`. The second `add` carries no pathspec, which is what
picks the two paths up. The sequence itself — the guarded commits, the skipped-empty rule, the
failure rule and the symlink case — is the chain **Git boundaries** (`git-boundaries.md`) gives, and is
not written out a second time here; `<agents repo>/scripts/commit-split.sh` is what runs it, at both this
call site and `/myflow-do`'s PR-exception path.

| Route | Then |
|-------|------|
| **Open a pull request** | push; open a PR via `gh` when usable for the host, else print the forge's create-PR URL and ask whether it was opened; record `prUrl` |
| **Merge and push** | push; merge into the base branch; push that |
| **Handle it manually** | push the branch only; say plainly what is left to do |

Run 1 ends at `IN_PROGRESS`, and its handoff's last line is `/myflow-finish <name>` again.

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

**Never fall back to `HEAD@{upstream}`.** `/myflow-finish` runs inside the apply worktree, where
`HEAD` *is* `spectre/<name>` — so that fallback resolves to the change's **own** upstream, making
the merge check `spectre/<name>` vs `origin/spectre/<name>`, which is true the moment the branch
is pushed. That silently reports an unmerged change as merged, and run 2 then archives it and
deletes its worktree. `resolve-base-branch.sh` is where this rule is now enforced: it never
consults `HEAD@{upstream}`, and its assertion that `BASE` differs from the current branch is
unconditional, which is what makes that class of misresolution impossible rather than merely
unlikely.

If no base branch resolves, **stop and ask**. An unresolvable base is an honest unknown; a guessed
one is a wrong answer at the only irreversible step.

**When the script is absent** — a harness whose repository does not carry it — resolve the base
branch by hand, in the same order, and say in the handoff that the resolution was done manually. The
guard is never skipped for want of the script. Run the wrapped fetch first — bounded and
credential-free, exactly as the guard's own header describes, so an unreachable remote refuses
quickly rather than hanging — then read `refs/remotes/origin/HEAD`, falling back to `git remote show
origin`'s reported `HEAD branch` only when that is empty. Once a candidate name is in hand, apply the
guard's assertions in order and by hand, never accepting a guess in place of any of them: refuse a
detached `HEAD`, refuse when no name resolved, refuse when the resolved name equals the current
branch, and refuse the resolved name unless it matches this exact shape: the first character is one
of `[A-Za-z0-9._]` — so a leading `-` a downstream git call would read as an option is refused, and
so is a leading `/` — and every character in the name, start to end, is one of `[A-Za-z0-9._/-]`,
which is what rules out control characters and anything else outside that set. That header is the
authority for the exact commands and their order; this paragraph does not repeat what it already
states. The character rule itself is stated here in full, not cited, because this paragraph's own
precondition is the script's absence — a fallback the operator applies by hand when the script is
absent cannot defer them to a file that, by the same precondition, is not there to read.

**A repository with no remote at all cannot be integrated by this command.** Every route needs a
push, and base resolution needs `origin` — `resolve-base-branch.sh` is where that check now lives,
and its exit `3` is how a caller recognises this case. Say exactly that — *"this repository has no
remote, so there is nothing to push to or merge into"* — rather than reporting a base-branch
failure, which sends the operator debugging the wrong thing. Offer to leave the change at
`IN_PROGRESS` with the work staged; there is nothing to lose, because nothing was pushed.

**No verification gate runs before integration.** No tests, no linters, no spec-coverage check.
Correctness was established during `/myflow-do` — TDD per task, per-task review, the final review
panel — and by the human gate. Re-running it here would repeat finished work immediately before
the one irreversible step. One exception exists: `skills/flow/integrate.md`'s own in-pipeline
rebase step runs a scoped re-verification, over the rebase's overlap set only, when a clean rebase
actually changed a file this change also touches — that file is canonical for the exception's
scope.

### Resolving a change's worktrees

The scan that finds the worktrees carrying a change's branch. This is `/myflow-finish`'s own
application of the rule stated once under **Resolving a change's worktrees**
(`skills/flow-contracts/worktree-resolution.md`) — that a step needing "the worktrees" resolves the set
rather than reading the state file's `worktrees` map directly, and that a resolved set which comes
back empty is never a vacuous pass. That rule and the commands it binds are not restated here; what
follows is specific to `/myflow-finish`: the preflight verdict, the unfinished-work gate, and run
2's removal all resolve the set through this same procedure.

The set of worktrees is the **keys of the state file's `worktrees` map**. When that is absent or
empty, scan each affected repository:

```bash
git -C "$REPO" worktree list --porcelain \
  | awk '/^worktree /{w=substr($0, 10)} /^branch /{if ($2=="refs/heads/spectre/<name>") print w}'
```

**Never guess a path.** Worktree layout differs per repository — this repo keeps worktrees under
`<agents repo>/.worktrees/` (git-ignored, per `superpowers:using-git-worktrees`), which is not where every
repository this pipeline is installed into keeps them.

**Here, a resolved set that is still empty means the map was absent or empty *and* the scan found no
worktree on the change's branch in any affected repository.** Per **Resolving a change's worktrees**
(`skills/flow-contracts/worktree-resolution.md`), that is a state the pipeline cannot explain: stop and
report it to the operator, exactly as a `REFUSE`
verdict would, rather than letting a zero-iteration loop read as "every worktree returned `RUN2`" or
"`CLEAR` from every worktree." This applies wherever this procedure is used — the preflight verdict,
the unfinished-work gate and run 2's removal alike.

**The path is taken with `substr`, never `$2`.** `worktree list --porcelain` emits it raw, so a
field reference truncates any path containing a space at the first one: fed
`worktree /tmp/my worktree` it yields `/tmp/my`, and the run then `--force`-removes a path that is
not the worktree, or fails having named the wrong one. `10` is one past the length of the literal
`worktree ` prefix. The branch on the next line is a ref name and cannot contain a space, so `$2` is
right for it. `<agents repo>/scripts/check-cleanup-complete.sh` parses the same stream the same way — the guard
and the snippet it verifies must not disagree, or the wrong one gets copied next.

