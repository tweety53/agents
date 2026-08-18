> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** Every self-review angle produces findings, every finding is explained before it is filed,
and the report that records all of it is checked by a guard rather than trusted to prose.

## Global Constraints

- **Task 1 first.** It restores a seventeenth report into `docs/self-review/`, and task 2's declared
  pre-rule list must cover it. Running task 2 first would produce a list that is wrong the moment
  task 1 lands.
- **Tasks 3, 4 and 5 are documentation-only and independent of each other**, but task 3 is the only
  one that touches `scripts/check-contract-budget.sh`, so no other task may edit that file.
- **The guard is never named in skill text.** `scripts/check-self-review-report.sh` is a repository
  lint guard, not something a `/myflow-*` run invokes. Naming it in an invoking position inside
  `skills/` would oblige `check-guard-symlinks.sh` rule 2 to require a symlink for a guard no command
  calls. The contracts state the report *shape*; only `.myflow/project.md` names the guard.
- **No existing rule is weakened.** Every task adds a rule or restates an existing one in one
  canonical place. A task that removes a check has exceeded its scope.
- **No task edits `openspec/` or `docs/superpowers/`** — except task 1, whose entire deliverable is a
  file under `docs/self-review/`, which is neither.

## Baseline

All measured 2026-08-18 against `d5cd0a5`.

- Files this change modifies, in bytes: `skills/myflow-contracts/finish-contract.md` 30712,
  `skills/myflow-contracts/jira-integration.md` 15758,
  `skills/myflow-contracts/jira-followups.md` 35721, `skills/myflow-finish/SKILL.md` 33879.
  <!-- measured: wc -c over each path @ d5cd0a5 -->
- Their declared budgets in `scripts/check-contract-budget.sh`: 33505, 19698, 44651, 39965
  respectively — so only `finish-contract.md`, with 2793 bytes of headroom, is expected to need a
  budget change.
  <!-- measured: budgets() table in scripts/check-contract-budget.sh @ d5cd0a5 -->
- `docs/self-review/` holds **16** reports; task 1 makes it 17.
  <!-- measured: ls docs/self-review/*.md | wc -l @ d5cd0a5 -->
- `scripts/lib/coverage.sh` exists (11253 bytes) and is already adopted by four guards; this change
  adds a fifth caller and no new helper.
  <!-- measured: wc -c scripts/lib/coverage.sh @ d5cd0a5 -->
- KAN-197's self-review report exists only as commit `0aea735`, on the merged-and-deleted
  `chore/archive-kan-197`; `main` has never carried it.
  <!-- measured: git log --oneline --all -- docs/self-review/kan-197-*.md @ d5cd0a5 -->

## The report shape this change defines

Every task below refers to this one shape; it is stated here once rather than in each task.

Each angle is an `##` heading whose line contains that angle's label:

```markdown verified:the shape this change introduces; asserted by task 2's harness
## Problems and fixes — `myflow-fix`

- **[myflow-fix]** Preflight compares against a stale local base ref — declined

## Cost — `myflow-cost`

- **[myflow-cost]** Every panel slot gathers the same context independently — filed: KAN-201

## What went well — `myflow-improvement`

_none — this angle produced no findings._
```

- a finding line is `- **[<label>]** <text> — filed: <KEY>` or `- **[<label>]** <text> — declined`;
- the label on a finding line matches the label in its section's heading;
- an angle with no findings carries `_none — this angle produced no findings._` and no finding lines.

---

### 1 Recover KAN-197's stranded self-review report

**Build:** green

**Files:**
- Create: `docs/self-review/kan-197-require-mutation-test-for-every-guard-self-review.md`

**Allowed-collateral:** *(none)*

**Tests:** *(none — this task restores a document and adds no behaviour)*

**Regression:** Reverting this commit removes the only copy of KAN-197's self-review report from
`main` again, returning it to existing solely on a merged-and-deleted branch.

**Baseline:** before=0 after=0

**Commit:** `docs(kan-200-self-review-filing-ask-per-angle): recover KAN-197's self-review report`

**Interfaces:**
- Consumes: nothing.
- Produces: the seventeenth report in `docs/self-review/`, which task 2's declared pre-rule list must
  name.

Commit `0aea735` was pushed to `chore/archive-kan-197` at 11:09Z on 2026-08-18, five minutes after
PR #13 merged at 11:04Z. The file it carries never reached `main`.

- [x] **Step 1: Restore the file from the commit**

```bash unverified:the object is reachable through the remote-tracking ref at planning time; fetch first if it is not
git -C <worktree> checkout 0aea735 -- docs/self-review/kan-197-require-mutation-test-for-every-guard-self-review.md
```

If the object is unreachable, `git -C <worktree> fetch origin` first. Do **not** cherry-pick: a
cherry-pick brings its own commit message, and this task's commit must carry the subject declared
above plus the `Task-Id:` trailer.

- [x] **Step 2: Verify the content is the whole report**

Confirm the restored file is byte-identical to the version in `0aea735` and that it is a complete
report, not a truncated one:

```bash unverified:run during implementation
git -C <worktree> diff --no-index --stat <(git -C <worktree> show 0aea735:docs/self-review/kan-197-require-mutation-test-for-every-guard-self-review.md) docs/self-review/kan-197-require-mutation-test-for-every-guard-self-review.md
```

**This report predates the five-angle shape and is not rewritten to fit it.** It is a record of a run
that is over; task 2 declares it pre-rule alongside the other sixteen.

---

### 2 `scripts/check-self-review-report.sh` — the guard and its harness

**Build:** green

**Files:**
- Create: `scripts/check-self-review-report.sh`
- Create: `scripts/test-check-self-review-report.sh`
- Modify: `.myflow/project.md`

**Allowed-collateral:** *(none)*

**Tests:** `scripts/test-check-self-review-report.sh`

**Regression:** Reverting this commit removes the only mechanical check that a self-review report
carries all five angles; a report omitting an angle would again land unnoticed, which is the exact
failure KAN-200 records.

**Baseline:** before=0 after=21 — the harness's twenty-one cases, from a corpus of zero before it exists. Nine at plan time; three panel fix rounds added twelve more, each mutation-proved — the fix reverted in a scratch tree, the case confirmed to fail, the fix restored.

**Commit:** `feat(kan-200-self-review-filing-ask-per-angle): check every self-review report's shape`

**Interfaces:**
- Consumes: task 1's restored report, which the declared pre-rule list must name; `scripts/lib/coverage.sh`.
- Produces: the guard `.myflow/project.md` registers in `## lint`, and its harness in `## test`.

The guard scans `docs/self-review/*.md`. It reports per-member coverage through
`scripts/lib/coverage.sh` — it is the library's fifth caller and adds no helper of its own — so a
report that is checked for nothing is a named, failing fact rather than a silent pass.

- [x] **Step 1: The harness first**

`scripts/test-check-self-review-report.sh`, driving the guard against `mktemp -d` fixtures. Nine
cases, each asserting the exit code **and** the verdict line, never one alone:

1. a fully compliant report passes, verdict `SELF-REVIEW-REPORT-OK`, exit 0;
2. a report missing the angle-5 section is named, with the missing label, exit 1;
3. a section carrying neither a finding line nor the none-marker is named, exit 1;
4. a finding line whose label does not match its section's heading is named, exit 1;
5. a finding marked `filed:` with no issue key is named, exit 1;
6. a declared pre-rule report is reported as declared, with its reason, exit 0;
7. **the KAN-197 regression shape** — a report present in the corpus, absent from the declared list,
   for which zero checks were performed — is named as an undeclared zero, exit 1;
8. an empty corpus is a violation, not a vacuous pass, exit 1;
9. an unreadable or absent `docs/self-review/` directory exits 2, distinctly from 1.

- [x] **Step 2: The guard**

Verdict line on success, carrying the coverage fragment `coverage_report` renders:

```text unverified:the intended verdict shape; asserted by step 1's case 1
SELF-REVIEW-REPORT-OK: <repo> — 17 report(s), 1 checked, 16 declared pre-rule
```

Exit-code contract in the script's own header: **0** clean, **1** violations found, **2** cannot
answer at all. Follow the three disciplines `scripts/lib/panel-record.sh`'s header states for every
guard here — `-a` on every `grep`, the `rc > 1` split between "no match" and a real error, and `--`
before every path — and say in the header that they were adopted from there.

Each per-member count is the number of checks actually performed on that report: its five section
checks plus one per finding line. A report that yields zero is either declared or a violation; there
is no third outcome.

- [x] **Step 3: The declared pre-rule list**

Seventeen `coverage_declare` calls in the guard's own source — the sixteen reports present at
`d5cd0a5` plus task 1's restored KAN-197 report — each with the reason
`predates the five-angle report shape (KAN-200)`. Written out by name, never derived from a date, a
sort order, or a marker in the file: a marker a new report can forget to write is a mechanism for
passing without being checked, and that is the outcome the declaration exists to prevent.

- [x] **Step 4: Register it**

Add `scripts/check-self-review-report.sh` to `.myflow/project.md`'s `## lint` list and
`scripts/test-check-self-review-report.sh` to its `## test` list. Add no count to either — the file
states, twice, that it cites those lists by count nowhere.

- [x] **Step 5: Verify by breaking the tree**

Delete the angle-5 section from a scratch copy of a compliant report and confirm the guard fails with
that report named; restore it and confirm the guard passes. A guard whose failure path has not been
observed on a real tree has not been verified.

---

### 3 The finish contract — five angles, per-angle filing, explain first

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/finish-contract.md`
- Modify: `skills/myflow-contracts/operator-prompts.md`
- Modify: `scripts/check-contract-budget.sh`

**Allowed-collateral:** *(none)*

**Tests:** *(none — this task changes documentation and adds no test)*

**Regression:** Reverting this commit returns step 8 to four angles and a per-finding ask, removing
the canonical statement every other file in this change cites.

**Baseline:** before=0 after=0

**Commit:** `feat(kan-200-self-review-filing-ask-per-angle): state the five self-review angles once`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: the canonical angle table and filing rules that tasks 4 and 5 cite rather than restate.

Step 8 of **Run 2 — the branch is merged** is the runtime statement of the procedure. It gains:

- [x] **Step 1: The angle table**

Five rows — angle number, angle, label — with angle 5 reading *what could move to the Go app or its
persistent storage*, label `myflow-stats-app`, and one sentence fixing its remit: records the
pipeline writes to files today, and derivation work now done in Bash or by the agent; **not** what
the SPA should display.

- [x] **Step 2: Every angle produces output**

Each angle produces zero or more findings, and an angle that produces none says so explicitly — the
same present-but-empty rule `## Decisions` and `## Open questions` already carry. State why in one
sentence: a silent angle and a skipped angle are indistinguishable to a reader, which is how KAN-73's
cost angle passed unnoticed while its section existed.

- [x] **Step 3: Explain before filing**

Every finding is explained in the message body before any prompt fires, carrying what was observed,
what breaks, and what the fix would be. The prompt records the decision only. State that a prompt's
option text cannot carry an explanation, and that a filed issue is durable — an explanation arriving
afterwards describes something the operator did not agree to.

- [x] **Step 4: One multi-select per angle, and the angle label**

The ask is one multi-select prompt per angle over that angle's findings, defaulting to filing none.
A filed issue carries its angle's label on top of the set **Labels on issues the pipeline creates**
(`skills/myflow-contracts/jira-integration.md`) already defines — cite that section for the inherited
set rather than restating it.

- [x] **Step 5: The report shape**

State the report shape from this plan's own section above: five sections, one parseable line per
finding carrying label, text and disposition, an explicit none-marker for an empty angle. Do **not**
name the guard here — it is a repository lint guard, per the global constraints.

- [x] **Step 6: Raise the budget**

Measure the file's new size and set its row in `scripts/check-contract-budget.sh`'s `budgets()` table
to that size plus 25%, exactly as the table's own rule states. Run
`scripts/check-contract-budget.sh` and confirm it exits 0.

---

### 4 Jira contracts — the angle label, and explain-before-filing everywhere

**Build:** green

**Files:**
- Modify: `skills/myflow-contracts/jira-integration.md`
- Modify: `skills/myflow-contracts/jira-followups.md`

**Allowed-collateral:** *(none)*

**Tests:** *(none — this task changes documentation and adds no test)*

**Regression:** Reverting this commit removes the rule that a filed finding carries its angle's
label, and lets run 1's follow-up filing ask without explaining first.

**Baseline:** before=0 after=0

**Commit:** `feat(kan-200-self-review-filing-ask-per-angle): label filed findings by angle`

**Interfaces:**
- Consumes: task 3's canonical angle table, cited rather than copied.
- Produces: the label rule and the explain-first binding for every filing site.

- [x] **Step 1: The angle label in `jira-integration.md`**

Extend **Labels on issues the pipeline creates** with one paragraph: an issue filed from a
self-review finding additionally carries the label naming the angle that produced it, and the table
of angles-to-labels lives in the finish contract's step 8 — cited, never copied, so the two cannot
drift.

- [x] **Step 2: Explain-before-filing in `jira-followups.md`**

State that a filing ask presents, per item, what was observed, what breaks and what the fix would be,
**in the message body**, before the prompt. Cite **Operator prompts**
(`skills/myflow-contracts/operator-prompts.md`) for the prompt's shape rather than restating it.

- [x] **Step 3: Confirm neither file needs a budget change**

Both have headroom at the baseline above. Run `scripts/check-contract-budget.sh`; if either row is
now over, raise that row the same way task 3 does and say so in the commit body.

---

### 5 `/myflow-finish` step 8 — the execution site

**Build:** green

**Files:**
- Modify: `skills/myflow-finish/SKILL.md`

**Allowed-collateral:** *(none)*

**Tests:** *(none — this task changes documentation and adds no test)*

**Regression:** Reverting this commit leaves the executing skill describing four angles and a
per-finding prompt, contradicting the contract it cites.

**Baseline:** before=0 after=0

**Commit:** `feat(kan-200-self-review-filing-ask-per-angle): run the five-angle filing ask`

**Interfaces:**
- Consumes: task 3's contract statement.
- Produces: the wording a run actually executes.

- [x] **Step 1: The reasoning step names five angles**

Replace the four-angle parenthetical with a citation of the contract's table plus angle 5's own
sentence. Keep the "one combined pass, never separate dispatches" rule exactly as it stands — five
angles, one pass.

- [x] **Step 2: The filing prompt**

Replace the per-finding yes/no prompt with the per-angle multi-select, and state that every finding's
explanation is printed in the message body first. Prompt shape per **Operator prompts**
(`skills/myflow-contracts/operator-prompts.md`), cited not restated.

- [x] **Step 3: The report write**

Update the report description from "the four-angle report" to the five-section shape, naming the
finding-line form. The commit shell below it is unchanged.

- [x] **Step 4: Confirm the budget**

`skills/myflow-finish/SKILL.md` has 6086 bytes of headroom at the baseline. Run
`scripts/check-contract-budget.sh`; raise its row only if the edit actually exceeds it.
