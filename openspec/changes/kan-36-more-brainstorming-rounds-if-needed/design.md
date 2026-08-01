# Design — kan-36-more-brainstorming-rounds-if-needed

The full brainstorming record is
`docs/superpowers/specs/2026-08-01-kan-36-more-brainstorming-rounds-if-needed-design.md`. This file
is the change's own design: the mechanisms, where each rule lives, and the decisions taken.

## Four slices, three subjects

| Slice | What it does | Capability |
|-------|--------------|------------|
| A | Brainstorming iterates to convergence, bounded by two prompts | `myflow-planning-gate` |
| B | What survives the last round is recorded and counted at the gate | `myflow-planning-gate`, `myflow-handoff-output` |
| C | `TO DO URGENT` is read as `To Do` | `myflow-jira-projection` |
| D | `/myflow-start` stages its planning artifacts and never commits them | *(none — complies with the unmodified `myflow-command-surface` requirement)* |

A and B are the issue's ask, split because the loop and the record are separable — the loop produces
more rounds, the record is what makes stopping early honest. C and D were raised by the operator at
gates this change is about: C at the unrecognised-status carve-out that fired for KAN-36 itself, D at
the spec-review gate.

## Slice A — one convergence test, applied after every exchange

The test is a single question the command asks itself after a question round, a design-section
approval, or the spec review: **does it now hold a question its inputs do not answer?** While the
answer is yes, another round opens.

**One test rather than a rule per gate is the design choice.** A design section that reopens
something and an answer that opens something are the same event — the command ends up holding an
unanswered question — so a gate added later inherits the loop instead of escaping it by not being
enumerated. Three separate rules would have to be extended by hand, and the one nobody extends is the
hole.

### Two prompts, recommending opposite courses

| Prompt | Fires when | Recommended option | Why that is honest |
|--------|-----------|--------------------|--------------------|
| **The confirm** | the test comes back empty | move on | unreachable while the command holds anything |
| **The soft-bound offer** | a round would open at or beyond the threshold | another round | reachable only while it genuinely holds questions |

The pair is the point, and it is stated in the spec so a later editor does not "harmonise" them.
Each is reachable only under the fact that makes its own recommendation true. The offer's shape is
the one `/myflow-finish` run 1's unfinished-work gate already uses, where **Stop** is recommended
because the gate only fires when something really is unfinished.

### Where the threshold lives

`3` is a **tuned value** and belongs in `skills/myflow-start/SKILL.md`; the requirement says only
that later rounds are offered rather than taken. This is the structure-here / thresholds-cited split
`pipeline.md`'s level-2 expansions already use, and it is why the *Brainstorm* expansion gains the
loop's shape without gaining its number.

Rounds one and two open silently because an initial round plus the round its answers open is the
ordinary shape of a converging conversation; prompting inside it taxes the common case. There is no
hard cap: no round count ends the stage on the command's own authority.

## Slice B — the record, and the count at the gate

`design.md` gains `## Open questions`, beside `## Decisions` and shaped like it, with an **immutable
ID**. A later round that answers a question sets `**Status:** answered by <decision-id>` and adds the
decision; the entry is never deleted or rewritten. A stage that left nothing open records none.

The `STARTED` handoff gains `**Open questions:** <N> | none` immediately after `Decisions recorded`,
counted from entries still at `open`. Because it is derived from `design.md` on disk it is
**regenerable, not run-only** — `/myflow-status <name>` renders it from the same template, and a
count that changes when a revision round answers a question is the correct behaviour rather than a
drift.

`skills/myflow-status/SKILL.md` needs no edit: it renders the block by citing the template rather
than carrying a copy, so the field flows through. That is the citation-over-copy property working as
designed, and it is worth naming because it is the reason a two-renderer field costs one edit.

## Slice C — an enumerated name, not an inference

The four ordered names become four **positions** carrying names, with `TO DO URGENT` mapped onto
`To Do`. The unrecognised-status ask survives, narrowed to a status matching no mapped name.

Two properties are preserved deliberately, and both are stated in the delta spec so a reader does not
have to reconstruct them:

- **kan-26's rejection of `statusCategory` derivation stands.** Enumerating a name states a position;
  deriving from the category deduces one, and the deduction is what would put `TO DO URGENT` at
  In Progress and freeze the board. This change does the former, which is the mechanism the follow-up
  join search already uses.
- **The carve-out count stays at two.** `jira-integration.md` bounds interactive Jira questions at
  exactly two and requires that sentence be amended before a third is added. This adds none — it
  makes fewer statuses reach the first.

The To Do name set is stated **once**, in `jira-integration.md`, and the join search cites it. That
is this repository's standing answer to a value two sites need, and `scripts/check-references.sh`
already checks a citation whose target heading moves. No sync guard is added: a script, a test
harness and a cleanup-registry entry to hold a two-name list in step is more machinery than the list
is worth, and a citation removes the divergence rather than detecting it.

## Slice D — reverted; the pipeline's existing prohibition is left standing

Three sources disagreed: `myflow-start/SKILL.md` said commit the design doc, `pipeline.md`'s Git
boundaries row said none, and `myflow-command-surface` says normatively that no command other than
`/myflow-finish` and `/myflow-do`-with-a-PR may commit. Seven archived changes committed a design
doc from `/myflow-start`, so the requirement had been violated by every change this pipeline has
planned. The first resolution tried was to make the practice legal and bounded rather than stopping
it — that attempt, and why it was abandoned, is the rest of this section.

**The bound was going to be the worktree, and that is exactly what broke.** `/myflow-do` builds its
worktree from the branch `/myflow-start` creates, not from a fresh branch off the base, so an
uncommitted plan would be absent from the worktree its implementers and reviewers read — the
worktree-visibility gap this slice originally set out to close. The fix that tried to close it was
rewritten three times across two review passes and produced every Critical finding either pass
raised. Its final form had `/myflow-start` create `openspec/<name>` and commit both planning paths
there, checked out in the main checkout. That leaves the branch checked out where `/myflow-do` needs
to check it out again to build a worktree — and git refuses to check an already-checked-out branch
out a second time, so the very next `/myflow-do` of every change would fail outright. The panel
reproduced this live: `git worktree add` exits with `fatal: … is already used by worktree at …`. The
harness's own worktree tool has no mode that checks out an existing branch either, only creates new
ones, so neither route to build `/myflow-do`'s worktree survives. The full record — what was tried,
across how many rewrites, and every finding it produced — is preserved in this change's own session
records; the operator's ruling is recorded below as a decision.

**The operator's ruling, at the panel handback: resolve the three-way contradiction the other way.**
Make `myflow-command-surface`'s existing prohibition *true* instead of carving an exception into it
— `/myflow-start` stages its planning artifacts and never commits them, exactly as `pipeline.md`'s
Git boundaries row already said before this slice was ever proposed. No delta spec is needed: the
stage-only behaviour already complies with the unmodified requirement in
`openspec/specs/myflow-command-surface/spec.md`, so restating it as a delta would be noise in the
permanent spec tree.

**The worktree-visibility problem does not disappear — it is deferred.** A plan that is only staged
is absent from the worktree `/myflow-do` builds, and nothing copies it across after the fact. That
gap is real and unresolved; it is recorded under **Open questions** below rather than solved here,
and is left for a change of its own.

## Decisions

### How brainstorming knows a round is the last one

**ID:** convergence-test-plus-confirm
**Status:** active
**Chosen:** Both — the command loops while it holds unanswered questions, then confirms with the
operator before moving on.
**Considered:** *Agent-judged only* — loops and moves on silently when it holds nothing; rejected
because "everything should be clarified" would rest on the command's own judgment, which is the
failure the issue reports. *Operator-gated only* — a prompt after every round; rejected because it
costs a prompt per round even when nothing is open, and the confirm already covers the case that
matters.

### How far back the loop reaches

**ID:** every-gate-routes-back
**Status:** active
**Chosen:** Every planning gate routes back — one convergence test after every exchange, so a design
section or a spec review that reopens a question opens a round rather than being approved over a
known gap.
**Considered:** *Questions stage only* — a later gap handled as a design revision, as today; rejected
because a design section is where an unasked question most often surfaces, and routing it to a
revision is what turns a question into an assumption.

### Whether the number of rounds is bounded

**ID:** soft-bound-from-third-round
**Status:** active
**Chosen:** A soft bound — early rounds open silently, and from the third onward the round is offered
as a named choice showing what is still open.
**Considered:** *No bound* — rounds continue until neither side has a question; rejected because a
command that keeps generating questions has no visible trajectory. *Hard cap* — move on after a fixed
number with the rest recorded; rejected because it ships a known gap on a schedule rather than on a
judgment.

### Where a question that survives the last round goes

**ID:** open-questions-section-in-design
**Status:** active
**Chosen:** An `## Open questions` section in `design.md`, beside `## Decisions`, reaching the
proposal artifact and the `STARTED` handoff count.
**Considered:** *Also blocking its task in `tasks.md`* — rejected because it couples two stages and
stalls implementation on a question the operator chose to defer. *Nowhere — stopping settles it* —
rejected because it makes the record transcript-only, which this pipeline refuses everywhere else.

### Whether a revision round re-enters the loop

**ID:** revision-round-scoped-loop
**Status:** active
**Chosen:** Yes, scoped to what the feedback reopened and whatever that opens in turn.
**Considered:** *Yes, in full* — rejected because it re-asks answered questions on a proposal re-run
precisely because most of it was right, which would make operators stop revising. *No — revisions
stay targeted* — rejected because a revision is exactly where a half-clarified plan surfaces.

### How `TO DO URGENT` is taught to the pipeline

**ID:** todo-urgent-enumerated-synonym
**Status:** active
**Chosen:** Enumerate it as a `To Do` synonym — an enumerated name at `To Do`'s position, in the
contract and the spec.
**Considered:** *A configurable synonym map in `.myflow/project.md`* — rejected as speculative
generality: it adds a second untrusted-input path into a transition decision, and nothing today needs
a second entry.

### Whether a guard holds the two status enumerations in sync

**ID:** synonym-set-cited-not-guarded
**Status:** active
**Chosen:** No guard — state the set once in `jira-integration.md` and have the join search cite it,
which `scripts/check-references.sh` already checks for a moved heading.
**Considered:** *A sync check* — rejected because a script, a test harness and a cleanup-registry
entry is more machinery than a two-name list warrants, and a citation removes the divergence rather
than detecting it.

### Where the rule is expressed across the layers

**ID:** extend-existing-capabilities
**Status:** active
**Chosen:** Extend the existing capabilities; structure in `pipeline.md`, tuned mechanics in
`myflow-start/SKILL.md`, one synonym set in `jira-integration.md`.
**Considered:** *A new capability `myflow-brainstorming-convergence`* — rejected because
`myflow-planning-gate` already owns "a question is asked, never assumed", and convergence is that
requirement read forward. *Skill-level only, no delta spec* — rejected because the repo is spec-first
by construction, and a behaviour with no requirement behind it leaves the capability permanently
silent after finish run 2 syncs.

### How wide the `/myflow-start` commit exception is

**ID:** start-commits-all-planning-artifacts
**Status:** superseded by prohibit-rather-than-legalise
**Chosen:** All planning artifacts it writes — the design document and `openspec/changes/<name>/` —
in two commits at the points the run reaches them.
**Considered:** *The design doc alone* — the narrowest carve-out, matching all seven historical
commits; rejected because it leaves the plan reaching the worktree by a route no document states.
*Any commit bounded only by "no implementation"* — rejected because a general permission gives up the
path list that makes the boundary checkable.

### Which way the commit contradiction is resolved

**ID:** legalise-rather-than-forbid
**Status:** superseded by prohibit-rather-than-legalise
**Chosen:** Make the practice legal and bounded — modify `myflow-command-surface` to name
`/myflow-start` for the planning paths.
**Considered:** *Stop committing and let finish run 1 commit the design doc* — the Git boundaries
row's literal reading; rejected because a worktree created from the base branch would then never
contain the plan its implementers read, and because it changes behaviour seven past changes relied on
in order to fix a document rather than a defect.

### Where `/myflow-start`'s commits land

**ID:** start-commits-on-its-own-branch
**Status:** superseded by prohibit-rather-than-legalise
**Chosen:** `/myflow-start` resolves the base branch — never derived from whatever happens to be
checked out — creates `openspec/<name>` from it, and makes both planning commits there, in the main
checkout; `/myflow-do` checks that branch out rather than creating one. Supersedes
`legalise-rather-than-forbid`, which bounded the exception by path alone and left it to land on
whatever branch the operator happened to have checked out.
**Considered:** *Leave the exception unbounded by branch, as `legalise-rather-than-forbid` did* —
rejected because two changes started back-to-back would then share one history: a second
`/myflow-start` run's commits land on whatever the first run left checked out, and `/myflow-do`'s
worktree — built from that branch — can silently carry a first change's entire unrelated proposal
into a second change's review surface. *Derive the branch from whatever is checked out when the run
starts* — rejected for the identical reason: resolving from the current checkout is exactly what
lets one change's branch get mistaken for another's, rather than always from the resolved base.

### Whether to keep legalising the commit exception or make the prohibition true

**ID:** prohibit-rather-than-legalise
**Status:** active
**Chosen:** Revert slice D to stage-only — `/myflow-start` stages the design document and the change
directory and never commits either, leaving `myflow-command-surface`'s existing prohibition standing
rather than amending it. Supersedes `start-commits-all-planning-artifacts`,
`legalise-rather-than-forbid` and `start-commits-on-its-own-branch`: the review panel found the
final form of the commit exception broke the ordinary run of every change — `/myflow-start` leaves
`openspec/<name>` checked out in the main checkout, and `/myflow-do`'s worktree build from that same
branch fails outright, reproduced live as `git worktree add` exiting with `fatal: … is already used
by worktree at …` — and two earlier review passes had already found every prior form of the same
exception Critical. Presented with that outcome at the panel handback, the operator chose to revert
rather than attempt a further rewrite. The full finding record is this change's own session
artifacts.
**Considered:** *A further rewrite — return the main checkout to the base branch before `/myflow-do`
checks the branch out* — rejected because nothing in this pipeline performs that step today and the
harness's native worktree tool has no mode that checks out an existing branch at all, so the fix
would need a new mechanism behind the same panel that had already found three rewrites Critical.
*Leave the worktree-visibility gap unresolved and unrecorded* — rejected because this pipeline's own
new `## Open questions` mechanism exists for exactly this kind of deferral; recording it is correct,
not merely convenient.

## Open questions

### Staging leaves the plan absent from the worktree /myflow-do builds

**ID:** start-plan-not-visible-in-worktree
**Status:** open
**Why it is open:** `/myflow-start` stages the design document and the change directory rather than
committing them (decision `prohibit-rather-than-legalise`, above). `/myflow-do` then builds its
worktree from the base branch, and `git worktree add` — along with the harness's native worktree
tool — carries no staged or untracked file into a new worktree. The plan `/myflow-start` just wrote
is therefore absent from the worktree its implementers and its review panel read, and nothing in
this pipeline copies it across. This is the same gap the reverted commit exception (slice D) existed
to close; the operator chose to defer it rather than accept the exception's own defect.
**What it affects:** every `/myflow-do` run, on every change, from the first task onward — an
implementer or reviewer reading only the worktree does not see the plan that was just approved. A
future change would need to either copy the planning paths into the worktree after it is built, or
find a different mechanism for making a staged-only plan visible where `/myflow-do` runs.

## Risks

- **The convergence test is agent behaviour and no guard can check it.** Nothing in this repository
  can mechanically prove a round was opened when one was needed; the requirement's scenarios and the
  operator at the confirm prompt are the whole enforcement. That is the same class of protection the
  planning gate's existing never-assume requirement relies on, and it is why the soft bound exists —
  it puts the trajectory in front of the operator rather than trusting it.
- **The soft bound's threshold is a judgment, not a measurement.** `3` comes from the shape of a
  converging conversation, not from data. It lives in the tuned-values file so it can move without
  touching the requirement.
- **`## Open questions` is a second immutable-ID section**, and IDs reworded rather than reassigned
  are a failure mode `## Decisions` already has. The rule is stated identically in both places for
  that reason, and the `answered by <decision-id>` link makes a stale ID visible.
- **The worktree-visibility gap slice D was going to close is now a known, deferred risk rather than
  an unrecognised one.** It is recorded in full under `## Open questions`
  (`start-plan-not-visible-in-worktree`, above) rather than restated here — the entry, not this
  bullet, is where a future change reads what it needs to close it.
