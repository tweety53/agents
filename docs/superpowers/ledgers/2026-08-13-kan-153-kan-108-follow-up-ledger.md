# SDD ledger — kan-153-kan-108-follow-up

**This ledger was compiled at finish, not maintained per dispatch, and that is a defect in the run
rather than a property of this format.** Model policy requires every subagent dispatch to record its
model in the ledger as it happens; this run recorded each model in the session's own reporting and
never wrote the file, so `scripts/preserve-session-records.sh` found no
`.superpowers/sdd/tasks/progress.md` to preserve. What follows is reconstructed from the dispatches
themselves. Every model named here was named explicitly on its dispatch — none was inherited — but a
record written after the fact is weaker evidence than one written during, and it is labelled so no
reader mistakes it for the latter.

## Configuration

Recorded in the change's state file, and applied on every dispatch below:

| Role | Value | Source |
|------|-------|--------|
| `models.implementation` | sonnet | `/myflow-fast`'s recorded defaults |
| `models.reviewPanel` | sonnet | same |
| `models.panelFix` | sonnet | same |
| `reviewPanelRoster` | light | same |

`/myflow-fast` records these defaults directly instead of asking, which is why `models.implementation`
is sonnet here rather than the Opus default the model policy sets for implementer subagents under
`/myflow-do`.

## Implementer dispatches

| Task | Dispatch | Model | Outcome |
|------|----------|-------|---------|
| 1, 2 | bundle 1 — the marker-helper library and the digit bound | sonnet | both committed; reported two plan errors, both confirmed against the code |
| 3 | the Markdown-integrity guard and its harness | sonnet | committed; calibration found 104 hits, seven rules tightened, one real defect left for Task 4 |
| 4 | declare the guard, repair what it finds | sonnet | committed; repaired the torn paragraph in `jira-integration-rationale.md` |
| 5 | the materiality qualifier and the size clause | sonnet | committed |
| 6 | the reproducer's process group | sonnet | committed; two plan deviations, each verified by experiment before being written |

`scripts/plan-dispatch-bundles.sh` produced five bundles — `1 2`, `3`, `4`, `5`, `6` — and dispatches
were serialised one implementer at a time against the single worktree, per the concurrency rule.

## Review dispatches

| Pass | Slot | Model | Outcome |
|------|------|-------|---------|
| per-task | Task 1 combined reviewer | sonnet | clean |
| per-task | Task 2 combined reviewer | sonnet | 1 Medium — the new case asserted the reason without the verdict |
| per-task | Task 3 combined reviewer | sonnet | 3 findings — one Critical exemption too broad, one forbidden suppression form, one latent false positive |
| per-task | Tasks 4 and 5 combined reviewer | sonnet | 3 findings — register, omission, trigger overlap |
| per-task | Task 6 combined reviewer | sonnet | 1 Medium — the survivor snapshot ordered correctly on only one path |
| panel 1 | 0 — Primary | sonnet | 5 findings |
| panel 1 | 2 — Principles | sonnet | clean |
| panel 1 | 3 — Code review (low) | sonnet | 3 findings |
| panel 2 | 0 — Primary | sonnet | clean |
| panel 2 | 3 — Code review (low) | sonnet | clean; ran as a general-purpose reviewer, the substitution its dispatch prescribes |

No slot was dispatched by `subagent_type`, so no ledger entry here reads `unknown (agent-defined)`.
The `light` roster's three required slots take an explicit model, and every one was given sonnet
rather than inheriting the parent's.

## Fix dispatches

| Round | Scope | Model | Outcome |
|-------|-------|-------|---------|
| Task 3 fix 1 | the three per-task findings | sonnet | fixed; the discriminator it added then over-fired on a legitimate shape |
| Task 3 fix 2 | the over-fire | sonnet | fixed; guard clean against the repository tree |
| Task 6 fix | the snapshot ordering | sonnet | fixed; required narrowing a group-wide kill, reported as wider than the finding |
| panel fix | four panel findings | sonnet | fixed |

Three plan-artifact findings were repaired by the parent rather than dispatched, because a subagent
may not write `openspec/`.

## Findings received outside the panel

A peer session sent two reuse findings unprompted. One matched a panel finding independently and was
closed by the same fix; the other was assessed and declined with reasons. Neither is a panel finding
and neither carries a marker line. The panel record holds the detail.
