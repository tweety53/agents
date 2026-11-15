package guard

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// close is `visual-verify-pass.sh close`: the report guards, the row's end,
// step 12's commit and the stage's end, from the report the verifier wrote
// for <key> and the stage file prepare wrote. Its outcomes are the comment above's.
func (v *vvp) close() {
	stage := v.readStage()
	w := ""
	for _, l := range stage {
		if k, rw, _ := strings.Cut(l[1], " "); l[0] == "row" && k == v.key {
			w = rw
		}
		if l[0] == "set" {
			fmt.Sscan(l[1], &v.setSize)
		}
	}
	if w == "" {
		v.die("stage-file-has-no-row", "the stage file %s has no row %s — close names a key prepare began", v.stagePath(), v.key)
	}
	v.stageOpen = true
	report := filepath.Join(w, ".superpowers", "sdd", "verify-report-"+v.key+".md")
	body, err := os.ReadFile(report)
	if os.IsNotExist(err) {
		v.endRow("aborted")
		v.stop("no-report", "stopped", "no report — "+report+" was never written: the verifier ended without its report or its agent died")
	}
	if err != nil {
		v.die("cannot-read", "cannot read %s: %v", report, err)
	}
	changeRoot := filepath.Join(v.canon, specRootLeaf(v.canon, io.Discard), "changes", v.name)
	design := filepath.Join(changeRoot, "design.md")

	// The report guards: an incomplete report is never accepted.
	var verdicts []string
	judgeGap := ""
	// A shard's report owes its slice alone (decision manifest-per-shard).
	var frames []string
	if slice := vvpSlice(stage, v.key); slice != "" {
		frames = []string{"--frames", slice}
	}
	rc, out, errs := v.guard("check-verify-report", nil, append([]string{report, strconv.Itoa(v.motions), "--design", design}, frames...)...)
	switch rc {
	case 0:
	case 1:
		verdicts = append(verdicts, out)
	default:
		v.die("check-verify-report-cannot-answer", "check-verify-report cannot answer: %s", errs)
	}
	if m := v.cells(w)["mockups"]; m != "" {
		rc, out, errs := v.guard("check-frame-list", nil, append([]string{design, cellPath(w, m), report}, frames...)...)
		switch {
		case rc == 0:
		case rc == 1 && vvcOnlyNoSidecar(out):
			judgeGap = out
		case rc == 1:
			verdicts = append(verdicts, out)
		default:
			v.die("check-frame-list-cannot-answer", "check-frame-list cannot answer: %s", errs)
		}
	}
	if len(verdicts) > 0 {
		v.endRow("aborted")
		for _, l := range verdicts {
			v.appendStage("verdict " + v.key + " " + l)
			v.say("%s", l)
		}
		v.next("rest " + v.restKey(w))
	}

	verdict := vvcVerdictBlock(string(body))
	exits, cause, departures := vvcParseVerdict(verdict)
	if exits {
		prior := 0
		for _, l := range stage {
			if k, _, _ := strings.Cut(l[1], " "); l[0] == "blocked" && v.rowWorktree(stage, k) == w && vvpShardSeg.FindString(k) == vvpShardSeg.FindString(v.key) {
				prior++
			}
		}
		v.appendStage("blocked " + v.key)
		nextKey := fmt.Sprintf("visual-verify-%d%s%s", prior+2, vvcFixRound.FindString(v.key), v.suffix(w))
		if cause == "" {
			cause = "environment"
		}
		switch {
		case prior == 0:
			v.endRow("blocked", "-cause", cause)
			v.next("redispatch " + nextKey)
		case cause == "environment":
			v.endRow("blocked", "-cause", cause)
			v.next("cure-environment " + nextKey)
		}
		v.endRow("blocked", "-cause", cause)
		v.commit(w, body)
		v.joinRecords(changeRoot)
		v.endStage()
		v.next("loop")
	}
	if departures > 0 || judgeGap != "" {
		v.say("%s", verdict)
		if judgeGap != "" {
			v.say("%s", judgeGap)
		}
		v.commit(w, body)
		v.joinRecords(changeRoot)
		v.next("judge")
	}
	snap := protoSnapshot(filepath.Dir(design))
	scoped, why := v.closesLoop(w, design, body, snap)
	if scoped && why == "" {
		v.endRow("completed", "-closes-loop")
	} else {
		v.endRow("completed")
	}
	if snap != "" {
		// A lost snapshot only keeps the next scoped round's full run.
		_ = os.WriteFile(filepath.Join(w, ".superpowers", "sdd", "visual-prototype.txt"), []byte(snap), 0o644)
	}
	v.commit(w, body)
	v.joinRecords(changeRoot)
	v.endStage()
	switch {
	case scoped && why == "":
		v.say("CLOSES-LOOP: yes")
	case scoped:
		v.say("CLOSES-LOOP: no — %s", why)
	}
	v.next("done")
}

// closesLoop derives whether this row closes the visual fix loop (decisions
// closes-loop-derived-in-close, capture-freshness-by-mtime, closing-frame-list,
// frame-map-change-keeps-full-run, prototype-change-keeps-full-run): scoped is
// false with no round file of the key; why is "" when every condition holds,
// else the first that fails.
func (v *vvp) closesLoop(w, design string, report []byte, snap string) (scoped bool, why string) {
	sdd := filepath.Join(w, ".superpowers", "sdd")
	base, at, defects, ok := readRound(sdd, v.key)
	if !ok {
		return false, ""
	}
	mpath := filepath.Join(sdd, "visual-manifest-"+v.key+".json")
	var m vvmManifestOut
	b, err := os.ReadFile(mpath)
	if err == nil {
		err = json.Unmarshal(b, &m)
	}
	if err != nil {
		return true, "no readable manifest " + mpath
	}
	// The closing list: the declared frames, the prototype map's, the manifest's.
	var list []string
	add := func(id string) {
		if !slices.Contains(list, id) {
			list = append(list, id)
		}
	}
	mock := ""
	if c := v.cells(w)["mockups"]; c != "" {
		mock = cellPath(w, c)
		d, derr := os.ReadFile(design)
		entries, eerr := os.ReadDir(mock)
		if derr == nil && eerr == nil {
			for _, id := range flDeclared(d, mock, entries) {
				add(id)
			}
		}
	}
	proto, _ := os.ReadFile(filepath.Join(filepath.Dir(design), "visual-verification", "prototype.mockups"))
	for _, f := range mockupsLines(proto) {
		add(f[1])
	}
	for _, p := range m.Pairs {
		add(p.Frame)
	}
	if len(list) == 0 {
		return true, "no frame on the change's list"
	}
	for _, id := range list {
		fresh := false
		for _, p := range m.Pairs {
			if fi, err := os.Stat(p.Capture); p.Frame == id && p.Capture != "" && err == nil && fi.ModTime().After(at) {
				fresh = true
			}
		}
		if !fresh {
			return true, id + ": no capture in " + mpath + " re-captured after " + at.Format(time.RFC3339Nano)
		}
	}
	lines := flReportLines(report)
	for _, p := range m.Pairs {
		if v.captureChanged(p.Capture, base) && !flPerFrame(p.Frame, lines) {
			return true, p.Frame + ": its capture changed since " + base + " and the report has no line of its own for it"
		}
	}
	for _, id := range defects {
		if !flPerFrame(id, lines) {
			return true, id + ": named under defects: and the report has no line of its own for it"
		}
	}
	var specs []string
	for _, p := range m.Pairs {
		if s := p.Spec + ".mockups"; !slices.Contains(specs, s) && vvmExists(s) {
			specs = append(specs, s)
			if v.captureChanged(s, base) {
				return true, s + " changed since " + base
			}
		}
	}
	if mock != "" {
		out, ok := v.git("-C", w, "diff", "--no-renames", "--name-only", base, "--", mock)
		if !ok {
			return true, "git diff " + base + " -- " + mock + " failed in " + w
		}
		if out != "" {
			return true, strings.Split(out, "\n")[0] + " changed since " + base
		}
	}
	if snap != "" {
		path := filepath.Join(sdd, "visual-prototype.txt")
		if old, _ := os.ReadFile(path); string(old) != snap {
			return true, path + " differs from prototype.mockups and prototype/*.png now"
		}
	}
	return true, ""
}

// protoSnapshot is the prototype frames a verifier judges against (decision
// prototype-change-keeps-full-run): one `<sha-256 hex> <name>` line for
// <changeRoot>/visual-verification/prototype.mockups, then one per
// prototype/*.png in name order; "" when the change has no prototype map.
func protoSnapshot(changeRoot string) string {
	vvDir := filepath.Join(changeRoot, "visual-verification")
	mp := filepath.Join(vvDir, "prototype.mockups")
	if !vvmExists(mp) {
		return ""
	}
	pngs, _ := filepath.Glob(filepath.Join(vvDir, "prototype", "*.png"))
	var b strings.Builder
	for _, p := range append([]string{mp}, pngs...) {
		data, _ := os.ReadFile(p)
		rel, _ := filepath.Rel(vvDir, p)
		fmt.Fprintf(&b, "%x %s\n", sha256.Sum256(data), rel)
	}
	return b.String()
}

// vvcOnlyNoSidecar is a FRAME-LIST-GAP line whose every gap is `no sidecar
// line: <frame id>`.
func vvcOnlyNoSidecar(out string) bool {
	for _, l := range strings.Split(out, "\n") {
		_, gaps, ok := strings.Cut(l, "FRAME-LIST-GAP: ")
		if !ok {
			continue
		}
		_, gaps, _ = strings.Cut(gaps, " — ")
		for _, g := range strings.Split(gaps, "; ") {
			if !strings.HasPrefix(g, "no sidecar line: ") {
				return false
			}
		}
		return true
	}
	return false
}

// vvcVerdict is the report's `## Verdict` block, up to the next heading.
func vvcVerdictBlock(report string) string {
	var out []string
	in := false
	for _, l := range strings.Split(report, "\n") {
		if strings.HasPrefix(l, "## ") {
			if in {
				break
			}
			in = strings.TrimSpace(l) == "## Verdict"
		}
		if in {
			out = append(out, l)
		}
	}
	return strings.TrimRight(strings.Join(out, "\n"), "\n")
}

var vvcExit = regexp.MustCompile(`-?[0-9]+`)

// vvcParseVerdict reads the `## Verdict` block: whether `exits:` names a
// non-zero exit, its `cause:` and its `departures:` count.
func vvcParseVerdict(verdict string) (bool, string, int) {
	nonZero, cause, departures := false, "", 0
	for _, l := range strings.Split(verdict, "\n") {
		k, val, ok := strings.Cut(strings.TrimPrefix(strings.TrimSpace(l), "- "), ": ")
		if !ok {
			continue
		}
		switch k {
		case "exits":
			for _, n := range vvcExit.FindAllString(val, -1) {
				if n != "0" {
					nonZero = true
				}
			}
		case "cause":
			cause = strings.TrimSpace(val)
		case "departures":
			departures, _ = strconv.Atoi(strings.TrimSpace(val))
		}
	}
	return nonZero, cause, departures
}

// rowWorktree is the worktree the stage file's row <key> names.
func (v *vvp) rowWorktree(stage [][2]string, key string) string {
	w := ""
	for _, l := range stage {
		if k, rw, _ := strings.Cut(l[1], " "); l[0] == "row" && k == key {
			w = rw
		}
	}
	return w
}

// suffix is the key's shard segment, `-s<i>of<n>`, then `-<repository name>`
// when the set holds more than one worktree.
func (v *vvp) suffix(w string) string {
	if v.setSize > 1 {
		return "-" + v.repoName(w)
	}
	return vvpShardSeg.FindString(v.key)
}

// vvpShardSeg is a shard key's `-s<i>of<n>` segment (decision shard-key-segment).
var vvpShardSeg = regexp.MustCompile(`-s([0-9]+)of([0-9]+)$`)

// vvpSlice is the stage file's `slice` of key's shard segment, "" for none.
func vvpSlice(stage [][2]string, key string) string {
	seg := vvpShardSeg.FindString(key)
	for _, l := range stage {
		if k, slice, _ := strings.Cut(l[1], " "); seg != "" && l[0] == "slice" && "-"+k == seg {
			return slice
		}
	}
	return ""
}

// vvcFixRound is a key's `-fix-<k>[-full]` segment, which a redispatch keeps
// (decision redispatch-key-keeps-round).
var vvcFixRound = regexp.MustCompile(`-fix-[0-9]+(-full)?`)

// joinRecords rewrites <changeRoot>/visual-verification.md as the shards' record-s<i>of<n>.md parts that exist, in shard order, on a
// shard's close (decision record-parts-merged-by-close).
func (v *vvp) joinRecords(changeRoot string) {
	m := vvpShardSeg.FindStringSubmatch(v.key)
	if m == nil {
		return
	}
	n, _ := strconv.Atoi(m[2])
	vvDir := filepath.Join(changeRoot, "visual-verification")
	var parts []string
	for i := 1; i <= n; i++ {
		part := filepath.Join(vvDir, fmt.Sprintf("record-s%dof%d.md", i, n))
		b, err := os.ReadFile(part)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			v.die("cannot-read", "cannot read %s: %v", part, err)
		}
		parts = append(parts, strings.TrimRight(string(b), "\n"))
	}
	v.writeOut(filepath.Join(changeRoot, "visual-verification.md"), strings.Join(parts, "\n\n")+"\n")
}

var vvcRest = regexp.MustCompile(`-rest(?:-([0-9]+))?$`)

// restKey is the remainder's key: `-rest`, then `-rest-<n>` from 2, placed
// before any `-<repository name>` suffix.
func (v *vvp) restKey(w string) string {
	suffix := v.suffix(w)
	base := strings.TrimSuffix(v.key, suffix)
	m := vvcRest.FindStringSubmatch(base)
	switch {
	case m == nil:
		base += "-rest"
	case m[1] == "":
		base += "-2"
	default:
		n, _ := strconv.Atoi(m[1])
		base = strings.TrimSuffix(base, m[0]) + "-rest-" + strconv.Itoa(n+1)
	}
	return base + suffix
}

// endRow is the row's `flow record dispatch end`, recorded in the stage file
// as `closed <key> <outcome>`.
func (v *vvp) endRow(outcome string, flags ...string) {
	argv := append([]string{"record", "dispatch", "end", "-C", v.canon, "-change", v.name, "-key", v.key,
		"-session-token", v.token, "-outcome", outcome}, flags...)
	if out, rc := v.flow(argv...); rc != 0 {
		v.die("flow-record-dispatch-end-failed", "flow record dispatch end failed for %s: %s", v.key, out)
	}
	v.appendStage("closed " + v.key + " " + outcome)
}

// endStage ends the stage once every surviving worktree's newest row is
// closed — `stopped` when any is blocked, else `completed` — renders the
// ledger and prints `STAGE: <outcome>`.
func (v *vvp) endStage() {
	stage := v.readStage()
	// The newest row of each worktree and shard index (decision shard-key-segment).
	newest, closed := map[string]string{}, map[string]string{}
	var order []string
	for _, l := range stage {
		k, rest, _ := strings.Cut(l[1], " ")
		switch l[0] {
		case "worktree":
			order = append(order, k)
		case "row":
			newest[rest+" "+vvpShardSeg.FindString(k)] = k
		case "closed":
			closed[k] = rest
		}
	}
	outcome := "completed"
	for _, w := range order {
		seen := false
		for lane, k := range newest {
			if !strings.HasPrefix(lane, w+" ") {
				continue
			}
			seen = true
			o, ok := closed[k]
			if !ok {
				return
			}
			if o == "blocked" {
				outcome = "stopped"
			}
		}
		if !seen {
			return
		}
	}
	v.markEnd(outcome)
	if out, rc := v.flow("record", "render", "-change", v.name, "-kind", "ledger", "-repo", v.canon); rc != 0 {
		fmt.Fprintf(v.stderr, "visual-verify-pass: warning: flow record render failed: %s\n", out)
	}
	v.say("STAGE: %s", outcome)
}

// vvcPath is an absolute path token of a report line.
var vvcPath = regexp.MustCompile(`/[^\s,;]+`)

// commit is step 12 (decision close-commits-final-reports): the report's
// `spec:` path, every PNG, and the full app suite's spec, zip and resolved
// PNGs — each in the checkout it lives in,
// pathspec-scoped, under `test: record the <key> visual verification` with
// trailer `Flow-Key: flow.visual-verify`. A checkout on a branch other than
// spectre/<name> prints `PUSH: git -C <checkout> push`. Then
// `ROUND-BASE: <worktree>=<HEAD>`.
func (v *vvp) commit(w string, report []byte) {
	byTop := map[string][]string{}
	var tops []string
	add := func(p string) {
		if _, err := os.Stat(p); err != nil {
			return
		}
		top, ok := v.git("-C", filepath.Dir(p), "rev-parse", "--show-toplevel")
		if !ok {
			return
		}
		top = realPath(top)
		// Ignored artifacts ride the change's archive at cleanup, not this commit.
		if _, ignored := v.gitChild(top, "check-ignore", "-q", "--", p); ignored {
			return
		}
		if _, seen := byTop[top]; !seen {
			tops = append(tops, top)
		}
		byTop[top] = append(byTop[top], p)
	}
	for _, l := range strings.Split(string(report), "\n") {
		l = strings.TrimSpace(l)
		if !strings.HasPrefix(l, "- ") {
			continue
		}
		isSpec := strings.HasPrefix(l, "- spec:") || strings.HasPrefix(l, "- full app suite")
		for _, p := range vvcPath.FindAllString(l, -1) {
			if isSpec || strings.HasSuffix(p, ".png") {
				add(p)
			}
			// The report counts the suite's PNGs without naming them: resolve
			// them as rebuild-full-app-suite-zip does.
			if strings.HasSuffix(p, "/full-app-suite.spec.ts") {
				var out, serr bytes.Buffer
				if rc := Registry["resolve-visual-screenshots"]([]string{w, "full-app-suite.spec.ts"}, quietReasons(v.env), &out, &serr); rc == 2 {
					fmt.Fprintf(v.stderr, "visual-verify-pass: warning: the full app suite's PNGs could not be resolved, so none is committed: %s", serr.String())
				}
				for _, png := range strings.Fields(out.String()) {
					add(png)
				}
			}
		}
	}
	for _, top := range tops {
		paths := byTop[top]
		if out, ok := v.gitChild(top, append([]string{"add", "--"}, paths...)...); !ok {
			v.die("git-add-in-failed", "git add in %s failed: %s", top, out)
		}
		if _, ok := v.gitChild(top, append([]string{"diff", "--cached", "--quiet", "--"}, paths...)...); ok {
			continue
		}
		if out, ok := v.gitChild(top, append([]string{"commit", "-q", "-m", "test: record the " + v.key + " visual verification",
			"-m", "Flow-Key: flow.visual-verify", "--"}, paths...)...); !ok {
			v.die("git-commit-in-failed", "git commit in %s failed: %s", top, out)
		}
		if branch, _ := v.gitChild(top, "branch", "--show-current"); branch != "spectre/"+v.name {
			v.say("PUSH: git -C %s push", top)
		}
	}
	head, ok := v.git("-C", w, "rev-parse", "HEAD")
	if !ok {
		v.die("head-unresolved", "cannot resolve HEAD in %s", w)
	}
	v.say("ROUND-BASE: %s=%s", w, head)
}

// gitChild runs git in dir under the caller's environment (Env.Environ):
// its combined output and whether it exited 0.
func (v *vvp) gitChild(dir string, args ...string) (string, bool) {
	git := "git"
	if p, ok := lookPath(v.env, "git"); ok {
		git = p
	}
	cmd := exec.Command(git, append([]string{"-C", dir}, args...)...)
	cmd.Env = icChildEnv(v.env)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err == nil
}
