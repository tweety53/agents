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

**Detect capability first — never assume `gh` exists or that the forge is GitHub:**

```bash
cd <worktree>
git remote get-url origin 2>/dev/null || echo "NO_REMOTE"
command -v gh >/dev/null 2>&1 && echo "gh: yes" || echo "gh: no"
```

Then take exactly one of three branches:

**6.1 — `gh` installed AND `origin` is a GitHub host** (`github.com` or a GitHub Enterprise host):

```bash
git push -u origin openspec/<name>
gh pr list --head openspec/<name> --state open --json number,url    # reuse if one already exists
gh pr create --base develop --head openspec/<name> --title "<change title>" --body "<summary + link to proposal>"
```

If a PR for this branch already exists (a `/myflow-do-fix` round pushed to it), skip creation and reuse it. → PR is open: `stage: awaiting-pr-review`, `gates.prOpened: true`.

**6.2 — a remote exists but there is no usable PR CLI for that host** (Bitbucket, GitLab without `glab`, or `gh` simply not installed — **this project's case**: `origin` is Bitbucket and `gh` is not installed):

```bash
git push -u origin openspec/<name>
```

Then derive the forge's create-PR URL from the `origin` URL, handling both SSH (`git@host:<workspace>/<repo>.git`) and HTTPS (`https://host/<workspace>/<repo>.git`) forms, and print it:

- Bitbucket: `https://bitbucket.org/<workspace>/<repo>/pull-requests/new?source=openspec/<name>&t=1`
- GitLab: `https://<host>/<workspace>/<repo>/-/merge_requests/new?merge_request[source_branch]=openspec/<name>`

Then **AskUserQuestion**:

> **Have you opened the PR?**
> - **No — not yet** *(default, recommended)*
> - **Yes — the PR is open**

- **Yes** → the human's confirmation is the evidence (the same trust model already used at Gates B and C): `stage: awaiting-pr-review`, `gates.prOpened: true`. Ask for the PR URL and record it in the summary if they have it; **do not block** if they don't.
- **No** → stay at `awaiting-test`, `gates.prOpened: false`. Say plainly what to do next: open the printed URL, create the PR, then re-run `/myflow-code-review <name>` (it will skip straight to this step) or confirm on the next run.

**Never** substitute a merge or a direct push to `develop`/`main` for opening a PR in this branch.

**6.3 — no remote at all:** pushing is impossible. Stop, stay at `awaiting-test`, `gates.prOpened: false`, and tell the user to add a remote. Do not fall back to merging.

### 6b. Write state

Resolve the state file path per **State file** in `rules/myflow-manual-review.mdc` (`--git-common-dir` → `<project-key>` → `/Users/tweety53/Agents/myflow/state/<project-key>/<name>.json`). It lives outside the repo: **never `git add` it, never commit it, never push it.**

**If branch 6.1 opened/reused a PR, or branch 6.2 got a "Yes":**

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

**If branch 6.2 got a "No", or branch 6.3 applied (no PR exists yet):** do not advance the stage — `awaiting-pr-review` would be a false claim. Leave the change at `awaiting-test` and say plainly that it does not advance to Gate D until a PR actually exists:

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

**Carry gates forward.** Read the existing state file first and preserve every gate this command does not own. Gate values are **monotonic** — never lower `gates.reviewed`, never overwrite `gates.tested: "skipped"`, never write a stage earlier than the one found.

There is no state commit and no follow-up commit: the state file is user-scoped and outside the repo, so step 5's code commit is the only commit this stage makes.

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
**PR:** <url> — **open, not merged** | opened by you, URL not recorded | **not opened yet** — create it at <compare-url>, still at `awaiting-test`

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
- **Never assume `gh` is installed or that the forge is GitHub** — detect capability (step 6) and use the confirmation branch otherwise.
- **Never `git add`, commit, or push the state file** — it is user-scoped and outside the repo.
- **Write `stage: awaiting-pr-review` only when a PR actually exists** (created by `gh`, or confirmed by the human); otherwise stay at `awaiting-test` with `gates.prOpened: false`.
- **Never lower a gate value** — gates are monotonic; `gates.tested: "skipped"`/`true` are sticky.
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
