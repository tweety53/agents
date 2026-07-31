## Why

Three of this pipeline's surfaces have no normative source at all, and each has already produced a
defect that nothing caught.

**Jira.** `jira-integration.md` is contract text with no capability behind it in `openspec/specs/`.
It transitions an issue to In Review *"once a PR is confirmed open"*, so a change that lands by the
merge-and-push route never reaches In Review — KAN-20's changelog records `To Do → In Progress`
then `In Progress → Done`, with no third transition. Its forward-only ordering also matches statuses
by name, so a workflow status outside the four known names has no defined position: KAN-13 and
KAN-17 both sit in `TO DO URGENT` and cannot be ordered against `In Progress` at all.

**The integration gate.** Nothing checks that a branch is finished before it is integrated. The
kan-6 branch merged carrying an unfixed defect and unticked manual-test boxes, and that is knowable
today only because it was typed into a chat transcript. Nothing in the repository records it.

**The review panel's bar.** "Zero open Critical/Important findings" lives only in
`myflow-do/SKILL.md`'s guardrails, in prose, with findings recorded as free text — so a minor
finding can be deferred with nothing able to notice.

Alongside these, the diff presented for human review is filtered rather than clean: planning
artifacts are staged and then hidden by a pathspec at display time, so they reappear in any other
view of the staging area.

## What Changes

- **Jira becomes normative.** A new capability holds the transition points, forward-only ordering,
  the unknown-status ask, labels on created issues, append-only description sync, and the invariant
  that Jira is never a gate.
- **In Progress fires when planning begins** — `/myflow-start` transitions immediately after the
  key resolves, before brainstorming, instead of after the state write. Failure still degrades to
  one skipped-with-reason line.
- **In Review fires at the end of finish run 1 on every landing route** — pull request, merge and
  push, or manual. **BREAKING** relative to the current contract, which conditions it on a PR.
- **An unrecognised current status is asked about, never inferred** from `statusCategory`.
- **Issues myflow creates carry the parent issue's labels plus `AI-generated`.** No label is
  invented.
- **A new run-1 gate refuses to integrate silently over unfinished work.**
  `scripts/check-unfinished-work.sh` reports unticked manual-test boxes, unchecked plan items,
  findings that are not closed, and a `## Known incomplete` section written by `/myflow-do`. The
  operator is offered continue, stop, or file a Jira task — and choosing to continue is recorded in
  the handoff and in the planning commit's message.
- **Review findings become countable.** The panel record gains a required table whose rows carry a
  status, and `/myflow-do` may not hand off while any row is open — at any severity.
- **Planning artifacts leave the staged diff.** `/myflow-do` stops staging `openspec/`,
  `docs/manual-test/` and `docs/superpowers/`; finish run 1 commits implementation first and those
  paths in a second commit. **BREAKING** relative to the current display-filter rule.
- **Cleanup becomes one registry** in `pipeline.md`, and the remote branch — which nothing deletes
  today — is removed at run 2 and verified by `scripts/check-cleanup-complete.sh`.
- **`/myflow-start` asks rather than assumes.** A new capability requires every unresolved question
  to be put to the operator, and every approval or choice to be offered as options rather than open
  prose.

## Capabilities

### New Capabilities

- `myflow-jira-projection`: how the pipeline projects its state onto a tracker issue — transition
  points, ordering, the unknown-status ask, labels on created issues, description sync, and the
  never-a-gate invariant.
- `myflow-planning-gate`: how `/myflow-start` conducts its planning — unresolved questions are put
  to the operator rather than resolved by assumption, and approvals are offered as options.

### Modified Capabilities

- `myflow-finish-cleanup`: run 1 gains the unfinished-work gate before it asks how the branch
  should land; run 2 gains remote-branch removal and a verification that the cleanup registry is
  empty.
- `myflow-handoff-output`: the requirement that planning artifacts are excluded from the review
  diff becomes a requirement that they are never staged before finish, and finish commits them
  separately.
- `myflow-command-surface`: the git-boundaries requirement follows the staging change and the
  two-commit split.
- `myflow-review-panel-economics`: the panel record gains a required findings table, and the
  handoff bar widens from Critical and Important to every severity.

## Impact

**New scripts:** `scripts/check-unfinished-work.sh` and `scripts/check-cleanup-complete.sh`, each
with a test harness beside it, following `check-finish-preflight.sh`'s verdict-line contract.

**Contract text:** `skills/myflow-contracts/pipeline.md`,
`skills/myflow-contracts/jira-integration.md`, `skills/myflow-start/SKILL.md`,
`skills/myflow-do/SKILL.md`, `skills/myflow-finish/SKILL.md`, and `README.md`'s command table.

**Behaviour operators will notice:** the staged diff at the human gate becomes implementation only;
a branch with unticked test-guide boxes now prompts before integrating; a manual-test guide written
before this change has no `## Known incomplete` section and will prompt once at its next finish.

**Not affected:** the three states and their transitions, the review panel's roster and model
policy, plan provenance, and the state file's schema — no field is added or removed.

**Scope.** This is slice A of KAN-17. Ask three, ask six, ask ten, ask eleven, ask fourteen, ask
fifteen and ask sixteen are deferred to a follow-up issue filed during this change's planning run;
they change what the operator sees and how the pipeline is configured, and share no requirement
with the asks above.
