#!/usr/bin/env python3
"""check-plan-shape.py — check that a `tasks.md` is SHAPED so the guard
that will later judge its commits (`check-task-commit-fields.py`) can
actually read it. Six findings, F1-F6 (canonical definitions, decisions and
rationale: `spectre/changes/kan-121-run-the-guards-own-parsers-over-tasks-
md-at-plan/design.md` — do not restate them here; a second copy is a
Single Source of Truth violation, the same class of drift
check-plan-provenance.py's own docstring warns against).

This guard IMPORTS the real parsers rather than reimplementing their
grammar — the whole point of the change: `parse_task_fields` and
`collect_task_ids` from `check-task-commit-fields.py`, and `iter_tasks`,
`select_task`, `unclosed_fence` from `lib/plan_grammar.py`. Parser identity
is structural (an import), never asserted (a comment claiming parity).

`check-task-commit-fields.py`'s name carries a hyphen, so it is not
importable by name; it is loaded through
`importlib.util.spec_from_file_location` against this file's own real
directory, and REGISTERED IN `sys.modules` before being executed — without
that registration, the module's `@dataclass` decorator raises
`AttributeError: 'NoneType' object has no attribute '__dict__'` on Python
3.14, because `dataclasses._is_type` looks the defining module up by name
(verified against Python 3.14.6 while this guard was written).

Scope is a single file per invocation, exactly like
check-task-build-green.py: this script takes exactly one `tasks.md` path on
argv and scans only that file. `check-plan-shape.sh` is the thin wrapper
that resolves WHICH files to pass — every non-archived change's tasks.md
when called with no arguments, or one explicit path when called with one.

Exit codes:
  0  clean — every task in the file is shaped so the real parsers read it
     without ambiguity (including a file with zero findings; a file with
     zero TASKS is F5, not clean — see below).
  1  one or more of F1-F6 found. Printed one per line as
     `file:line: message`, naming the task id (F5 excepted: a file with no
     tasks has no task id to name).
  2  invocation error — wrong argument count, or the file cannot be read.

Findings
--------

Every finding is reported as `file:line`, naming the task id, fences
excluded from every scan via `lib/plan_grammar.py`'s `FENCE_RE` — a
field-shaped line inside a worked example is not a declaration.

  F1  A second `**Files:**`, `**Tests:**`, `**Commit:**`, `**Baseline:**`
      or `**Allowed-collateral:**` line in one task body (`F1_FIELD_NAMES`
      below) — naming which occurrence currently wins, since the real
      parser's `fields[current_name] = current_parts` keeps the last.
  F2  No `**Files:**` line in a task body, or one whose backtick token
      list is empty — read directly off `parse_task_fields`'s own `files`
      list, which is `[]` in exactly these two cases (see
      `_extract_backtick_tokens` in check-task-commit-fields.py: a present
      field with prose but no backticks instead falls back to that prose
      as a one-element list, so this finding does not fire there).
  F3  Two distinct shapes, either replacing every check below for that
      task:
        a. a field-shaped line indented past column 0, invisible to
           `FIELD_RE` (which anchors at column 0) — detected by
           `INDENTED_FIELD_RE`, this module's own pattern, derived from
           `FIELD_RE`'s own field-name alternation rather than a second
           literal copy of the field name list;
        b. a task body whose fence never closes — read directly off
           `parse_task_fields`'s own `unclosed_fence_line`, which is
           `lib/plan_grammar.py`'s `unclosed_fence` under the hood. When
           this fires, F1/F2/F4/F6 are skipped for the same task: every
           field below the fence is code, not structure, and reporting
           those as separately empty or duplicated would name the
           consequence and hide the cause (the same convention
           check-task-build-green.py's own fix round 11 established).
  F4  A `**Tests:**` field opening with `none` (`NONE_OPEN_RE`, imported
      from check-task-commit-fields.py — see below) on a task whose
      `**Files:**` names a test-shaped path (`test-*`, `*_test.go`,
      `*.test.*`, `*.spec.ts`) — a contradiction between the opt-out and
      the declared files.
  F5  A `tasks.md` the grammar finds zero tasks in (`collect_task_ids`
      returns `[]`) — a vacuous pass, not a clean file, so it fails rather
      than reporting nothing. Has no task id to name; reported at line 1.
  F6  A `**Tests:**` field that is neither `none`-opening (F4's
      `NONE_OPEN_RE`), nor carries a `Case N` label, nor carries a
      backtick token — it declares nothing, so `check_tests` in
      check-task-commit-fields.py verifies nothing for it. Read as
      `parse_task_fields`'s own `tests` list being empty combined with the
      same `NONE_OPEN_RE` check, since `_parse_test_specs` itself now
      returns `[]` for a `none`-opening field (task 2's fix) as well as for
      case-labels-and-backticks-absent, and the two must not be reported
      twice under two different findings.

A duplicate task id (two task lines sharing one id) is check-task-build-
green.py's own violation, not this guard's: `collect_task_ids` de-
duplicates by first occurrence, and `select_task`/`parse_task_fields`
always resolve to that same first occurrence, so this guard checks each id
once, consistently with what a later commit-fields check will read.

Standard library only (see check-plan-provenance.py's module docstring for
why this repository restricts itself to that).
"""

from __future__ import annotations

import fnmatch
import importlib.util
import os
import re
import sys
from pathlib import Path
from typing import List, NamedTuple, Optional

SCRIPT_DIR = Path(os.path.dirname(os.path.realpath(__file__)))

# --- Load check-task-commit-fields.py -------------------------------------
#
# Named and checked explicitly (not just imported): a Python `import` is
# invisible to check-guard-symlinks.sh's rule 2, which derives a guard's
# required siblings by grepping its SOURCE for `$SCRIPT_DIR/<name>` — the
# .sh wrapper carries that literal reference for rule 2's benefit, and this
# check here is defense in depth for anyone invoking this .py directly.
_CFG_PATH = SCRIPT_DIR / "check-task-commit-fields.py"
if not _CFG_PATH.is_file():
    print(
        f"check-plan-shape.py: required sibling module not found: {_CFG_PATH}",
        file=sys.stderr,
    )
    sys.exit(2)

_cfg_spec = importlib.util.spec_from_file_location("cfg", str(_CFG_PATH))
if _cfg_spec is None or _cfg_spec.loader is None:
    print(f"check-plan-shape.py: cannot load module: {_CFG_PATH}", file=sys.stderr)
    sys.exit(2)
_cfg = importlib.util.module_from_spec(_cfg_spec)
# MUST be registered before exec_module: check-task-commit-fields.py's
# @dataclass decorated classes raise AttributeError on Python 3.14 without
# it (see module docstring above).
sys.modules["cfg"] = _cfg
_cfg_spec.loader.exec_module(_cfg)

parse_task_fields = _cfg.parse_task_fields
collect_task_ids = _cfg.collect_task_ids
FIELD_RE = _cfg.FIELD_RE
BACKTICK_RE = _cfg.BACKTICK_RE
CASE_LABEL_RE = _cfg.CASE_LABEL_RE
NONE_OPEN_RE = _cfg.NONE_OPEN_RE

# --- Load lib/plan_grammar.py ----------------------------------------------
_LIB_DIR = SCRIPT_DIR / "lib"
_GRAMMAR_PATH = _LIB_DIR / "plan_grammar.py"
if not _GRAMMAR_PATH.is_file():
    print(
        f"check-plan-shape.py: required sibling module not found: {_GRAMMAR_PATH}",
        file=sys.stderr,
    )
    sys.exit(2)
sys.path.insert(0, str(_LIB_DIR))

from plan_grammar import FENCE_RE, iter_tasks, select_task, unclosed_fence  # noqa: E402

# F1_FIELD_NAMES — the exact five field names design.md's F1 row names.
# `Regression` and `Build` are deliberately excluded: `Build`'s own
# precedence is lib/plan_grammar.py's select_build_tag's job, and
# `Regression` is not part of this finding's set.
F1_FIELD_NAMES = ("Files", "Tests", "Commit", "Baseline", "Allowed-collateral")

# INDENTED_FIELD_RE — the field-name alternation FIELD_RE anchors at column
# 0, requiring at least one leading whitespace column instead: a
# field-shaped line FIELD_RE never sees because it is indented under its
# own task line, which plan bodies are (F3a). Derived from FIELD_RE's own
# pattern string rather than a second literal copy of the field-name list,
# so the two can never drift apart on which names count as a field.
INDENTED_FIELD_RE = re.compile(r"^[ \t]+" + FIELD_RE.pattern.lstrip("^"))

# NONE_OPEN_RE is imported from check-task-commit-fields.py above (`cfg`),
# not redefined here: it is the same rule design.md's `tests-none-literal`
# decision records for the real parser fix (task 2), now the definition
# `_parse_test_specs` itself checks — one grammar, one regex, per the
# module docstring's "Parser identity is structural" line. This guard still
# applies it independently of what `parse_task_fields` returns, because F4
# and F6 are properties of what the PLAN says, not of what the real parser
# does with it.

# TEST_SHAPED_PATTERNS — the four glob shapes design.md's F4 row names for
# a "test-shaped path".
TEST_SHAPED_PATTERNS = ("test-*", "*_test.go", "*.test.*", "*.spec.ts")


def _is_test_shaped(path: str) -> bool:
    name = path.rsplit("/", 1)[-1]
    return any(fnmatch.fnmatch(name, pattern) for pattern in TEST_SHAPED_PATTERNS)


class _FieldLine(NamedTuple):
    """The first non-fenced line in a task body matching `FIELD_RE` for one
    given field name. `line` is the 1-based file line, used to anchor a
    violation message at the field's opening line; `value` is that line's
    own captured remainder only (never joined with a continuation line). F4
    and F6's NONE_OPEN_RE check does NOT read `value` — a `none` can open on
    a continuation line, so that check reads `fields.tests_value`
    (`check-task-commit-fields.py`'s own continuation-joined string)
    instead; see `_check_task` below."""

    line: int
    value: str


def _first_field_line(
    body: List[str], body_start: int, name: str
) -> Optional[_FieldLine]:
    """The first non-fenced line in `body` whose `FIELD_RE` match names
    `name` — reusing FIELD_RE itself (never a second field-name pattern) to
    find it, fence-aware exactly like every other scan in this module."""
    in_fence = False
    for offset, line in enumerate(body):
        if FENCE_RE.match(line):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        match = FIELD_RE.match(line)
        if match is None or match.group(1) != name:
            continue
        return _FieldLine(line=body_start + offset + 1, value=match.group(2))
    return None


def _check_f1(path: str, task_id: str, body: List[str], body_start: int) -> List[str]:
    # Only the decisive `if name in seen:` line below carries the `# F1`
    # tag the harness's mutation case sed-targets: the fence toggle, the
    # in-fence skip and the field-name match above it are scanning
    # machinery every check in this module shares, not this finding's own
    # decision, and tagging them too would let one sed edit silently
    # disable every check in this file at once instead of only F1's.
    violations: List[str] = []
    seen: dict = {}
    in_fence = False
    for offset, line in enumerate(body):
        if FENCE_RE.match(line):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        match = FIELD_RE.match(line)
        if match is None:
            continue
        name = match.group(1)
        if name not in F1_FIELD_NAMES:
            continue
        file_line = body_start + offset + 1
        if name in seen:  # F1
            violations.append(
                f"{path}:{file_line}: task {task_id} declares a second "
                f"**{name}:** line (first at line {seen[name]}); the "
                "parser keeps the last occurrence and silently discards "
                "this one"
            )
        else:
            seen[name] = file_line
    return violations


def _check_f3a(path: str, task_id: str, body: List[str], body_start: int) -> List[str]:
    # Same discipline as _check_f1 above: only the decisive, POSITIVE
    # `if match is not None:` line carries the `# F3a` tag, so mutating it
    # to `if False:` disables exactly this finding's append and nothing
    # else in the scan around it.
    violations: List[str] = []
    in_fence = False
    for offset, line in enumerate(body):
        if FENCE_RE.match(line):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        match = INDENTED_FIELD_RE.match(line)
        if match is not None:  # F3a
            file_line = body_start + offset + 1
            violations.append(
                f"{path}:{file_line}: task {task_id} carries a "
                f"**{match.group(1)}:** line indented past column 0, "
                "invisible to the field grammar (FIELD_RE anchors at "
                "column 0) and therefore never read as a declaration"
            )
    return violations


def _check_task(
    path: str,
    task_id: str,
    task_line: int,
    body: List[str],
    body_start: int,
    lines: List[str],
) -> List[str]:
    fields = parse_task_fields(lines, task_id)

    if fields.unclosed_fence_line is not None:  # F3b
        return [
            f"{path}:{fields.unclosed_fence_line}: task {task_id} opens a "
            "code fence here that is never closed in its body, so every "
            "field below it is inside the fence and was not read"
        ]

    violations: List[str] = []
    violations.extend(_check_f1(path, task_id, body, body_start))
    violations.extend(_check_f3a(path, task_id, body, body_start))

    if not fields.files:  # F2
        violations.append(
            f"{path}:{task_line}: task {task_id} has no **Files:** line, "
            "or one naming no file at all"
        )

    # NONE_OPEN_RE is applied to fields.tests_value — the SAME
    # continuation-joined string parse_task_fields itself builds and reads
    # (`" ".join(fields.get("Tests", []))`) before handing it to
    # _parse_test_specs — never to the field's own first physical line in
    # isolation: a `none` written on the line after `**Tests:**` (a
    # continuation) is invisible to the latter, which is exactly the
    # duplicate-grammar defect this guard exists to not have (KAN-121
    # panel-fix round, finding 1). `_first_field_line` is still used below,
    # for its line number only, to anchor F4/F6 messages at the field's
    # opening line.
    tests_field = _first_field_line(body, body_start, "Tests")
    tests_opens_none = NONE_OPEN_RE.match(fields.tests_value) is not None

    if tests_opens_none and any(_is_test_shaped(f) for f in fields.files):  # F4
        violations.append(
            f"{path}:{tests_field.line}: task {task_id}'s **Tests:** "
            "field opens with `none` but **Files:** names a test-shaped "
            "path — a contradiction between the opt-out and the declared "
            "files"
        )

    if not tests_opens_none and not fields.tests:  # F6
        file_line = tests_field.line if tests_field is not None else task_line
        violations.append(
            f"{path}:{file_line}: task {task_id}'s **Tests:** field "
            "declares nothing — it opens with neither `none`, a `Case N` "
            "label, nor a backtick-quoted test name, so nothing is "
            "verified"
        )

    return violations


def check_file(path: str) -> List[str]:
    """Read `path` and return its F1-F6 violations, formatted with `path`
    itself as the file portion of each `file:line: message` line."""
    with open(path, "r", encoding="utf-8") as handle:
        text = handle.read()
    lines = text.splitlines()

    ids = collect_task_ids(lines)
    if not ids:  # F5
        return [
            f"{path}:1: this plan defines zero tasks — the grammar finds "
            "no `- [ ] <n>. ` checkbox line, so every check that follows "
            "would be a vacuous pass rather than a real one"
        ]

    violations: List[str] = []
    for task_id in ids:
        found = select_task(lines, task_id)
        assert found is not None  # collect_task_ids only names ids select_task resolves
        violations.extend(
            _check_task(
                path, task_id, found.task_line, found.lines, found.body_start, lines
            )
        )
    return violations


def main(argv: List[str]) -> int:
    if len(argv) != 2:
        print(f"usage: {argv[0]} <path-to-tasks.md>", file=sys.stderr)
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
