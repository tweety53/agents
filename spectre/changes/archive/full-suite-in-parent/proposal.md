# full-suite-in-parent

## Why

- The plan-last implementer runs the full `## test` list; `guard-tests` alone takes 2–6 min.
  <!-- measured: flow suite list @ 2026-09-24 — guard-tests 2m04s–5m58s -->
- Subagents write their prompt cache with a 5-minute TTL (the main session: 1 hour), so a suite
  run that outlasts 5 minutes expires the implementer's cache and its next request re-writes the
  whole context — observed at 205K and 351K tokens in one implementer's transcript.
  <!-- measured: usage.cache_creation_input_tokens in an implementer transcript under ~/.claude/projects/-Users-tweety53-Projects-agents @ 2026-09-24 — 204911 after a 765s gap, 351188 after a 307s gap -->

## What changes

- The parent runs the full suite once after the last group, in every wave shape; no implementer
  carries FULL SUITE.
- A failure in the last group's own files is fixed by the parent inline; any other failure stops
  with `## Question`, as today.
- No subagent is ever resumed. Every dispatch is one-shot: a guard refusal, a wave pick conflict
  and a gated reviewer's `fix` verdict are fixed by the parent inline, where today each re-wakes
  the finished implementer after its cache has expired.
