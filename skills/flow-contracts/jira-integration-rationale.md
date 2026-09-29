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
