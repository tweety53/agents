> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** A guard that scans a corpus says what it actually checked, per member, on a passing run —
so a rule covering nothing announces itself instead of reporting clean forever.

## Global Constraints

- **Task 1 first; tasks 2-5 are independent of each other.** Task 1 builds the shared helper the
  other four adopt. After it lands, the four guard tasks share no file and may run in any order.
- **No existing check changes.** Every task adds reporting and the one new violation class. A task
  that alters what its guard already enforces has exceeded its scope.
- **One helper, not four copies.** KAN-73's own review raised `resolve_file` duplicated five times as
  a Critical; `scripts/lib/` exists because two duplicated helpers had already drifted. This plan
  extracts first and adopts second, deliberately.
- **Each guard's expected-zero set is written, never inferred.** Inferring it from the tree restates
  the assumption the zero already encodes and passes exactly the case this change exists to fail.
- **No task edits `openspec/` or `docs/superpowers/`.**

## Baseline

All measured 2026-08-18 against `a805d29`.

- The four in-scope guards and their harnesses, in bytes: `check-guard-symlinks` 32671 / 23997,
  `check-references` 18486 / 16287, `check-vocabulary` 30257 / 5952, `check-stage-mark-calls`
  13328 / 19547.
  <!-- measured: wc -c over each scripts/check-*.sh and scripts/test-check-*.sh @ a805d29 -->
- Their current clean verdicts, verbatim:
  <!-- measured: each guard run against the clean tree @ a805d29 -->

```text verified:run against the clean tree at a805d29 on 2026-08-18; these four lines are that run's output
GUARD-SYMLINKS-OK: <repo> — 53 guard(s) across 6 skill(s) validated
check-references: all referenced sections resolve
✓ Panel-vocabulary guard: clean
✓ Stage-mark-calls guard: clean (37 `stage begin` call(s) checked)
```

- So two guards report a corpus-wide total and two report no count at all. **None reports per
  member**, which is the gap: `53 across 6 skills` is true whether one of those six contributed 17 or
  contributed nothing.
- `scripts/lib/` already holds two shared helpers, `panel-record.sh` and `resolve-file.sh`.
  <!-- measured: ls scripts/lib/ @ a805d29 -->
- All 16 `scripts/test-check-*.sh` harnesses already assert a failure, which is why a
  failure-fixture rule was rejected rather than built.
  <!-- measured: grep for non-zero-exit and INVALID-verdict assertions across scripts/test-check-*.sh @ a805d29 -->

---

### 1 `scripts/lib/coverage.sh` — per-member coverage, defined once

**Build:** green

**Files:**
- Create: `scripts/lib/coverage.sh`
- Create: `scripts/test-lib-coverage.sh`
- Modify: `.myflow/project.md`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: the record/report/verdict functions tasks 2-5 each source, and the expected-zero
  comparison that turns an undeclared zero into a violation.

The helper owns three things, so that four guards cannot disagree about any of them: recording a
per-member count, rendering the members-and-counts fragment a verdict line carries, and deciding
whether a zero is declared or a violation.

It follows the three disciplines `scripts/lib/panel-record.sh`'s header states for every guard in
this repository — `-a` on every `grep`, the `rc > 1` split between "no match" and a real error, and
`--` before every path — and says in its own header that it adopted them from there rather than
inventing them.

- [x] **Step 1: The harness first**

`scripts/test-lib-coverage.sh`, sourcing the library directly rather than through a guard. Cases: a
member with a non-zero count renders and passes; a member at zero **and** declared renders as
declared and passes; a member at zero and **undeclared** is reported by name and fails; an empty
corpus is itself a violation rather than a vacuous pass; and a member name containing a space or a
leading `-` is handled rather than mis-parsed.

Watch every case fail before the library exists.

- [x] **Step 2: The library**

Portable to macOS `/bin/sh` and bash 3.2 — no associative arrays, no `mapfile`, no `readlink -f`.
Note this repo has been bitten by unquoted `$VAR` not word-splitting under zsh, and by macOS `awk`
ignoring `--`; write and run loops under `bash` explicitly.

- [x] **Step 3: Wire the harness into the project's test list**

Add `scripts/test-lib-coverage.sh` to `.myflow/project.md`'s `## test` list. It needs no `## lint`
entry: the library is not a guard and answers no question about the tree on its own.

**Tests:** the five cases in `scripts/test-lib-coverage.sh` named in step 1.

**Regression:** Reverting this task leaves each of tasks 2-5 to define its own coverage rendering and
its own expected-zero comparison, which is the five-copy `resolve_file` situation KAN-73's panel
raised as a Critical, reproduced deliberately.

**Baseline:** before=2 after=3 files in `scripts/lib/`.
<!-- measured: ls scripts/lib/ @ a805d29 listed panel-record.sh and resolve-file.sh -->

**Commit:** `feat(kan-197-require-mutation-test-for-every-guard): add scripts/lib/coverage.sh`

---

### 2 `check-guard-symlinks.sh` reports per-skill coverage

**Build:** green

**Files:**
- Modify: `scripts/check-guard-symlinks.sh`
- Modify: `scripts/test-check-guard-symlinks.sh`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: task 1's library.
- Produces: the first guard to carry the new verdict shape; tasks 3-5 follow it.

This is the guard whose defect motivated the change, so it carries the regression case.

Its corpus is the skills under `skills/`. Today it reports `53 guard(s) across 6 skill(s)` — a total
that was **already true** while rule 2 covered nothing whatever for `myflow-fast`.

- [x] **Step 1: The regression fixture, written first**

A fixture reproducing KAN-73's own shape: a skill that carries symlinks but names no guard in a form
rule 2's classifier can see — a delegating skill. Assert the guard **fails**, names that skill, and
says its coverage is zero.

Watch it fail before the guard changes: today that fixture passes, which is the defect.

Add the paired case too — a skill legitimately at zero and declared — so the declaration is exercised
in both directions.

- [x] **Step 2: Adopt the library and declare the expected zeros**

`myflow-start`, `myflow-status` and `openspec-explore` invoke no guard and are declared, each with
its reason recorded. `myflow-contracts` is **not** declared — it is filtered out of this guard's
corpus by name and is not a member, so a declaration for it would be inert. `myflow-fast` is **not**
declared either: it delegates, and its required set must resolve through the delegation rather than
be excused.

- [x] **Step 3: The verdict carries the breakdown**

```text unverified:the shape this task introduces; the guard prints only the corpus-wide total today
GUARD-SYMLINKS-OK: <repo> — 53 guard(s) across 6 skill(s) validated
  myflow-do 14 · myflow-fast 17 · myflow-finish 7 · myflow-status 0 (declared) ·
  myflow-start 0 (declared) · myflow-contracts 0 (declared)
```

- [x] **Step 4: Green against the real tree**

```bash verified:the guard and its harness both exist and pass at a805d29; the counts rise by the cases step 1 adds
scripts/test-check-guard-symlinks.sh && scripts/check-guard-symlinks.sh
```

A failure against the real tree is a real finding about this repository, not a reason to declare a
member expected-zero to make it quiet.

**Tests:** the delegating-skill regression case and the declared-zero case, both in
`scripts/test-check-guard-symlinks.sh`.

**Regression:** Reverting this task returns the guard to reporting a corpus-wide total that stays
true while a rule covers nothing for one skill — the exact state KAN-73 shipped and that a mutation,
not the guard, had to find.

**Baseline:** before=0 after=1 guards reporting per-member coverage.

**Commit:** `feat(kan-197-require-mutation-test-for-every-guard): report per-skill coverage in check-guard-symlinks.sh`

---

### 3 `check-references.sh` reports per-file coverage

**Build:** green

**Files:**
- Modify: `scripts/check-references.sh`
- Modify: `scripts/test-check-references.sh`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: task 1's library, and task 2's verdict shape as the pattern to follow.
- Produces: nothing later tasks depend on.

Its corpus is the Markdown files it scans for references. It reports `all referenced sections
resolve` today — no count, so a file contributing no checked reference is invisible.

- [x] **Step 1: The failing cases first**

A scanned file whose references all resolve reports its count. A scanned file contributing **zero**
checked references, undeclared, fails by name. A declared one passes.

- [x] **Step 2: Adopt the library, declare the expected zeros**

Establish which scanned files legitimately carry no checkable reference before declaring any — a
declaration written to silence a file nobody examined is the failure mode this change exists to stop.

- [x] **Step 3: Green against the real tree**

```bash verified:both exist and pass at a805d29
scripts/test-check-references.sh && scripts/check-references.sh
```

**Tests:** the zero-coverage and declared-zero cases added to `scripts/test-check-references.sh`.

**Regression:** Reverting this task lets a file drop out of the reference scan — a rename, a changed
glob — without the guard's output changing at all.

**Baseline:** before=1 after=2 guards reporting per-member coverage.

**Commit:** `feat(kan-197-require-mutation-test-for-every-guard): report per-file coverage in check-references.sh`

---

### 4 `check-vocabulary.sh` reports per-file coverage

**Build:** green

**Files:**
- Modify: `scripts/check-vocabulary.sh`
- Modify: `scripts/test-check-vocabulary.sh`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: task 1's library, and task 2's verdict shape.
- Produces: nothing later tasks depend on.

Its corpus is the files it scans for the closed vocabulary. It reports `clean` today, with no count
at all — the weakest of the four, since its output is identical whether it scanned the whole tree or
nothing.

- [x] **Step 1: The failing cases first**

Same three shapes as task 3: counted, undeclared zero fails by name, declared zero passes.

- [x] **Step 2: Adopt the library, declare the expected zeros**

- [x] **Step 3: Green against the real tree**

```bash verified:both exist and pass at a805d29
scripts/test-check-vocabulary.sh && scripts/check-vocabulary.sh
```

**Tests:** the zero-coverage and declared-zero cases added to `scripts/test-check-vocabulary.sh`.

**Regression:** Reverting this task returns the guard to a bare `clean`, which is the same output it
would print having examined no file whatsoever.

**Baseline:** before=2 after=3 guards reporting per-member coverage.

**Commit:** `feat(kan-197-require-mutation-test-for-every-guard): report per-file coverage in check-vocabulary.sh`

---

### 5 `check-stage-mark-calls.sh` reports per-file coverage

**Build:** green

**Files:**
- Modify: `scripts/check-stage-mark-calls.sh`
- Modify: `scripts/test-check-stage-mark-calls.sh`

**Allowed-collateral:** *(none)*

**Interfaces:**
- Consumes: task 1's library, and task 2's verdict shape.
- Produces: nothing later tasks depend on.

Its corpus is the skill files carrying `stage begin` call sites. It already reports a total —
`37 stage begin call(s) checked` — so this task splits that total per file rather than inventing one.

A skill that loses its marks entirely, or is renamed out of the glob, currently moves that total
without naming anything.

- [x] **Step 1: The failing cases first**

Same three shapes. The zero case here is a skill file the glob reaches that carries no `stage begin`
call at all.

- [x] **Step 2: Adopt the library, declare the expected zeros**

`myflow-status` marks nothing by contract — a read-only report that wrote stage runs would be
recording work nobody did — so it is a legitimate declared zero, and its declaration should cite that
reason rather than merely list the name.

- [x] **Step 3: Green against the real tree**

```bash verified:both exist and pass at a805d29
scripts/test-check-stage-mark-calls.sh && scripts/check-stage-mark-calls.sh
```

**Tests:** the zero-coverage and declared-zero cases added to
`scripts/test-check-stage-mark-calls.sh`.

**Regression:** Reverting this task returns the guard to a single total, which stays plausible while
one skill's marks vanish entirely.

**Baseline:** before=3 after=4 guards reporting per-member coverage.

**Commit:** `feat(kan-197-require-mutation-test-for-every-guard): report per-file coverage in check-stage-mark-calls.sh`
