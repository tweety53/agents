---
name: openspec-full-cycle-superpowers
description: Run full OpenSpec start → do (#2–#6) → manual review → manual test → review (commit+push+PR) → Gate D, stop (or, with automerge, ends at review-done — no PR to stop at). Use for /myflow-full. /myflow-finish is a separate, human-initiated step after the PR is merged.
allowed-tools: Bash(openspec:*)
license: MIT
compatibility: Requires openspec CLI and Superpowers plugin skills.
metadata:
  author: gymie
  version: "5.0"
---

Run **start → do (#2–#6) → manual review (Gate B) → manual test (Gate C, skip prompt) → review (commit + push + PR) → Gate D — STOP** with user gates (or, with `automerge`, review merges immediately and the cycle ends at `review-done` instead of Gate D). Gate B and Gate C each support an inline `/myflow-do-fix` loop before continuing — repeat as many times as needed.

**The cycle always stops at Gate D with an open PR — unless `automerge` was passed.** Merging is normally a human action performed on the forge, outside myflow — a composite command cannot automate past it on its own. `automerge` is the one explicit exception: passed through to `/myflow-review`, it merges immediately, so there is no PR to stop at and the cycle ends at `review-done` instead of Gate D. `/myflow-finish` is always a separate, human-initiated command run **after** the PR is merged (automatically, via `automerge`, or by the human at Gate D); this command never invokes it.

**Announce at start:** "Using openspec-full-cycle-superpowers."

Also follow **rules/myflow-manual-review.mdc** (Cursor: `.cursor/rules/myflow-manual-review.mdc`). In particular see **## Pipeline stages**, **## Stage transitions**, **## Full cycle gates**, and **## Opt-out (explicit only)** — the source of truth for stage names, gate behavior, and flags.

## Superpowers Basic Workflow — full map

| Step | Skill | Stage | Bridge skill |
|------|-------|-------|--------------|
| **1** | brainstorming | Start | openspec-propose-superpowers |
| **2** | using-git-worktrees | Do | openspec-apply-superpowers |
| **3** | writing-plans | Start (+ validate at do) | openspec-propose-superpowers / apply |
| **4** | subagent-driven-development | Do (+ Do-fix rounds) | openspec-apply-superpowers / openspec-apply-fix-superpowers |
| **5** | test-driven-development | Do (+ Do-fix rounds, every task) | openspec-apply-superpowers / openspec-apply-fix-superpowers |
| **6** | requesting-code-review + strict panel (Bugbot, Security, Adversarial, Senior, Economic Senior) | Do (+ Do-fix rounds, full re-run each time) | openspec-apply-superpowers / openspec-apply-fix-superpowers |
| **7** | finishing-a-development-branch — **commit + push + open PR only, no merge** | Review | openspec-review-superpowers |

**Mandatory:** Steps #1–#6 run in start/do/do-fix; none may be skipped. **Review runs the commit + push + open-PR portion of #7** and then stops — it never merges. The human merges the PR at Gate D, outside myflow.

## Required sub-skills (in order)

1. **openspec-propose-superpowers** — #1 + OpenSpec artifacts + #3
2. **openspec-apply-superpowers** — #2, #3 validate, #4–#6 (stage with `git add`, no commits, no #7)
3. **openspec-apply-fix-superpowers** — inline fix rounds at Gate B and Gate C (optional, repeatable)
4. **openspec-manual-test-superpowers** — Gate C guide (`docs/manual-test/<name>.md`, link only); always asks the skip prompt itself
5. **openspec-review-superpowers** — test coverage check + verification + commit + push + open PR (never merges)

Each stage delegates to its bridge skill; do not reimplement steps inline. `openspec-archive-superpowers` (the `/myflow-finish` skill) is deliberately **not** in this list — this cycle never reaches it; the human invokes `/myflow-finish` separately after merging.

## Input

- Change name (kebab-case) and/or description of what to build
- **If both are omitted:** run `openspec list --json`, filter to non-archived changes. Exactly one match → use it automatically (resume at whatever stage it's at), announce which; multiple matches → **AskUserQuestion** listing each (name, status, last modified) — do not guess; zero matches → ask what to build.
- Optional flags:

| Flag | Effect |
|------|--------|
| `skip-propose` | Start from `/myflow-do` using existing artifacts |
| `propose-only` | Stop after planning artifacts (Gate A) |
| `skip-review` | Skip Gate B only (advancing straight to `do-done`); Gate C still runs. Records `gates.reviewed: false` (Gate B explicitly skipped) — never `true` |
| `skip-manual-test` | Pre-answers the Gate C skip prompt with **Yes**, then carries the change on to `manual-test-done` with `gates.tested: "skipped"` (the Gate C question is not asked); review still runs and still checks coverage |
| `automerge` | Pass-through to `/myflow-review automerge` (see **Auto-merge (opt-in)** in the rule file). Merges immediately — no PR to stop at, so the cycle ends at `review-done` instead of Gate D |
| `commit-during-apply` | Legacy per-task SDD commits |

`no-archive` is **removed** — the cycle no longer reaches archiving, so the flag had no effect.

Without `skip-manual-test`, the manual-test stage asks the skip prompt normally (default **No**), exactly as the standalone `/myflow-manual-test` command does.

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

**Gate B — Manual review (+ optional fix loop)**

After do completes (#2–#6), **stop** unless user passed `skip-review`.

Provide manual review handoff (worktree path to open in IDE, branch, confirmation changes are **staged + uncommitted**, `git status`, `git diff --cached --stat`, full staged diff command). Do already ran `git add -A` in affected repos.

Use **AskUserQuestion** unless `skip-review`:

- "Do complete (#2–#6 ✓, staged + uncommitted). Review staged changes in the worktree IDE. Anything to fix, or continue to manual test?"
- Options: **Continue to manual test** (recommended), **Fix something (`/myflow-do-fix`)**, **Stop here**

If **Continue to manual test**: do ended at `awaiting-do-review`, but Phase C's skill accepts only `do-done` (advance) or `awaiting-manual-test` (refresh). **Write `stage: do-done` directly** — with `gates.reviewed: true` (the human did review the diff; under `skip-review` write `false` instead, per the flag's own rule) and every other field carried forward — then proceed to Phase C. This does **not** invoke `/myflow-do-done`; see **Crossing a gate on "Continue"** below.

If **Fix something**: invoke **openspec-apply-fix-superpowers** for `<name>` (this is a Gate B fix). After it completes, return to this same Gate B question — repeat as many rounds as needed until the user picks continue or stop.

If **Stop here**: exit with review commands and `/myflow-manual-test <name>` / `/myflow-do-fix <name>` hints. **Do not** write `do-done` — the human did not confirm the review.

**Do not commit or run #7 at Gate B.**

### Phase C — Manual test (Gate C, + optional fix loop)

Invoke **openspec-manual-test-superpowers** for `<name>` (writes `docs/manual-test/<name>.md`, replies with link only). That skill **always asks** whether to skip Gate C (default and recommended **No**):

- If the user passed `skip-manual-test` on `/myflow-full`, **pre-answer that prompt with Yes** and **announce that the flag pre-answered it** — do not silently skip. This is the only flag on `/myflow-full` that pre-answers a prompt instead of skipping a phase outright.
- Otherwise let the prompt run normally and honor whatever the user answers.

Either answer advances to `awaiting-manual-test`; the distinction lives in `gates.tested` (`"skipped"` vs `false`).

**With `skip-manual-test`, carry it through to `manual-test-done`.** Pre-answering the prompt leaves the change at `awaiting-manual-test`, which Phase D does not accept — so after the manual-test skill returns, **write `stage: manual-test-done` directly** with `gates.tested: "skipped"` (every other field carried forward) and skip the Gate C question entirely, announcing that the flag both pre-answered the skip prompt and carried the change past Gate C. This is the mechanism rule line "How `skip-review` and `skip-manual-test` reach a `*-done` stage" calls for: typing the flag at invocation is the human's explicit, in-the-moment decision, so writing the stage forward executes that decision rather than self-certifying testing nobody did. It does **not** invoke `/myflow-manual-test-done`. Never write `gates.tested: true` on this path — `"skipped"` is the honest value and only `/myflow-review` ever writes `true`.

**If `skip-review` was passed**, tell the manual-test skill to record `gates.reviewed: false` (Gate B was explicitly skipped) and note it in the final summary — the state must never claim a manual review that never happened.

Use **AskUserQuestion**:

- "Manual test guide ready. Run the apps, complete the checklist. Anything to fix, or continue to review?"
- Options: **Continue to review** (recommended), **Fix something (`/myflow-do-fix`)**, **Stop here**

If **Continue to review**: Phase C ended at `awaiting-manual-test`, but Phase D requires `manual-test-done`. **Write `stage: manual-test-done` directly** — carrying `gates.tested` forward exactly as the manual-test skill left it (`false` when the human ran the checklist, `"skipped"` under `skip-manual-test`), never promoting it to `true` — then proceed to Phase D. This does **not** invoke `/myflow-manual-test-done`; see **Crossing a gate on "Continue"** below.

If **Fix something**: invoke **openspec-apply-fix-superpowers** for `<name>` (this is a Gate C fix). Since the fix may touch tested behavior, refresh the guide by invoking **openspec-manual-test-superpowers** again afterward — at stage `awaiting-manual-test` it enters **refresh mode** automatically (no stage-mismatch prompt, no override, skip question not re-asked, checked boxes preserved) — then return to this question. Repeat as many rounds as needed.

If **Stop here**: exit with guide path and `/myflow-review <name>` hint. **Do not** write `manual-test-done` — the human did not confirm.

**Do not commit or run #7 at Gate C.**

### Returning from an inline fix

`/myflow-do-fix` always ends at **`awaiting-fix-review`** — it never returns the change to the
stage the fix was raised at. That return is `/myflow-do-fix-done`'s job, and this cycle may not
invoke it (it is a human confirmation, like every other `*-done`).

So when the human picks "Fix something" at Gate B or Gate C, the cycle:

1. invokes `/myflow-do-fix`, which records `originStage` and lands at `awaiting-fix-review`;
2. **stops** and names the two commands the human runs — `/myflow-do-fix-manual-review <name>`
   (optional, marks the fix review started) then `/myflow-do-fix-done <name>`, which reads
   `originStage`, returns the change to it, and clears the field;
3. resumes at that origin once the human re-invokes `/myflow-full <name>`.

Do **not** re-invoke the manual-test or review skill while the change sits at
`awaiting-fix-review` — their gates will refuse it, and answering the override would write a
`*-done` stage over a fix that was never confirmed, leaving a stale `originStage` behind.

### Crossing a gate on "Continue"

These two rules look contradictory and are not; they are stated together here so neither is read alone:

1. **`/myflow-full` never invokes a `*-done` or `*-manual-review` command.** That prohibition stands, unchanged, in the Guardrails below and in the rule file.
2. **Choosing "Continue" at a Gate B or Gate C prompt is the human's explicit, in-the-moment confirmation** — the same act `/myflow-do-done` / `/myflow-manual-test-done` performs, made inside the cycle. So on "Continue", the cycle **writes the `*-done` stage directly**, with honest gate values, and proceeds.

The two coexist because the prohibition exists to stop the agent from *self-certifying a review nobody performed*. Here a human was asked, stopped for, and answered — the cycle still stops and waits for that answer, and it still never runs the `*-done` command. This is the identical reasoning rule line "How `skip-review` and `skip-manual-test` reach a `*-done` stage" already applies to those flags: the human's explicit instruction is what authorizes the stage write, not an inference by the agent.

Constraints on this direct write, all mandatory:

- Only on an explicit **Continue** answer. **Stop here** never writes a `*-done` stage; neither does an unanswered or aborted prompt.
- Gate values must be **honest**: `gates.reviewed: true` only when Gate B actually ran (`false` under `skip-review`); `gates.tested` carried forward exactly as recorded and **never** promoted to `true` — only `/myflow-review` writes that.
- Write the whole state object, carrying every other field forward (including `artifactUrl` and `worktree`).
- Never invoke the `*-done` command itself, and never write a `*-done` stage at any other point in the cycle.

### Phase D — Review: coverage + verify + commit + push + open PR (or merge, with `automerge`) — STOP

Invoke **openspec-review-superpowers** for `<name>` — pass `automerge` through if the user passed it to `/myflow-full`.

If it finds coverage gaps, it will prompt for `/myflow-do-fix <name>` itself — that's an internal step of this phase, not a separate full-cycle gate.

Without `automerge`, this phase **commits, pushes, and opens a PR — it never merges**. The cycle ends here, at Gate D, every time: report the PR summary (see **Final summary** below) and stop. `/myflow-full` does not invoke `/myflow-finish`; the human reviews and merges the PR on the forge, then runs `/myflow-finish <name>` themselves whenever they're ready.

**With `automerge`,** the review phase commits, pushes, and merges immediately (per the rule file's **Auto-merge (opt-in)**) — there is no PR to stop at, so the cycle skips Gate D entirely and ends at stage `review-done`. Announce plainly that auto-merge was used, then report the automerge summary variant below and stop; `/myflow-finish <name>` is still a separate, human-initiated step.

### Final summary

Without `automerge`:

```
## Full Cycle Complete — PR Review Required (Gate D)

**Change:** <name>
**Basic Workflow:** #1 ✓ #2 ✓ #3 ✓ #4 ✓ #5 ✓ #6 ✓ #7 (commit+push+PR, no merge) ✓
**Started:** ✓ (or skipped)
**Applied (do):** ✓ N/N tasks (staged + uncommitted until review)
**Manual review (Gate B):** ✓ (or skipped via skip-review) — <M> fix round(s)
**Manual test (Gate C):** ✓ ran | ✓ skipped (skip-manual-test pre-answered the prompt) — <M> fix round(s)
**Review:** ✓ coverage checked, tests/linters green, committed, pushed
**PR:** <url> — **open, not merged**

**What to do (Gate D):**
1. Review the PR at the link above
2. Changes needed → `/myflow-do-fix <name>` (commits and pushes to this PR)
3. **Merge the PR yourself** — myflow never merges for you
4. Then, whenever you're ready → `/myflow-finish <name>` (verify merged, sync specs, archive) — a separate step you run manually
```

With `automerge`:

```
## Full Cycle Complete — Merged (automerge, stage: review-done)

**Change:** <name>
**Basic Workflow:** #1 ✓ #2 ✓ #3 ✓ #4 ✓ #5 ✓ #6 ✓ #7 (commit+push+merge via automerge) ✓
**Started:** ✓ (or skipped)
**Applied (do):** ✓ N/N tasks (staged + uncommitted until review)
**Manual review (Gate B):** ✓ (or skipped via skip-review) — <M> fix round(s)
**Manual test (Gate C):** ✓ ran | ✓ skipped (skip-manual-test pre-answered the prompt) — <M> fix round(s)
**Review:** ✓ coverage checked, tests/linters green, committed, pushed, **merged (automerge)**
**Gate D:** skipped — no PR to review, merge already landed

**What to do next:**
1. Whenever you're ready → `/myflow-finish <name>` (verify merged, sync specs, archive) — a separate step you run manually
```

## Resume semantics

| Situation | Command |
|-----------|---------|
| Fresh feature | `/myflow-full <name>` + description |
| Proposal done | `/myflow-do <name>` or `/myflow-full <name> skip-propose` |
| Do done, awaiting review | Review diff; `/myflow-do-fix <name>` if needed; then `/myflow-manual-test <name>` |
| Ready to test | `/myflow-manual-test <name>` |
| Testing done | `/myflow-review <name>` |
| Review done, PR open | Review + merge the PR yourself (Gate D), then `/myflow-finish <name>` |
| Review done, automerge used | No PR — go straight to `/myflow-finish <name>` |
| Partial do | `/myflow-do <name>` (SDD ledger resumes) |
| Gate B/C/D finding | `/myflow-do-fix <name>` |

## Guardrails

- **Always stop** at Gate A unless user passed `skip-propose` or `propose-only`.
- **Always stop** at Gate B for manual review unless user passed `skip-review`; loop on fix requests until the user chooses to continue or stop.
- **Always** let Gate C's skip prompt run; `skip-manual-test` pre-answers it with Yes (announce this) instead of bypassing the phase, and then carries the change to `manual-test-done` with `gates.tested: "skipped"`. Loop on fix requests (refreshing the guide after each fix) until the user chooses to continue or stop.
- **Never commit, push, merge, or run #7** during do or do-fix (unless `commit-during-apply`).
- **Review commits, pushes, and opens a PR — it never merges, unless the user explicitly passed `automerge`.** Without `automerge` the cycle always stops at Gate D; with it, the cycle ends at `review-done` instead — announce that auto-merge was used.
- **Never auto-invoke a `*-done` or `*-manual-review` command anywhere in this cycle.** `/myflow-start-done`, `/myflow-do-manual-review`, `/myflow-do-done`, `/myflow-manual-test-done`, and `/myflow-review-done` are human confirmations of a review the human actually performed — this cycle stops before each of them and names the exact command for the human to run; it never runs one itself.
- **The one thing the cycle may write is the `*-done` stage on an explicit "Continue" answer at Gate B or Gate C** — see **Crossing a gate on "Continue"**. That is not an exception to the rule above (no `*-done` command is ever invoked); it is the human's own in-the-moment confirmation being recorded, exactly as `skip-review`/`skip-manual-test` are. It still requires the cycle to stop and wait for the answer, and the gate values written must be honest — never `gates.tested: true`.
- **Never write a `*-done` stage on "Stop here", on an aborted prompt, or anywhere outside those two Continue answers.**
- **Never invoke `/myflow-finish` from this cycle.** It is always a separate, human-initiated command run after the PR is merged (or, with `automerge`, after the automatic merge).
- Do not substitute OpenSpec-only loops for #4–#6.
- Prefer stage bridge skills over one unstructured session.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Full cycle with gates, ending at Gate D | `/myflow-full <name>` + what to build |
| Proposal exists | `/myflow-full <name> skip-propose` |
| Propose only (#1 + #3) | `/myflow-full <name> propose-only` |
| Skip manual review | `/myflow-full <name> skip-review` |
| Pre-answer the Gate C skip prompt with Yes | `/myflow-full <name> skip-manual-test` |
| Auto-merge instead of stopping at Gate D | `/myflow-full <name> automerge` |

Individual stages:

- `/myflow-start <name>` — #1 + OpenSpec + #3
- `/myflow-do <name>` — #2–#6 (stage for IDE; no commits)
- `/myflow-do-fix <name>` — fix a Gate B/C/D finding (documents it first)
- `/myflow-manual-test <name>` — Gate C guide MD; always asks the skip prompt
- `/myflow-review <name>` — coverage check, tests/linters, commit, push, open PR (never merges)
- `/myflow-finish <name>` — separate, human-initiated: verify merged, sync specs, archive
