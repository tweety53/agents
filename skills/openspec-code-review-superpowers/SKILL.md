---
name: openspec-code-review-superpowers
description: Code review stage after Gate B/C — checks test coverage, verifies tests/linters, commits, pushes, opens PR (never merges). Use for /myflow-code-review.
allowed-tools: Bash(openspec:*)
license: MIT
compatibility: Requires openspec CLI and Superpowers plugin skills.
metadata:
  author: gymie
  version: "1.0"
---

Run the **code review** stage after manual review (Gate B) and manual test (Gate C): verify test coverage and quality, commit the apply work, then push and open a PR via Basic Workflow **#7** (finishing-a-development-branch, constrained to the PR path — never merge). This stage ends at Gate D (human PR review + merge); `/myflow-finish` runs after the human merges.

**Announce at start:** "Using openspec-code-review-superpowers for change `<name>`."

Also follow **rules/myflow-manual-review.mdc** (Cursor: `.cursor/rules/myflow-manual-review.mdc`).

## Superpowers Basic Workflow (this stage)

| Step | Skill | When |
|------|-------|------|
| 0 | Stage gate | First — requires `awaiting-test` |
| — | Gate C completion check | Verify manual-test checklist (or `SKIPPED` marker) |
| — | Test coverage check | Before verification — flag gaps, route to `/myflow-do-fix` |
| — | **verification-before-completion** | Tests/linters green |
| — | **git commit** | After verification passes |
| **7** | **finishing-a-development-branch** | **PR only** — push + open PR; never merge |

Steps **#1–#6** completed in the start/do/do-fix stages. **#7 runs here, constrained to the PR path** — this stage owns commit + push + PR; merging is Gate D, done by the human.

## Required sub-skills

1. **superpowers:verification-before-completion** — confirm tests/checks pass with evidence.
2. **superpowers:finishing-a-development-branch** — Basic Workflow **#7**.

## Workflow

### 0. Check stage

Requires stage **`awaiting-test`** per **Stage transitions** in `rules/myflow-manual-review.mdc`. On mismatch, stop with the standard mismatch handoff and AskUserQuestion override (default: **No**).

- At `awaiting-review` → recommend `/myflow-manual-test <name>` first.
- At `awaiting-pr-review` → the PR is already open; recommend `/myflow-do-fix <name>` for changes, or `/myflow-finish <name>` once merged.

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

Read `gates.tested` from the state file first, then `docs/manual-test/<name>.md`:

- **`gates.tested: "skipped"`** (or the guide contains `**Manual test status:** SKIPPED`): Gate C intentionally bypassed. Note it in the summary and continue — not a defect.
- **`gates.tested: false`**: parse every `- [ ]` / `- [x]` line in the functionality checklist and sign-off sections. If **any** remain unchecked, notify the user with the count and a short list of what is open, then **AskUserQuestion**: continue anyway, go finish testing, or fix via `/myflow-do-fix <name>`. Never proceed past this silently.
- **Guide missing entirely**: warn and offer `/myflow-manual-test <name>` first. (The stage gate in step 0 makes this unlikely.)

**Promote the flag.** `/myflow-code-review` is the **only** writer of `gates.tested: true`. When
`gates.tested` is `false` and every checkbox is ticked, set it to `true` before writing state —
that is what makes `true` mean "testing completed and verified" rather than a value nothing ever
produces. Never overwrite `"skipped"`: an intentional bypass stays `"skipped"` forever.

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

### 6. Push and open a PR — then stop

Invoke **superpowers:finishing-a-development-branch**, but constrain it to the **PR path only**:

- Push the branch to the remote.
- Open a PR against the base branch (`develop` unless the user says otherwise).
- **Do not merge.** Do not offer merge, direct push to base, or discard. Gate D is the human's.

```bash
cd <worktree>
git push -u origin openspec/<name>
gh pr create --base develop --head openspec/<name> --title "<change title>" --body "<summary + link to proposal>"
```

If a PR for this branch already exists (a `/myflow-do-fix` round pushed to it), skip creation and reuse it:

```bash
gh pr list --head openspec/<name> --state open --json number,url
```

If the user's setup has no forge remote, say so and stop at "pushed, no PR opened" — record `gates.prOpened: false` and tell the user to open the PR manually. Do not fall back to merging.

### 6b. Write state

**If step 6 opened (or reused) a PR:**

```json
{
  "stage": "awaiting-pr-review",
  "gates": { "reviewed": true, "tested": <true | "skipped" | false-with-user-override>, "prOpened": true, "prMerged": false },
  "worktree": "<unchanged>",
  "branch": "openspec/<name>",
  "updatedAt": "<ISO-8601 UTC now>",
  "updatedBy": "/myflow-code-review"
}
```

**If step 6 hit the no-forge-remote case (pushed, no PR opened):** do not advance the stage — a PR doesn't exist yet, so `awaiting-pr-review` would be false on write and immediately flagged by the rule file's own self-heal table. Leave the change at `awaiting-test`, and say plainly that it does not advance to Gate D until a PR actually exists:

```json
{
  "stage": "awaiting-test",
  "gates": { "reviewed": true, "tested": <true | "skipped" | false-with-user-override>, "prOpened": false, "prMerged": false },
  "worktree": "<unchanged>",
  "branch": "openspec/<name>",
  "updatedAt": "<ISO-8601 UTC now>",
  "updatedBy": "/myflow-code-review"
}
```

Include this file in the commit made in step 5 if it is written before committing; otherwise commit it in a small follow-up commit — the state file must not be left as the only uncommitted change on a PR branch.

**This project writes it as a follow-up commit.** The state file's `gates.prOpened`/`prMerged` and `stage` values are only known once step 6 has actually pushed and (re)confirmed the PR (or confirmed there is no forge remote), which happens after step 5's commit — so step 6b commits and pushes the state file on its own, after step 6:

```bash
cd <worktree>
git add openspec/changes/<name>/.myflow-state.json
git commit -m "chore(<name>): advance to awaiting-pr-review"
git push
```

(Omit the `git push` only in the no-forge-remote case if there is truly no remote configured at all; if a remote exists but simply has no PR-hosting forge, still push the commit.)

### 7. Summary

```
## Code Review Complete — PR Review Required (Gate D)

**Change:** <name>
**Nested fixes included:** <name>-fix-1, ... | none
**Manual test (Gate C):** all items checked | SKIPPED (intentional) | proceeded with N unchecked (user override)
**Test coverage:** all scenarios covered | proceeded with N gap(s) (user override)
**Committed:** ✓
**Tests/linters:** ✓ (commands run)
**Pushed:** ✓ openspec/<name>
**PR:** <url> — **open, not merged**

**What to do (Gate D):**
1. Review the PR at the link above
2. Changes needed → `/myflow-do-fix <name>` (commits and pushes to this PR)
3. **Merge the PR yourself** — myflow never merges for you
4. Then → `/myflow-finish <name>` (verify merged, sync specs, archive)
```

## Guardrails

- **Always** check the stage gate (step 0) before anything else.
- **Always** check Gate C completion before verification.
- **Always** run the test-coverage check before verification — never silently skip it.
- **Never** write missing tests yourself here — flag them and route to `/myflow-do-fix <name>`.
- **Always commit** uncommitted apply work before verification (unless already committed via `commit-during-apply`).
- Never mark verification complete with failing tests unless the user explicitly overrides after seeing failures.
- **Never merge.** This stage pushes and opens a PR; merging is Gate D, done by the human.
- **Never push directly to `develop`/`main`** as a substitute for opening a PR.
- **Always write `stage: awaiting-pr-review`** before handing off.
- Do not force-push or amend unless user rules allow.
- Do not archive here — that is `/myflow-finish`'s job.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Run code review (coverage, tests, commit, push, PR) | `/myflow-code-review <name>` |
| Fix a coverage gap or test/lint failure | `/myflow-do-fix <name>` |
| After code review — PR review (Gate D, human) | review the PR on the forge and merge it |
| Changes requested during Gate D | `/myflow-do-fix <name>` (commits and pushes to the existing PR) |
| After the PR is merged | `/myflow-finish <name>` |
