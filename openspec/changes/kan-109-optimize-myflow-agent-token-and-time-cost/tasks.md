> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** Cut the structural half of myflow's per-change agent cost — serialize implementers within
a worktree, dispatch by file-overlap bundle rather than by plan task number, measure the diff before
the panel reads it, and stop re-running conditional panel slots against subjects that did not change.

**Architecture:** Two new guard scripts plus edits to `skills/myflow-do/SKILL.md` sections 4 and 5.
`plan-dispatch-bundles.sh` is a thin Bash wrapper over `plan-dispatch-bundles.py`, the same split
`check-task-build-green.sh` and `check-task-commit-fields.sh` already use for `tasks.md` parsing.
`check-panel-diff-size.sh` is pure Bash over git, in the shape of `check-finish-preflight.sh`, since
it parses no Markdown. Two rules — serialization and conditional-slot re-run scoping — govern agent
behavior and have nothing on disk to measure, so they are carried by the skill and the delta specs
alone.

**Tech Stack:** Bash, Python 3 standard library only (`/usr/bin/python3`, no third-party imports, no
pip, no network), Markdown skills and contracts, the `openspec` CLI. No runnable application in this
repository; verification is the guard scripts and assertion harnesses declared under `## lint` and
`## test` in `.myflow/project.md`, plus reading the diff.

## Global Constraints

- **No new suppression markers, no guard weakening.** A lint hit is fixed by editing the offending
  line. This includes `check-contract-budget.sh`: its `skills/myflow-do/SKILL.md` row is re-anchored
  in Task 6 from the file's real post-change size, never narrowed in scope or deleted.
- **Python is standard library only.** `plan-dispatch-bundles.py` imports nothing outside the
  standard library, matching the constraint `.myflow/project.md` already states for
  `check-plan-provenance.py` and `check-task-build-green.py`.
- **Neither new script joins `## lint`.** Both need a change in flight — a `tasks.md` for a live
  change, or a worktree and a merge base — which is exactly why `check-finish-preflight.sh`,
  `check-unfinished-work.sh`, `preserve-session-records.sh` and `check-cleanup-complete.sh` are
  already excluded. Both join `## test` in Task 5.
- **Task 6 is the last content task.** No earlier task edits `scripts/check-contract-budget.sh`.
- **No change to the review panel roster presets, the optional-slot trigger table's own rows, the
  escalation ladder's trigger conditions, the panel record format, the marker-line rules, or the
  zero-open-findings bar.** Every rule here changes how much is read and how dispatches are paced,
  never what clears the gate.
- **`/myflow-fast` is not edited.** It inherits all four rules through the `/myflow-do` sections it
  already cites. A task that finds itself editing `skills/myflow-fast/SKILL.md` has drifted.

## Baseline

<!-- verified: wc -c on each path, working tree at the start of this change (commit f31cad1) -->

| File | Bytes now | Budget in `budgets()` |
|------|-----------|------------------------|
| `skills/myflow-do/SKILL.md` | 43690 | 54011 |
| `.myflow/project.md` | 7218 | not covered by the guard |

`skills/myflow-do/SKILL.md` has 10321 bytes of headroom under its current row. This plan adds prose
to sections 4 and 5, so Task 6 re-anchors the row from whatever the file measures after Task 5 —
raising it if the additions consumed the headroom, and lowering it to the post-change size plus 25%
if they did not, which is what the guard's own convention asks of a re-anchor.

`scripts/plan-dispatch-bundles.py`, `scripts/plan-dispatch-bundles.sh` and
`scripts/check-panel-diff-size.sh` are new and take no budget row: the guard covers
`skills/myflow-contracts/*.md`, `skills/*/SKILL.md` and `skills/*/SKILL-rationale.md`, and nothing
under `scripts/`.

---

### 1 `scripts/check-panel-diff-size.sh` — measure the diff before the panel reads it

**Build:** green

**Files:**
- Create: `scripts/check-panel-diff-size.sh`
- Create: `scripts/test-check-panel-diff-size.sh`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: `check-panel-diff-size.sh <worktree> <merge-base> [cap]`, called by `myflow-do` §5 in
  Task 4.

- [x] **Step 1: Write `scripts/check-panel-diff-size.sh`**

Pure Bash over git, argument-validated, with the exit-code contract in its own header — the shape
`check-finish-preflight.sh` already uses. It sums insertions and deletions across committed work and
the unstaged working tree, which together are the text `final-review.diff` will carry.

```bash unverified:authored in-tree for this change; the --shortstat output shape is git's documented format but has not been run against this script yet
#!/usr/bin/env bash
# check-panel-diff-size.sh — report whether a branch's diff is within the
# review panel's reading cap.
#
# Exit codes: 0 at or under the cap; 1 over the cap; 2 cannot answer at all.
set -euo pipefail

worktree="${1:-}"
merge_base="${2:-}"
cap="${3:-2000}"
```

The default cap is **2000** changed lines.

<!-- verified: KAN-109's own measurement section — 948k tokens for a 7-slot panel pass over a 9877-line diff -->

That figure is derived from the ticket's measurement: seven slots cost 948k tokens reading a
9877-line diff, so a cap at roughly a fifth of that keeps a full pass in the low hundreds of
thousands of tokens while leaving ordinary myflow changes untouched.

Validation, in this order, each failing with exit 2 and a message naming what was wrong: the
worktree argument is missing or is not a directory containing a git repository; the merge base is
missing or does not resolve via `git -C "$worktree" rev-parse --verify`; the cap is not a
non-negative decimal integer. Reporting the count and the verdict on stdout, one line each, so a
caller can quote the script rather than re-derive the number.

- [x] **Step 2: Write `scripts/test-check-panel-diff-size.sh`**

Same harness shape as `scripts/test-commit-split.sh`: a sandboxed git repository per case under
`TMPDIR`, an indexed `REPOS` array cleaned up on `EXIT` including on a failed assertion, and a
`fail`/`pass` counter pair. Six cases, listed under **Tests** below.

- [x] **Step 3: Verify**

```bash unverified:new script and harness, run for the first time in this task
chmod +x scripts/check-panel-diff-size.sh scripts/test-check-panel-diff-size.sh
scripts/test-check-panel-diff-size.sh
```

```bash verified:declared under ## lint in .myflow/project.md
scripts/check-references.sh
scripts/check-vocabulary.sh
```

Expected: the harness's six cases pass; both lint guards exit 0.

**Tests:** Case 1: a diff under the cap exits 0 and prints the count. Case 2: a diff over the cap
exits 1 and prints the count. Case 3: an explicit third argument overrides the default cap. Case 4:
unstaged working-tree changes count toward the total alongside committed ones. Case 5: a
non-numeric cap exits 2. Case 6: a path that is not a git worktree exits 2.
**Regression:** Case 4 (unstaged counted): dropping the unstaged term would let a run whose work sits
uncommitted in the working tree measure as zero and pass a cap it actually exceeds, which is the one
reading the panel is about to do. Case 5 and Case 6 (exit 2): collapsing either into exit 0 would
make an unmeasurable diff read as under the cap, the failure the spec's "unmeasurable is not small"
scenario names.
**Baseline:** before=0 after=6 cases in `scripts/test-check-panel-diff-size.sh` (new harness).
**Commit:** `feat(kan-109-optimize-myflow-agent-token-and-time-cost): add check-panel-diff-size.sh`

---

### 2 `scripts/plan-dispatch-bundles.{sh,py}` — group tasks by declared file overlap

**Build:** green

**Files:**
- Create: `scripts/plan-dispatch-bundles.py`
- Create: `scripts/plan-dispatch-bundles.sh`
- Create: `scripts/test-plan-dispatch-bundles.sh`

**Interfaces:**
- Consumes: nothing from earlier tasks. Reads the `**Files:**` and `**Allowed-collateral:**` fields
  `myflow-task-commit-fields` already requires of every task, and the `[ ]` / `[x]` checkbox markers
  `check-task-build-green.py` already parses.
- Produces: `plan-dispatch-bundles.sh [path-to-tasks.md]`, called by `myflow-do` §4 in Task 3.

- [x] **Step 1: Write `scripts/plan-dispatch-bundles.py`**

Standard library only. One `tasks.md` path per invocation, matching `check-task-build-green.py`'s
scope. It parses each task heading for its dotted id, skips tasks whose steps are all `[x]`, reads
the task's `**Files:**` block, and unions tasks sharing any declared path with union-find.
`**Allowed-collateral:**` is read and deliberately **not** unioned on — the spec's own scenario.

Path comparison is on the literal declared path text after stripping the `Create:` / `Modify:` /
`Delete:` prefix and surrounding backticks — no filesystem resolution, no globbing, no
normalization beyond stripping whitespace. Two tasks agree on a path when they wrote the same path.

Output on stdout, one line per bundle, bundles ordered by lowest task id and ids within a bundle in
plan order:

```text unverified:output format authored in-tree for this change
bundle 1: 1.1 1.2
bundle 2: 2.1
```

Exit 0 bundles computed; 1 an unchecked task carries no `**Files:**` field, with that task's id on
stderr; 2 cannot answer — unreadable path, unparsable file.

- [x] **Step 2: Write `scripts/plan-dispatch-bundles.sh`**

A thin wrapper in the shape `check-task-build-green.sh` already has: resolve `SCRIPT_DIR`, probe
that `python3` is both on `PATH` and able to run a trivial program (the macOS-stub case that
wrapper's own comment documents), `exec` the Python script on a single argument, and with no
argument resolve every non-archived `openspec/changes/*/tasks.md` under a root derived from the
script's own location, aggregating exit codes.

- [x] **Step 3: Write `scripts/test-plan-dispatch-bundles.sh`**

Fixture `tasks.md` files under `TMPDIR`, one per case, cleaned up on `EXIT`. Seven cases, listed
under **Tests** below.

- [x] **Step 4: Verify**

```bash unverified:new scripts and harness, run for the first time in this task
chmod +x scripts/plan-dispatch-bundles.sh scripts/test-plan-dispatch-bundles.sh
scripts/test-plan-dispatch-bundles.sh
```

```bash verified:declared under ## lint in .myflow/project.md
scripts/check-references.sh
scripts/check-vocabulary.sh
```

Expected: the harness's seven cases pass; both lint guards exit 0.

**Tests:** Case 1: three tasks declaring disjoint paths produce three bundles. Case 2: two tasks
sharing one path produce one bundle. Case 3: transitivity — A shares with B, B shares with C, all
three land in one bundle. Case 4: a task whose steps are all `[x]` takes no part in any bundle.
Case 5: an unchecked task with no `**Files:**` field exits 1 and names that task. Case 6: an
`**Allowed-collateral:**` glob matching another task's declared path does not join the two bundles.
Case 7: an unreadable path exits 2.
**Regression:** Case 3 (transitivity): a pairwise-only grouping would split a three-task file chain
into two bundles and dispatch two implementers into the same files, which is the seam breakage this
change exists to remove. Case 6 (collateral not joined): unioning on collateral globs would collapse
most plans into one bundle and serialize work that never needed to be.
**Baseline:** before=0 after=7 cases in `scripts/test-plan-dispatch-bundles.sh` (new harness).
**Commit:** `feat(kan-109-optimize-myflow-agent-token-and-time-cost): add plan-dispatch-bundles`

---

### 3 `myflow-do` §4 — serialize implementers and dispatch by bundle

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md` (section 4, "Execute (SDD + TDD)")

**Interfaces:**
- Consumes: `plan-dispatch-bundles.sh` from Task 2.
- Produces: the dispatch rules the review-panel edits in Task 4 sit alongside; no shared surface.

- [x] **Step 1: State the serialization rule**

In section 4, before the dispatch clauses, add the rule: at most one implementer subagent in flight
against a given worktree at any moment; the parent waits for the previous implementer's commit sha
before dispatching the next into that worktree; dispatches into different worktrees remain free to
run concurrently. Name the override explicitly — `superpowers:subagent-driven-development`'s parallel
dispatch guidance and `superpowers:dispatching-parallel-agents`, for same-worktree tasks only —
alongside the model-policy overrides the section already carries against the same upstream skill.

State the measured reason in one sentence and no more: two implementers on one build directory left
assertions red at file seams, idled agents on another's mid-edit compile, and corrupted test-result
XML. The full account stays in `design.md` and in the ticket.

- [x] **Step 2: Replace "each remaining checkbox (or a tightly coupled group)" with the bundle rule**

Section 4 currently opens "treating each remaining checkbox (or a tightly coupled group) as one
task". Replace the parenthetical judgment call with the computed grouping: run

```bash unverified:call site authored in-tree for this change; the script itself lands in Task 2
scripts/plan-dispatch-bundles.sh <changeRoot>/tasks.md
```

and dispatch one implementer per printed bundle, in the order printed. A non-zero exit is a plan
defect, not a review finding: exit 1 names a task missing its `**Files:**` field, which
`superpowers:writing-plans` repairs before any dispatch happens; exit 2 stops the run.

State that bundling does not change the commit-per-task model — an implementer handed a bundle still
makes one commit per task with that task's own `Task-Id:` trailer, and a `Build: red` task still
folds into the commit its `**Squash-with:**` field names.

- [x] **Step 3: Verify**

```bash verified:declared under ## lint in .myflow/project.md
scripts/check-references.sh
scripts/check-vocabulary.sh
scripts/check-contract-budget.sh
```

Expected: `check-references.sh` and `check-vocabulary.sh` exit 0. `check-contract-budget.sh` may
exit 1 on `skills/myflow-do/SKILL.md` if this task's additions consumed the headroom recorded under
**Baseline**; that is expected and is re-anchored in Task 6, not worked around here.

**Tests:** No automated test — this task edits skill prose, which no harness in this repository
executes. Verification is the three lint guards above plus reading the diff against the
myflow-dispatch-economy delta spec's scenarios.
**Regression:** Reverting this task leaves `plan-dispatch-bundles.sh` written but never called and
the parallel-dispatch guidance unoverridden, so `/myflow-do` keeps dispatching one implementer per
checkbox, concurrently — every failure mode the capability's scenarios name.
**Baseline:** before=0 after=0 automated cases; the lint guards' pass count is unchanged at 6
commands declared under `## lint`.
**Commit:** `feat(kan-109-optimize-myflow-agent-token-and-time-cost): serialize and bundle implementer dispatches`

---

### 4 `myflow-do` §5 — measure the diff, and scope conditional re-runs

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md` (section 5, "The review panel", and its "Panel re-runs"
  subsection)

**Interfaces:**
- Consumes: `check-panel-diff-size.sh` from Task 1.
- Produces: nothing later tasks read except the file's final size, which Task 6 measures.

- [x] **Step 1: Add the diff-size measurement before `final-review.diff` is written**

Section 5 currently opens by writing `.superpowers/sdd/final-review.diff`. Before that, add:

```bash unverified:call site authored in-tree for this change; the script itself lands in Task 1
scripts/check-panel-diff-size.sh <worktree> <merge-base>
```

Exit 0 proceeds. Exit 1 puts the choice to the operator as named options, shaped per **Operator
prompts** (`skills/myflow-contracts/operator-prompts.md`) and citing that contract rather than
restating its mechanics — proceed with the panel anyway *(default, recommended)*, or stop and split
the change, which ends the run at `IN_PROGRESS` with the implementation committed on the branch.
Exit 2 stops the run: a size the guard could not measure is not a size under the cap.

State that the measured count, the cap in force, and the operator's answer where one was given are
recorded in `.superpowers/sdd/final-review-panel.md` on every run, including exit-0 runs. State that
the cap moves nothing about the roster, the slots, the escalation ladder or the zero-open-findings
bar.

- [x] **Step 2: Scope Full-mode re-runs to slots whose subject changed**

In "Panel re-runs", the **Full** row currently reads "Every slot in this run's roster". Narrow it:
every **required** slot re-runs against the rewritten `final-review.diff`; a **conditional** slot —
Security, Adversarial, Lens B, Lens C — re-runs only when its own row in the optional-slot trigger
table still fires against `fix-round-N.diff`. A conditional slot whose trigger did not fire is not
re-run; its previous result stands and the record says `not re-run — subject unchanged`, distinctly
from a slot whose trigger never fired at all and from a slot that was declined.

Add the one sentence that keeps this a definition rather than a waiver: a result is stale when the
diff it read has since changed in the region that slot reads, a conditional slot's region is exactly
its trigger's subject, and a required slot has no bounded region — which is why the scoping reaches
conditional slots alone. Leave the **Targeted** row untouched.

- [x] **Step 3: Verify**

```bash verified:declared under ## lint in .myflow/project.md
scripts/check-references.sh
scripts/check-vocabulary.sh
scripts/check-contract-budget.sh
```

Expected: `check-references.sh` and `check-vocabulary.sh` exit 0. `check-contract-budget.sh` may
exit 1 on `skills/myflow-do/SKILL.md`, re-anchored in Task 6.

**Tests:** No automated test — skill prose, as in Task 3. Verification is the three lint guards plus
reading the diff against the myflow-review-panel-economics delta spec's two new requirements'
scenarios.
**Regression:** Reverting this task leaves `check-panel-diff-size.sh` written but never called and
Full-mode escalation re-running every conditional slot against unchanged subjects — the 76k-token
Security re-read the ticket measured.
**Baseline:** before=0 after=0 automated cases.
**Commit:** `feat(kan-109-optimize-myflow-agent-token-and-time-cost): cap the panel diff and scope conditional re-runs`

---

### 5 Declare the two new harnesses under `## test`

**Build:** green

**Files:**
- Modify: `.myflow/project.md` (`## test` list)

**Interfaces:**
- Consumes: the harnesses from Tasks 1 and 2.
- Produces: the declared list Task 7's sweep runs.

- [x] **Step 1: Add both harnesses**

Add `scripts/test-check-panel-diff-size.sh` and `scripts/test-plan-dispatch-bundles.sh` to the
`## test` list. The list is not alphabetically ordered today — it runs roughly in the order the
scripts were added — so append both at the end rather than reordering the existing entries.

- [x] **Step 2: State why neither joins `## lint`**

`.myflow/project.md` already carries a paragraph naming the four `/myflow-finish` helpers that are
deliberately not lint steps, with its reason: they need a change in flight and arguments passed in,
so a lint step that cannot run against a bare tree would fail on every unrelated invocation. Extend
that paragraph to name the two new scripts for the same reason — do not write a second paragraph
restating it.

- [x] **Step 3: Verify**

```bash verified:declared under ## lint in .myflow/project.md
scripts/check-references.sh
scripts/check-vocabulary.sh
```

Expected: both exit 0.

**Tests:** No automated test — this task edits a configuration file's declared command list.
Verification is the two lint guards plus Task 7's sweep, which runs every command the edited list
declares.
**Regression:** Reverting this task leaves both new harnesses undeclared, so nothing runs them and a
regression in either guard lands silently — the same gap the four `/myflow-finish` helpers were
given `## test` entries to close.
**Baseline:** before=15 after=17 commands declared under `## test` in `.myflow/project.md`.

<!-- verified: sed -n '/^## test/,/^## lint/p' .myflow/project.md | grep -c '^scripts/' at the start of this change -->
**Commit:** `chore(kan-109-optimize-myflow-agent-token-and-time-cost): declare the two new test harnesses`

---

### 6 Re-anchor `skills/myflow-do/SKILL.md`'s contract budget

**Build:** green

**Files:**
- Modify: `scripts/check-contract-budget.sh` (`budgets()` table, one row)

**Interfaces:**
- Consumes: the final size of `skills/myflow-do/SKILL.md` after Tasks 3 and 4.
- Produces: a green `check-contract-budget.sh` for Task 7's sweep.

- [x] **Step 1: Measure the file as it now stands**

```bash unverified:the number this prints is what Step 2 writes; it cannot be known before Tasks 3 and 4 land
wc -c skills/myflow-do/SKILL.md
```

- [x] **Step 2: Set the row to the measured size plus 25%**

Replace the `skills/myflow-do/SKILL.md` row's byte figure with the measured size multiplied by 1.25,
rounded down to a whole number of bytes — the convention the guard's own header states and every
existing row follows. Do not narrow the guard's scope, delete the row, or add a suppression marker.

This row moves in whichever direction the measurement says: if Tasks 3 and 4 consumed the 10321
bytes of headroom recorded under **Baseline**, the row rises; if they did not, the row falls to the
new size plus 25%, which is what a re-anchor means. A row left at its old figure because the file
still fits under it is a regrowth allowance this guard exists to deny.

- [x] **Step 3: Verify**

```bash verified:declared under ## lint in .myflow/project.md
scripts/check-contract-budget.sh
scripts/test-check-contract-budget.sh
```

Expected: `check-contract-budget.sh` exits 0; its harness still passes.

**Tests:** No new case — the budget guard's existing harness already asserts its over-budget,
under-budget and missing-row behavior against fixture trees, and a changed figure in the real table
exercises none of those paths differently.
**Regression:** Reverting this task returns the row to 54011 while the file has grown, so the guard
either passes on a file that outgrew its measured anchor or fails outright — the regrowth the
ratchet exists to catch.
**Baseline:** before=1 after=1 row for `skills/myflow-do/SKILL.md` in `budgets()`; the figure
changes, the row count does not.
**Commit:** `chore(kan-109-optimize-myflow-agent-token-and-time-cost): re-anchor the myflow-do budget row`

---

### 7 Full guard and test sweep

**Build:** green

**Files:**
- Modify: none — this task runs the declared commands and fixes whatever they report, in whichever
  file reports it.

**Allowed-collateral:** `scripts/*.sh`, `scripts/*.py`, `skills/**/*.md`, `.myflow/project.md`

**Interfaces:**
- Consumes: every earlier task.
- Produces: the green state `/myflow-do` §7 requires before it hands off.

- [x] **Step 1: Run every lint command**

```bash verified:the six commands declared under ## lint in .myflow/project.md
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-plan-provenance.sh
scripts/check-task-build-green.sh
scripts/check-workspace-isolation.sh
scripts/check-contract-budget.sh
```

Expected: every one exits 0.

- [x] **Step 2: Run every test harness**

```bash verified:the ## test list as it stands after Task 5 — the fifteen existing entries plus the two added there
scripts/test-setup.sh
scripts/test-check-references.sh
scripts/test-check-plan-provenance.sh
scripts/test-check-finish-preflight.sh
scripts/test-commit-split.sh
scripts/test-prepare-workspace.sh
scripts/test-preserve-session-records.sh
scripts/test-check-unfinished-work.sh
scripts/test-check-cleanup-complete.sh
scripts/test-gather-self-review-context.sh
scripts/test-check-task-build-green.sh
scripts/test-check-task-commit-fields.sh
scripts/test-check-workspace-isolation.sh
scripts/test-check-contract-budget.sh
scripts/test-check-vocabulary.sh
scripts/test-check-panel-diff-size.sh
scripts/test-plan-dispatch-bundles.sh
```

Expected: every harness exits 0.

- [x] **Step 3: Validate the change's own artifacts**

```bash verified:run in this worktree during task 7's sweep; `--change` is rejected by this CLI version, `--changes` is the accepted spelling
openspec validate --changes kan-109-optimize-myflow-agent-token-and-time-cost --strict
```

Expected: exit 0. `openspec validate --specs --strict` is known to fail on `dependency-versions`
for a missing `## Purpose` left by KAN-54's archive; that is pre-existing, tracked as KAN-135, and
is not touched by this change.

**Tests:** No new case — this task runs the existing suites. A failure it uncovers is fixed in the
file that caused it, and a fix that needs a new assertion adds it to that script's own harness.
**Regression:** Skipping this task lets a cross-task interaction — a call site naming a script
argument the script does not accept, a budget row measured before the last edit — reach the review
panel unmeasured, which is the whole reason `/myflow-do` §7 refuses to hand off on a non-zero exit.
**Baseline:** before=6 lint commands and 17 test commands after Task 5; after=the same, all exiting
0.
**Commit:** `chore(kan-109-optimize-myflow-agent-token-and-time-cost): full guard and test sweep`
