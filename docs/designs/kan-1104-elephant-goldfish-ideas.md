# kan-1104-elephant-goldfish-ideas
> Design record of change kan-1104-elephant-goldfish-ideas (KAN-1104), written at landing. Each decision's own **Status:** line is the status.

## Context

- **One change.** The operator asked for the Elephant-Goldfish Model's ideas as one ticket,
  KAN-1104, and settled each across three planning rounds; the parts share the planning contract
  (`skills/flow/brainstorm.md`, `skills/flow/brainstorm-planner.md`) and are landed together.
- **It edits the contract it is planned under.** This plan is written, implemented and landed by
  the pipeline as it is today: its own plan is not reviewed by the reviewer it adds, and its own
  design record is not written at its landing unless the installed `flow-guard` already carries
  Task 5 — the first later change gets both.
- **Who dispatches.** A planner's Agent tool is limited to Explore and Plan agents, so the parent
  dispatches the reviewer; the parent already owns the `PLAN-WRITTEN` redispatch loop the
  findings re-enter, and runs **Decide**.
- **The reviewer's dispatch must pass `hooks/enforce-agent-baseline.py`**: a description opening
  with a key `ROW_KEY` admits, an open ledger row under that key, and a `Role:` line whose
  **Allowed pairs** row admits the pair.
- **The Experiments view draws every row of `skills/flow-contracts/model-policy.md`**, a declared
  `ui paths` entry, so a new role row changes its visual baselines.
- **Run-loaded prose is guarded sentence by sentence.** `flow-guard check-verbatim-moves` fails
  any added or reworded sentence under `skills/` not listed in this change's
  `verbatim-moves.txt`, so the prose edits are carried as patches under `patches/` beside this
  file, each verified on a scratch copy, and `verbatim-moves.txt` lists their sentences.
- **`## decisions: recommended` is this project's mode.** Two points the approved design left
  open surfaced while planning and were taken by their recommended option under it — decisions
  `record-needs-a-subject` and `record-header-without-jira`.
- **`kan-1103-allowed-pairs-ui` is in flight** and rewrites the **Allowed pairs** tables to two
  columns. Whichever lands second re-applies its row edit; nothing here depends on it.

## Recommended approach

One context-free reviewer, dispatched by the parent after the final planner returns and before
**Decide**, inside a stage of its own, with its findings routed back through the planner — and
the smaller items each placed where its rule already lives. Rejected alternatives, by decision:

- reviewing at design approval, or at both points — `plan-review-after-artifacts`;
- all eight lenses, or the minimum two — `seven-lenses`;
- giving the reviewer the ticket or the planner's files — `reviewer-reads-documents-only`;
- the operator triaging every finding, or the planner fixing all unasked — `triage-fix-contest-escalate`;
- a review loop until clean — `two-plan-reviews`;
- running inside `flow.writing-plans` — `plan-review-stage`;
- vendoring the standards — `one-rubric-file`;
- a plan task, an archive commit or an opt-in key for the design record — `design-record-at-run1`.

## Critical files

- `skills/flow/brainstorm.md` — the parent's planning relay; gains **Plan review**, the summary
  relay and the one-per-call rule.
- `skills/flow/brainstorm-planner.md` — the planner's sections; gains the three dialogue rules,
  the two `design.md` sections, the planner's **Plan review** duties and the moved
  `flow.writing-plans` end.
- `skills/flow/plan-reviewer-prompt.md`, `skills/flow/plan-review-rubric.md` — new.
- `skills/flow/stage-keys.md`, `skills/flow-plan/SKILL.md` — name the new stage key.
- `skills/flow/review-panel.md`, `skills/flow/reviewer-calibration.md`,
  `skills/flow/principles-reviewer-prompt.md`, `skills/flow/primary-reviewer-prompt.md` — one
  sentence each.
- `skills/flow-contracts/model-policy.md` — the `design-reviewer` row.
- `skills/flow-contracts/finish-contract-run1.md`, `skills/flow-contracts/pipeline.md`,
  `skills/flow-contracts/artifacts-registry.md`, `skills/flow-contracts/finish-contract-rationale.md`
  — the design record and the dated reversal.
- `hooks/enforce-agent-baseline.py` — `ROW_KEY`.
- `stats/internal/stages/names.go`, `README.md` — the stage table and its Level 1 mirror.
- `stats/internal/store/aggregate.go`, `stats/internal/store/records.go`,
  `stats/internal/store/experimentstats.go` — weekly cost category, auto-opened stage, role rows.
- `stats/web/src/lib/phase.ts`, `stats/web/src/lib/stageLabel.ts`,
  `mods/flow-task-list/hooks/register.tsx` — the key's phase, label and task-band phase.
- `stats/internal/guard/planshape.go` — F19's section list.
- `stats/internal/guard/integratechange_commit.go`, `stats/internal/guard/integratechange_prepare.go`
  — the record's write and the subject it needs.
- `stats/web/tests/visual/experiments.spec.ts-snapshots/`,
  `stats/web/tests/visual/full-app-suite.spec.ts-snapshots/`,
  `stats/web/tests/visual/full-app-suite.zip` — baselines the new role row changes.

## Existing code to reuse

- `gateBrainstormDispatchesClosed` (`stats/cmd/flow/stage.go`) — already refuses
  `flow stage begin flow.decide` while a dispatch row whose key opens `plan-` is unended, so an
  open `plan-review-<n>` row holds **Decide** with no new gate.
- The `PLAN-WRITTEN` redispatch loop and the `QUESTIONS` relay of `skills/flow/brainstorm.md` —
  the findings and escalations re-enter planning through them.
- `psDesignSections` and `designSections()` (`stats/internal/guard/planshape.go`) — F19 gains two
  headings, no new check.
- `commit-split` (`stats/internal/guard/commitsplit.go`) stages every path outside the change
  directory with `git add -A`, so the record needs no `git add` of its own.
- `c.jiraKey`, `c.canon`, `c.implPaths`, `c.linkOnly`, `c.stop` (`stats/internal/guard/integratechange.go`).
- `roleDispatchSQL` and `roleMetrics` (`stats/internal/store/experimentstats.go`).
- The gated reviewer's stub-plus-brief shape (`skills/flow/gated-reviewer-brief.md`): the prompt
  names a file the reviewer reads itself, so the parent never loads the reviewer prompt.
- `flow-guard live-store-copy` — the live check's stand-in store.

## Decisions

### The review reads the finished artifacts, once

**ID:** plan-review-after-artifacts
**Status:** active
**Chosen:** one review after writing-plans and its guards, before Decide — the only point where the three artifacts exist, so the "a fresh engineer executes this without a choice of his own" test can run.
**Considered:** a review of the design draft before the design approval as well (two dispatches per change, and the tasks do not exist yet); before the approval only (the plan stays checked by guards alone); on `design.md` right after approval (same document, findings arrive after the approval and cannot judge tasks).

### Seven lenses in one bundled reviewer

**ID:** seven-lenses
**Status:** active
**Chosen:** comprehension, zoom-out, design critic, implementation readiness, codebase details, the inventive catalyst as advisory-only, and a four-question proposal exam — the lenses with no flow counterpart, in one subagent.
**Considered:** the five without the inventive catalyst (the operator wanted its trade-off pass); one lens per artifact (misses claims the code contradicts); all of the reference's lenses, structural and style included (duplicates F19 and the prose guards); an A/B double dispatch with a synthesizer (twice the cost for recall nobody measured).

### The reviewer reads the documents, the rubric and the repository — nothing else

**ID:** reviewer-reads-documents-only
**Status:** active
**Chosen:** the three artifacts, the rubric and read access to the repository; no ticket, no Jira key, no round or answers file — a design whose `## Context` does not carry the objective fails on that.
**Considered:** adding the ticket (the reviewer inherits the request's framing); everything under the planner directory (a second opinion, not an isolated one).

### The planner fixes, contests or escalates each finding

**ID:** triage-fix-contest-escalate
**Status:** active
**Chosen:** a fresh planner fixes a finding with one obvious fix, contests one with evidence the reviewer could have read, declines an advisory with a reason, and escalates as a `QUESTIONS` round any finding whose fix would change a `## Decisions` entry or that it cannot refute from such evidence.
**Considered:** every finding shown to the operator (a stop per finding on every non-trivial change); the planner fixing every finding unasked (a reviewer without the ticket would overturn operator decisions).

### One re-review, only when a decision changed

**ID:** one-re-review
**Status:** superseded by two-plan-reviews
**Chosen:** the plan is reviewed a second time only when a fix or an escalation's answer added or superseded a `## Decisions` entry; each finding still standing after that triage is written into `## Risks` as accepted, and nothing is reviewed a third time.
**Considered:** looping until two clean passes (unbounded cost on a reviewer that always finds an advisory); never re-reviewing (a changed decision reaches the implementer unread).
**Superseded because:** the operator's instruction during implementation: review once when the initial artifacts are created and once when all of them are done and the plan is ready for implementation, not a second time only when a decision changed.

### Two reviews on every non-trivial plan: once the artifacts exist, once the plan is ready

**ID:** two-plan-reviews
**Status:** active
**Chosen:** the plan is reviewed by two fresh reviewers at two points: `plan-review-1` once the artifacts are created — `proposal.md` and `design.md` written, `tasks.md` still a scaffold, before writing-plans — and `plan-review-2` after all the artifacts are done and the plan is ready for implementation, before Decide. Each review is triaged by a fresh planner; the findings still standing after the second triage are written into `## Risks` as accepted, and nothing is reviewed a third time. The early planner returns `ARTIFACTS-WRITTEN`, a fresh final planner then writes the plan, and `flow.plan-review` is marked around each review.
**Considered:** a second review only when a decision changed (the superseded `one-re-review`: a clean first review is never cross-checked, and a design flaw is found only after the plan is written on top of it); two reviews of the finished plan (the first adds nothing the second would not catch); looping until two clean passes (unbounded cost on a reviewer that always finds an advisory).

### `/flow` and `/flow-plan`, on a change that is not trivial

**ID:** plan-review-scope
**Status:** active
**Chosen:** the creating run of `/flow` and a `/flow-plan` capture, under the competing-designs test `skills/spectre-planning/SKILL.md` already applies; the final planner's return carries the verdict as a `Plan review: skip — <reason>` line, so the parent derives nothing.
**Considered:** every change (a review of a one-approach change finds only advisories); the self-review and in-run fix planners too (their plans fix named findings and are reviewed by the loop's own reviewer); the parent judging triviality (it has not read the design).

### The parent dispatches the reviewer

**ID:** parent-dispatches-reviewer
**Status:** active
**Chosen:** the parent makes the dispatch and both stage marks — a planner's Agent tool is limited to Explore and Plan agents, and the redispatch loop is the parent's.
**Considered:** the planner dispatching it (needs a third agent kind on the planner definition, and the planner would then triage inside the context that wrote the plan).

### `opus` at `flow-low`, a fixed `design-reviewer` row

**ID:** design-reviewer-row
**Status:** active
**Chosen:** a new unmarked row of **Allowed pairs**, `design-reviewer` — `opus/flow-low`, key `plan-review-<n>`, ledger role `reviewer`, recorded in no `decision.json` field, as the merger's pair is not.
**Considered:** `opus` at medium (a rubric read is short); the panel's first-pass pair with its CORE `fable` exception (extends an exception the operator scoped to three roles).

### A stage of its own, `flow.plan-review`

**ID:** plan-review-stage
**Status:** active
**Chosen:** a new stage valid for `/flow` and `/flow-plan`, begun and ended around each of the two reviews — once between `flow.create-artifacts` and `flow.writing-plans`, once between `flow.writing-plans` and `flow.decide` — each run spanning its review, its triage and any escalated round; `completed` whenever a verdict was reached, `failed` with reason `plan review not run` when the reviewer failed twice, and no mark at all on a skipped review. The `flow.writing-plans` end moves ahead of it.
**Considered:** running inside `flow.writing-plans` (the recommended option; no stage-table edits, but the operator wanted the review as its own leaderboard bar).

### One flow-owned rubric file

**ID:** one-rubric-file
**Status:** active
**Chosen:** `skills/flow/plan-review-rubric.md`, adapted from the reference's standards and personas with its source and licence named, citing flow's own rule wherever flow already states it.
**Considered:** the standards vendored verbatim with an upstream note (a second statement of rules flow owns); the lenses folded into the prompt (the planner and the reviewer then read different texts).

### The reviewer's prompt is a stub naming a file it reads

**ID:** reviewer-prompt-is-a-stub
**Status:** active
**Chosen:** the dispatch prompt carries the role, key, baseline, MODEL HANDSHAKE and six absolute paths; the instructions live in `skills/flow/plan-reviewer-prompt.md`, which the reviewer reads first — the gated reviewer's shape, so the parent never loads the prompt body. No `check-dispatch-paragraphs` site row is added: that guard checks a fixed table and requires nothing of a new file.
**Considered:** the parent pasting the prompt body with four values filled in (loads a subagent-facing file into the parent's context, which `skills/flow/review-panel.md` forbids for its own prompts).

### Four sentences in the existing reviewer prompts

**ID:** four-reviewer-sentences
**Status:** active
**Chosen:** `not run` is never zero findings (`skills/flow/review-panel.md`); a lens with nothing to report says so (`skills/flow/reviewer-calibration.md`); search the whole repository before judging new code (`skills/flow/principles-reviewer-prompt.md`); edited prose the code refutes, worded as a Critical finding because that is the primary reviewer's own scale (`skills/flow/primary-reviewer-prompt.md`). Operator-directed, so no `/flow-prompt-tuning` replay.
**Considered:** adding the churn test and claim-matches-evidence (more prompt text with no observed miss behind it); the anti-flattery block in every prompt (flow's prompts already ask for findings only); none.

### An architectural question says what each option fails under

**ID:** fails-under
**Status:** active
**Chosen:** a question whose options are approaches is headed `[architectural]`, every option closes with `Fails under:`, and the recommended mark stays so unattended runs remain possible.
**Considered:** no recommended mark and no default on such a question (it would stop every run, the recommended mode included); no change.

### Architectural questions are asked one per call

**ID:** architectural-one-per-call
**Status:** active
**Chosen:** the parent asks each `[architectural]` question in a call of its own, in round-file order; every other question stays four to a call.
**Considered:** four per call everywhere (the recommended option; the operator wanted an approach decided alone); one per call everywhere (many more turns for names and formats).

### A contradicting answer is asked about, in planning only

**ID:** contradiction-audit-planning-only
**Status:** active
**Chosen:** when an answer contradicts an earlier answer or a recorded decision, the planner changes nothing on that point, quotes both and asks which stands with no recommended option, so the recommended mode asks it too.
**Considered:** the same past the plan gate (a pivot would stop an unattended run the operator opted into); no change.

### Chat carries a round's summary, the file carries the text

**ID:** round-summary-relay
**Status:** active
**Chosen:** a round file opens with `## Summary`, at most ten lines, and the relay prints that section and the file's absolute path instead of the design draft; the plan-gate summary gains `**Design:**` and `**Plan review:**` lines.
**Considered:** publishing the design and the review as Artifacts each round (a publishing step per round for text the operator can open as a file); no change.

### `design.md` requires `## Non-goals` and `## Risks`

**ID:** non-goals-and-risks
**Status:** active
**Chosen:** both headings join `psDesignSections`, so F19 refuses a plan-mode design missing either; presence only, each may be one `none — <reason>` line, and the content is the reviewer's to judge.
**Considered:** the reviewer checking for them with no required section (a reviewer cannot flag what no section invites); no change.

### Run 1 writes the design record

**ID:** design-record-at-run1
**Status:** active
**Chosen:** `flow-guard integrate-change commit` writes the canonical worktree's `design.md` to `docs/designs/<name>.md` — a fixed two-line header with no status value, a blank line, the file byte for byte — in every project, where the implementation commit carries it; derived on every run of the step, never hand-edited, never kept current afterwards. This reverses kan-1072 for `design.md` alone.
**Considered:** a follow-up ticket (the recommended option; the operator wanted it now); a plan task that copies the file (stale the moment a fix round supersedes a decision); restoring the archive commit (two landings for one file); an opt-in `## design records` key or an `agents`-only check (configurability one project asked for, or a project-name test in a generic guard); a mutable status header (no guard would keep it true); a drift-report skill (a change of its own size).

### A worktree whose only implementation path is the record still takes a subject

**ID:** record-needs-a-subject
**Status:** active
**Chosen:** `integrate-change prepare` asks for the canonical worktree's implementation subject whenever its change directory holds a `design.md`, so `commit` never stops on a record nobody was asked to name a commit for. Taken by its recommended option under `## decisions: recommended`; the approved design did not cover a canonical worktree with no other implementation path.
**Considered:** skipping the record there (the design is lost for exactly the changes whose code lives in other repositories); a fixed literal subject as `link.md` has (an invented commit subject for a file that is not bookkeeping).

### A change with no Jira key drops the parenthesis

**ID:** record-header-without-jira
**Status:** active
**Chosen:** the header's second line reads `> Design record of change <name>, written at landing. …` when the state row holds no key. Taken by its recommended option under `## decisions: recommended`.
**Considered:** an empty `()` (reads as a defect); refusing to write the record (a project without Jira would get none).

### An auto-opened run for a plan-review row is `flow.plan-review`

**ID:** plan-review-auto-stage
**Status:** active
**Chosen:** `autoStageFor` maps a `plan-review-` key to `flow.plan-review`, ahead of its `plan-` case, so a row begun with no stage open is not filed under `flow.brainstorm`.
**Considered:** leaving the `plan-` case to catch it (the store would record planner time that was review time).

## Open questions

## Verification

- `flow-guard run-verify-list <worktree> lint` and `flow-guard run-verify-list <worktree> test`
  exit clean on the finished branch.
- The plan reviewer's dispatch is admitted: `TestEnforceAgentBaselineHookAdmitsPlanReviewKey`
  passes, and `TestEveryRoleLineIsAPolicyRow` finds the `design-reviewer` row.
- `go run ./cmd/flow stage keys` in `stats/` prints `flow.plan-review` between
  `flow.writing-plans` and `flow.decide`.
- `TestIntegrateChangeDesignRecord` shows the record in the implementation commit, header first.
- The Experiments view's roles panel shows a `design-reviewer` row under Review in the recaptured
  baselines, and `flow.visual-verify` passes against them.
- The first later non-trivial `/flow` change, after the reinstall this repository's `CLAUDE.md`
  requires, shows a `flow.plan-review` stage run and a `plan-review-1` dispatch row, and lands a
  `docs/designs/<name>.md`.

## Live check

Run from the change's worktree, against a copy of the dev store:
`scripts/flow-guard live-store-copy up kan-1104-live`, then, with the exports it prints:

- `flow stage begin -command '/flow' -stage flow.plan-review … kan-1104-live` exits clean. Not
  working looks like a refusal naming an unknown stage key.
- `flow record dispatch begin -change kan-1104-live -role reviewer -model opus -effort low -key
  plan-review-1 …`, then `flow stage begin … -stage flow.decide …` — the second is refused,
  naming `plan-review-1`. Not working looks like the begin succeeding while the row is open.
- End the row and the stage; `flow stage begin … -stage flow.decide …` now succeeds, and the copy's
  stage runs for the scratch change list `flow.plan-review`, not `flow.brainstorm`.

Then `scripts/flow-guard live-store-copy down kan-1104-live`. The design record has no live
stand-in beyond `TestIntegrateChangeDesignRecord`, which runs `integrate-change` end to end over
scratch repositories; running it against this change's own integrate state would write that state.

## Non-goals

- No review at design approval; no style or structural lens; no A/B double dispatch; no synthesizer.
- No drift report, no drift skill, no backfill of earlier designs, no status upkeep on a record.
- No change to auto-resolution past the plan gate, to the four-per-call rule for ordinary
  questions, or to `/flow`'s gate taking **Yes**.
- No Artifact publishing of the design or the review.
- No plan review, and no `flow.plan-review` mark, in `/flow-fast`, `/flow-interactive`, a fix run
  or a self-review or in-run fix loop.
- No `decision.json` field, no migration, no `uitest-seed` row for the new stage.

## Risks

- **Cost.** One `opus` low dispatch per non-trivial change and, on any finding — an advisory
  included, which the inventive catalyst usually produces — one planner redispatch; bounded by
  the second review.
- **An isolated reviewer flags what the ticket already answers.** That is the intended test, and
  a `contested` line needs evidence in a file the reviewer could read, so the usual fix is a
  sentence in `## Context`.
- **`flow.plan-review` includes operator wait** on an escalated question, as
  `flow.design-approval` does; the reviewer's own time is its `plan-review-<n>` dispatch rows.
- **Plans in flight fail F19 once this lands** until their `design.md` gains the two sections:
  `kan-1103-allowed-pairs-ui` and `self-review-kan-1083-s1-white-design-system-dark-theme` carry
  `Design shape: plan-mode` today. Each needs two `none — <reason>` lines, or real content.
- **Design records age.** Nothing checks one against the code, and a consuming project that
  lints or publishes everything under `docs/` receives a file its tooling may reject.
  `flow-guard check-done-when-paths` reads every tracked Markdown file, so a design that carried
  a `## Done when` section would hold its paths to later renames; no planning prompt writes one.
- **The task band's tests cannot be run in this repository** — `register.test.tsx` imports the
  harness's own test module — so Task 1's edit there is checked by reading.
