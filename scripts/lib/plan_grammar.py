"""plan_grammar.py — the one definition of the `tasks.md` structural
grammar that `check-task-build-green.py` and `check-task-commit-fields.py`
both read.

Why this module exists (fix round 6, F13). Both guards parse the SAME
`**Squash-with:**` field out of the SAME `tasks.md`, and each carried its
own regex literals for it. Kept in step only by a comment in each file
asserting parity with the other, they drifted twice:

  * fix round 5, F9 — one guard gated the field's whole value, the other ran
    a bare id pattern over the raw remainder of the line, so free text
    became partner ids there and not here. The comment asserting parity was
    itself wrong, and that false claim survived four review rounds.
  * fix round 6, F14 — one guard gated ONE PHYSICAL LINE, the other gated
    the value only after joining continuation lines, so the same plan was
    accepted by one guard and rejected by the other, in both directions.
  * fix round 8, F19 — the patterns were shared by then, but each guard
    still ran its OWN loop to decide WHICH line was the field, and the two
    loops disagreed: one committed to a line only once its value gated, the
    other took the first field-shaped line whatever its value. A body
    carrying a non-gating line ahead of a gating one therefore had a
    different field in each guard, and the plan passed one and failed every
    task in the other. The comment written in fix round 6 asserting the two
    "read one field the same way" was, like the one before it, wrong.

A comment cannot hold two grammars together; an import can. Every pattern
below is imported by both guards and defined nowhere else, and so is
`select_squash_with`, which is the whole of WHICH line is the field —
neither guard loops over a task body looking for it any more.

Scoping: `**Squash-with:**` is LINE-SCOPED
-----------------------------------------

A `Squash-with:` field's value is exactly what stands on its own
`**Squash-with:**` line — never a following continuation line, even where
the guard reading it joins continuation lines for its other fields:

  * the value's grammar is closed and short (`Task <id>[, <id>...]`), so it
    has no reason to wrap, unlike the prose and path lists that `Files:` and
    `Tests:` carry;
  * under the joined reading an unrelated prose line placed under the field
    with no blank line between silently changes what the field means, which
    is a non-local dependency invisible at the point the field is read;
  * check-task-build-green.py has always been line-scoped, and it is the
    guard that gates plans at lint time. Line scoping is therefore the rule
    both guards can hold without either changing what it accepts today.

A field wrapped across two lines is consequently a plan defect rather than a
longer field: its second line names nobody, and the guards fail it — the
commit check on the file union the narrower partner list implies, the lint
check by validating only the partners actually named on the line.

What this module defines
------------------------

Every structural element BOTH guards read out of a `tasks.md`, pattern AND
selection: which lines are a task, which task an id names, where a
task's body starts and ends, which line is its `**Build:**` tag, which line
is its `**Squash-with:**` field, and — `unclosed_fence`, fix round 11 —
whether the body opens a fence it never closes, which is the one condition
under which every selection after that line correctly finds nothing. Sharing a pattern while each guard
kept its own selection loop is what F19 was — and, for `**Build:**`, what
F20 was: one guard gated the whole physical line and took the FIRST such
line, the other read the value through its field alternation, joined
continuation lines onto it, and took the LAST. A body carrying `**Build:**
red` then `**Build:** green` was therefore red in one guard and green in the
other, so one resolved the fold while the other checked the folded commit
against a subject the fold had deleted.

`FENCE_RE` below is what every selection in this module uses, what
check-task-build-green.py uses throughout, and what
check-task-commit-fields.py's task-body field loop uses since fix round 10
— before which that loop gated a fence on a bare "^```", so a `~~~` or an
indented example block was invisible to it alone and the `**Files:**`
written inside a worked example overwrote the task's real declaration
(F21). check-task-commit-fields.py's remaining `FENCE_LINE_RE` parses
`.myflow/project.md`'s `## test` block, which is a different file with its
own cases.

Standard library only (see check-plan-provenance.py's module docstring for
why this repository restricts itself to that).
"""

from __future__ import annotations

import re
from typing import List, NamedTuple, Optional, Pattern, Sequence

# DOTTED_ID — the dotted id grammar ("1", "1.1", "9.9.2", ...). It is NOT a
# task id: `TASK_ID` below is, and it is flat. This pattern is every
# `Squash-with:` partner id — a dotted partner therefore names no task and
# is reported as a dangling reference, which is the correct answer rather
# than a gap — and (in check-task-commit-fields.py) the task-id-shaped
# commit scope it refuses, where the dotted form has to stay recognisable
# precisely so a scope written as `1.2` is still caught.
DOTTED_ID = r"\d+(?:\.\d+)*"

# TASK_ID — a task's own id: a flat integer, and spectre's exactly
# (`(\d+)` in its internal/parse/tasks.go). A dotted id on a task line is a
# "malformed task line" finding to spectre and no task at all, so admitting
# one here would recreate, in a narrower form, the very disagreement the
# checkbox line ended: myflow would dispatch and gate a task spectre never
# counts. Sub-tasks are therefore renumbered flat rather than written
# `1.1`.
TASK_ID = r"\d+"

# FENCE_RE — a fenced code block delimiter: three or more backticks or
# tildes, optionally preceded by up to 3 columns of leading whitespace (per
# CommonMark). Matching this line toggles fence state; a task line, a
# `**Build:**` tag line or a `**Squash-with:**` field line inside a fence is
# example text, not real structure. It lives here because
# `select_squash_with` below has to skip fenced candidates, and a second
# copy of the rule in a guard is what this module exists to prevent.
FENCE_RE: Pattern[str] = re.compile(r"^ {0,3}(`{3,}|~{3,})")

# TASK_LINE_RE — opens a task: the spectre checkbox line `- [ ] <id>. <title>`.
# Group "state" is the character between the brackets, a space for an open
# task and an `x` for a done one; group "id" is the task's id.
#
# This is spectre's own task grammar (`^- \[([ x])\] (\d+)\. (.*)$` in its
# internal/parse/tasks.go) character for character, and it replaces the
# `### <id> <title>` heading a myflow plan used to mark a task with.
# Recognising BOTH shapes was rejected: while the two grammars disagreed,
# spectre saw no tasks in a real myflow plan at all — `spectre list`
# reported 0/0, `spectre validate` reported one false "malformed task line"
# per step checkbox, and `spectre archive` refused with "tasks.md has no
# tasks" — and a grammar admitting a shape spectre rejects would leave that
# disagreement standing.
#
# The trailing space after the dot is spectre's too, and is what keeps the
# two tools reading the same line as a task: spectre's own malformed-task
# check is `^- \[[ x]\] \d+\. `, so `- [ ] 1.` with nothing after the dot
# is a finding there and, here, no task.
#
# The anchor is column 0, spectre's again: its malformed-task check tests
# the RAW line for a `- [` prefix. A task's STEPS are indented two columns
# beneath their task (`  - [x] **Step N: ...**`), which is what puts them
# outside this pattern AND outside that check — the same one line of indent
# makes a step a step to both tools at once. An UNindented step would be a
# "malformed task line" finding in spectre and still not a task here: never
# mistaken for a task, but never clean either.
TASK_LINE_RE: Pattern[str] = re.compile(
    rf"^- \[(?P<state>[ x])\] (?P<id>{TASK_ID})\. "
)

# BODY_BOUNDARY_RE — closes the current task's body: any level-2 or level-3
# heading. The next task's own TASK_LINE_RE line closes it too; that is
# `iter_tasks` below, which tests for a task line before this pattern.
BODY_BOUNDARY_RE: Pattern[str] = re.compile(r"^#{2,3}(?:\s|$)")

# BUILD_TAG_RE — the `**Build:**` tag's vocabulary and its scope: the WHOLE
# physical line, whose value is a bare `green` or `red`. Group "kind" is
# that word. Anything else on the line (`**Build:** yellow`, or a prose line
# joined onto it) is not a tag, and a body carrying no tag line has no tag.
BUILD_TAG_RE: Pattern[str] = re.compile(r"^\*\*Build:\*\*\s+(?P<kind>green|red)\s*$")

# SQUASH_WITH_FIELD_RE — the field's PRESENCE on one physical line. Group
# "value" is everything after the field name, and is what `partner_ids`
# gates. Presence and gate are separate because a red task carrying a
# present-but-ungated value is a distinct condition from one carrying no
# field at all, and each guard reports it in its own vocabulary.
SQUASH_WITH_FIELD_RE: Pattern[str] = re.compile(
    r"^\*\*Squash-with:\*\*\s*(?P<value>.*)$"
)

# SQUASH_WITH_VALUE_RE — the GATE. A `Squash-with:` value is `Task <ids>`
# and nothing else, its ids clause being digits, dots, commas and whitespace
# only. The whole value is gated BEFORE any id is extracted from it: a bare
# id pattern run over free text turns any digit in it into a partner and
# silently widens a fold's `Files:` union to an unrelated task's declared
# files (fix round 5, F9).
SQUASH_WITH_VALUE_RE: Pattern[str] = re.compile(
    r"^Task\s+(?P<partners>[\d.,\s]+)\s*$"
)

# PARTNER_ID_RE — every dotted-id token inside a GATED ids clause, so
# `Task 2`, `Task 2.1, 3.4` and `Task 2.1 3.4` all resolve the same way
# without either guard hand-rolling a splitter. Applied to
# SQUASH_WITH_VALUE_RE's `partners` group and never to a raw field value.
PARTNER_ID_RE: Pattern[str] = re.compile(DOTTED_ID)


def partner_ids(value: str) -> Optional[List[str]]:
    """The partner ids a `Squash-with:` field's VALUE names.

    `value` is what `SQUASH_WITH_FIELD_RE` captured — the one physical
    line's remainder, per this module's scoping rule.

    Returns `None` when the value does not gate as `Task <id>[, <id>...]`;
    the caller reports that in its own vocabulary and MUST NOT fall back to
    extracting ids from the raw value. Returns a possibly-empty list when it
    does gate: `Task ,` gates and names nobody, which is a different defect
    from a value the grammar does not admit at all.
    """
    match = SQUASH_WITH_VALUE_RE.match(value)
    if match is None:
        return None
    return PARTNER_ID_RE.findall(match.group("partners"))


class SquashWithField(NamedTuple):
    """The one `**Squash-with:**` field a task body carries.

    `offset` is the 0-based index, within the body sequence handed to
    `select_squash_with`, of the line the field was read from — a caller
    holding an absolute file position adds its own body start to it.
    `value` is that line's stripped remainder. `partners` is what
    `partner_ids` made of it: a possibly-empty list when the value gates,
    and `None` when it does not.
    """

    offset: int
    value: str
    partners: Optional[List[str]]


def select_squash_with(body: Sequence[str]) -> Optional[SquashWithField]:
    """The `**Squash-with:**` field `body` carries, or None if it carries
    none. `body` is a task's body lines, in order, fences included.

    WHICH line is the field is defined here and nowhere else (fix round 8,
    F19): the first non-fenced line whose value GATES as `Task
    <id>[, <id>...]` is the field, so a non-gating candidate ahead of it is
    skipped rather than claiming the field and shadowing the real one.

    When no candidate gates, the FIRST candidate is returned with
    `partners=None`, because the two guards report that condition
    differently and both readings need the line: check-task-build-green.py
    treats a non-gating value as no field at all (its case 18) and so
    ignores anything with `partners=None`, while
    check-task-commit-fields.py reports the malformed value itself (its
    cases 49, 50 and 55) and so needs to see it. What each guard makes of a
    field it cannot gate is its own rule; WHICH line it is looking at is
    not.

    Line scoping is this module's docstring's rule: a value is one physical
    line's remainder, never a continuation line.
    """
    ungated: Optional[SquashWithField] = None
    in_fence = False
    for offset, line in enumerate(body):
        if FENCE_RE.match(line):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        match = SQUASH_WITH_FIELD_RE.match(line)
        if match is None:
            continue
        value = match.group("value").strip()
        ids = partner_ids(value)
        if ids is not None:
            return SquashWithField(offset=offset, value=value, partners=ids)
        if ungated is None:
            ungated = SquashWithField(offset=offset, value=value, partners=None)
    return ungated


class BuildTag(NamedTuple):
    """A task body's `**Build:**` tag. `offset` is the 0-based index, within
    the body sequence handed to `select_build_tag`, of the line it was read
    from; `kind` is "green" or "red"."""

    offset: int
    kind: str


def select_build_tag(body: Sequence[str]) -> Optional[BuildTag]:
    """The `**Build:**` tag `body` carries, or None if it carries none.
    `body` is a task's body lines, in order, fences included.

    WHICH line is the tag is defined here and nowhere else (fix round 9,
    F20): the FIRST non-fenced line matching `BUILD_TAG_RE` in full. The tag
    is line-scoped for the same reason `Squash-with:` is — its grammar is
    one closed word, so it never needs to wrap, and a prose line placed
    under it with no blank line between must not change what it says.
    """
    in_fence = False
    for offset, line in enumerate(body):
        if FENCE_RE.match(line):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        match = BUILD_TAG_RE.match(line)
        if match is not None:
            return BuildTag(offset=offset, kind=match.group("kind"))
    return None


class TaskBody(NamedTuple):
    """One task found in a `tasks.md`. `task_line` is the 1-based line
    number of its `- [ ] <id>. <title>` line; `body_start` is the 0-based index, in
    the file's line list, of the first body line, so a caller turns an
    offset within `lines` into a file line number by adding it and 1;
    `lines` is the body itself, fences included."""

    id: str
    task_line: int
    body_start: int
    lines: List[str]

    # The task line's checkbox STATE is deliberately not a field here. Both
    # guards that import this module answer questions about a task's
    # `**Build:**` tag and its `**Squash-with:**` field, neither of which
    # depends on whether the task is done; the one guard that does read
    # done-ness, plan-dispatch-bundles.py, mirrors TASK_LINE_RE rather than
    # importing it and reads its own `state` group. A field no caller reads
    # is a field no test can pin — the reviewer inverted this one's
    # assignment and all three suites still passed — so it is not carried
    # until a caller exists. TASK_LINE_RE's `state` group is where it comes
    # from when one does.


def iter_tasks(lines: Sequence[str]) -> List[TaskBody]:
    """Every task in `lines`, in document order, duplicate ids included.

    A task opens at a non-fenced `TASK_LINE_RE` line and its body runs to
    the next task line, the next non-fenced `BODY_BOUNDARY_RE` line, or end
    of file. The body is everything BELOW the task line — its fields, its
    prose and its two-column-indented step checkboxes — and a step is never
    a task of its own.
    Fence state is tracked across the whole file, so a task line, tag or
    field shown inside a fenced example opens no task and closes no body —
    it is documentation, not structure.
    """
    tasks: List[TaskBody] = []
    open_index: Optional[int] = None
    open_id = ""
    in_fence = False

    def close(end_index: int) -> None:
        if open_index is None:
            return
        tasks.append(
            TaskBody(
                id=open_id,
                task_line=open_index + 1,
                body_start=open_index + 1,
                lines=list(lines[open_index + 1 : end_index]),
            )
        )

    for index, line in enumerate(lines):
        if FENCE_RE.match(line):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        task_match = TASK_LINE_RE.match(line)
        if task_match:
            close(index)
            open_index = index
            open_id = task_match.group("id")
            continue
        if BODY_BOUNDARY_RE.match(line):
            close(index)
            open_index = None
    close(len(lines))
    return tasks


def select_task(lines: Sequence[str], task_id: str) -> Optional[TaskBody]:
    """The task `task_id` names, or None if the plan defines no such task.

    An id defined more than once resolves to the FIRST task line carrying it
    (fix round 9). A duplicate id is itself a plan defect
    check-task-build-green.py reports, but until it is fixed both guards
    have to read the same task out of it.
    """
    for task in iter_tasks(lines):
        if task.id == task_id:
            return task
    return None


def unclosed_fence(body: Sequence[str]) -> Optional[int]:
    """The 0-based offset, within `body`, of a fence delimiter the body
    opens and never closes — or None when every fence it opens is closed.

    An unclosed fence in a task body is ALWAYS a plan defect (fix round 11,
    F22). CommonMark runs an unclosed fence to the end of the containing
    block, so every `**Files:**`, `**Commit:**`, `**Build:**` and
    `**Squash-with:**` line after it is code, not a field — which is what a
    markdown renderer shows a human, and what every selector in this module
    already reads. The selectors are therefore right to swallow those
    lines; what is wrong is reporting the CONSEQUENCE. Without this, a
    swallowed `**Files:**` reaches check-task-commit-fields.py as an empty
    declaration and the commit is blamed for touching an undeclared file,
    and a swallowed `**Build:**` reaches check-task-build-green.py as a
    task with no tag. Both guards name the fence instead.

    The toggle is the one every selector here uses, delimiter kind
    included: this reports exactly the state those selectors were in, so
    the diagnosis matches what the parse actually did.
    """
    opened_at: Optional[int] = None
    for offset, line in enumerate(body):
        if FENCE_RE.match(line):
            opened_at = None if opened_at is not None else offset
    return opened_at
