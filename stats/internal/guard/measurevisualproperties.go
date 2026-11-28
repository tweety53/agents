package guard

import (
	"bytes"
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"math"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/tweety53/agents/stats/internal/guard/core"
)

// The contract below moved here from the header of the deleted
// scripts/measure-visual-properties.sh shim (KAN-1058).
//
// flow-guard measure-visual-properties — measure one UI control's visual properties
// from pixels, in one image or in two (a mockup crop and the live capture of the
// same control), and report the numbers plus the calibrated delta as JSON.
//
// Interface:
//
//   flow-guard measure-visual-properties <image A> [<image B>] [options]
//
//     --region-a x,y,w,h    crop of A to measure (default: the whole image)
//     --region-b x,y,w,h    crop of B to measure (default: the whole image)
//     --scale <f>           calibration: B pixels per A pixel (a project's
//                           `mockup frame` scale is 1 here when A is the frame
//                           the composite already cropped and B its capture)
//     --ref-a x,y,w,h       calibration from two boxes already known correct in
//     --ref-b x,y,w,h       both images: scale = ref-b width / ref-a width
//                           (heights must agree with that factor within 5%)
//     --props <list>        comma-separated subset of box,radius,border,fill,
//                           shadow,content,gap,ink,runs,bands,seams (default: all),
//                           or the opt-in elements and matrix
//     --edge <n>            adjacent-pixel colour jump (RGB Euclidean, 0-441)
//                           that counts as a hard edge (default 24)
//     --noise <n>           colour distance to background below which a pixel
//                           is background (default 8)
//     --tolerance <px>      the length tolerance of `--props matrix`'s
//                           `summary.over` (default 2, mvpOverLength); never
//                           applied to a colour, a missing or an extra row
//
//   With two images one calibration input is mandatory — `--scale` or the
//   `--ref-*` pair; a cross-image comparison with neither is refused, because an
//   uncalibrated ruler is itself an unchecked claim (KAN-30 fix round 9).
//
// Each region is a crop around ONE control with a background margin on all
// four sides and nothing else in it — the background colour is the region's
// top-left pixel, and all four corners must be background. Everything is
// measured on the centre row and centre column of the region, from the outside
// inward, so the region's centre must fall inside the control. Only `shadow`
// and `gap` look past the region, out to the image's own edge, so a neighbour
// belongs outside the region, not in it.
//
// A control that runs off the image — a full-width bar, a card in the frame's
// corner — has no background margin on that side. A given `--region-*` whose
// side lies on the image's own edge, with that side's end of the centre scan
// line not background, is `clipped` there: the box edge on that side is the
// image's edge, the corners touching it need not be background (the
// background is then the first other corner colour that leaves every
// unclipped corner background, when the top-left pixel is not it), and `border`, `shadow` and `gap` on that side and `radius` on its
// two corners read the string "clipped" in place of a number — the other
// sides still measure. The output's `clipped` key lists the clipped sides,
// comma-joined, and is absent when none is; a clipped leaf's `delta` carries
// both readings and no `abs`. The whole image, the default region, clips
// nothing. `ink` and `runs` need no control edge
// at all and are the two properties to run before the crop is trusted.
//
// Properties (pixels of the region's own image; `null` where not found):
//
//   box      left/top/right/bottom edge coordinates (region-relative) and
//            width/height. An edge is the first pixel, walking inward along
//            the centre scan line, whose colour jumps from its neighbour by at
//            least `--edge`; a soft shadow gradient never jumps that far in
//            one pixel, a hard edge (with or without antialiasing) does.
//   radius   corner radius per corner: rows walked from the top (bottom) edge
//            until the row's own edge reaches the box's straight side; the
//            area those rows leave between the side and the curve is
//            r²(1 − π/4), inverted for r. 0 for a square corner.
//   border   per-side width in px and the border colour: pixels from the edge
//            inward that are neither the fill nor a background/fill blend.
//   fill     modal colour of the interior (inset a quarter of the box, so the
//            border and edge antialiasing are excluded) and its share of those
//            pixels.
//   shadow   per side: width of the band outside the edge whose pixels differ
//            from the background by more than `--noise`, the band's peak
//            distance from background, and an approximate opacity assuming a
//            black shadow. A flat design is width 0 on every side.
//   content  bounding box of non-fill pixels inside the border and outside the
//            corner curves (an icon, a glyph run, a label), its size, its size
//            as a ratio of the box, the padding from each box edge to it, and
//            `colour` — the modal colour of those pixels, the icon's or text's
//            tint. `fill` is the box behind the glyph, never the glyph: a grey
//            icon where the mockup draws an accent one matches on box, radius,
//            fill and content size alike, and only this colour's `delta`
//            distance sees it (KAN-437).
//   gap      per side: background pixels between the box edge (past any
//            shadow band) and the next non-background pixel on the same scan
//            line, out to the image's edge — the spacing to the nearest
//            neighbour; `null` when the scan line reaches the image edge.
//   ink      bounding box of every non-background pixel in the region, and
//            `colour` — the modal colour of those pixels. Needs no control
//            edge: crop to a run of capital letters and request `--props ink`
//            — its height is the cap-height, the font-size stand-in; its
//            `left`/`top` in a crop spanning the container are the run's
//            position and, against the crop's width, its alignment; its
//            `colour` is the text's tint, which no other property reads on a
//            run with no box (KAN-437).
//   runs     colour runs along the region's centre row and centre column:
//            `from`/`to` (region-relative), `length` and `colour` of every
//            stretch of pixels within `--noise` of the stretch's first pixel.
//            The scanline, needing no clean box. Two readings the other
//            properties cannot give: (1) an edge-to-edge claim — a fill flush
//            with its container's border is the run immediately after the
//            border's run, and an inset fill shows as the container-coloured
//            run between them, its `length` the inset in px (`content` cannot
//            answer this: it boxes every non-fill pixel in the container,
//            other rows' text included, and a fill covering half the box
//            flips which colour counts as `fill` — KAN-30 manual re-sweep);
//            (2) where to crop — every run boundary is an edge a region can
//            sit on, read before any property that needs a box. No `delta`:
//            compare the two lists by eye.
//   bands    the region's horizontal structure, whole-page: every maximal run
//            of rows holding a non-background pixel is a band — a text line,
//            a control, a card, a 1px divider — with `top`, `height`,
//            `gap_above` (background rows since the previous band), `edge`
//            (the modal non-background colour of the band's first row: a
//            control's border or a card's fill, whichever the eye meets
//            first), `colour` (the darkest non-background colour holding at
//            least 1% of the band's ink — its text, not a wider border's grey;
//            kan-1017 read a modal colour that flipped between the two) and
//            `share` (the two most frequent ink colours' share of the band's
//            ink, to 3 places). Background is the most frequent colour of the
//            region's outermost rows and columns, not its corner, so a card
//            larger than the page around it is ink, not background (kan-1083
//            self-review, F18); no corner needs to be background: the region is the whole
//            capture and the whole cropped frame. Unlike every other
//            property, `bands` has a structural `delta`: the two lists are
//            paired in order by the gap since the last pair, and every band
//            the frame draws with no counterpart in the capture is listed
//            `missing`, every capture band the frame lacks `extra`, and each
//            pair carries its `gap_above`, `since_pair` (rows since the last
//            paired band — the one number a lost padding shows as when the
//            divider beside it is also missing), `height`, `edge` and
//            `colour` deltas. Text content differing between the two does not
//            unpair a band — a band is geometry and non-background ink, never
//            what the text says — so the departures this lists are exactly
//            the ones a data difference cannot explain: a divider the frame
//            draws and the capture omits, a row padding the capture lacks, a
//            surface fill swapped for the page colour, a button border in the
//            wrong colour (KAN-437 fix round 5: all four passed a composite
//            whose diff ratio was already 0.3 from data and fonts alone).
//            Vertical structure — a segmented control's inter-cell dividers,
//            a wrapped label's internal alignment — is not a band; `seams`
//            reads those.
//   seams    the vertical structure inside each BOXED band — a band whose
//            first row's ink spans at least half the band's ink width: a
//            bordered or filled control, a card, a hairline; never a bare
//            text row, whose glyph stems would otherwise read as seams. Per
//            boxed band (its `top`, `height`, `gap_above`, `edge`, `colour`
//            as in `bands`): `seams`, every maximal run of columns inked over
//            at least 90% of the band's height and of one modal colour
//            (adjacent columns further than `--edge` apart start a new seam)
//            — an outer border's side, a cell divider, a filled cell — with
//            `left`, `width`, `gap_left` (columns since the previous seam)
//            and `colour`; and `cells`, the spans between consecutive seams,
//            each with `left`, `width` and its `lines`: every maximal run of
//            rows carrying ink inside the span, rows inked across 90% of the
//            span (a border, an underline) excluded, with `top`, `height`,
//            `left` and `right` (the ink's inset from each cell edge) and
//            `offset` (the ink's centre minus the cell's centre, negative
//            when left of it). `delta.seams` pairs the boxed bands as `bands`
//            does, then within each paired band pairs the seams in order by
//            `since_pair` (columns since the last paired seam) and `width`,
//            listing `missing` and `extra` seams and per pair the `gap_left`,
//            `since_pair`, `width` and `colour` deltas; cells pair where both
//            bounding seams paired consecutively, and their lines pair by
//            index with `count` per image and per pair the `offset`, `left`
//            and `right` deltas. Text width never moves a seam, so a missing
//            seam is a divider the capture omits; a wrapped label whose lines
//            are centred in the frame and left-anchored in the capture reads
//            as an `offset` delta of half the slack with a `left` delta of the
//            same size, while a left-anchored label whose width is data reads
//            as an `offset` delta with `left` unchanged (KAN-437 fix round 5:
//            the GOAL control's two dividers and its wrapped labels' centring
//            both shipped wrong past every band and every sweep). A filled
//            cell is a seam, not a cell — its lines are not read; a divider in
//            a row with no border or fill sits in an unboxed band and is not
//            read either. A control whose outer border the capture omits is
//            therefore not boxed there at all: it is one of `bands_unpaired`,
//            never a list of missing seams, and its labels' centring goes
//            unread until the border is back (the real KAN-437 pre-fix
//            capture: GOAL and PROTEIN each read as a plain band of the same
//            height and gap as the frame's boxed one).
//   elements — the app surface's top (`surface_top`: the first row, walking
//            column 3 up from the bottom, whose mean channel value is not
//            above 200, plus one; 0 when that row is above row 60 — the whole
//            image is the surface) and every ink run below it against the
//            surface's modal colour (distance over 30), each run's box, modal
//            colour, kind and gap above; paired in order across two images
//            like `bands`. Not in the default set. A flat scrim reaching the
//            top-left corner is cropped to its sheet before `bands` and `seams`;
//            a corner colour covering 5% or more of the sheet's bounding box is
//            a dark page, never a scrim.
//   matrix — opt-in, two whole images (no `--region-*`, refused with it):
//            beside other props, its `matrix` and `summary` follow their
//            output, each key as the separate calls print it. `elements`
//            on both, paired in order, then each paired element measured on
//            its own region (its box plus a 4px margin, clipped to the image)
//            with its kind's columns — `text/control` `ink`, plus the box
//            columns when its region resolves; `fill band` `box`, `radius`,
//            `border`, `fill`, `shadow`, `content`, `gap`; `rule` `runs`.
//            Output is `scale`, `matrix` (one row per pair: `kind`,
//            `class`, the pairing, and `measured` with `a`, `b` and `delta`,
//            or `unresolved` with the reason; an unpaired element is its
//            `missing` or `extra` row, carrying `class` and no `kind`) and
//            `summary` (kan-1017: every matrix
//            was filled one element per call): `paired`, `missing`, `extra`
//            and `over` (decision over-list) — one `{row, status, keys}`
//            entry per matrix row, `row` its 1-based index in `matrix`, that
//            is `missing` or `extra`, carries `unresolved`, or is paired with
//            a scaled numeric leaf of its pairing or its `measured.delta`
//            whose `abs` is over `--tolerance` px or a colour `distance`
//            over 4 (mvpOverColour) — step 10's band and seam tolerances.
//            `shift` is present only for a frame moved whole — `{dy, rows}`:
//            the elements voted one vertical offset over 8px and pairing by
//            position under it paired more rows than pairing by gap, at
//            least half of the shorter side; the matrix is then that
//            pairing's, `rows` its paired count, and the first paired row
//            carries `dy` on `since_pair`.
//            `within` lists every other paired row carrying a non-zero length
//            delta, as `{row, values}`, `values` its `<leaf>=<abs>` deltas.
//            A leaf under an unscaled key (`ratio`, `approx_opacity`,
//            `peak_delta`, `share`) is never over. `keys` names the leaves
//            over, `, `-separated (`unresolved` for an unresolved row, empty
//            for a missing or extra one). `class` is the row's matrix class
//            (decision matrix-row-class): `missing` or `extra` for an unpaired
//            row, `unresolved` for one carrying `unresolved`, else `control`
//            for a `text/control` element whose box columns resolved, `text
//            run` for one whose did not, and `fill band` or `rule` as its kind.
//
// A given `--region-*` that starts inside the image but runs past its edge is
// clamped to the image and reported `clipped` on that side; one starting
// outside the image is refused (exit 2).
//
// Output (stdout): one JSON object — `a`, `b` (when given), `scale`, and
// `delta`: for every numeric leaf, `a`, `a_scaled` (`a` × scale for lengths,
// `a` unchanged for ratios and opacities), `b`, `abs` (b − a_scaled) and `pct`
// (abs / a_scaled × 100, null when a_scaled is 0); for every colour, the RGB
// Euclidean distance; for `bands`, `delta.bands` is the paired list above and
// `delta.bands_summary` its `paired`/`missing`/`extra` counts; for `seams`,
// `delta.seams` is the per-band paired list above and `delta.seams_summary`
// its `paired`/`missing`/`extra` seam counts, `bands_unpaired` (boxed bands
// with no counterpart) and `lines` (`paired`/`unpaired`). Thresholds are
// the caller's: this script flags nothing.
//
// Exit codes:
//   0  measured.
//   1  a region cannot be resolved cleanly: a corner that is not background, a
//      scan line that finds no edge, a corner scan that never reaches the
//      straight side, or the region's centre outside the box. The reason is on
//      stderr; nothing is guessed.
//   2  cannot answer: usage error, two images with no calibration, a `--scale`
//      that is not a positive finite number or a measurement that overflows
//      to no finite number, `--ref-*` boxes whose aspect disagrees, a region starting outside its image,
//      `matrix` without two images or with a `--region-*`, or
//      an unreadable image (anything but a PNG
//      included).
//
// The body is Go: stats/internal/guard/measurevisualproperties.go. Option names
// are exact — argparse's prefix abbreviations are not accepted — and a usage
// error is one `measure-visual-properties: …` line on stderr.

// measure-visual-properties: one UI control's visual properties measured from
// pixels, in one image or two, printed as indented JSON (encoding/json), each
// object's keys in insertion order (mvpObj). The interface, the region layout,
// every property and the exit codes are the comment above's; the comments here are the algorithm's.

func init() {
	Registry["measure-visual-properties"] = measureVisualProperties
}

const mvpPrefix = "measure-visual-properties: "

var mvpAllProps = []string{"box", "radius", "border", "fill", "shadow", "content", "gap", "ink", "runs", "bands", "seams"}

// Properties that read the region as a whole page: background is the modal
// colour and the corners need not be background.
var mvpPageProps = []string{"bands", "seams", "elements", "matrix"}

// mvpOptIn are properties --props accepts that the default set leaves out.
var mvpOptIn = []string{"elements", "matrix"}

// mvpFull is the share of a band's height a column must ink to be a seam, and
// of a cell's width a row must ink to be a rule rather than a text line.
const mvpFull = 0.9

// Numeric leaves that are not lengths and therefore are not scaled.
var mvpUnscaled = []string{"ratio", "approx_opacity", "peak_delta", "share"}

// mvpUnresolved is a region that cannot be resolved to a clean bounding box:
// exit 1, the reason on stderr, nothing guessed.
type mvpUnresolved string

func (e mvpUnresolved) Error() string { return string(e) }

// mvpHelp is what argparse prints for -h with stdout not a terminal (80
// columns) and prog "flow-guard measure-visual-properties": the port prints
// the same text wherever it runs. `--tolerance` is the port's own option, added
// after the Python it is pinned against, and is documented in the header
// above alone.
const mvpHelp = `usage: flow-guard measure-visual-properties [-h] [--region-a REGION_A]
                                            [--region-b REGION_B]
                                            [--scale SCALE] [--ref-a REF_A]
                                            [--ref-b REF_B] [--props PROPS]
                                            [--edge EDGE] [--noise NOISE]
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

var mvpOptions = []string{"-h", "--help", "--region-a", "--region-b", "--scale", "--ref-a", "--ref-b", "--props", "--edge", "--noise", "--tolerance"}

type mvpArgs struct {
	images                       []string
	regionA, regionB, refA, refB *[4]int
	scale                        *float64
	props                        string
	edge, noise                  float64
	tolerance                    float64              // --tolerance: the length tolerance of summary.over, mvpOverLength by default
	decoded                      map[string]*rgbImage // each image file decoded once per call; regions only read it
}

func measureVisualProperties(args []string, env core.Env, stdout, stderr io.Writer) int {
	a, code := mvpParse(args, env, stdout, stderr)
	if code >= 0 {
		return code
	}
	refuse := func(id, msg string) int {
		core.Reason(env, id)
		fmt.Fprintln(stderr, mvpPrefix+msg)
		return 2
	}

	var props, unknown []string
	for _, s := range strings.Split(a.props, ",") {
		if s = core.PyStrip(s); s != "" {
			props = append(props, s)
			if !slices.Contains(mvpAllProps, s) && !slices.Contains(mvpOptIn, s) {
				unknown = append(unknown, s)
			}
		}
	}
	if len(unknown) > 0 || len(props) == 0 {
		quoted := make([]string, len(unknown))
		for i, s := range unknown {
			quoted[i] = ppRepr(s)
		}
		return refuse("unknown-props", fmt.Sprintf("unknown --props [%s]; choose from %s", strings.Join(quoted, ", "), strings.Join(mvpAllProps, ",")))
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
			return refuse("scale-and-refs-both-given", "give --scale or the --ref-a/--ref-b pair, not both")
		case a.scale != nil:
			if math.IsNaN(*a.scale) || math.IsInf(*a.scale, 0) {
				return refuse("scale-not-finite", "--scale must be finite")
			}
			if *a.scale <= 0 {
				return refuse("scale-must-be-positive", "--scale must be positive")
			}
			scale = *a.scale
		case a.refA != nil && a.refB != nil:
			scale = float64(a.refB[2]) / float64(a.refA[2])
			byHeight := float64(a.refB[3]) / float64(a.refA[3])
			if math.Abs(byHeight-scale)/scale > 0.05 {
				return refuse("ref-boxes-disagree", fmt.Sprintf("--ref boxes disagree: width factor %.4f, height factor %.4f", scale, byHeight))
			}
		default:
			return refuse("calibration-missing", "two images need a calibration — --scale <B px per A px>, or --ref-a and --ref-b boxes already known correct in both images")
		}
	} else if a.scale != nil || a.refA != nil || a.refB != nil || a.regionB != nil {
		return refuse("second-image-missing", "--scale/--ref-*/--region-b need a second image")
	}

	matrix := slices.Contains(props, "matrix")
	if matrix {
		if imageB == "" || a.regionA != nil || a.regionB != nil {
			return refuse("props-matrix-not-whole", "--props matrix reads two whole images: give both and drop any --region-*")
		}
		props = slices.DeleteFunc(props, func(p string) bool { return p == "matrix" })
	}
	result := mvpObj{}
	if len(props) > 0 {
		var code int
		if result, code = mvpPage(env, a, props, imageB, scale, refuse, stderr); code >= 0 {
			return code
		}
	}
	if matrix {
		out, err := mvpMatrix(env, a, scale)
		if err != nil {
			return refuse("matrix-failed", err.Error())
		}
		if len(result) > 0 {
			out = out[1:] // `scale`, already printed with the page props
		}
		result = append(result, out...)
	}
	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return refuse("measurement-not-finite", "a measurement is no finite number: "+err.Error())
	}
	fmt.Fprintln(stdout, string(out))
	return 0
}

// mvpObj is a JSON object whose keys print in insertion order. A value in an mvpObj tree is nil,
// int, float64, string, mvpObj or []mvpObj.
type mvpObj []mvpKV

type mvpKV struct {
	k string
	v any
}

func (o mvpObj) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')
	for i, kv := range o {
		if i > 0 {
			b.WriteByte(',')
		}
		k, _ := json.Marshal(kv.k)
		v, err := json.Marshal(kv.v)
		if err != nil {
			return nil, err
		}
		b.Write(k)
		b.WriteByte(':')
		b.Write(v)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

// roundTo is x rounded to n decimals, half to even on x's exact binary value (strconv's correctly
// rounded 'f' format, read back). -0.0 keeps its sign; NaN and ±Inf pass through.
func roundTo(x float64, n int) float64 {
	f, _ := strconv.ParseFloat(strconv.FormatFloat(x, 'f', n, 64), 64)
	return f
}

// mvpPage measures every prop but matrix on the two images, or the one; code
// is -1 when it measured.
func mvpPage(env core.Env, a mvpArgs, props []string, imageB string, scale float64, refuse func(id, msg string) int, stderr io.Writer) (mvpObj, int) {
	page := true
	for _, p := range props {
		page = page && slices.Contains(mvpPageProps, p)
	}
	result := mvpObj{}
	which := "a"
	measure := func(path string, box *[4]int) (mvpMeasured, error) {
		r, err := mvpOpen(env, path, box, a, page)
		if err != nil {
			return mvpMeasured{}, err
		}
		if page && !slices.Contains(props, "elements") { // elements finds its own surface below a scrim
			r.cropScrim()
		}
		return r.measure(props)
	}
	ma, err := measure(a.images[0], a.regionA)
	if err == nil {
		result = append(result, mvpKV{"a", ma.out})
		if imageB != "" {
			which = "b"
			var mb mvpMeasured
			if mb, err = measure(imageB, a.regionB); err == nil {
				result = append(result, mvpKV{"b", mb.out}, mvpKV{"scale", scale}, mvpKV{"delta", mvpDelta(ma, mb, props, scale)})
			}
		}
	}
	var unresolved mvpUnresolved
	if errors.As(err, &unresolved) {
		fmt.Fprintf(stderr, "%simage %s: %s\n", mvpPrefix, which, unresolved)
		core.Reason(env, "image-unresolved")
		return nil, 1
	}
	if err != nil {
		return nil, refuse("cannot-load-image", err.Error())
	}
	return result, -1
}

// mvpParse is the script's argparse surface under exact option names
// (design.md exact-option-names): `--opt value` and `--opt=value`, the last
// of a repeated option winning, positionals anywhere and everything after
// the first `--` positional, -h/--help printing the usage text. A usage error
// is one line carrying argparse's own message. Errors surface in argparse's
// order: an option's missing or bad value and -h act where they stand; a
// missing image and an unrecognized argument are reported after the scan.
// code is -1 when the arguments parsed.
func mvpParse(args []string, env core.Env, stdout, stderr io.Writer) (a mvpArgs, code int) {
	a = mvpArgs{props: strings.Join(mvpAllProps, ","), edge: 24, noise: 8, tolerance: mvpOverLength, decoded: map[string]*rgbImage{}}
	usage := func(msg string) (mvpArgs, int) {
		fmt.Fprintln(stderr, mvpPrefix+msg)
		core.Reason(env, "usage")
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
	if name == "--scale" || name == "--edge" || name == "--noise" || name == "--tolerance" { // type=float
		f, ok := mvpParseFloat(val)
		if !ok {
			return "invalid float value: " + ppRepr(val)
		}
		switch name {
		case "--scale":
			a.scale = &f
		case "--edge":
			a.edge = f
		case "--tolerance":
			if f < 0 {
				return "must not be negative"
			}
			a.tolerance = f
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
	s = core.PyStrip(s)
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
		p = core.PyStrip(p)
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
	bg               [3]byte
	box              mvpBox
	radius           [4]int // top_left, top_right, bottom_left, bottom_right
	fill             [3]byte
	border           [4]int // left, right, top, bottom
	borderColour     [3]byte
	haveBorderColour bool
	clipped          [4]bool // left, right, top, bottom: the control runs off the image there
	clamped          [4]bool // the given region ran past the image's edge there
	cropped          bool    // the region was given, not the whole image
	chrome           []bool  // image pixels of a device's rounded corners (page mode, whole image), read as background
}

// mvpCorners are the region's corners, each with the two sides it touches.
func (r *mvpRegion) corners() [4][4]int {
	return [4][4]int{{0, 0, 0, 2}, {r.w - 1, 0, 1, 2}, {0, r.h - 1, 0, 3}, {r.w - 1, r.h - 1, 1, 3}}
}

// clip marks each side of a given region lying on the image's own edge whose
// centre scan line starts on non-background: the control runs off the image
// there, so that side has no edge, shadow or gap to measure. The whole image
// is never a crop around one control, so it clips nothing.
func (r *mvpRegion) clip() {
	if !r.cropped {
		return
	}
	cx, cy := r.w/2, r.h/2
	onEdge := [4]bool{r.x0 == 0, r.x0+r.w == r.im.W, r.y0 == 0, r.y0+r.h == r.im.H}
	ends := [4][2]int{{0, cy}, {r.w - 1, cy}, {cx, 0}, {cx, r.h - 1}}
	for k := range r.clipped {
		r.clipped[k] = r.clamped[k] || onEdge[k] && !r.isBg(r.at(ends[k][0], ends[k][1]))
	}
}

// checkCorners requires every corner touching no clipped side to be background.
func (r *mvpRegion) checkCorners() error {
	for _, c := range r.corners() {
		if !r.clipped[c[2]] && !r.clipped[c[3]] && mvpDist(r.at(c[0], c[1]), r.bg) > r.noise {
			return mvpUnresolved(fmt.Sprintf("region corner (%d,%d) is not background %s — widen the crop", c[0], c[1], mvpHex(r.bg)))
		}
	}
	return nil
}

// mvpOpen is Region.__init__: the image read as convert("RGB") yields it, the
// region checked against it, and the background taken — the region's modal
// colour of its outermost rows and columns in page mode, else its top-left pixel, every other corner required
// to match it within the noise — save a corner on a clipped side, and with
// each other corner colour tried when the top-left pixel fails.
func mvpOpen(env core.Env, path string, box *[4]int, a mvpArgs, page bool) (*mvpRegion, error) {
	file := path
	if file != "" && !filepath.IsAbs(file) {
		file = env.Dir + "/" + file // unjoined: the OS resolves "..", as Python's open() did
	}
	im := a.decoded[file]
	var err error
	if im == nil {
		if im, err = decodeRGB(file); err != nil {
			return nil, fmt.Errorf("unreadable image %s: %s", path, pilImageError(err, file, path))
		}
		a.decoded[file] = im
	}
	b := [4]int{0, 0, im.W, im.H}
	if box != nil {
		b = *box
	}
	r := &mvpRegion{im: im, x0: b[0], y0: b[1], w: b[2], h: b[3], edge: a.edge, noise: a.noise, cropped: box != nil}
	if r.x0 >= im.W || r.y0 >= im.H {
		return nil, fmt.Errorf("region (%d, %d, %d, %d) exceeds %s (%dx%d)", b[0], b[1], b[2], b[3], path, im.W, im.H)
	}
	// sr11 (kan-1017): a region starting inside the image but running past
	// its edge is clamped to the image and clipped on that side.
	if r.x0+r.w > im.W {
		r.w, r.clamped[1] = im.W-r.x0, true
	}
	if r.y0+r.h > im.H {
		r.h, r.clamped[3] = im.H-r.y0, true
	}
	r.clipped = r.clamped
	if page {
		// The perimeter, not the whole region: a card larger than the page
		// around it is ink (kan-1083 self-review, F18).
		var c mvpCounter
		for y := 0; y < r.h; y++ {
			for x := 0; x < r.w; x++ {
				if y == 0 || y == r.h-1 || x == 0 || x == r.w-1 {
					c.add(r.at(x, y))
				}
			}
		}
		r.bg, _ = c.top()
		if box == nil {
			r.markChrome()
		}
		return r, nil
	}
	r.bg = r.at(0, 0)
	r.clip()
	err = r.checkCorners()
	if err == nil {
		return r, nil
	}
	// A control running off the image's edge leaves a corner on that side
	// non-background — two of the four when it spans the region, a 2:2 tie
	// no modal count breaks: retry with each other corner colour as the
	// background, keep the first that passes, and the top-left reading's
	// reason when none does.
	first := r.bg
	for _, k := range r.corners() {
		r.bg = r.at(k[0], k[1])
		if r.bg == first {
			continue
		}
		r.clip()
		if r.checkCorners() == nil {
			return r, nil
		}
	}
	return nil, err
}

// clippedText is the clipped sides, comma-joined.
func (r *mvpRegion) clippedText() string {
	var sides []string
	for k, c := range r.clipped {
		if c {
			sides = append(sides, mvpSides[k])
		}
	}
	return strings.Join(sides, ",")
}

func (r *mvpRegion) at(x, y int) [3]byte { return r.im.at(r.x0+x, r.y0+y) }

// cropScrim crops a page region dimmed by a modal's scrim to the sheet: when
// the top-left colour and the largest colour unlike it each cover 20% of the
// region and the corner is the darker, the corner is the scrim, and the region
// shrinks to the bounding box of the pixels unlike it, its background
// recounted there (kan-1044 self-review, F1-e). A scrim surrounds its sheet,
// so when the corner's colour covers 5% or more of that bounding box the
// corner is a dark page between cards, not a scrim, and nothing is cropped
// (kan-1083 self-review, F18: the dk-SL1 capture's black page was cropped as
// a scrim and its card read as the background).
// ponytail: a flat scrim reaching the corner only; a gradient scrim, or a
// sheet showing much of the scrim's colour inside it, needs a layer detector.
func (r *mvpRegion) cropScrim() {
	var c mvpCounter
	for y := 0; y < r.h; y++ {
		for x := 0; x < r.w; x++ {
			c.add(r.at(x, y))
		}
	}
	scrim := r.at(0, 0)
	total, ns, other, no := r.w*r.h, 0, [3]byte{}, 0
	for _, x := range c.order {
		if mvpDist(x, scrim) <= r.noise {
			ns += c.n[x]
		} else if c.n[x] > no {
			other, no = x, c.n[x]
		}
	}
	lum := func(x [3]byte) int { return int(x[0]) + int(x[1]) + int(x[2]) }
	if ns*5 < total || no*5 < total || lum(scrim) >= lum(other) {
		return
	}
	x0, y0, x1, y1 := r.w, r.h, -1, -1
	for y := 0; y < r.h; y++ {
		for x := 0; x < r.w; x++ {
			if mvpDist(r.at(x, y), scrim) > r.noise {
				x0, y0, x1, y1 = min(x0, x), min(y0, y), max(x1, x), max(y1, y)
			}
		}
	}
	inside := 0
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			if mvpDist(r.at(x, y), scrim) <= r.noise {
				inside++
			}
		}
	}
	if inside*20 >= (x1-x0+1)*(y1-y0+1) {
		return
	}
	r.x0, r.y0, r.w, r.h = r.x0+x0, r.y0+y0, x1-x0+1, y1-y0+1
	var in mvpCounter
	for y := 0; y < r.h; y++ {
		for x := 0; x < r.w; x++ {
			in.add(r.at(x, y))
		}
	}
	r.bg, _ = in.top()
}

// markChrome marks a device's rounded screen corners, which a frame draws and
// a capture does not: the ink connected to a non-background image corner pixel,
// a corner's arc with the fading strip along the edge beside it. Ink covering
// 5% or more of the image is a control flush with the corner, not chrome.
func (r *mvpRegion) markChrome() {
	W, H := r.im.W, r.im.H
	mark := make([]bool, W*H)
	for _, c := range [][2]int{{0, 0}, {W - 1, 0}, {0, H - 1}, {W - 1, H - 1}} {
		if r.isBg(r.im.at(c[0], c[1])) || mark[c[1]*W+c[0]] {
			continue
		}
		seen, queue := map[int]bool{c[1]*W + c[0]: true}, [][2]int{c}
		for i := 0; i < len(queue); i++ {
			for _, d := range [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
				x, y := queue[i][0]+d[0], queue[i][1]+d[1]
				if x < 0 || y < 0 || x >= W || y >= H || seen[y*W+x] || r.isBg(r.im.at(x, y)) {
					continue
				}
				seen[y*W+x] = true
				queue = append(queue, [2]int{x, y})
			}
			if len(queue)*20 >= W*H {
				break
			}
		}
		if len(queue)*20 < W*H {
			for k := range seen {
				mark[k] = true
			}
		}
	}
	r.chrome = mark
}

// inkAt: the pixel is ink — not background, not device-corner chrome.
func (r *mvpRegion) inkAt(x, y int) bool {
	return !r.isBg(r.at(x, y)) && (r.chrome == nil || !r.chrome[(r.y0+y)*r.im.W+r.x0+x])
}

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

// darkest is the lowest-luminance colour holding at least 1% of the counted
// pixels — a band's text ink rather than a wider border's grey (kan-1017),
// or the modal colour when none holds 1% — and share is the top two colours' share of the count, rounded to 3.
func (c *mvpCounter) darkest() (colour [3]byte, share float64) {
	total, n1, n2 := 0, 0, 0
	for _, x := range c.order {
		n := c.n[x]
		total += n
		if n > n1 {
			n1, n2 = n, n1
		} else if n > n2 {
			n2 = n
		}
	}
	lum := math.Inf(1)
	for _, x := range c.order {
		if l := 0.2126*float64(x[0]) + 0.7152*float64(x[1]) + 0.0722*float64(x[2]); c.n[x]*100 >= total && l < lum {
			colour, lum = x, l
		}
	}
	if math.IsInf(lum, 1) {
		colour, _ = c.top()
	}
	return colour, roundTo(float64(n1+n2)/float64(total), 3)
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

func (r *mvpRegion) measureBox() (mvpObj, error) {
	cx, cy := r.w/2, r.h/2
	lines := [][][2]int{r.row(cy, 0, cx), r.row(cy, r.w-1, cx), r.col(cx, 0, cy), r.col(cx, r.h-1, cy)}
	var found [4]int
	for k, pts := range lines {
		if r.clipped[k] {
			found[k] = pts[0][k/2] // the image's edge: the control runs past it
			continue
		}
		i := r.firstEdge(pts)
		if i < 0 {
			return nil, mvpUnresolved(fmt.Sprintf("no edge on the centre scan line from the %s (jump >= %s); the region's centre may lie outside the control, or lower --edge", mvpSides[k], strconv.FormatFloat(r.edge, 'g', -1, 64)))
		}
		found[k] = pts[i][k/2] // x for left/right, y for top/bottom
	}
	b := mvpBox{left: found[0], right: found[1], top: found[2], bottom: found[3]}
	b.width, b.height = b.right-b.left+1, b.bottom-b.top+1
	r.box = b
	return mvpObj{{"left", b.left}, {"right", b.right}, {"top", b.top}, {"bottom", b.bottom}, {"width", b.width}, {"height", b.height}}, nil
}

func (r *mvpRegion) measureRadius() (mvpObj, error) {
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
	out := mvpObj{}
	sides := r.corners()
	for k, c := range corners {
		if r.clipped[sides[k][2]] || r.clipped[sides[k][3]] {
			out = append(out, mvpKV{c.name, "clipped"})
			continue
		}
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
		r.radius[k] = int(roundTo(math.Sqrt(float64(area)/(1-math.Pi/4)), 0))
		out = append(out, mvpKV{c.name, r.radius[k]})
	}
	return out, nil
}

func (r *mvpRegion) interiorInset() (l, t, rt, btm int) {
	b := r.box
	ix, iy := b.width/4, b.height/4
	return b.left + ix, b.top + iy, b.right - ix, b.bottom - iy
}

func (r *mvpRegion) measureFill() mvpObj {
	l, t, rt, btm := r.interiorInset()
	var c mvpCounter
	for y := t; y <= btm; y++ {
		for x := l; x <= rt; x++ {
			c.add(r.at(x, y))
		}
	}
	colour, n := c.top()
	r.fill = colour
	return mvpObj{{"colour", mvpHex(colour)}, {"share", roundTo(float64(n)/float64((btm-t+1)*(rt-l+1)), 3)}}
}

func (r *mvpRegion) measureBorder() mvpObj {
	b, cx, cy := r.box, r.w/2, r.h/2
	lines := [][][2]int{r.row(cy, b.left, cx), r.row(cy, b.right, cx), r.col(cx, b.top, cy), r.col(cx, b.bottom, cy)}
	out := mvpObj{}
	for k, pts := range lines {
		if r.clipped[k] {
			out = append(out, mvpKV{mvpSides[k], "clipped"})
			continue
		}
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
		out = append(out, mvpKV{mvpSides[k], len(solid)})
		if len(solid) > 0 && !r.haveBorderColour {
			r.borderColour, r.haveBorderColour = solid[len(solid)/2], true
		}
	}
	if !r.haveBorderColour {
		return append(out, mvpKV{"colour", nil})
	}
	return append(out, mvpKV{"colour", mvpHex(r.borderColour)})
}

func (r *mvpRegion) measureShadow() mvpObj {
	black := mvpDist(r.bg, [3]byte{})
	if black == 0 {
		black = 1
	}
	out := mvpObj{}
	for k, side := range mvpSides {
		if r.clipped[k] {
			out = append(out, mvpKV{side, "clipped"})
			continue
		}
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
			peakDelta = roundTo(peak, 1)
		}
		out = append(out, mvpKV{side, mvpObj{{"width", width}, {"peak_delta", peakDelta}, {"approx_opacity", roundTo(math.Min(1.0, peak/black), 2)}}})
	}
	return out
}

func (r *mvpRegion) measureContent() mvpObj {
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
		return mvpObj{{"width", nil}, {"height", nil}, {"ratio", nil}, {"padding", nil}, {"colour", nil}}
	}
	w, h := cr-cl+1, cb-ct+1
	// The tint: the modal colour of the content pixels themselves. `fill`
	// is the box behind an icon, never the icon — a grey glyph on the
	// right fill reads as a match on every other property (KAN-437).
	colour, _ := tint.top()
	return mvpObj{
		{"width", w},
		{"height", h},
		{"colour", mvpHex(colour)},
		{"ratio", mvpObj{{"width", roundTo(float64(w)/float64(b.width), 3)}, {"height", roundTo(float64(h)/float64(b.height), 3)}}},
		{"padding", mvpObj{{"left", cl - b.left}, {"top", ct - b.top}, {"right", b.right - cr}, {"bottom", b.bottom - cb}}},
	}
}

func (r *mvpRegion) measureGap() mvpObj {
	out := mvpObj{}
	for k, side := range mvpSides {
		if r.clipped[k] {
			out = append(out, mvpKV{side, "clipped"})
			continue
		}
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
		out = append(out, mvpKV{side, gap})
	}
	return out
}

func (r *mvpRegion) measureInk() mvpObj {
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
		return mvpObj{{"width", nil}, {"height", nil}, {"colour", nil}}
	}
	// The tint of the ink itself — a text run has no box for `content` to
	// look inside, so this is the one reading of a label's colour (KAN-437
	// final verification: a caption and a summary row shipped in the wrong
	// colour on a frame whose every label had been transcribed).
	colour, _ := tint.top()
	return mvpObj{{"left", l}, {"top", t}, {"width", rt - l + 1}, {"height", btm - t + 1}, {"colour", mvpHex(colour)}}
}

func (r *mvpRegion) measureRuns() mvpObj {
	cx, cy := r.w/2, r.h/2
	out := mvpObj{}
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
		list := make([]mvpObj, len(runs))
		for i, u := range runs {
			list[i] = mvpObj{{"from", u.from}, {"to", u.to}, {"length", u.to - u.from + 1}, {"colour", mvpHex(u.colour)}}
		}
		out = append(out, mvpKV{line.name, list})
	}
	return out
}

type mvpBand struct {
	top, height, gapAbove int
	edge, colour          [3]byte
	share                 float64
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

func (b mvpBand) obj() mvpObj {
	return mvpObj{{"top", b.top}, {"height", b.height}, {"gap_above", b.gapAbove}, {"edge", mvpHex(b.edge)}, {"colour", mvpHex(b.colour)}, {"share", b.share}}
}

func (s mvpSeam) obj() mvpObj {
	return mvpObj{{"left", s.left}, {"width", s.width}, {"gap_left", s.gapLeft}, {"colour", mvpHex(s.colour)}}
}

// seamsObj is a boxed band as `seams` prints it: the band, its seams, its cells.
func (b mvpBand) seamsObj() mvpObj {
	seams := make([]mvpObj, len(b.seams))
	for i, s := range b.seams {
		seams[i] = s.obj()
	}
	cells := make([]mvpObj, len(b.cells))
	for i, c := range b.cells {
		lines := make([]mvpObj, len(c.lines))
		for j, ln := range c.lines {
			lines[j] = mvpObj{{"top", ln.top}, {"height", ln.height}, {"left", ln.left}, {"right", ln.right}, {"offset", ln.offset}}
		}
		cells[i] = mvpObj{{"left", c.left}, {"width", c.width}, {"lines", lines}}
	}
	return append(b.obj(), mvpKV{"seams", seams}, mvpKV{"cells", cells})
}

func (r *mvpRegion) measureBands() []mvpBand {
	var bands []mvpBand
	prevEnd, y := -1, 0
	for y < r.h {
		var edge mvpCounter
		for x := 0; x < r.w; x++ {
			if c := r.at(x, y); r.inkAt(x, y) {
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
				if c := r.at(x, y); r.inkAt(x, y) {
					ink.add(c)
					inked = true
				}
			}
			if !inked {
				break
			}
		}
		e, _ := edge.top()
		c, share := ink.darkest()
		bands = append(bands, mvpBand{top: top, height: y - top, gapAbove: top - prevEnd - 1, edge: e, colour: c, share: share})
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
					offset: roundTo(float64(ln.left+ln.right)/2-float64(x0+x1)/2, 1)}
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
	out          mvpObj
	bands, seams []mvpBand
	surfaceTop   int
	elements     []mvpElement
}

func (r *mvpRegion) measure(props []string) (mvpMeasured, error) {
	has := func(names ...string) bool {
		return slices.ContainsFunc(names, func(n string) bool { return slices.Contains(props, n) })
	}
	m := mvpMeasured{out: mvpObj{{"background", mvpHex(r.bg)}}}
	if c := r.clippedText(); c != "" {
		m.out = append(m.out, mvpKV{"clipped", c})
	}
	needsBox := slices.ContainsFunc(props, func(p string) bool { return !slices.Contains([]string{"ink", "runs", "bands", "seams", "elements"}, p) })
	if needsBox {
		box, err := r.measureBox()
		if err != nil {
			return m, err
		}
		if has("box") {
			m.out = append(m.out, mvpKV{"box", box})
		}
		if has("radius", "content") {
			radius, err := r.measureRadius()
			if err != nil {
				return m, err
			}
			if has("radius") {
				m.out = append(m.out, mvpKV{"radius", radius})
			}
		}
		if has("fill", "border", "content") {
			if fill := r.measureFill(); has("fill") {
				m.out = append(m.out, mvpKV{"fill", fill})
			}
		}
		if has("border", "content") {
			if border := r.measureBorder(); has("border") {
				m.out = append(m.out, mvpKV{"border", border})
			}
		}
		if has("shadow") {
			m.out = append(m.out, mvpKV{"shadow", r.measureShadow()})
		}
		if has("content") {
			m.out = append(m.out, mvpKV{"content", r.measureContent()})
		}
		if has("gap") {
			m.out = append(m.out, mvpKV{"gap", r.measureGap()})
		}
	}
	if has("ink") {
		m.out = append(m.out, mvpKV{"ink", r.measureInk()})
	}
	if has("runs") {
		m.out = append(m.out, mvpKV{"runs", r.measureRuns()})
	}
	if has("bands", "seams") {
		m.bands = r.measureBands()
		if has("bands") {
			list := make([]mvpObj, len(m.bands))
			for i, b := range m.bands {
				list[i] = b.obj()
			}
			m.out = append(m.out, mvpKV{"bands", list})
		}
		if has("seams") {
			m.seams = r.measureSeams(m.bands)
			list := make([]mvpObj, len(m.seams))
			for i, b := range m.seams {
				list[i] = b.seamsObj()
			}
			m.out = append(m.out, mvpKV{"seams", list})
		}
	}
	if has("elements") {
		var bg [3]byte
		list := []mvpObj{}
		m.surfaceTop, bg, m.elements = r.measureElements()
		for _, e := range m.elements {
			list = append(list, e.obj())
		}
		m.out = append(m.out, mvpKV{"elements", mvpObj{{"surface_top", m.surfaceTop}, {"background", mvpHex(bg)}, {"elements", list}}})
	}
	return m, nil
}

type mvpElement struct {
	top, height, left, width, right, gap int
	colour                               [3]byte
	kind                                 string
	segs                                 [][3]byte // a rail's marks' colours, left to right; nil on any other row
	group                                [2]int    // the widest ink group's width and height
}

// mvpSegMin is the narrowest run a rail mark is, and mvpGroupGap the widest
// blank a word space is: wider, it parts a badge or an icon from the text.
const (
	mvpSegMin   = 24
	mvpGroupGap = 24
)

// segments is the colours of the same-size marks on row y, left to right —
// at least three runs of pixels off the row's own background, its first
// pixel (over the noise, not the element's 30, so a light grey mark counts;
// the row's, not the surface's, so a rail above a card-filled page still
// parts at its white gaps), each mvpSegMin or wider and within 2px of the
// first's width — else nil.
func (r *mvpRegion) segments(y, x0, x1 int) [][3]byte {
	bg := r.at(0, y)
	var runs [][2]int
	for x := x0; x < x1; x++ {
		if mvpDist(r.at(x, y), bg) <= r.noise {
			continue
		}
		if n := len(runs); n > 0 && runs[n-1][1] == x {
			runs[n-1][1]++
		} else {
			runs = append(runs, [2]int{x, x + 1})
		}
	}
	runs = slices.DeleteFunc(runs, func(s [2]int) bool { return s[1]-s[0] < mvpSegMin })
	if len(runs) < 3 {
		return nil
	}
	w := runs[0][1] - runs[0][0]
	if slices.ContainsFunc(runs, func(s [2]int) bool { return s[1]-s[0] > w+2 || s[1]-s[0] < w-2 }) {
		return nil
	}
	out := make([][3]byte, len(runs))
	for i, s := range runs {
		var c mvpCounter
		for x := s[0]; x < s[1]; x++ {
			c.add(r.at(x, y))
		}
		out[i], _ = c.top()
	}
	return out
}

// group is the widest run of ink columns, blanks under mvpGroupGap merged,
// across rows (each a sorted ink-x list from top), and that run's ink height.
func mvpGroup(rows [][]int) [2]int {
	cols := map[int]bool{}
	for _, row := range rows {
		for _, x := range row {
			cols[x] = true
		}
	}
	xs := slices.Sorted(maps.Keys(cols))
	best, lo := [2]int{}, 0 // best is [left, right]
	for i := range xs {
		if i > 0 && xs[i]-xs[i-1] > mvpGroupGap {
			lo = i
		}
		if xs[i]-xs[lo] > best[1]-best[0] || i == 0 {
			best = [2]int{xs[lo], xs[i]}
		}
	}
	top, bot := -1, -1
	for k, row := range rows {
		if i, _ := slices.BinarySearch(row, best[0]); i < len(row) && row[i] <= best[1] {
			if top < 0 {
				top = k
			}
			bot = k
		}
	}
	if top < 0 {
		return [2]int{}
	}
	return [2]int{best[1] - best[0] + 1, bot - top + 1}
}

func (e mvpElement) obj() mvpObj {
	return mvpObj{{"top", e.top}, {"height", e.height}, {"left", e.left}, {"width", e.width}, {"right", e.right},
		{"colour", mvpHex(e.colour)}, {"kind", e.kind}, {"gap", e.gap}}
}

// measureElements is kan-1015's verifier matrix script (elems.py): the app
// surface's top — the first row, walking column 3 up from the bottom, whose
// mean channel is not above 200, plus one; 0 when that row is above row 60,
// the whole image then the surface, as it is when the bottom row itself is
// dark and the walk moves no row (a dark-theme page; kan-1041 self-review:
// every dark frame paired 0) — and every ink run below it against the
// surface's modal colour (distance over 30), 8px in from each side.
func (r *mvpRegion) measureElements() (int, [3]byte, []mvpElement) {
	y := r.h - 1
	for y > 0 {
		c := r.at(min(3, r.w-1), y)
		if int(c[0])+int(c[1])+int(c[2]) <= 600 {
			break
		}
		y--
	}
	t := y + 1
	if t < 60 || y == r.h-1 || t > r.h-60 {
		t = 0
	}
	x0, x1 := 8, r.w-8
	var bgc mvpCounter
	for yy := t; yy < r.h; yy += 4 {
		for x := x0; x < x1; x += 4 {
			bgc.add(r.at(x, yy))
		}
	}
	bg, _ := bgc.top()
	// A dark page is not a light one with the colours turned down: its cards
	// and sheets are the next shade up and can outnumber the page, so the two
	// sides of a pair pick different modal colours. The page is what the
	// surface's first row shows.
	if int(bg[0])+int(bg[1])+int(bg[2]) <= 600 {
		var top mvpCounter
		for x := x0; x < x1; x += 4 {
			top.add(r.at(x, t))
		}
		bg, _ = top.top()
	}
	rows := make([][]int, r.h-t)
	for i := range rows {
		for x := x0; x < x1; x++ {
			if mvpDist(r.at(x, t+i), bg) > 30 && (r.chrome == nil || !r.chrome[(r.y0+t+i)*r.im.W+r.x0+x]) {
				rows[i] = append(rows[i], x)
			}
		}
	}
	var out []mvpElement
	prev := t
	for i := 0; i < len(rows); {
		if len(rows[i]) == 0 {
			i++
			continue
		}
		s := i
		for i < len(rows) && len(rows[i]) > 0 {
			i++
		}
		left, right, span := rows[s][0], rows[s][0], 0
		for _, row := range rows[s:i] {
			left, right, span = min(left, row[0]), max(right, row[len(row)-1]), max(span, len(row))
		}
		mid := s + (i-s)/2
		var col mvpCounter
		for _, x := range rows[mid] {
			col.add(r.at(x, t+mid))
		}
		colour, _ := col.top()
		kind := "text/control"
		if inner := float64(x1 - x0); i-s <= 3 && float64(span) > 0.5*inner {
			kind = "rule"
		} else if float64(span) > 0.8*inner {
			kind = "fill band"
		}
		e := mvpElement{top: s + t, height: i - s, left: left, width: right - left + 1, right: r.w - 1 - right, colour: colour, kind: kind, gap: s + t - prev,
			segs: r.segments(t+mid, x0, x1), group: mvpGroup(rows[s:i])}
		prev = e.top + e.height
		out = append(out, e)
	}
	return t, bg, out
}

// mvpKindProps is the matrix columns each element kind fills: text its ink
// (and, for a control, its box columns when its region resolves), a fill band
// its box columns, a rule its runs.
var mvpKindProps = map[string][]string{
	"text/control": {"ink"},
	"fill band":    {"box", "radius", "border", "fill", "shadow", "content", "gap"},
	"rule":         {"runs"},
}

// mvpMatrix is `--props matrix`: both whole images' elements, paired in
// order, each pair measured on its own region — the element's box with a
// 4px margin, clipped to the image — with its kind's columns.
func mvpMatrix(env core.Env, a mvpArgs, scale float64) (mvpObj, error) {
	var sides [2]mvpMeasured
	for i, path := range a.images {
		r, err := mvpOpen(env, path, nil, a, true)
		if err != nil {
			return nil, err
		}
		_, _, sides[i].elements = r.measureElements()
	}
	ea, eb := sides[0].elements, sides[1].elements
	measure := func(path string, e mvpElement, props []string) (mvpMeasured, error) {
		const m = 4
		x, y := max(e.left-m, 0), max(e.top-m, 0)
		box := [4]int{x, y, e.left + e.width + m - x, e.top + e.height + m - y}
		r, err := mvpOpen(env, path, &box, a, false)
		if err != nil {
			return mvpMeasured{}, err
		}
		return r.measure(props)
	}
	ra, rb := mvpElementRecs(ea), mvpElementRecs(eb)
	pairs := mvpPairInOrder(ra, rb, scale, nil, "gap", "height", []string{"colour"})
	// A whole-frame shift — a scroll position, a bar the capture has and the
	// frame lacks — breaks the gap pairing above into missing and extra rows
	// (kan-1155 WU10: 24 missing, 19 extra, 3 paired). When the elements vote
	// for one offset and pairing by position under it pairs more of them — and
	// at least half of the shorter side, or it is another page, not this one
	// moved — that pairing stands and the summary names the shift once.
	var shift mvpObj
	if dy, ok := mvpShiftDy(ra, rb, scale); ok {
		alt := mvpPairInOrder(ra, rb, scale, &dy, "gap", "height", []string{"colour"})
		was, _, _ := mvpCountStatus(pairs)
		if n, _, _ := mvpCountStatus(alt); n > was && 2*n >= min(len(ra), len(rb)) {
			pairs, shift = mvpRepair(alt, ra, rb, scale), mvpObj{{"dy", dy}, {"rows", n}}
		}
	}
	rows := []mvpObj{}
	for _, p := range pairs {
		if p.status != "paired" {
			rows = append(rows, append(mvpObj{{"class", p.status}}, p.obj...))
			continue
		}
		p.obj = append(p.obj, mvpSegsAndScale(ea[p.i], eb[p.j], scale)...)
		kind := ea[p.i].kind
		props := mvpKindProps[kind]
		ma, errA := measure(a.images[0], ea[p.i], props)
		mb, errB := measure(a.images[1], eb[p.j], props)
		if errA != nil || errB != nil {
			row := append(mvpObj{{"kind", kind}, {"class", "unresolved"}}, p.obj...)
			rows = append(rows, append(row, mvpKV{"unresolved", errors.Join(errA, errB).Error()}))
			continue
		}
		class := kind
		if kind == "text/control" {
			class = "text run"
			box := mvpKindProps["fill band"]
			if ba, err := measure(a.images[0], ea[p.i], box); err == nil {
				if bb, err := measure(a.images[1], eb[p.j], box); err == nil {
					// Skip the box's own background and clipped keys: ma/mb carry them.
					rest := func(o mvpObj) mvpObj {
						if len(o) > 1 && o[1].k == "clipped" {
							return o[2:]
						}
						return o[1:]
					}
					ma.out, mb.out, props = append(ma.out, rest(ba.out)...), append(mb.out, rest(bb.out)...), append(slices.Clone(props), box...)
					class = "control"
				}
			}
		}
		row := append(mvpObj{{"kind", kind}, {"class", class}}, p.obj...)
		rows = append(rows, append(row, mvpKV{"measured", mvpObj{{"a", ma.out}, {"b", mb.out}, {"delta", mvpDelta(ma, mb, props, scale)}}}))
	}
	paired, missing, extra := mvpCountStatus(pairs)
	over, within := mvpOver(rows, a.tolerance)
	summary := mvpObj{{"paired", paired}, {"missing", missing}, {"extra", extra}, {"over", over}, {"within", within}}
	if shift != nil {
		summary = append(summary, mvpKV{"shift", shift})
	}
	return mvpObj{{"scale", scale}, {"matrix", rows}, {"summary", summary}}, nil
}

// mvpRepair pairs by gap each run of unpaired records the shift pairing left
// between two paired ones (or an image's edge): rows above a moved block sit
// at offsets of their own, smaller than the block's, and pair by their gaps
// from the last pair as an unshifted frame's do.
func mvpRepair(pairs []mvpPair, a, b []mvpRec, scale float64) []mvpPair {
	var out []mvpPair
	aEnd, bEnd := -1, -1 // the last pair's last row, per image
	for k := 0; k < len(pairs); {
		if pairs[k].status == "paired" {
			p := pairs[k]
			out = append(out, p)
			aEnd, bEnd = a[p.i].pos+a[p.i].size-1, b[p.j].pos+b[p.j].size-1
			k++
			continue
		}
		var is, js []int
		for ; k < len(pairs) && pairs[k].status != "paired"; k++ {
			if pairs[k].i >= 0 {
				is = append(is, pairs[k].i)
			} else {
				js = append(js, pairs[k].j)
			}
		}
		rebase := func(recs []mvpRec, idx []int, end int) []mvpRec {
			sub := make([]mvpRec, len(idx))
			for n, i := range idx {
				sub[n] = recs[i]
				sub[n].pos -= end + 1
			}
			return sub
		}
		for _, p := range mvpPairInOrder(rebase(a, is, aEnd), rebase(b, js, bEnd), scale, nil, "gap", "height", []string{"colour"}) {
			if p.i >= 0 {
				p.i = is[p.i]
			}
			if p.j >= 0 {
				p.j = js[p.j]
			}
			out = append(out, p)
		}
	}
	return out
}

// mvpSegsAndScale is a paired row's rail and scale keys. Per mark, when both
// rows are rails of as many marks, `segment<n>.colour` (1-based), so a mark
// whose grey moved past mvpOverColour is over though the row's own colour, its
// modal, is not. And `scale`, the widest ink groups' `<width>x<height>`, when
// that width moved by 5% or more and the height by within 5 points of the same
// ratio: the same glyphs drawn larger or smaller, which a badge or an icon
// beside them can hide from the row's own width. A text set in another face
// changes its width far more than its height, and carries none; nor does a
// rail, whose marks are no glyphs.
func mvpSegsAndScale(a, b mvpElement, scale float64) mvpObj {
	var out mvpObj
	if len(a.segs) == len(b.segs) {
		for k := range a.segs {
			out = append(out, mvpKV{fmt.Sprintf("segment%d.colour", k+1), mvpColourDelta(mvpHex(a.segs[k]), mvpHex(b.segs[k]))})
		}
	}
	if a.segs == nil && b.segs == nil && a.group[0] > 0 && b.group[0] > 0 {
		rw := float64(b.group[0]) / (float64(a.group[0]) * scale)
		rh := float64(b.group[1]) / (float64(a.group[1]) * scale)
		if math.Abs(rw-1) >= 0.05 && math.Abs(rw-rh) < 0.05 {
			size := func(g [2]int) string { return fmt.Sprintf("%dx%d", g[0], g[1]) }
			out = append(out, mvpKV{"scale", mvpObj{{"a", size(a.group)}, {"b", size(b.group)}, {"ratio", roundTo(rw, 3)}}})
		}
	}
	return out
}

// mvpShiftMin is the offset below which the gap pairing's own 8px tolerance
// already holds, and mvpShiftVotes how many elements must agree on one.
const (
	mvpShiftMin   = 8
	mvpShiftVotes = 3
)

// mvpShiftDy is the vertical offset most of a's elements sit from b's: every
// (a, b) couple whose sizes match within the pairing tolerance votes its
// b.pos − a.pos×scale, and the widest run of votes within 4px of each other
// wins, its median the offset. ok only when at least mvpShiftVotes distinct a
// elements are in the run and the offset is over mvpShiftMin either way.
func mvpShiftDy(a, b []mvpRec, scale float64) (float64, bool) {
	type vote struct {
		dy float64
		i  int
	}
	var votes []vote
	for i, ra := range a {
		for _, rb := range b {
			sa := float64(ra.size) * scale
			if math.Abs(float64(rb.size)-sa) <= 8+math.Max(sa, float64(rb.size))/4 {
				votes = append(votes, vote{float64(rb.pos) - float64(ra.pos)*scale, i})
			}
		}
	}
	slices.SortFunc(votes, func(x, y vote) int { return cmp.Compare(x.dy, y.dy) })
	distinct := func(run []vote) int {
		seen := map[int]bool{}
		for _, v := range run {
			seen[v.i] = true
		}
		return len(seen)
	}
	var best []vote
	n := 0
	for lo, hi := 0, 0; hi < len(votes); hi++ {
		for votes[hi].dy-votes[lo].dy > 4 {
			lo++
		}
		if d := distinct(votes[lo : hi+1]); d > n {
			best, n = votes[lo:hi+1], d
		}
	}
	if n < mvpShiftVotes {
		return 0, false
	}
	dy := best[len(best)/2].dy
	return dy, math.Abs(dy) > mvpShiftMin
}

// Step 10's band and seam tolerances (skills/flow/visual-verify-verifier.md
// cites them): a scaled length delta over mvpOverLength px, or a colour
// distance over mvpOverColour, puts a paired matrix row on the `over` list.
//
// mvpOverColour is 4, measured on kan-1155's 87 frame/capture pairs (2,884
// paired colour deltas): rendering noise — one unit a channel, #6c6c70 against
// #6b6b70, and an antialiased hairline's blend — reached 3.5 and no further,
// while two different greys (#e9e9ec against #e6e6e9) started at 5.2. The
// earlier 12 let those through; 0 would flag the noise.
const (
	mvpOverLength = 2
	mvpOverColour = 4
)

// mvpOver is summary.over (decision over-list): one {row, status, keys}
// entry per matrix row, row its 1-based index, that is missing or extra,
// carries `unresolved`, or is paired with a scaled numeric leaf of its
// pairing or its measured.delta over tol — `--tolerance`, mvpOverLength by
// default — or a colour distance over mvpOverColour; keys names the leaves
// over, comma-separated. The length tolerance never reaches a colour, a
// missing or an extra row. within is summary.within: one {row, values} entry
// per paired row off the over list that carries a non-zero length delta, every
// one of them inside tol, values its `<leaf>=<abs>` deltas comma-separated.
func mvpOver(rows []mvpObj, tol float64) (over, within []mvpObj) {
	over, within = []mvpObj{}, []mvpObj{}
	for i, row := range rows {
		status, keys, inside := "", []string{}, []string{}
		judge := func(key string, v any) {
			d, ok := v.(mvpObj)
			if !ok {
				return
			}
			for _, kv := range d {
				switch {
				case kv.k == "distance" && kv.v != nil && mvpF(kv.v) > mvpOverColour:
					keys = append(keys, key)
				case kv.k == "abs" && kv.v != nil && mvpF(kv.v) != 0 &&
					!slices.ContainsFunc(strings.Split(key, "."), func(seg string) bool { return slices.Contains(mvpUnscaled, seg) }):
					if math.Abs(mvpF(kv.v)) > tol {
						keys = append(keys, key)
					} else {
						inside = append(inside, key+"="+strconv.FormatFloat(mvpF(kv.v), 'f', -1, 64))
					}
				}
			}
		}
		for _, kv := range row {
			switch kv.k {
			case "status":
				status, _ = kv.v.(string)
			case "kind", "class", "a", "b":
			case "unresolved", "scale":
				keys = append(keys, kv.k)
			case "measured":
				for _, m := range kv.v.(mvpObj) {
					if m.k == "delta" {
						for _, dk := range m.v.(mvpObj) {
							judge(dk.k, dk.v)
						}
					}
				}
			default:
				judge(kv.k, kv.v)
			}
		}
		if status == "missing" || status == "extra" || len(keys) > 0 {
			over = append(over, mvpObj{{"row", i + 1}, {"status", status}, {"keys", strings.Join(keys, ", ")}})
		} else if len(inside) > 0 {
			within = append(within, mvpObj{{"row", i + 1}, {"values", strings.Join(inside, ", ")}})
		}
	}
	return over, within
}

func mvpElementRecs(elems []mvpElement) []mvpRec {
	recs := make([]mvpRec, len(elems))
	for i, e := range elems {
		recs[i] = mvpRec{e.top, e.height, e.gap, []string{mvpHex(e.colour)}, e.obj()}
	}
	return recs
}

// mvpLeaves is every non-object value under d, keyed by its dotted path.
func mvpLeaves(d mvpObj, prefix string) []mvpKV {
	var out []mvpKV
	for _, kv := range d {
		if sub, ok := kv.v.(mvpObj); ok {
			out = append(out, mvpLeaves(sub, prefix+kv.k+".")...)
		} else {
			out = append(out, mvpKV{prefix + kv.k, kv.v})
		}
	}
	return out
}

func mvpColourDelta(av, bv string) mvpObj {
	parse := func(h string) (c [3]byte) {
		for i := range c {
			n, _ := strconv.ParseUint(h[1+2*i:3+2*i], 16, 8)
			c[i] = byte(n)
		}
		return c
	}
	return mvpObj{{"a", av}, {"b", bv}, {"distance", roundTo(mvpDist(parse(av), parse(bv)), 1)}}
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
		return roundTo(f, n)
	}
	return v
}

// mvpNumberDelta is number_delta; scale is the int 1 for an unscaled leaf,
// else the float scale.
func mvpNumberDelta(av, bv, scale any) mvpObj {
	scaled := mvpMul(av, scale)
	ab := mvpSub(bv, scaled)
	var pct any
	if mvpF(scaled) != 0 {
		pct = roundTo(mvpF(ab)/mvpF(scaled)*100, 1)
	}
	return mvpObj{{"a", av}, {"a_scaled", mvpRoundN(scaled, 2)}, {"b", bv}, {"abs", mvpRoundN(ab, 2)}, {"pct", pct}}
}

// mvpRec is one band or seam as pairInOrder reads it.
type mvpRec struct {
	pos, size, gap int
	colours        []string
	obj            mvpObj
}

// mvpPair is one pairInOrder entry: i and j index the records it holds, -1
// for the side it lacks.
type mvpPair struct {
	status string
	i, j   int
	obj    mvpObj
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
//
// A non-nil shift pairs by position instead: a record pairs when it sits
// *shift from its counterpart, within the same tolerance, and its size
// matches — mvpMatrix's second pairing for a frame moved whole, where one
// record the gap pairing cannot place unpairs every record after it. A record
// within mvpOverLength of its counterpart's own position pairs too: a bar
// pinned to the frame's edge, which a scrolled page leaves where it was.
func mvpPairInOrder(a, b []mvpRec, scale float64, shift *float64, gapKey, sizeKey string, colourKeys []string) []mvpPair {
	var out []mvpPair
	missing := func(i int) mvpPair {
		return mvpPair{"missing", i, -1, mvpObj{{"status", "missing"}, {"a", a[i].obj}, {"b", nil}}}
	}
	extra := func(j int) mvpPair {
		return mvpPair{"extra", -1, j, mvpObj{{"status", "extra"}, {"a", nil}, {"b", b[j].obj}}}
	}
	i, j := 0, 0
	aAnchor, bAnchor := -1, -1 // last row/column of the last paired record, per image
	for i < len(a) || j < len(b) {
		switch {
		case i < len(a) && j < len(b):
			ga := float64(a[i].pos-aAnchor-1) * scale
			gb := b[j].pos - bAnchor - 1
			tol := 8 + math.Max(float64(a[i].size)*scale, float64(b[j].size))/4
			off, first := float64(gb)-ga, ga < float64(gb)
			if shift != nil {
				pa := float64(a[i].pos)*scale + *shift
				off, first = float64(b[j].pos)-pa, pa < float64(b[j].pos)
				// A bar pinned to the frame's edge did not move with the rest.
				if pinned := float64(b[j].pos) - float64(a[i].pos)*scale; math.Abs(pinned) <= mvpOverLength {
					off = pinned
				}
			}
			if math.Abs(off) <= tol && math.Abs(float64(b[j].size)-float64(a[i].size)*scale) <= tol {
				obj := mvpObj{{"status", "paired"}, {"a", a[i].obj}, {"b", b[j].obj}, {gapKey, mvpNumberDelta(a[i].gap, b[j].gap, scale)},
					// Distance from the last PAIRED record, so a padding lost
					// around a missing divider still reads as one number here.
					{"since_pair", mvpNumberDelta(a[i].pos-aAnchor-1, gb, scale)},
					{sizeKey, mvpNumberDelta(a[i].size, b[j].size, scale)}}
				for k, key := range colourKeys {
					obj = append(obj, mvpKV{key, mvpColourDelta(a[i].colours[k], b[j].colours[k])})
				}
				out = append(out, mvpPair{"paired", i, j, obj})
				aAnchor, bAnchor = a[i].pos+a[i].size-1, b[j].pos+b[j].size-1
				i, j = i+1, j+1
			} else if first {
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

func mvpPairObjs(pairs []mvpPair) []mvpObj {
	objs := make([]mvpObj, len(pairs))
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
func mvpPairSeams(a, b []mvpBand, scale float64) ([]mvpObj, mvpObj) {
	var out []mvpObj
	var paired, missing, extra, bandsUnpaired, linesPaired, linesUnpaired int
	for _, band := range mvpPairInOrder(mvpBandRecs(a), mvpBandRecs(b), scale, nil, "gap_above", "height", []string{"edge", "colour"}) {
		var ea, eb any
		if band.i >= 0 {
			ea = mvpObj{{"top", a[band.i].top}, {"height", a[band.i].height}}
		}
		if band.j >= 0 {
			eb = mvpObj{{"top", b[band.j].top}, {"height", b[band.j].height}}
		}
		entry := mvpObj{{"status", band.status}, {"a", ea}, {"b", eb}}
		if band.status != "paired" {
			bandsUnpaired++
			out = append(out, entry)
			continue
		}
		ba, bb := a[band.i], b[band.j]
		seams := mvpPairInOrder(mvpSeamRecs(ba.seams), mvpSeamRecs(bb.seams), scale, nil, "gap_left", "width", []string{"colour"})
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
		cells := []mvpObj{}
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
			lines := []mvpObj{}
			for n := 0; n < min(len(ca.lines), len(cb.lines)); n++ {
				la, lb := ca.lines[n], cb.lines[n]
				lines = append(lines, mvpObj{{"offset", mvpNumberDelta(la.offset, lb.offset, scale)},
					{"left", mvpNumberDelta(la.left, lb.left, scale)}, {"right", mvpNumberDelta(la.right, lb.right, scale)}})
			}
			linesPaired += len(lines)
			linesUnpaired += max(len(ca.lines)-len(cb.lines), len(cb.lines)-len(ca.lines))
			cells = append(cells, mvpObj{{"a", ia}, {"b", ib}, {"count", mvpObj{{"a", len(ca.lines)}, {"b", len(cb.lines)}}}, {"lines", lines}})
		}
		out = append(out, append(entry, mvpKV{"seams", mvpPairObjs(seams)}, mvpKV{"cells", cells}))
	}
	return out, mvpObj{{"paired", paired}, {"missing", missing}, {"extra", extra}, {"bands_unpaired", bandsUnpaired},
		{"lines", mvpObj{{"paired", linesPaired}, {"unpaired", linesUnpaired}}}}
}

func mvpDelta(a, b mvpMeasured, props []string, scale float64) mvpObj {
	out := mvpObj{}
	if slices.Contains(props, "bands") {
		pairs := mvpPairInOrder(mvpBandRecs(a.bands), mvpBandRecs(b.bands), scale, nil, "gap_above", "height", []string{"edge", "colour"})
		p, m, e := mvpCountStatus(pairs)
		out = append(out, mvpKV{"bands", mvpPairObjs(pairs)}, mvpKV{"bands_summary", mvpObj{{"paired", p}, {"missing", m}, {"extra", e}}})
	}
	if slices.Contains(props, "seams") {
		seams, summary := mvpPairSeams(a.seams, b.seams, scale)
		out = append(out, mvpKV{"seams", seams}, mvpKV{"seams_summary", summary})
	}
	if slices.Contains(props, "elements") {
		pairs := mvpPairInOrder(mvpElementRecs(a.elements), mvpElementRecs(b.elements), scale, nil, "gap", "height", []string{"colour"})
		for k, p := range pairs {
			if p.status == "paired" {
				pairs[k].obj = append(p.obj, mvpKV{"top", mvpNumberDelta(a.elements[p.i].top, b.elements[p.j].top, scale)})
			}
		}
		p, m, e := mvpCountStatus(pairs)
		out = append(out, mvpKV{"elements", mvpPairObjs(pairs)}, mvpKV{"elements_summary", mvpObj{{"paired", p}, {"missing", m}, {"extra", e}}})
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
		if key == "bands" || key == "seams" || strings.HasPrefix(key, "elements.") {
			continue
		}
		as, aok := av.(string)
		bs, bok := bv.(string)
		switch {
		case aok && bok && strings.HasPrefix(as, "#") && strings.HasPrefix(bs, "#"):
			out = append(out, mvpKV{key, mvpColourDelta(as, bs)})
		case isNum(av) && isNum(bv):
			var s any = scale
			if slices.ContainsFunc(strings.Split(key, "."), func(seg string) bool { return slices.Contains(mvpUnscaled, seg) }) {
				s = 1
			}
			out = append(out, mvpKV{key, mvpNumberDelta(av, bv, s)})
		case av == nil || bv == nil || av == "clipped" || bv == "clipped":
			out = append(out, mvpKV{key, mvpObj{{"a", av}, {"b", bv}, {"abs", nil}, {"pct", nil}}})
		}
	}
	return out
}
