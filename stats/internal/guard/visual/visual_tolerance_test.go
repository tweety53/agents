package visual

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/guardtest"
)

// The declared length tolerance decides `over` in code: a 2px offset is
// recorded as within tolerance and never flagged, 3px is flagged, a colour is
// flagged whatever the length tolerance, and a wider tolerance moves the line.
func TestMvpOverAppliesTheLengthTolerance(t *testing.T) {
	t.Parallel()
	length := func(abs float64) mvpObj {
		return mvpObj{{"a", 10}, {"a_scaled", 10.0}, {"b", 10 + abs}, {"abs", abs}, {"pct", 0.0}}
	}
	rows := []mvpObj{
		{{"kind", "rule"}, {"status", "paired"}, {"gap", length(2)}},
		{{"kind", "rule"}, {"status", "paired"}, {"gap", length(-3)}},
		{{"kind", "rule"}, {"status", "paired"}, {"colour", mvpObj{{"a", "#000000"}, {"b", "#0d0000"}, {"distance", 13.0}}}},
		{{"class", "missing"}, {"status", "missing"}},
	}
	rowsOf := func(list []mvpObj) string {
		var got []string
		for _, o := range list {
			got = append(got, fmt.Sprint(mvpDigObj(o, "row")))
		}
		return strings.Join(got, ",")
	}
	over, within := mvpOver(rows, mvpOverLength)
	if got := rowsOf(over); got != "2,3,4" {
		t.Fatalf("default tolerance: over rows %s, want 2,3,4", got)
	}
	if len(within) != 1 || mvpDigObj(within[0], "row") != 1 || mvpDigObj(within[0], "values") != "gap=2" {
		t.Fatalf("default tolerance: within %v, want row 1 with gap=2", within)
	}
	over, within = mvpOver(rows, 100)
	if got := rowsOf(over); got != "3,4" {
		t.Fatalf("tolerance 100: over rows %s, want the colour and the missing row alone", got)
	}
	if got := rowsOf(within); got != "1,2" {
		t.Fatalf("tolerance 100: within rows %s, want 1,2", got)
	}
}

func mvpDigObj(o mvpObj, key string) any {
	for _, kv := range o {
		if kv.k == key {
			return kv.v
		}
	}
	return nil
}

func TestVvToleranceCell(t *testing.T) {
	t.Parallel()
	for cell, want := range map[string]float64{"": 2, "2px at 2x": 2, "`4px at 2x`": 4, "1.5px at 3x": 1.5} {
		if got, ok := vvTolerance(cell); !ok || got != want {
			t.Fatalf("vvTolerance(%q) = %v, %v; want %v", cell, got, ok, want)
		}
	}
	for _, cell := range []string{"2", "2px", "2dp at 2x", "2px at 0x", "-2px at 2x"} {
		if _, ok := vvTolerance(cell); ok {
			t.Fatalf("vvTolerance(%q) parsed; want refused — the unit and scale are stated", cell)
		}
	}
}

// measure reads the project's `tolerance` row: a capture 2px off its frame
// passes and one 3px off flags by default, the row overrides the default, and
// what the tolerance let through is counted and listed with its value.
func TestVisualVerifyPassMeasureToleranceRow(t *testing.T) {
	t.Parallel()
	flagged := func(t *testing.T, dy int, row string) (int, string, vvmFix) {
		f := vvmFixture(t, "a.png frame-a\n")
		vvmPNG(t, filepath.Join(f.wt, "tests/visual/app.spec.ts-snapshots/a-darwin.png"), dy)
		guardtest.WriteFile(t, filepath.Join(f.wt, ".flow/project.md"), "# project\n\n"+vvpSect(t, "| `mockups` | `design/frames` |\n"+row))
		rc, stdout, stderr := f.run(t, f.wt, "demo", "visual-verify", "tests/visual/app.spec.ts")
		if rc != 0 {
			t.Fatalf("dy %d row %q: exit %d, stdout %q, stderr %q", dy, row, rc, stdout, stderr)
		}
		p := f.manifest(t, "visual-verify").Pairs[0]
		return len(p.Bands) + len(p.Seams) + len(p.Over), stdout, f
	}
	n, stdout, f := flagged(t, 2, "")
	if n != 0 {
		t.Fatalf("a 2px offset flagged %d rows under the default tolerance; want none", n)
	}
	appendix, _ := os.ReadFile(filepath.Join(f.changeRoot, "visual-verification/within-tolerance.md"))
	if strings.Contains(stdout, "within-tolerance: 0\n") || !strings.Contains(stdout, "within-tolerance: ") || !strings.Contains(string(appendix), "- frame-a #1: gap=2") {
		t.Fatalf("a 2px offset is not recorded: stdout %q, appendix %q", stdout, appendix)
	}
	if n, _, _ := flagged(t, 3, ""); n == 0 {
		t.Fatal("a 3px offset flagged nothing under the default tolerance")
	}
	if n, _, _ := flagged(t, 3, "| `tolerance` | `3px at 2x` |\n"); n != 0 {
		t.Fatalf("a 3px offset flagged %d rows under a declared 3px tolerance; want none", n)
	}
	if n, _, _ := flagged(t, 2, "| `tolerance` | `1px at 1x` |\n"); n == 0 {
		t.Fatal("a 2px offset flagged nothing under a declared 1px tolerance")
	}

	bad := vvmFixture(t, "a.png frame-a\n")
	guardtest.WriteFile(t, filepath.Join(bad.wt, ".flow/project.md"), "# project\n\n"+vvpSect(t, "| `mockups` | `design/frames` |\n| `tolerance` | `2` |\n"))
	if rc, _, stderr := bad.run(t, bad.wt, "demo", "visual-verify", "tests/visual/app.spec.ts"); rc != 2 || !strings.Contains(stderr, "tolerance") {
		t.Fatalf("a tolerance row with no unit: exit %d, stderr %q; want 2 naming the row", rc, stderr)
	}
}

// vvmShiftFixture is a frame of n paired rows all sitting dy px (±wobble) low,
// each over on a vertical key alone, then one row over on a colour and one
// missing row.
func vvmShiftFixture(n int, dy float64) (matrix []map[string]any, over []any) {
	for i := 0; i < n; i++ {
		top := float64(100 * i)
		matrix = append(matrix, map[string]any{"a": map[string]any{"top": top}, "b": map[string]any{"top": top + dy + float64(i%2)}})
		over = append(over, map[string]any{"row": float64(i + 1), "status": "paired", "keys": "gap, box.top"})
	}
	matrix = append(matrix, map[string]any{"a": map[string]any{"top": 9000.0}, "b": map[string]any{"top": 9000.0 + dy}}, map[string]any{"a": map[string]any{"top": 9100.0}})
	over = append(over, map[string]any{"row": float64(n + 1), "status": "paired", "keys": "gap, colour"}, map[string]any{"row": float64(n + 2), "status": "missing", "keys": ""})
	return matrix, over
}

// A constant offset over more than vvmShiftRows rows is one defect: the rows
// fold into the first of them and one shift record, and a row over on anything
// but a vertical position stays.
func TestVvmFoldShift(t *testing.T) {
	t.Parallel()
	matrix, over := vvmShiftFixture(vvmShiftRows+5, 55)
	shift := vvmFoldShift(matrix, &over, mvpOverLength)
	if shift == nil || shift.Rows != vvmShiftRows+5 || shift.Row != 1 || shift.Dy < 55 || shift.Dy > 56 {
		t.Fatalf("shift %+v, want %d rows at dy 55-56 standing on row 1", shift, vvmShiftRows+5)
	}
	var left []string
	for _, o := range over {
		left = append(left, fmt.Sprint(o.(map[string]any)["row"]))
	}
	if got := strings.Join(left, ","); got != fmt.Sprintf("1,%d,%d", vvmShiftRows+6, vvmShiftRows+7) {
		t.Fatalf("over rows left %s, want the shift's row 1, the colour row and the missing row", got)
	}

	// At the threshold with one paired row unmoved, and for offsets that
	// differ row by row, nothing folds.
	matrix, over = vvmShiftFixture(vvmShiftRows, 55)
	matrix = append(matrix, map[string]any{"a": map[string]any{"top": 9200.0}, "b": map[string]any{"top": 9200.0}})
	if shift := vvmFoldShift(matrix, &over, mvpOverLength); shift != nil || len(over) != vvmShiftRows+2 {
		t.Fatalf("%d rows folded (%+v, %d left); a shift needs more than %d", vvmShiftRows, shift, len(over), vvmShiftRows)
	}
	matrix, over = vvmShiftFixture(vvmShiftRows+5, 55)
	for i := range vvmShiftRows + 5 {
		matrix[i]["b"].(map[string]any)["top"] = float64(100*i + 10*i)
	}
	if shift := vvmFoldShift(matrix, &over, mvpOverLength); shift != nil {
		t.Fatalf("rows with differing offsets folded: %+v", shift)
	}
}

// A frame moved whole by a few px is one offset row however few rows it has;
// one paired row that did not move makes it per-element again. Every other
// row stays on over, and the length tolerance is still 2px.
func TestVvmFoldShiftUniformOffset(t *testing.T) {
	t.Parallel()
	if mvpOverLength != 2 {
		t.Fatalf("mvpOverLength %v, want 2 — folding never widens the tolerance", mvpOverLength)
	}
	if tol, _ := vvTolerance(""); tol != 2 {
		t.Fatalf("default tolerance %v, want 2", tol)
	}
	matrix, over := vvmShiftFixture(3, 4)
	shift := vvmFoldShift(matrix, &over, mvpOverLength)
	if shift == nil || !shift.Uniform || shift.Rows != 3 || shift.Row != 1 {
		t.Fatalf("shift %+v, want a uniform 3-row offset standing on row 1", shift)
	}
	var got []string
	for _, o := range over {
		got = append(got, fmt.Sprintf("%v", o.(map[string]any)["row"]))
	}
	if s := strings.Join(got, ","); s != "1,4,5" {
		t.Fatalf("over rows %s, want the offset row, the colour row and the pairing row", s)
	}

	matrix, over = vvmShiftFixture(3, 4)
	matrix = append(matrix, map[string]any{"a": map[string]any{"top": 9200.0}, "b": map[string]any{"top": 9200.0}})
	if shift := vvmFoldShift(matrix, &over, mvpOverLength); shift != nil || len(over) != 5 {
		t.Fatalf("a non-uniform shift folded (%+v, %d rows left); want every row kept", shift, len(over))
	}
}

// vvmRowsPNG paints dark rows on a white 400×1000 page, each `{top, height,
// left, width}`.
func vvmRowsPNG(t *testing.T, path string, rows [][4]int) {
	t.Helper()
	im := image.NewRGBA(image.Rect(0, 0, 400, 1000))
	draw.Draw(im, im.Bounds(), image.White, image.Point{}, draw.Src)
	for _, r := range rows {
		draw.Draw(im, image.Rect(r[2], r[0], r[2]+r[3], r[0]+r[1]), image.NewUniform(color.RGBA{0x22, 0x22, 0x22, 0xff}), image.Point{}, draw.Src)
	}
	var b bytes.Buffer
	if err := png.Encode(&b, im); err != nil {
		t.Fatal(err)
	}
	guardtest.WriteFile(t, path, b.String())
}

// kan-1155's WU10, reduced: the capture is the frame scrolled 54px — every
// row and rule 54px low — with a tick the frame lacks above the first row, the
// first two text lines touching where the frame parts them, and the bottom bar
// pinned where it was. Pairing by gap paired the pinned bar alone and listed
// every other row twice, missing and extra; measure names the shift once and
// pairs the rows under it.
func TestVisualVerifyPassMeasureNamesAWholeFrameShiftOnce(t *testing.T) {
	t.Parallel()
	rule := func(top int) [4]int { return [4]int{top, 2, 8, 384} }
	moved := [][4]int{rule(98), {128, 71, 38, 300}, rule(218), {272, 24, 38, 117}, {342, 43, 38, 300}, {388, 26, 38, 200},
		rule(434), {462, 71, 38, 300}, rule(554), {606, 21, 38, 82}, rule(700), {730, 40, 38, 300}, rule(800)}
	bar := [4]int{900, 60, 30, 300}
	frame := append([][4]int{{8, 43, 38, 300}, {52, 26, 38, 200}}, append(slices.Clone(moved), bar)...)
	capture := [][4]int{{7, 4, 115, 170}, {62, 70, 38, 300}}
	for _, r := range moved {
		capture = append(capture, [4]int{r[0] + 54, r[1], r[2], r[3]})
	}
	capture = append(capture, bar)

	f := vvmFixture(t, "s.png frame-s\n")
	vvmRowsPNG(t, filepath.Join(f.wt, "design/frames/frame-s.png"), frame)
	vvmRowsPNG(t, filepath.Join(f.wt, "tests/visual/app.spec.ts-snapshots/s-darwin.png"), capture)
	rc, stdout, stderr := f.run(t, f.wt, "demo", "visual-verify", "tests/visual/app.spec.ts")
	if rc != 0 {
		t.Fatalf("exit %d, stdout %q, stderr %q", rc, stdout, stderr)
	}
	if strings.Count(stdout, "shift: ") != 1 || !strings.Contains(stdout, "shift: dy=54 rows=14 frame=frame-s\n") {
		t.Fatalf("stdout %q, want one `shift: dy=54 rows=14 frame=frame-s` line", stdout)
	}
	unpaired := 0
	for _, o := range f.manifest(t, "visual-verify").Pairs[0].Over {
		if s := o.(map[string]any)["status"]; s == "missing" || s == "extra" {
			unpaired++
		}
	}
	// The frame's two parted lines, and the capture's tick and joined lines.
	if unpaired != 4 {
		t.Fatalf("%d missing or extra rows under the shift, want 4 — the rows that moved must pair", unpaired)
	}
}

// The colour line sits between what rendering does to one colour and two
// different greys, as measured on kan-1155's captures: 3.5 passes, 5.2 flags.
func TestMvpOverColourThreshold(t *testing.T) {
	t.Parallel()
	colour := func(a, b string, d float64) mvpObj {
		return mvpObj{{"kind", "rule"}, {"status", "paired"}, {"colour", mvpObj{{"a", a}, {"b", b}, {"distance", d}}}}
	}
	over, _ := mvpOver([]mvpObj{colour("#ebebeb", "#ededed", 3.5), colour("#e9e9ec", "#e6e6e9", 5.2)}, mvpOverLength)
	if len(over) != 1 || mvpDigObj(over[0], "row") != 2 {
		t.Fatalf("over %v, want row 2 alone — a 5.2 grey mismatch flags and 3.5 of rendering noise does not", over)
	}
}
