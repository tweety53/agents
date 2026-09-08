#!/usr/bin/env python3
"""check-baseline-fresh.py — refuse a `Baseline:` count whose source
test-results directory predates the worktree's newest commit.

KAN-339: a planner wrote its plan's `Baseline:` chain from a stale
build/test-results directory — 1690 against a real 1768 — and every task's
`Baseline:` had to be rewritten mid-run. Nothing required the baseline to
come from a fresh run. This guard makes the freshness mechanical at the
only point it is still observable, plan-writing time: by implementation
time mid-run test executions have already refreshed any results directory
and masked the original staleness, and KAN-442 deliberately keeps the
runtime side git/files-only.

Which directory a count's freshness is judged against is PROJECT-DECLARED:
an optional `## baseline results dirs` key in the project's
`.flow/project.md`, one worktree-relative path per line, outside any
fence — that contract file is canonical for the key. A project declaring
nothing is skipped, per the frozen tree's "skip, rather than fail, when
unsupported" precedent. A declared directory that does not exist is
reported stale like an old one: a directory that is not there cannot
attest the numbers came from a run in this worktree, and that absence is
the KAN-339 shape itself.

Parsing reuses the shared grammar rather than reimplementing it —
`lib/plan_grammar.py`'s `FENCE_RE` fence tracking and
`check-task-commit-fields.py`'s `FIELD_RE` field anchors — the same
structural-import discipline check-plan-shape.py states, and for the same
reason: parser identity is structural, never a comment claiming parity.

Scope is a single change root per invocation: exactly one argument, the
change directory holding the `tasks.md` to judge. The project root is
discovered by walking up from it to the nearest directory holding
`.flow/project.md`.

Exit codes:
  0  fresh, or nothing to check — no non-fenced `**Baseline:**` field in
     the plan, or no `.flow/project.md` above the change root, or the key
     declared nowhere. Skip reasons go to stderr, never stdout.
  1  stale hit — at least one declared directory is absent, or its
     recursive max mtime predates the worktree's newest commit. One line
     per offending directory, each anchored at the plan's first
     `**Baseline:**` field line; the fix is named on every line.
  2  cannot answer — usage, unreadable or missing `tasks.md`, the key
     declared more than once, the key declared with an empty body, or a
     declared path escaping the project root.

Standard library only (see check-plan-provenance.py's module docstring for
why this repository restricts itself to that).
"""

from __future__ import annotations

import importlib.util
import os
import subprocess
import sys
from pathlib import Path
from typing import List, Optional

SCRIPT_DIR = Path(os.path.dirname(os.path.realpath(__file__)))

# --- Load check-task-commit-fields.py --------------------------------------
#
# Named and checked explicitly (not just imported): a Python `import` is
# invisible to check-guard-symlinks.sh's rule 2, which derives a guard's
# required siblings by grepping its SOURCE for `$SCRIPT_DIR/<name>` — the
# .sh wrapper carries that literal reference for rule 2's benefit, and this
# check is defense in depth for anyone invoking this .py directly. The
# sys.modules registration before exec_module is check-plan-shape.py's:
# without it the loaded module's @dataclass decorators raise
# AttributeError on Python 3.14.
_CFG_PATH = SCRIPT_DIR / "check-task-commit-fields.py"
if not _CFG_PATH.is_file():
    print(
        f"check-baseline-fresh.py: required sibling module not found: {_CFG_PATH}",
        file=sys.stderr,
    )
    sys.exit(2)

_cfg_spec = importlib.util.spec_from_file_location("cfg", str(_CFG_PATH))
if _cfg_spec is None or _cfg_spec.loader is None:
    print(
        f"check-baseline-fresh.py: cannot load module: {_CFG_PATH}",
        file=sys.stderr,
    )
    sys.exit(2)
_cfg = importlib.util.module_from_spec(_cfg_spec)
sys.modules["cfg"] = _cfg
_cfg_spec.loader.exec_module(_cfg)

FIELD_RE = _cfg.FIELD_RE

# --- Load lib/plan_grammar.py ----------------------------------------------
_LIB_DIR = SCRIPT_DIR / "lib"
_GRAMMAR_PATH = _LIB_DIR / "plan_grammar.py"
if not _GRAMMAR_PATH.is_file():
    print(
        f"check-baseline-fresh.py: required sibling module not found: {_GRAMMAR_PATH}",
        file=sys.stderr,
    )
    sys.exit(2)
sys.path.insert(0, str(_LIB_DIR))

from plan_grammar import FENCE_RE  # noqa: E402

BASELINE_KEY = "baseline results dirs"


def baseline_field_lines(lines: List[str]) -> List[int]:
    """The 1-based lines of every non-fenced column-0 `**Baseline:**` field
    line, in document order. Fence state tracks across the whole file, so
    a field-shaped line inside a worked example is documentation, not a
    declaration."""
    hits: List[int] = []
    in_fence = False
    for index, line in enumerate(lines):
        if FENCE_RE.match(line):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        match = FIELD_RE.match(line)
        if match is not None and match.group(1) == "Baseline":
            hits.append(index + 1)
    return hits


def find_project_root(change_root: Path) -> Optional[Path]:
    """The nearest directory at or above `change_root` holding
    `.flow/project.md`, or None when no ancestor does — the undeclared
    case, which skips."""
    for candidate in (change_root, *change_root.parents):
        if (candidate / ".flow" / "project.md").is_file():
            return candidate
    return None


def extract_key_body(text: str) -> Optional[str]:
    """The body of the one `## baseline results dirs` section — the lines
    after its heading up to the next `## ` heading — or None when the key
    is declared nowhere. The caller has already refused a second
    declaration."""
    lines = text.splitlines()
    start: Optional[int] = None
    for index, line in enumerate(lines):
        if line.rstrip() == f"## {BASELINE_KEY}":
            start = index + 1
            break
    if start is None:
        return None
    body: List[str] = []
    for line in lines[start:]:
        if line.rstrip().startswith("## "):
            break
        body.append(line)
    return "\n".join(body)


def paths_from_body(body: str) -> tuple[List[str], List[str]]:
    """The declared paths and the body lines that were skipped. `paths` is
    one per line, outside any fence, backticks and surrounding whitespace
    stripped — blank lines are skipped, and so is any line carrying
    internal whitespace. `ignored` is every other non-blank, non-fence
    line, verbatim: a typo'd path must not vanish silently when a valid
    line shares the body."""
    paths: List[str] = []
    ignored: List[str] = []
    in_fence = False
    for line in body.splitlines():
        if FENCE_RE.match(line):
            in_fence = not in_fence
            continue
        if in_fence:
            continue
        token = line.strip().strip("`").strip()
        if not token:
            continue
        if any(c.isspace() for c in token):
            ignored.append(line.strip())
            continue
        paths.append(token)
    return paths, ignored


def _max_mtime(root: Path) -> Optional[float]:
    """The max mtime over `root` and everything under it, None when no
    entry has a readable mtime. The directory's own mtime moves only when
    its direct children change, so the walk is what makes a suite's write
    into a per-class subdirectory count as fresh."""
    newest: Optional[float] = None
    try:
        newest = os.path.getmtime(root)
    except OSError:
        pass
    for base, dirs, files in os.walk(root):
        for name in [*dirs, *files]:
            try:
                mtime = os.path.getmtime(os.path.join(base, name))
            except OSError:
                continue
            if newest is None or mtime > newest:
                newest = mtime
    return newest


def main(argv: List[str]) -> int:
    if len(argv) != 2:
        print(f"usage: {argv[0]} <change-root>", file=sys.stderr)
        return 2

    change_root = Path(argv[1])
    if not change_root.is_dir():
        print(
            f"check-baseline-fresh.py: not a directory: {change_root}",
            file=sys.stderr,
        )
        return 2
    tasks_md = change_root / "tasks.md"
    try:
        text = tasks_md.read_text(encoding="utf-8")
    except OSError as exc:
        print(
            f"check-baseline-fresh.py: cannot read {tasks_md}: {exc}",
            file=sys.stderr,
        )
        return 2
    lines = text.splitlines()

    anchors = baseline_field_lines(lines)
    if not anchors:
        print(
            f"check-baseline-fresh: {tasks_md}: no **Baseline:** field — "
            "nothing to check",
            file=sys.stderr,
        )
        return 0

    project_root = find_project_root(change_root)
    if project_root is None:
        print(
            "check-baseline-fresh: no .flow/project.md above "
            f"{change_root} — the source directory is undeclared, skip",
            file=sys.stderr,
        )
        return 0

    config = project_root / ".flow" / "project.md"
    config_text = config.read_text(encoding="utf-8")
    key_line = f"## {BASELINE_KEY}"
    declarations = sum(
        1 for line in config_text.splitlines() if line.rstrip() == key_line
    )
    if declarations == 0:
        print(
            f"check-baseline-fresh: {config} declares no '{key_line}' "
            "key — skip",
            file=sys.stderr,
        )
        return 0
    if declarations > 1:
        print(
            f"check-baseline-fresh: {config} declares {declarations} "
            f"'{key_line}' keys — a second declaration is ambiguous, so "
            "neither was read",
            file=sys.stderr,
        )
        return 2

    body = extract_key_body(config_text)
    assert body is not None  # declarations == 1
    declared, ignored = paths_from_body(body)
    if not declared:
        print(
            f"check-baseline-fresh: {config}'s '{key_line}' body names no "
            "directory — a declaration that says nothing cannot attest a "
            "count",
            file=sys.stderr,
        )
        return 2
    if ignored:
        print(
            "check-baseline-fresh: ignoring unparseable "
            f"'{key_line}' body line(s): {'; '.join(ignored)}",
            file=sys.stderr,
        )

    real_root = project_root.resolve()
    resolved: List[Path] = []
    for raw in declared:
        full = (project_root / raw).resolve()
        if full != real_root and real_root not in full.parents:
            print(
                "check-baseline-fresh: declared path "
                f"'{raw}' escapes the project root {real_root}",
                file=sys.stderr,
            )
            return 2
        resolved.append(full)

    commit = subprocess.run(
        ["git", "-C", str(real_root), "log", "-1", "--format=%ct", "HEAD"],
        capture_output=True,
        text=True,
    )
    if commit.returncode != 0 or not commit.stdout.strip():
        print(
            "check-baseline-fresh: git cannot answer the worktree's "
            f"newest commit: {commit.stderr.strip()}",
            file=sys.stderr,
        )
        return 2
    commit_epoch = int(commit.stdout.strip())

    violations: List[str] = []
    anchor = anchors[0]
    for raw, full in zip(declared, resolved):
        if not full.is_dir():  # ABSENT-VERDICT
            violations.append(
                f"{tasks_md}:{anchor}: **Baseline:** counts name no fresh "
                f"source: declared directory '{raw}' does not exist in "
                "this worktree — a directory that is not there cannot "
                "attest the numbers came from a run in it; re-run the "
                "suite here and re-derive the counts"
            )
            continue
        newest = _max_mtime(full)
        if newest is None:
            violations.append(
                f"{tasks_md}:{anchor}: **Baseline:** counts name no fresh "
                f"source: nothing under declared directory '{raw}' has a "
                "readable mtime; re-run the suite here and re-derive the "
                "counts"
            )
            continue
        if newest < float(commit_epoch):  # STALE-VERDICT
            violations.append(
                f"{tasks_md}:{anchor}: **Baseline:** counts are judged "
                f"against a stale source: '{raw}' last changed at epoch "
                f"{int(newest)}, before this worktree's newest commit at "
                f"epoch {commit_epoch} — re-run the suite here and "
                "re-derive the counts"
            )
    for violation in violations:
        print(violation)
    return 1 if violations else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
