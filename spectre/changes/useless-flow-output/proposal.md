## Why

Sessions narrate routine mechanics between status lines — "Main has moved, so I'll commit and then
rebase", "keeping upstream's version and dropping that edit", "Re-running lint and tests before
pushing", "origin/main already fails this check; fixing it per the new rule". The operator called
each of these completely useless (2026-10-02). `rules/be-brief.mdc` already limits a status line to
unit and state, but said nothing about prose between status lines.

## What changes

- `rules/be-brief.mdc`: between status lines, say nothing; routine mechanics and the reason a step
  was taken are never narrated; what the user needs reaches them once, in the final report.
