## Why

A plan written by `/myflow-start` asserts things about the world, and nothing checks any of them.

The KAN-6 run is the evidence. Six plan claims were wrong: `getToolWindowIds` was a Kotlin property
not a function; `DiffNotificationProvider`'s SAM parameter and return are `@Nullable`; the marker key
was said to be readable from `DiffRequest` when it only ever reaches `DiffContext`; the baseline was
"194 tests" when it was 197; "expect 4 fewer tests" when it was 11; and a rename fixture real git
cannot detect as a rename at any `-M` threshold.

That cost roughly forty minutes of a two-hour run — but the time is the smaller half. Had the
implementer transcribed the `DiffRequest` snippet instead of investigating it, the guard would have
been permanently false and the whole context-menu feature a no-op **that every unit test still
passed**. It was caught because that one task happened to carry a one-off instruction to check.

The pattern across all six: in-repo code was fine; claims about things outside the repository were
not — platform SDK calls, and numbers estimated rather than run.

## What Changes

- **A new contract** defining four provenance tags: `verified:<how>` and `unverified:<what-to-check>`
  on fenced code blocks, `measured:<command> @ <ref>` and `predicted:<what-confirms-it>` on numeric
  claims. **A number carrying neither tag may not appear in a plan at all** — "194 tests" was not
  mislabelled, it was invented, so labelling only helps if an unsourced number is itself a violation.
- **A guard script** `scripts/check-plan-provenance.sh` enforcing that provenance is *stated*, plus
  the sandboxed harness `scripts/test-check-plan-provenance.sh`. Scoped to
  `openspec/changes/*/tasks.md` excluding `archive/`.
- **`/myflow-start`** tags every block and claim while writing `tasks.md`, and runs the guard before
  publishing the proposal artifact.
- **`/myflow-do`**'s implementer dispatch gains a standing clause: a block tagged `unverified:` is a
  hypothesis, not code to transcribe.
- **`.myflow/project.md`** gains the new guard in `## lint` and the new harness in `## test`.
- **Self-heal carries forward the fields it did not infer.** Added during implementation, at the
  operator's direction, after a live state file was damaged during this very run: `/myflow-start`
  wrote `kan-14`'s file with `artifactUrl` and `jiraIssue` set, and 25 seconds later a self-heal
  rewrite replaced both with `null`. `state-file.md` already forbids this; neither
  `state-self-heal.md` nor `skills/myflow-status/SKILL.md`'s rewrite instruction restates the duty
  at the point of the write. Both are corrected, and an unrecoverable field must now be *announced*
  rather than silently nulled.
- **Two stale facts corrected in `agents-repo-verification`**, forced by having to restate its guard
  requirement accurately: it currently requires `/myflow-review`, a command KAN-8 deleted, and names
  `scripts/test-state-advance.sh`, which does not exist on disk. Both are the very failure this
  change exists to prevent, found in this repository's own specs.

## Capabilities

### New Capabilities
- `myflow-plan-provenance`: how a plan states the provenance of its claims, and what an unverified
  claim obliges of the implementer who reads it.
- `myflow-state-integrity`: what a state-file write owes the fields it does not own, and what it must
  announce when it cannot recover one.
- `myflow-model-policy`: which model each subagent role runs on, why the implementer and reviewer
  defaults point in opposite directions, how an operator overrides either, and the ledger record
  that makes the policy auditable.

### Modified Capabilities
- `agents-repo-verification`: one requirement — *This repository declares its own myflow project
  configuration* — so the declared `## lint` and `## test` sets name the new guard and its harness.

### Deliberately NOT included — the kan-8 collision

An earlier draft of this change also removed the dead `myflow-state-advance` and
`myflow-review-panel-economics` capabilities, and modified a second `agents-repo-verification`
requirement plus one in `myflow-contract-distribution`. **All four collided with
`kan-8-myflow-updates`**, a change that is still open, complete at 83/83, and whose own delta specs
already remove and rewrite exactly those requirements — verified by comparing the requirement-heading
sets, which match.

Those deltas are dropped. The stale live specs are not evidence that the drift went unnoticed; they
are stale because kan-8 has not archived yet. Two changes proposing the same removal would leave
whichever archives second applying a delta whose requirements are already gone.

That widening was **also reverted**, after the adversarial reviewer built fixtures from kan-8's
delta bodies and ran the real guard against them: **19 hits**.
<!-- measured: ./scripts/check-vocabulary.sh $(find openspec/changes/kan-8-myflow-updates/specs -iname spec.md) -->
The cause is structural, not
kan-8's — **a requirement that forbids a term must name that term**, and one of kan-8's reads verbatim
`SHALL NOT contain \`gates\`, \`tested\`, \`originStage\`, \`REVIEWED_TREE\`, \`fastPath\``.
Scanning live specs for retired vocabulary flags requirements for doing their job, with no honest
repair available: rewording deletes the requirement, and a suppression marker is forbidden.

So nothing from that line of work ships. What remains is a short comment at `DEFAULT_TARGETS`
recording that it was tried and why it was rejected, so it is not re-proposed. Detecting drift in
live specs is a real problem and deserves its own change, with a design that tolerates prohibitive
statements.

### Added after implementation — implementer model policy

`/myflow-do` dispatches implementer subagents through **subagent-driven-development**, whose model
guidance is to pick "the least powerful model that can handle each role". The pipeline never
overrode that, and never recorded what was actually chosen — so the SDD ledger for this very change
contains zero model entries across its task history, while the review panel's model choices are
fully documented because the panel record happens to write them down.

That asymmetry surfaced during this change's own review: the panel's models are auditable, the
implementers' are not. The operator's decision, recorded here:

- **Implementers run on Opus** (or the harness's strongest available model). This deliberately
  overrides subagent-driven-development's economy selection for this pipeline — implementation
  defects are what the review panel then spends its budget finding, so the saving is false.
- **Review-panel slots stay on Sonnet**, unchanged, for the reason already stated: panel cost must
  not depend on which model the operator is running.
- **An explicit operator instruction overrides either**, in both directions.
- **The ledger records the model used for every dispatch**, because a policy nothing records is a
  policy nothing can verify — the same argument this change makes about numeric claims.

## Impact

- **New:** `skills/myflow-contracts/plan-provenance.md`, `scripts/check-plan-provenance.sh`,
  `scripts/check-plan-provenance.py`, `scripts/test-check-plan-provenance.sh`.
- **Toolchain widening, decided during implementation:** the guard's fence/container classifier was
  reshaped from Bash into `scripts/check-plan-provenance.py` (Python 3, standard library only) after
  five review panel passes and seven fix waves found the distinct defect classes the Bash version
  could not structurally avoid, enumerated canonically in `check-plan-provenance.py`'s own module
  docstring rather than restated here as a bare number — full record in design.md's "Post-review
  reshape" section. `scripts/check-plan-provenance.sh` becomes a frozen, byte-identical
  CLI wrapper (`exec python3 ...`) rather than the implementation itself. This repository is now
  Bash + Python, not Bash-only; `.myflow/project.md` records that as a deliberate widening, not a
  drift.
- **Modified:** `skills/myflow-start/SKILL.md` (step D), `skills/myflow-do/SKILL.md` (implementer
  dispatch), `rules/myflow-manual-review.mdc` (contracts table row), `.myflow/project.md`
  (`## lint`, `## test`), `skills/myflow-contracts/state-self-heal.md` and
  `skills/myflow-status/SKILL.md` (the carry-forward duty).
- **No installer change.** `setup.sh` copies the contracts directory wholesale; a new file inside it
  ships to `~/.claude/skills/`, `~/.cursor/skills/` and `~/.codex/skills/` with no edit.
- **Guards that must stay green:** `check-vocabulary.sh`, `check-references.sh`,
  `test-check-references.sh`, `test-setup.sh`. The new contract introduces cross-references, which
  is exactly what `check-references.sh` polices.
- **No behaviour change to any existing plan.** Archived plans are out of the guard's scope and are
  not rewritten.
