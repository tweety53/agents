# Proposal — kan-853-reconcile-contradicting-rule-copies

**Jira:** KAN-853 (epic KAN-851). **Plan:** `docs/prompt-audit-2026-09-29/README.md`.

## Why

The 2026-09-29 prompt audit found co-loaded copies of the same rule that disagree, plus stale
pointers and positional words. The session trims (KAN-855..859) delete duplicate copies; deleting
either side of a disagreement would be a behaviour change nobody decided. This change settles every
listed disagreement first.

## What changes

Each in-scope `## Drift / bugs found` row is corrected to its canonical copy:

- `pipeline.md` for the state machine;
- `implement.md` **The handshake** for the model handshake;
- `finish-contract-run1.md` / `finish-contract-run2.md` for landing and archive;
- the Go guard over its header table.

Scope (IDs from each audit file):

- `audit-every-session.md` D1–D10, D12–D23, D25, D26;
- `audit-planning.md` #1–#4, #9–#14, #16;
- `audit-implement.md` #1–#5, #7–#10, #17, #18, #20, #21;
- `audit-review-panel.md` B2–B6, B8–B17;
- `audit-verify-and-contracts.md` D1, D3–D9, D11–D14, D16–D18;
- `audit-subagent-prompts.md` D3, D4, D7, D9, D10;
- `audit-finish.md` D1–D7, D9, D13, D15–D21, D24.

Every sentence reworded on purpose is listed in `verbatim-moves.txt`.

## Out of scope

- Rows that need a behaviour decision (KAN-854).
- Any trim (KAN-855..859).
