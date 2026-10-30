package guard

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// vvpFix is one visual-verify-pass fixture: a canonical worktree on
// `spectre/demo` whose `.flow/project.md` declares a `## visual verification`
// section, a merge base and one UI commit after it, and a fake `flow` on PATH
// logging each argv to log and answering `record dispatches` from rows.
type vvpFix struct {
	t                *testing.T
	wt, mb, bin, log string
	rows             string
	app              *httptest.Server
	environ          []string
}

const vvpSection = "## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `src/ui/**` |\n| `screenshots` | `tests/visual` |\n\n| Command | Runs |\n|---------|------|\n| `verify` | `npm run test:visual <specs>` |\n| `capture` | `npx playwright test <spec>` |\n| `specs` | `<frontend-root>/specs-since <merge-base>` |\n| `setup` | `npm ci` |\n"

func vvpFixture(t *testing.T, extra string) *vvpFix {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := &vvpFix{t: t, wt: filepath.Join(root, "demo"), bin: filepath.Join(root, "bin"), log: filepath.Join(root, "flow.log")}
	writeFile(t, filepath.Join(f.wt, ".flow/project.md"), "# project\n\n## run\n\n```bash\nnpm run dev\n```\n\n"+vvpSection+extra)
	writeFile(t, filepath.Join(f.wt, "src/ui/app.tsx"), "base\n")
	gitRun(t, f.wt, "init", "-q", "-b", "spectre/demo")
	gitRun(t, f.wt, "add", "-A")
	gitRun(t, f.wt, "commit", "-qm", "base")
	f.mb = vgHead(t, f.wt)
	f.commit("src/ui/app.tsx")
	f.green()
	writeExec(t, filepath.Join(f.bin, "flow"), "#!/bin/sh\necho \"$*\" >> "+f.log+"\ncase \"$1 $2\" in \"record dispatches\") cat "+filepath.Join(root, "rows.json")+";; esac\n")
	f.app = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) }))
	t.Cleanup(f.app.Close)
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	f.environ = append(os.Environ(), "PATH="+f.bin+":"+os.Getenv("PATH"),
		"FLOW_GUARD_REPO_ROOT="+repo,
		"FLOW_GUARD_WORKSPACE_ISOLATION="+filepath.Join(repo, "scripts/check-workspace-isolation.sh"))
	return f
}

// commit commits a change to path in the canonical worktree.
func (f *vvpFix) commit(path string) string { return vgCommit(f.t, f.wt, path) }

// green records a completed inline verify row at the canonical HEAD.
func (f *vvpFix) green() {
	f.rows = "[" + vgRow("verify", vgHead(f.t, f.wt)) + "]"
}

func (f *vvpFix) getenv(k string) string {
	for i := len(f.environ) - 1; i >= 0; i-- {
		if v, ok := strings.CutPrefix(f.environ[i], k+"="); ok {
			return v
		}
	}
	return ""
}

// run runs visual-verify-pass in process with args after the canonical
// worktree, name and key, returning its exit, stdout+stderr and the fake
// flow's logged calls.
func (f *vvpFix) run(sub string, args ...string) (int, string, string) {
	f.t.Helper()
	writeFile(f.t, filepath.Join(filepath.Dir(f.log), "rows.json"), f.rows)
	os.Remove(f.log)
	var out, errb bytes.Buffer
	env := Env{Getenv: f.getenv, Environ: func() []string { return f.environ }, Dir: f.wt}
	rc := Registry["visual-verify-pass"](append([]string{sub}, args...), env, &out, &errb)
	calls, _ := os.ReadFile(f.log)
	return rc, out.String() + errb.String(), string(calls)
}

// prepare is `prepare <wt> demo visual-verify` with the stage's required
// flags and the fixture's app.
func (f *vvpFix) prepare(extra ...string) (int, string, string) {
	f.t.Helper()
	args := []string{f.wt, "demo", "visual-verify", "--harness", "claude", "--session-token", "mf-tok",
		"--merge-base", f.wt + "=" + f.mb, "--app", f.wt + ":.=" + f.app.URL}
	return f.run("prepare", append(args, extra...)...)
}

func (f *vvpFix) stagePath() string {
	return filepath.Join(f.wt, ".superpowers/sdd/visual-verify-stage.txt")
}

func (f *vvpFix) stageFile() string {
	b, _ := os.ReadFile(f.stagePath())
	return string(b)
}

func TestVisualVerifyPassUsage(t *testing.T) {
	f := vvpFixture(t, "")
	base := []string{f.wt, "demo", "visual-verify"}
	ok := []string{"--harness", "claude", "--session-token", "mf-tok", "--merge-base", f.wt + "=" + f.mb}
	for _, c := range []struct {
		name string
		args []string
	}{
		{"no subcommand", nil},
		{"an unknown subcommand", append([]string{"open"}, base...)},
		{"no session token", append(append([]string{"prepare"}, base...), "--harness", "claude", "--merge-base", f.wt+"="+f.mb)},
		{"a token not mf-", append(append([]string{"prepare"}, base...), "--harness", "claude", "--session-token", "tok", "--merge-base", f.wt+"="+f.mb)},
		{"no harness", append(append([]string{"prepare"}, base...), "--session-token", "mf-tok", "--merge-base", f.wt+"="+f.mb)},
		{"a merge base not <wt>=<sha>", append(append([]string{"prepare"}, base...), "--harness", "claude", "--session-token", "mf-tok", "--merge-base", f.mb)},
		{"an app not <wt>:<root>=<url>", append(append(append([]string{"prepare"}, base...), ok...), "--app", ".="+f.app.URL)},
	} {
		os.Remove(f.log)
		var out, errb bytes.Buffer
		env := Env{Getenv: f.getenv, Environ: func() []string { return f.environ }, Dir: f.wt}
		rc := Registry["visual-verify-pass"](c.args, env, &out, &errb)
		calls, _ := os.ReadFile(f.log)
		if rc != 2 || !strings.Contains(errb.String(), "usage: visual-verify-pass.sh") || len(calls) != 0 {
			t.Errorf("%s: exit %d, calls %q\n%s%s", c.name, rc, calls, out.String(), errb.String())
		}
	}
}

func TestVisualVerifyPassPrepareKickoff(t *testing.T) {
	f := vvpFixture(t, "")
	other := filepath.Join(filepath.Dir(f.wt), "regress")
	writeFile(t, filepath.Join(other, "README"), "x\n")
	gitRun(t, other, "init", "-q", "-b", "main")
	gitRun(t, other, "add", "-A")
	gitRun(t, other, "commit", "-qm", "r")
	writeFile(t, filepath.Join(f.wt, ".flow/project.md"), "# project\n\n## apps\n\n| App | Repo root |\n|-----|-----------|\n| web | `"+other+"` |\n\n"+vvpSection+"| `regression checkout` | `"+other+"` |\n| `regression repo` | `x/regress` |\n")
	gitRun(t, f.wt, "commit", "-qam", "cfg")
	f.green()
	rc, out, calls := f.prepare()
	if rc != 1 || !strings.Contains(out, "STOP: kickoff regress") || calls != "" {
		t.Fatalf("exit %d, calls %q\n%s", rc, calls, out)
	}
}

func TestVisualVerifyPassPrepareSkips(t *testing.T) {
	f := vvpFixture(t, "")
	writeFile(t, filepath.Join(f.wt, ".superpowers/sdd/decision.json"), `{"visual":{"verify":"skipped","reason":"copy only"}}`)
	rc, out, calls := f.prepare()
	if rc != 0 || !strings.Contains(out, "NEXT: skip — skipped at Decide — copy only") || !strings.Contains(calls, "stage end") || !strings.Contains(calls, "-outcome skipped") || strings.Contains(calls, "dispatch begin") {
		t.Fatalf("decided skip: exit %d\n%s\n%s", rc, out, calls)
	}

	g := vvpFixture(t, "")
	rc, out, calls = g.prepare("--since", g.wt+"="+vgHead(t, g.wt))
	if rc != 0 || !strings.Contains(out, "NEXT: skip — ") || !strings.Contains(calls, "-outcome completed") || strings.Contains(calls, "dispatch begin") {
		t.Fatalf("no survivor: exit %d\n%s\n%s", rc, out, calls)
	}
}

func TestVisualVerifyPassPrepareComposeOnly(t *testing.T) {
	f := vvpFixture(t, "| `mockups` | `design/mockups` |\n")
	writeFile(t, filepath.Join(f.wt, "design/mockups/f1.png"), "png\n")
	gitRun(t, f.wt, "add", "-A")
	gitRun(t, f.wt, "commit", "-qm", "cfg")
	rb := vgHead(t, f.wt)
	f.commit("design/mockups/f1.png")
	f.green()
	rc, out, calls := f.prepare("--round-base", f.wt+"="+rb)
	if rc != 0 || !strings.Contains(out, "NEXT: compose-only") || strings.Contains(calls, "dispatch begin") {
		t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
	}
}

func TestVisualVerifyPassPrepareStopsOnVerifyNotGreen(t *testing.T) {
	f := vvpFixture(t, "")
	f.commit("src/ui/late.tsx")
	rc, out, calls := f.prepare()
	if rc != 1 || !strings.Contains(out, "STOP: verify-not-green") || !strings.Contains(calls, "stage end") || !strings.Contains(calls, "-outcome stopped") {
		t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
	}
}

func TestVisualVerifyPassPrepareStopsOnPreflight(t *testing.T) {
	f := vvpFixture(t, "| `mockups` | `design/mockups` |\n")
	writeFile(t, filepath.Join(f.wt, "tests/visual/a.spec.ts.mockups"), "gone.png f1\n")
	writeFile(t, filepath.Join(f.wt, "design/mockups/f1.png"), "png\n")
	writeFile(t, filepath.Join(f.wt, "tests/visual/a.spec.ts"), "test('a', () => {})\n")
	gitRun(t, f.wt, "add", "-A")
	gitRun(t, f.wt, "commit", "-qm", "cfg")
	f.green()
	args := []string{f.wt, "demo", "visual-verify", "--harness", "claude", "--session-token", "mf-tok",
		"--merge-base", f.wt + "=" + f.mb, "--app", f.wt + ":missing-app=" + f.app.URL}
	rc, out, calls := f.run("prepare", args...)
	stop := ""
	for _, l := range strings.Split(out, "\n") {
		if strings.HasPrefix(l, "STOP:") {
			stop = l
		}
	}
	if rc != 1 || strings.Count(out, "STOP:") != 1 || !strings.Contains(stop, "check-visual-preflight") || !strings.Contains(stop, "missing-app") || !strings.Contains(stop, "check-sidecar-reach") || !strings.Contains(stop, "gone.png") {
		t.Fatalf("exit %d\n%s", rc, out)
	}
	if !strings.Contains(calls, "stage end") || !strings.Contains(calls, "-outcome stopped") || strings.Contains(calls, "dispatch begin") {
		t.Fatalf("calls:\n%s", calls)
	}
}

func TestVisualVerifyPassPrepareMissStopsForAnalyst(t *testing.T) {
	f := vvpFixture(t, "")
	rc, out, calls := f.prepare("--miss")
	if rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: analyst") || strings.Contains(calls, "dispatch begin") || strings.Contains(calls, "stage end") {
		t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
	}
}

func TestVisualVerifyPassPrepareWritesStageFile(t *testing.T) {
	f := vvpFixture(t, "")
	rb := f.mb
	if rc, out, _ := f.prepare("--miss", "--round-base", f.wt+"="+rb); rc != 0 {
		t.Fatalf("exit %d\n%s", rc, out)
	}
	s := f.stageFile()
	for _, want := range []string{"token mf-tok\n", "begin ", "set 1\n", "worktree " + f.wt + " demo " + f.mb + " " + rb + "\n"} {
		if !strings.Contains(s, want) {
			t.Errorf("stage file lacks %q:\n%s", want, s)
		}
	}
}
