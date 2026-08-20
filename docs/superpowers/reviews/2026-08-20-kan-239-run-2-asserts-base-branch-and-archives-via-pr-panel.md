# Final review panel — kan-239-run-2-asserts-base-branch-and-archives-via-pr

**Roster:** `light` (recorded). Required slots: Primary · Principles · Code review (low).
**Panel model:** sonnet (recorded `models.reviewPanel`).
**Diff:** `git diff e61603b` — 1207 lines at the final round, under the cap
(`check-panel-diff-size.sh` exit 0 on every round, no operator prompt needed).

**Optional slots:** four triggers fired and **all four were declined by the operator**, in the single
multi-select prompt the `light` preset uses. Declined is recorded distinctly from a trigger that
never fired:

| Slot | Trigger that fired | Disposition |
|------|--------------------|-------------|
| Security | path and file handling — branch-name validation before git calls, and a caller-supplied trust anchor in `gather-self-review-context.sh` | **declined by the operator** |
| Adversarial | existing tests modified, and 1061 changed lines (over the ~300 threshold) | **declined by the operator** |
| Lens B — simplicity & state | 1061 changed lines (over the ~200 threshold) | **declined by the operator** |
| Lens C — robustness & ops | a new four-exit-code guard, refusal paths, a terminal `FINISHED` write | **declined by the operator** |

**Rounds:** 8. Rounds 2, 4, 5, 6 and 7 were **Full escalation** re-runs, each fired automatically by
the ladder rather than chosen — a fix had altered a delta spec or a guard's behaviour. Round 8 closed
with zero open findings at any severity and a non-stale clean result from all three slots.

## Findings

| ID | Round | Slot | Severity | Site | Finding | Resolution |
|----|-------|------|----------|------|---------|------------|
| F1 | 1 | Primary | **Critical** | `specs/myflow-self-review/spec.md` | a `## MODIFIED` block replaces the **whole** requirement, so the delta's 3 scenarios would have silently deleted the base spec's other 8 — KAN-102's symlink-hardening scenarios — plus a body paragraph on invalid-invocation handling and the `Silence runs it` scenario | all restored verbatim; every modified requirement in all delta specs then audited, bodies included |
| F2 | 1 | Primary + Code review | Minor | `skills/myflow-finish/SKILL.md` | run-2 preamble claimed two mark-sharing exceptions and "eleven steps, nine marks"; steps 5 and 6 already shared `finish.cleanup` | corrected to three exceptions and eight marks, verified by counting the marks |
| F3 | 1 | Code review | Minor | `skills/myflow-contracts/finish-contract.md` | said "the push is run 2's last step, step 10" while defining a step 11 | reworded — the push is step 10, step 11 closes the run |
| F4 | 2 | Principles | Minor | `specs/myflow-self-review/spec.md` | introduced by F1's fix: claimed an invalid `<repo-root>` exits 2 "exactly as a malformed `<archived-change-path>` is", which exits **0** | rewritten to state the exit-2 class and why the two deliberately differ |
| F5 | 2 | Code review | **Medium-high** | `scripts/gather-self-review-context.sh` | real bug: the archived path was resolved lexically against `$trusted_root` but semantically against the process cwd, so the new `<repo-root>` override plus a relative path diverged and reported **every source skipped at exit 0** — silent total loss | step 4 now resolves step 2's joined form; a second instance in the `tasks.md` lookup fixed too; 4 new assertions |
| F6 | 3 | Code review | Minor | `README.md` | prose still said run 2 would "commit and push the archive" | reworded to the archive branch and deferred push |
| F7 | 3 | Code review | Important | `openspec/specs/myflow-command-surface/spec.md` | a **capability spec** still mandated that run 2 "commit and push the archive… on its second run" — the behaviour this change removes — and no delta shipped to correct it | third delta spec added, carrying the full requirement: all 4 base scenarios plus one new |
| F8 | 4 | Primary | **Critical** | `specs/myflow-command-surface/spec.md` | introduced by F7's fix: asserted `/myflow-finish` never pushes the base branch "on either run", but run 1's merge-and-push route is "push; merge into the base branch; push that" | scoped to run 2, with an explicit paragraph stating run 1's route does push base and is untouched here |
| F9 | 4 | Primary | Minor | `README.md` | same overclaim, unscoped | scoped to run 2, naming run 1's route |
| F10 | 5 | Code review | Minor | `scripts/prepare-archive-branch.sh` | header documented exit 2 without naming three `git checkout` failure paths that exit 2 | enumerated, with the exit-1-versus-exit-2 boundary stated |
| F11 | 6 | Primary | Minor | `skills/myflow-contracts/finish-contract.md` | the exit-2 summary dropped "HEAD's own ref cannot be read", which the header names | added |
| F12 | 6 | Code review | Minor | `scripts/prepare-archive-branch.sh` | `validate_branch_name()` exits 1 on a malformed branch name — a fourth exit-1 cause named in neither document | enumerated in both |
| F13 | 6 | Dispatcher audit | Minor | `scripts/prepare-archive-branch.sh` | `origin/<base>` failing to resolve exits 3, a third exit-3 cause the header did not name | enumerated in both; all 16 exit sites then mapped to a documented cause |
| F14 | 7 | Principles | Minor | `skills/myflow-contracts/finish-contract.md` | six additive rounds had left the file **4 bytes** under its 40724 cap | 126 bytes of duplicated rationale cut, then the missing cause added; now 40594, 130 under |
| F15 | 7 | Code review | Minor | `skills/myflow-contracts/finish-contract.md` | the exit-2 restatement named three of the header's four causes, omitting "an argument is missing" | added |

finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed
finding-status: F7 fixed
finding-status: F8 fixed
finding-status: F9 fixed
finding-status: F10 fixed
finding-status: F11 fixed
finding-status: F12 fixed
finding-status: F13 fixed
finding-status: F14 fixed
finding-status: F15 fixed

findings-total: 15
reproducers-total: 15
finding-reproducer: F1 none — a spec-completeness defect whose loss occurs only at archive time, when the MODIFIED block replaces the base requirement; no command observes it beforehand
finding-reproducer: F2 sed -n 356,366p skills/myflow-finish/SKILL.md
finding-reproducer: F3 none — a prose self-contradiction between two paragraphs of one file; nothing executable observes it
finding-reproducer: F4 none — a self-contradiction within one requirement's prose, not an executable defect
finding-reproducer: F5 bash scripts/test-gather-self-review-context.sh
finding-reproducer: F6 sed -n 74,78p README.md
finding-reproducer: F7 none — a capability spec contradicting the shipped implementation; no guard resolves a spec requirement to code, which is why it survived to round 3
finding-reproducer: F8 none — a false normative sentence in a planning artifact; its falsity is established by reading the run-1 route table, not by running anything
finding-reproducer: F9 none — the same overclaim in prose
finding-reproducer: F10 grep -c exit scripts/prepare-archive-branch.sh
finding-reproducer: F11 none — an omission in one document relative to another; established by comparing the two
finding-reproducer: F12 bash scripts/prepare-archive-branch.sh . -badbase archive-branch
finding-reproducer: F13 grep -n origin scripts/prepare-archive-branch.sh
finding-reproducer: F14 scripts/check-contract-budget.sh
finding-reproducer: F15 none — an omission in one document relative to another; established by comparing the two

## What the panel is worth, on this change

**Two of the fifteen findings were defects in shipped behaviour** — F5, a silent total loss of the
self-review bundle, and F12/F13, exit paths the documented contract did not name. **Four were
introduced by earlier fixes in this same panel** (F4, F8, and the F2/F11 pair's successors), which is
the argument for re-running rather than accepting a fix on its author's word.

**F8 is the one to remember.** The code-review slot examined that exact delta spec in the same round
and passed it, having verified it was *complete* — no dropped scenarios — without checking whether
its new sentences were *true*. Only the Primary slot compared the claim against the run-1 route
table. A single-reviewer panel ships that.

**Three findings are the same class the change itself exists to fix**: a document that states
something false about the pipeline, with no guard able to notice. F7 in particular — a capability
spec mandating the behaviour being removed — would have been archived into the spec corpus.

**One defect no reviewer found.** `test-check-cleanup-complete.sh` failed in the worktree while
passing in the main checkout: `check-cleanup-complete.sh` requires every Temporary-artifacts-registry
row to be declared checked or not-checked, and this change added an "Archive branch" row without the
declaration. Five rounds of reading missed it; running the project's full `## test` list found it.
The declaration was added, with the reason.

## Advisory, not a finding

The Principles slot noted at round 8 that 130 bytes of headroom on `finish-contract.md` resolves
F14's literal complaint but not the trend behind it, and named candidates for a real reduction — the
two "when the script is absent" fallback blocks, or moving the worktree-cleanup shell block's inline
commentary into a script header. Recorded here rather than acted on: it is outside this change's
scope, and it is advice to whoever next touches that file.
