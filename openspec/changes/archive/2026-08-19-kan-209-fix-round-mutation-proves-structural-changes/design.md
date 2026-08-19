# Design — a fix round mutation-proves every behaviour it changes

Motivation is in `proposal.md`; the normative statement is in
`specs/myflow-review-panel-economics/spec.md`. The approved design this is adapted from is
`docs/superpowers/specs/2026-08-19-kan-209-fix-round-mutation-proves-structural-changes-design.md`.

## Context

`skills/myflow-do/SKILL.md` section 5 already describes what a fix round does: the parent unions the
open findings, verifies each finding's reproducer, dispatches one fix subagent on
`models.panelFix`, re-runs every dispatched finding's reproducer once that subagent reports, and
records the pass in `.superpowers/sdd/final-review-panel.md`.

Two of those steps are the anchors this change builds on.

The first is the re-run: *"Once the fix subagent reports, re-run every dispatched finding's
reproducer under the same constraints and require it now to exit 0."* The parent, not the subagent,
performs the verification. This change adds a second thing the parent does at that same point, for
the same reason.

The second is the pass record: *"Record every pass in `.superpowers/sdd/final-review-panel.md`:
mode, which agents ran, why, the diff path they read, and — when this pass bounced any finding —
each bounced finding's defect identity together with the reproducer output it carried back."* The
new lines join that list.

Bugbot already carries a reasoned mutation-testing brief, but it reads the diff as a **reviewer**,
and only where a preset dispatches it. Neither covers the case here: a fix round's own structural
change, in a run whose preset may dispatch no Bugbot at all.

## Goals

- A fix round's structural changes are proved to the same standard as the test cases it adds.
- The proof is isolated, so it cannot pass by cross-contamination.
- The proof is recorded, so the next round's reviewers can read it instead of having to think to
  ask for it.
- The obligation holds the round open until it is met.

## Non-Goals

- **Not a general mutation-testing requirement.** KAN-197 examined that and rejected it.
- **Not a new guard script.** No check parses the new lines.
- **Not a new finding class.** A surviving mutation is repaired in the round, not raised.
- **Not a roster change.** No slot added, no preset's required set altered.
- **Not a change to task implementation or the red-task-partner's fixups.**

## Where the text lands

One new subsection in `skills/myflow-do/SKILL.md` section 5, immediately after the paragraph
beginning *"Once the fix subagent reports, re-run every dispatched finding's reproducer"* and before
the materiality condition that follows it. That position is load-bearing: the rule reads as an
extension of a mechanism the reader has just met, rather than as an unrelated obligation bolted on
at the end of the section.

The `### Bugbot's mutation-testing brief` subsection is left untouched. The two are deliberately
separate: one briefs a reviewer about the diff, the other binds the round about its own edit, and
merging them would make the second inherit the first's dependence on a preset that dispatches
Bugbot.

## The record's placement, and the constraint it must respect

The new lines go in the pass log entry:

```text verified:authored in-tree for this change
fix-mutation: <path> — <what was mutated> — <the test that failed>
fix-mutation: <path> — none — <reason>
fix-mutations-total: <n>
```

`check-unfinished-work.sh` reads the same file for a different purpose, and two of its rules
constrain what may be added to it:

- Every `finding-status:` marker must occupy **one unbroken run of consecutive lines**. That rule
  exists so a marker written elsewhere in the record cannot stand in for a missing one. A
  `fix-mutation:` line placed between two markers would split the block and make the guard report a
  spread it would then refuse.
- The findings table's identifiers are derived from lines matching
  `^\|?[[:space:]]*F[0-9]+[[:space:]]*\|`. A new line shaped like a table row would enter that set
  and unbalance the row-versus-marker comparison.

Neither is a hazard the new lines run into as designed — they carry no `F<n>` in a pipe-delimited
first cell and sit in the pass log, not the marker block. But "as designed" is exactly the kind of
claim that is true when written and false after the next edit, so this change proves it with a case
in `scripts/test-check-unfinished-work.sh` rather than asserting it in a comment.

## Why the parent runs the mutations

The fix subagent has the worktree and knows what it changed, so having it mutation-prove its own
work is one dispatch cheaper. It is also the party whose completion is being judged.

The failure KAN-209 exists to prevent is precisely a round assessing its own work and concluding it
was thorough. Fix round 2 was not careless — it applied the standard rigorously where it had been
asked to, eight cases for eight, reported individually — and still missed the change it made
alongside them. A rule that asks the same actor to widen its own scope is asking for the judgment
that already failed.

`/myflow-do` has settled this question once already, in the paragraph directly above where this text
lands: the parent re-runs each reproducer rather than accepting the subagent's report of success.
Following that shape costs the parent a revert, a test run and a restore per named behaviour — the
subagent has already done the identification work — and keeps one story about who verifies what.

## Why a surviving mutation is repaired in the round

Bugbot's surviving mutant is an ordinary finding because Bugbot is a reviewer: it reads a diff it
did not write, and the fix belongs to whoever wrote it, so a finding is the only way to route it.

A mutation the parent runs over a fix round's own change is in a different position. The round has
the behaviour in hand, the diff is minutes old, and the missing test is a few lines. Raising a
finding would open a fresh dispatch, a fresh reproducer verification, and a fresh panel pass to
recover context the round has not yet lost. The bar is unchanged either way — nothing hands off
until the coverage exists — so the finding route buys no safety and spends a full round.

## Decisions

### Record the mutations in the panel record rather than stating the rule alone

**ID:** record-in-panel-log
**Status:** active
**Chosen:** The parent writes `fix-mutation:` lines and a `fix-mutations-total:` count into the pass
log entry — the rule is stated in `skills/myflow-do/SKILL.md`, and its evidence exists in the file
the next round's reviewers already read.
**Considered:**
- *Prose-only, exactly as KAN-209 proposed* — cheapest, and no risk to the record's existing
  readers. Ruled out because the ticket's own lesson is that the omission was caught only because
  one reviewer happened to ask an unprompted question; a rule whose satisfaction leaves no trace
  cannot be checked by the next round any more than it could by the last.
- *Record plus a guard script over the lines* — strongest on paper. Ruled out under **no-guard**
  below.

### No guard reads the new lines

**ID:** no-guard
**Status:** active
**Chosen:** No guard script parses `fix-mutation:` or `fix-mutations-total:`, and no existing guard
is extended to.
**Considered:**
- *A guard that fails when a fix round changed a script and declared no mutation* — to be more than
  a nag it must decide, from a diff, whether the round changed executable behaviour or edited a
  comment. That classification is not one a script can make, and in a repository that is mostly
  prose it fails in the direction that matters: it either fires on every prose fix round until
  someone narrows it into vacuity, or is written permissively from the start and checks nothing.
  KAN-197 rejected two mechanisms with exactly this shape — a per-harness failure fixture that every
  harness already satisfied while remaining compatible with the defect, and a `Mutation:` field on
  every task in `tasks.md`.

### The parent performs the mutations, after the fix subagent reports

**ID:** parent-verifies
**Status:** active
**Chosen:** The subagent names the executable behaviours it changed; the parent mutates each and
confirms a test fails.
**Considered:**
- *The fix subagent proves its own work* — one dispatch cheaper and it already holds the context.
  Ruled out because self-certification is the failure mode this change exists to prevent, and
  because `/myflow-do` already re-runs reproducers itself for the same reason, one paragraph above
  where this text lands.
- *Subagent runs them, parent spot-checks the ones touching shared state* — splits the cost, but
  adds a judgment call ("which ones touch shared state?") to the actor whose judgment is being
  checked, which is the original problem in miniature.

### A surviving mutation is repaired inside the round, not raised as a finding

**ID:** repair-in-round
**Status:** active
**Chosen:** The round adds the test that catches the surviving mutation before it closes.
**Considered:**
- *Raise it as an ordinary `F<n>` finding, as Bugbot's surviving mutant is* — consistent with the
  existing rule and needs no new disposition. Ruled out on cost: it spends a full extra fix round
  re-establishing context the round has not lost, for a bar that is identical either way.
- *Report it and put it to the operator with named options* — most flexible. Ruled out because it
  adds a prompt to every such round to decide something with one defensible answer.

### The obligation holds the fix round open

**ID:** blocks-the-round
**Status:** active
**Chosen:** The parent does not close the fix round, and so never reaches the handoff, while an
executable behaviour it changed carries neither a line nor an exemption.
**Considered:**
- *Record the gap and proceed* — cheaper and never stalls a run. Ruled out because it leaves exactly
  the hole KAN-209 describes, with a paper trail added; the ticket's complaint is not that the gap
  went unrecorded but that it went unproved.

### Scope is executable behaviour, with an explicit exemption for prose

**ID:** scope-executable
**Status:** active
**Chosen:** Anything a test could fail on. A prose-only round records `none — <reason>`.
**Considered:**
- *Only "structural" changes* — narrower, but "structural" is a judgment the round makes about its
  own work, which is the asymmetry KAN-209 says will recur. Ruled out for that reason.
- *Every change including prose, with no notion of executable* — uniform and needs no judgment, but
  is the same obligation with the word "executable" deleted; the exemption line already handles the
  prose case at one line's cost, without asking anyone to mutate a paragraph.

### The new lines sit in the pass log, never in the marker block

**ID:** lines-outside-marker-block
**Status:** active
**Chosen:** The pass log entry, alongside the mode, the slots, the diff path and the bounced
findings.
**Considered:**
- *Beside the marker lines, where a reader of finding state would see them* — better for a reader,
  fatal for `check-unfinished-work.sh`: its unbroken-block rule would report the markers as spread
  over more lines than there are markers, and refuse a record that is in fact correct. The reader
  benefit is recovered by keeping both in one file, which they already are.

## Risks / Trade-offs

- **The parent's per-round cost rises** → bounded to a revert, a test run and a restore per named
  behaviour. The identification work stays with the subagent, which has already done it as part of
  writing the fix.
- **A round could under-report which behaviours it changed, satisfying the letter of the rule** →
  not fully closed, and honestly so. The parent reads the fix diff it already has, so an omission is
  visible; but no mechanism forces the comparison. This is the residue of **no-guard**, accepted
  deliberately rather than papered over.
- **The new record lines could break `check-unfinished-work.sh` after a future edit** → a case in
  `scripts/test-check-unfinished-work.sh` asserting the guard's verdict is identical with and
  without them.
- **`skills/myflow-do/SKILL.md` may outgrow its `check-contract-budget.sh` row** → the budget is a
  ratchet, and raising a row for a genuine addition is the guard's documented correct response. The
  addition is kept to the shortest text that states the rule.

## Migration Plan

None. Nothing reads a stored artifact whose shape changes, and a panel record written before this
change carries no `fix-mutation:` lines and is read exactly as it was.

## Open Questions

None.
