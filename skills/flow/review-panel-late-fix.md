# Review panel — the late-fix reduction

Loaded during `flow.review-panel` by the load directive under **The late-fix reduction**
(`skills/flow/review-panel.md`), only on a fix run, before the stage opens its first round. Every
section name below without a path is a section of `skills/flow/review-panel.md`.

## The late-fix reduction

**A fix run whose delta is small against an already-clean, already-verified stage reduces pass 1
to `primary` alone reading the delta only** — the narrow late-fix review path. Like the docs-only reduction this only ever removes, and
it fires only on a **fix run** — a `/flow` invocation whose argument is fix instructions — never
on a creating run, whose pass 1 is always the full resolved roster. Every condition must hold,
checked once when this stage opens its first round on the fix run, by one call — the canonical
worktree's triple first:

```bash
check-late-fix-trigger.sh <change> <changeRoot>/tasks.md <worktree> <since-close-sha|-> <base-verdict> [<worktree> <since-close-sha|-> <base-verdict>…]
```

Its header (`<agents repo>/scripts/check-late-fix-trigger.sh`) is canonical for the five
conditions. `<since-close-sha>` is the sha this stage's last clean close on this change reviewed
in that worktree, the pass log or the rendered panel record naming that close — `-` where it has
none; the `-diff-base` its clean dispatches recorded names the canonical worktree's sha, and a
peer worktree's since-close sha comes from the panel record's per-worktree sha list.
`<base-verdict>` is that worktree's **Check base movement first** line this round. Exit 0 → the
reduction fires; exit 1 → the full path, each failed condition on its own line; exit 2 → it cannot
answer, read as the full path.

**On trigger, pass 1 is one dispatch: `primary` alone**, plus every slot the operator named at
this stage's start, reading not the whole `final-review.diff` but only
`<abs-worktree>/.superpowers/sdd/late-fix.diff`, written over the since-close range by
`write-panel-diff.sh late-fix <abs-worktree> <worktree> <since-close-sha> [<worktree> <since-close-sha>…]`,
whose exits read as the pass-1 write's. On a decided panel the dispatch runs on the decision's `panel.rerun_dispatch`
pair under the fix-round re-run's 5-minute ceiling; on a `default` panel, which carries no
decision to read a pair from, on `opus` at `low` effort under the ordinary 15-minute
ceiling. Mutation and Failure-modes are not dispatched, and each dropped slot is recorded
with `flow record pass -round <round> -note 'not dispatched — late-fix reduction: <slot>'`, the
docs-only reduction's own convention. Every entry check above still runs as any round's — base
movement, the diff-size cap, the docs-only guard — and where both reductions fire, the late-fix
one governs the read scope, the roster being `primary` alone either way. The reduction itself is
recorded with `flow record pass -round <round>`:
`late-fix reduction: <n> changed lines since <sha>`.

A finding the targeted dispatch raises feeds the ordinary fix-round loop unchanged, and a Minor-only
result is fixed inline under the standing rule — but **any Critical or Important it raises voids the reduction
for the rest of the run**: a delta that small producing a defect that severe means the narrow
read's context was not enough, so every later round this run opens takes the full path. When the
targeted dispatch reads the delta clean, or raises only Minors, fixed inline under the standing
rule, the close sha moves to the round's HEAD and the stage closes under the existing rules —
the reduction is a read-scope decision, never a weaker close.

**The staleness carve-out.** Beside the stale definition's rule that a fix against which a slot
raised no finding leaves that slot's result current (**Panel re-runs**, `skills/flow/review-panel.md`):

The same holds for a delta the late-fix reduction's targeted dispatch read — clean,
or with every finding it raised a Minor the parent fixed inline under the standing rule:
`primary`'s read on the since-close range leaves every slot it did not dispatch current —
the reduction's own conditions, already clean and already verified with the machinery untouched,
being what the full roster's coverage rests on for that delta.
