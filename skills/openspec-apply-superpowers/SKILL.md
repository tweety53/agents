---
name: openspec-apply-superpowers
description: Implement an OpenSpec change using Superpowers Basic Workflow #2–#6 (stage with git add; no commits; #7 deferred to review). Use for /myflow-do.
allowed-tools: Bash(openspec:*)
license: MIT
compatibility: Requires openspec CLI and Superpowers plugin skills.
metadata:
  author: gymie
  version: "3.5"
---

Implement an OpenSpec change with Superpowers Basic Workflow steps **#2–#6**. **No git commits, push, merge, or PR in this stage** — those happen in `/myflow-review` after the user's manual review (Gate B) and manual test (Gate C). **Do stage (`git add`) all apply changes** before handoff so the IDE shows them in Source Control.

**Announce at start:** "Using openspec-apply-superpowers for change `<name>`."

Also follow **rules/myflow-manual-review.mdc** (Cursor: `.cursor/rules/myflow-manual-review.mdc`).

## Superpowers Basic Workflow (this stage)

| Step | Skill | When |
|------|-------|------|
| **2** | **using-git-worktrees** | Before first code change |
| **3** | **writing-plans** | Validate plan; re-run if `tasks.md` not apply-ready |
| **4** | **subagent-driven-development** | Execute remaining tasks (default). **executing-plans** only if user explicitly wants a separate session |
| **5** | **test-driven-development** | Every implementer subagent, every task |
| **6** | **requesting-code-review** + **strict review panel** | Per-task review (via SDD) + final whole-branch review (**3 required slots + conditional slots**) |

**Deferred to code review:** **#7** `finishing-a-development-branch` (commit, push, open PR — never merges; PR review is Gate D, then `/myflow-finish` archives).

Step **#1** (brainstorming) completed in **openspec-propose-superpowers**.

## Required sub-skills

1. **openspec-apply-change** — steps 1–5 (load context, show progress). **Do not** use its minimal step-6 loop.
2. **superpowers:using-git-worktrees** — Basic Workflow **#2**.
3. **superpowers:writing-plans** — Basic Workflow **#3** (validate or repair plan).
4. **superpowers:subagent-driven-development** — Basic Workflow **#4** (with no-commit override below).
5. **superpowers:test-driven-development** — Basic Workflow **#5** (mandatory per implementer dispatch).
6. **superpowers:requesting-code-review** — Basic Workflow **#6** primary final reviewer (after SDD tasks).
7. **Strict review panel** — mandatory alongside #6: Bugbot and the Principles reviewer always, plus every optional slot the triggers select; see below.
8. **superpowers:verification-before-completion** — evidence before claiming apply done.

**Do not invoke** `finishing-a-development-branch` in this stage.

## No-commit SDD override (mandatory)

SDD implementers normally commit per task. **Override for myflow apply:**

Every implementer dispatch **must** include:

> **MYFLOW APPLY — NO COMMITS:** Do **not** run `git commit`, `git push`, merge, or open a PR. Leave all changes uncommitted in the worktree. You **may** `git add` / stage files. The parent agent records `TASK_BASE=$(git rev-parse HEAD)` before dispatch; your diff for review is `git diff TASK_BASE` (plus `git diff --cached TASK_BASE` if you staged files). The parent will `git add -A` before Gate B handoff.

Every implementer dispatch **must** include TDD requirement:

> **REQUIRED SUB-SKILL:** Use superpowers:test-driven-development — RED-GREEN-REFACTOR for this task. Delete any code written before tests.

Every implementer dispatch **must** include the principles requirement:

> **REQUIRED READING:** [engineering-principles.md](engineering-principles.md) — your implementation must satisfy these principles; the review panel's principles reviewer checks the diff against them.

### Per-task review without commits

- Before dispatch: `TASK_BASE=$(git rev-parse HEAD)`; record in progress ledger.
- After implementer returns: write review artifact — `git diff TASK_BASE > .superpowers/sdd/task-N.diff` (or equivalent path the reviewer reads).
- Dispatch task reviewer with the diff file path instead of commit SHAs.
- Progress ledger line: `Task N: complete (uncommitted, review clean)`.
- **Never** use `review-package` with `HEAD~1`; it assumes commits.

### Final whole-branch review (#6) — strict multi-agent panel

- `MERGE_BASE` = commit recorded at worktree setup (from progress ledger or `git merge-base HEAD main`).
- Write `.superpowers/sdd/final-review.diff` via `git diff MERGE_BASE` (include staged + unstaged).
- Run the **strict review panel** below in **every** affected repo/worktree. Resolve that set from the change's own record — the `MERGE_BASE` key set, or the apps listed under `## apps` in the project's `.myflow/project.md` (see **Project configuration** in `rules/myflow-manual-review.mdc`). Never assume a repo topology remembered from another project.
- Fix Critical/Important findings from **any** panel agent before handoff; re-diff and re-run per **Panel re-runs** below until clean.
- Record panel results in `.superpowers/sdd/final-review-panel.md` (one section per agent, one block per pass).

#### Strict review panel (three required slots + conditional slots)

Dispatch **separate** review subagents — one per selected slot. Prefer **parallel** spawn of the non-primary agents after (or with) the primary. Do **not** skip a slot the selection rules included. Do **not** merge their roles into one prompt.

| # | Slot | Required? | Model | How to spawn (Cursor) | Portable fallback |
|---|------|-----------|-------|------------------------|-------------------|
| 0 | **Primary** — plan alignment + code quality | **always** | parent (**omit** `model`) | `generalPurpose` via **superpowers:requesting-code-review** / `code-reviewer.md`; pass `final-review.diff` + plan/spec constraints | same |
| 1 | **Bugbot** — defect hunt | **always** | own | `subagent_type: bugbot`, `description: "Bugbot"`, prompt shape from `review-bugbot` skill with `Diff: uncommitted changes` and `Full Repository Path: <worktree>` | `generalPurpose` + [bug-hunter-reviewer-prompt.md](bug-hunter-reviewer-prompt.md) |
| 2 | **Principles** — merged principle list + project hard invariants | **always** | parent (**omit** `model`) | `generalPurpose` + [principles-reviewer-prompt.md](principles-reviewer-prompt.md) with `[LENS]` = **Merged**; `description: "Principles review (Merged)"` | same |
| 3 | **Security** — authZ / injection / secrets | conditional | own | `subagent_type: security-review`, `description: "Security Review"`, prompt shape from `review-security` skill with `Diff: uncommitted changes` and `Full Repository Path: <worktree>` | `generalPurpose` + [security-reviewer-prompt.md](security-reviewer-prompt.md) |
| 4 | **Adversarial** — skeptic / regressions / test theater | conditional | parent (**omit** `model`) | `generalPurpose` + [adversarial-reviewer-prompt.md](adversarial-reviewer-prompt.md) | same |
| 5+ | **Principles lens B / lens C** — extra breadth on a narrowed lens | conditional | **economy** (mapping below) | `generalPurpose` + [principles-reviewer-prompt.md](principles-reviewer-prompt.md) with `[LENS]` = **Lens B — simplicity & state** or **Lens C — robustness & ops**; **must** set `model` | same + economic `model` |

Slot 2 is the panel's only mandatory judgment check on *how* the code is built. It reads [engineering-principles.md](engineering-principles.md) — never a pasted copy of the list — and additionally owns the **hard invariants** read out of the project's own standards files (`[STANDARDS_PATHS]`): architecture/layer purity, new suppressions, weakened lint config. Those checks came from the retired conventions slot and still block; they were relocated, not dropped.

**Resolve `[PRINCIPLES_PATH]` before dispatching any principles slot (2 and 5+).** It is the **absolute** path of `engineering-principles.md` in the skill directory you are reading this file from — `<that directory>/engineering-principles.md`; under the global install, `~/.claude/skills/openspec-apply-superpowers/engineering-principles.md`. The subagent's working directory is the **project worktree**, which has no `skills/` tree, so a repo-relative `skills/…` path fails to open and the reviewer runs with no principle list. Confirm the file exists before spawning; if it does not, stop and report it rather than dispatching a blind reviewer.

**Resolve `[STANDARDS_PATHS]` before dispatching slot 2** — from the entries listed under the `## standards` section of the project's `.myflow/project.md` when it has one, else auto-detection. Entries are **not** paths to use as-is: each resolves to an absolute path through the entry-form table and the containment rule, and an entry that fails either is reported by name and dropped. The resolution order is stated once, with the placeholder it fills, in [principles-reviewer-prompt.md](principles-reviewer-prompt.md); follow it there rather than a copy here (the entry forms and containment rule themselves are defined under **Project configuration** in `rules/myflow-manual-review.mdc`). Pass the **resolved absolute paths**, and pass an **empty** value when none resolve — that empties the Hard Invariants section by design and is a correct outcome in an unconfigured project, not a reason to substitute standards from anywhere else. Record in `final-review-panel.md` which standards files were passed, or that none resolved.

Slots 5+ are the same template on a **narrowed** lens, which is what makes them worth an economy-tier agent: breadth over one theme, not a second opinion on everything. **No two principle reviewers in one run may share a lens.**

##### Optional slot selection

Evaluate these triggers against `final-review.diff` **before** dispatching, and record in `final-review-panel.md` which optional slots were included and which were excluded and why.

| Slot | Include when the diff touches | Ask when |
|------|-------------------------------|----------|
| 3 — Security | auth/authz, JWT/tokens, crypto, secrets or config, SQL/query construction, path or file handling, deserialization, CORS/HTTP edge, new dependencies, or any gateway/auth module file | a config or dependency file changed, but **only** comments or a version bump |
| 4 — Adversarial | DB migrations, concurrency/scheduling, behavior changes to code with existing tests, any test modified or deleted, or **>~300** changed lines | **150–300** changed lines with no other trigger |
| 5 — Lens B (simplicity & state) | **>~200** changed lines, or **≥3** new classes/modules | — |
| 5 — Lens C (robustness & ops) | error handling, retries, schedulers, external integrations, config/env, migrations | — |

**Borderline → ask.** When an "Ask when" cell fires and no other trigger for that slot already applies, use **AskUserQuestion**: name the slot, say why it is borderline, and offer **include** as the default/recommended answer. Erring toward including a reviewer costs tokens; erring toward excluding costs a defect.

**`full-panel` forces every slot** — 0 through 5+, including *both* extra lenses — and **bypasses trigger evaluation** entirely. It is opt-in and never inferred.

A documentation-, prompt-, or test-only diff with no trigger runs the three required slots alone. That is a correct outcome, not a skipped review — say so explicitly in the panel record.

##### Economic model mapping (slots 5+ only)

**This mapping applies only to the conditional extra lens reviewers (slots 5+).** The **required** principles reviewer (slot 2) inherits the **parent model** — pass **no** `model` override at all. Weighing principle tradeoffs is judgment work that degrades on a weaker agent; a narrowed breadth pass does not.

Detect the **parent / current agent provider family**, then pass the matching economy-tier slug to Task `model` for each slot 5+ reviewer.

| Parent provider family (examples) | Economic model slug |
|-----------------------------------|---------------------|
| Cursor Grok (`cursor-grok*`, Grok) | `composer-2.5-fast` |
| Composer (`composer*`) | `composer-2.5-fast` |
| Claude Opus / Sonnet / Fable (`claude-opus*`, `claude-sonnet*`, `claude-fable*`) | `claude-4.6-sonnet-medium-thinking` |
| GPT (`gpt-5.6*`, `gpt-5.5*`, `gpt-5.3*`) | `gpt-5.5-medium` |
| Unknown / other | `composer-2.5-fast` |

If the resolved economic slug is unavailable in the Task tool allowlist, fall back to `composer-2.5-fast`. Never skip a slot 5+ reviewer because of model selection — pick the closest economy sibling and continue.

**Aggregation rules:**

- Union all Critical/Important findings; dedupe by file:line + theme.
- One fix subagent gets the combined list (not one fixer per agent).
- After fixes: re-run per **Panel re-runs** below.
- Gate B handoff is blocked while any panel agent still reports Critical/Important.
- If Bugbot or Security Review fails to start (wrong prompt / empty diff), retry once per their skill rules, then fall back to the portable prompt templates — still as a **separate** spawned agent.

#### Panel re-runs (targeted by default)

**Pass 1 always runs the full roster selected for this change** over `final-review.diff` — every required slot plus every optional slot the trigger rules included. Only *re-runs* after a fix round are scoped; the first look at the branch is never narrowed for cost.

After each fix round, record `FIX_BASE` before the fix subagent runs and write the fix-scoped diff:

```bash
git diff FIX_BASE > .superpowers/sdd/fix-round-N.diff
```

Then choose the re-run shape:

| Mode | Who re-runs | Diff they get |
|------|-------------|---------------|
| **Targeted** (default) | Slot 0 primary (always, as integration check) + every agent that raised a finding in the round being fixed | `fix-round-N.diff` |
| **Full** (escalation or flag) | Every slot in this run's roster | rewritten `final-review.diff` (`git diff MERGE_BASE`) |

**Escalate a targeted re-run to full automatically** when any of these hold — do not ask, just escalate and say why in the panel record:

- The fix touched a file or module **outside** the set named in the findings it was fixing.
- The fix diff exceeds **~150 changed lines**.
- The fix altered a delta spec, a DB migration, or a public contract or port boundary as defined by the project's standards (the files under `## standards`, or auto-detection when none are configured).
- A targeted re-run surfaced a **new** Critical finding (not a restatement of the one being fixed).
- Three or more fix rounds have already run on this change — drift risk outweighs the saving.

**`full-panel` flag.** When the user passes `full-panel` to `/myflow-do` (or `/myflow-do-fix`), every slot is dispatched — trigger evaluation is bypassed, both extra lenses run — and every re-run is that full roster over the whole-branch diff. Use it when the change is large, security-sensitive, or when a previous targeted round missed something. The flag is opt-in and never inferred.

**Invariants** — targeting is a cost optimization, never a coverage waiver:

- A targeted re-run is **never fewer than two agents** (primary + the originating agent).
- Gate B handoff still requires **zero** open Critical/Important findings from every agent that has run, whatever the mode.
- The **final** pass before handoff must show a clean result for **every slot in this run's roster** — from that pass or an earlier one with no intervening change to the files that agent flagged. If any slot's clean result is stale under that test, run the full roster once before handing off.
- Record in `final-review-panel.md`, per pass: mode (targeted/full), which agents ran, why (finding IDs or escalation reason), and the diff path they reviewed.

## Workflow

### 0. Check stage

Read the change's state per **State file** and validate per **State self-heal** in `rules/myflow-manual-review.mdc`.

This command requires stage **`proposal-done`**. If the change is at any other stage, **stop** and emit the mismatch handoff from **Stage transitions**, then AskUserQuestion for an explicit override (default: **No — run the suggested command instead**).

The most common mismatch: the change is at `awaiting-do-review` or later, meaning the user probably wants `/myflow-do-fix <name>`. Recommend that. This replaces the older ad-hoc "guard against a mistaken re-run" heuristic — the stage check now covers it.

Missing state file → infer per **State self-heal**, write it, announce the correction, then apply the check to the inferred stage.

### 1. Load OpenSpec context (apply steps 1–5)

**Resolve the change name.** If given, use it. **If omitted:** run `openspec list --json`, filter to changes with apply-ready planning artifacts (not yet archived). Exactly one match → use it automatically, announce which; multiple matches → **AskUserQuestion** listing each (name, status, last modified) — do not guess; zero matches → stop, suggest `/myflow-start <name>`.

Follow **openspec-apply-change** steps 1–5:

```bash
openspec status --change "<name>" --json
openspec instructions apply --change "<name>" --json
```

- If `state: "blocked"`: stop; suggest `/myflow-start <name>` or `openspec-continue-change`.
- If `state: "all_done"`: suggest `/myflow-review <name>` (if not yet reviewed/committed) or `/myflow-finish <name>` (if the PR is already merged).
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

SDD per-task reviewers satisfy **between-task** review. Before handoff, run the **final whole-branch strict review panel** — the three required slots (primary requesting-code-review, Bugbot, Principles) plus every optional slot the trigger rules selected (Security, Adversarial, extra principle lenses B/C) — see above.

Critical/Important findings from **any** dispatched agent must be fixed and the panel re-run per **Panel re-runs** — targeted by default, full on escalation or the `full-panel` flag — before handoff.

### 6. Verify completion

Invoke **superpowers:verification-before-completion**:

- Run project-appropriate tests; show command output.
- Re-read `tasks.md`; all intended checkboxes must be `[x]`.
- Run `openspec instructions apply --change "<name>" --json` and confirm no pending tasks.
- Confirm **no commits** were made since `MERGE_BASE` (or since apply resume point): `git log MERGE_BASE..HEAD` should be empty unless user passed `commit-during-apply`.

### 7. Stage for IDE review, then hand off (not archive)

**Do not** commit, push, merge, or run #7. **Do** stage all apply changes so Gate B is visible in the IDE.

In **every** affected repo/worktree — the set resolved above from the `MERGE_BASE` key set or `## apps`, never a topology assumed from another project:

```bash
cd <worktree-or-repo>
git add -A
git status
git diff --cached --stat
```

- `git add -A` respects `.gitignore` (do not force-add secrets).
- Confirm implementation files appear under **Changes to be committed** (staged), not only as unstaged/untracked.
- If a sibling repo was modified outside the main worktree, stage there too and list each path in the handoff.

Write the state file before handing off. Resolve its path per **State file** in `rules/myflow-manual-review.mdc` (`--git-common-dir` → `<project-key>` → `/Users/tweety53/Agents/myflow/state/<project-key>/<name>.json`) — resolving via `--git-common-dir` is what makes the worktree and the main checkout agree on one file:

```json
{
  "stage": "awaiting-do-review",
  "gates": { "reviewed": false, "tested": null, "prOpened": null, "prMerged": null },
  "worktree": "<absolute worktree path>",
  "branch": "openspec/<name>",
  "originStage": null,
  "artifactUrl": "<unchanged — carried forward from the file as read>",
  "jiraIssue": "<unchanged — carried forward from the file as read>",
  "fastPath": <carried forward from the file as read>,
  "REVIEWED_TREE": <carried forward from the file as read>,
  "MERGE_BASE": <carried forward from the file as read>,
  "updatedAt": "<ISO-8601 UTC now>",
  "updatedBy": "/myflow-do"
}
```

The state file lives **outside** the repo — do not `git add` it, do not commit it, and do not archive it. Carry forward any gate values already present in the file; gates are monotonic.

**`artifactUrl`, `jiraIssue`, `fastPath`, `REVIEWED_TREE`, and `MERGE_BASE` are carried forward, never dropped.** Writes render the whole object, so omitting any of them destroys it permanently: `artifactUrl` is the proposal link `myflow-status` surfaces, `jiraIssue` is the link to the issue, `fastPath` and `REVIEWED_TREE` are what let a fast-path change resume, and `MERGE_BASE`'s key set is the authoritative list of affected worktrees for a multi-repo change. Read the existing file first and re-emit each value verbatim (`null` only if it was already `null`). This command makes **no** Jira call of its own — see **Jira integration** in `rules/myflow-manual-review.mdc`.

Stop here.

```
## Apply Complete — Manual Review Required

**Change:** <name>
**Basic Workflow:** #2 ✓ #3 ✓ #4 ✓ #5 ✓ #6 ✓ (strict panel — required: primary + Bugbot + Principles; optional selected: <Security / Adversarial / lens B / lens C, or "none — no triggers fired">)
**Deferred to review:** #7 (commit + push + open PR — never merges)
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

**Open in IntelliJ:**
open -na "IntelliJ IDEA" --args "<absolute worktree path>"

**Next:** `/myflow-do-manual-review <name>` to mark review started, or `/myflow-do-done <name>` when the diff looks right.

**Next steps:**
- Open the worktree folder in your IDE and review staged changes (Gate B).
- Request fixes: `/myflow-do-fix <name>` (resumes on same branch/worktree)
- After manual review looks good: `/myflow-do-done <name>` to advance to `do-done`, then `/myflow-manual-test <name>` (Gate C — run guide + checklist MD)
- After manual testing: `/myflow-review <name>` (coverage check, tests/linters, commit + push + open PR — never merges), then a human reviews and merges the PR (Gate D), then `/myflow-finish <name>` (verify PR merged, sync specs, archive)
```

## Guardrails

- **Never skip** Basic Workflow steps #2–#6 when implementing.
- **Never commit, push, merge, or open a PR** during apply unless user explicitly passed `commit-during-apply`.
- **Always check the incoming stage first** (step 0) — never start applying a change that is past `proposal-done` without an explicit user override.
- **Always write `stage: awaiting-do-review`** before the Gate B handoff — to the user-scoped state file, which is **never staged or committed**.
- **Always `git add -A`** (stage) in every affected repo before Gate B handoff — so the IDE shows the changes.
- **Never** run `finishing-a-development-branch` (#7) during apply.
- Do not use the lightweight openspec-apply-change step-6 loop.
- Do not skip per-task SDD review (#6) or final whole-branch review (#6).
- **Never skip a required slot:** primary, Bugbot, and Principles run on **every** panel pass 1, in every affected repo, and are never collapsed into one agent. Re-runs may be targeted per **Panel re-runs**; pass 1 may not.
- Do not decide the optional slots (Security, Adversarial, lens B, lens C) by feel — evaluate **Optional slot selection** against the diff, ask when a borderline cell fires (default **include**), and record what was excluded and why.
- Do not pass a `model` override to the required Principles reviewer (slot 2) — it inherits the parent model. Do not omit `model` on a slot 5+ lens reviewer — resolve it from the economic model mapping.
- Do not dispatch two principle reviewers with the same `[LENS]`, and do not paste the principle list into a prompt — the reviewer reads `engineering-principles.md`.
- Do not hand off while any slot's clean result is stale under the **Panel re-runs** invariant; run the full panel once if in doubt.
- Do not hand off to Gate B while any panel agent still has open Critical/Important findings.
- Do not skip TDD (#5) on any implementer dispatch.
- Do not mark OpenSpec checkboxes before task review passes.
- Pause on ambiguity, design conflicts, or failing verification — suggest `openspec-update-change` if artifacts need revision.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Apply (#2–#6, stage; no commits) | `/myflow-do <name>` |
| Apply, full panel on every re-run | `/myflow-do <name> full-panel` |
| Continue partial apply | Same command; SDD ledger + unchecked tasks resume |
| Fix a Gate B/C finding instead | `/myflow-do-fix <name>` |
| Legacy per-task commits | `/myflow-do <name> commit-during-apply` |
