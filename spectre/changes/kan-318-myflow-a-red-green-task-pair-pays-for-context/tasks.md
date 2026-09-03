# kan-318-myflow-a-red-green-task-pair-pays-for-context

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when that task
> passes spec + quality review.
> **Relocation:** no

## Global constraints

- `design.md` in this change root is canonical for the bundling edge, the pair's commit shape, the
  review rule and the `red-partner` decision; the prose each task adds cites it by section.
- `~/.claude/rules/be-brief.md` — cut, never paraphrase. A normative sentence in
  `skills/flow/implement.md` or `skills/flow-contracts/build-green.md` is deleted or kept
  word-for-word; only the sentences design.md §2 and §3 name are replaced.
- Every guard named in `.flow/project.md`'s `## lint` exits clean after every task, including
  `scripts/check-contract-budget.sh` (byte budgets at `01959d3`: `skills/flow/implement.md` 25786
  of 32500, `skills/flow-contracts/build-green.md` 5345 of 6678), `scripts/check-references.sh`,
  `scripts/check-dispatch-paragraphs.sh` and `scripts/check-markdown-integrity.py`.
  <!-- measured: wc -c skills/flow/implement.md skills/flow-contracts/build-green.md; grep -n 'implement.md\|build-green.md' scripts/check-contract-budget.sh @ 01959d3 -->
- `scripts/test-check-task-build-green.sh` and `scripts/test-check-task-commit-fields.sh` pass
  unchanged after task 1 — the guards are not touched, and `scripts/lib/plan_grammar.py` gains only
  a comment edit.
- Standard library only in `scripts/plan-dispatch-bundles.py`; `lib/plan_grammar.py` is imported
  the way `scripts/check-task-build-green.py:157-159` imports it.

---

- [x] 1. Bundler joins a red task with its `Squash-with:` partners

**Build:** green

**Files:**
- Modify: `scripts/plan-dispatch-bundles.py`
- Modify: `scripts/test-plan-dispatch-bundles.sh`
- Modify: `scripts/lib/plan_grammar.py`

**Tests:** `case 15`, `case 16`
**Regression:** `case 15` — reverting the commit puts the red task and its partner in separate
bundles, so the output is `bundle 1: 1`/`bundle 2: 2`/`bundle 3: 3` and the expected-output
assertion fails. `case 16` — reverting the commit changes nothing for a checked partner, so the
case passes on either side; it pins the non-join so a later edit cannot make a checked task
re-enter a bundle through the new edge.
**Baseline:** before=27 after=33
<!-- measured: scripts/test-plan-dispatch-bundles.sh | grep -c '^ok:' @ 01959d3 (before); predicted: the same command after this task's commit, two assertions per new case (after) -->
**Commit:** `feat(scripts): bundle a red task with its Squash-with partners`

  - [x] **Step 1: RED — add cases 15 and 16 to the harness**

  In `scripts/test-plan-dispatch-bundles.sh`, directly after case 14's two assertions and before
  the `if [ "$FAILURES" -gt 0 ]` block, add, in the file's own fixture shape:

```bash verified:fixture and assertion shape copied from case 14 of scripts/test-plan-dispatch-bundles.sh @ 01959d3
# ===========================================================================
# Case 15: a `**Squash-with:**` field is a bundling edge. A red task declaring
# only its test file and the green partner it names declaring only the source
# file share no path, so the Files: join alone splits the pair — which is
# what dispatched KAN-212's three pairs as six implementers. The contract
# (skills/flow-contracts/build-green.md) has always said the pair is
# dispatched as one unit; this edge is what makes that true.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. Red: the failing tests\n\n'
  printf '**Build:** red\n\n'
  printf '**Squash-with:** Task 2\n\n'
  printf '**Files:**\n- Create: `a_test.go`\n\n'
  printf -- '  - [ ] **Step 1: write them**\n\n'
  printf -- '- [ ] 2. Green: the implementation\n\n'
  printf '**Build:** green\n\n'
  printf '**Files:**\n- Modify: `a.go`\n\n'
  printf -- '  - [ ] **Step 1: make them pass**\n\n'
  printf -- '- [ ] 3. Unrelated task\n\n'
  printf '**Build:** green\n\n'
  printf '**Files:**\n- Modify: `b.go`\n\n'
  printf -- '  - [ ] **Step 1: do it**\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 15: a Squash-with pair exits 0" || fail "case 15: rc=$RC out=$OUT"
EXPECTED=$'bundle 1: 1 2\nbundle 2: 3'
[ "$OUT" = "$EXPECTED" ] && pass "case 15: a red task bundles with the partner it names" || fail "case 15: expected [$EXPECTED], got [$OUT]"

# ===========================================================================
# Case 16: the edge reaches only UNCHECKED partners. A partner already
# marked `[x]` takes no part in any bundle — the same rule the Files: join
# applies — so a red task naming it forms its own bundle rather than
# dragging a done task back into dispatch.
# ===========================================================================
new_fixture
{
  printf -- '- [ ] 1. Red: the failing tests\n\n'
  printf '**Build:** red\n\n'
  printf '**Squash-with:** Task 2\n\n'
  printf '**Files:**\n- Create: `a_test.go`\n\n'
  printf -- '  - [ ] **Step 1: write them**\n\n'
  printf -- '- [x] 2. Green: the implementation, already done\n\n'
  printf '**Build:** green\n\n'
  printf '**Files:**\n- Modify: `a.go`\n\n'
  printf -- '  - [x] **Step 1: make them pass**\n\n'
  printf -- '- [ ] 3. Unrelated task\n\n'
  printf '**Build:** green\n\n'
  printf '**Files:**\n- Modify: `b.go`\n\n'
  printf -- '  - [ ] **Step 1: do it**\n'
} > "$TASKS_MD"
run_guard "$TASKS_MD"
[ "$RC" -eq 0 ] && pass "case 16: a red task naming a checked partner exits 0" || fail "case 16: rc=$RC out=$OUT"
EXPECTED=$'bundle 1: 1\nbundle 2: 3'
[ "$OUT" = "$EXPECTED" ] && pass "case 16: a checked partner joins nothing" || fail "case 16: expected [$EXPECTED], got [$OUT]"
```

  - [x] **Step 2: run the harness and see case 15 fail**

  Run: `scripts/test-plan-dispatch-bundles.sh`
  Expected: `FAIL: case 15: expected [bundle 1: 1 2 / bundle 2: 3], got [bundle 1: 1 / bundle 2: 2 /
  bundle 3: 3]` (newlines shown as `/`), the four case-16 and case-15 exit-0 assertions pass, and
  the harness exits 1 with `1 failure(s)`.

  - [x] **Step 3: GREEN — read the field and union on it**

  In `scripts/plan-dispatch-bundles.py`:

  1. After `import sys` add `import os`, and after the `typing` import add the lib import in the
     shape `scripts/check-task-build-green.py:157-159` uses:

```python verified:import shape read from scripts/check-task-build-green.py:157-159 @ 01959d3
sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "lib"))
from plan_grammar import select_squash_with
```

     No suppression marker on the import line — `scripts/check-task-build-green.py` carries none
     on its own, and this repository forbids adding one.

  2. Give `Task` one more field:

```python verified:dataclass shape read from scripts/plan-dispatch-bundles.py:133-141 @ 01959d3
    partners: List[str] = field(default_factory=list)
```

  3. In `parse_tasks`'s `close_body`, immediately after `body_lines = lines[body_start:end_index]`:

```python verified:plan_grammar.select_squash_with(body: Sequence[str]) -> Optional[SquashWithField] with .partners Optional[List[str]], scripts/lib/plan_grammar.py:217 @ 01959d3
        squash = select_squash_with(body_lines)
        if squash is not None and squash.partners is not None:
            current.partners = squash.partners
```

     `select_squash_with` tracks fences itself and returns `partners=None` for a value that does
     not gate — that task then carries no edge, which is the same "reads it but does not judge it"
     stance the file already takes toward `**Allowed-collateral:**`.

  4. In `compute_bundles`, after the `for task in unchecked: for path in task.files:` loop and
     before `groups`:

```python verified:union-find API read from scripts/plan-dispatch-bundles.py:232-250 @ 01959d3
    unchecked_ids = {t.id for t in unchecked}
    for task in unchecked:
        for partner in task.partners:
            if partner in unchecked_ids:
                uf.union(task.id, partner)
```

     A partner absent from the plan, checked, or written dotted is not in `unchecked_ids` and joins
     nothing; those are `check-task-build-green.py`'s findings, not this script's (design.md §1).

  5. Replace the module docstring's rule paragraph opening "Two unchecked tasks join the same
     bundle when they declare any common path under **Files:**; the relation is transitive." with:
     "Two unchecked tasks join the same bundle when they declare any common path under **Files:**,
     or when one carries a `**Squash-with:**` field naming the other (the build-green contract's
     'dispatched together as a single unit', `skills/flow-contracts/build-green.md`); the relation
     is transitive." Delete the sentence at lines 94-95 ("The dotted id grammar lives on in
     lib/plan_grammar.py for `Squash-with:` partner ids, and is deliberately not mirrored here —
     this guard reads no partner.") and write in its place: "`Squash-with:` is read through
     lib/plan_grammar.py's `select_squash_with`, the one grammar both build-green guards gate it
     with, so this script never extracts a partner id from free text."

  6. In `scripts/lib/plan_grammar.py:304-312`, the `TaskBody` comment's clause "the one guard that
     does read done-ness, plan-dispatch-bundles.py, mirrors TASK_LINE_RE rather than importing it
     and reads its own `state` group" stays true (the bundler keeps its own task parser for
     done-ness); append one sentence after it: "That script does import `select_squash_with`, so
     the field's line-scoping and gate are read one way by all three callers."

  Run: `scripts/test-plan-dispatch-bundles.sh`
  Expected: every assertion `ok:`, `all cases passed`, 31 `ok:` lines.

  - [x] **Step 4: the neighbouring harnesses and the lint list**

  Run: `scripts/test-check-task-build-green.sh && scripts/test-check-task-commit-fields.sh`
  Expected: both end `all cases passed`.

  Run: `scripts/plan-dispatch-bundles.sh spectre/changes/kan-318-myflow-a-red-green-task-pair-pays-for-context/tasks.md`
  Expected: exit 0, one bundle per unchecked task of this plan (no `Squash-with:` here).

  Run every `## lint` command from `.flow/project.md` in order. Expected: each exits 0.

  - [x] **Step 5: commit**

  `git add scripts/plan-dispatch-bundles.py scripts/test-plan-dispatch-bundles.sh scripts/lib/plan_grammar.py`,
  then commit with the subject this task's `**Commit:**` field states and a `Task-Id: 1` trailer.

- [x] 2. `implement.md` §4 — one implementer, one commit per pair

**Build:** green

**Files:**
- Modify: `skills/flow/implement.md`

**Tests:** **none** — prose only
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `feat(flow): dispatch a red task and its partner as one implementer`

  - [x] **Step 1: replace the fold and the red-partner paragraphs**

  In `skills/flow/implement.md` section 4, delete both paragraphs at lines 328-334 — the one
  opening "**A `Build: red` task's commit folds into its green partner.**" and the one opening
  "**A `Build: red` task's own dispatch records `-role red-partner`, not `implementer`.**" — and
  write in their place, verbatim from design.md §2:

```markdown verified:authored in design.md §2 of this change
**A `Build: red` task is dispatched with its `Squash-with:` partner, in one bundle.** The
implementer runs the red task first — writes its tests, runs them, and reports the failing
output — then the green partner, and makes **one** commit for the pair: the partner's declared
`**Commit:**` subject and `Task-Id: <partner>` trailer. The red task never has a commit of its
own; `check-task-commit-fields.sh` resolves the pair from either id against that commit.
```

  - [x] **Step 2: the bundling sentence and the role list**

  In the paragraph at lines 313-317, change "an implementer handed a bundle still makes one commit
  per task, carrying that task's own `Task-Id:` trailer, and a `Build: red` task still folds into
  the commit its `**Squash-with:**` field names" to "an implementer handed a bundle still makes one
  commit per task, carrying that task's own `Task-Id:` trailer — a red task and its partner make
  one commit between them — and a `Build: red` task is bundled with, and commits with, the partner
  its `**Squash-with:**` field names".

  At line 292, change "`-role` is one of `implementer`, `reviewer`, `panel-fix`, `red-partner` or
  `verifier`" to "`-role` is one of `implementer`, `reviewer`, `panel-fix` or `verifier`". Nothing
  else on that line changes.

  - [x] **Step 3: one reviewer per commit, ticked together**

  At line 370, change `"bundle N's reviewers" is one reviewer per task in it.` to `"bundle N's
  reviewers" is one reviewer per commit in it — a red task and its partner share one.`

  In the **Per-task review** paragraph (line 399 onward), after the sentence ending "a step's
  checkbox tracks the step and gates nothing." append: "A red task's checkbox is ticked together
  with its partner's, when their one commit passes review."

  - [x] **Step 4: guards**

  Run: `scripts/check-dispatch-paragraphs.sh && scripts/check-contract-budget.sh && scripts/check-references.sh && scripts/check-vocabulary.sh && scripts/check-markdown-integrity.py`
  Expected: each exits 0; `implement.md` is smaller than before, so the budget row is untouched.

  Run: `grep -n 'red-partner\|--fixup=<partner\|autosquash' skills/flow/implement.md`
  Expected: `red-partner` absent; the only `--fixup`/`autosquash` hits are the fix-round step at
  line 376 (and the panel-fix references beyond section 4, if any), which design.md §2 leaves
  untouched.

  - [x] **Step 5: commit**

  `git add skills/flow/implement.md`, then commit with the subject this task's `**Commit:**` field
  states and a `Task-Id: 2` trailer.

- [x] 3. `build-green.md` names the bundler as the mechanism

**Build:** green

**Files:**
- Modify: `skills/flow-contracts/build-green.md`

**Tests:** **none** — prose only
**Regression:** n/a — no test declared
**Baseline:** n/a — no test declared
**Commit:** `docs(flow-contracts): name the bundler as the red task's dispatch mechanism`

  - [x] **Step 1: the clause**

  In `skills/flow-contracts/build-green.md`, the `**Build:** red` bullet at lines 15-20 reads
  "…that this task is dispatched together with as a single unit and whose commit this task's commit
  folds into." Change it to "…that this task is dispatched together with as a single unit —
  `plan-dispatch-bundles.py` places a red task in its partner's bundle — and whose commit this
  task's work lands in." Reflow the bullet's lines to the file's existing width; no other sentence
  changes.

  - [x] **Step 2: guards**

  Run: `scripts/check-contract-budget.sh && scripts/check-references.sh && scripts/check-vocabulary.sh && scripts/check-markdown-integrity.py && scripts/check-installed-citations.sh`
  Expected: each exits 0; `build-green.md` stays under its 6678-byte budget.

  - [x] **Step 3: commit**

  `git add skills/flow-contracts/build-green.md`, then commit with the subject this task's
  `**Commit:**` field states and a `Task-Id: 3` trailer.
