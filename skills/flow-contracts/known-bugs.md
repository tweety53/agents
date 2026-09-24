# Known bugs — the KNOWN-BUGS.md sweep

**This file is canonical for the KNOWN-BUGS.md sweep.** Skills reference it by name; none of them
restate it. If a rule below and a skill ever disagree, this file wins.

## The sweep

When a `/flow*` run's verify stage or `flow.visual-verify` stage fails a command whose output
names a failing test, and the suite that test belongs to is one whose results pixel coordinates,
viewport size or wall-clock time can move — a visual-regression or screenshot suite above all —
the run classifies the failure **pre-existing** exactly when it can name the commit that
introduced the defect and that commit is an ancestor of the default branch:

```bash
git merge-base --is-ancestor <introducing-sha> origin/<default-branch>
```

`git log` over the failing spec and the code it exercises is usually enough to name the
introducing commit. A failure no such commit can be named for is the change's own and follows
the ordinary failing-command rules; the sweep never absorbs it.

A pre-existing failure is **recorded, never repaired**:

1. **Record it.** One entry per failure in `<project-root>/KNOWN-BUGS.md` (create the file with a
   one-line title when absent), committed on the change's own branch:

   ```markdown
   - `<spec or test name>` — <root cause, one line> — introduced by <sha> (<subject>), an
     ancestor of <default branch>.
   ```

   The entry is the durable record — what fails, why, and since when — and the next change's
   known-failure baseline. Delete an entry only in the commit that actually fixes its bug. A
   project whose lint carries `check-contract-budget.sh` adds a `budgets()` row for
   `KNOWN-BUGS.md` the moment the sweep first creates the file — the ratchet refuses an owned
   Markdown file with no declared budget.

2. **Keep the defect out of the change's diff.** The sweep documents a failure; it never fixes,
   works around, skips or tolerance-widens one (**Fix determinism at the source, never by
   widening tolerance**, `rules/fix-determinism-at-the-source.mdc`). The failing spec and the
   code it exercises are left exactly as they are.

3. **Feed the next baseline.** When the project declares `## known failures`
   (**Project configuration**, `skills/flow-contracts/project-configuration.md`), append the
   failing test's identifier to it in the same commit, spelled exactly as the key's matching rule
   requires, so the next change's verify classifies the failure as known without re-deriving it.

A failure recorded this way is reported as a **known failure** — the same report shape and the
same no-block effect a baseline match gets (**Verify**, `skills/flow/verify-and-handoff.md`) —
with the entry's root cause as its reason, instead of blocking the run on a defect this change
did not introduce. The record is what keeps that from being silent: the entry, and the baseline
line when rule 3 wrote one, are committed on the change's branch, visible in its diff, and the
verify report names them.

## Where it runs

- **Verify** — the failing-command classification consults the sweep before the inline re-run: a
  failing test the sweep proves pre-existing takes the known-failure course above.
- **Visual verification** — the stage's `verify` and the full-suite runs a baseline regeneration
  forces consult it the same way.
- **Every `/flow*` command that runs those stages**, `/flow-fast` included — the sweep is stage
  behavior, not one command's.
