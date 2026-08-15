## Why

Every one of the dashboard's eight views is gated behind a period filter that is two free-text
`datetime-local` boxes (`stats/web/src/components/PeriodPicker.tsx`). Changing the range means
typing both bounds. There is no way to say "the last week", no way to say "everything", and no way
back to the default once it has been changed.

**The cost is not the typing — it is that a wrong range and no data look identical.** Every view
returns rows only within the period, so a bound that is slightly off renders an empty page
indistinguishable from a store with nothing in it. Nothing on screen says the period is why.

Observed on 2026-08-15: the live state board rendered `Changes: 0` and *"No changes updated in this
period"* while the database held the change and all seven other views rendered it. That instance was
caused by a record stamped outside the period rather than a mistyped bound, but it presented exactly
as a mistyped bound would — which is the point.

**The default period is already correct and is not being changed.** `defaultPeriod()` in `App.tsx`
already returns `now − 30 days` → `now`. Nothing asserts it, so this change pins it with a test.

## What Changes

- **Presets** on the period control — `Today` (local midnight → now), `7 days`, `30 days`,
  `90 days`, `All time`. `All time` resolves to a fixed floor rather than an open range, because the
  server requires both bounds: `from` and `to` are required RFC 3339 instants on every
  `/api/v1/stats/*` endpoint.
- **A reset** to the default, shown **only** when the period differs from it, so it is a signal
  rather than permanent furniture.
- **The active preset is marked**, so the current range is legible without reading two timestamps.
- **The manual boxes stay.** Presets cover the common cases; an arbitrary range must remain
  expressible, and typing clears the active-preset marking.
- **The period is carried in the URL hash**, so a range survives reload and can be shared.
- **The empty states say when the period may be why**, and offer the reset — the item that addresses
  the actual failure rather than only the ergonomics.

## The routing hazard this change has to clear

`currentRoute()` matches the **entire** hash: `isViewName(hash)` for the eight static views, and
`^run/([^/]+)/([^/]+)$` for the run-detail route. Appending a query string breaks both —
`cost-per-change?from=…` is not a view name, so **every static view would silently fall back to the
default view**, and the run-route regex would swallow the query into the change segment.

The hash is therefore split into path and query **before** any route matching, and the matching
below that split is unchanged. A malformed, partial, reversed or non-numeric period in the URL falls
back to the default rather than rendering an empty page: a URL must not be able to reintroduce the
exact failure this change exists to remove.

## Capabilities

### Added Capabilities

- `myflow-dashboard-period`: how the dashboard's period is chosen, defaulted, carried in the URL,
  and explained when a view comes back empty.

## Impact

**Code** — `stats/web/src/components/PeriodPicker.tsx`, `stats/web/src/App.tsx`, the components
owning the empty states, and their tests.

**Not changing** — the server's period contract (`from` and `to` required, half-open `[from, to)`),
any `/api/v1/*` endpoint, the default period itself, or the eight views' aggregation.
