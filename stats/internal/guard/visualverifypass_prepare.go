package guard

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// prepare is `visual-verify-pass.sh prepare`: the stage's begin mark and
// steps 1–3 of skills/flow/verify-and-handoff.md's **Visual verification**
// and skills/flow/visual-verify.md, in that order, each failing check
// ending the stage before any verifier row begins.
func (v *vvp) prepare() {
	v.kickoff()
	v.markBegin()
	survivors := v.trigger()
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
	v.die("prepare's steps 5 and 6 and the verifier rows are not implemented")
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
