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
entire tasks.md's tags in one pass. The one exception is a folded pair,
below, whose partner is looked up in the same tasks.md.
`check-task-commit-fields.sh` is the thin wrapper that resolves WHICH
tasks.md and worktree a caller means.

Folded red tasks
----------------

A task tagged `**Build:** red` carries `**Squash-with:** Task <N>`, naming
one or more partners in the same plan, and its commit is folded into that
one unit's commit before review (see skills/myflow-contracts/build-
green.md). After the fold the red task has NO commit of its own: its
declared `Commit:` subject exists nowhere, and one commit carries every
folded task's files. Checking it against a commit of its own is therefore
impossible, which is why `resolve_folded_task` resolves the whole fold
first — and the fold is every red task joined by a shared partner, not one
red task's own `Squash-with:` list, since two reds naming the same partner
both fold into that partner's one commit:

  * invoked for the red task -> the PARTNERS' agreed `Commit:` subject,
    with the UNION of every folded task's `Files:`/`Allowed-collateral:`
    sets;
  * invoked for a partner -> the fold's agreed subject, which is its own
    where it declared one, with that same union — the red task's files AND
    every SIBLING partner's — so either id reaches one verdict against the
    folded commit;
  * a `Squash-with:` value that is not `Task <id>[, <id>...]` -> a
    violation naming the value, reported from EVERY id in the plan, and
    nothing else is checked (fix round 6, F16). The whole field is gated
    before any id is extracted — by lib/plan_grammar.py, the one grammar
    check-task-build-green.py gates it with too — because a bare id pattern
    run over free text turns any digit in it into a partner and silently
    widens the fold's file union (fix round 5, F9). Reported rather than
    treated as no partners: the two guards read one field, build-green
    already fails a red task whose `Squash-with:` it cannot parse, and
    demoting the task to the ordinary path would check it against a
    `Commit:` subject the fold deleted. Plan-wide rather than from the red
    alone, because the guard cannot know which commit that task folded into
    and therefore cannot vouch for any fold in the plan; reported only from
    the red, the defect was invisible from the very task its free text
    named, which was told instead that the folded commit's other file was
    undeclared collateral — true of the commit, and about the wrong field;
  * a `Squash-with:` naming a task absent from the plan, or a partner that
    is itself red -> a violation naming it, and nothing else is checked.
    The partner set is re-validated from a partner's id too, so an invalid
    fold cannot pass through one. Such an edge joins NOTHING (fix round 6,
    F15): the named task, and any valid fold it belongs to, stay out of
    this red's membership set, so one red's broken reference cannot fail a
    separate and entirely correct folded commit from every id in it. This
    differs from the ungated value above because here both the named
    partner and the reason the edge is illegal are known, so the guard
    knows the fold does not exist; there, nothing is known;
  * partners declaring DIFFERENT `Commit:` subjects -> a violation, whether
    they are one red's partners or two joined folds'. One unit folds into
    one commit, so one subject; disagreement is a plan defect, not a
    mismatch to blame on the commit. A partner declaring NO `Commit:` field
    contributes no subject rather than a distinct one, so it never
    manufactures that disagreement;
  * NO partner declaring a `Commit:` subject anywhere in the fold -> a
    violation. The surviving commit's subject would otherwise be checked
    against nothing and any subject would pass. The question is asked of the
    whole combined fold, not of one red's own partner list: where any red in
    the group has a partner declaring a subject, that is the fold's subject
    and a sibling red whose partners declared none is no defect. This holds
    inside a fold only: `Commit:` stays optional for an ordinary task;
  * every other task -> unchanged. The union widens a folded pair, never an
    ordinary task.

`Tests:` and the declared-scope check are the task's own either way: both
are about what THIS task declared, not about which commit survived.

Field grammar
-------------

A task's body runs from its `- [ ] <DOTTED_ID>. ...` task line to the next
task line, to the next line matching `^#{2,3}(?:\\s|$)`, or to end of file,
fenced examples excluded and a duplicated id resolving to its FIRST task
line — all of which is lib/plan_grammar.py's `select_task`, not a scan in
this file (fix round 9). The task line is the spectre checkbox line, whose
mark carries whether the task is DONE; the `- [ ] **Step N: ...**` step
checkboxes below it are part of the body and are never tasks of their own.

A body that opens a fence and never closes it is a PLAN DEFECT, reported on
its own by `check_task_commit` and replacing every other check for that
task (fix round 11, F22). An unclosed fence runs to the end of the block,
so the fields below it are code — correctly unread here, and shown as code
by any renderer — but the guard used to report only the consequence: an
empty `Files:` set, and the commit blamed for touching the very file the
task declared. `lib/plan_grammar.py`'s `unclosed_fence` is the detection,
shared so check-task-build-green.py names the same defect instead of
reporting the task as untagged.
The dotted-id grammar, the whole of `Squash-with:` and the whole of
`**Build:**` come from that module too, which check-task-build-green.py
imports as well — one definition per piece of grammar, rather than a copy
in each guard held in step by a comment asserting parity (fix round 6, F13;
that comment had been wrong once already, and the false claim survived four
review rounds).

`**Build:**` is read from the body by that module's `select_build_tag`: the
first line that is `**Build:** green` or `**Build:** red` in full. It used
to be read through FIELD_RE's alternation below, which joined continuation
lines onto its value and let a later `**Build:**` line overwrite an earlier
one — so `**Build:** red` followed by `**Build:** green` was green here and
red in the other guard, and `**Build:** red` followed by an unblanked prose
line was no tag at all here and red there. Either way one guard resolved the
fold while this one checked the folded commit against a `Commit:` subject
the fold had deleted (fix round 9, F20).

Within that body, a field starts at a line matching

    ^\\*\\*(Files|Tests|Regression|Baseline|Commit|Allowed-collateral|Build):\\*\\*\\s*(.*)$

and its value continues onto every following non-blank line that does not
itself start a new field, until a blank line, a new field, or the body
boundary — this lets a long field (e.g. `Tests:` naming several cases) wrap
across source lines the way prose in this repository's plans normally does,
without the wrapped continuation being mistaken for a new field's absence.

`**Squash-with:**` is the ONE exception, and is LINE-SCOPED: its value is
what stands on its own line, and a following continuation line is not part
of it. Its grammar is closed and short, so it never needs to wrap; and
under the joined reading an unrelated prose line placed under it with no
blank line between silently changed what the field meant — accepted by the
other guard, which gates one physical line, and rejected here (fix round 6,
F14). WHICH line in the body is the field is lib/plan_grammar.py's
`select_squash_with`, called by both guards: the first candidate whose
value gates, so a non-gating line ahead of it is skipped rather than read
as the field (fix round 8, F19). A `Squash-with:`-shaped line still
terminates a preceding field's continuation wherever it stands, which is
all this guard's own field loop reads it for. See lib/plan_grammar.py's
docstring for the rule and its reasoning.

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
     declared test, a commit subject mismatch, a `Squash-with:` value that
     does not gate as `Task <id>[, <id>...]`, an unresolvable
     `Squash-with:` partner, a set of partners — or two folds joined by a
     shared partner — disagreeing about the folded commit's subject, or a
     fold declaring none at all, a `Regression:` revert that
     did not make its named test fail, or a `Baseline:` count mismatch. Printed one per line
     as `task <id>: message`.
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
from dataclasses import dataclass, field, replace
from typing import Dict, List, NamedTuple, Optional, Pattern, Tuple

# The tasks.md grammar this guard shares with check-task-build-green.py is
# defined once, in a module both import (fix round 6, F13); see that
# module's docstring for the two drifts a per-guard copy produced, and for
# why `Squash-with:` is line-scoped. `lib/` is resolved through this file's
# REAL path, so the import works when the guard is invoked through
# skills/myflow-do/scripts/ or skills/myflow-fast/scripts/ as well as from
# the repository root.
sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "lib"))

from plan_grammar import (
    DOTTED_ID,
    FENCE_RE,
    SQUASH_WITH_FIELD_RE,
    iter_tasks,
    select_build_tag,
    select_squash_with,
    select_task,
    unclosed_fence,
)

# FIELD_RE recognises the fields this guard reads out of a task body. Two
# names in it carry no value here: `Squash-with:` and `Build:` are matched
# only so that such a line CLOSES a preceding field's continuation (see
# parse_task_fields). What either of them SAYS is lib/plan_grammar.py's
# `select_squash_with` / `select_build_tag`, so that no loop in this file
# decides it (fix round 8, F19; fix round 9, F20). `Build:` matters here at
# all because a `Build: red` task's commit is folded into its
# `Squash-with:` partner's before review, so the fold has to be resolved
# before any field can be checked (see resolve_folded_task).
FIELD_RE = re.compile(
    r"^\*\*(Files|Tests|Regression|Baseline|Commit|Allowed-collateral|Build)"
    r":\*\*\s*(.*)$"
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
# of truth for this repository's test commands. FENCE_LINE_RE is scoped to
# that file: it gates the whole physical line on three backticks, which is
# all `## test`'s own block has ever used. A task body's fences are
# lib/plan_grammar.py's FENCE_RE instead, everywhere this guard reads one.
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
    build: Optional[str] = None  # "green", "red", or None (no tag)
    # The raw `Squash-with:` value as written, or None when the field is
    # absent. Kept so a red task whose field is present but yields no partner
    # id can be reported as the plan defect it is (see resolve_folded_task),
    # rather than silently demoted to the ordinary single-commit path.
    squash_value: Optional[str] = None
    squash_partners: List[str] = field(default_factory=list)
    # The 1-based tasks.md line of a fence this task's body opens and never
    # closes, or None when it closes every fence it opens. Set so that the
    # swallowed fields below it are reported as the plan defect they are,
    # rather than as an empty declaration the commit is then blamed for
    # exceeding (fix round 11, F22).
    unclosed_fence_line: Optional[int] = None


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
    """Raised when the requested task id has no task line in tasks.md."""


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
    """Find `task_id`'s task line in `lines` and read its declared fields."""
    found = select_task(lines, task_id)
    if found is None:
        raise TaskNotFoundError(f"task {task_id} not found")
    body = found.lines

    fields: Dict[str, List[str]] = {}
    current_name: Optional[str] = None
    current_parts: List[str] = []
    in_fence = False

    def commit_field() -> None:
        if current_name is not None:
            fields[current_name] = current_parts

    for line in body:
        if FENCE_RE.match(line):
            # lib/plan_grammar.py's fence rule, not a local one: CommonMark's
            # backticks-or-tildes with up to three columns of indent. Gating
            # this loop on a bare "^```" instead let a `~~~` or an indented
            # example block stay invisible HERE while every selector this
            # guard shares with check-task-build-green.sh saw it — so a
            # `**Files:**` written inside a worked example overwrote the
            # task's real declaration, and the guard then reported the
            # declared file as undeclared collateral and passed the example's
            # (fix round 10, F21).
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        if SQUASH_WITH_FIELD_RE.match(line):
            # A `Squash-with:`-shaped line terminates a preceding field's
            # continuation — the only thing this loop reads it for. WHICH
            # such line is the field is not decided here at all: it is
            # select_squash_with's answer, below, so that no loop in this
            # file can select a different line from the one
            # check-task-build-green.py selects (fix round 8, F19).
            commit_field()
            current_name = None
            current_parts = []
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

    # The tag is the first line-gated `**Build:** green|red` in the body,
    # per lib/plan_grammar.py's select_build_tag — never a value assembled
    # from FIELD_RE's capture above, which joined continuation lines onto it
    # and let a later tag line overwrite an earlier one (fix round 9, F20).
    tag = select_build_tag(body)

    # select_squash_with returns the first candidate whose value GATES, or —
    # when none gates — the first candidate with `partners=None`. Either of
    # the latter two shapes leaves this task with no partner, and
    # resolve_folded_task reports it as the plan defect it is rather than
    # demoting the task to the ordinary single-commit path.
    squash = select_squash_with(body)
    squash_value = None if squash is None else squash.value
    squash_partners = [] if squash is None else (squash.partners or [])

    fence_offset = unclosed_fence(body)
    fence_line = (
        None if fence_offset is None else found.body_start + fence_offset + 1
    )

    return TaskFields(
        id=task_id,
        files=_extract_backtick_tokens(files_value),
        tests=_parse_test_specs(tests_value),
        allowed_collateral=_extract_backtick_tokens(collateral_value),
        commit=commit_value,
        baseline=baseline_counts,
        build=tag.kind if tag is not None else None,
        squash_value=squash_value,
        squash_partners=squash_partners,
        unclosed_fence_line=fence_line,
    )


def collect_task_ids(lines: List[str]) -> List[str]:
    """Every task id with a task line in `lines`, in document order — the plan
    resolve_folded_task searches for a red task naming a given partner."""
    ids: List[str] = []
    for task in iter_tasks(lines):
        if task.id not in ids:
            ids.append(task.id)
    return ids


def _fold_group(
    lines: List[str], red_task: TaskFields
) -> Tuple[List[str], List[str], Optional[str], List[str]]:
    """Resolve every partner a `Build: red` task's `Squash-with:` names —
    "one or more other tasks in the same plan", per **The build-green tag**
    (`skills/myflow-contracts/build-green.md`). Returns `(partner files,
    partner allowed-collateral, the folded commit's expected subject,
    violations)`.

    The whole set is ONE unit folding into ONE commit ("whose commit this
    task's commit folds into" — singular, over a plural set). So every named
    partner has to declare the SAME `Commit:` subject: two partners naming
    different subjects cannot both describe the one surviving commit, and
    that disagreement is a plan defect reported here rather than a subject
    mismatch blamed on the commit. Taking the first-listed partner's subject
    and checking the commit against that would instead report a false
    mismatch whenever the real subject is a later-listed partner's.

    A partner declaring NO `Commit:` field contributes no subject (fix round
    2, F5/F6). It is skipped rather than counted, so it cannot manufacture a
    disagreement with a partner that did declare one. Whether the fold ends
    up with no subject at all is NOT decided here (fix round 4, item 2):
    this red's partners declaring none is only half the question when
    another red joined to the same commit supplies one, so
    `resolve_folded_task` asks it of the combined group and this function
    simply returns `None` for the subject.

    Whether a partner is missing, itself red, or disagrees about the
    subject, the violation names the RED task — that is where the defective
    field is, whichever id the guard was invoked for.

    Scope is ONE red task's own `Squash-with:` list. Where two red tasks
    share a partner their folds are one combined unit, and it is
    `resolve_folded_task` that walks every red in that unit and reconciles
    what each of them returns here (fix round 3, F8)."""
    files: List[str] = []
    collateral: List[str] = []
    subjects: List[str] = []
    violations: List[str] = []
    # partner_error: this fold already has a defect in the partner SET
    # itself — a partner that does not exist, or one that is itself red
    # (fix round 3, F7). The disagreement check below is about the subjects
    # the partners declared, which is only a meaningful question once the
    # set resolved; and a fold with an unresolvable partner reporting a
    # second violation would name one defect twice. The flag states that
    # precondition instead of inferring it from `violations` being empty,
    # which was true only because these are the sole violations that can
    # precede it. Case 47 pins it: an unresolvable partner alongside two
    # disagreeing resolvable ones reports the missing partner and nothing
    # else.
    partner_error = False
    for partner_id in red_task.squash_partners:
        try:
            partner = parse_task_fields(lines, partner_id)
        except TaskNotFoundError:
            violations.append(
                f"task {red_task.id}: Squash-with: names Task {partner_id}, "
                "which does not exist in this plan"
            )
            partner_error = True
            continue
        if partner.build == "red":
            violations.append(
                f"task {red_task.id}: Squash-with: names Task {partner_id}, "
                "which is itself red"
            )
            partner_error = True
            continue
        files += partner.files
        collateral += partner.allowed_collateral
        if partner.commit is not None and partner.commit not in subjects:
            subjects.append(partner.commit)
    if not partner_error and len(subjects) > 1:
        violations.append(
            f"task {red_task.id}: Squash-with: names partners declaring "
            "different Commit: subjects "
            + ", ".join(repr(subject) for subject in subjects)
            + " — one folded unit is one commit, so one subject"
        )
    subject = subjects[0] if len(subjects) == 1 else None
    return files, collateral, subject, violations


def _widen(
    task: TaskFields,
    extra_files: List[str],
    extra_collateral: List[str],
    commit: Optional[str],
) -> TaskFields:
    """The one widening step both of `resolve_folded_task`'s branches use:
    add the rest of the fold's declared paths to this task's own, and set
    the subject the surviving commit is expected to carry. Written once so
    the two branches cannot drift apart."""
    return replace(
        task,
        files=task.files + extra_files,
        allowed_collateral=task.allowed_collateral + extra_collateral,
        commit=commit,
    )


def _red_tasks(lines: List[str]) -> List[TaskFields]:
    """Every `Build: red` task in the plan that carries a `Squash-with:`
    field at all, in document order — the reds whose folds and whose
    defective fields both have to be accounted for before any task in this
    plan can be checked."""
    return [
        red
        for red in (parse_task_fields(lines, other_id) for other_id in
                    collect_task_ids(lines))
        if red.build == "red" and red.squash_value is not None
    ]


def _joined_partners(lines: List[str], red: TaskFields) -> List[str]:
    """The partner ids of `red` that RESOLVE — the named task exists in this
    plan and is not itself red — and therefore actually join `red`'s commit
    to theirs.

    An edge that does not resolve is not an edge (fix round 6, F15). It is
    a plan defect, reported by `_fold_group` against the red task carrying
    the field; but it names no commit, so it must not pull the named task —
    nor, transitively, the valid fold that task belongs to — into this red's
    membership set. Growing the group through such an edge made one red's
    broken reference fail a separate, entirely valid folded commit from
    every id in it."""
    joined: List[str] = []
    for partner_id in red.squash_partners:
        try:
            partner = parse_task_fields(lines, partner_id)
        except TaskNotFoundError:
            continue
        if partner.build == "red":
            continue
        joined.append(partner_id)
    return joined


def _fold_reds(
    lines: List[str], task: TaskFields, reds: List[TaskFields]
) -> List[str]:
    """Every red task whose fold folds into the ONE commit `task` belongs to,
    in document order (fix round 3, F8). `reds` is `_red_tasks(lines)`,
    already known to name at least one partner each.

    `Squash-with:` is an edge, not a tree. Two red tasks may name the SAME
    green partner, and each of their commits then folds into that one
    partner's commit — so both folds, and every task in either, are one
    combined unit carried by one commit. Resolving only the first red found
    in document order gave the shared partner's id a narrower file union and
    a different subject from the red ids: two verdicts on one commit, which
    is exactly what "either id gives the same verdict" forbids.

    The group is therefore grown to a fixed point: a red belongs if it IS a
    member or NAMES one through a RESOLVING edge, and admitting it makes
    those partners members in turn, which may admit further reds. Membership
    is grown over `_joined_partners`, never over the raw `squash_partners`
    list (fix round 6, F15). A task that is neither a red with partners nor
    named by one yields the empty list and is left alone — the union widens
    a fold, never an ordinary task."""
    edges = {red.id: _joined_partners(lines, red) for red in reds}
    members = {task.id}
    growing = True
    while growing:
        growing = False
        for red in reds:
            joined = edges[red.id]
            if red.id not in members and not any(
                partner in members for partner in joined
            ):
                continue
            reached = {red.id} | set(joined)
            if not reached <= members:
                members |= reached
                growing = True
    return [red.id for red in reds if red.id in members]


def _shared_partner(reds: List[TaskFields]) -> Optional[str]:
    """The first task named by more than one of `reds` — the join that made
    their folds one unit, and the task a disagreement between those folds is
    reported against. `None` when the reds share nobody, which a group of
    two or more reds cannot be (they are only in one group because they
    share a partner or because one names the other, and the latter is
    already a violation); the caller falls back rather than assume it."""
    seen: List[str] = []
    for red in reds:
        for partner in red.squash_partners:
            if partner in seen:
                return partner
            seen.append(partner)
    return None


def resolve_folded_task(
    lines: List[str], task: TaskFields
) -> Tuple[TaskFields, List[str]]:
    """Resolve a folded task against the whole fold its commit carries, per
    `myflow-task-commit-fields`'s requirement **A folded red task is checked
    against its partner's commit**. Returns `(task to check, violations found
    resolving it)`; a non-empty violation list is a plan defect and the caller
    reports it instead of checking anything.

    After the fold a red task has no commit of its own: its declared
    `Commit:` subject exists nowhere, and one commit carries every folded
    task's files. So every task in the fold — the red tasks and their green
    partners alike — is checked against the fold's agreed `Commit:` subject
    with the UNION of the whole fold's `Files:`/`Allowed-collateral:` sets.
    Either id therefore reaches one verdict, invalid folds included: every
    red in the group is re-validated from every id, so a fold one id rejects
    cannot pass through another's.

    The fold is `_fold_reds`'s combined group, not one red's partner list
    (fix round 3, F8): where two reds name the same partner, both commits
    fold into that partner's one commit. Their partner sets must then agree
    about that commit's subject exactly as one red's partners must, and a
    disagreement between the two folds is the same plan defect — reported
    against the shared task that joined them, and naming the reds, rather
    than silently taking the first red in document order.

    The no-subject rule is asked of that same combined group (fix round 4,
    item 2): it fires only when NO red in the group has a partner declaring
    a `Commit:` subject. A red whose own partners declared none takes the
    subject a joined red's partner supplied, because there is one commit and
    therefore one subject.

    A task that neither declares a `Squash-with:` nor is named by a red task
    in the same plan is returned unchanged."""
    reds = _red_tasks(lines)

    # A red whose `Squash-with:` value names no partner the grammar admits
    # is reported from EVERY id in the plan, not only from its own (fix
    # round 6, F16). The guard cannot know which commit that task folded
    # into — that is the whole content of the defect — so it cannot know
    # which folds the plan really has, and no verdict it reaches for any
    # task in this plan is one it can vouch for. Reported only from the red
    # itself, the defect was invisible from the task its free text names,
    # which was instead told that the folded commit's other file was
    # undeclared collateral: true of the commit, and about the wrong field.
    #
    # This is deliberately NOT how an unresolvable EDGE is treated. There
    # the named partner and the reason the edge is illegal are both known,
    # so the guard knows the fold does not exist and reports it against the
    # red task alone (see `_joined_partners`). Here nothing is known.
    unresolved = [red for red in reds if not red.squash_partners]
    if unresolved:
        return task, [
            f"task {red.id}: Squash-with: value {red.squash_value!r} is not "
            "`Task <id>[, <id>...]`, so no merge partner can be resolved "
            "from it"
            for red in unresolved
        ]

    fold_reds = _fold_reds(lines, task, reds)
    if not fold_reds:
        return task, []

    extra_files: List[str] = []
    extra_collateral: List[str] = []
    subjects: List[str] = []
    violations: List[str] = []
    for red_id in fold_reds:
        red = parse_task_fields(lines, red_id)
        group_files, group_collateral, group_subject, group_violations = _fold_group(
            lines, red
        )
        violations += group_violations
        # A red's own files, plus its partners' — including this task's own
        # where it is one of them (a harmless duplicate), its SIBLING
        # partners', and every task in a fold joined to this one.
        if red_id != task.id:
            extra_files += red.files
            extra_collateral += red.allowed_collateral
        extra_files += group_files
        extra_collateral += group_collateral
        if group_subject is not None and group_subject not in subjects:
            subjects.append(group_subject)

    # One combined fold is one commit, so it carries exactly one subject.
    # Both failures — the folds disagreeing, and nobody declaring one at all
    # — are asked of the COMBINED group (fix round 4, item 2 for the second
    # of them): a red whose own partners declared no subject is no defect
    # where a red joined to the same commit supplied one, since that is the
    # fold's subject. Both are gated the same way `_fold_group` gates its own
    # disagreement check: a group already carrying an unresolvable partner
    # has its violation, and what its remaining partners happened to declare
    # is not a second, separate defect. Both are anchored on the shared task
    # that joined the folds, where there is one.
    if not violations and len(subjects) != 1:
        reds = [parse_task_fields(lines, red_id) for red_id in fold_reds]
        shared = _shared_partner(reds)
        anchor = shared if shared is not None else fold_reds[0]
        if subjects:
            violations.append(
                f"task {anchor}: Task {shared or anchor} is folded by Tasks "
                + ", ".join(fold_reds)
                + ", whose partners declare different Commit: subjects "
                + ", ".join(repr(subject) for subject in subjects)
                + " — one folded unit is one commit, so one subject"
            )
        else:
            violations.append(
                f"task {anchor}: no partner named by Squash-with: declares "
                "a Commit: subject in the fold carried by Tasks "
                + ", ".join(fold_reds)
                + " — the folded commit's subject would be checked against "
                "nothing"
            )
    if violations:
        return task, violations
    subject = subjects[0]
    return _widen(task, extra_files, extra_collateral, subject), []


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

    # An unclosed fence in this task's body is reported on its own, before
    # anything else, and REPLACES every downstream check for it (fix round
    # 11, F22). CommonMark runs an unclosed fence to the end of the block,
    # so every field after it is code rather than a declaration: `Files:`
    # arrives empty, `Commit:` absent, `Build:` and `Squash-with:` gone.
    # Checking on against that field set does not merely add noise — it
    # blames the commit for a defect in the plan, reporting the file the
    # task really did declare as undeclared collateral. The precedent is
    # the unresolvable fold below: a task whose declaration cannot be read
    # is one no verdict can be reached for.
    if task.unclosed_fence_line is not None:
        return [
            f"task {task.id}: code fence opened at tasks.md line "
            f"{task.unclosed_fence_line} is never closed in this task's "
            "body, so every field below it is inside the fence and was not "
            "read"
        ], []

    # A folded pair is resolved before anything is checked; an unresolvable
    # one is a plan defect reported on its own, since checking a red task
    # against a commit that is not its own would report every field wrong.
    folded, squash_violations = resolve_folded_task(lines, task)
    if squash_violations:
        return squash_violations, []

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
    # `folded` carries the union file set and the surviving commit's expected
    # subject; `task` carries what this task itself declared, which is what
    # Tests: and the declared-scope check are about either way.
    violations += check_files(folded, changed_files)
    violations += check_tests(task, diff_text)
    violations += check_commit_subject(folded, actual_subject)
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
