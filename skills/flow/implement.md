# Implement (SDD + TDD)

Loaded by `skills/flow/SKILL.md` once planning artifacts exist — entered from
`skills/flow/resume.md` on a resumed `STARTED` run whose
plan is already ready, or on a fix run at `IN_PROGRESS`.

| Step | Skill | When |
|------|-------|------|
| **2** | resume the workspace created at `flow.kickoff`, stated in **2. Isolate the workspace** below | Before the first code change, every run |
| **3** | **superpowers:writing-plans** | Validate the plan; repair `tasks.md` if it is not apply-ready |
| **4** | **superpowers:subagent-driven-development** | Execute the remaining tasks — only when the recorded decision's `execution` is `sdd`; an `inline` run (**Inline — the parent implements** below) never loads it |
| **5** | **superpowers:test-driven-development** | Every implementer dispatch, every task |
| **6** | the review panel | The final whole-branch panel |
| **7** | **superpowers:systematic-debugging** | An unexpected test failure during implementation, or a review-panel finding confirmed as a real defect |
| **8** | **superpowers:verification-before-completion** | Evidence before claiming done |

**Never** invoke `finishing-a-development-branch`. Integration is `skills/flow/integrate.md`'s job.

## The parent orchestrates directly

Sections **1**, **2** and **4** below, `skills/flow/review-panel.md` and
`skills/flow/verify-and-handoff.md` are **the parent's own work** — no resumed subagent runs
`flow.load-context` through `flow.write-in-progress` on its behalf; the running session does it
itself. Every "you" in those files addresses the parent. On a fix run,
section **3** below runs first: the parent resolves the worktree from the state file's `worktrees`
map, runs section 3's inline plan-append, Jira sync and re-decision, then continues into load-context/isolate/
sdd-tdd on the same session, with no dispatch in between. A fix run's stage order is therefore
document-fix → decide → load-context → isolate (resume) → sdd-tdd → …, so the appended plan is validated
after the fix's edit.

**The parent's shell discipline:** the parent's own Bash work above runs in the session's
interactive shell, and that shell is zsh — never assign `path`, the `PATH`-tied special
variable, and never lean on unquoted variable word-splitting: loop and list values go through
arrays or explicit lists, the re-derived footgun **Process lessons** (`skills/flow-contracts/lessons.md`) exists to end. One git command per `git -C` call, and never
a compound command led by `cd` — the harness hook denies those by design — a compound chain
staying legitimate only when every member is `git -C`-led and pasted whole, as **The guarded two-commit chain** (`skills/flow-contracts/git-boundaries-commit-chain.md`) is. A background
call inherits no cwd, so every path it names is absolute or `-C`-given — the standing rule the
FOREGROUND BUILDS paragraph of section 4 already states for long-running commands
(`skills/flow/implement.md`).

**Resolve `REVIEWERS`** per **Model resolution** (`skills/flow/SKILL.md`), and
run the guard-presence check, before this run's first dispatch — on a fix run, before section
**3**'s re-decision. Read the decision JSON
(`<abs-worktree>/.superpowers/sdd/decision.json`) for the decided `execution`, `implementer`,
`fixer` and `panel` — per the `## Decision` block of **Decide** (`skills/flow/brainstorm-planner.md`) — and for the recorded
`groups` field, which section **4** below
dispatches by.

**Never end a turn with a child in flight** — wait for every implementer, reviewer, slot or fix
subagent launched before reporting a stage boundary or asking the operator anything. **Turn
discipline**, below, states the one-foreground-wait-call shape this applies through.

### Dispatch sites — the parent's closed list

These seven rows are **every** Agent-tool dispatch the parent may make, across sections **1**,
**2** and **4** below, `skills/flow/review-panel.md`, `skills/flow/verify-and-handoff.md` and
`skills/flow-contracts/pipeline.md`'s **Pipeline defects found mid-run**:

| Site | Role | Key shape | Owning section |
|---|---|---|---|
| implementer, one per group | `implementer` | `task-<n>-implementer` | section **4** below |
| gated reviewer bundle, one per implementer group the review gate fires in on `big`, one per run on `micro`/`small`/`regular` | `reviewer` | `task-<n+n+n>-reviewer` | section **4** below, **The gated per-task reviewer** |
| panel bundle, at most two per round | `reviewer` | `panel-<round>-<slot\|slot+slot+slot>` | `skills/flow/review-panel.md`, **Bundled dispatch** |
| panel-fix, one per chunk of at most 10 findings | `panel-fix` | `panel-fix-<round>[-<chunk>]` (`-retry` once per chunk) | `skills/flow/review-panel.md`, the fix step |
| verifier, one per worktree | `verifier` | `visual-verify` (`-2`, `-retry`, `-fix-<k>`) | `skills/flow/visual-verify.md`, **The verifier dispatch** |
| tooling analyst, one per worktree whose fix run reports a miss an earlier round passed | `planner` | `tooling-analysis-<n>` | `skills/flow/visual-verify.md`, **A missed defect — the tooling analysis** |
| in-run pipeline fix, review and fixer, one each per loop step | `implementer` for the fix and the fixer, `reviewer` for the review | `pipeline-fix-<k>`, `pipeline-fix-<k>-review-<r>`, `pipeline-fix-<k>-fix-<r>` | `skills/flow-contracts/pipeline.md`, **Pipeline defects found mid-run** |

**Everything else in those six sections is the parent's own Bash and Read work, never
delegated** — every `check-*.sh`, `run-reproducer.sh`, `gather-dispatch-context.sh`,
`prepare-workspace.sh`, `## lint` and `## test`, every `flow record` and `flow stage` call,
worktree add and remove, every report read and every diff walk. Not to a "verify" reader, a
"re-verify" or "mutation re-verify" agent, a helper, a background task, or a subagent under any
other name.

**The self-check.** Before any Agent-tool call, the parent names which row above the call is. A
call that names no row is not made.

**Every dispatch's Agent-tool `description` names what it carries, in the operator's status-line
shape.** A dispatch carrying plan tasks — an implementer, a gated reviewer bundle — names them
by plan id over the plan's task count, then a few words in parentheses: `Task 3/22 (port guard)`
for one task, `Tasks 3+4+7/22 (review)` for a group, its ids in the order its key lists them. Every other row is its key and a few
words, with no parentheses: `panel-1-primary correctness review`. The `flow-task-list` mod
(`mods/flow-task-list/`) reads that prefix to number its row.

**Every dispatch is one-shot — a finished child is never resumed.** The parent never sends a
`SendMessage` to a child that has written its report. Whatever a
finished child's work still needs — a guard refusal, a pick conflict, a reviewer's `fix` — is the
parent's own inline fix, under the commit mechanics and records **Inline — the parent
implements** below already states; a re-review is a fresh dispatch.

Every row's own prompt carries the NO DELEGATION paragraph (section **4** below,
`skills/flow/review-panel.md`, `skills/flow/visual-verify.md`, `skills/flow-contracts/pipeline.md`) — a leaf never dispatches, so
nothing exists below these rows.
**The `flow-<effort>` family (`agents/flow-low.md`, `agents/flow-medium.md`, `agents/flow-high.md`
— three definitions, one per effort, each carrying `effort:` and no
`model:`) carries a `tools:` allowlist that omits
`Agent`** — the NO DELEGATION paragraph is backed by a capability the dispatched agent
structurally does not have, not by prompt text alone. This covers the panel bundle and
panel-fix rows whenever the decision's `panel` is an object (`skills/flow/review-panel.md`'s own
**The roster**), and the gated per-task reviewer row on `big`, where it dispatches on its group's
`model`/`effort` pair. On `micro`, `small` and `regular` that row dispatches `flow-low`
(**The gated per-task reviewer**, section **4** below), and on a `default` panel a panel-bundle
row does too, so the reviewer rows are structurally fork-free on every class and both panel shapes. The verifier
row dispatches `flow-low` unconditionally, regardless of the decision (`skills/flow/visual-verify.md`),
so it is structurally fork-free too.

**Inline — the parent implements** below takes this same table minus the implementer and panel-fix
rows — the parent's only permitted dispatches inline are the panel-bundle, gated per-task-reviewer,
verifier and tooling-analyst rows.

**The handshake.** Every dispatched role in this
pipeline — implementer, gated per-task reviewer, panel slot, panel-fix, verifier, tooling analyst, in-run pipeline fix — opens its first reply with the `Model:`
line the MODEL HANDSHAKE paragraph (section **4** below) demands, and every dispatch prompt in
this pipeline carries that paragraph verbatim. Compare the line against the model this dispatch
requested — the dispatch's recorded pair's model (the literal `opus` where none is recorded), or
the run's session override. A match proceeds. A **first** mismatch
closes the open dispatch row `-outcome fallback` and re-dispatches once, on the same requested
model and `subagent_type`, under `<key>-retry`:

```bash
flow record dispatch end -change <name> -key <key> -session-token mf-<literal-token> \
  -outcome fallback
flow record dispatch begin -change <name> -role <role> -model <the model originally requested> \
  -key <key>-retry -session-token mf-<literal-token>
```

A **second** mismatch closes the retry row `-outcome fallback` too and the parent asks the
operator directly through **AskUserQuestion**, naming the requested model and both models that
actually answered, options **Continue on `<the model the second handshake named>`** — proceed on
that running agent, no third dispatch — or **Stop the run**. Silence takes **Continue** — the only
option that moves the run forward, the substitution already on the record in the two fallback rows
(**Unanswered mid-run asks**, `skills/flow-contracts/pipeline.md`). **A mark or a record never blocks** —
proceed on the handshake's outcome regardless of whether any `flow` call reached the store.

On a single-model harness, **Harness mapping** (`skills/flow-contracts/model-policy.md`) states when a
reply's `Model:` line is not a mismatch.

**The return.** Once a dispatch's report file appears, read its verdict, print the change's own
handoff or continue to the next stage, and close the record under whichever key is open:

```bash
flow record dispatch end -change <name> -key <the key currently open> -session-token mf-<literal-token> \
  -outcome completed
```

**A dispatch whose agent dies is closed with `-outcome aborted`, reported, and not retried**: print
`/flow <name>` for the operator — a re-run resumes from whatever was left (checkbox state, the
state file's worktrees, findings in the store) through this file's own re-entry rules.

## Inline — the parent implements

Entered instead of dispatching an implementer per group when the recorded decision's `execution`
is `inline` (**Decide**, `skills/flow/brainstorm-planner.md`). The parent itself runs sections **1**, **2**
and **4** below, then `skills/flow/review-panel.md` and `skills/flow/verify-and-handoff.md`, with
these substitutions:

- **No implementer, no panel-fix dispatch.** The parent does each bundle's TDD work in the
  canonical worktree, commits per task with the same `Task-Id:` trailer and declared `**Commit:**`
  subject, and runs `close-task.sh` exactly as section **4** states.
  Waves are not parallel inline: bundles run in plan order, one at a time, never launched into a
  throwaway worktree.
- **Every dispatch-prompt paragraph that instructs an implementer or fixer in its work** — FLOW —
  COMMIT-PER-TASK, the TDD sub-skill, TARGETED TESTS, DRIFT CHECK, MUTATION PROOF, PLAN FIELDS,
  FOREGROUND BUILDS, PROJECT HAZARDS, and the rest section **4** and `skills/flow/review-panel.md` list —
  **binds the parent in the same words**, as if the parent had dispatched itself. **Five
  paragraphs govern only a dispatched child's own channel and never bind the parent**: NO
  DELEGATION (inline, the parent keeps the dispatch rows the closed list above leaves it), MODEL
  HANDSHAKE, TOOLS, CONTEXT BUNDLE and REPORT FILE. Without a bundle, REQUIRED READING's principles
  are read from `engineering-principles.md` beside this file, and PROJECT HAZARDS' warnings as the
  next bullet states. A question REPORT, DON'T DECIDE would carry in a report file goes to the
  operator in the parent's own output instead.
- **Project hazards are read by the parent itself.** Inline gathers no context bundle, so the
  `## hazards` section PROJECT HAZARDS points at never reaches it. Once per run, before the first
  bundle's first edit, the parent runs `flow hazards -C <canonical-worktree> -shape <shape>`,
  where `<shape>` is the value section **4** computes for its gathers, and holds every row it
  prints to PROJECT HAZARDS' terms. `[]` means no hazards. A failed call is reported in one line
  and never blocks, just as the gather's own hazards read never does.
- **Panel slots, the gated per-task reviewer and the visual-verify verifier dispatch exactly as in
  sdd mode** — a session reviewing its own diff is not a review. Panel fixes and gated per-task
  review fixes are applied by the parent instead of a panel-fix subagent or a resumed implementer;
  the parent still runs every reproducer and the fix-diff walk
  (`skills/flow/review-panel.md`) before recording a finding `fixed`. `flow.verify` runs inline
  for the parent exactly as under `sdd` execution.
- **Records:** one `dispatches` row per bundle, `-role implementer -model <parent model> -effort
  <parent effort> -agent-id inline`, and one per fix round, `-role panel-fix -model <parent
  model> -effort <parent effort> -agent-id inline` — so cost attribution and the stats views see
  inline work under the same roles a dispatched run would use. An inline fix's `dispatch begin`
  is recorded when the parent starts the fix, before its first edit, and its `dispatch end` once
  the fix commit lands — never the two together, which records the fix as taking no time. A gated per-task reviewer's own
  rows carry the dispatched agent's id, never `inline`; the gated fix round it causes records
  `-role implementer -model <parent model> -effort <parent effort> -agent-id inline` under the task's fix key.

## 1. Load context and validate the plan

```bash
flow stage begin -command '/flow' -stage flow.load-context -harness <harness> -session-token mf-<literal-token> <name>
spectre validate "<name>"
spectre list --json
check-plan-shape.sh "<changeRoot>/tasks.md"
```

Exit `0` from `validate` is the only exit that proceeds. Exit `1` names findings in this change's
own artifacts — most often a step checkbox left at column 0 — and each is repaired here, before any
code is touched. Exit `2` is a usage or IO error, and `no such change "<name>"` is the one worth
naming: nothing has been proposed under that name, so stop and suggest `/flow <name>`. **A change
already archived skips `spectre validate` and `spectre list --json`** — both answer for open
changes only, and `validate` names an archived one `no such change`; `check-plan-shape.sh` still
runs on the archived `tasks.md`.

Exit `0` from `check-plan-shape.sh` proceeds. Exit `1` names a shape defect in this plan's own
`**Files:**`, `**Tests:**` or other declared fields — repaired here, before any task is dispatched
and before any code is touched, exactly as an exit-1 from `spectre validate` already is. Exit `2`
stops the run.

**The change root is the change's directory** (**A change's directory**,
`skills/flow-contracts/pipeline.md`). Read:

- `<changeRoot>/proposal.md` — what and why
- `<changeRoot>/tasks.md` — the plan
- `<changeRoot>/design.md` — how, when the change carries one
- `<project>/spectre/specs/<capability>.md` for every capability the proposal names

**The plan is refreshed against the base before task 1 runs.** Before the
first task dispatches — inline or as an implementer — run, with this run's working-notes merge
base and every capability spec the proposal names as `<spec-path>` — the spec's
`<project>/spectre/specs/<capability>.md`, relative to `<project>`:

```bash
refresh-plan-base.sh <worktree> <merge-base> <changeRoot>/tasks.md <spec-path>…
```

Exit 0 prints `UNMOVED:` — record that one line — or `MOVED:` followed by the `CHANGED:` paths
the plan's `**Files:**` declares and each spec's text at the moved base (`SPEC-ABSENT:` when it is
gone there); exit 2 stops the run. The script's header (`<agents repo>/scripts/refresh-plan-base.sh`)
is canonical for its output. On `MOVED:`, **name every route or requirement the plan adds that the
moved base now already carries**. A collision is reconciled before task 1 runs, under **A pivot
reconciles the three artifacts together** (section 4 below) — never discovered mid-task. The step
names, it never rebases: the change branch is synced onto the base at integrate (**Sync the branch
onto the base**, `skills/flow/sync-onto-base.md`).

**Whether there is anything left to implement is read off the task checkboxes**, from
`spectre list --json`'s `{"changes":[{"id","done","total"}]}` for this change. A change already
archived is absent from that list: `done` is the count of column-0 `- [x]` task lines in its
archived `tasks.md`, and `total` that count plus the column-0 `- [ ]` ones.

- `total == 0` → no plan spectre can read: stop, resume at `skills/flow/resume.md`.
- `total > 0` and `done == total` → every task is already checked: proceed to
  `skills/flow/integrate.md`.

Confirm `tasks.md` meets writing-plans quality; if it does not, invoke **superpowers:writing-plans**
to repair it before touching code.

Extract the **Global constraints** verbatim from the capability specs the proposal names and
`design.md` for the reviewers.

```bash
flow stage end -command '/flow' -stage flow.load-context -outcome completed <name>
```

## 2. Isolate the workspace

```bash
flow stage begin -command '/flow' -stage flow.isolate-workspace -harness <harness> -session-token mf-<literal-token> <name>
```

Resume `<project>-worktrees/<name>`, created at `flow.kickoff`. Never implement on the
default branch without explicit consent.

**Persist each worktree's merge base and absolute path to the state file as soon as the worktree
exists — never defer this to the end of the run.** A run that stops, is interrupted, or is
resumed after a context compaction anywhere between here and `flow.write-in-progress`
(`skills/flow/verify-and-handoff.md`) must not leave the state record looking like a creating run
with no worktrees, when real worktrees, branches and commits already exist. Record it with one
call, which keeps every peer's entry and leaves `state` exactly as read (a creating run stays
`STARTED`; `flow.write-in-progress` is still the only step entitled to flip it to `IN_PROGRESS`):

```bash
flow state add-worktree <name> <abs-path> <merge-base-sha>
```

It exits 2 on a relative path or a malformed sha, and 1 when no `STARTED` record exists yet,
writing nothing either way. Do this once per worktree, immediately after
`spectre link` succeeds for it (or immediately after resuming it, on a fix or resumed run), not
batched at the end. The kickoff worktree's own entry never waits for this stage: it is recorded
where the worktree is created — `skills/flow/brainstorm.md` step 2 — so on a first creating run
this stage finds it already present, and the paragraph's rule above keeps governing each
additional worktree.

**Load `skills/flow-contracts/worktree-resolution.md`** — it derives this run's resolved worktree
set.

**This run's resolved worktree set — the set `skills/flow/verify-and-handoff.md` iterates — is the
worktree resumed above, plus any additional worktree this change affects.** Per
**Resolving a change's worktrees** (`skills/flow-contracts/worktree-resolution.md`), non-empty by
construction on every ordinary run.

**Every `## apps` entry is resolved in this stage too, before any implementation runs — never
improvised mid-run by a later stage in need of a commit destination.** Read
`<project>/.flow/project.md`'s `## apps` table and resolve each entry per the roots-are-main-checkouts rule of **Where the agents repository is**
(`skills/flow-contracts/project-configuration.md`): an entry whose
repository already holds a worktree for this change — the kickoff worktree's repository, a
peer's repository linked above, a second entry naming a repository already resolved by an
earlier entry, or a repository where an earlier run of this change created one — records that
worktree as its root and creates nothing.

**Load `skills/flow/cross-repo-worktrees.md` only when** this change links a peer, an `## apps`
entry names a repository holding no worktree for this change, `## visual verification` names a
`regression checkout`, or this run's resolved worktree set spans more than one repository — it carries each created worktree's `## worktree setup`, the `spectre link`
step, the rest of `## apps` resolution, the regression checkout's toolchain and the merge-order
record.

**Then run `flow workspace-id <name>` for this worktree's workspace id**, once per run, on a fix
run exactly as on the first.

```bash
flow stage end -command '/flow' -stage flow.isolate-workspace -outcome completed <name>
```

## 3. Documenting a fix, before implementing it

**Load `skills/flow/document-fix.md` only when** this run is a fix run — state `IN_PROGRESS` with
fix instructions, or a plain message at `IN_PROGRESS` (`skills/flow/SKILL.md`); it carries the
`flow.document-fix` and `flow.decide` marks, the re-plan budget, the planning pass, the Jira
description sync, plan growth and the fix-run planning commit, and a first run marks nothing there.

## 4. Execute (SDD + TDD)

```bash
flow stage begin -command '/flow' -stage flow.sdd-tdd -harness <harness> -session-token mf-<literal-token> <name>
```

**At most one mutator — any Edit/Write-capable dispatch: an implementer, a panel-fix, a fix-round
agent of any name — may be in flight against a given worktree at any moment** — a reviewer is not
one: it reads an immutable commit range, and any number of them may run beside the one mutator.
Dispatches into different worktrees remain free to run concurrently. This explicitly overrides
`superpowers:subagent-driven-development`'s parallel dispatch guidance and
`superpowers:dispatching-parallel-agents` for same-worktree tasks. The invariant is the working
tree, not the build tool, and it holds however file-disjoint two tasks look on paper. **A mutating dispatch's
report is complete only when its build's own success line is quoted and `git -C <worktree>
status --porcelain` is empty or every entry it prints is explained in the report** — "waiting for
the build" is never a finished report, and uncommitted WIP is a finding, never a state the next
dispatch inherits; a fix to something visual names the fresh capture taken after that build, on
the same footing as step 6's fingerprint (`skills/flow/visual-verify.md`). A mutator that
finds unrelated uncommitted changes mid-task says so and stops rather than fixing around them.

**The parent records each dispatch in two calls — one as it goes out, one as it comes back.**
Immediately before dispatching:

```bash
flow record dispatch begin -change <name> -task <n> -role implementer -model <m> \
  -key task-<n>-implementer -session-token mf-<literal-token>
```

and as soon as that dispatch reports back, before the next one goes out:

```bash
flow record dispatch end -change <name> -key task-<n>-implementer \
  -session-token mf-<literal-token> -commit <sha> -outcome completed
```

Both calls are required. `begin` is recorded immediately before the dispatch — or, when several
launches go out together in one message, in the next Bash call, one call for all (step 3 of **The
next implementer overlaps the guard**, below) — carrying no
`-agent-id`: the daemon captures the agent's identifier. `-agent-id` is accepted on both
calls as recorded intent the daemon never overwrites, for a caller that knows the id — the
literal `inline` on a row the parent's own inline work records included. `-key` is this dispatch's own literal label,
unique within the run's session token — `task-<n>-implementer`, reused identically in both calls.
`-role` is one of `implementer`, `reviewer`, `panel-fix`, `verifier` (**Verify**,
`skills/flow/verify-and-handoff.md`) or `planner` (the tooling analyst, `skills/flow/visual-verify.md`); `-task` is the task's
flat integer id, omitted for a dispatch against no single task. `-session-token` takes a literal,
never a shell substitution. The start and end instants are the daemon's own — never
caller inputs.

**`-model` is the model this dispatch was actually given — the group's own `model` from the
decision's `groups` entry** (chosen per **Model and effort**, `skills/flow/brainstorm-planner.md`,
or the run's session-instruction override when one was given for the implementer role), `-effort` its `effort`, and the dispatch's `subagent_type` is
`flow-<effort>` with the group's `model` passed as the Agent tool's own `model` parameter — the
definition carries the effort, the dispatch carries the model. Name it explicitly — never by
omission. On harness `zcode` the pair given and recorded is `glm-5.3-flash` / `high` instead (**Harness mapping**, `skills/flow-contracts/model-policy.md`). A slot whose model the dispatcher cannot
read records the literal `unknown (agent-defined)` and never a guess.

**A record write never blocks.** An unreachable store journals the intent, prints one warning line,
and exits 0 — never branch on this command's exit code as a signal about the record.

**An incident one dispatch caused rides the next dispatch to the same role.** When a dispatched
child does what its brief should have prevented — a fixup folded against a sha that was not an
ancestor, a checkout that destroyed uncommitted planning artifacts, any damage a marker or
plan-tree verify later catches behind a clean report — the parent, before anything else
dispatches, records it in the store every later context bundle renders:

```bash
flow record incident -guard <the check that would have caught it> -symptom <what happened, verbatim> \
  -recovery <the exact commands that would have caught it> -minutes-lost <n> -change <name>
```

and opens the next dispatch to that same role with the incident carried verbatim: what happened,
those exact commands, and an explicit **stop and report** instruction — a repeat is handed back,
never self-recovered. A generic warning ("be careful with git") is not a carry. The child's own claim that nothing went wrong
never closes an incident.

Run `plan-dispatch-bundles.sh <changeRoot>/tasks.md` for this plan's bundles:

```bash
plan-dispatch-bundles.sh <changeRoot>/tasks.md
```

Exit 0 proceeds. A non-zero exit is a plan defect: exit 1 names a task missing its `**Files:**`
field, repaired by `superpowers:writing-plans` before any dispatch happens; exit 2 stops the run.

**Load `skills/flow/sdd-dispatch.md` only when** the recorded decision's `execution` is `sdd` — it
carries the per-group implementer dispatch, the **Waves**, the per-group context-bundle gather and
its failure handling, and the sixth-argument scoping; an `inline` run (**Inline — the parent
implements**, above) never loads it.

`<shape>` is this change's shape, computed once per run from the resolved worktree set — more than
one repository → `cross-repo`, otherwise `single-repo` — and passed on every gather this run
makes, this file's and `skills/flow/review-panel.md`'s alike. The context bundle's `## hazards`
section is filtered by it; a gather made without it carries only always-on hazards.

In every gather (`skills/flow/sdd-dispatch.md`, `skills/flow/review-panel.md`), `<changeRoot>` is
the change's directory (**A change's directory**, `skills/flow-contracts/pipeline.md`) resolved
inside this worktree — or inside `<canonical-worktree>` when this worktree holds none, an
`## apps` repository given no `spectre link` (`skills/flow/cross-repo-worktrees.md`) — and
`<principles-path>` is the **absolute** path of `engineering-principles.md` **beside this file** —
`skills/flow/`, always.

`<canonical-worktree>` — the seventh argument, passed on every call — is
the member of this run's resolved worktree set whose own change directory holds `tasks.md`, the
same argument
`check-unfinished-work.sh` takes at the integrate gate; on a single-repo change that member is
this worktree and the argument is inert, while on a satellite worktree's group it carries the
canonical plan under labeled sections while keeping this worktree's own project commands,
incidents and HEAD (`<agents repo>/scripts/gather-dispatch-context.sh`'s header is canonical for
the resolution).

**A guard you could not run is hand-substituted only on the record.** When a guard this
file calls — the gather above, `check-task-commit-fields.sh` at task close — exits non-zero,
cannot resolve this change's topology, or is absent, and you go on by composing its facts,
checking its contract, or running its step by hand, record the substitution at the moment you
make it, never in the handoff prose alone:

```bash
flow record substitution -change <name> -guard <guard-name> -shape <shape> -substitution "<what you ran instead>"
```

`<shape>` is the same computed value the gather above took (`cross-repo` or `single-repo`), and
`-substitution` carries the command or manual step actually used, verbatim.

Every implementer dispatch **must** carry:

> **FLOW — COMMIT-PER-TASK:** Do **not** run `git push`, merge, or open a PR. As soon as
> RED-GREEN-REFACTOR completes for this task — before the guard runs on it — commit
> your work with `git commit`, carrying a `Task-Id: <n>` trailer. The trailer identifies the task;
> the subject is this task's declared `**Commit:**` field, reproduced exactly. **The commit is
> pathspec-scoped** — `-- <this task's files>` — so it carries only the paths this task names,
> whatever else the index holds; a plain commit sweeps a pre-staged foreign tree in with the
> task's work. **Never weaken or
> bypass a project's commit validation to fit** — no `--no-verify`. Stage for that commit only
> through the guarded sequence below, in this order — the clearing pass runs first, before any
> `git add`, because a `:(exclude)` governs what an `add` adds and cannot retract what an earlier
> step already staged:
>
> ```bash
> git reset -q -- spectre/changes/ openspec/changes/ docs/superpowers/ \
>   && git add -- <this task's files> ':(exclude)spectre/changes/' ':(exclude)openspec/changes/' \
>   ':(exclude)docs/superpowers/' \
>   && git commit -m "<the task's declared subject>" -m "Task-Id: <n>" -- <this task's files>
> ```
>
> Paths are relative to the worktree root. Both spec-tree leaf spellings are named because the
> leaf is the project's — `spectre` or `openspec` — and a path the project does not use is a
> reset that clears nothing and an exclude that matches nothing. **Never a bare
> `git add -A`, `git add .`, or `git commit -a`**:
> any of them sweeps already-staged planning artifacts into the task commit, which
> `check-task-commit-planning-paths.sh` fails at the boundary. Never
> `<project>/spectre/changes/` either — the excludes do not licence a second, deliberate add of it.
> **A capability spec under `<project>/spectre/specs/` is your work, not theirs**: when this task's
> `**Files:**` names one, edit it and commit it here, in this task's own commit. **A withdrawn
> task is backed out, not erased**: when work an earlier commit landed is withdrawn, the
> backing-out commit's body carries the rejection reason — what was wrong, and what would have to
> change for it to come back (**Revert and back-out commits**,
> `skills/flow-contracts/git-boundaries.md`).

**A `Build: red` task is dispatched with its `Squash-with:` partner, in one bundle.** The
implementer runs the red task first — writes its tests, runs them, and reports the failing
output — then the green partner, and makes **one** commit for the pair: the partner's declared
`**Commit:**` subject and `Task-Id: <partner>` trailer. The red task never has a commit of its
own; `check-task-commit-fields.sh` resolves the pair from either id against that commit.

> **REQUIRED SUB-SKILL:** Use superpowers:test-driven-development — RED-GREEN-REFACTOR for this
> task. Delete any code written before its test.

> **REQUIRED SUB-SKILL:** When a test fails for a reason RED-GREEN-REFACTOR did not plan, invoke
> superpowers:systematic-debugging before writing a fix. An expected RED step needs no invocation.

> **PROVE THE GUARD BITES:** When this task's tests assert on configuration or file content — a
> guard script, an embedded config, a fixture file — a passing run alone is not evidence: break
> the property the assertion protects, run the guard or the test against the broken state and
> capture its failure, then restore and capture the pass. Report both runs. "The pattern is now
> stricter" is intent, not evidence — the failing run against the broken state is the evidence.

> **PIN BEFORE REFACTOR:** A task that refactors code whose observable behaviour includes
> produced output — a guard's warnings, a CLI's printed lines, an emitted file — pins that
> output before any code moves: its RED step first writes harness cases asserting the current
> outputs exactly as produced today, including the ones nothing asserted yet, and only then
> refactors. The refactor lands with two proofs, both recorded before the change closes: the
> pinned suite passing, and one end-to-end old-vs-new comparison that runs the whole producer
> against the same fixture on both sides of the refactor and diffs the outputs byte-for-byte.
> The suite proves only what it asserts; the old-vs-new diff is the proof for every output it
> never named.

> **REQUIRED READING:** the engineering principles section of the context bundle below — your
> implementation must satisfy these principles; the panel's principles reviewer checks the diff
> against them.

> **CONTEXT BUNDLE:** `<abs-worktree>/.superpowers/sdd/dispatch-context-group-<g>.md` carries this
> change's proposal, design, engineering principles, and — under `## tasks.md` — the plan header
> plus your group's own task(s) only, gathered for you. You **must** still read the
> actual diff and the actual code — the bundle is shared *input*, never a substitute for the source.
> It also carries this project's `## lint`/`## test`/`## run` commands, already resolved — you do
> not need to open `<project>/.flow/project.md` yourself for them.

> **PROJECT HAZARDS:** the bundle's `## hazards` section carries this project's recorded
> warnings, filtered to this change's shape — each one names a way this project specifically
> bites, recorded after it cost time. They are binding: read them before your first edit and
> never argue one away without measuring.

> **PLAN PROVENANCE:** a fenced block tagged `unverified:` is a hypothesis, not code to transcribe.
> Establish the real API before writing against it, and report what you found. An assumption tagged
> `unverified:` in your task is a guess to confirm before you build on it. When what you
> measure contradicts the plan, stop and report the measurement: see **When a measurement
> contradicts the plan** (`skills/flow-contracts/plan-amendment.md`). Report a correction your
> measurement produced — a tag to retag, a number to replace, a `**Files:**` entry to add — as the
> exact field and its corrected value, so it can be transcribed into the record in place
> (**The record carries its own corrections**, below).

Every implementer dispatch **must** also carry:

> **MODEL HANDSHAKE:** the first line of your first reply is `Model: <the model named in your own
> system prompt>` and nothing else on that line. Answer it before any tool call.

The parent compares that line against the group's recorded model (or the run's session override)
and applies **The handshake** stated above, unchanged.

> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or poll it to
> completion, before you stop.

> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,
> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch in your first turn,
> before anything else. Never ToolSearch for a tool already listed, and never a wildcard query: a
> schema loaded later changes your tool list and re-prices your whole context at full input rate.

> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and never spawn a
> subagent, background agent or helper of any kind — you are the leaf of this run, and any child
> you start is unrecorded and outside the parent's closed list (**Dispatch sites — the
> parent's closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and
> fixing are your own Read, Bash and Edit calls.

> **TARGETED TESTS:** Run only the tests this task's `**Tests:**` field names, through the build
> tool's own selector — `--tests '<class>'` for Gradle, `-run '<name>'` for `go test`, `-t
> '<name>'` for vitest — once for RED, once for GREEN, and again only after a source edit. Never
> run the module or repository suite mid-task: the full `## test` list runs once per worktree at
> the last bundle, and again in `flow.verify`. Pipe a test run's output through `tail` so a green
> run costs lines of context, not a build log.

> **DRIFT CHECK:** Before you commit, read your own diff once for what it leaves behind. For
> every behaviour, name, path, flag or exit code the diff changes, grep for the comments, file
> headers, usage lines, docs and cross-file citations that describe it, and bring each one in
> line in this same commit. Every refusal or exit path the diff adds gets a test case of its
> own.

> **PROOF RUNS:** When a check has to be run repeatedly to prove a flake reproduced or fixed, size
> the run count to the flake rate you measured, 10–15 runs by default, never a fixed count per
> step. Your report states the measured rate and the run count. Sizing the runs never replaces
> the proof `~/.claude/rules/fix-determinism-at-the-source.md` requires: a determinism fix still
> shows the check failing with the mechanism removed.

> **OUTPUT BUDGET:** Every tool result stays in your context, and every later turn re-reads your
> whole context — a large output is paid for again on every turn after it. Read a file over 200
> lines by line range — `grep -n` for the symbol, then `sed -n '<a>,<b>p'` or Read with
> `offset`/`limit` — and never re-read a file already in your context unless you have edited it
> since. Cap every search (`| head -40`) and every build, lint or install run (`| tail -30`), and
> reproduce a failing block from its log rather than printing the whole log. Never print a
> generated file — a lockfile, a snapshot, a bundle, a build artifact.

> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one test you write MUST exercise the real
> thing. A test backed by a fake or a hand-built value passes while the real integration is broken:
> the shape you construct by hand is not the shape the real producer emits. Build the value the way
> production builds it, or assert against the real boundary.

> **REPORT FILE:** write your report to
> `<abs-worktree>/.superpowers/sdd/implementer-report-<k>.md` as your **last** act — after your
> commit and your final test run — carrying each commit's sha, the failing RED output you saw,
> and anything the plan's `unverified:` tags asked you to establish. The dispatcher waits on that
> file's presence.

Every implementer dispatch also carries:

> **REPORT, DON'T DECIDE:** A question only the operator can settle — a spec point no recorded
> decision covers, a recorded decision the plan appears to contradict, a scope the plan does not
> name — is reported, never decided in code. Where the plan can proceed, implement it as written
> and carry the question in your REPORT FILE; where it cannot proceed without the answer, report
> BLOCKED. A decision you make silently is a decision nobody recorded.

**The parent runs the full suite — never an implementer.** Once the plan-last group — the group
holding the last `bundle <k>` line `plan-dispatch-bundles.sh` printed — has passed the guard, by
direct commit or by its wave's final pick, **the parent itself** runs the resolved `## test` list
once on the canonical worktree, in the foreground, in the order the context bundle carries it, its
output through `tail`. A failure in a file a plan-last group task's `**Files:**` field names **the parent fixes
itself**, as inline fixes are applied: it edits, commits on the route the branch's push state
dictates (**Panel re-runs**, `skills/flow/review-panel-fix-round.md`), records the round as one
`dispatches` pair `-role panel-fix -model <parent model> -effort <parent effort> -agent-id inline`
under `full-suite-fix-<n>`, and re-runs the list. Any other failure is the last boundary's
`## Question`, below.

**The next implementer overlaps the guard.** The unit is the group — the decision's `groups`
entry, one or more bundles `plan-dispatch-bundles.sh` emits. At each boundary, in this order:

1. **Group N+1's implementer commits** and writes its report; the wait above ends.
2. **One Bash call: `close-task.sh` and group N+2's gather.** Both are the parent's own Bash calls.
   `close-task.sh` runs the implementer's `record dispatch end`, the guard on every commit
   whose sha is new and `check-task-commit-planning-paths.sh <canonical-worktree> <merge-base>`
   over every commit since the merge base recorded in this run's working notes; only when both
   pass, the **review gate** on every
   task, `flow tasks tick` for every task the gate left `QUIET`, and
   `git -C <worktree> push origin spectre/<name>` per **Branch backup**
   (`skills/flow-contracts/git-boundaries.md`): the parent pushes, never the implementer.

   ```bash
   close-task.sh -end-key <key> -session-token mf-<literal-token> -end-commit <sha> \
     [-undeclared <task-id>=<path>[,<path>…]] <canonical-worktree> <name> <merge-base> <task-id>:<task-sha> …
   ```

   `<canonical-worktree>` is the worktree created or resumed in **2. Isolate the workspace**
   above, `<name>` this run's resolved name. **On a change whose resolved worktree set spans more
   than one repository, a task's sha is the per-repository commit map instead** —
   `<task-id>:<worktree>=<sha>[,<worktree>=<sha>…]`, one pair per repository carrying any of the
   task's commits; the guard's own header (`<agents repo>/scripts/check-task-commit-fields.sh`,
   THE COMMIT MAP) is canonical for the form. A task that landed in one repository passes that repository's
   pair alone, which is the single-sha form; skipping the guard on a cross-repo change and
   checking fields by hand is the substitution shape **A guard you could not run is
   hand-substituted only on the record** exists to prevent, not a licence — the guard runs.
   `-undeclared` carries the paths a fields refusal named for that task (**The record carries
   its own corrections**, below). The script's header (`<agents repo>/scripts/close-task.sh`) is
   canonical for its arguments and its order.

   The guard reads git objects and `tasks.md` only, so it is safe while the tree changes, and
   never stashes, reverts or resets. Every verdict is printed and read before anything
   launches: **exit 1** — a fields refusal, or a task commit whose diff touches
   `<project>/spectre/changes/` (leaf resolved per project) or `<project>/docs/superpowers/` —
   is the parent's own fix, with nothing ticked or pushed: it re-commits that task itself,
   without the swept paths, and re-runs `close-task.sh` without the `-end-*` flags before
   anything below; **exit 2 — a guard's not-a-verdict close (an unreadable plan, a task it
   cannot resolve, a commit range git cannot resolve, a usage error), or a failed tick or push —
   stops the run**: re-committing cannot repair an inability, so it is never folded into
   the re-commit loop. Exit 0 prints every task's `FIRE:` or `QUIET:` gate line; a fired
   gate defers the tick to the task's reviewer, below.

   **A guard call that times out is inspected before it is retried.** Run
   `git status --porcelain=v2 --branch` and `git stash list` in that worktree first. A
   reverting, rebasing or merging state on the `# branch` lines, a change the run did not
   make, or a stash entry the parent did not push means the tree is not the one the run
   left — end the turn with `## Question` carrying both outputs verbatim; never re-run the
   guard on top of it.
3. **One message launches group N+2's implementer and, when any gate fired in group N+1, that
   group's one reviewer bundle (below). The next Bash call records every launch's `begin`**.

**When the script cannot be located**, apply its task-commit field rules by hand: check the
commit's `Files:` against `git diff --name-only <task-sha>^..<task-sha>`, its `Tests:` against the
commit's diff, and its `Commit:` against the commit's actual subject line.

**The review gate.** After the guard passes a task's commit, `close-task.sh` judges that commit's
gate through `check-review-gate.sh`, with every path the guard's refusal named for that task —
the paths its `-undeclared` flag carries:

```bash
check-review-gate.sh <worktree> <task-id> <task-sha|map> <canonical-worktree> <name> [<refused-path>…]
```

**The gate fires when the commit changes more than 40 lines, or touches any path outside
the task's declared set** — its `**Files:**` plus its `**Allowed-collateral:**` globs; the
script's header (`<agents repo>/scripts/check-review-gate.sh`) is canonical for how both are
counted. It prints `FIRE: task <id> — …` or `QUIET: task <id> — …` and exits 0; **exit 2 — it
could not judge — stops the run**, as the guard's own exit 2 does.

**The guard's pass ticks an ungated task.** Mark a **task's** checkbox `[x]` (`flow tasks tick`)
once `check-task-commit-fields.sh` exits 0 on its commit and the gate does not fire — no reviewer
runs on that task; the whole-branch panel (`skills/flow/review-panel.md`) is still this branch's
review. **A gated task's tick defers until its reviewer closes clean.** A step's checkbox tracks
the step and gates nothing. A red task's checkbox is ticked together with its partner's, on their
one commit's gate verdict.

**The record carries its own corrections.** When a measurement or a commit disproves what a task
record declares, the record is corrected in place, at the task-close boundary and where the wrong
claim sits — a `predicted:` or `unverified:` tag the run has answered retagged `measured:` with
the command and ref that answered it, a wrong number replaced, a `**Files:**` field widened to the
paths the commit really needed, a fold into a squash partner stated in the task — never left to
disagree silently with what the run did. The implementer never edits `<project>/spectre/changes/`
(**FLOW — COMMIT-PER-TASK**, above): it discloses each correction in its report as the exact
field and its corrected value, and the parent transcribes it into the record before the guard runs
on the commit; inline, the session is both halves of that exchange. A `**Files:**` widening is the
one correction whose disclosure has a mechanical record: `check-task-commit-fields.sh` runs
against the record as it stood, refuses the commit, and names every undeclared path; the parent
judges the deviation legitimate or not on exactly that refusal, and only a legitimate one is
transcribed and the guard re-run green. The review gate's undeclared-path arm reads the paths the
refusal named, so the disclosure cannot disarm the gate. This is expected practice
on every task, not one implementer's habit.

**A deviation from the plan records as a dated `Correction:` paragraph.** When a task's commit
departs from what the plan declares — a file swap, a renamed helper, an added step, any course
the plan did not name — the parent appends a paragraph opening `Correction (YYYY-MM-DD):` to
that task's entry in `tasks.md`, at the task-close boundary where the task's other corrections
are transcribed (before the guard runs on the commit), stating what the plan declared, what
shipped instead, and why. The disclosure route is the same as any correction's: the implementer
reports the deviation, the parent transcribes it; inline, the session is both halves.

**A pivot reconciles the three artifacts together.** When implementation pivots — a reality
discovered mid-run (a route already taken on the base, a capability spec another change already
moved, a premise a measurement disproved) has remaining tasks redesigned rather than implemented
as written — the pivot is resolved **before the colliding code is written**: the run stops at the
discovery, and the redesign is a prompt resolved per **Auto-resolution**
(`skills/flow-contracts/operator-prompts.md`). Only then does the
parent edit `proposal.md`, `design.md` and `tasks.md` in the same pass, never
`tasks.md` alone: `proposal.md`'s `## What changes` is brought to the pivoted scope, the
capability spec the plan edits (`<project>/spectre/specs/<capability>.md`) is rewritten to the pivoted
scope in that same pass, and a
decision the pivot displaces is superseded by ID per **Decisions**
(`skills/flow/brainstorm-planner.md`) — the old entry's `**Status:**` set to `superseded by
<new-id>`, its reasoning retained and its `**Superseded because:**` trigger line appended, a
fresh entry appended, nothing deleted or rewritten — so the record carries the
superseded decision beside its replacement. The amended plan is
re-validated — `spectre validate` and `check-plan-shape.sh` — before the next task dispatches; a
change already archived runs `check-plan-shape.sh` alone, as section 1 does.

**The gated per-task reviewer.** One combined review per gate-fired task — spec compliance and
code quality together — but **one dispatch per bundle of gate-fired tasks, never one per task**,
the discipline **Bundled dispatch** (`skills/flow/review-panel.md`) applies to panel rounds.
**The bundle is the implementer group**: at a boundary, every task of group N+1 whose gate fired
goes out in one reviewer Agent call beside group N+2's implementer, on `opus` at that group's
effort from the decision's `groups` entry — a first-pass review, per **Model and effort**
(`skills/flow/brainstorm-planner.md`). **Groups join into one bundle by the
decision's `class`**: on `big`, one bundle per group; on `micro`, `small` or `regular`, every gate-fired
task of the run waits and goes out in one bundle at the last boundary, on `opus` at the
implementer pair's effort, or `default` when none is recorded; on harness `zcode`, **Harness mapping**
(`skills/flow-contracts/model-policy.md`) replaces it, as it does every pair. **Never one reviewer dispatch per gate-fired task, and never one per
group on `small`/`regular`**. Each task inside the bundle keeps its own pass: its own
commit-range diff `git diff <task-sha>^..<task-sha>` — a real commit diff, never a snapshot of
the working tree, which the next implementer is editing — its own verdict and its own report
file. Record the bundle as one `dispatches` row (`-role reviewer`, `-key task-<n+n+n>-reviewer`
with the task ids `+`-joined in plan order, the same convention as `panel-<round>-<slot|slot+slot+slot>`;
`-task <n>` only on a one-task bundle, omitted otherwise) and close it with **`-outcome clean`
when every pass is clean, `-outcome fix` when any pass is `fix`**; each pass's own verdict is
its report file's `## Verdict`, so per-task review yield stays measurable against the gate.
**A pass whose findings are all Minor is `clean`** — no fix round and no re-review; before
ticking the task the parent fixes each such Minor itself, inline, exactly as **Panel re-runs**
(`skills/flow/review-panel.md`) fixes a Minor-only round's Minors — one commit at the branch tip,
pushed, no dispatch, no dispatch record and no `flow record status` — and only while no
implementer dispatch is writing that worktree. A Minor no change to the tree can resolve is named,
one clause, in the parent's output. A `fix` pass
sends its Minors to the same fix. **A mixed-verdict bundle is handled per task**: every clean task is ticked in the same call that
closes the record, and every `fix` task takes the fix path below on its own sha, independently
of its bundle-mates.

**Load `skills/flow/gated-review-fix.md` only when** a gated reviewer pass closes `fix` — it
carries that fix path.

Every gated reviewer bundle dispatch **must** carry the shared paragraphs below once, then one
**PASS task-`<n>`** section per gate-fired task in plan order, each carrying that task's record
from `tasks.md`, its diff range `git diff <task-sha>^..<task-sha>` and its own REPORT FILE line:

> **MODEL HANDSHAKE:** the first line of your first reply is `Model: <the model named in your own
> system prompt>` and nothing else on that line. Answer it before any tool call.

> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,
> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch in your first turn,
> before anything else. Never ToolSearch for a tool already listed, and never a wildcard query: a
> schema loaded later changes your tool list and re-prices your whole context at full input rate.

> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or poll it to
> completion, before you stop.

> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one check you make MUST exercise the real
> thing. A claim you did not run is worth less than one you did: a doc comment, a type signature
> and a passing test can each read plausibly and be false. Run it before you accept it, and run it
> before you reject it.

> **NO DELEGATION:** Do this work yourself. Never call the `Agent` tool, and never spawn a
> subagent, background agent or helper of any kind — you are the leaf of this run, and any child
> you start is unrecorded and outside the parent's closed list (**Dispatch sites — the parent's
> closed list**, `skills/flow/implement.md`). Reading, searching, reproducing and
> fixing are your own Read, Bash and Edit calls.

> **READ-ONLY REVIEW:** You review a tree other agents are working in — never mutate it. No
> command that writes: no `git checkout`, `git restore`, `git reset`, `git stash`, `git clean`,
> no commit, no index change, and no file edit outside your own report file — the panel's
> mutating slots are the one declared exception, and they work in throwaway copies, not this
> tree. Inspect with Read, Grep, and the read-only git forms — `git show`, `git diff`,
> `git log`, `git status`. A mutation you cause is indistinguishable from a defect the next
> implementer inherits, and a restore you perform is a claim nobody can check: gymie KAN-635's
> reviewer ran `git checkout <sha> -- .` mid-review, destroyed uncommitted planning artifacts,
> and reported the tree restored.

> **INDEPENDENT PASSES:** each pass reviews its own task's diff range and the code, never an
> earlier pass's report or conclusions; a defect that sits in two passes' ranges is raised under
> each. Write each pass's report file before beginning the next pass.

and, inside each **PASS task-`<n>`** section:

> **REPORT FILE:** write this pass's report to
> `<abs-worktree>/.superpowers/sdd/reviewer-report-task-<n>.md` before beginning the next pass,
> the bundle's last one as your **last** act — a `## Verdict` section carrying exactly `clean` or
> `fix`, and each finding with its file, line and severity — Critical, Important or Minor as
> `skills/flow/reviewer-calibration.md` defines them. The verdict is `fix`
> only when a Critical or Important finding stands; Minor-only findings are `clean`. The dispatcher waits on every pass's
> file (`test -s` on each), and a fix round's re-review writes
> `reviewer-report-task-<n>-fix-<k>.md`.

**The plan tree survives every reviewer dispatch — asserted by git, never by the reviewer's
prose.** Before a gated reviewer bundle's Agent call goes out — original, fix re-review, or
handshake retry alike — the parent makes the reviewer-dispatch planning commit (**Planning
commits**, `skills/flow-contracts/git-boundaries.md`) over its own transcriptions and ticks
(task commits never carry plan paths, so uncommitted edits are the only copy — exactly what
`git checkout <sha> -- .` destroyed on gymie KAN-635), then runs
`check-plan-unchanged.sh snapshot <worktree> <name> <snapshot-file>`. When the bundle's last
report file exists, the parent runs `check-plan-unchanged.sh verify <worktree> <name>
<snapshot-file>` **before any verdict is read or acted on**: exit 0, the reports are read; exit
1 ends the turn with `## Question` carrying the guard's lines verbatim; exit 2, the same stop.

**Every dispatch that can touch the worktree is bracketed by content markers — the subagent's
own clean-state claim never answers for the tree.** The plan-tree guard above is one instance of
the general rule, and every other dispatched role gets the same treatment. Before any Agent call
whose child can write to the worktree — an implementer group, a per-task reviewer, a panel round,
the panel-fix subagent — the parent writes the dispatch's marker list beside the run's other
dispatch records: one `<path><TAB><ERE>` line per artifact the dispatch must not damage, each ERE
a string that artifact is known to contain (a decision ID, a task title, a section heading), then
runs `check-tree-markers.sh snapshot <worktree> <markers-file> <snapshot-file>`. When the
dispatch's report file exists, the parent runs `check-tree-markers.sh verify <worktree>
<markers-file> <snapshot-file>` before the report is read or acted on. Exit 1 is the dispatch
having mutated the tree, whatever its report claims — the same `## Question` stop as the
plan-tree guard's, with the guard's lines verbatim; exit 2, the same stop. The markers sit beside
the plan-tree guard, never in its place.

**The last group's guard pass is the stage's last boundary.** `final-review.diff` is written and
the slots dispatched once it has passed, every gate-fired reviewer has closed clean with any fix
landed, and the parent's full-suite run has passed, after any fix of its own; the review panel's
pre-work may share the last implementer's wait, in its one call. A full-suite failure outside the
plan-last group's files ends your turn with `## Question` — the failing
command and its output, verbatim — before `final-review.diff` is written: the panel never runs on
a red branch, and the operator resolves it through a fix run.

**Turn discipline.** A turn is spent only where an output must be read before the next action is
chosen. Calls that do not depend on one another share one Bash call — every verdict printed, each
read afterwards — and launches that do not depend on one another share one message. A stage's
`flow stage begin` rides its first command and its `flow stage end` its last, in the same Bash
call. **A wait on a child is one foreground call, never a chain of idle calls:**

```bash
for i in $(seq 1 48); do test -s <report> && break; sleep 5; done
test -s <report> && echo ready || echo still-running
```

`<report>` is the file the child's REPORT FILE paragraph names — every child kind writes one as
its last act, after its commit and its final test run, so the file's presence is the child's
completion. `still-running` re-issues
the wait, and a ceiling (**No forking, and a wall-clock ceiling on every slot**,
`skills/flow/review-panel.md`) is tracked across the calls. `skills/flow/review-panel.md` and
`skills/flow/verify-and-handoff.md` state their own batches under this paragraph and restate
none of it.

**Read discipline.** These five rules are the run's own size
control,
stated once here and cited — never restated — from `skills/flow/review-panel.md` and
`skills/flow/verify-and-handoff.md` wherever they read a report or a diff:

- **Never `cat` a report.** An implementer, panel-fix, panel or verifier report is read for its
  verdict section only — `sed -n '/^## Verdict/,/^## /p' <report>` or the equivalent for that
  report's own shape — never the whole file. The report file's existence (`test -s`) is the wait
  condition above; its body is read once, narrowly.
- **Never read `final-review.diff`, a dispatch-context bundle, or a panel-fix diff whole.** A slot
  reads the diff it was dispatched against; the parent walks a fix's hunks through `git diff
  --stat` and the specific hunks a finding names, never the whole diff. "Never read the bundle
  back" (`skills/flow/sdd-dispatch.md`) extends to every generated file the parent produces for a child.
- **Test/lint output through `tail`.** A targeted test or lint run's output is piped through
  `tail` (`| tail -20`, the failing block reproduced from the log file on a failure) — already the
  rule for implementers (TARGETED TESTS); it binds the parent's own `## lint`/`## test` runs in
  `flow.verify` and the full-suite run after the last group the same way.
- **Phase files read once per run.** `implement.md`, `review-panel.md`, `verify-and-handoff.md`
  — and `document-fix.md`, `cross-repo-worktrees.md`, `sdd-dispatch.md`, `gated-review-fix.md`,
  `review-panel-late-fix.md`, `review-panel-fix-round.md`, `review-panel-optional-slots.md`,
  `review-panel-experimental-slot.md`, `visual-verify.md` and `visual-verify-tooling-analysis.md` when their load directive or stage loads them — are each read in full
  once, at the start of the stage that needs them; a later need is served by
  `grep -n` for the heading plus `sed -n` for that section, never a second full read.
- **Change artifacts read once**, `proposal.md`/`design.md`/`tasks.md` at `flow.load-context`;
  `tasks.md` re-read only through `spectre list --json` — for a change already archived, its
  checkbox count — and `flow tasks tick` output afterward.

The parent's own `## lint`/`## test` runs, reproducer runs and boundary checks are bound by the same two paragraphs:

> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or poll it to
> completion, before you stop.

> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one check you make MUST exercise the real
> thing. A claim you did not run is worth less than one you did: a doc comment, a type signature
> and a passing test can each read plausibly and be false. Run it before you accept it, and run it
> before you reject it.

On BLOCKED: pause and report. Never guess. A question the report carries that takes one
recommended option is not a guess: **Auto-resolution** (`skills/flow-contracts/operator-prompts.md`)
takes it, and only a question with none pauses the run.

**Before closing the stage**, the parent runs this guard:

```bash
check-task-reviewer-single-dispatch.sh <worktree> <change> <session-token>
```

— the token this run stamped on its own dispatches, `<worktree>` the canonical one. Exit 0
proceeds to the stage close below. Exit 1 names every violation of the gated-
per-task-reviewer bundling contract above — a bundle carrying more than one non-retry dispatch, a
retry with no original, a key outside the canonical shape, two gate-fired tasks of the same
implementer group split across separate reviewer bundles, or (on `micro`/`small`/`regular`) more than one
original bundle for the whole run — and is a prompt, resolved on its recommended **Continue** per
**Auto-resolution** (`skills/flow-contracts/operator-prompts.md`), the same shape
`skills/flow/review-panel.md`'s own `check-panel-fix-single-dispatch.sh` handback carries:

> **The gated per-task reviewer broke the bundled-dispatch shape:** <the guard's violation lines>
> - **Continue — the violation stays recorded in this run's output** *(default, recommended)* —
>   the tasks may already be reviewed clean, and the work is real
> - **Stop the run**

Exit 2 stops the run — a close the guard cannot answer for is not a clean close.

```bash
flow stage end -command '/flow' -stage flow.sdd-tdd -outcome completed <name>
```

Once this stage completes, continue into `skills/flow/review-panel.md`.

### One stack for the verification tasks

**Before the first task whose `**Files:**` only add end-to-end specs, fidelity captures or their
baselines goes out** — dispatched under `sdd`, or started by the parent on an `inline` run — the
parent starts the canonical worktree's stack once, by the start rule of **Resolve the run
instructions** (`skills/flow/verify-and-handoff.md`) — its failure, refused-start and
protected-service handling included — then runs any build the plan's Global Constraints name
for those tasks, then `check-dev-stack-fresh.sh <canonical-worktree>`: exit 1 is a failed start,
and exit 2 puts `Freshness: unverified — <the guard's reason>` in every such task's prompt.
Every such task, and every wave copy it runs in, runs against that stack's worktree-resolved
URLs, which its prompt names, and starts, stops and rebuilds nothing. The stack stays up for
`flow.verify`.
