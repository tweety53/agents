# myflow-do — rationale

This file is the reasoning behind `skills/myflow-do/SKILL.md`.
**A `/myflow-*` run never loads it — appendices are for whoever edits a contract.**

## State gate

## Superpowers Basic Workflow

## 1. Load context and validate the plan

## 2. Isolate the workspace (first run only)

Why the merge base and path go into working notes rather than the state file at this point: they are
what section 7's workspace-isolation guard needs, and this run knows them from here on regardless of
what the on-disk state file says until section 7's own write.

Why the resolved worktree set is non-empty by construction on every ordinary run, first or fix
alike: the worktree is known the moment it is created or resumed above, well before section 7's
empty-set stop could ever fire on the ordinary shape of a first run.

Why the workspace id is computed fresh in this step, rather than carried over from an earlier one:

Compute it once per run, on a fix
run exactly as on the first: the derivation is deterministic, so a later run reproduces the same id
rather than reading one back, which is why nothing about it is written to the state file. Two later
steps consume that one value — section 6 resolves the run instructions' URLs from it, and section 7
resolves the project's declared isolation rows against it — so an id derived twice in one run is two
chances to disagree.

## 3. Documenting a fix, before implementing it

Appending is recommended because most fixes are corrections within the change's existing scope, and
a sub-change per fix round buys another change directory the operator has to read back. A
sub-change is never archived alone — it is archived with its parent, by its own call.

## 4. Execute (SDD + TDD)

## 5. The review panel

Why no preset moves the handoff bar: a preset able to lower it would be a way to skip review, not a
way to size the panel's reading.

Why the containment rule on `[STANDARDS_PATHS]` is non-negotiable: it stands between an
attacker-editable list in a tracked file (`<project>/.myflow/project.md`'s `## standards` entries) and an
arbitrary file read landing in a committed review record.

Free prose is not a record of a finding's state: a state that cannot be counted cannot be enforced.

Why a `withdrawn` marker's reason is checked for being there at all: a fix subagent rewriting
`open` to `withdrawn` would be closing its own gate.

Why panel dispatches never inherit a model from the parent session:

There is no
parent-model inheritance and no economy tier — the panel's cost must not depend on which model the
operator happens to be running, and a recorded value is a deliberate decision for one change rather
than an inheritance path.

Why a reviewer given a repo-relative path can end up with no principle list at all:

The subagent's working directory is
the project worktree, which has no `skills/` tree, so a repo-relative path opens nothing and the
reviewer runs with no principle list.

`<agents repo>/scripts/check-unfinished-work.sh` reads **only the marker block**. It never parses the table: not
its header, not its column order, not its cell boundaries, not where it starts or stops. So an
unescaped `|` inside a cell is just text, a reordered header changes nothing, and a row that lost a
boundary pipe still counts. That is the point of the split — the previous shape asked a hand-rolled
table parser to recover one fact from a grammar defined in prose, and it failed **open** six
distinct ways across three review passes before it was replaced.

- `<status>` is **exactly** `open`, `fixed` or `withdrawn`, compared byte for byte. `Open`,
  `WITHDRAWN`, `open (needs discussion)`, an empty value and a value carrying an invisible character
  are none of the three, and none of them reads as closed.
- `withdrawn` **carries its reason on the same line** — the reason is part of the state, not a note
  about it.
- Each `F<n>` names **one** finding. A reused identifier is reported on each side separately, so two
  distinct findings labelled `F1` in both the table and the marker block cannot cancel out — that
  shape hid an open Critical, with the word `open` never appearing in a marker at all.
- The marker lines sit on **consecutive lines**, one unbroken block. This is what stops a marker
  quoted elsewhere — inside a fenced example, say — standing in for a marker that was never written,
  which is the one route that still under-counted when the redesign was attacked.
- `findings-total: <n>` appears **exactly once** and equals the number of marker lines. A record
  with no total line is outstanding however clean it reads: zero findings is not something to infer
  from silence. A panel that raised nothing says `findings-total: 0` and carries no markers.

**The table carries no status column, on purpose.** A finding's state is written once, on its
marker line. A status cell beside the marker is a second surface that can silently disagree with the
line that governs: the machine's direction is protected — a marker reading `open` blocks whatever a
cell says — but nothing protects a reader who sees `fixed` in the table and believes it. State the
fact once. To read a finding's state, look up its `F<n>` in the marker block.

Why identity outranks the interval: a dispatch's `started_at`/`ended_at` are typed by the
dispatching agent at mark time, not read from any clock the harness itself keeps, so the window is
approximate by construction — the audit that motivated this capability found windows rounded to the
minute, `02:00:00` and `00:10:00`. The identifier the harness returns is exact, and it alone tells
two concurrent dispatches apart once their windows overlap. That audit of the live store, taken
2026-08-23, found 57 dispatch rows across 5 changes, only 20 carrying any token figure, and
`agent_id` NULL on all 46 rows this repository itself had recorded — the panel's concurrent slots
collapsing their whole group's cost onto whichever slot's window closed last.
<!-- measured: live-store audit, 2026-08-23; see kan-212 proposal.md -->

### No forking, and a wall-clock ceiling on every slot

- Why 15 minutes: the observed good run returned in 2.5 minutes, the harness's stall watchdog is
  600s and did not fire, and the hung slot ran 4 hours.
<!-- measured: KAN-295 panel rounds 3-4 and the re-dispatch after round 4; see proposal.md -->
- Why one automatic retry before asking: the ticket's own evidence is a re-dispatch returning in 2.5
  minutes, so a retry usually clears it, and an operator interrupt for that is a worse trade.
<!-- measured: KAN-295, the re-dispatch that followed round 4 -->
- Why the ceiling is not a guard script: a script has no handle on an in-flight subagent and could
  only read a timestamp the agent itself writes, which is the agent enforcing it with extra steps.
- Why not the harness's own stall watchdog: round 4 is the evidence it does not fire on a slot
  emitting output, and myflow does not configure three harnesses' watchdogs.
- Why panel dispatches only: implementers are serialised per worktree and report a commit sha, so a
  hang there is visible by a different route, and widening the rule to them would be speculative.

### Code review (low)

Why the light preset's third slot was the harness's built-in review skill rather than a narrowed
Superpowers reviewer, until KAN-295: a narrowed Superpowers reviewer is the same reviewer slot 0
already runs, so the panel would carry two readings from one reviewer rather than a genuinely
different check. The harness's skill was a distinct lens the panel did not otherwise get, which is
what a lighter roster was trading Bugbot's defect hunt for.

What KAN-295 found: the skill forks its own background agent, and the parent panel dispatcher never
observes that inner agent's completion, so the slot cannot report through the panel's contract.
Observed three times:

- **Round 3** — the inner agent bypassed the slot, returning its findings *directly to the session*.
  The slot itself stalled and was killed by the harness's stall watchdog with no result.
  It had run 39 minutes; the findings it delivered numbered 6; the watchdog's threshold is 600s.
<!-- measured: KAN-295 panel round 3 -->
- **Round 4** — the slot ran with no result and the stall watchdog never fired, because the slot
  kept emitting output while making no progress, so it sat alive and useless and the notification
  the dispatcher was waiting on never arrived. It was caught only because the operator asked
  whether it was stuck. Elapsed: **4 hours**.
<!-- measured: KAN-295 panel round 4 -->
- **Re-dispatched on the documented fallback** — a plain `general-purpose` reviewer doing the same
  checks, no skill — it returned `PANEL-SLOT-CLEAN` in **2.5 minutes**.
<!-- measured: KAN-295, the re-dispatch that followed round 4 -->

Why the concern that motivated the skill — that a narrowed Superpowers reviewer duplicates slot 0's
reading — did not survive: a duplicated reading that returns beats a distinct lens that does not.

### Bugbot's mutation-testing brief

Why the brief rides on the dispatch prompt rather than on Bugbot's own definition: Bugbot carries
its own agent definition, which the dispatcher does not edit, so the prompt is the only lever this
command has over what Bugbot is told to do.

Why a surviving mutant is an ordinary finding rather than an advisory note: an advisory class would
be a second kind of finding that nothing enforces, sitting beside a bar — zero open findings at any
severity — that enforces every other one. A finding class the guard does not check is a finding
class an operator can silently ignore.

### Optional slot selection

A reviewer too many costs tokens; one too
few costs a defect.

### Panel re-runs

Why **trigger-based** scoping reaches conditional slots alone: a conditional slot's region is
exactly its trigger's subject, so a trigger that no longer fires states positively that the region
is unchanged; a required slot has no trigger, and so no subject to re-evaluate. The empty-delta
skip is a second mechanism rather than the first one widened because it asks a different question
of a different fact — not *did this slot's subject change*, but *has anything changed since this
slot last read* — and the record keeps a wording for each, so that a panel record naming one says
which of the two questions was asked.

Why a delta is anchored at the slot's own last read rather than at the current round: for a slot
reaching a Full pass having missed rounds in between, `fix-round-N.diff` would withhold those rounds
silently, against *targeting is a cost optimization, never a coverage waiver*. What the anchor buys
is measured: KAN-315 ran twelve rounds, and three slots re-read a diff that reached 3,300 lines
every round while each round's actual subject was a 200-to-900-line delta.
<!-- measured: KAN-315's self-review, as reported in KAN-327's description -->

Why the range is tree-to-tree rather than `A..B`: an ancestry-dependent range is orphaned by the
same `git rebase --autosquash` a tree-to-tree range survives, and stops naming the change between
the two trees without erroring — the slot is handed a diff, just not the one the rule asked for.

Why Primary alone keeps the whole `final-review.diff`: its remit is plan alignment and the change's
history, and a delta carries neither — it shows what moved since one sha, never whether the change
as a whole still answers the plan it was written from. Principles is the near miss, and goes delta
anyway: it owns hard invariants, but judges them against the code it is shown, and pass 1 already
showed it everything.

Why the base is stored on the dispatch row and never read back: every `myflow record` write is on
the never-block guarantee — an unreachable store journals the intent and the run proceeds — which
holds only while nothing depends on the write having landed. A panel runs inside one `/myflow-do`
invocation and a re-run starts a fresh pass 1, so the dispatcher already holds each slot's
last-reviewed sha in session; adding a read path would convert a write that is allowed to fail into
one that is not.

Why checking the path alone was not enough: an argument such as `../../../etc/passwd` was free to
escape the worktree while the path token itself stayed clean.

Why an embedded newline needs no separate ban: the anchored `finding-reproducer:` marker line this
guard reads is one line by construction, so it could never carry one.

Why the 20-second bound is not `check-cleanup-complete.sh`'s 60-second `SURVIVORS_TIMEOUT`: a
reproducer is one short-lived check under test, not a wait for a whole cleanup pass to finish. Why a
killed reproducer's exit status is never read as pass or fail: this mirrors `run_survivors`'s own
`WS_TIMED_OUT` flag, kept apart from the exit code so a killed process is never mistaken for an
answer.

Why the post-fix re-run needs the materiality condition on top of the symmetric exit-0 check: a
symmetric re-run alone cannot tell "the production code changed" from "the reproducer's own target
changed," and membership alone cannot tell a real fix from a comment or whitespace edit on a named
path.

### The fix round mutation-proves what it changed

**Why the parent runs the mutations.** KAN-200's fix round 2 was not careless. It mutation-proved
eight added harness cases, eight for eight, reported individually. It still shipped its own largest
structural change — a record protocol's delimiter — unproved, and reverting that protocol left all
eighteen harness cases green, because the harness contained no literal tab byte and no `\037` byte
anywhere. The gap was found in round 3, by a reviewer who thought to ask whether the previous round
had held itself to its own standard. A rule that asks the same actor to widen its own scope is
asking for the judgment that already failed. `/myflow-do` settled a related question one paragraph
earlier, where the parent re-runs each reproducer rather than accepting the subagent's account of
its own success — related, not identical: a reproducer is a re-runnable command sitting behind two
guards (`check-panel-reproducers.sh`'s shape check and the parent's own re-run), while a
`fix-mutation:` line is prose the parent itself writes and no guard reads, so it could be
fabricated undetectably in a way a reproducer's flip cannot.

**Why no guard reads the record.** To be more than a nag, a guard would have to decide from a diff
whether a round changed executable behaviour or edited a comment. A script cannot make that
classification, and in a repository that is mostly prose it fails in the direction that matters: it
fires on every prose fix round until someone narrows it into vacuity. KAN-197 examined and rejected
two mechanisms of exactly this shape.

This rejects only the diff-classifying guard. The cheaper variant `check-panel-reproducers.sh`
already precedents — a shape-and-count check that `fix-mutations-total:` matches the number of
`fix-mutation:` lines and that each is well formed — was considered and is not adopted either, for
a reason specific to this record: a reproducer's shape check works because the shape it counts is
itself evidence — a well-formed `finding-reproducer:` line names a command the parent then runs, so
a malformed one is caught before it can hide anything. A `fix-mutation:` line names no command;
counting well-formed ones only proves the parent typed a well-formed sentence, never that a mutation
ran or that a test failed. The shape check would certify the report's grammar, not its truth, which
is the thing that matters here.

A third guard shape is different from both of the above: a grep-based lint flagging a reserved
marker label — `finding-status:`, `findings-total:`, `finding-reproducer:` — appearing outside its
own marker position, the family `<agents repo>/scripts/check-unfinished-work.sh` and
`<agents repo>/scripts/check-panel-reproducers.sh` already form. The evidence for it is real, and is stated
plainly rather than minimised: this exact defect class was independently rediscovered three times
in this change's own review — F2, F9 and F11, in three different scopes across three rounds — and
each time the fix was more prose rather than a check. That recurrence is a genuine argument for
mechanising it. It is nonetheless out of scope here: it is a guard over the record's
well-formedness, not over mutation-proof, which is the guard the operator declined above, and it
belongs beside the record-format guards it would extend rather than inside a rule about fix-round
behaviour. It is carried forward as a follow-up candidate rather than dropped.

What the record buys instead is that the question becomes askable at all — a recorded field turns a
lucky unprompted question into a line the next round's reviewers read.

## 6. Resolve the run instructions

In the same run, resolve the run instructions for the handoff. This is why reviewing and testing
are one gate: the diff and the way to run the change are produced together and can never drift
apart.

Nothing reads the run instructions back — no later command, no guard, no fix round consults a
recorded copy of them — so writing them to a file would be a record with no consumer. Printing them
into the handoff instead means they are only ever as stale as the run that just produced them, and
there is no second copy to fall out of sync with the worktree it describes.

A worktree's applications bind their own ports, so a URL taken from the project's declared base
instead of this worktree's resolved one reaches whichever workspace holds the default port — a
different change's application, answering plausibly and about the wrong work.

## 7. Verify, stage, and hand off

**This is the only place a project's declaration is validated, and that is why it happens here.**
The rules are mechanical, so re-deriving them by reading rows is the failure the guard exists to
remove; and the guard ships in the agents repository while the projects it judges do not, so a lint
list reaches one repository and this command reaches all of them. `/myflow-finish` does not repeat
it: run 2 reads the `survivors` row alone, and every input it cannot resolve is already a reported
skip under **Run 2 — the branch is merged** (`skills/myflow-contracts/finish-contract.md`). Adding a
blocking validation there would strand an already-merged change over text nothing in that session
can correct — the trade that **Creation and cleanup**
(`skills/myflow-contracts/workspace-isolation.md`) rejects when it weighs a change stranded short of
its terminal state against stale storage.

The script-absent case and the guard-exit-2 case are two distinct triggers that land in the same
outcome, and here is why that is deliberate rather than an oversight: both end with nothing having
run the guard, and a session that recognised only "the repository does not carry it" would answer a
failed resolution by doing nothing at all and saying nothing about it.

- **Every declared row, not a subset.** A value this run resolved and did not export is a value
  nothing reads, so the workspace is isolated on paper while the applications and their checks still
  reach the project's shared resource — the same silent wrong answer as having derived nothing.

Creating a workspace database and bucket for a run that
only ever linted would leave behind resources nobody asked for and only `/myflow-finish` run 2
removes.

Why the two planning paths are left unstaged rather than filtered out of a display:

The exclusion is what keeps them out of the diff, rather than a filter applied when the diff is
displayed: a filtered display leaves them in the staging area, where the IDE's staged-changes pane
and `git status` show them again. The list is fixed — the pipeline chooses these paths itself, so no
project can differ. `/myflow-finish` stages and commits them separately, so nothing is lost by
leaving them unstaged here.

**The `reset` is what enforces the rule; without it the `add` only assumes it** — the reason is
stated once under **Git boundaries** (`skills/myflow-contracts/git-boundaries.md`). What is specific to
this command is *whose* staging it retracts (an implementer subagent's own `git add`, or a worktree
resumed with a dirty index) and why `git reset -- <paths>` is the tool:
it touches the index only, restores a tracked path to its `HEAD` entry instead of staging a deletion
the way `git rm --cached` would, and succeeds when a path is absent — which `<project>/docs/superpowers/` is
on every run that has rendered no record into it yet, and where `git restore --staged` would refuse
the whole command and unstage nothing.

The outcome table under
**Rendering the session records** (`skills/myflow-contracts/session-records.md`) is canonical for all four
outcomes.

The template's third git state — committed and pushed with no PR — is one `/myflow-do` never emits
and `/myflow-status` does; the pairing is canonical under **The block each state renders**
(`skills/myflow-contracts/handoff-blocks.md`).

The block below is **not** a second definition of the handoff. It is this command's rendering of the
`IN_PROGRESS`-after-`/myflow-do` template, which is defined once under
**The block each state renders** (`skills/myflow-contracts/handoff-blocks.md`) and is canonical for the
labels, the field set and their order. What this block adds is the enumeration of the literal
alternatives `/myflow-do` writes. **Change the template first and bring this block with it** — a
field added here and not there is drift the moment `/myflow-status <name>` regenerates the same
state.

That ordering is what
makes a fix round raised after a PR is open refresh the rendered records rather than leave them a
round stale.

## Guardrails
