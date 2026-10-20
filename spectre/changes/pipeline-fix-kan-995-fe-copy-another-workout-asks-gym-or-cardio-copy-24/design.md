## Context

An in-run pipeline fix (**Pipeline defects found mid-run**, `skills/flow-contracts/pipeline.md`) for
gymie's kan-995 run, whose final report ended on two "still open" items the operator then had to
order done (`proposal.md`). Both trace to `skills/flow-self-review/SKILL.md`:

- Step 2 classifies "an observation with nothing to fix" as `noted`, and the pass put an observation
  that did name an action — merge develop into `feature/cardio` — under it, worded "a branch
  decision for the operator".
- Step 3 forbids dropping a non-`big` finding, but a planner that fixes the defect forward and
  leaves the rows it already corrupted is not dropping anything: decision `k995-rows-left` did that,
  and the operator had to ask for the repair (`b2555b98`, `flowd restamp-dispatch`).

The general rule the operator asked for — a finding naming a concrete action the run can take, not
`big`, not production, not an irreversible deletion, is done in the run; the final report carries
no "still open" or "your call" item but a genuine blocker — is already half-stated:
`rules/agent-baseline.md` ("Every defect you notice is fixed"), **Auto-resolution**
(`skills/flow-contracts/operator-prompts-auto-resolution.md`) for prompts, and
`skills/flow/cleanup.md`'s "carry no open item". The gap is the classification step that feeds
them, so each rule lands once, at the step that decides it.

Blast radius, counted per **Pipeline defects found mid-run**: the self-review skill, its land guard
(`stats/internal/guard/landselfreviewreport.go`, its test, the `scripts/land-self-review-report.sh`
header), the two files citing `LAND-UNTRACED-FINDING` (`skills/flow/integrate.md`,
`skills/flow-fast/SKILL.md`) and `skills/flow/cleanup.md` — 7 files, not `big`.
<!-- measured: grep -rln 'LAND-UNTRACED\|lsrrUntraced' stats scripts skills rules docs plus the three prose files @ b2555b98 -->


## Decisions

### An observation naming an action is a finding, ruled in step 2

**ID:** observation-with-action-is-a-finding
**Status:** active
**Chosen:** one sentence in step 2 of `skills/flow-self-review/SKILL.md`, beside the `noted` rule it narrows: an observation that names an action the run can take is a fixable finding, never `noted`, never an operator decision, unless the action is `big`, reaches production or deletes what cannot be restored — the classification is where kan-995's item went wrong, and step 3 already fixes every fixable finding.
**Considered:** a sentence in `skills/flow-contracts/operator-prompts-auto-resolution.md` — it governs prompts the run frames, and the pass framed no prompt, it recorded `noted`; a sentence in `skills/flow/cleanup.md` alone — it fires at the end of the run, after the pass already closed the finding as `noted` and no loop remains to fix it.

### A fix repairs what its defect already corrupted

**ID:** fix-repairs-corrupted-data
**Status:** active
**Chosen:** one sentence in step 3's planner paragraph of `skills/flow-self-review/SKILL.md`, right after the no-drop rule — **Pipeline defects found mid-run** cites step 3 as canonical for its planner, so the one sentence binds both loops.
**Considered:** a sentence in `skills/flow-contracts/pipeline.md` — it would bind the in-run loop only and restate a planner rule step 3 owns; folding it into the no-drop rule's own sentence — that sentence is about a design objection, and `k995-rows-left` raised none.

### A base missing a fix is synced by this session

**ID:** base-sync-by-session
**Status:** active
**Chosen:** a **The base sync** paragraph after item 5 of step 3: this session merges `origin/<default-branch>` into `<base>` in a detached worktree, runs that repository's declared `## lint` and `## test`, and on a clean merge and green run pushes the merge to `<base>` without asking, recording the finding `fixed` with the merge sha; a conflict or a red run pushes nothing and is offered in step 4 as a `big` finding is. The step-3 sentence "this session plans, changes and reviews nothing itself" names it as its second exception.
**Considered:** a task in the loop's plan for a fixer — the plan is a spectre change in `<agents repo>`, and a merge commit on another repository's base is no `**Files:**` list `check-task-commit-fields.sh` can attribute; syncing every `--base` at each run's kickoff — merges the default branch into a feature branch on every run, more than the operator asked; resolving a conflicting merge automatically — a judgement over another branch's work, which the operator's rule scoped out ("when it merges cleanly and verifies green"); asking the operator — the defect itself.

### The land guard refuses a deferred `noted` finding

**ID:** deferred-finding-guard
**Status:** active
**Chosen:** extend `lsrrUntraced` in `stats/internal/guard/landselfreviewreport.go`: a finding line (`- **[`) whose trimmed text ends `— noted` and that says `for the operator`, `operator decision`, `operator's call`, `operator's choice`, `your call` or `if you want` (case-insensitive) prints `LAND-DEFERRED-FINDING: <path>:<line>: <line> — fix it in the run, never leave it to the operator; nothing added, committed, pulled or pushed` and the run exits 4 — the code the untraced refusal already uses for "an open finding, nothing landed", so every caller's handling of exit 4 holds unchanged. Measured at plan time over every landed report — 188 in `<agents repo>` at `b2555b98`, gymie `origin/develop` (72) and `origin/feature/cardio` — the `noted`-restricted match hits exactly kan-995's line 15.
**Considered:** the check in `scripts/check-self-review-report.sh` — a shape guard run as lint, which runs after the report is already written and whose failure is not wired to stop the landing; matching every disposition — 3 false positives measured (kan-77 `filed`, kan-170 `declined`, kan-916 `filed`), whose wording describes the history, not a deferral; a guard over a planner's `design.md` for deferral wording — `k995-rows-left`'s text ("the fix is forward only") names no operator at all, so no phrase list catches it without flagging every forward-only decision: the rule is a wording fix (`fix-repairs-corrupted-data`); a new exit code — every caller already treats 4 as "resolve and land again".

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-medium` — Task 1 adds a refusal to the landing chain every self-review report passes through, and Task 2's base sync pushes to a shared branch without asking; a slip in either blocks every later landing or pushes what it should not.
**Considered:** `flow-low` — both tasks are small and test- or guard-pinned, but both sit on every run's path.

### Fixer grouping

**ID:** fixer-groups
**Status:** active
**Chosen:** one fixer dispatch, `fix`, Tasks 1 then 2 — both edit `skills/flow-self-review/SKILL.md`, and Task 2 is After Task 1.
**Considered:** one fixer per task — two dispatches editing the same file in series for a seven-file change.

## Open questions

## Live check

No running service or store is touched; the guard is checked against the real report, read-only.
In a throwaway repository (`git init -q <scratch>/g`, then
`git -C <scratch>/g fetch -q /Users/tweety53/Projects/gymie refs/remotes/origin/feature/cardio:refs/heads/feature/cardio`
and `git -C <scratch>/g checkout -q feature/cardio`),
`scripts/land-self-review-report.sh <scratch>/g feature/cardio "x" docs/self-review/kan-995-fe-copy-another-workout-asks-gym-or-cardio-copy-self-review.md`
from this worktree exits 4 with exactly one `LAND-DEFERRED-FINDING: …:15:` line and leaves the
clone's `HEAD` and index unchanged. Not working looks like: exit 0 and `LAND-NOTHING-TO-COMMIT`,
measured at plan time before the fix.
<!-- measured: the command above against the fetched feature/cardio @ b2555b98 printed LAND-NOTHING-TO-COMMIT, rc=0 -->

