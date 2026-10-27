## Context

- `setup.sh global` renders the core excerpt of every `rules/*.mdc` carrying `alwaysApply: true`
  into the managed blocks of `~/.claude/CLAUDE.md` and `~/.zcode/AGENTS.md`; subagents get the
  rules through `rules/agent-baseline.md`'s table. No existing rule states the operator's priority
  for ordinary sessions — `be-brief.mdc` (turn ends), `design-mockups-are-specs.mdc` (`/flow`-scoped
  auto-resolution) and `pipeline.md`'s **Fewest operator actions** each cover a slice.
- A survey of `skills/` found about forty asks or stops past planning, and about sixty call sites
  where a guard's exit 2 or an environment fault "stops the run". **Auto-resolution** covered only
  implementation and fix runs, and three review-panel prompts auto-resolved to a recommended
  **Stop**.
- One change: the rule and the contracts that contradict it must land together, or a run reads two
  instructions that disagree.
- Added during planning by the operator: `xhigh` flow agent definitions for later use, dispatched
  by nothing yet; this session's planning dispatches ran at `xhigh`, for this session only.
- `check-verbatim-moves.sh` fails every added or reworded run-loaded sentence not listed in this
  change's `verbatim-moves.txt`; the list is committed with the plan, since an implementer never
  writes `spectre/changes/`.

## Decisions

### A new rule file, not an extension of an existing rule

**ID:** new-rule-work-end-to-end
**Status:** active
**Chosen:** `rules/work-end-to-end.mdc`, titled "Work end to end — stop only to plan or to test by hand" — one rule, one subject, rendered into every session and cited by one `agent-baseline.md` row.
**Considered:** extending `be-brief.mdc`'s "never let a report become a stop" — brevity and autonomy are separate subjects, and the brevity rule would carry a stop list it cannot state briefly; extending `flow-manual-review.mdc` — that rule binds `/flow` only, and the operator asked for every session; the name `work-until-it-works.mdc` — longer and less specific than what it states.

### Auto-resolution covers everything past planning

**ID:** auto-resolution-past-planning
**Status:** active
**Chosen:** widen **Auto-resolution**'s scope from implementation and fix runs to everything from `flow.implement` to the run's final handoff — integrate, cleanup, the self-review pass, `/flow-fast` past its plan — and the prompts that open a run; delete the **Outside implementation and fix runs** bullet of **What still stops**; add a **Manual testing the operator asked for** bullet.
**Considered:** listing each integrate and cleanup prompt as auto-resolved one by one — the prompts that carry a recommended option and are not outward-facing are few, and a scope statement covers future ones too; keeping the carve-out — it is the stop the operator asked to remove.

### A tooling failure is cured in the run, by one precedence paragraph

**ID:** tooling-failure-precedence
**Status:** active
**Chosen:** one paragraph in **Fewest operator actions** (`skills/flow-contracts/pipeline.md`): a guard's exit 2, an erroring script, a stack that will not start, a stale environment or tool is cured in the run and the step re-runs; a phase file's stop on such a failure, and a prompt whose recommended option is itself a stop, hold only for a cause out of the run's reach. `pipeline.md` already wins wherever a phase file disagrees. **What still stops**' **Nothing to choose** becomes **Out of the run's reach**.
**Considered:** rewording each of the ~60 call sites — sixty rewordings of one rule, each a `verbatim-moves.txt` line and a chance to drift; putting the paragraph in `operator-prompts-auto-resolution.md` alone — that file loads only when a prompt arises, so integrate's and cleanup's failure stops would never read it.

### The moved-base prompt recommends Continue

**ID:** moved-base-continue
**Status:** active
**Chosen:** **Continue — review as is** becomes the recommended and silent option of the review panel's exit-3 prompt; integrate's sync onto the moved base settles the overlap before anything lands.
**Considered:** **Rebase now** — exit 3 means the rebase is not confirmed conflict-free, and a panel-stage conflict is never resolved there, so it would end the run; **Stop** — the stop being removed.

### The bundle-failure prompt recommends fixing the build

**ID:** bundle-failure-fix-then-dispatch
**Status:** active
**Chosen:** **Fix the build, then dispatch** — the run cures the cause, re-runs the gather and dispatches with the bundle; only a cause out of its reach closes the stage `-outcome stopped`. The five phrases `check-dispatch-paragraphs.sh` pins stay.
**Considered:** **Continue without the bundle** as the recommendation — every reviewer loses its context, a quality loss the run can avoid by fixing the build.

### A second wall-clock breach re-dispatches

**ID:** wall-clock-redispatch
**Status:** active
**Chosen:** **Re-dispatch it again** is recommended and silent; a third breach of the same slot is asked under **Auto-resolution**'s never-repeats clause, which bounds the loop.
**Considered:** **Proceed without the slot** — drops a reviewer's findings from the panel; **Stop the run** — the stop being removed.

### A self-review product defect continues into its fix run

**ID:** product-defect-continues
**Status:** active
**Chosen:** a product defect the in-run pass meets closes `flow.self-review` `-outcome stopped`, lands nothing, and the run continues straight into the fix run that defect names, in the same invocation (`/flow <name> <defect>`, `/flow-fast <name> <defect>`); the `## Landing stopped` block is deleted.
**Considered:** keeping the stop as the boundary of the fix run's planning (a design agent's recommendation) — the fix run's own planning asks still reach the operator wherever the project asks them, so the stop added a `/clear` and a retyped command and nothing else.

### No rating-only call

**ID:** rating-only-beside-an-ask
**Status:** active
**Chosen:** a self-review pass with nothing to file or mine makes no call; the rating is asked only beside a filing or mining ask.
**Considered:** dropping the rating everywhere — beyond the stop the operator asked to remove; their standing pause on rating and mining stays where they set it.

### The xhigh re-plan approval ask is struck

**ID:** xhigh-replan-ask-struck
**Status:** active
**Chosen:** where a `high` re-plan still cannot produce a design that holds, the next re-plan changes approach again under **Fewest operator actions** — never a stop; the `xhigh` approval ask and its filing fallback go.
**Considered:** keeping the ask — a stop past planning; dispatching `xhigh` without asking — the operator said the `xhigh` definitions are not to be used yet.

### `/flow-fast` without `## handoff` lands

**ID:** flow-fast-handoff-absent-lands
**Status:** active
**Chosen:** `## handoff` absent behaves as `none`; only an explicit `required` stops after the change summary — the operator's explicit request to test by hand.
**Considered:** keeping absent as `required` — a stop the operator never asked for.

### Stale three-course pointers are corrected

**ID:** stale-three-courses
**Status:** active
**Chosen:** `skills/flow/integrate.md`, `skills/flow-contracts/finish-contract-run1.md` and the `STOP: unfinished-work` message in `stats/internal/guard/integratechange_prepare.go` point at **Finish it — never a prompt** instead of a three-course prompt the gate never offers.
**Considered:** leaving them — the Go message prints the stale instruction at runtime and steers a run into the very prompt this change removes; deleting the gate's dead `## The three courses` section too — other files cite its headings, and it already states it is never offered.

### xhigh agents exist and are refused until wired

**ID:** xhigh-agents-defined-refused
**Status:** active
**Chosen:** `agents/flow-xhigh.md` and `agents/flow-planner-xhigh.md` (`effort: xhigh`), their `installedcitations.go` rows and the setup test's count; `hooks/enforce-agent-baseline.py` denies a dispatch on either type. `recordEfforts` keeps refusing `xhigh` until the change that dispatches them adds it, as `43125acf` removed it.
**Considered:** the definitions alone — nothing would stop a hand-picked dispatch, as `flow-high` once was; accepting `xhigh` in `recordEfforts` now — no dispatch records it yet, and it rewrites messages `decision_test.go` and `record_test.go` pin.

### Outward-facing and irreversible actions stay confirmed

**ID:** outward-facing-still-asked
**Status:** active
**Chosen:** pushes, merges, pull requests, Jira writes and deletions stay asked where their call sites ask, as do choices with no honest recommendation and causes out of the run's reach.
**Considered:** authorising every outward-facing action — the global confirmation rules require them confirmed, and the operator asked to remove recommended-pick stops, not confirmations.

### This session's planning ran at xhigh

**ID:** planning-xhigh-this-session
**Status:** active
**Chosen:** the operator's instruction "use xhigh for the planning — only in this session" set this run's Explore and Plan dispatches to `xhigh`; recorded in `decision.json`'s `overrides`, never persisted.
**Considered:** making it a default — the operator scoped it to this session.

## Open questions

## Recommended approach

The rule states the principle once and the `/flow` contracts are aligned to it: the exception list
lives in **What still stops**, one precedence paragraph in **Fewest operator actions** covers every
tooling-failure stop, and only the prompts whose recommended course was itself a stop are edited at
their call sites. Rejected alternatives: per-call-site rewording (`tooling-failure-precedence`),
extending an existing rule (`new-rule-work-end-to-end`), keeping the product-defect stop
(`product-defect-continues`), using or recording `xhigh` now (`xhigh-agents-defined-refused`).

## Critical files

- `rules/work-end-to-end.mdc` — the new rule (core excerpt rendered into every session).
- `rules/agent-baseline.md` — its row, read by every subagent.
- `skills/flow-contracts/operator-prompts-auto-resolution.md` — **Auto-resolution**'s scope and **What still stops**.
- `skills/flow-contracts/operator-prompts.md` — when the auto-resolution file loads.
- `skills/flow-contracts/pipeline.md` — **Fewest operator actions**, the precedence paragraph.
- `skills/flow-contracts/project-configuration.md` — the `## decisions` and `## handoff` rows.
- `skills/flow/SKILL-rationale.md` — the reasoning entry.
- `skills/flow/review-panel.md` — the three prompts whose recommended option was a stop.
- `skills/flow-contracts/known-bugs.md`, `skills/flow/integrate.md`, `skills/flow-fast/SKILL.md` — the product-defect continuation.
- `skills/flow-fast/SKILL.md`, `commands-claude/flow-fast.md` — the `## handoff` default.
- `skills/flow-self-review/SKILL.md` — the rating-only call and the `xhigh` ask.
- `skills/flow-contracts/finish-contract-run1.md`, `stats/internal/guard/integratechange_prepare.go` — stale three-course pointers.
- `agents/flow-xhigh.md`, `agents/flow-planner-xhigh.md`, `stats/internal/guard/installedcitations.go`, `stats/internal/setuptest/agents_test.go`, `hooks/enforce-agent-baseline.py`, `stats/internal/repotest/enforce_agent_baseline_hook_test.go` — the `xhigh` definitions and their refusal.

## Existing code to reuse

- `setup.sh` `always_on_rules` / `render_managed_block` / `install_agents` — render and link the new rule and agent files with no change.
- `hooks/enforce-agent-baseline.py` `flow_high_reason` — the shape `flow_xhigh_reason` mirrors.
- `43125acf^:agents/flow-xhigh.md` — the definition shape, `effort: xhigh` included.
- `rules/design-mockups-are-specs.mdc` — the precedent for a rule core citing **Auto-resolution** across a line break.

## Verification

- `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh`, `scripts/check-installed-citations.sh`, `scripts/check-ask-silence.sh`, `scripts/check-markdown-integrity.sh` and `scripts/check-dispatch-paragraphs.sh` exit 0 on the branch.
- `cd stats && go test ./internal/setuptest -run TestSetupAgents`: a sandboxed `setup.sh global` links eight `flow-*` definitions, `flow-xhigh.md` and `flow-planner-xhigh.md` among them.
- `cd stats && go test ./internal/repotest -run TestEnforceAgentBaselineHook`: a dispatch on `flow-xhigh` or `flow-planner-xhigh` is denied; every `flow-high` row keeps its verdict.
- A sandboxed render — `HOME=$(mktemp -d) ./setup.sh global` from the worktree — writes `<!-- rule: work-end-to-end.mdc -->` and the core excerpt into `$HOME/.claude/CLAUDE.md`.
- After landing: `./setup.sh global` and `make -C stats restart` from the main checkout; `~/.claude/CLAUDE.md` carries the new core; `scripts/check-installed-rules.sh` exits 0 there.

## Live check

None — the change touches no daemon, store or scheduler; the hook and the rendered managed block are exercised by `## Verification` above, against a sandboxed `$HOME`.
