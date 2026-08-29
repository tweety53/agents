# kan-362-myflow-guard-test-suite-takes-119s-sequentially

**Jira:** [KAN-362](https://tweety53.atlassian.net/browse/KAN-362)

## Why

`.flow/project.md`'s `## test` list names 40 Bash/Python guard harnesses and runs them
sequentially. Its own runtime note records the consequence — "roughly 144s total … against this
harness's 120000ms default tool timeout" — and works around it in prose, telling the caller to
split the run across invocations or raise the tool timeout. A verify stage that must be split
across invocations to fit is a verify stage that gets partially run.

The harnesses are already parallel-safe: each builds its own `mktemp -d` trees and sandboxes, so
nothing about running them together needs a harness change.

```text verified:xargs -P 8 over the same 40 harnesses, 10-core macOS (Darwin 25.5.0), in an apply worktree — the run KAN-362 reports
Sequential (documented)   ~118.6s
xargs -P 8                  49.6s, 40/40 passing
```

Two harnesses dominate the tail:

| Harness | Time | Cause |
|---|---|---|
| `test-run-reproducer.sh` | 17.7s | 21 `sleep` sites, several `sleep 30`/`sleep 31` — it exercises `run-reproducer.sh`'s real timeout and kill paths, so it genuinely waits on the wall clock |
| `test-check-installed-citations.sh` | 17.6s | 14 guard invocations, each of which makes the guard build a sandbox and run `setup.sh` twice |

```text verified:time scripts/test-check-installed-citations.sh, re-run for this change on 2026-08-29
17.649s wall, 3.62s user + 3.28s system = 6.90s CPU -> 39% utilisation
```

**39% utilisation is the load-bearing number.** That harness spends roughly 60% of its wall clock
waiting on process spawn and filesystem rather than computing, which is what decides how it is
fixed — see `citations-parallel-not-injected` in `design.md`.

### Two figures in KAN-362 are wrong, and nothing here plans against them

- **"invokes sandboxed `setup.sh` 21 times (`grep -c 'setup.sh'` = 21)."** That counts textual
  occurrences, not invocations. The harness invokes the guard **14** times (`grep -c 'run_guard'`
  = 14, `grep -c 'new_fixture_repo'` = 14) and the guard runs `setup.sh` **twice** per invocation —
  `global`, then `all` — inside a sandbox it creates itself. The real figure is 28.
- **"the parallel critical path falls to about 8s (`test-check-panel-reproducers.sh`, the next
  longest)."** That omits `test-run-reproducer.sh`, which the same ticket lists at 17.7s as the
  single longest harness and explicitly scopes out. With it untouched, it sets the floor.

**The target this change is verified against is ~20s**, from ~118.6s — about 6×.

## What changes

- A new `scripts/run-guard-tests.sh` discovers `scripts/test-*.sh` and runs them concurrently,
  quiet on pass and replaying a failing harness's own captured output in full.
- A new `scripts/lib/parallel.sh` owns the spawn/capture/replay primitive for both its callers.
- `scripts/test-check-installed-citations.sh` runs its 14 independent cases on that same
  primitive; `check-installed-citations.py` is not touched.
- `.flow/project.md`'s `## test` drops from 40 lines to one, and its runtime note is rewritten
  <!-- measured: awk '/^## test$/{f=1} f&&/^```/{c++; if(c==2) exit} f&&c==1&&/^scripts\/test-/' .flow/project.md | wc -l @ c38f24e -->
  against the new measurement instead of prescribing a split-the-run workaround.
- Two new harnesses, `scripts/test-lib-parallel.sh` and `scripts/test-run-guard-tests.sh`, are
  discovered by the runner's own glob — so the suite covers its own runner.
