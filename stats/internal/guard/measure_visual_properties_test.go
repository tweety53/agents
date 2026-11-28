package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/smoke"
)

// TestMvpJSON: an mvpObj prints its keys in insertion order through encoding/json, and roundTo
// rounds half to even on the float's exact binary value.
func TestMvpJSON(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		x    float64
		n    int
		want string
	}{
		{0.125, 2, "0.12"},
		{2.675, 2, "2.67"},
		{-0.04, 1, "-0"},
		{56.0089, 1, "56"},
		{-2.5, 0, "-2"},
	} {
		if got := strconv.FormatFloat(roundTo(c.x, c.n), 'g', -1, 64); got != c.want {
			t.Errorf("roundTo(%v, %d) = %s, want %s", c.x, c.n, got, c.want)
		}
	}
	for x, want := range map[float64]int{0.5: 0, 1.5: 2, 2.5: 2, -2.5: -2} {
		if got := int(roundTo(x, 0)); got != want {
			t.Errorf("int(roundTo(%v, 0)) = %d, want %d", x, got, want)
		}
	}
	obj := mvpObj{
		{"b", 1},
		{"a", []mvpObj{}},
		{"c", mvpObj{}},
		{"d", nil},
		{"e", []mvpObj{{{"x", 1.0}, {"y", "#00ff00"}}}},
		{"f", math.Copysign(0, -1)},
	}
	want := "{\n  \"b\": 1,\n  \"a\": [],\n  \"c\": {},\n  \"d\": null,\n  \"e\": [\n    {\n      \"x\": 1,\n      \"y\": \"#00ff00\"\n    }\n  ],\n  \"f\": -0\n}"
	if got, err := json.MarshalIndent(obj, "", "  "); err != nil || string(got) != want {
		t.Errorf("MarshalIndent(obj) = %s, %v\nwant\n%s", got, err, want)
	}
}

const mvpTestdata = "testdata/measure-visual-properties"

// mvpPNG writes a w×h PNG, each pixel from px, to a temporary file and
// returns its path.
func mvpPNG(t *testing.T, w, h int, px func(x, y int) color.RGBA) string {
	t.Helper()
	im := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			im.Set(x, y, px(x, y))
		}
	}
	path := filepath.Join(t.TempDir(), "im.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, im); err != nil {
		t.Fatal(err)
	}
	return path
}

// mvpSoft is step n (1–3, outward) of a soft shadow: each step under the
// default --edge from its neighbour, so no step reads as the control's edge.
func mvpSoft(n int) color.RGBA {
	v := uint8(0xff - 13*(4-n))
	return color.RGBA{v, v, v, 0xff}
}

// mvpRun runs the guard in-process with dir as its working directory.
func mvpRun(t *testing.T, dir string, args ...string) (rc int, stdout, stderr string) {
	t.Helper()
	fn := Registry["measure-visual-properties"]
	if fn == nil {
		t.Fatal("measure-visual-properties is not registered")
	}
	var out, errb bytes.Buffer
	rc = fn(args, core.Env{Getenv: os.Getenv, Dir: dir}, &out, &errb)
	return rc, out.String(), errb.String()
}

// mvpJSON runs the guard, requires exit 0 and decodes its stdout.
func mvpJSON(t *testing.T, dir string, args ...string) map[string]any {
	t.Helper()
	rc, stdout, stderr := mvpRun(t, dir, args...)
	if rc != 0 {
		t.Fatalf("exit %d, stderr %q", rc, stderr)
	}
	var d map[string]any
	if err := json.Unmarshal([]byte(stdout), &d); err != nil {
		t.Fatalf("stdout is not JSON: %v", err)
	}
	return d
}

// mvpDig walks decoded JSON by object key (string) or list index (int).
func mvpDig(v any, path ...any) any {
	for _, p := range path {
		switch k := p.(type) {
		case string:
			v = v.(map[string]any)[k]
		case int:
			v = v.([]any)[k]
		}
	}
	return v
}

func mvpList(v any) []any { l, _ := v.([]any); return l }

// mvpSp prints its operands space-separated, strings included (fmt.Sprint
// adds no space beside a string operand).
func mvpSp(v ...any) string { return strings.TrimSuffix(fmt.Sprintln(v...), "\n") }

// TestMeasureVisualProperties: every case of
// scripts/test-measure-visual-properties.sh at 3915fbc0 as a subtest named
// after its ok: label, each asserting the harness's own reading of the JSON;
// then every golden the live Python wrote under testdata/ replayed byte for
// byte (argv, exit code, stdout, stderr); then the argparse surface
// exact-option-names changes, where the Python's stderr is not the port's.
func TestMeasureVisualProperties(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	dir, err := filepath.Abs(mvpTestdata)
	if err != nil {
		t.Fatal(err)
	}
	pair := func(a, b, props string) []string { return []string{a, b, "--scale", "1", "--props", props} }
	runsAfterBorder := func(t *testing.T, image string) string {
		runs := mvpList(mvpDig(mvpJSON(t, dir, image, "--region-a", "10,10,420,120", "--props", "runs"), "a", "runs", "col"))
		for i, r := range runs {
			if mvpDig(r, "colour") == "#d7d3d3" {
				return fmt.Sprint(mvpDig(runs[i+1], "length"), " ", mvpDig(runs[i+1], "colour"))
			}
		}
		return "no border run"
	}
	harness := []struct {
		label string
		check func(t *testing.T) (got, want string)
	}{
		{"case 1: inset fill reads as a 10px sheet-coloured run after the border", func(t *testing.T) (string, string) {
			return runsAfterBorder(t, "inset.png"), "10 #eae9e9"
		}},
		{"case 2: flush fill is the run immediately after the border, the full row height", func(t *testing.T) (string, string) {
			return runsAfterBorder(t, "flush.png"), "44 #e9f8ff"
		}},
		{"case 3: runs measures a region with no edge", func(t *testing.T) (string, string) {
			rc, _, _ := mvpRun(t, dir, "flat.png", "--props", "runs")
			return strconv.Itoa(rc), "0"
		}},
		{"case 3: box still exits non-zero on the same region", func(t *testing.T) (string, string) {
			rc, _, _ := mvpRun(t, dir, "flat.png", "--props", "box")
			return strconv.FormatBool(rc != 0), "true"
		}},
		{"case 4: a two-image runs comparison has both lists and no runs delta", func(t *testing.T) (string, string) {
			d := mvpJSON(t, dir, "inset.png", "flush.png", "--region-a", "10,10,420,120", "--region-b", "10,10,420,120", "--scale", "1", "--props", "runs")
			runsDelta := false
			for k := range d["delta"].(map[string]any) {
				runsDelta = runsDelta || strings.HasPrefix(k, "runs")
			}
			return fmt.Sprint(runsDelta, mvpDig(d, "b", "runs") != nil), "false true"
		}},
		{"case 5: content.colour reads the glyph tint and the delta carries the distance between two tints", func(t *testing.T) (string, string) {
			d := mvpJSON(t, dir, pair("icon-blue.png", "icon-grey.png", "content")...)
			return fmt.Sprint(mvpDig(d, "a", "content", "colour"), " ", mvpDig(d, "b", "content", "colour"), " ",
				math.RoundToEven(mvpDig(d, "delta", "content.colour", "distance").(float64))), "#1e6fe0 #9e9e9e 152"
		}},
		{"case 6: ink reads a text run's position, cap-height and tint, and the delta carries the tint distance", func(t *testing.T) (string, string) {
			d := mvpJSON(t, dir, pair("text-grey.png", "text-blue.png", "ink")...)
			return fmt.Sprint(mvpDig(d, "a", "ink", "left"), " ", mvpDig(d, "a", "ink", "height"), " ", mvpDig(d, "a", "ink", "colour"), " ",
				mvpDig(d, "b", "ink", "colour"), " ", math.RoundToEven(mvpDig(d, "delta", "ink.colour", "distance").(float64))), "24 14 #9e9e9e #1e6fe0 152"
		}},
		{"case 7: bands lists the missing rules and header band, the halved row padding as since_pair and the button's grey border, and pairs the rows across a text-width difference", func(t *testing.T) (string, string) {
			d := mvpJSON(t, dir, pair("page-frame.png", "page-capture.png", "bands")...)
			s := mvpDig(d, "delta", "bands_summary")
			rules, header, button := 0, false, 0.0
			var gaps, since []any
			for _, p := range mvpList(mvpDig(d, "delta", "bands")) {
				switch mvpDig(p, "status") {
				case "missing":
					h, c := mvpDig(p, "a", "height"), mvpDig(p, "a", "colour")
					if h == 2.0 && c == "#d7d3d3" {
						rules++
					}
					header = header || h == 40.0 && c == "#eae9e9"
				case "paired":
					switch mvpDig(p, "a", "height") {
					case 16.0:
						gaps, since = append(gaps, mvpDig(p, "gap_above", "abs")), append(since, mvpDig(p, "since_pair", "abs"))
					case 48.0:
						if button == 0 {
							button = math.RoundToEven(mvpDig(p, "edge", "distance").(float64))
						}
					}
				}
			}
			return mvpSp(mvpDig(s, "paired"), mvpDig(s, "missing"), mvpDig(s, "extra"), rules, header, gaps[1:], since[1:], button),
				"4 4 0 3 true [2 2] [-12 -12] 230"
		}},
		{"case 8: seams lists the missing cell divider and the wrapped label's lines left-anchored where the frame centres them, reads no seam from a text row's stems, and pairs cells across a data width", func(t *testing.T) (string, string) {
			d := mvpJSON(t, dir, pair("control-frame.png", "control-capture.png", "seams")...)
			s := mvpDig(d, "delta", "seams_summary")
			var boxed []any
			for _, b := range mvpList(mvpDig(d, "a", "seams")) {
				boxed = append(boxed, mvpDig(b, "top"))
			}
			var band any
			for _, b := range mvpList(mvpDig(d, "delta", "seams")) {
				if band == nil && mvpDig(b, "status") == "paired" {
					band = b
				}
			}
			var missing []string
			for _, p := range mvpList(mvpDig(band, "seams")) {
				if mvpDig(p, "status") == "missing" {
					missing = append(missing, mvpSp(mvpDig(p, "a", "left"), mvpDig(p, "a", "width"), mvpDig(p, "a", "colour")))
				}
			}
			return mvpSp(boxed, mvpDig(s, "paired"), mvpDig(s, "missing"), mvpDig(s, "extra"), mvpDig(s, "bands_unpaired"), mvpDig(s, "lines"), missing, mvpCells(band)),
				"[60] 3 1 0 0 map[paired:2 unpaired:0] [120 1 #d7d3d3] [(2,1): [-11 -11] [-21 -21]]"
		}},
		{"case 9: touching border and fill seams bound no cell yet the cells past them still pair, and a control with no outer border is a band the frame boxes and the capture does not", func(t *testing.T) (string, string) {
			var got []string
			for _, capture := range []string{"filled-broken.png", "filled-unboxed.png"} {
				d := mvpJSON(t, dir, pair("filled-frame.png", capture, "seams")...)
				var statuses, cells []any
				for _, b := range mvpList(mvpDig(d, "delta", "seams")) {
					statuses = append(statuses, mvpDig(b, "status"))
					if mvpDig(b, "status") == "paired" {
						cells = append(cells, mvpCells(b))
					}
				}
				got = append(got, mvpSp(statuses, mvpDig(d, "delta", "seams_summary", "missing"), mvpDig(d, "delta", "seams_summary", "bands_unpaired"), cells))
			}
			return strings.Join(got, "\n"), "[paired] 1 0 [[(3,2): [-11 -11] [-21 -21]]]\n[missing] 0 1 []"
		}},
		{"case 11: a control running off the image's edge is measured on its other sides and reports the edge it runs past as clipped", func(t *testing.T) (string, string) {
			// A full-width bar with a 3px soft shadow below, and a card in the
			// image's top-left corner with one right and below.
			bar := mvpPNG(t, 200, 100, func(x, y int) color.RGBA {
				switch {
				case y >= 40 && y <= 59:
					return color.RGBA{0x33, 0x66, 0xcc, 0xff}
				case y >= 60 && y <= 62:
					return mvpSoft(y - 59)
				}
				return color.RGBA{0xff, 0xff, 0xff, 0xff}
			})
			card := mvpPNG(t, 200, 100, func(x, y int) color.RGBA {
				switch {
				case x <= 119 && y <= 39:
					return color.RGBA{0x33, 0x66, 0xcc, 0xff}
				case x <= 122 && y <= 42:
					return mvpSoft(max(x-119, y-39))
				}
				return color.RGBA{0xff, 0xff, 0xff, 0xff}
			})
			var got []string
			for _, c := range []struct{ image, region string }{{bar, "0,20,200,70"}, {card, "0,0,160,60"}} {
				d := mvpJSON(t, dir, c.image, "--region-a", c.region, "--props", "box,shadow,gap")
				sh := func(side string) any {
					if v, ok := mvpDig(d, "a", "shadow", side).(string); ok {
						return v
					}
					return mvpDig(d, "a", "shadow", side, "width")
				}
				got = append(got, mvpSp(mvpDig(d, "a", "clipped"), mvpDig(d, "a", "box", "width"), mvpDig(d, "a", "box", "height"),
					sh("left"), sh("right"), sh("top"), sh("bottom"), mvpDig(d, "a", "gap", "left")))
			}
			return strings.Join(got, "\n"), "left,right 200 20 clipped clipped 0 3 clipped\nleft,top 120 40 clipped 3 clipped 3 clipped"
		}},
		{"case 12: a control covering two of the crop's corners — a bar on the image's top edge, a full-height side panel — is measured with the other corners' colour as background", func(t *testing.T) (string, string) {
			blue, white := color.RGBA{0x33, 0x66, 0xcc, 0xff}, color.RGBA{0xff, 0xff, 0xff, 0xff}
			bar := mvpPNG(t, 200, 100, func(x, y int) color.RGBA {
				if y < 20 {
					return blue
				}
				return white
			})
			panel := mvpPNG(t, 200, 100, func(x, y int) color.RGBA {
				if x < 30 {
					return blue
				}
				return white
			})
			var got []string
			for _, c := range []struct{ image, region string }{{bar, "0,0,200,30"}, {panel, "0,0,45,100"}} {
				d := mvpJSON(t, dir, c.image, "--region-a", c.region, "--props", "box")
				got = append(got, mvpSp(mvpDig(d, "a", "clipped"), mvpDig(d, "a", "box", "width"), mvpDig(d, "a", "box", "height")))
			}
			return strings.Join(got, "\n"), "left,right,top 200 20\nleft,top,bottom 30 100"
		}},
		{"case 10: a small-caps caption whose flat tops ink most of its span is not a boxed band, while the bordered control under it still is", func(t *testing.T) (string, string) {
			d := mvpJSON(t, dir, pair("captioned-frame.png", "captioned-capture.png", "seams")...)
			var a, b []string
			for _, band := range mvpList(mvpDig(d, "a", "seams")) {
				a = append(a, fmt.Sprintf("(%v, %d)", mvpDig(band, "top"), len(mvpList(mvpDig(band, "seams")))))
			}
			for _, band := range mvpList(mvpDig(d, "b", "seams")) {
				b = append(b, fmt.Sprint(mvpDig(band, "top")))
			}
			return mvpSp(a, b, mvpDig(d, "delta", "seams_summary", "paired"), mvpDig(d, "delta", "seams_summary", "bands_unpaired")), "[(40, 3)] [] 0 1"
		}},
	}
	for _, c := range harness {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			if got, want := c.check(t); got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}

	goldens, err := filepath.Glob(filepath.Join(dir, "*.golden"))
	if err != nil || len(goldens) == 0 {
		t.Fatalf("no goldens under %s: %v", dir, err)
	}
	for _, g := range goldens {
		name := strings.TrimSuffix(filepath.Base(g), ".golden")
		t.Run("golden "+name, func(t *testing.T) {
			t.Parallel()
			args, rc, stdout, stderr := mvpGolden(t, dir, name)
			gotRC, gotOut, gotErr := mvpRun(t, dir, args...)
			if gotRC != rc || gotOut != stdout || gotErr != stderr {
				t.Fatalf("argv %q:\nexit %d, want %d\nstderr %q\nwant   %q\nstdout equal: %v", args, gotRC, rc, gotErr, stderr, gotOut == stdout)
			}
		})
	}

	// exact-option-names: argparse's prefix abbreviations are gone and a usage
	// error is one line under the guard's own name, carrying argparse's message.
	for _, c := range []struct {
		label string
		args  []string
		msg   string
	}{
		{"an abbreviated option is unrecognized", []string{"card-frame.png", "card-capture.png", "--sc", "1"}, "unrecognized arguments: --sc 1"},
		{"a float option refuses a word", []string{"card-frame.png", "--edge", "x"}, "argument --edge: invalid float value: 'x'"},
		{"a float option refuses the hex form float() refuses", []string{"card-frame.png", "--edge", "0x1p3"}, "argument --edge: invalid float value: '0x1p3'"},
		{"an option with no value at the end", []string{"card-frame.png", "--scale"}, "argument --scale: expected one argument"},
		{"an option whose value is --", []string{"card-frame.png", "--scale", "--", "2"}, "argument --scale: expected one argument"},
		{"an option whose value looks like an option", []string{"card-frame.png", "--edge", "-inf"}, "argument --edge: expected one argument"},
		{"help with an explicit argument", []string{"--help=x"}, "argument -h/--help: ignored explicit argument 'x'"},
		{"no image", []string{"--props", "box"}, "the following arguments are required: image_a"},
		{"a third image", []string{"a.png", "b.png", "c.png"}, "unrecognized arguments: c.png"},
		{"an unknown option", []string{"card-frame.png", "--bogus"}, "unrecognized arguments: --bogus"},
		{"a box of three numbers", []string{"card-frame.png", "--region-a", "1,2,3"}, "argument --region-a: expected x,y,w,h of non-negative integers, got '1,2,3'"},
		{"a box of zero width", []string{"card-frame.png", "card-capture.png", "--ref-a=1,2,0,4"}, "argument --ref-a: width and height must be positive, got '1,2,0,4'"},
		{"a box with a non-ASCII digit", []string{"card-frame.png", "--region-b", "١,2,3,4"}, "argument --region-b: expected x,y,w,h of non-negative integers, got '١,2,3,4'"},
		{"a bad value before help", []string{"--scale", "x", "-h"}, "argument --scale: invalid float value: 'x'"},
	} {
		t.Run("usage: "+c.label, func(t *testing.T) {
			t.Parallel()
			rc, stdout, stderr := mvpRun(t, dir, c.args...)
			if want := "measure-visual-properties: " + c.msg + "\n"; rc != 2 || stdout != "" || stderr != want {
				t.Fatalf("exit %d stdout %q stderr %q, want exit 2 and stderr %q", rc, stdout, stderr, want)
			}
		})
	}

	// One Python repr() for both image ports: a non-UTF-8 byte of a path is
	// the surrogate os.fsdecode made of it, `\udcXX`, in measure's and
	// compose's OSError text alike.
	t.Run("a non-UTF-8 path reprs as Python's surrogate in both ports", func(t *testing.T) {
		t.Parallel()
		const bad, repr = "x\xff.png", `'x\udcff.png'`
		_, _, got := mvpRun(t, t.TempDir(), bad)
		if want := "measure-visual-properties: unreadable image " + bad + ": [Errno 2] No such file or directory: " + repr + "\n"; got != want {
			t.Errorf("measure stderr %q, want %q", got, want)
		}
		var out, errb bytes.Buffer
		Registry["compose-mockup-frames"]([]string{bad, "root", "out"}, core.Env{Getenv: os.Getenv, Dir: t.TempDir(), Stdin: strings.NewReader("")}, &out, &errb)
		if got := errb.String(); !strings.HasSuffix(got, ": "+repr+"\n") {
			t.Errorf("compose stderr %q, want it to end with %q", got, repr)
		}
	})

	// A working directory reached through a symlink: the image is read through
	// it, and an unreadable one is named as given, as Python's open() did.
	t.Run("a working directory reached through a symlink", func(t *testing.T) {
		t.Parallel()
		link := filepath.Join(t.TempDir(), "link")
		if err := os.Symlink(dir, link); err != nil {
			t.Fatal(err)
		}
		args, _, want, _ := mvpGolden(t, dir, "all-single-card")
		if _, got, _ := mvpRun(t, link, args...); got != want {
			t.Errorf("stdout through the symlink differs from the golden")
		}
		_, _, got := mvpRun(t, link, "missing.png")
		if want := "measure-visual-properties: unreadable image missing.png: [Errno 2] No such file or directory: 'missing.png'\n"; got != want {
			t.Errorf("stderr %q, want %q", got, want)
		}
	})
}

// mvpGolden reads <name>.golden (argv as a JSON array, the exit code, then
// stdout) and <name>.stderr.
func mvpGolden(t *testing.T, dir, name string) (args []string, rc int, stdout, stderr string) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, name+".golden"))
	if err != nil {
		t.Fatal(err)
	}
	argv, rest, _ := strings.Cut(string(raw), "\n")
	code, stdout, _ := strings.Cut(rest, "\n")
	if err := json.Unmarshal([]byte(argv), &args); err != nil {
		t.Fatal(err)
	}
	if rc, err = strconv.Atoi(code); err != nil {
		t.Fatal(err)
	}
	errb, err := os.ReadFile(filepath.Join(dir, name+".stderr"))
	if err != nil {
		t.Fatal(err)
	}
	return args, rc, stdout, string(errb)
}

// mvpCells is the harness's cells reading of one paired delta.seams band:
// per cell pair "(a,b):" then each line's offset and left deltas.
func mvpCells(band any) string {
	var out []string
	for _, c := range mvpList(mvpDig(band, "cells")) {
		s := fmt.Sprintf("(%v,%v):", mvpDig(c, "a"), mvpDig(c, "b"))
		for _, l := range mvpList(mvpDig(c, "lines")) {
			s += fmt.Sprint(" ", []any{mvpDig(l, "offset", "abs"), mvpDig(l, "left", "abs")})
		}
		out = append(out, s)
	}
	return fmt.Sprint(out)
}

// TestMeasureVisualPropertiesElements pins kan-1015's self-review: the band
// read took each scrim-dimmed sheet frame as one band, so the verifier
// measured every matrix with a hand-written script. `elements` finds the app
// surface below the scrim and lists every ink run inside it.
func TestMeasureVisualPropertiesElements(t *testing.T) {
	t.Parallel()
	sheet := func(scrim bool, textTop int) func(x, y int) color.RGBA {
		return func(x, y int) color.RGBA {
			switch {
			case scrim && y < 80:
				return color.RGBA{0x55, 0x55, 0x55, 0xff}
			case y == 100:
				return color.RGBA{0xcc, 0xcc, 0xcc, 0xff}
			case x >= 20 && x < 60 && y >= textTop && y < textTop+10:
				return color.RGBA{0x22, 0x22, 0x22, 0xff}
			}
			return color.RGBA{0xff, 0xff, 0xff, 0xff}
		}
	}
	a, b := mvpPNG(t, 120, 200, sheet(true, 120)), mvpPNG(t, 120, 200, sheet(true, 124))
	d := mvpJSON(t, "", a, b, "--scale", "1", "--props", "elements")
	var got []string
	for _, side := range []string{"a", "b"} {
		got = append(got, mvpSp(side, mvpDig(d, side, "elements", "surface_top"), mvpDig(d, side, "elements", "background")))
		for _, e := range mvpList(mvpDig(d, side, "elements", "elements")) {
			got = append(got, mvpSp(mvpDig(e, "kind"), mvpDig(e, "top"), mvpDig(e, "height"), mvpDig(e, "gap"), mvpDig(e, "colour")))
		}
	}
	for _, p := range mvpList(mvpDig(d, "delta", "elements")) {
		got = append(got, mvpSp(mvpDig(p, "status"), mvpDig(p, "top", "abs"), mvpDig(p, "gap", "abs")))
	}
	s := mvpDig(d, "delta", "elements_summary")
	got = append(got, mvpSp(mvpDig(s, "paired"), mvpDig(s, "missing"), mvpDig(s, "extra")))
	want := []string{
		"a 80 #ffffff", "rule 100 1 20 #cccccc", "text/control 120 10 19 #222222",
		"b 80 #ffffff", "rule 100 1 20 #cccccc", "text/control 124 10 23 #222222",
		"paired 0 0", "paired 4 4", "2 0 0",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	c := mvpJSON(t, "", mvpPNG(t, 120, 200, sheet(false, 120)), "--props", "elements")
	if top := mvpDig(c, "a", "elements", "surface_top"); top != float64(0) {
		t.Fatalf("no scrim: surface_top %v, want 0", top)
	}
	if rc, _, stderr := mvpRun(t, "", a, "--props", "bands,elements"); rc != 0 {
		t.Fatalf("bands,elements: exit %d, stderr %q", rc, stderr)
	}
}

// kan-1041 self-review: on a dark-theme page the bottom row already stops
// the surface walk, so the whole image is the surface and its elements pair.
func TestMeasureVisualPropertiesElementsDarkPage(t *testing.T) {
	t.Parallel()
	page := mvpPNG(t, 120, 200, func(x, y int) color.RGBA {
		if x >= 20 && x < 60 && y >= 120 && y < 130 {
			return color.RGBA{0xee, 0xee, 0xee, 0xff}
		}
		return color.RGBA{0x12, 0x12, 0x12, 0xff}
	})
	d := mvpJSON(t, "", page, page, "--scale", "1", "--props", "elements")
	if top := mvpDig(d, "a", "elements", "surface_top"); top != float64(0) {
		t.Fatalf("dark page: surface_top %v, want 0", top)
	}
	if paired := mvpDig(d, "delta", "elements_summary", "paired"); paired != float64(1) {
		t.Fatalf("dark page: paired %v, want 1", paired)
	}
}

// A bare dark frame's bottom corners lie outside the device's rounded corner
// and are light: the surface walk stops at the first dark row above that
// 20-row strip, and the strip is not the app surface.
func TestMeasureVisualPropertiesElementsDarkPageLightCorners(t *testing.T) {
	t.Parallel()
	page := mvpPNG(t, 120, 200, func(x, y int) color.RGBA {
		if y >= 180 && x < 8 {
			return color.RGBA{0xe2, 0xdf, 0xdf, 0xff}
		}
		if x >= 20 && x < 60 && y >= 120 && y < 130 {
			return color.RGBA{0xee, 0xee, 0xee, 0xff}
		}
		return color.RGBA{0x12, 0x12, 0x12, 0xff}
	})
	d := mvpJSON(t, "", page, page, "--scale", "1", "--props", "elements")
	if top := mvpDig(d, "a", "elements", "surface_top"); top != float64(0) {
		t.Fatalf("dark page with light corners: surface_top %v, want 0", top)
	}
	if paired := mvpDig(d, "delta", "elements_summary", "paired"); paired != float64(1) {
		t.Fatalf("dark page with light corners: paired %v, want 1", paired)
	}
}

// A dark page whose card outnumbers the page itself: the card's shade must not
// be taken for the background, or a frame and its capture that differ in card
// size pick different backgrounds and pair no row (kan-1160 dk-NS9e).
func TestMeasureVisualPropertiesMatrixDarkCardOutnumbersPage(t *testing.T) {
	t.Parallel()
	page := func(cardTop int) string {
		return mvpPNG(t, 200, 300, func(x, y int) color.RGBA {
			switch {
			case x >= 20 && x < 80 && y >= 20 && y < 30:
				return color.RGBA{0xee, 0xee, 0xee, 0xff} // text on the page
			case y >= cardTop:
				return color.RGBA{0x1a, 0x1a, 0x1d, 0xff} // card
			}
			return color.RGBA{0, 0, 0, 0xff}
		})
	}
	d := mvpJSON(t, "", page(40), page(250), "--scale", "1", "--props", "matrix")
	if got := mvpDig(d, "summary", "paired"); got != float64(1) {
		t.Fatalf("dark page, card outnumbering it: paired %v, want 1", got)
	}
}

// kan-1017 (sr11): `--props matrix` fills each paired element's columns in
// one call, and a region running past the image's edge is clipped to it.
func TestMeasureVisualPropertiesMatrix(t *testing.T) {
	t.Parallel()
	page := func(dy int) func(x, y int) color.RGBA {
		return func(x, y int) color.RGBA {
			y -= dy
			switch {
			case x >= 20 && x < 80 && y >= 20 && y < 30:
				return color.RGBA{0x22, 0x22, 0x22, 0xff} // heading
			case x >= 20 && x <= 180 && y >= 60 && y <= 100:
				if x == 20 || x == 180 || y == 60 || y == 100 {
					return color.RGBA{0x88, 0x88, 0x88, 0xff} // border
				}
				return color.RGBA{0xf0, 0xf0, 0xf0, 0xff} // fill
			case x >= 10 && x < 190 && y == 140:
				return color.RGBA{0xcc, 0xcc, 0xcc, 0xff} // rule
			}
			return color.RGBA{0xff, 0xff, 0xff, 0xff}
		}
	}
	a, b := mvpPNG(t, 200, 300, page(0)), mvpPNG(t, 200, 300, page(2))
	d := mvpJSON(t, "", a, b, "--scale", "1", "--props", "matrix")
	var got []string
	for _, row := range mvpList(mvpDig(d, "matrix")) {
		kind := mvpDig(row, "kind")
		var filled any
		switch kind {
		case "text/control":
			filled = mvpDig(row, "measured", "a", "ink", "colour")
		case "fill band":
			filled = mvpDig(row, "measured", "a", "box", "width")
		case "rule":
			filled = mvpDig(row, "measured", "a", "runs", "row", 1, "colour")
		}
		got = append(got, mvpSp(mvpDig(row, "status"), kind, filled))
	}
	want := []string{"paired text/control #222222", "paired fill band 161", "paired rule #cccccc"}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}

	edge := mvpPNG(t, 100, 800, func(x, y int) color.RGBA {
		if y >= 795 && x >= 10 && x < 90 {
			return color.RGBA{0x20, 0x20, 0x20, 0xff}
		}
		return color.RGBA{0xff, 0xff, 0xff, 0xff}
	})
	c := mvpJSON(t, "", edge, "--region-a", "0,795,100,7", "--props", "bands")
	if got := mvpSp(mvpDig(c, "a", "clipped"), len(mvpList(mvpDig(c, "a", "bands")))); got != "bottom 1" {
		t.Fatalf("edge band: clipped/bands %q, want %q", got, "bottom 1")
	}
	if rc, _, stderr := mvpRun(t, "", edge, "--region-a", "0,800,100,7", "--props", "bands"); rc != 2 || !strings.Contains(stderr, "exceeds") {
		t.Fatalf("region starting outside: exit %d, stderr %q, want 2 naming exceeds", rc, stderr)
	}
}

// Decision matrix-row-class: every `--props matrix` row carries its class —
// a text/control element is a `control` only when its box columns resolve,
// and an unpaired or unresolved row is classed by that.
func TestMeasureVisualPropertiesMatrixClass(t *testing.T) {
	t.Parallel()
	page := func(only func(x, y int) bool) func(x, y int) color.RGBA {
		return func(x, y int) color.RGBA {
			switch {
			case x >= 20 && x < 80 && y >= 20 && y < 30:
				return color.RGBA{0x22, 0x22, 0x22, 0xff} // a solid block: its box resolves
			case (x >= 20 && x <= 22 || x >= 40 && x <= 42) && y >= 50 && y < 60:
				return color.RGBA{0x22, 0x22, 0x22, 0xff} // two strokes: the centre column misses both
			case x >= 20 && x <= 180 && y >= 80 && y <= 120:
				if x == 20 || x == 180 || y == 80 || y == 120 {
					return color.RGBA{0x88, 0x88, 0x88, 0xff}
				}
				return color.RGBA{0xf0, 0xf0, 0xf0, 0xff}
			case x >= 10 && x < 190 && y == 150:
				return color.RGBA{0xcc, 0xcc, 0xcc, 0xff}
			case (x >= 10 && x <= 90 || x >= 110 && x < 190) && y >= 180 && y < 200:
				return color.RGBA{0x44, 0x44, 0x44, 0xff} // a band split at its centre column: unresolved
			case only(x, y):
				return color.RGBA{0x22, 0x22, 0x22, 0xff}
			}
			return color.RGBA{0xff, 0xff, 0xff, 0xff}
		}
	}
	a := mvpPNG(t, 200, 400, page(func(x, y int) bool { return x >= 20 && x <= 60 && y >= 230 && y < 240 }))
	b := mvpPNG(t, 200, 400, page(func(x, y int) bool { return x >= 20 && x <= 60 && y >= 300 && y < 330 }))
	var got []string
	for _, row := range mvpList(mvpDig(mvpJSON(t, "", a, b, "--scale", "1", "--props", "matrix"), "matrix")) {
		got = append(got, mvpSp(mvpDig(row, "class")))
	}
	want := []string{"control", "text run", "fill band", "rule", "unresolved", "missing", "extra"}
	if strings.Join(got, ", ") != strings.Join(want, ", ") {
		t.Fatalf("classes %q, want %q", got, want)
	}
}

// kan-1017 (sr13): a band's colour is its darkest ink holding at least 1% of
// the band, not its modal colour, and `share` is the top two colours' share.
func TestMeasureVisualPropertiesBandColourDarkestInk(t *testing.T) {
	t.Parallel()
	im := mvpPNG(t, 100, 30, func(x, y int) color.RGBA {
		switch {
		case y < 10 || y >= 20 || x >= 71:
		case x < 60:
			return color.RGBA{0xe2, 0xe8, 0xf0, 0xff} // border grey, 60%
		case x < 70:
			return color.RGBA{0x02, 0x08, 0x17, 0xff} // text ink, 10%
		case y == 15:
			return color.RGBA{0x00, 0x00, 0x00, 0xff} // one anti-aliased pixel
		}
		return color.RGBA{0xff, 0xff, 0xff, 0xff}
	})
	d := mvpJSON(t, "", im, "--props", "bands")
	band := mvpDig(d, "a", "bands", 0)
	if got := mvpSp(mvpDig(band, "colour"), mvpDig(band, "share")); got != "#020817 0.999" {
		t.Fatalf("colour/share %q, want %q", got, "#020817 0.999")
	}
}

// A text ink at 1.5% of the band (a frame's 2.2% and its capture's 1.5% are the
// same drawn text) reads as the band's colour, not the card's.
func TestMeasureVisualPropertiesBandColourThinInk(t *testing.T) {
	t.Parallel()
	im := mvpPNG(t, 100, 30, func(x, y int) color.RGBA {
		if y == 10 && x < 14 && x >= 0 {
			return color.RGBA{0x00, 0x00, 0x00, 0xff} // 1.5% of 900
		}
		if y < 10 || y >= 20 || x >= 90 {
			return color.RGBA{0xff, 0xff, 0xff, 0xff}
		}
		return color.RGBA{0xf5, 0xf5, 0xf7, 0xff}
	})
	band := mvpDig(mvpJSON(t, "", im, "--props", "bands"), "a", "bands", 0)
	if got := mvpDig(band, "colour"); got != "#000000" {
		t.Fatalf("colour %v, want #000000", got)
	}
}

// A band where no colour holds 1% of its ink (a gradient) reports its modal
// colour, never a fabricated #000000.
func TestMeasureVisualPropertiesBandColourFallsBackToModal(t *testing.T) {
	t.Parallel()
	im := mvpPNG(t, 160, 30, func(x, y int) color.RGBA {
		switch {
		case y < 10 || y >= 20 || x >= 150:
		case x == 70 || x == 71:
			return color.RGBA{0x40, 0x40, 0x40, 0xff} // modal, 20 of 1500 pixels
		default:
			return color.RGBA{uint8(x), 0x80, 0x80, 0xff} // one colour per column
		}
		return color.RGBA{0xff, 0xff, 0xff, 0xff}
	})
	band := mvpDig(mvpJSON(t, "", im, "--props", "bands"), "a", "bands", 0)
	if got := mvpDig(band, "colour"); got != "#404040" {
		t.Fatalf("colour %v, want #404040", got)
	}
}

// `--props bands,seams,matrix` is the two calls `--props bands,seams` and
// `--props matrix` in one: every key either printed reads byte for byte as
// that call printed it, and a region still conflicts with matrix.
func TestMeasureVisualPropertiesMatrixCombined(t *testing.T) {
	t.Parallel()
	pair := []string{"page-frame.png", "page-capture.png", "--scale", "1", "--props"}
	keys := func(props string) map[string]json.RawMessage {
		rc, stdout, stderr := mvpRun(t, mvpTestdata, append(pair, props)...)
		var d map[string]json.RawMessage
		if rc != 0 || json.Unmarshal([]byte(stdout), &d) != nil {
			t.Fatalf("--props %s: exit %d, stdout %s, stderr %s", props, rc, stdout, stderr)
		}
		return d
	}
	got := keys("bands,seams,matrix")
	n := 0
	for _, props := range []string{"bands,seams", "matrix"} {
		for k, want := range keys(props) {
			n++
			if string(got[k]) != string(want) {
				t.Errorf("key %q of --props %s differs combined:\n%s\nwant\n%s", k, props, got[k], want)
			}
		}
	}
	if len(got) != n-1 { // `scale` is printed by both
		t.Errorf("combined keys %d, want %d", len(got), n-1)
	}
	rc, _, stderr := mvpRun(t, mvpTestdata, append(pair, "bands,matrix", "--region-a", "0,0,10,10")...)
	if rc != 2 || !strings.Contains(stderr, "whole images") {
		t.Fatalf("matrix with --region-a: exit %d, stderr %q, want the refusal", rc, stderr)
	}
}

// A text/control element at the image's edge carries one `clipped` key, not
// a second one from its fill-band box measurement.
func TestMeasureVisualPropertiesMatrixClippedOnce(t *testing.T) {
	t.Parallel()
	page := func(x, y int) color.RGBA {
		if x >= 20 && x < 80 && y >= 290 {
			return color.RGBA{0x22, 0x22, 0x22, 0xff}
		}
		return color.RGBA{0xff, 0xff, 0xff, 0xff}
	}
	a := mvpPNG(t, 200, 300, page)
	rc, stdout, stderr := mvpRun(t, "", a, a, "--scale", "1", "--props", "matrix")
	if rc != 0 || !strings.Contains(stdout, `"text/control"`) {
		t.Fatalf("exit %d, stdout %s, stderr %s", rc, stdout, stderr)
	}
	if n := strings.Count(stdout, `"clipped":`); n != 2 {
		t.Fatalf("%d clipped keys, want one per image:\n%s", n, stdout)
	}
}

// kan-1044 self-review (F1-e): a modal sheet under a flat scrim measures its
// bands inside the sheet, not as one scrim-plus-sheet band.
func TestMeasureVisualPropertiesBandsUnderScrim(t *testing.T) {
	t.Parallel()
	dimmed := mvpPNG(t, 400, 800, func(x, y int) color.RGBA {
		switch {
		case x < 40 || x >= 360 || y < 300:
			return color.RGBA{0x7f, 0x7f, 0x7f, 0xff} // scrim
		case x >= 80 && x < 320 && (y >= 350 && y < 370 || y >= 420 && y < 440):
			return color.RGBA{0x02, 0x08, 0x17, 0xff}
		}
		return color.RGBA{0xff, 0xff, 0xff, 0xff}
	})
	d := mvpJSON(t, "", dimmed, "--props", "bands")
	bands, _ := mvpDig(d, "a", "bands").([]any)
	if len(bands) != 2 {
		t.Fatalf("bands %v, want two text rows", mvpDig(d, "a", "bands"))
	}
	plain := mvpPNG(t, 320, 500, func(x, y int) color.RGBA { // the same sheet, alone
		if x >= 40 && x < 280 && (y >= 50 && y < 70 || y >= 120 && y < 140) {
			return color.RGBA{0x02, 0x08, 0x17, 0xff}
		}
		return color.RGBA{0xff, 0xff, 0xff, 0xff}
	})
	if got := mvpDig(mvpJSON(t, "", dimmed, plain, "--scale", "1", "--props", "bands"), "delta", "bands_summary", "paired"); got != float64(2) {
		t.Fatalf("paired %v, want 2", got)
	}
}

// over-list: `--props matrix`'s summary.over names, by 1-based row, every
// matrix row missing, extra, unresolved, or paired with a scaled length over
// mvpOverLength or a colour distance over mvpOverColour.
func TestMeasureVisualPropertiesMatrixOver(t *testing.T) {
	t.Parallel()
	page := func(wider int, extraRule bool) func(x, y int) color.RGBA {
		return func(x, y int) color.RGBA {
			switch {
			case x >= 20 && x < 80 && y >= 20 && y < 30:
				return color.RGBA{0x22, 0x22, 0x22, 0xff} // heading, unchanged
			case x >= 20 && x <= 180+wider && y >= 60 && y <= 100:
				if x == 20 || x == 180+wider || y == 60 || y == 100 {
					return color.RGBA{0x88, 0x88, 0x88, 0xff}
				}
				return color.RGBA{0xf0, 0xf0, 0xf0, 0xff} // fill band, 3px wider in b
			case x >= 10 && x < 190 && y == 140:
				return color.RGBA{0xcc, 0xcc, 0xcc, 0xff} // rule, unchanged
			case extraRule && x >= 10 && x < 190 && y == 220:
				return color.RGBA{0xcc, 0xcc, 0xcc, 0xff} // a rule only b draws
			}
			return color.RGBA{0xff, 0xff, 0xff, 0xff}
		}
	}
	a, b := mvpPNG(t, 200, 300, page(0, false)), mvpPNG(t, 200, 300, page(3, true))
	d := mvpJSON(t, "", a, b, "--scale", "1", "--props", "matrix")
	var got []string
	for _, o := range mvpList(mvpDig(d, "summary", "over")) {
		got = append(got, mvpSp(mvpDig(o, "row"), mvpDig(o, "status"), mvpDig(o, "keys")))
	}
	if want := []string{"2 paired box.right, box.width", "4 extra "}; strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("over:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}

	// A paired row whose only delta over 2 sits on an unscaled leaf, and one
	// within both tolerances, are absent; an unresolved row is named.
	within := mvpObj{{"a", 10}, {"a_scaled", 10.0}, {"b", 12}, {"abs", 2.0}, {"pct", 20.0}}
	rows := []mvpObj{
		{{"kind", "rule"}, {"status", "paired"}, {"gap", within}, {"colour", mvpObj{{"a", "#000000"}, {"b", "#030000"}, {"distance", 3.0}}},
			{"measured", mvpObj{{"delta", mvpObj{{"ink.ratio", mvpObj{{"a", 1}, {"a_scaled", 1}, {"b", 9}, {"abs", 8}, {"pct", 800.0}}}}}}}},
		{{"kind", "rule"}, {"status", "paired"}, {"unresolved", "no box"}},
		{{"kind", "rule"}, {"status", "paired"}, {"colour", mvpObj{{"a", "#000000"}, {"b", "#ffffff"}, {"distance", 441.7}}}},
	}
	got = nil
	synthetic, _ := mvpOver(rows, mvpOverLength)
	for _, o := range synthetic {
		b, _ := json.MarshalIndent(o, "", "  ")
		got = append(got, string(b))
	}
	if s := strings.Join(got, "|"); strings.Contains(s, `"row": 1`) || !strings.Contains(s, `"row": 2`) || !strings.Contains(s, `"keys": "unresolved"`) || !strings.Contains(s, `"row": 3`) || !strings.Contains(s, `"keys": "colour"`) {
		t.Fatalf("synthetic rows: %s", s)
	}
}

// TestMeasureVisualPropertiesBandsReadADarkPageByItsPerimeter: a page's
// background is its perimeter's colour, so a card larger than the black page
// around it is ink, not background (kan-1083 self-review, F18: the dk-SL1
// capture read as one band from top 0 on a #262629 background).
func TestMeasureVisualPropertiesBandsReadADarkPageByItsPerimeter(t *testing.T) {
	t.Parallel()
	dark := mvpPNG(t, 40, 60, func(x, y int) color.RGBA {
		switch {
		case x >= 4 && x <= 35 && y >= 5 && y <= 44:
			return color.RGBA{0x26, 0x26, 0x29, 0xff} // card, more pixels than the page around it
		case y >= 50 && y <= 55:
			return color.RGBA{0x1a, 0x1a, 0x1d, 0xff} // full-width strip
		}
		return color.RGBA{0, 0, 0, 0xff}
	})
	d := mvpJSON(t, "", dark, "--props", "bands")
	bands := mvpList(mvpDig(d, "a", "bands"))
	var tops []string
	strip := ""
	for _, b := range bands {
		tops = append(tops, mvpSp(mvpDig(b, "top")))
		if mvpDig(b, "top") == float64(50) {
			strip = mvpSp(mvpDig(b, "height"), mvpDig(b, "colour"))
		}
	}
	if got := mvpDig(d, "a", "background"); got != "#000000" || !slices.Contains(tops, "5") || strip != "6 #1a1a1d" {
		t.Fatalf("dark page: background %v, band tops %v, strip %q; want #000000, a band at 5, \"6 #1a1a1d\"", got, tops, strip)
	}

	light := mvpPNG(t, 40, 60, func(x, y int) color.RGBA {
		if x >= 10 && x < 30 && y >= 10 && y < 14 {
			return color.RGBA{0, 0, 0, 0xff}
		}
		return color.RGBA{0xff, 0xff, 0xff, 0xff}
	})
	d = mvpJSON(t, "", light, "--props", "bands")
	if got := mvpSp(mvpDig(d, "a", "background"), len(mvpList(mvpDig(d, "a", "bands"))), mvpDig(d, "a", "bands", 0, "top")); got != "#ffffff 1 10" {
		t.Fatalf("light page: background/bands/top %q, want %q", got, "#ffffff 1 10")
	}
}

// mvpOverRows is a matrix run's summary.over, one "<row> <status> <keys>" line each.
func mvpOverRows(t *testing.T, a, b string) []string {
	t.Helper()
	var got []string
	for _, o := range mvpList(mvpDig(mvpJSON(t, "", a, b, "--scale", "1", "--props", "matrix"), "summary", "over")) {
		got = append(got, mvpSp(mvpDig(o, "row"), mvpDig(o, "status"), mvpDig(o, "keys")))
	}
	return got
}

// A rail of same-size marks on one row: a mark whose grey moves by a few
// levels — past mvpOverColour, under the element threshold that keeps the
// lighter grey out of the row's ink — is over on its own segment key.
func TestMeasureVisualPropertiesMatrixSegmentColour(t *testing.T) {
	t.Parallel()
	rail := func(unreached color.RGBA) string {
		return mvpPNG(t, 400, 60, func(x, y int) color.RGBA {
			if y >= 20 && y < 28 {
				for k := range 4 {
					if l := 20 + 92*k; x >= l && x < l+80 {
						if k < 2 {
							return color.RGBA{0, 0, 0, 0xff}
						}
						return unreached
					}
				}
			}
			return color.RGBA{0xff, 0xff, 0xff, 0xff}
		})
	}
	frame := rail(color.RGBA{0xe6, 0xe6, 0xe9, 0xff})
	got := strings.Join(mvpOverRows(t, frame, rail(color.RGBA{0xee, 0xee, 0xf1, 0xff})), "\n")
	if !strings.Contains(got, "segment3.colour, segment4.colour") || strings.Contains(got, "segment1.") || strings.Contains(got, "segment2.") {
		t.Fatalf("over:\n%s\nwant segment3.colour, segment4.colour and no other segment", got)
	}
	// A rail's marks are no glyphs: one drawn a tenth larger is over on its
	// size, never on `scale`.
	sized := func(w, h, gap int) string {
		return mvpPNG(t, 460, 60, func(x, y int) color.RGBA {
			if y >= 20 && y < 20+h {
				for k := range 4 {
					if l := 20 + (w+gap)*k; x >= l && x < l+w {
						return color.RGBA{0, 0, 0, 0xff}
					}
				}
			}
			return color.RGBA{0xff, 0xff, 0xff, 0xff}
		})
	}
	if got := strings.Join(mvpOverRows(t, sized(80, 20, 12), sized(85, 21, 13)), "\n"); strings.Contains(got, "scale") {
		t.Fatalf("a larger rail is over on scale: %s", got)
	}
	if got := strings.Join(mvpOverRows(t, frame, frame), "\n"); strings.Contains(got, "segment") {
		t.Fatalf("an identical rail is over on a segment: %s", got)
	}
	// The same rail on a white strip above a card-filled page: its gaps are
	// the row's white, not the page's card grey, and still part the marks.
	carded := func(unreached color.RGBA) string {
		return mvpPNG(t, 400, 300, func(x, y int) color.RGBA {
			switch {
			case y >= 20 && y < 28:
				for k := range 4 {
					if l := 20 + 92*k; x >= l && x < l+80 {
						if k < 2 {
							return color.RGBA{0, 0, 0, 0xff}
						}
						return unreached
					}
				}
			case y >= 60:
				return color.RGBA{0xf4, 0xf4, 0xf6, 0xff}
			}
			return color.RGBA{0xff, 0xff, 0xff, 0xff}
		})
	}
	got = strings.Join(mvpOverRows(t, carded(color.RGBA{0xe6, 0xe6, 0xe9, 0xff}), carded(color.RGBA{0xee, 0xee, 0xf1, 0xff})), "\n")
	if !strings.Contains(got, "segment3.colour, segment4.colour") {
		t.Fatalf("carded page over:\n%s\nwant segment3.colour, segment4.colour", got)
	}
}

// A text row drawn larger on both axes — beside a badge the capture lacks, so
// the row's own width hides it — carries `scale` with its widest ink group's
// size; a row narrowed by its glyph shape, the height barely moving, does not.
func TestMeasureVisualPropertiesMatrixScale(t *testing.T) {
	t.Parallel()
	block := func(boxes ...[4]int) string {
		return mvpPNG(t, 400, 100, func(x, y int) color.RGBA {
			for _, b := range boxes {
				if x >= b[0] && x < b[0]+b[2] && y >= b[1] && y < b[1]+b[3] {
					return color.RGBA{0, 0, 0, 0xff}
				}
			}
			return color.RGBA{0xff, 0xff, 0xff, 0xff}
		})
	}
	frame := block([4]int{20, 30, 18, 17}, [4]int{72, 20, 270, 43})
	d := mvpJSON(t, "", frame, block([4]int{20, 20, 299, 46}), "--scale", "1", "--props", "matrix")
	s, _ := mvpDig(d, "matrix", 0, "scale").(map[string]any)
	if got := mvpSp(s["a"], s["b"]); got != "270x43 299x46" {
		t.Fatalf("scale %q, want %q", got, "270x43 299x46")
	}
	if keys, _ := mvpDig(d, "summary", "over", 0, "keys").(string); !strings.Contains(keys, "scale") {
		t.Fatalf("over keys %q lack scale", keys)
	}
	serif := mvpJSON(t, "", block([4]int{20, 20, 235, 48}), block([4]int{20, 20, 192, 44}), "--scale", "1", "--props", "matrix")
	if s := mvpDig(serif, "matrix", 0, "scale"); s != nil {
		t.Fatalf("a glyph-shape narrowing carries scale %v", s)
	}
}

// A frame paired by its shift leaves the rows above the moved block unpaired
// when their own offset is smaller: those rows pair by gap instead.
func TestMeasureVisualPropertiesMatrixShiftRepairsTheRunAbove(t *testing.T) {
	t.Parallel()
	page := func(rows ...[2]int) string {
		return mvpPNG(t, 200, 700, func(x, y int) color.RGBA {
			for _, r := range rows {
				if x >= 20 && x < 180 && y >= r[0] && y < r[0]+r[1] {
					return color.RGBA{0, 0, 0, 0xff}
				}
			}
			return color.RGBA{0xff, 0xff, 0xff, 0xff}
		})
	}
	a := [][2]int{{10, 10}, {50, 16}, {150, 43}}
	b := [][2]int{{10, 10}, {42, 8}, {121, 46}}
	for k := range 5 {
		a, b = append(a, [2]int{300 + 60*k, 20}), append(b, [2]int{251 + 60*k, 20})
	}
	d := mvpJSON(t, "", page(a...), page(b...), "--scale", "1", "--props", "matrix")
	if got := mvpSp(mvpDig(d, "summary", "paired"), mvpDig(d, "summary", "missing"), mvpDig(d, "summary", "extra")); got != "8 0 0" {
		t.Fatalf("paired/missing/extra %q, want %q", got, "8 0 0")
	}
}

// A device's rounded screen corners — a frame draws them, a capture does not —
// are chrome, not a band the capture lacks.
func TestMeasureVisualPropertiesBandsIgnoreDeviceCorners(t *testing.T) {
	t.Parallel()
	im := mvpPNG(t, 100, 200, func(x, y int) color.RGBA {
		switch {
		case y >= 196: // fading strip along the bottom edge, past the noise
			return color.RGBA{0xf0, 0xf0, 0xf0, 0xff}
		case y >= 180 && (x < 20-(y-180) || x >= 80+(y-180)):
			return color.RGBA{0xe6, 0xe6, 0xe9, 0xff} // corner arcs
		case y >= 50 && y < 60 && x >= 10 && x < 90:
			return color.RGBA{0, 0, 0, 0xff}
		}
		return color.RGBA{0xff, 0xff, 0xff, 0xff}
	})
	d := mvpJSON(t, "", im, "--props", "bands")
	if n := len(mvpList(mvpDig(d, "a", "bands"))); n != 1 {
		t.Fatalf("%d bands, want only the one at y 50", n)
	}
}
