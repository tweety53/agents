package guard

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
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
}

type vvmManifestOut struct {
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
	changeRoot := filepath.Join(w, specRootLeaf(w, io.Discard), "changes", name)
	vvDir := filepath.Join(changeRoot, "visual-verification")
	sdd := filepath.Join(w, ".superpowers", "sdd")
	bandsDir := filepath.Join(sdd, "visual-measure-"+key)
	matrixDir := filepath.Join(vvDir, "matrix")
	for _, d := range []string{bandsDir, matrixDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			v.die("cannot create output directory %s: %v", d, err)
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
	m := vvmManifestOut{Captures: []string{}, Identical: []string{}, Steps: map[string]vvmStep{}, Pairs: []vvmPair{}}

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
		v.writeManifest(sdd, key, spec, m, failed)
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
	own := "frame none\n"
	for _, raw := range strings.Split(string(body), "\n") {
		f := strings.Fields(raw)
		if len(f) != 2 || f[0] == "frame" || strings.HasPrefix(f[0], "#") {
			continue
		}
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
		p := vvmPair{Spec: spec, Frame: l.frame, Mockup: filepath.Join(mockups, l.frame+".png"),
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

		rc, out, serr := run(measureVisualProperties, "", p.Mockup, p.Capture, "--scale", "1", "--props", "bands,seams")
		p.Exits.Bands, p.Stderr["bands"] = note(rc), serr
		p.BandsFile = filepath.Join(bandsDir, l.frame+".json")
		v.writeOut(p.BandsFile, out)
		var bs map[string]any
		if json.Unmarshal([]byte(out), &bs) == nil {
			p.Bands, p.Seams = vvmFlagged(bs["bands"]), vvmFlagged(bs["seams"])
		}

		rc, out, serr = run(measureVisualProperties, "", p.Mockup, p.Capture, "--scale", "1", "--props", "matrix")
		p.Exits.Matrix, p.Stderr["matrix"] = note(rc), serr
		p.MatrixFile = filepath.Join(matrixDir, l.frame+".json")
		v.writeOut(p.MatrixFile, out)
		var mx struct {
			Matrix  []map[string]any                `json:"matrix"`
			Summary struct{ Over []map[string]any } `json:"summary"`
		}
		if json.Unmarshal([]byte(out), &mx) == nil && mx.Summary.Over != nil {
			for _, r := range mx.Summary.Over {
				r["figures"] = vvmFigures(mx.Matrix, r)
				p.Over = append(p.Over, r)
			}
		}
		m.Pairs = append(m.Pairs, p)
	}
	v.writeManifest(sdd, key, spec, m, failed)
}

func (v *vvp) writeOut(path, body string) {
	// A frame id carries a directory (`kan-1083/SE1`), so its output file sits below the key's directory.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		v.die("cannot write %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		v.die("cannot write %s: %v", path, err)
	}
}

// writeManifest merges m into the key's manifest, prints its path and exits 1
// when any step failed, else 0. An existing manifest that parses keeps its
// pairs of other specs whose frame m does not carry, and its captures and
// identical lines ahead of m's; steps are m's alone (kan-1102 self-review, F2).
func (v *vvp) writeManifest(sdd, key, spec string, m vvmManifestOut, failed bool) {
	path := filepath.Join(sdd, "visual-manifest-"+key+".json")
	var old vvmManifestOut
	if b, err := os.ReadFile(path); err == nil && json.Unmarshal(b, &old) == nil {
		mine := map[string]bool{}
		for _, p := range m.Pairs {
			mine[p.Frame] = true
		}
		var pairs []vvmPair
		for _, p := range old.Pairs {
			if p.Spec != spec && !mine[p.Frame] {
				pairs = append(pairs, p)
			}
		}
		m.Pairs = append(pairs, m.Pairs...)
		m.Captures = vvmUnion(old.Captures, m.Captures)
		m.Identical = vvmUnion(old.Identical, m.Identical)
	}
	b, _ := json.MarshalIndent(m, "", "  ")
	v.writeOut(path, string(b)+"\n")
	v.say("MANIFEST: %s", path)
	if failed {
		panic(vvpExit(1))
	}
	panic(vvpExit(0))
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

func vvmExists(p string) bool { _, err := os.Stat(p); return err == nil }

// vvmFlagged keeps the band or seam entries measure-visual-properties'
// tolerances flag: a missing or extra record, or a length delta over
// mvpOverLength or colour distance over mvpOverColour anywhere inside it.
func vvmFlagged(list any) []any {
	out := []any{}
	l, _ := list.([]any)
	for _, e := range l {
		if vvmOver(e) {
			out = append(out, e)
		}
	}
	return out
}

func vvmOver(v any) bool {
	switch t := v.(type) {
	case []any:
		for _, e := range t {
			if vvmOver(e) {
				return true
			}
		}
	case map[string]any:
		if s := t["status"]; s == "missing" || s == "extra" {
			return true
		}
		if d, ok := t["abs"].(float64); ok && (d > mvpOverLength || d < -mvpOverLength) {
			return true
		}
		if d, ok := t["distance"].(float64); ok && d > mvpOverColour {
			return true
		}
		for k, e := range t {
			if k != "a" && k != "b" && vvmOver(e) {
				return true
			}
		}
	}
	return false
}
