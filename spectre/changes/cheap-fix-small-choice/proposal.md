# cheap-fix-small-choice

## Why

Operator, 2026-10-02: a port left a same-second screenshot overwrite in `classify-untracked` because fixing it "meant inventing a naming scheme" and the bash did the same. That should have been fixed.

## What changes

- `rules/agent-baseline.md`: a defect whose fix needs a small open choice (a name, a suffix, a format) counts as cheap — take the simplest choice and fix it; "it needs a scheme" and "the old code did the same" are not reasons to leave it.
- `classify-untracked`: a second same-named capture within one second gets a counter after the timestamp instead of overwriting the first.
