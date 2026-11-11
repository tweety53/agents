# flow-plan — rationale

Reasoning behind `skills/flow-plan/SKILL.md`: why its lists are shaped as they are and the incidents
its rules were written against. Moved here verbatim from the run-loaded file; **no `/flow-plan` run
loads this file.** Each heading names the source file and the section the passage came from.

## SKILL.md — Research in the session, planning through the planner role

Until 2026-11-11 the session did the thinking itself: it ran `/flow`'s sections C, D and Decide
directly at capture, marked only `plan.session`, and recorded no dispatch (decision
`flow-plan-records-nothing`, kan-1033). A session planned inline that way, with no planner and no
plan designer, and the operator asked for parity: "flow-plan must behave the same way as planning
during the very first flow run." A capture now plans through the same `planner` dispatches as
`/flow`'s creating run (decision `flow-plan-plans-through-planner`); research before the capture
stays in the session because round files need the worktree and dispatch rows need the change.

The five planning rows of the Level 1 table admit `/flow-plan` (decision
`flow-plan-marks-planning-stages`) at the operator's choice, for per-stage timings of a capture,
over keeping `plan.session` the one mark. `plan.session` therefore ends at kickoff: a
`stage begin` closes every open run sharing its session token as `superseded`
(`<agents repo>/stats/internal/store/stageruns.go`), so a `plan.session` still open at the `flow.brainstorm`
begin would be closed `superseded` and the closing `started` end would find nothing open.

## SKILL.md — Capturing a new change

