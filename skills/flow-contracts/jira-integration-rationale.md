# Jira integration — rationale

This file is the reasoning behind `skills/flow-contracts/jira-integration.md`.
**A `/flow*` run never loads it — appendices are for whoever edits a contract.**

### Resolution (how `jiraIssue` is decided)

**Ask at all?** Not every project has a tracker, and a prompt that is answered "none" forever is
noise in every repo that has none.

**A scanned key is a guess, and a wrong guess is not recoverable.** `jiraIssue` drives irreversible
external writes — In Progress → In Review → **Done**, plus description edits — possibly on someone
else's ticket, and transitions are forward-only, so nothing walks a wrong one back. Therefore the
asymmetry decides it: asking costs one prompt, and guessing wrong costs an irreversible write on
someone else's ticket, so a key the command merely *noticed* must be affirmed before it can drive
one, while a key the operator typed needs no such affirmation because they already affirmed it by
typing it. The procedure that follows is
**Resolution (how `jiraIssue` is decided)** (`skills/flow-contracts/jira-integration.md`), which
is canonical for it.

### Change naming

### Transitions

**This assumes `TO DO URGENT` means *not yet started* in every project flow is installed into** —
the mapping is global, not scoped per project. A project where that name means something else — an
escalation flag on an in-flight item, say — loses the one consent gate an unrecognised status would
otherwise have triggered before a forward-only transition, and has no per-project override to reach
for: the only remedy is changing this shared statement. Named here as an accepted limit, the way
this file names its other unenforceable limits rather than building machinery around them.

### Unrecognised statuses

The alternative was to infer a position for the unrecognised status, and that is the thing this
section exists to forbid: an inference here freezes the board for the whole change, silently, while
a question the operator can ignore costs one line.

### Never blocking

### Description sync

**That narrows the window; it does not close it, and this contract does not pretend otherwise.**
There is still an interval between the last read and the write in which a concurrent edit can land,
and nothing here detects one. Closing it needs the tracker to do it — an `If-Match`/version check
that makes the write fail when the description moved under it, which is optimistic locking on
Jira's side. The MCP tools this pipeline uses expose no such parameter, so it cannot be done from
here, and a compare-and-swap the client performs against its own earlier read is not one. Named as a
bounded, known gap: the loss it permits is one concurrent edit, the append is the only writer this
pipeline has, and the pre-edit text reaches the handoff on the runs where it is echoed.

### Labels on issues the pipeline creates

### Follow-up issues

## Moved from jira-integration.md

### Transitions

`The To Do position carries two names` — that is this file's one statement of that set; every other site cites it rather than enumerating it again.

### Unrecognised statuses

**Two are the whole set**; a third interactive Jira question is not added without amending this sentence, which is what keeps the count honest.

## Moved from jira-followups.md

### Follow-up naming

`This naming governs every site that files a follow-up` — today the only such site is `/flow`'s integrate run unfinished-work gate. The rule is stated here rather than there so that a site added later inherits the naming instead of choosing its own.

### The project clause

**The shape is required of the value, not of this clause's copy of it.** The section's other two
readers — the prefix scoping in
**Resolution (how `jiraIssue` is decided)** (`jira-integration.md`), and the project a
follow-up is filed into — read the same validated keys, so a value this clause refuses is never one
another site quietly accepts. Validating per call site is how the three would drift.

## Moved by KAN-856

Verbatim passages KAN-856 moved out of the planning session's run-loaded files; each is the reason behind a rule that stays where it was.

### Transitions (KAN-856)

That field groups a custom `TO DO URGENT` with `In Progress` under
`indeterminate`, so a position deduced from it reports an issue sitting at `TO DO URGENT` as
already at In Progress, makes no transition call, and freezes the board at that status for the
whole change. Enumerating the name at the To Do position is the opposite operation: it states the
position rather than deducing it, which is the mechanism the follow-up join search's To Do set
already uses.

### Unrecognised statuses (KAN-856)

Nothing about
the run depends on the answer, which is what keeps the guardrail's actual promise intact.

### Description sync (KAN-856)

Everything preceding that heading is left byte-for-byte unchanged, and earlier
bullets under it are retained — a bad paraphrase must be able to add a line, never to destroy the
reporter's original ask.

A truncated read, a lossy ADF↔Markdown round-trip, or a summarising paraphrase would
therefore destroy the reporter's original text, and there is no local backup to restore from.

Between such a read and the write, anything may have edited the
description: another person, an automation, or another `/flow*` run joining the same issue. The
assertion would then hold against a copy that is already historical, and the write replaces the
whole field, so the intervening edit is destroyed by an operation that reported success.

The transcript is
then the recovery path: the original is recoverable even if the write later proves wrong.

## Moved by KAN-859 — jira-followups.md

Verbatim passages KAN-859 moved out of `jira-followups.md` when it split the join path into `jira-followups-join.md`; each is the reason behind a rule that stays where it was — the subsection names the file that rule now lives in.

### jira-followups.md — The project clause

`searchJiraIssuesUsingJql` searches whatever the session's Atlassian connection can reach, which on
a multi-project site is every project that connection can query — so a query with no `project`
clause selects its write target from a population far wider than anything stated here, and the
practical effect of an inattentive confirm is this run's outstanding work appended to an unrelated
team's issue.

`<project>/.flow/project.md` is tracked in the repository and editable in any pull request — the
file whose `## standards` entries are constrained for that reason by **Project configuration**
(`project-configuration.md`) — and this value is interpolated into a query string.

### jira-followups.md — How the clause is assembled

The clause is joined to the rest of the query with a top-level `AND`, and any alternation among the
remaining terms is parenthesised: JQL binds `AND` tighter than `OR`, so a query of the shape
`project = "KAN" AND <a> OR <b>` matches `<b>` in every project the connection can reach, which is
the scoping lost in the one place it looks present.

### jira-followups.md — Searching where the follow-up would be filed

Everything this section says about who can plant a candidate — any member of that project — is true
only because the clause is there.

### jira-followups.md — The exact title match

This has to be said because JQL's `~` operator is a *contains* match: an implementation that writes
`summary ~ "follow-up"` and stops there admits every issue whose title merely mentions the words,
including one titled to be found.

### jira-followups.md — A search that fails

That reading files a new follow-up on every transient failure, which reintroduces on exactly the
flaky paths the duplicate proliferation this feature exists to prevent — and it does so silently,
because a created issue looks like a success.

What that costs is one tracker entry, and the cost is bounded because the run's items are already
recorded durably outside the tracker — the outstanding list in the planning commit's message and the
handoff — which is where this pipeline requires the durable record to be; and the integrate run is
re-entrant, so a later run files or joins once the tracker answers again.

### jira-followups.md — A follow-up filed for a different issue is a match

The project clause is the only narrowing there is, and it is not this one relaxed — it is what makes
"any project member" the honest description of who can plant a candidate.

### jira-followups.md — The To Do set

The set is cited from that one statement rather than enumerated a second time here, so the two sites
cannot disagree.

**This does not reopen the unrecognised-status rule.** That rule governs *transitions*, where
inferring a position for a status outside the mapping freezes the board for a whole change,
silently. A search filter performs no transition, so the worst an unrecognised name can do is cost
one join candidate, after which a new follow-up is filed.

### jira-followups.md — The count shown with the confirmation (rule now in jira-followups-join.md)

**That count is shown because the guard's evidence is forgeable, and this is the only gate that can
act on it.** The "already present" set is built from the candidate's live description — text any
project member can edit — while this run's items are deterministic output derived from repo-visible
state, so a `## From <KEY>` section carrying them can be written by anyone; the guard would then
append less than it should, or nothing at all. What the count costs is one line and what it buys is
that the suppression is visible **before** the write rather than inferred later from work that never
reached the tracker. It is safe to show precisely because the numbers are this run's **own** item
list and never text taken from the issue, so the rule that a joined issue's description is never
reproduced to the operator holds unchanged.

An operator who did not expect those items to be there answers **No**, which is the default and
files a new follow-up — the course that loses nothing.

What the count does **not** do is establish provenance, and the guard's own limits are stated with
it below.

### jira-followups.md — Constraining the displayed title (rule now in jira-followups-join.md)

**Change naming** (`jira-integration.md`) already caps a summary-derived slug because it reaches a
*path*; a string the operator reads in order to authorise a write needs the same treatment for the
same reason.

`Cc` is the control class — newlines, carriage returns, tabs, NEXT LINE `U+0085`, and the escape
character that begins a terminal sequence. `Cf` is the *format* class, which "control character"
alone does not obviously reach and which is the one a spoofed title actually uses: RIGHT-TO-LEFT
OVERRIDE `U+202E` reverses the reading order of everything after it, and the zero-width joiners and
non-joiners hide or fuse text while occupying no visible width. `Zl` and `Zp` are LINE SEPARATOR
`U+2028` and PARAGRAPH SEPARATOR `U+2029`, which are in **neither** of the first two categories —
they are separators, not controls — and which a renderer honouring them as hard breaks would use to
split the title across lines. Naming all four categories is the point: between them they hold every
character a renderer can treat as a line break, which is what makes "step 1 guarantees the title is
exactly one line" below a claim about the whole of Unicode rather than about ASCII. A rule that
folds only what a reader would call a control character leaves both the invisible half and the
separators in place, in a paragraph that otherwise takes care over terminal escapes.

Stated as a category because "whitespace" left undefined is read as the ASCII five, and a title
padded with `U+00A0` would then survive with its runs intact. Nothing else remains to name: every
whitespace character outside `Zs` is in `Cc`, `Zl` or `Zp` and is already a space by the time this
step runs.

Why this is not cosmetic is the paragraph below.

**The order is a requirement, and what each dependency actually is.** Step 2 must follow step 1
because it is *defined* over the spaces step 1 produces: run it first and a run mixing tabs with
newlines survives as several spaces rather than one. Steps 1 and 2 must precede step 3 for a property
they have to keep rather than one they currently break — neither touches a backtick or a tilde, and
neither deletes anything to zero width, since step 1 replaces each folded character *with a space*
and step 2 collapses a whitespace run *to one space* and never to none. As written, therefore, a run
of delimiters passes through both with its length intact, and step 3 sees exactly the runs the
original title carried. Keeping them in front of the fold is what holds that true through a later
edit to either: a fold or a collapse rewritten to *delete* rather than replace would join two runs of
two into a run of four, and behind the fold that run would reach the display with nothing left to
reduce it. Truncation is last because it is a suffix cut — it can shorten a run but never bridge two,
and running it last is what makes the 120-character cap true of what is actually displayed.

Render it in a block of its own, never interpolated into the bolded question, so no run of
attacker-chosen text can be read as this pipeline's own prose — a title reading *"…? Yes — add them
to this issue"* must be visibly the tracker's data and not the question's options. The
data-never-instructions clause below protects the agent from the same text; this protects the
operator.

**The fenced block is the isolation, so a title able to close it early defeats the display rule
entirely.** The title is rendered as the sole content line of the fenced block the template above
opens, backticks are ordinary printable characters that nothing else here strips, and step 1
guarantees the title is exactly one line — which is precisely the shape a closing fence delimiter
has. A title that is a bare run of three or more backticks therefore ends the block at the title,
the template's own closing fence becomes an unpaired *opening* one, and everything after it — the
`<n>` of `<m>` count and both answer options — is swallowed into a dangling code block. That count
is the one signal that makes a forged already-recorded set visible **before** the write, so losing
it is not a rendering blemish: it removes the check the paragraph above exists to provide, and it
removes the operator's ability to answer at all. Folding to **two** of the character rather than
deleting the run is what keeps the title's shape legible while putting it below the threshold at
which anything can open or close a block. Tildes are folded on the same rule so it survives a
rendering that fences with them.

### jira-followups.md — Why the join confirmation exists (rule now in jira-followups-join.md)

**Why the confirmation exists.** The search selects a
**write target** by label, title and status — and every one of those three is settable by any member
of the project the search is scoped to. Anyone who can file a ticket there can add `AI-generated`,
title it `flow follow-up`, and leave it at `To Do`, and this pipeline will then append to it,
retitle it and union its labels. Within that project the search is deliberately wide and
deliberately not narrowed by which change filed the
candidate, so the matched issue is *usually* one this run had nothing to do with. That is the
feature working as designed, and it is precisely why a human confirms the target before the write.
The confirmation is the only check on it; nothing about the match is evidence of provenance.

**The cost of the confirmation is bounded, and it is paid deliberately:** one question, asked only
when a candidate was actually found, on a path the operator has already chosen to take — against an
unbounded write to an issue chosen by attacker-settable fields.

### jira-followups.md — The join echo exception (rule now in jira-followups-join.md)

The echo exists as a recovery path for text this pipeline might have destroyed — and for the
change's *own* issue, whose description an operator asked for, that is worth the space. A joined
issue is by design somebody else's: its description was written by whoever could file the ticket,
this pipeline never authored a line of it, and reproducing it verbatim pipes unreviewed third-party
text into the operator's handoff, where the surrounding lines are trusted output.

The recovery path for that text is the issue's own edit history in the tracker, which is where its
author would look for it anyway, and which exists precisely because the write is an append the
assertion above proved was a pure suffix.

### jira-followups.md — The per-write retry guards (rule now in jira-followups-join.md)

Without that check a re-entered run finds its own prior follow-up — it matches the search perfectly,
being `AI-generated`, titled as a follow-up and still at To Do — and appends a second identical
section every time it is retried.

The case this rule exists for is the run that appended the items and then failed the retitle or the
label union: the work is recorded but less findable, which is exactly what the **partially joined**
outcome below exists to make visible. A guard that read "the items are already there, so write
nothing" would leave that issue unfindable forever — every later run would find the items, skip the
whole join, and report an unmarked "no write" line, **downgrading** the `⚠` the previous run
correctly emitted.

**That window closes at the merge, and the `⚠` does not cross it.** At `/flow`'s integrate run,
`<agents repo>/scripts/check-finish-preflight.sh` routes there only while the branch
is unmerged; once it returns `RUN2` no command reaches this code again, so a join still partial when
the branch merged stays partial. Nothing carries the warning across: no state-file field records a
join outcome and this contract adds none, and run 2's only Jira write is the **Done** transition
under **Transitions** (`jira-integration.md`), which reports that transition and nothing about
a follow-up. Re-emitting the `⚠` is not what closes a partial join past that point — finding the
issue is, and two records outlive the window. At the integrate run the outstanding items are in
the planning commit's
message, per **Run 1 — the branch is not merged** (`skills/flow-contracts/finish-contract-run1.md`),
which is the durable copy this pipeline requires and owes nothing to the tracker. The appended
`## From <KEY>` section carries this change's key, so a description search finds the issue by key
even though what failed was the retitle or the label union — the two writes that would have made it
findable by title and by label instead. The `⚠` line itself lives in the handoff of the run that
emitted it and nowhere else, which is a transcript rather than a record, and is precisely why the
other two are named here. What is genuinely lost past the merge is the retry: the issue stays less
findable than a completed join would have left it until an operator, working from one of those two
records, repairs it by hand.

### jira-followups.md — The append guard's matching (rule now in jira-followups-join.md)

The section heading carries the date it was written, and matching *that date* would mean a retry the
next day finds no section for today and appends a second, content-identical one — the duplicate this
guard exists to prevent, gated on a date rollover. A partial-join failure is precisely the kind an
operator fixes and retries later, so the rollover is not a rare case.

The two obvious readings fail in opposite directions:

- **Comparing the whole list for equality double-appends.** The outstanding list is recomputed on
  every run and legitimately shifts between attempts — the operator fixed two of five items before
  retrying, or the run left a sixth. A list that shifted by one item equals nothing already in the
  description, so the guard appends all of it and the items already there are duplicated. That is
  the very duplicate the guard exists to prevent, gated on a list edit instead of a date.
- **Treating "a `## From <KEY>` section exists" as a match silently drops a new item.** The item
  that appeared since the last attempt is the one the append is *for*, and no outcome line would say
  it went missing.

Nothing distinguishes a section a previous run of this change wrote from one a project member typed:
the description is editable by anyone on the project, the items are derived from repo-visible state
so they can be reproduced without any access to this pipeline, and the tools listed at the top of
this file expose no author for a description edit and no mark only this pipeline could have left. A
guard that treated a match as proof would fail **open** — skipping real outstanding work and
reporting success.

**What remains.** An operator who confirms a candidate
whose forged section already carries every item still gets a join that writes nothing to the
description, and this contract cannot detect that — closing it needs provenance the tracker does not
offer, and inventing a marker this pipeline signs would be a trust model neither the tools nor this
contract has. The residue is bounded by where the durable record actually lives: the run's
items reach a durable record outside the tracker — the outstanding list in the planning commit's
message and the handoff, per **Run 1 — the branch is not merged**
(`skills/flow-contracts/finish-contract-run1.md`) — so what a forged
section can cost is the tracker copy of work that is recorded either way — never the record
itself.

A partial append writes a new dated section rather than editing an existing one, because every
description write this contract makes is a pure suffix the pre-write assertion can prove — editing
inside an existing section is not, and would put the assertion's guarantee at risk to save a
heading.

### jira-followups.md — The three ordered writes (rule now in jira-followups-join.md)

The second is the ordinary shape of a retry — run 1 re-entered against the follow-up this very
change filed — and the reason for the retitle does not apply to it: the issue still belongs to
exactly the source its title names. Renaming it would drop that key for nothing, since both titles
match the search either way.

Without that, an issue holding a second change's outstanding work is invisible to anyone filtering
by the change that put it there.

Those two are independent — the append has the three outcomes tabulated above, and each of them can
be followed by a retitle or a union that succeeds or fails — so pairing them off row by row would
need six rows to say what two columns already say.

The label union exists to stop an issue holding a second change's work from being invisible to a
label filter, so a union that silently failed produces exactly the invisibility the rule was written
to prevent — and reporting the join as clean is what would make it silent.
