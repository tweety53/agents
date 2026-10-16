# pipeline-fix-kan-954-user-row-loaded-twice-per-request-14

## Context

`checkLoopDispatch` (`stats/internal/guard/loop.go`) admits a loop's `review-1` only on the
subagent type `lpFirstReviewType` returns: the loop plan's `first-review-effort` decision, read from
`spectre/changes/<loop branch>/design.md` on the loop's branch (or the change's own branch for an
own-branch self-review). Today it takes the first `**Chosen:**` line after the decision's `**ID:**`,
matches a backticked `flow-low` or `flow-medium` there, and returns `flow-medium` when nothing
matches.

Two defects in that read:

1. **The value form.** `skills/flow-self-review/SKILL.md` step 3 item 1 tells the planner to record
   "the first review's effort, `medium` or `low`" — a bare effort — and `**Chosen:**` is written
   `<option> — <rationale>` (`skills/flow/brainstorm-planner.md`, **Decisions**). A planner that
   follows either text is misread.
2. **The entry bound.** The `**Chosen:**` search runs over the rest of the file, so an entry missing
   its own `**Chosen:**` line borrows the next decision's.

Reproduced at plan time at be156cf7: with the new test alone (code untouched), 6 of
`TestCheckLoopDispatchFirstReviewEffort`'s 7 cases fail — the bare and backticked-bare `low` forms
are refused as `flow-medium`, and the no-effort cases are admitted on `flow-medium`; with the task's
whole diff, `cd stats && go test ./internal/guard/ -count=1` is green.

Blast radius — not `big`:

- Code: 2 files name `lpFirstReviewType`, `lpReviewType` or `first-review-effort` outside `.git`
  <!-- measured: grep -rl -e lpFirstReviewType -e first-review-effort -e lpReviewType --exclude-dir=.git --exclude-dir=archive . @ be156cf7 -->
  and the spectre archive — `stats/internal/guard/loop.go` and its test — plus the refusal list in
  `scripts/check-loop-dispatch.sh`'s header; one caller of `lpFirstReviewType`.
- Recorded plans: 21 `first-review-effort` decisions across the agents tree and gymie's — 18 read
  the same before and after, 3 recorded `low` and were read as `flow-medium` (`self-review-kan-980`,
  `self-review-kan-959`, gymie `self-review-kan-954`), plus `pipeline-fix-13`'s on its loop branch;
  the new read refuses none of them.
  <!-- measured: a perl transcription of the old and new parse over every design.md under spectre/changes carrying the first-review-effort ID, in the agents worktree @ be156cf7 and gymie @ 1419b4f4 — agents: 2 low to low, 16 medium to medium, 2 medium to low; gymie: 1 medium to low; 0 refused -->

## Decisions

### Accept the effort as the skill names it

**ID:** bare-effort-accepted
**Status:** active
**Chosen:** a bare `low`/`medium` opening the `**Chosen:**` value wins; otherwise the line's first
backticked `low`, `medium`, `flow-low` or `flow-medium`; each maps to `flow-low`/`flow-medium` —
the bare check runs first, so a value `low` whose rationale mentions a backticked `medium` reads as
`low`.
**Considered:** rewording the skill to demand `flow-low` — the skill's bare form is the spec and the
regex the deviation, and two planners already wrote the bare form; any whole-word `low`/`medium`
anywhere on the line — a rationale such as "low risk" would pick the effort.

### Refuse an unparseable recorded decision

**ID:** unparseable-refuses
**Status:** active
**Chosen:** a recorded `first-review-effort` entry whose `**Chosen:**` names no effort (or is
missing) refuses `review-1` with an `EFFORT —` refusal naming the forms to write; a plan with no
such entry keeps the skill's `flow-medium`. A silent fallback is the defect this loop exists for;
the refusal names the fix, and no recorded plan trips it (measured above).
**Considered:** keep the silent `flow-medium` fallback — the planner's choice is lost without a
trace, the failure reported here; reuse the `PAIR` class — its header meaning is "off the suffix's
row", which an unreadable plan is not.

### Bound the read to the decision's own entry

**ID:** entry-bounded
**Status:** active
**Chosen:** the `**Chosen:**` search stops at the next `**ID:**` — one slice, so an entry missing
its line never borrows the next decision's effort.
**Considered:** leaving it — a missing line would pass on whatever effort the following decision
happens to mention instead of reaching the refusal.

### Skill text unchanged

**ID:** skill-text-unchanged
**Status:** active
**Chosen:** no edit to `skills/flow-self-review/SKILL.md` — its "`medium` or `low`" and
"`flow-medium` when it recorded none" both hold once the guard reads the bare form.
**Considered:** listing the accepted spellings in the skill — new run-loaded prose for what the
`EFFORT` refusal already tells the one reader who needs it.

### First review effort

**ID:** first-review-effort
**Status:** active
**Chosen:** `flow-low` — one guard function, one caller, a comment line and one table test, every
behaviour case shown failing without the fix; no concurrency, data-model or store seam.
**Considered:** `flow-medium` — for a change whose seam a cheap review could miss; this one has none.

### Fixer grouping

**ID:** fixer-grouping
**Status:** active
**Chosen:** one fixer, one task — the three files move together in a single commit.
**Considered:** splitting the script header into its own task — a comment with no behaviour of its
own, describing the refusal Task 1 adds.

## Open questions
