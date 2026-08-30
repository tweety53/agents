# kan-370-self-review-model — design

## Context

`/flow`'s archive-phase self-review (the 5-angle retrospective, `skills/flow/archive.md` step 9 /
`skills/flow-contracts/finish-contract.md` Run 2 step 9) runs inline in the same session as the
rest of `/flow`, so it has no model of its own — it inherits whatever the session is on. KAN-370
asks for a way to pin it to a stronger model (e.g. Opus) independent of the session and the other
two model roles (`skills/flow/SKILL.md`'s `DEFAULT_MODEL`, `skills/flow-contracts/model-policy.md`'s
retired three-role table). `proposal.md` covers why in full; this file covers how.

## Approach

Add `SelfReviewModel` to the `flow_settings` store row, following the exact pattern
`stats/internal/store/settings.go` already uses for `DefaultModel`/`ValidModels`, except that an
empty string is a **valid** value meaning "inherit `defaultModel`" (unlike `DefaultModel`, which
must always be a non-empty `ValidModels` member).

`/flow`'s **Model resolution** section (`skills/flow/SKILL.md`) resolves a third value,
`SELF_REVIEW_MODEL`, from the same `flow settings get` call, falling back to `DEFAULT_MODEL` when
the field is empty or the store is unreachable — same fallback story `DEFAULT_MODEL` and
`REVIEWERS` already have.

`skills/flow/archive.md` step 9 (self-review) changes from an inline reasoning pass to a subagent
dispatch: `gather-self-review-context.sh`'s output is handed to a subagent running on
`SELF_REVIEW_MODEL`, which returns the five angles' findings (each angle's findings list, or an
explicit none-marker) as its report. The main session receives that back, then runs the *unchanged*
per-angle filing-ask prompts, report assembly, and commit — because a subagent cannot drive
`AskUserQuestion`, and the filing decision must stay with the operator regardless of which model
did the reasoning.

## Decisions

### Self-review model is a new independent field, not a rename of an existing one

**ID:** self-review-model-new-field
**Status:** active
**Chosen:** add `SelfReviewModel` alongside `DefaultModel`, empty = inherit — leaves `DefaultModel`'s
meaning (implementer + panel + panel-fix) untouched.
**Considered:** splitting `DefaultModel` back into three separate role fields (implementer/panel/
panel-fix) as `model-policy.md`'s retired table describes — rejected as out of scope; design.md's
`models-fields-collapse` decision collapsed those roles deliberately and this ticket doesn't revisit
that call, it only adds the one role (self-review) that never had a field at all.

### Self-review becomes a subagent dispatch instead of an inline pass

**ID:** self-review-subagent-dispatch
**Status:** active
**Chosen:** dispatch the 5-angle reasoning pass as a subagent on the resolved model, keep the
filing prompts in the main session.
**Considered:** running self-review inline but asking the operator to switch the whole `/flow`
session to the self-review model for that one step — rejected: it would also move the model for
any other work happening in that session, and Claude Code's frontmatter can't switch a session's
model mid-run anyway.

## Open questions

(none)
