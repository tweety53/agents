## Context

- Angle 8 (`skills/flow-self-review/SKILL.md`, step 2) already defines a reviewer **miss** as "a defect raised only in a later round or by a later slot that a round-1 reviewer instruction would have caught", and only Critical/Important (or a genuinely useful Minor) findings qualify.
- KAN-1036 asks for the narrower, stronger signal: the *same* slot saw the *same* unchanged hunk in earlier rounds and said nothing. That is investigated on its own, regardless of severity, because the cause may be a prompt gap, a pipeline gap (a re-run shown only the delta) or neither.
- Prose-only change to one run-loaded file; no guard or case schema changes.

## Decisions

### Where the rule lives

**ID:** late-round-miss-in-angle-8
**Status:** active
**Chosen:** one paragraph in angle 8, directly after the `reviewer` bullet — the miss definition it sharpens sits there.
**Considered:** a new angle — rejected, it is a sub-case of angle 8's reviewer miss, and a ninth angle changes the report shape every guard reads; `skills/flow-prompt-tuning/SKILL.md` — rejected, that skill replays cases the pass already wrote and never reads a run's rounds.

### Severity bar

**ID:** late-round-miss-any-severity
**Status:** active
**Chosen:** traced whatever its severity — the ticket asks for the investigation, and the miss, not the finding, is the signal.
**Considered:** keep angle 8's Critical/Important bar — rejected, a Minor missed twice by the same slot is the same reviewer-variance signal.

### Recording the investigation

**ID:** late-round-miss-no-schema-field
**Status:** active
**Chosen:** the earlier rounds and the cause go in the case's existing `why_picked`; a non-instruction cause is an ordinary angle-8 finding classified by step 2's rules.
**Considered:** a new `case.json` field (`missed_rounds`) — rejected, it needs a `check-prompt-tuning-cases.sh` and renderer change for data `why_picked` already carries.

## Recommended approach

Add one paragraph to angle 8 defining a late-round miss and requiring its trace (decisions `late-round-miss-in-angle-8`, `late-round-miss-any-severity`, `late-round-miss-no-schema-field`), and acknowledge its sentences in `verbatim-moves.txt`.

## Critical files

- `skills/flow-self-review/SKILL.md` — angle 8; the paragraph is added after the `reviewer` bullet.
- `spectre/changes/kan-1036-late-round-finding-prompt-tuning/verbatim-moves.txt` — acknowledges the new sentences for `check-verbatim-moves.sh`.

## Existing code to reuse

- Angle 8's `reviewer` miss case and step 2a's `case.json` (`why_picked`, `source.round`, `source.slot`) — the case shape is reused unchanged.
- Step 2's classification rules (fixable / `big` / `noted`) — classify a non-instruction cause.

## Verification

- `scripts/check-verbatim-moves.sh` exits 0 with the new sentences acknowledged; `scripts/check-references.sh` and every other `scripts/check-*.sh` lint guard exit 0.

## Live check

None — prose in a skill file; no running service or persistent state changes.

## Open questions
