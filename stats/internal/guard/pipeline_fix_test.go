package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/tweety53/agents/stats/internal/guard/core"
	"github.com/tweety53/agents/stats/internal/guard/guardtest"
	"github.com/tweety53/agents/stats/internal/guard/hook"
	"github.com/tweety53/agents/stats/internal/smoke"
)

// pfFx is one pipeline-fix fixture: a bare origin, the agents main checkout
// cloned from it (model-policy.md's **Allowed pairs**, a .flow/project.md, a
// script and its harness), a second clone that moves origin, and a `flow` stub logging every
// call, its stdin included, answering `record dispatches` from rows and
// `state list` from states — flowd's change records, pipeline-fix-old-5's
// alone unless a test rewrites it.
type pfFx struct {
	lpFx
	origin, other, states string
}

const pfProject = "# fixture\n\n## worktree setup\n\n```bash\ntouch .setup-ran\n```\n\n" +
	"## lint\n\n```bash\ntrue\ncd stats && false\n```\n\n## default landing route\n\nmerge and push\n"

func pfNew(t *testing.T, rows string) pfFx {
	t.Helper()
	tmp, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	fx := pfFx{lpFx: lpFx{repo: tmp + "/agents", log: tmp + "/flow.log", rows: tmp + "/rows.json"},
		origin: tmp + "/origin.git", other: tmp + "/other", states: tmp + "/states.json"}
	table, err := os.ReadFile("../../../" + core.PolicyFile)
	if err != nil {
		t.Fatal(err)
	}
	guardtest.FixtureRepo(t, fx.origin, guardtest.RepoShape{Branch: "main", Bare: true})
	guardtest.GitRun(t, "", "clone", "-q", fx.origin, fx.repo)
	pfIdentity(t, fx.repo)
	guardtest.WriteFile(t, fx.repo+"/"+core.PolicyFile, string(table))
	guardtest.WriteFile(t, fx.repo+"/.flow/project.md", pfProject)
	guardtest.WriteFile(t, fx.repo+"/.gitignore", ".setup-ran\n.superpowers/\nspectre/changes/\n")
	guardtest.WriteExec(t, fx.repo+"/scripts/foo.sh", "#!/usr/bin/env bash\necho foo\n")
	guardtest.WriteExec(t, fx.repo+"/scripts/test-foo.sh", "#!/usr/bin/env bash\nscripts/foo.sh | grep -q foo\n")
	guardtest.GitRun(t, fx.repo, "add", "-A")
	guardtest.GitRun(t, fx.repo, "commit", "-q", "-m", "init")
	guardtest.GitRun(t, fx.repo, "push", "-q", "origin", "HEAD:main")
	guardtest.GitRun(t, fx.repo, "remote", "set-head", "origin", "main")
	guardtest.GitRun(t, "", "clone", "-q", fx.origin, fx.other)
	if rows != "" {
		guardtest.WriteFile(t, fx.rows, rows)
	}
	guardtest.WriteFile(t, fx.states, `{"source":"store","complete":true,"records":[{"name":"pipeline-fix-old-5","state":"FINISHED"}]}`)
	bin := tmp + "/bin"
	guardtest.WriteExec(t, bin+"/flow", "#!/usr/bin/env bash\n"+guardtest.PolicyStub+`printf '%s\n' "$*" >>`+fx.log+`
case "$1 $2" in
  "record dispatches") [ -f `+fx.rows+` ] && exec cat `+fx.rows+`; exit 1 ;;
  "state list") exec cat `+fx.states+` ;;
  "state set") cat >>`+fx.log+`; echo >>`+fx.log+` ;;
esac
`)
	projectGet := guardtest.GuardShim(t, "project-get")
	fx.env = core.Env{Dir: tmp, Getenv: func(k string) string {
		switch k {
		case "FLOW_GUARD_REPO_ROOT":
			return fx.repo
		case "FLOW_GUARD_PROJECT_GET":
			return projectGet
		}
		return guardtest.PathEnv(bin)(k)
	}}
	return fx
}

func (fx pfFx) pf(t *testing.T, args ...string) (int, string) {
	t.Helper()
	var o, e bytes.Buffer
	rc := pipelineFix(args, fx.env, &o, &e)
	return rc, o.String() + e.String()
}

func (fx pfFx) wt(branch string) string {
	return filepath.Join(filepath.Dir(fx.repo), "agents-worktrees", branch)
}

// pfIdentity gives the guard's own git calls in repo (a rebase commits, in the
// repository or a worktree of it) an identity: its local config, never the
// process environment, so these tests run in parallel.
func pfIdentity(t *testing.T, repo string) {
	t.Helper()
	guardtest.GitRun(t, repo, "config", "user.name", "test")
	guardtest.GitRun(t, repo, "config", "user.email", "test@example.com")
}

// smoke: core — a pipeline fix's start
func TestPipelineFixStart(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		args   []string
		branch string
	}{
		{[]string{"start", "kan-1-demo", "6"}, lpBranch},
		{[]string{"start", "kan-1-demo"}, "self-review-kan-1-demo"},
	} {
		fx := pfNew(t, "")
		rc, out := fx.pf(t, tc.args...)
		wt := fx.wt(tc.branch)
		if rc != 0 || !strings.Contains(out, "worktree: "+wt) {
			t.Fatalf("%v: rc %d: %s", tc.args, rc, out)
		}
		if b, _ := os.ReadFile(wt + "/.git"); len(b) == 0 {
			t.Fatalf("%v: no worktree at %s", tc.args, wt)
		}
		for _, p := range []string{wt + "/.superpowers/sdd/" + tc.branch, wt + "/.setup-ran"} {
			if _, err := os.Stat(p); err != nil {
				t.Fatalf("%v: %v", tc.args, err)
			}
		}
		calls := fx.calls(t)
		if !strings.Contains(calls, "state set -C "+wt+" "+tc.branch) || !strings.Contains(calls, `"state":"STARTED"`) {
			t.Fatalf("%v: no STARTED record for the loop: %s", tc.args, calls)
		}
	}
}

func TestPipelineFixStartLeavesNothingWhenTheStateWriteFails(t *testing.T) {
	t.Parallel()
	fx := pfNew(t, "")
	flow, ok := core.LookPath(fx.env, "flow")
	if !ok {
		t.Fatal("no flow stub")
	}
	guardtest.WriteExec(t, flow, "#!/usr/bin/env bash\necho 'flow: state set refused' >&2\nexit 1\n")
	if rc, out := fx.pf(t, "start", "kan-1-demo"); rc != 1 || !strings.Contains(out, "state set refused") {
		t.Fatalf("rc %d: %s", rc, out)
	}
	branch := "self-review-kan-1-demo"
	if _, err := os.Stat(fx.wt(branch)); !os.IsNotExist(err) {
		t.Fatalf("worktree left behind: %v", err)
	}
	if rc, _ := core.RunCombined(fx.env, "", "git", "-C", fx.repo, "rev-parse", "-q", "--verify", "refs/heads/"+branch); rc == 0 {
		t.Fatal("branch left behind")
	}
}

func TestPipelineFixStartRefusesANonUniqueK(t *testing.T) {
	t.Parallel()
	fx := pfNew(t, "")
	for _, k := range []string{"5", "7"} {
		if rc, out := fx.pf(t, "start", "kan-1-demo", k); rc != 1 || !strings.Contains(out, "<k> is 6") {
			t.Fatalf("k %s: rc %d: %s", k, rc, out)
		}
	}
	guardtest.GitRun(t, fx.repo, "branch", "pipeline-fix-other-9")
	if rc, out := fx.pf(t, "start", "kan-1-demo", "10"); rc != 0 {
		t.Fatalf("one past a live branch: rc %d: %s", rc, out)
	}
}

func TestPipelineFixEnd(t *testing.T) {
	t.Parallel()
	fx := pfNew(t, lpRowsJSON("pipeline-fix-6-", lpBranch, "plan completed"))
	guardtest.GitRun(t, fx.repo, "branch", lpBranch)
	if rc, out := fx.pf(t, "end", "pipeline-fix-6-review-1", "completed"); rc != 2 {
		t.Fatalf("a review closed completed: rc %d: %s", rc, out)
	}
	report := func(body string) string {
		p := filepath.Join(t.TempDir(), "report.md")
		guardtest.WriteFile(t, p, body)
		return p
	}
	if rc, out := fx.pf(t, "end", "pipeline-fix-6-review-1", report("Verdict: ok\n")); rc != 2 || !strings.Contains(out, "Critical: <n>") {
		t.Fatalf("a report with no counts: rc %d: %s", rc, out)
	}
	if rc, out := fx.pf(t, "end", "pipeline-fix-6-review-1", report("- **Critical:** 0\n- **Important:** 1\n")); rc != 0 ||
		!strings.Contains(fx.calls(t), "-key pipeline-fix-6-review-1 -session-token "+lpBranch+" -outcome fix") {
		t.Fatalf("an Important closed as fix: rc %d: %s\n%s", rc, out, fx.calls(t))
	}
	if rc, out := fx.pf(t, "end", "pipeline-fix-6-review-1", report("Critical: 0\nImportant: 0\nMinor: 2\n")); rc != 0 ||
		!strings.Contains(fx.calls(t), "record dispatch end -C "+fx.repo+" -change "+lpBranch+" -key pipeline-fix-6-review-1 -session-token "+lpBranch+" -outcome clean") {
		t.Fatalf("rc %d: %s\n%s", rc, out, fx.calls(t))
	}
	if rc, _ := fx.pf(t, "escalate", "pipeline-fix-6-fix-1"); rc != 2 {
		t.Fatalf("escalate is gone: rc %d, want usage 2", rc)
	}
}

func TestPipelineFixVerify(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	fx := pfNew(t, "")
	if rc, out := fx.pf(t, "start", "kan-1-demo", "6"); rc != 0 {
		t.Fatal(out)
	}
	wt := fx.wt(lpBranch)
	guardtest.WriteExec(t, wt+"/scripts/foo.sh", "#!/usr/bin/env bash\necho foo bar\n")
	guardtest.GitRun(t, wt, "commit", "-q", "-am", "fix: foo")
	rc, out := fx.pf(t, "verify", wt)
	if rc != 0 || !strings.Contains(out, "green: true") || !strings.Contains(out, "green: scripts/test-foo.sh") || strings.Contains(out, "cd stats") {
		t.Fatalf("rc %d: %s", rc, out)
	}
	guardtest.WriteExec(t, wt+"/scripts/foo.sh", "#!/usr/bin/env bash\necho bar\n")
	guardtest.GitRun(t, wt, "commit", "-q", "-am", "fix: break foo")
	if rc, out := fx.pf(t, "verify", wt); rc != 1 || !strings.Contains(out, "red: scripts/test-foo.sh (exit 1)") {
		t.Fatalf("a red harness: rc %d: %s", rc, out)
	}
}

// A touched Go file selects the repotest audits of the whole Go tree, which
// no Go file's name leads to; a touched script alone does not.
func TestPipelineFixChecksGoAudits(t *testing.T) {
	t.Parallel()
	wt := t.TempDir()
	guardtest.WriteFile(t, wt+"/.flow/project.md", pfProject)
	pe0 := pfEnv{}
	if cmds, _ := pfChecks(core.Env{Getenv: os.Getenv}, pe0, wt, []string{"stats/internal/guard/x.go", "hooks/h.py"}); slices.ContainsFunc(cmds, func(c string) bool { return strings.Contains(c, "internal/repotest") }) {
		t.Fatalf("a worktree without stats/internal/repotest: %q selects it", cmds)
	}
	guardtest.WriteFile(t, wt+"/stats/internal/repotest/doc.go", "package repotest\n")
	audits := "cd stats && go test ./internal/repotest/ -run '^Test(GitConfigPins|FlowAddrDeclaration)$' -count=1"
	pe := pfEnv{}
	env := core.Env{Getenv: os.Getenv}
	if cmds, rc := pfChecks(env, pe, wt, []string{"stats/internal/guard/x.go"}); rc != 0 || !slices.Contains(cmds, audits) {
		t.Fatalf("a touched Go file: rc %d, %q lacks %s", rc, cmds, audits)
	}
	if cmds, _ := pfChecks(env, pe, wt, []string{"scripts/foo.sh"}); slices.Contains(cmds, audits) {
		t.Fatalf("a touched script alone: %q selects %s", cmds, audits)
	}
}

// A `# concurrent` annotation line of the `## lint` fence is no check
// (request 72).
func TestPipelineFixChecksSkipLintAnnotations(t *testing.T) {
	t.Parallel()
	wt := t.TempDir()
	guardtest.WriteFile(t, wt+"/.flow/project.md", "## lint\n\n```bash\n# concurrent\ntrue\n```\n")
	cmds, rc := pfChecks(core.Env{Getenv: os.Getenv}, pfEnv{}, wt, nil)
	if rc != 0 || slices.ContainsFunc(cmds, func(c string) bool { return strings.HasPrefix(c, "#") }) || !slices.Contains(cmds, "true") {
		t.Fatalf("rc %d, checks %q: want `true` and no `#` line", rc, cmds)
	}
}

// A touched script or setup.sh selects the guard package's script audits
// where the worktree has stats/internal/guard (kan-1038).
func TestPipelineFixChecksScriptAudits(t *testing.T) {
	t.Parallel()
	wt := t.TempDir()
	guardtest.WriteFile(t, wt+"/.flow/project.md", pfProject)
	pe := pfEnv{}
	env := core.Env{Getenv: os.Getenv}
	if cmds, _ := pfChecks(env, pe, wt, []string{"scripts/pipeline-fix.sh"}); slices.Contains(cmds, pfScriptAudits) {
		t.Fatalf("a worktree without stats/internal/guard: %q selects %s", cmds, pfScriptAudits)
	}
	guardtest.WriteFile(t, wt+"/stats/internal/guard/doc.go", "package guard\n")
	for _, p := range []string{"scripts/pipeline-fix.sh", "scripts/lib/flow-guard.sh", "setup.sh"} {
		if cmds, rc := pfChecks(env, pe, wt, []string{p}); rc != 0 || !slices.Contains(cmds, pfScriptAudits) {
			t.Fatalf("touched %s: rc %d, %q lacks %s", p, rc, cmds, pfScriptAudits)
		}
	}
	if cmds, _ := pfChecks(env, pe, wt, []string{"stats/internal/guard/x.go"}); slices.Contains(cmds, pfScriptAudits) {
		t.Fatalf("a touched Go file alone: %q selects %s", cmds, pfScriptAudits)
	}
}

// A touched hook selects the whole repotest package, where its test lives.
func TestPipelineFixChecksHooks(t *testing.T) {
	t.Parallel()
	wt := t.TempDir()
	guardtest.WriteFile(t, wt+"/.flow/project.md", pfProject)
	guardtest.WriteFile(t, wt+"/stats/internal/repotest/doc.go", "package repotest\n")
	pe := pfEnv{}
	want := "cd stats && go test ./internal/repotest/ -count=1"
	if cmds, rc := pfChecks(core.Env{Getenv: os.Getenv}, pe, wt, []string{"hooks/block-unsafe-commands.py"}); rc != 0 || !slices.Contains(cmds, want) {
		t.Fatalf("a touched hook: rc %d, %q lacks %s", rc, cmds, want)
	}
}

func TestPipelineFixLand(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	reviewed := lpRowsJSON("pipeline-fix-6-", lpBranch, "plan completed", "fix completed", "review-1 clean")
	for _, tc := range []struct {
		name, rows string
		moved      bool
		rc         int
		red        bool
		refusal    string
	}{
		{name: "a reviewed loop lands", rows: reviewed},
		{name: "a moved base is rebased and verified first", rows: reviewed, moved: true},
		{name: "an unreviewed loop lands: its rows refuse nothing", rows: lpRowsJSON("pipeline-fix-6-", lpBranch, "plan completed", "fix completed")},
		{name: "a red branch is refused with no base move", rows: reviewed, red: true, rc: 1, refusal: "red: scripts/test-foo.sh"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fx := pfNew(t, tc.rows)
			if rc, out := fx.pf(t, "start", "kan-1-demo", "6"); rc != 0 {
				t.Fatal(out)
			}
			wt := fx.wt(lpBranch)
			guardtest.WriteFile(t, wt+"/docs/fix.md", "fix\n")
			guardtest.GitRun(t, wt, "add", "docs/fix.md")
			guardtest.GitRun(t, wt, "commit", "-q", "-m", "docs: fix")
			if tc.red {
				guardtest.WriteExec(t, wt+"/scripts/foo.sh", "#!/usr/bin/env bash\necho bar\n")
				guardtest.GitRun(t, wt, "commit", "-q", "-am", "fix: break foo")
			}
			if tc.moved {
				guardtest.WriteFile(t, fx.other+"/docs/other.md", "other\n")
				guardtest.GitRun(t, fx.other, "add", "docs/other.md")
				guardtest.GitRun(t, fx.other, "commit", "-q", "-m", "docs: other")
				guardtest.GitRun(t, fx.other, "push", "-q", "origin", "HEAD:main")
			}
			rc, out := fx.pf(t, "land", wt, "important")
			if rc != tc.rc {
				t.Fatalf("rc %d, want %d: %s", rc, tc.rc, out)
			}
			_, log := core.RunCombined(fx.env, "", guardtest.FixtureGit, "--git-dir", fx.origin, "log", "--format=%s", "main")
			if tc.rc != 0 {
				if strings.Contains(log, "docs: fix") || !strings.Contains(out, tc.refusal) {
					t.Fatalf("refused landing pushed or named no review: %s\n%s", out, log)
				}
				return
			}
			if !strings.Contains(log, "docs: fix") || !strings.Contains(out, "landed: ") {
				t.Fatalf("not landed: %s\n%s", out, log)
			}
			if tc.moved && (!strings.Contains(log, "docs: other") || !strings.Contains(out, "green: true")) {
				t.Fatalf("the moved base was not rebased onto and verified: %s\n%s", out, log)
			}
			if _, err := os.Stat(wt); err == nil || hook.BranchExists(fx.env, guardtest.FixtureGit, fx.repo, lpBranch) {
				t.Fatalf("worktree or branch left behind: %s", out)
			}
			if !strings.Contains(fx.calls(t), `"state":"FINISHED"`) {
				t.Fatalf("no FINISHED record: %s", fx.calls(t))
			}
		})
	}
}

// TestPipelineFixLandRecordsTheFixedRow pins that land of a pipeline-fix
// loop records the change's flow-fix `fixed` row itself, from the directory
// it runs in, and refuses a missing or misplaced grade before anything is
// pushed (kan-1104 self-review, F2: the row was left to the parent and went
// unrecorded).
func TestPipelineFixLandRecordsTheFixedRow(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	for _, tc := range []struct {
		name, start, branch string
		grade               []string
		rc                  int
	}{
		{name: "records", start: "6", branch: lpBranch, grade: []string{"important"}},
		{name: "no grade", start: "6", branch: lpBranch, rc: 2},
		{name: "grade on a self-review branch", branch: "self-review-kan-1-demo", grade: []string{"minor"}, rc: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prefix := "pipeline-fix-6-"
			if tc.start == "" {
				prefix = tc.branch + "-"
			}
			fx := pfNew(t, lpRowsJSON(prefix, tc.branch, "plan completed", "fix completed", "review-1 clean"))
			start := []string{"start", "kan-1-demo"}
			if tc.start != "" {
				start = append(start, tc.start)
			}
			if rc, out := fx.pf(t, start...); rc != 0 {
				t.Fatal(out)
			}
			wt := fx.wt(tc.branch)
			guardtest.WriteFile(t, wt+"/docs/fix.md", "fix\n")
			guardtest.GitRun(t, wt, "add", "docs/fix.md")
			guardtest.GitRun(t, wt, "commit", "-q", "-m", "docs: fix")
			_, first := core.RunCombined(fx.env, "", guardtest.FixtureGit, "-C", wt, "rev-parse", "HEAD")
			guardtest.WriteFile(t, wt+"/docs/more.md", "more\n")
			guardtest.GitRun(t, wt, "add", "docs/more.md")
			guardtest.GitRun(t, wt, "commit", "-q", "-m", "docs: more")
			var o, e bytes.Buffer
			rc := pipelineFix(append([]string{"land", wt}, tc.grade...), fx.env, &o, &e)
			_, log := core.RunCombined(fx.env, "", guardtest.FixtureGit, "--git-dir", fx.origin, "log", "--format=%s", "main")
			if rc != tc.rc {
				t.Fatalf("rc %d, want %d: %s%s", rc, tc.rc, o.String(), e.String())
			}
			if tc.rc != 0 {
				if strings.Contains(log, "docs: fix") || !strings.Contains(e.String(), "important|minor") {
					t.Fatalf("a refused grade pushed or named no grade: %s%s\n%s", o.String(), e.String(), log)
				}
				return
			}
			want := "self-review finding -C " + fx.env.Dir + " -change kan-1-demo -angle flow-fix -disposition fixed -ref " +
				strings.TrimSpace(first) + " -blast-radius 2 -severity important -note docs: fix\n"
			if !strings.Contains(log, "docs: fix") || strings.Count(fx.calls(t), "self-review finding") != 1 || !strings.Contains(fx.calls(t), want) {
				t.Fatalf("no fixed row %q: %s\n%s", want, o.String(), fx.calls(t))
			}
		})
	}
}

// TestPipelineFixLandFromItsOwnWorktree pins that land run from inside the
// loop worktree it removes -- the documented `cd <loop worktree> &&
// scripts/pipeline-fix.sh land <wt>` -- still deletes the branch and prints
// landed: (kan-1075 self-review, F13: every git call after the removal ran
// in the deleted cwd, failed to chdir and exited 1 printing nothing).
func TestPipelineFixLandFromItsOwnWorktree(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	fx := pfNew(t, lpRowsJSON("pipeline-fix-6-", lpBranch, "plan completed", "fix completed", "review-1 clean"))
	if rc, out := fx.pf(t, "start", "kan-1-demo", "6"); rc != 0 {
		t.Fatal(out)
	}
	wt := fx.wt(lpBranch)
	guardtest.WriteFile(t, wt+"/docs/fix.md", "fix\n")
	guardtest.GitRun(t, wt, "add", "docs/fix.md")
	guardtest.GitRun(t, wt, "commit", "-q", "-m", "docs: fix")
	fx.env.Dir = wt
	rc, out := fx.pf(t, "land", wt, "important")
	if rc != 0 || !strings.Contains(out, "landed: ") {
		t.Fatalf("rc %d, want 0 and landed: %s", rc, out)
	}
	fx.env.Dir = fx.repo // the checks below must not run in the removed worktree
	if _, err := os.Stat(wt); err == nil || hook.BranchExists(fx.env, guardtest.FixtureGit, fx.repo, lpBranch) {
		t.Fatalf("worktree or branch left behind: %s", out)
	}
}

// TestPipelineFixDropFromItsOwnWorktree pins the same F13 fix for drop: run
// from inside the loop worktree it removes, it still deletes the worktree
// and the branch and exits 0.
func TestPipelineFixDropFromItsOwnWorktree(t *testing.T) {
	t.Parallel()
	fx := pfNew(t, "")
	if rc, out := fx.pf(t, "start", "kan-1-demo", "6"); rc != 0 {
		t.Fatal(out)
	}
	wt := fx.wt(lpBranch)
	fx.env.Dir = wt
	rc, out := fx.pf(t, "drop", wt)
	if rc != 0 || !strings.Contains(out, "dropped: "+lpBranch) {
		t.Fatalf("rc %d, want 0 and dropped: %s", rc, out)
	}
	fx.env.Dir = fx.repo // the checks below must not run in the removed worktree
	if _, err := os.Stat(wt); err == nil || hook.BranchExists(fx.env, guardtest.FixtureGit, fx.repo, lpBranch) {
		t.Fatalf("worktree or branch left behind: %s", out)
	}
}

// TestPipelineFixRemoveNamesASilentFailure pins that a removal step failing
// with no output is named on stderr rather than leaving an exit 1 that
// prints nothing (kan-1075 self-review, F13).
func TestPipelineFixRemoveNamesASilentFailure(t *testing.T) {
	t.Parallel()
	fx := pfNew(t, "")
	if rc, out := fx.pf(t, "start", "kan-1-demo", "6"); rc != 0 {
		t.Fatal(out)
	}
	git := filepath.Join(t.TempDir(), "git")
	guardtest.WriteExec(t, git, "#!/usr/bin/env bash\ncase \" $* \" in *\" branch -D \"*) exit 1 ;; esac\nexec "+guardtest.FixtureGit+" \"$@\"\n")
	var stderr bytes.Buffer
	if rc := pfRemove(fx.env, pfEnv{git: git, main: fx.repo}, fx.wt(lpBranch), lpBranch, false, &stderr); rc != 1 || !strings.Contains(stderr.String(), "branch -D "+lpBranch+" exited 1") {
		t.Fatalf("rc %d, stderr %q: want 1 naming the failed branch -D", rc, stderr.String())
	}
}

func TestPipelineFixDrop(t *testing.T) {
	t.Parallel()
	fx := pfNew(t, "")
	if rc, out := fx.pf(t, "start", "kan-1-demo", "6"); rc != 0 {
		t.Fatal(out)
	}
	wt := fx.wt(lpBranch)
	guardtest.WriteFile(t, wt+"/scratch.txt", "left\n")
	if rc, out := fx.pf(t, "drop", wt); rc != 0 || !strings.Contains(out, "dropped: "+lpBranch) {
		t.Fatalf("rc %d: %s", rc, out)
	}
	if _, err := os.Stat(wt); err == nil || hook.BranchExists(fx.env, guardtest.FixtureGit, fx.repo, lpBranch) {
		t.Fatal("worktree or branch left behind")
	}
	if !strings.Contains(fx.calls(t), `"withdrawn":true`) {
		t.Fatalf("no withdrawn record: %s", fx.calls(t))
	}
	if rc, out := fx.pf(t, "start", "kan-1-demo", "6"); rc != 1 || !strings.Contains(out, "<k> is 7") {
		t.Fatalf("a dropped loop's <k> was reissued: rc %d: %s", rc, out)
	}
}

// end of the pass closes the row check-loop-dispatch began, under the
// change <name> its begin used; it takes completed or aborted only.
func TestPipelineFixEndClosesThePass(t *testing.T) {
	t.Parallel()
	fx := pfNew(t, "[]")
	rc, out := fx.pf(t, "end", "self-review-kan-1-demo-pass", "completed")
	if rc != 0 {
		t.Fatalf("rc %d: %s", rc, out)
	}
	if want := "-change kan-1-demo -key self-review-kan-1-demo-pass -session-token self-review-kan-1-demo -outcome completed"; !strings.Contains(fx.calls(t), "record dispatch end") || !strings.Contains(fx.calls(t), want) {
		t.Fatalf("end of the pass did not close it with %q:\n%s", want, fx.calls(t))
	}
	if rc, out := fx.pf(t, "end", "self-review-kan-1-demo-pass", "done"); rc != 2 {
		t.Fatalf("outcome done: rc %d, want 2: %s", rc, out)
	}
}

// end of an own-branch self-review in another project runs in
// that project -- the caller's working directory -- and verify diffs a
// worktree against its own origin/HEAD, not <agents repo>'s.
func TestPipelineFixOwnBranchInAnotherProject(t *testing.T) {
	t.Parallel()
	const token = "self-review-kan-1-demo"
	fx := pfNew(t, lpRowsJSON(token+"-", token, "plan completed", "fix completed"))
	tmp := filepath.Dir(fx.repo)
	porigin, proj := tmp+"/gymie.git", tmp+"/gymie"
	guardtest.FixtureRepo(t, porigin, guardtest.RepoShape{Branch: "develop", Bare: true})
	guardtest.GitRun(t, "", "clone", "-q", porigin, proj)
	pfIdentity(t, proj)
	guardtest.WriteFile(t, proj+"/README.md", "gymie\n")
	guardtest.GitRun(t, proj, "add", "-A")
	guardtest.GitRun(t, proj, "commit", "-q", "-m", "init")
	guardtest.GitRun(t, proj, "push", "-q", "origin", "HEAD:develop")
	guardtest.GitRun(t, proj, "remote", "set-head", "origin", "develop")
	guardtest.GitRun(t, proj, "checkout", "-q", "-b", "spectre/kan-1-demo")
	guardtest.WriteFile(t, proj+"/.flow/project.md", "# gymie\n\n## lint\n\n```bash\ntrue\n```\n")
	guardtest.GitRun(t, proj, "add", "-A")
	guardtest.GitRun(t, proj, "commit", "-q", "-m", "fix: tooling")
	fx.env.Dir = proj
	if rc, out := fx.pf(t, "end", token+"-fix-1", "completed"); rc != 0 ||
		!strings.Contains(fx.calls(t), "record dispatch end -C "+proj+" -change kan-1-demo -key "+token+"-fix-1") {
		t.Fatalf("end left the caller's project: rc %d: %s\n%s", rc, out, fx.calls(t))
	}
	if rc, out := fx.pf(t, "verify", proj); rc != 0 || !strings.Contains(out, "green: true") {
		t.Fatalf("verify against the worktree's own origin/HEAD: rc %d: %s", rc, out)
	}
	p := filepath.Join(t.TempDir(), "report.md")
	guardtest.WriteFile(t, p, "Critical: 0\nImportant: 0\n")
	if rc, out := fx.pf(t, "end", token+"-review-1", p); rc != 0 {
		t.Fatalf("end review clean: rc %d: %s", rc, out)
	}
}

func TestPipelineFixLandRestartsFlowdWhenStatsMoved(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	reviewed := lpRowsJSON("pipeline-fix-6-", lpBranch, "plan completed", "fix completed", "review-1 clean")
	for _, tc := range []struct {
		name    string
		paths   []string
		restart bool
	}{
		{"a stats/ landing restarts flowd", []string{"stats/Makefile", "stats/cmd/flowd/README"}, true},
		{"a docs-only landing does not", []string{"docs/fix.md"}, false},
		{"a stats/Makefile without cmd/flowd is another project's", []string{"stats/Makefile"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fx := pfNew(t, reviewed)
			guardtest.WriteExec(t, filepath.Dir(fx.repo)+"/bin/make", "#!/usr/bin/env bash\nprintf 'make %s\\n' \"$*\" >>"+fx.log+"\n")
			if rc, out := fx.pf(t, "start", "kan-1-demo", "6"); rc != 0 {
				t.Fatal(out)
			}
			wt := fx.wt(lpBranch)
			for _, p := range tc.paths {
				guardtest.WriteFile(t, wt+"/"+p, "restart:\n\ttrue\n")
			}
			guardtest.GitRun(t, wt, "add", "-A")
			guardtest.GitRun(t, wt, "commit", "-q", "-m", "fix: "+tc.paths[0])
			rc, out := fx.pf(t, "land", wt, "important")
			if rc != 0 {
				t.Fatalf("rc %d: %s", rc, out)
			}
			restarted := strings.Contains(out, "FLOWD-RESTARTED: "+fx.repo+"/stats")
			called := strings.Contains(fx.calls(t), "make -C "+fx.repo+"/stats restart")
			if tc.restart != restarted || tc.restart != called || (!tc.restart && (strings.Contains(fx.calls(t), "make ") || strings.Contains(out, "FLOWD-"))) {
				t.Fatalf("restart %v: %s\n%s", tc.restart, out, fx.calls(t))
			}
		})
	}
}

func TestPipelineFixVerifyRunsVisualOnUIPath(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	fx := pfNew(t, lpRowsJSON("pipeline-fix-6-", lpBranch, "plan completed", "fix completed", "review-1 clean"))
	if rc, out := fx.pf(t, "start", "kan-1-demo", "6"); rc != 0 {
		t.Fatal(out)
	}
	wt := fx.wt(lpBranch)
	b, err := os.ReadFile(wt + "/.flow/project.md")
	if err != nil {
		t.Fatal(err)
	}
	guardtest.WriteFile(t, wt+"/.flow/project.md", string(b)+"\n## visual verification\n\n"+
		"| Setting | Value |\n|---------|-------|\n| `ui paths` | `web/src/**` |\n\n"+
		"| Command | Runs |\n|---------|------|\n| `verify` | `test ! -e .visual-red` |\n")
	guardtest.WriteFile(t, wt+"/docs/a.md", "a\n")
	guardtest.GitRun(t, wt, "add", "-A")
	guardtest.GitRun(t, wt, "commit", "-q", "-m", "docs: visual section")
	const vis = "test ! -e .visual-red"
	if rc, out := fx.pf(t, "verify", wt); rc != 0 || strings.Contains(out, vis) {
		t.Fatalf("(a) rc %d, no ui path ran the visual row: %s", rc, out)
	}
	guardtest.WriteFile(t, wt+"/web/src/a.ts", "a\n")
	guardtest.GitRun(t, wt, "add", "-A")
	guardtest.GitRun(t, wt, "commit", "-q", "-m", "fix: ui")
	if rc, out := fx.pf(t, "verify", wt); rc != 0 || !strings.Contains(out, "green: "+vis) {
		t.Fatalf("(b) rc %d: %s", rc, out)
	}
	guardtest.WriteFile(t, wt+"/.visual-red", "x\n")
	if rc, out := fx.pf(t, "verify", wt); rc != 1 || !strings.Contains(out, "red: "+vis+" (exit 1)") {
		t.Fatalf("(c) rc %d: %s", rc, out)
	}
	rc, out := fx.pf(t, "land", wt, "important")
	_, log := core.RunCombined(fx.env, "", guardtest.FixtureGit, "--git-dir", fx.origin, "log", "--format=%s", "main")
	if rc != 1 || !strings.Contains(out, "red: "+vis) || strings.Contains(log, "fix: ui") {
		t.Fatalf("(d) rc %d: %s\n%s", rc, out, log)
	}
}

func TestPipelineFixVerifyFillsVisualSpecs(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	fx := pfNew(t, lpRowsJSON("pipeline-fix-6-", lpBranch, "plan completed", "fix completed", "review-1 clean"))
	if rc, out := fx.pf(t, "start", "kan-1-demo", "6"); rc != 0 {
		t.Fatal(out)
	}
	wt := fx.wt(lpBranch)
	b, err := os.ReadFile(wt + "/.flow/project.md")
	if err != nil {
		t.Fatal(err)
	}
	base := string(b)
	const vis = `test "<specs>" = web/a.spec.ts`
	section := func(uiPaths string) string {
		return base + "\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n" + uiPaths +
			"\n| Command | Runs |\n|---------|------|\n| `specs` | `echo web/a.spec.ts` |\n| `verify` | `" + vis + "` |\n"
	}
	guardtest.WriteFile(t, wt+"/.flow/project.md", section("| `ui paths` | `web/src/**` |\n"))
	guardtest.WriteFile(t, wt+"/web/src/a.ts", "a\n")
	guardtest.GitRun(t, wt, "add", "-A")
	guardtest.GitRun(t, wt, "commit", "-q", "-m", "fix: ui")
	if rc, out := fx.pf(t, "verify", wt); rc != 0 || !strings.Contains(out, `green: test "web/a.spec.ts" = web/a.spec.ts`) {
		t.Fatalf("(a) rc %d, specs paths not substituted: %s", rc, out)
	}
	// A specs row that maps nothing, and no specs row at all, run the smoke
	// set in `<specs>`, never a literal `<specs>` (kan-1132).
	const smokeVis = `test "<specs>" = "--grep @smoke"`
	smokeSection := func(specsRow string) string {
		return base + "\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `web/src/**` |\n" +
			"\n| Command | Runs |\n|---------|------|\n" + specsRow + "| `verify` | `" + smokeVis + "` |\n"
	}
	for name, row := range map[string]string{"empty specs row": "| `specs` | `true` |\n", "no specs row": ""} {
		guardtest.WriteFile(t, wt+"/.flow/project.md", smokeSection(row))
		if rc, out := fx.pf(t, "verify", wt); rc != 0 || !strings.Contains(out, `green: test "--grep @smoke" = "--grep @smoke"`) {
			t.Fatalf("(%s) rc %d, smoke set not filled: %s", name, rc, out)
		}
	}
	// No `ui paths` row: the trigger cannot answer, no merge base is read,
	// so a specs row has nothing to scope to and the visual line is red.
	guardtest.WriteFile(t, wt+"/.flow/project.md", section(""))
	if rc, out := fx.pf(t, "verify", wt); rc != 1 || !strings.Contains(out, "red: visual verify — no merge base to scope the specs row to") {
		t.Fatalf("(b) rc %d: %s", rc, out)
	}
}

// TestPipelineFixStartWaitsLongerForTheStoreList pins kan-1132's
// self-review, F3: start reads flow state list with a 15s store timeout, as
// it refuses anything but a complete store answer.
func TestPipelineFixStartWaitsLongerForTheStoreList(t *testing.T) {
	smoke.Heavy(t)
	t.Parallel()
	fx := pfNew(t, "")
	if rc, out := fx.pf(t, "start", "kan-1-demo", "6"); rc != 0 {
		t.Fatal(out)
	}
	if want := "state list -C " + fx.repo + " -timeout 15s\n"; !strings.Contains(fx.calls(t), want) {
		t.Errorf("flow calls lack %q:\n%s", want, fx.calls(t))
	}
}
