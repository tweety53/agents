# Tasks — the SDD ledger gets one canonical path

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

Five tasks. Tasks 1–2 are a test-first pair against `preserve-session-records.sh` and are dispatched
together. Tasks 3, 4 and 5 touch disjoint files and can run in any order among themselves, after the
pair.

The canonical path this whole change is about, written once here so no task has to re-derive it:

```text verified:scripts/sdd-workspace prints "$root/.superpowers/sdd/$(basename tasks.md .md)"; skills/myflow-do/SKILL.md:214 invokes superpowers:subagent-driven-development
<abs-worktree>/.superpowers/sdd/tasks/progress.md
```

---

### 1 Harness cases for rescue and MISSING

**Build:** red

**Squash-with:** Task 2

**Files:**
- Modify: `scripts/test-preserve-session-records.sh`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: the executable statement of what `preserve()`'s rescue and MISSING outcomes must do.
  Task 2 satisfies it.

Written before the script so the cases assert what the delta spec's **A record found at a
known-wrong path is rescued, not skipped** and **A record absent from every known path is reported
as missing, not skipped** require, rather than what the implementation happens to do. The harness's
own header (lines 5–12) states this rule and records the suite that once encoded a guard's defects
as its specification.

Follow the harness's existing idiom exactly: `new_tree` for a sandbox with all three sources
present, `run_it` to invoke the script for the change named `demo`, one `ok:`/`fail` line per
assertion. Add a helper beside `new_tree` — call it `new_tree_no_ledger` — that builds the same tree
and then removes `$WT/.superpowers/sdd/tasks/progress.md`, so a rescue case can place its own
wrong-path file. Do **not** invent a second fixture shape.

- [x] **Step 1: Rescue cases for the ledger, one per allowlist entry**

| Case | File placed | Expects |
|------|-------------|---------|
| flat `ledger.md` | `.superpowers/sdd/ledger.md` | `rescued:` naming that path; the file lands at `docs/superpowers/ledgers/<date>-demo.md` |
| flat `progress.md` | `.superpowers/sdd/progress.md` | same |
| `tasks/ledger.md` | `.superpowers/sdd/tasks/ledger.md` | same |

Each case asserts the *content* arrives at the destination, not merely that a `rescued:` line was
printed — a report of a copy that copied nothing would satisfy a message-only assertion, which is
the trap case 1b in this harness already guards against for the ordinary copy.

- [x] **Step 2: Allowlist ordering**

Place `.superpowers/sdd/ledger.md` and `.superpowers/sdd/progress.md` in the same tree. Assert the
destination carries `ledger.md`'s body, and that the `rescued:` line names `ledger.md`. This is the
case that stops the implementation being written as an unordered glob.

- [x] **Step 3: MISSING for the ledger**

Remove every ledger path. Assert the output carries a `MISSING:` line, that it names the canonical
path **and** all three fallbacks, that no `skipped:` line is printed for the ledger, and that
`RC` is 0.

- [x] **Step 4: The same four cases for the panel record**

Fallbacks `.superpowers/sdd/panel.md`, `.superpowers/sdd/review-panel.md`,
`.superpowers/sdd/final-review.md`; destination `docs/superpowers/reviews/<date>-demo-panel.md`. This
needs the panel's canonical path absent too, so add a `new_tree_no_panel` sibling one-liner beside
`new_tree_no_ledger` rather than inlining an `rm` at each call site.

Existing **case 3** must also be amended: once task 2 lands, a tree with the canonical ledger removed
and no fallbacks present prints `MISSING:`, not `skipped:`, so its `skipped:` assertion breaks. Fix it
to expect the outcome each source's allowlist implies — `MISSING:` for the ledger and panel record,
`skipped:` for the proposal artifact — leaving its RC and REMAINING assertions untouched.

- [x] **Step 5: The proposal artifact is untouched**

With no `$STATE_DIR/demo-proposal-artifact.html`, assert the output still carries the ordinary
`skipped:` line for it and **no** `MISSING:` line mentioning the artifact. This is the case that
pins the two-sources-not-three decision.

- [x] **Step 6: A rescued source outside the worktree is refused**

Point `.superpowers/sdd/ledger.md` at a file outside `$WT` via a symlink, with the canonical path
absent. Assert stderr carries the refusal, `RC` is non-zero, and the panel record and artifact were
still preserved. This is the case that proves the rescue did not bypass protection 3.

- [x] **Step 7: Idempotency holds for a rescue**

Pre-place a preserved ledger under a fixed past date, then rescue from a wrong path. Assert the
existing dated file is overwritten and no second dated file exists — the same assertion case 2
makes for an ordinary re-copy.

- [x] **Step 8: Confirm the harness fails against the unmodified script**

The red half of this pair: run it before task 2 and confirm every new case above fails loudly
rather than passing vacuously.

```bash unverified:run this before task 2 lands; the expected result is a non-zero exit naming the new cases as failures
bash scripts/test-preserve-session-records.sh 2>&1 | tail -5
```

**Tests:** the cases enumerated in steps 1–7, in `scripts/test-preserve-session-records.sh`.

**Regression:** Reverting this task leaves the rescue and MISSING outcomes with no executable
statement of what they must do, so a rescue that silently copies nothing, an unordered allowlist, a
MISSING that omits the paths tried, and a rescue that bypasses the source-containment protection all
become unasserted.

**Baseline:** before=135 after=169 `ok:` assertions in `scripts/test-preserve-session-records.sh`.
<!-- measured: bash scripts/test-preserve-session-records.sh 2>&1 | grep -c '^ok:' @ branch main, before this change -->
<!-- measured: the same command @ branch openspec/kan-77-sdd-ledger-canonical-path after tasks 1+2 landed as f0905f2 — the original predicted=157 was wrong -->

**Commit:** `test(preserve-session-records): assert rescue and MISSING outcomes`

---

### 2 Rescue and MISSING in preserve-session-records.sh

**Build:** green

**Files:**
- Modify: `scripts/preserve-session-records.sh`

**Interfaces:**
- Consumes: task 1's cases.
- Produces: the `rescued:` and `MISSING:` outcomes tasks 3's contract rows describe.

`preserve()` currently takes four arguments — `<source> <source-root> <dest-dir> <suffix>` — and its
three call sites are at the foot of the file. Add a **fifth**: a newline-separated, ordered list of
fallback source paths, empty for the proposal artifact.

The control flow, replacing only the leading `[ ! -f "$src" ]` early return:

```bash unverified:confirm the label argument reads naturally in the MISSING line before settling the wording
# 1. canonical present            -> unchanged path, "preserved:"
# 2. canonical absent, fallback   -> "rescued: <dest> (found at <fallback>)"
# 3. canonical absent, none, list non-empty -> "MISSING: <canonical> — tried <paths>", return 0
# 4. canonical absent, list empty -> "skipped: <src> (absent)", return 0
```

Requirements the implementation must hold, each pinned by a case in task 1:

- The fallback search walks the list **in order** and stops at the first `[ -f ]` hit.
- A rescued file is assigned to `$src` and then falls through to **every existing protection
  unchanged** — the `src_root` containment check, the `dest_dir` resolution and containment check,
  and acting through `$src_real`/`$dest_real` rather than the arguments. Do not add a second copy
  path that skips them.
- A refused fallback is `return 1` with its message on stderr, exactly as a refused canonical source
  is; it is never downgraded to `skipped:` or `MISSING:`.
- Both new outcomes `return 0`, so `RC` stays 0 and the caller's exit-status contract is unchanged.
- The `MISSING:` line names the canonical path first, then each fallback in list order.

The three call sites become:

| Source | Fallbacks, in order |
|--------|---------------------|
| ledger | `.superpowers/sdd/ledger.md`, `.superpowers/sdd/progress.md`, `.superpowers/sdd/tasks/ledger.md` |
| panel record | `.superpowers/sdd/panel.md`, `.superpowers/sdd/review-panel.md`, `.superpowers/sdd/final-review.md` |
| proposal artifact | *(empty)* |

The ledger's order is taken from KAN-77's own occurrence list, not invented: `.superpowers/sdd/ledger.md`
is recorded for six changes and `.superpowers/sdd/progress.md` for one.
<!-- measured: KAN-77's description, "Paths observed in practice" bullet @ the issue as fetched 2026-08-21 -->

Update the script's header comment block: the "Prints one line per source" sentence gains the two new
outcomes, and the sentence asserting that `skipped:` is the only silent outcome stays true and is
left alone. Add a line to the source-list comment citing the Temporary artifacts registry row task 3
writes, so the path agrees by reference rather than by coincidence.

- [x] **Step 1: The fifth parameter and the four-way branch**
- [x] **Step 2: The three call sites' allowlists**
- [x] **Step 3: Header comment block**
- [x] **Step 4: Run the harness green**

```bash verified:this is the harness's own invocation, already in .myflow/project.md's ## test list
bash scripts/test-preserve-session-records.sh
```

**Tests:** none added here — task 1 carries them.

**Regression:** Reverting this task restores the silent `skipped: (absent)` for a misfiled ledger,
and every case task 1 added fails.

**Baseline:** before=169 after=169 `ok:` assertions; task 1's count is unchanged and all of them pass.
<!-- measured: bash scripts/test-preserve-session-records.sh | grep -c '^ok:' @ branch openspec/kan-77-sdd-ledger-canonical-path after f0905f2 -->

**Commit:** `feat(preserve-session-records): rescue a misfiled record and report a missing one`

---

### 3 The registry row and the outcome table

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md`

**Interfaces:**
- Consumes: task 2's outcome wording.
- Produces: the contract rows `preserve-session-records.sh` and its callers cite.

Two edits, both in `skills/myflow-contracts/pipeline.md`:

- **Temporary artifacts registry** — the `SDD ledger` row's Location cell changes from
  `<abs-worktree>/.superpowers/sdd/` to `<abs-worktree>/.superpowers/sdd/tasks/progress.md`. The
  `Panel record` and `Per-task and review diffs` rows are left alone: they name a directory too, but
  only the ledger's path is in dispute and widening the edit would put four rows into a change about
  one.
- **Preserving the session records** — the outcome table gains two rows, after the existing
  `preserved:` row and before the non-zero row, so the table reads in increasing order of alarm:

| Outcome | What it means | What you do |
|---------|---------------|-------------|
| `rescued: <dest> (found at <path>)`, exit 0 | The record was written to a non-canonical path and has been copied to the canonical destination. | **Report it.** The record is safe; the writer that chose the path is not. Proceed. |
| `MISSING: <canonical> — tried <paths>`, exit 0 | A record that should exist for every change was found at none of its known paths. | **Report it, naming the paths tried.** Proceed with the integration. |

The label is the **canonical source path itself**, not a separate parameter — `preserve()` takes a
fifth argument, not a sixth. The line reads `MISSING: <canonical> — tried <fb1>, <fb2>, <fb3>`, and
the paths it prints are absolute, as every other path the script prints is; the relative spellings in
this plan's tables are readability only.

The paragraph below the table already says a non-zero exit is never silent and never a stop; extend
it to state that neither new outcome is non-zero, so a caller branching on exit status is not
tempted to read a rescue or a missing record as a failure.

`skills/myflow-contracts/finish-contract.md` is deliberately **not** edited: it describes the
invocation at lines 143–152 and cites this table rather than restating it, which is the property that
lets this row be added in one place.

- [x] **Step 1: The registry row**
- [x] **Step 2: The two outcome rows and the paragraph below the table**
- [x] **Step 3: Guards green**

```bash verified:both commands are in .myflow/project.md's ## lint list
scripts/check-references.sh && scripts/check-contract-budget.sh
```

`check-contract-budget.sh` needs no table edit: `pipeline.md` is 47,565 bytes against a declared
budget of 55,728, so both additions fit inside the existing headroom.
<!-- measured: wc -c skills/myflow-contracts/pipeline.md; scripts/check-contract-budget.sh budgets() @ branch main -->

**Tests:** none — this task changes contract prose; the lint guards named in the step above cover
it. No test name is declared here on purpose: the guard reads every backticked token in this field
as a declared test name and then looks for it in the diff.

**Regression:** Reverting this task returns the registry to naming a directory, which settles no
path, and removes the only statement of what a caller does with `rescued:` and `MISSING:`.

**Baseline:** before=0 after=0 tests; the two guards above exit 0 before and after.
<!-- predicted: scripts/check-references.sh; scripts/check-contract-budget.sh after task 3 -->

**Commit:** `docs(pipeline): name the ledger file and add the rescue and missing outcomes`

---

### 4 The writer names the path, and section 7 asserts it

**Build:** green

**Files:**
- Modify: `skills/myflow-do/SKILL.md`

**Interfaces:**
- Consumes: task 3's registry row, cited rather than copied.
- Produces: the writer-side half of the canonical path.

Two edits, both in `skills/myflow-do/SKILL.md`:

- **Section 4**, at the `Invoke **superpowers:subagent-driven-development**` paragraph (around
  line 214): state that the skill's workspace script writes this plan's ledger to
  `<abs-worktree>/.superpowers/sdd/tasks/progress.md` — because it derives the directory from the
  plan file's basename and myflow's plan is always `tasks.md` — and that this is the path
  preservation reads, citing the **Temporary artifacts registry**
  (`skills/myflow-contracts/pipeline.md`) row task 3 wrote. Say explicitly that the ledger is
  **not** written flat under `.superpowers/sdd/` like the other artifacts this section names, since
  that is the pattern-match the whole change exists to break.
- **Section 7**, beside the existing verification steps: a non-gating presence check, reported at its
  call site and never a stop.

```bash unverified:confirm the exact surrounding wording in section 7 when the file is open
test -f <abs-worktree>/.superpowers/sdd/tasks/progress.md
```

The report shape follows section 4's existing dispatch-context check verbatim in spirit: report
plainly rather than letting the run continue silently, and state that the missing ledger never gates
the run. **No new field in the handoff block** — the block is shared with `/myflow-fast` and
`/myflow-finish`, and a field added for one command changes all three.

- [x] **Step 1: Section 4 names the path and the reason**
- [x] **Step 2: Section 7's `test -f` and its report**
- [x] **Step 3: Guards green**

```bash verified:both commands are in .myflow/project.md's ## lint list
scripts/check-references.sh && scripts/check-contract-budget.sh
```

`skills/myflow-do/SKILL.md` is 71,528 bytes against a declared budget of 89,566, so no budget row
moves.
<!-- measured: wc -c skills/myflow-do/SKILL.md; scripts/check-contract-budget.sh budgets() @ branch main -->

**Tests:** none — this task changes skill prose, covered by the two guards above.

**Regression:** Reverting this task returns the writer to naming every artifact except the ledger,
which is the silence that lets each run choose its own path.

**Baseline:** before=0 after=0 tests; the two guards above exit 0 before and after.
<!-- predicted: scripts/check-references.sh; scripts/check-contract-budget.sh after task 4 -->

**Commit:** `docs(myflow-do): name the ledger's canonical path and assert it before handoff`

---

### 5 Rename the two misnamed preserved ledgers

**Build:** green

**Files:**
- Rename: `docs/superpowers/ledgers/2026-08-20-kan-102-citations-resolve-to-installed-paths-ledger.md`
- Rename: `docs/superpowers/ledgers/2026-08-08-kan-95-slim-the-myflow-contract-files-per-move-ledgers.md`

**Interfaces:**
- Consumes: nothing.
- Produces: two files `gather-self-review-context.sh`'s `find_dated()` can reach.

Both carry a hand-typed suffix that `find_dated()`'s anchored pattern
`[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]-${NAME}.md` can never match, so the self-review gather
reports them absent.

- [x] **Step 1: Read each file's first line and identify its change**

Upstream writes the ledger's identity as its first line — `# SDD ledger — plan: <plan file path>`.
Read both files' first lines and the surrounding few, and establish which archived change each
belongs to before touching either.

```bash verified:head is being used to read the file's own identity line, exactly as the upstream skill writes it
head -3 docs/superpowers/ledgers/2026-08-20-kan-102-citations-resolve-to-installed-paths-ledger.md
head -3 docs/superpowers/ledgers/2026-08-08-kan-95-slim-the-myflow-contract-files-per-move-ledgers.md
```

- [x] **Step 2: Rename the kan-102 ledger**

Its change is `kan-102-citations-resolve-to-installed-paths`, and no correctly-named ledger exists
for it, so the trailing `-ledger` is a typo:

```bash verified:the target name is the source name minus the trailing "-ledger", matching find_dated's pattern
git mv docs/superpowers/ledgers/2026-08-20-kan-102-citations-resolve-to-installed-paths-ledger.md \
       docs/superpowers/ledgers/2026-08-20-kan-102-citations-resolve-to-installed-paths.md
```

- [x] **Step 3: Decide the kan-95 file on its first line — resolved: leave it**

Its first line reads `# Per-move ledgers — kan-95-slim-the-myflow-contract-files`. It is the
**per-move ledger** `myflow-contract-economy`'s **A move or eviction is recorded in a per-move
ledger** requires — not an SDD ledger at all, merely filed in the same directory. The real SDD
ledger for that change already sits beside it as
`2026-08-08-kan-95-slim-the-myflow-contract-files.md`, so renaming would have destroyed one of the
two. Left in place; recorded in `design.md`.


`2026-08-08-kan-95-slim-the-myflow-contract-files.md` already exists beside it, so the
`-per-move-ledgers` file may be a genuinely separate record rather than a misnamed one — renaming it
onto the existing name would destroy one of the two. If step 1 shows it is the same change's ledger
split across two files, **leave it in place** and record why in the change's `design.md` under a new
`### The kan-95 second ledger` subsection of `## Design`. Only rename it if its first line shows it
belongs to a change with no preserved ledger at all.

**Tests:** none — this is a data fix with no code path.

**Regression:** Reverting this task returns both files to names `find_dated()` cannot match, making
their changes' ledgers invisible to the self-review gather.

**Baseline:** before=32 after=32 files in `docs/superpowers/ledgers/`; the count does not change,
only one name does — step 3 resolved to leaving the kan-95 file alone.
<!-- measured: ls docs/superpowers/ledgers/ | wc -l @ branch main -->

**Commit:** `chore(openspec): plan and session records` — **this task produces no task commit of its
own.** `docs/superpowers/` is a planning path: pipeline.md's Git-boundaries table and section 4's
COMMIT-PER-TASK clause both forbid a task or fixup commit from touching it, and `/myflow-finish`
run 1's second commit — that fixed literal — is what carries it. The rename is left in the working
tree and swept up there. `check-task-commit-fields.sh` is therefore **not** run for this task, having
no commit to check.

---

### 6 Artifact brevity, stated where every run loads it

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md`

**Interfaces:**
- Consumes: nothing from earlier tasks — independent of tasks 1–5.
- Produces: the rule `specs/myflow-artifact-economy/spec.md` requires.

Add one section to `skills/myflow-contracts/pipeline.md`, stating **A change's artifacts are written
brief** as the delta spec requires. Place it beside **Handoff output**, whose subject — what a run
writes for a human — is the closest neighbour.

`pipeline.md` is the one file every `/myflow-*` command loads before any other step, which is why the
rule goes there and **not** into each skill: four copies would drift, and a skill-local copy would
miss whichever command did not carry it.

Cover, in the section: the artifacts it binds; that brevity never withholds a fact; the
never-compress list; that it narrows the be-brief rule's "docs and specs stay full" carve-out for
these files only; and that no length guard is added.

**Do not** add a guard, a budget row, or a length check — the spec forbids it, since a byte budget on
a per-change artifact rewards dropping the facts the rule requires kept.

- [x] **Step 1: Write the section**
- [x] **Step 2: Guards green**

```bash verified:all four commands are in .myflow/project.md's ## lint list
scripts/check-references.sh && scripts/check-contract-budget.sh \
  && scripts/check-vocabulary.sh && scripts/check-markdown-integrity.py
```

**Tests:** none — contract prose, covered by the four guards above.

**Regression:** Reverting this task leaves artifact brevity stated nowhere a run reads, so it holds
only for whichever session was told directly.

**Baseline:** `pipeline.md` before=48,240 bytes, after=50,002, against its declared budget of 55,728
— the addition fits the existing headroom and no budget row moves. The `before` is measured on this
branch **after task 3**, not on `main`: task 3 grows the same file, so a `main` baseline is stale by
the time this task runs.
<!-- measured: wc -c skills/myflow-contracts/pipeline.md @ branch openspec/kan-77-sdd-ledger-canonical-path after 9cfca22 -->
<!-- measured: the same command after 4cf067e -->

**Commit:** `docs(pipeline): require a change's artifacts to be written brief`

---

### 7 Harness cases for a folded red task

**Build:** red

**Squash-with:** Task 8

**Files:**
- Modify: `scripts/test-check-task-commit-fields.sh`
- Modify: `scripts/test-check-task-build-green.sh`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: the executable statement of how the guard must resolve a red task. Task 8 satisfies it.

Written before the guard, asserting what `specs/myflow-task-commit-fields/spec.md`'s **A folded red
task is checked against its partner's commit** requires.

Follow the harness's existing idiom exactly: a throwaway repository under `TMPDIR`, a synthesised
`openspec/changes/<name>/tasks.md`, a real commit, one `ok:`/`FAIL:` line per assertion. Do **not**
invent a second fixture shape.

The fixture is a two-task plan: task 1 `Build: red` with `**Squash-with:** Task 2`, task 2 green.
One commit carries both tasks' files and task 2's declared subject — the shape this change's own
`f0905f2` has.

- [x] **Step 1: The four acceptance cases**

| Case | Invoked for | Expects |
|------|-------------|---------|
| red task, partner's subject | task 1 | exit 0 — the red task's own declared subject is not required |
| partner task, same commit | task 2 | exit 0 — same verdict from either id |
| union file set, red's files | task 1 | exit 0 — the partner's files are not undeclared collateral |
| union file set, partner's files | task 2 | exit 0 — and neither are the red task's |

- [x] **Step 2: The two refusal cases**

| Case | Fixture | Expects |
|------|---------|---------|
| `Squash-with:` names a missing task | red task points at Task 9, absent | exit 1; message names the missing partner |
| partner is itself `red` | both tasks tagged red | exit 1 |

- [x] **Step 3: A green task with no `Squash-with:` is unaffected**

Assert the existing single-task behaviour is unchanged — a green task with its own commit still
fails on an undeclared file and on a subject mismatch, exactly as before. This is the case that stops
task 8 widening the check for every task rather than only a folded one.

- [x] **Step 4: Confirm the harness fails against the unmodified guard**

The red half of this pair: run it before task 8 and confirm the new cases fail loudly rather than
passing vacuously.

```bash unverified:run this before task 8 lands; the expected result is a non-zero exit naming the new cases as failures
bash scripts/test-check-task-commit-fields.sh 2>&1 | tail -5
```

**Tests:** the cases enumerated in steps 1–3, in `scripts/test-check-task-commit-fields.sh`.

**Regression:** Reverting this task leaves the folded-red-task resolution with no executable
statement, so a guard that ignores `Squash-with:`, one that widens the file set for every task, and
one that resolves against a missing partner all become unasserted.

**Baseline:** before=55 after=170 `ok:` assertions in `scripts/test-check-task-commit-fields.sh`,
beside 54 in `scripts/test-check-task-build-green.sh` (34 before this change). The task itself landed
68; the review panel's eleven fix rounds took it to 170 as they closed F1-F23 across multi-partner
folds, joined folds, the field gate, the shared grammar, field selection, the `Build:` tag, task
headings, body boundaries, duplicated ids, the fence rule and unclosed fences. `Baseline:` is
skipped-not-verified in this project, so no guard catches a stale figure here — it has been corrected
by hand three times, which is itself worth knowing.
<!-- measured: bash scripts/test-check-task-commit-fields.sh | grep -c '^ok:' @ branch openspec/kan-77-sdd-ledger-canonical-path, before this task -->
<!-- measured: the same command @ the same branch after fix round 11 (55e029c) -->

**Commit:** `test(check-task-commit-fields): assert a folded red task resolves to its partner`

---

### 8 Resolve a red task against its partner in the guard

**Build:** green

**Files:**
- Modify: `scripts/check-task-commit-fields.py`
- Modify: `scripts/check-task-commit-fields.sh`
- Modify: `scripts/check-task-build-green.py`
- Modify: `scripts/check-task-build-green.sh`
- Add: `scripts/lib/plan_grammar.py`

**Interfaces:**
- Consumes: task 7's cases.
- Produces: a guard that passes on the folded commits the pipeline already produces.

`check-task-commit-fields.py` parses `**Squash-with:**` only so the field terminates a preceding
field's continuation — its own comment at lines 94–97 says it reads neither that value nor
`**Build:**`. Give it the resolution the delta spec requires:

- When the task named on the command line is tagged `**Build:** red`, read its `**Squash-with:**`
  partner id, resolve that task in the same plan, and check against the partner's `**Commit:**`
  subject with a `**Files:**` set that is the union of both tasks' declared files.
- Apply the **same** resolution when invoked for the partner, so either id gives one verdict.
- A `**Squash-with:**` naming a task absent from the plan fails, with the missing id named.
- A partner itself tagged `red` fails.
- A task with no `**Squash-with:**` is unchanged — do not widen the file set for ordinary tasks.

This is why the change's own `f0905f2` currently exits 1 for both task 1 and task 2; once this task
lands, re-running the guard against that commit for both ids is the first real check of the fix.

- [x] **Step 1: Read `Build:` and `Squash-with:`, resolve the partner**
- [x] **Step 2: Union the file sets and take the partner's subject**
- [x] **Step 3: The two refusal paths**
- [x] **Step 4: Run the harness green, then re-check `f0905f2` for both ids**

```bash unverified:f0905f2 is this branch's tasks 1+2 commit; both invocations are expected to exit 0 once this task lands
bash scripts/test-check-task-commit-fields.sh
scripts/check-task-commit-fields.sh . 1 f0905f2 f09f2b1
scripts/check-task-commit-fields.sh . 2 f0905f2 f09f2b1
```

**Tests:** none added here — task 7 carries them.

**Regression:** Reverting this task makes every `Build: red` task fail the guard again, since its
folded commit can never carry its own declared subject.

**Baseline:** before=170 after=170 `ok:` assertions; task 7's count is unchanged and all pass.
<!-- measured: bash scripts/test-check-task-commit-fields.sh | grep -c '^ok:' @ the same branch after fix round 11 (55e029c) -->

**Commit:** `fix(check-task-commit-fields): resolve a folded red task against its partner`
