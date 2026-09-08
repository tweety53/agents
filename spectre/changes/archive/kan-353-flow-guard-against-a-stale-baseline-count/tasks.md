# kan-353-flow-guard-against-a-stale-baseline-count

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** no

---

- [x] 1. The freshness guard and its harness

**Build:** green

**Files:**
- Create: `scripts/check-baseline-fresh.py`
- Create: `scripts/check-baseline-fresh.sh`
- Create: `skills/flow/scripts/check-baseline-fresh.sh` (symlink → `../../../scripts/check-baseline-fresh.sh`)
- Create: `skills/flow/scripts/check-baseline-fresh.py` (symlink → `../../../scripts/check-baseline-fresh.py`)
- Create: `scripts/test-check-baseline-fresh.sh`

**Tests:** `scripts/test-check-baseline-fresh.sh` — 16 cases added (no-Baseline skip, no-project-config skip, undeclared-key skip, fresh dir passes, absent dir fails, stale dir fails, fresh+absent reports only the offender, two stale dirs in one report, duplicate key exit 2, empty key body exit 2, escaping path exit 2, fenced Baseline line ignored, usage exit 2, two mutation cases proving the stale and absent verdicts each depend on their own check, boundary mtime == commit epoch passes)
**Regression:** reverting this task's guard changes while keeping the harness fails all 16 cases — no exit-code contract holds and no stale-directory message is printed
**Baseline:** before=0 after=16 on `scripts/test-check-baseline-fresh.sh` (new harness; before=0 is the base tree's no-harness state)
<!-- predicted: scripts/test-check-baseline-fresh.sh after task 1 — 16 cases, all pass -->
**Commit:** `feat(scripts): refuse a stale Baseline count at plan time`

  - [x] **Step 1: Write the failing harness**

  Create `scripts/test-check-baseline-fresh.sh`, modeled on `test-check-plan-shape.sh`'s
  fixture-driven pattern: fixtures under `mktemp -d`, a `run_guard` helper capturing RC and
  combined output, every case ending in an explicit pass/fail assertion, and `fail`/`pass`
  counters with a non-zero exit when `FAILURES` is non-zero.

```bash verified:authored in-tree for this change
#!/usr/bin/env bash
# Assertion harness for check-baseline-fresh.sh / check-baseline-fresh.py.
#
# Modeled on test-check-plan-shape.sh's fixture-driven pattern: fixtures
# live under mktemp -d, the guard is invoked via a thin run_guard helper
# that captures RC/OUT, and every case ends with an explicit pass/fail
# assertion. Runs the REAL scripts/check-baseline-fresh.sh against REAL
# fixture files on disk — never a copy of its logic. Two MUTATION cases
# disable the guard's stale-verdict and absent-verdict checks (via sed
# against lines tagged `# STALE-VERDICT` and `# ABSENT-VERDICT`) and assert
# the tagged check's own MESSAGE vanishes from the mutant's output — not
# rc 0, which the guard's second layer (the unreadable-mtime verdict) may
# still legitimately produce for the same fixture. A suite that cannot
# detect the guard's own defect is not a suite.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GUARD="$SCRIPT_DIR/check-baseline-fresh.sh"
FAILURES=0

fail() { printf 'FAIL: %s\n' "$1" >&2; FAILURES=$((FAILURES + 1)); }
pass() { printf 'ok: %s\n' "$1"; }

DIRS=()
cleanup() {
  [ "${#DIRS[@]}" -eq 0 ] && return 0
  local d
  for d in "${DIRS[@]}"; do
    rm -rf "$d"
  done
}
trap cleanup EXIT

# run_guard [<changeRoot>] -> sets RC and OUT (combined stdout+stderr)
run_guard() {
  set +e
  OUT="$("$GUARD" "$@" 2>&1)"
  RC=$?
  set -e
}

# new_fixture -> FIXTURE (a git repo whose newest commit sits at epoch
# 1577836800 / 2020-01-01T00:00:00Z), CHANGE (its change root), TASKS_MD
new_fixture() {
  FIXTURE="$(mktemp -d "${TMPDIR:-/tmp}/baseline-fresh-test.XXXXXX")"
  DIRS+=("$FIXTURE")
  git -C "$FIXTURE" init -q
  GIT_COMMITTER_DATE="2020-01-01T00:00:00Z" \
    GIT_AUTHOR_DATE="2020-01-01T00:00:00Z" \
    git -C "$FIXTURE" commit -q --allow-empty -m base
  CHANGE="$FIXTURE/spectre/changes/baseline-fresh-fixture"
  mkdir -p "$CHANGE"
  TASKS_MD="$CHANGE/tasks.md"
}

# write_tasks <path> — a minimal one-task plan carrying a **Baseline:** field
write_tasks() {
  cat > "$1" <<'EOF'
- [ ] 1. Do the thing

**Build:** green
**Files:**
- `src/thing.go`
**Tests:** `none`
**Baseline:** before=10 after=12
**Commit:** `feat(thing): do the thing`
EOF
}

# write_config <path> — a project config declaring one results dir (pass the
# body lines as extra arguments; no arguments declares no key)
write_config() {
  local cfg="$1"
  shift
  {
    printf '# fixture project\n'
    if [ "$#" -gt 0 ]; then
      printf '\n## baseline results dirs\n\n'
      printf '%s\n' "$@"
    fi
  } > "$cfg"
}

# fresh_dir / stale_dir — set a directory's mtime after/before the fixture's
# fixed newest-commit epoch (1577836800)
fresh_dir() { mkdir -p "$1" && touch "$1"; }
stale_dir() { mkdir -p "$1" && touch -t 201901010000 "$1"; }
```

  Then the 15 cases, each `new_fixture` → arrange → `run_guard "$CHANGE"` → assert:

```bash verified:authored in-tree for this change
# case 1 — no **Baseline:** field anywhere -> RC 0, nothing to check
new_fixture
printf -- '- [ ] 1. Do the thing\n' > "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results'
run_guard "$CHANGE"
[ "$RC" -eq 0 ] || fail "case 1: expected RC 0, got $RC: $OUT"
pass "case 1: no Baseline field -> skip"

# case 2 — Baseline field, no .flow/project.md above the change root -> RC 0
new_fixture
write_tasks "$TASKS_MD"
run_guard "$CHANGE"
[ "$RC" -eq 0 ] || fail "case 2: expected RC 0, got $RC: $OUT"
pass "case 2: no project config -> skip"

# case 3 — Baseline field, config without the key -> RC 0
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md"
run_guard "$CHANGE"
[ "$RC" -eq 0 ] || fail "case 3: expected RC 0, got $RC: $OUT"
pass "case 3: key undeclared -> skip"

# case 4 — declared dir fresh (mtime after the fixed commit epoch) -> RC 0
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results'
fresh_dir "$FIXTURE/build/test-results"
run_guard "$CHANGE"
[ "$RC" -eq 0 ] || fail "case 4: expected RC 0, got $RC: $OUT"
pass "case 4: fresh dir -> pass"

# case 5 — declared dir absent -> RC 1, message names the dir
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results'
run_guard "$CHANGE"
[ "$RC" -eq 1 ] || fail "case 5: expected RC 1, got $RC: $OUT"
printf '%s' "$OUT" | grep -q 'build/test-results' || fail "case 5: dir not named: $OUT"
printf '%s' "$OUT" | grep -q 'does not exist' || fail "case 5: absence not stated: $OUT"
pass "case 5: absent dir -> stale"

# case 6 — declared dir stale (mtime 2019 < commit 2020) -> RC 1, both epochs
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results'
stale_dir "$FIXTURE/build/test-results"
run_guard "$CHANGE"
[ "$RC" -eq 1 ] || fail "case 6: expected RC 1, got $RC: $OUT"
printf '%s' "$OUT" | grep -q 'stale' || fail "case 6: staleness not stated: $OUT"
printf '%s' "$OUT" | grep -q '1577836800' || fail "case 6: commit epoch not named: $OUT"
pass "case 6: stale dir -> stale"

# case 7 — one fresh + one absent dir -> RC 1 naming only the offender
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results' 'dist/test-results'
fresh_dir "$FIXTURE/build/test-results"
run_guard "$CHANGE"
[ "$RC" -eq 1 ] || fail "case 7: expected RC 1, got $RC: $OUT"
printf '%s' "$OUT" | grep -q 'dist/test-results' || fail "case 7: offender not named: $OUT"
printf '%s' "$OUT" | grep -q 'build/test-results' && fail "case 7: fresh dir reported: $OUT"
pass "case 7: only the offender reported"

# case 8 — two stale dirs -> RC 1, both named in the one report
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results' 'dist/test-results'
stale_dir "$FIXTURE/build/test-results"
stale_dir "$FIXTURE/dist/test-results"
run_guard "$CHANGE"
[ "$RC" -eq 1 ] || fail "case 8: expected RC 1, got $RC: $OUT"
printf '%s' "$OUT" | grep -q 'build/test-results' || fail "case 8: first dir not named: $OUT"
printf '%s' "$OUT" | grep -q 'dist/test-results' || fail "case 8: second dir not named: $OUT"
pass "case 8: both offenders in one report"

# case 9 — duplicate key declarations -> RC 2
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results'
printf '\n## baseline results dirs\n\ndist/test-results\n' >> "$FIXTURE/.flow/project.md"
run_guard "$CHANGE"
[ "$RC" -eq 2 ] || fail "case 9: expected RC 2, got $RC: $OUT"
pass "case 9: duplicate key -> cannot answer"

# case 10 — key declared with an empty body -> RC 2
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md"
printf '\n## baseline results dirs\n' >> "$FIXTURE/.flow/project.md"
run_guard "$CHANGE"
[ "$RC" -eq 2 ] || fail "case 10: expected RC 2, got $RC: $OUT"
pass "case 10: empty key body -> cannot answer"

# case 11 — declared path escaping the project root -> RC 2
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" '../outside-results'
run_guard "$CHANGE"
[ "$RC" -eq 2 ] || fail "case 11: expected RC 2, got $RC: $OUT"
pass "case 11: escaping path -> cannot answer"

# case 12 — a **Baseline:** line inside a fenced example is not a field -> RC 0
new_fixture
FENCE='```'
{
  printf -- '- [ ] 1. Do the thing\n\n'
  printf -- '**Build:** green\n\n'
  printf -- '%s\n' "$FENCE"
  printf -- '**Baseline:** before=1 after=2\n'
  printf -- '%s\n' "$FENCE"
} > "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results'
run_guard "$CHANGE"
[ "$RC" -eq 0 ] || fail "case 12: expected RC 0, got $RC: $OUT"
pass "case 12: fenced Baseline line ignored"

# case 13 — usage: no arguments -> RC 2
run_guard
[ "$RC" -eq 2 ] || fail "case 13: expected RC 2, got $RC: $OUT"
pass "case 13: usage -> exit 2"

# cases 14–15 — mutations: each verdict depends on its own check. The
# assertion is the house one (test-check-plan-shape.sh's): the tagged
# check's own MESSAGE must vanish from the mutant's output — not rc 0,
# which the guard's second layer (the unreadable-mtime verdict) may still
# legitimately produce for the same fixture.
assert_mutation_removes_finding() {
  local tag="$1" fixture_change="$2" label="$3" needle="$4"
  local mutant_dir out rc
  mutant_dir="$(mktemp -d "${TMPDIR:-/tmp}/baseline-fresh-mutant.XXXXXX")"
  DIRS+=("$mutant_dir")
  mkdir -p "$mutant_dir/lib"
  ln -s "$SCRIPT_DIR/check-task-commit-fields.py" "$mutant_dir/check-task-commit-fields.py"
  ln -s "$SCRIPT_DIR/lib/plan_grammar.py" "$mutant_dir/lib/plan_grammar.py"
  cp "$SCRIPT_DIR/check-baseline-fresh.sh" "$mutant_dir/check-baseline-fresh.sh"
  chmod +x "$mutant_dir/check-baseline-fresh.sh"
  sed "/# ${tag}\$/s/if .*/if False:  # ${tag} (mutated)/" \
    "$SCRIPT_DIR/check-baseline-fresh.py" > "$mutant_dir/check-baseline-fresh.py"
  grep -q "${tag} (mutated)" "$mutant_dir/check-baseline-fresh.py" \
    || { fail "$label: sed edit did not apply"; return 0; }
  set +e
  out="$("$mutant_dir/check-baseline-fresh.sh" "$fixture_change" 2>&1)"
  rc=$?
  set -e
  printf '%s' "$out" | grep -qF -- "$needle" \
    && fail "$label: mutant still reports the finding (rc=$rc): $out"
  pass "$label"
}

# case 14 — stale verdict disabled: the case-6 fixture's stale-source message gone
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results'
stale_dir "$FIXTURE/build/test-results"
STALE_CHANGE="$CHANGE"
assert_mutation_removes_finding 'STALE-VERDICT' "$STALE_CHANGE" 'case 14: stale mutation' 'stale source'

# case 15 — absent verdict disabled: the case-5 fixture's absence message gone
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results'
ABSENT_CHANGE="$CHANGE"
assert_mutation_removes_finding 'ABSENT-VERDICT' "$ABSENT_CHANGE" 'case 15: absent mutation' 'does not exist'

# case 16 — dir mtime exactly at the commit epoch -> RC 0: the comparison is
# >=, and this case pins that boundary so a `<` -> `<=` mutation is caught
new_fixture
write_tasks "$TASKS_MD"
mkdir -p "$FIXTURE/.flow"
write_config "$FIXTURE/.flow/project.md" 'build/test-results'
mkdir -p "$FIXTURE/build/test-results"
touch -t "$(date -r 1577836800 +%Y%m%d%H%M)" "$FIXTURE/build/test-results"
run_guard "$CHANGE"
[ "$RC" -eq 0 ] || fail "case 16: expected RC 0, got $RC: $OUT"
pass "case 16: boundary mtime == commit epoch -> fresh"

printf '%s\n' "failures: $FAILURES"
[ "$FAILURES" -eq 0 ]
```

  - [x] **Step 2: Run the harness to verify it fails**

  Run: `bash scripts/test-check-baseline-fresh.sh`
  Expected: every case FAILs — the guard does not exist yet.

  - [x] **Step 3: Write the guard**

  Create `scripts/check-baseline-fresh.py`. It loads `FIELD_RE` from
  `check-task-commit-fields.py` and `FENCE_RE` from `lib/plan_grammar.py` through
  `check-plan-shape.py`'s exact import discipline (explicit sibling checks, `sys.modules`
  registration before `exec_module`); its docstring states the exit contract and names
  `skills/flow-contracts/project-configuration.md` as canonical for the key.

```python verified:authored in-tree for this change
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
```

  - [x] **Step 4: Write the wrapper and the shipped symlink**

  Create `scripts/check-baseline-fresh.sh` mirroring `check-plan-shape.sh`'s wrapper
  discipline (sibling declarations for `check-guard-symlinks.sh` rule 2, python3 presence
  and runnability probes, argument check), then link both the wrapper and the Python guard
  into the shipped tree — rule 2 requires a symlink beside the shipped wrapper for every
  sibling the wrapper's own source names:

```bash verified:authored in-tree for this change
ln -s ../../../scripts/check-baseline-fresh.sh skills/flow/scripts/check-baseline-fresh.sh
ln -s ../../../scripts/check-baseline-fresh.py skills/flow/scripts/check-baseline-fresh.py
```

  - [x] **Step 5: Run the harness to verify it passes, then lint**

  Run: `bash scripts/test-check-baseline-fresh.sh`
  Expected: 15 `ok:` lines, `failures: 0`, exit 0.
  Run: `scripts/check-guard-symlinks.sh && scripts/check-vocabulary.sh && scripts/check-references.sh && scripts/check-plan-shape.sh`
  Expected: all exit 0.

  - [x] **Step 6: Commit**

```bash unverified:confirm exact paths against the worktree at execution time
git add scripts/check-baseline-fresh.py scripts/check-baseline-fresh.sh \
  scripts/test-check-baseline-fresh.sh skills/flow/scripts/check-baseline-fresh.sh
git commit -m "feat(scripts): refuse a stale Baseline count at plan time"
```

---

- [x] 2. Wire the guard into the planner and the contracts

**Build:** green

**Files:**
- Modify: `skills/flow/brainstorm-planner.md`
- Modify: `skills/flow/SKILL.md`
- Modify: `skills/flow-contracts/project-configuration.md`

**Tests:** **none** — procedure prose and a budget table; the guards below verify shape, budget, vocabulary and references
**Regression:** reverting this task's edits while keeping task 1 leaves a shipped guard nothing invokes — section D never runs it, the guard-presence list does not name it, and the key it reads is undocumented
**Baseline:** n/a — no test declared
**Commit:** `docs(flow): run check-baseline-fresh at plan time`

  - [x] **Step 1: Teach section D to run the guard**

  In `skills/flow/brainstorm-planner.md` (section D), replace the sentence

  > Before continuing, run `check-plan-shape.sh` — a shipped guard, run unconditionally — and the project's configured plan-provenance guard and its configured build-green guard, if the project declares them, and fix any hit.

  with

  > Before continuing, run `check-plan-shape.sh` — a shipped guard, run unconditionally — and `check-baseline-fresh.sh <changeRoot>` — a shipped guard, run unconditionally; it skips when the plan carries no `**Baseline:**` field or the project declares no `## baseline results dirs` key, and refuses a `Baseline:` count whose declared source directory is absent or predates the worktree's newest commit — and the project's configured plan-provenance guard and its configured build-green guard, if the project declares them, and fix any hit.

  - [x] **Step 2: Name the guard in `skills/flow/SKILL.md`'s guard-presence list**

  Insert `check-baseline-fresh.sh` into the union list in **Check guard presence**, in
  alphabetical order after `check-base-moved.sh`.

  - [x] **Step 3: Document the key in `project-configuration.md`**

  In `skills/flow-contracts/project-configuration.md`, add one row to the optional-keys
  table, beside the other single-shape optional keys:

  > `## baseline results dirs` | Optional. One worktree-relative directory path per line, outside any fence — the test-results directories the project's own suites write. `check-baseline-fresh.sh` stats every declared directory against the worktree's newest commit before trusting a plan's `**Baseline:**` counts; the guard is canonical for the parse and its exit codes. Absent means the guard skips — this key is the skip switch, per the frozen tree's "skip, rather than fail, when unsupported" precedent. A duplicated key, a declared-but-empty body, or a path escaping the project root is a cannot-answer exit, never a silent skip. |

  - [x] **Step 4: Check the budgets the edits spent**

  Run: `scripts/check-contract-budget.sh`
  Measured at execution: the three grown files sit at 20544, 14659 and 48151 bytes against
  rows of 23248, 16278 and 48175 — `BUDGET-OK`, no raise needed. A row is raised (new byte
  size plus 25%, rounded up, as the change's last edit) only when a file is measured past
  its row; none was.

  - [x] **Step 5: Lint and commit**

  Run: `scripts/check-vocabulary.sh && scripts/check-references.sh && scripts/check-plan-shape.sh && scripts/check-contract-budget.sh && scripts/check-guard-symlinks.sh && scripts/check-normative-inventory.sh > /dev/null`
  Expected: all exit 0.

```bash unverified:confirm exact paths against the worktree at execution time
git add skills/flow/brainstorm-planner.md skills/flow/SKILL.md \
  skills/flow-contracts/project-configuration.md
git commit -m "docs(flow): run check-baseline-fresh at plan time"
```
