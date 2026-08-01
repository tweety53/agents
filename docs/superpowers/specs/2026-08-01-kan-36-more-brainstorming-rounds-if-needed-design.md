# KAN-36 — more brainstorming rounds if needed

**Issue:** KAN-36 "More brainstorming rounds if needed"
**Change:** `kan-36-more-brainstorming-rounds-if-needed`
**Planning effort:** default · **Date:** 2026-08-01

## Scope

KAN-36 asks for one thing — *"do more brainstorming rounds if needed, everything should be
clarified"* — and one more was added by the operator during this planning run: **always interpret
`TO DO URGENT` as `To Do`**. The two are unrelated in subject; they share a change because the second
was raised at the gate the first is about, and both are small.

| Ask | Slices | Lands in |
|-----|--------|----------|
| Brainstorming iterates until nothing is unclarified, and what survives is recorded | A, B | `myflow-planning-gate`, `myflow-handoff-output` |
| `TO DO URGENT` is read as `To Do` rather than asked about | C | `myflow-jira-projection` |
| *(raised at the spec-review gate)* `/myflow-start`'s git boundary says two different things | D | `myflow-command-surface` |

The first ask splits into two slices because the loop and the record are separable: the loop is what
produces more rounds, and the record is what makes stopping early honest. Either could ship without
the other, and the review panel reads them as different subjects.

Three capabilities, all existing, all amended. No new capability and no new guard script.

### Why the first ask is a real gap

`skills/myflow-start/SKILL.md:130` invokes superpowers:brainstorming "in full: checklist items 1–8",
and that upstream skill's flow is linear: explore → clarifying questions → approaches → design →
approval. Its only loop is *design revision* — `present design → not approved → revise` — which
routes a **correction** back to the design, never a **question** back to the questions.

`myflow-planning-gate` already requires that *"an unresolved question is put to the operator, never
resolved by assumption"*. That covers a question the command **notices**. It says nothing about a
question a previous answer **created**, because nothing takes the command back to a place where such
a question could be asked. The gap is not that questions get assumed away; it is that the stage has
no way to reach a second round.

### Why the second ask is not a reopening of kan-26

kan-26 enumerated `To Do` and `TO DO URGENT` by name as the To Do statuses the follow-up join search
accepts, and **rejected** deriving that set from Jira's `statusCategory` — which groups a custom
`TO DO URGENT` with `In Progress` under `indeterminate`. This slice does the same thing one layer
over: it enumerates `TO DO URGENT` as a **name** at `To Do`'s position in the transition order. The
forbidden inference stays forbidden. What changes is that a status the pipeline already recognised
in one place is now recognised in both.

## Slice A — brainstorming converges instead of passing once

### The convergence test

After **every** planning-stage exchange — a question round, a design-section approval, the spec
review — `/myflow-start` applies one test: *does it now hold a question it cannot answer?* While the
answer is yes it opens another round, rather than proceeding to approaches, to the design, or out of
the stage.

One test rather than three rules is the design. A design section that reopens something and an
answer that opens something are the same event — the command ends up holding an unanswered question
— so neither needs a rule of its own, and no gate can be added later that quietly bypasses the loop
by not being enumerated.

### The two prompts, and why they recommend opposite things

**The confirm** fires when the test comes back empty. The command states what it believes settled
and asks whether anything is still unclear, with **"nothing unclear — move on"** as the recommended
option. That recommendation is honest precisely because the prompt is unreachable while the command
holds anything.

**The soft-bound offer** fires from the **third** round onward, replacing the silent opening of a
round: the command shows what is still open and what it would ask, and offers the round as a named
choice, with **"another round"** recommended. This is the same shape as `/myflow-finish` run 1's
unfinished-work gate, where **Stop** is recommended because the gate only fires when something
really is unfinished — here the offer only fires while the command really does hold questions.

The two prompts therefore recommend opposite courses, and both are honest, because each is reachable
only under the fact that makes its recommendation true. Stating that here is what stops a later
editor from "harmonising" them.

### What the soft bound buys, and where its number lives

Rounds one and two open without asking: an initial round plus the round its answers open is the
ordinary shape of a converging conversation, and prompting inside it would tax the common case. A
third round means convergence is not happening on its own, which is the point at which the operator
should be able to see the trajectory and stop it — so from there the round is offered rather than
taken.

`3` is a **tuned value** and lives in `skills/myflow-start/SKILL.md`. The requirement states only
that later rounds are offered rather than taken. This is the structure-here / thresholds-cited split
`pipeline.md`'s level-2 expansions already use, and it is why the *Brainstorm* expansion gains the
shape of the loop while the number stays out of it.

### Every planning effort level runs the loop

`low`, `default` and `detailed` all converge. A level changes how many questions are grouped into a
round — one at a time at `detailed`, batched at `low` — never whether another round opens.
`myflow-planning-effort` already forbids a level switching a gate off, and a level able to end the
loop early would be exactly that: a way to skip the gate rather than a way to size the thinking
inside it.

### A revision round re-enters the loop, scoped

Re-running `/myflow-start` at `STARTED` applies the same test to **what the feedback reopened, plus
whatever that opens in turn**. Settled parts of the plan are not re-brainstormed.

Scoping it is what makes the loop usable on a revision. A revision round exists because most of the
proposal was right; re-asking answered questions would make the cheap path expensive, and the
operator would stop taking it — which would remove the loop from the one place a half-clarified plan
most reliably surfaces.

## Slice B — what survives the last round is recorded

The soft bound lets the operator stop with something still open. That is deliberate — deferring a
question is a legitimate call — but it must leave a record, because a decision held only in a
transcript is the outcome this pipeline refuses everywhere else.

`design.md` gains `## Open questions`, beside `## Decisions` and shaped like it:

```markdown
### <the question>

**ID:** <kebab-case-slug>
**Status:** open
**Why it is open:** <deferred by the operator, blocked on something external, …>
**What it affects:** <what would change depending on the answer>
```

The **ID is assigned once and is immutable**, for the same reason a decision's is: it is the match
key a later round uses. A revision round that answers the question sets `**Status:** answered by
<decision-id>` and adds the decision entry beside it. **An answered entry is never deleted or
rewritten** — the record of what was knowingly left open, and when it was closed, is the point.

A stage that left nothing open records none. The section stays empty rather than being filled with
invented entries — the same rule `## Decisions` already carries.

### It reaches the gate, not just the file

The published proposal artifact carries the section, and the `STARTED` handoff gains one line
immediately after `Decisions recorded`:

```text
**Open questions:** <N> | none
```

counted from entries still at `**Status:** open`, so an answered one does not inflate it. The count
is derived from `design.md`, which is on disk, so it is **regenerable and not run-only**:
`/myflow-status <name>` renders it from the same template. Adding it to `pipeline.md`'s `STARTED`
template and to `myflow-start/SKILL.md`'s rendering in the same change is what
**The block each state renders** requires of any new field.

`skills/myflow-status/SKILL.md` needs no edit — it renders the block by citing that template rather
than carrying a copy, so the line flows through.

## Slice C — `TO DO URGENT` is a `To Do` synonym

The ordered set becomes four **positions** with names mapped onto them, and `TO DO URGENT` maps to
`To Do`. Nothing else about transitions changes: resolved by name, case-insensitively, allowing the
usual spellings; forward-only; an issue at or past the target is left alone with no call made.

**The unrecognised-status ask survives, narrowed.** It fires for a status matching no recognised
name *including the synonyms*. The run that planned this change is exactly the case it should stop
firing for: KAN-36 sat at `TO DO URGENT`, the ask fired, and the operator's answer was that the
status should simply have been read as `To Do`.

**The synonym set is stated once**, in `skills/myflow-contracts/jira-integration.md`. The follow-up
join search's *"exactly two names — `To Do` and `TO DO URGENT`"* becomes a **citation** of it rather
than a second enumeration, which is this repository's standing answer to a value that two sites need:
one source, cited, checked by `scripts/check-references.sh`. No sync guard is added — a new script,
harness and registry entry to hold a two-name list in step is more machinery than the list is worth,
and the citation is what makes divergence impossible rather than merely detectable.

**The carve-out count stays at two.** `jira-integration.md` bounds interactive Jira questions at
exactly two — the unrecognised-status ask and the join confirmation — and says a third may not be
added without amending that sentence. This slice adds none; it narrows when the first fires. The
sentence is untouched, and `/myflow-start`'s guardrail that exactly one carve-out is reachable from
it stays true.

## Slice D — `/myflow-start` commits the planning artifacts it writes

The contract says two different things about whether `/myflow-start` may commit, and the
disagreement is three-way:

| Source | What it says |
|--------|--------------|
| `skills/myflow-start/SKILL.md:133` | Save the design doc "**and commit it** when the brainstorming skill requires it" |
| `pipeline.md` **Git boundaries** | `/myflow-start` → "None — planning artifacts only" |
| `myflow-command-surface` | "**No command other than** `/myflow-finish`, and `/myflow-do` when a PR is already open, **SHALL create a commit**" |

Seven changes have committed a design doc from `/myflow-start`, so the normative sentence has been
violated by every change this pipeline has planned. The resolution is to make the practice legal and
bounded rather than to stop it.

### The exception is the planning artifacts, and the reason is the worktree

`/myflow-do` creates its worktree from the base branch, so a file that is merely *staged* or
untracked in the main checkout is **not present in the worktree**. Yet run 1 commits the whole
`openspec/changes/<name>/` directory with `git -C <worktree> add -A`, and `/myflow-do` marks
checkboxes in `tasks.md` and appends fix notes to `proposal.md` — all inside the worktree. **No
document states how those files get there, and no mechanism does either.** All three candidate sites
were checked during planning: `skills/myflow-do/SKILL.md` has no copy step and reads the plan through
`openspec instructions` at stage 1, *before* the worktree exists at stage 2; superpowers:using-git-worktrees
runs a plain `git worktree add`, which carries nothing untracked; and the harness's native worktree
tool branches from a ref, likewise carrying nothing uncommitted. The only reading that explains a
past run committing the plan from the worktree is the documented sandbox fallback, where worktree
creation fails and the "worktree" is the main checkout.

Committing them at `/myflow-start` makes the mechanism explicit and removes the gap: the worktree
inherits the plan from the base branch, the way it inherits everything else. That is a portability
fix as much as a tidiness one — on any harness where the worktree really is a separate checkout, the
plan is currently absent from it.

**`skills/myflow-do/SKILL.md` is therefore not in this change's surface.** There is no dead copy step
to remove, because there was never one to begin with.

So the boundary becomes: `/myflow-start` **commits the planning artifacts it writes** — the design
document under `docs/superpowers/specs/` and the change directory under `openspec/changes/<name>/` —
and performs **no other git action**: no branch, no worktree, no push, no merge, no pull request, and
nothing outside those two paths. `myflow-command-surface`'s "no command other than" sentence names
`/myflow-start` for exactly those paths.

### Two commits, at the two points the run reaches them

The design doc is committed mid-run, when the brainstorming skill requires it and before the change
directory exists; the change directory is committed at the end of the run. They cannot be one commit
without holding the design doc uncommitted through the rest of the run, which is the thing the
brainstorming skill asks for. The existing message convention is kept for the first
(`docs(<key>): design for <topic>`) and matched for the second (`chore(<name>): plan and delta
specs`).

A revision round rewrites the artifacts and commits again, which makes the revision history visible
in git rather than only in a republished artifact.

### What this leaves for finish run 1

Run 1's planning commit carries what remains: `docs/manual-test/<name>.md`, the preserved session
records, and any artifact edits made after `/myflow-start`. Its guarded skip-if-empty rule already
covers a run where nothing is left, so no chain changes. The two-commit split — implementation
first, planning second — is untouched, and `/myflow-do`'s staging exclusions are unaffected: a plan
that is already committed cannot appear in the review diff at all, which strengthens the property
those exclusions exist to protect rather than weakening it.

## Surface

| File | Change |
|------|--------|
| `specs/myflow-planning-gate/spec.md` | **ADDED** — brainstorming iterates to convergence, with the two prompts and the record |
| `specs/myflow-command-surface/spec.md` | **MODIFIED** — `/myflow-start` commits the planning artifacts it writes, and nothing else |
| `specs/myflow-jira-projection/spec.md` | **MODIFIED** — the four ordered names become positions carrying names, `TO DO URGENT` at `To Do` |
| `specs/myflow-handoff-output/spec.md` | **MODIFIED** — the `STARTED` field row gains the open-questions count |
| `skills/myflow-contracts/pipeline.md` | *Brainstorm* level-2 expansion gains the loop's structure; `STARTED` template gains the line |
| `skills/myflow-start/SKILL.md` | Section B round mechanics and the threshold; `## Open questions` in C; artifact content in E; handoff block in F; guardrails |
| `skills/myflow-contracts/jira-integration.md` | The To Do synonym set, stated once; the join search cites it |
| `skills/myflow-contracts/pipeline.md` *(slice D)* | The **Git boundaries** row for `/myflow-start` states the commit and its bounds |
| `skills/myflow-start/SKILL.md` *(slice D)* | Section F commits rather than stages; the message convention; guardrails name the bound |

**Out of scope.** `/myflow-do`, `/myflow-finish` and `/opsx:explore` are untouched. No new guard
script. Open questions do **not** block tasks in `tasks.md` — coupling the planning record to
implementation would stall a task on a question the operator deliberately deferred, which is the
opposite of what deferring one means. `skills/myflow-status/SKILL.md` needs no edit, per slice B.

## Decisions

### How brainstorming knows a round is the last one

**ID:** convergence-test-plus-confirm
**Status:** active
**Chosen:** Both — the command loops while it holds unanswered questions, then confirms with the
operator before moving on.
**Considered:** *Agent-judged only* — the command loops and moves on silently when it holds nothing;
rejected because "everything should be clarified" would then rest on the command's own judgment,
which is the failure the issue reports. *Operator-gated only* — a prompt after every round; rejected
because it costs a prompt per round even when nothing is open, and the confirm already covers the
case that matters.

### How far back the loop reaches

**ID:** every-gate-routes-back
**Status:** active
**Chosen:** Every planning gate routes back — one convergence test applied after every exchange, so
a design section or a spec review that reopens a question opens a round rather than being approved
over a known gap.
**Considered:** *Questions stage only* — the clarifying-questions stage iterates and a later gap is
handled as a design revision, as today; rejected because a design section is where an unasked
question most often surfaces, and routing it to a revision is what turns a question into an
assumption.

### Whether the number of rounds is bounded

**ID:** soft-bound-from-third-round
**Status:** active
**Chosen:** A soft bound — early rounds open silently, and from the third onward the round is
offered as a named choice showing what is still open.
**Considered:** *No bound* — rounds continue until neither side has a question; rejected because a
command that keeps generating questions has no visible trajectory and nothing but the operator's
patience stops it. *Hard cap* — move on after a fixed number with the rest recorded; rejected
because it ships a known gap on a schedule rather than on a judgment.

### Where a question that survives the last round goes

**ID:** open-questions-section-in-design
**Status:** active
**Chosen:** An `## Open questions` section in `design.md`, beside `## Decisions`, reaching the
proposal artifact and the `STARTED` handoff count.
**Considered:** *Also blocking its task in `tasks.md`* — writing-plans links each open question to
the tasks it affects and `/myflow-do` pauses there; rejected because it couples two stages and
stalls implementation on a question the operator chose to defer. *Nowhere — stopping settles it* —
rejected because it makes the record transcript-only, which this pipeline refuses everywhere else.

### Whether a revision round re-enters the loop

**ID:** revision-round-scoped-loop
**Status:** active
**Chosen:** Yes, scoped to what the feedback reopened and whatever that opens in turn.
**Considered:** *Yes, in full* — re-run brainstorming from the top; rejected because it re-asks
answered questions on a proposal re-run precisely because most of it was right, which would make
operators stop revising. *No — revisions stay targeted* — rejected because a revision is exactly
where a half-clarified plan surfaces, so excluding it excludes the case the issue is about.

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
**Considered:** *A sync check* — a new guard asserting both enumerations list the same names;
rejected because a script, a test harness and a cleanup-registry entry is more machinery than a
two-name list warrants, and a citation removes the divergence rather than detecting it.

### Where the rule is expressed across the layers

**ID:** extend-existing-capabilities
**Status:** active
**Chosen:** Extend the two existing capabilities (plus the handoff one for the count); structure in
`pipeline.md`, tuned mechanics in `myflow-start/SKILL.md`, one synonym set in `jira-integration.md`.
**Considered:** *A new capability `myflow-brainstorming-convergence`* — rejected because
`myflow-planning-gate` already owns "a question is asked, never assumed", and convergence is that
requirement read forward; two capabilities on one subject is the drift this repo keeps collapsing
into citations. *Skill-level only, no delta spec* — rejected because the repo is spec-first by
construction, and a behaviour with no requirement behind it leaves the capability permanently silent
after finish run 2 syncs.

### How wide the `/myflow-start` commit exception is

**ID:** start-commits-all-planning-artifacts
**Status:** active
**Chosen:** All planning artifacts it writes — the design document and the `openspec/changes/<name>/`
directory — committed in two commits at the points the run reaches them.
**Considered:** *The design doc alone* — the narrowest carve-out, matching all seven historical
commits; rejected because it leaves the plan reaching the worktree by an undocumented route, which is
the gap this slice found. *Any commit bounded only by "no implementation"* — rejected because a
general permission gives up the path list that makes the boundary checkable, leaving the guardrails
as the only limit.

### Which way the commit contradiction is resolved

**ID:** legalise-rather-than-forbid
**Status:** active
**Chosen:** Make the practice legal and bounded — modify `myflow-command-surface` to name
`/myflow-start` for the planning paths.
**Considered:** *Stop committing and let finish run 1 commit the design doc* — the table's literal
reading; rejected because a worktree created from the base branch would then never contain the plan
its implementers read, and because it changes behaviour seven past changes relied on to fix a
document rather than a defect.

## Risks

- **The convergence test is agent behaviour and no guard can check it.** Nothing in this repository
  can mechanically prove a round was opened when one was needed; the requirement's scenarios and the
  operator at the confirm prompt are the whole enforcement. That is the same class of protection the
  planning gate's existing "never resolved by assumption" requirement already relies on, and it is
  why the soft bound exists — it puts the trajectory in front of the operator rather than trusting it.
- **The soft bound's threshold is a judgment, not a measurement.** `3` is chosen from the shape of a
  converging conversation, not from data. It lives in the tuned-values file precisely so it can move
  without touching the requirement.
- **`## Open questions` is a second immutable-ID section**, and IDs that are reworded rather than
  reassigned are a failure mode `## Decisions` already has. The rule is stated identically in both
  places for that reason, and the `answered by <decision-id>` link is what makes a stale ID visible.
- **The `STARTED` block gains a field, and two renderers must agree.** The handoff contract already
  requires a field added to a command to be added to the template in the same change; this design
  names both edits, and `myflow-status` is unaffected because it cites rather than copies.
- **Slice D legalises a commit on the base branch.** `/myflow-start` creates no branch, so its two
  commits land wherever the operator is checked out — `main`, in every run this repository has seen.
  That is the existing behaviour and the slice does not widen it, but it is now written down, which
  makes it reviewable rather than incidental. The bound that keeps it safe is the path list: two
  paths, both planning-only, with no push, so nothing leaves the machine until `/myflow-finish`.

## Open Questions

None. Every question raised during brainstorming was put to the operator and answered; each is
recorded above as a decision.
