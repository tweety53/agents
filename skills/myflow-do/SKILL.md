---
name: myflow-do
description: Implement an OpenSpec change with Superpowers TDD and a multi-agent review panel, emit the manual test guide, and stage everything for one human gate. Re-run to apply a fix. Use for /myflow-do.
allowed-tools: Bash(openspec:*)
license: MIT
---

Implement an OpenSpec change, write its manual test guide, and stage both for the human gate at
`IN_PROGRESS`. **No commits** — unless a PR already exists, which is the one exception below.

**Announce at start:** "Using myflow-do for change `<name>`."

**Load `skills/myflow-contracts/pipeline.md` first** — it is canonical for the states, the
command→state transition table, git boundaries, and the handoff output shape.

## State gate

Accepts **`STARTED`** (first run) or **`IN_PROGRESS`** (fix run).

- From `STARTED`: create the worktree, implement the plan, end at `IN_PROGRESS`.
- From `IN_PROGRESS`: resume the **existing** worktree, apply a fix, and **write the state back
  unchanged**. There is no `originStage` and no fix re-entry table — a fix never moves the state. <!-- vocab-guard:allow -->

At `FINISHED` the change is archived; emit the wrong-state handoff and stop.

## Superpowers Basic Workflow

| Step | Skill | When |
|------|-------|------|
| **2** | **superpowers:using-git-worktrees** | Before the first code change, on a first run |
| **3** | **superpowers:writing-plans** | Validate the plan; repair `tasks.md` if it is not apply-ready |
| **4** | **superpowers:subagent-driven-development** | Execute the remaining tasks |
| **5** | **superpowers:test-driven-development** | Every implementer dispatch, every task |
| **6** | **superpowers:requesting-code-review** + the review panel | Per-task review, then the final whole-branch panel |
| **8** | **superpowers:verification-before-completion** | Evidence before claiming done |

**Never** invoke `finishing-a-development-branch` — integration is `/myflow-finish`'s job.

## 1. Load context and validate the plan

```bash
openspec status --change "<name>" --json
openspec instructions apply --change "<name>" --json
```

- `state: "blocked"` → stop; suggest `/myflow-start <name>`.
- `state: "all_done"` → suggest `/myflow-finish <name>`.
- Read every path in `contextFiles`. Resolve paths from the CLI JSON, never from an assumed layout.

Confirm `tasks.md` meets **writing-plans** quality — exact paths, verification commands,
bite-sized steps, no placeholders. If it does not, invoke **superpowers:writing-plans** to repair
it before touching code.

Extract the **Global constraints** verbatim from the delta specs and `design.md` for the reviewers.

## 2. Isolate the workspace (first run only)

Invoke **superpowers:using-git-worktrees**. Branch `openspec/<name>`. Never implement on the
default branch without explicit consent. Record each worktree's merge base — it goes into the state
file's `worktrees` map, which is the authoritative list of affected worktrees.

On a fix run, resume the existing worktree. **Never create a second one.**

## 3. Documenting a fix, before implementing it

On a fix run, record what changed **before** writing code, so the proposal never goes stale: either
append to `proposal.md`/`tasks.md`, or create a linked nested `<name>-fix-N` sub-change. Ask which.
A nested sub-change is never archived alone — it goes with its parent.

If the fix adds scope the linked Jira issue does not describe, sync the issue **description** per
**Description sync** in **Jira integration** (`skills/myflow-contracts/jira-integration.md`). Never
transition the issue here.

## 4. Execute (SDD + TDD)

Invoke **superpowers:subagent-driven-development**, treating each remaining checkbox (or a tightly
coupled group) as one task. Every implementer dispatch **must** carry all three of:

> **MYFLOW — NO COMMITS:** Do **not** run `git commit`, `git push`, merge, or open a PR. Leave all
> changes uncommitted in the worktree. You **may** `git add`. The parent records
> `TASK_BASE=$(git rev-parse HEAD)` before dispatch; your diff for review is `git diff TASK_BASE`.

> **REQUIRED SUB-SKILL:** Use superpowers:test-driven-development — RED-GREEN-REFACTOR for this
> task. Delete any code written before its test.

> **REQUIRED READING:** [engineering-principles.md](engineering-principles.md) — your
> implementation must satisfy these principles; the panel's principles reviewer checks the diff
> against them.

Per-task review without commits: write `git diff TASK_BASE > .superpowers/sdd/task-N.diff` and give
the reviewer that path, never a commit range. Ledger line: `Task N: complete (uncommitted, review
clean)`. Mark a checkbox `[x]` only after its task passes spec **and** quality review.

On BLOCKED: pause and report. Never guess.

## 5. The review panel

Write `.superpowers/sdd/final-review.diff` from `git diff <merge-base>` (staged and unstaged), then
dispatch **separate** review subagents — one per selected slot, in **every** affected worktree.
Never merge two slots into one prompt.

**Every slot runs on Sonnet.** There is no parent-model inheritance and no economy tier — the
panel's cost must not depend on which model the operator happens to be running.

| # | Slot | Required? | Model | How to spawn |
|---|------|-----------|-------|--------------|
| 0 | **Primary** — plan alignment + code quality | **always** | `sonnet` | **superpowers:requesting-code-review** with `final-review.diff` + the plan/spec constraints |
| 1 | **Bugbot** — defect hunt | **always** | its own | `subagent_type: bugbot`, `Diff: uncommitted changes`, `Full Repository Path: <worktree>` |
| 2 | **Principles** | **always** | `sonnet` | general-purpose + [principles-reviewer-prompt.md](principles-reviewer-prompt.md), `[LENS]` = **Merged** |
| 3 | **Security** | conditional | its own | `subagent_type: security-review`, same shape as Bugbot |
| 4 | **Adversarial** | conditional | `sonnet` | general-purpose + [adversarial-reviewer-prompt.md](adversarial-reviewer-prompt.md) |
| 5+ | **Principles lens B / lens C** | conditional | `sonnet` | same template, `[LENS]` = **Lens B — simplicity & state** or **Lens C — robustness & ops** |

Slots 1 and 3 are dispatched by `subagent_type` and carry their own agent definitions — pass them
no model override. Every other slot names Sonnet explicitly.

Slot 2 is the panel's only mandatory judgment check on *how* the code is built. It reads
`engineering-principles.md` — never a pasted copy — and owns the project's **hard invariants** from
its standards files: architecture and layer purity, new suppressions, weakened lint config.

**Resolve `[PRINCIPLES_PATH]` before dispatching any principles slot.** It is the **absolute** path
of `engineering-principles.md` in the directory you are reading this file from — under a global
install, `~/.claude/skills/myflow-do/engineering-principles.md`. The subagent's working directory is
the project worktree, which has no `skills/` tree, so a repo-relative path opens nothing and the
reviewer runs with no principle list. Confirm the file exists before spawning; if it does not,
stop and report rather than dispatching a blind reviewer.

**Resolve `[STANDARDS_PATHS]` before dispatching slot 2**, from the `## standards` entries in the
project's `.myflow/project.md`. Entries are **not** paths to use as-is: each resolves through the
entry-form table and the containment rule in **Project configuration**
(`skills/myflow-contracts/project-configuration.md`), and an entry failing either is reported by
name and dropped. Resolve that contract file by **absolute** path too, for the same reason as
above; if it is not readable, **stop** — do not resolve entries without the containment rule, which
is the only thing between an attacker-editable list in a tracked file and an arbitrary file read
whose output lands in a committed review record. Pass an **empty** value when none resolve; that
correctly empties the Hard Invariants section rather than substituting another project's standards.
Record which standards files were passed, or that none resolved.

**No two principle reviewers may share a lens.**

### Optional slot selection

Evaluate against `final-review.diff` **before** dispatching, and record which optional slots were
included and which were excluded and why.

| Slot | Include when the diff touches | Ask when |
|------|-------------------------------|----------|
| 3 — Security | auth/authz, tokens, crypto, secrets or config, query construction, path or file handling, deserialization, CORS/HTTP edge, new dependencies | a config or dependency file changed, but only comments or a version bump |
| 4 — Adversarial | migrations, concurrency/scheduling, behavior changes to code with existing tests, any test modified or deleted, or **>~300** changed lines | **150–300** changed lines with no other trigger |
| 5 — Lens B (simplicity & state) | **>~200** changed lines, or **≥3** new classes/modules | — |
| 5 — Lens C (robustness & ops) | error handling, retries, schedulers, external integrations, config/env, migrations | — |

**Borderline → ask**, with **include** as the default. A reviewer too many costs tokens; one too
few costs a defect.

A documentation-, prompt-, or test-only diff with no trigger runs the three required slots alone.
That is a correct outcome, not a skipped review — say so explicitly.

### Panel re-runs

**Pass 1 always runs the full roster selected for this change.** Only re-runs after a fix are
scoped. Record `FIX_BASE` before each fix, then `git diff FIX_BASE > .superpowers/sdd/fix-round-N.diff`.

| Mode | Who re-runs | Diff they get |
|------|-------------|---------------|
| **Targeted** (default) | Slot 0 (always, as integration check) + every agent that raised a finding | `fix-round-N.diff` |
| **Full** (escalation) | Every slot in this run's roster | rewritten `final-review.diff` |

**Escalate automatically** — do not ask, and say why in the record — when the fix touched a file
outside the set named in the findings; the fix diff exceeds ~150 changed lines; the fix altered a
delta spec, a migration, or a public contract; a targeted re-run surfaced a **new** Critical
finding; or three or more fix rounds have already run.

Targeting is a cost optimization, never a coverage waiver: a targeted re-run is never fewer than
two agents, and handoff still requires **zero** open Critical/Important findings from every agent
that has run, with the final pass showing a non-stale clean result for every slot in the roster.

Union all Critical/Important findings, dedupe by file:line + theme, and give **one** fix subagent
the combined list. Record every pass in `.superpowers/sdd/final-review-panel.md`: mode, which
agents ran, why, and the diff path they read.

## 6. Write the manual test guide

In the same run, write or refresh `docs/manual-test/<name>.md` — how to run every app in scope, and
a functionality checklist derived from the delta specs. This is why reviewing and testing are one
gate: both surfaces are produced together and can never drift apart.

- **Every path is absolute**, resolved from `git worktree list` or the state file's `worktrees`
  keys. Never a relative sibling path (`../<other-app>`), and never a main-checkout path while a
  worktree holds the work.
- Apps in scope come from `## apps` in the project's `.myflow/project.md`, or from auto-detection
  when that file or key is absent — see **Project configuration**.
- On a fix run, **refresh** the guide: preserve already-ticked boxes, and re-open only what the fix
  invalidated.
- There is no skip prompt and no `SKIPPED` marking. The guide is there to use or ignore; nothing
  records whether it was used.

## 7. Verify, stage, and hand off

Run the project's `## lint` and `## test` commands from `.myflow/project.md` (auto-detect if
absent) and show the output. **Nothing runs them later** — `/myflow-finish` has no verification
gate — so a non-zero exit blocks this handoff.

Confirm every intended checkbox is `[x]`, and that no commits were made:
`git log <merge-base>..HEAD` must be empty, unless a PR already exists (below).

In **every** affected worktree:

```bash
git -C <worktree> add -A
git -C <worktree> status
git -C <worktree> diff --cached --stat -- . ':(exclude)openspec/'
```

`git add -A` respects `.gitignore`; never force-add.

**The one commit exception.** If the state file records a `prUrl`, a PR is already open and a
staged-only fix would be invisible on it — commit and push to the PR branch instead of leaving the
work staged. Otherwise never commit.

Write the state file: `IN_PROGRESS` from `STARTED`, otherwise **the state exactly as read**.
Populate `worktrees` with one absolute-path key per affected worktree and its merge base. Carry
`artifactUrl`, `jiraIssue` and `prUrl` forward verbatim. The state file lives outside the repo —
never `git add` it.

```
## Implementation staged — review and test

**Change:** <name>
**Panel:** clean — required: primary + Bugbot + Principles; optional: <selected, or "none — no triggers fired">
**Progress:** N/N tasks
**Git:** staged and uncommitted | committed and pushed to the PR branch

Worktree:   <absolute worktree path>
Test guide: <absolute path to docs/manual-test/<name>.md>

Review the staged diff, then run the apps against the guide:
  git -C <absolute worktree path> diff --cached -- . ':(exclude)openspec/'
  open -na "IntelliJ IDEA" --args "<absolute worktree path>"

Re-run this command to fix anything you find.

Next:
/myflow-finish <name>
```

## Guardrails

- **Never commit, push, merge, or open a PR** — except the `prUrl` exception above.
- **Never** run `finishing-a-development-branch`.
- **Never** create a second worktree for the same change.
- **Never** advance the state from `IN_PROGRESS`; write back what you read.
- **Always `git add -A`** in every affected worktree before handing off.
- **Never skip** a required panel slot, and never collapse two slots into one prompt.
- **Never** pass a model override to Bugbot or Security Review; **always** name Sonnet on every
  other slot.
- **Never** paste the principle list into a prompt — the reviewer reads the file.
- **Never** hand off with an open Critical/Important finding, or a stale clean result.
- **Never** mark a checkbox before its task review passes.
- **No flags.** The only argument is the optional change name; report anything else.
