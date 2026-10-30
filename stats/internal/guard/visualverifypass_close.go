package guard

import (
	"bytes"
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
// for <key> and the stage file prepare wrote. Its outcomes are the shim
// header's.
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
		v.die("the stage file %s has no row %s — close names a key prepare began", v.stagePath(), v.key)
	}
	v.stageOpen = true
	report := filepath.Join(w, ".superpowers", "sdd", "verify-report-"+v.key+".md")
	body, err := os.ReadFile(report)
	if os.IsNotExist(err) {
		v.endRow("aborted")
		v.stop("stopped", "no report — "+report+" was never written: the verifier ended without its report or its agent died")
	}
	if err != nil {
		v.die("cannot read %s: %v", report, err)
	}
	changeRoot := filepath.Join(v.canon, specRootLeaf(v.canon, io.Discard), "changes", v.name)
	design := filepath.Join(changeRoot, "design.md")

	// The report guards: an incomplete report is never accepted.
	var verdicts []string
	judgeGap := ""
	rc, out, errs := v.guard("check-verify-report", nil, report, strconv.Itoa(v.motions), "--design", design)
	switch rc {
	case 0:
	case 1:
		verdicts = append(verdicts, out)
	default:
		v.die("check-verify-report cannot answer: %s", errs)
	}
	if m := v.cells(w)["mockups"]; m != "" {
		rc, out, errs := v.guard("check-frame-list", nil, design, cellPath(w, m), report)
		switch {
		case rc == 0:
		case rc == 1 && vvcOnlyNoSidecar(out):
			judgeGap = out
		case rc == 1:
			verdicts = append(verdicts, out)
		default:
			v.die("check-frame-list cannot answer: %s", errs)
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
			if k, _, _ := strings.Cut(l[1], " "); l[0] == "blocked" && v.rowWorktree(stage, k) == w {
				prior++
			}
		}
		v.appendStage("blocked " + v.key)
		suffix := v.suffix(w)
		nextKey := fmt.Sprintf("visual-verify-%d%s", prior+2, suffix)
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
		v.commit(w, body, changeRoot)
		v.endStage()
		v.next("loop")
	}
	if departures > 0 || judgeGap != "" {
		v.say("%s", verdict)
		if judgeGap != "" {
			v.say("%s", judgeGap)
		}
		v.commit(w, body, changeRoot)
		v.next("judge")
	}
	v.endRow("completed")
	v.commit(w, body, changeRoot)
	v.endStage()
	v.next("done")
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

// suffix is `-<repository name>` when the set holds more than one worktree.
func (v *vvp) suffix(w string) string {
	if v.setSize > 1 {
		return "-" + v.repoName(w)
	}
	return ""
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
		v.die("flow record dispatch end failed for %s: %s", v.key, out)
	}
	v.appendStage("closed " + v.key + " " + outcome)
}

// endStage ends the stage once every surviving worktree's newest row is
// closed — `stopped` when any is blocked, else `completed` — renders the
// ledger and prints `STAGE: <outcome>`.
func (v *vvp) endStage() {
	stage := v.readStage()
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
		o, ok := closed[newest[w]]
		if !ok {
			return
		}
		if o == "blocked" {
			outcome = "stopped"
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
// `spec:` path, every PNG, the full app suite's spec, zip and resolved PNGs, and
// <changeRoot>/visual-verification/ — each in the checkout it lives in,
// pathspec-scoped, under `test: record the <key> visual verification` with
// trailer `Flow-Key: flow.visual-verify`. A checkout on a branch other than
// spectre/<name> prints `PUSH: git -C <checkout> push`. Then
// `ROUND-BASE: <worktree>=<HEAD>`.
func (v *vvp) commit(w string, report []byte, changeRoot string) {
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
				if rc := Registry["resolve-visual-screenshots"]([]string{w, "full-app-suite.spec.ts"}, v.env, &out, &serr); rc == 2 {
					fmt.Fprintf(v.stderr, "visual-verify-pass: warning: the full app suite's PNGs could not be resolved, so none is committed: %s", serr.String())
				}
				for _, png := range strings.Fields(out.String()) {
					add(png)
				}
			}
		}
	}
	add(filepath.Join(changeRoot, "visual-verification"))
	for _, top := range tops {
		paths := byTop[top]
		if out, ok := v.gitChild(top, append([]string{"add", "--"}, paths...)...); !ok {
			v.die("git add in %s failed: %s", top, out)
		}
		if _, ok := v.gitChild(top, append([]string{"diff", "--cached", "--quiet", "--"}, paths...)...); ok {
			continue
		}
		if out, ok := v.gitChild(top, append([]string{"commit", "-q", "-m", "test: record the " + v.key + " visual verification",
			"-m", "Flow-Key: flow.visual-verify", "--"}, paths...)...); !ok {
			v.die("git commit in %s failed: %s", top, out)
		}
		if branch, _ := v.gitChild(top, "branch", "--show-current"); branch != "spectre/"+v.name {
			v.say("PUSH: git -C %s push", top)
		}
	}
	head, ok := v.git("-C", w, "rev-parse", "HEAD")
	if !ok {
		v.die("cannot resolve HEAD in %s", w)
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
