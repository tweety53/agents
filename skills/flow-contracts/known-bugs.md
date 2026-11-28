# Known bugs — fixed in the run that finds them

**This file is canonical for what a `/flow*` run does with a defect it finds in the project's own
product code — its change's own, or one the change did not introduce: a pre-existing test failure,
a finding a verification change's sweep makes on a
surface another change owns.** Skills reference it as **the sweep**; none of them restate it.

## The sweep

A product defect a run finds is fixed inside that run, before integrate lands the change — never deferred to self-review.

**A defect the change did not introduce is no exception**, except a failing test proven
pre-existing and unrelated (below). It blocks exactly as the change's own
does and takes the same course — the fix procedure (`skills/flow/fix.md`) in verify and
visual verification, an appended task for a verification change's sweep (**D. Basic Workflow #3 —
Writing plans**, `skills/flow/brainstorm-planner.md`). It is never logged to a `<project>/KNOWN-BUGS.md`
file, and never skipped, worked around or
tolerance-widened (**Fix determinism at the source, never by widening tolerance**,
`rules/fix-determinism-at-the-source.mdc`).

**A failing test suspected to pre-date the change is classified by its guard**, before anything
else is done with it:

```bash
flow-guard pre-existing-failure <worktree> <base-ref> '<failing command>' <path>...
```

each `<path>` a file or directory, relative to the repository root, that the failing test covers.
The guard compares the paths with the change's diff against the merge-base, then runs the command
in a detached checkout of the merge-base:

- `CAUSED` (exit 1) — a named path is in the change's diff, or the command passes on the
  merge-base: the change's own defect. It blocks and is fixed in the change.
- `PRE-EXISTING` (exit 0) — no named path is in the change's diff and the command fails on the
  merge-base too: the one defect that does not block the landing. The run files it at once as its
  own issue, titled as **Follow-up issues** (`skills/flow-contracts/jira-followups.md`) names a
  follow-up and never joined to one, labelled and bodied per **Labels and body of issues the
  pipeline creates** (`skills/flow-contracts/jira-integration-finish.md`) — its `## Why` carrying
  the guard's `PRE-EXISTING` line and the failing output — records it with
  `flow record finding -change <name> -ref <the ref flow record next-ref prints> -slot parent -severity important -status 'withdrawn pre-existing: <issue key>' -reproducer '<failing command>' -note '<the failing test, one line>'`,
  and the change lands with that test still failing, the issue key named in its handoff. The fix
  is its own change.
- Exit 2 classifies nothing: the failure blocks.

A failing `lint` command is never classified: the landing reads that list's green.

**A stale snapshot baseline on the base is re-baselined, never asked.** A screenshot mismatch is a
failure whose run produced an `-actual.png` or `-diff.png` attachment; a "waiting for stable
screenshot" timeout is a timeout, never a mismatch. A mismatch met during a sync's targeted
recapture onto a moved base (`skills/flow/sync-onto-base.md`) is the base's baselines lagging the
base's own code only when the same tests also fail against the base's committed baselines on
captures of the base's own code — a spec's being untouched by this change is not enough, because the
change's own app code can move its pixels. Any other mismatch, a `verify` run's among them, is the
branch's own defect and takes The loop. Whoever ran it recaptures exactly the failing tests — `capture`
with `<spec>` substituted by each failing test's `<file>:<line>` — stages only the PNGs the failing
run named, restores every other file the capture rewrote, commits them on the change branch as their
own `test(fidelity): re-baseline <spec basenames>` commit, and re-runs the failing tests.
**No rewritten baseline is staged unread.** Each PNG a capture rewrote is read against its old
baseline, and against its frame where a sidecar names one, before it is staged; one showing
another screen than the test meant is restored, never committed — `--update-snapshots` writes
whatever the run captured, a missed click included. A failure
that is no screenshot mismatch — a timeout, a screen only production serves, a missing service — is
never re-baselined: it is reported, and blocks as above.

**A product defect the self-review pass meets is never dropped.** It stays out of the pass's
report and filing, and stops the landing. Met inside a run, the pass stops at the end of its
step 2 — no fix, no filing, no store row, no report: nothing is committed for the pass, so the
next run meets no committed report, runs the pass again, and meets the defect again if it was not
fixed. The `flow.self-review` mark closes `-outcome stopped`, nothing lands, the change keeps its
state, and the run continues straight into the fix run that defect names, in the same invocation —
never a handoff for the operator to start it.

**A landing HEAD whose UI no verifier saw is such a defect.** A pass finding that a UI-path commit
landed after the last HEAD a visual verifier saw — departures fixed and never re-captured, a fix
run's UI change never verified — stops the landing exactly as above. It is never a risk line, a
`noted` finding or an angle item the landing proceeds over.

- **`/flow`'s integrate run 1** stops before step 5's route and continues as
  `/flow <name> <the defect, one line>` would: the fix run writes in the change directory still on
  disk and, once it writes its state, continues into integrate as a re-run with new work, which
  runs the pass again (**Run 1 — the branch is not merged**,
  `skills/flow-contracts/finish-contract-run1.md`).

Met by the pass `/flow` runs from a saved bundle after landing, the pass's closing report names it for the
operator to start that fix run.
