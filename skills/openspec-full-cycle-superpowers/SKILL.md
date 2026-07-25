---
name: openspec-full-cycle-superpowers
description: Run full OpenSpec start → do (#2–#6) → manual review → manual test → code review (commit+#7) → finish (archive) with approval gates. Use for /myflow-full.
allowed-tools: Bash(openspec:*)
license: MIT
compatibility: Requires openspec CLI and Superpowers plugin skills.
metadata:
  author: gymie
  version: "4.0"
---

Run **start → do (#2–#6) → manual review (Gate B) → manual test (Gate C) → code review (commit + #7) → finish (archive)** with user gates. Gate B and Gate C each support an inline `/myflow-do-fix` loop before continuing — repeat as many times as needed.

**Announce at start:** "Using openspec-full-cycle-superpowers."

Also follow **rules/myflow-manual-review.mdc** (Cursor: `.cursor/rules/myflow-manual-review.mdc`).

## Superpowers Basic Workflow — full map

| Step | Skill | Stage | Bridge skill |
|------|-------|-------|--------------|
| **1** | brainstorming | Start | openspec-propose-superpowers |
| **2** | using-git-worktrees | Do | openspec-apply-superpowers |
| **3** | writing-plans | Start (+ validate at do) | openspec-propose-superpowers / apply |
| **4** | subagent-driven-development | Do (+ Do-fix rounds) | openspec-apply-superpowers / openspec-apply-fix-superpowers |
| **5** | test-driven-development | Do (+ Do-fix rounds, every task) | openspec-apply-superpowers / openspec-apply-fix-superpowers |
| **6** | requesting-code-review + strict panel (Bugbot, Security, Adversarial, Senior, Economic Senior) | Do (+ Do-fix rounds, full re-run each time) | openspec-apply-superpowers / openspec-apply-fix-superpowers |
| **7** | finishing-a-development-branch | **Code review** | openspec-code-review-superpowers |

**Mandatory:** Steps #1–#6 run in start/do/do-fix. **#7 always runs in code review** (moved out of the old combined archive stage). None may be skipped.

## Required sub-skills (in order)

1. **openspec-propose-superpowers** — #1 + OpenSpec artifacts + #3
2. **openspec-apply-superpowers** — #2, #3 validate, #4–#6 (stage with `git add`, no commits, no #7)
3. **openspec-apply-fix-superpowers** — inline fix rounds at Gate B and Gate C (optional, repeatable)
4. **openspec-manual-test-superpowers** — Gate C guide (`docs/manual-test/<name>.md`, link only)
5. **openspec-code-review-superpowers** — test coverage check + verification + commit + #7
6. **openspec-archive-superpowers** — verify merged + sync delta specs + archive

Each stage delegates to its bridge skill; do not reimplement steps inline.

## Input

- Change name (kebab-case) and/or description of what to build
- **If both are omitted:** run `openspec list --json`, filter to non-archived changes. Exactly one match → use it automatically (resume at whatever stage it's at), announce which; multiple matches → **AskUserQuestion** listing each (name, status, last modified) — do not guess; zero matches → ask what to build.
- Optional flags:
  - **propose-only** — stop after Gate A (#1 + #3 + artifacts done)
  - **skip-propose** — change apply-ready; start at do (#2–#6)
  - **skip-review** — skip Gate B manual code review pause; still run Gate C unless also skipped
  - **skip-manual-test** — skip Gate C; proceed to code review without generating/waiting on the guide
  - **no-archive** — stop after code review; no finish/archive
  - **commit-during-apply** — legacy SDD per-task commits during apply; only when user explicitly requests

If ambiguous, ask once before starting.

## Workflow

### Phase A — Start: #1 + OpenSpec + #3 (skip if `skip-propose` or apply-ready)

1. Check existing change:
   ```bash
   openspec list --json
   openspec status --change "<name>" --json
   ```
2. If change missing, artifacts incomplete, or `tasks.md` not writing-plans-ready: invoke **openspec-propose-superpowers**.
3. If change exists and apply-ready: skip to Gate A.

**Gate A — Proposal review**

Use **AskUserQuestion**:

- "Proposal ready (#1 brainstorming ✓, #3 plan ✓). Continue to do (#2–#6)?"
- Options: **Continue to do** (recommended), **Revise proposal**, **Stop here**

If **Stop here** or **propose-only**: exit with summary and `/myflow-do <name>` hint.

### Phase B — Do: #2–#6 only (skip if only proposing)

Invoke **openspec-apply-superpowers** for `<name>`.

If blocked mid-apply: stop; report progress. Resume with `/myflow-full <name> skip-propose` or `/myflow-do <name>`.

**Gate B — Manual code review (+ optional fix loop)**

After do completes (#2–#6), **stop** unless user passed `skip-review`.

Provide manual review handoff (worktree path to open in IDE, branch, confirmation changes are **staged + uncommitted**, `git status`, `git diff --cached --stat`, full staged diff command). Do already ran `git add -A` in affected repos.

Use **AskUserQuestion** unless `skip-review` or `no-archive` with intent to stop early:

- "Do complete (#2–#6 ✓, staged + uncommitted). Review staged changes in the worktree IDE. Anything to fix, or continue to manual test?"
- Options: **Continue to manual test** (recommended), **Fix something (`/myflow-do-fix`)**, **Stop here**

If **Fix something**: invoke **openspec-apply-fix-superpowers** for `<name>` (this is a Gate B fix). After it completes, return to this same Gate B question — repeat as many rounds as needed until the user picks continue or stop.

If **Stop here**: exit with review commands and `/myflow-manual-test <name>` / `/myflow-do-fix <name>` hints.

**Do not commit or run #7 at Gate B.**

### Phase C — Manual test (Gate C, + optional fix loop)

Unless user passed `skip-manual-test`:

1. Invoke **openspec-manual-test-superpowers** for `<name>` (writes `docs/manual-test/<name>.md`, replies with link only).
2. Use **AskUserQuestion** unless `no-archive`:

- "Manual test guide ready. Run the apps, complete the checklist. Anything to fix, or continue to code review?"
- Options: **Continue to code review** (recommended), **Fix something (`/myflow-do-fix`)**, **Stop here**

If **Fix something**: invoke **openspec-apply-fix-superpowers** for `<name>` (this is a Gate C fix). Since the fix may touch tested behavior, refresh the guide via **openspec-manual-test-superpowers** afterward before returning to this question. Repeat as many rounds as needed.

If **Stop here** or `no-archive`: exit with guide path and `/myflow-code-review <name>` hint.

**Do not commit or run #7 at Gate C.**

### Phase D — Code review: coverage + verify + commit + #7

Invoke **openspec-code-review-superpowers** for `<name>`.

If it finds coverage gaps, it will prompt for `/myflow-do-fix <name>` itself — that's an internal step of this phase, not a separate full-cycle gate.

If `no-archive`: stop after this phase; report summary and `/myflow-finish <name>` hint for later.

### Phase E — Finish: verify merged + sync specs + archive

Invoke **openspec-archive-superpowers** for `<name>`.

### Final summary

```
## Full Cycle Complete

**Change:** <name>
**Basic Workflow:** #1 ✓ #2 ✓ #3 ✓ #4 ✓ #5 ✓ #6 ✓ #7 ✓
**Started:** ✓ (or skipped)
**Applied (do):** ✓ N/N tasks (staged + uncommitted until code review)
**Manual review (Gate B):** ✓ (or skipped via skip-review) — <M> fix round(s)
**Manual test (Gate C):** ✓ (or skipped via skip-manual-test) — <M> fix round(s)
**Code review:** ✓ coverage checked, tests/linters green, committed, #7 done
**Finished:** ✓ <archive path>
```

## Resume semantics

| Situation | Command |
|-----------|---------|
| Fresh feature | `/myflow-full <name>` + description |
| Proposal done | `/myflow-do <name>` or `/myflow-full <name> skip-propose` |
| Do done, awaiting review | Review diff; `/myflow-do-fix <name>` if needed; then `/myflow-manual-test <name>` |
| Ready to test | `/myflow-manual-test <name>` |
| Testing done | `/myflow-code-review <name>` |
| Code review done | `/myflow-finish <name>` |
| Partial do | `/myflow-do <name>` (SDD ledger resumes) |
| Gate B/C finding | `/myflow-do-fix <name>` |

## Guardrails

- **Always stop** at Gate A unless user passed `skip-propose` or `propose-only`.
- **Always stop** at Gate B for manual review unless user passed `skip-review`; loop on fix requests until the user chooses to continue or stop.
- **Always stop** at Gate C for manual test unless user passed `skip-manual-test`; loop on fix requests (refreshing the guide after each fix) until the user chooses to continue or stop.
- **Never commit, push, merge, or run #7** during do or do-fix (unless `commit-during-apply`).
- **Always run #7** during code review in standard myflow.
- Do not substitute OpenSpec-only loops for #4–#6.
- Prefer stage bridge skills over one unstructured session.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Full cycle with gates | `/myflow-full <name>` + what to build |
| Proposal exists | `/myflow-full <name> skip-propose` |
| Propose only (#1 + #3) | `/myflow-full <name> propose-only` |
| Skip manual code review | `/myflow-full <name> skip-review` |
| Skip manual test | `/myflow-full <name> skip-manual-test` |

Individual stages:

- `/myflow-start <name>` — #1 + OpenSpec + #3
- `/myflow-do <name>` — #2–#6 (stage for IDE; no commits)
- `/myflow-do-fix <name>` — fix a Gate B/C finding (documents it first)
- `/myflow-manual-test <name>` / `/myflow-manual-test-skip <name>` — Gate C guide MD
- `/myflow-code-review <name>` — coverage check, tests/linters, commit, #7
- `/myflow-finish <name>` — verify merged, sync specs, archive
