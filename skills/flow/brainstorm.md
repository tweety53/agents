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
  "planningEffort": null,
  "models": { "default": null },
  "prUrl": null,
  "updatedAt": "<ISO-8601 UTC now>",
  "updatedBy": "/flow"
}
```

`planningEffort` and `models.default` are written `null` and stay `null` for the life of the
change.

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
configuration**, `skills/flow-contracts/project-configuration.md`) and `flow-guard ignore-scratch`,
and pushes the branch (**Branch backup**, `skills/flow-contracts/git-boundaries.md`). `<base>` is
the creating run's `--base <branch>` (`skills/flow/SKILL.md`), omitted when none was given; the
base is recorded on the branch for `flow-guard resolve-base-branch`, and `origin/<base>` missing
is exit 1.

Exit 0 prints `worktree: <abs-worktree>` and `merge-base: <sha>`. Exit 1 or 2 stops the run with
the script's own lines: **a command's non-zero exit ends your turn** naming the command and its
output.

## Run brainstorming and planning through the planner role

**The parent never brainstorms or writes the plan itself; a `planner` dispatch does** (decision
`planner-role`). A `/flow-plan` capture runs this section too, with the differences
**Capturing a new change** (`skills/flow-plan/SKILL.md`) states.

**Every planner dispatch is fresh and one-shot**, on `subagent_type: flow-planner-<effort>` with
the planner role's pair (**Model and effort**, `skills/flow/brainstorm-planner.md`). Its prompt
carries the two agent-baseline sentences, the report-tersely line, the PLANNING DELEGATION
paragraph (`skills/flow-self-review/SKILL.md`) — its auto-resolution sentence replaced by "every
choice is written as a question of the round file" — the change name and the absolute worktree.
The Agent-tool `description` opens with `plan-<n>` — `plan-2 design round` — `<n>` counting this
run's planner dispatches from 1, so the hook records the dispatch; once the planner returns:

```bash
flow record dispatch end -change <name> -key plan-<n> -session-token mf-<literal-token> -outcome completed
```

A failed recording command is one line in the run's narrative, never a stop.

No `fork` from kickoff to the plan gate: a fork copies the whole conversation and doubles the
token spend.

1. **Design round.** The planner runs section **B** of `skills/flow/brainstorm-planner.md` without <!-- refs-guard:allow -->
   plan mode, reading the ticket, `findings.md`, the latest `round-<r>.md` and the latest
   `answers-<r>.md` under `<abs-worktree>/.superpowers/sdd/planner/` — the three paths named in its
   prompt, never an earlier round or answers file — dispatching its own Explore and Plan agents
   (**Explore and design agents**, `skills/flow/brainstorm-planner.md`), writing or updating the
   prototype file a visible change needs and never publishing it, and writes `round-<r>.md`: the
   prototype file's absolute path when the round wrote or changed it, `## Summary` — at most ten
   lines, the decisions the round settled and what is still open, which is the settled prose the
   convergence confirm shows — the design draft, then `## Questions` — each question with two to
   four options, one marked recommended where one honestly is, the convergence confirm last when
   nothing else is open. `## Questions` holds only questions headed `[architectural]` or
   `[impactful: <criterion>]`; the round's auto-taken ones sit under `## Auto-taken`.
   It asks the operator nothing.
2. **Relay.** The parent first publishes the prototype
   (**The checklist**, `skills/flow/brainstorm-planner.md`). The parent prints the round's `## Summary` verbatim and
   the round file's absolute path — the design draft stays in the file — with the
   prototype's Artifact URL beside them whenever a prototype exists, asks its `## Questions` verbatim and no other —
   the `[architectural]` ones first, four to an **AskUserQuestion** call, then the
   `[impactful: <criterion>]` ones, four to a call, the convergence confirm last, in round-file order — the URL appended to the text of the question that depends on the
   prototype, or else of the convergence confirm, since the question dialog may hide the text
   above it — and writes the answers verbatim to `answers-<r>.md`; it never answers
   a question itself. Every later planner dispatch's prompt, the final one included, carries the
   prototype's latest Artifact URL. An explicit approve of the convergence confirm closes
   brainstorming; any other answer dispatches round `<r+1>`. An operator message mid-planning is
   written to the next `answers-<r>.md` and relayed the same way.
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

The parent then runs **Plan review** below, then **Decide** (`skills/flow/brainstorm-planner.md`)
and the gate.

**Prose preceding a question is shown too, not dropped.** When the checklist carries a summary
before a question — most concretely the convergence confirm's "state what you believe settled"
paragraph — show that prose to the operator as ordinary text before the **AskUserQuestion** call,
not only the bare question and options.

Once the Decide step finishes, print the `## Decision` block verbatim — the output of `flow
decision render` (**Decide**, `skills/flow/brainstorm-planner.md`) — then record it:

```bash
flow record decision -change <name> -session-token mf-<literal-token> -file <abs-worktree>/.superpowers/sdd/decision.json
```

**The gate comes next**: **Plan review gate** (`skills/flow/brainstorm-planner.md`), recording each
re-decision the same way. On **Yes**, end the planning mark and **end the run** with the `STARTED`
handoff (**The `STARTED` handoff block**, below), `/clear` above `/flow <name>`, on a creating and
a resumed run alike; the next run finds the plan ready (**Resuming at `STARTED`**,
`skills/flow/resume.md`) and implements on a fresh context.

```bash
flow stage end -command '/flow' -stage flow.plan -outcome completed <name>
```

### Plan review

**A fresh reviewer that has seen nothing but the documents reads every plan once**, after the
final planner returned `PLAN-WRITTEN`, in `/flow`'s creating run and a `/flow-plan` capture. A fix
run's re-plan, the self-review and in-run fix planners and `/flow-fast` never run it.

1. **Dispatch.** One Agent call on the design-reviewer role's pair (**Model and effort**,
   `skills/flow/brainstorm-planner.md`), its `description` opening with `plan-review-1`. The
   prompt is the two agent-baseline sentences, the report-tersely line, the paragraph below, and
   six absolute paths: `<skill-dir>/plan-reviewer-prompt.md` as the file to read in full first,
   `<skill-dir>/plan-review-rubric.md`, the change's `proposal.md`, `design.md` and `tasks.md`
   under `<changeRoot>`, the repository root `<abs-worktree>`, and the report
   `<abs-worktree>/.superpowers/sdd/plan-review-1.md` — `<skill-dir>` the directory this file
   was read from. It carries no ticket text, no Jira key, no round or answers file and no
   summary of the conversation. Once it returns, end its row as a planner's is ended, under key
   `plan-review-1`.
2. **`PLAN-REVIEW-CLEAN`** — go on.
3. **`PLAN-REVIEW-FINDINGS`** — dispatch a fresh final planner, its prompt naming the report's
   absolute path; it triages the report per **Plan review** (`skills/flow/brainstorm-planner.md`).
   Its `QUESTIONS` return is relayed as step 2 of the section above relays a round, and a fresh
   final dispatch takes the answers. Its `PLAN-WRITTEN` return closes the review; nothing is
   reviewed a second time.
4. **A failed reviewer.** A dispatch that errors, or returns neither verdict line, is
   redispatched once as `plan-review-2`. Failing again, the plan review is `not run`, never
   clean: say so in the gate summary and go on.

### The `STARTED` handoff block

Defined by **The block each state renders** (`skills/flow-contracts/handoff-blocks.md`), cited here and never loaded by this run.

````text
## Plan ready

<2–4 lines: what the plan implements, in product terms, from `proposal.md`'s `## What changes`>

<the `flow decision render` output for the change's latest recorded decision, verbatim>

**Prototype:** <the Artifact URL `design.md`'s `## Context` cites — printed only when it cites a prototype; where none could be published, the file's absolute path and that it was not published>
**Open questions:** <each still-open entry's ID and question, one line each — printed only when `design.md` records one>
**Jira:** <the skip reason, or the line appended to the description — printed only when a Jira call failed or the run appended>

Next:
/clear
```
/flow <name>
```
````

**Only what the operator acts on** — on a clean run the block is the summary, the decision and
the next commands; a conditional line prints only when its condition holds.

**The block is the run's final message, whole — never the `Next:` lines alone.** The summary and
the decision sit directly above `Next:` even though the gate already showed them: text printed
beside a question is hidden where the operator sees only a turn's last message (focus mode), so
the handoff is the one place the settled plan is guaranteed to be read.
