## Context

- Five findings of the kan-1000 self-review pass, planned per item 1 of **3. Fix every finding
  that is not `big`** (`skills/flow-self-review/SKILL.md`), on branch
  `self-review-kan-1000-fe-group-sessions-through-the-same-doors` off agents `origin/main`
  (`19fd4a8c`), worktree
  `/Users/tweety53/Projects/agents-worktrees/self-review-kan-1000-fe-group-sessions-through-the-same-doors`.
  Each finding is carried by exactly one task; the brainstorm design is
  `.superpowers/sdd/2026-10-21-self-review-kan-1000-fe-group-sessions-through-the-same-doors-design.md`
  in this worktree. No path the findings name changed on `origin/main` since the run, and
  `spectre/changes/archive/` records no rejection of any remedy below.
- The kan-1000 rows live in the dev flow store (`flow-postgres`, `localhost:5433`, project
  `gymie-7c1f238a`), read with `flow record dispatches -C /Users/tweety53/Projects/gymie -change
  kan-1000-fe-group-sessions-through-the-same-doors`. Row numbers below are the rows' `seq`; the
  repair commands take the row's `id`. The parent transcript is
  `/Users/tweety53/.claude/projects/-Users-tweety53-Projects-gymie/45e57ddc-ab4c-4f0e-bffc-e105ed098d03.jsonl`,
  its agents under that path without `.jsonl`, in `subagents/`.
- `flow record cost-status -change kan-1000-fe-group-sessions-through-the-same-doors` (from
  `/Users/tweety53/Projects/gymie`) prints `20 unattributed — not measured: 20`.
  <!-- measured: flow record cost-status -change kan-1000-fe-group-sessions-through-the-same-doors @ live store, 2026-10-21 (not re-runnable once repaired) -->
- **1a (Task 1), blast radius 8.** `stampDispatchAgents` (`stats/internal/harvest/watcher.go`)
  pairs a launch with a begin by time, same-instant begins by their text order. Two drifts beat it
  in kan-1000:
  - Round 9's begins went out in one Bash call before the launch message, the verifier's begin
    last (`panel-9-primary+principles`, `panel-9-test-sanity+call-efficiency`,
    `visual-verify-fixrun1-final-gymie`), while the launches went out verifier first, so each
    launch took the next begin's row: rows 62, 63 and 64 are rotated by one, and rows 69, 70 and
    71 the same way; rows 47 and 51 carry their own round's primary-slot agent.
  - Rounds 3 to 8 begun their slots with `-key panel-$R-<slot>` after `R=$(flow record next-round
    …)`. That half of the pass's cause is already refused in the tree: `dispatchBeginFromInvocation`
    drops a key holding `$` (`bb1f82c6`), so such a begin claims no launch — but the launch then
    has no begin at all and its row keeps no agent: rows 25, 26, 30, 33, 34, 40, 41, 43, 45, 46,
    49, 50, 57 and 58.
  - Every launch result carries the Agent call's `description` (`toolUseResult.description`),
    and the description rule (**Every dispatch's Agent-tool `description`**,
    `skills/flow/implement.md`) already opens every non-task row's description with its key: the
    first word of each of the 22 launches is its row's `dispatch_key` exactly (`panel-3-primary`,
    `visual-verify-fix-4-full-gymie verifier`, …). Every kan-1000 row carries `session_id`, the
    transcript's `sessionId`.
    <!-- measured: docker exec flow-postgres psql -U flow -d flow -Atc "select count(*), count(session_id) from dispatches where session_token in ('mf-k1000b3','mf-k1000f1')" → 76|76 @ live store, 2026-10-21 -->
  - The true agent of each row is the agent whose `subagents/agent-<id>.meta.json` description's
    first word is the row's key; Task 1's repair table lists all 22. Rows 52 and 53 are backfill
    duplicates of row 51's verifier and stay unstamped.
- **1b (Task 2), blast radius 6.** Rows 24 `task-11-implementer-fix-1`, 42
  `verify-fix-3-gymie-frontend`, 60 `verify-fixrun1-gymie-frontend` and 73 `task-19-implementer`
  are `inline`, recorded by `flow record dispatch backfill`, and carry no `tokens`: the main-thread
  records inside each window were committed by earlier batches, before the row existed, and the
  watcher never re-attributes committed usage. `flowd restamp-dispatch` (`b2555b98`) repairs only
  an agent-file row.
- **1c (Task 3), blast radius 4.** `.superpowers/sdd/panel-report-11-primary.md` in the kan-1000
  gymie worktree raises one Important (`gymie-playwright:scripts/specs-for-diff.mjs:56-80`); the
  store holds F1–F9 and nothing for round 11; the parent fixed it inline as gymie-playwright
  `b2212bb`. `check-panel-findings-closed.sh` on that worktree prints `FINDINGS-CLOSED`, exit 0.
  <!-- measured: scripts/check-panel-findings-closed.sh /Users/tweety53/Projects/gymie-worktrees/kan-1000-fe-group-sessions-through-the-same-doors kan-1000-fe-group-sessions-through-the-same-doors → FINDINGS-CLOSED rc=0 @ agents 19fd4a8c -->
  Report files have no fixed shape: round 0's primary report heads its one Important `**P1 — …**`
  followed by six sub-bullets, round 11's writes it as one bullet under `#### Important` — a
  bullet count would read 6 against 1 recorded for round 0 and misfire.
- **1d (Task 4), blast radius 2.** Step 8 of `skills/flow/visual-verify-verifier.md` zips the
  resolver's list with `zip -X -j -@`; `-X` drops extra attributes but the DOS timestamp of each
  entry is the file's mtime. Spiked: two zips of the same bytes with mtimes pinned by `touch -t
  198001010000` compare identical, and differ once one file's mtime moves. `sync-onto-base.md`
  step 3 rebuilds "exactly as step 8", so it follows with no edit.
- **4a (Task 5), blast radius 2.** `runRecordDispatchBegin` (`stats/cmd/flow/record.go`) requires
  no `-diff-base` for any row; kan-1000 rows 30, 43, 47 and 51 (`visual-verify-fix-*`) carry none,
  which `visual-verify.md`'s **Recording** paragraph requires and only
  `check-visual-verify-dispatched.sh` notices, at integrate. `dispatch backfill` runs the same
  begin, so it inherits the refusal.
- Prose edits were applied at plan time, `scripts/check-verbatim-moves.sh` run, and its FAIL
  sentences listed in this change's `verbatim-moves.txt`; the edits were reverted. With them
  applied `check-dispatch-paragraphs.sh`, `check-markdown-integrity.py` and `go test
  ./internal/guard/ -run 'RenderSlot|DispatchParagraph|PanelFindingsClosed'` passed.

## Decisions

### Stamp a launch by its description's first word in the launching session

**ID:** launch-stamp-by-description-key
**Status:** active
**Chosen:** `ParseAgentLaunches` keeps the first whitespace-delimited word of the launch result's
`description` as `AgentLaunch.Key`; before time pairing, the watcher calls the new
`StampSessionDispatchAgent(sessionID, key, agentID)` — the latest row with that `dispatch_key`,
that `session_id` and an empty `agent_id` — for every carried and new launch with a key; a stamped
launch leaves the pairing, and every begin naming its key, now or in a later batch, is dropped.
A launch whose key names no row pairs by time exactly as before.
**Considered:** scoping a key lookup to session tokens parsed from the transcript's begins — a
`$R` begin is refused before its token is read, and a second token parser is a copy of one the
watcher has; a `Key:` line in every panel and verifier prompt — a new prompt contract for an
identity every launch already carries; tightening the time ordering — kan-1000's two drifts are
both orderings the protocol permits.

### Measure a back-filled inline row from the parent transcript and replace its figures

**ID:** inline-backfill-recompute
**Status:** active
**Chosen:** when a committed batch carries a `flow record dispatch backfill … -agent-id inline`
command with a literal `-key` and `-session-token`, the watcher looks the row up
(`InlineDispatchID`), runs the dispatch attributor over the parent transcript's records up to the
committed offset, and replaces the row's `tokens`/`signals` with that row's share
(`RestampInlineDispatch`, which touches only an `inline` row). `flowd restamp-dispatch <id>
<parent-transcript>` runs the same computation over a whole parent transcript, the repair path.
**Considered:** computing the figures in the CLI at backfill time — the CLI reads no transcript
and writes through flowd's HTTP surface; merging the share additively — a watcher restart re-reads
and would double it; reading the file past the committed offset — the next batch would credit
the same records again.

### Count a slot report's findings by its own count lines

**ID:** panel-report-count-lines
**Status:** active
**Chosen:** the REPORT FILE paragraph has every slot open its report with one `Critical: <n>` and
one `Important: <n>` line; `check-panel-findings-closed.sh` reads each
`<worktree>/.superpowers/sdd/panel-report-<round>-<id>.md` carrying both lines and exits 1 when
their sum exceeds the store's Critical and Important rows of that round whose slot's `+`-components
include `<id>`, any status. A report carrying neither line is not compared.
**Considered:** counting bullets under `#### Critical` / `#### Important` — report shapes vary,
and round 0's one Important reads as six; refusing a report without the lines — every in-flight
run's reports predate the paragraph and would stop at their next panel close.

### Record kan-1000's round-11 Important as fixed

**ID:** f10-recorded-fixed
**Status:** active
**Chosen:** F10, round 11, slot `primary`, Important, `-dispatch-seq 74`, task 19, recorded
`open` then set `fixed` (gymie-playwright `b2212bb`). `check-panel-findings-closed.sh` then
reports F10 as fixed with no clean re-run, which is true: the parent fixed it inline.
**Considered:** recording it `withdrawn` — it was fixed, and the reason would be false; leaving
it unrecorded — the defect this task exists to close.

### Pin each PNG's mtime before zipping the full-app suite

**ID:** zip-pin-mtime
**Status:** active
**Chosen:** step 8 captures the resolver's list once, `touch -t 198001010000` over it, then zips
it with `zip -X -j -@`.
**Considered:** comparing the new zip's entries with the old before committing — a second
procedure where pinning the input makes the output stable; `zip -D` — drops directory entries
only, the mtime stays.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** flow-medium — Tasks 1 and 2 change which rows flowd's harvester credits and add two
store writes every harvested transcript can reach.
**Considered:** flow-low — fits Tasks 3–5, but a mis-stamp in Task 1 or a double count in Task 2
corrupts every later run's cost figures silently.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** two fixer dispatches in series: `fix-g1` carries Tasks 1 and 2 (the harvester and
store seam they share: `watcher.go`, `deps.go`, `records.go`, `watcher_test.go`, `records_test.go`),
then `fix-g2` carries Tasks 3, 4 and 5 (guard, verifier prose, CLI refusal).
**Considered:** one fixer for all five — Tasks 1 and 2 alone fill a fixer's context; one per task
— five dispatches where Tasks 4 and 5 are a few lines each.

## Open questions

## Live check

Against the live dev `flowd` after the landing's `make restart` (`stats/`), and the repaired
kan-1000 rows:

- `flow record cost-status -change kan-1000-fe-group-sessions-through-the-same-doors` from
  `/Users/tweety53/Projects/gymie`: before `20 unattributed — not measured: 20`, after
  `2 unattributed — not measured: 2` (rows 52 and 53).
  <!-- predicted: flow record cost-status after Tasks 1 and 2's repair steps -->
- `flow record dispatches` for that change: rows 45, 47, 62, 64 and 71 carry agents
  `a7b53a4363e1874c1`, `a6ff0d4128631e27b`, `a4dea4a68353bbbd1`, `a769b929d38dce23e` and
  `ae7f31e03b6f094bf`; row 47's sidechain output is no longer `329`; rows 24, 42, 60 and 73 carry
  `tokens.main` with a non-zero `cache_read`.
  <!-- predicted: flow record dispatches after Tasks 1 and 2's repair steps -->
- The next panel round any run begins after the restart, its begins in any order: every slot row
  and verifier row carries the agent whose `description` names its key. Not working looks like
  today: a verifier row with a 46-line reviewer's figures, or a panel row with no agent.
- The next `flow record dispatch backfill -agent-id inline` any run records: its row reads
  `tokens.main` within one watcher cycle. Not working: `cost-status` counts it `not measured`.
- `flow record findings` for kan-1000 lists F10 `fixed`, and `check-panel-findings-closed.sh` on
  the kan-1000 gymie worktree names F10 as fixed with no clean re-run (exit 1) — the true state.
