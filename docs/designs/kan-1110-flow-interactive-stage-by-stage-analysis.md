# kan-1110-flow-interactive-stage-by-stage-analysis
> Design record of change kan-1110-flow-interactive-stage-by-stage-analysis (KAN-1110), written at landing. Each decision's own **Status:** line is the status.

## Context

KAN-1110. The operator's instruction for this run, as clarified: prepare the data for the stage
analysis — for each stage its key, a summary, its branching where it has any, and whatever else
helps. The planning rounds widened "the next 5 stages" to all 31 keys in one run, with the analysis
sessions still held five at a time.

**Specification.** The approved prototype is the page's specification: the left list, the page head
and `flow.brainstorm`'s section, as every other section is built.

- File: `flow-stage-analysis-prototype.html` in the worktree\'s `.superpowers/sdd` directory
- Artifact: https://claude.ai/artifact/3U5SBVcPcYVpZSwNGCLrQC

`docs/**` is outside `ui paths` (`.flow/project.md`, `## visual verification`), so
`flow.visual-verify` does not run for this page; each task's reviewer compares the section markup
against the prototype as text (decision `text-only-check`).

**What the store offers.** Figures come from the dev workspace's `flow` database, read-only:
`stage_runs` (harness, command, attempt, outcome, cost, `metrics.dispatches`) joined to the
change's recorded decision for the class. Caveats that hold for every stage, stated once in the
page head: a stage left open is closed after 6 hours as `abandoned`, and those and open runs are
out of the duration percentiles; dollar percentiles use costed runs only; no column marks a fix
run; `suite_runs` carries no change or stage link.

**Keys.** `flow stage keys` prints 31 keys, `flow.kickoff` through `plan.session`.
`flow.pipeline-fix` has no row in `skills/flow/stage-keys.md`: no phase file marks it, the store
opens it itself (`stats/internal/store/records.go`). `plan.session` is marked only by `/flow-plan`.
The retired `flow.commit-archive` and `flow.sync-archive` are named by no file in the repository
(searched in every spelling, tracked and untracked); the 77 tracked files that still say "archive"
are live `changes/archive/` skips in guards, recorded reasons, applied migrations and a landed
design, none a leftover of a stage key.

**State.** `reviewBeforeLanding` is already `true` on the change's record, set during planning.

## Decisions

### All 31 keys in one run

**ID:** all-31-keys-one-run
**Status:** active
**Chosen:** every key `flow stage keys` prints gets its section in this change, on one query date; only the analysis sessions go five at a time — the whole pipeline is readable before the first session.
**Considered:** the first five keys only, each later batch its own `/flow` run on KAN-1110 — a cross-stage pattern is then visible only after the last batch, and the ticket stays open across seven runs. All 31 with the run pausing after each batch for its session — the branch drifts from `main` and later batches' citations are read at a different commit from the first. The five costliest keys — abandons the ticket's start-to-finish order.

### A hand-assembled HTML page

**ID:** hand-assembled-html
**Status:** active
**Chosen:** one self-contained HTML file written by hand, its figures pasted from one read-only query run, its stage SELECT kept in the page — no code, and one place to read and annotate before a session.
**Considered:** hand-assembled Markdown — first chosen, then replaced by the operator's "hand assembled html page". A new `flow-guard stage-stats` command — a Go command, a store method and a flowd route for an analysis that may end after a few sessions, leaving one caller. Extending the SPA's Stage leaderboard — store, API, SPA and tests, and live figures move between preparation and session, so findings would cite numbers that have changed.

### One committed file under docs/briefs

**ID:** one-committed-file
**Status:** active
**Chosen:** `docs/briefs/flow-stage-analysis.html`, one file — the durable record of which stages were covered and what each session agreed.
**Considered:** one file per batch — spreads the covered-stages record. A gitignored file under the change directory — gone at cleanup. Terminal only — kept nowhere.

### The section's fields

**ID:** section-fields
**Status:** active
**Chosen:** thirteen fields in the order `## Recommended approach` lists; every key carries fields 1 to 9, 12 and 13, a field with nothing to say reading `none`; field 10 appears only where it changes how a number reads, field 11 only for a stage above $100 in total — "no exceptions" holds visibly, and a stage with no loop says so.
**Considered:** key, summary, branching and stats row only — findings would be worked out live with nothing prepared. Tokens, mean dollars, maximum duration, a model list and a harness split on every stage — more columns than the hunt for repeated, uncapped or unselected work reads. A short section for a key under $10 with no loop — a threshold would decide what is looked at.

### Class split on every key

**ID:** class-split-every-key
**Status:** active
**Chosen:** all 31 sections carry the class split, as the prototype draws it — the table, then one line saying whether the stage itself branches on class; runs of a change with no recorded decision are the row `no decision recorded` — the operator's answer to round 4.
**Considered:** only where the stage branches on class — `flow.brainstorm` and every planning key lose the split that shows where their money goes (big changes: $310.97 of $456.36). Only on a stage above $100 in total — recommended in round 4, and overruled by the operator.

### Branching as condition bullets with file:line

**ID:** branching-bullets
**Status:** active
**Chosen:** nested `condition -> what runs` bullets, each citing `file:line` — nests, carries citations and reads as text.
**Considered:** a table — nested conditions do not fit cells. A Mermaid flowchart per stage, alone or beside the bullets — no precedent in this repository, and it carries no citations.

### A citation is base name and line

**ID:** citation-base-name
**Status:** active
**Chosen:** a citation reads as the prototype draws it, the file's base name and line, `brainstorm.md:127`; where one section cites two files sharing a base name, both carry their parent directory, `flow-plan/SKILL.md:12` — short as drawn, and never ambiguous.
**Considered:** every citation carrying its repository-relative path — always unambiguous, and lines far longer than the prototype's.

### An empty table cell reads —

**ID:** empty-cell-dash
**Status:** active
**Chosen:** a table cell the store has no value for reads `—` — `plan.session`'s `small` row has no costed run and so no median dollars; it reads as "no value", never as a zero.
**Considered:** `none` — the word `section-fields` gives a whole field with nothing to say. `0.00` — reads as a measured zero, which it is not.

### Which runs the figures count

**ID:** stats-window
**Status:** active
**Chosen:** the last 30 days, command `/flow`, every change state, with a fixed upper bound on `started_at` that every table of the page shares — a change stuck in a loop is what the hunt is for, and a finished-only filter hides it.
**Considered:** finished changes only — drops stuck and withdrawn changes. 90 days — mixes in the pipeline as it was before recent changes to it. Adding `/flow-fast` and `/flow-interactive` runs — those commands mark other keys and skip stages, so per-key figures would mix unlike runs.

### Claude Code only, figures and branching

**ID:** claude-code-only
**Status:** active
**Chosen:** figures are filtered with `harness = 'claude-code'`, and a branch taken only on another harness is left out of every branching list, the page head saying so once — the page describes what a Claude Code run does and nothing else.
**Considered:** zcode rows included in counts and durations — about a third of all rows, with cost figures of unchecked comparability. Every harness branch kept and marked `not analysed` — longer lists with no figure behind them. A per-stage count of excluded rows — one more number per section for no decision it informs.

### plan.session is queried under /flow-plan

**ID:** plan-session-flow-plan
**Status:** active
**Chosen:** `plan.session` gets a full section whose figures are queried with command `/flow-plan`, its caveat line naming that command — all 31 keys carry figures, and the one exception to the `/flow` filter is stated where it applies.
**Considered:** a stats row reading `no /flow runs in the window` — the key would have no figures. No section — breaks "every key `flow stage keys` prints".

### Retired keys are not on the page

**ID:** retired-keys-absent
**Status:** active
**Chosen:** the page names no key `flow stage keys` no longer prints, and this change cleans nothing up — the search found neither `flow.commit-archive` nor `flow.sync-archive` in any file, so there is no leftover.
**Considered:** one page-head line with their run counts and dollars — the operator: "if they are retired no need to show". A Jira issue auditing the guards' `changes/archive/` handling, or removing that handling here — about 45 Go files behind a review panel on what is a documentation change, and wrong if a consuming project still carries such a directory.

### Key order and six session batches

**ID:** six-session-batches
**Status:** active
**Chosen:** sections in `flow stage keys` order; the left list is headed `Session 1` to `Session 6`, holding 5, 5, 5, 5, 5 and 6 keys — the page reads start to finish, and each session's keys are visible as a group.
**Considered:** a seventh batch holding `plan.session` alone — a session for one key. The ticket's nine coarse stages — batches of one to eight keys. Order by 30-day spend — loses pipeline order.

### Layout: a sticky left key list

**ID:** sticky-left-list
**Status:** active
**Chosen:** a sticky list of the 31 keys down the left, each an anchor, the sections in one column beside it, the list moving above the content below 860 px; plain anchors and CSS, no script — every stage is one click away at any scroll position.
**Considered:** a key list at the top only — the operator scrolls back up to jump. One stage visible at a time behind tabs — needs script, and a browser search no longer finds text in a hidden stage.

### No key is marked in the left list

**ID:** no-current-mark
**Status:** active
**Chosen:** the page's left list marks no key: the prototype's `class="current"` on the `flow.brainstorm` link and its `nav.keys li a.current` CSS rule are left out of the page, the hover style staying; the prototype file itself is not edited — the one place the page departs from it, and nothing on a page with no script is ever wrong.
**Considered:** marking the key the address bar's `#anchor` names, by CSS alone — one `:has(:target)` selector per key, and it marks the last key clicked, not the section scrolled to. As drawn, `flow.brainstorm` marked whatever is on screen — matches the prototype literally and is wrong for thirty of the keys.

### Light colours only

**ID:** light-only
**Status:** active
**Chosen:** the light set of `stats/web/src/theme.css`, copied into the page — the operator's choice; least CSS.
**Considered:** following the operating system through `prefers-color-scheme` — recommended in round 2, and overruled. A manual switch — needs script and stored state for a page read a handful of times.

### Finding labels and the Session outcome box

**ID:** tag-labels-outcome-box
**Status:** active
**Chosen:** each candidate finding sits behind a small coloured label, `repeated`, `uncapped` or `unselected`; each section ends on a bordered `Session outcome` box reading `not yet held` until a fix run fills it with what was agreed and the Jira keys filed — the three kinds are told apart at a glance, and an unanalysed stage is visibly unanalysed.
**Considered:** plain bracketed words and an ordinary paragraph — nothing stands out. A page-top table of finding counts per stage — one more place to keep in step with the sections by hand. Outcomes in Jira only — the page would not record what its own analysis decided.

### Held for review; named fixes are fix runs

**ID:** held-for-review-fix-runs
**Status:** active
**Chosen:** the run stops at `IN_PROGRESS` with `reviewBeforeLanding` set; the sessions run against the page in the worktree; a fix the operator names is a fix run of this change, which also fills that stage's `Session outcome` box; a fix the fix planner classes as big gets its own ticket; the page and its fix runs land together — the operator: "yes with fixes if I say so after the analysis".
**Considered:** landing the page now with every fix its own ticket and change — many one-line fixes would each pay a full change's ceremony. Every named fix a fix run whatever its size — a large fix with no plan gate in front of it, and a landing that mixes a documentation page with pipeline changes in one review.

### The query result is a gitignored side file

**ID:** query-result-side-file
**Status:** active
**Chosen:** task 1 runs every SELECT of the page once — the stage row, per-attempt, outcomes, dispatches by model, dispatches per run, costliest runs and class split — inside one read-only `REPEATABLE READ` transaction with one fixed upper bound, and writes the result and the bound to `.superpowers/sdd/stage-stats.tsv`; tasks 2 to 7 read their keys' rows from that file and query nothing; the page carries the stage SELECT's text in a collapsed block — every table of the page is of one instant, so a section's per-attempt and class dollars add up to its stats total, and nothing extra is committed.
**Considered:** tasks 2 to 7 running the per-attempt, per-model, costliest-run and class SELECTs for their own keys with task 1's bound — first chosen; measured while planning, the stage SELECT with the bound `2026-11-13T13:20:54Z` gave `flow.brainstorm` $460.43 at 13:21Z and $461.49 at 13:41Z, because the store keeps adding harvested cost to runs already inside the window, so a section's tables would disagree with its own stats row. The same, each task also re-running the stage SELECT for its keys — each section adds up, and the sections are of six different instants. Task 1 writing every stats row into its empty section — the later tables would still need their figures from somewhere.

### Citations are read at the head task 1 starts on

**ID:** citations-at-task-1-head
**Status:** active
**Chosen:** every `file:line` is read at the worktree's head when task 1 starts, named once in the page head; a fix run that edits a cited file re-reads that stage's branching and marks the section `re-read at <commit>`; a change landing on `main` meanwhile refreshes nothing — a citation is always checkable against a named commit, and only what this change moved is redone.
**Considered:** re-reading all 31 sections before landing — 31 re-checks for citations nothing else moved. File and heading only, no line numbers — slower to follow.

### Seven tasks, one per phase-file group

**ID:** tasks-by-phase-file-group
**Status:** active
**Chosen:** a skeleton task, then one task per group of keys sharing phase files — 7, 4, 1, 5, 8 and 6 keys — run in order since all edit the one file; an implementer reads one set of skill files per task.
**Considered:** one task per session batch — a batch crosses phase files, so two tasks re-read the same files. Two tasks, skeleton then all 31 — one task reads every phase file and one review covers the whole page.

### No rendered check

**ID:** text-only-check
**Status:** active
**Chosen:** the page is checked as text against the prototype; no screenshot, spec, baseline or `ui paths` row — the operator's choice.
**Considered:** a headless open with the Playwright under `stats/web`, screenshots read and not committed — recommended in round 2, and overruled. A committed spec with baselines — every batch of sections would recapture every baseline.

### The commit hash named in the page head is a main-branch commit

**ID:** page-head-names-base-commit
**Status:** active
**Chosen:** the commit the page head names is `git merge-base HEAD main` as task 1 starts — `c07258a3` today — the `main` commit the branch's phase files are read at, since no commit of this branch before task 1 edits a cited file; a sha on the change's own branch stops resolving once integrate reshapes it.
**Considered:** naming task 1's own commit — unknowable while writing it, and unresolvable after landing.

## Open questions

## Recommended approach

Write `docs/briefs/flow-stage-analysis.html` by hand from the approved prototype: one file, inline
CSS, light colours, no script, no external font, image or stylesheet (`hand-assembled-html`,
`one-committed-file`, `light-only`, `sticky-left-list`, `no-current-mark`).

**Page head, stated once:** the citation commit (`citations-at-task-1-head`,
`page-head-names-base-commit`); the query date and filters (`stats-window`, `claude-code-only`);
that harness-only branches are left out; the store caveats of `## Context`; the SELECT in a
collapsed `<details><pre>` block.

**Each stage, `<section class="stage" id="<key>">`, in `flow stage keys` order (`six-session-batches`), with
these fields (`section-fields`):**

1. Key, and the phase file or files that define the stage (`flow.pipeline-fix`: the store code and the pipeline-fix contract; `plan.session`: `skills/flow-plan/SKILL.md`).
2. Summary — what the stage is for and what ends it.
3. Branching (`branching-bullets`, `citation-base-name`).
4. Stats row (`empty-cell-dash`) — runs, changes, median / p90 / total dollars, p50 / p90 duration, attempts per change, highest attempt, uncosted count, `abandoned` and open counts; where the highest attempt is above 1, a second table of runs and dollars per attempt. A key with no run in the window keeps the row, every cell `—`, and carries field 10 saying so.
5. Loops and their caps — each loop, its exit condition, and its cap or the word "uncapped".
6. Subagent dispatches — role, count per run, where the model and effort come from; then the recorded dispatches by model, from `metrics.dispatches`.
7. Full-suite or whole-list runs, and whether each repeats.
8. Candidate findings (`tag-labels-outcome-box`).
9. Open questions for the session.
10. A one-line data caveat, only where it changes how a number in the section reads.
11. The three costliest runs, only for a stage above $100 in total.
12. Class split (`class-split-every-key`).
13. `Session outcome` (`tag-labels-outcome-box`).

Seven tasks in order (`tasks-by-phase-file-group`, `query-result-side-file`):

1. Every SELECT, the result file, the page head, the CSS, the left list and 31 empty sections.
2. `flow.kickoff` through `flow.decide` (7 keys).
3. `flow.load-context` through `flow.sdd-tdd` (4).
4. `flow.review-panel` (1).
5. `flow.verify` through `flow.write-in-progress` (5).
6. `flow.preflight` through `flow.landing-routes` (8).
7. `flow.verify-merge` through `plan.session` (6; `plan-session-flow-plan`).

Rejected, each with its reason under the decision named: a first-five-only change
(`all-31-keys-one-run`); a `flow-guard` command or an SPA view (`hand-assembled-html`); landing
before the sessions (`held-for-review-fix-runs`); any archive cleanup (`retired-keys-absent`).

## Critical files

- `docs/briefs/flow-stage-analysis.html` — the one file the change adds.
- `flow-stage-analysis-prototype.html` in the worktree's `.superpowers/sdd` directory — the approved prototype; gitignored, the specification.
- `.superpowers/sdd/stage-queries.sql` — task 1's query file; gitignored, never committed.
- `.superpowers/sdd/stage-stats.tsv` — task 1's query result and time bound; gitignored, never committed.
- `skills/flow/stage-keys.md` — which phase file marks each key; read, not edited.
- `skills/flow/*.md`, `skills/flow-contracts/*.md`, `skills/flow-plan/SKILL.md` — the phase files and contracts the branching lists cite; read, not edited.
- `stats/internal/store/records.go` — where the store opens `flow.pipeline-fix`; read, not edited.
- `stats/internal/guard/*.go` — the guards a stage runs, cited where a loop or cap lives in one; read, not edited.
- `stats/web/src/theme.css` — source of the light colour values; read, not edited.

## Existing code to reuse

- The prototype's markup and inline CSS — copied as the page's skeleton and as the model for every section.
- `flow stage keys` — the key list and its order.
- The prototype's `<details>` SELECT — the one-row-per-stage query, re-run with a fixed upper bound.
- The dev workspace's `flow` database on host port 5433, read-only, as `.flow/project.md`'s `## run` gives the connection.

## Verification

- `scripts/flow-guard run-verify-list <worktree> lint` exits 0.
- `flow stage keys | while read k; do grep -c "<section class=\"stage\" id=\"$k\">" docs/briefs/flow-stage-analysis.html; done` prints `1` thirty-one times, and `grep -c '<section class="stage" id=' docs/briefs/flow-stage-analysis.html` prints `31`.
- `git grep -n -i -e 'commit-archive' -e 'sync-archive' -- docs/briefs/flow-stage-analysis.html` prints nothing.
- `grep -c 'not yet held' docs/briefs/flow-stage-analysis.html` prints `31`; `grep -c '<h3>Class split</h3>'` prints `31`.
- `grep -n -E 'https?://|<script|<link|@import|prefers-color-scheme|class="current"' docs/briefs/flow-stage-analysis.html` prints nothing but Jira or artifact links inside prose, if any.
- Read against the prototype as text: the head, the left list and `flow.brainstorm`'s section match it element for element but for `no-current-mark`; every other section carries the same fields in the same order.
- Spot-check three `file:line` citations per task against `git show <the commit the page head names>:<file>`.

## Live check

None — the change adds one static HTML document; it touches no service, daemon, store or scheduler, and its only contact with the store is a read-only SELECT run while writing.

## Non-goals

- No Go, SPA, skill, contract or guard change; no new command or view.
- No fix to any finding in the creating run — fixes come only from the operator, after a session, as fix runs.
- No cleanup of the word "archive" or of the guards' `changes/archive/` handling.
- No dark colour scheme, no script, no rendered check, no `ui paths` row.
- No figures for zcode runs, and none for `/flow-fast` or `/flow-interactive`.
- No refresh of figures or citations when `main` moves while the change is held.

## Risks

- **A branching list misreads a phase file.** Limited by `file:line` on every bullet, the per-task review and the spot-check in `## Verification`; the sessions read each list with the operator.
- **Figures age while the change is held.** The page head names the query date; the page is a snapshot by design (`all-31-keys-one-run`).
- **The branch drifts from `main` across the sessions.** The change adds one new file, so a conflict needs a fix run to touch a file `main` also moved; integrate's preflight catches it.
- **A fix run moves cited lines.** `citations-at-task-1-head` re-reads the edited stage and marks it.
- **A long hand-written page drifts from the prototype section by section.** Every task's reviewer compares against the same prototype; fields in a fixed order with `none` where empty make a missing field visible.
- **Thin data misleads.** `plan.session` has 4 rows and several mechanical keys cost cents; the stats row shows run counts beside every percentile, and field 10 states it where it matters.
- Accepted: the six section tasks each repeat the same verify block, where one loop over all keys in task 7 would do (plan-review-2 F3) — a per-task check catches a malformed section before the next task builds on the same file, where a single loop in task 7 would find it five tasks late.
