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
	cmd := exec.Command("bash", append([]string{"-c",
		`set -uo pipefail; . "$1" || exit 99; shift; "$@"`, "_", "../../../scripts/lib/" + lib, fn}, args...)...)
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
