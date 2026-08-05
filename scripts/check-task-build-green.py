#!/usr/bin/env python3
"""check-task-build-green.py — fail when a tasks.md's dotted-id tasks carry
no **Build:** tag, or a `red` tag whose `merges with Task …` clause names no
partner, an absent partner, or a partner that is itself `red`.

Rule (canonical definition: skills/myflow-contracts/build-green.md — do not
restate it here; a second copy is a Single Source of Truth violation, the
same class of drift check-plan-provenance.py's own docstring warns against).

Scope is a single file per invocation (unlike check-plan-provenance.py's
whole-repo scan): this script takes exactly one `tasks.md` path on argv and
scans only that file. `check-task-build-green.sh` is the thin wrapper that
resolves WHICH files to pass — every non-archived change's tasks.md when
called with no arguments, or one explicit path when called with one.

Exit codes:
  0  clean — every task in the file carries a resolvable **Build:** tag
     (including a file with zero tasks).
  1  violations found — one or more of: a duplicate task id, a task with no
     tag, a `red` tag naming no partner, a named partner that does not exist
     in this file, or a named partner that is itself `red`. Printed one per
     line as `file:line: message`.
  2  invocation error — wrong argument count, or the file cannot be read.

Fenced code blocks (a line matching `^ {0,3}(`{3,}|~{3,})`, toggling fence
state each time it is seen) are opaque to this parser: a `### ...` heading or
`**Build:** ...` line inside a fence opens no task, closes no task's body,
and is never read as a real tag — it is documentation/example text, not
structure. This keeps a `tasks.md` free to show worked examples of the tag
grammar without those examples being mistaken for real tasks.

Parsing model
-------------

A task begins at a line (outside any fence) matching
`^### (DOTTED_ID)(?:\\s|$)`, where `DOTTED_ID` is `\\d+(?:\\.\\d+)*` — a
level-3 heading whose text starts with a dotted id, optionally followed by
more heading text, or by nothing at all (the id runs to end of line). That
id is the task's identity for every violation message and every `merges
with Task <id>` reference. Two or more tasks sharing the same id is itself a
violation (see list item 0 below); lookups against that id resolve against
whichever task was parsed first.

A task's BODY runs from its heading line to the next line (outside any
fence) matching `^#{2,3}(?:\\s|$)` (a level-2 or level-3 heading, whether or
not the level-3 one is itself a task heading) or end of file. Within that
body, the FIRST line (also outside any fence) matching

    ^\\*\\*Build:\\*\\*\\s+(green|red\\s+—\\s+merges with Task ([\\d.,\\s]+))\\s*$

(note: `—` is an em-dash, U+2014, not a hyphen) is the task's tag. A body
with no such line has no tag at all — this includes a line that merely looks
like an attempt at one (`**Build:** yellow`), which is deliberately treated
the same as no tag rather than as a separate parse-error class: the fixed
vocabulary is `green` or `red — merges with Task …`, and anything else is
simply absent, reported the same way an entirely missing line is.

Note the tag grammar's own boundary: `red` alone, with no `— merges with
Task …` clause at all, does not match this regex and is therefore "no tag"
(the missing-tag violation), not the red-without-partner violation below.
The red-without-partner violation fires only when the clause is present
syntactically (`merges with Task` appears) but the digit-id group it
requires (`[\\d.,\\s]+`, one or more characters) captures no digits at all —
e.g. trailing punctuation or whitespace with nothing else after `Task `.

Validation, run over every task in document order:

  0. A task id that duplicates an earlier task's id in the same file →
     `<file>:<this heading's line>: task <id> is defined more than once
     (first at line <first heading's line>)`, reported for every occurrence
     after the first.
  1. Missing tag → `<file>:<heading line>: task <id> has no **Build:** tag`
  2. `red` with a clause naming zero ids →
     `<file>:<tag line>: task <id> is red with no merge partner named`
  3. A named partner id that is not any task's id in this file →
     `<file>:<tag line>: task <id> merges with Task <partner>, which does
     not exist in this plan`
  4. A named partner that exists and is itself tagged `red` →
     `<file>:<tag line>: task <id> merges with Task <partner>, which is
     itself red` — reported once, from the task carrying the offending tag,
     never requiring the mutual reference to also be checked from the
     partner's own tag.
  5. A `red` task whose partner exists and is `green` is not a violation,
     and neither is a `red` task with no incoming reference from anything
     else in the file — "unreferenced" is not itself a violation shape this
     guard checks for.

A partner id named more than once in the same clause (`merges with Task
2.1, 2.1`) is validated once, not once per occurrence — the partner list is
de-duplicated, order preserving, before validation, so a repeated id never
produces duplicate violation lines for the same pair.

Standard library only (see check-plan-provenance.py's module docstring for
why this repository restricts itself to that: no markdown/CommonMark
package is installed anywhere on this machine and this repository has no
dependency management).
"""

from __future__ import annotations

import re
import sys
from dataclasses import dataclass, field
from typing import List, Optional

# DOTTED_ID — the dotted task-id grammar ("1", "1.1", "9.9.2", ...), shared
# by TASK_HEADING_RE and PARTNER_ID_RE so the two never drift apart.
DOTTED_ID = r"\d+(?:\.\d+)*"

# TASK_HEADING_RE — opens a new task. Group 1 is the task's id. The id may
# be followed by more heading text (`### 1.1 Some title`) or nothing at all
# (`### 1.1` at end of line) — both are recognised task headings.
TASK_HEADING_RE = re.compile(rf"^### ({DOTTED_ID})(?:\s|$)")

# FENCE_RE — a fenced code block delimiter: three or more backticks or
# tildes, optionally preceded by up to 3 columns of leading whitespace (per
# CommonMark). Matching this line toggles fence state; a task heading or
# **Build:** tag line inside a fence is example text, not real structure.
FENCE_RE = re.compile(r"^ {0,3}(`{3,}|~{3,})")

# BODY_BOUNDARY_RE — closes the current task's body: any level-2 or level-3
# heading, whether or not the level-3 one is itself a recognised task
# heading (a "### Notes" aside inside a task's own body still ends it).
BODY_BOUNDARY_RE = re.compile(r"^#{2,3}(?:\s|$)")

# BUILD_TAG_RE — the tag vocabulary. `—` is U+2014 (em-dash), not `-`.
# Group "partners" captures the raw digit/dot/comma/whitespace run after
# "Task " for the red form; it is `None` for the green form.
BUILD_TAG_RE = re.compile(
    r"^\*\*Build:\*\*\s+(?:(?P<green>green)"
    r"|red\s+—\s+merges with Task (?P<partners>[\d.,\s]+))\s*$"
)

# PARTNER_ID_RE — extracts every dotted-id token out of a partners clause,
# so "9.9", "2.1, 3.4" and "2.1 3.4" (comma- or whitespace-separated) all
# resolve the same way without this guard hand-rolling a splitter.
PARTNER_ID_RE = re.compile(DOTTED_ID)


@dataclass
class Task:
    """One task found in a tasks.md file."""

    id: str
    heading_line: int
    tag_kind: Optional[str] = None  # "green", "red", or None (no tag)
    tag_line: Optional[int] = None
    partners: List[str] = field(default_factory=list)


def parse_tasks(lines: List[str]) -> List[Task]:
    """Split `lines` (1-indexed by position, 0-indexed in the list) into
    tasks, then resolve each task's **Build:** tag from its own body.
    """
    tasks: List[Task] = []
    current: Optional[Task] = None
    body_start = 0

    def close_body(end_index: int) -> None:
        if current is None:
            return
        in_fence = False
        for offset, body_line in enumerate(lines[body_start:end_index]):
            if FENCE_RE.match(body_line):
                in_fence = not in_fence
                continue
            if in_fence:
                continue
            m = BUILD_TAG_RE.match(body_line)
            if m is None:
                continue
            lineno = body_start + offset + 1
            if m.group("green"):
                current.tag_kind = "green"
                current.tag_line = lineno
            else:
                current.tag_kind = "red"
                current.tag_line = lineno
                current.partners = PARTNER_ID_RE.findall(m.group("partners"))
            break

    in_fence = False
    for index, line in enumerate(lines):
        if FENCE_RE.match(line):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        heading_match = TASK_HEADING_RE.match(line)
        if heading_match:
            close_body(index)
            current = Task(id=heading_match.group(1), heading_line=index + 1)
            tasks.append(current)
            body_start = index + 1
            continue
        if BODY_BOUNDARY_RE.match(line):
            close_body(index)
            current = None

    close_body(len(lines))
    return tasks


def check_tasks(tasks: List[Task], relfile: str) -> List[str]:
    """Validate `tasks` (already parsed, in document order) and return one
    formatted `file:line: message` violation string per finding, in the
    order described in the module docstring.
    """
    by_id = {}
    duplicate_violations: List[str] = []
    for task in tasks:
        first = by_id.get(task.id)
        if first is None:
            by_id[task.id] = task
        else:
            duplicate_violations.append(
                f"{relfile}:{task.heading_line}: task {task.id} is defined "
                f"more than once (first at line {first.heading_line})"
            )
    violations: List[str] = list(duplicate_violations)

    for task in tasks:
        if task.tag_kind is None:
            violations.append(
                f"{relfile}:{task.heading_line}: task {task.id} has no "
                "**Build:** tag"
            )
            continue
        if task.tag_kind != "red":
            continue
        if not task.partners:
            violations.append(
                f"{relfile}:{task.tag_line}: task {task.id} is red with no "
                "merge partner named"
            )
            continue
        seen_partners: List[str] = []
        for partner in task.partners:
            if partner in seen_partners:
                continue
            seen_partners.append(partner)
            partner_task = by_id.get(partner)
            if partner_task is None:
                violations.append(
                    f"{relfile}:{task.tag_line}: task {task.id} merges "
                    f"with Task {partner}, which does not exist in this plan"
                )
            elif partner_task.tag_kind == "red":
                violations.append(
                    f"{relfile}:{task.tag_line}: task {task.id} merges "
                    f"with Task {partner}, which is itself red"
                )

    return violations


def check_file(path: str) -> List[str]:
    """Read `path` and return its violations, formatted with `path` itself
    as the file portion of each `file:line: message` line.
    """
    with open(path, "r", encoding="utf-8") as handle:
        text = handle.read()
    tasks = parse_tasks(text.splitlines())
    return check_tasks(tasks, path)


def main(argv: List[str]) -> int:
    if len(argv) != 2:
        print(
            f"usage: {argv[0]} <path-to-tasks.md>",
            file=sys.stderr,
        )
        return 2

    path = argv[1]
    try:
        violations = check_file(path)
    except OSError as exc:
        print(f"{path}: cannot read file: {exc}", file=sys.stderr)
        return 2

    if not violations:
        return 0
    for violation in violations:
        print(violation)
    return 1


if __name__ == "__main__":
    sys.exit(main(sys.argv))
