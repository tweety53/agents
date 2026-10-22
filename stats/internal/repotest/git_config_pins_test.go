package repotest

// TestGitConfigPins is the assertion harness for the git-config-pin audit
// (KAN-596, deferred self-review of KAN-540). The audit's rule: a guard whose
// verdict reads git state must not inherit the invoking machine's config —
// rename detection, whitespace action and untracked-file visibility are each
// pinned explicitly at every call whose output decides something. KAN-540 fixed
// one instance (check-task-commit-fields's changed-file list, now
// stats/internal/guard/taskcommitfields.go); this test pins the class, both by
// asserting the audit's verdicts on seeded fixtures and by running the audit
// live against this repository's own corpus, so a future unpinned invocation
// lands as a red test instead of as a machine-dependent guard verdict.
//
// WHY THE AUDITOR LIVES INSIDE THIS FILE. It is a test concern, not a
// project-configured guard: .flow/project.md's `## lint` list is not extended.
// The auditor (auditPy, run by python3) takes one argument — the scripts-like
// directory to scan — so the fixture cases point it at a sandbox and the live
// cases point it at the real corpus. The same auditor scans a directory's
// non-test *.go files too, reading each git argument list ("diff", "status",
// ... up to its call's closing parenthesis): the guards ported to Go (KAN-760)
// carry the pins KAN-540 and KAN-596 put in their bash and python originals,
// so case 7 points it at stats/internal/guard/ — without it those pins could
// be deleted and `go test` stay green (panel round 2, F18). The auditor skips
// test-* and *_test.go files; case 6 asserts that skip.
//
// THE THREE RULES, and their exemptions (all deliberate):
//
//	R1  every `git diff`/`git diff-tree` whose output carries names or text
//	    (anything without --quiet, whose emptiness verdict a rename cannot
//	    flip) must carry --no-renames — rename detection on by default elides
//	    a rename's source path from the name list (KAN-540's panel F1);
//	R2  every `git apply` except `--numstat` (a patch read, not an
//	    application) must carry --whitespace=nowarn — apply.whitespace/
//	    core.whitespace otherwise decide whether a patch fails or is
//	    silently rewritten;
//	R3  every `git status --porcelain` must carry an explicit
//	    --untracked-files= — status.showUntrackedFiles otherwise decides
//	    whether newly-created untracked residue is visible at all.
//
// A bash `g()` git wrapper, and the `"$GIT_BIN"` / `$GIT_BIN` command form the
// panel lib resolves git to, are scanned too: a call through either is a git
// call.

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runAudit runs the auditor over root, returning its stdout and exit code.
func runAudit(t *testing.T, root string) (string, int) {
	t.Helper()
	cmd := exec.Command("python3", "-", root)
	cmd.Stdin = strings.NewReader(auditPy)
	cmd.Env = append(os.Environ(), "HOME="+t.TempDir())
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	err := cmd.Run()
	var ee *exec.ExitError
	switch {
	case errors.As(err, &ee):
		return out.String(), ee.ExitCode()
	case err != nil:
		t.Fatalf("run auditor: %v", err)
	}
	return out.String(), 0
}

// sandbox writes files (relative path -> content) under a fresh temp dir and
// returns it.
func sandbox(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestGitConfigPins(t *testing.T) {
	t.Parallel()
	has := strings.Contains

	t.Run("case 1: R1 — the unpinned diff is named, the pinned and --quiet ones are not", func(t *testing.T) {
		t.Parallel()
		out, rc := runAudit(t, sandbox(t, map[string]string{
			"guard-bad-rename.sh": `STAGED="$(git -C "$ROOT" diff --cached --name-only 2>/dev/null)"` + "\n",
			"guard-good-rename.sh": `STAGED="$(git -C "$ROOT" diff --no-renames --cached --name-only 2>/dev/null)"
QUIET_OK="$(git -C "$ROOT" diff --cached --quiet)"
`,
		}))
		if rc != 1 || !has(out, "guard-bad-rename.sh:1: R1") || has(out, "guard-good-rename.sh") {
			t.Errorf("rc=%d out=[%s]", rc, out)
		}
	})

	t.Run("case 2: R2 — the unpinned apply is named, --numstat and --whitespace= are not", func(t *testing.T) {
		t.Parallel()
		out, rc := runAudit(t, sandbox(t, map[string]string{
			"guard-bad-apply.sh": `APPLY_CHECK_ERR="$(git apply --check "$PATCH_FILE" 2>&1)"
NUMSTAT_OK="$(git apply --numstat "$PATCH_FILE")"
`,
			"guard-good-apply.sh": `git apply --whitespace=nowarn "$PATCH_FILE"` + "\n",
		}))
		if rc != 1 || !has(out, "guard-bad-apply.sh:1: R2") || has(out, "guard-good-apply.sh") || has(out, "guard-bad-apply.sh:2") {
			t.Errorf("rc=%d out=[%s]", rc, out)
		}
	})

	t.Run("case 3: R3 — the unpinned porcelain status is named, the pinned one is not", func(t *testing.T) {
		t.Parallel()
		out, rc := runAudit(t, sandbox(t, map[string]string{
			"guard-bad-status.sh": `STATUS_OUT="$(git -C "$WORKTREE" status --porcelain --untracked-files=normal 2>/dev/null)"
HIDDEN="$(git status --porcelain > "$DIR/before.status")"
`,
		}))
		if rc != 1 || !has(out, "guard-bad-status.sh:2: R3") || has(out, "guard-bad-status.sh:1") {
			t.Errorf("rc=%d out=[%s]", rc, out)
		}
	})

	t.Run("case 4: python lists and the g() wrapper are scanned; --quiet stays exempt", func(t *testing.T) {
		t.Parallel()
		out, rc := runAudit(t, sandbox(t, map[string]string{
			"guard-bad.py": `changed = [
    p
    for p in run_git(
        worktree,
        ["diff", "--name-only", f"{parent}..{sha}"],
    ).splitlines()
    if p
]
`,
			"guard-good.py": `changed = [
    p
    for p in run_git(
        worktree,
        ["diff", "--no-renames", "--name-only", f"{parent}..{sha}"],
    ).splitlines()
    if p
]
`,
			"lib/wrapper.sh": `g() { git -C "$REPO" "$@"; }
g diff --quiet || refuse "has unstaged changes"
g status --porcelain > "$DIR/after.status"
`,
		}))
		if rc != 1 || !has(out, "guard-bad.py:5: R1") || has(out, "guard-good.py") ||
			!has(out, "lib/wrapper.sh:3: R3") || has(out, "wrapper.sh:2") {
			t.Errorf("rc=%d out=[%s]", rc, out)
		}
	})

	t.Run("case 5: live corpus — every verdict-bearing git call is pinned", func(t *testing.T) {
		t.Parallel()
		if out, rc := runAudit(t, filepath.Join(repoRoot(t), "scripts")); rc != 0 || out != "" {
			t.Errorf("live corpus is not pinned (rc=%d):\n%s", rc, out)
		}
	})

	t.Run(`case 6: go — unpinned diff (incl. exec.Command(<var>, "diff")) and status are named; pinned, multi-line, composite-literal, map-index, diff(1)-program and _test.go are not`, func(t *testing.T) {
		t.Parallel()
		out, rc := runAudit(t, sandbox(t, map[string]string{
			"guard.go": `func a() { names, err := runGit(env, wt, "diff", "--name-only", parent+".."+commit) }
func b() { names, err := runGit(env, wt, "diff", "--no-renames", "--name-only", parent+".."+commit) }
func c() { out, err := exec.Command("git", "-C", wt, "status", "--porcelain").Output() }
func d() {
	out, err := exec.Command("git", "-C", wt,
		"status", "--porcelain", "--untracked-files=normal").Output()
	s := pcRaw(o["status"])
}
func e() { args := append([]string{"-C", repo, "diff", "--name-only", "--no-renames"}, cached...) }
func f() { out, err := exec.Command(gitBin, "diff", "--name-only").Output() }
func g() { p, ok := lookPath(env, "diff"); d := exec.Command("diff", a, b) }
`,
			"guard_test.go": `func t() { runGit(env, wt, "diff", "--name-only") }` + "\n",
		}))
		lines := strings.Count(strings.TrimRight(out, "\n"), "\n") + 1
		if rc != 1 || !has(out, "guard.go:1: R1") || !has(out, "guard.go:3: R3") || !has(out, "guard.go:10: R1") || lines != 3 {
			t.Errorf("rc=%d out=[%s]", rc, out)
		}
	})

	t.Run("case 7: live Go guards — every verdict-bearing git call is pinned", func(t *testing.T) {
		t.Parallel()
		if out, rc := runAudit(t, filepath.Join(repoRoot(t), "stats", "internal", "guard")); rc != 0 || out != "" {
			t.Errorf("live Go guards are not pinned (rc=%d):\n%s", rc, out)
		}
	})
}

// auditPy is the auditor: one scanner, three rules, hits printed as
// file:line lines; exit 1 on any hit, 2 on a bad root.
const auditPy = `
import re
import sys
from pathlib import Path

root = Path(sys.argv[1]).resolve()
if not root.is_dir():
    print(f"audit-git-config-pins: not a directory: {root}", file=sys.stderr)
    sys.exit(2)

targets = sorted(
    p
    for pat in ("*.sh", "*.py", "lib/*.sh", "lib/*.py", "*.go")
    for p in root.glob(pat)
    if not p.name.startswith("test-") and not p.name.endswith("_test.go")
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
    # Shell syntax rides along in a raw token — ` + "`" + `--quiet)"` + "`" + `, ` + "`" + `"--no-renames",` + "`" + `,
    # a Go composite literal's ` + "`" + `"--no-renames"}` + "`" + ` — so comparisons see the
    # flag, never the quoting around it.
    return [t.strip("\"'(){},;") for t in tokens]


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


def scan_go(path, hits):
    # A git argument list in Go is a call's string arguments: the subcommand
    # literal (never a map index like o["status"]) through the call's
    # closing parenthesis.
    text = path.read_text(errors="replace")
    for m in re.finditer(r'(?<![\w\[])"(diff|diff-tree|apply|status)"\s*[,)]', text):
        # A literal naming the program itself — lookPath(env, "diff"),
        # exec.Command("diff", …) — is diff(1), never a git subcommand. Only
        # lookPath takes a leading identifier: exec.Command(gitBin, "diff", …)
        # is a git diff and stays audited.
        if re.search(r'(?:exec\.Command\(|lookPath\(\s*\w+\s*,)\s*$', text[: m.start()]):
            continue
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
            elif ch in "\"` + "`" + `":
                quote = ch
            elif ch == "(":
                depth += 1
            elif ch == ")":
                if depth == 0:
                    end = i
                    break
                depth -= 1
            i += 1
        lineno = text[: m.start()].count("\n") + 1
        classify(f"{path}:{lineno}", m.group(1), text[m.start() : end], hits)


hits = []
for path in targets:
    if path.suffix == ".py":
        scan_python(path, hits)
    elif path.suffix == ".go":
        scan_go(path, hits)
    else:
        scan_shell(path, hits)

for h in hits:
    print(h)
sys.exit(1 if hits else 0)
`
