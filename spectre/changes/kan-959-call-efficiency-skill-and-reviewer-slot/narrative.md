# kan-959-call-efficiency-skill-and-reviewer-slot — session narrative

## 2026-10-15 — creating run (resumed at STARTED, inline)

- Operator renamed the measuring skill `/call-efficiency` → `/call-efficiency-audit` at resume; plan artifacts edited before task 1. The panel slot keeps the id `call-efficiency`.
- Task 2's planned render test omitted the required `-session-model` flag; added it (correction recorded).
- Task 3's planned verify contradicted itself: moving `experimental/test-sanity.md` broke the existing exp-slot render subtest, which Task 5 was meant to own. Rewrote that subtest to render a probe candidate from a copy of the skill's top-level files and folded it into Task 3.
- Forward citations to **The always-on extra dispatch** stayed red in `check-references.sh` from Task 3 until Task 7 added the heading.
- The new skill tripped three coverage guards (`check-references`, `check-stage-mark-calls`, `check-guard-symlinks`) — each now declares it expected-zero; the last two were only found in Task 7's full lint run.
- `verbatim-moves.txt` treats a leading `#` as a comment, so heading sentences are listed backslash-escaped.
- The installed `render-slot-prompt.sh` (main's build) cannot render the new slots; the extra bundle's brief was rendered with the worktree's own shim.
- Panel round 0: primary raised 2 Important (stale Decide step citations; recording rules contradicting the extra pair) + 5 Minor, principles 2 Minor on the same roots; test-sanity and call-efficiency clean. Fixed inline in one commit; sonnet/low re-review clean.
- The inline panel-fix row's begin was recorded after the work, so its end was refused by the 10 s window; closed `aborted` with a pass note.
- Not done here: the Models-rule carve-out in `~/.claude/CLAUDE.md` (decision `extra-dispatch-sonnet-carve-out`) is the parent's step after landing.

## 2026-10-15 — integrate run

- Preflight clean; unfinished-work gate clear; `origin/main` unmoved since the merge base — no rebase.
- Route: merge and push, from the project's configured default.
