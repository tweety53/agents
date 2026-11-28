# Plan triage

Loaded only by the planner dispatch that triages a plan check (**Plan check**, step 3,
`skills/flow/brainstorm.md`).

**A dispatch whose prompt names a `plan-check.md` report triages it.** It reads the report,
`proposal.md`, `design.md` and `tasks.md`; it opens a file under
`<abs-worktree>/.superpowers/sdd/planner/` only for a finding whose fix turns on a round's answer.
Fix in the artifacts a finding with one obvious fix. A finding whose fix would change a
`## Decisions` entry returns `QUESTIONS`, the next `round-<r>.md` quoting the finding verbatim
beside the options. Write each finding still standing into `design.md`'s `## Risks` as
`Accepted: <the finding> — <why>` — a contest of the reviewer's reading names the path:line it
rests on — re-run `spectre validate`, and return `PLAN-WRITTEN`.

**A `closes with: plan` gap is fixed, never deferred.** The triaging planner applies it in the
artifacts and re-runs `spectre validate`, whatever the score and whatever the mode, `## decisions:
recommended` included; it is never written to `## Risks` as `Accepted`.

**A `closes with: running it` gap is a probe step in the plan.** The planner writes the gap's
probe and its fallback into the task it concerns, so the implementer runs it on reaching that
task, and also writes the gap into `## Risks` as `Accepted: <the gap> — <why>`.

**A confidence below 100% is raised, never asked.** When
`<abs-worktree>/.superpowers/sdd/confidence.json` exists and its `score` is below 100, **Raise it**
is taken automatically, in every mode: no question asks whether to raise it. The triaging dispatch
triages the report as above with every gap in scope, applying the suggestion that closes each gap;
a `closes with: operator` gap is a `QUESTIONS` round asking the operator that gap's question. It
then files the estimate with `-disposition raised`.

**Filing is list, then reuse or create.** Run `flow confidence tags`; for each kind of gap, reuse
a printed tag that fits it, and write a new lowercase kebab-case tag only when none does — a tag
names the kind of gap (`unchecked-premise`, `missing-fixture`), never the change. Then, once:

```bash
flow confidence record -change <name> -stage plan -score <score> -disposition raised \
  -gap '<one entry of gaps, verbatim>' [-gap …] -tag <tag> [-tag …]
```

`<score>` is `confidence.json`'s, the cold reader's number: it is never re-estimated after the
gaps are closed. Exit 2 is a refused estimate — correct the flag its stderr names and re-run; an
unreachable store warns and exits 0, and the planner goes on.
