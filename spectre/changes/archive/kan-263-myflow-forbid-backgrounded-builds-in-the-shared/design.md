## Context

The "shared dispatch preamble" the finding refers to is not a single file — it is a set of
`>`-quoted paragraphs repeated across `skills/flow/implement.md` and `skills/flow/review-panel.md`
wherever an implementer/reviewer subagent is dispatched. `CONTEXT BUNDLE` is de-duplicated by
citation. `REPRODUCE, DON'T READ` and `VERBATIM REPORT — THE FACT` are instead duplicated in full
at every required site — deliberately, and mechanically enforced: `scripts/check-dispatch-paragraphs.sh`
is a guard, with its own fixture-based test suite (`scripts/test-check-dispatch-paragraphs.sh`),
that fails the build if a required paragraph — or any of its load-bearing phrases — goes missing
from any of its required sites. Its own header names why: KAN-289's root cause was exactly a
required instruction living in no template, so a later prose edit could trim it away silently;
KAN-217 added the table-driven generalization when a second paragraph (`VERBATIM REPORT`) needed
the same protection.

`FOREGROUND BUILDS` is a behavioral dispatch instruction born from the same failure class this
guard exists to prevent — an instruction added ad hoc, per-prompt, after repeated real stalls
(KAN-253's finding), with nothing stopping it from being silently dropped again once folded into
skill prose. It belongs in the guarded set, not the cited-only set `CONTEXT BUNDLE` sits in.

## Decisions

### Duplicate the paragraph in full at every required site, and guard it

**ID:** foreground-builds-guarded-duplication
**Status:** active
**Chosen:** add the full `FOREGROUND BUILDS` blockquote, verbatim, at all four dispatch sites —
`skills/flow/implement.md` (implementer dispatch, per-task reviewer dispatch) and
`skills/flow/review-panel.md` (panel slot dispatch, panel-fix subagent dispatch) — and add a new
entry to `scripts/check-dispatch-paragraphs.sh`'s paragraph table (`SITE_ENTRY`/`SITE_PATHS`/
`SITE_MIN_BLOCKS`/`SITE_VARIANTS` plus `ENTRY_LABEL`/`ENTRY_SHARED_PHRASES`) requiring it at all
four, with new fixture cases in `scripts/test-check-dispatch-paragraphs.sh` mirroring KAN-217's
addition of `VERBATIM REPORT — THE FACT`.
**Considered:** the citation pattern originally approved (one canonical paragraph plus three
one-line citations, no guard changes) — superseded once `check-dispatch-paragraphs.sh` was found:
it is the established, enforced pattern for exactly this class of instruction, and using the
weaker, unguarded citation pattern here would leave the new instruction exposed to the same silent
drift KAN-289 already burned once.
**Superseded:** `foreground-builds-single-source`, below, which this decision replaces before
implementation began — recorded rather than deleted, per the decisions contract.

### Add the new instruction once, cited elsewhere

**ID:** foreground-builds-single-source
**Status:** superseded by foreground-builds-guarded-duplication
**Chosen:** define `FOREGROUND BUILDS` once, as canonical text, in `skills/flow/implement.md`'s
implementer dispatch block (next to `PLAN PROVENANCE` / `REPRODUCE, DON'T READ`) — every other
dispatch site cites it by name, one line, exactly as `review-panel.md` already cites `CONTEXT
BUNDLE`.
**Considered:** duplicating the full paragraph at each of the four sites, matching how
`REPRODUCE, DON'T READ` is currently duplicated — not yet known at the time this decision was made
to be a guarded, mechanically-enforced pattern rather than an unexplained repetition; discovering
`scripts/check-dispatch-paragraphs.sh` mid-task is what superseded this decision.

### Scope: which dispatch sites get the citation

**ID:** foreground-builds-site-scope
**Status:** active
**Chosen:** four sites — the implementer's initial dispatch (canonical text, `implement.md`), the
per-task reviewer dispatch (`implement.md`), the review panel's per-slot dispatch
(`review-panel.md`), and the panel-fix subagent dispatch (`review-panel.md`) — since panel-fix is
implementer-shaped (it runs `systematic-debugging` and produces a fix commit) and can run a build
the same way an implementer can.
**Considered:** the three sites literally named "implementer and reviewer" in the Jira description,
excluding panel-fix — rejected per operator's explicit choice to include panel-fix, since excluding
it would leave the one other build-capable dispatch role uncovered.
