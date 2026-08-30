# kan-300-myflow-reviewing-a-relocation-through-a-diff-is

## Why

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
directions twice. A mechanical no-loss check cannot substitute for the ledger (rule extraction
authors new text a pure-move check would flag as loss), but a mechanical *comparison* — not a
verdict — can still turn "reconstruct 44 extraction points by reading two files" into "audit a
table."

## What changes

- A plan (`tasks.md`) declares itself a relocation via a required `**Relocation:** yes|no` header
  line, alongside the existing `**Execution:**` line.
- When a plan declares `**Relocation:** yes`, `/flow`'s review-panel stage generates
  `<worktree>/.superpowers/sdd/relocation-comparison.md` before writing `final-review.diff`: a
  before/after passage-level (paragraph/bullet/table row) comparison across the files named by the
  plan's own `**Files:**` fields, classifying each differing passage as moved, repointed, added or
  removed.
- Every panel slot's dispatch prompt names this file's absolute path, alongside the existing
  dispatch-context bundle pointer, when it exists.
- Generation never blocks the panel: a failure prints one line and the panel dispatches without
  the file.
- This does not replace the per-move ledger `myflow-contract-economy` requires, and makes no
  correctness judgment about a relocation — it is a review input, not a gate.
