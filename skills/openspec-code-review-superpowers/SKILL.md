---
name: openspec-code-review-superpowers
description: Code review stage after Gate B/C — checks test coverage, verifies tests/linters, commits apply work, runs Basic Workflow #7 (finishing-a-development-branch). Use for /myflow-code-review.
allowed-tools: Bash(openspec:*)
license: MIT
compatibility: Requires openspec CLI and Superpowers plugin skills.
metadata:
  author: gymie
  version: "1.0"
---

Run the **code review** stage after manual review (Gate B) and manual test (Gate C): verify test coverage and quality, commit the apply work, and integrate via Basic Workflow **#7** (finishing-a-development-branch). This is the last stage before `/myflow-finish` archives the change.

**Announce at start:** "Using openspec-code-review-superpowers for change `<name>`."

Also follow **rules/myflow-manual-review.mdc** (Cursor: `.cursor/rules/myflow-manual-review.mdc`).

## Superpowers Basic Workflow (this stage)

| Step | Skill | When |
|------|-------|------|
| — | Gate C completion check | First — verify manual-test checklist (or `SKIPPED` marker) |
| — | Test coverage check | Before verification — flag gaps, route to `/myflow-do-fix` |
| — | **verification-before-completion** | Tests/linters green |
| — | **git commit** | After verification passes |
| **7** | **finishing-a-development-branch** | **Always** — merge / PR / push per user choice |

Steps **#1–#6** completed in the start/do/do-fix stages. **#7 always runs here** — this stage owns commit + integrate, split out of the old combined archive stage.

## Required sub-skills

1. **superpowers:verification-before-completion** — confirm tests/checks pass with evidence.
2. **superpowers:finishing-a-development-branch** — Basic Workflow **#7**.

## Workflow

### 1. Select change and locate apply work

- Use the provided name. **If omitted:** run `openspec list --json`, filter to changes with an apply worktree that have finished Gate B/C (not yet code-reviewed). Exactly one match → use it automatically, announce which; multiple matches → **AskUserQuestion** listing each (name, status, last modified) — do not guess; zero matches → stop, suggest `/myflow-do <name>`.
- Announce: "Code review for change: `<name>`."
- Locate the apply worktree/branch (`openspec/<name>` or path from progress ledger).
- Confirm implementation is present (staged/uncommitted changes from apply and any `/myflow-do-fix` rounds).

```bash
cd <worktree>
git status
git diff --cached --stat
git diff --stat <MERGE_BASE>
```

If worktree missing or no changes: **stop** — suggest `/myflow-do <name>`.

**Check for nested fix sub-changes** (created by `/myflow-do-fix` when the user chose "nested sub-change"): `openspec list --json`, look for `<name>-fix-*`. Their code already lives in this same worktree's diff (implemented in-place) — nothing extra to do here; note them in the summary. They get archived alongside `<name>` at `/myflow-finish`.

### 2. Verify Gate C completion

Read `docs/manual-test/<name>.md`:

- **Missing entirely**: warn once and offer — generate via `/myflow-manual-test <name>` first, or proceed without it if the user explicitly skipped Gate C.
- **Contains `**Manual test status:** SKIPPED`**: Gate C was intentionally bypassed (via `/myflow-manual-test-skip`). Note it in the summary and continue — not a defect.
- **No marker**: parse every `- [ ]` / `- [x]` line in the functionality checklist and sign-off sections. If **any** remain unchecked, **notify** the user with the count and a short list of what's open, then **AskUserQuestion**: continue anyway, go finish testing, or fix via `/myflow-do-fix <name>`. Never proceed past this silently.

### 3. Test coverage check

Before running verification, assess whether the change's tests actually cover what was built — this is a **quality gate**, not just "tests pass":

- Walk every **delta spec** Scenario for `<name>` (and any nested `<name>-fix-N` changes) and confirm a corresponding test exists — check test file names/content, not just that *some* test changed in the diff.
- Cross-check `tasks.md`: every implementation task should have traceable RED-GREEN-REFACTOR evidence in the progress ledger (per the TDD requirement enforced during apply/do-fix).
- If the project has a coverage tool wired up, run it and note the delta for touched modules; otherwise rely on the scenario-by-scenario check above.
- **If coverage looks thin** (a scenario with no matching test, or a task with no test evidence): report the specific gaps and **AskUserQuestion** — run `/myflow-do-fix <name>` to add the missing tests (**recommended**), or proceed anyway. **Do not write the missing tests yourself in this skill** — routing through `/myflow-do-fix` keeps the gap documented in the proposal, same as any other fix.

### 4. Verification gate

Invoke **superpowers:verification-before-completion**:

- Run project-appropriate tests (e.g. `./gradlew test`, `./gradlew ktlintCheck detekt`, frontend compile checks) in every affected repo/worktree.
- Show command output before claiming ready.
- If tests fail: **stop** — fix in worktree and re-verify, or suggest `/myflow-do-fix <name>` for larger fixes.

### 5. Commit apply work

**First commit in this stage** unless user asks for multiple commits.

- Prefer committing what apply/do-fix already staged. If anything is still unstaged, `git add -A` then commit.
- Write a concise commit message reflecting the change purpose (why, not just what).
- Commit per user git rules (hooks must pass).
- If `commit-during-apply` was used and changes are already committed: skip this step; confirm `git log MERGE_BASE..HEAD` shows expected commits.

### 6. Basic Workflow #7 — Branch finishing (always)

Invoke **superpowers:finishing-a-development-branch**:

- Present merge / PR / keep branch / discard options.
- Execute the user's choice.
- Detached HEAD / external worktree: follow that skill's reduced menu.

**Do not skip #7** in standard myflow.

### 7. Summary

```
## Code Review Complete

**Change:** <name>
**Nested fixes included:** <name>-fix-1, <name>-fix-2, ... | none
**Manual test (Gate C):** all items checked | SKIPPED (intentional) | proceeded with N unchecked (user override)
**Test coverage:** all scenarios covered | proceeded with N gap(s) (user override) — see notes
**Committed:** ✓ (this stage)
**Tests/linters:** ✓ (commands run)
**Basic Workflow #7:** ✓ <merged | PR url | deferred | n/a>

**Next:** `/myflow-finish <name>` — verify merged, sync specs, archive.
```

## Guardrails

- **Always** check Gate C completion before anything else.
- **Always** run the test-coverage check before verification — never silently skip it.
- **Never** write missing tests yourself here — flag them and route to `/myflow-do-fix <name>`.
- **Always commit** uncommitted apply work before verification (unless already committed via `commit-during-apply`).
- Never mark verification complete with failing tests unless the user explicitly overrides after seeing failures.
- **Always run #7** — never assume a prior stage finished integration.
- Do not force-push or amend unless user rules allow.
- Do not archive here — that is `/myflow-finish`'s job.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Run code review (coverage, tests, commit, #7) | `/myflow-code-review <name>` |
| Fix a coverage gap or test/lint failure | `/myflow-do-fix <name>` |
| After code review | `/myflow-finish <name>` |
