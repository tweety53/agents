# kan-844-agents-speed-up-scripts-test-setup-sh

> **Execution:** `/flow` implements this plan. Mark a task's own checkbox when
> `check-task-commit-fields.sh` passes on that task's commit.
> **Relocation:** yes — `scripts/test-setup.sh`'s header and per-group comments move into `stats/internal/setuptest/*_test.go`

Ports `scripts/test-setup.sh` to a Go test package whose groups run concurrently (tasks 1–5),
adds `xhigh` implementer effort (task 6), then measures and verifies live (task 7). `design.md` is
canonical for every decision; each task cites its entry by ID.

**The port's source of truth** is the bash harness at the merge base:
`git show 75410c1c:scripts/test-setup.sh`. Every group is found by its `group "<title>"` line;
each Go test ports exactly the bash between that line and the next `group` line (or the next
`# ====` banner) — every assertion, every seed, every comment that explains a case. A comment is
carried over as a Go comment, reworded only where it names bash mechanics that no longer exist.

**Assertion descriptions are byte-identical.** The Go description string equals the bash one after
the shell's own expansion — `"run $run succeeds"` becomes `fmt.Sprintf("run %d succeeds", run)`,
`${leftover#"$home"/}` becomes the path relative to that home. The per-group baseline is
`spectre/changes/kan-844-agents-speed-up-scripts-test-setup-sh/setup-harness-baseline.tsv`: one
line per assertion, `<group title>\t<description>`, captured from a full bash run at `75410c1c`
(580 lines).
<!-- measured: scripts/test-setup.sh > log; grep '^== \|^  ✓ ' log, grouped and sorted; wc -l setup-harness-baseline.tsv @ merge-base 75410c1c -->

**Parity check (used by tasks 2–5 and 7).** With `$TITLES` the `|`-joined group titles a task
ports and `$RUN` its `-run` regex:

```bash verified:ran the baseline half against setup-harness-baseline.tsv @ branch spectre/kan-844-agents-speed-up-scripts-test-setup-sh
B=spectre/changes/kan-844-agents-speed-up-scripts-test-setup-sh/setup-harness-baseline.tsv
T="$(mktemp -d)"
awk -F'\t' -v want="$TITLES|Nothing outside the sandbox was touched" \
  'BEGIN { n = split(want, w, "|"); for (i = 1; i <= n; i++) g[w[i]] = 1 } ($1 in g) { print $2 }' "$B" \
  | grep -v '/scripts/__pycache__ ' | LC_ALL=C sort > "$T/base"
(cd stats && go test ./internal/setuptest/ -run "$RUN" -count=1 -v) > "$T/go.log" 2>&1; echo "go test rc=$?"
grep -oE '✓ .*' "$T/go.log" | sed 's/^✓ //' | grep -v '^leak detection: ' \
  | grep -v '/scripts/__pycache__ ' | LC_ALL=C sort > "$T/go"
diff "$T/base" "$T/go" && echo "parity: $(wc -l < "$T/go") assertions"
```

`TestMain`'s close-out group runs on every invocation, so its title is always part of the slice.
Lines naming `/scripts/__pycache__ ` are dropped from both sides: an untracked `__pycache__` under
`skills/*/scripts/` appears and disappears with Python runs and would make the guard-reachability
list depend on tree state rather than on the port. `leak detection: ` lines are the new group's
(**Decision:** setuptest-leak-detection) and exist only on the Go side.

**Per-task verify** is the task's own `gofmt`/`go vet` plus its targeted tests and parity slice;
the full `## lint` list, `scripts/run-guard-tests.sh`, `cd stats && go test ./... -race -count=1`
and `cd stats/web && npm test` run in task 7 and `flow.verify`.

**Live verification:** task 7 — the change's runtime is the harness itself and the installer it
drives; it measures both harnesses on this machine and proves the containment checks trip. The
`xhigh` CLI change needs no live store: `dispatches.effort` is unconstrained `TEXT` (migration
`0020_dispatch_effort.sql`), and the CLI's acceptance is pinned against an `httptest` daemon.

## Review Focus

- A ported loop that iterates an empty set asserts nothing and still passes — the parity diff is
  the check; a group whose slice diffs by missing lines is not done.
- A group writing into the shared base fixture races every parallel reader — the close-out's
  `leak detection: the shared fixture repo is unchanged after every group` fingerprints it.
- A refusal weakened in the port (`runSetup` or `seedGuard` accepting a path outside the sandbox,
  or writing through a symlink) — task 1 carries both refusals verbatim.
- `KEEP_SANDBOX` ignored, or cleanup removing a path other than `/tmp/flow-test-setup.*` — task 1
  step 3.
- A Go description differing from the bash one by expansion (a `%d`, a trimmed prefix, a
  backtick) — the parity diff names it.

---

- [x] 1. setuptest: harness infrastructure, fingerprints, leak detection

**Files:** `stats/internal/setuptest/main_test.go`, `stats/internal/setuptest/helpers_test.go`, `stats/internal/setuptest/fingerprint_test.go`
**Tests:** `TestPycacheChurnLeavesFingerprintUnchanged`, `TestContainmentChecksDetectLeaks`
**Regression:** `TestPycacheChurnLeavesFingerprintUnchanged` fails if `treeFingerprint` stops
pruning `__pycache__` or stops seeing an ordinary file's mtime; `TestContainmentChecksDetectLeaks`
fails if either fingerprint stops moving when a sampled path is written.
**Baseline:** before=0 after=2
<!-- measured: cat stats/internal/setuptest/fingerprint_test.go 2>/dev/null | grep -cE '^func Test[A-Za-z0-9]+\(t \*testing\.T\)' @ branch spectre/kan-844-agents-speed-up-scripts-test-setup-sh -->
**After:** none
**Commit:** `test(setuptest): add the Go setup.sh harness core`
**Build:** green

**Decision:** setup-harness-go-parallel

**Decision:** setuptest-leak-detection

**Decision:** setuptest-in-stats-go

Correction (2026-09-28): panel round 0 fix commits `1ac33883`, `a2e74e9d` and `49f9d8e0`, on top of this
task's commit. The plan declared three files and two tests. What shipped adds `selfcheck_test.go`
with five tests that pin the harness's own safety logic: the `setupRefusal`/`seedRefusal`
refusals, `sandboxRemovable`, `closeOut`/`exitCode` and `countLinesMatching`, each extracted
so it can be tested. It also widens `TestContainmentChecksDetectLeaks` from a sample of the
locations to every sampled location, moves `appendTo` into `helpers_test.go`, bounds every installer
run by the test deadline in its own process group, and cleans the sandbox up on SIGINT/SIGTERM.
The installed guard run now gets a sandboxed `FLOW_GUARD_CACHE_DIR`, and `TestGuardsReachInstallAndRunFromIt`
asserts the guard built its binary there. Findings F1–F3, F5–F10 and
F12–F18. These fields still describe this task's own commit; the fix commits carry no `Task-Id:`.
<!-- measured: go test ./internal/setuptest/ -run '^TestContainmentChecksDetectLeaks$' -v | grep -c '✓ leak detection' → 27 (26 + the close-out line); five Test funcs in selfcheck_test.go @ branch spectre/kan-844-agents-speed-up-scripts-test-setup-sh -->

  - [x] **Step 1: `main_test.go`.** Package doc = the bash header's WHY THIS EXISTS and WHAT THIS
    CAN AND CANNOT PROVE paragraphs and the two-source-trees note, reworded only where they name
    bash (the portability-helper notes on BSD/GNU `stat` and a missing SHA tool are dropped: Go's
    standard library has neither problem). Then:

```go unverified:go vet ./internal/setuptest/ compiles it
package setuptest

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// kotlinRule is the opt-in rule two groups in different files assert on.
const kotlinRule = "kotlin-backend-development-standard.mdc"

var (
	repoRoot string // this checkout, the physical parent of stats/
	realHome string // $HOME as the test binary started
	sandbox  string // /tmp/flow-test-setup.*; every byte the harness writes lies under it
	begin    string // CLAUDE_MD_BEGIN, read out of setup.sh
	end      string // CLAUDE_MD_END, read out of setup.sh
	fixture  string // the shared base fixture repo — read-only for every group using it
)

func TestMain(m *testing.M) { os.Exit(runMain(m)) }

func runMain(m *testing.M) int {
	die := func(format string, a ...any) int {
		fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", a...)
		return 2
	}
	var err error
	if repoRoot, err = filepath.Abs("../../.."); err != nil {
		return die("cannot resolve the repo root: %v", err)
	}
	setupSh := filepath.Join(repoRoot, "setup.sh")
	if fi, err := os.Stat(setupSh); err != nil || fi.Mode().Perm()&0o111 == 0 {
		return die("setup.sh not found or not executable at %s", setupSh)
	}
	begin, end = readLiteral(setupSh, "CLAUDE_MD_BEGIN"), readLiteral(setupSh, "CLAUDE_MD_END")
	if begin == "" || end == "" || begin == end {
		return die("could not read the block delimiters out of setup.sh")
	}
	if realHome = os.Getenv("HOME"); realHome == "" {
		return die("HOME must be set")
	}
	// /tmp explicitly, never $TMPDIR (a per-user /var/folders path on macOS): the safety
	// claim is the concrete one — every byte written is under /tmp and removed again.
	if sandbox, err = os.MkdirTemp("/tmp", "flow-test-setup."); err != nil {
		return die("cannot create the sandbox under /tmp: %v", err)
	}
	defer cleanup()

	homeBefore := realHomeFingerprint(realHome)
	srcBefore := sourceTreeFingerprint(repoRoot)
	fixture = filepath.Join(sandbox, "fixture-repo")
	if err := makeFixtureRepo(fixture, false); err != nil {
		return die("%v", err)
	}
	fixtureBefore := treeFingerprint(fixture)

	fmt.Printf("setup.sh regression harness\n  repo    : %s\n  sandbox : %s\n", repoRoot, sandbox)
	code := m.Run()
	if closeOut(homeBefore, srcBefore, fixtureBefore) != 0 && code == 0 {
		code = 1
	}
	return code
}

// closeOut is the safety group: every test above ran with a sandboxed HOME; this proves it.
func closeOut(homeBefore, srcBefore, fixtureBefore string) int {
	fmt.Print("\n== Nothing outside the sandbox was touched ==\n")
	code := 0
	check := func(desc, want, got string) {
		if want == got {
			fmt.Printf("  ✓ %s\n", desc)
			return
		}
		code = 1
		fmt.Fprintf(os.Stderr, "  ✗ %s\n      expected [%s], got [%s]\n", desc, want, got)
	}
	check("~/.claude and ~/.zcode are unchanged", homeBefore, realHomeFingerprint(realHome))
	check("the repo's own skills, rules and commands are unchanged", srcBefore, sourceTreeFingerprint(repoRoot))
	check("leak detection: the shared fixture repo is unchanged after every group", fixtureBefore, treeFingerprint(fixture))
	return code
}

func cleanup() {
	if os.Getenv("KEEP_SANDBOX") != "" {
		fmt.Fprintf(os.Stderr, "KEEP_SANDBOX set — sandbox left at %s\n", sandbox)
		return
	}
	// Belt and braces: only ever remove a path this harness created under /tmp.
	if strings.HasPrefix(sandbox, "/tmp/flow-test-setup.") {
		_ = os.RemoveAll(sandbox)
	}
}

// readLiteral returns NAME's single-quoted value from the first `NAME='…'` line of setup.sh.
// Read, never restated: a stale copy would make the whole delimiter group pass vacuously.
func readLiteral(setupSh, name string) string {
	f, err := os.Open(setupSh)
	if err != nil {
		return ""
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if !strings.HasPrefix(line, name+"=") {
			continue
		}
		if i := strings.Index(line, "='"); i >= 0 {
			line = line[i+2:]
		}
		return strings.TrimSuffix(line, "'")
	}
	return ""
}
```

    `makeFixtureRepo(dir string, badRule bool) error` also lives here: a port of
    `make_fixture_repo`, every file byte-identical to what the bash `printf` wrote (copy
    `setup.sh` with mode `0755`; write each fixture file with `os.WriteFile`; the bad rule's body
    carries `begin`). It returns an error rather than calling `t.Fatal`, because `TestMain` calls
    it before any `*testing.T` exists.
  - [x] **Step 2: `helpers_test.go`.** The per-test state and the ported plumbing:

```go unverified:go vet ./internal/setuptest/ compiles it
package setuptest

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// group is one assertion group's private state: the bash harness's CASE_SEQ, HOME_DIR,
// RUN_RC and RUN_LOG globals, one copy per parallel test.
type group struct {
	t   *testing.T
	dir string // $SANDBOX/<TestName>; every path this group writes lies under it
	seq int
	rc  int    // the last runSetup's exit status
	log string // the last runSetup's combined-output log
}

// newGroup marks the test parallel and gives it its own sub-sandbox.
func newGroup(t *testing.T) *group {
	t.Helper()
	t.Parallel()
	dir := filepath.Join(sandbox, t.Name())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("cannot create the group sandbox %s: %v", dir, err)
	}
	return &group{t: t, dir: dir}
}

// newHome returns a fresh empty HOME for one case.
func (g *group) newHome() string {
	g.t.Helper()
	g.seq++
	home := filepath.Join(g.dir, fmt.Sprintf("home-%d", g.seq))
	if err := os.MkdirAll(home, 0o755); err != nil {
		g.t.Fatalf("cannot create sandbox home %s: %v", home, err)
	}
	return home
}

func inSandbox(p string) bool { return strings.HasPrefix(p, sandbox+"/") }

// runSetup invokes the installer against a sandboxed HOME; g.rc and g.log are published.
// The HOME and project arguments are re-checked here rather than trusted: this is the single
// point every case funnels through, so one check covers all of them. proj "" means the group's
// own cwd directory.
func (g *group) runSetup(repo, home, mode, proj string) {
	g.t.Helper()
	if proj == "" {
		proj = filepath.Join(g.dir, "cwd")
	}
	if !inSandbox(home) {
		g.t.Fatalf("refusing to run setup.sh with HOME=%s — not inside %s", home, sandbox)
	}
	if !inSandbox(proj) {
		g.t.Fatalf("refusing to run setup.sh with project dir %s — not inside %s", proj, sandbox)
	}
	mkdirAll(g.t, proj)
	g.log = filepath.Join(g.dir, fmt.Sprintf("log-%d-%s.txt", g.seq, mode))
	f, err := os.Create(g.log)
	if err != nil {
		g.t.Fatal(err)
	}
	defer f.Close()
	cmd := exec.Command(filepath.Join(repo, "setup.sh"), mode, proj)
	cmd.Dir = proj
	cmd.Env = withHome(os.Environ(), home)
	cmd.Stdout, cmd.Stderr = f, f
	g.rc = 0
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			g.t.Fatalf("cannot run setup.sh: %v", err)
		}
		g.rc = ee.ExitCode()
	}
}

func withHome(env []string, home string) []string {
	out := make([]string, 0, len(env)+1)
	for _, kv := range env {
		if !strings.HasPrefix(kv, "HOME=") {
			out = append(out, kv)
		}
	}
	return append(out, "HOME="+home)
}

// seedGuard refuses a path outside the sandbox, or one reached through a symlink at the path
// or at any parent inside the sandbox (the write-through incident the bash header records).
func (g *group) seedGuard(p string) {
	g.t.Helper()
	if !inSandbox(p) {
		g.t.Fatalf("refusing to seed %s — outside the sandbox %s", p, sandbox)
	}
	if isSymlink(p) {
		g.t.Fatalf("refusing to seed %s — it is a symlink, and writing through it would modify the link's target instead of the sandbox.", p)
	}
	for parent := filepath.Dir(p); inSandbox(parent); parent = filepath.Dir(parent) {
		if isSymlink(parent) {
			g.t.Fatalf("refusing to seed %s — its parent %s is a symlink, and writing under it would modify the link's target instead of the sandbox.", p, parent)
		}
	}
}

// seedFile creates p with content and an explicit mode. Content carries its own trailing
// newline, as the bash heredoc and here-string did.
func (g *group) seedFile(p string, mode os.FileMode, content string) {
	g.t.Helper()
	g.seedGuard(p)
	mkdirAll(g.t, filepath.Dir(p))
	if err := os.WriteFile(p, []byte(content), mode); err != nil {
		g.t.Fatal(err)
	}
	if err := os.Chmod(p, mode); err != nil { // WriteFile's mode is umask-masked
		g.t.Fatal(err)
	}
}

// seedStaleSkillDir puts a real skill directory where a symlink must go.
func (g *group) seedStaleSkillDir(p string) {
	g.t.Helper()
	g.seedGuard(p)
	mkdirAll(g.t, p)
	g.seedGuard(filepath.Join(p, "SKILL.md"))
	if err := os.WriteFile(filepath.Join(p, "SKILL.md"), []byte("# stale copy\nSENTINEL-PREEXISTING-SKILL\n"), 0o644); err != nil {
		g.t.Fatal(err)
	}
}

// seedProjectMD writes a .flow/project.md whose ## standards section lists entries, one
// backticked bullet each; the sections around it exercise the boundary parsing.
func (g *group) seedProjectMD(proj string, entries ...string) {
	g.t.Helper()
	var bullets strings.Builder
	for _, e := range entries {
		fmt.Fprintf(&bullets, "- `%s` — seeded by the harness\n", e)
	}
	g.seedFile(filepath.Join(proj, ".flow/project.md"), 0o644,
		"# flow project configuration — fixture project\n\n## test\n\n```bash\n./gradlew test\n```\n\n"+
			"## standards\n\nFiles the principles reviewer receives, and the rules this project opts into.\n\n"+
			bullets.String()+"\nA bullet inside a fenced block is illustration, not an entry:\n\n"+
			"```text\n- fenced-not-an-entry.mdc\n```\n\n## jira\n\nnone\n")
}

// --- assertions: every one names what it checked, byte-identical to the bash harness ---

func (g *group) pass(desc string)         { g.t.Helper(); g.t.Log("✓ " + desc) }
func (g *group) fail(desc, detail string) { g.t.Helper(); g.t.Errorf("✗ %s\n      %s", desc, detail) }

func (g *group) assertEq(desc string, want, got any) {
	g.t.Helper()
	if fmt.Sprint(want) == fmt.Sprint(got) {
		g.pass(desc)
	} else {
		g.fail(desc, fmt.Sprintf("expected [%v], got [%v]", want, got))
	}
}
func (g *group) assertNe(desc string, notWant, got any) {
	g.t.Helper()
	if fmt.Sprint(notWant) != fmt.Sprint(got) {
		g.pass(desc)
	} else {
		g.fail(desc, fmt.Sprintf("expected something other than [%v], got [%v]", notWant, got))
	}
}
func (g *group) assertExists(desc, p string) {
	g.t.Helper()
	if lexists(p) {
		g.pass(desc)
	} else {
		g.fail(desc, p+" does not exist")
	}
}

// assertFileExists requires a regular FILE (following links, as `[[ -f ]]` does).
func (g *group) assertFileExists(desc, p string) {
	g.t.Helper()
	if fi, err := os.Stat(p); err == nil && fi.Mode().IsRegular() {
		g.pass(desc)
	} else {
		g.fail(desc, p+" is not a regular file")
	}
}
func (g *group) assertAbsent(desc, p string) {
	g.t.Helper()
	if lexists(p) {
		g.fail(desc, p+" exists and must not")
	} else {
		g.pass(desc)
	}
}
func (g *group) assertSymlink(desc, p string) {
	g.t.Helper()
	if isSymlink(p) {
		g.pass(desc)
	} else {
		g.fail(desc, p+" is not a symlink")
	}
}
func (g *group) assertExecutable(desc, p string) {
	g.t.Helper()
	if executable(p) {
		g.pass(desc)
	} else {
		g.fail(desc, p+" is not executable")
	}
}
func (g *group) assertIdentical(desc, a, b string) {
	g.t.Helper()
	x, errA := os.ReadFile(a)
	y, errB := os.ReadFile(b)
	if errA == nil && errB == nil && bytes.Equal(x, y) {
		g.pass(desc)
	} else {
		g.fail(desc, a+" and "+b+" differ")
	}
}
func (g *group) assertContains(desc, file, s string) {
	g.t.Helper()
	if b, err := os.ReadFile(file); err == nil && bytes.Contains(b, []byte(s)) {
		g.pass(desc)
	} else {
		g.fail(desc, fmt.Sprintf("[%s] not found in %s", s, file))
	}
}
func (g *group) assertNotContains(desc, file, s string) {
	g.t.Helper()
	if b, err := os.ReadFile(file); err != nil || !bytes.Contains(b, []byte(s)) {
		g.pass(desc)
	} else {
		g.fail(desc, fmt.Sprintf("[%s] found in %s and must not be", s, file))
	}
}
func (g *group) assertRCNonzero(desc string, rc int) {
	g.t.Helper()
	if rc != 0 {
		g.pass(desc)
	} else {
		g.fail(desc, "exit status was 0; the run should have aborted")
	}
}
func (g *group) assertRCZero(desc string, rc int, log string) {
	g.t.Helper()
	if rc == 0 {
		g.pass(desc)
	} else {
		g.fail(desc, fmt.Sprintf("exit status %d — see %s", rc, log))
	}
}

// --- filesystem helpers ---

// countLinesMatching counts the lines of file equal to line (grep -cFx); 0 when unreadable.
func countLinesMatching(file, line string) int {
	b, err := os.ReadFile(file)
	if err != nil {
		return 0
	}
	n := 0
	for _, l := range strings.Split(strings.TrimSuffix(string(b), "\n"), "\n") {
		if l == line {
			n++
		}
	}
	return n
}

// fileMode is the permission bits in octal, e.g. "640".
func fileMode(p string) string {
	fi, err := os.Stat(p)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%o", fi.Mode().Perm())
}

func lexists(p string) bool { _, err := os.Lstat(p); return err == nil }

func isSymlink(p string) bool {
	fi, err := os.Lstat(p)
	return err == nil && fi.Mode()&os.ModeSymlink != 0
}

// executable follows links and asks the kernel, as `[[ -x ]]` does.
func executable(p string) bool { return syscall.Access(p, 0x1) == nil }

func mkdirAll(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
}

// copyFile is `cp -p`: content, permission bits and modification time.
func copyFile(t *testing.T, src, dst string) {
	t.Helper()
	b, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	fi, err := os.Stat(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, b, fi.Mode().Perm()); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dst, fi.Mode().Perm()); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(dst, time.Now(), fi.ModTime()); err != nil {
		t.Fatal(err)
	}
}

// findFollow lists every path under root named name, following symlinks (`find -L`); a
// directory already on the current path is not re-entered, as find -L refuses a loop.
func findFollow(root, name string) []string {
	var out []string
	var walk func(p string, ancestors map[string]bool)
	walk = func(p string, ancestors map[string]bool) {
		resolved, err := filepath.EvalSymlinks(p)
		if err != nil {
			return
		}
		fi, err := os.Stat(resolved)
		if err != nil {
			return
		}
		if filepath.Base(p) == name {
			out = append(out, p)
		}
		if !fi.IsDir() || ancestors[resolved] {
			return
		}
		ancestors[resolved] = true
		defer delete(ancestors, resolved)
		entries, _ := os.ReadDir(resolved)
		for _, e := range entries {
			walk(filepath.Join(p, e.Name()), ancestors)
		}
	}
	walk(root, map[string]bool{})
	return out
}
```

    Any further helper a later task finds it needs (a cheaper form of one above, a shared seed)
    is added in that task's own file, never here, so tasks 2–5 stay file-disjoint.
  - [x] **Step 3: `fingerprint_test.go`.** The two containment fingerprints, parameterised by root,
    carrying the bash comments that explain what is sampled and why:

```go unverified:go vet ./internal/setuptest/ compiles it
package setuptest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

// statLine is `stat` without -L: name, size, mtime seconds, permission bits.
func statLine(p string) string {
	fi, err := os.Lstat(p)
	if err != nil {
		return "absent " + p + "\n"
	}
	return fmt.Sprintf("%s %d %d %o\n", p, fi.Size(), fi.ModTime().Unix(), fi.Mode().Perm())
}

func sha256hex(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }

// realHomeFingerprint samples only the paths setup.sh can write under home.
func realHomeFingerprint(home string) string {
	var b strings.Builder
	for _, d := range []string{".claude", ".zcode"} {
		for _, sub := range []string{"", "skills", "commands", "rules"} {
			p := filepath.Join(home, d, sub)
			if fi, err := os.Stat(p); err != nil || !fi.IsDir() {
				fmt.Fprintf(&b, "absent %s\n", p)
				continue
			}
			fmt.Fprintf(&b, "dir %s\n", p)
			entries, _ := os.ReadDir(p)
			for _, e := range entries {
				fmt.Fprintf(&b, "  %s\n", e.Name())
			}
		}
		for _, f := range []string{"CLAUDE.md", "AGENTS.md", "CLAUDE.md.flow.bak", "AGENTS.md.flow.bak"} {
			b.WriteString(statLine(filepath.Join(home, d, f)))
		}
	}
	for _, f := range []string{".zshrc", ".bashrc"} {
		b.WriteString(statLine(filepath.Join(home, f)))
	}
	return sha256hex(b.String())
}

// treeFingerprint stat-lines every regular file under paths, sorted, skipping any directory
// named __pycache__ (generated, gitignored bytecode — not source).
func treeFingerprint(paths ...string) string {
	var files []string
	for _, root := range paths {
		_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() && d.Name() == "__pycache__" {
				return filepath.SkipDir
			}
			if d.Type().IsRegular() {
				files = append(files, p)
			}
			return nil
		})
	}
	sort.Strings(files)
	var b strings.Builder
	for _, f := range files {
		b.WriteString(statLine(f))
	}
	return sha256hex(b.String())
}

// sourceTreeFingerprint includes scripts/ and README.md deliberately: without them the harness
// cannot detect damage to itself.
func sourceTreeFingerprint(root string) string {
	var ps []string
	for _, p := range []string{"skills", "rules", "commands-claude", "scripts", "setup.sh", "CLAUDE.md", "AGENTS.md", "README.md"} {
		ps = append(ps, filepath.Join(root, p))
	}
	return treeFingerprint(ps...)
}

func TestPycacheChurnLeavesFingerprintUnchanged(t *testing.T) {
	g := newGroup(t)
	fx := filepath.Join(g.dir, "pycache-fixture")
	mkdirAll(t, filepath.Join(fx, "scripts/__pycache__"))
	mkdirAll(t, filepath.Join(fx, "skills"))
	src := filepath.Join(fx, "scripts/real.py")
	pyc := filepath.Join(fx, "scripts/__pycache__/real.cpython-314.pyc")
	g.seedFile(src, 0o644, "x\n")
	g.seedFile(pyc, 0o644, "y\n")
	fp := func() string { return treeFingerprint(filepath.Join(fx, "scripts"), filepath.Join(fx, "skills")) }
	before := fp()
	// A fixed future mtime guarantees a changed timestamp even inside one mtime tick.
	future := time.Date(2030, 1, 1, 0, 0, 0, 0, time.Local)
	if err := os.Chtimes(pyc, future, future); err != nil {
		t.Fatal(err)
	}
	g.assertEq("a .pyc rewrite under __pycache__ leaves the fingerprint unchanged", before, fp())
	if err := os.Chtimes(src, future, future); err != nil {
		t.Fatal(err)
	}
	g.assertNe("a real source-file change still moves the fingerprint", before, fp())
}

// TestContainmentChecksDetectLeaks proves both containment fingerprints still trip on a leak,
// on fake roots under the sandbox — never by writing to the real HOME. Every write changes a
// size or adds an entry, so no case depends on the filesystem's mtime resolution.
func TestContainmentChecksDetectLeaks(t *testing.T) {
	g := newGroup(t)
	home := filepath.Join(g.dir, "fake-home")
	mkdirAll(t, filepath.Join(home, ".claude/skills"))
	g.seedFile(filepath.Join(home, ".claude/CLAUDE.md"), 0o644, "# mine\n")
	g.seedFile(filepath.Join(home, ".zshrc"), 0o644, "# mine\n")

	before := realHomeFingerprint(home)
	g.seedFile(filepath.Join(home, ".claude/skills/leaked-skill"), 0o644, "leak\n")
	g.assertNe("leak detection: a new entry under ~/.claude/skills moves the real-home fingerprint", before, realHomeFingerprint(home))

	before = realHomeFingerprint(home)
	appendTo(t, filepath.Join(home, ".claude/CLAUDE.md"), "leak\n")
	g.assertNe("leak detection: a rewrite of ~/.claude/CLAUDE.md moves the real-home fingerprint", before, realHomeFingerprint(home))

	before = realHomeFingerprint(home)
	appendTo(t, filepath.Join(home, ".zshrc"), "leak\n")
	g.assertNe("leak detection: a rewrite of ~/.zshrc moves the real-home fingerprint", before, realHomeFingerprint(home))

	src := filepath.Join(g.dir, "fake-repo")
	g.seedFile(filepath.Join(src, "skills/demo/SKILL.md"), 0o644, "# demo\n")
	g.seedFile(filepath.Join(src, "setup.sh"), 0o755, "#!/usr/bin/env bash\n")
	before = sourceTreeFingerprint(src)
	g.seedFile(filepath.Join(src, "skills/demo/leaked.md"), 0o644, "leak\n")
	g.assertNe("leak detection: a new file under skills/ moves the source-tree fingerprint", before, sourceTreeFingerprint(src))
}

func appendTo(t *testing.T, p, s string) {
	t.Helper()
	f, err := os.OpenFile(p, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(s); err != nil {
		t.Fatal(err)
	}
}
```

  - [x] **Step 4: Run, expect green.** `cd stats && gofmt -l internal/setuptest && go vet
    ./internal/setuptest/ && go test ./internal/setuptest/ -count=1 -v 2>&1 | grep -E '✓|✗|^(ok|FAIL|---)'`
    — both tests PASS, the three close-out lines print `✓`, `gofmt -l` prints nothing.
  - [x] **Step 5: Prove the leak group can fail.** Temporarily make `realHomeFingerprint` return
    a constant; re-run `-run '^TestContainmentChecksDetectLeaks$'` — expect three `✗` lines;
    restore it. Temporarily delete the `__pycache__` `SkipDir` branch; re-run
    `-run '^TestPycacheChurnLeavesFingerprintUnchanged$'` — expect its first assertion to fail;
    restore it. Commit only the restored files.

- [x] 2. setuptest: delimiter, convergence and all-or-nothing groups

**Files:** `stats/internal/setuptest/delimiters_test.go`
**Tests:** `TestMalformedDelimitersAbortByteIdentical`, `TestThreeGlobalRunsConverge`, `TestHandWrittenContentAndModeSurvive`, `TestDelimiterRuleInstallsNothing`
**Regression:** each fails if `setup.sh` regresses on the shape its bash group guards — a
malformed marker pair no longer aborting byte-identically, a re-run adding a block, a hand-written
line or file mode lost, or a delimiter-carrying rule installing anything.
**Baseline:** before=0 after=4
<!-- measured: cat stats/internal/setuptest/delimiters_test.go 2>/dev/null | grep -cE '^func Test[A-Za-z0-9]+\(t \*testing\.T\)' @ branch spectre/kan-844-agents-speed-up-scripts-test-setup-sh -->
**After:** Task 1
**Commit:** `test(setuptest): port the delimiter and convergence groups`
**Build:** green

**Decision:** setup-harness-go-parallel

Correction (2026-09-28): panel round 0 fix commit `1ac33883`, on top of this task's commit,
adds `TestUnbalancedDelimitersAbortByteIdentical`, covering a lone begin and a lone end. The bash
harness never seeded that shape, so the package doc's defect 2 could come back undetected (F11).
Its six assertions are the only Go-side lines the parity check reports beyond the baseline.
<!-- measured: the tasks.md parity check with every group title, RUN='.' → diff of 6 added lines, all `lone begin:`/`lone end:` @ branch spectre/kan-844-agents-speed-up-scripts-test-setup-sh -->
It also corrects the 640-mode rationale comment (F19).

  - [x] **Step 1: Port**, one test per bash group, each opening `g := newGroup(t)`, each case
    opening `home := g.newHome()`, every `run_setup "$FIXTURE" …` a `g.runSetup(fixture, home,
    "global", "")`, every `cp -p` a `copyFile` into `g.dir`, every `$SANDBOX/<name>` a
    `filepath.Join(g.dir, "<name>")`:
    - `TestMalformedDelimitersAbortByteIdentical` ← `group "Malformed delimiters abort with the
      target byte-identical"` (the reversed, duplicated and CRLF cases).
    - `TestThreeGlobalRunsConverge` ← `group "Three consecutive global runs converge"`:

```go unverified:go vet ./internal/setuptest/ compiles it
func TestThreeGlobalRunsConverge(t *testing.T) {
	g := newGroup(t)
	home := g.newHome()
	claudeMD := filepath.Join(home, ".claude/CLAUDE.md")
	zcodeMD := filepath.Join(home, ".zcode/AGENTS.md")
	claude2 := filepath.Join(g.dir, "converge-claude-run2.md")
	zcode2 := filepath.Join(g.dir, "converge-zcode-run2.md")
	for run := 1; run <= 3; run++ {
		g.runSetup(fixture, home, "global", "")
		g.assertRCZero(fmt.Sprintf("run %d succeeds", run), g.rc, g.log)
		g.assertEq(fmt.Sprintf("run %d: exactly one begin in CLAUDE.md", run), 1, countLinesMatching(claudeMD, begin))
		g.assertEq(fmt.Sprintf("run %d: exactly one end in CLAUDE.md", run), 1, countLinesMatching(claudeMD, end))
		g.assertEq(fmt.Sprintf("run %d: exactly one begin in AGENTS.md", run), 1, countLinesMatching(zcodeMD, begin))
		g.assertEq(fmt.Sprintf("run %d: exactly one end in AGENTS.md", run), 1, countLinesMatching(zcodeMD, end))
		if run == 2 {
			copyFile(t, claudeMD, claude2)
			copyFile(t, zcodeMD, zcode2)
		}
	}
	g.assertIdentical("run 3 leaves CLAUDE.md byte-identical to run 2", claudeMD, claude2)
	g.assertIdentical("run 3 leaves AGENTS.md byte-identical to run 2", zcodeMD, zcode2)
}
```

    - `TestHandWrittenContentAndModeSurvive` ← `group "Hand-written content and file mode survive
      every run"` (seed mode `0o640`; the `HANDWRITTEN-BELOW` append via task 1's
      `appendTo`).
    - `TestDelimiterRuleInstallsNothing` ← `group "A rule carrying a delimiter installs nothing at
      all"`, on its own `makeFixtureRepo(filepath.Join(g.dir, "fixture-repo-bad-rule"), true)`.
  - [x] **Step 2: Verify.** `cd stats && gofmt -l internal/setuptest && go vet
    ./internal/setuptest/`; then the preamble's parity check with
    `TITLES='Malformed delimiters abort with the target byte-identical|Three consecutive global runs converge|Hand-written content and file mode survive every run|A rule carrying a delimiter installs nothing at all'`
    and `RUN='^(TestMalformedDelimitersAbortByteIdentical|TestThreeGlobalRunsConverge|TestHandWrittenContentAndModeSurvive|TestDelimiterRuleInstallsNothing)$'`
    — `go test rc=0`, empty diff.

- [x] 3. setuptest: global containment groups

**Files:** `stats/internal/setuptest/containment_test.go`
**Tests:** `TestOnlyAlwaysApplyRulesInstall`, `TestCoreExcerptsLinksAndBaseline`, `TestKotlinRuleInstalledByNoMode`, `TestGlobalInstallPopulatesSkillDirs`, `TestGuardsReachInstallAndRunFromIt`, `TestPreexistingSkillDirMovedOut`
**Regression:** each fails if a containment guarantee its bash group guards regresses — an opt-in
or malformed rule installed, a block inlining a full rule, the Kotlin rule installed by any mode,
a skill link missing or dangling, a guard missing or not executable at its installed path, a
stale skill copy left reachable.
**Baseline:** before=0 after=6
<!-- measured: cat stats/internal/setuptest/containment_test.go 2>/dev/null | grep -cE '^func Test[A-Za-z0-9]+\(t \*testing\.T\)' @ branch spectre/kan-844-agents-speed-up-scripts-test-setup-sh -->
**After:** Task 1
**Commit:** `test(setuptest): port the global containment groups`
**Build:** green

**Decision:** setup-harness-go-parallel

  - [x] **Step 1: Port**, conventions as task 2 step 1:
    - `TestOnlyAlwaysApplyRulesInstall` ← `group "Only rules declaring alwaysApply: true
      install"` (the rule count: every `os.ReadDir` entry of `.claude/rules` but
      `agent-baseline.md`; `assertEq(…, 2, n)`).
    - `TestCoreExcerptsLinksAndBaseline` ← `group "Core excerpts, full-text links and the agent
      baseline"`; its unbalanced-core case builds its own fixture under `g.dir`.
    - `TestKotlinRuleInstalledByNoMode` ← `group "The opt-in Kotlin rule is installed by no mode"`
      (`kotlinRule` from task 1; `hits` = every path under `home` and `proj` whose base name is
      `kotlinRule`, walked with `filepath.WalkDir` — no link following, as `find` without `-L` —
      joined by `"\n"`; a missing rule file is `g.fail("the opt-in Kotlin rule is present to be
      tested", …+" is missing")`).
    - `TestGlobalInstallPopulatesSkillDirs` ← `group "A global install populates both skill
      directories with live links"` (expected skills = the non-dot entries of `repoRoot/skills`
      that `os.Stat` as directories; `linked` = `isSymlink` and `os.Stat` a directory; `dangling`
      = every symlink found by `filepath.WalkDir(home)` whose `os.Stat` fails, joined by `"\n"`).
    - `TestGuardsReachInstallAndRunFromIt` ← `group "Every guard in a command skill's scripts/
      directory reaches the install"` **followed in the same test by** `group
      "check-panel-reproducers.sh runs through the installed path and finds its dependencies"` —
      the second runs the guard out of the first's HOME (**Decision:**
      setup-harness-go-parallel). Skills = the non-dot `repoRoot/skills/*` whose `scripts` is a
      directory; entries = the non-dot names in each `scripts/`; `want_exec` = `os.Stat` regular
      and `executable`. The guard runs as `exec.Command(guardPath, g.dir)` with the inherited
      environment, output to `filepath.Join(g.dir, "check-panel-reproducers-installed.log")`.
    - `TestPreexistingSkillDirMovedOut` ← `group "A pre-existing skill directory is moved outside
      the scanned tree"` (victim = the first non-dot directory entry of `repoRoot/skills`;
      both hit counts = the files `findFollow(root, "SKILL.md")` returns that contain
      `SENTINEL-PREEXISTING-SKILL`).
  - [x] **Step 2: Verify.** `cd stats && gofmt -l internal/setuptest && go vet
    ./internal/setuptest/`; then the parity check with
    `TITLES="Only rules declaring alwaysApply: true install|Core excerpts, full-text links and the agent baseline|The opt-in Kotlin rule is installed by no mode|A global install populates both skill directories with live links|Every guard in a command skill's scripts/ directory reaches the install|check-panel-reproducers.sh runs through the installed path and finds its dependencies|A pre-existing skill directory is moved outside the scanned tree"`
    and `RUN='^(TestOnlyAlwaysApplyRulesInstall|TestCoreExcerptsLinksAndBaseline|TestKotlinRuleInstalledByNoMode|TestGlobalInstallPopulatesSkillDirs|TestGuardsReachInstallAndRunFromIt|TestPreexistingSkillDirMovedOut)$'`
    — `go test rc=0`, empty diff.

- [x] 4. setuptest: project rendering groups

**Files:** `stats/internal/setuptest/project_test.go`
**Tests:** `TestProjectOptInRuleRendered`, `TestProjectRenderingIdempotent`, `TestModesRenderingProjectStandards`, `TestProjectWithNothingToRenderLeftAlone`, `TestMissingNamedRuleReportedAndSkipped`, `TestProjectDelimiterGuardsFire`, `TestRealKotlinStandardReachesProject`
**Regression:** each fails if project-standards rendering regresses on the shape its bash group
guards — an opted-in rule not rendered into both files, a re-run changing them, a mode rendering
when it must not, an untouched project touched, a missing rule not reported, a reversed marker
pair not refused, the real Kotlin standard not reaching a project.
**Baseline:** before=0 after=7
<!-- measured: cat stats/internal/setuptest/project_test.go 2>/dev/null | grep -cE '^func Test[A-Za-z0-9]+\(t \*testing\.T\)' @ branch spectre/kan-844-agents-speed-up-scripts-test-setup-sh -->
**After:** Task 1
**Commit:** `test(setuptest): port the project rendering groups`
**Build:** green

**Decision:** setup-harness-go-parallel

  - [x] **Step 1: Port**, conventions as task 2 step 1, every `seed_project_md` a
    `g.seedProjectMD`, every project dir `filepath.Join(g.dir, fmt.Sprintf("project-<kind>-%d",
    g.seq))` right after the case's `newHome`:
    - `TestProjectOptInRuleRendered` ← `group "A project's opted-in shared rule is rendered into
      both its instruction files"`.
    - `TestProjectRenderingIdempotent` ← `group "Project rendering is idempotent and preserves
      hand-written content"`.
    - `TestModesRenderingProjectStandards` ← `group "Which modes render project standards"`.
    - `TestProjectWithNothingToRenderLeftAlone` ← `group "A project with nothing to render is
      left alone, silently"` (`seed_instruction_files` becomes a closure in the test).
    - `TestMissingNamedRuleReportedAndSkipped` ← `group "A named rule that does not exist is
      reported and skipped"`.
    - `TestProjectDelimiterGuardsFire` ← `group "The delimiter guards fire on a project file
      exactly as on a global one"`.
    - `TestRealKotlinStandardReachesProject` ← `group "The real Kotlin standard reaches a project
      that opts into it"` (`kotlinRule` from task 1).
  - [x] **Step 2: Verify.** `cd stats && gofmt -l internal/setuptest && go vet
    ./internal/setuptest/`; then the parity check with
    `TITLES="A project's opted-in shared rule is rendered into both its instruction files|Project rendering is idempotent and preserves hand-written content|Which modes render project standards|A project with nothing to render is left alone, silently|A named rule that does not exist is reported and skipped|The delimiter guards fire on a project file exactly as on a global one|The real Kotlin standard reaches a project that opts into it"`
    and `RUN='^(TestProjectOptInRuleRendered|TestProjectRenderingIdempotent|TestModesRenderingProjectStandards|TestProjectWithNothingToRenderLeftAlone|TestMissingNamedRuleReportedAndSkipped|TestProjectDelimiterGuardsFire|TestRealKotlinStandardReachesProject)$'`
    — `go test rc=0`, empty diff.

- [x] 5. setuptest: prune, retired-harness and zcode groups; test-setup.sh runs the Go harness

**Files:** `stats/internal/setuptest/harnesses_test.go`, `scripts/test-setup.sh`, `stats/internal/guard/installedcitations.go`
**Tests:** `TestReinstallPrunesDeletedSourceLinks`, `TestGlobalWritesNothingUnderCursorCodex`, `TestRetiredModesRefused`, `TestZcodeGlobalSelfContained`, `TestZcodeProjectModeProjectPathsOnly`, `TestZcodeCompactWindowEnvBlock`
**Regression:** each fails if the installer regresses on the shape its bash group guards — a stale
link kept or a user's own file pruned, anything written under `~/.cursor`/`~/.codex`, a retired
mode accepted, the zcode layer pointing outside `~/.zcode` or duplicating on re-run, the env block
duplicated or an unmanaged export overwritten. Reverting the shim restores the bash harness,
which still passes.
**Baseline:** before=0 after=6
<!-- measured: cat stats/internal/setuptest/harnesses_test.go 2>/dev/null | grep -cE '^func Test[A-Za-z0-9]+\(t \*testing\.T\)' @ branch spectre/kan-844-agents-speed-up-scripts-test-setup-sh -->
**After:** Task 1, 2, 3, 4
**Commit:** `test(setuptest): port the last groups and run test-setup.sh through them`
**Build:** green

**Decision:** setup-harness-go-parallel

  - [x] **Step 1: Port**, conventions as task 2 step 1:
    - `TestReinstallPrunesDeletedSourceLinks` ← `group "A re-install prunes links whose source
      was deleted"`, on its own `makeFixtureRepo(filepath.Join(g.dir, "fixture-repo-prune"),
      false)`; the user's broken link targets `filepath.Join(g.dir,
      "not-mounted-right-now/thing.md")`. The bash harness left `FIXTURE` pointing here for every
      later group; the groups below use the shared `fixture` instead — identical content
      (**Decision:** setup-harness-go-parallel).
    - `TestGlobalWritesNothingUnderCursorCodex` ← `group "global install writes nothing under
      .cursor or .codex"` (`old_harness_fp` = the sorted `filepath.WalkDir` paths of both
      directories plus `treeFingerprint` of both, hashed with `sha256hex`).
    - `TestRetiredModesRefused` ← `group "The retired cursor, codex and all modes are refused"`
      (project dir `filepath.Join(g.dir, "retired-"+mode)`; `ls -A` = the joined `os.ReadDir`
      names).
    - `TestZcodeGlobalSelfContained` ← `group "ZCode: global installs a self-contained ~/.zcode
      layer"`.
    - `TestZcodeProjectModeProjectPathsOnly` ← `group "ZCode: the per-project mode writes project
      paths only"`.
    - `TestZcodeCompactWindowEnvBlock` ← `group "ZCode: the compact-window env block in the shell
      rc"`.
  - [x] **Step 2: Verify.** `cd stats && gofmt -l internal/setuptest && go vet
    ./internal/setuptest/`; then the parity check with
    `TITLES='A re-install prunes links whose source was deleted|global install writes nothing under .cursor or .codex|The retired cursor, codex and all modes are refused|ZCode: global installs a self-contained ~/.zcode layer|ZCode: the per-project mode writes project paths only|ZCode: the compact-window env block in the shell rc'`
    and `RUN='^(TestReinstallPrunesDeletedSourceLinks|TestGlobalWritesNothingUnderCursorCodex|TestRetiredModesRefused|TestZcodeGlobalSelfContained|TestZcodeProjectModeProjectPathsOnly|TestZcodeCompactWindowEnvBlock)$'`
    — `go test rc=0`, empty diff.
  - [x] **Step 3: Full parity, before the swap.** The preamble's parity check with every group
    title: `TITLES="$(cut -f1 spectre/changes/kan-844-agents-speed-up-scripts-test-setup-sh/setup-harness-baseline.tsv | sort -u | paste -sd'|' -)"`
    and `RUN='.'` — `go test rc=0`, empty diff, `parity: 580 assertions`; and
    `grep -c '✓ leak detection: ' "$T/go.log"` prints 5.
    <!-- predicted: 580 = the baseline's line count, less any /scripts/__pycache__ line present in neither; 5 = four TestContainmentChecksDetectLeaks lines plus the close-out's fixture line -->
  - [x] **Step 4: Shim.** Replace `scripts/test-setup.sh` whole:

```bash verified:scripts/test-go-guards.sh @ branch spectre/kan-844-agents-speed-up-scripts-test-setup-sh is the same shape
#!/usr/bin/env bash
# test-setup.sh — regression harness for setup.sh.
#
# The harness is the Go test package stats/internal/setuptest/: its assertion
# groups run concurrently, each against its own sandboxed HOME under
# /tmp/flow-test-setup.*, and the real ~/.claude, ~/.zcode and the source tree
# are fingerprinted around all of them. Its package doc says what a green run
# does and does not prove. This shim keeps the one name run-guard-tests.sh's
# test-*.sh glob and every caller use.
#
# Usage: scripts/test-setup.sh
# Set KEEP_SANDBOX=1 to leave the sandbox behind for inspection.
# Exit codes are `go test`'s: 0 every assertion passed, non-zero otherwise.
set -euo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/../stats" && exec go test ./internal/setuptest/ -count=1
```

  - [x] **Step 5: Repoint the one citation of the bash internals.**
    `stats/internal/guard/installedcitations.go`'s comment "adopting scripts/test-setup.sh's own
    refusal" names a refusal that now lives in `stats/internal/setuptest/helpers_test.go`'s
    `runSetup`; repoint it there. The other prose mentions of `test-setup.sh`
    (`scripts/check-model-resolution-shell.sh`, `scripts/test-check-model-resolution-shell.sh`)
    name its containment case, which the shim still runs — unchanged.
  - [x] **Step 6: Verify.** `scripts/test-setup.sh; echo rc=$?` — `rc=0`; `KEEP_SANDBOX=1
    scripts/test-setup.sh` leaves one `/tmp/flow-test-setup.*` directory (remove it after
    looking); `scripts/check-references.sh`; `scripts/check-installed-citations.sh`; `cd stats &&
    gofmt -l . && go vet ./internal/guard/`.

- [x] 6. flow: xhigh implementer effort

**Files:** `agents/flow-xhigh.md`, `stats/internal/guard/installedcitations.go`, `scripts/test-setup-agents.sh`, `stats/cmd/flow/record.go`, `stats/cmd/flow/record_test.go`, `stats/internal/records/types.go`, `skills/flow/brainstorm-planner.md`, `skills/flow/implement.md`, `README.md`
**Tests:** `TestRecordDispatchBeginAcceptsEffort` — extended, not added: its table gains an `xhigh`
row; TestRunRecordDispatchBeginRejectsUnknownEffort's accepted-word loop and
scripts/test-setup-agents.sh's effort list gain `xhigh` too.
**Regression:** `TestRecordDispatchBeginAcceptsEffort` — reverting `record.go` fails its xhigh row;
removing `agents/flow-xhigh.md` fails `scripts/test-setup-agents.sh`; dropping its allowlist entry
fails `scripts/check-installed-citations.sh`.
**Baseline:** before=91 after=91
<!-- measured: cat stats/cmd/flow/record_test.go | grep -cE '^func Test' @ branch spectre/kan-844-agents-speed-up-scripts-test-setup-sh -->
**After:** Task 5
**Commit:** `feat(flow): allow xhigh effort for implementers`
**Build:** green

**Decision:** xhigh-implementers-only

  - [x] **Step 1: Failing tests.** In `stats/cmd/flow/record_test.go`, add `"xhigh"` to
    TestRecordDispatchBeginAcceptsEffort's word list (and its doc comment: four accepted
    non-default words) and to TestRunRecordDispatchBeginRejectsUnknownEffort's accepted-word loop.
    In `scripts/test-setup-agents.sh`, `EFFORTS="low medium high xhigh"`, both counts 3 → 4, the
    `three` wording → `four`. Run `cd stats && go test ./cmd/flow/ -run
    '^(TestRecordDispatchBeginAcceptsEffort|TestRunRecordDispatchBeginRejectsUnknownEffort)$'
    -count=1` and `scripts/test-setup-agents.sh` — expect both to fail.
  - [x] **Step 2: The agent definition.** `agents/flow-xhigh.md`:

```markdown verified:agents/flow-high.md @ branch spectre/kan-844-agents-speed-up-scripts-test-setup-sh with effort xhigh; Claude Code sub-agents docs list xhigh as a frontmatter effort
---
name: flow-xhigh
description: Flow role at xhigh effort — an implementer group carrying the change's hardest seam; the dispatch prompt carries every instruction; pass `model` at dispatch time.
effort: xhigh
tools: Read, Glob, Grep, Bash, Write, Edit, ToolSearch
---
General-purpose flow role at xhigh effort; the dispatch prompt carries every instruction.
```

    and its entry in `stats/internal/guard/installedcitations.go`'s allowlist, after
    `agents/flow-high.md`'s: `{"agents/flow-xhigh.md", "generic dispatch-target agent definition
    — cites no .md/.mdc path at all"},`. Without the entry the guard exits 1.
    <!-- measured: scripts/check-installed-citations.sh with a copy of flow-high.md as agents/flow-xhigh.md, no allowlist entry → rc=1 "agents/flow-xhigh.md:0: 0 checked, and not declared expected-zero" @ branch spectre/kan-844-agents-speed-up-scripts-test-setup-sh -->
  - [x] **Step 3: The CLI.** `recordEfforts = []string{"low", "medium", "high", "xhigh",
    "default"}`, its comment "the four efforts a flow dispatch can carry"; the `Dispatch.Effort`
    doc comment in `stats/internal/records/types.go` lists `xhigh`. No migration:
    `dispatches.effort` is unconstrained `TEXT`.
  - [x] **Step 4: The prose.** `skills/flow/brainstorm-planner.md` **Model and effort**: the
    sentence "every other `effort` is the planner's own choice, one of `low`/`medium`/`high`"
    becomes "…one of `low`/`medium`/`high` — and, for the implementer pair and an implementer group
    alone, `xhigh` —"; the fixer, panel dispatches and rerun pair keep the three. `skills/flow/implement.md`'s
    `flow-<effort>` family paragraph names `agents/flow-xhigh.md` and "four definitions, one per
    effort, `xhigh` dispatched for implementer groups alone". `README.md`'s class table `big` row:
    `` `low`/`medium`/`high`/`xhigh` per group ``.
  - [x] **Step 5: Verify.** `cd stats && gofmt -l . && go vet ./cmd/flow/ ./internal/guard/
    ./internal/records/ && go test ./cmd/flow/ -run
    '^(TestRecordDispatchBeginAcceptsEffort|TestRunRecordDispatchBeginRejectsUnknownEffort)$'
    -count=1 && go test ./internal/guard/ -run '^TestCheckInstalledCitations$' -count=1`;
    `scripts/test-setup-agents.sh`; `scripts/check-installed-citations.sh`;
    `scripts/check-references.sh` — all exit 0.

- [x] 7. Live verification: timings, parity, containment

**Files:** none
**Tests:** none — measurement task; the figures it records are the check
**Regression:** none — no commit
**Baseline:** before=0 after=0
<!-- predicted: no test is added by this task -->
**After:** Task 1, 2, 3, 4, 5, 6
**Build:** green

**Decision:** setuptest-timing-interleaved-median

**Decision:** setuptest-leak-detection

  - [x] **Step 1: Baseline checkout.** `git -C /Users/tweety53/Projects/agents worktree add --detach
    /Users/tweety53/Projects/agents/.worktrees/kan-844-baseline 75410c1c`.
  - [x] **Step 2: Interleaved runs.** One warm-up each, then five pairs, recording each run's
    `/usr/bin/time -p` real/user/sys and `uptime`'s load averages:
    `A = /Users/tweety53/Projects/agents/.worktrees/kan-844-baseline/scripts/test-setup.sh`,
    `B = /Users/tweety53/Projects/agents/.worktrees/kan-844-agents-speed-up-scripts-test-setup-sh/scripts/test-setup.sh`, run A, B, A, B, … Every run must exit 0.
  - [x] **Step 3: Parity, once more on the head.** The preamble's parity check with every title
    (task 5 step 3's `TITLES`, `RUN='.'`) — empty diff; `✓ leak detection: ` count 5.
  - [x] **Step 4: Record** in `design.md` `## Measurements`: a table of the ten timed runs
    (harness, real, user+sys, load average), both medians and median(A)/median(B), each figure
    tagged `measured:` with the command and `@ branch
    spectre/kan-844-agents-speed-up-scripts-test-setup-sh` (A: `@ 75410c1c`).
  - [x] **Step 5: Full suite.** Every command in `.flow/project.md`'s `## lint`;
    `scripts/run-guard-tests.sh`; `cd stats && go test ./... -race -count=1`; `cd stats/web &&
    npm test` — all exit 0.
  - [x] **Step 6: Judge, then clean up.** Failure looks like: median(A)/median(B) < 2; any parity
    line missing or extra; a leak-detection count other than 5; any run or suite red. Any of these
    is reported, not recorded as success. Then `git -C /Users/tweety53/Projects/agents worktree remove
    /Users/tweety53/Projects/agents/.worktrees/kan-844-baseline`.
