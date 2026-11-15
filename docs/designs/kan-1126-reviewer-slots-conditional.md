# kan-1126-reviewer-slots-conditional
> Design record of change kan-1126-reviewer-slots-conditional (KAN-1126), written at landing. Each decision's own **Status:** line is the status.

## Context

KAN-1126 decided the four conditions the proposal lists. The design was settled over three
planning rounds:

- Round 1, Q1 [architectural] How does panel-slot-triggers decide call-efficiency and
  failure-modes? — operator: "A. Built-in rules, changed lines only (Recommended)".
- Round 2, operator message mid-planning — "make it also llm powered on the stage of planning. It
  is pretty cheap".
- Round 3, convergence confirm — "Nothing unclear — approve the design and move on".

Constraints:

- Decide (`skills/flow/brainstorm-planner.md`) runs before any diff exists, so a diff-based
  verdict can only be taken at review pass 1; a plan-based one only at Decide. Fix runs never
  re-plan.
- Every existing panel reduction (docs-only, late-fix, small append, fix-run judge) only removes,
  and only on a definite answer; an unanswered question never reduces a panel. The new reductions
  keep both properties.
- The extra dispatches' pairs (`extra_dispatch`, `extra_dispatch_mutation`) are recorded by Decide
  on every class and read by `records.go`, `aggregate.go` and the stats; their shape does not
  change.
- New logic is Go in `flow-guard` (`rules/no-bash-scripts.mdc`).

It is one change because the guard, the judge, the panel prose that combines them and the bundle
matchers that the shrinking bundles make live only work together: either source alone, or the
combination without the matcher fixes, leaves a panel that dispatches the wrong pair or never
gates.

## Recommended approach

**1. `flow-guard panel-slot-triggers <worktree> <merge-base> [<worktree> <merge-base>…]`** in
`stats/internal/guard/panelslottriggers.go`, registered in `guard.Registry` by its own `init()`, its test reusing the
temporary-repo fixtures of `check_panel_docs_only_test.go`.

- Inputs per worktree: `panelTouchedPaths`, so `## panel diff exclude` already applies; and one
  `git diff --no-renames -U0 --end-of-options <merge-base>` call with `wpdExcludes`, which against
  the working tree covers committed, staged and unstaged changes in one go.
- Output is exactly three lines, in this fixed order:
  - `test-sanity: run|skip — <reason>`
  - `call-efficiency: run|skip — <reason>`
  - `failure-modes: run|skip — <reason>`
- `<reason>` is the first match (`<path>:<line> <rule>`), or `no <lens> signal in the changed
  code`.
- Exit 0: answered. Exit 2: cannot answer, and every gated slot runs. There is no exit 1, since
  three verdicts do not fit one bit.
- A slot runs when any worktree prints `run` for it.

**2. test-sanity** runs when a touched path (added, changed or deleted) is a test path:

- `_test\.go$`
- `(Test|Tests|Spec|IT)\.(kt|java)$`
- `(^|/)src/[A-Za-z]*[Tt]est[A-Za-z]*/`
- `\.(test|spec)\.[cm]?[jt]sx?$`
- `(^|/)(tests?|__tests__|e2e|testdata|fixtures?)/`
- `__snapshots__/`, `-snapshots/`
- `{jest,vitest,playwright}.config.*`

**3. call-efficiency and failure-modes** are decided by built-in content and path rules over
changed lines only (`-U0`, added and removed lines). There is no project configuration.

- Doc paths (`isDocPath`), lockfiles and test paths are excluded.
- A deleted call still counts.
- call-efficiency runs on:
  - handler, query and client tokens: `ServeHTTP|HandleFunc|*http.Request`,
    `@*Mapping|@RestController|routing {`, Next.js `export function GET…|NextResponse`,
    `.Query|QueryRow|Exec(Context)?(`, SQL verbs, `@Query`, `http.(Get|Post|NewRequest)`,
    `fetch(`, `axios`, `HttpClient`, `useQuery|useMutation`;
  - a changed path under a handler, store or client directory:
    `api|handlers?|controllers?|routes?|presentation|store|persistence|repositor(y|ies)|clients?|integrations`,
    or a `.sql` path.
- failure-modes runs on everything call-efficiency matches, plus:
  - `os.(Open|Create|ReadFile|WriteFile|Rename|Remove|MkdirAll)`, `exec.Command`;
  - `go func`, `sync.`, `chan`, `<-`, `select {`, `context.With`, `atomic.`;
  - Kotlin `launch|async|withContext|Mutex|suspend fun`;
  - TS `await|Promise.|setTimeout|localStorage|fs.`;
  - `scheduler|jobs?|workers?|migrations` directories.
- Known ceiling: a query or call reached through an unnamed project helper outside the listed
  directories (an `agents` guard edit adding a second `capture(panelGit(…))`, a composable
  calling a repository method) is invisible to the guard. The slot judge (item 10) closes it where
  the plan names the work; the slot still skips when neither sees it. The live check (item 9)
  measures it; a skip on a finding's own slot is a missing rule, added before landing.

**4. Where it runs.** `skills/flow/review-panel.md`, in the pass-1 entry Bash call right after
`check-panel-docs-only`.

- A new `### The slot triggers` section sits under the extra-dispatches section, which loses
  "always-on" and the sentence "No other reduction removes either".
- The "only automatic reductions" paragraph names it, and every reduction still only removes.
- Order of application:
  1. docs-only
  2. late-fix / small append
  3. slot triggers — a gated slot is skipped only when both the guard and the plan-time slot
     judge (item 10) say skip
  4. fix-run judge
  5. operator-named slots, which always run
- Each skipped role is recorded as `flow record pass -round <r> -note 'not dispatched — slot
  trigger: <slot> — guard: <guard reason>; judge: <judge reason>'`. The guard's three lines are
  recorded on every run.

**5. Bundles** reuse the fix-run judge's rule verbatim: a bundle dispatches the roles it still
names, on its own pair, and a bundle naming none is not dispatched.

- `test-sanity+call-efficiency` can become `test-sanity`, `call-efficiency`, or nothing.
- `mutation+failure-modes` can become `mutation` alone. Mutation is never gated.
- Two latent defects become live and are fixed:
  - `renderslotprompt.go` `rspRole` maps a lone `test-sanity` or `call-efficiency` to
    `panel-reviewer`; it maps them to `panel-sanity-efficiency`.
  - `panelroundbegins.go:113-116` holds only the literal `mutation+failure-modes` to the later
    begin; any dispatch whose roles are within {mutation, failure-modes} and include mutation is
    held.

**6. principles.** `pcTree` (`stats/internal/guard/planclass.go`) keys on the effective class.

- On `small`, roster and dispatches are `primary` alone.
- A planner raise to `regular` brings principles back.
- Micro is unchanged.
- The `extra:` lines and Decide's `extra_dispatch` / `extra_dispatch_mutation` pairs stay on every
  class, so `records.go`, `aggregate.go` and stats need no change. pair-judge changes only by the
  extraction item 10 names.

**7. Fix rounds.** `review-panel-fix-round.md` re-runs the guard beside its docs-only re-check.

- A gated slot that now fires on the whole branch joins that round, reading the whole
  `final-review.diff`.
- On fix runs (`review-panel-late-fix.md`) the guard reads the whole branch, and the fix-run
  judge's input list becomes the post-trigger roster.

**8. Prose.** Edited to match:

- `skills/flow/review-panel.md`, `skills/flow/review-panel-fix-round.md`,
  `skills/flow/review-panel-late-fix.md`, `skills/flow/review-panel-optional-slots.md:3`,
  `skills/flow/verify-fix-loop.md:141` (a citation of the renamed section);
- `skills/flow/brainstorm-planner.md` Decide steps 3–4, the new Decide step 7 (item 10), and the
  `slot_triggers` field in the decision JSON paragraph;
- `skills/flow-fast/SKILL.md:110,119,122` — its decide line runs step 7 too;
- the headers of `skills/flow/test-sanity-reviewer-prompt.md` and
  `skills/flow/call-efficiency-reviewer-prompt.md`;
- `skills/flow-contracts/model-policy.md:123-125`;
- `skills/flow/SKILL-rationale.md`;
- `README.md`'s rosters table and its "always-on" bullet;
- `stats/internal/repotest/reviewers_never_add_a_worktree_test.go:51-69`, which string-matches
  the old heading and sentence and is updated with them.

**9. Live check (`flow.verify`).** Read every Critical/Important call-efficiency, failure-modes
and test-sanity finding since 2026-10-01 from the local dev flowd store.

- Run the guard over each finding's change range on a `live-store-copy`.
- Run `slot-judge` over that change's archived `tasks.md` where one exists; where none exists the
  judge counts as run, which is what the panel would do.
- Record how many each would have skipped alone, and how many the combination skips.
- A combined skip on a finding's own slot is a missing rule, fixed before landing.

**10. `flow-guard slot-judge <digest-file>`** in `stats/internal/guard/slotjudge.go`, registered
in `guard.Registry` by its own `init()`, tested beside it through the `FLOW_PAIR_JUDGE_CMD` seam as
`pairjudge_test.go` does.

- **Shared call.** pair-judge's run of the judge moves out of `pjAsk` into one function in
  `pairjudge.go` that takes a system prompt and a digest and returns the answer text, the spend
  suffix, or a cause. pair-judge and slot-judge both call it. Unchanged: the model is the
  `pair-judge` row's first base pair (no new policy row); the `claude -p` flags; the
  `FLOW_PAIR_JUDGE_CMD` seam; the 60 s deadline (`Env.PairJudgeTimeout`); the 200 KB digest cap;
  the ` (judge <s>s, $<cost>)` spend suffix.
- **Prompt.** A fixed template naming the three slots, each with the ticket's condition verbatim.
  It says the digest is a plan written before any code exists and that a slot runs unless the plan
  clearly does no such work, and asks for exactly three lines, `<slot>: run|skip — <one line>`.
- **Output.** `panel-slot-triggers`'s three lines in the same order and format. Each reason ends
  with `judge` or `fallback: <cause>`, then the spend suffix whenever the command ran.
- **Fallback.** Exit 0 whatever the judge did. Any cause pair-judge falls back on prints all three
  as `run`; so does an answer missing a line, carrying a slot twice, or carrying a verdict other
  than `run|skip`. Exit 2, nothing on stdout, only on usage or an unreadable digest.
- **Where it runs.** A new Decide step 7 in `brainstorm-planner.md`, on every class, micro
  included, and on a fix run's Decide. The digest is the plan's tasks and the `inputs:` line, the
  same digest as Decide's other calls, written to
  `<abs-worktree>/.superpowers/sdd/pair-judge/decide-slots.md`.
- **Recorded.** `decision.json` gains
  `slot_triggers`: `{"test-sanity": {"verdict": "run|skip", "reason": "…"}, "call-efficiency": {…}, "failure-modes": {…}}`.
  `flow decision render` (`stats/cmd/flow/decision.go`) prints one line, `Slot judge`, listing
  each slot as `run` or `skip`; the field is optional, so older decisions still render.
  `records.go` stores the decision JSON as-is, so flowd needs no change.
- **Combination at panel time** (item 4, step 3). The panel reads `slot_triggers` from the run's
  `decision.json`. A gated slot is skipped only when the judge's verdict and every worktree's
  guard line both say `skip`; either `run` runs it. A missing `slot_triggers` — an older decision,
  a no-decision panel, an unreadable file — counts as `run` for all three. Guard exit 2 still
  runs every gated slot. So the judge only ever adds a slot back, and every reduction still only
  removes.
- **Fix rounds** (item 7). The plan-time verdict is not re-asked; a slot the guard fires on the
  whole branch joins the round as before, and a slot the judge said `run` already ran at pass 1.
- **Not judged.** principles stays keyed on the effective class (item 6); mutation is never gated.

Rejected alternatives, one line each: built-in rules widened to the enclosing function plus
project-declared triggers over-trigger in Go and Kotlin (`trigger-rules-builtin-changed-lines`);
project-declared globs only keep the always-on cost everywhere until declared
(`trigger-rules-builtin-changed-lines`); plan-time triggers from `**Files:**` predate the diff
(`triggers-at-panel-time`); the judge alone, the guard alone, and skip-on-either each fail
(`slot-judge-union`); a panel-time LLM judge is not what was asked (`slot-judge-at-plan-time`); a
new policy row needs a contract row and migration for one call (`slot-judge-reuses-pair-judge-call`).

## Critical files

- `stats/internal/guard/panelslottriggers.go` — new guard: the three diff-based verdicts.
- `stats/internal/guard/panelslottriggers_test.go` — its tests on temporary repositories.
- `stats/internal/guard/slotjudge.go` — new guard: the plan-time LLM verdicts.
- `stats/internal/guard/slotjudge_test.go` — its tests through `FLOW_PAIR_JUDGE_CMD`.
- `stats/internal/guard/pairjudge.go` — the judge call extracted from `pjAsk` into a shared function.
- `stats/internal/guard/pairjudge_test.go` — keeps pair-judge's behaviour pinned across the extraction.
- `stats/internal/guard/planclass.go` — `pcTree` drops principles on the effective class `small`.
- `stats/internal/guard/renderslotprompt.go` — `rspRole` maps a lone `test-sanity` or `call-efficiency`.
- `stats/internal/guard/panelroundbegins.go` — later-begin hold for any mutation-carrying sub-bundle.
- `stats/cmd/flow/decision.go` — renders the optional `slot_triggers` as a `Slot judge` line.
- `stats/internal/repotest/reviewers_never_add_a_worktree_test.go` — string-matches the renamed heading.
- `skills/flow/review-panel.md` — `### The slot triggers`, the reduction order, the skip notes.
- `skills/flow/review-panel-fix-round.md` — re-runs the guard on the whole branch.
- `skills/flow/review-panel-late-fix.md` — guard on the whole branch; fix-run judge reads the post-trigger roster.
- `skills/flow/review-panel-optional-slots.md` — line 3's always-on wording.
- `skills/flow/verify-fix-loop.md` — line 141 cites the renamed section.
- `skills/flow/brainstorm-planner.md` — Decide steps 3–4, new step 7, `slot_triggers` in the decision JSON.
- `skills/flow-fast/SKILL.md` — lines 110, 119 and 122.
- `skills/flow/test-sanity-reviewer-prompt.md`, `skills/flow/call-efficiency-reviewer-prompt.md` — headers.
- `skills/flow-contracts/model-policy.md` — lines 123–125.
- `skills/flow/SKILL-rationale.md` — the reasoning for the conditional slots.
- `README.md` — rosters table and the extra-dispatch bullet.

## Existing code to reuse

- `panelTouchedPaths` (`stats/internal/guard/paneltouchedpaths.go`) — touched paths with `## panel diff exclude` applied.
- `wpdExcludes`, `smcAbs`, `panelGit`, `capture` — the excluded-path diff call pattern used by `paneltouchedpaths.go` and `planclass.go`.
- `isDocPath` (`stats/internal/guard/planclass.go`) — doc-path exclusion.
- `check_panel_docs_only_test.go` — temporary-repo fixtures for the guard's tests.
- `pjAsk` and its `claude -p` call, `FLOW_PAIR_JUDGE_CMD`, `Env.PairJudgeTimeout`, the 200 KB cap and spend suffix (`stats/internal/guard/pairjudge.go`) — extracted and shared.
- `pairjudge_test.go`'s seam usage — the pattern for `slotjudge_test.go`.
- The fix-run judge's bundle rule (`skills/flow/review-panel-late-fix.md`) — reused verbatim for bundle shrinking.
- `flow record pass -note` — the existing not-dispatched record.

## Verification

- `cd stats && make -s web-dist && go vet ./... && gofmt -l .` clean, and `go test ./...` green.
- Every `scripts/flow-guard check-*` guard in `.flow/project.md`'s `## lint` green.
- `scripts/flow-guard panel-slot-triggers <worktree> <merge-base>` over this change's own branch
  prints `test-sanity: run` (it adds `_test.go` files) and `failure-modes: run` (it moves
  pair-judge's `exec.CommandContext` call); over a docs-only commit range it prints three
  `skip` lines.
- `scripts/flow-guard slot-judge <digest>` over this change's own Decide digest prints three
  lines ending in `judge` and the spend suffix; with `FLOW_PAIR_JUDGE_CMD` pointing at a command
  that fails, three `run` lines ending in `fallback: <cause>`.
- `flow decision render` over a decision with and without `slot_triggers` renders both.
- The live check (item 9) at `flow.verify`, as `## Live check` states.

## Live check

Item 9's recall backtest, run by `flow.verify` before the change lands. Nothing it runs writes a
store, a shared service or the change's own state.

- **The findings.** `scripts/flow-guard live-store-copy up kan-1126-live`, then, read-only on that
  copy, every finding whose `slot` is `test-sanity`, `call-efficiency` or `failure-modes`, whose
  `severity` is `Critical` or `Important`, and whose dispatch started on or after 2026-10-01 —
  53 findings across 33 changes (agents 13, gymie 19, flow-eval 1): call-efficiency 3 Critical
  and 12 Important, failure-modes 23 Important, test-sanity 15 Important.
  <!-- measured: docker exec flow-postgres psql -U flow -d flow -Atc "select f.slot, f.severity, count(*), count(distinct f.change_id) from findings f join dispatches d on d.id=f.dispatch_id where f.slot in ('test-sanity','call-efficiency','failure-modes') and f.severity in ('Critical','Important') and d.started_at >= '2026-10-01' group by 1,2" read-only on the dev store on 2026-11-15; a live store, not re-runnable to the same figure -->
  Three more such findings carry no dispatch and so no date; they are left out and counted.
  `scripts/flow-guard live-store-copy down kan-1126-live` when done.
- **Each finding's range.** Its change's commits recorded on implementer dispatches
  (`dispatches.commit_sha`) that started before the finding's dispatch and resolve in that
  project's main checkout (`~/Projects/agents`, `~/Projects/gymie`, `~/Projects/flow-eval`): the
  tip is the latest of them, the base the parent of the change's earliest one. 26 of the 33
  changes have at least one such commit; the other 7 are reported as unresolvable and not scored.
  <!-- measured: git -C <checkout> cat-file -e <sha>^{commit} over the 341 recorded commits of the 33 changes, on 2026-11-15 at agents b3314640 and gymie 07f5d7ca -->
- **The guard.** For each finding, a detached worktree of its project at the tip,
  `git -C <checkout> worktree add --detach <checkout>-worktrees/kan-1126-live-<n> <tip>`, then
  `scripts/flow-guard panel-slot-triggers <that worktree> <base>` from this change's worktree, and
  `git -C <checkout> worktree remove <that worktree>`.
- **The judge.** `scripts/flow-guard slot-judge <digest>` over the change's archived `tasks.md`
  (`change_files`) where one exists — 10 of the 33 changes — the digest that file alone; where none
  exists the judge counts as `run`, which is what the panel does without a verdict.
  <!-- measured: count of the 33 changes with a change_files row ending tasks.md, read-only on the dev store on 2026-11-15 -->
- **Figures.** Before: every finding's slot ran, none skipped. After: how many findings' own
  slot the guard alone would skip, the judge alone, and both together; and, over the 26 scored
  changes, how many gated dispatches the combination skips.
- **What failure looks like.** A combined skip on a finding's own slot: a missing rule, added to
  the guard before landing and the backtest run again. A guard exit 2 on most ranges, or the same
  three lines for every range, is a check that never read the diffs, not a pass.

## Non-goals

- principles is not judged by the LLM and not gated by the diff; it stays keyed on the class.
- mutation is never gated.
- No project configuration of the trigger rules; no `## panel slot triggers` section.
- No new policy row, Allowed pairs entry or policy migration.
- No flowd, `records.go`, `aggregate.go` or stats change; the decision JSON is stored as-is.
- The plan-time verdict is not re-asked on fix rounds.
- Micro's panel is unchanged.

## Risks

- A call or query reached through an unnamed helper outside the listed directories, in a plan
  that does not name it, skips the slot. Limited by the judge adding slots back, and measured by
  the live check, whose misses become rules before landing.
- The content rules over-trigger on common tokens (`<-`, `await`, `chan`), which keeps a slot that
  could have skipped; this costs a dispatch, never a review.
- The judge adds one `claude -p` call per Decide, bounded by the 60 s deadline and the cheap
  pair; any failure falls back to `run`.
- Extracting the judge call from `pjAsk` can change pair-judge's behaviour; `pairjudge_test.go`
  pins it across the extraction.

## Decisions

### Built-in rules over changed lines decide call-efficiency and failure-modes

**ID:** trigger-rules-builtin-changed-lines
**Status:** active
**Chosen:** A. built-in content and path rules over changed lines only, no project configuration — the operator's round-1 pick.
**Considered:** B. rules widened to the enclosing function (`git diff -W`), a boundary-named-receiver rule `\w*(api|client|repo|dao|store|service|cache|db|tx|conn)\.\w+(`, and an additive `## panel slot triggers` section in `.flow/project.md` — its rules over-trigger in Go and Kotlin, so `agents`, `spectre` and `flow-eval` would run both slots on nearly every code change, shrinking the saving the ticket asks for, and it adds a `project.md` section with its contract row and parser. C. project-declared globs only, fail-open without a declaration — every project, today all of them, keeps the always-on cost until it writes globs, and path-only misses calls in scattered layouts, so the globs would have to be broad.

### The LLM judge runs at plan time

**ID:** slot-judge-at-plan-time
**Status:** active
**Chosen:** a plan-time LLM judge at Decide, in addition to the diff guard — the operator's round-2 message asked for it "on the stage of planning".
**Considered:** an LLM judge at panel time over the real diff — not what was asked, and the guard already reads the diff.

### A slot is skipped only when both sources say skip

**ID:** slot-judge-union
**Status:** active
**Chosen:** at panel time a gated slot is skipped only when the plan-time judge and every worktree's guard line both say skip.
**Considered:** the judge alone decides and the guard is advisory — the plan predates the diff, so a plan that later grows a query skips with no check on the landed code. The guard alone decides and the judge is recorded only — the judge costs a call and changes nothing. Skip when either says skip — each source's miss becomes a silent skip, and it breaks the only-removes-on-a-definite-answer rule.

### slot-judge reuses pair-judge's call

**ID:** slot-judge-reuses-pair-judge-call
**Status:** active
**Chosen:** one shared function extracted from `pjAsk`, on the `pair-judge` row's first pair with its deadline, digest cap and fallback.
**Considered:** a second `claude -p` caller — duplicates the call. A new `slot-judge` policy row — needs a model-policy contract row, an Allowed pairs entry and a policy migration for one call per change.

### A failed or missing judge verdict runs every gated slot

**ID:** slot-judge-fail-open
**Status:** active
**Chosen:** a judge failure, a malformed or partial answer, or a missing `slot_triggers` counts as `run` for every gated slot.
**Considered:** deferring to the guard alone — an unanswered question never reduces a panel.

### The judge runs on every class

**ID:** slot-judge-every-class
**Status:** active
**Chosen:** Decide step 7 runs on every class, micro and fix runs included.
**Considered:** regular and big only — the extra dispatches it gates are recorded on every class.

### The judge reads Decide's digest

**ID:** slot-judge-digest
**Status:** active
**Chosen:** the plan's tasks and the `inputs:` line, the digest Decide's other calls use.
**Considered:** adding `design.md` — it repeats the tasks' `**Files:**` at greater length.

### The plan-time verdict is not re-asked

**ID:** slot-judge-not-re-asked
**Status:** active
**Chosen:** the verdict holds for the run's fix rounds; only the guard re-runs on the whole branch.
**Considered:** re-judging each round — the judge reads the plan, and the plan does not change.

### The judge covers three slots

**ID:** slot-judge-three-slots
**Status:** active
**Chosen:** test-sanity, call-efficiency and failure-modes only.
**Considered:** also judging principles — the ticket keys it on the `small` class.

### slot-judge prints the guard's format

**ID:** slot-judge-same-output-shape
**Status:** active
**Chosen:** `panel-slot-triggers`'s three-line format.
**Considered:** a JSON answer — the panel would read and record the two sources differently.

### principles drops on the effective small class

**ID:** principles-small-effective-class
**Status:** active
**Chosen:** `pcTree` drops principles on the effective class `small`.
**Considered:** the mechanical class — it would ignore a planner's recorded raise.

### Micro keeps its panel

**ID:** principles-micro-unchanged
**Status:** active
**Chosen:** micro keeps its `default` panel.
**Considered:** also dropping principles there — the ticket does not state it, and a micro surface is docs-only, which the docs-only reduction already narrows to `primary`.

### The diff guard decides at panel time

**ID:** triggers-at-panel-time
**Status:** active
**Chosen:** the guard decides test-sanity, call-efficiency and failure-modes at pass 1 from the diff.
**Considered:** plan time from `**Files:**` in `plan-class` — it predates the diff and is never re-planned on a fix run.

### Guard exit 2 runs every gated slot

**ID:** triggers-fail-open
**Status:** active
**Chosen:** exit 2 runs every gated slot.
**Considered:** skipping them — an unanswered question never reduces a panel (the docs-only rule).

### One exit code for three verdicts is not used

**ID:** triggers-three-lines-not-exit-code
**Status:** active
**Chosen:** three printed lines, exit 0 answered, exit 2 cannot answer.
**Considered:** an exit code per slot — three verdicts do not fit one bit.

### Bundles shrink to the roles they still name

**ID:** bundle-shrink-rule
**Status:** active
**Chosen:** a bundle dispatches the roles it still names, and one naming none is not dispatched — the fix-run judge's existing rule.
**Considered:** always dispatching both roles of a half-triggered bundle — keeps the cost the ticket removes.

### A slot firing on a fix round joins it

**ID:** fix-round-join
**Status:** active
**Chosen:** a gated slot that fires on a later fix round's whole-branch diff joins that round.
**Considered:** pass-1-only gating — would never review a test or I/O a fix adds.

### Fix the latent bundle matchers

**ID:** fix-latent-bundle-matchers
**Status:** active
**Chosen:** fix `rspRole` and the `panelroundbegins` later-begin match for partial bundles.
**Considered:** leaving them — a lone `mutation` would re-snapshot the plan, and a lone `test-sanity` would be judged on the wrong pair row.

### Decide still records both extra pairs

**ID:** extra-pairs-still-recorded
**Status:** active
**Chosen:** Decide records `extra_dispatch` and `extra_dispatch_mutation` on every class.
**Considered:** recording only the triggered slots — Decide cannot know them before the diff exists.

### Docs-only drops failure-modes too

**ID:** docs-drop-failure-modes
**Status:** active
**Chosen:** a docs-only diff drops failure-modes, leaving `mutation` alone.
**Considered:** scanning `.md` for failure-modes tokens — the ticket gates it on I/O, persistence or concurrency, and docs are none of these.

### File and exec calls trigger failure-modes only

**ID:** file-exec-is-failure-modes
**Status:** active
**Chosen:** file reads and writes and `exec.Command` trigger failure-modes, not call-efficiency.
**Considered:** triggering both — the ticket's call-efficiency condition names handlers, queries and client calls only.

### Test files are not content-scanned

**ID:** test-paths-excluded-from-content-scan
**Status:** active
**Chosen:** test paths are excluded from the call-efficiency and failure-modes content scan.
**Considered:** scanning them — test helpers doing I/O would trigger both.

### Directory rules match paths

**ID:** dir-rule-is-path-match
**Status:** active
**Chosen:** the handler/store/client and scheduler/jobs/workers/migrations directory rules match a changed file's path.
**Considered:** content matching — the option's "directories" wording rules it out.

### The recall backtest is the live check

**ID:** backtest-as-live-check
**Status:** active
**Chosen:** the backtest against findings since 2026-10-01 is the `## Verification` live check, run at `flow.verify`.
**Considered:** a plan task — a live check is where the plan contract puts work against the real store.

## Open questions
