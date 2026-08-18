# A corpus-scanning guard reports what it checked

## Why

A guard exits 0 when it finds no violation. That result is produced identically by a guard that
examined the whole corpus and found it clean, and by one that examined **nothing at all**.

KAN-73 shipped exactly that. `scripts/check-guard-symlinks.sh` rule 2 required every guard a skill
invokes to be symlinked into that skill. For `skills/myflow-fast/` — the skill carrying the **most**
symlinks, 17 of them — the required set computed to empty, because that skill delegates by citation
and names no guard the classifier could see. Deleting one of its symlinks still produced
`GUARD-SYMLINKS-OK`, exit 0.

Three reviewers read that guard. Its own harness passed. It was caught only because someone
deliberately broke the tree and noticed the guard did not care.

## What Changes

- **A corpus-scanning guard reports per-member coverage** in its verdict, naming every member whose
  count is zero — on a **passing** run, so a rule that covers nothing announces itself without anyone
  having to imagine the missing case.
- **Each such guard declares which members legitimately check nothing.** An undeclared zero is a
  named violation with a non-zero exit. `myflow-start` is a legitimate zero; `myflow-fast` was not.
- **Scope is the four guards that discover their corpus from the tree** —
  `check-guard-symlinks.sh`, `check-references.sh`, `check-vocabulary.sh`,
  `check-stage-mark-calls.sh`.

## What this deliberately does not do

**It does not require a failure fixture per harness** — KAN-197's own cheaper proposal, checked and
rejected. All 16 `scripts/test-check-*.sh` harnesses already assert a failure, and
`test-check-guard-symlinks.sh` carried two rule-2 violating fixtures while missing `myflow-fast`
entirely. The rule is already universally satisfied and is compatible with the defect it was meant to
catch.

**It does not add a `Mutation:` field to `tasks.md`** — KAN-197's other proposal. That puts per-task
friction on every future guard change to catch a defect living in the guard rather than the task, and
would be authored by the same blind spot that produced it.

Both are recorded under `## Decisions` in `design.md`, and KAN-197 has been corrected so the ticket
and this change do not contradict each other.

## Impact

- **Affected specs:** `agents-repo-verification`
- **Affected code:** `scripts/check-guard-symlinks.sh`, `scripts/check-references.sh`,
  `scripts/check-vocabulary.sh`, `scripts/check-stage-mark-calls.sh`, and each of their
  `scripts/test-check-*.sh` harnesses
- **No existing check changes.** This adds reporting and one new violation class; no rule any guard
  already enforces is altered.
