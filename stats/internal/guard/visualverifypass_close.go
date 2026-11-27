package guard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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
	var frames []string
	if strings.Contains(v.key, "-rest") { // the repository suffix follows the segment
		// A remainder answers for the frames its own report names; the other
		// captures of the specs it measures belong to earlier passes.
		if ids := vvcReportFrames(body); ids != "" {
			frames = []string{"--frames", ids}
		}
	}
	mockups := ""
	if m := v.cells(w)["mockups"]; m != "" {
		mockups = cellPath(w, m)
	}
	verdicts, judgeGap := v.reportChecks(report, design, mockups, frames)
	if l := v.recordPassCheck(changeRoot); l != "" {
		verdicts = append(verdicts, l)
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
			if k, _, _ := strings.Cut(l[1], " "); l[0] == "blocked" && v.rowWorktree(stage, k) == w {
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
		v.recordJudged(w, body, design)
		v.commit(w, body)
		v.endStage()
		v.escalate(w, changeRoot, departures)
		v.next("loop")
	}
	if departures > 0 || judgeGap != "" {
		v.say("%s", verdict)
		if judgeGap != "" {
			v.say("%s", judgeGap)
		}
		v.recordJudged(w, body, design)
		v.commit(w, body)
		v.escalate(w, changeRoot, departures)
		v.next("judge")
	}
	v.endRow("completed")
	v.recordJudged(w, body, design)
	v.commit(w, body)
	v.endStage()
	v.next("done")
}

// recordJudged rewrites the worktree's visual-judged.txt from the row's
// manifest (judgedAfter); a lost write only makes the next scoped round
// re-judge those frames.
func (v *vvp) recordJudged(w string, report []byte, design string) {
	sdd := filepath.Join(w, ".superpowers", "sdd")
	var m vvmManifestOut
	b, err := os.ReadFile(filepath.Join(sdd, "visual-manifest-"+vrRestSegment.ReplaceAllString(v.key, "")+".json"))
	if err != nil || json.Unmarshal(b, &m) != nil {
		return
	}
	d, _ := os.ReadFile(design)
	_ = writeJudged(sdd, judgedAfter(readJudged(sdd), m.Pairs, report, d))
}

// reportChecks runs the report guards on report: the verdict lines of an
// incomplete report, and the one gap that is only a missing sidecar, left to
// the judge. mockups is "" when none is declared.
func (v *vvp) reportChecks(report, design, mockups string, frames []string) (verdicts []string, judgeGap string) {
	args := []string{report, strconv.Itoa(v.motions)}
	if design != "" {
		args = append(args, "--design", design)
	}
	rc, out, errs := v.guard("check-verify-report", nil, append(args, frames...)...)
	switch rc {
	case 0:
	case 1:
		verdicts = append(verdicts, out)
	default:
		v.die("check-verify-report-cannot-answer", "check-verify-report cannot answer: %s", errs)
	}
	if mockups != "" {
		rc, out, errs := v.guard("check-frame-list", nil, append([]string{design, mockups, report}, frames...)...)
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
	return verdicts, judgeGap
}

// checkOnly is `close --check-only <report> <motions named> ...`: the report
// guards alone, exit 1 with their messages, 0 on a complete report, and no
// dispatch row, stage line or verdict written either way.
func (v *vvp) checkOnly(args []string) {
	if len(args) < 2 {
		v.usage("--check-only takes <report> <motions named>")
	}
	report := v.abs(args[0])
	if n, err := fmt.Sscan(args[1], &v.motions); n != 1 || err != nil || v.motions < 0 {
		v.usage("<motions named> '" + args[1] + "' is not a count")
	}
	design, mockups := "", ""
	for i := 2; i < len(args); i += 2 {
		if i+1 >= len(args) {
			v.usage(args[i] + " takes a value")
		}
		switch args[i] {
		case "--design":
			design = v.abs(args[i+1])
		case "--mockups":
			mockups = v.abs(args[i+1])
		default:
			v.usage("unknown argument '" + args[i] + "'")
		}
	}
	if mockups != "" && design == "" {
		v.usage("--mockups needs --design")
	}
	if _, err := os.Stat(report); err != nil {
		v.die("cannot-read", "cannot read %s: %v", report, err)
	}
	verdicts, gap := v.reportChecks(report, design, mockups, nil)
	if gap != "" {
		verdicts = append(verdicts, gap)
	}
	for _, l := range verdicts {
		v.say("%s", l)
	}
	if len(verdicts) > 0 {
		Reason(v.env, "report-incomplete")
		panic(vvpExit(1))
	}
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

// vvcExitNote is a parenthesised note on an exit (`capture 0 (28 passed)`): its counts are not exits.
var vvcExitNote = regexp.MustCompile(`\([^)]*\)`)

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
			for _, n := range vvcExit.FindAllString(vvcExitNote.ReplaceAllString(val, ""), -1) {
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

// suffix is `-<repository name>` when the set holds more than one worktree.
func (v *vvp) suffix(w string) string {
	if v.setSize > 1 {
		return "-" + v.repoName(w)
	}
	return ""
}

// vvcFixRound is a key's `-fix-<k>[-full]` segment, which a redispatch keeps
// (decision redispatch-key-keeps-round).
var vvcFixRound = regexp.MustCompile(`-fix-([0-9]+)(-full)?`)

// vvcOpener is one `In-run fix <k>` task's title and its `**Opener:**` key.
var vvcOpener = regexp.MustCompile(`(?m)In-run fix ([0-9]+) —[^\n]*\n\*\*Opener:\*\* *(\S+)`)

// escalate prints `ESCALATE: <why>` when the round this row closes failed
// without progress, so the loop's next round is the No-progress escalation:
// round 3 or later, or more departures than the newest visual report that
// opened a round from this one's down — read from <changeRoot>/tasks.md's
// **Opener:** lines, its report in w (kan-1091: rounds 4 to 8 ran on the
// base pair while the departures went 9, 19, 26, 24).
func (v *vvp) escalate(w, changeRoot string, departures int) {
	m := vvcFixRound.FindStringSubmatch(v.key)
	if m == nil {
		return
	}
	k, _ := strconv.Atoi(m[1])
	why := ""
	if k >= 3 {
		why = fmt.Sprintf("round %d is the loop's third failed round or later", k)
	} else if tasks, err := os.ReadFile(filepath.Join(changeRoot, "tasks.md")); err == nil {
		openers := map[int]string{}
		for _, o := range vvcOpener.FindAllStringSubmatch(string(tasks), -1) {
			j, _ := strconv.Atoi(o[1])
			openers[j] = o[2]
		}
		for j := k; j >= 1; j-- {
			key := openers[j]
			body, err := os.ReadFile(filepath.Join(w, ".superpowers", "sdd", "verify-report-"+key+".md"))
			if !strings.HasPrefix(key, "visual-verify") || err != nil {
				continue
			}
			if _, _, base := vvcParseVerdict(vvcVerdictBlock(string(body))); departures > base {
				why = fmt.Sprintf("departures %d → %d since %s", base, departures, key)
			}
			break
		}
	}
	if why != "" {
		v.say("ESCALATE: %s — the next round is a No-progress escalation", why)
	}
}

// vvcRecordPass is the line a record opens with: the key of the pass that wrote it.
var vvcRecordPass = regexp.MustCompile(`(?m)^Pass: *(\S+) *$`)

// recordPassCheck refuses a record a pass leaves stale (KAN-1155: the change
// landed with a visual-verification.md naming fix-13-rest after fix-21-rest
// closed done). visual-verification.md must carry a `Pass: <key>` line naming
// this row's key; one that names another pass, or none, is the verdict line an
// incomplete report is. No record file at all is no verdict: step 11 writes it
// only when a view was captured.
func (v *vvp) recordPassCheck(changeRoot string) string {
	rec := filepath.Join(changeRoot, "visual-verification.md")
	b, err := os.ReadFile(rec)
	if err != nil {
		return ""
	}
	want := v.key
	got := "no pass"
	if m := vvcRecordPass.FindSubmatch(b); m != nil {
		got = string(m[1])
	}
	if got == want {
		return ""
	}
	return fmt.Sprintf("RECORD-STALE: %s names %s, not this pass %s — rewrite it whole under this pass's key, opening with `Pass: %s` (step 11 of skills/flow/visual-verify-verifier.md)", rec, got, want, want)
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
	// The newest row of each worktree.
	newest, closed := map[string]string{}, map[string]string{}
	var order []string
	for _, l := range stage {
		k, rest, _ := strings.Cut(l[1], " ")
		switch l[0] {
		case "worktree":
			order = append(order, k)
		case "row":
			newest[rest] = k
		case "closed":
			closed[k] = rest
		}
	}
	outcome := "completed"
	for _, w := range order {
		k, seen := newest[w]
		if !seen {
			return
		}
		o, ok := closed[k]
		if !ok {
			return
		}
		if o == "blocked" {
			outcome = "stopped"
		}
	}
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

var vvcFrameID = regexp.MustCompile(`[A-Za-z0-9][\w.-]*/[A-Za-z0-9][\w.-]*`)

// vvcReportFrames is the comma-joined frame ids of a report's `- frames:` line
// (`<n>/<m> — <id>; <id>: <reason>`), "" when it names none.
func vvcReportFrames(report []byte) string {
	var ids []string
	for _, l := range strings.Split(string(report), "\n") {
		m := flFramesLine.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		ids = append(ids, vvcFrameID.FindAllString(m[3], -1)...)
	}
	return strings.Join(ids, ",")
}
