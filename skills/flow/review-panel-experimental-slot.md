# Review panel — experimental slot

Loaded by `skills/flow/review-panel.md` only for a round whose roster carries an `exp-` slot. Every
section name below without a path is a section of `skills/flow/review-panel.md`.

## Experimental slot

When the decision's `panel.roster` carries an entry whose `slot` starts `exp-` — at most one, per
**Decide** (`skills/flow/brainstorm-planner.md`) — it is dispatched once, in pass 1 alongside the rest of the roster,
exactly like any other slot in **The roster** table (`skills/flow/review-panel.md`), carrying the same paragraphs every
slot's dispatch already carries there.

Its prompt is not `principles-reviewer-prompt.md` or any other fixed template: it is the file the
roster entry names in `prompt` — `skills/flow/experimental/<name>.md` inside the agents repo, never
a path inside the project worktree — read by **absolute** path and substituted into the dispatch
prompt the same way `[PRINCIPLES_PATH]` is resolved for Principles. Confirm the file exists before
dispatching; an absent file at dispatch time (the roster was decided against a prompt that has since
moved) is reported and this slot dropped from this run, never dispatched against nothing.

Its `-slot` on the dispatch record, and every `flow record finding -slot` this slot raises, is the
full `exp-<name>` id verbatim — never shortened to `experimental` or to `<name>` alone. Its report
file is `<abs-worktree>/.superpowers/sdd/panel-report-<round>-exp-<name>.md`, the same
`panel-report-<round>-<id>` shape every slot's REPORT FILE paragraph already names with `<id>`
substituted. The rendered panel record's Slot column therefore shows the `exp-` id unchanged, so the
prefix survives into the archive.

It is a diff-reading slot like Primary and Principles: **Panel
re-runs** governs it unchanged. **The docs-only reduction** still narrows a
docs-only branch to `primary` alone, and so does **The late-fix reduction**
(`skills/flow/review-panel.md`) on a qualifying fix run: the experimental slot is never part of
either reduced roster, and is dispatched again when a later round reclassifies the run off the
reduction — a docs-only guard reclassification, or a late-fix round whose Critical or Important
voids the reduction for the rest of the run. **The append scope**
(`skills/flow/review-panel-late-fix.md`) narrows the read scope alone, so the experimental slot
rides an append-scope pass 1 as it rides a full one.

It runs at most once per change, whether or not the roster is `compact` — the experimental roll and
the compact roll are independent per **Decide** (`skills/flow/brainstorm-planner.md`) — and never at all on a `default`
panel (a `micro` class), or when the decision recorded `experimental: none available`.

Per **Decide** (`skills/flow/brainstorm-planner.md`), it never joins the floor bundle: it joins
the second dispatch, last among the reading passes, only when one exists and has room; otherwise
the decision already records it `skipped — bundle cap` and it is recorded with
`flow record pass -round <round> -note 'experimental: skipped — bundle cap'` rather than displacing a persistent role.
