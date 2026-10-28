## Context

- Approved prototype: the real page, built from this branch on a store copy — `.superpowers/sdd/heatmap-by-project-prototype.html`, https://claude.ai/artifact/2Dy9Ug6FqUpnkcNmNPahF9 (gymie + Big). It is the spec.
- Operator rounds (2026-10-28): split-into-panels rejected for "All / gymie / agents" buttons driving all graphs; then a Reviewers toggle, then Task size, then the defaults.

## Decisions

### project-buttons
**ID:** project-buttons
**Status:** decided
- The bar sets App's shared `project` (new optional `ViewProps.onProjectChange`), never a page-local filter: one filter, header and bar always agree.
- The list is `ReviewDynamics.Projects` — distinct `project_key` of changes with a reviewer dispatch in the period, ignoring the project and class filters — so the list does not shrink to the pick. Rejected: the `projects` table (holds 14 keys, most `tmp.*` test leftovers) and a new endpoint (no second caller).
- Buttons keep the last answer's list while a click refetches.

### class-buttons
**ID:** class-buttons
**Status:** decided
- `class` ∈ micro, small, regular, big (plan-class.sh's classes; the store holds exactly these plus 8 class-less decisions). A change's class is its latest decision's `decision->>'class'`; a change with no decision is excluded under any class.
- Server side, as `$5` on every review-dynamics query (`classFilterSQL`), so per-change rates stay correct; the stats endpoint rejects `class` on other views and any other value with 400, as it does an unknown parameter.

### retired-slots-hidden
**ID:** retired-slots-hidden
**Status:** decided
- Current slots are `CURRENT_SLOTS` in `ReviewDynamics.tsx`, the same six `panelSlots` `internal/api` accepts on a reviewer row. Filtered client side (heat cells and the `per-dispatch-slot` series); the data is already in the answer.

### four-week-default
**ID:** four-week-default
**Status:** decided
- The view-owned default range is Last 4 weeks.

## Live check

- `scripts/live-store-copy.sh up k1047 4175` from this worktree; open `/#/review-dynamics`: Projects lists agents, flow-eval, gymie, gymie-frontend, spectre; gymie + Big shows 43 changes reviewed and no retired rows; failure looks like the header field not following a click, the counts not changing with Task size, or bugbot rows under Current.
