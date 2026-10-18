# self-review-backlog-flow-3

## Why

Five tickets in KAN's "Backlog Flow" status, filed by earlier self-review passes, are planned here
as one self-review fix change. One is already fixed (`design.md`, **Context**); the rest name gaps
the tree still has:

- **KAN-609** — on harness `zcode` the contract and `hooks/enforce-agent-baseline.py` name
  `subagent_type` `flow-high`, which ZCode's Agent tool never offers, so kan-553's panel ran on
  `general-purpose` with nothing on the record saying so.
- **KAN-629** — kan-361's handoff left an out-of-scope `admin-frontend` unstarted behind a port
  another session held, and nothing in the handoff said the stack was partial.
- **KAN-631** — kan-361's `friendship_ordered_pair` check constraint caught a nondeterministic
  test insert before it landed; no principle asks for such constraints.
- **KAN-633** — kan-361's `ui paths`-flagged DTO/converter diff drew nothing; its verifier checked
  the wire contract live on the API by its own judgment, with no contract asking for it.

## What changes

- The zcode harness mapping (`skills/flow-contracts/model-policy.md`) and the hook's `ZCODE_PAIR`
  name `general-purpose`; **The roster** (`skills/flow/review-panel.md`) records any spawn type the
  harness does not offer on the round's pass log.
- **A refused start is relayed, never resolved** (`skills/flow/verify-and-handoff.md`) gains the
  out-of-worktree exception and its `Not started:` handoff line.
- The Robustness principle (`skills/flow/engineering-principles.md`) gains the schema-enforceable
  invariant held in application code alone.
- `flow.visual-verify` names a `wire contract:` for a flagged diff that draws nothing, and its
  verifier checks it live on the API (`skills/flow/visual-verify.md`,
  `skills/flow/visual-verify-verifier.md`).
