#!/usr/bin/env python3
"""check-task-commit-fields.py — check a task's declared `Files:`, `Tests:`
and `Commit:` fields in `tasks.md` against the real commit that closed it.

Rule (canonical definition: openspec/changes/kan-100-myflow-get-rid-of-
staging-use-commits/specs/myflow-task-commit-fields/spec.md — "A runtime
guard checks each field against the real commit", and "Regression and
Baseline checks skip, rather than fail, when unsupported"). `Files:`,
`Tests:` and `Commit:` are checked directly against git (`check_files`,
`check_tests`, `check_commit_subject`). `Regression:` and `Baseline:`
(`check_regression`, `check_baseline`) additionally need the project's
`## test` command, read out of `.myflow/project.md` in the worktree
(`read_single_test_command`) — when that command cannot be targeted at a
single named test or does not report a parseable count, both checks report
skipped-not-verified instead of failing the run.

Scope is one task, in one tasks.md, in one worktree, checked against one
commit (and its parent) — unlike check-task-build-green.py, which scans an
entire tasks.md's tags in one pass. `check-task-commit-fields.sh` is the
thin wrapper that resolves WHICH tasks.md and worktree a caller means.

Field grammar
-------------

A task's fields are read from its body the same way check-task-build-
green.py reads a `**Build:**` tag: the task's body runs from its `###
<DOTTED_ID> ...` heading to the next line matching `^#{2,3}(?:\\s|$)` or end
of file. Within that body, a field starts at a line matching

    ^\\*\\*(Files|Tests|Regression|Baseline|Commit|Allowed-collateral):\\*\\*\\s*(.*)$

and its value continues onto every following non-blank line that does not
itself start a new field, until a blank line, a new field, or the body
boundary — this lets a long field (e.g. `Tests:` naming several cases) wrap
across source lines the way prose in this repository's plans normally does,
without the wrapped continuation being mistaken for a new field's absence.

`Files:` and `Allowed-collateral:` each name one or more backtick-quoted
tokens (`` `path/to/file` ``, comma-separated) — the convention every real
`Files:` line in this plan's own tasks.md already uses. `Commit:` names
exactly one commit subject line, backtick-quoted or bare.

`Tests:` is different: every real `Tests:` field written in this plan's own
tasks.md (including this task's own) is free descriptive prose naming
cases by number — `"Case 1: files subset of declared passes; Case 2:
undeclared file fails; ..."` — not backtick-quoted identifiers. Requiring
backticks there would make this guard false-fail on every real `Tests:`
field in the corpus it has to run against, so `Tests:` is parsed
differently: every `Case <N>` label mentioned anywhere in the field's text
becomes a declared test named "Case <N>", and it is checked by asking
whether `Case <N>` (any case, e.g. as a `# Case <N>:` comment or a `case
<N>:` pass/fail line — the convention this repository's own test harnesses
already use, per `test-check-task-build-green.sh`'s numbered `# Case N:`
comments) appears anywhere in the commit's diff — a loose existence check
by label, not a verbatim match of the surrounding sentence. When a `Tests:`
field names no `Case <N>` label at all, it falls back to backtick-quoted
tokens if present (for a task that names one or two specific test
identifiers rather than numbered cases), and declares no checkable tests at
all when neither shape is present — deliberately vacuous rather than
literal-matching a whole free-prose sentence, which is what produced false
failures against this plan's own `Tests:` fields before this parsing was
added.

Exit codes:
  0  clean — every checked field matches the real commit (a `Regression:`
     or `Baseline:` check that fell back to skipped-not-verified still
     prints its own notice line, but does not affect this exit code).
  1  violations found — one or more of: an undeclared file, a missing
     declared test, a commit subject mismatch, a `Regression:` revert that
     did not make its named test fail, or a `Baseline:` count mismatch.
     Printed one per line as `task <id>: message`.
  2  invocation error — wrong argument count, the tasks.md cannot be read,
     the task id does not exist in it, or git cannot resolve the commit
     range in the given worktree.

Standard library only (see check-plan-provenance.py's module docstring for
why this repository restricts itself to that).
"""

from __future__ import annotations

import contextlib
import fnmatch
import os
import re
import shlex
import subprocess
import sys
from dataclasses import dataclass, field
from typing import Dict, List, NamedTuple, Optional, Pattern, Tuple

DOTTED_ID = r"\d+(?:\.\d+)*"
TASK_HEADING_RE = re.compile(rf"^### ({DOTTED_ID})(?:\s|$)")
BODY_BOUNDARY_RE = re.compile(r"^#{2,3}(?:\s|$)")
# FIELD_RE recognises every field name this plan's grammar defines, not
# only the ones this guard checks — `Build:` and `Squash-with:` must still
# close a preceding field's continuation (see parse_task_fields), even
# though this guard reads neither of their values itself.
FIELD_RE = re.compile(
    r"^\*\*(Files|Tests|Regression|Baseline|Commit|Allowed-collateral|Build|"
    r"Squash-with):\*\*\s*(.*)$"
)
BACKTICK_RE = re.compile(r"`([^`]+)`")
CASE_LABEL_RE = re.compile(r"\bCase\s+(\d+)\b", re.IGNORECASE)
BASELINE_COUNTS_RE = re.compile(r"before=(\d+)\s+after=(\d+)")

# check_commit_scope's grammar: a declared Commit: subject is
# `<type>(<scope>): <rest>` or the Conventional Commits breaking-change form
# `<type>(<scope>)!: <rest>`; a subject with no parenthesised scope before
# the first colon (or `!:`) has no scope at all (SUBJECT_SCOPE_RE then does
# not match). The optional `!` (pass 2, finding A) must sit between the
# closing paren and the colon, or a subject like
# `feat(kan-202-commit-split-and-module-scopes)!: add alpha` bypassed the
# scope check entirely — a real, valid Conventional Commits shape this
# guard has to see. TASK_ID_SCOPE_RE reuses DOTTED_ID: a scope that is
# nothing but digits and dots is a task id, never a module.
SUBJECT_SCOPE_RE = re.compile(r"^[A-Za-z]+\(([^)]*)\)!?:")
TASK_ID_SCOPE_RE = re.compile(rf"^{DOTTED_ID}$")

# The minimal contract this guard requires of a project's `## test` command
# in order to verify Regression:/Baseline: at all — not a real test-runner
# protocol, just what these two checks need out of it. A project whose
# `## test` command does not speak this (this repository's own included,
# whose `## test` section lists several independent shell scripts rather
# than one command that understands either line) never matches these, and
# the corresponding check falls back to skipped-not-verified rather than
# failing — seeing PROJECT_TEST_SECTION_RE.
#
#   `<command> --only <name>`  ->  a line `RESULT <name>: pass` or
#                                   `RESULT <name>: fail` (Regression:)
#   `<command>`                ->  a line `COUNT: <N>` (Baseline:)
TARGETED_RESULT_RE = re.compile(r"^RESULT\s+(\S+):\s*(pass|fail)\s*$", re.MULTILINE)
TOTAL_COUNT_RE = re.compile(r"^COUNT:\s*(\d+)\s*$", re.MULTILINE)

# PROJECT_TEST_SECTION_RE / MARKDOWN_HEADING_RE / FENCE_LINE_RE — read the
# fenced command list under `.myflow/project.md`'s own `## test` heading,
# the same section `.myflow/project.md`'s own prose documents as the source
# of truth for this repository's test commands.
PROJECT_TEST_SECTION_RE = re.compile(r"^## test\s*$", re.IGNORECASE)
MARKDOWN_HEADING_RE = re.compile(r"^#{1,6}\s")
FENCE_LINE_RE = re.compile(r"^```")


class TestSpec(NamedTuple):
    """One declared test extracted from a `Tests:` field: `label` is what a
    violation message names, `pattern` is what must appear in the diff."""

    label: str
    pattern: Pattern[str]


@dataclass
class TaskFields:
    """The subset of one task's declared fields this guard checks."""

    id: str
    files: List[str] = field(default_factory=list)
    tests: List[TestSpec] = field(default_factory=list)
    allowed_collateral: List[str] = field(default_factory=list)
    commit: Optional[str] = None
    baseline: Optional[Tuple[int, int]] = None


class CheckOutcome(NamedTuple):
    """The result of `check_regression`/`check_baseline` when they have
    something to report: `skipped=True` is skipped-not-verified (goes to
    `main`'s notices, never affects the exit code), `skipped=False` is a
    real violation (goes to `main`'s violations, exit 1). Returning `None`
    from either check (not this type at all) means the check passed
    outright, or the task declared nothing checkable — both silent, the
    same convention `check_files`/`check_tests` use for "nothing to
    report"."""

    skipped: bool
    message: str


class TaskNotFoundError(Exception):
    """Raised when the requested task id has no heading in tasks.md."""


def _extract_backtick_tokens(value: str) -> List[str]:
    tokens = BACKTICK_RE.findall(value)
    if tokens:
        return tokens
    stripped = value.strip()
    return [stripped] if stripped else []


def _parse_test_specs(value: str) -> List[TestSpec]:
    """See the module docstring's `Tests:` section for why case labels take
    priority over backtick tokens, and why an untagged prose field yields no
    checkable tests at all rather than one literal-matched sentence."""
    case_numbers: List[str] = []
    for number in CASE_LABEL_RE.findall(value):
        if number not in case_numbers:
            case_numbers.append(number)
    if case_numbers:
        return [
            TestSpec(
                label=f"Case {number}",
                pattern=re.compile(rf"case\s*{re.escape(number)}\b", re.IGNORECASE),
            )
            for number in case_numbers
        ]
    return [
        TestSpec(label=token, pattern=re.compile(re.escape(token)))
        for token in BACKTICK_RE.findall(value)
    ]


def parse_task_fields(lines: List[str], task_id: str) -> TaskFields:
    """Find `task_id`'s heading in `lines` and read its declared fields."""
    body_start: Optional[int] = None
    body_end = len(lines)
    for index, line in enumerate(lines):
        heading_match = TASK_HEADING_RE.match(line)
        if heading_match and heading_match.group(1) == task_id:
            body_start = index + 1
            continue
        if body_start is not None and BODY_BOUNDARY_RE.match(line):
            body_end = index
            break

    if body_start is None:
        raise TaskNotFoundError(f"task {task_id} not found")

    fields: Dict[str, List[str]] = {}
    current_name: Optional[str] = None
    current_parts: List[str] = []
    in_fence = False

    def commit_field() -> None:
        if current_name is not None:
            fields[current_name] = current_parts

    for line in lines[body_start:body_end]:
        if FENCE_LINE_RE.match(line):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        field_match = FIELD_RE.match(line)
        if field_match:
            commit_field()
            current_name = field_match.group(1)
            current_parts = [field_match.group(2)]
            continue
        if line.strip() == "":
            commit_field()
            current_name = None
            current_parts = []
            continue
        if current_name is not None:
            current_parts.append(line.strip())
    commit_field()

    files_value = " ".join(fields.get("Files", []))
    tests_value = " ".join(fields.get("Tests", []))
    collateral_value = " ".join(fields.get("Allowed-collateral", []))
    baseline_value = " ".join(fields.get("Baseline", []))
    commit_parts = fields.get("Commit", [])
    commit_value = " ".join(commit_parts).strip() if commit_parts else None
    if commit_value:
        stripped_backtick = BACKTICK_RE.fullmatch(commit_value)
        if stripped_backtick:
            commit_value = stripped_backtick.group(1)

    baseline_match = BASELINE_COUNTS_RE.search(baseline_value)
    baseline_counts = (
        (int(baseline_match.group(1)), int(baseline_match.group(2)))
        if baseline_match
        else None
    )

    return TaskFields(
        id=task_id,
        files=_extract_backtick_tokens(files_value),
        tests=_parse_test_specs(tests_value),
        allowed_collateral=_extract_backtick_tokens(collateral_value),
        commit=commit_value,
        baseline=baseline_counts,
    )


def run_git(worktree: str, args: List[str]) -> str:
    result = subprocess.run(
        ["git", "-C", worktree] + args,
        capture_output=True,
        text=True,
    )
    if result.returncode != 0:
        raise RuntimeError(
            f"git -C {worktree} {' '.join(args)} failed: {result.stderr.strip()}"
        )
    return result.stdout


def resolve_parent(worktree: str, commit_sha: str) -> str:
    return run_git(worktree, ["rev-parse", f"{commit_sha}^"]).strip()


def check_files(task: TaskFields, changed_files: List[str]) -> List[str]:
    violations = []
    for path in changed_files:
        if path in task.files:
            continue
        if any(fnmatch.fnmatch(path, glob) for glob in task.allowed_collateral):
            continue
        violations.append(
            f"task {task.id}: file {path} is not declared in Files: and not "
            "covered by Allowed-collateral:"
        )
    return violations


def check_tests(task: TaskFields, diff_text: str) -> List[str]:
    violations = []
    for spec in task.tests:
        if not spec.pattern.search(diff_text):
            violations.append(
                f"task {task.id}: declared test {spec.label} not found in "
                "the diff"
            )
    return violations


def check_commit_subject(task: TaskFields, actual_subject: str) -> List[str]:
    if task.commit is None:
        return []
    if task.commit == actual_subject:
        return []
    return [
        f"task {task.id}: commit subject {actual_subject!r} does not match "
        f"declared Commit: {task.commit!r}"
    ]


# _LEADING_KEY_RE: a Jira key is `<letters>-<digits>` — never digits alone —
# at the very start of the change name (per skills/myflow-contracts/
# jira-integration.md's own `[A-Z]{2,10}-\d+` shape and this repository's
# "Change naming" convention, `<key>-<slug>`; matched case-insensitively
# here since check_commit_scope's own comparison is, per finding B).
_LEADING_KEY_RE = re.compile(r"^[A-Za-z]+-\d+")


def _leading_jira_key(change_name: str) -> Optional[str]:
    """The change name's leading Jira key, e.g. `kan-202` from
    `kan-202-commit-split-and-module-scopes`. A change name that does not
    begin with `<letters>-<digits>` at all has no key and this returns
    `None` (pass 2, finding C — the prior implementation returned the name
    up through its first ALL-DIGIT hyphen segment, which both over-matched
    a plain `<word>-<year>`-shaped module name and, worse, could not tell
    that shape apart from a real key at all).

    A leading `<letters>-<digits>` match is itself ambiguous, not a key,
    when the text immediately following it is ANOTHER `<letters>-<digits>`
    pair — e.g. `release-2026-kan-450-cleanup`, where `release-2026` and
    `kan-450` are equally plausible candidates and nothing here can tell
    which one (if either) is the real Jira key; per the same "does not
    begin with one has no key" rule, this returns `None` rather than guess.
    A single unambiguous match, e.g. `kan-202` in
    `kan-202-commit-split-and-module-scopes` (followed by `commit-split`,
    which is not itself a `<letters>-<digits>` pair), is returned as-is.
    """
    match = _LEADING_KEY_RE.match(change_name)
    if not match:
        return None
    key = match.group(0)
    remainder = change_name[len(key):]
    if remainder.startswith("-") and _LEADING_KEY_RE.match(remainder[1:]):
        return None
    return key


def check_commit_scope(task: TaskFields, change_name: str) -> List[str]:
    """Fail when the **declared** `Commit:` field's scope — parsed out of
    `task.commit`, never the real commit's subject, per myflow-commit-
    scope's requirement — equals the change name, the change name's bare
    Jira key, or a dotted/numeric task id. A field with no scope, or any
    other scope, passes: there is no vocabulary of legal module names to
    maintain.

    The change-name and Jira-key comparisons are case-insensitive (pass 2,
    finding B): a declared scope of `feat(KAN-202): ...` against change
    `kan-202-...` is the shape a human is *most* likely to type, since Jira
    keys are conventionally written uppercase, and a case-sensitive compare
    let it straight through. The task-id comparison is deliberately left
    case-sensitive: `TASK_ID_SCOPE_RE` only ever matches a scope that is
    digits and dots (`^\\d+(?:\\.\\d+)*$`), which carries no letters at all,
    so case does not apply to it.
    """
    if task.commit is None:
        return []
    match = SUBJECT_SCOPE_RE.match(task.commit)
    if not match:
        return []
    scope = match.group(1)
    if not scope:
        return []
    scope_lower = scope.lower()
    if scope_lower == change_name.lower():
        return [
            f"task {task.id}: declared Commit: scope {scope!r} names the "
            "change, not a module"
        ]
    jira_key = _leading_jira_key(change_name)
    if jira_key is not None and scope_lower == jira_key.lower():
        return [
            f"task {task.id}: declared Commit: scope {scope!r} names the "
            "change's Jira key, not a module"
        ]
    if TASK_ID_SCOPE_RE.match(scope):
        return [
            f"task {task.id}: declared Commit: scope {scope!r} is a task "
            "id, not a module"
        ]
    return []


def read_single_test_command(worktree: str) -> Optional[str]:
    """Read `<worktree>/.myflow/project.md`'s `## test` section and return
    its one command line, or `None` when the section is absent, empty, or
    names more than one command.

    A single command is what `check_regression`/`check_baseline` need: one
    they can re-invoke with `--only <name>` (Regression:) or with no
    arguments to get a total count (Baseline:). A list of several
    independent commands — this repository's own `## test` section, a list
    of shell scripts with no way to target one named test — can never
    satisfy either need, which is exactly the case the skip-not-verified
    fallback exists for; it is not a parsing failure, so this returns
    `None` rather than raising.
    """
    project_md = os.path.join(worktree, ".myflow", "project.md")
    try:
        with open(project_md, "r", encoding="utf-8") as handle:
            lines = handle.read().splitlines()
    except OSError:
        return None

    commands: List[str] = []
    in_section = False
    in_fence = False
    for line in lines:
        if PROJECT_TEST_SECTION_RE.match(line):
            in_section = True
            in_fence = False
            continue
        if in_section and not in_fence and MARKDOWN_HEADING_RE.match(line):
            break
        if not in_section:
            continue
        if FENCE_LINE_RE.match(line):
            in_fence = not in_fence
            continue
        if in_fence and line.strip():
            commands.append(line.strip())

    return commands[0] if len(commands) == 1 else None


def _run_test_command(worktree: str, command: str, extra_args: List[str]) -> str:
    result = subprocess.run(
        shlex.split(command) + extra_args,
        cwd=worktree,
        capture_output=True,
        text=True,
    )
    return result.stdout


@contextlib.contextmanager
def _commit_reverted(worktree: str, commit_sha: str):
    """Revert `commit_sha` in `worktree`'s working tree (uncommitted, via
    `git revert --no-commit`) for the body of the `with` block, then
    unconditionally restore the pre-revert state with `git reset --hard
    commit_sha` on the way out — the "revert, ..., un-revert" `myflow-task-
    commit-fields` describes for `Regression:`, reused for `Baseline:`'s
    "run at the parent commit" (a task commit's revert lands exactly on its
    parent's tree, since this runs immediately after that commit with
    nothing built on top of it yet).

    A failure in the revert itself (conflict against dirty local state, or
    any other reason `git revert` exits non-zero) still leaves the worktree
    exactly as found: the revert attempt is cleaned up with the same `git
    reset --hard commit_sha` before the original error is re-raised, so
    every exit path — success, a later failure inside the `with` block, or
    the revert itself failing — restores state rather than leaving the
    worktree mid-conflict."""
    try:
        run_git(worktree, ["revert", "--no-commit", "--no-edit", commit_sha])
    except Exception:
        run_git(worktree, ["reset", "--hard", commit_sha])
        raise
    try:
        yield
    finally:
        run_git(worktree, ["reset", "--hard", commit_sha])


def check_regression(task: TaskFields, worktree: str, commit_sha: str) -> Optional[CheckOutcome]:
    """Verify `Regression:` by reverting `commit_sha`, running the task's
    first declared test with `--only`, and confirming it now reports
    `fail` — then always restoring, whatever the outcome. Declares nothing
    checkable when the task names no tests at all (mirrors `check_tests`).
    Falls back to skipped-not-verified when the project's `## test`
    command cannot be targeted at all, or does not report a recognisable
    `RESULT <name>: pass|fail` line for the named test."""
    if not task.tests:
        return None

    command = read_single_test_command(worktree)
    if command is None:
        return CheckOutcome(
            skipped=True,
            message=(
                f"task {task.id}: Regression: skipped, not verified — the "
                "project's `## test` command cannot target a single named "
                "test"
            ),
        )

    label = task.tests[0].label
    with _commit_reverted(worktree, commit_sha):
        output = _run_test_command(worktree, command, ["--only", label])

    match = next(
        (m for m in TARGETED_RESULT_RE.finditer(output) if m.group(1) == label),
        None,
    )
    if match is None:
        return CheckOutcome(
            skipped=True,
            message=(
                f"task {task.id}: Regression: skipped, not verified — the "
                "project's `## test` command does not support targeting a "
                "single named test"
            ),
        )
    if match.group(2) == "fail":
        return None
    return CheckOutcome(
        skipped=False,
        message=(
            f"task {task.id}: Regression: reverting the commit did not "
            f"make {label} fail"
        ),
    )


def check_baseline(task: TaskFields, worktree: str, commit_sha: str) -> Optional[CheckOutcome]:
    """Verify `Baseline:` by running the project's `## test` command at the
    task's own commit and, reverted, at its parent, comparing the two
    `COUNT: <N>` totals against the declared `before=<N> after=<N>`.
    Declares nothing checkable when the task's `Baseline:` field carries no
    parseable `before=/after=` counts (e.g. "not applicable — ..."). Falls
    back to skipped-not-verified when the project's `## test` command
    cannot be targeted at all, or does not report a parseable count at
    either commit."""
    if task.baseline is None:
        return None
    declared_before, declared_after = task.baseline

    command = read_single_test_command(worktree)
    if command is None:
        return CheckOutcome(
            skipped=True,
            message=(
                f"task {task.id}: Baseline: skipped, not verified — the "
                "project's `## test` command cannot be run as a single, "
                "targetable command"
            ),
        )

    after_output = _run_test_command(worktree, command, [])
    after_match = TOTAL_COUNT_RE.search(after_output)

    with _commit_reverted(worktree, commit_sha):
        before_output = _run_test_command(worktree, command, [])
    before_match = TOTAL_COUNT_RE.search(before_output)

    if after_match is None or before_match is None:
        return CheckOutcome(
            skipped=True,
            message=(
                f"task {task.id}: Baseline: skipped, not verified — the "
                "project's `## test` command does not report a parseable "
                "count"
            ),
        )

    actual_before = int(before_match.group(1))
    actual_after = int(after_match.group(1))
    if (actual_before, actual_after) == (declared_before, declared_after):
        return None
    return CheckOutcome(
        skipped=False,
        message=(
            f"task {task.id}: Baseline: declared before={declared_before} "
            f"after={declared_after} but measured before={actual_before} "
            f"after={actual_after}"
        ),
    )


def check_task_commit(
    tasks_md_path: str, task_id: str, worktree: str, commit_sha: str,
    parent_sha: Optional[str] = None,
) -> Tuple[List[str], List[str]]:
    """Returns `(violations, notices)`: `violations` fail the run (exit 1),
    `notices` are skipped-not-verified reports from `Regression:`/
    `Baseline:` that are printed but never affect the exit code, per
    `myflow-task-commit-fields`'s requirement **Regression and Baseline
    checks skip, rather than fail, when unsupported**."""
    with open(tasks_md_path, "r", encoding="utf-8") as handle:
        lines = handle.read().splitlines()
    task = parse_task_fields(lines, task_id)
    change_name = os.path.basename(os.path.dirname(os.path.abspath(tasks_md_path)))

    resolved_parent = parent_sha or resolve_parent(worktree, commit_sha)
    changed_files = [
        p
        for p in run_git(
            worktree, ["diff", "--name-only", f"{resolved_parent}..{commit_sha}"]
        ).splitlines()
        if p
    ]
    diff_text = run_git(worktree, ["diff", f"{resolved_parent}..{commit_sha}"])
    actual_subject = run_git(
        worktree, ["log", "-1", "--format=%s", commit_sha]
    ).strip()

    violations: List[str] = []
    notices: List[str] = []
    violations += check_files(task, changed_files)
    violations += check_tests(task, diff_text)
    violations += check_commit_subject(task, actual_subject)
    violations += check_commit_scope(task, change_name)

    for outcome in (
        check_regression(task, worktree, commit_sha),
        check_baseline(task, worktree, commit_sha),
    ):
        if outcome is None:
            continue
        (notices if outcome.skipped else violations).append(outcome.message)

    return violations, notices


def main(argv: List[str]) -> int:
    if len(argv) not in (5, 6):
        print(
            f"usage: {argv[0]} <tasks.md> <task-id> <worktree> <commit-sha> "
            "[parent-sha]",
            file=sys.stderr,
        )
        return 2

    tasks_md_path, task_id, worktree, commit_sha = argv[1:5]
    parent_sha = argv[5] if len(argv) == 6 else None

    try:
        violations, notices = check_task_commit(
            tasks_md_path, task_id, worktree, commit_sha, parent_sha
        )
    except (OSError, TaskNotFoundError, RuntimeError) as exc:
        print(f"{argv[0]}: {exc}", file=sys.stderr)
        return 2

    for notice in notices:
        print(notice)
    if not violations:
        return 0
    for violation in violations:
        print(violation)
    return 1


if __name__ == "__main__":
    sys.exit(main(sys.argv))
