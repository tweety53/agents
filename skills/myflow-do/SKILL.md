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

**Load `skills/myflow-contracts/pipeline.md` first** — it is canonical for the states, the
command→state transition table, git boundaries, and the handoff output shape.

**Then register this run's steps** with the harness's task-list mechanism, before any work begins,
and keep each entry's status current as the run proceeds, per
**Progress visibility** (`skills/myflow-contracts/pipeline.md`) — that section names which steps
this command registers and is the one to read. What is specific to this command, and so stated
here: an entry moves to in-progress when its implementer is dispatched, and to completed when that
task passes **both** its spec and quality review — the same moment its `tasks.md` checkbox is
allowed to be ticked, so the progress view and the file never disagree.

The reasoning behind this file lives in `skills/myflow-do/SKILL-rationale.md`; **a
`/myflow-*` run never loads it.**

## State gate

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
| **8** | **superpowers:verification-before-completion** | Evidence before claiming done |

**Never** invoke `finishing-a-development-branch` — integration is `/myflow-finish`'s job.

## 1. Load context and validate the plan

```bash
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

## 2. Isolate the workspace (first run only)

Invoke **superpowers:using-git-worktrees**. Branch `openspec/<name>`. Never implement on the
default branch without explicit consent. Record each worktree's merge base and absolute path as
soon as the worktree exists, in this run's own working notes — not in the state file, whose
`worktrees` map is written only at the end of section 7. That merge base and path are what section
7's workspace-isolation guard needs, and this run knows them from here on regardless of what the
on-disk state file currently says. At the section 7 write, that same map becomes the state file's
`worktrees` map, which is the authoritative recorded list of affected worktrees.

On a fix run, resume the existing worktree. **Never create a second one.**

**This run's resolved worktree set — the set section 7's guard iterates — is the worktree just
created or resumed above, plus any additional worktree this change affects.** Per **Resolving a
change's worktrees** (`skills/myflow-contracts/pipeline.md`), how a command resolves the set beyond
reading the state file's map is that command's own; this is `/myflow-do`'s. It is non-empty by
construction on every ordinary run, first or fix alike — the worktree is known the moment it is
created or resumed above, well before section 7 runs and well before the state file's `worktrees`
map is next written. Section 7's empty-set stop is for the genuinely anomalous case where this step
produced no worktree at all, never the ordinary shape of a first run.

**Then compute this worktree's workspace id from the change name.** The derivation is stated once
under **The workspace id** (`skills/myflow-contracts/workspace-isolation.md`), which is canonical
for it — do not restate it here, and do not re-derive it by hand. Compute it once per run, on a fix
run exactly as on the first. See **2. Isolate the workspace (first run only)**
(`skills/myflow-do/SKILL-rationale.md`) for why.

The main checkout has no id, and a project that declares no isolation at all is that same case
wherever it runs: every value resolves to the project's declared default, and neither is reported as
a misconfiguration. See **The empty id** (`skills/myflow-contracts/workspace-isolation.md`).

## 3. Documenting a fix, before implementing it

On a fix run, record what changed **before** writing code, so the proposal never goes stale. Ask
which of exactly two, with named options rather than open prose — this is a choice between courses
of action, which the planning-gate capability governs wherever a `/myflow-*` command asks for one,
not only in `/myflow-start`:

> **This fix has to be recorded before it is written — where should it go?**
> - **Append to `proposal.md` and `tasks.md`** *(default, recommended)* — the fix is recorded in the
>   change's own artifacts; nothing new is created, and the plan stays one file
> - **Create a linked nested `<name>-fix-N` sub-change** — its own proposal and plan, for a fix that
>   adds scope the parent change does not describe

If the fix adds scope the linked Jira issue does not describe, sync the issue **description** per
**Description sync** in Jira integration (`skills/myflow-contracts/jira-integration.md`). Never
transition the issue here.

## 4. Execute (SDD + TDD)

Invoke **superpowers:subagent-driven-development**, treating each remaining checkbox (or a tightly
coupled group) as one task. Every implementer dispatch **must** carry all four of:

> **MYFLOW — COMMIT-PER-TASK:** Do **not** run `git push`, merge, or open a PR. As soon as
> RED-GREEN-REFACTOR completes for this task — before the parent dispatches review for it — commit
> your work with `git commit`, subject line `task(<n>): <subject>` where `<n>` is this task's
> dotted id from its `tasks.md` heading and `<subject>` is a short description, and a `Task-Id: <n>`
> trailer. You **may** `git add`/`git commit` your own work, but never `openspec/` or
> `docs/superpowers/` — `/myflow-finish` stages and commits those.

**A `Build: red` task's commit folds into its green partner.** A task tagged `Build: red` also
carries `**Squash-with:** Task <N>`, naming the green partner whose commit it folds into. Once that
partner task has its own commit, fold the red task's commit into it using the same
fixup-and-autosquash mechanism used for fix rounds (see "Panel re-runs" below): `git commit
--fixup=<partner-task-sha>` followed by `git rebase --autosquash`, where `<partner-task-sha>` is the
green partner's own commit — the one named by the red task's `Squash-with:` field. This satisfies
`myflow-task-commits`'s requirement that a red task's commit folds into its green partner's commit.

> **REQUIRED SUB-SKILL:** Use superpowers:test-driven-development — RED-GREEN-REFACTOR for this
> task. Delete any code written before its test.

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
guidance; see **Model policy** in `skills/myflow-contracts/pipeline.md` for why, for how a recorded
choice and a session instruction relate, and for the operator-override rule. The panel's slots
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
script, and a mismatch found by hand sends the task back to the same implementer exactly as a
guard failure would.

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
on `models.reviewPanel`. Under `full`, the spec-compliance and code-quality reviewers both run,
which is today's behaviour. See the roster table in section 5 for what each preset means; this
section does not restate it.

On BLOCKED: pause and report. Never guess.

## 5. The review panel

**Read `reviewPanelRoster` from the state file before selecting slots**, defaulting to `light` when
the field is absent or null. It names the preset in force for this run, per
**State file** (`skills/myflow-contracts/state-file.md`), which is canonical for the field's shape,
its values and the absent-key rule.

| Preset | Required slots |
|--------|----------------|
| `light` *(default)* | Primary · Principles · Code review (low) |
| `standard` | Primary · Principles · Bugbot |
| `full` | Primary · Bugbot · Principles |

Every preset dispatches exactly three required slots, and no preset reduces that number. `full`
reproduces the roster in force before this table existed.

**No preset moves the handoff bar.** A preset selects how much reading the panel does and nothing
else: handoff still requires zero open findings at any severity under every preset, a minor finding
still blocks exactly as a critical one does, and the escalation ladder, fix-round rules, panel
record format, marker-line rules and operator handback are all unchanged. A preset able to lower the
handoff bar would not be sizing the reading — it would be a way to skip review.

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

### Code review (low)

Slot 3, the `light` preset's third required slot, is a `general-purpose` subagent on
`models.reviewPanel`, told to invoke the harness's `code-review` skill at effort `low` against
`.superpowers/sdd/final-review.diff` in the worktree. Because the skill reports through a host
surface the parent does not read, tell the subagent to return its findings **in its report back**
rather than leaving them wherever the skill itself displays them. Its findings take ordinary
`F<n>` rows and marker lines, exactly like every other slot's. Because the dispatcher names the
model explicitly, the ledger records that real model for this slot and never `unknown
(agent-defined)` — that value is reserved for slots dispatched by `subagent_type`.

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

The brief applies wherever Bugbot is dispatched, and nowhere else: no other slot acquires it, and
this brief adds no slot to any preset. See the roster table above for which presets dispatch Bugbot.
Bugbot is still dispatched by `subagent_type` with no model override, and its ledger entry still
reads `unknown (agent-defined)` — carrying the brief on the prompt changes neither.

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
entry-form table and the containment rule in
**Project configuration** (`skills/myflow-contracts/project-configuration.md`), and an entry
failing either is reported by name and dropped. Resolve that contract file by **absolute** path
too, for the same reason as
above; if it is not readable, **stop** — do not resolve entries without the containment rule, which
is the only thing between an attacker-editable list in a tracked file and an arbitrary file read
whose output lands in a committed review record. Pass an **empty** value when none resolve; that
correctly empties the Hard Invariants section rather than substituting another project's standards.
Record which standards files were passed, or that none resolved.

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

`scripts/check-unfinished-work.sh` reads **only the marker block**. It never parses the table: not
its header, not its column order, not its cell boundaries, not where it starts or stops. See
**5. The review panel** (`skills/myflow-do/SKILL-rationale.md`) for why.

**The rules the guard does enforce, each of which reports outstanding when broken:**

- A marker line **begins its line** and reads `finding-status: F<n> <status>`. Indented, inside a
  blockquote, or missing its `F<n>`, it is reported as a line naming `finding-status:` that is not
  one — never silently skipped. Do not quote the marker format inside the record itself.
- `<status>` is **exactly** `open`, `fixed` or `withdrawn`, compared byte for byte. `Open`,
  `WITHDRAWN`, `open (needs discussion)`, an empty value and a value carrying an invisible character
  are none of the three, and none of them reads as closed.
- `withdrawn` **carries its reason on the same line** — the reason is part of the state, not a note
  about it.
- The table's `F<n>` identifiers and the marker block's must name the **same** findings. A row with
  no marker and a marker with no row are each reported.
- Each `F<n>` names **one** finding. A reused identifier is reported on each side separately. See
  **5. The review panel** (`skills/myflow-do/SKILL-rationale.md`) for why.
- The marker lines sit on **consecutive lines**, one unbroken block. See **5. The review panel**
  (`skills/myflow-do/SKILL-rationale.md`) for why.
- `findings-total: <n>` appears **exactly once** and equals the number of marker lines. A record
  with no total line is outstanding however clean it reads: zero findings is not something to infer
  from silence. A panel that raised nothing says `findings-total: 0` and carries no markers.

**The table carries no status column, on purpose.** A finding's state is written once, on its
marker line. To read a finding's state, look up its `F<n>` in the marker block. See
**5. The review panel** (`skills/myflow-do/SKILL-rationale.md`) for why.

**A `withdrawn` marker's reason is checked for being there at all.** A withdrawal is the operator's
decision at the panel's handback, not a status a run may write for itself: a fix subagent rewriting
`open` to `withdrawn` would be closing its own gate, which is the one move this bar exists to
prevent. Fix subagents write `fixed` when they fixed it, and leave `open` when they did not —
including when they believe the finding is wrong, which belongs in the note and in the report back,
never in the status.

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

**Under `full`**, a fired trigger auto-includes its slot, which is today's behaviour, and the
table's borderline *ask* rows keep their current behaviour there.

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
fired — the two are different facts about the same diff, and a reader of the panel record must be
able to tell them apart.

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
created when the task was first implemented, not the sha from a previous fixup-and-autosquash
round — so that every fixup round for a given task targets the same original commit. Immediately
run `git rebase --autosquash` to fold the fixup into that commit, before the next review pass reads
the diff. This keeps the working tree at one commit per task throughout the panel's re-runs, never
a trailing separate fixup commit.

| Mode | Who re-runs | Diff they get |
|------|-------------|---------------|
| **Targeted** (default) | Slot 0 (always, as integration check) + every agent that raised a finding | `fix-round-N.diff` |
| **Full** (escalation) | Every slot in this run's roster | rewritten `final-review.diff` |

**Escalate automatically** — do not ask, and say why in the record — when the fix touched a file
outside the set named in the findings; the fix diff exceeds ~150 changed lines; the fix altered a
delta spec, a migration, or a public contract; a targeted re-run surfaced a **new** Critical
finding; or three or more fix rounds have already run.

Targeting is a cost optimization, never a coverage waiver: a targeted re-run is never fewer than
two agents, and handoff still requires **zero open findings at any severity** from every agent that
has run, with the final pass showing a non-stale clean result for every slot in the roster.

Union all **open** findings, dedupe by file:line + theme, and give **one** fix subagent the
combined list. **Dispatch it on the model recorded under `models.panelFix`**, defaulting to Opus
(or the harness's strongest available model) when that field is absent or null — deliberately not
the panel's own default, for the reason stated under
**Model policy** in `skills/myflow-contracts/pipeline.md`. Record every pass in
`.superpowers/sdd/final-review-panel.md`: mode, which agents ran, why, and the diff path they read.

A minor finding blocks the handoff exactly as a critical one does. The escalation ladder is what
makes that terminate: when fix rounds do not converge the run hands back to the operator, who
resolves the disagreement — including by marking a finding `withdrawn` with a reason. That handback is
the existing human gate, not a routine way to defer a finding.

**The handback is an actual prompt, not a claim that one happened.** When a finding survives its
last fix round, stop and put it to the operator, one finding at a time, with the finding's text, the
fixer's reason for disputing it, and named options:

> **`<location>` — <the finding, in one line>. The fix round did not resolve it.**
> - **Take another round on it** *(default, recommended)*
> - **Withdraw it — I'll give the reason** — the reason is recorded on the finding's marker line
> - **Stop the run and hand it back to me**

Only that answer writes `withdrawn`, and only with the reason the operator gives. Nothing else in
the run may write it: the guard reads a `withdrawn` marker with nothing after the status as
outstanding, so a withdrawal with no stated reason does not clear the gate it appears to.

## 6. Resolve the run instructions

In the same run, resolve the run instructions for the handoff's `Run it:` section. It writes no
file. See **6. Resolve the run instructions** (`skills/myflow-do/SKILL-rationale.md`) for why.

Resolve:

- **Every app root is absolute**, resolved from `git worktree list` or the state file's `worktrees`
  keys. Never a relative sibling path (`../<other-app>`), and never a main-checkout path while a
  worktree holds the work.
- **Every start command comes from `.myflow/project.md`'s `## run`**, with every path in it made
  absolute.
- **Every URL is the one this worktree resolved**, never the project's declared base. **Resolve
  each URL from this worktree's workspace id, the way section 2 computed it** — every derived value
  is a function of that id, so the run instructions are resolved from the id and the project's own
  declaration rather than from a variable some later step exports. The project's
  `## workspace isolation` rows name the variable each URL is carried by and what it becomes in a
  workspace, per **Project configuration** (`skills/myflow-contracts/project-configuration.md`);
  what a workspace id moves at all is listed under
  **What the id derives** (`skills/myflow-contracts/workspace-isolation.md`).

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
  application shape the project does not have. This repository is that case: it is the source of
  the myflow skills, commands and rules, and "running it" here means running its guard scripts, its
  assertion harnesses and a sandboxed installer pass. An application-shaped `Run it:` section
  written for it would name an app, a port and a URL that do not exist.

## 7. Verify, stage, and hand off

**First, validate the section — with the guard, not by eye.** Run

```bash
<agents repo>/scripts/check-workspace-isolation.sh <worktree>
```

once per worktree in this run's resolved set — per section 2, the worktree this run created or
resumed, plus any additional worktree this change affects — never a raw read of the state file's
`worktrees` map, which a `{}` or absent map would make this check pass having examined nothing (and
is exactly what the map reads as on every first run, before this run's own section-7 write). Per
**Resolving a change's worktrees** (`skills/myflow-contracts/pipeline.md`), a resolved set that
comes back empty stops rather than passes: report it and do not proceed to validate a worktree the
state file cannot name. Because section 2 already made the set non-empty by construction, that stop
fires here only in the genuinely anomalous case where section 2 produced no worktree at all — not
on the ordinary shape of a first run. Run this **before** resolving a single row and before
anything else below this line. `<agents repo>` is the same root a bare `.mdc` `## standards` entry
resolves against, and the two steps that find it — from a global install and from a project-local
one alike — are stated once under
**Where the agents repository is** (`skills/myflow-contracts/project-configuration.md`)
and are not repeated here. **Resolve it before you run anything, and check the script is actually
there**: not finding it is the absent-script case below, which is a different outcome from the
guard exiting 2, and is reported rather than passed over.

| Exit | What it means | What this command does |
|------|---------------|------------------------|
| 0 | every row is well formed, **or the project declares no section at all** | resolve and export, below |
| 1 | one or more rows are malformed; each is named on stdout with the rule it broke | this is the dropped-row case below — relay the guard's lines verbatim and stop |
| 2 | it could not answer — a `.myflow/project.md` that is not a regular file, cannot be read, or a scan of it that failed | stop, for the same reason: an unvalidated declaration is not a validated one |

**When the script cannot be located** — a harness whose repository does not carry it, or a skill
directory copied rather than linked, so the two steps above resolve to something that is not the
agents repository — apply the same rules by hand from **Project configuration**
(`skills/myflow-contracts/project-configuration.md`), which is canonical for them, and say in the
handoff that the validation was performed manually **and why the script was not run**. See
**7. Verify, stage, and hand off** (`skills/myflow-do/SKILL-rationale.md`) for why the two reasons
are treated as one case. It is never skipped for want of the script, and "the
declaration was validated" is never reported for a run in which nothing validated it.

**Then export the variables the project's `## workspace isolation` section declares**, resolved
against the workspace id section 2 computed. How each row resolves — the four `In a workspace`
forms, the `<id>`, `<id_underscored>` and `<value:VARIABLE>` tokens, the validation every row
passes, and the single pass that makes declaration order free — is stated once under
**Project configuration** (`skills/myflow-contracts/project-configuration.md`) and is not re-derived
here. What belongs to this command is when the resolution happens and what follows from it:

- **When.** Once, before the `## lint` command and before the `## test` command below — never
  between them, and never per task. Both then run in an environment that already carries every
  declared variable, and every process either one starts inherits the same values. **The test
  command itself is unchanged**: this step changes what the tests run against, never how they are
  invoked or which they are.
- **Every declared row, not a subset.** See **7. Verify, stage, and hand off**
  (`skills/myflow-do/SKILL-rationale.md`) for why.
- **The cache index is claimed here, not looked up.** A `cache index` row resolves by probing the
  cache and claiming a free index, so this step is the moment the claim happens, per
  **The cache index** (`skills/myflow-contracts/workspace-isolation.md`). Claim once and export that
  one value; probing again later would hand two processes of one run two different indices.
- **A dropped row stops this run.** A row failing validation is reported by name and dropped, and in
  an apply worktree the run then refuses rather than falling back to the shared default — see
  **Project configuration** (`skills/myflow-contracts/project-configuration.md`). The guard above is
  what names it: its exit 1 *is* this case, and its stdout already carries the row, the cell and the
  rule, so relay those lines rather than restating them. Report the row,
  the cell that failed and the shared value being declined; export neither that value nor an unset
  variable; and stop **before** `## lint` and `## test`, without writing the state file. Correcting
  the row in the project's `.myflow/project.md` and re-running this command is the whole remedy — a
  dropped row moves no state.
- **A project declaring no such section exports nothing** and behaves exactly as it does today. That
  is the ordinary case for a repository with no runnable application, and this repository is one; it
  is never reported as a misconfiguration.

**This step does not call the project's `create` command, and that is a decision rather than a
gap.** `create` is called by whatever starts the project's applications, per
**Project configuration** (`skills/myflow-contracts/project-configuration.md`), and this command
starts none of them — it exports, lints, tests, and hands off. The applications are started at the
review gate by the operator, through the project's own `## run` commands, which is where the
creation and its one-time notice belong. See **7. Verify, stage, and hand off**
(`skills/myflow-do/SKILL-rationale.md`) for why.

Run the project's `## lint` and `## test` commands from `.myflow/project.md` (auto-detect if
absent) and show the output. **Nothing runs them later** — `/myflow-finish` has no verification
gate — so a non-zero exit blocks this handoff.

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
section 4 and section 5 — that part is unconditional. If the state file records a `prUrl`, a PR is
already open, so this run also commits `openspec/` and `docs/superpowers/` — the only paths a task
or fixup commit never touches — and pushes everything to the PR branch; otherwise this step commits
and pushes nothing. That path makes **one more** commit on top of the task/fixup commits already on
the branch, keeping the code history free of planning artifacts. On that path only — and in this
order — run `scripts/preserve-session-records.sh <worktree> <name> <state-dir>`; then `git add -A`,
which picks up only `openspec/` and `docs/superpowers/` since everything else is already committed;
then commit those as the one additional commit; then push the branch, which carries that commit
along with every task and fixup commit accumulated since the PR was opened. That ordering is what
makes a fix round raised after a PR is open refresh the preserved records rather than leave them a
round stale.

The script overwrites in place; it never creates a second dated copy. A source that does not exist is
reported and skipped; **a non-zero exit means a copy was attempted and refused or failed** — report
it with the script's own stderr message and continue committing the fix. The outcome table under
**Preserving the session records** (`skills/myflow-contracts/pipeline.md`) is canonical for all
three outcomes.

**This additional commit is guarded exactly as run 1's are** — the chain, the skipped-empty rule,
the stop-on-failure rule and the symlinked-planning-path case are all under
**Git boundaries** (`skills/myflow-contracts/pipeline.md`), which this path follows rather than
restates. The empty case is ordinary here: a fix round that touched neither `openspec/` nor the test
guide has nothing to add — that is not an error, and it is not silent — say in the handoff that no
planning-artifacts commit was made.

Write the state file: `IN_PROGRESS` from `STARTED`, otherwise **the state exactly as read**.
Populate `worktrees` with one absolute-path key per affected worktree and its merge base. Carry
`artifactUrl`, `jiraIssue`, `planningEffort`, `models`, `prUrl` and `reviewPanelRoster` forward
verbatim, per the carry-forward rule in **State file** (`skills/myflow-contracts/state-file.md`),
which is canonical for what a write must re-emit. The state file lives outside the repo — never
`git add` it.

The one field where *verbatim* is not a byte copy is the planning effort: a file that recorded it
under the retired key is carried forward as the **mapped level under `planningEffort`**, per that
same carry-forward rule, which is canonical and is not restated here. What matters at this call site
is that reading only `planningEffort` and writing what it found would erase the recorded level.

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
and committed and pushed to an open PR — because both leave the branch fully committed, so the same
commit-range diff reads either one; there is no `--cached` case left under commit-per-task. The
merge base comes from this worktree's entry in the state file's `worktrees` map. The template's
third git state — committed and pushed with no PR — is one `/myflow-do` never emits and
`/myflow-status` does; the pairing is canonical under **The block each state renders**
(`skills/myflow-contracts/handoff-blocks.md`).

The pre-edit description line is present only on a fix run that synced the description in section
**3**, and reproduces that text without summarising or reflowing it — the transcript is then the
recovery path, since there is no local backup. A run that wrote nothing omits the line rather than
printing an empty one. See **Description sync** (`skills/myflow-contracts/jira-integration.md`).

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
