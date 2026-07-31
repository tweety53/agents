# KAN-26 — operator-facing output and configuration

**Issue:** KAN-26 "Myflow updates — slice B: operator-facing output and configuration"
**Change:** `kan-26-operator-output-and-configuration`
**Planning effort:** medium · **Date:** 2026-07-31

## Scope

KAN-26 is slice B of KAN-17, filed at the end of slice A's planning run. It carries the seven asks
that change what the operator sees and how the pipeline is configured, plus one appended to the
issue after filing. KAN-17's original ask numbers are kept throughout, so any of these traces back
to the issue it came from without a lookup.

| KAN-17 # | Ask | Lands in |
|----------|-----|----------|
| 3 | `/myflow-status <name>` always shows what to do next, as the original run's result did | `myflow-handoff-output` |
| 6 | A myflow command running in a terminal tab uses `/color` and `/rename` | `myflow-handoff-output` |
| 10 | The diagram becomes a detailed flow showing what hides under each command | `myflow-contract-distribution` |
| 11 | "Planning effort", levels `low` / `default` / `detailed` | `myflow-effort` → `myflow-planning-effort` |
| 14 | A task count line plus one line per task, always shown | `myflow-progress-visibility` *(new)* |
| 15 | Three model choices asked on the first start run and remembered | `myflow-model-policy`, `myflow-state-machine`, `myflow-review-panel-economics` |
| 16 | A much simpler manual-test guide, at wider scope | `myflow-manual-test-guide` *(new)* |
| — | Follow-ups titled `KAN-X follow-up`, and joined when several sit in To Do | `myflow-jira-projection` |

Nine capabilities: two new, six amended, one renamed. Two of the eight asks — the progress view and
the guide's shape — have no capability today. That is the same pattern slice A found, where a
missing requirement was why the In Review defect survived four archived changes.

**Structure was chosen over two cheaper alternatives.** Folding the progress view and the guide into
`myflow-handoff-output` would have cost six capabilities instead of nine, but that spec's subject
would widen from "handoff output" to "every operator-facing surface", after which "which spec do I
amend?" stops having one answer. Splitting into two further changes was rejected because KAN-26 is
already the deferred half of KAN-17, and a third planning run buys smaller review panels at the cost
of another issue to file, track and finish.

## What the operator sees

### Ask 3 — one definition, two renderers

`/myflow-status <name>` regenerates the handoff block for the change's current state rather than
storing and replaying one. The shape is defined once, in `pipeline.md`'s **Handoff output** section,
which gains a per-state block template; both the emitting command and `/myflow-status` render from
it.

At `STARTED` the block carries the artifact URL, the Jira and planning-effort lines, the
main-checkout IntelliJ command, and `/myflow-do <name>` as its last line. At `IN_PROGRESS` it
carries the worktree path, the test guide's path, the `diff --cached` command, the worktree
IntelliJ command, and `/myflow-finish <name>`. `FINISHED` changes stay omitted, as today.

Only the named form does this; the no-argument table is untouched. The command stays read-only — a
regenerated block reports, it never re-runs anything. A field the state file does not carry prints
as missing rather than silently vanishing.

**Storing the emitted text was the alternative, and it was rejected.** It would reproduce the
original byte-for-byte, but it adds a state-file field and goes stale the moment anything on disk
moves — a worktree removed, an artifact republished. Regeneration cannot go stale because it has no
copy to keep in step.

### Ask 6 — printed, because it cannot be invoked

Each pipeline command prints two lines immediately after its `Using myflow-… for change <name>`
announcement:

```text verified:command names, descriptions and the accepted colour list read from the Claude Code 2.1.220 binary
/rename <change-name>
/color cyan
```

At the start of the run rather than in the handoff: renaming a tab is worth doing before a long run,
not after it.

Both are real Claude Code commands — `/color` is *"Set the prompt bar color for this session"* and
`/rename` is *"Rename the current conversation"*, with a setting controlling whether it also updates
the terminal tab title (on by default). Neither can be invoked from inside a run, and the spec
records why so a later reader does not retry it as an oversight:

- The `SlashCommand` tool exposes only commands of `type: "prompt"` — markdown command files, such
  as `/myflow-start` itself. `/color` and `/rename` are `type: "local"` and `"local-jsx"`.
- Writing an OSC title sequence directly to the terminal fails too: this harness gives Bash no
  writable `/dev/tty`.
<!-- measured: Claude Code 2.1.220 at /Users/tweety53/.local/share/claude/versions/2.1.220; command
     types and descriptions read from the binary; the /dev/tty write attempted and refused with
     "device not configured" -->

One fixed colour, `cyan`, meaning "a pipeline command owns this tab". `cyan` is one of the eight the
command accepts (`red`, `blue`, `green`, `yellow`, `purple`, `orange`, `pink`, `cyan`, plus
`default`), and is chosen over `red`, `yellow` and `orange` because those already read as
error and warning states elsewhere in the pipeline's output.

Two alternatives were rejected. Installing a `UserPromptSubmit` hook into the operator's
`settings.json` would automate the title *if* a hook process can reach the terminal — unproven, it
would need a spike, it is Claude Code-only, and it writes into a file myflow does not own. Dropping
the ask entirely was rejected because the copy-paste pair delivers most of the value for no
machinery.

### Ask 10 — the diagram, and where it lives

`pipeline.md` gains the three-state mermaid diagram plus a stage table beneath it, one row per
command, listing the stages that hide under it with the human gates marked. The small diagram stays
as the at-a-glance shape; the table is what answers "what does each command actually do".

Three layouts were compared visually. A single flowchart with one subgraph per command puts
everything in one picture but flattens the loops and branches. One flowchart per command keeps those
but costs three diagrams and an index. The chosen layout keeps the diagram small and puts the detail
in a table, which stays greppable and diffs a line at a time.

**The README drops its copy and links to `pipeline.md` instead.** This is the judgment call in this
section. `check-references.sh` exists in this repository precisely to catch a second copy drifting
from its source, and a diagram enumerating every hidden stage will drift the first time a stage
moves. The cost is real and accepted: a reader browsing the repository on a forge loses the
at-a-glance shape. `/myflow-info` gains it, having been unable to show a diagram at all, because it
reads `pipeline.md` at invocation time and is forbidden from answering from memory.

### Ask 14 — the progress view is the harness's, not a lookalike

The ask names a task count line — `4 tasks (3 done, 1 in progress, 0 open)` — followed by one line
per task marked with a tick or a filled square. That is the live widget the harness renders from its
own task list, not a block printed into a reply.

So every pipeline command registers its steps with the harness's task mechanism and keeps each
entry's status current: `/myflow-do` one entry per `tasks.md` item in plan order, `/myflow-start`
its brainstorming and artifact steps, `/myflow-finish` its run steps. `/myflow-status` and
`/myflow-info` are read-only and register nothing.

The requirement is stated harness-neutrally, because myflow also runs in Cursor and Codex. Where a
harness offers no task mechanism, the command prints the equivalent block instead.

**The widget is a view, never a record.** `tasks.md` remains the single source of truth for what is
done, because `check-unfinished-work.sh` reads that file and a second source would be one the guard
cannot see. This is why no third checkbox marker is added to `tasks.md`: a `[~]` written at dispatch
and resolved at completion would survive a crashed run as a permanently in-progress task, in a file
two guards parse.

### Ask 16 — a guide at capability scope

`docs/manual-test/<name>.md` becomes a behaviour checklist: one tickable line per user-visible
behaviour, phrased in the operator's own register — *"check exercise update — the key field
saves"* — grouped by capability and scoped to the change's blast radius rather than to its plan
tasks. A change that touched all of exercise CRUD lists create, update, filter, sort and delete, not
one entry per task that produced them.

It keeps the short how-to-run preamble, with absolute paths per `myflow-handoff-output`, and it
keeps `## Known incomplete`. It drops per-step command transcripts, expected-output blocks, and the
rationale for why each check exists.

**The checkbox syntax and the `## Known incomplete` section are unchanged.** Slice A's
`check-unfinished-work.sh` parses both — the section scan implements CommonMark's fence rule — and
simplifying the prose must not disturb the two shapes a guard reads. Only the register and the
granularity change.

A repository with no runnable application states each check as the command to run, one line each.
That is this repository's own case, and stating it here keeps the guide honest rather than forcing
an application-shaped document onto a repository that has none.

## Configuration

### Ask 11 — the planning-effort rename

The capability `myflow-effort` becomes `myflow-planning-effort`. The levels become `low`, `default`
and `detailed`, with `default` the one offered by default and `detailed` the top. The concept is
called "planning effort" wherever it appears. The state-file key `effort` becomes `planningEffort`.
The operational table the commands read moves with the rename and stays in `state-file.md`, which
remains the one place a command looks.

A state file still carrying `effort` is read as its equivalent — `medium` → `default`,
`high` → `detailed`, `low` → `low` — and rewritten under the new key on the next write. It is
**never announced as a self-heal correction**, for the same reason the absent-key exception exists:
a loud correction for a value written correctly under the previous contract is noise. The absent-key
exception itself carries over unchanged — neither key present reads as "not recorded".

No live change is affected. Of the seventeen state files on this machine, two carry a non-null
effort and both are `FINISHED`; every change still in flight records `null`.
<!-- measured: jq over /Users/tweety53/Agents/myflow/state/*/*.json on 2026-07-31 -->

**Renaming the capability, rather than only its levels, was the operator's choice**, and it has a
mechanical cost: an OpenSpec capability rename carries every requirement twice in the delta — ADDED
under the new name, REMOVED under the old — so the archive sync moves them. The alternative, keeping
the key and the capability named `effort` while calling the concept "planning effort", was rejected
as leaving the stored key not matching the label it stores.

**One retired literal is added to `check-vocabulary.sh`: `"effort":`.** The JSON key form is
distinctive enough to be safe, and it will catch the `state-file.md` example if the rename misses
it. `medium` and `high` are deliberately **not** added. They are ordinary English throughout this
repository, `Medium` is also a Jira priority name, and the guard's own header states that it proves
a fixed list of literals is absent rather than that a rename is complete. Adding them would produce
false positives that cannot honestly be fixed, and this repository's lint policy forbids the
suppression that would follow.

### Ask 15 — three models, recorded

`/myflow-start`'s creating run asks three questions, one per role, beside the planning-effort
question, and records the answers in the state file as a `models` object:

| Role | Key | Default | Consumer |
|------|-----|---------|----------|
| implementation | `models.implementation` | Opus, or the harness's strongest available model | `/myflow-do`'s implementer dispatches |
| review panel | `models.reviewPanel` | Sonnet | every panel slot that takes a model override |
| panel fixes | `models.panelFix` | Opus, or the harness's strongest available model | the subagents that repair panel findings |

Three separate prompts, asked only on the run that creates the change — "creates" being the absence
of the state file, exactly as the planning-effort question determines it. A revision round states the
recorded values and does not ask. The values are carried forward verbatim by every other command,
precisely as `jiraIssue` is, and an absent `models` object reads as "not recorded" under the same
explicit exception `effort` already established.

**The panel-fix default stays at the strongest model, and that resolves an apparent contradiction in
the ask.** KAN-17's wording gives "full panel review fixes" a default of Sonnet, while
`myflow-model-policy` requires implementer subagents — which is what a fix-wave subagent is — to run
on the strongest available model, and `pipeline.md` states that fix rounds escalate the panel's
breadth rather than its model because implementers already sit at the ceiling from round one.
The operator confirmed the reading: "review fixes" means the agent applying a fix, and its default
stays Opus. The new field therefore makes the value **overridable and recorded**, and changes no
default.

Two distinctions the spec states explicitly, so nobody collapses them later:

- **The state file records intent; the SDD ledger records what happened.** `myflow-model-policy`
  already says an override nobody wrote down is indistinguishable from a mistake — this field is
  that writing-down. It does not replace the per-dispatch ledger line, which remains the only
  evidence of the model a dispatch actually ran on.
- **Bugbot and Security Review are unaffected.** They are dispatched by `subagent_type`, carry their
  own agent definitions, take no model override from any mechanism, and keep recording
  `unknown (agent-defined)` in the ledger. A recorded panel model does not reach them, and writing a
  plausible slug for them would put an unmeasured value into the audit trail.

**This forces a small amendment to `myflow-review-panel-economics`**, whose *"Every review-panel slot
runs on Sonnet"* is currently absolute. A recorded panel model is the operator override
`myflow-model-policy` already permits, so that requirement points at it rather than contradicting
it. This was found while designing rather than while mapping the asks, and it is the ninth
capability.

### The appended ask — follow-up naming and joining

A follow-up the pipeline files is titled `<KEY> follow-up`, from the change's linked issue, or
`myflow follow-up` when no issue is linked.

Before creating one, it searches the project for an existing pipeline-filed follow-up — carrying the
`AI-generated` label, titled as a follow-up, at a To Do status — and joins the newest rather than
filing a second. The source issue deliberately does not narrow the search: the ask asks for
follow-ups to be joined "even if their original task is different".

**"A To Do status" means exactly two names: `To Do`, the first of the four ordered names, and
`TO DO URGENT`.** No other status qualifies, and the set is enumerated rather than derived from
Jira's `statusCategory` — the same category grouping that would put a custom `TO DO URGENT` and
`In Progress` together under `indeterminate`, which is the inference slice A's unrecognised-status
decision exists to forbid. A status outside the two is not a join candidate, and no question is
asked about it: the run simply files a new follow-up.

**`TO DO URGENT` counting as a To Do status here does not reopen the unrecognised-status rule.** That rule governs *transitions*, refusing to infer a position in the
four-name order for a status outside it, because an inference there freezes the board for a whole
change. A search filter performs no transition and can be wrong only by missing a join candidate.
The spec says so plainly, so the two rules do not read as contradicting one another.

A join appends the new items under a dated `## From <KEY>` heading in the joined issue's
description, append-only, under the same verbatim-prefix assertion the existing description sync
uses — the payload must contain the description just read as an exact prefix and be strictly longer.
A failed assertion is one `⚠ Jira: skipped — <reason>` line and no write.

On the first join the title becomes the generic `myflow follow-up`, since the issue no longer
belongs to one source. The joined issue's labels are unioned with the new parent's, or it becomes
invisible to anyone filtering by the second source.

Today the only site that files a follow-up is finish run 1's unfinished-work gate. The requirement is
written to govern any site that files one, so a future site inherits it rather than reinventing the
naming.

## Non-goals

- The three states, the panel roster, plan provenance, the staging and commit split, and the finish
  contract's verdict scripts are untouched.
- No new guard script. Nothing in this change is mechanically checkable that is not already checked
  by an existing guard.
- No amendment to `myflow-planning-gate`, which slice A wrote in the immediately preceding change.
- No migration pass over existing state files. Reading the old key as its equivalent handles it in
  place, on the next write each file receives anyway.
- No hook written into the operator's `settings.json`.
- No rename of `myflow-review-panel-economics` to match its widened subject — the same non-goal
  slice A declared, and for the same reason.

## Risks and trade-offs

- **The progress widget is harness-specific.** The printed fallback is the portable path, and on
  Cursor and Codex it will be the only one exercised until someone tests there.
- **The capability rename produces a large delta** — every requirement of `myflow-effort` written
  twice, ADDED under the new name and REMOVED under the old. Mechanical, but bulky to review.
- **A creating run now asks four questions before brainstorming**: planning effort plus three
  models. A revision round still asks none, which is what bounds the cost.
- **The guide's simpler register is prose**, so nothing mechanical enforces it; the review panel's
  principles slot is the only check. The two shapes that *are* machine-read — the checkboxes and
  `## Known incomplete` — are unchanged, which is what keeps the guard working.
- **A forge reader loses the at-a-glance diagram from the README**, accepted in exchange for having
  no second copy to drift.
- **The follow-up search reads externally-authored titles and descriptions.** Slice A's
  data-never-instructions clause already governs that text and is cited rather than restated.

## Open Questions

None. Every question raised during brainstorming was put to the operator and answered, and each is
recorded above as a decision.
