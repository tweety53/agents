# pipeline-fix-kan-1037-stats-weekly-costs-dashboard-34

## Why

During `/flow kan-1037-stats-weekly-costs-dashboard`'s implementation run the parent stopped twice
where no stop belongs: it put a visual verifier's three spec questions, each with a recommended
option, to the operator via AskUserQuestion — reading `rules/design-mockups-are-specs.mdc`'s "Ask
rather than invent" as overriding `skills/flow-contracts/operator-prompts-auto-resolution.md` — and
it ended its turn after an in-run fix task citing token cost, with the panel re-run, verify,
handoff and landing still ahead. Both are already banned, but nothing states the run's whole shape
in one place, and the mockup rule reads as an unconditional "ask". The operator asked: "State that
the general flow thru any task must be: interactive planning into implementation+integration till
the task is done, non-stop (unless i say to stop for the operator manual testing after the
implement)."

## What changes

- `skills/flow-contracts/pipeline.md` **Fewest operator actions** opens with the run's shape:
  planning is the only interactive phase; the plan gate's `/clear` handoff is its boundary; from
  the resuming `/flow <name>` the run goes implementation → review → verify → integrate → cleanup
  without stopping; the only operator-chosen stop is `reviewBeforeLanding`; otherwise only **What
  still stops**; no other file's rule licenses asking a question that carries a recommended option.
- `rules/design-mockups-are-specs.mdc`'s core "Ask rather than invent" paragraph yields inside a
  `/flow` run past its plan gate: a question with a recommended option is taken, never asked.
