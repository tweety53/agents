## Context

Defect raised by the operator on 2026-10-27 in `/flow kan-1037-stats-weekly-costs-dashboard`
(implementation run): (a) three visual-verifier spec questions put via AskUserQuestion, each with a
"(Recommended)" option, under `rules/design-mockups-are-specs.mdc`'s "Ask rather than invent";
(b) a mid-run turn end after an in-run fix task, citing token cost. About 380k uncached tokens
spent on the two stops.

Already canonical: **Auto-resolution** (`skills/flow-contracts/operator-prompts-auto-resolution.md`)
forbids asking a mid-run question with a recommended option; **Fewest operator actions**
(`skills/flow-contracts/pipeline.md`) bans run-size stops. Missing: one statement of the run's
whole shape, and the mockup rule's yield to auto-resolution — its core reads "ask" with no scope.

`rules/flow-manual-review.mdc`'s state diagram is the same block as pipeline.md's **States**: the
STARTED line's `/clear` is the plan-gate boundary and the other lines already say "lands straight
away unless review-before-landing is set or a question is open". It does not contradict the shape,
so it is not edited.

The `.mdc` core is rendered into `~/.claude/CLAUDE.md`'s managed block by `setup.sh global`; no
rendered copy is committed in this repository (`git grep -l "Ask rather than invent"` lists only
the `.mdc` and prose citing it outside it). The repo's post-landing reinstall
(`CLAUDE.md` **Reinstall after a change to this repository is fully done**) refreshes the installed copy.
<!-- measured: git grep -l "Ask rather than invent" @ 0e970879 → rules/design-mockups-are-specs.mdc plus docs/specs citing it -->

Blast radius: files citing **Fewest operator actions** or "Ask rather than invent" outside archives
are under 15; the fix edits two of them and cites the rest unchanged.

## Decisions

### Where the run's shape is stated

**ID:** shape-in-fewest-operator-actions
**Status:** active
**Chosen:** one opening paragraph of **Fewest operator actions** in `skills/flow-contracts/pipeline.md`, citing **What still stops** and `reviewBeforeLanding` rather than restating them.
**Considered:** a new section beside **States** — a second home for what **Fewest operator actions** already governs; editing `rules/flow-manual-review.mdc` — it carries only the trigger by design and loads the pipeline on demand.

### The plan gate's /clear stays

**ID:** plan-gate-clear-is-the-boundary
**Status:** active
**Chosen:** keep the `/clear` handoff at the plan gate and name it as the boundary between interactive planning and the non-stop run.
**Considered:** removing it — the creating run's context is spent on brainstorming, and the implementation run starts clean by design; the operator's wording ("interactive planning into implementation") targets stops after the plan, not the gate.

### How the mockup rule yields

**ID:** mockup-ask-yields-in-flow
**Status:** active
**Chosen:** one sentence appended to the `.mdc` core's "Ask rather than invent" paragraph, citing **Auto-resolution**; the body section is not edited since the core sits atop the same file.
**Considered:** editing the `## Ask rather than invent` body as well — repeats the core sentence in the same file.

### No guard

**ID:** no-guard-docs-only
**Status:** active
**Chosen:** no new guard: both edits are prose a reviewer reads. A lexical guard pinning a phrase would pass while the parent misreads the sentence again, as it misread the existing ban.
**Considered:** a `check-*` grep pinning the new sentences — guards wording, not behaviour; a normative-inventory entry — that guard reports a set and has no violation exit, so it pins nothing.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — two prose paragraphs in two files, wording fixed by the plan.
**Considered:** `flow-medium` — pipeline.md is loaded by every run, but the edit adds a paragraph citing existing rules and changes no transition.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one group, one fixer — the plan holds one task.
**Considered:** none — a single task has no split.

## Open questions

None.

## Live check

None — prose contracts only; no running service or persistent state is touched.
