# Review panel — kan-357-record-each-dispatch-own-token-usage

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | primary | minor | spectre/changes/kan-357-record-each-dispatch-own-token-usage/tasks.md:157 | Commit fab1223 creates undeclared stats/internal/harvest/agentfile_test.go; the four declared TestAgentFileRecords tests live there instead of attribute_test.go (forced — attribute_test.go is package harvest_test and cannot reach the unexported function). |
| F2 | primary | minor | spectre/changes/kan-357-record-each-dispatch-own-token-usage/tasks.md:191 | Plan's verified:authored snippets assert Sidechain.InputTokens; the real Bucket fields are Input/Output (attribute.go:140-142), so the snippets as written do not compile. |
| F3 | primary | minor | spectre/changes/kan-357-record-each-dispatch-own-token-usage/tasks.md:231 | Routing test lacks the declared top-level-path-keeps-inference control (attribute_test.go:1699-1736); that route stays covered only by pre-existing tests. |
| F4 | simple-reviewer | minor | stats/internal/harvest/watcher.go:852 | A subagents/ file not named agent-<id>.jsonl is routed away from DispatchAttributor by directory name, then silently dropped when AgentIDFromTranscriptPath's prefix check fails — pre-change it attributed via inference. Unreachable today (no harness writes such a file). |

findings-total: 4
finding-status: F1 deferred — /flow-fast defers every Minor; fix rounds cover Critical and Major only
finding-status: F2 deferred — /flow-fast defers every Minor; fix rounds cover Critical and Major only
finding-status: F3 deferred — /flow-fast defers every Minor; fix rounds cover Critical and Major only
finding-status: F4 deferred — /flow-fast defers every Minor; fix rounds cover Critical and Major only

reproducers-total: 4
finding-reproducer: F1 none — plan declares attribute_test.go; the unexported function forced an in-package test file
finding-reproducer: F2 none — snippets are plan text; the real assertions use Bucket.Input/Output
finding-reproducer: F3 none — the declared control case is absent from the test file
finding-reproducer: F4 write subagents/helper.jsonl with a resolvable agentId — pre-change it attributed, on this branch merged=map[]
