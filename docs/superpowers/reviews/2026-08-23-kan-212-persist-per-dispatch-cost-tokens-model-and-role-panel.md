# Review panel — kan-212-persist-per-dispatch-cost-tokens-model-and-role

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | per-task review (task 5) | major | stats/internal/harvest/watcher_test.go:2266 | TestRetryStillBounded asserts only that a give-up happened at or before cycle 60 and not again after; it never checks state before the bound, so an implementation seeding a retried token's cycle counter near the max passes unchanged, violating 'bounded exactly as the first attempt is'. |
| F2 | per-task review (task 6) | major | stats/internal/harvest/watcher.go:747 | Attribute computes deltas and ambiguous independently per record. Two windows sharing an AgentID make every id-carrying record ambiguous over both, while a no-agent-id record in the same batch attributes cleanly to one of them by interval. attributeDispatches then merges real tokens for that dispatch and also stamps it unattributed, contradicting task 6 step 4's 'Do not stamp a dispatch that attributed'. Repair: filter each AmbiguousDispatch.DispatchIDs against deltas before stamping, keeping candidates at the full ambiguity size. |
| F3 | per-task review (task 6.1) | major | stats/internal/harvest/watcher.go:1186 | The session-ambiguous branch stamps candidates=len(sessions) into the same metrics.unattributed.candidates key the dispatch-grain path uses for a count of concurrent dispatches. With no distinct wording, task 8 would either drop the count or render N matched sessions as 'indistinguishable from N concurrent dispatches'. Delta spec now enumerates four states and forbids the two ambiguities sharing a wording; task 8 renders all four. |
| F4 | per-task review (task 6.1) | major | stats/internal/harvest/watcher_test.go:2360 | The never-block subtest builds its Watcher with a nil logger, and w.warn is a no-op when logger is nil, so the test cannot observe whether the stamp failure was reported or silently dropped. The file already has the right pattern (bytes.Buffer + slog.NewTextHandler) in TestSessionTokenStopsBeingScannedAfterBoundedGiveUp. |
| F5 | per-task review (tasks 7-8) | minor | stats/internal/records/render.go:288 | tokenLine's default branch (verbatim render of an unrecognised reason through neutraliseMarkers) is required by task 8 step 3 and documented in the THE FOUR STATES comment, but render_test.go asserts nothing for it. A future edit could reinstate the not-measured fallback for that branch with no test failing. |
| F6 | per-task review (tasks 7-8) | minor | stats/internal/records/render_test.go:290 | dispatchMetrics.Unattributed's null and empty shapes have no dedicated case, unlike Tokens which has explicit table rows for exactly those. Behaviour is correct but asserted only incidentally. |
| F7 | per-task review (tasks 7-8) | minor | stats/internal/records/render.go:199 | The three reason-string constants are duplicated verbatim between internal/records/render.go and internal/harvest/watcher.go with nothing pinning them together. The tradeoff is named in a comment and the drift mode is visible rather than silent (a renamed reason falls to the verbatim branch), but no automated signal exists. Repair: a test in a package importing both asserting the three literals are equal. |
| F8 | per-task review (task 9) | major | stats/internal/api/records_test.go:590 | Nothing tests that an end call omitting -agent-id leaves an identifier begin recorded intact. TestDispatchEndAgentIDReachesTheStore covers only the opposite direction, and the CLI test proves only that the client omits the key. store.EndDispatch has no agent_id handling yet, so a naive unconditional SET agent_id would silently clear it and no test would catch it. Sent to task 10's implementer, which owns that SQL. |
| F9 | per-task review (tasks 9-10) | major | stats/internal/records/types.go:185 | records.CostStatusOf has zero test coverage anywhere, including its load-bearing tokens-outranks-unattributed precedence rule. The costStatus HTTP handler and its route are likewise unexercised: the CLI tests fake the daemon response and never touch the real handler/route/CostStatusOf chain. |
| F10 | per-task review (tasks 9-10) | major | stats/internal/store/records.go:275 | The COALESCE($8, d.agent_id) SQL, this commit's stated heart, is untested against real Postgres. TestDispatchEndOmittedAgentIDPreservesBeginsIdentifier lives in internal/api and exercises only fakeStore's Go conditional. Real behaviour verified correct by probe, but no shipped test asserts it. |
| F11 | per-task review (tasks 9-10) | minor | stats/internal/records/types.go:112 | An end call supplying a different agent id silently overwrites what begin recorded. Currently unreachable in the one real call site, and not forbidden by the spec, but the doc comments on DispatchEnd.AgentID and EndDispatch describe only the omitted case, so a future caller has no documented answer. |
| F12 | per-task review (tasks 9-10) | minor | stats/cmd/myflow/main.go:22 | main.go's top-level usage constant omits the new record cost-status verb, though record.go's own recordUsage was correctly extended. It is printed on bare invocation and on unknown-subcommand errors, so it is user-facing documentation. |
| F13 | per-task review (tasks 9-10) | minor | stats/internal/records/types.go:186 | CostStatusOf declares a local int named unattributed, shadowing the package-level type of the same name for the rest of the function body. |
| F14 | task 12 live verification | major | stats/internal/harvest/watcher.go:980 | The persisted give-up is retried, but the retry cannot bind: matchSessionTokens only scans commands NEWLY READ from a transcript this cycle, and harvest_offsets for the owning transcript already equals the file size (3147539 = full length). Seeding the token back into the pending set therefore searches a stream with no new bytes, exhausts the 60-cycle window and gives up again, permanently. The delta spec's scenario 'A given-up token is retried after a restart ... THEN the token is searched for again and binds' is unsatisfied. Binding is being gated on the usage-attribution offset, which is a different concern. |
| F15 | task 12 live verification | major | openspec/changes/kan-212-persist-per-dispatch-cost-tokens-model-and-role/design.md | The retry now binds (15/15 for mf-kan302-a3f9, 11/11 for mf-kan190-a3f7, measured live) but recovers no cost: the usage records behind the harvest offset are deliberately not re-read, so a bound-after-the-fact stage run gains identity, not retroactive token figures. design.md's persist-the-give-up decision and proposal.md both claim kan-302's uncosted half recovers, which overstates what was built. Either the claim is narrowed to binding, or re-attribution for a newly bound token is added. |
| F16 | Principles | minor | stats/internal/harvest/watcher_test.go:1301 | fakeSessionTokenStore and togglableSessionTokenBinder each carry byte-identical no-op implementations of the four widened SessionTokenBinder methods, ~30 duplicated lines. A shared embeddable no-op struct would remove the duplication. DRY vs a defensible WET call each fake's own comment explains. |
| F17 | Principles | minor | stats/cmd/myflow/record.go:280 | runRecordCostStatus's fallback.ProjectKey failure prints unknown and exits 0 without ever contacting the store, but the verb's doc comment tells the reader it contacts the store. Least Astonishment: the never-block path is wider than the comment states. |
| F18 | Primary | major | stats/internal/store/records.go:480 | jsonb_deep_add sums numeric leaves, which is correct for token deltas and wrong for unattributed.candidates, a snapshot fact. attributeDispatches restamps every harvest cycle a multi-minute ambiguity persists, and MarkDispatchesUnattributed restamps on every retried give-up, so candidates doubles per repeat and the ledger reports a count that grows with cycle count. attribute.go:665's doc comment asserts the stamp is idempotent under jsonb_deep_add; that claim is false. Repair: metrics \|\| $2::jsonb in both methods, correct the comment, add a store test that two identical stamps leave candidates unchanged. |

findings-total: 18
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 fixed
finding-status: F13 fixed
finding-status: F14 fixed
finding-status: F15 fixed
finding-status: F16 fixed
finding-status: F17 fixed
finding-status: F18 fixed

reproducers-total: 18
finding-reproducer: F1 cd stats && go test ./internal/harvest -run TestRetryStillBounded -count=1  # passes against a 1-cycle retry window
finding-reproducer: F2 cd stats && go test ./internal/harvest -run TestReproDispatchThatAttributedIsAlsoAmbiguous -count=1  # dispatch 10 appears in deltas AND in ambiguous from one Attribute call
finding-reproducer: F3 none — spec/plan gap in a not-yet-written task, not a runtime defect in this commit
finding-reproducer: F4 cd stats && go test ./internal/harvest -run 'TestGiveUpStampsItsOwnDispatches/stamp_failure_never_blocks' -v -count=1  # passes, but w.warn is a no-op under the nil logger the test constructs
finding-reproducer: F5 none — the defect is an absent test, not a wrong behaviour
finding-reproducer: F6 cd stats && go test ./internal/records -run TestRenderLedgerDistinguishesAnAbsentMeasurementFromAZeroOne -v  # passes, but no case carries an unattributed key at all
finding-reproducer: F7 none — a pinning gap, not a functional bug
finding-reproducer: F8 In a scratch worktree off 3bd6e2f, replace fakeStore.EndDispatch's guarded assignment with an unconditional d.dispatch.AgentID = in.AgentID, then: cd stats && go test ./internal/api -run Dispatch -v -count=1  # all pass with the clearing bug in place
finding-reproducer: F9 cd stats && remove the 'if m.Tokens != nil { continue }' guard from CostStatusOf, then go test ./internal/records ./internal/api ./cmd/myflow ./internal/store -count=1  # still all PASS
finding-reproducer: F10 cd stats && revert nullIfEmpty(in.AgentID) to in.AgentID at internal/store/records.go:280, then go test ./internal/store ./internal/api ./cmd/myflow -count=1  # still passes
finding-reproducer: F11 none — no test exercises this path; behaviour confirmed by ad hoc probe against the live store
finding-reproducer: F12 cd stats && go run ./cmd/myflow 2>&1 | grep -c cost-status  # 0
finding-reproducer: F13 none — cosmetic; confirmed harmless via go build ./internal/records
finding-reproducer: F14 cd stats && make restart; then after ~6 min: docker compose exec -T myflow-postgres psql -U myflow -d myflow -c "SELECT session_token, count(session_id) AS bound FROM stage_runs WHERE session_token='mf-kan302-a3f9' GROUP BY session_token"  # bound=0, and session_token_giveups holds the token again
finding-reproducer: F15 cd stats && docker compose exec -T myflow-postgres psql -U myflow -d myflow -c "SELECT stage, (metrics ? 'tokens') FROM stage_runs WHERE session_token='mf-kan302-a3f9'"  # all false, though session_id is now set on all 15
finding-reproducer: F16 none — a style preference, not a correctness issue; grep both fakes for 'no-op stand-ins' to see the duplicated blocks
finding-reproducer: F17 none — cosmetic; the doc comment's own scope is broader than one of its two early-return branches
finding-reproducer: F18 docker exec -i myflow-postgres psql -U myflow -d myflow -c "SELECT jsonb_deep_add('{\"unattributed\":{\"reason\":\"matched more than one dispatch\",\"candidates\":2}}'::jsonb, '{\"unattributed\":{\"reason\":\"matched more than one dispatch\",\"candidates\":2}}'::jsonb);"  # candidates reads back 4, not 2
