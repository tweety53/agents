# Finish contract — rationale

This file is the reasoning behind `skills/flow-contracts/finish-contract-run1.md` and
`skills/flow-contracts/finish-contract-run2.md`.
**A `/flow*` run never loads it — appendices are for whoever edits a contract.**

## finish-contract-run1.md — Surface foreign staged work before the preflight

Incident: KAN-546.

Moved verbatim from the contract, where it closed "a main checkout's staged residue is exactly what
a resumed run is tempted to clear by hand once a later REFUSE arrives": — the high-judgment surgery
gymie kan-437's run 2 performed inline across three repos, hard reset where it judged a clean
revert and stash where the work was distinct WIP.

## finish-contract-run1.md — Run 1 — the branch is not merged

Sync the branch onto the base (now `skills/flow/sync-onto-base.md`, KAN-859) — incident behind `aside-planning-artifacts.sh`: KAN-628.

Reshape, moved verbatim from the contract, where it followed "is the chain
**Git boundaries** (`git-boundaries.md`) gives": , and is not written out a second time here.

Resolve the base branch, moved verbatim from the end of the hand-fallback paragraph (now in `finish-hand-fallbacks.md`, KAN-859): That header is
the authority for the exact commands and their order; this paragraph does not repeat what it already
states. The character rule itself is stated here in full, not cited, because this paragraph's own
precondition is the script's absence — a fallback the operator applies by hand when the script is
absent cannot defer them to a file that, by the same precondition, is not there to read.

## finish-contract-run1.md — Resolving a change's worktrees

Moved verbatim from the contract, where it opened the sentence whose call-site list stays there:
"That rule and the commands it binds are not restated here; what follows is specific to bare
`/flow`:".

## finish-contract-run1.md — one task, one worktree (2026-10-04)

The operator's rule: one task gets one worktree, and its planning, archive and self-review
artifacts land on the same pull request as the code, as separate commits. So run 1 archives the
change and saves its self-review context bundle on `spectre/<name>`, in the apply worktree, before
any route pushes, and run 2 only cleans up.

Rejected, and removed: a second, throwaway landing worktree `<project>/.worktrees/_landing-<name>`
and a second branch `chore/archive-<name>`. Run 1's merge-and-push route positioned that worktree
on `<base>` itself and merged there, which moved the local `<base>` ref under the main checkout:
the main checkout then showed the change reverse-staged until run 2 reset its index (observed in
KAN-875's landing). Run 2 then archived on `chore/archive-<name>` and landed it through a second
push or pull request — two landings per change, and a window in which the code had landed with its
change still open.

A re-run of run 1 on an already-archived change skips the reshape because `reshape-branch.sh`
keeps only commits touching `<project>/spectre/changes/`: the bundle commit touches
`<project>/docs/self-review/` alone, so a reshape would fold it into the implementation commit. It
still runs `commit-archive.sh` and the bundle step, each a no-op once its commit exists, because the
earlier run may have stopped between `spectre archive` and either commit — a refused archive commit,
or a bundle commit refused after the archive commit landed — and a skip keyed on the working tree's
layout alone would leave that commit missing on every later re-run. Archived-ness is read from the
canonical repository alone because only it holds the change directory; a satellite worktree has
nothing to judge it by.

The merge-and-push route pushes `spectre/<name>` before `<base>`: the branch's upstream is
`origin/spectre/<name>`, last pushed at its pre-reshape tip, and `git branch -d` at run 2 refuses
a branch not merged into its upstream — a cleanup that would end `LEFTOVER` and fail again on every
re-run. A rejected push of `<base>` re-resolves the base before re-syncing because a rejected push
fetches nothing: `origin/<base>` would still name the tip the push was refused against, and the
re-sync would rebase onto it and be refused again.

## finish-contract-run1.md — Archive on the change branch

Step 1, moved verbatim from the contract, where it closed "**There is nothing to sync into
`<project>/spectre/specs/` first**": , and no sync step is ever to be added back.

Step 2, moved verbatim from the contract, where it followed "rather than let a stray path land on
`spectre/<name>` unremarked.": This has happened: an archive commit made this way reverted
skill-file content another change had shipped minutes earlier.

Step 2, the archive-scope guard's cannot-answer exit — incident: KAN-601.

## finish-contract-run1.md — Save the self-review context bundle

`skills/flow/integrate.md`'s step 4 carries only what is specific to *executing* it: the
script invocation and its arguments and the commit shell. It is not a second statement of this
rule.

## Moved by KAN-859 — the finish session trims

### finish-contract-run1.md — Finish contract (the base ref)

Moved verbatim (KAN-859), where it followed "never the bare local name on its own.":

A bare local branch is
wrong here because the ancestor test resolves whatever ref it is handed, and `resolve-base-branch.sh`'s
fetch refreshes only remote-tracking refs, never local branches; a stale local branch of the same
name would then silently feed the RUN1/RUN2/REFUSE decision, the gate in front of this command's most
destructive step.

### finish-contract-run1.md — Surface foreign staged work before the preflight (KAN-859)

Moved verbatim (KAN-859), where it closed "the foreign staged work the affected main checkouts carry.":

It is pre-run on purpose: which run this is is not yet
known, and a main checkout's staged residue is exactly what a resumed run is tempted to clear by
hand once a later REFUSE arrives. That judgment belongs to the operator; this surface is what hands
it to them before the run reaches the refusal it would otherwise improvise around.

### finish-contract-run1.md — the main checkouts' drift (KAN-859)

Moved verbatim (KAN-859); the run-loaded copy keeps the sentence up to "counts tracked entries", without the KAN-647 provenance clause:

`check-main-checkout-drift.sh` runs once per
distinct main checkout, and its header is canonical for the verdict grammar it prints: a
`DRIFT-BRANCH` line names a checkout on any other branch than the one `refs/remotes/origin/HEAD`
points at, a `DRIFT-DIRTY` line counts tracked entries — the shapes KAN-647's reverse-image
incidents took, which the staged-only listing above cannot see.

### finish-contract-run1.md — Run 1, `[canonical-worktree]` (KAN-859)

Moved verbatim (KAN-859), where it followed "whose own `<project>/<spec-root>/changes/<change-name>/tasks.md` exists."; the empty-set sentence after it was cut as a duplicate of **Resolving a change's worktrees** (`skills/flow-contracts/worktree-resolution.md`):

Without it, a
satellite worktree's call falls back to resolving the link's peer name through `peers`, which cannot
resolve from inside a worktree.

### finish-contract-run1.md — Run 1, `check-visual-verify-dispatched.sh` (KAN-859)

Moved verbatim (KAN-859), where it followed "not a second prompt.":

This guard exists because a stage
mark is not evidence a stage ran — a run can mark `flow.visual-verify` begun and completed around no
dispatch at all; see that guard's own header for the account.

### unfinished-work-gate.md — Stop is the recommendation (KAN-859)

Moved verbatim (KAN-859) from `finish-contract-run1.md`, where it followed "**Stop is marked as the recommendation, and the reason is stated rather than left to be inferred.**":

The gate only fires because something really is unfinished, and finishing it is the cheapest of the
three to recover from — Continue is the only course that reaches an irreversible step, and it exists
for work the operator deliberately deferred, which is a judgment only they hold. Marking a
recommendation is not a courtesy here: the planning-gate capability requires every choice a
`/flow*` command offers to name its recommended option, and this prompt is one of them.

### finish-contract-run1.md — Never fall back to `HEAD@{upstream}` (KAN-859)

Moved verbatim (KAN-859), where it followed "**Never fall back to `HEAD@{upstream}`.**":

bare `/flow` runs inside the apply worktree, where
`HEAD` *is* `spectre/<name>` — so that fallback resolves to the change's **own** upstream, making
the merge check `spectre/<name>` vs `origin/spectre/<name>`, which is true the moment the branch
is pushed. That silently reports an unmerged change as merged, and run 2 then archives it and
deletes its worktree. `resolve-base-branch.sh` is where this rule is enforced: it never
consults `HEAD@{upstream}`, and its assertion that `BASE` differs from the current branch is
unconditional, which is what makes that class of misresolution impossible rather than merely
unlikely.

### finish-contract-run1.md — an unresolvable base (KAN-859)

Moved verbatim (KAN-859), where it followed "If no base branch resolves, **stop and ask**.":

An unresolvable base is an honest unknown; a guessed
one is a wrong answer at the only irreversible step.

### finish-contract-run1.md — No verification gate (KAN-859)

Moved verbatim (KAN-859), where it followed "No tests, no linters, no spec-coverage check.":

Correctness was established during `/flow`'s implement phase — TDD per task, the final review
panel — and by the human gate. Re-running it here would repeat finished work immediately before
the one irreversible step.

### finish-contract-run1.md — Resolving a change's worktrees, `substr` (KAN-859)

Moved verbatim (KAN-859), where it followed "**The path is taken with `substr`, never `$2`.**":

`worktree list --porcelain` emits it raw, so a
field reference truncates any path containing a space at the first one: fed
`worktree /tmp/my worktree` it yields `/tmp/my`, and the run then `--force`-removes a path that is
not the worktree, or fails having named the wrong one. `10` is one past the length of the literal
`worktree ` prefix. The branch on the next line is a ref name and cannot contain a space, so `$2` is
right for it. `<agents repo>/scripts/check-cleanup-complete.sh` parses the same stream the same way
(`<agents repo>/stats/internal/guard/cleanupcomplete.go`, row one: the rest of the `worktree ` line, the
branch's first field) — the guard
and the snippet it verifies must not disagree, or the wrong one gets copied next.

### finish-contract-run2.md — step 1 (KAN-859)

Moved verbatim (KAN-859), where it followed "otherwise `git merge-base --is-ancestor`.":

That fallback must stay reachable on its own — it is the only
   merge evidence available on a non-GitHub forge.

### finish-contract-run1.md — Archive on the change branch, the archived leaf (KAN-859)

Moved verbatim (KAN-859); the run-loaded copy keeps the bold claim alone:

**The archived leaf carries no date prefix**, because
   `spectre archive` adds none: a prefix re-added here would describe a move the tool does not
   perform.

### finish-contract-run1.md — Archive on the change branch, no spec sync (KAN-859)

Moved verbatim (KAN-859); the run-loaded copy keeps the bold claim alone:

**There is nothing to sync into
   `<project>/spectre/specs/` first**: a change edits that
   tree directly on its own branch, so its spec edits reach the base branch with the same merge.

### finish-contract-run1.md — Archive on the change branch, why the scope check (KAN-859)

Moved (KAN-859), where it followed "run between the add and the commit.":

`add -A` stages the whole worktree, not only the archive move, so it stages and commits whatever
   else the tree carried right alongside it.

### finish-contract-run1.md — Archive on the change branch, why the renders are preserved (KAN-859)

Moved (KAN-859), where it followed "under the scope the check above verifies.":

The store's rows are the terminal record, but rows that never
   reached it leave the worktree renders the only copies, and run 2's cleanup destroys those with the
   worktree; the committed copies are what the bundle serves when the store renders report
   skipped.

### finish-contract-run2.md — step 4, the `SKIPPED:` relay (KAN-859)

Moved verbatim (KAN-859), where it followed the `COMPLETE: <repo> — … — SKIPPED:` example:

A run that reported only "cleanup verified" would have told the
   operator the opposite of what the guard said, while following this table to the letter.

### finish-contract-run2.md — step 4, not a cue to restart the stack (KAN-859)

Moved verbatim (KAN-859), where it followed "**None of this is a cue to bring the stack back up.**":

Once check 5 in **Worktree cleanup** below
   has stopped the project's declared stack, every later run-2 step that touches it — a
   reported-and-continued removal failure at step 2 above, or a `SKIPPED:` clause on a `COMPLETE:`
   line here — is that same stack's absence showing up again, correctly, one step later. Restarting
   it to make one of those steps succeed undoes what check 5 was for and answers a question this
   procedure never asked.

### finish-contract-run2.md — step 4, a leftover blocks `FINISHED` (KAN-859)

Moved verbatim (KAN-859), where it followed "**A leftover blocks the `FINISHED` write, and that is the whole point of having a verdict.**":

`FINISHED` is terminal: bare `/flow` stops at it and `/flow-status` does not list it, so a
   change written `FINISHED` over a known leftover has exactly one record of that leftover — the
   console line — which is the transcript-only record this pipeline refuses everywhere else. Left at
   `IN_PROGRESS` instead, the change stays listed, stays re-runnable, and the state file it already
   has is the durable record; no new field is invented to carry a fact the state itself carries.

### finish-contract-run2.md — step 6, why the refresh

Nothing in the pipeline moves the local `<base>` ref: every change lands from its own apply
worktree, as a push of its branch or a forge merge, so the operator's main checkout is simply
behind `origin/<base>` afterwards. Step 6 fast-forwards it so it shows the landed change, and only
when nothing in it could be lost; anything else is the operator's to settle.

### finish-contract-run2.md — Worktree cleanup, check 3's comment (KAN-859)

Moved verbatim (KAN-859) out of check 3's shell block, where it followed the `@{upstream}` comment:

```bash
#
#    Step 1 already proved the branch is an ancestor of the base branch, which is STRICTLY
#    STRONGER evidence than "pushed to its own upstream": the commits are in the base branch.
#    Accept that first. Requiring the upstream regardless would lock out the ordinary
#    squash-merge workflow — GitHub's "delete head branch on merge" plus `fetch.prune=true`
#    removes the tracking ref, after which no upstream can ever resolve and the branch cannot be
#    re-pushed because it no longer exists on the remote.
```

### finish-contract-run2.md — Worktree cleanup, check 6's cwd (KAN-859)

Moved verbatim (KAN-859), where it followed "before it runs.":

`check-worktree-processes.sh`'s own header treats a process whose working
directory is at or under the worktree as held, and a shell that is itself `cd`'d into `$WT` (or into
another worktree in the same set) is exactly such a process — it would report `HELD:` against
itself, not against the stack this check exists to catch.

### finish-contract-run2.md — Worktree cleanup, check 6 is a gate (KAN-859)

Moved verbatim (KAN-859), where it followed "or the worktree path is not readable.":

Neither is a pass: an inability that proceeded to
removal would be the failure this check exists to prevent, arrived at more quietly.

### finish-contract-run2.md — Worktree cleanup, no confirmation past check 6 (KAN-859)

Moved verbatim (KAN-859), where it followed "unlike check 4's one ask.":

That ask is safe
because the operator sees exactly the irreplaceable entry at stake and decides; a live process is
different in kind. Confirming it destroys the only records that can reach the process afterwards,
and the resulting orphan holds ports shared across every workspace.

### finish-contract-run2.md — Worktree cleanup, check 4's empty bucket (KAN-859)

Moved verbatim (KAN-859); the run-loaded copy keeps the sentence up to "proceed without asking":

An empty unclassified bucket → **show the regeneratable bucket's count and proceed without
asking** — every entry in it is reproduced identically by the next run of whatever wrote it, so
confirming its loss adds nothing the operator can act on, and asking every single time a routine
archive leaves nothing but build output behind is a prompt with no real decision behind it.

### finish-contract-run2.md — Worktree cleanup, a known limit (KAN-859)

Moved verbatim (KAN-859) from the bullets after the removal commands:

- **Neither check sees a file whose `assume-unchanged` bit is set.** `git status` is blind to it
  by design. Rare, operator-inflicted, and named here so it is a known limit rather than a
  surprise.

### finish-contract-run2.md — Worktree cleanup, verify each removal (KAN-859)

Moved verbatim (KAN-859), where it followed "leave that worktree's entry in `worktrees`.":

Writing `worktrees: {}` regardless would drop it from
  the only authoritative list, and nothing would ever find it again.

### finish-contract-run2.md — the remote branch, the shell's comment (KAN-859)

Moved verbatim (KAN-859) from the top of the remote-delete shell block:

```bash
# `push --delete` exits non-zero BOTH when the branch was already gone and when the push was
# refused, so the two are told apart by git's message and never by the exit code alone. Measured
# against a scratch remote on 2026-08-31 (git 2.50.1): an already-absent branch prints
# `error: unable to delete 'spectre/<name>': remote ref does not exist` and exits 1, and the
# stale remote-tracking ref SURVIVES that failure.
```

### finish-contract-run2.md — the remote branch, no prompt (KAN-859)

Moved verbatim (KAN-859), where it followed "**The remote branch is deleted without a further prompt.**":

- **The remote branch is deleted without a further prompt.** Run 2 is reached only by proving the
  branch is an ancestor of the base branch, so its commits are in the base branch and nothing can be
  lost — which is why this is not gated the way check 4's disclosure is.

### finish-contract-run2.md — the remote branch, a refused push (KAN-859)

Moved verbatim (KAN-859), where it followed "**A refused push is reported, never swallowed.**":

- **A refused push is reported, never swallowed.** A bare `|| true` would make an expired
  credential indistinguishable from a branch the forge already removed, and leave the remote branch
  standing with nothing said about it.

### finish-contract-run2.md — the remote branch, not gated (KAN-859)

Moved verbatim (KAN-859), where it followed "**The remote delete is not gated on the local one succeeding.**":

- **The remote delete is not gated on the local one succeeding.** Gating it would leave the remote
  branch behind whenever anything unrelated failed, which is the state this step exists to end.

### finish-contract-run2.md — the stack-stopped check (KAN-859)

Moved verbatim (KAN-859); the run-loaded copy keeps the sentence up to "the other checks":

When the key
or the file is absent the check is **skipped, not failed**, and cleanup proceeds on the strength of
the other checks — check 6 among them, which is what makes an undeclared `## stop` key survivable:
a project that declares nothing to stop is still checked for a process running from its worktree.
