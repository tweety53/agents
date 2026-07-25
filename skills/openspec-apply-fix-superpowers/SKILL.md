---
name: openspec-apply-fix-superpowers
description: Apply a fix for something found during manual review (Gate B) or manual test (Gate C) of an already-applied OpenSpec change. Documents the fix in the change's proposal first (append or nested sub-change) so proposals never go stale after review/test rounds, then runs Superpowers #4–#6 in the existing apply worktree. Use for /myflow-do-fix.
allowed-tools: Bash(openspec:*)
license: MIT
compatibility: Requires openspec CLI and Superpowers plugin skills.
metadata:
  author: gymie
  version: "1.2"
---

Apply a **fix** for a problem found during **manual review (Gate B)** or **manual test (Gate C)** of a change that has already been through `/myflow-do`. Unlike `/myflow-do`, this skill never creates a new worktree — it resumes the existing apply worktree/branch — but it always **documents the fix in OpenSpec artifacts first**, so the proposal/tasks never drift out of sync with what was actually built (the source of "stale proposal" drift this skill exists to prevent).

**Announce at start:** "Using openspec-apply-fix-superpowers for change `<name>`."

Also follow **rules/myflow-manual-review.mdc** (Cursor: `.cursor/rules/myflow-manual-review.mdc`).

## When to use this vs `/myflow-do`

- Change has **already completed** `/myflow-do` (worktree exists, tasks mostly/fully checked) and something needs fixing because of a Gate B or Gate C finding → use `/myflow-do-fix`.
- Change has never been applied, or you're resuming interrupted/unchecked **original** tasks (not a review/test finding) → use `/myflow-do`.

## Superpowers Basic Workflow (this stage)

| Step | Skill | When |
|------|-------|------|
| — | Fix-organization choice | Before any edit — append vs nested sub-change (below) |
| **4** | **subagent-driven-development** | Execute fix task(s) |
| **5** | **test-driven-development** | Every implementer subagent, every fix task |
| **6** | **requesting-code-review** + **strict review panel** | Full re-run after the fix, same bar as apply |

No new worktree (**#2** already done by the original apply), no branch finishing (**#7** deferred to archive) — same constraints as apply.

## Required sub-skills

Same as **openspec-apply-superpowers** steps 4–6 (reuse [../openspec-apply-superpowers/](../openspec-apply-superpowers/) prompt files verbatim — do not fork copies):

1. **superpowers:subagent-driven-development** — Basic Workflow **#4**, with the no-commit override below.
2. **superpowers:test-driven-development** — Basic Workflow **#5**, mandatory per implementer dispatch.
3. **superpowers:requesting-code-review** — Basic Workflow **#6** primary reviewer.
4. **Strict review panel (5 additional agents)** — same roster/prompts/economic-model mapping as apply; see `../openspec-apply-superpowers/SKILL.md`.
5. **superpowers:verification-before-completion** — evidence before re-handoff.

Plus, for documenting the fix:

6. **openspec-update-change** conventions — confirm-before-write discipline for the append path.
7. **openspec-propose** conventions — artifact shape for the nested sub-change path.

**Do not invoke** `superpowers:using-git-worktrees` (**#2**) or `finishing-a-development-branch` (**#7**) in this stage.

## No-commit SDD override (mandatory)

Identical to apply — every implementer dispatch **must** include:

> **MYFLOW APPLY-FIX — NO COMMITS:** Do **not** run `git commit`, `git push`, merge, or open a PR. Leave all changes uncommitted in the worktree. You **may** `git add` / stage files. The parent agent records `TASK_BASE=$(git rev-parse HEAD)` before dispatch; your diff for review is `git diff TASK_BASE` (plus `git diff --cached TASK_BASE` if you staged files). The parent will `git add -A` before handoff.

And the TDD requirement:

> **REQUIRED SUB-SKILL:** Use superpowers:test-driven-development — RED-GREEN-REFACTOR for this task. Delete any code written before tests.

## Workflow

### 1. Resolve the target change and locate its worktree

```bash
openspec status --change "<name>" --json
```

- **If no name given:** run `openspec list --json`, filter to changes that already have an apply worktree (not yet archived). Exactly one match → use it automatically, announce which; multiple matches → **AskUserQuestion** listing each (name, status, last modified) — do not guess; zero matches → stop, suggest `/myflow-do <name>` first.
- Confirm an apply worktree already exists for this change (from `.superpowers/sdd/progress*.md` or `git worktree list`, branch `openspec/<name>`). If none exists: **stop** — this would be a first apply; use `/myflow-do <name>` instead.
- Load `changeRoot`, current `proposal.md`, `tasks.md`, delta specs, `MERGE_BASE` (from the progress ledger).

### 2. Identify which gate the fix addresses

Infer from conversation context (which gate the user is currently at); if unclear, **AskUserQuestion**:

> Is this fix for something found in **manual review (Gate B — code)** or **manual test (Gate C — running app)**?

This decides the section title used in step 4 (`Manual Review Fixes` vs `Manual Test Fixes`) and which gate to hand back to in step 7.

### 3. Ask how to organize the fix

**AskUserQuestion** — always ask, never default silently (first option is recommended):

1. **Append to this change's proposal** (Recommended) — adds a `Manual Review Fixes` / `Manual Test Fixes` section to this change's `proposal.md`, plus matching tasks in `tasks.md`. Simplest; one proposal per change.
2. **New nested sub-change** — creates a separate OpenSpec change (`<name>-fix-N`) that references this change as its parent, implemented in the **same** worktree/branch, and archived together with the parent later.

### 4a. Append path

- **proposal.md**: append (never replace existing content) a `## Manual Review Fixes` or `## Manual Test Fixes` section. If that heading already exists from an earlier fix round, do **not** duplicate it — add a dated sub-entry under it instead (`### Fix — YYYY-MM-DD`). Each entry: what was found, why, what changes. Keep it terse — this documents *why*, not a diff. Otherwise follow standard OpenSpec proposal formatting (matches the rest of the file).
- **tasks.md**: append matching checkbox task(s) under a `## Manual Review Fixes` / `## Manual Test Fixes` heading (create once, append entries on repeat rounds), meeting **writing-plans** quality — exact file paths, verification commands, no placeholders.
- **specs**: only touch delta spec files if the fix changes *intended* behavior (not just restores it to match the existing spec). Most Gate B/C fixes correct an implementation defect against an already-correct spec and need no spec edit. If the fix reveals genuinely new/changed intended behavior, add an `ADDED`/`MODIFIED` requirement the same way `/myflow-start` would.
- Show the user the proposed additions before writing (same confirm-before-write discipline as `openspec-update-change`).

### 4b. Nested sub-change path

- Determine the next fix number: list `openspec/changes/<name>-fix-*` and `openspec/changes/archive/*-<name>-fix-*`; use the next unused N.
- `openspec new change "<name>-fix-N"` for a minimal proposal: `Why` / `What Changes` / `Impact`, plus `tasks.md`. The **first line** of `proposal.md` **must** read `**Parent change:** \`<name>\`` so the link is unambiguous even outside this repo's own bookkeeping.
- In the **parent** change's `proposal.md`, append/maintain a `## Related Fixes` section listing every `<name>-fix-N` (a live index — this is what keeps the parent from going stale even though the fix content lives in separate files).
- Implementation still happens in the **same** apply worktree/branch as the parent — do **not** create a new worktree or branch for the fix.
- The nested change **must not** be archived on its own; see the archive-time guardrail in `openspec-archive-superpowers`.

### 5. Implement the fix (Basic Workflow #4–#6)

Same discipline as `openspec-apply-superpowers` steps 4–6:

- Invoke **superpowers:subagent-driven-development** with the no-commit override and TDD requirement above.
- One SDD task per fix item from step 4a/4b's `tasks.md` entries.
- After each implementer returns: per-task review via `git diff TASK_BASE`, same pattern as apply.
- Mark fix task checkboxes `[x]` only after task review passes.
- Progress ledger: append to the same `.superpowers/sdd/progress-<name>.md` used by the original apply (do not start a new ledger file).

### 6. Full strict review panel — always re-run in full

Do **not** review only the fix diff in isolation. Re-run the **entire final whole-branch strict review panel** (primary + Bugbot + Security + Adversarial + Senior + Economic Senior — same six agents, same prompts, same economic-model mapping as `openspec-apply-superpowers`) against `git diff MERGE_BASE`, exactly as a normal apply handoff. A fix can introduce regressions elsewhere; the full panel is what makes handoff to another Gate B trustworthy.

Fix any new Critical/Important findings, then re-run the full panel again until clean — identical aggregation rules to apply.

### 7. Stage and hand off (not archive)

Same as `openspec-apply-superpowers` step 7 — in every affected repo/worktree:

```bash
cd <worktree-or-repo>
git add -A
git status
git diff --cached --stat
```

Confirm the fix appears under **Changes to be committed** (staged), then stop.

```
## Fix Applied — Manual Review Required Again

**Change:** <name> (fix for: Manual Review | Manual Test)
**Tracked as:** appended to proposal.md | openspec/changes/<name>-fix-N (nested, parent: <name>)
**Basic Workflow:** #4 ✓ #5 ✓ #6 ✓ (strict panel re-run clean)
**Fix tasks:** N/N complete
**Branch / worktree:** <same as original apply — unchanged>
**Git state:** staged + uncommitted (not pushed)

**Next steps:**
- If this was a Gate B fix: manual review again on the updated staged diff, then continue to `/myflow-manual-test <name>` (Gate C)
- If this was a Gate C fix: refresh the guide first — `/myflow-manual-test <name>` — since prior checked items may need re-verification, then re-test
- More fixes: `/myflow-do-fix <name>` again
- Once Gate B and Gate C are both satisfied: `/myflow-code-review <name>` (coverage check, tests, commit, #7), then `/myflow-finish <name>` (also archives any `<name>-fix-N` nested changes together)
```

## Guardrails

- **Never** create a new worktree or branch — this always resumes the existing apply worktree.
- **Never** commit, push, merge, or run #7 — same as apply (commits happen in `/myflow-code-review`).
- **Always** document the fix in OpenSpec artifacts (append or nested) **before** implementing — this is the whole point of this skill over ad hoc reuse of `/myflow-do`.
- **Never** duplicate a `## Manual Review Fixes` / `## Manual Test Fixes` heading in `proposal.md` or `tasks.md` — append a dated sub-entry to the existing one.
- **Never** archive a nested `<name>-fix-N` change on its own; it archives only together with its parent (see `openspec-archive-superpowers`).
- **Always** re-run the **full** strict review panel (not a fix-only partial review) before handoff.
- Do not skip TDD on fix tasks.
- Pause on ambiguity (which gate, how to organize) — never guess silently past the AskUserQuestion prompts in steps 2–3.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Fix something found in manual review or manual test | `/myflow-do-fix <name>` |
| Original (first) apply | `/myflow-do <name>` |
| After fix, re-test | `/myflow-manual-test <name>` |
| After both gates satisfied | `/myflow-code-review <name>` then `/myflow-finish <name>` |
