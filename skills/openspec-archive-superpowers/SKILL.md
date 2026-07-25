---
name: openspec-archive-superpowers
description: Finish stage — verify the PR actually merged (Gate D), sync delta specs into main specs, archive the OpenSpec change. Use for /myflow-finish.
allowed-tools: Bash(openspec:*)
license: MIT
compatibility: Requires openspec CLI.
metadata:
  author: gymie
  version: "5.0"
---

Finish an OpenSpec change **after review (`/myflow-review`) and the human's PR review + merge (Gate D)**: verify the PR actually merged, sync delta specs into main specs, and archive the change. This is the last stage in myflow.

**Announce at start:** "Using openspec-archive-superpowers for change `<name>`."

Also follow **rules/myflow-manual-review.mdc** (Cursor: `.cursor/rules/myflow-manual-review.mdc`).

## Pipeline position

```text
... → manual test (Gate C) → review (commit + PR, no merge) → PR review (Gate D, human merges) → finish (this skill)
```

`/myflow-review` committed, pushed, and opened the PR, then deliberately stopped without merging — unless the user passed `automerge`, in which case `/myflow-review` merged directly and there was no PR to review (see **Auto-merge (opt-in)**). Absent `automerge`, merging is a **human** action (Gate D) that happens on the forge, outside myflow, and may simply not have happened yet: an open PR is the normal, expected state when this skill runs, not an anomaly. This skill only **verifies** the PR (or the automerge) actually landed on the base branch, then archives; it never commits, tests, merges, or pushes anything itself.

## Required sub-skills

1. **openspec-archive-change** — delta sync assessment + move to archive.

Optional: **openspec-sync-specs** when the user chooses to sync delta specs to main.

## Workflow

### 0. Check stage

Requires stage **`review-done`** per **Stage transitions** in `rules/myflow-manual-review.mdc`. On mismatch, stop with the standard mismatch handoff and AskUserQuestion override (default: **No**).

- At `manual-test-done` → recommend `/myflow-review <name>` first.
- At `awaiting-pr-review` → recommend `/myflow-review-done <name>` once the PR has actually been reviewed and merged.
- At `finished` → already archived; stop.

**`review-done` is a claim, not proof.** It only means a human confirmed they reviewed (and, outside `automerge`, merged) the PR — it is not independent evidence the merge actually happened. This skill's PR-merge check below is what verifies it; do not treat `stage: review-done` alone as sufficient to archive, and do not let a future edit remove the check just because the two now look redundant.

### 1. Select change and validate it's ready to finish

- Use the provided name. **If omitted:** run `openspec list --json`, filter to changes that have gone through review (committed, no longer just staged-and-uncommitted). Exactly one match → use it automatically, announce which; multiple matches → **AskUserQuestion** listing each (name, status, last modified) — do not guess; zero matches → stop, suggest `/myflow-review <name>`.
- Announce: "Finishing change: `<name>`."
- Locate the apply worktree/branch (`openspec/<name>` or path from progress ledger).

**Verify the PR actually merged** — `/myflow-review` opened it but deliberately did not merge it (Gate D is the human's):

```bash
cd <main-repo-checkout>
git fetch origin
# Ancestry check — works on every forge (Bitbucket, GitLab, GitHub) and needs no PR CLI.
# Run it FIRST and always; it is the primary merge evidence.
git merge-base --is-ancestor origin/openspec/<name> origin/<base-branch> && echo "merged"
# GitHub only, and only when gh is actually installed — extra detail, never a prerequisite:
command -v gh >/dev/null 2>&1 && gh pr list --head openspec/<name> --state all --json number,state,mergedAt,url
```

- **PR state `MERGED`** (or the ancestor check prints "merged"): continue.
- **PR still `OPEN`**: **stop.** This is the normal, expected block — Gate D has not happened yet. Tell the user to review and merge the PR, then re-run. Do not offer to merge it for them.
- **PR `CLOSED` unmerged**: stop and ask what happened — the work was abandoned or superseded.
- **No PR and no merge evidence**: stop — suggest `/myflow-review <name>` first.
- **User insists on archiving unmerged** (deliberate, e.g. the work landed another way): **AskUserQuestion** confirming that archiving without the code on `<base-branch>` is intentional. Default: **No**.

If `gh` is unavailable or the forge is not GitHub (this project's `origin` is Bitbucket), say so honestly and rely on the `git merge-base --is-ancestor` check alone — it is sufficient and must never be gated behind a `gh` call. Never guess `MERGED`. If neither check can be performed at all (no remote, no network), report the PR state as **unknown** and ask the user — do not rewind the stage on an inconclusive probe.

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
- Write the terminal state into the change's **user-scoped** state file — resolve the path per **State file** in `rules/myflow-manual-review.mdc` (`--git-common-dir` → `<project-key>` → `/Users/tweety53/Agents/myflow/state/<project-key>/<name>.json`). It is **not** moved into the archive and **not** committed; it stays at that path as the terminal record:

```json
{
  "stage": "finished",
  "gates": { "reviewed": true, "tested": <as recorded>, "prOpened": <as recorded — false if automerge was used>, "prMerged": true },
  "worktree": null,
  "branch": "openspec/<name>",
  "originStage": null,
  "artifactUrl": "<unchanged — carried forward from the file as read>",
  "updatedAt": "<ISO-8601 UTC now>",
  "updatedBy": "/myflow-finish"
}
```

**`artifactUrl` is carried forward, never dropped** — this is the terminal record, so the published proposal link must survive into it. Writes render the whole object; omitting the field would erase it.

Set `worktree` to `null` — the worktree is removed or stale after finishing. Carry `gates.tested` forward exactly as recorded (`true` or `"skipped"` are sticky — never demote them). Write the same terminal state to each nested `<name>-fix-N` change's own user-scoped state file before archiving it.

### 4. Summary

```
## Finish Complete

**Change:** <name>
**Nested fixes archived:** <name>-fix-1, <name>-fix-2, ... | none
**PR:** <url> — merged <date> | archived unmerged (user override)
**Archived to:** <path>
**Specs:** <synced | skipped | none>
```

## Guardrails

- **Never commit, run tests, merge, or push here** — review opened the PR and the human merged it; this skill only verifies and archives.
- **Never archive while the PR is still open** without an explicit user override.
- **Always write `stage: finished`** into the change (and every nested fix) before the archive move.
- Always show incomplete task/artifact warnings before archive.
- Preserve `.openspec.yaml` in the archived directory (moves with change).
- **Never archive a nested `<name>-fix-N` sub-change on its own** — if the user names a fix change directly, redirect to archiving its parent (`<name>`), which pulls the fix along with it.
- **Never leave a nested fix unarchived** while its parent gets archived — check for `<name>-fix-*` every time (step 2) and archive them together.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Review (must run first) | `/myflow-review <name>` |
| Finish — verify merged, sync specs, archive | `/myflow-finish <name>` |
