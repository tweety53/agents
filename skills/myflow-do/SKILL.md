---
name: myflow-do
description: Implement an OpenSpec change with Superpowers TDD and a multi-agent review panel, committing each task as it completes and printing the run instructions in the handoff for one human gate. Re-run to apply a fix. Use for /myflow-do.
allowed-tools: Bash(openspec:*)
license: MIT
---

Implement an OpenSpec change, committing each task as it completes, and print the run instructions
in the handoff for the human gate at `IN_PROGRESS`: the operator reviews the diff and runs the
apps. **Never pushes, merges, or opens a PR** — unless a PR already exists, which is the one
exception below.

**Announce at start:** "Using myflow-do for change `<name>`."

Immediately after that line, print these two commands for the operator to paste, per
**Handoff output** (`skills/myflow-contracts/pipeline.md`) — that section fixes the colour and
records why they are printed rather than invoked; do not restate its reasoning here:

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

Every citation below is canonical at its target. Never restate its content here and never act on
a remembered version of it — read it fresh each time it is needed.

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
openspec status --change "<name>" --json
openspec instructions apply --change "<name>" --json
```

- `state: "blocked"` → stop; suggest `/myflow-start <name>`.
- `state: "all_done"` → suggest `/myflow-finish <name>`.
- Read every path in `contextFiles`. Resolve paths from the CLI JSON, never from an assumed layout.

Confirm `tasks.md` meets **writing-plans** quality — exact paths, verification commands,
bite-sized steps, no placeholders. If it does not, invoke **superpowers:writing-plans** to repair
it before touching code.

Extract the **Global constraints** verbatim from the delta specs and `design.md` for the reviewers.

```bash
myflow stage end -command '/myflow-do' -stage do.load-context -outcome completed <name>
```

## 2. Isolate the workspace (first run only)

**This stage runs, and is marked, on a first run only** — its own name in the Level 1 table carries
`*(first run only)*`. A fix run resumes the existing worktree instead, per section 3 below, and
marks nothing here:

```bash
myflow stage begin -command '/myflow-do' -stage do.isolate-workspace -harness <harness> -session-token mf-<literal-token> <name>
```

Invoke **superpowers:using-git-worktrees**. Branch `openspec/<name>`. Never implement on the
default branch without explicit consent. Record each worktree's merge base and absolute path in
this run's own working notes as soon as the worktree exists — the state file's `worktrees` map is
written only at the end of section 7. See **2. Isolate the workspace (first run only)**
(`skills/myflow-do/SKILL-rationale.md`) for why the working notes come first.

On a fix run, resume the existing worktree. **Never create a second one.**

**This run's resolved worktree set — the set section 7's guard iterates — is the worktree just
created or resumed above, plus any additional worktree this change affects.** Per **Resolving a
change's worktrees** (`skills/myflow-contracts/pipeline.md`), how a command resolves the set beyond
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

```bash
myflow stage begin -command '/myflow-do' -stage do.sdd-tdd -harness <harness> -session-token mf-<literal-token> <name>
```

**At most one implementer subagent may be in flight against a given worktree at any moment.** The
parent waits for the previous implementer's commit sha for that worktree before dispatching the
next implementer into it; dispatches into different worktrees remain free to run concurrently. This
explicitly overrides `superpowers:subagent-driven-development`'s parallel dispatch guidance and
`superpowers:dispatching-parallel-agents` for same-worktree tasks, alongside the model-policy
override this section already carries against the same upstream skill. See `design.md` for the
concurrency failures — assertions left red at file seams, an agent idling on another's mid-edit
compile, corrupted test-result XML — this rule closes.

Invoke **superpowers:subagent-driven-development**, dispatching one implementer per bundle from

```bash
scripts/plan-dispatch-bundles.sh <changeRoot>/tasks.md
```

where `<changeRoot>` is section 1's `openspec status` output's `changeRoot` field, resolved inside
this worktree. Exit 0 proceeds to dispatch. A non-zero exit is a plan defect, not a review finding:
exit 1 names a task missing its `**Files:**` field, which `superpowers:writing-plans` repairs
before any dispatch happens; exit 2 stops the run. Bundling does not change the commit-per-task
model — an implementer handed a bundle still makes one commit per task, carrying that task's own
`Task-Id:` trailer, and a `Build: red` task still folds into the commit its `**Squash-with:**`
field names. Every implementer dispatch **must** carry all five of:

> **MYFLOW — COMMIT-PER-TASK:** Do **not** run `git push`, merge, or open a PR. As soon as
> RED-GREEN-REFACTOR completes for this task — before the parent dispatches review for it — commit
> your work with `git commit`, carrying a `Task-Id: <n>` trailer where `<n>` is this task's dotted
> id from its `tasks.md` heading. **The trailer identifies the task; the subject follows this
> project's own commit convention** — where that convention has a scope, `<n>` is the scope, so a
> Conventional Commits project writes `fix(<n>): <subject>` or `feat(<n>): <subject>` with the type
> describing the change, and a project with no stated convention may write `task(<n>): <subject>`.
> **Never weaken or bypass a project's commit validation to fit** — no `--no-verify`, and no edit to
> its commit-message validator; a rejected subject means writing one the project accepts. You **may**
> `git add`/`git commit` your own work, but never `openspec/` or `docs/superpowers/` —
> `/myflow-finish` stages and commits those.

**A `Build: red` task's commit folds into its green partner.** A task tagged `Build: red` also
carries `**Squash-with:** Task <N>`, naming the green partner whose commit it folds into. Once that
partner task has its own commit, fold the red task's commit into it using the same
fixup-and-autosquash mechanism used for fix rounds (see "Panel re-runs" below): `git commit
--fixup=<partner-task-sha>` followed by `git rebase --autosquash`, where `<partner-task-sha>` is the
green partner's own commit — the one named by the red task's `Squash-with:` field.

> **REQUIRED SUB-SKILL:** Use superpowers:test-driven-development — RED-GREEN-REFACTOR for this
> task. Delete any code written before its test.

> **REQUIRED SUB-SKILL:** When a test fails for a reason RED-GREEN-REFACTOR did not plan, invoke
> superpowers:systematic-debugging before writing a fix. An expected RED step needs no invocation.

> **REQUIRED READING:** [engineering-principles.md](engineering-principles.md) — your
> implementation must satisfy these principles; the panel's principles reviewer checks the diff
> against them.

> **PLAN PROVENANCE:** a fenced block tagged `unverified:` is a hypothesis, not code to transcribe.
> Establish the real API before writing against it, and report what you found. A block tagged
> `verified:<how>` was checked as stated; if it does not compile, report that — do not contort the
> code to match it.

**Dispatch every implementer on the model the state file records under `models.implementation`**,
defaulting to Opus (or the harness's strongest available model) when that field is absent or null.
Name it explicitly — never by omission, which silently inherits the parent's model. This
**overrides** subagent-driven-development's "least powerful model that can handle each role"
guidance; see **Model policy** in `skills/myflow-contracts/pipeline.md` for why. The panel's slots
default to Sonnet — the two rules differ on purpose.

**Guard the commit before dispatching review.** As soon as the implementer reports the task's
commit sha back, and **before** the parent dispatches that task for review, the parent runs

```bash
scripts/check-task-commit-fields.sh <worktree> <task-id> <task-sha> <task-base>
```

naming the worktree path, this task's dotted id from its `tasks.md` heading, the commit sha the
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
commit. Ledger line: `Task N: complete (commit <sha7>, review clean, model: <model>, review:
<combined|spec+quality>)` — **record the model and the per-task review shape on every dispatch**,
implementer and reviewer alike, so both the model and the shape choice are auditable after the
fact. Mark a checkbox `[x]` only after its task passes spec **and** quality review.

**The per-task review's shape depends on `reviewPanelRoster`.** Under `light` and `standard`, a
**single** combined reviewer per task covers spec compliance and code quality together, dispatched
on `models.reviewPanel`. Under `full`, the spec-compliance and code-quality reviewers both run. See
the roster table in section 5 for what each preset means.

On BLOCKED: pause and report. Never guess.

```bash
myflow stage end -command '/myflow-do' -stage do.sdd-tdd -outcome completed <name>
```

## 5. The review panel

```bash
myflow stage begin -command '/myflow-do' -stage do.review-panel -harness <harness> -session-token mf-<literal-token> <name>
```

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
scripts/check-panel-diff-size.sh <worktree> <merge-base>
```

Exit 0 proceeds. Exit 1 puts the choice to the operator, shaped per **Operator prompts**
(`skills/myflow-contracts/operator-prompts.md`) — that section is canonical for the mechanics, not
restated here:

> **The panel diff measured `<count>`, over the `<cap>` cap. How should this proceed?**
> - **Proceed with the panel anyway** *(default, recommended)*
> - **Stop and split the change** — ends the run at `IN_PROGRESS` with the implementation committed
>   on the branch

Exit 2 stops the run: a size the guard could not measure is not a size under the cap.

Record the measured count, the cap in force, and the operator's answer where one was given in
`.superpowers/sdd/final-review-panel.md` on **every** run, including exit-0 runs. The cap moves
nothing about the roster, the slots, the escalation ladder or the zero-open-findings bar.

Write `.superpowers/sdd/final-review.diff` from `git diff <merge-base>` (staged and unstaged), then
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
| 3 | **Code review (low)** | `light` | `models.reviewPanel` | general-purpose + the harness's `code-review` skill at effort `low`, against `final-review.diff` |
| 4 | **Security** | conditional | its own | `subagent_type: security-review`, same shape as Bugbot |
| 5 | **Adversarial** | conditional | `models.reviewPanel` | general-purpose + [adversarial-reviewer-prompt.md](adversarial-reviewer-prompt.md) |
| 6+ | **Principles lens B / lens C** | conditional | `models.reviewPanel` | same template, `[LENS]` = **Lens B — simplicity & state** or **Lens C — robustness & ops** |

Slots 1 and 4 are dispatched by `subagent_type` and carry their own agent definitions — pass them
**no** model override, whatever `models.reviewPanel` records, and record `unknown (agent-defined)`
for them in the ledger. Every other slot, slot 3 included, names its model explicitly.

**Every slot the panel dispatches must supply, per finding, a reproducer**: a runnable command that
demonstrates the defect, or the literal exemption form `none — <reason>`. Carry this requirement on
every slot's dispatch prompt. Slots dispatched by `subagent_type` (Bugbot, Security) receive it as
prompt text, the same way Bugbot already receives the mutation-testing brief below — no agent
definition is edited to carry it.

### Code review (low)

Slot 3, the `light` preset's third required slot, is a `general-purpose` subagent on
`models.reviewPanel`, told to invoke the harness's `code-review` skill at effort `low` against
`.superpowers/sdd/final-review.diff` in the worktree. Because the skill reports through a host
surface the parent does not read, tell the subagent to return its findings **in its report back**
rather than leaving them wherever the skill itself displays them, as ordinary `F<n>` rows and marker
lines like every other slot's. Because the dispatcher names the model explicitly, the ledger records
that real model for this slot and never `unknown (agent-defined)`, which is reserved for slots
dispatched by `subagent_type`.

**Where the harness offers no `code-review` skill**, the slot becomes a `general-purpose` reviewer
on `models.reviewPanel`, briefed to report high-confidence defects only, and the panel record names
the substitution. The slot is never dropped on that account, and the panel never falls back to two
required slots: an unavailable harness feature is not a way to weaken review.

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
`subagent_type` dispatch nor its `unknown (agent-defined)` ledger entry.

Slot 2 is the panel's only mandatory judgment check on *how* the code is built. It reads
`engineering-principles.md` — never a pasted copy — and owns the project's **hard invariants** from
its standards files: architecture and layer purity, new suppressions, weakened lint config.

**Resolve `[PRINCIPLES_PATH]` before dispatching any principles slot.** It is the **absolute** path
of `engineering-principles.md` in the directory you are reading this file from — under a global
install, `~/.claude/skills/myflow-do/engineering-principles.md`. Confirm the file exists before
spawning; if it does not, stop and report rather than dispatching a blind reviewer. See
**5. The review panel** (`skills/myflow-do/SKILL-rationale.md`) for why it must be absolute.

**Resolve `[STANDARDS_PATHS]` before dispatching slot 2**, from the `## standards` entries in the
project's `.myflow/project.md`. Entries are **not** paths to use as-is: each resolves through the
entry-form table and the containment rule in **Project configuration**
(`skills/myflow-contracts/project-configuration.md`), and an entry failing either is reported by
name and dropped. Resolve that contract file by **absolute** path too; if it is not readable,
**stop** — do not resolve entries without the containment rule. Pass an **empty** value when none
resolve; that correctly empties the Hard Invariants section rather than substituting another
project's standards. Record which standards files were passed, or that none resolved. See
**5. The review panel** (`skills/myflow-do/SKILL-rationale.md`) for why the containment rule is
non-negotiable here.

**No two principle reviewers may share a lens.**

**Every finding is recorded twice: as a row for the reader, and as a marker line for the guard.**
The panel record is `.superpowers/sdd/final-review-panel.md`. Write the table:

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Bugbot | Minor | `src/Foo.kt:42` | replaced the silent catch |

and, below it, the marker block — one line per row, plus the count:

```
findings-total: 1
finding-status: F1 fixed
```

Do not quote the marker format inside the record itself — a validly-formatted marker written as an
example inside prose or a fenced block reads the same as a real one.

**The reproducer each finding's slot supplied gets a marker block of its own**, separate from the
`finding-status:` block above — the two are kept apart because `scripts/check-unfinished-work.sh`
requires the `finding-status:` lines to occupy one unbroken span, and interleaving the two blocks
would break that. Write one `reproducers-total: <n>` count line and one `finding-reproducer: F<n>
<command | none — reason>` line per finding:

```
reproducers-total: 1
finding-reproducer: F1 scripts/test-check-panel-reproducers.sh
```

The same rule against quoting the format inside the record applies to this block as well.

`scripts/check-unfinished-work.sh` reads **only the marker block**. It never parses the table: not
its header, not its column order, not its cell boundaries, not where it starts or stops. See
**5. The review panel** (`skills/myflow-do/SKILL-rationale.md`) for why.

**The rules the guard does enforce, each of which reports outstanding when broken.** Run the guard
and read its own output for the reject reason on any given record — it already reports one in
human-readable form; the list below is not repeated here as a parse grammar:

- `<status>` is **exactly** `open`, `fixed` or `withdrawn`, compared byte for byte.
- `withdrawn` **carries its reason on the same line**.
- Each `F<n>` names **one** finding. A reused identifier is reported on each side separately.
- The marker lines sit on **consecutive lines**, one unbroken block.
- `findings-total: <n>` appears **exactly once** and equals the number of marker lines. A panel that
  raised nothing writes `findings-total: 0` and carries no markers.

**The table carries no status column, on purpose.** A finding's state is written once, on its
marker line. To read a finding's state, look up its `F<n>` in the marker block.

**A `withdrawn` marker's reason is checked for being there at all.** A withdrawal is the operator's
decision at the panel's handback, not a status a run may write for itself. Fix subagents write
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
`git rebase --autosquash`, write `.superpowers/sdd/fix-round-N.diff` from
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
| **Full** (escalation) | Every **required** slot; a **conditional** slot (Security, Adversarial, Lens B, Lens C) only when its own row in the optional-slot trigger table still fires against `fix-round-N.diff` | rewritten `final-review.diff` |

A conditional slot whose trigger did not fire is **not** re-run; its previous result stands, and the
record states `not re-run — subject unchanged`, distinct from a slot whose trigger never fired at
all and from a slot the operator declined. The **Targeted** row is unchanged. Not re-running a slot
never closes, softens or expires a finding it has already raised — the zero-open-findings bar still
governs every slot in the roster. See **Panel re-runs** (`skills/myflow-do/SKILL-rationale.md`) for
why this scoping reaches conditional slots alone.

**Escalate automatically** — do not ask, and say why in the record — when the fix touched a file
outside the set named in the findings; the fix altered a delta spec, a migration, or a guard's
behaviour; a targeted re-run surfaced a **new** Critical finding; three or more fix rounds have
already run; or the fix diff exceeds ~150 changed lines **and** adds a new file. Size alone carries
no risk signal — a mechanical rename is large and harmless, a one-line change to a guard's behaviour
is small and dangerous — so size never escalates on its own. **The signal it is paired with is `adds
a new file` and nothing else, because every other risk signal already escalates by itself**: a delta
spec, a migration, a guard's behaviour and a file outside the findings set each have their own clause
above, so repeating them here would add no case the ladder does not already reach. What the pairing
does reach is the one gap those clauses leave — a large body of brand-new, wholly unreviewed code in
a file the findings themselves named. See the delta spec for the longer argument. A
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
keys off it cites this paragraph rather than restating it.

**Worked example, on a compound note.** F20's Note reads: "swapping the two `comm` directions
mislabels every finding and still passes all 20 cases, because the assertions match on the
identifier substring rather than the message." Its three clauses are the root defect, a symptom,
and why testing missed it — only the first is the theme: "still passes all 20 cases" says the bug
went undetected, not what it is, and "the assertions match on the identifier substring" is itself a
separate later finding (F35) that folding in would wrongly merge into this one's identity. The
theme is the root cause alone: **swapped `comm` directions**.

**Before dispatching the fix subagent**, run

```bash
scripts/check-panel-reproducers.sh <worktree>
```

Exit 0 proceeds. Exit 1 covers two classes, handled differently. A **missing or malformed field** —
no `finding-reproducer:` line for some `F<n>`, a bad `reproducers-total:` count — is a
record-completeness defect, and the field is added before dispatch. A
**rejected reproducer shape** — a command carrying a shell metacharacter, an absolute path, a `..`
segment, a leading `-` on its path token, a URL, or a NUL byte — is a **refusal, not something to
rewrite until the guard accepts it**: the line is recorded **unverifiable** and put to the operator,
the same disposition `scripts/run-reproducer.sh` reaches for a refusal of its own below. It is
**never silently rewritten to satisfy the guard** — the line may be exactly the injection the guard
exists to catch, and rewriting it to pass defeats the check. Exit 2 stops the run: a record the
guard could not read is not a record in which every finding declared a reproducer.

**For each open finding whose record carries a runnable `finding-reproducer:` command** (not the
`none — <reason>` exemption form), run

```bash
scripts/run-reproducer.sh <worktree> "<the finding's finding-reproducer: text>"
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

**A bounce is a guard-class failure, not a review finding** — the accounting section 4 already
gives `check-task-commit-fields.sh`: no fix-round slot consumed, no round-count advance, and it
never closes, softens or expires the finding.

Give the surviving findings — every dispatched finding from the union above — to **one** fix
subagent as the combined list. Where a finding is confirmed as a real defect — as opposed to a
style or principles nit — the fix subagent invokes **superpowers:systematic-debugging** before
writing its fix. **Dispatch it on the model recorded under `models.panelFix`**, defaulting to Opus
(or the harness's strongest available model) when that field is absent or null — deliberately not
the panel's own default, for the reason stated under
**Model policy** in `skills/myflow-contracts/pipeline.md`. Record every pass in
`.superpowers/sdd/final-review-panel.md`: mode, which agents ran, why, the diff path they read,
and — when this pass bounced any finding — each bounced finding's defect identity (file:line plus
theme, as defined above) together with the reproducer output it carried back to its raising slot.
A pass that bounced nothing states that plainly rather than omitting the field.

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

Only that answer writes `withdrawn`, and only with the reason the operator gives. Nothing else in
the run may write it: the guard reads a `withdrawn` marker with nothing after the status as
outstanding, so a withdrawal with no stated reason does not clear the gate it appears to.

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
- **Every start command comes from `.myflow/project.md`'s `## run`**, with every path in it made
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
- Apps in scope come from `## apps` in the project's `.myflow/project.md`, or from auto-detection
  when that file or key is absent — see
  **Project configuration** (`skills/myflow-contracts/project-configuration.md`).
- **Where the project declares no runnable application**, resolve the `## lint` and `## test`
  commands instead, with every path in them made absolute, and do not give the handoff an
  application shape the project does not have.

```bash
myflow stage end -command '/myflow-do' -stage do.run-instructions -outcome completed <name>
```

## 7. Verify, stage, and hand off

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
<agents repo>/scripts/prepare-workspace.sh <worktree>
```

once per worktree in this run's resolved set — the same set section 2 resolved, non-empty by
construction — never a raw read of the state file's `worktrees` map, which reads as `{}` or absent
on every first run, before this run's own section-7 write, and would make this check pass having
examined nothing. Per **Resolving a change's worktrees** (`skills/myflow-contracts/pipeline.md`),
report an empty resolved set and do not proceed to validate a worktree the state file cannot name;
see section 2 above for why that stop is the anomalous case. Run this **before anything else below
this line.**
`<agents repo>` resolves the same way a bare `.mdc` `## standards` entry does, from a global or a
project-local install alike, stated once under **Where the agents repository is**
(`skills/myflow-contracts/project-configuration.md`). Not finding the script there is the
absent-script case below, and is reported rather than passed over.

`prepare-workspace.sh` runs `check-workspace-isolation.sh` against the worktree first, then — only
if that passes — derives and exports the variables the project's `## workspace isolation` section
declares, resolved against the workspace id section 2 computed, and prints one `KEY=value` line per
exported variable to stdout. Read its exit code for what happened and read those printed lines
rather than re-deriving any of them: exit 0 means the printed lines are what to carry forward into
`## lint` and `## test` below (an exit 0 with nothing printed means the project declares no
`## workspace isolation` section). A
non-zero exit is the dropped-row case (exit 1, relay the script's own lines verbatim and stop) or the
cannot-answer case (exit 2, stop the same way) — in either case, stop **before** `## lint` and
`## test`, without writing the state file. Correcting the row in the project's `.myflow/project.md`
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

**When the script cannot be located**, apply the same rules by hand — do not restate them here — from
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

Run the project's `## lint` and `## test` commands from `.myflow/project.md` (auto-detect if
absent) and show the output. **Nothing runs them later** — `/myflow-finish` has no verification
gate — so a non-zero exit blocks this handoff.

```bash
myflow stage end -command '/myflow-do' -stage do.lint-and-test \
  -outcome completed <name>
myflow stage begin -command '/myflow-do' -stage do.stage-diff -harness <harness> -session-token mf-<literal-token> <name>
```

Confirm every intended checkbox is `[x]`, and that `git log <merge-base>..HEAD` shows one commit
per completed task, per section 4's commit-per-task model, with every fix-round and red-task-partner
fixup already folded in via `git rebase --autosquash` — no stray `fixup!` commit should remain
unsquashed on the branch, unless a PR already exists (below), in which case the section's one
additional commit also sits on top.

In **every** affected worktree:

```bash
git -C <worktree> status
git -C <worktree> log <merge-base>..HEAD --oneline
```

> **`openspec/` and `docs/superpowers/` are never part of a task commit.** Section 4's
> COMMIT-PER-TASK clause already excludes both paths from every task and fixup commit; this step
> only confirms nothing slipped in, it does not stage anything itself. See **7. Verify, stage, and
> hand off** (`skills/myflow-do/SKILL-rationale.md`) for why.

**The one push exception.** Every task and fixup commit already sits on the branch, unpushed, per
sections 4 and 5. If the state file records a `prUrl`, a PR is already open, so this run also
commits `openspec/` and `docs/superpowers/` — the only paths a task or fixup commit never touches —
and pushes everything to the PR branch; otherwise this step commits and pushes nothing. On that path
only — and in this order — run
`scripts/preserve-session-records.sh <worktree> <name> <state-dir>`; then
`scripts/commit-split.sh <worktree> <name> "<impl-msg>" "chore(<name>): plan and session records"`;
then push the branch, which carries whatever that call committed along with every task and fixup
commit accumulated since the PR was opened. `<impl-msg>` covers the one case a task or fixup commit
does not: working-tree edits the operator made at the human gate without staging them — normally
none, in which case the script's own guard skips that commit, and only the planning commit lands.
When there is something to describe, derive `<impl-msg>` the same way a fixup commit's subject is
derived — `fix(<name>): <what changed since the last task commit>`. See **7. Verify, stage, and
hand off** (`skills/myflow-do/SKILL-rationale.md`) for why that ordering matters.

The preservation script overwrites in place; it never creates a second dated copy. A source that does
not exist is reported and skipped; **a non-zero exit means a copy was attempted and refused or
failed** — report it with the script's own stderr message and continue committing the fix. See
**Preserving the session records** (`skills/myflow-contracts/pipeline.md`).

**`commit-split.sh` is the same guarded chain run 1 uses** — the skipped-empty rule, the
stop-on-failure rule and the symlinked-planning-path case are all under **Git boundaries**
(`skills/myflow-contracts/pipeline.md`), which this call implements rather than restates. The empty
case is ordinary here — a fix round that touched neither `openspec/` nor the test guide has nothing
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

```
## Implementation committed — review and test

**Change:** <name>
**Panel:** clean — roster: <light | standard | full>, required: <that roster's required slots, per the table under **5. The review panel**>; optional: <selected, or "none — no triggers fired">
**Staged:** N/N tasks committed on branch | committed, plus one planning-artifacts commit, and pushed to the PR branch
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

The pre-edit description line is present only on a fix run that synced the description in section
**3**, and reproduces that text without summarising or reflowing it. A run that wrote nothing omits
the line rather than printing an empty one. See **Description sync**
(`skills/myflow-contracts/jira-integration.md`).

## Guardrails

- **Commit per task and per fixup, as section 4 and section 5 require** — never `openspec/` or
  `docs/superpowers/` in a task or fixup commit. **Never push, merge, or open a PR** — except the
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
- **Never** mark a checkbox before its task review passes.
- **No flags.** The only argument is the optional change name; report anything else.
