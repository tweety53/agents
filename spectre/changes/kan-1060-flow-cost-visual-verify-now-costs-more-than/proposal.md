# kan-1060-flow-cost-visual-verify-now-costs-more-than

## Why

`flow.visual-verify` costs about as much as implementation, and the main session carries a large share
of it (KAN-1060, with KAN-1063 merged into it). Checked against the flow stats store on 2026-10-30
(`stage_runs.metrics`, priced at that day's `pricing` rows), the finding reproduces and the change is
not withdrawn:

- Weekly stage cost, last two weeks: visual-verify $254 / $246 against sdd-tdd $319 / $222 —
  visual-verify now equals or exceeds implementation.
- Main session's share of visual-verify: 34% over 30 days ($332 of $984); 38% since the 2026-10-28
  stage-close fix ($39.45 of $105.05 over 45 stage runs; $0.88 of main per run; 10.4 main tool calls
  per run at ~284k context).
- Nothing on the base thins the parent's visual-verify turns: KAN-858 only slimmed prompt text, and
  the one-shot merger (`983be8de`) covers the review panel alone.

<!-- measured: stage_runs.metrics cost_usd and metrics.dispatches.*.cost_usd summed over the flow store, priced at the 2026-10-30 pricing rows @ live dev store read 2026-10-30, not re-runnable at a ref -->

### Measured breakdown of visual-verify cost

30 days to 2026-10-30, opus-5-5 prices — the Done-when's "measured breakdown of visual-verify cost per
change".

| bucket | n | $ | share |
|---|---|---|---|
| verifier passes (opus/low) | 273 | 338 | 34% |
| main session (parent's own turns) | 237 stage runs | 332 | 34% |
| fix-loop dispatches inside the stage window (implementer, panel-fix, reviewers, merger) | 152 | 139 | 14% |
| sidechain not matched to a dispatch row (rows with no harvested tokens, almost all Oct 1–16) | — | ~175 | 18% |

<!-- measured: flow store dispatches and stage_runs for flow.visual-verify, 30 days to 2026-10-30 @ live dev store read 2026-10-30, not re-runnable at a ref -->

- **Per change.** The costliest: kan-1016 $95 of visual-verify (12 stage runs, 27 verifier passes, 11
  fix rounds) against $36 of sdd-tdd; kan-1003 $43 against $20; kan-1044 $40 against $10. Fix rounds
  per change: 0:12, 1:10, 2:7, 3:10, 4:6, 5:3, 6–11: 8 changes.
- **Verifier passes by kind** (14 days, $306 over 173 passes, $1.57 a pass): scoped fix re-runs 35%,
  `-full` runs 25%, first passes 18%, non-zero-exit re-dispatches 13%, incomplete-report remainders
  (`-rest`) 8%.
- **Inside one pass:** cache creation 51%, cache read 31%, output 17%; 15–78 turns, ~19 minutes, the
  longest single wait a polled `verify` suite of 2–18 minutes. Screenshots are not the sink: ~8 PNGs a <!-- measured: verifier subagent transcripts of the last 14 days, per-pass usage and turn timings @ machine-local transcripts read 2026-10-30, not re-runnable at a ref -->
  pass at ~2k tokens each, ~12% of a pass's cache creation. Turn count drives the cost. Over 189
  verifier transcripts of the last 14 days a pass averages ~35 tool turns: 4–5 are the resolve,
  identical-check, compose and measure chain (12–15%), ~6 are one image read each (15–18%), ~2 go on
  small scripts the verifier writes to summarise the measuring JSON.
- **Inside the main session's visual-verify windows** (26 recent stage runs, transcript-classified):
  pre-dispatch mechanics 16%, report reading and guards 12.5%, stage marks / ledger / steps 12–13 11%,
  verifier launches 9%, loading `visual-verify.md` 4%, polling 2%, PNG reads 0.2% (one read in 26
  windows); the rest is fix-loop and incidental work. Each main message re-reads a 130k–460k context,
  so each costs ~$0.05–0.09. Verifier reports handed back into main run 2k–43k characters and stay in
  its context for the rest of the session.
- **Already fixed on the base, so no task here:** the fix loop's implementers, reviewers and panel
  rounds used to be billed to `flow.visual-verify`, because its stage mark stayed open and was later
  `superseded` ($530 of the $984 sat in superseded runs). `e37af3f4` (2026-10-27) closes it `stopped`;
  no visual-verify run has closed `superseded` since 2026-10-28.
- **Already enforced, so no task here:** round counts come from `In-run fix <k>` task titles, never
  from `dispatches.outcome`. Before 2026-10-28 some verifier rows whose report opened a round were
  closed `completed` (kan-1000 fix-2/3/4, kan-1016 fix-3/4/5, kan-930), against
  `skills/flow/verify-fix-loop.md` step 1, in force since `6ab7ef9b` (2026-10-17). Every round-opening
  row checked since 2026-10-28 reads `blocked` (kan-1041, kan-1044, kan-1047, kan-1059, kan-1061,
  kan-1067).

<!-- measured: verifier subagent transcripts and main-session transcripts of the last 14 to 30 days, classified on 2026-10-30 @ machine-local transcripts read 2026-10-30, not re-runnable at a ref -->

### The ticket's starting questions, answered

- Dispatches and rounds per change: above. A closing `-full` run is already once per loop, after a
  clean scoped re-run (**One full run closes the loop**, `skills/flow/verify-fix-loop.md`).
- Image tokens against cached prompt reads: images are ~12% of a pass's creation; turns and the
  context they re-read dominate.
- Opus for capture as well as judging: the verifier stays on opus (Done-when). Its mechanical steps
  are scripts already; this change collapses the measuring chain into one call. What a script cannot
  take over is spec authoring and extension, seeding, interactive states, motions and merge-base
  attribution.
- Skipping frames byte-identical to the last approved run: rejected on evidence (decision
  `no-skip-identical-frames`).
- KAN-858: Done; it slimmed prompt text only, and no remaining lever here is prompt text.

## What changes

- **A thin parent.** One script call before each verifier dispatch (`visual-verify-pass.sh prepare`)
  and one after it (`visual-verify-pass.sh close`) run the mechanical steps the main session runs by
  hand today — the stage marks, the trigger, the pre-flight, the probe and fingerprint, the verifier's
  resolved prompt bundle and its record, the report guards, the record end, the captures commit and
  the ledger — so the main session reads one short verdict per pass.
- **The verifier stops re-running `verify`.** `flow.verify` already ran the same visual specs green;
  the verifier's step 7 retires, and the green verify row is pinned to the HEAD it ran at.
- **Main-session columns on the Stage leaderboard.** **Main session** and **Main share** after
  **Total cost**, exactly as the approved prototype draws them — the board the Done-when measures on,
  against the baseline of decision `done-when-baseline`.
- **One call for the measuring chain** (`visual-verify-pass.sh measure`): the verifier's resolve,
  identical-check, compose and band/seam/matrix measuring become one call writing one manifest, read
  with every image it names in one turn.
- **Every flagged element explained; one report, one round.** The matrix measurement lists every
  over-tolerance element; the report explains each by a closed list of causes; every in-run fix task
  names the report it transcribes, and two tasks never transcribe one report.
- **The by-hand frame/matrix check becomes automatic** in `check-verify-report.sh`; the main session
  reads no images.

Expected saving per 30 days, overlapping — fewer rounds also means fewer passes for the other parts to
save on: ~$100–160 from the thin parent, ~$55–90 from dropping the verifier's `verify`, ~$30–60 from
the one measuring call, ~$100–160 plus 30–70 minutes per avoided round from explaining every flagged
element.
<!-- predicted: the stats board's Main session column and flow.visual-verify's stage cost over the 30 days after landing -->
