# Review panel — kan-110-lighter-auto-code-review-by-default

## Roster

The operator instructed this run to use the roster this change itself proposes as its `light`
preset, ahead of the change landing: Primary, Principles, and Code review (low). That drops Bugbot,
which the contract in force still lists as a required slot, so the override was put to the operator
explicitly before the panel ran and approved explicitly. It is recorded here rather than left to be
inferred from the slot list.

Every slot ran on Sonnet, the model recorded under `models.reviewPanel`.

| Slot | Ran? | Model |
|---|---|---|
| Primary | yes | sonnet |
| Principles (Merged lens) | yes | sonnet |
| Code review (low) | yes | sonnet |
| Bugbot | no — excluded by the operator's roster override | — |

## Optional slots

Evaluated against `final-review.diff` before dispatch: 12 files, 385 insertions, 48 deletions.

| Slot | Trigger | Outcome |
|---|---|---|
| Security | a config file changed (`.myflow/project.md`, one added line) | not selected — the change touches no auth, crypto, secret, query, deserialization or dependency |
| Adversarial | fired — more than ~300 changed lines | fired, offered to the operator, declined |
| Lens B — simplicity & state | fired — more than ~200 changed lines | fired, offered to the operator, declined |
| Lens C — robustness & ops | fired — the guard's error handling and a config change | fired, offered to the operator, declined |

Three triggers fired and the operator declined all three after being shown which fired and why.
That is a decline, not an absence of triggers, and the distinction is recorded deliberately.

## Pass 1

Mode: full roster as selected above. Diff read: `.superpowers/sdd/final-review.diff`.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Primary | Important | `skills/myflow-contracts/state-file.md:44` | The closed-schema sentence still reads "omits a documented field other than `planningEffort` or `models`", so it calls a file omitting `reviewPanelRoster` unparseable — while line 82 of the same file says omitting it is valid. The file states its own absent-key rule twice, inconsistently, and it is the file the canonical split makes the owner of that rule. |

findings-total: 1
finding-status: F1 fixed

## Pass 2 — targeted re-review

Mode: targeted. Diff read: `.superpowers/sdd/fix-round-1.diff`.

Who re-ran, and why: Primary, always, as the integration check; and Principles, because F1 is a
canonical-split failure and that is the Principles lens. A targeted re-run is never fewer than two
agents, which is the other reason the second slot is here. Code review (low) did not re-run: it
raised nothing, and the fix touched a file outside anything it reported on.

One fix subagent handled the single open finding, dispatched on sonnet, the model recorded under
`models.panelFix`.

Both re-reviewers verdicted F1 ADDRESSED and found no new breakage in the fix diff. Every slot in
the roster now shows a non-stale clean result: Principles and Code review (low) were clean on pass
1 and Principles re-confirmed against the fix, and Primary's own finding is closed by its own
re-review.

## Deferred minors

Neither blocks handoff; both are recorded in the SDD ledger and were shown to the panel, which
judged neither worse than recorded.

- `scripts/check-vocabulary.sh`'s exemption matches the singular word `skill` only. A future mention
  writing the plural would be a false positive. No test case covers it.
- That exemption is line-scoped, so a line wrap separating the backticked skill name from the
  following word produces a false hit. It fired once during implementation and was fixed by
  reflowing the line.
