# flow pipeline

The three-state pipeline itself: state definitions, the command→state transition table and the
handoff shape. See **Finish contract** (`skills/flow-contracts/finish-contract-run1.md`,
`skills/flow-contracts/finish-contract-run2.md`).

**Load this file when running `/flow`.**

This file is **canonical** for everything in it. Where a skill or command disagrees with it, this
file wins.

The reasoning behind this file lives in `skills/flow-contracts/pipeline-rationale.md`;
**a `/flow` run never loads it.**

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
/flow  (reachability end)  → FINISHED (withdrawn)     nothing to plan — the operator may withdraw
/flow  (STARTED, planned)  → IN_PROGRESS              you: review the staged diff — the stack is running
/flow  <fix instructions>  → IN_PROGRESS (unchanged)  you: review the staged diff — the stack is running
/flow  (bare, IN_PROGRESS) → IN_PROGRESS or FINISHED  terminal only on the merge-and-push route — see the finish contract
```

**The human gate is a property of the state, not a separate stage.** `IN_PROGRESS` *means* a
staged diff is waiting for the human to review, alongside a stack the handoff already started.
Nothing records that the review or the testing happened; the operator running the next
command is what carries the change forward.

| State | Means | Waiting on |
|-------|-------|-----------|
| `STARTED` | The change exists; planning is under way or the plan awaits implementation | you — `/clear`, then `/flow <name>` |
| `IN_PROGRESS` | The implementation is staged and the stack is running | you — review the diff — the stack is running |
| `FINISHED` | Archived, pushed, worktrees removed — or withdrawn: a change abandoned before planning, its worktree and branches deleted and its record's `withdrawn` field set (`state-file.md`) | — |

**Reviewing and testing are one gate.** A creating or fix run of `/flow` produces both surfaces in
the same run, so the human does both at one sitting. There is no state between implementation and
finishing — integration is not a stage, it is the first half of finishing.

## Command surface

One command, `/flow`, drives the whole pipeline, plus one read-only command (`/flow-status`),
one that creates a change at `STARTED` and stops there (`/flow-plan`), one minimal-ceremony
variant that writes no state file (`/flow-fast`) and two standalone, non-pipeline commands
(`/flow-settings`, `/flow-self-review`). **No command accepts
a flag**, save one: `--base <branch>` on a run that creates a change — `/flow`, `/flow-plan`,
`/flow-fast` — names the branch it is cut from and lands on, in place of the one `origin/HEAD`
points at (`skills/flow/SKILL.md`). The only argument is the optional change name — see **Change name resolution** — or,
on `/flow` and `/flow-fast`, a description or Jira key seeding a new change; on `/flow-plan`, a
topic; on `/flow`, fix instructions at `IN_PROGRESS`.

An argument that is none of those is **reported**, not silently ignored.

## State transitions

| Command | Accepts | Ends at |
|---------|---------|---------|
| `/flow` | *(no state — creates the change)* | `STARTED` — the run ends at the plan gate with a `/clear` handoff (see **Resuming at `STARTED`** in `skills/flow/resume.md`); or `FINISHED` (withdrawn) when the reachability check ends the run and the operator answers the withdraw ask |
| `/flow` | `STARTED` | resumes the creating run from wherever it stopped; ends at `STARTED` (still planning, or planning just finished — `/clear` handoff) or `IN_PROGRESS` (the plan was already ready); or `FINISHED` (withdrawn) on an explicit withdraw answer at a planless resume (**The withdrawal route**, `skills/flow/withdrawal.md`) |
| `/flow` | `IN_PROGRESS`, with an argument | fix run; state unchanged |
| *(none — a plain message)* | `IN_PROGRESS`, in the session whose last `/flow` run marked the change | fix run; state unchanged — **A plain message at `IN_PROGRESS`** (`skills/flow/SKILL.md`) |
| `/flow` | `IN_PROGRESS`, bare | integrate run; ends at `IN_PROGRESS` (run 1) or `FINISHED` (run 1 chained into run 2) |
| `/flow` | `FINISHED` | wrong-state handoff — the change is archived |
| `/flow-status` | any — read-only, never block | unchanged |

Which phase file marks each `flow.*` key is **Stage keys** (`skills/flow/stage-keys.md`).

**This table is authoritative.**

### Every invocation is re-entrant

**The bare invocation that starts integrate must be an actual new `/flow` (or `/flow <name>`)
command from the operator — never inferred from anything said inside a still-running turn.** A
running `/flow` turn is mid fix run, mid a question the run itself asked, or simply still executing
— and an instruction the operator gives during that turn, however explicit ("merge and push", "go
ahead", an answer to an unrelated question with a trailing aside about landing the branch), is
conversation, not a re-invocation. This is what the human gate at `IN_PROGRESS`
(**States**, above) actually rests on: the state transition table's "bare" column means a bare
*command*, not a quiet moment in an ongoing one. **Concretely: never chain straight from a fix, a
mid-turn question, or a dispatched-agent report into integrate's unfinished-work gate, its landing
question, or any git action the chosen route performs — even when the operator's own words include
"merge and push" — without the turn first ending and a fresh bare invocation starting it.** On a
mismatch, treat it exactly as **Wrong state for this command** below already requires: report what
was said, name the bare re-invocation as what actually starts integrate, and stop rather than
inferring consent.

Implementation advances the state **only** from `STARTED` to `IN_PROGRESS`. A fix run at
`IN_PROGRESS` writes the state back exactly as it found it.

No field records where a fix was raised. Whether the human re-reviews the diff or re-runs the apps
after a fix is their decision.

## Fewest operator actions

**The operator's actions and the run's stops are minimised: a run resolves everything it can
itself and continues.** It stops only for a decision only the operator can make, or for an
action that is irreversible or outward-facing — **What still stops**
(`skills/flow-contracts/operator-prompts-auto-resolution.md`) — and every phase's contract is read
under this one.

**No round count ends a re-review or a re-verification.** A review-panel fix round, a gated
per-task re-review, a verify stage's fix loop and a verifier re-dispatch each repeat until their
own clean close. **A round that makes no progress — the same defect identity, on the same
evidence, after a fix — changes the run's approach instead of stopping or asking.** The next fix
is a fresh dispatch of that loop's own fixer row on `opus` at effort `high` (`subagent_type:
flow-high`, recorded with that pair, mapped on harness `zcode` per **Harness mapping**,
`skills/flow-contracts/model-policy.md`), its prompt carrying every earlier round's report
verbatim and naming the approach that failed; a loop whose fix the parent applies inline applies
it from those same reports by a different approach. Where the evidence itself may be wrong, the
defect is first reproduced another way — another reproducer, another capture route. A further
round without progress changes the approach again; none of them stops the run.

## Pipeline defects found mid-run

**A `/flow` or `/flow-fast` run that hits a defect in the pipeline itself fixes it within the
run, unless the fix is `big`.** The pipeline is everything `<agents repo>` ships — a skill, a
contract, a guard, the `flow` CLI or flowd, a rule — and any saved memory that encodes its
behaviour. For this section alone `<agents repo>` is a repository the run was given
(**Reporting back**, `rules/agent-baseline.md`), whichever project the run is in.

**Predict the class before any work, from the blast radius — never from a plan.** Count the
files the fix must touch: the defective file, its test, and every file `grep -rl` finds in
`<agents repo>` citing the name or section the fix changes. The fix is `big` — `plan-class.sh`'s
own `files>=60` threshold — when that count is 60 or more, when it reaches outside
`<agents repo>`, or when it needs a design choice only the operator can make. A `big` fix is
deferred to self-review, and so is one found `big` once under way, which stops there.

`<agents-base>` is `<agents repo>`'s own default branch —
`git -C <agents repo> symbolic-ref --short refs/remotes/origin/HEAD` with its `origin/` prefix
dropped — never the project's `<default-branch>`, which `<agents repo>` may not carry.

**Every dispatch in the loop is one-shot.** No `SendMessage` to a finished child, and no child
parks a question for a resume. Each runs on `opus`, `subagent_type: flow-high`, its key per the
steps below and `<k>` counting this run's in-run fixes. Each prompt carries the MODEL HANDSHAKE,
TOOLS, FOREGROUND BUILDS and NO DELEGATION paragraphs of **4. Execute (SDD + TDD)**
(`skills/flow/implement.md`) verbatim, the review's READ-ONLY REVIEW too, and **The handshake**
there applies to each reply:

1. **Fix** — `pipeline-fix-<k>`, its prompt carrying the defect, its evidence and the counted
   blast radius. It works in its own worktree,
   `git -C <agents repo> worktree add -b fix-<slug> <agents repo>/.worktrees/<slug> origin/<agents-base>`
   — never the main checkout, never the change's own worktree — runs `<agents repo>`'s
   `## worktree setup` in it (`project-get.sh <agents repo> 'worktree setup'`), adds one test or guard that
   fails without the fix, updates a saved memory that encoded the defect, runs the
   `<agents repo>/.flow/project.md` `## lint` lines its files need, and commits with a module
   scope. It never merges or pushes.
2. **Review** — a fresh `pipeline-fix-<k>-review-<r>` over
   `git diff origin/<agents-base>...fix-<slug>`.
3. **Fix the findings** — the parent fixes them inline in that worktree, or dispatches a fresh
   `pipeline-fix-<k>-fix-<r>`; then step 2 again, `<r>` plus one. The loop runs to a clean
   review under **Fewest operator actions** above.
4. **Land** — by `<agents repo>`'s `## default landing route`
   (`project-get.sh <agents repo> 'default landing route'`), without asking. Merge and push is:
   `git -C <agents repo> fetch origin`, rebase `fix-<slug>` onto `origin/<agents-base>`,
   `git push origin fix-<slug>:<agents-base>`, `git -C <agents repo> pull --ff-only`, then
   remove the worktree and the branch.

The loop runs in the background while the run's own work continues; the run still never ends a
turn with one of its children in flight.

**Record each one.** Every in-run fix, landed or deferred, is one line in the run's narrative —
`narrative.md` on `/flow` (**Write `IN_PROGRESS`**, `skills/flow/verify-and-handoff.md`),
`## Session narrative` on `/flow-fast` — and one decision bullet in the run's summary
(**Summary and live-stack line, before every handoff**, below):

```text
In-run pipeline fix: <agents sha | deferred> — <the defect, one line> (blast radius <N> files)
```

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
**Suggested instead:** <what to do instead, e.g. "nothing — the change is archived (or withdrawn, per the record's `withdrawn` field)">
```

## Progress visibility

**`/flow` registers nothing with the harness's task-list mechanism.** Its live progress view on
Claude Code is the `subagent-board` mod (`mods/subagent-board/`): a band above the prompt with one
row per subagent, drawn from each dispatch's description, and a hint-line tail naming the tracker
key where the change name leads with one, the phase, and the running stage's emoji where it has one
— read from the run's own `flow stage` marks (**Stage marks**, below) — followed by a tally of those
rows by emoji. Work no subagent carries — brainstorming, an inline-executed task, a finish step —
shows on the main agent's own `main` row, first on the band while the main turn runs, labelled with its
latest Bash command's description, if any, and never numbered as a task.

**Every subagent dispatch's description is the board row's label**, in the shape
**Dispatch sites — the parent's closed list** (`skills/flow/implement.md`) states — no emoji and
no state of its own: the mod adds both, and numbers a task's row from its `Task <x>/<n>` prefix.
Claude Code's own agent panel, not the board, shows a running agent's latest tool-call
description, so every dispatch prompt also tells the agent to open each tool call's description
with its dispatch description's unit and its round (`Task 30/31 review-1 — run the drawer tests`,
`Visual verify review-2 — resolve the specs`).

**The progress view is a view, never a record.** No command, guard or contract reads it back as
evidence of what was done. `tasks.md` remains the single source of truth for a
plan's completion state, and `<agents repo>/scripts/check-unfinished-work.sh` reads that file.

**No third checkbox marker is added to `tasks.md`** to carry an in-progress state; the in-progress
count comes from no persisted record at all — the board's rows are session state, never written
to disk. See **Progress visibility** (`skills/flow-contracts/pipeline-rationale.md`) for why a marker would be unsafe.

**On a harness without the mod (ZCode), `/flow` prints the equivalent block instead** — a count
line naming how many steps are done, in progress and open, followed by one line per step marked
done or not done. One step per whichever cited stage is running at the time, at that stage's own
granularity — brainstorming checklist items and artifacts on the creating/resuming branch, tasks on
the implementation branch, a finish run's steps on the integrate/archive branch.

## Quiet progress

**Opt-in, off by default.** A `/flow` run resolves the project's `## progress` key once, at its
start, with `project-get.sh <main-checkout> progress --enum quiet` (**Project configuration**,
`skills/flow-contracts/project-configuration.md`): exit 0 turns the mode on for the whole run; exit
1 leaves it off; exit 3 is reported by name and leaves it off.

**With the mode on, the run writes no status prose of its own on a routine event** — dispatching a
subagent, a subagent returning clean, sending a subagent a decision, waiting on a background agent,
a guard passing, a commit landing. This suspends the per-unit and what-I-am-on lines of the
"Keep the user posted" paragraph of `rules/be-brief.mdc` for the run, and nothing else in that
rule: a turn still ends only where it says.

**It likewise writes nothing for** a bookkeeping note about a plan-field correction ("Q2
re-captured unchanged, so its baseline is dropped from task 14's Files.") or an end-of-turn
progress summary ("Task 13 passed review and is ticked … Task 14 … is running now."), a narrated
step ("Now commit the frontend link and run close-task for tasks 1–6.", "Run RED first.", "Now the
fix."), or a stage-level status line ("🔍 Review panel (primary + principles) — in review"). When
nothing needs the operator — no stop, no auto-resolution line, no handoff — the turn ends with no
text at all.

**What still prints:**

- a stop for the operator — a question, a `## Question` handback, a blocked step;
- one line per auto-resolution taken, naming the question and the option (**Auto-resolution**,
  `skills/flow-contracts/operator-prompts.md`);
- the handoff block;
- everything on the **Never compress** list of `rules/be-brief.mdc`, in full.

**Progress visibility** above is unaffected: the subagent-board mod's view, and the equivalent block
a harness without it prints, are not the run's status prose.

## Stage marks

`/flow` marks each of its own stages: `flow stage begin` when the stage starts,
`flow stage end` when it closes, both naming the command, the stage and the change. The stage
identifier is the **key**, never the prose name, from **Level 1 — the stages of each command**
(`<agents repo>/README.md`).

**A `stage begin` mark MUST carry `-session-token` and `-harness`.** Generate the token once, near the start of
the run, before the first `stage begin`, and reuse that exact value at every later mark site in the
same run; do not invent a fresh one per mark. `-harness` names the harness actually running the mark
— `claude-code` or `zcode`.

**Neither `-session-token` nor `-harness` is ever a hardcoded value in the skill text: both are
filled in by the agent at call time, from a placeholder — `<literal-token>` and `<harness>` below.**

```bash
flow stage begin -command '/flow' -stage flow.review-panel -harness <harness> -session-token mf-<literal-token> <name>
# … the stage's own work …
flow stage end   -command '/flow' -stage flow.review-panel -outcome completed <name>
# … a later stage in the same run reuses the same token, not a new one …
flow stage begin -command '/flow' -stage flow.verify -harness <harness> -session-token mf-<literal-token> <name>
```

**The session token MUST be a literal, written directly into the command — never a shell
substitution.**

**A mark never blocks, delays, or alters the stage it marks.** The only nonzero exit is a caller mistake — an
undocumented stage key, a missing required flag, or a session token carrying a substitution shape —
which is a defect in the skill's own call, not an outcome of the stage, and is fixed by correcting the
call rather than worked around. On any store failure the CLI journals the intent, prints one warning
line, and exits 0 — the same never-block guarantee **State file**
(`skills/flow-contracts/state-file.md`) already states for `state set`. Do not branch on
`flow stage`'s exit code as a signal about the stage itself: a mark that could not reach the store
still exits 0, so there is nothing to react to.

**`stage end` carries no session token and no harness of its own.**

## Handoff output

Every run ends in the same shape, and prints **nothing** after it:

```
<1–3 lines: what actually happened>

<absolute paths to anything the operator needs to open>

Next:
/flow <name>
```

- **A block is printed as rendered Markdown, never inside a code fence.** The fence around every
  block shape in a skill or contract file — this one, `## Finished`, `## Branch integrated`, the
  wrong-state block and every other — delimits the template, not the output: printed fenced, its
  `**Field:**` markup reaches the operator as literal asterisks. Each `**Field:** value` line is
  printed as a list item (`- **Field:** value`), since consecutive plain lines fold into one
  paragraph, and a path, branch or change name inside a value is inline code. The next command
  and `/clear` stay bare lines, per the next rule.
- **No HTML in anything printed to the terminal** — no `<details>`, `<summary>`, `<br>` or any
  other tag. The terminal renders GitHub-flavoured Markdown without HTML, so a tag prints
  literally and whatever it was meant to fold spills into the output.
- **The next command is the last line** — bare, copy-pasteable, with no prose after it. See
  **Handoff output** (`skills/flow-contracts/pipeline-rationale.md`) for why.
- **A handoff that leaves the change at `IN_PROGRESS`, or at `STARTED` once the plan gate answered
  Yes, puts `/clear` on the line above the next command.** It is a recommendation the operator may skip, never a gate, and the next
  command stays the last line.
- **A bare invocation at `IN_PROGRESS` that opened a PR or handed off manually names itself** as the
  next command, because that is what the operator runs once the branch is merged. Only a run that
  **completed** archive — or completed a withdrawal (**The withdrawal route**,
  `skills/flow/withdrawal.md`) — is terminal and names nothing; a run that stopped on a cleanup
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
- **Every path is absolute** — in handoffs, in IntelliJ commands, in run instructions. Never a
  relative path, never `../<other-app>`, and never a main-checkout path while an apply worktree
  holds the work — and no `/flow` step checks out, stages or commits in the main checkout. Resolve
  app roots from `git worktree list` or the state file's `worktrees` keys.
- **Implementation never stages `<project>/spectre/changes/` into its own commits**, and the path
  is fixed here rather than configured per project. `<project>/spectre/specs/`
  and a change directory's `link.md` are deliberately not on it — see **Git boundaries**
  (`skills/flow-contracts/git-boundaries.md`) for why. The planning artifacts reach the branch
  through their own commits (**Planning commits**, `skills/flow-contracts/git-boundaries.md`), so
  nothing is lost. See
  **Handoff output** (`skills/flow-contracts/pipeline-rationale.md`) for why leaving them unstaged
  — rather than filtering a display — is what keeps them out of every view of the staging area:

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
- the models the recorded decision chose for this run's dispatches — each pair, per **Model and
  effort** (`skills/flow/brainstorm-planner.md`) — so the operator reads the run's model policy
  off the summary without opening the ledger;
- findings fixed, and findings withdrawn;
- commits made — repo and a one-line description each;
- any decision recorded along the way — an automatic rebase onto a moved base, a diff-size cap
  exceeded and proceeded past, and the like.

The live-stack line names which services are up, on what ports, in which worktree — or reads
`no dev stack running` when none is.

This is distinct from the `<1–3 lines: what actually happened>` that follows it: the summary may
be a short bulleted list and is not held to the three-line cap, but it keeps the be-brief
conventions — bullets over prose, no preamble, no recap of the plan. **Every run that reaches a
stopping point prints it** — a fix round closing, a stage closing, a full handoff — not only the
terminal handoff, so the operator always holds a current account of what changed and what is
still running without having to ask.

**The `IN_PROGRESS` handoff of a creating, resuming or fix run prints neither.** Its own
`**Summary:**` carries the first bullet above alone, and its block is the whole handoff (**Write
`IN_PROGRESS`**, `skills/flow/verify-and-handoff.md`).

### The block each state renders

The block a state hands off is defined in **The block each state renders**
(`skills/flow-contracts/handoff-blocks.md`). `/flow-status` loads it; `/flow`
carries only the block it prints.

## Artifact brevity

**Every artifact a `/flow` run writes is written brief** — bullets over prose, no preamble, no
recap, no restatement of what another artifact in the same change already says. It binds
`proposal.md`, `design.md`, the change's spec edits, `tasks.md`, the SDD ledger, the review panel record and
the self-review report.

**Brevity never withholds a fact.** Wording is compressed; a decision, a reason, a measured number,
an alternative that was ruled out, or a caveat never is. An artifact that got shorter by losing one
of those is a defect, not a brief artifact.

**Never compressed** — a guard or a contract parses each of these byte for byte:

- a task's `Files:`, `Tests:`, `Regression:`, `Baseline:`, `Commit:`, `Build:` and `Squash-with:`
  fields;
- a plan's `verified:` / `unverified:` / `measured:` / `predicted:` provenance tags;
- a decision's or an open question's `ID:` and `Status:` lines;
- a spec's normative statements and their scenarios;
- the review panel record's marker blocks and its findings table.

This narrows the "code, commits, docs and specs stay full" carve-out the be-brief rule
(`rules/be-brief.mdc`) states — for the artifacts named above, and for no other file.

**No length guard and no byte budget measures a change artifact.**

See **Artifact brevity** (`skills/flow-contracts/pipeline-rationale.md`) for why this is stated
here rather than in each artifact-writing skill.

## IntelliJ commands

Every state that waits on a human prints a copy-paste command in its handoff, save where the table
below says none:

```bash
open -na "IntelliJ IDEA" --args "<absolute path>"
```

Use `open -na`, not the `idea` shim — that shim is not on this machine's PATH. See
**IntelliJ commands** (`skills/flow-contracts/pipeline-rationale.md`) for what `open` buys.

| State | Path to open |
|-------|--------------|
| `STARTED` | none — the `STARTED` handoff is the summary, the decision and the next commands alone |
| `IN_PROGRESS` | apply worktree root — `/flow-status`'s regenerated block only; `/flow`'s own handoff omits it |

Paths are absolute, resolved from `git worktree list`. Never emit a relative path.

## Guard resolution

**A named guard resolves to `<skill-dir>/scripts/<name>`.**
Skills and contracts name a guard by **basename**, never by a path relative to a repository root.

`skills/flow-contracts/` is never a running command and carries no
`skills/flow-contracts/scripts/` directory. See **Guard resolution**
(`skills/flow-contracts/pipeline-rationale.md`) for what resolving against the running command's
own skill directory buys.

## Guard presence check

`/flow` checks, once at the start of its run, that every guard it can invoke —
per **Guard resolution** above — is present in `<skill-dir>/scripts/`. A complete
set prints nothing. Any absence prints exactly one block, naming every missing guard, the
directory searched, and the install command, then the run continues:

```text
⚠ GUARDS MISSING — 3 of 6 not found at
  <skill-dir>/scripts/
    check-finish-preflight.sh
    check-unfinished-work.sh
    check-cleanup-complete.sh
These checks will be performed BY HAND.
Re-run ./setup.sh global to install them.
```

**This is a report, never a gate.** Each contract's existing hand-run fallback still governs the
call site a missing guard would have covered, and the handoff says those checks were run by hand.
A guard is never skipped for want of the script — this changes only when its absence is noticed,
not what happens next. The block is printed once, at the start of the run; a later call site for
one of the guards it already named performs the check by hand without printing the block again.

**The check covers a named guard's own sibling dependencies too, not only the guard itself.** See
**Guard presence check** (`skills/flow-contracts/pipeline-rationale.md`) for the mechanism and
examples. Derive a
guard's siblings from its own source — grep it for `$SCRIPT_DIR/<name>` and `$(dirname -- "${BASH_SOURCE[0]}")/<name>`, the spelling most `flow-guard` shims load `<agents repo>/scripts/lib/flow-guard.sh` by — rather than trusting a
hardcoded map, exactly as `<agents repo>/scripts/check-guard-symlinks.sh`'s rule 2 already does; a hardcoded list
here would drift from that guard's own dependencies the moment they change.

The guards a command can invoke are derived, never listed: every `<name>.sh` the command's own
skill files and the contracts they load name — the set `<agents repo>/scripts/check-guard-symlinks.sh`'s
rule 2 derives at lint time. Each command cites this section for the block shape.

## Hand-verifying a guard verdict

**Load `skills/flow-contracts/guard-verdict-verification.md`** only when a gate guard's verdict fires — `STAGED-FOREIGN`, `REFUSE`, `OUTSTANDING`, `MOVED` or `LEFTOVER`.

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

`<name>` is **optional** on `/flow` and `/flow-status`.

**The candidate set is `flow state resolve -C <main-checkout>`'s `candidates` — never a
hand-written HTTP call, never a directory listing of your own.** It prints one JSON object:
`"source"` (`"store"` or `"fallback"`), `"complete"` (`true` only for `"source":"store"`),
`"candidates"` (each carrying `name`, `state`, `updatedAt`, `updatedBy`) and `"unreadable"` (the
names of fallback records that could not be read). A `FINISHED` change is never a candidate, and
neither is one already under `<project>/spectre/changes/archive/` whose record is not
`IN_PROGRESS` — a merged change carries its archive while it waits on run 2.

**A record `flow state resolve` marks `"unreadable":true` is reported and skipped from the union — never
silently dropped.** Name the unreadable file in the resolution's own output; do not fold it into a
"zero matches" or "no change" result as if it were never there.

**Every command that resolves this candidate set reports which of the two sources produced it** —
`flow state resolve`'s own `"source"` field, echoed rather than re-derived.

Once the candidate set is built:

- Exactly one match → use it automatically; announce which change was picked.
- Multiple matches → **AskUserQuestion** listing each (name, state, last modified) — never guess.
- Zero matches → fall back to that command's normal "no change" handling (e.g. a creating `/flow`
  run asks what to build; `/flow-status` reports no open changes).

A change linked to a Jira issue is named `<lowercased-key>-<slug>` — see
**Change naming** in `skills/flow-contracts/jira-integration.md`.

## A change's directory

A change's directory is `<project>/spectre/changes/<name>/` while that exists, else
`<project>/spectre/changes/archive/<name>/`. Integrate's run 1 archives the change on its branch
before the route (**Archive on the change branch**, `skills/flow-contracts/finish-contract-run1.md`),
so a fix run after a run 1 that stopped at a pull request or a manual route finds the change
archived, and writes into the archived directory — its appended tasks, its narrative and its
planning commits. Every step that reads or writes a change's planning artifacts resolves the
directory this way; the guards resolve it through `changePlanDir`
(`<agents repo>/stats/internal/guard/changeplan.go`) and its bash twin
`<agents repo>/scripts/lib/change-plan.sh`, and `flow tasks` through the same rule.
