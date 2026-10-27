# Design — pipeline-fix-kan-1037-stats-weekly-costs-dashboard-35

## Context

The kan-1037 run ended on a hand-written summary, so the Finished handoff's prompt-tuning line
never reached the operator (`proposal.md`).

## Decisions

- **prompt-tuning-line-renamed** — `**Prompt-tuning:**` becomes `**Prompt tuning:**`, the
  operator's spelling, in the Finished block, its explanatory paragraph and `/flow-fast`'s citation.
- **none-without-reason** — the `NONE` value carries no reason suffix. The previous reason ("angle
  8's own empty-angle sentence") is dropped because the operator's format is `Prompt tuning: NONE`.
  `Learning:` keeps its `NONE — <reason>`; the operator asked only about the prompt-tuning line.
- **always-in-final-message** — the always-present rule is stated once, in `skills/flow/cleanup.md`
  beside the block that is canonical for the line; `/flow-fast` already cites that block and says
  "never omitted". Rejected: restating the rule in `skills/flow-fast/SKILL.md`, which would
  duplicate a canonical statement.
- **test-script-guard** — the regression check is a standalone `scripts/test-prompt-tuning-line.sh`
  grepping the two files. Rejected: a Go guard in flow-guard, which is for guards run as lint lines;
  this is a one-off regression test over two literal strings.
- **verbatim-moves-acknowledged** — the reworded and new sentences are listed in this change's
  `verbatim-moves.txt`, as `scripts/check-verbatim-moves.sh`'s header requires.
