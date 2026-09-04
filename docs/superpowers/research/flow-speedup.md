# flow speed-up — research notes

Source: none

Explored via `/flow-research` on 2026-09-04 against this repository's own `spectre/` tree, the
phase files under `skills/flow/`, the guards under `scripts/`, and the dev stats store at
`http://127.0.0.1:4173` (read-only queries against `/api/v1/stage-runs`, `/api/v1/stats/stage-leaderboard`
and `/api/v1/records/agents-a740d89c/<change>`). No spectre change exists yet. The topic the operator
brought: "How can I speed up `flow` implementation+review? What stages can be merged in all the flow
(beginning to end)? What can be added to flow?" The session first chose **wall clock per change** as
the target to go deep on (sections 3–5), with cost and operator attention investigated to one round
each; the operator then reopened the session and took **cost** through three rounds (section 6) and
**operator attention** through two (section 7). A third session (section 8) closed every item the
first two left open, reading the per-slot sidechain transcripts under
`~/.claude/projects/-Users-tweety53-Projects-agents/<session>/subagents/agent-<agentId>.jsonl` (the
`agentId` each dispatch row carries) and `pmset -g log` beside the store.

## 1. The measured end-to-end map

All figures are `/flow` runs in project `agents-a740d89c`, roughly 49 runs per stage, wall clock
taken as `endedAt - startedAt` of each stage run in the store. "Share" is that stage's total across
all runs divided by the sum over every stage. **Medians are the figures to use.** A stage run's
wall clock includes every minute the machine slept or the operator was away, and nothing in the
store distinguishes those from work (section 8); kan-402 — a run on a sleeping laptop — is
excluded from every wall-clock figure below, which moves the panel mean from 57 to 51 minutes and
sdd-tdd's from 49 to 47 and leaves every median where it was. The means that remain are still
mostly waiting.

| Stage | median | mean | share of total wall clock | dispatches / gates |
|---|---|---|---|---|
| `flow.review-panel` | 23 min | 51 min | 25% | one subagent per resolved roster slot (store roster today: `primary`, `principles`, `code-review-low`, `mutation`), one `panel-fix` per round, delta re-runs; no operator gate on an ordinary run |
| `flow.sdd-tdd` | 17 min | 47 min | 23% | one implementer plus one per-task reviewer per bundle, serial per worktree, inside the conductor |
| `flow.brainstorm` | 12 min | 29 min | 14% | one planner (fable); at least two question rounds, a convergence confirm, the design approval — mostly human time |
| `flow.verify` | 5 min | 26 min | 12% | one verifier (sonnet) running the full `## lint` and `## test` lists, strictly after the panel closes |
| `flow.push-archive` | 23 s | 15 min | 7% | the mean is PR-route waiting; the median is the push |
| `flow.self-review` | 53 s | 14 min | 6% | one subagent, a skip prompt, five per-angle filing prompts, a rating prompt |
| `flow.landing-routes` | 84 s | 9 min | 4% | push or merge |
| `flow.writing-plans` | 2.5 min | 3 min | 1% | the same planner dispatch, resumed |
| the other 22 stages | 0–60 s each | — | about 3% combined | thirteen of them have a median under 15 s: their whole runtime is two `flow stage` marks plus a few parent tool calls |

Cost per change (mean): `flow.sdd-tdd` about $23, `flow.review-panel` about $18, every other stage
together about $2–4. The stage leaderboard puts the panel's p90 at $51 and sdd-tdd's at $85.

Per-dispatch rows from the last seven archived changes (kan-380, kan-389, kan-390, kan-394, kan-395,
kan-402, kan-404):

- A typical change is 2–4 tasks, 12–23 dispatches, one panel round, one or two `panel-fix`
  dispatches.
- Panel `primary` looked like the long pole inside the panel: kan-395 `primary` 23 min against
  `principles` 16 s, `code-review-low` 50 s, `mutation` 3 min; kan-402 `primary` 93 min, `bugbot`
  76 min. Section 8 shows both to be stalls outside the model — a sleeping laptop and a harness
  retry — and `primary` at 94–159 s in every awake change, no slower than bugbot.
- Per-task reviewer yield: 1 `fix` outcome in 19 per-task reviews (kan-402 task 1). Its wall-clock
  cost is small because it overlaps the next implementer; its cost is one sonnet dispatch per task.
- Panel findings by slot across the seven: bugbot 11, primary 3, code-review-low 3, mutation 2,
  principles 1. Seventeen of twenty were Minor. A Minor finding triggers no slot re-run, but every
  round still pays a `panel-fix` dispatch (1–16 min), the reproducer re-run and the mutation-proof
  walk.
- Human prompts on a creating run, minimum: n brainstorm questions + the convergence confirm + the
  design approval (asked back to back) + the landing question (skipped in this repository by the
  configured `## default landing route`) + the self-review skip prompt + five per-angle filing
  prompts + the rating — about 9 + n.

Structural observations the data confirmed:

- `flow.verify` is loaded only "once `skills/flow/review-panel.md` closes clean", although the
  verifier and the pass-1 panel slots read the same immutable HEAD whenever the panel raises no
  finding.
- `stage-diff → run-instructions → write-in-progress`, `preserve-sessions → commit-two`,
  `verify-merge → sync-archive → commit-archive` and `cleanup → verify-cleanup → write-finished` are
  mark-only chains: no decision sits between the marks, and each stage is seconds long.

## 2. The three targets, expanded before choosing

### Target A — wall clock per change (median about 60 min today)

- Enforce the panel ceiling. `skills/flow/review-panel.md`, **No forking, and a wall-clock ceiling
  on every slot**, states 15 minutes per slot, but the conductor blocks on the Agent call and
  observes elapsed time only at the next turn boundary. Gain: caps the p90 (mean 57 min toward the
  23-min median). Loss: a truncated review may miss what it never reached; the record says so.
- Make `primary` read less. It is the only slot that also invokes
  `superpowers:requesting-code-review` and reads `code-reviewer.md` on top of `final-review.diff`
  and the bundle, and it appeared 5–20× slower than the other slots in every change measured.
  Gain claimed: 10–20 min on a typical panel. Loss: the skill's checklist; its measured yield here
  is three findings over seven changes, two of them prose width or indentation. Section 8 withdrew
  the wall-clock gain and narrowed the slot to plan alignment on coverage grounds instead.
- Overlap `verify` with panel pass 1. Gain: median 5 min, mean 26 min hidden under the panel. Loss:
  one wasted verifier dispatch (about $3) when a fix round follows, which is about 60% of runs.
- Minor-only rounds fixed by resuming the last per-task implementer instead of a fresh `panel-fix`
  subagent that re-reads the bundle. Gain: 3–5 min per round. Loss: none in coverage; the conductor
  still runs the reproducer re-run and the diff check.
- The 22 short stages sum to about 3% and are not a wall-clock lever.

### Target B — dollar cost per change (about $45 mean)

- Drop per-task review (`skills/flow/implement.md` §4, **Per-task review**). Gain: one sonnet
  dispatch per task, $1–3 each, 2–4 per change, plus the conductor's relay turns. Loss: a task defect
  is found about ten minutes later by the panel and fixed as a fixup instead of before the next task
  builds on it; `flow tasks tick` would move to "guard passed".
- Size-scale the roster: below a diff-size threshold dispatch `primary` + `mutation` only, the same
  one-way reduction shape **The docs-only reduction** already establishes. Gain: two slots at $2–4
  each per small change. Loss: `principles` (1 finding in 7) and `code-review-low` (3, one Major in
  kan-404 — a real loss on that change).
- Slots with an empty delta are already not dispatched (their 0-second rows are records only).
- The planner on fable costs $94 over 50 brainstorm runs; the model is not the cost.

### Target C — operator attention (about 9 + n prompts, three sittings)

- Merge the convergence confirm and the design approval in `skills/flow/brainstorm-planner.md`
  **Convergence** — `flow.design-approval` has a 0-second median because it is asked seconds after
  "anything still unclear". One prompt, both marks kept. Gain: one prompt per change.
- Collapse self-review's five per-angle filing asks and the rating into one multi-select
  (`skills/flow/archive.md` step 9, canonical in `skills/flow-contracts/finish-contract-run2.md`
  step 9). Gain: five prompts; `flow.self-review`'s 15-minute mean is mostly this waiting.
- A project-level default for the self-review skip prompt, the same shape as `## default landing
  route`. Gain: one prompt.
- Fold the four mark-only chains named in section 1 into one stage each. Touches `skills/flow/SKILL.md`
  **Stage keys**, `README.md` Level 1, `stats/internal/stages/names_test.go` and the four phase
  files. Gain: about 14 fewer mark pairs, roughly 30 parent tool calls, 1–2 minutes and about $1 per
  change. Loss: leaderboard rows for stages that are all under 15 seconds.
- The three sittings are the pipeline's human gates by design; only the archive sitting is
  reducible.

The operator chose Target A.

## 3. Wall clock, round two — panel anatomy, the unrecorded slot, the ceiling, the verify floor

- Panel anatomy, kan-395 (40-minute stage): `primary` 23 min in parallel with `principles` 16 s,
  `code-review-low` 50 s, `mutation` 3 min; conductor glue 1.3 min (reproducers, bundle rebuild);
  `panel-fix-1` 6 min; delta re-run 3 min (only `code-review-low` and `mutation` had a delta; the
  other two slots have 0-second rows); `panel-fix-2` 1.3 min; glue. About 60% is one slot, about 20%
  is fix rounds for Minor findings (F1: an error-message vocabulary list with no test pinning it;
  F3: a 108-character line), about 15% is conductor glue.
- Why `primary` was slow was taken as unrecoverable from the records — wrongly, as section 8
  found: the sidechain transcript of every dispatch survives under the session's `subagents/`
  directory, keyed by the `agentId` the dispatch row records. Its extra inputs are small
  (`requesting-code-review/SKILL.md` 95 lines, `code-reviewer.md` 181 lines). Each slot's report
  goes to `<worktree>/.superpowers/sdd/panel-report-<round>-<id>.md`, which archive cleanup removes
  with the worktree, and the rendered ledger prints `Started:` but no end time and `Tokens: not
  measured` on every dispatch row (`stats/internal/records/render.go`). Only the session transcript
  can say whether it re-ran the test suite, walked every spec, or deliberated. The store does hold
  `endedAt` per dispatch — every duration in these notes came from it — so a duration line is a
  render change, not a new measurement; per-dispatch tokens stay unmeasured because nothing
  attributes sidechain transcript spans to dispatch rows.
- The ceiling is not enforceable as written: kan-402's two `timed-out` rows closed after 16 and 76
  minutes. Every dispatch in `implement.md` and `review-panel.md` is a blocking Agent call; nothing
  polls. kan-402 itself is an overnight-gap outlier (conductor 9 h, dispatches starting at 01:27 and
  02:31 local) and is excluded from the medians above.
- `flow.verify` stage time exceeds the verifier's: kan-404 stage 892 s against verifier 234 s plus a
  299-second `verify-2`; kan-389 stage 906 s with no verifier dispatch recorded at all; kan-395 237 s
  against 200 s. The `## test` list (`scripts/run-guard-tests.sh` about 55 s, `go test ./... -race`,
  `npm test`) is a 3–4 minute floor; the remainder is conductor work (`prepare-workspace.sh`, the
  ledger render, reading the report) or an unrecorded second run. Overlapping the verifier with
  panel pass 1 hides the floor; it does not remove the glue.
- `flow.sdd-tdd` choreography overhead is about 40%: kan-395 stage 541 s against an implementer sum
  of 312 s. Per-task reviewers overlap the next implementer, so the gap is guard, bundle, record
  calls and relay turns between bundles — quantified in section 4.
- Conductor launch to first implementer is 3–5 minutes in every change — quantified in section 4.
- `visual-verify`, `stage-diff`, `run-instructions`, `write-in-progress`: 5–10 s each, already free.

The operator had not watched what `primary` does in the transcript, so recording duration and
preserving the report (lever 1 in section 5) was ordered before any change to `primary` itself;
section 8 read the transcripts directly.

## 4. Wall clock, round three — pre-implementation gap, sdd-tdd choreography, background poll

### (a) Pre-implementation gap

- Split, kan-395: conductor launch → `flow.load-context` begin 11 s; `load-context` 14 s; gap 8 s;
  `flow.isolate-workspace` 147 s; gap 8 s; `flow.sdd-tdd` begin → first implementer launch 38 s
  (bundling, gather, record). kan-390: isolate 178 s, first launch five minutes after the conductor
  started. kan-404 and kan-389: isolate 42–51 s.
- What isolate actually costs in commands: `git worktree add` measures 0.3 s in this repository; the
  `spectre/changes/<name>` copy, `flow workspace-id` and `spectre link` are sub-second. The
  2–3 minutes is `superpowers:using-git-worktrees`, which `implement.md` §2 invokes: its Step 2 runs
  `npm install` / `cargo build` / `pip install` when it finds a manifest at the worktree root (none
  here — `stats/web/package.json` is not at the root — but gymie has one), and its Step 3 "Verify
  Clean Baseline" runs the project's test suite before a line is written. That is the same `## test`
  list `flow.verify` runs at the end, on a HEAD `resolve-base-branch.sh` and `check-base-moved.sh`
  already establish is the base branch.
- Cut: replace the skill invocation with the three commands the stage needs (`git worktree add -b
  spectre/<name>`, the spectre copy, `flow workspace-id`) and no baseline test. Gain: about 2 minutes
  per change here, more where `npm install` fires. Risk: a red base branch surfaces at `flow.verify`
  instead of before implementation; `flow.verify` already blocks with the same diagnostics, and in
  this repository the base is `main` after a merged PR.

### (b) sdd-tdd choreography

- Boundary gaps, implementer end → next launch: kan-395 39 / 33 / 42 s; kan-390 43 / 35 s; kan-389
  55 / 56 s; kan-404 40 / 50 s. Constant 35–55 s per bundle regardless of task size, plus the last
  reviewer running alone (55–90 s) and about 40 s of stage lead-in.
- Every script in a boundary is sub-second: `gather-dispatch-context.sh` 0.15 s,
  `check-task-commit-fields.sh` 0.04 s, a `flow record` call 0.05 s. The 40 s is conductor model
  turns — one Bash call per `flow record end`, `flow record begin`, the guard, the gather, `flow
  tasks tick`, two Agent launches, the `## Stage` relay block — each a 3–8 s round trip with a large
  context. About 8–10 tool calls per boundary.
- Cuts, by payoff per line changed: (1) dropping the per-task reviewer removes one launch, one record
  pair, one tick decision per boundary and the trailing solo reviewer — about 15 s per boundary plus
  about a minute at the end; (2) `implement.md` §4 says to gather each bundle "immediately before that
  bundle's implementer goes out", but every bundle derives from `tasks.md`, which no implementer
  edits, so all bundles can be gathered in one Bash call at stage start; (3) the record end/begin
  pair and the guard can share one Bash call per boundary — the contract requires the calls, not
  separate turns. Expected gain: 20–25 s per boundary, 1–2 minutes on a 3–4-task change, about zero
  on a 1-task change. Risk: only (1) has a quality trade (1 fix in 19); (2)–(3) lose nothing — a
  record write never blocks.
- Not cuttable without changing the model: "at most one implementer subagent in flight against a
  given worktree". Bundles with disjoint `**Files:**` could run in throwaway worktrees and be
  cherry-picked, at the price of a merge step and a conflict path, for a median gain of one
  implementer-length (1–2 min) at the current 2–4-task change size.

### (c) Background dispatch and poll for panel slots

Dropped in section 8: the two breaches that motivated it were a sleeping laptop and a harness
retry, and a background `sleep 900` freezes with the machine. Kept as the record of the mechanics.

- Mechanics: the Agent tool already runs a subagent in the background and delivers a completion
  notification; the blocking shape comes from the rule that the conductor never ends a turn with a
  child in flight. A ceiling needs a timer the conductor hears: `Bash` with `run_in_background` and
  `sleep 900; echo ceiling` yields exactly one notification at 15 minutes; `TaskStop <agent-id>`
  stops a slot by the id the launch returned — the same id `-agent-id` already records. `Monitor` is
  for repeated events and is not needed.
- Shape, in **No forking, and a wall-clock ceiling on every slot**: at pass-1 dispatch, launch the
  slots, then the background timer; on each slot's completion record its `end`; on the `ceiling`
  notification, `TaskStop` every slot still open, close each `-outcome timed-out`, re-dispatch once
  per the existing rule with its own timer. The conductor still cannot end its turn with a slot in
  flight; the timer is a within-turn event the harness delivers.
- Expected gain: none on the median (every slot but `primary` finishes in 16 s–3 min); caps
  `primary` at 15 min on the kan-395 shape (23 → 15 plus a re-dispatch) and turns kan-402's 93- and
  76-minute slots into 15-minute ones; the panel's p90 moves from the 57-minute mean toward the
  23-minute median. Risk: a thorough `primary` on a large diff gets cut twice and reaches the
  operator prompt ("re-dispatch / proceed without / stop"), which today is dead text. Pair the timer
  with an in-prompt budget ("write what you have to the report file by minute 12") so the first cut
  yields a partial report rather than nothing.

Shape change from this round: the worktree skill's baseline test run is a second full test-suite run
per change that nothing else in the pipeline relies on — a deletion worth about 2 minutes, on a par
with the verify overlap.

## 5. Ranked wall-clock levers

Ordered by gain per change on the 60-minute median, as revised by section 8:

1. **Wave-parallel bundles** — a `**After:**` task field; bundles whose declared predecessors have
   landed dispatch together, each in its own throwaway worktree, cherry-picked onto the branch in
   plan order as it returns. Absent field = after every earlier task, so existing plans stay serial.
   One to five implementer-lengths (3–15 min) on the 6–9-bundle plans that are now a quarter of
   changes; nothing on a 1–2-task change; nothing until the planner writes the field.
2. **Verifier alongside panel pass 1** — re-run only after a fix round. About 4 min when no fix round
   follows (about 40% of runs), zero otherwise.
3. **Delete the worktree baseline test, build the SPA once** — `implement.md` §2 runs the three
   commands directly instead of `superpowers:using-git-worktrees`; this repository's `## test` list
   opens with `cd stats && make web-build`, because `stats/internal/web/dist/` is gitignored and
   `//go:embed all:dist` makes a fresh worktree's first `go test` fail without it (kan-389, section
   8). About 2 min, more in a project with a root manifest.
4. **Minor-only fixes via the last implementer** — `SendMessage` to the implementer whose context
   holds the code, recorded under its own `task-<n>-implementer-fix-<k>` key as the per-task fix
   already is, instead of a fresh `panel-fix`. 3–5 min per round.
5. **Delete the per-task reviewer** — one launch, one record pair and one tick decision per boundary
   (about 15 s each) plus the 55–90 s trailing solo reviewer; the coverage reasoning is section 8.
   About 1–2 min on a multi-task change.
6. **Boundary batching** — gather all bundles at stage start, one Bash call per boundary for record
   end/begin plus guard. About 1 min on a multi-task change once the reviewer is gone.
7. **Keep the panel's evidence** — copy every `panel-report-<round>-<id>.md` into
   `docs/superpowers/reviews/` at `flow.preserve-sessions`. No gain; the transcripts already answer
   what a slot did, and the ledger's missing `Ended:` line is a nice-to-have, not a prerequisite.

Withdrawn: **slim `primary`** (no wall-clock basis — the slot is 94–159 s when the machine is
awake; its narrowing to plan alignment is a coverage change, section 8) and the **ceiling timer**
(section 4(c)). Levers 2–6 take about 10 minutes off the median of a 3-task change without touching
review coverage; lever 1 is the only one that scales with plan size.

## 6. Cost

First investigated to one round while the session was on wall clock (the findings from that round
open this section), then reopened by the operator and taken through three rounds.

### One round, from the wall-clock session

- Per-task review is the cheapest deletion: 1 fix outcome in 19 dispatches across seven changes, $1–3
  per dispatch, 2–4 per change. Its quality trade is stated in section 2 and left open in the Open
  list.
- A size-scaled roster below a `check-panel-diff-size.sh` threshold would drop `principles` and
  `code-review-low` on small diffs; `code-review-low` raised kan-404's one Major finding, so the
  threshold is a real trade, not a free cut.
- Minor findings cost a fix round each; lever 5 above reduces the dispatch but not the verification.
- Empty-delta slots are already not dispatched; the planner's model is not a cost lever.

### Round one — the per-role split

Four of the seven changes carry full dispatch attribution (kan-389, kan-390, kan-394, kan-404); the
other three have `session never bound` rows (kan-395, kan-402, kan-380) and cannot be split. Each
dispatch row's own sidechain tokens, at the seeded rates (`stats/internal/store/pricing_seed.go`:
sonnet $2 / $10 per Mtok in / out, $0.20 cache read, $2.50 cache write; fable $10 / $50 / $0.25 /
$12.50):

| Role | per change | share |
|---|---|---|
| conductor's own turns | $8.70 – $14.99 | 50–80% |
| planner (fable) | $0.86 – $6.26 | 5–25% |
| per-task reviewers | $0.19 – $0.74 | 1–4% |
| implementers | $0.41 – $1.54 | 2–6% |
| panel-fix | $0.25 – $0.83 | 1–5% |
| all four panel slots together | $0.68 – $1.63 | 4–8% |
| verifier | $0.31 – $0.63 | 2–3% |
| total | $17 – $24 | |

- The stage leaderboard's "panel $18 / sdd-tdd $23 mean" is the conductor's context churn *during*
  those stages, not the slots: kan-404's `flow.verify` stage cost $5.00 on 15.7 M cache-read tokens
  with a $0.63 verifier inside it.
- The conductor's cost is turns × context. Every `flow record`, `flow stage`, guard, gather, tick and
  relay block is one turn re-reading the whole context at $0.20 per Mtok.
- kan-380 (pre-conductor; kan-377 moved this work out of the parent) is unattributed, so whether the
  conductor added cost or only relocated it from the parent session is not answerable from the store.
- Re-priced levers: dropping per-task review saves $0.2–0.7 per change in reviewer tokens plus about
  four conductor turns per task — 3–5% of a change; `flow tasks tick` (`stats/cmd/flow/tick.go`)
  only flips checkboxes and `check-unfinished-work.sh` counts column-0 boxes without knowing about
  review, so the gate would move to "guard passed" in one sentence of `implement.md` §4.
- **Size-scaled roster: rejected.** Implementation diff sizes: kan-404 79 lines, kan-395 147, kan-402
  399, kan-390 450, kan-380 513 (the cap is 5000). At a 300-line threshold only kan-404 and kan-395
  would reduce, saving about $0.30 each — and kan-404's dropped `code-review-low` raised that
  change's one Major finding. It saves 1–2% and would have lost the most severe finding in the
  sample.
- The operator chose to investigate batching the conductor's calls.

### Round two — the conductor's turns, enumerated from the phase files

A turn is one assistant message; the harness lets one message carry several tool calls and a Bash
call chain several commands, so a turn is spent only where the model must read an output before
choosing the next action.

- **Task boundary today** (`implement.md` §4, bundle N+1 just committed): `flow record dispatch end`
  for the implementer; `check-task-commit-fields.sh`; `gather-dispatch-context.sh` + `test -f` for
  bundle N+2; Agent launch, reviewer N+1; `record begin` reviewer; Agent launch, implementer N+2;
  `record begin` implementer; reviewer returns → `record dispatch end`; `flow tasks tick`. Nine turns;
  ten to twelve when a per-task fix folds in. **After:** one Bash (record end + guard + gather); one
  message with both launches; one Bash with both `record begin`s — the very next action after the
  launches return; one Bash on reviewer return (record end + tick). Four turns; three without the
  per-task reviewer.
- **Panel round today** (`review-panel.md`, pass 1 through a one-fix-round close, three findings):
  pre-work guards one per turn (base movement, citation, bundle rebuild, relocation comparison,
  diff size, docs-only, `final-review.diff`, principles/standards paths, throwaway worktree), four
  launches, four `record begin`, four completions with `record end` + `test -s`, `record finding`
  per finding, worktree removal, `final-review-panel.md`, `check-panel-reproducers`,
  `run-reproducer` per finding, bundle rebuild, fix launch + `record begin`, fix return, reproducer
  re-runs + diff walk + `record status fixed` per finding, mutation-proofs, fixup check, the delta
  re-run's own set, `check-panel-findings-closed`, render, stage end. About 55–75 turns. **After:**
  pre-work in one Bash (every verdict printed; only the dispatch depends on them); four launches in
  one message; `record begin`s in one Bash; completions in one or two; all `record finding` calls in
  one; reproducers + worktree removal in one; fix launch and record; fix return with every
  reproducer re-run and the diff printed in one; all `status fixed` in one after the model judges
  each diff touch; mutation-proofs chained; delta re-run in three or four; close in one. About 22–26
  turns.
- **Verify/handoff tail today** (`verify-and-handoff.md`): `prepare-workspace`, `project-get`
  lint/test, verifier launch, `record begin`, return → `record end`, ledger render, five stage-mark
  pairs, visual trigger, `git status`/`log`, `project-get run`/`apps`, `flow state set`,
  `journal-count`, `cost-status`. About 25 turns. **After:** six — each stage's `end` mark and the
  next stage's `begin` mark ride the next stage's first command.
- **Stage marks and relays:** 18 `flow stage` calls, each its own turn today, fold to zero extra
  turns; the nine `## Stage` returns each cost a conductor turn and a parent turn and stay unless
  **Progress visibility** is amended (round three).
- **Contract sentences that shape the batches, none forbidding them:** `implement.md` §4 "`begin`
  carries `-agent-id <id>` and is recorded immediately after the launch returns, before any other
  action" (two launches in one message, two `begin`s in the next Bash, satisfies it);
  `review-panel.md` "Record it per finding, as its verdict is reached, not batched at the round's
  end" (its purpose — an aborted round keeps verified findings closed — survives one Bash recording
  every verdict reached in that same turn; one sentence to reword); `pipeline.md` **Progress
  visibility** (keeps the relays). Unaffected: "never end a turn with a child in flight", "never read
  the bundle back", "read the outcome word, not the exit code", the literal-token rule.
- Estimate for a 3-task change with one fix round: about 170 turns today → about 65.

### Round three — verified against kan-404's conductor transcript

`~/.claude/projects/-Users-tweety53-Projects-agents/5d204fa9-d2bd-4e22-8bf8-25b8710cc811/subagents/agent-a0a0c76b220754d6f.jsonl`
(2 tasks, one panel fix round, one verify re-run):

- **191 API calls** (unique message ids — the JSONL's 419 assistant records split one call into text
  and tool blocks), 221 tool calls, never more than one tool call per message. Cost from its own
  usage rows: **$9.02**. Context 46 k tokens at the first call (harness, `CLAUDE.md`, rules, the
  dispatch prompt — before reading anything), 87 k after the three phase-file reads (+41 k, so about
  3 characters per token), 111 k at the first implementer launch, 240 k at the handoff. Average
  $0.047 per call.

| Window | API calls | idle polls | what else |
|---|---|---|---|
| load-context + isolate | 28 | 0 | 3 phase-file reads, contracts, `spectre validate`, plan repair, worktree skill steps |
| sdd-tdd (2 tasks) | 32 | 4 | one boundary = 17 calls, including an Agent launched only to "wait for the notification" and `echo`/`sleep`/`git log` polls |
| review-panel | 48 | 8 | 12 single-guard shells before the launches; 4 launches in 4 turns; 4 `record begin` batched in one call; 4 `record end` as each arrived |
| verify | 66 | 17 | the verifier reported a lint failure and the conductor fixed it inline — about 35 calls — then a second verifier |
| visual + tail | 17 | 0 | five stage-mark pairs, `state set`, `journal-count`, `cost-status` |

- **Idle polling is 29 of 191 calls (15%, about $1.3, about 2 minutes).** The conductor cannot await
  a child without making a tool call, so it emits `echo "still waiting"` every 1–2 seconds until the
  completion notification lands. The harness's own guidance is one foreground
  `until <condition>; do sleep 5; done`: for an implementer the condition is a new commit
  (`git -C <wt> log -1 --format=%H` changed), for a reviewer or slot `test -s <report path>`, for
  the verifier `test -s` on its report. One turn per wait.
- **The inline lint fix inside `flow.verify` is a hidden fix round.** `verify-and-handoff.md` says
  the verifier "fixes nothing, edits no source"; nothing says the conductor may not. It changed
  reviewed code after the panel closed — which **Panel re-runs**' stale-result rule should have
  caught — with no dispatch record for the fix and no slot re-run, at about 35 turns of a 240 k
  context (about $1.8) and 8 minutes. A contract gap, not a batching item: dispatch a fix and re-run
  the affected slots, or hand back.
- Revised before/after for this change: 191 → about 62 (load 12, sdd-tdd 12, panel 20, verify 8 when
  clean, visual 2, tail 5, relays as today, 3 reads). **$5.5–6.5 saved of $9** and, at the
  transcript's 4–6 s per call, **9–11 minutes.** The round-two estimate was right in ratio and low in
  absolute terms because it did not count polls.
- **Tail relay amendment.** The five stages after the panel each end with a conductor turn-end, a
  parent turn on a 200 k+ session context and a `SendMessage continue` — about $0.10 and 10–20 s each
  for stages whose medians are 5–10 s (except `verify`). Smallest form: `pipeline.md` **Progress
  visibility**, the paragraph "On the implementation branch the granularity is the stage", gains that
  the conductor returns one `## Stage` for the four stages after `flow.verify` (which still returns
  on its own since it can block), naming all four outcomes; `pipeline-rationale.md` gets the reason;
  `implement.md` **Dispatch the conductor**'s `## Stage` bullet and `verify-and-handoff.md` carry the
  exception in one sentence each. **Nothing moves in `SKILL.md`'s Stage keys, `README.md`'s Level 1
  table, `stats/internal/stages/names.go` or `names_test.go`** — keys and marks are unchanged, only
  the visibility relay coarsens; merging the four keys into one `flow.handoff` would touch all four
  for the same saving. The parent's task list shows the four entries open together for about 30 s,
  within the trade `conductor-stage-returns` already accepted. Saving about $0.30–0.40 and 45 s per
  change.
- **The conductor's context.** Fixed load: 46 k (not reducible from this repository) + 41 k for
  `implement.md` (443 lines), `review-panel.md` (764), `verify-and-handoff.md` (477) + about 10 k for
  `worktree-resolution.md`, `artifacts-registry.md`, the change artifacts and `ls scripts`. The 41 k
  of phase files costs about $0.008 per call — about $1.5 per change at 191 calls, about $0.5 at 62 —
  and a larger context is a slower call. Read against the files: `review-panel.md` is roughly 35–40%
  rationale (the KAN-366 collision, the `git diff HEAD --binary` semantics, the diff-cap history,
  design.md citations on nearly every rule, the throwaway-worktree reasoning); `verify-and-handoff.md`
  roughly 30% (the protected-daemon paragraph three times, the gymie worked example, the `embed.go`
  measurement); `implement.md` about 20%. A split on the contracts' existing pattern —
  `implement-rationale.md`, `review-panel-rationale.md`, `verify-and-handoff-rationale.md`, never
  loaded by a run, cited from the procedure file the way `pipeline.md` cites `pipeline-rationale.md`
  — takes the load from about 41 k to 26–28 k tokens with no generated copy; `scripts/check-references.sh`
  already polices the citations. A generated condensed copy was rejected: two files saying the same
  thing.
- **Cost ranking, revised:** (1) batch calls and replace idle polls with one `until` wait — $5.5–6.5
  and about 10 min per change; (2) close the inline-work gap in `flow.verify` — the conductor
  neither edits source nor runs the `## lint`/`## test` lists itself; a failed verifier is
  re-dispatched with its failure or handed back — $1.8 and 8 min on kan-404, about 4 min and a
  second verifier on kan-389 (section 8), plus correctness; (3) the rationale split — about $1 per
  change and faster calls; (4) the tail relay — about $0.35; (5) per-task review deletion, decided
  in section 8 — about $0.4 and the boundary turns; size-scaled roster rejected.

## 7. Operator attention

First investigated to one round while the session was on wall clock (the findings from that round
open this section), then reopened by the operator and taken through two rounds.

### One round, from the wall-clock session

- Convergence confirm and design approval are two prompts seconds apart (`flow.design-approval`
  median 0 s); one prompt with both marks kept is a pure win.
- Self-review's five per-angle filing prompts plus the rating dominate `flow.self-review`'s 15-minute
  mean; one multi-select with the angle label on each option, plus a project default for the skip
  prompt, removes six prompts.
- The four mark-only stage chains can fold into one stage each at the cost of leaderboard rows that
  are all under 15 seconds; the README Level 1 table, `SKILL.md` **Stage keys**, `names_test.go` and
  the phase files move together.
- The three sittings (brainstorm, `IN_PROGRESS` review, archive) are the pipeline's human gates and
  stay.

### Round one — every prompt the pipeline can fire

Creating run → integrate → archive, with the file that fires it, whether it is a gate by contract or
incidental, and the wait where the store or kan-404's parent transcript
(`~/.claude/projects/-Users-tweety53-Projects-agents/5d204fa9-d2bd-4e22-8bf8-25b8710cc811.jsonl`)
captures it:

| # | Prompt | Fired by | Contract or incidental | Measured wait |
|---|---|---|---|---|
| 1 | Which change? / what to build? | `pipeline.md` **Change name resolution**; `brainstorm.md` A | contract, conditional | — |
| 2 | Which staged research note? (glob matched more than one) | `brainstorm-planner.md` **Seed from a staged research note** | conditional, defensive | never fired in the sample |
| 3 | n brainstorming questions (rounds 1–2, no cap) | `brainstorm-planner.md` **The checklist**, relayed per `brainstorm.md` | contract — the design gate's substance | kan-404: 2 questions, answered in 56 s and 15 s; `flow.brainstorm` median 12 min over 50 runs |
| 4 | Round ≥ 3 offer | **Convergence** | contract (no-cap rule) | conditional |
| 5 | Convergence confirm | **Convergence** | contract — **Stage exit — never the command's own judgment** | kan-404: 36 s |
| 6 | Design approval (HARD GATE) | **The checklist**; marked `flow.design-approval` | contract | kan-404: 5 s; store median 0 s, mean 107 s (49 runs) |
| 7 | Where should this fix go? | `implement.md` §3 | contract, fix runs only | `flow.document-fix` n = 1 |
| 8 | Base moved with overlap — panel | `review-panel.md` **Check base movement first** | conditional | not fired in the sample |
| 9 | Diff over cap | `review-panel.md` **The roster** | conditional (cap 5000) | not fired |
| 10 | Slot breached the ceiling twice | **No forking…** | conditional — dead text while the ceiling is unenforced | never |
| 11 | Finding did not converge | **Panel re-runs** handback | conditional | not fired in the sample |
| 12 | Plan-quality repair, BLOCKED, empty worktree set | `implement.md` conductor relay list | conditional | not fired |
| — | The `IN_PROGRESS` human gate | `pipeline.md` **States** | contract — the state is the gate, no prompt | kan-404: handoff 10:23:47 → bare `/flow` 10:24:05, **18 s** |
| 13 | Unfinished work | `integrate.md` §1 | conditional | `flow.unfinished-work-gate` median 9 s |
| 14 | Base moved with overlap — integrate | `integrate.md` §2 | conditional | — |
| 15 | How should this branch land? | `integrate.md` §2 | contract, skipped here by `## default landing route` | `flow.landing-question` median 30 s |
| 16 | PR-exists confirmation on a forge without CLI | `integrate.md` §4 | conditional | — |
| 17 | Jira follow-up confirmation | `jira-followups.md` | conditional, only on 13's third option | — |
| 18 | Run self-review? | `archive.md` step 9 | contract, "running it the default" | kan-404: **20 min** (10:30:39 → 10:50:37, then "No"); `flow.self-review` median 53 s, mean 14.6 min over 44 runs |
| 19–23 | File any of this angle's findings? × 5 | `archive.md` step 9, canonical in `finish-contract-run2.md` step 9 | contract | inside 18's stage |
| 24 | Rate this run 1–5 | same | contract | inside 18's stage |
| 25 | Cleanup leftover | `archive.md` step 7 | conditional — stops, not a prompt | `flow.verify-cleanup` median 14 s |
| 26 | Wrong-state override | `pipeline.md` **Wrong state** | conditional | — |
| x | "Staged-but-uncommitted design spec on main — copy / leave / discard" | none — improvised at integrate | **incidental**, a planner side effect | kan-404: 18 s |

- kan-404 fired 3, 3, 5, 6 (the planner's four prompts, about 2 minutes), x, and 18 (answered "No"
  after 20 minutes, so 19–24 never fired): six prompts, about 2.5 minutes of answering, one 20-minute
  absence. With the landing default configured and self-review declined, the whole non-brainstorm
  pipeline asked one contract question and one incidental one.
- **The `IN_PROGRESS` gate was 18 seconds** — the operator did not open the diff or run anything
  between the handoff and the bare `/flow`. Asked which of three readings was true, the operator
  first answered "I review later, in the PR / IntelliJ"; section 8 corrected that — this
  repository's landing route is `merge and push`, there is no PR, and the actual reason is **"I just
  don't review agents repo changes at all, only when I need to check the stats app UI."** The gate
  is closed quickly because review mostly does not happen, not because it happens elsewhere. It
  stays as the pipeline's human gate; auto-integrate on a clean panel was offered twice and
  declined both times. The thread is the three merges below.
- The convergence confirm and the design approval are one decision asked twice, 36 s and 5 s apart.
- Self-review is where attention leaks: the skip prompt fires only after `FINISHED` is written, which
  is when the operator walks away.
- The incidental prompt was a defect: the planner saved the design spec into the main checkout's
  `docs/superpowers/specs/` and staged it; nothing moves it into the worktree, so `integrate.md` §3
  found a dirty main checkout.

### Round two — the three merges, section by section

**Merge 1 — the brainstorm double gate.**

- `skills/flow/brainstorm-planner.md` **Convergence**: the confirm prompt gains a third option and
  becomes the approval — *Nothing unclear — approve the design and move on* (recommended) / *Another
  round — I have something* (default on silence, unchanged) / *Revise — I have a change to the
  design*. The sentence "the design approval the HARD GATE requires is a separate, later act by the
  same operator and gets its own mark" is replaced: the same answer closes the checklist and is the
  approval; the parent still marks `flow.brainstorm` end, then `flow.design-approval` begin and end,
  around that one relayed answer.
- **The checklist**, HARD GATE bullet: "do not run `spectre new` until the user approves the design"
  stays; "approves" means the merged prompt's first option. The existing scoped override already
  removed the per-section approvals; this is the last duplicate.
- `skills/flow-contracts/pipeline.md` **Stage exit — never the command's own judgment**: unchanged in
  substance; one clause noting the same explicit answer may close the checklist and grant the
  approval. `pipeline-rationale.md` gets the measurement.
- `skills/flow/brainstorm.md` **The relay**: the three relayed prompts become the merged one.
- `skills/flow-contracts/operator-prompts.md`: nothing — the shape is preserved, and this remains
  the one prompt where the safe default and the recommended option differ.
- Guards and tests: none parse either prompt; stage keys are unchanged, so `README.md` Level 1 and
  `names_test.go` stay.
- Risk: the beat between "nothing unclear" and "approve" is lost; the design is presented before the
  confirm, and *Revise* covers wanting a change. Prompts per change: −1.

**Merge 2 — self-review: a project default, and one prompt.**

- `skills/flow-contracts/project-configuration.md`, optional-keys table: a new `## self review` row,
  the single-line-literal shape of `## default landing route` — body `run` or `skip`, matched
  byte-for-byte after trimming; anything else reported by name and dropped, resolving as absent.
  `project-get.sh <main-checkout> "self review"` reads it (exit 1: absent); no new script.
  `scripts/check-model-keys.sh` validates only the two model keys; a `## self review` body is
  validated by the phase file the way the landing route is today, or by one more case in that guard
  if a hard check is wanted.
- `skills/flow/archive.md` step 9: resolve the key first; `skip` → step 9 ends with `Self-review:
  skipped — project default`; `run` → no prompt; absent → the prompt as today. The guardrail "Never
  ask the self-review skip prompt … before `FINISHED`" is unchanged.
- `skills/flow-contracts/finish-contract-run2.md` step 9 (canonical): "one multi-select prompt per
  angle" → one multi-select prompt for the whole pass, each option prefixed with its angle label,
  default *None — file nothing*; the rating rides in the same `AskUserQuestion` call (kan-404's
  parent transcript already issues a four-question call for `/flow-settings`). `archive.md` step 9's
  wording follows. "Every finding is explained in the message body before any prompt fires" and the
  per-angle none-marker in the report are unchanged.
- `skills/flow-contracts/operator-prompts.md` **The multi-select variant** names two call sites; the
  second, `skills/myflow-do/SKILL.md`, no longer exists — stale text to fix in passing.
- Guards and renderers: `scripts/check-self-review-report.sh` parses the report (five `##` sections,
  one `- **[label]** … — filed: KEY | declined` line per finding), not the prompts — unchanged.
  `stats/` has no self-review renderer and records no self-review dispatch.
- **Asking before `FINISHED`.** Neither "Never ask the self-review skip prompt … before `FINISHED`"
  (`archive.md` guardrails) nor "a skip, a failure, or a decline never moves the change off
  `FINISHED`" (`finish-contract-run2.md`) is violated by asking the skip question at run 1 beside the
  landing question: nothing about self-review would block a state write, and the answer is a stored
  intent, not a stage. What blocks it is where to store it: the state file is a closed schema
  (`stats/internal/api/changes.go` decodes with `DisallowUnknownFields`), so a `selfReview` field
  means a `state-file.md` shape change, the Go DTO, carry-forward at every write site and the
  fallback file — and run 2 on the PR route is a separate invocation, possibly days later. The
  project key answers the same need with no schema: asked once, before `FINISHED`, for every change.
  Per-change intent without a schema change exists only on the merge-and-push route (same
  invocation). Recommendation: the key.
- Risk: a `skip` default stops the self-review reports — the pipeline's own improvement loop, which
  this session drew on — from accruing; the choice becomes explicit rather than a 20-minute prompt
  answered "No". Prompts per change: −1 with the key, −5 when it runs. **This repository's body is
  `skip`** (section 8): the series already ended at kan-380, and its yield was one Jira ticket per
  three reports.

**Merge 3 — the staged design spec (incidental prompt → zero).**

- `skills/flow/brainstorm-planner.md` **The checklist**, first bullet, saves the design to
  `<project>/docs/superpowers/specs/YYYY-MM-DD-<name>-design.md` and stages it — in the main
  checkout, since the planner runs before any worktree exists. **C** creates `spectre/changes/<name>/`
  beside it; `implement.md` §2 copies only `spectre/changes/<name>/` into the worktree and removes the
  main checkout's copy. `integrate.md` §3 then finds a dirty main checkout (kan-404's prompt at
  10:26:26) and `prepare-archive-branch.sh` refuses on one at run 2 (the KAN-271 note in §2 is the
  same class).
- Fix: one sentence in `implement.md` §2 **First run only** — copy the design spec into the same
  relative path in the worktree alongside the change directory and remove the main checkout's staged
  copy; `docs/superpowers/` is already a planning path `commit-split.sh` commits in the second commit
  (`git-boundaries.md`). The checklist bullet gains "in the main checkout; `implement.md` §2 moves it
  into the worktree".
- Guards: `check-unfinished-work.sh`, `check-finish-preflight.sh`, `prepare-archive-branch.sh` read
  git state and benefit; none parses the spec path. Risk: none. Prompts per change: −1 incidental.

**Prompt count**, an ordinary creating run + integrate (default landing route) + archive in this
repository: today n + 5 when self-review is declined, n + 10 when it runs (n = the brainstorm's own
questions). After the three merges: **n + 1** with the key set to `skip`, n + 2 with `run` and one
combined filing/rating prompt. The brainstorm's n questions and the `IN_PROGRESS` gate stay.

## 8. The open items, closed

A third session on 2026-09-04 took each item of the first two sessions' Open list through two
rounds of investigate-then-ask and closed it. Sources beyond the store: every dispatch's sidechain
transcript (`<session>/subagents/agent-<agentId>.jsonl`), the parent and conductor transcripts, and
`pmset -g log`.

### What `primary` spends its time on — nothing; the two long slots were stalls

- Pass-1 `primary` durations across the seven changes: kan-380 159 s, kan-389 103 s, kan-390 94 s,
  kan-404 96 s, kan-394 481 s (all four slots' `end` records landed within 20 s of each other — a
  batched record, not slot time), kan-395 1377 s, kan-402 5570 s. Bugbot was the longer slot in
  kan-380 (261 s), kan-389 (159 s), kan-390 (297 s) and kan-404 (120 s).
- **kan-395's 23 minutes were about 6 minutes of work and one 997-second stall on a single Bash
  call.** 45 tool calls in the first 293 s: the diff and bundle, the guard tests, `make ui-test-up`,
  `npm ci`, two vite builds, the served bundle compared against `dist`, then an attempt to append a
  probe line to `stats/web/src/App.tsx` and a `git stash` — both denied by the auto-mode classifier
  (a review slot mutating the worktree is what KAN-366 forbids; the classifier held). The next call,
  `scripts/check-normative-inventory.sh 2>&1 | tail -8`, went out at 17:26:21 local and returned at
  17:42:59 with normal output; the script runs in 0.5 s. The machine was awake and lid-open the
  whole time (`pmset -g log`: `lidopen`, `UserIsActive`, no Sleep entry), the conductor was idle
  waiting on this slot, and the operator was away and answered no prompt. Every Bash call over five
  minutes in this project's sessions that is not on a sleeping machine ends in either "Command did
  not complete within its 120s timeout and was moved to background" or **"claude-sonnet-5[1m] is
  temporarily unavailable (timed out)"**, and every one of the latter is 908–1005 s long — the
  harness's API retry budget. The permission classifier's own API call went into that loop and
  came out the far end. Nothing in `/flow` can shorten it.
- **kan-402's 93-minute `primary`, 76-minute bugbot, 2.8-hour `panel-fix-1` and 9-hour conductor
  are a laptop with the lid closed.** The slot made 26 tool calls separated by gaps of 982, 935,
  999, 649, 912 and 935 s, and the parent, the conductor and every subagent have gaps at the same
  wall-clock moments. `pmset -g log`: "Entering Sleep state due to 'Maintenance Sleep'" at 05:31:51
  local, then DarkWakes of 45–60 s at 05:47:39, 06:03:40, 06:21:02, 06:31:52, 06:48:01 and
  07:04:10 — the pipeline advanced only inside those windows. The `caffeinate -i -t 300` Claude
  Code spawns every few minutes holds off idle sleep only; `pmset -g` is `sleep 1`, `powernap 1`,
  and lid-closed sleep ignores `-i`. The lid closing was incidental, not a deliberate overnight
  run; no operator-side fix is recorded.
- Consequences: lever "slim `primary`" has no wall-clock basis; the ceiling timer is dropped (a
  background `sleep 900` sleeps with the laptop, and cutting the slots would have been wrong in
  both cases); lever 1's duration half is unnecessary because the transcripts already answer the
  question; the store's durations are wall clock including sleep, so medians only, and kan-402 is
  excluded from every wall-clock figure by name. **No store change**: the leaderboard is
  cost-ranked and already reports mean, median and p90 — a frozen process spends no tokens, so its
  cost figures were never distorted — it has no duration column, its filters (`from`, `to`,
  `project`, `change`, `model`) include and never exclude, and teaching the harvester to read the
  power log is a dependency for one incident.

### `primary` narrowed to plan alignment

- `requesting-code-review` is a dispatcher's skill: `SKILL.md` tells the caller how to dispatch a
  reviewer, and `code-reviewer.md` is a prompt template with `[DESCRIPTION]/[PLAN]/[BASE_SHA]/
  [HEAD_SHA]` placeholders and a "Strengths / Issues / Recommendations / Assessment" shape. Handing
  it to the slot asks the slot to self-brief with a template meant for whoever briefs it. Only
  kan-380 and kan-390's `primary` invoked it at all; the other five ran on the conductor's brief and
  wrote the same "Verdict / What was checked / Reproduced" report.
- Findings by slot across the seven changes: bugbot 11 (2 Major/Medium), `code-review-low` 4 (2
  Major), principles 2 (1 Major), mutation 2 (Minor), `primary` 3 — a list item indented three
  spaces (rated Important), a 108-character line, a paragraph's wording. No code defect.
- Decision: **plan alignment only.** The roster row drops the skill reference and the
  `code-reviewer.md` path sentence; `primary` is briefed on `proposal.md`, `design.md` and
  `tasks.md`'s Files/Tests/Commit fields against the whole branch; code quality is
  `code-review-low`'s and bugbot's. The docs-only reduction and the empty-store fallback still
  resolve to `primary` unchanged.

### Per-task review deleted

- The one `fix` in nineteen (kan-402 task 1) was a fence regex at column 0 where the cited grammar
  allows 0–3 leading spaces; the reviewer itself rated it "not a blocker given zero real-world
  impact and full test pass". The plan's `verified:` tag had asked the implementer to confirm that
  regex and it had not. The early-catch benefit did not materialise: task 2 was already implemented
  and reviewed against the unfixed task 1 when the fix landed.
- After the `primary` decision the whole-branch `primary` is exactly the per-task reviewer's
  spec-compliance mandate at branch granularity, `code-review-low` and bugbot are its code-quality
  half, and `check-task-commit-fields.sh` checks Files/Tests/Commit per task mechanically.
- Decision: **delete it.** `flow tasks tick` on guard-passed — one sentence in `implement.md` §4;
  the `-role reviewer` rows and "the last bundle's reviewers run alone" go; `tick.go` and
  `check-unfinished-work.sh` know nothing about review. The deterrent value of a reviewer following
  every implementer is unmeasured and recorded as such.

### Wave-parallel bundles — the lever the task counts reopened

- Task counts across the 48 changes in the tree: 1–4 tasks in 37, 5+ in 12, and five of those
  twelve are the last week's `/flow` changes (kan-372 9, kan-374 6, kan-377 8, kan-378 8, kan-379
  7). The seven-change sample the earlier sessions measured was the small end.
- Bundles are file-disjoint by construction (`plan-dispatch-bundles.py` joins on a shared
  `**Files:**` path or `**Squash-with:**` only). Re-run over the recent plans: kan-377 → 8
  singleton bundles, kan-374 → 6 singletons, kan-379 → 5 bundles of 7 tasks, kan-378 → 4 of 8. An
  8-bundle plan is eight serial implementer-lengths plus seven 40-second boundaries, of which the
  file-overlap rule forces none; `implement.md`'s "at most one implementer in flight against a
  given worktree" leaves "dispatches into different worktrees free to run concurrently", and
  `review-panel.md` already runs the throwaway-worktree mechanics.
- What `**Files:**` cannot see is semantic order — task 3 calling a helper task 1 adds in another
  file. Plan order between every pair is total, so "dispatch a bundle when its plan-order
  predecessors have landed" is serial; parallelism needs a declared independence.
- Decision: **an `**After:**` task field, `Task <ids>` or `none`, the same line-scoped id-list
  shape `plan_grammar.py` already parses for `**Squash-with:**`; `plan-dispatch-bundles.py` prints
  each bundle's `after` set; `check-plan-shape.sh` validates it; the planner's plan template gains
  the field. Absent = after every earlier task**, so every existing plan runs serially as today and
  the planner opts a task into a wave explicitly. Waves: every bundle whose `after` set has landed
  dispatches together, each into its own throwaway worktree; each returns and is cherry-picked
  onto the change branch in plan order (a pick conflict is impossible on disjoint paths), then
  `check-task-commit-fields.sh` runs on the picked commit and `flow tasks tick` is unchanged; a
  guard failure after the pick hands the bundle back to its implementer with the rebased tree.

### kan-389's unrecorded verify — two verifiers and an inline re-run

- kan-389 is the change that introduced the dedicated verifier; the installed `flow` CLI predated
  the branch and rejected `-role verifier`, so neither of its two verifiers was recorded.
  `TestRecordAcceptsVerifierRole` now pins the role. The 906 s decompose as: verifier 1, 329 s;
  the conductor itself, about 4 min — read the failing Go test, `make web-build`, `go vet`,
  `go test ./...`, `npm test`; verifier 2 "SPA now built", 258 s; ledger and marks. The first
  verifier hit Go tests that need the embedded SPA `dist`, a fresh-worktree fact and not the
  branch's defect. The verify numbers stand; the 3–4-minute floor is the `## test` list.
- Decision: **cost lever 2 is widened** — the conductor never edits source after the panel and
  never runs the `## lint`/`## test` lists itself; a failed verifier is re-dispatched with its
  failure, or the run hands back. Its verify-stage work is `prepare-workspace.sh`, the dispatch and
  the ledger render. **And the `dist` build joins lever 3 (section 5)** as the first line of this
  repository's `## test` list.
- Two facts surfaced here and left in the Open list: a change that extends the `flow` CLI runs its
  own pipeline on the stale installed binary, and kan-395's conductor waited by ending its turn.

### Self-review key, planner question count, and the rest

- **`## self review`: `skip` for this repository.** `docs/self-review/` holds 30 reports over 48
  changes, the newest kan-380's; kan-389 through kan-404 all declined. Across the 30: 75 findings
  declined, 15 filed into 9 Jira tickets (KAN-56, 77, 205, 206, 239, 287, 381, 384, 388) — one
  ticket per three reports at one fable dispatch and a 15-minute prompt wait each. `skip` ratifies
  the last six changes and ends the series; `run` with merge 2b's one prompt was the alternative.
- **The planner's n needs no change.** Question blocks in the planner transcripts that carry an
  agent id: kan-390 4, kan-394 2, kan-402 3 — the two mandated rounds plus at most two.
  `flow.brainstorm`'s wall clock (kan-404 3 min, kan-389 19 min, kan-380 49 min, kan-395 57 min) is
  operator thinking, not question count; merge 1 is the only prompt-count lever there.
- **The per-change self-review intent** (item 7) stays declined for the project key; the
  `DisallowUnknownFields` schema fact stands.
- **Auto-integrate on a clean panel** stays declined, with the reasoning corrected in section 7.
- **The conductor's call count scales linearly with tasks; its idle polls do not.** kan-395's 4-task
  conductor: 159 API calls, 150 tool calls, $9.96, context 46 k → 285 k — load-context 6, isolate
  14, sdd-tdd 41 (kan-404: 32 for 2 tasks), review-panel 65 (two fix rounds), verify 9, tail 19.
  About 10 calls per task; the $9–10 per change holds at 4 tasks. It made **1** idle poll against
  kan-404's 29 because it ended its turn with "Still waiting on the primary reviewer" and the
  harness resumed it with the completion notification twenty minutes later at zero calls. That is
  cheaper than cost lever 1's `until` wait and violates the relay contract's "never end a turn with
  a child in flight"; it worked only because the parent ignores a conductor turn that is none of
  the three blocks. The `until` wait is the contract-clean form.

## Step-by-step breakdown

### Wave-parallel bundles with an `After:` field (wall-clock lever 1)

**What:** A `**After:** Task <ids>` or `**After:** none` task field; every bundle whose `after` set has
landed dispatches together, each into its own throwaway worktree, and is cherry-picked onto the
change branch in plan order as it returns, guard-checked after the pick. Absent = after every earlier
task.
**Why:** A quarter of changes are now 6–9 file-disjoint bundles run serially for no reason the
bundling rule imposes; the safe default keeps every existing plan serial and lets the planner opt a
task into a wave.
**Uses:** `scripts/lib/plan_grammar.py` (the `**Squash-with:**` field shape), `scripts/plan-dispatch-bundles.py`,
`scripts/check-plan-shape.sh`, `skills/flow/brainstorm-planner.md` plan template, `skills/flow/implement.md`
**4. Execute (SDD + TDD)** and the one-implementer-per-worktree rule, `skills/flow/review-panel.md`
**The throwaway worktree**, `scripts/check-task-commit-fields.sh`, `git cherry-pick`.

### Keep the panel's evidence (wall-clock lever 7)

**What:** Copy each slot's `panel-report-<round>-<id>.md` out of the worktree into
`docs/superpowers/reviews/` when the session records are preserved.
**Why:** The reports are deleted with the worktree; the sidechain transcripts that also answer what
a slot did are per-session files nothing copies out either.
**Uses:** `skills/flow/integrate.md` **3. Commit the staged work** (`flow.preserve-sessions`),
`skills/flow-contracts/session-records.md`, `scripts/commit-split.sh`.

### Narrow the primary slot to plan alignment

**What:** Drop the `superpowers:requesting-code-review` reference and the `code-reviewer.md` path from
the roster; brief `primary` on `proposal.md`, `design.md` and `tasks.md`'s Files/Tests/Commit fields
against the whole branch and nothing else.
**Why:** The skill is a dispatcher's template the slot self-briefed with, its three findings in seven
changes were all prose, and its code-quality half duplicates `code-review-low` on the same diff; the
docs-only reduction and the empty-store fallback still name `primary`.
**Uses:** `skills/flow/review-panel.md` **The roster**, the `primary` row and the subagent-facing-file
sentence; `skills/flow/SKILL.md` **Model resolution** (unchanged).

### Delete the per-task reviewer

**What:** No reviewer dispatch per bundle; `flow tasks tick` marks a task once `check-task-commit-fields.sh`
passes; the `-role reviewer` rows and the trailing solo reviewer go.
**Why:** One harmless finding in nineteen dispatches, caught after the next task had already built on
it; the whole-branch `primary`, `code-review-low` and bugbot carry the mandate.
**Uses:** `skills/flow/implement.md` **4. Execute (SDD + TDD)** (**Per-task review** paragraph),
`scripts/check-task-commit-fields.sh`, `flow tasks tick`.

### Verifier alongside panel pass 1 (lever 2)

**What:** Dispatch the `flow.verify` verifier at the same moment as the pass-1 slots; if a fix round
lands, dispatch it again after the round; keep every mark and block decision where it is.
**Why:** Both read the same HEAD when the panel is clean, and the verifier's 3–4 minute floor is
serial today.
**Uses:** `skills/flow/verify-and-handoff.md` **The verifier dispatch**, `skills/flow/review-panel.md`
pass-1 dispatch, `scripts/prepare-workspace.sh`, `scripts/project-get.sh`, `scripts/check-spec-reach.sh`.

### Delete the worktree baseline test, build the SPA once (lever 3)

**What:** Replace the `superpowers:using-git-worktrees` invocation in `implement.md` §2 with
`git worktree add -b spectre/<name>`, the `spectre/changes/<name>` copy-and-remove, `spectre link`
where a peer exists, and `flow workspace-id`; this repository's `## test` list opens with
`cd stats && make web-build`.
**Why:** The skill's Step 3 runs the whole test suite as a baseline and its Step 2 installs
dependencies, 2–3 minutes that `flow.verify` repeats at the end and that the base-branch guards make
redundant; `stats/internal/web/dist/` is gitignored and embedded, so a fresh worktree's first
`go test` fails until the SPA is built (kan-389).
**Uses:** `skills/flow/implement.md` **2. Isolate the workspace**, `skills/flow-contracts/worktree-resolution.md`,
`skills/flow-contracts/artifacts-registry.md`, `flow workspace-id`, `.flow/project.md` `## test`,
`stats/Makefile` `web-build`, `stats/internal/web/embed.go`.

### Minor-only fixes via the last implementer (lever 4)

**What:** When a round raised only Minor findings, resume the last per-task implementer with the
structured finding blocks instead of dispatching a fresh `panel-fix` subagent; record the resumption
under `task-<n>-implementer-fix-<k>`.
**Why:** The fresh subagent re-reads the bundle and the code the implementer already holds; the
conductor's reproducer re-run, diff check and mutation-proof are unchanged.
**Uses:** `skills/flow/review-panel.md` **Panel re-runs** and the fix-subagent dispatch, `SendMessage`,
`flow record dispatch`, `scripts/run-reproducer.sh`, `scripts/mutate-and-verify.sh`.

### Boundary batching (lever 6)

**What:** Gather every dispatch bundle in one Bash call at `flow.sdd-tdd` start, and run the record
end, record begin and guard for a boundary in one Bash call.
**Why:** Each boundary is 35–55 s of conductor turns over sub-second scripts.
**Uses:** `skills/flow/implement.md` **4. Execute (SDD + TDD)**, `scripts/plan-dispatch-bundles.sh`,
`scripts/gather-dispatch-context.sh`, `scripts/check-task-commit-fields.sh`, `flow record dispatch`.

### Planning stages — kickoff, brainstorm, design-approval, create-artifacts, writing-plans

**What:** The parent writes `STARTED`, dispatches one planner on `PLANNING_MODEL`, relays its
questions, and marks five stages around the planner's `## Design`, `## Artifacts` and `## Plan`
returns.
**Why:** Human-paced; 12-minute median is operator thinking time, and the two adjacent gates are the
only mechanical cost.
**Uses:** `skills/flow/brainstorm.md`, `skills/flow/brainstorm-planner.md`, `spectre new`,
`scripts/check-plan-shape.sh`, `flow stage`, `flow record dispatch`.

### Implementation stages — load-context, isolate-workspace, sdd-tdd

**What:** The conductor validates the plan, creates the worktree, and runs one implementer per
bundle — serial per worktree, in waves across throwaway worktrees once a plan declares `**After:**`.
**Why:** The second-largest wall-clock block; about 40% of it is choreography, and the isolate step
carries a redundant baseline test.
**Uses:** `skills/flow/implement.md`, `spectre validate`, `spectre list --json`,
`scripts/check-plan-shape.sh`, `scripts/plan-dispatch-bundles.sh`, `scripts/gather-dispatch-context.sh`,
`scripts/check-task-commit-fields.sh`, `flow tasks tick`.

### Review panel stage

**What:** Base-movement check, citation pre-check, bundle rebuild, diff-size and docs-only guards,
`final-review.diff`, one subagent per roster slot, findings as store rows, fix rounds with reproducer
and mutation proof, delta re-runs.
**Why:** The largest wall-clock block; one slot dominates, and each Minor finding costs a fix round.
**Uses:** `skills/flow/review-panel.md`, `scripts/resolve-base-branch.sh`, `scripts/check-base-moved.sh`,
`scripts/check-panel-citation-trigger.sh`, `scripts/check-panel-diff-size.sh`,
`scripts/check-panel-docs-only.sh`, `scripts/check-panel-reproducers.sh`, `scripts/run-reproducer.sh`,
`scripts/mutate-and-verify.sh`, `scripts/check-panel-findings-closed.sh`, `flow record finding`,
`flow record render -kind panel`.

### Verify and handoff stages — verify, visual-verify, stage-diff, run-instructions, write-in-progress

**What:** One verifier per worktree runs lint, test and spec-reach; visual verification when
configured; then three mark-only stages produce the `IN_PROGRESS` handoff.
**Why:** The verifier's 3–4 minute floor is serial after the panel; the rest is seconds.
**Uses:** `skills/flow/verify-and-handoff.md`, `scripts/prepare-workspace.sh`,
`scripts/check-workspace-isolation.sh`, `scripts/check-spec-reach.sh`, `scripts/check-visual-trigger.sh`,
`scripts/resolve-visual-screenshots.sh`, `flow record render -kind ledger`,
`flow record journal-count`, `flow record cost-status`.

### Integrate stages — preflight, unfinished-work-gate, landing-question, preserve-sessions, commit-two, landing-routes

**What:** Run 1 of the bare invocation: preflight verdict, unfinished-work guard, landing route
(configured default here), ledger render, two commits, push or merge, Jira transition.
**Why:** Seconds of machine time; the mean is dominated by push and PR waits, not pipeline work.
**Uses:** `skills/flow/integrate.md`, `skills/flow-contracts/finish-contract-run1.md`,
`scripts/check-finish-preflight.sh`, `scripts/check-unfinished-work.sh`, `scripts/check-base-moved.sh`,
`scripts/commit-split.sh`, `scripts/project-get.sh`.

### Archive stages — verify-merge, sync-archive, commit-archive, cleanup, verify-cleanup, write-finished, self-review, push-archive

**What:** Run 2: prove the merge, position the archive branch, `spectre archive`, commit, remove the
worktrees and workspace, verify the cleanup, write `FINISHED`, self-review with its prompts, push.
**Why:** Machine time is seconds; the self-review prompts and the push route are the human waits.
**Uses:** `skills/flow/archive.md`, `skills/flow-contracts/finish-contract-run2.md`,
`scripts/prepare-archive-branch.sh`, `spectre archive`, `scripts/workspace.sh`,
`scripts/check-cleanup-complete.sh`, `scripts/check-worktree-processes.sh`,
`scripts/gather-self-review-context.sh`.

### Batch the conductor's calls and replace idle polls with one `until` wait (cost lever 1)

**What:** One Bash call per task boundary (record end + guard + gather, then both `record begin`s after both launches in one message), one Bash for the panel's pre-work, one for all `record finding` calls, one for reproducers, one for all `status fixed` after the model's judgment, stage marks riding the adjacent work call; every wait on a child is one foreground `until <condition>; do sleep 5; done` (new commit for an implementer, `test -s <report>` for a reviewer, slot or verifier) instead of `echo` polls.
**Why:** kan-404's conductor made 191 API calls at $0.047 each, 29 of them idle polls; the batched shape is about 62 calls — $5.5–6.5 and 9–11 minutes per change — and no contract sentence forbids it (one to reword: "Record it per finding … not batched at the round's end").
**Uses:** `skills/flow/implement.md` §4, `skills/flow/review-panel.md` (pre-work, findings, reproducers, fix round, re-runs), `skills/flow/verify-and-handoff.md`, `flow record dispatch`, `flow record finding`, `flow record status`, `flow stage`, `scripts/gather-dispatch-context.sh`, `scripts/check-task-commit-fields.sh`, `scripts/run-reproducer.sh`.

### Close the inline-work gap in `flow.verify` (cost lever 2)

**What:** When the verifier reports a lint or test failure, the conductor dispatches a fix subagent (recorded, `-role panel-fix` or a new `verify-fix`) and re-runs the affected panel slots on the delta, or hands back — it never edits source itself and never runs the `## lint` or `## test` lists itself; a failed verifier is re-dispatched with its failure.
**Why:** kan-404's conductor fixed a lint failure inline after the panel closed — 35 turns, about $1.8, 8 minutes, no dispatch record, no slot re-run — which **Panel re-runs**' stale-result rule exists to prevent; kan-389's ran the whole test list inline between two verifiers, about 4 minutes on a 200 k context.
**Uses:** `skills/flow/verify-and-handoff.md` **Verify** and **The verifier dispatch**, `skills/flow/review-panel.md` **Panel re-runs**, `flow record dispatch`.

### Split the conductor's phase files into procedure and rationale (cost lever 3)

**What:** Move the reasoning, history and worked examples out of `implement.md`, `review-panel.md` and `verify-and-handoff.md` into `implement-rationale.md`, `review-panel-rationale.md` and `verify-and-handoff-rationale.md`, cited from the procedure file and loaded by no run — the contracts' existing `-rationale.md` pattern.
**Why:** The three files are 41 k tokens of the conductor's context, roughly a third rationale the guards and the parent already enforce; the split takes them to 26–28 k, about $1 per change and faster calls, with no generated copy.
**Uses:** `skills/flow/implement.md`, `skills/flow/review-panel.md`, `skills/flow/verify-and-handoff.md`, `skills/flow-contracts/pipeline-rationale.md` (the pattern), `scripts/check-references.sh`.

### One `## Stage` return for the handoff tail (cost lever 4)

**What:** The conductor returns one `## Stage` block for `visual-verify`, `stage-diff`, `run-instructions` and `write-in-progress` together; `flow.verify` still returns on its own; every stage keeps its own `flow stage` marks.
**Why:** Each relay costs a conductor turn, a parent turn on the session's larger context and a `SendMessage` — more than the stage it reports; keys, README table and `names_test.go` are untouched because only the visibility relay coarsens.
**Uses:** `skills/flow-contracts/pipeline.md` **Progress visibility**, `skills/flow-contracts/pipeline-rationale.md`, `skills/flow/implement.md` **Dispatch the conductor**, `skills/flow/verify-and-handoff.md`.

### Merged brainstorm gate (attention merge 1)

**What:** The convergence confirm becomes the approval: *Nothing unclear — approve the design and move on* / *Another round* / *Revise*; the parent marks `flow.brainstorm` end and `flow.design-approval` begin/end around the one relayed answer.
**Why:** `flow.design-approval`'s median is 0 s over 49 runs — the approval is a reflex seconds after the confirm; one prompt per change removed with the keys, README table and stage-exit rule unchanged.
**Uses:** `skills/flow/brainstorm-planner.md` **Convergence** and **The checklist**, `skills/flow/brainstorm.md` **The relay**, `skills/flow-contracts/pipeline.md` **Stage exit**, `skills/flow-contracts/operator-prompts.md` (shape, unchanged).

### `## self review` project key (attention merge 2a)

**What:** An optional `.flow/project.md` key with body `run` or `skip`, the single-line-literal shape of `## default landing route`; `archive.md` step 9 resolves it before the skip prompt and asks only when it is absent. This repository's body is `skip`.
**Why:** The skip prompt fires after `FINISHED`, when the operator has walked away (20 minutes on kan-404, 14.6-minute mean over 44 runs); a per-change answer stored at run 1 would need a state-file schema change (`DisallowUnknownFields`), the key needs none.
**Uses:** `skills/flow-contracts/project-configuration.md` optional-keys table, `scripts/project-get.sh`, `skills/flow/archive.md` step 9, optionally `scripts/check-model-keys.sh`.

### One filing-and-rating prompt for self-review (attention merge 2b)

**What:** The five per-angle multi-select prompts and the rating become one `AskUserQuestion` call: one multi-select listing every finding with its angle label (default *None — file nothing*) plus the rating.
**Why:** Six prompts become one with the report shape, its guard and the per-angle none-marker unchanged; the operator-prompts contract's multi-select variant already covers it, and its stale `skills/myflow-do/SKILL.md` reference gets fixed in passing.
**Uses:** `skills/flow-contracts/finish-contract-run2.md` step 9, `skills/flow/archive.md` step 9, `skills/flow-contracts/operator-prompts.md`, `scripts/check-self-review-report.sh` (unchanged).

### Move the design spec into the worktree with the change directory (attention merge 3)

**What:** `implement.md` §2's first-run copy also moves `<project>/docs/superpowers/specs/<date>-<name>-design.md` from the main checkout into the worktree and removes the staged main-checkout copy.
**Why:** The planner stages the spec in the main checkout before any worktree exists; leaving it there dirties the checkout, producing an improvised prompt at integrate (kan-404) and a `prepare-archive-branch.sh` refusal at run 2.
**Uses:** `skills/flow/implement.md` **2. Isolate the workspace**, `skills/flow/brainstorm-planner.md` **The checklist**, `scripts/commit-split.sh`, `skills/flow-contracts/git-boundaries.md`.

## Open / undesigned

Every item of the first two sessions' list is closed in section 8. Two facts that session surfaced
and did not design:

- A change that extends the `flow` CLI runs its own pipeline on the stale installed binary (kan-389's
  `-role verifier` was refused by the CLI it was adding the role to); nothing builds or prefers the
  worktree's own `stats/flow`.
- kan-395's conductor waited by ending its turn with plain text and was resumed by the completion
  notification — zero polling calls, outside the relay contract, tolerated only because the parent
  ignores a non-block return. Whether the contract should admit that form or cost lever 1's
  `until` wait stays the rule is undecided.
