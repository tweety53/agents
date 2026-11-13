package guard

import (
	"bytes"
	"github.com/tweety53/agents/stats/internal/smoke"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// rfPage is a tiny design page: frame F1 with a light and a dark phone,
// frame L1 with a light phone only, and lazy frame Z1, which sets
// data-mig-filled before filling its panes (light and dark phone), as the
// design pages' _migLazy does.
const rfPage = `<!doctype html><html><body style="margin:0;background:#888">
<div id="F1" data-frame style="display:flex;gap:40px;padding:20px;width:820px">
<div style="width:390px;height:844px;border-radius:28px;background:#ffffff"></div>
<div style="width:390px;height:844px;border-radius:28px;background:#000000"></div>
</div>
<div id="L1" data-frame style="padding:20px;width:820px">
<div style="width:390px;height:844px;border-radius:28px;background:#ffffff"></div>
</div>
<div id="Z1" data-frame data-mig style="display:flex;gap:40px;padding:20px;width:820px">
<div data-mig-pane></div>
<div data-mig-pane></div>
</div>
<script>
document.getElementById('Z1').dataset.migFilled = '1';
setTimeout(() => {
  const ps = document.querySelectorAll('#Z1 [data-mig-pane]');
  ps[0].innerHTML = '<div style="width:390px;height:844px;border-radius:28px;background:#ffffff"></div>';
  ps[1].innerHTML = '<div style="width:390px;height:844px;border-radius:28px;background:#000000"></div>';
}, 1500);
</script>
</body></html>
`

// rfFixture is a design repo tagged `t1` carrying rfPage (plus a generator
// copy of it under _gen/, which must not count against uniqueness), and a
// project root, itself a git repo, whose `## visual verification` names
// that repo, docs/design/mockups and checkout as the regression checkout.
func rfFixture(t *testing.T, checkout string) (design, root string) {
	t.Helper()
	base := t.TempDir()
	design, root = filepath.Join(base, "design"), filepath.Join(base, "project")
	mkdir(t, design)
	gitRun(t, design, "init", "-q")
	writeFile(t, filepath.Join(design, "Page.dc.html"), rfPage)
	writeFile(t, filepath.Join(design, "_gen", "Page.dc.html"), rfPage)
	gitRun(t, design, "add", "-A")
	gitRun(t, design, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "-m", "page")
	gitRun(t, design, "tag", "t1")
	mkdir(t, filepath.Join(root, ".flow"))
	gitRun(t, root, "init", "-q")
	writeFile(t, filepath.Join(root, ".flow", "project.md"), "## visual verification\n\n| Setting | Value |\n|---------|-------|\n"+
		"| `regression checkout` | `"+checkout+"` |\n| `mockups` | `docs/design/mockups` |\n"+
		"| `design repo` | `"+design+" "+design+"` |\n")
	return design, root
}

func rfRun(t *testing.T, env Env, args ...string) (string, int) {
	t.Helper()
	var out bytes.Buffer
	rc := Registry["render-frames"](args, env, &out, &out)
	return out.String(), rc
}

func rfCite(t *testing.T, root, body string) string {
	t.Helper()
	p := filepath.Join(root, "design.md")
	writeFile(t, p, body)
	return p
}

func TestRenderFramesCannotAnswer(t *testing.T) {
	t.Parallel()
	env := vvEnv("/")
	t.Run("no design repo row", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		f := rfCite(t, root, "x")
		if out, rc := rfRun(t, env, root, f); rc != 2 || !strings.Contains(out, "`design repo`") {
			t.Errorf("rc=%d out=%q, want 2 naming `design repo`", rc, out)
		}
	})
	t.Run("clone missing", func(t *testing.T) {
		t.Parallel()
		design, root := rfFixture(t, "/nowhere")
		os.RemoveAll(design)
		f := rfCite(t, root, "docs/design/mockups/t1/F1.png")
		if out, rc := rfRun(t, env, root, f); rc != 2 || !strings.Contains(out, "missing or not a git checkout") {
			t.Errorf("rc=%d out=%q, want 2", rc, out)
		}
	})
	t.Run("tag names a committed sub-directory", func(t *testing.T) {
		t.Parallel()
		_, root := rfFixture(t, "/nowhere")
		writeFile(t, filepath.Join(root, "docs/design/mockups/t1/F1.png"), "png")
		gitRun(t, root, "add", "-A")
		f := rfCite(t, root, "docs/design/mockups/t1/F1.png")
		if out, rc := rfRun(t, env, root, f); rc != 2 || !strings.Contains(out, "committed sub-directory") {
			t.Errorf("rc=%d out=%q, want 2 naming the committed sub-directory", rc, out)
		}
	})
	t.Run("committed sub-directory of a tag the design repo lacks stands", func(t *testing.T) {
		t.Parallel()
		_, root := rfFixture(t, "/nowhere")
		writeFile(t, filepath.Join(root, "docs/design/mockups/t9/F1.png"), "png")
		gitRun(t, root, "add", "-A")
		f := rfCite(t, root, "docs/design/mockups/t9/F1.png")
		if out, rc := rfRun(t, env, root, f); rc != 0 {
			t.Errorf("rc=%d out=%q, want 0: the committed frames are the only source", rc, out)
		}
		f = rfCite(t, root, "docs/design/mockups/t9/F2.png")
		if out, rc := rfRun(t, env, root, f); rc != 2 || !strings.Contains(out, "missing from it") {
			t.Errorf("rc=%d out=%q, want 2 naming the missing frame", rc, out)
		}
	})
	t.Run("tag unfetchable", func(t *testing.T) {
		t.Parallel()
		_, root := rfFixture(t, "/nowhere")
		f := rfCite(t, root, "docs/design/mockups/t9/F1.png")
		if out, rc := rfRun(t, env, root, f); rc != 2 || !strings.Contains(out, "cannot be fetched") {
			t.Errorf("rc=%d out=%q, want 2", rc, out)
		}
	})
	t.Run("node absent", func(t *testing.T) {
		t.Parallel()
		_, root := rfFixture(t, "/nowhere")
		f := rfCite(t, root, "docs/design/mockups/t1/F1.png")
		gitDir := filepath.Dir(must(exec.LookPath("git")))
		noNode := Env{Getenv: func(k string) string {
			if k == "PATH" {
				return gitDir
			}
			return os.Getenv(k)
		}, Dir: "/"}
		if _, ok := lookPath(noNode, "node"); ok {
			t.Skip("node sits beside git on PATH, so PATH cannot drop it")
		}
		if out, rc := rfRun(t, noNode, root, f); rc != 2 || !strings.Contains(out, "node is not on PATH") {
			t.Errorf("rc=%d out=%q, want 2", rc, out)
		}
	})
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func TestRenderFramesFindings(t *testing.T) {
	t.Parallel()
	env := vvEnv("/")
	t.Run("an id not at its tag, from a sidecar, fetching the tag once", func(t *testing.T) {
		t.Parallel()
		design, root := rfFixture(t, "/nowhere")
		// The clone lacks t1; the declared remote (the fixture repo) has it.
		clone := filepath.Join(filepath.Dir(design), "clone")
		gitRun(t, filepath.Dir(design), "clone", "-q", "--no-tags", design, clone)
		cfg := filepath.Join(root, ".flow", "project.md")
		body, _ := os.ReadFile(cfg)
		writeFile(t, cfg, strings.Replace(string(body), "`"+design+" ", "`"+clone+" ", 1))
		sidecar := filepath.Join(root, "x.spec.ts.mockups")
		writeFile(t, sidecar, "# comment\nframe scale=2 status=26 border=1 bare\na.png t1/NOPE\n")
		out, rc := rfRun(t, env, root, sidecar)
		if rc != 1 || !strings.Contains(out, "RENDER-FINDING: t1/NOPE — no frame `NOPE` at the tag") || !strings.Contains(out, "FRAMES-MISSING: 1") {
			t.Errorf("rc=%d out=%q, want 1 naming t1/NOPE", rc, out)
		}
		if exec.Command("git", "-C", clone, "rev-parse", "-q", "--verify", "refs/tags/t1").Run() != nil {
			t.Error("the tag was not fetched into the clone")
		}
	})
	t.Run("an id not unique at its tag", func(t *testing.T) {
		t.Parallel()
		design, root := rfFixture(t, "/nowhere")
		writeFile(t, filepath.Join(design, "Copy.dc.html"), rfPage)
		gitRun(t, design, "add", "-A")
		gitRun(t, design, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-q", "-m", "copy")
		gitRun(t, design, "tag", "t2")
		f := rfCite(t, root, "see docs/design/mockups/t2/F1.png.")
		if out, rc := rfRun(t, env, root, f); rc != 1 || !strings.Contains(out, "is not unique at the tag (2 in Copy.dc.html, Page.dc.html)") {
			t.Errorf("rc=%d out=%q, want 1", rc, out)
		}
	})
	t.Run("already rendered references render nothing", func(t *testing.T) {
		t.Parallel()
		_, root := rfFixture(t, "/nowhere")
		writeFile(t, filepath.Join(root, "docs/design/mockups/t1/F1.png"), "png")
		f := rfCite(t, root, "docs/design/mockups/t1/F1.png and other/docs/design/mockups/t1/NOPE.png")
		if out, rc := rfRun(t, env, root, f); rc != 0 || out != "FRAMES-RENDERED: 0 rendered, 1 already present\n" {
			t.Errorf("rc=%d out=%q, want 0 with one present", rc, out)
		}
	})
}

// TestRenderFramesRenders runs the real renderer on stats/web's own
// @playwright/test: each phone of the fixture frame at 780×1688 (390×844 at
// DPR 2), the light one as <id>.png and the dark one as dk-<id>.png, a
// light-only frame with no dk-, and the directory ignoring itself.
func TestRenderFramesRenders(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	_, root := rfFixture(t, rfPlaywright(t))
	f := rfCite(t, root, "docs/design/mockups/t1/F1.png docs/design/mockups/t1/dk-F1.png docs/design/mockups/t1/L1.png")
	out, rc := rfRun(t, vvEnv("/"), root, f)
	if rc != 0 || out != "FRAMES-RENDERED: 3 rendered, 0 already present\n" {
		t.Fatalf("rc=%d out=%q, want 0 with three rendered", rc, out)
	}
	dir := filepath.Join(root, "docs/design/mockups/t1")
	for name, wantLight := range map[string]bool{"F1.png": true, "dk-F1.png": false, "L1.png": true} {
		fh, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		img, err := png.Decode(fh)
		fh.Close()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if b := img.Bounds(); b.Dx() != 780 || b.Dy() != 1688 {
			t.Errorf("%s is %dx%d, want 780x1688", name, b.Dx(), b.Dy())
		}
		r, _, _, _ := img.At(390, 844).RGBA()
		if light := r > 0x8000; light != wantLight {
			t.Errorf("%s centre red=%#x, want light=%v", name, r, wantLight)
		}
	}
	if isFile(filepath.Join(dir, "dk-L1.png")) {
		t.Error("a light-only frame got a dk- render")
	}
	if b, _ := os.ReadFile(filepath.Join(dir, ".gitignore")); string(b) != "*\n" {
		t.Errorf(".gitignore = %q, want \"*\\n\"", b)
	}
	if st, _ := capture(exec.Command("git", "-C", root, "status", "--porcelain", "--", "docs")); st != "" {
		t.Errorf("rendered frames show in git status: %q", st)
	}
}

// rfPlaywright returns stats/web, skipping the test when node, @playwright/test
// or its Chromium does not resolve from it.
func rfPlaywright(t *testing.T) string {
	t.Helper()
	web, err := filepath.Abs("../../web")
	if err != nil {
		t.Fatal(err)
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("SKIPPED, not run: node is not on PATH, so the renderer cannot be exercised")
	}
	probe := "const p=require(require.resolve('@playwright/test',{paths:[process.argv[1]]}));if(!require('fs').existsSync(p.chromium.executablePath()))throw new Error('no chromium')"
	if b, err := exec.Command(node, "-e", probe, web).CombinedOutput(); err != nil {
		t.Skipf("SKIPPED, not run: @playwright/test or its Chromium does not resolve from %s (run `make -s web-dist` and `npx playwright install chromium` there): %s", web, strings.TrimSpace(string(b)))
	}
	return web
}

// TestRenderFramesWaitsForLazyPanes: a lazy frame whose panes fill after its
// data-mig-filled flag still renders both phones (KAN-1087).
func TestRenderFramesWaitsForLazyPanes(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	_, root := rfFixture(t, rfPlaywright(t))
	f := rfCite(t, root, "docs/design/mockups/t1/Z1.png docs/design/mockups/t1/dk-Z1.png")
	out, rc := rfRun(t, vvEnv("/"), root, f)
	if rc != 0 || out != "FRAMES-RENDERED: 2 rendered, 0 already present\n" {
		t.Fatalf("rc=%d out=%q, want 0 with two rendered", rc, out)
	}
}
