# Known bugs — fixed in the run that finds them

**This file is canonical for what a `/flow*` run does with a defect it finds in the project's own
product code — its change's own, or one the change did not introduce: a pre-existing test failure,
a visual departure present at the merge base, a finding a verification change's sweep makes on a
surface another change owns.** Skills reference it as **the sweep**; none of them restate it.

## The sweep

A product defect a run finds is fixed inside that run, before integrate lands the change — never deferred to self-review.

**A defect the change did not introduce is no exception.** It blocks exactly as the change's own
does and takes the same course — **The loop** (`skills/flow/verify-fix-loop.md`) in verify and
visual verification, an appended task for a verification change's sweep (**D. Basic Workflow #3 —
Writing plans**, `skills/flow/brainstorm-planner.md`). It is never logged to a `<project>/KNOWN-BUGS.md`
file, never filed for later, and never skipped, worked around or
tolerance-widened (**Fix determinism at the source, never by widening tolerance**,
`rules/fix-determinism-at-the-source.mdc`).

**A failure suspected to pre-date the change is isolated before it is classified.** Revert every
file the change touched in the failing area to its merge-base content, re-run the failing tests,
then restore the files: a failure that persists pre-dates the change, and one that goes away is the
change's own. The classification decides only whose defect it is and where its fix is recorded —
never whether it blocks. KAN-570 is the case that set the standard: kan-485 reverted its own
calculator-area touches, saw the same failures, and so kept a pre-existing defect out of its own
scope on evidence rather than assertion.

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
whatever the run captured, a missed click included (KAN-803). A failure
that is no screenshot mismatch — a timeout, a screen only production serves, a missing service — is
never re-baselined: it is reported, and blocks as above.

**A product defect the self-review pass meets is never dropped.** It stays out of the pass's
report and filing prompt, and stops the landing. Met inside a run, the pass stops at the end of its
step 2 — no fix, no prompt, no store row, no report: nothing is committed for the pass, so the
next run meets no committed report, runs the pass again, and meets the defect again if it was not
fixed. The `flow.self-review` mark closes `-outcome stopped`, nothing lands, the change keeps its
state, and the run ends on the block below.

**A landing HEAD whose UI no verifier saw is such a defect.** A pass finding that a UI-path commit
landed after the last HEAD a visual verifier saw — departures fixed and never re-captured, a fix
run's UI change never verified — stops the landing exactly as above. It is never a risk line, a
`noted` finding or an angle item the landing proceeds over.

- **`/flow`'s integrate run 1** stops before step 5's route. The fix run writes its plan above the
  archive commit, so the next bare `/flow <name>` is an archived re-run with new work, which runs
  the pass (**Run 1 — the branch is not merged**, `skills/flow-contracts/finish-contract-run1.md`).
- **`/flow-fast`'s verify** stops before `flow.stage-diff`; the fix re-run lands normally.

Met by a standalone `/flow-self-review` after landing, the pass's closing report names it for the
operator to start that fix run.

```
## Landing stopped — a product defect to fix first

**Change:** <name>
**Defect:** <the defect, one line>

Next:
/clear
/flow <name> <the defect, one line>
```

`/flow-fast` prints `/flow-fast <name> <the defect, one line>` as the last line instead.
