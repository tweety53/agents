# Auto-resolution

The auto-resolution rules of **Operator prompts** (`skills/flow-contracts/operator-prompts.md`).

**Past planning, a prompt with a recommended option is never asked: the run takes the recommended
option itself, at once.** Past planning is everything from `flow.implement` to the run's final
handoff — `skills/flow/implement.md`, `skills/flow/review-panel.md`,
`skills/flow/visual-verify.md`, `skills/flow/verify-and-handoff.md`, `skills/flow/integrate.md`,
`skills/flow/cleanup.md` and the self-review pass (`skills/flow-self-review/SKILL.md`) — every fix
run (`/flow <fix instructions>`, or a plain message run as a fix), `/flow-fast` from its
implementation on, and the prompts that open a run. The operator asked for this for every flow run
(**Work end to end — stop only to plan or to test by hand**, `rules/work-end-to-end.mdc`): a
question they would answer with the recommended option costs them a stop for nothing.

It holds for a call site's own prompt and for any question the run frames itself in those phases —
a finding that seems to need a product decision, a departure from a mockup — one whose root
cause is a shared style, token or component, whose fix moves other screens' baselines, included
(`skills/flow/fix.md` fixes it at that source) — a
question an implementer's report carries, a pivot (`skills/flow/implement.md`).
A pivot is auto-resolved although it alters approved scope: the operator asked that a recommended
option be taken even when it reopens an earlier choice of theirs. Frame it with the option the
pipeline's own rules favour marked recommended (for a mockup departure, matching the mockup, per `rules/design-mockups-are-specs.mdc`),
and take that option.

**A child's hand-back never reaches the operator as an ask.** A fixer's or implementer's "please
spot-check", "your call" or any decision its report leaves open is resolved inside the run: a check
it asks for — a regenerated golden or snapshot among them — is made by the loop's reviewer, its
dispatch naming each regenerated file and the one change each may carry; a decision left over is
taken by its recommended option by the parent and recorded as below, never listed as open.

**The `## decisions: recommended` mode widens the scope to planning.** When the project's
`## decisions` key resolves to `recommended` (**Project configuration**,
`skills/flow-contracts/project-configuration.md`), the paragraph above holds for the whole run:
the planning asks the Planning bullet of **What still stops** below names
are auto-resolved exactly as the phases past planning are. The key absent, or a head that
resolves to nothing, leaves the scope exactly as the paragraph above states it. The remaining
**What still stops** bullets are the mode's limits: the manual testing the operator asked for, a
prompt with no recommended option (an undrawn visible element, asked in planning per **Every visible element a task builds is drawn before the design is approved**, `skills/flow/brainstorm-planner.md`, is one), a cause out of the run's reach and anything outward-facing or
irreversible stop under the mode exactly as off it.

**A blocked verify or visual-verify is never turned into questions.** Each class below has one
answer the pipeline's own rules already give, so the run takes it and goes on to the fix procedure
(`skills/flow/fix.md`) without asking. A list of "Decisions for you" built from these
cases is a stop for nothing:

- **A failure this branch introduced**, such as a capture's seed data crowding another spec's search:
  fix the branch so the other spec passes again, with the smallest change that does it (rename the
  seed, for example).
- **Failures that are already on the base:** attribute them to their cause first, then fix them in this
  run (**The sweep**, `skills/flow-contracts/known-bugs.md`). If the cause is a base branch that has
  fallen behind the branch it tracks (a feature branch that never took in develop or main), the
  recommended option is to sync that base and re-run. The parent checks this at the start of
  whichever visual or lint stage of `flow.verify` runs first, before any failure arrives: for
  each repository whose recorded base is a feature branch, after `git fetch origin`, a
  `git rev-list --count origin/<base>..origin/<the branch it tracks>` above 0 — develop or main
  holding commits the base lacks — syncs the base first. `flow-guard check-base-behind` makes that count for every worktree of the change and `flow-guard sync-base-behind` syncs them (**Verify**, `skills/flow/verify-and-handoff.md`). Pushing the synced base is not asked (below).
  **Never merge the tracked branch into a change branch**: the change branch then carries the tracked branch's commits as its own, and integrate's reshape folds them into the change's squash — the base is what gets synced.
- **A departure from the mockup:** match the mockup (`rules/design-mockups-are-specs.mdc`). An element
  the mockup does not draw stays only when the screen cannot work without it. Record that reason
  with the frame.
- **A declared frame that was not captured:** add the capture. A missing capture means the report is
  incomplete. It is never a scope choice.
- **The ticket or a design decision conflicts with a project rule** — a requested entrance motion
  against a no-blank-first-frame rule, for example: take the reading that satisfies both (the row
  drawn from its first frame and sliding into place, never starting invisible), record it as a
  decision superseding the plan's, and go on to the fix procedure. A conflict is never a reason to stop.
- **A verify or visual-verify loop that keeps finding departures** — round after round, the same
  frames again or new ones: run another fix round. There is no round count, no usage or time
  threshold, and no "land with leftovers or keep going?" question: the departures are fixed until a
  round reports none (`skills/flow/fix.md`). **The run never lands, and never writes
  its handoff, with a departure open** — no "leftovers", no "known departures" list. Two frames that
  contradict each other are settled by taking the frame of the screen's own section (the one the
  ticket names for that screen), recording that as a decision superseding the other frame's reading,
  and fixing to it; the other frame's departure is then closed, not carried.
- **A guard that refuses evidence the run really holds** — a freshness check that reads a
  byte-identical recapture as "not recaptured", a closing check a remainder row cannot satisfy: the
  guard is the defect. Fix it in an `agents` worktree (**Pipeline defects found mid-run**,
  `skills/flow-contracts/pipeline.md`), with a test that fails without it, land it and re-run the
  step. It is never a `## Question`, never "RUN FULL or fix the guard?", and never a request for the
  operator's go-ahead: the operator's "never a full run unless asked" is kept by fixing the guard,
  not by asking which way to pass it. Faking the evidence — touching a file, editing a ledger row —
  stays forbidden.

**A blocked verify never ends the turn.** The report of a blocked `flow.verify` or
`flow.verify` is the input to the fix procedure, dispatched in the same turn; a handoff that
lists the failures and stops, or ends on a "Decision for you" whose recommended option the run
could have taken, is the stop this file exists to prevent.

**A scope or cost question with a recommended option is never asked during integrate or the
self-review either** (operator). "Fix every finding (recommended), or only the
blocker?", whether to run one more fix or review round, whether a fix is worth its cost: the run
takes the recommended option, records it as below and goes on. A question written as prose in a
message, a final message's included, is an ask exactly as an `AskUserQuestion` call is. The
outward-facing and irreversible actions of **What still stops** below are still asked.

Every auto-resolution:

- is recorded in the pass log when the mode is on — a `flow
  record pass -round <round> -note 'auto-resolved: <question> → <option>'` row, round 0 at every
  site outside the review panel, the same row shape an operator answer is recorded in; off the
  mode, it is recorded where the call site records an operator's answer — in the review panel, that
  same row — and a site that records no answer records nothing new
- is named in the handoff's `**Decisions:**` lines, 🤖-marked, question and option taken, so the
  operator can overrule it afterwards with a fix run
- never repeats: the same prompt arising again for the same subject after its recommended option
  was already taken this run is asked, so an auto-taken **another round** cannot loop — except a
  fix, re-review or re-verify round, which **Fewest operator actions**
  (`skills/flow-contracts/pipeline.md`) leaves unbounded

### What still stops

These are asked, or stop with `## Question`, exactly as their call sites state — and nothing else
past planning is (**Work end to end — stop only to plan or to test by hand**,
`rules/work-end-to-end.mdc`):

- **Planning.** Brainstorm and design questions, the convergence-and-approval confirm, the plan
  review gate, every `/flow-plan` prompt. The operator scoped
  auto-resolution to implementation and fix options, not planning. The `## decisions: recommended` mode lifts the planning asks in this bullet and lifts no
  other bullet here.
- **Manual testing the operator asked for.** `/flow-fast` with `## handoff` `required` stops after
  its change summary; the fixes that testing finds are the operator's fix runs. It stops only when
  the operator set it. A `/flow` run that leaves a `?` in its `Decisions:` stops at its `IN_PROGRESS`
  handoff (**What still stops**).
- **No recommended option.** A multiple-match pick, a finding
  recorded unverifiable, and a question the run cannot honestly give one recommended option.
- **Out of the run's reach.** A failure whose cure lies outside the run — production, a credential
  or account only the operator holds, a repository the run was not given — and a `## Question`
  handback that carries no options. A guard's exit 2, a failing command or tool, a stack that will
  not start and an environment fault the run's own commands can repair are not this bullet: they
  are cured in the run (**Fewest operator actions**, `skills/flow-contracts/pipeline.md`). Nor is
  the flow tooling itself: a guard, script, skill or contract of the `agents` repository that blocks
  the run — two guards that contradict, a check that refuses correct work — is fixed in an `agents`
  worktree with a test that fails without the fix, pushed to `main`, and the run continues; that push
  is not asked (`rules/work-end-to-end.mdc`).
- **Outward-facing or irreversible actions.** Push, merge, opening a PR, deleting, and
  every Jira write — the follow-up filing included.
  The global rules require these confirmed.
  One push is not asked: creating a `--base` a repository's origin lacks, from that repository's default branch — kickoff's (`skills/flow/SKILL.md`) and integrate's `prepare` (`BASE-CREATED:`, `skills/flow/integrate.md`) — adds a branch at the default branch's own tip, overwrites nothing, and is what the operator's `--base` already asked for.
  Nor is the base sync `flow-guard sync-base-behind` runs: a merge commit of the tracked branch onto the
  `--base`, pushed fast-forward — it rewrites no history and only brings the base up to what it
  tracks.
  Nor is the self-review pass's filing (step 4, `skills/flow-self-review/SKILL.md`): it creates
  new issues only, by the operator's standing decision, and touches no existing one.
