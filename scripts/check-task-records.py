"""check-task-records.py — lint one repository's in-flight task records
against the commits on its branch (KAN-410).

The per-commit guard, check-task-commit-fields.py, checks ONE task's fields
against ONE real commit at the moment that commit is made. Its findings in
kan-29's self-review (F14, F1 and the malformed **Build:** lines of tasks
43-45) were all drift it cannot see, because it runs before the drift
happens: a checkbox ticked with no commit behind it, a commit that touched
files its task never declared, a Build tag mangled by a later hand edit.
This guard is the retroactive, whole-plan pass those findings asked for:
every non-archived spectre/changes/*/tasks.md, every task, checked against
the branch's commits since the base ref. Zero plans in flight is clean —
the bare-tree case a lint step runs in.

Usage (via the check-task-records.sh wrapper):

    check-task-records.py <repo-root> [base-ref]

`base-ref` is the branch the change branched from, bare (`main`) or
remote-tracking (`origin/main`); a bare name prefers
refs/remotes/origin/<name> when that resolves, so a stale local branch is
never compared against. Absent, it is the branch `refs/remotes/origin/HEAD`
points at. The commit range is `<base-ref>..HEAD`.

Three checks, per task:

1. Tick state vs commit existence. A task declaring a `**Commit:**`
   subject that is ticked must have a commit carrying that subject
   reachable from HEAD — a landed plan whose commits are already in the
   base reads clean, which is the state every unarchived plan left behind
   by a merged change sits in. A task that is unticked must have no commit
   carrying its subject in `<base-ref>..HEAD` (work the branch started
   that the record does not claim); a commit reachable from HEAD but older
   than the base is not this branch's business and is not a violation. A
   task declaring no subject carries no expectation and is skipped here.
2. Declared `**Files:**` vs the commit's real files. For each ticked
   task's subject, the files `git diff-tree` reports for the newest
   reachable commit carrying it are checked against the
   union of the declaring tasks' `**Files:**` plus `**Allowed-collateral:**`
   globs — in both directions, which is the part the per-commit guard
   cannot do: a file the commit touched that nothing declared, and a file
   the record declares that the commit never touched. Tasks sharing one
   subject (a red fold: every partner declares the folded commit's agreed
   subject) form one group keyed on that subject, so the fold's union is
   what the commit is checked against — the same semantics
   resolve_folded_task enforces at commit time, reached here from the
   subjects themselves without re-implementing bundle resolution. A merge
   commit yields no diff-tree file list and is accepted silently; a flow
   branch carries none.
3. `**Build:**` tag format. First column-0 `**Build:**` line per task, via
   the shared grammar's select_build_tag: absent is a violation, and a
   value opening with neither green nor red is a violation naming the line.
   This deliberately re-states a check the plan-time lint
   (check-task-build-green.sh) already makes: a record drifts after plan
   time — kan-29's tasks 43-45 are the case that motivated this guard —
   and because both guards read the same shared grammar, the overlap
   carries no second definition to drift apart.

Parsing is imported, never re-stated: task lines, bodies, Build tags and
fences come from lib/plan_grammar.py, field parsing (continuation joining,
backtick stripping, collateral globs) and the undeclared-file check from
check-task-commit-fields.py itself — the same parser identity
check-plan-shape.py loads.

Output: one violation per line, `<tasks.md path relative to the repo
root>: task <id> ...`, in first-come order. Exit codes:

  0  clean, or nothing to check (no plans in flight, no declaring tasks)
  1  violations found
  2  could not judge — usage, an unreadable plan, a sibling module or a
     git call that cannot evaluate its arguments; never a verdict about a
     record. Every exit-2 refusal prints the stderr opening
     "check-task-records: COULD NOT JUDGE — not a record verdict:" so a
     caller mistake is never read as the guard refusing a record.
"""

from __future__ import annotations

import importlib.util
import subprocess
import sys
from pathlib import Path
from typing import Dict, List, NoReturn, Optional

SCRIPT_DIR = Path(__file__).resolve().parent
GRAMMAR_PATH = SCRIPT_DIR / "lib" / "plan_grammar.py"
COMMIT_FIELDS_PATH = SCRIPT_DIR / "check-task-commit-fields.py"


def could_not_judge(detail: str) -> None:
    print(
        f"check-task-records: COULD NOT JUDGE — not a record verdict: {detail}",
        file=sys.stderr,
    )


def die(detail: str) -> NoReturn:
    could_not_judge(detail)
    sys.exit(2)


def load_module(name: str, path: Path):
    if not path.is_file():
        die(f"required sibling module not found: {path}")
    spec = importlib.util.spec_from_file_location(name, str(path))
    if spec is None or spec.loader is None:
        die(f"cannot load sibling module: {path}")
    module = importlib.util.module_from_spec(spec)
    # Registered before exec: @dataclass resolves annotations through
    # sys.modules[cls.__module__], which does not exist for a module that
    # was never registered.
    sys.modules[name] = module
    spec.loader.exec_module(module)
    return module


sys.path.insert(0, str(SCRIPT_DIR / "lib"))
grammar = load_module("plan_grammar", GRAMMAR_PATH)
# check-task-commit-fields.py resolves its own plan_grammar import through
# its own real path; the sys.path entry above keeps that resolution working
# when this guard is invoked through the skills/flow/scripts symlink, where
# CWD is arbitrary.
cff = load_module("check_task_commit_fields", COMMIT_FIELDS_PATH)


def git(root: Path, args: List[str]) -> str:
    proc = subprocess.run(
        ["git", "-C", str(root), *args], capture_output=True, text=True
    )
    if proc.returncode != 0:
        die(f"git {' '.join(args)} failed: {proc.stderr.strip() or 'non-zero exit'}")
    return proc.stdout


def ref_resolves(root: Path, ref: str) -> bool:
    proc = subprocess.run(
        ["git", "-C", str(root), "rev-parse", "--verify", "--quiet", ref],
        capture_output=True,
        text=True,
    )
    return proc.returncode == 0


def resolve_base(root: Path, base_arg: Optional[str]) -> str:
    if base_arg:
        for candidate in (f"refs/remotes/origin/{base_arg}", base_arg):
            if ref_resolves(root, candidate):
                return candidate
        die(f"base-ref does not resolve: {base_arg}")
    proc = subprocess.run(
        [
            "git", "-C", str(root), "symbolic-ref", "--short",
            "refs/remotes/origin/HEAD",
        ],
        capture_output=True,
        text=True,
    )
    if proc.returncode != 0:
        die("refs/remotes/origin/HEAD does not resolve and no base-ref "
            "argument was given")
    return proc.stdout.strip()


def subject_map(root: Path, rev: str) -> Dict[str, List[str]]:
    """subject -> shas reachable from `rev`, newest first (`git log` order).
    A subject a record declares should identify exactly one commit; the
    NEWEST one is what the file comparison reads (the state a rework or
    re-squash left behind), and more than one is its own violation, raised
    once per subject group."""
    out = git(root, ["log", rev, "--format=%H%x09%s"])
    subjects: Dict[str, List[str]] = {}
    for line in out.splitlines():
        if not line:
            continue
        sha, _, subject = line.partition("\t")
        subjects.setdefault(subject, []).append(sha)
    return subjects


def commit_files(root: Path, sha: str) -> List[str]:
    out = git(
        root, ["diff-tree", "--root", "-r", "--name-only", "--no-commit-id", sha]
    )
    return [line for line in out.splitlines() if line]


def main(argv: List[str]) -> int:
    if len(argv) > 2:
        could_not_judge(
            "usage: check-task-records.py <repo-root> [base-ref] — got "
            f"{len(argv)} argument(s)"
        )
        return 2
    root = Path(argv[0]).resolve() if argv else Path.cwd().resolve()
    if not root.is_dir():
        die(f"repo root not found: {root}")

    changes_dir = root / "spectre" / "changes"
    tasks_files: List[Path] = []
    if changes_dir.is_dir():
        for entry in sorted(changes_dir.iterdir()):
            candidate = entry / "tasks.md"
            if entry.is_dir() and candidate.is_file():
                tasks_files.append(candidate)
    if not tasks_files:
        return 0

    base = resolve_base(root, argv[1] if len(argv) == 2 else None)
    head_subjects = subject_map(root, "HEAD")
    range_subjects = subject_map(root, f"{base}..HEAD")

    violations: List[str] = []
    # subject -> {sha, rel, ids, files, collateral}. A subject's partners
    # live in one plan, so the first declaring task's plan path is the
    # attribution for the whole group; a commit already reported under one
    # plan is never re-checked under another.
    groups: Dict[str, Dict[str, object]] = {}

    for tasks_md in tasks_files:
        rel = tasks_md.relative_to(root).as_posix()
        try:
            lines = tasks_md.read_text().splitlines()
        except OSError as err:
            die(f"cannot read {tasks_md}: {err}")
        for task in grammar.iter_tasks(lines):
            match = grammar.TASK_LINE_RE.match(lines[task.task_line - 1])
            ticked = match is not None and match.group("state") == "x"

            tag = grammar.select_build_tag(task.lines)
            if tag is None:
                violations.append(f"{rel}: task {task.id} has no **Build:** tag")
            elif tag.kind is None:
                violations.append(
                    f"{rel}: task {task.id} has a malformed **Build:** tag: "
                    f"'{tag.value.strip()}'"
                )

            try:
                fields = cff.parse_task_fields(lines, task.id)
            except cff.TaskNotFoundError:
                die(f"task {task.id} found by the grammar but not by the "
                    f"field parser in {tasks_md}")
            if not fields.commit:
                continue
            subject = fields.commit
            shas = head_subjects.get(subject) or []
            if not ticked:
                if range_subjects.get(subject):
                    violations.append(
                        f"{rel}: task {task.id} not ticked but commit with "
                        f"subject '{subject}' exists in {base}..HEAD "
                        f"({shas[0][:12]})"
                    )
                continue
            if not shas:
                violations.append(
                    f"{rel}: task {task.id} ticked but no commit with "
                    f"subject '{subject}' reachable from HEAD"
                )
                continue
            sha = shas[0]
            group = groups.setdefault(
                subject,
                {
                    "sha": sha,
                    "rel": rel,
                    "ids": [],
                    "files": [],
                    "collateral": [],
                    "commit_count": len(range_subjects.get(subject) or []),
                },
            )
            if task.id not in group["ids"]:
                group["ids"].append(task.id)
            for path in fields.files:
                if path not in group["files"]:
                    group["files"].append(path)
            for glob in fields.allowed_collateral:
                if glob not in group["collateral"]:
                    group["collateral"].append(glob)

    for subject, group in groups.items():
        sha = str(group["sha"])
        rel = str(group["rel"])
        if int(str(group["commit_count"])) > 1:
            violations.append(
                f"{rel}: task {'+'.join(str(i) for i in group['ids'])}: "
                f"subject '{subject}' is carried by {group['commit_count']} "
                f"commits in {base}..HEAD; expected one"
            )
        changed = commit_files(root, sha)
        if not changed:
            # A merge commit: diff-tree lists no files for it, so there is
            # no file comparison to make — it is accepted silently, as the
            # module docstring states. A flow branch carries none.
            continue
        changed_set = set(changed)
        merged = cff.TaskFields(
            id="+".join(str(i) for i in group["ids"]),
            files=list(group["files"]),
            allowed_collateral=list(group["collateral"]),
        )
        for violation in cff.check_files(merged, changed):
            violations.append(f"{rel}: {violation}")
        for path in merged.files:
            if path not in changed_set:
                violations.append(
                    f"{rel}: task {merged.id}: declared file {path} "
                    f"is not touched by its commit {sha[:12]}"
                )

    for violation in violations:
        print(violation)
    return 1 if violations else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
