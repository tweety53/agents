package guard

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/smoke"
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
	// The `specs` row's command: it prints no spec until a test rewrites it.
	writeExec(t, filepath.Join(f.wt, "specs-since"), "#!/bin/sh\n")
	fixtureRepo(t, f.wt, repoShape{Branch: "spectre/demo"})
	gitRun(t, f.wt, "add", "-A")
	gitRun(t, f.wt, "commit", "-qm", "base")
	f.mb = vgHead(t, f.wt)
	f.commit("src/ui/app.tsx")
	f.green()
	// `policy get` answers policy.json when a test wrote one, else the
	// seeded pairs of the next `flow` on PATH (TestMain's stub).
	writeExec(t, filepath.Join(f.bin, "flow"), "#!/bin/sh\necho \"$*\" >> "+f.log+"\ncase \"$1 $2\" in \"record dispatches\") cat "+filepath.Join(root, "rows.json")+";;\n"+
		"\"policy get\") if [ -f "+filepath.Join(root, "policy.json")+" ]; then cat "+filepath.Join(root, "policy.json")+"; else PATH=${PATH#*:} exec flow \"$@\"; fi;; esac\n")
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
		{"defects not <wt>=<frames>", append(append(append([]string{"prepare"}, base...), ok...), "--round-base", f.wt+"="+f.mb, "--defects", "f1"), ""},
		{"an empty defect frame", append(append(append([]string{"prepare"}, base...), ok...), "--round-base", f.wt+"="+f.mb, "--defects", f.wt+"=f1,"), ""},
		{"defects with no round base", append(append(append([]string{"prepare"}, base...), ok...), "--defects", f.wt+"=f1"), ""},
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
	fixtureRepo(t, other, repoShape{Branch: "main"})
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
	fixtureRepo(t, other, repoShape{Branch: "main"})
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
	writeExec(t, filepath.Join(f.wt, "specs-since"), "#!/bin/sh\necho tests/visual/a.spec.ts\n")
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
	fixtureRepo(t, w, repoShape{Branch: "spectre/demo"})
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

// smoke: core — the bundle a visual-verify pass prepares
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
		// The two guards close runs on the report, so the verifier runs them first.
		"check: flow-guard check-verify-report " + filepath.Join(f.wt, ".superpowers/sdd/verify-report-visual-verify.md") + " <motions named> --design " + filepath.Join(f.wt, "spectre/changes/demo/design.md") + "\n",
		"check: flow-guard check-frame-list " + filepath.Join(f.wt, "spectre/changes/demo/design.md") + " " + filepath.Join(f.wt, "design/mockups") + " " + filepath.Join(f.wt, ".superpowers/sdd/verify-report-visual-verify.md") + "\n",
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
	if !strings.Contains(string(b), "check: flow-guard check-verify-report ") || strings.Contains(string(b), "check-frame-list") {
		t.Errorf("no `mockups` declared, so the report check alone:\n%s", b)
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

// TestVisualVerifyPassPrepareReadsThePair pins KAN-1110: the verifier's pair
// is the first base pair of its allowed-pairs row, not a constant; an empty
// row, and a failed allowed-pairs read, stops prepare.
func TestVisualVerifyPassPrepareReadsThePair(t *testing.T) {
	smoke.Heavy(t)
	f := vvpFixture(t, "")
	policy := filepath.Join(filepath.Dir(f.log), "policy.json")
	writeFile(t, policy, `{"visual-verifier":{"base":["sonnet/flow-low"],"optin":[],"marked":[]}}`)
	rc, out, calls := f.prepare()
	if rc != 0 || !strings.Contains(calls, "-role verifier -model sonnet -effort low -key visual-verify ") || !strings.Contains(out, "subagent_type flow-low, model sonnet,") {
		t.Fatalf("sonnet row: exit %d\n%s\n%s", rc, out, calls)
	}
	g := vvpFixture(t, "")
	writeFile(t, filepath.Join(filepath.Dir(g.log), "policy.json"), `{"visual-verifier":{"base":[],"optin":[],"marked":[]}}`)
	rc, out, calls = g.prepare()
	if rc != 1 || !strings.Contains(out, "STOP: pair — ") || strings.Contains(calls, "record dispatch begin") || !strings.Contains(calls, "-outcome stopped") {
		t.Fatalf("empty row: exit %d\n%s\n%s", rc, out, calls)
	}
	h := vvpFixture(t, "")
	writeFile(t, filepath.Join(filepath.Dir(h.log), "policy.json"), `not json`)
	rc, out, calls = h.prepare()
	if rc != 1 || !strings.Contains(out, "STOP: pair — no allowed pairs: `flow policy get`") || strings.Contains(calls, "record dispatch begin") {
		t.Fatalf("failed policy read: exit %d\n%s\n%s", rc, out, calls)
	}
}

func TestVisualVerifyPassPrepareBeginsRows(t *testing.T) {
	smoke.Heavy(t)
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
	writeExec(t, filepath.Join(h.bin, "flow"), "#!/bin/sh\necho \"$*\" >> "+h.log+"\ncase \"$1 $2 $3\" in \"record dispatches \"*) cat "+filepath.Join(filepath.Dir(h.log), "rows.json")+";; \"record dispatch begin\") echo dirty tree; exit 1;; \"policy get \") PATH=${PATH#*:} exec flow \"$@\";; esac\n")
	rc, out, calls = h.prepare()
	if rc != 1 || !strings.Contains(out, "STOP: row — flow record dispatch begin refused for visual-verify: dirty tree") || !strings.Contains(calls, "-outcome stopped") || strings.Contains(out, "DISPATCH:") {
		t.Fatalf("a refused begin: exit %d\n%s\n%s", rc, out, calls)
	}

	// A begin refused for the second row ends the first row it already began.
	i := vvpFixture(t, "")
	w, mb = i.sibling("gymie-frontend")
	writeExec(t, filepath.Join(i.bin, "flow"), "#!/bin/sh\necho \"$*\" >> "+i.log+"\ncase \"$*\" in \"record dispatches \"*) cat "+filepath.Join(filepath.Dir(i.log), "rows.json")+";; *\"-key visual-verify-gymie-frontend \"*) echo dirty tree; exit 1;; \"policy get\") PATH=${PATH#*:} exec flow \"$@\";; esac\n")
	rc, out, calls = i.prepare("--worktree", w, "--merge-base", w+"="+mb, "--app", w+":.="+i.app.URL)
	if rc != 1 || !strings.Contains(calls, "record dispatch end -C "+i.wt+" -change demo -key visual-verify-demo -session-token mf-tok -outcome aborted\n") || !strings.Contains(calls, "-outcome stopped") {
		t.Fatalf("a refused second begin: exit %d\n%s\n%s", rc, out, calls)
	}
}

// TestVisualVerifyPassPrepareWritesTheRoundFile pins decisions
// round-file-per-row and no-round-file-for-rest: a row begun with a round
// base gets visual-round-<key>.txt holding the base and the instant read
// before its begin; the same key re-begun unscoped loses it; a -rest key
// gets none.
func TestVisualVerifyPassPrepareWritesTheRoundFile(t *testing.T) {
	smoke.Heavy(t)
	f := vvpFixture(t, "")
	sdd := filepath.Join(f.wt, ".superpowers/sdd")
	before := time.Now()
	if rc, out, _ := f.prepare("--round-base", f.wt+"="+f.mb); rc != 0 {
		t.Fatalf("scoped prepare: exit %d\n%s", rc, out)
	}
	after := time.Now()
	b, err := os.ReadFile(filepath.Join(sdd, "visual-round-visual-verify.txt"))
	if err != nil {
		t.Fatalf("scoped prepare wrote no round file: %v", err)
	}
	fields := strings.Fields(string(b))
	if len(fields) != 2 || fields[0] != f.mb {
		t.Fatalf("round file %q, want `%s <instant>`", b, f.mb)
	}
	if bb, err := os.ReadFile(filepath.Join(sdd, "visual-round-visual-verify.bases")); err != nil || !strings.Contains(string(bb), f.wt+" "+f.mb) {
		t.Fatalf("scoped prepare wrote no per-checkout round bases %q (%v), want a `<worktree> <sha>` line", bb, err)
	}
	at, err := time.Parse(time.RFC3339Nano, fields[1])
	if err != nil || at.Before(before) || at.After(after) {
		t.Fatalf("round file instant %q (%v), want an RFC3339Nano instant within [%s, %s]", fields[1], err, before.Format(time.RFC3339Nano), after.Format(time.RFC3339Nano))
	}

	cont := func(key string, extra ...string) {
		t.Helper()
		args := append([]string{f.wt, "demo", key, "--harness", "claude", "--session-token", "mf-tok", "--continue", "--app", f.wt + ":.=" + f.app.URL}, extra...)
		if rc, out, _ := f.run("prepare", args...); rc != 0 {
			t.Fatalf("continue %s: exit %d\n%s", key, rc, out)
		}
	}
	cont("visual-verify")
	if _, err := os.Stat(filepath.Join(sdd, "visual-round-visual-verify.txt")); !os.IsNotExist(err) {
		t.Fatalf("an unscoped re-begin left the round file (stat err %v)", err)
	}
	cont("visual-verify-rest", "--round-base", f.wt+"="+f.mb)
	if _, err := os.Stat(filepath.Join(sdd, "visual-round-visual-verify-rest.txt")); !os.IsNotExist(err) {
		t.Fatalf("a -rest key got a round file (stat err %v)", err)
	}
}

// TestVisualVerifyPassPrepareWritesTheDefects pins decisions defects-flag and
// defects-in-round-file: a scoped row's --defects frames are the round
// file's third field, comma-joined (kan-1122 self-review, F1).
func TestVisualVerifyPassPrepareWritesTheDefects(t *testing.T) {
	smoke.Heavy(t)
	f := vvpFixture(t, "")
	if rc, out, _ := f.prepare("--round-base", f.wt+"="+f.mb, "--defects", f.wt+"=f1,f2"); rc != 0 {
		t.Fatalf("scoped prepare with --defects: exit %d\n%s", rc, out)
	}
	b, err := os.ReadFile(filepath.Join(f.wt, ".superpowers/sdd/visual-round-visual-verify.txt"))
	fields := strings.Fields(string(b))
	if err != nil || len(fields) != 3 || fields[0] != f.mb || fields[2] != "f1,f2" {
		t.Fatalf("round file %q (%v), want `%s <instant> f1,f2`", b, err, f.mb)
	}
}

// TestVisualVerifyPassPrepareForwardsBeginWarning: a stage begin that
// succeeds with output still reaches the operator.
func TestVisualVerifyPassPrepareForwardsBeginWarning(t *testing.T) {
	smoke.Heavy(t)
	f := vvpFixture(t, "")
	const warn = "flow: warning: this stage begin superseded 1 still-open run of the same session"
	writeExec(t, filepath.Join(f.bin, "flow"), "#!/bin/sh\necho \"$*\" >> "+f.log+"\ncase \"$1 $2\" in \"record dispatches\") cat "+filepath.Join(filepath.Dir(f.log), "rows.json")+";; \"stage begin\") echo '"+warn+"' >&2;; \"policy get\") PATH=${PATH#*:} exec flow \"$@\";; esac\n")
	rc, out, calls := f.prepare()
	if rc != 0 || !strings.Contains(calls, "stage begin ") || !strings.Contains(out, warn) {
		t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
	}
}

func TestVisualVerifyPassPrepareContinue(t *testing.T) {
	smoke.Heavy(t)
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
	smoke.Heavy(t)
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

// TestVisualVerifyPassPrepareChecksOnlyCapturedSidecars pins kan-1086's
// self-review, F11: preflight judged every sidecar under the capture root,
// so a foreign spec's stale sidecar stopped a pass that never captures it.
// With `specs` declared only the printed specs' sidecars are checked.
func TestVisualVerifyPassPrepareChecksOnlyCapturedSidecars(t *testing.T) {
	smoke.Heavy(t)
	setup := func(t *testing.T, specs string) *vvpFix {
		section := "| `mockups` | `design/mockups` |\n"
		f := vvpFixture(t, section)
		if specs == "" {
			b, _ := os.ReadFile(filepath.Join(f.wt, ".flow/project.md"))
			writeFile(t, filepath.Join(f.wt, ".flow/project.md"), strings.Replace(string(b), "| `specs` | `<frontend-root>/specs-since <merge-base>` |\n", "", 1))
		} else {
			writeExec(t, filepath.Join(f.wt, "specs-since"), "#!/bin/sh\n"+specs)
		}
		writeFile(t, filepath.Join(f.wt, "design/mockups/f1.png"), "png\n")
		writeFile(t, filepath.Join(f.wt, "a.spec.ts"), "shot('a.png')\n")
		writeFile(t, filepath.Join(f.wt, "a.spec.ts.mockups"), "a.png f1\n")
		writeFile(t, filepath.Join(f.wt, "foreign.spec.ts"), "shot('b.png')\n")
		writeFile(t, filepath.Join(f.wt, "foreign.spec.ts.mockups"), "b.png gone\n")
		gitRun(t, f.wt, "add", "-A")
		gitRun(t, f.wt, "commit", "-qm", "cfg")
		f.green()
		return f
	}
	t.Run("the captured spec's sidecar alone", func(t *testing.T) {
		if rc, out, _ := setup(t, "echo a.spec.ts\n").prepare(); rc != 0 || strings.Contains(out, "STOP:") {
			t.Fatalf("exit %d\n%s", rc, out)
		}
	})
	t.Run("a captured spec's stale sidecar", func(t *testing.T) {
		if rc, out, _ := setup(t, "echo foreign.spec.ts\n").prepare(); rc != 1 || !strings.Contains(out, "check-sidecar-reach") || !strings.Contains(out, "gone") {
			t.Fatalf("exit %d\n%s", rc, out)
		}
	})
	t.Run("a failing specs command", func(t *testing.T) {
		if rc, out, _ := setup(t, "exit 1\n").prepare(); rc != 1 || !strings.Contains(out, "specs-since") {
			t.Fatalf("exit %d\n%s", rc, out)
		}
	})
	t.Run("no specs row checks the whole root", func(t *testing.T) {
		if rc, out, _ := setup(t, "").prepare(); rc != 1 || !strings.Contains(out, "check-sidecar-reach") || !strings.Contains(out, "foreign") {
			t.Fatalf("exit %d\n%s", rc, out)
		}
	})
}

// kan-1096: with `design repo` declared, prepare runs render-frames on the
// sidecars before check-sidecar-reach; a reference absent at its tag stops
// preflight naming render-frames, and one already rendered passes.
func TestVisualVerifyPassPrepareRendersFrames(t *testing.T) {
	setup := func(t *testing.T, frame string) *vvpFix {
		design, _ := rfFixture(t, "/nowhere")
		f := vvpFixture(t, "| `mockups` | `design/mockups` |\n| `regression checkout` | `.` |\n| `design repo` | `"+design+" "+design+"` |\n")
		b, _ := os.ReadFile(filepath.Join(f.wt, ".flow/project.md"))
		writeFile(t, filepath.Join(f.wt, ".flow/project.md"), strings.Replace(string(b), "| `specs` | `<frontend-root>/specs-since <merge-base>` |\n", "", 1))
		writeFile(t, filepath.Join(f.wt, "a.spec.ts"), "shot('a.png')\n")
		writeFile(t, filepath.Join(f.wt, "a.spec.ts.mockups"), "a.png "+frame+"\n")
		gitRun(t, f.wt, "add", "-A")
		gitRun(t, f.wt, "commit", "-qm", "cfg")
		f.green()
		rfWritePNG(t, filepath.Join(f.wt, "design/mockups/t1/F1.png"), true)
		return f
	}
	t.Run("a reference absent at its tag", func(t *testing.T) {
		if rc, out, _ := setup(t, "t1/NOPE").prepare(); rc != 1 || !strings.Contains(out, "render-frames") || !strings.Contains(out, "t1/NOPE") {
			t.Fatalf("exit %d\n%s", rc, out)
		}
	})
	t.Run("a reference already rendered", func(t *testing.T) {
		if rc, out, _ := setup(t, "t1/F1").prepare(); rc != 0 || strings.Contains(out, "STOP:") {
			t.Fatalf("exit %d\n%s", rc, out)
		}
	})
}

// TestVisualVerifyPassPrepareRoundDiffReadsTheSiblings: the canonical worktree
// survives on a sibling's diff, so a --round-base on the sibling decides
// "round diff empty" too, never the canonical's own empty diff alone.
func TestVisualVerifyPassPrepareRoundDiffReadsTheSiblings(t *testing.T) {
	section := strings.NewReplacer("`src/ui/**`", "`src/ui/**, gymie-frontend/**`").Replace(vvpSection)
	f := vvpFixture(t, "")
	writeFile(t, filepath.Join(f.wt, ".flow/project.md"), "# project\n\n## run\n\n```bash\nnpm run dev\n```\n\n"+section)
	gitRun(t, f.wt, "commit", "-qam", "cfg")
	f.green()
	w, mb := f.sibling("gymie-frontend")
	writeFile(t, filepath.Join(w, ".flow/project.md"), "# project\n")
	gitRun(t, w, "commit", "-qam", "no visual verification")
	rc, out, calls := f.prepare("--worktree", w, "--merge-base", w+"="+mb, "--app", w+":.="+f.app.URL,
		"--round-base", f.wt+"="+vgHead(t, f.wt), "--round-base", w+"="+mb)
	if rc != 0 || strings.Contains(out, "round diff empty") || !strings.Contains(calls, "dispatch begin") {
		t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
	}
}

// vvpFrames is "<prefix><from>" … "<prefix><to>".
func vvpFrames(prefix string, from, to int) []string {
	var ids []string
	for i := from; i <= to; i++ {
		ids = append(ids, prefix+strconv.Itoa(i))
	}
	return ids
}

// vvpOwnShots is one (spec, screenshot, frame) line per frame, each frame on
// a screenshot of its own.
func vvpOwnShots(spec string, frames []string) [][3]string {
	var lines [][3]string
	for _, f := range frames {
		lines = append(lines, [3]string{spec, "shot-" + f + ".png", f})
	}
	return lines
}

// TestVvpSlices pins decisions four-shards-slices-grow and slice-by-frame-id:
// n = min(4, ⌈F/25⌉) contiguous slices, each closing at the first group
// boundary at or past ⌈F/n⌉ frames, a group being the frames one screenshot
// joins across every spec; fewer slices when the groups run out, never an
// empty one; a declared frame with no sidecar line in the first.
func TestVvpSlices(t *testing.T) {
	t.Parallel()
	sizes := func(s [][]string) []int {
		var n []int
		for _, x := range s {
			n = append(n, len(x))
		}
		return n
	}
	joined := vvpOwnShots("a.spec.ts", vvpFrames("F", 1, 60))
	joined[20][1] = joined[19][1] // F20 and F21 drawn on one screenshot
	twoSpecs := append(vvpOwnShots("a.spec.ts", vvpFrames("F", 1, 60)),
		[3]string{"b.spec.ts", "y.png", "F5"}, [3]string{"b.spec.ts", "y.png", "F50"})
	oneBig := vvpOwnShots("a.spec.ts", vvpFrames("F", 1, 60))
	for i := 0; i < 45; i++ {
		oneBig[i][1] = "all.png"
	}
	for _, tc := range []struct {
		name  string
		lines [][3]string
		extra []string
		want  []int
		check func(s [][]string) bool
	}{
		{"25 frames stay one slice", vvpOwnShots("a.spec.ts", vvpFrames("F", 1, 25)), nil, []int{25}, nil},
		{"60 frames cut into three of 20", vvpOwnShots("a.spec.ts", vvpFrames("F", 1, 60)), nil, []int{20, 20, 20}, func(s [][]string) bool { return s[1][0] == "F21" }},
		{"130 frames keep four, slices grow", vvpOwnShots("a.spec.ts", vvpFrames("F", 1, 130)), nil, []int{33, 33, 33, 31}, nil},
		{"a cut inside a screenshot moves forward", joined, nil, []int{21, 20, 19}, func(s [][]string) bool { return s[0][20] == "F21" }},
		{"a frame two specs map joins its group", twoSpecs, nil, []int{20, 20, 20}, func(s [][]string) bool { return slices.Contains(s[0], "F50") }},
		{"groups that run out make fewer slices", oneBig, nil, []int{45, 15}, nil},
		{"a frame with no sidecar line goes first", vvpOwnShots("a.spec.ts", vvpFrames("F", 1, 59)), []string{"Z1"}, []int{20, 20, 20}, func(s [][]string) bool { return s[0][0] == "Z1" }},
	} {
		got := vvpSlices(tc.lines, tc.extra)
		if !slices.Equal(sizes(got), tc.want) || tc.check != nil && !tc.check(got) {
			t.Errorf("%s: slices %v, want sizes %v", tc.name, got, tc.want)
		}
	}
}

// vvpShardFixture is a prepare fixture declaring `mockups`, whose `specs` row
// prints one capture spec mapping n frames, one screenshot each, its `setup`
// and `capture` rows fixture scripts logging every run to run.log; capture
// writes shot-F1.png when png is set.
func vvpShardFixture(t *testing.T, n int, capture string) *vvpFix {
	t.Helper()
	f := vvpFixture(t, "| `mockups` | `design/mockups` |\n")
	b, _ := os.ReadFile(filepath.Join(f.wt, ".flow/project.md"))
	// The section's `capture` row is its one cell ending `<spec>`.
	project := regexp.MustCompile("`[^`|]*<spec>`").ReplaceAllString(string(b), "`./cap <spec>`")
	writeFile(t, filepath.Join(f.wt, ".flow/project.md"), strings.Replace(project, "`npm ci`", "`./setup`", 1))
	log := filepath.Join(f.wt, "run.log")
	writeExec(t, filepath.Join(f.wt, "setup"), "#!/bin/sh\necho setup >> "+log+"\n")
	writeExec(t, filepath.Join(f.wt, "cap"), "#!/bin/sh\necho \"capture $*\" >> "+log+"\n"+capture)
	writeExec(t, filepath.Join(f.wt, "specs-since"), "#!/bin/sh\necho tests/visual/a.spec.ts\n")
	var spec, sidecar strings.Builder
	for _, id := range vvpFrames("F", 1, n) {
		writeFile(t, filepath.Join(f.wt, "design/mockups", id+".png"), "png\n")
		fmt.Fprintf(&spec, "shot('shot-%s.png')\n", id)
		fmt.Fprintf(&sidecar, "shot-%s.png %s\n", id, id)
	}
	writeFile(t, filepath.Join(f.wt, "tests/visual/a.spec.ts"), spec.String())
	writeFile(t, filepath.Join(f.wt, "tests/visual/a.spec.ts.mockups"), sidecar.String())
	writeFile(t, filepath.Join(f.wt, "spectre/changes/demo/design.md"), fmt.Sprintf("Frames F1 … F%d.\n", n))
	writeFile(t, filepath.Join(f.wt, ".gitignore"), "run.log\nspectre/changes/\n")
	gitRun(t, f.wt, "add", "-A")
	gitRun(t, f.wt, "commit", "-qm", "frames")
	f.green()
	return f
}

// TestVisualVerifyPassPrepareShards pins decisions
// shard-unscoped-single-worktree, prepare-captures-sharded,
// capture-failure-falls-back-serial, bundle-names-measure-call and
// shard-key-segment (KAN-1121).
func TestVisualVerifyPassPrepareShards(t *testing.T) {
	smoke.Heavy(t)
	runLog := func(f *vvpFix) string { b, _ := os.ReadFile(filepath.Join(f.wt, "run.log")); return string(b) }
	bundle := func(f *vvpFix, key string) string {
		b, _ := os.ReadFile(filepath.Join(f.wt, ".superpowers/sdd/verify-prompt-"+key+".md"))
		return string(b)
	}
	spec := func(f *vvpFix) string { return filepath.Join(f.wt, "tests/visual/a.spec.ts") }
	part := func(f *vvpFix) string {
		return filepath.Join(f.wt, "spectre/changes/demo/visual-verification/record-s1of2.md")
	}

	t.Run("25 frames run one unsharded verifier", func(t *testing.T) {
		f := vvpShardFixture(t, 25, "")
		writeFile(t, part(f), "# old\n")
		rc, out, _ := f.prepare()
		if rc != 0 || strings.Contains(f.stageFile(), "slice ") || len(vvpLines(out, "DISPATCH: ")) != 1 || !strings.Contains(out, "DISPATCH: visual-verify ") || runLog(f) != "" {
			t.Fatalf("exit %d, run log %q\n%s\n%s", rc, runLog(f), out, f.stageFile())
		}
		if !strings.Contains(bundle(f, "visual-verify"), "measure: flow-guard visual-verify-pass measure "+f.wt+" demo visual-verify <spec path>\n") {
			t.Fatalf("the unsharded bundle names no measure call:\n%s", bundle(f, "visual-verify"))
		}
		if _, err := os.Stat(part(f)); err != nil {
			t.Fatalf("an unsharded pass deleted an old record part: %v", err)
		}
	})

	t.Run("60 frames run three shards", func(t *testing.T) {
		f := vvpShardFixture(t, 60, "")
		writeFile(t, part(f), "# old\n")
		rc, out, calls := f.prepare()
		s := f.stageFile()
		slice2 := strings.Join(vvpFrames("F", 21, 40), ",")
		for _, want := range []string{"spec " + spec(f) + "\n", "slice s1of3 " + strings.Join(vvpFrames("F", 1, 20), ",") + "\n", "slice s2of3 " + slice2 + "\n", "captured\n", "row visual-verify-s3of3 " + f.wt + "\n"} {
			if !strings.Contains(s, want) {
				t.Errorf("stage file lacks %q", want)
			}
		}
		d := vvpLines(out, "DISPATCH: ")
		if rc != 0 || len(d) != 3 || !strings.HasPrefix(d[1], "DISPATCH: visual-verify-s2of3 ") || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: dispatch 3") ||
			!strings.Contains(calls, "-key visual-verify-s1of3 ") || runLog(f) != "setup\ncapture "+spec(f)+" "+vvpWorkersFlag+"--update-snapshots=all\n" {
			t.Fatalf("exit %d, run log %q\n%s\n%s", rc, runLog(f), out, s)
		}
		b := bundle(f, "visual-verify-s2of3")
		report := filepath.Join(f.wt, ".superpowers/sdd/verify-report-visual-verify-s2of3.md")
		for _, want := range []string{"slice: s2of3 " + slice2 + "\n", "measure: flow-guard visual-verify-pass measure " + f.wt + " demo visual-verify-s2of3 <spec path> --frames " + slice2 + "\n",
			"design.md --frames " + slice2 + "\n", report + " --frames " + slice2 + "\n"} {
			if !strings.Contains(b, want) {
				t.Errorf("the s2of3 bundle lacks %q:\n%s", want, b)
			}
		}
		if _, err := os.Stat(part(f)); !os.IsNotExist(err) {
			t.Errorf("a sharded pass left an old record part (stat err %v)", err)
		}
	})

	t.Run("every spec is captured in one run", func(t *testing.T) {
		f := vvpShardFixture(t, 60, "")
		writeFile(t, filepath.Join(f.wt, "tests/visual/b.spec.ts"), "shot('b.png')\n")
		writeExec(t, filepath.Join(f.wt, "specs-since"), "#!/bin/sh\necho tests/visual/a.spec.ts tests/visual/b.spec.ts\n")
		gitRun(t, f.wt, "add", "-A")
		gitRun(t, f.wt, "commit", "-qm", "second spec")
		f.green()
		if rc, out, _ := f.prepare(); rc != 0 || runLog(f) != "setup\ncapture "+spec(f)+" "+filepath.Join(f.wt, "tests/visual/b.spec.ts")+" "+vvpWorkersFlag+"--update-snapshots=all\n" {
			t.Fatalf("exit %d, run log %q\n%s", rc, runLog(f), out)
		}
	})

	t.Run("a failed capture falls back to one verifier", func(t *testing.T) {
		f := vvpShardFixture(t, 60, "exit 1\n")
		rc, out, _ := f.prepare()
		if rc != 0 || !strings.Contains(out, "SHARDS: none — ./cap "+spec(f)+" "+vvpWorkersFlag+"--update-snapshots=all exit 1") || len(vvpLines(out, "DISPATCH: ")) != 1 ||
			!strings.Contains(out, "DISPATCH: visual-verify ") || strings.Contains(f.stageFile(), "slice ") {
			t.Fatalf("exit %d\n%s\n%s", rc, out, f.stageFile())
		}
	})

	t.Run("a missing merge-base baseline stops no shard", func(t *testing.T) {
		shots := filepath.Join("tests/visual/a.spec.ts-snapshots")
		f := vvpShardFixture(t, 60, "mkdir -p "+shots+" && printf png > "+shots+"/shot-F1-darwin.png\n")
		rc, out, _ := f.prepare()
		if rc != 0 || strings.Contains(out, "MERGE-BASE-SPEC: ") || len(vvpLines(out, "DISPATCH: ")) != 3 {
			t.Fatalf("exit %d\n%s", rc, out)
		}
		// A shard's remainder after a merge-base capture that failed.
		args := []string{f.wt, "demo", "visual-verify-rest-s3of3", "--harness", "claude", "--session-token", "mf-tok", "--continue", "--app", f.wt + ":.=" + f.app.URL, "--merge-base-failed", "stack down"}
		rc, out, _ = f.run("prepare", args...)
		if rc != 0 || len(vvpLines(out, "DISPATCH: ")) != 1 || strings.Count(runLog(f), "capture ") != 1 || !strings.Contains(bundle(f, "visual-verify-rest-s3of3"), "merge-base capture failed — stack down\n") {
			t.Fatalf("--merge-base-failed: exit %d, run log %q\n%s", rc, runLog(f), out)
		}
		args = args[:len(args)-2]
		args[2] = "visual-verify-2-s3of3"
		if rc, out, _ = f.run("prepare", args...); rc != 0 || !strings.Contains(bundle(f, "visual-verify-2-s3of3"), "merge-base capture failed — stack down\n") {
			t.Fatalf("a later shard bundle lost the recorded merge-base failure: exit %d\n%s\n%s", rc, out, bundle(f, "visual-verify-2-s3of3"))
		}
		args[2] = "visual-verify-2-s2of3"
		if rc, out, _ = f.run("prepare", args...); rc != 0 {
			t.Fatalf("s2 redispatch: exit %d\n%s", rc, out)
		}
		report := func(k string) string { return filepath.Join(f.wt, ".superpowers/sdd/verify-report-"+k+".md") }
		if b := bundle(f, "visual-verify-2-s2of3"); !strings.Contains(b, report("visual-verify-s2of3")) || strings.Contains(b, report("visual-verify-s1of3")) {
			t.Fatalf("the s2 redispatch's previous attempt is not its own index's alone:\n%s", b)
		}
	})

	// kan-1099 ran ~32 frames as one verifier: its set held the regression
	// checkout's worktree beside the one surviving worktree.
	t.Run("one survivor of a two-worktree set shards", func(t *testing.T) {
		f := vvpShardFixture(t, 60, "")
		// The regression checkout, a worktree of the set whose diff touches no
		// UI path: it holds the spec, its sidecar and a baseline at its merge base.
		w := filepath.Join(filepath.Dir(f.wt), "gymie-playwright")
		shot := "tests/visual/a.spec.ts-snapshots/shot-F1-darwin.png"
		for _, p := range []string{"tests/visual/a.spec.ts", "tests/visual/a.spec.ts.mockups"} {
			b, _ := os.ReadFile(filepath.Join(f.wt, p))
			writeFile(t, filepath.Join(w, p), string(b))
		}
		writeFile(t, filepath.Join(w, shot), "png\n")
		writeFile(t, filepath.Join(w, ".flow/project.md"), "# project\n")
		writeFile(t, filepath.Join(w, ".gitignore"), "setup\ncap\n")
		log := filepath.Join(f.wt, "run.log")
		writeExec(t, filepath.Join(w, "setup"), "#!/bin/sh\necho setup >> "+log+"\n")
		writeExec(t, filepath.Join(w, "cap"), "#!/bin/sh\necho \"capture $*\" >> "+log+"\nprintf new > "+shot+"\n")
		fixtureRepo(t, w, repoShape{Branch: "spectre/demo"})
		gitRun(t, w, "add", "-A")
		gitRun(t, w, "commit", "-qm", "base")
		mb := vgHead(t, w)
		vgCommit(t, w, "notes.md")
		b, _ := os.ReadFile(filepath.Join(f.wt, ".flow/project.md"))
		writeFile(t, filepath.Join(f.wt, ".flow/project.md"), string(b)+"| `regression checkout` | `"+w+"` |\n| `regression repo` | `x/regress` |\n")
		gitRun(t, f.wt, "add", "-A")
		gitRun(t, f.wt, "commit", "-qm", "regression checkout")
		f.green()

		rc, out, calls := f.prepare("--worktree", w, "--merge-base", w+"="+mb)
		s := f.stageFile()
		d := vvpLines(out, "DISPATCH: ")
		if rc != 0 || len(d) != 3 || !strings.HasPrefix(d[1], "DISPATCH: visual-verify-demo-s2of3 ") || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: dispatch 3") ||
			!strings.Contains(calls, "-key visual-verify-demo-s1of3 ") || !strings.Contains(s, "row visual-verify-demo-s3of3 "+f.wt+"\n") ||
			runLog(f) != "setup\ncapture "+filepath.Join(w, "tests/visual/a.spec.ts")+" "+vvpWorkersFlag+"--update-snapshots=all\n" {
			t.Fatalf("exit %d, run log %q\n%s\n%s", rc, runLog(f), out, s)
		}
		args := []string{f.wt, "demo", "visual-verify-rest-demo-s2of3", "--harness", "claude", "--session-token", "mf-tok", "--continue", "--app", f.wt + ":.=" + f.app.URL}
		rc, out, _ = f.run("prepare", args...)
		if d := vvpLines(out, "DISPATCH: "); rc != 0 || len(d) != 1 || !strings.HasPrefix(d[0], "DISPATCH: visual-verify-rest-demo-s2of3 ") || strings.Count(runLog(f), "capture ") != 1 {
			t.Fatalf("a suffixed shard's remainder: exit %d, run log %q\n%s", rc, runLog(f), out)
		}
		// A key the parent already suffixed shards under that suffix.
		args = []string{f.wt, "demo", "visual-verify-fix-1-full-demo", "--harness", "claude", "--session-token", "mf-tok",
			"--merge-base", f.wt + "=" + f.mb, "--app", f.wt + ":.=" + f.app.URL, "--worktree", w, "--merge-base", w + "=" + mb}
		rc, out, _ = f.run("prepare", args...)
		if d := vvpLines(out, "DISPATCH: "); rc != 0 || len(d) != 3 || !strings.HasPrefix(d[2], "DISPATCH: visual-verify-fix-1-full-demo-s3of3 ") {
			t.Fatalf("a suffixed pass key: exit %d\n%s", rc, out)
		}
	})

	t.Run("a scoped pass and a set with two survivors never shard", func(t *testing.T) {
		f := vvpShardFixture(t, 60, "")
		if rc, out, _ := f.prepare("--round-base", f.wt+"="+f.mb); rc != 0 || strings.Contains(f.stageFile(), "slice ") || len(vvpLines(out, "DISPATCH: ")) != 1 {
			t.Fatalf("--round-base: exit %d\n%s", rc, out)
		}
		g := vvpShardFixture(t, 60, "")
		w, mb := g.sibling("gymie-frontend")
		if rc, out, _ := g.prepare("--worktree", w, "--merge-base", w+"="+mb, "--app", w+":.="+g.app.URL); rc != 0 || strings.Contains(g.stageFile(), "slice ") {
			t.Fatalf("two worktrees: exit %d\n%s", rc, out)
		}
	})

	t.Run("the merge-base flags need --continue", func(t *testing.T) {
		f := vvpShardFixture(t, 60, "")
		if rc, out, _ := f.prepare("--merge-base-failed", "stack down"); rc != 2 || !strings.Contains(out, "--continue") {
			t.Fatalf("exit %d\n%s", rc, out)
		}
	})
}
