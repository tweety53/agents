# myflow pipeline

The three-state pipeline itself: state definitions, the command→state transition table and the
handoff shape. See **Finish contract** (`skills/flow-contracts/finish-contract.md`).

**Load this file when running `/flow`.** It is split out of
`rules/myflow-manual-review.mdc` so the always-on rule layer carries only the trigger, not the
whole state machine — the same reason the other contract files beside it were split out.

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
/flow  (no state)          → STARTED → IN_PROGRESS   you: review the staged diff and run the apps
/flow  <fix instructions>  → IN_PROGRESS (unchanged)  you: review the staged diff and run the apps
/flow  (bare, IN_PROGRESS) → IN_PROGRESS or FINISHED  terminal only on the merge-and-push route — see the finish contract
```

**Each command ends in the state named after it**, so the state vocabulary and the command
vocabulary are the same words.

**The human gate is a property of the state, not a separate stage.** `IN_PROGRESS` *means* a
staged diff is waiting for the human to review, alongside the run instructions the handoff
printed. Nothing records that the review or the testing happened; the operator running the next
command is what carries the change forward. This is why no `*-done` command exists — there would
be nothing for one to write.

| State | Means | Waiting on |
|-------|-------|-----------|
| `STARTED` | The proposal exists and is published | you — read the artifact |
| `IN_PROGRESS` | The implementation is staged and the handoff printed the run instructions | you — review the diff, run the apps |
| `FINISHED` | Archived, pushed, worktrees removed | — |

**Reviewing and testing are one gate.** A creating or fix run of `/flow` produces both surfaces in
the same run, so the human does both at one sitting. There is no state between implementation and
finishing — integration is not a stage, it is the first half of finishing.

## Stage exit — never the command's own judgment

Within a single run, a stage that loops — most concretely `/flow`'s brainstorm
stage, whose convergence test reopens after every planning-stage exchange that leaves a question the
command's inputs do not answer — never closes on the command's own judgment. It closes only on an
explicit operator answer: at a confirm, or by declining an offer, recording what is still open
rather than assuming it away. The one bounded exception is a session that cannot ask at all: it
records the confirm itself as an open question and ends the stage there, since no operator answer
could ever arrive through it. An operator who is present but silent is not that exception and still
gets another round.

See **Convergence** (`skills/flow/brainstorm.md`).

## Command surface

One command, `/flow`, drives the whole pipeline, plus two read-only commands (`/flow-status`,
`/flow-research`) and one standalone, non-pipeline command (`/flow-settings`). **No command accepts
a flag.** The only argument is the optional change name — see **Change name resolution**.

An argument that is not a known change name is **reported**, not silently ignored — a silently
ignored word is indistinguishable from a flag that stopped working.

## State transitions

| Command | Accepts | Ends at |
|---------|---------|---------|
| `/flow` | *(none — creates the change)* · `STARTED` · `IN_PROGRESS` | `STARTED` from a creating run that stops early, or `IN_PROGRESS` in the same invocation; from `STARTED`, resumes and ends at `STARTED` or `IN_PROGRESS`; from `IN_PROGRESS` with an argument, unchanged (fix run); from a bare invocation at `IN_PROGRESS`, `IN_PROGRESS` or `FINISHED` depending on the route chosen — see `skills/flow/SKILL.md`'s own **State transitions** and **Stage keys** for the full detail this row summarises |
| `/flow-status` | any — read-only, never block | unchanged |

**This table is authoritative.** Every command file — in **both** command trees (`commands/` and
`commands-claude/`) — must state exactly the states its row lists, and must agree with the skill it
delegates to. When a command and its skill disagree, whichever the agent reads first wins, which is
non-determinism in the one layer that must be deterministic.

### Every invocation is re-entrant

Re-invoking `/flow` is the supported way to revise its output. There is no separate `*-fix`
command:

- **At `STARTED`**, `/flow` resumes the creating run from wherever brainstorming stopped.
- **At `IN_PROGRESS`, with an argument**, `/flow` resumes the existing worktree and applies a fix,
  documenting it in `proposal.md`/`tasks.md` or a `<name>-fix-N` sub-change first, and refreshing
  the test guide alongside the code so the two surfaces never drift apart.
- **At `IN_PROGRESS`, bare**, `/flow` integrates on the first such invocation and archives on the
  next, once the branch has merged.

### A fix never moves the state

Implementation advances the state **only** from `STARTED` to `IN_PROGRESS`. A fix run at
`IN_PROGRESS` writes the state back exactly as it found it.

No field records where a fix was raised. Whether the human re-reviews the diff or re-runs the apps
after a fix is their decision.

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
**Suggested instead:** <what to do instead, e.g. "nothing — the change is archived">
```

## Progress visibility

**`/flow` drives the harness's task-list mechanism.** It registers its steps with it at the start of
a run and keeps each entry's status current — in progress when its step begins, completed when
that step finishes — so the harness's live progress view, a count line and one line per task,
renders throughout the run rather than arriving with the handoff. The count line then distinguishes
done, in progress and open at every point. One entry per whichever cited stage is running at the
time, at that stage's own granularity — brainstorming checklist items and artifacts on the
creating/resuming branch, `tasks.md` items on the implementation branch, a finish run's steps on
the integrate/archive branch.

`/flow-status` is read-only and **registers nothing**. Registering steps for a
report would put entries on the operator's task list for work nobody is doing.

**The progress view is a view, never a record.** No command, guard or contract reads the harness's
task list back as evidence of what was done. `tasks.md` remains the single source of truth for a
plan's completion state, and `<agents repo>/scripts/check-unfinished-work.sh` reads that file — a second source of
completion state would be one that guard cannot see.

**No third checkbox marker is added to `tasks.md`** to carry an in-progress state; the in-progress
count comes from the harness's task list alone, which no run persists. See **Progress visibility**
(`skills/flow-contracts/pipeline-rationale.md`) for why a marker would be unsafe.

**Stated against the mechanism, never against one harness's tool.** Where a harness offers no
task-list mechanism, the command prints the equivalent block instead — a count line naming how many
steps are done, in progress and open, followed by one line per step marked done or not done — and no
harness has to gain a task tool to satisfy the rule. See **Progress visibility**
(`skills/flow-contracts/pipeline-rationale.md`) for why the rule is stated against the mechanism.

## Stage marks

`/flow` marks each of its own stages: `myflow stage begin` when the stage starts,
`myflow stage end` when it closes, both naming the command, the stage and the change. The stage
identifier is the **key**, never the prose name, from **Level 1 — the stages of each command**
(`<agents repo>/README.md`) — that table is restated nowhere here, on purpose: a second copy is exactly what its
own README-parsing test (`<agents repo>/stats/internal/stages/names_test.go`) exists to make impossible. Each
skill's own mark calls sit at that skill's own stage boundaries, in that skill's own `SKILL.md`, and
name the stage by its Key column exactly — the CLI compares the `-stage` value byte for byte against
the documented key and rejects anything else, naming the documented alternatives.

**A `stage begin` mark MUST carry `-session-token` and `-harness`.** Neither is optional, and a call
missing either is a caller mistake the CLI rejects before it ever reaches the store (see below).
`-session-token` is a literal, unique token this **run** — not this mark — generates once, at the
start of the run, and then passes unchanged on every mark that run makes, so the daemon can later
find it in the calling session's own transcript and bind every stage run carrying it to that
transcript's own `sessionId` — design.md's "bind after the fact, by a correlator the caller writes"
and "one token per session, not one per mark" (kan-172). Generate the token once, near the start of
the run, before the first `stage begin`, and reuse that exact value at every later mark site in the
same run; do not invent a fresh one per mark. `-harness` names the harness actually running the mark
— `claude-code`, `cursor` or `codex`.

**The token is per run, generated fresh, never reused across runs.** See **Stage marks**
(`skills/flow-contracts/pipeline-rationale.md`) for why. A run
that starts identifies itself with its own new token; a run that already has one (mid-run, at a
later mark) reuses it.

**The `<change>` argument is always a resolved change name, never a guess.** See **Stage marks**
(`skills/flow-contracts/pipeline-rationale.md`) for what marking writes when it is not.
`<agents repo>/scripts/check-stage-mark-calls.sh` rejects a `stage begin` call site whose change argument is
written as a placeholder naming a guess.

**Neither `-session-token` nor `-harness` is ever a hardcoded value in the skill text: both are
filled in by the agent at call time, from a placeholder — `<literal-token>` and `<harness>` below —
because one skill source installs into `~/.claude/skills/`, `~/.cursor/skills/` and
`~/.codex/skills/` alike, and a hardcoded `-harness claude-code` would mislabel every Cursor and
Codex run as Claude Code, hiding the very thing the field exists to record: that Cursor and Codex
write no transcript, so their runs are *explicitly unavailable* rather than zero.
`<agents repo>/scripts/check-stage-mark-calls.sh` rejects a hardcoded `-harness` literal in skill source the same
way it rejects a substituted session token.

```bash
myflow stage begin -command '/flow' -stage flow.review-panel -harness <harness> -session-token mf-<literal-token> <name>
# … the stage's own work …
myflow stage end   -command '/flow' -stage flow.review-panel -outcome completed <name>
# … a later stage in the same run reuses the same token, not a new one …
myflow stage begin -command '/flow' -stage flow.verify -harness <harness> -session-token mf-<literal-token> <name>
```

**The session token MUST be a literal, written directly into the command — never a shell
substitution.** `-session-token "mf-$(date +%s)-$$"` is rejected, and so is any token carrying a
backtick or a `$VAR` reference. See **Stage marks** (`skills/flow-contracts/pipeline-rationale.md`)
for why. A reader who does not know this will
"improve" the literal into a substitution the first chance they get, which is exactly the regression
this paragraph, `<agents repo>/stats/cmd/myflow/stage.go`'s `validateSessionToken`, `<agents repo>/stats/internal/api/stages.go`'s
`validateSessionTokenShape`, and `<agents repo>/scripts/check-stage-mark-calls.sh` all exist to stop. Write a
concrete token in its place — `<literal-token>` above means "invent a short, unique string right
here, once, and reuse it", not "leave this placeholder in the invocation" and not "invent a new one
at every mark".

**A mark never blocks, delays, or alters the stage it marks.** On any store failure the CLI
journals the intent, prints one warning line, and exits 0 — the same never-block guarantee **State
file** (`skills/flow-contracts/state-file.md`) already states for `state set`. Do not branch on
`myflow stage`'s exit code as a signal about the stage itself: a mark that could not reach the store
still exits 0, so there is nothing to react to, and treating its output as a stage failure would make
the mark exactly the block it is required not to be. The only nonzero exit is a caller mistake — an
undocumented stage key, a missing required flag, or a session token carrying a substitution shape —
which is a defect in the skill's own call, not an outcome of the stage, and is fixed by correcting the
call rather than worked around.

**`stage end` carries no session token and no harness of its own** — attribution happens once, at
`begin`, and the harness recorded there is immutable, so an end mark can never contradict the harness
a stage began under.

`/flow-status` marks nothing — its own row in the Level 1 table says so, and a read-only report
that wrote stage runs would be recording work nobody did.

## Handoff output

Every run ends in the same shape, and prints **nothing** after it:

```
<1–3 lines: what actually happened>

<absolute paths to anything the operator needs to open>

Next:
/flow <name>
```

- **The next command is the last line** — bare, copy-pasteable, with no prose after it. See
  **Handoff output** (`skills/flow-contracts/pipeline-rationale.md`) for why.
- **A bare invocation at `IN_PROGRESS` that opened a PR or handed off manually names itself** as the
  next command, because that is what the operator runs once the branch is merged. Only a run that
  **completed** archive is terminal and names nothing — a run that stopped on a cleanup leftover
  names itself too, for the same reason: the operator clears what remains and runs it again.
- **Only what the operator must act on.** Do not restate the plan, enumerate completed internal
  steps, or repeat content available at a path you just gave.
- **Link, never paste.** Diffs and plans are given as absolute paths.
- **Every path is absolute** — in handoffs, in IntelliJ commands, in run instructions. Never a
  relative path, never `../<other-app>`, and never a main-checkout path while an apply worktree
  holds the work. Resolve app roots from `git worktree list` or the state file's `worktrees` keys.
- **Implementation never stages `<project>/spectre/changes/` or `<project>/docs/superpowers/` before
  integrating**, and the list is fixed here rather than configured per project. `<project>/spectre/specs/`
  is deliberately not on it — a capability spec is implementation, per **Git boundaries**
  (`skills/flow-contracts/git-boundaries.md`). The integrate phase
  stages them and commits them separately from the implementation, so nothing is lost. See
  **Handoff output** (`skills/flow-contracts/pipeline-rationale.md`) for why leaving them unstaged
  — rather than filtering a display — is what keeps them out of every view of the staging area:

  ```bash
  git -C <abs-worktree> diff --cached --stat
  git -C <abs-worktree> diff --cached
  ```

### The block each state renders

The block a state hands off is defined in **The block each state renders**
(`skills/flow-contracts/handoff-blocks.md`). `/flow-status` loads it; `/flow`
carries only the block it prints.

### The tab commands, printed at the start of a run

`/flow` prints, immediately after its announcement
line and before any work, two commands for the operator to paste:

```text
/rename <change-name>
/color cyan
```

**They sit at the start of the run, not in the handoff block, and the colour is one fixed value —
`cyan` — for every command and change, signifying only that a pipeline command owns the tab.**
`/flow-status` prints neither line: a read-only report does not own the tab. See **The tab
commands, printed at the start of a run** (`skills/flow-contracts/pipeline-rationale.md`) for why
cyan was chosen and why the lines are printed at the start rather than the end.

**They are printed rather than invoked because neither is reachable from inside a run.** See **The
tab commands, printed at the start of a run**
(`skills/flow-contracts/pipeline-rationale.md`) for the measurement behind it.

- **Where a harness offers a reachable way to set the tab's name and colour from inside a run**, the
  command may use it, and then prints nothing — the lines exist to be pasted, and there is nothing
  to paste once the thing is done.
- **Where it does not** — Claude Code today, for the two reasons measured, and any harness
  with no tab concept at all — the command prints the two lines. In a harness with no such commands
  they are inert text the operator ignores, which costs two lines and leaves nothing broken.

The rule is satisfied by whichever mechanism the harness provides, and no harness has to gain a tab
API to satisfy it. What is **not** optional is that the naming happens at the start of the run: a
command that silently skips it because its harness offers no tool has dropped the requirement, not
adapted it.

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

**No length guard and no byte budget measures a change artifact**, and none is to be added: a budget
on a per-change artifact rewards dropping the facts the paragraph above requires kept. Brevity
here is a judgment the review panel and the operator make, never a number a script checks.

See **Artifact brevity** (`skills/flow-contracts/pipeline-rationale.md`) for why this is stated
here rather than in each artifact-writing skill.

## IntelliJ commands

Every state that waits on a human must print a copy-paste command in its handoff:

```bash
open -na "IntelliJ IDEA" --args "<absolute path>"
```

Use `open -na`, not the `idea` shim — that shim is not on this machine's PATH. See
**IntelliJ commands** (`skills/flow-contracts/pipeline-rationale.md`) for what `open` buys.

| State | Path to open |
|-------|--------------|
| `STARTED` | main checkout (artifacts live there; no worktree exists yet) |
| `IN_PROGRESS` | apply worktree root |

Paths are absolute, resolved from `git worktree list`. Never emit a relative path.

## Guard resolution

**A named guard resolves to `<skill-dir>/scripts/<name>`.**
Skills and contracts name a guard by **basename**, never by a path relative to a repository
root — such a path resolves only when the project being worked on *is* the agents repository,
which is the one case that is never the interesting one.

`skills/flow-contracts/` is never a running command and carries no
`skills/flow-contracts/scripts/` directory. See **Guard resolution**
(`skills/flow-contracts/pipeline-rationale.md`) for what resolving against the running command's
own skill directory buys.

**Prose describing this repository's own guard is not an invocation.** A guard invoked by name
uses the basename form above. Prose that describes **this repository's own** lint and test
guards — resolved through `<agents repo>/.flow/project.md`'s `## lint` and `## test` lists
rather than through `<skill-dir>/scripts/` — names the guard as
`<agents repo>/scripts/<name>` instead of a bare repository-relative path: a bare path there
resolves, for a reader standing in an installed project, against that project's own tree, so the
sentence would name a file the reader may be able to write. See **Guard resolution**
(`skills/flow-contracts/pipeline-rationale.md`) for why carrying the prefix matters.

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
guard's siblings from its own source — grep it for `$SCRIPT_DIR/<name>` — rather than trusting a
hardcoded map, exactly as `<agents repo>/scripts/check-guard-symlinks.sh`'s rule 2 already does; a hardcoded list
here would drift from that guard's own dependencies the moment they change.

Each command names, in its own text, the guards *it* can invoke — exactly the guards
`<skill-dir>/scripts/` carries — and cites this section for the block shape.

## Finish contract

**Finish contract** (`skills/flow-contracts/finish-contract.md`) governs the preflight signals,
both runs' procedures, base-branch resolution and worktree cleanup, and `/flow`'s integrate/archive
phase is the only phase that loads it.

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

## Change name resolution (all `/myflow-*` commands)

`<name>` is **optional** on `/flow` and `/flow-status`. When omitted, the candidate set is built
from the store when the daemon answers, and from the filesystem only when it does not — and the
command building it says which of the two produced the set, per **State file**
(`skills/flow-contracts/state-file.md`).

**The store is reached first, through the CLI — `myflow state list [-C dir]`, never a hand-written
HTTP call.** `state list` enumerates the store on the caller's behalf (`GET
/api/v1/stats/state-board` under the hood, over a period wide enough to cover every change ever
recorded, since the store starts empty per **State file**) and prints one JSON object:
`"source"` (`"store"` or `"fallback"`), `"complete"` (`true` only for `"source":"store"`), and
`"records"` (each carrying `name`, `state`, `updatedAt`, `updatedBy`). See
**Change name resolution (all `/myflow-*` commands)** (`skills/flow-contracts/pipeline-rationale.md`)
for why this goes through the CLI rather than a skill calling `curl` directly.

**When `"source":"store"`**, the candidate set is every record's `name`, dropping one whose `state`
is `FINISHED` — **States** above already defines `FINISHED` as archived, so this is the store-backed
form of the same exclusion.

**When `"source":"fallback"`** — `state list` could not reach the daemon and instead scanned the
local on-disk fallback directory — the candidate set falls back to the union of two filesystem
sources:

- the non-archived ids `spectre list --json` reports, read from each entry's `id` field in its
  `{"changes":[{"id","done","total"}]}` output; and
- the names `state list`'s own `"records"` carries in this mode — the basenames of whatever is
  directly under the project's state directory, `/Users/tweety53/Agents/flow/state/<project-key>/`
  — which, per **State file**'s "The store starts empty", now holds only the CLI's on-disk fallback
  records, never a second live source. A name found only here is one whose last write could not
  reach the store.

From that union, drop any name whose `<project>/spectre/changes/<name>/` directory has already reached
`<project>/spectre/changes/archive/`. The state directory is per-project rather than per-worktree, so a
fallback record is reachable from the main checkout regardless of which worktree wrote it. See
**Change name resolution (all `/myflow-*` commands)** (`skills/flow-contracts/pipeline-rationale.md`)
for why the filesystem source is needed at all now that the store is the normal path.

**A record `state list` marks `"unreadable":true` is reported and skipped from the union — never
silently dropped.** Name the unreadable file in the resolution's own output; do not fold it into a
"zero matches" or "no change" result as if it were never there.

**Every command that resolves this candidate set reports which of the two sources produced it** —
`state list`'s own `"source"` field, echoed rather than re-derived. A
report built from the filesystem fallback during an outage must never be presented the way a report
built from the live store is — that is precisely the silent-stale-data outcome this resolution
exists to prevent.

Once the candidate set is built, resolution proceeds exactly as before:

- Exactly one match → use it automatically; announce which change was picked.
- Multiple matches → **AskUserQuestion** listing each (name, state, last modified) — never guess.
- Zero matches → fall back to that command's normal "no change" handling (e.g. a creating `/flow`
  run asks what to build; `/flow-status` reports no open changes).

This resolution is defined **once, here**, and `/flow`'s own enumeration step
cites this section rather than repeating or re-deriving the union — so its list of open
changes can never drift from what this section defines.

A change linked to a Jira issue is named `<lowercased-key>-<slug>` — see
**Change naming** in `skills/flow-contracts/jira-integration.md`.
