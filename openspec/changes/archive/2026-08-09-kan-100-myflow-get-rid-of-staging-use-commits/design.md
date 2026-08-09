## Context

`/myflow-do` currently implements every task uncommitted (`NO-COMMITS` dispatch clause), isolating
each task's and each fix round's diff with `skills/myflow-do/scripts/checkpoint`
(`git stash create`, falling back to `HEAD`) and reviewing it via
`skills/myflow-do/scripts/uncommitted-review-package` (`git diff BASE` against the live working
tree). `/myflow-finish` run 1 makes the branch's only two commits: implementation, then planning
artifacts. `Build: red` tasks name their green merge partner inline in the tag's own text.

KAN-71 evidence (cited in the Jira ticket) showed this self-report model failing silently:

- "a task declaring 3 files touched 11"
- "a task claiming 4 new tests wrote zero"
- three defects shipped behind green suites
- "a reported test count off by 600"

Nothing mechanically checked any of it against real state, because there was no real state — only
a workspace diff and an agent's own account of it.

## Goals / Non-Goals

**Goals:**
- Give each task a real commit, so review and new mechanical guards check something real instead of
  a snapshot.
- Add per-task fields (`Files:`, `Tests:`, `Regression:`, `Baseline:`, `Squash-with:`, `Commit:`)
  that a runtime guard checks against the actual commit, immediately after it's made.
- Keep the final branch shape `/myflow-finish` produces unchanged — two commits, same as today.

**Non-Goals:**
- Bisectable per-task history on the merged branch. Per-task commits exist for review and
  mechanical checking during `/myflow-do`; they do not survive `/myflow-finish`'s reshape.
- Any change to the review panel's slot roster, model policy, or escalation ladder — only the diff
  source it reads changes.
- Any change to `/myflow-finish` run 2 (archive/cleanup).

## Decisions

### Squash everything at finish, not bisectable-per-task history

**Chosen:** `/myflow-finish` run 1 collapses every task/fixup commit with
`git reset --soft <recorded-merge-base>` before its existing two-commit sequence. Final branch
history is unchanged from today's shape — one implementation commit, one planning-artifacts commit.

**Considered:** keeping one commit per task on the merged branch for bisectability — the reason the
ticket itself gives for `Squash-with:` ("keeping history bisectable"). Rejected: confirmed twice
with the operator, who was shown the tension with the ticket's own wording and chose the
single-commit final shape regardless. Per-task granularity is valuable during `/myflow-do` for
review and mechanical checking; it is not wanted in the merged history.

### Fix rounds commit as `git commit --fixup`, autosquashed immediately

**Chosen:** every fix-round commit targets its task via `--fixup=<task-sha>` and is folded in with
`git rebase --autosquash` immediately, before re-review reads the diff.

**Considered:** leaving fix commits unfolded, as separate commits. Rejected — the ticket names
`git commit --fixup` explicitly as `Task-Id:`'s reason for existing, and unfolded fixups would leave
the `Files:`/`Tests:` guards checking a moving target spread across multiple commits per task.

### A runtime guard, not a plan-time-only guard

**Chosen:** the `Files:`/`Tests:`/`Regression:`/`Baseline:` guard runs during `/myflow-do`, against
the real commit, immediately after each task is committed and before review is dispatched.

**Considered:** a plan-time-only check, matching `plan-provenance`/`build-green`'s existing pattern
of only asserting the plan text is well-formed. Rejected — the entire point of these fields (per the
KAN-71 evidence the ticket cites) is to catch a mismatch between what was claimed and what was
actually committed, which a plan-time check structurally cannot see.

### `Squash-with:` replaces `Build: red`'s inline partner text

**Chosen:** `Build: red` becomes a bare tag; `**Squash-with:** Task <N>` names the partner.
`myflow-build-green`'s guard resolves the partner from the new field.

**Considered:** keeping the partner in `Build: red — merges with Task <N>` and adding the new
fields alongside it. Rejected — the ticket explicitly states `Squash-with:` *replaces* that prose,
and the field is also the literal target the fixup-and-autosquash mechanism uses for red tasks, so
one field serving both readers (the guard and the fixup step) is simpler than two.

### `Regression:` and `Baseline:` degrade to skipped, not failed, when unsupported

**Chosen:** when the project's `## test` command can't target a named test, or doesn't report a
parseable count, these two checks report skipped-not-verified rather than failing the run.

**Considered:** failing the run whenever a project's tooling doesn't support named-test execution or
parseable counts. Rejected — that would make the guard a hard requirement on every project's test
runner shape, when the pipeline already has a precedent for this exact trade-off (workspace-isolation
survivor reports skip rather than fail when a service can't be reached).

### One change, not split into commit-model and mechanical-fields

**Chosen:** both halves ship together.

**Considered:** shipping commit-per-task alone first, with the mechanical fields as a follow-up once
real commits exist to check against. The operator chose to do both together, accepting the larger
surface.

## Risks / Trade-offs

- **[Risk]** `Baseline:` runs the project's full test command twice per task (parent and task
  commit) → **Mitigation:** none needed beyond the existing skip-not-verified fallback; this is an
  accepted cost of the mechanical check the ticket asks for.
- **[Risk]** `git rebase --autosquash` on a task commit that already has downstream fixups could
  conflict → **Mitigation:** tasks are dispatched serially per `subagent-driven-development`, so a
  task's fixups are always folded before the next task starts; no interleaving is possible.
- **[Risk]** Retiring `checkpoint`/`uncommitted-review-package` is a breaking change for any in-flight
  `/myflow-do` run created before this change lands → **Mitigation:** none — this pipeline has no
  cross-version compatibility guarantee for in-flight runs; an in-flight change should finish under
  the old skill version or be re-planned.

## Migration Plan

No data migration. In-flight changes at `IN_PROGRESS` created before this lands should either finish
under the prior skill version or be re-planned. No rollback beyond reverting the merged change.

## Open Questions

*(none)*
