## Why

`/myflow-finish` decides between its integrating run and its archiving run with
`git merge-base --is-ancestor HEAD origin/$BASE`. A branch whose work is staged but never committed
has `HEAD` still at the merge base, so it is trivially an ancestor of the base branch and the check
answers *merged*. Run 2 then archives the change, pushes the archive, and removes the worktree with
`--force`. This was reproduced on KAN-14 with 29 staged entries and 8,543 insertions inside the
worktree that `--force` deletes; only a manual check stopped it.

Two smaller gaps in the same family — a check or a step that leaves an artifact unaccounted for —
are closed alongside it, because both live in the same command. The SDD ledger recording which model
ran each dispatch is gitignored and dies with the worktree, so `myflow-model-policy` currently has
to state that its own record does not survive archival. And nothing removes the proposal artifact
source `/myflow-start` writes beside the state file, so whether a finished change leaves one behind
has depended on whether someone noticed.

## What Changes

- **The run-1/run-2 decision becomes an executable, tested script.** `scripts/check-finish-preflight.sh`
  prints one verdict — `RUN1`, `RUN2` or `REFUSE` — from three signals in a fixed order: `HEAD`
  against the merge base recorded in the state file's `worktrees` map, then the ancestor test, then
  the worktree's cleanliness. A change with no recorded merge base refuses and asks rather than
  guessing.
- **Session records are preserved into the repository** by a new
  `scripts/preserve-session-records.sh`, invoked from `/myflow-finish` run 1 before its `git add -A`
  and from `/myflow-do`'s commit path. The SDD ledger and the review panel record land under
  `docs/superpowers/`, in the same commit as the work they describe.
- **The proposal artifact source is preserved in-repo at run 1 and removed from the state directory
  at run 2**, in the same disclosed sequence as the worktree removal.
- **`myflow-model-policy` stops disclaiming its own durability.** The scope paragraph and the
  "reader asks which model implemented a given task" scenario are rewritten; the
  `unknown (agent-defined)` rule is left untouched.
- **`/myflow-start` asks how much effort to spend, once per change.** On a change's first run only,
  it asks which reasoning effort to apply to that run's brainstorming and planning, records the
  answer in a new state-file field, and never asks again — a revision round reuses the recorded
  value and says so.

Not breaking. Every change already at `STARTED` or `IN_PROGRESS` continues to work; one whose state
file carries no recorded merge base gets a refusal with an explicit operator override rather than a
silent archive, and one whose state file predates the `effort` field reads as "not recorded" rather
than becoming unparseable.

## Capabilities

### New Capabilities

- `myflow-effort`: when `/myflow-start` asks for an effort level, what the levels mean, what they
  may and may not change, and how the answer is stored and carried.

### Modified Capabilities

- `myflow-finish-cleanup`: the run-2 merge verification gains the two signals that make it correct
  and a refusal path; two requirements are added for record preservation and artifact-source removal.
- `myflow-model-policy`: the model record becomes durable, so the requirement's scope paragraph and
  one scenario no longer state that it is gone once the change is archived.
- `agents-repo-verification`: this repository's declared `## test` commands gain the two new
  harnesses, and the rule that every new guard joins `## lint` is narrowed to guards that scan the
  tree.
- `myflow-state-integrity`: `effort` joins the fields a write must carry forward and a failed
  recovery must announce.

## Impact

- **New:** `scripts/check-finish-preflight.sh`, `scripts/test-check-finish-preflight.sh`,
  `scripts/preserve-session-records.sh`, `scripts/test-preserve-session-records.sh`.
- **Modified:** `skills/myflow-contracts/pipeline.md` (Finish contract, and the `## Model policy`
  section's scope paragraph, which asserted the model record was session-scoped and told a future
  reader to make it durable — the thing this change does), `skills/myflow-finish/SKILL.md`
  (sections 1.2 and run 2), `skills/myflow-do/SKILL.md` (the commit path),
  `skills/myflow-start/SKILL.md` (the effort ask), `skills/myflow-contracts/state-file.md` and
  `skills/myflow-contracts/state-self-heal.md` (the `effort` field and its absence carve-out),
  `.myflow/project.md` (`## test`), and the specs above.
- **New directories in the repository:** `docs/superpowers/artifacts/`, joining the existing
  `docs/superpowers/ledgers/` and `docs/superpowers/reviews/`.
- **Distribution:** the two new scripts are repository-local guards, not installed by `setup.sh`
  into other harnesses; the contract text that points at them is.

Item 1 of KAN-19 — widening the provenance guard's scan scope past `tasks.md` — is deliberately
**not** in this change. It shares no file with these three, and item 3 destroys work, so it should
not wait behind a parser rewrite.
