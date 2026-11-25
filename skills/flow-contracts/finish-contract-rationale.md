# Finish contract — rationale

This file is the reasoning behind `skills/flow-contracts/finish-contract-run1.md` and
`skills/flow-contracts/finish-contract-run2.md`.
**A `/flow*` run never loads it — appendices are for whoever edits a contract.**

## finish-contract-run1.md — Surface foreign staged work before the preflight

## finish-contract-run1.md — Run 1 — the branch is not merged

Reshape, moved verbatim from the contract, where it followed "is the chain
**Git boundaries** (`git-boundaries.md`) gives": , and is not written out a second time here.

Resolve the base branch, moved verbatim from the end of the hand-fallback paragraph (now in `finish-hand-fallbacks.md`): That header is
the authority for the exact commands and their order; this paragraph does not repeat what it already
states. The character rule itself is stated here in full, not cited, because this paragraph's own
precondition is the script's absence — a fallback the operator applies by hand when the script is
absent cannot defer them to a file that, by the same precondition, is not there to read.

## finish-contract-run1.md — Resolving a change's worktrees

Moved verbatim from the contract, where it opened the sentence whose call-site list stays there:
"That rule and the commands it binds are not restated here; what follows is specific to bare
`/flow`:".

## finish-contract-run1.md — one task, one worktree

The operator's rule: one task gets one worktree, and its self-review fixes land on the same pull
request as the code. So run 1 runs its self-review pass on `spectre/<name>`, in the apply
worktree, before any route pushes, and run 2 only cleans up. No planning, archive or self-review artifact is committed at all.

Reversed for `design.md` alone: run 1 writes it out as
`<project>/docs/designs/<name>.md` inside the implementation commit, because the chosen option, the
rejected ones and why were surviving only in commit bodies and Jira. Rejected: a plan task that
copies the file — stale the moment a fix round supersedes a decision; restoring the archive
commit — two landings' worth of mechanics for one file; an opt-in `## design records` key —
configurability one project asked for; a mutable status header — no guard would keep it true.

Rejected, and removed: a second, throwaway landing worktree `<project>-worktrees/_landing-<name>`
and a second branch `chore/archive-<name>`. Run 1's merge-and-push route positioned that worktree
on `<base>` itself and merged there, which moved the local `<base>` ref under the main checkout: the main checkout then showed the change reverse-staged until run 2 reset its index. Run 2 then archived on `chore/archive-<name>` and landed it through a second
push or pull request — two landings per change, and a window in which the code had landed with its
change still open.

The merge-and-push route pushes `spectre/<name>` before `<base>`, so `origin/spectre/<name>` never
stays at its pre-reshape tip. Run 2's `git branch -d` does not depend on it: `flow-guard remove-change-worktrees`
points the branch's upstream at `origin/<base>` first, because `-d` refuses a branch not merged into
its upstream — else into the main checkout's `HEAD`, which run 2 refreshes only after cleanup — and a
forge merge deletes `origin/spectre/<name>`; judged against either, a merged branch would fail again on every re-run. A rejected push of `<base>` re-resolves the base before re-syncing because a rejected push
fetches nothing: `origin/<base>` would still name the tip the push was refused against, and the
re-sync would rebase onto it and be refused again.

## finish-contract-run1.md — a clean drift branch is switched back

`flow-guard integrate-change prepare` switches a clean drift checkout back to its default branch.

Only the clean shape switches — `STAGED-CLEAN` and a lone `DRIFT-BRANCH` — because then nothing
tracked can be lost: no staged, modified or unmerged entry rides along or blocks the switch. A
detached HEAD is excluded because its commits sit on no branch; switching away would leave them
reachable only through the reflog. A checkout on a base the change records is excluded because run 2 refreshes a main checkout sitting on the change's recorded base onto that base, so
moving it to the default branch would undo what run 2 relies on. The switch is a plain
`git switch`, never `checkout -B` or a reset, so the branch left behind keeps every commit it
carried.

The switch lives in `flow-guard integrate-change`, not in the drift guard nor a new Go subcommand, for the
reasons decision `prepare-switch-clean-drift`'s Considered line records.

## finish-contract-run1.md — Run the self-review pass

`skills/flow/integrate.md`'s step 4 carries only what is specific to *executing* it: the
script invocation and its arguments and the commit shell. It is not a second statement of this
rule.

## Moved — the finish session trims

### finish-contract-run1.md — Finish contract (the base ref)

Moved verbatim, where it followed "never the bare local name on its own.":

A bare local branch is
wrong here because the ancestor test resolves whatever ref it is handed, and `flow-guard resolve-base-branch`'s
fetch refreshes only remote-tracking refs, never local branches; a stale local branch of the same
name would then silently feed the RUN1/RUN2/REFUSE decision, the gate in front of this command's most
destructive step.

### finish-contract-run1.md — Surface foreign staged work before the preflight

Moved verbatim, where it closed "the foreign staged work the affected main checkouts carry.":

It is pre-run on purpose: which run this is is not yet
known, and a main checkout's staged residue is exactly what a resumed run is tempted to clear by
hand once a later REFUSE arrives. That judgment belongs to the operator; this surface is what hands
it to them before the run reaches the refusal it would otherwise improvise around.

### finish-contract-run1.md — the main checkouts' drift

Moved verbatim; the run-loaded copy keeps the sentence up to "counts tracked entries":

`flow-guard check-main-checkout-drift` runs once per
distinct main checkout, and its header is canonical for the verdict grammar it prints: a
`DRIFT-BRANCH` line names a checkout on any other branch than the one `refs/remotes/origin/HEAD`
points at, a `DRIFT-DIRTY` line counts tracked entries — shapes the staged-only listing above cannot see.

### finish-contract-run1.md — Run 1, `[canonical-worktree]`

Moved verbatim, where it followed "whose own `<project>/<spec-root>/changes/<change-name>/tasks.md` exists."; the empty-set sentence after it was cut as a duplicate of **Resolving a change's worktrees** (`skills/flow-contracts/worktree-resolution.md`):

Without it, a
satellite worktree's call falls back to resolving the link's peer name through `peers`, which cannot
resolve from inside a worktree.

### unfinished-work-gate.md — Stop is the recommendation

Moved verbatim from `finish-contract-run1.md`, where it followed "**Stop is marked as the recommendation, and the reason is stated rather than left to be inferred.**":

The gate only fires because something really is unfinished, and finishing it is the cheapest of the
three to recover from — Continue is the only course that reaches an irreversible step, and it exists
for work the operator deliberately deferred, which is a judgment only they hold. Marking a
recommendation is not a courtesy here: the planning-gate capability requires every choice a
`/flow*` command offers to name its recommended option, and this prompt is one of them.

### finish-contract-run1.md — Never fall back to `HEAD@{upstream}`

Moved verbatim, where it followed "**Never fall back to `HEAD@{upstream}`.**":

bare `/flow` runs inside the apply worktree, where
`HEAD` *is* `spectre/<name>` — so that fallback resolves to the change's **own** upstream, making
the merge check `spectre/<name>` vs `origin/spectre/<name>`, which is true the moment the branch
is pushed. That silently reports an unmerged change as merged, and run 2 then deletes its
worktree. `flow-guard resolve-base-branch` is where this rule is enforced: it never
consults `HEAD@{upstream}`, and its assertion that `BASE` differs from the current branch is
unconditional, which is what makes that class of misresolution impossible rather than merely
unlikely.

### finish-contract-run1.md — an unresolvable base

Moved verbatim, where it followed "If no base branch resolves, **stop and ask**.":

An unresolvable base is an honest unknown; a guessed
one is a wrong answer at the only irreversible step.

### finish-contract-run1.md — No verification gate

Moved verbatim, where it followed "No tests, no linters, no spec-coverage check.":

Correctness was established during `/flow`'s implement phase — TDD per task, the final review
panel — and by the human gate. Re-running it here would repeat finished work immediately before
the one irreversible step.

### finish-contract-run1.md — Resolving a change's worktrees, `substr`

Moved verbatim, where it followed "**The path is taken with `substr`, never `$2`.**":

`worktree list --porcelain` emits it raw, so a
field reference truncates any path containing a space at the first one: fed
`worktree /tmp/my worktree` it yields `/tmp/my`, and the run then `--force`-removes a path that is
not the worktree, or fails having named the wrong one. `10` is one past the length of the literal
`worktree ` prefix. The branch on the next line is a ref name and cannot contain a space, so `$2` is
right for it.

### finish-contract-run2.md — step 1

Moved verbatim, where it followed "otherwise `git merge-base --is-ancestor`.":

That fallback must stay reachable on its own — it is the only
   merge evidence available on a non-GitHub forge.

### finish-contract-run2.md — step 5, why the refresh

Nothing in the pipeline moves the local `<base>` ref: every change lands from its own apply
worktree, as a push of its branch or a forge merge, so the operator's main checkout is simply
behind `origin/<base>` afterwards. Step 5 fast-forwards it so it shows the landed change, and only
when nothing in it could be lost; anything else is the operator's to settle.

### finish-contract-run2.md — Worktree cleanup, check 3's comment

Moved verbatim out of check 3's shell block, where it followed the `@{upstream}` comment:

```bash
#
#    Step 1 already proved the branch is an ancestor of the base branch, which is STRICTLY
#    STRONGER evidence than "pushed to its own upstream": the commits are in the base branch.
#    Accept that first. Requiring the upstream regardless would lock out the ordinary
#    squash-merge workflow — GitHub's "delete head branch on merge" plus `fetch.prune=true`
#    removes the tracking ref, after which no upstream can ever resolve and the branch cannot be
#    re-pushed because it no longer exists on the remote.
```

### finish-contract-run2.md — Worktree cleanup, a known limit

Moved verbatim from the bullets after the removal commands:

- **Neither check sees a file whose `assume-unchanged` bit is set.** `git status` is blind to it
  by design. Rare, operator-inflicted, and named here so it is a known limit rather than a
  surprise.

### finish-contract-run2.md — Worktree cleanup, verify each removal

Moved verbatim, where it followed "leave that worktree's entry in `worktrees`.":

Writing `worktrees: {}` regardless would drop it from
  the only authoritative list, and nothing would ever find it again.

### finish-contract-run2.md — the remote branch, the shell's comment

Moved verbatim from the top of the remote-delete shell block:

```bash
# `push --delete` exits non-zero BOTH when the branch was already gone and when the push was
# refused, so the two are told apart by git's message and never by the exit code alone. Measured
# against a scratch remote (git 2.50.1): an already-absent branch prints
# `error: unable to delete 'spectre/<name>': remote ref does not exist` and exits 1, and the
# stale remote-tracking ref SURVIVES that failure.
```

### finish-contract-run2.md — the remote branch, no prompt

Moved verbatim, where it followed "**The remote branch is deleted without a further prompt.**":

- **The remote branch is deleted without a further prompt.** Run 2 is reached only by proving the
  branch is an ancestor of the base branch, so its commits are in the base branch and nothing can be
  lost.

### finish-contract-run2.md — the remote branch, a refused push

Moved verbatim, where it followed "**A refused push is reported, never swallowed.**":

- **A refused push is reported, never swallowed.** A bare `|| true` would make an expired
  credential indistinguishable from a branch the forge already removed, and leave the remote branch
  standing with nothing said about it.

### finish-contract-run2.md — the remote branch, not gated

Moved verbatim, where it followed "**The remote delete is not gated on the local one succeeding.**":

- **The remote delete is not gated on the local one succeeding.** Gating it would leave the remote
  branch behind whenever anything unrelated failed, which is the state this step exists to end.

### finish-contract-run2.md — the stack-stopped check

Moved verbatim; the run-loaded copy keeps the sentence up to "the other checks":

When the key
or the file is absent the check is **skipped, not failed**, and cleanup proceeds on the strength of
the other checks.
