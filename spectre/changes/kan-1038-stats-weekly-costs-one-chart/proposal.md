# kan-1038-stats-weekly-costs-one-chart

## Why

The Weekly costs dashboard (KAN-1037) is hard to read: an 8-week table beside a small picker-driven chart with unlabelled axes, a filter bar that does not help, and no week-over-week comparison. Separately, planning that dashboard showed a clickable prototype converges a UI design faster than prose, and `/flow` never offers one.

## What changes

- Weekly costs is one large chart of the last 8 Wednesday-to-Wednesday weeks; the "Last 8 weeks" table and the picker chart are gone.
- Each of the 5 series (Overall, Planning, Implementation, Review, Self-review) toggles on/off from a pill row; the Y scale follows the visible series.
- Every point carries a permanent week-over-week label — red ▲ for an increase, green ▼ for a decrease, whole %, non-overlapping — and a hover tooltip with series, week, cost and the change vs the previous week.
- X axis (week start, Wed UTC) and Y axis (Cost, USD) are labelled with tick values.
- The filter bar (period, project, model) is not rendered on Weekly costs; the dashboard always covers every project and model.
- `/flow` planning builds and publishes a clickable visual prototype whenever the change alters something a user sees, and iterates it with the operator before design approval.
