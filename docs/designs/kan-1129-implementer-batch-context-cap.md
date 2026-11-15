# kan-1129-implementer-batch-context-cap
> Design record of change kan-1129-implementer-batch-context-cap (KAN-1129), written at landing. Each decision's own **Status:** line is the status.

## Context

The approved design is `.superpowers/sdd/planner/round-3.md` in this worktree (rounds 1–3, approved
in round 3). The operator settled three points, verbatim:

- Round 1, context bound: "Option 1 but cap at 3" — the task cap is the bound; no mid-run
  mechanism.
- Round 2, cap shape: "The cap must be smart. 3 small to medium subtasks is ok. 3 big are not.
  Then approve" — the cap is weighted by task size.
- Round 3, size source: "A. New optional Size field (Recommended)"; design approved.

**No declared task field tracks an implementer's final context**, so the size is stated per task.
Measured on the four plans still on disk, joined with `flow record dispatches` (kan-1087's own
tasks.md is gone):

| Change | Dispatch | Per task (Files, Tests, plan lines) | context_end |
|---|---|---|---|
| kan-1124 | task-2+3+6+7+8+9 | (4,4,30) (10,5,19) (4,2,15) (40,2,14) (25,2,12) (13,3,12) — 96 files | 152,859 <!-- measured: grep over kan-1124's tasks.md fields joined with flow record dispatches -change kan-1124 context_end @ live flowd store, read during this planning (not re-runnable at a ref) --> |
| kan-1124 | task-1+14..18 | (5,2,33) (6,1,28) (23,0,12) (19,0,12) (29,0,12) (7,8,22) — 89 files | 221,388 <!-- measured: grep over kan-1124's tasks.md fields joined with flow record dispatches -change kan-1124 context_end @ live flowd store, read during this planning (not re-runnable at a ref) --> |
| kan-1122 | task-7 | (6,0,260) — longest task text | 81,783 |
| kan-1122 | task-3+5+6 | (2,1,20) (5,2,109) (4,1,187) | 155,099 |
| kan-1017 | task-4+5+6+7 | (10,5,14) (7,6,24) (9,6,19) (7,3,17) | 171,489 |
| kan-1017 | task-8 | (11,4,16) | 101,771 |

`**Files:**` count, `**Tests:**` count and task text length each invert against context; kan-1087's
single-task dispatches spread 71k–375k on the difficulty of the work alone.

This is one change because the cap, the field it reads and the contract text that names both
must land together: a cap without the field cannot run a big task alone, and a field the guard
does not read is decoration.

## Decisions

### Size-aware weight cap on implementer groups

**ID:** group-weight-cap
**Status:** active
**Chosen:** a third pass in `flow-guard plan-dispatch-groups`, after the fold: each group is cut,
in plan order, into consecutive pieces whose weight stays within `pdgMaxWeight = 3`. A task
weighs 3 when its `**Size:**` is `big`, otherwise 1; a bundle weighs the sum of its tasks. A
bundle joins the current piece when the piece's weight plus the bundle's stays within 3,
otherwise it opens the next piece. Pieces are printed as groups, ordered by lowest bundle id —
deterministic, carried by `groups_mechanical`; `decision.json`, `stampGroupTasks` and
`offDecisionGroups` are unchanged because a piece is a group, and pieces of one lane wait on each
other through the existing ready rule of **Waves** (`skills/flow/sdd-dispatch.md`).
**Considered:** a flat cap of 2 tasks (round 1's auto-taken value — overridden by the operator to
3); a flat cap of 3 tasks (round 1's answer — superseded by round 2's answer: three big tasks
must not share an implementer); a Decide-step judgement rule for the planner (the existing
`groups_override` split already failed to fire in kan-1087); a record-time refusal in
`stampGroupTasks` (would refuse an unsplittable `**Squash-with:**` bundle).

### The cap is the context bound

**ID:** cap-is-context-bound
**Status:** active
**Chosen:** the weight cap alone bounds an implementer's context; no mid-run mechanism.
**Considered:** a context hand-back rule with a continuation dispatch (`-cont-<k>` key) or a hook
watching context size — the operator chose the cap alone ("Option 1 but cap at 3").

### Where a task's size comes from

**ID:** task-size-field
**Status:** active
**Chosen:** a new optional `**Size:**` task field the planner writes, parsed in `pdbParseTasks`
beside `**Files:**` — the planner who wrote the task is the one source that knows its difficulty.
**Considered:** inferring size from the `**Files:**` count (e.g. 8+ files = big) — the measurement
in `## Context` shows it would run a wide mechanical sweep alone and pack hard narrow tasks
together, the opposite of the intent.

### A big task runs alone

**ID:** big-runs-alone
**Status:** active
**Chosen:** `big` weighs 3 against a budget of 3.
**Considered:** `big` = 2 — would pair a big task with a small one, while a single big kan-1087
task already reached 323k–375k alone.

### One weight budget, no separate task count

**ID:** weight-only-budget
**Status:** active
**Chosen:** one budget of weight 3.
**Considered:** a weight budget plus a task-count cap — with small and medium at weight 1, weight
3 already means at most 3 tasks.

### Small and medium weigh alike

**ID:** small-equals-medium
**Status:** active
**Chosen:** `small` and `medium` both weigh 1.
**Considered:** `small` = 0.5 — the operator's words put them together ("3 small to medium
subtasks is ok").

### An unmarked task is medium

**ID:** size-absent-is-medium
**Status:** active
**Chosen:** a task without `**Size:**` weighs 1.
**Considered:** requiring the field on every task — existing plans and `/flow-fast` plans would
stop parsing and a new check guard would be needed.

### An unknown size is a plan violation

**ID:** size-invalid-exit-1
**Status:** active
**Chosen:** a `**Size:**` value other than `small`, `medium` or `big` is a violation, exit 1, in
`plan-dispatch-bundles`' violation list beside a malformed `**After:**`, and passes through
`plan-dispatch-groups` unchanged.
**Considered:** reading it as medium — a typo of `big` would silently pack three heavy tasks.

### The plan gate refuses a bad size too

**ID:** size-checked-at-plan-gate
**Status:** active
**Chosen:** `check-plan-shape` gains rule F32 — a task's `**Size:**` value is `small`, `medium` or
`big` — so a bad value fails at the plan gate, not first at dispatch; `plan-dispatch-bundles` keeps
its own violation (`size-invalid-exit-1`) for plans that never passed the gate. Auto-taken from
plan review 1 (F1).
**Considered:** dispatch-time only — the plan gate owns field validation, and a typo found after it
costs a fix round on an approved plan.

### A duplicated or fenced size line

**ID:** size-duplicate-and-fence
**Status:** active
**Chosen:** `Size` joins `tcfFieldRE`'s names and `psF1Fields`, so a second `**Size:**` line in one
task is an F1 violation and an indented one an F3a violation; `pdbParseTasks` reads `**Size:**` only
at column 0 outside a fence, as it reads `**Files:**`. Auto-taken from plan review 1 (F2).
**Considered:** first line wins — a silent pick hides a contradictory plan.

### An oversized bundle stays whole

**ID:** oversized-bundle-stays-whole
**Status:** active
**Chosen:** a single bundle heavier than 3 (a `**Squash-with:**` pair holding a big task) is one
group.
**Considered:** refusing it — a `**Squash-with:**` pair must commit together.

### The implementer runs the commit guards before hand-back

**ID:** implementer-precheck-commit
**Status:** active
**Chosen:** `skills/flow/implementer-brief.md`'s commit paragraph tells the implementer to run
`flow-guard check-task-commit-fields` and `flow-guard check-task-commit-planning-paths` on its own
commit and amend before handing back, so the parent's `close-task` exit 1 and its
`-recommit-<k>` fixer stop being those guards' first run.
**Considered:** a new `flow-guard task-verify` wrapper — no second caller; the guards already
exist.

### A search that needs context is one call

**ID:** brief-search-with-context
**Status:** active
**Chosen:** one rule in the brief's OUTPUT BUDGET paragraph: a search whose hit needs its
surrounding lines is one call — `grep -n -C<k>` or the Grep tool with context — never a grep
followed by a `sed -n` window.
**Considered:** a new search wrapper tool — one brief rule covers it.

### Fixer grouping is unchanged

**ID:** fixer-grouping-unchanged
**Status:** active
**Chosen:** gated fix (≤20 findings) and panel-fix chunks (≤10) stay finding-count based.
**Considered:** capping fixers by tasks too — the cap already shrinks what a gated fixer
inherits, and the ticket names implementers.

### Full-suite fixes are out of scope

**ID:** full-suite-fix-out-of-scope
**Status:** active
**Chosen:** the 19 `full-suite-fix` causes are recorded in proposal.md's `## Why` as gymie-side;
no task.
**Considered:** planning gymie fixes — a repository this change was not given.

### No cost-per-task CLI

**ID:** no-cost-per-task-cli
**Status:** active
**Chosen:** the ticket's before/after figure is computed from `flow record dispatches` JSON.
**Considered:** a new `flow record` subcommand — no second caller.

### Size correlation measured inline

**ID:** explore-inline-plan-agents-skipped
**Status:** active
**Chosen:** the field-to-context correlation was measured directly (four plans, one join) and no
Plan agents were dispatched.
**Considered:** Explore and Plan dispatches — one bounded query, and the sizing alternatives are
the two under `task-size-field`.

## Open questions

## Recommended approach

Add `size string` to `pdbTask`, read from a `**Size:**` line in `pdbParseTasks`, and report an
unknown value through `pdbBundlesOf`'s violation list (exit 1). In `pdgComputeGroups`, after the
fold, walk each folded group's bundle ids in order, summing each bundle's weight (3 per `big`
task, 1 per other task), and start a new piece whenever adding the next bundle would take the
current non-empty piece past `pdgMaxWeight = 3`; sort all pieces by lowest bundle id as today. The
planner learns when to write `big` from one new field bullet, Decide step 5 replaces "no per-group
ceiling" with the cap, and the implementer brief gains the commit pre-check and the
search-with-context rule. Rejected alternatives: Files-count inference (`task-size-field`), a
flat cap (`group-weight-cap`), a mid-run hand-back (`cap-is-context-bound`), a planner judgement
rule or a record-time refusal (`group-weight-cap`), new wrapper tools
(`implementer-precheck-commit`, `brief-search-with-context`, `no-cost-per-task-cli`).

## Critical files

- `stats/internal/guard/planparse.go` — `pdbTask` and `pdbParseTasks`: gain the `**Size:**` field.
- `stats/internal/guard/planparse.go` — `tcfFieldRE` gains `Size`.
- `stats/internal/guard/planshape.go` — `psF1Fields` gains `Size`; new rule F32 (`size-checked-at-plan-gate`).
- `stats/internal/guard/check_plan_shape_test.go` — F32, F1 and F3a cases for `**Size:**`, beside
  the other per-rule tests (`TestCheckPlanShapeF29` and its like).
- `stats/internal/guard/plandispatchbundles.go` — `pdbBundlesOf`'s violation list: gains the
  unknown-size violation; its header's exit-1 list names it.
- `stats/internal/guard/plandispatchgroups.go` — `pdgComputeGroups`: gains the weight-cap pass;
  the header states the third pass and the weights, and drops the "fewer, larger groups" bias
  wording it contradicts.
- `stats/internal/guard/plan_dispatch_groups_test.go` — cases for the cap and the field.
- `stats/internal/guard/plan_dispatch_bundles_test.go` — the unknown-size violation case.
- `skills/flow/brainstorm-planner.md` — section D's task-field list gains `**Size:**`; Decide step
  5 states the cap in place of "no per-group ceiling".
- `skills/flow/sdd-dispatch.md` — **Dispatch one implementer per group** names the cap.
- `skills/flow/implementer-brief.md` — the commit paragraph's pre-check and the OUTPUT BUDGET
  paragraph's search-with-context rule (its current "`grep -n` for the symbol, then `sed -n`"
  advice for reading a large file by range is reconciled with it).
- `README.md` — the class table's `big` row ("one per task group") names the cap.

## Existing code to reuse

- `pdbBundlesOf` (`stats/internal/guard/plandispatchbundles.go`) — bundles and the violation
  path the size check joins.
- `tcfIterTasks` and the `**Files:**`/`**After:**` field reading in `pdbParseTasks`
  (`stats/internal/guard/planparse.go`) — the field grammar `**Size:**` follows.
- `pdgComputeGroups`' chain and fold passes (`stats/internal/guard/plandispatchgroups.go`) — the
  cap is a pass over their output.
- `psFixture` (`stats/internal/guard/check_plan_shape_test.go:503`) and `runGuard`
  (`stats/internal/guard/check_base_moved_test.go:113`) — test fixture and runner.
- `flow-guard check-task-commit-fields`, `flow-guard check-task-commit-planning-paths` — the
  guards the implementer now runs itself.

## Verification

- `cd stats && go test ./internal/guard -run 'TestPlanDispatch'`, covering:
  - a 5-task serial chain with no `**Size:**` yields pieces of 3,2; a 7-task chain 3,3,1; a 3-task
    chain stays one group;
  - `small`, `medium` and absent weigh alike: a chain of small, medium, (absent) stays one piece;
  - a chain small, big, small yields pieces 1,1,1 (the big task alone; the small after it does not
    join it);
  - a chain big, big, big yields three pieces of one task;
  - a chain small, small, big, small yields 2,1,1;
  - an oversized single bundle (a `**Squash-with:**` pair with one task `big`) stays one group;
  - `**Size:** huge` exits 1 with a violation naming the task and the field;
  - existing cases pass, the expectations of those whose groups weigh more than 3 — case 1,
    case 4, and both plans of `TestPlanDispatchGroupsKeepsRepositoriesApart` — restated to the
    capped pieces.
- The project's `## lint` commands from `.flow/project.md` exit clean.
- The ticket's before/after comparison: `flow record dispatches -change <n>` gives
  `metrics.cost_usd` and `metrics.signals.sidechain.context_end` per dispatch, tasks counted from
  the key's `+` segments; the before figures are kan-1087's above, the after figures come from
  the next comparable run.

## Live check

`flow-guard plan-dispatch-groups` is a CLI over a plan file; no daemon or store holds its state, so
the real thing is the built guard run on a plan reproducing kan-1087's nine-task chain. Write a
fixture tasks.md under the change's state dir (`flow state dir`): nine unchecked tasks in plan
order, task `<i>` declaring `**Files:** \`f<i>.txt\`` and no `**After:**`, tasks 3 and 7 also
carrying a blank line then `**Size:** big`. Run the worktree's `scripts/flow-guard
plan-dispatch-groups <fixture>`.

- Before (base): `group 1: 1 2 3 4 5 6 7 8 9`, exit 0.
  <!-- measured: scripts/flow-guard plan-dispatch-groups on that fixture @ merge-base f3e458be (the grouping BEFORE this change) -->
- After (branch): `group 1: 1 2`, `group 2: 3`, `group 3: 4 5 6`, `group 4: 7`, `group 5: 8 9`,
  exit 0.
  <!-- predicted: the same command on the branch after task 3 -->
- Not working looks like: any group naming four or more tasks, task 3 or 7 sharing a group, or
  the base's single nine-task line unchanged.
- Then change task 3's value to `huge`: exit 1, stderr naming task 3 and its unknown
  **Size:** value.

## Non-goals

- Fixer grouping changes (`fixer-grouping-unchanged`).
- A `flow-guard task-verify` wrapper or a cost-per-task stats CLI (`implementer-precheck-commit`,
  `no-cost-per-task-cli`).
- Size inference from `**Files:**` or `**Tests:**` counts (`task-size-field`).
- Deriving the size weights from recorded `context_end` per `**Size:**` (plan review 1, F3,
  advisory) — the weights stay hand-set until a later run's data asks otherwise.
- A mid-run context mechanism (`cap-is-context-bound`).
- Gymie-side full-suite causes (`full-suite-fix-out-of-scope`).

## Risks

- Size is the planner's judgement: a task left unmarked that turns out big still shares a piece
  with two others and can run past 300k. The next comparable run's `context_end` per piece, read
  against each task's `**Size:**`, is the data to tune the weights from.
- The cap cannot cut inside a task: one big task alone can still reach ~375k (kan-1087 task-58).
- Three small or medium tasks do not hold the ticket's ~200k target: kan-1087's 3-task dispatches
  ended ~305k. Accepted by the operator (round 2).
- More groups means more dispatches, each paying its context-bundle start (~100k+); a plan of
  many tiny tasks pays more starts.
- The cap reverses part of KAN-513's "fewer, larger groups" bias; that design's reasoning lives in
  a gitignored kan-513 change dir, so the new rule is stated in the guard header and Decide
  step 5.
