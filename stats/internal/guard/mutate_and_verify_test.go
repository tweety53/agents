package guard

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Every case of scripts/test-mutate-and-verify.sh at d71a2327, one subtest per
// ok: label, nested under the harness's own case. The bash harness built a
// throwaway repository per case carrying a fixture guard.sh and a fixture
// test-*.sh harness for it, and ran the script from inside it; here each case
// copies one of three repositories built once, and the guard runs in-process
// from the same directory. Each case also carries an "output pinned" subtest:
// the whole stdout, stderr and exit the bash script printed for that fixture
// at d71a2327, its paths replaced by <repo> and <patch>. The cases after the
// harness's twelve are the plan's Review Focus rows: tree and HEAD unchanged
// after every refusal and cannot-answer exit, a repository path with a space,
// the real shim with no go on PATH, and a signal mid-run.

// mvGuard is build_guard: `guard.sh <n>` exits 0 for even n, 1 for odd n in
// 1..cases, and 1 for anything else; mutation alters exactly one thing (see
// the harness's own comment: "", "flip:<i>", "catchall", "always-fail").
func mvGuard(cases int, mutation string) string {
	var b strings.Builder
	b.WriteString("#!/usr/bin/env bash\nn=\"$1\"\n")
	if mutation == "always-fail" {
		b.WriteString("exit 1\n")
	}
	b.WriteString("case \"$n\" in\n")
	for i := 1; i <= cases; i++ {
		want := i % 2
		if mutation == fmt.Sprintf("flip:%d", i) {
			want = 1 - want
		}
		fmt.Fprintf(&b, "  %d) exit %d ;;\n", i, want)
	}
	if mutation == "catchall" {
		b.WriteString("  *) exit 0 ;;\n")
	} else {
		b.WriteString("  *) exit 1 ;;\n")
	}
	b.WriteString("esac\n")
	return b.String()
}

// mvHarness is build_harness: guard.sh called once per n, `ok: case-<n>` or
// `FAIL: case-<n>` per result.
func mvHarness(cases int) string {
	var b strings.Builder
	b.WriteString("#!/usr/bin/env bash\nset -euo pipefail\n" +
		"DIR=\"$(cd \"$(dirname \"${BASH_SOURCE[0]}\")\" && pwd)\"\nGUARD=\"$DIR/guard.sh\"\n")
	for i := 1; i <= cases; i++ {
		fmt.Fprintf(&b, "if \"$GUARD\" %d >/dev/null 2>&1; then got=0; else got=1; fi\n", i)
		fmt.Fprintf(&b, "if [ \"$got\" -eq %d ]; then echo \"ok: case-%d\"; else echo \"FAIL: case-%d\"; fi\n",
			i%2, i, i)
	}
	b.WriteString("exit 0\n")
	return b.String()
}

// mvMutationMark is MUTATION_MARK: the text the flip:2 patch writes.
const mvMutationMark = "2) exit 1"

// mvStashHarness is stash_harness: `ok: harness case 1`, and on the mutated
// pass only, the applied mutation stashed away and left behind; with stage it
// also overwrites and stages the touched file.
func mvStashHarness(touched string, stage bool) string {
	b := "#!/usr/bin/env bash\nset -e\necho \"ok: harness case 1\"\n" +
		fmt.Sprintf("if grep -q \"%s\" \"%s\" 2>/dev/null; then\n", mvMutationMark, touched) +
		"  git stash push -q --include-untracked -m kan-448-residue\n"
	if stage {
		b += fmt.Sprintf("  printf %%s \"%s\" > \"%s\"\n  git add \"%s\"\n", mvMutationMark, touched, touched)
	}
	return b + "fi\n"
}

// mvFixtures is the three repositories new_fixture_repo built (4, 10 and 12
// cases) and every patch the harness made against them, built once.
type mvFixtures struct {
	base    map[int]string    // cases -> repository
	patches map[string]string // "<cases>/<mutation>" -> patch text
}

func mvBuild(t *testing.T) mvFixtures {
	t.Helper()
	fx := mvFixtures{map[int]string{}, map[string]string{}}
	dir := t.TempDir()
	for _, n := range []int{4, 10, 12} {
		repo := fmt.Sprintf("%s/base-%d", dir, n)
		gitRun(t, "", "init", "-q", "-b", "main", "--template=", repo)
		gitRun(t, repo, "config", "user.email", "test@example.invalid")
		gitRun(t, repo, "config", "user.name", "Test")
		writeExec(t, repo+"/guard.sh", mvGuard(n, ""))
		writeExec(t, repo+"/test-fixture.sh", mvHarness(n))
		gitRun(t, repo, "add", "guard.sh", "test-fixture.sh")
		gitRun(t, repo, "commit", "-qm", "fixture")
		fx.base[n] = repo
	}
	// make_patch: the diff between the committed guard and the same guard
	// rebuilt with the mutation.
	for _, p := range []struct {
		n        int
		mutation string
	}{{4, "flip:2"}, {4, "catchall"}, {10, "always-fail"}, {12, "always-fail"}} {
		repo := fx.base[p.n]
		writeExec(t, repo+"/guard.sh", mvGuard(p.n, p.mutation))
		fx.patches[fmt.Sprintf("%d/%s", p.n, p.mutation)] = mvGit(t, repo, "diff", "--", "guard.sh")
		writeExec(t, repo+"/guard.sh", mvGuard(p.n, ""))
	}
	// Case 10's patch creates new-file.txt as a brand-new file.
	repo := fx.base[4]
	writeFile(t, repo+"/new-file.txt", "net new content\n")
	gitRun(t, repo, "add", "new-file.txt")
	fx.patches["4/newfile"] = mvGit(t, repo, "diff", "--cached", "--", "new-file.txt")
	gitRun(t, repo, "reset", "-q", "--", "new-file.txt")
	if err := os.Remove(repo + "/new-file.txt"); err != nil {
		t.Fatal(err)
	}
	for n, repo := range fx.base {
		if s := mvGit(t, repo, "status", "--porcelain"); s != "" {
			t.Fatalf("base-%d left dirty: %s", n, s)
		}
	}
	return fx
}

// mvGit runs git -C dir under fixtureGitEnv and returns its whole stdout.
func mvGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(fixtureGit, append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), fixtureGitEnv...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, stderr.String())
	}
	return string(out)
}

// mvCopyTree copies src to dst; an executable is linked to its master inode
// (writeExec), so a copied guard or harness costs no fresh assessment.
func mvCopyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := dst + strings.TrimPrefix(p, src)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		fi, err := d.Info()
		if err != nil {
			return err
		}
		if fi.Mode()&0o111 != 0 {
			writeExec(t, target, string(b))
			return nil
		}
		return os.WriteFile(target, b, fi.Mode().Perm())
	})
	if err != nil {
		t.Fatal(err)
	}
}

// mvCase is one case's own repository and patch file.
type mvCase struct {
	t                     *testing.T
	repo, guard, harness  string
	patch                 string
	statusBefore, headBef string
}

// newFixture is new_fixture_repo plus make_patch: a copy of the n-case
// repository under dir, and the named patch written outside it.
func (fx mvFixtures) newFixture(t *testing.T, dir string, n int, patch string) *mvCase {
	t.Helper()
	repo := dir + "/repo"
	mvCopyTree(t, fx.base[n], repo)
	c := &mvCase{t: t, repo: repo, guard: repo + "/guard.sh", harness: repo + "/test-fixture.sh"}
	if patch != "" {
		c.writePatch(fx.patches[patch])
	}
	return c
}

func (c *mvCase) writePatch(body string) {
	c.patch = filepath.Dir(c.repo) + "/patch"
	writeFile(c.t, c.patch, body)
}

// snap records `git status --porcelain` and HEAD, for tree-unchanged checks.
func (c *mvCase) snap() {
	c.statusBefore = mvGit(c.t, c.repo, "status", "--porcelain")
	c.headBef = mvGit(c.t, c.repo, "rev-parse", "HEAD")
}

// unchanged is the Review Focus row: `git status --porcelain` and HEAD
// exactly as snap recorded them.
func (c *mvCase) unchanged(label string) {
	c.t.Helper()
	st, head := mvGit(c.t, c.repo, "status", "--porcelain"), mvGit(c.t, c.repo, "rev-parse", "HEAD")
	mvCheck(c.t, label, st == c.statusBefore && head == c.headBef,
		"status %q head %q, want status %q head %q", st, head, c.statusBefore, c.headBef)
}

func (c *mvCase) clean() bool { return mvGit(c.t, c.repo, "status", "--porcelain") == "" }

type mvRes struct {
	rc       int
	out, err string
}

// all is the harness's OUT: stdout and stderr together (2>&1).
func (r mvRes) all() string { return r.out + r.err }

// run is run_mutate: the guard from inside repo, with vars on its Env.
func (c *mvCase) run(vars map[string]string, args ...string) mvRes {
	return mvRunIn(c.repo, vars, args...)
}

func mvRunIn(dir string, vars map[string]string, args ...string) mvRes {
	var out, errb bytes.Buffer
	rc := mutateAndVerify(args, crEnv(dir, vars), &out, &errb)
	return mvRes{rc, out.String(), errb.String()}
}

func mvCheck(t *testing.T, label string, ok bool, format string, a ...any) {
	t.Helper()
	t.Run(label, func(t *testing.T) {
		if !ok {
			t.Fatalf(format, a...)
		}
	})
}

// pinned asserts r is, byte for byte, what the bash script printed for this
// fixture at d71a2327 (mvPins, keyed by the subtest's name).
func (c *mvCase) pinned(r mvRes, extra ...string) {
	c.t.Helper()
	repoPhys, err := filepath.EvalSymlinks(c.repo)
	if err != nil {
		c.t.Fatal(err)
	}
	pairs := []string{repoPhys, "<repo>", c.repo, "<repo>"}
	if c.patch != "" {
		pairs = append(pairs, c.patch, "<patch>")
	}
	pairs = append(pairs, extra...)
	norm := strings.NewReplacer(pairs...)
	got := fmt.Sprintf("%s--- stderr\n%s--- exit %d\n", norm.Replace(r.out), norm.Replace(r.err), r.rc)
	want := mvPins[c.t.Name()]
	mvCheck(c.t, "output pinned", got == want, "got:\n%q\nwant:\n%q", got, want)
}

func TestMutateAndVerify(t *testing.T) {
	t.Parallel()
	fx := mvBuild(t)
	cases := []struct {
		name string
		fn   func(t *testing.T)
	}{
		// 1. caught — a patch that flips exactly one fixture-harness case.
		{"1 caught", func(t *testing.T) {
			c := fx.newFixture(t, t.TempDir(), 4, "4/flip:2")
			r := c.run(nil, c.patch, c.harness)
			mvCheck(t, "caught: exit 0", r.rc == 0, "rc=%d out=%s", r.rc, r.all())
			mvCheck(t, "caught: verdict names caught", has(r.all(), "caught"), "%s", r.all())
			mvCheck(t, "caught: does not flag suspicious blast radius", !has(r.all(), "suspicious"), "%s", r.all())
			mvCheck(t, "caught: names case-2 as the new failure", has(r.all(), "case-2"), "%s", r.all())
			mvCheck(t, "caught: names exactly one new failure",
				!has(r.all(), "case-1", "new failure") && !has(r.all(), "case-3", "new failure") &&
					!has(r.all(), "case-4", "new failure"), "%s", r.all())
			mvCheck(t, "caught: touched file clean after restore", c.clean(), "dirty")
			c.pinned(r)
		}},
		// 2. surviving mutant — the catch-all arm, which no case observes.
		{"2 surviving mutant", func(t *testing.T) {
			c := fx.newFixture(t, t.TempDir(), 4, "4/catchall")
			r := c.run(nil, c.patch, c.harness)
			mvCheck(t, "surviving mutant: exit 0", r.rc == 0, "rc=%d out=%s", r.rc, r.all())
			mvCheck(t, "surviving mutant: verdict names surviving mutant", has(r.all(), "surviving mutant"), "%s", r.all())
			mvCheck(t, "surviving mutant: reports 0 new failures", has(r.all(), "0 new failures"), "%s", r.all())
			mvCheck(t, "surviving mutant: touched file clean after restore", c.clean(), "dirty")
			c.pinned(r)
		}},
		// 3. suspicious blast radius — 12 cases, every even one broken (6).
		{"3 blast radius", func(t *testing.T) {
			c := fx.newFixture(t, t.TempDir(), 12, "12/always-fail")
			r := c.run(nil, c.patch, c.harness)
			mvCheck(t, "blast radius: exit 0", r.rc == 0, "rc=%d out=%s", r.rc, r.all())
			mvCheck(t, "blast radius: verdict names suspicious blast radius", has(r.all(), "suspicious blast radius"), "%s", r.all())
			mvCheck(t, "blast radius: reports the actual new-failure count (6)", has(r.all(), "6 new failures"), "%s", r.all())
			mvCheck(t, "blast radius: names the default bound (5)", has(r.all(), "bound 5"), "%s", r.all())
			mvCheck(t, "blast radius: touched file clean after restore", c.clean(), "dirty")
			c.pinned(r)
		}},
		// 4. MUTATE_AND_VERIFY_MAX_NEW_FAILURES override — the bound raised
		//    above the new-failure count flips the verdict to caught.
		{"4 bound override", func(t *testing.T) {
			c := fx.newFixture(t, t.TempDir(), 12, "12/always-fail")
			r := c.run(map[string]string{"MUTATE_AND_VERIFY_MAX_NEW_FAILURES": "10"}, c.patch, c.harness)
			mvCheck(t, "bound override: exit 0", r.rc == 0, "rc=%d out=%s", r.rc, r.all())
			mvCheck(t, "bound override: no longer flags suspicious blast radius", !has(r.all(), "suspicious"), "%s", r.all())
			mvCheck(t, "bound override: verdict names caught", has(r.all(), "caught"), "%s", r.all())
			mvCheck(t, "bound override: names the overridden bound (10)",
				has(r.all(), "bound: 10") || has(r.all(), "bound 10"), "%s", r.all())
			mvCheck(t, "bound override: touched file clean after restore", c.clean(), "dirty")
			c.pinned(r)
		}},
		// 5. patch does not apply — refused before anything is mutated.
		{"5 patch does not apply", func(t *testing.T) {
			c := fx.newFixture(t, t.TempDir(), 4, "")
			c.writePatch("diff --git a/nonexistent.sh b/nonexistent.sh\nindex 0000000..1111111 100644\n" +
				"--- a/nonexistent.sh\n+++ b/nonexistent.sh\n@@ -1,1 +1,1 @@\n" +
				"-old line that is not present anywhere\n+new line\n")
			before := shasum(t, c.guard)
			c.snap()
			r := c.run(nil, c.patch, c.harness)
			mvCheck(t, "patch does not apply: exit 2", r.rc == 2, "rc=%d out=%s", r.rc, r.all())
			mvCheck(t, "patch does not apply: fixture file byte-for-byte unchanged", shasum(t, c.guard) == before, "changed")
			mvCheck(t, "patch does not apply: no harness ran, no report printed",
				!has(r.all(), "baseline pass") && !has(r.all(), "mutated pass"), "%s", r.all())
			mvCheck(t, "patch does not apply: repo stays clean", c.clean(), "dirty")
			c.unchanged("patch does not apply: status and HEAD unchanged")
			c.pinned(r)
		}},
		// 6. dirty touched file refused — the uncommitted edit survives.
		{"6 dirty touched file", func(t *testing.T) {
			c := fx.newFixture(t, t.TempDir(), 4, "4/flip:2")
			dirty := mvGuard(4, "") + "# pre-existing uncommitted edit\n"
			writeFile(t, c.guard, dirty)
			if err := os.Chmod(c.guard, 0o755); err != nil {
				t.Fatal(err)
			}
			c.snap()
			r := c.run(nil, c.patch, c.harness)
			mvCheck(t, "dirty touched file: exit 2", r.rc == 2, "rc=%d out=%s", r.rc, r.all())
			after, err := os.ReadFile(c.guard)
			mvCheck(t, "dirty touched file: uncommitted edit survives unchanged", err == nil && string(after) == dirty, "altered")
			mvCheck(t, "dirty touched file: names the reason", has(r.all(), "uncommitted changes"), "%s", r.all())
			c.unchanged("dirty touched file: status and HEAD unchanged")
			c.pinned(r)
		}},
		// 7. missing harness refused.
		{"7 missing harness", func(t *testing.T) {
			c := fx.newFixture(t, t.TempDir(), 4, "4/flip:2")
			c.snap()
			r := c.run(nil, c.patch, c.repo+"/no-such-harness.sh")
			mvCheck(t, "missing harness: exit 2", r.rc == 2, "rc=%d out=%s", r.rc, r.all())
			mvCheck(t, "missing harness: names the harness", has(r.all(), "no-such-harness.sh"), "%s", r.all())
			mvCheck(t, "missing harness: repo stays clean", c.clean(), "dirty")
			c.unchanged("missing harness: status and HEAD unchanged")
			c.pinned(r)
		}},
		// 8. non-executable harness refused.
		{"8 non-executable harness", func(t *testing.T) {
			c := fx.newFixture(t, t.TempDir(), 4, "4/flip:2")
			noexec := c.repo + "/test-fixture-noexec.sh"
			writeFile(t, noexec, mvHarness(4))
			c.snap()
			r := c.run(nil, c.patch, noexec)
			mvCheck(t, "non-executable harness: exit 2", r.rc == 2, "rc=%d out=%s", r.rc, r.all())
			mvCheck(t, "non-executable harness: names the harness", has(r.all(), "test-fixture-noexec.sh"), "%s", r.all())
			mvCheck(t, "non-executable harness: touched file stays clean",
				mvGit(t, c.repo, "status", "--porcelain", "--", "guard.sh") == "", "dirty")
			c.unchanged("non-executable harness: status and HEAD unchanged")
			c.pinned(r)
		}},
		// 9. boundary — new-failure count exactly the default bound (5).
		{"9 boundary", func(t *testing.T) {
			c := fx.newFixture(t, t.TempDir(), 10, "10/always-fail")
			r := c.run(nil, c.patch, c.harness)
			mvCheck(t, "boundary: exit 0", r.rc == 0, "rc=%d out=%s", r.rc, r.all())
			mvCheck(t, "boundary: verdict names caught at new_count == bound", has(r.all(), "caught"), "%s", r.all())
			mvCheck(t, "boundary: does not flag suspicious blast radius", !has(r.all(), "suspicious"), "%s", r.all())
			mvCheck(t, "boundary: reports exactly 5 new failures", has(r.all(), "5 new failure(s)"), "%s", r.all())
			mvCheck(t, "boundary: touched file clean after restore", c.clean(), "dirty")
			c.pinned(r)
		}},
		// 10. could not fully restore — the patch creates a brand-new file,
		//     which `git checkout --` cannot remove.
		{"10 cannot restore", func(t *testing.T) {
			c := fx.newFixture(t, t.TempDir(), 4, "4/newfile")
			r := c.run(nil, c.patch, c.harness)
			mvCheck(t, "cannot restore: exit 3", r.rc == 3, "rc=%d out=%s", r.rc, r.all())
			mvCheck(t, "cannot restore: names the reason", has(r.all(), "could not fully restore"), "%s", r.all())
			mvCheck(t, "cannot restore: does not claim ran clean", !has(r.all(), "ran clean"), "%s", r.all())
			mvCheck(t, "cannot restore: leftover untracked file still present", isFile(c.repo+"/new-file.txt"), "absent")
			c.pinned(r)
		}},
		// 11. post-restore stash residue — exits 2 naming the stash.
		{"11 stash residue", func(t *testing.T) {
			c := fx.newFixture(t, t.TempDir(), 4, "4/flip:2")
			h := c.repo + "/stash-harness.sh"
			writeExec(t, h, mvStashHarness(c.guard, false))
			r := c.run(nil, c.patch, h)
			mvCheck(t, "post-restore stash residue exits 2 naming the stash",
				r.rc == 2 && has(r.all(), "new stash entry: ", "kan-448-residue"), "rc=%d out=%s", r.rc, r.all())
			c.pinned(r)
		}},
		// 12. touched-file residual keeps exit 3, the stash still named.
		{"12 residual keeps exit 3", func(t *testing.T) {
			c := fx.newFixture(t, t.TempDir(), 4, "4/flip:2")
			h := c.repo + "/stash-harness.sh"
			writeExec(t, h, mvStashHarness(c.guard, true))
			r := c.run(nil, c.patch, h)
			mvCheck(t, "touched-file residual keeps exit 3 and the stash is still named",
				r.rc == 3 && has(r.all(), "could not fully restore") && has(r.all(), "new stash entry: ", "kan-448-residue"),
				"rc=%d out=%s", r.rc, r.all())
			c.pinned(r)
		}},
		// The cannot-answer exits (4) leave the tree and HEAD untouched.
		{"usage", func(t *testing.T) {
			c := fx.newFixture(t, t.TempDir(), 4, "4/flip:2")
			c.snap()
			r := c.run(nil, c.patch)
			mvCheck(t, "usage: one argument exits 4", r.rc == 4, "rc=%d out=%s", r.rc, r.all())
			c.unchanged("usage: status and HEAD unchanged")
			c.pinned(r)
		}},
		{"not a worktree", func(t *testing.T) {
			c := fx.newFixture(t, t.TempDir(), 4, "4/flip:2")
			plain := t.TempDir()
			r := mvRunIn(plain, nil, c.patch, c.harness)
			mvCheck(t, "not a worktree: exits 4", r.rc == 4, "rc=%d out=%s", r.rc, r.all())
			c.pinned(r, plain, "<plain>")
		}},
		{"unreadable patch", func(t *testing.T) {
			c := fx.newFixture(t, t.TempDir(), 4, "")
			c.snap()
			r := c.run(nil, "no-such.patch", c.harness)
			mvCheck(t, "unreadable patch: exits 4", r.rc == 4, "rc=%d out=%s", r.rc, r.all())
			c.unchanged("unreadable patch: status and HEAD unchanged")
			c.pinned(r)
		}},
		{"silent harness on the baseline pass", func(t *testing.T) {
			c := fx.newFixture(t, t.TempDir(), 4, "4/flip:2")
			h := c.repo + "/silent.sh"
			writeExec(t, h, "#!/usr/bin/env bash\necho nothing to report\n")
			c.snap()
			r := c.run(nil, c.patch, h)
			mvCheck(t, "silent harness on the baseline pass: exits 4", r.rc == 4, "rc=%d out=%s", r.rc, r.all())
			c.unchanged("silent harness on the baseline pass: status and HEAD unchanged")
			c.pinned(r)
		}},
		{"silent harness on the mutated pass", func(t *testing.T) {
			c := fx.newFixture(t, t.TempDir(), 4, "4/flip:2")
			h := c.repo + "/silent.sh"
			writeExec(t, h, "#!/usr/bin/env bash\ngrep -q '"+mvMutationMark+"' guard.sh || echo 'ok: baseline'\n")
			c.snap()
			r := c.run(nil, c.patch, h)
			mvCheck(t, "silent harness on the mutated pass: exits 4", r.rc == 4, "rc=%d out=%s", r.rc, r.all())
			c.unchanged("silent harness on the mutated pass: status and HEAD unchanged")
			c.pinned(r)
		}},
		// Relative arguments resolve against the caller's cwd, a
		// subdirectory here, before the run moves to the repository root.
		{"relative arguments", func(t *testing.T) {
			c := fx.newFixture(t, t.TempDir(), 4, "4/flip:2")
			mkdir(t, c.repo+"/sub")
			r := mvRunIn(c.repo+"/sub", nil, "../../patch", "../test-fixture.sh", "../no-such.sh")
			mvCheck(t, "relative arguments: a missing harness is named by its joined path",
				r.rc == 2 && has(r.err, c.repo+"/sub/../no-such.sh"), "rc=%d out=%s", r.rc, r.all())
			r = mvRunIn(c.repo+"/sub", nil, "../../patch", "../test-fixture.sh")
			mvCheck(t, "relative arguments: caught from a subdirectory", r.rc == 0 && has(r.out, "caught"), "rc=%d out=%s", r.rc, r.all())
			c.pinned(r)
		}},
		{"path with a space", func(t *testing.T) {
			c := fx.newFixture(t, t.TempDir()+"/a dir", 4, "4/flip:2")
			r := c.run(nil, c.patch, c.harness)
			mvCheck(t, "path with a space: caught, exit 0", r.rc == 0 && has(r.out, "caught — 1 new failure(s): case-2"), "rc=%d out=%s", r.rc, r.all())
			mvCheck(t, "path with a space: touched file clean after restore", c.clean(), "dirty")
			c.pinned(r)
		}},
		// The real shim, no go on PATH: the loader cannot build flow-guard
		// and exits the cannot-answer code, 4 — shaped as check-task-commit-
		// fields' case 56.
		{"shim without go", func(t *testing.T) {
			cmd := exec.Command("/bin/bash", tcfScriptsDir(t)+"/mutate-and-verify.sh", "p", "h")
			cmd.Env = []string{"PATH=/usr/bin:/bin", "FLOW_GUARD_CACHE_DIR=" + t.TempDir()}
			out, err := cmd.CombinedOutput()
			if cmd.ProcessState == nil {
				t.Fatal(err)
			}
			mvCheck(t, "shim without go: exits 4", cmd.ProcessState.ExitCode() == 4, "rc=%d out=%s", cmd.ProcessState.ExitCode(), out)
			mvCheck(t, "shim without go: names the missing go",
				has(string(out), "mutate-and-verify: cannot build flow-guard from", "no go on PATH"), "%s", out)
		}},
		// A SIGTERM mid-run still restores, as the bash EXIT trap did, and
		// the process then dies of that signal; so does a SIGPIPE, which
		// bash 3.2's trap caught too. A SIGHUP ignored at start stays
		// ignored: the mutated pass runs on and the run ends normally.
		{"signal mid-run", func(t *testing.T) { mvSignalCase(t, fx, syscall.SIGTERM, false) }},
		{"SIGPIPE mid-run", func(t *testing.T) { mvSIGPIPECase(t, fx) }},
		// A sent SIGABRT, which the Go runtime would answer with a stack
		// dump and exit 2, restores and exits 134, as bash's trap and death
		// read to a shell.
		{"SIGABRT mid-run", func(t *testing.T) { mvSignalCase(t, fx, syscall.SIGABRT, false) }},
		{"EPIPE in git apply's output", func(t *testing.T) { mvApplyEPIPECase(t, fx) }},
		// Once a signal is being handled, the patch is never applied: the
		// restore may already have run.
		{"no apply once a signal is being handled", func(t *testing.T) {
			c := fx.newFixture(t, t.TempDir(), 4, "4/flip:2")
			r := &mvRun{root: c.repo, stderr: io.Discard}
			r.dying.Store(true)
			_, ok := r.apply(io.Discard, c.patch)
			mvCheck(t, "dying: nothing applied", !ok && c.clean(), "ok=%v clean=%v", ok, c.clean())
			r.dying.Store(false)
			rc, ok := r.apply(io.Discard, c.patch)
			mvCheck(t, "not dying: applied", ok && rc == 0 && !c.clean(), "ok=%v rc=%d", ok, rc)
		}},
		{"ignored SIGHUP mid-run", func(t *testing.T) { mvSignalCase(t, fx, syscall.SIGHUP, true) }},
		// Mechanisms the bash got implicitly, each run beside the bash at
		// d71a2327 on its own copy of the fixture: $(...) dropping NUL bytes,
		// a bare `FAIL: ` taking no part in the difference, and a bound
		// `[ -le ]` cannot read falling to FLAG.
		{"port: NUL bytes in harness output are dropped", func(t *testing.T) {
			mvParity(t, fx, nil, "#!/usr/bin/env bash\n"+
				"if grep -q '"+mvMutationMark+"' guard.sh; then printf 'ok: a\\n\\0ok: z\\nFAIL: x\\0y\\n'\n"+
				"else printf '\\0ok: only\\n'; fi\n")
		}},
		{"port: a bare FAIL: line is no new failure", func(t *testing.T) {
			mvParity(t, fx, nil, "#!/usr/bin/env bash\necho 'ok: a'\n"+
				"if grep -q '"+mvMutationMark+"' guard.sh; then echo 'FAIL: '; fi\n")
		}},
		{"port: an out-of-range bound falls to FLAG", func(t *testing.T) {
			mvParity(t, fx, map[string]string{"MUTATE_AND_VERIFY_MAX_NEW_FAILURES": "99999999999999999999"}, "")
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			c.fn(t)
		})
	}
}

// mvSignalCase runs the real shim on a harness that, on the mutated pass,
// reports through one FIFO and then blocks on another; the test sends sig
// once it has reported, then releases it. With ignore, sig is ignored from
// the start (`trap ” <sig>` before the exec), the mutation must still be
// in place after it, and the run must end normally.
func mvSignalCase(t *testing.T, fx mvFixtures, sig syscall.Signal, ignore bool) {
	dir := t.TempDir()
	c := fx.newFixture(t, dir, 4, "4/flip:2")
	ready, release := dir+"/ready", dir+"/release"
	for _, p := range []string{ready, release} {
		if err := syscall.Mkfifo(p, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	h := c.repo + "/block.sh"
	writeExec(t, h, "#!/usr/bin/env bash\necho 'ok: case-1'\n"+
		"if grep -q '"+mvMutationMark+"' guard.sh; then echo ready > \"$MV_READY\"; read -r _ < \"$MV_RELEASE\"; fi\n")
	c.snap()
	shim := tcfScriptsDir(t) + "/mutate-and-verify.sh"
	cmd := exec.Command("/bin/bash", shim, c.patch, h)
	if ignore {
		cmd = exec.Command("/bin/bash", "-c", fmt.Sprintf(`trap '' %d; exec /bin/bash "$0" "$@"`, int(sig)), shim, c.patch, h)
	}
	cmd.Dir = c.repo
	cmd.Env = append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "MV_READY="+ready, "MV_RELEASE="+release)
	// Files, not buffers: Wait then waits for the process alone, never for
	// a pipe an orphaned harness might still hold.
	out, errf := mvCreate(t, dir+"/stdout"), mvCreate(t, dir+"/stderr")
	cmd.Stdout, cmd.Stderr = out, errf
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.ReadFile(ready); err != nil { // blocks until the harness writes
		t.Fatal(err)
	}
	if err := cmd.Process.Signal(sig); err != nil {
		t.Fatal(err)
	}
	if ignore {
		// An ignored signal leaves no trace to wait on; half a second is
		// long past a handler's restore, which would undo the mutation
		// under the running harness.
		time.Sleep(500 * time.Millisecond)
		g, _ := os.ReadFile(c.guard)
		mvCheck(t, "ignored signal: the mutation is still applied", strings.Contains(string(g), mvMutationMark), "restored early")
		if f, err := os.OpenFile(release, os.O_WRONLY, 0); err == nil {
			f.Close()
		}
		_ = cmd.Wait()
		o, _ := os.ReadFile(dir + "/stdout")
		mvCheck(t, "ignored signal: the run ends normally", cmd.ProcessState.ExitCode() == 0 && has(string(o), "ran clean"),
			"status %v\n%s", cmd.ProcessState, o)
		c.unchanged("ignored signal: status and HEAD unchanged")
		return
	}
	_ = cmd.Wait()
	if f, err := os.OpenFile(release, os.O_WRONLY, 0); err == nil { // unblock the orphaned harness
		f.Close()
	}
	ws, _ := cmd.ProcessState.Sys().(syscall.WaitStatus)
	if sig == syscall.SIGHUP || sig == syscall.SIGINT || sig == syscall.SIGTERM {
		mvCheck(t, "signal mid-run: dies of "+sig.String(), ws.Signaled() && ws.Signal() == sig, "status %v", cmd.ProcessState)
	} else {
		mvCheck(t, "signal mid-run: exits 128+"+sig.String(), ws.Exited() && ws.ExitStatus() == 128+int(sig), "status %v", cmd.ProcessState)
	}
	c.unchanged("signal mid-run: status and HEAD unchanged")
	o, _ := os.ReadFile(dir + "/stdout")
	e, _ := os.ReadFile(dir + "/stderr")
	if sig == syscall.SIGTERM {
		c.pinned(mvRes{-1, string(o), string(e)})
	}
}

// mvSIGPIPECase runs the real shim with its stdout on a pipe the test
// closes once the mutated pass has begun, as `| head -n 7` would; the next
// line the guard prints raises SIGPIPE. The touched file is restored and the
// guard exits 141 -- where bash 5.3 died of the SIGPIPE and left the
// mutation in place, and bash 3.2's trap restored it.
func mvSIGPIPECase(t *testing.T, fx mvFixtures) {
	dir := t.TempDir()
	c := fx.newFixture(t, dir, 4, "4/flip:2")
	release := dir + "/release"
	if err := syscall.Mkfifo(release, 0o600); err != nil {
		t.Fatal(err)
	}
	h := c.repo + "/block.sh"
	writeExec(t, h, "#!/usr/bin/env bash\necho 'ok: case-1'\n"+
		"if grep -q '"+mvMutationMark+"' guard.sh; then read -r _ < \"$MV_RELEASE\"; fi\n")
	// A second harness marks its mutated pass: bash died at the failed
	// write, so it never started.
	h2 := c.repo + "/second.sh"
	writeExec(t, h2, "#!/usr/bin/env bash\necho 'ok: case-1'\n"+
		"if grep -q '"+mvMutationMark+"' guard.sh; then touch \"$MV_SECOND\"; fi\n")
	second := dir + "/second-ran"
	c.snap()
	cmd := exec.Command("/bin/bash", tcfScriptsDir(t)+"/mutate-and-verify.sh", c.patch, h, h2)
	cmd.Dir = c.repo
	cmd.Env = append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "MV_RELEASE="+release, "MV_SECOND="+second)
	cmd.Stderr = mvCreate(t, dir+"/stderr")
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stdout = pw
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	pw.Close()
	var seen strings.Builder
	buf := make([]byte, 4096)
	for !strings.Contains(seen.String(), "mutated pass (after mutation)\n") {
		n, err := pr.Read(buf)
		if err != nil {
			t.Fatalf("stdout ended before the mutated pass: %v\n%s", err, seen.String())
		}
		seen.Write(buf[:n])
	}
	pr.Close()
	if f, err := os.OpenFile(release, os.O_WRONLY, 0); err == nil {
		f.Close()
	}
	_ = cmd.Wait()
	mvCheck(t, "SIGPIPE mid-run: exits 141", cmd.ProcessState.ExitCode() == 141, "status %v", cmd.ProcessState)
	_, err = os.Stat(second)
	mvCheck(t, "SIGPIPE mid-run: the next harness never starts its mutated pass", os.IsNotExist(err), "it ran")
	c.unchanged("SIGPIPE mid-run: status and HEAD unchanged")
}

// mvApplyEPIPECase runs the real shim with a `git` first on PATH that
// prints while applying the patch, and closes stdout right after the last
// line before the apply: the EPIPE lands in `git apply`'s output, copied
// while the restore lock is held. The trap must still restore and exit
// 141, as the bash did, not hang.
func mvApplyEPIPECase(t *testing.T, fx mvFixtures) {
	dir := t.TempDir()
	c := fx.newFixture(t, dir, 4, "4/flip:2")
	stub := dir + "/bin"
	mkdir(t, stub)
	writeExec(t, stub+"/git", "#!/usr/bin/env bash\n"+
		"if [ \"$1\" = apply ] && [ \"$2\" = --whitespace=nowarn ]; then yes applying | head -n 100000; fi\n"+
		"exec "+fixtureGit+" \"$@\"\n")
	c.snap()
	cmd := exec.Command("/bin/bash", tcfScriptsDir(t)+"/mutate-and-verify.sh", c.patch, c.harness)
	cmd.Dir = c.repo
	cmd.Env = append(os.Environ(), "PATH="+stub+":"+os.Getenv("PATH"), "FLOW_GUARD_CACHE_DIR="+guardCache(t))
	cmd.Stderr = mvCreate(t, dir+"/stderr")
	pr, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	cmd.Stdout = pw
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	pw.Close()
	var seen strings.Builder
	buf := make([]byte, 4096)
	for !strings.Contains(seen.String(), "baseline ok=") {
		n, err := pr.Read(buf)
		if err != nil {
			t.Fatalf("stdout ended before the baseline verdict: %v\n%s", err, seen.String())
		}
		seen.Write(buf[:n])
	}
	pr.Close()
	done := make(chan struct{})
	go func() { _ = cmd.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(60 * time.Second):
		_ = cmd.Process.Kill()
		<-done
		t.Fatal("hung after an EPIPE in git apply's output")
	}
	mvCheck(t, "EPIPE in git apply's output: exits 141", cmd.ProcessState.ExitCode() == 141, "status %v", cmd.ProcessState)
	c.unchanged("EPIPE in git apply's output: status and HEAD unchanged")
}

// bashAtBase writes each scripts/<rel> as it stood at d71a2327 into one
// temporary directory, keeping rel's layout, and returns that directory.
func bashAtBase(t *testing.T, rels ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, rel := range rels {
		src, err := exec.Command(fixtureGit, "-C", "../../..", "show", "d71a2327:scripts/"+rel).Output()
		if err != nil {
			t.Fatalf("git show d71a2327:scripts/%s: %v", rel, err)
		}
		writeFile(t, filepath.Join(dir, rel), string(src))
	}
	return dir
}

// mvParity runs the bash at d71a2327 and the port on separate copies of
// the 4-case flip:2 fixture, with vars in the environment and harness (the
// fixture's own when empty), and asserts stdout, stderr and exit are the
// same. Bash 5's "ignored null byte" warning and bash's `[: ... integer`
// diagnostic are dropped from its stderr: the port records both as the
// bytes it does not reproduce.
func mvParity(t *testing.T, fx mvFixtures, vars map[string]string, harness string) {
	t.Helper()
	run := func(bash bool) mvRes {
		c := fx.newFixture(t, t.TempDir(), 4, "4/flip:2")
		h := c.harness
		if harness != "" {
			h = c.repo + "/h.sh"
			writeExec(t, h, harness)
		}
		var r mvRes
		if bash {
			cmd := exec.Command("bash", bashAtBase(t, "mutate-and-verify.sh", "lib/post-mutation-check.sh")+"/mutate-and-verify.sh", c.patch, h)
			cmd.Dir = c.repo
			cmd.Env = os.Environ()
			for k, v := range vars {
				cmd.Env = append(cmd.Env, k+"="+v)
			}
			var out, errb bytes.Buffer
			cmd.Stdout, cmd.Stderr = &out, &errb
			_ = cmd.Run()
			var kept []string
			for _, l := range strings.SplitAfter(errb.String(), "\n") {
				if !strings.Contains(l, "ignored null byte") && !strings.Contains(l, ": integer ex") {
					kept = append(kept, l)
				}
			}
			r = mvRes{cmd.ProcessState.ExitCode(), out.String(), strings.Join(kept, "")}
		} else {
			r = c.run(vars, c.patch, h)
		}
		norm := strings.NewReplacer(c.patch, "<patch>", c.repo, "<repo>")
		return mvRes{r.rc, norm.Replace(r.out), norm.Replace(r.err)}
	}
	want, got := run(true), run(false)
	mvCheck(t, "matches the bash at d71a2327", got == want, "got:\n%+v\nwant:\n%+v", got, want)
}

func mvCreate(t *testing.T, path string) *os.File {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

// mvPins is each fixture's whole result from the bash script at d71a2327,
// captured by running scripts/mutate-and-verify.sh over the same fixture.
var mvPins = map[string]string{
	"TestMutateAndVerify/10_cannot_restore":                   "mutate-and-verify: patch <patch> touches:\n  new-file.txt\n\nmutate-and-verify: baseline pass (before mutation)\n  test-fixture.sh: baseline ok=4 fail=0\n\nmutate-and-verify: mutated pass (after mutation)\n  test-fixture.sh: mutated ok=4 fail=0\n\nmutate-and-verify: verdict (blast-radius bound: 5)\n  test-fixture.sh: surviving mutant — 0 new failures\n\nmutate-and-verify: restoring touched files\n--- stderr\nmutate-and-verify: could not fully restore — residual status:\n?? new-file.txt\nmutate-and-verify: post-restore drift detected:\nunexpected status line: ? new-file.txt\n--- exit 3\n",
	"TestMutateAndVerify/11_stash_residue":                    "mutate-and-verify: patch <patch> touches:\n  guard.sh\n\nmutate-and-verify: baseline pass (before mutation)\n  stash-harness.sh: baseline ok=1 fail=0\n\nmutate-and-verify: mutated pass (after mutation)\n  stash-harness.sh: mutated ok=1 fail=0\n\nmutate-and-verify: verdict (blast-radius bound: 5)\n  stash-harness.sh: surviving mutant — 0 new failures\n\nmutate-and-verify: restoring touched files\nmutate-and-verify: ran clean — touched files restored\n--- stderr\nmutate-and-verify: post-restore drift detected:\nnew stash entry: stash@{0}: On main: kan-448-residue\n--- exit 2\n",
	"TestMutateAndVerify/12_residual_keeps_exit_3":            "mutate-and-verify: patch <patch> touches:\n  guard.sh\n\nmutate-and-verify: baseline pass (before mutation)\n  stash-harness.sh: baseline ok=1 fail=0\n\nmutate-and-verify: mutated pass (after mutation)\n  stash-harness.sh: mutated ok=1 fail=0\n\nmutate-and-verify: verdict (blast-radius bound: 5)\n  stash-harness.sh: surviving mutant — 0 new failures\n\nmutate-and-verify: restoring touched files\n--- stderr\nmutate-and-verify: could not fully restore — residual status:\nM  guard.sh\nmutate-and-verify: post-restore drift detected:\nnew stash entry: stash@{0}: On main: kan-448-residue\nunexpected status line: 1 M. N... 100755 100755 100755 8705ad30ffc1113fc5ec461b1188a7a1ed47bb62 216ae20b9cce71cd3b04b59eb89a4ed9f26a989b guard.sh\n--- exit 3\n",
	"TestMutateAndVerify/1_caught":                            "mutate-and-verify: patch <patch> touches:\n  guard.sh\n\nmutate-and-verify: baseline pass (before mutation)\n  test-fixture.sh: baseline ok=4 fail=0\n\nmutate-and-verify: mutated pass (after mutation)\n  test-fixture.sh: mutated ok=3 fail=1\n\nmutate-and-verify: verdict (blast-radius bound: 5)\n  test-fixture.sh: caught — 1 new failure(s): case-2\n\nmutate-and-verify: restoring touched files\nmutate-and-verify: ran clean — touched files restored\n--- stderr\n--- exit 0\n",
	"TestMutateAndVerify/2_surviving_mutant":                  "mutate-and-verify: patch <patch> touches:\n  guard.sh\n\nmutate-and-verify: baseline pass (before mutation)\n  test-fixture.sh: baseline ok=4 fail=0\n\nmutate-and-verify: mutated pass (after mutation)\n  test-fixture.sh: mutated ok=4 fail=0\n\nmutate-and-verify: verdict (blast-radius bound: 5)\n  test-fixture.sh: surviving mutant — 0 new failures\n\nmutate-and-verify: restoring touched files\nmutate-and-verify: ran clean — touched files restored\n--- stderr\n--- exit 0\n",
	"TestMutateAndVerify/3_blast_radius":                      "mutate-and-verify: patch <patch> touches:\n  guard.sh\n\nmutate-and-verify: baseline pass (before mutation)\n  test-fixture.sh: baseline ok=12 fail=0\n\nmutate-and-verify: mutated pass (after mutation)\n  test-fixture.sh: mutated ok=6 fail=6\n\nmutate-and-verify: verdict (blast-radius bound: 5)\n  test-fixture.sh: FLAG suspicious blast radius — 6 new failures (bound 5): case-2 case-4 case-6 case-8 case-10 case-12\n\nmutate-and-verify: restoring touched files\nmutate-and-verify: ran clean — touched files restored\n--- stderr\n--- exit 0\n",
	"TestMutateAndVerify/4_bound_override":                    "mutate-and-verify: patch <patch> touches:\n  guard.sh\n\nmutate-and-verify: baseline pass (before mutation)\n  test-fixture.sh: baseline ok=12 fail=0\n\nmutate-and-verify: mutated pass (after mutation)\n  test-fixture.sh: mutated ok=6 fail=6\n\nmutate-and-verify: verdict (blast-radius bound: 10)\n  test-fixture.sh: caught — 6 new failure(s): case-2 case-4 case-6 case-8 case-10 case-12\n\nmutate-and-verify: restoring touched files\nmutate-and-verify: ran clean — touched files restored\n--- stderr\n--- exit 0\n",
	"TestMutateAndVerify/5_patch_does_not_apply":              "--- stderr\nmutate-and-verify: git apply --check failed:\nerror: nonexistent.sh: No such file or directory\nmutate-and-verify: refused — <patch> does not apply cleanly — nothing was mutated\n--- exit 2\n",
	"TestMutateAndVerify/6_dirty_touched_file":                "--- stderr\nmutate-and-verify: touched files already have uncommitted changes:\n M guard.sh\nmutate-and-verify: refused — a file the patch touches is not clean — nothing was mutated\n--- exit 2\n",
	"TestMutateAndVerify/7_missing_harness":                   "--- stderr\nmutate-and-verify: refused — harness does not exist: <repo>/no-such-harness.sh — nothing was mutated\n--- exit 2\n",
	"TestMutateAndVerify/8_non-executable_harness":            "--- stderr\nmutate-and-verify: refused — harness is not executable: <repo>/test-fixture-noexec.sh — nothing was mutated\n--- exit 2\n",
	"TestMutateAndVerify/9_boundary":                          "mutate-and-verify: patch <patch> touches:\n  guard.sh\n\nmutate-and-verify: baseline pass (before mutation)\n  test-fixture.sh: baseline ok=10 fail=0\n\nmutate-and-verify: mutated pass (after mutation)\n  test-fixture.sh: mutated ok=5 fail=5\n\nmutate-and-verify: verdict (blast-radius bound: 5)\n  test-fixture.sh: caught — 5 new failure(s): case-2 case-4 case-6 case-8 case-10\n\nmutate-and-verify: restoring touched files\nmutate-and-verify: ran clean — touched files restored\n--- stderr\n--- exit 0\n",
	"TestMutateAndVerify/not_a_worktree":                      "--- stderr\nmutate-and-verify: cannot answer — not inside a git worktree: <plain>\n--- exit 4\n",
	"TestMutateAndVerify/path_with_a_space":                   "mutate-and-verify: patch <patch> touches:\n  guard.sh\n\nmutate-and-verify: baseline pass (before mutation)\n  test-fixture.sh: baseline ok=4 fail=0\n\nmutate-and-verify: mutated pass (after mutation)\n  test-fixture.sh: mutated ok=3 fail=1\n\nmutate-and-verify: verdict (blast-radius bound: 5)\n  test-fixture.sh: caught — 1 new failure(s): case-2\n\nmutate-and-verify: restoring touched files\nmutate-and-verify: ran clean — touched files restored\n--- stderr\n--- exit 0\n",
	"TestMutateAndVerify/relative_arguments":                  "mutate-and-verify: patch ../../patch touches:\n  guard.sh\n\nmutate-and-verify: baseline pass (before mutation)\n  test-fixture.sh: baseline ok=4 fail=0\n\nmutate-and-verify: mutated pass (after mutation)\n  test-fixture.sh: mutated ok=3 fail=1\n\nmutate-and-verify: verdict (blast-radius bound: 5)\n  test-fixture.sh: caught — 1 new failure(s): case-2\n\nmutate-and-verify: restoring touched files\nmutate-and-verify: ran clean — touched files restored\n--- stderr\n--- exit 0\n",
	"TestMutateAndVerify/signal_mid-run":                      "mutate-and-verify: patch <patch> touches:\n  guard.sh\n\nmutate-and-verify: baseline pass (before mutation)\n  block.sh: baseline ok=1 fail=0\n\nmutate-and-verify: mutated pass (after mutation)\nmutate-and-verify: ran clean — touched files restored\n--- stderr\n--- exit -1\n",
	"TestMutateAndVerify/silent_harness_on_the_baseline_pass": "mutate-and-verify: patch <patch> touches:\n  guard.sh\n\nmutate-and-verify: baseline pass (before mutation)\n--- stderr\nmutate-and-verify: cannot answer — harness silent.sh produced no ok:/FAIL: line on the baseline pass\n--- exit 4\n",
	"TestMutateAndVerify/silent_harness_on_the_mutated_pass":  "mutate-and-verify: patch <patch> touches:\n  guard.sh\n\nmutate-and-verify: baseline pass (before mutation)\n  silent.sh: baseline ok=1 fail=0\n\nmutate-and-verify: mutated pass (after mutation)\n--- stderr\nmutate-and-verify: cannot answer — harness silent.sh produced no ok:/FAIL: line on the mutated pass\n--- exit 4\n",
	"TestMutateAndVerify/unreadable_patch":                    "--- stderr\nmutate-and-verify: cannot answer — cannot read patch file: no-such.patch\n--- exit 4\n",
	"TestMutateAndVerify/usage":                               "--- stderr\nmutate-and-verify: usage: mutate-and-verify.sh <patch-file> <harness>...\n--- exit 4\n",
}
