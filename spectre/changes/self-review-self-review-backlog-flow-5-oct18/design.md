## Context

Found by the self-review pass of `/flow` change `self-review-backlog-flow-5-oct18` (angle
`flow-fix`). Its archived `tasks.md` line 16 reads "Self-review fixes: one commit per task with its
`**Commit:**` subject and no `Task-Id:` trailer"; `scripts/close-task.sh` step 2 refuses a `/flow`
task commit whose `Task-Id:` trailer does not name the task. Reproduced at `045471f7`: that
`tasks.md` copied into a live `spectre/changes/self-review-backlog-flow-5-oct18/` passes
`scripts/check-plan-shape.sh` with exit 0.
<!-- measured: mkdir -p spectre/changes/self-review-backlog-flow-5-oct18 && cp spectre/changes/archive/self-review-backlog-flow-5-oct18/tasks.md spectre/changes/self-review-backlog-flow-5-oct18/ && scripts/check-plan-shape.sh spectre/changes/self-review-backlog-flow-5-oct18/tasks.md @ 045471f7 → exit 0 -->

Loop plans carry the same sentence legitimately: their fixers commit with no trailer
(`skills/flow-self-review/SKILL.md` step 3 item 1; the pipeline-fix fixer likewise). Loop plans and
`/flow` plans share the same `> **Execution:**` header, and a `/flow` change may itself be named
`self-review-…` (this one is), so neither the file's text nor the change's name tells them apart.
The branch does: `/flow` gives a change `spectre/<name>`; a standalone loop runs on
`self-review-<name>`, a pipeline fix on `pipeline-fix-<name>-<k>`, and an own-branch loop on
`spectre/<reviewed change>` — never `spectre/<its own name>`. `hooks/enforce-loop-dispatch.py`
already discriminates this way (eebc4f8c).

Measured with the planned guard applied to this worktree at `045471f7`: the live copy above
exits 1 naming line 16; `self-review-kan-urgent-5-oct18`'s loop plan copied the same way exits 0;
the no-argument `scripts/check-plan-shape.sh` scan and `scripts/test-check-plan-shape.sh` exit 0;
`CHECK_PLAN_SHAPE_ROOT=/Users/tweety53/Projects/agents scripts/check-plan-shape.sh` prints no
`Task-Id` line; of every archived plan in the main checkout stating "no `Task-Id:`", none has a
`spectre/<name>` branch locally or on origin.
<!-- measured: the commands above, run with the Task 1 diff applied @ 045471f7 -->

Blast radius: 3 files — the precedent `plan-script-paths` (973d3451) took the same three: one Go
guard, its Go table test, and one `check_one` line in `scripts/check-plan-shape.sh`.
<!-- measured: git show --stat 973d3451 → 3 files changed; Task 1's **Files:** names 3 @ 045471f7 -->

## Decisions

### Guard over prose

**ID:** guard-over-prose
**Status:** active
**Chosen:** a `check-plan-shape.sh` refusal, so a planner who copies the rule into a `/flow` plan again is stopped by a guard that fails without the fix.
**Considered:** confining the rule in `skills/flow-self-review/SKILL.md` / `skills/flow/brainstorm-planner.md` text — a prose change fails nothing when a planner copies the sentence anyway, which is how this one arrived.

### New Go subcommand

**ID:** go-subcommand
**Status:** active
**Chosen:** `flow-guard plan-task-id-rule <tasks.md>` in `stats/internal/guard/plantaskidrule.go`, run by `check_one` beside `plan-script-paths` — new guard logic lives in Go, the shim only calls it.
**Considered:** a new F14 in `scripts/check-plan-shape.py` — grows a Python guard where new logic goes to Go; folding into `plan-script-paths` — a different contract in one subcommand.

### Discriminator

**ID:** spectre-branch-discriminator
**Status:** active
**Chosen:** a plan is a `/flow` plan when its change `<name>` has `refs/heads/spectre/<name>` or `refs/remotes/origin/spectre/<name>` in its repository — the test `hooks/enforce-loop-dispatch.py` applies.
**Considered:** the change name's prefix (`self-review-`/`pipeline-fix-` exempt) — exempts `self-review-backlog-flow-5-oct18`, the very offender; the current branch equal to `spectre/<name>` — silent for every live plan scanned from another branch, where a ref test answers the same anywhere; the `> **Execution:**` header — loop plans carry the same one.

### Unreadable repository is clean

**ID:** no-repo-clean
**Status:** active
**Chosen:** a git call that fails (a directory outside any repository) reads as no `spectre/<name>` branch, so exit 0.
**Considered:** exit 2 — `scripts/test-check-plan-shape.sh`'s fixture trees are not repositories and would all turn "cannot answer"; a tree git cannot read holds no branch, so clean is the true answer.

### What is matched

**ID:** match-no-task-id
**Status:** active
**Chosen:** any unfenced line of the whole file matching `(?i)\bno\s+` followed by `` `Task-Id:` `` — Global constraints sit above the first task, so task bodies alone would miss it; archived plans and files outside `changes/<name>/` are skipped, `plan-script-paths`' convention.
**Considered:** task bodies only — misses the offending line; a broader "Task-Id" match — flags plans that require the trailer.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — three files on the `plan-script-paths` precedent's shape, a pure-addition guard whose seven table cases pin both directions, the loop-plan pass included.
**Considered:** `flow-medium` — the guard gates every plan, but a false refusal shows at once in `check-plan-shape.sh` and the measured history carries none.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-self-review-backlog-flow-5-oct18-fix`, Task 1.
**Considered:** none — a single task has no split.

## Open questions

## Live check

No runtime to verify against: the change is a plan-time guard run from the checkout, no service,
store or daemon. The guard's own measurement is the `## Context` runs above.
