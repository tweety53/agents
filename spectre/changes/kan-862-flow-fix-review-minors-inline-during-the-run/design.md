## Context

One change, because the guard, the review prose, the handoff and the dashboard all retire the same
`deferred` status — §§ 1–7 below.

## 1. Where Minors go today, and what that costs

- **Per-task gated reviewer** (`skills/flow/implement.md`, **The gated per-task reviewer**): a pass
  whose findings are all Minor closes `clean`, and the parent appends each Minor to
  `KNOWN-BUGS.md` before ticking the task. A `fix` pass sends its Minors to the gated fix.
- **Panel round** (`skills/flow/review-panel.md`, **Panel re-runs**): a round with no Critical and
  no Important records every Minor `deferred <reason>` with a category, then appends it to
  `KNOWN-BUGS.md` (**Deferred findings go to KNOWN-BUGS.md at round close**). A round with a
  Critical or Important sends its Minors to that round's fix.
- **The pile:** at `0b7e1c58`, `KNOWN-BUGS.md` held 97 entries — 96 deferred Minors and one
  sweep entry for a pre-existing load-sensitive test failure. Kan-860 alone added 30.
  <!-- measured: grep -c '^- ' KNOWN-BUGS.md → 97; grep -c Minor KNOWN-BUGS.md → 96 @ 0b7e1c58 -->
- **The later fix is the expensive part.** KAN-861 (`/flow-fast`, class `big`, 17 tasks, 67
  findings, opus/high implementers, four waves) is the Minor-fixing route this change replaces.
- **What the Minors are:** about 46 are drift (comments, file headers, usage lines or citations
  that no longer match what the task changed), about 16 are missing test cases (an untested
  refusal or exit path), 18 are marked cosmetic, and the rest are edge cases.
  <!-- measured: keyword counts over the `## Deferred review findings` entries @ 0b7e1c58 -->

## 2. The inline Minor fix

- **Who fixes:** the parent, inline, at the point it wrote the `KNOWN-BUGS.md` entry before —
  reading the Minor costs about what writing the entry did. No dispatch, no dispatch record.
- **Panel:** a round with no Critical and no Important has every Minor fixed by the parent at the
  round's close: one commit per worktree at the branch tip (the on-top route, pathspec-scoped,
  pushed per **Branch backup**), then `flow record status -status fixed` per finding. No fix
  subagent, no reproducer run, no mutation proof, no slot re-run.
- **Per-task:** a Minor-only pass is still `clean`. Before ticking the task, the parent fixes its
  Minors the same way, one commit at the tip, applied only while no implementer dispatch is
  writing that worktree.
- **Unchanged:** a Minor beside a Critical or Important still goes to that round's fix or the
  gated fix, closed by that fix's verification.
- **Verification:** none of the fix's own. Per-task fixes are read by the panel, which reads the
  whole branch diff later. Panel fixes are covered by `flow.verify`'s lint and tests, which run
  after the panel.
- **Staleness carve-out:** the inline Minor commit leaves every slot's result current — the one
  source change after a slot's last read that does not make that result stale. Without it, a
  Minor fix would force the re-runs this path exists to avoid.
- **The only exit is `withdrawn`:** a Minor no tree change can resolve, or one that is not a
  defect, is withdrawn with a one-clause reason, as the handback loop withdraws. Out of scope,
  pre-existing and cosmetic are not reasons: they are how 96 entries piled up.
- **`/flow-fast` parity:** `/flow-fast` runs `review-panel.md` "**Check base movement first**
  through **Panel re-runs**" and never runs `close-task.sh`, so no gated reviewer fires there. The
  new panel rule therefore has to live inside `## Panel re-runs` to reach `/flow-fast` with no
  edit of its own. `/flow-fast` runs no guards, so there the prose rule is the whole enforcement.

## 3. The findings-closed guard

`check-panel-findings-closed.sh` (`stats/internal/guard/panelfindingsclosed.go`) today:

- counts `deferred <reason>` as closed;
- fails a Minor recorded `deferred` in a round with a live Critical or Important;
- fails every `fixed` finding whose slot has no clean re-run in a later round, Minors included.
  This contradicts **Panel re-runs**' own "a fixed Minor closes on the verification below alone"
  wherever a slot raised only Minors in a round whose Important went to a fix.

After: `deferred` is open (so the misdeferred class collapses into "still open" and is deleted),
and a Minor recorded `fixed` needs no re-run. The store and CLI keep accepting and reading
`deferred`, because historical rows and migration `0025` carry it.

## 4. Retiring the deferral surface

- `skills/flow-contracts/known-bugs.md` loses `## Deferred review findings` and its **Where it
  runs** bullet. The sweep for pre-existing test failures stays.
- `skills/flow/review-panel.md` loses **Deferred findings go to KNOWN-BUGS.md at round close**,
  and the ROUND SCOPE paragraph's sentence on deferral. `review-panel-fix-round.md`'s
  non-convergence loop and `review-panel-late-fix.md`'s "Minors defer under the standing rule"
  switch to the inline fix.
- **Handoff:** the `**Deferred:**` line and `### Deferred minors` list are always `0`/`none` now,
  so they are removed from `verify-and-handoff.md`, `handoff-blocks.md` and
  `flow record handoff-lines`. That verb's findings read existed only for them, so it goes too,
  along with its unknown spelling.
- **Stray mentions:** `jira-followups.md`, `pipeline.md`'s "findings deferred or withdrawn",
  `operator-prompts-auto-resolution.md`'s "a Minor's disposition", `flow-self-review/SKILL.md`'s
  "every finding the review panel deferred", `README.md`'s guard line.
- `KNOWN-BUGS.md`: the `## Deferred review findings` heading and all 96 entries under it are
  deleted, unfixed. The title and the sweep entry stay.
- **Rationale:** the reversal of the Minor-deferral default is recorded in
  `skills/flow/SKILL-rationale.md`'s **review-panel.md — Panel re-runs** section. The old rule's
  recorded reasons stay where they are, as a rejected alternative's reasons.

## 5. Preventing Minors at the source

A new implementer dispatch paragraph in `skills/flow/implement.md`'s "Every implementer
dispatch **must** carry" list:

> **DRIFT CHECK:** Before you commit, read your own diff once for what it leaves behind. For every
> behaviour, name, path, flag or exit code the diff changes, grep for the comments, file headers,
> usage lines, docs and cross-file citations that describe it, and bring each one in line in this
> same commit. Every refusal or exit path the diff adds gets a test case of its own.

It targets the two biggest Minor classes above (about 62 of 96). Under `inline` execution it binds
the parent in the same words, as every implementer paragraph does (**Inline — the parent
implements**). It is not pinned in `check-dispatch-paragraphs`.

## 6. Stats: drop the deferred metrics

- Reviewers view: the `Deferred` column (`ReviewerRow.DeferredShare`, the aggregate's
  `status ILIKE 'deferred%'` count, the API's `deferredShare`).
- Run detail: the `Deferred minor` panel (KAN-508's `DeferredMinorRatio`). The run-record fetch
  in `useRunDetail` fed that panel alone, so the fetch, `fetchRunRecord` and the SPA's run-record
  DTO types go with it. The server's records endpoint stays: the CLI reads it.
- The store's `deferred` status vocabulary, categories and migration stay (**3** above).

## 7. KAN-861

KAN-861 (`kan-861-fix-all-minor-findings-in-known-bugs-md`, in flight, 6 commits) is withdrawn by
the operator in favour of deleting the backlog unfixed. This change never touches its worktrees
or its Jira issue; withdrawing it is the operator's act. This change deletes whatever Minor
entries its own base holds.

## Step-by-step breakdown

### Findings-closed guard

**What:** `check-panel-findings-closed` treats `deferred` as open and exempts a `fixed` Minor
from the clean-re-run requirement.
**Why:** it enforces "nothing is deferred" where the pipeline runs guards, and stops failing the
re-run-free Minor fix.
**Uses:** `stats/internal/guard/panelfindingsclosed.go`,
`stats/internal/guard/check_panel_findings_closed_test.go`, `scripts/check-panel-findings-closed.sh`.

### Handoff lines without Deferred

**What:** `flow record handoff-lines` prints `**Records:**` and `**Costs:**` only.
**Why:** with nothing deferred, the Deferred line and list are always `0`/`none`.
**Uses:** `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `stats/cmd/flow/main.go`.

### Dashboard without deferred metrics

**What:** remove the Reviewers `Deferred` column and the Run detail `Deferred minor` panel, end
to end (store aggregate, API field, SPA).
**Why:** they would read 0% forever. The operator chose dropping them over keeping them as an
alarm.
**Uses:** `stats/internal/store/aggregate.go`, `stats/internal/api/stats.go`, `stats/web/src/`.

### Inline Minor fix in the review panel and per-task review

**What:** the panel's and the gated reviewer's Minor-only paths fix Minors inline; the KNOWN-BUGS
close step and the deferral contract section go.
**Why:** a Minor fixed where it is raised costs a few edits; fixed later, it costs a `/flow` run.
**Uses:** `skills/flow/review-panel.md`, `skills/flow/review-panel-fix-round.md`,
`skills/flow/review-panel-late-fix.md`, `skills/flow/implement.md`,
`skills/flow-contracts/known-bugs.md`, `skills/flow-contracts/jira-followups.md`.

### DRIFT CHECK paragraph

**What:** a new implementer dispatch paragraph requiring the drift grep and a test per new exit
path, before the commit.
**Why:** it prevents the two largest Minor classes, instead of fixing them after review.
**Uses:** `skills/flow/implement.md`.

### Handoff and stray mentions

**What:** remove the Deferred handoff line and list from the skill templates; fix every
remaining deferral mention; record the reversal's rationale.
**Why:** no run-loaded text may describe a deferral that no longer exists.
**Uses:** `skills/flow/verify-and-handoff.md`, `skills/flow-contracts/handoff-blocks.md`,
`skills/flow-contracts/pipeline.md`, `skills/flow-contracts/operator-prompts-auto-resolution.md`,
`skills/flow-self-review/SKILL.md`, `README.md`, `skills/flow/SKILL-rationale.md`.

### Delete the KNOWN-BUGS Minors

**What:** delete `## Deferred review findings` and its 96 entries from `KNOWN-BUGS.md`.
**Why:** the operator chose deleting over fixing; the entries' home no longer exists.
**Uses:** `KNOWN-BUGS.md`.

### Live verification

**What:** exercise the guard, `flow record handoff-lines` and the reviewers API against the
worktree's isolated store, with before (merge-base build) and after figures.
**Why:** the guard and the CLI read the running store, and flowd's aggregate query changes.
**Uses:** the worktree's isolated `flowd` (`## workspace isolation`), `scripts/workspace.sh`,
`FLOW_ADDR`/`FLOW_RECORDS_ADDR`.

## Decisions

### The parent fixes Minors inline

**ID:** parent-fixes-minors-inline
**Status:** active
**Chosen:** the parent, inline, at the point the Minor is read — no dispatch, one commit at the tip.
**Considered:** the raising reviewer fixing its own Minors in the same dispatch — it would edit
the worktree while the next group's implementer is also editing it, and reviewers would stop
being read-only. One sonnet/low fixer per boundary — the operator picked it first and then
reversed in favour of the parent: it adds a dispatch at every boundary.

### A Minor fix has no verification of its own

**ID:** minor-fix-no-own-verification
**Status:** active
**Chosen:** no re-review, reproducer or mutation proof; the panel (per-task fixes) and
`flow.verify` (panel fixes) cover it.
**Considered:** lint plus the task's targeted tests right after each Minor commit — catches a
break earlier, but costs a test run at every boundary.

### Nothing is deferred; withdraw is the only exit

**ID:** no-deferral-withdraw-only
**Status:** active
**Chosen:** every Minor is fixed; a Minor no tree change can resolve, or that is not a defect,
is withdrawn with one clause.
**Considered:** also allowing an out-of-scope withdrawal for code the change never touched — it
reopens the escape that let 96 entries accumulate.

### Delete the KNOWN-BUGS Minors unfixed

**ID:** delete-known-bugs-minors
**Status:** active
**Chosen:** delete every Minor entry and the heading.
**Considered:** fixing them all in this change — the expensive route this change exists to end.
Triage (fix what still stands, drop the stale) — still a sizeable pass for Minors.

### DRIFT CHECK implementer paragraph, unpinned

**ID:** drift-check-paragraph
**Status:** active
**Chosen:** one new implementer paragraph, prompt text only.
**Considered:** the same paragraph pinned in `check-dispatch-paragraphs` — a Go row and a test
nobody asked for. No prevention — leaves the two largest Minor classes to review.

### Drop the dashboard's deferred metrics

**ID:** drop-deferred-metrics
**Status:** active
**Chosen:** remove the Reviewers `Deferred` column and the Run detail `Deferred minor` panel end
to end.
**Considered:** leave as-is (no code, and any non-zero value on a new run would flag a broken
rule) — the operator chose dropping. Replace with a Minor-fix count — per-task Minors never reach
the store, so it would count panel Minors only.

### KAN-861 is withdrawn, not waited on

**ID:** kan-861-withdrawn
**Status:** active
**Chosen:** the operator withdraws KAN-861; this change deletes every Minor entry on its own base.
**Considered:** treating KAN-861 as independent — the operator chose to withdraw it. Landing
KAN-861's commits first — ties this change to a big run it replaces.

### The store keeps the `deferred` vocabulary

**ID:** store-keeps-deferred-vocabulary
**Status:** active
**Chosen:** the store and CLI keep accepting and reading `deferred <reason>` and its categories;
the guard reads it as open.
**Considered:** removing it from the store, the CLI, reconcile and the API errors — historical
rows and migration `0025` carry it, and nothing writes it once the skills stop.

### The Minor rule lives inside Panel re-runs

**ID:** minor-rule-inside-panel-reruns
**Status:** active
**Chosen:** state the inline Minor fix inside `## Panel re-runs`.
**Considered:** a section of its own after it — `/flow-fast`'s range ("through **Panel
re-runs**") would miss it.

### Drop the handoff's Deferred line and list

**ID:** handoff-drops-deferred-line
**Status:** active
**Chosen:** remove `**Deferred:**` and `### Deferred minors` from the templates and
`flow record handoff-lines`, along with the findings read that fed them.
**Considered:** keeping them — they would print `0`/`none` forever.

### One Minor commit at the tip

**ID:** minor-commit-at-tip
**Status:** active
**Chosen:** one pathspec commit per worktree at the branch tip, pushed like any other.
**Considered:** folding into the task commit — `fold-fixup.sh`'s set-aside, rebase and guard
re-runs are the cost this change removes; the branch is pushed at every boundary, so the on-top
route always applies.

## Open questions
