# Operator prompts

**This file is canonical for the shape of an operator-facing prompt.** It is the one prose shape every approval or choice takes — offered as
options, not as open prose — stated once so every call site stops restating it.

## The shape

A prompt in this shape states:

- the question, with named options
- exactly one option marked (recommended)
- what happens if the operator is silent — the safe default, always the recommended option
- a ⚠ marker in the handoff when that silent default actually fired, or when **Auto-resolution**
  below took it

## The multi-select variant

Some prompts ask the operator to choose any subset of several options, not exactly one. This
variant states:

- the question, with each option listed separately
- that the operator may select any subset of the listed options — none, one, or several
- one explicitly stated default — named by the call site — for what happens if the operator is
  silent

This contract fixes the shape, not the default's polarity: the call site chooses it, and states it
plainly. A safe default may resolve silence to the empty set (an explicit "None" option, marked
recommended) or to the full set (every listed option, silence needing no option of its own to name
it) — whichever matches what the options actually control. The one live multi-select call site,
the self-review filing ask, chooses the empty set: silence selects **None — file nothing**.

## The doctrine

Every call site cites this contract for the mechanics and states only its own question text
and options — never the shape itself.

## Auto-resolution

**During implementation and every fix run, a prompt with a recommended option is never asked: the
run takes the recommended option itself, at once.** The phases are everything from `flow.implement`
to the `IN_PROGRESS` handoff of a creating run — `skills/flow/implement.md`,
`skills/flow/review-panel.md`, `skills/flow/visual-verify.md`, `skills/flow/verify-and-handoff.md`
— and every fix run (`/flow <fix instructions>`, or a plain message run as a fix). The operator
asked for this once, for every flow run: a mid-run question they would answer with the recommended
option costs them a stop for nothing.

It holds for a call site's own prompt and for any question the run frames itself in those phases —
a finding that seems to need a product decision, a Minor's disposition, a departure from a mockup, a
question an implementer's report carries. Frame it with the option the pipeline's own rules favour
marked recommended (for a mockup departure, matching the mockup, per `rules/design-mockups-are-specs.mdc`),
and take that option.

Every auto-resolution:

- is recorded where the call site records an operator's answer — in the review panel, a `flow
  record pass -round <round> -note 'auto-resolved: <question> → <option>'`; a site that records no
  answer records nothing new
- is named in the handoff's `**Auto-resolved:**` line, ⚠-marked, question and option taken, so the
  operator can overrule it afterwards with a fix run
- never repeats: the same prompt arising again for the same subject after its recommended option
  was already taken this run is asked, so an auto-taken **another round** cannot loop

### What still stops

These are asked, or stop with `## Question`, exactly as their call sites state:

- **Planning.** Brainstorm and design questions, the convergence-and-approval confirm, the plan
  review gate, every `/flow-plan` prompt, a fix run's own planning pass (the re-plan-budget and
  where-the-fix-goes prompts in `skills/flow/implement.md`), and a pivot, which alters scope the
  operator approved. The operator scoped auto-resolution to implementation and fix options, not
  planning.
- **No recommended option.** The second model-handshake mismatch, a multiple-match pick, a finding
  recorded unverifiable, and a question the run cannot honestly give one recommended option.
- **Nothing to choose.** A guard's exit 2 (it cannot answer), a command that failed twice, and every
  other `## Question` handback that carries no options.
- **Outward-facing or irreversible actions.** Push, merge, opening a PR, archiving, deleting, and
  every Jira write — the deferred-findings follow-up filing and the self-review filing included.
  The global rules require these confirmed.
- **Outside implementation and fix runs.** The wrong-state override, the plain-message ambiguity
  prompt that decides whether a fix run starts at all, and every integrate and archive prompt.
