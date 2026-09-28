package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestPyJSON pins pyjson.go against Python's own output. Each expected string
// is copied from the python3 command in the comment above its table.
func TestPyJSON(t *testing.T) {
	t.Parallel()
	// python3 -c 'import json
	// for v in [3.0, 0.1, 1e-05, 0.0001, 1e16, 1.5e16, -0.0, 123456789012345.0, 0.0, 1e15,
	//           9999999999999998.0, 9.999999999999999e-05, float("nan"), float("inf"), -float("inf")]:
	//     print(repr(v), json.dumps(v))'
	for _, c := range []struct {
		f          float64
		repr, json string
	}{
		{3.0, "3.0", "3.0"},
		{0.1, "0.1", "0.1"},
		{1e-05, "1e-05", "1e-05"},
		{0.0001, "0.0001", "0.0001"},
		{1e16, "1e+16", "1e+16"},
		{1.5e16, "1.5e+16", "1.5e+16"},
		{math.Copysign(0, -1), "-0.0", "-0.0"},
		{123456789012345.0, "123456789012345.0", "123456789012345.0"},
		{0, "0.0", "0.0"},
		{1e15, "1000000000000000.0", "1000000000000000.0"},
		{9999999999999998.0, "9999999999999998.0", "9999999999999998.0"},
		{9.999999999999999e-05, "9.999999999999999e-05", "9.999999999999999e-05"},
		{math.NaN(), "nan", "NaN"},
		{math.Inf(1), "inf", "Infinity"},
		{math.Inf(-1), "-inf", "-Infinity"},
	} {
		if got := pyFloat(c.f); got != c.repr {
			t.Errorf("pyFloat(%v) = %q, python repr says %q", c.f, got, c.repr)
		}
		if got := pyDump(c.f); got != c.json {
			t.Errorf("pyDump(%v) = %q, python json.dumps says %q", c.f, got, c.json)
		}
	}
	// python3 -c 'import json; print(json.dumps(0), json.dumps(-12))'
	if got := pyDump(0) + " " + pyDump(-12); got != "0 -12" {
		t.Errorf("pyDump of ints = %q, want %q", got, "0 -12")
	}
	// python3 -c 'print(repr(round(0.125, 2)), repr(round(2.675, 2)), repr(round(-0.04, 1)),
	//                   repr(round(56.0089, 1)), repr(round(-2.5, 0)))'
	for _, c := range []struct {
		x    float64
		n    int
		want string
	}{
		{0.125, 2, "0.12"},
		{2.675, 2, "2.67"},
		{-0.04, 1, "-0.0"},
		{56.0089, 1, "56.0"},
		{-2.5, 0, "-2.0"},
	} {
		if got := pyFloat(pyRound(c.x, c.n)); got != c.want {
			t.Errorf("pyRound(%v, %d) = %s, python round says %s", c.x, c.n, got, c.want)
		}
	}
	// python3 -c 'print(round(0.5), round(1.5), round(2.5), round(-2.5))' — ndigits omitted, an int
	for x, want := range map[float64]int{0.5: 0, 1.5: 2, 2.5: 2, -2.5: -2} {
		if got := int(pyRound(x, 0)); got != want {
			t.Errorf("int(pyRound(%v, 0)) = %d, python round says %d", x, got, want)
		}
	}
	// python3 -c 'import json; print(json.dumps({"b": 1, "a": [], "c": {}, "d": None,
	//     "e": [{"x": 1.0, "y": "#00ff00"}], "f": -0.0}, indent=2))'
	obj := pyObj{
		{"b", 1},
		{"a", []pyObj{}},
		{"c", pyObj{}},
		{"d", nil},
		{"e", []pyObj{{{"x", 1.0}, {"y", "#00ff00"}}}},
		{"f", math.Copysign(0, -1)},
	}
	want := "{\n  \"b\": 1,\n  \"a\": [],\n  \"c\": {},\n  \"d\": null,\n  \"e\": [\n    {\n      \"x\": 1.0,\n      \"y\": \"#00ff00\"\n    }\n  ],\n  \"f\": -0.0\n}"
	if got := pyDump(obj); got != want {
		t.Errorf("pyDump(obj) =\n%s\npython json.dumps(indent=2) says\n%s", got, want)
	}
	if got := pyDump([]pyObj(nil)); got != "[]" {
		t.Errorf("pyDump of an empty list = %q, want []", got)
	}
}

const mvpTestdata = "testdata/measure-visual-properties"

// mvpRun runs the guard in-process with dir as its working directory.
func mvpRun(t *testing.T, dir string, args ...string) (rc int, stdout, stderr string) {
	t.Helper()
	fn := Registry["measure-visual-properties"]
	if fn == nil {
		t.Fatal("measure-visual-properties is not registered")
	}
	var out, errb bytes.Buffer
	rc = fn(args, Env{Getenv: os.Getenv, Dir: dir}, &out, &errb)
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
		Registry["compose-mockup-frames"]([]string{bad, "root", "out"}, Env{Getenv: os.Getenv, Dir: t.TempDir(), Stdin: strings.NewReader("")}, &out, &errb)
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
