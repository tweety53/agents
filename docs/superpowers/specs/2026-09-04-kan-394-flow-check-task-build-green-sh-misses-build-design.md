# kan-394-flow-check-task-build-green-sh-misses-build — design

## Context

KAN-394, from KAN-29's self-review: `check-task-build-green.sh` reports `task <id> has no
**Build:** tag` for a `**Build:**` line that carries prose after the keyword. KAN-29's archived
`tasks.md` in `gymie` holds eight such lines, not three — six of the shape
`**Build:** green — <prose>`, one `**Build:** green (unchanged; …)`, and one with no keyword at
all, `` **Build:** `tests/` 33 of 34 — see the flake below ``. Every one failed the guard for the
change's whole life, and the message named the consequence (no tag) rather than the cause (a tag
line whose value the grammar does not read).

No staged research note existed (`docs/superpowers/research/kan-394.md`, `kan-394-*.md` absent).
`spectre/specs/` is empty, so no capability spec is edited.

Measured against `main` at `cba777b`:

- The tag is read by one regex, `BUILD_TAG_RE` in `scripts/lib/plan_grammar.py:155`:
  `^\*\*Build:\*\*\s+(?P<kind>green|red)\s*$`. Both `scripts/check-task-build-green.py` and
  `scripts/check-task-commit-fields.py` read the tag through that module's `select_build_tag`;
  `scripts/plan-dispatch-bundles.py` recognises the `**Build:**` field name only to end a
  `**Files:**` bullet run and never reads its kind. One regex change fixes every reader.
- The issue's third bullet — "treat an unparseable line as a failure rather than a pass" — is half
  true. An unparseable line already fails: `select_build_tag` returns `None` for it, and the guard
  reports `task <id> has no **Build:** tag`. `skills/flow-contracts/build-green.md:44-46` and
  test case 7 in `scripts/test-check-task-build-green.sh` pin that "treated the same as no tag"
  behaviour deliberately, so changing it supersedes a recorded contract sentence, not only a regex.
  The operator chose to supersede it.
- `check-task-commit-fields.py` stores the tag as `build: Optional[str]` (line 339) and consults
  it only as `== "red"` (lines 585, 634, 656). A malformed tag reaching it as `None` changes
  nothing there.
- Test baselines: `scripts/test-check-task-build-green.sh` 67 `ok` lines,
  `scripts/test-check-task-commit-fields.sh` 213 `ok` lines, both passing.

## 1. The grammar

`scripts/lib/plan_grammar.py` replaces `BUILD_TAG_RE` with two patterns:

- `BUILD_LINE_RE = ^\*\*Build:\*\*(?P<value>.*)$` — a column-0 `**Build:**` line. The first
  non-fenced match in a task's body is the task's tag line, whatever its value says.
- `BUILD_KIND_RE = ^\s+(?P<kind>green|red)\b` — applied to that line's `value`. The keyword is a
  prefix ending on a word boundary; everything after it is ignored, so `green — prose`,
  `green (unchanged)` and `red — see below` read as `green`, `green`, `red`, while `greenish`,
  `yellow` and `` `tests/` 33 of 34 `` read as no kind.

`BuildTag` becomes `(offset: int, kind: Optional[str], value: str)`: `kind` is `None` when the
line matched `BUILD_LINE_RE` but its value did not match `BUILD_KIND_RE`; `value` is the text
after `**Build:**`, stripped, for the violation message. `select_build_tag` returns `None` only
for a body with no non-fenced `**Build:**` line at all.

Because the first `**Build:**` line decides, a malformed line followed by a well-formed one is a
malformed tag, not the later line's kind. Today the later line wins silently; the contract's
"first line matching the vocabulary" rule was written for the closed-word grammar and is restated
as "first `**Build:**` line" below.

## 2. The guard

`scripts/check-task-build-green.py`:

- `Task` gains `tag_value: Optional[str]`; `parse_tasks` copies `kind` and `value` from the
  `BuildTag` and sets `tag_line` whenever a tag line exists, malformed or not.
- `check_tasks` reports, before the "no tag" branch, a task whose tag line exists with
  `tag_kind is None`:

  `<file>:<tag_line>: task <id> has a **Build:** line reading "<value>", which is neither green nor red`

  and `continue`s, exactly as the "no tag" and unclosed-fence branches do. The "no tag" message and
  its line (the task line) are unchanged for a body with no `**Build:**` line.
- The module docstring's regex and its "treated the same as no tag" paragraph are rewritten to
  state the two-pattern grammar and the new violation.

`scripts/check-task-commit-fields.py`: no logic change — `build=tag.kind if tag is not None else
None` already yields `None` for a malformed tag. Its docstring (lines 122–127) and the comment at
line 479 stop saying "in full" and name the prefix grammar instead.

## 3. The contract

`skills/flow-contracts/build-green.md`:

- Line 35's regex sentence names the two patterns and the prefix rule: the keyword ends on a word
  boundary and the rest of the line is free text, still line-scoped.
- Lines 44–46 are rewritten: the tag is the **first** column-0 `**Build:**` line in the body; a
  body with none has no tag; a first line whose value does not open with `green` or `red` is a
  **malformed tag**, reported by its own line and value rather than as a missing tag — the same
  cause-not-consequence reasoning the unclosed-fence finding already carries.

## 4. Tests

`scripts/test-check-task-build-green.sh`:

- Case 7 asserts the malformed-tag message for `**Build:** yellow` (exit 1 unchanged).
- New cases, each one fixture and one `run_guard`: `**Build:** green — prose` → exit 0;
  `**Build:** green (unchanged; …)` → exit 0; `**Build:** red — prose` with
  `**Squash-with:** Task 2` and a green task 2 → exit 0; `**Build:** greenish` → exit 1 with the
  malformed message; `` **Build:** `tests/` 33 of 34 — see the flake below `` → exit 1 with the
  malformed message quoting that value; `**Build:** yellow` followed by `**Build:** green` → exit
  1, malformed, naming the first line.

`scripts/test-check-task-commit-fields.sh`: one new case — a `**Build:** red — prose` task with
`**Squash-with:** Task 2` folds into task 2's commit exactly as a bare `red` does.

Each script's `ok`-line count rises by the cases added; both must end `all cases passed`.

## Decisions

### An unparseable `**Build:**` line is a malformed-tag violation, not a missing tag

**ID:** malformed-build-tag-is-its-own-finding
**Status:** active
**Chosen:** distinct violation naming the line and its value; the first `**Build:**` line in a
body decides — the guard names the cause, as it already does for an unclosed fence.
**Considered:** regex-only change, keeping `task <id> has no **Build:** tag` for a malformed line —
smaller diff, but leaves the guard naming the consequence and hiding the cause, the defect this
change exists to remove; ruled out by the operator.

### The keyword is a word-boundary prefix

**ID:** build-keyword-prefix-word-boundary
**Status:** active
**Chosen:** `^\s+(green|red)\b` on the value after `**Build:**`; trailing text ignored.
**Considered:** requiring whitespace or end-of-line after the keyword — rejects `green.` and
`green,` for no gain; `\b` is the stdlib answer and rejects `greenish` just as well.

## Open questions

None.
