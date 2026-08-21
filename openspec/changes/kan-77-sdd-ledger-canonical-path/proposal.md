## Why

The SDD ledger records every implementer dispatch, its model, every review finding and every
operator ruling. **Model policy** (`skills/myflow-contracts/pipeline.md`) requires each dispatch to
record its model and names the *preserved* ledger as what makes that answerable after archiving.
Nobody keeps the transcript, so a lost ledger is a lost audit trail.

Three components disagree on its path:

- `skills/myflow-do/SKILL.md` names every other artifact flat under `.superpowers/sdd/`
  (`dispatch-context.md`, `final-review-panel.md`, `final-review.diff`) and **never names the
  ledger** — so a controller writes it flat by pattern-match.
- `scripts/preserve-session-records.sh:227` reads only `.superpowers/sdd/tasks/progress.md`.
- `pipeline.md`'s Temporary artifacts registry names a *directory*, settling nothing.

On disagreement the script prints `skipped: <src> (absent)`, exit 0 — which the outcome table in
**Preserving the session records** (`pipeline.md`) defines as legitimate, since a change may
genuinely have no such record. So *never written* and *written elsewhere* are indistinguishable to
every caller, and the ledger dies with the worktree at run 2.

Evidence here: **32 preserved ledgers against 40 panel records**;
<!-- measured: ls docs/superpowers/ledgers/ | wc -l; ls docs/superpowers/reviews/ | wc -l @ branch main -->
two of the 32 carry a hand-typed `-ledger` suffix `find_dated()` can never match. KAN-77 merges nine
observations. `kan-13-myflow-planning-and-status-fixes` lost its ledger outright; KAN-73's 192 lines
were one `git worktree remove --force` from destruction.
<!-- measured: KAN-77's own description, "Occurrences" section @ the issue as fetched 2026-08-21 -->

## What Changes

- **One canonical path — `<abs-worktree>/.superpowers/sdd/tasks/progress.md` — stated in three
  places.** Not new: `sdd-workspace` derives `<root>/.superpowers/sdd/<plan-basename>/`, myflow's
  plan is always `tasks.md`, and `skills/myflow-do/SKILL.md:214` already invokes that skill.
  myflow-do names it; the registry row gains the filename; `preserve-session-records.sh` — already
  correct — cites the registry instead of agreeing by coincidence.
- **`preserve()` rescues a misfiled record.** A fifth parameter carries an ordered allowlist of
  known-wrong paths; on canonical-absent it walks the list, first hit wins, copies to the canonical
  destination, and names every path tried. Ledger list from KAN-77's occurrence log; panel list
  seeded.
- **Two outcomes, both exit 0:** `rescued: <dest> (found at <path>)` and
  `MISSING: <canonical> — tried <paths>`. Non-zero keeps its one meaning — a copy attempted and refused
  or failed. `pipeline.md`'s outcome table gains a row each.
- **Rescue and MISSING cover the ledger and panel record only.** The proposal artifact keeps the
  plain skip: `/myflow-fast` publishes none, so MISSING would fire on every fast run.
- **myflow-do § 7 asserts the ledger** with a non-gating `test -f`, reported at its call site — the
  shape § 4's dispatch-context check already uses.
- **Two misnamed preserved ledgers renamed** to `<date>-<name>.md`, each first line read first.
- **`check-task-commit-fields.py` resolves a folded red task against its partner.** It reads
  `**Squash-with:**` today only so the field terminates the previous one, so a `Build: red` task can
  never pass: its commit is folded away and its declared subject stops existing. Found by this
  change's own tasks 1+2 commit; `377b7dd` (kan-102) is a prior instance. Added scope, synced to Jira.
- **Artifact brevity, stated in `pipeline.md`.** Every artifact a `/myflow-*` run writes is written
  brief — bullets over prose, no recap, no restatement across artifacts — with a never-compress list
  for the fields guards parse, and no length guard added. It goes in `pipeline.md` because every
  command loads that file first, so one statement reaches every run. Added scope, not in KAN-77's
  description; the Jira description is synced accordingly.

## Impact

- **Affected specs:** `myflow-finish-cleanup`, `myflow-artifact-economy` (new),
  `myflow-task-commit-fields`
- **Affected code:** `scripts/preserve-session-records.sh`,
  `scripts/test-preserve-session-records.sh`, `skills/myflow-do/SKILL.md`,
  `skills/myflow-contracts/pipeline.md`, `scripts/check-task-commit-fields.py`,
  `scripts/test-check-task-commit-fields.sh`, `docs/superpowers/ledgers/` (two renames)
- **Not affected:** `scripts/gather-self-review-context.sh` — `find_dated()` already searches the
  dated shape; `docs/superpowers/ledgers/<name>.md` is only a display label. KAN-77's claim against
  it is stale.
- **Out of scope:** KAN-155 — that nothing *writes* the ledger. This change names the path and
  asserts presence; it does not specify the writing.
