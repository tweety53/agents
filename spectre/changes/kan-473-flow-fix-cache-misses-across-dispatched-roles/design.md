# kan-473-flow-fix-cache-misses-across-dispatched-roles — design

## Context

Three prose edits in `skills/flow/`, one guard row with harness cases, one analysis script, one
budget row. No spec: `spectre/specs/` holds no capability this change alters.

## What the transcripts show

Source: `~/.claude/projects/-Users-tweety53-Projects-agents/a7273417-c46a-4128-9f73-619e4a99c84d.jsonl`
and the 12 files under that session's `subagents/`. Every role's `message.model` is
<!-- measured: ls <session>/subagents/*.jsonl | wc -l @ machine-local transcript, cannot be re-run from a ref -->
`glm-5.3-flash`, and `cache_creation_input_tokens` is 0 on every turn — the TTL and pricing
measured are that provider's, with implicit caching written asynchronously after a turn.

Every number in the table below was read from those transcripts' `message.usage` and
`timestamp` fields by the analysis `scripts/cache-misses.py` (task 1) makes repeatable.
<!-- measured: the ad-hoc predecessor of scripts/cache-misses.py over the session above @ machine-local transcript, cannot be re-run from a ref -->

Every zero-cache-read turn above 5k input tokens, classified:

| Class | Rows | Cause | In scope |
|---|---|---|---|
| Cold start | implementer 12:02:39, primary 12:27:18, planner 11:29:24, main 11:22:43 (turn 1); conductor 11:55:24, panel-fix 12:49:32 and 12:49:46, delta-primary 13:08:22, delta-principles 13:08:33, main 12:18:59 (turns 2–3, 0.2–14 s after a miss) | a fresh agent's system prompt plus tool schemas (~17k), or a turn landing before the provider's asynchronous cache write from the prior miss | no |
| Schema load | conductor 11:58:32 (67,917), principles 12:28:29 (23,490) | a `ToolSearch "."` query loaded seven deferred tools; the tools array precedes every message, so the prefix changed. Every zero-match `ToolSearch` (nine in the same two transcripts) left the cache intact | yes |
| Idle child-wait | conductor 12:19:28 (966 s), 12:43:01 (14m17s), 13:03:32 (13m20s); main 12:18:52 (921 s), 13:03:07 (782 s) | the conductor returned `## Stage … In progress — awaiting …` at 12:03:21, 12:28:30 and 12:49:54 with a child running; the parent answered "Conductor holding; waiting" and ended its own turn; both resumed only when the child finished | yes |
| Provider flake | planner 11:41:58 (139 s after a `SendMessage` resume) | ten of the thirteen coordinator resumes across the run hit the cache, including two at ~2.5 minutes | no |
<!-- measured: every row above — the ad-hoc predecessor of scripts/cache-misses.py over the KAN-445 session @ machine-local transcript, cannot be re-run from a ref -->

## The edits

### `skills/flow/implement.md`, **Turn discipline**

The wait loop `for i in $(seq 1 110); do test -s <report> && break; sleep 5; done` becomes
`seq 1 48` — 240 s — and the paragraph states why in one sentence: a wait longer than the
prompt-cache TTL re-prices the whole context on return, while a bounded wait's `still-running`
turn reads the context at the cache rate and keeps it warm. The paragraph already says
`still-running` re-issues the wait; the ceiling tracking is unchanged. `review-panel.md` and
`verify-and-handoff.md` cite this paragraph and change nothing.

### `skills/flow/implement.md`, **Dispatch the conductor**

- The relay contract's "never with a child subagent still in flight" sentence gains the cost: a
  turn that ends with a child running idles this role and the parent until the child finishes,
  and both re-price their whole context on resume.
- A parent backstop, one paragraph after the three blocks: a conductor return that names a child
  in flight — `awaiting`, `in flight`, a dispatch key with no `## Stage` end mark behind it — is
  not one of the three blocks; the parent resumes it with `continue` in its very next action and
  never waits on a grandchild itself.

### The TOOLS paragraph

Verbatim at six dispatch sites, in the same "also carries" shape the sites already use:

> **TOOLS:** Every tool you need that is not already listed in your tool set — `SendMessage`,
> `Monitor`, an MCP tool — is loaded in one `select:<name>,<name>` ToolSearch in your first turn,
> before anything else. Never ToolSearch for a tool already listed, and never a wildcard query: a
> schema loaded later changes your tool list and re-prices your whole context at full input rate.

Sites and guard minimums: `skills/flow/implement.md` 2 (conductor dispatch, implementer
dispatch), `skills/flow/review-panel.md` 2 (panel slot, panel-fix subagent),
`skills/flow/brainstorm.md` 1 (planner), `skills/flow/verify-and-handoff.md` 1 (verifier).
Shared phrases for `scripts/check-dispatch-paragraphs.sh`'s table, no variants: "in your first
turn", "never a wildcard query", "re-prices your whole context". The harness
`scripts/test-check-dispatch-paragraphs.sh` gains a `TOOLS_BLOCK` on every fixture asserted
clean, sites `brainstorm.md` and `verify-and-handoff.md` in its sandbox, and one case per
required phrase dropped plus one label-absent case.

### `scripts/cache-misses.py`

Python 3, standard library only, like the repository's other `.py` scripts. Usage:
`cache-misses.py <transcript.jsonl>…` — a `subagents/` directory expands to its `*.jsonl` files,
labelled by the sibling `.meta.json`'s `description`. For each transcript, one line per assistant
message whose `usage.cache_read_input_tokens` is 0 and `usage.input_tokens` > 5000, deduplicated
on (timestamp, input, cache_read) because a streamed message writes several rows carrying the
same usage: `turn#<n> <HH:MM:SS> gap=<s> in=<tokens> prev_user=<text|tool_result>
tools_since_last_usage=[…]`. Exit 0 always; a file that is not JSONL is reported on stderr and
skipped. Its check is the KAN-445 session: run against it, the output reproduces the table above.

### `scripts/check-contract-budget.sh`

`skills/flow/implement.md` sits at 31,948 of 32,500 bytes; the edits above add roughly 900. Its
row is raised to the edited file's size plus 25%, as the guard's header defines a row.
`review-panel.md` (50,049 of 58,732), `brainstorm.md` (11,908 of 35,015) and
`verify-and-handoff.md` (31,337 of 33,199) absorb their blocks within budget.

## Decisions

### Fix the two measured mechanisms, not the ticket's two

**ID:** measured-mechanisms-only
**Status:** active
**Chosen:** bound the wait loop under the TTL, add the parent `continue` backstop, and add the
TOOLS paragraph — the three causes the transcripts actually show.
**Considered:** keeping a role's cache warm across operator gates (the ticket's mechanism-1
direction) — rejected: no idle gap in the run was an operator gate; every one was a child wait,
and a turn that has ended awaiting a relayed answer cannot fire anything. Keeping large tool
results out of the cached region (the ticket's mechanism-2 direction) — rejected: no tool-result
size correlates with a miss; the two real invalidations follow schema loads.

### The wait loop is the keep-alive

**ID:** wait-loop-is-keepalive
**Status:** active
**Chosen:** shorten the existing bounded wait to 240 s; each return is one cheap cached turn.
**Considered:** a separate periodic no-op turn — rejected: it is the same turn the wait already
produces, added as a second mechanism. Leaving 550 s and accepting one miss per long wait —
rejected: at the cache-read rate four keep-alive returns cost well under one full re-price of a
77–120k context.

### TOOLS as a guarded dispatch paragraph

**ID:** tools-paragraph-guarded
**Status:** active
**Chosen:** one verbatim blockquote at every dispatch site, registered in
`check-dispatch-paragraphs.sh` like FOREGROUND BUILDS.
**Considered:** a single sentence in `implement.md` cited from the other sites — rejected: the
guard scans sites for the label and phrases, and a cited sentence is what a later trim removes
silently. Naming the loaded tool set explicitly — rejected: the set differs per session
(`SendMessage` is deferred in some), so the paragraph names the rule and examples, not a list.

### Reproducer is a report, not a guard

**ID:** cache-misses-report-not-guard
**Status:** active
**Chosen:** `scripts/cache-misses.py` prints the analysis and exits 0; its check is one run on
the KAN-445 session.
**Considered:** adding it to `.flow/project.md`'s `## lint` with a fixture — rejected: it has no
pass/fail answer about the repository, and a synthetic fixture is the hand-built value the
REPRODUCE, DON'T READ paragraph warns against.

## Open questions

None.
