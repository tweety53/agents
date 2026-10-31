# kan-1051-rework-the-trend-dashboard

## Why

The Trend over time view claims to answer whether the pipeline is getting cheaper per change as
flow changes, and nothing on it does. It shows daily total spend and stage-run counts, which move
with how many changes ran that day, not with what one change costs, and daily and weekly spend is
already Weekly costs' job. Flow health shows one period-wide mean cost per finished change with no
trend. A reader deciding whether flow's own changes are paying off has no view that plots what a
finished change costs week by week.

The per-change figures such a view would plot are also wrong today. Harvest attributes a
transcript record to a stage run only when its timestamp falls inside one of its session's stage
windows, and drops every other record while consuming its bytes. `/flow-fast` marks its stages in
bursts after the work, so its windows last milliseconds and about 85 % of a run's spend falls
between them; every other command loses some share too (roughly `/flow` 11 %, `/flow-plan` 13 %,
`/flow-interactive` 16 %, `/flow-fast` 41 % over the last 45 days' finished changes with
transcripts on disk). The ticket's scope therefore takes in the cost-capture fix as well as the
view — the operator's answer 7 of design round 1 chose to fix it in this change rather than file a
follow-up — because plotting the undercounted figures would answer the question wrongly.

## What changes

The view answers one question — **"Is a finished change getting cheaper as flow changes?"** — as
the approved prototype draws it:

- A chip filter bar in KAN-1048's pattern replaces the shared dashboard bar and the model filter:
  Period `8w · 12w · 26w` (default `12w`), Command (one chip per command that finished a change in
  the window with its count, exactly one always selected, default `/flow`), Project (All plus one
  chip per project with finished changes of the selected command, readable label with count), and
  Clear filters when not at the defaults.
- Three tiles: Median per change (the window's latest half against its first half), Changes
  finished (with what was spent on them), and Flow changes landed (changes finished in the
  `agents` project, every command).
- A weekly chart of median and mean cost per finished change, with flow changes landed as bars
  beneath, a hover guide and tip; and a By week table, newest first.
- The Days and Total cost tiles, the Daily cost chart and the By-day table are gone; spend over
  time stays on Weekly costs.
- The `trend` endpoint returns one row per finished change in the window — its finish week,
  command, project and whole-life cost — and refuses a `model` parameter.
- Harvest attributes a record that falls between two stage windows to the stage that began last
  before it, up to the change's finish, for every command; `flowd` back-fills, once at startup,
  the spend the old rule dropped from transcripts still on disk.
- In-run fix 1 (verify): a finished change's finish bound is the later of its state write and its last stage run — the live check showed `/flow-fast` changes write FINISHED at kickoff, so the old bound dropped their whole gap spend — and `scripts/live-store-copy.sh` can point its scratch flowd at the real transcripts.
- In-run fix 2 (visual-verify): the Trend chart's guide line and tip are committed inside the pointer event, so they appear and disappear in the same frame as the hover band.
