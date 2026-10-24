package guard

import (
	"os"
	"strings"
	"testing"
)

// Every case of scripts/test-check-planning-commit-location.sh at c4f26c84, one subtest
// per run label, each pinning the guard's stdout whole where the harness
// globbed it: 0 with the OK line in a linked worktree on spectre/<name>; 1
// with the main-checkout line in a main checkout (on any branch,
// spectre/<name> included); 1 with the wrong-branch line in a linked
// worktree on another branch or a detached HEAD; both lines together when
// both hold; 2 with nothing on stdout on bad arguments, a missing directory
// or a directory that is not a git work tree.

// pclFx is the harness's sandbox: a main checkout on `main` with linked
// worktrees on spectre/demo, spectre/other and a detached HEAD.
func pclFx(t *testing.T) (sandbox, main string) {
	t.Helper()
	sandbox = t.TempDir()
	main = sandbox + "/repo"
	var g fxGit
	g.git("", "init", "-q", "-b", "main", main)
	g.write(main+"/README.md", "seed")
	g.git(main, "add", "-A")
	g.git(main, "commit", "-q", "-m", "seed")
	g.git(main, "worktree", "add", "-q", "-b", "spectre/demo", main+"/.worktrees/demo")
	g.git(main, "worktree", "add", "-q", "-b", "spectre/other", main+"/.worktrees/other")
	g.git(main, "worktree", "add", "-q", "--detach", main+"/.worktrees/detached")
	if g.err != nil {
		t.Fatal(g.err)
	}
	mkdir(t, sandbox+"/plain")
	mkdir(t, main+"/.worktrees/demo/sub")
	return sandbox, main
}

func TestCheckPlanningCommitLocation(t *testing.T) {
	t.Parallel()
	env := Env{Getenv: os.Getenv}
	sandbox, main := pclFx(t)
	run := func(t *testing.T, rc int, want string, args ...string) {
		t.Helper()
		r := runGuard("check-planning-commit-location", args, env)
		if r.rc != rc || r.stdout != want {
			t.Fatalf("rc=%d stdout=%q stderr=%q; want %d and %q", r.rc, r.stdout, r.err, rc, want)
		}
		if rc == 2 && r.err == "" {
			t.Fatal("cannot answer with nothing on stderr")
		}
	}

	// Every case reads the fixture as pclFx left it; the main checkout moved
	// onto spectre/<name> below gets a fixture of its own.
	t.Run("linked worktree on spectre/<name> passes", func(t *testing.T) {
		t.Parallel()
		run(t, 0, "PLANNING-COMMIT-LOCATION-OK: "+main+"/.worktrees/demo on spectre/demo\n", main+"/.worktrees/demo", "demo")
	})
	// The harness's "main checkout on its default branch is refused" and "…
	// is also the wrong branch": both lines, in this order.
	t.Run("main checkout on its default branch is refused and is also the wrong branch", func(t *testing.T) {
		t.Parallel()
		run(t, 1, "PLANNING-COMMIT-MAIN-CHECKOUT: "+main+"\nPLANNING-COMMIT-WRONG-BRANCH: "+main+" on main — expected spectre/demo\n", main, "demo")
	})
	t.Run("linked worktree of another change is refused", func(t *testing.T) {
		t.Parallel()
		run(t, 1, "PLANNING-COMMIT-WRONG-BRANCH: "+main+"/.worktrees/other on spectre/other — expected spectre/demo\n", main+"/.worktrees/other", "demo")
	})
	t.Run("detached worktree is refused", func(t *testing.T) {
		t.Parallel()
		run(t, 1, "PLANNING-COMMIT-WRONG-BRANCH: "+main+"/.worktrees/detached on detached — expected spectre/demo\n", main+"/.worktrees/detached", "demo")
	})
	t.Run("a subdirectory of the change worktree passes", func(t *testing.T) {
		t.Parallel()
		run(t, 0, "PLANNING-COMMIT-LOCATION-OK: "+main+"/.worktrees/demo/sub on spectre/demo\n", main+"/.worktrees/demo/sub", "demo")
	})
	t.Run("beyond the harness: a relative worktree resolves against the caller's cwd and is printed as given", func(t *testing.T) {
		t.Parallel()
		r := runGuard("check-planning-commit-location", []string{".worktrees/demo", "demo"}, Env{Getenv: os.Getenv, Dir: main})
		if want := "PLANNING-COMMIT-LOCATION-OK: .worktrees/demo on spectre/demo\n"; r.rc != 0 || r.stdout != want {
			t.Fatalf("rc=%d stdout=%q; want 0 and %q", r.rc, r.stdout, want)
		}
	})

	// The main checkout moved onto spectre/<name> is still the main checkout,
	// and the only line is the main-checkout one.
	t.Run("main checkout on spectre/<name> is still refused, with only the main-checkout line", func(t *testing.T) {
		t.Parallel()
		_, main := pclFx(t)
		var g fxGit
		g.git(main+"/.worktrees/demo", "checkout", "-q", "--detach")
		g.git(main, "checkout", "-q", "spectre/demo")
		if g.err != nil {
			t.Fatal(g.err)
		}
		run(t, 1, "PLANNING-COMMIT-MAIN-CHECKOUT: "+main+"\n", main, "demo")
	})

	for _, c := range []struct {
		name string
		args []string
	}{
		{"no arguments cannot answer", nil},
		{"one argument cannot answer", []string{main}},
		{"empty name cannot answer", []string{main, ""}},
		{"missing directory cannot answer", []string{sandbox + "/nope", "demo"}},
		{"non-git directory cannot answer", []string{sandbox + "/plain", "demo"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			run(t, 2, "", c.args...)
		})
	}
	t.Run("beyond the harness: the usage line", func(t *testing.T) {
		t.Parallel()
		r := runGuard("check-planning-commit-location", nil, env)
		if want := "check-planning-commit-location.sh: usage: check-planning-commit-location.sh <worktree> <name>\n"; r.err != want {
			t.Errorf("stderr %q, want %q", r.err, want)
		}
	})
}

// TestCheckPlanningCommitTickedAppend pins kan-1010's self-review: a fix
// run appended Task 3 to an archived plan already ticked, and its planning
// commit went through. A task appended ticked is refused, in an active or an
// archived change folder; ticking a task HEAD carries, appending one
// unticked and a plan HEAD lacks all pass; reshape-branch and commit-split
// still call the location check alone.
func TestCheckPlanningCommitTickedAppend(t *testing.T) {
	t.Parallel()
	env := Env{Getenv: os.Getenv}
	plan := "# demo\n\n" + pclTaskMd("x", "1", "green") + pclTaskMd(" ", "2", "green")
	const refusal = " — appended already ticked; append it unticked, commit, and tick it once its commit passes check-task-commit-fields.sh\n"
	for _, c := range []struct {
		name, dir, head, now string
		rc                   int
		want                 string
	}{
		{"an appended ticked task is refused", "spectre/changes/demo", plan, plan + pclTaskMd("x", "3", "green"), 1,
			"PLANNING-COMMIT-TICKED-APPEND: spectre/changes/demo/tasks.md task 3" + refusal},
		{"an archived plan's ticked append is refused", "spectre/changes/archive/demo", plan, plan + "- [X] 3. Three\n", 1,
			"PLANNING-COMMIT-TICKED-APPEND: spectre/changes/archive/demo/tasks.md task 3" + refusal},
		{"a tick of a carried task passes", "spectre/changes/demo", plan, "# demo\n\n" + pclTaskMd("x", "1", "green") + pclTaskMd("x", "2", "green"), 0, ""},
		{"an unticked append passes", "spectre/changes/demo", plan, plan + pclTaskMd(" ", "3", "green"), 0, ""},
		{"a new plan passes", "spectre/changes/demo", "", plan + pclTaskMd("x", "3", "green"), 0, ""},
		{"another change's ticked append is not this change's", "spectre/changes/other", plan, plan + pclTaskMd("x", "3", "green"), 0, ""},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, main := pclFx(t)
			wt := main + "/.worktrees/demo"
			var g fxGit
			if c.head != "" {
				writeFile(t, wt+"/"+c.dir+"/tasks.md", c.head)
				g.git(wt, "add", "-A")
				g.git(wt, "commit", "-q", "-m", "plan")
			}
			writeFile(t, wt+"/"+c.dir+"/tasks.md", c.now)
			if g.err != nil {
				t.Fatal(g.err)
			}
			r := runGuard("check-planning-commit-location", []string{wt, "demo"}, env)
			want := "PLANNING-COMMIT-LOCATION-OK: " + wt + " on spectre/demo\n" + c.want
			if r.rc != c.rc || r.stdout != want {
				t.Fatalf("rc=%d stdout=%q stderr=%q; want %d and %q", r.rc, r.stdout, r.err, c.rc, want)
			}
		})
	}
	t.Run("the in-process location check never reads the plan", func(t *testing.T) {
		t.Parallel()
		_, main := pclFx(t)
		wt := main + "/.worktrees/demo"
		var g fxGit
		writeFile(t, wt+"/spectre/changes/demo/tasks.md", plan)
		g.git(wt, "add", "-A")
		g.git(wt, "commit", "-q", "-m", "plan")
		writeFile(t, wt+"/spectre/changes/demo/tasks.md", plan+pclTaskMd("x", "3", "green"))
		if g.err != nil {
			t.Fatal(g.err)
		}
		var out, errb strings.Builder
		if rc := checkPlanningCommitLocation([]string{wt, "demo"}, env, &out, &errb); rc != 0 {
			t.Fatalf("rc=%d stdout=%q stderr=%q; want 0", rc, out.String(), errb.String())
		}
	})
}

// pclTaskMd is one task block check-plan-shape and check-task-build-green
// read as well formed, its checkbox tick and its **Build:** value given.
func pclTaskMd(tick, n, build string) string {
	return "- [" + tick + "] " + n + ". Task " + n + "\n\n**Build:** " + build + "\n**Files:** `f" + n + ".go`\n**Tests:** `TestF" + n + "`\n**Commit:** `fix: task " + n + "`\n**After:** none\n\n"
}

// TestCheckPlanningCommitAppendRunsPlanGuards pins kan-1015's self-review:
// tasks 15-24 were appended with **Build:** pending and no declared tests,
// and the planning commit ran neither plan guard. An appended task now runs
// check-plan-shape and check-task-build-green on the working tree's file.
func TestCheckPlanningCommitAppendRunsPlanGuards(t *testing.T) {
	t.Parallel()
	env := Env{Getenv: os.Getenv}
	green := "# demo\n\n" + pclTaskMd(" ", "1", "green")
	pending := "# demo\n\n" + pclTaskMd(" ", "1", "pending")
	for _, c := range []struct {
		name, head, tree, want string
		rc                     int
	}{
		{"an append reading Build pending is refused", green, green + pclTaskMd(" ", "2", "pending"), `task 2 has a **Build:** line reading "pending"`, 1},
		{"a well-formed append passes", green, green + pclTaskMd(" ", "2", "green"), "", 0},
		{"a tick of an existing task alone runs neither guard", pending, "# demo\n\n" + pclTaskMd("x", "1", "pending"), "", 0},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			_, main := pclFx(t)
			wt := main + "/.worktrees/demo"
			var g fxGit
			writeFile(t, wt+"/spectre/changes/demo/tasks.md", c.head)
			g.git(wt, "add", "-A")
			g.git(wt, "commit", "-q", "-m", "plan")
			writeFile(t, wt+"/spectre/changes/demo/tasks.md", c.tree)
			if g.err != nil {
				t.Fatal(g.err)
			}
			r := runGuard("check-planning-commit-location", []string{wt, "demo"}, env)
			if r.rc != c.rc || (c.want != "" && !strings.Contains(r.out, c.want)) {
				t.Fatalf("rc=%d out=%q stderr=%q; want %d carrying %q", r.rc, r.out, r.err, c.rc, c.want)
			}
		})
	}
}
