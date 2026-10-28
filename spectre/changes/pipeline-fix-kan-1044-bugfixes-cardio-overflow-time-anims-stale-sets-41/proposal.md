# pipeline-fix-kan-1044-bugfixes-cardio-overflow-time-anims-stale-sets-41

## Why

The flow pipeline prescribes load reproductions on the operator's laptop. `rules/fix-determinism-at-the-source.mdc`
("make it fail: throttle … the CPU (a parallel load …)"), `skills/flow/implement.md`'s PROOF RUNS
paragraph ("10–15 runs by default") and `skills/flow/verify-fix-loop.md` ("reproduces it under … the
same concurrent load and the full-suite ordering"; "runs the worktree's whole `## test` list twice
more") together licensed kan-1044's self-review to plan "Reproduce each under the failing run's load
and ordering … then the worktree's whole ## test list twice more". Its implementer ran concurrent
wasm-boot probes and multi-worker repeats until the laptop went out of memory (operator, 2026-10-28).

## What changes

- No flow step runs a load reproduction: no multi-worker suite repeats, no `--repeat-each` or loop of
  a suite, no concurrent browser or app-boot probes, no "reproduce under full-suite load".
- A flake seen only under parallel load is diagnosed by reading the code and the failing run's saved
  artifacts. Its fix is proven by forcing the cause deterministically in one isolated single-worker
  run (fault or delay injection); where that is impossible it lands on code-reading evidence, the
  commit body saying so.
- Repeat proof runs, where still needed, are single-worker, isolated, at most 5.
- The project's `## test` list still runs as defined, but is never repeated "twice more" for a flake.
- Never-widen-tolerance and mechanism-removed proof stay intact.
- The `proofruns` dispatch-paragraph guard pins the new wording.
