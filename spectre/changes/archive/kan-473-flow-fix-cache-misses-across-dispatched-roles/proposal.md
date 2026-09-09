# kan-473-flow-fix-cache-misses-across-dispatched-roles

## Why

Every dispatched role in a `/flow` run pays full input price on some turns instead of the
cache-read rate (KAN-473, consolidating KAN-470 and KAN-471). Re-analysing the KAN-445 run the
ticket cites (session `a7273417-c46a-4128-9f73-619e4a99c84d`, main thread plus its 12 subagent
transcripts) shows the ticket's two mechanisms are not what the transcripts record:

- The "idle-gap TTL lapse" turns are not the operator taking minutes to answer. Each one follows
  the conductor returning `## Stage … In progress — awaiting …` with an implementer, panel or fix
  child still running — a turn `skills/flow/implement.md` already forbids — after which the
  parent also ended its turn ("Conductor holding; waiting") instead of resuming the conductor
  with `continue`. Both roles then sat idle 13–16 minutes and re-priced 77–120k (conductor) and
  83–107k (main) tokens on resume. Even a conductor that obeys the rule lapses: the canonical
  <!-- measured: message.usage and timestamp fields of the KAN-445 transcripts, the analysis scripts/cache-misses.py makes repeatable @ machine-local transcript, cannot be re-run from a ref -->
  wait loop is `seq 1 110` × 5 s, 9m10s of silence per call, longer than any prompt-cache TTL.
- The "sub-TTL prefix invalidation" turns split two ways. Two are real invalidations, and both
  follow a `ToolSearch` that loaded deferred tool schemas (a `"."` query loading seven tools):
  the tools array precedes every message, so a loaded schema re-prices the whole context. The
  agents were hunting for `Read`, `Bash` and `Agent`, which they already had. Every other row is
  a fresh agent's turn 1 (system prompt plus tool schemas, ~17k tokens) or a turn 2–3 arriving
  0.2–14 s later, before the provider's asynchronous cache write exists — cold start, not
  invalidation, and no tool-result size correlates.

## What changes

- `skills/flow/implement.md` **Turn discipline**: the wait loop is bounded under the prompt-cache
  TTL (`seq 1 48`, 240 s) so every `still-running` return is a cheap cache-read turn that keeps
  the role's prefix warm — the keep-alive, using the loop that already exists.
- `skills/flow/implement.md` **Dispatch the conductor**: the never-end-with-a-child-in-flight
  sentence carries its cost, and the parent gains a backstop — a conductor return naming a child
  in flight is resumed with `continue` at once; the parent never waits on a grandchild.
- A **TOOLS** dispatch paragraph — load every deferred tool you need in one `select:` ToolSearch
  in your first turn, never search for a tool already listed, a schema loaded later re-prices
  your whole context — verbatim at the conductor, implementer, panel-slot, panel-fix and planner
  dispatch sites, with a row in `scripts/check-dispatch-paragraphs.sh`'s table and cases in its
  harness so it cannot be trimmed away.
- `scripts/cache-misses.py`, the analysis above made repeatable: for a session transcript or a
  `subagents/` directory, every zero-cache-read turn above 5k input tokens with its timestamp,
  turn number, gap since the prior turn, the preceding user-message kind and the tools used
  since the last usage row.
- `scripts/check-contract-budget.sh`'s row for `skills/flow/implement.md` is raised to the
  file's new size plus 25%, the guard's own rule for a row.
- Out of scope, recorded in `design.md`: cold start per spawn, and the planner's one 2m19s
  resume miss (ten of thirteen sibling resumes, some at 2.5 minutes, hit the cache).
  <!-- measured: every user entry with origin.kind=coordinator in the KAN-445 transcripts and the next usage row's cache_read_input_tokens @ machine-local transcript, cannot be re-run from a ref -->
