package guard

import (
	"bytes"
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
)

// Every case of scripts/test-compose-mockup-frames.sh at 3915fbc0, one
// subtest per ok: label, each running the guard in-process over the harness's
// own fixtures, drawn once by its Pillow snippets and committed under
// testdata/compose-mockup-frames/<invocation>/ with the live Python's output
// on them beside as golden/ (python-goldens-as-testdata). Besides its label's
// own assertion, every subtest of an invocation with a golden requires the
// port's exit code, stdout (the golden run's output directory spelled
// <out>) and stderr to equal the Python's byte for byte, and every PNG it
// wrote to equal the Python's pixel for pixel -- a three-panel composite's
// difference panel against an any-channel mask of its own capture and frame
// panels instead (any-channel-diff-mask). Case 11 (Pillow absent) has no
// counterpart; case 12 runs a lone copy of the shim, the sibling it needs now
// being lib/flow-guard.sh. The rows after the harness's pin the any-channel
// mask, Python's text input (python-io-semantics) and the physical absolute
// path printed from a cwd reached through a symlink.

const cmfGeom = "scale=2 status=26 border=1"

// cmfInv is one run of the guard. dir is the fixture directory under
// testdata/compose-mockup-frames (the invocation's own name when empty);
// args nil means the harness's `map.mockups root <out>` plus geom when set.
type cmfInv struct {
	dir, stdin, geom string
	args             []string
	stdout           string // overrides the golden stdout when set
	viaSymlink       bool   // run from a symlinked copy, out dir relative
	shim             bool   // run a lone copy of scripts/compose-mockup-frames.sh
}

var cmfInvs = map[string]cmfInv{
	"case-01":  {stdin: "caps/j5-finish-dialog-darwin.png\n"},
	"case-02":  {stdin: "caps/k2-menu.png\n"},
	"case-03":  {stdin: "caps/ok-darwin.png\n"},
	"case-04":  {stdin: "caps/a-darwin.png\n"},
	"case-05":  {stdin: "\n"},
	"case-06":  {stdin: "\n"},
	"case-07":  {stdin: "caps/a-darwin.png\ncaps/a-linux.png\n"},
	"case-08":  {stdin: "caps/bad.png\n"},
	"case-09":  {args: []string{"a", "b"}},
	"case-10":  {args: []string{"map.mockups", "root/not-a-dir", "<out>"}},
	"case-12":  {shim: true},
	"case-13":  {stdin: "caps/bad-frame.png\n"},
	"case-14":  {stdin: "caps/a-darwin.png\n"},
	"case-15":  {stdin: "caps/a-darwin.png\n"},
	"case-16":  {stdin: "caps/bg-check-darwin.png\n"},
	"case-17":  {stdin: "caps/align-check-darwin.png\n"},
	"case-18":  {stdin: "caps/add-participant-finished-excluded-darwin.png\ncaps/add-participant-finished-excluded-added-darwin.png\n"},
	"case-19":  {stdin: "caps/a-bogusplatform.png\n"},
	"case-20":  {stdin: "caps/g1-dash.png\n", geom: cmfGeom},
	"case-21":  {stdin: "caps/t1.png\ncaps/t2.png\n", geom: cmfGeom},
	"case-22":  {stdin: "caps/m1.png\ncaps/m2.png\n", geom: cmfGeom},
	"case-23a": {stdin: "caps/d1.png\n", geom: cmfGeom},
	"case-23b": {stdin: "caps/d2.png\n", geom: cmfGeom},
	"case-24":  {stdin: "caps/n1.png\n"},
	"case-25a": {stdin: "caps/b1.png\n", geom: "scale=2 status=26"},
	"case-25b": {stdin: "caps/b1.png\n", geom: "scale=0 status=26 border=1"},
	"case-25c": {stdin: "caps/b1.png\n", geom: "scale=2 status=-1 border=1"},
	"case-25d": {stdin: "caps/b1.png\n", geom: "scale=two status=26 border=1"},
	"case-25e": {stdin: "caps/b1.png\n", geom: "scale=2 status=26 border=1 extra=9"},
	"case-25f": {stdin: "caps/b1.png\n", geom: "scale=2 scale=2 border=1"},
	"case-26":  {stdin: "caps/c1.png\n", geom: cmfGeom},
	"case-27":  {stdin: "caps/c1.png\n", geom: cmfGeom},
	"case-28":  {stdin: "caps/b1.png\n", geom: cmfGeom},
	// The Python's luma mask printed diff=0.0000 for this pair: one pixel
	// one level off in red, one in blue, 2 of 96x20.
	"any-channel":   {stdin: "caps/d1.png\n", geom: cmfGeom, stdout: "<out>/d1.png diff=0.0010\n"},
	"fs-separator":  {stdin: "caps/k2-menu.png\n"},
	"stdin-crlf":    {stdin: "caps/k2-menu.png\r\n"},
	"stdin-lone-cr": {stdin: "caps/p1.png\rcaps/p2.png\n"},
	"map-not-utf8":  {stdin: "caps/k2-menu.png\n"},
	"stdin-nul":     {stdin: "caps/k\x00.png\n"},
	"symlink-cwd":   {dir: "case-02", stdin: "caps/k2-menu.png\n", viaSymlink: true},
}

// cmfResult is one run's outcome; out is the output directory, physical, and
// golden every departure from the Python's golden output ("" when none).
type cmfResult struct {
	rc                  int
	stdout, stderr, out string
	golden              string
}

type cmfCheck func(r cmfResult) string

type cmfCase struct {
	label, inv string
	checks     []cmfCheck
}

var cmfCases = []cmfCase{
	{"case 1: happy path exits 0 with the composite path on stdout", "case-01", []cmfCheck{cmfRC(0), cmfOut("<out>/j5-finish-dialog.png diff=n/a")}},
	{"case 1: composite size is 1996x2028", "case-01", []cmfCheck{cmfSize("j5-finish-dialog.png", 1996, 2028)}},
	{"case 2: exact-name match with no platform suffix exits 0", "case-02", []cmfCheck{cmfRC(0)}},
	{"case 3: stderr carries 'no captured PNG matches'", "case-03", []cmfCheck{cmfErr(`no captured PNG matches`)}},
	{"case 3: exit 1", "case-03", []cmfCheck{cmfRC(1)}},
	{"case 3: the well-formed second line is still composed", "case-03", []cmfCheck{cmfOut("<out>/ok.png diff=n/a")}},
	{"case 4: exit 1", "case-04", []cmfCheck{cmfRC(1)}},
	{"case 4: stderr carries 'no frame'", "case-04", []cmfCheck{cmfErr(`no frame`)}},
	{"case 5: exit 1", "case-05", []cmfCheck{cmfRC(1)}},
	{"case 5: stderr carries 'expected'", "case-05", []cmfCheck{cmfErr(`expected`)}},
	{"case 6: comment/blank-only map exits 0 with no output", "case-06", []cmfCheck{cmfRC(0), cmfOut("")}},
	{"case 7: exit 1", "case-07", []cmfCheck{cmfRC(1)}},
	{"case 7: stderr names both paths", "case-07", []cmfCheck{cmfErr(`a-darwin\.png.*a-linux\.png|a-linux\.png.*a-darwin\.png`)}},
	{"case 8: unreadable PNG on stdin exits 2", "case-08", []cmfCheck{cmfRC(2)}},
	{"case 9: wrong argument count exits 2", "case-09", []cmfCheck{cmfRC(2)}},
	{"case 9: stderr carries the usage line", "case-09", []cmfCheck{cmfErr(`usage: compose-mockup-frames\.sh`)}},
	{"case 10: mockups root not a directory exits 2", "case-10", []cmfCheck{cmfRC(2)}},
	{"case 12: a shim copy with no lib/flow-guard.sh sibling exits 2", "case-12", []cmfCheck{cmfRC(2)}},
	{"case 12: the message names the missing module", "case-12", []cmfCheck{cmfErr(`lib/flow-guard\.sh`)}},
	{"case 13: unreadable mockup frame exits 2", "case-13", []cmfCheck{cmfRC(2)}},
	{"case 13: stderr names the unreadable frame", "case-13", []cmfCheck{cmfErr(`unreadable PNG.*X\.png`)}},
	{"case 14: duplicate screenshot name in map exits 1", "case-14", []cmfCheck{cmfRC(1)}},
	{"case 14: stderr names the duplicate", "case-14", []cmfCheck{cmfErr(`duplicate screenshot name.*a\.png`)}},
	{"case 15: screenshot name without .png suffix exits 1", "case-15", []cmfCheck{cmfRC(1)}},
	{"case 15: stderr names the missing .png suffix", "case-15", []cmfCheck{cmfErr(`does not end in \.png`)}},
	{"case 16: happy path for background check exits 0", "case-16", []cmfCheck{cmfRC(0)}},
	{"case 16: gutter pixel is (40, 40, 40)", "case-16", []cmfCheck{cmfPixel("bg-check.png", 100, 5, 40, 40, 40)}},
	{"case 17: happy path for alignment check exits 0", "case-17", []cmfCheck{cmfRC(0)}},
	{"case 17: mockup pixel at top edge is the mockup's color", "case-17", []cmfCheck{cmfPixel("align-check.png", 66, 0, 0, 255, 0)}},
	{"case 17: pixel below the top-aligned mockup is the background", "case-17", []cmfCheck{cmfPixel("align-check.png", 66, 199, 40, 40, 40)}},
	{"case 18: a shorter name is not confused with a longer name's own platform-suffixed capture", "case-18", []cmfCheck{cmfRC(0), cmfOut("<out>/add-participant-finished-excluded.png diff=n/a")}},
	{"case 18: the composed capture is the exact-stem one, not the longer-named one", "case-18", []cmfCheck{cmfPixel("add-participant-finished-excluded.png", 5, 5, 255, 0, 0)}},
	{"case 19: a suffix outside the closed platform vocabulary does not match", "case-19", []cmfCheck{cmfRC(1)}},
	{"case 19: stderr carries 'no captured PNG matches'", "case-19", []cmfCheck{cmfErr(`no captured PNG matches`)}},
	{"case 20: geometry declared exits 0", "case-20", []cmfCheck{cmfRC(0)}},
	{"case 20: composite is three 96px panels and two gutters", "case-20", []cmfCheck{cmfSize("g1-dash.png", 320, 60)}},
	{"case 20: the middle panel starts at the content area, not the status line", "case-20", []cmfCheck{cmfPixel("g1-dash.png", 112, 0, 255, 0, 0)}},
	{"case 20: the cropped frame is written as <stem>.frame.png at the capture's size", "case-20", []cmfCheck{cmfSize("g1-dash.frame.png", 96, 60)}},
	{"case 21: two frames of different heights both compose", "case-21", []cmfCheck{cmfRC(0)}},
	{"case 21: each frame's caption bottom is found for that frame", "case-21", []cmfCheck{cmfSize("t1.png", 320, 60), cmfSize("t2.png", 320, 40)}},
	{"case 22: a size mismatch exits 1", "case-22", []cmfCheck{cmfRC(1)}},
	{"case 22: the finding names both sizes and the frame id", "case-22", []cmfCheck{cmfErr(`capture 96×44 vs frame 96×60 for M1`)}},
	{"case 22: the mismatched pair is skipped, the sibling still composed", "case-22", []cmfCheck{cmfExists("m2.png", true), cmfExists("m1.png", false)}},
	{"case 23: an identical pair prints diff=0.0000", "case-23a", []cmfCheck{cmfRC(0), cmfOut("<out>/d1.png diff=0.0000")}},
	{"case 23: a fully inverted pair prints diff=1.0000", "case-23b", []cmfCheck{cmfRC(0), cmfOut("<out>/d2.png diff=1.0000")}},
	{"case 24: no geometry prints diff=n/a and exits 0", "case-24", []cmfCheck{cmfRC(0), cmfOut("<out>/n1.png diff=n/a")}},
	{"case 24: the composite is two native-size panels, uncropped", "case-24", []cmfCheck{cmfSize("n1.png", 1316, 1164)}},
	{"case 25: malformed geometry `scale=2 status=26` exits 2 with usage on stderr", "case-25a", []cmfCheck{cmfRC(2), cmfErr(`usage:`)}},
	{"case 25: malformed geometry `scale=0 status=26 border=1` exits 2 with usage on stderr", "case-25b", []cmfCheck{cmfRC(2), cmfErr(`usage:`)}},
	{"case 25: malformed geometry `scale=2 status=-1 border=1` exits 2 with usage on stderr", "case-25c", []cmfCheck{cmfRC(2), cmfErr(`usage:`)}},
	{"case 25: malformed geometry `scale=two status=26 border=1` exits 2 with usage on stderr", "case-25d", []cmfCheck{cmfRC(2), cmfErr(`usage:`)}},
	{"case 25: malformed geometry `scale=2 status=26 border=1 extra=9` exits 2 with usage on stderr", "case-25e", []cmfCheck{cmfRC(2), cmfErr(`usage:`)}},
	{"case 25: malformed geometry `scale=2 scale=2 border=1` exits 2 with usage on stderr", "case-25f", []cmfCheck{cmfRC(2), cmfErr(`usage:`)}},
	{"case 26: a caption band directly below the border crops at the frame's own bottom border", "case-26", []cmfCheck{cmfRC(0), cmfOut("<out>/c1.png diff=0.0000")}},
	{"case 26: the cropped frame is the content area's size", "case-26", []cmfCheck{cmfSize("c1.frame.png", 96, 60)}},
	{"case 27: a border-coloured content row far below a broken side border is not taken for the bottom border", "case-27", []cmfCheck{cmfRC(0)}},
	{"case 27: the cropped frame runs to the last row above the bottom border", "case-27", []cmfCheck{cmfSize("c1.frame.png", 96, 120)}},
	{"case 28: a frame with no border, its content background at (0, 0), exits 1 rather than cropping", "case-28", []cmfCheck{cmfRC(1), cmfOut(""), cmfErr(`map\.mockups:1: frame has no border: root/B1\.png\n`)}},
	{"case 28: the borderless frame writes no composite and no cropped frame", "case-28", []cmfCheck{cmfExists("b1.png", false), cmfExists("b1.frame.png", false)}},

	{"one-level red and blue differences each count toward diff=", "any-channel", []cmfCheck{cmfRC(0), cmfOut("<out>/d1.png diff=0.0010")}},
	{"a one-level red difference is white in the difference panel", "any-channel", []cmfCheck{cmfPixel("d1.png", 96+16+96+16+10, 5, 255, 255, 255), cmfPixel("d1.png", 96+16+96+16+11, 5, 0, 0, 0)}},
	{"a one-level blue difference is white in the difference panel", "any-channel", []cmfCheck{cmfPixel("d1.png", 96+16+96+16+12, 5, 255, 255, 255)}},
	{"a NUL byte in a stdin path is Python's embedded null byte", "stdin-nul", []cmfCheck{cmfRC(2), cmfErr(`: embedded null byte\n\z`)}},
	{"a map line split by \\x1c composes", "fs-separator", []cmfCheck{cmfRC(0), cmfOut("<out>/k2-menu.png diff=n/a")}},
	{"a CRLF stdin line reads as its path", "stdin-crlf", []cmfCheck{cmfRC(0), cmfOut("<out>/k2-menu.png diff=n/a")}},
	{"a lone \\r on stdin does not end a line, as Python's sys.stdin", "stdin-lone-cr", []cmfCheck{cmfRC(2), cmfErr(`'caps/p1\.png\\rcaps/p2\.png'`)}},
	{"a map file that is not UTF-8 exits 2 with one compose-mockup-frames: line", "map-not-utf8", []cmfCheck{cmfRC(2), cmfOut(""), cmfErr(`\Acompose-mockup-frames: [^\n]*\n\z`)}},
	{"a relative out dir from a cwd reached through a symlink prints the physical absolute path", "symlink-cwd", []cmfCheck{cmfRC(0), cmfOut("<out>/k2-menu.png diff=n/a")}},
}

func TestComposeMockupFrames(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	runs := map[string]func() cmfResult{}
	for name, inv := range cmfInvs {
		runs[name] = sync.OnceValue(func() cmfResult { return cmfRun(base, name, inv) })
	}
	for _, c := range cmfCases {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			r := runs[c.inv]()
			if r.golden != "" {
				t.Errorf("departs from the Python's golden output: %s", r.golden)
			}
			for _, check := range c.checks {
				if msg := check(r); msg != "" {
					t.Errorf("%s\nrc=%d\nstdout=%q\nstderr=%q", msg, r.rc, r.stdout, r.stderr)
				}
			}
		})
	}
	t.Run("stdin captures are validated, not retained", func(t *testing.T) {
		t.Parallel()
		// The Python loaded each stdin capture only to validate it and
		// re-opened the matched one per map line, so peak memory followed one
		// pair, not the stdin list. Measured as the guard's peak RSS in a
		// child process: 20 captures on stdin against 1, the growth held under
		// half of what retaining the 19 extra 1000×1000 RGB decodes costs.
		// Random pixels, since macOS compresses a uniform page out of RSS.
		const w, h, n = 1000, 1000, 20
		dir := t.TempDir()
		im := &rgbImage{W: w, H: h, Pix: make([]byte, 3*w*h)}
		_, _ = rand.NewChaCha8([32]byte{}).Read(im.Pix)
		if err := encodeRGB(dir+"/c0.png", im); err != nil {
			t.Fatal(err)
		}
		paths := []string{dir + "/c0.png"}
		for i := 1; i < n; i++ {
			p := fmt.Sprintf("%s/c%d.png", dir, i)
			if err := os.Link(dir+"/c0.png", p); err != nil {
				t.Fatal(err)
			}
			paths = append(paths, p)
		}
		mkdir(t, dir+"/frames")
		writeFile(t, dir+"/map", "c0.png f\n")
		peak := func(stdin []string) int64 {
			cmd := exec.Command(guardBinary(t), "compose-mockup-frames", dir+"/map", dir+"/frames", dir+"/out")
			cmd.Stdin = strings.NewReader(strings.Join(stdin, "\n") + "\n")
			if out, err := cmd.CombinedOutput(); !strings.Contains(string(out), "no frame `f.png`") {
				t.Fatalf("compose did not reach the map line: %v\n%s", err, out)
			}
			rss := cmd.ProcessState.SysUsage().(*syscall.Rusage).Maxrss
			if runtime.GOOS != "darwin" {
				rss *= 1024 // kilobytes on Linux, bytes on macOS
			}
			return rss
		}
		one, many := peak(paths[:1]), peak(paths)
		if limit := int64((n - 1) * 3 * w * h / 2); many-one > limit {
			t.Errorf("peak RSS grew %d B from 1 to %d stdin captures, over %d B: the captures are retained", many-one, n, limit)
		}
	})
}

// cmfRun runs one invocation into base/<name> and compares it with the
// Python's golden output where the fixture carries one.
func cmfRun(base, name string, inv cmfInv) cmfResult {
	fail := func(err error) cmfResult { return cmfResult{rc: -1, golden: "setup: " + err.Error()} }
	if inv.dir == "" {
		inv.dir = name
	}
	fix, err := filepath.Abs("testdata/compose-mockup-frames/" + inv.dir)
	if err != nil {
		return fail(err)
	}
	var r cmfResult
	if inv.shim {
		// The harness's case 12: a lone copy of the guard, its sibling absent.
		dir := base + "/" + name
		body, err := os.ReadFile("../../../scripts/compose-mockup-frames.sh")
		if err != nil {
			return fail(err)
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fail(err)
		}
		if err := os.WriteFile(dir+"/compose-mockup-frames.sh", body, 0o755); err != nil {
			return fail(err)
		}
		var stdout, stderr bytes.Buffer
		cmd := exec.Command("bash", dir+"/compose-mockup-frames.sh", "a", "b", "c")
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		_ = cmd.Run()
		return cmfResult{rc: cmd.ProcessState.ExitCode(), stdout: stdout.String(), stderr: stderr.String()}
	}
	dir, out, outArg := fix, base+"/"+name, base+"/"+name
	if inv.viaSymlink {
		real := base + "/" + name + "/real"
		if err := os.CopyFS(real, os.DirFS(fix)); err != nil {
			return fail(err)
		}
		if err := os.RemoveAll(real + "/golden"); err != nil {
			return fail(err)
		}
		if err := os.Symlink(real, base+"/"+name+"/link"); err != nil {
			return fail(err)
		}
		phys, err := filepath.EvalSymlinks(real)
		if err != nil {
			return fail(err)
		}
		dir, out, outArg = base+"/"+name+"/link", phys+"/out", "out"
	}
	args := slices.Clone(inv.args)
	if args == nil {
		args = []string{"map.mockups", "root", outArg}
		if inv.geom != "" {
			args = append(args, inv.geom)
		}
	}
	for i, a := range args {
		args[i] = strings.ReplaceAll(a, "<out>", outArg)
	}
	env := Env{Getenv: os.Getenv, Dir: dir}
	if inv.stdin != "" {
		env.Stdin = strings.NewReader(inv.stdin)
	}
	var stdout, stderr bytes.Buffer
	r.rc = -1
	if fn := Registry["compose-mockup-frames"]; fn != nil {
		r.rc = fn(args, env, &stdout, &stderr)
	} else {
		stderr.WriteString("compose-mockup-frames is not registered\n")
	}
	r.stdout, r.stderr, r.out = stdout.String(), stderr.String(), out
	if _, err := os.Stat(fix + "/golden"); err == nil {
		r.golden = cmfGolden(fix+"/golden", r, inv)
	}
	return r
}

// cmfGolden lists every way r departs from the Python's golden output.
func cmfGolden(golden string, r cmfResult, inv cmfInv) string {
	var diffs []string
	read := func(name string) string {
		b, err := os.ReadFile(golden + "/" + name)
		if err != nil {
			diffs = append(diffs, err.Error())
		}
		return string(b)
	}
	if want := strings.TrimSpace(read("rc")); want != strconv.Itoa(r.rc) {
		diffs = append(diffs, fmt.Sprintf("rc %d, golden %s", r.rc, want))
	}
	wantOut := read("stdout")
	if inv.stdout != "" {
		wantOut = inv.stdout
	}
	if got := strings.ReplaceAll(r.stdout, r.out, "<out>"); got != wantOut {
		diffs = append(diffs, fmt.Sprintf("stdout %q, golden %q", got, wantOut))
	}
	if want := read("stderr"); r.stderr != want {
		diffs = append(diffs, fmt.Sprintf("stderr %q, golden %q", r.stderr, want))
	}
	wantPNGs, _ := filepath.Glob(golden + "/*.png")
	gotPNGs, _ := filepath.Glob(r.out + "/*.png")
	if len(gotPNGs) != len(wantPNGs) {
		diffs = append(diffs, fmt.Sprintf("wrote %d PNGs, golden %d", len(gotPNGs), len(wantPNGs)))
	}
	for _, w := range wantPNGs {
		name := filepath.Base(w)
		if d := cmfComparePNG(r.out+"/"+name, w, inv.geom != "" && !strings.HasSuffix(name, ".frame.png")); d != "" {
			diffs = append(diffs, name+": "+d)
		}
	}
	return strings.Join(diffs, "; ")
}

// cmfComparePNG compares got with want pixel for pixel; a three-panel
// composite's difference panel is compared with an any-channel mask of
// want's own capture and frame panels instead.
func cmfComparePNG(got, want string, threePanel bool) string {
	g, err := decodeRGB(got)
	if err != nil {
		return err.Error()
	}
	w, err := decodeRGB(want)
	if err != nil {
		return err.Error()
	}
	if g.W != w.W || g.H != w.H {
		return fmt.Sprintf("%dx%d, golden %dx%d", g.W, g.H, w.W, w.H)
	}
	panel := (w.W - 2*cmfGutter) / 3
	diffX := 2 * (panel + cmfGutter)
	for y := 0; y < w.H; y++ {
		for x := 0; x < w.W; x++ {
			want := w.at(x, y)
			if threePanel && x >= diffX {
				want = [3]byte{}
				if w.at(x-diffX, y) != w.at(x-diffX+panel+cmfGutter, y) {
					want = [3]byte{255, 255, 255}
				}
			}
			if got := g.at(x, y); got != want {
				return fmt.Sprintf("pixel (%d, %d) is %v, want %v", x, y, got, want)
			}
		}
	}
	return ""
}

func cmfRC(want int) cmfCheck {
	return func(r cmfResult) string {
		if r.rc != want {
			return fmt.Sprintf("exit %d, want %d", r.rc, want)
		}
		return ""
	}
}

// cmfOut is the harness's `$(…)` compare: stdout, trailing newlines
// dropped, against want with <out> spelled as the run's output directory.
func cmfOut(want string) cmfCheck {
	return func(r cmfResult) string {
		if got, w := strings.TrimRight(r.stdout, "\n"), strings.ReplaceAll(want, "<out>", r.out); got != w {
			return fmt.Sprintf("stdout %q, want %q", got, w)
		}
		return ""
	}
}

// cmfErr is the harness's `case "$ERR" in *…*` match, as a regexp.
func cmfErr(pattern string) cmfCheck {
	re := regexp.MustCompile(pattern)
	return func(r cmfResult) string {
		if !re.MatchString(r.stderr) {
			return fmt.Sprintf("stderr does not match %q", pattern)
		}
		return ""
	}
}

func cmfSize(name string, w, h int) cmfCheck {
	return func(r cmfResult) string {
		im, err := decodeRGB(r.out + "/" + name)
		if err != nil {
			return err.Error()
		}
		if im.W != w || im.H != h {
			return fmt.Sprintf("%s is %dx%d, want %dx%d", name, im.W, im.H, w, h)
		}
		return ""
	}
}

func cmfPixel(name string, x, y int, rgb ...byte) cmfCheck {
	return func(r cmfResult) string {
		im, err := decodeRGB(r.out + "/" + name)
		if err != nil {
			return err.Error()
		}
		if got := im.at(x, y); !bytes.Equal(got[:], rgb) {
			return fmt.Sprintf("%s (%d, %d) is %v, want %v", name, x, y, got, rgb)
		}
		return ""
	}
}

func cmfExists(name string, want bool) cmfCheck {
	return func(r cmfResult) string {
		if _, err := os.Stat(r.out + "/" + name); (err == nil) != want {
			return fmt.Sprintf("%s exists=%v, want %v", name, err == nil, want)
		}
		return ""
	}
}
