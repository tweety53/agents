# KAN-183 — a project has a name; the key is not it

**Date:** 2026-08-15
**Change:** `kan-183-project-name-not-hashed-key`

## The observation

The dashboard's Project column reads `agents-a740d89c` and `gymie-7c1f238a`. The eight hex digits are
the first 8 of the sha1 of the checkout's absolute path, and they are there so two repositories
called `agents` in different directories address different records. That is a good reason for the
*key* to carry them, and no reason at all for the *column* to.

## What names a project today

| Surface | File | Kind |
|---------|------|------|
| Live state board's `Project` column | `stats/web/src/views/StateBoard.tsx` | display |
| Cost-per-change's `Project` column | `stats/web/src/views/CostPerChange.tsx` | display |
| Change dropdown's option label | `stats/web/src/components/ChangeVariable.tsx` | display |
| Run-detail header | `stats/web/src/views/RunDetail.tsx` | display |
| `Project` filter box | `stats/web/src/components/ProjectFilter.tsx` | **round-trips to the server** |

The fifth is what makes this more than a formatting change. Every view query in
`stats/internal/store/aggregate.go` filters with `AND ($3::text IS NULL OR project_key = $3)` — an
exact match. A display-only change would leave the screen saying `agents` and the filter demanding
`agents-a740d89c`, so typing what you see would empty the view. That is a worse outcome than the
wordy column, and it is the trap this design exists to avoid.

## The design

### 1. One client helper for the display name

```ts
projectLabel(key: string): string
```

Strips a trailing `-[0-9a-f]{8}`; returns the key **unchanged** when it does not match that shape.

The fallback is the important half. The helper's input is whatever the server put in `projectKey`,
and its own contract cannot be "assume the derivation" — a key that was derived differently, or a
future key shape, must stay readable rather than lose its last segment to a pattern that was never
about it. Stripping only what provably matches the documented suffix is the difference between a
formatter and a guess.

### 2. The full key stays reachable

It remains the row identity (`rowKey`), the run-detail link target, and a `title` on the project
cell. Where two checkouts share a basename — the exact case the hash exists for — the screen shows
`agents` twice and the hover tells them apart. Nothing that disambiguates is destroyed; it is moved
off the default reading path, which is what a label is for.

### 3. The server accepts a display name

`parsePeriodAndProject` (`stats/internal/api/stats.go`) is the single place every stats view's
`project` parameter is parsed; the changes endpoint has its own, equally single, parse. A value that
is not an exact key is resolved against the `projects` table by display name:

| Matches | Result |
|---------|--------|
| exactly one key | resolve to that key |
| more than one key | **400**, naming the ambiguity and listing the candidate keys |
| none | pass through unchanged — an unknown project yields no rows, exactly as today |

**The ambiguous case is an error rather than a choice.** A filter whose value means two projects is
not a filter, and picking one silently would answer a different question than the one asked, with
nothing on screen to say so. Listing the keys makes the full key — the thing that disambiguates —
the obvious next thing to type.

**Why the server and not the client.** `ProjectFilter.tsx`'s own header records the rule: the client
does not re-derive what the server restricts by. The `projects` table already holds every key, so a
client-side map would be reconstructing, from whatever rows happened to be on screen, an answer the
server can give exactly. It would also be wrong in the one case that matters — a project with no rows
in the current period is absent from the client's view and present in the table.

**Why not a dropdown.** `ProjectFilter` is free text because no endpoint lists distinct projects, and
that is recorded in its own header. Making the filter tolerant of both forms needs no such endpoint,
so none is added.

### 4. The eight view queries do not change

Resolution happens above the store, so `project_key = $3` stays exactly as it is in all eight
queries. One resolution query is added; none of the existing ones is touched.

## Verification

- SPA unit tests: `projectLabel` (the documented suffix, a key without one, a key whose tail merely
  looks hex-ish but is the wrong length, an empty string), both `Project` columns rendering the short
  name, the change dropdown's label, and the run-detail header.
- Go tests: resolution of an exact key, of a display name matching one project, of a display name
  matching two (400, both keys named), and of an unknown value (passed through).
- The full lint list, `go test ./... -race -count=1`, and the SPA suite.
- The dashboard itself: the Project column reads `agents` and `gymie`, and typing `agents` into the
  filter returns that project's rows.
