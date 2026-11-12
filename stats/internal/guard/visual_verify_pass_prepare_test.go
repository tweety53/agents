package guard

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
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
		"FLOW_GUARD_REPO_ROOT="+repo)
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
		hint string
	}{
		{"no subcommand", nil, ""},
		{"an unknown subcommand", append([]string{"open"}, base...), ""},
		{"no session token", append(append([]string{"prepare"}, base...), "--harness", "claude", "--merge-base", f.wt+"="+f.mb), ""},
		{"a token not mf-", append(append([]string{"prepare"}, base...), "--harness", "claude", "--session-token", "tok", "--merge-base", f.wt+"="+f.mb), ""},
		{"no harness", append(append([]string{"prepare"}, base...), "--session-token", "mf-tok", "--merge-base", f.wt+"="+f.mb), ""},
		{"a merge base not <wt>=<sha>", append(append([]string{"prepare"}, base...), "--harness", "claude", "--session-token", "mf-tok", "--merge-base", f.mb), ""},
		{"--started without --continue", append(append(append([]string{"prepare"}, base...), ok...), "--started"), ""},
		{"an app not <wt>:<root>=<url>", append(append(append([]string{"prepare"}, base...), ok...), "--app", ".="+f.app.URL), "${W}"},
		{"a zsh-mangled app", append(append(append([]string{"prepare"}, base...), ok...), "--app", "/s/web=h/btp://x"), "${W}"},
	} {
		os.Remove(f.log)
		var out, errb bytes.Buffer
		env := Env{Getenv: f.getenv, Environ: func() []string { return f.environ }, Dir: f.wt}
		rc := Registry["visual-verify-pass"](c.args, env, &out, &errb)
		calls, _ := os.ReadFile(f.log)
		if rc != 2 || !strings.Contains(errb.String(), "usage: flow-guard visual-verify-pass") || len(calls) != 0 || !strings.Contains(errb.String(), c.hint) {
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

func TestVisualVerifyPassPrepareSkipNeedsNoKickoff(t *testing.T) {
	f := vvpFixture(t, "")
	other := filepath.Join(filepath.Dir(f.wt), "regress")
	writeFile(t, filepath.Join(other, "README"), "x\n")
	gitRun(t, other, "init", "-q", "-b", "main")
	gitRun(t, other, "add", "-A")
	gitRun(t, other, "commit", "-qm", "r")
	writeFile(t, filepath.Join(f.wt, ".flow/project.md"), "# project\n\n## apps\n\n| App | Repo root |\n|-----|-----------|\n| web | `"+other+"` |\n\n"+vvpSection+"| `regression checkout` | `"+other+"` |\n| `regression repo` | `x/regress` |\n")
	gitRun(t, f.wt, "commit", "-qam", "cfg")
	writeFile(t, filepath.Join(f.wt, ".superpowers/sdd/decision.json"), `{"visual":{"verify":"skipped","reason":"backend only"}}`)
	f.green()
	rc, out, calls := f.prepare()
	if rc != 0 || strings.Contains(out, "STOP: kickoff") || !strings.Contains(out, "NEXT: skip — skipped at Decide — backend only") || !strings.Contains(calls, "-outcome skipped") {
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

// lines is out's lines starting with prefix.
func vvpLines(out, prefix string) []string {
	var got []string
	for _, l := range strings.Split(out, "\n") {
		if strings.HasPrefix(l, prefix) {
			got = append(got, l)
		}
	}
	return got
}

// sibling adds a second repository's worktree of the change, with its own
// `ui paths` and one UI commit, and returns it with its merge base.
func (f *vvpFix) sibling(name string) (string, string) {
	t := f.t
	w := filepath.Join(filepath.Dir(f.wt), name)
	writeFile(t, filepath.Join(w, ".flow/project.md"), "# project\n\n"+vvpSection)
	writeFile(t, filepath.Join(w, "src/ui/b.tsx"), "base\n")
	gitRun(t, w, "init", "-q", "-b", "spectre/demo")
	gitRun(t, w, "add", "-A")
	gitRun(t, w, "commit", "-qm", "base")
	mb := vgHead(t, w)
	vgCommit(t, w, "src/ui/b.tsx")
	return w, mb
}

func TestVisualVerifyPassPrepareStartStack(t *testing.T) {
	f := vvpFixture(t, "")
	dead := httptest.NewServer(http.NotFoundHandler())
	dead.Close()
	args := []string{f.wt, "demo", "visual-verify", "--harness", "claude", "--session-token", "mf-tok",
		"--merge-base", f.wt + "=" + f.mb, "--app", f.wt + ":.=" + dead.URL}
	rc, out, calls := f.run("prepare", args...)
	if rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: start-stack") || strings.Contains(calls, "dispatch begin") {
		t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
	}
	rc, out, calls = f.run("prepare", append(args, "--continue", "--started")...)
	if rc != 1 || !strings.Contains(out, "STOP: stack — ") || !strings.Contains(calls, "-outcome stopped") || strings.Contains(calls, "dispatch begin") {
		t.Fatalf("started yet unanswered: exit %d\n%s\n%s", rc, out, calls)
	}
}

func TestVisualVerifyPassPrepareRestartStack(t *testing.T) {
	f := vvpFixture(t, "| `fingerprint` | `echo stale bundle; exit 1` |\n")
	rc, out, calls := f.prepare()
	if rc != 0 || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: restart-stack") || strings.Contains(calls, "dispatch begin") {
		t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
	}
	rc, out, calls = f.prepare("--continue", "--started")
	if rc != 1 || !strings.Contains(out, "STOP: fingerprint — ") || !strings.Contains(out, "stale bundle") || !strings.Contains(calls, "-outcome stopped") || strings.Contains(calls, "stage begin") || strings.Contains(calls, "dispatch begin") {
		t.Fatalf("under --started: exit %d\n%s\n%s", rc, out, calls)
	}

	g := vvpFixture(t, "")
	rc, out, _ = g.prepare()
	b, _ := os.ReadFile(filepath.Join(g.wt, ".superpowers/sdd/verify-prompt-visual-verify.md"))
	if rc != 0 || !strings.Contains(string(b), "fingerprint: not declared — ") {
		t.Fatalf("exit %d\n%s\n%s", rc, out, b)
	}
}

func TestVisualVerifyPassPrepareBundle(t *testing.T) {
	f := vvpFixture(t, "| `mockups` | `design/mockups` |\n| `mockup frame` | `phone` |\n")
	writeFile(t, filepath.Join(f.wt, "design/mockups/f1.png"), "png\n")
	writeFile(t, filepath.Join(f.wt, "design/proto.html"), "<html></html>\n")
	writeFile(t, filepath.Join(f.wt, "spectre/changes/demo/design.md"), "See `design/proto.html` and `design/gone.css`.\n")
	gitRun(t, f.wt, "add", "-A")
	gitRun(t, f.wt, "commit", "-qm", "cfg")
	f.green()
	head := vgHead(t, f.wt)
	rc, out, _ := f.prepare()
	bundle := filepath.Join(f.wt, ".superpowers/sdd/verify-prompt-visual-verify.md")
	if rc != 0 || !strings.Contains(out, "DISPATCH: visual-verify "+bundle) {
		t.Fatalf("exit %d\n%s", rc, out)
	}
	b, _ := os.ReadFile(bundle)
	s := string(b)
	repo, _ := filepath.Abs("../../..")
	for _, want := range []string{
		"worktree: " + f.wt,
		"setup: npm ci",
		"capture: npx playwright test <spec>",
		"specs: " + f.wt + "/specs-since " + f.mb,
		"screenshots: " + filepath.Join(f.wt, "tests/visual"),
		"mockups: " + filepath.Join(f.wt, "design/mockups"),
		"mockup frame: phone",
		"app: " + f.app.URL,
		"npm run dev",
		"stack: found running",
		"fingerprint: not declared — ",
		"verify: verify@" + head[:7],
		"changeRoot: " + filepath.Join(f.wt, "spectre/changes/demo"),
		"**The relay contract.**",
		filepath.Join(f.wt, ".superpowers/sdd/verify-report-visual-verify.md"),
		filepath.Join(repo, "skills/flow/visual-verify-verifier.md"),
		"mockup sources: " + filepath.Join(f.wt, "design/proto.html") + "\n",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("bundle lacks %q:\n%s", want, s)
		}
	}
	if strings.Contains(s, "<abs-worktree>") || strings.Contains(s, "gone.css") {
		t.Errorf("bundle carries an unresolved path or a missing source:\n%s", s)
	}

	g := vvpFixture(t, "")
	writeFile(t, filepath.Join(g.wt, "spectre/changes/demo/design.md"), "No mockup.\n")
	gitRun(t, g.wt, "add", "-A")
	gitRun(t, g.wt, "commit", "-qm", "cfg")
	g.green()
	g.prepare()
	b, _ = os.ReadFile(filepath.Join(g.wt, ".superpowers/sdd/verify-prompt-visual-verify.md"))
	if !strings.Contains(string(b), "mockup sources: none") {
		t.Errorf("no source:\n%s", b)
	}
}

func TestVisualVerifyPassPrepareDecodesBundledPage(t *testing.T) {
	f := vvpFixture(t, "")
	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	_, _ = zw.Write([]byte("body{color:red}"))
	_ = zw.Close()
	manifest := `{"app.css":{"data":"` + base64.StdEncoding.EncodeToString(gz.Bytes()) + `","compressed":true},"a.txt":{"data":"` + base64.StdEncoding.EncodeToString([]byte("plain")) + `","compressed":false}}`
	page := "<html><script type=\"__bundler/template\">\n\"<p>hi</p>\"\n</script><script type=\"__bundler/manifest\">" + manifest + "</script></html>"
	writeFile(t, filepath.Join(f.wt, "design/proto.html"), page)
	writeFile(t, filepath.Join(f.wt, "design/other.html"), "<html><script type=\"__bundler/template\">\n\"<p>two</p>\"\n</script></html>")
	writeFile(t, filepath.Join(f.wt, "spectre/changes/demo/design.md"), "Prototype `design/proto.html`, and `design/other.html`.\n")
	gitRun(t, f.wt, "add", "-A")
	gitRun(t, f.wt, "commit", "-qm", "cfg")
	f.green()
	if rc, out, _ := f.prepare(); rc != 0 {
		t.Fatalf("exit %d\n%s", rc, out)
	}
	// Each bundled page decodes into its own directory (kan-1060 gated review).
	dir := filepath.Join(f.wt, ".superpowers/sdd/verify-mockup-visual-verify/design-proto")
	if b, _ := os.ReadFile(filepath.Join(f.wt, ".superpowers/sdd/verify-mockup-visual-verify/design-other/index.html")); string(b) != "<p>two</p>" {
		t.Errorf("the second page's index.html = %q", b)
	}
	for name, want := range map[string]string{"index.html": "<p>hi</p>", "app.css": "body{color:red}", "a.txt": "plain"} {
		if b, _ := os.ReadFile(filepath.Join(dir, name)); string(b) != want {
			t.Errorf("%s = %q, want %q", name, b, want)
		}
	}
	b, _ := os.ReadFile(filepath.Join(f.wt, ".superpowers/sdd/verify-prompt-visual-verify.md"))
	if !strings.Contains(string(b), dir) {
		t.Errorf("the bundle names no decoded directory:\n%s", b)
	}
}

func TestVisualVerifyPassPrepareBeginsRows(t *testing.T) {
	f := vvpFixture(t, "")
	head := vgHead(t, f.wt)
	rc, out, calls := f.prepare("--round-base", f.wt+"="+f.mb)
	begin := "record dispatch begin -C " + f.wt + " -change demo -role verifier -model opus -effort low -key visual-verify -session-token mf-tok -diff-base " + head + " -round-base " + f.mb + "\n"
	if rc != 0 || !strings.Contains(calls, begin) || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: dispatch 1") || !strings.Contains(f.stageFile(), "row visual-verify "+f.wt+"\n") {
		t.Fatalf("exit %d\n%s\n%s\n%s", rc, out, calls, f.stageFile())
	}

	g := vvpFixture(t, "")
	w, mb := g.sibling("gymie-frontend")
	rc, out, calls = g.prepare("--worktree", w, "--merge-base", w+"="+mb, "--app", w+":.="+g.app.URL)
	keys := vvpLines(out, "DISPATCH: ")
	if rc != 0 || len(keys) != 2 || !strings.HasPrefix(keys[0], "DISPATCH: visual-verify-demo ") || !strings.HasPrefix(keys[1], "DISPATCH: visual-verify-gymie-frontend ") ||
		!strings.Contains(calls, "-key visual-verify-gymie-frontend -session-token mf-tok -diff-base "+vgHead(t, w)+"\n") ||
		!strings.Contains(g.stageFile(), "row visual-verify-gymie-frontend "+w+"\n") || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: dispatch 2") {
		t.Fatalf("two worktrees: exit %d\n%s\n%s", rc, out, calls)
	}

	h := vvpFixture(t, "")
	writeExec(t, filepath.Join(h.bin, "flow"), "#!/bin/sh\necho \"$*\" >> "+h.log+"\ncase \"$1 $2 $3\" in \"record dispatches \"*) cat "+filepath.Join(filepath.Dir(h.log), "rows.json")+";; \"record dispatch begin\") echo dirty tree; exit 1;; esac\n")
	rc, out, calls = h.prepare()
	if rc != 1 || !strings.Contains(out, "STOP: row — flow record dispatch begin refused for visual-verify: dirty tree") || !strings.Contains(calls, "-outcome stopped") || strings.Contains(out, "DISPATCH:") {
		t.Fatalf("a refused begin: exit %d\n%s\n%s", rc, out, calls)
	}

	// A begin refused for the second row ends the first row it already began.
	i := vvpFixture(t, "")
	w, mb = i.sibling("gymie-frontend")
	writeExec(t, filepath.Join(i.bin, "flow"), "#!/bin/sh\necho \"$*\" >> "+i.log+"\ncase \"$*\" in \"record dispatches \"*) cat "+filepath.Join(filepath.Dir(i.log), "rows.json")+";; *\"-key visual-verify-gymie-frontend \"*) echo dirty tree; exit 1;; esac\n")
	rc, out, calls = i.prepare("--worktree", w, "--merge-base", w+"="+mb, "--app", w+":.="+i.app.URL)
	if rc != 1 || !strings.Contains(calls, "record dispatch end -C "+i.wt+" -change demo -key visual-verify-demo -session-token mf-tok -outcome aborted\n") || !strings.Contains(calls, "-outcome stopped") {
		t.Fatalf("a refused second begin: exit %d\n%s\n%s", rc, out, calls)
	}
}

func TestVisualVerifyPassPrepareContinue(t *testing.T) {
	f := vvpFixture(t, "")
	if rc, out, _ := f.prepare(); rc != 0 {
		t.Fatalf("exit %d\n%s", rc, out)
	}
	writeFile(t, f.stagePath(), f.stageFile()+"verdict visual-verify VERIFY-REPORT-INCOMPLETE: step 10 undone\n")
	cont := func(key string, extra ...string) (int, string, string, string) {
		args := append([]string{f.wt, "demo", key, "--harness", "claude", "--session-token", "mf-tok", "--continue", "--app", f.wt + ":.=" + f.app.URL}, extra...)
		rc, out, calls := f.run("prepare", args...)
		b, _ := os.ReadFile(filepath.Join(f.wt, ".superpowers/sdd/verify-prompt-"+key+".md"))
		return rc, out, calls, string(b)
	}
	rc, out, calls, b := cont("visual-verify-rest")
	if rc != 0 || strings.Contains(calls, "stage begin") || !strings.Contains(calls, "-key visual-verify-rest ") ||
		!strings.Contains(b, "## Incomplete attempt") || !strings.Contains(b, filepath.Join(f.wt, ".superpowers/sdd/verify-report-visual-verify.md")) || !strings.Contains(b, "VERIFY-REPORT-INCOMPLETE: step 10 undone") {
		t.Fatalf("rest: exit %d\n%s\n%s\n%s", rc, out, calls, b)
	}
	rc, out, _, b = cont("visual-verify-2")
	if rc != 0 || !strings.Contains(b, "## Previous attempt") || !strings.Contains(b, "verify-report-visual-verify.md") || !strings.Contains(b, "verify-report-visual-verify-rest.md") {
		t.Fatalf("redispatch: exit %d\n%s\n%s", rc, out, b)
	}

	g := vvpFixture(t, "")
	w, mb := g.sibling("gymie-frontend")
	if rc, out, _ := g.prepare("--worktree", w, "--merge-base", w+"="+mb, "--app", w+":.="+g.app.URL); rc != 0 {
		t.Fatalf("exit %d\n%s", rc, out)
	}
	args := []string{g.wt, "demo", "visual-verify-2-gymie-frontend", "--harness", "claude", "--session-token", "mf-tok", "--continue", "--app", g.wt + ":.=" + g.app.URL, "--app", w + ":.=" + g.app.URL}
	rc, out, calls = g.run("prepare", args...)
	if d := vvpLines(out, "DISPATCH: "); rc != 0 || len(d) != 1 || !strings.HasPrefix(d[0], "DISPATCH: visual-verify-2-gymie-frontend ") || strings.Contains(calls, "-key visual-verify-2-gymie-frontend-") {
		t.Fatalf("one worktree's key: exit %d\n%s\n%s", rc, out, calls)
	}
}

// TestVisualVerifyPassPrepareSpecsNameTheSibling: when the canonical worktree
// survives because a sibling's diff matched its `ui paths` under `--repo`, the
// bundle's `specs` row names that sibling and its merge base, the frontend the
// specs command reads (kan-1083 self-review, F3).
func TestVisualVerifyPassPrepareSpecsNameTheSibling(t *testing.T) {
	section := strings.NewReplacer("`src/ui/**`", "`src/ui/**, gymie-frontend/**`",
		"`<frontend-root>/specs-since <merge-base>`", "`node specs.mjs <frontend-root> <merge-base>`").Replace(vvpSection)
	newFix := func() *vvpFix {
		f := vvpFixture(t, "")
		writeFile(t, filepath.Join(f.wt, ".flow/project.md"), "# project\n\n## run\n\n```bash\nnpm run dev\n```\n\n"+section)
		gitRun(t, f.wt, "commit", "-qam", "cfg")
		f.green()
		return f
	}
	// bundle is the one dispatched bundle's text; with a sibling in the set its key is
	// visual-verify-demo, so it is read from the DISPATCH line.
	bundle := func(out string) string {
		d := vvpLines(out, "DISPATCH: ")
		if len(d) != 1 {
			return ""
		}
		b, _ := os.ReadFile(strings.Fields(d[0])[2])
		return string(b)
	}

	f := newFix()
	w, mb := f.sibling("gymie-frontend")
	writeFile(t, filepath.Join(w, ".flow/project.md"), "# project\n")
	gitRun(t, w, "commit", "-qam", "no visual verification")
	rc, out, _ := f.prepare("--worktree", w, "--merge-base", w+"="+mb)
	got := vvpLines(bundle(out), "specs: ")
	if rc != 0 || len(got) != 1 || got[0] != "specs: node specs.mjs "+w+" "+mb ||
		strings.Contains(got[0], f.wt+" ") || strings.Contains(got[0], f.mb) {
		t.Fatalf("sibling: exit %d, specs %q\n%s", rc, got, out)
	}
	// A --continue re-dispatch reads the sibling and its merge base back from the stage file.
	rc, out, _ = f.run("prepare", f.wt, "demo", "visual-verify-2-demo", "--harness", "claude", "--session-token", "mf-tok", "--continue", "--app", f.wt+":.="+f.app.URL)
	if got := vvpLines(bundle(out), "specs: "); rc != 0 || len(got) != 1 || got[0] != "specs: node specs.mjs "+w+" "+mb {
		t.Fatalf("continue: exit %d, specs %q\n%s", rc, got, out)
	}

	g := newFix()
	rc, out, _ = g.prepare()
	if got := vvpLines(bundle(out), "specs: "); rc != 0 || len(got) != 1 || got[0] != "specs: node specs.mjs "+g.wt+" "+g.mb {
		t.Fatalf("no sibling: exit %d, specs %q\n%s", rc, got, out)
	}
}

// TestVisualVerifyPassPrepareContinueNamesSurvivorWithoutApp pins decision
// f3-continue-usage-error (kan-1102 self-review, F3): under --continue a
// survivor no --app pair names is a usage error naming it, never a
// `NEXT: start-stack` for a stack nothing can probe.
func TestVisualVerifyPassPrepareContinueNamesSurvivorWithoutApp(t *testing.T) {
	g := vvpFixture(t, "")
	w, mb := g.sibling("gymie-frontend")
	if rc, out, _ := g.prepare("--worktree", w, "--merge-base", w+"="+mb, "--app", w+":.="+g.app.URL); rc != 0 {
		t.Fatalf("exit %d\n%s", rc, out)
	}
	var out, errb bytes.Buffer
	env := Env{Getenv: g.getenv, Environ: func() []string { return g.environ }, Dir: g.wt}
	writeFile(t, filepath.Join(filepath.Dir(g.log), "rows.json"), g.rows)
	rc := Registry["visual-verify-pass"]([]string{"prepare", g.wt, "demo", "visual-verify-2", "--harness", "claude", "--session-token", "mf-tok", "--continue", "--app", w + ":.=" + g.app.URL}, env, &out, &errb)
	if rc != 2 || !strings.Contains(errb.String(), "no --app passed for "+g.wt) || strings.Contains(out.String(), "NEXT: start-stack") {
		t.Fatalf("exit %d\nstdout %s\nstderr %s", rc, out.String(), errb.String())
	}
}

// TestVisualVerifyPassPrepareNamesTheAgentCall pins kan-1086's self-review,
// F9: the verifier's Agent call was hand-assembled around the bundle path;
// the DISPATCH line now names its description, subagent_type, model and
// Role line, from the pair the row begins on.
func TestVisualVerifyPassPrepareNamesTheAgentCall(t *testing.T) {
	f := vvpFixture(t, "")
	f.green()
	rc, out, _ := f.prepare()
	d := vvpLines(out, "DISPATCH: ")
	want := ` — description "visual-verify", subagent_type flow-low, model opus, prompt opens "Role: visual-verifier"`
	if rc != 0 || len(d) != 1 || !strings.HasSuffix(d[0], want) {
		t.Fatalf("exit %d, DISPATCH lines %q; want one ending %q\n%s", rc, d, want, out)
	}
}

// TestVisualVerifyPassPrepareDefaultsAppsFromTheProject pins kan-1086's
// self-review, F10: a survivor no --app names takes its apps from the
// project's `preflight apps` row, each `${VAR}` replaced from that
// worktree's prepare-workspace output and a pair whose root is absent
// dropped; an explicit --app replaces the declared set.
func TestVisualVerifyPassPrepareDefaultsAppsFromTheProject(t *testing.T) {
	iso := "\n## workspace isolation\n\n| Resource | Variable | Default | In a workspace |\n|---|---|---|---|\n| `port` | `FRONTEND_PORT` | `3000` | `+<offset>` |\n"
	f := vvpFixture(t, "| `preflight apps` | `web=http://localhost:${FRONTEND_PORT}`, `missing=http://localhost:1` |\n"+iso)
	mkdir(t, filepath.Join(f.wt, "web"))
	var kv bytes.Buffer
	if rc := Registry["prepare-workspace"]([]string{f.wt}, Env{Getenv: f.getenv, Environ: func() []string { return f.environ }, Dir: f.wt}, &kv, &bytes.Buffer{}); rc != 0 {
		t.Fatalf("prepare-workspace exit %d", rc)
	}
	port := ""
	for _, l := range strings.Split(kv.String(), "\n") {
		if v, ok := strings.CutPrefix(l, "FRONTEND_PORT="); ok {
			port = v
		}
	}
	if port == "" || port == "3000" {
		t.Fatalf("prepare-workspace exported FRONTEND_PORT %q", port)
	}
	args := []string{f.wt, "demo", "visual-verify", "--harness", "claude", "--session-token", "mf-tok", "--merge-base", f.wt + "=" + f.mb}
	if rc, out, _ := f.run("prepare", args...); rc != 0 || !strings.Contains(out, "NEXT: start-stack") || strings.Contains(out, "no --app passed") {
		t.Fatalf("declared apps: exit %d; want preflight passed and the stack asked for\n%s", rc, out)
	}
	// Under --started an unanswered probe stops naming every URL it tried.
	rc, out, _ := f.run("prepare", append(args, "--continue", "--started")...)
	want := "stack — http://localhost:" + port + " answers nothing"
	if rc != 1 || !strings.Contains(out, want) || strings.Contains(out, "no --app passed") {
		t.Fatalf("declared apps under --continue: exit %d; want a stop naming %q alone\n%s", rc, want, out)
	}
	if rc, out, _ := f.prepare(); rc != 0 || !strings.Contains(out, "DISPATCH: visual-verify ") {
		t.Fatalf("explicit --app: exit %d\n%s", rc, out)
	}
}
