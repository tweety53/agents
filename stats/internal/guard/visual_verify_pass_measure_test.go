package guard

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tweety53/agents/stats/internal/smoke"
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

// vvmProtoPNG is vvmPNG under a 52px status line the capture does not have.
func vvmProtoPNG(t *testing.T, path string, dy int) {
	t.Helper()
	im := image.NewRGBA(image.Rect(0, 0, 200, 352))
	px := vvmPage(dy + 52)
	for y := 0; y < 352; y++ {
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
	writeFile(t, filepath.Join(f.wt, ".flow/project.md"), "# project\n\n"+vvpSect(t, "| `mockups` | `design/frames` |\n"))
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
	smoke.Heavy(t)
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
		if p.Frame == "frame-a" && len(p.Bands) == 0 {
			t.Fatalf("%s: the bands the measure flagged under `delta` never reached the manifest", p.Frame)
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
	vvmProtoPNG(t, filepath.Join(f.changeRoot, "visual-verification/prototype/a-darwin.png"), 4)
	writeFile(t, filepath.Join(f.changeRoot, "visual-verification/prototype.mockups"), "frame none\na.png a-darwin\n")
	if rc, _, stderr := f.run(t, f.wt, "demo", "proto", "tests/visual/app.spec.ts", "--prototype"); rc != 0 {
		t.Fatalf("--prototype: exit %d, stderr %q", rc, stderr)
	}
	pm := f.manifest(t, "proto")
	if len(pm.Pairs) != 1 || pm.Pairs[0].Mockup != filepath.Join(f.changeRoot, "visual-verification/a.frame.png") {
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

// TestVisualVerifyPassMeasureRecordsItsSpec pins the pipeline-fix-82 fix: a
// measure that pairs nothing -- no `mockups` root declared, so no sidecar
// frame resolves -- still names the spec it ran over, and a later measure of
// another spec keeps it, so close-task can tell the measure ran.
func TestVisualVerifyPassMeasureRecordsItsSpec(t *testing.T) {
	t.Parallel()
	f := vvmFixture(t, "a.png frame-a\n")
	writeFile(t, filepath.Join(f.wt, ".flow/project.md"), "# project\n\n"+vvpSection)
	writeFile(t, filepath.Join(f.wt, "tests/visual/other.spec.ts"), "// spec\n")
	vvmPNG(t, filepath.Join(f.wt, "tests/visual/other.spec.ts-snapshots/c-darwin.png"), 60)
	for _, spec := range []string{"tests/visual/app.spec.ts", "tests/visual/other.spec.ts"} {
		if rc, _, stderr := f.run(t, f.wt, "demo", "task-1", spec); rc != 0 {
			t.Fatalf("%s: exit %d, stderr %q", spec, rc, stderr)
		}
	}
	b, _ := os.ReadFile(filepath.Join(f.wt, ".superpowers/sdd/visual-manifest-task-1.json"))
	var m struct{ Specs []string }
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	want := []string{filepath.Join(f.wt, "tests/visual/app.spec.ts"), filepath.Join(f.wt, "tests/visual/other.spec.ts")}
	if strings.Join(m.Specs, " ") != strings.Join(want, " ") {
		t.Fatalf("specs %q, want %q", m.Specs, want)
	}
}

func TestVisualVerifyPassMeasureUsage(t *testing.T) {
	t.Parallel()
	f := vvmFixture(t, "b.png frame-b\n")
	if rc, _, stderr := f.run(t, f.wt, "demo", "visual-verify"); rc != 2 || !strings.Contains(stderr, "usage") {
		t.Fatalf("missing spec: exit %d, stderr %q, want 2 with usage", rc, stderr)
	}
	// A spec path that resolves nowhere would leave the last run's pairs standing as this run's.
	if rc, _, stderr := f.run(t, f.wt, "demo", "visual-verify", "elsewhere/app.spec.ts"); rc != 2 || !strings.Contains(stderr, "does not exist") {
		t.Fatalf("absent spec: exit %d, stderr %q, want 2 naming it", rc, stderr)
	}
	writeFile(t, filepath.Join(f.wt, ".superpowers/sdd/visual-measure-k"), "a file, not a directory\n")
	if rc, _, stderr := f.run(t, f.wt, "demo", "k", "tests/visual/app.spec.ts"); rc != 2 {
		t.Fatalf("unwritable output directory: exit %d, stderr %q, want 2", rc, stderr)
	}
}

// Regression (kan-1057 self-review, F2): a per-spec --prototype measure over a
// map naming another spec's screenshot failed at compose and deleted that
// spec's matrix and band files.
func TestVisualVerifyPassMeasurePrototypeKeepsOtherSpec(t *testing.T) {
	t.Parallel()
	f := vvmFixture(t, "a.png frame-a\n")
	writeFile(t, filepath.Join(f.wt, "tests/visual/other.spec.ts"), "// spec\n")
	vvmPNG(t, filepath.Join(f.wt, "tests/visual/other.spec.ts-snapshots/c-darwin.png"), 0)
	vvmProtoPNG(t, filepath.Join(f.changeRoot, "visual-verification/prototype/a-darwin.png"), 4)
	vvmProtoPNG(t, filepath.Join(f.changeRoot, "visual-verification/prototype/c-darwin.png"), 0)
	sidecar := filepath.Join(f.changeRoot, "visual-verification/prototype.mockups")
	const body = "frame none\na.png a-darwin\nc.png c-darwin\n"
	writeFile(t, sidecar, body)
	kept := []string{
		filepath.Join(f.changeRoot, "visual-verification/matrix/c-darwin.json"),
		filepath.Join(f.wt, ".superpowers/sdd/visual-measure-proto/c-darwin.json"),
	}
	for _, k := range kept {
		writeFile(t, k, `{"keep":true}`)
	}
	rc, stdout, stderr := f.run(t, f.wt, "demo", "proto", "tests/visual/app.spec.ts", "--prototype")
	if rc != 0 {
		t.Fatalf("exit %d, stdout %q, stderr %q", rc, stdout, stderr)
	}
	m := f.manifest(t, "proto")
	if len(m.Pairs) != 1 || m.Pairs[0].Frame != "a-darwin" {
		t.Fatalf("pairs %+v, want exactly a-darwin", m.Pairs)
	}
	if e := m.Pairs[0].Exits["compose"]; e == nil || *e != 0 {
		t.Fatalf("a-darwin compose exit %v, want 0", e)
	}
	for _, k := range kept {
		if b, err := os.ReadFile(k); err != nil || string(b) != `{"keep":true}` {
			t.Fatalf("%s: %q %v, want it untouched", k, b, err)
		}
	}
	if b, _ := os.ReadFile(sidecar); string(b) != body {
		t.Fatalf("prototype.mockups rewritten: %q", b)
	}
}

// A --prototype measure with no prototype.mockups still fails: compose runs
// over the missing map and its exit 2 surfaces (kan-1057 self-review, R2).
func TestVisualVerifyPassMeasurePrototypeMissingMap(t *testing.T) {
	t.Parallel()
	f := vvmFixture(t, "a.png frame-a\n")
	vvmProtoPNG(t, filepath.Join(f.changeRoot, "visual-verification/prototype/a-darwin.png"), 4)
	rc, stdout, stderr := f.run(t, f.wt, "demo", "proto", "tests/visual/app.spec.ts", "--prototype")
	if rc == 0 {
		t.Fatalf("exit 0 with no prototype.mockups, stdout %q, stderr %q", stdout, stderr)
	}
}

// TestVisualVerifyPassMeasureOverFigures pins kan-1057's self-review, F20:
// each manifest over row carries its own matrix row's figures, so a report
// item can be checked against the row it names, never a neighbour's.
func TestVisualVerifyPassMeasureOverFigures(t *testing.T) {
	t.Parallel()
	f := vvmFixture(t, "a.png frame-a\n")
	if rc, stdout, stderr := f.run(t, f.wt, "demo", "visual-verify", "tests/visual/app.spec.ts"); rc != 0 {
		t.Fatalf("exit %d, stdout %q, stderr %q", rc, stdout, stderr)
	}
	var m struct {
		Pairs []struct {
			MatrixFile string `json:"matrix_file"`
			Over       []struct {
				Row     int                          `json:"row"`
				Status  string                       `json:"status"`
				Keys    string                       `json:"keys"`
				Figures map[string]map[string]string `json:"figures"`
			} `json:"over"`
		} `json:"pairs"`
	}
	body, _ := os.ReadFile(filepath.Join(f.wt, ".superpowers/sdd/visual-manifest-visual-verify.json"))
	if err := json.Unmarshal(body, &m); err != nil || len(m.Pairs) != 1 {
		t.Fatalf("manifest %s: %v", body, err)
	}
	var mx struct {
		Matrix []map[string]any `json:"matrix"`
	}
	mb, _ := os.ReadFile(m.Pairs[0].MatrixFile)
	if err := json.Unmarshal(mb, &mx); err != nil {
		t.Fatal(err)
	}
	str := func(v any) string {
		if n, ok := v.(float64); ok {
			return strconv.FormatFloat(n, 'f', -1, 64)
		}
		s, _ := v.(string)
		return s
	}
	keyed := 0
	for _, r := range m.Pairs[0].Over {
		var want []string
		for _, k := range strings.Split(r.Keys, ", ") {
			if k != "" && k != "unresolved" {
				want = append(want, k)
			}
		}
		if r.Status == "missing" || r.Status == "extra" || len(want) == 0 {
			if len(r.Figures) != 0 {
				t.Errorf("row %d (%s, keys %q): figures %v, want none", r.Row, r.Status, r.Keys, r.Figures)
			}
			continue
		}
		keyed++
		if len(r.Figures) != len(want) {
			t.Errorf("row %d: figures %v, want keys %v", r.Row, r.Figures, want)
		}
		row := mx.Matrix[r.Row-1]
		for _, k := range want {
			v, ok := row[k].(map[string]any)
			if !ok {
				meas, _ := row["measured"].(map[string]any)
				d, _ := meas["delta"].(map[string]any)
				v, _ = d[k].(map[string]any)
			}
			if got := r.Figures[k]; got["a"] != str(v["a"]) || got["b"] != str(v["b"]) || got["a"] == "" {
				t.Errorf("row %d %s: figures %v, want a=%v b=%v", r.Row, k, got, v["a"], v["b"])
			}
		}
	}
	if keyed == 0 {
		t.Fatalf("no paired over row with keys: %+v", m.Pairs[0].Over)
	}
}

// TestVisualVerifyPassMeasurePrintsFrameLines pins the verifier's one input
// per frame: a `frame:` line naming the composite and every over row with its
// figures, frame→capture, so no frame is judged by parsing the manifest.
func TestVisualVerifyPassMeasurePrintsFrameLines(t *testing.T) {
	t.Parallel()
	f := vvmFixture(t, "a.png frame-a\nb.png frame-b\n")
	rc, stdout, stderr := f.run(t, f.wt, "demo", "visual-verify", "tests/visual/app.spec.ts")
	if rc != 0 {
		t.Fatalf("exit %d, stdout %q, stderr %q", rc, stdout, stderr)
	}
	m := f.manifest(t, "visual-verify")
	lines := map[string]string{}
	for _, l := range strings.Split(stdout, "\n") {
		if rest, ok := strings.CutPrefix(l, "frame: "); ok {
			id, _, _ := strings.Cut(rest, " ")
			lines[id] = l
		}
	}
	for _, p := range m.Pairs {
		l, ok := lines[p.Frame]
		if !ok {
			t.Fatalf("no `frame:` line for %s in %q", p.Frame, stdout)
		}
		if !strings.HasPrefix(l, "frame: "+p.Frame+" — "+p.Composite+" diff="+p.Diff+" — over: ") {
			t.Fatalf("%s: line %q does not name its composite and diff", p.Frame, l)
		}
		if len(p.Over) == 0 {
			if !strings.HasSuffix(l, "— over: none") {
				t.Fatalf("%s: line %q, want over: none", p.Frame, l)
			}
			continue
		}
		for _, r := range p.Over {
			row := r.(map[string]any)
			want := "#" + strconv.FormatFloat(row["row"].(float64), 'f', -1, 64)
			figs, _ := row["figures"].(map[string]any)
			for k, v := range figs {
				ab := v.(map[string]any)
				if !strings.Contains(l, k+" "+ab["a"].(string)+"→"+ab["b"].(string)) {
					t.Fatalf("%s: line %q lacks %s's %s figures %v", p.Frame, l, want, k, ab)
				}
			}
			if !strings.Contains(l, want) {
				t.Fatalf("%s: line %q lacks over row %s", p.Frame, l, want)
			}
		}
	}
}

// A frame id with a directory (`kan-1083/SE1`) writes its bands file below the key's directory.
func TestVvpWriteOutCreatesAFrameIdsDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "visual-measure-k", "kan-1083", "SE1.json")
	(&vvp{stderr: io.Discard}).writeOut(path, "{}")
	if b, err := os.ReadFile(path); err != nil || string(b) != "{}" {
		t.Fatalf("bands file not written: %q %v", b, err)
	}
}

// TestVisualVerifyPassMeasureMergesSpecs pins decision f2-merge-by-spec
// (kan-1102 self-review, F2): a second spec's measure under the same key merges
// into the manifest rather than overwriting the first spec's pairs, and a
// re-measured spec replaces its own pairs rather than duplicating them.
func TestVisualVerifyPassMeasureMergesSpecs(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	f := vvmFixture(t, "a.png frame-a\n")
	writeFile(t, filepath.Join(f.wt, "tests/visual/other.spec.ts"), "// spec\n")
	writeFile(t, filepath.Join(f.wt, "tests/visual/other.spec.ts.mockups"), "frame none\nb.png frame-b\n")
	vvmPNG(t, filepath.Join(f.wt, "tests/visual/other.spec.ts-snapshots/b-darwin.png"), 30)
	specA, specB := filepath.Join(f.wt, "tests/visual/app.spec.ts"), filepath.Join(f.wt, "tests/visual/other.spec.ts")
	var captures []string
	read := func() map[string][]string {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(f.wt, ".superpowers/sdd/visual-manifest-k.json"))
		if err != nil {
			t.Fatal(err)
		}
		var m struct {
			Pairs    []struct{ Frame, Spec string }
			Captures []string
		}
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatal(err)
		}
		captures = m.Captures
		got := map[string][]string{}
		for _, p := range m.Pairs {
			got[p.Frame] = append(got[p.Frame], p.Spec)
		}
		return got
	}
	// Spec A's capture line stays in captures after spec B's call, once, ahead of B's.
	capturesKept := func(when string) {
		t.Helper()
		at := func(sub string) int {
			i, n := -1, 0
			for j, c := range captures {
				if strings.Contains(c, sub) {
					i, n = j, n+1
				}
			}
			if n != 1 {
				t.Fatalf("%s: captures %v carry %s %d times, want once", when, captures, sub, n)
			}
			return i
		}
		if a, b := at("app.spec.ts-snapshots/a-darwin.png"), at("other.spec.ts-snapshots/b-darwin.png"); a > b {
			t.Fatalf("%s: captures %v, want spec A's line ahead of spec B's", when, captures)
		}
	}
	for _, spec := range []string{"tests/visual/app.spec.ts", "tests/visual/other.spec.ts"} {
		if rc, stdout, stderr := f.run(t, f.wt, "demo", "k", spec); rc != 0 {
			t.Fatalf("%s: exit %d, stdout %q, stderr %q", spec, rc, stdout, stderr)
		}
	}
	got := read()
	if len(got["frame-a"]) != 1 || got["frame-a"][0] != specA || len(got["frame-b"]) != 1 || got["frame-b"][0] != specB {
		t.Fatalf("pairs by frame %v; want frame-a from %s and frame-b from %s", got, specA, specB)
	}
	capturesKept("after spec B")
	if rc, _, stderr := f.run(t, f.wt, "demo", "k", "tests/visual/app.spec.ts"); rc != 0 {
		t.Fatalf("re-measure: exit %d, stderr %q", rc, stderr)
	}
	if got := read(); len(got["frame-a"]) != 1 || len(got["frame-b"]) != 1 {
		t.Fatalf("after re-measuring spec A: pairs by frame %v; want each frame once", got)
	}
	capturesKept("after re-measuring spec A")
}

// TestVisualVerifyPassMeasureNamesChangedFrames pins decisions
// changed-frames-in-measure, changed-by-blob-hash and
// changed-line-names-round-base: with a round file for its key, measure
// prints, before MANIFEST:, the frames whose capture differs from its blob at
// the round base — a rewritten or an untracked one — with the base; with none
// it prints no changed: line.
func TestVisualVerifyPassMeasureNamesChangedFrames(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	f := vvmFixture(t, "a.png frame-a\nb.png frame-b\n")
	gitRun(t, f.wt, "init", "-q")
	gitRun(t, f.wt, "add", "-A")
	gitRun(t, f.wt, "commit", "-qm", "base")
	base := vgHead(t, f.wt)
	measure := func(when string) []string {
		t.Helper()
		rc, stdout, stderr := f.run(t, f.wt, "demo", "visual-verify-fix-1", "tests/visual/app.spec.ts")
		if rc != 0 {
			t.Fatalf("%s: exit %d, stdout %q, stderr %q", when, rc, stdout, stderr)
		}
		return strings.Split(stdout, "\n")
	}
	// changed is the changed: line, failing unless it comes before MANIFEST:.
	changed := func(when string) string {
		t.Helper()
		got := ""
		for _, l := range measure(when) {
			if strings.HasPrefix(l, "MANIFEST:") && got == "" {
				t.Fatalf("%s: MANIFEST: before any changed: line", when)
			}
			if strings.HasPrefix(l, "changed:") {
				got = l
			}
		}
		return got
	}
	for _, l := range measure("no round file") {
		if strings.HasPrefix(l, "changed:") {
			t.Fatalf("no round file: printed %q", l)
		}
	}
	writeFile(t, filepath.Join(f.wt, ".superpowers/sdd/visual-round-visual-verify-fix-1.txt"), base+" "+time.Now().UTC().Format(time.RFC3339Nano)+"\n")
	if got, want := changed("unchanged"), "changed: none — round base "+base; got != want {
		t.Fatalf("unchanged: got %q, want %q", got, want)
	}
	vvmPNG(t, filepath.Join(f.wt, "tests/visual/app.spec.ts-snapshots/a-darwin.png"), 8)
	if got, want := changed("rewritten"), "changed: frame-a — round base "+base; got != want {
		t.Fatalf("rewritten: got %q, want %q", got, want)
	}
	writeFile(t, filepath.Join(f.wt, "tests/visual/app.spec.ts.mockups"), "frame none\na.png frame-a\nb.png frame-b\nc.png frame-c\n")
	// dy 20: a capture identical to b-darwin.png fails check-identical-baselines.
	vvmPNG(t, filepath.Join(f.wt, "design/frames/frame-c.png"), 20)
	vvmPNG(t, filepath.Join(f.wt, "tests/visual/app.spec.ts-snapshots/c-darwin.png"), 20)
	if got, want := changed("untracked"), "changed: frame-a frame-c — round base "+base; got != want {
		t.Fatalf("untracked: got %q, want %q", got, want)
	}
}

// TestVvmCells pins decisions cells-in-measure, text-stretch-from-element-insets
// and content-cell-for-text: each matrix row's 11 columns are read from the
// leaves its class names, n/a exactly where vrNAColumns lists the column, and a
// missing or extra row's absent side and delta are null.
func TestVvmCells(t *testing.T) {
	var matrix []map[string]any
	if err := json.Unmarshal([]byte(`[
	{"kind": "text/control", "class": "text run", "status": "paired",
	 "a": {"top": 20, "height": 10, "left": 30, "width": 60, "right": 110, "colour": "#222222", "kind": "text/control", "gap": 20},
	 "b": {"top": 24, "height": 10, "left": 50, "width": 60, "right": 90, "colour": "#222222", "kind": "text/control", "gap": 24},
	 "gap": {"a": 20, "b": 24, "abs": 4}, "height": {"a": 10, "b": 10, "abs": 0}, "colour": {"a": "#222222", "b": "#222222", "distance": 0},
	 "measured": {"a": {"background": "#ffffff", "ink": {"left": 4, "top": 4, "width": 60, "height": 10, "colour": "#222222"}},
	              "b": {"background": "#ffffff", "ink": {"left": 4, "top": 4, "width": 60, "height": 9, "colour": "#333333"}},
	              "delta": {"ink.height": {"a": 10, "b": 9, "abs": -1}, "ink.colour": {"a": "#222222", "b": "#333333", "distance": 29.4}}}},
	{"kind": "text/control", "class": "control", "status": "paired",
	 "a": {"top": 50, "height": 40, "left": 20, "width": 160, "right": 20, "colour": "#888888", "kind": "text/control", "gap": 20},
	 "b": {"top": 54, "height": 40, "left": 20, "width": 160, "right": 20, "colour": "#888888", "kind": "text/control", "gap": 20},
	 "measured": {"a": {"background": "#ffffff", "clipped": "left", "box": {"left": 4, "top": 4, "width": 160, "height": 40},
	                    "border": {"left": "clipped", "top": 1, "colour": "#888888"}, "radius": {"top_left": "clipped", "top_right": 0},
	                    "gap": {"left": null, "right": 12}, "content": {"padding": {"left": 8, "right": 8}}},
	              "b": {"background": "#ffffff", "clipped": "left", "box": {"left": 4, "top": 4, "width": 158, "height": 40},
	                    "border": {"left": "clipped", "top": 1, "colour": "#888888"}, "radius": {"top_left": "clipped", "top_right": 0},
	                    "gap": {"left": null, "right": 14}, "content": {"padding": {"left": 8, "right": 6}}},
	              "delta": {"box.width": {"a": 160, "b": 158, "abs": -2}, "border.left": {"a": "clipped", "b": "clipped", "abs": null}}}},
	{"kind": "rule", "class": "rule", "status": "paired",
	 "a": {"top": 140, "height": 1, "left": 10, "width": 180, "right": 10, "colour": "#cccccc", "kind": "rule", "gap": 50},
	 "b": {"top": 144, "height": 1, "left": 10, "width": 170, "right": 20, "colour": "#cccccc", "kind": "rule", "gap": 50},
	 "height": {"a": 1, "b": 1, "abs": 0},
	 "measured": {"a": {"background": "#ffffff", "runs": {"row": []}}, "b": {"background": "#ffffff", "runs": {"row": []}}, "delta": {}}},
	{"kind": "fill band", "class": "unresolved", "status": "paired",
	 "a": {"top": 180, "height": 20, "left": 10, "width": 180, "right": 10, "colour": "#444444", "kind": "fill band", "gap": 39},
	 "b": {"top": 184, "height": 20, "left": 10, "width": 180, "right": 10, "colour": "#444444", "kind": "fill band", "gap": 39},
	 "unresolved": "no edge on the centre scan line from the top"},
	{"class": "missing", "status": "missing",
	 "a": {"top": 230, "height": 10, "left": 20, "width": 41, "right": 139, "colour": "#222222", "kind": "text/control", "gap": 30}, "b": null},
	{"class": "extra", "status": "extra", "a": null,
	 "b": {"top": 300, "height": 30, "left": 20, "width": 41, "right": 139, "colour": "#222222", "kind": "text/control", "gap": 96}}
	]`), &matrix); err != nil {
		t.Fatal(err)
	}
	cells := vvmCells(matrix)
	got := func(row int, col string) string {
		b, _ := json.Marshal(cells[row-1].(map[string]any)[col])
		return string(b)
	}
	for _, tc := range []struct {
		row  int
		col  string
		want string
	}{
		{1, "content", `"n/a — transcribed by sweep 5"`},
		{1, "border", `"n/a — does not apply to a text run row"`},
		{1, "colour", `{"a":{"ink.colour":"#222222"},"b":{"ink.colour":"#333333"},"delta":{"ink.colour":{"a":"#222222","b":"#333333","distance":29.4}}}`},
		{1, "stretch", `{"a":{"element.left":30,"element.right":110},"b":{"element.left":50,"element.right":90},"delta":{}}`},
		{1, "position", `{"a":{"element.gap":20,"element.left":30,"element.top":20},"b":{"element.gap":24,"element.left":50,"element.top":24},"delta":{"element.gap":{"a":20,"abs":4,"b":24}}}`},
		{2, "size", `{"a":{"box.height":40,"box.width":160},"b":{"box.height":40,"box.width":158},"delta":{"box.width":{"a":160,"abs":-2,"b":158}}}`},
		{2, "border", `{"a":{"border.colour":"#888888","border.left":"clipped","border.top":1},"b":{"border.colour":"#888888","border.left":"clipped","border.top":1},"delta":{"border.left":{"a":"clipped","abs":null,"b":"clipped"}}}`},
		{2, "alignment", `{"a":{"content.padding.left":8,"content.padding.right":8},"b":{"content.padding.left":8,"content.padding.right":6},"delta":{}}`},
		{3, "alignment", `"n/a — does not apply to a rule row"`},
		{3, "size", `{"a":{"element.height":1,"element.width":180},"b":{"element.height":1,"element.width":170},"delta":{"element.height":{"a":1,"abs":0,"b":1}}}`},
		{4, "kind", `{"a":{"element.kind":"fill band"},"b":{"element.kind":"fill band"},"delta":{}}`},
		{4, "size", `"unresolved — no edge on the centre scan line from the top"`},
		{5, "size", `{"a":{"element.height":10,"element.width":41},"b":null,"delta":null}`},
		{5, "radius", `"n/a — does not apply to a missing row"`},
		{6, "stretch", `{"a":null,"b":{"element.left":20,"element.right":139},"delta":null}`},
	} {
		if g := got(tc.row, tc.col); g != tc.want {
			t.Errorf("row %d %s = %s, want %s", tc.row, tc.col, g, tc.want)
		}
	}
	for i, c := range cells {
		e := c.(map[string]any)
		if e["row"] != i+1 || len(e) != len(vrColumns)+2 {
			t.Errorf("entry %d: row %v, %d keys, want row %d and %d keys", i, e["row"], len(e), i+1, len(vrColumns)+2)
		}
	}
}

// TestVisualVerifyPassMeasureCells pins decision cells-in-measure end to end:
// `measure` writes one cells entry per matrix row of each composed frame, and
// check-verify-report admits every one of them.
func TestVisualVerifyPassMeasureCells(t *testing.T) {
	t.Parallel()
	f := vvmFixture(t, "a.png frame-a\n")
	if rc, stdout, stderr := f.run(t, f.wt, "demo", "visual-verify", "tests/visual/app.spec.ts"); rc != 0 {
		t.Fatalf("exit %d, stdout %q, stderr %q", rc, stdout, stderr)
	}
	var m struct {
		Pairs []struct {
			Frame      string           `json:"frame"`
			MatrixFile string           `json:"matrix_file"`
			Cells      []map[string]any `json:"cells"`
		} `json:"pairs"`
	}
	body, _ := os.ReadFile(filepath.Join(f.wt, ".superpowers/sdd/visual-manifest-visual-verify.json"))
	if err := json.Unmarshal(body, &m); err != nil || len(m.Pairs) != 1 {
		t.Fatalf("manifest %s: %v", body, err)
	}
	p := m.Pairs[0]
	var mx struct {
		Matrix []map[string]any `json:"matrix"`
	}
	mb, _ := os.ReadFile(p.MatrixFile)
	if err := json.Unmarshal(mb, &mx); err != nil || len(mx.Matrix) == 0 {
		t.Fatalf("matrix %s: %v", mb, err)
	}
	if len(p.Cells) != len(mx.Matrix) {
		t.Fatalf("%d cells entries for %d matrix rows", len(p.Cells), len(mx.Matrix))
	}
	for i, c := range p.Cells {
		if c["row"] != float64(i+1) || c["class"] != mx.Matrix[i]["class"] {
			t.Errorf("entry %d: row %v class %v, want row %d class %v", i, c["row"], c["class"], i+1, mx.Matrix[i]["class"])
		}
	}
	if bad := vrCells(p.Frame, p.Cells); len(bad) != 0 {
		t.Errorf("vrCells = %q, want none", bad)
	}
}

// TestVisualVerifyPassMeasureFrames pins decisions manifest-per-shard and
// slice-by-frame-id (KAN-1121): `measure --frames` composes and measures only
// the slice's frames, through a filtered copy of the spec's sidecar that
// compose-mockup-frames accepts without --prototype, and leaves another
// slice's cropped frame, matrix and the sidecar itself untouched.
func TestVisualVerifyPassMeasureFrames(t *testing.T) {
	t.Parallel()
	f := vvmFixture(t, "a.png frame-a\nb.png frame-b\n")
	vv := filepath.Join(f.changeRoot, "visual-verification")
	kept := []string{filepath.Join(vv, "b.frame.png"), filepath.Join(vv, "matrix/frame-b.json")}
	for _, k := range kept {
		writeFile(t, k, "keep\n")
	}
	sidecar := filepath.Join(f.wt, "tests/visual/app.spec.ts.mockups")
	rc, stdout, stderr := f.run(t, f.wt, "demo", "visual-verify-s1of2", "tests/visual/app.spec.ts", "--frames", "frame-a")
	if rc != 0 {
		t.Fatalf("exit %d, stdout %q, stderr %q", rc, stdout, stderr)
	}
	m := f.manifest(t, "visual-verify-s1of2")
	if len(m.Pairs) != 1 || m.Pairs[0].Frame != "frame-a" {
		t.Fatalf("pairs %+v, want frame-a alone", m.Pairs)
	}
	if e := m.Pairs[0].Exits["compose"]; e == nil || *e != 0 {
		t.Fatalf("frame-a compose exit %v, want 0 from the filtered sidecar", e)
	}
	filtered, err := os.ReadFile(filepath.Join(f.wt, ".superpowers/sdd/visual-measure-visual-verify-s1of2/app.spec.ts.mockups"))
	if err != nil || string(filtered) != "frame none\na.png frame-a\n" {
		t.Fatalf("filtered sidecar %q (%v), want the frame line and frame-a's line", filtered, err)
	}
	for _, k := range kept {
		if b, err := os.ReadFile(k); err != nil || string(b) != "keep\n" {
			t.Fatalf("%s: %q %v, want it untouched", k, b, err)
		}
	}
	if b, _ := os.ReadFile(sidecar); string(b) != "frame none\na.png frame-a\nb.png frame-b\n" {
		t.Fatalf("the spec's sidecar was rewritten: %q", b)
	}
	if _, err := os.Stat(filepath.Join(vv, "b.png")); !os.IsNotExist(err) {
		t.Fatalf("frame-b was composed outside the slice (stat err %v)", err)
	}
	if rc, _, stderr := f.run(t, f.wt, "demo", "k", "tests/visual/app.spec.ts", "--frames"); rc != 2 || !strings.Contains(stderr, "--frames") {
		t.Fatalf("--frames with no value: exit %d, stderr %q, want 2 naming it", rc, stderr)
	}
}

// TestVisualVerifyPassMeasureManifestKeysPairsByPrototype pins the manifest's pair key
// (spec, frame, prototype): a plain run then a --prototype run over one spec keeps both
// kinds of pair, and re-measuring either kind replaces only that kind's pair.
func TestVisualVerifyPassMeasureManifestKeysPairsByPrototype(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	f := vvmFixture(t, "a.png frame-a\n")
	vvmProtoPNG(t, filepath.Join(f.changeRoot, "visual-verification/prototype/a-darwin.png"), 4)
	writeFile(t, filepath.Join(f.changeRoot, "visual-verification/prototype.mockups"), "frame none\na.png a-darwin\n")
	read := func() (plain, proto int) {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(f.wt, ".superpowers/sdd/visual-manifest-k.json"))
		if err != nil {
			t.Fatal(err)
		}
		var m struct {
			Pairs []struct {
				Frame     string
				Prototype bool
			}
			Steps map[string]any
		}
		if err := json.Unmarshal(b, &m); err != nil {
			t.Fatal(err)
		}
		for _, p := range m.Pairs {
			if p.Prototype {
				proto++
			} else {
				plain++
			}
		}
		if len(m.Steps) == 0 {
			t.Fatalf("manifest carries no steps: %s", b)
		}
		return plain, proto
	}
	measure := func(extra ...string) {
		t.Helper()
		args := append([]string{f.wt, "demo", "k", "tests/visual/app.spec.ts"}, extra...)
		if rc, stdout, stderr := f.run(t, args...); rc != 0 {
			t.Fatalf("measure %v: exit %d, stdout %q, stderr %q", extra, rc, stdout, stderr)
		}
	}
	measure()
	measure("--prototype")
	if plain, proto := read(); plain != 1 || proto != 1 {
		t.Fatalf("plain then --prototype: %d plain and %d prototype pairs, want 1 and 1", plain, proto)
	}
	measure()
	if plain, proto := read(); plain != 1 || proto != 1 {
		t.Fatalf("re-measured plain: %d plain and %d prototype pairs, want 1 and 1", plain, proto)
	}
	measure("--prototype")
	if plain, proto := read(); plain != 1 || proto != 1 {
		t.Fatalf("re-measured --prototype: %d plain and %d prototype pairs, want 1 and 1", plain, proto)
	}
}

// A prototype frame carries the 52px status line the capture lacks; measure crops it, so a
// correct capture compares at zero offset.
func TestVisualVerifyPassMeasurePrototypeCropsStatusLine(t *testing.T) {
	t.Parallel()
	f := vvmFixture(t, "a.png frame-a\n")
	vvmProtoPNG(t, filepath.Join(f.changeRoot, "visual-verification/prototype/a-darwin.png"), 4)
	writeFile(t, filepath.Join(f.changeRoot, "visual-verification/prototype.mockups"), "frame none\na.png a-darwin\n")
	if rc, _, stderr := f.run(t, f.wt, "demo", "proto", "tests/visual/app.spec.ts", "--prototype"); rc != 0 {
		t.Fatalf("--prototype: exit %d, stderr %q", rc, stderr)
	}
	pm := f.manifest(t, "proto")
	if len(pm.Pairs) != 1 || len(pm.Pairs[0].Bands)+len(pm.Pairs[0].Over) != 0 {
		t.Fatalf("a status-line-offset prototype flagged at nonzero offset: %+v", pm.Pairs)
	}
}
