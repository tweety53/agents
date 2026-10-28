# The in-run fix loop

Loaded by the load directive under **Verify** (`skills/flow/verify-and-handoff.md`) only when the
final report of `flow.verify` or `flow.visual-verify` carries a fixable defect. Every "you" here
addresses the parent (**The parent orchestrates directly**, `skills/flow/implement.md`).

## The loop

**A fixable defect the verify stages find is fixed by this run, never handed to the operator as a
question.** The fixable defects are a lint or test failure **Inline verify**
(`skills/flow/verify-and-handoff.md`) routes here, and every item the **Blocking** paragraph of
**Steps 3–13** (`skills/flow/visual-verify.md`) routes here. The choice "fix it in-run, or hand
off with it open?" is never asked: the run takes **fix it in-run** under **Auto-resolution**
(`skills/flow-contracts/operator-prompts.md`), recorded and named on the handoff's
`**Decisions:**` lines as that section states, one entry per defect, the defect named as the
report names it.

**A defect whose root cause is a shared source — a style, a token, a component other screens draw
too — is fixed at that source, whatever else the fix moves.** The round's task never patches the
changed screen alone to spare the others, and its steps carry two more. Every baseline the fix
moves is recaptured: every spec `specs` prints for the fix's own diff — the whole suite when
`specs` is not declared or prints nothing — runs through `capture` with `--update-snapshots=all`
appended (step 8, `skills/flow/visual-verify-verifier.md`), and every PNG whose bytes changed is
read first, per **The sweep** (`skills/flow-contracts/known-bugs.md`), and
committed with the task, in the checkout step 12 of **Steps 3–13** (`skills/flow/visual-verify.md`)
commits the stage's PNGs to. And every step-4 `visual-verify` re-run of the round, the closing full
run included, names under `moved:` in its prompt every frame a `<spec>.mockups` sidecar maps to a
spec whose baseline moved — the verifier compares each against its mockup exactly as it does the
change's own frames, so a screen the shared fix regressed is a departure the next round fixes.

**A suite that also fails for a reason the fix did not cause is re-baselined per capture, never per
flag run** — `--update-snapshots=all` would write the unrelated failures' captures over their
baselines too. The suite runs without the flag; for each failing capture, every changed-pixel
blob of its diff image is classified against the fix's expected footprint — where the shared
source draws and how large its change is — and only a capture whose every blob fits has its
`-actual.png` copied over its baseline, each copied baseline read before it is staged. A test
stops at its first failing capture, so the run repeats until no failure inside the footprint
remains, and a capture with a blob outside it is a defect per **The sweep**
(`skills/flow-contracts/known-bugs.md`), never copied (KAN-837).

**Mockup departures are classified before the round's task is written.** Where every frame that
draws the departing component agrees, the departure is a defect this round fixes. Where frames
disagree on a shared component, the component's own design source wins — the frame that draws it
as its subject, or the design-system frame it is taken from: the departures that source backs are
fixed, and the other frame's departure closes as a mockup conflict, named on the handoff's
`**Decisions:**` lines beside the frame it lost to — never fixed, and never asked (KAN-801).

One round runs these steps in order:

1. **Transcribe the fix.** Run **Documenting a fix, before implementing it**
   (`skills/flow/document-fix.md`) with the stage's final `## Report` — every defect it names,
   verbatim — as the fix instructions. Neither of its two asks is made: the fix is appended to
   `proposal.md` and `tasks.md`, and `**Tasks appended:**` is not raised, since it counts appends
   made at the human gate. The appended task carries the field family the verification-change
   paragraph of **D. Basic Workflow #3 — Writing plans** (`skills/flow/brainstorm-planner.md`)
   states for a found defect, and its title reads `In-run fix <k> — <stage>: <the defects>`, `<k>`
   this stage's round in this run. On a `flow.visual-verify` round its `**Files:**` lists the
   change's `<changeRoot>/visual-verification.md` beside the fix's own paths, and the task rewrites
   the record's sections its fix changes in the same commit, as DRIFT CHECK
   (`skills/flow/implement.md`) requires (kan-1038). `<changeRoot>/visual-verification.md` is
   untracked until that task's commit, which `git add`s it, and `close-task` runs the plan guard on
   the task before ticking it (kan-1046 self-review, 1f). Its planning commit and the re-decide run as
   that file states.
   **One round appends one task**, carrying every defect the report names — never a task or a
   round per defect. Before the append, the stage's failing row is closed with `flow record dispatch
   end` at `-outcome blocked -cause test-failure`, so no row of the stage stays open while this
   round's rows open. A failing `flow.visual-verify` also has its stage mark closed beside that row, `flow stage end -command '/flow' -stage flow.visual-verify -outcome stopped <name>`, and a run that re-enters this loop while `flow.write-in-progress` is open closes that mark the same way first, so the round's next `flow stage begin` supersedes no open run (kan-1016: `flow.visual-verify` attempts 1–9 and 11 closed `superseded`).
2. **Implement it.** **4. Execute (SDD + TDD)** (`skills/flow/implement.md`) executes the appended
   task exactly as a plan-time task: the decision's implementer pair, `close-task.sh` with its
   guard and its review gate. The task closes through `close-task.sh` — an inline-committed one included — before step 3 opens any `panel-<round>-*` dispatch; `flow record dispatch begin` refuses a panel key while `tasks.md` holds an unticked task (kan-1007: panel round 4 began 16 seconds after task 8's inline commit, before its close). The implementer's dispatch also carries the MUTATION PROOF
   and PIXEL PROBE paragraphs (`skills/flow/review-panel.md`), and the round closes only when every reported
   `fix-mutation:` line is checked against the fix diff the way **The fix round mutation-proves
   what it changed** (`skills/flow/review-panel-fix-round.md`) walks them — a mutant nothing
   kills is a new fixable defect this loop takes another round on.
   **Every mockup departure the round fixes lands a permanent probe in the task's one commit** —
   a test that drives the real composable or component, never a hand-copied constant, and asserts
   the value the verifier measured against the frame — so a later change that moves it again fails
   a test instead of waiting for the next visual verification (KAN-790).
   **Every new branch the fix adds lands a test of its own in the same commit.** A condition, flag,
   guard or state the fix introduces beyond the departure it closes — a hidden placeholder, a reset
   on return, a skipped refresh — carries a test that fails with that branch removed, and the task's
   `**Regression:**` names each such test beside the departure's probe, never the probe alone
   (kan-1023: panel rounds 3, 5, 7 and 9 each raised an Important the round's own fix had introduced).
   **The implementer measures before it hands back.** When the round's defects include a mockup
   departure, its dispatch names the absolute path of `skills/flow/visual-verify-verifier.md`, the
   capture spec with its `<spec>.mockups` sidecar, the resolved `mockups` root and `mockup frame`
   value, and the project's `capture` command, and tells it: on every view its fix touched,
   recapture against the stack rebuilt from its fix, as step 4's confirming re-run rebuilds it, and
   run step 10's measurements on each — `compose-mockup-frames.sh` and the band, seam and matrix
   reads of `measure-visual-properties.sh` — fixing until they come back clean or each leftover is
   named with its cause, and report every view's measured numbers (KAN-1000: rounds 2–5 each found
   a new 1–3px offset the previous fix had judged by eye, 45–70 minutes a round). It measures each
   pair on the 2× capture and the 2× mockup frame themselves, never on a `compose-mockup-frames.sh`
   composite (kan-1015 self-review). A leftover of any size — a 1dp offset included — is closed or named with its cause, never reported clean because it is small (kan-1007: a fix round called a first-try 1dp offset clean).
   **A task fixing a visual departure runs its capture spec before it closes.** Its verify step
   runs the project's `capture` command with the round's capture spec for `<spec>`, against the
   stack rebuilt from its fix, and the task closes through `close-task.sh` only on that run's exit
   0 — a red capture is fixed in the same task and folded into its one commit (kan-1015: Tasks 21–24
   closed on desktop geometry tests, and each departure surfaced only at the next verifier).
   **A departure observed on a running clock closes on repeated running-clock passes.** When the
   departure was observed on a running clock — a live walk, a recorded frame sequence, any read the
   capture spec's pinned clock (`skills/flow/visual-verify-verifier.md`, step 8) does not
   reproduce — the task first adds a capture test that runs under the observer's conditions (clock
   running, same viewport and data) and fails without the fix, and closes only once that test
   passes five times over in one run under `run-capped.sh` with one worker (Playwright
   `--repeat-each 5 --workers 1`), beside the capture spec's own exit 0 (kan-1044 self-review A8-a:
   Tasks 7 and 8 closed green on the pinned-clock capture, and Task 9 on one running-clock pass
   against a four-in-five failure rate, three extra visual-verify rounds following).
   **Every recapture the round runs rebuilds the full app suite's zip after it** —
   `rebuild-full-app-suite-zip.sh <worktree> <screenshots root>` in the checkout the recapture
   wrote to, the rebuilt `full-app-suite.zip` committed with the task's PNGs — so the zip never
   trails the baselines it archives (kan-1003: two rounds left it stale).
   **An observed failure is never closed as "not reproduced".** This rule is canonical for every
   fixer of a test failure a run saw — this loop's, a full-suite fix's, a self-review loop's. The
   fixer traces it from the code and the failing run's saved artifacts (a load reproduction is allowed only under run-capped.sh with a bounded worker count —
   PROOF RUNS, `skills/flow/implement.md`), and keeps every
   failing run's result files (the `result files:` and `full output:` paths `run-verify-list.sh`
   prints, or the JUnit XMLs and logs copied out by hand) before the next run overwrites them. A
   failure that reproduces gets its determinism fix, held to the PROOF RUNS paragraph
   (`skills/flow/implement.md`). A forced cause also shows the test failing with the mechanism
   removed; only one that does not reproduce within the cap waives that run, and it goes straight
   to code. One that does not reproduce within the cap is fixed anyway: every source of nondeterminism the trace names as a candidate — a container
   readiness wait, a shared port, an unpinned clock or order — is fixed at its source, each with a
   forced reproduction as its evidence (the candidate made to fail on purpose, a delay or a fault
   injected, failing the test without the fix and passing with it) where the candidate can be forced. The task's one commit carries
   those fixes and its body states each forced reproduction, or that the fix
   went straight to code; a fixer that names no candidate has not finished the trace, and is sent back to it.
   The task ticks through `close-task.sh <canonical-worktree> <name> <merge-base> <n>:<that commit's sha>`
   as any task.
3. **Re-review the delta.** Open `flow.review-panel` again and run `primary` beside **The always-on
   extra dispatches** (`skills/flow/review-panel.md`), reading `late-fix.diff` over the range since
   the panel's last clean close, written, paired, ceilinged and recorded as **The late-fix
   reduction** (`skills/flow/review-panel-late-fix.md`) states its dispatch, its prompt carrying
   the FIX-ROUND SCOPE paragraph of **Panel re-runs** (`skills/flow/review-panel-fix-round.md`).
   The three dispatches are keyed `panel-<round>-primary`,
   `panel-<round>-test-sanity+call-efficiency` and `panel-<round>-mutation+failure-modes`, `<round>` the number `flow record next-round`
   prints, and the round records its `flow record pass -round <round>` rows like any panel round.
   That file's finding, voiding and staleness rules
   apply as written — a clean or Minor-only read leaves every slot it did not dispatch current.
   Once any round's read raises a Critical or Important, the reduction is void for the rest of the
   run: every later round's step 3 opens the decided full roster, never `primary` alone.
   `check-late-fix-trigger.sh` is not run: its scope-growth condition denies every appended task the late-fix reduction,
   and this task's premise is a defect a verifier measured, not an operator's unverified flag.
   The stage then closes through its own close (**Review panel**, `skills/flow/review-panel.md`).
4. **Re-capture and re-run.** Run **Verify** (`skills/flow/verify-and-handoff.md`) again from its
   `begin` mark, then **Visual verification** from its `begin` mark, so the fix's own lint, tests and
   captures are what the handoff reports. This round's `verify` and `visual-verify` dispatch keys
   carry the suffix `-fix-<k>`, before any `-<repository name>` suffix. **The `verify` re-run
   runs only in a worktree whose HEAD moved since its previous verify of this run, or whose test
   commands drive the stack of one that did** — an unmoved worktree's lint and tests already
   passed on that same HEAD. **The `visual-verify`
   re-run is scoped to the round's own diff** in every worktree with a round base — the HEAD step
   12 of **Steps 3–13** (`skills/flow/visual-verify.md`) read after that worktree's previous
   verifier dispatch: its `motions:` names only the motions `git diff <round
   base>..HEAD` adds, changes or removes, plus every motion this round's defects name. Its prompt
   names the `<frontend-root>` worktree's own round base, which the verifier's step 7 substitutes
   for `specs`' `<merge-base>` — never another worktree's of the change; when the
   `<frontend-root>` worktree has no round base, the prompt names none and `specs` runs unscoped.
   Before that dispatch the parent runs `specs` itself with the round base for `<merge-base>`, names
   the count of paths it prints in the prompt as `scoped specs: <n>`, and passes `<n>` to
   `check-verify-report.sh` as its third argument, which refuses a report whose `verify:` line
   names another count (kan-1015: a re-run under round base `c75529d2` printed 2 specs, and the
   verifier ran 95). A worktree
   whose round diff is empty is not dispatched. A worktree with no round base — no verifier of
   this run has verified it, or the parent no longer holds the sha — runs the stage unscoped. **A round whose diff touches only the mockups root composes instead of verifying.** When `git diff <round base>..HEAD` matches no `ui paths` entry in any worktree and touches nothing outside the mockups root, no `visual-verify` verifier is dispatched: for each `<spec>.mockups` sidecar with a line naming a changed `<frame id>.png`, the parent runs the compose command and the `measure-visual-properties.sh … --props bands,seams` read of step 10 (`skills/flow/visual-verify-verifier.md`) on every pair; a leftover is a defect that opens the next round, and clean composites close the round with no `-full` run, the last verifier's report standing for every path outside the mockups root (kan-1007: a mockups-only round re-ran the whole suite, 18m28s).
   The confirming re-run
   verifies a fresh build, never a stale stack: before this round's re-run reports the fix
   confirmed live, rebuild the stack from the branch HEAD — current source through the project's
   own build commands, brought up the way step 5 of **Steps 3–13** (`skills/flow/visual-verify.md`)
   starts one — and re-run the defect's exact scenario end to end against that build. The
   fingerprint step there proves only that a stack serves its worktree's build, never that the
   build is current source, and verification against a stale build is not verification. The
   rebuild is itself a defect-finder: a defect the fresh build exposes that the fix's own task
   does not cover is reported the way any stage-found defect is, never absorbed into the fix's
   confirmation.

**One full run closes the loop.** A scoped re-run whose final report carries no fixable defect is
followed by **Visual verification** once more from its `begin` mark, unscoped — no round base,
every motion of the change — its dispatch keys suffixed `-fix-<k>-full` before any `-<repository
name>` suffix. A fixable defect it reports opens the next round, whose step-4 re-run is scoped
like any round's, its round base the HEAD this full run verified: a `-full` run follows only a
clean scoped re-run, never another round's fix (KAN-1000: three rounds in a row re-ran unscoped).
A scoped re-run's verifier row is begun with `-round-base <sha>`, the round base its prompt
names, and `check-visual-verify-dispatched.sh` refuses a worktree whose newest completed verifier
row carries one — the closing unscoped run is not recorded.
After an `In-run fix <k>` task is ticked, and until a `-fix-<k>-full` verifier row has completed,
`flow record dispatch begin` refuses a `visual-verify` key that does not carry `-fix-<k>`, and one
that is not the `-full` run, carries no `-round-base` and shares its repository suffix with a
completed verifier row.
Its clean report is the one the handoff reports.

## No cap

**No round count ends the loop.** **Fewest operator actions** (`skills/flow-contracts/pipeline.md`)
is canonical for that, and for a round that makes no progress.
