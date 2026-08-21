## Why

This project's commit convention is `<type>(<scope>): <subject>` where `<scope>` names the **module
or area inside the repository**, so a reader scanning `git log` sees what part of the codebase
moved. The myflow pipeline instructs the opposite in three places — `/myflow-do` says the dotted
task id is the scope, `pipeline.md` hard-codes the change name into the planning commit, and
`/myflow-start`'s `**Commit:**` field spec states no rule at all, so every plan written under it
declares a change-name scope. The result reaches `main` as
`feat(kan-239-run-2-asserts-base-branch-and-archives-via-pr): …` — a 50-character scope that
repeats the branch and the ticket and names no part of the codebase.

It cannot be corrected at commit time: `check-task-commit-fields.py` compares the real commit
subject against the task's declared `**Commit:**` field, so writing the correct module scope against
a plan that declares a change-name scope **fails the guard**. Following the convention currently
breaks a green build, which is what makes this actionable rather than cosmetic.

## What Changes

- **`/myflow-start`'s `**Commit:**` field spec gains the scope rule** — the scope names the module or
  area, derived from the paths in the task's own `**Files:**` field; never the change name, never a
  task id, never a list.
- **`/myflow-do`'s COMMIT-PER-TASK block drops "`<n>` is the scope."** The `Task-Id: <n>` trailer
  already identifies the task; the subject follows the plan's declared `**Commit:**` field.
- **Finish run 1's two commit messages stop carrying the change name.** The implementation commit
  derives a module scope from the reshaped diff; the planning commit becomes the fixed literal
  `chore(openspec): plan and session records`. Changed at four sites: `pipeline.md`'s Git-boundaries
  chain, `finish-contract.md`'s two-commit section, and the two `commit-split.sh` call sites in
  `myflow-finish/SKILL.md` and `myflow-do/SKILL.md`.
- **`check-task-commit-fields.py` gains a scope check** that fails a `**Commit:**` field whose scope
  is the change name, the change name's bare Jira key, or a dotted/numeric task id. A scope is
  optional — a field declaring no scope at all still passes; only a present-and-wrong scope fails.
- **A new always-on rule** `rules/commit-scope-is-the-module.mdc` carries the convention beyond this
  pipeline, with a matching row in `rules/agent-baseline.md` so it reaches dispatched agents at every
  depth.
- **A fifth rule in `check-guard-symlinks.sh`** rejects a symlink placed directly under a skill
  directory. `skills/myflow-do/SKILL.md` states that `[PRINCIPLES_PATH]` resolves beside itself, in
  `skills/myflow-do/`, and names symlinking a copy into the running command's own skill directory as
  the wrong fix — and a session created three such symlinks in `skills/myflow-fast/` about five hours
  after that sentence was committed. Rule 1 validates entries *under* `skills/*/scripts/` and reaches
  nothing at the skill directory's own top level, so prose was the only countermeasure. The three
  stray symlinks were untracked and have been removed.
- **`finish-contract.md` names `commit-split.sh`** at the point it specifies the two commits —
  KAN-202's proposed fix. Its stated symptom is stale: `/myflow-fast` already reaches the script
  through `myflow-finish/SKILL.md` §1.2, verified against the file as it stood the day the ticket
  was filed. This is defensive documentation, folded in because the file is being edited anyway.

## Capabilities

### New Capabilities

- `myflow-commit-scope`: what a commit's scope names across every message the pipeline produces —
  per-task subjects declared in `tasks.md` and the two messages finish run 1 writes — and what it
  may never name.

### Modified Capabilities

- `myflow-task-commit-fields`: the runtime guard gains a check on the **declared** `**Commit:**`
  field's scope, alongside the existing check of the field against the real commit subject.
- `myflow-contract-distribution`: a skill directory carries symlinks only under its `scripts/`
  directory; a symlink at its top level is a lint violation.

## Impact

- `skills/myflow-start/SKILL.md` — the `**Commit:**` field spec
- `skills/myflow-do/SKILL.md` — the COMMIT-PER-TASK block and the `commit-split.sh` call site
- `skills/myflow-finish/SKILL.md` — the `commit-split.sh` call site
- `skills/myflow-contracts/pipeline.md` — the Git-boundaries chain
- `skills/myflow-contracts/finish-contract.md` — the two-commit section
- `scripts/check-task-commit-fields.py` and `scripts/test-check-task-commit-fields.sh`
- `rules/commit-scope-is-the-module.mdc` (new) and `rules/agent-baseline.md`
- `scripts/check-guard-symlinks.sh` and `scripts/test-check-guard-symlinks.sh`
- No change to `setup.sh`: `always_on_rules()` discovers a rule from its frontmatter.
- Not affected: git history, and archived plans under `openspec/changes/archive/`.
