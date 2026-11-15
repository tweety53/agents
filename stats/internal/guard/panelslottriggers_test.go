package guard

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/smoke"
)

// pstFx is one copy of the panel-slot-triggers template repository: main,
// one commit carrying README.md, stats/store/q.go and cmd/db.go (one
// db.QueryRow( line each), cmd/a.go and cmd/a_test.go -- the merge base every case diffs against.
type pstFx struct {
	dir, repo string
	g         fxGit
	env       Env
}

// pstTemplate builds the template once per test and returns a constructor
// for an isolated copy of it, plus its merge base.
func pstTemplate(t *testing.T) (func(t *testing.T) *pstFx, string) {
	t.Helper()
	tmpl := t.TempDir() + "/repo"
	var g fxGit
	g.git("", "init", "-q", "-b", "main", tmpl)
	for p, body := range map[string]string{
		"README.md":        "# readme",
		"stats/store/q.go": "row := db.QueryRow(q)",
		"cmd/db.go":        "row := db.QueryRow(q)",
		"cmd/a.go":         "package cmd",
		"cmd/a_test.go":    "package cmd",
	} {
		if err := os.MkdirAll(filepath.Dir(tmpl+"/"+p), 0o755); err != nil {
			t.Fatal(err)
		}
		g.write(tmpl+"/"+p, body)
	}
	g.git(tmpl, "add", ".")
	g.git(tmpl, "commit", "-qm", "base")
	mb := g.git(tmpl, "rev-parse", "HEAD")
	if g.err != nil {
		t.Fatal(g.err)
	}
	return func(t *testing.T) *pstFx {
		fx := &pstFx{dir: t.TempDir()}
		fx.repo = fx.dir + "/repo"
		gdcCopyTree(t, tmpl, fx.repo)
		fx.env = Env{Dir: fx.dir, Getenv: os.Getenv, LookupEnv: hermeticGitLookup}
		return fx
	}, mb
}

// put writes body to the repository path p, creating its directory.
func (fx *pstFx) put(t *testing.T, p, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(fx.repo+"/"+p), 0o755); err != nil {
		t.Fatal(err)
	}
	fx.g.write(fx.repo+"/"+p, body)
}

// commit stages everything and commits it.
func (fx *pstFx) commit() {
	fx.g.git(fx.repo, "add", "-A")
	fx.g.git(fx.repo, "commit", "-qm", "change")
}

func pstRun(t *testing.T, env Env, args ...string) guardResult {
	t.Helper()
	return runGuard("panel-slot-triggers", args, env)
}

// pstWant is the three lines, each slot's argument its verdict and reason,
// "" for a skip with the no-signal reason.
func pstWant(ts, ce, fm string) string {
	line := func(slot, v string) string {
		if v == "" {
			return slot + ": skip — no " + slot + " signal in the changed code"
		}
		return slot + ": run — " + v
	}
	return line("test-sanity", ts) + "\n" + line("call-efficiency", ce) + "\n" + line("failure-modes", fm) + "\n"
}

func pstIs(t *testing.T, r guardResult, want string) {
	t.Helper()
	if r.rc != 0 || r.stdout != want {
		t.Fatalf("want exit 0 stdout\n%s\ngot exit %d stdout\n%s\nstderr %q", want, r.rc, r.stdout, r.err)
	}
}

// A touched test path -- added, changed or deleted -- runs test-sanity; a
// docs-only range skips all three.
func TestPanelSlotTriggersTestSanity(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	newFx, mb := pstTemplate(t)
	for _, c := range []struct {
		name   string
		change func(t *testing.T, fx *pstFx)
		want   string
	}{
		{"added go test", func(t *testing.T, fx *pstFx) { fx.put(t, "cmd/b_test.go", "package cmd") },
			pstWant("cmd/b_test.go go-test", "", "")},
		{"added ts test", func(t *testing.T, fx *pstFx) { fx.put(t, "web/x.test.ts", "it()") },
			pstWant("web/x.test.ts js-test", "", "")},
		{"deleted go test", func(t *testing.T, fx *pstFx) { fx.g.err = os.Remove(fx.repo + "/cmd/a_test.go") },
			pstWant("cmd/a_test.go go-test", "", "")},
		{"docs only", func(t *testing.T, fx *pstFx) { fx.put(t, "README.md", "# changed") },
			pstWant("", "", "")},
	} {
		t.Run(c.name, func(t *testing.T) {
			fx := newFx(t)
			c.change(t, fx)
			fx.commit()
			if fx.g.err != nil {
				t.Fatal(fx.g.err)
			}
			pstIs(t, pstRun(t, fx.env, fx.repo, mb), c.want)
		})
	}
}

// A changed handler, query or client line -- a deleted one too -- and a
// changed path under a handler, store or client directory run
// call-efficiency and failure-modes; the same token in a doc, a lockfile or a
// test path runs neither.
func TestPanelSlotTriggersCallEfficiency(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	newFx, mb := pstTemplate(t)
	both := func(r string) string { return pstWant("", r, r) }
	for _, c := range []struct {
		name   string
		change func(t *testing.T, fx *pstFx)
		want   string
	}{
		{"added client call", func(t *testing.T, fx *pstFx) {
			fx.g.appendLine(fx.repo+"/cmd/a.go", `req, _ := http.NewRequest("GET", u, nil)`)
		}, both("cmd/a.go:2 http-client")},
		{"deleted query", func(t *testing.T, fx *pstFx) { fx.put(t, "stats/store/q.go", "package store") },
			both("stats/store/q.go:1 query")},
		{"deleted file outside handler dirs", func(t *testing.T, fx *pstFx) { fx.g.err = os.Remove(fx.repo + "/cmd/db.go") },
			both("cmd/db.go:1 query")},
		{"new file under api", func(t *testing.T, fx *pstFx) { fx.put(t, "internal/api/h.go", "package api") },
			both("internal/api/h.go handler-dir")},
		{"http-handler line", func(t *testing.T, fx *pstFx) {
			fx.g.appendLine(fx.repo+"/cmd/a.go", `func (h H) ServeHTTP(w W, r *http.Request) {}`)
		},
			both("cmd/a.go:2 http-handler")},
		{"spring-handler line", func(t *testing.T, fx *pstFx) { fx.g.appendLine(fx.repo+"/cmd/a.go", `@GetMapping("/x")`) },
			both("cmd/a.go:2 spring-handler")},
		{"next-handler line", func(t *testing.T, fx *pstFx) { fx.g.appendLine(fx.repo+"/cmd/a.go", `export async function GET() {}`) },
			both("cmd/a.go:2 next-handler")},
		{"sql line", func(t *testing.T, fx *pstFx) { fx.g.appendLine(fx.repo+"/cmd/a.go", `q := "SELECT 1"`) },
			both("cmd/a.go:2 sql")},
		{"jpa-query line", func(t *testing.T, fx *pstFx) { fx.g.appendLine(fx.repo+"/cmd/a.go", `@Query("x")`) },
			both("cmd/a.go:2 jpa-query")},
		{"fetch line", func(t *testing.T, fx *pstFx) { fx.g.appendLine(fx.repo+"/cmd/a.go", `fetch(u)`) },
			both("cmd/a.go:2 fetch")},
		{"axios line", func(t *testing.T, fx *pstFx) { fx.g.appendLine(fx.repo+"/cmd/a.go", `axios.get(u)`) },
			both("cmd/a.go:2 axios")},
		{"http-client-jvm line", func(t *testing.T, fx *pstFx) { fx.g.appendLine(fx.repo+"/cmd/a.go", `val c = HttpClient()`) },
			both("cmd/a.go:2 http-client-jvm")},
		{"react-query line", func(t *testing.T, fx *pstFx) { fx.g.appendLine(fx.repo+"/cmd/a.go", `useQuery(k)`) },
			both("cmd/a.go:2 react-query")},
		{"sql file", func(t *testing.T, fx *pstFx) { fx.put(t, "db/x.sql", "-- x") },
			both("db/x.sql sql-file")},
		{"fetch in a doc", func(t *testing.T, fx *pstFx) { fx.g.appendLine(fx.repo+"/README.md", "fetch(x)") },
			pstWant("", "", "")},
		{"fetch in a lockfile", func(t *testing.T, fx *pstFx) { fx.put(t, "go.sum", "fetch(x)") },
			pstWant("", "", "")},
		{"fetch in a .lockfile", func(t *testing.T, fx *pstFx) { fx.put(t, "deps.lockfile", "fetch(x)") },
			pstWant("", "", "")},
		{"fetch in a test", func(t *testing.T, fx *pstFx) { fx.put(t, "web/x.test.ts", "fetch(x)") },
			pstWant("web/x.test.ts js-test", "", "")},
	} {
		t.Run(c.name, func(t *testing.T) {
			fx := newFx(t)
			c.change(t, fx)
			fx.commit()
			if fx.g.err != nil {
				t.Fatal(fx.g.err)
			}
			pstIs(t, pstRun(t, fx.env, fx.repo, mb), c.want)
		})
	}
}

// File and exec calls, concurrency and a scheduler-style directory run
// failure-modes alone.
func TestPanelSlotTriggersFailureModes(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	newFx, mb := pstTemplate(t)
	for _, c := range []struct {
		name   string
		change func(t *testing.T, fx *pstFx)
		want   string
	}{
		{"file write", func(t *testing.T, fx *pstFx) {
			fx.g.appendLine(fx.repo+"/cmd/a.go", "_ = os.WriteFile(p, b, 0o644)")
		}, pstWant("", "", "cmd/a.go:2 file-io")},
		{"goroutine", func(t *testing.T, fx *pstFx) { fx.g.appendLine(fx.repo+"/cmd/a.go", "go func() {}()") },
			pstWant("", "", "cmd/a.go:2 goroutine")},
		{"exec line", func(t *testing.T, fx *pstFx) { fx.g.appendLine(fx.repo+"/cmd/a.go", `exec.Command("ls")`) },
			pstWant("", "", "cmd/a.go:2 exec")},
		{"sync line", func(t *testing.T, fx *pstFx) { fx.g.appendLine(fx.repo+"/cmd/a.go", `var mu sync.Mutex`) },
			pstWant("", "", "cmd/a.go:2 sync")},
		{"channel line", func(t *testing.T, fx *pstFx) { fx.g.appendLine(fx.repo+"/cmd/a.go", `c := make(chan int)`) },
			pstWant("", "", "cmd/a.go:2 channel")},
		{"context line", func(t *testing.T, fx *pstFx) {
			fx.g.appendLine(fx.repo+"/cmd/a.go", `ctx, cancel := context.WithCancel(ctx)`)
		},
			pstWant("", "", "cmd/a.go:2 context")},
		{"kotlin-coroutine line", func(t *testing.T, fx *pstFx) { fx.g.appendLine(fx.repo+"/cmd/a.go", `launch { }`) },
			pstWant("", "", "cmd/a.go:2 kotlin-coroutine")},
		{"ts-async line", func(t *testing.T, fx *pstFx) { fx.g.appendLine(fx.repo+"/cmd/a.go", `await x()`) },
			pstWant("", "", "cmd/a.go:2 ts-async")},
		{"space in the name", func(t *testing.T, fx *pstFx) { fx.put(t, "cmd/my file.go", "go func() {}()") },
			pstWant("", "", "cmd/my file.go:1 goroutine")},
		{"non-ASCII name", func(t *testing.T, fx *pstFx) { fx.put(t, "cmd/café.go", "go func() {}()") },
			pstWant("", "", "cmd/café.go:1 goroutine")},
		{"jobs directory", func(t *testing.T, fx *pstFx) { fx.put(t, "jobs/run.go", "package jobs") },
			pstWant("", "", "jobs/run.go worker-dir")},
	} {
		t.Run(c.name, func(t *testing.T) {
			fx := newFx(t)
			c.change(t, fx)
			fx.commit()
			if fx.g.err != nil {
				t.Fatal(fx.g.err)
			}
			pstIs(t, pstRun(t, fx.env, fx.repo, mb), c.want)
		})
	}
}

// git C-quotes a name holding `"` or `\` whatever core.quotePath says; its
// changed lines are still read under the decoded name.
func TestPanelSlotTriggersQuotedNames(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	newFx, mb := pstTemplate(t)
	for name, p := range map[string]string{
		"double quote in the name": `cmd/we"ird.go`,
		"backslash in the name":    `cmd/back\slash.go`,
	} {
		t.Run(name, func(t *testing.T) {
			fx := newFx(t)
			fx.put(t, p, "go func() {}()")
			fx.commit()
			if fx.g.err != nil {
				t.Fatal(fx.g.err)
			}
			pstIs(t, pstRun(t, fx.env, fx.repo, mb), pstWant("", "", p+":1 goroutine"))
		})
	}
}

// One worktree's run wins over another's skip, in either order, and the
// first pair's reason wins when both run; a path the worktree's `## panel
// diff exclude` leaves out is never scanned. The code worktree's change is
// unstaged.
func TestPanelSlotTriggersAcrossWorktrees(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	newFx, mb := pstTemplate(t)
	docs, code := newFx(t), newFx(t)
	docs.put(t, "README.md", "# changed")
	docs.commit()
	code.g.appendLine(code.repo+"/cmd/a.go", "go func() {}()")
	for _, fx := range []*pstFx{docs, code} {
		if fx.g.err != nil {
			t.Fatal(fx.g.err)
		}
	}
	want := pstWant("", "", "cmd/a.go:2 goroutine")
	pstIs(t, pstRun(t, docs.env, docs.repo, mb, code.repo, mb), want)
	pstIs(t, pstRun(t, docs.env, code.repo, mb, docs.repo, mb), want)

	jobs := newFx(t)
	jobs.put(t, "jobs/run.go", "package jobs")
	jobs.commit()
	if jobs.g.err != nil {
		t.Fatal(jobs.g.err)
	}
	pstIs(t, pstRun(t, jobs.env, jobs.repo, mb, code.repo, mb), pstWant("", "", "jobs/run.go worker-dir"))

	ex := newFx(t)
	ex.put(t, ".flow/project.md", "# p\n\n## panel diff exclude\n\n```text\ngen/**\n```")
	ex.put(t, "gen/z.go", "go func() {}()")
	ex.commit()
	if ex.g.err != nil {
		t.Fatal(ex.g.err)
	}
	pstIs(t, pstRun(t, ex.env, ex.repo, mb), pstWant("", "", ""))
}

// A question the guard cannot answer exits 2 with nothing on stdout.
func TestPanelSlotTriggersCannotAnswer(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	newFx, mb := pstTemplate(t)
	fx := newFx(t)
	for _, c := range []struct {
		name string
		args []string
	}{
		{"no arguments", nil},
		{"odd count", []string{fx.repo, mb, fx.repo}},
		{"not a directory", []string{fx.dir + "/missing", mb}},
		{"not a repository", []string{t.TempDir(), mb}},
		{"merge base does not resolve", []string{fx.repo, strings.Repeat("0", 40)}},
		{"second pair refused", []string{fx.repo, mb, fx.dir + "/missing", mb}},
	} {
		t.Run(c.name, func(t *testing.T) {
			r := pstRun(t, fx.env, c.args...)
			if r.rc != 2 || r.stdout != "" || r.err == "" {
				t.Fatalf("want exit 2, empty stdout, a message on stderr: exit %d stdout %q stderr %q", r.rc, r.stdout, r.err)
			}
		})
	}
}
