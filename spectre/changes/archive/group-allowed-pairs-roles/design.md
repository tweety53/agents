## Context

The roles table is `## Allowed pairs` in `skills/flow-contracts/model-policy.md`. It is parsed by
`stats/internal/modelpolicy` — the one parser the guards (`check-dispatch-pair`,
`check-loop-dispatch`, `plan-agent-pair`, `plan-class`, `pipeline-fix`) and flowd share — and
flowd serves it, embedded, on `GET /api/v1/roles`, which the Experiments view
(`stats/web/src/views/Experiments.tsx`) draws row for row. The operator asked for the roles grouped
in the table and in what flowd serves (planning being `explorer` + `plan-designer` + `planner`),
a plain-English line per role, a more compact table, and every review-panel bundle as a role row
of its own.

The table stays the one source: group and description are parsed from it and served, never
stored in a database (derived, not stored). Experiment 4 `planning-opus-medium-to-xhigh`, already
open outside this change, covers the three planning roles as one combined experiment; nothing
here touches experiment cells.

**The visual spec** is the approved prototype
`/Users/tweety53/Projects/agents-worktrees/group-allowed-pairs-roles/.superpowers/sdd/roles-prototype.html`,
published at https://claude.ai/artifact/D82LpNhBssL7oEaewaeu5Y. It is the spec the implementation
and `flow.visual-verify` compare against: the roles panel, its group header rows, its row hover
and its `#id` link focus. Its sample override cells are experiment 4's, as flowd serves them on
`GET /api/v1/experiments`. The approved design record is
`/Users/tweety53/Projects/agents-worktrees/group-allowed-pairs-roles/.superpowers/sdd/2026-10-30-group-allowed-pairs-roles-design.md`.

**Groups and roles:**

| Group | Roles |
|---|---|
| Session | `main` (API and view only — no table row) |
| Planning | `explorer`, `plan-designer`, `planner` |
| Implementation | `implementer`, `panel-fix` |
| Review | `gated-reviewer`, `gated-re-reviewer`, `panel-reviewer`, `panel-sanity-efficiency`, `panel-mutation-failure`, `panel-rerun`, `late-fix`, `panel-merger` |
| Visual verification | `visual-verifier`, `tooling-analyst` |
| Self-review | `self-review-pass`, `self-review-plan`, `self-review-fix`, `self-review-review-1`, `self-review-review-r` |
| Pipeline fix | `pipeline-fix-plan`, `pipeline-fix-fix`, `pipeline-fix-review-1`, `pipeline-fix-review-r` |

Each role's `Does` line is the prototype's text; task 1 of `tasks.md` carries the table verbatim.

**Not in scope:** grouped experiment cells (`-cell 'planning=…'`) — nothing asks for a
group-level override; any change to which pairs a role admits; splitting the fix-round re-run
bundle.

## Decisions

### Groups live as headings over one table each

**ID:** groups-in-markdown-headings
**Status:** active
**Chosen:** one `### <Group>` subheading plus one table per group under `## Allowed pairs`, header `| Role | Allowed pairs | Requires marker | Does |`; the parser reads every table and takes each row's group from the nearest preceding `###` heading — the markdown is grouped for its reader too (round 1, Q1 A).
**Considered:** a `Group` column on the one table — the markdown stays a flat list and grouping only appears in the view; a separate `role → group, does` table below `## Allowed pairs` — a second parse joining two tables that can drift apart.

### Groups follow the pipeline area

**ID:** groups-by-pipeline-area
**Status:** active
**Chosen:** Session, Planning, Implementation, Review, Visual verification, Self-review, Pipeline fix (round 1, Q2 A).
**Considered:** groups by job across loops (every planner under Planning, every fixer under Fixing) — the self-review and pipeline-fix loops, each a unit an experiment is opened on, disappear; Self-review and Pipeline fix merged into one "Fix loops" group — two loops with distinct triggers read as one.

### One compact row per role

**ID:** compact-one-row-per-role
**Status:** active
**Chosen:** as the prototype — one row per role, its marked rows folded into the pairs cell as a muted second line `<pair> with <marker>`, columns Role, Does, Allowed pairs, the `Requires marker` and `Experiment` columns gone, an override's `#<id>` link after its pairs (round 1, Q3 A).
**Considered:** twin loop roles sharing one row — two roles with separate pairs and experiments behind one line; pair lists shortened to `opus · low, medium` notation — the drawn pair no longer reads as the string a dispatch is checked against.

### Group and description are parsed, never stored

**ID:** group-and-does-from-table
**Status:** active
**Chosen:** `modelpolicy.Entry` gains `Group` and `Does`, parsed from the embedded table; `GET /api/v1/roles` serves them — derived from the one source.
**Considered:** a flowd table or column holding each role's group — a second copy of what the markdown already states, kept in sync by hand.

### `main` is grouped in the API, not the table

**ID:** main-group-literal
**Status:** active
**Chosen:** the synthetic leading `main` entry carries group `Session` and does `The session you talk to: runs the pipeline and sends out every other role.`, both literals in `stats/internal/api/experiments.go` beside the `main` literal it already carries.
**Considered:** a `### Session` table holding a `main` row — `Parse` would key `main` as a dispatch role every guard then admits, which no dispatch is.

### Rows keep their current order inside a group

**ID:** row-order-kept
**Status:** active
**Chosen:** inside each group, rows keep the relative order they have today (an escalation row stays where it is), so every reader's per-role row order and `TestRolesRoute`'s table-order assertion hold; the view orders roles by first appearance.
**Considered:** moving each escalation row directly under its role — reorders rows across roles for a markdown reader only, and rewrites a test that pins table order.

### Every review-panel bundle with a fixed composition is its own role

**ID:** split-panel-extra
**Status:** active
**Chosen:** `panel-extra` is replaced by `panel-sanity-efficiency` (`test-sanity+call-efficiency`) and `panel-mutation-failure` (`mutation+failure-modes`), each on `panel-extra`'s pairs; `panel-reviewer` keeps the `primary+principles` bundle (round 3, Q1 A). No open or closed experiment names `panel-extra` (flowd `GET /api/v1/experiments`, 2026-10-30), so no stored cell is orphaned.
**Considered:** `panel-extra-sanity` / `panel-extra-mutation` — longer ids in a nowrap Role column for no gain in meaning; renaming `panel-reviewer` to `panel-primary-principles` too — touches the hook tests, the experiment stats and every `Role: panel-reviewer` in the skills for a bundle that already has its own row.

### The re-run bundle stays one row

**ID:** one-panel-rerun-row
**Status:** active
**Chosen:** `panel-rerun` stays one row: a re-run carries whichever roles raised findings, so no fixed bundle exists to give a row of its own (round 3, Q2 A).
**Considered:** one re-run role per pass-1 bundle — splits a mixed re-run into separate dispatches the panel does not make today.

### Extra-bundle stats select pass 1 alone

**ID:** extra-stats-pass-one
**Status:** active
**Chosen:** in `stats/internal/store/experimentstats.go`, `panel-sanity-efficiency` selects `^panel-0-test-sanity\+call-efficiency`, `panel-mutation-failure` selects `^panel-0-mutation\+failure-modes`, `panel-reviewer` (`^panel-0-`) excludes both keys, and `panel-rerun` is `^panel-[1-9][0-9]*-` with no exclusion. Extra dispatches run at pass 1 alone, so a later-round key opening `test-sanity+call-efficiency` is a re-run; this fixes the two misattributions the shared row hid.
**Considered:** keeping `^panel-[0-9]+-` for the extras — the re-run dispatches whose key opens `test-sanity+call-efficiency` stay counted as an extra bundle, not as `panel-rerun`.

### The roles panel drops its footnote

**ID:** roles-footnote-dropped
**Status:** active
**Chosen:** the `p.sub` footnote under the roles table goes: the approved prototype draws the panel as its title and the table alone.
**Considered:** keeping it — a departure from the approved prototype, which is the source of truth.

### The roles panel keeps the view's panel chrome

**ID:** roles-panel-chrome-unchanged
**Status:** active
**Chosen:** the table's rules come from the prototype, scoped `.experiments .roles`; the panel's own box and title stay the view's shared `.experiments .panel` and `.experiments .ttl`, so it matches the Experiments and Detail panels beside it. `flow.visual-verify` reads a difference in the panel box's padding or radius against this decision.
**Considered:** restyling the roles panel's box to the prototype's standalone-page `.panel` — splits one panel's chrome from its neighbours on the same view, which nothing asked for.

### The UI-test seed carries experiment 4's cells

**ID:** seed-matches-experiment-4
**Status:** active
**Chosen:** `stats/cmd/uitest-seed/seed.go`'s open experiment "planning on xhigh" carries experiment 4's three cells — `explorer=sonnet/Explore@low`, `plan-designer=opus/Plan@medium,opus/Plan@high,opus/Plan@xhigh`, `planner=opus/flow-planner-medium,opus/flow-planner-high,opus/flow-planner-xhigh` — so the capture draws the prototype's override text; the `#<id>` the link reads is the seed's own id, data rather than design.
**Considered:** leaving the seed's two cells — the capture departs from the prototype on its data, which a later fix round would then have to repair.

### The roles captures join the view's spec

**ID:** roles-captures-in-experiments-spec
**Status:** active
**Chosen:** `stats/web/tests/visual/experiments.spec.ts` gains element captures of the `Current roles` region — resting, a role row hovered, the `#<id>` link focused; its `.mockups` sidecar is unchanged, since this project declares no `mockups` root and `flow.visual-verify` renders the cited prototype as the frame (step 10, `skills/flow/visual-verify-verifier.md`).
**Considered:** a spec file of its own — a second spec for one view.

## Open questions

## Recommended approach

Parse the grouped tables in `stats/internal/modelpolicy` (every table carrying the header, each
row's group from the nearest `###` heading, the scan ending at the next `##` heading), and rewrite
`## Allowed pairs` into one table per group with a `Does` column in the same commit, the
`panel-extra` row replaced by the two bundle rows and `plan-class.sh` reading them — the table, its
parser and its one row-keyed reader moving together keep every commit green. Name the new roles
in the dispatch-site table and the model-and-effort rules, retarget the experiment-stats
selectors, serve `group` and `does` on `GET /api/v1/roles`, then redraw the view's roles table
to the prototype and capture it. Rejected: a `Group` column or a joined second table
(groups-in-markdown-headings), stored groups (group-and-does-from-table), a `main` table row
(main-group-literal), longer or renamed bundle ids (split-panel-extra), split re-runs
(one-panel-rerun-row).

## Critical files

- `skills/flow-contracts/model-policy.md` — `## Allowed pairs`, regrouped with the `Does` column and the two bundle rows.
- `stats/internal/modelpolicy/modelpolicy.go` — the shared parser: every grouped table, `Entry.Group`, `Entry.Does`.
- `stats/internal/modelpolicy/modelpolicy_test.go` — the parse cases and the real-table role list.
- `stats/internal/guard/check_dispatch_pair_test.go` — the malformed-row fixture's header.
- `stats/internal/guard/planclass.go` — each extra bundle's pair off its own row.
- `stats/internal/guard/plan_class_test.go` — the extra-pair tests over the two rows.
- `skills/flow/implement.md` — the dispatch-site table's panel-bundle Role line cell.
- `skills/flow/brainstorm-planner.md` — **Model and effort**'s extra-dispatches bullet.
- `stats/internal/store/experimentstats.go` — `roleMetrics` and `roleDispatchSQL`.
- `stats/internal/store/experimentstats_test.go` — the per-bundle attribution case.
- `stats/internal/api/experiments.go` — `roleEntry.Group`, `roleEntry.Does`, the `main` literals.
- `stats/internal/api/experiments_test.go` — `TestRolesRoute`.
- `stats/web/src/api.ts` — `RoleEntry.group`, `RoleEntry.does`.
- `stats/web/src/views/Experiments.tsx` — `RolesTable`.
- `stats/web/src/styles.css` — the compact roles-table rules.
- `stats/web/src/views/views.test.tsx` — the `Experiments` roles case.
- `stats/cmd/uitest-seed/seed.go` — the open experiment's cells.
- `stats/web/tests/visual/experiments.spec.ts` — the roles captures; its snapshots and the full-app suite's.
- Unchanged, generic over the table: `hooks/enforce-agent-baseline.py`, `scripts/check-dispatch-pair.sh`, `stats/internal/guard/modelpolicy.go`, `stats/internal/repotest/role_lines_test.go`, `stats/internal/api/records.go`.

## Existing code to reuse

- `modelpolicy.Rows` / `ticks` (`stats/internal/modelpolicy/modelpolicy.go`) — extended, not replaced; `Parse` keeps its map over `Rows`.
- `policyTable` (`stats/internal/guard/modelpolicy.go`) — `plan-class` already reads the overlaid table through it.
- `pcExtra` (`stats/internal/guard/planclass.go`) — unchanged; only the pairs it is handed change.
- `experimentsHandler.roles` (`stats/internal/api/experiments.go`) — gains two fields per entry.
- `tablePairs` (`stats/web/src/views/Experiments.tsx`) — the Detail table keeps reading the unmarked row through it.
- `newStatsFixture`, `fx.dispatch`, `fx.finding`, `wantMeans` (`stats/internal/store/experimentstats_test.go`) — the attribution case's fixture.
- `seedExperiments` (`stats/cmd/uitest-seed/seed.go`), `open` and `PINNED_QUERY` (`stats/web/tests/visual/experiments.spec.ts`, `./support`) — the capture's stack and navigation.
- `scripts/rebuild-full-app-suite-zip.sh`, `scripts/compose-mockup-frames.sh`, `scripts/measure-visual-properties.sh` — the capture task's rebuild and measurements.

## Verification

- `cd stats && make -s web-dist && go test -race -count=1 ./internal/modelpolicy ./internal/guard ./internal/api ./internal/store ./cmd/uitest-seed` passes, and `scripts/run-verify-list.sh <worktree> lint` exits clean.
- `cd stats && make -s web-dist && cd web && npx vitest run src/views/views.test.tsx -t 'Experiments'` passes.
- The project's `verify` (`npm run test:visual`) passes on the recaptured baselines, and the roles captures compose against the prototype rendered at the same state with every leftover departure named.
- `scripts/plan-class.sh spectre/changes/group-allowed-pairs-roles/tasks.md 1` prints both `extra:` lines.
- The live check below.

## Live check

Against a copy of the real store, never the dev stack: `scripts/live-store-copy.sh up
group-allowed-pairs-roles` builds flowd from this worktree on its own port against a copy of the
dev `flow` database, and prints the `FLOW_ADDR` export that points `flow` at it.

1. `curl -s http://127.0.0.1:4175/api/v1/roles | jq -c '[.[] | {role, group, does}]'` — every
   entry carries a non-empty `group`; groups appear in the order Session, Planning,
   Implementation, Review, Visual verification, Self-review, Pipeline fix; every unmarked entry
   carries a non-empty `does` and every marked entry an empty one; `panel-extra` is absent and
   `panel-sanity-efficiency` and `panel-mutation-failure` are present. Before (the dev flowd on
   4173, the old build): no `group` or `does` key, and a `panel-extra` entry. Not working looks
   like: a missing `group` key, an empty `group`, or `panel-extra` still served.
2. With `FLOW_ADDR` exported to the copy: `flow experiment open -name live-check -cell
   'panel-mutation-failure=opus/flow-low' -cell 'panel-sanity-efficiency=opus/flow-low'`, then
   `curl -s http://127.0.0.1:4175/api/v1/experiments/<id>/stats | jq '.metrics[] |
   select(.name=="findingsPerDispatch")'`. The copy's reviewer dispatches hold pass-1
   `mutation+failure-modes` keys, pass-1 `test-sanity+call-efficiency` keys, and later-round keys
   opening `test-sanity+call-efficiency` (counts below), so the experiment's before window draws
   a non-null `findingsPerDispatch` mean. Not working looks like: a null before mean while the
   copy holds those dispatches — the selectors match nothing.
   <!-- measured: docker exec flow-postgres psql -U flow -d flow -Atc "select count(*) filter (where dispatch_key ~ '^panel-0-mutation\+failure-modes'), count(*) filter (where dispatch_key ~ '^panel-0-test-sanity\+call-efficiency'), count(*) filter (where dispatch_key ~ '^panel-[1-9][0-9]*-test-sanity\+call-efficiency') from dispatches where role='reviewer'" → 23|76|105 @ dev store, 2026-10-30, read-only -->
3. `scripts/live-store-copy.sh down group-allowed-pairs-roles`.
