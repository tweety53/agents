# kan-1019 — live verification

Written by the self-review pass's fix plan (`self-review-kan-1019-flow-improvement-port-integrate-change-sh-to-go`,
decision `repair-kan-1019-live-check`): the creating run deferred this check to integrate, and
integrate ran only its shim half.

**Stand-in.** `prepare` writes this change's own integrate stage marks and state, so the two runs the
design names were made on a scratch two-repository fixture of the integrate-change harness's own
shape (bare origins, a `spectre/demo` worktree per repository, `link.md` merge order, a stub `flow`
on `PATH` answering the state record and logging stage marks), built afresh at one fixed path for
each run, every commit date pinned so both runs see the same shas. Each run was made from inside
the fixture's canonical worktree, as a parent session makes it.

Before figures: `prepare` through the bash at `a03f3c4b` (`git show a03f3c4b:scripts/integrate-change.sh`,
saved beside symlinks to this tree's other scripts). After figures: `prepare` through the shim at
HEAD `1420a779`. stdout, stderr, exit code and the stage-mark log compared with `cmp`.

| Fixture | Before exit | After exit | stdout / stderr / stage marks | Last line |
|---|--:|--:|---|---|
| clean | 0 | 0 | identical | `NEXT: append this run's narrative to …` |
| staged foreign work in `b` | 1 | 1 | identical | `STOP: foreign-staged — ask the operator …` |

Failure look (a shim build error, an exit 2 naming a missing sibling, any differing line or exit
code, a stop the bash would not make): not seen — both pairs matched byte for byte.

The design's second half, this change's own integrate through the shim: `prepare` and `commit` ran
through it with no stop (the self-review pass's session narrative); `land` follows this pass.
