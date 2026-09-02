# Review panel — kan-380-flow-self-review-model-overridable-per-project

Rendered from the store. Do not edit: the findings are rows, and the next render overwrites this file.

| ID | Slot | Severity | Location | Note |
|---|---|---|---|---|
| F1 | principles | Major | stats/internal/store/migrations/0017_flow_settings_planning_model.sql:8 | migration 0017's comment still asserts self_review_model empty means inherit default_model, contradicting 0016's reworded comment now saying empty means fable |
| F2 | bugbot | Major | skills/flow/archive.md:163 | the handshake paragraph describes the subagent's report opening with Model: <name> but never states this requirement is included in the dispatch prompt itself, unlike brainstorm.md's planner handshake which explicitly says the relay contract is stated in the same prompt |
| F3 | bugbot | Minor | .flow/project.md | no guard mechanically validates ## self review model's body against ValidModels; a surviving mutant, but mirrors the identical pre-existing gap already accepted for ## planning model |
| F4 | bugbot | Minor | skills/flow/SKILL.md | no guard catches a logic-reversal mutation in the SELF_REVIEW_MODEL bash resolution block; a surviving mutant, but mirrors the identical pre-existing unenforced-prose gap already accepted for the PLANNING_MODEL bash block |
| F5 | code-review-low | Minor | scripts/check-model-keys.sh:114 | a project with no .flow/project.md is printed as an OK line but excluded from the final CHECKED tally, so the printed project(s) checked count undercounts when multiple project roots are passed |
| F6 | principles | Minor | scripts/check-model-keys.sh | neither new guard (check-model-keys.sh, check-model-resolution-shell.sh) has a companion scripts/test-check-*.sh mutation-test harness, unlike this repo's established convention (run-guard-tests.sh auto-discovers test-check-*.sh by glob) that every guard needs one |

findings-total: 6
finding-status: F1 fixed
finding-status: F2 fixed
finding-status: F3 fixed
finding-status: F4 fixed
finding-status: F5 fixed
finding-status: F6 fixed

reproducers-total: 6
finding-reproducer: F1 grep -n "inherit default_model" stats/internal/store/migrations/*.sql
finding-reproducer: F2 grep -n "Model:" skills/flow/archive.md skills/flow-contracts/finish-contract-run2.md
finding-reproducer: F3 sed -i 's/^`fable`$/`bogus-model`/' .flow/project.md && bash scripts/check-references.sh; echo $?
finding-reproducer: F4 sed -i 's/\[ -z "$SELF_REVIEW_MODEL" \] \&\& SELF_REVIEW_MODEL=fable/[ -n "$SELF_REVIEW_MODEL" ] \&\& SELF_REVIEW_MODEL=fable/' skills/flow/SKILL.md && bash scripts/check-references.sh; echo $?
finding-reproducer: F5 none-see-note-multi-root-invocation-needed
finding-reproducer: F6 none-see-note-multi-command-reproducer
