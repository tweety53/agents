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

**A confidence below 100% is put to the operator, before anything is triaged.** When
`<abs-worktree>/.superpowers/sdd/confidence.json` exists and its `score` is below 100, the
triaging dispatch's first return is `QUESTIONS`, its round file holding one question:
`### Q<n>. The plan check is <score>% confident this plan is executed as written — raise it now?`
Its `**Context:**` lists every entry of the file's `gaps` that closes with `operator` or `running
it`, verbatim, each followed by the suggestion that closes it — the question the operator would be
asked, or what only running the work can show. It has exactly two options, **Raise it** — apply
every suggestion now — and **Later** *(recommended)* — file the estimate and go on. Under the
`## decisions: recommended` mode (**Auto-resolution**,
`skills/flow-contracts/operator-prompts-auto-resolution.md`) the question is not asked and
**Later** is taken. A gap that closes with `operator` is not asked in this round: it is asked
only once the answer is **Raise it**.

- **Raise it** — the fresh final planner that takes the answer triages the report as above, with
  every gap in scope: a `closes with: operator` gap is one more `QUESTIONS` round. It then files
  the estimate with `-disposition raised`.
- **Later** — it defers only the `operator` gaps: each is written into `## Risks` as
  `Accepted: <the gap> — deferred at the confidence question`, the report's other
  findings are triaged as above, and the estimate is filed with `-disposition later`.

**Filing is list, then reuse or create.** Run `flow confidence tags`; for each kind of gap, reuse
a printed tag that fits it, and write a new lowercase kebab-case tag only when none does — a tag
names the kind of gap (`unchecked-premise`, `missing-fixture`), never the change. Then, once:

```bash
flow confidence record -change <name> -stage plan -score <score> -disposition <raised|later> \
  -gap '<one entry of gaps, verbatim>' [-gap …] -tag <tag> [-tag …]
```

`<score>` is `confidence.json`'s, the cold reader's number: it is never re-estimated after the
gaps are closed. Exit 2 is a refused estimate — correct the flag its stderr names and re-run; an
unreachable store warns and exits 0, and the planner goes on.
