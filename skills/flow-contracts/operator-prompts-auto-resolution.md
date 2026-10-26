# Auto-resolution

The auto-resolution rules of **Operator prompts** (`skills/flow-contracts/operator-prompts.md`).

**During implementation and every fix run, a prompt with a recommended option is never asked: the
run takes the recommended option itself, at once.** The phases are everything from `flow.implement`
to the `IN_PROGRESS` handoff of a creating run — `skills/flow/implement.md`,
`skills/flow/review-panel.md`, `skills/flow/visual-verify.md`, `skills/flow/verify-and-handoff.md`
— and every fix run (`/flow <fix instructions>`, or a plain message run as a fix). The operator
asked for this once, for every flow run: a mid-run question they would answer with the recommended
option costs them a stop for nothing.

It holds for a call site's own prompt and for any question the run frames itself in those phases —
a finding that seems to need a product decision, a departure from a mockup — one whose root
cause is a shared style, token or component, whose fix moves other screens' baselines, included
(**The loop**, `skills/flow/verify-fix-loop.md`, fixes it at that source) — a
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
are auto-resolved exactly as the implementation phases are. The key absent, or a head that
resolves to nothing, leaves the scope exactly as the paragraph above states it. The remaining
**What still stops** bullets are the mode's limits: a prompt with no recommended option, nothing
to choose, anything outward-facing or irreversible, and every prompt outside the run's own phases
are asked under the mode exactly as off it.

**A blocked verify or visual-verify is never turned into questions.** Each class below has one
answer the pipeline's own rules already give, so the run takes it and goes on to **The loop**
(`skills/flow/verify-fix-loop.md`) without asking. A list of "Decisions for you" built from these
cases is a stop for nothing (KAN-1000):

- **A failure this branch introduced**, such as a capture's seed data crowding another spec's search:
  fix the branch so the other spec passes again, with the smallest change that does it (rename the
  seed, for example).
- **Failures that are already on the base:** attribute them to their cause first, then fix them in this
  run (**The sweep**, `skills/flow-contracts/known-bugs.md`). If the cause is a base branch that has
  fallen behind the branch it tracks (a feature branch that never took in develop or main), the
  recommended option is to sync that base and re-run. The parent checks this at the start of
  whichever of `flow.verify` and `flow.visual-verify` runs first, before any failure arrives: for
  each repository whose recorded base is a feature branch, after `git fetch origin`, a
  `git rev-list --count origin/<base>..origin/<the branch it tracks>` above 0 — develop or main
  holding commits the base lacks — syncs the base first (KAN-1000: a stale feature/cardio base
  cost 61 failures and about an hour). `check-base-behind.sh` makes that count for every worktree of the change and prints the sync commands (**Verify**, `skills/flow/verify-and-handoff.md`). Pushing the synced base is not asked (below).
  **Never merge the tracked branch into a change branch** (kan-1007): the change branch then carries the tracked branch's commits as its own, and integrate's reshape folds them into the change's squash — the base is what gets synced.
- **A departure from the mockup:** match the mockup (`rules/design-mockups-are-specs.mdc`). An element
  the mockup does not draw stays only when the screen cannot work without it. Record that reason
  with the frame.
- **A declared frame that was not captured:** add the capture. A missing capture means the report is
  incomplete. It is never a scope choice.
- **The ticket or a design decision conflicts with a project rule** — a requested entrance motion
  against a no-blank-first-frame rule, for example: take the reading that satisfies both (the row
  drawn from its first frame and sliding into place, never starting invisible), record it as a
  decision superseding the plan's, and go on to **The loop**. A conflict is never a reason to stop.

**A blocked verify never ends the turn.** The report of a blocked `flow.verify` or
`flow.visual-verify` is the input to **The loop**, dispatched in the same turn; a handoff that
lists the failures and stops, or ends on a "Decision for you" whose recommended option the run
could have taken, is the stop this file exists to prevent (kan-1007: the run stopped after
visual-verify to ask whether a requested rise could start visible).

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

These are asked, or stop with `## Question`, exactly as their call sites state:

- **Planning.** Brainstorm and design questions, the convergence-and-approval confirm, the
  third-round offer, the plan
  review gate, every `/flow-plan` prompt, a fix run's own planning pass (the re-plan-budget and
  where-the-fix-goes prompts in `skills/flow/document-fix.md`). The operator scoped
  auto-resolution to implementation and fix options, not planning. The `## decisions: recommended` mode lifts the planning asks in this bullet and lifts no
  other bullet here; the withdrawal offers in `skills/flow/withdrawal.md` stay asked under the
  mode too, deleting a change being irreversible and the bullet below governing.
- **No recommended option.** The second model-handshake mismatch, a multiple-match pick, a finding
  recorded unverifiable, and a question the run cannot honestly give one recommended option.
- **Nothing to choose.** A guard's exit 2 (it cannot answer), a command that failed twice on an
  environment cause the run cannot cure itself, and every other `## Question` handback that carries no options.
- **Outward-facing or irreversible actions.** Push, merge, opening a PR, archiving, deleting, and
  every Jira write — the follow-up filing and the self-review filing included — and the
  self-review's flow-eval mining.
  The global rules require these confirmed.
  One push is not asked: creating a `--base` a repository's origin lacks, from that repository's default branch — kickoff's (`skills/flow/SKILL.md`) and integrate's `prepare` (`BASE-CREATED:`, `skills/flow/integrate.md`) — adds a branch at the default branch's own tip, overwrites nothing, and is what the operator's `--base` already asked for.
  Nor is the base sync `check-base-behind.sh` prints: a merge commit of the tracked branch onto the
  `--base`, pushed fast-forward — it rewrites no history and only brings the base up to what it
  tracks (gymie KAN-1015: the run stopped mid-verify to ask it, and the operator said to answer it
  automatically).
- **Outside implementation and fix runs.** The wrong-state override, the plain-message ambiguity
  prompt that decides whether a fix run starts at all, and every integrate and cleanup prompt.
