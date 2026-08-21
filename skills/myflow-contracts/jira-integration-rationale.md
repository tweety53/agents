# Jira integration — rationale

This file is the reasoning behind `skills/myflow-contracts/jira-integration.md`.
**A `/myflow-*` run never loads it — appendices are for whoever edits a contract.**

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
**Resolution (how `jiraIssue` is decided)** (`skills/myflow-contracts/jira-integration.md`), which
is canonical for it.

### Change naming

### Transitions

**This assumes `TO DO URGENT` means *not yet started* in every project myflow is installed into** —
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
