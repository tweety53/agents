# myflow-commit-scope

## Purpose

What a commit's scope names across every message the myflow pipeline produces — the
per-task subjects declared in `tasks.md`, the two messages finish run 1 writes, and the two finish
run 2 writes — and what it may never name.

## Requirements

### Requirement: A commit's scope names the module, never the change or the task

Every commit message the pipeline produces or prescribes SHALL, where it carries a Conventional
Commits scope, name the **module or area inside the repository** that the commit moved —
`finish-contract`, `gather-self-review-context`, `stages`, `openspec`, `scripts`. The scope SHALL
NOT be the change name, the change's Jira key, or a task id.

A scope is **optional**. A commit subject carrying no scope at all — `fix: <subject>` — satisfies
this requirement; only a scope that is present and names one of the prohibited things violates it.

Where a commit spans several modules, the scope SHALL name the one carrying the change's substance,
or a broader area covering them. It SHALL NOT be a list.

#### Scenario: A change-name scope is prohibited

- **WHEN** a commit subject is written as `feat(kan-239-run-2-asserts-base-branch-and-archives-via-pr): …`
- **THEN** it violates this requirement, because the scope names the change rather than a module

#### Scenario: A task-id scope is prohibited

- **WHEN** a commit subject is written as `fix(3.2): …`, where `3.2` is the task's dotted id
- **THEN** it violates this requirement, because a task id names no part of the codebase

#### Scenario: A module scope satisfies the requirement

- **WHEN** a commit touching `skills/myflow-contracts/finish-contract.md` is written as
  `feat(finish-contract): name commit-split.sh at the two-commit step`
- **THEN** it satisfies this requirement

#### Scenario: An absent scope satisfies the requirement

- **WHEN** a commit subject is written as `fix: drop an invented spec citation`, with no scope
- **THEN** it satisfies this requirement, because a scope is optional

### Requirement: A task's declared Commit subject carries a module scope

`/myflow-start`'s writing-plans stage SHALL write each task's `**Commit:**` field with a scope
derived from the paths in that same task's `**Files:**` field, per **A commit's scope names the
module, never the change or the task**.

`/myflow-do`'s COMMIT-PER-TASK instruction SHALL NOT name the task id as the scope. The task's
identity is carried by the `Task-Id: <n>` trailer, and the subject SHALL follow the plan's declared
`**Commit:**` field.

This is why the scope must be chosen at planning time: the runtime guard compares the real commit
subject against the declared field byte for byte, so a plan declaring a prohibited scope makes the
correct subject fail.

#### Scenario: The plan declares the scope the implementer must use

- **WHEN** a task's `**Files:**` field names only paths under `scripts/`
- **THEN** its `**Commit:**` field declares a scope naming that area, and the implementer's commit
  reproduces that subject exactly

#### Scenario: The task-id instruction is gone

- **WHEN** an implementer reads the COMMIT-PER-TASK block
- **THEN** it finds no instruction to use the task id as the scope, and follows the declared
  `**Commit:**` field instead

### Requirement: Finish run 1's two commits carry module scopes

The two commits `/myflow-finish` run 1 makes SHALL NOT carry the change name as their scope.

- The **implementation** commit's scope SHALL be derived from the reshaped diff — the module
  carrying the change's substance, or a broader area where it spans several.
- The **planning** commit's message SHALL be the fixed literal
  `chore(openspec): plan and session records`. It is fixed rather than derived because every
  planning commit stages the same two trees in every change, so a derived value would compute a
  constant.

### Requirement: Finish run 2's two commits carry module scopes

The two commits `/myflow-finish` run 2 makes SHALL NOT carry the change name as their scope, for the
same reason run 1's may not. Both are fixed literals, because each stages the same trees in every
change and a derived scope would therefore compute a constant.

- The **archive** commit's message SHALL be the fixed literal
  `chore(openspec): archive change, sync delta specs`. It stages the archived change dir and the
  delta-synced specs, both under `openspec/`.
- The **self-review report** commit's message SHALL be
  `docs(self-review): <name> self-review report`. It stages one file under `docs/self-review/`, so
  that directory is the area it moved; the change name appears in the subject's description, which
  names the report rather than a module and is therefore not a scope.

Naming these literals here is what keeps `skills/myflow-finish/SKILL.md`'s two commit shells from
drifting back to a change-name scope — the drift this requirement exists to close, having reached
production in KAN-244, KAN-245 and KAN-253 while the run 1 requirement above was already in force.

#### Scenario: The archive commit names openspec, not the change

- **WHEN** run 2 commits the archived change dir and the synced delta specs
- **THEN** the subject is `chore(openspec): archive change, sync delta specs`, carrying no change
  name in its scope

#### Scenario: The self-review commit names the report directory

- **WHEN** run 2 commits `docs/self-review/<name>-self-review.md`
- **THEN** the subject's scope is `self-review`, and `<name>` appears only in the description

Every site that states either message SHALL state it this way: `pipeline.md`'s Git-boundaries chain,
`finish-contract.md`'s two-commit section, and both `commit-split.sh` call sites.

#### Scenario: The planning commit is a fixed literal

- **WHEN** run 1 commits the planning artifacts and session records
- **THEN** the subject is exactly `chore(openspec): plan and session records`, with no change name
  in it

#### Scenario: The implementation commit's scope comes from the diff

- **WHEN** run 1 commits the reshaped implementation
- **THEN** its scope names the module that diff moved, not the change name

#### Scenario: Neither message is guarded

- **WHEN** run 1 writes either message
- **THEN** no guard checks the scope, because the message is derived at integration time rather than
  declared in a file a guard reads; the rule is stated in the contract and enforced by review

### Requirement: The convention is carried as an always-on rule

The convention SHALL be stated once as an always-on rule at `rules/commit-scope-is-the-module.mdc`,
declaring `alwaysApply: true` and marking a core excerpt, so `setup.sh` renders that core into every
managed `CLAUDE.md` block and installs the full text at `~/.claude/rules/`.

`rules/agent-baseline.md`'s rule table SHALL carry a matching row, so the convention reaches
dispatched agents at every depth.

`setup.sh` SHALL NOT require an edit: `always_on_rules()` already discovers a rule from its
frontmatter.

#### Scenario: A sandboxed install renders the rule

- **WHEN** `setup.sh global` runs against a sandboxed home directory
- **THEN** the managed block carries the rule's core excerpt and a pointer to
  `~/.claude/rules/commit-scope-is-the-module.md`, and that file exists

#### Scenario: A dispatched agent inherits the convention

- **WHEN** an agent reads `~/.claude/rules/agent-baseline.md`
- **THEN** the rule table names this convention with a pointer to its full text
