package guard

import (
	"bytes"
	"encoding/json"
	"github.com/tweety53/agents/stats/internal/smoke"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// frFx is one fixture repository: a base commit carrying `.flow/project.md`
// (a `test` list with one input-less command and one reading go/, and a
// `ui paths` of web/src/**) and a tracked base.go, the merge base every case
// starts from. `.superpowers/` is ignored, as in a flow project.
type frFx struct {
	t    *testing.T
	repo string
	mb   string
	log  string
	g    fxGit
	env  Env
}

func newFrFx(t *testing.T) *frFx {
	t.Helper()
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	fx := &frFx{t: t, repo: dir + "/repo", log: dir + "/ran"}
	fx.g.git("", "init", "-q", "-b", "main", fx.repo)
	writeFile(t, fx.repo+"/.flow/project.md", "## test\n\n```bash\necho all >> "+fx.log+"\n# inputs: go/\necho go >> "+fx.log+"\n```\n\n"+
		"## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `web/src/**`, `policy.md` |\n")
	writeFile(t, fx.repo+"/.gitignore", ".superpowers/\n")
	writeFile(t, fx.repo+"/base.go", "base\n")
	fx.g.git(fx.repo, "add", "-A")
	fx.g.git(fx.repo, "commit", "-qm", "base")
	fx.mb = fx.g.git(fx.repo, "rev-parse", "HEAD")
	environ := append(os.Environ(), "FLOW_GUARD_CACHE_DIR="+guardCache(t), "GIT_CONFIG_GLOBAL=/dev/null")
	fx.env = Env{Dir: dir, Environ: func() []string { return environ }, LookupEnv: hermeticGitLookup,
		Getenv: func(k string) string {
			if k == "GIT_CONFIG_GLOBAL" {
				return "/dev/null"
			}
			return os.Getenv(k)
		},
		Findings: func(string) ([]byte, error) { return []byte("[]"), nil }}
	return fx
}

// commit writes each file and commits them together.
func (fx *frFx) commit(files ...string) {
	fx.t.Helper()
	for _, f := range files {
		writeFile(fx.t, fx.repo+"/"+f, f+"\n")
	}
	fx.g.git(fx.repo, "add", "-A")
	fx.g.git(fx.repo, "commit", "-qm", "add "+strings.Join(files, " "))
	if fx.g.err != nil {
		fx.t.Fatal(fx.g.err)
	}
}

// judge writes the judge's JSON and runs the guard's -judge form in round.
func (fx *frFx) judge(round, verdict string) guardResult {
	fx.t.Helper()
	file := filepath.Dir(fx.repo) + "/judge.json"
	writeFile(fx.t, file, verdict)
	return runGuard("fix-run-floor", []string{"-judge", file, "-round", round, fx.repo, fx.mb}, fx.env)
}

const (
	frProse = "class: prose\nfloor_roster: primary\nfloor_verify: fast\n"
	frCode  = "class: code\nfloor_roster: full\nfloor_verify: full\n"
	frUI    = "class: ui\nfloor_roster: full\nfloor_verify: full\n"
	frFast  = `{"roster": ["primary"], "verify": "fast", "reasons": ["one reworded paragraph"]}`
)

// TestFixRunFloorClass pins the class and the floor it carries for each kind
// of touched-path set, each read by the classifier the run already trusts.
func TestFixRunFloorClass(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, tc := range []struct {
		name  string
		files []string
		want  string
	}{
		{"markdown and a docs/ brief are prose", []string{"skills/contract.md", "docs/briefs/x.html", "rules/r.mdc"}, frProse},
		{"a .go file is code", []string{"notes.md", "main.go"}, frCode},
		{"a .ts file is code", []string{"notes.md", "tool/app.ts"}, frCode},
		{"an .html outside docs/ is code", []string{"notes.md", "site/index.html"}, frCode},
		{"a ui path is ui", []string{"notes.md", "web/src/App.tsx"}, frUI},
		{"documentation named in `ui paths` is ui", []string{"policy.md"}, frUI},
		{"documentation under a suite's `# inputs:` is code", []string{"go/README.md"}, frCode},
		{"nothing touched is code", nil, frCode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fx := newFrFx(t)
			if tc.files != nil {
				fx.commit(tc.files...)
			}
			r := runGuard("fix-run-floor", []string{fx.repo, fx.mb}, fx.env)
			if r.rc != 0 || r.stdout != tc.want {
				t.Fatalf("rc=%d stdout=%q stderr=%q, want rc=0 stdout=%q", r.rc, r.stdout, r.err, tc.want)
			}
		})
	}
	t.Run("staged and unstaged paths count", func(t *testing.T) {
		t.Parallel()
		fx := newFrFx(t)
		fx.commit("notes.md")
		fx.g.appendLine(fx.repo+"/base.go", "unstaged edit")
		if r := runGuard("fix-run-floor", []string{fx.repo, fx.mb}, fx.env); r.rc != 0 || r.stdout != frCode {
			t.Fatalf("rc=%d stdout=%q stderr=%q", r.rc, r.stdout, r.err)
		}
	})
	t.Run("cannot answer", func(t *testing.T) {
		t.Parallel()
		fx := newFrFx(t)
		for _, args := range [][]string{{}, {fx.repo}, {fx.repo, "no-such-ref"}, {fx.repo + "/absent", fx.mb},
			{"-judge", "x.json", fx.repo, fx.mb}, {"-round", "1", fx.repo, fx.mb}, {"-judge", "x.json", "-round", "one", fx.repo, fx.mb},
			{"-judge", fx.repo + "/absent.json", "-round", "1", fx.repo, fx.mb}} {
			if r := runGuard("fix-run-floor", args, fx.env); r.rc != 2 || r.err == "" {
				t.Errorf("%q: rc=%d stderr=%q, want rc=2 and a cause", args, r.rc, r.err)
			}
		}
	})
}

// TestFixRunFloorVerdict pins the gate: a verdict is admitted only at or above
// the floor, written with its round and base, and removed by the next call.
// smoke: core — the one gate that lets a run skip its tests
func TestFixRunFloorVerdict(t *testing.T) {
	t.Parallel()
	t.Run("an admitted verdict is written and the next plain call removes it", func(t *testing.T) {
		t.Parallel()
		fx := newFrFx(t)
		fx.commit("notes.md")
		r := fx.judge("3", frFast)
		if r.rc != 0 || r.stdout != frProse+"verdict: roster primary · verify fast\n" {
			t.Fatalf("rc=%d stdout=%q stderr=%q", r.rc, r.stdout, r.err)
		}
		var v frVerdict
		raw, err := os.ReadFile(frPath(fx.repo))
		if err != nil || json.Unmarshal(raw, &v) != nil {
			t.Fatalf("fix-run.json: %v\n%s", err, raw)
		}
		if v.Round != 3 || v.Base != fx.mb || v.Class != "prose" || strings.Join(v.Roster, ",") != "primary" || v.Verify != "fast" || len(v.Reasons) != 1 {
			t.Errorf("fix-run.json = %+v", v)
		}
		if got, ok := fixRunVerdict(fx.env, fx.repo); !ok || got.Round != 3 {
			t.Errorf("fixRunVerdict = %+v, %v: want the verdict in force", got, ok)
		}
		if r := runGuard("fix-run-floor", []string{fx.repo, fx.mb}, fx.env); r.rc != 0 {
			t.Fatalf("plain call: rc=%d stderr=%q", r.rc, r.err)
		}
		if _, err := os.Stat(frPath(fx.repo)); !os.IsNotExist(err) {
			t.Errorf("an earlier fix run's verdict survived the next floor call: %v", err)
		}
	})
	for name, verdict := range map[string]string{
		"a roster without primary":  `{"roster": ["principles"], "verify": "full", "reasons": ["r"]}`,
		"an empty roster":           `{"roster": [], "verify": "fast", "reasons": ["r"]}`,
		"a roster entry off the id": `{"roster": ["primary", "all of them"], "verify": "fast", "reasons": ["r"]}`,
		"an unknown verify mode":    `{"roster": ["primary"], "verify": "none", "reasons": ["r"]}`,
		"no reason":                 `{"roster": ["primary"], "verify": "fast", "reasons": []}`,
		"an unknown field":          `{"roster": ["primary"], "verify": "fast", "reasons": ["r"], "skip_lint": true}`,
		"not JSON":                  `roster: primary`,
	} {
		t.Run("refused: "+name, func(t *testing.T) {
			t.Parallel()
			fx := newFrFx(t)
			fx.commit("notes.md")
			if r := fx.judge("1", frFast); r.rc != 0 {
				t.Fatalf("seeding an earlier verdict: rc=%d stderr=%q", r.rc, r.err)
			}
			r := fx.judge("2", verdict)
			if r.rc != 1 || r.err == "" {
				t.Fatalf("rc=%d stderr=%q, want rc=1 and a cause", r.rc, r.err)
			}
			if _, err := os.Stat(frPath(fx.repo)); !os.IsNotExist(err) {
				t.Errorf("a refused verdict left a file in force: %v", err)
			}
		})
	}
	t.Run("refused: code has no room above its floor", func(t *testing.T) {
		t.Parallel()
		fx := newFrFx(t)
		fx.commit("notes.md", "main.go")
		if r := fx.judge("1", frFast); r.rc != 1 || r.stdout != frCode || !strings.Contains(r.err, "class code") {
			t.Fatalf("rc=%d stdout=%q stderr=%q", r.rc, r.stdout, r.err)
		}
		if _, err := os.Stat(frPath(fx.repo)); !os.IsNotExist(err) {
			t.Errorf("a verdict was written under a full floor: %v", err)
		}
	})
	t.Run("a verdict stops governing once the change is no longer prose", func(t *testing.T) {
		t.Parallel()
		fx := newFrFx(t)
		fx.commit("notes.md")
		if r := fx.judge("1", frFast); r.rc != 0 {
			t.Fatalf("rc=%d stderr=%q", r.rc, r.err)
		}
		fx.commit("main.go")
		if _, ok := fixRunVerdict(fx.env, fx.repo); ok {
			t.Error("fixRunVerdict held a prose verdict over a tree carrying main.go")
		}
	})
}

// TestPanelRoundBeginsEnforcesTheJudgedRoster: in the verdict's round a slot
// carrying a role the roster omits is refused before any begin; another
// round, or a tree that is no longer prose, is not restricted.
// smoke: core — the only test that a judged roster is enforced
func TestPanelRoundBeginsEnforcesTheJudgedRoster(t *testing.T) {
	fx := newFrFx(t)
	fx.commit("notes.md")
	if r := fx.judge("4", frFast); r.rc != 0 {
		t.Fatalf("rc=%d stderr=%q", r.rc, r.err)
	}
	bin := t.TempDir()
	calls := filepath.Join(bin, "calls")
	writeExec(t, filepath.Join(bin, "flow"), "#!/bin/sh\necho \"$*\" >> "+calls+"\necho recorded\n")
	env := fx.env
	getenv := env.Getenv
	env.Getenv = func(k string) string {
		if k == "PATH" {
			return bin + ":" + os.Getenv("PATH")
		}
		return getenv(k)
	}
	begins := func(round string, slots ...string) (int, string, string) {
		os.Remove(calls)
		var out, errb bytes.Buffer
		rc := panelRoundBegins(append([]string{fx.repo, "kan-1", round, "mf-k1"}, slots...), env, &out, &errb)
		b, _ := os.ReadFile(calls)
		return rc, string(b), errb.String()
	}
	if rc, made, stderr := begins("4", "primary:opus:low", "test-sanity+call-efficiency:sonnet:low"); rc != 3 || made != "" ||
		!strings.Contains(stderr, "slot test-sanity+call-efficiency carries test-sanity") {
		t.Errorf("a slot off the judged roster: rc=%d calls=%q stderr=%q, want rc=3 and no begin", rc, made, stderr)
	}
	if rc, made, _ := begins("4", "primary+principles:opus:low"); rc != 3 || made != "" {
		t.Errorf("a bundle carrying one omitted role: rc=%d calls=%q, want rc=3 and no begin", rc, made)
	}
	if rc, made, stderr := begins("4", "primary:opus:low"); rc != 0 || strings.Count(made, "-key panel-4-primary ") != 1 {
		t.Errorf("the judged roster itself: rc=%d calls=%q stderr=%q", rc, made, stderr)
	}
	if rc, made, stderr := begins("5", "mutation+failure-modes:opus:low"); rc != 0 || made == "" {
		t.Errorf("a later round is not the verdict's: rc=%d calls=%q stderr=%q", rc, made, stderr)
	}
	fx.commit("main.go")
	if rc, made, stderr := begins("4", "primary+principles:opus:low"); rc != 0 || made == "" {
		t.Errorf("a tree no longer prose: rc=%d calls=%q stderr=%q, want the verdict out of force", rc, made, stderr)
	}
}

// TestRunVerifyListVerifyFast: under a judged verify fast the `test` list
// runs no command, ends its row skipped and stamps no green; a Critical or
// Important finding of the verdict's round or later, or a tree that is no
// longer prose, runs the whole list.
// smoke: core — the only test that verify fast is enforced
func TestRunVerifyListVerifyFast(t *testing.T) {
	fx := newFrFx(t)
	fx.commit("notes.md")
	if r := fx.judge("2", frFast); r.rc != 0 {
		t.Fatalf("rc=%d stderr=%q", r.rc, r.err)
	}
	findings := "[]"
	env := fx.env
	env.Findings = func(string) ([]byte, error) { return []byte(findings), nil }
	run := func() (int, string, string) {
		os.Remove(fx.log)
		var out, errb bytes.Buffer
		rc := runVerifyList([]string{fx.repo, "test"}, env, &out, &errb)
		ran, _ := os.ReadFile(fx.log)
		return rc, out.String(), string(ran)
	}
	green := fx.g.git(fx.repo, "rev-parse", "--path-format=absolute", "--git-path", "flow-test-green")

	rc, out, ran := run()
	if rc != 0 || ran != "" || strings.Count(out, "— skipped, fix-run verify fast: the change is prose\n") != 2 {
		t.Fatalf("verify fast: rc=%d ran=%q out:\n%s", rc, ran, out)
	}
	if _, err := os.Stat(green); !os.IsNotExist(err) {
		t.Errorf("a fast run stamped a green sha no suite earned: %v", err)
	}

	findings = `[{"ref":"F1","round":1,"severity":"Critical","status":"fixed","slot":"primary","note":"n"}]`
	if rc, out, ran := run(); rc != 0 || ran != "" {
		t.Errorf("a severe finding of an earlier round voided fast: rc=%d ran=%q out:\n%s", rc, ran, out)
	}
	findings = `[{"ref":"F2","round":2,"severity":"minor","status":"fixed","slot":"primary","note":"n"}]`
	if rc, out, ran := run(); rc != 0 || ran != "" {
		t.Errorf("a Minor voided fast: rc=%d ran=%q out:\n%s", rc, ran, out)
	}
	findings = `[{"ref":"F3","round":2,"severity":"Important","status":"fixed","slot":"primary","note":"n"}]`
	if rc, out, ran := run(); rc != 0 || ran != "all\ngo\n" && ran != "go\nall\n" || !strings.Contains(out, "fix-run verify fast voided — Important finding F3 of round 2") {
		t.Errorf("an Important of the verdict's round: rc=%d ran=%q out:\n%s", rc, ran, out)
	}
	// The full run above stamped its green; drop it so the next full run runs.
	dropGreen := func() {
		os.Remove(green)
		os.Remove(green + "-cmds")
	}
	dropGreen()
	env.Findings = func(string) ([]byte, error) { return nil, os.ErrDeadlineExceeded }
	if rc, _, ran := run(); rc != 0 || len(ran) == 0 {
		t.Errorf("unreadable findings kept fast: rc=%d ran=%q", rc, ran)
	}

	env.Findings = func(string) ([]byte, error) { return []byte("[]"), nil }
	dropGreen()
	fx.commit("main.go")
	if rc, out, ran := run(); rc != 0 || !strings.Contains(ran, "all\n") || strings.Contains(out, "verify fast") {
		t.Errorf("a tree no longer prose: rc=%d ran=%q out:\n%s", rc, ran, out)
	}
}
