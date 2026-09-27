# kan-839-flow-fix-reproducers-pass-vacuously-when-a-fix

## Why

A panel reproducer whose target file, test/class name or task id is renamed by a fix still
exits 0 on the post-fix verification re-run: its checks enumerate nothing, `run-reproducer.sh`
reads the exit as "defect not demonstrated" — the expected verdict — and the vacuous green is
counted as fix verified. Nothing audits what a reproducer's checks read once its finding leaves
`open`: the exit-contract guard skips non-open findings by design and its KAN-606 audit covers
only the defect-present side. Verified at base `4a278320`: `grep -rn premise
stats/internal/guard/` returns nothing.

Filed from the deferred self-review pass of
kan-692-flow-fix-the-wasm-dev-server-never-hot-reloads-a, where three reproducers
(0-primary-2, 0-principles-2, 0-principles-3) passed exactly this way after fix round 1 renamed
their targets, and the parent verified those fixes by hand (`DevStackFreshnessTest`, a grep)
instead of counting the green exits.

## What changes

- A `# premise: <path>:<line>:<content>` declaration (same first-10-lines window and citation
  shape as `# demonstrates:`) for every file, test/class name or `tasks.md` task id a
  reproducer's checks read — at least one mandatory for every newly authored or repaired
  runnable non-mutation reproducer.
- The script body asserts each declared premise before its real checks: a missing premise is a
  loud non-zero failure naming it, never exit 0.
- `pcAudit` resolves premise citations at dispatch time with the demonstrates machinery —
  tolerant of absence, strict about declared-but-unresolvable.
- Post-fix, a renamed premise lands in the KAN-524 ambiguity refusal (identical verdicts)
  instead of a green count, forcing the re-author path.

Observable difference: a fix that renames what a reproducer reads can no longer produce a green
verification exit — it produces a named refusal and a re-authored reproducer.
