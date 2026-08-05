# Tasks — kan-13-myflow-planning-and-status-fixes

> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Repository:** `/Users/tweety53/Projects/agents` — a documentation-and-guards repository. Most
tasks below edit Markdown contracts/skills; one group adds a new Bash+Python script plus its test
harness, mirroring `scripts/check-plan-provenance.py`/`.sh`/`test-check-plan-provenance.sh`. There
is no compiled artifact and no application to run. Verification is the repository's own lint guards
declared in `.myflow/project.md` (`check-vocabulary.sh`, `check-references.sh`,
`check-plan-provenance.sh`), the new script's own test harness, and targeted `grep` assertions.

**Order:** the new guard and its test (group 1) have no dependency on the contract/skill edits and
are built first, so group 2 can cite the guard's real invocation rather than guessing it. Group 2
(the new contract file) defines the tag vocabulary group 3 (`/myflow-start`'s writing-plans stage)
requires plans to use. Group 4 (the shared change-name-resolution fix) and group 5
(`.myflow/project.md` bookkeeping) are independent of groups 1-3 and of each other. Group 6
(cross-file verification) runs last because it depends on every prior group's edits existing.
**Every task below leaves the repository in a runnable state for the guards named above** — there
is no compiled build to break, and no task depends on a file another task has not yet created,
so no task in this plan requires a `**Build:** red` tag.

**Global constraints, carried into every dispatch:**

- **No suppression markers, ever.** Fix the offending line; never silence a guard.
- **Citation over copy.** Where a rule already lives in `pipeline.md`, `plan-provenance.md` or
  `state-file.md`, cite it with a bold token plus its backticked path rather than restating it, so
  `scripts/check-references.sh` checks the pointer — every new bold-token/path pair must have a
  matching heading in the target file, or the guard fails.
- Every fenced code block this plan or its implementation adds to a change's `tasks.md`,
  `design.md` or `proposal.md` carries `verified:`/`unverified:` on its info string, and every
  numeric claim carries a `measured:`/`predicted:` comment within two lines, per
  `skills/myflow-contracts/plan-provenance.md` — the guard that enforces this on `agents`' own
  changes is `scripts/check-plan-provenance.sh`, already declared in `.myflow/project.md`.
- Every path written into a contract, skill or handoff is **absolute** where it names something the
  operator opens, exactly as the rest of this pipeline's output already is.
- This change touches only the `agents` repo. The `gymie`/`gymie-frontend` items on KAN-13 are
  separate changes and are out of scope here.

---

## 1. `scripts/check-task-build-green.py`, its wrapper, and its test

### 1.1 Write the parser/checker

- [x] Create `scripts/check-task-build-green.py` (Python 3, standard library only — no third-party
      imports, no pip, no network — matching `check-plan-provenance.py`'s constraint stated in
      `.myflow/project.md`'s `## apps` section).
- [x] Task-boundary detection: a line matching `^### (\d+(?:\.\d+)*)\s` (a level-3 heading whose
      text starts with a dotted task id, e.g. `### 2.1 Write the parser`) opens a new task. The
      task's id is the captured group.
- [x] Build-tag detection: within a task's body (from its heading up to the next `^###`/`^##`
      heading or end of file), the **first** line matching
      `^\*\*Build:\*\*\s+(green|red\s+—\s+merges with Task ([\d.,\s]+))\s*$` is that task's tag. A
      task body with no matching line has **no tag**.
- [x] Validation, run over every task found in the file, in document order:
      1. **Missing tag** — a task with no `**Build:**` line → violation
         `<file>:<line of heading>: task <id> has no **Build:** tag`.
      2. **Red without a partner** — a task tagged `red` whose `merges with Task …` clause parses
         to zero task ids → violation `<file>:<line of tag>: task <id> is red with no merge
         partner named`.
      3. **Partner does not exist** — a named partner id that is not any task's id in this file →
         violation `<file>:<line of tag>: task <id> merges with Task <partner>, which does not
         exist in this plan`.
      4. **Partner is not green** — a named partner whose own tag is `red` (whether or not *it*
         resolves further) → violation `<file>:<line of tag>: task <id> merges with Task
         <partner>, which is itself red`. (A red→red chain is caught at both ends; each reported
         once, from the task carrying the offending tag.)
      5. **Unreferenced red is not itself a violation** — a `red` task whose partner exists and is
         `green` passes: the guard does not require the partner to name it back. (The relationship
         is directional: the merge is a fact about the red task's dispatch, not a mutual
         cross-reference.)
- [x] Exit 0 with no output when the file has zero violations (including when it has zero tasks —
      an empty or task-free plan is not this guard's concern). Exit 1 and print every violation,
      one per line in `file:line: message` shape, when there is at least one. Exit 2 on an
      invocation error (bad argv, unreadable file).
- [x] CLI: `check-task-build-green.py <path-to-tasks.md>` — a single required positional argument,
      no flags. (Unlike `check-plan-provenance.py`, this guard's scope is one file per call, not a
      whole-repository scan — the caller, `check-task-build-green.sh`, resolves which `tasks.md`
      to pass. See 1.2.)
- [x] Module docstring states the contract this script implements (mirrors
      `check-plan-provenance.py`'s own docstring convention) and cites
      **The build-green tag and its guard** (`skills/myflow-contracts/build-green.md`) as
      canonical for the rule — the script's comment does not restate the rule, only implements it.

**Build:** green

### 1.2 Write the thin Bash wrapper

- [x] Create `scripts/check-task-build-green.sh`, structured like
      `scripts/check-plan-provenance.sh`: probes for a working `python3` (not just present on
      `PATH` — run a trivial program, exit 2 with a clear message on failure, exactly as the
      existing wrapper does), then execs the Python script.
- [x] With no arguments, scan every non-archived change's `tasks.md` under
      `openspec/changes/*/tasks.md` (excluding `openspec/changes/archive/`), calling the Python
      script once per file and aggregating exit codes (non-zero if any file has a violation).
      With one argument, treat it as an explicit `tasks.md` path and scan only that file — this is
      what `/myflow-start` calls before publishing a specific change's proposal artifact (see task
      3.2).
- [x] `.myflow/project.md`'s `## lint` row must be able to name this script with no arguments and
      have it do the right thing repo-wide, matching how `check-plan-provenance.sh` is declared
      there today.

**Build:** green

### 1.3 Write the test harness

- [x] Create `scripts/test-check-task-build-green.sh`, structured like
      `scripts/test-check-plan-provenance.sh`: builds fixture `tasks.md` files under a sandboxed
      `mktemp -d` tree, runs the guard against each via
      `scripts/check-task-build-green.sh <fixture-path>`, and asserts exit code plus the presence
      of the expected violation message.
- [x] Fixture cases, at minimum:
      1. All tasks tagged `green` → exit 0, no output.
      2. One task with no `**Build:**` line → exit 1, message names that task's id.
      3. A task tagged `red` with no `merges with Task …` clause → exit 1.
      4. A task tagged `red — merges with Task 9.9` where no task `9.9` exists → exit 1.
      5. A task tagged `red — merges with Task 2` where task `2` is itself tagged `red` → exit 1,
         reported against the first task.
      6. A task tagged `red — merges with Task 2` where task `2` is tagged `green` → exit 0.
      7. A malformed tag line (e.g. `**Build:** yellow`) is treated as **no tag** (falls through to
         case 2's violation, not a separate parse error) — the regex in 1.1 simply does not match.
- [x] Run the new test. All fixtures must pass before this task is marked done.

```bash verified:this file is executable and self-contained per the existing check-plan-provenance test pattern it mirrors
scripts/test-check-task-build-green.sh
```

**Build:** green

---

## 2. `skills/myflow-contracts/build-green.md`

### 2.1 Write the contract file

- [x] Create `skills/myflow-contracts/build-green.md`, modeled structurally on
      `skills/myflow-contracts/plan-provenance.md` (opening statement that it is canonical; a
      section defining the tag vocabulary; a section on the guard's scope and what it does and
      does not verify).
- [x] State the tag vocabulary exactly as specified in
      **Requirement: Every task in a plan declares its build state**
      (`openspec/changes/kan-13-myflow-planning-and-status-fixes/specs/myflow-build-green/spec.md`):
      `**Build:** green` and `**Build:** red — merges with Task <N>`, where the tag is the first
      matching line in a task's body (this is the file `check-task-build-green.py`'s docstring
      cites — task 1.1's last checklist item — so the two must describe the same placement rule).
- [x] State the guard's scope and limit exactly as
      **Requirement: A guard enforces the build-state tags before the proposal publishes**
      (same spec file) states it: the guard verifies a tag is present and consistent, never that a
      `green` tag is true — same accepted limit `plan-provenance.md`'s own "What the guard does
      not do" section states for its tags, and this section should say so by citing that section
      rather than re-deriving the reasoning.
- [x] Include a heading whose exact text a future cross-reference can point at — this file will be
      cited from `skills/myflow-start/SKILL.md` (task 3.1) via a bold-token/path pair, so
      `scripts/check-references.sh` requires a matching heading here. Name it plainly, e.g.
      `## The build-green tag` and `## The guard's scope`.

**Build:** green

### 2.2 Add it to the `myflow-contracts` index

- [x] Read `skills/myflow-contracts/SKILL.md`'s index table (the one listing `pipeline.md`,
      `state-file.md`, `state-self-heal.md`, `project-configuration.md`, `jira-integration.md`,
      `plan-provenance.md` with one "load it when you need to" row each).
- [x] Add a row for `build-green.md`: *Load it when you need to* — "write or check a plan's
      build-state tags: the tag vocabulary, the merge-partner rule, and the guard's scope" (adapt
      wording to match the table's existing voice; do not deviate from the row shape the other
      five entries use).

**Build:** green

---

## 3. `/myflow-start` writing-plans-stage integration

### 3.1 Require the tag in section D

- [x] In `skills/myflow-start/SKILL.md` section **D. Basic Workflow #3 — Writing plans**, add an
      instruction alongside the existing plan-provenance tagging paragraph: while enriching
      `tasks.md`, tag every task with `**Build:**` per
      **The build-green tag** (`skills/myflow-contracts/build-green.md`) — cite, do not restate,
      exactly as the existing plan-provenance sentence cites `skills/myflow-contracts/plan-provenance.md`.
- [x] Update the existing "Before publishing the artifact, run the project's configured
      plan-provenance guard if it declares one, and fix any hit" sentence to also name the
      build-green guard, in the same "if it declares one, fix any hit" shape — both guards are
      optional-by-presence in a project, exactly as documented in
      **Requirement: A guard enforces the build-state tags before the proposal publishes**
      (`openspec/changes/kan-13-myflow-planning-and-status-fixes/specs/myflow-build-green/spec.md`).

**Build:** green

### 3.2 Add the guardrail line

- [x] In `skills/myflow-start/SKILL.md`'s **Guardrails** section, add a line alongside "Never
      publish a plan carrying an untagged block or an unsourced number": **Never** publish a plan
      carrying a task with no `**Build:**` tag, or a `red` tag with no resolvable green partner.
- [x] Confirm (by re-reading the edited section) that the new sentence sits next to the existing
      provenance guardrail rather than in an unrelated part of the list, so a future reader sees
      both publish-blocking checks together.

**Build:** green

---

## 4. Shared change-name-resolution fix

### 4.1 Rewrite `pipeline.md`'s `## Change name resolution` section

- [x] In `skills/myflow-contracts/pipeline.md`, locate `## Change name resolution (all /myflow-*
      commands)` (currently: run `openspec list --json`, filter to non-archived, exactly-one/many/
      zero matches).
- [x] Replace the enumeration step with the union rule from
      **Requirement: Change name resolution enumerates every state-tracked change**
      (`openspec/changes/kan-13-myflow-planning-and-status-fixes/specs/myflow-command-surface/spec.md`):
      the set of names is `openspec list --json`'s non-archived names **union** the basenames
      (minus `.json`) of every file directly under the project's state directory
      (`/Users/tweety53/Agents/myflow/state/<project-key>/*.json`, `<project-key>` resolved exactly
      as `skills/myflow-contracts/state-file.md` already defines it — cite that file for the
      formula rather than re-deriving it), then drop any name whose
      `openspec/changes/<name>/` directory has reached `openspec/changes/archive/`.
- [x] State the unreadable-file rule: a state-directory file that cannot be parsed is **reported**
      (named) and skipped from the union, never silently absent — matches the design decision
      recorded in `design.md`'s Risks / Trade-offs.
- [x] Keep the rest of the section's shape (exactly-one/many/zero-match handling; the Jira-derived
      naming pointer at the end) unchanged — this task changes only how the candidate-name set is
      built, not what happens once it exists.

**Build:** green

### 4.2 Update `/myflow-status`'s own step 1 to cite the shared section

- [x] In `skills/myflow-status/SKILL.md`, step **1. List open changes** currently runs
      `openspec list --json` directly. Replace its enumeration with a citation to
      **Change name resolution** (`skills/myflow-contracts/pipeline.md`) — the same shared logic
      every other command now uses — rather than keeping a second, independent `openspec list
      --json` call that could drift from it.
- [x] Leave the rest of step 1 (the `<name>` argument behavior, the "zero open changes" handling)
      unchanged; only the enumeration source changes.

**Build:** green

---

## 5. `.myflow/project.md` bookkeeping

### 5.1 Register the new guard and its test

- [x] In `.myflow/project.md`'s `## lint` section, add `scripts/check-task-build-green.sh` to the
      list alongside `check-vocabulary.sh`, `check-references.sh`, `check-plan-provenance.sh`.
- [x] In `.myflow/project.md`'s `## test` section, add `scripts/test-check-task-build-green.sh` to
      the list alongside the existing nine test scripts.

**Build:** green

---

## 6. Cross-file verification

### 6.1 Run the repository's own guards

- [x] Run every lint guard `.myflow/project.md` now declares, and fix any hit before this task is
      marked done (per this repository's global Lint Fix Priority rule — no suppressions, no
      config weakening).

```bash unverified:run against the working tree after every prior task in this plan is complete
scripts/check-vocabulary.sh
scripts/check-references.sh
scripts/check-plan-provenance.sh
scripts/check-task-build-green.sh
```

- [x] Run every test script `.myflow/project.md` now declares, including the new one added in
      task 1.3.

```bash unverified:run against the working tree after every prior task in this plan is complete
scripts/test-setup.sh
scripts/test-check-references.sh
scripts/test-check-plan-provenance.sh
scripts/test-check-finish-preflight.sh
scripts/test-preserve-session-records.sh
scripts/test-check-unfinished-work.sh
scripts/test-check-cleanup-complete.sh
scripts/test-gather-self-review-context.sh
scripts/test-uncommitted-review-package.sh
scripts/test-check-task-build-green.sh
```

**Build:** green

### 6.2 Confirm the cross-references resolve

- [x] Grep for every new bold-token/path pair this plan introduced (task 2.1's citation into
      `plan-provenance.md`, task 3.1's citation into `build-green.md`, task 4.1's citation into
      `state-file.md`) and confirm each target file carries a heading matching the cited bold
      token, so `check-references.sh` (already run in 6.1) has something real to check rather than
      passing vacuously because no such pair exists yet.
- [x] Confirm `openspec validate kan-13-myflow-planning-and-status-fixes --strict` still passes
      after every file edit in this plan (it passed at proposal/design/specs time; this task
      re-confirms nothing broke it along the way).

```bash unverified:run once every prior task's edits are in place
openspec validate kan-13-myflow-planning-and-status-fixes --strict
```

**Build:** green
