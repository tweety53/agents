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
flow-guard check-late-fix-trigger <change> <changeRoot>/tasks.md <worktree> <since-close-sha|-> <base-verdict> [<worktree> <since-close-sha|-> <base-verdict>…]
```

Its header (`<agents repo>/scripts/flow-guard check-late-fix-trigger`) is canonical for the five
conditions; its plan-grew condition reads `<project>/.superpowers/sdd/<name>/tasks-at-close.md` in the canonical worktree,
the copy the review panel's stage close writes (`skills/flow/review-panel.md`). `<since-close-sha>` is the sha this stage's last clean close on this change reviewed
in that worktree, the pass log or the rendered panel record naming that close — `-` where it has
none; the `-diff-base` its clean dispatches recorded names the canonical worktree's sha, and a
peer worktree's since-close sha comes from the panel record's per-worktree sha list.
**After integrate's reshape** (**3. Commit the staged work**, `skills/flow/integrate.md`) the shas
the last clean close reviewed are gone from the branch, so the since-close sha is each worktree's
reshaped branch head.
`<base-verdict>` is that worktree's **Check base movement first** line this round. Exit 0 → the
reduction fires; exit 1 → the full path, each failed condition on its own line; exit 2 → it cannot
answer, read as the full path; exit 3 → **The append scope** below — the full path for the
roster, never for the read scope; exit 4 → **The small append** below.

**On trigger, pass 1 is one dispatch: `primary` alone**, beside **The always-on extra dispatches** (`skills/flow/review-panel.md`), which still run, plus every slot the operator named at
this stage's start, reading not the whole `final-review.diff` but only
`<abs-worktree>/.superpowers/sdd/late-fix.diff`, written over the since-close range by
`flow-guard write-panel-diff late-fix <abs-worktree> <worktree> <since-close-sha> [<worktree> <since-close-sha>…]`,
whose exits read as the pass-1 write's. On a decided panel the dispatch runs on the decision's `panel.rerun_dispatch`
pair, its Role line `panel-rerun`, under the fix-round re-run's 5-minute ceiling; on a `default`
panel, which carries no decision to read a pair from, on the `late-fix` row of **Allowed pairs**
(`skills/flow-contracts/model-policy.md`) under the ordinary 15-minute ceiling. Each dropped slot is recorded
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
being what the full roster's coverage rests on for that delta. A delta the small append's dispatch
read — clean, or with every finding it raised a Minor the parent fixed inline under the standing
rule — leaves every slot it did not dispatch current the same way.

## The append scope

**A fix run whose late-fix check fails on its size or scope-growth condition alone — exit 3 —
keeps the decided roster and narrows only the read scope.** Pass 1 runs the roster a full pass 1
would: the same `panel.dispatches`, pairs and ceilings, and on a `default` panel its
settings-store grouping. Every dispatch reads `<abs-worktree>/.superpowers/sdd/late-fix.diff`,
written over the since-close range by the same `flow-guard write-panel-diff late-fix` call the late-fix
reduction makes, whose exits read as the pass-1 write's, and rendered with `-diff late-fix`,
never the whole `final-review.diff` — the render points every shared paragraph at that diff.
Mutation, which reads no diff file, takes each worktree's since-close sha as its diff base in
place of the merge base its MUTATION ENTRY CONTEXT names.

Every entry check still runs as any round's, against the merge base — base movement, the
diff-size cap and the docs-only guard, whose reduction to `primary` still applies. The append scope is
recorded with `flow record pass -round <round>`:
`append scope: <n> changed lines since <sha>`.

A Critical or Important finding it raises feeds the ordinary fix-round loop and **voids nothing**:
the full roster already read the delta, so a severe finding says nothing about the read being too
narrow. A clean close moves the close sha to the round's HEAD, as any close does.

## The small append

**A fix run whose late-fix check exits 4 — the append scope's verdict on a countable delta of at
most 250 changed lines — runs pass 1 as one dispatch: `primary` alone**, beside **The always-on extra dispatches** (`skills/flow/review-panel.md`), which still run, plus every slot the
operator named at this stage's start, on the `late-fix` row of **Allowed pairs**
(`skills/flow-contracts/model-policy.md`) under the ordinary 15-minute ceiling, reading `late-fix.diff` written and rendered as **The append scope** writes and renders
it. `primary` reads every line of the new source, so nothing appended goes unreviewed; what the
small append drops is the second lens and the higher effort. It runs on `opus`, never the rerun pair: the delta is new
source, a first read rather than a re-run. Every other slot pass 1 would carry — the experimental
slot included — is not dispatched, each recorded with
`flow record pass -round <round> -note 'not dispatched — small append: <slot>'`, and the small
append itself with `flow record pass -round <round>`:
`small append: <n> changed lines since <sha>`. Every entry check still runs as the append scope's.

A Critical or Important the small append raises feeds the ordinary fix-round loop and leaves every
slot it did not dispatch stale, so each runs on the next round under **Panel re-runs**; a Minor-only result is fixed inline under the standing rule,
and a clean close moves the close sha to the round's HEAD.

## The fix-run judge

**A fix run decides its panel roster and its verify mode once, through one judge held above a
mechanical floor.** When this stage opens its first round on the fix run, after the late-fix
check above and in the same Bash call, run — in a change whose resolved set holds one worktree;
a wider set takes the full path, with no floor and no judge:

```bash
flow-guard fix-run-floor <worktree> <merge-base>
```

Its header (`<agents repo>/scripts/flow-guard fix-run-floor`) is canonical for the class, the
floor each class carries, the verdict file and the exit codes. `<merge-base>` is the one the
docs-only guard reads. Exit 2 is reported with its stderr and takes the full path. A
`floor_roster: full` line leaves the judge nothing to decide, so none is dispatched: record
`flow record pass -round <round> -note 'fix-run judge: not dispatched — floor full (<class>)'`.
A `floor_roster: primary` line dispatches the judge, once per fix run and never again within it.

**The judge dispatch.** Its pair is its **Allowed pairs** row
(`skills/flow-contracts/model-policy.md`), mapped on harnesses `zcode` and `muse` as that file's
**Harness mapping** maps it. Its prompt opens with the line `Role: fix-run-judge` and carries,
verbatim:

> Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task.
> Include this instruction verbatim in any prompt you write for another agent.

**The prompt also carries the TOOLS paragraph**:

> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,
> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch in your first turn,
> before anything else. Never ToolSearch for a tool already listed, and never a wildcard query: a
> schema loaded later changes your tool list and re-prices your whole context at full input rate.

**The prompt also carries the NO DELEGATION paragraph**:

> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and never spawn a
> subagent, background agent or helper of any kind — you are the leaf of this run, and any child
> you start is unrecorded and outside the parent's closed list (**Dispatch sites — the
> parent's closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and
> fixing are your own Read, Bash and Edit calls.

**The prompt also carries the MODEL HANDSHAKE paragraph**:

> **MODEL HANDSHAKE:** the first line of your first reply is `Model: <the model named in your own
> system prompt>` and nothing else on that line. Answer it before any tool call.

It also states: the absolute worktree path; the fix instructions, verbatim; the output of
`git -C <worktree> diff --stat <merge-base>`; the three lines `flow-guard fix-run-floor` printed;
the roles this round's pass 1 would dispatch without a verdict — the roster the reductions above
leave, the always-on extra dispatches' roles and every slot the operator named — one id per
line; the absolute path of the rendered panel record under
`<abs-worktree>/.superpowers/sdd/reviews/`, whose findings are the earlier rounds'; and this
task, as written:

> Decide what this fix run's review panel and verify still have to run. The floor is the least
> you may choose and you may choose more, never less. Answer with one JSON object,
> `{"roster": [<role id>, …], "verify": "fast"|"full", "reasons": [<one sentence each>]}`.
> `roster` lists, from the role ids you were given, the roles pass 1 dispatches: `primary`
> always, every role the fix instructions name, and any other role whose lens the fix gives
> something to find — a role an earlier finding of the same kind came from, a fix that changes
> what a guard, a test or a dispatched prompt does rather than how a sentence reads. `verify` is
> `fast` — the lint list alone — only when no test could read the changed text differently, and
> `full` otherwise. Give one reason per choice that departs from the floor, and one for staying
> on it. You edit nothing and run no test. Your last act is writing that object, alone, to
> `<abs-worktree>/.superpowers/sdd/fix-run-judge-<n>.json`, and your reply after the `Model:`
> line is the same object.

`<n>` is this fix run's ordinal, the one `flow.document-fix` recorded.

**The verdict passes through the guard, never around it:**

```bash
flow-guard fix-run-floor -judge <abs-worktree>/.superpowers/sdd/fix-run-judge-<n>.json -round <round> <worktree> <merge-base>
```

`<round>` is this stage's first round on the fix run. Exit 0 puts the verdict in force. Exit 1 —
a verdict below the floor, or no JSON — and exit 2 take the full path. A judge that has not
answered within 5 minutes, or whose agent dies, is closed `-outcome aborted` and takes the full
path too.

**Recording.** The parent records the dispatch `-role planner`, `-task` omitted, on its pair's
model and effort, `-key fix-run-judge-<n>`, and compares its `Model:` line as the handshake
paragraph of **Dispatch sites — the parent's closed list** (`skills/flow/implement.md`) states. The outcome is one row of the pass log:
`flow record pass -round <round> -note 'fix-run judge: <the guard's verdict: line> · floor <class> — <the reasons>'`,
or `fix-run judge: full path — <the guard's stderr, or aborted>`.

**What a verdict in force changes, and what holds it there:**

- **The roster.** Pass 1 dispatches a role only when the verdict's roster names it. The verdict
  is applied after the reductions above, to the roster and to the always-on extra dispatches
  alike, and only ever removes: a bundle dispatches the roles it still names, on its own pair
  and read scope, and a bundle left naming none is not dispatched. Each dropped role is recorded
  with `flow record pass -round <round> -note 'not dispatched — fix-run judge: <slot>'`.
  `flow-guard panel-round-begins` refuses the round's begins — exit 3, none recorded — while a
  slot carries a role the roster omits.
- **Verify.** Under `"verify": "fast"`, `flow-guard run-verify-list <worktree> test` skips every
  command and ends its row `skipped`; the `lint` group runs whole.
- **The void.** Any Critical or Important finding this fix run's panel raises voids the verdict
  for the rest of the run, as it voids the late-fix reduction: every role the verdict dropped is
  stale and runs on the next round under **Panel re-runs**, and verify runs `full` —
  `flow-guard run-verify-list` reads the finding from the store and runs the whole list unasked.
  A clean or Minor-only pass 1 leaves every role the verdict dropped current, as the late-fix
  reduction's carve-out above leaves its own.
- **The tree.** Both guards re-read the class from the tree: a verdict stops governing the moment
  a touched path is no longer prose.
