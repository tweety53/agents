# kan-512-flow-defer-self-review-to-a-stronger-model-via — design

Seeded from `docs/superpowers/research/kan-512.md`, sections 2–6. That note is canonical for the
reasoning behind each decision below until this file absorbs it; it is deleted once adopted here.

## Context

KAN-512: run 2's step 9 self-review reasoning pass always runs synchronously, inside the same
archive dispatch, with no way to save the gathered context and run the reasoning pass later on a
different model. See `proposal.md`'s `## Why` for the full problem statement.

## How

### The bundle: one committed file beside the report it becomes

`<project>/docs/self-review/<name>-context.md`, written by step 9 physically under
`<landing-worktree>`, committed on `chore/archive-<name>` with subject `docs(self-review): <name>
self-review context bundle`. "Self-review pending" is derived from that file existing — no new
state field, no new stage outcome, no store write. The deferred pass deletes it in the same commit
that adds the report.

Contents, in order: `gather-self-review-context.sh`'s output verbatim (its header line and four
sections), then `## design.md` holding the archived `design.md` verbatim, then `## Session
narrative` holding the archived `narrative.md` (or an `absent` marker for a change that predates
it) followed by one paragraph the step-9 session appends for run 2 itself.

`check-self-review-report.sh`'s find gains `-not -name '*-context.md'`.

### Choosing to defer

`## self review` accepts a third literal, `defer`; the absent-key prompt at step 9 gains a third
option, **Defer — save the bundle for `/flow-self-review`**, between Yes and No. On `defer`, step 9
gathers, writes and commits the bundle, runs no reasoning pass, and ends `flow.self-review` with
`-outcome completed` exactly as `skip` does. The handoff's `Self-review` line gains `deferred —
docs/self-review/<name>-context.md`.

### `run` also drops its subagent dispatch — every self-review pass is inline now

Operator instruction, mid-run: self-review never dispatches a subagent, `run` included. Step 9's
`run` path — like `defer`'s `/flow-self-review` — runs the five-angle reasoning pass directly in
the session already executing step 9, on whatever model that session is on. This removes the
`Model:` handshake and the `opus` re-dispatch entirely: there is no second model to hand off to and
nothing to compare against, so nothing to reconcile. `SELF_REVIEW_MODEL` (**Model resolution**,
`skills/flow/SKILL.md`) stops governing a dispatch — no self-review pass is ever dispatched — and
that section is corrected to say so; the settings-store field itself is untouched, out of scope for
this change.

`/flow-fast` is untouched — it does not run step 9.

### The narrative: every invocation appends, the archive carries it

`<changeRoot>/narrative.md`, a dated `## <YYYY-MM-DD> — <run kind>` section per invocation, appended
at `flow.write-in-progress` (every creating and fix run) and `flow.preserve-sessions` (run 1). Run
2 has no commit of its own that reaches the change directory, so step 9 writes its own paragraph
straight into the bundle's `## Session narrative` section instead. It rides the existing
`spectre/changes/` commit pathspec (`commit-split.sh`), the existing archive move, and the
registry's **Change directory** row — no new mechanism.

### `/flow-self-review <name>` — standalone, inline, on the session's own model

A new skill, `skills/flow-self-review/SKILL.md`, with stubs `commands/flow-self-review.md` and
`commands-claude/flow-self-review.md`, shaped like `/flow-settings`: standalone, not a pipeline
stage — no state file, no `flow stage` mark. The operator picks the model with `/model` before
invoking it. The reasoning pass runs inline in the session — no subagent dispatch, no `Model:`
handshake, no fallback.

Procedure, run from the project's main checkout on `<default-branch>`: resolve the bundle (absent →
list every pending bundle or `none pending`, stop; wrong branch → print it, stop); read it and run
the five angles over it; explain every finding, then one filing-and-rating `AskUserQuestion`; file
chosen findings per **Labels on issues the pipeline creates**
(`skills/flow-contracts/jira-integration.md`); write the report with a `**Deferred:**` line naming
the model and the bundle path; delete the bundle; one commit, `pull --rebase`, push (a rejected push
left local and named); end naming the report path, rating and filed keys.

### Project configuration

`## self review` becomes three literals: `run`, `skip`, `defer`. This change sets both `agents` and
`gymie` (reached through the peer worktree `spectre/peers` already declares) to `defer`.
`gymie-frontend` and `gymie-admin-frontend` are not peers of `agents` and are an operator step after
this lands.

## Decisions

### Committed bundle over a gitignored `.superpowers/sdd/` home

**ID:** context-bundle-location
**Status:** active
**Chosen:** commit `docs/self-review/<name>-context.md` on the archive branch — visible in `git
status`, survives a clone, needs no separate staleness handling.
**Considered:** `.superpowers/sdd/` (gitignored, so invisible and lost on another clone, and would
need its own staleness handling that a committed file gets from `git status` for free);
`docs/self-review/pending/` (one directory hop away from the report it becomes, and the registry
row would then have to name two directories instead of one).

### Deferral records nothing beyond the bundle file

**ID:** defer-outcome-parity
**Status:** active
**Chosen:** `defer` ends `flow.self-review -outcome completed`, identically to `skip` — the bundle
file is the whole "pending" record.
**Considered:** a distinct stage outcome or a state-file field for "pending self-review" — rejected
as a second copy of the same fact the bundle's existence already carries.

### The narrative is written by every invocation, not only by step 9

**ID:** narrative-authorship
**Status:** active
**Chosen:** append a dated section at `flow.write-in-progress` and `flow.preserve-sessions`, so the
session that actually lived a verify saga or a merge conflict records it.
**Considered:** a narrative written by the step-9 session alone — rejected because that session is
the integrate/archive invocation and usually never ran implement or verify, reproducing the
parity gap the ticket's own caveat describes.

### A deferred pass is a documented subset, not a claimed parity with a same-run pass

**ID:** deferred-pass-scope
**Status:** active
**Chosen:** the report's `**Deferred:**` line states plainly that the reasoning ran from the saved
bundle, on the named model — no claim that it reproduces everything a same-run pass would have
found.
**Considered:** silently presenting a deferred report as equivalent to a same-run one — rejected per
the ticket's own caveat: several of KAN-459's highest-value findings existed only in the parent
session's own observations, not in the four static sources the bundle gathers.

### Every self-review pass runs inline — no dispatch left, `run` included

**ID:** self-review-always-inline
**Status:** active
**Chosen:** step 9's `run` path drops its subagent dispatch and runs the five-angle pass directly in
the archive session, exactly as `/flow-self-review` already runs it inline. No `Model:` handshake,
no `opus` re-dispatch.
**Considered:** keeping `run` as a dispatch to `SELF_REVIEW_MODEL` while only `defer` goes inline —
rejected per an explicit operator instruction mid-run: self-review dispatches no subagent, in
either mode.

## Open questions

None — this design was fully seeded from `docs/superpowers/research/kan-512.md`, itself produced by
an investigate-then-ask `/flow-plan` session with no question left open, plus one operator
instruction taken mid-run (self-review-always-inline, above).
