# flow pipeline

The three-state pipeline itself: state definitions, the command→state transition table and the
handoff shape. See **Finish contract** (`skills/flow-contracts/finish-contract-run1.md`,
`skills/flow-contracts/finish-contract-run2.md`).

**Load this file when running `/flow`.**

This file is **canonical** for everything in it. Where a skill or command disagrees with it, this
file wins.

The reasoning behind this file lives in `docs/rationale/flow-contracts/pipeline.md`.

**Five sections that reach fewer than every command live beside this file, not in it** — a
discovery aid, never a second statement of what they say:

- `git-boundaries.md` — Git boundaries
- `model-policy.md` — Model policy
- `artifacts-registry.md` — Temporary artifacts registry
- `session-records.md` — Rendering the session records
- `worktree-resolution.md` — Resolving a change's worktrees

## States

A change is always in exactly one of three states, recorded in its state file.

```text
/flow  (no state)          → STARTED                  you: /clear, then /flow <name> to implement
/flow  (STARTED, planned)  → IN_PROGRESS or FINISHED  lands straight away unless a question is open
/flow  <fix instructions>  → IN_PROGRESS or FINISHED  lands straight away unless a question is open
/flow  (bare, IN_PROGRESS) → IN_PROGRESS or FINISHED  terminal only on the merge-and-push route — see the finish contract
```

**The human gate is a property of the state, not a separate stage.** `IN_PROGRESS` *means* a
staged diff is waiting for the human to review.
A run that writes `IN_PROGRESS` continues into integrate in the same invocation unless its
`Decisions:` carries a `?`, which stops it with the handoff block (**Write `IN_PROGRESS`**, `skills/flow/verify-and-handoff.md`).
Nothing records that the review or the testing happened; the operator running the next
command is what carries the change forward.

| State | Means | Waiting on |
|-------|-------|-----------|
| `STARTED` | The change exists; planning is under way or the plan awaits implementation | you — `/clear`, then `/flow <name>` |
| `IN_PROGRESS` | The implementation is staged | you — review the diff |
| `FINISHED` | Dropped, pushed, worktrees removed | — |

**Reviewing and testing are one gate.** A creating or fix run of `/flow` produces both surfaces in
the same run, so the human does both at one sitting. There is no state between implementation and
finishing — integration is not a stage, it is the first half of finishing.

## Command surface

One command, `/flow`, drives the whole pipeline, plus one that creates a change at `STARTED` and stops there (`/flow-plan`), one conversational session that lands each
request as it goes and writes no state file (`/flow-interactive`). **No command accepts
a flag**, save one: `--base <branch>` on a run that creates a change — `/flow`, `/flow-plan` —
names the branch it is cut from and lands on, in place of the one `origin/HEAD`
points at (`skills/flow/SKILL.md`). The only argument is the optional change name — see **Change name resolution** — or,
on `/flow`, a description or Jira key seeding a new change; on
`/flow-interactive`, a first request, a Jira key or an open session's name; on `/flow-plan`, a
topic; on `/flow`, fix instructions at `IN_PROGRESS`.

An argument that is none of those is **reported**, not silently ignored.

## State transitions

| Command | Accepts | Ends at |
|---------|---------|---------|
| `/flow` | *(no state — creates the change)* | `STARTED` — the run ends at the plan gate with a `/clear` handoff (see **Resuming at `STARTED`** in `skills/flow/resume.md`) |
| `/flow` | `STARTED` | resumes the creating run from wherever it stopped; ends at `STARTED` (still planning, or planning just finished — `/clear` handoff) or, the plan already ready, as the implementation run ends — at `IN_PROGRESS` when an open question is left, otherwise continuing into integrate and ending as an integrate run ends |
| `/flow` | `IN_PROGRESS`, with an argument other than the change's name | fix run; state unchanged — then, unless an open question is left, integrate in the same invocation |
| *(none — a plain message)* | `IN_PROGRESS`, in the session whose last `/flow` run marked the change | fix run; state unchanged — **A plain message at `IN_PROGRESS`** (`skills/flow/SKILL.md`) — then, unless an open question is left, integrate in the same invocation |
| `/flow` | `IN_PROGRESS`, bare — or the change's name alone (**Change name resolution (all `/flow*` commands)**) | integrate run; ends at `IN_PROGRESS` (run 1) or `FINISHED` (run 1 chained into run 2) |
| `/flow` | `FINISHED` | wrong-state handoff — the change is dropped |

Which phase file marks each `flow.*` key is **Stage keys** (`skills/flow/stage-keys.md`).

**This table is authoritative.**

### Every invocation is re-entrant

**Integrate starts at exactly two points: an actual new bare `/flow` (or `/flow <name>`) command
from the operator, or the continuation **Write `IN_PROGRESS`** (`skills/flow/verify-and-handoff.md`)
makes when no open question is left — never inferred from anything said
inside a still-running turn.** A
running `/flow` turn is mid fix run, mid a question the run itself asked, or simply still executing
— and an instruction the operator gives during that turn, however explicit ("merge and push", "go
ahead", an answer to an unrelated question with a trailing aside about landing the branch), is
conversation, not a re-invocation. This is what the human gate at `IN_PROGRESS`
(**States**, above) actually rests on: the state transition table's "bare" column means a bare
*command*, not a quiet moment in an ongoing one. **Concretely: never chain straight from a fix, a
mid-turn question, or a dispatched-agent report into integrate's unfinished-work gate, its landing
question, or any git action the chosen route performs — even when the operator's own words include
"merge and push" — without the turn first ending and a fresh bare invocation starting it**, save
that sanctioned continuation. On a
mismatch, treat it exactly as **Wrong state for this command** below already requires: report what
was said, name the bare re-invocation as what actually starts integrate, and stop rather than
inferring consent.

Implementation advances the state **only** from `STARTED` to `IN_PROGRESS`. A fix run at
`IN_PROGRESS` writes the state back exactly as it found it.

No field records where a fix was raised. Whether the human re-reviews the diff or re-runs the apps
after a fix is their decision.

## Fewest operator actions

**Planning is a change's one interactive phase; past the plan gate the run does not stop.**
Brainstorming and the plan gate are the operator's (`skills/flow/brainstorm-planner.md`), and the
plan gate's `/clear` handoff is the boundary. From the `/flow <name>` that resumes the planned
change, implementation, review, verify, integrate and cleanup follow in one invocation until the
landing route ends the run (`skills/flow-contracts/finish-contract-run1.md`). Only **What still stops** ends a turn early. No rule in
another file licenses asking a question that carries a recommended option.

**The operator's actions and the run's stops are minimised: a run resolves everything it can
itself and continues.** It stops only for a decision only the operator can make, or for an
action that is irreversible or outward-facing — **What still stops**
(`skills/flow-contracts/operator-prompts-auto-resolution.md`) — and every phase's contract is read
under this one.

**A run's size is never a stop.** Tokens spent, wall-clock time, context length, the number of
rounds or tasks still ahead, a natural-looking checkpoint, the length of a phase file still to be
read or the lines already loaded, and a list of open items that each carry
a recommended option are never reasons to end a turn or to hand a status report to the operator
mid-run (**Auto-resolution**, `skills/flow-contracts/operator-prompts-auto-resolution.md`). A phase file too long for one read is read in
offset chunks and the run continues, and, with no explicit operator stop, is never handed to the
operator as 're-run /flow to continue'.

**A failing tool, guard or environment is cured in the run, never handed off.** A guard's exit 2,
a script that errors, a stack that will not start, a stale environment and an out-of-date tool are
defects to fix, in the project's dev tooling or in `<agents repo>`
(**Pipeline defects found mid-run**, below), and the step runs again, however many rounds that
takes. Where a phase file says such a failure stops the run, asks, or ends its turn — and where a
prompt's recommended option is itself a stop — the stop holds only for a cause out of the run's
reach (**What still stops**, `skills/flow-contracts/operator-prompts-auto-resolution.md`);
otherwise this paragraph wins, as **Work end to end — stop only to plan or to test by hand**
(`rules/work-end-to-end.mdc`) requires.

**A turn ends only at a gate, a block, or the run's final handoff.** Between any two steps of a
phase — findings recorded and the fix round not yet made, a fix committed and its re-review not
yet launched, a dispatch returned and its result not yet acted on — the next step is taken in the
same turn, never announced as "next". A turn whose last action is anything other than an operator
prompt **What still stops** names, a blocked report, or the handoff the state's run ends on is
this rule broken. Ending on
"the turn has been long", "each step is a real dispatch or build", "send `/flow` to continue" or
any status report that offers to continue is this rule broken too.

**A wait on background subagents is one `flow-guard await-rows` call, never a turn end** (**Turn
discipline**, `skills/flow/implement.md`).

A review-panel fix round, a verify stage's fix loop, a verifier re-dispatch, the implement phase's
full-suite fix and a guard or integrate call re-run after a stop each repeat until their own clean
close or **The round ceiling** below. **A round that makes no progress — the same defect identity, on the same
evidence, after a fix — changes the run's approach instead of stopping or asking.** The next fix
is a fresh dispatch of that loop's own fixer row on the `no-progress-fixer` role (**Allowed
pairs**, `skills/flow-contracts/model-policy.md`, recorded with its pair), its prompt
carrying `Role: no-progress-fixer` and every earlier round's report
verbatim and naming the approach that failed; a loop whose fix the parent applies inline applies
it from those same reports by a different approach. Where the evidence itself may be wrong, the
defect is first reproduced another way — another reproducer, another capture route. A further
round without progress changes the approach again until the ceiling defers the item; none of them
stops the run.

### Three failed rounds

**Three failed rounds of one loop are a round without progress too, whatever their defects.** A
failed round is one that ends without the loop's own clean close. The fix loop counts three failed rounds of one site of **The fix loop** (`skills/flow/fix.md`), a round as that site's row defines it. The fourth round, and each
later one below the ceiling, is the fresh dispatch above — its fixer —
on the `no-progress-fixer` role carrying every earlier round's report
verbatim and naming the approach that failed. It is never a stop and never an ask.

**A `no-progress-fixer` dispatch carries one defect family.** Where the reports name several
independent families, each gets its own fixer, dispatched one after another — the next only
once the previous has reported, so no two fixers' changes intersect — and each commits its own
work — a WIP commit counts — before it reports, so a stalled fixer resumes from committed work
.

**A fixer that never reported has its row ended before it is resumed or replaced.** When the
parent gives up on or resumes any fixer whose dispatch row is still open, it first ends that row
with `flow record dispatch end -change <name> -key <key> -outcome aborted`, so no row stays open with no outcome.

**A verify round whose backend is down is no round.** Before each verify or visual-verify round,
`flow-guard check-verify-env <url>...` dials every backend URL the project's `## run` starts; exit
1 is an environment failure, cured by starting or restarting that backend before the round runs,
and never counted as a failed round.

### The round ceiling

**The round ceiling bounds a loop's rounds on one item, never the run.** After each failed round
of a loop above, `flow-guard round-ceiling <worktree> <failed rounds>` names the next one, the
failed rounds counted as **Three failed rounds** counts them: `NEXT: retry`, the loop's own fixer;
`NEXT: rethink`, the escalation of **Three failed rounds**; `NEXT: defer`, once the failed rounds
reach the project's `## fix round ceiling` (`skills/flow-contracts/project-configuration.md`), 6
when the key is absent. The guard's header is canonical for its output and exit codes.

**`NEXT: defer` takes the item out of its loop and the run continues.** No further round is
dispatched on that item; every other step of the run goes on in the same turn. The item is
recorded once, as one line in the run's narrative (**Write `IN_PROGRESS`**,
`skills/flow/verify-and-handoff.md`; a `/flow-interactive` request's report) and one
`**Decisions:**` bullet of the run's summary:

```text
Deferred at the round ceiling: <the item, one line> — <failed rounds> failed rounds; approaches tried: <each, one phrase>; reports: <each round's report path>
```

**A deferral waives no gate.** A deferred item that a landing check still reads — a red `lint` or
`test` line, a `flow.verify` not ended `completed`, an open Critical or Important finding — holds
the landing: the change never lands red (**The sweep**, `skills/flow-contracts/known-bugs.md`), and
the run's handoff carries the item on its `**Decisions:**` line for the operator to take up.

### Caps never stop the run

**Every cap, ceiling, budget or count limit this pipeline names bounds an approach, never the
run** — a planner's agent cap, a dispatch's wall-clock ceiling, a read or git budget. Meeting or being
refused by one is a failed round of that approach: the next attempt differs — a smaller scope, a
split dispatch, another route — and where
no different approach exists the work goes past the bound with the fact recorded. It is never a
stop and never an ask, and it counts toward **Three failed rounds** above.

## Unanswered mid-run asks

**An operator ask that goes unanswered — the operator absent, the prompt stalled — resolves to
the safest course that still makes progress, and never hangs the run on it.** Of the ask's own
options, silence takes the one whose risk to the change is lowest among those that move it
forward; an option that stops the run is taken only where every option that moves it forward is
unsafe, and silence is never consent to an irreversible or outward-facing action (**What still
stops**, `skills/flow-contracts/operator-prompts-auto-resolution.md`). The run neither re-asks
nor invents a course the options do not name. A silence outcome a call site states for itself —
a marked default, a gate that acts only on an explicit operator instruction — governs over this
rule. The resolution is recorded with the reasoning: a panel run in the pass log
(`flow record pass`), any other run in its narrative (**Write `IN_PROGRESS`**,
`skills/flow/verify-and-handoff.md`), naming the options considered and why the one taken is the
safest that still makes progress.

**An operator message that arrives mid-run is not a stop by default.** A question, an aside or a
side request the operator sends while a run is executing — about the run, the change or anything
else — is answered or done in the run's own output, and the same turn then continues into the
run's next step. **An explicit instruction
to stop is the one exception** — "stop", "pause", "wait for me", "don't continue": the run
finishes the step in flight, waits for every child it launched, answers what the message asked,
and ends the turn at the state it reached, naming the next step and `/flow <name>` to resume. A
question, a complaint or a request to change something is never read as that instruction; only
words that ask the run to stop are.

## The parent is the only orchestrator

**Every dispatch a `/flow` or `/flow-self-review` run makes does the work itself;
none of them is an orchestrator.** The parent session dispatches each implementer,
fixer, reviewer and verifier directly, every one with the NO DELEGATION paragraph
(`skills/flow/implement.md`) and a `subagent_type` that carries no `Agent` tool — never a
`general-purpose` or other agent whose job is to dispatch, sequence or wait on others. Where a
contract says "this session" or "the parent" runs a loop, the parent runs it, even when the loop
is long. An orchestrator duplicates the parent's context, idles between its children's returns
until its prompt cache expires and re-reads it all on every wake, and resuming it is the
`SendMessage`-to-a-finished-child shape this pipeline bans. The brainstorm planner alone dispatches
through `Agent` (**PLANNING DELEGATION**, `skills/flow/brainstorm.md`).

**Only the parent waits.** A subagent starts fresh, does one job, writes its result to disk and
exits; work that follows it starts in a fresh one from that file (**Dispatch sites — the parent's closed list**,
`skills/flow/implement.md`). The parent runs on the one-hour prompt cache and a subagent on the
five-minute one, so a subagent that waits, or is resumed after a gap, pays a full cache rewrite.

## Pipeline defects found mid-run

**Load `skills/flow-contracts/pipeline-defects.md` only when** a run hits a defect in the pipeline
itself — a skill, a contract, a guard, the `flow` CLI or flowd, a rule, or a saved memory that
encodes its behaviour; it carries the whole in-run fix procedure, its blast-radius class and its
record.

**A confidence estimate below 100% is one row of its own**, never a self-review finding:
`flow confidence tags` lists the tags in use and `flow confidence record` files the estimate —
by the planner triaging a plan check (**Plan triage**, `skills/flow/plan-triage.md`) and by
whoever turns a task group green (**CONFIDENCE ONCE GREEN**, `skills/flow/implementer-brief.md`).
`flow confidence` with no argument prints both usages.

**A gap names a missing capability, never a particular thing.** Each `-gap` names a capability or
process step the pipeline or its tooling lacks, whose presence would move the score to 100 for any
change ("no guard compares the capture against the frame"). A gap about one particular thing of the
change at hand — a frame, a row, a plan line, a file, a fixture — is never filed: it is asked to the
operator at the plan gate, or fixed in the task.

## Wrong state for this command

**On a mismatch, stop.** Report the actual state, the states the command expects, and the command
that should run instead, then **AskUserQuestion** for an explicit override with **"No — run the
suggested command instead"** as the default and recommended answer. Only proceed when the user
explicitly chooses to override. Never advance from a wrong starting state silently.

```
## Wrong state for this command

**Change:** <name>
**Current state:** <actual> (set by <updatedBy>, <updatedAt>)
**This command expects:** <expected>
**Suggested instead:** <what to do instead, e.g. "nothing — the change is dropped">
```

## Progress visibility

**`/flow` registers nothing with the harness's task-list mechanism.** Its live progress view on
Claude Code is the `flow-task-list` mod (`mods/flow-task-list/`). **Load
`skills/flow-contracts/progress-visibility.md` only when** the harness lacks the mod, or the run
edits the mod, the board's rows or labels, or `tasks.md`'s checkbox markers — it carries the
band's rows, the numbering rules, the view-never-a-record rule and the equivalent block a harness
without the mod prints.

**Every subagent dispatch's description is the board row's label**, in the shape
**Dispatch sites — the parent's closed list** (`skills/flow/implement.md`) states — no emoji and
no state of its own: the mod adds both, and numbers a task's row from its `Task <x>/<n>` prefix.
Claude Code's own agent panel, not the board, shows a running agent's latest tool-call
description, so every dispatch prompt also tells the agent to open each tool call's description
with its dispatch description's unit and its round (`Task 30/31 review-1 — run the drawer tests`,
`Visual verify review-2 — resolve the specs`).

## Quiet progress

**Opt-in, off by default.** A `/flow` run resolves the project's `## progress` key once, at its
start, with `flow-guard project-get <main-checkout> progress --enum quiet` (**Project configuration**,
`skills/flow-contracts/project-configuration.md`): exit 0 turns the mode on for the whole run; exit
1 leaves it off; exit 3 is reported by name and leaves it off.

**Load `skills/flow-contracts/quiet-progress.md` only when** that call exits 0 — it carries what the
mode silences and what still prints.

## Stage marks

A stage mark is a passive record for the stats: `flow stage begin` when a phase starts, `flow
stage end` when it closes, naming the command, the stage key and the change. A mark never blocks,
delays or alters the phase it marks — it refuses nothing but a caller mistake (an undocumented
key, a missing flag), and on a store failure the CLI journals the intent, prints one warning and
exits 0. Never branch on its exit code; a failed mark is one line in the run's narrative.

`/flow` marks one stage per phase. Planning marks `flow.plan`, from kickoff to the plan gate
(`skills/flow/brainstorm.md`); the other phases' keys are listed in **Stage keys**
(`skills/flow/stage-keys.md`). A mark and a `flow-guard` call carry no `-harness`: the CLI names
the harness actually running the mark — `$FLOW_HARNESS`, else `claude-code` on Claude Code; a
harness that sets neither passes `-harness <its name>`. A `flow record` call given `-C` inside the
change's worktree carries no `-change`: the CLI reads the change off that checkout's branch.
**The session token is the calling agent's, and inferred.** A `flow stage` or `flow record` call,
and a `flow-guard` call that marks or records, carries no `-session-token`: the CLI uses
`cc-<harness session id>-<agent id>`, `main` on the main thread. Subagents share their parent's
harness session id, and the store supersedes a token's open run of the same `-request` (or of none,
when the begin carries none) on its next begin, so the agent id
is what keeps a subagent's mark from closing its parent's run: the block-unsafe-commands hook adds
`-caller-agent <agent id>` to a subagent's `flow` and `flow-guard` calls
(`stats/internal/caller/caller.go`). A call the hook cannot see as one — inside `bash -c` or a
quoted string — runs under the main thread's token, so a mark or record is always its own
top-level call. A harness with no session id passes `-session-token <literal>`; an explicit token
always wins. A `flow stage end` closes the open run of its command, stage and change begun under
the caller's inferred token, so two agents inside the same stage each close their own; with no
inferred token, or none of its runs open, it closes that stage's highest open attempt. An end given
`-request <n>` looks only among request `<n>`'s runs, so two requests in the same stage never close
each other's.

```bash
flow stage begin -command '/flow' -stage flow.plan <name>
# … the phase's own work …
flow stage end   -command '/flow' -stage flow.plan -outcome completed <name>
```

The dispatch hooks begin and end a dispatch's row under its dispatcher's token, the same way.

## Handoff output

Every run ends in the same shape, and prints **nothing** after it:

````
<1–3 lines: what actually happened>

<absolute paths to anything the operator needs to open>

Next:
```
/flow <name>
```
````

- **A block is printed as rendered Markdown, never inside a code fence.** The fence around every
  block shape in a skill or contract file — this one, `## Finished`, `## Branch integrated`, the
  wrong-state block and every other — delimits the template, not the output: printed fenced, its
  `**Field:**` markup reaches the operator as literal asterisks. Each `**Field:** value` line is
  printed as a list item (`- **Field:** value`), since consecutive plain lines fold into one
  paragraph, and a path, branch or change name inside a value is inline code. `/clear` stays
  a bare line, and the next command is fenced, per the next rule.
- **No HTML in anything printed to the terminal** — no `<details>`, `<summary>`, `<br>` or any
  other tag. The terminal renders GitHub-flavoured Markdown without HTML, so a tag prints
  literally and whatever it was meant to fold spills into the output.
- **The next command is the last thing printed, alone in its own fenced code block** — a plain
  triple-backtick fence, no language tag, holding the command and nothing else, with no prose after
  it, so a client with a code-block copy button copies it in one tap.
- **A handoff printed at `IN_PROGRESS`, or at `STARTED` once planning is complete,
  puts `/clear` on the line above the next command's fence.** It is a recommendation the operator may skip, never a gate, and the next
  command's fence stays last.
- **A bare invocation at `IN_PROGRESS` that opened a PR or handed off manually names itself** as the
  next command, because that is what the operator runs once the branch is merged. Only a run that
  **completed** cleanup is terminal and names nothing; a run that stopped on a cleanup
  leftover names itself too, for the same reason: the operator clears what remains and runs it
  again.
- **A red check is fixed, never handed off.** Every lint, guard, test or build the run executes
  that comes back red — the change's own failure or one already red on the base — is fixed in
  that run, at its root cause (a flake per `rules/fix-determinism-at-the-source.mdc`), before the
  handoff prints. The handoff never asks whether to fix it. Only a fix outside the run's reach —
  production, a repository the run was not given — is reported, with the evidence.
- **Only what the operator must act on.** Do not restate the plan, enumerate completed internal
  steps, or repeat content available at a path you just gave.
- **Link, never paste.** Diffs and plans are given as absolute paths.
- **Every path is absolute** — in handoffs. Never a
  relative path, never `../<other-app>`, and never a main-checkout path while an apply worktree
  holds the work — and no `/flow` step checks out, stages or commits in the main checkout. Resolve
  app roots from `git worktree list` or the state file's `worktrees` keys.
- **Implementation never stages `<project>/spectre/changes/` into its own commits**, and the path
  is fixed here rather than configured per project. A change directory's `link.md` is deliberately
  not on it — see **Git boundaries** (`skills/flow-contracts/git-boundaries.md`) for why. Leaving them unstaged — rather than filtering a display — is what keeps them out of every view of the staging area:

  ```bash
  git -C <abs-worktree> diff --cached --stat
  git -C <abs-worktree> diff --cached
  ```

### Summary and live-stack line, before every handoff

Immediately before the block above is printed, the run prints a summary of what this run actually
did, then one line naming the state of any local dev stack it started or left running. The summary
covers:

- **what this changes for a user of the app, and why** — the feature or fix in product terms
  (what was added, changed, fixed or removed, and the purpose stated in the plan's own `## Why`),
  never only the mechanics of the run. A reader who never opens the diff still learns what shipped.
- the pair each of this run's dispatches ran on, per **Model and effort**
  (`skills/flow-contracts/model-policy.md`) — so the operator reads the run's model policy off the
  summary without opening the ledger;
- findings fixed, and findings withdrawn;
- commits made — repo and a one-line description each, each commit named by its subject, never by
  a sha integrate's reshape rewrites;
- any decision recorded along the way — an automatic rebase onto a moved base, a diff-size cap
  exceeded and proceeded past, and the like.

The live-stack line names which services are up, at the URLs their worktree resolved, in which
worktree, with the local credentials `<project>/.flow/project.md`'s `## credentials` declares — or
reads `no dev stack running` when none is.

This is distinct from the `<1–3 lines: what actually happened>` that follows it: the summary may
be a short bulleted list and is not held to the three-line cap, but it keeps the be-brief
conventions — bullets over prose, no preamble, no recap of the plan. **Every run that reaches a
stopping point prints it** — a fix round closing, a stage closing, a full handoff — not only the
terminal handoff, so the operator always holds a current account of what changed and what is
still running without having to ask.

**Every summary a run prints is recorded as it prints** — `flow record summary -C <canonical-worktree> -file -`, the summary's Markdown on stdin, the `IN_PROGRESS` handoff's `**Summary:**` bullets included; one row per change, the last write wins — so the self-review bundle's first source is the run's own statement, as **Write `IN_PROGRESS`** prints it (`skills/flow/verify-and-handoff.md`).

**The `IN_PROGRESS` handoff of a creating, resuming or fix run prints neither before its block.**
Its own `**Summary:**` carries the first bullet above alone, and its block is the whole handoff (**Write `IN_PROGRESS`**,
`skills/flow/verify-and-handoff.md`).

## Artifact brevity

**Every artifact a `/flow` run writes is written brief** — bullets over prose, no preamble, no
recap, no restatement of what another artifact in the same change already says. It binds
`proposal.md`, `design.md`, `tasks.md`, the SDD ledger, the review panel record and
the self-review report.

**Brevity never withholds a fact.** Wording is compressed; a decision, a reason, a measured number,
an alternative that was ruled out, or a caveat never is. An artifact that got shorter by losing one
of those is a defect, not a brief artifact.

**Never compressed** — a guard or a contract parses each of these byte for byte:

- a task's `Files:`, `Tests:`, `After:`, `Size:` and `Commit:` fields;
- a decision's or an open question's `ID:` and `Status:` lines;
- the review panel record's marker blocks and its findings table.

This narrows the "code, commits, docs and specs stay full" carve-out the be-brief rule
(`rules/be-brief.mdc`) states — for the artifacts named above, and for no other file.

**No length guard and no byte budget measures a change artifact.**

## Guard resolution

**A named guard runs as `flow-guard <name>`.** `flow-guard` resolves on `PATH` to the shim
`<agents repo>/setup.sh global` installs, which execs `<agents repo>/scripts/flow-guard` — the binary built
from that checkout, never a `flow-guard` installed anywhere else. Skills and contracts name a guard
by that command, never by a path relative to a repository root.

## Guard presence check

`/flow` checks, once at the start of its run, that `flow-guard` resolves on `PATH`. A resolving
command prints nothing. Otherwise it prints exactly one block, then the run stops:

```text
⛔ GUARDS MISSING — flow-guard not found on PATH
Re-run ./setup.sh global to install it, then re-run /flow.
```

**This is a gate.** A guard that cannot run stops the run — never a hand substitute.

## Hand-verifying a guard verdict

**Load `skills/flow-contracts/guard-verdict-verification.md`** only when a gate guard's verdict fires — `DRIFT-DIRTY`, `REFUSE`, `OUTSTANDING` or `MOVED`.

## State file

The contract governing where a change's state file lives, its full JSON shape, monotonic state
writes, and carry-forward rules. **State file** (`skills/flow-contracts/state-file.md`) — load it
before reading or writing a state file.

## Project configuration

The contract governing `<project>/.flow/project.md` — its optional keys, how a `## standards` entry
resolves to a file, and the containment rules that keep resolution safe.
**Project configuration** (`skills/flow-contracts/project-configuration.md`) — load it before
resolving project configuration.

## Jira integration

The contract governing how a change is linked to a Jira issue, transitioned through the pipeline,
and has its description synced — including that Jira is never a gate and never blocks a state
write. **Jira integration** (`skills/flow-contracts/jira-integration.md`) — load it before any
Jira-related step.

## Change name resolution (all `/flow*` commands)

`<name>` is **optional** on `/flow`.

**The candidate set is `flow state resolve -C <main-checkout>`'s `candidates` — never a
hand-written HTTP call, never a directory listing of your own.** It prints one JSON object:
`"source"` (`"store"` or `"fallback"`), `"complete"` (`true` only for `"source":"store"`),
`"candidates"` (each carrying `name`, `state`, `updatedAt`, `updatedBy`) and `"unreadable"` (the
names of fallback records that could not be read). A `FINISHED` change is never a candidate.

**A record `flow state resolve` marks `"unreadable":true` is reported and skipped from the union — never
silently dropped.** Name the unreadable file in the resolution's own output; do not fold it into a
"zero matches" or "no change" result as if it were never there.

**Every command that resolves this candidate set reports which of the two sources produced it** —
`flow state resolve`'s own `"source"` field, echoed rather than re-derived.

Once the candidate set is built:

- Exactly one match → use it automatically; announce which change was picked.
- Multiple matches → **AskUserQuestion** listing each (name, state, last modified) — never guess.
- Zero matches → fall back to that command's normal "no change" handling (e.g. a creating `/flow`
  run asks what to build).

**An argument that is exactly a candidate's name names that change and nothing else.** At
`IN_PROGRESS`, `/flow <name>` is a bare invocation for that change — an integrate run, never a fix
run carrying the name as its instructions — which is why every handoff's `Next: /flow <name>`
starts integrate.

A change linked to a Jira issue is named `<lowercased-key>-<slug>` — see
**Change naming** in `skills/flow-contracts/jira-integration.md`.

## A change's directory

A change's directory is `<project>/spectre/changes/<name>/`, on disk in the worktree, never
committed: `<project>/spectre/changes/` is gitignored and there is no archived path — nothing of it lands. Every step that reads or writes a change's planning artifacts resolves the
directory this way; the guards resolve it through `changePlanDir`
(`<agents repo>/stats/internal/guard/changeplan.go`), and `flow tasks`
through the same rule.
