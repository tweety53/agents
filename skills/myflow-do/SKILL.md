---
name: myflow-do
description: Implement an OpenSpec change with Superpowers TDD and a multi-agent review panel, emit the manual test guide, and stage everything for one human gate. Re-run to apply a fix. Use for /myflow-do.
allowed-tools: Bash(openspec:*)
license: MIT
---

Implement an OpenSpec change, write its manual test guide, and stage both for the human gate at
`IN_PROGRESS`. **No commits** — unless a PR already exists, which is the one exception below.

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
default branch without explicit consent. Record each worktree's merge base — it goes into the state
file's `worktrees` map, which is the authoritative list of affected worktrees.

On a fix run, resume the existing worktree. **Never create a second one.**

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

Appending is recommended because most fixes are corrections within the change's existing scope, and
a sub-change per fix round buys a directory tree the operator has to read back. A nested sub-change
is never archived alone — it goes with its parent.

If the fix adds scope the linked Jira issue does not describe, sync the issue **description** per
**Description sync** in Jira integration (`skills/myflow-contracts/jira-integration.md`). Never
transition the issue here.

## 4. Execute (SDD + TDD)

Invoke **superpowers:subagent-driven-development**, treating each remaining checkbox (or a tightly
coupled group) as one task. Every implementer dispatch **must** carry all four of:

> **MYFLOW — NO COMMITS:** Do **not** run `git commit`, `git push`, merge, or open a PR. Leave all
> changes uncommitted in the worktree. You **may** `git add` your own work, but never `openspec/`,
> `docs/manual-test/` or `docs/superpowers/` — `/myflow-finish` stages and commits those. The parent
> records `TASK_BASE=$(skills/myflow-do/scripts/checkpoint)` before dispatch; your diff for review is
> `skills/myflow-do/scripts/uncommitted-review-package <plan-file> "$TASK_BASE"`.

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

Per-task review without commits: the parent runs
`skills/myflow-do/scripts/uncommitted-review-package <plan-file> "$TASK_BASE"` and gives the
reviewer the printed path, never a commit range. Ledger line: `Task N: complete (uncommitted, review
clean, model: <model>)` — **record the model on every dispatch**, implementer and reviewer alike, so
the policy is auditable after the fact. Mark a checkbox `[x]` only after its task passes spec **and**
quality review.

On BLOCKED: pause and report. Never guess.

## 5. The review panel

Write `.superpowers/sdd/final-review.diff` from `git diff <merge-base>` (staged and unstaged), then
dispatch **separate** review subagents — one per selected slot, in **every** affected worktree.
Never merge two slots into one prompt.

**Every slot the panel spawns directly runs on the model the state file records under
`models.reviewPanel`, defaulting to Sonnet** when that field is absent or null. There is no
parent-model inheritance and no economy tier — the panel's cost must not depend on which model the
operator happens to be running, and a recorded value is a deliberate decision for one change rather
than an inheritance path.

| # | Slot | Required? | Model | How to spawn |
|---|------|-----------|-------|--------------|
| 0 | **Primary** — plan alignment + code quality | **always** | `models.reviewPanel` | **superpowers:requesting-code-review** with `final-review.diff` + the plan/spec constraints |
| 1 | **Bugbot** — defect hunt | **always** | its own | `subagent_type: bugbot`, `Diff: uncommitted changes`, `Full Repository Path: <worktree>` |
| 2 | **Principles** | **always** | `models.reviewPanel` | general-purpose + [principles-reviewer-prompt.md](principles-reviewer-prompt.md), `[LENS]` = **Merged** |
| 3 | **Security** | conditional | its own | `subagent_type: security-review`, same shape as Bugbot |
| 4 | **Adversarial** | conditional | `models.reviewPanel` | general-purpose + [adversarial-reviewer-prompt.md](adversarial-reviewer-prompt.md) |
| 5+ | **Principles lens B / lens C** | conditional | `models.reviewPanel` | same template, `[LENS]` = **Lens B — simplicity & state** or **Lens C — robustness & ops** |

Slots 1 and 3 are dispatched by `subagent_type` and carry their own agent definitions — pass them
**no** model override, whatever `models.reviewPanel` records, and record `unknown (agent-defined)`
for them in the ledger. Every other slot names its model explicitly.

Slot 2 is the panel's only mandatory judgment check on *how* the code is built. It reads
`engineering-principles.md` — never a pasted copy — and owns the project's **hard invariants** from
its standards files: architecture and layer purity, new suppressions, weakened lint config.

**Resolve `[PRINCIPLES_PATH]` before dispatching any principles slot.** It is the **absolute** path
of `engineering-principles.md` in the directory you are reading this file from — under a global
install, `~/.claude/skills/myflow-do/engineering-principles.md`. The subagent's working directory is
the project worktree, which has no `skills/` tree, so a repo-relative path opens nothing and the
reviewer runs with no principle list. Confirm the file exists before spawning; if it does not,
stop and report rather than dispatching a blind reviewer.

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
its header, not its column order, not its cell boundaries, not where it starts or stops. So an
unescaped `|` inside a cell is just text, a reordered header changes nothing, and a row that lost a
boundary pipe still counts. That is the point of the split — the previous shape asked a hand-rolled
table parser to recover one fact from a grammar defined in prose, and it failed **open** six
distinct ways across three review passes before it was replaced.

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
- Each `F<n>` names **one** finding. A reused identifier is reported on each side separately, so two
  distinct findings labelled `F1` in both the table and the marker block cannot cancel out — that
  shape hid an open Critical, with the word `open` never appearing in a marker at all.
- The marker lines sit on **consecutive lines**, one unbroken block. This is what stops a marker
  quoted elsewhere — inside a fenced example, say — standing in for a marker that was never written,
  which is the one route that still under-counted when the redesign was attacked.
- `findings-total: <n>` appears **exactly once** and equals the number of marker lines. A record
  with no total line is outstanding however clean it reads: zero findings is not something to infer
  from silence. A panel that raised nothing says `findings-total: 0` and carries no markers.

Free prose is not a record of a finding's state: a state that cannot be counted cannot be enforced.

**The table carries no status column, on purpose.** A finding's state is written once, on its
marker line. A status cell beside the marker is a second surface that can silently disagree with the
line that governs: the machine's direction is protected — a marker reading `open` blocks whatever a
cell says — but nothing protects a reader who sees `fixed` in the table and believes it. State the
fact once. To read a finding's state, look up its `F<n>` in the marker block.

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
| 3 — Security | auth/authz, tokens, crypto, secrets or config, query construction, path or file handling, deserialization, CORS/HTTP edge, new dependencies | a config or dependency file changed, but only comments or a version bump |
| 4 — Adversarial | migrations, concurrency/scheduling, behavior changes to code with existing tests, any test modified or deleted, or **>~300** changed lines | **150–300** changed lines with no other trigger |
| 5 — Lens B (simplicity & state) | **>~200** changed lines, or **≥3** new classes/modules | — |
| 5 — Lens C (robustness & ops) | error handling, retries, schedulers, external integrations, config/env, migrations | — |

**Borderline → ask**, with **include** as the default. A reviewer too many costs tokens; one too
few costs a defect.

A documentation-, prompt-, or test-only diff with no trigger runs the three required slots alone.
That is a correct outcome, not a skipped review — say so explicitly.

### Panel re-runs

**Pass 1 always runs the full roster selected for this change.** Only re-runs after a fix are
scoped. Record `FIX_BASE=$(skills/myflow-do/scripts/checkpoint)` before each fix, then
`skills/myflow-do/scripts/uncommitted-review-package <plan-file> "$FIX_BASE"
.superpowers/sdd/fix-round-N.diff`.

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

## 6. Write the manual test guide

In the same run, write or refresh `docs/manual-test/<name>.md`. This is why reviewing and testing
are one gate: both surfaces are produced together and can never drift apart.

**The guide is a behaviour checklist at capability scope.** One tickable line per user-visible
behaviour, grouped by capability, scoped to the change's **blast radius** rather than to its plan
tasks. Phrase each line as the check to perform, in the register an operator would use — say
`check exercise update — the "key" field saves` — and never as a restatement of the requirement it
came from. A change that touched every part of exercise CRUD lists create, update, filter, sort and
delete; not one entry per plan task that produced them. Carry **no** per-step command transcripts,
**no** expected-output blocks, and **no** explanation of why a check exists.

A guide written per plan task grows with the implementation rather than with the behaviour, which is
what made earlier guides long without making them more thorough: several entries could exercise one
behaviour while another went unlisted.

Above the checklist, write a **short preamble stating how to run whatever is in scope**. Nothing
else belongs in it.

- **Every path is absolute**, resolved from `git worktree list` or the state file's `worktrees`
  keys. Never a relative sibling path (`../<other-app>`), and never a main-checkout path while a
  worktree holds the work.
- Apps in scope come from `## apps` in the project's `.myflow/project.md`, or from auto-detection
  when that file or key is absent — see
  **Project configuration** (`skills/myflow-contracts/project-configuration.md`).
- **Where the project declares no runnable application**, state each check as the command to run,
  one line each, tickable in the same way, and do not give the guide an application shape the
  project does not have. This repository is that case: it is the source of the myflow skills,
  commands and rules, and "running the apps" here means running its guard scripts, its assertion
  harnesses and a sandboxed installer pass. An application-shaped guide written for it would name
  an app, a port and a URL that do not exist.
- On a fix run, **refresh** the guide: preserve already-ticked boxes, and re-open only what the fix
  invalidated.
- There is no skip prompt and no `SKIPPED` marking. The guide is there to use or ignore; nothing
  records whether it was used.
- **Always write a `## Known incomplete` section.** Either the single word `None.` or a bullet per
  item the run knows is unfinished — a defect instrumented but not fixed, a case deliberately left
  for later, a box that cannot be ticked yet. Refresh it on every fix run.

  Finish runs in a different session and has no memory of this one, so anything not written here
  is invisible at the integration gate. `scripts/check-unfinished-work.sh` reads this section, and
  treats its **absence** as outstanding rather than as clear.

**The register above is prose; two shapes in this guide are machine-read, and neither changes.**
The checks stay an unordered list written with the `- [ ]` and `- [x]` markers, and the
`## Known incomplete` section stays exactly as described. `scripts/check-unfinished-work.sh` parses
both, so altering either would break a guard while appearing only to shorten a document. Shortening
what a line *says* is the whole of this change; the markers it is written with are not part of it.

## 7. Verify, stage, and hand off

Run the project's `## lint` and `## test` commands from `.myflow/project.md` (auto-detect if
absent) and show the output. **Nothing runs them later** — `/myflow-finish` has no verification
gate — so a non-zero exit blocks this handoff.

Confirm every intended checkbox is `[x]`, and that no commits were made:
`git log <merge-base>..HEAD` must be empty, unless a PR already exists (below).

In **every** affected worktree:

```bash
git -C <worktree> reset -q -- openspec/ docs/manual-test/ docs/superpowers/
git -C <worktree> add -A -- . ':(exclude)openspec/' ':(exclude)docs/manual-test/' ':(exclude)docs/superpowers/'
git -C <worktree> status
git -C <worktree> diff --cached --stat
```

**The `reset` is what enforces the rule; without it the `add` only assumes it** — the reason is
stated once under **Git boundaries** (`skills/myflow-contracts/pipeline.md`) and is not re-derived
here. What is specific to this command is *whose* staging it retracts (an implementer subagent's own
`git add`, or a worktree resumed with a dirty index) and why `git reset -- <paths>` is the tool:
it touches the index only, restores a tracked path to its `HEAD` entry instead of staging a deletion
the way `git rm --cached` would, and succeeds when a path is absent — which `docs/superpowers/` is
on every run that has not preserved records yet, and where `git restore --staged` would refuse the
whole command and unstage nothing.

> **Those three paths are never staged.** The exclusion is what keeps them out of the diff, rather
> than a filter applied when the diff is displayed: a filtered display leaves them in the staging
> area, where the IDE's staged-changes pane and `git status` show them again. The list is fixed —
> the pipeline chooses these paths itself, so no project can differ. `/myflow-finish` stages and
> commits them separately, so nothing is lost by leaving them unstaged here.

`git add -A` respects `.gitignore`; never force-add.

**The one commit exception.** If the state file records a `prUrl`, a PR is already open and a
staged-only fix would be invisible on it — commit and push to the PR branch instead of leaving the
work staged. Otherwise never commit. That path makes **two** commits, implementation first, so a fix
pushed to an open PR keeps its code commit free of planning artifacts. On that path only — and in
this order — run `scripts/preserve-session-records.sh <worktree> <name> <state-dir>`; commit what
the staging above left in the index, which is the implementation alone; then `git add -A` **again**,
which is what picks up the three excluded paths; then commit those as the second commit, and push
both. The staging above excluded `docs/superpowers/`, where the script writes, so without that
second `add` neither commit would carry the preserved records. That ordering is what makes a fix
round raised after a PR is open refresh the preserved records rather than leave them a round stale.

The script overwrites in place; it never creates a second dated copy. A source that does not exist is
reported and skipped; **a non-zero exit means a copy was attempted and refused or failed** — report
it with the script's own stderr message and continue committing the fix. The outcome table under
**Finish contract** (`skills/myflow-contracts/pipeline.md`) is canonical for all three outcomes.

**Both commits are guarded exactly as run 1's are** — the chain, the skipped-empty rule, the
stop-on-failure rule and the symlinked-planning-path case are all under
**Git boundaries** (`skills/myflow-contracts/pipeline.md`), which this path follows rather than
restates. The empty cases are ordinary here: a fix round that touched only `openspec/` and the test
guide has nothing for the first commit, and one that touched only code has nothing for the second.
Neither is an error, and neither is silent — say in the handoff which commit was skipped.

Write the state file: `IN_PROGRESS` from `STARTED`, otherwise **the state exactly as read**.
Populate `worktrees` with one absolute-path key per affected worktree and its merge base. Carry
`artifactUrl`, `jiraIssue`, `planningEffort`, `models` and `prUrl` forward verbatim, per the
carry-forward rule in **State file** (`skills/myflow-contracts/state-file.md`), which is canonical
for what a write must re-emit. The state file lives outside the repo — never `git add` it.

The one field where *verbatim* is not a byte copy is the planning effort: a file that recorded it
under the retired key is carried forward as the **mapped level under `planningEffort`**, per that
same carry-forward rule, which is canonical and is not restated here. What matters at this call site
is that reading only `planningEffort` and writing what it found would erase the recorded level.

The block below is **not** a second definition of the handoff. It is this command's rendering of the
`IN_PROGRESS`-after-`/myflow-do` template, which is defined once under
**The block each state renders** (`skills/myflow-contracts/pipeline.md`) and is canonical for the
labels, the field set and their order. What this block adds is the enumeration of the literal
alternatives `/myflow-do` writes. **Change the template first and bring this block with it** — a
field added here and not there is drift the moment `/myflow-status <name>` regenerates the same
state.

```
## Implementation staged — review and test

**Change:** <name>
**Panel:** clean — required: primary + Bugbot + Principles; optional: <selected, or "none — no triggers fired">
**Progress:** N/N tasks
**Git:** staged and uncommitted | committed as two commits and pushed to the PR branch
**Jira description (pre-edit):** <the text as it stood before the write, verbatim in a fenced block, inside <details> when long> | omitted — this run wrote no description

Worktree:   <absolute worktree path>
Test guide: <absolute path to docs/manual-test/<name>.md>

Review the diff, then run the apps against the guide:
  git -C <absolute worktree path> diff --cached          # when staged and uncommitted
  git -C <absolute worktree path> diff <merge base>..HEAD  # when committed and pushed
  open -na "IntelliJ IDEA" --args "<absolute worktree path>"

Re-run this command to fix anything you find.

Next:
/myflow-finish <name>
```

**Print one review command, the one that matches the `Git` line** — the two are shown together above
only because this block serves both of this command's cases. `--cached` on a committed branch exits
0 printing nothing, which reads as *there is nothing to review*; the merge base comes from this
worktree's entry in the state file's `worktrees` map. The template's third `Git` option — committed
and pushed with no PR — is one `/myflow-do` never emits and `/myflow-status` does; the pairing is
canonical under **The block each state renders** (`skills/myflow-contracts/pipeline.md`).

The pre-edit description line is present only on a fix run that synced the description in section
**3**, and reproduces that text without summarising or reflowing it — the transcript is then the
recovery path, since there is no local backup. A run that wrote nothing omits the line rather than
printing an empty one. See **Description sync** (`skills/myflow-contracts/jira-integration.md`).

## Guardrails

- **Never commit, push, merge, or open a PR** — except the `prUrl` exception above.
- **Never** run `finishing-a-development-branch`.
- **Never** create a second worktree for the same change.
- **Never** advance the state from `IN_PROGRESS`; write back what you read.
- **Always stage with the exclusion pathspec above** in every affected worktree before handing off —
  never a bare `git add -A`.
- **Never skip** a required panel slot, and never collapse two slots into one prompt.
- **Never dispatch an implementer without the provenance clause.**
- **Never** pass a model override to Bugbot or Security Review; **always** name the panel's model
  explicitly on every other slot.
- **Never** paste the principle list into a prompt — the reviewer reads the file.
- **Never** hand off with an open finding of any severity, or a stale clean result.
- **Never** mark a checkbox before its task review passes.
- **No flags.** The only argument is the optional change name; report anything else.
