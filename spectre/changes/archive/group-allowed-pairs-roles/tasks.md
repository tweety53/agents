# group-allowed-pairs-roles

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no
> **Design shape:** plan-mode

**Goal:** the Allowed pairs table grouped by pipeline area with a plain-English `Does` line per role, every fixed review-panel bundle its own role, flowd serving each role's group and description, and the Experiments view drawing one compact row per role under group headers, as the approved prototype draws.

**Architecture:** the markdown table stays the one source; `stats/internal/modelpolicy` parses group and description from it, flowd serves them on `GET /api/v1/roles`, and the SPA groups by them. `plan-class.sh` and the experiment stats name the two bundle roles that replace `panel-extra`.

**Tech Stack:** Markdown, Go (`stats/`), React/TypeScript (`stats/web`), Playwright.

**Spec:** `design.md` beside this file; the approved prototype `/Users/tweety53/Projects/agents-worktrees/group-allowed-pairs-roles/.superpowers/sdd/roles-prototype.html` (https://claude.ai/artifact/D82LpNhBssL7oEaewaeu5Y).

## Global constraints

- Every task runs in `/Users/tweety53/Projects/agents-worktrees/group-allowed-pairs-roles` (`<wt>`), branch `spectre/group-allowed-pairs-roles`; `**Files:**` paths are relative to `<wt>`. Never touch the main checkout, never create a worktree, never write `spectre/changes/` — this change's `verbatim-moves.txt` is committed with the plan and already lists every sentence tasks 1 and 2 add or reword.
- Go: `cd stats && gofmt -w .` first; `cd stats && make -s web-dist && go vet ./... && gofmt -l .` exits clean. SPA: `cd stats && make -s web-dist && cd web && npx tsc -b` exits 0.
- Store tests use the package's existing test-database helper, never the live `flow` database. Never stop, restart or drop `flowd` on 4173, the `flow-postgres` container or the `flow` database; the visual tasks run on the disposable UI-test stack the project's `capture` and `verify` commands start.
- Every Go test run sets a Bash `timeout` of 600000.
- Mutation proofs restore with `scripts/break-and-prove.sh` — never `git checkout --` over uncommitted edits.

## Review Focus

- Task 1: the parser refuses, never loads fail-open — a role under two groups, an unmarked row with no `Does`, a row of the wrong width; a table after the next `## ` heading is not read; every guard's per-role row order is unchanged.
- Task 3: each bundle's pass-1 dispatches count under its own role alone, and every later-round dispatch counts under `panel-rerun`.
- Task 5: the view derives one row per role from the served rows — no role, pair or marker literal in the component but `main`'s.

- [x] 1. Group the Allowed pairs table, give each role a Does line, and split panel-extra
  - [x] **Step 1: Write the failing tests.** `stats/internal/modelpolicy/modelpolicy_test.go`: `header` becomes `"| Role | Allowed pairs | Requires marker | Does |\n|---|---|---|---|\n"`; every fixture row gains a fourth cell — `Does a.`-style text on an unmarked row, `—` on a marked one — and `wantEntries` carries each entry's `Does` (`""` on the marked row); the refused-table map gains `"three cells": header + "| `a` | `opus/flow-low` | — |\n"`, `"no does": header + "| `a` | `opus/flow-low` | — | — |\n"` and `"two groups": "### G\n\n" + header + "| `a` | `opus/flow-low` | — | x |\n\n### H\n\n" + header + "| `a` | `opus/flow-medium` | — | x |\n"`; the real-table role list swaps `panel-extra` for `panel-sanity-efficiency` and `panel-mutation-failure` and adds `planner`, asserts `panel-extra` has no row, and over `Rows` of the real file asserts every entry's `Group` is non-empty, every unmarked entry's `Does` is non-empty, and the groups in order of first appearance are exactly `Planning`, `Implementation`, `Review`, `Visual verification`, `Self-review`, `Pipeline fix`. New `TestParseGroups`: `Rows` over `"## Allowed pairs\n\nintro\n\n### G\n\n" + header + "| `a` | `opus/flow-low` | — | Does a. |\n| `a` | `opus/flow-high` | `M` | — |\n\n### H\n\n" + header + "| `b` | `opus/Plan@low` | — | Does b. |\n\n## Next\n\n### I\n\n" + header + "| `c` | `opus/flow-low` | — | Does c. |\n"` returns exactly `{a, G, "Does a."}`, `{a, G, "", marker M}`, `{b, H, "Does b."}` — `c`, past the next `## ` heading, is not read; a table with no `###` heading above it reads `Group` `""`. `stats/internal/guard/check_dispatch_pair_test.go` (`TestPolicyTableRefusesMalformedRows`): the fixture header becomes the four-column header and separator, and each row but `"two cells"` gains a fourth cell (`| x |`, `| — |` on the two-markers row), so each still exits 2 for its own reason. `stats/internal/guard/plan_class_test.go`: `pcExtraPairs(t, role string)` returns `rows[role][0].Pairs`; `TestPlanClassExtra` passes `"panel-sanity-efficiency"`, `TestPlanClassExtraBundles` passes `"panel-mutation-failure"`; `TestPlanClassExtraFollowsPolicy` rewrites both rows' prefixes — ``"| `panel-sanity-efficiency` | `opus/flow-low`, `sonnet/flow-low` |"`` to ``"| `panel-sanity-efficiency` | `opus/flow-low` |"`` and the same for `panel-mutation-failure` — fatal when either is absent, and still wants both extra lines on `opus low`; a further run over the real table with the whole `panel-mutation-failure` row deleted exits 2 with `r.out` naming `panel-mutation-failure`. Run `cd stats && go test -count=1 ./internal/modelpolicy -run 'TestParse'` — fails.
  - [x] **Step 2: Parse the grouped tables.** `stats/internal/modelpolicy/modelpolicy.go`: `Entry` becomes `struct { Role, Group, Does string; Row }`. `Rows` scans every line: a line opening `### ` sets the current group to the rest of the line, trimmed; a line opening `## ` ends the scan once at least one table row was read; the four-column header `| Role | Allowed pairs | Requires marker | Does |` opens a table; inside a table a line not opening `|` closes it and the scan goes on; a `|---` line is skipped. A row splits into exactly 4 cells, else the existing cell-count error (`want 4`); role, pairs and marker go through `ticks` as today; `Does` is cell 4 trimmed, `—` read as `""`. An unmarked row with an empty `Does` is an error naming the row; a role already read under a different group is an error naming the role and both groups. The missing-table error and `Parse`'s doc comment name the four-column header. `Parse` keeps building its map from `Rows`.
  - [x] **Step 3: Rewrite the table.** In `skills/flow-contracts/model-policy.md`, keep `## Allowed pairs` and its first paragraph, and replace everything after that paragraph up to (not including) `## Experiments` with exactly this block, followed by one blank line:

```markdown verified:applied to a scratch clone of this branch, where check-verbatim-moves.sh against HEAD exits 0 with this change's verbatim-moves.txt
Each `###` heading below is one pipeline area, and a role sits under exactly one; a role's
unmarked row carries its `Does` cell, a plain-English line on what the role does, and a marked
row's `Does` is `—`. flowd serves both on `GET /api/v1/roles`, and the Experiments view groups its
roles table by them.

### Planning

| Role | Allowed pairs | Requires marker | Does |
|---|---|---|---|
| `explorer` | `opus/Explore@low`, `opus/Explore@medium` | — | Reads the codebase to answer a planner's question. |
| `plan-designer` | `opus/Plan@low`, `opus/Plan@medium` | — | Proposes a design or plan for the planner to choose from. |
| `planner` | `opus/flow-planner-medium`, `opus/flow-planner-low` | — | Writes the design and the task plan. |

### Implementation

| Role | Allowed pairs | Requires marker | Does |
|---|---|---|---|
| `implementer` | `opus/flow-low`, `opus/flow-medium`, `sonnet/flow-low`, `sonnet/flow-medium` | — | Writes the code for a group of tasks, tests first. |
| `panel-fix` | `opus/flow-low`, `opus/flow-medium`, `sonnet/flow-low`, `sonnet/flow-medium` | — | Fixes what the review panel found. |
| `implementer` | `opus/flow-high` | `No-progress escalation` | — |
| `panel-fix` | `opus/flow-high` | `No-progress escalation` | — |

### Review

| Role | Allowed pairs | Requires marker | Does |
|---|---|---|---|
| `gated-reviewer` | `opus/flow-medium` | — | Reviews risky tasks as they land, on big changes only. |
| `gated-re-reviewer` | `opus/flow-low` | — | Re-checks a gated review's findings after their fix. |
| `panel-reviewer` | `opus/flow-low`, `opus/flow-medium` | — | Reviews the whole branch, first pass: the primary + principles bundle. |
| `panel-reviewer` | `opus/flow-high` | `Operator-approved high` | — |
| `panel-sanity-efficiency` | `opus/flow-low`, `sonnet/flow-low` | — | Always-on extra bundle: checks each new test would fail if its behaviour broke, and spots wasted calls. |
| `panel-mutation-failure` | `opus/flow-low`, `sonnet/flow-low` | — | Always-on extra bundle: breaks the code on purpose to prove the tests notice, and checks error, timeout and partial-write paths. |
| `panel-rerun` | `opus/flow-low`, `sonnet/flow-low` | — | Re-reviews the branch after a fix round. |
| `late-fix` | `opus/flow-low` | — | Reviews just a small fix made after the branch already passed review. |
| `panel-merger` | `sonnet/flow-low` | — | Merges the reviewers' findings into one list without duplicates. |

### Visual verification

| Role | Allowed pairs | Requires marker | Does |
|---|---|---|---|
| `visual-verifier` | `opus/flow-low` | — | Screenshots the UI and compares it with the mockup. |
| `tooling-analyst` | `opus/flow-medium` | — | Works out why visual verification missed a defect. |

### Self-review

| Role | Allowed pairs | Requires marker | Does |
|---|---|---|---|
| `self-review-pass` | `opus/flow-medium` | — | Reviews a finished change and the run that built it. |
| `self-review-plan` | `opus/flow-planner-medium`, `opus/flow-planner-low` | — | Plans the fixes for the self-review's findings. |
| `self-review-fix` | `opus/flow-low` | — | Makes one planned self-review fix. |
| `self-review-fix` | `opus/flow-high` | `No-progress escalation` | — |
| `self-review-review-1` | `opus/flow-medium`, `opus/flow-low` | — | First review of a self-review fix. |
| `self-review-review-r` | `sonnet/flow-low` | — | Re-review once that fix is corrected. |

### Pipeline fix

| Role | Allowed pairs | Requires marker | Does |
|---|---|---|---|
| `pipeline-fix-plan` | `opus/flow-planner-medium`, `opus/flow-planner-low` | — | Plans the fix for a defect found in the flow tooling mid-run. |
| `pipeline-fix-fix` | `opus/flow-low` | — | Makes that tooling fix. |
| `pipeline-fix-fix` | `opus/flow-high` | `No-progress escalation` | — |
| `pipeline-fix-review-1` | `opus/flow-medium`, `opus/flow-low` | — | First review of the tooling fix. |
| `pipeline-fix-review-r` | `sonnet/flow-low` | — | Re-review once that fix is corrected. |
```

  - [x] **Step 4: Roll each extra bundle off its own row.** `stats/internal/guard/planclass.go`: read `policy["panel-sanity-efficiency"]` and `policy["panel-mutation-failure"]`; when the table fails to read or either is empty, print `plan-class.sh: the allowed-pairs table names no <role> row (<err>)` naming the missing role and return 2; `pcExtra("test-sanity+call-efficiency", …)` gets the first row's pairs, `pcExtra("mutation+failure-modes", …)` the second's; roll keys `extra` / `extramf` and the sonnet-below-50 rule unchanged. `pcExtra`'s doc comment names the bundle's own row in place of the panel-extra row.
  - [x] **Step 5: Prove and verify.** `scripts/break-and-prove.sh` with Step 2's two-groups check removed — `TestParse`'s `"two groups"` case fails; restore. `cd stats && make -s web-dist && go test -race -count=1 ./internal/modelpolicy ./internal/repotest ./internal/api` and `cd stats && go test -race -count=1 ./internal/guard -run 'TestParse|TestPolicyTable|TestCheckDispatchPair|TestPlanAgentPair|TestPlanClass|TestCheckLoopDispatch|TestPipelineFix|TestWritePanelDiff'` pass; `scripts/run-verify-list.sh /Users/tweety53/Projects/agents-worktrees/group-allowed-pairs-roles lint` exits clean (its `check-verbatim-moves.sh` reads the committed `verbatim-moves.txt`). Commit.

**Build:** green
**Files:** `skills/flow-contracts/model-policy.md`, `stats/internal/modelpolicy/modelpolicy.go`, `stats/internal/modelpolicy/modelpolicy_test.go`, `stats/internal/guard/check_dispatch_pair_test.go`, `stats/internal/guard/planclass.go`, `stats/internal/guard/plan_class_test.go`
**Allowed-collateral:** `stats/internal/guard/*_test.go`
**Tests:** `TestParse`, `TestParseGroups`, `TestPlanClassExtraFollowsPolicy`
**Regression:** `TestParseGroups` fails — the parser reads one table, so every group after the first is lost to every guard; `TestPlanClassExtraFollowsPolicy` fails — `plan-class.sh` still reads a `panel-extra` row the table no longer has.
**Commit:** feat(modelpolicy): group the allowed-pairs table with a Does line per role
**After:** none

**Decision:** groups-in-markdown-headings
**Decision:** groups-by-pipeline-area
**Decision:** group-and-does-from-table
**Decision:** row-order-kept
**Decision:** split-panel-extra
**Decision:** one-panel-rerun-row

- [x] 2. Name the two extra-bundle roles in the dispatch rules
  - [x] **Step 1: Edit.** `skills/flow/implement.md`, the dispatch-site table's panel-bundle row, **Role line** cell: replace ``; `panel-extra` for an always-on extra dispatch;`` with ``; `panel-sanity-efficiency` for the `test-sanity+call-efficiency` extra dispatch and `panel-mutation-failure` for the `mutation+failure-modes` one;``. `skills/flow/brainstorm-planner.md`, **Model and effort**, the extra-dispatches bullet: replace ``each its own rolled pair off the `panel-extra` row of **Allowed pairs**`` with ``each its own rolled pair off its own row of **Allowed pairs** — `panel-sanity-efficiency` and `panel-mutation-failure`` — the rest of the bullet unchanged.
  - [x] **Step 2: Verify.** `git grep -n -F -e 'panel-extra' -- skills/flow/implement.md skills/flow/brainstorm-planner.md` prints nothing. `scripts/run-verify-list.sh /Users/tweety53/Projects/agents-worktrees/group-allowed-pairs-roles lint` exits clean. Commit.

**Build:** green
**Files:** `skills/flow/implement.md`, `skills/flow/brainstorm-planner.md`
**Tests:** **none** — prose; Step 2's sweep and guards are the check.
**Regression:** none — prose.
**Commit:** docs(flow): name the two extra-bundle roles in the dispatch rules
**After:** Task 1

**Decision:** split-panel-extra

- [x] 3. Attribute each extra bundle's dispatches to its own role in the experiment stats
  - [x] **Step 1: Write the failing test.** `stats/internal/store/experimentstats_test.go`, `TestExperimentStatsExtraBundleRoles`: for each of `panel-reviewer` → 3, `panel-sanity-efficiency` → 1, `panel-mutation-failure` → 2, `panel-rerun` → 4, a fresh `newStatsFixture(t)` opens an experiment with that one cell (`opus/flow-low`); a before change `e1` at `fx.day(-1)` carries reviewer dispatches `panel-0-primary+principles` with three round-0 findings, `panel-0-test-sanity+call-efficiency` with one, `panel-0-mutation+failure-modes` with two and `panel-1-test-sanity+call-efficiency` with four round-1 findings; an after change `e2` at `fx.day(2)` carries one dispatch of each of those four keys and no finding; `wantMeans(t, got, "findingsPerDispatch", f64(<want>), f64(0))`. Run `cd stats && go test -count=1 ./internal/store -run 'TestExperimentStats'` — fails (the two new roles select nothing; `panel-rerun` misses the re-run that opens `test-sanity+call-efficiency`).
  - [x] **Step 2: Retarget the selectors.** `stats/internal/store/experimentstats.go`: `roleMetrics`'s reviewer case lists `panel-sanity-efficiency` and `panel-mutation-failure` in place of `panel-extra`. `roleDispatchSQL`: drop `panel-extra`; `panel-reviewer` becomes `d.role = 'reviewer' AND d.dispatch_key ~ '^panel-0-' AND d.dispatch_key !~ '^panel-0-(test-sanity\+call-efficiency|mutation\+failure-modes)'`; add `panel-sanity-efficiency`: `d.role = 'reviewer' AND d.dispatch_key ~ '^panel-0-test-sanity\+call-efficiency'` and `panel-mutation-failure`: `d.role = 'reviewer' AND d.dispatch_key ~ '^panel-0-mutation\+failure-modes'`; `panel-rerun` becomes `d.role = 'reviewer' AND d.dispatch_key ~ '^panel-[1-9][0-9]*-'`.
  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && go test -race -count=1 ./internal/store -run 'TestExperimentStats'` passes; `cd stats && make -s web-dist && go vet ./... && gofmt -l .` exits clean. Commit.

**Build:** green
**Files:** `stats/internal/store/experimentstats.go`, `stats/internal/store/experimentstats_test.go`
**Tests:** `TestExperimentStatsExtraBundleRoles`
**Regression:** `TestExperimentStatsExtraBundleRoles` fails — an experiment on either extra-bundle role draws no figures, and re-runs keyed `test-sanity+call-efficiency` drop out of `panel-rerun`.
**Commit:** fix(store): attribute each extra review bundle's dispatches to its own role
**After:** none

**Decision:** extra-stats-pass-one

- [x] 4. Serve each role's group and description on GET /api/v1/roles
  - [x] **Step 1: Write the failing test.** `stats/internal/api/experiments_test.go`, `TestRolesRoute`: its local `entry` type gains `Group string `json:"group"`` and `Does string `json:"does"``; entry 0 wants `Group: "Session"`, `Does: "The session you talk to: runs the pipeline and sends out every other role."`; entry 1 wants `explorer` under `Planning` with `Does` `Reads the codebase to answer a planner's question.`; the marked implementer row wants group `Implementation` and `Does` `""`; every entry's `Group` is non-empty. Run `cd stats && go test -count=1 ./internal/api -run TestRolesRoute` — fails.
  - [x] **Step 2: Implement.** `stats/internal/api/experiments.go`: `roleEntry` gains `Group string `json:"group"`` and `Does string `json:"does"``; the leading `main` entry carries the two literals above; each table row's entry carries `row.Group` and `row.Does`; the `roles` doc comment says each entry carries its group and description from the table, `main`'s from these literals.
  - [x] **Step 3: Verify.** `cd stats && make -s web-dist && go test -race -count=1 ./internal/api` passes; `cd stats && make -s web-dist && go vet ./... && gofmt -l .` exits clean. Commit.

**Build:** green
**Files:** `stats/internal/api/experiments.go`, `stats/internal/api/experiments_test.go`
**Tests:** `TestRolesRoute`
**Regression:** `TestRolesRoute` fails — the roles API serves no group or description, so the view cannot group.
**Commit:** feat(api): serve each role's group and description on the roles route
**After:** Task 1

**Decision:** main-group-literal

- [x] 5. Draw the roles table grouped, one compact row per role
  - [x] **Step 1: Write the failing test.** `stats/web/src/views/views.test.tsx`, `describe("Experiments")`: the `roles` fixture gains `group` and `does` on every entry — `main` `Session` / `The session you talk to: runs the pipeline and sends out every other role.`, `explorer` and `plan-designer` `Planning` with their prototype lines, and a new unmarked `{ role: "implementer", pairs: ["opus/flow-low"], marker: "", group: "Implementation", does: "Writes the code for a group of tasks, tests first." }` before the existing marked implementer row (`group: "Implementation"`, `does: ""`). The case `Experiments lists every roles row, …` is replaced by `Experiments groups the roles, one row per role, folding escalations and linking overrides`: the column headers are exactly `Role`, `Does`, `Allowed pairs`; the rows' first cells read, in order, `Role`, `Session`, `main`, `Planning`, `explorer`, `plan-designer`, `Implementation`, `implementer` — one `implementer` row; each group row's one header cell spans 3 columns; the `plan-designer` row's `s` reads `opus/Plan@low, opus/Plan@medium`, its `opus/Plan@xhigh` has class `exp`, it links `#3` and its Does cell reads its line; the `implementer` row holds `opus/flow-low` and an `.esc` element reading `opus/flow-high with No-progress escalation`; `main`'s `s` reads `session's own pair`; `explorer` has no link and no `s`; no `Requires marker` or `Experiment` header and no `From skills/flow-contracts/model-policy.md` text render; clicking `#2` selects `main agent opus@medium`. Run `cd stats/web && npx vitest run src/views/views.test.tsx -t 'Experiments'` — fails.
  - [x] **Step 2: Type.** `stats/web/src/api.ts`: `RoleEntry` gains `group: string;` and `does: string;`.
  - [x] **Step 3: Draw.** `stats/web/src/views/Experiments.tsx`, `RolesTable`, as the prototype draws it: a `thead` row `Role | Does | Allowed pairs`; then, per group in order of first appearance, `<tr className="grp"><th colSpan={3}>{group}</th></tr>` and one `<tr className="role">` per role in order of first appearance — `td.role` the `code` id; `td.does` the unmarked row's `does`; `td.pairs` the unmarked row's pairs joined `, ` in one `code` (`MAIN_PAIR` for `main`) or, under an override, that `code` inside `s`, a space, the override's pairs in `code.exp` and the `#<id>` link selecting the experiment as today, then one `<span className="esc"><code>{pair list}</code> with <code>{marker}</code></span>` per marked row. The `p.sub` footnote goes. The file's header comment also cites `.superpowers/sdd/roles-prototype.html` for the roles table. `tablePairs` is unchanged.
  - [x] **Step 4: Style.** `stats/web/src/styles.css`, after `.experiments .roles .exp`, the prototype's table rules scoped `.experiments .roles`: `table` `font-size: 12px; line-height: 1.35`; `th, td` `padding: 2px 8px; vertical-align: top`, `td` `border-bottom: none`; `th` `font-weight: 600; color: var(--text-secondary); border-bottom: 1px solid var(--border)`; `tr.grp th` `padding-top: 8px; font-size: 0.68rem; letter-spacing: 0.06em; text-transform: uppercase; color: var(--text-muted)`; `tbody tr.role:hover td` `background: var(--surface-tip)`; `td.role, td.pairs` `white-space: nowrap; width: 1%`; `td.does` `color: var(--text-secondary)`; `code` `font-size: 11.5px`; `.esc` `display: block; color: var(--text-muted)`; `a` `color: var(--series-1); text-decoration: none; margin-left: 4px`; `a:hover, a:focus` `text-decoration: underline; outline: none`. The panel's own box and title rules are untouched.
  - [x] **Step 5: Verify.** `cd stats/web && npx vitest run src/views/views.test.tsx -t 'Experiments'` passes; `cd stats && make -s web-dist && cd web && npx tsc -b` exits 0. Commit.

**Build:** green
**Files:** `stats/web/src/api.ts`, `stats/web/src/views/Experiments.tsx`, `stats/web/src/styles.css`, `stats/web/src/views/views.test.tsx`
**Allowed-collateral:** `stats/web/src/api.test.ts`, `stats/web/src/App.test.tsx`
**Tests:** `Experiments groups the roles`
**Regression:** the `Experiments groups the roles` case fails — the view draws the flat per-row table with its marker and experiment columns.
**Commit:** feat(web): draw the roles table grouped, one compact row per role
**After:** Task 4

Correction (2026-10-30): the plan named only the roles case in `views.test.tsx`; this task's commit also fixes `Experiments draws a mean-after line with no before mean as neutral`, red on the base, which read the trend panel's charts synchronously before the stats fetch drew them — it now awaits `findAllByRole("img")`.

**Decision:** compact-one-row-per-role
**Decision:** roles-footnote-dropped
**Decision:** roles-panel-chrome-unchanged

- [x] 6. Capture the roles panel against the prototype
  - [x] **Step 1: Seed experiment 4's cells.** `stats/cmd/uitest-seed/seed.go`, `seedExperiments`: the open experiment "planning on xhigh" carries `explorer: {"sonnet/Explore@low"}`, `plan-designer: {"opus/Plan@medium", "opus/Plan@high", "opus/Plan@xhigh"}`, `planner: {"opus/flow-planner-medium", "opus/flow-planner-high", "opus/flow-planner-xhigh"}`; its name and start date unchanged; the doc comment names the cells as experiment 4's.
  - [x] **Step 2: Capture tests.** `stats/web/tests/visual/experiments.spec.ts` gains `experiments roles panel, row hover and link focus`: `open(page)`; `const roles = page.getByRole("region", { name: "Current roles" })`; assert `roles.locator("table")` has CSS `font-size` `12px` and its first `td` `padding` `2px 8px`; `await expect(roles).toHaveScreenshot("experiments-roles.png")`; hover `roles.getByRole("row", { name: /^plan-designer/ })` and screenshot `experiments-roles-hover.png`; `page.mouse.move(0, 0)`, focus that row's `#<id>` link, assert it is focused, screenshot `experiments-roles-focus.png`. Every locator by role or name.
  - [x] **Step 3: Capture.** Run the project's `capture` command (`.flow/project.md`, `## visual verification`) for `tests/visual/experiments.spec.ts` and `tests/visual/full-app-suite.spec.ts`; read every new and changed PNG; rebuild `stats/web/tests/visual/full-app-suite.zip` with `scripts/rebuild-full-app-suite-zip.sh`.
  - [x] **Step 4: Measure against the prototype.** In `/Users/tweety53/Projects/agents-worktrees/group-allowed-pairs-roles/.superpowers/sdd/roles-frames/` (gitignored), a throwaway Node script on `stats/web`'s installed Playwright opens `file:///Users/tweety53/Projects/agents-worktrees/group-allowed-pairs-roles/.superpowers/sdd/roles-prototype.html` at the capture's viewport, device scale factor and dark colour scheme, sized so its `Current roles` region is the capture's width, and screenshots that region at rest, with its `plan-designer` row hovered and with that row's link focused, as `experiments-roles.png`, `experiments-roles-hover.png`, `experiments-roles-focus.png`; a map file there opens `frame none` and pairs each capture with its frame. Run `compose-mockup-frames.sh` on that map and directory and `measure-visual-properties.sh`'s band, seam and matrix reads on every pair; fix `stats/web/src/views/Experiments.tsx` or `stats/web/src/styles.css` and recapture until each pair is clean or each leftover is named with its cause in the commit body — the `#<id>` text (`seed-matches-experiment-4`) and the panel box (`roles-panel-chrome-unchanged`) are the expected ones.
  - [x] **Step 5: Verify.** The project's `verify` command passes. Commit.

**Build:** green
**Files:** `stats/cmd/uitest-seed/seed.go`, `stats/web/tests/visual/experiments.spec.ts`, `stats/web/tests/visual/experiments.spec.ts-snapshots/experiments-roles-darwin.png`, `stats/web/tests/visual/experiments.spec.ts-snapshots/experiments-roles-hover-darwin.png`, `stats/web/tests/visual/experiments.spec.ts-snapshots/experiments-roles-focus-darwin.png`, `stats/web/tests/visual/experiments.spec.ts-snapshots/experiments-open-darwin.png`, `stats/web/tests/visual/experiments.spec.ts-snapshots/experiments-closed-darwin.png`, `stats/web/tests/visual/experiments.spec.ts-snapshots/experiments-hover-darwin.png`, `stats/web/tests/visual/experiments.spec.ts-snapshots/experiments-focus-darwin.png`, `stats/web/tests/visual/full-app-suite.spec.ts-snapshots/full-experiments-darwin.png`, `stats/web/tests/visual/full-app-suite.zip`
**Allowed-collateral:** `stats/web/tests/visual/*-snapshots/*`, `stats/cmd/uitest-seed/seed_test.go`, `stats/web/src/views/Experiments.tsx`, `stats/web/src/styles.css`
**Tests:** `experiments roles panel`
**Regression:** `experiments roles panel` fails — the roles panel drifted from the approved prototype.
**Commit:** test(web): capture the grouped roles panel against its prototype
**After:** Task 5

**Decision:** seed-matches-experiment-4
**Decision:** roles-captures-in-experiments-spec
