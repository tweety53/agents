## Why

KAN-36 asks that brainstorming run "more rounds if needed, everything should be clarified". Today it
cannot: `skills/myflow-start/SKILL.md` invokes superpowers:brainstorming "in full: checklist items
1–8", and that upstream skill's flow is linear — explore, ask, propose, present, approve. Its only
loop is *design revision*, which routes a **correction** back to the design. Nothing routes a
**question** back to the questions.

`myflow-planning-gate` already requires that an unresolved question be put to the operator rather
than assumed. That covers a question the command **notices**; it says nothing about a question a
previous answer **created**, because the stage has no way to reach a second round. The gap is not
that questions get assumed away — it is that there is nowhere to ask them.

Two further asks arrived during this change's own planning run, both raised by the operator at gates
this change is about:

- **`TO DO URGENT` should be read as `To Do`.** KAN-36 sat at that status, the unrecognised-status
  carve-out fired, and the answer was that the status simply means To Do.
- **`/myflow-start`'s git boundary contradicts itself three ways** — the skill says commit the design
  doc, `pipeline.md` says no git actions, and `myflow-command-surface` says in normative language
  that no command other than `/myflow-finish` and `/myflow-do`-with-a-PR may commit at all. Seven
  archived changes committed a design doc from `/myflow-start`, so the requirement has been violated
  by every change this pipeline has planned.

The full brainstorming record is
`docs/superpowers/specs/2026-08-01-kan-36-more-brainstorming-rounds-if-needed-design.md`.

## What Changes

- **Brainstorming converges instead of passing once.** After every planning-stage exchange — a
  question round, a design-section approval, the spec review — `/myflow-start` applies one test: does
  it now hold a question it cannot answer? While the answer is yes it opens another round. One test
  rather than a rule per gate is deliberate: a design section that reopens something and an answer
  that opens something are the same event, so no gate can later be added that bypasses the loop by
  not being enumerated.
- **Two prompts bound the loop, and they recommend opposite courses.** The **confirm** fires when the
  test comes back empty: the command states what it believes settled and asks whether anything is
  still unclear, recommending *move on* — honest because the prompt is unreachable while it holds
  anything. The **soft-bound offer** fires from the third round onward in place of silently opening
  one: it shows what is still open and recommends *another round* — honest for the mirror reason,
  the same shape as finish run 1's unfinished-work gate where **Stop** is recommended. The threshold
  `3` is a tuned value in `skills/myflow-start/SKILL.md`; the requirement says only that later rounds
  are offered rather than taken.
- **Every planning effort level runs the loop.** A level changes how many questions are grouped into
  a round, never whether another opens. `myflow-planning-effort` already forbids a level switching a
  gate off, and a level able to end the loop early would be exactly that.
- **A revision round re-enters the loop, scoped** to what the feedback reopened and whatever that
  opens in turn. Settled parts are not re-brainstormed — otherwise the cheap path becomes expensive
  and operators stop taking it, which would remove the loop from the place a half-clarified plan most
  reliably surfaces.
- **`design.md` gains `## Open questions`**, beside `## Decisions` and shaped like it, with an
  immutable ID. A question still open when the stage ends is recorded there; a later round that
  answers it sets `**Status:** answered by <decision-id>` and adds the decision, and the entry is
  never deleted. Stopping early stays legitimate and stops being invisible.
- **The `STARTED` handoff gains `Open questions: <N> | none`**, counted from entries still at
  `**Status:** open`. It is derived from `design.md`, so it is regenerable rather than run-only and
  `/myflow-status <name>` renders it from the same template.
- **`TO DO URGENT` becomes a `To Do` synonym.** The four ordered names become four **positions**
  carrying names. The unrecognised-status ask survives, narrowed to a status matching no recognised
  name including the synonyms. This adds an enumerated **name**, which is the mechanism kan-26
  already used for the follow-up join search — the `statusCategory` inference it rejected stays
  rejected, and the count of interactive Jira carve-outs stays at two.
- **`/myflow-start` stages its planning artifacts and never commits them.** The design document
  under `docs/superpowers/specs/` and the change directory under `openspec/changes/<name>/` stay
  staged, exactly as `myflow-command-surface`'s existing prohibition already requires — the
  three-way contradiction is resolved by making that prohibition true, not by carving an exception
  into it. The worktree-visibility gap the commit exception would have closed — a plan that is only
  staged is absent from the worktree `/myflow-do` builds, and nothing copies it across — is recorded
  in `design.md` under `## Open questions` rather than addressed here, and is left for a later
  change.

## Impact

- **Affected specs:** `myflow-planning-gate` (added), `myflow-handoff-output` (modified),
  `myflow-jira-projection` (modified). No delta touches `myflow-command-surface`: the stage-only
  resolution complies with the requirement already in `openspec/specs/myflow-command-surface/spec.md`,
  so restating it as a delta would be noise in the permanent spec tree.
- **Affected code:** `skills/myflow-contracts/pipeline.md`, `skills/myflow-start/SKILL.md`,
  `skills/myflow-contracts/jira-integration.md`.
- **Not affected:** `/myflow-do`, `/myflow-finish` and `/opsx:explore` are untouched.
  `skills/myflow-status/SKILL.md` needs no edit — it renders the handoff block by citing the
  template rather than copying it, so the new line flows through. No new guard script: the To Do
  synonym set is stated once and cited, which `scripts/check-references.sh` already checks.
- **Behavioural change for operators:** `/myflow-start` will ask more often before leaving
  brainstorming, recording anything still open under `## Open questions` rather than assuming it
  away. Its git behaviour is unchanged: it stages the planning artifacts exactly as before, and
  `/myflow-finish` run 1's planning commit continues to carry them.
