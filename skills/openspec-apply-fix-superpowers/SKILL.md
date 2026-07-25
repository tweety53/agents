---
name: openspec-apply-fix-superpowers
description: Apply a fix for something found during manual review (Gate B), manual test (Gate C), or PR review (Gate D) of an already-applied OpenSpec change. Documents the fix in the change's proposal first (append or nested sub-change) so proposals never go stale after review/test rounds, then runs Superpowers #4–#6 in the existing apply worktree. Stage-only (stage + no commit) at Gate B/C; PR-fix (commit + push to the PR branch) at Gate D — the only place in myflow this command commits. Use for /myflow-do-fix.
allowed-tools: Bash(openspec:*)
license: MIT
compatibility: Requires openspec CLI and Superpowers plugin skills.
metadata:
  author: gymie
  version: "1.3"
---

Apply a **fix** for a problem found during **manual review (Gate B)**, **manual test (Gate C)**, or **PR review (Gate D)** of a change that has already been through `/myflow-do`. Unlike `/myflow-do`, this skill never creates a new worktree — it resumes the existing apply worktree/branch — but it always **documents the fix in OpenSpec artifacts first**, so the proposal/tasks never drift out of sync with what was actually built (the source of "stale proposal" drift this skill exists to prevent). The incoming stage — one of the six accepted origins (`awaiting-do-review`, `do-review-started`, `do-done`, `awaiting-manual-test`, `manual-test-done`, `awaiting-pr-review`) — is recorded as **`originStage`** on entry and determines the git mode — see **step 0**. This command always ends at **`awaiting-fix-review`**; `/myflow-do-fix-done` later reads `originStage` and returns the change to the right stage per **Fix re-entry**.

**Announce at start:** "Using openspec-apply-fix-superpowers for change `<name>`."

Also follow **rules/myflow-manual-review.mdc** (Cursor: `.cursor/rules/myflow-manual-review.mdc`).

## When to use this vs `/myflow-do`

- Change has **already completed** `/myflow-do` (worktree exists, tasks mostly/fully checked) and something needs fixing because of a Gate B, Gate C, or Gate D finding → use `/myflow-do-fix`.
- Change has never been applied, or you're resuming interrupted/unchecked **original** tasks (not a review/test finding) → use `/myflow-do`.

## Superpowers Basic Workflow (this stage)

| Step | Skill | When |
|------|-------|------|
| 0 | Stage gate + mode selection | First — requires one of the six origins (`awaiting-do-review`, `do-review-started`, `do-done`, `awaiting-manual-test`, `manual-test-done`, `awaiting-pr-review`); records `originStage`; derives stage-only vs PR-fix from `originStage` |
| — | Fix-organization choice | Before any edit — append vs nested sub-change (below) |
| **4** | **subagent-driven-development** | Execute fix task(s) |
| **5** | **test-driven-development** | Every implementer subagent, every fix task |
| **6** | **requesting-code-review** + **strict review panel** | Full re-run after the fix, same bar as apply |
| — | Stage/commit and hand off | Stage-only when `originStage != awaiting-pr-review`; **commit + push** when `originStage == awaiting-pr-review` (PR-fix) |

No new worktree (**#2** already done by the original apply), no branch finishing (**#7** — merging stays Gate D, done by the human). The only exception: PR-fix mode (`originStage == awaiting-pr-review`) commits and pushes to the existing PR branch — see step 0.

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

### 0. Check stage and select mode

Requires one of the six accepted origins — **`awaiting-do-review`**, **`do-review-started`**, **`do-done`**, **`awaiting-manual-test`**, **`manual-test-done`**, **`awaiting-pr-review`** — per **Stage transitions** in `rules/myflow-manual-review.mdc`. At `proposal-done`, stop and recommend `/myflow-do <name>`. At `finished`, stop — the change is archived.

`do-done` and `manual-test-done` are accepted because both mean "the work is complete and the human confirmed it, but something was found before the next command ran". Accepting `manual-test-done` in particular is what lets `/myflow-review`'s coverage check recommend `/myflow-do-fix` without a stage-mismatch override.

**Record `originStage` on entry.** Read the state file's current `stage` and write it into `originStage` before doing anything else — this is what `/myflow-do-fix-done` later reads to return the change to the right place per **Fix re-entry**. A second fix round overwrites `originStage` with the stage that round began from.

`originStage` selects the git mode. **Do not ask** — derive it:

| `originStage` | Mode | Git behavior at the end |
|----------------|------|-------------------------|
| `awaiting-do-review` | stage-only | `git add -A`; **no commit** |
| `do-review-started` | stage-only | `git add -A`; **no commit** |
| `do-done` | stage-only | `git add -A`; **no commit** |
| `awaiting-manual-test` | stage-only | `git add -A`; **no commit** |
| `manual-test-done` | stage-only | `git add -A`; **no commit** |
| `awaiting-pr-review` (Gate D) | **PR-fix** | `git add -A`, **commit, and push to the PR branch** |

**This command always advances `stage` to `awaiting-fix-review`** at the end (see step 7) — it does not preserve or return to the incoming stage itself. Returning to the right stage is `/myflow-do-fix-done`'s job, driven by the `originStage` this step records. The commit/push decision above is keyed on `originStage`, **not** on the current stage (which is `awaiting-fix-review` throughout this run) — that is what keeps the Gate D commit-and-push behavior working correctly.

#### PR-fix mode (`originStage == awaiting-pr-review`)

A PR already exists remotely, so staging alone would leave the fix invisible to the reviewer. This is the **only** myflow stage where `/myflow-do-fix` commits:

```bash
cd <worktree>
git add -A
git commit -m "fix: <what the PR review surfaced>"
git push
command -v gh >/dev/null 2>&1 && gh pr view --json url -q .url   # GitHub only; skip elsewhere
```

Constraints:

- Commit **after** the full strict review panel passes, never before.
- **Never** merge, never force-push, never amend an already-pushed commit — a reviewer may have commented on it.
- Reply with the PR URL (if `gh` is unavailable for this forge, reply with the branch name and ask the user to open their PR) and tell the user to re-review, then merge.

On mismatch (any other stage), emit the standard mismatch handoff and AskUserQuestion override (default: **No — run the suggested command instead**), per **Stage transitions**.

### 1. Resolve the target change and locate its worktree

```bash
openspec status --change "<name>" --json
```

- **If no name given:** run `openspec list --json`, filter to changes that already have an apply worktree (not yet archived). Exactly one match → use it automatically, announce which; multiple matches → **AskUserQuestion** listing each (name, status, last modified) — do not guess; zero matches → stop, suggest `/myflow-do <name>` first.
- Confirm an apply worktree already exists for this change (from `.superpowers/sdd/progress*.md` or `git worktree list`, branch `openspec/<name>`). If none exists: **stop** — this would be a first apply; use `/myflow-do <name>` instead.
- Load `changeRoot`, current `proposal.md`, `tasks.md`, delta specs, `MERGE_BASE` (from the progress ledger).

### 2. Identify which gate the fix addresses

`originStage` recorded in step 0 already identifies the gate — no need to ask:

- `awaiting-do-review`, `do-review-started`, or `do-done` → **Gate B** (manual review — code)
- `awaiting-manual-test` or `manual-test-done` → **Gate C** (manual test — running app)
- `awaiting-pr-review` → **Gate D** (PR review)

This decides the section title used in step 4 (`Manual Review Fixes` / `Manual Test Fixes` / `PR Review Fixes`) and which gate `/myflow-do-fix-done` hands back to (via **Fix re-entry**).

### 3. Ask how to organize the fix

**AskUserQuestion** — always ask, never default silently (first option is recommended):

1. **Append to this change's proposal** (Recommended) — adds a `Manual Review Fixes` / `Manual Test Fixes` / `PR Review Fixes` section to this change's `proposal.md`, plus matching tasks in `tasks.md`. Simplest; one proposal per change.
2. **New nested sub-change** — creates a separate OpenSpec change (`<name>-fix-N`) that references this change as its parent, implemented in the **same** worktree/branch, and archived together with the parent later.

### 4a. Append path

- **proposal.md**: append (never replace existing content) a `## Manual Review Fixes`, `## Manual Test Fixes`, or `## PR Review Fixes` section. If that heading already exists from an earlier fix round, do **not** duplicate it — add a dated sub-entry under it instead (`### Fix — YYYY-MM-DD`). Each entry: what was found, why, what changes. Keep it terse — this documents *why*, not a diff. Otherwise follow standard OpenSpec proposal formatting (matches the rest of the file).
- **tasks.md**: append matching checkbox task(s) under a `## Manual Review Fixes` / `## Manual Test Fixes` / `## PR Review Fixes` heading (create once, append entries on repeat rounds), meeting **writing-plans** quality — exact file paths, verification commands, no placeholders.
- **specs**: only touch delta spec files if the fix changes *intended* behavior (not just restores it to match the existing spec). Most Gate B/C/D fixes correct an implementation defect against an already-correct spec and need no spec edit. If the fix reveals genuinely new/changed intended behavior, add an `ADDED`/`MODIFIED` requirement the same way `/myflow-start` would.
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

### 7. Stage/commit and hand off (not archive)

Branch on the mode selected in **step 0** (derived from `originStage`). In both branches, write the state file with **`stage: awaiting-fix-review`** and **`originStage`** set to the value recorded in step 0; `updatedAt` and `updatedBy` (`"/myflow-do-fix"`) also change; every other gate value is carried forward exactly as read (gates are monotonic). Resolve its path per **State file** in `rules/myflow-manual-review.mdc` (`--git-common-dir` → `<project-key>` → `/Users/tweety53/Agents/myflow/state/<project-key>/<name>.json`). It lives outside the repo — **never stage, commit, or push it.**

**Stage-only mode** (`originStage` is `awaiting-do-review`, `do-review-started`, `do-done`, `awaiting-manual-test`, or `manual-test-done`) — same as `openspec-apply-superpowers` step 7, in every affected repo/worktree:

```bash
cd <worktree-or-repo>
git add -A
git status
git diff --cached --stat
```

Confirm the fix appears under **Changes to be committed** (staged), then stop. **No commit.**

```
## Fix Applied — Manual Review Required Again

**Change:** <name> (fix for: Manual Review | Manual Test)
**Tracked as:** appended to proposal.md | openspec/changes/<name>-fix-N (nested, parent: <name>)
**Basic Workflow:** #4 ✓ #5 ✓ #6 ✓ (strict panel re-run clean)
**Fix tasks:** N/N complete
**Branch / worktree:** <same as original apply — unchanged>
**Git state:** staged + uncommitted (not pushed)
**Stage:** awaiting-fix-review (originStage: <awaiting-do-review | do-review-started | do-done | awaiting-manual-test | manual-test-done>)

**Open in IntelliJ:**
open -na "IntelliJ IDEA" --args "<absolute worktree path>"

**Next steps:**
- `/myflow-do-fix-manual-review <name>` to mark the fix review as started, or `/myflow-do-fix-done <name>` once the fix looks right — that returns the change to `<originStage>` per **Fix re-entry**
- If this was a Gate B fix: manual review again on the updated staged diff, then continue to `/myflow-manual-test <name>` (Gate C)
- If this was a Gate C fix: refresh the guide first — `/myflow-manual-test <name>` — since prior checked items may need re-verification, then re-test
- More fixes: run `/myflow-do-fix <name>` **after** `/myflow-do-fix-done <name>` has returned the change to `<originStage>` — `/myflow-do-fix` does not accept `awaiting-fix-review` or `fix-review-started`, so starting another round from here would be a stage mismatch
- Once Gate B and Gate C are both satisfied: `/myflow-review <name>` (coverage check, tests, commit + push + open PR — never merges), then a human reviews and merges the PR (Gate D), then `/myflow-finish <name>` (also archives any `<name>-fix-N` nested changes together)
```

**Refreshing the guide after a Gate C fix:** once `/myflow-do-fix-done <name>` returns the change to `awaiting-manual-test`, run `/myflow-manual-test <name>`. It accepts `awaiting-manual-test` as a first-class stage and enters **refresh mode** automatically — no stage mismatch, no override prompt. It preserves checked boxes, does not re-ask the skip question, and re-emits `stage: awaiting-manual-test` with all gates carried forward.

**PR-fix mode** (`originStage == awaiting-pr-review`) — only after the full strict review panel (step 6) is clean, per **PR-fix mode** in step 0:

```bash
cd <worktree>
git add -A
git commit -m "fix: <what the PR review surfaced>"
git push
command -v gh >/dev/null 2>&1 && gh pr view --json url -q .url   # GitHub only; skip elsewhere
```

Never merge, force-push, or amend. Stop after pushing.

```
## Fix Applied — Pushed to PR

**Change:** <name> (fix for: PR review, Gate D)
**Tracked as:** appended to proposal.md | openspec/changes/<name>-fix-N (nested, parent: <name>)
**Basic Workflow:** #4 ✓ #5 ✓ #6 ✓ (strict panel re-run clean)
**Fix tasks:** N/N complete
**Branch / worktree:** <same as original apply — unchanged>
**Git state:** committed and pushed to the PR branch
**Stage:** awaiting-fix-review (originStage: awaiting-pr-review)
**PR:** <PR URL>

**Open in IntelliJ:**
open -na "IntelliJ IDEA" --args "<absolute worktree path>"

**Next steps:**
- `/myflow-do-fix-manual-review <name>` to mark the fix review as started, or `/myflow-do-fix-done <name>` once reviewed — that returns the change to `awaiting-pr-review`
- Ask the reviewer to re-review the PR, then merge it on the forge (never merged by this skill)
- More fixes: run `/myflow-do-fix <name>` **after** `/myflow-do-fix-done <name>` has returned the change to `awaiting-pr-review` — `/myflow-do-fix` does not accept `awaiting-fix-review` or `fix-review-started`, so starting another round from here would be a stage mismatch
- Once merged: `/myflow-finish <name>`
```

## Guardrails

- **Never** create a new worktree or branch — this always resumes the existing apply worktree.
- **Never** run #7 (`finishing-a-development-branch`) or open/merge a PR — merging is always Gate D, done by the human.
- **Always** document the fix in OpenSpec artifacts (append or nested) **before** implementing — this is the whole point of this skill over ad hoc reuse of `/myflow-do`.
- **Never** duplicate a `## Manual Review Fixes` / `## Manual Test Fixes` / `## PR Review Fixes` heading in `proposal.md` or `tasks.md` — append a dated sub-entry to the existing one.
- **Never** archive a nested `<name>-fix-N` change on its own; it archives only together with its parent (see `openspec-archive-superpowers`).
- **Always** re-run the **full** strict review panel (not a fix-only partial review) before handoff.
- Do not skip TDD on fix tasks.
- Pause on ambiguity (which gate, how to organize) — never guess silently past the AskUserQuestion prompts in steps 2–3.
- **Always record `originStage`** on entry (step 0) and write it into the state file — `/myflow-do-fix-done` depends on it to return the change to the right stage.
- **Always end at `stage: awaiting-fix-review`** — this command never writes any other stage itself; returning the change to `originStage` is `/myflow-do-fix-done`'s job.
- **Never commit when `originStage != awaiting-pr-review`** — stage only, exactly as `/myflow-do`.
- **Only PR-fix mode commits** (`originStage == awaiting-pr-review`), and only after the full review panel passes. Key this on `originStage`, never on the current stage.
- **Never force-push or amend** a commit that is already on an open PR.
- **Never point the user at `/myflow-do-fix` while the change sits at `awaiting-fix-review` or `fix-review-started`.** Those are not accepted origins; another round starts only after `/myflow-do-fix-done` has returned the change to its `originStage`. Recording `originStage: awaiting-fix-review` via an override would make `/myflow-do-fix-done` target the stage it just accepted — a livelock.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Fix something found in manual review, manual test, or PR review | `/myflow-do-fix <name>` |
| Original (first) apply | `/myflow-do <name>` |
| After fix, re-test | `/myflow-manual-test <name>` |
| After both gates satisfied | `/myflow-review <name>` (commit + push + open PR) then, after the PR merges (Gate D), `/myflow-finish <name>` |
