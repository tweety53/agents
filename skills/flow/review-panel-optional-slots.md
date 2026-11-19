# Review panel — optional slots

Loaded by `skills/flow/review-panel.md` only for a round that dispatches `mutation`, always a bundle alone. Every
section name below without a path is a section of `skills/flow/review-panel.md`.

## The mutation slot's prompt

The mutating
role (`mutation`) works in its
throwaway copy (**The throwaway worktree**, `skills/flow/review-panel-optional-slots.md`); every other bundle reads the shared
`<worktree>`.

**The mutation slot's dispatch prompt also carries the MUTATION ENTRY CONTEXT paragraph**, every
test the plan's `**Tests:**` fields name and every test file in the touched-files list inlined
directly beneath it:

> **MUTATION ENTRY CONTEXT:** `final-review.diff`, at the merge-base sha each `# worktree:`
> header carries, is your diff base; the tests beneath this paragraph are the tests your
> mutations target. Begin the brief's search for a behaviour's covering test from those two —
> never a whole-tree sweep. Read a test outside the list only when the behaviour you are mutating
> cannot otherwise be judged caught or survived, and name that read in your report.

## The throwaway worktree

**Independent multi-slot detection is the
panel's core signal**: Mutation's
dispatch, carrying the mutation-testing brief, therefore runs
against its own throwaway worktree, never the
shared `<worktree>` the other slots read, so every slot's findings are raised against the same
pristine snapshot and no slot's view ever contains another slot's work:

**The parent creates and removes every throwaway copy itself, in its own Bash calls — never a
subagent.** Run `flow-guard throwaway-worktree create <worktree> <worktree>-<slot>-<round> --sdd` once per
worktree in the resolved set per slot, producing one `<worktree>-<slot>-<round>` per repository per
slot — `<slot>` is the id (`mutation`). Exit 2 means a step failed, named on stderr: that slot is
not dispatched against that copy.

Never symlink a shared `node_modules` into a copy: the copy's own `npm ci` installs one, and over a symlink it empties the shared directory — `<agents repo>/stats/Makefile` refuses to run it there.

Dispatch each slot present in this round's panel **once**, its brief hand-assembled
(**Bundled dispatch**, `skills/flow/review-panel.md`) from the shared resolved set, never from a copy: the
brief names each copy, `<worktree>-<slot>-<round>`, as the repository path to mutate and test
in, and never lists a worktree whose name already ends `-mutation-<round>` (kan-1119 self-review, F1).
Remove every copy unconditionally once that slot's dispatch closes — completed, timed out
(including after the wall-clock re-dispatch), or the run stopped — with
`flow-guard throwaway-worktree remove <worktree> <worktree>-<slot>-<round> --fold-back`, run as its own Bash call with literal absolute paths, never chained into a compound command, which first copies
each non-empty `<worktree>-<slot>-<round>/.superpowers/sdd/panel-report-*.md`,
`<worktree>-<slot>-<round>/.superpowers/sdd/reproducers/*.sh`,
`<worktree>-<slot>-<round>/.superpowers/sdd/dispatch-context-group-*.md` and
`<worktree>-<slot>-<round>/.superpowers/sdd/implementer-report-*.md` back to the same path under `<worktree>`, unless `<worktree>` already holds a non-empty file the copy's is not newer than.
It also folds a non-empty `panel-report-*.md` found at the root of the copy or of `<worktree>` — a report a slot wrote by a relative path after a `cd` — into the directory the copy's own reports fold back to, by the same rule, moving the one at `<worktree>`'s root, so a misplaced report is never moved by hand (kan-1119 self-review, F2).

The fold-back runs inside the removal step itself — on every path that removes a copy,
completed, timed out or run stopped — because a report the slot resolved onto its dispatched root
dies with the copy.

Findings are unaffected: a finding's `file:line` is repo-relative.
