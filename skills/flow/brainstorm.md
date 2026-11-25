# Brainstorm and plan

Superpowers Basic Workflow steps **#1** (brainstorming) and **#3** (writing-plans), intertwined
with spectre artifact creation, run here. Loaded by `skills/flow/SKILL.md` on a creating run (no state), and by
`skills/flow/resume.md` when a resumed run still has planning to do.

## A. Resolve the change and write `STARTED`

**Load `skills/flow-contracts/jira-integration.md`** — this section resolves the linked issue and
derives the change name from it.

**Resolve the linked Jira issue first** — it decides the change name. Follow **Resolution (how
`jiraIssue` is decided)** in `skills/flow-contracts/jira-integration.md` exactly. This is the
only phase that resolves a key.

Then the change name:

- **With a linked issue**, first enumerate the candidate set exactly as **Change name resolution (all `/flow*` commands)**
  (`skills/flow-contracts/pipeline.md`) defines it: exactly one candidate whose name
  starts with `<lowercased-key>-` is this change — a `/flow-plan` capture or an earlier run
  already named it — resumed at its recorded state, announcing which; more than one is an
  **AskUserQuestion** listing each (name, state, last modified); none means the name is
  `<lowercased-key>-<slug>`, per **Change naming** (`skills/flow-contracts/jira-integration.md`).
- **Without one**, the name is the descriptive slug alone.
- If a name or description was given, use it (derive kebab-case from the description if only a
  description was given).
- **If both are omitted:** enumerate the candidate set exactly as **Change name resolution (all `/flow*` commands)**
  (`skills/flow-contracts/pipeline.md`) defines it, restricted to changes with incomplete planning
  artifacts.

**Load `skills/flow/resume.md`** only when, on `/flow`, the name resolves to a change already
recorded at `STARTED`, and continue there rather than with the rest of **A**; `/flow-plan` routes
such a name to its own existing-change destination instead.

**Transition the issue to In Progress now**, per **Transitions** in Jira integration
(`skills/flow-contracts/jira-integration.md`) — before brainstorming, so the board is correct
while planning runs. A failure is one skipped-with-reason line and planning continues; nothing about
this call may delay or alter the run.

**Mark `flow.plan` begun and write `STARTED` immediately — before brainstorming begins.** The
mark is a passive record for the stats (**Stage marks**, `skills/flow-contracts/pipeline.md`); it
never refuses, and a failed mark is one line in the run's narrative, never a stop.

```bash
flow stage begin -command '/flow' -stage flow.plan -harness <harness> -session-token mf-<literal-token> <name>
```

```json
{
  "state": "STARTED",
  "branch": null,
  "worktrees": {},
  "artifactUrl": null,
  "jiraIssue": "<resolved key, or null>",
  "prUrl": null,
  "updatedAt": "<ISO-8601 UTC now>",
  "updatedBy": "/flow"
}
```

**Then create the worktree — before brainstorming, before any file is read or written for this
change.** From this point on, the main checkout is never read, checked out, staged, committed or
written by any phase of `/flow`; every path below resolves inside `<project>-worktrees/<name>`,
including the design spec, `spectre new`, the three artifacts, the plan and the decision JSON.

```bash
flow-guard kickoff-worktree <project> <name> [<base>]
```

It adds the worktree at `<project>-worktrees/<name>`, the sibling of the main checkout, on
`spectre/<name>` — tracking the remote branch when a `/flow-plan` capture or an earlier run pushed
it, else a new branch from `origin/<default-branch>`, never HEAD — records it in the state through
`flow state add-worktree`, runs the project's `## worktree setup` commands (**Project
configuration**, `skills/flow-contracts/project-configuration.md`),
and pushes the branch (**Branch backup**, `skills/flow-contracts/git-boundaries.md`). `<base>` is
the creating run's `--base <branch>` (`skills/flow/SKILL.md`), omitted when none was given; the
base is recorded on the branch for `flow-guard resolve-base-branch`, and `origin/<base>` missing
is exit 1.

Exit 0 prints `worktree: <abs-worktree>` and `merge-base: <sha>`. Exit 1 or 2 stops the run with
the script's own lines: **a command's non-zero exit ends your turn** naming the command and its
output.

## Run brainstorming and planning through the planner role

**The parent never brainstorms or writes the plan itself; a `planner` dispatch does**. Before the first dispatch, write the ticket — the Jira issue, or the operator's
request verbatim — to `<abs-worktree>/.superpowers/sdd/planner/ticket.md`. A `/flow-plan` capture runs this section too, with the differences
**Capturing a new change** (`skills/flow-plan/SKILL.md`) states.

**Every planner dispatch is fresh and one-shot**, on `subagent_type: flow-planner-<effort>` with
the `planner` role's pair (**Model and effort**, `skills/flow/brainstorm-planner.md`). Its prompt
carries the Role line, `Role: planner`, the two agent-baseline sentences, the
report-tersely line, the **PLANNING DELEGATION** paragraph below, the change name and the absolute
worktree.
The Agent-tool `description` opens with `plan-<n>` — `plan-2 design round` — `<n>` counting this
run's planner dispatches from 1, so the hook records the dispatch; once the planner returns:

```bash
flow record dispatch end -change <name> -key plan-<n> -session-token mf-<literal-token> -outcome completed
```

A failed recording command is one line in the run's narrative, never a stop.

> **PLANNING DELEGATION:** You may dispatch fresh `Explore` and `Plan` subagents through the
> `Agent` tool while you plan, per **B. Basic Workflow #1 — Brainstorming** (`skills/flow/brainstorm-planner.md`) — never a
> `fork`, and never an agent that edits files or commits. Each dispatch runs on its own Role's pair
> (**Model and effort**, `skills/flow/brainstorm-planner.md`), and every dispatch prompt carries the line
> `Role: explorer` (an Explore agent) or `Role: plan-designer` (a Plan agent) and the two sentences of **Propagate this**
> (`~/.claude/rules/agent-baseline.md`) verbatim. The agents investigate and propose; you
> decide, and every choice is written as a question of the round file. Fixing, implementing and
> reviewing are never delegated.

No `fork` from kickoff to the plan gate: a fork copies the whole conversation and doubles the
token spend.

1. **Design round.** The planner runs section **B** of `skills/flow/brainstorm-planner.md` without <!-- refs-guard:allow -->
   plan mode, reading the ticket, `findings.md`, the latest `round-<r>.md` and the latest
   round file's `## Answers` under `<abs-worktree>/.superpowers/sdd/planner/` — the three paths named in its
   prompt, never an earlier round file — reading the code itself, and dispatching its
   own Explore and Plan agents only where that cannot settle a question (**Explore and design
   agents**, `skills/flow/brainstorm-planner.md`), writing or updating the
   prototype file a visible change needs and never publishing it, and writes `round-<r>.md`: the
   prototype file's absolute path when the round wrote or changed it, `## Summary` — at most ten
   lines, the decisions the round settled and what is still open, which is the settled prose the
   convergence confirm shows — the design draft, then `## Questions` — each question with two to
   four options, one marked recommended where one honestly is, the convergence confirm last when
   nothing else is open. `## Questions` holds only the questions the operator is asked; the
   round's auto-taken ones sit under `## Auto-taken`.
   It asks the operator nothing.
2. **Relay.** When the round wrote or changed the prototype, the parent publishes it with the
   Artifact tool. The parent prints the round's `## Summary` verbatim and the round file's
   absolute path — the design draft stays in the file — with the prototype's Artifact URL beside
   them whenever a prototype exists, asks its `## Questions` verbatim and no other —
   four to an **AskUserQuestion** call in round-file order, the convergence confirm last — the URL appended to the text of the question that depends on the
   prototype, or else of the convergence confirm, since the question dialog may hide the text
   above it — and appends the answers verbatim to that round file under `## Answers`; it never answers
   a question itself. A round whose `## Questions` holds only the convergence confirm is asked
   with the round file's full design draft as the approve option's `preview` (**Convergence**,
   `skills/flow/brainstorm-planner.md`). Every later planner dispatch's prompt, the final one included, carries the
   prototype's latest Artifact URL. An explicit approve of the convergence confirm closes
   brainstorming; any other answer dispatches round `<r+1>`. An operator message mid-planning is
   appended to the latest round file's `## Answers` and relayed the same way.
3. **Artifacts and plan.** One final planner dispatch runs sections **C** and **D** of <!-- refs-guard:allow -->
   `skills/flow/brainstorm-planner.md` — `spectre new`, the three artifacts from the approved
   design (read from the approved round file, its answers and `findings.md`), the plan,
   `spectre validate` — and returns `PLAN-WRITTEN`. It returns `QUESTIONS` instead the moment
   writing the artifacts or the plan surfaces an ambiguity (**The checklist**,
   `skills/flow/brainstorm-planner.md`), the next `round-<r>.md` holding it: the parent relays that
   round as step 2 does, and a fresh final dispatch — never a design round — takes its answers,
   revising the design, the artifacts and the plan they change. A dispatch that returns a failure
   it could not clear is redispatched fresh once, its prompt carrying that output and the absolute
   paths of `proposal.md`, `design.md` and `tasks.md`; a second failure stops the run with both
   returns.

The parent then runs **Plan check** below.

**Prose preceding a question is shown too, not dropped.** When the checklist carries a summary
before a question — most concretely the convergence confirm's "state what you believe settled"
paragraph — show that prose to the operator as ordinary text before the **AskUserQuestion** call,
not only the bare question and options.

Once the plan check finishes, render the decision — the parent never decides itself — then record
it:

```bash
flow decision render -file <abs-worktree>/.superpowers/sdd/decision.json -session-model '<the model named in this session's own system prompt>'
flow record decision -change <name> -session-token mf-<literal-token> -file <abs-worktree>/.superpowers/sdd/decision.json
```

The render exits 0 on a printed block and 2 on a usage error or an unreadable file.

**`/flow` has no gate question.** It ends the planning mark and **ends the run** with the `STARTED`
handoff (**The `STARTED` handoff block**, below), `/clear` above `/flow <name>`, on a creating and
a resumed run alike; the next run finds the plan ready (**Resuming at `STARTED`**,
`skills/flow/resume.md`) and implements on a fresh context. `/flow-plan` goes on at **The Save the
plan? gate** (`skills/flow-plan/SKILL.md`) instead.

```bash
flow stage end -command '/flow' -stage flow.plan -outcome completed <name>
```

**A plan the operator wants changed** — on `/flow`, a message before the handoff's `/flow <name>`
runs: take the operator's changes, write them verbatim to the latest round file's `## Answers`, and
dispatch a fresh final planner on them as a `QUESTIONS` return is taken (step 3 above) — it revises
`tasks.md` and re-runs `spectre validate "<name>"`. Then re-decide as **Plan check** below states and
print the `STARTED` handoff again.

### Plan check

**One headless pass reads every plan once and decides how it is built**, after the final planner
returned `PLAN-WRITTEN`, in `/flow`'s creating run and a `/flow-plan` capture. A fix run's re-plan never runs it.

1. **Run it.**

   ```bash
   flow-guard plan-check <worktree> <changeRoot> <abs-worktree>/.superpowers/sdd/decision.json
   ```

   Its header is canonical for the call, its fallback and its exits. It writes
   `<abs-worktree>/.superpowers/sdd/plan-check.md`, the decision and, when the report states one,
   the execution confidence as `<abs-worktree>/.superpowers/sdd/confidence.json`; exit 2 stops the
   run with its stderr, and so does exit 1: the decider's decision was unusable twice and none is
   written — there is no fallback decision, the run is never planned by a made-up one.
   `skills/flow/plan-check-prompt.md` is canonical for the lenses, the rules and the shape. The
   report's last line is `PLAN-CHECK-CLEAN`, `PLAN-CHECK-FINDINGS <n>` or `PLAN-CHECK-NOT-RUN`.
2. **`PLAN-CHECK-CLEAN`** and no `confidence <n>%` below 100 on the guard's stdout line — go on.
3. **`PLAN-CHECK-FINDINGS`, or a `confidence <n>%` below 100** — dispatch a fresh final planner, its prompt naming the report's
   absolute path; it triages the report per **Plan review** (`skills/flow/brainstorm-planner.md`).
   Its `QUESTIONS` return is relayed as step 2 of the section above relays a round, and a fresh
   final dispatch takes the answers. Its `PLAN-WRITTEN` return closes the check; nothing is
   checked a second time.
4. **`PLAN-CHECK-NOT-RUN`** — the check could not run: say so on the `**Plan check:**` line and go
   on, never as clean.

**A changed `tasks.md` re-decides once.** `plan-check … -decide-only` rewrites `decision.json` and
leaves `plan-check.md` as it is. **That re-run is the only one: it happens once whenever `tasks.md`
changed after the check — a triage fix or an operator's change — and the decision is recorded
again.**

### The `STARTED` handoff block

Defined by **The block each state renders** (`skills/flow-contracts/handoff-blocks.md`), cited here and never loaded by this run.

````text
## Plan ready

<2–4 lines: what the plan implements, in product terms, from `proposal.md`'s `## What changes`>

<the `flow decision render` output for the change's latest recorded decision, verbatim>

**Design:** <absolute path of design.md>
**Plan check:** <absolute path of the report — clean | <n> fixed, <n> accepted | not run — no confidence estimate (the report ends `PLAN-CHECK-NOT-RUN`; never a bare `not run`), followed by ` · confidence <n>% at plan check` when the guard's stdout line carried a number>
**Prototype:** <the Artifact URL `design.md`'s `## Context` cites — printed only when it cites a prototype; where none could be published, the file's absolute path and that it was not published>
**Open questions:** <each still-open entry's ID and question, one line each — printed only when `design.md` records one>
**Jira:** <the skip reason, or the line appended to the description — printed only when a Jira call failed or the run appended>

Next:
/clear
```
/flow <name>
```
````

**Only what the operator acts on** — on a clean run the block is the summary, the decision, the
design and plan-check lines and the next commands; a conditional line prints only when its
condition holds.

**The block is the run's final message, whole — never the `Next:` lines alone.**
