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

Run the kickoff steps 1–4 in one call — 1 `flow-guard check-worktree-location <project>`; 2 the worktree
add at `<project>-worktrees/<name>`, the sibling of the main checkout, on
`spectre/<name>` tracking the remote branch when a `/flow-plan` capture or an earlier run pushed
it, else a new `spectre/<name>` from `origin/<default-branch>` — never HEAD: the main checkout may
be on any branch and is never moved; 3 the project's `## worktree setup` commands (**Project
configuration**, `skills/flow-contracts/project-configuration.md`); 4 the push, per **Branch
backup** (`skills/flow-contracts/git-boundaries.md`):

```bash
flow-guard kickoff-worktree <project> <name> [<base>]
```

`<base>` is the creating run's `--base <branch>` (`skills/flow/SKILL.md`), omitted when none was
given: the new branch is then cut from `origin/<base>` instead of `origin/<default-branch>`, and the
base is recorded on it for `flow-guard resolve-base-branch`; `origin/<base>` missing is exit 1.
Right after the add, the script runs `flow-guard ignore-scratch` on the new worktree.

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
`planner-role`, KAN-1067). A `/flow-plan` capture runs this section too, with the differences
**Capturing a new change** (`skills/flow-plan/SKILL.md`) states. Every planner dispatch is fresh and one-shot, on `subagent_type:
flow-planner-<effort>` with `model` and effort from the **Allowed pairs** row `planner`
(`skills/flow-contracts/model-policy.md`) — its first pair unless an open experiment overrides
it; a CORE change (its Jira issue carries an `**Agents (CORE):**` section) takes `flow-guard plan-agent-pair
<name> <n> planner core` instead, `fable xhigh`, and its prompt carries the line `CORE change`. Its prompt carries `Role: planner`, `Key: plan-<n>`, the two agent-baseline sentences, the
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
   `-key explore-<n>|design-<n>`, each pair from `flow-guard plan-agent-pair <name> <n> explorer|plan-designer`, `<n>` continuing
   the run's count across rounds — the parent names the next free `<n>` in each round's prompt, since
   a reused key restarts the earlier round's row), writing or
   updating the prototype file a visible change needs and never publishing it, and writes
   `round-<r>.md`: the prototype file's absolute path when the round wrote or changed it, `## Summary` — at
   most ten lines, the decisions the round settled and what is still open, which is the settled
   prose the convergence confirm shows — the design draft, then `## Questions` — each question with two to
   four options, one marked recommended where one honestly is, the convergence confirm last when
   nothing else is open.
   It asks the operator nothing.
2. **Relay.** The parent first publishes the prototype
   (**The checklist**, `skills/flow/brainstorm-planner.md`). The parent prints the round's `## Summary` verbatim and
   the round file's absolute path — the design draft stays in the file — with the
   prototype's Artifact URL beside them whenever a prototype exists, asks its questions verbatim, four to
   an **AskUserQuestion** call and each question headed `[architectural]` in a call of its own, in
   round-file order — the URL appended to the text of the question that depends on the
   prototype, or else of the convergence confirm, since the question dialog may hide the text
   above it — and writes the answers verbatim to `answers-<r>.md`; it never answers
   a question itself. Every later planner dispatch's prompt, the final one included, carries the
   prototype's latest Artifact URL. An explicit approve of the convergence confirm closes brainstorming —
   `flow.brainstorm` end, `flow.design-approval` begin and end, as **Convergence**
   (`skills/flow/brainstorm-planner.md`) marks them; any other answer dispatches round `<r+1>`. An
   operator message mid-planning is written to the next `answers-<r>.md` and relayed the same way.
3. **Artifacts and plan.** Two final planner dispatches, under the run's session token:
   - the first runs section **C** of `skills/flow/brainstorm-planner.md` — its <!-- refs-guard:allow -->
     `flow.create-artifacts` marks, `spectre new` and the three artifacts from the approved design,
     read from every file under `<abs-worktree>/.superpowers/sdd/planner/`, where the approved
     design lives — and returns `ARTIFACTS-WRITTEN`; the parent then runs **Plan review** #1 below;
   - a fresh one then runs section **D** up to, not including, **Decide** — its
     `flow.writing-plans` begin mark, writing-plans, `spectre validate`, `flow-guard
     check-plan-shape`, the configured guards and `flow tasks count` — and returns `PLAN-WRITTEN`.

   A dispatch returning the guard output it could not clear is redispatched fresh with that output
   until it returns its word; a redispatch skips each mark already made and resumes at the guard
   that failed. Either dispatch returns
   `QUESTIONS` instead the moment writing the artifacts or the plan surfaces an ambiguity (**The
   checklist**, `skills/flow/brainstorm-planner.md`), the next `round-<r>.md` holding it: the parent
   relays that round as step 2 does, and a fresh final dispatch — never a design round — takes its
   answers, skipping each mark already made and revising the design, the artifacts and the plan
   they change.

The parent then closes `flow.writing-plans`, runs **Plan review** #2 below unless the planner's
return skips it, and runs **Decide**, the record sequence and the gate, as the rest of this file
states.

Run each `begin` immediately before its Agent call and its `end` once that agent returns — never
both after it returns, which records a window of zero seconds (kan-1034: both explorer rows read
0s).

No `fork` from kickoff to the plan gate: a fork copies the whole conversation and doubles the
token spend.

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
re-decision through the sequence above. On **Yes**, take the plan snapshot (**Plan snapshots**,
below), then **end the run** with
the `STARTED` handoff (**The `STARTED` handoff block**, below),
`/clear` above `/flow <name>`, on a creating and a resumed run alike; the next run finds
the plan ready (**Resuming at `STARTED`**, `skills/flow/resume.md`) and implements on a fresh context.

### Plan review

**A fresh reviewer that has seen nothing but the documents reads the plan, twice**
(decisions `plan-review-after-artifacts`, `two-plan-reviews`, KAN-1104), in `/flow`'s creating
run and a `/flow-plan` capture:
- `plan-review-1` once the first final planner returned `ARTIFACTS-WRITTEN` and
  `flow.create-artifacts` has ended;
- `plan-review-2` once the second returned `PLAN-WRITTEN` and `flow.writing-plans` has ended.

A return whose second line is `Plan review: skip — <reason>` skips both: print that line,
mark no `flow.plan-review` and go on. A fix run's re-plan, the self-review and
in-run fix planners and `/flow-fast` never run it.

```bash
flow stage begin -command '/flow' -stage flow.plan-review -harness <harness> -session-token mf-<literal-token> <name>
flow record dispatch begin -change <name> -role reviewer -model opus -effort low -key plan-review-<n> -session-token mf-<literal-token>
flow record dispatch end -change <name> -key plan-review-<n> -session-token mf-<literal-token> -outcome completed
flow stage end -command '/flow' -stage flow.plan-review -outcome completed <name>
```

`flow.plan-review` is begun and ended around each review, and steps 1–4 apply to each; every
dispatch, planner redispatch and question round of a review sits inside its stage.

1. **Dispatch.** One Agent call, `subagent_type: flow-low`, `model: opus` — the **Allowed pairs**
   row `design-reviewer` (`skills/flow-contracts/model-policy.md`) — its `description` opening
   with `plan-review-<n>`, `<n>` counting this run's plan-review dispatches from 1. The prompt is
   the lines `Role: design-reviewer` and `Key: plan-review-<n>`, the two agent-baseline
   sentences, the report-tersely line, the paragraph below, and six absolute paths:
   `<skill-dir>/plan-reviewer-prompt.md` as the file to read in full first,
   `<skill-dir>/plan-review-rubric.md`, the change's `proposal.md`, `design.md` and `tasks.md`
   under `<changeRoot>`, the repository root `<abs-worktree>`, and the report
   `<abs-worktree>/.superpowers/sdd/plan-review-<n>.md` — `<skill-dir>` the directory this file
   was read from. It carries no ticket text, no Jira key, no round or answers file and no
   summary of the conversation.

   > **MODEL HANDSHAKE:** the first line of your first reply is `Model: <the model named in your own
   > system prompt>` and nothing else on that line. Answer it before any tool call.

2. **`PLAN-REVIEW-CLEAN`** — end `flow.plan-review` and go on.
3. **`PLAN-REVIEW-FINDINGS`** — redispatch a fresh final planner as for guard output it could
   not clear, its prompt naming the report's absolute path; it triages the report per **Plan
   review** (`skills/flow/brainstorm-planner.md`) and makes no stage mark. Its `QUESTIONS` return
   is relayed as step 2 of the section above relays a round, and a fresh final dispatch takes
   the answers. Its `ARTIFACTS-WRITTEN` or `PLAN-WRITTEN` return ends `flow.plan-review`;
   nothing is reviewed a third time.
4. **A failed reviewer.** A dispatch that errors, or returns neither verdict line, is
   redispatched once under the next `<n>`. Failing again, the plan review is `not run`, never
   clean: end the stage with `-outcome failed -reason 'plan review not run'` and go on.

A `/flow-plan` capture marks the stage under `-command '/flow-plan'`.

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
the handoff is the one place the settled plan is guaranteed to be read (kan-941 printed only
`Next:`).

## Plan snapshots

`<project>/spectre/changes/` is gitignored, so no commit records the plan as it stood at a gate. A plan
snapshot does: on the plan gate's **Yes**, at `/flow-plan`'s capture and after a fix run appends
tasks, the run copies `tasks.md`, `proposal.md` and `design.md` of `<project>/spectre/changes/<name>/`
in the canonical worktree into the next numbered directory
`<abs-worktree>/.superpowers/sdd/<name>/plan-snapshots/<k>/`
(`<k>` = 1 + the number of directories already there), never committed, with exactly this command:

```bash
d=<worktree>/.superpowers/sdd/<name>/plan-snapshots
k=$(( $(find "$d" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l) + 1 ))
mkdir -p "$d/$k" && cp <worktree>/spectre/changes/<name>/{tasks,proposal,design}.md "$d/$k/"
```

A non-zero exit stops the run before anything else happens. The snapshots have one reader:
flow-eval's `case add impl` reads, for task n, the earliest snapshot holding task n, so a case sees
the plan as it stood before any review fix wrote a Correction. No snapshot is ever rewritten.
