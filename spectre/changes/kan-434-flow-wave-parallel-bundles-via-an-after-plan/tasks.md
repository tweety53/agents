# kan-434-flow-wave-parallel-bundles-via-an-after-plan

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

## Global constraints

- Every `scripts/check-*.sh` guard named in `.flow/project.md`'s `## lint` exits clean after
  every task; a guard is never weakened and no suppression is added.
- Guard scripts stay Python 3, standard library only.
- Every structural pattern for the new field is defined once in `scripts/lib/plan_grammar.py`;
  no guard carries its own regex copy.
- Prose edits cut, never paraphrase; normative sentences, exit-code contracts and
  rejected-alternative rationales are never reworded.
- A task commit never stages `<project>/spectre/changes/` or `<project>/docs/superpowers/`.

- [x] 1. `**After:**` field grammar in `scripts/lib/plan_grammar.py`

**Build:** green
**After:** none

**Files:**
- Modify: `scripts/lib/plan_grammar.py`

**Tests:** **none** — pure grammar addition; its behavior is pinned by the harness cases tasks 2
and 3 add, which exercise `after_ids` and `select_after` through `plan-dispatch-bundles.py` and
`check-plan-shape.py`
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `feat(plan-grammar): add the After field grammar`

  - [x] **Step 1: Add the field and value patterns beside the `SQUASH_WITH_*` pair**

  In `scripts/lib/plan_grammar.py`, directly below `SQUASH_WITH_VALUE_RE` / `PARTNER_ID_RE`, add:

```python verified:mirrors SQUASH_WITH_FIELD_RE / SQUASH_WITH_VALUE_RE in scripts/lib/plan_grammar.py, read this session
# AFTER_FIELD_RE — the field's PRESENCE on one physical line, same
# presence/gate split as SQUASH_WITH_FIELD_RE.
AFTER_FIELD_RE: Pattern[str] = re.compile(
    r"^\*\*After:\*\*\s*(?P<value>.*)$"
)

# AFTER_VALUE_RE — the GATE. An `After:` value is `Task <ids>` or the
# literal `none`, and nothing else.
AFTER_VALUE_RE: Pattern[str] = re.compile(
    r"^(?:Task\s+[\d.,\s]+|none)\s*$"
)
```

  Add `after_ids(value) -> Optional[List[str]]` mirroring `partner_ids`: `None` when the value
  does not gate, `[]` when it gates as `none`, otherwise the id list from
  `PARTNER_ID_RE.findall`. Add `AFTER_VALUE_RE` to the module docstring's "What this module
  defines" enumeration.

  - [x] **Step 2: Add `select_after`, mirroring `select_squash_with`**

  Add an `AfterField(NamedTuple)` (`offset`, `value`, `ids`) and `select_after(body:
  Sequence[str]) -> Optional[AfterField]` with `select_squash_with`'s exact discipline: fence
  tracking via `FENCE_RE`, line-scoped (one physical line's remainder, never a continuation),
  the first non-fenced gating line wins, and when no candidate gates the FIRST candidate is
  returned with `ids=None` so the malformed-report path can name the line. Its docstring states
  that the serial default (absent field = every plan-order earlier task) is applied where
  after-sets are resolved (`scripts/plan-dispatch-bundles.py`), never in the grammar — the
  grammar records only what the plan declares.

  - [x] **Step 3: Verify and commit**

  Run: `python3 -c "import sys; sys.path.insert(0, 'scripts/lib'); import plan_grammar;
  assert plan_grammar.after_ids('Task 2, 3') == ['2', '3'] and
  plan_grammar.after_ids('none') == [] and plan_grammar.after_ids('Task sometime') is None"`
  Expected: no assertion error.
  Run: `scripts/check-vocabulary.sh && scripts/check-references.sh`
  Expected: both exit 0.

```bash unverified:confirm exact paths against the worktree at execution time
git add scripts/lib/plan_grammar.py
git commit -m "feat(plan-grammar): add the After field grammar"
```

---

- [x] 2. Bundle after-sets in `scripts/plan-dispatch-bundles.py`

**Build:** green
**After:** Task 1

**Files:**
- Modify: `scripts/plan-dispatch-bundles.py`
- Modify: `scripts/test-plan-dispatch-bundles.sh`

**Tests:** `scripts/test-plan-dispatch-bundles.sh` — cases 18–23 added (declared set printed,
default expansion, `none` literal, red-pair union, resumed-plan checked task, mixed-width
after-set numeric ordering)
**Regression:** reverting this task's `plan-dispatch-bundles.py` change while keeping the new
cases fails cases 18–22 — no `after <k>:` line is printed and declared `After:` fields are not
resolved
**Baseline:** before=17 after=23 on `scripts/test-plan-dispatch-bundles.sh`
<!-- measured: scripts/test-plan-dispatch-bundles.sh @ branch spectre/kan-434-flow-wave-parallel-bundles-via-an-after-plan -->
<!-- predicted: scripts/test-plan-dispatch-bundles.sh after this task -->
**Commit:** `feat(plan-dispatch): print each bundle's resolved after-set`

  - [x] **Step 1: Write the failing cases**

  Extend `scripts/test-plan-dispatch-bundles.sh`, following its existing fixture/run_guard
  pattern, with five cases:

  1. Tasks declaring `**After:** Task 1` print their resolved set:
     `bundle`/`after` lines for a two-task plan where task 2 declares `Task 1`.
  2. A task with NO `**After:**` field resolves to every plan-order earlier id (a three-task
     plan, no fields anywhere, prints `after` lines `none`, `1`, `1 2`).
  3. `**After:** none` resolves to an empty set, printed as `after <k>: none`, and puts the task
     in the first wave's position (its after line lists nobody even though earlier tasks exist).
  4. A `**Squash-with:**` red pair where the red task declares `Task 1` and its green partner
     declares nothing: the shared bundle's after-set is the union (task 1 plus every
     partner-earlier id).
  5. A resumed plan (task 1 checked `- [x]`) where task 2 declares no field: task 1 is not
     bundled, task 2's after-set still resolves to `1`.

  - [x] **Step 2: Confirm RED**

  Run: `scripts/test-plan-dispatch-bundles.sh`
  Expected: FAIL on cases 18–22.

  - [x] **Step 3: Implement resolution and output**

  In `scripts/plan-dispatch-bundles.py`:

  - Add `After` to `ANY_FIELD_RE`'s alternation so an `**After:**` line ends a Files bullet run.
  - In `parse_tasks`' `close_body`, read the field via `select_after`/`after_ids` (imported from
    `lib/plan_grammar.py`): a gated value stores the declared ids or `[]` for `none`; a
    non-gating value is reported like a missing `**Files:**` field is today — exit 1, one
    `<path>:<task line>: task <id> has a malformed **After:** value: <value>` line per task —
    and stores nothing.
  - In `compute_bundles`, resolve each bundle's after-set as the union over its members: a
    member with declared ids contributes them; a member with NO field contributes every
    task id that precedes it in document order (checked tasks included); a member with `none`
    contributes nothing; deduplicate and sort numerically.
  - Extend the printed output: beneath each existing `bundle <k>: <ids>` line (byte-identical to
    today), print `after <k>: <ids space-separated>` or `after <k>: none` when the set is empty.
  - Update the module docstring's exit-0 contract to describe both line shapes and the new exit-1
    finding.

  - [x] **Step 4: Update the existing cases' expected output**

  All 17 existing cases assert exact stdout; every one now also carries an `after` line. Update
  each case's `EXPECTED` to the new two-line shape — the assertions' bundle content is
  unchanged.

  - [x] **Step 5: Confirm GREEN and commit**

  Run: `scripts/test-plan-dispatch-bundles.sh`
  Expected: PASS, 22 cases green.
  Run: `scripts/check-guard-symlinks.sh`
  Expected: exit 0 (the script's imports of `lib/plan_grammar.py` are already declared).

```bash unverified:confirm exact paths against the worktree at execution time
git add scripts/plan-dispatch-bundles.py scripts/test-plan-dispatch-bundles.sh
git commit -m "feat(plan-dispatch): print each bundle's resolved after-set"
```

---

- [x] 3. Shape findings F7–F10 in `scripts/check-plan-shape.py`

**Build:** green
**After:** Task 1

**Files:**
- Modify: `scripts/check-plan-shape.py`
- Modify: `scripts/check-task-commit-fields.py`
- Modify: `scripts/test-check-plan-shape.sh`
- Modify: `scripts/test-check-task-commit-fields.sh`

**Tests:** `scripts/test-check-plan-shape.sh` — cases 20–25 added (F7 duplicate, F8 non-gating
value, F9 dangling id, F10 cycle, a clean `After:`-using plan, an absent-field plan staying
clean); `scripts/test-check-task-commit-fields.sh` — case 90 added (an `**After:**` line adjacent
to a `**Files:**` bullet run leaves Files parsing intact)
**Regression:** reverting this task's guard changes while keeping the new cases fails shape cases
20–25 (no After finding fires) and commit-fields case 90 (the After line disturbs the parsed
`**Files:**` set)
**Baseline:** before=19 after=25 on `scripts/test-check-plan-shape.sh`; before=89 after=90 on
`scripts/test-check-task-commit-fields.sh`
<!-- measured: scripts/test-check-plan-shape.sh and scripts/test-check-task-commit-fields.sh @ branch spectre/kan-434-flow-wave-parallel-bundles-via-an-after-plan -->
<!-- predicted: both harnesses after this task -->
**Commit:** `feat(plan-shape): validate the After field's references and order`

  - [x] **Step 1: Write the failing cases**

  Shape harness (`scripts/test-check-plan-shape.sh`), cases 20–25:

  1. F7 — a task body with two gating `**After:**` lines: one finding at the second line.
  2. F8 — a task whose `**After:**` value is free text (`Task sometime`): one malformed-value
     finding at that line.
  3. F9 — `**After:** Task 9` in a two-task plan: one finding naming the absent task.
  4. F10 — task 2 declares `**After:** Task 5` and task 5 declares no field: one cycle finding
     naming both tasks (task 5's serial default closes the cycle); also a direct self-reference
     `**After:** Task 2` on task 2: one finding.
  5. A clean plan using `Task <ids>` and `none` declarations: exit 0.
  6. A plan with no `**After:**` fields at all: exit 0, output byte-identical to before this
     task.

  Commit-fields harness (`scripts/test-check-task-commit-fields.sh`), case 90: a task whose
  `**Files:**` bullet run is immediately followed by `**After:** Task 2` — the declared Files set
  is unchanged and the commit check passes as if the line were absent.

  - [x] **Step 2: Confirm RED**

  Run: `scripts/test-check-plan-shape.sh && scripts/test-check-task-commit-fields.sh`
  Expected: FAIL on cases 20–25 and case 90 respectively (field boundary first: case 90's Files
  parse is disturbed until `FIELD_RE` knows the field).

  - [x] **Step 3: Register the field in `check-task-commit-fields.py`**

  Add `After` to `FIELD_RE`'s alternation (one word, keeping the anchored
  `^\*\*(...):\*\*\s*(.*)$` shape). `parse_task_fields` then treats the line as a field boundary
  and `check-plan-shape.py`'s `INDENTED_FIELD_RE`, derived from `FIELD_RE.pattern`, covers an
  indented copy under F3a automatically. `**After:**` stays optional — no check requires it.

  - [x] **Step 4: Add F7–F10 to `scripts/check-plan-shape.py`**

  - Import `select_after`/`after_ids` from `lib/plan_grammar.py`.
  - F7: per body, every non-fenced gating `**After:**` line after the first is a finding
    (`task <id> declares a second **After:** line (first at line <n>); select_after keeps the
    first`), the F1 message discipline.
  - F8: a body whose first `**After:**` candidate does not gate (`task <id> has a
    malformed **After:** value: <value>`), reported at that line.
  - F9: an id in any task's declared set that names no task in the plan (`task <id>'s **After:**
    names task <n>, which this plan does not define`).
  - F10: resolve every task's after-set exactly as `plan-dispatch-bundles.py` does (absent =
    every plan-order earlier id, `none` = empty), DFS over the resolved edges, and report each
    cycle back-edge (`task <a> -> task <b> -> task <a>: the resolved After graph has a cycle, so
    neither bundle can ever dispatch`); a self-reference is the one-node cycle. Findings are
    canonical in the module docstring's Findings list, which this step extends to F10.
  - All four skip a task whose fence never closes (F3b's existing early return).

  - [x] **Step 5: Confirm GREEN, run the repo's own plans through the guard, and commit**

  Run: `scripts/test-check-plan-shape.sh && scripts/test-check-task-commit-fields.sh`
  Expected: PASS, 25 and 90 cases green.
  Run: `scripts/check-plan-shape.sh`
  Expected: exit 0 — every non-archived plan in this repository, this change's own included,
  passes the new findings (this plan's own `**After:**` lines are among the validated input).

```bash unverified:confirm exact paths against the worktree at execution time
git add scripts/check-plan-shape.py scripts/check-task-commit-fields.py \
  scripts/test-check-plan-shape.sh scripts/test-check-task-commit-fields.sh
git commit -m "feat(plan-shape): validate the After field's references and order"
```

---

- [x] 4. Wave dispatch in `skills/flow/implement.md` §4

**Build:** green
**After:** Task 2

**Files:**
- Modify: `skills/flow/implement.md`

**Allowed-collateral:** `scripts/check-contract-budget.sh` (raise this file's row only if Step 5
measures the edited file over it)

**Tests:** **none** — procedure prose; the guards below verify shape, budget and vocabulary
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow): dispatch ready bundles in waves`

  - [x] **Step 1: Capture the normative inventory, then add the wave-dispatch paragraph**

  Run `scripts/check-normative-inventory.sh > /tmp/normative-before.txt` and keep the output.
  Then, in §4 after the bundling paragraph ("Dispatch one implementer per bundle from…"), add
  the wave rule: readiness (a bundle is ready when every id in its `after <k>:` line has landed —
  committed and guard-passed, by direct commit or pick); a bundle alone in a wave dispatches into
  the canonical worktree and commits directly, exactly as today; two or more ready bundles launch
  in one message, each into its own throwaway worktree created by the sequence below, each copy
  then running the project's resolved `## worktree setup` command once before its implementer
  dispatches:

```bash verified:the worktree-add/transplant sequence quoted from skills/flow/review-panel.md, The throwaway worktree
git -C <worktree> worktree add --detach <worktree>-wave-bundle-<k> HEAD
git -C <worktree> diff HEAD --binary | git -C <worktree>-wave-bundle-<k> apply --allow-empty
git -C <worktree> status --porcelain -z | \
  while IFS= read -r -d '' entry; do
    st="${entry:0:2}"; f="${entry:3}"
    [ "$st" = "??" ] || continue
    mkdir -p "<worktree>-wave-bundle-<k>/$(dirname "$f")"
    cp -a "<worktree>/$f" "<worktree>-wave-bundle-<k>/$f"
  done
```

  - [x] **Step 2: Add the pick, guard and handback rules**

  After the launches paragraph, state: as wave members return, each is cherry-picked onto the
  change branch in plan order — a member is picked once every plan-earlier member of its wave is
  picked; the unchanged `check-task-commit-fields.sh` call (canonical worktree fifth argument,
  resolved `<name>` sixth) runs on each picked commit and the dispatch `end` records the picked
  sha; a pick conflict or a guard failure hands the bundle back to its own implementer — its
  throwaway worktree rebased onto the advanced branch HEAD, re-commit, re-pick — while sibling
  members and already-ready later waves are unaffected; a copy is removed once its bundle is
  picked, or after handback resolves. A member reporting BLOCKED follows the existing BLOCKED
  handback. The one-implementer-per-worktree rule is untouched: each wave member has its own
  worktree.

  - [x] **Step 3: Add the full-suite exception**

  In the `FULL SUITE` paragraph, append one sentence: when the plan-last bundle belongs to a
  shared wave, its implementer does not carry `FULL SUITE` — the conductor instead runs the
  resolved `## test` list once on the canonical worktree after that wave's final pick passes the
  guard, and a failure is the same verbatim-output `## Question` handback as below. The
  existing last-boundary sentence about a full-suite failure report keeps governing the
  singleton case.

  - [x] **Step 4: Verify the prose guards**

  Run: `scripts/check-references.sh && scripts/check-dispatch-paragraphs.sh && scripts/check-stage-mark-calls.sh`
  Expected: all exit 0 — no required dispatch paragraph (REPRODUCE DON'T READ, FOREGROUND
  BUILDS) was touched, every new path citation resolves.
  Run: `diff <(scripts/check-normative-inventory.sh) /tmp/normative-before.txt`
  Expected: the only differences are sentences this task deliberately added; any difference in a
  pre-existing sentence is resolved by restoring that sentence.

  - [x] **Step 5: Measure the budget and commit**

  Run: `wc -c skills/flow/implement.md`
  Expected: at most 32500 — the file measures 29117 bytes before this task, its budget row allows
  32500, and the wave paragraphs must fit inside the 3383-byte headroom or the row is raised to
  the new actual size plus 25% in `scripts/check-contract-budget.sh` (the ratchet's own rule)
  in this same commit.
<!-- measured: wc -c skills/flow/implement.md and grep budgets() scripts/check-contract-budget.sh @ branch spectre/kan-434-flow-wave-parallel-bundles-via-an-after-plan -->

```bash unverified:confirm exact paths against the worktree at execution time
git add skills/flow/implement.md
git commit -m "docs(flow): dispatch ready bundles in waves"
```

---

- [x] 5. `**After:**` field in `skills/flow/brainstorm-planner.md`'s plan template

**Build:** green
**After:** none

**Files:**
- Modify: `skills/flow/brainstorm-planner.md`

**Allowed-collateral:** `scripts/check-contract-budget.sh` (raise this file's row only if Step 2
measures the edited file over it)

**Tests:** **none** — procedure prose
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow): teach the plan template the After field`

  - [x] **Step 1: Add the field to §D's field family**

  In section **D**'s mechanically-checkable field list, after the `**Squash-with:**` sentence,
  add: an optional `**After:**` field, `Task <ids>` or `none`, whose absence means the task runs
  after every earlier task; the planner writes it on file-disjoint tasks with no caller/helper
  relationship and writes it consistently across a `**Squash-with:**` pair (union semantics
  merge the pair into one bundle); a task the planner does not annotate stays fully serial, so
  opting in is per task.

  - [x] **Step 2: Measure the budget, verify, and commit**

  Run: `wc -c skills/flow/brainstorm-planner.md`
  Expected: at most 23248 — the file measures 19792 bytes before this task, its row allows
  23248, and the addition must fit inside the 3456-byte headroom or the row is raised to the new
  actual size plus 25% in this same commit.
  Run: `scripts/check-references.sh && scripts/check-contract-budget.sh`
  Expected: both exit 0.
<!-- measured: wc -c skills/flow/brainstorm-planner.md and grep budgets() scripts/check-contract-budget.sh @ branch spectre/kan-434-flow-wave-parallel-bundles-via-an-after-plan -->

```bash unverified:confirm exact paths against the worktree at execution time
git add skills/flow/brainstorm-planner.md
git commit -m "docs(flow): teach the plan template the After field"
```
