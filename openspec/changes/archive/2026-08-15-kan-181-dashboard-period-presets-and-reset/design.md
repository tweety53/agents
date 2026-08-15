# Design

## The problem worth fixing

The period control is two `datetime-local` boxes. The ergonomic complaint is real but secondary; the
substantive defect is that **a wrong range and an empty store render identically**. Every view
filters on the period, so a bound that is off by hours produces a page that says "no data" and means
"not in this range". An operator reads that as "nothing recorded" and stops looking.

Everything below serves that distinction. The presets reduce how often a range is wrong; the reset
makes a wrong one recoverable in one click; the empty-state hint makes a wrong one legible.

## Presets

| Preset | From | To |
|--------|------|-----|
| Today | local midnight today | now |
| 7 days | now − 7 days | now |
| 30 days | now − 30 days | now |
| 90 days | now − 90 days | now |
| All time | `2020-01-01T00:00:00Z` | now |

**`Today` is local midnight, not "the last 24 hours".** The boxes are `datetime-local` and the
operator reasons in local time; "today" meaning "since 11pm yesterday" would be surprising.

**`All time` is a fixed floor rather than a queried minimum.** The server requires both bounds —
`from` and `to` are required RFC 3339 instants on every `/api/v1/stats/*` endpoint — so an open range
is not expressible. The alternative, querying the earliest stage run, buys a tighter bound at the
cost of a new endpoint, a loading state, and a fallback for an empty store. `2020-01-01` predates the
project, so nothing is excluded, and the label is accurate for every practical purpose.

**A preset is a computation, not stored state.** Which preset is "active" is derived by comparing the
current period against each preset's range, not remembered in a variable that could disagree with the
bounds actually in effect. That is why typing a bound clears the marking without any explicit
handler: the comparison simply stops matching.

## Reset

Shown only while the period differs from the default. A control that is always present is furniture
the eye stops seeing; one that appears when it does something is itself the signal that the period
has been changed — which is exactly the fact an operator staring at an empty page has lost track of.

## The URL hash

This is the part with a real hazard, and it is worth stating precisely.

`currentRoute()` matches the **whole** hash:

- `isViewName(hash)` for the eight static views, and
- `^run/([^/]+)/([^/]+)$` for the run-detail route.

Appending `?from=…&to=…` breaks both. `cost-per-change?from=…` is not a view name, so **every static
view silently falls back to the default view**; and `[^/]+` happily absorbs `?from=…` into the change
segment of a run route, so run detail opens on a change name that does not exist.

**So the hash is split into path and query before any route matching happens, and the matching below
that split is untouched.** Route resolution therefore cannot be affected by the presence, absence or
content of a period — which is the property worth having, because the alternative failure is silent
and total.

**An unreadable period falls back to the default.** Missing one bound, unparsable, or `to` before
`from` all resolve to the default rather than being partially applied. A URL is untrusted input: if a
crafted one could produce an empty view with no explanation, this change would have added a route to
the very failure it exists to remove.

**Writes use `replaceState`.** Adjusting a range is exploration, not navigation; one history entry
per adjustment would make the back button useless.

## Empty states

The messages already exist — `No changes updated in this period.`, `No stage runs in this period.`
They gain the reset, and the "the period may be why" framing, **only when the period is non-default**.
Under the default, empty means empty, and saying otherwise would train the operator to distrust a
true negative.

## Decisions

### The all-history preset uses a fixed floor

**ID:** all-time-fixed-floor
**Status:** active
**Chosen:** `2020-01-01T00:00:00Z` — no extra request, no loading state, correct on an empty store,
and below any record this project can hold.
**Considered:** querying the earliest stage run — a tighter and more honest bound, at the cost of an
endpoint, a loading state and an empty-store fallback. Dropping the preset — leaves the widest range
reachable only by typing, which is the ergonomics complaint restated.

### The period lives in the URL hash

**ID:** period-in-url
**Status:** active
**Chosen:** encode it in the hash — survives reload, and a range becomes shareable, which
`localStorage` cannot do.
**Considered:** no persistence — simplest, and what the dashboard does today, but loses the range on
every reload. `localStorage` — sticky but unshareable, and a stale stored range silently reproduces
the empty page this change exists to prevent, with no URL to explain it.

### Route matching happens on the path alone

**ID:** split-before-match
**Status:** active
**Chosen:** split the hash into path and query before matching, leaving the matchers unchanged —
route resolution becomes independent of the period by construction.
**Considered:** teaching each matcher to tolerate a query string — spreads the concern across every
matcher and every future one, and each is a place to get it wrong silently.

### Preset activity is derived

**ID:** derived-active-preset
**Status:** active
**Chosen:** compare the current period against each preset's range — no stored "active preset" to
fall out of step with the bounds in effect.
**Considered:** storing the selected preset — creates a second source of truth for a value already
implied by the bounds, and needs an explicit rule for what happens when a bound is typed.

### The empty-state hint is conditional on a non-default period

**ID:** hint-only-off-default
**Status:** active
**Chosen:** show it only when the period differs from the default — under the default, empty is a
true negative and saying "the period may be why" would be misleading.
**Considered:** showing it whenever a view is empty — simpler, but trains the operator to discount a
genuine empty result.

## Open questions

None. Every question this design raised was put to the operator and answered.

## Testing

- The default period is `now − 30 days` → `now` — the assertion that pins what nothing currently
  asserts.
- Each preset produces its stated bounds; `Today` starts at local midnight; `All time` carries the
  floor.
- Active-preset derivation: a preset's range marks it active; a typed range marks none.
- Reset appears off-default, is absent on the default, and restores the default.
- Hash: a period-carrying URL selects the same view as one without; a missing, unparsable or reversed
  period falls back to the default and still renders; a run-detail route survives a period.
- Empty state: the hint and reset appear on an empty view under a non-default period, and do not
  under the default.

`cd stats/web && npm test` and `npx tsc -b` must pass.
