# kan-490-flow-fast-a-reduced-ceremony-flow-variant-drop — design

## Context

`/flow`'s cost is mostly context and subagent ceremony, not the guards it runs — loading
`skills/flow/*.md` (~10k lines) plus a planner subagent, a conductor subagent, a settings-store
review roster and 20+ guards. That is the right shape for changes big enough to need SDD and
multiple reviewers; it is unnecessary weight for a change small enough to run inline start to
finish. `/myflow-fast` (KAN-111) existed before as a composite command chaining `/myflow-start` →
`/myflow-do` → `/myflow-finish` at their **full** rigor and was folded into `/flow`; `/flow-fast`
here is a different thing — the same three-state pipeline with less ceremony, not the same
ceremony issued as one command.

Two facts about the current `/flow` already narrow the gap this change closes: Minor findings
already defer by default (commit 722341a), and `execution: inline` / a dynamic panel roster
already exist as `## execution mode` / `## review panel` project toggles with a docs-only panel
reduction. `/flow-fast` fixes every one of those choices rather than deciding them per run, and
adds no toggle of its own.

Seeded from `docs/superpowers/research/flow-fast.md` — every decision below is the operator's own
answer from that brainstorm, not a default chosen here.

Design captured for the operator's own record and this task set alone; `spectre/specs/` gains no
new capability from this change (there is no myflow-pipeline capability spec today, and none is
started here).

## Decisions

### Thin router + slim phase files, sharing the contracts

**ID:** flow-fast-shape
**Status:** active
**Chosen:** own `skills/flow-fast/SKILL.md` and short phase files (`brainstorm.md`,
`implement.md`, `review.md`, `finish.md`), sharing `skills/flow-contracts/*` as canonical, the
reviewer prompts under `skills/flow/` (`principles-reviewer-prompt.md`, `simple-reviewer-prompt.md`,
`engineering-principles.md`) and `skills/flow/scripts/` unchanged. Each fast file states the
procedure it runs and cites the contract it follows; nothing is restated from `skills/flow/*.md`.
**Considered:** a profile flag inside `/flow` — rejected, every run still loads the full phase
files, so the context-cost goal is lost; a fully standalone skill that copies and trims the
contracts too — rejected, smallest per-run context but largest drift risk, and it would need an
exemption from this repository's own non-repetition rule.

### One state record, interchangeable with `/flow`

**ID:** flow-fast-state-interop
**Status:** active
**Chosen:** `/flow-fast` writes the same state-file shape and the same `flow.*` stage keys as
`/flow`, marking `-command '/flow-fast'` on every mark; it simply marks fewer keys (skips
`flow.design-approval`). The recorded decision is fixed (`execution: inline`, roster `primary` +
`simple-reviewer`) so a `/flow` run resuming the same change reads a valid record and can finish it
under full rigor if the operator chooses. `/flow-status` renders both commands' changes identically.
**Considered:** a `mode: fast` field that makes `/flow` refuse a change started under `/flow-fast` —
rejected, defeats interop; a separate record shape invisible to `/flow-status` — rejected, splits
the one dashboard into two.

### Model resolution — one call, `DEFAULT_MODEL` only

**ID:** flow-fast-model-resolution
**Status:** active
**Chosen:** `/flow-fast` resolves only `DEFAULT_MODEL` via one `flow settings get` (same
unreachable-store fallback `/flow`'s own resolution block uses, per `skills/flow/SKILL.md`'s
**Model resolution**). It resolves no `PLANNING_MODEL`, `SELF_REVIEW_MODEL` or `VERIFY_MODEL`, and
reads no `## execution mode` / `## implementer model` / `## review panel` project toggle — every
one of those three choices is fixed by design, not decided per run.
**Considered:** reusing `/flow`'s full model-resolution block and ignoring the extra fields —
rejected, the point of the router being slim is that it does not carry logic it never uses.

### Brainstorm — inline, auto-pick, no design gate

**ID:** flow-fast-brainstorm
**Status:** active
**Chosen:** planning runs in the parent session — no planner subagent dispatch. Every options round
in `superpowers:brainstorming`'s checklist auto-picks the recommended option; the parent asks the
operator only when no option is recommended or the request cannot proceed without an answer (a true
ambiguity, ships-something-wrong-if-guessed). Writes `proposal.md`, `design.md`, `tasks.md` in the
exact shape `skills/flow/brainstorm-planner.md` sections B–D define (so `check-unfinished-work.sh`
and a later `/flow` resume both read them), and records the fixed decision (below) instead of
running `skills/flow/brainstorm-planner.md`'s **Decide** roll. Marks `flow.kickoff`,
`flow.brainstorm`, `flow.create-artifacts`, `flow.writing-plans`, `flow.decide`; **never**
`flow.design-approval`. The `IN_PROGRESS` staged-diff review is the only human gate before landing.
**Considered:** keeping one design gate (a short summary plus a yes before implementing) —
rejected by the operator: the value in `/flow`'s brainstorm stage is the plan it produces, and the
operator already reviews the staged diff before landing.

### Implement — inline, TDD, targeted runs only

**ID:** flow-fast-implement
**Status:** active
**Chosen:** workspace isolation via `prepare-workspace.sh` / `check-workspace-isolation.sh`,
exactly as `/flow` does. The parent (no conductor) implements task by task from `tasks.md`, failing
test first, one commit per task carrying the same commit-field shape `/flow` requires
(`**Files:**`, `**Tests:**`, `**Regression:**`, `**Baseline:**`, `**Commit:**` — the format is
kept, `check-task-commit-fields.sh` itself is dropped from the guard set below and never runs).
After each task: that task's own touched-package tests and lint on touched files only, via the
build tool's own selector — never the project's whole `## test` / `## lint` list. A full test or
lint run happens only when the operator's own instruction text asks for it, at any point — never
automatically, not even before handoff. SDD ceremony (spec deltas, the plan-provenance guard) is
dropped entirely; a capability's `spectre/specs/` edit is still planned and committed when a task's
`**Files:**` names one, the writing step is simply not gated by the guard.
**Considered:** a full run before handoff whenever the diff touches shared surface (build config, a
shared library, a module without targeted tests) — rejected, reintroduces the "when is it enough"
judgment call the fixed rule exists to remove; one full run always at handoff — rejected, the
whole point is that full runs are opt-in; dropping TDD — rejected, out of scope, not requested;
dropping `tasks.md` — rejected, would break `check-unfinished-work.sh` and the state-interop
decision above.

### Review — fixed two-slot bundle, Critical/Major only

**ID:** flow-fast-review
**Status:** active
**Chosen:** the roster is always exactly `primary` (briefed against `final-review.diff`,
`proposal.md`, `design.md` and each task's commit fields, on `DEFAULT_MODEL`) + `simple-reviewer`
(briefed by `skills/flow/simple-reviewer-prompt.md` for high-confidence defects only, on `haiku`),
bundled as one dispatch (`-slot primary+simple-reviewer`) — the settings-store reviewer list and
the docs-only reduction are never consulted. Critical and Major findings are fixed inline by the
parent, one round at a time. Every Minor is recorded and deferred, never fixed — no
"trivially easy" exception, unlike `/flow`'s own default (commit 722341a). After a fix, both slots
re-run on the round's delta diff until no Critical or Major is open.
`check-panel-findings-closed.sh` gates the `IN_PROGRESS` handoff exactly as it gates `/flow`'s. The
panel record (`flow record dispatch` / `flow record finding`) keeps `/flow`'s shape so
`flow record render -kind panel` works unchanged.
**Considered:** re-running `primary` alone on the delta — rejected, `simple-reviewer` could
introduce or miss a defect the re-run would then not catch; one round with no re-run — rejected,
leaves a just-introduced Critical/Major unchecked; `simple-reviewer` on `DEFAULT_MODEL` instead of
haiku — rejected, the point of a fixed compact bundle is the cost floor; honouring the project's
`## review panel` toggle — rejected, `/flow-fast`'s roster is fixed by design, not dynamic.

### Guard set — seven kept, the rest dropped outright

**ID:** flow-fast-guards
**Status:** active
**Chosen:** `check-unfinished-work.sh`, `check-base-moved.sh`, `check-finish-preflight.sh`,
`check-cleanup-complete.sh`, `check-workspace-isolation.sh`, `check-worktree-processes.sh` and
`check-panel-findings-closed.sh` are presence-checked and run, per **Guard presence check**
(`skills/flow-contracts/pipeline.md`). Every other guard `skills/flow/scripts/` carries —
`check-panel-citation-trigger.sh`, `check-panel-diff-size.sh`, `check-panel-docs-only.sh`,
`check-panel-fix-single-dispatch.sh`, `check-panel-reproducers.sh`, `check-plan-shape.sh`,
`plan-class.sh`, `check-spec-reach.sh`, `check-task-commit-fields.sh`, `check-visual-trigger.sh`,
`check-visual-verification.sh`, `commit-split.sh`, `gather-dispatch-context.sh`,
`gather-self-review-context.sh`, `mutate-and-verify.sh`, `plan-dispatch-bundles.sh`,
`resolve-visual-screenshots.sh`, `run-reproducer.sh` — is dropped for `/flow-fast`, not hand-run:
`/flow-fast`'s own phase files never cite them.
**Considered:** state and git safety only, dropping `check-panel-findings-closed.sh` too —
rejected, would let a Critical slip past unreviewed into a handoff, which no amount of "faster"
justifies; keeping the full guard set — rejected, defeats the point of a reduced-ceremony command.

### Finish — the same two contracts, run 2 minus two steps

**ID:** flow-fast-finish
**Status:** active
**Chosen:** a bare `IN_PROGRESS` `/flow-fast` invocation follows
`skills/flow-contracts/finish-contract-run1.md` and `finish-contract-run2.md` exactly as `/flow`
does — preflight, the unfinished-work gate, the landing question (or a project's `## default
landing route`), session preservation, the two commits, the landing routes; then run 2's merge
verification, archive move, archive commit, cleanup (step 5), write `FINISHED` (step 8), push and
land (step 10), remove the landing worktree (step 11) — **except** run 2 skips step 7
(`check-cleanup-complete.sh`, the verify-cleanup pass) and step 9 (the self-review subagent)
entirely for a `/flow-fast` run. Cleanup itself (step 5) still runs; only its separate verification
pass does not.
**Considered:** dropping self-review only, keeping verify-cleanup — rejected by the operator,
both are the same category of "slow, optional tail" the fast variant exists to cut; also dropping
session preservation — rejected, session preservation protects against losing in-flight work on a
crash, which is not ceremony.

### Drop `/rename` and `/color` from `/flow` and the pipeline contract

**ID:** flow-fast-drop-rename-color
**Status:** active
**Chosen:** remove the printed `/rename <change-name>` / `/color cyan` lines from
`skills/flow/SKILL.md` (announce-line block) and from **Handoff output** /
**The tab commands, printed at the start of a run** in `skills/flow-contracts/pipeline.md`, and
delete the **The tab commands, printed at the start of a run** rationale section from
`skills/flow-contracts/pipeline-rationale.md`. `/flow-fast` never prints them (nothing to remove
there — it never had them). The archived KAN-26 design and artifact under `docs/superpowers/` that
introduced the two lines are history and are left as they are.
**Considered:** keeping the lines in `/flow` and only omitting them from `/flow-fast` — rejected
by the operator: both are deprecated, not `/flow-fast`-specific.

### Raise the dynamic plan-class thresholds

**ID:** flow-fast-plan-class-thresholds
**Status:** active
**Chosen:** raise `scripts/plan-class.sh`'s class thresholds and the compact-roll cutoffs `
skills/flow/brainstorm-planner.md`'s **Decide** section interprets them with:

| Value | Now | New |
|---|---|---|
| small | `tasks≤5` and `files≤12` and `repos=1`, no migration, no spec | `tasks≤10` and `files≤25`, other conditions unchanged |
| big | `tasks≥15` or `files≥40` or (`repos>1` and `tasks≥8`) or (`migration` and `tasks≥8`) | `tasks≥30` or `files≥80` or (`repos>1` and `tasks≥15`) or (`migration` and `tasks≥15`) |
| regular | everything else | everything else, unchanged |
| compact-roll cutoff | `<70` small, `<30` regular/big | `<90` small, `<60` regular/big |
| experimental-roll cutoff | `<30` every class | unchanged |
| static-grouping (bundle) roll cutoff | `bundle<30` | unchanged |

This changes every `/flow` and `/flow-fast` run's dynamic decision (where the `## execution mode`
/ `## review panel` toggles are `dynamic`), not only `/flow-fast`'s own — it is independent of the
fast skill itself, in scope of this change by the operator's own decision (note section 12). The
archived `kan-472-flow-dynamic-review-panel-roster-repo-scoped/design.md` is frozen and historical;
it is cited nowhere in the edited files and stays exactly as it is — only `plan-class.sh`, its own
header comment, `test-plan-class.sh`, and `brainstorm-planner.md`'s **Decide** section change.
**Considered:** leaving `test-plan-class.sh`'s existing boundary cases as regression fixtures at
their old numbers — rejected, they assert the *old* small/big boundary and would then assert the
wrong behavior; each boundary case moves to the new numbers instead.

## Open questions

None — every question the brainstorming round raised was answered by the operator's opening list
and the eight follow-up decisions above; none was deferred.
