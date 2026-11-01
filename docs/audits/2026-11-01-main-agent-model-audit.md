# Main-agent model audit — 2026-11-01

Whether the main (orchestrating) session of a flow run should move from Opus 5.5 to Sonnet 5.5 at
low effort (KAN-1071). Every number below sits beside the SQL or command that produced it. §1–§6
were measured on 2026-11-01 over the 30 days 2026-10-02 00:00 UTC → 2026-11-01 00:00 UTC, on a
read-only copy of the dev store built from the change branch (`scripts/live-store-copy.sh up
kan-1071-audit`, database `flow_kan_1071_audit_live`); a separate `/flow` change re-measures them
with the window ending its own date and writes §7–§8.

## Data sources

- `stage_runs.metrics.models.<model>.tokens.{main,sidechain}` — per stage run, the tokens each
  model served, split by the harvester into the main thread and the subagent sidechains
  (`stats/internal/harvest/attribute.go`). A transcript record outside every marked stage window
  goes to the change's last-begun stage run (the last-begun fallback in the same file).
- `dispatches` — one row per subagent dispatch, keyed by `dispatch_key`.
- `pricing` — per-model rates; each bucket is priced at its model's latest row.
  `claude-sonnet-5-5` is exactly half `claude-opus-5-5` on input, output, cache read and both
  cache-write rates (`SELECT * FROM pricing`).

Every spend figure below is computed from tokens, not read from `metrics.cost_usd`, so a run whose
top-level `cost_usd` is absent but whose tokens were harvested still counts (§2). The shared CTE,
`base.sql`:

```sql
WITH w AS (SELECT timestamptz '2026-10-02 00:00+00' AS f, timestamptz '2026-11-01 00:00+00' AS t),
p AS (SELECT DISTINCT ON (model) * FROM pricing ORDER BY model, effective_from DESC),
b AS (
  SELECT sr.id, sr.change_id, sr.command, sr.stage, sr.harness, m.k AS model, bk.k AS bucket, bk.v AS tok,
    (COALESCE((bk.v->>'input')::numeric,0)*p.input_per_mtok
     + COALESCE((bk.v->>'output')::numeric,0)*p.output_per_mtok
     + COALESCE((bk.v->>'cache_read')::numeric,0)*p.cache_read_per_mtok
     + CASE WHEN bk.v ? 'cache_creation_1h' OR bk.v ? 'cache_creation_5m'
         THEN COALESCE((bk.v->>'cache_creation_5m')::numeric,0)*p.cache_write_5m_per_mtok
            + COALESCE((bk.v->>'cache_creation_1h')::numeric,0)*COALESCE(p.cache_write_1h_per_mtok,p.cache_write_5m_per_mtok)
         ELSE COALESCE((bk.v->>'cache_creation')::numeric,0)*p.cache_write_5m_per_mtok END)/1e6 AS usd
  FROM w, stage_runs sr, jsonb_each(sr.metrics->'models') m(k,v), jsonb_each(m.v->'tokens') bk(k,v), p
  WHERE p.model = m.k AND sr.started_at >= w.f AND sr.started_at < w.t AND jsonb_typeof(sr.metrics->'models')='object')
```

## Coverage

Stage runs per command and harness, and those with no top-level `cost_usd`:

| Command | Harness | Runs | No `cost_usd` | Share |
|---|---|---|---|---|
| /flow | claude-code | 4860 | 2094 | 43% |
| /flow | claude | 35 | 28 | 80% |
| /flow | unknown | 8 | 5 | 63% |
| /flow-fast | claude-code | 103 | 78 | 76% |
| /flow-fast | zcode | 1096 | 922 | 84% |
| /flow-fast | muse | 71 | 71 | 100% |
| /flow-interactive | claude-code | 171 | 34 | 20% |

```sql
SELECT command, harness, count(*), count(*) FILTER (WHERE NOT (metrics ? 'cost_usd'))
FROM stage_runs WHERE started_at >= '2026-10-02' AND started_at < '2026-11-01'
  AND command IN ('/flow','/flow-fast','/flow-interactive') GROUP BY 1,2;
```

The Claude Code runs without `cost_usd` (2206), each classified by scanning its session transcript
(`~/.claude/projects/*/<session_id>.jsonl`) for a main-thread (`isSidechain` false) assistant
record inside `[started_at, ended_at)`:

| Bucket | /flow | /flow-fast | /flow-interactive | What it means |
|---|---|---|---|---|
| No main turn in window | 2002 | 54 | 34 | Nothing to attribute — nothing lost |
| No transcript on disk | 20 | 24 | 0 | Transcript gone — unrecoverable |
| Main turns in window, tokens harvested, no `cost_usd` | 69 | 0 | 0 | Tokens present, priced by this audit's CTE — counted |
| Main turns in window, nothing attributed (late auto runs) | 3 | 0 | 0 | Spend counted in a neighbouring run — see below |

```sql
SELECT id, command, session_id, started_at, ended_at FROM stage_runs
WHERE harness='claude-code' AND NOT (metrics ? 'cost_usd')
  AND started_at >= '2026-10-02' AND started_at < '2026-11-01'
  AND command IN ('/flow','/flow-fast','/flow-interactive');
-- each row's transcript scanned with a Python reader for main-thread assistant timestamps in the window
```

**The late auto runs.** Runs 18126, 18225 and 18245 are `auto_opened` `flow.sdd-tdd` runs whose
bounds equal a `task-*-implementer*` dispatch's own `started_at`/`ended_at`, opened behind a stage
run that had already ended after that instant — time the watcher had already committed past. Their
main-thread spend is not missing from the totals: the last-begun fallback placed it in the
preceding run of the same change — 18126's in 18124 (`flow.visual-verify`), 18225's in 18222
(`flow.verify`), 18245's in 18242 (`flow.visual-verify`), all three costed. Nothing is
re-attributed; this change's store fix stops new empty runs of this kind being opened.

```sql
SELECT l.id, n.id, n.stage, n.metrics ? 'cost_usd' FROM stage_runs l CROSS JOIN LATERAL (
  SELECT * FROM stage_runs o WHERE o.change_id = l.change_id AND o.id <> l.id
    AND o.started_at <= l.started_at AND o.metrics ? 'models'
  ORDER BY o.started_at DESC LIMIT 1) n WHERE l.id IN (18126, 18225, 18245);
```

**The zcode loss** (922 `/flow-fast` runs) is outside the harvester: `harvest_offsets` shows every
tracked rollout file read to its full size within seconds of its mtime, nothing in this repository
deletes files under `~/.zcode/cli/rollout`, and 607 of the 610 tracked files are gone. The deleter
is unidentified; the rows stay lost, and no harvester code changes for it.

```sql
SELECT transcript_path, byte_offset, updated_at FROM harvest_offsets WHERE transcript_path LIKE '%/.zcode/%';
-- against stat of each path; and grep -rn 'os.Remove|RemoveAll' over stats/internal/harvest scripts setup.sh Makefile
```

**Muse** (71 runs) writes no tokens by design.

## Current spend

Per command (main thread against subagents):

| Command | Main | Subagents | Total | Main share |
|---|---|---|---|---|
| /flow | $3,314 | $3,894 | $7,208 | 46% |
| /flow-interactive | $49 | $9 | $58 | 84% |
| /flow-fast | $27 | $5 | $32 | 85% |
| /flow-plan | $21 | $0 | $21 | 100% |
| **All** | **$3,411** | **$3,908** | **$7,319** | **47%** |

```sql
-- base.sql, then
SELECT command, sum(usd) FILTER (WHERE bucket='main'), sum(usd) FILTER (WHERE bucket='sidechain'), sum(usd)
FROM b GROUP BY ROLLUP(command);
```

Per change (mean / median / p90):

| Command | Changes | Main | Subagents |
|---|---|---|---|
| /flow | 132 | $25 / $19 / $51 | $29 / $13 / $87 |
| /flow-fast | 35 | $1 / $1 / $2 | $0 / $0 / $0 |
| /flow-plan | 4 | $5 / $3 / $10 | $0 / $0 / $0 |
| /flow-interactive | 3 | $16 / $16 / $23 | $3 / $1 / $6 |

```sql
-- base.sql, then
SELECT command, count(*), avg(m), percentile_cont(.5) WITHIN GROUP (ORDER BY m), percentile_cont(.9) WITHIN GROUP (ORDER BY m),
       avg(s), percentile_cont(.5) WITHIN GROUP (ORDER BY s), percentile_cont(.9) WITHIN GROUP (ORDER BY s)
FROM (SELECT command, change_id, sum(usd) FILTER (WHERE bucket='main') m,
             COALESCE(sum(usd) FILTER (WHERE bucket='sidechain'), 0) s FROM b GROUP BY 1, 2) x GROUP BY 1;
```

Main model served: `claude-opus-5-5` $3,392, `glm-5.3-flash` $19, `claude-sonnet-5-5` $0 — every
Claude Code main session ran Opus 5.5. Main effort recorded on the stage run (Claude Code): none
recorded $1,528, low $1,192, medium $641, high $13, xhigh $4.

```sql
-- base.sql, then
SELECT model, sum(usd) FROM b WHERE bucket='main' GROUP BY 1;
SELECT sr.metrics->>'effort', sum(usd) FROM b JOIN stage_runs sr USING (id)
WHERE bucket='main' AND b.harness='claude-code' GROUP BY 1;
```

Main Opus 5.5 spend by token type: cache read $2,103 (62%), cache write $709 (21%), output $580
(17%), input $0. Main cost is context size × turn count, not generation.

```sql
-- base.sql, then
SELECT sum((tok->>'input')::numeric)*4/1e6, sum((tok->>'output')::numeric)*20/1e6,
       sum((tok->>'cache_read')::numeric)*0.2/1e6, sum(usd)
FROM b WHERE bucket='main' AND model='claude-opus-5-5';  -- cache write = total minus the other three
```

## Where it goes

`/flow` main spend by stage (stages under 1% omitted; 31 stages in all):

| Stage | Main | Subagents | Share of /flow main | Contracts the main session holds there |
|---|---|---|---|---|
| flow.sdd-tdd | $1,081 | $1,350 | 33% | `implement.md` (81 KB), `implementer-brief.md` (16 KB) on top of `SKILL.md` + `pipeline.md` |
| flow.review-panel | $562 | $558 | 17% | `review-panel.md` (92 KB) |
| flow.visual-verify | $352 | $647 | 11% | `visual-verify.md` |
| flow.self-review | $309 | $792 | 9% | the self-review skill and bundle |
| flow.verify | $169 | $15 | 5% | project `## lint` / `## test` |
| flow.decide | $157 | $195 | 5% | planning skills |
| flow.brainstorm | $134 | $123 | 4% | planning skills |
| flow.document-fix | $95 | $56 | 3% | — |
| flow.write-in-progress | $95 | $10 | 3% | `pipeline.md`, state-file contract |
| flow.create-artifacts | $93 | $57 | 3% | planning skills |
| flow.writing-plans | $73 | $39 | 2% | planning skills |
| flow.landing-routes, load-context, cleanup, unfinished-work-gate | $92 | $26 | 3% | `finish-contract-run1.md` (25 KB), `finish-contract-run2.md` (15 KB) |

```sql
-- base.sql, then
SELECT stage, sum(usd) FILTER (WHERE bucket='main'), sum(usd) FILTER (WHERE bucket='sidechain')
FROM b WHERE command='/flow' GROUP BY 1 ORDER BY 2 DESC;
```

```bash
wc -c skills/flow/SKILL.md skills/flow-contracts/pipeline.md skills/flow/implement.md skills/flow/review-panel.md \
  skills/flow-contracts/finish-contract-run1.md skills/flow-contracts/finish-contract-run2.md skills/flow/implementer-brief.md
cat skills/flow/*.md skills/flow-contracts/*.md | wc -c   # 1,174,603 bytes in all
```

The two stages that orchestrate subagents (sdd-tdd, review-panel) carry half of `/flow`'s main
spend; each re-reads a 80–92 KB contract plus `pipeline.md` (49 KB) on every turn as cache read.

## Worth verdict

- **Ceiling.** At equal token counts Sonnet 5.5 halves every main-thread price component, so the
  most the switch saves is half of main Opus 5.5 spend: $3,392 / 2 = $1,696 per 30 days, 23% of
  total spend ($7,319).
- **Erosion.** If Sonnet-low needs a fraction *x* more main-thread tokens (more turns, re-reads,
  retries), the saving is $3,392 × (1 − (1 + *x*) / 2): $1,357 at *x* = 20%, $848 at 50%, nothing
  at 100%. Because 83% of main cost is cache (§3), extra turns erode the saving roughly
  linearly — a turn re-reads the whole context.
- **Quality** is unmeasured here; §7 carries the A/B. The switch is worth making only if the
  experiment shows no quality regression and a token growth well under 100%.

## Offload candidates

Jobs the main session does today, independent of the main-model verdict. Handoff cost is a
dispatch's own bootstrap: the median `flow-low` dispatch costs $0.12 on Sonnet and $0.53 on Opus.

```sql
SELECT b.v->>'model', count(*), percentile_cont(.5) WITHIN GROUP (ORDER BY (b.v->>'cost_usd')::numeric)
FROM stage_runs sr, jsonb_each(sr.metrics->'dispatches') b(k,v)
WHERE sr.started_at >= '2026-10-02' AND sr.started_at < '2026-11-01'
  AND b.v ? 'cost_usd' AND b.v->>'agent_type' = 'flow-low' GROUP BY 1;
```

| Job | Stage | Main cost | Kind | Target | Handoff cost |
|---|---|---|---|---|---|
| Run lint and tests, read the result | flow.verify | $169 | mechanical | `flow-low` verifier on Sonnet | ~$0.12 per run |
| Write the state record, stage diff, run instructions | flow.write-in-progress, flow.stage-diff, flow.run-instructions | $112 | mechanical | existing scripts (one call each) | none |
| Landing, archive commits, cleanup | flow.landing-routes, flow.commit-archive, flow.cleanup, flow.verify-cleanup, sync/push-archive | $58 | mechanical | existing scripts | none |
| Session preservation, workspace isolation, kickoff | flow.preserve-sessions, flow.isolate-workspace, flow.kickoff | $43 | mechanical | existing scripts | none |
| Task dispatch and close loop | flow.sdd-tdd | $1,081 | judgement | stays on main | — |
| Panel dispatch and finding triage | flow.review-panel | $562 | judgement | stays on main | — |
| Visual-verify orchestration | flow.visual-verify | $352 | judgement | stays on main | — |
| Self-review triage | flow.self-review | $309 | judgement | stays on main | — |
| Planning and design gates | flow.decide, flow.brainstorm, flow.create-artifacts, flow.writing-plans | $457 | judgement | stays on main | — |

The mechanical rows total $382 per 30 days (11% of main spend).

## Experiment result

*Stub — the finalize run writes this section.*

Experiment `5` (`kan-1071-main-sonnet-low`, cell `main=sonnet@low`). The finalize run fills:
`mainMetrics` before and after, including `selfReviewFindings`, per plan class; the counted and
excluded changes (at least five counted after-window changes are required before this section is
written); and the confounds — different changes in each window, and experiment #4 running over the
same window.

`selfReviewFindings` counts only graded, reviewed changes: a change with no `flow.self-review` run
yet, or with a fixed/filed self-review finding carrying no severity, is excluded from that metric
rather than counted as 0. Severity grading began 2026-10-30, so the before window's coverage is
thin; report the counted changes per window alongside the means.

Stop rule: any regression on a quality metric is a no-go for the switch.

Finalize: start a new `/flow` change that writes §7–§8 from experiment 5. This change,
`kan-1071-audit-moving-the-main-flow-agent-to-a-cheaper`, is `FINISHED` once it lands, and `/flow`
on a `FINISHED` change is a wrong-state handoff.

## Verdict and follow-ups

*Stub — the finalize run writes this section.*

The verdict on experiment `5` follows the stop rule above (any regression on a quality metric is a
no-go for the switch), read against §5's ceiling and erosion. The finalize run files one ticket per
offload candidate §6 rates mechanical and, on a positive verdict, one ticket making
`main=sonnet@low` the default. Finalize: the new `/flow` change §7 names.
