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

**Mark `flow.kickoff` now that the name is fixed, and write `STARTED` immediately — before
brainstorming begins.**

```bash
flow stage begin -command '/flow' -stage flow.kickoff -harness <harness> -session-token mf-<literal-token> <name>
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

**Nothing is written before this point on a creating run** but the In Progress transition
above — the state write above is the
first thing this invocation writes once the name is fixed, ahead of even the design conversation.

**Then create the worktree, still inside `flow.kickoff` — before brainstorming, before any file
is read or written for this change.** From this point on, the main checkout is never read, checked
out, staged, committed or written by any phase of `/flow`; every path below resolves inside
`<project>-worktrees/<name>`, including the design spec, `spectre new`, the three artifacts, the
plan and the decision JSON.

Run the kickoff steps 1–4 in one call — 1 `check-worktree-location.sh <project>`; 2 the worktree
add at `<project>-worktrees/<name>`, the sibling of the main checkout, on
`spectre/<name>` tracking the remote branch when a `/flow-plan` capture or an earlier run pushed
it, else a new `spectre/<name>` from `origin/<default-branch>` — never HEAD: the main checkout may
be on any branch and is never moved; 3 the project's `## worktree setup` commands (**Project
configuration**, `skills/flow-contracts/project-configuration.md`); 4 the push, per **Branch
backup** (`skills/flow-contracts/git-boundaries.md`):

```bash
kickoff-worktree.sh <project> <name> [<base>]
```

`<base>` is the creating run's `--base <branch>` (`skills/flow/SKILL.md`), omitted when none was
given: the new branch is then cut from `origin/<base>` instead of `origin/<default-branch>`, and the
base is recorded on it for `resolve-base-branch.sh`; `origin/<base>` missing is exit 1.

Exit 0 prints `worktree: <abs-worktree>` and `merge-base: <sha>`. **Persist the worktree into the
state record before this step returns** — the script does, through `flow state add-worktree` right
after the add, so a later step's failure leaves the worktree recorded. Exit 1 or 2 stops the run
with the script's own lines: **a command's non-zero exit ends your turn** naming the command and
its output.

```bash
flow stage end -command '/flow' -stage flow.kickoff -outcome completed <name>
```

## Run brainstorming and planning through the planner role

```bash
flow stage begin -command '/flow' -stage flow.brainstorm -harness <harness> -session-token mf-<literal-token> <name>
```

**The parent never brainstorms or writes the plan itself; a `planner` dispatch does** (decision
`planner-role`, KAN-1067). Every planner dispatch is fresh and one-shot, on `subagent_type:
flow-planner-<effort>` with `model` and effort from the **Allowed pairs** row `planner`
(`skills/flow-contracts/model-policy.md`) — its first pair unless an open experiment overrides
it. Its prompt carries `Role: planner`, `Key: plan-<n>`, the two agent-baseline sentences, the
report-tersely line, the PLANNING DELEGATION paragraph (`skills/flow-self-review/SKILL.md`) — its
auto-resolution sentence replaced by "every choice is written as a question of the round file" —
the change name, the absolute worktree, the session token and `REVIEWERS` (**Model resolution**,
`skills/flow/SKILL.md`), already in scope from this run's own earlier resolution. `<n>` counts
this run's planner dispatches from 1, and the Agent-tool `description` opens with `plan-<n>` —
`plan-2 design round` — the row `hooks/enforce-agent-baseline.py` admits:

```bash
flow record dispatch begin -change <name> -role planner -model <m> -effort <e> -key plan-<n> -session-token mf-<literal-token>
flow record dispatch end -change <name> -key plan-<n> -session-token mf-<literal-token> -outcome completed
```

1. **Design round.** The planner runs section **B** of `skills/flow/brainstorm-planner.md` without <!-- refs-guard:allow -->
   plan mode, reading the ticket and every file under `<abs-worktree>/.superpowers/sdd/planner/`,
   dispatching its own Explore and Plan agents (it records their rows, `-role explorer|planner`,
   `-key explore-<n>|design-<n>`, each pair from `plan-agent-pair.sh <name> <n> explorer|plan-designer`, `<n>` continuing
   the run's count across rounds — the parent names the next free `<n>` in each round's prompt, since
   a reused key restarts the earlier round's row), publishing or
   updating the prototype a visible change needs, and writes `round-<r>.md`: the design draft, the
   settled prose the convergence confirm shows, then `## Questions` — each question with two to
   four options, one marked recommended, the convergence confirm last when nothing else is open.
   It asks the operator nothing.
2. **Relay.** The parent prints the round's settled prose, asks its questions verbatim in one
   **AskUserQuestion** call, and writes the answers verbatim to `answers-<r>.md`; it never answers
   a question itself. An explicit approve of the convergence confirm closes brainstorming —
   `flow.brainstorm` end, `flow.design-approval` begin and end, as **Convergence**
   (`skills/flow/brainstorm-planner.md`) marks them; any other answer dispatches round `<r+1>`. An
   operator message mid-planning is written to the next `answers-<r>.md` and relayed the same way.
3. **Artifacts and plan.** A final planner dispatch runs sections **C** and **D** of <!-- refs-guard:allow -->
   `skills/flow/brainstorm-planner.md` up to, not including, **Decide**: its `flow.create-artifacts`
   and `flow.writing-plans` begin marks under the run's session token, `spectre new`, the three
   artifacts from the approved design — read from every file under
   `<abs-worktree>/.superpowers/sdd/planner/`, where the approved design lives — writing-plans,
   `spectre validate`, `check-plan-shape.sh`, the configured guards and `flow tasks count`. It
   returns `PLAN-WRITTEN` or the guard output it could not clear, and the parent redispatches a
   fresh one with that output until `PLAN-WRITTEN`; a redispatch skips section **C** and the
   `flow.writing-plans` begin, both already made, and resumes at the guard that failed.

The parent then closes `flow.writing-plans` and runs **Decide**, the record sequence and the gate,
as the rest of this file states.

Run each `begin` immediately before its Agent call and its `end` once that agent returns — never
both after it returns, which records a window of zero seconds (kan-1034: both explorer rows read
0s).

No `fork` from kickoff to the plan gate: a fork copies the whole conversation and doubles the
token spend (kan-941).

**Prose preceding a question is shown too, not dropped.** When the checklist carries a summary
before a question — most concretely the convergence confirm's "state what you believe settled"
paragraph — show that prose to the operator as ordinary text before the **AskUserQuestion** call,
not only the bare question and options.

Once the Decide step finishes, print the `## Decision`
block verbatim — the output of `flow decision render` (**Decide**, `skills/flow/brainstorm-planner.md`) — then run the record
sequence below:

```bash
flow stage begin -command '/flow' -stage flow.decide -harness <harness> -session-token mf-<literal-token> <name>
flow record decision -change <name> -session-token mf-<literal-token> -file <abs-worktree>/.superpowers/sdd/decision.json
flow stage end -command '/flow' -stage flow.decide -outcome completed <name>
```

**A refused `flow record decision`** — exit non-zero naming a pair's JSON path and model — is a
Decide defect, not a store failure: correct that pair in `decision.json` per **Model and effort**
(`skills/flow/brainstorm-planner.md`) and record again before the gate; no dispatch reads a
decision the store refused.

**The gate comes next**: **Plan review gate** (`skills/flow/brainstorm-planner.md`), recording each
re-decision through the sequence above. On **Yes**, make the plan-gate planning commit and push
it (**Planning commits**, `skills/flow-contracts/git-boundaries.md`), then **end the run** with
the `STARTED` handoff (**The `STARTED` handoff block**, below),
`/clear` above `/flow <name>`, on a creating and a resumed run alike; the next run finds
the plan ready (**Resuming at `STARTED`**, `skills/flow/resume.md`) and implements on a fresh context.

### The `STARTED` handoff block

Defined by **The block each state renders** (`skills/flow-contracts/handoff-blocks.md`), cited here and never loaded by this run.

````text
## Plan ready

<2–4 lines: what the plan implements, in product terms, from `proposal.md`'s `## What changes`>

<the `flow decision render` output for the change's latest recorded decision, verbatim>

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
the handoff is the one place the settled plan is guaranteed to be read (kan-941 printed only
`Next:`).
