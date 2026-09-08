## Context

Every model-bearing dispatch resolves its model through a tier order — session override, then the
project `.flow/project.md` key, then the settings store, then a literal fallback — defined once in
`skills/flow/SKILL.md`'s Model resolution. The resolved value lives only in the dispatcher's head:
nothing in the run's own output states what was resolved or which tier produced it. KAN-381
records the near-miss that follows: the archive-phase dispatcher read the store's raw
`selfReviewModel` field instead of the resolved `SELF_REVIEW_MODEL`, nearly dispatching
self-review on `opus` instead of `fable`. The handshake (`Model:` first-line check) is
structurally blind to this class — a dispatcher holding the wrong value expects the wrong line,
and the subagent honestly matches it.

## Approaches considered

- **State the resolved value at each dispatch site (chosen)** — makes the resolution an
  observable claim the operator can challenge mid-run.
- **A mechanical resolver command (rejected)** — a `flow` CLI subcommand that prints the resolved
  value and tier is the related automation finding KAN-381 points at, not this fix; it changes
  the Go app and ships separately.
- **Fix only `skills/flow/archive.md` step 9 (rejected)** — the stale-value class exists at every
  model-bearing dispatch site; the issue's own scope phrase is "any other model-bearing dispatch
  site".

## Design

**The statement.** Immediately before every model-bearing dispatch, the run's own output carries
one line:

```markdown verified:authored in-tree for this change
<role> model: <value> (<tier>)
```

`<tier>` names the tier that produced the value — `session override`, `project key`, `store`, or
`fallback` — plus three special cases: `fixed literal` for `VERIFY_MODEL`, which is never
resolved; `unknown (agent-defined)` for Bugbot and Security dispatched by their own
`subagent_type`, matching the dispatch record's own `-model` literal so the absence of an
override is itself observable; and the model actually given for a substituted slot. A
handshake-mismatch re-dispatch states again before each re-dispatch.

**Canonical statement.** The rule lives once, in `skills/flow/SKILL.md`'s Model resolution — the
section already canonical for what each model variable is and which tiers resolve it. Each
dispatch site adds one sentence instantiating its own role token and citing Model resolution;
the rule itself is never restated.

**Sites.**

| File | Dispatch | Line |
|---|---|---|
| `skills/flow/archive.md` (step 9) | self-review, plus its mismatch re-dispatch | `self-review model: …` |
| `skills/flow/brainstorm.md` | planner (creating run) | `planner model: …` |
| `skills/flow/implement.md` | conductor; `flow.document-fix` planner | `conductor model: …` / `planner model: …` |
| `skills/flow/review-panel.md` | panel slots (per round); panel-fix subagent | `<slot> model: …` / `panel-fix model: …` |
| `skills/flow/verify-and-handoff.md` | Verify and Visual verification | `verify model: …` |
| `skills/flow-research/SKILL.md` | research subagent | `research model: …` |

**Dogfooding.** None available to this change's own run: the operator instructed it to run
inline with no subagent layer, so it performs no dispatch. The statement's first live exercise is
the next dispatching run.

## Decisions

### State the value in the run's own output, not only in the dispatch record

**ID:** observable-claim-in-output
**Status:** active
**Chosen:** print `<role> model: <value> (<tier>)` to the run's own output immediately before the
dispatch — the operator sees it mid-run, while a challenge can still change the run.
**Considered:** relying on the dispatch record's `-model` field — rejected because the record
faithfully stores whatever wrong value the dispatcher held; the near-miss would have recorded
`opus` with no discrepancy anywhere to see. A post-hoc record audit is not a mid-run check.

### One canonical statement in Model resolution, per-site instantiation only

**ID:** one-canonical-statement
**Status:** active
**Chosen:** the rule and the line's shape are stated once in `skills/flow/SKILL.md`'s Model
resolution; each site adds one sentence naming its role token and citing it.
**Considered:** restating the full rule at each site — rejected as six copies that can drift;
bare citations with no site sentence — rejected because each site must say when in its own
sequence the line prints.

## Open questions

None.
