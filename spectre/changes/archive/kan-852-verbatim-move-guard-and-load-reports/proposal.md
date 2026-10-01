# Proposal — kan-852-verbatim-move-guard-and-load-reports

**Jira:** KAN-852 (epic KAN-851). **Plan:** `docs/prompt-audit-2026-09-29/README.md`.

## Why

Every trim in KAN-851 moves or de-duplicates prompt text. Today nothing in `## lint` fails a trim
that rewords or drops a rule. `check-normative-inventory.sh` sees only the `MUST`/`SHALL`
sentences, 7 of them in the whole corpus. `check-references` reads one line at a time, so a
`**Heading**` split from its path by a line break is never verified. The epic's per-session sizes
are also static estimates, and nothing yet measures what a real session reads.

## What changes

1. A new guard, `check-verbatim-moves`, in Go in `flow-guard`, with a `scripts/check-verbatim-moves.sh` shim. It
   compares the sentences of `skills/`, `commands-claude/` and `rules/` at a base ref against the
   working tree. It fails on a deleted or reworded sentence and on new run-loaded prose. It reports a
   rule that moved into a `-rationale.md` as REVIEW. It runs in `## lint`.
2. `check-references` also verifies a bold heading and a path that a line break splits.
3. `scripts/loaded-files.py` is a transcript report: what each session Read or Skill-loaded from
   the flow files, the turn it was read at, re-reads, and turn-weighted cost.
4. `scripts/load-sets.sh` is a static report of the bytes in each session's load set. It never fails.
5. The first measurement goes in `design.md`.
6. `.flow/project.md`'s note on `check-normative-inventory.sh` is corrected to state that guard's limit.

## Out of scope

Any trim itself. Those are the sibling tickets of KAN-851.
