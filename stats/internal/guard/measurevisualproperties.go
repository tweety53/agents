package guard

import (
	"errors"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// measure-visual-properties: one UI control's visual properties measured from
// pixels, in one image or two, printed as the Python script's JSON byte for
// byte (design.md byte-identical-json). The interface, the region layout,
// every property and the exit codes are scripts/measure-visual-properties.sh's
// header; the comments here are the algorithm's.

func init() {
	Registry["measure-visual-properties"] = measureVisualProperties
}

const mvpPrefix = "measure-visual-properties: "

var mvpAllProps = []string{"box", "radius", "border", "fill", "shadow", "content", "gap", "ink", "runs", "bands", "seams"}

// Properties that read the region as a whole page: background is the modal
// colour and the corners need not be background.
var mvpPageProps = []string{"bands", "seams"}

// mvpFull is the share of a band's height a column must ink to be a seam, and
// of a cell's width a row must ink to be a rule rather than a text line.
const mvpFull = 0.9

// Numeric leaves that are not lengths and therefore are not scaled.
var mvpUnscaled = []string{"ratio", "approx_opacity", "peak_delta", "share"}

// mvpUnresolved is a region that cannot be resolved to a clean bounding box:
// exit 1, the reason on stderr, nothing guessed.
type mvpUnresolved string

func (e mvpUnresolved) Error() string { return string(e) }

// mvpHelp is what argparse printed for -h with stdout not a terminal (80
// columns): the port prints the same text wherever it runs.
const mvpHelp = `usage: measure-visual-properties.sh [-h] [--region-a REGION_A]
                                    [--region-b REGION_B] [--scale SCALE]
                                    [--ref-a REF_A] [--ref-b REF_B]
                                    [--props PROPS] [--edge EDGE]
                                    [--noise NOISE]
                                    image_a [image_b]

positional arguments:
  image_a
  image_b

options:
  -h, --help           show this help message and exit
  --region-a REGION_A
  --region-b REGION_B
  --scale SCALE
  --ref-a REF_A
  --ref-b REF_B
  --props PROPS
  --edge EDGE
  --noise NOISE
`

var mvpOptions = []string{"-h", "--help", "--region-a", "--region-b", "--scale", "--ref-a", "--ref-b", "--props", "--edge", "--noise"}

type mvpArgs struct {
	images                       []string
	regionA, regionB, refA, refB *[4]int
	scale                        *float64
	props                        string
	edge, noise                  float64
	edgeText                     string // str(edge): argparse left the default an int
}

func measureVisualProperties(args []string, env Env, stdout, stderr io.Writer) int {
	a, code := mvpParse(args, stdout, stderr)
	if code >= 0 {
		return code
	}
	refuse := func(msg string) int {
		fmt.Fprintln(stderr, mvpPrefix+msg)
		return 2
	}

	var props, unknown []string
	for _, s := range strings.Split(a.props, ",") {
		if s = pyStrip(s); s != "" {
			props = append(props, s)
			if !slices.Contains(mvpAllProps, s) {
				unknown = append(unknown, s)
			}
		}
	}
	if len(unknown) > 0 || len(props) == 0 {
		quoted := make([]string, len(unknown))
		for i, s := range unknown {
			quoted[i] = ppRepr(s)
		}
		return refuse(fmt.Sprintf("unknown --props [%s]; choose from %s", strings.Join(quoted, ", "), strings.Join(mvpAllProps, ",")))
	}

	// An empty second image is argparse's falsy image_b: one image.
	imageB := ""
	if len(a.images) > 1 {
		imageB = a.images[1]
	}
	var scale float64
	if imageB != "" {
		switch {
		case a.scale != nil && (a.refA != nil || a.refB != nil):
			return refuse("give --scale or the --ref-a/--ref-b pair, not both")
		case a.scale != nil:
			if *a.scale <= 0 {
				return refuse("--scale must be positive")
			}
			scale = *a.scale
		case a.refA != nil && a.refB != nil:
			scale = float64(a.refB[2]) / float64(a.refA[2])
			byHeight := float64(a.refB[3]) / float64(a.refA[3])
			if math.Abs(byHeight-scale)/scale > 0.05 {
				return refuse(fmt.Sprintf("--ref boxes disagree: width factor %.4f, height factor %.4f", scale, byHeight))
			}
		default:
			return refuse("two images need a calibration — --scale <B px per A px>, or --ref-a and --ref-b boxes already known correct in both images")
		}
	} else if a.scale != nil || a.refA != nil || a.refB != nil || a.regionB != nil {
		return refuse("--scale/--ref-*/--region-b need a second image")
	}

	page := true
	for _, p := range props {
		page = page && slices.Contains(mvpPageProps, p)
	}
	result := pyObj{}
	which := "a"
	measure := func(path string, box *[4]int) (mvpMeasured, error) {
		r, err := mvpOpen(env, path, box, a, page)
		if err != nil {
			return mvpMeasured{}, err
		}
		return r.measure(props)
	}
	ma, err := measure(a.images[0], a.regionA)
	if err == nil {
		result = append(result, pyKV{"a", ma.out})
		if imageB != "" {
			which = "b"
			var mb mvpMeasured
			if mb, err = measure(imageB, a.regionB); err == nil {
				result = append(result, pyKV{"b", mb.out}, pyKV{"scale", scale}, pyKV{"delta", mvpDelta(ma, mb, props, scale)})
			}
		}
	}
	var unresolved mvpUnresolved
	if errors.As(err, &unresolved) {
		fmt.Fprintf(stderr, "%simage %s: %s\n", mvpPrefix, which, unresolved)
		return 1
	}
	if err != nil {
		return refuse(err.Error())
	}
	fmt.Fprintln(stdout, pyDump(result))
	return 0
}

// mvpParse is the script's argparse surface under exact option names
// (design.md exact-option-names): `--opt value` and `--opt=value`, the last
// of a repeated option winning, positionals anywhere and everything after
// the first `--` positional, -h/--help printing the usage text. A usage error
// is one line carrying argparse's own message. Errors surface in argparse's
// order: an option's missing or bad value and -h act where they stand; a
// missing image and an unrecognized argument are reported after the scan.
// code is -1 when the arguments parsed.
func mvpParse(args []string, stdout, stderr io.Writer) (a mvpArgs, code int) {
	a = mvpArgs{props: strings.Join(mvpAllProps, ","), edge: 24, noise: 8, edgeText: "24"}
	usage := func(msg string) (mvpArgs, int) {
		fmt.Fprintln(stderr, mvpPrefix+msg)
		return a, 2
	}
	var extras []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			a.images = append(a.images, args[i+1:]...)
			break
		}
		name, val, explicit := arg, "", false
		if n, v, ok := strings.Cut(arg, "="); ok && !slices.Contains(mvpOptions, arg) && slices.Contains(mvpOptions, n) {
			name, val, explicit = n, v, true
		}
		switch {
		case name == "-h" || name == "--help":
			if explicit {
				return usage("argument -h/--help: ignored explicit argument " + ppRepr(val))
			}
			io.WriteString(stdout, mvpHelp)
			return a, 0
		case slices.Contains(mvpOptions, name):
			if !explicit {
				if i+1 == len(args) || !mvpPositional(args[i+1]) {
					return usage("argument " + name + ": expected one argument")
				}
				i++
				val = args[i]
			}
			if msg := a.set(name, val); msg != "" {
				return usage("argument " + name + ": " + msg)
			}
		case mvpPositional(arg):
			a.images = append(a.images, arg)
		default:
			extras = append(extras, arg)
		}
	}
	if len(a.images) == 0 {
		return usage("the following arguments are required: image_a")
	}
	if len(a.images) > 2 {
		extras = append(extras, a.images[2:]...)
	}
	if len(extras) > 0 {
		return usage("unrecognized arguments: " + strings.Join(extras, " "))
	}
	return a, -1
}

// mvpPositional is argparse's _parse_optional returning None: an argument
// that is not an option string — empty, not starting with '-', a lone '-', a
// negative number (`-\.?\d`, the parser having no option that looks like
// one), or one holding a space. "--" is neither.
func mvpPositional(s string) bool {
	if s == "" || s[0] != '-' || s == "-" {
		return true
	}
	if n, _, _ := strings.Cut(s, "="); s == "--" || slices.Contains(mvpOptions, s) || slices.Contains(mvpOptions, n) {
		return false
	}
	r, _ := utf8.DecodeRuneInString(strings.TrimPrefix(s[1:], "."))
	return unicode.IsDigit(r) || strings.Contains(s, " ")
}

// set stores one option's value, converted as its argparse type did; the
// returned message is argparse's for a value it refused.
func (a *mvpArgs) set(name, val string) string {
	if name == "--props" {
		a.props = val
		return ""
	}
	if name == "--scale" || name == "--edge" || name == "--noise" { // type=float
		f, ok := mvpParseFloat(val)
		if !ok {
			return "invalid float value: " + ppRepr(val)
		}
		switch name {
		case "--scale":
			a.scale = &f
		case "--edge":
			a.edge, a.edgeText = f, pyFloat(f)
		default:
			a.noise = f
		}
		return ""
	}
	box, msg := mvpParseBox(val) // type=parse_box
	switch name {
	case "--region-a":
		a.regionA = box
	case "--region-b":
		a.regionB = box
	case "--ref-a":
		a.refA = box
	default:
		a.refB = box
	}
	return msg
}

// mvpParseFloat is Python's float() on a string: Python whitespace trimmed,
// decimal only (Go's hex floats refused), a signed nan accepted, and an
// overflow read as ±inf rather than refused. strconv already takes Python's
// underscores between digits and its inf/infinity/nan spellings.
func mvpParseFloat(s string) (float64, bool) {
	s = pyStrip(s)
	if strings.ContainsAny(s, "xX") {
		return 0, false
	}
	if len(s) == 4 && (s[0] == '+' || s[0] == '-') && strings.EqualFold(s[1:], "nan") {
		return math.NaN(), true
	}
	f, err := strconv.ParseFloat(s, 64)
	return f, err == nil || errors.Is(err, strconv.ErrRange)
}

// mvpParseBox is parse_box: x,y,w,h of non-negative integers, each trimmed
// of Python whitespace. Digits are ASCII only, where str.isdigit() also
// passed other scripts' digits; a number too large for an int is refused
// with argparse's wording for a type function's ValueError.
func mvpParseBox(text string) (*[4]int, string) {
	parts := strings.Split(text, ",")
	bad := "expected x,y,w,h of non-negative integers, got " + ppRepr(text)
	if len(parts) != 4 {
		return nil, bad
	}
	var b [4]int
	for i, p := range parts {
		p = pyStrip(p)
		if p == "" || strings.Trim(p, "0123456789") != "" {
			return nil, bad
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, "invalid parse_box value: " + ppRepr(text)
		}
		b[i] = n
	}
	if b[2] < 1 || b[3] < 1 {
		return nil, "width and height must be positive, got " + ppRepr(text)
	}
	return &b, ""
}

type mvpBox struct{ left, right, top, bottom, width, height int }

type mvpRegion struct {
	im               *rgbImage
	x0, y0, w, h     int
	edge, noise      float64
	edgeText         string
	bg               [3]byte
	box              mvpBox
	radius           [4]int // top_left, top_right, bottom_left, bottom_right
	fill             [3]byte
	border           [4]int // left, right, top, bottom
	borderColour     [3]byte
	haveBorderColour bool
}

// mvpOpen is Region.__init__: the image read as convert("RGB") yields it, the
// region checked against it, and the background taken — the region's modal
// colour in page mode, else its top-left pixel, every other corner required
// to match it within the noise.
func mvpOpen(env Env, path string, box *[4]int, a mvpArgs, page bool) (*mvpRegion, error) {
	file := path
	if file != "" && !filepath.IsAbs(file) {
		file = env.Dir + "/" + file // unjoined: the OS resolves "..", as Python's open() did
	}
	im, err := decodeRGB(file)
	if err != nil {
		return nil, fmt.Errorf("unreadable image %s: %s", path, pilImageError(err, file, path))
	}
	b := [4]int{0, 0, im.W, im.H}
	if box != nil {
		b = *box
	}
	r := &mvpRegion{im: im, x0: b[0], y0: b[1], w: b[2], h: b[3], edge: a.edge, noise: a.noise, edgeText: a.edgeText}
	if r.w > im.W || r.x0 > im.W-r.w || r.h > im.H || r.y0 > im.H-r.h {
		return nil, fmt.Errorf("region (%d, %d, %d, %d) exceeds %s (%dx%d)", b[0], b[1], b[2], b[3], path, im.W, im.H)
	}
	if page {
		var c mvpCounter
		for y := 0; y < r.h; y++ {
			for x := 0; x < r.w; x++ {
				c.add(r.at(x, y))
			}
		}
		r.bg, _ = c.top()
		return r, nil
	}
	r.bg = r.at(0, 0)
	for _, p := range [][2]int{{r.w - 1, 0}, {0, r.h - 1}, {r.w - 1, r.h - 1}} {
		if mvpDist(r.at(p[0], p[1]), r.bg) > r.noise {
			return nil, mvpUnresolved(fmt.Sprintf("region corner (%d,%d) is not background %s — widen the crop", p[0], p[1], mvpHex(r.bg)))
		}
	}
	return r, nil
}

func (r *mvpRegion) at(x, y int) [3]byte { return r.im.at(r.x0+x, r.y0+y) }

func (r *mvpRegion) isBg(c [3]byte) bool { return mvpDist(c, r.bg) <= r.noise }

// isBlend: c lies on the segment between p and q — antialiasing.
func (r *mvpRegion) isBlend(c, p, q [3]byte) bool {
	return mvpDist(c, p)+mvpDist(c, q) <= mvpDist(p, q)+r.noise
}

// mvpDist is the RGB Euclidean distance: an integer sum of squares, then its
// square root, as math.sqrt(sum(...)) computed it.
func mvpDist(a, b [3]byte) float64 {
	s := 0
	for i := range a {
		d := int(a[i]) - int(b[i])
		s += d * d
	}
	return math.Sqrt(float64(s))
}

func mvpHex(c [3]byte) string { return fmt.Sprintf("#%02x%02x%02x", c[0], c[1], c[2]) }

// mvpCounter is Counter(...).most_common(1): counts in first-seen order, the
// earliest colour winning a tie.
type mvpCounter struct {
	n     map[[3]byte]int
	order [][3]byte
}

func (c *mvpCounter) add(x [3]byte) {
	if c.n == nil {
		c.n = map[[3]byte]int{}
	}
	if c.n[x] == 0 {
		c.order = append(c.order, x)
	}
	c.n[x]++
}

func (c *mvpCounter) top() (colour [3]byte, n int) {
	for _, x := range c.order {
		if c.n[x] > n {
			colour, n = x, c.n[x]
		}
	}
	return colour, n
}

// row and col are the region-relative points of a scan line, both ends
// included, walked from `from` towards `to`.
func (r *mvpRegion) row(y, from, to int) [][2]int {
	var pts [][2]int
	for x, step := from, mvpStep(from, to); x != to+step; x += step {
		pts = append(pts, [2]int{x, y})
	}
	return pts
}

func (r *mvpRegion) col(x, from, to int) [][2]int {
	var pts [][2]int
	for y, step := from, mvpStep(from, to); y != to+step; y += step {
		pts = append(pts, [2]int{x, y})
	}
	return pts
}

func mvpStep(from, to int) int {
	if to >= from {
		return 1
	}
	return -1
}

// firstEdge is the index into pts (outside → inward) of the first hard jump,
// or -1.
func (r *mvpRegion) firstEdge(pts [][2]int) int {
	prev := r.at(pts[0][0], pts[0][1])
	for i := 1; i < len(pts); i++ {
		cur := r.at(pts[i][0], pts[i][1])
		if mvpDist(cur, prev) >= r.edge {
			return i
		}
		prev = cur
	}
	return -1
}

// outward is the region-relative points from just outside the box's side to
// the IMAGE's edge — shadow and gap look past the region, so the region can
// be a tight crop and still measure the spacing to a neighbour outside it.
func (r *mvpRegion) outward(side string) [][2]int {
	b, cx, cy := r.box, r.w/2, r.h/2
	switch side {
	case "left":
		return r.row(cy, b.left-1, -r.x0)
	case "right":
		return r.row(cy, b.right+1, r.im.W-1-r.x0)
	case "top":
		return r.col(cx, b.top-1, -r.y0)
	}
	return r.col(cx, b.bottom+1, r.im.H-1-r.y0)
}

var mvpSides = []string{"left", "right", "top", "bottom"}

func (r *mvpRegion) measureBox() (pyObj, error) {
	cx, cy := r.w/2, r.h/2
	lines := [][][2]int{r.row(cy, 0, cx), r.row(cy, r.w-1, cx), r.col(cx, 0, cy), r.col(cx, r.h-1, cy)}
	var found [4]int
	for k, pts := range lines {
		i := r.firstEdge(pts)
		if i < 0 {
			return nil, mvpUnresolved(fmt.Sprintf("no edge on the centre scan line from the %s (jump >= %s); the region's centre may lie outside the control, or lower --edge", mvpSides[k], r.edgeText))
		}
		found[k] = pts[i][k/2] // x for left/right, y for top/bottom
	}
	b := mvpBox{left: found[0], right: found[1], top: found[2], bottom: found[3]}
	b.width, b.height = b.right-b.left+1, b.bottom-b.top+1
	r.box = b
	return pyObj{{"left", b.left}, {"right", b.right}, {"top", b.top}, {"bottom", b.bottom}, {"width", b.width}, {"height", b.height}}, nil
}

func (r *mvpRegion) measureRadius() (pyObj, error) {
	b, cx := r.box, r.w/2
	limit := min(b.width, b.height)/2 + 1
	corners := []struct {
		name                 string
		y0, dy, from, target int
	}{
		{"top_left", b.top, 1, 0, b.left},
		{"top_right", b.top, 1, r.w - 1, b.right},
		{"bottom_left", b.bottom, -1, 0, b.left},
		{"bottom_right", b.bottom, -1, r.w - 1, b.right},
	}
	out := pyObj{}
	for k, c := range corners {
		// The area between the straight side and the curve is r²(1 − π/4);
		// summing each row's inset from the side and inverting that is
		// robust to antialiasing where counting rows until the inset hits
		// zero is not (it converges ~√(2r) rows early).
		area, reached := 0, false
		for n := 0; n < limit; n++ {
			y := c.y0 + n*c.dy
			if y < 0 || y >= r.h {
				break
			}
			pts := r.row(y, c.from, cx)
			i := r.firstEdge(pts)
			if i < 0 {
				break
			}
			inset := pts[i][0] - c.target
			if inset == 0 {
				reached = true
				break
			}
			area += max(inset, -inset)
		}
		if !reached {
			return nil, mvpUnresolved(fmt.Sprintf("%s corner never reaches the straight side x=%d within %d rows — not a clean rectangle", c.name, c.target, limit))
		}
		r.radius[k] = int(pyRound(math.Sqrt(float64(area)/(1-math.Pi/4)), 0))
		out = append(out, pyKV{c.name, r.radius[k]})
	}
	return out, nil
}

func (r *mvpRegion) interiorInset() (l, t, rt, btm int) {
	b := r.box
	ix, iy := b.width/4, b.height/4
	return b.left + ix, b.top + iy, b.right - ix, b.bottom - iy
}

func (r *mvpRegion) measureFill() pyObj {
	l, t, rt, btm := r.interiorInset()
	var c mvpCounter
	for y := t; y <= btm; y++ {
		for x := l; x <= rt; x++ {
			c.add(r.at(x, y))
		}
	}
	colour, n := c.top()
	r.fill = colour
	return pyObj{{"colour", mvpHex(colour)}, {"share", pyRound(float64(n)/float64((btm-t+1)*(rt-l+1)), 3)}}
}

func (r *mvpRegion) measureBorder() pyObj {
	b, cx, cy := r.box, r.w/2, r.h/2
	lines := [][][2]int{r.row(cy, b.left, cx), r.row(cy, b.right, cx), r.col(cx, b.top, cy), r.col(cx, b.bottom, cy)}
	out := pyObj{}
	for k, pts := range lines {
		var solid [][3]byte
		for _, p := range pts {
			c := r.at(p[0], p[1])
			if mvpDist(c, r.fill) <= r.noise {
				break
			}
			if !r.isBlend(c, r.bg, r.fill) {
				solid = append(solid, c)
			}
		}
		r.border[k] = len(solid)
		out = append(out, pyKV{mvpSides[k], len(solid)})
		if len(solid) > 0 && !r.haveBorderColour {
			r.borderColour, r.haveBorderColour = solid[len(solid)/2], true
		}
	}
	if !r.haveBorderColour {
		return append(out, pyKV{"colour", nil})
	}
	return append(out, pyKV{"colour", mvpHex(r.borderColour)})
}

func (r *mvpRegion) measureShadow() pyObj {
	black := mvpDist(r.bg, [3]byte{})
	if black == 0 {
		black = 1
	}
	out := pyObj{}
	for _, side := range mvpSides {
		width, peak := 0, 0.0
		for _, p := range r.outward(side) {
			d := mvpDist(r.at(p[0], p[1]), r.bg)
			if d <= r.noise {
				break
			}
			width++
			peak = max(peak, d)
		}
		// max() of no deltas was the int 0, and round(0, 1) keeps it an int.
		var peakDelta any = 0
		if width > 0 {
			peakDelta = pyRound(peak, 1)
		}
		out = append(out, pyKV{side, pyObj{{"width", width}, {"peak_delta", peakDelta}, {"approx_opacity", pyRound(math.Min(1.0, peak/black), 2)}}})
	}
	return out
}

func (r *mvpRegion) measureContent() pyObj {
	b, bd, rad := r.box, r.border, r.radius
	l, t := b.left+bd[0]+2, b.top+bd[2]+2
	rt, btm := b.right-bd[1]-2, b.bottom-bd[3]-2
	corners := [4][4]int{ // (x0, y0, x1, y1) squares the curve cuts out of the rectangle
		{b.left, b.top, b.left + rad[0], b.top + rad[0]},
		{b.right - rad[1], b.top, b.right, b.top + rad[1]},
		{b.left, b.bottom - rad[2], b.left + rad[2], b.bottom},
		{b.right - rad[3], b.bottom - rad[3], b.right, b.bottom},
	}
	cl, cr, ct, cb := math.MaxInt, math.MinInt, math.MaxInt, math.MinInt
	var tint mvpCounter
	for y := t; y <= btm; y++ {
	pixel:
		for x := l; x <= rt; x++ {
			for _, k := range corners {
				if k[0] <= x && x <= k[2] && k[1] <= y && y <= k[3] {
					continue pixel
				}
			}
			c := r.at(x, y)
			if mvpDist(c, r.fill) <= r.noise || r.isBlend(c, r.bg, r.fill) || r.haveBorderColour && r.isBlend(c, r.borderColour, r.fill) {
				continue
			}
			cl, cr, ct, cb = min(cl, x), max(cr, x), min(ct, y), max(cb, y)
			tint.add(c)
		}
	}
	if tint.order == nil {
		return pyObj{{"width", nil}, {"height", nil}, {"ratio", nil}, {"padding", nil}, {"colour", nil}}
	}
	w, h := cr-cl+1, cb-ct+1
	// The tint: the modal colour of the content pixels themselves. `fill`
	// is the box behind an icon, never the icon — a grey glyph on the
	// right fill reads as a match on every other property (KAN-437).
	colour, _ := tint.top()
	return pyObj{
		{"width", w},
		{"height", h},
		{"colour", mvpHex(colour)},
		{"ratio", pyObj{{"width", pyRound(float64(w)/float64(b.width), 3)}, {"height", pyRound(float64(h)/float64(b.height), 3)}}},
		{"padding", pyObj{{"left", cl - b.left}, {"top", ct - b.top}, {"right", b.right - cr}, {"bottom", b.bottom - cb}}},
	}
}

func (r *mvpRegion) measureGap() pyObj {
	out := pyObj{}
	for _, side := range mvpSides {
		pts := r.outward(side)
		i := 0
		for i < len(pts) && !r.isBg(r.at(pts[i][0], pts[i][1])) {
			i++ // shadow band
		}
		for i < len(pts) && r.isBg(r.at(pts[i][0], pts[i][1])) {
			i++ // the gap itself
		}
		var gap any
		if i < len(pts) {
			gap = i
		}
		out = append(out, pyKV{side, gap})
	}
	return out
}

func (r *mvpRegion) measureInk() pyObj {
	l, rt, t, btm := math.MaxInt, math.MinInt, math.MaxInt, math.MinInt
	var tint mvpCounter
	for y := 0; y < r.h; y++ {
		for x := 0; x < r.w; x++ {
			if c := r.at(x, y); !r.isBg(c) {
				l, rt, t, btm = min(l, x), max(rt, x), min(t, y), max(btm, y)
				tint.add(c)
			}
		}
	}
	if tint.order == nil {
		return pyObj{{"width", nil}, {"height", nil}, {"colour", nil}}
	}
	// The tint of the ink itself — a text run has no box for `content` to
	// look inside, so this is the one reading of a label's colour (KAN-437
	// final verification: a caption and a summary row shipped in the wrong
	// colour on a frame whose every label had been transcribed).
	colour, _ := tint.top()
	return pyObj{{"left", l}, {"top", t}, {"width", rt - l + 1}, {"height", btm - t + 1}, {"colour", mvpHex(colour)}}
}

func (r *mvpRegion) measureRuns() pyObj {
	cx, cy := r.w/2, r.h/2
	out := pyObj{}
	for _, line := range []struct {
		name string
		pts  [][2]int
	}{{"row", r.row(cy, 0, r.w-1)}, {"col", r.col(cx, 0, r.h-1)}} {
		type run struct {
			from, to int
			colour   [3]byte
		}
		var runs []run
		for i, p := range line.pts {
			c := r.at(p[0], p[1])
			// Compared with the run's FIRST pixel, not the previous one, so a
			// gradient breaks into runs instead of drifting into one.
			if len(runs) > 0 && mvpDist(c, runs[len(runs)-1].colour) <= r.noise {
				runs[len(runs)-1].to = i
			} else {
				runs = append(runs, run{i, i, c})
			}
		}
		list := make([]pyObj, len(runs))
		for i, u := range runs {
			list[i] = pyObj{{"from", u.from}, {"to", u.to}, {"length", u.to - u.from + 1}, {"colour", mvpHex(u.colour)}}
		}
		out = append(out, pyKV{line.name, list})
	}
	return out
}

type mvpBand struct {
	top, height, gapAbove int
	edge, colour          [3]byte
	seams                 []mvpSeam // measureSeams' boxed bands only
	cells                 []mvpCell
}

type mvpSeam struct {
	left, width, gapLeft int
	colour               [3]byte
}

type mvpCell struct {
	left, width int
	lines       []mvpLine
}

type mvpLine struct {
	top, height, left, right int
	offset                   float64
}

func (b mvpBand) obj() pyObj {
	return pyObj{{"top", b.top}, {"height", b.height}, {"gap_above", b.gapAbove}, {"edge", mvpHex(b.edge)}, {"colour", mvpHex(b.colour)}}
}

func (s mvpSeam) obj() pyObj {
	return pyObj{{"left", s.left}, {"width", s.width}, {"gap_left", s.gapLeft}, {"colour", mvpHex(s.colour)}}
}

// seamsObj is a boxed band as `seams` prints it: the band, its seams, its cells.
func (b mvpBand) seamsObj() pyObj {
	seams := make([]pyObj, len(b.seams))
	for i, s := range b.seams {
		seams[i] = s.obj()
	}
	cells := make([]pyObj, len(b.cells))
	for i, c := range b.cells {
		lines := make([]pyObj, len(c.lines))
		for j, ln := range c.lines {
			lines[j] = pyObj{{"top", ln.top}, {"height", ln.height}, {"left", ln.left}, {"right", ln.right}, {"offset", ln.offset}}
		}
		cells[i] = pyObj{{"left", c.left}, {"width", c.width}, {"lines", lines}}
	}
	return append(b.obj(), pyKV{"seams", seams}, pyKV{"cells", cells})
}

func (r *mvpRegion) measureBands() []mvpBand {
	var bands []mvpBand
	prevEnd, y := -1, 0
	for y < r.h {
		var edge mvpCounter
		for x := 0; x < r.w; x++ {
			if c := r.at(x, y); !r.isBg(c) {
				edge.add(c)
			}
		}
		if edge.order == nil {
			y++
			continue
		}
		top := y
		var ink mvpCounter
		for ; y < r.h; y++ {
			inked := false
			for x := 0; x < r.w; x++ {
				if c := r.at(x, y); !r.isBg(c) {
					ink.add(c)
					inked = true
				}
			}
			if !inked {
				break
			}
		}
		e, _ := edge.top()
		c, _ := ink.top()
		bands = append(bands, mvpBand{top: top, height: y - top, gapAbove: top - prevEnd - 1, edge: e, colour: c})
		prevEnd = y - 1
	}
	return bands
}

func (r *mvpRegion) measureSeams(bands []mvpBand) []mvpBand {
	var out []mvpBand
	for _, band := range bands {
		y0, h := band.top, band.height
		cols := make([][][3]byte, r.w)
		first, last := -1, -1
		for x := range cols {
			for y := y0; y < y0+h; y++ {
				if c := r.at(x, y); !r.isBg(c) {
					cols[x] = append(cols[x], c)
				}
			}
			if cols[x] != nil {
				if first < 0 {
					first = x
				}
				last = x
			}
		}
		// A boxed band's first row is one unbroken run — a border or a
		// fill — across the band's span. A text row's first row is not:
		// even 8px small caps whose flat tops ink most of the span are
		// separate glyphs with a background gap between letters, so its
		// longest run is a glyph or two. Counting inked pixels missed
		// that, and read GOAL/PROTEIN captions as controls whose every
		// stem became a seam.
		run, longest := 0, 0
		for x := first; x <= last; x++ {
			run++
			if r.isBg(r.at(x, y0)) {
				run = 0
			}
			longest = max(longest, run)
		}
		if float64(longest) < mvpFull*float64(last-first+1) {
			continue // a bare text row: its glyph stems are not seams
		}
		type span struct {
			left, right int
			rgb         [3]byte
		}
		var seams []span
		for x, col := range cols {
			if float64(len(col)) < mvpFull*float64(h) {
				continue
			}
			var c mvpCounter
			for _, p := range col {
				c.add(p)
			}
			colour, _ := c.top()
			if n := len(seams); n > 0 && seams[n-1].right == x-1 && mvpDist(colour, seams[n-1].rgb) < r.edge {
				seams[n-1].right = x
			} else {
				seams = append(seams, span{x, x, colour})
			}
		}
		for k := 0; k+1 < len(seams); k++ {
			x0, x1 := seams[k].right+1, seams[k+1].left-1
			if x1 < x0 {
				continue
			}
			cell := mvpCell{left: x0, width: x1 - x0 + 1}
			var line *mvpLine
			for y := y0; y < y0+h; y++ {
				l, rt, n := -1, -1, 0
				for x := x0; x <= x1; x++ {
					if !r.isBg(r.at(x, y)) {
						if l < 0 {
							l = x
						}
						rt, n = x, n+1
					}
				}
				if n == 0 || float64(n) >= mvpFull*float64(cell.width) { // empty, or a rule across the cell
					line = nil
					continue
				}
				if line == nil {
					cell.lines = append(cell.lines, mvpLine{top: y, left: l, right: rt})
					line = &cell.lines[len(cell.lines)-1]
				}
				line.height++
				line.left, line.right = min(line.left, l), max(line.right, rt)
			}
			// Absolute columns become each line's inset from the cell's edges
			// and its centre's offset from the cell's centre.
			for i, ln := range cell.lines {
				cell.lines[i] = mvpLine{top: ln.top - y0, height: ln.height, left: ln.left - x0, right: x1 - ln.right,
					offset: pyRound(float64(ln.left+ln.right)/2-float64(x0+x1)/2, 1)}
			}
			band.cells = append(band.cells, cell)
		}
		prev := -1
		for _, s := range seams {
			band.seams = append(band.seams, mvpSeam{left: s.left, width: s.right - s.left + 1, gapLeft: s.left - prev - 1, colour: s.rgb})
			prev = s.right
		}
		out = append(out, band)
	}
	return out
}

// mvpMeasured is one image's measurement: the printed object, plus the bands
// and boxed bands the delta pairs.
type mvpMeasured struct {
	out          pyObj
	bands, seams []mvpBand
}

func (r *mvpRegion) measure(props []string) (mvpMeasured, error) {
	has := func(names ...string) bool {
		return slices.ContainsFunc(names, func(n string) bool { return slices.Contains(props, n) })
	}
	m := mvpMeasured{out: pyObj{{"background", mvpHex(r.bg)}}}
	needsBox := slices.ContainsFunc(props, func(p string) bool { return !slices.Contains([]string{"ink", "runs", "bands", "seams"}, p) })
	if needsBox {
		box, err := r.measureBox()
		if err != nil {
			return m, err
		}
		if has("box") {
			m.out = append(m.out, pyKV{"box", box})
		}
		if has("radius", "content") {
			radius, err := r.measureRadius()
			if err != nil {
				return m, err
			}
			if has("radius") {
				m.out = append(m.out, pyKV{"radius", radius})
			}
		}
		if has("fill", "border", "content") {
			if fill := r.measureFill(); has("fill") {
				m.out = append(m.out, pyKV{"fill", fill})
			}
		}
		if has("border", "content") {
			if border := r.measureBorder(); has("border") {
				m.out = append(m.out, pyKV{"border", border})
			}
		}
		if has("shadow") {
			m.out = append(m.out, pyKV{"shadow", r.measureShadow()})
		}
		if has("content") {
			m.out = append(m.out, pyKV{"content", r.measureContent()})
		}
		if has("gap") {
			m.out = append(m.out, pyKV{"gap", r.measureGap()})
		}
	}
	if has("ink") {
		m.out = append(m.out, pyKV{"ink", r.measureInk()})
	}
	if has("runs") {
		m.out = append(m.out, pyKV{"runs", r.measureRuns()})
	}
	if has("bands", "seams") {
		m.bands = r.measureBands()
		if has("bands") {
			list := make([]pyObj, len(m.bands))
			for i, b := range m.bands {
				list[i] = b.obj()
			}
			m.out = append(m.out, pyKV{"bands", list})
		}
		if has("seams") {
			m.seams = r.measureSeams(m.bands)
			list := make([]pyObj, len(m.seams))
			for i, b := range m.seams {
				list[i] = b.seamsObj()
			}
			m.out = append(m.out, pyKV{"seams", list})
		}
	}
	return m, nil
}

// mvpLeaves is every non-object value under d, keyed by its dotted path.
func mvpLeaves(d pyObj, prefix string) []pyKV {
	var out []pyKV
	for _, kv := range d {
		if sub, ok := kv.v.(pyObj); ok {
			out = append(out, mvpLeaves(sub, prefix+kv.k+".")...)
		} else {
			out = append(out, pyKV{prefix + kv.k, kv.v})
		}
	}
	return out
}

func mvpColourDelta(av, bv string) pyObj {
	parse := func(h string) (c [3]byte) {
		for i := range c {
			n, _ := strconv.ParseUint(h[1+2*i:3+2*i], 16, 8)
			c[i] = byte(n)
		}
		return c
	}
	return pyObj{{"a", av}, {"b", bv}, {"distance", pyRound(mvpDist(parse(av), parse(bv)), 1)}}
}

// Python's arithmetic on an int or a float64: int op int stays an int, any
// float makes a float — the promotion number_delta's output types follow.
func mvpF(v any) float64 {
	if i, ok := v.(int); ok {
		return float64(i)
	}
	return v.(float64)
}

func mvpMul(a, b any) any {
	ai, aok := a.(int)
	bi, bok := b.(int)
	if aok && bok {
		return ai * bi
	}
	return mvpF(a) * mvpF(b)
}

func mvpSub(a, b any) any {
	ai, aok := a.(int)
	bi, bok := b.(int)
	if aok && bok {
		return ai - bi
	}
	return mvpF(a) - mvpF(b)
}

// mvpRoundN is round(v, n): an int comes back unchanged, a float rounded.
func mvpRoundN(v any, n int) any {
	if f, ok := v.(float64); ok {
		return pyRound(f, n)
	}
	return v
}

// mvpNumberDelta is number_delta; scale is the int 1 for an unscaled leaf,
// else the float scale.
func mvpNumberDelta(av, bv, scale any) pyObj {
	scaled := mvpMul(av, scale)
	ab := mvpSub(bv, scaled)
	var pct any
	if mvpF(scaled) != 0 {
		pct = pyRound(mvpF(ab)/mvpF(scaled)*100, 1)
	}
	return pyObj{{"a", av}, {"a_scaled", mvpRoundN(scaled, 2)}, {"b", bv}, {"abs", mvpRoundN(ab, 2)}, {"pct", pct}}
}

// mvpRec is one band or seam as pairInOrder reads it.
type mvpRec struct {
	pos, size, gap int
	colours        []string
	obj            pyObj
}

// mvpPair is one pairInOrder entry: i and j index the records it holds, -1
// for the side it lacks.
type mvpPair struct {
	status string
	i, j   int
	obj    pyObj
}

func mvpBandRecs(bands []mvpBand) []mvpRec {
	recs := make([]mvpRec, len(bands))
	for i, b := range bands {
		recs[i] = mvpRec{b.top, b.height, b.gapAbove, []string{mvpHex(b.edge), mvpHex(b.colour)}, b.obj()}
	}
	return recs
}

func mvpSeamRecs(seams []mvpSeam) []mvpRec {
	recs := make([]mvpRec, len(seams))
	for i, s := range seams {
		recs[i] = mvpRec{s.left, s.width, s.gapLeft, []string{mvpHex(s.colour)}, s.obj()}
	}
	return recs
}

// mvpPairInOrder pairs the frame's bands (or seams) with the capture's in
// order along one axis — gapKey/sizeKey/colourKeys name the printed keys. A
// record pairs when its gap since the last pair and its size both match
// within a tolerance of 8px plus a quarter of the larger record, so a padding
// defect shows once as that pair's gap delta and the records after it still
// pair, instead of every one after reading as shifted — and a 2px rule never
// pairs with a text row that happens to sit where the rule was.
//
// ponytail: greedy in-order pairing; a shift larger than the tolerance
// unpairs one record and resyncs on the next — enough for a page of rows,
// replace with sequence alignment if frames with many near-identical bands
// mis-pair.
func mvpPairInOrder(a, b []mvpRec, scale float64, gapKey, sizeKey string, colourKeys []string) []mvpPair {
	var out []mvpPair
	missing := func(i int) mvpPair {
		return mvpPair{"missing", i, -1, pyObj{{"status", "missing"}, {"a", a[i].obj}, {"b", nil}}}
	}
	extra := func(j int) mvpPair {
		return mvpPair{"extra", -1, j, pyObj{{"status", "extra"}, {"a", nil}, {"b", b[j].obj}}}
	}
	i, j := 0, 0
	aAnchor, bAnchor := -1, -1 // last row/column of the last paired record, per image
	for i < len(a) || j < len(b) {
		switch {
		case i < len(a) && j < len(b):
			ga := float64(a[i].pos-aAnchor-1) * scale
			gb := b[j].pos - bAnchor - 1
			tol := 8 + math.Max(float64(a[i].size)*scale, float64(b[j].size))/4
			if math.Abs(float64(gb)-ga) <= tol && math.Abs(float64(b[j].size)-float64(a[i].size)*scale) <= tol {
				obj := pyObj{{"status", "paired"}, {"a", a[i].obj}, {"b", b[j].obj}, {gapKey, mvpNumberDelta(a[i].gap, b[j].gap, scale)},
					// Distance from the last PAIRED record, so a padding lost
					// around a missing divider still reads as one number here.
					{"since_pair", mvpNumberDelta(a[i].pos-aAnchor-1, gb, scale)},
					{sizeKey, mvpNumberDelta(a[i].size, b[j].size, scale)}}
				for k, key := range colourKeys {
					obj = append(obj, pyKV{key, mvpColourDelta(a[i].colours[k], b[j].colours[k])})
				}
				out = append(out, mvpPair{"paired", i, j, obj})
				aAnchor, bAnchor = a[i].pos+a[i].size-1, b[j].pos+b[j].size-1
				i, j = i+1, j+1
			} else if ga < float64(gb) {
				out, i = append(out, missing(i)), i+1
			} else {
				out, j = append(out, extra(j)), j+1
			}
		case i < len(a):
			out, i = append(out, missing(i)), i+1
		default:
			out, j = append(out, extra(j)), j+1
		}
	}
	return out
}

func mvpPairObjs(pairs []mvpPair) []pyObj {
	objs := make([]pyObj, len(pairs))
	for i, p := range pairs {
		objs[i] = p.obj
	}
	return objs
}

func mvpCountStatus(pairs []mvpPair) (paired, missing, extra int) {
	for _, p := range pairs {
		switch p.status {
		case "paired":
			paired++
		case "missing":
			missing++
		default:
			extra++
		}
	}
	return paired, missing, extra
}

// mvpPairSeams pairs the boxed bands, then each band's seams; a cell pairs
// where its two bounding seams paired with consecutive seams in the other
// image, and its lines pair by index.
func mvpPairSeams(a, b []mvpBand, scale float64) ([]pyObj, pyObj) {
	var out []pyObj
	var paired, missing, extra, bandsUnpaired, linesPaired, linesUnpaired int
	for _, band := range mvpPairInOrder(mvpBandRecs(a), mvpBandRecs(b), scale, "gap_above", "height", []string{"edge", "colour"}) {
		var ea, eb any
		if band.i >= 0 {
			ea = pyObj{{"top", a[band.i].top}, {"height", a[band.i].height}}
		}
		if band.j >= 0 {
			eb = pyObj{{"top", b[band.j].top}, {"height", b[band.j].height}}
		}
		entry := pyObj{{"status", band.status}, {"a", ea}, {"b", eb}}
		if band.status != "paired" {
			bandsUnpaired++
			out = append(out, entry)
			continue
		}
		ba, bb := a[band.i], b[band.j]
		seams := mvpPairInOrder(mvpSeamRecs(ba.seams), mvpSeamRecs(bb.seams), scale, "gap_left", "width", []string{"colour"})
		p, m, e := mvpCountStatus(seams)
		paired, missing, extra = paired+p, missing+m, extra+e
		var idx [][2]int // each paired seam's index in its own list, in order
		for _, s := range seams {
			if s.status == "paired" {
				idx = append(idx, [2]int{s.i, s.j})
			}
		}
		// Cells keyed by their left column: two touching seams bound no
		// cell, so cell index and seam index disagree past the first pair.
		cellAt := func(cells []mvpCell, left int) (mvpCell, bool) {
			for _, c := range cells {
				if c.left == left {
					return c, true
				}
			}
			return mvpCell{}, false
		}
		cells := []pyObj{}
		for k := 0; k+1 < len(idx); k++ {
			ia, ib, ja, jb := idx[k][0], idx[k][1], idx[k+1][0], idx[k+1][1]
			if ja != ia+1 || jb != ib+1 {
				continue
			}
			ca, okA := cellAt(ba.cells, ba.seams[ia].left+ba.seams[ia].width)
			cb, okB := cellAt(bb.cells, bb.seams[ib].left+bb.seams[ib].width)
			if !okA || !okB {
				continue
			}
			lines := []pyObj{}
			for n := 0; n < min(len(ca.lines), len(cb.lines)); n++ {
				la, lb := ca.lines[n], cb.lines[n]
				lines = append(lines, pyObj{{"offset", mvpNumberDelta(la.offset, lb.offset, scale)},
					{"left", mvpNumberDelta(la.left, lb.left, scale)}, {"right", mvpNumberDelta(la.right, lb.right, scale)}})
			}
			linesPaired += len(lines)
			linesUnpaired += max(len(ca.lines)-len(cb.lines), len(cb.lines)-len(ca.lines))
			cells = append(cells, pyObj{{"a", ia}, {"b", ib}, {"count", pyObj{{"a", len(ca.lines)}, {"b", len(cb.lines)}}}, {"lines", lines}})
		}
		out = append(out, append(entry, pyKV{"seams", mvpPairObjs(seams)}, pyKV{"cells", cells}))
	}
	return out, pyObj{{"paired", paired}, {"missing", missing}, {"extra", extra}, {"bands_unpaired", bandsUnpaired},
		{"lines", pyObj{{"paired", linesPaired}, {"unpaired", linesUnpaired}}}}
}

func mvpDelta(a, b mvpMeasured, props []string, scale float64) pyObj {
	out := pyObj{}
	if slices.Contains(props, "bands") {
		pairs := mvpPairInOrder(mvpBandRecs(a.bands), mvpBandRecs(b.bands), scale, "gap_above", "height", []string{"edge", "colour"})
		p, m, e := mvpCountStatus(pairs)
		out = append(out, pyKV{"bands", mvpPairObjs(pairs)}, pyKV{"bands_summary", pyObj{{"paired", p}, {"missing", m}, {"extra", e}}})
	}
	if slices.Contains(props, "seams") {
		seams, summary := mvpPairSeams(a.seams, b.seams, scale)
		out = append(out, pyKV{"seams", seams}, pyKV{"seams_summary", summary})
	}
	bl := map[string]any{}
	for _, kv := range mvpLeaves(b.out, "") {
		bl[kv.k] = kv.v
	}
	isNum := func(v any) bool {
		switch v.(type) {
		case int, float64:
			return true
		}
		return false
	}
	for _, kv := range mvpLeaves(a.out, "") {
		key, av, bv := kv.k, kv.v, bl[kv.k]
		if key == "bands" || key == "seams" {
			continue
		}
		as, aok := av.(string)
		bs, bok := bv.(string)
		switch {
		case aok && bok && strings.HasPrefix(as, "#") && strings.HasPrefix(bs, "#"):
			out = append(out, pyKV{key, mvpColourDelta(as, bs)})
		case isNum(av) && isNum(bv):
			var s any = scale
			if slices.ContainsFunc(strings.Split(key, "."), func(seg string) bool { return slices.Contains(mvpUnscaled, seg) }) {
				s = 1
			}
			out = append(out, pyKV{key, mvpNumberDelta(av, bv, s)})
		case av == nil || bv == nil:
			out = append(out, pyKV{key, pyObj{{"a", av}, {"b", bv}, {"abs", nil}, {"pct", nil}}})
		}
	}
	return out
}
