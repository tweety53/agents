#!/usr/bin/env bash
# Assertion harness for the git-config-pin audit (KAN-596, deferred self-review
# of KAN-540). The audit's rule: a guard whose verdict reads git state must not
# inherit the invoking machine's config — rename detection, whitespace action
# and untracked-file visibility are each pinned explicitly at every call whose
# output decides something. KAN-540 fixed one instance
# (check-task-commit-fields.py's changed-file list); this harness pins the
# class, both by asserting the audit's verdicts on seeded fixtures and by
# running the audit live against this repository's own corpus, so a future
# unpinned invocation lands as a red harness instead of as a machine-dependent
# guard verdict.
#
# WHY THE AUDITOR LIVES INSIDE THIS FILE. It is a harness concern, not a
# project-configured guard: .flow/project.md's `## lint` list is not extended,
# and run-guard-tests.sh discovers this file by glob. The auditor takes one
# argument — the scripts-like directory to scan — so the fixture cases below
# point it at a sandbox and the live case points it at the real corpus; the
# default is this repository's own scripts/ directory.
#
# THE THREE RULES, and their exemptions (all deliberate):
#   R1  every `git diff`/`git diff-tree` whose output carries names or text
#       (anything without --quiet, whose emptiness verdict a rename cannot
#       flip) must carry --no-renames — rename detection on by default elides
#       a rename's source path from the name list (KAN-540's panel F1);
#   R2  every `git apply` except `--numstat` (a patch read, not an
#       application) must carry --whitespace=nowarn — apply.whitespace/
#       core.whitespace otherwise decide whether a patch fails or is
#       silently rewritten;
#   R3  every `git status --porcelain` must carry an explicit
#       --untracked-files= — status.showUntrackedFiles otherwise decides
#       whether newly-created untracked residue is visible at all.
# The `g()` wrapper used by refresh-main-checkout.sh and
# land-self-review-report.sh, and the `"$GIT_BIN"` / `$GIT_BIN` command form
# the panel lib resolves git to, are scanned too: a call through either is a
# git call.
#
# Bash 3.2 is the floor, as test-check-finish-preflight.sh's header records.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SELF_NAME="test-git-config-pins.sh"

PASS=0
FAIL=0
FAILURES=()

pass() { PASS=$((PASS + 1)); printf 'ok:   %s\n' "$1"; }
fail() { FAIL=$((FAIL + 1)); FAILURES+=("$1"); printf 'FAIL: %s\n' "$1"; }

# run_audit <root> — print the auditor's output; its exit code rides in
# AUDIT_RC. The auditor is the embedded python below: one scanner, three
# rules, hits printed as file:line lines.
run_audit() {
  AUDIT_RC=0
  AUDIT_OUT="$(python3 - "$1" <<'PYAUDIT'
import re
import sys
from pathlib import Path

root = Path(sys.argv[1]).resolve()
if not root.is_dir():
    print(f"audit-git-config-pins: not a directory: {root}", file=sys.stderr)
    sys.exit(2)

targets = sorted(
    p
    for pat in ("*.sh", "*.py", "lib/*.sh", "lib/*.py")
    for p in root.glob(pat)
    if not p.name.startswith("test-")
)

OPTION_WITH_VALUE = {"-C", "--git-dir", "--work-tree", "-c"}


def subcommand(tokens):
    i = 0
    while i < len(tokens):
        t = tokens[i]
        if t in OPTION_WITH_VALUE:
            i += 2
            continue
        if t.startswith("-"):
            i += 1
            continue
        return t
    return None


def norm(tokens):
    # Shell syntax rides along in a raw token — `--quiet)"`, `"--no-renames",`
    # — so comparisons see the flag, never the quoting around it.
    return [t.strip("\"'(),;") for t in tokens]


def classify(location, sub, joined, hits):
    tokens = norm(joined.split())
    if sub in ("diff", "diff-tree"):
        if "--quiet" in tokens:
            return
        if "--no-renames" not in tokens:
            hits.append(
                f"{location}: R1 — this {sub} read decides on names or text; "
                "pin --no-renames so the verdict cannot inherit the machine's "
                "rename detection"
            )
    elif sub == "apply":
        if "--numstat" in tokens:
            return
        if "--whitespace=nowarn" not in tokens:
            hits.append(
                f"{location}: R2 — this apply's success or content depends on "
                "whitespace config; pin --whitespace=nowarn explicitly"
            )
    elif sub == "status" and any(t.startswith("--porcelain") for t in tokens):
        if not any(t.startswith("--untracked-files=") for t in tokens):
            hits.append(
                f"{location}: R3 — this porcelain status decides on visible "
                "entries; pin --untracked-files= so status.showUntrackedFiles "
                "cannot hide residue"
            )


def scan_shell(path, hits):
    uses_wrapper = False
    lines = path.read_text(errors="replace").splitlines()
    for lineno, line in enumerate(lines, 1):
        if line.lstrip().startswith("#"):
            continue
        # An echo or printf line emits text about git; it never runs it.
        if re.match(r"\s*(echo|printf)\b", line):
            continue
        if re.match(r"\s*g\(\)\s*\{", line):
            uses_wrapper = True
        for m in re.finditer(
            r'(?<![\w.$/-])(?:git\b|"\$GIT_BIN"|\$\{GIT_BIN\}|\$GIT_BIN)', line
        ):
            tail = line[m.end():]
            sub = subcommand(tail.split())
            if sub:
                classify(f"{path}:{lineno}", sub, tail, hits)
        if uses_wrapper:
            for m in re.finditer(r"(?<![\w.$/-])g\s+", line):
                if line[: m.start()].rstrip().endswith("{"):
                    continue  # the wrapper definition itself
                tail = line[m.end():]
                sub = subcommand(tail.split())
                if sub:
                    classify(f"{path}:{lineno}", sub, tail, hits)


def scan_python(path, hits):
    text = path.read_text(errors="replace")
    for m in re.finditer(r"\[\s*(['\"])(diff|diff-tree|apply|status)\1\s*,", text):
        depth = 0
        i = m.start()
        quote = None
        end = len(text)
        while i < end:
            ch = text[i]
            if quote:
                if ch == "\\":
                    i += 2
                    continue
                if ch == quote:
                    quote = None
            elif ch in "\"'":
                quote = ch
            elif ch == "[":
                depth += 1
            elif ch == "]":
                depth -= 1
                if depth == 0:
                    end = i
                    break
            i += 1
        span = text[m.start() : end]
        lineno = text[: m.start()].count("\n") + 1
        sub = m.group(2)
        classify(f"{path}:{lineno}", sub, span, hits)


hits = []
for path in targets:
    if path.suffix == ".py":
        scan_python(path, hits)
    else:
        scan_shell(path, hits)

for h in hits:
    print(h)
sys.exit(1 if hits else 0)
PYAUDIT
)" || AUDIT_RC=$?
}

SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/test-git-config-pins.XXXXXX")"
trap 'rm -rf "$SANDBOX"' EXIT
mkdir -p "$SANDBOX/scripts/lib"

# --- case 1: R1 — rename detection ---------------------------------------
cat > "$SANDBOX/scripts/guard-bad-rename.sh" <<'EOF'
STAGED="$(git -C "$ROOT" diff --cached --name-only 2>/dev/null)"
EOF
cat > "$SANDBOX/scripts/guard-good-rename.sh" <<'EOF'
STAGED="$(git -C "$ROOT" diff --no-renames --cached --name-only 2>/dev/null)"
QUIET_OK="$(git -C "$ROOT" diff --cached --quiet)"
EOF
run_audit "$SANDBOX/scripts"
if [ "$AUDIT_RC" -eq 1 ] && grep -q 'guard-bad-rename.sh:1: R1' <<<"$AUDIT_OUT" \
  && ! grep -q 'guard-good-rename.sh' <<<"$AUDIT_OUT"; then
  pass "case 1: R1 — the unpinned diff is named, the pinned and --quiet ones are not"
else
  fail "case 1: R1 — rc=$AUDIT_RC out=[$AUDIT_OUT]"
fi

# --- case 2: R2 — whitespace action --------------------------------------
cat > "$SANDBOX/scripts/guard-bad-apply.sh" <<'EOF'
APPLY_CHECK_ERR="$(git apply --check "$PATCH_FILE" 2>&1)"
NUMSTAT_OK="$(git apply --numstat "$PATCH_FILE")"
EOF
cat > "$SANDBOX/scripts/guard-good-apply.sh" <<'EOF'
git apply --whitespace=nowarn "$PATCH_FILE"
EOF
run_audit "$SANDBOX/scripts"
if [ "$AUDIT_RC" -eq 1 ] && grep -q 'guard-bad-apply.sh:1: R2' <<<"$AUDIT_OUT" \
  && ! grep -q 'guard-good-apply.sh' <<<"$AUDIT_OUT" \
  && ! grep -q 'guard-bad-apply.sh:2' <<<"$AUDIT_OUT"; then
  pass "case 2: R2 — the unpinned apply is named, --numstat and --whitespace= are not"
else
  fail "case 2: R2 — rc=$AUDIT_RC out=[$AUDIT_OUT]"
fi

# --- case 3: R3 — untracked visibility -----------------------------------
cat > "$SANDBOX/scripts/guard-bad-status.sh" <<'EOF'
STATUS_OUT="$(git -C "$WORKTREE" status --porcelain --untracked-files=normal 2>/dev/null)"
HIDDEN="$(git status --porcelain > "$DIR/before.status")"
EOF
run_audit "$SANDBOX/scripts"
if [ "$AUDIT_RC" -eq 1 ] && grep -q 'guard-bad-status.sh:2: R3' <<<"$AUDIT_OUT" \
  && ! grep -q 'guard-bad-status.sh:1' <<<"$AUDIT_OUT"; then
  pass "case 3: R3 — the unpinned porcelain status is named, the pinned one is not"
else
  fail "case 3: R3 — rc=$AUDIT_RC out=[$AUDIT_OUT]"
fi

# --- case 4: python list forms, the g() wrapper, and the live corpus ------
cat > "$SANDBOX/scripts/guard-bad.py" <<'EOF'
changed = [
    p
    for p in run_git(
        worktree,
        ["diff", "--name-only", f"{parent}..{sha}"],
    ).splitlines()
    if p
]
EOF
cat > "$SANDBOX/scripts/guard-good.py" <<'EOF'
changed = [
    p
    for p in run_git(
        worktree,
        ["diff", "--no-renames", "--name-only", f"{parent}..{sha}"],
    ).splitlines()
    if p
]
EOF
cat > "$SANDBOX/scripts/lib/wrapper.sh" <<'EOF'
g() { git -C "$REPO" "$@"; }
g diff --quiet || refuse "has unstaged changes"
g status --porcelain > "$DIR/after.status"
EOF
run_audit "$SANDBOX/scripts"
if [ "$AUDIT_RC" -eq 1 ] && grep -q 'guard-bad.py:5: R1' <<<"$AUDIT_OUT" \
  && ! grep -q 'guard-good.py' <<<"$AUDIT_OUT" \
  && grep -q 'lib/wrapper.sh:3: R3' <<<"$AUDIT_OUT" \
  && ! grep -q 'wrapper.sh:2' <<<"$AUDIT_OUT"; then
  pass "case 4: python lists and the g() wrapper are scanned; --quiet stays exempt"
else
  fail "case 4: rc=$AUDIT_RC out=[$AUDIT_OUT]"
fi

# --- case 5: the live corpus — the pins hold in this repository -----------
run_audit "$SCRIPT_DIR"
if [ "$AUDIT_RC" -eq 0 ] && [ -z "$AUDIT_OUT" ]; then
  pass "case 5: live corpus — every verdict-bearing git call is pinned"
else
  fail "case 5: live corpus is not pinned (rc=$AUDIT_RC):
$AUDIT_OUT"
fi

printf '\n%s cases, %s passed, %s failed\n' "$((PASS + FAIL))" "$PASS" "$FAIL"
if [ "$FAIL" -gt 0 ]; then
  printf '%s\n' "${FAILURES[@]}"
  exit 1
fi
