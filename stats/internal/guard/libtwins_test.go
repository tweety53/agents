package guard

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// The parity tests below run each bash library and its Go twin over the same
// inputs and fail on any difference in stdout, stderr or status
// (shared-helper-go-twins, per helper-parity-tests). The bash side runs under
// `set -uo pipefail`, the state every sourcing script's `$(...)` calls it in:
// the callers set -euo pipefail, and a command substitution does not inherit
// errexit.

// libRes is one call's observable result, bash or Go.
type libRes struct {
	stdout, stderr string
	status         int
}

// bashLib sources scripts/lib/<lib> and runs fn with args in the test
// process's own working directory, the one the Go twin resolves against.
func bashLib(t *testing.T, lib, fn string, args ...string) libRes {
	t.Helper()
	return bashLibEnv(t, nil, lib, fn, args...)
}

// bashLibEnv is bashLib with env appended to the process environment.
func bashLibEnv(t *testing.T, env []string, lib, fn string, args ...string) libRes {
	t.Helper()
	return bashLibIO(t, env, nil, lib, fn, args...)
}

// bashLibIO is bashLibEnv with stdin fed to fn (nil: no stdin).
func bashLibIO(t *testing.T, env []string, stdin io.Reader, lib, fn string, args ...string) libRes {
	t.Helper()
	cmd := exec.Command("bash", append([]string{"-c",
		`set -uo pipefail; . "$1" || exit 99; shift; "$@"`, "_", "../../../scripts/lib/" + lib, fn}, args...)...)
	if env != nil {
		cmd.Env = append(os.Environ(), env...)
	}
	cmd.Stdin = stdin
	var out, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errb
	status := 0
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			t.Fatalf("bash %s: %v", fn, err)
		}
		status = ee.ExitCode()
	}
	return libRes{out.String(), errb.String(), status}
}

func assertParity(t *testing.T, goRes, bashRes libRes) {
	t.Helper()
	if goRes != bashRes {
		t.Errorf("Go twin differs from the bash library\n  go: %#v\nbash: %#v", goRes, bashRes)
	}
}

func symlink(t *testing.T, target, link string) {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
}

// TestResolveFileParity pins resolveFile against lib/resolve-file.sh's
// resolve_file.
func TestResolveFileParity(t *testing.T) {
	t.Parallel()
	d := t.TempDir()
	writeFile(t, d+"/file.txt", "x\n")
	writeFile(t, d+"/dir/inner.txt", "x\n")
	writeFile(t, d+"/sp ace/f", "x\n")
	symlink(t, d+"/file.txt", d+"/absLink")
	symlink(t, "file.txt", d+"/relLink")
	symlink(t, "../file.txt", d+"/dir/up")
	symlink(t, "relLink", d+"/c2")
	symlink(t, "c2", d+"/c1")
	symlink(t, "nowhere", d+"/dang")
	symlink(t, "dir", d+"/dirLink")
	symlink(t, "loopB", d+"/loopA")
	symlink(t, "loopA", d+"/loopB")
	// chainN: N links, the last pointing at file.txt — resolve_file follows
	// at most 40 hops.
	for _, n := range []int{40, 41} {
		pre := d + "/chain" + strconv.Itoa(n) + "_"
		for i := 0; i < n; i++ {
			next := filepath.Base(pre) + strconv.Itoa(i+1)
			if i == n-1 {
				next = "file.txt"
			}
			symlink(t, next, pre+strconv.Itoa(i))
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(cwd, d+"/relLink")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, in string }{
		{"plain file", d + "/file.txt"},
		{"absolute symlink", d + "/absLink"},
		{"relative symlink", d + "/relLink"},
		{"relative symlink up a directory", d + "/dir/up"},
		{"symlink chain", d + "/c1"},
		{"dangling link", d + "/dang"},
		{"directory", d + "/dir"},
		{"directory with a trailing slash", d + "/dir/"},
		{"directory with trailing slashes", d + "/dir//"},
		{"symlinked directory with a trailing slash", d + "/dirLink/"},
		{"leaf under a symlinked directory", d + "/dirLink/inner.txt"},
		{"trailing dot", d + "/dir/."},
		{"trailing dot-dot", d + "/dir/.."},
		{"bare dot", "."},
		{"bare dot-dot", ".."},
		{"empty path", ""},
		{"root", "/"},
		{"root-parented system path", "/tmp"},
		{"relative path", rel},
		{"path with a space", d + "/sp ace/f"},
		{"symlink cycle", d + "/loopA"},
		{"40-link chain", d + "/chain40_0"},
		{"41-link chain", d + "/chain41_0"},
		{"missing parent directory", d + "/nope/x"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var got libRes
			if p, ok := resolveFile(tc.in); ok {
				got.stdout = p + "\n"
			} else {
				got.status = 1
			}
			assertParity(t, got, bashLib(t, "resolve-file.sh", "resolve_file", tc.in))
		})
	}
}

// TestProjectSectionParity pins projectSection against
// lib/project-section.sh's project_section.
func TestProjectSectionParity(t *testing.T) {
	t.Parallel()
	d := t.TempDir()
	for i, tc := range []struct{ name, body, key string }{
		{"present section", "# P\n\n## lint\n\nrun lint\n\n## test\n\nrun test\n", "lint"},
		{"absent section", "# P\n\n## lint\n\nrun lint\n", "run"},
		{"empty body", "## lint\n\n \t\n## test\nx\n", "lint"},
		{"fenced body", "## lint\n\n```bash\nmake lint\n```\n\n## test\n", "lint"},
		{"heading inside a fence still ends the section", "## lint\n```\n## not a heading\n```\n", "lint"},
		{"subheadings stay in the body", "## lint\n### one\na\n#### two\nb\n## test\n", "lint"},
		{"deeper heading of the key does not match", "### lint\nx\n## lint\ny\n", "lint"},
		{"heading without a space does not match", "##lint\nx\n", "lint"},
		{"heading with a trailing space does not match", "## lint \nx\n", "lint"},
		{"longer heading does not match", "## linter\nx\n## lint\ny\n", "lint"},
		{"heading with a CR does not match", "## lint\r\nx\r\n", "lint"},
		{"trailing prose at EOF without a newline", "## run\n\nsome prose\n\n  more prose", "run"},
		{"interior blank lines kept, edge whitespace lines trimmed", "## run\n\t\n\r\na\n\nb\n \n\v\n", "run"},
		{"a repeated key heading keeps grabbing", "## lint\na\n## lint\nb\n## test\n", "lint"},
		{"leading BOM", "\xef\xbb\xbf## lint\nx\n", "lint"},
		{"empty file", "", "lint"},
	} {
		file := d + "/p" + strconv.Itoa(i) + ".md"
		writeFile(t, file, tc.body)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var got libRes
			if s := projectSection(file, tc.key); s != "" {
				got.stdout = s + "\n"
			}
			assertParity(t, got, bashLib(t, "project-section.sh", "project_section", file, tc.key))
		})
	}
}

// TestPostMutationCheckParity pins snapshotTreeState and checkTreeRestored
// against lib/post-mutation-check.sh: each row snapshots its tree with both
// sides, mutates it, then checks it against the bash snapshot with both.
func TestPostMutationCheckParity(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name          string
		before, after func(t *testing.T, wt string)
	}{
		{"clean tree", nil, nil},
		{"dirty tracked file present at the snapshot", func(t *testing.T, wt string) {
			writeFile(t, wt+"/a.txt", "dirty\n")
		}, nil},
		{"tracked file dirtied after the snapshot", nil, func(t *testing.T, wt string) {
			writeFile(t, wt+"/a.txt", "dirty\n")
		}},
		{"new untracked file", nil, func(t *testing.T, wt string) {
			writeFile(t, wt+"/new.txt", "x\n")
		}},
		{"new stash entry, an older one renumbered", func(t *testing.T, wt string) {
			writeFile(t, wt+"/a.txt", "one\n")
			gitRun(t, wt, "stash", "push", "-m", "old")
		}, func(t *testing.T, wt string) {
			writeFile(t, wt+"/a.txt", "two\n")
			gitRun(t, wt, "stash", "push", "-m", "new")
		}},
		{"stash present at the snapshot and still present", func(t *testing.T, wt string) {
			writeFile(t, wt+"/a.txt", "one\n")
			gitRun(t, wt, "stash", "push", "-m", "old")
		}, nil},
		{"drift restored before the check", func(t *testing.T, wt string) {
			writeFile(t, wt+"/a.txt", "dirty\n")
		}, func(t *testing.T, wt string) {
			gitRun(t, wt, "checkout", "--", "a.txt")
		}},
		{"stash and status drift together", nil, func(t *testing.T, wt string) {
			writeFile(t, wt+"/a.txt", "one\n")
			gitRun(t, wt, "stash", "push", "-m", "s")
			writeFile(t, wt+"/b.txt", "x\n")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			wt := t.TempDir() + "/work tree"
			gitRun(t, "", "init", "-q", "-b", "main", wt)
			writeFile(t, wt+"/a.txt", "a\n")
			gitRun(t, wt, "add", "a.txt")
			gitRun(t, wt, "commit", "-q", "-m", "init")
			if tc.before != nil {
				tc.before(t, wt)
			}
			snap := bashLib(t, "post-mutation-check.sh", "snapshot_tree_state", wt)
			assertParity(t, goSnapshot(wt), snap)
			if tc.after != nil {
				tc.after(t, wt)
			}
			// The callers capture the snapshot through $(...), which strips
			// its trailing newlines.
			snapshot := strings.TrimRight(snap.stdout, "\n")
			assertParity(t, goCheck(wt, snapshot),
				bashLib(t, "post-mutation-check.sh", "check_tree_restored", wt, snapshot))
		})
	}
	t.Run("missing worktree directory", func(t *testing.T) {
		t.Parallel()
		wt := t.TempDir() + "/missing"
		snap := bashLib(t, "post-mutation-check.sh", "snapshot_tree_state", wt)
		assertParity(t, goSnapshot(wt), snap)
		assertParity(t, goCheck(wt, snap.stdout),
			bashLib(t, "post-mutation-check.sh", "check_tree_restored", wt, snap.stdout))
	})
	t.Run("not a git repository", func(t *testing.T) {
		t.Parallel()
		wt := t.TempDir()
		snap := bashLib(t, "post-mutation-check.sh", "snapshot_tree_state", wt)
		assertParity(t, goSnapshot(wt), snap)
		assertParity(t, goCheck(wt, snap.stdout),
			bashLib(t, "post-mutation-check.sh", "check_tree_restored", wt, snap.stdout))
	})
}

func goSnapshot(wt string) libRes {
	var errb strings.Builder
	out, status := snapshotTreeState(wt, &errb)
	return libRes{out, errb.String(), status}
}

func goCheck(wt, snapshot string) libRes {
	var errb strings.Builder
	findings, status := checkTreeRestored(wt, snapshot, &errb)
	var out string
	for _, f := range findings {
		out += f + "\n"
	}
	return libRes{out, errb.String(), status}
}

// TestGitExecSignalStatus pins gitExec's status for a git killed by a signal
// at bash's 128+n: mutate-and-verify and prepare-archive-branch return it as
// their own exit (set -e), where a raw ExitCode() of -1 left mutate-and-verify
// exiting 255, outside its 0/2/3/4 contract. The alias's shell SIGTERMs its
// parent, the git gitExec started.
func TestGitExecSignalStatus(t *testing.T) {
	t.Parallel()
	_, rc := gitExec(t.TempDir(), nil, io.Discard, "-c", "alias.die=!kill -TERM $PPID", "die")
	if rc != 143 {
		t.Fatalf("gitExec status for a SIGTERM'd git: got %d, want 143", rc)
	}
}

// envWith is an Env over the test process's own environment and working
// directory with over's NAME=value pairs laid on top, the same pairs
// bashLibEnv hands the bash side.
func envWith(t *testing.T, over []string) Env {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	m := map[string]string{}
	for _, kv := range over {
		k, v, _ := strings.Cut(kv, "=")
		m[k] = v
	}
	look := func(k string) (string, bool) {
		if v, ok := m[k]; ok {
			return v, true
		}
		return os.LookupEnv(k)
	}
	return Env{Dir: cwd, LookupEnv: look, Getenv: func(k string) string { v, _ := look(k); return v }}
}

// TestPanelTouchedPathsParity pins panelResolveGit, panelValidateWorktree and
// panelTouchedPaths against lib/panel-touched-paths.sh's three functions.
func TestPanelTouchedPathsParity(t *testing.T) {
	t.Parallel()
	d := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relTo := func(p string) string {
		r, err := filepath.Rel(cwd, p)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}

	// The change: a base commit, then a committed rename and additions,
	// staged and unstaged edits, and an untracked file git diff never lists.
	var g fxGit
	repo := d + "/work tree"
	g.git("", "init", "-q", "-b", "main", repo)
	for _, f := range []string{"a.txt", "old.txt", "tracked.md", "sp ace.txt", "B.md"} {
		g.write(repo+"/"+f, "base "+f)
	}
	g.git(repo, "add", ".")
	g.git(repo, "commit", "-qm", "base")
	mb := g.git(repo, "rev-parse", "HEAD")
	g.git(repo, "mv", "old.txt", "new.txt")
	for _, f := range []string{"with space.md", "a.md", "_u.txt", "Z.mdc"} {
		g.write(repo+"/"+f, f)
	}
	g.appendLine(repo+"/B.md", "committed")
	g.git(repo, "add", ".")
	g.git(repo, "commit", "-qm", "change")
	g.write(repo+"/staged.go", "package x")
	g.appendLine(repo+"/a.txt", "staged")
	g.git(repo, "add", "staged.go", "a.txt")
	g.appendLine(repo+"/tracked.md", "unstaged")
	g.appendLine(repo+"/sp ace.txt", "unstaged")
	g.appendLine(repo+"/B.md", "unstaged")
	g.write(repo+"/untracked.txt", "untracked")
	clean := d + "/clean"
	g.git("", "init", "-q", "-b", "main", clean)
	g.write(clean+"/a.txt", "a")
	g.git(clean, "add", "a.txt")
	g.git(clean, "commit", "-qm", "base")
	cleanMB := g.git(clean, "rev-parse", "HEAD")
	if g.err != nil {
		t.Fatal(g.err)
	}
	writeFile(t, d+"/file", "x\n")
	mkdir(t, d+"/plain")
	mkdir(t, d+"/nogit")
	stubGit(t, d+"/override", "false", "never")
	stubGit(t, d+"/committed", `has `+mb+`..HEAD "$@"`, "simulated committed failure")
	stubGit(t, d+"/staged", `has --cached "$@"`, "simulated staged failure")
	stubGit(t, d+"/unstaged", `[ "${!#}" = --name-only ]`, "simulated unstaged failure")
	path := os.Getenv("PATH")

	for _, tc := range []struct{ name, path string }{
		{"resolve: git on PATH", path},
		{"resolve: an override git first on PATH", d + "/override:" + path},
		{"resolve: no git on PATH", d + "/nogit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			over := []string{"PATH=" + tc.path}
			var got libRes
			var errb strings.Builder
			if p, ok := panelResolveGit(envWith(t, over), "prog", &errb); ok {
				got.stdout = p + "\n"
			} else {
				got.status = 2
			}
			got.stderr = errb.String()
			assertParity(t, got, bashLibEnv(t, over, "panel-touched-paths.sh", "panel_resolve_git", "prog"))
		})
	}

	for _, tc := range []struct{ name, wt, mb string }{
		{"validate: missing worktree argument", "", mb},
		{"validate: missing merge-base argument", repo, ""},
		{"validate: worktree absent", d + "/absent", mb},
		{"validate: worktree a regular file", d + "/file", mb},
		{"validate: worktree not a git repository", d + "/plain", mb},
		{"validate: unknown merge base", repo, "0000000000000000000000000000000000000000"},
		{"validate: flag-shaped merge base", repo, "--evil"},
		{"validate: valid worktree and merge base", repo, mb},
		{"validate: relative worktree path", relTo(repo), mb},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var got libRes
			var errb strings.Builder
			if !panelValidateWorktree(envWith(t, nil), "prog", tc.wt, tc.mb, fixtureGit, &errb) {
				got.status = 2
			}
			got.stderr = errb.String()
			assertParity(t, got, bashLib(t, "panel-touched-paths.sh", "panel_validate_worktree", "prog", tc.wt, tc.mb, fixtureGit))
		})
	}

	for _, tc := range []struct {
		name, wt, mb, git string
		over              []string
	}{
		{"touched: committed, staged and unstaged, a rename and spaces, untracked left out", repo, mb, fixtureGit, nil},
		{"touched: C collation", repo, mb, fixtureGit, []string{"LC_ALL=C"}},
		{"touched: en_US.UTF-8 collation", repo, mb, fixtureGit, []string{"LC_ALL=en_US.UTF-8"}},
		{"touched: relative worktree path", relTo(repo), mb, fixtureGit, nil},
		{"touched: empty union", clean, cleanMB, fixtureGit, nil},
		{"touched: committed-paths failure", repo, mb, d + "/committed/git", nil},
		{"touched: staged-paths failure", repo, mb, d + "/staged/git", nil},
		{"touched: unstaged-paths failure", repo, mb, d + "/unstaged/git", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var got libRes
			var errb strings.Builder
			paths, ok := panelTouchedPaths(envWith(t, tc.over), "prog", tc.wt, tc.mb, tc.git, &errb)
			for _, p := range paths {
				got.stdout += p + "\n"
			}
			if !ok {
				got.status = 2
			}
			got.stderr = errb.String()
			assertParity(t, got, bashLibEnv(t, tc.over, "panel-touched-paths.sh", "panel_touched_paths", "prog", tc.wt, tc.mb, tc.git))
		})
	}
}

// TestStripBOMParity pins stripBOM against lib/strip-bom.sh's strip_bom_cat.
func TestStripBOMParity(t *testing.T) {
	t.Parallel()
	d := t.TempDir()
	for i, tc := range []struct{ name, body string }{
		{"leading BOM", "\xef\xbb\xbf## visual verification\n"},
		{"BOM not at byte 0", "x\xef\xbb\xbfy\n"},
		{"two BOM bytes only", "\xef\xbb"},
		{"two BOMs", "\xef\xbb\xbf\xef\xbb\xbfz"},
		{"empty file", ""},
		{"missing file", ""},
	} {
		file := d + "/f" + strconv.Itoa(i)
		if tc.name != "missing file" {
			writeFile(t, file, tc.body)
		}
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			bashRes := bashLib(t, "strip-bom.sh", "strip_bom_cat", file)
			b, err := os.ReadFile(file)
			if err != nil {
				// The twin takes bytes; a missing file is its caller's read
				// error. The bash prints nothing on stdout and fails.
				if bashRes.stdout != "" || bashRes.status == 0 {
					t.Errorf("missing file: bash printed %q, status %d", bashRes.stdout, bashRes.status)
				}
				return
			}
			assertParity(t, libRes{stdout: string(stripBOM(b))}, bashRes)
		})
	}
}

// TestSanitizeDisplayParity pins sanitizeDisplay against
// lib/sanitize-display.sh's sanitize_display, one line per input as every
// caller's `printf '...\n' | sanitize_display` feeds it.
func TestSanitizeDisplayParity(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"DEL":                   "a\x7fb",
		"backslash":             `a\b\\c`,
		"multi-byte UTF-8 run":  "héllo — ✓ 日本",
		"CR at end":             "line\r",
		"NUL mid-line":          "a\x00b\\c\x1b",
		"NUL first":             "\x00abc",
		"invalid UTF-8":         "\xff\xfe\xc3",
		"escape sequence":       "\x1b[31mred\x1b[0m",
		"empty line":            "",
		"interior newline kept": "a\x01\nb\x02",
	}
	for c := 0; c < 0x20; c++ {
		cases["byte "+strconv.Itoa(c)] = "x" + string(rune(c)) + "y"
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			bashRes := bashLibIO(t, nil, strings.NewReader(in+"\n"), "sanitize-display.sh", "sanitize_display")
			assertParity(t, libRes{stdout: sanitizeDisplay(in + "\n")}, bashRes)
		})
	}
}

// vtDriver prints, per input line, split_cells' count and every cell raw,
// trimmed and folded, then trimcell and foldcell of the whole line — fields
// separated by \x1e (\036 to awk, whose \x escape would swallow a
// following hex letter), which no input carries.
const vtDriver = `{
  n = split_cells($0, cells)
  printf "%d", n
  for (i = 1; i <= n; i++) printf "\036%s\036%s\036%s", cells[i], trimcell(cells[i]), foldcell(cells[i])
  printf "\036T%s\036F%s\n", trimcell($0), foldcell($0)
}
`

func vtGoDriver(line string) string {
	cells := vtSplitCells(line)
	var b strings.Builder
	b.WriteString(strconv.Itoa(len(cells)))
	for _, c := range cells {
		b.WriteString("\x1e" + c + "\x1e" + vtTrimCell(c) + "\x1e" + vtFoldCell(c))
	}
	b.WriteString("\x1eT" + vtTrimCell(line) + "\x1eF" + vtFoldCell(line) + "\n")
	return b.String()
}

// TestVisualTableCellsParity pins vtSplitCells, vtTrimCell and vtFoldCell
// against lib/visual-table-cells.awk, run the way its callers run it — as a
// second -f beside the caller's program — under LC_ALL=C: the C locale's
// [[:space:]] and an ASCII-only tolower (the non-ASCII rows pin that).
func TestVisualTableCellsParity(t *testing.T) {
	t.Parallel()
	d := t.TempDir()
	driver := d + "/driver.awk"
	writeFile(t, driver, vtDriver)
	for i, line := range []string{
		"| Setting | Value |",
		"| Setting | Value",
		`| a \| b | c \\ d |`,
		`| a \x b | c\ |`,
		`| trailing backslash \`,
		"| UI Paths | `a/**`, `b/**` |\r",
		"  \t| ` spaced ` |  `x`  |  ",
		"| Ui \t  Paths |   many   \t spaces |",
		"not a row | x |",
		"| a | |",
		"| a ||",
		"|",
		"||",
		"",
		"| MiXeD CaSe | VaLuE |",
		"| \v\fvt\v | \x85x\xc2\xa0 |",
		"| ÄRGER | \xc2\xa0nbsp\xc2\xa0 |",
		"| `` | ` |",
	} {
		t.Run(strconv.Itoa(i)+" "+strconv.Quote(line), func(t *testing.T) {
			t.Parallel()
			cmd := exec.Command("awk", "-f", "../../../scripts/lib/visual-table-cells.awk", "-f", driver)
			cmd.Env = append(os.Environ(), "LC_ALL=C")
			cmd.Stdin = strings.NewReader(line + "\n")
			out, err := cmd.Output()
			if err != nil {
				t.Fatalf("awk: %v", err)
			}
			assertParity(t, libRes{stdout: vtGoDriver(line)}, libRes{stdout: string(out)})
		})
	}
}
