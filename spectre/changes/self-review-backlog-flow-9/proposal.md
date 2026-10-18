# self-review-backlog-flow-9

## Why

The next five live tickets in KAN's "Backlog Flow" status, oldest first, filed by earlier
self-review passes and not already fixed on `origin/main`, are planned here as one self-review fix
change (`design.md`, **Context**):

- **KAN-886** — four refusals of the kan-778 guard ports have no test: the null state record in
  `check-panel-reproducers`, the null finding status in `check-unfinished-work`, and an unset
  `FLOW_GUARD_REPO_ROOT` in `check-references` and `check-installed-citations`; removing any one
  keeps its guard's tests green, and losing the first fails open.
- **KAN-889** — `check-task-commit-fields`' commit-map refusals of a malformed entry and of a
  worktree named twice have no test, and no map case resolves its plan through a satellite's
  `link.md`.
- **KAN-890** — `check-panel-fix-single-dispatch` reads only the invoking session token's rows, so
  a fix round resumed under a fresh token records a `-retry` the guard calls a retry with no
  original.
- **KAN-896** — `check-verbatim-moves` stands a list item's first line alone as a block, so a
  sentence wrapped across it is cut in two, and an insertion ahead of it on that line flags the
  untouched sentence as reworded.
- **KAN-898** — the reviewer row's `-model`/`-effort` paragraph in **review-panel.md** names the
  decision pair with no harness exception, and the store refuses that pair on a `zcode` session
  token.

## What changes

- The four ported-guard refusals each get a subtest that fails when the refusal is removed.
- `check_task_commit_fields_test.go` gains cases 160–162: a map entry with no sha, a worktree
  named twice, and a link-route satellite under the commit map.
- `check-panel-fix-single-dispatch` counts an original another session token of the change
  recorded under a key as that key's original — for a `-retry`, a chunked round's bare key and its
  chunk contiguity — while every dispatch count stays the invoking token's.
- `check-verbatim-moves` joins a list item's indented continuation lines into the item's block.
- **review-panel.md**'s reviewer-row paragraph states the `zcode` and `muse` recorded pairs.
