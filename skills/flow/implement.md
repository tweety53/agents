# Implement (SDD + TDD)

Loaded by `skills/flow/SKILL.md` once planning artifacts exist — continuing straight from
`skills/flow/brainstorm.md` on a creating run, or entered directly on a resumed `STARTED` run whose
plan is already ready, or on a fix run at `IN_PROGRESS`.

| Step | Skill | When |
|------|-------|------|
| **2** | resume the workspace created at `flow.create-artifacts`, stated in **2. Isolate the workspace** below | Before the first code change, every run |
| **3** | **superpowers:writing-plans** | Validate the plan; repair `tasks.md` if it is not apply-ready |
| **4** | **superpowers:subagent-driven-development** | Execute the remaining tasks |
| **5** | **superpowers:test-driven-development** | Every implementer dispatch, every task |
| **6** | the review panel | The final whole-branch panel |
| **7** | **superpowers:systematic-debugging** | An unexpected test failure during implementation, or a review-panel finding confirmed as a real defect |
| **8** | **superpowers:verification-before-completion** | Evidence before claiming done |

**Never** invoke `finishing-a-development-branch`. Integration is `skills/flow/integrate.md`'s job.

## The parent orchestrates directly

Sections **1**, **2** and **4** below, `skills/flow/review-panel.md` and
`skills/flow/verify-and-handoff.md` are **the parent's own work** — no resumed subagent runs
`flow.load-context` through `flow.write-in-progress` on its behalf; the running session does it
itself, in its own Bash and Read calls, `flow record`/`flow stage` marks, worktree add/remove,
report reads and diff walks. Every "you" in those files addresses the parent. On a fix run,
section **3** below runs first: the parent resolves the worktree from the state file's `worktrees`
map, runs section 3's inline plan-append and Jira sync, then continues into load-context/isolate/
sdd-tdd on the same session, with no dispatch in between. A fix run's stage order is therefore
document-fix → load-context → isolate (resume) → sdd-tdd → …, so the appended plan is validated
after the fix's edit.

**Resolve `DEFAULT_MODEL` and `REVIEWERS`** per **Model resolution** (`skills/flow/SKILL.md`), and
run the guard-presence check, before this run's first dispatch. Read the decision JSON
(`<abs-worktree>/.superpowers/sdd/decision.json`) for the three resolved toggles —
`EXECUTION_MODE_TOGGLE`, `IMPLEMENTER_MODEL_TOGGLE`, `REVIEW_PANEL_TOGGLE` — per **The `##
Decision` block** (`design.md`), and for the recorded `groups` field, which section **4** below
dispatches by.

**Never end a turn with a child in flight** — wait for every implementer, reviewer, slot or fix
subagent launched before reporting a stage boundary or asking the operator anything. **Turn
discipline**, below, states the one-foreground-wait-call shape this applies through.

### Dispatch sites — the parent's closed list

These four rows are **every** Agent-tool dispatch the parent may make, across sections **1**,
**2** and **4** below, `skills/flow/review-panel.md` and `skills/flow/verify-and-handoff.md`:

| Site | Role | Key shape | Owning section |
|---|---|---|---|
| implementer, one per group | `implementer` | `task-<n>-implementer` | section **4** below |
| panel bundle, at most two per round | `reviewer` | `panel-<round>-<slot+slot>` | `skills/flow/review-panel.md`, **Bundled dispatch** |
| panel-fix, one per chunk of at most 10 findings | `panel-fix` | `panel-fix-<round>[-<chunk>]` (`-retry` once per chunk) | `skills/flow/review-panel.md`, the fix step |
| verifier, one per worktree | `verifier` | `visual-verify` (`-2`, `-retry`) | `skills/flow/verify-and-handoff.md`, **Visual verification** |

**Everything else in those five sections is the parent's own Bash and Read work, never
delegated** — every `check-*.sh`, `run-reproducer.sh`, `gather-dispatch-context.sh`,
`prepare-workspace.sh`, `## lint` and `## test`, every `flow record` and `flow stage` call,
worktree add and remove, every report read and every diff walk. Not to a "verify" reader, a
"re-verify" or "mutation re-verify" agent, a helper, a background task, or a subagent under any
other name — the KAN-449 run's six unrecorded subagents (four rogue panel-fix dispatches, a
"verify fixes" reader and a "mutation re-verify" agent) are exactly the shape this forbids.

**The self-check.** Before any Agent-tool call, the parent names which row above the call is. A
call that names no row is not made.

**These four rows are the whole run's dispatch tree.** Every row's own prompt carries the NO
DELEGATION paragraph (section **4** below, `skills/flow/review-panel.md`,
`skills/flow/verify-and-handoff.md`) — a leaf never dispatches, so nothing exists below these rows.
**The `flow-<model>-<effort>` family (`agents/flow-*.md`) carries a `tools:` allowlist that omits
`Agent`** — the NO DELEGATION paragraph is now backed by a capability the dispatched agent
structurally does not have, not only by prompt text (KAN-487). This covers the panel bundle and
panel-fix rows whenever `REVIEW_PANEL_TOGGLE` is `dynamic` (`skills/flow/review-panel.md`'s own
**The roster**). The verifier row is unaffected regardless of any toggle: it dispatches
`subagent_type: general-purpose` unconditionally (`skills/flow/verify-and-handoff.md`) — a
harness-provided type this repository does not own and cannot restrict this way. `general-purpose`
is likewise what a reviewer row dispatches on `REVIEW_PANEL_TOGGLE: default`.

**Inline — the parent implements** below takes this same table minus the implementer and panel-fix
rows — the parent's only permitted dispatches inline are the panel-bundle and verifier rows.

**The handshake — stated once here, cited everywhere else.** Every dispatched role in this
pipeline — implementer, panel slot, panel-fix, verifier — opens its first reply with the `Model:`
line the MODEL HANDSHAKE paragraph (section **4** below) demands, and every dispatch prompt in
this pipeline carries that paragraph verbatim. Compare the line against the model this dispatch
requested (`DEFAULT_MODEL`, or the run's session override). A match proceeds. A **first** mismatch
closes the open dispatch row `-outcome fallback` and re-dispatches once, on the same requested
model and `subagent_type`, under `<key>-retry`:

```bash
flow record dispatch end -change <name> -key <key> -session-token mf-<literal-token> \
  -outcome fallback -ended-at <ts>
flow record dispatch begin -change <name> -role <role> -model <the model originally requested> \
  -key <key>-retry -agent-id <id> -session-token mf-<literal-token> -started-at <ts>
```

A **second** mismatch closes the retry row `-outcome fallback` too and the parent asks the
operator directly through **AskUserQuestion**, naming the requested model and both models that
actually answered, options **Continue on `<the model the second handshake named>`** — proceed on
that running agent, no third dispatch — or **Stop the run**. **A mark or a record never blocks** —
proceed on the handshake's outcome regardless of whether any `flow` call reached the store.

This rule is cited, never restated, at every dispatch site in this pipeline: the implementer
dispatch below; `skills/flow/review-panel.md`'s panel slot and panel-fix subagent dispatch;
`skills/flow/verify-and-handoff.md`'s verifier dispatch (compared against `sonnet`, never
`DEFAULT_MODEL`).

**The return.** Once a dispatch's report file appears, read its verdict, print the change's own
handoff or continue to the next stage, and close the record under whichever key is open:

```bash
flow record dispatch end -change <name> -key <the key currently open> -session-token mf-<literal-token> \
  -outcome completed -ended-at <ts>
```

**A dispatch whose agent dies is closed with `-outcome aborted`, reported, and not retried**: print
`/flow <name>` for the operator — a re-run resumes from whatever was left (checkbox state, the
state file's worktrees, findings in the store) through this file's own re-entry rules, and the
operator should see the death rather than have it hidden by a second dispatch.

**Bugbot and Security are prompt-driven roles, dispatched general-purpose like every other panel
slot** (**The roster**, `skills/flow/review-panel.md`) — never a fixed `bugbot` or `security-review`
Agent-tool type, so there is nothing for the parent to substitute.

## Inline — the parent implements

Entered instead of dispatching an implementer per group when the recorded decision's `execution`
is `inline` (**The `## Decision` block**, `design.md`). The parent itself runs sections **1**, **2**
and **4** below, then `skills/flow/review-panel.md` and `skills/flow/verify-and-handoff.md`, with
these substitutions:

- **No implementer, no panel-fix dispatch.** The parent does each bundle's TDD work in the
  canonical worktree, commits per task with the same `Task-Id:` trailer and declared `**Commit:**`
  subject, runs `check-task-commit-fields.sh` and ticks the task exactly as section **4** states.
  Waves are not parallel inline: bundles run in plan order, one at a time, never launched into a
  throwaway worktree.
- **Every dispatch-prompt paragraph that instructs an implementer or fixer** — FLOW —
  COMMIT-PER-TASK, the TDD sub-skill, TARGETED TESTS, MUTATION PROOF, PLAN FIELDS, FOREGROUND
  BUILDS, and the rest section **4** and `skills/flow/review-panel.md` list — **binds the parent
  in the same words**, as if the parent had dispatched itself.
- **Panel slots and the visual-verify verifier dispatch exactly as in sdd mode** — a session
  reviewing its own diff is not a review. Panel fixes are applied by the parent instead of a
  panel-fix subagent; the parent still runs every reproducer and the fix-diff walk
  (`skills/flow/review-panel.md`) before recording a finding `fixed`. The parent's own permitted
  dispatches inline are the closed list's panel-bundle and verifier rows alone
  (**Dispatch sites — the parent's closed list** above); `flow.verify` runs inline for the parent
  exactly as under `sdd` execution.
- **Records:** one `dispatches` row per bundle, `-role implementer -model <parent model> -effort
  <parent effort> -agent-id inline`, and one per fix round, `-role panel-fix -model <parent
  model> -effort <parent effort> -agent-id inline` — so cost attribution and the stats views see
  inline work under the same roles a dispatched run would use.

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
naming: nothing has been proposed under that name, so stop and suggest `/flow <name>`.

Exit `0` from `check-plan-shape.sh` proceeds. Exit `1` names a shape defect in this plan's own
`**Files:**`, `**Tests:**` or other declared fields — repaired here, before any task is dispatched
and before any code is touched, exactly as an exit-1 from `spectre validate` already is. Exit `2`
stops the run.

**The change root is `<project>/spectre/changes/<name>/`, by construction.** Read:

- `<changeRoot>/proposal.md` — what and why
- `<changeRoot>/tasks.md` — the plan
- `<changeRoot>/design.md` — how, when the change carries one
- `<project>/spectre/specs/<capability>.md` for every capability the proposal names

**Whether there is anything left to implement is read off the task checkboxes**, from
`spectre list --json`'s `{"changes":[{"id","done","total"}]}` for this change:

- `total == 0` → no plan spectre can read: stop, resume at `skills/flow/brainstorm.md`.
- `total > 0` and `done == total` → every task is already checked: proceed to
  `skills/flow/integrate.md`.

Confirm `tasks.md` meets writing-plans quality; if it does not, invoke **superpowers:writing-plans**
to repair it before touching code.

Extract the **Global constraints** verbatim from the capability specs the proposal names and
`design.md` for the reviewers.

```bash
flow stage end -command '/flow' -stage flow.load-context -outcome completed <name>
```

## 2. Isolate the workspace (first run only)

**Load `skills/flow-contracts/artifacts-registry.md`** — the worktree and branch created at
`flow.create-artifacts` are rows in it.

```bash
flow stage begin -command '/flow' -stage flow.isolate-workspace -harness <harness> -session-token mf-<literal-token> <name>
```

Resume `<project>/.worktrees/<name>`, created at `flow.create-artifacts`. Never implement on the
default branch without explicit consent.

Record each worktree's merge base and absolute path in this run's own working notes as soon as the
worktree exists — the state file's `worktrees` map is written only at the end of
`skills/flow/verify-and-handoff.md`.

**Load `skills/flow-contracts/worktree-resolution.md`** — it derives this run's resolved worktree
set.

**This run's resolved worktree set — the set `skills/flow/verify-and-handoff.md` iterates — is the
worktree resumed above, plus any additional worktree this change affects.** Per
**Resolving a change's worktrees** (`skills/flow-contracts/worktree-resolution.md`), non-empty by
construction on every ordinary run.

**After creating each additional worktree, run
`spectre link --root <abs-worktree>/spectre <canonical-peer>:<name>` with the working directory at
that repository's primary checkout.** The working directory is what resolves the peers file's
relative entries — `ResolvePeer` stats a declared peer path against the process working
directory, so from inside a worktree `../<peer>` resolves into `<project>/.worktrees/` and the
link is always refused — and `--root <abs-worktree>/spectre` is what writes the satellite-side
`link.md` into the worktree, where `check-unfinished-work.sh` reads it at integrate.
`<canonical-peer>` is the canonical repository's own name in that worktree's
`<project>/spectre/peers` file. Record what the command wrote alongside that worktree's merge
base in this run's working notes. **A refusal is a hard failure of this stage**: report it and
stop the run — a change whose cross-repo link cannot be established lands at integrate with a
false OUTSTANDING verdict that forces hand verification. A change with one worktree runs
nothing here.

**Then run `flow workspace-id <name>` for this worktree's workspace id**, once per run, on a fix
run exactly as on the first.

```bash
flow stage end -command '/flow' -stage flow.isolate-workspace -outcome completed <name>
```

## 3. Documenting a fix, before implementing it

**Parent work, run before the plan is executed** — see **The parent orchestrates directly**
above. Everything below is the parent's own.

**Fix runs only** — a first run creates the worktree instead, per **2** above, and marks nothing
here:

```bash
flow stage begin -command '/flow' -stage flow.document-fix -harness <harness> -session-token mf-<literal-token> <name>
flow record dispatch begin -change <name> -role planner -model <PLANNING_MODEL> \
  -key planner-fix-<n> -session-token mf-<literal-token> -started-at <ts>
```

Record what changed **before** writing code, so the proposal never goes stale. `<n>` is this fix
run's own ordinal — one more than the number of fix rounds already recorded in `proposal.md`/
`tasks.md` or as `<name>-fix-N` sub-changes, the same `N` the "where should it go" prompt's
sub-change option below names — so the first fix run's dispatch is `planner-fix-1`, the second
`planner-fix-2`, and so on. Dispatch this fix's planner the same way **Dispatch the planner**
(`skills/flow/brainstorm.md`) dispatches a creating run's — same handshake, same `opus` fallback
**and the same key-suffix rule that section states: the opus re-dispatch records under
`planner-fix-<n>-opus`, and a second mismatch's under `planner-fix-<n>-<model>`, never a repeat of
`planner-fix-<n>`** — same relay contract, same `Model:` first line — with the fix instructions in
place of the design checklist; that section is canonical for the mechanics and is not restated
here.

The planner opens with a `## Question` asking where the fix should go, relayed through the parent
exactly as any other, shape per Operator prompts (`skills/flow-contracts/operator-prompts.md`):

> **This fix has to be recorded before it is written — where should it go?**
> - **Append to `proposal.md` and `tasks.md`** *(default, recommended)* — nothing new is created
> - **Create a linked `<name>-fix-N` sub-change** — its own proposal and plan, for a fix that adds
>   scope the parent change does not describe

The planner writes the append, or the sub-change's own proposal and plan, and returns `## Plan`.
**The Jira description sync stays in the parent** — never the planner's job. **Load
`skills/flow-contracts/jira-integration.md`.** If the fix adds scope the linked Jira issue does not
describe, sync the issue **description** per **Description sync** in Jira integration
(`skills/flow-contracts/jira-integration.md`). Never transition the issue here.

```bash
flow stage end -command '/flow' -stage flow.document-fix -outcome completed <name>
flow record dispatch end -change <name> -key <the key currently open> -session-token mf-<literal-token> \
  -outcome completed -ended-at <ts> -agent-id <id>
```

`<the key currently open>` is `planner-fix-<n>` on a clean handshake, `planner-fix-<n>-opus` after
one mismatch, or `planner-fix-<n>-<model>` after a second — the same rule stated above.

## 4. Execute (SDD + TDD)

```bash
flow stage begin -command '/flow' -stage flow.sdd-tdd -harness <harness> -session-token mf-<literal-token> <name>
```

**At most one implementer subagent may be in flight against a given worktree at any moment** —
a reviewer is not one: it reads an immutable commit range, and any number of them may run beside
the one implementer. Dispatches into different worktrees remain free to run concurrently. This
explicitly overrides `superpowers:subagent-driven-development`'s parallel dispatch guidance and
`superpowers:dispatching-parallel-agents` for same-worktree tasks.

**The parent records each dispatch in two calls — one as it goes out, one as it comes back.**
Immediately before dispatching:

```bash
flow record dispatch begin -change <name> -task <n> -role implementer -model <m> \
  -key task-<n>-implementer -agent-id <id> -session-token mf-<literal-token> -started-at <ts>
```

and as soon as that dispatch reports back, before the next one goes out:

```bash
flow record dispatch end -change <name> -key task-<n>-implementer \
  -session-token mf-<literal-token> -commit <sha> -outcome completed -ended-at <ts> \
  -agent-id <id>
```

**Both calls are required. Every launch is asynchronous and returns the agent's identifier at
launch, so `begin` carries `-agent-id <id>` and is recorded immediately after the launch returns,
before any other action; `end` may repeat the id.** `-key` is this dispatch's own literal label,
unique within the run's session token — `task-<n>-implementer`, reused identically in both calls.
`-role` is one of `implementer`, `reviewer`, `panel-fix` or `verifier` (**Verify**,
`skills/flow/verify-and-handoff.md`); `-task` is the task's
flat integer id, omitted for a dispatch against no single task; `-started-at`/`-ended-at` are
RFC 3339 — `-started-at` the launch time. `-session-token` takes a literal, never a shell
substitution. Two dispatches starting at one instant are told apart only by id, and a resumed
dispatch shares its id with the original — which is why the id is recorded at launch.

**`-model` is the model this dispatch was actually given — `DEFAULT_MODEL`** (`skills/flow/SKILL.md`'s
**Model resolution**), or the run's session-instruction override when one was given for the
implementer role; on `IMPLEMENTER_MODEL_TOGGLE` `dynamic` it is the group's own `model` from the
decision's `groups` entry, `-effort` its `effort`, and the dispatch's `subagent_type` is
`flow-<model>-<effort>`. Name it explicitly — never by omission. A slot whose model the dispatcher cannot
read records the literal `unknown (agent-defined)` and never a guess.

**A record write never blocks.** An unreachable store journals the intent, prints one warning line,
and exits 0 — never branch on this command's exit code as a signal about the record.

Run `plan-dispatch-bundles.sh <changeRoot>/tasks.md` for this plan's bundles:

```bash
plan-dispatch-bundles.sh <changeRoot>/tasks.md
```

Exit 0 proceeds. A non-zero exit is a plan defect: exit 1 names a task missing its `**Files:**`
field, repaired by `superpowers:writing-plans` before any dispatch happens; exit 2 stops the run.

**Dispatch one implementer per group, not per bundle.** The unit is the recorded decision's
`groups` entry — `{bundles, model, effort}`, `bundles` an array of bundle ids from the same
`plan-dispatch-bundles.sh` output above, `model`/`effort` what this group's implementer runs on; a
`null` `groups` field, which only inline execution ever records, never reaches this section, since
inline runs bundles in plan order with no implementer dispatch at all. A group's implementer works
its bundles in plan order, one commit per task, carrying that task's own `Task-Id:` trailer — a red
task and its partner make one commit between them — and a `Build: red` task is bundled with, and
commits with, the partner its `**Squash-with:**` field names.

**Waves — concurrent dispatch of ready groups.** Inline (**Inline — the parent implements**
above) runs bundles in plan order, never in waves. A group is ready when every id in the union of
its bundles' `after <k>:` lines **that is not itself a task of one of the group's own bundles** has
landed — committed and guard-passed, by direct commit or pick.
**At most two implementer dispatches are in flight per wave**, on both `## execution mode` values.
A group alone in its wave, with no other group ready alongside it, dispatches into the canonical
worktree and commits directly, exactly as today; two ready groups launch together in one message,
each into its own throwaway worktree created by the sequence below, each copy then running the
project's resolved `## worktree setup` command once before its implementer dispatches. A third or
later ready group queues in plan order and launches, into its own throwaway worktree by the same
sequence, as soon as one of the two in-flight groups is picked — the cap bounds dispatches in
flight, never how many groups may be ready at once:

```bash
git -C <worktree> worktree add --detach <worktree>-wave-group-<g> HEAD
git -C <worktree> diff HEAD --binary | git -C <worktree>-wave-group-<g> apply --allow-empty
git -C <worktree> status --porcelain -z | \
  while IFS= read -r -d '' entry; do
    st="${entry:0:2}"; f="${entry:3}"
    [ "$st" = "??" ] || continue
    mkdir -p "<worktree>-wave-group-<g>/$(dirname "$f")"
    cp -a "<worktree>/$f" "<worktree>-wave-group-<g>/$f"
  done
```

**As wave members return**, each is cherry-picked onto the change branch in plan order — a member
is picked once every plan-earlier member of its wave is picked. The unchanged
`check-task-commit-fields.sh` call (canonical worktree fifth argument, resolved `<name>` sixth)
runs on each picked commit, and the dispatch `end` records the picked sha. A pick conflict or a
guard failure hands the group back to its own implementer — its throwaway worktree rebased onto
the advanced branch HEAD, re-commit, re-pick — while sibling members, queued groups and
already-ready later waves are unaffected. A copy is removed once its group is picked, or after
handback resolves. A member reporting BLOCKED follows the existing BLOCKED handback. The
one-implementer-per-worktree rule is untouched: each wave member has its own worktree.

**Gather one context bundle per group, immediately before that group's implementer goes out.**
Take `<g>` and the union of the ids from the `bundle <k>: <ids>` lines `plan-dispatch-bundles.sh`
printed for every bundle in the group, comma-separated:

```bash
mkdir -p <worktree>/.superpowers/sdd
gather-dispatch-context.sh <worktree> <changeRoot> <name> <principles-path> \
  <worktree>/.superpowers/sdd/dispatch-context-group-<g>.md <id>[,<id>…] <canonical-worktree> <shape>
```

`<shape>` is this change's shape, computed once per run from the resolved worktree set — more than
one repository → `cross-repo`, otherwise `single-repo` — and passed on every gather this run
makes, this file's and `skills/flow/review-panel.md`'s alike. The context bundle's `## hazards`
section is filtered by it; a gather made without it carries only always-on hazards.

where `<changeRoot>` is `<project>/spectre/changes/<name>/` resolved inside this worktree, and
`<principles-path>` is the **absolute** path of `engineering-principles.md` **beside this file** —
`skills/flow/`, always.

`<canonical-worktree>` — the seventh argument, passed on every call — is
the member of this run's resolved worktree set whose own
`<project>/<spec-root>/changes/<name>/tasks.md` exists, the same argument
`check-unfinished-work.sh` takes at the integrate gate; on a single-repo change that member is
this worktree and the argument is inert, while on a satellite worktree's group it carries the
canonical plan under labeled sections while keeping this worktree's own project commands,
incidents and HEAD (`<agents repo>/scripts/gather-dispatch-context.sh`'s header is canonical for
the resolution).

A non-zero exit — including the guard being absent — is reported, and
dispatching proceeds with the prompt shape this stage used before this capability existed; the
context bundle never gates a run. Confirm the bundle was actually written (`test -f
<worktree>/.superpowers/sdd/dispatch-context-group-<g>.md`) and report plainly if it is not.
**Never read the bundle back into this context** — `test -f` is the whole check; its content is the
implementer's input, not the dispatcher's. Report the script's stderr line for this stage (`bundle
unchanged — reusing …` or `bundle rebuilt — …`) as part of this stage's own reporting.

The sixth argument scopes the group's `## tasks.md` section to the plan header and the named
tasks' blocks (per design.md's `scope-tasks-not-files`); a named id the plan does not carry is
exit 2, a plan defect reported like a missing `**Files:**` field. The panel's and the fix
subagent's bundles (`skills/flow/review-panel.md`) keep the five-argument call and the whole plan.

Every implementer dispatch **must** carry:

> **FLOW — COMMIT-PER-TASK:** Do **not** run `git push`, merge, or open a PR. As soon as
> RED-GREEN-REFACTOR completes for this task — before the guard runs on it — commit
> your work with `git commit`, carrying a `Task-Id: <n>` trailer. The trailer identifies the task;
> the subject is this task's declared `**Commit:**` field, reproduced exactly. **Never weaken or
> bypass a project's commit validation to fit** — no `--no-verify`. You **may** `git add`/`git
> commit` your own work, but never `<project>/spectre/changes/` or `<project>/docs/superpowers/`.
> **A capability spec under `<project>/spectre/specs/` is your work, not theirs**: when this task's
> `**Files:**` names one, edit it and commit it here, in this task's own commit.

**A `Build: red` task is dispatched with its `Squash-with:` partner, in one bundle.** The
implementer runs the red task first — writes its tests, runs them, and reports the failing
output — then the green partner, and makes **one** commit for the pair: the partner's declared
`**Commit:**` subject and `Task-Id: <partner>` trailer. The red task never has a commit of its
own; `check-task-commit-fields.sh` resolves the pair from either id against that commit.

> **REQUIRED SUB-SKILL:** Use superpowers:test-driven-development — RED-GREEN-REFACTOR for this
> task. Delete any code written before its test.

> **REQUIRED SUB-SKILL:** When a test fails for a reason RED-GREEN-REFACTOR did not plan, invoke
> superpowers:systematic-debugging before writing a fix. An expected RED step needs no invocation.

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
> Establish the real API before writing against it, and report what you found. When what you
> measure contradicts the plan, stop and report the measurement: see **When a measurement
> contradicts the plan** (`skills/flow-contracts/plan-provenance.md`).

Every implementer dispatch **must** also carry:

> **MODEL HANDSHAKE:** the first line of your first reply is `Model: <the model named in your own
> system prompt>` and nothing else on that line. Answer it before any tool call.

The parent compares that line against `DEFAULT_MODEL` (or the run's session override) and
applies **The handshake** stated above, unchanged: a first mismatch is a fallback plus one retry
under `<key>-retry`; a second is a fallback plus `## Question`.

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

> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one test you write MUST exercise the real
> thing. A test backed by a fake or a hand-built value passes while the real integration is broken:
> the shape you construct by hand is not the shape the real producer emits. Build the value the way
> production builds it, or assert against the real boundary.

> **REPORT FILE:** write your report to
> `<abs-worktree>/.superpowers/sdd/implementer-report-<k>.md` as your **last** act — after your
> commit and your final test run — carrying each commit's sha, the failing RED output you saw,
> and anything the plan's `unverified:` tags asked you to establish. The dispatcher waits on that
> file's presence; a resumed fix writes `implementer-report-<k>-fix-<n>.md` instead.

**The plan-last group's implementer dispatch — the group holding the last `bundle <k>` line
`plan-dispatch-bundles.sh` printed — alone also carries:**

> **FULL SUITE:** Yours is the plan-last group. After GREEN and before your commit, run the
> resolved `## test` list once, in the foreground, in the order the context bundle carries it. A
> failure in a file this task's `**Files:**` field names is yours: fix it and re-run. Any other
> failure is not: record the command and its output verbatim in your REPORT FILE under a `## Full
> suite` heading, unfixed, and still commit your own task. When the plan-last group belongs to a
> shared wave, its implementer does not carry FULL SUITE — **the parent itself**, never a
> subagent, instead runs the resolved `## test` list once on the canonical worktree after that
> wave's final pick passes the guard, and a failure is the same verbatim-output `## Question`
> handback as below. The existing last-boundary sentence about a full-suite failure report keeps
> governing the singleton case.

**The next implementer overlaps the guard.** The unit is the group — the decision's `groups`
entry, one or more bundles `plan-dispatch-bundles.sh` emits. At each boundary, in this order:

1. **Group N+1's implementer commits** and writes its report; the wait above ends.
2. **One Bash call: the implementer's `record dispatch end`, the guard on every commit whose sha
   is new, `flow tasks tick` for every task the guard passed, and group N+2's gather.** The guard,
   the tick and the gather are the parent's own Bash calls, never a subagent's. The
   guard takes the canonical worktree's absolute path (the worktree created or resumed in
   **2. Isolate the workspace** above) as its fourth argument and this run's resolved `<name>` as
   its fifth. No task base is passed: the guard derives the commit's parent itself (KAN-330),
   so the argument a mistyped merge base once corrupted is never typed at all:

   ```bash
   check-task-commit-fields.sh <worktree> <task-id> <task-sha> <canonical-worktree> <name>
   ```

   The guard reads git objects and `tasks.md` only, so it is safe while the tree changes, and
   never stashes, reverts or resets (KAN-442). Every verdict is printed and read before anything
   launches: a nonzero exit sends that task back to the **same implementer**, which re-commits and
   re-runs the guard before anything below; exit 0 ticks the task in the same call.

   **A guard call that times out is inspected before it is retried.** Run
   `git status --porcelain=v2 --branch` and `git stash list` in that worktree first. A
   reverting, rebasing or merging state on the `# branch` lines, a change the run did not
   make, or a stash entry the parent did not push means the tree is not the one the run
   left — end the turn with `## Question` carrying both outputs verbatim; never re-run the
   guard on top of it. (KAN-423: a re-run over a mid-flight revert cost ~55 minutes of hand
   recovery.)
3. **One message launches group N+2's implementer. The next Bash call records its `begin`** —
   the very next action after the launch returns, which is what "recorded immediately after the
   launch returns" above requires.

**When the script cannot be located**, apply `flow-task-commit-fields`'s rules by hand: check the
commit's `Files:` against `git diff --name-only <task-sha>^..<task-sha>`, its `Tests:` against the
commit's diff, and its `Commit:` against the commit's actual subject line.

**The guard's pass is the tick.** Mark a **task's** checkbox `[x]` (`flow tasks tick`) once
`check-task-commit-fields.sh` exits 0 on its commit — no reviewer runs per task; the whole-branch
panel (`skills/flow/review-panel.md`) is this branch's review. A step's checkbox tracks the step
and gates nothing. A red task's checkbox is ticked together with its partner's, on their one
commit's guard pass.

**The last group's guard pass is the stage's last boundary.** `final-review.diff` is written and
the slots dispatched once it has passed and the last implementer's report carries no `## Full
suite` failure; the review panel's pre-work may share the last implementer's wait, in its one
call. A report that records a full-suite failure ends your turn with `## Question` — the failing
command and its output, verbatim — before `final-review.diff` is written: the panel never runs on
a red branch, and the operator resolves it through a fix run.

**Never end a turn with a child in flight** — wait for every implementer launched before
reporting a stage boundary or asking the operator anything.

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
completion. The loop is bounded at 240 s, under the prompt-cache TTL rather than the Bash tool's
ten-minute cap: a wait longer than the TTL re-prices the whole context on return, while a bounded
wait's `still-running` turn reads it at the cache rate and keeps it warm. `still-running` re-issues
the wait, and a ceiling (**No forking, and a wall-clock ceiling on every slot**,
`skills/flow/review-panel.md`) is tracked across the calls. `skills/flow/review-panel.md` and
`skills/flow/verify-and-handoff.md` state their own batches under this paragraph and restate
none of it.

**Read discipline.** The parent now does the reading a resumed conductor once did on a 5-minute
TTL; on the 1-hour TTL a large context costs 0.1x per call and no rewrite, but only if it stays
small enough that a warm call is still cheap. These five rules are the run's own size control,
stated once here and cited — never restated — from `skills/flow/review-panel.md` and
`skills/flow/verify-and-handoff.md` wherever they read a report or a diff:

- **Never `cat` a report.** An implementer, panel-fix, panel or verifier report is read for its
  verdict section only — `sed -n '/^## Verdict/,/^## /p' <report>` or the equivalent for that
  report's own shape — never the whole file. The report file's existence (`test -s`) is the wait
  condition above; its body is read once, narrowly.
- **Never read `final-review.diff`, a dispatch-context bundle, or a panel-fix diff whole.** A slot
  reads the diff it was dispatched against; the parent walks a fix's hunks through `git diff
  --stat` and the specific hunks a finding names, never the whole diff. "Never read the bundle
  back" (**4**, above) extends to every generated file the parent produces for a child.
- **Test/lint output through `tail`.** A targeted test or lint run's output is piped through
  `tail` (`| tail -20`, the failing block reproduced from the log file on a failure) — already the
  rule for implementers (TARGETED TESTS); it binds the parent's own `## lint`/`## test` runs in
  `flow.verify` and the full-suite run after a shared wave the same way.
- **Phase files read once per run.** `implement.md`, `review-panel.md`, `verify-and-handoff.md`
  are each read in full once, at the start of the stage that needs them; a later need is served by
  `grep -n` for the heading plus `sed -n` for that section, never a second full read.
- **Change artifacts read once**, `proposal.md`/`design.md`/`tasks.md` at `flow.load-context`;
  `tasks.md` re-read only through `spectre list --json` and `flow tasks tick` output afterward.

> **FOREGROUND BUILDS:** Never end your turn with a build, test run, or other long-running
> command still executing in the background. Run it in the foreground, or poll it to
> completion, before you stop.

> **REPRODUCE, DON'T READ:** Where a behaviour crosses a boundary — the store, the filesystem, a
> guard, a real transcript, a real process — at least one check you make MUST exercise the real
> thing. A claim you did not run is worth less than one you did: a doc comment, a type signature
> and a passing test can each read plausibly and be false. Run it before you accept it, and run it
> before you reject it.

On BLOCKED: pause and report. Never guess.

```bash
flow stage end -command '/flow' -stage flow.sdd-tdd -outcome completed <name>
```

Once this stage completes, continue into `skills/flow/review-panel.md`.
