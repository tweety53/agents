## Context

One prose change across the pipeline contract, `implement.md`'s closed dispatch list, `/flow-fast`'s
guardrail and `/flow-self-review`'s angle 1. No guard, CLI or SPA change.

## 1. The rule's home and reach

- New `## Pipeline defects found mid-run` in `skills/flow-contracts/pipeline.md`, after
  `## Fewest operator actions`, whose no-round-count loop rule it runs under.
- `/flow` loads `pipeline.md` first, so it reaches every `/flow` phase. `/flow-fast` does not load
  it, so its guardrail sentence cites the section.
- Scope: the pipeline is everything `<agents repo>` ships, plus a saved memory encoding its
  behaviour. For this section alone `<agents repo>` is a repository the run was given, overriding
  `rules/agent-baseline.md`'s "a repository you were not given" for that one repository.

## 2. Predicting the class without a plan

- Blast radius: the defective file, its test, and every `grep -rl` hit in `<agents repo>` citing
  the name or section the fix changes.
- `big` when the count is 60 or more (`plan-class.sh`'s own `files>=60` big threshold), when the
  fix reaches outside `<agents repo>`, or when it needs a design choice only the operator can make.
  Everything else (`micro`/`small`/`regular`) is fixed in-run.
- A fix found `big` once under way stops and is deferred the same way.

## 3. The one-shot loop

- Fix subagent → fresh reviewer → parent inline fix or fresh fixer → fresh re-review, until a clean
  review; then land. Every dispatch is fresh; nothing is resumed and no question is parked.
- Each dispatch runs on `opus` at `high` (`flow-high`), keys `pipeline-fix-<k>`,
  `pipeline-fix-<k>-review-<r>`, `pipeline-fix-<k>-fix-<r>`.
- The fix branch lives in its own `<agents repo>` worktree from `origin/<default-branch>` — never
  the main checkout (read-only) and never the change's own worktree.
- Landing: `<agents repo>`'s `## default landing route`, no ask. Merge and push is
  `git push origin fix-<slug>:<default-branch>` after a rebase, then `pull --ff-only` in the main
  checkout so the installed symlinks see the fix.
- The loop runs in the background; the run's own work continues.

## 4. Recording and self-review

- One narrative line per fix: `In-run pipeline fix: <agents sha | deferred> — <defect> (blast radius
  <N> files)`, in `narrative.md` (`/flow`) or `## Session narrative` (`/flow-fast`), both of which
  already reach the self-review bundle; plus one decision bullet in the run summary.
- `/flow-self-review` angle 1 reports a line carrying a sha as fixed and offers no ticket for it; a
  `deferred` line is offered as any other angle-1 finding.

## Decisions

### Where the rule lives

**ID:** rule-home-pipeline-contract
**Status:** active
**Chosen:** a section of `skills/flow-contracts/pipeline.md` — loaded first by every `/flow` run; `/flow-fast` cites it.
**Considered:** amending `rules/agent-baseline.md` so `<agents repo>` is always "given" — reaches every agent, /flow or not, wider than asked; a `docs/briefs/` lesson only — changes no behaviour.

### How a fix lands

**ID:** land-by-default-route
**Status:** active
**Chosen:** `<agents repo>`'s `## default landing route`, without asking — the declared route is the durable authorization.
**Considered:** push the branch and leave merging to the operator — the fix does not reach the running pipeline; ask per fix — a stop the run does not need.

### Which defects are fixed in-run

**ID:** blast-radius-class
**Status:** active
**Chosen:** predicted class from a blast-radius file count, `big` at 60+ files, an outside reach, or an operator-only choice — deferred when `big`.
**Considered:** the agent-baseline cheap-fix test alone — too narrow, the operator asked for a bigger threshold; one module and no contract text — superseded by the operator's class-based framing; `plan-class.sh` on a throwaway task sketch — a mini plan, which the prediction must avoid; unquantified judgment — no checkable number.

### One-shot dispatch

**ID:** one-shot-fix-loop
**Status:** active
**Chosen:** every fix, review, fixer and re-review is a fresh dispatch run to completion; the parent never resumes a child and no child parks a question.
**Considered:** resuming the fix subagent with review findings — the pause-resume cycle the operator ruled out.

### Self-review sees in-run fixes

**ID:** narrative-line-record
**Status:** active
**Chosen:** one narrative line per fix, read by self-review from the bundle it already holds.
**Considered:** a new bundle section or store record — a CLI change for what one existing prose channel carries.

## Open questions
