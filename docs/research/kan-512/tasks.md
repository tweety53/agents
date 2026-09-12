> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

# kan-512 — defer self-review to a stronger model via a saved-and-deleted context bundle

KAN-512 · implementation plan.

**Goal:** run 2's step 9 learns a third answer, `defer`: it writes the gathered bundle plus the
archived `design.md` and the change's per-invocation narrative to one committed file,
`docs/self-review/<name>-context.md`, and dispatches nothing. A new standalone
`/flow-self-review <name>` runs the five-angle pass inline on whatever model the session is on,
files and rates, writes the report, deletes the bundle and lands both on the default branch.
`docs/research/kan-512.md` (sections 2–6) is canonical for every decision a task
implements until `/flow` folds it into this change's `design.md`.

Seven tasks, dependency order, two repositories (`agents`, and its peer `gymie` for task 7's
second file). No spec, no migration.

**Baseline, measured before any edit, on `main` at `ac4cdc3`:** `test-check-self-review-report.sh`
carries 25 cases.
<!-- measured: grep -c '^# Case [0-9]' scripts/test-check-self-review-report.sh @ main ac4cdc3 -->

**Every task's verify step is the lint guards its own files need plus its own harness**, never
`scripts/run-guard-tests.sh` — that run is the last bundle's FULL SUITE paragraph and
`flow.verify`'s.

---

- [ ] 1. The report guard skips `*-context.md`

**Build:** green
**Files:** `scripts/check-self-review-report.sh`, `scripts/test-check-self-review-report.sh`
**Tests:** `test-check-self-review-report.sh` — `Case 25: a <name>-context.md beside a compliant
report is neither scanned nor an undeclared-zero coverage violation — exit 0, the bundle's
basename absent from the output`
**Regression:** reverting this commit makes the guard read the bundle as a report with none of the
five headings and fail the archive branch with an undeclared-zero coverage violation naming it.
**Baseline:** `test-check-self-review-report.sh` before=25 after=26
<!-- predicted: grep -c '^# Case [0-9]' scripts/test-check-self-review-report.sh after task 1 -->
**Commit:** `fix(scripts): skip self-review context bundles in the report guard`

  - [ ] **Step 1: The find.** `find "$TARGET" -maxdepth 1 -type f -name '*.md' -not -name
    '*-context.md' -print0`. Add one paragraph to the header comment naming why: run 2 step 9
    commits `docs/self-review/<name>-context.md` on `defer` (**Run 2 — the branch is merged**,
    `skills/flow-contracts/finish-contract-run2.md`, step 9) and `/flow-self-review` deletes it; it
    is a bundle, never a report.
  - [ ] **Step 2: Harness.** Case 25 on `new_fixture`/`compliant_report`: write
    `fixture-self-review.md` compliant and `fixture-context.md` holding the line `# Self-review
    context bundle for fixture` and nothing else; `run_guard "$FIXTURE"`; assert `RC -eq 0` and
    `$OUT` does not contain `fixture-context.md`.
  - [ ] **Step 3: Verify.**

  ```bash verified:both scripts exist on main ac4cdc3 and take no arguments
  scripts/test-check-self-review-report.sh
  scripts/check-vocabulary.sh
  ```

- [ ] 2. The contracts: `defer`, the bundle row, the deferred handoff line

**Build:** green
**Files:** `skills/flow-contracts/project-configuration.md`, `skills/flow-contracts/finish-contract-run2.md`, `skills/flow-contracts/artifacts-registry.md`, `skills/flow-contracts/handoff-blocks.md`
**Tests:** **none** — contracts carry prose; the guards in the verify step are their checks.
**Regression:** reverting this commit leaves `defer` an unrecognised `## self review` body that
step 9 reports and drops, and the bundle an artifact no registry row accounts for.
**Baseline:** `test-check-self-review-report.sh` before=26 after=26 (unchanged — this task adds no test)
<!-- predicted: no harness change in task 2 -->
**Commit:** `docs(flow-contracts): defer literal, context bundle row and deferred handoff line`

  - [ ] **Step 1: `## self review` row.** In `project-configuration.md`, the row's literals become
    `run`, `skip` or `defer`; add: `defer` saves the context bundle without asking (**Run 2 — the
    branch is merged**, `skills/flow-contracts/finish-contract-run2.md`, step 9) for
    `/flow-self-review` to consume.
  - [ ] **Step 2: Step 9 procedure.** In `finish-contract-run2.md` step 9, after the sentence
    ending "the per-run prompt is the absent case": `defer` — by key or by the prompt's third
    option — gathers as below, appends `## design.md` (the archived `design.md` verbatim) and
    `## Session narrative` (the archived `narrative.md` verbatim, or `narrative.md: absent —
    change predates the narrative rule`, then one paragraph this session writes for run 2 itself),
    writes the whole to `<project>/docs/self-review/<name>-context.md` physically under
    `<landing-worktree>`, commits it on `chore/archive-<name>` with subject `docs(self-review):
    <name> self-review context bundle`, and dispatches no reasoning pass; step 10 carries the
    bundle as it carries the report. The pass then runs in `/flow-self-review <name>`
    (`skills/flow-self-review/SKILL.md`), canonical for the deferred pass, which deletes the
    bundle in its report commit. A deferred pass covers what the bundle holds and nothing a
    same-run session could still remember beyond it — the report's `**Deferred:**` line states that.
  - [ ] **Step 3: Registry row.** In `artifacts-registry.md`'s table, after the **Panel record**
    row: `| Self-review context bundle | run 2 step 9, on \`defer\` |
    \`<project>/docs/self-review/<name>-context.md\`, committed on \`chore/archive-<name>\` |
    \`/flow-self-review <name>\`, in the same commit as the report |`.
  - [ ] **Step 4: Handoff line.** In `handoff-blocks.md`, the `Self-review` field's value set
    becomes `<path> (rating: <n>/5) | deferred — docs/self-review/<name>-context.md | skipped |
    skipped — project default`.
  - [ ] **Step 5: Verify.**

  ```bash verified:the three guards are .flow/project.md's ## lint rows for markdown
  scripts/check-vocabulary.sh
  scripts/check-references.sh
  scripts/check-contract-budget.sh
  ```

- [ ] 3. The per-invocation narrative

**Build:** green
**Files:** `skills/flow/verify-and-handoff.md`, `skills/flow/integrate.md`
**Tests:** **none** — stage files carry procedure; the guards in the verify step are their checks.
**Regression:** reverting this commit leaves no run writing `narrative.md`, so every deferred
bundle carries the `absent` marker and the ticket's caveat stays unaddressed.
**Baseline:** `test-check-self-review-report.sh` before=26 after=26 (unchanged — this task adds no test)
<!-- predicted: no harness change in task 3 -->
**Commit:** `feat(flow): append a per-invocation narrative to the change directory`

  - [ ] **Step 1: `flow.write-in-progress`.** In `verify-and-handoff.md`, before "Write the state
    file": append to `<abs-worktree>/spectre/changes/<name>/narrative.md` (create it with the
    title `# <name> — session narrative` when absent) one section `## <YYYY-MM-DD> — <creating
    run | fix run>` holding this session's own prose account of the run — problems hit,
    workarounds, time sinks, environment gaps, operator decisions taken mid-run — and nothing the
    ledger or panel record already holds. It rides the next planning commit through
    `commit-split.sh`'s existing `spectre/changes/` pathspec; nothing else stages it.
  - [ ] **Step 2: `flow.preserve-sessions`.** In `integrate.md`, before "Before any route commits,
    reshape the branch": the same append, heading `## <YYYY-MM-DD> — integrate run`, covering
    preflight, the unfinished-work gate and the rebase.
  - [ ] **Step 3: Verify.**

  ```bash verified:the three guards are .flow/project.md's ## lint rows for markdown
  scripts/check-vocabulary.sh
  scripts/check-references.sh
  scripts/check-contract-budget.sh
  ```

- [ ] 4. Step 9 defers

**Build:** green
**Files:** `skills/flow/archive.md`
**Tests:** **none** — a stage file carries procedure; the guards in the verify step are its checks.
**Regression:** reverting this commit leaves step 9 with two answers, so a `defer` key is dropped
as unrecognised and the bundle is never written.
**Baseline:** `test-check-self-review-report.sh` before=26 after=26 (unchanged — this task adds no test)
<!-- predicted: no harness change in task 4 -->
**Commit:** `feat(flow): defer self-review by saving a committed context bundle`

  - [ ] **Step 1: Key resolution.** The `project-get.sh <main-checkout> "self review"` sentence
    matches three literals; `defer` proceeds to the bundle write below with no prompt.
  - [ ] **Step 2: The prompt.** Insert `- **Defer — save the bundle for \`/flow-self-review\`**`
    between the Yes and No options; "On anything but No" becomes "On Yes".
  - [ ] **Step 3: The bundle write.** New paragraph before "On Yes": on `defer`, run the script
    exactly as above and write its stdout, then `## design.md` and `## Session narrative` per the
    contract (task 2 step 2), to `<landing-worktree>/docs/self-review/<name>-context.md`; commit
    with the report's own branch-assert shell, path and subject swapped:

    ```bash verified:the shape is archive.md step 9's existing report-commit shell with the path and subject swapped
    [ "$(git -C <landing-worktree> branch --show-current)" = "chore/archive-<name>" ] \
      && git -C <landing-worktree> add -- docs/self-review/<name>-context.md \
      && { git -C <landing-worktree> diff --cached --quiet \
           || git -C <landing-worktree> commit -m "docs(self-review): <name> self-review context bundle"; }
    ```

    Then straight to the `flow stage end … flow.self-review -outcome completed` mark; the
    handoff's `Self-review` line reads `deferred — docs/self-review/<name>-context.md`.
  - [ ] **Step 4: Handoff template.** The `**Self-review:**` line in the run-2 block gains the
    `deferred — …` alternative from task 2 step 4.
  - [ ] **Step 5: Verify.**

  ```bash verified:the four guards are .flow/project.md's ## lint rows for markdown and stage marks
  scripts/check-vocabulary.sh
  scripts/check-references.sh
  scripts/check-contract-budget.sh
  scripts/check-stage-mark-calls.sh
  ```

- [ ] 5. `/flow-self-review <name>`

**Build:** green
**Files:** `skills/flow-self-review/SKILL.md`, `commands/flow-self-review.md`, `commands-claude/flow-self-review.md`
**Tests:** **none** — a skill and its stubs carry procedure; the guards in the verify step are
their checks.
**Regression:** reverting this commit leaves every deferred bundle with no command that consumes
it, so `defer` becomes `skip` with a tracked leftover.
**Baseline:** `test-check-self-review-report.sh` before=26 after=26 (unchanged — this task adds no test)
<!-- predicted: no harness change in task 5 -->
**Commit:** `feat(flow-self-review): run a deferred self-review pass from a saved bundle`

  - [ ] **Step 1: Skill frontmatter.** `name: flow-self-review`; description "Run the self-review
    pass a `/flow` run deferred, inline on this session's model, from the saved context bundle;
    file, rate, write the report, delete the bundle. Standalone, not a pipeline stage. Use for
    /flow-self-review."; `allowed-tools: Bash(git:*), Bash(scripts/check-self-review-report.sh:*)`;
    the `license`/`compatibility`/`metadata` block as `skills/flow-settings/SKILL.md` carries it.
    Announce "Using flow-self-review."
  - [ ] **Step 2: Procedure.** Sections 1–6 exactly as section 5 of
    `docs/research/kan-512.md` lists them: resolve and refuse (bundle absent → list
    `docs/self-review/*-context.md` or `none pending`, stop; not on `<default-branch>` → print
    the branch, stop); the five-angle table with `flow-` labels cited from
    `skills/flow-contracts/finish-contract-run2.md` step 9, run inline, every angle explicit;
    explain-then-ask with the one `AskUserQuestion` shape and overflow rule cited from that same
    step; Jira filing per **Labels on issues the pipeline creates**
    (`skills/flow-contracts/jira-integration.md`) with the skipped-line fallback; the report in
    the shape `check-self-review-report.sh` checks plus `**Deferred:** reasoning pass run on
    <model> from <bundle path>` under the title; delete, guard, one commit
    `docs(self-review): <name> self-review report`, `pull --rebase`, push, rejected push left local
    and named; end naming the report path, rating and filed keys.
  - [ ] **Step 3: Stubs.** `commands/flow-self-review.md` with the `name`/`id`/`category`/
    `description` frontmatter and `commands-claude/flow-self-review.md` with `model: opus` and the
    description, both delegating to the **flow-self-review** skill by name as the `flow-settings`
    stubs do; **Input:** one change name, required; **Model:** whatever the session runs on —
    that is the stronger-model choice, so pick it with `/model` before invoking.
  - [ ] **Step 4: Verify.**

  ```bash verified:the three guards are .flow/project.md's ## lint rows for markdown
  scripts/check-vocabulary.sh
  scripts/check-references.sh
  scripts/check-installed-citations.sh
  ```

- [ ] 6. Guard declarations and the skill index

**Build:** green
**Files:** `scripts/check-contract-budget.sh`, `scripts/check-references.sh`, `scripts/check-installed-citations.sh`, `scripts/check-guard-symlinks.sh`, `scripts/check-stage-mark-calls.sh`, `CLAUDE.md`, `AGENTS.md`, `README.md`
**Tests:** **none** — declaration tables and index rows; the guards in the verify step are their checks.
**Regression:** reverting this commit makes every enumerating guard fail on the three undeclared
files task 5 added.
**Baseline:** `test-check-self-review-report.sh` before=26 after=26 (unchanged — this task adds no test)
<!-- predicted: no harness change in task 6 -->
**Commit:** `chore(scripts): declare the flow-self-review skill and stubs to every enumerating guard`

  - [ ] **Step 1: Budgets.** Three `budgets()` rows — `commands-claude/flow-self-review.md`,
    `commands/flow-self-review.md`, `skills/flow-self-review/SKILL.md` — each the file's size
    after task 5 plus 25%, per the ratchet rule in `.flow/project.md`.
  - [ ] **Step 2: Expected-zero and declared entries.** `check-references.sh`
    `EXPECTED_ZERO_COMMAND_DISPATCH_STUBS` gains both stubs; `check-installed-citations.sh` gains
    a `declare_if_present` pair mirroring the `flow-settings` pair; `check-guard-symlinks.sh` gains
    `declare_if_present "flow-self-review" "invokes no guard of its own — runs the project's
    report guard by path, with no implementation or verification stage"`;
    `check-stage-mark-calls.sh` gains `skills/flow-self-review/SKILL.md` and a reason mirroring
    `flow-settings`'.
  - [ ] **Step 3: Index.** One skill-index row after `flow-settings` in `CLAUDE.md` and
    `AGENTS.md`; the `commands/` tree line and a `flow-self-review/` line in `README.md`.
  - [ ] **Step 4: Verify.**

  ```bash verified:each guard is a .flow/project.md ## lint row and runs without arguments
  scripts/check-contract-budget.sh
  scripts/check-references.sh
  scripts/check-installed-citations.sh
  scripts/check-guard-symlinks.sh
  scripts/check-stage-mark-calls.sh
  scripts/check-vocabulary.sh
  ```

- [ ] 7. Project configs default to `defer`

**Build:** green
**Files:** `.flow/project.md`, `../gymie/.flow/project.md`
**Tests:** **none** — two one-word config edits; the guard in the verify step is their check.
**Regression:** reverting this commit returns both projects to `skip`, so no run ever writes a
bundle and `/flow-self-review` has nothing to consume.
**Baseline:** `test-check-self-review-report.sh` before=26 after=26 (unchanged — this task adds no test)
<!-- predicted: no harness change in task 7 -->
**Commit:** `chore(flow): defer self-review by default`

  - [ ] **Step 1: agents.** `.flow/project.md` `## self review` body: `` `skip` `` → `` `defer` ``.
  - [ ] **Step 2: gymie.** In the gymie peer worktree, its own `.flow/project.md` `## self review`
    body: `skip` → `defer` (unbackticked, as it is today).
  - [ ] **Step 3: Verify.**

  ```bash verified:check-vocabulary.sh is a .flow/project.md ## lint row and scans .flow/
  scripts/check-vocabulary.sh
  ```
