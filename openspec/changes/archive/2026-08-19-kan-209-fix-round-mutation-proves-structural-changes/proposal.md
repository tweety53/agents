# A fix round mutation-proves every behaviour it changes

## Why

A review panel's fix round proves the test cases it was asked to add, and does not prove the
structural change it made alongside them. KAN-200's fix round 2 mutation-proved eight new harness
cases individually, then replaced the record protocol's delimiter — its largest structural change —
and proved nothing; reverting that protocol left all eighteen cases green. The gap was caught in
round 3, only because the Principles slot thought to ask whether the previous round had held itself
to its own standard.

The asymmetry is the natural shape of "prove the thing you were told to prove", so it will recur
every time a fix round both adds tests and changes structure. Nothing in the pipeline asks the
question, and nothing records the answer.

## What Changes

- **A fix round mutation-proves every executable behaviour it changes**, not only the test cases it
  adds. A prose-only round has nothing to mutate and records an explicit exemption.
- **Each mutation alters one mechanism.** Where a single revert would touch state that more than one
  check reads, it splits into surgical mutations. A mutation touching shared state can pass by
  cross-contamination and report coverage that does not exist — KAN-200's round 3 found exactly this
  and split its own revert in two.
- **The parent runs the mutations, after the fix subagent reports** — the same shape
  `skills/myflow-do/SKILL.md` already uses one paragraph earlier for re-running each dispatched
  finding's reproducer.
- **A surviving mutation is repaired inside the round**, by adding the test that catches it. It is
  not raised as a finding and costs no extra panel pass.
- **The parent records each mutation** in the pass log it already keeps in
  `.superpowers/sdd/final-review-panel.md`, as `fix-mutation:` lines with a `fix-mutations-total:`
  count, reusing the record's existing `none — <reason>` exemption form.
- **The fix round does not close, and so the run does not reach the handoff**, while an executable
  behaviour it changed carries neither a line nor an exemption.
- **No new guard script.** The record is read by the next round's reviewers, not by a check.

## What this deliberately does not do

**It does not require mutation tests generally.** KAN-197 examined that broader form and rejected
both mechanisms it considered — a failure fixture per harness, already universally satisfied and
compatible with the defect it was meant to catch, and a `Mutation:` field on every task in
`tasks.md`, which puts per-task friction on every future change. This is the narrower, situational
sibling: a fix round's own structural edits, nothing else.

**It does not add a guard over the new record lines.** To be more than a nag, such a guard would
have to decide from a diff whether a round changed executable behaviour or edited a comment — a
classification it cannot make, and one that fails in the direction that matters in a repository
that is mostly prose.

**It adds no slot to any preset** and changes no preset's required set. The obligation belongs to
the round, not to a slot, so it binds `light` — where no Bugbot runs, and the round's own proof is
therefore the only mutation reasoning that happens at all.

## Capabilities

### New Capabilities

<!-- none -->

### Modified Capabilities

- `myflow-review-panel-economics`: adds one requirement stating that a fix round mutation-proves
  every executable behaviour it changes, that each mutation is isolated to one mechanism, that the
  parent performs the mutations after the fix subagent reports, that a surviving mutation is
  repaired within the round, and that the pass log carries a `fix-mutation:` line or exemption per
  changed behaviour — placed outside the marker block the record's existing guard reads.

## Impact

- **Affected specs:** `myflow-review-panel-economics`
- **Affected code:** `skills/myflow-do/SKILL.md` (section 5, the fix-round text);
  `scripts/test-check-unfinished-work.sh` (a regression case for the new record lines);
  `scripts/check-contract-budget.sh` (the budget row for `skills/myflow-do/SKILL.md`, if the added
  section outgrows it)
- **No existing check changes.** `check-unfinished-work.sh` is not modified — the new lines are
  placed so its unbroken-marker-block rule and its `| F<n> |` row matching are untouched, and that
  is proved by a test rather than asserted.
