---
name: openspec-fast-path-superpowers
description: Shortened single-session myflow variant for small, well-understood features. Writes minimal OpenSpec artifacts, implements inline with TDD, reviews with a three-agent panel, and ends at a PR — collapsing five human gates to one. Escalates to the standard pipeline on any size trigger. Use for /myflow-fast-path.
allowed-tools: Bash(openspec:*)
license: MIT
compatibility: Requires openspec CLI and Superpowers plugin skills.
metadata:
  author: gymie
  version: "1.0"
---

Implement a **small, well-understood feature** in one session. The premise is that the design is
already settled — no brainstorming, no design approval gate, no manual test guide. What survives:
**TDD**, an **OpenSpec record**, **agent code review**, and the **worktree + PR** shape.

**Announce at start:** "Using openspec-fast-path-superpowers for change `<name>`."

Also follow **rules/myflow-manual-review.mdc** (Cursor: `.cursor/rules/myflow-manual-review.mdc`).

## When to use this vs `/myflow-start` + `/myflow-do`

- You already know what to build and roughly how, it touches one or two modules, and you want a PR
  today → `/myflow-fast-path`.
- The design needs exploring, the change spans modules, or it touches contracts/migrations/specs →
  the standard pipeline. This skill will tell you when it thinks you crossed that line (see
  **step 4a**).

## What this skill cuts

| Cut | Kept instead |
|-----|--------------|
| Brainstorming (#1) + Gate A proposal artifact | A terse `proposal.md` you confirm once, inline |
| `design.md`, delta specs by default | Delta specs only when intended behaviour genuinely changes |
| SDD subagent per task (#4) | Inline RED-GREEN-REFACTOR by the main agent |
| The conditional review panel slots | Primary + Bugbot + Principles — the three required slots (every slot with `full-panel`) |
| Gate B staged-diff stop | Optional, via `checkpoint` |
| Gate C manual test guide + run | Nothing — `gates.tested: "skipped"`, honestly recorded |

## Flags

- `checkpoint` — add a Gate B stop on the staged diff before anything is pushed.
- `full-panel` — run every review slot (required plus both extra principle lenses) instead of the lean three.
- **`automerge` is not accepted.** This command always ends at a PR a human reviews. If the user
  passes it, say so and continue without it.

## Stage machine

This skill adds **no new stages**. It writes only stages that already exist in **Pipeline stages**:

| Point | Stage written | Gate values |
|-------|---------------|-------------|
| Entry | no state file (new change), `proposal-done`, or `awaiting-do-review` **with `fastPath: true`** (checkpoint resume — step 6a) | — |
| After artifacts written and confirmed (step 2) | `proposal-done` | — |
| `checkpoint` stop (step 6a) | `awaiting-do-review` | — |
| On "Continue" past the checkpoint | `do-done` | `reviewed: false` |
| End of run (step 7), **only once a PR exists** | `awaiting-pr-review` | `prOpened: true`, `tested: "skipped"` |

Also write **`fastPath: true`** into the state file, so `/myflow-status` and `/myflow-finish` can
report how the change got here. Resolve the state file path per **State file** in `skills/myflow-contracts/state-file.md`. It lives outside the repo — **never stage, commit, or push it.**

**`jiraIssue`** is carried forward verbatim on every write this skill makes. On a change this
command creates from scratch, resolve it the same way `/myflow-start` does (issue-key scan → fetch
to confirm → ask once, accepting "none") and record it on the first write; on a change entering at
`proposal-done`, `/myflow-start` already recorded it — never overwrite or reset it to `null`.

**Gate honesty.** Never write `gates.reviewed: true` or `gates.tested: true`. Invoking this command
is the human's explicit, in-the-moment decision to skip those gates — which authorizes writing the
stage forward, exactly as `skip-review`/`skip-manual-test` do under **Opt-out (explicit only)** —
but it never authorizes claiming a gate ran. Carry every other gate value forward unchanged (gates
are monotonic).

**Timestamps.** Every state write sets `updatedAt` to the **actual current time** in ISO-8601 UTC —
read it, never invent it:

```bash
date -u +%Y-%m-%dT%H:%M:%SZ
```

A placeholder like `T00:00:00Z` is a defect: `/myflow-status` reports "last update" from this field,
and a fabricated timestamp makes a stalled change look freshly touched. Set `updatedBy` to
`"/myflow-fast-path"` on every write this skill makes.

Downstream is unmodified: the human merges the PR (Gate D) → `/myflow-review-done <name>` →
`/myflow-finish <name>`.

## Workflow

### 0. Check stage and eligibility

Resolve the change name from `$ARGUMENTS` or the conversation. If omitted, follow **Change name
resolution**: `openspec list --json`, exactly one active candidate → use it and announce; multiple →
**AskUserQuestion**; zero → treat as a new change and ask for a name.

Read the state per **State file** in `skills/myflow-contracts/state-file.md` and validate per **State self-heal** in `skills/myflow-contracts/state-self-heal.md`.

| Current stage | Action |
|---------------|--------|
| no state file (new change) | Proceed — this is the normal entry. |
| `proposal-done` | Proceed — artifacts may already exist; step 1 adds only what is missing and never overwrites an existing file. |
| `awaiting-do-review` | Proceed **only** in checkpoint resume — the state file has `fastPath: true` (see step 6a). Then **skip steps 1–6 and go straight to the resume path at the end of step 6a**, which decides whether step 7 may be entered directly. Without `fastPath: true`, refuse. |
| any other stage | **Refuse.** |

On refusal, emit the standard mismatch handoff from **Stage transitions** and recommend by stage:
`awaiting-proposal-review` → `/myflow-start-done`; `do-done`/`awaiting-manual-test`/
`manual-test-done` → `/myflow-manual-test` or `/myflow-review`; `awaiting-fix-review`/
`fix-review-started` → `/myflow-do-fix-done`; `awaiting-pr-review`/`review-done` → `/myflow-do-fix`
or `/myflow-finish`; `finished` → stop, the change is archived. Then **AskUserQuestion** for an
explicit override (default: **No — run the suggested command instead**).

A change that is already past `proposal-done` has had real work done under the standard pipeline's
guarantees; silently continuing it on the fast path would misrepresent what was reviewed.

**Resolve `jiraIssue` and move the issue to In Progress — the run has begun.** Read it from the
state file when one exists; for a **new** change, resolve it per **Resolution** under **Jira
integration** in `rules/myflow-manual-review.mdc` — that section is canonical and its rules are
**not** restated here in any form; follow it there. Name the change `<key>-<slug>` when one is
linked and record it on the first state write in step 2.

**This stage's row: In Progress, when the run begins.** The transition mechanism is the same
canonical section. On a checkpoint resume the issue is already In Progress, so this is a no-op
report.

### 1. Write minimal OpenSpec artifacts

Ensure `openspec/changes/<name>/` exists and holds **at most two new files**:

**This step is strictly non-destructive.** If `openspec/changes/<name>/` already exists — the normal
case for a change entering at `proposal-done` — **never delete, truncate, or overwrite any file in
it.** Read what is there and add only what is missing. A change at `proposal-done` arrived via
`/myflow-start` → Gate A → `/myflow-start-done`, so its `proposal.md` is **human-approved** and its
`design.md` and delta specs were written deliberately and reviewed at Gate A. All of them stay
exactly as they are; "tops them up" means *adds what is absent*, never *replaces what is present*.

- **`proposal.md`** — `## Why` / `## What Changes` / `## Impact`. Terse: this documents *why*, not a
  diff. Write it only if the file does not already exist.
- **`tasks.md`** — checkbox tasks meeting **writing-plans** quality: exact file paths, exact
  verification commands, no placeholders. Group by module, not by finding or by layer. If it
  already exists, extend it rather than rewriting it.

**Do not author a `design.md` yourself** — if you find yourself needing to write one, that is an
escalation trigger (step 4a). An **existing** `design.md` is not a problem and is never removed.

**Delta specs:** only when the change introduces genuinely new or changed *intended* behaviour. A
feature that fits an existing requirement needs no spec edit. When it does, add the `ADDED`/
`MODIFIED` requirement the way `/myflow-start` would — and note that touching a delta spec is an
escalation trigger.

Do **not** publish a proposal artifact and do **not** wait for Gate A.

### 2. Confirm the task list once

Show the user `tasks.md` (and the one-paragraph Why) and ask to proceed. This is the only planning
checkpoint on this path — the cheapest possible place to catch a misread requirement.

Run the **step 4a eligibility check** now, before any code, and fold its result into this question.

On confirmation, write `stage: proposal-done` with `fastPath: true`.

**On a step 4a hand-off to the standard pipeline here, also write `stage: proposal-done` before
stopping** — that is the stage `/myflow-do` requires, and it is what the escalation handoff text
promises the user. But do **not** write `fastPath: true` on a handoff: the change never took the
fast path, and that field exists to record how a change reached `awaiting-pr-review`.

**Gate values on either step-2 write.** Nothing has been reviewed, tested, or opened at this point,
so write all four gates as `null` — `reviewed`, `tested`, `prOpened`, `prMerged`. `null` means "that
stage not reached", which is exactly true here. Do **not** write `reviewed: false` or
`tested: "skipped"` yet: those record a gate the human deliberately skipped, and on a step-2 handoff
the standard pipeline is about to run those gates for real.

### 3. Create the worktree

**REQUIRED SUB-SKILL:** `superpowers:using-git-worktrees` (Basic Workflow **#2**), branch
`openspec/<name>` — identical to `/myflow-do`. Record `MERGE_BASE=$(git rev-parse HEAD)` before any
edit; the review panel in step 5 diffs against it.

**Persist the worktree coordinates immediately, in the same write.** Record into the state file:

- `worktree` — the absolute worktree path (the schema already has this field)
- `branch` — `openspec/<name>` (already in the schema)
- `MERGE_BASE` — the commit recorded above, as an object **keyed by absolute worktree path** (see
  **State file** in `skills/myflow-contracts/state-file.md` for the shape)

Note this differs from `/myflow-do` and `/myflow-do-fix`, which record `MERGE_BASE` in the SDD
progress ledger at `.superpowers/sdd/progress-<name>.md`. The fast path keeps no ledger, so the
state file is its only durable place to put it.

Without these, a later session — the checkpoint resume in step 6a especially — cannot run
`git diff MERGE_BASE` and would silently review nothing.

**When the change spans repos** (more than one app in scope): `worktree` and `branch` name the
**primary** repo only, and every affected worktree gets its own entry in `MERGE_BASE`, keyed by
absolute path — the same keying `REVIEWED_TREE` uses. The key set of `MERGE_BASE` is the
authoritative list of affected worktrees; step 6a, the resume comparison, and step 7 all iterate it.

If the state file's `worktree` is missing or stale, fall back to resolving it from
`git worktree list` (branch `openspec/<name>`) per **State self-heal** in `skills/myflow-contracts/state-self-heal.md`, and rewrite the field.

### 4. Implement inline with TDD

**REQUIRED SUB-SKILL:** `superpowers:test-driven-development`. The main agent implements directly —
**no implementer subagents**. Per task in `tasks.md`:

1. **RED** — write the failing test, run it, confirm it fails for the expected reason.
2. **GREEN** — minimal code to pass. Run the test.
3. **REFACTOR** — clean up with tests green.
4. Check the box `[x]`.

**Test scope while iterating:** take the command(s) from `## test` in
`<main checkout>/.myflow/project.md`, using its narrower iteration subset when it names one; when
that file or key is absent, **auto-detect from the repository** (build files, `package.json`
scripts, CI config) and announce what you detected from. Still **write** tests and implementations
for every platform the change touches; only their execution is deferred. Never substitute a task
name remembered from another project — if nothing resolves, say so and ask.

**REQUIRED READING:** [../openspec-apply-superpowers/engineering-principles.md](../openspec-apply-superpowers/engineering-principles.md)
— your implementation must satisfy these principles; the review panel's principles reviewer checks
the diff against them. This applies whether the task is implemented inline by the main agent or
handed to a subagent, and it is **not** waived because the change is small — pass the same line
verbatim in any implementer dispatch this command makes.

**No commits during implementation.** The first commit happens in step 7 (or step 6a's continue
path), after review is clean. Do not `git commit`, `git push`, merge, or open a PR here.

Delete any implementation code written before its test, per TDD.

### 4a. Escalation check (run twice)

Run this **at step 2** (from `tasks.md` and the paths it names, before any code) and **again after
step 5** (from the actual diff). Triggers:

- more than **3** tasks in `tasks.md`. **Count every checkbox line at any nesting depth, checked or
  unchecked** — not top-level groupings, not headings. Two agents counting differently is itself a
  defect; count checkboxes, which is unambiguous:
  `grep -cE '^[[:space:]]*- \[[ x]\]' openspec/changes/<name>/tasks.md`
  Counting only `- [ ]` is wrong: this check also runs after step 5, when tasks are already checked
  off, and would then undercount to near zero precisely when it is re-verifying scope drift.
- touches a public contract or port boundary as defined by the project's standards (the files under
  `## standards` in `.myflow/project.md`, or auto-detection when none are configured) — never a
  package layout remembered from another project
- touches a DB migration, wherever this project keeps them (resolve from the repository: the
  migration tool's configured directory, or the layout the project's standards mandate)
- touches a delta spec
- the review panel raised a **Critical** finding

On any trigger, **stop and ask** — never escalate silently, and never silently stay fast:

```
## Fast Path — Escalation Trigger

**Change:** <name>
**Trigger:** <which one, with the concrete evidence — file paths, task count, or finding ID>

The fast path is scoped to small, well-understood changes. This one crossed that line.
```

**AskUserQuestion:**

1. **Hand off to the standard pipeline** (recommended) — stop here; the change is at a real stage
   with honest gates, so `/myflow-do <name>` (from `proposal-done`) or `/myflow-do-fix <name>`
   (from `awaiting-do-review` or later) picks it up with no special-casing.
2. **Continue on the fast path** — proceed, and record the overridden trigger in the final report.

Escalating at step 2 costs nothing but the artifacts already written — which the standard pipeline
reuses. That is the cheap moment to catch this, which is why the check runs there first.

### 5. Lean review panel

Write the whole-branch diff, then dispatch the **three required** review agents in parallel:

```bash
git diff MERGE_BASE > .superpowers/sdd/fast-path-review.diff
```

| # | Role | How to spawn | Portable fallback |
|---|------|--------------|-------------------|
| 0 | **Primary** — plan alignment + code quality | `superpowers:requesting-code-review`; pass the diff + `tasks.md` | same |
| 1 | **Bugbot** — defect hunt | `subagent_type: bugbot`, `description: "Bugbot"`, `Diff: uncommitted changes`, `Full Repository Path: <worktree>` | `generalPurpose` + [../openspec-apply-superpowers/bug-hunter-reviewer-prompt.md](../openspec-apply-superpowers/bug-hunter-reviewer-prompt.md) |
| 2 | **Principles** — merged principle list + project hard invariants | `generalPurpose` + [../openspec-apply-superpowers/principles-reviewer-prompt.md](../openspec-apply-superpowers/principles-reviewer-prompt.md) with `[LENS]` = **Merged**; **omit** `model` (inherit parent) | same |

These are the same three required slots the standard panel always runs — the fast path cuts the
*conditional* slots (Security, Adversarial, extra principle lenses), not the required ones.

**Resolve `[PRINCIPLES_PATH]` and `[STANDARDS_PATHS]` before dispatching slot 2.** Both are
mandatory pre-dispatch steps, defined once under **Resolve `[PRINCIPLES_PATH]`** and **Resolve
`[STANDARDS_PATHS]`** in [../openspec-apply-superpowers/SKILL.md](../openspec-apply-superpowers/SKILL.md)
— canonical, and followed there rather than restated here, exactly as the `full-panel` paragraph
below defers for the roster. Dispatching with either placeholder unresolved is a defect, not a
degraded run: the subagent's working directory is the **project worktree**, which has no `skills/`
tree, so an unresolved `[PRINCIPLES_PATH]` produces a principles reviewer holding **no principles**
that still reports "Principles-compliant? Yes". Confirm the principles file exists before spawning;
if it does not, stop and report it rather than dispatching a blind reviewer.

With **`full-panel`**, run every slot instead, exactly as defined in
[../openspec-apply-superpowers/SKILL.md](../openspec-apply-superpowers/SKILL.md) — same roster, same
prompts, same optional-slot triggers, same economic-model mapping for slots 4 and 5+. Reference those
prompt files by path; never copy them.

**Aggregation:** union Critical/Important findings, dedupe by file:line + theme, hand the combined
list to one fix pass, then re-run every slot that ran. **Handoff is blocked while any slot reports an open
Critical or Important finding.** Record results in `.superpowers/sdd/final-review-panel.md` with
`mode: fast-path-lean` (or `fast-path-full`) and the diff path reviewed.

**Re-run the full step 4a escalation check against this diff now** — all five triggers, not just the Critical-finding one. Step 2's check ran against what `tasks.md` predicted; this one runs against what actually landed, which is the only way scope drift discovered during implementation (a contract or port boundary crossed, a migration added, a delta spec edited) gets caught.

### 6. Verify

Run the project's tests and linters in **every** affected repo/worktree, each from its own absolute
worktree root:

```bash
cd <absolute worktree root>      # repeat for each affected repo
<the project's test command>     # from `## test`, or auto-detected
<the project's lint-fix command> # from `## lint`, when it names one
<the project's lint command>     # from `## lint`, or auto-detected — never scoped
```

Resolve both from `<main checkout>/.myflow/project.md` (`## test`, `## lint`); when that file or a
key is absent, **auto-detect from the repository** and announce what you detected from. Both keys
and the fallback are defined once under **Project configuration** in `skills/myflow-contracts/project-configuration.md`
— canonical, not restated here.

Unlike step 4's iteration subset, verification here runs the **full** test command, not the narrow
one. Lint must pass with zero violations: fix, never suppress and never weaken config. Run the
auto-fix command first when the project names one.

If `checkpoint` was passed, go to **6a** before step 7.

### 6a. Checkpoint stop (only with `checkpoint`)

The stop comes **after** verification, not before — the human should be reading a diff whose tests
and lint already pass, not one that may still change.

Run this **in every affected repo/worktree** — the `MERGE_BASE` key set recorded in step 3, which is
the same set step 6 verified:

```bash
cd <worktree-or-repo>
git add -A
git status
git diff --cached --stat
git write-tree
```

Record the output of `git write-tree` in the state file as `REVIEWED_TREE` — **one entry per repo**,
keyed by that repo's absolute worktree path, so each repo's tree can be compared independently on
resume.

Write `stage: awaiting-do-review` (gates unchanged), then present:

```
## Fast Path — Checkpoint

**Change:** <name>
**Review:** primary ✓ bugbot ✓ principles ✓ (clean)   **Tests:** ✓   **Lint:** ✓
**Git state:** staged + uncommitted (nothing pushed)

**Open in IntelliJ:**
open -na "IntelliJ IDEA" --args "<absolute worktree path>"

Review with: git diff --cached

**To resume:** re-invoke `/myflow-fast-path <name>`.
**Do not run `/myflow-do-done`** — this change is on the fast path; `/myflow-do-done` would
strand it outside this command's resume path and discard the recorded review state.
```

Then **AskUserQuestion**:

1. **Continue — commit and open the PR** (recommended)
2. **Stop here — I'll come back to it**

**On "Continue":** write `stage: do-done` with `gates.reviewed: false`, then proceed to step 7.

This is the human's confirmation made in the moment, so writing `do-done` directly is honoring it —
the same reasoning as **Crossing Gate B/C on an explicit "Continue"** in the rules file. Note the
gate stays `false`: what the human confirmed is "this is good enough to push", not the full Gate B
review. **Never invoke `/myflow-do-done`.**

**On "Stop here":** leave the change at `awaiting-do-review` and end the run. Report that re-running
`/myflow-fast-path <name>` resumes from step 7. A review that takes hours must not require a live
session.

**Resume path.** When `/myflow-fast-path` is invoked on a change at `awaiting-do-review` that has
`fastPath: true`, skip steps 1–6 and go straight to step 7 — the work is already implemented,
reviewed, verified, and staged. **Before resuming, prove the diff has not moved since the review.** Record `REVIEWED_TREE` (the output of `git write-tree` after staging) in the state file at the checkpoint stop. On resume, recompute it and compare — **in every affected repo/worktree, once per recorded `REVIEWED_TREE` entry**:

```bash
cd <worktree-or-repo>           # repeat for each repo recorded at the checkpoint
git add -A                      # pick up anything the human left unstaged
git write-tree                  # compare against that repo's REVIEWED_TREE from the state file
```

If **any** repo's tree hash differs, the human changed something during their review — **re-run step 5's panel and step 6's verification against the updated diff before landing.** Only an exact match in *every* repo may skip to step 7.

**A missing, empty, or unparseable `REVIEWED_TREE` — for the change as a whole or for any one affected repo — counts as a MISMATCH, not as a match:** re-run step 5's panel and step 6's verification before landing. Absence of evidence is never evidence that the diff is unchanged.

`git status` alone is **not** sufficient here: a human who edits *and stages* their change leaves a clean-looking status while the diff has in fact moved, which would push unreviewed code to the PR. Compare tree hashes, not status output.

If the change is at `awaiting-do-review` **without** `fastPath: true`, it came from `/myflow-do` —
refuse per step 0 and recommend `/myflow-do-done`.

### 7. Land — commit, push, open the PR

Only after step 5 is clean and step 6 passes:

Commit and push **in every affected repo/worktree** — the `MERGE_BASE` key set recorded in step 3
(the same set step 6 verified and step 6a staged). A repo that was verified
but never committed and pushed would be missing from the PR while the final report claims the change
is complete:

```bash
cd <worktree-or-repo>           # repeat for each affected repo
git add -A
git commit -m "feat: <what the change delivers>"
git push -u origin openspec/<name>
```

**Opening the PR is forge-agnostic.** Follow **PR opening is forge-agnostic** under **Review** in
`rules/myflow-manual-review.mdc` exactly, with all three of its cases — that section is canonical
and is **not** restated here in simplified form. In particular: `gh` being installed does **not**
mean the remote is a GitHub host, and `gh pr create` is a *write*, so it requires both conditions
the rules file names. Open one PR per affected repo.

**The stage may only be written once a PR actually exists.** Per that section:

- **Case 1** (`gh` installed **and** the remote is a GitHub host) → PR created → write
  `stage: awaiting-pr-review` with `gates.prOpened: true`.
- **Case 2** (remote exists, no usable PR CLI for that host) → push, print the forge's create-PR
  URL, **AskUserQuestion** "Have you opened the PR?" (default **No**). **Yes** → write
  `stage: awaiting-pr-review` with `gates.prOpened: true` (record the URL if given). **No** → do
  **not** advance the stage and do **not** set `prOpened: true`; leave the change where it is with
  `gates.prOpened: false`, say plainly what to do next, and end the run.
- **Case 3** (no remote at all) → push is impossible; stop, do not advance the stage,
  `gates.prOpened: false`.

On the branches that do write `stage: awaiting-pr-review`, write alongside it:

- `gates.prOpened: true` — a PR exists; leaving it `null` would mean "stage not reached", which
  contradicts the stage written beside it.
- `gates.tested: "skipped"` — Gate C was never run.
- `gates.reviewed: false` — **only when it is currently `null` or `false`.** Gates are monotonic:
  if the state file already carries `gates.reviewed: true`, carry that `true` forward untouched.
  (Step 0 permits an explicit user override from a later stage, so `true` is reachable here.)
- `fastPath: true`.
- `jiraIssue` — carried forward exactly as read.

Every other gate is carried forward unchanged.

**This stage's second row: In Review, after the PR is confirmed open and after the state write
above** — the state must be recorded whether or not Jira answers. The mechanism is defined once
under **Jira integration** in `skills/myflow-contracts/jira-integration.md`; follow it there. Fires on those same
branches only (cases 1 and 2-**Yes**); cases 2-**No** and 3 open no PR, so they make **no** Jira
call.

**Never merge.** Never force-push.

Final report — emitted **only** on a branch where a PR actually exists (cases 2-**No** and 3 report
the pushed branch and what the user must do instead):

```
## Fast Path Complete — PR Open

**Change:** <name>
**Mode:** fast-path (lean panel | full-panel) (checkpoint: yes | no)
**Artifacts:** openspec/changes/<name>/ (proposal.md, tasks.md<, delta specs>)
**Tasks:** N/N complete, TDD per task
**Review:** primary ✓ bugbot ✓ principles ✓ (clean)
**Tests:** ✓ (commands run, per repo)  **Lint:** ✓
**Gates:** reviewed: false (Gate B skipped) · tested: "skipped" (Gate C skipped)
**Stage:** awaiting-pr-review
**PR:** <URL>
**Jira:** <KEY> → In Progress → In Review | <KEY> already In Review (no transition) | none linked | ⚠ Jira: skipped — <reason>

**Next steps:**
- Review and merge the PR (Gate D) — never merged by this command
- Something to fix on the PR? `/myflow-do-fix <name>` (Gate D origin: commits, pushes, full panel)
- After merge: `/myflow-review-done <name>` → `/myflow-finish <name>`
```

## Guardrails

- **Never** accept `automerge`; **never** merge a PR.
- **Never** write `gates.reviewed: true` or `gates.tested: true` — this path skips Gate B (unless
  `checkpoint`) and always skips Gate C. Equally, **never lower an existing `true`** — gates are
  monotonic, so `gates.reviewed: false` is written only over `null` or `false`.
- **Never** write `stage: awaiting-pr-review` unless a PR actually exists (step 7, cases 1 and
  2-**Yes**), and never delete or overwrite an existing artifact in `openspec/changes/<name>/`.
- **Never** escalate silently, in either direction — step 4a always stops and asks.
- **Never** commit before the review panel is clean and verification passes.
- **Never** implement outside a worktree, and never on `develop`.
- **Never** dispatch implementer subagents — inline TDD is what makes this path fast. If the work
  is big enough to want subagents, that is exactly the escalation signal in step 4a.
- **Never** copy reviewer prompt files — reference `../openspec-apply-superpowers/` by path.
- **Never** stage, commit, or push the state file; it lives outside the repo.
- **Refuse** when the change is past `proposal-done`, except the `fastPath: true` resume at
  `awaiting-do-review`.
- **Always** write `fastPath: true` so downstream commands can report how the change got here.
- Do not skip TDD, and do not weaken lint to pass.

## Commands (user-facing)

| Intent | Say |
|--------|-----|
| Small, well-understood feature, one session | `/myflow-fast-path <name>` |
| Same, but let me eyeball the diff before it pushes | `/myflow-fast-path <name> checkpoint` |
| Same, but I want every reviewer slot | `/myflow-fast-path <name> full-panel` |
| Design needs exploring first | `/myflow-start <name>` |
| Something found on the PR | `/myflow-do-fix <name>` |
| After the PR merges | `/myflow-review-done <name>` → `/myflow-finish <name>` |
