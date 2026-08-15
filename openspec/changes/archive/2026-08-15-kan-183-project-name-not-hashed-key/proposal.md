## Why

Every place the dashboard names a project, it prints the raw project key:

```text verified:read off the running dashboard at http://127.0.0.1:4173 on 2026-08-15, and returned by the state-board API for the same two projects
agents-a740d89c
gymie-7c1f238a
```

A project key is `<basename of the main checkout>-<first 8 hex of sha1 of that checkout's absolute
path>` (**State file**, `skills/myflow-contracts/state-file.md`). The hash exists so two same-named
repositories in different directories address different records. It is an **identity**, and identity
is not a label: nothing a human reads is improved by eight hex digits, and the column is wide enough
to wrap because of them.

Five surfaces name a project. Four of them are display:

- the `Project` column of the live state board (`stats/web/src/views/StateBoard.tsx`),
- the `Project` column of cost-per-change (`stats/web/src/views/CostPerChange.tsx`),
- the change dropdown's `<projectKey>/<name>` option label
  (`stats/web/src/components/ChangeVariable.tsx`),
- the run-detail header (`stats/web/src/views/RunDetail.tsx`).

**The fifth is not.** The `Project` filter box sends its text to the server, which matches the key
**exactly** — `AND ($3::text IS NULL OR project_key = $3)`, in every one of
`stats/internal/store/aggregate.go`'s view queries. Shortening only the display would leave the
filter demanding `agents-a740d89c` while the screen everywhere says `agents`: typing what you see
would return an empty view, which is worse than the wordy column it replaced.

## What Changes

- **A display name is derived from a key, in one client helper.** `projectLabel(key)` strips a
  trailing `-[0-9a-f]{8}` and returns the key **unchanged** when it does not match that shape — a key
  derived some other way stays legible rather than being silently mangled. The four display surfaces
  above use it.
- **The full key stays reachable.** It remains the row identity, the run-detail link target, and a
  `title` tooltip on the project cell — so where two checkouts share a basename, the thing that tells
  them apart is one hover away rather than gone.
- **The server accepts a display name where it accepts a key.** A `project` parameter that is not an
  exact key is resolved against the `projects` table by display name. Exactly one match resolves to
  that key. **Several is a 400** naming the ambiguity and listing the candidate keys, because a
  filter whose value means two projects is not a filter and silently picking one would be a wrong
  answer rather than a slow one. None passes through unchanged, exactly as today — an unknown
  project already yields no rows.

**The resolution is server-side on purpose.** `ProjectFilter.tsx`'s own header records why the client
must not re-derive what the server restricts by, and the `projects` table already holds the mapping,
so the client would be reconstructing something the server can answer.

## Capabilities

### Modified Capabilities

- `myflow-stats-views`: how a project is named on screen, and what the `project` parameter accepts.

## Impact

**Code** — `stats/web/src/` (a new helper plus the four display surfaces and their tests) and
`stats/internal/api/` (`parsePeriodAndProject`, the changes endpoint's own parse, and a resolution
query against the `projects` table).

**Not changing** — the project key itself, how it is derived, the state file, the `projects` table's
shape, the eight view queries' `project_key = $3` predicate, the run-detail route's shape, or any
`/api/v1/*` response body.
