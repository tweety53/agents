---
name: myflow-do
description: Implement a spectre change with Superpowers TDD and a multi-agent review panel, committing each task as it completes and printing the run instructions in the handoff for one human gate. Re-run to apply a fix. Use for /myflow-do.
allowed-tools: Bash(spectre:*)
license: MIT
---

Implement a spectre change, committing each task as it completes, and print the run instructions
in the handoff for the human gate at `IN_PROGRESS`: the operator reviews the diff and runs the
apps. **Never pushes, merges, or opens a PR** — unless a PR already exists, which is the one
exception below.

**Announce at start:** "Using myflow-do for change `<name>`."

Immediately after that line, print these two commands for the operator to paste, per
**Handoff output** (`skills/myflow-contracts/pipeline.md`) — that section fixes the colour and
records why they are printed rather than invoked:

```text
/rename <change-name>
/color cyan
```

**Load `skills/myflow-contracts/pipeline.md` first.**

**Then register this run's steps** with the harness's task-list mechanism, before any work begins,
and keep each entry's status current as the run proceeds, per **Progress visibility**
(`skills/myflow-contracts/pipeline.md`), which names which steps this command registers. Specific to
this command: an entry moves to in-progress when its implementer is dispatched, and to completed
when that task passes **both** its spec and quality review — the same moment its `tasks.md` checkbox
is allowed to be ticked, so the progress view and the file never disagree.

The reasoning behind this file lives in `skills/myflow-do/SKILL-rationale.md`; **a
`/myflow-*` run never loads it.**

## State gate

**Generate this run's session token once, right now, before the first mark below — a short, unique
literal string — and reuse that exact same value, unchanged, at every `stage begin` this run makes
below.** Never mint a fresh token per mark (design.md's "one token per session, not one per mark").

```bash
myflow stage begin -command '/myflow-do' -stage do.state-gate -harness <harness> -session-token mf-<literal-token> <name>
```

Accepts **`STARTED`** (first run) or **`IN_PROGRESS`** (fix run).

- From `STARTED`: create the worktree, implement the plan, end at `IN_PROGRESS`.
- From `IN_PROGRESS`: resume the **existing** worktree, apply a fix, and **write the state back
  unchanged** — a fix never moves the state.

At `FINISHED` the change is archived; emit the wrong-state handoff and stop.

## Superpowers Basic Workflow

| Step | Skill | When |
|------|-------|------|
| **2** | **superpowers:using-git-worktrees** | Before the first code change, on a first run |
| **3** | **superpowers:writing-plans** | Validate the plan; repair `tasks.md` if it is not apply-ready |
| **4** | **superpowers:subagent-driven-development** | Execute the remaining tasks |
| **5** | **superpowers:test-driven-development** | Every implementer dispatch, every task |
| **6** | **superpowers:requesting-code-review** + the review panel | Per-task review, then the final whole-branch panel |
| **7** | **superpowers:systematic-debugging** | An unexpected test failure during implementation, or a review-panel finding confirmed as a real defect |
| **8** | **superpowers:verification-before-completion** | Evidence before claiming done |

**Never** invoke `finishing-a-development-branch` — integration is `/myflow-finish`'s job.

```bash
myflow stage end -command '/myflow-do' -stage do.state-gate -outcome completed <name>
```

## 1. Load context and validate the plan

```bash
myflow stage begin -command '/myflow-do' -stage do.load-context -harness <harness> -session-token mf-<literal-token> <name>
spectre validate "<name>"
```

Exit `0` is the only exit that proceeds. Exit `1` names findings in this change's own artifacts —
most often a step checkbox left at column 0, which spectre reads as a malformed task line — and each
is repaired here, before any code is touched, exactly as a plan defect is. Exit `2` is a usage or IO
error, and `no such change "<name>"` is the one worth naming: nothing has been proposed under that
name, so stop and suggest `/myflow-start <name>`.

**The change root is `<project>/spectre/changes/<name>/`, by construction.** Nothing reports it and
nothing needs to: under spectre the layout is the contract, so it is derived rather than asked for.
**Read these files, and this list is the whole of it** — no command supplies one:

- `<changeRoot>/proposal.md` — what and why
- `<changeRoot>/tasks.md` — the plan
- `<changeRoot>/design.md` — how, when the change carries one; a change may legitimately carry none
- `<project>/spectre/specs/<capability>.md` for every capability the proposal names — one flat file
  per capability, edited directly on this change's branch, never a delta

**Whether there is anything left to implement is read off the task checkboxes**, which
`spectre list --json` reports per change as `done` and `total` in its
`{"changes":[{"id","done","total"}]}` output:

- `total == 0` → the change has no plan spectre can read: stop, and suggest `/myflow-start <name>`
  to write one.
- `total > 0` and `done == total` → every task is already checked: suggest `/myflow-finish <name>`.

**Check guard presence.** Per **Guard presence check** (`skills/myflow-contracts/pipeline.md`),
confirm every guard this command invokes — `check-panel-diff-size.sh`,
`check-panel-reproducers.sh`, `check-task-commit-fields.sh`, `check-unfinished-work.sh`,
`commit-split.sh`, `gather-dispatch-context.sh`, `plan-dispatch-bundles.sh`, `prepare-workspace.sh`
and `run-reproducer.sh` — is present in `skills/myflow-do/scripts/`.
A complete set prints nothing; any absence prints that section's block once, and the run continues
under each guard's own hand-run fallback.

Confirm `tasks.md` meets **writing-plans** quality — exact paths, verification commands,
bite-sized steps, no placeholders. If it does not, invoke **superpowers:writing-plans** to repair
it before touching code.

Extract the **Global constraints** verbatim from the capability specs the proposal names and
`design.md` for the reviewers.

```bash
myflow stage end -command '/myflow-do' -stage do.load-context -outcome completed <name>
```

## 2. Isolate the workspace (first run only)

**Load `skills/myflow-contracts/artifacts-registry.md`** — the worktree and branch this step
creates are rows in it.

**This stage runs, and is marked, on a first run only** — its own name in the Level 1 table carries
`*(first run only)*`. A fix run resumes the existing worktree instead, per section 3 below, and
marks nothing here:

```bash
myflow stage begin -command '/myflow-do' -stage do.isolate-workspace -harness <harness> -session-token mf-<literal-token> <name>
```

Invoke **superpowers:using-git-worktrees**. Branch `spectre/<name>`. Never implement on the
default branch without explicit consent. Record each worktree's merge base and absolute path in
this run's own working notes as soon as the worktree exists — the state file's `worktrees` map is
written only at the end of section 7. See **2. Isolate the workspace (first run only)**
(`skills/myflow-do/SKILL-rationale.md`) for why the working notes come first.

On a fix run, resume the existing worktree. **Never create a second one.**

**This run's resolved worktree set — the set section 7's guard iterates — is the worktree just
created or resumed above, plus any additional worktree this change affects.** Per **Resolving a
change's worktrees** (`skills/myflow-contracts/worktree-resolution.md`), how a command resolves the set beyond
reading the state file's map is that command's own; this is `/myflow-do`'s, and it is non-empty by
construction on every ordinary run — section 7's empty-set stop is for the genuinely anomalous case
where this step produced no worktree at all. See **2. Isolate the workspace (first run only)**
(`skills/myflow-do/SKILL-rationale.md`) for why.

**Then compute this worktree's workspace id from the change name.** The derivation is stated once
under **The workspace id** (`skills/myflow-contracts/workspace-isolation.md`) — do not re-derive it
by hand. Compute it once per run, on a fix run exactly as on the first. See
**2. Isolate the workspace (first run only)** (`skills/myflow-do/SKILL-rationale.md`) for why.

The main checkout has no id, and a project that declares no isolation at all is that same case
wherever it runs: every value resolves to the project's declared default, and neither is reported as
a misconfiguration. See **The empty id** (`skills/myflow-contracts/workspace-isolation.md`).

On a first run, close the stage once the worktree exists and the workspace id is computed:

```bash
myflow stage end -command '/myflow-do' -stage do.isolate-workspace -outcome completed <name>
```

## 3. Documenting a fix, before implementing it

**This stage runs, and is marked, on a fix run only** — its own name in the Level 1 table carries
`*(re-runs only)*`. A first run creates the worktree instead, per section 2 above, and marks
nothing here:

```bash
myflow stage begin -command '/myflow-do' -stage do.document-fix -harness <harness> -session-token mf-<literal-token> <name>
```

On a fix run, record what changed **before** writing code, so the proposal never goes stale. Ask
which of exactly two, with named options rather than open prose — shape per Operator prompts
(`skills/myflow-contracts/operator-prompts.md`):

> **This fix has to be recorded before it is written — where should it go?**
> - **Append to `proposal.md` and `tasks.md`** *(default, recommended)* — the fix is recorded in the
>   change's own artifacts; nothing new is created, and the plan stays one file
> - **Create a linked nested `<name>-fix-N` sub-change** — its own proposal and plan, for a fix that
>   adds scope the parent change does not describe

If the fix adds scope the linked Jira issue does not describe, sync the issue **description** per
**Description sync** in Jira integration (`skills/myflow-contracts/jira-integration.md`). Never
transition the issue here.

On a fix run, close the stage once the fix is recorded:

```bash
myflow stage end -command '/myflow-do' -stage do.document-fix -outcome completed <name>
```

## 4. Execute (SDD + TDD)

**Load `skills/myflow-contracts/model-policy.md`** before dispatching an implementer, below.

```bash
myflow stage begin -command '/myflow-do' -stage do.sdd-tdd -harness <harness> -session-token mf-<literal-token> <name>
```

**Gather the dispatch context bundle before dispatching any implementer.** Create the bundle's
directory as part of this same step — never rely on an earlier stage having created it, which is
exactly what left this redirect targeting a directory nothing had yet created on a fresh worktree,
since `superpowers:subagent-driven-development` (whose own workspace script creates
`<abs-worktree>/.superpowers/sdd/`) is not invoked until later in this section. Run

```bash
mkdir -p <worktree>/.superpowers/sdd
gather-dispatch-context.sh <worktree> <changeRoot> <name> <principles-path> \
  > <worktree>/.superpowers/sdd/dispatch-context.md
```

where `<changeRoot>` is `<project>/spectre/changes/<name>/`, the path section 1 derives, resolved
inside this worktree, and `<principles-path>` is the same absolute path of
`engineering-principles.md` that
`[PRINCIPLES_PATH]` (section 5) names — resolve it here too, since implementers dispatch before the
principles slot does. A non-zero exit — including the guard being absent — is reported, and
dispatching proceeds with the prompt shape this stage used before this capability existed; the bundle
never gates a run. **Confirm the bundle was actually written** — `test -f
<worktree>/.superpowers/sdd/dispatch-context.md` — and if it is not, report that plainly rather than
letting the run continue silently: this capability's failure mode is a silent fallback to the
pre-capability prompt shape, and only a visible report distinguishes that from an ordinary run. The
missing bundle still never gates or stops the run.

**At most one implementer subagent may be in flight against a given worktree at any moment.** The
parent waits for the previous implementer's commit sha for that worktree before dispatching the
next implementer into it; dispatches into different worktrees remain free to run concurrently. This
explicitly overrides `superpowers:subagent-driven-development`'s parallel dispatch guidance and
`superpowers:dispatching-parallel-agents` for same-worktree tasks, alongside the model-policy
override this section already carries against the same upstream skill. See `design.md` for the
concurrency failures — assertions left red at file seams, an agent idling on another's mid-edit
compile, corrupted test-result XML — this rule closes.

**The parent records each dispatch in two calls — one as it goes out, one as it comes back.
Nothing in this run writes a ledger file.** Immediately before dispatching, run

```bash
myflow record dispatch begin -change <name> -task <n> -role implementer -model <m> \
  -key task-<n>-implementer -session-token mf-<literal-token> -started-at <ts>
```

and as soon as that dispatch reports back, before the next one goes out, run

```bash
myflow record dispatch end -change <name> -key task-<n>-implementer \
  -session-token mf-<literal-token> -commit <sha> -outcome completed -ended-at <ts> \
  -agent-id <id>
```

**Both calls are required, and `begin` must go out BEFORE the dispatch does — never delayed to
obtain an identifier.** The harvester
commits the transcript offset it has consumed every few seconds and never re-reads behind it, so
each transcript record is offered to cost attribution exactly once — at the tick that consumed it.
A dispatch row that appears only at the close was therefore missing for every tick that ran while
the subagent worked, and all of that usage was dropped or credited to an unrelated earlier
dispatch. A dispatch runs for minutes; a tick is seconds. Symmetrically, a row never closed is an
open attribution window, and an open window goes on claiming its successors' tokens forever.

`-key` is this dispatch's own literal label, unique within the run its `-session-token` names. It
is what `end` closes and what makes a replayed write land on one row rather than two, so **write it
as a literal and reuse the identical string in both calls** — `task-<n>-implementer` for an
implementer, and the analogous form for each role below. `-role` is one of `implementer`,
`reviewer`, `panel-fix` or `red-partner`; `-task` is the task's id — the flat integer on its
`tasks.md` task line, per the `Placement` paragraph under **The build-green tag**
(`skills/myflow-contracts/build-green.md`) — omitted for a dispatch that ran against no single
task; `-started-at` and `-ended-at` are
that dispatch's own start and end, RFC 3339. **`-session-token` takes a literal, never a shell
substitution**, exactly as every `myflow stage` call above it does, and both halves take the same
one. `-agent-id` goes on `end` here, not `begin`: a serialized implementer dispatch reports its own
identifier only once it comes back, unlike a panel slot's concurrent launch — see **`-agent-id`**
under **5. The review panel** below for what the flag carries and why an absent one is never
invented.

**The model named here is the recorded intent Model policy requires**, not a figure derived after
the fact — see **Model policy** (`skills/myflow-contracts/model-policy.md`), canonical for it. Name the
model the dispatch was actually given. **A slot whose model the dispatcher cannot read records the
literal `unknown (agent-defined)` and never a guess**: a slot dispatched by `subagent_type` resolves
its model from an agent definition this pipeline does not read, and a plausible-looking slug written
for it puts an unmeasured value into the audit trail.

**A record write never blocks.** An unreachable store journals the intent, prints one warning line
and exits 0, so **never branch on this command's exit code as a signal about the record** — its only
non-zero exits are caller mistakes (a missing required flag, an unrecognised `-role`, a
`-session-token` carrying a substitution). What makes a journalled write visible is the handoff's
`Records:` line, defined once under **The block each state renders**
(`skills/myflow-contracts/handoff-blocks.md`).

The rows are the record; the SDD ledger is a rendering of them, written into
`<project>/docs/superpowers/ledgers/` by `myflow record render` at finish run 1. See the `SDD ledger`
and `Rendered ledger and panel record` rows under **Temporary artifacts registry**
(`skills/myflow-contracts/artifacts-registry.md`), canonical for both.

Invoke **superpowers:subagent-driven-development**, dispatching one implementer per bundle from

```bash
plan-dispatch-bundles.sh <changeRoot>/tasks.md
```

using the same `<changeRoot>` the gather invocation above already resolved. Exit 0 proceeds
to dispatch. A non-zero exit is a plan defect, not a review finding:
exit 1 names a task missing its `**Files:**` field, which `superpowers:writing-plans` repairs
before any dispatch happens; exit 2 stops the run. Bundling does not change the commit-per-task
model — an implementer handed a bundle still makes one commit per task, carrying that task's own
`Task-Id:` trailer, and a `Build: red` task still folds into the commit its `**Squash-with:**`
field names. Every implementer dispatch **must** carry each of the six blocks below:

> **MYFLOW — COMMIT-PER-TASK:** Do **not** run `git push`, merge, or open a PR. As soon as
> RED-GREEN-REFACTOR completes for this task — before the parent dispatches review for it — commit
> your work with `git commit`, carrying a `Task-Id: <n>` trailer where `<n>` is this task's id from
> its `tasks.md` task line. **The trailer identifies the task; the subject is this task's declared
> `**Commit:**` field, reproduced exactly** — already what `check-task-commit-fields.sh`
> enforces. **Never weaken or bypass a project's commit validation to fit** — no `--no-verify`, and
> no edit to its commit-message validator; a rejected subject means writing one the project accepts.
> You **may** `git add`/`git commit` your own work, but never `<project>/spectre/` or
> `<project>/docs/superpowers/` — `/myflow-finish` stages and commits those.

**A `Build: red` task's commit folds into its green partner.** A task tagged `Build: red` also
carries `**Squash-with:** Task <N>`, naming the green partner whose commit it folds into. Once that
partner task has its own commit, fold the red task's commit into it using the same
fixup-and-autosquash mechanism used for fix rounds (see "Panel re-runs" below): `git commit
--fixup=<partner-task-sha>` followed by `git rebase --autosquash`, where `<partner-task-sha>` is the
green partner's own commit — the one named by the red task's `Squash-with:` field.

**A `Build: red` task's own dispatch records `-role red-partner`, not `implementer`.** That role
exists to mark exactly this case: a dispatch whose work ends up carrying no commit of its own,
because it was folded into the green partner's. Record it as a pair like any other, with `-task`
its own id and the end call's `-commit` the green partner's sha as it stands after the fold:

```bash
myflow record dispatch begin -change <name> -task <n> -role red-partner -model <m> \
  -key task-<n>-red-partner -session-token mf-<literal-token> -started-at <ts>
myflow record dispatch end -change <name> -key task-<n>-red-partner \
  -session-token mf-<literal-token> -commit <partner-task-sha> -outcome completed -ended-at <ts> \
  -agent-id <id>
```

> **REQUIRED SUB-SKILL:** Use superpowers:test-driven-development — RED-GREEN-REFACTOR for this
> task. Delete any code written before its test.

> **REQUIRED SUB-SKILL:** When a test fails for a reason RED-GREEN-REFACTOR did not plan, invoke
> superpowers:systematic-debugging before writing a fix. An expected RED step needs no invocation.

> **REQUIRED READING:** [engineering-principles.md](engineering-principles.md) — your
> implementation must satisfy these principles; the panel's principles reviewer checks the diff
> against them.

> **CONTEXT BUNDLE:** `<abs-worktree>/.superpowers/sdd/dispatch-context.md` carries this change's
> proposal, design, plan and engineering principles, gathered for you — you need not go looking
> for them. You may open any file it names. You **must** still read the actual diff and the
> actual code you are reviewing or changing: the bundle is shared *input*, never a substitute for the
> source, and never a shared conclusion.

> **PLAN PROVENANCE:** a fenced block tagged `unverified:` is a hypothesis, not code to transcribe.
> Establish the real API before writing against it, and report what you found. A block tagged
> `verified:<how>` was checked as stated; if it does not compile, report that — do not contort the
> code to match it.

**Dispatch every implementer on the model the state file records under `models.implementation`**,
defaulting to Opus (or the harness's strongest available model) when that field is absent or null.
Name it explicitly — never by omission, which silently inherits the parent's model. This
**overrides** subagent-driven-development's "least powerful model that can handle each role"
guidance; see **Model policy** in `skills/myflow-contracts/model-policy.md` for why. The panel's slots
default to Sonnet — the two rules differ on purpose.

**Guard the commit before dispatching review.** As soon as the implementer reports the task's
commit sha back, and **before** the parent dispatches that task for review, the parent runs

```bash
check-task-commit-fields.sh <worktree> <task-id> <task-sha> <task-base>
```

naming the worktree path, this task's id from its `tasks.md` task line, the commit sha the
implementer just made, and the commit the task started from. A nonzero exit is a guard failure,
not a review finding — per `myflow-task-commit-fields`'s requirement **A runtime guard checks each
field against the real commit** — so it does **not** consume one of the review loop's fix-round
slots. The parent sends the task back to the **same implementer** (never a fresh one, never the
reviewer) to correct the mismatched field or the commit, then re-runs the guard before proceeding.
Only a clean exit clears the task for the dispatch described below.

**When the script cannot be located** — the same two cases as `check-workspace-isolation.sh` in
section 7: a harness whose repository does not carry it, or a skill directory copied rather than
linked — apply `myflow-task-commit-fields`'s rules by hand: check the commit's `Files:` against
`git diff --name-only <task-base>..<task-sha>`, its `Tests:` against the commit's diff, and its
`Commit:` against the commit's actual subject line. The check is never skipped for want of the
script, and a mismatch found by hand sends the task back to the same implementer exactly as a guard
failure would.

**Per-task review:** the parent gives the reviewer the commit-range diff
`git diff <task-base>..<task-sha>`, where `<task-base>` is the commit the task started from and
`<task-sha>` is the task's own commit — a real commit diff, never a snapshot of the uncommitted
working tree. If a fix round folds a fixup into that commit, via `git commit --fixup=<task-sha>`
followed by `git rebase --autosquash` (see "Panel re-runs" below), the
range still resolves as `<task-base>..<task-sha>`, now pointing at the rewritten, fixup-folded
commit. **Record the reviewer's dispatch too** — `-role reviewer`, the same `-task <n>`, and its own
`-model` — so implementer and reviewer alike leave a row. Nothing writes a ledger line by hand: the
model is each row's own `-model` field, and the per-task review shape is read off the rows
themselves, one reviewer row where a combined reviewer ran and two under `full`, so no second
statement of it can disagree with them. Mark a **task's** checkbox `[x]` only after that task passes
spec **and** quality review; a step's checkbox tracks the step and gates nothing.

**The per-task review's shape depends on `reviewPanelRoster`.** Under `light` and `standard`, a
**single** combined reviewer per task covers spec compliance and code quality together, dispatched
on `models.reviewPanel`. Under `full`, the spec-compliance and code-quality reviewers both run. See
the roster table in section 5 for what each preset means.

On BLOCKED: pause and report. Never guess.

```bash
myflow stage end -command '/myflow-do' -stage do.sdd-tdd -outcome completed <name>
```

## 5. The review panel

**Load `skills/myflow-contracts/model-policy.md`** before dispatching a panel slot, below.

```bash
myflow stage begin -command '/myflow-do' -stage do.review-panel -harness <harness> -session-token mf-<literal-token> <name>
```

**Rebuild the dispatch context bundle at the start of this stage too** — never reused from section
4's run. Create the directory again as part of this same step, same as section 4 — never assumed
still there from an earlier stage. Run

```bash
mkdir -p <worktree>/.superpowers/sdd
gather-dispatch-context.sh <worktree> <changeRoot> <name> <principles-path> \
  > <worktree>/.superpowers/sdd/dispatch-context.md
```

overwriting the same path. The bundle is rebuilt at the start of every dispatching stage, never
gathered once per run: a fix documented under section 3 edits `proposal.md` and `tasks.md`, and a
run-scoped bundle would leave every later dispatch reading a plan that no longer exists.

**Read `reviewPanelRoster` from the state file before selecting slots**, defaulting to `light` when
the field is absent or null. It names the preset in force for this run, per
**State file** (`skills/myflow-contracts/state-file.md`).

| Preset | Required slots |
|--------|----------------|
| `light` *(default)* | Primary · Principles · Code review (low) |
| `standard` | Primary · Principles · Bugbot |
| `full` | Primary · Bugbot · Principles |

Every preset dispatches exactly three required slots, and no preset reduces that number.

**No preset moves the handoff bar.** A preset selects how much reading the panel does and nothing
else: handoff still requires zero open findings at any severity under every preset, a minor finding
still blocks exactly as a critical one does, and the escalation ladder, fix-round rules, panel
record format, marker-line rules and operator handback are all unchanged. See **5. The review
panel** (`skills/myflow-do/SKILL-rationale.md`) for why.

**Before writing `final-review.diff`**, run

```bash
check-panel-diff-size.sh <worktree> <merge-base>
```

Exit 0 proceeds. Exit 1 puts the choice to the operator, shaped per **Operator prompts**
(`skills/myflow-contracts/operator-prompts.md`) — that section is canonical for the mechanics:

> **The panel diff measured `<count>`, over the `<cap>` cap. How should this proceed?**
> - **Proceed with the panel anyway** *(default, recommended)*
> - **Stop and split the change** — ends the run at `IN_PROGRESS` with the implementation committed
>   on the branch

Exit 2 stops the run: a size the guard could not measure is not a size under the cap.

Record the measured count, the cap in force, and the operator's answer where one was given in
`<abs-worktree>/.superpowers/sdd/final-review-panel.md` on **every** run, including exit-0 runs. The cap moves
nothing about the roster, the slots, the escalation ladder or the zero-open-findings bar.

Write `<abs-worktree>/.superpowers/sdd/final-review.diff` from `git diff <merge-base>` (staged and unstaged), then
dispatch **separate** review subagents — one per selected slot, in **every** affected worktree.
Never merge two slots into one prompt.

**Every slot the panel spawns directly runs on the model the state file records under
`models.reviewPanel`, defaulting to Sonnet** when that field is absent or null. There is no
parent-model inheritance and no economy tier. See **5. The review panel**
(`skills/myflow-do/SKILL-rationale.md`) for why.

| # | Slot | Required? | Model | How to spawn |
|---|------|-----------|-------|--------------|
| 0 | **Primary** — plan alignment + code quality | **always** | `models.reviewPanel` | **superpowers:requesting-code-review** with `final-review.diff` + the plan/spec constraints |
| 1 | **Bugbot** — defect hunt | `standard`, `full` | its own | `subagent_type: bugbot`, `Diff: uncommitted changes`, `Full Repository Path: <worktree>`, + the mutation-testing brief below |
| 2 | **Principles** | **always** | `models.reviewPanel` | general-purpose + [principles-reviewer-prompt.md](principles-reviewer-prompt.md), `[LENS]` = **Merged** |
| 3 | **Code review (low)** | `light` | `models.reviewPanel` | general-purpose reviewer briefed for high-confidence defects only, against `final-review.diff` |
| 4 | **Security** | conditional | its own | `subagent_type: security-review`, same shape as Bugbot |
| 5 | **Adversarial** | conditional | `models.reviewPanel` | general-purpose + [adversarial-reviewer-prompt.md](adversarial-reviewer-prompt.md) |
| 6+ | **Principles lens B / lens C** | conditional | `models.reviewPanel` | same template, `[LENS]` = **Lens B — simplicity & state** or **Lens C — robustness & ops** |

Slots 1 and 4 are dispatched by `subagent_type` and carry their own agent definitions — pass them
**no** model override, whatever `models.reviewPanel` records, and record `unknown (agent-defined)`
as their dispatch row's `-model`. Every other slot, slot 3 included, names its model explicitly.

**Every slot's dispatch is recorded, exactly as section 4 records an implementer's.** The panel is
where most of a run's dispatches happen, so a panel that recorded none of its own would leave the
audit trail describing a minority of the run. The same pair per slot — `begin` before the slot is
dispatched, `end` at its close, and for the same reasons section 4 gives:

```bash
myflow record dispatch begin -change <name> -role reviewer -slot <slot> -model <m> \
  -agent-id <id> -diff-base <sha> -key panel-<round>-<slot> \
  -session-token mf-<literal-token> -started-at <ts>
myflow record dispatch end -change <name> -key panel-<round>-<slot> \
  -session-token mf-<literal-token> -outcome completed -ended-at <ts> -agent-id <id>
```

`-slot` is the slot's own name from the table above — `Primary`, `Bugbot`, `Principles`,
`Code review (low)`, `Security`, `Adversarial`, `Lens B` or `Lens C` — so the rows themselves say
which slots ran, and nothing has to state it a second time. `-role` is `reviewer` for every one of
them, and `-task` is omitted: a panel slot runs against no single task.
`-started-at` and `-ended-at` are that slot's own start and end, RFC 3339. `-key` is
`panel-<round>-<slot>`, the same literal in both halves. **`-session-token` takes a literal, never a
shell substitution**, exactly as the `myflow stage begin` that opened this stage does, and both
halves take the same one.

**`-diff-base <sha>` is the sha a delta-slot's delta starts from**, passed on a slot dispatched
against a delta and on no other: a slot reading the whole `final-review.diff` passes none, and an
absent base means *not recorded* rather than a base of nothing. **Scheduling reads the dispatcher's
own in-session value of each slot's last-reviewed sha, never the store** — the row is the durable
audit trail, and nothing waits on it, which is what keeps this flag inside *A record write never
blocks*, stated earlier in this section.

**`-model` is recorded intent here too, per Model policy.** A slot the dispatcher named a model for
records that model; **slots 1 and 4, dispatched by `subagent_type`, record the literal
`unknown (agent-defined)` and never a plausible-looking slug**, for the reason the paragraph above
gives. Slot 3 names its real model.

**On Claude Code, `-agent-id` is the identifier an asynchronous agent launch returns in the parent's
own tool result, at launch — `agentId: a392afd1eacbdfebc` — and the dispatcher reads it from there
rather than hoping the harness exposes one.** This is the flag's whole reason for existing: the
panel dispatches its slots concurrently against one parent session, so their time windows overlap,
and the window rule alone credits every record in the overlap to whichever slot started last.
Two of the three supported harnesses expose no id at all — a dispatch recorded without one is
ordinary, not degraded, and its cost is attributed by its own window instead. **Never invent one.**
An absent id means *not reported* and never matches another absent id, so a placeholder would pair
two unrelated slots together.

**Record a slot's dispatch before recording that slot's findings**, and carry the seq the command
printed — `recorded: dispatch <seq>` — into each of that slot's `myflow record finding` calls as
`-dispatch-seq <seq>`. Every panel finding names the dispatch that raised it; a finding left without
one is a finding **no single dispatch raised**, and that is the only thing the absence ever means.

**Every slot the panel dispatches must supply, per finding, a reproducer**: a runnable command that
demonstrates the defect, or the literal exemption form `none — <reason>`. Carry this requirement on
every slot's dispatch prompt. Slots dispatched by `subagent_type` (Bugbot, Security) receive it as
prompt text, the same way Bugbot already receives the mutation-testing brief below — no agent
definition is edited to carry it.

**Every slot's dispatch prompt also carries the CONTEXT BUNDLE paragraph:**

> **CONTEXT BUNDLE:** `<abs-worktree>/.superpowers/sdd/dispatch-context.md` carries this change's
> proposal, design, plan and engineering principles, gathered for you — you need not go looking
> for them. You may open any file it names. You **must** still read the actual diff and the
> actual code you are reviewing or changing: the bundle is shared *input*, never a substitute for the
> source, and never a shared conclusion.

Slots dispatched by `subagent_type` (Bugbot, Security) receive it as prompt text too, exactly as they
already receive the reproducer requirement and the mutation-testing brief — no agent definition is
edited to carry it.

### No forking, and a wall-clock ceiling on every slot

**No panel slot SHALL be dispatched onto a skill or agent that forks its own background agent.** The
parent dispatcher never observes a forked agent's completion, so a slot dispatched that way cannot
report through the panel's contract: its findings land on a surface the panel does not read, or they
never arrive at all, and either way the slot neither returns a result nor reliably ends. The repair
is to dispatch that slot on a shape that reports back to the dispatcher directly. **Never drop the
slot** — *No preset moves the handoff bar*, earlier in this section, already forbids it.

**Every panel slot SHALL carry a 15-minute wall-clock ceiling from its dispatch.** The dispatcher
SHALL NOT block indefinitely on a slot's completion notification; it tracks each in-flight slot's
elapsed time and enforces the ceiling itself, so a slot emitting output while making no progress is
still bounded. This is stated against the mechanism, not against one harness's tool — the same form
**Progress visibility** (`skills/myflow-contracts/pipeline.md`) uses — so no harness has to gain a
particular tool to satisfy it; a dispatcher waiting on a notification alone has dropped the
requirement rather than adapted it.

On a breach, in order:

1. Stop the slot.
2. Close its dispatch row: `myflow record dispatch end … -outcome timed-out`.
3. Record the breach in the panel record, naming the slot and its elapsed time.
4. Re-dispatch that one slot once. Other slots are unaffected.

A second breach of the same slot is never resolved silently. Stop and put it to the operator, shape
per **Operator prompts** (`skills/myflow-contracts/operator-prompts.md`):

> **Slot `<slot>` breached the wall-clock ceiling a second time. How should this proceed?**
> - **Re-dispatch it again**
> - **Proceed without the slot**
> - **Stop the run** *(default, recommended)* — ends at `IN_PROGRESS` with the implementation
>   committed on the branch

A timed-out slot raises no finding, consumes no fix round, and is not a clean result for the final
pass. Where the operator chooses to proceed without it, the panel record names that slot as **not
run** — distinct from a slot the operator declined and from a slot whose trigger never fired, both of
which **Optional slot selection**, below, already distinguishes.

### Code review (low)

Slot 3, the `light` preset's third required slot, is unconditionally a `general-purpose` subagent on
`models.reviewPanel`, briefed to report high-confidence defects only against
`<abs-worktree>/.superpowers/sdd/final-review.diff` in the worktree. It invokes no skill, so there is
no harness-availability condition on it and nothing for the panel record to name as a substitution.
Its findings are ordinary `F<n>` rows and marker lines, exactly like every other slot's. Because the
dispatcher names the model explicitly, this slot's dispatch row records that real model and never
`unknown (agent-defined)`, which is reserved for slots dispatched by `subagent_type`.

### Bugbot's mutation-testing brief

Wherever the panel dispatches Bugbot, its dispatch prompt carries a mutation-testing brief: for
each behaviour the diff changes, mutate it — flip a condition, drop a guard, move a boundary, remove
a branch — and establish whether an existing test fails. A mutation no test catches is a
**surviving mutant**. This is reasoned mutation testing performed by the reviewer: no
mutation-testing framework is added, adopted or executed, and no mutation score is computed.

A surviving mutant is an ordinary finding — an `F<n>` row and a marker line — and blocks the handoff
under the existing zero-open-findings bar until a test is added or the operator withdraws it with a
reason. It is not an advisory note outside the findings table.

The brief applies wherever Bugbot is dispatched, and nowhere else — no other slot acquires it, and
it adds no slot to any preset. Carrying it on the dispatch prompt changes neither Bugbot's
`subagent_type` dispatch nor the `unknown (agent-defined)` its dispatch row records.

Slot 2 is the panel's only mandatory judgment check on *how* the code is built. It reads
`engineering-principles.md` — never a pasted copy — and owns the project's **hard invariants** from
its standards files: architecture and layer purity, new suppressions, weakened lint config.

**Resolve `[PRINCIPLES_PATH]` before dispatching any principles slot.** It is the **absolute** path
of `engineering-principles.md` **beside this file** — `skills/myflow-do/`, always, including when a
composite command such as `/myflow-fast` is the one running this section. Under a global install
that is `~/.claude/skills/myflow-do/engineering-principles.md`. The three reviewer-prompt files this
section links resolve the same way, beside this file.

**It is deliberately not resolved against the running command's own skill directory.** That rule is
**Guard resolution** (`skills/myflow-contracts/pipeline.md`), it governs `<skill-dir>/scripts/` and nothing else,
and applying it here sends a `/myflow-fast` run to `skills/myflow-fast/`, which carries no such file
— after which the fix looks like symlinking one in, rather than reading the sentence above.

Confirm the file exists before spawning; if it does not, stop and report rather than dispatching a
blind reviewer. See **5. The review panel** (`skills/myflow-do/SKILL-rationale.md`) for why it must
be absolute.

**Resolve `[STANDARDS_PATHS]` before dispatching slot 2**, from the `## standards` entries in
`<project>/.myflow/project.md`. Entries are **not** paths to use as-is: each resolves through the
entry-form table and the containment rule in **Project configuration**
(`skills/myflow-contracts/project-configuration.md`), and an entry failing either is reported by
name and dropped. Resolve that contract file by **absolute** path too; if it is not readable,
**stop** — do not resolve entries without the containment rule. Pass an **empty** value when none
resolve; that correctly empties the Hard Invariants section rather than substituting another
project's standards. Record which standards files were passed, or that none resolved. See
**5. The review panel** (`skills/myflow-do/SKILL-rationale.md`) for why the containment rule is
non-negotiable here.

**No two principle reviewers may share a lens.**

**Every finding is a row in the store. The panel record is rendered from those rows, and no part of
it is written by hand.** As each slot raises a finding, record it — one call, as it is raised:

```bash
myflow record finding -change <name> -ref F<n> -round <r> -slot <slot> -severity <sev> \
  -location <file:line> -status open -reproducer <command | none — reason> \
  -dispatch-seq <seq> -note <the finding>
```

`-dispatch-seq` is the seq the raising slot's own `myflow record dispatch begin` call printed,
which is another reason that call comes before the dispatch rather than after it. A `begin` that
fell back to the journal printed no seq — the store allocates it, and no store was reached — so
omit `-dispatch-seq` in that case rather than guessing a number. `-round` is `0` for the initial panel and `1..n` for a fix round. `-ref`
is unique within the change and the store's own constraint enforces it, so **a reused identifier is
refused at the write rather than found later on a read**; the command prints `recorded: F<n>` on
the insert and `updated: F<n>` when it replaced an existing row, which is how a finding just raised
is told from one a fix round restated. **A fix round updates the finding it resolved rather than
appending a second row:**

```bash
myflow record status -change <name> -ref F<n> -status fixed
```

**`-status` carries the whole status text the marker line shows**, so a withdrawal passes its
reason with it — `-status 'withdrawn <the operator's reason>'` — because the reason is checked for
on that same line and a withdrawal with nothing after the status does not clear the gate it appears
to.

**Render the record when the panel closes** — the point this section ends at, with every slot's
result clean and no finding open — before any guard reads it:

```bash
myflow record render -change <name> -kind panel -repo <abs-worktree>
```

**A panel that raised nothing still renders**, declaring `findings-total: 0`. There is no
skip-the-render-when-there-were-no-findings shortcut, and adding one is the exact defect this
sentence exists to prevent: zero findings is a declaration and clears, where an unwritten record is
silence and reads as outstanding. The render's own outcome words, and what to do with each, are the
table under **Rendering the session records** (`skills/myflow-contracts/session-records.md`), canonical for
them.

**Every rule about the record's format below is unchanged, and every one of them now binds the
renderer rather than the agent.** They are stated here because they are the contract the record must
satisfy, not because anybody types them: `<agents repo>/stats/internal/records/render.go` is the
only writer, and the guards read exactly what they read before. **Do not delete a format rule on the grounds that
nobody hand-writes the record any more** — deleting one removes the only statement of what the
renderer owes the guard, and the guard then fails on records nothing produces incorrectly.

The record carries a findings table, one row per finding, each beginning with its identifier:

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Bugbot | Minor | `src/Foo.kt:42` | replaced the silent catch |

and, below it, the marker block — one line per row, plus the count:

```
findings-total: 1
finding-status: F1 fixed
```

The marker format is never quoted inside the record itself — a validly-formatted marker written as
an example inside prose or a fenced block reads the same as a real one. The renderer honours this by
neutralising any marker label a finding's note or location happens to carry, on the way out only:
the store keeps the slot's own words verbatim, because a note is evidence and the rendering is what
can be regenerated.

**The reproducer each finding's slot supplied gets a marker block of its own**, separate from the
`finding-status:` block above — the two are kept apart because `<agents repo>/scripts/check-unfinished-work.sh`
requires the `finding-status:` lines to occupy one unbroken span, and interleaving the two blocks
would break that. The record carries one `reproducers-total: <n>` count line and one
`finding-reproducer: F<n> <command | none — reason>` line per finding:

```
reproducers-total: 1
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh
```

A finding recorded with no reproducer renders the `none — <reason>` exemption form rather than no
line at all: `check-panel-reproducers.sh` requires exactly one reproducer line per finding, and an
omitted line is reported as a missing one.

The same rule against quoting the format inside the record applies to this block as well.

**This quoting rule is one instance of a broader constraint**, binding these marker labels anywhere
in the record, not only in this block — see **The fix round mutation-proves what it changed** below
for the full label-collision rule.

`<agents repo>/scripts/check-unfinished-work.sh` reads **only the marker block**. It never parses the table: not
its header, not its column order, not its cell boundaries, not where it starts or stops. See
**5. The review panel** (`skills/myflow-do/SKILL-rationale.md`) for why.

**The rules the guard does enforce, each of which reports outstanding when broken.** Run
`check-unfinished-work.sh` and read its own output for the reject reason on any given record — it
already reports one in
human-readable form; the list below is not repeated here as a parse grammar:

- `<status>` is **exactly** `open`, `fixed` or `withdrawn`, compared byte for byte.
- `withdrawn` **carries its reason on the same line**.
- Each `F<n>` names **one** finding. A reused identifier is reported on each side separately.
- The marker lines sit on **consecutive lines**, one unbroken block.
- `findings-total: <n>` appears **exactly once** and equals the number of marker lines. A panel that
  raised nothing renders `findings-total: 0` and carries no markers.

**The table carries no status column, on purpose.** A finding's state is one column of its row and
is rendered once, on its marker line. To read a finding's state, look up its `F<n>` in the marker
block.

**A `withdrawn` marker's reason is checked for being there at all.** A withdrawal is the operator's
decision at the panel's handback, not a status a run may write for itself. Fix subagents record
`fixed` when they fixed it, and leave `open` when they did not — including when they believe the
finding is wrong, which belongs in the note and in the report back, never in the status.

See **5. The review panel** (`skills/myflow-do/SKILL-rationale.md`) for why each of the rules above
exists — the five marker-format rules, the missing status column, and the withdrawn-reason check.

### Optional slot selection

Evaluate against `final-review.diff` **before** dispatching, and record which optional slots were
included and which were excluded and why.

| Slot | Include when the diff touches | Ask when |
|------|-------------------------------|----------|
| 4 — Security | auth/authz, tokens, crypto, secrets or config, query construction, path or file handling, deserialization, CORS/HTTP edge, new dependencies | a config or dependency file changed, but only comments or a version bump |
| 5 — Adversarial | migrations, concurrency/scheduling, behavior changes to code with existing tests, any test modified or deleted, or **>~300** changed lines | **150–300** changed lines with no other trigger |
| 6 — Lens B (simplicity & state) | **>~200** changed lines, or **≥3** new classes/modules | — |
| 6 — Lens C (robustness & ops) | error handling, retries, schedulers, external integrations, config/env, migrations | — |

**Borderline → ask**, with **include** as the default. See **Optional slot selection**
(`skills/myflow-do/SKILL-rationale.md`) for why.

The triggers above fire the same way under every preset — nothing about evaluating the table
changes. What happens once a trigger fires depends on `reviewPanelRoster`:

**Under `full`**, a fired trigger auto-includes its slot, and the table's borderline *ask* rows
keep their behaviour there.

**Under `light` and `standard`**, every slot whose trigger fired goes into **one** multi-select
prompt instead of being auto-included or asked about individually:

> **These triggers fired on this diff. Which slots should the panel include?**
> - **Security** — <the trigger that fired>
> - **Adversarial** — <the trigger that fired>
> - **Lens B — simplicity & state** — <the trigger that fired>
> - **Lens C — robustness & ops** — <the trigger that fired>
>
> Including all of them is the recommended answer.

Only slots whose triggers actually fired appear in the prompt, and the prompt is not shown at all
when nothing fired.

Record which optional slots were included and which were excluded and why, under every preset. A
slot the operator declined is recorded as **declined**, distinctly from a slot whose trigger never
fired.

A documentation-, prompt-, or test-only diff with no trigger runs the three required slots alone.
That is a correct outcome, not a skipped review — say so explicitly.

### Panel re-runs

**Pass 1 always runs the full roster selected for this change.** Only re-runs after a fix are
scoped. Record `FIX_BASE=<task-sha>` — the task's commit as it stood before this fix round — then,
once the fix is folded into that commit via `git commit --fixup=<task-sha>` and
`git rebase --autosquash`, write `<abs-worktree>/.superpowers/sdd/fix-round-N.diff` from
`git diff "$FIX_BASE"..<task-sha>`, where `<task-sha>` now names the rewritten, fixup-folded
commit.

When a review finding requires a code change to a task that is already committed, commit the fix
as `git commit --fixup=<task-sha>`, where `<task-sha>` is the **original** task commit — the one
created when the task was first implemented, never the sha from a previous fixup-and-autosquash
round, so every fixup round for a given task targets the same commit. Immediately run
`git rebase --autosquash` to fold the fixup into that commit, before the next review pass reads the
diff — one commit per task throughout the panel's re-runs, never a trailing separate fixup commit.

| Mode | Who re-runs | Diff they get |
|------|-------------|---------------|
| **Targeted** (default) | Slot 0 (always, as integration check) + every agent that raised a finding | `fix-round-N.diff` |
| **Full** (escalation) | Every **required** slot; a **conditional** slot (Security, Adversarial, Lens B, Lens C) only when its own row in the optional-slot trigger table still fires against `fix-round-N.diff` | Slot 0 the rewritten `final-review.diff`; every other diff-reading slot its own delta, below |

**A delta is `git diff <the HEAD sha that slot last reviewed> HEAD`**, written to
`<abs-worktree>/.superpowers/sdd/slot-delta-<round>-<slot>.diff`. It is anchored at that slot's own
last read, never at the current round: **Targeted** mode re-runs slot 0 and the slots that raised
findings alone, so a slot reaching a Full pass may have missed rounds in between. Each dispatch sets
that slot's sha to the HEAD it was dispatched against — pass 1, a Targeted re-run and a Full-mode
delta dispatch alike — and a slot not dispatched in a round, including one skipped for an empty
delta, keeps the sha it had. The range is tree-to-tree and needs no ancestry between its ends, so a
base recorded before a `git rebase --autosquash` stays valid after that rebase rewrote the task
commit. A slot for which no last-reviewed sha is held reads the whole `final-review.diff` — an
unrecorded base is not a small delta. Bugbot and Security read no diff file and are unaffected.
Coverage holds because pass 1 always runs the full roster against the full `final-review.diff`, so
every slot has a real starting sha and every change made since sits inside some slot's delta, and
there is no exemption for the pass that closes the panel. **Every slot's dispatch prompt names the
path it was given and, for a delta, the sha that delta starts from** — that is what makes a
reviewer's "I did not see X" checkable after the fact rather than a matter of recollection.

**A required slot whose delta is empty is not dispatched**, and the record states `not re-run —
nothing new since its last read`. Slot 0 reads the whole diff, has no delta, and is never scoped out
by this. It says the slot's reading is current, never that its remit was waived.

A conditional slot whose trigger did not fire is **not** re-run; its previous result stands, and the
record states `not re-run — subject unchanged`, distinct from a slot whose trigger never fired at
all and from a slot the operator declined. The **Targeted** row is unchanged. Not re-running a slot
never closes, softens or expires a finding it has already raised — the zero-open-findings bar still
governs every slot in the roster. **Trigger-based** scoping reaches conditional slots alone; a
required slot is scoped, if at all, only by its own delta being empty — two mechanisms, two
recorded reasons. See **Panel re-runs** (`skills/myflow-do/SKILL-rationale.md`) for why.

**Escalate automatically** — do not ask, and say why in the record — when the fix touched a file
outside the set named in the findings; the fix altered a capability spec, a migration, or a guard's
behaviour; a targeted re-run surfaced a **new** Critical finding; three or more fix rounds have
already run; or the fix diff exceeds ~150 changed lines **and** adds a new file. Size alone carries
no risk signal — a mechanical rename is large and harmless, a one-line change to a guard's behaviour
is small and dangerous — so size never escalates on its own. **The signal it is paired with is `adds
a new file` and nothing else, because every other risk signal already escalates by itself**: a
capability spec, a migration, a guard's behaviour and a file outside the findings set each have
their own clause above, so repeating them here would add no case the ladder does not already reach. What the pairing
does reach is the one gap those clauses leave — a large body of brand-new, wholly unreviewed code in
a file the findings themselves named. See the capability spec for the longer argument. A
trigger that fires on every fix round of every change in a given repository selects nothing there,
so where a trigger is found not to discriminate, the correct repair is to narrow the condition, not
to remove the escalation.

Targeting is a cost optimization, never a coverage waiver: a targeted re-run is never fewer than
two agents, and handoff still requires **zero open findings at any severity** from every agent that
has run, with the final pass showing a non-stale clean result for every slot in the roster.

Union all **open** findings, dedupe by **defect identity — file:line + theme.** *File:line* is the
finding's own recorded location, taken verbatim from the findings table (the `Location` column),
never re-derived from a diff hunk that may have moved since. *Theme* is the finding's one-sentence
Note column, reduced to its own defect noun phrase: the shortest phrase in that sentence naming
what is wrong (e.g. "no `--` before the record path", "leading-dash worktree path"), with severity
words, slot names and prose padding stripped out — not the sentence itself, so a reviewer's
rewording of the same sentence still reduces to the same phrase. Two reviewers describing the same
defect at the same file:line, in different words, are the same identity; two different defects at
the same file:line are not. This is the one place the identity is defined; every rule below that
keys off it cites this paragraph.

**Worked example, on a compound note.** F20's Note reads: "swapping the two `comm` directions
mislabels every finding and still passes all 20 cases, because the assertions match on the
identifier substring rather than the message." Its three clauses are the root defect, a symptom,
and why testing missed it — only the first is the theme: "still passes all 20 cases" says the bug
went undetected, not what it is, and "the assertions match on the identifier substring" is itself a
separate later finding (F35) that folding in would wrongly merge into this one's identity. The
theme is the root cause alone: **swapped `comm` directions**.

**Before dispatching the fix subagent**, run

```bash
check-panel-reproducers.sh <worktree> <change>
```

**The change name is the second argument** because the record the guard reads is the one `myflow
record render -kind panel` wrote, at `<worktree>/docs/superpowers/reviews/<date>-<change>-panel.md`
— one file per change in a directory every change shares, so a worktree alone no longer names a
record. `<abs-worktree>/.superpowers/sdd/final-review-panel.md` is the pass log, carries no marker
block, and is read by neither panel guard.

Exit 0 proceeds. Exit 1 covers two classes, handled differently. A **missing or malformed field** —
no `finding-reproducer:` line for some `F<n>`, a bad `reproducers-total:` count — is a
record-completeness defect, and the field is added before dispatch. A
**rejected reproducer shape** — a command carrying a shell metacharacter, an absolute path, a `..`
segment, a leading `-` on its path token, a URL, or a NUL byte — is a **refusal, not something to
rewrite until the guard accepts it**: the line is recorded **unverifiable** and put to the operator,
the same disposition `<agents repo>/scripts/run-reproducer.sh` reaches for a refusal of its own below. It is
**never silently rewritten to satisfy the guard** — the line may be exactly the injection the guard
exists to catch, and rewriting it to pass defeats the check. Exit 2 stops the run: a record the
guard could not read is not a record in which every finding declared a reproducer.

**For each open finding whose record carries a runnable `finding-reproducer:` command** (not the
`none — <reason>` exemption form), run

```bash
run-reproducer.sh <worktree> "<the finding's finding-reproducer: text>"
```

which is what makes the constraints this section used to only describe in prose actually true
rather than merely claimed: the direct exec with an argument vector, the resolved containment on
every token, the bound, and the timed-out and surviving-process dispositions — its own header is
canonical for all of it, including the session-kill mechanism it verified by running candidates on
this platform rather than by citing a flag that turned out not to exist here (`ps -o sid=`). Read
its exit code. **0** dispatches the finding — the reproducer demonstrated the defect as a direct
exec, with no shell ever seeing the line. **1** bounces the finding once, back to the slot that
raised it, carrying the reproducer's passing output, rather than dispatching it — the instruction
built on a reproducer that exits 0 cannot be verified as a fix. **2** is a refusal: recorded
**unverifiable** and put to the operator, never run, the same disposition as a rejected shape from
`check-panel-reproducers.sh` above. **3** is a timeout or a detached survivor: recorded
**unverifiable** and put to the operator — with a **surviving process's pid named** when the
script's own output names one — and the worktree is re-checked (`git status`) before the run
continues, and again when the operator resumes. **4** stops this finding's dispatch decision
entirely, the same way an unreadable record does above. A finding recorded `none — <reason>` is
dispatched without a run: the rule binds findings claiming a mechanical defect, and a principles,
prose or naming finding has no runnable check to demand.

If a bounced finding's replacement reproducer also exits 1, the run stops and puts that finding to
the operator through the handback prompt below — take another round, withdraw it with a reason, or
stop the run — rather than dispatching it silently.

**Bounce accounting keys off defect identity, never off `F<n>`** — the same identity defined once
above under **Union all open findings**, so a fresh identifier assigned to the same defect in a
later pass does not reset its one-bounce budget to zero. Each bounce is recorded in the existing
pass log entry this section already requires below, keyed by that identity, so a run resumed after
an interruption sees the prior bounce and goes straight to the operator rather than bouncing again.

Where a slot **dispatched by `subagent_type`** supplies nothing at all for a finding, record
`none — not supplied by <slot>` and dispatch the finding unverified; this exemption is legal only
for such a slot, since the pipeline does not control a third-party agent's definition. A
**general-purpose** slot — one whose prompt this pipeline fully controls — that supplies nothing has
not supplied a legal exemption: record that omission as its own open finding rather than as
`none — not supplied by <slot>`, so the record distinguishes a genuine no-runnable-check exemption
from a slot that was simply never asked to comply.

**Once the fix subagent reports, re-run every dispatched finding's reproducer** under the same
constraints and require it now to exit **0**. A reproducer that still exits non-zero means the fix
did not fix it, and the finding is not closed on that pass — this also catches a reproducer that
failed both before and after the fix for a reason unrelated to it, since it is the same command
under the same constraints, not a fresh judgment call. **The flip alone does not close a
finding — the fix's diff must also touch at least one path the finding named, with a non-comment,
non-whitespace change in that hunk**, and preferably at the finding's own `file:line`. A fix whose
diff, for a given finding, touches only the reproducer's own target **and no path the finding
named**, or touches a named path with no non-comment, non-whitespace change, is not a fix: the
finding stays open and goes to the operator through the handback below, carrying that fact as the
reason. Where the reproducer's target *is* a path the finding named — the ordinary shape for a
finding about a guard script, whose reproducer is that script's own test harness or the script
itself — the fix is material, and the finding closes on the reproducer's flip. See **Panel
re-runs** (`skills/myflow-do/SKILL-rationale.md`) for why the re-run and the materiality condition
are both needed.

### The fix round mutation-proves what it changed

**This binds the review panel's fix round — this section (5) — and not the per-task review's fix
in section 4**, which uses the same words, "fix round", for a different mechanism: section 4's
fixup-and-autosquash on a single task's commit, reviewed by that task's own reviewer. The two are
never conflated below.

**Every executable behaviour the fix changed is mutation-proved, not only the test cases the round
adds.** When the fix subagent reports, it names the executable behaviours its fix changed — a guard
script, Go, TypeScript, shell, anything a test could fail on. **You** then mutate each one — revert
it in a scratch tree, or flip the single value it turns on — confirm an existing test fails, and
restore. The subagent does not certify its own mutations, for the same reason it does not re-run its
own reproducers one paragraph above. Whether a given change counts as an executable behaviour is
sometimes unclear (a refactor with no new branch, a comment beside live code); an unclear case goes
to the operator through the same handback used below for a finding that does not converge, rather
than the parent deciding it alone.

**Each mutation alters one mechanism.** Where a single revert would also change state a second check
reads, split it into surgical mutations, one per mechanism. A mutation touching shared state can
pass by cross-contamination — the case goes red because the *other* check broke, not because the
mutated mechanism works — and that reports coverage which does not exist, which is worse than
recording no mutation at all.

**A surviving mutation is repaired in this round.** Where no existing test fails, add the test that
catches it before the round closes. It is not raised as an `F<n>` finding and costs no extra pass:
the round has the behaviour in hand, and a finding would spend a full round recovering context it
has not lost. This is deliberately a different disposition from a surviving mutant Bugbot reports,
which is a reviewer's reading of a diff someone else wrote.

**Record each one in this pass's log entry**, beside the mode, the slots that ran and the diff path:

```text
fix-mutation: <path> — <what was mutated> — <the test that failed>
fix-mutation: <path> — none — <reason>
fix-mutations-total: <n>
```

One line per changed behaviour, using the same `none — <reason>` exemption form the record already
uses for `finding-reproducer:`. A round that changed only prose records the exemption rather than
recording nothing.

**These lines go in the pass log entry and never inside the marker block.**
`check-unfinished-work.sh` requires every `finding-status:` marker to occupy one unbroken run of
consecutive lines, and reads the findings table's identifiers off lines matching
`^\|?[[:space:]]*F[0-9]+[[:space:]]*\|`. A line that split that run, or that looked like a row,
changes that guard's verdict on a record this rule does not otherwise touch — which is why
`<agents repo>/scripts/test-check-unfinished-work.sh` carries a case for it.

**No line anywhere in the panel record may carry the literal label `finding-status:`,
`findings-total:`, or `finding-reproducer:` outside its own marker use — the constraint is the whole
record's, not one line's.** `<agents repo>/scripts/check-unfinished-work.sh` counts the first two unanchored —
`M_NAMED` (line 337), `T_NAMED` (line 342) — and `<agents repo>/scripts/check-panel-reproducers.sh` counts the
third the same way, `R_NAMED` (line 139): a substring match over every line, fences included, so a
"Format example" that merely *quotes* the marker grammar counts as a real marker line. This list is
derived from those guards' unanchored counts, not invented by hand — extend it by re-reading the
guards, never by guessing. `reproducers-total:` is counted too but anchored (`^`,
check-panel-reproducers.sh:282), so it is not a hazard the same way. The rule reaches past
`fix-mutation:` free text to any prose, and lands hardest on a round documenting the marker-parsing
logic itself. Write around it: paraphrase the label, or break it with a non-word character. The
failure is safe — a clean round reads `OUTSTANDING` rather than a gap reading clear — but still worth
avoiding by construction. `<agents repo>/scripts/test-check-unfinished-work.sh` case 12 pins the lines as inert; it
does not pin this constraint, which is why it is stated here in words.

**No guard reads these lines, and none is added.** What holds the rule instead is this: **the fix
round does not close, and the run does not reach the handoff, while an executable behaviour it
changed carries neither a line nor an exemption.** See **The fix round mutation-proves what it
changed** (`skills/myflow-do/SKILL-rationale.md`) for why the parent runs the mutations and why no
guard reads the record.

**The parent checks the reported list against the fix diff before the round can close** — the
report is the round's account of itself, and nothing above requires it to be checked against what
actually changed. Walk every hunk of the fix diff with a non-comment, non-whitespace change: each
one is either covered by a reported line — mutated, or exempted with a reason — or is not an
executable behaviour at all (prose, a comment, a rename with no behaviour change). A hunk that
changes executable behaviour and names neither a reported line nor an exemption is treated exactly
as an unproved behaviour — the round does not close until it is added to the list and mutated. A
hunk that removes or weakens a test or an assertion is a third case, neither of the two above: it is
executable, but nothing can be reverted and run to prove it unsafe, because the removed assertion
was itself the thing that would have failed. Mutation cannot cover it, so it takes a different
obligation: state in the record what the removed or weakened assertion used to cover, and name what
still covers that same behaviour now. That claim is checked the same way a mutation is — run the
named covering test against the **pre-fix** code and confirm it fails; a named test that passes
against the pre-fix code did not cover the behaviour, and the claim is false. Where nothing can be
named, restore the coverage instead of arguing it away.

This binds the fix round under every roster, `light` included — the obligation is the round's, not a
slot's, so a preset that dispatches no Bugbot is exactly where the round's own proof is the only
mutation reasoning that happens at all. It adds no slot to any preset.

**A bounce is a guard-class failure, not a review finding** — the accounting section 4 already
gives `check-task-commit-fields.sh`: no fix-round slot consumed, no round-count advance, and it
never closes, softens or expires the finding.

**Rebuild the dispatch context bundle before dispatching the fix subagent.** Create the directory
again as part of this same step, same as sections 4 and 5 — never assumed still there. Run

```bash
mkdir -p <worktree>/.superpowers/sdd
gather-dispatch-context.sh <worktree> <changeRoot> <name> <principles-path> \
  > <worktree>/.superpowers/sdd/dispatch-context.md
```

overwriting the same path — the plan may have changed since this stage's own start, per section 3's
fix documentation.

**Carry each surviving finding to the fix subagent as a structured block, not a bare restatement of
its prose.** For every finding in the union above, carry its `F<n>`, the slot that raised it, its
severity, its `file:line` taken verbatim from the findings table's Location column, its theme (as
defined above), the text of its `finding-reproducer:` line, and any bounce already recorded against
its defect identity. State plainly that these locations were established by the slot that raised
them and are not to be re-derived. **Inline no source excerpt** for any finding: the fix round edits
the code it is given, so an excerpt taken now would be invalidated by the fixer's own work before it
is even read — the fix subagent opens the named file itself.

> **CONTEXT BUNDLE:** `<abs-worktree>/.superpowers/sdd/dispatch-context.md` carries this change's
> proposal, design, plan and engineering principles, gathered for you — you need not go looking
> for them. You may open any file it names. You **must** still read the actual diff and the
> actual code you are reviewing or changing: the bundle is shared *input*, never a substitute for the
> source, and never a shared conclusion.

Give the surviving findings — every dispatched finding from the union above, carried as the
structured dossier above — to **one** fix subagent as the combined list. Where a finding is confirmed
as a real defect — as opposed to a style or principles nit — the fix subagent invokes
**superpowers:systematic-debugging** before writing its fix. **Dispatch it on the model recorded
under `models.panelFix`**, defaulting to Opus
(or the harness's strongest available model) when that field is absent or null — deliberately not
the panel's own default, for the reason stated under
**Model policy** in `skills/myflow-contracts/model-policy.md`. Record every pass in
`<abs-worktree>/.superpowers/sdd/final-review-panel.md`: mode, which agents ran, why, the diff path they read,
and — when this pass bounced any finding — each bounced finding's defect identity (file:line plus
theme, as defined above) together with the reproducer output it carried back to its raising slot.
A pass that bounced nothing states that plainly rather than omitting the field.

**The fix subagent's own dispatch is recorded too, with `-role panel-fix`** — the same pair, naming
the model `models.panelFix` resolved to:

```bash
myflow record dispatch begin -change <name> -role panel-fix -model <m> \
  -key panel-fix-<round> -session-token mf-<literal-token> -started-at <ts>
myflow record dispatch end -change <name> -key panel-fix-<round> \
  -session-token mf-<literal-token> -commit <partner-task-sha> -outcome completed -ended-at <ts> \
  -agent-id <id>
```

`-commit` is the task commit the fixup was folded into, since the fix leaves no commit of its own;
`-task` is omitted where the pass carried findings against more than one task, and given where it
carried them against one. `-agent-id` goes on `end` here, for the same reason as an implementer's:
the fix subagent is a single dispatch, not a panel slot's concurrent launch, so its identifier is
known only once it reports back — never invented. **Recording this dispatch is what keeps the fix round inside the run's
audit trail**: it runs on the strongest of the three roles' models, so a panel that recorded every
slot and not its fixer would understate precisely the dispatch that costs the most.

A minor finding blocks the handoff exactly as a critical one does. When fix rounds do not converge,
the run hands back to the operator, who resolves the disagreement — including by marking a finding
`withdrawn` with a reason. **The handback is an actual prompt, not a claim that one happened.** When a finding survives its
last fix round, stop and put it to the operator, one finding at a time, with the finding's text, the
fixer's reason for disputing it, and named options — shape per Operator prompts
(`skills/myflow-contracts/operator-prompts.md`):

> **`<location>` — <the finding, in one line>. The fix round did not resolve it.**
> - **Take another round on it** *(default, recommended)*
> - **Withdraw it — I'll give the reason** — the reason is recorded on the finding's marker line
> - **Stop the run and hand it back to me**

Only that answer records `withdrawn`, and only with the reason the operator gives, carried in the
`-status` text so the reason lands on the marker line. Nothing else in the run may record it: the
guard reads a `withdrawn` marker with nothing after the status as outstanding, so a withdrawal with
no stated reason does not clear the gate it appears to.

```bash
myflow stage end -command '/myflow-do' -stage do.review-panel -outcome completed <name>
```

## 6. Resolve the run instructions

```bash
myflow stage begin -command '/myflow-do' -stage do.run-instructions -harness <harness> -session-token mf-<literal-token> <name>
```

In the same run, resolve the run instructions for the handoff's `Run it:` section. It writes no
file. See **6. Resolve the run instructions** (`skills/myflow-do/SKILL-rationale.md`) for why.

Resolve:

- **Every app root is absolute**, resolved from `git worktree list` or the state file's `worktrees`
  keys. Never a relative sibling path (`../<other-app>`), and never a main-checkout path while a
  worktree holds the work.
- **Every start command comes from `<project>/.myflow/project.md`'s `## run`**, with every path in it made
  absolute.
- **Every URL is the one this worktree resolved**, never the project's declared base. **Resolve
  each URL from this worktree's workspace id, the way section 2 computed it.** The project's
  `## workspace isolation` rows name the variable each URL is carried by and what it becomes in a
  workspace, per **Project configuration** (`skills/myflow-contracts/project-configuration.md`);
  what a workspace id moves at all is listed under **What the id derives**
  (`skills/myflow-contracts/workspace-isolation.md`).

  A project that declares no isolation resolves nothing, so the handoff names that project's
  declared URLs unchanged.

  An application whose port is fixed outside that project's own repository keeps its default, per
  **Project configuration** (`skills/myflow-contracts/project-configuration.md`), and so does every
  URL built from that port. Name such a URL with a short note on the same line — the project's
  default, shared, and holdable by one workspace at a time — and leave the worktree's own URLs
  plain.
- Apps in scope come from `## apps` in `<project>/.myflow/project.md`, or from auto-detection
  when that file or key is absent — see
  **Project configuration** (`skills/myflow-contracts/project-configuration.md`).
- **Where the project declares no runnable application**, resolve the `## lint` and `## test`
  commands instead, with every path in them made absolute, and do not give the handoff an
  application shape the project does not have.

```bash
myflow stage end -command '/myflow-do' -stage do.run-instructions -outcome completed <name>
```

## 7. Verify, stage, and hand off

**Load `skills/myflow-contracts/worktree-resolution.md`** before resolving this run's worktree set,
below.

This section carries four of the Level 1 table's stages, marked one at a time as each opens and
closes rather than all at once, per **Stage marks** (`skills/myflow-contracts/pipeline.md`).

```bash
myflow stage begin -command '/myflow-do' \
  -stage do.workspace-export \
  -harness <harness> \
  -session-token mf-<literal-token> \
  <name>
```

**First, validate the section and export what it declares — with the script, not by eye.** Run

```bash
prepare-workspace.sh <worktree>
```

once per worktree in this run's resolved set — the same set section 2 resolved, non-empty by
construction — never a raw read of the state file's `worktrees` map, which reads as `{}` or absent
on every first run, before this run's own section-7 write, and would make this check pass having
examined nothing. Per **Resolving a change's worktrees** (`skills/myflow-contracts/worktree-resolution.md`),
report an empty resolved set and do not proceed to validate a worktree the state file cannot name;
see section 2 above for why that stop is the anomalous case. Run this **before anything else below
this line.**
`prepare-workspace.sh` resolves per **Guard resolution** (`skills/myflow-contracts/pipeline.md`) —
against this running command's own skill directory. Not finding it there is the absent-script case
below, and is reported rather than passed over.

`prepare-workspace.sh` runs `check-workspace-isolation.sh` against the worktree first, then — only
if that passes — derives and exports the variables the project's `## workspace isolation` section
declares, resolved against the workspace id section 2 computed, and prints one `KEY=value` line per
exported variable to stdout. Read its exit code for what happened and read those printed lines
rather than re-deriving any of them: exit 0 means the printed lines are what to carry forward into
`## lint` and `## test` below (an exit 0 with nothing printed means the project declares no
`## workspace isolation` section). A
non-zero exit is the dropped-row case (exit 1, relay the script's own lines verbatim and stop) or the
cannot-answer case (exit 2, stop the same way) — in either case, stop **before** `## lint` and
`## test`, without writing the state file. Correcting the row in `<project>/.myflow/project.md`
and re-running this command is the whole remedy.

**A declared `cache index` row is never among the printed `KEY=value` lines — the script reports it
by name on stderr instead, and claiming it is this step's own job, not the script's.** Per **The
cache index** (`skills/myflow-contracts/workspace-isolation.md`), the index is claimed by probing
the project's own cache, never derived from the workspace id. On an exit-0 run whose stderr names a
`cache index` row, probe the project's cache here, claim a free index atomically, and record that claim in the
cache itself under an entry naming this workspace — exactly as **The cache index** requires, and per
the registry's `Claimed cache index | /myflow-do, by probing, when it exports the workspace's
variables` row (`skills/myflow-contracts/pipeline.md`) — before carrying the exported lines forward
into `## lint` and `## test` below.

**When the script cannot be located**, apply the same rules by hand from
**Project configuration** (`skills/myflow-contracts/project-configuration.md`) and
**Workspace isolation** (`skills/myflow-contracts/workspace-isolation.md`), and say in the handoff
that the validation and export were performed manually and why. See
**7. Verify, stage, and hand off** (`skills/myflow-do/SKILL-rationale.md`) for the full procedure and
why the script-absent case and the guard-exit-2 case are treated as one.

**This step does not call the project's `create` command, and that is a decision rather than a
gap.** `create` is called by whatever starts the project's applications, per
**Project configuration** (`skills/myflow-contracts/project-configuration.md`), and this command
starts none of them — it exports, lints, tests, and hands off. The applications are started at the
review gate by the operator, through the project's own `## run` commands, which is where the
creation and its one-time notice belong. See **7. Verify, stage, and hand off**
(`skills/myflow-do/SKILL-rationale.md`) for why.

```bash
myflow stage end -command '/myflow-do' \
  -stage do.workspace-export \
  -outcome completed <name>
myflow stage begin -command '/myflow-do' -stage do.lint-and-test -harness <harness> -session-token mf-<literal-token> <name>
```

Run the `## lint` and `## test` commands from `<project>/.myflow/project.md` (auto-detect if
absent) and show the output. **Nothing runs them later** — `/myflow-finish` has no verification
gate — so a non-zero exit blocks this handoff.

**Load `skills/myflow-contracts/session-records.md`** before reading the render outcome below.

**Confirm this run recorded a ledger** — there is no file to test for, so ask the store, by
rendering the ledger this run's dispatch rows produce:

```bash
myflow record render -change <name> -kind ledger -repo <abs-worktree>
```

Read the outcome word, not the exit code. `rendered: <dest>` is the ordinary case. **`MISSING:
ledger — no rows for <name>` means this run recorded no dispatch at all**, and it is reported
plainly here, at this call site, rather than being discovered at finish run 1 or not at all — the
whole point of asking now. `journalled: ledger` and a non-zero exit are reported the same way, with
the command's own message. **Unlike the lint and test exits above, none of these gates or stops the
run**; each is reported and the handoff proceeds. The outcome words and what to do with each are the
table under **Rendering the session records** (`skills/myflow-contracts/session-records.md`), canonical for
them.

```bash
myflow stage end -command '/myflow-do' -stage do.lint-and-test \
  -outcome completed <name>
myflow stage begin -command '/myflow-do' -stage do.stage-diff -harness <harness> -session-token mf-<literal-token> <name>
```

Confirm every intended task checkbox is `[x]`, and that `git log <merge-base>..HEAD` shows one commit
per completed task, per section 4's commit-per-task model, with every fix-round and red-task-partner
fixup already folded in via `git rebase --autosquash` — no stray `fixup!` commit should remain
unsquashed on the branch, unless a PR already exists (below), in which case the section's one
additional commit also sits on top.

In **every** affected worktree:

```bash
git -C <worktree> status
git -C <worktree> log <merge-base>..HEAD --oneline
```

> **`<project>/spectre/` and `<project>/docs/superpowers/` are never part of a task commit.** Section 4's
> COMMIT-PER-TASK clause already excludes both paths from every task and fixup commit; this step
> only confirms nothing slipped in, it does not stage anything itself. See **7. Verify, stage, and
> hand off** (`skills/myflow-do/SKILL-rationale.md`) for why.

**Load `skills/myflow-contracts/git-boundaries.md`** before committing below.

**The one push exception.** Every task and fixup commit already sits on the branch, unpushed, per
sections 4 and 5. If the state file records a `prUrl`, a PR is already open, so this run also
commits `<project>/spectre/` and `<project>/docs/superpowers/` — the only paths a task or fixup commit never touches —
and pushes everything to the PR branch; otherwise this step commits and pushes nothing. On that path
only — and in this order — run
`myflow record render -change <name> -kind all -repo <worktree>`; then
`commit-split.sh <worktree> <name> "<impl-msg>" "chore(spectre): plan and session records"`;
then push the branch, which carries whatever that call committed along with every task and fixup
commit accumulated since the PR was opened. `<impl-msg>` covers the one case a task or fixup commit
does not: working-tree edits the operator made at the human gate without staging them — normally
none, in which case the script's own guard skips that commit, and only the planning commit lands.
When there is something to describe, derive `<impl-msg>` the same way a fixup commit's subject is
derived — `fix(<module>): <what changed since the last task commit>`, `<module>` naming the area
those edits touch. See **7. Verify, stage, and
hand off** (`skills/myflow-do/SKILL-rationale.md`) for why that ordering matters.

The render overwrites in place; the date is fixed at a change's first render, so a fix round refreshes
the same file rather than leaving one dated copy per round. `MISSING: <kind>` means the store holds
no rows of that kind and nothing was written — report it, and never read it as a record written
somewhere else. **A non-zero exit means a destination was refused or could not be written** — report
it with the command's own stderr message, and continue committing the fix; the remaining kind is
still attempted after any one failure. See **Rendering the session records**
(`skills/myflow-contracts/session-records.md`), canonical for every outcome.

**`commit-split.sh` is the same guarded chain run 1 uses** — the skipped-empty rule, the
stop-on-failure rule and the symlinked-planning-path case are all under **Git boundaries**
(`skills/myflow-contracts/git-boundaries.md`). The empty
case is ordinary here — a fix round that touched neither `<project>/spectre/` nor the test guide has nothing
to add — but say in the handoff which of the two commits, if either, was made.

```bash
myflow stage end   -command '/myflow-do' -stage do.stage-diff -outcome completed <name>
myflow stage begin -command '/myflow-do' -stage do.write-in-progress -harness <harness> -session-token mf-<literal-token> <name>
```

Write the state file: `IN_PROGRESS` from `STARTED`, otherwise **the state exactly as read**.
Populate `worktrees` with one absolute-path key per affected worktree and its merge base. Carry
`artifactUrl`, `jiraIssue`, `planningEffort`, `models`, `prUrl` and `reviewPanelRoster` forward
verbatim, per the carry-forward rule in **State file** (`skills/myflow-contracts/state-file.md`).
The state file lives outside the repo — never
`git add` it.

Read the planning effort through the retired-key fallback, not from `planningEffort` alone, per
**State file** (`skills/myflow-contracts/state-file.md`).

```bash
myflow stage end -command '/myflow-do' -stage do.write-in-progress -outcome completed <name>
```

**Produce the handoff's `Records:` count before printing the block**, one call per affected
worktree:

```bash
myflow record journal-count -change <name> -C <abs-worktree>
```

It prints one decimal count, or `unknown` where no count could be produced, and exits 0 either way.
Render what it printed — never a hand-derived path, never a substituted number for `unknown`. The
line's three alternatives are defined once under **The block each state renders**
(`skills/myflow-contracts/handoff-blocks.md`).

**Produce the handoff's `Costs:` line the same way**, one call per affected worktree:

```bash
myflow record cost-status -change <name>
```

It prints one line naming how many of this change's dispatches are unattributed and why, or that
none are, and exits 0 always — `unknown` included. Render exactly what it printed, never a
hand-derived figure.

```
## Implementation committed — review and test

**Change:** <name>
**Panel:** clean — roster: <light | standard | full>, required: <that roster's required slots, per the table under **5. The review panel**>; optional: <selected, or "none — no triggers fired">
**Staged:** N/N tasks committed on branch | committed, plus one planning-artifacts commit, and pushed to the PR branch
**Records:** all writes reached the store | N write(s) journalled — the store was unreachable | unknown — the journal could not be counted
**Costs:** <the line `myflow record cost-status` printed>
**Guards:** all present | N missing — those checks were performed by hand (see the guard presence check above)
**Jira description (pre-edit):** <the text as it stood before the write, verbatim in a fenced block, inside <details> when long> | omitted — this run wrote no description

Worktree:   <absolute worktree path>

Run it:
  <command>          # <app or check name>
  <command>

Review the diff, then run it:
  git -C <absolute worktree path> diff <merge base>..HEAD
  open -na "IntelliJ IDEA" --args "<absolute worktree path>"

Re-run this command to fix anything you find.

Next:
/myflow-finish <name>
```

**One review command covers both of this command's cases** — committed on branch with no PR yet,
and committed and pushed to an open PR — both leave the branch fully committed, so the same
commit-range diff reads either one. The merge base comes from this worktree's entry in the state
file's `worktrees` map. The template's third git state — committed and pushed with no PR — is one
`/myflow-do` never emits and `/myflow-status` does — see **The block each state renders**
(`skills/myflow-contracts/handoff-blocks.md`).

**The `Records` line is printed whether or not anything was journalled**, and its three alternatives
above are this command's enumeration of the placeholder the line carries under **The block each
state renders** (`skills/myflow-contracts/handoff-blocks.md`), canonical for it — including why a
run that journalled nothing still prints the line, and why the count is read from the change's
record journal rather than remembered from the run.

**The `Costs:` line is printed the same way — always, `unknown` included — and never suppressed by
`cost-status` itself failing to answer.** `cost-status` always exits 0, so it can never become the
reason the handoff does not print.

The pre-edit description line is present only on a fix run that synced the description in section
**3**, and reproduces that text without summarising or reflowing it. A run that wrote nothing omits
the line rather than printing an empty one. See **Description sync**
(`skills/myflow-contracts/jira-integration.md`).

## Guardrails

- **Commit per task and per fixup, as section 4 and section 5 require** — never `<project>/spectre/` or
  `<project>/docs/superpowers/` in a task or fixup commit. **Never push, merge, or open a PR** — except the
  `prUrl` exception above.
- **Never** run `finishing-a-development-branch`.
- **Never** create a second worktree for the same change.
- **Never** advance the state from `IN_PROGRESS`; write back what you read.
- **Never skip** a required panel slot, and never collapse two slots into one prompt.
- **Never dispatch an implementer without the provenance clause.**
- **Never** pass a model override to Bugbot or Security Review; **always** name the panel's model
  explicitly on every other slot.
- **Never** paste the principle list into a prompt — the reviewer reads the file.
- **Never** hand off with an open finding of any severity, or a stale clean result.
- **Never** mark a task's checkbox before that task's review passes.
- **No flags.** The only argument is the optional change name; report anything else.
