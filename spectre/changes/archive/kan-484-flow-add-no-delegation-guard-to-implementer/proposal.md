# kan-484-flow-add-no-delegation-guard-to-implementer

## Why

KAN-483 closed the conductor's Agent-tool dispatches to four rows — implementer, panel bundle,
panel-fix, verifier (**Dispatch sites — the conductor's closed list**, `skills/flow/implement.md`).
That rule binds the conductor and stops one level down: each of those four children is dispatched
`general-purpose` or `flow-<model>-<effort>` with the full tool set, `Agent` included, and nothing
in its own dispatch prompt tells it not to fork. On the kan-30-group-workouts-final-verification
fix run (2026-09-10) the Tasks 17–18 implementer spawned an unrecorded `general-purpose` grandchild
that spent over ten minutes reading admin credentials in `api.ts`, out of scope for its task; it was
stopped before it wrote anything. It is the KAN-449 failure shape — unrecorded subagents outside
the closed list — recurring one level further down than the KAN-483 fix reaches.

Every other obligation the pipeline needs a leaf agent to honour is a mandatory dispatch paragraph
propagated verbatim (FOREGROUND BUILDS, TOOLS, MODEL HANDSHAKE, REPRODUCE, DON'T READ, …), and
`scripts/check-dispatch-paragraphs.sh` keeps each one from being trimmed away by a later prose
edit. The no-delegation obligation has no paragraph and no guard row.

## What changes

- A `NO DELEGATION` blockquote — do the work yourself, never call the `Agent` tool, never spawn a
  subagent, background agent or helper; you are the leaf of this run and any child is unrecorded
  and outside the closed list — is added, verbatim and in the same "every dispatch prompt also
  carries" shape as its siblings, at four sites: the implementer dispatch (`skills/flow/implement.md`
  section 4), the panel slot dispatch and the panel-fix subagent dispatch
  (`skills/flow/review-panel.md`), and the verifier dispatch (`skills/flow/verify-and-handoff.md`,
  **The verifier dispatch**). The **Experimental slot** paragraph enumeration in `review-panel.md`
  names it too, so the experimental slot carries it like every other slot. **Dispatch sites — the
  conductor's closed list** gains one sentence naming the paragraph as what closes the list one
  level down.
- `scripts/check-dispatch-paragraphs.sh` gains a `NO DELEGATION` table entry — label plus three
  shared phrases, no variants — required once in `implement.md`, twice in `review-panel.md`, once in
  `verify-and-handoff.md`; `scripts/test-check-dispatch-paragraphs.sh` gains the block in every
  fixture asserted clean and one case per failure shape: label absent per site, one per phrase
  dropped, and the min-blocks threshold in `review-panel.md`.

No store, spec, contract-budget or planner-dispatch changes. Observable difference: a `/flow`
run's implementer, panel slot, panel-fix and verifier prompts each carry the paragraph, and
`scripts/check-dispatch-paragraphs.sh` exits 1 if any of them loses it.
