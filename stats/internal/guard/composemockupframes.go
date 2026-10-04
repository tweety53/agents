package guard

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// composeMockupFrames is scripts/compose-mockup-frames.sh: that script's
// header comment is the contract -- one composite PNG per map line pairing a
// captured screenshot with the mockup frame it was drawn from, exit 0/1/2.
// It ports compose-mockup-frames.py (3915fbc0), whose algorithm prose and
// comments sit here beside the code they explain. Its departures from the
// Python: the difference mask counts any differing channel
// (any-channel-diff-mask), a 16-bit PNG converts by high byte
// (png-rgb-decode), and an input the Python died on with a traceback -- a map
// or stdin that is not UTF-8, an output PNG it could not write -- exits 2 with
// one line (python-io-semantics).
func init() {
	Registry["compose-mockup-frames"] = composeMockupFrames
}

const (
	cmfPrefix = "compose-mockup-frames: "
	cmfUsage  = "usage: compose-mockup-frames.sh <map file> <mockups root> <out dir>" +
		" [scale=<n> status=<px> border=<px>]" +
		"   (resolved PNG paths on stdin, one per line)\n"
	cmfGutter = 16
	// The side border's run stops at the frame's rounded bottom corner, this
	// many logical rows at most above the bottom border's first row (1.5 on
	// every real frame; content rows mis-taken for the border sat 56+ logical
	// rows away).
	cmfBottomCornerRows = 4
)

var cmfBackground = [3]byte{40, 40, 40}

// Node's process.platform values (https://nodejs.org/api/process.html#processplatform) --
// the exhaustive set Playwright's own snapshot suffix is drawn from. Matching against this
// closed vocabulary, rather than "any continuation", is what keeps one screenshot name from
// being mistaken for another's platform-suffixed capture merely because it is a string prefix
// of it (`add-participant-finished-excluded.png` vs.
// `add-participant-finished-excluded-added-darwin.png` -- a real collision, not hypothetical).
var cmfPlatformSuffixes = []string{"aix", "android", "darwin", "freebsd", "linux", "openbsd", "sunos", "win32"}

type cmfGeometry struct{ scale, status, border int }

type cmfOutput struct {
	path  string
	ratio string
}

func composeMockupFrames(args []string, env Env, stdout, stderr io.Writer) int {
	if len(args) != 3 && len(args) != 4 {
		fmt.Fprint(stderr, cmfUsage)
		return 2
	}
	mapPath, mockupsRoot, outDir := args[0], args[1], args[2]
	var geometry *cmfGeometry
	if len(args) == 4 {
		g, ok := cmfParseGeometry(args[3])
		if !ok {
			fmt.Fprint(stderr, cmfUsage)
			return 2
		}
		geometry = &g
	}
	at := func(p string) string {
		if p == "" || filepath.IsAbs(p) {
			return p
		}
		return env.Dir + "/" + p
	}

	data, err := os.ReadFile(at(mapPath))
	if err != nil {
		fmt.Fprintf(stderr, cmfPrefix+"cannot read map file %s: %s\n", mapPath, ppOSError(err, mapPath))
		return 2
	}
	if !utf8.Valid(data) {
		fmt.Fprintf(stderr, cmfPrefix+"cannot read map file %s: not valid UTF-8\n", mapPath)
		return 2
	}
	// open(…, encoding="utf-8").readlines(): universal newlines, so a lone
	// \r ends a line as \n and \r\n do.
	mapLines := strings.Split(strings.ReplaceAll(strings.ReplaceAll(string(data), "\r\n", "\n"), "\r", "\n"), "\n")

	if fi, err := os.Stat(at(mockupsRoot)); err != nil || !fi.IsDir() {
		fmt.Fprintf(stderr, cmfPrefix+"mockups root is not a directory: %s\n", mockupsRoot)
		return 2
	}

	if name, err := cmfMakedirs(at, outDir); err != nil {
		fmt.Fprintf(stderr, cmfPrefix+"cannot create output directory %s: %s\n", outDir, ppOSError(err, name))
		return 2
	}

	// sys.stdin splits on \n alone on POSIX -- CPython opens it with
	// newline="\n", not universal newlines -- so a lone \r stays inside the
	// path it sits in, and strip() drops a \r\n's \r.
	var in []byte
	if env.Stdin != nil {
		if in, err = io.ReadAll(env.Stdin); err != nil {
			fmt.Fprintf(stderr, cmfPrefix+"cannot read stdin: %v\n", err)
			return 2
		}
	}
	if !utf8.Valid(in) {
		fmt.Fprint(stderr, cmfPrefix+"cannot read stdin: not valid UTF-8\n")
		return 2
	}
	var stdinPaths []string
	for _, line := range strings.Split(string(in), "\n") {
		if p := pyStrip(line); p != "" {
			stdinPaths = append(stdinPaths, p)
		}
	}
	// Each capture is decoded to validate it and let go, as the Python's
	// load() did; the matched one is decoded again per map line, so peak
	// memory follows one pair rather than the stdin list.
	decodeCapture := func(p string) (*rgbImage, int) {
		im, err := decodeRGB(at(p))
		if err != nil {
			fmt.Fprintf(stderr, cmfPrefix+"unreadable PNG on stdin: %s: %s\n", p, pilImageError(err, at(p), p))
			return nil, 2
		}
		return im, 0
	}
	for _, p := range stdinPaths {
		if _, code := decodeCapture(p); code != 0 {
			return code
		}
	}

	var findings []string
	var outputs []cmfOutput
	seenNames := map[string]bool{}
	finding := func(lineno int, format string, a ...any) {
		findings = append(findings, fmt.Sprintf("%s:%d: ", mapPath, lineno)+fmt.Sprintf(format, a...))
	}

	for i, raw := range mapLines {
		lineno := i + 1
		line := pyStrip(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := pySplit(line)
		if len(fields) != 2 {
			finding(lineno, "expected `<screenshot name> <frame id>`, got %d fields", len(fields))
			continue
		}
		name, frameID := fields[0], fields[1]

		if seenNames[name] {
			finding(lineno, "duplicate screenshot name `%s`", name)
			continue
		}
		seenNames[name] = true

		if !strings.HasSuffix(name, ".png") {
			finding(lineno, "expected `<screenshot name> <frame id>`, `%s` does not end in .png", name)
			continue
		}

		matches := cmfMatchCapture(name, stdinPaths)
		if len(matches) == 0 {
			finding(lineno, "no captured PNG matches `%s`", name)
			continue
		}
		if len(matches) > 1 {
			finding(lineno, "multiple captured PNGs match `%s`: %s", name, strings.Join(matches, ", "))
			continue
		}

		framePath := cmfJoin(mockupsRoot, frameID+".png")
		if fi, err := os.Stat(at(framePath)); err != nil || !fi.Mode().IsRegular() {
			finding(lineno, "no frame `%s.png` under %s", frameID, mockupsRoot)
			continue
		}

		stem := strings.TrimSuffix(name, ".png")
		outPath := cmfJoin(outDir, stem+".png")

		mock, err := decodeRGB(at(framePath))
		if err != nil {
			fmt.Fprintf(stderr, cmfPrefix+"unreadable PNG: %s: %s\n", framePath, pilImageError(err, at(framePath), framePath))
			return 2
		}
		capture, code := decodeCapture(matches[0])
		if code != 0 {
			return code
		}

		var canvas *rgbImage
		ratio := "n/a"
		if geometry == nil {
			canvas = cmfCanvas(capture.W+cmfGutter+mock.W, max(capture.H, mock.H))
			cmfPaste(canvas, capture, 0, 0)
			cmfPaste(canvas, mock, capture.W+cmfGutter, 0)
		} else {
			box, err := cmfCropBox(mock, *geometry)
			if errors.Is(err, errCmfNoBorder) {
				finding(lineno, "frame has no border: %s", framePath)
				continue
			}
			if err != nil {
				finding(lineno, "frame %s: declared geometry leaves no content area", frameID)
				continue
			}
			mock = cmfCrop(mock, box)
			// The cropped frame is the calibrated ruler every later
			// measurement is taken against (scale 1 to the capture), so it is
			// written beside the composite rather than re-derived by hand.
			if code := cmfSave(stderr, at, cmfJoin(outDir, stem+".frame.png"), mock); code != 0 {
				return code
			}

			if capture.W != mock.W || capture.H != mock.H {
				finding(lineno, "capture %d×%d vs frame %d×%d for %s", capture.W, capture.H, mock.W, mock.H, frameID)
				continue
			}

			diff, differing := cmfDifferencePanel(capture, mock)
			ratio = fmt.Sprintf("%.4f", float64(differing)/float64(capture.W*capture.H))
			canvas = cmfCanvas(capture.W+cmfGutter+mock.W+cmfGutter+diff.W, capture.H)
			cmfPaste(canvas, capture, 0, 0)
			cmfPaste(canvas, mock, capture.W+cmfGutter, 0)
			cmfPaste(canvas, diff, capture.W+cmfGutter+mock.W+cmfGutter, 0)
		}
		if code := cmfSave(stderr, at, outPath, canvas); code != 0 {
			return code
		}
		outputs = append(outputs, cmfOutput{cmfAbspath(env, outPath), ratio})
	}

	for _, o := range outputs {
		fmt.Fprintf(stdout, "%s diff=%s\n", o.path, o.ratio)
	}
	for _, f := range findings {
		fmt.Fprintln(stderr, f)
	}
	if len(findings) > 0 {
		return 1
	}
	return 0
}

func cmfSave(stderr io.Writer, at func(string) string, path string, im *rgbImage) int {
	if err := encodeRGB(at(path), im); err != nil {
		fmt.Fprintf(stderr, cmfPrefix+"cannot write %s: %s\n", path, ppOSError(err, path))
		return 2
	}
	return 0
}

// cmfParseGeometry reads `scale=<int> status=<int> border=<int>`.
//
// Every key required exactly once, in any order; every value a non-negative
// integer and `scale` at least 1. Anything else is a usage error rather than
// a silently-ignored argument: a geometry the caller meant to declare and
// mistyped must not compose uncropped as though none had been given.
//
// A value saturates at 1<<30, past any dimension a decodable PNG has, so
// every comparison cmfCropBox makes answers as Python's unbounded int did.
func cmfParseGeometry(arg string) (cmfGeometry, bool) {
	values := map[string]int{}
	for _, field := range pySplit(arg) {
		key, raw, sep := strings.Cut(field, "=")
		if _, dup := values[key]; !sep || dup || (key != "scale" && key != "status" && key != "border") {
			return cmfGeometry{}, false
		}
		if raw == "" {
			return cmfGeometry{}, false
		}
		v := 0
		for _, c := range []byte(raw) {
			if c < '0' || c > '9' {
				return cmfGeometry{}, false
			}
			v = min(v*10+int(c-'0'), 1<<30)
		}
		values[key] = v
	}
	if len(values) != 3 || values["scale"] < 1 {
		return cmfGeometry{}, false
	}
	return cmfGeometry{values["scale"], values["status"], values["border"]}, true
}

// cmfMatchCapture returns the stdin path(s) whose basename is name or
// `<stem>-<platform>.png` where <stem> is name without `.png` and <platform> is
// exactly one Node `process.platform` value (Playwright's own snapshot suffix).
// Deliberately not `<stem>-<anything>.png`: that loosely matches a longer,
// unrelated screenshot name's own capture whenever it happens to start with
// this name's stem followed by a dash.
func cmfMatchCapture(name string, stdinPaths []string) []string {
	stem := strings.TrimSuffix(name, ".png")
	var matches []string
	for _, p := range stdinPaths {
		base := p[strings.LastIndex(p, "/")+1:] // os.path.basename
		if base == name {
			matches = append(matches, p)
			continue
		}
		if b, ok := strings.CutSuffix(base, ".png"); ok {
			for _, s := range cmfPlatformSuffixes {
				if b == stem+"-"+s {
					matches = append(matches, p)
					break
				}
			}
		}
	}
	return matches
}

var (
	errCmfNoContent = errors.New("declared geometry leaves no content area")
	errCmfNoBorder  = errors.New("frame has no border")
)

// cmfCropBox is the frame's content area in PNG pixels -- left, top, right,
// bottom -- or errCmfNoContent when the declared geometry leaves none, or
// errCmfNoBorder when a border is declared and the frame's side border pixel
// (rule 2's, just left of the crop at the probe row) is the colour at (0, 0):
// the image has no border telling page from frame -- gymie's X4 was drawn
// without its top and left border, its content background running to (0, 0)
// -- and rule 1 would take a uniform content row for the margin and crop short
// in silence.
//
// The sides and the top come from the declared values; the status line and
// the border are declared rather than detected. The bottom is found in the
// image, because content-hugging frames differ in height from one another,
// by two rules in turn:
//
//  1. The page background is the pixel at (0, 0) -- outside the frame's
//     rounded corner, so always the page and never the border -- and the
//     first row below the top crop whose every pixel between the side
//     borders equals it is the margin above the caption band; the bottom
//     border sits directly above it.
//  2. No such row -- gymie's J/K floor frames draw their caption band on the
//     row directly below the bottom border, with no page-coloured margin --
//     and a border declared: the side border column (just left of the crop)
//     holds one colour from the probe row (below the rounded top corner, or the top
//     crop when that is lower) down to the bottom corner, and the
//     first row within the corner's height below that run (a few rows: 3 on
//     every real frame at 2x) whose middle pixel is that colour is the
//     bottom border's first row. A matching row further down is content in
//     the border colour under a side border that breaks mid-height (the
//     copy-session C1/C4/C7 frames), never the bottom border.
//
// Neither rule finding a row means the frame runs to the last row.
func cmfCropBox(frame *rgbImage, g cmfGeometry) ([4]int, error) {
	s := g.scale
	left := g.border * s
	top := (g.border + g.status) * s
	right := frame.W - g.border*s
	if right <= left || top >= frame.H {
		return [4]int{}, errCmfNoContent
	}

	page := frame.at(0, 0)
	// The side border is read at the probe row, below the rounded top corner:
	// with status=0 the top crop row still sits inside the corner, where the
	// pixel is the page or the corner's anti-aliased edge, never the border.
	probe := min(max(top, (g.border+cmfBottomCornerRows)*s), frame.H-1)
	if g.border > 0 && frame.at(left-1, probe) == page {
		return [4]int{}, errCmfNoBorder
	}
	bottom := frame.H - g.border*s
	found := false
	for y := top; y < frame.H && !found; y++ {
		found = true
		for x := left; x < right; x++ {
			if frame.at(x, y) != page {
				found = false
				break
			}
		}
		if found {
			bottom = y - g.border*s
		}
	}
	if !found && g.border > 0 {
		edge := frame.at(left-1, probe)
		y := probe
		for y+1 < frame.H && frame.at(left-1, y+1) == edge {
			y++
		}
		mid := frame.W / 2
		for yy := y + 1; yy < min(y+1+cmfBottomCornerRows*s, frame.H); yy++ {
			if frame.at(mid, yy) == edge {
				bottom = yy
				break
			}
		}
	}

	if bottom <= top {
		return [4]int{}, errCmfNoContent
	}
	return [4]int{left, top, right, bottom}, nil
}

// cmfDifferencePanel is white wherever any channel differs and black elsewhere,
// so a departure reads as shape rather than as a colour blend, plus the count
// of differing pixels. Pillow's difference().convert("L"), the Python's mask,
// rounded a one-level red or blue difference to luma 0 and missed it
// (any-channel-diff-mask).
func cmfDifferencePanel(capture, frame *rgbImage) (*rgbImage, int) {
	mask := &rgbImage{W: capture.W, H: capture.H, Pix: make([]byte, len(capture.Pix))}
	differing := 0
	for i := 0; i < len(capture.Pix); i += 3 {
		if !bytes.Equal(capture.Pix[i:i+3], frame.Pix[i:i+3]) {
			mask.Pix[i], mask.Pix[i+1], mask.Pix[i+2] = 255, 255, 255
			differing++
		}
	}
	return mask, differing
}

func cmfCanvas(w, h int) *rgbImage {
	c := &rgbImage{W: w, H: h, Pix: make([]byte, 3*w*h)}
	for i := 0; i < len(c.Pix); i += 3 {
		copy(c.Pix[i:], cmfBackground[:])
	}
	return c
}

// cmfPaste copies src onto im with its top-left corner at (x0, y0); src fits.
func cmfPaste(im, src *rgbImage, x0, y0 int) {
	for y := 0; y < src.H; y++ {
		copy(im.Pix[3*((y0+y)*im.W+x0):], src.Pix[3*y*src.W:3*(y+1)*src.W])
	}
}

// cmfCrop is im's box -- left, top, right, bottom -- as a new image.
func cmfCrop(im *rgbImage, box [4]int) *rgbImage {
	out := &rgbImage{W: box[2] - box[0], H: box[3] - box[1]}
	for y := box[1]; y < box[3]; y++ {
		out.Pix = append(out.Pix, im.Pix[3*(y*im.W+box[0]):3*(y*im.W+box[2])]...)
	}
	return out
}

// cmfJoin is os.path.join(a, b): b wins when absolute, and nothing is cleaned.
func cmfJoin(a, b string) string {
	if strings.HasPrefix(b, "/") {
		return b
	}
	if a == "" || strings.HasSuffix(a, "/") {
		return a + b
	}
	return a + "/" + b
}

// cmfAbspath is os.path.abspath: joined onto the physical working directory
// (os.getcwd(); env.Dir is the logical $PWD) and normalised lexically, a
// leading `//` -- exactly two slashes -- kept as posixpath.normpath keeps it.
func cmfAbspath(env Env, p string) string {
	if !strings.HasPrefix(p, "/") {
		cwd, err := filepath.EvalSymlinks(env.Dir)
		if err != nil {
			cwd = env.Dir
		}
		p = cwd + "/" + p
	}
	c := filepath.Clean(p)
	if strings.HasPrefix(p, "//") && !strings.HasPrefix(p, "///") {
		c = "/" + c
	}
	return c
}

// cmfMakedirs is os.makedirs(name, exist_ok=True): it creates each missing
// parent first, and a failure names the path whose mkdir failed, as the
// Python's OSError did.
func cmfMakedirs(at func(string) string, name string) (string, error) {
	head, tail := cmfSplitPath(name)
	if tail == "" {
		head, tail = cmfSplitPath(head)
	}
	if _, err := os.Stat(at(head)); head != "" && tail != "" && err != nil {
		if n, err := cmfMakedirs(at, head); err != nil && !errors.Is(err, os.ErrExist) {
			return n, err
		}
		if tail == "." {
			return "", nil
		}
	}
	if err := os.Mkdir(at(name), 0o777); err != nil {
		if fi, serr := os.Stat(at(name)); serr != nil || !fi.IsDir() {
			return name, err
		}
	}
	return "", nil
}

// cmfSplitPath is os.path.split: the head keeps no trailing slash unless it is
// nothing but slashes.
func cmfSplitPath(p string) (string, string) {
	i := strings.LastIndex(p, "/") + 1
	head, tail := p[:i], p[i:]
	if strings.Trim(head, "/") != "" {
		head = strings.TrimRight(head, "/")
	}
	return head, tail
}
