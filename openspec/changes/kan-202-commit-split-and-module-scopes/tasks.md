# Tasks — commit scopes name the module

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

Eight tasks in four independent groups. The guard pair (1–2) and the rule pair (7–8) are each
test-first; the four text tasks (3–6) touch disjoint sentences and can run in any order among
themselves.

Every task's `**Commit:**` field below declares a **module scope**, which is what this change is
about. Tasks 3–8 land after task 2's guard exists, so each of their subjects is checked by the very
rule this change adds.

---

### 1 Scope-check cases in the commit-fields harness

**Build:** red

**Squash-with:** Task 2

**Files:**
- Modify: `scripts/test-check-task-commit-fields.sh`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: the executable statement of what `check_commit_scope` must refuse. Task 2 satisfies it.

Written before the guard so the cases assert what
`specs/myflow-task-commit-fields/spec.md`'s modified **Requirement: A runtime guard checks each
field against the real commit** says the guard must do, rather than what the implementation
happens to do.

Follow the harness's existing idiom exactly: a throwaway repository under `TMPDIR`, a synthesised
`openspec/changes/<name>/tasks.md`, a real commit whose subject matches the declared `**Commit:**`
field, then one `ok:`/`FAIL:` line per assertion. Do **not** invent a second fixture shape.

The change name must come from the fixture's directory, so each case's fixture directory is named
for the change whose scope it is testing.

- [x] **Step 1: The four refusal cases**

| Case | Declared `**Commit:**` | Fixture change name | Expects |
|------|------------------------|---------------------|---------|
| change-name scope | `feat(kan-900-some-change): add alpha` | `kan-900-some-change` | exit 1; message names the task and the scope |
| bare-key scope | `feat(kan-900): add alpha` | `kan-900-some-change` | exit 1 |
| numeric task id | `feat(3): add alpha` | `kan-900-some-change` | exit 1 |
| dotted task id | `feat(3.2): add alpha` | `kan-900-some-change` | exit 1 |

Each refusal case asserts the real commit's subject **matches** the declared field, so the failure
can only come from the new scope check and never from the existing `check_commit_subject`.

- [x] **Step 2: The three acceptance cases**

| Case | Declared `**Commit:**` | Expects |
|------|------------------------|---------|
| module scope | `feat(scripts): add alpha` | exit 0 |
| no scope at all | `feat: add alpha` | exit 0 — a scope is optional |
| scope merely containing the key | `feat(kan-900-helpers): add alpha` | exit 0 — the check is equality, not a substring match |

The third case is the one that stops the check from being written as `in` rather than `==`.

- [x] **Step 3: Confirm the harness fails against the unmodified guard**

The red half of this pair: run it before task 2 and confirm the four refusal cases fail loudly
rather than passing vacuously.

```bash verified:run against this tree before task 2 exists; the four refusal cases are absent, so the harness passes at its current count
bash scripts/test-check-task-commit-fields.sh 2>&1 | tail -1
```

**Tests:** the seven cases enumerated in steps 1 and 2, in `scripts/test-check-task-commit-fields.sh`.

**Regression:** Reverting this task leaves `check_commit_scope` with no executable statement of what
it refuses, so every prohibited scope shape becomes unasserted.

**Baseline:** before=29 after=45 assertions in `scripts/test-check-task-commit-fields.sh`.
<!-- measured: bash scripts/test-check-task-commit-fields.sh | grep -c '^ok:' @ branch main, before this change -->
<!-- measured: the same command @ branch openspec/kan-202-commit-split-and-module-scopes after the F1 fix round — eight cases, the eighth added by that round to pin equality against substring on the change-name branch -->

**Commit:** `test(check-task-commit-fields): pin which declared Commit scopes the guard must refuse`

---

### 2 The scope check itself

**Build:** green

**Files:**
- Modify: `scripts/check-task-commit-fields.py`
- Modify: `scripts/test-check-task-commit-fields.sh` — task 1's file. Task 1 is `Build: red` with
  `**Squash-with:** Task 2`, so its commit folds into this one and this commit is the squash unit's.

**Interfaces:**
- Consumes: task 1's cases.
- Produces: `check_commit_scope(task, change_name) -> List[str]`, called from `check_task_commit`
  beside the existing `check_commit_subject`.

- [x] **Step 1: Derive the change name from the tasks.md path**

`check_task_commit` already receives `tasks_md_path`, which the wrapper has resolved to
`<worktree>/openspec/changes/<name>/tasks.md`. The change name is that path's parent directory's
basename. No new argument, and no new call-site change in `check-task-commit-fields.sh`.

```python unverified:confirm the exact import list at the top of check-task-commit-fields.py before adding os.path usage
change_name = os.path.basename(os.path.dirname(os.path.abspath(tasks_md_path)))
```

- [x] **Step 2: Parse the scope out of the declared subject**

Parse `task.commit`, not the real commit's subject — the existing `check_commit_subject` already
proves the two agree, and checking the declared field is what makes a bad plan fail rather than a
bad implementer.

A subject is `<type>(<scope>): <rest>`. A subject with no parenthesised scope before the first
colon has no scope, and the check returns no violation.

- [x] **Step 3: Refuse the three prohibited shapes**

Fail when the parsed scope, compared by **equality** rather than by substring:

* equals the change name;
* equals the change name's leading Jira key — the change name up to and including its first
  numeric segment, e.g. `kan-202` from `kan-202-commit-split-and-module-scopes`. A change name with
  no such prefix yields no key, and this clause then matches nothing; or
* matches a dotted or numeric task id — digits separated by dots, and nothing else.

Anything else passes. Introduce **no** vocabulary of legal module names: a list of legal scopes
would need keeping in sync with the tree, and the guard's job is refusing three known-wrong shapes.

The violation message names the task, the offending scope, and which of the three shapes it hit,
following the phrasing of the existing violations in this file.

- [x] **Step 4: Wire it in and run the harness**

```bash unverified:confirm check_task_commit's violation-accumulation lines are still spelled this way after the edits above
grep -n 'violations += check_' scripts/check-task-commit-fields.py
```

```bash verified:this is the harness this repository already runs for this guard, listed under .myflow/project.md's ## test
bash scripts/test-check-task-commit-fields.sh
```

**Tests:** task 1's seven cases now pass; the four refusal cases fail the guard and the three
acceptance cases do not.

**Regression:** Reverting this task lets a plan declare `feat(<change-name>): …` and pass the guard,
which is the exact state this change exists to end.

**Baseline:** after=45 assertions in `scripts/test-check-task-commit-fields.sh`, all passing.
<!-- measured: bash scripts/test-check-task-commit-fields.sh @ branch openspec/kan-202-commit-split-and-module-scopes after the F1 fix round -->

**Commit:** `feat(check-task-commit-fields): reject a declared Commit scope naming the change or a task id`

---

### 3 The `**Commit:**` field spec gains the scope rule

**Build:** green

**Files:**
- Modify: `skills/myflow-start/SKILL.md`

**Interfaces:**
- Consumes: nothing.
- Produces: the planning-time statement of the rule task 2's guard enforces.

The `**Commit:**` entry in section **D**'s mechanically-checkable field list currently reads "the
commit subject line this task's implementer must use" and states nothing about the scope, which is
why every plan written under it declares a change-name scope.

- [x] **Step 1: State the scope rule on the field**

The scope names the module or area the commit moves, **derived from the paths in that same task's
`**Files:**` field**. Where a task spans modules, name the one carrying the substance or a broader
area covering them — never a list, never the change name, never the task id.

Say that a scope is optional and that an absent scope is correct where no single module carries the
task, so a writer does not manufacture one to satisfy a field.

- [x] **Step 2: Point at the guard**

One sentence naming `check-task-commit-fields.sh` as what checks this at `/myflow-do`, so a plan
writer knows a wrong scope is a build failure and not a style note. Cite the capability rather than
restating the three prohibited shapes — `specs/myflow-commit-scope/spec.md` is canonical for them.

```bash unverified:confirm the field list's exact line numbers have not shifted under earlier tasks in this run
grep -n 'the commit subject line this task' skills/myflow-start/SKILL.md
```

**Tests:** none — prose-only change to contract text, no executable surface. The citation this task
adds is validated by the references and installed-citations guards in the repository's lint list,
named without backticks here because this field's backticked tokens are parsed as declared test
names.

**Regression:** Reverting this task leaves the field spec silent on the scope, so the next plan
written declares a change-name scope and fails task 2's guard with no instruction saying why.

**Baseline:** unchanged — this task adds no test.

**Commit:** `docs(myflow-start): require a module scope on every task's declared Commit subject`

---

### 4 COMMIT-PER-TASK stops naming the task id as the scope

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md`

**Interfaces:**
- Consumes: nothing.
- Produces: an implementer dispatch block that no longer contradicts task 3.

The block currently says the subject follows the project's convention and then "where that
convention has a scope, `<n>` is the scope", with `fix(<n>)`, `feat(<n>)` and `task(<n>)` examples
built on it. `<n>` is the dotted task id.

- [x] **Step 1: Remove the task-id-as-scope instruction and its examples**

Keep everything else in the block intact: the `Task-Id: <n>` trailer, the never-weaken-commit-
validation rule, and the prohibition on committing `openspec/` or `docs/superpowers/`. Only the
scope sentence and the three examples resting on it go.

- [x] **Step 2: Point the subject at the plan's declared field**

Replace it with: the subject is the task's declared `**Commit:**` field, reproduced exactly — which
is already what `check-task-commit-fields.sh` enforces, so this states the existing contract rather
than adding one.

```bash unverified:confirm the sentence has not been reworded by an earlier task in this run
grep -n 'is the scope' skills/myflow-do/SKILL.md
```

**Tests:** none — contract text, covered by the repository's lint list as in task 3.

**Regression:** Reverting this task restores an instruction telling every implementer to write
`feat(3.2): …`, which task 2's guard now refuses — a dispatch block that guarantees a guard failure.

**Baseline:** unchanged — this task adds no test.

**Commit:** `docs(myflow-do): drop the task id as the commit scope from COMMIT-PER-TASK`

---

### 5 Finish run 1's two messages, at all four sites

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md`
- Modify: `skills/myflow-contracts/finish-contract.md`
- Modify: `skills/myflow-finish/SKILL.md`
- Modify: `skills/myflow-do/SKILL.md`
- Modify: `scripts/check-contract-budget.sh` — raising the ratchet row for
  `skills/myflow-contracts/finish-contract.md`, which this task's edits grow past its current
  budget. Raising a row for a file a task genuinely grew is the guard's documented remedy; narrowing
  the guard or deleting a row is not. A raised row is set to the file's new size **plus 25%**, per
  the formula in the guard's own header — never to the size exactly, which would leave zero headroom
  and fail the next byte added.

**Interfaces:**
- Consumes: nothing.
- Produces: one statement of each message, identical at every site.

All four change in one commit because a message stated two ways is a message that will be
transcribed the wrong way — which is the defect class this whole change addresses.

- [x] **Step 1: `pipeline.md`'s Git-boundaries chain**

The chain's two `commit -m` literals become `"<type>(<module>): <what the implementation does>"` and
`"chore(openspec): plan and session records"`. Add one sentence saying `<module>` is derived from
the reshaped diff and that the planning message is a **fixed literal**, fixed rather than derived
because every planning commit stages the same two trees in every change.

```bash unverified:confirm the chain's literals have not shifted since this plan was written
grep -n 'chore(<name>): plan and session records' skills/myflow-contracts/pipeline.md skills/myflow-finish/SKILL.md skills/myflow-do/SKILL.md
```

- [x] **Step 2: `finish-contract.md`'s two-commit section**

The prose describing the two commits states the same two messages. Keep the citation to
`pipeline.md`'s Git boundaries as the canonical statement of the *sequence*; this task changes only
what the messages say.

- [x] **Step 3: The two `commit-split.sh` call sites**

`skills/myflow-finish/SKILL.md`'s run-1 call and `skills/myflow-do/SKILL.md`'s PR-exception call
each pass the two messages as arguments. Both get the new literals. In `myflow-do`'s case the
`<impl-msg>` derivation sentence — currently `fix(<name>): <what changed since the last task
commit>` — becomes a module scope too.

`scripts/commit-split.sh` itself needs **no** edit: it takes both messages as arguments, and its
`<name>` parameter is already documented as accepted-and-unused.

- [x] **Step 4: Check the contract budget**

Editing two files under `skills/myflow-contracts/` can trip the ratchet.

```bash verified:this guard is in .myflow/project.md's ## lint list and runs against a bare tree
scripts/check-contract-budget.sh
```

Raise the budget for a file this task genuinely grew; never narrow the guard or delete a row.

**Tests:** none — prose-only change to contract text, no executable surface. Covered by the
contract-budget, references and installed-citations guards in the repository's lint list, named
without backticks for the reason task 3's field states.

**Regression:** Reverting this task puts the change name back into the only two commits that reach
`main`, which is where the defect is most visible.

**Baseline:** unchanged — this task adds no test.

**Commit:** `docs(myflow-contracts): scope finish run 1's two commits by module, not by change name`

---

### 6 The finish contract names `commit-split.sh`

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/finish-contract.md`
- Modify: `scripts/check-contract-budget.sh` — same ratchet row as task 5, for the same reason and
  under the same plus-25% rule.

**Interfaces:**
- Consumes: task 5's edit to the same section.
- Produces: KAN-202's proposed fix.

This is KAN-202. Its stated symptom is stale — `/myflow-fast` reaches the script through
`skills/myflow-finish/SKILL.md` §1.2 and did so on the day the ticket was filed — so this task is
defensive documentation, not a bug fix, and the plan says so rather than implying otherwise.

- [x] **Step 1: Name the script at the point the two commits are specified**

The section describes the clearing pass, the pathspec exclusion and the second bare `add`, and cites
`pipeline.md`'s Git boundaries for the sequence, without naming the script that implements it. Add
that name, so a reader treating the contract as the authority has something to call.

Keep `pipeline.md`'s Git boundaries as the canonical spec of the *behaviour*; the contract points at
the implementation of it.

```bash verified:this is the script's own stated reason for existing, in its header
sed -n '1,12p' scripts/commit-split.sh
```

**Tests:** none — contract text, covered by the lint list as in task 5.

**Regression:** Reverting this task returns the canonical statement of the two-commit step to
describing a chain and naming no implementation of it.

**Baseline:** unchanged — this task adds no test.

**Commit:** `docs(finish-contract): name commit-split.sh at the two-commit step`

---

### 7 Install-harness cases for the new rule

**Build:** red

**Squash-with:** Task 8

**Files:**
- Modify: `scripts/test-setup.sh`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: the executable statement that a sandboxed install renders and installs the new rule.
  Task 8 satisfies it.

- [x] **Step 1: Assert the managed block carries the rule's core**

Following the harness's existing sandboxed-`HOME` idiom, assert that after `setup.sh global` the
managed block in `$HOME/.claude/CLAUDE.md` contains the new rule's core excerpt and the
`Full rule: ~/.claude/rules/commit-scope-is-the-module.md` pointer that core extraction appends.

- [x] **Step 2: Assert the full text is installed and reachable**

Assert `$HOME/.claude/rules/commit-scope-is-the-module.md` exists and resolves — it is a symlink
into the repository, like every other installed rule.

```bash unverified:confirm the harness's helper names for asserting a managed-block substring and an installed rule path before writing new cases
grep -n 'rules/\|managed block\|assert_' scripts/test-setup.sh | head -20
```

- [x] **Step 3: Confirm the harness fails against the absent rule**

The red half of this pair: run it before task 8 and confirm the new cases fail rather than passing
vacuously.

**Tests:** the two cases enumerated in steps 1 and 2, in `scripts/test-setup.sh`.

**Regression:** Reverting this task leaves the rule's installation unasserted, so a rule that stops
rendering into the managed block fails silently.

**Baseline:** before=468 after=472 assertions in `scripts/test-setup.sh`.
<!-- measured: bash scripts/test-setup.sh @ branch main, before this change — the harness prints its own assertion count -->
<!-- predicted: the same command after task 7 — two cases at two assertions each -->

**Commit:** `test(setup): assert the commit-scope rule renders into the managed block and installs`

---

### 8 The always-on rule and its baseline row

**Build:** green

**Files:**
- Add: `rules/commit-scope-is-the-module.mdc`
- Modify: `rules/agent-baseline.md`
- Modify: `scripts/check-references.sh` — registering the new rule file in
  `EXPECTED_ZERO_RULE_FILES`. This is not optional and not collateral: that guard's per-file
  coverage check fails a rule file it does not know about
  (`rules/commit-scope-is-the-module.mdc:0: 0 checked, and not declared expected-zero`), and every
  one of the ten existing always-on rule files is registered there for the identical reason.
- Modify: `scripts/test-setup.sh` — task 7's file. Task 7 is `Build: red` with
  `**Squash-with:** Task 8`, so its commit folds into this one and this commit is the squash unit's.

**Interfaces:**
- Consumes: task 7's cases.
- Produces: the convention as an always-on rule, reaching every session and every dispatched agent.

- [x] **Step 1: Write the rule**

Frontmatter declaring `alwaysApply: true`, and `<!-- core -->` / `<!-- /core -->` markers around the
part that belongs in every session's prompt: what the scope names, what it may never be, and that
it is optional. The full text below the closing marker carries the reasoning and the guard.

Follow the shape of an existing always-on rule rather than inventing a second one.

```bash verified:these are the always-on rules this repository already ships; the frontmatter is what setup.sh reads
head -6 rules/build-the-simplest-thing.mdc
```

`setup.sh` needs **no** edit: `always_on_rules()` discovers a rule from its frontmatter, and
`render_managed_block` extracts the core and appends the pointer. Confirm both markers are present
and balanced — an unbalanced pair is a hard `die` in the installer.

- [x] **Step 2: Register the rule file with the references guard**

Add `rules/commit-scope-is-the-module.mdc` to `EXPECTED_ZERO_RULE_FILES` in
`scripts/check-references.sh`, in the array's existing alphabetical order. The array's
`_REASON` string already covers why a rule file cites zero paths in a bold-adjacent shape; this
adds a member, never a new reason and never a weakened check.

- [x] **Step 3: Add the `agent-baseline.md` row**

One row in the rule table: the one-liner, and a pointer to `~/.claude/rules/commit-scope-is-the-module.md`.
Match the existing rows' phrasing — a one-line statement of the rule, never a copy of it.

- [x] **Step 4: Run the install harness and the lint list**

```bash verified:this is the sandboxed installer run .myflow/project.md's ## run section documents, and the harness that wraps it
bash scripts/test-setup.sh
```

```bash verified:these two guards validate the citations and symlinks a new rule file introduces; both are in .myflow/project.md's ## lint list
scripts/check-references.sh && scripts/check-guard-symlinks.sh
```

**Tests:** task 7's two cases now pass.

**Regression:** Reverting this task removes the convention from every session prompt and from
`agent-baseline.md`, leaving it stated only inside the myflow skills — so a commit made outside a
`/myflow-*` run has no rule at all.

**Baseline:** after=472 assertions in `scripts/test-setup.sh`, all passing.
<!-- predicted: bash scripts/test-setup.sh after task 8 -->

**Commit:** `feat(rules): carry the module-scope commit convention as an always-on rule`

---

### 9 A harness case for a top-level skill symlink

**Build:** red

**Squash-with:** Task 10

**Files:**
- Modify: `scripts/test-check-guard-symlinks.sh`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: the executable statement that a symlink directly under `skills/<skill>/` is a violation.
  Task 10 satisfies it.

Written before rule 5 so the case asserts what
`specs/myflow-contract-distribution/spec.md`'s added **Requirement: A skill directory SHALL carry
symlinks only under its `scripts/` directory** says, rather than what the implementation happens to
do.

Use the harness's existing sandboxed-fixture idiom under `TMPDIR` with
`CHECK_GUARD_SYMLINKS_ROOT`, exactly as its other rule cases do. **Never** point the guard at this
repository to produce a violation.

- [x] **Step 1: The violation case**

A fixture skill whose directory holds a symlink at its top level — `skills/<skill>/<name>.md`
pointing at a file in a sibling skill — and no other defect. Assert exit 1, and assert the reported
line names **both** the symlink's path and its target, since a report naming only the path leaves
the reader unable to tell an anti-fix symlink from a stray one.

- [x] **Step 2: The two acceptance cases**

A fixture whose only symlinks sit under `skills/<skill>/scripts/` — exit 0, so rule 5 does not
re-report what rule 1 already governs. And a fixture whose skill directory holds only regular files
and directories — exit 0.

- [x] **Step 3: Confirm the harness fails against the unmodified guard**

The red half of this pair: run it before task 10 and confirm the violation case fails rather than
passing vacuously.

```bash verified:run against this tree before task 10 exists; the harness passes at its current count, so the new case is the only thing that can move it
bash scripts/test-check-guard-symlinks.sh 2>&1 | tail -2
```

**Tests:** the three cases enumerated in steps 1 and 2, in `scripts/test-check-guard-symlinks.sh`.

**Regression:** Reverting this task leaves rule 5 with no executable statement of what it refuses, so
the placement it exists to catch becomes unasserted.

**Baseline:** before=72 after=78 assertions in `scripts/test-check-guard-symlinks.sh`.
<!-- measured: bash scripts/test-check-guard-symlinks.sh | grep -c '^ok:' @ branch openspec/kan-202-commit-split-and-module-scopes, before this task -->
<!-- predicted: the same command after task 9 — three cases at two assertions each -->

**Commit:** `test(check-guard-symlinks): pin that a symlink at a skill's top level is a violation`

---

### 10 Rule 5 — no symlink directly under a skill directory

**Build:** green

**Files:**
- Modify: `scripts/check-guard-symlinks.sh`
- Modify: `scripts/test-check-guard-symlinks.sh` — task 9's file. Task 9 is `Build: red` with
  `**Squash-with:** Task 10`, so its commit folds into this one and this commit is the squash unit's.

**Interfaces:**
- Consumes: task 9's cases.
- Produces: a fifth rule in the guard, reported by name like the existing four.

Rule 1 validates entries **under** `skills/*/scripts/`. Nothing validates the skill directory's own
top level, which is where the anti-fix lands.

- [x] **Step 1: Add the rule**

For each skill directory in the scan set, report every entry directly under it that is a symlink.
`scripts/` itself is a directory, not a symlink, so it is not a hit; a symlink **inside** `scripts/`
is rule 1's, not this rule's.

Report `path:0: message` in the guard's existing violation shape — naming the symlink's path and its
resolved target — and fold it into the existing `GUARD-SYMLINKS-INVALID` count. Add **no** new exit
status and **no** second report shape: the header documents 0 clean, 1 any violation, 2 cannot
answer, and this rule is an ordinary violation class.

- [x] **Step 2: Extend the header's rule list**

The header enumerates four rules by number and describes each. Add the fifth in the same voice, and
say what it is for: a file a skill's text names but does not carry is resolved by reading where the
text says it lives — `skills/myflow-do/SKILL.md`'s `[PRINCIPLES_PATH]` paragraph is the case that
motivated this — never by symlinking a copy into the running command's own skill directory.

- [x] **Step 3: Verify against the real tree and the harness**

```bash verified:this tree currently holds no top-level skill symlink, so rule 5 must leave the verdict at OK
find skills -mindepth 2 -maxdepth 2 -type l | wc -l
```

```bash verified:this guard is in .myflow/project.md's ## lint list and runs against a bare tree
scripts/check-guard-symlinks.sh && bash scripts/test-check-guard-symlinks.sh
```

Both must exit clean: the repository has no such symlink, so a rule that fires here is over-broad.

**Tests:** task 9's three cases now pass.

**Regression:** Reverting this task returns the repository to prose as its only countermeasure
against an anti-fix that prose had already failed to prevent once.

**Baseline:** after=78 assertions in `scripts/test-check-guard-symlinks.sh`, all passing, and
`scripts/check-guard-symlinks.sh` still reports `GUARD-SYMLINKS-OK` against this repository.
<!-- predicted: bash scripts/test-check-guard-symlinks.sh after task 10 -->

**Commit:** `feat(check-guard-symlinks): reject a symlink placed directly under a skill directory`

---

### 11 Cases pinning commit resolution under the new subjects

**Build:** red

**Squash-with:** Task 12

**Files:**
- Modify: `scripts/test-gather-self-review-context.sh`

**Interfaces:**
- Consumes: tasks 5 and 6, which changed the two subjects finish run 1 writes.
- Produces: the executable statement that `gather-self-review-context.sh` still resolves both
  commits under those new subjects. Task 12 satisfies it.

**Why this task exists.** The review panel found that
`scripts/gather-self-review-context.sh` resolves three commits by grepping for the **change name as
the commit scope** — `ARCHIVE_SHA`, `PLAN_SHA` and `IMPL_SHA`. Tasks 5 and 6 removed the change name
from two of those three subjects, so `PLAN_SHA` and `IMPL_SHA` silently resolve to empty for every
change finished after this lands. `ARCHIVE_SHA` is unaffected: run 2's archive commit still uses
`chore(<name>)`, which this change does not touch.

The existing harness passes only because **its fixtures still use the old subjects**, which is
exactly why the defect reached the panel rather than a guard.

- [x] **Step 1: Cases for the new subjects**

Following the harness's existing fixture idiom, add a section whose fixture repository commits the
subjects finish run 1 writes **now**:

| Fixture commit | Must resolve as |
|----------------|-----------------|
| `feat(some-module): <impl body marker>` | `IMPL_SHA` |
| `chore(openspec): plan and session records` | `PLAN_SHA` |
| `chore(<name>): sync delta specs and archive the change` | `ARCHIVE_SHA`, unchanged |

Assert each resolves to the right commit, by a body marker unique to it — never by position.

- [x] **Step 2: A case that pins the change-specificity**

`chore(openspec): plan and session records` is now **identical across every change**, so a fixture
carrying two changes' plan commits must resolve to *this* change's, not merely the most recent. Give
the fixture a second, later plan commit belonging to a different change and assert the resolution
still picks the one whose `openspec/changes/<name>/` path this run was asked about.

This is the case that distinguishes a path-scoped resolution from a widened grep, and it is the
reason the plan takes the former.

- [x] **Step 3: The old subjects keep resolving**

Every pre-existing case in this harness must still pass unchanged. A change finished **before** this
one lands carries `chore(<name>): plan and session records` and a change-name-scoped implementation
subject, and its self-review must still gather them. Do not delete or rewrite an existing case to
make room.

- [x] **Step 4: Confirm the harness fails against the unmodified script**

The red half of this pair.

```bash verified:run against this worktree before task 12 exists; the harness passes at its current count because every fixture still uses the old subjects
bash scripts/test-gather-self-review-context.sh 2>&1 | tail -1
```

**Tests:** the cases enumerated in steps 1 to 3, in `scripts/test-gather-self-review-context.sh`.

**Regression:** Reverting this task leaves the new subjects' resolution unasserted, so the harness
goes back to passing while the script it guards returns nothing for two of its three commits.

**Baseline:** before=104 after=112 assertions in `scripts/test-gather-self-review-context.sh`.
<!-- measured: bash scripts/test-gather-self-review-context.sh | grep -c '^ok:' @ branch openspec/kan-202-commit-split-and-module-scopes, before this task -->
<!-- predicted: the same command after task 11 — four cases at two assertions each -->

**Commit:** `test(gather-self-review-context): pin commit resolution under the new finish subjects`

---

### 12 Resolve the plan and implementation commits by path

**Build:** green

**Files:**
- Modify: `scripts/gather-self-review-context.sh`
- Modify: `scripts/test-gather-self-review-context.sh` — task 11's file. Task 11 is `Build: red`
  with `**Squash-with:** Task 12`, so its commit folds into this one and this commit is the squash
  unit's.

**Interfaces:**
- Consumes: task 11's cases.
- Produces: a resolution that does not depend on the change name appearing in a commit's scope.

- [x] **Step 1: Resolve `PLAN_SHA` by path, not by subject**

The plan commit is the one that touched this change's own `openspec/changes/<name>/` directory. That
is exact and change-specific whatever the scope says, and it is why the panel's alternative — merely
widening the grep — was rejected: `chore(openspec): plan and session records` is now identical
across every change, so a subject-only match with `--max-count=1` can return a different change's
commit entirely.

Search both the live and archived locations, since run 2 moves the directory:
`openspec/changes/<name>/` and `openspec/changes/archive/*<name>/`.

- [x] **Step 2: Derive `IMPL_SHA` from the plan commit**

`commit-split.sh` makes the implementation commit and the planning commit **back to back**, in that
order, so the implementation commit is the planning commit's first parent.

**Handle the skipped-empty case.** Either commit is skipped rather than failed when nothing is
staged, per **Git boundaries** (`<agents repo>/skills/myflow-contracts/pipeline.md`). When the
implementation commit was skipped, the plan commit's parent is not one — resolve nothing rather than
returning the wrong commit. Returning a confident wrong answer is the failure mode the script's own
`IMPL_SHA` comment already warns about; preserve that judgment.

- [x] **Step 3: Keep the old subjects resolving**

A change finished before this lands has no `openspec/changes/<name>/` commit under the new shape but
does match the old greps. Keep the existing subject-based lookup as a fallback when the path-based
one finds nothing, so both eras resolve. Do not delete the old patterns.

- [x] **Step 4: Update the comments that state the old contract**

The block above these lookups explains the grep-and-exclude approach and the sync requirement
between `PLAN_SHA`'s grep and `IMPL_SHA`'s exclusion. Rewrite it to describe what the code now does.
A comment describing a mechanism the file no longer uses is worse than none.

```bash verified:this is the harness for this script, listed under .myflow/project.md's ## test
bash scripts/test-gather-self-review-context.sh
```

**Tests:** task 11's cases now pass, and every pre-existing case still passes.

**Regression:** Reverting this task returns `PLAN_SHA` and `IMPL_SHA` to greps whose scope no longer
matches, so every self-review after this change silently gathers two fewer commits.

**Baseline:** after=112 assertions in `scripts/test-gather-self-review-context.sh`, all passing.
<!-- predicted: bash scripts/test-gather-self-review-context.sh after task 12 -->

**Commit:** `fix(gather-self-review-context): resolve the plan and implementation commits by path`
