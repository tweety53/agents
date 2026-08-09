# Final review panel — kan-106-slim-the-myflow-skills-cut-meta-prose-extract

Roster: light. Required slots: Primary, Principles, Code review (low). Optional slots: Adversarial,
Lens B, Lens C — all three triggered (>300 changed lines, new modules, new script/config
integration) but declined by the operator when offered as one bundled multi-select.

## Pass 1 (full roster, whole-branch diff `.superpowers/sdd/final-review.diff`, merge-base
`649fb1fa0e2cd1a2d1f2ee264c746410a1e81249`)

- **Primary** (models.reviewPanel: sonnet): clean, 0 findings.
- **Principles** (models.reviewPanel: sonnet, lens: Merged): 1 Important finding.
- **Code review (low)** (models.reviewPanel: sonnet): same finding as Principles, reported
  informally (not yet in F-id/marker form) — converges independently.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles + Code review (low) | Important | `scripts/prepare-workspace.sh` (cell-split awk block) | Duplicates `scripts/check-workspace-isolation.sh`'s cell-splitting/trim/fold awk logic near-verbatim (~70 lines) instead of sharing it — the same duplication class this whole change exists to remove elsewhere. |
| F2 | Code review (low) | High | `skills/myflow-finish/SKILL.md:340` (self-review skip prompt, Task 4) | Task 4 deleted "Silence or a session that cannot ask runs self-review, exactly as an explicit Yes would" as a pure restatement of `operator-prompts.md`'s silent-default rule — but `operator-prompts.md` only states what happens when the operator is *silent*, not the distinct "a session that cannot ask at all" case (the two are handled differently elsewhere in this corpus, e.g. `myflow-planning-gate`'s convergence rules). Real content loss, not a redundancy. |
| F3 | Code review (low) | Medium | `scripts/prepare-workspace.sh:306` (Task 10) | Port offset arithmetic uses bash `$((...))`, which parses a `Default` value with a leading zero as octal — a declared default like `0070` silently becomes a wrong port (verified: `$((0070+20))` = 76, not 90) or `0080` hard-crashes with "value too great for base". |
| F4 | Code review (low) | Low | `scripts/prepare-workspace.sh:78` (Task 10) | Guard-script presence check tests only `-f`, not the executable bit — a present-but-non-executable `check-workspace-isolation.sh` fails exec with an undocumented exit 126, outside the stated 0/1/2 contract. |
| F5 | Code review (low) | Medium | `scripts/prepare-workspace.sh:385` (Task 10) | `resolve_value_refs` resolves `url` rows in table declaration order; a `url` row referencing a later-declared `url` row silently substitutes empty string instead of erroring, since that row's `WSVAL` isn't populated yet in this pass. |

Pass 1 raised 5 findings total (F1-F5 above), all subsequently fixed — see Fix round 1 and the
final marker block at the end of this record, which is this document's only live one.

## Fix round 1

All 5 findings fixed:
- F1: `prepare-workspace.sh` no longer duplicates the awk cell-parser — it delegates row-parsing to
  `check-workspace-isolation.sh` via a new `CHECK_WORKSPACE_ISOLATION_PRINT_ROWS` env-var switch
  (default off, zero behavior change for existing callers), folded into Task 10's commit.
- F2: restored the "a session that cannot ask still runs self-review" rule in
  `skills/myflow-finish/SKILL.md`, distinct from the silent-default case `operator-prompts.md`
  already covers, folded into Task 4's commit.
- F3: `prepare-workspace.sh`'s port-offset arithmetic now forces base-10 (`10#${DEF[i]}`), fixing
  the leading-zero-as-octal bug, folded into Task 10's commit.
- F4: the guard-runnable check now tests `-x` as well as `-f`, folded into Task 10's commit.
- F5: `resolve_value_refs` now rejects a reference to a non-`database`/`bucket`/`port` row
  (defense in depth; the guard already forbids url→url at the table level) instead of silently
  substituting empty string, folded into Task 10's commit. New test coverage added for F3/F4/F5
  (cases 4-6 in `test-prepare-workspace.sh`).

Targeted re-run: Slot 0 (Primary, integration check) + Principles + Code review (low) — every slot
that ran in pass 1 — against the fix diff (`git diff 649fb1fa0e2cd1a2d1f2ee264c746410a1e81249`,
now including all 5 fixes).

## Re-check pass 1

- **Primary**: confirmed all 5 fixes, 0 new findings.
- **Code review (low)**: confirmed all 5 fixes, 0 new findings (one pre-existing, out-of-scope
  duplication noted for awareness only — `check-cleanup-complete.sh`'s own copy of the workspace-id
  formula, untouched by this change).
- **Principles**: confirmed the F1 DRY fix is solid. Raised a NEW Important finding: Task 11
  (compress the report-only reviewer prompts) wrongly compressed two persisted, repo-committed
  dispatch-prompt files under a misapplied reading of the caveman Boundaries rule.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F6 | Principles | Important | `skills/myflow-do/principles-reviewer-prompt.md`, `skills/myflow-do/adversarial-reviewer-prompt.md` (Task 11) | Task 11 caveman-compressed two persisted, repo-committed dispatch-prompt templates under a misapplied reading of the caveman Boundaries rule (that rule licenses compressing a subagent's chat report-back, not a persisted file the subagent reads as its own instructions). No other file in this 26-file change was compressed this way. |

## Fix round 2

F6: Task 11 reverted entirely — both prompt files restored byte-for-byte to their pre-Task-11
content, Task 11's commit dropped from history. See `tasks.md`'s Task 11 section for the record.

## Re-check pass 2

Targeted re-run: Slot 0 (Primary) + Principles (the raiser).

- **Primary**: clean, 0 findings.
- **Principles**: clean, 0 findings. Final principles-compliance assessment for the whole branch: **Yes**.

Every finding raised across both passes (F1-F6) was fixed and re-confirmed. Final state:

```
findings-total: 6
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
```

## Handoff

Zero open findings at any severity, from every slot in the roster, on the final pass. Panel clear.
