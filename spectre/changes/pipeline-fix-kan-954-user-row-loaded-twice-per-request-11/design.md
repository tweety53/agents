## Context

A self-review finding left "not traced" reached a FINISHED run's final report as "Still open",
beside an unverified store-write claim. Blast radius, counted per **Pipeline defects found
mid-run** (`skills/flow-contracts/pipeline.md`): the files touched are `landselfreviewreport.go`,
its test, `scripts/land-self-review-report.sh`, `skills/flow-self-review/SKILL.md`,
`skills/flow/cleanup.md`, `skills/flow/integrate.md` and `skills/flow-fast/SKILL.md`; with every
file citing `land-self-review-report`, "Record every finding in the flow store" or `## Finished`,
the union is 36 files, under 60: not `big`. <!-- measured: sort -u of the touched files plus git grep -lF -- "<name>" 752f3a28 per name -->

Measured at 752f3a28: the gymie report `docs/self-review/kan-954-user-row-loaded-twice-per-request-self-review.md`
carries the line `- **[flow-stats-app]** … cause in flowd recording, not traced — noted`. Gymie
declares no `check-self-review-report.sh` (its `## lint` is Gradle only), and that guard run over
gymie's corpus reports 102 violations across 60 reports — so the report guard never reaches the
project the incident came from. `land-self-review-report` does: every project's report lands
through it.
<!-- measured: scripts/check-self-review-report.sh /Users/tweety53/Projects/gymie/docs/self-review at 752f3a28, exit 1 -->

## Decisions

### Gate at the landing chain

**ID:** gate-at-landing-chain
**Status:** active
**Chosen:** `land-self-review-report` refuses a report with an untraced finding line, exit 4,
checked right after the start-of-run branch assert — the one chain every project's report passes
through, in `/flow`'s integrate, `/flow-fast`'s verify and standalone `/flow-self-review`, before
FINISHED can be written.
**Considered:** a new check in `check-self-review-report.sh` — it runs only in projects that
declare it, and gymie does not (its corpus fails 102 ways), so it would not have caught this
incident; a check in `integrate-change.sh land` — covers `/flow` alone, and its fixtures commit
stub reports; running the report guard over a temp dir holding the one new report — gymie's
latest report already fails its shape (`flow-speed` carries `- None.`), widening this fix into a
shape migration.

### Refuse "not traced" and "untraced" only

**ID:** untraced-phrases
**Status:** active
**Chosen:** a finding line (`- **[`-prefixed) containing `not traced` or `untraced`,
case-insensitive.
**Considered:** `unverified` — legitimate in finding lines that describe a plan's
`unverified:` provenance tag (gymie's kan-680 report, lines 20 and 23); `still open` —
legitimate in agents' kan-560 report line 25 ("while flow.review-panel was still open"); either
would refuse true reports. The prose rule (step 2) names all three; the gate pins the incident's
own shape.

### Untraced findings go to the planner, no new pairs-table row

**ID:** trace-via-planner
**Status:** active
**Chosen:** an untraced finding is sent to step 3's planner as a non-`big` finding whose plan
opens on the trace (the existing reproduce-first task step of `skills/flow/brainstorm-planner.md`),
which ends it fixed or disproved — both existing outcomes of step 3.
**Considered:** a new `trace` key suffix on the pairs table — a new row `check-loop-dispatch.sh`
would have to learn, for what the planner already does.

### Store rows re-read, not assumed

**ID:** store-rows-reread
**Status:** active
**Chosen:** step 6 states a row as recorded only once `flow self-review findings -change <name>`
lists it; a missing row is recorded again, and a store that still refuses is fixed as a pipeline
defect.
**Considered:** a mechanical check in the store CLI — a timeout that did write is invisible to
the client; the listing is the existing read that answers it.

### Finished carries no open item

**ID:** finished-no-open-items
**Status:** active
**Chosen:** one paragraph under the Finished handoff in `skills/flow/cleanup.md`, citing step 4 of
`/flow-self-review` for the `big` filing route.
**Considered:** a new handoff field such as `**Open:** none` — a field whose only legal value is
`none` is ceremony.

### First review on medium

**ID:** first-review-effort
**Status:** active
**Chosen:** the first review, `pipeline-fix-11-review-1`, runs on `flow-medium` (effort `medium`)
on `opus`: the change is normative prose across four skill files plus a new guard exit code whose
wording must stay consistent with the rest of the self-review contract.
**Considered:** `low` — the Go change is small and its test pinned, but the prose edits are the
substance and benefit from a careful read.

### Two fixer groups

**ID:** fixer-grouping
**Status:** active
**Chosen:** two groups in series — Task 1 in `pipeline-fix-11-fix-g1`, Task 2 in
`pipeline-fix-11-fix-g2`; Task 2's prose names Task 1's refusal line.
**Considered:** one group — fine too, but the Go task runs the test suite and the prose task runs
the verbatim-moves guard; separate dispatches keep each small.

## Open questions

## Live check

No runtime state changes: `land-self-review-report` is a one-shot process per call. After landing,
copy the gymie kan-954 report into a scratch git repo and run
`<agents repo>/scripts/land-self-review-report.sh <repo> main 'x' docs/self-review/<file>`: it
exits 4 with one `LAND-UNTRACED-FINDING: …:33: - **[flow-stats-app]** …` line and HEAD unmoved.
Failure looks like exit 0 and a new commit.
