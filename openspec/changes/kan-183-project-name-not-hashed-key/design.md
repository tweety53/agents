# Design — a project has a name; the key is not it

Adapted from `docs/superpowers/specs/2026-08-15-kan-183-project-name-not-hashed-key-design.md`, which
carries the full walk-through.

## The shape of it

`projectLabel(key)` strips a trailing `-[0-9a-f]{8}` and returns the key unchanged otherwise. The
four display surfaces use it; the full key stays the row identity, the run-detail link target, and a
`title` tooltip. The server resolves a display name to a key in the two places a `project` parameter
is parsed, leaving all eight view queries' `project_key = $3` untouched.

## Decisions

### How the display name is derived

**ID:** strip-the-documented-suffix
**Status:** active
**Chosen:** Strip a trailing `-[0-9a-f]{8}` and return the key unchanged when it does not match. The
suffix is documented in **State file** (`skills/myflow-contracts/state-file.md`), so matching it
exactly is matching a stated contract.
**Considered:**
- *Split on the last `-` and drop the tail* — ruled out: it mangles any key not derived this way, and
  a project literally named `my-repo` with a differently-shaped key would render as `my`.
- *Have the server send a display name alongside the key* — ruled out: it widens every DTO carrying a
  project for a value the client can derive from the key it already has, and the derivation is a
  documented one, not a judgment.

### Where a display name is resolved back to a key

**ID:** resolve-on-the-server
**Status:** active
**Chosen:** In the API layer, at the two places a `project` parameter is parsed, against the
`projects` table.
**Considered:**
- *A client-side map built from the rows on screen* — ruled out: `ProjectFilter.tsx`'s own header
  records that the client must not re-derive what the server restricts by, and the map would be wrong
  for a project with no rows in the current period — present in the table, absent from the screen.
- *A new `GET /api/v1/projects` endpoint and a dropdown* — ruled out as unrequested: tolerance of
  both forms needs no endpoint, and the filter's free-text shape is a recorded decision.
- *Changing the eight view queries to match either form* — ruled out: eight copies of one rule, in
  SQL, where one resolution above the store does it once.

### What an ambiguous display name does

**ID:** ambiguous-is-an-error
**Status:** active
**Chosen:** 400, naming the ambiguity and listing the candidate keys.
**Considered:**
- *Pick the first match* — ruled out: it answers a different question than the one asked, with
  nothing on screen to say so.
- *Filter by every matching key* — ruled out: it changes the store's predicate from `=` to `= ANY`
  across all eight queries to serve a case that is rare by construction, and "rows from two projects
  under one label" is itself a confusing answer.

### What happens to the hash on screen

**ID:** hash-moves-to-the-tooltip
**Status:** active
**Chosen:** Keep the full key as the row identity, the link target and a `title` tooltip.
**Considered:** *Drop the hash from the UI entirely* — ruled out: it is exactly what distinguishes
two same-named checkouts, and the operator was offered that option and chose to keep it reachable.

## Open questions

*(none — the two questions this change raised, how far the short name should reach and how it should
land, were both answered before the design was written.)*
