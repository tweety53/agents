# Review panel — kan-36-more-brainstorming-rounds-if-needed

Diff reviewed: `.superpowers/sdd/final-review.diff` (227 changed lines, 3 files)
Pass 1 — full roster selected for this change.

## Roster

| Slot | Selected? | Model | Note |
|---|---|---|---|
| 0 Primary | required | sonnet | |
| 1 Bugbot | **required, COULD NOT RUN** | sonnet (substitute) | `bugbot` subagent type unavailable in this harness. A general-purpose defect-only reviewer was dispatched in its place. **This is a substitute, not Bugbot**, and carries no Bugbot agent definition. |
| 2 Principles (Merged) | required | sonnet | |
| 3 Security | **not selected** | — | No trigger fired: no auth, tokens, crypto, secrets, path handling, deserialization, HTTP edge or dependency change; JQL assembly untouched, only which status names map to a position. |
| 4 Adversarial | selected | sonnet | 227 lines is in the 150-300 ask band with no other trigger; operator asked and chose include. |
| 5 Principles Lens B (simplicity & state) | selected | sonnet | Trigger: >~200 changed lines. |
| 5 Principles Lens C (robustness & ops) | selected | sonnet | Trigger: external-integration contract (`jira-integration.md`) changed. |

## Findings

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles (Merged) | Critical | `skills/myflow-contracts/pipeline.md:98` | Verbatim duplication of the exchange enumeration and the "one test, not a rule per gate" rationale between pipeline.md and `skills/myflow-start/SKILL.md:156`, despite SKILL.md:153 explicitly claiming the structure is stated once in pipeline.md. Violates this repo's cite-don't-copy hard invariant. `check-references.sh` cannot detect this class. Confirmed by parent. |
| F2 | Principles (Merged) | Minor | `skills/myflow-contracts/pipeline.md:76` | Level 1 stage table's `/myflow-start` row omits the two commits slice D adds, while the `/myflow-finish` row below enumerates its commits as a stage. Completeness gap created by this change's own scope. Confirmed by parent. |
| F3 | Principles Lens B | Important | `skills/myflow-start/SKILL.md:267` | The `answered by <decision-id>` cross-section reference is the first ID reference in this repo that crosses two independently-edited sections, and nothing enforces it: the decision ID must exist, IDs must stay unique across both sections, and every answered question's status must actually be flipped. Consistency is instruction-only and the text does not say so, while the count it feeds is what the operator reads at the gate. Fix is to name the limit, not to build a validator. |
| F4 | Principles Lens B | Minor | `skills/myflow-start/SKILL.md:186` | The threshold counts *rounds* while planning effort changes how many questions a round groups, so the offer's friction arrives after ~3 questions at `detailed` and far later (or never) at `low` — a hidden coupling between two axes the text asserts are independent. |
| F5 | Principles Lens B | Minor | `skills/myflow-start/SKILL.md:169` | Round numbering is under-specified for a round opened by the confirm's *Another round — I have something* answer: unstated whether an operator-initiated round consumes a slot in the 1-2-silent / 3-onward-offered sequence. Load-bearing, since it decides when the offer starts appearing. |
| F6 | Principles Lens C | Critical | `skills/myflow-contracts/pipeline.md:297` | The new `/myflow-start` commit asserts no branch precondition. `/myflow-start` runs in the main checkout and creates no branch, so with any other branch checked out both commits land there and `/myflow-do`'s worktree — created from the base branch — silently lacks the plan. That is the exact failure slice D exists to close, reopened, with no detection and nothing in the handoff to reveal it. `pipeline.md:950` already establishes the pattern for run 2 ("never assume it, and never derive it from the current branch"). Confirmed by parent, and reproduced live: this session sits on `openspec/kan-36-...`, not `main`. |
| F7 | Principles Lens C | Important | `skills/myflow-start/SKILL.md:139` | The two new commits have no skip-vs-fail contract. `pipeline.md:325` specifies the guarded `&&` chain, the skipped-empty rule, the stop-on-failure rule and the symlink case for every other commit surface in this file; `/myflow-start`'s are prose only. An empty second commit on a revision round, or a hook rejecting either, is unhandled — and a failed second commit reproduces F6's silent loss by another trigger. |
| F8 | Principles Lens C | Important | `skills/myflow-contracts/pipeline.md:487` | The `STARTED` template gained no field for the new commit action, though `pipeline.md:446` states that a field added to a command is added to the template in the same change, and `/myflow-do`'s `IN_PROGRESS` template carries a `Git:` line precisely because it writes to git. The operator cannot tell from the handoff whether the commits happened, on which branch, or whether one failed — and `branch` stays `null` by design, so the state file does not say either. This change added a field to that same template for slice B, so the rule was plainly in view. |
| F9 | Principles Lens C | Important | `skills/myflow-contracts/jira-integration.md:106` | `TO DO URGENT` is hardcoded into a contract shared by every project myflow is installed into, removing a previously-required consent for an irreversible external write. A different project with a status colliding on that literal name but different semantics (escalation rather than not-yet-started) gets auto-transition with no opt-out short of editing the shared file. **CONFLICTS WITH A RECORDED OPERATOR DECISION** (`todo-urgent-enumerated-synonym`), which considered and rejected a per-project synonym map as speculative generality. Routed to the operator rather than adjudicated by the parent. |
| F10 | Principles Lens C | Minor | `skills/myflow-contracts/jira-integration.md:106` | The new "position" abstraction inherits undefined near-miss matching — "usual spellings" is not defined for whitespace variance, where this file is meticulous elsewhere (the Unicode-category collapse rules for join titles). Pre-existing, but the diff rewrote this exact passage. |
| F11 | Primary | Minor | `skills/myflow-contracts/jira-integration.md:125` | The Unrecognised statuses paragraph cites the fuller `statusCategory` prohibition as applying "in full", then immediately restates its consequence in near-identical words. Diff-introduced duplication of the kind this change's own citation-over-copy constraint warns against. Distinct from the To Do name set, which is single-sourced correctly. |
| F12 | Defect hunt (Bugbot substitute) | Critical | `skills/myflow-start/SKILL.md:419` | Self-contradiction introduced by this diff. The soft-bound offer's decline option ends the stage while a question is open — the stated purpose of the `## Open questions` record — but `SKILL.md:158` ("or out of the stage"), the new guardrail at `SKILL.md:419` ("never leave brainstorming while holding a question … ends only on the operator's answer at the confirm") and `pipeline.md:99` all forbid it. An agent reading the guardrail literally must refuse the offer's own "No". Adjudicated by parent: the DESIGN is correct (decisions `soft-bound-from-third-round` and `open-questions-section-in-design` deliberately permit stopping with something open); the guardrail and pipeline wording are the error and must be corrected to say the stage ends on an explicit operator answer — at the confirm, or by declining the offer. |
| F13 | Defect hunt (Bugbot substitute) | Minor | `skills/myflow-start/SKILL.md:139` | The new `docs(<key>): design for <topic>` instruction does not define what `<key>` resolves to when no issue is linked, though an unlinked change is first-class and `jira-integration.md:72` defines the `<name>` fallback for exactly that case. The `docs(<key>)` convention itself was verified against real git history and is correct. |
| F14 | Adversarial | Critical | `skills/myflow-contracts/pipeline.md:297` | `/myflow-start`'s commits are local-only (D forbids push) and `main` is pushed only at finish run 2. Two proposals started back-to-back leave change B's planning commits on local `main`; `/myflow-do a` forks its worktree from local HEAD (`git worktree add -b` passes no start-point), so change A's PR against `origin/main` carries change B's entire proposal — breaking the standing invariant that planning artifacts never reach a review surface. Local `main` also diverges from `origin/main` indefinitely with nothing reconciling it. **VERIFIED LIVE by parent:** local `main` is currently 1 commit ahead of `origin/main` (`99b751e`, this session's design-doc commit). **CONFLICTS WITH RECORDED DECISIONS** `start-commits-all-planning-artifacts` and `legalise-rather-than-forbid`. Routed to the operator. Distinct from F6: F6 is the wrong-branch case, this is the right-branch case. |
| F16 | Adversarial | Important | `skills/myflow-start/SKILL.md:72` | A revision round that answers a recorded open question has no instructed path to update the entry: the only `answered by <decision-id>` instruction sits in section C, which a revision round has no reason to revisit. The count billed as the operator's evidence can over-report a resolved question indefinitely. |
| F17 | Adversarial | Important | `skills/myflow-start/SKILL.md:267` | No uniqueness constraint on IDs, though the ID is the sole match key for closing an entry and for the count. Two entries slugifying to the same ID make the lookup ambiguous, marking the wrong one answered and leaving the resolved one `open` forever. Pre-existing in Decisions; this diff gives it a counted, operator-facing consequence. |
| F18 | Adversarial | Important | `skills/myflow-start/SKILL.md:169` | The confirm and the offer are the only yes/no gates in these files with no stated fallback for a non-explicit answer. Every comparable gate states one ("No — leave the status alone *(default, recommended)*"). With no hard cap, a stalled or malformed prompt at these two gates has no documented resolution. |
| F19 | Adversarial | Important | `skills/myflow-start/SKILL.md:139` | Convergence explicitly counts the spec review (sections D/E) as an exchange that can reopen a question, but the design doc was already committed in section B. A reopened design has no instructed way to reach git without a forbidden third commit, leaving the committed design stale against what was approved. |
| F20 | Adversarial | Minor | `skills/myflow-start/SKILL.md:169` | The convergence test is self-administered with no requirement to disclose what was considered and set aside; the operator's authoritative answer rests on the agent's own summary of its own test. |
| F21 | Adversarial | Minor | `skills/myflow-start/SKILL.md:166` | An "I cannot answer" response has no defined handling before round 3, since the record-and-defer path is only reachable through the round-3 offer. |

**Deduped:** Adversarial's `TO DO URGENT` consent finding is the same finding as F9 (Lens C) and is merged into it; Adversarial corroborates it with a concrete sequence — an issue linked casually via the conversation-scan confirmation, auto-transitioned before brainstorming begins.

## Fix round 1 — outcome

One fix subagent, model=sonnet (`models.panelFix`). All 20 findings reported addressed; guards 0/0/0.
Files changed: `pipeline.md`, `myflow-start/SKILL.md`, `jira-integration.md`, `myflow-do/SKILL.md`,
`commands/myflow-start.md`, `commands-claude/myflow-start.md`; the `myflow-command-surface` delta
spec edited and left unstaged. The two command-tree files were not in the brief — the fixer found the
guardrail duplicated there and stale, which the both-command-trees rule requires be kept in step.

F6 is marked **fixed**, not withdrawn: the F14 change removed the wrong-branch case outright, so the
finding no longer applies to the code. Only an operator's answer may write a `withdrawn` marker, and
no withdrawal was asked for or given.

### New findings raised by the fix round itself

| ID | Source | Severity | Location | Note |
|---|---|---|---|---|
| F22 | Fixer, self-reported | Critical | `openspec/changes/kan-36-more-brainstorming-rounds-if-needed/specs/myflow-planning-gate/spec.md:9` | F12 is fixed in the skill but NOT in the delta spec, which still reads "or out of the stage" (line 9) and "SHALL end the stage only on an answer" (line 45). The normative contract still carries the contradiction the implementation just removed, and finish run 2 syncs this delta into the live spec. F12 is therefore only half-addressed. |
| F23 | Parent, verifying the fix | Important | `skills/myflow-contracts/pipeline.md:299` | The F19 fix widened the delta spec to permit a third commit (design document only, reopened design), but `pipeline.md:299` still says "commits twice … and nothing else" and `skills/myflow-start/SKILL.md:474` says "those two commits are this command's whole git surface". New contradiction introduced by the fix wave. |
| F24 | Fixer, self-reported | Minor | `openspec/changes/kan-36-more-brainstorming-rounds-if-needed/design.md:253` | The Risks section still describes pre-F14 behaviour — commits landing "wherever the operator is checked out — `main`". They now land on the change's own branch. |
| F25 | Fixer, self-reported | Minor | `openspec/changes/kan-36-more-brainstorming-rounds-if-needed/specs/myflow-handoff-output/spec.md:41` | The run-only-fields sentence does not account for the new `Git` line added to the `STARTED` template. |

## Fix round 2 — outcome

One fix subagent, model=sonnet. F12/F22, F23, F24, F25 all fixed; guards 0/0/0.
F24's fix used the file's own supersede convention rather than rewriting: decision
`legalise-rather-than-forbid` is now `superseded by start-commits-on-its-own-branch`, with the new
entry appended and the old text left intact.
F25 resolved by determining the `Git` line **is** run-only — it reports a per-commit outcome nothing
on disk records, exactly as the Jira line reports a transition nothing records.

The fixer was scoped to those four and correctly reported rather than fixed one thing it found
outside them:

| ID | Source | Severity | Location | Note |
|---|---|---|---|---|
| F26 | Fix round 2, self-reported | Important | `skills/myflow-start/SKILL.md:408` | Residual of F23: "Together they are its whole git surface" is still unqualified, while line 477 was corrected to admit the third design-only commit. Parent swept `skills/`, `commands/` and `commands-claude/` — this is the only remaining instance. |



## Pass 2 — full roster (escalated: delta spec altered, files touched outside the findings' set, public contract changed)

Diff: 381 changed lines across six files. Roster unchanged except Adversarial, which is now a hard
trigger (>~300 lines) rather than a borderline ask. Security still not selected — the two new files
are contract prose about git branches; no auth, crypto, secrets, query construction or dependencies.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F27 | Primary | Critical | `docs/manual-test/kan-36-more-brainstorming-rounds-if-needed.md:74` | The manual test guide still tells the operator to verify **pre-F14** behaviour: "check it creates no branch" (line 74) and "check the worktree contains the plan, since the base branch now carries it" (line 76). Both are now exactly backwards. Written by the parent before the operator's branch ruling and never refreshed. No guard scans this file, and it is read at the human gate. Confirmed by parent. |
| F28 | Primary | Important | `openspec/changes/kan-36-more-brainstorming-rounds-if-needed/proposal.md:76` | The proposal states `/myflow-do` is "not affected" when `skills/myflow-do/SKILL.md` is one of the six changed files; its affected-code list omits that file and both command-tree files; and its rationale still says the worktree is created from the base branch. This is the artifact published at `STARTED` and archived at finish. Confirmed by parent. |
| F29 | Primary | Important | `skills/myflow-start/SKILL.md:68` | Cites **Git boundaries** as stating the base-branch "resolution mechanism", but `pipeline.md:316` carries only rationale and points onward to **Finish contract**, whose snippet asserts `BASE != CUR` and exits 1 — which is `/myflow-start`'s ordinary case, starting a proposal from `main`. A citation pointing at content that is not there, wrapping a functional trap if reused literally. Confirmed by parent. |
| F30 | Primary | Minor | `openspec/changes/kan-36-more-brainstorming-rounds-if-needed/design.md:117` | Design still says "**Two commits, not one**" with no mention of the third design-only commit that the spec, the contract and the skill all now carry. |
| F31 | Primary | Minor | `skills/myflow-do/SKILL.md:79` | "already carries `/myflow-start`'s two commits" does not admit the possible third. Behaviourally harmless; same documentation-precision class as F30. |
| F32 | Principles Lens C | **Critical** | `skills/myflow-start/SKILL.md:68` + `skills/myflow-do/SKILL.md:75` | **The F14 fix breaks the ordinary run.** `/myflow-start` leaves `openspec/<name>` checked out in the main checkout; `/myflow-do` then builds its worktree from that same branch. Git refuses a branch already checked out elsewhere. **VERIFIED EMPIRICALLY by the parent** in a scratch repo: `git worktree add` exits 128 with `fatal: 'openspec/kan-36' is already used by worktree at '<main checkout>'`. Nothing in either skill, or in superpowers:using-git-worktrees, returns the main checkout to the base branch first. Every `/myflow-do` following every `/myflow-start` would fail. **Corroborated independently by Adversarial, which added a second dimension:** the harness's `EnterWorktree` creates a worktree only *on a new branch* (base ref `fresh` or `head`) — it has no mode that checks out an existing named branch. So "check it out; never create it" is unsatisfiable through the tool `superpowers:using-git-worktrees` tells agents to prefer *first*, as well as through the git fallback. Both routes dead-end on the first `/myflow-do` of every change. |
| F33 | Principles Lens C | Important | `skills/myflow-start/SKILL.md:67` | No precondition for a clean tree or an in-progress rebase/merge before branch creation. `/myflow-finish` has an explicit detached-HEAD guard for its comparable operation; `/myflow-start` has none. The case that succeeds silently carries unrelated uncommitted changes onto the new topic branch. |
| F34 | Principles Lens C | Important | `skills/myflow-start/SKILL.md:67` | The "never fail because it already exists" carve-out is scoped to revision rounds. A **creating** run that finds the branch already present — a prior run that crashed after branch creation but before the state write — has no stated behaviour; `git checkout -b` simply fails with no described recovery. |
| F35 | Principles Lens C | Important | `skills/myflow-start/SKILL.md:378` | Section F writes the `STARTED` state file before describing the second commit, inverting this document's own verify-then-declare pattern (run 2 blocks the `FINISHED` write on a cleanup leftover). A hook rejecting the change-directory commit after the state write leaves the handoff reading "Proposal ready" while the branch lacks the plan. |
| F36 | Principles Lens C | Important | `skills/myflow-contracts/state-self-heal.md` | `/myflow-start`'s commits are local-only for the whole `STARTED` → finish span, and self-heal's contradiction table has no row for `branch` — a recorded branch that no longer exists in git is not a checked contradiction. No recovery is stated for a deleted branch or a lost machine, and the operator is never warned these commits are unpushed. |
| F37 | Principles Lens C | Minor | `skills/myflow-start/SKILL.md:434` | The `Git:` handoff line names the branch in only one of its four alternatives, so a revision round that makes no new commits never surfaces the branch name at the gate. |
| F38 | Principles (Merged) + Lens B | Important | `skills/myflow-start/SKILL.md:472` | The two-or-three-commit fact is restated at **five or more** sites with only one citation: `commands/myflow-start.md:16` and `commands-claude/myflow-start.md:12` byte-identically and uncited; `SKILL.md:472` uncited; `SKILL.md:475` duplicating the content *before* its citation; and re-derived again at `pipeline.md:307`, `:316` and `:367`. The same file cites correctly elsewhere (`SKILL.md:364`, `:543`), so this is inconsistent application rather than a style choice. A future change to the exception now has five-plus sites to keep in lockstep. |
| F39 | Principles (Merged) | Minor | `skills/myflow-contracts/pipeline.md:107` | The paragraph disclaims restating the two prompts, then paraphrases their behaviour anyway. Same class as F1, reduced to paraphrase with no factual disagreement. |

**Deduped:** Lens C's citation finding is the same as F29 (Primary) and is merged into it. Lens C adds
that the consequence is `/myflow-start` having no stated stop-and-ask on an unresolvable base branch,
where `/myflow-finish` explicitly has one.

**Confirmed still holding from pass 1** (reported by the slots, not re-raised): F1's duplication fix
holds — `pipeline.md` states structure, the skill cites it; the two command-tree files are
byte-identical in the changed passage, the sanctioned duplication; hook rejection and empty-commit
handling are explicit for both commits; `/myflow-do` checking out an existing branch is stated as the
ordinary case; detached HEAD does not block `/myflow-start`'s create step, since the start point is
the resolved base rather than the current HEAD.
| F40 | Principles Lens B | Important | `skills/myflow-contracts/pipeline.md:523` | The `STARTED` handoff's `Git` line enumerates exactly four values, all built on a two-commit model (`pipeline.md:557` says so explicitly: "which of this run's **two** commits actually happened"), but the change itself creates a three-commit case. An agent hitting it must either invent unlisted text or report "both commits made", concealing a third commit on the branch being handed off for review. Duplicated at `skills/myflow-start/SKILL.md:434`, so both sites need the fifth value. |

**F29 corroborated independently by three slots** (Primary, Lens C, Lens B). Lens B adds the sharpest
remedy: `/myflow-start` should not consult `CUR` at all — `pipeline.md:317` says its resolution never
derives from what is checked out — so the citation must scope to the origin-lookup lines only, not to
finish's guarded script which bundles finish-specific validation.
| F41 | Adversarial | Important | `skills/myflow-start/SKILL.md:406` | The third commit's trigger is stated at two contradictory points. Section B ties it to the spec review in sections **D**/**E**; section F says a reopening "after this point" — after F's own commit — but F is terminal, so read literally it describes an event no stage can produce. The delta spec sides with B. An agent building the logic from F alone would look for a trigger that cannot occur. |
| F42 | Adversarial | Important | `skills/myflow-start/SKILL.md:186` | Two fix-round additions interact into a possible non-terminating loop. "**Recording a question never satisfies the test**" (unqualified) meets the "I cannot answer" path that records a question and continues, while the test asks whether the command "now holds a question its inputs do not answer" — which a recorded-but-unanswered question still is. Below the round-3 threshold the next test silently reopens a round on the very question the operator just said they could not answer, with no offer and no stated exit. The intended exception — that explicit operator deferral removes that question from "held" — is never drawn. |

**F40 corroborated by Adversarial**, which identified the cause precisely: round 1's fix added the
`Git` field (closing F8) and round 2's fix widened the git surface to permit a third commit (closing
F19/F23), and neither round revisited the other's work.

## Pass 2 — outcome

Sixteen findings: **2 Critical, 9 Important, 5 Minor.** Every slot reported; guards 0/0/0 throughout,
and every finding was invisible to them.

**Two slots reported the change is not ready for the human gate.** Both named F32 as sufficient on its
own: the fix for pass 1's worst finding introduced a defect that breaks the ordinary run of every
change, on the first command after planning.
| F43 | Defect hunt | Minor | `openspec/changes/kan-36-more-brainstorming-rounds-if-needed/specs/myflow-command-surface/spec.md` | The three-commit branch is a `SHALL` in the requirement text but has no `#### Scenario` among the file's eight, so it would ship into the permanent spec tree with nothing to catch a regression. |

The defect-hunt slot found **no Critical** and confirmed pass 1's sharpest finding (F12) is genuinely
fixed, with no residual absolute prohibition and no gap or overlap at the confirm/offer boundary. All
four of its findings share one root cause: the third-commit concession, added late in the fix rounds,
was not propagated to every location that had baked in "exactly two" as fact.

---

# OPERATOR DECISION AT THE PANEL HANDBACK — slice D is reverted

Presented with pass 2's outcome and three courses, the operator chose: **revert slice D entirely to
stage-only.** The original three-way contradiction is resolved the other way — by making
`myflow-command-surface`'s existing prohibition true, so `/myflow-start` stages and never commits.
The worktree-visibility problem becomes a documented known gap and gets its own change.

**This closes, as no longer applicable, every finding that exists only because of slice D:**
F14, F29, F30, F31, F32, F33, F34, F35, F36, F37, F38, F39(partly), F40, F41, F43.

**What survives, against slices A/B/C and the parent's own artifacts** — each is tabulated once at
its own pass above, and is named here in prose so no identifier carries two table rows:

- **F42** (Important) — the possible non-terminating loop: "Recording a question never satisfies the
  test" meets the "I cannot answer" path, and nothing states that an explicitly deferred question
  stops counting as held. Slice A; survives the revert.
- **F39** (Minor) — `pipeline.md`'s Brainstorm expansion disclaims restating the two prompts, then
  paraphrases their behaviour. Slice A; survives the revert.
- **F27** (Critical) — the manual test guide describes behaviour that will not exist after the
  revert. Rewritten wholesale rather than patched.
- **F28** (Important) — `proposal.md` misstates scope; rewritten for a three-slice change.

## Pass 3 — full roster on the reduced diff (273 lines, 3 files)

Escalated automatically: four fix rounds had run and the fix altered delta specs.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F44 | Primary | **Critical** | git index | The index for `openspec/changes/<name>/` was stale: `/myflow-start` staged those files, and the fix rounds plus the revert then modified them on disk without re-staging. `specs/myflow-command-surface/spec.md` stood at `AD` — staged as *added* in its pre-revert form, deleted on disk — so committing that index would have resurrected the delta the operator's decision removed. Fixed by the parent with `/myflow-do`'s own clearing pass, not by re-staging: only `git reset -- <paths>` retracts what an earlier command staged. |
| F45 | Lens B **+ Merged, independently** | Important | `skills/myflow-contracts/pipeline.md:101` | "The stage ends only on an explicit operator answer" stated in full four times across two files with no canonical-statement discipline — unlike this diff's `TO DO URGENT` set and `## Open questions` shape. The plan's own task A1 describes it as one fact to add. |
| F46 | Lens B | Minor | `skills/myflow-start/SKILL.md:175` | Scar tissue from the F42 fix: "at any round" established, then re-argued before the paragraph reaches its new content. |
| F47 | Lens C | Important | `skills/myflow-start/SKILL.md:192` | **The confirm's fallback was circular.** It fires only when the test came back empty, yet defaulted to "another round" — nothing to explore, and with no hard cap a session that cannot ask at all had no exit. `jira-integration.md` names that failure mode for its own ask and terminates; the convergence section never did. The offer's fallback is sound by contrast. |
| F48 | Lens C | Minor | `skills/myflow-contracts/jira-integration.md:103` | "Punctuation" left undefined in the loose-match bound. |
| F49 | Adversarial | Important | `skills/myflow-start/SKILL.md:199` | The offer promised the full backlog *and* this round's slice but collapsed both into one interpolation. When one answer opens two questions, declining left the operator under-informed or silently dropped the unseen question from the record and the count — against slice B's whole purpose. |
| F50 | Adversarial | Minor | `skills/myflow-start/SKILL.md:182` | "Considered and set aside" reads as rejected approaches, not deferred questions, so a round-1 deferral could go unmentioned at a round-3 confirm. |
| F51 | Adversarial | Minor | `skills/myflow-contracts/pipeline.md:413` | "never staged before finish" is literally false — `/myflow-start` stages two of those paths. Correctly scoped by context and pre-dating this diff. |
| F52 | Defect hunt | **Critical** | `openspec/changes/.../specs/myflow-handoff-output/spec.md:1` | A `## MODIFIED` block replaces the whole requirement body. This delta carried **3** scenarios against **17** live. The installed CLI's `findMissingCurrentScenarios` would hard-fail `openspec archive` at finish run 2; on the manual sync path the text is silently deleted from the permanent tree. |
| F53 | Defect hunt | **Critical** | `openspec/changes/.../specs/myflow-jira-projection/spec.md:62` | Same defect: **4** scenarios against **13** live on the join requirement, including the prompt-injection defences whose implementation is live and untouched. A renamed heading counted as a drop, since matching is by exact string. |
| F54 | Defect hunt | **Critical** | `openspec/changes/.../specs/myflow-planning-gate/spec.md:19` | **The F42 fix never reached the delta spec**, which still carried the unqualified rule producing the non-terminating loop, with no scenario for the deferral path. Identical class to F22 earlier in this change. |
| F55 | Defect hunt | Important | `openspec/changes/.../specs/myflow-handoff-output/spec.md:40` | Slice-D residue: the run-only enumeration still named "the Git line" on `STARTED`, which the revert removed and the live spec never had. |
| F56 | Defect hunt | Minor | `.superpowers/sdd/revert-report.md` | The revert report's scope note was inaccurate about what it left undone, after the parent fixed it. |
| F57 | Fix wave 5, self-reported | Important | `skills/myflow-contracts/pipeline.md:101` | F45 made this the canonical absolute statement; F47 then added a third terminating path documented only in `SKILL.md`, making it false in the degenerate case. **Third occurrence of this class** (F12, F22). |

**The defect-hunt slot found what five other slots missed**, by comparing each delta against its
**live** counterpart rather than against the implementation. Every other slot checked
delta-vs-code and found agreement — true, and insufficient. The parent's own pass-3 brief asked for
delta-vs-implementation, which is exactly the check that passes.

**Adversarial verified the revert mechanically:** `git diff 99b751e` on the three restored files is
empty and on the three changed files is *exactly* `final-review.diff` — writing slice D, rewriting it
three times and reverting it has a provably zero net effect.

## Fix waves 5 and 6 — outcome

Wave 5 fixed the seven pass-3 prose findings and **self-reported F57** rather than leaving it for a
reviewer. Wave 6 fixed the delta specs and **corrected the parent**: the "29 live scenarios" figure
the parent verified with an unbounded `awk` range was wrong — 29 combined two requirements; the true
count is 13, as the reviewer originally said. Wave 6 restored per requirement rather than per file.

**Verified by the parent after wave 6, by set difference rather than by count:** no live scenario
heading is absent from any `## MODIFIED` block, and `openspec validate --strict` reports the change
valid.

**Record damage, disclosed:** an earlier parent edit to this file removed the pass-3 section and the
marker-block heading while deduplicating four table rows. The section above was reconstructed from
the findings as reported by each slot; the marker block itself was never lost.

## Marker block — all findings, all three passes

`F15` was deduped into `F9` and never existed separately. `withdrawn` markers carry the operator's
own decision at the panel handback, which is the only thing that may write one.

findings-total: 56
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
finding-status: F16 fixed
finding-status: F17 fixed
finding-status: F18 fixed
finding-status: F19 fixed
finding-status: F20 fixed
finding-status: F21 fixed
finding-status: F22 fixed
finding-status: F23 fixed
finding-status: F24 fixed
finding-status: F25 fixed
finding-status: F26 fixed
finding-status: F27 fixed
finding-status: F28 fixed
finding-status: F29 withdrawn slice D reverted by operator decision at the panel handback
finding-status: F30 withdrawn slice D reverted by operator decision at the panel handback
finding-status: F31 withdrawn slice D reverted by operator decision at the panel handback
finding-status: F32 withdrawn slice D reverted by operator decision at the panel handback
finding-status: F33 withdrawn slice D reverted by operator decision at the panel handback
finding-status: F34 withdrawn slice D reverted by operator decision at the panel handback
finding-status: F35 withdrawn slice D reverted by operator decision at the panel handback
finding-status: F36 withdrawn slice D reverted by operator decision at the panel handback
finding-status: F37 withdrawn slice D reverted by operator decision at the panel handback
finding-status: F38 withdrawn slice D reverted by operator decision at the panel handback
finding-status: F39 fixed
finding-status: F40 withdrawn slice D reverted by operator decision at the panel handback
finding-status: F41 withdrawn slice D reverted by operator decision at the panel handback
finding-status: F42 fixed
finding-status: F43 withdrawn slice D reverted by operator decision at the panel handback
finding-status: F44 fixed
finding-status: F45 fixed
finding-status: F46 fixed
finding-status: F47 fixed
finding-status: F48 fixed
finding-status: F49 fixed
finding-status: F50 fixed
finding-status: F51 fixed
finding-status: F52 fixed
finding-status: F53 fixed
finding-status: F54 fixed
finding-status: F55 fixed
finding-status: F56 fixed
finding-status: F57 fixed
