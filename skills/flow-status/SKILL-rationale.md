# flow-status — rationale

Reasoning behind `skills/flow-status/SKILL.md`: which guard each prose step mirrors, where a rule's
derivation lives, and the incident a rule was written against. Moved here verbatim from the
run-loaded file; **no `/flow-status` run loads this file.** Each heading names the source file and
the section the passage came from.

## SKILL.md — 1. List open changes

> Step 2 below still
> reimplements `check-finish-preflight.sh`'s other merge-status steps in prose rather than invoking
> that script (see the note there); only base-branch resolution is delegated to a real guard.

## SKILL.md — 2. Resolve each change's state

> Read it there — it is not re-derived here, and it is not
> re-derived from `<agents repo>/scripts/check-finish-preflight.sh` either: that script's resolve-first guard and
> its comment (b) are what `pipeline.md` cites in turn.

## SKILL.md — 4. Detail view (only when `<name>` was given)

On the run-only fields the handoff-blocks contract marks:

> and this file does not list them again

On the one-way `prUrl` test:

> Do not restate that
> reasoning here, and do not present the test as conclusive.

On the `Running:` section:

> Do not restate the resolution *procedure* here — the steps that
> compute each app root, start command and URL; a second copy of those steps is the failure this
> repository's contracts are built to avoid, and naming the invariant above is not one.

On reusing step 2's merge status rather than re-deriving it:

> Reading `prUrl` in front of that answer is what let one invocation
> report *branch merged → it will archive* in the table and *waiting on the merge* in the block, for
> the same change, in the same run — a change stopped at a run-2 cleanup leftover is exactly that
> case, and it is not rare.

> - **The two splits do not compete.** The next-command column splits `IN_PROGRESS` on merge
> status to say which bare `/flow` run the operator gets; the block splits on it to say which
> wait the operator is in. Both read the same signal first, so they cannot disagree about the
> branch — and because both blocks end in `/flow <name>`, neither can contradict the other
> about what to run next.
