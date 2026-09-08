# Design — kan-440-review-panel-roster-drift-mutation-slot-and

## Context

Two small defects, one change because both are flow-process hygiene filed together on KAN-440:

- `DefaultReviewers` (`stats/internal/store/settings.go:58`) is the value `GetSettings` reports when
  the `flow_settings` table holds no row (:154), and the list `skills/flow/SKILL.md`'s
  Model-resolution table names for the store-unreachable case. It predates the mutation slot's
  promotion to a default panel member and now disagrees with the live store row.
- `spec_root_leaf()` (`scripts/lib/spec-root.sh:73`) warns on **every** call when a project holds
  both `spectre/changes/` and `openspec/changes/`. Guards are separate processes, so the warning
  fires once per invocation; this repository is dual-tree, so every run prints the identical line
  once per guard call. The shipped copy `skills/flow/scripts/lib/spec-root.sh` is a regular file
  kept byte-identical to `scripts/lib/spec-root.sh` — both change together.

The approved design is also saved, unchanged, at
`docs/superpowers/specs/2026-09-08-kan-440-review-panel-roster-drift-mutation-slot-and-design.md`.

## Decisions

### The default roster includes the mutation slot

**ID:** default-roster-includes-mutation
**Status:** active
**Chosen:** `DefaultReviewers` becomes `primary, principles, code-review-low, mutation` — the
operator ratified the 4-slot list as the default in the brainstorming round.
**Considered:** dropping `mutation` from the store row (KAN-440's original premise) — overruled by
the operator: the store row is correct, the code fallback is the drift. A settings-schema entry
asserting a roster rule — speculative hardening for a row only `/flow-settings` writes.
Supersedes kan-404's recorded decision that `DefaultReviewers` stays
`primary, principles, code-review-low`.

### Warning deduped by a TMPDIR marker

**ID:** tmpdir-marker-dedup
**Status:** active
**Chosen:** `spec_root_leaf()` writes `${TMPDIR:-/tmp}/spec-root-dual-tree.<dir>` (path with `/`
folded to `-`) when it warns and skips the stderr line while the marker exists; marker write is
best-effort and a write failure never suppresses the warning.
**Considered:** an env var — cannot cross guard processes; a per-run marker written by the
pipeline — changes `/flow` itself and leaks a file into the project tree; a daily TTL on the
marker — date logic in the lib for a static condition. Accepted trade: silenced until the OS cleans
tmp, not strictly per run — the dual-tree condition is static, so one telling per tmp generation is
enough.

## Open questions

None — both decisions were settled explicitly by the operator during brainstorming.
