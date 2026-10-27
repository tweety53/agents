# pipeline-fix-kan-1040-copyable-next-command-38

## Why

`/flow` run 2's Finished handoff was printed without its mandatory `**Prompt tuning:**` line
(`skills/flow/cleanup.md`: "`Learning:` and `Prompt tuning:` are never omitted"), and the operator
reports it as recurring. In kan-1040 the parent received `NEXT: print the Finished handoff` from
`integrate-change.sh land` (`stats/internal/guard/integratechange_cleanup.go`), never opened
`skills/flow/cleanup.md`, and composed the block from memory. Nothing at that point names the
template or its never-omitted lines.

## What changes

- `land`'s final line becomes `NEXT: print the Finished handoff (skills/flow/cleanup.md) — its
  Learning: and Prompt tuning: lines included`, naming the template and the two lines a block
  composed from memory drops.
- `stats/internal/guard/integrate_change_test.go`'s happy path gains a check that fails without it.
- `hooks/flow-stop-guard.py` blocks a turn whose last assistant text is a Finished handoff (or a
  `/flow-fast` landing report carrying one of the two lines) lacking `**Learning:**` or
  `**Prompt tuning:**`, naming the missing line(s) and `skills/flow/cleanup.md`; pinned in
  `stats/internal/repotest/flow_stop_guard_hook_test.go`.
- `/flow-fast`'s skill text is unchanged: its landing report is written from `skills/flow-fast/SKILL.md` itself,
  which already names both lines; no `NEXT:` line hands it off.
