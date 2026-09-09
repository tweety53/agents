# kan-479-harvest-zcode-rollout-transcripts

## Why

- `zcode` stage runs record no tokens, cost or model — structural, not harvesting lag: the
  pipeline treats ZCode like Cursor/Codex, a harness that writes no harvestable transcript.
  Measured (store, 2026-09-09): 267 zcode stage runs, 0 with tokens, 0 with a bound session,
  0 with models; claude-code: 4328 runs, 2738 harvested. Every recent zcode run carries
  `tokens_available: false`.
- ZCode already writes everything needed at `~/.zcode/cli/rollout/model-io-sess_<uuid>.jsonl`:
  one JSON line (`type: "model_io"`) per API call with `sessionId`, `startedAt`/`completedAt`,
  `model.modelId`, and `response.usage.{inputTokens, outputTokens, cacheReadTokens,
  cacheWriteTokens}`. Stage-mark commands appear in the line's tool-call `input.command` text.
- Cost stays blank even after harvesting until GLM rates are seeded, and ZCode reports a single
  `cacheWriteTokens` with no 5m/1h split, which the pricer refuses today.
- Harness mislabeling recurrence: the kan-252 archive run of 2026-09-09 (stage runs 4617–4628)
  is stamped `-harness claude-code` but was driven from ZCode — `pipeline.md`'s harness roster
  names no zcode, so the agent guessed.

## What changes

- `internal/harvest`: a second transcript source — root `~/.zcode/cli/rollout` (env override
  `FLOW_ZCODE_ROLLOUTS_DIR`), files `model-io-sess_*.jsonl` — plus a parser for the `model_io`
  line shape. One line is one API response; no per-content-block dedup. The Watcher takes a
  list of sources instead of one root; flowd wires both, one Watcher.
- Session-token resolution reads the token from rollout tool-call command text
  (`request.messages[*].toolCalls[*].input.command` and `response.toolCalls[*].input.command`);
  attribution and binding are unchanged.
- `internal/store`: cache-creation usage with an unknown 5m/1h split becomes priceable when the
  model's rate carries a single cache-write rate (exact, never a guess); `glm-5.3-flash` seeded
  at the discounted plan rates (source: docs.z.ai/guides/overview/pricing, read 2026-09-09;
  keyed to the canonical lowercase form per fix 1 below).
- `internal/api/stages.go`: the `claudeCodeHarness` single-constant check becomes the set of
  harnesses whose transcripts are harvested (`claude-code`, `zcode`).
- `skills/flow-contracts/pipeline.md`: `zcode` joins the `-harness` roster.
- Out of scope, stated: per-dispatch figures stay Claude-only (rollout lines carry no agentId
  and no sidechain marker).

## Fix 1 — 2026-09-09: canonical model bucket keys

- Found before landing: the rollout `modelId` is **not** stable — live files carry both
  `"modelId":"GLM-5.3-Flash"` and `"modelId":"glm-5.3-flash"` within one session (the design's
  premise that only `request.body.model` varies is falsified by measurement). Verbatim bucket
  keys split one model's usage across two buckets; the lowercase bucket has no pricing row, so
  `Price` reports `ErrPricingNotFound` and writes **no** top-level `cost_usd` for every
  mixed-spelling session — effectively all of them.
- Fix: model bucket keys are canonicalized to lowercase where they are minted
  (`internal/harvest` attribution), and the seed row is keyed `glm-5.3-flash`. See design
  decision `model-id-canonical-lowercase`.
