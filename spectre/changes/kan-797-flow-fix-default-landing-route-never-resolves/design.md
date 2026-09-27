# kan-797-flow-fix-default-landing-route-never-resolves — design

## Context

`.flow/project.md`'s single-line-literal keys are read by every `/flow*` phase through
`project-get.sh` (the shared `project_section` extractor) and matched against their vocabularies
per `project-configuration.md`'s match paragraph. The extractor must stay generic — `## lint`,
`## worktree setup` and friends are genuinely multi-line — so the fix lives in the **match rule**,
not the extraction. The one Go consumer in the family is `check-model-keys`
(`stats/internal/guard/modelkeys.go`), which validates `## self review model` against
`ValidModels` via its `mkSectionBody` twin. Full problem statement and the reproduced evidence:
`proposal.md` `## Why`. The approved design this file adapts:
`.superpowers/sdd/2026-09-28-kan-797-flow-fix-default-landing-route-never-resolves-design.md`
(gitignored, never committed).

## Resolution rule

For `## default landing route`, `## self review model`, `## self review` and `## handoff`:

- The value is the body's **first non-blank line**, whitespace-trimmed, surrounding backticks
  removed, matched byte-for-byte against that key's vocabulary. No case-folding, no synonym list.
- Lines below the head are documentation for the reader, never read — the `## workspace
  isolation` "prose beside it is for the reader" pattern.
- A head matching no literal stays a malformed row: reported by name (quoting what was found),
  dropped, resolving as if the key were absent.
- The four table rows drop their "holds that value and nothing else, never free-form prose"
  wording for the head shape.
- `## jira` is untouched — its body legitimately holds multiple keys and is JQL-constrained.

## Guard follow

`modelkeys.go` reads the head — first non-blank line, trim, backtick-strip — before the
`ValidModels` compare for `## self review model`, so a documented body the contract calls legal
passes. Its Go tests gain a prose-bearing-valid case and a malformed-head case. No new
family-wide validation: a malformed head is already reported at resolution time.

## Unchanged, deliberately

- `project_section` / `projectSection` / `project-get.sh` — generic extraction is load-bearing.
- This repo's own `.flow/project.md` — bare already; resolves under both rules.
- No head-extraction helper, no family-wide head-validation guard — wider than the finding.

## Decisions

### Resolve the value as the body's head, not guard authoring

**ID:** head-of-body-resolution
**Status:** active
**Chosen:** head-of-body rule — fixes the recorded failure mode (two runs tripped) while keeping
a documented default expressible.
**Considered:** authoring guard alone — keeps authoring strict but leaves the recorded failure
one prose line away from recurring; head rule + head-validating guard — declined as wider than
the finding (operator chose follow-only).

### One shared rule across the four literal keys

**ID:** four-literal-keys-family
**Status:** active
**Chosen:** all four byte-for-byte keys take the head rule, stated once in the shared match
paragraph.
**Considered:** landing route only — leaves the siblings carrying the same failure mode and
desyncs `check-model-keys` from the contract.

### The guard follows the rule without new validation

**ID:** guard-follows-head-rule
**Status:** active
**Chosen:** `modelkeys.go` reads the head for `## self review model` only.
**Considered:** validating the whole family's heads at lint time — declined: wider than the
finding; resolution-time reporting already covers a malformed head.

## Open questions
