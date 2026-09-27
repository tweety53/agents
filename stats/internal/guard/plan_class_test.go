package guard

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
)

// Every case of scripts/test-plan-class.sh at c5379c0a, one subtest per ok:
// label, plus pins the harness never asserted: the three output lines byte
// for byte against the bash at c5379c0a for fixed change names (the rolls
// /flow's Decide step records must not move), the refusal messages, and the
// fail-closed binary surface. The bash harness ran the script with stdout
// and stderr merged; here the guard runs in-process, each case building its
// own fixture — a tasks.md at <tmp>/changes/<name>/tasks.md, so the change
// name reads back as the harness's did, and a copy of one template
// repository for the four-argument form.

// pcTaskLines is task_line <n> <files...> for each entry: "- [ ] <n>. Task
// <n>" plus, when it names files, "**Files:** `f1`, `f2`".
func pcTaskLines(tasks ...[]string) string {
	var b strings.Builder
	for i, files := range tasks {
		fmt.Fprintf(&b, "- [ ] %d. Task %d\n", i+1, i+1)
		if len(files) > 0 {
			b.WriteString("**Files:** `" + strings.Join(files, "`, `") + "`\n")
		}
	}
	return b.String()
}

// pcMake is make_tasks: count tasks, each carrying one distinct src/ file,
// task 1 a migration path instead when migration is set.
func pcMake(count int, migration bool) string {
	var tasks [][]string
	for i := 1; i <= count; i++ {
		f := fmt.Sprintf("src/file%d.go", i)
		if i == 1 && migration {
			f = "stats/internal/store/migrations/0099_x.sql"
		}
		tasks = append(tasks, []string{f})
	}
	return pcTaskLines(tasks...)
}

// pcDocs is one_docs_task's plan: one task whose only file is docs/note.md.
var pcDocs = pcTaskLines([]string{"docs/note.md"})

// pcTasks writes body to <tmp>/changes/<name>/tasks.md.
func pcTasks(t *testing.T, name, body string) string {
	t.Helper()
	p := t.TempDir() + "/changes/" + name + "/tasks.md"
	writeFile(t, p, body)
	return p
}

// pcSeq is `seq a b`.
func pcSeq(a, b int) string {
	var s strings.Builder
	for i := a; i <= b; i++ {
		fmt.Fprintf(&s, "%d\n", i)
	}
	return s.String()
}

func TestPlanClass(t *testing.T) {
	t.Parallel()
	// new_git_repo, built once and copied per case: one empty commit, the
	// merge base every four-argument case measures from.
	tmpl := t.TempDir() + "/repo"
	var g fxGit
	g.git("", "init", "-q", tmpl)
	g.git(tmpl, "commit", "-q", "--allow-empty", "-m", "init")
	mb := g.git(tmpl, "rev-parse", "HEAD")
	if g.err != nil {
		t.Fatal(g.err)
	}
	env := Env{Getenv: os.Getenv, LookupEnv: os.LookupEnv, Dir: t.TempDir()}
	run := func(args ...string) guardResult { return runGuard("plan-class", args, env) }

	// repo is a copy of the template; commit stages and commits everything
	// written into it (repo_commit).
	type repoFx struct{ dir string }
	newRepo := func(t *testing.T) repoFx {
		dir := t.TempDir() + "/repo"
		gdcCopyTree(t, tmpl, dir)
		return repoFx{dir}
	}
	commit := func(t *testing.T, r repoFx) {
		var g fxGit
		g.git(r.dir, "add", "-A")
		g.git(r.dir, "commit", "-q", "-m", "surface")
		if g.err != nil {
			t.Fatal(g.err)
		}
	}
	has := func(r guardResult, line string) bool {
		return regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(line) + `$`).MatchString(r.out)
	}
	rolls := func(r guardResult) string {
		return regexp.MustCompile(`(?m)^rolls:.*$`).FindString(r.out)
	}
	classIs := func(t *testing.T, r guardResult, class string) {
		t.Helper()
		if r.rc != 0 || !has(r, "class: "+class) {
			t.Fatalf("rc=%d out=%s", r.rc, r.out)
		}
	}
	exit2 := func(t *testing.T, r guardResult) {
		t.Helper()
		if r.rc != 2 {
			t.Fatalf("rc=%d out=%s", r.rc, r.out)
		}
	}

	for _, tc := range []struct {
		name string
		fn   func(t *testing.T)
	}{
		{"8 tasks / 19 files / repos=1 -> small", func(t *testing.T) {
			var tasks [][]string
			for i := 1; i <= 6; i++ {
				tasks = append(tasks, []string{fmt.Sprintf("f%da.go", i), fmt.Sprintf("f%db.go", i), fmt.Sprintf("f%dc.go", i)})
			}
			tasks = append(tasks, []string{"f7a.go"}, nil)
			r := run(pcTasks(t, "small-8-19", pcTaskLines(tasks...)), "1")
			classIs(t, r, "small")
			if !has(r, "inputs: tasks=8 files=19 repos=1 migration=no spec=no red=no unverified=no") {
				t.Fatalf("out=%s", r.out)
			}
		}},
		{"9 tasks -> regular", func(t *testing.T) {
			classIs(t, run(pcTasks(t, "regular-9", pcMake(9, false)), "1"), "regular")
		}},
		{"22 tasks -> big", func(t *testing.T) {
			classIs(t, run(pcTasks(t, "big-22", pcMake(22, false)), "1"), "big")
		}},
		{"1 migration + 11 tasks -> big", func(t *testing.T) {
			r := run(pcTasks(t, "migration-11-big", pcMake(11, true)), "1")
			classIs(t, r, "big")
			if !strings.Contains(r.out, "migration=yes") {
				t.Fatalf("out=%s", r.out)
			}
		}},
		{"1 migration + 10 tasks -> regular", func(t *testing.T) {
			r := run(pcTasks(t, "migration-10-regular", pcMake(10, true)), "1")
			classIs(t, r, "regular")
			if !strings.Contains(r.out, "migration=yes") {
				t.Fatalf("out=%s", r.out)
			}
		}},
		{"override never appears in plan-class.sh output", func(t *testing.T) {
			r := run(pcTasks(t, "no-override", pcMake(15, false)), "1")
			if r.rc != 0 || strings.Contains(strings.ToLower(r.out), "override") {
				t.Fatalf("rc=%d out=%s", r.rc, r.out)
			}
		}},
		{"rolls are reproducible for a fixed change name", func(t *testing.T) {
			p := pcTasks(t, "roll-stability-name", pcMake(4, false))
			first, second := rolls(run(p, "1")), rolls(run(p, "1"))
			if first == "" || first != second {
				t.Fatalf("first=%s second=%s", first, second)
			}
		}},
		{"rolls line carries compact, experimental and bundle", func(t *testing.T) {
			r := run(pcTasks(t, "bundle-shape", pcMake(4, false)), "1")
			if r.rc != 0 || !regexp.MustCompile(`(?m)^rolls: compact [0-9]+ · experimental [0-9]+ · bundle [0-9]+$`).MatchString(r.out) {
				t.Fatalf("rc=%d out=%s", r.rc, r.out)
			}
		}},
		{"all three rolls are reproducible for a fixed change name", func(t *testing.T) {
			p := pcTasks(t, "roll-stability-name-three", pcMake(4, false))
			first, second := rolls(run(p, "1")), rolls(run(p, "1"))
			if first == "" || first != second {
				t.Fatalf("first=%s second=%s", first, second)
			}
		}},
		// sha256("bundle-static-2bundle") mod 100 = 26.
		{"bundle-static-2 rolls bundle 26 (<30 -> static)", func(t *testing.T) {
			r := run(pcTasks(t, "bundle-static-2", pcMake(4, false)), "1")
			if r.rc != 0 || !strings.HasSuffix(rolls(r), "bundle 26") {
				t.Fatalf("rc=%d out=%s", r.rc, r.out)
			}
		}},
		// sha256("bundle-free-1bundle") mod 100 = 46.
		{"bundle-free-1 rolls bundle 46 (>=30 -> free)", func(t *testing.T) {
			r := run(pcTasks(t, "bundle-free-1", pcMake(4, false)), "1")
			if r.rc != 0 || !strings.HasSuffix(rolls(r), "bundle 46") {
				t.Fatalf("rc=%d out=%s", r.rc, r.out)
			}
		}},
		{"two-argument form on a docs-tiny plan -> small, never micro", func(t *testing.T) {
			classIs(t, run(pcTasks(t, "micro-two-args", pcDocs), "1"), "small")
		}},
		{"one .go task with four args -> small", func(t *testing.T) {
			classIs(t, run(pcTasks(t, "micro-go-plan", pcMake(1, false)), "1", newRepo(t).dir, mb), "small")
		}},
		{"three docs tasks -> small", func(t *testing.T) {
			p := pcTasks(t, "micro-three-docs", pcTaskLines([]string{"a.md"}, []string{"b.md"}, []string{"c.md"}))
			classIs(t, run(p, "1", newRepo(t).dir, mb), "small")
		}},
		{"Build: red tag -> small", func(t *testing.T) {
			classIs(t, run(pcTasks(t, "micro-red", pcDocs+"**Build:** red\n"), "1", newRepo(t).dir, mb), "small")
		}},
		{"one docs task, empty touched surface -> micro", func(t *testing.T) {
			classIs(t, run(pcTasks(t, "micro-empty-surface", pcDocs), "1", newRepo(t).dir, mb), "micro")
		}},
		{"one docs task, 20 changed surface lines -> micro", func(t *testing.T) {
			r := newRepo(t)
			writeFile(t, r.dir+"/note.md", pcSeq(1, 20))
			commit(t, r)
			classIs(t, run(pcTasks(t, "micro-docs-at-cap", pcDocs), "1", r.dir, mb), "micro")
		}},
		{"one docs task, 25 changed surface lines -> small", func(t *testing.T) {
			r := newRepo(t)
			writeFile(t, r.dir+"/note.md", pcSeq(1, 25))
			commit(t, r)
			classIs(t, run(pcTasks(t, "micro-docs-over-cap", pcDocs), "1", r.dir, mb), "small")
		}},
		{"one docs task over a non-docs surface -> small", func(t *testing.T) {
			r := newRepo(t)
			writeFile(t, r.dir+"/src.go", "code\n")
			commit(t, r)
			classIs(t, run(pcTasks(t, "micro-nondoc-surface", pcDocs), "1", r.dir, mb), "small")
		}},
		{"three arguments -> exit 2", func(t *testing.T) {
			exit2(t, run(pcTasks(t, "micro-three-args", pcDocs), "1", newRepo(t).dir))
		}},
		{"non-git worktree argument -> exit 2", func(t *testing.T) {
			exit2(t, run(pcTasks(t, "micro-notgit", pcDocs), "1", t.TempDir(), mb))
		}},
		// F1/F5: the diff side is answered on every four-argument plan, so
		// a bad merge base exits 2 on a .go plan exactly as on a docs one.
		{"unresolving merge base on a .go plan -> exit 2", func(t *testing.T) {
			exit2(t, run(pcTasks(t, "micro-badmb-go", pcMake(1, false)), "1", newRepo(t).dir, "deadbeefdeadbeefdeadbeefdeadbeefdeadbeef"))
		}},
		// F2: a docs file committed and then rewritten unstaged is 20
		// changed lines against the merge base, not 10 + 20 summed.
		{"committed-then-rewritten docs file counts once -> micro", func(t *testing.T) {
			r := newRepo(t)
			writeFile(t, r.dir+"/note.md", pcSeq(1, 10))
			commit(t, r)
			writeFile(t, r.dir+"/note.md", pcSeq(11, 20))
			classIs(t, run(pcTasks(t, "micro-churn-single-count", pcDocs), "1", r.dir, mb), "micro")
		}},
		{"missing file -> exit 2", func(t *testing.T) {
			exit2(t, run("/nonexistent/tasks.md", "1"))
		}},
		{"non-integer repos -> exit 2", func(t *testing.T) {
			exit2(t, run(pcTasks(t, "bad-repos", pcMake(3, false)), "notanumber"))
		}},

		// Pins: each expected value is what the bash printed at c5379c0a for
		// the same fixture.
		{"pin: every input, class and rolls byte for byte against the bash at c5379c0a", func(t *testing.T) {
			p := pcTasks(t, "plan-class-pin", "- [ ] 1. First\n"+
				"**Files:** `docs/a.md`, `spectre/specs/x/spec.md`, `src/b.go`\n"+
				"**Build:** red\n"+
				"- [x] 2. Second\n"+
				"**Files:** `src/b.go`, `stats/internal/store/migrations/0001_x.sql`\n"+
				"unverified: maybe\n")
			r := run(p, "1")
			want := "inputs: tasks=2 files=4 repos=1 migration=yes spec=yes red=yes unverified=yes\n" +
				"class: regular\n" +
				"rolls: compact 71 · experimental 0 · bundle 75\n"
			if r.rc != 0 || r.stdout != want || r.err != "" {
				t.Fatalf("rc=%d stdout=%q stderr=%q", r.rc, r.stdout, r.err)
			}
		}},
		{"pin: micro class and rolls byte for byte against the bash at c5379c0a", func(t *testing.T) {
			r := run(pcTasks(t, "plan-class-micro-pin", "- [ ] 1. Docs\n**Files:** `docs/note.md`\n"), "1", newRepo(t).dir, mb)
			want := "inputs: tasks=1 files=1 repos=1 migration=no spec=no red=no unverified=no\n" +
				"class: micro\n" +
				"rolls: compact 65 · experimental 0 · bundle 43\n"
			if r.rc != 0 || r.stdout != want || r.err != "" {
				t.Fatalf("rc=%d stdout=%q stderr=%q", r.rc, r.stdout, r.err)
			}
		}},
		// Recorded divergence (task 6's Correction): grep read a tasks file
		// with a NUL byte as binary and the bash counted files=0; the port
		// counts the **Files:** union, as the header defines `files`.
		{"port: a tasks file with a NUL byte still counts its Files union", func(t *testing.T) {
			r := run(pcTasks(t, "plan-class-nul", "- [ ] 1. A\n**Files:** `a.md`, `b.md`\nx\x00y\n"), "1", newRepo(t).dir, mb)
			if r.rc != 0 || !strings.HasPrefix(r.stdout, "inputs: tasks=1 files=2 ") {
				t.Fatalf("rc=%d stdout=%q stderr=%q", r.rc, r.stdout, r.err)
			}
		}},
		{"pin: a binary touched path is never micro", func(t *testing.T) {
			r := newRepo(t)
			writeFile(t, r.dir+"/img.md", "a\x00b\n")
			commit(t, r)
			classIs(t, run(pcTasks(t, "micro-binary", pcDocs), "1", r.dir, mb), "small")
		}},
		// grep's \b is the locale's: measured against the bash at c5379c0a
		// under both.
		{"pin: the red tag's word boundary follows the caller's locale", func(t *testing.T) {
			for _, c := range []struct {
				locale, tail, want string
			}{
				{"C", "é", "red=yes"}, {"en_US.UTF-8", "é", "red=no"},
				{"C", "\xe9", "red=yes"}, {"en_US.UTF-8", "\xe9", "red=no"},
				{"en_US.UTF-8", "-ish", "red=yes"}, {"C", "dish", "red=no"}, {"C", "_x", "red=no"}, {"C", "", "red=yes"},
			} {
				p := pcTasks(t, "red-locale", "**Build:** red"+c.tail+"\n")
				e := envWith(t, []string{"LC_ALL=" + c.locale})
				r := runGuard("plan-class", []string{p, "1"}, e)
				if r.rc != 0 || !strings.Contains(r.out, " "+c.want+" ") {
					t.Errorf("%s %q: rc=%d out=%s", c.locale, c.tail, r.rc, r.out)
				}
			}
		}},
		{"pin: refusal messages word for word", func(t *testing.T) {
			repo := newRepo(t).dir
			docs := pcTasks(t, "refusals", pcDocs)
			for _, c := range []struct {
				args []string
				want string
			}{
				{[]string{"a"}, "usage: plan-class.sh <tasks.md> <repos> [<worktree> <merge-base>]\n"},
				{[]string{"/nonexistent/tasks.md", "1"}, "plan-class.sh: no such file: /nonexistent/tasks.md\n"},
				{[]string{docs, "0x1"}, "plan-class.sh: <repos> must be a non-negative integer, got: 0x1\n"},
				{[]string{docs, "", repo, mb}, "plan-class.sh: <repos> must be a non-negative integer, got: \n"},
				{[]string{docs, "1", repo, "deadbeef"}, "plan-class: merge base 'deadbeef' does not resolve in " + repo + "\n"},
			} {
				r := run(c.args...)
				if r.rc != 2 || r.stdout != "" || r.err != c.want {
					t.Errorf("%q: rc=%d stdout=%q stderr=%q", c.args, r.rc, r.stdout, r.err)
				}
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.fn(t)
		})
	}
}
