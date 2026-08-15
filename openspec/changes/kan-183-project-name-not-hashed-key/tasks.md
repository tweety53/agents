> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** The dashboard names projects `agents` and `gymie`, and the filter accepts what the screen
shows — without the client ever reconstructing what the server restricts by.

## Global Constraints

- **The project key does not change**, nor how it is derived, nor the state file, nor the `projects`
  table's shape. This change is about what is *shown* and what a `project` parameter *accepts*.
- **The eight view queries keep `project_key = $3`.** Resolution sits above the store; no query in
  `stats/internal/store/aggregate.go` is edited.
- **No response body changes.** No DTO gains a display-name field — the client derives it from the
  key it already has.
- **No task edits `openspec/` or `docs/superpowers/`.**

## Baseline

**Measured 2026-08-15 against `c6f6dab`:** `cd stats/web && npm test` → 8 test files, 141 tests, all
passing. `cd stats && go test ./internal/api/ -count=1` → ok, 21.0s.
<!-- measured: both commands run on 2026-08-15 from the main checkout at c6f6dab -->

---

### 1 Derive a display name from a project key

**Build:** green

**Files:**
- Create: `stats/web/src/lib/projectLabel.ts`
- Create: `stats/web/src/lib/projectLabel.test.ts`

**Interfaces:**
- Consumes: nothing — a pure function over a string.
- Produces: `projectLabel(key: string): string`, imported by task 2's four surfaces.

The whole risk of this task is in the fallback, so the tests carry it: a key that does not end in the
documented suffix must come back **unchanged**, because the helper's input is whatever the server put
in `projectKey` and it must not lose a segment to a pattern that was never about it.

- [x] **Step 1: Failing tests, written first**

Cover, at minimum: `agents-a740d89c` → `agents`; `gymie-7c1f238a` → `gymie`; a key with no suffix
(`agents`) unchanged; a key whose tail is hex but the wrong length (`repo-a740d89` — seven, and
`repo-a740d89c1` — nine) unchanged; a tail of the right length that is not hexadecimal
(`repo-zzzzzzzz`) unchanged; an uppercase tail (`repo-A740D89C`) unchanged, since the derivation
produces lowercase hex; a basename that itself contains hyphens (`my-repo-a740d89c` → `my-repo`); and
the empty string, which comes back empty rather than throwing.

- [x] **Step 2: The helper**

One exported function, one regular expression anchored at the end, a doc comment stating where the
suffix comes from — **State file** (`skills/myflow-contracts/state-file.md`) — and why a
non-matching key is returned untouched. No options object, no second export.

**Tests:** `stats/web/src/lib/projectLabel.test.ts` — the cases in step 1.

**Regression:** Reverting this task removes the only definition of a display name; task 2's surfaces
would have nothing to call.

**Baseline:** before=141 after=150 SPA tests.
<!-- predicted: nine cases enumerated in step 1; none written yet -->

**Commit:** `feat(1): derive a project's display name from its key`

---

### 2 Name projects on screen, keep the key reachable

**Build:** green

**Files:**
- Modify: `stats/web/src/views/StateBoard.tsx`
- Modify: `stats/web/src/views/CostPerChange.tsx`
- Modify: `stats/web/src/components/ChangeVariable.tsx`
- Modify: `stats/web/src/views/RunDetail.tsx`
- Modify: `stats/web/src/views/views.test.tsx`
- Modify: `stats/web/src/views/RunDetail.test.tsx`

**Allowed-collateral:** `stats/web/src/**/*.test.tsx`

**Interfaces:**
- Consumes: `projectLabel` from task 1.
- Produces: four surfaces reading a display name, each keeping the full key reachable.

- [x] **Step 1: Failing tests, written first**

Assert that the state board's and cost-per-change's `Project` cell renders `agents` for a row keyed
`agents-a740d89c`, that the full key is still reachable from that cell (its `title`), that the change
dropdown's option label reads `agents/kan-1`, and that the run-detail header reads `agents`. Assert
too that the row's link target still carries the **full key**, since that is the route's identity.

- [x] **Step 2: The four surfaces**

Both `Project` columns take a `render` that shows `projectLabel(r.projectKey)` with the full key as
`title`; their `accessor` also returns the display name, so sorting, searching and the exact-match
filter dropdown all agree with what is on screen. `ChangeVariable`'s option label and `RunDetail`'s
header call the helper directly. `rowKey`, `runDetailHash` and every link target keep the full key —
do not touch them.

**Tests:** `stats/web/src/views/views.test.tsx` and `stats/web/src/views/RunDetail.test.tsx` — the
assertions in step 1.

**Regression:** Reverting this task puts the hashed key back in every column, which is the reported
defect.

**Baseline:** before=150 after=156 SPA tests.
<!-- predicted: six assertions across the four surfaces; none written yet -->

**Commit:** `feat(2): name projects on screen and keep the key one hover away`

---

### 3 Accept a display name wherever a project key is accepted

**Build:** green

**Files:**
- Modify: `stats/internal/api/stats.go`
- Modify: `stats/internal/api/changes.go`
- Modify: `stats/internal/api/server.go`
- Modify: `stats/internal/store/changes.go`
- Modify: `stats/internal/api/stats_test.go`
- Modify: `stats/internal/store/changes_test.go`

**Allowed-collateral:** `stats/internal/api/*_test.go`, `stats/internal/client/client_test.go`,
`stats/internal/web/embed_test.go`

`server.go` and the two test fakes are named because the change to them is **forced, not chosen**: a
store method the handlers call has to be declared on the interface they depend on, which lives in
`server.go`, and every fake implementing that interface must gain the method or its package stops
compiling. Declaring them here is the difference between a sweep the plan anticipated and one nobody
checked.

**Interfaces:**
- Consumes: the `projects` table (`stats/internal/store/migrations/0001_init.sql`).
- Produces: a store lookup resolving a display name to project keys, and its use at the two places a
  `project` parameter is parsed — `parsePeriodAndProject` (`stats/internal/api/stats.go`) and the
  changes list's own filter parse (`stats/internal/api/changes.go`).

- [x] **Step 1: Failing tests, written first**

Store: a lookup that returns both keys for a display name shared by two projects, one key for a
unique display name, and none for an unknown name. API: a stats view requested with a display name
returns that project's rows; with an exact key, resolution is not attempted and the key is used
as-is; with a display name shared by two projects, the response is **400** and its message names both
candidate keys; with an unknown value, the request succeeds and returns no rows.

- [x] **Step 2: The store lookup**

One method on `*Store` returning the project keys whose display name equals the given value. Derive
the display name in SQL by trimming the documented suffix — the same shape task 1's helper matches,
anchored, lowercase hex, exactly eight — so the two sides agree by construction rather than by
comment. Say in a doc comment that this is the server-side twin of `projectLabel`.

- [x] **Step 3: Resolution at the two parse sites**

A value that is already an exact project key is used unchanged and **costs no query** — check that
first. Otherwise resolve: one match wins, several is a 400 naming the ambiguity and listing the
candidate keys, none passes through unchanged. Both call sites share one helper; neither reimplements
the rule.

**Tests:** `stats/internal/api/stats_test.go` and `stats/internal/store/changes_test.go` — the cases
in step 1.

**Regression:** Reverting this task makes the filter demand the full key again while the screen shows
the short one, which is the trap this change exists to avoid.

**Baseline:** before=141 after=141 SPA tests — this task adds Go tests only.
<!-- predicted: the task edits no SPA file, so the SPA count cannot move -->

**Commit:** `feat(3): accept a project's display name wherever its key is accepted`
