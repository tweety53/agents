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

   **Fix instructions that dispute a visual judgement this session already made — a spacing, size
   or alignment an earlier round eyeballed as fine — open with the measurement, never with another
   look.** Run `flow-guard measure-visual-properties` on the disputed region of the current capture
   and the mockup (**10** in `skills/flow/visual-verify-verifier.md`) and put the numbers in the
   implementer's prompt as a `Measured:` line; a spacing dispute is measured on every side the
   complaint names. The complaint's own wording names the property — "too big", "oversized" is a size
   (`box` and `ink`); "cramped", "uneven", "too close" is a spacing (`gap`); "not filled to the
   border", "flush", "reaches" is an edge alignment (`runs` through the container) — so the
   measurement answers the property named, never the screen area the complaint sits in.

   **Every fix task for a visual mismatch starts with the measurement, disputed or not.** The
   first step of the task is the `Measured:` line above for the reported element on both sides; when
   the numbers show no difference, the task's result is `not a departure: <the measured numbers, both
   sides>` and no code changes, and in every case the measured numbers are in the task's result, so
   a later round that contradicts it must quote them.

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
   only the frames step 9's `measure` names on its `changed:` line, beside the `moved:` frames and
   every frame this fix's defects name, which its prompt names under `defects:`
   (`skills/flow/visual-verify-verifier.md`). Its `prepare` is passed each such worktree's
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
   whose final report carries no fixable defect closes it, `CLOSES-LOOP:` line or not: no
   `-fix-<k>-full` run is dispatched unless the operator asks for one in this session, and `flow
   record dispatch begin` counts a completed scoped row (one begun with `-round-base`) as the
   closing run. A scoped re-run that judged every changed frame closes it itself: when its `close`
   prints `CLOSES-LOOP: yes` — every frame on the change's list was re-captured after its row began
   or stands on a `captured:` spec, every frame whose capture changed since its round base, and
   every `--defects` frame, has a line of its own in the report, and no mockup or sidecar changed
   since that round base (`flow-guard visual-verify-pass`'s header) — `close` has ended its row
   `-closes-loop`. After a fix's step 2 commit, and until a `-fix-<k>-full` verifier row, or a
   `-fix-<k>` one carrying `closesLoop` or `-round-base`, has completed, `flow record dispatch
   begin` refuses a `visual-verify` key that carries neither `-fix-<k>` nor a later fix's `-fix-<n>`,
   and one that is not the `-full` run, carries no `-round-base` and shares its repository suffix
   with a completed verifier row. A completed `-fix-<j>-full` verifier row, or a completed
   `-fix-<j>` one carrying `closesLoop` or `-round-base`, closes every fix up to `<j>`. The clean
   report of the run that closed the check is the one the handoff reports.

   **A failing visual `verify` row is re-run narrowed** — **Narrowed visual re-runs**, below.

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
applies to the defect it fixes.

**A defect whose root cause is a shared source — a style, a token, a component other screens draw
too, the data every spec renders such as a UI-test seed or fixture — is fixed at that source, whatever else the fix moves.** A shared source's shipped look departing
from one feature's frames is not such a defect (**A departure from the mockup**,
`skills/flow-contracts/operator-prompts-auto-resolution.md`). The fix never patches the
changed screen alone to spare the others, and its steps carry two more. Every baseline the fix
moves is recaptured: every spec `flow-guard affected-specs <worktree> <round base> <capture root>`
prints for the fix's own diff where the capture root holds a `<spec>.sources` sidecar, else every
spec `specs` prints for it — none when it prints
nothing — plus every capture spec the fix's task names in its `**Tests:**` line — when it fixes a shared
component, every spec whose frames show that component, named by the planner or the finding that
raised the fix, never left for the diff to reveal — runs through one `capture` run with `--update-snapshots=all`
appended (step 8, `skills/flow/visual-verify-verifier.md`), and every PNG whose bytes changed is
read first, per **The sweep** (`skills/flow-contracts/known-bugs.md`), and
committed with the fix, in the checkout step 12 of **Steps 3–13** (`skills/flow/visual-verify.md`)
commits the stage's PNGs to. The fix's run takes the tests a
sidecar maps a changed file to and the tests its fix adds, a bare
spec path its every test, never
every spec a shared file reaches, which step 3's `visual-verify` re-run captures once. And every step-3 `visual-verify` re-run of the fix names under `moved:` in its prompt every frame a `<spec>.mockups` sidecar maps to a
spec whose baseline moved — the verifier compares each against its mockup exactly as it does the
change's own frames, so a screen the shared fix regressed is a departure the next fix fixes.

**A suite that also fails for a reason the fix did not cause is re-baselined per capture, never per
flag run** — `--update-snapshots=all` would write the unrelated failures' captures over their
baselines too. The suite runs without the flag; for each failing capture, every changed-pixel
blob of its diff image is classified against the fix's expected footprint — where the shared
source draws and how large its change is — and only a capture whose every blob fits has its
`-actual.png` copied over its baseline, each copied baseline read before it is staged. A test
stops at its first failing capture, so the run repeats until no failure inside the footprint
remains, and a capture with a blob outside it is a defect per **The sweep**
(`skills/flow-contracts/known-bugs.md`), never copied.

**A fix round from the second on carries the full departure list.** Its prompt carries every
departure of the report that opened the loop plus every earlier round's report, never the latest
report alone, and the fixer compares every frame row and departure of that list before it reports
done.

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

**A fixer checks every consumer of what it changed before it reports done.** It lists every view
and frame that draws a component, copy or style its diff changed — the listed departures' siblings
included — reads each one's recaptured PNG against its frame, and fixes in the same round any it
no longer matches. A departure whose frame contradicts older spec text is fixed to
the frame and the spec text updated in the same round, never handed back as an open question
.

**Mockup departures are classified before the fix is written.** Where every frame that
draws the departing component agrees, the departure is a defect this round fixes.

**A departure that persists two rounds is re-judged before a third fixer opens.** Before the
round's fix is written, each departure that stood in both previous rounds is checked against the
frame's sidecar comment and the recorded decisions (`design.md`'s `## Decisions`), with its
measurement beside the declaration: one a declaration covers is closed as `expected` with the
declaration cited, one the evidence does not bear out (a mis-measurement) is dropped, and neither
opens a fixer round.

**A size the frames, or the agents reading them, disagree on is picked once, from a value a frame
draws, and never reopened.** Where frames disagree on one element's size, position, spacing or
radius — or two verifiers, fixers or rounds read it off the frames at different values — the
parent picks one value a frame draws: the element's own design source when one exists (the frame
that draws it as its subject, or the design-system frame it is taken from), else the value most
frames draw, else the first of those frames in the handoff's order. The pick is written once, before
the fix, as a `design.md` decision whose `**ID:**` is `drift-<element slug>` and whose
`**Chosen:**` names the element, the property, the kept value and its frame, and every other frame
with the value it draws; and as that drift in the store, a withdrawn finding of the parent's:

```bash
flow record finding -change <name> -ref <the ref `flow record next-ref -change <name>` prints> -slot parent -severity minor -location '<element> <property>' -status 'withdrawn drift: kept <value> from <frame>' -note '<every frame and the value it draws>'
```

From then on every verifier attributes the rows that pick explains `decision drift-<element slug>`
(step 10's closed list, `skills/flow/visual-verify-verifier.md`), so they count as no departure; no
round fixes the element toward another frame's value, picks it again or asks about it, and the
handoff's `**Decisions:**` lines name each drift. Whatever looks wrong in use the operator reports
after landing.

**A shared element that departs on some screens and not on others is fixed where those screens
differ, never by a per-screen offset.** When the frames agree on where an element sits and the
captures put it in different places on different screens, the cause is what those screens draw
differently around it — a title block, an inset, a hero — and the fix fixes that. A constant
added for one screen or one tab to land the element where that frame puts it moves the element off
the frames the other screens matched, and the next round reports them.

One round runs steps 1, 2 and 4 in order; step 3 runs once, at the loop's close:

**Every mockup departure the fix lands a permanent probe in the fix's one commit** —
a test that drives the real composable or component, never a hand-copied constant, and asserts
the value the verifier measured against the frame — so a later change that moves it again fails
a test instead of waiting for the next visual verification. The fix's tests name one probe per
departure it lists, and the implementer hands back a listed departure with no probe it watched
fail before the fix and pass after it as unfixed, never as fixed.

A departure seen in motion or over time — a frame range, a gesture, a delay — is probed
under the input the verifier recorded, its pace, distance and step timing, and sampled inside
the window the verifier saw it in, never with instantaneous steps or a settle wait that outlasts
that window; where the harness cannot reproduce those conditions, the commit body names the
difference.

**Every new branch the fix adds lands a test of its own in the same commit.** A condition, flag,
guard or state the fix introduces beyond the departure it closes — a hidden placeholder, a reset
on return, a skipped refresh — carries a test that fails with that branch removed.

**The implementer measures before it hands back.** When the defects include a mockup departure,
its dispatch names the absolute path of `skills/flow/visual-verify-verifier.md`, the capture spec
with its `<spec>.mockups` sidecar, the resolved `mockups` root and `mockup frame` value, and the
project's `capture` command, and tells it: on every view its fix touched, recapture against the
stack rebuilt from its fix, as step 3's confirming re-run rebuilds it, and run step 9's
`flow-guard visual-verify-pass measure` on each under the key `fix-<k>-implementer`, reading its
manifest's flagged bands, seams and `over` rows — fixing until they come back clean or each
leftover is named with its cause, and report every view's measured numbers. It measures each pair
on the 2× capture and the 2× mockup frame themselves, never on a `flow-guard
compose-mockup-frames` composite. A leftover of any size — a 1dp offset included — is closed or
named with its cause, never reported clean because it is small.

**A fix for a visual departure runs its capture spec before it closes.** Its verify step runs the
project's `capture` command with the recapture list above for `<spec>`, against the stack rebuilt
from its fix, and the fix closes only on that run's exit 0 — a red capture is fixed in the same
commit. The capture command takes the list `flow-guard affected-specs <worktree> <round base> <capture root>` prints for the fix's diff; `flow-guard recapture` needs a `tasks.md` task a fix has none of, so it is not used.

**A departure observed on a running clock closes on repeated running-clock passes.** When the
departure was observed on a running clock — a live walk, a recorded frame sequence, any read the
capture spec's pinned clock (`skills/flow/visual-verify-verifier.md`, step 8) does not
reproduce — the fix first adds a capture test that runs under the observer's conditions (clock
running, same viewport and data) and fails without the fix, and closes only once that test
passes five times over in one run with one worker (Playwright
`--repeat-each 5 --workers 1`), beside the capture spec's own exit 0.

**Every recapture the fix runs rebuilds the full app suite's zip after it** —
`flow-guard rebuild-full-app-suite-zip <worktree> <screenshots root>` in the checkout the recapture
wrote to, the rebuilt `full-app-suite.zip` committed with the fix's PNGs — so the zip never
trails the baselines it archives.

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

## Narrowed visual re-runs

**A failing visual `verify` row is re-run narrowed, never whole.** When the `## visual verification`
`verify` row a `test` group ends with fails, its one inline re-run is `flow-guard run-verify-list
-only <visual-failing.txt> <worktree> visual [<canonical worktree>]`, the `visual-failing.txt` file
the failing run prints, run once with one retry, and one fixer gets every test that run fails and
every test it prints as `flaky` in one batch. A flaky test is a failure the fixer fixes at its
source (`rules/fix-determinism-at-the-source.mdc`), and never a reason to run the selection again.
Each fix round then re-runs that file the same way. Once a narrowed run is green and prints no
`flaky` line, one plain run of the `test`
group under the round's `-fix-<k>` key is the gate, and when it fails the next round narrows to
its failures. That re-run selects from the round's diff, never the change's, plus the spec file of
each `visual-failing.txt` entry, and `-only` selects as the row it follows did — `flow-guard
run-verify-list` reads the round base from the previous verify row itself (THE ROUND'S SELECTION,
its header canonical). `-only` records no dispatch row, writes no green cache and never stands as the gate.
A plain run is the `specs` row's selection — the
change's tests plus the smoke tail —
and never the whole suite. A selection naming more than 40 spec files runs the specs the change edits plus that tail, and is never refused (A WIDE SELECTION, `flow-guard run-verify-list`'s header canonical). Each failing run prints its failures in two groups, the
change's specs and the specs outside the `specs` row's output. Two consecutive rounds whose failures all fall outside it make no progress, and escalate as **Fewest operator actions** states (`skills/flow-contracts/pipeline.md`). A non-visual command keeps its own inline re-run.

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
