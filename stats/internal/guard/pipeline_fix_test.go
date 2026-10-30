package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// pfFx is one pipeline-fix fixture: a bare origin, the agents main checkout
// cloned from it (model-policy.md's **Allowed pairs**, a .flow/project.md, a
// script and its harness), a second clone that moves origin, and a `flow` stub logging every
// call, its stdin included, and answering `record dispatches` from rows.
type pfFx struct {
	lpFx
	origin, other string
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
		origin: tmp + "/origin.git", other: tmp + "/other"}
	table, err := os.ReadFile("../../../" + mpFile)
	if err != nil {
		t.Fatal(err)
	}
	gitRun(t, "", "init", "-q", "--bare", "-b", "main", fx.origin)
	gitRun(t, "", "clone", "-q", fx.origin, fx.repo)
	writeFile(t, fx.repo+"/"+mpFile, string(table))
	writeFile(t, fx.repo+"/.flow/project.md", pfProject)
	writeFile(t, fx.repo+"/.gitignore", ".setup-ran\n.superpowers/\n")
	writeFile(t, fx.repo+"/spectre/changes/archive/pipeline-fix-old-5/proposal.md", "old\n")
	writeExec(t, fx.repo+"/scripts/foo.sh", "#!/usr/bin/env bash\necho foo\n")
	writeExec(t, fx.repo+"/scripts/test-foo.sh", "#!/usr/bin/env bash\nscripts/foo.sh | grep -q foo\n")
	gitRun(t, fx.repo, "add", "-A")
	gitRun(t, fx.repo, "commit", "-q", "-m", "init")
	gitRun(t, fx.repo, "push", "-q", "origin", "HEAD:main")
	gitRun(t, fx.repo, "remote", "set-head", "origin", "main")
	gitRun(t, "", "clone", "-q", fx.origin, fx.other)
	if rows != "" {
		writeFile(t, fx.rows, rows)
	}
	bin := tmp + "/bin"
	writeExec(t, bin+"/flow", `#!/usr/bin/env bash
printf '%s\n' "$*" >>`+fx.log+`
case "$1 $2" in
  "record dispatches") [ -f `+fx.rows+` ] && exec cat `+fx.rows+`; exit 1 ;;
  "state set") cat >>`+fx.log+`; echo >>`+fx.log+` ;;
esac
`)
	scripts := tcfScriptsDir(t)
	fx.env = Env{Dir: tmp, Getenv: func(k string) string {
		switch k {
		case "FLOW_GUARD_REPO_ROOT":
			return fx.repo
		case "FLOW_GUARD_PROJECT_GET":
			return scripts + "/project-get.sh"
		}
		return pathEnv(bin)(k)
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

// pfIdentity gives the guard's own git calls (a rebase commits) an identity:
// they run under the process environment, so these tests do not run in parallel.
func pfIdentity(t *testing.T) {
	for _, k := range []string{"GIT_AUTHOR_NAME", "GIT_COMMITTER_NAME"} {
		t.Setenv(k, "test")
	}
	for _, k := range []string{"GIT_AUTHOR_EMAIL", "GIT_COMMITTER_EMAIL"} {
		t.Setenv(k, "test@example.com")
	}
}

func TestPipelineFixStart(t *testing.T) {
	pfIdentity(t)
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

func TestPipelineFixStartRefusesANonUniqueK(t *testing.T) {
	pfIdentity(t)
	fx := pfNew(t, "")
	for _, k := range []string{"5", "7"} {
		if rc, out := fx.pf(t, "start", "kan-1-demo", k); rc != 1 || !strings.Contains(out, "<k> is 6") {
			t.Fatalf("k %s: rc %d: %s", k, rc, out)
		}
	}
	gitRun(t, fx.repo, "branch", "pipeline-fix-other-9")
	if rc, out := fx.pf(t, "start", "kan-1-demo", "10"); rc != 0 {
		t.Fatalf("one past a live branch: rc %d: %s", rc, out)
	}
}

func TestPipelineFixEndAndEscalate(t *testing.T) {
	pfIdentity(t)
	fx := pfNew(t, lpRowsJSON("pipeline-fix-6-", lpBranch, "plan completed"))
	gitRun(t, fx.repo, "branch", lpBranch)
	if rc, out := fx.pf(t, "end", "pipeline-fix-6-review-1", "completed"); rc != 2 {
		t.Fatalf("a review closed completed: rc %d: %s", rc, out)
	}
	report := func(body string) string {
		p := filepath.Join(t.TempDir(), "report.md")
		writeFile(t, p, body)
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
	if rc, out := fx.pf(t, "escalate", "pipeline-fix-6-fix-1"); rc != 1 || !strings.Contains(out, "no completed earlier round") {
		t.Fatalf("a first-round escalation: rc %d: %s", rc, out)
	}
	if rc, out := fx.pf(t, "escalate", "pipeline-fix-6-review-2"); rc != 1 || !strings.Contains(out, "admits no escalation") {
		t.Fatalf("a review escalation: rc %d: %s", rc, out)
	}
	if rc, out := fx.pf(t, "escalate", "pipeline-fix-6-plan-2"); rc != 1 || !strings.Contains(out, "admits no escalation") {
		t.Fatalf("a re-plan escalation: rc %d: %s", rc, out)
	}
}

func TestPipelineFixVerify(t *testing.T) {
	pfIdentity(t)
	fx := pfNew(t, "")
	if rc, out := fx.pf(t, "start", "kan-1-demo", "6"); rc != 0 {
		t.Fatal(out)
	}
	wt := fx.wt(lpBranch)
	writeExec(t, wt+"/scripts/foo.sh", "#!/usr/bin/env bash\necho foo bar\n")
	gitRun(t, wt, "commit", "-q", "-am", "fix: foo")
	rc, out := fx.pf(t, "verify", wt)
	if rc != 0 || !strings.Contains(out, "green: true") || !strings.Contains(out, "green: scripts/test-foo.sh") || strings.Contains(out, "cd stats") {
		t.Fatalf("rc %d: %s", rc, out)
	}
	writeExec(t, wt+"/scripts/foo.sh", "#!/usr/bin/env bash\necho bar\n")
	gitRun(t, wt, "commit", "-q", "-am", "fix: break foo")
	if rc, out := fx.pf(t, "verify", wt); rc != 1 || !strings.Contains(out, "red: scripts/test-foo.sh (exit 1)") {
		t.Fatalf("a red harness: rc %d: %s", rc, out)
	}
}

// A touched Go file selects the repotest audits of the whole Go tree, which
// no Go file's name leads to; a touched script alone does not.
func TestPipelineFixChecksGoAudits(t *testing.T) {
	t.Parallel()
	wt := t.TempDir()
	writeFile(t, wt+"/.flow/project.md", pfProject)
	pe0 := pfEnv{projectGet: tcfScriptsDir(t) + "/project-get.sh"}
	if cmds, _ := pfChecks(Env{Getenv: os.Getenv}, pe0, wt, []string{"stats/internal/guard/x.go", "hooks/h.py"}); slices.ContainsFunc(cmds, func(c string) bool { return strings.Contains(c, "internal/repotest") }) {
		t.Fatalf("a worktree without stats/internal/repotest: %q selects it", cmds)
	}
	writeFile(t, wt+"/stats/internal/repotest/doc.go", "package repotest\n")
	audits := "cd stats && go test ./internal/repotest/ -run '^Test(GitConfigPins|FlowAddrDeclaration)$' -count=1"
	pe := pfEnv{projectGet: tcfScriptsDir(t) + "/project-get.sh"}
	env := Env{Getenv: os.Getenv}
	if cmds, rc := pfChecks(env, pe, wt, []string{"stats/internal/guard/x.go"}); rc != 0 || !slices.Contains(cmds, audits) {
		t.Fatalf("a touched Go file: rc %d, %q lacks %s", rc, cmds, audits)
	}
	if cmds, _ := pfChecks(env, pe, wt, []string{"scripts/foo.sh"}); slices.Contains(cmds, audits) {
		t.Fatalf("a touched script alone: %q selects %s", cmds, audits)
	}
}

// A touched script or setup.sh selects the guard package's script audits
// where the worktree has stats/internal/guard (kan-1038).
func TestPipelineFixChecksScriptAudits(t *testing.T) {
	t.Parallel()
	wt := t.TempDir()
	writeFile(t, wt+"/.flow/project.md", pfProject)
	pe := pfEnv{projectGet: tcfScriptsDir(t) + "/project-get.sh"}
	env := Env{Getenv: os.Getenv}
	if cmds, _ := pfChecks(env, pe, wt, []string{"scripts/pipeline-fix.sh"}); slices.Contains(cmds, pfScriptAudits) {
		t.Fatalf("a worktree without stats/internal/guard: %q selects %s", cmds, pfScriptAudits)
	}
	writeFile(t, wt+"/stats/internal/guard/doc.go", "package guard\n")
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
	writeFile(t, wt+"/.flow/project.md", pfProject)
	writeFile(t, wt+"/stats/internal/repotest/doc.go", "package repotest\n")
	pe := pfEnv{projectGet: tcfScriptsDir(t) + "/project-get.sh"}
	want := "cd stats && go test ./internal/repotest/ -count=1"
	if cmds, rc := pfChecks(Env{Getenv: os.Getenv}, pe, wt, []string{"hooks/protect-main-checkout.py"}); rc != 0 || !slices.Contains(cmds, want) {
		t.Fatalf("a touched hook: rc %d, %q lacks %s", rc, cmds, want)
	}
}

func TestPipelineFixLand(t *testing.T) {
	pfIdentity(t)
	reviewed := lpRowsJSON("pipeline-fix-6-", lpBranch, "plan completed", "fix completed", "review-1 clean")
	for _, tc := range []struct {
		name, rows string
		moved      bool
		rc         int
		red, open  bool
		lsTreeFail bool
		refusal    string
	}{
		{name: "a reviewed loop lands", rows: reviewed},
		{name: "a moved base is rebased and verified first", rows: reviewed, moved: true},
		{name: "an unreviewed loop is refused", rows: lpRowsJSON("pipeline-fix-6-", lpBranch, "plan completed", "fix completed"), rc: 1, refusal: "LOOP-LANDING-REFUSED: REVIEW"},
		{name: "a red branch is refused with no base move", rows: reviewed, red: true, rc: 1, refusal: "red: scripts/test-foo.sh"},
		{name: "an unarchived change is refused", rows: reviewed, open: true, rc: 1, refusal: "is not archived"},
		{name: "an ls-tree failure fails closed", rows: reviewed, lsTreeFail: true, rc: 2, refusal: "cannot tell whether"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fx := pfNew(t, tc.rows)
			if rc, out := fx.pf(t, "start", "kan-1-demo", "6"); rc != 0 {
				t.Fatal(out)
			}
			wt := fx.wt(lpBranch)
			writeFile(t, wt+"/docs/fix.md", "fix\n")
			gitRun(t, wt, "add", "docs/fix.md")
			gitRun(t, wt, "commit", "-q", "-m", "docs: fix")
			if tc.red {
				writeExec(t, wt+"/scripts/foo.sh", "#!/usr/bin/env bash\necho bar\n")
				gitRun(t, wt, "commit", "-q", "-am", "fix: break foo")
			}
			if tc.open {
				writeFile(t, wt+"/spectre/changes/"+lpBranch+"/tasks.md", "- [x] 1\n")
				gitRun(t, wt, "add", "-A")
				gitRun(t, wt, "commit", "-q", "-m", "chore(spectre): plan")
			}
			gitRun(t, fx.repo, "config", "--add", "flow.escalation", lpBranch+":pipeline-fix-6-fix-1")
			if tc.lsTreeFail {
				writeExec(t, filepath.Dir(fx.repo)+"/bin/git", "#!/usr/bin/env bash\n"+
					"case \" $* \" in *\" ls-tree \"*\" spectre/changes/"+lpBranch+"\"*) exit 128 ;; esac\n"+
					"exec "+fixtureGit+" \"$@\"\n")
			}
			if tc.moved {
				writeFile(t, fx.other+"/docs/other.md", "other\n")
				gitRun(t, fx.other, "add", "docs/other.md")
				gitRun(t, fx.other, "commit", "-q", "-m", "docs: other")
				gitRun(t, fx.other, "push", "-q", "origin", "HEAD:main")
			}
			rc, out := fx.pf(t, "land", wt)
			if rc != tc.rc {
				t.Fatalf("rc %d, want %d: %s", rc, tc.rc, out)
			}
			_, log := kwRun(fx.env, "", fixtureGit, "--git-dir", fx.origin, "log", "--format=%s", "main")
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
			if _, err := os.Stat(wt); err == nil || lpBranchExists(fx.env, fixtureGit, fx.repo, lpBranch) {
				t.Fatalf("worktree or branch left behind: %s", out)
			}
			if _, out := kwRun(fx.env, "", fixtureGit, "-C", fx.repo, "config", "--get-all", "flow.escalation"); strings.Contains(out, lpBranch) {
				t.Fatalf("the loop's escalation marker outlived its landing: %s", out)
			}
			if !strings.Contains(fx.calls(t), `"state":"FINISHED"`) {
				t.Fatalf("no FINISHED record: %s", fx.calls(t))
			}
		})
	}
}

func TestPipelineFixDrop(t *testing.T) {
	pfIdentity(t)
	fx := pfNew(t, "")
	if rc, out := fx.pf(t, "start", "kan-1-demo", "6"); rc != 0 {
		t.Fatal(out)
	}
	wt := fx.wt(lpBranch)
	writeFile(t, wt+"/scratch.txt", "left\n")
	if rc, out := fx.pf(t, "drop", wt); rc != 0 || !strings.Contains(out, "dropped: "+lpBranch) {
		t.Fatalf("rc %d: %s", rc, out)
	}
	if _, err := os.Stat(wt); err == nil || lpBranchExists(fx.env, fixtureGit, fx.repo, lpBranch) {
		t.Fatal("worktree or branch left behind")
	}
	if !strings.Contains(fx.calls(t), `"withdrawn":true`) {
		t.Fatalf("no withdrawn record: %s", fx.calls(t))
	}
	if rc, out := fx.pf(t, "start", "kan-1-demo", "6"); rc != 1 || !strings.Contains(out, "<k> is 7") {
		t.Fatalf("a dropped loop's <k> was reissued: rc %d: %s", rc, out)
	}
}

func TestPipelineFixEndClearsOwnBranchEscalation(t *testing.T) {
	pfIdentity(t)
	const token = "self-review-kan-1-demo"
	fx := pfNew(t, lpRowsJSON(token+"-", "kan-1-demo", "plan completed"))
	gitRun(t, fx.repo, "config", "--add", "flow.escalation", token+":"+token+"-plan-2")
	p := filepath.Join(t.TempDir(), "report.md")
	writeFile(t, p, "Critical: 0\nImportant: 0\n")
	if rc, out := fx.pf(t, "end", token+"-review-1", p); rc != 0 {
		t.Fatalf("rc %d: %s", rc, out)
	}
	if _, out := kwRun(fx.env, "", fixtureGit, "-C", fx.repo, "config", "--get-all", "flow.escalation"); strings.Contains(out, token) {
		t.Fatalf("the own-branch escalation marker outlived its final clean review: %s", out)
	}
}

// end of the pass closes nothing: check-loop-dispatch admits the pass and
// records no row, so a flow call would answer 404 (kan-1001's self-review).
func TestPipelineFixEndPassRecordsNothing(t *testing.T) {
	const token = "self-review-kan-1-demo"
	fx := pfNew(t, "[]")
	rc, out := fx.pf(t, "end", token+"-pass", "completed")
	if rc != 0 || !strings.Contains(out, "PIPELINE-FIX-END-NA "+token+"-pass") {
		t.Fatalf("rc %d: %s", rc, out)
	}
	if strings.Contains(fx.calls(t), "record dispatch end") {
		t.Fatalf("end of the pass called flow:\n%s", fx.calls(t))
	}
}

// end and escalate of an own-branch self-review in another project run in
// that project -- the caller's working directory -- and verify diffs a
// worktree against its own origin/HEAD, not <agents repo>'s.
func TestPipelineFixOwnBranchInAnotherProject(t *testing.T) {
	pfIdentity(t)
	const token = "self-review-kan-1-demo"
	fx := pfNew(t, lpRowsJSON(token+"-", token, "plan completed", "fix completed"))
	tmp := filepath.Dir(fx.repo)
	porigin, proj := tmp+"/gymie.git", tmp+"/gymie"
	gitRun(t, "", "init", "-q", "--bare", "-b", "develop", porigin)
	gitRun(t, "", "clone", "-q", porigin, proj)
	writeFile(t, proj+"/README.md", "gymie\n")
	gitRun(t, proj, "add", "-A")
	gitRun(t, proj, "commit", "-q", "-m", "init")
	gitRun(t, proj, "push", "-q", "origin", "HEAD:develop")
	gitRun(t, proj, "remote", "set-head", "origin", "develop")
	gitRun(t, proj, "checkout", "-q", "-b", "spectre/kan-1-demo")
	writeFile(t, proj+"/.flow/project.md", "# gymie\n\n## lint\n\n```bash\ntrue\n```\n")
	gitRun(t, proj, "add", "-A")
	gitRun(t, proj, "commit", "-q", "-m", "fix: tooling")
	fx.env.Dir = proj
	if rc, out := fx.pf(t, "end", token+"-fix-1", "completed"); rc != 0 ||
		!strings.Contains(fx.calls(t), "record dispatch end -C "+proj+" -change kan-1-demo -key "+token+"-fix-1") {
		t.Fatalf("end left the caller's project: rc %d: %s\n%s", rc, out, fx.calls(t))
	}
	if rc, out := fx.pf(t, "escalate", token+"-fix-2"); rc != 0 {
		t.Fatalf("escalate: rc %d: %s", rc, out)
	}
	if _, out := kwRun(fx.env, "", fixtureGit, "-C", proj, "config", "--get-all", "flow.escalation"); !strings.Contains(out, token+":"+token+"-fix-2") {
		t.Fatalf("the escalation was not recorded in the caller's project: %q", out)
	}
	lp, _, _, msg, ok := lpFromKey(fx.env, fixtureGit, proj, token+"-fix-2")
	if !ok || !lpEscalated(fx.env, fixtureGit, lp.main, lp, token+"-fix-2") {
		t.Fatalf("check-loop-dispatch does not read the escalation in the caller's project: %s", msg)
	}
	if rc, out := fx.pf(t, "verify", proj); rc != 0 || !strings.Contains(out, "green: true") {
		t.Fatalf("verify against the worktree's own origin/HEAD: rc %d: %s", rc, out)
	}
	p := filepath.Join(t.TempDir(), "report.md")
	writeFile(t, p, "Critical: 0\nImportant: 0\n")
	if rc, out := fx.pf(t, "end", token+"-review-1", p); rc != 0 {
		t.Fatalf("end review clean: rc %d: %s", rc, out)
	}
	if _, out := kwRun(fx.env, "", fixtureGit, "-C", proj, "config", "--get-all", "flow.escalation"); strings.Contains(out, token) {
		t.Fatalf("the escalation outlived its final clean review in the caller's project: %q", out)
	}
}

func TestPipelineFixLandRestartsFlowdWhenStatsMoved(t *testing.T) {
	pfIdentity(t)
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
			writeExec(t, filepath.Dir(fx.repo)+"/bin/make", "#!/usr/bin/env bash\nprintf 'make %s\\n' \"$*\" >>"+fx.log+"\n")
			if rc, out := fx.pf(t, "start", "kan-1-demo", "6"); rc != 0 {
				t.Fatal(out)
			}
			wt := fx.wt(lpBranch)
			for _, p := range tc.paths {
				writeFile(t, wt+"/"+p, "restart:\n\ttrue\n")
			}
			gitRun(t, wt, "add", "-A")
			gitRun(t, wt, "commit", "-q", "-m", "fix: "+tc.paths[0])
			gitRun(t, fx.repo, "config", "--add", "flow.escalation", lpBranch+":pipeline-fix-6-fix-1")
			rc, out := fx.pf(t, "land", wt)
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
	pfIdentity(t)
	fx := pfNew(t, lpRowsJSON("pipeline-fix-6-", lpBranch, "plan completed", "fix completed", "review-1 clean"))
	if rc, out := fx.pf(t, "start", "kan-1-demo", "6"); rc != 0 {
		t.Fatal(out)
	}
	wt := fx.wt(lpBranch)
	b, err := os.ReadFile(wt + "/.flow/project.md")
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, wt+"/.flow/project.md", string(b)+"\n## visual verification\n\n"+
		"| Setting | Value |\n|---------|-------|\n| `ui paths` | `web/src/**` |\n\n"+
		"| Command | Runs |\n|---------|------|\n| `verify` | `test ! -e .visual-red` |\n")
	writeFile(t, wt+"/docs/a.md", "a\n")
	gitRun(t, wt, "add", "-A")
	gitRun(t, wt, "commit", "-q", "-m", "docs: visual section")
	const vis = "test ! -e .visual-red"
	if rc, out := fx.pf(t, "verify", wt); rc != 0 || strings.Contains(out, vis) {
		t.Fatalf("(a) rc %d, no ui path ran the visual row: %s", rc, out)
	}
	writeFile(t, wt+"/web/src/a.ts", "a\n")
	gitRun(t, wt, "add", "-A")
	gitRun(t, wt, "commit", "-q", "-m", "fix: ui")
	if rc, out := fx.pf(t, "verify", wt); rc != 0 || !strings.Contains(out, "green: "+vis) {
		t.Fatalf("(b) rc %d: %s", rc, out)
	}
	writeFile(t, wt+"/.visual-red", "x\n")
	if rc, out := fx.pf(t, "verify", wt); rc != 1 || !strings.Contains(out, "red: "+vis+" (exit 1)") {
		t.Fatalf("(c) rc %d: %s", rc, out)
	}
	gitRun(t, fx.repo, "config", "--add", "flow.escalation", lpBranch+":pipeline-fix-6-fix-1")
	rc, out := fx.pf(t, "land", wt)
	_, log := kwRun(fx.env, "", fixtureGit, "--git-dir", fx.origin, "log", "--format=%s", "main")
	if rc != 1 || !strings.Contains(out, "red: "+vis) || strings.Contains(log, "fix: ui") {
		t.Fatalf("(d) rc %d: %s\n%s", rc, out, log)
	}
}

func TestPipelineFixVerifyFillsVisualSpecs(t *testing.T) {
	pfIdentity(t)
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
	writeFile(t, wt+"/.flow/project.md", section("| `ui paths` | `web/src/**` |\n"))
	writeFile(t, wt+"/web/src/a.ts", "a\n")
	gitRun(t, wt, "add", "-A")
	gitRun(t, wt, "commit", "-q", "-m", "fix: ui")
	if rc, out := fx.pf(t, "verify", wt); rc != 0 || !strings.Contains(out, `green: test "web/a.spec.ts" = web/a.spec.ts`) {
		t.Fatalf("(a) rc %d, specs paths not substituted: %s", rc, out)
	}
	// No `ui paths` row: the trigger cannot answer, no merge base is read,
	// so a specs row has nothing to scope to and the visual line is red.
	writeFile(t, wt+"/.flow/project.md", section(""))
	if rc, out := fx.pf(t, "verify", wt); rc != 1 || !strings.Contains(out, "red: visual verify — no merge base to scope the specs row to") {
		t.Fatalf("(b) rc %d: %s", rc, out)
	}
}
