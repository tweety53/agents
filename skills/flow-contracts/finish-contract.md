# Finish contract — the two runs and their cleanup

**This file is canonical for `/myflow-finish`'s two-run contract** — the preflight signals, both
runs' procedures, base-branch resolution and worktree cleanup.

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
for the host, as in run 2 below — that option belongs to the human doing this by hand, never to the
script — but signals 1 and 3 still run, and still run in this order.

### Run 1 — the branch is not merged

**Check for unfinished work first — before the landing question and before any git action.**
`check-unfinished-work.sh <worktree> <change-name>` prints one verdict line and exits 0
whenever it reached a verdict. It exits 2 with **no** verdict line when it cannot read the worktree.
Run it once per worktree in the set found by **Resolving a change's worktrees** below — never a raw
read of the state file's `worktrees` map, for the same reason the preflight verdict above does not
read it raw. A resolved set that comes back empty stops the run rather than passing as `CLEAR` from
every worktree.

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

Only then ask, **before any git action**, how the branch should land:

> **How should this branch land?**
> - **Open a pull request** *(default, recommended)*
> - **Merge and push**
> - **Handle it manually**

Then run to completion without asking again. The answer is never remembered between runs.

**Before any route commits, reshape the branch.** Run
`git -C <abs-worktree> reset --soft <recorded-merge-base>`, where `<recorded-merge-base>` is the
merge base recorded in the state file's `worktrees` map for this worktree — the same merge base
**Resolving a change's worktrees** and the finish-preflight verdict above both reference. This
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
the one irreversible step.

### Run 2 — the branch is merged

1. **Verify the merge.** Use a PR CLI when one is usable for the host; otherwise
   `git merge-base --is-ancestor`. That fallback must stay reachable on its own — it is the only
   merge evidence available on a non-GitHub forge. **Not merged → this is not run 2.**
2. **Position the main checkout on the archive branch, before anything else touches it.** Resolve
   `<base>` with `resolve-base-branch.sh` against the apply worktree, exactly as Run 1 does above,
   then invoke `prepare-archive-branch.sh <main-checkout> <base> chore/archive-<name>`. The exit
   contract: `0` positioned — the checkout is on `chore/archive-<name>`, cut from a fast-forwarded
   `<base>`; `1` a named refusal — a dirty working tree, **on `<base>` or off it**, a detached
   `HEAD`, an existing archive branch not descended from `origin/<base>`, or a `<base>` or
   `<archive-branch>` name that fails the guard's shape check; `2` an argument is missing,
   `<main-checkout>` is missing, unreadable, or not a git worktree, `HEAD`'s own ref cannot be read,
   or a checkout the guard performs fails; `3` `<base>` cannot be reconciled with `origin` — it
   has diverged, `origin/<base>` does not resolve, or there is no `origin` remote at all.
   Anything but exit `0` stops run 2 here, with nothing staged, committed, pushed or removed.

   **When the script is absent** — a harness whose repository does not carry it — perform the same
   positioning by hand, in this order, and say in the handoff that it was done manually. The guard is
   never skipped for want of the script. The steps are stated in full rather than cited, for the
   reason the resolver's own fallback above gives: the guard being absent takes its header with it.
   Run the wrapped, credential-free fetch first, so an
   unreachable remote refuses quickly rather than hanging. Then read `HEAD`: **refuse a detached
   `HEAD`.** Then read the working tree with `git -C <main-checkout> status --porcelain` — the same
   test the preflight's signal 3 uses — and **refuse a dirty tree wherever it is found, on `<base>` as
   well as off it**, naming both the branch found and `<base>`; uncommitted changes would otherwise
   ride onto the archive branch unremarked. On a clean tree, check out `<base>` if the checkout is not
   already on it, then fast-forward it to `origin/<base>`, **refusing a base that cannot fast-forward**
   rather than merging or resetting it. Finally create `chore/archive-<name>` from that base and check
   it out — or, when it already exists, **reuse it only if it is descended from `origin/<base>`** and
   refuse it otherwise. Apply each refusal in that order and never accept a guess in place of any of
   them.
3. **Archive the change** — `spectre archive <name>` moves it into
   `<project>/spectre/changes/archive/<name>/`. **The archived leaf carries no date prefix**, because
   `spectre archive` adds none: the date-ordered archive OpenSpec gave this pipeline is a real loss,
   accepted, and a prefix re-added here would describe a move the tool does not perform.
   **One call per change, parent and sub-change alike.** A `<name>-fix-N` sub-change is a flat
   sibling under `<project>/spectre/changes/`, never a directory inside its parent — `spectre new`
   refuses an id that is not a single flat directory name — so the parent's call cannot reach it and
   each sub-change is archived by its own call in this same step. Never left behind, never archived
   alone. **There is nothing to sync into
   `<project>/spectre/specs/` first, and no sync step is ever to be added back**: a change edits that
   tree directly on its own branch, so its spec edits reached the base branch with the merge step 1
   proved.
4. **Commit the archive on `chore/archive-<name>` — no push.** There is no merge to do: the change
   branch was already merged, which step 1 proved. Run 2 never merges anything into the base branch,
   and never commits the archive on the base branch itself. Every commit run 2 makes after step 2 —
   this one and the self-review report at step 9 alike — asserts `chore/archive-<name>` rather than
   assuming it: naming the directory with `git -C <main-checkout>` fixes the directory, not the
   branch. A finished change never leaves the archive move uncommitted in the working tree. The push
   happens at step 10, after self-review; step 11, which restores the checkout, closes the run.
5. **Clean up the worktrees, the local branch and the remote branch, then remove the workspace's
   database and bucket** — the worktree half being **Worktree cleanup**
   (`skills/flow-contracts/finish-contract.md`) below.

   **`BASE` is resolved per worktree, inside the cleanup loop below — never once for the whole
   change.** A multi-repo change has one `origin` and one default branch per repository, so a name
   resolved against one worktree can be the wrong ref, or no ref at all, against another
   repository's `origin`. Worktree cleanup's check 3 below is where this resolution actually runs:
   for each worktree in the resolved set it invokes `resolve-base-branch.sh` against that same
   worktree, immediately before that worktree's own removal — the same call and exit contract as
   **Resolve the base branch** under Run 1 above, run again here because this is a separate
   invocation and nothing carries `BASE` over from run 1's. Step 4's archive commit runs before this
   step, so every worktree in the set is still present when its own resolution runs — satisfying the
   same "before cleanup removes it" ordering this step has always required, just per worktree rather
   than once for the change. Anything but exit `0` for a given worktree — stop and ask, exactly as
   Run 1 does, and leave every worktree alone, per **Any failed check leaves every worktree alone**
   below.

   The removal runs the project's `remove` command, read from the command table
   **Project configuration** (`skills/flow-contracts/project-configuration.md`) is canonical for,
   with the workspace id substituted into its text by the mechanism that same file defines. **Run 2
   is not handed that id and does not need to be**: it is derived from the change name and from
   nothing else, deterministically and without ever being recorded, per
   **The workspace id** (`skills/flow-contracts/workspace-isolation.md`) — so run 2 re-derives it
   and arrives at the id `/myflow-do` used, in a session that shared nothing with it.

   **A project declaring no `## workspace isolation` section, or no `remove` command in it, has this
   half skipped rather than failed** — a step whose artifact is already absent is a success, which is
   the same re-entrancy rule every other removal in run 2 follows.

   **A failed removal does not stop run 2 here; it is reported, and step 7 decides the verdict** —
   from the project's survivor report and never from this command's exit code, per
   **Creation and cleanup** (`skills/flow-contracts/workspace-isolation.md`).
6. **Remove the proposal artifact source** from the state directory, on the condition its row in
   **Temporary artifacts registry** (`artifacts-registry.md`) gives. That section carries the condition and
   the reason for it; this step does not repeat either.
7. **Verify the cleanup.** Run `check-cleanup-complete.sh <repo> <name> <state-dir>` once
   per repository, **after** every removal above — it is there to judge what the run actually left
   behind, which is the one thing run 2 previously assumed.

   | Verdict | What run 2 does |
   |---------|-----------------|
   | `COMPLETE:` | report the cleanup as verified, **relay every clause the line carries after ` — ` word for word**, and go on to step 8 |
   | `LEFTOVER:` | name what remains, **do not write `FINISHED`**, and stop at `IN_PROGRESS` |

   **A `SKIPPED:` clause on a `COMPLETE:` line is relayed, never dropped, and the two rows are
   symmetric for that reason.** The guard appends its notes to the verdict after ` — `, and a
   `SKIPPED:` note there says a registry row was *not* verified — reached, for instance, as
   `COMPLETE: <repo> — … — SKIPPED: the workspace survivor verification — '<cmd>' exited 7, so the
   service could not be reached`. A run that reported only "cleanup verified" would have told the
   operator the opposite of what the guard said, while following this table to the letter. **A skip
   is never a pass**: `<agents repo>/scripts/check-cleanup-complete.sh`'s own header is canonical for why, and it
   is the reason the clause is quoted rather than summarised — the row it leaves unverified and the
   reason it could not be verified are both inside it. The relay does **not** block step 8; why an
   unreachable service must not strand an already-merged change is stated once under
   **Creation and cleanup** (`skills/flow-contracts/workspace-isolation.md`).

   A non-zero exit with **no verdict line** is the third outcome and not a verdict: report it, leave
   the affected `worktrees` entries in the state file, and treat it exactly as `LEFTOVER` — an
   unverified cleanup is not a verified one. The exit code is checked as well as the line, because a
   caller that greps for `COMPLETE` in empty output finds nothing.

   **None of this is a cue to bring the stack back up.** Once check 5 in **Worktree cleanup** below
   has stopped the project's declared stack, every later run-2 step that touches it — a
   reported-and-continued removal failure at step 5 above, or a `SKIPPED:` clause on a `COMPLETE:`
   line here — is that same stack's absence showing up again, correctly, one step later. Restarting
   it to make one of those steps succeed undoes what check 5 was for and answers a question this
   procedure never asked.

   **A leftover blocks the `FINISHED` write, and that is the whole point of having a verdict.**
   `FINISHED` is terminal: `/myflow-finish` stops at it and `/flow-status` does not list it, so a
   change written `FINISHED` over a known leftover has exactly one record of that leftover — the
   console line — which is the transcript-only record this pipeline refuses everywhere else. Left at
   `IN_PROGRESS` instead, the change stays listed, stays re-runnable, and the state file it already
   has is the durable record; no new field is invented to carry a fact the state itself carries.

   **Run 2 is re-entrant, which is what makes that safe.** Every step is remove-or-move *if present*
   and a step whose artifact is already gone is success, not an error — so a re-run after the
   operator clears the leftover repeats the verification and nothing else. An already-archived change
   directory means step 3 is already done: the archive move is skipped, not repeated, and the run
   continues to cleanup and verification.

   **When the script is absent** — a repository that does not carry it — check the same registry rows
   by hand, in the same order, and say in the handoff that the verification was done manually. The
   check is never skipped for want of the script, and "not verified" is never reported as verified.
8. **Write `FINISHED`**, clearing from `worktrees` **only the entries whose removal actually
   succeeded** — see **Worktree cleanup**
   (`skills/flow-contracts/finish-contract.md`) below — and carry every other field
   forward. This step is reached only on `COMPLETE:`.
9. **Run self-review** — after `FINISHED` is written; a skip, a failure, or a decline never moves
   the change off `FINISHED`. It is skippable per run, with running it the default. It gathers its
   input by invoking `gather-self-review-context.sh` rather than having the reasoning pass
   re-read files inline, and runs **one** combined reasoning pass covering all five angles below,
   together with the operator's 1-5 rating, never as five separate dispatches.

   | # | Angle | Label |
   |---|-------|-------|
   | 1 | Problems encountered, and what pipeline change would avoid them | `myflow-fix` |
   | 2 | Token/time cost, and what would reduce it without quality loss | `myflow-cost` |
   | 3 | What went well, and how to reproduce it | `myflow-improvement` |
   | 4 | What could be automated or moved to a script | `myflow-automation` |
   | 5 | What could move to the Go app or its persistent storage | `myflow-stats-app` |

   Angle 5's remit covers the records the pipeline writes to files today and the derivation work
   now done in Bash or by the agent — **not** what the SPA should display.

   **Every angle produces zero or more findings, and an angle that produces none says so
   explicitly** — present-but-empty, the way `## Decisions` and `## Open questions` already are,
   rather than omitted. A silent angle and a skipped angle are indistinguishable to a reader, which
   is how KAN-73's cost angle passed unnoticed while its section existed.

   **Every finding is explained in the message body before any prompt fires** — what was observed,
   what breaks, and what the fix would be. A prompt's option text cannot carry that explanation, so
   the prompt records the decision only: a filed issue is durable, and an explanation arriving
   afterward describes something the operator did not agree to.

   The filing ask is **one multi-select prompt per angle**, listing that angle's findings and
   defaulting to filing none of them, shape per **Operator prompts**
   (`skills/flow-contracts/operator-prompts.md`). A finding filed this way carries its angle's
   label on top of the set **Labels on issues the pipeline creates**
   (`skills/flow-contracts/jira-integration.md`) already defines.

   The report is committed onto `chore/archive-<name>` — asserting that branch rather than assuming
   it, and not pushed here; step 10 carries the push — to
   `<project>/docs/self-review/<name>-self-review.md`. It carries one section per angle,
   all five present; each finding is one line naming its angle's label, the finding, and its
   disposition — the issue key when filed, an explicit declined marker when not — and an angle with
   no findings carries an explicit none-marker instead of finding lines. **This procedure is
   canonical here.** Step 9 of `skills/flow/archive.md`'s own run 2 carries only what is
   specific to *executing* it: the script invocation and its arguments, the exact prompt wording,
   and the report-commit shell. It is not a second statement of this rule.

   **There is no requirements layer above this one; change this file.** The procedure was first
   written as **Requirement: Self-review runs only after FINISHED is written**
   (`<agents repo>/openspec/specs/myflow-self-review/spec.md`), alongside sibling requirements in that
   same file for context gathering, the combined pass, the per-angle filing ask, the rating and
   the report path. That capability was frozen with the rest of the `<agents repo>/openspec/` tree at the spectre
   cutover and not migrated, so it records where the rule came from and governs nothing: the
   statement above is both the requirement and the runtime source of the procedure.
10. **Push the archive branch and open its pull request.** Push `chore/archive-<name>`; open a pull
    request against `<base>` via a PR CLI when usable for the host, else print the forge's
    create-PR URL and ask whether it was opened — the same shape Run 1's pull-request route above
    uses. This push carries both the archive commit and the self-review report from step 9, so
    there is no window in which the archive pull request merges while the report is still
    unwritten. Run 2 never pushes to `<base>`.

    A failed push, or a failed pull-request creation, is reported with the command's own output. It
    never moves the change off `FINISHED` — the change is already terminal by step 8 — and the
    handoff names the unpushed branch `chore/archive-<name>` and prints the exact push and
    pull-request commands needed to land it by hand. The archive is never reported as landed until
    this step actually lands it.
11. **Restore the main checkout to `<base>`.** Successful or not, run 2 leaves the main checkout on
    `<base>`, never on the archive branch.

**The Jira `Done` transition fires before step 9, not after it.** Per **Jira integration**
(`skills/flow-contracts/jira-integration.md`)'s own timing — the issue moves to `Done` after the
archive move and the state write — that transition has already happened by the time step 9 begins,
so self-review has nothing to delay: there is no Jira write left in run 2 for it to sit in front of.

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

### Worktree cleanup

Which worktrees those are, and the `git worktree list --porcelain` scan that finds them when
the state file's map is absent or empty, are **Resolving a change's worktrees** above.

For each worktree, run **every** check below before removing anything:

```bash
# 1. no uncommitted tracked changes — must be empty
git -C "$WT" status --porcelain --untracked-files=no

# 2. no untracked files that git does not already ignore — must be empty.
#    `--others --exclude-standard` lists exactly the files `--force` would destroy and
#    `.gitignore` does NOT cover. This is the check that makes `--force` safe.
git -C "$WT" ls-files --others --exclude-standard

# 3. no commits that exist only here. Resolve `BASE` fresh for THIS worktree — never reused from
#    another worktree in the set. A multi-repo change has one `origin` and one default branch per
#    repository, so a name resolved against one worktree's `origin` can be the wrong ref, or absent
#    entirely, in another repository. Resolving it here, before this worktree's own removal below,
#    is what "before cleanup removes it" (run 2 step 5, above) means per worktree.
BASE="$(resolve-base-branch.sh "$WT")" || { echo "cannot resolve the base branch for $WT — stop and ask"; false; }
#    `@{upstream}` ERRORS when no upstream is configured, and an empty capture would read as
#    "nothing unpushed" — so resolve it explicitly and never let a failed lookup pass as success.
#
#    Step 1 already proved the branch is an ancestor of the base branch, which is STRICTLY
#    STRONGER evidence than "pushed to its own upstream": the commits are in the base branch.
#    Accept that first. Requiring the upstream regardless would lock out the ordinary
#    squash-merge workflow — GitHub's "delete head branch on merge" plus `fetch.prune=true`
#    removes the tracking ref, after which no upstream can ever resolve and the branch cannot be
#    re-pushed because it no longer exists on the remote.
if git -C "$WT" merge-base --is-ancestor HEAD "origin/$BASE" 2>/dev/null; then
  :                                            # already merged into base — nothing can be lost
elif UP="$(git -C "$WT" rev-parse --abbrev-ref --symbolic-full-name '@{upstream}' 2>/dev/null)"; then
  git -C "$WT" log --oneline "$UP..HEAD"       # must be empty
else
  echo "not merged, and no upstream — cannot prove these commits exist anywhere else"; false
fi

# 4. what `--force` WILL destroy: ignored files. This does not gate removal — it is shown to
#    the operator, who decides. `--exclude-standard` in check 2 hides everything matched by
#    .gitignore, <project>/.git/info/exclude or the global excludes file, and "ignored" is NOT "disposable":
#    a deliberately-ignored .env, a local override config, or this pipeline's own
#    <abs-worktree>/.superpowers/sdd/ records are all ignored and all irreplaceable.
git -C "$WT" ls-files --others --ignored --exclude-standard

# 5. the project's local stack is stopped — run its `## stop` command if declared. Give it a
#    bounded wait — 60 seconds — and treat a timeout as a FAILED check.

# 6. no process is still running FROM this worktree. Runs whether or not `## stop` is declared,
#    and whatever that command exited: a project that declares no stop command can still have a
#    stack running from its worktree, and a stop command that exits 0 is not evidence that it
#    worked. Verifying that is the whole of what this check adds.
check-worktree-processes.sh "$WT"
```

**Check 6 requires the orchestrating shell's own cwd to be outside every worktree in the resolved
set before it runs.** `check-worktree-processes.sh`'s own header treats a process whose working
directory is at or under the worktree as held, and a shell that is itself `cd`'d into `$WT` (or into
another worktree in the same set) is exactly such a process — it would report `HELD:` against
itself, not against the stack this check exists to catch. `cd` out first, for every worktree, before
this check runs for any of them.

**Check 6 is a gate, and both of its bad outcomes are failures.** `HELD:` names each process
holding the worktree, and exit 2 says the guard could not answer at all — the scanning tool is
absent, or the worktree path is not readable. Neither is a pass: an inability that proceeded to
removal would be the failure this check exists to prevent, arrived at more quietly. On either, stop
at `IN_PROGRESS`, leave every worktree alone, and report each pid, its working directory, and the
command that reaches it:

```bash
ps -o pid,command -p <pid>
```

**There is no confirmation to proceed past check 6**, unlike check 4's disclosure. That disclosure
is safe to confirm because the operator can see what is at stake and decide; a live process is
different in kind. Confirming it destroys the only records that can reach the process afterwards,
and the resulting orphan holds ports shared across every workspace — which is how this incident
reached the operator both times it happened. The remedy is to clear the process and re-run, while
the worktree still exists and the project's own stop command can still read what it started.

**Check 4 is a disclosure, not a gate.** When it lists anything, **stop and show the list**, and
ask for explicit confirmation before removing that worktree. Do not try to classify the entries as
build output: no allowlist of names can be trusted, because the operator decides what they ignore.
Empty list → proceed without asking.

**`/myflow-fast` overrides the ask, and only the ask.** Its own **Guardrails**
(`skills/myflow-fast/SKILL.md`) state that override and why it is safe there: that command reports
what `--force` will destroy and proceeds, having already preserved and committed the records worth
keeping before it reaches cleanup. The override is named here too, so two files cannot silently
disagree about a step that destroys files. It reaches nothing else — checks 1, 2, 3, 5 and 6 stay
gates under every command, and an irreplaceable **unpreserved** entry still stops the run and asks,
under `/myflow-fast` as much as here.

Then, and only then:

```bash
git -C "$REPO" worktree remove --force "$WT"
git -C "$REPO" branch -d "spectre/<name>"
git -C "$REPO" worktree prune
```

- **`--force` destroys every ignored file in the worktree, and no check prevents that.** Checks 1
  and 2 establish only that nothing *tracked-and-modified* and nothing *untracked-and-unignored*
  is at risk. They say nothing about ignored files, because `--exclude-standard` is what hides
  them — and "ignored" is not "disposable". Check 4 exists to make that visible rather than to
  prevent it: it lists exactly what will die, and the operator confirms. Claiming the checks make
  `--force` safe would be false, and was: a gitignored `.env` passes checks 1 and 2 and is
  destroyed silently.
- **Neither check sees a file whose `assume-unchanged` bit is set.** `git status` is blind to it
  by design. Rare, operator-inflicted, and named here so it is a known limit rather than a
  surprise.
- **`git branch -d`, never `-D`.** It must be free to refuse an unmerged branch.
- **An already-removed worktree is success**, not an error.
- **Any failed check leaves every worktree alone** and reports why. There is no partial cleanup.
  This includes check 5: a `## stop` command that **exits non-zero, is not found, or has to be
  interrupted** is a *failed* check, not an absent one — only an undeclared key is skipped. Give it
  a bounded wait rather than letting it hang the run.
- **Verify each removal actually succeeded** before writing state. If any `git worktree remove`
  fails for a reason the checks did not predict — a file lock, a permission error — report it and
  leave that worktree's entry in `worktrees`. Writing `worktrees: {}` regardless would drop it from
  the only authoritative list, and nothing would ever find it again.

Then the change's **remote** branch:

```bash
# `push --delete` exits non-zero BOTH when the branch was already gone and when the push was
# refused, so the two are told apart by git's message and never by the exit code alone. Measured
# against a scratch remote on this machine (git 2.50.1): an already-absent branch prints
# `error: unable to delete 'spectre/<name>': remote ref does not exist` and exits 1, and the
# stale remote-tracking ref SURVIVES that failure.
OUT="$(git -C "$REPO" push origin --delete "spectre/<name>" 2>&1)"; RC=$?
if [ "$RC" -eq 0 ]; then
  echo "remote branch deleted: origin/spectre/<name>"
elif printf '%s' "$OUT" | grep -q 'remote ref does not exist'; then
  # The forge deleted it on merge. Prune the ref it left behind: check-cleanup-complete.sh reads a
  # surviving refs/remotes/origin/spectre/<name> as a leftover, and it would be a real one.
  git -C "$REPO" fetch --prune --quiet origin
  echo "remote branch already gone — the forge deleted it on merge"
else
  echo "remote branch NOT deleted: $OUT"
fi
```

- **The remote branch is deleted without a further prompt.** Run 2 is reached only by proving the
  branch is an ancestor of the base branch, so its commits are in the base branch and nothing can be
  lost — which is why this is not gated the way check 4's disclosure is.
- **An already-absent remote branch is success**, not an error, and the outcome is reported either
  way: deleted, already gone, or refused.
- **A refused push is reported, never swallowed.** A bare `|| true` would make an expired
  credential indistinguishable from a branch the forge already removed, and leave the remote branch
  standing with nothing said about it.
- **The remote delete is not gated on the local one succeeding.** Gating it would leave the remote
  branch behind whenever anything unrelated failed, which is the state this step exists to end.

The stack-stopped check reads the optional `## stop` key from `<project>/.flow/project.md` —
see **Project configuration** in `skills/flow-contracts/project-configuration.md`. When the key
or the file is absent the check is **skipped, not failed**, and cleanup proceeds on the strength of
the other checks — check 6 among them, which is what makes an undeclared `## stop` key survivable:
a project that declares nothing to stop is still checked for a process running from its worktree.

