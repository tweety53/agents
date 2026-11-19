package guard

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// TestRunVerifyList carries the cases of the retired
// scripts/test-run-verify-list.sh, in its order and over its fixtures. Every
// step runs twice, on twin roots: through the Go guard and through the bash
// at 8c54d6d3, whose project-get.sh and check-spec-reach.sh siblings are the
// live shims. Each step asserts the harness's expectations on the Go output
// and byte-for-byte stdout/stderr/exit parity with the bash, each twin's root
// written as <ROOT>.
func TestRunVerifyList(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	bashDir := t.TempDir()
	writeExec(t, filepath.Join(bashDir, "run-verify-list.sh"),
		string(historicalFile(t, repo, "8c54d6d3", "scripts/run-verify-list.sh")))
	// The bash's two siblings, now run through the bootstrap.
	for _, l := range []string{"project-get", "check-spec-reach"} {
		writeExec(t, filepath.Join(bashDir, l+".sh"), "#!/bin/sh\nexec '"+filepath.Join(repo, "scripts/flow-guard")+"' "+l+" \"$@\"\n")
	}
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t))
	goEnviron := environ
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
		// The recording arguments (kan-1033) and the visual list (kan-1052
		// self-review, F2) postdate the bash: its usage line lacks them.
		g := res{grc, strings.NewReplacer(" [<change> <key> <session-token> <model> <effort> [<canonical-worktree>]]", "", "lint|test|visual", "lint|test").Replace(norm(gr, o.String(), e.String()))}

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
		// The port spells its usage and the spec-reach row as flow-guard calls.
		respell := strings.NewReplacer("usage: run-verify-list.sh", "usage: flow-guard run-verify-list", "run-verify-list.sh: ", "run-verify-list: ", "`check-spec-reach.sh ", "`flow-guard check-spec-reach ")
		if b := (res{brc, respell.Replace(norm(br, bo.String(), be.String()))}); g != b {
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
		has(t, r, "- `flow-guard check-spec-reach <ROOT>` — exit 0\n  Spec reach: not configured")
		lacks(t, r, "full output:")
	})

	t.Run("a passing list", func(t *testing.T) {
		r := run(t, "<ROOT>", "test")
		rc(t, r, 0)
		has(t, r, "- `printf 'ok\\n'` — exit 0\n  ok")
		has(t, r, "  nonl\n- `echo next` — exit 0")
		lacks(t, r, "check-spec-reach")
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
			fixtureRepo(t, g, repoShape{})
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
			fixtureRepo(t, f, repoShape{})
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
// smoke: core — the verify row a list records for itself
func TestRunVerifyListRecordsTheVerifyRow(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	writeExec(t, filepath.Join(dir, "flow"), "#!/bin/sh\n[ \"$1\" = state ] && exit 0\n[ \"$2\" = dispatches ] && exit 0\necho \"flow $*\" >> "+log+"\n")
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	path := dir + ":" + os.Getenv("PATH")
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "PATH="+path)
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
		// THE RUN'S COST row is TestRunVerifyListRecordsItsCost's.
		return rc, regexp.MustCompile(`(?m)^flow record run .*\n`).ReplaceAllString(string(calls), "")
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
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	writeExec(t, filepath.Join(dir, "flow"), "#!/usr/bin/env bash\necho \"$*\" >> "+log+"\nif [ \"$1 $2\" = \"suite record\" ]; then\n  while [ \"$1\" != -- ]; do shift; done\n  shift\n  exec \"$@\"\nfi\n")
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(wt, ".flow/project.md"), "## test\n\n```bash\n# suite: demo-suite\necho one\necho two\n```\n")
	run := func(path string) (int, string) {
		environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "PATH="+path)
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
	if want := "suite record -C " + wt + " -suite demo-suite -- bash -c echo one\n"; string(calls) != want {
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
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "PATH="+dir+":"+os.Getenv("PATH"))
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
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	writeExec(t, filepath.Join(dir, "flow"), "#!/bin/sh\n[ \"$1\" = state ] && exit 0\n[ \"$2\" = dispatches ] && exit 0\necho \"flow $*\" >> "+log+"\n")
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
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "PATH="+dir+":"+os.Getenv("PATH"))
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
		return len(lines) == 3 // begin, THE RUN'S COST, end
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
	smoke.Heavy(t)
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	writeExec(t, filepath.Join(dir, "flow"), "#!/bin/sh\n[ \"$1\" = state ] && exit 0\n[ \"$2\" = dispatches ] && exit 0\necho \"flow $*\" >> "+log+"\n")
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "PATH="+dir+":"+os.Getenv("PATH"))
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
	fixtureRepo(t, wt, repoShape{})

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
	// Both groups run at once under one key, so either may end first: a
	// blocked test group is never masked by the lint group ending green later
	// (kan-1058 self-review, R2).
	writeFile(t, filepath.Join(wt, ".flow/project.md"), "## lint\n\n```bash\ntrue\n```\n\n## test\n\n```bash\nexit 1\n```\n")
	gitIn(t, wt, "add", "-A")
	gitIn(t, wt, "commit", "-q", "-m", "seven")
	if rc, calls := run(); rc != 1 || !strings.HasSuffix(calls, endWith("-outcome blocked -cause test-failure")) {
		t.Errorf("a failing test list: exit %d, calls:\n%s", rc, calls)
	}
	list = "lint"
	if rc, calls := run(); rc != 0 || !strings.HasSuffix(calls, endWith("-outcome blocked -cause test-failure")) {
		t.Errorf("a green lint list after a blocked test: exit %d, calls:\n%s", rc, calls)
	}
}

// TestRunVerifyListLeavesTheRowOpenWhileASiblingRuns pins kan-1086
// self-review, F3: lint and test share one row, and a group that ends green
// while its sibling still runs — or was killed before its end — records no
// end; the last group to finish ends the row.
func TestRunVerifyListLeavesTheRowOpenWhileASiblingRuns(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	gate := filepath.Join(dir, "go")
	writeExec(t, filepath.Join(dir, "flow"), "#!/bin/sh\n[ \"$1\" = state ] && exit 0\n[ \"$2\" = dispatches ] && exit 0\necho \"flow $*\" >> "+log+"\n")
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "PATH="+dir+":"+os.Getenv("PATH"))
	getenv := func(k string) string {
		for i := len(environ) - 1; i >= 0; i-- {
			if v, ok := strings.CutPrefix(environ[i], k+"="); ok {
				return v
			}
		}
		return ""
	}
	git := func(args ...string) string {
		t.Helper()
		out, err := exec.Command(fixtureGit, append([]string{"-C", wt, "-c", "user.name=t", "-c", "user.email=t@t"}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run := func(list string) int {
		var o, e bytes.Buffer
		return runVerifyList([]string{wt, list, "kan-1", "verify", "mf-k1", "opus", "low"},
			Env{Getenv: getenv, Environ: func() []string { return environ }, Dir: wt}, &o, &e)
	}
	calls := func() string { b, _ := os.ReadFile(log); return string(b) }
	fixtureRepo(t, wt, repoShape{})
	writeFile(t, filepath.Join(wt, ".flow/project.md"),
		"## lint\n\n```bash\ntrue\n```\n\n## test\n\n```bash\nwhile [ ! -f "+gate+" ]; do sleep 0.05; done\n```\n")
	git("add", "-A")
	git("commit", "-q", "-m", "one")

	done := make(chan int)
	go func() { done <- run("test") }()
	for deadline := time.Now().Add(30 * time.Second); !strings.Contains(calls(), "dispatch begin"); time.Sleep(20 * time.Millisecond) {
		if time.Now().After(deadline) {
			t.Fatal("the test group never began")
		}
	}
	before := len(calls())
	if rc := run("lint"); rc != 0 {
		t.Fatalf("lint exit %d", rc)
	}
	lint := calls()[before:]
	if !strings.Contains(lint, "dispatch begin") || strings.Contains(lint, "dispatch end") {
		t.Errorf("a lint group ending green while test runs must begin and not end the row, calls:\n%s", lint)
	}
	writeFile(t, gate, "")
	if rc := <-done; rc != 0 {
		t.Fatalf("test exit %d", rc)
	}
	all := strings.Split(strings.TrimRight(calls(), "\n"), "\n")
	if last := all[len(all)-1]; !strings.Contains(last, "dispatch end") || !strings.HasSuffix(last, "-outcome completed") {
		t.Errorf("the last group to finish must end the row completed, last call %q", last)
	}

	// A test group killed before its end leaves its running mark: a later
	// green lint group leaves the row open for the parent.
	writeFile(t, git("rev-parse", "--path-format=absolute", "--git-path", "flow-verify-running-verify-mf-k1-test"), "")
	os.Remove(log)
	if rc := run("lint"); rc != 0 {
		t.Fatalf("lint exit %d", rc)
	}
	if c := calls(); strings.Contains(c, "dispatch end") {
		t.Errorf("a green lint group beside a killed test group's mark must not end the row, calls:\n%s", c)
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
	fixtureRepo(t, wt, repoShape{})
	writeFile(t, filepath.Join(wt, ".flow/project.md"), "## test\n\n```bash\nmkdir -p build/test-results && echo \"<x n='$$'/>\" > build/test-results/TEST-a.xml; exit 1\n```\n")
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t))
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
	smoke.Heavy(t)
	if wt := os.Getenv("RVL_SIGPIPE_WT"); wt != "" {
		os.Exit(runVerifyList([]string{wt, "test", "kan-1", "verify", "mf-k1", "opus", "low"},
			Env{Getenv: os.Getenv, Environ: os.Environ, Dir: wt}, os.Stdout, os.Stderr))
	}
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	writeExec(t, filepath.Join(dir, "flow"), "#!/bin/sh\n[ \"$1\" = state ] && exit 0\n[ \"$2\" = dispatches ] && exit 0\necho \"flow $*\" >> "+log+"\n")
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(wt, ".flow/project.md"), "## test\n\n```bash\nseq 1 200000\ntrue\n```\n")
	fixtureRepo(t, wt, repoShape{})
	for _, args := range [][]string{{"add", "-A"}, {"commit", "-q", "-m", "one"}} {
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
		"PATH="+dir+":"+os.Getenv("PATH"))
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
	t.Setenv("FLOW_GUARD_CACHE_DIR", guardCache(t))
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

// rvlVisualRun runs runVerifyList's list on a fresh seed→origin→wt fixture:
// a project whose `## lint` and `## test` lines `echo LIST-RAN`, with `ui
// paths` `web/src/**` and the given `verify` and `specs` rows, and a `work`
// branch committing file. decision, when set, is written to decision.json;
// noOrigin drops the remote, so the base does not resolve; checkout declares
// a `regression checkout` whose `work` worktree is appWT, unless noAppWT
// leaves that checkout with no worktree on the branch. It returns the exit
// code and stdout+stderr.
func rvlVisualRun(t *testing.T, list, verify, file, decision, specs string, noOrigin, checkout, noAppWT bool) (rc int, out, appWT string) {
	t.Helper()
	wt, appWT := rvlVisualSetup(t, verify, file, decision, specs, noOrigin, checkout, noAppWT)
	rc, out = rvlRunIn(t, wt, "", wt, list)
	return rc, out, appWT
}

// rvlRunIn runs run-verify-list with args from dir, path prepended to PATH
// when set, and returns its exit code and stdout+stderr. PATH never holds
// the real `flow`, whose `suite record` would write to the dev store: a test
// wanting one passes its stub's directory as path.
func rvlRunIn(t *testing.T, dir, path string, args ...string) (int, string) {
	t.Helper()
	return rvlRunEnv(t, dir, path, nil, args...)
}

// rvlOperator is the environment entry only the operator sets (THE
// OPERATOR'S SWITCH).
var rvlOperator = []string{rvlOperatorFull + "=1"}

// rvlRunEnv is rvlRunIn with extra appended to the environment, after the
// entries that make every run a default one no operator asked to widen.
func rvlRunEnv(t *testing.T, dir, path string, extra []string, args ...string) (int, string) {
	t.Helper()
	return rvlRunRows(t, dir, path, extra, nil, args...)
}

// rvlRunRows is rvlRunEnv with dispatches standing for `flow record
// dispatches` when set.
func rvlRunRows(t *testing.T, dir, path string, extra []string, dispatches func(string) ([]byte, error), args ...string) (int, string) {
	t.Helper()
	var bare []string
	for _, d := range filepath.SplitList(os.Getenv("PATH")) {
		if _, err := os.Stat(filepath.Join(d, "flow")); err != nil {
			bare = append(bare, d)
		}
	}
	if path != "" {
		bare = append([]string{path}, bare...)
	}
	// FLOW_VERIFY_FULL= and FLOW_SMOKE_KEEP= last: every run through here is
	// a default run, whatever the process carries.
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "PATH="+strings.Join(bare, ":"), "FLOW_VERIFY_FULL=", "FLOW_SMOKE_KEEP=", rvlOperatorFull+"=")
	environ = append(environ, extra...)
	getenv := func(k string) string {
		for i := len(environ) - 1; i >= 0; i-- {
			if v, ok := strings.CutPrefix(environ[i], k+"="); ok {
				return v
			}
		}
		return ""
	}
	var o, e bytes.Buffer
	rc := runVerifyList(args, Env{Getenv: getenv, Environ: func() []string { return environ }, Dir: dir, Dispatches: dispatches}, &o, &e)
	return rc, o.String() + e.String()
}

// rvlGit runs git in dir with a fixed identity, failing the test on error.
func rvlGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command(fixtureGit, append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@t"}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// rvlVisualSetup builds rvlVisualRun's fixture and returns the worktree on
// branch `work` and the regression checkout's worktree path.
func rvlVisualSetup(t *testing.T, verify, file, decision, specs string, noOrigin, checkout, noAppWT bool) (wt, appWT string) {
	t.Helper()
	git := func(t *testing.T, dir string, args ...string) { t.Helper(); rvlGit(t, dir, args...) }
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
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	seed := filepath.Join(base, "seed")
	wt = filepath.Join(base, "wt")
	app := filepath.Join(base, "app")
	appWT = filepath.Join(base, "app-wt")
	checkoutPath := ""
	if checkout {
		checkoutPath = app
		writeFile(t, filepath.Join(app, "a"), "x\n")
		fixtureRepo(t, app, repoShape{Branch: "main"})
		git(t, app, "add", "-A")
		git(t, app, "commit", "-q", "-m", "base")
		if !noAppWT {
			git(t, app, "worktree", "add", "-q", "-b", "work", appWT)
		}
	}
	writeFile(t, filepath.Join(seed, ".flow/project.md"), project(verify, checkoutPath, specs))
	writeFile(t, filepath.Join(seed, ".gitignore"), ".superpowers/\n")
	fixtureRepo(t, seed, repoShape{Branch: "main"})
	git(t, seed, "add", "-A")
	git(t, seed, "commit", "-q", "-m", "base")
	if noOrigin {
		git(t, base, "clone", "-q", seed, wt)
		git(t, wt, "remote", "remove", "origin")
	} else {
		git(t, base, "clone", "-q", "--bare", seed, filepath.Join(base, "origin.git"))
		git(t, base, "clone", "-q", filepath.Join(base, "origin.git"), wt)
	}
	git(t, wt, "checkout", "-q", "-b", "work")
	writeFile(t, filepath.Join(wt, file), "x\n")
	git(t, wt, "add", "-A")
	git(t, wt, "commit", "-q", "-m", "change")
	if decision != "" {
		writeFile(t, filepath.Join(wt, ".superpowers/sdd/decision.json"), decision)
	}
	return wt, appWT
}

// TestRunVerifyListRunsTheVisualVerify pins verify-list-runs-visual: a `test`
// list ends with the `## visual verification` `verify` row when the branch's
// diff from its base touches `ui paths`, unless decision.json records the
// visual verify skipped; a `lint` list never runs it.
func TestRunVerifyListRunsTheVisualVerify(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, tc := range []struct {
		name, verify, file, decision, list string
		specs, wantOut                     string
		noOrigin                           bool
		want                               int
		ran, checkout                      bool
	}{
		{"a ui change runs it", "echo VISUAL-RAN <specs>", "web/src/a.tsx", "", "test", "", "", false, 0, true, false},
		{"a failing visual verify fails the list", "false <specs>", "web/src/a.tsx", "", "test", "", "", false, 1, false, false},
		{"a skipped visual verify", "echo VISUAL-RAN <specs>", "web/src/a.tsx", `{"visual":{"verify":"skipped","reason":"r"}}`, "test", "", "", false, 0, false, false},
		{"a change outside ui paths", "echo VISUAL-RAN <specs>", "docs/a.md", "", "test", "", "", false, 0, false, false},
		{"an unresolved base runs it", "echo VISUAL-RAN <specs>", "web/src/a.tsx", "", "test", "", "", true, 0, true, false},
		{"a lint list never runs it", "echo VISUAL-RAN <specs>", "web/src/a.tsx", "", "lint", "", "", false, 0, false, false},
		{"a regression checkout runs it from the change's worktree of that repo", "echo VISUAL-RAN in $PWD <specs>", "web/src/a.tsx", "", "test", "", "", false, 0, true, true},
		// verify-list-runs-targeted: a `specs` row scopes the run, and never
		// falls back to the whole suite.
		{"specs scopes it to the mapped specs", "echo VISUAL-RAN [<specs>]", "web/src/a.tsx", "", "test", "test -d <frontend-root> && git rev-parse -q --verify <merge-base> >/dev/null && { echo a.spec.ts; echo b.spec.ts; }", "VISUAL-RAN [a.spec.ts b.spec.ts]", false, 0, true, false},
		{"specs notes on stderr never reach the list", "echo VISUAL-RAN [<specs>]", "web/src/a.tsx", "", "test", "echo 'note: shared rendering code -- no spec selected' >&2; echo a.spec.ts", "VISUAL-RAN [a.spec.ts]", false, 0, true, false},
		// THE SMOKE FALLBACK: the smoke set fills `<specs>` wherever no
		// selection can, and the row never runs with it empty.
		{"no specs row runs the smoke set", "echo VISUAL-RAN [<specs>]", "web/src/a.tsx", "", "test", "", "VISUAL-RAN [--grep @smoke]", false, 0, true, false},
		{"no specs row and an unresolved base runs the smoke set", "echo VISUAL-RAN [<specs>]", "web/src/a.tsx", "", "test", "", "VISUAL-RAN [--grep @smoke]", true, 0, true, false},
		{"specs mapping nothing runs the smoke set", "echo VISUAL-RAN [<specs>]", "web/src/a.tsx", "", "test", "true <frontend-root> <merge-base>", "VISUAL-RAN [--grep @smoke]", false, 0, true, false},
		{"a failing specs fails the list and runs the smoke set", "echo VISUAL-RAN [<specs>]", "web/src/a.tsx", "", "test", "exit 3", "VISUAL-RAN [--grep @smoke]", false, 1, true, false},
		{"specs with an unresolved base fails the list and runs the smoke set", "echo VISUAL-RAN [<specs>]", "web/src/a.tsx", "", "test", "echo a.spec.ts <merge-base>", "VISUAL-RAN [--grep @smoke]", true, 1, true, false},
		{"a verify cell with no <specs> is not run", "echo WHOLE", "web/src/a.tsx", "", "test", "", "no <specs> placeholder", false, 1, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rc, out, appWT := rvlVisualRun(t, tc.list, tc.verify, tc.file, tc.decision, tc.specs, tc.noOrigin, tc.checkout, false)
			if rc != tc.want {
				t.Fatalf("exit %d, want %d\n%s", rc, tc.want, out)
			}
			if strings.Contains(out, "VISUAL-RAN") != tc.ran {
				t.Fatalf("VISUAL-RAN printed = %v, want %v\n%s", !tc.ran, tc.ran, out)
			}
			if !strings.Contains(out, tc.wantOut) {
				t.Fatalf("output lacks %q\n%s", tc.wantOut, out)
			}
			if strings.Contains(out, "VISUAL-RAN []") || strings.Contains(out, "VISUAL-RAN \n") {
				t.Fatalf("the visual verify ran with <specs> empty\n%s", out)
			}
			if smoke := strings.Contains(out, "--grep @smoke"); smoke != strings.Contains(out, "runs the smoke set alone") {
				t.Fatalf("the smoke set ran = %v, and stderr does not say the same\n%s", smoke, out)
			}
			if tc.checkout && !strings.Contains(out, "VISUAL-RAN in "+appWT+" --grep @smoke\n") {
				t.Fatalf("the visual verify did not run from %s\n%s", appWT, out)
			}
			if tc.ran && strings.Index(out, "VISUAL-RAN") < strings.Index(out, "LIST-RAN") {
				t.Fatalf("the visual verify ran before the list\n%s", out)
			}
		})
	}
}

// TestRunVerifyListVisualList pins the kan-1052 self-review fix (F2): a
// `visual` list runs the `verify` row alone, under the conditions a `test`
// list's tail applies, never the `## lint` or `## test` commands, and prints a
// not-run line and exits 0 when those conditions skip it.
func TestRunVerifyListVisualList(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, tc := range []struct {
		name, verify, file, decision string
		want                         int
		ran, notRun, noAppWT         bool
	}{
		{"a ui change runs the verify row alone", "echo VISUAL-RAN <specs>", "web/src/a.tsx", "", 0, true, false, false},
		{"a failing verify row fails the list", "false <specs>", "web/src/a.tsx", "", 1, false, false, false},
		{"a change outside ui paths runs nothing", "echo VISUAL-RAN <specs>", "docs/a.md", "", 0, false, true, false},
		{"a skipped visual verify runs nothing", "echo VISUAL-RAN <specs>", "web/src/a.tsx", `{"visual":{"verify":"skipped","reason":"r"}}`, 0, false, true, false},
		// kan-1052 self-review R1: /flow-fast and /flow-interactive create no
		// regression-checkout worktree; the row must never run from the main
		// checkout, against a build that is not the change's.
		{"a regression checkout with no worktree on the branch refuses", "echo VISUAL-RAN <specs>", "web/src/a.tsx", "", 1, false, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rc, out, _ := rvlVisualRun(t, "visual", tc.verify, tc.file, tc.decision, "", false, tc.noAppWT, tc.noAppWT)
			if rc != tc.want {
				t.Fatalf("exit %d, want %d\n%s", rc, tc.want, out)
			}
			if strings.Contains(out, "VISUAL-RAN") != tc.ran {
				t.Fatalf("VISUAL-RAN printed = %v, want %v\n%s", !tc.ran, tc.ran, out)
			}
			if strings.Contains(out, "LIST-RAN") {
				t.Fatalf("a visual list ran a lint or test command\n%s", out)
			}
			if strings.Contains(out, "- visual verify — not run") != tc.notRun {
				t.Fatalf("not-run line printed = %v, want %v\n%s", !tc.notRun, tc.notRun, out)
			}
		})
	}
}

// TestRunVerifyListRefusesAnOffContractKey pins the kan-1067 self-review fix,
// 1-C: a recording key that does not fold to verify or verify-<repository
// name> is refused, exit 2, before the begin and before any command.
func TestRunVerifyListRefusesAnOffContractKey(t *testing.T) {
	smoke.Heavy(t)
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	writeExec(t, filepath.Join(dir, "flow"), "#!/bin/sh\n[ \"$1\" = state ] && exit 0\n[ \"$2\" = dispatches ] && exit 0\necho \"flow $*\" >> "+log+"\n")
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	wt := filepath.Join(tmp, "repo-x")
	fixtureRepo(t, wt, repoShape{})
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "PATH="+dir+":"+os.Getenv("PATH"))
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

// TestRunVerifyListRecordsHead pins verify-row-records-head: the recording
// begin carries `-diff-base` with the HEAD of the verified worktree (the first
// argument), never the canonical one.
func TestRunVerifyListRecordsHead(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	writeExec(t, filepath.Join(dir, "flow"), "#!/bin/sh\n[ \"$1\" = state ] && exit 0\n[ \"$2\" = dispatches ] && exit 0\necho \"flow $*\" >> "+log+"\n")
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	canon, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(wt, ".flow/project.md"), "## test\n\n```bash\ntrue\n```\n")
	fixtureRepo(t, wt, repoShape{})
	gitRun(t, wt, "add", "-A")
	gitRun(t, wt, "commit", "-qm", "a")
	fixtureRepo(t, canon, repoShape{Commit: "c"})
	head := func(d string) string {
		out, err := exec.Command("git", "-C", d, "rev-parse", "HEAD").Output()
		if err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(string(out))
	}
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "PATH="+dir+":"+os.Getenv("PATH"))
	getenv := func(k string) string {
		for i := len(environ) - 1; i >= 0; i-- {
			if v, ok := strings.CutPrefix(environ[i], k+"="); ok {
				return v
			}
		}
		return ""
	}
	var o, e bytes.Buffer
	rc := runVerifyList([]string{wt, "test", "kan-1", "verify", "mf-k1", "opus", "low", canon}, Env{Getenv: getenv, Environ: func() []string { return environ }, Dir: wt}, &o, &e)
	calls, _ := os.ReadFile(log)
	first := strings.SplitN(string(calls), "\n", 2)[0]
	if rc != 0 || !strings.HasPrefix(first, "flow record dispatch begin ") || !strings.HasSuffix(first, " -diff-base "+head(wt)) {
		t.Fatalf("exit %d, begin %q, want -diff-base %s\n%s", rc, first, head(wt), e.String())
	}
}

// TestProjectTestListRunsGoSuitesOnACoreShare pins the kan-1052 self-review
// fix (F5): the `## test` list's suites run at once, so each Go suite's
// command runs under a third of the cores — both `-race` suites taking every
// core starved the SPA suite until a 235 ms vitest test passed its 5 s
// timeout.
func TestProjectTestListRunsGoSuitesOnACoreShare(t *testing.T) {
	t.Parallel()
	const share = "GOMAXPROCS=$(( ($(getconf _NPROCESSORS_ONLN) + 2) / 3 ))"
	b, err := os.ReadFile("../../../.flow/project.md")
	if err != nil {
		t.Fatal(err)
	}
	_, sec, ok := strings.Cut(string(b), "\n## test\n")
	if !ok {
		t.Fatal(".flow/project.md has no `## test` section")
	}
	_, fence, ok := strings.Cut(sec, "```bash\n")
	if !ok {
		t.Fatal("`## test` has no ```bash fence")
	}
	fence, _, _ = strings.Cut(fence, "```")
	lines := strings.Split(fence, "\n")
	for _, suite := range []string{"guard-tests", "stats-go"} {
		i := slices.Index(lines, "# suite: "+suite)
		if i < 0 || i+1 >= len(lines) {
			t.Errorf("`## test` has no `# suite: %s` marker with a command beneath it", suite)
			continue
		}
		if !strings.Contains(lines[i+1], share) {
			t.Errorf("suite %s runs without the core share %q: %s", suite, share, lines[i+1])
		}
	}
}

// TestRunVerifyListStampsLintGreen pins the kan-1081 self-review fix (F1): a
// `lint` list that exits 0 on a tree clean before and after it writes HEAD to
// the worktree's `flow-lint-green`, the evidence a landed request end reads;
// a failing or dirty run writes nothing.
func TestRunVerifyListStampsLintGreen(t *testing.T) {
	head := func(t *testing.T, d string) string {
		out, err := exec.Command("git", "-C", d, "rev-parse", "HEAD").Output()
		if err != nil {
			t.Fatal(err)
		}
		return strings.TrimSpace(string(out))
	}
	setup := func(t *testing.T, cmd string) (string, string) {
		wt, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(wt, ".flow/project.md"), "## lint\n\n```bash\n"+cmd+"\n```\n")
		fixtureRepo(t, wt, repoShape{})
		gitRun(t, wt, "add", "-A")
		gitRun(t, wt, "commit", "-qm", "a")
		out, err := exec.Command("git", "-C", wt, "rev-parse", "--path-format=absolute", "--git-path", "flow-lint-green").Output()
		if err != nil {
			t.Fatal(err)
		}
		return wt, strings.TrimSpace(string(out))
	}
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t))
	getenv := func(k string) string {
		for i := len(environ) - 1; i >= 0; i-- {
			if v, ok := strings.CutPrefix(environ[i], k+"="); ok {
				return v
			}
		}
		return ""
	}
	lint := func(wt string) (int, string) {
		var o, e bytes.Buffer
		rc := runVerifyList([]string{wt, "lint"}, Env{Getenv: getenv, Environ: func() []string { return environ }, Dir: wt}, &o, &e)
		return rc, o.String() + e.String()
	}
	stamp := func(f string) (string, bool) {
		b, err := os.ReadFile(f)
		return string(b), err == nil
	}
	t.Run("green", func(t *testing.T) {
		wt, f := setup(t, "true")
		rc, out := lint(wt)
		if got, ok := stamp(f); rc != 0 || !ok || got != head(t, wt)+"\n" {
			t.Fatalf("exit %d, stamp %q (%v), want %s\n%s", rc, got, ok, head(t, wt), out)
		}
	})
	t.Run("failing", func(t *testing.T) {
		wt, f := setup(t, "false")
		rc, _ := lint(wt)
		if _, ok := stamp(f); rc != 1 || ok {
			t.Fatalf("exit %d, stamp written %v; want exit 1, no stamp", rc, ok)
		}
	})
	t.Run("dirty", func(t *testing.T) {
		wt, f := setup(t, "true")
		writeFile(t, filepath.Join(wt, "untracked"), "x")
		rc, out := lint(wt)
		if _, ok := stamp(f); rc != 0 || ok {
			t.Fatalf("exit %d, stamp written %v; want exit 0, no stamp\n%s", rc, ok, out)
		}
	})
	t.Run("stale", func(t *testing.T) {
		wt, f := setup(t, "true")
		old := head(t, wt)
		writeFile(t, f, old+"\n")
		gitRun(t, wt, "commit", "-q", "--allow-empty", "-m", "b")
		rc, out := lint(wt)
		if got, _ := stamp(f); rc != 0 || got != head(t, wt)+"\n" || got == old+"\n" {
			t.Fatalf("exit %d, stamp %q, want the newer %s\n%s", rc, got, head(t, wt), out)
		}
	})
}

// TestRunVerifyListSkipsACommandGreenAtHead: a `test` re-run at the same
// clean HEAD skips each command an earlier run passed while another failed,
// and a new commit runs it again (kan-1103 self-review, F10).
// smoke: core — the green-commands skip at an unchanged HEAD
func TestRunVerifyListSkipsACommandGreenAtHead(t *testing.T) {
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	a := `echo ran >> "$(git rev-parse --git-dir)/a-runs"; echo a`
	b := `test -f "$(git rev-parse --git-dir)/b-ok"`
	writeFile(t, filepath.Join(wt, ".flow/project.md"), "## test\n\n```bash\n"+a+"\n"+b+"\n```\n")
	fixtureRepo(t, wt, repoShape{})
	gitRun(t, wt, "add", "-A")
	gitRun(t, wt, "commit", "-qm", "a")
	gitDir := filepath.Join(wt, ".git")
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t))
	getenv := func(k string) string {
		for i := len(environ) - 1; i >= 0; i-- {
			if v, ok := strings.CutPrefix(environ[i], k+"="); ok {
				return v
			}
		}
		return ""
	}
	test := func() (int, string) {
		var o, e bytes.Buffer
		rc := runVerifyList([]string{wt, "test"}, Env{Getenv: getenv, Environ: func() []string { return environ }, Dir: wt}, &o, &e)
		return rc, o.String() + e.String()
	}
	runs := func() int {
		b, _ := os.ReadFile(filepath.Join(gitDir, "a-runs"))
		return strings.Count(string(b), "\n")
	}

	if rc, out := test(); rc != 1 || runs() != 1 {
		t.Fatalf("run 1: exit %d, a-runs %d; want 1, 1\n%s", rc, runs(), out)
	}
	writeFile(t, filepath.Join(gitDir, "b-ok"), "")
	head, err := exec.Command("git", "-C", wt, "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	rc, out := test()
	if want := "- `" + a + "` — skipped, passed at " + strings.TrimSpace(string(head)) + "\n"; rc != 0 || !strings.Contains(out, want) || runs() != 1 || !strings.Contains(out, "- `"+b+"` — exit 0\n") {
		t.Fatalf("run 2: exit %d, a-runs %d; want 0, 1, %q and b exit 0\n%s", rc, runs(), want, out)
	}
	writeFile(t, filepath.Join(wt, "more.txt"), "x")
	gitRun(t, wt, "add", "-A")
	gitRun(t, wt, "commit", "-qm", "b")
	if rc, out := test(); runs() != 2 {
		t.Fatalf("run 3: exit %d, a-runs %d; want 2\n%s", rc, runs(), out)
	}
}

// rvlSummary is a Playwright 1.62.1 list-reporter run: passing lines, an error
// header, and the failed / interrupted / flaky summary blocks, one block
// header wrapped in ANSI colour codes.
const rvlSummary = "  ✓  1 [fidelity] › fidelity/a.spec.ts:3:1 › a\n" +
	"  1) [fidelity] › fidelity/templates.spec.ts:97:3 › templates › list\n" +
	"\n" +
	"\x1b[31m  2 failed\x1b[39m\n" +
	"    [fidelity] › fidelity/templates.spec.ts:97:3 › templates › list\n" +
	"    [functional] › tests/auth.spec.ts:132:1 › signs in @smoke\n" +
	"  1 interrupted\n" +
	"    [fidelity] › fidelity/b.spec.ts:8:1 › b\n" +
	"  1 flaky\n" +
	"    [fidelity] › fidelity/kan-754-offline-l3.spec.ts:380:5 › offline\n" +
	"  140 passed (3.2m)\n"

var rvlSummaryFailures = []string{"fidelity/templates.spec.ts:97", "tests/auth.spec.ts:132", "fidelity/b.spec.ts:8", "fidelity/kan-754-offline-l3.spec.ts:380"}

func TestRvlVisualFailuresReadsTheSummaryBlocks(t *testing.T) {
	t.Parallel()
	if got := rvlVisualFailures([]byte(rvlSummary + rvlSummary)); !slices.Equal(got, rvlSummaryFailures) {
		t.Errorf("rvlVisualFailures = %q, want %q", got, rvlSummaryFailures)
	}
}

// rvlScript writes body as an executable script and returns `bash <path>`,
// a table-safe `verify` or `specs` cell.
func rvlScript(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "cell.sh")
	writeExec(t, p, "#!/usr/bin/env bash\n"+body)
	return "bash " + p
}

// rvlFailingList is the worktree's kept visual-failing.txt path.
func rvlFailingList(t *testing.T, wt string) string {
	t.Helper()
	return filepath.Join(rvlGit(t, wt, "rev-parse", "--path-format=absolute", "--git-path", "flow-verify-failed"), "visual-failing.txt")
}

func TestRunVerifyListWritesTheFailingVisualTests(t *testing.T) {
	t.Parallel()
	summary := filepath.Join(t.TempDir(), "summary")
	writeFile(t, summary, rvlSummary)
	wt, _ := rvlVisualSetup(t, rvlScript(t, "cat "+summary+"\nexit 1\n")+" <specs>", "web/src/a.tsx", "", "", false, false, false)
	rc, out := rvlRunIn(t, wt, "", wt, "visual")
	list := rvlFailingList(t, wt)
	b, _ := os.ReadFile(list)
	if rc != 1 || string(b) != strings.Join(rvlSummaryFailures, "\n")+"\n" || !strings.Contains(out, "  failing list: "+list+"\n") {
		t.Errorf("exit %d, %s holds %q, output:\n%s", rc, list, b, out)
	}
}

func TestRunVerifyListGroupsFailuresByTheChangeSpecs(t *testing.T) {
	t.Parallel()
	summary := filepath.Join(t.TempDir(), "summary")
	writeFile(t, summary, rvlSummary)
	verify := rvlScript(t, "cat "+summary+"\nexit 1\n") + " <specs>"
	wt, _ := rvlVisualSetup(t, verify, "web/src/a.tsx", "", "echo fidelity/templates.spec.ts:12 tests/auth.spec.ts", false, false, false)
	_, out := rvlRunIn(t, wt, "", wt, "visual")
	for _, want := range []string{
		"  failing in the change's specs (2): fidelity/templates.spec.ts:97 tests/auth.spec.ts:132\n",
		"  failing elsewhere (2): fidelity/b.spec.ts:8 fidelity/kan-754-offline-l3.spec.ts:380\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
	// With no `specs` row every failure is in the change's specs.
	wt, _ = rvlVisualSetup(t, verify, "web/src/a.tsx", "", "", false, false, false)
	if _, out := rvlRunIn(t, wt, "", wt, "visual"); !strings.Contains(out, "  failing in the change's specs (4): "+strings.Join(rvlSummaryFailures, " ")+"\n") || !strings.Contains(out, "  failing elsewhere (0):\n") {
		t.Errorf("no specs row: output:\n%s", out)
	}
}

func TestRunVerifyListEmptiesTheFailingListWhenNoneParse(t *testing.T) {
	t.Parallel()
	wt, _ := rvlVisualSetup(t, rvlScript(t, "echo boom\nexit 1\n")+" <specs>", "web/src/a.tsx", "", "", false, false, false)
	list := rvlFailingList(t, wt)
	writeFile(t, list, "fidelity/old.spec.ts:1\n")
	rc, out := rvlRunIn(t, wt, "", wt, "visual")
	b, err := os.ReadFile(list)
	if rc != 1 || err != nil || len(b) != 0 || !strings.Contains(out, "  failing list: none parsed\n") {
		t.Errorf("exit %d, %s holds %q (%v), output:\n%s", rc, list, b, err, out)
	}
}

func TestRunVerifyListReportsAnUnwritableFailingList(t *testing.T) {
	t.Parallel()
	summary := filepath.Join(t.TempDir(), "summary")
	writeFile(t, summary, rvlSummary)
	wt, _ := rvlVisualSetup(t, rvlScript(t, "cat "+summary+"\nexit 1\n")+" <specs>", "web/src/a.tsx", "", "", false, false, false)
	list := rvlFailingList(t, wt)
	if err := os.MkdirAll(list, 0o777); err != nil {
		t.Fatal(err)
	}
	_, out := rvlRunIn(t, wt, "", wt, "visual")
	if strings.Contains(out, "  failing list: "+list+"\n") || !strings.Contains(out, "  failing list: not written — ") {
		t.Errorf("an unwritten list was reported as written:\n%s", out)
	}
}

// rvlNarrowFixture is a visual-verify fixture whose `verify` cell writes its
// argv to argv and prints out, exiting rc, and whose `specs` cell touches
// marker and prints fidelity/a.spec.ts.
func rvlNarrowFixture(t *testing.T, out string, rc int) (wt, argv, marker string) {
	t.Helper()
	d := t.TempDir()
	argv, marker = filepath.Join(d, "argv"), filepath.Join(d, "marker")
	writeFile(t, filepath.Join(d, "out"), out)
	verify := rvlScript(t, "echo \"$*\" > "+argv+"\ncat "+filepath.Join(d, "out")+"\nexit "+strconv.Itoa(rc)+"\n") + " <specs>"
	specs := rvlScript(t, "touch "+marker+"\necho fidelity/a.spec.ts\n")
	wt, _ = rvlVisualSetup(t, verify, "web/src/a.tsx", "", specs, false, false, false)
	return wt, argv, marker
}

func TestRunVerifyListOnlyRunsTheFailingList(t *testing.T) {
	t.Parallel()
	wt, argv, _ := rvlNarrowFixture(t, "", 0)
	list := rvlFailingList(t, wt)
	writeFile(t, list, "fidelity/b.spec.ts:8\ntests/c.spec.ts:4\n")
	rc, out := rvlRunIn(t, wt, "", "-only", list, wt, "visual")
	got, _ := os.ReadFile(argv)
	if want := "fidelity/b.spec.ts:8 tests/c.spec.ts:4 --retries=1\n"; rc != 0 || string(got) != want {
		t.Errorf("exit %d, argv %q, want %q\n%s", rc, got, want, out)
	}
	for _, f := range []string{"flow-test-green", "flow-test-green-cmds"} {
		if _, err := os.Stat(rvlGit(t, wt, "rev-parse", "--path-format=absolute", "--git-path", f)); err == nil {
			t.Errorf("-only wrote %s", f)
		}
	}
}

// TestRunVerifyListOnlyGroupsFailuresByTheChangeSpecs: a -only round groups
// its failures by the `specs` row's specs, so a failure outside them counts
// as elsewhere, while the list alone picks what runs (panel-fix-1, F9).
func TestRunVerifyListOnlyGroupsFailuresByTheChangeSpecs(t *testing.T) {
	t.Parallel()
	wt, argv, _ := rvlNarrowFixture(t, "  2 failed\n    [fidelity] › fidelity/a.spec.ts:20:3 › a\n    [fidelity] › tests/c.spec.ts:4:3 › c\n", 1)
	list := filepath.Join(t.TempDir(), "only")
	writeFile(t, list, "fidelity/a.spec.ts:20\ntests/c.spec.ts:4\n")
	_, out := rvlRunIn(t, wt, "", "-only", list, wt, "visual")
	got, _ := os.ReadFile(argv)
	if want := "fidelity/a.spec.ts:20 tests/c.spec.ts:4 --retries=1\n"; string(got) != want {
		t.Errorf("argv %q, want %q\n%s", got, want, out)
	}
	for _, want := range []string{"  failing in the change's specs (1): fidelity/a.spec.ts:20\n", "  failing elsewhere (1): tests/c.spec.ts:4\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
}

func TestRvlNextRoundIsNumericAndClaimsTheRound(t *testing.T) {
	t.Parallel()
	keep := t.TempDir()
	writeFile(t, filepath.Join(keep, "visual-plain-r2.log"), "")
	writeFile(t, filepath.Join(keep, "visual-plain-r10.log"), "")
	// r10 sorts before r2: the highest is taken numerically (F12), and a
	// second claim before any log is written takes the next round (F14).
	if a, b := rvlNextRound(keep), rvlNextRound(keep); a != 11 || b != 12 {
		t.Errorf("rounds %d, %d, want 11, 12", a, b)
	}
}

// TestRunVerifyListOnlyReportsARetryPassAsFlaky pins the collapsed
// failing-run protocol (kan-1091: plain 40 minutes, a detector 36, a
// five-repeat 49): the failing list runs once with one retry, a test that
// passes on it is reported flaky and never fails the run, and a test that
// fails twice still does.
func TestRunVerifyListOnlyReportsARetryPassAsFlaky(t *testing.T) {
	t.Parallel()
	wt, argv, _ := rvlNarrowFixture(t, "  1 flaky\n    [fidelity] › fidelity/a.spec.ts:20:3 › a\n", 0)
	list := rvlFailingList(t, wt)
	writeFile(t, list, "fidelity/a.spec.ts:20\n")
	rc, out := rvlRunIn(t, wt, "", "-only", list, wt, "visual")
	got, _ := os.ReadFile(argv)
	if want := "fidelity/a.spec.ts:20 --retries=1\n"; rc != 0 || string(got) != want {
		t.Errorf("exit %d, argv %q, want 0 and %q\n%s", rc, got, want, out)
	}
	if !strings.Contains(out, "  flaky (1): fidelity/a.spec.ts:20 — passed on retry: fix each at its source\n") {
		t.Errorf("the retry-pass is not reported flaky:\n%s", out)
	}
	if b, _ := os.ReadFile(list); string(b) != "fidelity/a.spec.ts:20\n" {
		t.Errorf("a flaky-only run left %q in the failing list, want its entry kept for the fixer", b)
	}
	wt, _, _ = rvlNarrowFixture(t, "  1 failed\n    [fidelity] › tests/c.spec.ts:4:3 › c\n  1 flaky\n    [fidelity] › fidelity/a.spec.ts:20:3 › a\n", 1)
	list = rvlFailingList(t, wt)
	writeFile(t, list, "fidelity/a.spec.ts:20\ntests/c.spec.ts:4\ntests/d.spec.ts:9\n")
	rc, out = rvlRunIn(t, wt, "", "-only", list, wt, "visual")
	if rc != 1 || !strings.Contains(out, "  flaky (1): fidelity/a.spec.ts:20 — ") {
		t.Errorf("a twice-failed test beside a flaky one: exit %d, want 1 and the flaky line\n%s", rc, out)
	}
	if b, _ := os.ReadFile(list); string(b) != "tests/c.spec.ts:4\nfidelity/a.spec.ts:20\n" {
		t.Errorf("the red run left %q, want what is still to fix", b)
	}
}

func TestRunVerifyListNarrowedRunRefusals(t *testing.T) {
	t.Parallel()
	wt, argv, marker := rvlNarrowFixture(t, "", 0)
	empty := filepath.Join(t.TempDir(), "empty")
	writeFile(t, empty, "")
	full := filepath.Join(t.TempDir(), "full")
	writeFile(t, full, "tests/c.spec.ts:4\n")
	missing := filepath.Join(t.TempDir(), "missing")
	noSpecs, _ := rvlVisualSetup(t, "echo whole", "web/src/a.tsx", "", "", false, false, false)
	for _, tc := range []struct {
		name, want string
		args       []string
	}{
		{"-only with a missing file", missing, []string{"-only", missing, wt, "visual"}},
		{"-only with an empty file", "is empty", []string{"-only", empty, wt, "visual"}},
		{"a verify cell with no <specs>", "echo whole", []string{"-only", full, noSpecs, "visual"}},
		{"the retired detector flag", "usage", []string{"-detect-flaky", wt, "visual"}},
		{"-only with the test list", "usage", []string{"-only", full, wt, "test"}},
		{"-only with the lint list", "usage", []string{"-only", full, wt, "lint"}},
		{"-only with recording arguments", "usage", []string{"-only", full, wt, "visual", "kan-1", "verify", "tok", "opus", "low"}},
	} {
		rc, out := rvlRunIn(t, wt, "", tc.args...)
		_, a := os.Stat(argv)
		_, m := os.Stat(marker)
		if rc != 2 || !strings.Contains(out, tc.want) || a == nil || m == nil {
			t.Errorf("%s: exit %d, ran verify %v, ran specs %v, output:\n%s", tc.name, rc, a == nil, m == nil, out)
		}
	}
}

func TestRunVerifyListKeepsEveryVisualRound(t *testing.T) {
	t.Parallel()
	counter := filepath.Join(t.TempDir(), "n")
	// Each run writes test-results/<n>/shot-diff.png holding its own n.
	verify := rvlScript(t, "n=$(( $(cat "+counter+" 2>/dev/null || echo 0) + 1 ))\necho $n > "+counter+"\nmkdir -p test-results/$n\necho $n > test-results/$n/shot-diff.png\nexit 1\n") + " <specs>"
	wt, _ := rvlVisualSetup(t, verify, "web/src/a.tsx", "", "", false, false, false)
	keep := filepath.Dir(rvlFailingList(t, wt))
	rvlRunIn(t, wt, "", wt, "visual")
	only := filepath.Join(t.TempDir(), "only")
	writeFile(t, only, "tests/c.spec.ts:4\n")
	_, out := rvlRunIn(t, wt, "", "-only", only, wt, "visual")
	for f, want := range map[string]string{
		"visual-plain-r1.log":             "",
		"visual-only-r2.log":              "",
		"visual-results-r1/shot-diff.png": "1\n",
		"visual-results-r2/shot-diff.png": "2\n",
	} {
		b, err := os.ReadFile(filepath.Join(keep, f))
		if err != nil || (want != "" && string(b) != want) {
			t.Errorf("%s: %q (%v), want %q\n%s", f, b, err, want, out)
		}
	}
	if !strings.Contains(out, "  result files: "+filepath.Join(keep, "visual-results-r2")+"\n") {
		t.Errorf("output does not name the kept results:\n%s", out)
	}
	rvlRunIn(t, wt, "", wt, "visual")
	for _, f := range []string{"visual-plain-r1.log", "visual-only-r2.log", "visual-plain-r3.log", "visual-results-r1", "visual-results-r2", "visual-results-r3"} {
		if _, err := os.Stat(filepath.Join(keep, f)); err != nil {
			t.Errorf("a third run left no %s: %v", f, err)
		}
	}
}

func TestRunVerifyListKeepsTheTestListsRunStartCleanup(t *testing.T) {
	t.Parallel()
	wt, _ := rvlVisualSetup(t, "true <specs>", "docs/a.md", "", "", false, false, false)
	keep := filepath.Dir(rvlFailingList(t, wt))
	writeFile(t, filepath.Join(keep, "test-0.log"), "stale\n")
	writeFile(t, filepath.Join(keep, "visual-plain-r1.log"), "round 1\n")
	rvlRunIn(t, wt, "", wt, "test")
	if _, err := os.Stat(filepath.Join(keep, "test-0.log")); err == nil {
		t.Error("a test run left the stale test-0.log")
	}
	if _, err := os.Stat(filepath.Join(keep, "visual-plain-r1.log")); err != nil {
		t.Errorf("a test run removed visual-plain-r1.log: %v", err)
	}
}

// rvlFlowStub writes a `flow` stub into a new directory and returns it with
// its call log: every call appends its argv to the log; `state get` prints
// the file state, exiting stateRC; `suite record` runs the command after
// `--`; anything else exits 0.
func rvlFlowStub(t *testing.T, state string, stateRC int) (dir, log string) {
	t.Helper()
	dir = t.TempDir()
	log = filepath.Join(dir, "calls")
	writeFile(t, filepath.Join(dir, "state.json"), state)
	writeExec(t, filepath.Join(dir, "flow"), "#!/usr/bin/env bash\necho \"$*\" >> "+log+"\n"+
		"if [ \"$1 $2\" = \"state get\" ]; then cat "+filepath.Join(dir, "state.json")+"; exit "+strconv.Itoa(stateRC)+"; fi\n"+
		"if [ \"$1 $2\" = \"suite record\" ]; then\n  while [ \"$1\" != -- ]; do shift; done\n  shift\n  exec \"$@\"\nfi\n")
	return dir, log
}

// rvlStateJSON is a state record whose worktrees map each path to its base.
func rvlStateJSON(t *testing.T, bases map[string]string) string {
	t.Helper()
	b, err := json.Marshal(map[string]any{"worktrees": bases})
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestRunVerifyListScopesSpecsToTheKickoffBase(t *testing.T) {
	t.Parallel()
	// origin/main: K, then a ui change, then M; the branch spectre/demo
	// leaves M with a docs change only, so only K's diff touches ui paths.
	wt, _ := rvlVisualSetup(t, "echo VISUAL [<specs>]", "docs/a.md", "", "echo <merge-base>", false, false, false)
	k := rvlGit(t, wt, "rev-parse", "HEAD~1")
	rvlGit(t, wt, "checkout", "-q", "-b", "spectre/demo", k)
	writeFile(t, filepath.Join(wt, "web/src/k.tsx"), "x\n")
	rvlGit(t, wt, "add", "-A")
	rvlGit(t, wt, "commit", "-q", "-m", "ui")
	writeFile(t, filepath.Join(wt, "README"), "m\n")
	rvlGit(t, wt, "add", "-A")
	rvlGit(t, wt, "commit", "-q", "-m", "m")
	rvlGit(t, wt, "push", "-q", "-f", "origin", "HEAD:main")
	rvlGit(t, wt, "fetch", "-q", "origin")
	writeFile(t, filepath.Join(wt, "docs/b.md"), "x\n")
	rvlGit(t, wt, "add", "-A")
	rvlGit(t, wt, "commit", "-q", "-m", "change")
	stub, _ := rvlFlowStub(t, rvlStateJSON(t, map[string]string{wt: k}), 0)
	if rc, out := rvlRunIn(t, wt, stub, wt, "visual"); rc != 0 || !strings.Contains(out, "VISUAL ["+k+"]") {
		t.Errorf("exit %d, the verify row did not run on the kickoff base %s:\n%s", rc, k, out)
	}
}

// TestRunVerifyListMatchesTheKickoffBaseThroughSymlinks: a record naming
// the worktree by a symlinked path still gives its kickoff base.
func TestRunVerifyListMatchesTheKickoffBaseThroughSymlinks(t *testing.T) {
	t.Parallel()
	wt, _ := rvlVisualSetup(t, "echo VISUAL [<specs>]", "web/src/a.tsx", "", "echo <merge-base>", false, false, false)
	k := rvlGit(t, wt, "rev-parse", "HEAD~1")
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(wt, link); err != nil {
		t.Fatal(err)
	}
	stub, _ := rvlFlowStub(t, rvlStateJSON(t, map[string]string{link: k}), 0)
	if _, out := rvlRunIn(t, wt, stub, wt, "visual"); !strings.Contains(out, "VISUAL ["+k+"]") || strings.Contains(out, "kickoff base: not recorded") {
		t.Errorf("the symlinked record key gave no kickoff base %s:\n%s", k, out)
	}
}

func TestRunVerifyListFallsBackWithoutAKickoffBase(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	wt, _ := rvlVisualSetup(t, "echo VISUAL [<specs>]", "web/src/a.tsx", "", "echo <merge-base>", false, false, false)
	mb := rvlGit(t, wt, "merge-base", "HEAD", "origin/main")
	for name, rc := range map[string]int{"no entry": 0, "flow fails": 1} {
		stub, _ := rvlFlowStub(t, rvlStateJSON(t, map[string]string{"/elsewhere": "abc"}), rc)
		_, out := rvlRunIn(t, wt, stub, wt, "visual")
		if !strings.Contains(out, "VISUAL ["+mb+"]") || !strings.Contains(out, "- kickoff base: not recorded for "+wt+" — selecting from merge-base HEAD origin/main\n") {
			t.Errorf("%s: output:\n%s", name, out)
		}
	}
}

func TestRunVerifyListReadsTheRecordFromTheCanonicalWorktree(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	wt, _ := rvlVisualSetup(t, "echo VISUAL [<specs>]", "web/src/a.tsx", "", "echo <merge-base>", false, false, false)
	rvlGit(t, wt, "branch", "-m", "spectre/demo")
	canon := t.TempDir()
	only := filepath.Join(t.TempDir(), "only")
	writeFile(t, only, "tests/c.spec.ts:4\n")
	for _, tc := range []struct {
		want string
		args []string
	}{
		{"state get -C " + canon + " kan-9\n", []string{wt, "visual", "kan-9", "verify", "mf-t", "opus", "low", canon}},
		{"state get -C " + canon + " demo\n", []string{"-only", only, wt, "visual", canon}},
		{"state get -C " + wt + " demo\n", []string{wt, "visual"}},
	} {
		stub, log := rvlFlowStub(t, "{}", 0)
		_, out := rvlRunIn(t, wt, stub, tc.args...)
		if calls, _ := os.ReadFile(log); !strings.Contains(string(calls), tc.want) {
			t.Errorf("%v: flow calls %q, want %q\n%s", tc.args, calls, tc.want, out)
		}
	}
}

func TestSiblingFrontendPicksTheFirstMatchingSibling(t *testing.T) {
	t.Parallel()
	bases := map[string]string{"canon": "c", "s1": "b1", "s2": "b2"}
	for _, tc := range []struct {
		name       string
		match      []string
		root, base string
	}{
		{"s2 only", []string{"s2"}, "s2", "b2"},
		{"both", []string{"s1", "s2"}, "s1", "b1"},
		{"none", nil, "canon", "c"},
	} {
		root, base := siblingFrontend("canon", []string{"s1", "s2"}, bases, func(s string) bool { return slices.Contains(tc.match, s) })
		if root != tc.root || base != tc.base {
			t.Errorf("%s: %s %s, want %s %s", tc.name, root, base, tc.root, tc.base)
		}
	}
}

func TestRunVerifyListRunsSpecsFromTheSiblingFrontend(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	wt, _ := rvlVisualSetup(t, "echo VISUAL [<specs>]", "docs/a.md", "", "echo <frontend-root> <merge-base>", false, false, false)
	// A sibling worktree of a repository named `web`, whose diff touches
	// src/a.tsx -- web/src/a.tsx under the canonical `ui paths`.
	sib := filepath.Join(t.TempDir(), "web")
	writeFile(t, filepath.Join(sib, "README"), "x\n")
	fixtureRepo(t, sib, repoShape{Branch: "main"})
	rvlGit(t, sib, "add", "-A")
	rvlGit(t, sib, "commit", "-q", "-m", "base")
	sibBase := rvlGit(t, sib, "rev-parse", "HEAD")
	writeFile(t, filepath.Join(sib, "src/a.tsx"), "x\n")
	rvlGit(t, sib, "add", "-A")
	rvlGit(t, sib, "commit", "-q", "-m", "ui")
	stub, _ := rvlFlowStub(t, rvlStateJSON(t, map[string]string{wt: rvlGit(t, wt, "merge-base", "HEAD", "origin/main"), sib: sibBase}), 0)
	if rc, out := rvlRunIn(t, wt, stub, wt, "visual"); rc != 0 || !strings.Contains(out, "VISUAL ["+sib+" "+sibBase+"]") {
		t.Errorf("exit %d, the specs row did not read the sibling frontend %s at %s:\n%s", rc, sib, sibBase, out)
	}
}

// TestRunVerifyListRecordsTheVisualSuiteOnPlainRuns pins decision
// suite-records-scope: a plain run of the visual `verify` row, alone or as a
// `test` list's tail, runs under `flow suite record -suite visual`; a
// narrowed run never does, and with no `flow` the row runs bare.
func TestRunVerifyListRecordsTheVisualSuiteOnPlainRuns(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	wt, _ := rvlVisualSetup(t, "echo VISUAL-RAN <specs>", "web/src/a.tsx", "", "", false, false, false)
	list := rvlFailingList(t, wt)
	writeFile(t, list, "tests/c.spec.ts:4\n")
	for _, tc := range []struct {
		args   []string
		record bool
	}{
		{[]string{wt, "visual"}, true},
		{[]string{wt, "test"}, true},
		{[]string{"-only", list, wt, "visual"}, false},
	} {
		stub, log := rvlFlowStub(t, "{}", 0)
		rc, out := rvlRunIn(t, wt, stub, tc.args...)
		calls, _ := os.ReadFile(log)
		if got := strings.Contains(string(calls), "suite record -C "+wt+" -suite visual -- bash -c echo VISUAL-RAN"); rc != 0 || got != tc.record || !strings.Contains(out, "  VISUAL-RAN") {
			t.Errorf("%v: exit %d, recorded %v, want %v; flow calls %q\n%s", tc.args, rc, got, tc.record, calls, out)
		}
	}
	if rc, out := rvlRunIn(t, wt, "", wt, "visual"); rc != 0 || !strings.Contains(out, "  VISUAL-RAN") {
		t.Errorf("without flow: exit %d\n%s", rc, out)
	}
}

// TestRunVerifyListRecordsNoRowForAGuardSkippedSuite pins
// skipped-suite-records-no-row: a `flow-guard unchanged <pathspecs> ||` line
// the guard skips runs in process, prints a skip line and records no suite
// row; a line of another shape, or whose `||` is followed by more than one
// command, runs as before.
func TestRunVerifyListRecordsNoRowForAGuardSkippedSuite(t *testing.T) {
	t.Parallel()
	// fixture: the base commit holds the list, origin/main; one commit adds docs/a.md.
	fixture := func(t *testing.T, list string) (repo, base string) {
		t.Helper()
		repo, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(repo, ".flow/project.md"), "## test\n\n```bash\n"+list+"```\n")
		fixtureRepo(t, repo, repoShape{Branch: "main"})
		rvlGit(t, repo, "add", "-A")
		rvlGit(t, repo, "commit", "-q", "-m", "base")
		rvlGit(t, repo, "update-ref", "refs/remotes/origin/main", "HEAD")
		base = rvlGit(t, repo, "rev-parse", "HEAD")
		writeFile(t, filepath.Join(repo, "docs/a.md"), "a\n")
		rvlGit(t, repo, "add", "-A")
		rvlGit(t, repo, "commit", "-q", "-m", "docs")
		return repo, base
	}
	const quiet = "# suite: quiet\nflow-guard unchanged . ':!docs/' || { echo QUIET-RAN; }\n"
	t.Run("only the guard-skipped line records no row", func(t *testing.T) {
		t.Parallel()
		repo, base := fixture(t, quiet+
			"# suite: loud\nflow-guard unchanged docs/ || { echo LOUD-RAN; }\n"+
			"# suite: odd\nflow-guard unchanged $(echo src/) || { echo ODD-RAN; }\n"+
			"# suite: chain\nflow-guard unchanged . ':!docs' || echo A && echo CHAIN-RAN\n")
		stub, flowLog := rvlFlowStub(t, "", 1)
		guardLog := filepath.Join(stub, "guard-calls")
		writeExec(t, filepath.Join(stub, "flow-guard"), "#!/usr/bin/env bash\necho \"$*\" >> "+guardLog+"\nexit 1\n")

		rc, out := rvlRunIn(t, repo, stub, repo, "test")
		if rc != 0 {
			t.Fatalf("exit %d, want 0\n%s", rc, out)
		}
		if want := "`flow-guard unchanged . ':!docs/' || { echo QUIET-RAN; }` — skipped, unchanged: nothing under . :!docs/ changed since " + base + " — suite skipped\n"; !strings.Contains(out, want) || strings.Contains(out, "QUIET-RAN\n") {
			t.Errorf("the quiet suite was not skipped by its guard, want %q:\n%s", want, out)
		}
		for _, c := range []string{"`flow-guard unchanged docs/ || { echo LOUD-RAN; }` — exit 0\n  LOUD-RAN", "`flow-guard unchanged $(echo src/) || { echo ODD-RAN; }` — exit 0\n  ODD-RAN", "CHAIN-RAN"} {
			if !strings.Contains(out, c) {
				t.Errorf("output lacks %q:\n%s", c, out)
			}
		}
		calls, _ := os.ReadFile(flowLog)
		if !strings.Contains(string(calls), "-suite loud ") || !strings.Contains(string(calls), "-suite odd ") || !strings.Contains(string(calls), "-suite chain ") || strings.Contains(string(calls), "-suite quiet ") {
			t.Errorf("suite record calls want loud and odd, not quiet:\n%s", calls)
		}
		if g, _ := os.ReadFile(guardLog); strings.Contains(string(g), ":!docs/") {
			t.Errorf("the quiet suite's guard ran as a child:\n%s", g)
		}
	})

	t.Run("a list the guard skips whole ends its row skipped", func(t *testing.T) {
		t.Parallel()
		repo, _ := fixture(t, quiet)
		stub, flowLog := rvlFlowStub(t, "", 1)
		writeExec(t, filepath.Join(stub, "flow-guard"), "#!/usr/bin/env bash\necho \"$*\" >> "+filepath.Join(stub, "guard-calls")+"\nexit 1\n")
		if rc, out := rvlRunIn(t, repo, stub, repo, "test", "kan-1", "verify", "mf-k1", "opus", "low"); rc != 0 {
			t.Fatalf("quiet alone: exit %d, want 0\n%s", rc, out)
		}
		calls, _ := os.ReadFile(flowLog)
		if !strings.Contains(string(calls), "-outcome skipped") || strings.Contains(string(calls), "-outcome completed") {
			t.Errorf("the row of a guard-skipped list did not end skipped:\n%s", calls)
		}
	})
}

// rvlFullRepo is a clean repository whose origin/main is its HEAD and whose
// `## test` fence is list.
func rvlFullRepo(t *testing.T, list string) string {
	t.Helper()
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(repo, ".flow/project.md"), "## test\n\n```bash\n"+list+"```\n")
	fixtureRepo(t, repo, repoShape{Branch: "main"})
	rvlGit(t, repo, "add", "-A")
	rvlGit(t, repo, "commit", "-q", "-m", "base")
	rvlGit(t, repo, "update-ref", "refs/remotes/origin/main", "HEAD")
	return repo
}

// TestRunVerifyListFullRunsEverySuite pins full-is-the-list-unskipped and
// full-rows-carry-a-suffix.
func TestRunVerifyListFullRunsEverySuite(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	repo := rvlFullRepo(t, "# suite: quiet\n"+guardBinary(t)+" unchanged . || { echo QUIET-RAN; }\n# suite: plain\necho \"FULL=$FLOW_VERIFY_FULL\"\n")
	stub, log := rvlFlowStub(t, "", 1)
	rc, out := rvlRunIn(t, repo, stub, repo, "test")
	if rc != 0 || !strings.Contains(out, "— suite skipped") || strings.Contains(out, "  QUIET-RAN\n") || !strings.Contains(out, "  FULL=\n") {
		t.Fatalf("default run: exit %d, want 0, quiet skipped by its guard and FULL= printed:\n%s", rc, out)
	}
	rvlGit(t, repo, "commit", "-q", "--allow-empty", "-m", "empty")
	head := rvlGit(t, repo, "rev-parse", "HEAD")
	os.Remove(log)
	rc, out = rvlRunEnv(t, repo, stub, rvlOperator, "-full", repo, "test", "kan-1", "verify", "mf-k1", "opus", "low")
	if rc != 0 || strings.Contains(out, "skipped") || !strings.Contains(out, "  QUIET-RAN\n") || !strings.Contains(out, "  FULL=1\n") {
		t.Errorf("-full run: exit %d, want 0, nothing skipped, QUIET-RAN and FULL=1:\n%s", rc, out)
	}
	calls, _ := os.ReadFile(log)
	for _, want := range []string{"-suite quiet-full ", "-suite plain-full ", "dispatch begin", "-role verifier", "-outcome completed"} {
		if !strings.Contains(string(calls), want) {
			t.Errorf("flow log lacks %q:\n%s", want, calls)
		}
	}
	if strings.Contains(string(calls), "-suite quiet --") || strings.Contains(string(calls), "-suite plain --") {
		t.Errorf("-full recorded a default row:\n%s", calls)
	}
	rc, out = rvlRunIn(t, repo, stub, repo, "test")
	if want := " — skipped, no input changed since " + head + " passed the list\n"; rc != 0 || strings.Count(out, want) != 2 {
		t.Errorf("default run after -full: exit %d, want both commands skipped on %s:\n%s", rc, head, out)
	}
}

// TestRunVerifyListFullNamesAnEscape pins full-failure-names-its-escape.
func TestRunVerifyListFullNamesAnEscape(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	const a, b, c = `test -z "$FLOW_VERIFY_FULL"`, `test -n "$FLOW_VERIFY_FULL"`, "test -f .git/never"
	repo := rvlFullRepo(t, a+"\n"+b+"\n"+c+"\n")
	head := rvlGit(t, repo, "rev-parse", "HEAD")
	if rc, out := rvlRunIn(t, repo, "", repo, "test"); rc != 1 || !strings.Contains(out, "`"+a+"` — exit 0") {
		t.Fatalf("default run: exit %d, want 1 with a passed:\n%s", rc, out)
	}
	rc, out := rvlRunEnv(t, repo, "", rvlOperator, "-full", repo, "test")
	escape := "  escaped the default list — it passed at " + head + "\n"
	block := func(cmd, next string) string {
		_, rest, _ := strings.Cut(out, "- `"+cmd+"` — ")
		if next != "" {
			rest, _, _ = strings.Cut(rest, "- `"+next+"`")
		}
		return rest
	}
	if rc != 1 || !strings.Contains(block(a, b), escape) || !strings.Contains(out, "`"+b+"` — exit 0") ||
		strings.Contains(block(c, ""), "escaped") || strings.Count(out, "escaped the default list") != 1 {
		t.Errorf("-full run: exit %d, want 1 and the escape line in a's block alone:\n%s", rc, out)
	}

	repo = rvlFullRepo(t, a+"\n")
	head = rvlGit(t, repo, "rev-parse", "HEAD")
	if rc, out := rvlRunIn(t, repo, "", repo, "test"); rc != 0 {
		t.Fatalf("default run of a alone: exit %d:\n%s", rc, out)
	}
	if rc, out := rvlRunEnv(t, repo, "", rvlOperator, "-full", repo, "test"); rc != 1 || !strings.Contains(out, "  escaped the default list — it passed at "+head+"\n") {
		t.Errorf("-full run of a alone: exit %d, want 1 and the escape line on the green sha:\n%s", rc, out)
	}
}

// TestRunVerifyListFullRunsTheVisualSuiteWhole pins
// full-runs-the-visual-suite-whole.
func TestRunVerifyListFullRunsTheVisualSuiteWhole(t *testing.T) {
	t.Parallel()
	wt, _ := rvlVisualSetup(t, "echo VISUAL [<specs>]", "docs/x.md", `{"visual":{"verify":"skipped"}}`, "echo SPECS-ROW-RAN", false, false, false)
	stub, log := rvlFlowStub(t, "", 1)
	if _, out := rvlRunIn(t, wt, stub, wt, "visual"); !strings.Contains(out, "- visual verify — not run:") {
		t.Errorf("default visual list ran the row:\n%s", out)
	}
	for _, list := range []string{"visual", "test"} {
		os.Remove(log)
		rc, out := rvlRunEnv(t, wt, stub, rvlOperator, "-full", wt, list)
		calls, _ := os.ReadFile(log)
		if rc != 0 || !strings.Contains(out, "VISUAL []") || strings.Contains(out, "SPECS-ROW-RAN") || !strings.Contains(string(calls), "-suite visual-full ") {
			t.Errorf("-full %s: exit %d, want the whole suite run as visual-full:\n%s\nflow log:\n%s", list, rc, out, calls)
		}
	}
}

// TestRunVerifyListFullUsage: -full refuses the lint list, a narrowing flag
// and a missing worktree before any command runs.
func TestRunVerifyListFullUsage(t *testing.T) {
	t.Parallel()
	repo := rvlFullRepo(t, "touch ran\n")
	writeFile(t, filepath.Join(repo, "only.txt"), "a.spec.ts\n")
	for _, args := range [][]string{
		{"-full", repo, "lint"},
		{"-full", "-only", filepath.Join(repo, "only.txt"), repo, "visual"},
		{"-full"},
	} {
		var out, errb bytes.Buffer
		rc := runVerifyList(args, Env{Getenv: func(k string) string { return map[string]string{rvlOperatorFull: "1"}[k] }, Environ: os.Environ, Dir: repo}, &out, &errb)
		if _, err := os.Stat(filepath.Join(repo, "ran")); rc != 2 || errb.Len() == 0 || err == nil {
			t.Errorf("%q: rc=%d stderr=%q, want 2, a stderr line and no command run", args, rc, errb.String())
		}
	}
}

// TestRunVerifyListReRunsTheVisualRowWhenASiblingMoves pins kan-1087 self-review F7: the visual
// row's green keys on every sibling worktree's HEAD, since its specs read the siblings' code.
func TestRunVerifyListReRunsTheVisualRowWhenASiblingMoves(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	wt, _ := rvlVisualSetup(t, "echo VISUAL <specs>", "docs/a.md", "", "", false, false, false)
	sib := filepath.Join(t.TempDir(), "web")
	writeFile(t, filepath.Join(sib, "README"), "x\n")
	fixtureRepo(t, sib, repoShape{Branch: "main"})
	rvlGit(t, sib, "add", "-A")
	rvlGit(t, sib, "commit", "-q", "-m", "base")
	sibBase := rvlGit(t, sib, "rev-parse", "HEAD")
	writeFile(t, filepath.Join(sib, "src/a.tsx"), "x\n")
	rvlGit(t, sib, "add", "-A")
	rvlGit(t, sib, "commit", "-q", "-m", "ui")
	stub, _ := rvlFlowStub(t, rvlStateJSON(t, map[string]string{wt: rvlGit(t, wt, "merge-base", "HEAD", "origin/main"), sib: sibBase}), 0)
	ran := "`echo VISUAL --grep @smoke` — exit 0"
	args := []string{wt, "test", "demo", "verify-" + filepath.Base(wt), "mf-test", "opus", "low"}
	if rc, out := rvlRunIn(t, wt, stub, args...); rc != 0 || !strings.Contains(out, ran) {
		t.Fatalf("run 1: exit %d, the visual row did not run:\n%s", rc, out)
	}
	if rc, out := rvlRunIn(t, wt, stub, args...); rc != 0 || strings.Contains(out, ran) {
		t.Fatalf("run 2: exit %d, the visual row ran with nothing moved:\n%s", rc, out)
	}
	writeFile(t, filepath.Join(sib, "src/b.tsx"), "x\n")
	rvlGit(t, sib, "add", "-A")
	rvlGit(t, sib, "commit", "-q", "-m", "more ui")
	if rc, out := rvlRunIn(t, wt, stub, args...); rc != 0 || !strings.Contains(out, ran) {
		t.Fatalf("run 3: exit %d, the visual row was skipped after the sibling moved:\n%s", rc, out)
	}
}

// TestRunVerifyListFullNeedsTheOperator pins THE OPERATOR'S SWITCH: -full
// runs nothing unless FLOW_OPERATOR_FULL_SUITE=1 is set, and a default run
// drops an inherited FLOW_VERIFY_FULL.
func TestRunVerifyListFullNeedsTheOperator(t *testing.T) {
	t.Parallel()
	repo := rvlFullRepo(t, "touch ran\necho \"FULL=$FLOW_VERIFY_FULL\"\n")
	for _, list := range []string{"test", "visual"} {
		for _, extra := range [][]string{nil, {rvlOperatorFull + "=yes"}} {
			rc, out := rvlRunEnv(t, repo, "", extra, "-full", repo, list)
			if _, err := os.Stat(filepath.Join(repo, "ran")); rc != 2 || err == nil || !strings.Contains(out, rvlOperatorFull+"=1") {
				t.Errorf("-full %s with %q: exit %d, want 2, no command run and the variable named:\n%s", list, extra, rc, out)
			}
		}
	}
	if rc, out := rvlRunEnv(t, repo, "", []string{"FLOW_VERIFY_FULL=1"}, repo, "test"); rc != 0 || !strings.Contains(out, "  FULL=\n") {
		t.Errorf("default run under an inherited FLOW_VERIFY_FULL=1: exit %d, want 0 and FULL= printed:\n%s", rc, out)
	}
}

// rvlWideFixture is a worktree whose change edits tests/s7.spec.ts beside a
// UI file, and whose `specs` row prints n spec files for a non-empty diff and
// tests/smoke.spec.ts:1, its smoke tail, for any diff. The `verify` row writes
// its arguments to argv.
func rvlWideFixture(t *testing.T, n int) (wt, argv string) {
	t.Helper()
	argv = filepath.Join(t.TempDir(), "argv")
	verify := rvlScript(t, "echo \"$*\" > "+argv+"\n") + " <specs>"
	specs := rvlScript(t, `[ -z "$(git -C "$1" diff --name-only "$2")" ] || for i in $(seq 0 `+strconv.Itoa(n-1)+`); do echo "tests/s$i.spec.ts:3"; done
echo tests/smoke.spec.ts:1
`) + " <frontend-root> <merge-base>"
	wt, _ = rvlVisualSetup(t, verify, "web/src/a.tsx", "", specs, false, false, false)
	writeFile(t, filepath.Join(wt, "tests/s7.spec.ts"), "x\n")
	rvlGit(t, wt, "add", "-A")
	rvlGit(t, wt, "commit", "-q", "-m", "spec")
	return wt, argv
}

// TestRunVerifyListNarrowsAWideSelection pins A WIDE SELECTION (kan-1089,
// kan-1091, kan-1099: a change touching every screen selected 646-679 of 680
// tests, 17 runs of 35-50 minutes): a selection naming more than
// rvlWideSelection spec files runs the specs the change edits plus the
// `specs` row's tail, and is never refused; one naming that many or fewer,
// and a -only list of any size, runs as it is.
func TestRunVerifyListNarrowsAWideSelection(t *testing.T) {
	t.Parallel()
	wt, argv := rvlWideFixture(t, 500)
	for _, tc := range []struct {
		want string
		args []string
	}{
		{"tests/s7.spec.ts:3 tests/smoke.spec.ts:1\n", []string{wt, "visual"}},
		{"tests/s7.spec.ts:3 tests/smoke.spec.ts:1\n", []string{wt, "test"}},
	} {
		os.Remove(argv)
		rc, out := rvlRunIn(t, wt, "", tc.args...)
		got, _ := os.ReadFile(argv)
		if note := "- visual selection: 501 spec files reached, more than 40 — narrowed to the 1 the change edits plus the smoke tail"; rc != 0 || string(got) != tc.want || !strings.Contains(out, note) {
			t.Errorf("%v: exit %d, argv %q, want %q and %q:\n%.900s", tc.args, rc, got, tc.want, note, out)
		}
	}
	var b strings.Builder
	for i := range 500 {
		fmt.Fprintf(&b, "tests/s%d.spec.ts:3 ", i)
	}
	only := filepath.Join(t.TempDir(), "only")
	writeFile(t, only, b.String())
	if rc, out := rvlRunIn(t, wt, "", "-only", only, wt, "visual"); rc != 0 || strings.Contains(out, "narrowed") {
		t.Errorf("-only: exit %d, want 0 and no narrowing:\n%.600s", rc, out)
	}
	if got, _ := os.ReadFile(argv); !strings.Contains(string(got), "tests/s499.spec.ts:3 --retries=1") {
		t.Errorf("-only ran %.200q, want all 500 entries", got)
	}
	wt, argv = rvlWideFixture(t, 39) // 39 files and the tail's: 40, not over
	if rc, out := rvlRunIn(t, wt, "", wt, "visual"); rc != 0 || strings.Contains(out, "narrowed") {
		t.Errorf("40 spec files: exit %d, want 0 and no narrowing:\n%.600s", rc, out)
	}
	if got, _ := os.ReadFile(argv); !strings.Contains(string(got), "tests/s38.spec.ts:3 tests/smoke.spec.ts:1") {
		t.Errorf("40 spec files ran %.200q, want the whole selection", got)
	}
}

// TestRunVerifyListWideSelectionWithNothingOfItsOwnRunsTheSmokeSet: a wide
// selection whose change edits no selected spec, under a `specs` row that
// prints no tail, runs the smoke set (THE SMOKE FALLBACK), never nothing.
func TestRunVerifyListWideSelectionWithNothingOfItsOwnRunsTheSmokeSet(t *testing.T) {
	t.Parallel()
	specs := rvlScript(t, `[ -z "$(git -C "$1" diff --name-only "$2")" ] || for i in $(seq 0 60); do echo "tests/s$i.spec.ts"; done
`) + " <frontend-root> <merge-base>"
	wt, _ := rvlVisualSetup(t, "echo VISUAL-RAN [<specs>]", "web/src/a.tsx", "", specs, false, false, false)
	if rc, out := rvlRunIn(t, wt, "", wt, "visual"); rc != 0 || !strings.Contains(out, "VISUAL-RAN [--grep @smoke]") {
		t.Errorf("exit %d, want 0 and the smoke set:\n%.900s", rc, out)
	}
}

// rvlRoundFixture is a worktree whose change touches a theme file, so the
// `specs` row maps every spec from the kickoff base, and whose fix round then
// touches screen b alone. The `verify` row writes its arguments to argv and
// exits rc; h1 is the HEAD the first verify ran at.
func rvlRoundFixture(t *testing.T, rc int) (wt, argv, h1 string) {
	t.Helper()
	argv = filepath.Join(t.TempDir(), "argv")
	verify := rvlScript(t, "echo \"$*\" > "+argv+"\ncat "+argv+".summary 2>/dev/null\nexit "+strconv.Itoa(rc)+"\n") + " <specs>"
	specs := rvlScript(t, `for f in $(git -C "$1" diff --name-only "$2"); do case $f in
web/src/theme.css) echo a.spec.ts; echo b.spec.ts; echo c.spec.ts;;
web/src/*.tsx) f=${f##*/}; echo "${f%.tsx}.spec.ts";;
esac; done | sort -u
`) + " <frontend-root> <merge-base>"
	wt, _ = rvlVisualSetup(t, verify, "web/src/theme.css", "", specs, false, false, false)
	h1 = rvlGit(t, wt, "rev-parse", "HEAD")
	writeFile(t, filepath.Join(wt, "web/src/b.tsx"), "x\n")
	rvlGit(t, wt, "add", "-A")
	rvlGit(t, wt, "commit", "-q", "-m", "fix")
	return wt, argv, h1
}

// rvlRows is a `flow record dispatches` answer of verifier rows, key then
// diffBase.
func rvlRows(t *testing.T, pairs ...string) func(string) ([]byte, error) {
	t.Helper()
	var rows []map[string]string
	for i := 0; i < len(pairs); i += 2 {
		rows = append(rows, map[string]string{"key": pairs[i], "role": "verifier", "diffBase": pairs[i+1]})
	}
	b, err := json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	return func(string) ([]byte, error) { return b, nil }
}

// TestRunVerifyListReRunSelectsFromTheRoundBase pins THE ROUND'S SELECTION
// (kan-1099): a first verify selects from the kickoff base -- every spec, the
// change touching the theme -- and a re-run from the HEAD the previous verify
// row recorded, never from a row of its own round.
func TestRunVerifyListReRunSelectsFromTheRoundBase(t *testing.T) {
	t.Parallel()
	wt, argv, h1 := rvlRoundFixture(t, 0)
	head := rvlGit(t, wt, "rev-parse", "HEAD")
	stub, _ := rvlFlowStub(t, "{}", 1)
	rows := rvlRows(t, "verify", h1, "verify-fix-1-"+filepath.Base(wt), head)
	for key, want := range map[string]string{"verify": "a.spec.ts b.spec.ts c.spec.ts\n", "verify-fix-1": "b.spec.ts\n"} {
		rc, out := rvlRunRows(t, wt, stub, nil, rows, wt, "visual", "work", key, "tok", "opus", "low")
		if got, _ := os.ReadFile(argv); rc != 0 || string(got) != want {
			t.Errorf("key %s: exit %d, argv %q, want %q\n%s", key, rc, got, want, out)
		}
	}
}

// TestRunVerifyListReRunCoversTheFailingList: a plain re-run adds the spec
// file of each visual-failing.txt entry to the round's selection, and a green
// one empties the file.
func TestRunVerifyListReRunCoversTheFailingList(t *testing.T) {
	t.Parallel()
	wt, argv, h1 := rvlRoundFixture(t, 0)
	list := rvlFailingList(t, wt)
	writeFile(t, list, "b.spec.ts:8\ntests/c.spec.ts:4\n")
	stub, _ := rvlFlowStub(t, "{}", 1)
	rc, out := rvlRunRows(t, wt, stub, nil, rvlRows(t, "verify", h1), wt, "visual", "work", "verify-fix-1", "tok", "opus", "low")
	got, _ := os.ReadFile(argv)
	if want := "b.spec.ts tests/c.spec.ts\n"; rc != 0 || string(got) != want {
		t.Errorf("exit %d, argv %q, want %q\n%s", rc, got, want, out)
	}
	if b, _ := os.ReadFile(list); len(b) != 0 {
		t.Errorf("a green re-run left %q in the failing list", b)
	}
}

// TestRunVerifyListOnlyGroupsAsTheFailingRowDid: -only carries no key, so
// its `specs` row selects as the newest verify row did -- from the kickoff
// base after a first verify, from the round base after a re-run.
func TestRunVerifyListOnlyGroupsAsTheFailingRowDid(t *testing.T) {
	t.Parallel()
	wt, argv, h1 := rvlRoundFixture(t, 1)
	head := rvlGit(t, wt, "rev-parse", "HEAD")
	writeFile(t, argv+".summary", "  1 failed\n    [p] › a.spec.ts:1:1 › a\n")
	only := filepath.Join(t.TempDir(), "only")
	writeFile(t, only, "a.spec.ts:1\n")
	for want, rows := range map[string]func(string) ([]byte, error){
		"  failing in the change's specs (1): a.spec.ts:1\n": rvlRows(t, "verify", head),
		"  failing elsewhere (1): a.spec.ts:1\n":             rvlRows(t, "verify", h1, "verify-fix-1", head),
	} {
		rc, out := rvlRunRows(t, wt, "", nil, rows, "-only", only, wt, "visual")
		if rc != 1 || !strings.Contains(out, want) {
			t.Errorf("exit %d, want 1 and %q\n%s", rc, want, out)
		}
	}
}

// TestRunVerifyListRefusesAnUnsuffixedKeyInAMultiWorktreeChange pins
// f5-verify-key-suffix (kan-1099 self-review, F5): a key folding to bare
// `verify` is refused when the change's state record holds more than one
// worktree, before any `begin`; the suffixed key, or a one-worktree change,
// runs.
func TestRunVerifyListRefusesAnUnsuffixedKeyInAMultiWorktreeChange(t *testing.T) {
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	wt := filepath.Join(tmp, "repo-a")
	if err := os.MkdirAll(wt, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(wt, ".flow/project.md"), "## lint\n\n```bash\ntrue\n```\n")
	run := func(state, key string) (int, string, string) {
		dir, log := rvlFlowStub(t, state, 0)
		environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "PATH="+dir+":"+os.Getenv("PATH"))
		getenv := func(k string) string {
			for i := len(environ) - 1; i >= 0; i-- {
				if v, ok := strings.CutPrefix(environ[i], k+"="); ok {
					return v
				}
			}
			return ""
		}
		var o, e bytes.Buffer
		rc := runVerifyList([]string{wt, "lint", "kan-1", key, "mf-k1", "sonnet", "medium", wt}, Env{Getenv: getenv, Environ: func() []string { return environ }, Dir: wt}, &o, &e)
		calls, _ := os.ReadFile(log)
		return rc, string(calls), e.String()
	}
	two := rvlStateJSON(t, map[string]string{wt: "abc", filepath.Join(tmp, "repo-b"): "def"})
	one := rvlStateJSON(t, map[string]string{wt: "abc"})
	if rc, calls, errs := run(two, "verify-fix-5"); rc != 2 || strings.Contains(calls, "record dispatch begin") || !strings.Contains(errs, "-<repository name>") {
		t.Errorf("unsuffixed key, two worktrees: exit %d, calls %q, stderr %q", rc, calls, errs)
	}
	if rc, calls, errs := run(two, "verify-fix-5-repo-a"); rc != 0 || !strings.Contains(calls, "record dispatch begin") {
		t.Errorf("suffixed key, two worktrees: exit %d, calls %q, stderr %q", rc, calls, errs)
	}
	if rc, calls, errs := run(one, "verify-fix-5"); rc != 0 || !strings.Contains(calls, "record dispatch begin") {
		t.Errorf("unsuffixed key, one worktree: exit %d, calls %q, stderr %q", rc, calls, errs)
	}
}

// TestRunVerifyListAppliesPrepareWorkspace pins f16-run-prepares-workspace
// (kan-1099 self-review, F16): the list runs prepare-workspace itself and
// gives every command its KEY=value lines, the caller's own value winning; a
// failing prepare-workspace stops the list, exit 2, before any command.
func TestRunVerifyListAppliesPrepareWorkspace(t *testing.T) {
	iso := func(res string) string {
		return "## workspace isolation\n\n| Resource | Variable | Default | In a workspace |\n|----------|----------|---------|----------------|\n" +
			"| `" + res + "` | `FRONTEND_URL` | `http://localhost:3000` | `http://localhost:4290` |\n" + pwCommands +
			"\n## lint\n\n```bash\ntouch ran\nprintf '%s\\n' \"$FRONTEND_URL\" > out.txt\n```\n"
	}
	run := func(wt string, extra ...string) (int, string) {
		environ := append(append(slices.DeleteFunc(os.Environ(), func(kv string) bool { return strings.HasPrefix(kv, "FRONTEND_URL=") }), "FLOW_GUARD_CACHE_DIR="+guardCache(t)), extra...)
		getenv := func(k string) string {
			for i := len(environ) - 1; i >= 0; i-- {
				if v, ok := strings.CutPrefix(environ[i], k+"="); ok {
					return v
				}
			}
			return ""
		}
		var o, e bytes.Buffer
		rc := runVerifyList([]string{wt, "lint"}, Env{Getenv: getenv, Environ: func() []string { return environ }, Dir: wt}, &o, &e)
		b, _ := os.ReadFile(filepath.Join(wt, "out.txt"))
		return rc, strings.TrimSpace(string(b)) + "\n" + o.String() + e.String()
	}
	wt := pwRepo(t, "demo", iso("url"))
	if _, out := run(wt); !strings.HasPrefix(out, "http://localhost:4290\n") {
		t.Errorf("unset: out.txt not the workspace value:\n%s", out)
	}
	if _, out := run(wt, "FRONTEND_URL=mine"); !strings.HasPrefix(out, "mine\n") {
		t.Errorf("caller's value: out.txt not mine:\n%s", out)
	}
	bad := pwRepo(t, "demo", iso("widget"))
	if rc, out := run(bad); rc != 2 || strings.Contains(out, "`touch ran`") {
		t.Errorf("failing prepare-workspace: exit %d, a command ran:\n%s", rc, out)
	}
	// A /flow-fast or /flow-interactive worktree is on `<name>`, carries no
	// workspace, and ran the list before f16: it still does, with no value.
	fast := pwRepo(t, "demo", iso("url"))
	gitRun(t, fast, "checkout", "-q", "-b", "kan-1-demo")
	if rc, out := run(fast); rc != 0 || !strings.HasPrefix(out, "\n") || !strings.Contains(out, "`touch ran`") {
		t.Errorf("a worktree on a bare <name> branch: exit %d, want 0, the list run and no workspace value:\n%s", rc, out)
	}
}

// TestRunVerifyListAbortsTheSupersededOpenRow pins kan-1099 self-review F11:
// a verify group's begin ends `-outcome aborted` the open inline verifier row
// of its own token and group that a killed run left, and nothing else.
func TestRunVerifyListAbortsTheSupersededOpenRow(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	rows := `[{"key":"verify-a","role":"verifier","agentId":"inline","sessionToken":"mf-k1"},` +
		`{"key":"verify-b","role":"verifier","agentId":"inline","sessionToken":"mf-k1"},` +
		`{"key":"verify-fix-1-a","role":"verifier","agentId":"inline","sessionToken":"mf-k1"},` +
		`{"key":"verify-a","role":"verifier","agentId":"inline","sessionToken":"mf-other"},` +
		`{"key":"verify-2-a","role":"verifier","agentId":"inline","sessionToken":"mf-k1","outcome":"completed"}]`
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	a := filepath.Join(wt, "a")
	fixtureRepo(t, a, repoShape{})
	writeFile(t, filepath.Join(a, ".flow/project.md"), "## lint\n\n```bash\necho ran >> "+log+"\n```\n")
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "PATH="+dir+":"+os.Getenv("PATH"))
	getenv := func(k string) string {
		for i := len(environ) - 1; i >= 0; i-- {
			if v, ok := strings.CutPrefix(environ[i], k+"="); ok {
				return v
			}
		}
		return ""
	}
	run := func(endRC int) (int, string, string) {
		os.Remove(log)
		writeExec(t, filepath.Join(dir, "flow"), fmt.Sprintf("#!/bin/sh\n[ \"$1\" = state ] && exit 0\n"+
			"[ \"$2\" = dispatches ] && { echo '%s'; exit 0; }\necho \"flow $*\" >> %s\n"+
			"[ \"$3\" = end ] && echo nope && exit %d\nexit 0\n", rows, log, endRC))
		var o, e bytes.Buffer
		rc := runVerifyList([]string{a, "lint", "kan-1", "verify-fix-1-a", "mf-k1", "opus", "low"},
			Env{Getenv: getenv, Environ: func() []string { return environ }, Dir: a}, &o, &e)
		calls, _ := os.ReadFile(log)
		return rc, string(calls), e.String()
	}
	abort := "flow record dispatch end -C " + a + " -change kan-1 -key verify-a -session-token mf-k1 -outcome aborted\n"
	rc, calls, _ := run(0)
	if i := strings.Index(calls, "dispatch begin"); rc != 0 || !strings.HasPrefix(calls, abort) || i < len(abort) ||
		strings.Count(calls, "-outcome aborted") != 1 {
		t.Errorf("the superseded row must be ended aborted before the begin, alone: exit %d, calls:\n%s", rc, calls)
	}
	rc, calls, stderr := run(1)
	if !strings.Contains(stderr, "superseded row verify-a failed") || !strings.Contains(calls, "ran\n") {
		t.Errorf("a failing abort must warn and run the list: exit %d, stderr %q, calls:\n%s", rc, stderr, calls)
	}
}

// TestRunVerifyListRecordsItsCost pins THE RUN'S COST: one `flow record run`
// row per run, under the branch's change, with the visual verify's test
// counts and the list's exit.
func TestRunVerifyListRecordsItsCost(t *testing.T) {
	t.Parallel()
	wt, _, _ := rvlNarrowFixture(t, "  1 failed\n    [fidelity] › fidelity/a.spec.ts:20:3 › a\n  1 flaky\n    [fidelity] › tests/c.spec.ts:4:3 › c\n  2 passed (1.0s)\n", 1)
	only := filepath.Join(t.TempDir(), "only")
	writeFile(t, only, "fidelity/a.spec.ts:20\n")
	for label, args := range map[string][]string{"visual": {wt, "visual"}, "visual-only": {"-only", only, wt, "visual"}} {
		stub, log := rvlFlowStub(t, "{}", 0)
		rvlRunIn(t, wt, stub, args...)
		calls, _ := os.ReadFile(log)
		if want := `record run -C ` + regexp.QuoteMeta(wt) + ` -change work -kind run-verify-list -label ` + label + ` -seconds \d+ -tests 4 -failed 1 -exit 1 -flaky 1\n`; !regexp.MustCompile(want).Match(calls) {
			t.Errorf("%s: flow calls %q, want %q", label, calls, want)
		}
	}
}

// TestRunVerifyListRecordsItsFlakyCount pins that the recorded run carries
// the summary's flaky count, 0 when the summary has no flaky block.
func TestRunVerifyListRecordsItsFlakyCount(t *testing.T) {
	t.Parallel()
	for summary, want := range map[string]string{
		"  1 flaky\n    [fidelity] › tests/c.spec.ts:4:3 › c\n  2 passed (1.0s)\n": "-flaky 1\n",
		"  2 passed (1.0s)\n": "-flaky 0\n",
	} {
		wt, _, _ := rvlNarrowFixture(t, summary, 0)
		stub, log := rvlFlowStub(t, "{}", 0)
		rvlRunIn(t, wt, stub, wt, "visual")
		calls, _ := os.ReadFile(log)
		if !regexp.MustCompile(`record run .* -exit 0 ` + want).Match(calls) {
			t.Errorf("summary %q: flow calls %q, want the run to end %q", summary, calls, want)
		}
	}
}
