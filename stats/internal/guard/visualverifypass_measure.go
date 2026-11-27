package guard

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// vvmStep is one in-process guard run: its exit and stderr.
type vvmStep struct {
	Exit   *int   `json:"exit"`
	Stderr string `json:"stderr,omitempty"`
}

type vvmExits struct {
	Resolve   *int `json:"resolve"`
	Identical *int `json:"identical"`
	Compose   *int `json:"compose"`
	Bands     *int `json:"bands"`
	Matrix    *int `json:"matrix"`
}

type vvmPair struct {
	Spec       string            `json:"spec"`
	Frame      string            `json:"frame"`
	Prototype  bool              `json:"prototype,omitempty"` // measured against the approved prototype, not a mockup
	Capture    string            `json:"capture"`
	Mockup     string            `json:"mockup"`
	Composite  string            `json:"composite"`
	Diff       string            `json:"diff"`
	Exits      vvmExits          `json:"exits"`
	Stderr     map[string]string `json:"stderr,omitempty"`
	BandsFile  string            `json:"bands_file,omitempty"`
	MatrixFile string            `json:"matrix_file,omitempty"`
	Bands      []any             `json:"bands"`
	Seams      []any             `json:"seams"`
	Over       []any             `json:"over"`
	Within     []any             `json:"within,omitempty"` // rows moved by at most the declared tolerance: recorded, never flagged
	Shift      *vvmShift         `json:"shift,omitempty"`
}

// vvmShift is one whole-frame offset, found one of two ways. `--props matrix`
// paired the frame by position under Dy (its `summary.shift`): Rows is the rows
// it paired and Row the first of them. Or vvmFoldShift folded Rows over rows
// whose only over keys are vertical positions and whose elements all sit Dy px
// (within 1px of each other) from the frame's: Row is the one of them left on
// the pair's `over` list to stand for the defect. Uniform marks a fold whose
// dy every paired row of the frame shares, however few rows that is.
type vvmShift struct {
	Dy      float64 `json:"dy"`
	Rows    int     `json:"rows"`
	Row     int     `json:"row"`
	Uniform bool    `json:"uniform,omitempty"`
}

// vvmShiftRows is how many rows must share one dy before they are one defect.
const vvmShiftRows = 20

type vvmManifestOut struct {
	Specs     []string           `json:"specs"` // every spec a measure ran over, pairs or none
	Captures  []string           `json:"captures"`
	Identical []string           `json:"identical"`
	Steps     map[string]vvmStep `json:"steps"`
	Pairs     []vvmPair          `json:"pairs"`
}

// measure is `visual-verify-pass.sh measure <worktree> <name> <key> <spec>
// [--prototype]` (decision measure-one-call): the verifier's measuring chain
// in process, carrying on past every non-zero exit, each one recorded in the
// manifest. Exit 0 when every step exited 0, 1 otherwise, 2 on a usage error
// or an output directory it cannot create.
func (v *vvp) measure(args []string) {
	prototype := false
	var pos []string
	for _, a := range args {
		switch {
		case a == "--prototype":
			prototype = true
		case strings.HasPrefix(a, "--"):
			v.usage("unknown argument '" + a + "'")
		default:
			pos = append(pos, a)
		}
	}
	if len(pos) != 4 {
		v.usage("measure takes <worktree> <name> <key> <spec path>")
	}
	w, name, key, spec := v.abs(pos[0]), pos[1], pos[2], pos[3]
	if !plainChangeName(name) || key == "" || strings.ContainsRune(key, '/') {
		v.usage("change name '" + name + "' or key '" + key + "' is not plain")
	}
	spec = cellPath(w, spec)
	tol, ok := vvTolerance(v.cells(w)["tolerance"])
	if !ok {
		v.usage("`tolerance` row '" + v.cells(w)["tolerance"] + "' is not `<n>px at <scale>x`")
	}
	tolArg := strconv.FormatFloat(tol, 'f', -1, 64)
	// A spec that is not there measures nothing, and the merge would keep the
	// last run's pairs under a manifest that reads as this run's.
	if !vvmExists(spec) {
		v.usage("spec '" + spec + "' does not exist")
	}
	changeRoot := filepath.Join(w, specRootLeaf(w, io.Discard), "changes", name)
	vvDir := filepath.Join(changeRoot, "visual-verification")
	sdd := filepath.Join(w, ".superpowers", "sdd")
	bandsDir := filepath.Join(sdd, "visual-measure-"+key)
	matrixDir := filepath.Join(vvDir, "matrix")
	for _, d := range []string{bandsDir, matrixDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			v.die("cannot-create-output-directory", "cannot create output directory %s: %v", d, err)
		}
	}

	env := v.env
	env.Dir = w
	run := func(fn func([]string, Env, io.Writer, io.Writer) int, stdin string, a ...string) (int, string, string) {
		var out, errb bytes.Buffer
		e := env
		e.Stdin = strings.NewReader(stdin)
		rc := fn(a, e, &out, &errb)
		return rc, out.String(), errb.String()
	}
	ip := func(n int) *int { return &n }
	failed := false
	note := func(rc int) *int {
		if rc != 0 {
			failed = true
		}
		return ip(rc)
	}
	m := vvmManifestOut{Specs: []string{spec}, Captures: []string{}, Identical: []string{}, Steps: map[string]vvmStep{}, Pairs: []vvmPair{}}

	rc, out, serr := run(resolveVisualScreenshots, "", w, filepath.Base(spec))
	resolveExit := note(rc)
	m.Steps["resolve"] = vvmStep{resolveExit, serr}
	m.Captures = vvmLines(out)
	resolved := out

	rc, out, serr = run(checkIdenticalBaselines, resolved)
	identicalExit := note(rc)
	m.Steps["identical"] = vvmStep{identicalExit, serr}
	m.Identical = vvmLines(out)

	sidecar, mockups, frameArg := spec+".mockups", "", v.cells(w)["mockup frame"]
	if prototype {
		sidecar, mockups, frameArg = filepath.Join(vvDir, "prototype.mockups"), filepath.Join(vvDir, "prototype"), ""
	} else if c := v.cells(w)["mockups"]; c != "" {
		mockups = cellPath(w, c)
	}
	if _, err := os.Stat(sidecar); err != nil && !prototype || mockups == "" {
		v.writeManifest(sdd, vvDir, key, spec, prototype, m, failed)
	}

	// Each map line is one pair; a stale cropped frame from an earlier run
	// must not stand in for one this compose did not write. The prototype map
	// is shared by every spec: a line whose screenshot this spec does not
	// resolve belongs to another spec's measure and is left alone (kan-1057
	// self-review, F2). A map that cannot be read is composed over as it
	// stands, so compose's own refusal surfaces.
	body, readErr := os.ReadFile(sidecar)
	type line struct{ name, frame string }
	var lines []line
	// A prototype frame carries the 52px status line the capture does not: crop it, keep the rest at 1x.
	own := "frame scale=1 status=52 border=0 bare\n"
	for _, f := range mockupsLines(body) {
		if prototype && len(cmfMatchCapture(f[0], m.Captures)) == 0 {
			continue
		}
		lines = append(lines, line{f[0], f[1]})
		own += f[0] + " " + f[1] + "\n"
		_ = os.Remove(filepath.Join(vvDir, strings.TrimSuffix(f[0], ".png")+".frame.png"))
	}
	if prototype && readErr == nil {
		sidecar = filepath.Join(bandsDir, "prototype.mockups")
		v.writeOut(sidecar, own)
	}
	composeArgs := []string{sidecar, mockups, vvDir}
	if frameArg != "" {
		composeArgs = append(composeArgs, strings.Trim(frameArg, "`"))
	}
	composeRC, out, serr := run(composeMockupFrames, resolved, composeArgs...)
	m.Steps["compose"] = vvmStep{note(composeRC), serr}
	composed := map[string]string{} // composite path → diff ratio
	for _, l := range vvmLines(out) {
		if i := strings.LastIndex(l, " diff="); i > 0 {
			composed[l[:i]] = l[i+len(" diff="):]
		}
	}

	for _, l := range lines {
		stem := strings.TrimSuffix(l.name, ".png")
		p := vvmPair{Spec: spec, Frame: l.frame, Prototype: prototype, Mockup: filepath.Join(mockups, l.frame+".png"),
			Stderr: map[string]string{}, Bands: []any{}, Seams: []any{}, Over: []any{}}
		p.Exits.Resolve, p.Exits.Identical = resolveExit, identicalExit
		if c := cmfMatchCapture(l.name, m.Captures); len(c) == 1 {
			p.Capture = c[0]
		}
		composite := filepath.Join(vvDir, stem+".png")
		ratio, ok := composed[composite]
		if !ok || p.Capture == "" {
			// An earlier round's measurements would read as this round's.
			os.Remove(filepath.Join(bandsDir, l.frame+".json"))
			os.Remove(filepath.Join(matrixDir, l.frame+".json"))
			p.Exits.Compose = ip(max(composeRC, 1))
			failed = true
			m.Pairs = append(m.Pairs, p)
			continue
		}
		p.Composite, p.Diff, p.Exits.Compose = composite, ratio, ip(0)
		if frame := filepath.Join(vvDir, stem+".frame.png"); vvmExists(frame) {
			p.Mockup = frame
		}

		rc, out, serr := run(measureVisualProperties, "", p.Mockup, p.Capture, "--scale", "1", "--props", "bands,seams", "--tolerance", tolArg)
		p.Exits.Bands, p.Stderr["bands"] = note(rc), serr
		p.BandsFile = filepath.Join(bandsDir, l.frame+".json")
		v.writeOut(p.BandsFile, out)
		var bs map[string]any
		if json.Unmarshal([]byte(out), &bs) == nil {
			// measure-visual-properties nests the page props under `delta`.
			d, _ := bs["delta"].(map[string]any)
			p.Bands, p.Seams = vvmFlagged(d["bands"], tol), vvmFlagged(d["seams"], tol)
		}

		rc, out, serr = run(measureVisualProperties, "", p.Mockup, p.Capture, "--scale", "1", "--props", "matrix", "--tolerance", tolArg)
		p.Exits.Matrix, p.Stderr["matrix"] = note(rc), serr
		p.MatrixFile = filepath.Join(matrixDir, l.frame+".json")
		v.writeOut(p.MatrixFile, out)
		var mx struct {
			Matrix  []map[string]any `json:"matrix"`
			Summary struct {
				Over   []map[string]any
				Within []any
				Shift  *vvmShift
			} `json:"summary"`
		}
		if json.Unmarshal([]byte(out), &mx) == nil && mx.Summary.Over != nil {
			for _, r := range mx.Summary.Over {
				r["figures"] = vvmFigures(mx.Matrix, r)
				p.Over = append(p.Over, r)
			}
			p.Within = mx.Summary.Within
			// The matrix's own shift pairing names a frame moved whole; its first
			// paired row carries the offset on `since_pair` and stands for it.
			if p.Shift = mx.Summary.Shift; p.Shift != nil {
				p.Shift.Row = 1 + slices.IndexFunc(mx.Matrix, func(r map[string]any) bool { return r["status"] == "paired" })
			} else {
				p.Shift = vvmFoldShift(mx.Matrix, &p.Over, tol)
			}
			if p.Shift != nil {
				v.say("shift: dy=%s rows=%d frame=%s", strconv.FormatFloat(p.Shift.Dy, 'f', -1, 64), p.Shift.Rows, p.Frame)
			}
		}
		v.say("%s", vvmFrameLine(p))
		m.Pairs = append(m.Pairs, p)
	}
	v.writeManifest(sdd, vvDir, key, spec, prototype, m, failed)
}

func (v *vvp) writeOut(path, body string) {
	// A frame id carries a directory (`kan-1083/SE1`), so its output file sits below the key's directory.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		v.die("cannot-write", "cannot write %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		v.die("cannot-write", "cannot write %s: %v", path, err)
	}
}

// writeManifest merges m into the key's manifest, prints its path and exits 1
// when any step failed, else 0. A pair is keyed by (spec, frame, prototype): an
// existing manifest that parses keeps every pair m does not replace — m replaces
// the pairs of its own spec, and of a frame it measured, that were measured the
// same way (a plain run keeps a spec's prototype pairs and a --prototype run its
// plain ones). Captures and identical lines are unioned, existing first, and so
// are steps, m's winning where both carry one (kan-1102 self-review, F2).
func (v *vvp) writeManifest(sdd, vvDir, key, spec string, prototype bool, m vvmManifestOut, failed bool) {
	path := filepath.Join(sdd, "visual-manifest-"+key+".json")
	// Only this call's pairs count, so the line is read before the merge.
	if base, _, defects, ok := readRound(sdd, key); ok {
		v.loadBases(sdd, key)
		var carried []string
		if len(defects) == 0 {
			carried = vvpCarried(readJudged(sdd), m.Pairs)
		}
		var changed []string
		for _, p := range m.Pairs {
			if !slices.Contains(changed, p.Frame) && !slices.Contains(carried, p.Frame) && v.captureChanged(p.Capture, base) {
				changed = append(changed, p.Frame)
			}
		}
		if len(carried) > 0 {
			v.say("carried: %d — %s: capture identical to the one a report judged frame by frame", len(carried), strings.Join(carried, " "))
		}
		line := "none"
		if len(changed) > 0 {
			line = strings.Join(changed, " ")
		}
		v.say("changed: %s — round base %s", line, base)
	}
	var old vvmManifestOut
	if b, err := os.ReadFile(path); err == nil && json.Unmarshal(b, &old) == nil {
		mine := map[string]bool{}
		for _, p := range m.Pairs {
			mine[p.Frame] = true
		}
		var pairs []vvmPair
		for _, p := range old.Pairs {
			if p.Prototype != prototype || p.Spec != spec && !mine[p.Frame] {
				pairs = append(pairs, p)
			}
		}
		m.Pairs = append(pairs, m.Pairs...)
		m.Specs = vvmUnion(old.Specs, m.Specs)
		m.Captures = vvmUnion(old.Captures, m.Captures)
		m.Identical = vvmUnion(old.Identical, m.Identical)
		for k, s := range old.Steps {
			if _, ok := m.Steps[k]; !ok {
				m.Steps[k] = s
			}
		}
	}
	b, _ := json.MarshalIndent(m, "", "  ")
	v.writeOut(path, string(b)+"\n")
	// The appendix visual-verification.md cites: every row the declared
	// tolerance leaves off the over list, with its measured value.
	within, appendix := 0, "# Within tolerance\n\nRows `visual-verify-pass measure` left off the `over` list because every position or size delta is inside the declared `tolerance` (`## visual verification`), each with its measured deltas in capture px.\n\n"
	for _, p := range m.Pairs {
		for _, r := range p.Within {
			if row, ok := r.(map[string]any); ok {
				within++
				appendix += fmt.Sprintf("- %s #%v: %v\n", p.Frame, row["row"], row["values"])
			}
		}
	}
	if within > 0 {
		v.writeOut(filepath.Join(vvDir, "within-tolerance.md"), appendix)
	} else {
		os.Remove(filepath.Join(vvDir, "within-tolerance.md"))
	}
	v.say("within-tolerance: %d", within)
	// A task's own measure (key `task-<id>`) is the record check-ui-task-measured reads.
	if strings.HasPrefix(key, "task-") {
		v.writeOut(filepath.Join(vvDir, key+"-measure.json"), string(b)+"\n")
	}
	v.say("MANIFEST: %s", path)
	if failed {
		Reason(v.env, "measure-failed")
		panic(vvpExit(1))
	}
	panic(vvpExit(0))
}

// vvmFrameLine is measure's `frame:` line for one measured pair
// (visualverifypass.go's header, step 3, is canonical for its shape).
func vvmFrameLine(p vvmPair) string {
	var rows []string
	for _, o := range p.Over {
		r, _ := o.(map[string]any)
		n, _ := r["row"].(float64)
		s := "#" + strconv.FormatFloat(n, 'f', -1, 64)
		st, _ := r["status"].(string)
		if st != "" && st != "paired" {
			s += " " + st
		}
		keys, _ := r["keys"].(string)
		figures, _ := r["figures"].(map[string]map[string]string)
		var parts []string
		for _, k := range strings.Split(keys, ", ") {
			if f, ok := figures[k]; ok {
				parts = append(parts, k+" "+f["a"]+"→"+f["b"])
			} else if k != "" && k != st {
				parts = append(parts, k)
			}
		}
		if len(parts) > 0 {
			s += " " + strings.Join(parts, ", ")
		}
		rows = append(rows, s)
	}
	over := "none"
	if len(rows) > 0 {
		over = strings.Join(rows, "; ")
	}
	return "frame: " + p.Frame + " — " + p.Composite + " diff=" + p.Diff + " — over: " + over
}

// vvmUnion is a then b, each line once.
func vvmUnion(a, b []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, l := range append(append([]string{}, a...), b...) {
		if !seen[l] {
			seen[l] = true
			out = append(out, l)
		}
	}
	return out
}

func vvmLines(s string) []string {
	out := []string{}
	for _, l := range strings.Split(s, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}

// vvmFigures is an over row's own figures (decision over-items-quote-figures):
// per over key other than `unresolved`, the a and b of the matrix row the
// over row indexes (1-based), its top-level key else measured.delta's.
func vvmFigures(matrix []map[string]any, over map[string]any) map[string]map[string]string {
	figures := map[string]map[string]string{}
	row, _ := over["row"].(float64)
	keys, _ := over["keys"].(string)
	i := int(row) - 1
	if i < 0 || i >= len(matrix) {
		return figures
	}
	for _, k := range strings.Split(keys, ", ") {
		if k == "" || k == "unresolved" {
			continue
		}
		v, ok := matrix[i][k].(map[string]any)
		if !ok {
			measured, _ := matrix[i]["measured"].(map[string]any)
			delta, _ := measured["delta"].(map[string]any)
			v, _ = delta[k].(map[string]any)
		}
		a, okA := vvmFigure(v["a"])
		b, okB := vvmFigure(v["b"])
		if okA && okB {
			figures[k] = map[string]string{"a": a, "b": b}
		}
	}
	return figures
}

// vvmFigure formats a number shortest-exact and keeps a string as written.
func vvmFigure(v any) (string, bool) {
	switch t := v.(type) {
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64), true
	case string:
		return t, true
	}
	return "", false
}

// vvmShiftKeys are the over keys a constant vertical offset alone can raise.
var vvmShiftKeys = []string{"gap", "since_pair", "gap.top", "gap.bottom", "box.top", "box.bottom", "ink.top"}

// vvmFoldShift folds a whole-frame shift into one over row: when more than
// vvmShiftRows paired over rows — or any number of them, when every paired
// matrix row's dy sits inside their 1px window (a uniform offset) — carry only vvmShiftKeys and sit the same dy
// (b.top − a.top, within 1px of each other, beyond tol) from the frame's, a
// constant offset is one defect — a shared bar, a scroll position — not one
// per row. The first of them stays on over, the rest leave it, and the
// returned shift names dy and their count; nil when no such run exists. A row
// over on a colour, a size, a horizontal leaf, or missing, extra or
// unresolved is never folded.
func vvmFoldShift(matrix []map[string]any, over *[]any, tol float64) *vvmShift {
	type cand struct {
		row int
		dy  float64
	}
	var cands []cand
	for _, o := range *over {
		r, _ := o.(map[string]any)
		row, _ := r["row"].(float64)
		keys, _ := r["keys"].(string)
		i := int(row) - 1
		if r["status"] != "paired" || keys == "" || i < 0 || i >= len(matrix) ||
			slices.ContainsFunc(strings.Split(keys, ", "), func(k string) bool { return !slices.Contains(vvmShiftKeys, k) }) {
			continue
		}
		a, _ := matrix[i]["a"].(map[string]any)
		b, _ := matrix[i]["b"].(map[string]any)
		at, okA := a["top"].(float64)
		bt, okB := b["top"].(float64)
		if dy := bt - at; okA && okB && (dy > tol || dy < -tol) {
			cands = append(cands, cand{int(row), dy})
		}
	}
	slices.SortStableFunc(cands, func(x, y cand) int { return cmp.Compare(x.dy, y.dy) })
	lo, hi := 0, 0 // the widest window of dys within 1px of each other
	for i, j := 0, 0; j < len(cands); j++ {
		for cands[j].dy-cands[i].dy > 1 {
			i++
		}
		if j+1-i > hi-lo {
			lo, hi = i, j+1
		}
	}
	if hi == lo {
		return nil
	}
	run := cands[lo:hi]
	uniform := true
	for _, r := range matrix {
		a, _ := r["a"].(map[string]any)
		b, _ := r["b"].(map[string]any)
		at, okA := a["top"].(float64)
		bt, okB := b["top"].(float64)
		if okA && okB && (bt-at < run[0].dy || bt-at > run[len(run)-1].dy) {
			uniform = false
		}
	}
	if hi-lo <= vvmShiftRows && !uniform {
		return nil
	}
	folded := map[int]bool{}
	first := run[0].row
	for _, c := range run {
		folded[c.row] = true
		first = min(first, c.row)
	}
	*over = slices.DeleteFunc(*over, func(o any) bool {
		r, _ := o.(map[string]any)
		row, _ := r["row"].(float64)
		return folded[int(row)] && int(row) != first
	})
	return &vvmShift{Dy: run[len(run)/2].dy, Rows: len(run), Row: first, Uniform: uniform}
}

func vvmExists(p string) bool { _, err := os.Stat(p); return err == nil }

// vvmFlagged keeps the band or seam entries measure-visual-properties'
// tolerances flag: a missing or extra record, or a length delta over
// tol or colour distance over mvpOverColour anywhere inside it.
func vvmFlagged(list any, tol float64) []any {
	out := []any{}
	l, _ := list.([]any)
	for _, e := range l {
		if vvmOver(e, tol) {
			out = append(out, e)
		}
	}
	return out
}

func vvmOver(v any, tol float64) bool {
	switch t := v.(type) {
	case []any:
		for _, e := range t {
			if vvmOver(e, tol) {
				return true
			}
		}
	case map[string]any:
		if s := t["status"]; s == "missing" || s == "extra" {
			return true
		}
		if d, ok := t["abs"].(float64); ok && (d > tol || d < -tol) {
			return true
		}
		if d, ok := t["distance"].(float64); ok && d > mvpOverColour {
			return true
		}
		for k, e := range t {
			if k != "a" && k != "b" && vvmOver(e, tol) {
				return true
			}
		}
	}
	return false
}

// mockupsLines is a `.mockups` map's `<screenshot> <frame id>` lines, less
// its `frame` line, comments and malformed lines.
func mockupsLines(body []byte) [][2]string {
	var out [][2]string
	for _, raw := range strings.Split(string(body), "\n") {
		if f := strings.Fields(raw); len(f) == 2 && f[0] != "frame" && !strings.HasPrefix(f[0], "#") {
			out = append(out, [2]string{f[0], f[1]})
		}
	}
	return out
}
