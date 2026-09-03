# kan-318-myflow-a-red-green-task-pair-pays-for-context — design

## Context

KAN-318: a `Build: red` task and its `Squash-with:` green partner are dispatched as two
implementers doing one unit of work. The red dispatch reads the plan, the specs, the principles and
the surrounding code to write failing tests; the green dispatch re-reads the same material minutes
later to implement against them. KAN-212 ran 5 red-partner dispatches for 30.5M tokens, 15% of its
subagent spend, essentially all of it context the green partner re-acquired.
<!-- measured: dispatches table, reported in docs/self-review/kan-212-persist-per-dispatch-cost-tokens-model-and-role-self-review.md @ main 01959d3 -->

The operator chose "bundler joins on `Squash-with:`" over "retire `Build: red` at planning time"
and "resume the same agent for the green half". No staged research note existed
(`docs/superpowers/research/kan-318.md`, `kan-318-*.md` absent). `spectre/specs/` is empty and
`openspec/` is frozen, so no capability spec is edited.

Measured against `main` at `01959d3`:

- The contract already states the pair is one unit and the mechanism never delivered it.
  `skills/flow-contracts/build-green.md` — a red task's partner is one "that this task is
  dispatched together with as a single unit"; `openspec/specs/myflow-build-green/spec.md:24` —
  "the task it is dispatched with as one unit". `scripts/plan-dispatch-bundles.py` joins tasks on
  `**Files:**` overlap only and deliberately reads no `Squash-with:` (its own constants comment:
  "this guard reads no partner"). A red task declares the test file and its partner the source
  file, so the pair rarely shares a path — KAN-212's pairs 1→2, 3→4, 5→6 each split into two
  bundles.
- `skills/flow/implement.md` §4 then folds the two commits after the fact — `git commit
  --fixup=<partner-task-sha>` plus `git rebase --autosquash` — and records the red dispatch under
  `-role red-partner`.
- `scripts/check-task-commit-fields.py` (`resolve_folded_task`) already checks a fold from either
  id against the one folded commit: the partners' agreed `Commit:` subject, the union of every
  folded task's `Files:`. A pair that produces one commit directly is the state it already
  expects; no guard change.
- Review independence: today the red task's per-task review reads a test-only diff and the green
  partner's review reads the folded commit — tests and implementation — minutes later. The red
  review is a strict subset (KAN-319). Merging the pair loses nothing the panel sees.
- `stats/cmd/flow/record.go:47` — `recordRoles` is a closed list that includes `red-partner`;
  `record_test.go` pins it. The store column is plain `TEXT`.
- `scripts/lib/plan_grammar.py` exports `iter_tasks` and `select_squash_with`, imported by
  `check-task-build-green.py` via `sys.path.insert(0, <script dir>/lib)`;
  `skills/flow/scripts/lib` is a symlink to `scripts/lib`, so the same import works from the
  installed path.

## 1. The bundler joins a red task with its partners

`scripts/plan-dispatch-bundles.py` adds one edge class to its union-find: for every unchecked task
whose body carries a `**Squash-with:**` field that gates (`plan_grammar.select_squash_with(...)`
returns `partners` that is not `None`), union the task with each named partner id that is itself an
unchecked task in the plan.

```python verified:plan_grammar.select_squash_with(body: Sequence[str]) and TaskBody.lines, scripts/lib/plan_grammar.py @ 01959d3
sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "lib"))
from plan_grammar import iter_tasks, select_squash_with

# in compute_bundles, after the by_path joins:
for body in iter_tasks(lines):
    if body.id not in unchecked_ids:
        continue
    field = select_squash_with(body.lines)
    if field is None or field.partners is None:
        continue
    for partner in field.partners:
        if partner in unchecked_ids:
            uf.union(body.id, partner)
```

What this does not do, deliberately:

- A partner that does not exist, is checked, is itself red, or is written dotted joins nothing.
  Those are `check-task-build-green.py`'s violations and the bundler is not a second guard for
  them; it computes a grouping over the ids it has. A red task whose partner is already `[x]` forms
  its own bundle, as it does today.
- The `Files:` join, the `Allowed-collateral:` exclusion, exit codes and output shape are
  unchanged. Exit 1 still names a task missing `**Files:**`.
- The module docstring's "this guard reads no partner" sentence and `plan_grammar.py`'s comment
  that the bundler "mirrors TASK_LINE_RE rather than importing it" are corrected, since both become
  false.

`scripts/test-plan-dispatch-bundles.sh` gains one case: a red task declaring `a_test.go` with
`**Squash-with:** Task 2`, a green task 2 declaring `a.go`, a green task 3 declaring `b.go` —
expected output `bundle 1: 1 2` then `bundle 2: 3`. A second case pins the non-join: the same
plan with task 2 marked `[x]` yields `bundle 1: 1` then `bundle 2: 3`.

## 2. `skills/flow/implement.md` §4 — one implementer, one commit

The paragraph "**A `Build: red` task's commit folds into its green partner.**" (the
`--fixup`/`--autosquash` instruction) and the paragraph "**A `Build: red` task's own dispatch
records `-role red-partner`, not `implementer`.**" are deleted. In their place, one paragraph after
the `plan-dispatch-bundles.sh` exit-code paragraph:

> **A `Build: red` task is dispatched with its `Squash-with:` partner, in one bundle.** The
> implementer runs the red task first — writes its tests, runs them, and reports the failing
> output — then the green partner, and makes **one** commit for the pair: the partner's declared
> `**Commit:**` subject and `Task-Id: <partner>` trailer. The red task never has a commit of its
> own; `check-task-commit-fields.sh` resolves the pair from either id against that commit.

The COMMIT-PER-TASK blockquote's "one commit per task" sentence gains the qualifier "a red task
and its partner make one commit between them"; the sentence "a `Build: red` task still folds into
the commit its `**Squash-with:**` field names" is reworded to "a `Build: red` task is bundled with,
and commits with, the partner its `**Squash-with:**` field names".

The `-role` sentence drops `red-partner`: "`-role` is one of `implementer`, `reviewer`, `panel-fix`
or `verifier`". `recordRoles` in `stats/cmd/flow/record.go` keeps `red-partner` so old rows and
any journaled record still validate — no Go change.

**Review.** "bundle N's reviewers is one reviewer per task in it" becomes "one reviewer per
*commit* in it — a red task and its partner share one". The tick rule gains: a red task's checkbox
is ticked together with its partner's when their one commit passes review. The fix-round fold
(`git commit --fixup=<task-sha>` on a reviewer's finding) is untouched — that is a fold between a
task's commit and its own fix, not between two tasks.

## 3. `skills/flow-contracts/build-green.md`

One clause added to the `Build: red` bullet: "…dispatched together with as a single unit —
`plan-dispatch-bundles.py` places a red task in its partner's bundle — and whose commit this
task's work lands in." The guard-scope section is unchanged.

## 4. Relationship to KAN-319

KAN-319 ("a red task's per-task review duplicates its green partner's") is not implemented here.
It stops existing as a defect once section 2 lands: with one commit per pair and one reviewer per
commit, there is no separate red review to duplicate. The proposal names this so KAN-319 can be
closed against this change rather than planned.

## Testing

- `scripts/test-plan-dispatch-bundles.sh` — the two new cases plus the existing fourteen.
- `scripts/test-check-task-commit-fields.sh` and `scripts/test-check-task-build-green.sh` —
  unchanged and must still pass, since the guards are untouched.
- The project's `## lint` list, in particular `check-dispatch-paragraphs.sh` (the implementer site
  keeps every required paragraph), `check-contract-budget.sh` (`implement.md` shrinks;
  `build-green.md` grows by one clause under its 6678-byte budget) and
  `check-normative-inventory.sh`.

## Decisions

### bundler-joins-on-squash-with

**ID:** bundler-joins-on-squash-with
**Status:** active
**Chosen:** add a `Squash-with:` edge to `plan-dispatch-bundles.py`'s union-find and have the one
implementer run RED then GREEN and commit once — the contracts already state the pair is dispatched
as one unit, so this closes a drift rather than adding a rule.
**Considered:** retiring `Build: red` at planning time, writing each pair as one task — rejected:
touches both build-green guards, the commit-fields fold logic and every document naming the tag,
and loses the plan-level record of which tests were written first. Resuming the red implementer via
`SendMessage` for the green half — rejected: keeps two dispatch records and the fold choreography,
which cost KAN-212 a correction round when the fold had to step over an intervening fix commit.

### red-partner-role-kept-in-store

**ID:** red-partner-role-kept-in-store
**Status:** active
**Chosen:** `red-partner` stays in `recordRoles`; `implement.md` stops naming it.
**Considered:** removing it from `recordRoles` and the pinned test — rejected: a journaled
`red-partner` record from an older run would be refused on replay, for the gain of deleting one
string.

### one-reviewer-per-commit

**ID:** one-reviewer-per-commit
**Status:** active
**Chosen:** a bundle's per-task review is one reviewer per commit; a red task and its partner share
one, and are ticked together.
**Considered:** keeping one reviewer per task, the red reviewer reading the same folded commit —
rejected: that is KAN-319's duplication reproduced by construction.

## Open questions

None.
