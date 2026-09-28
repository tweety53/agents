package setuptest

import (
	"bytes"
	"context"
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
	if err := setupRefusal(home, proj); err != nil {
		g.t.Fatal(err)
	}
	mkdirAll(g.t, proj)
	g.log = filepath.Join(g.dir, fmt.Sprintf("log-%d-%s.txt", g.seq, mode))
	f, err := os.Create(g.log)
	if err != nil {
		g.t.Fatal(err)
	}
	defer f.Close()
	// The installer runs in its own process group, killed whole when the test's deadline nears
	// or an interrupt arrives, so no orphaned child keeps writing into a sandbox being removed.
	ctx, cancel := context.WithCancel(runCtx)
	if dl, ok := g.t.Deadline(); ok {
		ctx, cancel = context.WithDeadline(runCtx, dl.Add(-5*time.Second))
	}
	defer cancel()
	cmd := exec.CommandContext(ctx, filepath.Join(repo, "setup.sh"), mode, proj)
	cmd.Dir = proj
	cmd.Env = withHome(os.Environ(), home)
	cmd.Stdout, cmd.Stderr = f, f
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = 5 * time.Second
	running.Add(1)
	defer running.Done()
	g.rc = 0
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			g.t.Fatalf("cannot run setup.sh: %v", err)
		}
		g.rc = ee.ExitCode()
	}
}

// setupRefusal names why runSetup must not hand the installer home and proj, or nil.
func setupRefusal(home, proj string) error {
	if !inSandbox(home) {
		return fmt.Errorf("refusing to run setup.sh with HOME=%s — not inside %s", home, sandbox)
	}
	if !inSandbox(proj) {
		return fmt.Errorf("refusing to run setup.sh with project dir %s — not inside %s", proj, sandbox)
	}
	return nil
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
	if err := seedRefusal(p); err != nil {
		g.t.Fatal(err)
	}
}

// seedRefusal names why seedGuard must refuse p, or nil.
func seedRefusal(p string) error {
	if !inSandbox(p) {
		return fmt.Errorf("refusing to seed %s — outside the sandbox %s", p, sandbox)
	}
	if isSymlink(p) {
		return fmt.Errorf("refusing to seed %s — it is a symlink, and writing through it would modify the link's target instead of the sandbox.", p)
	}
	for parent := filepath.Dir(p); inSandbox(parent); parent = filepath.Dir(parent) {
		if isSymlink(parent) {
			return fmt.Errorf("refusing to seed %s — its parent %s is a symlink, and writing under it would modify the link's target instead of the sandbox.", p, parent)
		}
	}
	return nil
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

// appendTo appends s to the existing file p, as a shell `>>` does.
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
