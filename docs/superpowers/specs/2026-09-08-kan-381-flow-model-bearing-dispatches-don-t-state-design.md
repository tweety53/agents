# kan-381 — flow: model-bearing dispatches don't state the resolved value before dispatching — design

Approved design from the brainstorming dialogue of 2026-09-08. `design.md` in the change root is
derived from this document; this file is the brainstorming record.

## Problem

KAN-381 (self-review finding from KAN-380, problems-and-fixes angle): at archive time the
dispatcher read the settings store's raw `selfReviewModel` value instead of re-resolving
`SELF_REVIEW_MODEL` through the three-tier order (session override, project key, store, literal
fallback), nearly dispatching the self-review subagent on `opus` instead of the resolved `fable` —
caught only because the operator asked "why opus?" mid-run.

The handshake cannot catch this class of error: the dispatcher believed the wrong value was
right, so the subagent's honest `Model:` first line matched the wrong expectation and the run
would have proceeded silently on the wrong model.

## Approaches considered

- **State the resolved value at each dispatch site (chosen)** — makes the resolution an
  observable claim the operator can challenge mid-run.
- **A mechanical resolver command (rejected)** — a `flow` CLI subcommand printing the resolved
  value and tier is the related automation finding the issue points at, not this fix; it changes
  the Go app and ships separately.
- **Fix only `skills/flow/archive.md` step 9 (rejected)** — the stale-value class exists at every
  model-bearing dispatch site; the issue's own scope phrase is "any other model-bearing dispatch
  site".

## Design

Immediately before every model-bearing dispatch, the run's own output carries one line:

```
<role> model: <value> (<tier>)
```

`<tier>` names the tier that produced the value — `session override`, `project key`, `store`, or
`fallback` — plus three special cases: `fixed literal` for `VERIFY_MODEL`, which is never
resolved; `unknown (agent-defined)` for Bugbot and Security dispatched by their own
`subagent_type`, matching the dispatch record's own `-model` literal; and the model actually
given for a substituted slot. A handshake-mismatch re-dispatch states again before each
re-dispatch.

The rule lives once, in `skills/flow/SKILL.md`'s Model resolution; each dispatch site adds one
sentence instantiating its own role token and citing Model resolution. Sites: the planner
dispatch (creating run and `flow.document-fix`), the conductor, every review-panel slot per
round, the panel-fix subagent, both verifiers, the self-review dispatch (step 9 and its
re-dispatch), and `/flow-research`'s research subagent.

A decision this run itself cannot exercise: the change ships the statement for future runs to
print — this change's own run is inline (operator instruction) and performs no dispatch.
