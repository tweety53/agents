# kan-440-review-panel-roster-drift-mutation-slot-and

## Why

KAN-440 (self-review finding from KAN-423; the issue's "3-reviewer convention" premise was
corrected by the operator during brainstorming):

- The reviewer-list fallback `DefaultReviewers` (`stats/internal/store/settings.go`) carries the
  old 3-slot roster (`primary, principles, code-review-low`) while the ratified default roster —
  and the live store row — is 4 slots (`primary, principles, code-review-low, mutation`). A run
  resolving reviewers with no store row (fresh install, wiped table, store unreachable) silently
  reviews with a different panel than the intended default.
- `spec_root_leaf()` (`scripts/lib/spec-root.sh`) prints its dual-tree warning on every guard call
  in a project holding both `spectre/changes/` and `openspec/changes/` — true for this repository,
  so one run prints the identical stderr line 3+ times (once per guard invocation).

## What changes

- `DefaultReviewers` becomes the 4-slot list; doc comment updated; `skills/flow/SKILL.md`'s
  Model-resolution fallback row, which enumerates the trio, gains `mutation`. A pinning test in
  `stats/internal/store` makes the list an intentional, guarded choice.
- `spec_root_leaf()` writes `${TMPDIR:-/tmp}/spec-root-dual-tree.<dir>` when it warns and stays
  silent while that marker exists; the spectre/openspec selection is unchanged on every call. Both
  repo copies (`scripts/lib/`, `skills/flow/scripts/lib/`) change together. New
  `scripts/test-lib-spec-root.sh` covers selection and dedup (auto-discovered by
  `scripts/run-guard-tests.sh`).
