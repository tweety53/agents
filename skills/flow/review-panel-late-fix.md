# Review panel — the late-fix reduction

Loaded during `flow.review-panel` by the load directive under **The late-fix reduction**
(`skills/flow/review-panel.md`), only on a fix run, before the stage opens its first round. Every
section name below without a path is a section of `skills/flow/review-panel.md`.

## The late-fix reduction

**A fix run whose delta is small against an already-clean, already-verified stage reduces pass 1
to `primary` alone reading the delta only** — the narrow late-fix review path. Like the docs-only reduction this only ever removes, and
it fires only on a **fix run** — a `/flow` invocation whose argument is fix instructions — never
on a creating run, whose pass 1 is always the full resolved roster. Every condition must hold,
checked once when this stage opens its first round on the fix run:

1. **Already clean, already verified** — the change's store findings carry no `open` row
   (`check-panel-findings-closed.sh <worktree> <change>` exits 0), and this stage has closed
   clean before on this change, the pass log or the rendered panel record naming that close;
2. **the base has not moved since that close** — every worktree's `Check base movement first`
   verdict this round was `CLEAR`, so no rebase has invalidated the reviewed state;
3. **the delta is small** — per worktree, `git diff --numstat` from the sha that close reviewed
   there to the current tree, insertions and deletions summed across the resolved set, is at
   most **40 changed lines**; the `-diff-base` its clean dispatches recorded names the canonical
   worktree's sha, and a peer worktree's since-close sha comes from the panel record's
   per-worktree sha list;
4. **no scope growth** — the delta adds no task line to the change's plan `tasks.md`;
5. **the panel's own machinery is untouched** — the delta names no path under
   `skills/flow/review-panel*.md`, no `scripts/check-panel-*.sh` guard, and no reviewer-prompt or
   principles file the panel's own slots read as their instructions
   (`skills/flow/*-reviewer-prompt.md`, `skills/flow/engineering-principles.md`) — of the
   repository the change edits.

**On trigger, pass 1 is one dispatch: `primary` alone**, plus every slot the operator named at
this stage's start, reading not the whole `final-review.diff` but only a
`<abs-worktree>/.superpowers/sdd/late-fix.diff` written from the since-close range — the same
per-worktree sectioned shape as `final-review.diff`, each `# worktree:` header naming the
since-close sha. On a decided panel the dispatch runs on the decision's `panel.rerun_dispatch`
pair under the fix-round re-run's 5-minute ceiling; on a `default` panel, which carries no
decision to read a pair from, on `opus` at `low` effort under the ordinary 15-minute
ceiling. Mutation and Failure-modes are not dispatched, and each dropped slot is recorded
with `flow record pass -round <round> -note 'not dispatched — late-fix reduction: <slot>'`, the
docs-only reduction's own convention. Every entry check above still runs as any round's — base
movement, the diff-size cap, the docs-only guard — and where both reductions fire, the late-fix
one governs the read scope, the roster being `primary` alone either way. The reduction itself is
recorded with `flow record pass -round <round>`:
`late-fix reduction: <n> changed lines since <sha>`.

A finding the targeted dispatch raises feeds the ordinary fix-round loop unchanged, and Minors
defer under the standing rule — but **any Critical or Important it raises voids the reduction
for the rest of the run**: a delta that small producing a defect that severe means the narrow
read's context was not enough, so every later round this run opens takes the full path. When the
targeted dispatch reads the delta clean, or raises only Minors that defer under the standing
rule, the close sha moves to the round's HEAD and the stage closes under the existing rules —
the reduction is a read-scope decision, never a weaker close.

**The staleness carve-out.** Beside the stale definition's rule that a fix against which a slot
raised no finding leaves that slot's result current (**Panel re-runs**, `skills/flow/review-panel.md`):

The same holds for a delta the late-fix reduction's targeted dispatch read — clean,
or with every finding it raised a deferred Minor under the standing rule:
`primary`'s read on the since-close range leaves every slot it did not dispatch current —
the reduction's own conditions, already clean and already verified with the machinery untouched,
being what the full roster's coverage rests on for that delta.
