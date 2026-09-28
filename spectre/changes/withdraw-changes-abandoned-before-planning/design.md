## Context

`/flow`'s creating run can end before the design gate — today only at the `flow-fix`/`flow-cost`
reachability check, whose rule is "the finding the base already delivers ends the run … the issue
is the operator's to close". What the rule leaves behind is the change itself: a `STARTED` record
that every `flow state resolve`, `state list` and `/flow-status` carries forever, plus the
worktree, local branch and pushed remote branch `flow.kickoff` created. Nothing in the pipeline
retires them, because `FINISHED` is reachable only through run 2's archive, which presumes planned,
merged work, and no command deletes a record.

The change is one change because the field, the route and the docs are one contract: the route is
the only writer of the field, and the docs are where both ends learn to offer it. Constraints that
shaped it: the record's field vocabulary is closed (`DisallowUnknownFields` on the one decode path
shared by live writes and journal replay); the monotonic rule ranks `STARTED`(0) < `IN_PROGRESS`(1)
< `FINISHED`(2); the dev daemon is protected and cannot be restarted by an agent, so a live
verification task would have nothing new to exercise against.

## Decisions

### Terminal shape — FINISHED plus a `withdrawn` boolean, no fourth state

**ID:** finished-withdrawn-boolean
**Status:** active
**Chosen:** the withdrawal route's single state write moves `STARTED` → `FINISHED` carrying
`withdrawn: true` — the operator's "move to FINISHED, mark it withdrawn somehow (the simplest
way)". Every FINISHED exclusion (candidate resolution, `/flow-status`'s open set) applies
unchanged, and the record stays as the audit trail: the change existed, when it was withdrawn, by
whom. The store refuses `withdrawn: true` with a state other than `FINISHED` (one invariant beside
`ErrInvalidState`).
**Considered:** a fourth `WITHDRAWN` state — rejected: it touches `state_rank`'s migration, the
store's valid set, and every state consumer, for a distinction only a human reader ever makes;
deleting the record — rejected: a delete verb in a write path built on monotonic upserts, the
audit trail lost, and the change's stage marks orphaned.

### Offer points — both pre-plan ends, no new command

**ID:** both-preplan-ends
**Status:** active
**Chosen:** the creating run offers withdrawal at the reachability-check end (Yes recommended —
the run just proved nothing should be planned) and at a resume of a planless `STARTED` change
(Resume default — the operator invoked it to continue). The explicit answer is the consent: it
names the worktree path, the local branch and the remote branch it deletes.
**Considered:** the reachability end only — rejected: a `STARTED` change abandoned for any other
reason keeps littering; a dedicated `/flow-withdraw` command — rejected: it grows every harness's
command surface (wrappers, README Level 1) for a route the creating run already carries.

### Route order — git first, record last

**ID:** git-first-record-last
**Status:** active
**Chosen:** the worktree and branch deletions run before the state write, so a crash mid-route
leaves a `STARTED` record and a partially cleaned tree — the route re-runs to the same end —
never a `FINISHED` record over a live worktree.
**Considered:** record first — rejected: a terminal record whose cleanup then fails strands the
remaining deletions outside any route that would run them.

### No Jira action on withdrawal

**ID:** no-jira-on-withdraw
**Status:** active
**Chosen:** a withdrawal performs no Jira transition and no description write, even when
`jiraIssue` is set — closing the issue is the operator's own act, as the KAN-828 precedent
already showed.
**Considered:** transitioning the linked issue to Done — rejected: the pipeline's Jira transitions
belong to named run phases, and a withdrawal has no issue-side truth to state that the operator's
own close would not state better.

## Open questions
