> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** One row per change on the dashboard. Stop `/myflow-fast` from marking a stage under a name
nobody has resolved yet, make that rule mechanical, and delete the row the defect already produced.

## Global Constraints

- **The synthetic bootstrap does not change.** `ApplyBeginStageMark`
  (`stats/internal/api/stages.go`), `stages.SyntheticChangeUpdatedBy`
  (`stats/internal/stages/synthetic.go`) and `state get`'s `"synthetic": true` reporting
  (`stats/cmd/myflow/state.go`) are kan-174's recorded decision and are left exactly as they are.
  **No Go or TypeScript source is edited by any task in this plan.**
- **The state gate's *read* keeps its best guess.** `myflow state get` writes nothing; it is the
  mark, and only the mark, that must wait for a resolved name.
- **No task edits `openspec/` or `docs/superpowers/`.**
- **Task 4 runs against the shared development database on `localhost:5433`, never an apply
  worktree's isolated `myflow_<id>` database.** Getting this wrong deletes nothing and reports
  success.

## Baseline

**Measured 2026-08-15 against `c07fed7`:** `scripts/test-check-stage-mark-calls.sh` prints 19 `ok:`
assertions across its 10 cases and exits 0. `scripts/check-stage-mark-calls.sh` over `skills/`
reports `clean (37 stage begin call(s) checked)`.
<!-- measured: both scripts run on 2026-08-15 from the main checkout at c07fed7 -->

**Contract budget headroom, measured the same way:** `skills/myflow-contracts/pipeline.md` 41042
bytes of 44574 budgeted; `skills/myflow-fast/SKILL.md` 16632 of 18225.
<!-- measured: wc -c against the two files, budgets read from scripts/check-contract-budget.sh -->

---

### 1 Teach the guard that a mark may not name a guess

**Build:** green

**Files:**
- Modify: `scripts/test-check-stage-mark-calls.sh`
- Modify: `scripts/check-stage-mark-calls.sh`

**Interfaces:**
- Consumes: `assemble_calls`, which already folds continuation lines into one logical command
  string — the new check reads that same assembled text, so a change argument written on a
  continuation line is checked like any other.
- Produces: a fourth finding class on `stage begin` calls, alongside the missing/substituted
  session token, the missing harness, and the hardcoded harness.

The guard cannot know whether `<name>` is resolved at a given call site. It can know that a
placeholder whose own text says "guess" is not one — that is the whole claim, and it is enough to
fail today's `skills/myflow-fast/SKILL.md`.

- [x] **Step 1: Failing fixtures, written first**

Add cases to `scripts/test-check-stage-mark-calls.sh`, following the existing `new_fixture` /
`run_guard` / assert shape exactly:

- a `stage begin` whose trailing argument is `<name-or-best-guess>` → exit 1, and the finding text
  names the argument;
- the same shape written across continuation lines → exit 1, proving the check reads the assembled
  command and not the physical line;
- a `stage begin` whose trailing argument is `<name>` → exit 0, so the check does not fire on the
  compliant form every other skill uses;
- `myflow stage end ... <name-or-best-guess>` → exit 0, since `stage end` is deliberately not
  examined at all.

Run the harness and watch the first two fail before touching the guard.

- [x] **Step 2: The check**

In the per-call loop, extract the **last** whitespace-delimited token of the assembled command as
the change argument, and emit a violation when it matches a bracketed placeholder whose text
contains `guess`, case-insensitively — `<name-or-best-guess>` and any restatement of it. The
message names the file, the line, the offending argument, and why: a mark writes, so a guessed name
bootstraps a change row that outlives the run.

Leave the extraction narrow and obvious. A token that is not bracketed is not this check's business.

- [x] **Step 3: Document it in the header**

Extend the `WHAT THIS CATCHES` header block with the fourth rule and the incident behind it —
`kan-175` and `kan-175-more-ui-ux-fixes` appearing as two open changes on 2026-08-15, 29 seconds
apart. The other three rules each carry their own reasoning there; this one is no different.
<!-- measured: the two rows' updatedAt values, read from the state-board API on 2026-08-15 -->

**Tests:** `scripts/test-check-stage-mark-calls.sh` — the four cases in step 1.

**Regression:** Reverting this task lets a `stage begin` name a guessed change in skill source with
nothing objecting, which is exactly how `kan-175` was created.

**Baseline:** before=19 after=23 `ok:` assertions from `scripts/test-check-stage-mark-calls.sh`.
<!-- predicted: four new assertions, one per case in step 1; none written yet -->

**Commit:** `feat(1): reject a stage mark that names a guessed change`

---

### 2 Defer `/myflow-fast`'s state-gate mark until the name is resolved

**Build:** green

**Files:**
- Modify: `skills/myflow-fast/SKILL.md`

**Interfaces:**
- Consumes: `/myflow-start` section A's creating-run rule, which this section now mirrors.
- Produces: a **State gate** section whose marks carry `<name>`, and a stated firing point for a
  creating run.

**Task 1's guard fails against this file until this task lands.** That ordering is deliberate: the
guard is the test, and this is the fix it proves.

- [x] **Step 1: The marks take a resolved name**

In the **State gate** section, change both `myflow stage begin` and `myflow stage end` for
`do.state-gate` to take `<name>`. Leave the `myflow state get <name-or-best-guess>` read, and the
three-way exit-code reading beneath it, byte-for-byte unchanged — that read is what makes the gate
work before a name exists.

- [x] **Step 2: Say when the pair fires**

State, in that section, that on a creating run the `begin`/`end` pair fires back to back at the
point section A fixes the change name, before `start.resolve-change`'s own marks, and cite
`/myflow-start`'s **A. Resolve the change** for the identical rule rather than restating its
reasoning. State that at `IN_PROGRESS` the name is already resolved in every path, so the marks fire
where the gate runs. Say plainly what this costs — a creating run's `do.state-gate` records a
near-zero duration, accepted because the alternative buys a duration by creating a change row for a
name nobody chose.

- [x] **Step 3: Check the budget**

`scripts/check-contract-budget.sh` must still pass. The file has 1593 bytes of headroom; if the
edit exceeds it, raise that file's row in the guard's `budgets()` table deliberately and say so in
the commit body — never trim the explanation to fit.

**Tests:** No new test — this task's test is the guard task 1 added, run over the skills tree: red
before this task, green after. The contract-budget, references and vocabulary guards must be clean
too. Named in prose rather than as declared tests, since this commit adds no test of its own.

**Regression:** Reverting this task restores `<name-or-best-guess>` in the mark, which task 1's
guard then rejects — the two tasks hold each other in place.

**Baseline:** before=37 after=37 `stage begin` calls checked by `check-stage-mark-calls.sh`, with
the run moving from exit 1 to exit 0.
<!-- predicted: the edit changes an argument, adds no call and removes none -->

**Commit:** `fix(2): mark /myflow-fast's state gate only once the name resolves`

---

### 3 State the rule once, in the contract

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/pipeline.md`

**Interfaces:**
- Consumes: the **Stage marks** section, which already states the session-token and harness rules
  that `check-stage-mark-calls.sh` enforces.
- Produces: one paragraph stating that a mark's `<change>` argument is a resolved change name.

- [x] **Step 1: The paragraph**

Under **Stage marks**, beside the existing required-flag rules, state that the `<change>` argument
is always a resolved change name and never a guess; that marking writes, so a guessed name
bootstraps a change row that outlives the run, appears among the open changes and is never archived;
and that `scripts/check-stage-mark-calls.sh` rejects a call site written that way. Point at the
sibling rule already in the specs — a state gate reads the state before it marks — rather than
re-deriving why marking writes.

Do not restate the bootstrap's own reasoning; `stats/internal/stages/synthetic.go` is canonical for
it.

- [x] **Step 2: Check the budget**

`scripts/check-contract-budget.sh` must still pass — 3532 bytes of headroom on this file. The same
rule as task 2 applies if it does not: raise the row deliberately, do not trim the rule.

**Tests:** No new test — a contract paragraph has no executable surface. Verified by the
contract-budget, references, vocabulary and markdown-integrity guards, named in prose rather than as
declared tests, since this commit adds no test of its own.

**Regression:** Reverting this leaves the rule enforced by a guard but stated nowhere a reader of
the contract would find it, which is how the placeholder survived review the first time.

**Baseline:** before=41042 after≤44574 bytes for `skills/myflow-contracts/pipeline.md`.
<!-- measured: wc -c on 2026-08-15; the budget is this file's row in scripts/check-contract-budget.sh -->

**Commit:** `docs(3): state that a stage mark names a resolved change`

---

### 4 Delete the stray `kan-175` row

**Build:** green

**Files:**
- No repository file is modified by this task.

**Allowed-collateral:** *(none — this task's whole effect is on the development database)*

**Interfaces:**
- Consumes: the `changes`, `change_repos` and `stage_runs` tables
  (`stats/internal/store/migrations/0001_init.sql`, `0002_change_repos.sql`, `0003_stage_runs.sql`).
- Produces: one fewer row in `changes`.

**This task edits no file and therefore produces no commit of its own.** Record its evidence in the
SDD ledger and in the handoff instead.

- [x] **Step 1: Confirm the target, before deleting anything**

Against the shared development database — **`localhost:5433`, database `myflow`, not any worktree's
`myflow_<id>`** — read back both rows so the one being deleted is identified by `id`, not by name
alone:

```sql unverified:written from the migration DDL, not yet run — check that exactly two rows come back and that only `kan-175` carries the synthetic sentinel
SELECT id, project_key, name, state, updated_by
FROM changes
WHERE project_key = 'gymie-7c1f238a' AND name LIKE 'kan-175%';
```

Expect two rows: `kan-175-more-ui-ux-fixes` (keep) and `kan-175` (delete).

- [x] **Step 2: Delete it, in one transaction, children first**

Neither `stage_runs.change_id` nor `change_repos.change_id` cascades, so the order is fixed:

```sql unverified:written from the migration DDL's foreign keys; run it only after step 1 has fixed :id to the `kan-175` row
BEGIN;
DELETE FROM stage_runs   WHERE change_id = :id;
DELETE FROM change_repos WHERE change_id = :id;
DELETE FROM changes      WHERE id = :id;
COMMIT;
```

Match on `id`, never on `name LIKE 'kan-175%'` — that pattern also matches the change being kept.

- [x] **Step 3: Verify from the outside**

Re-run step 1's SELECT: one row, `kan-175-more-ui-ux-fixes`. Then confirm through the API that the
board agrees, which is the surface the defect was reported against:

```bash verified:this exact request was run on 2026-08-15 and returned both kan-175 rows, which is how the defect was confirmed
curl -s -H 'Myflow-Daemon: 1' \
  'http://127.0.0.1:4173/api/v1/stats/state-board?from=2020-01-01T00:00:00Z&to=2030-01-01T00:00:00Z'
```

**Tests:** No automated test. The verification is step 3's two reads, and their output belongs in
the handoff.

**Regression:** Nothing re-creates this row once task 2 has landed; if it reappears, task 2 did not
take.

**Baseline:** before=2 after=1 rows matching `kan-175%` in project `gymie-7c1f238a`.
<!-- measured: the two rows were read from the state-board API on 2026-08-15 -->

**Commit:** *(none — this task changes no file)*
