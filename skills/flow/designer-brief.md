# The designer

Loaded by the session that dispatches a `designer` (**Every visible element a task builds is
drawn**, `skills/flow/brainstorm-ui.md`), never by a planner.

**Before asking anything, read `## Undrawn`** in the round
file's design draft: non-empty, or a designer review open, the parent refuses to ask the
convergence confirm, runs `flow-guard design-repo <abs-worktree>` and, on exit 0, dispatches
the one `designer` in this same turn;
on exit 1 the next round has the planner draw the items into the prototype. A final dispatch
that leaves `design.md`'s `## Undrawn` non-empty is handled the same way, then a fresh final
dispatch takes the drawn frames.

**Whenever a planner
return leaves `## Undrawn` non-empty (the round file's design draft, or `design.md`), the session
that dispatched it dispatches one `designer`** for the whole list in
the same turn, unasked; the next planner dispatch replaces each item with the frame id the
designer returned. With a design
repo the only operator step for the items is **The designer review** (below), which ends before
the convergence confirm (**Convergence**, `skills/flow/brainstorm-planner.md`) is asked.

**One `designer` dispatch draws every `## Undrawn` item** (**Every visible element a task builds is
drawn**, `skills/flow/brainstorm-ui.md`), key `designer-<k>`, `<k>` counting from 1 per change, on the pair the pair judge picks for it
(**Model and effort**, `skills/flow-contracts/model-policy.md`, the prompt saved as `pair-judge/designer-<k>.md` first — a small
adjustment round is judged on its own prompt, never given the role's first pair by default), on a
`flow-*` agent type, since it writes. Its prompt carries the two
agent-baseline sentences, the report-tersely line, `Role: designer`, the `## Undrawn` list
verbatim, the `<clone> <remote>` `flow-guard design-repo` printed, the frames already cited, and on an adjustment the operator's words verbatim. The
designer:

1. works in a git worktree of that design-repo clone, `<design clone>-worktrees/<change
   name>` on branch `<change name>` — created from the clone's default branch when absent, reused
   as it stands when present;
2. reads that repository's own instructions first — its `CLAUDE.md`, `AGENTS.md`, README, style
   guide and tokens — and follows them exactly, inventing no style they do not state;
3. runs that repository's own stack where its instructions name one (a local server, a build, a
   render tool — gymie-design's is a local node server and Playwright), from its worktree, on a
   port of its own and never touching a server the operator runs; where they name no server, it
   serves the worktree with a plain static one. The server the review page (item 4) is served on
   stays running until the operator approves (**The designer review**, below);
4. draws each listed item as frames in that repository's own frame format, its copy in the voice
   those instructions set, and where they give no answer the simplest plausible copy, listed in
   its report — it never asks. Every frame it draws carries the description that repository's frames
   carry, in the same form (gymie-design's: id, title, a behaviour caption, a LINKED line) — never
   fewer than the frames beside it have. **Every drawn frame is part of the repository's current
   pages** — placed on the existing page and section its instructions say such a frame belongs to,
   never an orphan file — **and every new thing is also collected on one separate review page** for
   the change, in the repository's own page format, named after the change: each new frame with its
   description and a link to where it lives in the current pages, so the operator opens one page and
   searches for nothing. That review page is the only page of the review: every drawn or changed
   frame is on it, rendered in full, with its id, title, behaviour caption and LINKED line, and the
   operator needs no other URL, page or published copy. Where the stack renders, it renders every drawn frame,
   reads each image and fixes what the instructions' own checks would flag. Before it reports it
   opens the review page and checks that every `#<frame id>` anchor resolves and every frame
   renders;
5. commits on the branch, tags that commit `<change name>-designer-<k>` (`<k>` the dispatch's) and
   pushes the branch and the tag — `flow-guard render-frames` resolves a frame at a tag only
   (`skills/flow-contracts/project-configuration-visual.md`, **Frames at a design-repo tag**) — and
   reports the review page's URL, once with a `#<frame id>` anchor per drawn or changed frame, each
   with the frame's title — the only URL in the report: no link to another page, no PNG, no
   Artifact copy; a report without it is incomplete and the planner sends the designer back for
   it. It also reports each frame id and path drawn, on an adjustment the frame ids changed, the
   copy it wrote without a rule, the tag and the commit sha. The planner cites each frame as
   `<mockups root>/<tag>/<frame id>.png` with that tag; an adjustment's dispatch cites its own,
   newer tag.

The dispatching session appends that report to the latest round file's `## Answers`, where the next planner dispatch reads it. The designer is outside **The cap** (`skills/flow/brainstorm-planner.md`).

**The designer review** is an interactive pre-step of the approval, run by the session that
dispatched the planner after every return that drew or changed frames. That session prints the
review page's anchored URLs, every one in full, first — the operator reviews on that live page, and
is given no other URL, no PNG and no Artifact copy — then, after an adjustment, the frame ids that
changed, then the design-repo branch and commit.
It then asks through **AskUserQuestion**: **Continue with the planning** *(recommended)* or the
operator's own words on what to change. Words naming a change are one more `designer` dispatch on
the same worktree, carrying them verbatim; it applies them on the review page, keeps its server
running and reports which frame ids changed, then the review runs
again on its return; **Continue with the planning** ends the review, the session stops the server the
designer left running, **lands the design branch** and planning goes on. No convergence confirm is
asked while a review is open, and the confirm never carries the drawn frames.

**Landing the design branch is the approval's own step, unasked.** From the designer's worktree the
session pushes `HEAD:<default branch>` of the `design repo` clone — a fast-forward, the default
branch being what `git symbolic-ref --short refs/remotes/origin/HEAD` names with its `origin/` prefix
dropped — after rebasing the branch onto a fresher `origin/<default branch>` where the push is
refused as non-fast-forward, then runs `git -C <design clone> pull --ff-only`. The branch and its
`<change name>-designer-<k>` tags stay pushed, so `flow-guard render-frames` still resolves each
frame at its tag. A rebase conflict or any other refusal is a defect fixed in the run, never a stop.
The unapproved branch is never landed: a review that never reaches **Continue with the planning**
lands nothing. Where the review is not asked (below) the drawn frames stand and the same landing
runs once the plan is written.

Where the confirm is not asked (`## decisions: recommended`) the review is not asked either: the drawn frames stand.
