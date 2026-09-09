# kan-472 — dynamic execution mode, implementer model and review panel — design

Date: 2026-09-09. Change: `kan-472-flow-dynamic-review-panel-roster-repo-scoped` (KAN-472, joined
with KAN-475). Canonical artifact form: this change's `spectre/changes/<name>/design.md`
(every rule, table, shape, Decisions and Open questions live there); this document records what
was presented at the approval gate and nothing the canonical file already states.

## Presented and approved, in the order asked

1. The planner decides at the end of writing-plans and returns a `## Decision` block with its
   `## Plan`; the parent prints it read-only, records it, then proceeds. No gate.
2. Inline = the parent implements and fixes itself; panel slots and the verifier stay subagents.
3. Class from mechanical plan inputs with a one-step-up planner override.
4. Thresholds small ≤5/≤12/1 repo/no migration/no spec; big ≥15 tasks or ≥40 files or
   (>1 repo ∧ ≥8) or (migration ∧ ≥8); a first, lower proposal was rejected as too low.
5. The v1 tree with per-slot model and effort, compact and experimental rolls hashed from the
   change name, delta rerun for small/regular and full for big.
6. Nine `flow-<model>-<effort>` agent definitions, universal `Model:` handshake, fallback + one
   retry, second mismatch stops with a question.
7. One `decisions` JSONB row per run; `exp-` slot prefix; prompts under `skills/flow/experimental/`.
8. Three independent toggles `## execution mode`, `## implementer model`, `## review panel`.
9. Inline context ceiling 250k before a bundle / 400k before the panel / 6 bundles fallback;
   clear-and-resume handoff; the stored decision is reused on resume.
10. `deferred <reason>` finding status, Minor-only; 10% is guidance, never computed.
11. Both `reviewers` and `decisions` stats views.
12. One change, seven task groups; KAN-475 linked to KAN-472 and closed with it.

Convergence closed with no open questions.

## Post-approval amendment

After the gate above, per operator instruction, Group 8 (bundling — the review-panel dispatch cap,
implementer merging, and retiring the `subagent_type` Bugbot/Security mechanism; KAN-478 joined)
was brainstormed and added. Detail: `design.md`'s decision records `two-dispatch-cap-everywhere`
through `kan-478-joined-eighth-group`.
