## Context

The self-review fix loop of kan-1036, planned on the change's own branch
`spectre/kan-1036-late-round-finding-prompt-tuning` in worktree
`/Users/tweety53/Projects/agents-worktrees/kan-1036-late-round-finding-prompt-tuning`, fix base
`c580312c`. Each finding was checked against that commit, the dev store (read-only `SELECT`s) and
the parent session's transcript before its fix was designed. The brainstorm trail is
`.superpowers/sdd/2026-10-26-self-review-kan-1036-late-round-finding-prompt-tuning-design.md` in
this worktree.

**Rounds count from 0.** `Pass.Round` mirrors `Finding.Round`: 0 for the initial panel, 1..n for
fix rounds (`stats/internal/records/types.go`). A miss first raised in round 1 is therefore a
late-round miss, and the earliest reviewer instruction is round 0's, the first round.

**Off-shape reproducer names are common and collide.** Of the store's findings whose reproducer
sits under `.superpowers/sdd/reproducers/`, many in-flight rows carry names off
`<round>-<id>-<n>.sh`: `r0-P1-1.sh` recorded by both `primary` and `failure-modes` in one change,
`r1-M1-1.sh` for a round-0 `mutation` finding (kan-1036 F3), `panel0-mutation-F1-1.sh` and
`r1-ce-I1-1.sh` (kan-1017). A finding joined across slots records the `+`-joined slot set
(`primary+principles`) while its script carries one of those slots' ids (`0-primary-1.sh`) — the
shape the guard must accept.
<!-- measured: docker exec flow-postgres psql … select … from findings f join changes c … where f.reproducer like '.superpowers/sdd/reproducers/%' and c.state <> 'FINISHED' and f.reproducer !~ ('^\.superpowers/sdd/reproducers/' || f.round || '-' || <escaped slot> || '-[1-9][0-9]*\.sh( |$)') @ dev store, 2026-10-26 -->
kan-1036's F3 points at `.superpowers/sdd/reproducers/r1-M1-1.sh`, whose content is not F3's
mutation check (deleting the late-round paragraph survives) but a prompt-text check — the other
slot's script, written over F3's on the case-insensitive filesystem. F3 is Minor and withdrawn.

**The verifier row's gap is a missing zero, not lost spend.** The harvest attributes a
main-thread transcript record to the inline window containing it, by timestamp alone
(`DispatchAttributor.Attribute`, `stats/internal/harvest/attribute.go`). A window holding no such
record gets no delta, so its metrics bag stays `{}` and the ledger renders `not measured`.
`run-verify-list.sh` records its row around one blocking Bash call: the parent's assistant record
that issued the call lands before `begin`, the next one after `end`. kan-1036's dispatch 6
(`verify`, inline) spans `13:31:27.823Z`–`13:37:25.519Z`; the parent transcript's assistant
records fall at `13:31:25.743Z` and next at `13:37:28.709Z`, none inside. Its sibling inline rows
1 and 5, which span several parent turns, are measured.
<!-- measured: select seq, dispatch_key, agent_id, started_at, ended_at, metrics ? 'tokens' from dispatches … where change = kan-1036-late-round-finding-prompt-tuning; jq over ~/.claude/projects/-Users-tweety53-Projects-agents/2419c446-935a-4755-ae1b-dc145717161c.jsonl for records between 13:31:00Z and 13:38:00Z @ dev store and transcript, 2026-10-26 -->
`MergeDispatchMetrics` adds a patch with `jsonb_deep_add`, and `MetricsPatch.Tokens` always
marshals, so a zero patch writes a `tokens` object of zeros and adds nothing to one already there
(`stats/internal/store/records.go`, `stats/internal/harvest/attribute.go`).

## Decisions

### Count round 1 as a late round

**ID:** late-round-from-1
**Status:** active
**Chosen:** "first raises in round 1 or later" and "a first-round reviewer instruction" in `skills/flow-self-review/SKILL.md`; no other run-loaded file carries either phrase.
**Considered:** "round 0" for the reviewer instruction — exact, but "first-round" reads without knowing the numbering.

### Name the reproducer shape in the slot prompt and refuse an off-shape path

**ID:** reproducer-name-shape-guarded
**Status:** active
**Chosen:** the slot-facing REPRODUCER paragraph states `<round>-<id>-<n>.sh` with what each part is; `check-panel-reproducers` adds a violation (exit 1) for a finding whose reproducer's path token starts `.superpowers/sdd/reproducers/` and whose remainder is not `<round>-<id>-<n>.sh`, `<round>` the finding's own round, `<id>` the finding's slot or any one of its `+`-separated parts, `<n>` a positive integer without a leading zero. A reproducer elsewhere (`scripts/…`) and the `none — <reason>` exemption are untouched. One sentence in `skills/flow/review-panel.md` beside "there is no rename step" names the refusal as a malformed field, the class `review-panel-fix-round.md` already has the parent repair before dispatch.
**Considered:** a write-time refusal in `flow record finding` — the parent records the slot's path verbatim, so the refusal would land on the parent mid-recording with no class to route it; the guard already runs before every fix round. Matching the slot case-insensitively — the collision is between two names that differ only in case, so the exact slot id is what keeps them apart.

### Repair kan-1036's F3 row only

**ID:** repair-f3-only
**Status:** active
**Chosen:** re-record kan-1036's F3 with every field unchanged but `-reproducer "none — its mutation script was overwritten by another slot's r1-m1-1.sh on a case-insensitive filesystem (kan-1036)"`, legal on a Minor finding; the overwritten file is left in the gitignored reproducers directory. Other in-flight changes' off-shape rows are not rewritten: the guard reads them only before a fix round of their own change, where the malformed-field class already has the parent repair them, and their scripts live in worktrees of other runs (kan-1017 is another project's live change).
**Considered:** renaming `r1-M1-1.sh` to `0-mutation-1.sh` — the file's content is not F3's check, so the rename would record a false reproducer.

### Round column after ID

**ID:** round-column
**Status:** active
**Chosen:** `| ID | Round | Slot | Severity | Location | Note | Lineage |`, the cell `strconv.Itoa(f.Round)`; the KAN-857 example table in `skills/flow/SKILL-rationale.md` shows the current header with the new column.
**Considered:** a Round column at the end — the round is a finding's identity, read before its slot.

### Measure an empty inline window as zero

**ID:** inline-window-zero
**Status:** active
**Chosen:** `DispatchWindow` gains `Measured bool`, read by `DispatchWindowsForSession` as `(d.metrics -> 'tokens') IS NOT NULL`. In `DispatchAttributor.Attribute`, after a session's records are attributed, every inline window with an `EndedAt`, `Measured` false and no entry in `deltas`, whose `EndedAt` is at or before the latest main-thread record of that session in the batch, gets an empty `DispatchDelta` — written as a `tokens` object of zeros. The ledger then reads `input 0, output 0, cache read 0, cache creation 0`: the parent spent no tokens inside the window, a measurement.
**Considered:** render-only wording for a closed inline `{}` row — the renderer cannot tell "the harvest found no turn inside" from "the harvest has not read this far yet" or a harness with no transcript, so the wording would be false for those. Widening the window to the issuing turn — charges the parent's surrounding turns to the row, a guess. A batch-local test (window end inside the batch's span) — misses exactly this case, where the batch boundary falls in the window's gap.

### Past rows stay as they are

**ID:** inline-zero-no-repair
**Status:** active
**Chosen:** no repair of past `{}` inline rows. `not measured` is a true statement about a row nothing ever measured, not a wrong figure; the harvest offsets of finished sessions are past those windows, and kan-1036's own row 6 is measured once the operator's `flowd` runs this build and its session writes one more record.
**Considered:** a restamp command writing zeros to past rows — it would assert a measurement nothing re-read.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 4 changes what the always-running harvest writes on every batch, and Task 2 adds a refusal every panel fix round runs.
**Considered:** `flow-low` — Tasks 1 and 3 are mechanical, but a wrong condition in Task 4 writes zeros over rows or misses them silently.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — key `self-review-kan-1036-late-round-finding-prompt-tuning-fix`, Tasks 1 to 4 in order; their files are disjoint.
**Considered:** one fixer per task — four dispatches for small tasks with no ordering between them.

## Open questions

## Recommended approach

Two prose edits applied as verified diffs (Tasks 1 and 2), a lexical name check added to the
existing reproducer guard with one table test and a store-row repair (Task 2), one table column in
the panel renderer (Task 3), and a zero delta from the dispatch attributor for an inline window the
harvest has read past without finding a turn (Task 4). Rejected: a write-time reproducer refusal
(`reproducer-name-shape-guarded`), render-only token wording (`inline-window-zero`), past-row
repair (`inline-zero-no-repair`).

## Critical files

- `skills/flow-self-review/SKILL.md` — angle 8's late-round-miss definition (Task 1).
- `skills/flow/review-panel.md` — the reproducer rule and the slot-facing REPRODUCER paragraph (Task 2).
- `stats/internal/guard/panelreproducers.go`, `stats/internal/guard/check_panel_reproducers_test.go` — the guard and its table test (Task 2).
- `stats/internal/records/render.go`, `stats/internal/records/render_test.go` — `RenderPanel`'s findings table (Task 3).
- `skills/flow/SKILL-rationale.md` — the KAN-857 example of the rendered table (Task 3).
- `stats/internal/harvest/attribute.go`, `stats/internal/harvest/attribute_test.go` — `DispatchWindow`, `DispatchAttributor.Attribute` (Task 4).
- `stats/internal/store/records.go`, `stats/internal/store/records_test.go` — `DispatchWindowsForSession` (Task 4).

## Existing code to reuse

- `cprCommandViolations` and `cprSeverities` (`stats/internal/guard/panelreproducers.go`) — the violation shape and the index-aligned side read the new round/slot read copies.
- `cprFindings`, `cprRun`, `cprExpect` (`stats/internal/guard/check_panel_reproducers_test.go`) — the test fixtures.
- `tableCell`, `lineageCell` (`stats/internal/records/render.go`) — cell rendering.
- `InlineAgentID`, `groupBySession`, `bestDispatchWindow` (`stats/internal/harvest/attribute.go`) and `TestInlineDispatchTakesMainThreadUsage` (`stats/internal/harvest/attribute_test.go`) — the inline path and its test fixture.
- `TestDispatchWindowsForSessionResolvesTheBoundToken` (`stats/internal/store/records_test.go`) — the store read's existing test.

## Verification

- `cd stats && make -s web-dist && go vet ./... && gofmt -l .` clean; `go test ./internal/guard/ ./internal/records/ ./internal/harvest/ ./internal/store/`.
- `scripts/check-verbatim-moves.sh`, `scripts/check-references.sh` exit 0 after Tasks 1 and 2.
- `scripts/check-panel-reproducers.sh /Users/tweety53/Projects/agents-worktrees/kan-1036-late-round-finding-prompt-tuning kan-1036-late-round-finding-prompt-tuning` exits 0 after Task 2's repair.
- `go test ./internal/records/ -run TestRenderPanelCarriesEachFindingsRound` shows the Round column rendered from a finding's stored round.

## Live check

Two runtime surfaces, each run at `flow.verify` on this worktree:

- **The guard against the real store (Task 2).** `scripts/check-panel-reproducers.sh <this worktree> kan-1036-late-round-finding-prompt-tuning`: after Task 2 it prints `REPRODUCERS-OK (3 finding(s) declared)` at exit 0, and `flow record findings -change kan-1036-late-round-finding-prompt-tuning` shows F3's reproducer as the exemption. Stand-in for "before": the same command against the worktree checked out at `c580312c` prints `REPRODUCERS-OK` too — the old guard never looked at the name. Not working looks like exit 1 naming F3's `r1-M1-1.sh` (repair not made), or a `REPRODUCERS-OK` from this build on a scratch findings file carrying F3's old row (guard not refusing).
- **The attributor against the real transcript (Task 4).** The protected `flowd` is never restarted, so the stand-in is a scratch test file in `stats/internal/harvest/`, deleted after the run and never committed, that reads `~/.claude/projects/-Users-tweety53-Projects-agents/2419c446-935a-4755-ae1b-dc145717161c.jsonl` with the harvester's own transcript reader, keeps the main-thread records between `2026-10-26T13:31:00Z` and `2026-10-26T13:38:00Z`, and runs `DispatchAttributor.Attribute` over them with a fake window source returning kan-1036's dispatch 5 (`13:30:22.713Z`–`13:31:00.301Z`, measured) and dispatch 6 (`13:31:27.823Z`–`13:37:25.519Z`, unmeasured), both inline. Record before (this build's parent, `c580312c`) and after: before, `deltas` has no entry for dispatch 6; after, it has one whose tokens are all zero, and dispatch 5 gets none (already measured, no record inside). Not working looks like no entry for dispatch 6, or non-zero figures on it.
