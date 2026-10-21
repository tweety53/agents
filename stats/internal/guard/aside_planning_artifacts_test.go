package guard

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// apaBashAt writes scripts/aside-planning-artifacts.sh and the
// lib/spec-root.sh it sources, as they stood at ae805186 (Decision:
// parity-ref-on-main), into one temporary directory and returns the script's
// path there.
func apaBashAt(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, rel := range []string{"aside-planning-artifacts.sh", "lib/spec-root.sh"} {
		src := historicalFile(t, "../../..", "ae805186", "scripts/"+rel)
		writeFile(t, filepath.Join(dir, rel), string(src))
	}
	return dir + "/aside-planning-artifacts.sh"
}

// apaRepo is one fixture repository: a commit carrying a tracked file under
// each planning directory named in trees (`spectre`, `openspec`, `docs`) plus
// an implementation file, on branch demo. The identity is the repository's
// own config, since both guards run git under the process environment.
func apaRepo(t *testing.T, trees ...string) string {
	t.Helper()
	repo := t.TempDir() + "/repo"
	gitRun(t, "", "init", "-q", "-b", "demo", repo)
	gitRun(t, repo, "config", "user.email", "test@example.invalid")
	gitRun(t, repo, "config", "user.name", "Test")
	for _, tree := range trees {
		if tree == "docs" {
			writeFile(t, repo+"/docs/superpowers/ledger.md", "ledger\n")
		} else {
			writeFile(t, repo+"/"+tree+"/changes/demo/tasks.md", "plan\n")
		}
	}
	writeFile(t, repo+"/src.txt", "code\n")
	gitRun(t, repo, "add", "-A")
	gitRun(t, repo, "commit", "-qm", "base")
	return repo
}

// apaStep is one guard call, after an optional fixture edit. args may name
// the repository as $REPO. want pins the bash side's verdict, so two sides
// refusing alike can never pass for parity on the path a case names.
type apaStep struct {
	before func(t *testing.T, repo string)
	args   []string
	want   string // the verdict token stdout starts with; "" for an exit-2 refusal's empty stdout
}

type apaRes struct {
	code     int
	out, err string
}

var apaSha = regexp.MustCompile(`\b[0-9a-f]{7,40}\b`)

// apaRunSteps builds a fresh fixture with build, runs every step through run
// in the repository's parent directory, and returns each step's result with
// the repository path and every object id normalised to its length: the two
// sides run on two fixtures, whose stash commits differ in their timestamps.
func apaRunSteps(t *testing.T, build func(t *testing.T) string, steps []apaStep, run func(args []string, dir string) apaRes) []apaRes {
	t.Helper()
	repo := build(t)
	var got []apaRes
	for _, s := range steps {
		if s.before != nil {
			s.before(t, repo)
		}
		args := make([]string, len(s.args))
		for i, a := range s.args {
			args[i] = strings.ReplaceAll(a, "$REPO", repo)
		}
		r := run(args, filepath.Dir(repo))
		norm := func(s string) string {
			return apaSha.ReplaceAllStringFunc(strings.ReplaceAll(s, repo, "<repo>"), func(id string) string { return fmt.Sprintf("<sha%d>", len(id)) })
		}
		got = append(got, apaRes{r.code, norm(r.out), norm(r.err)})
	}
	return got
}

func TestAsidePlanningArtifactsParity(t *testing.T) {
	t.Parallel()
	bash := apaBashAt(t)
	edit := func(rel, body string) func(t *testing.T, repo string) {
		return func(t *testing.T, repo string) { writeFile(t, repo+"/"+rel, body) }
	}
	dirtyPlanning := func(t *testing.T, repo string) {
		writeFile(t, repo+"/spectre/changes/demo/tasks.md", "plan edited\n")
		writeFile(t, repo+"/docs/superpowers/new.md", "untracked\n")
		writeFile(t, repo+"/src.txt", "code edited\n")
	}
	both := func(t *testing.T) string { return apaRepo(t, "spectre", "docs") }
	aside, restore := []string{"aside", "$REPO"}, []string{"restore", "$REPO"}
	cases := []struct {
		label string
		build func(t *testing.T) string
		steps []apaStep
	}{
		{"aside then restore", both, []apaStep{{dirtyPlanning, aside, "PLANNING-ARTIFACTS-ASIDE:"}, {nil, restore, "PLANNING-ARTIFACTS-RESTORED:"}}},
		{"nothing to aside", both, []apaStep{{nil, aside, "PLANNING-ARTIFACTS-CLEAN:"}, {nil, restore, "PLANNING-ARTIFACTS-NONE:"}}},
		{"only docs/superpowers is dirty", both, []apaStep{{edit("docs/superpowers/ledger.md", "x\n"), aside, "PLANNING-ARTIFACTS-ASIDE:"}, {nil, restore, "PLANNING-ARTIFACTS-RESTORED:"}}},
		{"both spec roots", func(t *testing.T) string { return apaRepo(t, "spectre", "openspec", "docs") },
			[]apaStep{{edit("spectre/changes/demo/tasks.md", "x\n"), aside, "PLANNING-ARTIFACTS-ASIDE:"}, {nil, restore, "PLANNING-ARTIFACTS-RESTORED:"}, {nil, []string{"bogus", "$REPO"}, ""}}},
		{"openspec only", func(t *testing.T) string { return apaRepo(t, "openspec") },
			[]apaStep{{edit("openspec/changes/demo/tasks.md", "x\n"), aside, "PLANNING-ARTIFACTS-ASIDE:"}, {nil, restore, "PLANNING-ARTIFACTS-RESTORED:"}}},
		{"no planning directory", func(t *testing.T) string { return apaRepo(t) },
			[]apaStep{{nil, aside, "PLANNING-ARTIFACTS-CLEAN:"}, {nil, restore, "PLANNING-ARTIFACTS-NONE:"}, {nil, []string{"bogus", "$REPO"}, ""}}},
		{"a relative worktree is printed as given", both, []apaStep{{dirtyPlanning, []string{"aside", "repo"}, "PLANNING-ARTIFACTS-ASIDE: repo "}, {nil, []string{"restore", "repo"}, "PLANNING-ARTIFACTS-RESTORED: repo "}}},
		{"not a directory", both, []apaStep{{nil, []string{"aside", "$REPO/missing"}, ""}, {nil, []string{"restore", ""}, ""}}},
		{"not a git repository", func(t *testing.T) string { d := t.TempDir() + "/repo"; mkdir(t, d); return d },
			[]apaStep{{nil, aside, ""}, {nil, restore, ""}}},
		{"unknown action", both, []apaStep{{nil, []string{"bogus", "$REPO"}, ""}}},
		{"wrong argument count", both, []apaStep{{nil, nil, ""}, {nil, []string{"aside"}, ""}, {nil, []string{"aside", "$REPO", "x"}, ""}}},
		{"restore refuses mid-merge", both, []apaStep{{dirtyPlanning, aside, "PLANNING-ARTIFACTS-ASIDE:"},
			{func(t *testing.T, repo string) { writeFile(t, repo+"/.git/MERGE_HEAD", "x\n") }, restore, ""}}},
		{"an operator's stash on top is never popped", both, []apaStep{{dirtyPlanning, aside, "PLANNING-ARTIFACTS-ASIDE:"},
			{func(t *testing.T, repo string) {
				writeFile(t, repo+"/src.txt", "operator\n")
				gitRun(t, repo, "stash", "push", "-q", "-m", "mine")
			}, restore, "PLANNING-ARTIFACTS-NONE:"}}},
		{"a conflicted restore keeps the stash", both, []apaStep{{edit("spectre/changes/demo/tasks.md", "aside side\n"), aside, "PLANNING-ARTIFACTS-ASIDE:"},
			{func(t *testing.T, repo string) {
				writeFile(t, repo+"/spectre/changes/demo/tasks.md", "base side\n")
				gitRun(t, repo, "commit", "-qam", "moved")
			}, restore, "PLANNING-ARTIFACTS-CONFLICT:"}}},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			want := apaRunSteps(t, c.build, c.steps, func(args []string, dir string) apaRes {
				cmd := exec.Command("bash", append([]string{bash}, args...)...)
				cmd.Dir = dir
				var out, errb bytes.Buffer
				cmd.Stdout, cmd.Stderr = &out, &errb
				_ = cmd.Run()
				return apaRes{cmd.ProcessState.ExitCode(), out.String(), errb.String()}
			})
			got := apaRunSteps(t, c.build, c.steps, func(args []string, dir string) apaRes {
				fn := Registry["aside-planning-artifacts"]
				if fn == nil {
					t.Fatal("aside-planning-artifacts is not registered")
				}
				var out, errb bytes.Buffer
				code := fn(args, Env{Getenv: os.Getenv, Dir: dir}, &out, &errb)
				return apaRes{code, out.String(), errb.String()}
			})
			for i, s := range c.steps {
				if s.want == "" && want[i].out != "" || !strings.HasPrefix(want[i].out, s.want) {
					t.Errorf("step %d %v: the bash printed %+v, want a %q verdict", i, s.args, want[i], s.want)
				}
				if got[i] != want[i] {
					t.Errorf("step %d %v:\ngot  %+v\nwant %+v", i, c.steps[i].args, got[i], want[i])
				}
			}
		})
	}
}

// TestAsidePlanningArtifactsState asserts what the parity test cannot: the
// tree and stash state each verdict leaves behind — every case of the
// deleted bash harness (KAN-887) beyond stdout, stderr and exit code.
func TestAsidePlanningArtifactsState(t *testing.T) {
	t.Parallel()
	run := func(t *testing.T, args ...string) apaRes {
		t.Helper()
		var out, errb bytes.Buffer
		code := Registry["aside-planning-artifacts"](args, Env{Getenv: os.Getenv}, &out, &errb)
		return apaRes{code, out.String(), errb.String()}
	}
	expect := func(t *testing.T, r apaRes, code int, prefix string) {
		t.Helper()
		if r.code != code || !strings.HasPrefix(r.out, prefix) || prefix == "" && r.out != "" {
			t.Fatalf("got %+v, want exit %d and a %q verdict", r, code, prefix)
		}
	}
	read := func(t *testing.T, path string) string {
		t.Helper()
		b, _ := os.ReadFile(path)
		return string(b)
	}
	status := func(t *testing.T, repo string, paths ...string) string {
		t.Helper()
		out, _ := exec.Command(fixtureGit, append([]string{"-C", repo, "status", "--porcelain", "--"}, paths...)...).Output()
		return string(out)
	}
	stashes := func(t *testing.T, repo string) int {
		t.Helper()
		out, _ := exec.Command(fixtureGit, "-C", repo, "stash", "list").Output()
		return strings.Count(string(out), "\n")
	}
	both := func(t *testing.T) string { return apaRepo(t, "spectre", "docs") }

	t.Run("case 1: aside clears the planning paths and restore brings them back", func(t *testing.T) {
		t.Parallel()
		repo := both(t)
		writeFile(t, repo+"/spectre/changes/demo/tasks.md", "plan-edited\n")
		writeFile(t, repo+"/spectre/changes/demo/notes.md", "scratch\n")
		expect(t, run(t, "aside", repo), 0, "PLANNING-ARTIFACTS-ASIDE:")
		if s := status(t, repo, "spectre/changes", "docs/superpowers"); s != "" {
			t.Fatalf("planning paths still dirty after aside:\n%s", s)
		}
		expect(t, run(t, "restore", repo), 0, "PLANNING-ARTIFACTS-RESTORED:")
		if read(t, repo+"/spectre/changes/demo/tasks.md") != "plan-edited\n" || read(t, repo+"/spectre/changes/demo/notes.md") != "scratch\n" {
			t.Fatal("restore did not bring the tracked edit and the untracked file back")
		}
	})
	t.Run("case 2: docs/superpowers alone is set aside and restored", func(t *testing.T) {
		t.Parallel()
		repo := both(t)
		writeFile(t, repo+"/docs/superpowers/ledger.md", "ledger-edited\n")
		expect(t, run(t, "aside", repo), 0, "PLANNING-ARTIFACTS-ASIDE:")
		if read(t, repo+"/docs/superpowers/ledger.md") != "ledger\n" {
			t.Fatal("aside did not clear docs/superpowers")
		}
		expect(t, run(t, "restore", repo), 0, "PLANNING-ARTIFACTS-RESTORED:")
		if read(t, repo+"/docs/superpowers/ledger.md") != "ledger-edited\n" {
			t.Fatal("restore did not bring the docs/superpowers edit back")
		}
	})
	t.Run("case 3: implementation WIP outside the planning paths is untouched", func(t *testing.T) {
		t.Parallel()
		repo := both(t)
		writeFile(t, repo+"/spectre/changes/demo/tasks.md", "plan-edited\n")
		writeFile(t, repo+"/src.txt", "wip\n")
		expect(t, run(t, "aside", repo), 0, "PLANNING-ARTIFACTS-ASIDE:")
		if read(t, repo+"/src.txt") != "wip\n" || status(t, repo, "src.txt") == "" {
			t.Fatal("src.txt was stashed or modified by aside")
		}
	})
	t.Run("case 4b: an operator stash on top is never popped", func(t *testing.T) {
		t.Parallel()
		repo := both(t)
		writeFile(t, repo+"/src.txt", "operator-wip\n")
		gitRun(t, repo, "stash", "push", "-q", "-m", "operator wip", "--", "src.txt")
		writeFile(t, repo+"/spectre/changes/demo/tasks.md", "plan-edited\n")
		expect(t, run(t, "aside", repo), 0, "PLANNING-ARTIFACTS-ASIDE:")
		writeFile(t, repo+"/src.txt", "operator-later\n")
		gitRun(t, repo, "stash", "push", "-q", "-m", "operator second", "--", "src.txt")
		expect(t, run(t, "restore", repo), 0, "PLANNING-ARTIFACTS-NONE:")
		if n := stashes(t, repo); n != 3 {
			t.Fatalf("%d stashes, want 3", n)
		}
	})
	t.Run("case 5: a conflicted restore exits 1 and keeps the stash", func(t *testing.T) {
		t.Parallel()
		repo := both(t)
		writeFile(t, repo+"/spectre/changes/demo/tasks.md", "line-one-aside\n")
		expect(t, run(t, "aside", repo), 0, "PLANNING-ARTIFACTS-ASIDE:")
		writeFile(t, repo+"/spectre/changes/demo/tasks.md", "line-one-upstream\n")
		gitRun(t, repo, "commit", "-qam", "upstream moves the planning file")
		expect(t, run(t, "restore", repo), 1, "PLANNING-ARTIFACTS-CONFLICT:")
		if n := stashes(t, repo); n != 1 {
			t.Fatalf("%d stashes, want 1", n)
		}
	})
	t.Run("case 6: restore is refused mid-rebase and succeeds once it aborts", func(t *testing.T) {
		t.Parallel()
		repo := both(t)
		gitRun(t, repo, "checkout", "-q", "-b", "upstream")
		writeFile(t, repo+"/src.txt", "upstream-line\n")
		gitRun(t, repo, "commit", "-qam", "upstream rewrites src")
		gitRun(t, repo, "checkout", "-q", "demo")
		writeFile(t, repo+"/src.txt", "demo-line\n")
		gitRun(t, repo, "commit", "-qam", "demo rewrites src")
		writeFile(t, repo+"/spectre/changes/demo/tasks.md", "plan-edited\n")
		expect(t, run(t, "aside", repo), 0, "PLANNING-ARTIFACTS-ASIDE:")
		if exec.Command(fixtureGit, "-C", repo, "rebase", "upstream").Run() == nil {
			t.Fatal("the fixture rebase did not stop on its conflict")
		}
		expect(t, run(t, "restore", repo), 2, "")
		_ = exec.Command(fixtureGit, "-C", repo, "rebase", "--abort").Run()
		expect(t, run(t, "restore", repo), 0, "PLANNING-ARTIFACTS-RESTORED:")
	})
	t.Run("case 8: clean planning paths stash nothing", func(t *testing.T) {
		t.Parallel()
		repo := both(t)
		expect(t, run(t, "aside", repo), 0, "PLANNING-ARTIFACTS-CLEAN:")
		if n := stashes(t, repo); n != 0 {
			t.Fatalf("%d stashes, want 0", n)
		}
	})
	t.Run("case 9: a long stash list is read without a SIGPIPE refusal", func(t *testing.T) {
		t.Parallel()
		repo := both(t)
		head, err := exec.Command(fixtureGit, "-C", repo, "rev-parse", "HEAD").Output()
		if err != nil {
			t.Fatal(err)
		}
		sha := strings.TrimSpace(string(head))
		gitRun(t, repo, "update-ref", "-m", "operator wip 0", "--create-reflog", "refs/stash", "HEAD")
		var log strings.Builder
		log.WriteString(read(t, repo+"/.git/logs/refs/stash"))
		for n := 1; n <= 2000; n++ {
			fmt.Fprintf(&log, "%s %s Test <test@example.invalid> %d +0000\toperator wip %d\n", sha, sha, 1700000000+n, n)
		}
		writeFile(t, repo+"/.git/logs/refs/stash", log.String())
		expect(t, run(t, "restore", repo), 0, "PLANNING-ARTIFACTS-NONE:")
	})
}
