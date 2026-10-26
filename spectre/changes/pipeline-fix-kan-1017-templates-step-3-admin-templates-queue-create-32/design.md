# pipeline-fix-kan-1017-templates-step-3-admin-templates-queue-create-32

## Context

**Fewest operator actions** (`skills/flow-contracts/pipeline.md`) lists when a run may stop for the operator; it never named run size or a checkpoint as non-reasons, so kan-1017's parent read them as grounds to stop. `check-installed-citations.sh` treats every backticked path as a citation that must resolve under a known root; an illustrative placeholder path needs the guard's allow marker.

## Decisions

### Where the stop rule lives

**ID:** run-size-never-a-stop
**Status:** active
**Chosen:** one new paragraph in **Fewest operator actions** (`skills/flow-contracts/pipeline.md`) — the section every stop decision is already read against.
**Considered:** a line in `skills/flow/implement.md` — rejected: the stop happened between stages, so a per-stage file would leave verify and integrate uncovered.

### How the illustrative path passes the citations guard

**ID:** learning-path-allow-marker
**Status:** active
**Chosen:** append `<!-- citations-guard:allow -->` to the line — the guard's existing escape for an illustration.
**Considered:** rewriting the path to a resolvable one — rejected: the path is a per-project placeholder with no fixed root.

### Planning effort for the first review

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-low — two prose edits, one paragraph and one marker; session cap is opus at low effort until 2026-10-28.
**Considered:** flow-medium — rejected: nothing in the diff calls for wider reasoning.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group — Tasks 1 and 2, file-disjoint docs edits.
**Considered:** one group per task — rejected: two one-file edits do not repay a second dispatch.

## Open questions

## Recommended approach

Both edits are already committed on the branch; each task matches its commit exactly. The new sentences of Task 1 are acknowledged in `.superpowers/sdd/pipeline-fix-kan-1017-templates-step-3-admin-templates-queue-create-32/verbatim-moves.txt`.

## Critical files

- `skills/flow-contracts/pipeline.md` — **Fewest operator actions**.
- `skills/flow/cleanup.md` — the `Learning:` paragraph.

## Existing code to reuse

- The `<!-- citations-guard:allow -->` marker `scripts/check-installed-citations.sh` already honours.

## Verification

- `scripts/check-installed-citations.sh` exits 0.

## Live check

None — documentation-only change; no running service or persistent state is touched.
