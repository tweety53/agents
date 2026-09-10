# flow-fast speed-up — research notes

Source: none

Explored via `/flow-plan` on 2026-09-11 against `skills/flow-fast/*.md` (landed 2026-09-10,
commit 90e61be, kan-490), the contracts it cites, the dev stats store at `http://127.0.0.1:4173`,
the KAN Jira workflow and `spectre`'s own scaffold. The operator's topic: "`/flow-fast` is not fast
enough. What else can be cut? Simpler artifacts? Faster cleanup (just merge and push, fast worktree
clean, move Jira to Done)? The `flow` Go app interaction stays." The one `/flow-fast` run so far was
driven by GLM 5.3; every stage mark it made was rejected because the `/flow-fast` command was never
added to `README.md`'s Level 1 table and `stats/internal/stages` (KAN-491, in progress), so the
store holds **no** `/flow-fast` rows and this note has no timings for the command. The operator
named brainstorm+artifacts and finish as the slow phases; `/myflow-fast` data in the store is a
different command and was set aside at the operator's instruction. Every decision below is the
operator's answer to a question this session asked.

## 1. What a `/flow-fast` run actually loads and does today

- **The four fast phase files are 35 KB, but "cited, never restated" pulls in most of `/flow`.**
  Citations resolve into `skills/flow/review-panel.md` (64 KB — for three dispatch paragraphs,
  the `final-review.diff` step and the re-run mechanism), `brainstorm-planner.md` (30 KB — the
  artifact shapes), `finish-contract-run1.md` + `finish-contract-run2.md` (59 KB),
  `integrate.md`, `archive.md`, `pipeline.md`, `handoff-blocks.md` (17 KB — one block),
  `jira-integration.md`, `build-green.md`, `plan-provenance.md`. Every inline turn then carries
  that context.
- **Brainstorm runs `superpowers:brainstorming`'s full checklist** (explore, visual-companion
  offer, one-at-a-time questions, 2–3 approaches, section-by-section design approval, a
  `docs/superpowers/specs/<date>-<name>-design.md` file written and committed by the skill's own
  item 6, spec self-review, user spec review) with auto-pick, then `superpowers:writing-plans`.
  The spec file is then folded into `design.md`; two writes of the same design.
- **Artifacts are written to `/flow`'s full shape** — `design.md` decision entries with
  `ID:`/`Status:`/supersede rules, `tasks.md` with `**Files:**`, `**Tests:**`, `**Regression:**`,
  `**Baseline:**`, `**Commit:**`, `**Squash-with:**`, `**After:**`, the `**Execution:**` /
  `**Relocation:**` headers, plan-provenance tag, build-green tag — and only two things in
  `/flow-fast` read any of it: `check-unfinished-work.sh` (column-0 checkboxes) and the task loop
  (`**Files:**`, `**Tests:**`, `**Commit:**`). `spectre` itself needs `proposal.md` (`## Why`,
  `## What changes`), `design.md` (`## Context`, `## Decisions`) and at least one task line in
  `tasks.md` (`spectre validate` reports "no tasks"; `spectre archive` refuses without `--force`).
- **Finish on merge-and-push is two merges, two pushes, two landing-worktree positionings.** Run 1
  pushes the change branch, positions `_landing-<name>` on `<base>`, merges `--no-ff`, pushes base,
  removes the landing worktree. Run 2 (same invocation) positions the landing worktree again on
  `chore/archive-<name>`, `spectre archive`, `check-archive-scope.sh`, commits, pushes that branch,
  merges it into base, pushes base again, deletes the remote change branch run 1 pushed, then
  removes the landing worktree again. Plus `flow record render -kind ledger` and `-kind panel` into
  `docs/superpowers/` and the two Jira transitions (In Review, then Done) seconds apart.
- **Worktree cleanup runs six checks per worktree.** Checks 1–3 (clean tracked, no untracked,
  merged) are true by construction after commit-two and the merge in the same invocation. Check 4
  is a live operator prompt (ignored-files disclosure), and the override that lets a fast run
  report-and-proceed is still written as `/myflow-fast` in `finish-contract-run2.md`, so
  `/flow-fast` is prompted about `.superpowers/sdd/` and build output on every change. Checks 5
  (`## stop`) and 6 (`check-worktree-processes.sh`) guard the KAN-129 orphaned-ports incident and
  cost seconds.
- **Forty `flow stage` calls.** Twenty keys; ten of them (`decide`, `verify`, `stage-diff`,
  `run-instructions`, `write-in-progress`, `preflight`, `landing-question`, `preserve-sessions`,
  `verify-merge`, `sync-archive`) are mark-only in `/flow-fast`. Each begin and end is its own
  Bash turn unless chained into adjacent work.
- **Two spec contradictions.** `skills/flow-fast/brainstorm.md` section D says to run
  `check-plan-shape.sh` unconditionally while `SKILL.md`'s **Guard set** drops it.
  `finish-contract-run1.md`'s artifact-copy skip and `finish-contract-run2.md`'s check-4 override
  name `/myflow-fast`, not `/flow-fast`.
- **Jira.** KAN's workflow offers `Done` (transition 41) directly from `In Progress` (checked on
  KAN-459), so the In Review hop is not required by the board.

## 2. Decisions

Each item below was chosen by the operator from options this session presented; rejected
alternatives are recorded with the reason.

### Brainstorm — direct write, no skill invocations

Read the project context, ask **one batched `AskUserQuestion`** only for true blockers (no
recommended option, or shipping-something-wrong-if-guessed), then write `proposal.md`, `design.md`
and `tasks.md` straight into `spectre/changes/<name>/`. No `superpowers:brainstorming`, no
`superpowers:writing-plans`, no approaches round, no `docs/superpowers/specs/` file, no spec
self-review, no user spec review. The staged-research-note seed (exact-filename lookup) stays.
Rejected: keeping the brainstorming checklist with auto-pick and only dropping writing-plans — the
checklist's items 5, 6 and 8 still cost turns and a double write even with every option auto-picked;
keeping both skills — no change.

### Artifacts — minimal shape, the store is the record

- `tasks.md`: `- [ ] <n>. <title>` plus `**Files:**`, `**Tests:**`, `**Commit:**` per task — the
  three fields the task loop reads. No `**Regression:**`, `**Baseline:**`, `**Squash-with:**`,
  `**After:**`, no `**Execution:**`/`**Relocation:**` headers, no plan-provenance tag, no
  build-green tag, no `check-plan-shape.sh` (resolving the section-D contradiction in favour of the
  guard set).
- `design.md`: `## Context` and `## Decisions` as short prose bullets — no `ID:`/`Status:` entries,
  no supersede rules. `proposal.md`: `## Why` / `## What changes`, unchanged.
- No `flow record render` at finish and no `docs/superpowers/` commit: the ledger and the panel
  record live in the store, which `artifacts-registry.md` already names as their terminal record.
  `flow record dispatch`, `flow record finding`, `flow record decision` (the fixed decision) and
  `flow state set` all stay — the Go app interaction is kept in full.
- Consequence accepted: a `/flow` resume of a `/flow-fast` change needs plan repair before its
  guards pass; `flow-fast-state-interop`'s "valid record" now means the state file, stage keys and
  decision record, not the plan shape.

Rejected: minimal plan but still rendering the records — two copies of a record the store already
holds; keeping the full `/flow` shape — the shape is written for guards `/flow-fast` never runs.

### Finish — one merge, one push, one landing worktree, no run 2

`/flow-fast` lands by **merge-and-push only**. The landing question is never asked and the project's
`## default landing route` is not read; a change that needs a pull request is a `/flow` run. The
bare `IN_PROGRESS` invocation is one run:

1. `check-finish-preflight.sh`, `check-unfinished-work.sh`, `check-base-moved.sh` — kept, all
   three, as one chained Bash call; they prompt only on `OUTSTANDING:` or an overlapping `MOVED`.
   Rejected: dropping `check-base-moved.sh` — a conflict then surfaces inside the merge instead of
   in front of it.
2. `git reset --soft <recorded-merge-base>` and `commit-split.sh`'s two commits (implementation,
   then `chore(spectre): plan and session records` carrying only `spectre/changes/<name>/`).
3. `prepare-archive-branch.sh <project>/.worktrees/_landing-<name> <base> <base>`; `git merge
   --no-ff spectre/<name>` in the landing worktree. The change branch is **not pushed**.
4. `spectre archive <name>` **in the landing worktree, on `<base>`**, `check-archive-scope.sh`,
   commit. No `chore/archive-<name>` branch, no second positioning, no verify-merge step.
5. `git push origin <base>` — once. A merge conflict or a refused push stops the run and reports
   the landing worktree path.
6. Cleanup (below), workspace `remove` command via `flow workspace-id <name>`, remove the landing
   worktree, write `FINISHED`, Jira → Done.

Rejected: keeping all three routes with merge-and-push collapsed — PR/manual are what keep a
standalone run 2 and the archive branch alive; keeping two runs but reusing the landing worktree —
still a second branch and a second push.

### Cleanup — stop and process scan only

Checks 1–4 are skipped: 1–3 were established by this invocation (commit-two left the tree clean,
the merge made HEAD an ancestor of `<base>`), and 4's disclosure prompt goes with them — the
`.superpowers/sdd/` records it would disclose are in the store. Check 5 (`## stop`, 60-second
bound) and check 6 (`check-worktree-processes.sh`, cwd outside every worktree) stay as gates:
`HELD:` or exit 2 stops at `IN_PROGRESS` with the pids. Then `git worktree remove --force`, `git
branch -d spectre/<name>`, `git worktree prune`. No remote branch delete — nothing was pushed. The
project's workspace `remove` command still runs (gymie's cache-index pool has a ceiling of fifteen).
`check-cleanup-complete.sh` stays skipped as before.

Rejected: dropping only the check-4 prompt — checks 1–3 would still re-prove the invocation's own
work; keeping all six — no change.

### Jira — straight to Done

`In Progress` at kickoff (unchanged). After the push, one transition to `Done` — no `In Review`
hop, no second transitions read. The forward-only rule and never-blocking rule from
`jira-integration.md` still apply; a project whose workflow has no direct transition gets the
skipped-with-reason line. `/flow`'s In Review timing is untouched.

### Stage marks — every key kept, every call chained

All twenty `flow.*` keys stay so the Go app's stage rows are complete. Each stage's `end` mark and
the next stage's `begin` ride in the same Bash call as the next stage's first command; mark-only
stages are one call carrying both marks. Zero dedicated turns. The literal-token and `-harness`
rules are unchanged. Depends on KAN-491 landing (the `/flow-fast` command in the stage vocabulary).

Rejected: fewer keys for `/flow-fast` — loses leaderboard rows the operator wants; leaving 40
separate calls — no change.

### Review panel — shared dispatch file, simple-reviewer-only delta re-run

The three dispatch paragraphs (INDEPENDENT PASSES, TOOLS, MODEL HANDSHAKE) and the
`final-review.diff` step move to a ~40-line `skills/flow/panel-dispatch.md` cited by both
`skills/flow/review-panel.md` and `skills/flow-fast/review.md`; `/flow-fast` never loads
`review-panel.md`. After a Critical/Major fix, only the `simple-reviewer` (haiku) slot re-runs on
the round's delta; `primary` keeps its pass-1 plan-alignment verdict. `check-panel-findings-closed.sh`
still gates the handoff. The kan-490 rejection of "re-run `primary` alone" is not reopened — this
re-runs the defect slot, not the alignment slot.

Rejected: extract only, both slots re-running — `primary`'s job is plan alignment and a fix for a
defect does not change it; leaving the panel as is — no change.

### Handoff — carry the literal block

`skills/flow-fast/review.md` carries its own `IN_PROGRESS` block and run instructions verbatim
instead of citing `handoff-blocks.md` (17 KB) — the same "carries only the block it prints" rule
`pipeline.md` already grants `/flow`.

### Spec fixes folded into the same change

`brainstorm.md` section D's `check-plan-shape.sh` sentence goes; the `/myflow-fast` names in
`finish-contract-run1.md` (artifact-copy skip) and `finish-contract-run2.md` (check-4 override)
become `/flow-fast` or are deleted where the step no longer applies to it.

## 3. Implementation touchpoints the decisions imply

Derived from the decisions above and the repository's own guards; none is a new decision.

- **`skills/flow-fast/finish.md` becomes its own procedure.** It can no longer follow
  `finish-contract-run1.md`/`run2.md` "exactly": it states the one-run order of section 2 itself,
  citing only the shared mechanics by their canonical homes — the reshape (`reset --soft`) and
  `commit-split.sh` chain (`git-boundaries.md`), the preflight/unfinished-work/base-moved verdict
  tables (`finish-contract-run1.md`), `prepare-archive-branch.sh`'s and `check-archive-scope.sh`'s
  own headers, and cleanup checks 5–6 (`finish-contract-run2.md` **Worktree cleanup**). The
  `**Deciding which run this is**` section goes: there is one run, and a `RUN2` preflight verdict on
  a `/flow-fast` change is a wrong-state handoff, not a branch.
- **Stage keys, one-run mapping** (`SKILL.md` **Stage keys** table): `flow.preflight` (the three
  guards), `flow.unfinished-work-gate` and `flow.landing-question` (mark-only — never asked),
  `flow.preserve-sessions` (mark-only — nothing rendered), `flow.commit-two`, `flow.landing-routes`
  (landing worktree + merge), `flow.verify-merge` (mark-only), `flow.sync-archive` (`spectre
  archive`), `flow.commit-archive`, `flow.cleanup`, `flow.write-finished`, `flow.push-archive` (the
  one push, then Jira Done). All twenty keys still marked; `README.md` Level 1 and
  `stats/internal/stages` need nothing beyond KAN-491.
- **Guard set** (`SKILL.md` **Guard set** and `skills/flow-fast/scripts/` symlinks):
  `check-archive-scope.sh` joins (it runs at the archive commit); `check-cleanup-complete.sh` leaves
  (it was never run). The set becomes `check-unfinished-work.sh`, `check-base-moved.sh`,
  `check-finish-preflight.sh`, `check-archive-scope.sh`, `check-workspace-isolation.sh`,
  `check-worktree-processes.sh`, `check-panel-findings-closed.sh`.
- **`artifacts-registry.md` rows that no longer apply to `/flow-fast`** need a per-command note
  where the row says who removes what: *Remote branch* (never pushed), *Archive branch* (never
  created), *Rendered ledger and panel record* (never rendered). The registry's own rule — an
  artifact no row accounts for is a defect — cuts both ways, so the rows say the command has none
  rather than staying silent.
- **`jira-integration.md` Transitions table** gains a `/flow-fast` row: `In Progress` at kickoff,
  `Done` after the push; the forward-only and never-blocking rules apply unchanged.
- **Repository guards the change must pass:** `scripts/check-contract-budget.sh` carries a byte
  ceiling per skill file and needs a row for the new `skills/flow/panel-dispatch.md` (and the
  reduced `flow-fast/*.md` sizes); `scripts/check-stage-mark-calls.sh` lists
  `skills/flow-fast/SKILL.md` and reads the phase files' mark calls — chained marks are still
  literal `flow stage` calls, so the shape it checks is unchanged; `scripts/check-references.sh`
  polices every citation moved by the panel-dispatch extraction and the finish rewrite.
- **`check-unfinished-work.sh` parses no task fields** — only checkboxes (verified in
  `scripts/lib/change-plan.sh`), so the minimal `tasks.md` shape passes it as is.
- **The archived kan-490 `design.md` is frozen.** The new change's `design.md` records each
  superseded decision (`flow-fast-brainstorm`, `flow-fast-finish`, `flow-fast-guards`,
  `flow-fast-review`, the plan-shape half of `flow-fast-implement`) by id, per the supersede rule.

## Step-by-step breakdown

### Brainstorm — direct write

**What:** Read context, one batched blocker question, then write `proposal.md`, `design.md`,
`tasks.md` directly; seed from a staged research note when one exists.
**Why:** The brainstorming checklist's approvals, approaches round and spec file cost turns and a
double write on a command whose only human gate is the staged diff.
**Uses:** `skills/flow-fast/brainstorm.md`, `spectre new`, `AskUserQuestion`,
`docs/superpowers/research/<name-or-key>.md` seed lookup, `flow stage` marks `flow.kickoff`,
`flow.brainstorm`, `flow.create-artifacts`, `flow.writing-plans`, `flow.decide`.

### Minimal artifacts

**What:** `tasks.md` = checkbox + `**Files:**`/`**Tests:**`/`**Commit:**`; `design.md` = prose
`## Context` / `## Decisions`; no rendered ledger/panel, no `docs/superpowers/` commit.
**Why:** Nothing `/flow-fast` runs reads the other fields; the store is the terminal record for the
ledger and panel.
**Uses:** `skills/flow-fast/brainstorm.md` section C/D, `scripts/check-unfinished-work.sh`,
`spectre validate`, `flow record decision`, `<abs-worktree>/.superpowers/sdd/decision.json`.

### One-run finish

**What:** Guards → reset/commit-split → landing worktree on `<base>` → merge `--no-ff` → `spectre
archive` on `<base>` → `check-archive-scope.sh` → commit → one push → cleanup → `FINISHED` → Jira
Done → remove landing worktree.
**Why:** The two-run shape exists for routes where the merge happens elsewhere; merge-and-push in
one invocation makes the second landing, the archive branch, the second merge, the second push and
the remote-branch delete redundant.
**Uses:** `skills/flow-fast/finish.md`, `scripts/check-finish-preflight.sh`,
`scripts/check-unfinished-work.sh`, `scripts/check-base-moved.sh`, `scripts/commit-split.sh`,
`scripts/prepare-archive-branch.sh`, `scripts/check-archive-scope.sh`, `spectre archive`,
`scripts/resolve-base-branch.sh`, `flow state set`, `flow stage` marks `flow.preflight` through
`flow.push-archive`.

### Cleanup — stop + process scan

**What:** `## stop` (bounded), `check-worktree-processes.sh`, then `worktree remove --force`,
`branch -d`, `prune`, the project's workspace `remove` command.
**Why:** Checks 1–3 restate what this invocation proved; check 4's prompt discloses records the
store holds; 5 and 6 guard a real incident and cost seconds.
**Uses:** `skills/flow-contracts/finish-contract-run2.md` **Worktree cleanup** (checks 5–6 only),
`scripts/check-worktree-processes.sh`, `flow workspace-id`, `<project>/.flow/project.md` `## stop`
and `## workspace isolation` `remove`.

### Jira straight to Done

**What:** One `Done` transition after the push; `In Progress` at kickoff unchanged.
**Why:** KAN offers `Done` from `In Progress`; In Review marks a wait this route never has.
**Uses:** `skills/flow-contracts/jira-integration.md` **Transitions** (a `/flow-fast` row),
`mcp__jira__jira_get_transitions`, `mcp__jira__jira_transition_issue`.

### Chained stage marks

**What:** Every `flow stage end` + next `begin` in the same Bash call as the next stage's first
command.
**Why:** Keeps all twenty rows for the Go app at zero dedicated turns.
**Uses:** `flow stage begin/end`, every `skills/flow-fast/*.md` phase file, KAN-491 (vocabulary).

### Shared panel-dispatch file, narrower re-run

**What:** `skills/flow/panel-dispatch.md` with the three dispatch paragraphs and the
`final-review.diff` step; `/flow-fast` re-runs `simple-reviewer` only on a fix delta.
**Why:** Removes the 64 KB `review-panel.md` load from every fast turn; `primary`'s pass-1 verdict
is about plan alignment, which a defect fix does not move.
**Uses:** `skills/flow/review-panel.md`, `skills/flow-fast/review.md`,
`skills/flow/simple-reviewer-prompt.md`, `scripts/check-panel-findings-closed.sh`,
`scripts/check-references.sh`, `flow record dispatch`, `flow record finding`.

### Literal handoff block

**What:** `review.md` carries the `IN_PROGRESS` block and run instructions verbatim.
**Why:** Avoids loading `handoff-blocks.md` for one block, on the rule `/flow` already has.
**Uses:** `skills/flow-fast/review.md`, `skills/flow-contracts/pipeline.md` **Handoff output**.

### Spec fixes

**What:** Drop the `check-plan-shape.sh` sentence in `brainstorm.md` D; rename or delete the
`/myflow-fast` references in both finish contracts.
**Why:** Two contradictions a run would otherwise resolve differently each time.
**Uses:** `skills/flow-fast/brainstorm.md`, `skills/flow-contracts/finish-contract-run1.md`,
`skills/flow-contracts/finish-contract-run2.md`, `scripts/check-references.sh`.

## Open / undesigned

None — every item this session surfaced was decided above. Timing evidence for `/flow-fast`
itself arrives with the first run after KAN-491 lands.

## Decision

Plan: `docs/superpowers/research/flow-fast-speedup/tasks.md` (`check-plan-shape.sh` clean).
Decision: `docs/superpowers/research/flow-fast-speedup/decision.json`.

| Setting | Toggle | Result |
|---|---|---|
| execution mode | dynamic | inline |
| implementer model | dynamic | skipped — inline |
| review panel | dynamic | compact: primary sonnet/medium, simple-reviewer haiku/medium, principles haiku/medium + exp-failure-modes sonnet/medium; delta rerun; dispatches: primary+simple-reviewer+principles · exp-failure-modes |
| implementer groups | — | skipped — inline |

grouping: free — bundle roll 67 ≥ 30; the floor bundle fills the first dispatch and the experimental slot is the whole second
class: small (mechanical: small; override: none)
inputs: tasks=9 files=13 repos=1 migration=no spec=no red=no unverified=no
rolls: compact 58 (< 90 — compact) · experimental 10 (< 30 — `failure-modes.md`, index 10 mod 1) · bundle 67 (≥ 30 — free grouping)
