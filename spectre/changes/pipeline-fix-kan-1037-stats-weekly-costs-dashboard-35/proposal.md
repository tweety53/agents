# pipeline-fix-kan-1037-stats-weekly-costs-dashboard-35

## Why

The operator never saw the prompt-tuning line at the end of the kan-1037 run: the Finished handoff
in `skills/flow/cleanup.md` defines `**Prompt-tuning:** <paths> | NONE — <reason>`, but the run
ended on a hand-written summary instead of the block. The operator asked, verbatim: "Why do I still
not see the prompt-tuning string in the end? Update flow to write it in the following format:
Prompt tuning: html path or Prompt tuning: NONE" and then "It always must be present".

## What changes

- The Finished handoff line is renamed `**Prompt tuning:**`; its value is each case's `index.html`
  absolute path, or the literal `NONE` with no reason suffix. `skills/flow-fast/SKILL.md` cites the
  new name.
- `skills/flow/cleanup.md` states that the `Prompt tuning:` line is present in every `/flow` and
  `/flow-fast` run's final message: the Finished handoff is printed as that message, and a final
  report written beside or instead of it still carries the line.
- `scripts/test-prompt-tuning-line.sh` fails while either file carries the old shape.
- No other reader exists: a grep of skills, rules, scripts, hooks and stats finds `Prompt-tuning:`
  only in those two files.
