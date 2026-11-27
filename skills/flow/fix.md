# The fix procedure

Loaded by `/flow` for every fix: the in-run fix a verify stage's defect opens (**Verify**,
`skills/flow/verify-and-handoff.md`), and a fix run — state `IN_PROGRESS` with fix instructions, or a
plain message at `IN_PROGRESS` (`skills/flow/SKILL.md`). One procedure serves both. Every "you"
addresses the parent (**The parent orchestrates directly**, `skills/flow/implement.md`).

**A fixable defect the verify stages find is fixed by this run, never handed to the operator as a
question.** The fixable defects are a lint or test failure **Inline verify**
(`skills/flow/verify-and-handoff.md`) routes here, and every item the **Blocking** paragraph of
**Steps 3–13** (`skills/flow/visual-verify.md`) routes here. The choice "fix it in-run, or hand
off with it open?" is never asked: the run takes **fix it in-run** under **Auto-resolution**
(`skills/flow-contracts/operator-prompts.md`), recorded and named on the handoff's
`**Decisions:**` lines as that section states, one entry per defect, the defect named as the
report names it.

**Input:** the fix instructions, verbatim — the verifying stage's final `## Report`, every defect
it names, in-run; the operator's message on a fix run. `<k>` is this fix's number: one more than the
number of `fix-<k>.md` files in `<abs-worktree>/.superpowers/sdd/`, counted across both stages, so no
two fixes share a `-fix-<k>` key. A fix run keeps the planning run's decision: no plan-check decision runs and
no `flow.plan` is marked.

One fix runs steps 1 to 4 in order; a Critical or Important finding of step 4 opens fix `<k+1>`.

1. **Open the fix.** Write the fix instructions, verbatim, to
   `<abs-worktree>/.superpowers/sdd/fix-<k>.md`, beneath each worktree's **fix-start sha**,
   `git -C <worktree> rev-parse HEAD`, read before anything is edited. No planner is dispatched,
   nothing is appended to `proposal.md` or `tasks.md`, `spectre validate` does not run and no Jira
   description sync is made. Before the fix, the stage's failing dispatch row is closed with `flow
   record dispatch end` at `-outcome blocked -cause test-failure`, so no row of the stage stays open
   while this fix's rows open; the failing stage's mark — `flow.verify` — is closed beside it,
   `flow stage end -command '/flow' -stage flow.verify -outcome stopped <name>`, and a run that
   re-enters this procedure while `flow.verify` is open closes that mark the same way first, so
   step 3's `flow stage begin` supersedes no open run. On `NEXT: loop`,
   `flow-guard visual-verify-pass close` has closed the row, once every surviving worktree's
   newest row is closed (**The verifier dispatch**, `skills/flow/visual-verify.md`).

   **A visual fix** — one whose defect or fix instructions name a mockup departure, a capture or a
   `ui paths` file — loads `skills/flow/fix-visual.md` here; its **Measuring first** section binds
   this step.

2. **Fix it test-first**, under `flow.implement` (`flow stage begin -command '/flow' -stage
   flow.implement <name>` when no `flow.implement` mark is open). One fresh implementer, role
   `implementer`, key `fix-<k>-implementer`, its pair from `flow-guard pair-judge implementer`,
   launched like any implementer (**4. Execute (SDD + TDD)**, `skills/flow/implement.md`): its
   prompt is the implementer stub with a `FIX INPUT:` line naming `fix-<k>.md` and the `Measured:`
   line when step 1 wrote one. It writes the failing test, sees it fail, fixes, sees it pass, and
   runs its own lint and its tests by name. Its commit is
   `git -C <abs-worktree> commit -m "fix(<module>): <subject>" --trailer "Flow-Key: fix-<k>-implementer" --trailer "Invariant: <one line>" -- <the changed paths>`,
   no `Task-Id:` trailer and no `flow-guard close-task` (**Commits outside every task**,
   `skills/flow-contracts/git-boundaries.md`); the parent pushes it per its branch-last-read rule
   (`skills/flow/review-panel.md`). The fix's own `flow.verify` record
   (`<changeRoot>/visual-verification.md`) sections are rewritten beside the fix's paths as DRIFT
   CHECK (`skills/flow/implement.md`) requires; the record and `<changeRoot>/visual-verification/`
   are written on disk only, the change directory gitignored. The fix is held to the paragraphs
   below and to `skills/flow/fixer-brief.md`, its prompt the fixer stub (**Every fix is a fresh
   one-shot**, `skills/flow/implement.md`); the fix closes when the
   re-run of step 3 passes. **A verifier's report naming many departures is split from the first
   dispatch**: one fixer per family (a frame group or a screen), dispatched in series, each its own
   `fix-<k>`, never one fixer handed the whole list. **A large fix task — one fix whose family
   touches more than one implementation group of files — runs one group at a time by default,
   each group amended into the fix's one commit** (`git commit --amend --no-edit -- <the group's
   files>`, its trailers kept), so a fixer that runs out of working memory part-way leaves a commit
   holding every finished group. The fix's next dispatch resumes from that commit: it reads
   `git -C <abs-worktree> show --stat HEAD` for the groups done, takes the first group the fix
   instructions name that the commit does not yet touch, and redoes nothing.
   **A deferred item is ruled on before step 3 dispatches a re-verify round.** Every `not-fixed:`
   line of the fixer's report, and every item it dismisses as covered by a decision, goes back to a
   fixer — the fix's next dispatch — or to the judge for a ruling: fixed, or withdrawn with its
   reason recorded beside the frame in `visual-verification.md`. Step 3 runs only once no item is
   left unruled, so no deferral is first re-raised by the next verifier. A withdrawn item's frames
   are still read by the closing full pass.

3. **Re-run the failing check**, under `flow.verify`. In-run: the stage that reported the defect,
   re-run as `skills/flow/verify-and-handoff.md`'s Verify and Visual verification
   (`skills/flow/visual-verify.md`) say from their `begin` marks, under the dispatch keys
   `verify-fix-<k>[-<repository name>]` and `visual-verify-fix-<k>[-<repository name>]`, never
   `fix-<k>-verify…`, which `flow-guard run-verify-list` refuses (exit 2); when visual
   verify failed and code changed since the last green lint/test run, the lint/test list re-runs
   first, so the landing gate is unchanged. On a fix run: the lint/test list, then visual verify
   only when the fix touched `ui paths` — a fix that touches none needs no `visual-verify` run.
   **A fix round re-runs only the test commands that failed, until they pass;** the one default-list
   run that confirms the whole list follows, before landing (**Inline verify — the confirm run**,
   `skills/flow/verify-and-handoff.md`).
   **The `verify` re-run runs only in a worktree whose HEAD moved since its previous verify of
   this run, or whose test commands drive the stack of one that did.** **The `visual-verify`
   re-run is scoped to the fix's own diff** in every worktree with a round base — the HEAD step
   12 of **Steps 3–13** (`skills/flow/visual-verify.md`) read after that worktree's previous
   verifier dispatch: its `motions:` names only the motions `git diff <round base>..HEAD` adds,
   changes or removes, plus every motion this fix's defects name. Its verifier still captures every
   frame on the change's list — save the specs its bundle names on a `captured:` line, which the
   fix's recapture passed on HEAD's sources (`flow-guard visual-verify-pass`'s header) — and judges
   only the frames its prompt and step 9's `changed:` line name (`skills/flow/visual-verify-verifier.md`). Its `prepare` is passed each such worktree's
   `--round-base`, which the verifier row records, and `--defects <worktree>=<frame>[,<frame>…]`
   naming that worktree's `defects:` frames. A worktree whose round diff is empty is not dispatched; a
   worktree with no round base — no verifier of this run has verified it, or the parent no longer
   holds the sha — runs the stage unscoped. **A fix whose diff touches only the mockups root
   composes instead of verifying.** When `git diff <round base>..HEAD` matches no `ui paths` entry
   in any worktree and touches nothing outside the mockups root, `prepare` prints `NEXT:
   compose-only` and no `visual-verify` verifier is dispatched: for each capture spec whose
   `<spec>.mockups` sidecar has a line naming a changed `<frame id>.png`, the parent runs step 9's
   `flow-guard visual-verify-pass measure` (`skills/flow/visual-verify-verifier.md`) under the fix's
   key and reads each manifest's flagged bands, seams and `over` rows; a leftover is a defect that
   opens the next fix, and clean composites close the check with no `-full` run — either way the
   `flow.verify` ends as **Verify** states (`skills/flow/verify-and-handoff.md`), `stopped` when a
   leftover opens the next fix. **The confirming re-run
   verifies a fresh build, never a stale stack:** before it reports the fix confirmed live,
   rebuild the stack from the branch HEAD — current source through the project's own build
   commands, brought up the way step 5 of **Steps 3–13** (`skills/flow/visual-verify.md`) starts
   one — and re-run the defect's exact scenario end to end against that build. The rebuild is itself a defect-finder: a defect the fresh build exposes that the fix
   does not cover is reported the way any stage-found defect is, never absorbed into the fix's
   confirmation. A re-run that reports a fixable defect opens fix `<k+1>` at step 1.

   **A scoped re-run closes the check; a full run is the operator's to ask for.** A scoped re-run
   whose final report carries no fixable defect closes it: no
   `-fix-<k>-full` run is dispatched unless the operator asks for one in this session. The clean
   report of the run that closed the check is the one the handoff reports.

   **A failing visual `verify` row is re-run narrowed** — **Narrowed visual re-runs**
   (`skills/flow/fix-visual.md`).

4. **One reviewer**, under `flow.review` (`flow stage begin -command '/flow' -stage flow.review
   <name>`). Once step 3 is clean, `flow-guard write-panel-diff late-fix <abs-worktree> <worktree>
   <fix-start sha> [<worktree> <fix-start sha>…]` writes `<abs-worktree>/.superpowers/sdd/late-fix.diff`
   from the fix-start shas, and one dispatch reads it: the `primary` slot, role `panel-reviewer`, key
   `fix-<k>-review`, its pair from `flow-guard pair-judge panel-reviewer`, its brief rendered from
   `skills/flow/review-panel.md`'s placeholder table with `[DIFF_PATH]` the `late-fix.diff` and the
   FIX-ROUND SCOPE paragraph of **Panel re-runs** (`skills/flow/review-panel-fix-round.md`). A
   Critical or Important finding opens fix `<k+1>` through this procedure, the Minors riding it;
   a read with neither ends the stage with `flow stage end -command '/flow' -stage flow.review
   -outcome completed <name>`. The fix's `verify` re-run always runs after a fix the review
   opened — `flow-guard run-verify-list` records `skipped` itself where nothing a test command reads
   changed — since a code commit no completed or skipped verify row covers is unfinished work
   (`flow-guard check-unfinished-work`).

## The fix implementer's brief

The fix's dispatch prompt names this section, and the implementer follows every paragraph that
applies to the defect it fixes; a visual fix's prompt also names **The visual fix implementer's
brief** (`skills/flow/fix-visual.md`).

**The fix ledger: no fix silently undoes an earlier one**. Every fix commit of step 2
carries an `Invariant: <the one line its fix protects>` trailer beside its `Flow-Key:`. The ledger
is derived from git, never stored: one row per fix commit since `<base>` — the branch's flow base,
`branch.<current>.flowBase` when recorded, else `refs/remotes/origin/HEAD`, so a fix committed before
this loop began is a row too — read by
`git -C <abs-worktree> log --stat --format='%h %(trailers:key=Flow-Key,valueonly,separator=%x2C) %(trailers:key=Invariant,valueonly)' --grep='^Flow-Key: fix-' <base>..HEAD` —
the finding id (`fix-<k>`, its finding in `fix-<k>.md`), the sha, the invariant and the files
touched. Every fixer and reviewer prompt of the loop carries that output in full under
`LEDGER:`. A fix touching a line a ledger row's commit wrote keeps that row's invariant, or
returns `CONFLICT: <row sha> <invariant>` with no commit, never reverting it; the parent then
dispatches the next fixer with both findings to reconcile. After each step 2 commit the parent runs
`flow-guard check-fix-revert <abs-worktree>` (it resolves the same `<base>`); its exit 1 — the fix inverts more than three
lines of an earlier fix, both shas printed — is that same `CONFLICT`, and the reconciling fixer
amends that fix commit rather than adding one on top.

One round runs steps 1, 2 and 4 in order; step 3 runs once, at the loop's close:

**Every new branch the fix adds lands a test of its own in the same commit.** A condition, flag,
guard or state the fix introduces beyond the departure it closes — a hidden placeholder, a reset
on return, a skipped refresh — carries a test that fails with that branch removed.

**A fixer batches its edits, then builds once.** It makes every edit the round's findings call
for before it builds anything, builds once, and then repairs only what that build or its capture
still shows — never one build per edit.

**A passing test that asserts the behaviour the fix instructions report as wrong is evidence of
the code, not of the spec — it decides nothing on its own.** Before treating such a test as the
tie-breaker, search this change's `design.md`, `proposal.md`, panel records and the linked Jira
issue for a sentence that decided *this* point. One found: cite it, change nothing, and end with
`won't fix: <cite>`, which the parent hands back through `## Question` rather than silently
changing what the spec required. None found: the test guarded an unexamined implementation
choice, the report wins, and the fix changes the test alongside the behaviour, its commit saying so
("no design decision covers this; the prior test locked in the behaviour the report flags"). A
test whose own name reads as a description of the reported bug is a signal to pause on, not
reassurance. A `Measured:` line answers a visual dispute the same way: its numbers decide whether
the fix changes anything or answers "it matches", the latter a `won't fix:` naming them.

**An observed failure is never closed as "not reproduced".** This rule is canonical for every
fixer of a test failure a run saw — this procedure's, a full-suite fix's, a self-review loop's. The
fixer traces it from the code and the failing run's saved artifacts (a load reproduction is allowed only with a bounded worker count —
PROOF RUNS, `skills/flow/implementer-brief.md`), and keeps every
failing run's result files (the `result files:` and `full output:` paths `flow-guard run-verify-list`
prints, or the JUnit XMLs and logs copied out by hand) before the next run overwrites them. A
failure that reproduces gets its determinism fix, held to the PROOF RUNS paragraph
(`skills/flow/implementer-brief.md`). A forced cause also shows the test failing with the mechanism
removed; only one that does not reproduce within its bounded runs waives that run, and it goes straight
to code. One that does not reproduce within its bounded runs is fixed anyway: every source of nondeterminism the trace names as a candidate — a container
readiness wait, a shared port, an unpinned clock or order — is fixed at its source, each with a
forced reproduction as its evidence (the candidate made to fail on purpose, a delay or a fault
injected, failing the test without the fix and passing with it) where the candidate can be forced. The fix's one commit carries
those fixes and its body states each forced reproduction, or that the fix
went straight to code; a fixer that names no candidate has not finished the trace, and is sent back to it.

## No cap

**No round count ends the loop.** **Fewest operator actions** (`skills/flow-contracts/pipeline.md`)
is canonical for that, for a round that makes no progress, and for **Three failed rounds** of this
procedure — a round being one fix.

## The late review

**A loop that ran more than one fix round ends on one late review before the stage hands off.**
Once the last round is green, the parent dispatches one reviewer over the union of the loop's fix
diffs — `<first fix round's base>..HEAD` — reading for what no single round's reviewer saw because
each saw its own slice: a value hard-coded or drawn from a frame that no model holds, a branch
without a test, a baseline moved without a reason. The dispatch takes its pair from the role the
reviewer's work falls under in `flow policy get`, never a pair named here. What it finds takes
this procedure as any defect does, and the late review runs again only if that fix round added
more than one commit.
