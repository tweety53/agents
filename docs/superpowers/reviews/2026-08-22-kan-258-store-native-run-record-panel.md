# Review panel - kan-258-store-native-run-record

**Roster:** `light` - Primary, Principles (Merged lens), Code review (low). All three dispatched on `sonnet`,
the model this change records under `models.reviewPanel`.

**Diff:** `.superpowers/sdd/final-review.diff`, 11743 lines over 17 commits, measured at 9762 by
`check-panel-diff-size.sh` against the 2000 cap. The guard exited 1; the operator was asked and first
chose to stop and split, then reversed and chose to proceed with the panel. Both answers are recorded
because the first one produced a handoff.

**Every commit had already passed its own combined spec-and-quality review as it landed.** These are the
cross-commit findings that per-task review cannot reach by construction.

## Findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Primary | Major | `openspec/changes/kan-258-store-native-run-record/specs/myflow-run-record/spec.md:224` | States the guard SHALL require no change; task 23 changed both panel guards' resolved path. Task 23 fixed the identical wording in the sibling delta spec and missed this near-duplicate. |
| F2 | Primary | Major | `openspec/specs/agents-repo-verification/spec.md:12` | A live capability spec normatively requires project.md's ## test to name scripts/test-preserve-session-records.sh, which task 13 deleted. This change carries no delta spec for that capability, so after merge the repository violates its own spec. |
| F3 | Primary | Minor | `openspec/specs/myflow-self-review/spec.md:65` | A live spec still cites the deleted script as canonical for gather-self-review-context.sh's locate logic and outcome vocabulary. |
| F4 | Primary | Minor | `stats/cmd/myflow/record.go:701` | resolveRenderKinds' comment claims no pipeline step asks for -kind all; section 7's prUrl push path does. |
| F5 | Principles | Minor | `skills/myflow-do/SKILL.md:1288` | Says the Records line has two alternatives; it has three, as the same file says correctly at :1252 and as handoff-blocks.md states. The miscount was copied into myflow-fast/SKILL.md:278. |
| F6 | Code review (low) | Major | `stats/cmd/myflow/record.go:295` | The dispatch row is inserted only at close, but the harvester commits its offset every 5s and never re-reads past it, so a long dispatch's own tokens are dropped or credited to a sibling. |
| F7 | Code review (low) | Major | `stats/internal/store/records.go:138` | No production path ever sets Dispatch.EndedAt, so every dispatch window stays open forever and later usage is credited to a stale sibling by the latest-start fallback. |
| F8 | Code review (low) | Major | `stats/internal/store/records.go:84` | RecordDispatch has no idempotency key, so a journalled write replayed after a lost response inserts a second row for one logical dispatch - double-counted cost, unlike UpsertFinding. |
| F9 | Code review (low) | Major | `skills/myflow-contracts/finish-contract.md:163` | The inline artifact cp kept destination containment and the name allowlist but dropped the retired script's source-path containment, so a symlinked source is followed into the committed, pushed repository. |
| F10 | Code review (low) | Minor | `stats/internal/api/stats_test.go:1340` | newIntegrationTestServer passes a fake RecordStore although a real migrated store satisfying the interface is already in hand, so no records route is exercised against Postgres. |
| F11 | Code review (low) | Minor | `stats/internal/harvest/attribute.go:641` | DispatchAttributor.Attribute rebuilds a second per-session index over the same batch Attributor.Attribute already grouped, adding a redundant pass each cycle. |
| F12 | Code review (low) | Minor | `stats/internal/store/records.go:104` | insertDispatch and readDispatches duplicate the same fifteen-column scan and decode block, so one path can silently drift from the other on a future column change. |
| F13 | Code review (low) | Minor | `stats/internal/store/records.go:376` | DispatchWindow.SessionID is populated but read nowhere in production code. |
| F14 | Code review (low) | Minor | `stats/internal/client/client.go:640` | writeRecord and readRecordResponse leave BeginStage and EndStage hand-rolling the same prologue, and the two differ on 409 classification. |
| F15 | Code review (low) | Minor | `stats/internal/harvest/attribute.go:559` | DispatchWindow.contains duplicates Window.contains byte for byte, including the open-window semantics F6 and F7 hinge on, so fixing one grain would leave the other broken. |

findings-total: 15
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
finding-status: F11 withdrawn the redundant index-building code is gone (shared groupBySession); the second pass is kept deliberately, because sharing the grouped structure would couple both attributors' signatures and call order to save one O(n) walk, against their documented pure-computation-over-a-slice contract. Reason recorded in groupBySession's own doc comment.
finding-status: F12 fixed
finding-status: F13 fixed
finding-status: F14 fixed
finding-status: F15 fixed

reproducers-total: 15
finding-reproducer: F1 grep -rn require openspec/changes/kan-258-store-native-run-record/specs/myflow-run-record/spec.md
finding-reproducer: F2 grep -rn test-preserve-session-records openspec/specs/agents-repo-verification/spec.md
finding-reproducer: F3 grep -rn preserve-session-records openspec/specs/myflow-self-review/spec.md
finding-reproducer: F4 grep -rn resolveRenderKinds stats/cmd/myflow/record.go
finding-reproducer: F5 grep -rn alternatives skills/myflow-do/SKILL.md skills/myflow-fast/SKILL.md
finding-reproducer: F6 none — needs a live myflowd and a harvest tick elapsing inside an unclosed dispatch; established by code trace across record.go, watcher.go and attribute.go, and now covered by TestUsageDuringAnUnfinishedDispatchIsAttributedToIt
finding-reproducer: F7 grep -rn EndedAt stats/internal/store/records.go
finding-reproducer: F8 none — needs a lost daemon response followed by a reconcile pass; established by contrast with the UpsertFinding path in the same file, and now covered by the dispatch-key idempotency test
finding-reproducer: F9 grep -rn resolve skills/myflow-contracts/finish-contract.md
finding-reproducer: F10 grep -rn newFakeStore stats/internal/api/stats_test.go
finding-reproducer: F11 grep -rn groupBySession stats/internal/harvest/attribute.go
finding-reproducer: F12 grep -rn scanDispatchRow stats/internal/store/records.go
finding-reproducer: F13 grep -rn SessionID stats/internal/harvest/attribute.go
finding-reproducer: F14 grep -rn sendJSON stats/internal/client/client.go
finding-reproducer: F15 grep -rn intervalContains stats/internal/harvest/attribute.go

