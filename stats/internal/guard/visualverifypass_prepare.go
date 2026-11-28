package guard

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/records"
)

// prepare is `visual-verify-pass.sh prepare`: the stage's begin mark and
// steps 1–3 of skills/flow/verify-and-handoff.md's **Visual verification**
// and skills/flow/visual-verify.md, in that order, each failing check
// ending the stage before any verifier row begins.
func (v *vvp) prepare() error {
	var survivors []string
	var err error
	if v.cont {
		if survivors, err = v.continueSet(); err != nil {
			return err
		}
	} else {
		// The `## visual verification` section is validated before any capture.
		var ok bytes.Buffer
		if rc := checkVisualVerification([]string{v.canon}, v.env, &ok, v.stderr); rc != 0 {
			v.stdout.Write(ok.Bytes())
			return guardExit(rc)
		}
		// A change decided `skipped` verifies nothing, so it needs no worktree to serve a stack.
		if _, skip := v.decidedSkip(); !skip {
			if err := v.kickoff(); err != nil {
				return err
			}
		}
		if survivors, err = v.trigger(); err != nil {
			return err
		}
		if err := v.writeStage(survivors); err != nil {
			return err
		}
		if reason, ok := v.decidedSkip(); ok {
			for range survivors {
				v.say("Visual: skipped at Decide — %s", reason)
			}
			return v.next("skip — skipped at Decide — " + reason)
		}
		if len(survivors) == 0 {
			return v.next("skip — no worktree's diff matched `ui paths`")
		}
		if err := v.roundDiff(survivors); err != nil {
			return err
		}
		if err := v.preflight(survivors); err != nil {
			return err
		}
	}
	stack, err := v.probe(survivors)
	if err != nil {
		return err
	}
	return v.dispatch(survivors, stack)
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
func (v *vvp) kickoff() error {
	roots := tcfAppRoots(core.ProjectSection(filepath.Join(v.canon, ".flow", "project.md"), "apps"))
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
			named := v.linkRoots(strings.TrimPrefix(branch, "spectre/"))
			for _, r := range roots {
				if core.IsDir(r) && v.inSet(r) == "" && (named == nil || named[realPath(r)]) {
					add(r)
				}
			}
		}
	}
	if len(missing) > 0 {
		return v.stop("kickoff-missing", "", "kickoff "+strings.Join(missing, " "))
	}
	return nil
}

// linkRoots is the real main checkout of every repository the change's
// link.md names (backticked, in its `## Merge order`), resolved through
// spectre/peers; nil when there is no link.md or peers file, so every `## apps`
// repository is then required.
func (v *vvp) linkRoots(name string) map[string]bool {
	link, err := os.ReadFile(filepath.Join(v.canon, "spectre", "changes", name, "link.md"))
	if err != nil {
		return nil
	}
	peers, err := os.ReadFile(filepath.Join(v.canon, "spectre", "peers"))
	if err != nil {
		return nil
	}
	main := v.repoRoot(v.canon)
	named := map[string]bool{main: true}
	for _, l := range strings.Split(string(peers), "\n") {
		f := strings.Fields(l)
		if len(f) == 2 && strings.Contains(string(link), "`"+f[0]+"`") {
			p := f[1]
			if !filepath.IsAbs(p) {
				p = filepath.Join(main, p)
			}
			named[realPath(p)] = true
		}
	}
	return named
}

// changed is a worktree's `git diff --name-only <base>...HEAD`.
func (v *vvp) changed(w, base, dots string) ([]string, error) {
	out, ok := v.git("-C", w, "diff", "--no-renames", "--name-only", base+dots+"HEAD")
	if !ok {
		return nil, v.die("git-diff-failed", "git diff %s%sHEAD failed in %s", base, dots, w)
	}
	if out == "" {
		return nil, nil
	}
	return strings.Split(out, "\n"), nil
}

// trigger is steps 1 and 2: the worktrees whose diff — from --since when
// passed, else --merge-base — matched `ui paths`. Every other worktree's
// diff matched against the canonical worktree's `ui paths` under `--repo`
// makes the canonical worktree survive too.
func (v *vvp) trigger() ([]string, error) {
	match := map[string]bool{}
	v.siblings = map[string][]string{}
	for _, w := range v.worktrees {
		base := v.since[w]
		if base == "" {
			base = v.mergeBase[w]
		}
		diff, err := v.changed(w, base, "...")
		if err != nil {
			return nil, err
		}
		names := strings.Join(diff, "\n") + "\n"
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
	return survivors, nil
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
func (v *vvp) roundDiff(survivors []string) error {
	some, onlyMockups, empty := false, true, true
	for _, sv := range survivors {
		// A survivor stands for its siblings too: a sibling's diff is what made it survive.
		for _, w := range append([]string{sv}, v.siblings[sv]...) {
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
			diff, err := v.changed(w, rb, "..")
			if err != nil {
				return err
			}
			for _, p := range diff {
				empty = false
				if mock == "" || !strings.HasPrefix(p, mock) {
					onlyMockups = false
				}
			}
		}
	}
	switch {
	case !some:
	case empty:
		return v.next("skip — round diff empty")
	case onlyMockups:
		return v.next("compose-only")
	}
	return nil
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
// declared, check-sidecar-reach -- on each existing `<spec>.mockups` of the
// specs the `specs` row prints, run from the capture root, or on the whole
// capture root when no `specs` row is declared; a failing `specs` command
// is a failing check (kan-1086 self-review, F11). With `design repo` declared
// too, render-frames runs first on design.md and those sidecars, so the
// tagged frames they cite exist before check-sidecar-reach, compose and
// check-frame-list read them (kan-1096). Of the sidecars reach-checked, only
// the change's own (own) go on to the bundle's `own specs:` line; a --defects pass keeps every sidecar holding a defect frame. Every failing check is collected, then one
// `STOP:` line names them all.
func (v *vvp) preflight(survivors []string) error {
	var fails, sidecars []string
	collect := func(guard, out, errs string) {
		ev := strings.TrimSpace(out + "\n" + errs)
		fails = append(fails, guard+": "+strings.Join(strings.Fields(strings.ReplaceAll(ev, "\n", " | ")), " "))
	}
	others := []string{v.canon, v.name, v.token}
	others = append(others, v.worktrees[1:]...)
	switch rc, out, errs := v.guard("check-verify-green", nil, others...); rc {
	case 0:
	case 1:
		return v.stop("verify-not-green", "stopped", "verify-not-green — "+out)
	default:
		collect("check-verify-green", out, errs)
	}
	for _, w := range survivors {
		if rc, kv, errs := v.guard("prepare-workspace", nil, w); rc != 0 {
			collect("prepare-workspace", kv, errs)
		} else if v.declaredApps(w, kv); len(v.apps[w]) == 0 {
			collect("check-visual-preflight", "", "no --app passed for "+w)
		} else if rc, out, errs := v.guard("check-visual-preflight", strings.NewReader(kv+"\n"), append([]string{w}, v.apps[w]...)...); rc != 0 {
			collect("check-visual-preflight", out, errs)
		}
		if cmd := strings.TrimSpace(v.cells(w)["smoke"]); cmd != "" {
			start := time.Now()
			rc, out := core.RunCombined(v.env, v.captureRoot(w), "sh", "-c", cmd)
			if rc != 0 {
				collect("smoke", "", cmd+" exited "+fmt.Sprint(rc)+": "+out)
			} else {
				v.say("SMOKE: %s ok in %.1fs", w, time.Since(start).Seconds())
			}
		}
		if c := v.cells(w)["mockups"]; c != "" {
			root := v.captureRoot(w)
			targets := []string{root}
			if cmd := v.specsCommand(w); cmd != "" {
				rc, out := kwStdout(v.env, root, "sh", "-c", cmd)
				if rc != 0 {
					collect("specs", "", cmd+" exited "+fmt.Sprint(rc)+": "+out)
					continue
				}
				targets = nil
				for _, spec := range strings.Fields(out) {
					if sc := cellPath(root, spec) + ".mockups"; core.IsFile(sc) {
						targets = append(targets, sc)
					}
				}
			}
			files := v.frameFiles(targets)
			own, err := v.own(w)
			if err != nil {
				return err
			}
			if d := v.defects[w]; len(v.roundBase) > 0 && len(d) > 0 {
				own = nil
				// A scoped pass renders and reach-checks only the sidecars holding its defect frames.
				files, targets = nil, nil
				for _, p := range v.frameFiles([]string{root}) {
					body, _ := os.ReadFile(p)
					for _, f := range mockupsLines(body) {
						if slices.Contains(d, f[1]) {
							files, targets = append(files, p), append(targets, p)
							break
						}
					}
				}
			}
			for _, p := range files {
				if strings.HasSuffix(p, ".mockups") && (own == nil || own(p)) {
					sidecars = append(sidecars, p)
				}
			}
			if v.cells(w)["design repo"] != "" && len(files) > 0 {
				if rc, out, errs := v.guard("render-frames", nil, append([]string{w}, files...)...); rc != 0 {
					collect("render-frames", out, errs)
				}
			}
			for _, tgt := range targets {
				if rc, out, errs := v.guard("check-sidecar-reach", nil, tgt, cellPath(w, c)); rc != 0 {
					collect("check-sidecar-reach", out, errs)
				}
			}
		}
	}
	if len(fails) > 0 {
		return v.stop("preflight-failed", "stopped", "preflight — "+strings.Join(fails, "; "))
	}
	v.sidecars = sidecars
	return nil
}

// own is whether a `<spec>.mockups` sidecar is the change's own: it or its
// spec changed in the diff of w's capture root from its --round-base, else
// its --since, else its --merge-base. nil, every sidecar counting, when the
// capture root has no base to diff from.
func (v *vvp) own(w string) (func(string) bool, error) {
	root := v.captureRoot(w)
	base, dots := v.roundBase[root], ".."
	if base == "" {
		base, dots = v.since[root], "..."
	}
	if base == "" {
		base = v.mergeBase[root]
	}
	if base == "" {
		return nil, nil
	}
	changed, err := v.changed(root, base, dots)
	if err != nil {
		return nil, err
	}
	return func(sc string) bool {
		for _, p := range []string{sc, strings.TrimSuffix(sc, ".mockups")} {
			if rel, err := filepath.Rel(root, p); err == nil && slices.Contains(changed, filepath.ToSlash(rel)) {
				return true
			}
		}
		return false
	}, nil
}

// ownSpecs is the bundle's `own specs:` value: the specs, under w's capture
// root, of the sidecars preflight kept (the change's own, by own) and that map
// a frame; `none` when there are none, so an empty `<spec>` never reaches the
// verifier's capture run.
func (v *vvp) ownSpecs(w string) string {
	var specs []string
	root := v.captureRoot(w) + string(filepath.Separator)
	for _, sc := range v.sidecars {
		if body, _ := os.ReadFile(sc); strings.HasPrefix(sc, root) && len(mockupsLines(body)) > 0 {
			specs = append(specs, strings.TrimSuffix(sc, ".mockups"))
		}
	}
	if len(specs) == 0 {
		return "none"
	}
	return strings.Join(specs, " ")
}

// taskCaptured is the specs of w's capture root a task's capture still
// stands for (rcTaskCaptured), over every worktree with a merge base.
func (v *vvp) taskCaptured(w string) []string {
	return rcTaskCaptured(v.env, v.captureRoot(w), v.mergeBase)
}

// frameFiles is the change's design.md, when present, and every
// `<spec>.mockups` sidecar of targets — a sidecar itself, or each one under
// a capture root, `node_modules` and `.git` skipped as check-sidecar-reach
// skips them.
func (v *vvp) frameFiles(targets []string) []string {
	var files []string
	if d := filepath.Join(v.canon, core.SpecRootLeaf(v.canon, io.Discard), "changes", v.name, "design.md"); core.IsFile(d) {
		files = append(files, d)
	}
	for _, tgt := range targets {
		if !core.IsDir(tgt) {
			files = append(files, tgt)
			continue
		}
		_ = filepath.WalkDir(tgt, func(p string, d fs.DirEntry, err error) error {
			if err == nil && d.IsDir() && (d.Name() == "node_modules" || d.Name() == ".git") {
				return filepath.SkipDir
			}
			if err == nil && !d.IsDir() && strings.HasSuffix(p, ".mockups") {
				files = append(files, p)
			}
			return nil
		})
	}
	return files
}

// declaredApps fills a survivor no --app names from its project's
// `preflight apps` row: comma-separated `<app-root>=<url>` cells, `<name>`
// and each `${VAR}` replaced from kv (prepare-workspace's `KEY=value`
// lines), a pair whose root does not exist under w dropped (kan-1086
// self-review, F10). An explicit --app wins.
func (v *vvp) declaredApps(w, kv string) {
	if len(v.apps[w]) > 0 {
		return
	}
	repl := []string{"<name>", v.name}
	for _, l := range strings.Split(kv, "\n") {
		if k, val, ok := strings.Cut(strings.TrimSpace(l), "="); ok {
			repl = append(repl, "${"+k+"}", val)
		}
	}
	r := strings.NewReplacer(repl...)
	for _, cell := range strings.Split(v.cells(w)["preflight apps"], ",") {
		root, url, ok := strings.Cut(r.Replace(strings.Trim(strings.TrimSpace(cell), "`")), "=")
		if ok && root != "" && url != "" && core.IsDir(filepath.Join(w, root)) {
			v.apps[w] = append(v.apps[w], root+"="+url)
		}
	}
}

// specsCommand is w's `specs` row with `<frontend-root>` and `<merge-base>`
// substituted: the frontend the command reads is the sibling whose diff made
// w survive, when one did. "" when no `specs` row is declared.
func (v *vvp) specsCommand(w string) string {
	// The frontend the specs command reads is the sibling whose diff made w survive, when one did.
	root, base := siblingFrontend(w, v.siblings[w], v.mergeBase, func(string) bool { return true })
	return strings.NewReplacer("<frontend-root>", root, "<merge-base>", base).Replace(v.cells(w)["specs"])
}

// continueSet is `--continue`: the survivors, their merge bases and siblings
// and the set's size, read back from the stage file; no stage begin and no
// step-3 check runs again.
func (v *vvp) continueSet() ([]string, error) {
	v.siblings = map[string][]string{}
	var survivors []string
	stage, err := v.readStage()
	if err != nil {
		return nil, err
	}
	for _, l := range stage {
		switch l[0] {
		case "set":
			fmt.Sscan(l[1], &v.setSize)
		case "worktree":
			f := strings.Fields(l[1])
			if len(f) < 4 {
				return nil, v.die("malformed-stage-file-line", "malformed stage file line: worktree %s", l[1])
			}
			survivors = append(survivors, f[0])
			v.mergeBase[f[0]] = f[2]
			if len(f) > 4 {
				for _, pair := range strings.Split(f[4], ",") {
					sib, mb, _ := strings.Cut(pair, "=")
					v.siblings[f[0]] = append(v.siblings[f[0]], sib)
					if mb != "" && v.mergeBase[sib] == "" {
						v.mergeBase[sib] = mb
					}
				}
			}
		}
	}
	if len(survivors) == 0 {
		return nil, v.die("no-surviving-worktree", "the stage file %s names no surviving worktree", v.stagePath())
	}
	return survivors, nil
}

// probe is step 5: each survivor's --app URLs (or its declared `preflight
// apps`, prepare-workspace run first to resolve them), a short GET each, any answer
// counting. A survivor nothing answers for prints `NEXT: start-stack`, or
// under --started stops the stage. The result is the bundle's `stack:` line.
func (v *vvp) probe(survivors []string) (string, error) {
	client := http.Client{Timeout: 3 * time.Second}
	for _, w := range survivors {
		// A survivor with no app to probe is the caller's omission, never a
		// stack to start (kan-1102 self-review, F3).
		if len(v.apps[w]) == 0 {
			if rc, kv, _ := v.guard("prepare-workspace", nil, w); rc == 0 {
				v.declaredApps(w, kv)
			}
		}
		if len(v.apps[w]) == 0 {
			return "", v.usage("no --app passed for " + w)
		}
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
			return "", v.stop("stack-unreachable", "stopped", "stack — "+strings.Join(urls, " ")+" answers nothing after the stack was started, for "+w)
		}
		return "", v.next("start-stack")
	}
	if v.started {
		return "started by this stage", nil
	}
	return "found running", nil
}

func vvpOneLine(s string) string {
	return strings.Join(strings.Fields(strings.ReplaceAll(strings.TrimSpace(s), "\n", " | ")), " ")
}

// vvpRedispatch is a re-dispatch's key, `visual-verify-<n>[-…]`.
var vvpRedispatch = regexp.MustCompile(`^visual-verify-[0-9]+(-|$)`)

// vvpPairRe is a `<model>/flow-<effort>` allowed pair.
var vvpPairRe = regexp.MustCompile(`^([a-z0-9.-]+)/flow-([a-z]+)$`)

// pair is the model and effort every verifier row begins on: the first base
// pair of the `visual-verifier` allowed-pairs row (PolicyTable); the DISPATCH line names the Agent
// call from it (kan-1086 self-review, F9). A failed read or an empty or
// unparsable row stops prepare.
func (v *vvp) pair() (model, effort string, err error) {
	rows, err := core.PolicyTable(v.env)
	if err != nil {
		return "", "", v.stop("policy-unreadable", "stopped", "pair — "+err.Error())
	}
	var m []string
	if rs := rows["visual-verifier"]; len(rs) > 0 && len(rs[0].Pairs) > 0 {
		m = vvpPairRe.FindStringSubmatch(rs[0].Pairs[0])
	}
	if m == nil {
		return "", "", v.stop("visual-pair-missing", "stopped", "pair — the allowed pairs name no `<model>/flow-<effort>` visual-verifier pair")
	}
	return m[1], m[2], nil
}

// dispatch writes each verifier's bundle, begins its row and prints
// `DISPATCH: <key> <bundle> — description "<key>", subagent_type
// flow-<effort>, model <model>, prompt opens "Role: visual-verifier"` per row,
// the pair read by pair(), then `NEXT: dispatch <n>`. A set of
// more than one worktree suffixes each key `-<repository name>`; a
// --continue key already ending in a survivor's suffix runs that worktree
// alone.
func (v *vvp) dispatch(survivors []string, stack string) error {
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
	stage, err := v.readStage()
	if err != nil {
		return err
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
	relay, err := v.relayContract()
	if err != nil {
		return err
	}
	green, err := v.greenRows()
	if err != nil {
		return err
	}
	model, effort, err := v.pair()
	if err != nil {
		return err
	}
	begun := []string{}
	if !v.cont {
		keys := []string{}
		for _, r := range rows {
			keys = append(keys, r.key)
		}
		v.closeSuperseded(keys)
	}
	for _, r := range rows {
		bundle, err := v.bundle(r.key, r.w, stack, relay, green, stage)
		if err != nil {
			return err
		}
		heads := []string{}
		for _, w := range append([]string{r.w}, v.siblings[r.w]...) {
			h, ok := v.git("-C", w, "rev-parse", "HEAD")
			if !ok {
				return v.die("head-unresolved", "cannot resolve HEAD in %s", w)
			}
			heads = append(heads, h)
		}
		argv := []string{"record", "dispatch", "begin", "-C", v.canon, "-change", v.name, "-role", "visual-verifier",
			"-model", model, "-effort", effort, "-key", r.key, "-session-token", v.token, "-diff-base", strings.Join(heads, ",")}
		// The round file's instant is read right before the begin (decision
		// round-file-per-row); a -rest key or an unscoped row has none.
		rf := roundFile(filepath.Join(r.w, ".superpowers", "sdd"), r.key)
		rb := v.roundBase[r.w]
		if rb != "" {
			argv = append(argv, "-round-base", rb)
		}
		if rb != "" && !vrRestSegment.MatchString(r.key) {
			line := rb + " " + time.Now().UTC().Format(time.RFC3339Nano)
			if d := v.defects[r.w]; len(d) > 0 {
				line += " " + strings.Join(d, ",")
			}
			if err := v.writeOut(rf, line+"\n"); err != nil {
				return err
			}
			var bl strings.Builder
			for _, w := range v.worktrees {
				if sha := v.roundBase[w]; sha != "" {
					bl.WriteString(w + " " + sha + "\n")
				}
			}
			if err := v.writeOut(basesFile(filepath.Join(r.w, ".superpowers", "sdd"), r.key), bl.String()); err != nil {
				return err
			}
		} else if err := os.Remove(rf); err != nil && !os.IsNotExist(err) {
			// A stale round file would let close read this row as scoped.
			return v.die("cannot-write", "cannot remove %s: %v", rf, err)
		} else if bf := basesFile(filepath.Join(r.w, ".superpowers", "sdd"), r.key); !vrRestSegment.MatchString(r.key) {
			_ = os.Remove(bf) // a stale bases file would outlive its round file
		}
		// A new row measures into an empty manifest, never an earlier pass's or round's.
		if err := os.Remove(filepath.Join(r.w, ".superpowers", "sdd", "visual-manifest-"+r.key+".json")); err != nil && !os.IsNotExist(err) {
			return v.die("cannot-write", "cannot remove the manifest of %s: %v", r.key, err)
		}
		if out, rc := v.flow(argv...); rc != 0 {
			// End the rows already begun so none is left open in the store.
			for _, b := range begun {
				v.flow("record", "dispatch", "end", "-C", v.canon, "-change", v.name, "-key", b,
					"-session-token", v.token, "-outcome", "aborted")
			}
			return v.stop("dispatch-begin-refused", "stopped", "row — flow record dispatch begin refused for "+r.key+": "+vvpOneLine(out))
		}
		begun = append(begun, r.key)
		if err := v.appendStage("row " + r.key + " " + r.w); err != nil {
			return err
		}
		v.say("DISPATCH: %s %s — description %q, subagent_type flow-%s, model %s, prompt opens \"Role: visual-verifier\"", r.key, bundle, r.key, effort, model)
	}
	return v.next(fmt.Sprintf("dispatch %d", len(rows)))
}

// closeSuperseded ends, aborted, every still-open visual-verifier row of
// this change whose key a new pass (keep) replaces, so none outlives its
// supersession until a later sweep (kan-1160 self-review, F2).
func (v *vvp) closeSuperseded(keep []string) {
	raw, err := core.RecordRead(v.env, v.env.Dispatches, "dispatches", v.name, v.canon)
	var rows []struct{ Key, Role, Outcome string }
	if err == nil {
		err = json.Unmarshal(raw, &rows)
	}
	if err != nil {
		return // the begin below reports an unreachable store
	}
	for _, r := range rows {
		if r.Role == "visual-verifier" && r.Outcome == "" && r.Key != "" && !slices.Contains(keep, r.Key) {
			v.flow("record", "dispatch", "end", "-C", v.canon, "-change", v.name, "-key", r.Key,
				"-session-token", v.token, "-outcome", "aborted")
		}
	}
}

// relayContract is skills/flow/visual-verify.md's paragraph opening
// `**The relay contract.**`, read from the checkout this guard runs from.
func (v *vvp) relayContract() (string, error) {
	root := v.env.Getenv("FLOW_GUARD_REPO_ROOT")
	b, err := os.ReadFile(filepath.Join(root, "skills", "flow", "visual-verify.md"))
	if root == "" || err != nil {
		return "", v.die("cannot-read-visual-verify-md", "cannot read skills/flow/visual-verify.md under FLOW_GUARD_REPO_ROOT '%s' — run flow-guard visual-verify-pass, which sets it", root)
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
		return "", v.die("relay-paragraph-missing", "skills/flow/visual-verify.md carries no paragraph opening **The relay contract.**")
	}
	return strings.Join(para, "\n"), nil
}

// greenRows is this session's completed or skipped inline verify rows with
// a recorded HEAD, latest last, as `flow record dispatches` returns them.
func (v *vvp) greenRows() ([]struct{ Key, DiffBase string }, error) {
	raw, err := core.RecordRead(v.env, v.env.Dispatches, "dispatches", v.name, v.canon)
	var rows []struct{ Key, Role, SessionToken, Outcome, DiffBase string }
	if err == nil {
		err = json.Unmarshal(raw, &rows)
	}
	if err != nil {
		return nil, v.die("record-dispatches-failed", "flow record dispatches failed for '%s'", v.name)
	}
	var out []struct{ Key, DiffBase string }
	for _, r := range rows {
		if records.Class(r.Role) == "verifier" && r.SessionToken == v.token && (r.Key == "verify" || strings.HasPrefix(r.Key, "verify-")) &&
			(r.Outcome == "completed" || r.Outcome == "skipped") && r.DiffBase != "" {
			out = append(out, struct{ Key, DiffBase string }{r.Key, r.DiffBase})
		}
	}
	return out, nil
}

// bundle writes <w>/.superpowers/sdd/verify-prompt-<key>.md, every line of
// the verifier's prompt a script can resolve (skills/flow/visual-verify.md,
// **Steps 3–13**), and returns its path.
func (v *vvp) bundle(key, w, stack, relay string, green []struct{ Key, DiffBase string }, stage [][2]string) (string, error) {
	sdd := filepath.Join(w, ".superpowers", "sdd")
	report := filepath.Join(sdd, "verify-report-"+key+".md")
	changeRoot := filepath.Join(v.canon, core.SpecRootLeaf(v.canon, io.Discard), "changes", v.name)
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
		return "", v.die("prepare-workspace-failed", "prepare-workspace %s: %s", w, errs)
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
	line("specs: %s", vvpOr(v.specsCommand(w)))
	line("screenshots: %s", vvpOr(cellPath(v.captureRoot(w), cells["screenshots"])))
	line("workers: %d", captureWorkers(runtime.NumCPU()))
	if v.roundBase[w] != "" {
		if c := rcCaptured(v.env, v.captureRoot(w), v.roundBase); len(c) > 0 {
			line("captured: %s", strings.Join(c, " "))
		}
	} else if c := v.taskCaptured(w); len(c) > 0 {
		line("captured: %s", strings.Join(c, " "))
	}
	if m := cells["mockups"]; m != "" {
		line("own specs: %s", v.ownSpecs(w))
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
	line("%s", vvpOr(core.ProjectSection(filepath.Join(w, ".flow", "project.md"), "run")))
	line("")
	line("## Stack")
	line("")
	line("stack: %s", stack)
	verify := "none recorded"
	for _, g := range green {
		if vvpResolves(v, w, g.DiffBase) {
			verify = g.Key + "@" + vgShort(g.DiffBase)
		}
	}
	line("verify: %s", verify)
	line("")
	line("## Measure")
	line("")
	line("measure: flow-guard visual-verify-pass measure %s %s %s <spec path>", w, v.name, vrRestSegment.ReplaceAllString(key, ""))
	line("")
	line("## Report checks")
	line("")
	design := filepath.Join(changeRoot, "design.md")
	mockups := ""
	if m := cells["mockups"]; m != "" {
		mockups = " --mockups " + cellPath(w, m)
	}
	line("check: flow-guard visual-verify-pass close --check-only %s <motions named>%s --design %s", report, mockups, design)
	line("")
	line("## Mockup sources")
	line("")
	sources, decoded, err := v.mockupSources(key, w, changeRoot)
	if err != nil {
		return "", err
	}
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
		return "", v.die("cannot-write", "cannot write %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", v.die("cannot-write", "cannot write %s: %v", path, err)
	}
	return path, nil
}

func vvpOr(s string) string {
	if strings.TrimSpace(s) == "" {
		return "not declared"
	}
	return s
}

func vvpResolves(v *vvp, w, sha string) bool {
	return core.EnvGit(v.env)("-C", w, "cat-file", "-e", sha+"^{commit}").Run() == nil
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
func (v *vvp) mockupSources(key, w, changeRoot string) ([]string, []string, error) {
	body, err := os.ReadFile(filepath.Join(changeRoot, "design.md"))
	if err != nil {
		return nil, nil, nil
	}
	var sources, decoded []string
	seen := map[string]bool{}
	for _, m := range vrMockupSource.FindAllStringSubmatch(string(body), -1) {
		p := filepath.Join(v.canon, m[1])
		if seen[m[1]] || !core.IsFile(p) {
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
			return nil, nil, v.die("cannot-decode-the-bundled-page", "cannot decode the bundled page %s: %v", p, err)
		}
		decoded = append(decoded, m[1]+" → "+dir)
	}
	return sources, decoded, nil
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

// siblingFrontend is the frontend a `specs` row reads and its merge base:
// the first of siblings, in order, that matches admits, with its base, else
// canon with its own -- the one rule `visual-verify-pass prepare` and
// `run-verify-list` share (decision shared-sibling-frontend-helper).
func siblingFrontend(canon string, siblings []string, bases map[string]string, matches func(string) bool) (root, base string) {
	for _, s := range siblings {
		if matches(s) {
			return s, bases[s]
		}
	}
	return canon, bases[canon]
}
