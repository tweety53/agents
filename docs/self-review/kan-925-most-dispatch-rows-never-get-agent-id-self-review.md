# Self-review report for kan-925-most-dispatch-rows-never-get-agent-id

**Deferred:** reasoning pass run on Muse Code (Muse Spark) from the in-run bundle

In-run pass: the change fixes KAN-925 (dispatch agent_id capture) with two
commits on the branch (`a56f576d` loop expansion + template guard,
`dc8928fc` brace-depth stack and multi-line loops), a mutation-proved
watcher-level regression test, a compact primary+principles panel (round 0
with one retry, round 1 re-run — all four findings fixed, verdicts cited),
full `## lint` green and the scoped harvest suite green. The pass covers
the in-run bundle — ledger, panel record, findings — and nothing beyond it.
No product defect met; the panel's product findings were fixed and
re-reviewed before this pass.

## Session narrative

Implemented the stamper fix inline: traced kan-916's empty agent_ids from
the store through the KAN-322 transcript pairing, replayed kan-916's real
5.5MB parent transcript through the Watcher with a recording fake (13 of 20
launched rows stamped pre-fix — the 7 misses exactly the `for`-loop and
`eval` begins), wrote the failing test first, then expanded literal-word
loops, split `&&`/`;` top-level only, and skipped unresolvable template
keys. Post-fix replay stamps all 19 launches onto their own rows in order;
each fix condition was mutation-killed and restored. Approaches tried and
dropped: a skill-side ban on loop-recorded begins (the daemon now handles
the shape the skill's one-call batching encourages, and a skill edit would
churn run-loaded prose), and store-side row-time pairing fallback
(inference, against the pairing discipline).

## Problems encountered, and what pipeline change would avoid them — `flow-fix`

- **[flow-fix]** panel round-0 first attempt returned its handshake only and did no review; one manual retry succeeded — a durable empty-return retry rule is left uncodified — declined

## Token/time cost, and what would reduce it without quality loss — `flow-cost`

- **[flow-cost]** duplicate spend from the empty first panel attempt: a full-context dispatch for zero output plus one retry roundtrip — declined

## What went well, and how to reproduce it — `flow-improvement`

- **[flow-improvement]** replaying kan-916's real parent transcript through the Watcher with a recording fake isolated the root cause pre-code (13/20, then 19/19 after the fix); mutation proof per fix condition; panel caught two real bugs pre-land — declined

## What could be automated or moved to a script — `flow-automation`

- **[flow-automation]** a reusable transcript-replay harness would automate the hand-built replay probe; out of scope for this run — declined

## What could move to the Go app or its persistent storage — `flow-stats-app`

_none — this angle produced no findings._

## What can be sped up — `flow-speed`

- **[flow-speed]** the `## lint` list ran serially in batches; parallelizing the invocation is run technique, not a pipeline change — declined
