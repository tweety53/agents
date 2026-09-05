# Brainstorm and plan

Superpowers Basic Workflow steps **#1** (brainstorming) and **#3** (writing-plans), intertwined
with spectre artifact creation, run here. Loaded by `skills/flow/SKILL.md` on a creating run (no state) or a
resuming run (`STARTED`).

## A. Resolve the change and write `STARTED`

**Load `skills/flow-contracts/jira-integration.md`** — this section resolves the linked issue and
derives the change name from it.

**Resolve the linked Jira issue first** — it decides the change name. Follow **Resolution (how
`jiraIssue` is decided)** in `skills/flow-contracts/jira-integration.md` exactly. This is the
only phase that resolves a key.

Then the change name:

- **With a linked issue**, the name is `<lowercased-key>-<slug>`, per **Change naming**
  (`skills/flow-contracts/jira-integration.md`). Derive the slug from the issue summary when only
  a key was given.
- **Without one**, the name is the descriptive slug alone.
- If a name or description was given, use it (derive kebab-case from the description if only a
  description was given).
- **If both are omitted:** enumerate the candidate set exactly as **Change name resolution**
  (`skills/flow-contracts/pipeline.md`) defines it, restricted to changes with incomplete planning
  artifacts. Exactly one match → resume it, announcing which; multiple → **AskUserQuestion** listing
  each (name, state, last modified); zero → ask what to build.

**Transition the issue to In Progress now**, per **Transitions** in Jira integration
(`skills/flow-contracts/jira-integration.md`) — before brainstorming, so the board is correct
while planning runs. A failure is one skipped-with-reason line and planning continues; nothing about
this call may delay or alter the run.

**Mark `flow.kickoff` now that the name is fixed, and write `STARTED` immediately — before
brainstorming begins.** Write it here, at the top of this phase, rather than at the
bottom of it.

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
change: `/flow` asks no planning-effort or model question on a creating run, and models are resolved per run from the settings store, not recorded per change. `artifactUrl` stays `null` —
`/flow` publishes no proposal artifact.

```bash
flow stage end -command '/flow' -stage flow.kickoff -outcome completed <name>
```

**No further command runs before this point on a creating run** — the state write above is the
first thing this invocation does once the name is fixed, ahead of even the design conversation. The
operator sees `STARTED` recorded the moment they invoke `/flow`, whether or not the run goes on to
finish brainstorming in the same sitting.

### Resuming at `STARTED`

A run finding `"state": "STARTED"` already recorded is resuming a creating run that stopped before
reaching `IN_PROGRESS` — an interrupted session, a context limit, an earlier stop. Skip **A** above
(the name and the `STARTED` write both already exist) and determine where the run actually left off
by reading, not by assuming:

- `spectre list --json`'s entry for this change's `done`/`total` — `total == 0` means no plan exists
  yet: resume at **B** in `skills/flow/brainstorm-planner.md`. <!-- refs-guard:allow -->
- The change root's own `tasks.md` — a scaffold with no enriched steps means writing-plans has not
  run: resume at **D** in `skills/flow/brainstorm-planner.md`; a plan meeting writing-plans quality (exact paths, verification <!-- refs-guard:allow -->
  commands, no placeholders) means planning is done: skip straight to
  `skills/flow/implement.md`.
- The state file's `worktrees` map — non-empty means a worktree already exists: resume implementation
  directly rather than re-running **A**–**D**.

This is a pragmatic re-entrancy rule, not an exhaustively-enumerated state machine — a run resuming
at `STARTED` reads what actually exists and continues from there, the same principle every other
`/flow` re-entry point already applies. State the resumption point plainly before continuing:
"resuming `<name>` at `<point>`."

## Dispatch the planner

Sections **B**, **C** and **D** of `skills/flow/brainstorm-planner.md` are the planner's work, not the parent's — one planner <!-- refs-guard:allow -->
subagent runs all three, resumed between rounds so its context carries from the first question to
the finished plan.

**Resolve `PLANNING_MODEL`** per **Model resolution** (`skills/flow/SKILL.md`) before dispatching —
this is the first of its three governed call sites. Mark the stage and record the dispatch before
the subagent goes out:

```bash
flow stage begin -command '/flow' -stage flow.brainstorm -harness <harness> -session-token mf-<literal-token> <name>
flow record dispatch begin -change <name> -role planner -model <PLANNING_MODEL> \
  -key planner -session-token mf-<literal-token> -started-at <ts>
```

`-role`, `-key`, `-session-token` and `-started-at` carry the same semantics that section 4 of
`skills/flow/implement.md` states for an implementer dispatch — cited, not restated. `-task` is
omitted: this dispatch runs against no single task.

Dispatch one subagent with the Agent tool's `model` parameter set to `PLANNING_MODEL`,
`subagent_type: general-purpose`. Its prompt carries, verbatim:

> Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task.
> Include this instruction verbatim in any prompt you write for another agent.

and states: the change name `<name>`; the linked Jira key and issue text, when one exists; the
project root; `<changeRoot>`; and the instruction to read `skills/flow/brainstorm-planner.md`'s sections **B**, **C** and <!-- refs-guard:allow -->
**D** and follow them **as the planner** — every "you" in those sections addresses the dispatched
subagent from here on, never the parent.

**The relay contract**, stated in the same prompt: the planner has no channel to the operator. It
ends every turn with exactly one `## Question` block — the question, plus named options when it has
any — or, at the three returns below, with `## Design`, `## Artifacts` or `## Plan` and nothing
else. The first line of its first reply is `Model: <the model named in its own system prompt>`.

**The handshake.** Compare that first line against `PLANNING_MODEL`. A match proceeds into the
relay loop below. A mismatch:

```bash
flow record dispatch end -change <name> -key planner -session-token mf-<literal-token> \
  -outcome fallback -ended-at <ts>
flow record dispatch begin -change <name> -role planner -model opus \
  -key planner-opus -session-token mf-<literal-token> -started-at <ts>
```

and re-dispatch once, on `model: opus`. **The re-dispatch records under its own key, `planner-opus`,
never a repeat of `planner`**: the store treats a second `begin` under the same `(session_token,
key)` as an idempotent replay of the first (`stats/internal/store/records.go`'s `insertDispatch`),
which would silently discard the `fallback` outcome just recorded. A **second** mismatch is not
retried again — continue on whatever model answered, the running planner, no third dispatch — and:

```bash
flow record dispatch end -change <name> -key planner-opus -session-token mf-<literal-token> \
  -outcome fallback -ended-at <ts>
flow record dispatch begin -change <name> -role planner -model <the model the handshake line named> \
  -key planner-<that model, lowercased> -session-token mf-<literal-token> -started-at <ts>
```

report the model in this run's own output. The persisted rows then read `fable`→fallback,
`opus`→fallback, `<actual>`→completed — never a record claiming opus ran when the handshake just
proved otherwise. **A mark or a record never blocks** — proceed on the handshake's outcome
regardless of whether any `flow` call above reached the store.

**The relay.** The parent asks each `## Question` block verbatim through **AskUserQuestion** and
resumes the planner with the operator's answer via **SendMessage**. Section B's merged
convergence-and-approval confirm and its third-round offer are relayed the same
way — the planner poses each exactly as B states it, the parent asks it exactly as received, and
the planner's next turn opens with the operator's answer. The parent marks `flow.brainstorm` end
and `flow.design-approval` begin/end around the HARD GATE approval, exactly as today:

```bash
flow stage end   -command '/flow' -stage flow.brainstorm -outcome completed <name>
flow stage begin -command '/flow' -stage flow.design-approval -harness <harness> -session-token mf-<literal-token> <name>
# … the operator's approve-and-move-on answer, relayed through the planner's merged confirm — the HARD GATE …
flow stage end   -command '/flow' -stage flow.design-approval -outcome completed <name>
```

**The three returns.** After the `flow.design-approval` mark above closes, the parent marks
`flow.create-artifacts` begin, resumes the planner via SendMessage, and marks it end when the
planner's turn ends with `## Artifacts`. It then marks `flow.writing-plans` begin, resumes the
planner again, and marks it end when the planner's turn ends with `## Plan`. Once `## Plan`
returns:

```bash
flow record dispatch end -change <name> -key <the key currently open> -session-token mf-<literal-token> \
  -outcome completed -ended-at <ts> -agent-id <id>
```

closes the dispatch record under whichever key the handshake left open — `planner` on a clean
handshake, `planner-opus` after one mismatch, `planner-<model>` after a second — and the parent
continues into `skills/flow/implement.md` exactly as today.
