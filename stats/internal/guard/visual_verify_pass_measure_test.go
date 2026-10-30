package guard

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// vvmFix is a worktree with a capture spec, its captures under the declared
// `screenshots` root, a `<spec>.mockups` sidecar (`frame none`) and frames.
type vvmFix struct{ wt, changeRoot string }

// vvmPage draws a heading, a bordered fill band and a rule, dy px down.
func vvmPage(dy int) func(x, y int) color.RGBA {
	return func(x, y int) color.RGBA {
		y -= dy
		switch {
		case x >= 20 && x < 80 && y >= 20 && y < 30:
			return color.RGBA{0x22, 0x22, 0x22, 0xff}
		case x >= 20 && x <= 180 && y >= 60 && y <= 100:
			if x == 20 || x == 180 || y == 60 || y == 100 {
				return color.RGBA{0x88, 0x88, 0x88, 0xff}
			}
			return color.RGBA{0xf0, 0xf0, 0xf0, 0xff}
		case x >= 10 && x < 190 && y == 140:
			return color.RGBA{0xcc, 0xcc, 0xcc, 0xff}
		}
		return color.RGBA{0xff, 0xff, 0xff, 0xff}
	}
}

func vvmPNG(t *testing.T, path string, dy int) {
	t.Helper()
	im := image.NewRGBA(image.Rect(0, 0, 200, 300))
	px := vvmPage(dy)
	for y := 0; y < 300; y++ {
		for x := 0; x < 200; x++ {
			im.Set(x, y, px(x, y))
		}
	}
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	writeFile(t, path, b.String())
}

// vvmFixture: capture a.png sits 4px below frame-a (flagged), b.png matches
// frame-b exactly; sidecar is the sidecar body after its `frame none` line.
func vvmFixture(t *testing.T, sidecar string) vvmFix {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	f := vvmFix{wt: filepath.Join(root, "demo")}
	f.changeRoot = filepath.Join(f.wt, "spectre/changes/demo")
	writeFile(t, filepath.Join(f.wt, ".flow/project.md"), "# project\n\n"+vvpSection+"| `mockups` | `design/frames` |\n")
	writeFile(t, filepath.Join(f.changeRoot, "design.md"), "# design\n")
	writeFile(t, filepath.Join(f.wt, "tests/visual/app.spec.ts"), "// spec\n")
	writeFile(t, filepath.Join(f.wt, "tests/visual/app.spec.ts.mockups"), "frame none\n"+sidecar)
	vvmPNG(t, filepath.Join(f.wt, "tests/visual/app.spec.ts-snapshots/a-darwin.png"), 4)
	vvmPNG(t, filepath.Join(f.wt, "tests/visual/app.spec.ts-snapshots/b-darwin.png"), 30)
	vvmPNG(t, filepath.Join(f.wt, "design/frames/frame-a.png"), 0)
	vvmPNG(t, filepath.Join(f.wt, "design/frames/frame-b.png"), 30)
	return f
}

func (f vvmFix) run(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var out, errb bytes.Buffer
	rc := visualVerifyPass(append([]string{"measure"}, args...), Env{Getenv: os.Getenv, Dir: f.wt}, &out, &errb)
	return rc, out.String(), errb.String()
}

type vvmManifest struct {
	Pairs []struct {
		Frame, Capture, Mockup, Composite, Diff string
		Exits                                   map[string]*int
		Bands, Seams, Over                      []any
	}
}

func (f vvmFix) manifest(t *testing.T, key string) vvmManifest {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(f.wt, ".superpowers/sdd/visual-manifest-"+key+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var m vvmManifest
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("manifest is not JSON: %v", err)
	}
	return m
}

func TestVisualVerifyPassMeasureManifest(t *testing.T) {
	t.Parallel()
	f := vvmFixture(t, "a.png frame-a\nb.png frame-b\n")
	rc, stdout, stderr := f.run(t, f.wt, "demo", "visual-verify", "tests/visual/app.spec.ts")
	if rc != 0 {
		t.Fatalf("exit %d, stdout %q, stderr %q", rc, stdout, stderr)
	}
	m := f.manifest(t, "visual-verify")
	if len(m.Pairs) != 2 {
		t.Fatalf("pairs %d, want 2", len(m.Pairs))
	}
	for _, p := range m.Pairs {
		for _, step := range []string{"resolve", "identical", "compose", "bands", "matrix"} {
			if e := p.Exits[step]; e == nil || *e != 0 {
				t.Fatalf("%s: exit %s %v, want 0", p.Frame, step, e)
			}
		}
		for _, file := range []string{p.Capture, p.Mockup, p.Composite,
			filepath.Join(f.changeRoot, "visual-verification/matrix", p.Frame+".json"),
			filepath.Join(f.wt, ".superpowers/sdd/visual-measure-visual-verify", p.Frame+".json")} {
			if _, err := os.Stat(file); err != nil {
				t.Fatalf("%s: %v", p.Frame, err)
			}
		}
		flagged := len(p.Bands) + len(p.Over)
		if (p.Frame == "frame-a") != (flagged > 0) {
			t.Fatalf("%s: %d flagged bands, %d over rows; want flags on frame-a alone", p.Frame, len(p.Bands), len(p.Over))
		}
	}
	matrix, _ := os.ReadFile(filepath.Join(f.changeRoot, "visual-verification/matrix/frame-a.json"))
	if !strings.Contains(string(matrix), `"over"`) {
		t.Fatalf("matrix file carries no summary.over: %.200s", matrix)
	}

	// --prototype composes against <changeRoot>/visual-verification/prototype/.
	vvmPNG(t, filepath.Join(f.changeRoot, "visual-verification/prototype/a-darwin.png"), 4)
	writeFile(t, filepath.Join(f.changeRoot, "visual-verification/prototype.mockups"), "frame none\na.png a-darwin\n")
	if rc, _, stderr := f.run(t, f.wt, "demo", "proto", "tests/visual/app.spec.ts", "--prototype"); rc != 0 {
		t.Fatalf("--prototype: exit %d, stderr %q", rc, stderr)
	}
	pm := f.manifest(t, "proto")
	if len(pm.Pairs) != 1 || pm.Pairs[0].Mockup != filepath.Join(f.changeRoot, "visual-verification/prototype/a-darwin.png") {
		t.Fatalf("--prototype pairs %+v, want one against the prototype frame", pm.Pairs)
	}
	if len(pm.Pairs[0].Bands)+len(pm.Pairs[0].Over) != 0 {
		t.Fatalf("--prototype: an identical prototype flagged %+v", pm.Pairs[0])
	}
}

// Regression: one failed compose hides every other frame's measurement.
func TestVisualVerifyPassMeasureCarriesOn(t *testing.T) {
	t.Parallel()
	f := vvmFixture(t, "a.png frame-missing\nb.png frame-b\n")
	rc, _, stderr := f.run(t, f.wt, "demo", "visual-verify", "tests/visual/app.spec.ts")
	if rc != 1 {
		t.Fatalf("exit %d, want 1; stderr %q", rc, stderr)
	}
	m := f.manifest(t, "visual-verify")
	var sawFail, sawB bool
	for _, p := range m.Pairs {
		switch p.Frame {
		case "frame-missing":
			sawFail = p.Exits["compose"] != nil && *p.Exits["compose"] != 0
		case "frame-b":
			sawB = p.Exits["matrix"] != nil && *p.Exits["matrix"] == 0 && p.Exits["bands"] != nil && *p.Exits["bands"] == 0
		}
	}
	if !sawFail || !sawB {
		t.Fatalf("pairs %+v: want frame-missing's non-zero compose exit and frame-b measured", m.Pairs)
	}
}

func TestVisualVerifyPassMeasureUsage(t *testing.T) {
	t.Parallel()
	f := vvmFixture(t, "b.png frame-b\n")
	if rc, _, stderr := f.run(t, f.wt, "demo", "visual-verify"); rc != 2 || !strings.Contains(stderr, "usage") {
		t.Fatalf("missing spec: exit %d, stderr %q, want 2 with usage", rc, stderr)
	}
	writeFile(t, filepath.Join(f.wt, ".superpowers/sdd/visual-measure-k"), "a file, not a directory\n")
	if rc, _, stderr := f.run(t, f.wt, "demo", "k", "tests/visual/app.spec.ts"); rc != 2 {
		t.Fatalf("unwritable output directory: exit %d, stderr %q, want 2", rc, stderr)
	}
}
