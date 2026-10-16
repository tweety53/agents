## Context

- Three non-`big` findings from the kan-960 self-review pass, planned as one change per item 1 of
  **3. Fix every finding that is not `big`** (`skills/flow-self-review/SKILL.md`). Each is carried by
  exactly one task; none is dropped. Checked against this worktree at `7890df95` and the gymie
  worktree `/Users/tweety53/Projects/gymie-worktrees/kan-960-group-session-floor-hidden-tab-polling`
  at `b4efac39`:
  - A (Task 1): `stats/internal/store/records.go` declares `const minInlineWindow = 10 * time.Second`
    and `EndDispatch` refuses a `completed` inline row with no token report closing under it.
    `grep -rn 'minInlineWindow\|ErrInlineWindowTooShort\|ten seconds'` over `stats/`, `scripts/`
    and `skills/` (excluding `node_modules`) finds the constant only in `records.go`, the test
    `TestEndDispatchRefusesAZeroWidthInlineRow` in `records_test.go`, and the sentence in
    `skills/flow/implement.md`; `reconcile.go`, `server.go` and `record_test.go` name only the error
    and need no change. Blast radius 3.
  - B (Task 2): the implementer dispatch paragraphs in `skills/flow/implement.md` say nothing on
    where a throwaway harness lives. Blast radius 1. The stray worktree and branch were removed by
    hand during the pass.
  - C (Task 3): gymie's `docs/audits/kan-924-call-efficiency.md` F16 row records the measurement
    (Playwright's focus emulation keeps every page `visible`; CDP activation fires neither `blur`
    nor `focus`; the pause rides the window `blur`). Gymie's `CLAUDE.md` and `AGENTS.md` carry the
    same hand-maintained Frontend section (`diff` of the two sections is empty), every gymie
    session reads one of them, and its "A green suite is not evidence of a frame" bullet already
    names CDP screencasts. Blast radius 2 (the mirror), not the 1 the pass counted.
- **Task 3 reaches outside `<agents repo>`.** Per **Pipeline defects found mid-run**
  (`skills/flow-contracts/pipeline.md`) it commits on the change's own branch in that repository's
  worktree — `spectre/kan-960-group-session-floor-hidden-tab-polling` in the gymie worktree above —
  never in this agents worktree.
- Spiked at plan time, every edit reverted: Task 1's diff applied here — `gofmt -l ./internal/store`
  prints nothing, `go vet ./internal/store/` passes, `go test ./internal/store/ -run
  TestEndDispatchRefusesAZeroWidthInlineRow -count=1 -v` passes, and with `records.go` reverted it
  fails at `records_test.go:3990` (`closed 7s after its begin`). Tasks 1 and 2 applied together:
  `check-references.sh`, `check-markdown-integrity.py`, `check-dispatch-paragraphs.sh`,
  `check-installed-citations.sh` and `check-ask-silence.sh` exit 0, and `check-verbatim-moves.sh`
  prints exactly the four sentences this change's `verbatim-moves.txt` acknowledges. Task 3's diff
  applied in the gymie worktree and `git apply -R --check` passes.

## Decisions

### The inline floor is three seconds

**ID:** inline-floor-three-seconds
**Status:** active
**Chosen:** `minInlineWindow = 3 * time.Second` — far above the milliseconds a begin and end in one
Bash call land apart, the case the floor exists to refuse, and under the 7.6 s a real cached verify
took.
**Considered:** one second — a begin and end in two consecutive Bash calls can land a second or two
apart and would pass; five seconds — no measured inline work needs it, and every second raised is
another legitimate quick close refused; dropping the floor — loses the zero-width refusal entirely.

### The scratch rule is a dispatch paragraph, unguarded

**ID:** scratch-harness-paragraph
**Status:** active
**Chosen:** one `SCRATCH HARNESS` blockquote among the implementer dispatch paragraphs of
`skills/flow/implement.md`, after REPRODUCE, DON'T READ.
**Considered:** a row in `check-dispatch-paragraphs.sh`'s required table — a guard for a one-off
paragraph nobody has trimmed; a cleanup step that sweeps unknown worktrees — cleanup cannot tell a
stray from an operator's own worktree.

### The Playwright note lives in gymie's instruction files

**ID:** playwright-note-in-claude-md
**Status:** active
**Chosen:** one bullet in the Frontend section of gymie's `CLAUDE.md` and its `AGENTS.md` mirror,
after "A green suite is not evidence of a frame", committed on gymie's
`spectre/kan-960-group-session-floor-hidden-tab-polling` branch.
**Considered:** `.flow/project.md` — only its command sections reach a dispatch bundle; a new
`docs/` file — read only when something points at it; a `flow hazard add` row — a store write, not a
reviewed commit, and the pass asked for a docs note; `gymie-playwright`'s README — outside the
change's gymie worktree the pass named.

### First review effort and fixer grouping

**ID:** review-effort-and-grouping
**Status:** active
**Chosen:** first review on `flow-low` (effort `low`); one fixer dispatch (key
`self-review-kan-960-group-session-floor-hidden-tab-polling-fix`) carrying Tasks 1–3 — each task
is a verified diff applied with `git apply`, three small file-disjoint edits across two
repositories.
**Considered:** `flow-medium` for the first review — nothing in three applied-and-tested diffs needs
the deeper pass; one fixer per task — three dispatches for edits that are each a `git apply`.

## Open questions

## Live check

None for this change's run: Task 1 changes a store constant its own Postgres-backed test pins, and
Tasks 2 and 3 are prose.
