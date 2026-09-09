# kan-472-flow-dynamic-review-panel-roster-repo-scoped

**Jira:** KAN-472 — joins KAN-475 (linked to KAN-472, closed with it) and KAN-478 (linked as a
duplicate, closed with it).

## Why

Every `/flow` run today makes the same per-run choices regardless of the change in front of it:
the review panel is the harness-wide settings-store list on `DEFAULT_MODEL`, the implementer and
fixer run on `DEFAULT_MODEL`, and implementation always goes through a conductor subagent that
dispatches an implementer per bundle, a slot per reviewer and a fix subagent per round. A two-task
prose edit and a fifteen-task schema change pay the same panel and the same subagent tree. Nothing
records why a roster, a model or an execution shape was chosen, so nothing can be tuned from data:
there is no per-reviewer catch-rate view, no cost comparison between execution shapes, and no way
to trial a new reviewer prompt against the persistent slots. Every finding, Minor included, blocks
the handoff until fixed, so a run spends fix rounds on cosmetic findings at the same price as real
defects. A six-role panel is six subagents, and a wave of ready bundles is one implementer each;
nothing bounds either, and the two review roles that are agent definitions rather than prompts
cannot share a dispatch with anything.

## What changes

One change, eight task groups, in dependency order:

1. **Store and CLI.** A `decisions` table (one JSONB row per run) and `flow record decision`; an
   `effort` column on `dispatches` and `-effort` on `flow record dispatch begin`; a third terminal
   finding status, `deferred <reason>`, accepted only on a Minor.
2. **Toggles, classifier and decision tree.** Three `.flow/project.md` keys — `## execution mode`,
   `## implementer model`, `## review panel` — each `default` or `dynamic`. The planner classifies
   the plan (small / regular / big) from mechanical inputs with a one-step-up override, rolls the
   compact and experimental samples from the change name, and appends a `## Decision` block to its
   `## Plan` return; the parent prints it read-only and records it before anything is dispatched.
3. **Enforcement.** Nine `flow-<model>-<effort>` agent definitions installed by `setup.sh`; every
   dispatched role opens with the `Model:` handshake; a first mismatch is a fallback and one retry,
   a second stops with a question.
4. **Dynamic panel.** `review-panel.md` dispatches the decided roster with per-slot model and
   effort, the decided rerun policy, and at most one experimental slot whose prompt lives under
   `skills/flow/experimental/` and whose id carries the `exp-` prefix everywhere it is recorded.
5. **Inline execution.** With `## execution mode` dynamic and a small or regular class, the parent
   implements and fixes itself — no conductor, no implementer, no fix subagent — under a context
   ceiling that stops with a clear-and-resume handoff; panel slots and the verifier stay subagents.
6. **Severity-gated minors.** Critical and Major are always fixed; a Minor is fixed when easy or
   when judgment says so, otherwise deferred with a reason and listed in the handoff — in both
   default and dynamic mode.
7. **Stats views.** `reviewers` (per-slot severity counts, experimental slots badged with their
   prompt's description) and `decisions` (every decision row against its run's wall-clock, tokens,
   cost, findings and fallbacks).
8. **Bundling** (KAN-478). At most two review dispatches per round, each one to three roles, on
   both toggle values; `bugbot` and `security` become prompt-driven roles so every role can share
   a dispatch; grouping is rolled — 30% the class's static table, 70% the planner's own choice
   with a recorded reason; a bundle is one `dispatches` row under a `+`-joined slot, findings
   keep their role, the `reviewers` view splits the slot; implementer bundles merge into
   planner-decided groups with at most two in flight per wave; the `decisions` view shows the
   grouping.

`design.md` is canonical for every rule, table and shape named above.
