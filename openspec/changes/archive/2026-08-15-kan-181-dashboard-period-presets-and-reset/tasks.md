> **Execution:** `/myflow-do` implements this plan. Mark each checkbox when its task passes spec +
> quality review.

**Goal:** Make the dashboard's period easy to change, easy to undo, and — when a view comes back
empty — make it obvious whether the period is why.

## Global Constraints

- **The default period does not change.** `now − 30 days` → `now` is already correct; this change
  pins it with a test and builds everything else around it.
- **The server's period contract does not change.** `from` and `to` stay required RFC 3339 instants
  on every `/api/v1/stats/*` endpoint, and `[from, to)` stays half-open. No backend file is touched.
- **Route resolution must not depend on the period.** Adding or removing a period from a URL must
  select the same view either way.
- **No task edits `openspec/` or `docs/superpowers/`.**

## Baseline

**Measured 2026-08-15 against `0256da3`:** 112 SPA tests across 7 files
(`cd stats/web && npm test`). `npx tsc -b` clean.
<!-- measured: npm test run on 2026-08-15 in stats/web, against commit 0256da3 -->

---

### 1 Split the hash into path and query before route matching

**Build:** green

**Files:**
- Modify: `stats/web/src/App.tsx`
- Modify: `stats/web/src/App.test.tsx` *(create if absent)*

**Interfaces:**
- Consumes: `currentRoute()`, `parseRunRoute()` and `isViewName()` as they stand.
- Produces: a hash split into path and query, with the existing matchers applied to the path alone.

**This task carries the whole risk of the change and is deliberately first, alone, and behind
tests.** `currentRoute()` matches the entire hash today. Appending a query string makes
`cost-per-change?from=…` fail `isViewName`, so **every static view silently falls back to the default
view**; and `^run/([^/]+)/([^/]+)$` absorbs `?from=…` into the change segment, so run detail opens on
a change that does not exist. Both failures are silent.

- [x] **Step 1: Failing tests for route resolution under a query string**

Write these first and watch them fail: `#/cost-per-change?from=X&to=Y` resolves to the
`cost-per-change` view; `#/run/proj/change?from=X&to=Y` resolves to that run route with the change
name **not** carrying the query; a hash with no query resolves exactly as it does today for both a
view and a run route.

- [x] **Step 2: The split**

Split the hash on its first `?` into path and query. Apply `parseRunRoute` and `isViewName` to the
path only. Do not teach either matcher about query strings — the point is that they never see one.

- [x] **Step 3: Say why the split exists**

A comment at the split recording that route matching must not depend on the period, and what the two
silent failures were. The next person to touch this file needs the hazard visible where it lives.

**Tests:** `App.test.tsx` — view route with and without a query, run route with and without a query.

**Regression:** Reverting this sends every static view to the default whenever a period is present in
the URL, and points run detail at a nonexistent change.

**Baseline:** before=112 after=116 SPA tests.
<!-- predicted: four new route-resolution cases, none written yet -->

**Commit:** `feat(1): match the route on the hash path alone`

---

### 2 Read and write the period in the URL

**Build:** green

**Files:**
- Modify: `stats/web/src/App.tsx`
- Modify: `stats/web/src/App.test.tsx`

**Interfaces:**
- Consumes: task 1's path/query split, and `defaultPeriod()`.
- Produces: a period read from the query on load, and written back on change.

- [x] **Step 1: Failing tests for the fallback, written before the happy path**

The invalid cases are the point, so write them first: a query with only `from`; only `to`; an
unparsable bound; `to` earlier than `from`; no query at all. **Every one falls back to the default
period and still renders the view.** Then the happy path: a valid pair is applied as given.

- [x] **Step 2: Parse, validate, fall back**

Read both bounds from the query. Apply them only when both are present, both parse, and `from` is
strictly before `to`. Anything else is the default — never a partial application. A URL is untrusted
input, and a crafted one must not be able to produce the empty page this change exists to remove.

- [x] **Step 3: Write with `replaceState`**

Changing the period updates the URL without adding a history entry. Adjusting a range is exploration,
not navigation; one entry per adjustment makes the back button useless. Ensure the write does not
feed back into the hash listener as a route change.

**Tests:** `App.test.tsx` — the five invalid cases above, the valid case, and that a period write
does not push history.

**Regression:** Reverting this loses the period on every reload and makes a range unshareable.

**Baseline:** before=116 after=123 SPA tests.
<!-- predicted: six fallback/parse cases plus one history assertion, none written yet -->

**Commit:** `feat(2): carry the dashboard period in the URL, falling back to the default`

---

### 3 Presets and reset on the period control

**Build:** green

**Files:**
- Modify: `stats/web/src/components/PeriodPicker.tsx`
- Modify: `stats/web/src/components/DashboardBar.test.tsx`
- Modify: `stats/web/src/App.tsx`

**Interfaces:**
- Consumes: `Period`, `PeriodPickerProps.onChange`, and `defaultPeriod()`.
- Produces: preset buttons, derived active-preset marking, and a conditional reset.

**`defaultPeriod()` moves into `PeriodPicker.tsx` and is exported, which is why `App.tsx` is in the
list.** The 30-day preset and the reset both need the default, and `App.tsx` needs it too; leaving it
private to `App.tsx` would put the 30-day formula in two files. The global constraint here is that
the default does not change — two copies of it is precisely how it would change in one place and not
the other. `App.tsx`'s edit is the import, nothing more.

- [x] **Step 1: Failing tests for the presets and the derived marking**

Each preset sets its stated bounds: `Today` from local midnight; `7 days`, `30 days`, `90 days` from
now minus that span; `All time` from the `2020-01-01T00:00:00Z` floor. A period equal to a preset's
range marks that preset active; a hand-typed range marks none.

- [x] **Step 2: The presets**

Render the five buttons. **Derive which is active by comparing the current period against each
preset's range** — do not store a selected preset. A stored one is a second source of truth that can
disagree with the bounds actually in effect, and deriving it is why typing a bound clears the marking
with no explicit handler.

- [x] **Step 3: The reset, present only off-default**

A control returning to the default, rendered **only** while the period differs from it. Always-on
furniture stops being seen; a control that appears when it does something is itself the signal that
the period has been changed.

- [x] **Step 4: Keep manual entry working**

The `From`/`To` boxes keep their current behaviour, including rejecting an unparsable value rather
than propagating `NaN`.

**Tests:** `DashboardBar.test.tsx` — five preset cases, active-marking derivation both ways, reset
present off-default, reset absent on the default, reset restores the default, manual entry unchanged.

**Regression:** Reverting this returns the control to two free-text boxes with no way back to the
default.

**Baseline:** before=123 after=133 SPA tests.
<!-- predicted: ten preset/marking/reset cases, none written yet -->

**Review found the derivation compared against a moving clock.** `defaultPeriod()` and each
preset's `range()` recompute `now` per call, while `period.to` is a snapshot — so any re-render after
a few seconds unmarked a just-selected preset and made the reset appear on an untouched default.
`sameRange` is now a single exported tolerance comparison (±60s), shared with `DataTable`, and the
regression tests advance the clock and re-render rather than asserting immediately.
<!-- measured: both failures reproduced live during review on 2026-08-15, then re-run green -->

**Commit:** `feat(3): add period presets and a reset to the dashboard bar`

---

### 4 An empty view says whether the period is why

**Build:** green

**Files:**
- Modify: `stats/web/src/components/DataTable.tsx`
- Modify: `stats/web/src/components/DataTable.test.tsx`
- Modify: `stats/web/src/viewTypes.ts`
- Modify: `stats/web/src/App.tsx`

**Allowed-collateral:** `stats/web/src/views/*.tsx`

**Interfaces:**
- Consumes: the existing `emptyMessage` prop and the period-versus-default comparison from task 3.
- Produces: an empty state that names the period as a possible cause, and offers the reset.

**The reset needs a callback that actually reaches `App.tsx`'s period state, which is why this task
reaches past `DataTable`.** No view threads one today. Without it the button would render in
`DataTable`'s unit tests and be dead in the running application — a control that looks like it works
and does not, which is worse than not shipping it. So `ViewProps` gains an optional
`onPeriodChange`, `App.tsx` passes `setPeriod` to the active view, and each of the eight views
forwards `period`/`onPeriodChange` to its `DataTable`. The eight view files are mechanical
forwarding and are declared as allowed collateral rather than listed one by one.

- [x] **Step 1: Failing tests for both directions**

Empty under a non-default period: the message says the period may be why and the reset is offered.
Empty under the default period: neither appears. **The second case matters as much as the first** —
under the default, empty is a true negative, and blaming the period would train the operator to
distrust it.

- [x] **Step 2: The conditional hint**

Extend the existing empty states rather than adding a new surface. Read the current empty messages
first (`No changes updated in this period.`, `No stage runs in this period.`) and keep their wording
where it is already right.

**Tests:** `DataTable.test.tsx` — empty off-default shows hint and reset; empty on default shows
neither.

**Regression:** Reverting this restores the ambiguity between "no data" and "wrong range", which is
the defect this change exists to fix.

**Baseline:** before=133 after=137 SPA tests.
<!-- predicted: four empty-state cases, none written yet -->

**Commit:** `feat(4): tell an empty view apart from a mistyped period`

---

### 5 Pin the default period

**Build:** green

**Files:**
- Modify: `stats/web/src/App.test.tsx`

**Interfaces:**
- Consumes: `defaultPeriod()`.
- Produces: the assertion that nothing currently makes.

- [x] **Step 1: Assert the default**

`defaultPeriod()` returns `now − 30 days` → `now`. Everything else in this change is built around
that value and nothing asserts it, so it can drift with no test failing. Use a fixed clock rather
than comparing against a second `new Date()`.

**Tests:** `App.test.tsx` — the default period's span and end.

**Regression:** Reverting this lets the default drift silently.

**Baseline:** before=137 after=138 SPA tests.
<!-- predicted: one assertion, not yet written -->

**Commit:** `test(5): pin the dashboard's default period`
