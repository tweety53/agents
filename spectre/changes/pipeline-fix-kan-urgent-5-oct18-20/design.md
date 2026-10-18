# pipeline-fix-kan-urgent-5-oct18-20

## Context

Found while landing self-review-kan-urgent-5-oct18's gymie and spectre satellites, after the loop
itself landed as agents `fe238ed0` and went FINISHED. Reproduced at `fe238ed0` against the dev
store (read-only):

- `flow record dispatches -change self-review-kan-urgent-5-oct18 -C <agents>` returns the loop's
  five rows (plan completed, fix completed, review-1 fix, fix-1 completed, review-2 clean);
  `-C <gymie>` and `-C <spectre>` return `[]`.
- `scripts/check-loop-landing.sh <gymie> self-review-kan-urgent-5-oct18` exits 1 with
  `PLAN —`, `FIX —` and `REVIEW —` refusals; the same call with `<agents>` exits 0,
  `LOOP-LANDING-OK`.

Blast radius: 11 files cite
<!-- measured: grep -rlE 'check-loop-landing|looplanding' . --exclude-dir=archive --exclude-dir=.git | wc -l @ fe238ed0 --> `check-loop-landing` or `looplanding` outside archives
(`grep -rlE 'check-loop-landing|looplanding' . --exclude-dir=archive --exclude-dir=.git`); the
fix touches three of them plus the guard's test.

## Decisions

### Where the satellite's rows are read

**ID:** fallback-to-agents-project
**Status:** active
**Chosen:** when the pushing repository's project holds no rows for the loop, read them under the main checkout of `FLOW_GUARD_REPO_ROOT` — the agents checkout the shim sits in, the one the hook and `pipeline-fix.sh` are installed from — skipped when that is the pushing repository itself.
**Considered:** filing the rows under every repository the plan touches — `check-loop-dispatch` cannot know the plan's repositories when it admits the plan dispatch; a `-project` argument on the check — the hook would have to know the loop's home, which it cannot; always reading both projects — a pushing repository with its own rows for the loop never needs the second read.

### A change's own branch keeps its own project

**ID:** own-branch-no-fallback
**Status:** active
**Chosen:** the fallback applies to `pipeline-fix-*` and `self-review-*` branches only; a `spectre/<name>` branch reads its own project's rows alone.
**Considered:** applying it to `spectre/<name>` too — a project's change branch would be judged by an agents-side loop sharing its name, gating an ordinary `/flow` landing on rows that are not its own.

### A FINISHED loop still admits its satellites

**ID:** finished-loop-admits-satellites
**Status:** active
**Chosen:** yes, with no code for it — the check reads dispatch rows, never the loop's state record, and the satellites land after the loop's own branch, so they always meet a FINISHED loop.
**Considered:** refusing once FINISHED — every satellite would be unlandable by construction.

### No satellite mode in pipeline-fix.sh land

**ID:** no-satellite-land
**Status:** active
**Chosen:** a satellite lands by its repository's own `## default landing route`, a hand-typed push the hook now admits; `pipeline-fix.sh land` stays agents-only.
**Considered:** a satellite mode in `land` — a second landing implementation for a path the admitted push already covers.

### Unreadable agents project

**ID:** fallback-read-failure-keeps-refusal
**Status:** active
**Chosen:** a failed fallback read keeps the empty rows, so the push is refused — the check still fails closed.
**Considered:** exiting 2 — the first read already proved flowd answers, and a refusal is equally closed.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — the check gates every loop landing push, a hook every operator push passes.
**Considered:** `flow-low` — the diff is small, but a wrong admission silently lands an unreviewed branch.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — the plan holds one task.
**Considered:** none — a single task has no split.

## Open questions

## Live check

`scripts/check-loop-landing.sh /Users/tweety53/Projects/gymie self-review-kan-urgent-5-oct18`, run
from this worktree, exits 0 with `LOOP-LANDING-OK` (it exits 1 with three refusals at `fe238ed0`);
the same call for `self-review-no-such-loop` still exits 1 with `PLAN —`, `FIX —` and `REVIEW —`.
A refusal on the first call means the fallback never ran.
