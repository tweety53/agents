---
name: openspec-archive-superpowers
description: Finish stage — validate the branch is merged into main/develop, sync delta specs into main specs, archive the OpenSpec change. Use for /myflow-finish.
allowed-tools: Bash(openspec:*)
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: gymie
  version: "4.0"
---

Finish an OpenSpec change **after code review (`/myflow-code-review`)**: validate the branch actually landed on `main`/`develop`, sync delta specs into main specs, and archive the change. This is the last stage in myflow.

**Announce at start:** "Using openspec-archive-superpowers for change `<name>`."

Also follow **rules/myflow-manual-review.mdc** (Cursor: `.cursor/rules/myflow-manual-review.mdc`).

## Pipeline position

```text
... → manual test (Gate C) → code review (commit + #7) → finish (this skill)
```

Commit, verification, and Basic Workflow **#7** (branch finishing) already happened in `/myflow-code-review`. This skill assumes that already happened and **verifies** it, rather than doing it — it never commits, tests, or merges anything itself.

## Required sub-skills

1. **openspec-archive-change** — delta sync assessment + move to archive.

Optional: **openspec-sync-specs** when the user chooses to sync delta specs to main.

## Workflow

### 1. Select change and validate it's ready to finish

- Use the provided name. **If omitted:** run `openspec list --json`, filter to changes that have gone through code review (committed, no longer just staged-and-uncommitted). Exactly one match → use it automatically, announce which; multiple matches → **AskUserQuestion** listing each (name, status, last modified) — do not guess; zero matches → stop, suggest `/myflow-code-review <name>`.
- Announce: "Finishing change: `<name>`."
- Locate the apply worktree/branch (`openspec/<name>` or path from progress ledger).

**Validate the branch is actually merged into `main`/`develop`** — this replaces the inline #7 step that used to live here:

```bash
cd <main-repo-checkout>
git fetch origin
git merge-base --is-ancestor openspec/<name> <base-branch> && echo "merged"
```

- **Merged** (command prints "merged", or the PR shows merged on the forge): continue.
- **Open PR, not yet merged**: **stop** — this isn't ready to finish. Wait for the PR to merge, or ask the user how they want to proceed.
- **Branch kept as-is** (user chose "keep branch" at `/myflow-code-review`'s #7 step, deliberately deferring integration): **AskUserQuestion** — finishing now will archive the OpenSpec change while the code isn't on `<base-branch>` yet; confirm this is intentional before proceeding, or stop and suggest finishing #7 first via `/myflow-code-review <name>`.
- **No commits found at all** (evidence #7 never ran): **stop** — suggest `/myflow-code-review <name>` first.

### 2. OpenSpec pre-archive checks

Follow **openspec-archive-change** steps 2–4:

```bash
openspec status --change "<name>" --json
```

- Warn on incomplete artifacts or unchecked tasks; **AskUserQuestion** to proceed.
- Assess delta spec sync vs `openspec/specs/`; show summary.
- Offer: sync now (recommended), archive without sync, cancel.

If user chooses sync: invoke **openspec-sync-specs** for this change.

**Check for nested fix sub-changes** (created by `/myflow-do-fix` when the user chose "nested sub-change"):

```bash
openspec list --json   # look for names matching "<name>-fix-*"
```

If any `<name>-fix-N` changes exist and are not yet archived, they **must** be archived together with `<name>` in this same operation — a nested fix is never left archived alone or skipped. Include them in spec sync and the archive move alongside the parent.

### 3. Archive

Follow **openspec-archive-change** step 5:

- Target: `<planningHome.changesDir>/archive/YYYY-MM-DD-<name>`
- Fail if target exists; otherwise move `changeRoot`.
- **If nested `<name>-fix-N` sub-changes were found in step 2**, archive each of them the same way, in the same session, right alongside `<name>` (same date). Do not stop after archiving the parent while a nested fix remains unarchived.

### 4. Summary

```
## Finish Complete

**Change:** <name>
**Nested fixes archived:** <name>-fix-1, <name>-fix-2, ... | none
**Integration verified:** merged into <base-branch> | PR <url> merged | deferred (user override)
**Archived to:** <path>
**Specs:** <synced | skipped | none>
```

## Guardrails

- **Never commit, run tests, or run #7 here** — that already happened in `/myflow-code-review`; this skill only verifies it landed.
- **Never archive a branch that's still an open, unmerged PR** without explicit user confirmation.
- Always show incomplete task/artifact warnings before archive.
- Preserve `.openspec.yaml` in the archived directory (moves with change).
- **Never archive a nested `<name>-fix-N` sub-change on its own** — if the user names a fix change directly, redirect to archiving its parent (`<name>`), which pulls the fix along with it.
- **Never leave a nested fix unarchived** while its parent gets archived — check for `<name>-fix-*` every time (step 2) and archive them together.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Code review (must run first) | `/myflow-code-review <name>` |
| Finish — verify merged, sync specs, archive | `/myflow-finish <name>` |
