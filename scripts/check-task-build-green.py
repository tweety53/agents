#!/usr/bin/env python3
"""check-task-build-green.py — fail when a tasks.md's tasks carry
no **Build:** tag, or a `red` tag whose task has no **Squash-with:** field,
an empty one, one naming an absent partner, or one naming a partner that is
itself `red`.

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
     tag, a `red` task with no **Squash-with:** field at all, a
     **Squash-with:** field naming zero ids, a named partner that does not
     exist in this file, or a named partner that is itself `red`. Printed
     one per line as `file:line: message`.
  2  invocation error — wrong argument count, or the file cannot be read.

Fenced code blocks (a line matching `^ {0,3}(`{3,}|~{3,})`, toggling fence
state each time it is seen) are opaque to this parser: a `- [ ] <id>. ...`
task line, a `**Build:** ...` line, or a `**Squash-with:** ...` line inside a
fence opens no task, closes no task's body, and is never read as a real tag or field —
it is documentation/example text, not structure. This keeps a `tasks.md`
free to show worked examples of the tag grammar without those examples
being mistaken for real tasks.

A fence a task body OPENS and never CLOSES is a different matter, and is
itself a violation (list item 0b below, fix round 11, F22): an unclosed
fence runs to the end of the block, so the `**Build:**` tag and
`**Squash-with:**` field below it are code and are correctly not read —
but reporting the task as untagged names the consequence and hides the
cause. `lib/plan_grammar.py`'s `unclosed_fence` is the detection, shared
with check-task-commit-fields.py, which reports the same defect rather
than blaming the commit for touching a file the swallowed `**Files:**`
declared.

Parsing model
-------------

A task begins at a line (outside any fence) matching
`^- \\[([ x])\\] (TASK_ID)\\. `, where `TASK_ID` is `\\d+` — the spectre
checkbox line `- [ ] <id>. <title>`, whose mark carries whether the task is
DONE and never whether it is a task at all, so `- [x]` opens one exactly as
`- [ ]` does. A task's id is a FLAT integer, spectre's own: a dotted `1.1`
is a "malformed task line" finding there and is no task here either, so a
sub-task is renumbered flat rather than written dotted. The dotted grammar
survives as `lib/plan_grammar.py`'s `DOTTED_ID` for `Squash-with:` partner
ids, where a dotted value now names no task and is reported as a partner
that does not exist. That id is the task's identity for every violation
message and every `Squash-with: Task <id>` reference. Two or more tasks
sharing the same id is itself a violation (see list item 0 below); lookups
against that id resolve against whichever task was parsed first.

A task's BODY runs from its task line to the next task line, to the next
line (outside any fence) matching `^#{2,3}(?:\\s|$)` (a level-2 or level-3
heading), or to end of file. The body is everything BELOW the task line, its
`  - [ ] **Step N: ...**` step checkboxes included: a step is indented two
columns beneath its task, is part of that task's body, and is never a task
of its own — the same distinction spectre's own parser makes, and the indent
is what keeps spectre's malformed-task check off those lines as well. Splitting a file into tasks, and resolving an id
to a task, are `lib/plan_grammar.py`'s `iter_tasks` and `select_task`. Within that body, the FIRST line (also
outside any fence) matching

    ^\\*\\*Build:\\*\\*\\s+(green|red)\\s*$

is the task's tag, and that selection is `lib/plan_grammar.py`'s
`select_build_tag` (fix round 9, F20). A body with no such line has no tag at all — this
includes a line that merely looks like an attempt at one (`**Build:**
yellow`), which is deliberately treated the same as no tag rather than as a
separate parse-error class: the fixed vocabulary is `green` or `red`, and
anything else is simply absent, reported the same way an entirely missing
line is. `Build:` no longer carries any inline suffix — a `red` task's
partner is read from a separate field.

Independently, within that same body, the FIRST line (also outside any
fence) that is a `**Squash-with:**` field whose value gates as `Task
<id>[, <id>...]` is the task's Squash-with field — a non-gating candidate
ahead of it is skipped rather than read as the field. The field's shape,
the gate on its value, and the selection of which line satisfies both, are
all `lib/plan_grammar.py`'s `select_squash_with`, which
check-task-commit-fields.py calls too — one grammar and one selection for
one field, rather than a copy per guard held in step by a comment (fix
round 6, F13; fix round 8, F19). The
field is LINE-SCOPED: its value is what stands on that one line, and a
following prose line is never part of it. A body with no such line has no
Squash-with field at all — a distinct condition, checked only for `red`
tasks, from a Squash-with field that is present but names zero ids. A value
that does not gate (`Task 3 (see step 2)`) is reported as no field, the
condition it has always been read as here.

Validation, run over every task in document order:

  0. A task id that duplicates an earlier task's id in the same file →
     `<file>:<this task line>: task <id> is defined more than once
     (first at line <first task line>)`, reported for every occurrence
     after the first.
  0b. A task body that opens a fence it never closes →
     `<file>:<fence line>: task <id> opens a code fence here that is never
     closed in its body, ...`, reported INSTEAD of checks 1-5 for that
     task, whose fields the fence swallowed.
  1. Missing **Build:** tag → `<file>:<task line>: task <id> has no
     **Build:** tag`
  2. `red` with no **Squash-with:** field at all →
     `<file>:<tag line>: task <id> is red with no **Squash-with:** field`
  3. `red` with a **Squash-with:** field naming zero ids →
     `<file>:<squash-with line>: task <id> is red with no merge partner
     named`
  4. A named partner id that is not any task's id in this file →
     `<file>:<squash-with line>: task <id> merges with Task <partner>,
     which does not exist in this plan`
  5. A named partner that exists and is itself tagged `red` →
     `<file>:<squash-with line>: task <id> merges with Task <partner>,
     which is itself red` — reported once, from the task carrying the
     offending Squash-with field, never requiring the mutual reference to
     also be checked from the partner's own field.
  6. A `red` task whose partner exists and is `green` is not a violation,
     and neither is a `red` task with no incoming reference from anything
     else in the file — "unreferenced" is not itself a violation shape this
     guard checks for.

A partner id named more than once in the same **Squash-with:** field
(`Squash-with: Task 2.1, 2.1`) is validated once, not once per occurrence —
the partner list is de-duplicated, order preserving, before validation, so a
repeated id never produces duplicate violation lines for the same pair.

Standard library only (see check-plan-provenance.py's module docstring for
why this repository restricts itself to that: no markdown/CommonMark
package is installed anywhere on this machine and this repository has no
dependency management).
"""

from __future__ import annotations

import os
import sys
from dataclasses import dataclass, field
from typing import List, Optional

# Every piece of tasks.md grammar this guard reads is defined once, in a
# module check-task-commit-fields.py imports too (fix round 6, F13); see
# that module's docstring for the drifts a per-guard copy produced. `lib/`
# is resolved through this file's REAL path, so the import works when the
# guard is invoked through a skill directory's symlink as well as from the
# repository root. Nothing about the shape of a tasks.md is spelled out
# below this line — splitting the file into tasks, reading a task's
# `**Build:**` tag and reading its `**Squash-with:**` field are all calls
# into that module.
sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "lib"))

from plan_grammar import (
    iter_tasks,
    select_build_tag,
    select_squash_with,
    unclosed_fence,
)


@dataclass
class Task:
    """One task found in a tasks.md file."""

    id: str
    task_line: int
    tag_kind: Optional[str] = None  # "green", "red", or None (no tag)
    tag_line: Optional[int] = None
    squash_line: Optional[int] = None  # None: no **Squash-with:** field
    partners: List[str] = field(default_factory=list)
    # The line of a fence this task's body opens and never closes, or None
    # when it closes every fence it opens (fix round 11, F22).
    fence_line: Optional[int] = None


def parse_tasks(lines: List[str]) -> List[Task]:
    """Split `lines` into tasks, then resolve each task's **Build:** tag and
    **Squash-with:** field from its own body. Which lines are a task, which
    line is its tag and which is its field are all lib/plan_grammar.py's.
    """
    tasks: List[Task] = []
    for found in iter_tasks(lines):
        task = Task(id=found.id, task_line=found.task_line)
        tag = select_build_tag(found.lines)
        if tag is not None:
            task.tag_kind = tag.kind
            task.tag_line = found.body_start + tag.offset + 1
        squash = select_squash_with(found.lines)
        # A value that does not gate is not a Squash-with field at all — a
        # red task carrying one is reported as having no field, not as
        # having an empty one, which is what `partners is None` means here.
        # Case 18 pins both halves of that against this guard.
        if squash is not None and squash.partners is not None:
            task.squash_line = found.body_start + squash.offset + 1
            task.partners = squash.partners
        fence_offset = unclosed_fence(found.lines)
        if fence_offset is not None:
            task.fence_line = found.body_start + fence_offset + 1
        tasks.append(task)
    return tasks


def check_tasks(tasks: List[Task], relfile: str) -> List[str]:
    """Validate `tasks` (already parsed, in document order) and return one
    formatted `file:line: message` violation string per finding, in the
    order described in the module docstring.
    """
    # This loop does not route through lib/plan_grammar.py's `select_task`,
    # and the duplication is deliberate (fix round 11, F23): it needs the
    # FULL index — every id, so the partner lookups below are one dict hit
    # each — and every duplicate OCCURRENCE, so each one after the first is
    # reported. `select_task` answers a single lookup and discards both, and
    # rebuilding either from it costs a pass over the plan per id.
    by_id = {}
    duplicate_violations: List[str] = []
    for task in tasks:
        first = by_id.get(task.id)
        if first is None:
            by_id[task.id] = task
        else:
            duplicate_violations.append(
                f"{relfile}:{task.task_line}: task {task.id} is defined "
                f"more than once (first at line {first.task_line})"
            )
    violations: List[str] = list(duplicate_violations)

    for task in tasks:
        # An unclosed fence REPLACES this task's tag and partner checks (fix
        # round 11, F22). Every line below the fence is code, so the
        # `**Build:**` tag and `**Squash-with:**` field the body may well
        # carry were correctly not read — reporting the task as untagged
        # would name the consequence and hide the cause.
        if task.fence_line is not None:
            violations.append(
                f"{relfile}:{task.fence_line}: task {task.id} opens a code "
                "fence here that is never closed in its body, so every "
                "field below it is inside the fence and was not read"
            )
            continue
        if task.tag_kind is None:
            violations.append(
                f"{relfile}:{task.task_line}: task {task.id} has no "
                "**Build:** tag"
            )
            continue
        if task.tag_kind != "red":
            continue
        if task.squash_line is None:
            violations.append(
                f"{relfile}:{task.tag_line}: task {task.id} is red with no "
                "**Squash-with:** field"
            )
            continue
        if not task.partners:
            violations.append(
                f"{relfile}:{task.squash_line}: task {task.id} is red with "
                "no merge partner named"
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
                    f"{relfile}:{task.squash_line}: task {task.id} merges "
                    f"with Task {partner}, which does not exist in this plan"
                )
            elif partner_task.tag_kind == "red":
                violations.append(
                    f"{relfile}:{task.squash_line}: task {task.id} merges "
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
