# A multi-glob `ui paths` matches nothing

**Jira:** [KAN-359](https://tweety53.atlassian.net/browse/KAN-359)

## Why

`check-visual-trigger.sh` mis-parses a `ui paths` value carrying more than one glob, so
`flow.visual-verify` **silently skips** for every project declaring more than one — which is every
realistic project. Gymie declares eight and would never trigger.

It shipped because every test used a single glob: all eight cases in
`scripts/test-check-visual-trigger.sh` use the same `stats/web/src/**`, and none contains a comma.
The harness pins `**` spanning directories, a leading `./`, absolute paths and globs containing
spaces — but never the one thing the contract says the value is, "comma-separated globs".

A second defect surfaced with it: `check-visual-verification.sh` reports `VISUAL-OK` on a value the
trigger guard cannot use. Two guards read the same setting and disagree silently, and the validator
is the one an author trusts.

## What changes

- `check-visual-trigger.sh` splits on commas **first**, then strips each element's own backticks.
- `check-visual-verification.sh` parses `ui paths` the same way and rejects a value yielding no
  usable glob, so the two guards cannot bless different things.
- Both harnesses gain multi-glob cases, verified RED against the current parsers.
- Both real declarations are verified to actually **trigger**, not merely validate.
