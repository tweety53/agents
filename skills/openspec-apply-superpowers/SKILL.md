---
name: openspec-apply-superpowers
description: Implement an OpenSpec change using Superpowers Basic Workflow #2–#6 (stage with git add; no commits; #7 deferred to code review). Use for /myflow-do.
allowed-tools: Bash(openspec:*)
license: MIT
compatibility: Requires openspec CLI and Superpowers plugin skills.
metadata:
  author: gymie
  version: "3.5"
---

Implement an OpenSpec change with Superpowers Basic Workflow steps **#2–#6**. **No git commits, push, merge, or PR in this stage** — those happen in `/myflow-code-review` after the user's manual review (Gate B) and manual test (Gate C). **Do stage (`git add`) all apply changes** before handoff so the IDE shows them in Source Control.

**Announce at start:** "Using openspec-apply-superpowers for change `<name>`."

Also follow **rules/myflow-manual-review.mdc** (Cursor: `.cursor/rules/myflow-manual-review.mdc`).

## Superpowers Basic Workflow (this stage)

| Step | Skill | When |
|------|-------|------|
| **2** | **using-git-worktrees** | Before first code change |
| **3** | **writing-plans** | Validate plan; re-run if `tasks.md` not apply-ready |
| **4** | **subagent-driven-development** | Execute remaining tasks (default). **executing-plans** only if user explicitly wants a separate session |
| **5** | **test-driven-development** | Every implementer subagent, every task |
| **6** | **requesting-code-review** + **strict review panel** | Per-task review (via SDD) + final whole-branch review (**1 primary + 5 additional agents**) |

**Deferred to code review:** **#7** `finishing-a-development-branch` (commit, push, open PR — never merges; PR review is Gate D, then `/myflow-finish` archives).

Step **#1** (brainstorming) completed in **openspec-propose-superpowers**.

## Required sub-skills

1. **openspec-apply-change** — steps 1–5 (load context, show progress). **Do not** use its minimal step-6 loop.
2. **superpowers:using-git-worktrees** — Basic Workflow **#2**.
3. **superpowers:writing-plans** — Basic Workflow **#3** (validate or repair plan).
4. **superpowers:subagent-driven-development** — Basic Workflow **#4** (with no-commit override below).
5. **superpowers:test-driven-development** — Basic Workflow **#5** (mandatory per implementer dispatch).
6. **superpowers:requesting-code-review** — Basic Workflow **#6** primary final reviewer (after SDD tasks).
7. **Strict review panel (5 additional agents)** — mandatory alongside #6; see below.
8. **superpowers:verification-before-completion** — evidence before claiming apply done.

**Do not invoke** `finishing-a-development-branch` in this stage.

## No-commit SDD override (mandatory)

SDD implementers normally commit per task. **Override for myflow apply:**

Every implementer dispatch **must** include:

> **MYFLOW APPLY — NO COMMITS:** Do **not** run `git commit`, `git push`, merge, or open a PR. Leave all changes uncommitted in the worktree. You **may** `git add` / stage files. The parent agent records `TASK_BASE=$(git rev-parse HEAD)` before dispatch; your diff for review is `git diff TASK_BASE` (plus `git diff --cached TASK_BASE` if you staged files). The parent will `git add -A` before Gate B handoff.

Every implementer dispatch **must** include TDD requirement:

> **REQUIRED SUB-SKILL:** Use superpowers:test-driven-development — RED-GREEN-REFACTOR for this task. Delete any code written before tests.

### Per-task review without commits

- Before dispatch: `TASK_BASE=$(git rev-parse HEAD)`; record in progress ledger.
- After implementer returns: write review artifact — `git diff TASK_BASE > .superpowers/sdd/task-N.diff` (or equivalent path the reviewer reads).
- Dispatch task reviewer with the diff file path instead of commit SHAs.
- Progress ledger line: `Task N: complete (uncommitted, review clean)`.
- **Never** use `review-package` with `HEAD~1`; it assumes commits.

### Final whole-branch review (#6) — strict multi-agent panel

- `MERGE_BASE` = commit recorded at worktree setup (from progress ledger or `git merge-base HEAD main`).
- Write `.superpowers/sdd/final-review.diff` via `git diff MERGE_BASE` (include staged + unstaged).
- Run the **strict review panel** below in **every** affected repo/worktree (backend + sibling frontends).
- Fix Critical/Important findings from **any** panel agent before handoff; re-diff and **re-run the full panel** until clean.
- Record panel results in `.superpowers/sdd/final-review-panel.md` (one section per agent).

#### Strict review panel (mandatory — 1 primary + 5 additional)

Dispatch **six separate** review subagents. Prefer **parallel** spawn of the five additional agents after (or with) the primary. Do **not** skip any agent. Do **not** merge their roles into one prompt.

| # | Role | How to spawn (Cursor) | Portable fallback |
|---|------|------------------------|-------------------|
| 0 | **Primary** — plan alignment + code quality | `generalPurpose` via **superpowers:requesting-code-review** / `code-reviewer.md`; pass `final-review.diff` + plan/spec constraints | same |
| 1 | **Bugbot** — defect hunt | `subagent_type: bugbot`, `description: "Bugbot"`, prompt shape from `review-bugbot` skill with `Diff: uncommitted changes` and `Full Repository Path: <worktree>` | `generalPurpose` + [bug-hunter-reviewer-prompt.md](bug-hunter-reviewer-prompt.md) |
| 2 | **Security** — authZ / injection / secrets | `subagent_type: security-review`, `description: "Security Review"`, prompt shape from `review-security` skill with `Diff: uncommitted changes` and `Full Repository Path: <worktree>` | `generalPurpose` + [security-reviewer-prompt.md](security-reviewer-prompt.md) |
| 3 | **Adversarial** — skeptic / regressions / test theater | `generalPurpose` + [adversarial-reviewer-prompt.md](adversarial-reviewer-prompt.md) | same |
| 4 | **Senior engineer** — pragmatic teammate PR review | `generalPurpose` + [senior-engineer-reviewer-prompt.md](senior-engineer-reviewer-prompt.md); **omit** `model` (inherit parent) | same |
| 5 | **Economic senior engineer** — same persona, economy model | `generalPurpose` + [senior-engineer-reviewer-prompt.md](senior-engineer-reviewer-prompt.md); **must** set `model` to the economic sibling of the parent agent (table below); `description: "Economic senior engineer review"` | same + economic `model` |

##### Economic model mapping (slot 5 only)

Detect the **parent / current agent provider family**, then pass the matching economy-tier slug to Task `model`. Do **not** use the parent's main model for slot 5.

| Parent provider family (examples) | Economic model slug |
|-----------------------------------|---------------------|
| Cursor Grok (`cursor-grok*`, Grok) | `composer-2.5-fast` |
| Composer (`composer*`) | `composer-2.5-fast` |
| Claude Opus / Sonnet / Fable (`claude-opus*`, `claude-sonnet*`, `claude-fable*`) | `claude-4.6-sonnet-medium-thinking` |
| GPT (`gpt-5.6*`, `gpt-5.5*`, `gpt-5.3*`) | `gpt-5.5-medium` |
| Unknown / other | `composer-2.5-fast` |

If the resolved economic slug is unavailable in the Task tool allowlist, fall back to `composer-2.5-fast`. Never skip slot 5 because of model selection — pick the closest economy sibling and continue.

**Aggregation rules:**

- Union all Critical/Important findings; dedupe by file:line + theme.
- One fix subagent gets the combined list (not one fixer per agent).
- After fixes: rewrite `final-review.diff`, re-dispatch **all six** agents (full panel), not only the agent that complained.
- Gate B handoff is blocked while any panel agent still reports Critical/Important.
- If Bugbot or Security Review fails to start (wrong prompt / empty diff), retry once per their skill rules, then fall back to the portable prompt templates — still as a **separate** spawned agent.

## Workflow

### 0. Check stage

Read the change's state per **State file** and validate per **State self-heal** in `rules/myflow-manual-review.mdc`.

This command requires stage **`start`**. If the change is at any other stage, **stop** and emit the mismatch handoff from **Stage transitions**, then AskUserQuestion for an explicit override (default: **No — run the suggested command instead**).

The most common mismatch: the change is at `awaiting-review` or later, meaning the user probably wants `/myflow-do-fix <name>`. Recommend that. This replaces the older ad-hoc "guard against a mistaken re-run" heuristic — the stage check now covers it.

Missing state file → infer per **State self-heal**, write it, announce the correction, then apply the check to the inferred stage.

### 1. Load OpenSpec context (apply steps 1–5)

**Resolve the change name.** If given, use it. **If omitted:** run `openspec list --json`, filter to changes with apply-ready planning artifacts (not yet archived). Exactly one match → use it automatically, announce which; multiple matches → **AskUserQuestion** listing each (name, status, last modified) — do not guess; zero matches → stop, suggest `/myflow-start <name>`.

Follow **openspec-apply-change** steps 1–5:

```bash
openspec status --change "<name>" --json
openspec instructions apply --change "<name>" --json
```

- If `state: "blocked"`: stop; suggest `/myflow-start <name>` or `openspec-continue-change`.
- If `state: "all_done"`: suggest `/myflow-code-review <name>` (if not yet reviewed/committed) or `/myflow-finish <name>` (if the PR is already merged).
- Read every path in `contextFiles` (proposal, specs, design, tasks).

Resolve paths from CLI JSON (`changeRoot`, `planningHome`) — do not assume repo layout.

Show: schema, progress N/M, remaining tasks overview.

### 2. Basic Workflow #3 — Validate plan

Before any code change, confirm `<changeRoot>/tasks.md` meets **writing-plans** quality:

- Exact file paths, verification commands, bite-sized steps for unchecked tasks
- No placeholders ("TBD", "implement later", "add tests")

If inadequate: invoke **superpowers:writing-plans** to repair `tasks.md` (and optional `docs/superpowers/plans/…` mirror). Do not proceed to #2 until plan is apply-ready.

Extract **Global constraints** verbatim from delta specs + `design.md` for SDD reviewers.

### 3. Basic Workflow #2 — Isolate workspace

Invoke **superpowers:using-git-worktrees** before the first code change:

- Branch name: `openspec/<change-name>` unless user specifies otherwise.
- Never implement on `main`/`master` without explicit user consent.
- Complete worktree setup + clean baseline per that skill.
- Record `MERGE_BASE` in `.superpowers/sdd/progress.md`.

### 4. Basic Workflow #4 — Execute (SDD)

Invoke **superpowers:subagent-driven-development** with the **no-commit override** above:

- Treat each remaining OpenSpec checkbox (or tightly coupled group) as one SDD task.
- Use SDD `task-brief` scripts; pass spec constraints + design excerpts via task briefs.
- After each task passes **spec ✅ and quality ✅** review: update `tasks.md` `- [ ]` → `- [x]`.
- On BLOCKED: pause and report; do not guess.

**Progress ledger:** `.superpowers/sdd/progress.md` + OpenSpec checkboxes — both stay in sync.

### 5. Basic Workflow #6 — Code review (strict panel)

SDD per-task reviewers satisfy **between-task** review. Before handoff, run the **final whole-branch strict review panel** (primary requesting-code-review **plus** Bugbot, Security Review, Adversarial, Senior engineer, and Economic senior engineer — see above).

Critical/Important findings from **any** of the six agents must be fixed and the **full panel** re-run before handoff.

### 6. Verify completion

Invoke **superpowers:verification-before-completion**:

- Run project-appropriate tests; show command output.
- Re-read `tasks.md`; all intended checkboxes must be `[x]`.
- Run `openspec instructions apply --change "<name>" --json` and confirm no pending tasks.
- Confirm **no commits** were made since `MERGE_BASE` (or since apply resume point): `git log MERGE_BASE..HEAD` should be empty unless user passed `commit-during-apply`.

### 7. Stage for IDE review, then hand off (not archive)

**Do not** commit, push, merge, or run #7. **Do** stage all apply changes so Gate B is visible in the IDE.

In **every** affected repo/worktree (backend worktree and any sibling frontend repos touched):

```bash
cd <worktree-or-repo>
git add -A
git status
git diff --cached --stat
```

- `git add -A` respects `.gitignore` (do not force-add secrets).
- Confirm implementation files appear under **Changes to be committed** (staged), not only as unstaged/untracked.
- If a sibling repo was modified outside the main worktree, stage there too and list each path in the handoff.

Write the state file per **State file** before handing off:

```json
{
  "stage": "awaiting-review",
  "gates": { "reviewed": false, "tested": null, "prOpened": null, "prMerged": null },
  "worktree": "<absolute worktree path>",
  "branch": "openspec/<name>",
  "updatedAt": "<ISO-8601 UTC now>",
  "updatedBy": "/myflow-do"
}
```

Include it in the `git add -A` so it is staged with the rest of the work.

Stop here.

```
## Apply Complete — Manual Review Required

**Change:** <name>
**Basic Workflow:** #2 ✓ #3 ✓ #4 ✓ #5 ✓ #6 ✓ (strict panel: primary + Bugbot + Security + Adversarial + Senior + Economic Senior)
**Deferred to code review:** #7 (commit + push + open PR — never merges)
**Progress:** N/N tasks complete
**Branch:** openspec/<name>
**Worktree:** <absolute path>
**Merge base:** <MERGE_BASE>
**Final review panel:** clean (see `.superpowers/sdd/final-review-panel.md`)
**Git state:** staged + uncommitted (not pushed) — open worktree in IDE Source Control

**Review locally:**
  # Open this folder in Cursor/IDE to see staged changes:
  <absolute path>
  cd <worktree>
  git status
  git diff --cached --stat
  git diff --cached              # full staged diff
  git diff <MERGE_BASE>         # full tree vs merge base (staged + any leftover unstaged)

**Next steps:**
- Open the worktree folder in your IDE and review staged changes (Gate B).
- Request fixes: `/myflow-do-fix <name>` (resumes on same branch/worktree)
- After manual review looks good: `/myflow-manual-test <name>` (Gate C — run guide + checklist MD)
- After manual testing: `/myflow-code-review <name>` (coverage check, tests/linters, commit + push + open PR — never merges), then a human reviews and merges the PR (Gate D), then `/myflow-finish <name>` (verify PR merged, sync specs, archive)
```

## Guardrails

- **Never skip** Basic Workflow steps #2–#6 when implementing.
- **Never commit, push, merge, or open a PR** during apply unless user explicitly passed `commit-during-apply`.
- **Always check the incoming stage first** (step 0) — never start applying a change that is past `start` without an explicit user override.
- **Always write `stage: awaiting-review`** before the Gate B handoff, and stage the state file.
- **Always `git add -A`** (stage) in every affected repo before Gate B handoff — so the IDE shows the changes.
- **Never** run `finishing-a-development-branch` (#7) during apply.
- Do not use the lightweight openspec-apply-change step-6 loop.
- Do not skip per-task SDD review (#6) or final whole-branch review (#6).
- Do not skip any of the **five additional** final-review agents (Bugbot, Security, Adversarial, Senior engineer, Economic senior engineer), and do not collapse them into one agent.
- Do not omit the `model` parameter on the Economic senior engineer (slot 5); resolve it from the economic model mapping.
- Do not hand off to Gate B while any panel agent still has open Critical/Important findings.
- Do not skip TDD (#5) on any implementer dispatch.
- Do not mark OpenSpec checkboxes before task review passes.
- Pause on ambiguity, design conflicts, or failing verification — suggest `openspec-update-change` if artifacts need revision.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Apply (#2–#6, stage; no commits) | `/myflow-do <name>` |
| Continue partial apply | Same command; SDD ledger + unchecked tasks resume |
| Fix a Gate B/C finding instead | `/myflow-do-fix <name>` |
| Legacy per-task commits | `/myflow-do <name> commit-during-apply` |
