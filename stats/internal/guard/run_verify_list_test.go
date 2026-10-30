package guard

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
)

// TestRunVerifyList carries the cases of the retired
// scripts/test-run-verify-list.sh, in its order and over its fixtures. Every
// step runs twice, on twin roots: through the Go guard and through the bash
// at 8c54d6d3, whose project-get.sh and check-spec-reach.sh siblings are the
// live shims. Each step asserts the harness's expectations on the Go output
// and byte-for-byte stdout/stderr/exit parity with the bash, each twin's root
// written as <ROOT>.
func TestRunVerifyList(t *testing.T) {
	t.Parallel()
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	bashDir := t.TempDir()
	writeExec(t, filepath.Join(bashDir, "run-verify-list.sh"),
		string(historicalFile(t, repo, "8c54d6d3", "scripts/run-verify-list.sh")))
	for _, l := range []string{"lib", "project-get.sh", "check-spec-reach.sh"} {
		if err := os.Symlink(filepath.Join(repo, "scripts", l), filepath.Join(bashDir, l)); err != nil {
			t.Fatal(err)
		}
	}
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t))
	goEnviron := append(environ[:len(environ):len(environ)], "FLOW_GUARD_SCRIPT_DIR="+filepath.Join(repo, "scripts"))
	getenv := func(k string) string {
		for i := len(goEnviron) - 1; i >= 0; i-- {
			if v, ok := strings.CutPrefix(goEnviron[i], k+"="); ok {
				return v
			}
		}
		return ""
	}
	gitIn := func(t *testing.T, dir string, args ...string) {
		t.Helper()
		cmd := exec.Command(fixtureGit, append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@t"}, args...)...)
		// A pinned date gives both twins the same shas, so the skip lines
		// naming one compare byte for byte.
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE=2026-01-01T00:00:00Z", "GIT_COMMITTER_DATE=2026-01-01T00:00:00Z")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	gitOutIn := func(t *testing.T, dir string, args ...string) string {
		t.Helper()
		out, err := exec.Command(fixtureGit, append([]string{"-C", dir}, args...)...).Output()
		if err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
		return strings.TrimRight(string(out), "\n")
	}

	// Two twin roots, built as the harness built $ROOT; the root's physical
	// path, so `pwd` and git's absolute paths agree with it.
	newRoot := func() string {
		r, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(r, ".flow/project.md"), "## lint\n\nProse outside the fence is never run.\n\n```bash\n# a comment line is never run\nprintf 'one\\n'\nprintf 'to-stderr\\n' >&2; exit 3\ncat\nseq 1 25\npwd\n```\n\n## test\n\n```bash\nprintf 'ok\\n'\nprintf nonl\necho next\n```\n")
		return r
	}
	gr, br := newRoot(), newRoot()
	both := func(f func(root string)) { f(gr); f(br) }

	type res struct {
		rc  int
		out string // stdout then stderr, <ROOT> for the twin's root
	}
	run := func(t *testing.T, args ...string) res {
		t.Helper()
		norm := func(root string, s ...string) string { return strings.ReplaceAll(strings.Join(s, ""), root, "<ROOT>") }
		gargs := make([]string, len(args))
		for i, a := range args {
			gargs[i] = strings.ReplaceAll(a, "<ROOT>", gr)
		}
		var o, e bytes.Buffer
		grc := Registry["run-verify-list"](gargs, Env{Getenv: getenv, Environ: func() []string { return goEnviron },
			Dir: gr, Stdin: strings.NewReader("LEAKED-STDIN\n")}, &o, &e)
		// The recording arguments postdate the bash (kan-1033): its usage line lacks them.
		g := res{grc, strings.Replace(norm(gr, o.String(), e.String()), " [<change> <key> <session-token> <model> <effort> [<canonical-worktree>]]", "", 1)}

		bargs := make([]string, len(args))
		for i, a := range args {
			bargs[i] = strings.ReplaceAll(a, "<ROOT>", br)
		}
		cmd := exec.Command("bash", append([]string{filepath.Join(bashDir, "run-verify-list.sh")}, bargs...)...)
		cmd.Dir, cmd.Env, cmd.Stdin = br, append(environ[:len(environ):len(environ)], "PWD="+br), strings.NewReader("LEAKED-STDIN\n")
		var bo, be bytes.Buffer
		cmd.Stdout, cmd.Stderr = &bo, &be
		brc := 0
		if err := cmd.Run(); err != nil {
			var ee *exec.ExitError
			if !errors.As(err, &ee) {
				t.Fatal(err)
			}
			brc = ee.ExitCode()
		}
		if b := (res{brc, norm(br, bo.String(), be.String())}); g != b {
			t.Errorf("port rc=%d out=%q\nbash rc=%d out=%q", g.rc, g.out, b.rc, b.out)
		}
		return g
	}
	rc := func(t *testing.T, r res, want int) {
		t.Helper()
		if r.rc != want {
			t.Errorf("exit %d, want %d; out=%q", r.rc, want, r.out)
		}
	}
	has := func(t *testing.T, r res, s string) {
		t.Helper()
		if !strings.Contains(r.out, s) {
			t.Errorf("missing %q in %q", s, r.out)
		}
	}
	lacks := func(t *testing.T, r res, s string) {
		t.Helper()
		if strings.Contains(r.out, s) {
			t.Errorf("unexpected %q in %q", s, r.out)
		}
	}
	project := func(dir, body string) { writeFile(t, filepath.Join(dir, ".flow/project.md"), body) }

	t.Run("a failing command fails the list", func(t *testing.T) {
		r := run(t, "<ROOT>", "lint")
		rc(t, r, 1)
		has(t, r, "- `printf 'one\\n'` — exit 0\n  one")
		has(t, r, "exit 3\n  to-stderr")
		has(t, r, "- `pwd` — exit 0")
		lacks(t, r, "LEAKED-STDIN")
		lacks(t, r, "comment line")
		has(t, r, "(truncated — last 20 of 25 lines)\n  6\n")
		lacks(t, r, "\n  5\n")
		has(t, r, "  <ROOT>")
		has(t, r, "- `check-spec-reach.sh <ROOT>` — exit 0\n  Spec reach: not configured")
		lacks(t, r, "full output:")
	})

	t.Run("a passing list", func(t *testing.T) {
		r := run(t, "<ROOT>", "test")
		rc(t, r, 0)
		has(t, r, "- `printf 'ok\\n'` — exit 0\n  ok")
		has(t, r, "  nonl\n- `echo next` — exit 0")
		lacks(t, r, "check-spec-reach.sh")
	})

	// Beyond the harness: the threshold itself, which 25 lines never pinned.
	t.Run("21 lines are truncated", func(t *testing.T) {
		both(func(root string) { project(root, "## test\n\n```bash\nseq 1 21\nseq 1 20\n```\n") })
		r := run(t, "<ROOT>", "test")
		has(t, r, "(truncated — last 20 of 21 lines)\n  2\n")
		has(t, r, "- `seq 1 20` — exit 0\n  1\n")
	})

	t.Run("a list with a race fails", func(t *testing.T) {
		both(func(root string) {
			project(root, "## test\n\n```bash\nprintf 'WARNING: DATA RACE\\n'; seq 1 25; exit 1\nprintf 'testing.go:1: race detected during execution of test\\n'; exit 1\nprintf 'plain\\n'; exit 1\n```\n")
		})
		r := run(t, "<ROOT>", "test")
		rc(t, r, 1)
		has(t, r, "  25\n  data race — never environmental")
		if n := strings.Count(r.out, "data race — never environmental"); n != 2 {
			t.Errorf("race marks: %d", n)
		}
	})

	// The first command waits for a file only the second creates: run one
	// after the other it gives up after five seconds and fails the list.
	t.Run("test commands run concurrently", func(t *testing.T) {
		both(func(root string) {
			project(root, "## test\n\n```bash\nfor i in $(seq 50); do [ -f made-by-two ] && echo saw-two && exit 0; sleep 0.1; done; exit 1\ntouch made-by-two\n```\n")
		})
		r := run(t, "<ROOT>", "test")
		both(func(root string) { os.Remove(filepath.Join(root, "made-by-two")) })
		rc(t, r, 0)
		has(t, r, "  saw-two\n- `touch made-by-two` — exit 0")
	})

	t.Run("an undeclared list", func(t *testing.T) {
		both(func(root string) { project(root, "## lint\n\n```bash\ntrue\n```\n") })
		rc(t, run(t, "<ROOT>", "test"), 3)
	})

	t.Run("a placeholder line refuses the list", func(t *testing.T) {
		both(func(root string) {
			project(root, "## test\n\n```bash\nprintf ran-first\n./gradlew verifyChange :<module>:test --tests '<selectors>'\n```\n")
		})
		r := run(t, "<ROOT>", "test")
		rc(t, r, 2)
		has(t, r, "test line is a template, not a command — move it out of the fence: ./gradlew verifyChange :<module>:test --tests '<selectors>'")
		lacks(t, r, "ran-first")
	})

	// THE GREEN SHA: a git worktree whose `test` list counts its own runs.
	t.Run("the green sha", func(t *testing.T) {
		both(func(root string) {
			g := filepath.Join(root, "git")
			project(g, "## test\n\n```bash\n# inputs: web/\necho ran >> \"$(git rev-parse --git-dir)/web-runs\"; echo web\n# inputs: go/\necho go\necho all\n```\n")
			writeFile(t, filepath.Join(g, "web/a"), "a\n")
			writeFile(t, filepath.Join(g, "go/b"), "b\n")
			gitIn(t, g, "init", "-q")
			gitIn(t, g, "add", "-A")
			gitIn(t, g, "commit", "-q", "-m", "one")
		})
		g := filepath.Join(gr, "git")
		sha1 := gitOutIn(t, g, "rev-parse", "HEAD")
		if b := gitOutIn(t, filepath.Join(br, "git"), "rev-parse", "HEAD"); b != sha1 {
			t.Fatalf("twin shas differ: %s, %s", sha1, b)
		}

		r := run(t, "<ROOT>/git", "test")
		rc(t, r, 0)
		has(t, r, "- `echo all` — exit 0")
		green := gitOutIn(t, g, "rev-parse", "--path-format=absolute", "--git-path", "flow-test-green")
		if b, _ := os.ReadFile(green); strings.TrimSpace(string(b)) != sha1 {
			t.Errorf("green sha: %q", b)
		}
		if st := gitOutIn(t, g, "status", "--porcelain"); st != "" {
			t.Errorf("the record dirtied the tree: %q", st)
		}

		r = run(t, "<ROOT>/git", "test")
		rc(t, r, 0)
		has(t, r, "- `echo all` — skipped, no input changed since "+sha1+" passed the list")
		runs, _ := os.ReadFile(filepath.Join(gitOutIn(t, g, "rev-parse", "--absolute-git-dir"), "web-runs"))
		if strings.Count(string(runs), "\n") != 1 {
			t.Errorf("a command re-ran at the green sha: %q", runs)
		}

		both(func(root string) {
			writeFile(t, filepath.Join(root, "git/go/b"), "c\n")
			gitIn(t, filepath.Join(root, "git"), "commit", "-q", "-am", "two")
		})
		r = run(t, "<ROOT>/git", "test")
		has(t, r, "- `echo ran >> \"$(git rev-parse --git-dir)/web-runs\"; echo web` — skipped, no input changed since "+sha1)
		has(t, r, "- `echo go` — exit 0")
		has(t, r, "- `echo all` — exit 0")

		both(func(root string) { writeFile(t, filepath.Join(root, "git/web/a"), "dirty\n") })
		has(t, run(t, "<ROOT>/git", "test"), "echo web` — exit 0")

		both(func(root string) { gitIn(t, filepath.Join(root, "git"), "checkout", "-q", "--", "web/a") })
		run(t, "<ROOT>/git", "test")
		both(func(root string) {
			p := filepath.Join(root, "git/.flow/project.md")
			b, _ := os.ReadFile(p)
			writeFile(t, p, strings.Replace(string(b), "\necho go\n", "\necho go2\n", 1))
			gitIn(t, filepath.Join(root, "git"), "commit", "-q", "-am", "three")
		})
		has(t, run(t, "<ROOT>/git", "test"), "- `echo go2` — exit 0")
	})

	// A failing command's full output is kept in the git dir, never the tree.
	t.Run("a failing list in a git worktree", func(t *testing.T) {
		both(func(root string) {
			f := filepath.Join(root, "fail")
			project(f, "## test\n\n```bash\nseq 1 25; exit 4\ntrue\n```\n")
			gitIn(t, f, "init", "-q")
			gitIn(t, f, "add", "-A")
			gitIn(t, f, "commit", "-q", "-m", "one")
		})
		f := filepath.Join(gr, "fail")
		r := run(t, "<ROOT>/fail", "test")
		kept := gitOutIn(t, f, "rev-parse", "--path-format=absolute", "--git-path", "flow-verify-failed") + "/test-0.log"
		rc(t, r, 1)
		has(t, r, "  full output: "+strings.ReplaceAll(kept, gr, "<ROOT>"))
		if b, _ := os.ReadFile(kept); !strings.HasPrefix(string(b), "1\n") {
			t.Errorf("kept output: %q", b)
		}
		lacks(t, r, "test-1.log")
		if st := gitOutIn(t, f, "status", "--porcelain"); st != "" {
			t.Errorf("the kept output dirtied the tree: %q", st)
		}
		both(func(root string) {
			project(filepath.Join(root, "fail"), "## test\n\n```bash\ntrue\n```\n")
			gitIn(t, filepath.Join(root, "fail"), "commit", "-q", "-am", "two")
		})
		run(t, "<ROOT>/fail", "test")
		if _, err := os.Stat(kept); err == nil {
			t.Error("a stale kept output survived")
		}
	})

	t.Run("an unknown list", func(t *testing.T) { rc(t, run(t, "<ROOT>", "build"), 2) })
	t.Run("a missing worktree", func(t *testing.T) { rc(t, run(t, "<ROOT>/missing", "lint"), 2) })
}

// TestRunVerifyListRecordsTheVerifyRow pins kan-1033's self-review: given
// the five trailing arguments the list records its own verify row — begin
// before its first command, end once its commands ran (TestRunVerifyListEndsEveryRun
// pins each outcome).
func TestRunVerifyListRecordsTheVerifyRow(t *testing.T) {
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	writeExec(t, filepath.Join(dir, "flow"), "#!/bin/sh\necho \"flow $*\" >> "+log+"\n")
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	path := dir + ":" + os.Getenv("PATH")
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "FLOW_GUARD_SCRIPT_DIR="+filepath.Join(repo, "scripts"), "PATH="+path)
	getenv := func(k string) string {
		for i := len(environ) - 1; i >= 0; i-- {
			if v, ok := strings.CutPrefix(environ[i], k+"="); ok {
				return v
			}
		}
		return ""
	}
	run := func(args ...string) (int, string) {
		os.Remove(log)
		var o, e bytes.Buffer
		rc := runVerifyList(args, Env{Getenv: getenv, Environ: func() []string { return environ }, Dir: wt}, &o, &e)
		calls, _ := os.ReadFile(log)
		if strings.Contains(string(calls), "-tokens") {
			t.Errorf("an end carries -tokens:\n%s", calls)
		}
		return rc, string(calls)
	}
	project := func(cmds string) {
		writeFile(t, filepath.Join(wt, ".flow/project.md"), "## test\n\n```bash\n"+cmds+"\n```\n")
	}
	rec := []string{"kan-1", "verify", "mf-k1", "opus", "low"}
	begin := "flow record dispatch begin -C " + wt + " -change kan-1 -role verifier -model opus -effort low -agent-id inline -key verify -session-token mf-k1\n"
	end := "flow record dispatch end -C " + wt + " -change kan-1 -key verify -session-token mf-k1 -outcome completed\n"

	project("echo first >> " + log + "\necho last >> " + log)
	rc, calls := run(append([]string{wt, "test"}, rec...)...)
	if rc != 0 || !strings.HasPrefix(calls, begin) || !strings.HasSuffix(calls, end) || strings.Count(calls, "flow ") != 2 {
		t.Errorf("a passing list: exit %d, calls:\n%s", rc, calls)
	}

	project("echo ran >> " + log + "; exit 1")
	if rc, calls := run(append([]string{wt, "test"}, rec...)...); rc != 1 || calls != begin+"ran\n"+strings.Replace(end, "-outcome completed", "-outcome blocked -cause test-failure", 1) {
		t.Errorf("a failing list: exit %d, calls:\n%s", rc, calls)
	}

	project("true")
	if rc, calls := run(wt, "test"); rc != 0 || calls != "" {
		t.Errorf("the two-argument form: exit %d, calls:\n%s", rc, calls)
	}
	for n := 3; n <= 6; n++ {
		if rc, calls := run(append([]string{wt, "test"}, rec[:n-2]...)...); rc != 2 || calls != "" {
			t.Errorf("%d arguments: exit %d, calls %q", n, rc, calls)
		}
	}
}

// A `# suite:` line runs the command below it under `flow suite record`, so
// its runtime reaches the store; no `flow` on PATH runs it bare.
func TestRunVerifyListRecordsSuiteRuntimes(t *testing.T) {
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	writeExec(t, filepath.Join(dir, "flow"), "#!/usr/bin/env bash\necho \"$*\" >> "+log+"\nif [ \"$1 $2\" = \"suite record\" ]; then\n  while [ \"$1\" != -- ]; do shift; done\n  shift\n  exec \"$@\"\nfi\n")
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(wt, ".flow/project.md"), "## test\n\n```bash\n# suite: demo-suite\necho one\necho two\n```\n")
	run := func(path string) (int, string) {
		environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "FLOW_GUARD_SCRIPT_DIR="+filepath.Join(repo, "scripts"), "PATH="+path)
		getenv := func(k string) string {
			for i := len(environ) - 1; i >= 0; i-- {
				if v, ok := strings.CutPrefix(environ[i], k+"="); ok {
					return v
				}
			}
			return ""
		}
		var o, e bytes.Buffer
		rc := runVerifyList([]string{wt, "test"}, Env{Getenv: getenv, Environ: func() []string { return environ }, Dir: wt}, &o, &e)
		return rc, o.String()
	}
	rc, out := run(dir + ":" + os.Getenv("PATH"))
	calls, _ := os.ReadFile(log)
	if rc != 0 || !strings.Contains(out, "- `echo one` — exit 0\n  one\n") || !strings.Contains(out, "- `echo two` — exit 0\n  two\n") {
		t.Errorf("with flow: exit %d, output:\n%s", rc, out)
	}
	if want := "suite record -C " + wt + " -suite demo-suite -- " + filepath.Join(repo, "scripts", "run-capped.sh") + " -- bash -c echo one\n"; string(calls) != want {
		t.Errorf("flow calls %q, want %q", calls, want)
	}
	if rc, out := run("/usr/bin:/bin"); rc != 0 || !strings.Contains(out, "  one\n") {
		t.Errorf("without flow: exit %d, output:\n%s", rc, out)
	}
}

// TestRunVerifyListStopsOnARefusedBegin pins kan-1032's self-review: a begin
// the store refuses stops the list, exit 4, before any command; a failed end
// never changes the exit code.
func TestRunVerifyListStopsOnARefusedBegin(t *testing.T) {
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "FLOW_GUARD_SCRIPT_DIR="+filepath.Join(repo, "scripts"), "PATH="+dir+":"+os.Getenv("PATH"))
	getenv := func(k string) string {
		for i := len(environ) - 1; i >= 0; i-- {
			if v, ok := strings.CutPrefix(environ[i], k+"="); ok {
				return v
			}
		}
		return ""
	}
	writeFile(t, filepath.Join(wt, ".flow/project.md"), "## test\n\n```bash\necho ran >> "+log+"\n```\n")
	run := func(flow string) (int, string, string) {
		os.Remove(log)
		writeExec(t, filepath.Join(dir, "flow"), flow)
		var o, e bytes.Buffer
		rc := runVerifyList([]string{wt, "test", "kan-1", "verify", "mf-k1", "opus", "low"},
			Env{Getenv: getenv, Environ: func() []string { return environ }, Dir: wt}, &o, &e)
		calls, _ := os.ReadFile(log)
		return rc, string(calls), e.String()
	}

	rc, calls, stderr := run("#!/bin/sh\ncase \"$3\" in begin) echo 'flow: record dispatch refused: dirty'; exit 2;; esac\n")
	want := "run-verify-list: flow record dispatch begin refused — no command ran: flow: record dispatch refused: dirty"
	if rc != 4 || calls != "" || !strings.Contains(stderr, want) {
		t.Errorf("a refused begin: exit %d, calls %q, stderr %q", rc, calls, stderr)
	}

	if rc, calls, _ := run("#!/bin/sh\ncase \"$3\" in end) exit 1;; esac\n"); rc != 0 || calls != "ran\n" {
		t.Errorf("a failed end: exit %d, calls %q", rc, calls)
	}
}

// TestRunVerifyListRecordsAgainstTheCanonicalWorktree pins kan-1014's
// self-review: verifying a worktree that holds no spectre/changes/<name>, the
// optional eighth argument names the worktree that does, and every recording
// call carries it as -C.
func TestRunVerifyListRecordsAgainstTheCanonicalWorktree(t *testing.T) {
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	writeExec(t, filepath.Join(dir, "flow"), "#!/bin/sh\necho \"flow $*\" >> "+log+"\n")
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	canon, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(canon, "spectre/changes/kan-1/tasks.md"), "")
	writeFile(t, filepath.Join(wt, ".flow/project.md"), "## test\n\n```bash\ntrue\n```\n")
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "FLOW_GUARD_SCRIPT_DIR="+filepath.Join(repo, "scripts"), "PATH="+dir+":"+os.Getenv("PATH"))
	getenv := func(k string) string {
		for i := len(environ) - 1; i >= 0; i-- {
			if v, ok := strings.CutPrefix(environ[i], k+"="); ok {
				return v
			}
		}
		return ""
	}
	run := func(args ...string) (int, string) {
		os.Remove(log)
		var o, e bytes.Buffer
		rc := runVerifyList(args, Env{Getenv: getenv, Environ: func() []string { return environ }, Dir: wt}, &o, &e)
		calls, _ := os.ReadFile(log)
		return rc, string(calls)
	}
	rec := []string{wt, "test", "kan-1", "verify", "mf-k1", "opus", "low"}
	allCarry := func(calls, target string) bool {
		lines := strings.Split(strings.TrimSpace(calls), "\n")
		for _, l := range lines {
			if !strings.Contains(l, " -C "+target+" ") {
				return false
			}
		}
		return len(lines) == 2
	}

	if rc, calls := run(append(rec, canon)...); rc != 0 || !allCarry(calls, canon) {
		t.Errorf("eight arguments: exit %d, calls:\n%s", rc, calls)
	}
	if rc, calls := run(rec...); rc != 0 || !allCarry(calls, wt) {
		t.Errorf("seven arguments: exit %d, calls:\n%s", rc, calls)
	}
	for _, args := range [][]string{rec[:6], append(rec, canon, "extra")} {
		if rc, calls := run(args...); rc != 2 || calls != "" {
			t.Errorf("%d arguments: exit %d, calls %q", len(args), rc, calls)
		}
	}
}

// TestRunVerifyListEndsEveryRun pins kan-1017's self-review: every list that
// began its row ends it -- blocked with cause test-failure on a failure,
// skipped when every command of a test list was skipped, completed otherwise -- and no end
// carries -tokens.
func TestRunVerifyListEndsEveryRun(t *testing.T) {
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	writeExec(t, filepath.Join(dir, "flow"), "#!/bin/sh\necho \"flow $*\" >> "+log+"\n")
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "FLOW_GUARD_SCRIPT_DIR="+filepath.Join(repo, "scripts"), "PATH="+dir+":"+os.Getenv("PATH"))
	getenv := func(k string) string {
		for i := len(environ) - 1; i >= 0; i-- {
			if v, ok := strings.CutPrefix(environ[i], k+"="); ok {
				return v
			}
		}
		return ""
	}
	rec := []string{"kan-1", "verify", "mf-k1", "opus", "low"}
	list := "test"
	run := func() (int, string) {
		os.Remove(log)
		var o, e bytes.Buffer
		rc := runVerifyList(append([]string{wt, list}, rec...), Env{Getenv: getenv, Environ: func() []string { return environ }, Dir: wt}, &o, &e)
		calls, _ := os.ReadFile(log)
		return rc, string(calls)
	}
	endWith := func(outcome string) string {
		return "flow record dispatch end -C " + wt + " -change kan-1 -key verify -session-token mf-k1 " + outcome + "\n"
	}
	gitIn := func(t *testing.T, dir string, args ...string) {
		t.Helper()
		if out, err := exec.Command(fixtureGit, append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@t"}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	commit := func(cmds, msg string) {
		writeFile(t, filepath.Join(wt, ".flow/project.md"), "## test\n\n```bash\n"+cmds+"\n```\n")
		gitIn(t, wt, "add", "-A")
		gitIn(t, wt, "commit", "-q", "-m", msg)
	}
	gitIn(t, wt, "init", "-q")

	commit("exit 1", "one")
	if rc, calls := run(); rc != 1 || !strings.HasSuffix(calls, endWith("-outcome blocked -cause test-failure")) {
		t.Errorf("a failing list: exit %d, calls:\n%s", rc, calls)
	}
	commit("true", "two")
	if rc, calls := run(); rc != 0 || !strings.HasSuffix(calls, endWith("-outcome completed")) {
		t.Errorf("a green list: exit %d, calls:\n%s", rc, calls)
	}
	if rc, calls := run(); rc != 0 || !strings.HasSuffix(calls, endWith("-outcome skipped")) {
		t.Errorf("an all-skipped list: exit %d, calls:\n%s", rc, calls)
	}
	// Every test command skipped but the visual verify ran: a check ran, so the
	// row ends completed.
	writeFile(t, filepath.Join(wt, ".flow/project.md"), "## test\n\n```bash\n# inputs: go/\ntrue\n```\n\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `web/src/**` |\n| `screenshots` | `web/tests/visual` |\n\n| Command | Runs |\n|---------|------|\n| `verify` | `echo VISUAL-RAN <specs>` |\n| `capture` | `true <spec>` |\n")
	gitIn(t, wt, "add", "-A")
	gitIn(t, wt, "commit", "-q", "-m", "visual")
	if rc, calls := run(); rc != 0 || !strings.HasSuffix(calls, endWith("-outcome completed")) {
		t.Errorf("a green list with a visual verify: exit %d, calls:\n%s", rc, calls)
	}
	writeFile(t, filepath.Join(wt, "web/src/a.tsx"), "x\n")
	gitIn(t, wt, "add", "-A")
	gitIn(t, wt, "commit", "-q", "-m", "ui")
	if rc, calls := run(); rc != 0 || !strings.HasSuffix(calls, endWith("-outcome completed")) {
		t.Errorf("an all-skipped list whose visual verify ran: exit %d, calls:\n%s", rc, calls)
	}
	// A lint list whose every command is skipped still runs check-spec-reach.sh,
	// so a check ran and the row ends completed.
	list = "lint"
	writeFile(t, filepath.Join(wt, ".flow/project.md"), "## lint\n\n```bash\ntrue\n```\n")
	gitIn(t, wt, "add", "-A")
	gitIn(t, wt, "commit", "-q", "-m", "three")
	if rc, calls := run(); rc != 0 || !strings.HasSuffix(calls, endWith("-outcome completed")) {
		t.Errorf("a green lint list: exit %d, calls:\n%s", rc, calls)
	}
	if rc, calls := run(); rc != 0 || !strings.HasSuffix(calls, endWith("-outcome completed")) {
		t.Errorf("an all-skipped lint list: exit %d, calls:\n%s", rc, calls)
	}
	// A blocked lint group is never masked by the test group ending the same
	// row later, whether that test group ends skipped or completed.
	writeFile(t, filepath.Join(wt, ".flow/project.md"), "## lint\n\n```bash\nexit 1\n```\n\n## test\n\n```bash\ntrue\n```\n")
	gitIn(t, wt, "add", "-A")
	gitIn(t, wt, "commit", "-q", "-m", "four")
	list = "test"
	run()
	list = "lint"
	if rc, calls := run(); rc != 1 || !strings.HasSuffix(calls, endWith("-outcome blocked -cause test-failure")) {
		t.Errorf("a failing lint list: exit %d, calls:\n%s", rc, calls)
	}
	list = "test"
	if rc, calls := run(); rc != 0 || !strings.HasSuffix(calls, endWith("-outcome blocked -cause test-failure")) {
		t.Errorf("a skipped test list after a blocked lint: exit %d, calls:\n%s", rc, calls)
	}
	writeFile(t, filepath.Join(wt, ".flow/project.md"), "## lint\n\n```bash\nexit 1\n```\n\n## test\n\n```bash\necho changed\n```\n")
	gitIn(t, wt, "add", "-A")
	gitIn(t, wt, "commit", "-q", "-m", "five")
	if rc, calls := run(); rc != 0 || !strings.HasSuffix(calls, endWith("-outcome blocked -cause test-failure")) {
		t.Errorf("a completed test list after a blocked lint: exit %d, calls:\n%s", rc, calls)
	}
	list = "lint"
	writeFile(t, filepath.Join(wt, ".flow/project.md"), "## lint\n\n```bash\ntrue\n```\n\n## test\n\n```bash\ntrue\n```\n")
	gitIn(t, wt, "add", "-A")
	gitIn(t, wt, "commit", "-q", "-m", "six")
	run()
	list = "test"
	if rc, calls := run(); rc != 0 || !strings.HasSuffix(calls, endWith("-outcome completed")) {
		t.Errorf("a test list after a lint list re-run green: exit %d, calls:\n%s", rc, calls)
	}
}

// TestRunVerifyListKeepsResultFiles pins kan-1017's self-review: a failing
// command's JUnit result files are copied out of the tree into a per-run
// directory the next run does not overwrite.
func TestRunVerifyListKeepsResultFiles(t *testing.T) {
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(fixtureGit, "-C", wt, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	writeFile(t, filepath.Join(wt, ".flow/project.md"), "## test\n\n```bash\nmkdir -p build/test-results && echo \"<x n='$$'/>\" > build/test-results/TEST-a.xml; exit 1\n```\n")
	repo, _ := filepath.Abs("../../..")
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "FLOW_GUARD_SCRIPT_DIR="+filepath.Join(repo, "scripts"))
	getenv := func(k string) string {
		for i := len(environ) - 1; i >= 0; i-- {
			if v, ok := strings.CutPrefix(environ[i], k+"="); ok {
				return v
			}
		}
		return ""
	}
	var dirs []string
	for range 2 {
		var o, e bytes.Buffer
		if rc := runVerifyList([]string{wt, "test"}, Env{Getenv: getenv, Environ: func() []string { return environ }, Dir: wt}, &o, &e); rc != 1 {
			t.Fatalf("exit %d: %s %s", rc, o.String(), e.String())
		}
		_, dir, ok := strings.Cut(o.String(), "  result files: ")
		if !ok {
			t.Fatalf("no result files line:\n%s", o.String())
		}
		dirs = append(dirs, strings.TrimSpace(strings.SplitN(dir, "\n", 2)[0]))
	}
	for _, d := range dirs {
		if _, err := os.Stat(filepath.Join(d, "build/test-results/TEST-a.xml")); err != nil {
			t.Errorf("%s: %v", d, err)
		}
	}
	if dirs[0] == dirs[1] {
		t.Errorf("both runs kept into %s", dirs[0])
	}
}

// TestRunVerifyListSurvivesAClosedStdout pins decision f6-ignore-sigpipe
// (kan-1037): with stdout a pipe whose reader is gone (`| head`), the list is
// not killed by SIGPIPE and still records its row's end. The guard runs in a
// child copy of this test binary, since SIGPIPE ends its process.
func TestRunVerifyListSurvivesAClosedStdout(t *testing.T) {
	if wt := os.Getenv("RVL_SIGPIPE_WT"); wt != "" {
		os.Exit(runVerifyList([]string{wt, "test", "kan-1", "verify", "mf-k1", "opus", "low"},
			Env{Getenv: os.Getenv, Environ: os.Environ, Dir: wt}, os.Stdout, os.Stderr))
	}
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	writeExec(t, filepath.Join(dir, "flow"), "#!/bin/sh\necho \"flow $*\" >> "+log+"\n")
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(wt, ".flow/project.md"), "## test\n\n```bash\nseq 1 200000\ntrue\n```\n")
	for _, args := range [][]string{{"init", "-q"}, {"add", "-A"}, {"commit", "-q", "-m", "one"}} {
		if out, err := exec.Command(fixtureGit, append([]string{"-C", wt, "-c", "user.name=t", "-c", "user.email=t@t"}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	r.Close()
	cmd := exec.Command(os.Args[0], "-test.run=^TestRunVerifyListSurvivesAClosedStdout$")
	cmd.Env = append(os.Environ(), "RVL_SIGPIPE_WT="+wt, "FLOW_GUARD_CACHE_DIR="+guardCache(t),
		"FLOW_GUARD_SCRIPT_DIR="+filepath.Join(repo, "scripts"), "PATH="+dir+":"+os.Getenv("PATH"))
	var stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = w, &stderr
	err = cmd.Run()
	w.Close()
	var exit *exec.ExitError
	if errors.As(err, &exit) {
		if ws := exit.Sys().(syscall.WaitStatus); ws.Signaled() {
			t.Fatalf("run-verify-list died of %v; stderr %q", ws.Signal(), stderr.String())
		}
	}
	calls, _ := os.ReadFile(log)
	if want := "flow record dispatch end -C " + wt + " -change kan-1 -key verify -session-token mf-k1 -outcome completed\n"; !strings.HasSuffix(string(calls), want) {
		t.Errorf("err %v, calls:\n%s\nwant the row's end %q; stderr %q", err, calls, want, stderr.String())
	}
}

// TestRunVerifyListKeepsSIGPIPEForCommands pins decision f6-ignore-sigpipe's
// Notify form: a list command's pipeline producer still dies of SIGPIPE (141),
// which an inherited signal.Ignore would turn into an EPIPE write error.
func TestRunVerifyListKeepsSIGPIPEForCommands(t *testing.T) {
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("FLOW_GUARD_CACHE_DIR", guardCache(t))
	t.Setenv("FLOW_GUARD_SCRIPT_DIR", filepath.Join(repo, "scripts"))
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(wt, ".flow/project.md"), "## test\n\n```bash\nbash -c 'yes | head -1 >/dev/null; test ${PIPESTATUS[0]} = 141'\n```\n")
	var out, errb bytes.Buffer
	if rc := runVerifyList([]string{wt, "test"}, Env{Getenv: os.Getenv, Environ: os.Environ, Dir: wt}, &out, &errb); rc != 0 {
		t.Fatalf("rc %d, want 0: the producer did not die of SIGPIPE\nstdout:\n%s\nstderr:\n%s", rc, out.String(), errb.String())
	}
}

// TestRunVerifyListRunsTheVisualVerify pins verify-list-runs-visual: a `test`
// list ends with the `## visual verification` `verify` row when the branch's
// diff from its base touches `ui paths`, unless decision.json records the
// visual verify skipped; a `lint` list never runs it.
func TestRunVerifyListRunsTheVisualVerify(t *testing.T) {
	t.Parallel()
	repo, _ := filepath.Abs("../../..")
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "FLOW_GUARD_SCRIPT_DIR="+filepath.Join(repo, "scripts"))
	getenv := func(k string) string {
		for i := len(environ) - 1; i >= 0; i-- {
			if v, ok := strings.CutPrefix(environ[i], k+"="); ok {
				return v
			}
		}
		return ""
	}
	git := func(t *testing.T, dir string, args ...string) {
		t.Helper()
		if out, err := exec.Command(fixtureGit, append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@t"}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	project := func(verify, checkout, specs string) string {
		row, specsRow := "", ""
		if checkout != "" {
			row = "| `regression checkout` | `" + checkout + "` |\n"
		}
		if specs != "" {
			specsRow = "| `specs` | `" + specs + "` |\n"
		}
		return "## lint\n\n```bash\necho LIST-RAN\n```\n\n## test\n\n```bash\necho LIST-RAN\n```\n\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `web/src/**` |\n| `screenshots` | `web/tests/visual` |\n" + row + "\n| Command | Runs |\n|---------|------|\n| `verify` | `" + verify + "` |\n| `capture` | `true <spec>` |\n" + specsRow
	}
	for _, tc := range []struct {
		name, verify, file, decision, list string
		specs, wantOut                     string
		noOrigin                           bool
		want                               int
		ran, checkout                      bool
	}{
		{"a ui change runs it", "echo VISUAL-RAN <specs>", "web/src/a.tsx", "", "test", "", "", false, 0, true, false},
		{"a failing visual verify fails the list", "exit 1", "web/src/a.tsx", "", "test", "", "", false, 1, false, false},
		{"a skipped visual verify", "echo VISUAL-RAN <specs>", "web/src/a.tsx", `{"visual":{"verify":"skipped","reason":"r"}}`, "test", "", "", false, 0, false, false},
		{"a change outside ui paths", "echo VISUAL-RAN <specs>", "docs/a.md", "", "test", "", "", false, 0, false, false},
		{"an unresolved base runs it", "echo VISUAL-RAN <specs>", "web/src/a.tsx", "", "test", "", "", true, 0, true, false},
		{"a lint list never runs it", "echo VISUAL-RAN <specs>", "web/src/a.tsx", "", "lint", "", "", false, 0, false, false},
		{"a regression checkout runs it from the change's worktree of that repo", "echo VISUAL-RAN in $PWD <specs>", "web/src/a.tsx", "", "test", "", "", false, 0, true, true},
		// verify-list-runs-targeted: a `specs` row scopes the run, and never
		// falls back to the whole suite.
		{"specs scopes it to the mapped specs", "echo VISUAL-RAN [<specs>]", "web/src/a.tsx", "", "test", "test -d <frontend-root> && git rev-parse -q --verify <merge-base> >/dev/null && { echo a.spec.ts; echo b.spec.ts; }", "VISUAL-RAN [a.spec.ts b.spec.ts]", false, 0, true, false},
		{"specs mapping nothing runs no spec", "echo VISUAL-RAN [<specs>]", "web/src/a.tsx", "", "test", "true <frontend-root> <merge-base>", "mapped no spec", false, 0, false, false},
		{"a failing specs fails the list", "echo VISUAL-RAN [<specs>]", "web/src/a.tsx", "", "test", "exit 3", "exit 3", false, 1, false, false},
		{"specs with an unresolved base fails the list", "echo VISUAL-RAN [<specs>]", "web/src/a.tsx", "", "test", "echo a.spec.ts <merge-base>", "no merge base", true, 1, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			base, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			seed, wt := filepath.Join(base, "seed"), filepath.Join(base, "wt")
			app, appWT := filepath.Join(base, "app"), filepath.Join(base, "app-wt")
			checkout := ""
			if tc.checkout {
				checkout = app
				writeFile(t, filepath.Join(app, "a"), "x\n")
				git(t, base, "init", "-q", "-b", "main", app)
				git(t, app, "add", "-A")
				git(t, app, "commit", "-q", "-m", "base")
				git(t, app, "worktree", "add", "-q", "-b", "work", appWT)
			}
			writeFile(t, filepath.Join(seed, ".flow/project.md"), project(tc.verify, checkout, tc.specs))
			writeFile(t, filepath.Join(seed, ".gitignore"), ".superpowers/\n")
			git(t, base, "init", "-q", "-b", "main", seed)
			git(t, seed, "add", "-A")
			git(t, seed, "commit", "-q", "-m", "base")
			if tc.noOrigin {
				git(t, base, "clone", "-q", seed, wt)
				git(t, wt, "remote", "remove", "origin")
			} else {
				git(t, base, "clone", "-q", "--bare", seed, filepath.Join(base, "origin.git"))
				git(t, base, "clone", "-q", filepath.Join(base, "origin.git"), wt)
			}
			git(t, wt, "checkout", "-q", "-b", "work")
			writeFile(t, filepath.Join(wt, tc.file), "x\n")
			git(t, wt, "add", "-A")
			git(t, wt, "commit", "-q", "-m", "change")
			if tc.decision != "" {
				writeFile(t, filepath.Join(wt, ".superpowers/sdd/decision.json"), tc.decision)
			}
			var o, e bytes.Buffer
			rc := runVerifyList([]string{wt, tc.list}, Env{Getenv: getenv, Environ: func() []string { return environ }, Dir: wt}, &o, &e)
			out := o.String() + e.String()
			if rc != tc.want {
				t.Fatalf("exit %d, want %d\n%s", rc, tc.want, out)
			}
			if strings.Contains(out, "VISUAL-RAN") != tc.ran {
				t.Fatalf("VISUAL-RAN printed = %v, want %v\n%s", !tc.ran, tc.ran, out)
			}
			if !strings.Contains(out, tc.wantOut) {
				t.Fatalf("output lacks %q\n%s", tc.wantOut, out)
			}
			if tc.checkout && !strings.Contains(out, "VISUAL-RAN in "+appWT+"\n") {
				t.Fatalf("the visual verify did not run from %s\n%s", appWT, out)
			}
			if tc.ran && strings.Index(out, "VISUAL-RAN") < strings.Index(out, "LIST-RAN") {
				t.Fatalf("the visual verify ran before the list\n%s", out)
			}
		})
	}
}

// TestRunVerifyListCapsMemory pins kan-1044's cap: every listed command runs
// under run-capped.sh, so one allocating past `## resource cap` (64 MB here,
// the child ~128 MB) is killed and the list fails naming CAPPED:.
func TestRunVerifyListCapsMemory(t *testing.T) {
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(fixtureGit, "-C", wt, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}
	writeFile(t, filepath.Join(wt, ".flow/project.md"), "## resource cap\n\n64\n\n## test\n\n```bash\npython3 -c 'import os, time; b = bytearray(os.urandom(128 << 20)); [(b.__setitem__(slice(None, None, 4096), os.urandom(len(b) // 4096)), time.sleep(0.2)) for _ in range(50)]'\n```\n")
	repo, _ := filepath.Abs("../../..")
	// An outer run-capped.sh's RUN_CAPPED_PGID would make the inner one run
	// in place under the outer cap, never this fixture's 64 MB.
	environ := slices.DeleteFunc(os.Environ(), func(kv string) bool { return strings.HasPrefix(kv, "RUN_CAPPED_PGID=") })
	environ = append(environ, "FLOW_GUARD_CACHE_DIR="+guardCache(t), "FLOW_GUARD_SCRIPT_DIR="+filepath.Join(repo, "scripts"))
	getenv := func(k string) string {
		for i := len(environ) - 1; i >= 0; i-- {
			if v, ok := strings.CutPrefix(environ[i], k+"="); ok {
				return v
			}
		}
		return ""
	}
	var o, e bytes.Buffer
	if rc := runVerifyList([]string{wt, "test"}, Env{Getenv: getenv, Environ: func() []string { return environ }, Dir: wt}, &o, &e); rc != 1 {
		t.Fatalf("exit %d: %s %s", rc, o.String(), e.String())
	}
	if !strings.Contains(o.String()+e.String(), "CAPPED:") {
		t.Errorf("no CAPPED: line:\n%s\n%s", o.String(), e.String())
	}
}

// TestRunVerifyListRefusesAnOffContractKey pins the kan-1067 self-review fix,
// 1-C: a recording key that does not fold to verify or verify-<repository
// name> is refused, exit 2, before the begin and before any command.
func TestRunVerifyListRefusesAnOffContractKey(t *testing.T) {
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	writeExec(t, filepath.Join(dir, "flow"), "#!/bin/sh\necho \"flow $*\" >> "+log+"\n")
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	wt := filepath.Join(tmp, "repo-x")
	gitRun(t, tmp, "init", "-q", wt)
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "FLOW_GUARD_SCRIPT_DIR="+filepath.Join(repo, "scripts"), "PATH="+dir+":"+os.Getenv("PATH"))
	getenv := func(k string) string {
		for i := len(environ) - 1; i >= 0; i-- {
			if v, ok := strings.CutPrefix(environ[i], k+"="); ok {
				return v
			}
		}
		return ""
	}
	run := func(on, key string) (int, string, string) {
		os.Remove(log)
		var o, e bytes.Buffer
		rc := runVerifyList([]string{on, "lint", "kan-1", key, "mf-k1", "opus", "low"}, Env{Getenv: getenv, Environ: func() []string { return environ }, Dir: on}, &o, &e)
		calls, _ := os.ReadFile(log)
		return rc, string(calls), e.String()
	}
	project := func(on string) {
		writeFile(t, filepath.Join(on, ".flow/project.md"), "## lint\n\n```bash\ntrue\n```\n")
	}
	project(wt)
	base := filepath.Base(wt)
	admitted := func(on, key string) {
		rc, calls, errs := run(on, key)
		if rc != 0 || !strings.HasPrefix(calls, "flow record dispatch begin ") || !strings.Contains(strings.SplitN(calls, "\n", 2)[0], "-key "+key+" ") {
			t.Errorf("key %s on %s: exit %d, calls %q, stderr %q", key, on, rc, calls, errs)
		}
	}
	for _, key := range []string{"verify", "verify-2", "verify-fix-1", "verify-" + base, "verify-fix-2-" + base} {
		admitted(wt, key)
	}
	for _, key := range []string{"verify-fix-1b", "verify-2b", "verify-other"} {
		rc, calls, errs := run(wt, key)
		if rc != 2 || calls != "" || !strings.Contains(errs, key) || !strings.Contains(errs, "verify-and-handoff.md") {
			t.Errorf("key %s: exit %d, calls %q, stderr %q", key, rc, calls, errs)
		}
	}
	gitRun(t, wt, "commit", "-q", "--allow-empty", "-m", "init")
	other := filepath.Join(tmp, "other")
	gitRun(t, wt, "worktree", "add", "-q", other)
	project(other)
	admitted(other, "verify-"+base)
}
