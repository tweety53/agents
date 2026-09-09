# kan-452-inject-project-hazards-into-dispatch-prompts

Implementation plan — inject project hazards into dispatch prompts automatically from the store.

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

**Goal:** Per-project hazards live in the store, and every dispatch context bundle carries the
ones matching the change's shape.

**Architecture:** A `hazards` table (migration 0019) behind store accessors, HTTP routes beside
the incidents pair, a `flow hazard`/`flow hazards` CLI in the `flow record incident` pattern, and
a `## hazards` section in `gather-dispatch-context.sh`'s bundle, filtered by a shape the caller
(the conductor) computes once and passes on every gather.

**Tech Stack:** Go (pgx, net/http, flag), PostgreSQL migration, Bash (macOS bash 3.2 compatible).

**Spec:** `spectre/changes/kan-452-inject-project-hazards-into-dispatch-prompts/design.md`
(canonical); `docs/superpowers/specs/2026-09-09-kan-452-inject-project-hazards-into-dispatch-prompts-design.md`
(the approved design record).

## Global Constraints

- The closed shape vocabulary is exactly `all` | `cross-repo` | `single-repo`, everywhere: the SQL
  CHECK, the CLI validation, the script argument, the store filter.
- Migration files are tracked by filename with no checksum — write a NEW file `0019_hazards.sql`;
  never edit an applied migration.
- No new Go dependencies. `gofmt -w` before `gofmt -l` claims clean; `go vet ./...` clean.
- Bash in `scripts/` runs under macOS bash 3.2: no associative arrays, no `declare -A`.
- The dev daemon on `127.0.0.1:4173` and the `myflow-postgres` container are protected (`.flow/project.md`
  `## stop`): an agent never stops or restarts them. Live proofs use the disposable UI-test stack
  (`make ui-test-up` / `make ui-test-down`, port 4174).
- Never `git add` `<project>/spectre/changes/` or `<project>/docs/superpowers/` in a task commit;
  never `--no-verify`; never push or merge.
- Store/API/CLI behaviour on an unreachable store: journal/fallback, one warning line, exit 0 —
  never a gate.

---

- [x] 1. Hazards table and store accessors

**Build:** green
**Files:**
- Create: `stats/internal/store/migrations/0019_hazards.sql`
- Modify: `stats/internal/records/types.go`
- Create: `stats/internal/store/hazards.go`
- Create: `stats/internal/store/hazards_test.go`
**Tests:** `TestAddListRetireHazard`, `TestListHazardsFiltersByShape`, `TestListHazardsIncludeInactive`, `TestAddHazardDuplicateRefused` — run with the selector in Step 4
**Regression:** reverting this task's commit drops the only coverage of the hazard accessors and the shape filter; the table no longer migrates, so every later task's store-touching tests fail.
**Baseline:** before=0 after=4
<!-- predicted: go test ./internal/store/ -run 'TestAddListRetireHazard|TestListHazardsFiltersByShape|TestListHazardsIncludeInactive|TestAddHazardDuplicateRefused' after task 1 -->
**Commit:** feat(store): per-project hazards table and accessors

  - [x] **Step 1: Write the failing store tests**

    Follow `stats/internal/store/records_test.go`'s use of the per-test database harness
    (`newTestDatabase` from `testsupport_test.go`; tests needing Postgres skip with a clear
    message when the compose stack is down). Four tests:

    - `TestAddListRetireHazard` — `AddHazard` returns the stored row (`Active: true`,
      `CreatedAt` set); `ListHazards(ctx, project, "all", false)` returns it newest-first;
      `RetireHazard` returns the row with `Active: false` and a later `ListHazards` omits it;
      `ListHazards(..., ..., true)` includes it again.
    - `TestListHazardsFiltersByShape` — seed three rows (`applies` = `all`, `cross-repo`,
      `single-repo`); listing with `all` returns 1, `cross-repo` returns 2 (`all` + `cross-repo`),
      `single-repo` returns 2 (`all` + `single-repo`).
    - `TestListHazardsIncludeInactive` — a retired `all` row reappears only when
      `includeInactive` is true.
    - `TestAddHazardDuplicateRefused` — a second `AddHazard` with the same `(project_key, name)`
      returns a unique-violation error; the table still holds one row.

    Use `records.Hazard` (Step 3's struct) in the tests.

  - [x] **Step 2: Run the tests to verify they fail**

Run: `cd stats && go test ./internal/store/ -run 'TestAddListRetireHazard|TestListHazardsFiltersByShape|TestListHazardsIncludeInactive|TestAddHazardDuplicateRefused' -count=1`
Expected: FAIL to compile — `records.Hazard` and the store methods undefined.

  - [x] **Step 3: Write the migration, the type and the accessors**

    `stats/internal/store/migrations/0019_hazards.sql`:

```sql verified:conventions read from 0018_guard_log.sql at plan time
-- 0019_hazards.sql: proactive, per-project warnings a dispatch bundle
-- carries before they cost time -- the record KAN-423's self-review found
-- missing, when hazards that predicted an incident lived only in the
-- operator's memory notes.
--
-- applies is a closed vocabulary matching how the pipeline classifies a
-- change (one repository vs satellites): 'all' injects into every bundle,
-- 'cross-repo'/'single-repo' only into bundles whose caller passed that
-- shape. name is a stable identifier; UNIQUE makes a re-add a clean
-- refusal, not a duplicate row. Rows are retired (active=false), never
-- deleted, the same disposition incidents get.
--
-- A NEW file, not an edit to an applied migration: 0018_guard_log.sql's
-- header carries the reason, and it still holds.

CREATE TABLE hazards (
  id          BIGSERIAL PRIMARY KEY,
  project_key TEXT NOT NULL REFERENCES projects(project_key),
  name        TEXT NOT NULL,
  body        TEXT NOT NULL,
  applies     TEXT NOT NULL CHECK (applies IN ('all','cross-repo','single-repo')),
  active      BOOLEAN NOT NULL DEFAULT true,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (project_key, name)
);

CREATE INDEX hazards_project_key ON hazards (project_key);
```

    `stats/internal/records/types.go` — beside `Incident`:

```go verified:authored in-tree for this change
// Hazard is a proactive, per-project warning a dispatch bundle carries
// before it costs time -- incidents' proactive sibling (KAN-452).
type Hazard struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Body      string    `json:"body"`
	Applies   string    `json:"applies"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
}
```

    `stats/internal/store/hazards.go` — follow `records.go`'s incident functions
    (`incidentColumns`, scan helpers) for shape:

```go verified:authored in-tree for this change
// AddHazard inserts one hazard row; a duplicate (project_key, name) comes
// back as the unique violation the caller maps to a conflict.
func (s *Store) AddHazard(ctx context.Context, projectKey string, h records.Hazard) (records.Hazard, error)

// ListHazards reads a project's hazards newest first. shape must be one of
// the closed-vocabulary values and filters to applies IN ('all', shape);
// includeInactive lifts the active filter for the operator's -all view.
func (s *Store) ListHazards(ctx context.Context, projectKey, shape string, includeInactive bool) ([]records.Hazard, error)

// RetireHazard sets active=false on the named row and returns it; an
// unknown name surfaces pgx.ErrNoRows for the handler to map to 404.
func (s *Store) RetireHazard(ctx context.Context, projectKey, name string) (records.Hazard, error)
```

    (Bodies implemented in this step; the signatures above are the contract tasks 2–4 consume.
    Validate `shape` against the closed set in `ListHazards` and return an error for anything
    else — the caller's own validation is UX, this one is the boundary.)

  - [x] **Step 4: Run the tests to verify they pass**

Run: `cd stats && go test ./internal/store/ -run 'TestAddListRetireHazard|TestListHazardsFiltersByShape|TestListHazardsIncludeInactive|TestAddHazardDuplicateRefused' -count=1`
Expected: PASS (4 tests).
<!-- predicted: the Step 4 selector run, after this plan's new tests exist -->

  - [x] **Step 5: Lint and commit**

Run: `cd stats && gofmt -w . && gofmt -l . && go vet ./...`
Expected: `gofmt -l` prints nothing; `go vet` exits 0.

```bash verified:authored in-tree for this change
git add stats/internal/store/migrations/0019_hazards.sql stats/internal/records/types.go \
  stats/internal/store/hazards.go stats/internal/store/hazards_test.go
git commit -m "feat(store): per-project hazards table and accessors"
```

- [x] 2. Hazards API routes

**Build:** green
**Files:**
- Create: `stats/internal/api/hazards.go`
- Modify: `stats/internal/api/server.go`
- Modify: `stats/internal/api/records.go` — the `RecordStore` interface gains the three hazard methods
- Create: `stats/internal/api/hazards_test.go`
- Modify: `stats/internal/store/hazards.go`, `stats/internal/store/hazards_test.go` — the `ErrHazardNotFound` sentinel the API's retire handler maps to 404 lives beside the store's other hazard sentinels, with its test flipped from `pgx.ErrNoRows`
**Allowed-collateral:** `stats/internal/**/*_test.go` — the other RecordStore fakes (web, client, reconcile, api's own struct) gain the three hazard methods the widened interface requires
**Tests:** `TestHazardRoutesAddListRetire`, `TestHazardRouteRejectsInvalidApplies`, `TestHazardRouteDuplicateConflict`, `TestHazardRouteRetireUnknownName` — run with the selector in Step 4
**Regression:** reverting drops the only HTTP-level coverage of the hazard routes; the CLI and the bundle script would then hit routes no test proves exist.
**Baseline:** before=0 after=4
<!-- predicted: go test ./internal/api/ -run 'TestHazardRoutesAddListRetire|TestHazardRouteRejectsInvalidApplies|TestHazardRouteDuplicateConflict|TestHazardRouteRetireUnknownName' after task 2 -->
**Commit:** feat(api): hazards routes beside incidents

  - [x] **Step 1: Write the failing handler tests**

    Follow `stats/internal/api/records_test.go`'s handler-test pattern (the file's own
    in-memory `fakeStore` RecordStore behind `httptest`; real-Postgres coverage lives in the
    store-layer tests task 1 carries). Coverage:

    - `TestHazardRoutesAddListRetire` — `POST /api/v1/hazards/<project>` with
      `{"name":"probe","body":"b","applies":"all"}` → 201 and the stored row echoed;
      `GET /api/v1/hazards/<project>` → 200, array of one; `GET` with `?shape=all` → same;
      `PATCH /api/v1/hazards/<project>/probe` → 200 with `active:false`; `GET` → 200, empty array.
    - `TestHazardRouteRejectsInvalidApplies` — POST with `applies:"always"` → 400; GET with
      `?shape=bogus` → 400.
    - `TestHazardRouteDuplicateConflict` — POST the same name twice → second is 409.
    - `TestHazardRouteRetireUnknownName` — PATCH an unrecorded name → 404.

  - [x] **Step 2: Run the tests to verify they fail**

Run: `cd stats && go test ./internal/api/ -run 'TestHazardRoutesAddListRetire|TestHazardRouteRejectsInvalidApplies|TestHazardRouteDuplicateConflict|TestHazardRouteRetireUnknownName' -count=1`
Expected: FAIL to compile — the handler type and routes undefined.

  - [x] **Step 3: Write the handlers and register the routes**

    `stats/internal/api/hazards.go` — a handler struct in the shape of `records.go`'s (store
    field, JSON helpers), exposing:

```go verified:authored in-tree for this change
// POST   /api/v1/hazards/{project}          -> 201, stored row
// GET    /api/v1/hazards/{project}           -> 200, []records.Hazard
//                                            ?shape=<closed-set value> filters
//                                            as the store does; absent = unfiltered
// PATCH  /api/v1/hazards/{project}/{name}    -> 200, retired row
```

    `stats/internal/api/server.go` — register directly after the incidents pair
    (`internal/api/server.go:311-312`), wiring the new handler the same way `rh`/`sth` are:

```go verified:authored in-tree for this change
mux.HandleFunc("POST /api/v1/hazards/{project}", hz.add)
mux.HandleFunc("GET /api/v1/hazards/{project}", hz.list)
mux.HandleFunc("PATCH /api/v1/hazards/{project}/{name}", hz.retire)
```

    Map `pgx.ErrNoRows` from retire to 404 and the unique violation from add to 409 — follow how
    the existing handlers map store errors.

  - [x] **Step 4: Run the tests to verify they pass**

Run: `cd stats && go test ./internal/api/ -run 'TestHazardRoutesAddListRetire|TestHazardRouteRejectsInvalidApplies|TestHazardRouteDuplicateConflict|TestHazardRouteRetireUnknownName' -count=1`
Expected: PASS (4 tests).
<!-- predicted: the Step 4 selector run, after this plan's new tests exist -->

  - [x] **Step 5: Lint and commit**

Run: `cd stats && gofmt -w . && gofmt -l . && go vet ./...`
Expected: `gofmt -l` prints nothing; `go vet` exits 0.

```bash verified:authored in-tree for this change
git add stats/internal/api/hazards.go stats/internal/api/server.go stats/internal/api/hazards_test.go
git commit -m "feat(api): hazards routes beside incidents"
```

- [x] 3. Hazards CLI and client

**Build:** green
**Files:**
- Modify: `stats/internal/client/client.go`
- Create: `stats/cmd/flow/hazard.go`
- Modify: `stats/cmd/flow/main.go`
- Create: `stats/cmd/flow/hazard_test.go`
**Tests:** `TestHazardCommandRoundTrip`, `TestHazardAddRefusesBadFlags`, `TestHazardsListFlags` — run with the selector in Step 4
**Regression:** reverting drops the only CLI-level coverage; the bundle script's `flow hazards` call and the operator's record path are then unproven at the process boundary.
**Baseline:** before=0 after=3
<!-- predicted: go test ./cmd/flow/ -run 'TestHazardCommandRoundTrip|TestHazardAddRefusesBadFlags|TestHazardsListFlags' after task 3 -->
**Commit:** feat(flow): hazard subcommands to record, list and retire

  - [x] **Step 1: Write the failing CLI tests**

    Follow `stats/cmd/flow/record_test.go`'s pattern (httptest server behind `-addr`, run the
    command function, assert exit code and stdout). Coverage:

    - `TestHazardCommandRoundTrip` — `hazard add -name probe -text b -applies all` against a
      recording stub server prints `recorded: hazard` and exits 0; `hazards` prints the stub's
      JSON array; `hazard remove -name probe` prints `retired: hazard probe`.
    - `TestHazardAddRefusesBadFlags` — missing `-name`, missing `-text`, missing `-applies`, and
      `-applies always` (not in the closed set) each exit 2 with a stderr line, contacting no
      server.
    - `TestHazardsListFlags` — `-shape cross-repo` forwards `shape=cross-repo` as a query
      parameter to the stub; `-all` forwards its include-inactive parameter; `-shape bogus`
      exits 2.

  - [x] **Step 2: Run the tests to verify they fail**

Run: `cd stats && go test ./cmd/flow/ -run 'TestHazardCommandRoundTrip|TestHazardAddRefusesBadFlags|TestHazardsListFlags' -count=1`
Expected: FAIL to compile — the command functions undefined.

  - [x] **Step 3: Write the client methods and the commands**

    `stats/internal/client/client.go` — beside `RecordIncident`/`ListIncidents`
    (`internal/client/client.go:855-880`), reuse `writeRecord`/`send` and the `incidentsURL`
    URL-building pattern:

```go verified:read internal/client/client.go RecordIncident/ListIncidents at plan time
func (c *Client) AddHazard(ctx context.Context, project string, h records.Hazard) (records.Hazard, error)
func (c *Client) ListHazards(ctx context.Context, project, shape string, includeInactive bool) ([]records.Hazard, error)
func (c *Client) RetireHazard(ctx context.Context, project, name string) (records.Hazard, error)
```

    `stats/cmd/flow/hazard.go` — `runHazardAdd`, `runHazards`, `runHazardRemove`, reusing
    `registerRecordConnFlags`, `requireRecordFlags`, `fallback.ProjectKey` and
    `classifyRecordWrite` from `cmd/flow/record.go` exactly as `runRecordIncident` does
    (`cmd/flow/record.go:1342-1391`). `-applies` is validated against the closed set before any
    network call. `flow hazards` prints the array as JSON; `-shape`/`-all` map onto the client
    call.

    `stats/cmd/flow/main.go` — dispatch `hazard` (with subcommand `add`/`remove`) and `hazards`
    alongside `record`, following the existing subcommand routing; add usage lines to the help
    text.

  - [x] **Step 4: Run the tests to verify they pass**

Run: `cd stats && go test ./cmd/flow/ -run 'TestHazardCommandRoundTrip|TestHazardAddRefusesBadFlags|TestHazardsListFlags' -count=1`
Expected: PASS (3 tests).
<!-- predicted: the Step 4 selector run, after this plan's new tests exist -->

  - [x] **Step 5: Lint and commit**

Run: `cd stats && gofmt -w . && gofmt -l . && go vet ./...`
Expected: `gofmt -l` prints nothing; `go vet` exits 0.

```bash verified:authored in-tree for this change
git add stats/internal/client/client.go stats/cmd/flow/hazard.go stats/cmd/flow/main.go \
  stats/cmd/flow/hazard_test.go
git commit -m "feat(flow): hazard subcommands to record, list and retire"
```

- [x] 4. Bundle section, shape argument, harness, live proof

**Build:** green
**Files:**
- Modify: `scripts/gather-dispatch-context.sh`
- Modify: `scripts/test-gather-dispatch-context.sh`
**Tests:** `hazards with two rows`, `hazards empty`, `hazards no flow on PATH`, `eighth argument cross-repo forwarded`, `shape omitted` — run via bash scripts/test-gather-dispatch-context.sh
**Regression:** reverting drops the only proof the bundle renders, skips and shape-filters hazards; the injection the change exists for would be untested at the script boundary.
**Baseline:** before=83 after=88
<!-- measured: bash scripts/test-gather-dispatch-context.sh 2>&1 | grep -c '^ok:' @ branch spectre/kan-452-inject-project-hazards-into-dispatch-prompts -->
<!-- predicted: bash scripts/test-gather-dispatch-context.sh 2>&1 | grep -c '^ok:' after task 4 -->
**Commit:** feat(scripts): inject project hazards into dispatch context bundles

  - [x] **Step 1: Write the failing harness cases**

    Follow the existing stub machinery in `scripts/test-gather-dispatch-context.sh`
    (`make_no_flow_dir`, `make_flow_stub_dir` — a stub `flow` answering
    `record incidents -C <dir>`). Extend the stub to also answer `hazards -C <dir>` and add five
    cases: a populated array renders a `## hazards` section after `## incidents` with one
    `- **name (applies):** body` bullet per row and appears in the census as a found source; an
    empty array skips as `hazards (none)`; PATH without `flow` skips as
    `hazards (flow unavailable)`; the stub records its argv and the case asserts
    `-shape cross-repo` was passed when the 8th argument is `cross-repo`; the 8th argument
    omitted or empty makes the stub receive `-shape all`.

  - [x] **Step 2: Run the harness to verify the new cases fail**

Run: `bash scripts/test-gather-dispatch-context.sh 2>&1 | tail -5`
Expected: FAIL — the five new cases fail (the section never renders, the stub never sees
`-shape`), the 83 existing cases still pass.

  - [x] **Step 3: Implement the section and the argument**

    `scripts/gather-dispatch-context.sh`:

    - New optional 8th positional `SHAPE="${8:-}"`; non-empty and not one of
      `all|cross-repo|single-repo` → exit 2, malformed invocation (the same contract as the
      task-ids argument).
    - After the incidents block (the script's `INCIDENTS_*` section, before `render_body`), a
      hazards block with the incidents block's three-outcome shape:
      `HAZ_SHAPE="${SHAPE:-all}"`; `flow hazards -C "$WORKTREE_REAL" -shape "$HAZ_SHAPE"`
      succeeds with a JSON array → `jq` renders one `- **<name> (<applies>):** <body>` line per
      row (pipe characters in `body` need no escaping in a bullet — plain text, unlike the
      incidents table) into `HAZARDS_BODY`, counted as a found source labelled `hazards`; `[]` →
      skipped as `hazards (none)`; `flow` absent or failing → skipped as
      `hazards (flow unavailable)`. Render inside `render_body()` as a `## hazards` section
      immediately after `## incidents`; it is part of BODY and so of the skip-when-unchanged
      hash.

```bash verified:authored in-tree for this change, mirroring the incidents block's structure
HAZ_SHAPE="${SHAPE:-all}"
if HAZARDS_JSON="$(flow hazards -C "$WORKTREE_REAL" -shape "$HAZ_SHAPE" 2>/dev/null)" \
  && [ -n "$HAZARDS_JSON" ] && [ "$HAZARDS_JSON" != "[]" ]; then
  HAZARDS_BODY="$(printf '%s' "$HAZARDS_JSON" | jq -r '.[] | "- **" + .name + " (" + .applies + "):** " + .body')"
  FOUND_LABELS+=("hazards")
  FOUND_PATHS+=("@hazards")
elif [ -n "$HAZARDS_JSON" ]; then
  SKIPPED_LABELS+=("hazards (none)")
else
  SKIPPED_LABELS+=("hazards (flow unavailable)")
fi
```

    (`@hazards` joins `@scoped-tasks`/`@incidents` in `render_body`'s case dispatch. `set -euo
    pipefail` note: the `flow hazards` call is inside `if`, so its failure is consumed — the same
    pattern the incidents block uses.)

  - [x] **Step 4: Run the harness to verify all cases pass**

Run: `bash scripts/test-gather-dispatch-context.sh 2>&1 | tail -3`
Expected: 88 `ok:` lines, `all cases passed`.

  - [x] **Step 5: Live proof against the disposable UI-test stack**

Run:
```bash verified:authored in-tree for this change
cd stats && make ui-test-up && cd ..
FLOW_ADDR=http://127.0.0.1:4174 flow hazard add -C /Users/tweety53/Projects/agents \
  -name commit-fields-guard-reverts-head-only \
  -text "Never stash, revert or reset inside a worktree a guard may read; inspect a timed-out guard call's git state before retrying it" \
  -applies all
FLOW_ADDR=http://127.0.0.1:4174 flow hazard add -C /Users/tweety53/Projects/agents \
  -name commit-fields-guard-single-repo \
  -text "check-task-commit-fields.sh's canonical-worktree argument is inert on a single-repo change: pass the change's own worktree, never a sibling's" \
  -applies single-repo
FLOW_ADDR=http://127.0.0.1:4174 flow hazards -C /Users/tweety53/Projects/agents
```
Expected: the two rows listed. Then one real gather against this change's own worktree and change
root, shape `single-repo`, asserting the section renders both rows (the `all` row and the
`single-repo` row):

```bash verified:authored in-tree for this change
WT=/Users/tweety53/Projects/agents/.worktrees/kan-452-inject-project-hazards-into-dispatch-prompts
FLOW_ADDR=http://127.0.0.1:4174 bash scripts/gather-dispatch-context.sh "$WT" \
  "$WT/spectre/changes/kan-452-inject-project-hazards-into-dispatch-prompts" \
  kan-452-inject-project-hazards-into-dispatch-prompts \
  "$PWD/skills/flow/engineering-principles.md" /tmp/k452-proof-bundle.md "" "" single-repo
grep -c "^- \*\*commit-fields-guard" /tmp/k452-proof-bundle.md   # expected: 2
grep "^## hazards" /tmp/k452-proof-bundle.md
```

    This run's bundle goes through the real CLI → real daemon (port 4174) → real migrated
    database — the migration is proven on a fresh database by the stack's own boot. Tear down:

Run: `cd stats && make ui-test-down`
Expected: stack removed; nothing on port 4174 afterwards.

  - [x] **Step 6: Commit**

```bash verified:authored in-tree for this change
git add scripts/gather-dispatch-context.sh scripts/test-gather-dispatch-context.sh
git commit -m "feat(scripts): inject project hazards into dispatch context bundles"
```

- [x] 5. Skill wiring: shape through every gather call site

**Build:** green
**Files:**
- Modify: `skills/flow/implement.md`
- Modify: `skills/flow/review-panel.md`
- Modify: `skills/flow/brainstorm.md` — restores the **TOOLS:** paragraph at the planner dispatch that the kan-474 archive revert dropped; check-dispatch-paragraphs.sh required it on the base commit and failed before this change touched anything
**Tests:** none — prose and invocation-shape changes; the guard suite covers them
**Regression:** reverting leaves the shape argument undocumented at its call sites, so every dispatch passes no shape and only always-on hazards inject — the feature silently degrades to its fail-open floor.
**Baseline:** before=0 after=0
<!-- measured: no test files touched @ branch spectre/kan-452-inject-project-hazards-into-dispatch-prompts -->
**Commit:** docs(flow): pass change shape to dispatch bundle gathers

  - [x] **Step 1: Wire implement.md**

    In **4. Execute (SDD + TDD)**'s gather invocation
    (`skills/flow/implement.md:337-340`), append the eighth argument:

```markdown verified:authored in-tree for this change
gather-dispatch-context.sh <worktree> <changeRoot> <name> <principles-path> \
  <worktree>/.superpowers/sdd/dispatch-context-bundle-<k>.md <id>[,<id>…] <canonical-worktree> <shape>
```

    Directly after that block, add:

```markdown verified:authored in-tree for this change
`<shape>` is this change's shape, computed once per run from the resolved worktree set — more than
one repository → `cross-repo`, otherwise `single-repo` — and passed on every gather this run
makes, this file's and `skills/flow/review-panel.md`'s alike. The bundle's `## hazards` section is
filtered by it; a gather made without it carries only always-on hazards.
```

    In the REQUIRED READING block beside **CONTEXT BUNDLE**
    (`skills/flow/implement.md:391-400`), add one paragraph:

```markdown verified:authored in-tree for this change
> **PROJECT HAZARDS:** the bundle's `## hazards` section carries this project's recorded
> warnings, filtered to this change's shape — each one names a way this project specifically
> bites, recorded after it cost time. They are binding: read them before your first edit and
> never argue one away without measuring.
```

  - [x] **Step 2: Wire review-panel.md**

    At the panel's bundle rebuild (`skills/flow/review-panel.md:98-102`), append the same eighth
    argument and one sentence naming it the conductor's computed shape value:

```markdown verified:authored in-tree for this change
gather-dispatch-context.sh <worktree> <changeRoot> <name> <principles-path> \
  <worktree>/.superpowers/sdd/dispatch-context.md "" <canonical-worktree> <shape>
```

  - [x] **Step 3: Run the guard suite over the touched prose**

Run: `scripts/check-vocabulary.sh && scripts/check-references.sh && scripts/check-dispatch-paragraphs.sh && scripts/check-stage-mark-calls.sh && scripts/check-contract-budget.sh`
Expected: all exit 0 (both files have ≥8 KB of budget headroom — no ratchet raise is needed).

  - [x] **Step 4: Commit**

```bash verified:authored in-tree for this change
git add skills/flow/implement.md skills/flow/review-panel.md
git commit -m "docs(flow): pass change shape to dispatch bundle gathers"
```
