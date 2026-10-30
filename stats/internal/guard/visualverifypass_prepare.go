package guard

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// prepare is `visual-verify-pass.sh prepare`: the stage's begin mark and
// steps 1–3 of skills/flow/verify-and-handoff.md's **Visual verification**
// and skills/flow/visual-verify.md, in that order, each failing check
// ending the stage before any verifier row begins.
func (v *vvp) prepare() {
	var survivors []string
	if v.cont {
		v.stageOpen = true
		survivors = v.continueSet()
	} else {
		v.kickoff()
		v.markBegin()
		survivors = v.trigger()
		v.writeStage(survivors)
		if reason, ok := v.decidedSkip(); ok {
			for range survivors {
				v.say("Visual: skipped at Decide — %s", reason)
			}
			v.markEnd("skipped")
			v.next("skip — skipped at Decide — " + reason)
		}
		if len(survivors) == 0 {
			v.markEnd("completed")
			v.next("skip — no worktree's diff matched `ui paths`")
		}
		v.roundDiff(survivors)
		v.preflight(survivors)
		if v.miss {
			v.next("analyst")
		}
	}
	stack := v.probe(survivors)
	fingerprints := v.fingerprint(survivors)
	v.dispatch(survivors, stack, fingerprints)
}

// realPath is p with symlinks resolved, or cleaned when it cannot be.
func realPath(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		return r
	}
	return filepath.Clean(p)
}

// repoRoot is the real main checkout a worktree belongs to.
func (v *vvp) repoRoot(w string) string {
	if common, ok := v.git("-C", w, "rev-parse", "--path-format=absolute", "--git-common-dir"); ok {
		return realPath(filepath.Dir(common))
	}
	return realPath(w)
}

// inSet is the set's worktree in the repository whose main checkout is root,
// or "".
func (v *vvp) inSet(root string) string {
	for _, w := range v.worktrees {
		if v.repoRoot(w) == realPath(root) {
			return w
		}
	}
	return ""
}

// kickoff is step 3's first rule: a declared `regression checkout`, or with
// a recorded flowBase every `## apps` repository, the set holds no worktree
// in prints `STOP: kickoff <repo>` before anything is begun.
func (v *vvp) kickoff() {
	roots := tcfAppRoots(projectSection(filepath.Join(v.canon, ".flow", "project.md"), "apps"))
	declared := func(p string) bool {
		for _, r := range roots {
			if realPath(r) == realPath(p) {
				return true
			}
		}
		return false
	}
	var missing []string
	add := func(root string) {
		name := filepath.Base(realPath(root))
		for _, m := range missing {
			if m == name {
				return
			}
		}
		missing = append(missing, name)
	}
	if c := cellPath(v.canon, v.cells(v.canon)["regression checkout"]); c != "" && declared(c) && v.inSet(c) == "" {
		add(c)
	}
	if branch, _ := v.git("-C", v.canon, "symbolic-ref", "--short", "HEAD"); strings.HasPrefix(branch, "spectre/") {
		if fb, _ := v.git("-C", v.canon, "config", "--get", "branch."+branch+".flowBase"); fb != "" {
			for _, r := range roots {
				if isDir(r) && v.inSet(r) == "" {
					add(r)
				}
			}
		}
	}
	if len(missing) > 0 {
		v.stop("", "kickoff "+strings.Join(missing, " "))
	}
}

// changed is a worktree's `git diff --name-only <base>...HEAD`.
func (v *vvp) changed(w, base, dots string) []string {
	out, ok := v.git("-C", w, "diff", "--no-renames", "--name-only", base+dots+"HEAD")
	if !ok {
		v.die("git diff %s%sHEAD failed in %s", base, dots, w)
	}
	if out == "" {
		return nil
	}
	return strings.Split(out, "\n")
}

// trigger is steps 1 and 2: the worktrees whose diff — from --since when
// passed, else --merge-base — matched `ui paths`. Every other worktree's
// diff matched against the canonical worktree's `ui paths` under `--repo`
// makes the canonical worktree survive too.
func (v *vvp) trigger() []string {
	match := map[string]bool{}
	v.siblings = map[string][]string{}
	for _, w := range v.worktrees {
		base := v.since[w]
		if base == "" {
			base = v.mergeBase[w]
		}
		names := strings.Join(v.changed(w, base, "..."), "\n") + "\n"
		if rc, _, _ := v.guard("check-visual-trigger", strings.NewReader(names), w); rc == 0 {
			match[w] = true
		}
		if w != v.canon {
			if rc, _, _ := v.guard("check-visual-trigger", strings.NewReader(names), "--repo", v.repoName(w), v.canon); rc == 0 {
				match[v.canon] = true
				v.siblings[v.canon] = append(v.siblings[v.canon], w)
			}
		}
	}
	var survivors []string
	for _, w := range v.worktrees {
		if match[w] {
			survivors = append(survivors, w)
		}
	}
	return survivors
}

// decidedSkip is the decision's `visual.verify` `skipped` and its reason.
func (v *vvp) decidedSkip() (string, bool) {
	raw, err := os.ReadFile(filepath.Join(v.canon, ".superpowers", "sdd", "decision.json"))
	if err != nil {
		return "", false
	}
	var d struct {
		Visual struct{ Verify, Reason string } `json:"visual"`
	}
	if json.Unmarshal(raw, &d) != nil || d.Visual.Verify != "skipped" {
		return "", false
	}
	return d.Visual.Reason, true
}

// roundDiff applies the --round-base diffs: all empty skips the stage, and
// one touching only each worktree's `mockups` root composes only.
func (v *vvp) roundDiff(survivors []string) {
	some, onlyMockups, empty := false, true, true
	for _, w := range survivors {
		rb := v.roundBase[w]
		if rb == "" {
			continue
		}
		some = true
		mock := ""
		if c := v.cells(w)["mockups"]; c != "" {
			if rel, err := filepath.Rel(w, cellPath(w, c)); err == nil {
				mock = filepath.ToSlash(rel) + "/"
			}
		}
		for _, p := range v.changed(w, rb, "..") {
			empty = false
			if mock == "" || !strings.HasPrefix(p, mock) {
				onlyMockups = false
			}
		}
	}
	switch {
	case !some:
	case empty:
		v.markEnd("completed")
		v.next("skip — round diff empty")
	case onlyMockups:
		v.next("compose-only")
	}
}

// captureRoot is where a worktree's capture specs live: the set's worktree
// of a declared `regression checkout`, else that checkout, else w.
func (v *vvp) captureRoot(w string) string {
	c := cellPath(w, v.cells(w)["regression checkout"])
	if c == "" {
		return w
	}
	if s := v.inSet(c); s != "" {
		return s
	}
	return c
}

// preflight is step 3's checks: check-verify-green once, then per survivor
// prepare-workspace into check-visual-preflight and, with `mockups`
// declared, check-sidecar-reach. Every failing check is collected, then one
// `STOP:` line names them all.
func (v *vvp) preflight(survivors []string) {
	var fails []string
	fail := func(guard, out, errs string) {
		ev := strings.TrimSpace(out + "\n" + errs)
		fails = append(fails, guard+": "+strings.Join(strings.Fields(strings.ReplaceAll(ev, "\n", " | ")), " "))
	}
	others := []string{v.canon, v.name, v.token}
	others = append(others, v.worktrees[1:]...)
	switch rc, out, errs := v.guard("check-verify-green", nil, others...); rc {
	case 0:
	case 1:
		v.stop("stopped", "verify-not-green — "+out)
	default:
		fail("check-verify-green", out, errs)
	}
	for _, w := range survivors {
		if len(v.apps[w]) == 0 {
			fail("check-visual-preflight", "", "no --app passed for "+w)
		} else if rc, kv, errs := v.guard("prepare-workspace", nil, w); rc != 0 {
			fail("prepare-workspace", kv, errs)
		} else if rc, out, errs := v.guard("check-visual-preflight", strings.NewReader(kv+"\n"), append([]string{w}, v.apps[w]...)...); rc != 0 {
			fail("check-visual-preflight", out, errs)
		}
		if c := v.cells(w)["mockups"]; c != "" {
			if rc, out, errs := v.guard("check-sidecar-reach", nil, v.captureRoot(w), cellPath(w, c)); rc != 0 {
				fail("check-sidecar-reach", out, errs)
			}
		}
	}
	if len(fails) > 0 {
		v.stop("stopped", "preflight — "+strings.Join(fails, "; "))
	}
}

// continueSet is `--continue`: the survivors, their merge bases and siblings
// and the set's size, read back from the stage file; no stage begin and no
// step-3 check runs again.
func (v *vvp) continueSet() []string {
	v.siblings = map[string][]string{}
	var survivors []string
	for _, l := range v.readStage() {
		switch l[0] {
		case "set":
			fmt.Sscan(l[1], &v.setSize)
		case "worktree":
			f := strings.Fields(l[1])
			if len(f) < 4 {
				v.die("malformed stage file line: worktree %s", l[1])
			}
			survivors = append(survivors, f[0])
			v.mergeBase[f[0]] = f[2]
			if len(f) > 4 {
				v.siblings[f[0]] = strings.Split(f[4], ",")
			}
		}
	}
	if len(survivors) == 0 {
		v.die("the stage file %s names no surviving worktree", v.stagePath())
	}
	return survivors
}

// probe is step 5: each survivor's --app URLs, a short GET each, any answer
// counting. A survivor nothing answers for prints `NEXT: start-stack`, or
// under --started stops the stage. The result is the bundle's `stack:` line.
func (v *vvp) probe(survivors []string) string {
	client := http.Client{Timeout: 3 * time.Second}
	for _, w := range survivors {
		answered := false
		var urls []string
		for _, a := range v.apps[w] {
			_, u, _ := strings.Cut(a, "=")
			urls = append(urls, u)
			if resp, err := client.Get(u); err == nil {
				resp.Body.Close()
				answered = true
				break
			}
		}
		if answered {
			continue
		}
		if v.started {
			v.stop("stopped", "stack — "+strings.Join(urls, " ")+" answers nothing after the stack was started, for "+w)
		}
		v.next("start-stack")
	}
	if v.started {
		return "started by this stage"
	}
	return "found running"
}

// fingerprint is step 6 per survivor through check-dev-stack-fresh: exit 1
// prints `NEXT: restart-stack`, or under --started stops the stage; exit 2
// reads `not declared — <its stderr reason>`.
func (v *vvp) fingerprint(survivors []string) map[string]string {
	got := map[string]string{}
	for _, w := range survivors {
		rc, out, errs := v.guard("check-dev-stack-fresh", nil, w)
		switch rc {
		case 0:
			got[w] = "exit 0 — the served bundle is the worktree's build"
		case 1:
			if v.started {
				v.stop("stopped", "fingerprint — "+vvpOneLine(out+"\n"+errs))
			}
			v.next("restart-stack")
		default:
			got[w] = "not declared — " + strings.TrimPrefix(vvpLastLine(errs), "check-dev-stack-fresh: ")
		}
	}
	return got
}

func vvpOneLine(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(strings.TrimSpace(s), "\n", " | ")), " ")
}

func vvpLastLine(s string) string {
	ls := strings.Split(strings.TrimSpace(s), "\n")
	return ls[len(ls)-1]
}

// vvpRedispatch is a re-dispatch's key, `visual-verify-<n>[-…]`.
var vvpRedispatch = regexp.MustCompile(`^visual-verify-[0-9]+(-|$)`)

// dispatch writes each verifier's bundle, begins its row and prints
// `DISPATCH: <key> <bundle>` per row, then `NEXT: dispatch <n>`. A set of
// more than one worktree suffixes each key `-<repository name>`; a
// --continue key already ending in a survivor's suffix runs that worktree
// alone.
func (v *vvp) dispatch(survivors []string, stack string, fingerprints map[string]string) {
	size := v.setSize
	if !v.cont {
		size = len(v.worktrees)
	}
	type row struct{ key, w string }
	var rows []row
	for _, w := range survivors {
		if size > 1 && strings.HasSuffix(v.key, "-"+v.repoName(w)) {
			rows = []row{{v.key, w}}
			break
		}
	}
	if rows == nil {
		for _, w := range survivors {
			k := v.key
			if size > 1 {
				k += "-" + v.repoName(w)
			}
			rows = append(rows, row{k, w})
		}
	}
	relay := v.relayContract()
	green := v.greenRows()
	var stage [][2]string
	if v.cont {
		stage = v.readStage()
	}
	for _, r := range rows {
		bundle := v.bundle(r.key, r.w, stack, fingerprints[r.w], relay, green, stage)
		heads := []string{}
		for _, w := range append([]string{r.w}, v.siblings[r.w]...) {
			h, ok := v.git("-C", w, "rev-parse", "HEAD")
			if !ok {
				v.die("cannot resolve HEAD in %s", w)
			}
			heads = append(heads, h)
		}
		argv := []string{"record", "dispatch", "begin", "-C", v.canon, "-change", v.name, "-role", "verifier",
			"-model", "opus", "-effort", "low", "-key", r.key, "-session-token", v.token, "-diff-base", strings.Join(heads, ",")}
		if rb := v.roundBase[r.w]; rb != "" {
			argv = append(argv, "-round-base", rb)
		}
		if out, rc := v.flow(argv...); rc != 0 {
			v.stop("stopped", "row — flow record dispatch begin refused for "+r.key+": "+vvpOneLine(out))
		}
		v.appendStage("row " + r.key + " " + r.w)
		v.say("DISPATCH: %s %s", r.key, bundle)
	}
	v.next(fmt.Sprintf("dispatch %d", len(rows)))
}

// relayContract is skills/flow/visual-verify.md's paragraph opening
// `**The relay contract.**`, read from the checkout this guard runs from.
func (v *vvp) relayContract() string {
	root := v.env.Getenv("FLOW_GUARD_REPO_ROOT")
	b, err := os.ReadFile(filepath.Join(root, "skills", "flow", "visual-verify.md"))
	if root == "" || err != nil {
		v.die("cannot read skills/flow/visual-verify.md under FLOW_GUARD_REPO_ROOT '%s' — run scripts/visual-verify-pass.sh, which sets it", root)
	}
	var para []string
	for _, l := range strings.Split(string(b), "\n") {
		if len(para) == 0 && !strings.HasPrefix(l, "**The relay contract.**") {
			continue
		}
		if strings.TrimSpace(l) == "" {
			break
		}
		para = append(para, l)
	}
	if len(para) == 0 {
		v.die("skills/flow/visual-verify.md carries no paragraph opening **The relay contract.**")
	}
	return strings.Join(para, "\n")
}

// greenRows is this session's completed or skipped inline verify rows with
// a recorded HEAD, latest last, as `flow record dispatches` returns them.
func (v *vvp) greenRows() []struct{ Key, DiffBase string } {
	raw, err := pfdRead(v.env, v.env.Dispatches, "dispatches", v.name, v.canon)
	var rows []struct{ Key, Role, SessionToken, Outcome, DiffBase string }
	if err == nil {
		err = json.Unmarshal(raw, &rows)
	}
	if err != nil {
		v.die("flow record dispatches failed for '%s'", v.name)
	}
	var out []struct{ Key, DiffBase string }
	for _, r := range rows {
		if r.Role == "verifier" && r.SessionToken == v.token && (r.Key == "verify" || strings.HasPrefix(r.Key, "verify-")) &&
			(r.Outcome == "completed" || r.Outcome == "skipped") && r.DiffBase != "" {
			out = append(out, struct{ Key, DiffBase string }{r.Key, r.DiffBase})
		}
	}
	return out
}

// bundle writes <w>/.superpowers/sdd/verify-prompt-<key>.md, every line of
// the verifier's prompt a script can resolve (skills/flow/visual-verify.md,
// **Steps 3–13**), and returns its path.
func (v *vvp) bundle(key, w, stack, fp, relay string, green []struct{ Key, DiffBase string }, stage [][2]string) string {
	sdd := filepath.Join(w, ".superpowers", "sdd")
	report := filepath.Join(sdd, "verify-report-"+key+".md")
	changeRoot := filepath.Join(v.canon, specRootLeaf(v.canon, io.Discard), "changes", v.name)
	cells := v.cells(w)
	var b strings.Builder
	line := func(format string, a ...any) { fmt.Fprintf(&b, format+"\n", a...) }
	line("# Visual verifier bundle — %s", key)
	line("")
	line("worktree: %s", w)
	line("changeRoot: %s", changeRoot)
	line("verifier steps: %s — run its steps 4 and 8–11 and its motion step as written",
		filepath.Join(v.env.Getenv("FLOW_GUARD_REPO_ROOT"), "skills", "flow", "visual-verify-verifier.md"))
	line("")
	line("## Workspace")
	line("")
	if rc, kv, errs := v.guard("prepare-workspace", nil, w); rc != 0 {
		v.die("prepare-workspace %s: %s", w, errs)
	} else if kv == "" {
		line("(no `## workspace isolation` value exported)")
	} else {
		line("%s", kv)
	}
	line("")
	line("## Commands")
	line("")
	for _, c := range []string{"setup", "capture"} {
		line("%s: %s", c, vvpOr(cells[c]))
	}
	specs := strings.NewReplacer("<frontend-root>", w, "<merge-base>", v.mergeBase[w]).Replace(cells["specs"])
	line("specs: %s", vvpOr(specs))
	line("screenshots: %s", vvpOr(cellPath(v.captureRoot(w), cells["screenshots"])))
	if m := cells["mockups"]; m != "" {
		line("mockups: %s", cellPath(w, m))
	}
	if m := cells["mockup frame"]; m != "" {
		line("mockup frame: %s", m)
	}
	line("")
	line("## Apps")
	line("")
	for _, a := range v.apps[w] {
		_, u, _ := strings.Cut(a, "=")
		line("app: %s", u)
	}
	line("")
	line("## Run")
	line("")
	line("%s", vvpOr(projectSection(filepath.Join(w, ".flow", "project.md"), "run")))
	line("")
	line("## Stack")
	line("")
	line("stack: %s", stack)
	line("fingerprint: %s", fp)
	verify := "none recorded"
	for _, g := range green {
		if vvpResolves(v, w, g.DiffBase) {
			verify = g.Key + "@" + vgShort(g.DiffBase)
		}
	}
	line("verify: %s", verify)
	line("")
	line("## Mockup sources")
	line("")
	sources, decoded := v.mockupSources(key, w, changeRoot)
	if len(sources) == 0 {
		line("mockup sources: none")
	} else {
		line("mockup sources: %s", strings.Join(sources, ", "))
	}
	for _, d := range decoded {
		line("mockup decoded: %s", d)
	}
	line("")
	line("## Relay contract")
	line("")
	line("%s", strings.ReplaceAll(relay, "<abs-worktree>/.superpowers/sdd/verify-report-<key>.md", report))
	if v.cont {
		v.attempts(&b, key, w, stage)
	}
	path := filepath.Join(sdd, "verify-prompt-"+key+".md")
	if err := os.MkdirAll(sdd, 0o755); err != nil {
		v.die("cannot write %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		v.die("cannot write %s: %v", path, err)
	}
	return path
}

func vvpOr(s string) string {
	if strings.TrimSpace(s) == "" {
		return "not declared"
	}
	return s
}

func vvpResolves(v *vvp, w, sha string) bool {
	return envGit(v.env)("-C", w, "cat-file", "-e", sha+"^{commit}").Run() == nil
}

// attempts adds a remainder's `## Incomplete attempt` — the incomplete
// report's path and the guard verdict lines `close` recorded for it — or a
// re-dispatch's `## Previous attempt` — every earlier report of w.
func (v *vvp) attempts(b *strings.Builder, key, w string, stage [][2]string) {
	var earlier []string
	for _, l := range stage {
		if l[0] != "row" {
			continue
		}
		k, rw, _ := strings.Cut(l[1], " ")
		if rw == w && k != key {
			earlier = append(earlier, k)
		}
	}
	report := func(k string) string {
		return filepath.Join(w, ".superpowers", "sdd", "verify-report-"+k+".md")
	}
	switch {
	case strings.Contains(key, "-rest") && len(earlier) > 0:
		prev := earlier[len(earlier)-1]
		fmt.Fprintf(b, "\n## Incomplete attempt\n\nreport: %s\n", report(prev))
		for _, l := range stage {
			if k, verdict, _ := strings.Cut(l[1], " "); l[0] == "verdict" && k == prev {
				fmt.Fprintf(b, "%s\n", verdict)
			}
		}
	case vvpRedispatch.MatchString(key) && len(earlier) > 0:
		b.WriteString("\n## Previous attempt\n\n")
		for _, k := range earlier {
			fmt.Fprintf(b, "report: %s\n", report(k))
		}
	}
}

// mockupSources is the absolute path of every HTML or CSS file design.md
// cites that exists in the canonical worktree, and a `<src> → <dir>` line per
// bundled page, each decoded into its own
// <w>/.superpowers/sdd/verify-mockup-<key>/<page path, extension dropped, `/` as `-`>/.
func (v *vvp) mockupSources(key, w, changeRoot string) ([]string, []string) {
	body, err := os.ReadFile(filepath.Join(changeRoot, "design.md"))
	if err != nil {
		return nil, nil
	}
	var sources, decoded []string
	seen := map[string]bool{}
	for _, m := range vrMockupSource.FindAllStringSubmatch(string(body), -1) {
		p := filepath.Join(v.canon, m[1])
		if seen[m[1]] || !isFile(p) {
			continue
		}
		seen[m[1]] = true
		sources = append(sources, p)
		page, err := os.ReadFile(p)
		if err != nil || !bytes.Contains(page, []byte(`<script type="__bundler/template">`)) {
			continue
		}
		dir := filepath.Join(w, ".superpowers", "sdd", "verify-mockup-"+key, strings.ReplaceAll(strings.TrimSuffix(m[1], filepath.Ext(m[1])), "/", "-"))
		if err := vvpDecode(page, dir); err != nil {
			v.die("cannot decode the bundled page %s: %v", p, err)
		}
		decoded = append(decoded, m[1]+" → "+dir)
	}
	return sources, decoded
}

var vvpBundler = regexp.MustCompile(`(?s)<script type="__bundler/(template|manifest)">\s*(.*?)\s*</script>`)

// vvpDecode writes a bundled page's template as dir/index.html and each
// manifest entry beside it, gzip-decompressed when `compressed`.
func vvpDecode(page []byte, dir string) error {
	parts := map[string][]byte{}
	for _, m := range vvpBundler.FindAllSubmatch(page, -1) {
		parts[string(m[1])] = m[2]
	}
	var tmpl string
	if err := json.Unmarshal(parts["template"], &tmpl); err != nil {
		return fmt.Errorf("template: %w", err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(tmpl), 0o644); err != nil {
		return err
	}
	if parts["manifest"] == nil {
		return nil
	}
	var manifest map[string]struct {
		Data       string `json:"data"`
		Compressed bool   `json:"compressed"`
	}
	if err := json.Unmarshal(parts["manifest"], &manifest); err != nil {
		return fmt.Errorf("manifest: %w", err)
	}
	for name, e := range manifest {
		if !filepath.IsLocal(name) {
			return fmt.Errorf("manifest entry %q leaves the directory", name)
		}
		data, err := base64.StdEncoding.DecodeString(e.Data)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		if e.Compressed {
			zr, err := gzip.NewReader(bytes.NewReader(data))
			if err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
			if data, err = io.ReadAll(zr); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}
		out := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(out, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}
