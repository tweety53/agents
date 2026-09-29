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
`<project>/.worktrees/<name>`, including the design spec, `spectre new`, the three artifacts, the
plan and the decision JSON.

1. `check-worktree-location.sh <project>` — exit 1 or 2 stops the run with the guard's own lines.
2. `git check-ignore -q .worktrees` from the project root. Where it exits non-zero, append
   `<project>/.worktrees/` to `<project>/.git/info/exclude` — never a commit on any branch.
3. `git -C <project> fetch origin`, then — when `git -C <project> rev-parse -q --verify
   origin/spectre/<name>` succeeds, the branch already exists on the remote (a `/flow-plan`
   capture pushed it, or an earlier run's worktree was removed) — `git worktree add
   <project>/.worktrees/<name> spectre/<name>`, a local branch tracking that remote one; otherwise
   `git worktree add <project>/.worktrees/<name> -b spectre/<name> origin/<default-branch>` — the
   default branch by name, never HEAD: the main checkout may be on any branch and is never moved.
   **Persist the worktree into the state record before this step returns**, by the read-merge-write
   **2. Isolate the workspace** (`skills/flow/implement.md`) applies to each additional worktree:
   merge the entry `"<abs-worktree>": "<merge-base>"` into the record's `worktrees` map and write
   it back, where `<merge-base>` is the sha `git -C <worktree> rev-parse HEAD` prints immediately
   after the add — the commit the worktree starts from in either case above.
4. `project-get.sh <worktree> "worktree setup"`. Exit 0: run every printed line from the worktree
   root, in order, in the foreground — the printed body can carry fence markers and trailing prose
   outside the fence (as `<project>/.flow/project.md`'s `## worktree setup` section does); run only
   the fenced command lines, not those. Exit 1: the project declares no `## worktree setup`; say so
   and continue. Exit 2: stop the run, relaying the script's own line. **A command's non-zero exit
   ends your turn** naming the command and its output. The key is canonical in
   **Project configuration** (`skills/flow-contracts/project-configuration.md`).
5. `git -C <worktree> push -u origin spectre/<name>` — the branch exists on the remote from its
   first minute, per **Branch backup** (`skills/flow-contracts/git-boundaries.md`).

```bash
flow stage end -command '/flow' -stage flow.kickoff -outcome completed <name>
```

## Run brainstorming and planning directly

```bash
flow stage begin -command '/flow' -stage flow.brainstorm -harness <harness> -session-token mf-<literal-token> <name>
```

Read `skills/flow/brainstorm-planner.md`'s sections **B**, **C** and **D** and follow them <!-- refs-guard:allow -->
directly — `REVIEWERS` (**Model resolution**, `skills/flow/SKILL.md`) is
already in scope from this run's own earlier resolution.

**Prose preceding a question is shown too, not dropped.** When the checklist carries a summary
before a question — most concretely the convergence confirm's "state what you believe settled"
paragraph — show that prose to the operator as ordinary text before the **AskUserQuestion** call,
not only the bare question and options.

Once the Decide step finishes, print the `## Decision`
block verbatim — the two-table shape (input side, then decision side) **Decide** (`skills/flow/brainstorm-planner.md`) prints — then run the record
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

**A value the state file does not carry is reported as missing, not dropped.**

```text
## Proposal ready — review required

**Change:** <name>
**Artifact:** <artifactUrl, or "missing">
**Recorded:** <N> decisions · <N> open questions · effort <level, or "not recorded — planned at default"> · model <models.default, or "not recorded">
**Jira:** <issue key and the transition made, or "none linked", or a skipped-with-reason line>
**Jira description (pre-edit):** <the text as it stood before the write, verbatim in a fenced block>

Open in IntelliJ:
open -na "IntelliJ IDEA" --args "<absolute apply-worktree path>"

<what the operator does next>

Next:
/clear
/flow <name>
```

**Both the decisions count and the open-questions count render `none` when zero — never `0` — by
the missing-rather-than-dropped rule above.**
