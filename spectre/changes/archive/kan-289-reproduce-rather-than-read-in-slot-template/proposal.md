# kan-289 — "reproduce, don't read" in the dispatch template, and `myflow` → `flow`

**Jira:** KAN-289

This change carries two concerns, at the operator's explicit request. They land as separate commits
with separate scopes; they share only this change's branch.

## Why

### The dispatch template (KAN-289 proper)

KAN-289 records a self-review finding from `kan-258-store-native-run-record`. Every review dispatch
in that change carried an instruction to verify by running rather than by reading, typed into the
dispatcher's prose each time. It sat in no template, so it held only as long as the dispatcher
remembered it.

Four defects in that change survived a reading review and died to a reproduction: a doc comment
asserting a snapshot the CTE did not share; `jsonb_deep_add`'s real merge semantics; twelve marker
encodings run against the real `check-unfinished-work.sh`; and a hand-built `Dispatch{}` literal
whose `len == 0` is a shape the store never produces, which would have rendered "0 tokens" instead
of "not measured" for every Cursor and Codex ledger entry, indefinitely.

The fourth is the general case: a test backed by a fake or a hand-built value passes while the real
integration is broken. It happened four times in one change.

### The rename

The pipeline's single command is `/flow`. Its skills are `flow`, `flow-status`, `flow-research`,
`flow-settings`. Everything underneath still says `myflow` — the CLI binary, the `myflowd` daemon,
the `myflow-postgres` container, the `myflow` database, the `MYFLOWD_*` and `MYFLOW_*` environment
variables, `.myflow/`, the `myflow-contracts` skill directory, and the contract vocabulary
throughout. The product name and the implementation name disagree at every layer below the command
surface.

## What changes

### The dispatch template

One labelled blockquote, `**REPRODUCE, DON'T READ:**`, in two variants — one for a reviewer judging
someone else's claim, one for an implementer writing the test that will be judged — carried verbatim
into three dispatch-prompt families: every review-panel slot, the per-task combined reviewer, and the
implementer dispatch.

A new guard, `scripts/check-reproduce-not-read.sh`, asserts the label and each variant's
load-bearing phrases are present at each required site, so a later prose edit cannot reproduce the
same failure one level up.

### The rename

`myflow` becomes `flow` throughout the live corpus, identifiers included: the CLI binary, the
daemon, the container, the database, the environment variables, `.myflow/`, the contract skill
directory and its citations, the Go and TypeScript identifiers, and the contract vocabulary.

`scripts/check-vocabulary.sh` gains `myflow` and its identifier shapes as retired literals, which is
what keeps the rename from silently regrowing.

**Out of scope, deliberately:** the frozen `openspec/` tree; the historical records under
`docs/superpowers/` and `spectre/changes/archive/`; the retired command-name constants in
`stats/internal/stages/names.go` and the `mf-` session-token prefix, both of which are values already
written into the live store; and relabelling existing Jira issues. Each is recorded as a decision in
`design.md`.

**Four steps are handed to the operator rather than run**: renaming the container, renaming the
database, reinstalling the launchd agent, and installing the renamed binary. All four require the
dev workspace's daemon and storage to stop, which no agent action does.
