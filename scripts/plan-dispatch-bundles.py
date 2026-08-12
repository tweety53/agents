#!/usr/bin/env python3
"""plan-dispatch-bundles.py — group a tasks.md's unchecked tasks into
dispatch bundles by the overlap of their declared **Files:** paths.

Rule (canonical definition: openspec/changes/kan-109-optimize-myflow-agent-
token-and-time-cost/specs/myflow-dispatch-economy/spec.md, "Requirement:
Implementer dispatches are bundled by declared file overlap" — do not
restate it here, the same Single Source of Truth discipline check-task-
build-green.py's own docstring already states). Two unchecked tasks join
the same bundle when they declare any common path under **Files:**; the
relation is transitive. An **Allowed-collateral:** glob is read but
deliberately never joins two tasks — it names paths a legitimate sweep may
also touch, not paths the task owns.

Scope is a single file per invocation, matching check-task-build-green.py's
own scope. `plan-dispatch-bundles.sh` is the thin wrapper that resolves
WHICH files to pass — every non-archived change's tasks.md when called with
no arguments, or one explicit path when called with one.

Exit codes:
  0  bundles computed — printed one per line on stdout, ordered by each
     bundle's lowest task id, as `bundle <k>: <ids>` with the ids in plan
     (document) order. A file with zero unchecked tasks prints nothing and
     still exits 0.
  1  one or more unchecked tasks carry no **Files:** field at all — printed
     one per line as `<path>:<heading line>: task <id> has no **Files:**
     field`, naming every such task rather than stopping at the first.
  2  invocation error — wrong argument count, the file cannot be read, or
     it cannot be decoded as text.

Parsing model
-------------

A task begins at a line (outside any fenced code block) matching `^###
(DOTTED_ID)(?:\\s|$)`, the same TASK_HEADING_RE grammar check-task-build-
green.py uses. Its body runs to the next line matching `^#{2,3}(?:\\s|$)`
or end of file, the same BODY_BOUNDARY_RE. Fenced code blocks (three or
more backticks or tildes, toggling fence state each time one is seen) are
opaque to this parser exactly as they are to check-task-build-green.py: a
heading, a checkbox, or a `**Files:**` line inside a fence is worked
example text, not real structure.

Within a task's body:

  - a line matching `^- \\[([ xX])\\]` is one of that task's step
    checkboxes. A task is UNCHECKED when it has at least one such line
    whose mark is exactly a space (`- [ ]`); a task with zero such lines,
    or whose checkboxes are all `[x]`/`[X]`, takes no part in any bundle —
    it is simply not read for its **Files:** field, and never reported as
    missing one.
  - the first line matching `^\\*\\*Files:\\*\\*` opens the Files field. Its
    entries are every immediately following line matching `^- ` (a bullet),
    read until the first line that is not such a bullet (a blank line, a
    new `**Field:**` line, or the body boundary). Each entry may name one or
    more comma-separated backtick-quoted paths, the same convention check-
    task-commit-fields.py's own docstring states for `Files:` and `Allowed-
    collateral:` — strip the leading `- `, strip a leading `Create:`,
    `Modify:` or `Delete:` prefix if present, then take every backtick-
    quoted substring the entry has, or the remaining trimmed text as a
    single path if it has none.
  - `**Allowed-collateral:**` is a distinct bold field. It is read (so a
    Files field's bullet run correctly stops before it) but its own
    entries never contribute a path to the union-find join — this is the
    spec's own "Allowed collateral does not join bundles" scenario.

Path comparison is on the literal extracted text: no filesystem resolution,
no globbing, no normalization beyond the stripping described above. Two
tasks agree on a path only when they wrote the identical remaining text.

Standard library only (see check-plan-provenance.py's module docstring for
why this repository restricts itself to that).
"""

from __future__ import annotations

import re
import sys
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Tuple

# DOTTED_ID / TASK_HEADING_RE / FENCE_RE / BODY_BOUNDARY_RE mirror check-
# task-build-green.py's own constants exactly, so the two guards never read
# the same tasks.md's structure two different ways.
DOTTED_ID = r"\d+(?:\.\d+)*"
TASK_HEADING_RE = re.compile(rf"^### ({DOTTED_ID})(?:\s|$)")
FENCE_RE = re.compile(r"^ {0,3}(`{3,}|~{3,})")
BODY_BOUNDARY_RE = re.compile(r"^#{2,3}(?:\s|$)")

# CHECKBOX_RE — a task step's checkbox. Group "mark" is the character
# between the brackets: a literal space means the step is unchecked.
CHECKBOX_RE = re.compile(r"^- \[(?P<mark>[ xX])\]")

# FILES_FIELD_RE / COLLATERAL_FIELD_RE / ANY_FIELD_RE — the bold field
# lines that open, and that stop, a **Files:** bullet run. ANY_FIELD_RE
# recognises every bold field name this plan grammar defines (matching
# check-task-commit-fields.py's own FIELD_RE vocabulary), so a **Tests:**
# or **Commit:** line correctly ends a Files run too, not only
# **Allowed-collateral:**.
FILES_FIELD_RE = re.compile(r"^\*\*Files:\*\*")
ANY_FIELD_RE = re.compile(
    r"^\*\*(Files|Tests|Regression|Baseline|Commit|Allowed-collateral|"
    r"Build|Squash-with):\*\*"
)

# BULLET_RE — a Files field entry line. CHECKBOX_RE describes a strict
# subset of what this pattern would otherwise accept (both start with
# `- `), so a checkbox line immediately following a **Files:** bullet run,
# with no blank line between them, would be consumed as a bogus file-path
# entry instead of falling through to the checkbox check below. The
# negative lookahead excludes exactly the `- [ ]` / `- [x]` / `- [X]`
# prefix CHECKBOX_RE matches, so the two patterns partition the line space
# instead of overlapping.
BULLET_RE = re.compile(r"^- (?!\[[ xX]\])\s*(.*)$")

# ENTRY_PREFIX_RE — the recognised Create:/Modify:/Delete: prefix on one
# entry, stripped before the path is extracted.
ENTRY_PREFIX_RE = re.compile(r"^(?:Create|Modify|Delete):\s*")

# BACKTICK_RE — every backtick-quoted substring in an entry. Same regex
# shape as check-task-commit-fields.py's own BACKTICK_RE, and read the same
# way: via findall, so a comma-separated multi-path bullet yields every
# token rather than only the first.
BACKTICK_RE = re.compile(r"`([^`]+)`")


@dataclass
class Task:
    """One task found in a tasks.md file."""

    id: str
    heading_line: int
    unchecked: bool = False
    files_present: bool = False
    files: List[str] = field(default_factory=list)


def _extract_paths(entry_text: str) -> List[str]:
    """Turn one **Files:** bullet entry's text (already stripped of its
    leading `- `) into every path it declares. An entry may name one or
    more comma-separated backtick-quoted paths; every such token is a
    separate declared path, matching check-task-commit-fields.py's own
    BACKTICK_RE.findall precedent for the same field convention."""
    stripped = ENTRY_PREFIX_RE.sub("", entry_text.strip(), count=1).strip()
    tokens = BACKTICK_RE.findall(stripped)
    if tokens:
        return [token.strip() for token in tokens]
    return [stripped]


def parse_tasks(lines: List[str]) -> List[Task]:
    """Split `lines` (0-indexed in the list, 1-indexed as line numbers) into
    tasks, then resolve each task's unchecked status and **Files:** field
    from its own body."""
    tasks: List[Task] = []
    current: Optional[Task] = None
    body_start = 0

    def close_body(end_index: int) -> None:
        if current is None:
            return
        in_fence = False
        in_files = False
        offset = 0
        body_lines = lines[body_start:end_index]
        while offset < len(body_lines):
            body_line = body_lines[offset]
            if FENCE_RE.match(body_line):
                in_fence = not in_fence
                offset += 1
                continue
            if in_fence:
                offset += 1
                continue
            if in_files:
                bullet_match = BULLET_RE.match(body_line)
                if bullet_match:
                    current.files.extend(_extract_paths(bullet_match.group(1)))
                    offset += 1
                    continue
                in_files = False
                # Fall through: this line still needs its own checks below
                # (it may itself open a new field or be a checkbox).
            if CHECKBOX_RE.match(body_line):
                mark = CHECKBOX_RE.match(body_line).group("mark")
                if mark == " ":
                    current.unchecked = True
                offset += 1
                continue
            if FILES_FIELD_RE.match(body_line):
                current.files_present = True
                in_files = True
                offset += 1
                continue
            if ANY_FIELD_RE.match(body_line):
                offset += 1
                continue
            offset += 1

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


def _id_key(task_id: str) -> Tuple[int, ...]:
    """Numeric sort key for a dotted task id, so "1.10" sorts after "1.2"
    rather than before it."""
    return tuple(int(part) for part in task_id.split("."))


class UnionFind:
    """Standard union-find with path compression, over a fixed set of
    task ids."""

    def __init__(self, ids: List[str]) -> None:
        self._parent: Dict[str, str] = {i: i for i in ids}

    def find(self, x: str) -> str:
        root = x
        while self._parent[root] != root:
            root = self._parent[root]
        while self._parent[x] != root:
            self._parent[x], x = root, self._parent[x]
        return root

    def union(self, a: str, b: str) -> None:
        ra, rb = self.find(a), self.find(b)
        if ra != rb:
            self._parent[rb] = ra


def compute_bundles(tasks: List[Task]) -> Tuple[List[List[Task]], List[str]]:
    """Return (bundles, missing_files_violations) for `tasks` (already
    parsed, in document order). `bundles` is empty when
    `missing_files_violations` is non-empty — the caller must not use one
    without checking the other, matching exit 1 taking priority over exit
    0's bundle output."""
    unchecked = [t for t in tasks if t.unchecked]

    missing = [t for t in unchecked if not t.files_present]
    if missing:
        return [], [
            f"{t.heading_line}: task {t.id} has no **Files:** field"
            for t in missing
        ]

    uf = UnionFind([t.id for t in unchecked])
    by_path: Dict[str, str] = {}
    for task in unchecked:
        for path in task.files:
            if path in by_path:
                uf.union(by_path[path], task.id)
            else:
                by_path[path] = task.id

    groups: Dict[str, List[Task]] = {}
    for task in unchecked:
        root = uf.find(task.id)
        groups.setdefault(root, []).append(task)  # plan order preserved

    ordered = sorted(
        groups.values(), key=lambda members: min(_id_key(t.id) for t in members)
    )
    return ordered, []


def check_file(path: str) -> Tuple[List[List[Task]], List[str]]:
    """Read `path` and return (bundles, violations) as compute_bundles
    does, with `violations` already formatted with `path` as the file
    portion of each `file:line: message` line."""
    with open(path, "r", encoding="utf-8") as handle:
        text = handle.read()
    tasks = parse_tasks(text.splitlines())
    bundles, raw_violations = compute_bundles(tasks)
    violations = [f"{path}:{v}" for v in raw_violations]
    return bundles, violations


def main(argv: List[str]) -> int:
    if len(argv) != 2:
        print(f"usage: {argv[0]} <path-to-tasks.md>", file=sys.stderr)
        return 2

    path = argv[1]
    try:
        bundles, violations = check_file(path)
    except (OSError, UnicodeDecodeError) as exc:
        print(f"{path}: cannot read file: {exc}", file=sys.stderr)
        return 2

    if violations:
        for violation in violations:
            print(violation, file=sys.stderr)
        return 1

    for index, members in enumerate(bundles, start=1):
        ids = " ".join(t.id for t in members)
        print(f"bundle {index}: {ids}")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
