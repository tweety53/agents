## Why

Sessions narrate routine mechanics between status lines — "Main has moved, so I'll commit and then
rebase", "keeping upstream's version and dropping that edit", "Re-running lint and tests before
pushing", "origin/main already fails this check; fixing it per the new rule". The operator called
each of these completely useless (2026-10-02). `rules/be-brief.mdc` already limits a status line to
unit and state, but said nothing about prose between status lines.

## What changes

- `rules/be-brief.mdc`: between status lines, say nothing; routine mechanics and the reason a step
  was taken are never narrated; what the user needs reaches them once, in the final report.
- `rules/be-brief.mdc`, from an audit of two full session transcripts (2026-10-02):
  - no acknowledgement lines ("Noted", "Agreed");
  - a status line's state is one of the five words and ends the line; an unchanged state is not
    reprinted;
  - the final report carries only the delta: no re-telling of a subagent's report, no per-file
    change list unless asked, all-green checks as one line, a pending question or running unit
    stated once, no memory-saving announcements, no internal bookkeeping the user cannot act on.
- `rules/agent-baseline.md` **Reporting back**: a subagent's report has four parts — Verdict,
  Findings, Decisions, Not done — plus any field the dispatch prompt asks for. Hand-backs ran 30–90
  lines of model lines, coverage lists and passing-check lists the parent re-told anyway.
