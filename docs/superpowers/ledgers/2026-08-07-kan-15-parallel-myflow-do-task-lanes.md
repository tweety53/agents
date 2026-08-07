# SDD ledger — kan-15-parallel-myflow-do-task-lanes

Recorded models come from the change's state file: `models.implementation` = Opus,
`models.reviewPanel` = Sonnet, `models.panelFix` = Opus.

Worktrees:

- agents: `/Users/tweety53/Projects/agents-worktrees/openspec-kan-15-parallel-myflow-do-task-lanes`
  (merge base `88793e0c63411ba403a652922a59ba832534d42b`)
- gymie: `/Users/tweety53/Projects/gymie-worktrees/openspec-kan-15-parallel-myflow-do-task-lanes`
  (merge base `fa8242a9377b4f0070fb7408035f683dd76d673a`)

## Dispatches

| Task | Role | Model | TASK_BASE | Outcome |
|------|------|-------|-----------|---------|
| 1.1 | implementer | Opus | `24b9082` (agents) | complete (uncommitted) |
| 1.1 | reviewer (spec + quality) | Sonnet | `24b9082` (agents) | 2 Minor, both fixed in-session; review clean |
| 3.1 | implementer | Opus | `fa8242a` (gymie) | complete (uncommitted) |
| 3.1 | reviewer (spec + quality) | Sonnet | `fa8242a` (gymie) | no findings against the diff; 3 Major plan gaps raised |
| 2.1 | implementer | Opus | `f27b462` (agents) | complete (uncommitted) |
| 2.1 | reviewer (spec + quality) | Sonnet | `f27b462` (agents) | 5 Major, 2 Minor |
| 2.1 | fix round 1 | Opus | `413322f` (agents) | all 7 closed; added the `url` kind and the `survivors` verb |
| 2.1 | re-review (targeted) | Sonnet | `413322f` (agents) | 7 closed; 1 Critical + 3 Major raised against the new shape |
| 2.1 | fix round 2 | Opus | `8f2ef08` (agents) | 4 closed; rationale moved into `workspace-isolation.md` |
| 2.1 | verification re-review | Sonnet | `8f2ef08` (agents) | clean — no Critical or Major; 1 Minor against the fixer's own report |
| 2.1 | fresh-reader review | Sonnet | (both files whole) | 1 Critical + 1 Major + 5 Minor — the Critical two reviews had missed |
| 2.1 | fix round 3 | Opus | `1574d28` (agents) | all 7 closed; 3 residuals reported, fixed by the parent |
| 2.1 | final verification | Sonnet | `1574d28` (agents) | 1 Major (a fourth citation-target mismatch), fixed by the parent |
| 2.5 | implementer | Opus | `e560666` (agents) | complete; found every contract index already stale |
| 2.5 | reviewer | Sonnet | `e560666` (agents) | 1 Minor, fixed; two rulings given |
| 2.3 | implementer | Opus | `3e654e5` (agents) | complete (uncommitted) |
| 2.3 | reviewer | Sonnet | `3e654e5` (agents) | 1 Major, fixed by the parent; 2 Minor |
| 2.2 | implementer | Opus | `8064e83` (agents) | complete; declined a bullet belonging to task 3.2 |
| 2.2 | reviewer | Sonnet | `8064e83` (agents) | 2 Minor; one closed by the parent |
| 2.4 | implementer | Opus | `273b43b` (agents) | complete; harness 54 → 108 assertions |
| 2.4 | reviewer | Sonnet | `273b43b` (agents) | 2 Major, 1 Minor — one of them plan-level |
| 2.4 | fix round 1 | Opus | `7028c09` (agents) | the unbounded `survivors` execution |
| 2.6 | implementer | Opus | `7028c09` (agents) | complete (uncommitted) |
| 2.6 | reviewer | Sonnet | `7028c09` (agents) | 2 Major, both on the ordering decision |
| 2.6 | fix round 1 | Opus | `1f4749c` (agents) | ordering inverted; invocation directory pinned |
| 3.2 | implementer | Opus | `a55ec5a` (gymie) | complete; killed the operator's backend with an unscoped `pkill`, disclosed |
| 3.2 | reviewer | Sonnet | `a55ec5a` (gymie) | 2 Critical (both routed to 3.3), 2 Major |
| 3.2 | fix round 1 | Opus | `fc183f7` (gymie + agents) | locale-pinned the derivation; buildSrc brought into the lint gate |
| 3.2 | drift fix | Opus | `1500f6a` (agents) | the guard's copy of the derivation, which the contract change had left behind |
| 3.3 | implementer | Opus | `c16fbed` (gymie) | complete; `remove`/`survivors`, OAuth redirect, `.env`, `pgDb` defaults |
| 3.3 | reviewer | Sonnet | `c16fbed` (gymie) | 1 Major — unvalidated id reaching destructive SQL — and 2 Minor |
| 3.3 | fix round 1 | Opus | `cdce9ce` (gymie) | whole-id validation **and** SQL escaping; found a second injection site |
| 3.4 | implementer | Opus | `778e87e` (gymie) | complete; the declaration, exercised end to end against the guard |
| 3.4 | reviewer | Sonnet | `778e87e` (gymie) | 1 Critical in the guard (`pipefail`), 2 Minor |
| 3.4 | fix round 1 | Opus | `b5df347` (agents) | `pipefail`; duplicate-heading skip |
| panel | slot 0 primary | Sonnet | both `final-review.diff` | 1 Major (4.1's live proof), 1 Minor (this ledger) |
| panel | slot 1 defect hunt | Sonnet | both `final-review.diff` | 1 Critical, 1 Major |
| panel | slot 2 principles (Merged) | Sonnet | both `final-review.diff` | clean — 2 Minor |
| panel | slot 3 security | Sonnet | both `final-review.diff` | clean — 2 Minor latent |
| panel | slot 4 adversarial | Sonnet | both `final-review.diff` | in flight |
| panel | slot 5 principles (Lens B) | Sonnet | both `final-review.diff` | 2 Major, 1 Important, 2 Minor |
| panel | slot 6 principles (Lens C) | Sonnet | both `final-review.diff` | 1 Critical, 2 Major, 3 Minor |
| 6.1 | implementer | Opus | `a1b3b76` (gymie) | complete (uncommitted) — `selectedServices`, 12 tests |
| 6.2 | implementer | Opus | `a1b3b76` (gymie) | complete (uncommitted) — `portBlockDecision`, 9 tests |
| 6.1 + 6.2 | reviewer (spec + quality) | Sonnet | `a1b3b76` (gymie) | dispatched over both, since the two ran in one worktree in parallel |
| 6.3 | implementer | Opus | `a1b3b76` (gymie) | the wiring, the four task descriptions and the README |
| panel | pass 4, full roster | Sonnet | gymie `final-review.diff` | 6 findings; agents not re-reviewed, its diff unchanged |
| fix 5 | fix round | Opus | `17b0547` (gymie) | 6 closed; its repair produced 8 |
| panel | pass 5, full roster | Sonnet | gymie `final-review.diff` | 8 findings; 3 slots clean |
| fix 6 | fix round | Opus | `57753b4` (gymie) | 8 closed; its repair produced 13 |
| panel | pass 6, four slots | Sonnet | gymie `final-review.diff` | **13, three Critical** — the pass that ended the inference approach |
| 7.1–7.3 | implementer | Opus | `b2d1ea2` (gymie) | the redesign: `PortBlock.kt` and its 29-test suite deleted, `ResolvedBlock.kt` in |
| panel | pass 7, four slots | Sonnet | gymie `final-review.diff` | 10 findings, 2 Critical — **Lens B endorsed the design** |
| fix 8 | fix round | Opus | `f3076d5` (gymie) | 10 closed; the F64 seam closed on both sides |
| panel | pass 8, two slots | Sonnet | gymie `fix-round-8.diff` | both pass-7 Criticals confirmed closed; 1 new Critical |
| fix 9 | fix round | Opus | `c7a885d` (gymie) | 8 closed; F74's rewritten test proven falsifiable by mutation |
| panel | pass 9, two slots | Sonnet | gymie `fix-round-9.diff` | **no Critical**; parser checked against real porcelain output |
| fix 10 | fix round | Opus | `55869cc` (gymie) | 4 closed; the record invariant stated and the header made to match |
| panel | pass 10, one slot | Sonnet | gymie `fix-round-10.diff` | closing verification, scoped to that round |

**Group 6 and 7 cost six fix rounds and seven panel passes, and the count is the finding.** Rounds 5
and 6 each closed every finding put to them and each produced more than it closed — 6 → 8 → 13 — and
every one of those was against machinery inferring which port a service was on, never against
`-Pservices`, the feature actually asked for. The operator was shown the trend and chose to delete
the inference rather than repair it again. **The two rounds after the reversal went 10 → 1 → 0
Criticals**, which is what convergence looks like and what its absence had looked like for three
rounds before.

**Group 6 was dispatched in parallel, and the contention it hit is the reason to record it.** 6.1 and
6.2 touch different files and were dispatched at once into the **same** worktree. Both landed, but
6.1's first green run failed on 6.2's half-written file — a test class present before its
implementation. Neither agent was wrong and neither retry was needed beyond waiting. That is the
in-worktree form of exactly what this change's Jira issue proposes solving with per-lane build
isolation, observed here rather than argued about: shared `buildSrc` compilation is the coupling,
not the Gradle daemon lock the issue predicted.

**Two panel slots were adapted, and the adaptation is recorded rather than left to look like a
choice.** This harness exposes no `bugbot` or `security-review` agent type, so slots 1 and 3 ran as
general-purpose agents against the skill's own `bug-hunter-reviewer-prompt.md` and
`security-reviewer-prompt.md`, on the panel's model. Their ledger entries therefore name Sonnet
rather than `unknown (agent-defined)`: the model *is* known here, because the dispatch chose it.

## What the reviews cost and bought

Recorded because the shape of it is the argument for the panel that follows, not a complaint about it.

**The same defect class appeared four times: a citation that resolves while the cited section does
not contain the claim.** `scripts/check-references.sh` matches a bold token against a *heading*, so
every one of them passed the guard. Two were introduced by fix rounds that were themselves closing
citation findings. What caught the worst of them — `<id_underscored>` cited as canonically defined
in a file that never defined it — was not a checklist review but a **cold read**: an agent given
only the two files and told to build against them, which failed at the first job and said so.

**Two defects were invisible to every reviewer that read the diff.** Task 2.4's review found that
**nothing in the pipeline called the project's `remove` command** — the registry row, the contract
and the guard all described a step no file performed. Task 2.6's review then found the fix's
ordering argument rested on an unstated assumption about which directory the command runs from, and
that the order chosen would have made a leftover the routine outcome rather than the rare one. Both
came from reading the *pipeline* rather than the change.
| 3.1b | implementer | Opus | `63b2920` (gymie) | complete (uncommitted) |
| 3.1b | reviewer (spec + quality) | Sonnet | `63b2920` (gymie) | 3 Minor, all resolved; review clean |

## Plan repairs made during the run

Two, both because a per-task review found the plan could not satisfy the delta spec as written.
Recorded here because the plan's own history does not otherwise show why a task appeared mid-run.

- **Task 3.1b was added** after task 3.1's review confirmed three Major gaps: `application-local.yml`
  also shadows the object-store bucket, endpoint and public base URL and both frontend URLs, and the
  gateway carried no cache-index key at all. Each blocked a requirement the delta spec already makes.
  A checkbox was added to task 3.2 requiring every derived value to be exported, since a value
  derived and never exported is the same failure one level out.
- **Task 3.1b gained a checkbox** after its own review found the README documenting the remaining
  Redis host, port and password as environment variables that the local profile silently ignored.
  Not an isolation gap — a false claim, fixed by making the three overridable.
- **The delta spec was amended in three places** after task 2.1's re-review, because the contract
  the change was producing had outgrown the requirement it was written against: containment now
  splits by what the pipeline *does* with an entry rather than naming one rule for all of them; a
  survivor is established by asking rather than inferred from the removal's exit code; and an
  application whose port is fixed outside the declaring project's repository is recorded as a
  stated limitation rather than claimed as isolated.
- **`design.md` gained the decision `kmp-frontend-not-port-isolated`**, taken by the operator when
  the re-review found the KMP frontend's port fixed in a repository this change may not touch. The
  port block is the backend, the gateway and the admin panel; the KMP frontend keeps 3000. Tasks
  3.2, 3.3 and 3.4, `proposal.md` and the plan header were brought into line with it.

## Notes

- The worktree's `openspec/changes/kan-15-parallel-myflow-do-task-lanes/` held a stale
  container-era revision of the plan. It was re-synced from the main checkout before the
  per-task review of 1.1, so every later dispatch reads the current plan from the worktree.
- `scripts/check-plan-provenance.sh` scopes to `openspec/changes/*/{tasks,design,proposal}.md`
  only, so `verified:` tags on contract files under `skills/` are preserved by convention
  rather than enforced by a guard.
- The design decision the plan calls `ports-deterministic-with-block-fallback` is recorded in
  `design.md` as `ports-deterministic-from-digest`. Same content, different id.
