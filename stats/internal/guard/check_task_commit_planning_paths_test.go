package guard

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of scripts/test-check-task-commit-planning-paths.sh at
// c5379c0a, one subtest per ok: label. Each case builds its own throwaway
// repository with real commits — task commits carrying the pipeline's
// `Task-Id:` trailer, some deliberately sweeping the planning paths — and
// runs the guard in-process against it. The harness grepped for substrings
// on several cases; each case here pins exit code, stdout and stderr whole,
// as the bash printed them at c5379c0a.

// tcppRepo is new_repo: a repository with one commit on main, whose spec
// tree lives under <leaf>/changes/ when a leaf other than spectre is wanted.
func tcppRepo(t *testing.T, leaf string) string {
	t.Helper()
	repo := t.TempDir() + "/repo"
	gitRun(t, "", "init", "-q", "-b", "main", repo)
	if leaf != "spectre" {
		mkdir(t, repo+"/"+leaf+"/changes")
	}
	writeFile(t, repo+"/base.txt", "base\n")
	gitRun(t, repo, "add", "base.txt")
	gitRun(t, repo, "commit", "-qm", "base")
	return repo
}

// tcppCommit is task_commit (tid non-empty) or plain_commit (tid empty):
// one real commit writing path, its trailer naming the task.
func tcppCommit(t *testing.T, repo, msg, tid, path string) {
	t.Helper()
	writeFile(t, repo+"/"+path, path+"\n")
	gitRun(t, repo, "add", "-A")
	if tid == "" {
		gitRun(t, repo, "commit", "-qm", msg)
		return
	}
	gitRun(t, repo, "commit", "-qm", msg, "-m", "Task-Id: "+tid)
}

// tcppSha is `git rev-parse <rev> | cut -c1-12`.
func tcppSha(t *testing.T, repo, rev string) string {
	t.Helper()
	out, err := exec.Command(fixtureGit, "-C", repo, "rev-parse", rev).Output()
	if err != nil {
		t.Fatal(err)
	}
	return string(out[:12])
}

// tcppMerge starts the harness's evil/carry merge: side branched off main
// with one commit, merged back --no-ff --no-commit, left for the caller's
// task commit to conclude.
func tcppMerge(t *testing.T, repo, sideTid, sidePath string) {
	t.Helper()
	gitRun(t, repo, "checkout", "-q", "-b", "side")
	tcppCommit(t, repo, "side work", sideTid, sidePath)
	gitRun(t, repo, "checkout", "-q", "main")
	gitRun(t, repo, "merge", "--no-ff", "--no-commit", "side")
}

func TestCheckTaskCommitPlanningPaths(t *testing.T) {
	t.Parallel()
	const p = "check-task-commit-planning-paths.sh: "
	const usage = "usage: check-task-commit-planning-paths.sh <worktree> <base>\n"
	clean := func(repo, n string) string {
		return "PLANNING-PATHS-CLEAN: " + repo + " — " + n + " task commit(s) checked\n"
	}
	swept := func(repo, sha, tid, path string) string {
		return "TASK-COMMIT-SWEEP: " + sha + " " + tid + " " + path + "\nPLANNING-PATHS-SWEPT: " + repo + " — 1 task commit(s)\n"
	}
	type result struct {
		code           int
		stdout, stderr string
	}
	for _, tc := range []struct {
		name  string
		build func(t *testing.T) ([]string, result)
	}{
		{"no arguments: exit 2, stdout empty, usage on stderr", func(t *testing.T) ([]string, result) {
			return nil, result{2, "", usage}
		}},
		{"one argument: exit 2, stdout empty, usage on stderr", func(t *testing.T) ([]string, result) {
			return []string{tcppRepo(t, "spectre")}, result{2, "", usage}
		}},
		{"non-git directory: exit 2, nothing on stdout", func(t *testing.T) ([]string, result) {
			dir := t.TempDir() + "/not-git"
			mkdir(t, dir)
			return []string{dir, "main"}, result{2, "", p + "not a git repository: " + dir + "\n"}
		}},
		{"unresolvable base: exit 2, nothing on stdout", func(t *testing.T) ([]string, result) {
			return []string{tcppRepo(t, "spectre"), "no-such-ref"}, result{2, "", p + "base does not resolve to a commit: no-such-ref\n"}
		}},
		{"unborn repository: exit 2, nothing on stdout, HEAD named on stderr", func(t *testing.T) ([]string, result) {
			repo := t.TempDir() + "/unborn"
			gitRun(t, "", "init", "-q", "-b", "main", repo)
			return []string{repo, "main"}, result{2, "", p + "HEAD does not resolve in: " + repo + "\n"}
		}},
		{"implementation-only task commit: exit 0, clean verdict names 1 checked", func(t *testing.T) ([]string, result) {
			repo := tcppRepo(t, "spectre")
			base := tcppSha(t, repo, "HEAD")
			tcppCommit(t, repo, "feat(app): do the thing", "1", "src/app.go")
			return []string{repo, base}, result{0, clean(repo, "1"), ""}
		}},
		{"spectre/changes/ sweep: exit 1, commit and path named, swept verdict", func(t *testing.T) ([]string, result) {
			repo := tcppRepo(t, "spectre")
			base := tcppSha(t, repo, "HEAD")
			tcppCommit(t, repo, "feat(app): do the thing", "1", "src/app.go")
			tcppCommit(t, repo, "wip", "2", "spectre/changes/kan-1/tasks.md")
			return []string{repo, base}, result{1, swept(repo, tcppSha(t, repo, "HEAD"), "2", "spectre/changes/kan-1/tasks.md"), ""}
		}},
		{"docs/superpowers/ sweep: exit 1, commit and path named, swept verdict", func(t *testing.T) ([]string, result) {
			repo := tcppRepo(t, "spectre")
			base := tcppSha(t, repo, "HEAD")
			tcppCommit(t, repo, "fix(app): patch", "3", "docs/superpowers/research/notes.md")
			return []string{repo, base}, result{1, swept(repo, tcppSha(t, repo, "HEAD"), "3", "docs/superpowers/research/notes.md"), ""}
		}},
		{"mixed commit: exit 1, only the planning path named", func(t *testing.T) ([]string, result) {
			repo := tcppRepo(t, "spectre")
			base := tcppSha(t, repo, "HEAD")
			writeFile(t, repo+"/src/app.go", "code\n")
			writeFile(t, repo+"/spectre/changes/kan-1/design.md", "plan\n")
			gitRun(t, repo, "add", "-A")
			gitRun(t, repo, "commit", "-qm", "feat(app): do the thing", "-m", "Task-Id: 4")
			return []string{repo, base}, result{1, swept(repo, tcppSha(t, repo, "HEAD"), "4", "spectre/changes/kan-1/design.md"), ""}
		}},
		{"trailer-less commit: exit 0, zero task commits checked", func(t *testing.T) ([]string, result) {
			repo := tcppRepo(t, "spectre")
			base := tcppSha(t, repo, "HEAD")
			tcppCommit(t, repo, "chore(spectre): plan", "", "spectre/changes/kan-1/tasks.md")
			return []string{repo, base}, result{0, clean(repo, "0"), ""}
		}},
		{"spectre/specs/ commit: exit 0, capability specs are implementation", func(t *testing.T) ([]string, result) {
			repo := tcppRepo(t, "spectre")
			base := tcppSha(t, repo, "HEAD")
			tcppCommit(t, repo, "docs(specs): state the rule", "5", "spectre/specs/guards.md")
			return []string{repo, base}, result{0, clean(repo, "1"), ""}
		}},
		{"openspec leaf: exit 1, the project's own changes directory guarded", func(t *testing.T) ([]string, result) {
			repo := tcppRepo(t, "openspec")
			base := tcppSha(t, repo, "HEAD")
			tcppCommit(t, repo, "feat(app): do the thing", "6", "openspec/changes/kan-1/tasks.md")
			return []string{repo, base}, result{1, swept(repo, tcppSha(t, repo, "HEAD"), "6", "openspec/changes/kan-1/tasks.md"), ""}
		}},
		// The shape a task-close boundary always sees mid-run, and the case
		// whose first implementation mis-parsed: the newline git log emits
		// between entries landed at the head of the second sha.
		{"walk: swept commit below a clean one flagged, newer one parsed", func(t *testing.T) ([]string, result) {
			repo := tcppRepo(t, "spectre")
			base := tcppSha(t, repo, "HEAD")
			tcppCommit(t, repo, "wip", "8", "spectre/changes/kan-1/tasks.md")
			tcppCommit(t, repo, "feat(app): do the thing", "9", "src/app.go")
			return []string{repo, base}, result{1, swept(repo, tcppSha(t, repo, "HEAD~1"), "8", "spectre/changes/kan-1/tasks.md"), ""}
		}},
		// KAN-611's depth: the sweep in the branch's first commit, the LAST
		// entry the walk visits.
		{"four-commit branch: deepest sweep flagged, every walk entry observed", func(t *testing.T) ([]string, result) {
			repo := tcppRepo(t, "spectre")
			base := tcppSha(t, repo, "HEAD")
			tcppCommit(t, repo, "wip", "13", "spectre/changes/kan-1/tasks.md")
			tcppCommit(t, repo, "feat(app): second", "14", "src/two.go")
			tcppCommit(t, repo, "feat(app): third", "15", "src/three.go")
			tcppCommit(t, repo, "feat(app): fourth", "16", "src/four.go")
			return []string{repo, base}, result{1, swept(repo, tcppSha(t, repo, "HEAD~3"), "13", "spectre/changes/kan-1/tasks.md"), ""}
		}},
		{"four-commit clean range: 4 checked pinned exactly", func(t *testing.T) ([]string, result) {
			repo := tcppRepo(t, "spectre")
			base := tcppSha(t, repo, "HEAD")
			tcppCommit(t, repo, "feat(app): first", "13", "src/one.go")
			tcppCommit(t, repo, "feat(app): second", "14", "src/two.go")
			tcppCommit(t, repo, "feat(app): third", "15", "src/three.go")
			tcppCommit(t, repo, "feat(app): fourth", "16", "src/four.go")
			return []string{repo, base}, result{0, clean(repo, "4"), ""}
		}},
		// KAN-607: diff-tree without a merge flag prints nothing for merges.
		{"evil merge: exit 1, the merge's smuggled planning path flagged", func(t *testing.T) ([]string, result) {
			repo := tcppRepo(t, "spectre")
			base := tcppSha(t, repo, "HEAD")
			tcppMerge(t, repo, "10", "src/side.go")
			tcppCommit(t, repo, "merge side", "11", "spectre/changes/kan-1/tasks.md")
			if out, _ := exec.Command(fixtureGit, "-C", repo, "cat-file", "-p", "HEAD").Output(); strings.Count(string(out), "\nparent ") != 2 {
				t.Fatal("evil merge: fixture HEAD is not a merge commit")
			}
			return []string{repo, base}, result{1, swept(repo, tcppSha(t, repo, "HEAD"), "11", "spectre/changes/kan-1/tasks.md"), ""}
		}},
		{"carry merge: exit 0, planning content carried over from a parent is the parent's", func(t *testing.T) ([]string, result) {
			repo := tcppRepo(t, "spectre")
			base := tcppSha(t, repo, "HEAD")
			tcppMerge(t, repo, "", "spectre/changes/kan-1/tasks.md")
			tcppCommit(t, repo, "feat(app): do the thing", "12", "src/app.go")
			if out, _ := exec.Command(fixtureGit, "-C", repo, "cat-file", "-p", "HEAD").Output(); strings.Count(string(out), "\nparent ") != 2 {
				t.Fatal("carry merge: fixture HEAD is not a merge commit")
			}
			return []string{repo, base}, result{0, clean(repo, "1"), ""}
		}},
		{"range: swept planning commit before <base> not flagged", func(t *testing.T) ([]string, result) {
			repo := tcppRepo(t, "spectre")
			tcppCommit(t, repo, "chore(spectre): plan", "", "spectre/changes/kan-1/tasks.md")
			before := tcppSha(t, repo, "HEAD")
			tcppCommit(t, repo, "feat(app): do the thing", "7", "src/app.go")
			return []string{repo, before}, result{0, clean(repo, "1"), ""}
		}},
		{"interleaved planning commits: exit 0, two task commits checked", func(t *testing.T) ([]string, result) {
			repo := tcppRepo(t, "spectre")
			base := tcppSha(t, repo, "HEAD")
			tcppCommit(t, repo, "chore(spectre): plan", "", "spectre/changes/kan-1/tasks.md")
			tcppCommit(t, repo, "chore(spectre): link peer", "", "spectre/changes/kan-1/link.md")
			tcppCommit(t, repo, "feat(app): first", "1", "src/one.go")
			tcppCommit(t, repo, "chore(spectre): plan", "", "spectre/changes/kan-1/design.md")
			tcppCommit(t, repo, "feat(app): second", "2", "src/two.go")
			return []string{repo, base}, result{0, clean(repo, "2"), ""}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			args, want := tc.build(t)
			guard := Registry["check-task-commit-planning-paths"]
			if guard == nil {
				t.Fatal("check-task-commit-planning-paths is not registered")
			}
			env := Env{Getenv: os.Getenv, LookupEnv: os.LookupEnv, Dir: filepath.Dir(t.TempDir())}
			var stdout, stderr bytes.Buffer
			got := guard(args, env, &stdout, &stderr)
			if got != want.code || stdout.String() != want.stdout || stderr.String() != want.stderr {
				t.Errorf("exit %d, want %d\nstdout:\n%s\nwant:\n%s\nstderr:\n%s\nwant:\n%s",
					got, want.code, stdout.String(), want.stdout, stderr.String(), want.stderr)
			}
		})
	}
}
