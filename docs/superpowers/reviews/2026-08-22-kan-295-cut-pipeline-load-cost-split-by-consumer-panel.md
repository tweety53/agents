# Review panel — kan-295-cut-pipeline-load-cost-split-by-consumer

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | Principles | critical | skills/myflow-contracts/git-boundaries.md:29 | A moved sentence was split at its em-dash and the tail capitalised into a new sentence, which is an edit to a moved passage rather than a citation repoint. |
| F2 | Principles | critical | skills/myflow-contracts/model-policy-rationale.md:42 | An appositive gerund phrase was stripped from its governing clause and capitalised as a standalone line with no subject or main verb. |
| F3 | Principles | critical | skills/myflow-contracts/artifacts-registry-rationale.md:22 | Fragments of two non-adjacent original sentences were stitched into one ungrammatical paragraph that never existed in the source. |
| F4 | Principles | major | skills/myflow-contracts/model-policy-rationale.md:45 | Two orphaned pronouns whose antecedents now live only in the sibling core file, so the appendix cannot be read on its own. |
| F5 | Primary | major | scripts/check-references.sh:578 | The new EXPECTED_ZERO_EMPTY_APPENDICES reason claims the appendix has no body when it has three lines of prose, and the file belongs in the existing EXPECTED_ZERO_RATIONALE_DOCS category. |
| F6 | Primary | major | openspec/changes/kan-295-cut-pipeline-load-cost-split-by-consumer/verification.md | The per-move ledger for tasks 2-7 was never persisted in any commit body, under .superpowers/sdd/, or in verification.md, so the design's designated evidence path is absent for five of seven tasks. |
| F7 | Primary | minor | openspec/changes/kan-295-cut-pipeline-load-cost-split-by-consumer/tasks.md | Nine touched files were never declared in their owning task's Files field and tasks.md was not corrected to match what the commits actually changed. |
| F8 | Primary | minor | openspec/changes/kan-295-cut-pipeline-load-cost-split-by-consumer/tasks.md | All 31 step checkboxes remain unticked despite every task being implemented and committed. |
| F9 | Code review (low) | minor | scripts/test-check-cleanup-complete.sh:1399 | A stale comment still cites Worktree cleanup in pipeline.md, a section this diff moved to finish-contract.md. |
| F10 | Code review (low) | minor | skills/myflow-contracts/session-records.md | The Loaded by line omits /myflow-fast even though myflow-fast/SKILL.md in this diff adds an explicit load directive for the file. |
| F11 | Primary | critical | skills/myflow-contracts/artifacts-registry.md | Five pointer sentences into the rationale were deleted outright rather than repointed, and survive nowhere in the live corpus. |
| F12 | Principles | critical | skills/myflow-contracts/model-policy-rationale.md:49 | The core pointer was dropped and the appendix paragraph's bold lead sentence discarded, leaving an orphaned opening with no antecedent. |
| F13 | Principles | critical | skills/myflow-contracts/model-policy-rationale.md:59 | Two separately-worded base sentences were merged into one new sentence verbatim in neither original, and bold emphasis was stripped from another. |
| F14 | Principles | major | skills/myflow-contracts/model-policy.md:88 | The trailing rationale pointer was dropped with no replacement in either file. |
| F15 | Principles | critical | skills/myflow-contracts/pipeline.md | A Stage marks passage was moved to the appendix but the core paragraph ends with no pointer and the move has no ledger row. |
| F16 | Principles | critical | skills/myflow-contracts/pipeline.md | A Guard resolution passage was moved to the appendix but the core paragraph ends with no pointer and the move has no ledger row. |
| F17 | Primary | major | openspec/changes/kan-295-cut-pipeline-load-cost-split-by-consumer/verification.md | A ledger row quotes a removed passage that was never removed and is still present in pipeline.md. |
| F18 | Primary | major | openspec/changes/kan-295-cut-pipeline-load-cost-split-by-consumer/verification.md | At least four genuine extractions have no ledger row, so the ledger fails the deletion-to-row direction the requirement states. |
| F19 | Code review (low) | major | scripts/check-contract-budget.sh:169 | Ten of the twelve family budget rows do not match the landing-size-plus-25-percent formula, five too generous and six too tight, so the ratchet does not hold what it claims. |
| F20 | Code review (low) | major | openspec/specs/myflow-contract-distribution/spec.md:19 | A live SHALL still states pipeline.md is canonical for git boundaries, which this change contradicts, and no delta spec corrects it. |
| F21 | Code review (low) | major | rules/myflow-manual-review.mdc:32 | The always-on rule loaded in every session still claims pipeline.md is canonical for git boundaries, which could lead an agent to skip loading git-boundaries.md. |
| F22 | Code review (low) | minor | commands/myflow-finish.md:21 | Eight command stub files carry the same stale canonical-for-git-boundaries claim. |
| F23 | Code review (low) | major | AGENTS.md:157 | The Model summary line still cites pipeline.md as canonical for Model policy, though the parallel sentence in skills/README.md and the command files was repointed. |
| F24 | Code review (low) | major | CLAUDE.md:109 | The same stale Model policy citation survives in CLAUDE.md while its sibling citation two lines below was repointed. |
| F25 | Code review (low) | major | README.md:226 | A second Model policy citation still points at pipeline.md though the one later in the same file was repointed. |
| F26 | Code review (low) | major | skills/myflow-contracts/finish-contract.md:196 | A Temporary artifacts registry citation still names pipeline.md while the identical citation five lines below was repointed. |
| F27 | Code review (low) | major | scripts/gather-self-review-context.sh:26 | Header comments still cite Git boundaries and Rendering the session records as living in pipeline.md after both moved. |
| F28 | Code review (low) | major | skills/myflow-contracts/git-boundaries.md:28 | A moved sentence says Handoff output names the planning paths below, but that section stayed in pipeline.md so there is nothing below in this file. |
| F29 | Primary | major | skills/myflow-contracts/SKILL.md:23 | The contracts skill index was never updated: it lists none of the five new contract files and still describes pipeline.md as holding git boundaries and the session records. |
| F30 | Primary | minor | skills/myflow-contracts/SKILL.md:47 | The appendix pairing table lists five core-to-appendix pairs and omits the five new ones. |
| F31 | Primary | minor | openspec/changes/kan-295-cut-pipeline-load-cost-split-by-consumer/verification.md:181 | The per-move ledger reports 18 rows but the table carries 19, and a second self-count later in the file repeats the error. |
| F32 | Primary | minor | scripts/check-contract-budget.sh:171 | The git-boundaries.md budget row is 5512, computed before round 3 grew the file to 4428, where the guard's own formula gives 5535. |

findings-total: 32
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
finding-status: F29 fixed
finding-status: F30 fixed
finding-status: F31 fixed
finding-status: F32 fixed

reproducers-total: 32
finding-reproducer: F1 grep -n pathspec skills/myflow-contracts/git-boundaries.md
finding-reproducer: F2 grep -n Dispatching skills/myflow-contracts/model-policy-rationale.md
finding-reproducer: F3 grep -n revision skills/myflow-contracts/artifacts-registry-rationale.md
finding-reproducer: F4 grep -n Which skills/myflow-contracts/model-policy-rationale.md
finding-reproducer: F5 sed -n 5,10p skills/myflow-contracts/worktree-resolution-rationale.md
finding-reproducer: F6 git log --format=%B -7 15f11cc..HEAD
finding-reproducer: F7 git show --stat ce87c60
finding-reproducer: F8 grep -n Step openspec/changes/kan-295-cut-pipeline-load-cost-split-by-consumer/tasks.md
finding-reproducer: F9 grep -n cleanup skills/myflow-contracts/pipeline.md
finding-reproducer: F10 grep -n Loaded skills/myflow-contracts/session-records.md
finding-reproducer: F11 grep -rn Temporary skills/myflow-contracts/artifacts-registry.md
finding-reproducer: F12 grep -n panel-fix skills/myflow-contracts/model-policy-rationale.md
finding-reproducer: F13 grep -n outlives skills/myflow-contracts/model-policy-rationale.md
finding-reproducer: F14 grep -n persisting skills/myflow-contracts/model-policy.md
finding-reproducer: F15 grep -n concurrent skills/myflow-contracts/pipeline-rationale.md
finding-reproducer: F16 grep -n prefix skills/myflow-contracts/pipeline-rationale.md
finding-reproducer: F17 grep -n restated skills/myflow-contracts/pipeline.md
finding-reproducer: F18 grep -n IntelliJ skills/myflow-contracts/pipeline-rationale.md
finding-reproducer: F19 wc -c skills/myflow-contracts/pipeline-rationale.md
finding-reproducer: F20 sed -n 17,22p openspec/specs/myflow-contract-distribution/spec.md
finding-reproducer: F21 sed -n 30,34p rules/myflow-manual-review.mdc
finding-reproducer: F22 grep -rn canonical commands/myflow-finish.md
finding-reproducer: F23 sed -n 156,157p AGENTS.md
finding-reproducer: F24 sed -n 108,109p CLAUDE.md
finding-reproducer: F25 sed -n 225,226p README.md
finding-reproducer: F26 sed -n 195,196p skills/myflow-contracts/finish-contract.md
finding-reproducer: F27 sed -n 24,27p scripts/gather-self-review-context.sh
finding-reproducer: F28 sed -n 28p skills/myflow-contracts/git-boundaries.md
finding-reproducer: F29 sed -n 23p skills/myflow-contracts/SKILL.md
finding-reproducer: F30 sed -n 47,51p skills/myflow-contracts/SKILL.md
finding-reproducer: F31 sed -n 181,209p openspec/changes/kan-295-cut-pipeline-load-cost-split-by-consumer/verification.md
finding-reproducer: F32 wc -c skills/myflow-contracts/git-boundaries.md
