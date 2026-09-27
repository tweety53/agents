package guard

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of scripts/test-recover-guard-incident.sh at c5379c0a, one
// subtest per ok: label, plus port: rows pinning what the harness matched only
// loosely — every refusal's stderr line and every dry-run/apply stdout byte
// for byte, every refusal leaving `status --porcelain`, HEAD, REVERT_HEAD and
// the stash list as it found them — and the preconditions the harness never
// reached (no stash, a repo-dir that is a file, --help, a misplaced --apply).
// The harness built throwaway git repositories; here new_repo is built once
// and copied per case, and the guard runs in-process over the copy. Every
// case's repository is its own temp tree: the guard's --apply mutates git.

const (
	rgiP1    = "spectre/changes/kan-423/proposal.md"
	rgiP2    = "spectre/changes/kan-423/tasks.md"
	rgiPlan1 = "incident plan body\n"
	rgiPlan2 = "- restore the planning files\n"
)

// rgiState is what a case observes of its repository, read once before the
// run and once after it, so no parallel subtest runs git on the repo.
type rgiState struct {
	revert                        bool
	porcelain, head, stash, index string
}

type rgiFx struct {
	dir, repo, cwd string // repo is physical, as the harness's pwd -P REPO
	g              fxGit
	reflog         string // `git reflog -g HEAD -n 15` before the run
	before, after  rgiState
}

// out is a git read under the environment the guard's own git runs in, so
// the reflog it captures is the one the guard prints.
func (fx *rgiFx) out(args ...string) (string, bool) {
	cmd := exec.Command(fixtureGit, append([]string{"-C", fx.repo}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	b, err := cmd.Output()
	return string(b), err == nil
}

func (fx *rgiFx) state() rgiState {
	var s rgiState
	_, s.revert = fx.out("rev-parse", "-q", "--verify", "REVERT_HEAD")
	s.porcelain, _ = fx.out("status", "--porcelain", "-uall")
	s.head, _ = fx.out("rev-parse", "HEAD")
	s.stash, _ = fx.out("stash", "list", "--format=%H %gs")
	s.index, _ = fx.out("diff", "--cached")
	return s
}

// stashPlanning is stash_planning_files: P1 and P2 untracked, captured into
// stash@{0}'s third parent and gone from the worktree afterwards.
func (fx *rgiFx) stashPlanning() {
	fx.write(rgiP1, rgiPlan1)
	fx.write(rgiP2, rgiPlan2)
	fx.g.git(fx.repo, "stash", "-q", "-u")
}

// startRevert is start_conflicting_revert: REVERT_HEAD set, failure ignored.
func (fx *rgiFx) startRevert() {
	if fx.g.err != nil {
		return
	}
	cmd := exec.Command(fixtureGit, "-C", fx.repo, "revert", "--no-commit", "HEAD~1")
	cmd.Env = append(os.Environ(), fixtureGitEnv...)
	_ = cmd.Run()
}

func (fx *rgiFx) write(rel, body string) {
	if fx.g.err != nil {
		return
	}
	p := filepath.Join(fx.repo, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		fx.g.err = err
		return
	}
	fx.g.err = os.WriteFile(p, []byte(body), 0o644)
}

func (fx *rgiFx) read(rel string) string {
	b, _ := os.ReadFile(filepath.Join(fx.repo, rel))
	return string(b)
}

func (fx *rgiFx) exists(rel string) bool {
	_, err := os.Stat(filepath.Join(fx.repo, rel))
	return err == nil
}

// porcelain is `git status --porcelain -- <rel>` read off the after-run
// snapshot.
func (fx *rgiFx) porcelain(rel string) string {
	for _, l := range strings.Split(fx.after.porcelain, "\n") {
		if strings.HasSuffix(l, " "+rel) {
			return l
		}
	}
	return ""
}

// plan is the dry-run stdout for files, each marked when in overwrites.
func (fx *rgiFx) plan(files []string, overwrites ...string) string {
	var b strings.Builder
	b.WriteString("reflog diagnosis — the alternating reset pattern the incident showed:\n")
	b.WriteString(fx.reflog)
	b.WriteString("plan:\n  git -C " + fx.repo + " revert --abort\n")
	for _, f := range files {
		for _, o := range overwrites {
			if o == f {
				b.WriteString("  (overwrites an existing untracked file: " + f + ")\n")
			}
		}
		b.WriteString("  mkdir -p " + fx.repo + "/" + filepath.Dir(f) + "\n")
		b.WriteString("  git -C " + fx.repo + ` show "stash@{0}^3:` + f + `" > ` + fx.repo + "/" + f + "\n")
	}
	return b.String()
}

type rgiCheck struct {
	label string
	ok    func(r guardResult, fx *rgiFx) bool
}

func rgiRC(label string, rc int) rgiCheck {
	return rgiCheck{label, func(r guardResult, _ *rgiFx) bool { return r.rc == rc }}
}

func rgiNoStdout(label string) rgiCheck {
	return rgiCheck{label, func(r guardResult, _ *rgiFx) bool { return r.stdout == "" }}
}

func rgiOutHas(label, s string) rgiCheck {
	return rgiCheck{label, func(r guardResult, _ *rgiFx) bool { return strings.Contains(r.stdout, s) }}
}

func rgiErrHas(label, s string) rgiCheck {
	return rgiCheck{label, func(r guardResult, _ *rgiFx) bool { return strings.Contains(r.err, s) }}
}

func rgiRevert(label string, want bool) rgiCheck {
	return rgiCheck{label, func(_ guardResult, fx *rgiFx) bool { return fx.after.revert == want }}
}

// rgiUnchanged is the Review Focus row: a refusal leaves the repository's
// porcelain, HEAD, REVERT_HEAD, index and stash list as it found them.
func rgiUnchanged(label string) rgiCheck {
	return rgiCheck{label, func(_ guardResult, fx *rgiFx) bool { return fx.after == fx.before }}
}

// rgiErrIs pins the whole stderr, built from the case's fixture.
func rgiErrIs(label string, want func(fx *rgiFx) string) rgiCheck {
	return rgiCheck{label, func(r guardResult, fx *rgiFx) bool { return r.err == want(fx) }}
}

func rgiAnchored(label, f string) rgiCheck {
	return rgiCheck{label, func(r guardResult, fx *rgiFx) bool {
		return strings.Contains(r.stdout, "> "+fx.repo+"/"+f)
	}}
}

func rgiCount(out, sub string) int {
	n := 0
	for _, l := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if strings.Contains(l, sub) {
			n++
		}
	}
	return n
}

func TestRecoverGuardIncident(t *testing.T) {
	t.Parallel()
	// new_repo, built once and copied per case: one tracked file three
	// commits touch, so a revert of HEAD~1 leaves REVERT_HEAD set.
	tmpl := t.TempDir() + "/repo"
	var g fxGit
	g.git("", "init", "-q", tmpl)
	for i, body := range []string{"one", "two", "one"} {
		g.write(tmpl+"/shared.txt", body)
		g.git(tmpl, "add", "shared.txt")
		g.git(tmpl, "commit", "-qm", []string{"base", "two", "three"}[i])
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
	const self = "recover-guard-incident: "

	type rgiCase struct {
		fresh  bool                  // a plain directory, no repository
		setup  func(fx *rgiFx)       // nil: new_repo only
		args   func(*rgiFx) []string // nil: {repo}
		checks []rgiCheck
	}
	defaultPlan := func(fx *rgiFx) {
		fx.stashPlanning()
		fx.startRevert()
	}
	repoArg := func(extra ...string) func(*rgiFx) []string {
		return func(fx *rgiFx) []string { return append([]string{fx.repo}, extra...) }
	}
	cases := []rgiCase{
		// Case 1: usage-error-exit-2 — a plain directory that is not a git repo.
		{fresh: true, args: func(fx *rgiFx) []string { return []string{fx.dir} }, checks: []rgiCheck{
			rgiRC("case 1: exits 2 on a non-repo directory", 2),
			rgiNoStdout("case 1: writes nothing to stdout"),
			{"case 1: names the failure on stderr", func(r guardResult, _ *rgiFx) bool { return r.err != "" }},
			{"case 1: stderr carries the cause only, not the exit code", func(r guardResult, _ *rgiFx) bool {
				return !strings.HasSuffix(strings.TrimRight(r.err, "\n"), " 2")
			}},
			rgiErrIs("case 1: port: stderr is the not-a-repository line", func(fx *rgiFx) string {
				return self + "not a git repository: " + fx.repo + "\n"
			})}},
		// Case 1b: unknown-flag-usage-error — an unknown flag after repo-dir is a
		// usage error (exit 2), not a planning path falling through to exit 1.
		{setup: func(fx *rgiFx) { fx.stashPlanning() }, args: repoArg("--bogus"), checks: []rgiCheck{
			rgiRC("case 1b: unknown flag exits 2", 2),
			rgiNoStdout("case 1b: stdout empty on the usage error"),
			rgiErrHas("case 1b: stderr names the offending flag", "--bogus"),
			rgiErrIs("case 1b: port: stderr is the unknown-option line", func(*rgiFx) string {
				return self + "unknown option: --bogus\n"
			}),
			rgiUnchanged("case 1b: port: refusal leaves porcelain, HEAD and stash list unchanged")}},
		// Case 2: no-revert-refusal — no revert in progress, nothing changed.
		{setup: func(fx *rgiFx) { fx.stashPlanning() }, checks: []rgiCheck{
			rgiRC("case 2: exits 1 with no revert in progress", 1),
			rgiNoStdout("case 2: refusal leaves stdout empty"),
			rgiErrHas("case 2: stderr names the missing REVERT_HEAD", "REVERT_HEAD"),
			rgiRevert("case 2: still no revert in progress", false),
			rgiErrIs("case 2: port: stderr is the no-revert line", func(fx *rgiFx) string {
				return self + "no revert in progress in " + fx.repo + " (REVERT_HEAD missing) — nothing to recover\n"
			}),
			rgiUnchanged("case 2: port: refusal leaves porcelain, HEAD and stash list unchanged")}},
		// Case 3: stash-without-third-parent-refusal — the refusal distinguishes
		// the cause, revert still in progress.
		{setup: func(fx *rgiFx) {
			fx.write("shared.txt", "one\nx\n")
			fx.g.git(fx.repo, "stash", "-q") // tracked-only stash: no untracked third parent
			fx.startRevert()
		}, checks: []rgiCheck{
			rgiRC("case 3: exits 1 on a third-parent-less stash", 1),
			rgiNoStdout("case 3: refusal leaves stdout empty"),
			rgiErrHas("case 3: stderr names the missing third parent", "third parent"),
			{"case 3: cause is distinct from the missing-stash message", func(r guardResult, _ *rgiFx) bool {
				return !strings.Contains(r.err, "no stash entry")
			}},
			rgiRevert("case 3: revert still in progress", true),
			rgiErrIs("case 3: port: stderr is the no-third-parent line", func(*rgiFx) string {
				return self + "stash@{0} has no untracked third parent — it was not created with 'git stash -u'; re-stash with -u before any abort\n"
			}),
			rgiUnchanged("case 3: port: refusal leaves porcelain, HEAD and stash list unchanged")}},
		// Port: no stash at all — the precondition between REVERT_HEAD and the
		// third parent, named apart from both.
		{setup: func(fx *rgiFx) { fx.startRevert() }, checks: []rgiCheck{
			rgiRC("port: no stash: exits 1", 1),
			rgiNoStdout("port: no stash: stdout empty"),
			rgiErrIs("port: no stash: stderr is the no-stash line", func(fx *rgiFx) string {
				return self + "no stash entry in " + fx.repo + " — recovery restores stash@{0}^3, which needs a stash\n"
			}),
			rgiUnchanged("port: no stash: refusal leaves porcelain, HEAD and stash list unchanged")}},
		// Case 4: dry-run-changes-nothing — lists the plan, changes nothing.
		{setup: defaultPlan, checks: []rgiCheck{
			rgiRC("case 4: dry-run exits 0", 0),
			rgiOutHas("case 4: plan lists the revert --abort", "revert --abort"),
			rgiOutHas("case 4: plan restores "+rgiP1+" via a git show redirect", `show "stash@{0}^3:`+rgiP1+`"`),
			rgiOutHas("case 4: plan restores "+rgiP2+" via a git show redirect", `show "stash@{0}^3:`+rgiP2+`"`),
			rgiOutHas("case 4: stdout carries the reflog diagnosis block", "reflog diagnosis"),
			rgiRevert("case 4: REVERT_HEAD untouched", true),
			{"case 4: no planning file exists on disk", func(_ guardResult, fx *rgiFx) bool {
				return !fx.exists(rgiP1) && !fx.exists(rgiP2)
			}},
			{"case 4: nothing staged", func(_ guardResult, fx *rgiFx) bool { return fx.after.index == "" }},
			{"case 4: port: stdout is the plan byte for byte, stderr empty", func(r guardResult, fx *rgiFx) bool {
				return r.stdout == fx.plan([]string{rgiP1, rgiP2}) && r.err == ""
			}},
			rgiUnchanged("case 4: port: the dry-run leaves porcelain, HEAD and stash list unchanged")}},
		// Case 5: apply-restores-unstaged-after-abort — aborts first, restores
		// unstaged.
		{setup: defaultPlan, args: func(fx *rgiFx) []string { return []string{"--apply", fx.repo} }, checks: []rgiCheck{
			rgiRC("case 5: apply exits 0", 0),
			// The ordering is normative (design decision redirect-restore-after-
			// single-abort): everything printed before the FIRST restore run-line
			// must carry the abort run-line.
			{"case 5: the abort precedes every restore", func(r guardResult, _ *rgiFx) bool {
				head, _, _ := strings.Cut(r.stdout, "running: git show")
				i := strings.Index(head, "running: git -C")
				return i >= 0 && strings.Contains(head[i:], "revert --abort")
			}},
			rgiRevert("case 5: REVERT_HEAD gone", false),
			{"case 5: " + rgiP1 + " back with stash content", func(_ guardResult, fx *rgiFx) bool {
				return fx.read(rgiP1) == rgiPlan1
			}},
			{"case 5: " + rgiP2 + " back with stash content", func(_ guardResult, fx *rgiFx) bool {
				return fx.read(rgiP2) == rgiPlan2
			}},
			{"case 5: every restored file is untracked", func(_ guardResult, fx *rgiFx) bool {
				return strings.HasPrefix(fx.porcelain(rgiP1), "?? ") && strings.HasPrefix(fx.porcelain(rgiP2), "?? ")
			}},
			{"case 5: nothing staged — the incident's property holds", func(_ guardResult, fx *rgiFx) bool {
				return fx.after.index == ""
			}},
			{"case 5: port: stdout is the plan then the run lines, abort first, byte for byte", func(r guardResult, fx *rgiFx) bool {
				run := func(f string) string {
					return `running: git show "stash@{0}^3:` + f + `" > ` + fx.repo + "/" + f + "\n"
				}
				return r.stdout == fx.plan([]string{rgiP1, rgiP2})+
					"running: git -C "+fx.repo+" revert --abort\n"+run(rgiP1)+run(rgiP2) && r.err == ""
			}},
			{"case 5: port: HEAD and the stash list are what they were", func(_ guardResult, fx *rgiFx) bool {
				return fx.after.head == fx.before.head && fx.after.stash == fx.before.stash
			}}}},
		// Case 6: tracked-target-refusal — a tracked file at a restore target
		// refuses the whole run.
		{setup: func(fx *rgiFx) {
			defaultPlan(fx)
			fx.write(rgiP1, "tracked content\n")
			fx.g.git(fx.repo, "add", rgiP1)
		}, checks: []rgiCheck{
			rgiRC("case 6: exits 1 on a tracked target", 1),
			rgiNoStdout("case 6: refusal leaves stdout empty"),
			rgiErrHas("case 6: stderr names the tracked file", rgiP1),
			rgiRevert("case 6: revert still in progress", true),
			{"case 6: tracked content untouched", func(_ guardResult, fx *rgiFx) bool {
				return fx.read(rgiP1) == "tracked content\n"
			}},
			{"case 6: the staged file is still staged", func(_ guardResult, fx *rgiFx) bool {
				return strings.HasPrefix(fx.porcelain(rgiP1), "A  ")
			}},
			{"case 6: refusal message carries the cause only", func(r guardResult, _ *rgiFx) bool {
				return !strings.HasSuffix(strings.TrimRight(r.err, "\n"), " 1")
			}},
			rgiErrIs("case 6: port: stderr is the refusing line", func(fx *rgiFx) string {
				return self + "refusing: " + rgiP1 + " is tracked in " + fx.repo + " — recovery never clobbers tracked state\n"
			}),
			rgiUnchanged("case 6: port: refusal leaves porcelain, HEAD and stash list unchanged")}},
		// Case 7: subdirectory-invocation-anchors-at-root — repo-dir (explicit or
		// cwd) resolves to the toplevel, so path... is repo-root-relative.
		{setup: func(fx *rgiFx) {
			defaultPlan(fx)
			fx.write("sub/.keep", "")
			fx.cwd = fx.repo + "/sub"
		}, args: func(*rgiFx) []string { return nil }, checks: []rgiCheck{
			rgiRC("case 7: cwd-in-subdirectory dry-run exits 0", 0),
			rgiAnchored("case 7: "+rgiP1+" anchored at the repo root from cwd", rgiP1),
			rgiAnchored("case 7: "+rgiP2+" anchored at the repo root from cwd", rgiP2)}},
		{setup: func(fx *rgiFx) {
			defaultPlan(fx)
			fx.write("sub/.keep", "")
		}, args: func(fx *rgiFx) []string { return []string{fx.repo + "/sub"} }, checks: []rgiCheck{
			rgiRC("case 7: explicit subdirectory repo-dir exits 0", 0),
			rgiAnchored("case 7: "+rgiP1+" anchored at the repo root from explicit dir", rgiP1),
			rgiAnchored("case 7: "+rgiP2+" anchored at the repo root from explicit dir", rgiP2),
			rgiRevert("case 7: REVERT_HEAD untouched", true)}},
		// Port: a relative repo-dir resolves against the caller's cwd, and a
		// symlinked one to its physical toplevel (pwd -P).
		{setup: func(fx *rgiFx) {
			defaultPlan(fx)
			if fx.g.err == nil {
				fx.g.err = os.Symlink(fx.repo, fx.dir+"/link")
			}
		}, args: func(*rgiFx) []string { return []string{"link"} }, checks: []rgiCheck{
			rgiRC("port: relative symlinked repo-dir: exits 0", 0),
			rgiAnchored("port: relative symlinked repo-dir: anchored at the physical root", rgiP1)}},
		// Case 8: overwrite-naming — an existing untracked target is named as an
		// overwrite, an absent sibling is not, and the mark is never inside a
		// command line (the plan stays paste-runnable).
		{setup: func(fx *rgiFx) {
			fx.write("a/notes.md", "aa\n")
			fx.write("za/notes.md", "za\n")
			fx.g.git(fx.repo, "stash", "-q", "-u")
			fx.startRevert()
			fx.write("za/notes.md", "za\n") // back untracked; a/notes.md stays absent
		}, args: repoArg("a", "za"), checks: []rgiCheck{
			rgiRC("case 8: dry-run with an existing untracked target exits 0", 0),
			{"case 8: exactly one overwrite mark (za.md, not absent a.md)", func(r guardResult, _ *rgiFx) bool {
				return rgiCount(r.stdout, "overwrites an existing untracked file") == 1
			}},
			rgiOutHas("case 8: the mark names za/notes.md on its own line", "(overwrites an existing untracked file: za/notes.md)"),
			{"case 8: every mkdir line parses as a command", func(r guardResult, _ *rgiFx) bool {
				for _, l := range strings.Split(r.stdout, "\n") {
					if strings.HasPrefix(l, "  mkdir") && strings.Contains(l, "(") {
						return false
					}
				}
				return true
			}},
			{"case 8: port: stdout is the plan byte for byte, overwrite line before its mkdir", func(r guardResult, fx *rgiFx) bool {
				return r.stdout == fx.plan([]string{"a/notes.md", "za/notes.md"}, "za/notes.md")
			}},
			{"case 8: port: the dry-run leaves the untracked target as it was", func(_ guardResult, fx *rgiFx) bool {
				return fx.read("za/notes.md") == "za\n" && !fx.exists("a/notes.md")
			}}}},
		// Case 9: HEAD-tracked-target-refusal — a target tracked in HEAD only
		// (committed, then removed from the index) refuses the whole run.
		{setup: func(fx *rgiFx) {
			fx.write(rgiP1, "committed\n")
			fx.g.git(fx.repo, "add", rgiP1)
			fx.g.git(fx.repo, "commit", "-qam", "sneak")
			fx.g.git(fx.repo, "rm", "-q", "--cached", rgiP1)
			fx.stashPlanning()
			// `stash -u` restores the index to HEAD, re-tracking P1 — remove it
			// from the index again so only the cat-file HEAD-tracked half of
			// the guard can fire.
			fx.g.git(fx.repo, "rm", "-q", "--cached", rgiP1)
			fx.startRevert()
		}, checks: []rgiCheck{
			rgiRC("case 9: exits 1 on a HEAD-tracked target", 1),
			rgiNoStdout("case 9: refusal leaves stdout empty"),
			rgiErrHas("case 9: stderr names the HEAD-tracked file", rgiP1),
			rgiRevert("case 9: revert still in progress", true),
			rgiErrIs("case 9: port: stderr is the refusing line", func(fx *rgiFx) string {
				return self + "refusing: " + rgiP1 + " is tracked in " + fx.repo + " — recovery never clobbers tracked state\n"
			}),
			rgiUnchanged("case 9: port: refusal leaves porcelain, HEAD and stash list unchanged")}},
		// Case 10: custom-paths-replace-defaults — explicit path... replaces the
		// default paths and resolves repo-root-relative.
		{setup: func(fx *rgiFx) {
			fx.write("notes/incident.md", "incident note\n")
			fx.g.git(fx.repo, "stash", "-q", "-u")
			fx.startRevert()
		}, args: repoArg("notes"), checks: []rgiCheck{
			rgiRC("case 10: custom path dry-run exits 0", 0),
			rgiOutHas("case 10: plan restores the custom path's file", `show "stash@{0}^3:notes/incident.md"`),
			rgiAnchored("case 10: custom path resolved repo-root-relative", "notes/incident.md"),
			{"case 10: default paths absent from the plan", func(r guardResult, _ *rgiFx) bool {
				return !strings.Contains(r.stdout, "stash@{0}^3:spectre/changes")
			}}}},
		// Case 11: empty-restore-set-refusal — a path the stash's third parent
		// does not hold refuses (exit 1) before the abort, even with --apply.
		{setup: defaultPlan, args: func(fx *rgiFx) []string { return []string{"--apply", fx.repo, "spectre/nope"} }, checks: []rgiCheck{
			rgiRC("case 11: exits 1 when the restore set is empty", 1),
			rgiNoStdout("case 11: refusal leaves stdout empty"),
			rgiErrHas("case 11: stderr names the empty-restore-set cause", "holds no files"),
			rgiRevert("case 11: --apply aborted nothing — revert still in progress", true),
			{"case 11: nothing restored", func(_ guardResult, fx *rgiFx) bool {
				return !fx.exists(rgiP1) && !fx.exists(rgiP2)
			}},
			rgiErrIs("case 11: port: stderr is the holds-no-files line", func(*rgiFx) string {
				return self + "stash@{0}^3 holds no files under: spectre/nope\n"
			}),
			rgiUnchanged("case 11: port: refusal leaves porcelain, HEAD and stash list unchanged")}},
		// Case 12: reflog-block-bounded-at-15 — the diagnosis block carries the
		// last 15 entries when the reflog holds more.
		{setup: func(fx *rgiFx) {
			fx.stashPlanning()
			for i := 0; i < 14; i++ {
				fx.g.git(fx.repo, "commit", "-q", "--allow-empty", "-m", "fill")
			}
			fx.startRevert()
		}, checks: []rgiCheck{
			rgiRC("case 12: dry-run exits 0", 0),
			{"case 12: diagnosis block carries exactly 15 entries", func(r guardResult, _ *rgiFx) bool {
				return rgiCount(r.stdout, "HEAD@{") == 15
			}}}},
		// Case 13: default-path-excludes-docs-research — with NO path argument,
		// the plan restores spectre/changes only.
		{setup: func(fx *rgiFx) {
			fx.write("docs/research/kan-423.md", "retired note\n")
			fx.write(rgiP2, "plan\n")
			fx.g.git(fx.repo, "stash", "-q", "-u")
			fx.startRevert()
		}, checks: []rgiCheck{
			rgiRC("case 13: default-path dry-run exits 0", 0),
			rgiOutHas("case 13: default plan restores the spectre/changes file", `show "stash@{0}^3:`+rgiP2+`"`),
			{"case 13: docs/research absent from the default plan", func(r guardResult, _ *rgiFx) bool {
				return !strings.Contains(r.stdout, "docs/research")
			}}}},
		// Port: a failed restore step ends the run with its exit status, the
		// abort already done — the redirect onto a directory, and a mkdir -p
		// blocked by a file.
		{setup: func(fx *rgiFx) {
			defaultPlan(fx)
			if fx.g.err == nil {
				fx.g.err = os.MkdirAll(fx.repo+"/"+rgiP2, 0o755)
			}
		}, args: func(fx *rgiFx) []string { return []string{"--apply", fx.repo} }, checks: []rgiCheck{
			rgiRC("port: restore onto a directory: exits 1", 1),
			rgiRevert("port: restore onto a directory: the abort ran first", false),
			{"port: restore onto a directory: the earlier file is restored", func(_ guardResult, fx *rgiFx) bool {
				return fx.read(rgiP1) == rgiPlan1
			}},
			rgiErrHas("port: restore onto a directory: stderr names the target", rgiP2)}},
		{setup: func(fx *rgiFx) {
			fx.write("x/f.md", "y\n")
			fx.g.git(fx.repo, "stash", "-q", "-u")
			fx.startRevert()
			fx.write("x", "blocker\n")
		}, args: func(fx *rgiFx) []string { return []string{"--apply", fx.repo, "x"} }, checks: []rgiCheck{
			rgiRC("port: mkdir blocked by a file: exits 1", 1),
			rgiRevert("port: mkdir blocked by a file: the abort ran first", false),
			rgiErrIs("port: mkdir blocked by a file: stderr is mkdir's own line", func(fx *rgiFx) string {
				cmd := exec.Command("mkdir", "-p", fx.repo+"/x")
				b, _ := cmd.CombinedOutput()
				return string(b)
			})}},
		// Port: the usage paths the harness never reached.
		{fresh: true, args: func(*rgiFx) []string { return []string{"--help"} }, checks: []rgiCheck{
			rgiRC("port: --help exits 0", 0),
			{"port: --help prints the usage on stdout, stderr empty", func(r guardResult, _ *rgiFx) bool {
				return r.stdout == "usage: recover-guard-incident [--apply] [repo-dir] [path...]\n"+
					"  --apply    execute; default is a dry-run plan\n"+
					"  repo-dir   git repository, default cwd\n"+
					"  path...    planning paths, repo-root-relative; default spectre/changes\n" && r.err == ""
			}}}},
		{setup: defaultPlan, args: repoArg("--apply"), checks: []rgiCheck{
			rgiRC("port: --apply after repo-dir is an unknown option: exits 2", 2),
			rgiErrIs("port: --apply after repo-dir: stderr names it", func(*rgiFx) string {
				return self + "unknown option: --apply\n"
			}),
			rgiUnchanged("port: --apply after repo-dir: nothing changed")}},
		{setup: func(fx *rgiFx) { fx.write("file", "") }, args: func(fx *rgiFx) []string {
			return []string{fx.repo + "/file"}
		}, checks: []rgiCheck{
			rgiRC("port: repo-dir that is a file: exits 2", 2),
			rgiNoStdout("port: repo-dir that is a file: stdout empty"),
			rgiErrIs("port: repo-dir that is a file: stderr names it as given", func(fx *rgiFx) string {
				return self + "not a directory: " + fx.repo + "/file\n"
			})}},
		// `cd "$1"` refuses these three where a lexical Join would not: an
		// empty argument (bash 5's "null directory"), a missing component
		// before `..`, and a directory without search permission.
		{setup: func(fx *rgiFx) { defaultPlan(fx); fx.cwd = fx.repo }, args: func(*rgiFx) []string { return []string{"--apply", ""} }, checks: []rgiCheck{
			rgiRC("port: empty repo-dir under --apply: exits 2", 2),
			rgiNoStdout("port: empty repo-dir under --apply: stdout empty"),
			rgiErrIs("port: empty repo-dir under --apply: stderr names it", func(*rgiFx) string {
				return self + "not a directory: \n"
			}),
			rgiUnchanged("port: empty repo-dir under --apply: nothing changed")}},
		// cd resolves an absolute `<symlink>/..` lexically, to the link's
		// own parent -- here the sandbox, not the repository.
		{setup: func(fx *rgiFx) {
			defaultPlan(fx)
			fx.write("sub/deep/x", "")
			if err := os.Symlink(fx.repo+"/sub/deep", fx.dir+"/ldeep"); err != nil {
				panic(err)
			}
		}, args: func(fx *rgiFx) []string { return []string{"--apply", fx.dir + "/ldeep/.."} }, checks: []rgiCheck{
			rgiRC("port: absolute symlink/.. under --apply: exits 2", 2),
			rgiErrIs("port: absolute symlink/.. under --apply: resolved lexically", func(fx *rgiFx) string {
				return self + "not a git repository: " + fx.dir + "\n"
			}),
			rgiUnchanged("port: absolute symlink/.. under --apply: nothing changed")}},
		// When the lexical path does not exist, cd falls back to the argument
		// as given, resolved physically through the symlink -- here into the
		// repository, so the bash acts.
		{setup: func(fx *rgiFx) {
			defaultPlan(fx)
			fx.write("a/b/x", "")
			fx.write("a/sub/x", "")
			if err := os.Symlink(fx.repo+"/a/b", fx.dir+"/yb"); err != nil {
				panic(err)
			}
		}, args: func(fx *rgiFx) []string { return []string{fx.dir + "/yb/../sub"} }, checks: []rgiCheck{
			rgiRC("port: absolute symlink/../x with no lexical target falls back to the physical path: exits 0", 0)}},
		// Logical cd checks only the lexical path: `w/c/../wsub` is the
		// directory w/wsub (here a plain directory), though its physical
		// reading through the w/c link does not exist.
		{setup: func(fx *rgiFx) {
			fx.write("a/c/x", "")
			for _, d := range []string{"/w", "/other"} {
				if err := os.Mkdir(fx.dir+d, 0o755); err != nil {
					panic(err)
				}
			}
			for l, tgt := range map[string]string{"/w/c": "../repo/a/c", "/w/wsub": "../other"} {
				if err := os.Symlink(tgt, fx.dir+l); err != nil {
					panic(err)
				}
			}
		}, args: func(*rgiFx) []string { return []string{"w/c/../wsub"} }, checks: []rgiCheck{
			rgiRC("port: a lexically valid symlink/.. path is used though its physical reading is absent: exits 2", 2),
			rgiErrIs("port: a lexically valid symlink/.. path resolves lexically", func(fx *rgiFx) string {
				return self + "not a git repository: " + fx.dir + "/other\n"
			})}},
		// A prefix before a `..` that is not a directory fails the lexical
		// canonicalisation, so cd falls back to the physical path -- into
		// the repository, never the lexical lv/sub.
		{setup: func(fx *rgiFx) {
			defaultPlan(fx)
			fx.write("a/b/x", "")
			fx.write("a/m/x", "")
			fx.write("a/sub/x", "")
			for _, d := range []string{"/lv", "/other"} {
				if err := os.Mkdir(fx.dir+d, 0o755); err != nil {
					panic(err)
				}
			}
			for l, tgt := range map[string]string{"/lv/lb": "../repo/a/b", "/lv/sub": "../other"} {
				if err := os.Symlink(tgt, fx.dir+l); err != nil {
					panic(err)
				}
			}
		}, args: func(*rgiFx) []string { return []string{"lv/lb/../m/../sub"} }, checks: []rgiCheck{
			rgiRC("port: a missing prefix before .. falls back to the physical path: exits 0", 0)}},
		{setup: func(fx *rgiFx) { fx.write("sub/x", "") }, args: func(*rgiFx) []string {
			return []string{"repo/nonexist/../sub"}
		}, checks: []rgiCheck{
			rgiRC("port: missing component before ..: exits 2", 2),
			rgiErrIs("port: missing component before ..: stderr names it as given", func(*rgiFx) string {
				return self + "not a directory: repo/nonexist/../sub\n"
			})}},
		{setup: func(fx *rgiFx) {
			if err := os.Mkdir(fx.repo+"/d", 0o600); err != nil {
				panic(err)
			}
		}, args: func(fx *rgiFx) []string { return []string{fx.repo + "/d"} }, checks: []rgiCheck{
			rgiRC("port: repo-dir without search permission: exits 2", 2),
			rgiErrIs("port: repo-dir without search permission: stderr names it as given", func(fx *rgiFx) string {
				return self + "not a directory: " + fx.repo + "/d\n"
			})}},
	}

	for i, c := range cases {
		// The sandbox belongs to the parent test, so it outlives the
		// parallel subtest reading this case's one run.
		dir, err := filepath.EvalSymlinks(t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		fx := &rgiFx{dir: dir, repo: dir, cwd: dir}
		if !c.fresh {
			fx.repo = dir + "/repo"
			gdcCopyTree(t, tmpl, fx.repo)
		}
		run := func() (guardResult, error) {
			if c.setup != nil {
				c.setup(fx)
			}
			if fx.g.err != nil {
				return guardResult{}, fx.g.err
			}
			if !c.fresh {
				fx.reflog, _ = fx.out("reflog", "-g", "HEAD", "-n", "15")
				fx.before = fx.state()
			}
			args := []string{fx.repo}
			if c.args != nil {
				args = c.args(fx)
			}
			r := runGuard("recover-guard-incident", args, Env{Dir: fx.cwd, Getenv: os.Getenv, LookupEnv: os.LookupEnv})
			if r.rc == -1 {
				return r, errors.New(r.out)
			}
			if !c.fresh {
				fx.after = fx.state()
			}
			return r, nil
		}
		// One parallel subtest runs the case, its checks nested beneath it:
		// a parallel subtest per check held a -parallel slot apiece while
		// the one running the case worked and the rest waited on it.
		t.Run(fmt.Sprint("run ", i), func(t *testing.T) {
			t.Parallel()
			r, err := run()
			for _, chk := range c.checks {
				t.Run(chk.label, func(t *testing.T) {
					if err != nil {
						t.Fatal(err)
					}
					if !chk.ok(r, fx) {
						t.Fatalf("rc=%d\nstdout=%s\nstderr=%s", r.rc, r.stdout, r.err)
					}
				})
			}
		})
	}
}
