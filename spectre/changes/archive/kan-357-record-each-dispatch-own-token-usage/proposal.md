# kan-357-record-each-dispatch-own-token-usage

## Why

Dispatch-grain cost attribution discards a resumed subagent's whole spend. The kan-377 run's
store rows show the shape: three agentIds each appear on two dispatch rows (an implementer
resumed per task), and every row carrying one of them is stamped
`unattributed — matched more than one dispatch` — 6 of 28 rows in that one run; KAN-357 records
the same failure as 9 of 27 on the run it was filed from. A third of a run's cost is therefore
unmeasurable at exactly the grain the cost figures are read at.

Root cause: one resumed agent = one agentId across several dispatch rows. The identity pass
matches more than one row, timestamp narrowing against hand-typed windows fails, and the
refuse-to-guess rule (kan-212) correctly discards the usage rather than crediting a guess.

<!-- measured: flow record dispatches -change kan-377-flow-cut-run-cost-and-time-conductor-subagent @ live myflow store, 2026-09-10 -->

## What changes

The harvester stops inferring a dispatch's usage from time windows wherever the source itself
names its dispatch: records harvested from a per-agent transcript
(`subagents/agent-<id>.jsonl`) are credited directly to the dispatch row(s) carrying that file's
agentId. A single row takes the batch whole; a resumed agent's several rows split the file's
usage among themselves by row start order. Time-window inference remains only for sidechain
records arriving inside top-level session transcripts. Stage-grain attribution is unchanged.
