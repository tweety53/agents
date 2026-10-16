# self-review-kan-988-ledger-test-pool-starvation

## Why

The self-review pass over kan-988 found three non-`big` defects, each held against the tree
(`design.md`, **Context**):

- A gymie `/flow` reproducer named its live worktree by absolute path, so
  `prove-reproducer.sh`'s pre-fix leg read the already-fixed tree from inside its scratch and the
  proof meant nothing. `run-reproducer` refuses only absolute argv tokens, never the script body,
  and `scripts/prove-reproducer.sh`'s header claims a hardcoding script cannot reach past the
  scratch.
- Inline dispatch rows recorded an invented `effort=high` after `-effort unknown` was refused:
  `skills/flow-contracts/model-policy.md` says an unstated parent effort records as `default`, but
  the refusal in `flow record dispatch begin` does not say so.
- Harvested output tokens undercount: `ParseAssistantRecords` keeps a message's first line, but an
  early line can carry a streaming-time `output_tokens` the message's last line supersedes; the
  kan-988 ledger holds output 28 and 178 for reviewers that wrote whole reports.

## What changes

- `prove-reproducer.sh` exits 2, naming the path, when the reproducer's body names `<worktree>`'s
  absolute path; its header states that refusal instead of the false structural claim.
- The `-effort` refusal names the rule: an effort the session does not state is recorded as
  `default`, never guessed.
- The harvester counts each message's final output and thinking counts: the largest across its
  lines within one read, and the increase a later line carries when the message straddles two
  reads. Rows already stored are not rewritten.
