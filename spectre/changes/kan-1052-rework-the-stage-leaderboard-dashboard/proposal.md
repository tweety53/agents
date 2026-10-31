# kan-1052-rework-the-stage-leaderboard-dashboard

## Why

The Stage leaderboard is read before deciding which stage of the pipeline to make cheaper next, and
today it cannot answer that. It lists the ten costliest command/stage pairs of a period with mean,
median and P90, but shows neither how much of the period's spend those ten take nor whether a stage
is growing: over October against September `flow.self-review` grew roughly eighteenfold and
`flow.visual-verify` nearly doubled while `flow.review-panel` more than halved, and the view showed
none of it. Its filters are the shared dashboard bar (a free period range, a project dropdown of
every registered checkout, a model dropdown) and a Command column dropdown, its stage cells show raw
stage ids, and its two tiles (a stage count and a run count) answer no question a reader brings.

## What changes

The view answers one question — **which stages take the most of the spend, and which are growing?**
— laid out as the approved prototype draws it:

- A chip filter bar in KAN-1048's pattern replaces the shared dashboard bar and the Command column
  dropdown: Period `7d · 30d · 90d` (a rolling window ending now, default `30d`), Project (only
  projects with spend in the window, each named by its readable label with its spend), Command
  (each command with its spend), and Clear filters when not at the defaults. The model filter is
  gone from this view.
- Three tiles replace "Stages" and "Total runs": Spend (the window's total and its change against
  the previous equal window), Top 3 stages (their combined share and a ranked list) and Biggest
  riser (the stage whose spend grew most in dollars, previous → current). Spend sits over Biggest
  riser in a narrower column beside a wider Top 3, collapsing to one column on narrow screens.
- The table ranks the top ten stages by spend and sums the rest into an "Other N stages" row. It
  keeps Command, Stage, Runs, Spend (now drawn with a bar), Median, P90 and Main share, adds Share
  and "vs previous <period>", and drops Mean cost, Main session $ and column sorting. Every stage on
  the page shows its plain-English name with the stage id in small type beneath.
- The `stage-leaderboard` endpoint stops truncating at ten rows and returns each project's spend in
  the window beside its rows; the previous window is a second call of the same endpoint.
- The page stays inside the app's own shell and theme; P95 and P99 are not added.
