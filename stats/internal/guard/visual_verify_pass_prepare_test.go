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

// vvpSectionWith is vvpSection with extra's rows placed: a row of the closed
// command vocabulary joins the commands table, any other table row the
// settings table, and any other line follows the section.
func vvpSectionWith(extra string) (section, tail string) {
	section = vvpSection
	for _, l := range strings.SplitAfter(extra, "\n") {
		key, _, _ := strings.Cut(strings.TrimPrefix(l, "| `"), "`")
		switch {
		case tail != "" || !strings.HasPrefix(l, "| `"):
			tail += l
		case slices.Contains([]string{"setup", "verify", "capture", "fingerprint", "start", "specs", "smoke"}, key):
			section += l
		default:
			section = strings.Replace(section, "| `screenshots` | `tests/visual` |\n", "| `screenshots` | `tests/visual` |\n"+l, 1)
		}
	}
	return section, tail
}

func vvpSect(t *testing.T, extra string) string {
	t.Helper()
	sec, tl := vvpSectionWith(extra)
	return sec + tl
}

func vvpFixture(t *testing.T, extra string) *vvpFix {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := &vvpFix{t: t, wt: filepath.Join(root, "demo"), bin: filepath.Join(root, "bin"), log: filepath.Join(root, "flow.log")}
	section, tail := vvpSectionWith(extra)
	writeFile(t, filepath.Join(f.wt, ".flow/project.md"), "# project\n\n## run\n\n```bash\nnpm run dev\n```\n\n"+section+tail)
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
	// seeded pairs (testPolicy) -- never the dev store's.
	writeExec(t, filepath.Join(f.bin, "flow"), "#!/bin/sh\necho \"$*\" >> "+f.log+"\ncase \"$1 $2\" in \"record dispatches\") cat "+filepath.Join(root, "rows.json")+";;\n"+
		"\"policy get\") if [ -f "+filepath.Join(root, "policy.json")+" ]; then cat "+filepath.Join(root, "policy.json")+"; else printf '%s\\n' '"+testPolicy+"'; fi;; esac\n")
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
	t.Parallel()
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
	t.Parallel()
	f := vvpFixture(t, "")
	other := filepath.Join(filepath.Dir(f.wt), "regress")
	writeFile(t, filepath.Join(other, "README"), "x\n")
	fixtureRepo(t, other, repoShape{Branch: "main"})
	gitRun(t, other, "add", "-A")
	gitRun(t, other, "commit", "-qm", "r")
	gitRun(t, other, "remote", "add", "origin", "x/regress")
	sec, tl := vvpSectionWith("| `regression checkout` | `" + other + "` |\n| `regression repo` | `x/regress` |\n")
	writeFile(t, filepath.Join(f.wt, ".flow/project.md"), "# project\n\n## apps\n\n| App | Repo root |\n|-----|-----------|\n| web | `"+other+"` |\n\n"+sec+tl)
	gitRun(t, f.wt, "commit", "-qam", "cfg")
	f.green()
	rc, out, calls := f.prepare()
	if rc != 1 || !strings.Contains(out, "STOP: kickoff regress") || calls != "" {
		t.Fatalf("exit %d, calls %q\n%s", rc, calls, out)
	}
}

func TestVisualVerifyPassPrepareSkipNeedsNoKickoff(t *testing.T) {
	t.Parallel()
	f := vvpFixture(t, "")
	other := filepath.Join(filepath.Dir(f.wt), "regress")
	writeFile(t, filepath.Join(other, "README"), "x\n")
	fixtureRepo(t, other, repoShape{Branch: "main"})
	gitRun(t, other, "add", "-A")
	gitRun(t, other, "commit", "-qm", "r")
	gitRun(t, other, "remote", "add", "origin", "x/regress")
	sec, tl := vvpSectionWith("| `regression checkout` | `" + other + "` |\n| `regression repo` | `x/regress` |\n")
	writeFile(t, filepath.Join(f.wt, ".flow/project.md"), "# project\n\n## apps\n\n| App | Repo root |\n|-----|-----------|\n| web | `"+other+"` |\n\n"+sec+tl)
	gitRun(t, f.wt, "commit", "-qam", "cfg")
	writeFile(t, filepath.Join(f.wt, ".superpowers/sdd/decision.json"), `{"visual":{"verify":"skipped","reason":"backend only"}}`)
	f.green()
	rc, out, calls := f.prepare()
	if rc != 0 || strings.Contains(out, "STOP: kickoff") || !strings.Contains(out, "NEXT: skip — skipped at Decide — backend only") || strings.Contains(calls, "stage ") {
		t.Fatalf("exit %d, calls %q\n%s", rc, calls, out)
	}
}

func TestVisualVerifyPassPrepareSkips(t *testing.T) {
	t.Parallel()
	f := vvpFixture(t, "")
	writeFile(t, filepath.Join(f.wt, ".superpowers/sdd/decision.json"), `{"visual":{"verify":"skipped","reason":"copy only"}}`)
	rc, out, calls := f.prepare()
	if rc != 0 || !strings.Contains(out, "NEXT: skip — skipped at Decide — copy only") || strings.Contains(calls, "stage ") || strings.Contains(calls, "dispatch begin") {
		t.Fatalf("decided skip: exit %d\n%s\n%s", rc, out, calls)
	}

	g := vvpFixture(t, "")
	rc, out, calls = g.prepare("--since", g.wt+"="+vgHead(t, g.wt))
	if rc != 0 || !strings.Contains(out, "NEXT: skip — ") || strings.Contains(calls, "stage ") || strings.Contains(calls, "dispatch begin") {
		t.Fatalf("no survivor: exit %d\n%s\n%s", rc, out, calls)
	}
}

func TestVisualVerifyPassPrepareComposeOnly(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	f := vvpFixture(t, "")
	f.commit("src/ui/late.tsx")
	rc, out, calls := f.prepare()
	if rc != 1 || !strings.Contains(out, "STOP: verify-not-green") || strings.Contains(calls, "stage ") {
		t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
	}
}

func TestVisualVerifyPassPrepareStopsOnPreflight(t *testing.T) {
	t.Parallel()
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
	if strings.Contains(calls, "stage ") || strings.Contains(calls, "dispatch begin") {
		t.Fatalf("calls:\n%s", calls)
	}
}

func TestVisualVerifyPassPrepareWritesStageFile(t *testing.T) {
	t.Parallel()
	f := vvpFixture(t, "")
	rb := f.mb
	if rc, out, _ := f.prepare("--round-base", f.wt+"="+rb); rc != 0 {
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
	t.Parallel()
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
	if rc != 1 || !strings.Contains(out, "STOP: stack — ") || strings.Contains(calls, "stage ") || strings.Contains(calls, "dispatch begin") {
		t.Fatalf("started yet unanswered: exit %d\n%s\n%s", rc, out, calls)
	}
}

func TestPrepareDoesNotFingerprintStack(t *testing.T) {
	t.Parallel()
	f := vvpFixture(t, "| `fingerprint` | `echo stale bundle; exit 1` |\n")
	rc, out, calls := f.prepare()
	if strings.Contains(out, "NEXT: restart-stack") || !strings.Contains(calls, "dispatch begin") {
		t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
	}
}

// smoke: core — the bundle a visual-verify pass prepares
func TestVisualVerifyPassPrepareBundle(t *testing.T) {
	t.Parallel()
	f := vvpFixture(t, "| `mockups` | `design/mockups` |\n| `mockup frame` | `scale=1 status=0 border=0` |\n")
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
		"mockup frame: scale=1 status=0 border=0",
		"app: " + f.app.URL,
		"npm run dev",
		"stack: found running",
		"verify: verify@" + head[:7],
		"changeRoot: " + filepath.Join(f.wt, "spectre/changes/demo"),
		"**The relay contract.**",
		filepath.Join(f.wt, ".superpowers/sdd/verify-report-visual-verify.md"),
		filepath.Join(repo, "skills/flow/visual-verify-verifier.md"),
		"mockup sources: " + filepath.Join(f.wt, "design/proto.html") + "\n",
		// The one check close runs on the report, so the verifier runs it first.
		"check: flow-guard visual-verify-pass close --check-only " + filepath.Join(f.wt, ".superpowers/sdd/verify-report-visual-verify.md") + " <motions named> --mockups " + filepath.Join(f.wt, "design/mockups") + " --design " + filepath.Join(f.wt, "spectre/changes/demo/design.md") + "\n",
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
	if !strings.Contains(string(b), "check: flow-guard visual-verify-pass close --check-only ") || strings.Contains(string(b), "--mockups") {
		t.Errorf("no `mockups` declared, so the report check alone:\n%s", b)
	}
}

func TestVisualVerifyPassPrepareDecodesBundledPage(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	f := vvpFixture(t, "")
	policy := filepath.Join(filepath.Dir(f.log), "policy.json")
	writeFile(t, policy, `{"visual-verifier":{"base":["sonnet/flow-low"],"optin":[],"marked":[]}}`)
	rc, out, calls := f.prepare()
	if rc != 0 || !strings.Contains(calls, "-role visual-verifier -model sonnet -effort low -key visual-verify ") || !strings.Contains(out, "subagent_type flow-low, model sonnet,") {
		t.Fatalf("sonnet row: exit %d\n%s\n%s", rc, out, calls)
	}
	g := vvpFixture(t, "")
	writeFile(t, filepath.Join(filepath.Dir(g.log), "policy.json"), `{"visual-verifier":{"base":[],"optin":[],"marked":[]}}`)
	rc, out, calls = g.prepare()
	if rc != 1 || !strings.Contains(out, "STOP: pair — ") || strings.Contains(calls, "record dispatch begin") || strings.Contains(calls, "stage ") {
		t.Fatalf("empty row: exit %d\n%s\n%s", rc, out, calls)
	}
	h := vvpFixture(t, "")
	writeFile(t, filepath.Join(filepath.Dir(h.log), "policy.json"), `not json`)
	rc, out, calls = h.prepare()
	if rc != 1 || !strings.Contains(out, "STOP: pair — no allowed pairs: `flow policy get`") || strings.Contains(calls, "record dispatch begin") {
		t.Fatalf("failed policy read: exit %d\n%s\n%s", rc, out, calls)
	}
}

// vvpReceipt appends a passing flow-guard recapture receipt of spec to wt,
// taken on path at blob.
func vvpReceipt(t *testing.T, wt, spec, path, blob string) {
	t.Helper()
	file := filepath.Join(tcfGit(t, wt, "rev-parse", "--absolute-git-dir"), "flow-recapture.jsonl")
	old, _ := os.ReadFile(file)
	writeFile(t, file, string(old)+`{"specs":["`+spec+`"],"exit":0,"blobs":{"`+path+`":"`+blob+`"}}`+"\n")
}

// TestVisualVerifyPassPrepareNamesCapturedSpecs pins the kan-1091 fix: a
// scoped row's bundle names the specs the round's own recapture passed on
// HEAD's sources, so the verifier's capture run does not repeat them -- and
// names none a receipt on other sources would; an unscoped row names the
// specs a task's capture still stands for (rcTaskCaptured).
func TestVisualVerifyPassPrepareNamesCapturedSpecs(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	bundle := func(f *vvpFix) string {
		b, _ := os.ReadFile(filepath.Join(f.wt, ".superpowers/sdd/verify-prompt-visual-verify.md"))
		return string(b)
	}
	const line = "captured: tests/visual/home.spec.ts\n"
	f := vvpFixture(t, "")
	vvpReceipt(t, f.wt, "tests/visual/home.spec.ts", "src/ui/app.tsx", tcfGit(t, f.wt, "rev-parse", "HEAD:src/ui/app.tsx"))
	if rc, out, _ := f.prepare("--round-base", f.wt+"="+f.mb); rc != 0 || !strings.Contains(bundle(f), line) {
		t.Errorf("a receipt on HEAD's sources: exit %d, bundle lacks %q\n%s\n%s", rc, line, out, bundle(f))
	}
	g := vvpFixture(t, "")
	vvpReceipt(t, g.wt, "tests/visual/home.spec.ts", "src/ui/app.tsx", tcfGit(t, g.wt, "rev-parse", g.mb+":src/ui/app.tsx"))
	if rc, out, _ := g.prepare("--round-base", g.wt+"="+g.mb); rc != 0 || strings.Contains(bundle(g), "captured:") {
		t.Errorf("a receipt on an earlier draft: exit %d, bundle names captured specs\n%s\n%s", rc, out, bundle(g))
	}
	// An unscoped row names the specs a task captured on its own commit and
	// nothing has moved since; a later edit of the spec's source un-names it.
	h := vvpFixture(t, "")
	vvpReceipt(t, h.wt, "tests/visual/home.spec.ts", "src/ui/app.tsx", tcfGit(t, h.wt, "rev-parse", "HEAD:src/ui/app.tsx"))
	if rc, out, _ := h.prepare(); rc != 0 || !strings.Contains(bundle(h), line) {
		t.Errorf("an unscoped row over a task's capture: exit %d, bundle lacks %q\n%s\n%s", rc, line, out, bundle(h))
	}
	k := vvpFixture(t, "")
	vvpReceipt(t, k.wt, "tests/visual/home.spec.ts", "src/ui/app.tsx", tcfGit(t, k.wt, "rev-parse", "HEAD:src/ui/app.tsx"))
	writeFile(t, filepath.Join(k.wt, "tests/visual/home.spec.ts"), "test('home', () => {})\n")
	writeFile(t, filepath.Join(k.wt, "tests/visual/home.spec.ts.sources"), "src/ui/app.tsx\n")
	writeFile(t, filepath.Join(k.wt, "src/ui/app.tsx"), "export const later = 1\n")
	gitRun(t, k.wt, "add", "-A")
	gitRun(t, k.wt, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-q", "-m", "a later edit")
	k.green()
	if rc, out, _ := k.prepare(); rc != 0 || strings.Contains(bundle(k), "captured:") {
		t.Errorf("a source edited after its task's capture: exit %d, the bundle still names it\n%s\n%s", rc, out, bundle(k))
	}
}

func TestVisualVerifyPassPrepareBeginsRows(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	f := vvpFixture(t, "")
	head := vgHead(t, f.wt)
	rc, out, calls := f.prepare("--round-base", f.wt+"="+f.mb)
	begin := "record dispatch begin -C " + f.wt + " -change demo -role visual-verifier -model opus -effort low -key visual-verify -session-token mf-tok -diff-base " + head + " -round-base " + f.mb + "\n"
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
	if rc != 1 || !strings.Contains(out, "STOP: row — flow record dispatch begin refused for visual-verify: dirty tree") || strings.Contains(calls, "stage ") || strings.Contains(out, "DISPATCH:") {
		t.Fatalf("a refused begin: exit %d\n%s\n%s", rc, out, calls)
	}

	// A begin refused for the second row ends the first row it already began.
	i := vvpFixture(t, "")
	w, mb = i.sibling("gymie-frontend")
	writeExec(t, filepath.Join(i.bin, "flow"), "#!/bin/sh\necho \"$*\" >> "+i.log+"\ncase \"$*\" in \"record dispatches \"*) cat "+filepath.Join(filepath.Dir(i.log), "rows.json")+";; *\"-key visual-verify-gymie-frontend \"*) echo dirty tree; exit 1;; \"policy get\") PATH=${PATH#*:} exec flow \"$@\";; esac\n")
	rc, out, calls = i.prepare("--worktree", w, "--merge-base", w+"="+mb, "--app", w+":.="+i.app.URL)
	if rc != 1 || !strings.Contains(calls, "record dispatch end -C "+i.wt+" -change demo -key visual-verify-demo -session-token mf-tok -outcome aborted\n") || strings.Contains(calls, "stage ") {
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
	t.Parallel()
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

// TestVisualVerifyPassPrepareStartsAnEmptyManifest: a row prepare begins measures into a manifest
// of its own — an earlier pass's or round's manifest under the same key is deleted, never merged into.
func TestVisualVerifyPassPrepareStartsAnEmptyManifest(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	f := vvpFixture(t, "")
	manifest := filepath.Join(f.wt, ".superpowers/sdd/visual-manifest-visual-verify.json")
	writeFile(t, manifest, `{"pairs":[{"spec":"old.spec.ts","frame":"stale"}]}`)
	if rc, out, _ := f.prepare(); rc != 0 {
		t.Fatalf("prepare: exit %d\n%s", rc, out)
	}
	if _, err := os.Stat(manifest); !os.IsNotExist(err) {
		t.Fatalf("an earlier pass's manifest survived prepare (stat err %v)", err)
	}
}

// TestVisualVerifyPassPrepareWritesTheDefects pins decisions defects-flag and
// defects-in-round-file: a scoped row's --defects frames are the round
// file's third field, comma-joined (kan-1122 self-review, F1).
func TestVisualVerifyPassPrepareWritesTheDefects(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
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

func TestVisualVerifyPassPrepareContinue(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
		t.Parallel()
		if rc, out, _ := setup(t, "echo a.spec.ts\n").prepare(); rc != 0 || strings.Contains(out, "STOP:") {
			t.Fatalf("exit %d\n%s", rc, out)
		}
	})
	t.Run("a captured spec's stale sidecar", func(t *testing.T) {
		t.Parallel()
		if rc, out, _ := setup(t, "echo foreign.spec.ts\n").prepare(); rc != 1 || !strings.Contains(out, "check-sidecar-reach") || !strings.Contains(out, "gone") {
			t.Fatalf("exit %d\n%s", rc, out)
		}
	})
	t.Run("stderr diagnostics are not specs", func(t *testing.T) {
		t.Parallel()
		if rc, out, _ := setup(t, "echo a.spec.ts\necho 'specs-for-diff: skipped foreign.spec.ts not selected' >&2\n").prepare(); rc != 0 || strings.Contains(out, "STOP:") {
			t.Fatalf("exit %d\n%s", rc, out)
		}
	})
	t.Run("a failing specs command", func(t *testing.T) {
		t.Parallel()
		if rc, out, _ := setup(t, "exit 1\n").prepare(); rc != 1 || !strings.Contains(out, "specs-since") {
			t.Fatalf("exit %d\n%s", rc, out)
		}
	})
	t.Run("a scoped pass with defects checks only the sidecars holding them", func(t *testing.T) {
		t.Parallel()
		f := setup(t, "")
		if rc, out, _ := f.prepare("--round-base", f.wt+"="+f.mb, "--defects", f.wt+"=f1"); rc != 0 || strings.Contains(out, "STOP:") {
			t.Fatalf("exit %d\n%s", rc, out)
		}
	})
	t.Run("no specs row checks the whole root", func(t *testing.T) {
		t.Parallel()
		if rc, out, _ := setup(t, "").prepare(); rc != 1 || !strings.Contains(out, "check-sidecar-reach") || !strings.Contains(out, "foreign") {
			t.Fatalf("exit %d\n%s", rc, out)
		}
	})
}

// kan-1096: with `design repo` declared, prepare runs render-frames on the
// sidecars before check-sidecar-reach; a reference absent at its tag stops
// preflight naming render-frames, and one already rendered passes.
func TestVisualVerifyPassPrepareRendersFrames(t *testing.T) {
	t.Parallel()
	setup := func(t *testing.T, frame string) *vvpFix {
		design, _ := rfFixture(t, "/nowhere")
		gitRun(t, design, "remote", "add", "origin", "x/design")
		f := vvpFixture(t, "| `mockups` | `design/mockups` |\n| `regression checkout` | `.` |\n| `regression repo` | `x/regress` |\n| `design repo` | `"+design+" x/design` |\n")
		gitRun(t, f.wt, "remote", "add", "origin", "x/regress")
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
		t.Parallel()
		if rc, out, _ := setup(t, "t1/NOPE").prepare(); rc != 1 || !strings.Contains(out, "render-frames") || !strings.Contains(out, "t1/NOPE") {
			t.Fatalf("exit %d\n%s", rc, out)
		}
	})
	t.Run("a reference already rendered", func(t *testing.T) {
		t.Parallel()
		if rc, out, _ := setup(t, "t1/F1").prepare(); rc != 0 || strings.Contains(out, "STOP:") {
			t.Fatalf("exit %d\n%s", rc, out)
		}
	})
}

// TestVisualVerifyPassPrepareRoundDiffReadsTheSiblings: the canonical worktree
// survives on a sibling's diff, so a --round-base on the sibling decides
// "round diff empty" too, never the canonical's own empty diff alone.
func TestVisualVerifyPassPrepareRoundDiffReadsTheSiblings(t *testing.T) {
	t.Parallel()
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

// vvpFramesFixture is a prepare fixture declaring `mockups`, whose `specs` row
// prints one capture spec mapping n frames, one screenshot each, its `setup`
// and `capture` rows fixture scripts logging every run to run.log; capture
// writes shot-F1.png when png is set.
func vvpFramesFixture(t *testing.T, n int, capture string) *vvpFix {
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

// TestVisualVerifyPassPrepareOneVerifier: a pass is one verifier per worktree
// whatever its frame count — 60 frames begin one row, prepare captures
// nothing itself, and the bundle names the plain measure and check calls.
func TestVisualVerifyPassPrepareOneVerifier(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	runLog := func(f *vvpFix) string { b, _ := os.ReadFile(filepath.Join(f.wt, "run.log")); return string(b) }
	bundle := func(f *vvpFix, key string) string {
		b, _ := os.ReadFile(filepath.Join(f.wt, ".superpowers/sdd/verify-prompt-"+key+".md"))
		return string(b)
	}
	spec := func(f *vvpFix) string { return filepath.Join(f.wt, "tests/visual/a.spec.ts") }

	t.Run("60 frames run one verifier", func(t *testing.T) {
		t.Parallel()
		f := vvpFramesFixture(t, 60, "")
		rc, out, _ := f.prepare()
		d := vvpLines(out, "DISPATCH: ")
		if rc != 0 || len(d) != 1 || !strings.HasPrefix(d[0], "DISPATCH: visual-verify ") || !strings.HasSuffix(strings.TrimSpace(out), "NEXT: dispatch 1") || runLog(f) != "" {
			t.Fatalf("exit %d, run log %q\n%s\n%s", rc, runLog(f), out, f.stageFile())
		}
		b := bundle(f, "visual-verify")
		if !strings.Contains(b, "measure: flow-guard visual-verify-pass measure "+f.wt+" demo visual-verify <spec path>\n") || strings.Contains(b, "--frames") || strings.Contains(b, "slice:") {
			t.Fatalf("the bundle's measure or check call:\n%s", b)
		}
	})

	// earlier is a frames fixture whose `specs` row reaches a.spec.ts and two
	// sidecar-carrying specs of earlier changes, b and c, committed before the
	// merge base; only a.spec.ts changes after it.
	earlier := func(t *testing.T, frames int) *vvpFix {
		f := vvpFramesFixture(t, frames, "")
		for _, n := range []string{"b", "c"} {
			writeFile(t, filepath.Join(f.wt, "design/mockups/"+n+"1.png"), "png\n")
			writeFile(t, filepath.Join(f.wt, "tests/visual/"+n+".spec.ts"), "shot('shot-"+n+"1.png')\n")
			writeFile(t, filepath.Join(f.wt, "tests/visual/"+n+".spec.ts.mockups"), "shot-"+n+"1.png "+n+"1\n")
		}
		writeExec(t, filepath.Join(f.wt, "specs-since"), "#!/bin/sh\necho tests/visual/a.spec.ts tests/visual/b.spec.ts tests/visual/c.spec.ts\n")
		gitRun(t, f.wt, "add", "-A")
		gitRun(t, f.wt, "commit", "-qm", "earlier changes' specs")
		f.mb = vgHead(t, f.wt)
		own, _ := os.ReadFile(spec(f))
		writeFile(t, spec(f), string(own)+"// the change's own\n")
		writeFile(t, filepath.Join(f.wt, "src/ui/app.tsx"), "own\n")
		gitRun(t, f.wt, "add", "-A")
		gitRun(t, f.wt, "commit", "-qm", "the change's own")
		return f
	}

	t.Run("the bundle names only the change's own spec to capture", func(t *testing.T) {
		t.Parallel()
		f := earlier(t, 10)
		f.green()
		rc, out, _ := f.prepare()
		if b := bundle(f, "visual-verify"); rc != 0 || !strings.Contains(b, "own specs: "+spec(f)+"\n") || strings.Contains(b, "b.spec.ts") || strings.Contains(b, "c.spec.ts") {
			t.Fatalf("exit %d\n%s\n%s", rc, out, b)
		}
	})

	t.Run("a pass with no own sidecar names none to capture", func(t *testing.T) {
		t.Parallel()
		f := earlier(t, 10)
		rb := vgHead(t, f.wt)
		writeFile(t, filepath.Join(f.wt, "src/ui/app.tsx"), "round\n")
		gitRun(t, f.wt, "commit", "-qam", "the round's fix")
		f.green()
		rc, out, _ := f.prepare("--round-base", f.wt+"="+rb)
		if b := bundle(f, "visual-verify"); rc != 0 || !strings.Contains(b, "own specs: none\n") {
			t.Fatalf("exit %d\n%s\n%s", rc, out, b)
		}
	})

}

func TestFoldedVisualGuardsAreUnregistered(t *testing.T) {
	t.Parallel()
	for _, n := range []string{"check-frame-list", "check-verify-report", "check-sidecar-reach", "check-verify-green", "check-visual-preflight", "check-visual-verification"} {
		if _, ok := Registry[n]; ok {
			t.Errorf("%s is a guard.Registry key", n)
		}
		if foldedGuards[n] == nil {
			t.Errorf("%s is not run in process", n)
		}
	}
}

func TestPrepareStopsOnInvalidVisualSection(t *testing.T) {
	t.Parallel()
	f := vvpFixture(t, "")
	writeFile(t, filepath.Join(f.wt, ".flow/project.md"), "# project\n\n## visual verification\n\n| Setting | Value |\n|---|---|\n| `ui paths` | |\n")
	rc, out, calls := f.prepare()
	if rc == 0 || strings.Contains(calls, "dispatch begin") || strings.Contains(out, "NEXT:") {
		t.Fatalf("exit %d, want the validator's refusal before any capture\n%s\n%s", rc, out, calls)
	}
	if _, err := os.Stat(f.stagePath()); err == nil {
		t.Errorf("a stage file was written")
	}
}

// TestVisualVerifyPassPrepareRunsTheSmokeRow: a declared `smoke` command runs
// before any verifier row; a non-zero exit is a failing preflight check that
// begins no row, a zero exit prints its measured time and the pass goes on.
func TestVisualVerifyPassPrepareRunsTheSmokeRow(t *testing.T) {
	t.Parallel()
	bad := vvpFixture(t, "| `smoke` | `echo broken-harness; exit 3` |\n")
	rc, out, _ := bad.prepare()
	if rc != 1 || !strings.Contains(out, "STOP: preflight") || !strings.Contains(out, "smoke:") || !strings.Contains(out, "broken-harness") || strings.Contains(out, "DISPATCH:") {
		t.Fatalf("failing smoke: exit %d; want a preflight stop naming smoke and no dispatch\n%s", rc, out)
	}
	good := vvpFixture(t, "| `smoke` | `true` |\n")
	if rc, out, _ := good.prepare(); rc != 0 || !strings.Contains(out, "SMOKE: "+good.wt+" ok in ") || !strings.Contains(out, "DISPATCH: visual-verify ") {
		t.Fatalf("passing smoke: exit %d\n%s", rc, out)
	}
}

func TestVisualVerifyPassPrepareKickoffSkipsReposLinkDoesNotName(t *testing.T) {
	t.Parallel()
	f := vvpFixture(t, "")
	other := filepath.Join(filepath.Dir(f.wt), "admin")
	writeFile(t, filepath.Join(other, "README"), "x\n")
	fixtureRepo(t, other, repoShape{Branch: "main"})
	gitRun(t, other, "add", "-A")
	gitRun(t, other, "commit", "-qm", "r")
	sec, tl := vvpSectionWith("")
	writeFile(t, filepath.Join(f.wt, ".flow/project.md"), "# project\n\n## apps\n\n| App | Repo root |\n|-----|-----------|\n| admin | `"+other+"` |\n\n"+sec+tl)
	writeFile(t, filepath.Join(f.wt, "spectre/peers"), "admin "+other+"\n")
	writeFile(t, filepath.Join(f.wt, "spectre/changes/demo/link.md"), "## Merge order\n\n1. `.`\n")
	gitRun(t, f.wt, "add", "-A")
	gitRun(t, f.wt, "commit", "-qm", "cfg")
	gitRun(t, f.wt, "config", "branch.spectre/demo.flowBase", "develop")
	f.green()
	_, out, _ := f.prepare()
	if strings.Contains(out, "STOP: kickoff") {
		t.Fatalf("an unnamed repository demanded a worktree\n%s", out)
	}
}

func TestVisualVerifyPassPrepareClosesSupersededRows(t *testing.T) {
	t.Parallel()
	f := vvpFixture(t, "")
	f.rows = "[" + vgRow("verify", vgHead(t, f.wt)) +
		`,{"key":"visual-verify-3","role":"visual-verifier","outcome":""},{"key":"visual-verify-4","role":"visual-verifier","outcome":"completed"}]`
	rc, out, calls := f.prepare()
	end := "record dispatch end -C " + f.wt + " -change demo -key visual-verify-3 -session-token mf-tok -outcome aborted\n"
	if rc != 0 || !strings.Contains(calls, end) || strings.Contains(calls, "-key visual-verify-4 -session-token") || strings.Contains(calls, "dispatch end -C "+f.wt+" -change demo -key visual-verify ") {
		t.Fatalf("exit %d\n%s\n%s", rc, out, calls)
	}
}
