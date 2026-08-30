# Generate a sentence-set relocation comparison for the review panel

**Jira:** KAN-300
**Date:** 2026-08-30

## Problem

KAN-295 moved five sections between contract files under a byte-for-byte constraint. Reviewing
that through a unified diff cost 3 panel rounds, 3 fix rounds and ~2.4M subagent tokens, because a
diff shows lines removed in one file and added in another and proving those are the *same
sentence* — not a paraphrase, not two fragments of one sentence — requires cross-file comparison
against merge-base text, which the diff format actively hides. Every reviewer reconstructed ~44
extraction points by hand; in round 1 the Primary slot sampled 5 and found nothing, while the
Principles slot sampled 4 and found 4 defects, including one paragraph stitched from two
non-adjacent original sentences.

`myflow-contract-economy`'s per-move ledger already requires the implementer to log every removed
passage, but KAN-295's own self-review found that ledger hand-maintained and wrong in both
directions. A mechanical no-loss check cannot substitute for it (rule extraction authors new text
a pure-move check would flag as loss), but a mechanical *comparison* — not a verdict — can still
turn "reconstruct 44 extraction points by reading two files" into "audit a table."

## What changes

For a change whose plan declares itself a relocation, `/flow`'s review-panel stage generates a
before/after passage comparison as a review input, alongside `final-review.diff`, before the panel
runs.

### Declaration

`tasks.md`'s header — the block writing-plans already writes for every plan
(`skills/flow/brainstorm.md`'s **D**) — gains a second required line beside `**Execution:**`:

```markdown
> **Relocation:** yes — <one-line reason>
```

or `**Relocation:** no`. Required and explicit on every plan, never omitted — this repository's
established "missing rather than dropped" convention, so a reader can tell a plan was examined for
this property rather than the question never being asked. The scope of the comparison is the union
of every task's own `**Files:**` field across the plan; no separate per-task tag is added.

### Generation

In `skills/flow/review-panel.md`, immediately before writing `final-review.diff`: if `tasks.md`
declares `**Relocation:** yes`, run a new script,
`generate-relocation-comparison.py` (+ a `.sh` wrapper, the same pairing `check-plan-shape` and
`plan-dispatch-bundles` already use), taking the worktree, the merge-base sha, and the scoped file
list derived from `tasks.md`'s `**Files:**` fields.

The script extracts **passages** — paragraph, bullet-list item, or table row, the same unit
`myflow-contract-economy`'s per-move ledger uses — from each scoped file's merge-base blob and its
current worktree blob, then classifies every passage that differs between the two sets:

| Class | Meaning |
|-------|---------|
| moved | identical text, different file or heading |
| repointed | identical text modulo the two edits `myflow-contract-economy` already permits a moved passage — a repointed citation path/token, or a deleted stale `above`/`below` |
| added | no match in the before set |
| removed | no match in the after set |

Output is a table (passage's first eight words, source location, destination location, class) at
`<worktree>/.superpowers/sdd/relocation-comparison.md`. This is a comparison, never a verdict: it
does not decide whether a passage should have moved, only whether it can be found.

### Dispatch

Every panel slot's dispatch prompt, when this file exists, carries its absolute path alongside the
existing `dispatch-context.md` pointer, framed as a review input to audit against the diff — never
a substitute for reading `final-review.diff` itself.

### Failure handling

Generation never blocks the panel. A failure — an unreadable merge-base blob, an empty scope, a
malformed header — prints one line and the panel dispatches without the comparison file, the same
fallback shape `dispatch-context.md`'s own missing-bundle case already uses.

### Spec

A new requirement lands under `myflow-review-panel-economics` (the capability already owning
"Every panel slot is dispatched with the shared context bundle"). The `**Relocation:**` header
field itself — writing-plans writes it, review-panel reads it — is consolidated into that same
new requirement, not split into a separate addition to `myflow-contract-economy`.

## Out of scope

- Replacing or auto-populating the per-move ledger `myflow-contract-economy` requires — the ledger
  is authored by the implementer and catches rule extractions this mechanical check cannot.
- Any verdict on whether a relocation is *correct* — the comparison is an input the panel audits,
  not a gate.
