# self-review-kan-1025-flow-improvement-require-a-suite-regression-test

## Why

The self-review pass of `/flow` change `kan-1025-flow-improvement-require-a-suite-regression-test`
found three non-`big` defects:

- F-A: the mutation-testing brief (`skills/flow/review-panel.md`) demands a whole-package test run
  with no rule for a package whose run eats a large share of the slot's 15-minute ceiling. <!-- measured: kan-1025 self-review pass evidence, quoted from the plan dispatch -->
  `go test ./internal/guard` (332 tests) ran 165 s idle and about 455 s contended; the round-0 <!-- measured: kan-1025 self-review pass evidence, quoted from the plan dispatch -->
  slot's 300 s `# reproducer-bound` timed out, the slot ran 55 minutes, and the parent narrowed <!-- measured: kan-1025 self-review pass evidence, quoted from the plan dispatch -->
  the reproducer to a single-test `-run` filter, the form the brief forbids. The slot also wrote
  generic exits under `# mutation-reproducer`: the brief states the inverted convention only
  indirectly.
- F-B: **No forking, and a wall-clock ceiling on every slot** says the dispatcher tracks each
  slot's elapsed time itself but names no mechanism; a breach was caught at about three times the
  ceiling.
- F-C: `hooks/protect-main-checkout.py`'s `strip_data_heredocs` scans a heredoc body as shell
  unless the opener is `cat`/`tee` to a named file or a python/node program, so
  `flow record summary -change x -file - <<'EOF'` with `->`, `rm` or `>` in its body is denied
  from the main checkout.

## What changes

- The mutation brief permits a test-family name-prefix filter when one whole-package run takes
  more than a third of the slot's ceiling, sizes `# reproducer-bound:` from that measured run, and
  states the inverted exit convention plainly.
- The ceiling paragraph names the wake-up: the dispatch instant noted at dispatch, the bounded
  wait loop of **Turn discipline**, and an elapsed check before every re-issue that sends a slot
  past its ceiling straight to the breach path.
- The hook reads a heredoc body as data whenever its opener runs no code (no shell interpreter or
  other runner word) and the rest of the command runs none either; `$(…)`/backtick spans of an
  unquoted body stay scanned.
