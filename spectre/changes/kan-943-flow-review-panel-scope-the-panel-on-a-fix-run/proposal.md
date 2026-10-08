# kan-943-flow-review-panel-scope-the-panel-on-a-fix-run

## Why

KAN-758's fix runs (rounds 5, 7, 8, 9) re-ran the full compact panel over the ~16k-line
whole-branch `final-review.diff` for small appended tasks. The base still does this:
`skills/flow/document-fix.md` states "the append never narrows the panel", and
`check-late-fix-trigger.sh` condition 4 (scope growth) sends every append to the full path, whose
pass 1 reads the whole branch although everything before the last clean close was already
reviewed clean.

## What changes

- `check-late-fix-trigger.sh` gains a third verdict, exit 3 `append scope: <n> changed lines since
  <sha>`: the change is clean, the base has not moved and the panel machinery is untouched, and
  only the size (condition 3) or scope-growth (condition 4) condition failed.
- On that verdict, a fix run's pass 1 runs the decided roster unchanged, every dispatch reading
  the since-close delta (`late-fix.diff`) instead of the whole branch.
- The late-fix reduction (exit 0) and the full path (exit 1/2) are unchanged.
