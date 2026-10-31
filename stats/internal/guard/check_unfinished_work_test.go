package guard

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Every case of scripts/test-check-unfinished-work.sh at 3e48ecac, one
// subtest per ok: label; a label's scenario is built afresh in its own
// subtest. The harness's stub `flow` answered `record findings`, `record
// verdict`, `record verdicts` and `state find`; here the findings read is
// Env.Findings, except where a case asserts on the exec itself (the -C
// anchor, the stderr split), and every other call stays one stub `flow` on
// Env's PATH. That stub logs each call's arguments to bin/<verb>.args,
// answers from bin/<verb>.json, fails like an unreachable store when
// bin/<verb>.fail exists, and prints a FLOW_ADDR diagnostic on stderr when
// bin/<verb>.diag does -- one body for every fixture, so one assessed inode
// (writeExec).
const uwStub = `#!/usr/bin/env bash
d="${0%/*}"
v="$2"
[ "$1" = state ] && v=find
printf '%s\n' "$*" >> "$d/$v.args"
if [ -e "$d/$v.fail" ]; then echo "flow: connect: connection refused" >&2; exit 1; fi
[ -e "$d/$v.diag" ] && echo "flow: using FLOW_ADDR=http://127.0.0.1:4174" >&2
[ "$v" = verdict ] || cat "$d/$v.json" 2>/dev/null
exit 0
`

// uwFix is the harness's new_fixture: a worktree holding a finished change
// "demo" (every plan item checked, one fixed finding) and a stub `flow` in
// <wt>/bin.
type uwFix struct {
	wt, bin  string
	env      Env
	findings []byte
	fErr     error
	ranges   []string // <worktree>=<merge-base> arguments, after an empty canonical
}

// commits makes dir a repository whose base commit carries baseMsg and
// whose one later commit carries msg -- both empty, so the fixture's files
// stay untracked -- and returns `<dir>=<base sha>`, a range argument.
func uwCommits(t *testing.T, dir, baseMsg, msg string) string {
	t.Helper()
	gitRun(t, "", "init", "-q", dir)
	gitRun(t, dir, "commit", "-q", "--allow-empty", "-m", baseMsg)
	base := tcfGit(t, dir, "rev-parse", "HEAD")
	gitRun(t, dir, "commit", "-q", "--allow-empty", "-m", msg)
	return dir + "=" + base
}

// uwResult is one run: exit code, stdout (trailing newlines stripped, as
// command substitution does), stderr, and the paths its checks read.
type uwResult struct {
	rc             int
	out, err       string
	bin, wt, canon string
}

func newUWBare(t *testing.T, wt string) *uwFix {
	t.Helper()
	f := &uwFix{wt: wt, bin: wt + "/bin"}
	writeExec(t, f.bin+"/flow", uwStub)
	f.setFile(t, "verdicts.json", "[]\n")
	f.setFile(t, "dispatches.json", `[{"key":"task-1-implementer","outcome":"completed"}]`+"\n")
	// A pass row for every round a fixture's panel rows run in (0..2).
	f.setFile(t, "passes.json", `[{"id":1,"round":0,"note":"roster: full"},{"id":2,"round":1,"note":"n"},{"id":3,"round":2,"note":"n"}]`+"\n")
	// A fix-mutation row for the round fixing the fixture's round-0 findings.
	f.setFile(t, "mutations.json", `[{"id":1,"round":1,"path":"a","mutated":"none","test":"doc-only"}]`+"\n")
	// A recorded change summary: the bundle header signal six reads.
	f.setFile(t, "bundle.json", "# Self-review context bundle for demo\n\nfound: 1 of 1 sources; skipped: 0 of 1 sources\n\n## change summary\n\nwhat changed\n")
	f.setFindings(t)
	get := pathEnv(f.bin)
	f.env = Env{Getenv: get, Dir: wt,
		Findings: func(string) ([]byte, error) { return f.findings, f.fErr }}
	return f
}

func newUW(t *testing.T) *uwFix {
	t.Helper()
	f := newUWBare(t, t.TempDir())
	f.plan(t, "- [x] 1.1 done\n")
	f.setFindings(t, "fixed")
	return f
}

// newUWSatellite is new_satellite_fixture: a link-only change "sat-demo"
// whose link.md names <peer>:canon-demo under ## Part of, no local plan, and
// no findings.
func newUWSatellite(t *testing.T, wt, peer string) *uwFix {
	t.Helper()
	f := newUWBare(t, wt)
	writeFile(t, wt+"/spectre/changes/sat-demo/link.md", "## Part of\n\n`"+peer+":canon-demo`\n")
	return f
}

// uwCanonTree is new_canonical_tree: <parent>/canon-tree carrying
// canon-demo's plan with one item in the given state.
func uwCanonTree(t *testing.T, parent, line string) string {
	t.Helper()
	dir := parent + "/canon-tree"
	writeFile(t, dir+"/spectre/changes/canon-demo/tasks.md", "- "+line+"\n")
	return dir
}

func (f *uwFix) plan(t *testing.T, body string) {
	t.Helper()
	writeFile(t, f.wt+"/spectre/changes/demo/tasks.md", body)
}

func (f *uwFix) setFile(t *testing.T, name, body string) {
	t.Helper()
	writeFile(t, f.bin+"/"+name, body)
}

// setFindings is set_findings: refs F1..Fn, one per status, written both as
// Env.Findings' answer and as the stub's findings.json.
func (f *uwFix) setFindings(t *testing.T, statuses ...string) {
	t.Helper()
	type finding struct {
		Ref        string `json:"ref"`
		Status     string `json:"status"`
		Reproducer string `json:"reproducer"`
	}
	all := []finding{}
	for i, s := range statuses {
		all = append(all, finding{"F" + strconv.Itoa(i+1), s, "scripts/x.sh"})
	}
	b, err := json.Marshal(all)
	if err != nil {
		t.Fatal(err)
	}
	f.findings = b
	f.setFile(t, "findings.json", string(b)+"\n")
}

func (f *uwFix) run(t *testing.T, args ...string) uwResult {
	t.Helper()
	fn := Registry["check-unfinished-work"]
	if fn == nil {
		t.Fatal("check-unfinished-work is not registered")
	}
	var out, errb bytes.Buffer
	rc := fn(args, f.env, &out, &errb)
	return uwResult{rc: rc, out: strings.TrimRight(out.String(), "\n"), err: errb.String(), bin: f.bin, wt: f.wt}
}

func (r uwResult) args(verb string) string {
	b, _ := os.ReadFile(r.bin + "/" + verb + ".args")
	return string(b)
}

// anchored is `grep -q -- "-C <dir>$" <verb>.args`.
func (r uwResult) anchored(verb, dir string) bool {
	for _, l := range strings.Split(r.args(verb), "\n") {
		if strings.HasSuffix(l, "-C "+dir) {
			return true
		}
	}
	return false
}

// uwCheck is one ok: label and what it asserts of its scenario's run.
type uwCheck struct {
	label string
	check func(t *testing.T, r uwResult)
}

func uwVerdict(label, want string) uwCheck {
	return uwCheck{label, func(t *testing.T, r uwResult) {
		t.Helper()
		switch {
		case r.rc != 0:
			t.Fatalf("expected exit 0, got rc=%d out=%s err=%s", r.rc, r.out, r.err)
		case r.out == "":
			t.Fatalf("expected a verdict line, got empty stdout (err=%s)", r.err)
		case strings.Contains(r.out, "\n"):
			t.Fatalf("expected exactly one stdout line: %s", r.out)
		case !strings.HasPrefix(r.out, want):
			t.Fatalf("expected a line beginning %s, got: %s", want, r.out)
		}
	}}
}

func uwReason(label, needle string) uwCheck {
	return uwCheck{label, func(t *testing.T, r uwResult) {
		if !strings.Contains(r.out, needle) {
			t.Fatalf("the breakdown does not name the signal (%s): %s", needle, r.out)
		}
	}}
}

func uwRC(label string, want int) uwCheck {
	return uwCheck{label, func(t *testing.T, r uwResult) {
		if r.rc != want {
			t.Fatalf("expected exit %d, got rc=%d out=%s err=%s", want, r.rc, r.out, r.err)
		}
	}}
}

func uwNonZero(label string) uwCheck {
	return uwCheck{label, func(t *testing.T, r uwResult) {
		if r.rc == 0 {
			t.Fatalf("expected a non-zero exit, got 0 out=%s", r.out)
		}
	}}
}

func uwNoOut(label string) uwCheck {
	return uwCheck{label, func(t *testing.T, r uwResult) {
		if r.out != "" {
			t.Fatalf("emitted a verdict line: %s", r.out)
		}
	}}
}

func uwErr(label, needle string) uwCheck {
	return uwCheck{label, func(t *testing.T, r uwResult) {
		if !strings.Contains(r.err, needle) {
			t.Fatalf("no %q on stderr: %s", needle, r.err)
		}
	}}
}

func uwTrue(label string, ok func(r uwResult) bool) uwCheck {
	return uwCheck{label, func(t *testing.T, r uwResult) {
		if !ok(r) {
			t.Fatalf("rc=%d out=%s err=%s\nfindings.args=%s\nverdict.args=%s\nfind.args=%s",
				r.rc, r.out, r.err, r.args("findings"), r.args("verdict"), r.args("find"))
		}
	}}
}

func TestCheckUnfinishedWork(t *testing.T) {
	t.Parallel()
	// each runs scenario afresh inside every check's own subtest.
	each := func(scenario func(t *testing.T) uwResult, checks ...uwCheck) {
		for _, c := range checks {
			t.Run(c.label, func(t *testing.T) {
				t.Parallel()
				c.check(t, scenario(t))
			})
		}
	}
	// demo is new_fixture, then setup, then the guard on (wt, demo).
	demo := func(setup func(t *testing.T, f *uwFix)) func(t *testing.T) uwResult {
		return func(t *testing.T) uwResult {
			f := newUW(t)
			if setup != nil {
				setup(t, f)
			}
			if len(f.ranges) > 0 {
				return f.run(t, append([]string{f.wt, "demo", ""}, f.ranges...)...)
			}
			return f.run(t, f.wt, "demo")
		}
	}
	unticked := func(t *testing.T, f *uwFix) { f.plan(t, "- [ ] 1.1 not done\n") }
	findings := func(statuses ...string) func(t *testing.T, f *uwFix) {
		return func(t *testing.T, f *uwFix) { f.setFindings(t, statuses...) }
	}
	file := func(rel, body string) func(t *testing.T, f *uwFix) {
		return func(t *testing.T, f *uwFix) { writeFile(t, f.wt+"/"+rel, body) }
	}

	// 1. The whole point of the CLEAR verdict: a finished change is not interrupted.
	each(demo(nil), uwVerdict("a finished change is CLEAR", "CLEAR:"))
	// 1b. The other closed statuses are closed, not open.
	each(demo(findings("fixed", "withdrawn retracted, the guard already covers it", "fixed")),
		uwVerdict("fixed and withdrawn findings are closed, not open", "CLEAR:"))
	// 2. docs/manual-test/ is not a signal at all, even when a leftover guide
	//    is still in the worktree when the guard runs.
	each(demo(file("docs/manual-test/guide.md", "a leftover guide\n")),
		uwVerdict("a leftover docs/manual-test/ present at invocation time is not a signal", "CLEAR:"))

	// 2b. Signal three — a dispatch row never closed with an outcome
	//     (kan-929's three "Outcome: not recorded" rows reached integrate).
	dispatches := func(body string) func(t *testing.T, f *uwFix) {
		return func(t *testing.T, f *uwFix) { f.setFile(t, "dispatches.json", body+"\n") }
	}
	each(demo(dispatches(`[{"key":"task-1-implementer","outcome":"completed","sessionToken":"mf-run1"},{"key":"task-14-reviewer","sessionToken":"mf-run2"}]`)),
		uwVerdict("a dispatch row with no outcome is OUTSTANDING", "OUTSTANDING:"),
		uwReason("the breakdown names the unclosed row's key", "1 dispatch row(s) with no outcome (task-14-reviewer)"),
		uwReason("the close command carries the row's own key and session token", "flow record dispatch end -change demo -key task-14-reviewer -session-token mf-run2 -outcome <outcome> -C "))
	each(demo(dispatches(`[{"seq":7,"key":"task-9-implementer"}]`)),
		uwVerdict("a tokenless open row is still OUTSTANDING", "OUTSTANDING:"),
		uwReason("a tokenless row is named as not closable instead of a broken command", "row seq 7 has no key or session token and cannot be closed by `flow record dispatch end`"))
	each(demo(dispatches("not json")), uwRC("unreadable dispatch rows cannot answer", 2))
	each(demo(file("bin/dispatches.fail", "")), uwRC("a failed dispatches read cannot answer", 2))
	// 2c. Signal four — a ticked appended task no dispatch row names
	//     (kan-979's fix-round task reached integrate with no rows at all).
	planned := func(body, rows string) func(t *testing.T, f *uwFix) {
		return func(t *testing.T, f *uwFix) {
			f.plan(t, body)
			f.setFile(t, "dispatches.json", rows+"\n")
			f.ranges = []string{uwCommits(t, f.wt, "base", "fix: the appended task\n\nTask-Id: 7")}
		}
	}
	// covered is a ticked appended task 7 with its implementer row and a
	// later panel reviewer row: only its commit is left to judge.
	covered := `[{"seq":1,"key":"task-7-implementer","role":"implementer","outcome":"completed"},{"seq":2,"key":"panel-2-primary","role":"reviewer","outcome":"clean"}]`
	committedAs := func(baseMsg, msg string, peer bool) func(t *testing.T, f *uwFix) {
		return func(t *testing.T, f *uwFix) {
			f.plan(t, "> **Tasks appended:** 1\n\n- [x] 1. done\n- [x] 7. appended by a fix run\n")
			f.setFile(t, "dispatches.json", covered+"\n")
			if peer {
				// A peer the merge-order record names, so signal ten is satisfied.
				writeFile(t, f.wt+"/spectre/peers", "peer ../peer\n")
				writeFile(t, f.wt+"/spectre/changes/demo/link.md", "## Merge order\n\n1. `.`\n2. `peer`\n")
				f.ranges = []string{uwCommits(t, f.wt, "base", "fix: no trailer"), uwCommits(t, filepath.Dir(f.wt)+"/peer", "base", msg)}
				return
			}
			f.ranges = []string{uwCommits(t, f.wt, baseMsg, msg)}
		}
	}
	appended := "> **Tasks appended:** 1\n\n- [x] 1. done\n- [x] 2. grouped\n- [x] 7. appended by a fix run\n"
	each(demo(planned(appended, `[{"key":"task-1-implementer","outcome":"completed"}]`)),
		uwVerdict("an appended task with no implementer dispatch row is OUTSTANDING", "OUTSTANDING:"),
		uwReason("the breakdown names the undispatched appended task", "1 ticked appended task(s) with no dispatch row (Task 7)"))
	each(demo(planned(appended, `[{"seq":1,"key":"task-1-implementer","role":"implementer","outcome":"completed"},{"seq":2,"key":"task-7-implementer","role":"implementer","outcome":"completed"},{"seq":3,"key":"panel-2-primary","role":"reviewer","outcome":"clean"}]`)),
		uwVerdict("an appended task with its implementer row and a later reviewer row is CLEAR", "CLEAR:"))
	each(demo(planned(appended, `[{"seq":1,"key":"impl-7","taskId":"7","role":"implementer","outcome":"completed"},{"seq":2,"key":"panel-2-primary","role":"reviewer","outcome":"clean"}]`)),
		uwVerdict("a row whose taskId names the appended task covers it", "CLEAR:"))
	each(demo(planned(appended, `[{"seq":1,"key":"task-6+7-implementer-fix-1","role":"implementer","outcome":"completed"},{"seq":2,"key":"panel-2-primary","role":"reviewer","outcome":"clean"}]`)),
		uwVerdict("a +-joined key covers the appended task it names", "CLEAR:"))
	each(demo(planned(appended, `[{"seq":1,"key":"task-7-implementer","role":"implementer","outcome":"completed"}]`)),
		uwVerdict("an appended task with an implementer row and no reviewer row is OUTSTANDING", "OUTSTANDING:"),
		uwReason("the breakdown names the unreviewed appended task", "1 ticked appended task(s) with no panel reviewer row (key `panel-<round>-<slot>`) begun after its implementer row (Task 7)"))
	each(demo(planned(appended, `[{"seq":1,"key":"panel-1-primary","role":"reviewer","outcome":"clean"},{"seq":2,"key":"task-7-implementer","role":"implementer","outcome":"completed"}]`)),
		uwVerdict("a reviewer row begun before the implementer row does not review the appended task", "OUTSTANDING:"))
	each(demo(planned(appended, `[{"seq":1,"key":"task-7-implementer","role":"implementer","outcome":"completed"},{"seq":2,"key":"pipeline-fix-1-review-1","role":"reviewer","outcome":"clean"}]`)),
		uwVerdict("a non-panel reviewer row does not review the appended task", "OUTSTANDING:"))
	// An in-run fix task (verify-fix-loop.md step 1) never raises
	// **Tasks appended:**, and is judged all the same (kan-941's tasks 12-16).
	inRun := "- [x] 1. done\n- [x] 7. In-run fix 1 — flow.visual-verify: the menu shadow\n"
	each(demo(planned(inRun, `[{"seq":1,"key":"task-1-implementer","role":"implementer","outcome":"completed"}]`)),
		uwVerdict("an in-run fix task with no dispatch row is OUTSTANDING", "OUTSTANDING:"),
		uwReason("the breakdown names the undispatched in-run fix task", "1 ticked appended task(s) with no dispatch row (Task 7)"))
	each(demo(planned(inRun, `[{"seq":1,"key":"task-7-implementer","role":"implementer","outcome":"completed"},{"seq":2,"key":"panel-2-primary","role":"reviewer","outcome":"clean"}]`)),
		uwVerdict("an in-run fix task with its implementer row and a later reviewer row is CLEAR", "CLEAR:"))
	each(demo(planned("- [x] 1. done\n- [x] 2. grouped\n- [x] 3. grouped\n", `[{"key":"task-2-implementer","outcome":"completed"}]`)),
		uwVerdict("plan-time tasks grouped under one key, with no Tasks appended line, are CLEAR", "CLEAR:"))
	// 2e. Signal five — a panel round dispatched with no pass row (kan-939:
	//     round 1 ran panel-1-primary+principles and recorded nothing).
	withPasses := func(rows, passes string) func(t *testing.T, f *uwFix) {
		return func(t *testing.T, f *uwFix) {
			f.setFile(t, "dispatches.json", rows+"\n")
			f.setFile(t, "passes.json", passes+"\n")
		}
	}
	panelRows := `[{"seq":1,"key":"panel-0-primary+principles","role":"reviewer","outcome":"clean"},{"seq":2,"key":"panel-fix-1","role":"panel-fix","outcome":"completed"},{"seq":3,"key":"panel-1-primary+principles","role":"reviewer","outcome":"clean"},{"seq":4,"key":"task-1-reviewer","role":"reviewer","outcome":"clean"}]`
	each(demo(withPasses(panelRows, `[{"id":1,"round":0,"note":"roster: full"}]`)),
		uwVerdict("a panel round dispatched with no pass row is OUTSTANDING", "OUTSTANDING:"),
		uwReason("the breakdown names the round and its keys", "1 panel round(s) dispatched with no pass row (round 1: panel-1-primary+principles)"))
	each(demo(withPasses(panelRows, `[{"id":1,"round":0,"note":"roster: full"},{"id":2,"round":1,"note":"n"}]`)),
		uwVerdict("every dispatched panel round with a pass row is CLEAR", "CLEAR:"))
	each(demo(withPasses(`[{"seq":1,"key":"task-1-reviewer","role":"reviewer","outcome":"clean"},{"seq":2,"key":"panel-fix-3","role":"panel-fix","outcome":"completed"}]`, "not json")),
		uwVerdict("no panel reviewer row reads no pass rows", "CLEAR:"))
	each(demo(withPasses(panelRows, "not json")), uwRC("unreadable pass rows cannot answer", 2))
	each(demo(func(t *testing.T, f *uwFix) {
		withPasses(panelRows, "[]")(t, f)
		f.setFile(t, "passes.fail", "")
	}), uwRC("a failed pass-row read cannot answer", 2))

	// 2f. Signal six — no change summary recorded (kan-957: 35 of 36 gymie
	//     changes landed with none).
	bundle := func(body string) func(t *testing.T, f *uwFix) {
		return func(t *testing.T, f *uwFix) { f.setFile(t, "bundle.json", body) }
	}
	each(demo(bundle("# Self-review context bundle for demo\n\nfound: 0 of 1 sources; skipped: 1 of 1 sources\nskipped: change summary (absent)\n\n")),
		uwVerdict("a change with no recorded summary is OUTSTANDING", "OUTSTANDING:"),
		uwReason("the breakdown names the summary and its record command", "no change summary recorded — record the run's summary with `flow record summary -change demo -C "))
	each(demo(bundle("# Self-review context bundle for demo\n\nfound: 1 of 1 sources; skipped: 0 of 1 sources\n\n## change summary\n\nskipped: change summary (absent)\n")),
		uwVerdict("the skipped line inside a recorded summary's text is not the header's", "CLEAR:"))
	each(demo(bundle("not a bundle\n")), uwRC("a bundle with no found: line cannot answer", 2))
	each(demo(func(t *testing.T, f *uwFix) { f.setFile(t, "bundle.fail", "") }),
		uwRC("a failed bundle read cannot answer", 2))

	// 2d. Signal four's commit half — a dispatch row is not a commit (a
	//     back-filled row named a task whose work never landed).
	each(demo(committedAs("base", "fix: the appended task\n\nFlow-Key: panel-fix-2", false)),
		uwVerdict("an appended task with its rows but no Task-Id commit is OUTSTANDING", "OUTSTANDING:"),
		uwReason("the breakdown names the uncommitted appended task", "1 ticked appended task(s) with no commit carrying its `Task-Id:` trailer on the branch (Task 7)"))
	each(demo(committedAs("fix: before the branch\n\nTask-Id: 7", "fix: no trailer", false)),
		uwVerdict("a Task-Id commit at or before the merge base does not count", "OUTSTANDING:"))
	each(demo(committedAs("base", "fix: the appended task\n\nTask-Id: 7", true)),
		uwVerdict("a Task-Id commit in a peer worktree's range covers the appended task", "CLEAR:"))
	each(func(t *testing.T) uwResult {
		f := newUW(t)
		f.plan(t, "> **Tasks appended:** 1\n\n- [x] 1. done\n- [x] 7. appended by a fix run\n")
		f.setFile(t, "dispatches.json", covered+"\n")
		gitRun(t, "", "init", "-q", f.wt)
		gitRun(t, f.wt, "add", "spectre/changes/demo/tasks.md")
		gitRun(t, f.wt, "commit", "-q", "-m", "chore(spectre): drop demo")
		base := tcfGit(t, f.wt, "rev-parse", "HEAD")
		gitRun(t, f.wt, "commit", "-q", "--allow-empty", "-m", "fix: no trailer")
		return f.run(t, f.wt, "demo", "", f.wt+"="+base)
	}, uwVerdict("an appended task already ticked in the plan at a dropped re-run's base needs no new commit", "CLEAR:"))
	each(func(t *testing.T) uwResult {
		f := newUW(t)
		f.plan(t, "> **Tasks appended:** 1\n\n- [x] 7. appended by a fix run\n")
		f.setFile(t, "dispatches.json", covered+"\n")
		return f.run(t, f.wt, "demo")
	}, uwRC("ticked appended tasks with no range passed cannot answer", 2))
	each(demo(nil), uwVerdict("a plan with no appended task needs no range", "CLEAR:"))
	each(func(t *testing.T) uwResult {
		f := newUW(t)
		return f.run(t, f.wt, "demo", "", "no-equals-sign")
	}, uwRC("a range argument not shaped <worktree>=<merge-base> is refused", 2))

	// 2d. Signal seven and the folded visual verdict (kan-989 self-review):
	//     a missing range is derived, a code commit after the last completed
	//     verify row is unfinished, and a ui-path commit no visual verifier
	//     saw is too.
	const (
		before = "2026-01-01T10:00:00Z"
		after  = "2026-01-01T12:00:00Z"
	)
	verifyRow := func(at string) string {
		return `{"seq":9,"key":"verify","role":"verifier","outcome":"completed","startedAt":"` + at + `"}`
	}
	// coded makes f.wt a repository whose base commit is empty and whose one
	// later commit, dated 11:00, touches path; f.ranges is its range.
	coded := func(path string) func(t *testing.T, f *uwFix) {
		return func(t *testing.T, f *uwFix) {
			gitRun(t, "", "init", "-q", f.wt)
			gitRun(t, f.wt, "commit", "-q", "--allow-empty", "-m", "base")
			base := tcfGit(t, f.wt, "rev-parse", "HEAD")
			writeFile(t, f.wt+"/"+path, "x\n")
			gitRun(t, f.wt, "add", path)
			gitRun(t, f.wt, "commit", "-q", "--date=2026-01-01T11:00:00Z", "-m", "fix: code")
			f.ranges = []string{f.wt + "=" + base}
		}
	}
	verified := func(path, at string) func(t *testing.T, f *uwFix) {
		return func(t *testing.T, f *uwFix) {
			coded(path)(t, f)
			f.setFile(t, "dispatches.json", "["+verifyRow(at)+"]\n")
		}
	}
	each(func(t *testing.T) uwResult {
		f := newUW(t)
		f.plan(t, "> **Tasks appended:** 1\n\n- [x] 1. done\n- [x] 7. appended by a fix run\n")
		f.setFile(t, "dispatches.json", `[{"seq":1,"key":"task-7-implementer","role":"implementer","outcome":"completed"},{"seq":2,"key":"panel-1-primary","role":"reviewer","outcome":"clean"},`+verifyRow(after)+"]\n")
		origin := t.TempDir() + "/origin.git"
		gitRun(t, "", "init", "-q", "--bare", "-b", "main", origin)
		gitRun(t, "", "init", "-q", "-b", "main", f.wt)
		gitRun(t, f.wt, "commit", "-q", "--allow-empty", "-m", "base")
		gitRun(t, f.wt, "remote", "add", "origin", origin)
		gitRun(t, f.wt, "push", "-q", "origin", "main")
		gitRun(t, f.wt, "checkout", "-q", "-b", "feature")
		writeFile(t, f.wt+"/src/a.go", "package a\n")
		gitRun(t, f.wt, "add", "src/a.go")
		gitRun(t, f.wt, "commit", "-q", "--date=2026-01-01T11:00:00Z", "-m", "fix: the appended task\n\nTask-Id: 7")
		return f.run(t, f.wt, "demo")
	}, uwVerdict("a call with no range derives the worktree's own merge base", "CLEAR:"))
	each(demo(verified("src/a.go", before)),
		uwVerdict("a code commit after the last completed verify row is OUTSTANDING", "OUTSTANDING:"),
		uwReason("the breakdown names the code commit no verify row covers", "no completed inline verify row since"))
	each(demo(verified("src/a.go", after)),
		uwVerdict("a completed verify row after the last code commit is CLEAR", "CLEAR:"))
	each(demo(func(t *testing.T, f *uwFix) {
		verified("src/a.go", after)(t, f)
		writeFile(t, f.wt+"/src/b.go", "y\n")
		gitRun(t, f.wt, "add", "src/b.go")
		gitRun(t, f.wt, "commit", "-q", "--date=2026-01-01T13:00:00Z", "-m", "feat: reshaped\n\nFlow-Key: flow.commit-two")
	}),
		uwVerdict("run 1's reshaped flow.commit-two commit after the verify row is CLEAR", "CLEAR:"))
	each(demo(verified("spectre/changes/demo/tasks.md", before)),
		uwVerdict("planning-only commits after the verify row are CLEAR", "CLEAR:"))
	uiCommit := func(t *testing.T, f *uwFix) {
		writeFile(t, f.wt+"/.flow/project.md", "# project\n\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `src/web/**` |\n")
		f.setFile(t, "decisions.json", "[]\n")
		verified("src/web/App.tsx", after)(t, f)
	}
	each(demo(uiCommit),
		uwVerdict("a ui-path commit no visual verifier saw is OUTSTANDING", "OUTSTANDING:"),
		uwReason("the breakdown carries the visual guard's VISUAL-VERIFY-MISSING line", "VISUAL-VERIFY-MISSING"))
	each(demo(func(t *testing.T, f *uwFix) {
		uiCommit(t, f)
		wt, _ := filepath.EvalSymlinks(f.wt)
		out, _ := exec.Command(fixtureGit, "-C", f.wt, "rev-parse", "HEAD").Output()
		head := strings.TrimSpace(string(out))
		f.setFile(t, "verdicts.json", `[{"change":"demo","guard":"check-visual-verify-dispatched","worktree":"`+wt+`","verdict":"VISUAL-VERIFY-MISSING: x — HEAD `+head+`","falsePositive":true}]`+"\n")
	}),
		uwVerdict("a ui-path commit whose MISSING verdict an operator flagged a false positive for this HEAD is CLEAR", "CLEAR:"))
	each(demo(func(t *testing.T, f *uwFix) {
		uiCommit(t, f)
		wt, _ := filepath.EvalSymlinks(f.wt)
		out, _ := exec.Command(fixtureGit, "-C", f.wt, "rev-parse", "HEAD").Output()
		head := strings.TrimSpace(string(out))
		f.setFile(t, "verdicts.json", `[{"change":"demo","guard":"check-visual-verify-dispatched","worktree":"`+wt+`","verdict":"VISUAL-VERIFY-MISSING: x HEAD `+head+`","falsePositive":true}]`+"\n")
	}),
		uwVerdict("a flagged verdict ending HEAD <sha> without the \" — \" separator integrate requires is still OUTSTANDING", "OUTSTANDING:"))

	// 3. Signal one — the plan.
	each(demo(unticked),
		uwVerdict("an unchecked plan item is OUTSTANDING", "OUTSTANDING:"),
		uwReason("an unchecked plan item names its signal", "unchecked plan item"))
	// 3b. The same signal in a fix sub-change, both layouts.
	each(demo(file("spectre/changes/demo-fix-1/tasks.md", "- [ ] 1.1 the fix is not done\n")),
		uwVerdict("an unchecked item in a sibling fix sub-change is OUTSTANDING", "OUTSTANDING:"))
	each(demo(file("spectre/changes/demo/demo-fix-1/tasks.md", "- [ ] 1.1 the fix is not done\n")),
		uwVerdict("an unchecked item in a nested fix sub-change is OUTSTANDING", "OUTSTANDING:"))
	// 3c. A fix OF a fix, nested inside the change and beside it.
	each(demo(file("spectre/changes/demo/demo-fix-1/demo-fix-1-fix-2/tasks.md", "- [ ] 1.1 the fix of the fix is not done\n")),
		uwVerdict("a plan two levels deep under the change is OUTSTANDING", "OUTSTANDING:"),
		uwReason("a plan two levels deep names its signal", "unchecked plan item"))
	each(demo(file("spectre/changes/demo-fix-1-fix-2/tasks.md", "- [ ] 1.1 the sibling fix of the fix is not done\n")),
		uwVerdict("a sibling fix-of-a-fix plan is OUTSTANDING", "OUTSTANDING:"))
	// 3d. Another change's plan is not this change's signal.
	each(demo(file("spectre/changes/other-change/tasks.md", "- [ ] 1.1 not our problem\n")),
		uwVerdict("another change's unchecked plan is not this change's signal", "CLEAR:"))
	// 3e. The PRIMARY plan is not optional, and its absence is not silence.
	rmAll := func(rel string) func(t *testing.T, f *uwFix) {
		return func(t *testing.T, f *uwFix) {
			if err := os.RemoveAll(f.wt + "/" + rel); err != nil {
				t.Fatal(err)
			}
		}
	}
	each(demo(rmAll("spectre/changes/demo")),
		uwVerdict("a missing primary plan is OUTSTANDING, not CLEAR", "OUTSTANDING:"),
		uwReason("a missing primary plan names its signal", "no plan at"))
	each(demo(rmAll("spectre")),
		uwVerdict("no spectre tree at all is OUTSTANDING, not CLEAR", "OUTSTANDING:"),
		uwReason("a missing spectre tree names the plan signal", "no plan at"))

	// 4. Signal two — findings whose recorded status is not closed.
	each(demo(findings("open")),
		uwVerdict("an open finding is OUTSTANDING", "OUTSTANDING:"),
		uwReason("an open finding names its signal, with its count", "1 open finding(s)"))
	each(demo(findings("open", "fixed", "open", "open")),
		uwVerdict("several open findings are OUTSTANDING", "OUTSTANDING:"),
		uwReason("every open finding is counted, not just the first", "3 open finding(s)"))
	each(demo(findings("withdrawn the operator retracted it: the guard already covers this")),
		uwVerdict("a withdrawal with a reason is closed", "CLEAR:"))
	each(demo(findings("withdrawn")),
		uwVerdict("a bare withdrawn finding is open (KAN-791)", "OUTSTANDING:"),
		uwReason("a bare withdrawn finding is counted open", "1 open finding(s)"))
	each(demo(findings("withdrawn   ")),
		uwVerdict("a whitespace-only withdrawal reason is open", "OUTSTANDING:"),
		uwReason("a whitespace-only withdrawal reason is counted open", "1 open finding(s)"))
	each(demo(findings("deferred cosmetic, not worth a fix round")),
		uwVerdict("a deferred finding is open", "OUTSTANDING:"),
		uwReason("a deferred finding is counted open", "1 open finding(s)"))
	each(demo(findings()), uwVerdict("a store with no findings for this run is CLEAR", "CLEAR:"))
	// 4e. The store is unreachable: exit 2, never 0 or 1.
	each(demo(func(t *testing.T, f *uwFix) {
		f.findings, f.fErr = []byte("flow: connect: connection refused"), errors.New("exit status 1")
	}),
		uwRC("an unreachable store exits 2, never 0 or 1", 2),
		uwNoOut("an unreachable store emits no verdict line"),
		uwErr("an unreachable store names the failure on stderr", "cannot read findings for 'demo' from the store — cannot determine anything"))
	// 4e, exec path: production never sets Env.Findings, so the refusal the
	// shim actually runs is the failed `flow record findings` exec.
	each(demo(func(t *testing.T, f *uwFix) {
		f.env.Findings = nil
		f.setFile(t, "findings.fail", "")
	}),
		uwRC("an unreachable store through the flow exec exits 2, never 0 or 1", 2),
		uwNoOut("an unreachable store through the flow exec emits no verdict line"),
		uwErr("an unreachable store through the flow exec names the failure on stderr", "cannot read findings for 'demo' from the store — cannot determine anything"))

	// Findings output that is not one array of objects is refused at exit 2,
	// through the injected read and the flow exec alike (decided 2026-09-27,
	// tasks.md task 4's Correction). None is a shape `flow record findings`
	// emits.
	for _, bad := range []struct{ label, json string }{
		{"empty output", ""}, {"an object", `{}`}, {"an object of findings", `{"a":{"status":"open"}}`},
		{"a null element", `[null]`}, {"two arrays", `[{"status":"open"}] []`},
		// KAN-886: a null status decodes into "" with no error, which
		// would be counted as one open finding instead of refused.
		{"a finding with a null status", `[{"ref":"F1","status":null}]`},
	} {
		each(demo(func(t *testing.T, f *uwFix) { f.findings = []byte(bad.json) }),
			uwRC("malformed findings refuse: "+bad.label+" exits 2", 2),
			uwNoOut("malformed findings refuse: "+bad.label+" emits no verdict line"),
			uwErr("malformed findings refuse: "+bad.label+" names jq failing", "jq failed — cannot determine anything"))
	}
	each(demo(func(t *testing.T, f *uwFix) {
		f.env.Findings = nil
		f.setFile(t, "findings.json", "")
	}),
		uwRC("malformed findings refuse: empty output through the flow exec exits 2", 2))

	// 5. Both signals at once are counted independently, on the one line.
	each(demo(func(t *testing.T, f *uwFix) { unticked(t, f); f.setFindings(t, "open") }),
		uwVerdict("both signals firing at once is OUTSTANDING", "OUTSTANDING:"),
		uwReason("the combined line names the plan signal", "unchecked plan item"),
		uwReason("the combined line also names the findings signal", "1 open finding(s)"))

	// 5b. A plan that exists but cannot be read is a refusal, never zero
	//     matches. Skipped as root, where chmod 000 does not block reads.
	if os.Geteuid() != 0 {
		each(demo(func(t *testing.T, f *uwFix) {
			p := f.wt + "/spectre/changes/demo/tasks.md"
			if err := os.Chmod(p, 0); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chmod(p, 0o644) })
		}),
			uwNonZero("an unreadable tasks.md exits non-zero"),
			uwNoOut("an unreadable tasks.md emits no verdict line"),
			uwErr("an unreadable tasks.md names the failure on stderr", "check-unfinished-work: cannot read"))
	}

	// 8e. The change name is PR-controlled; the allowlist is the same shapes
	//     TestCheckCleanupComplete/12 rejects.
	for _, bad := range []string{"../../../planted/clear", "demo*", "demo/../demo", ".hidden", "demo?x"} {
		named := func(t *testing.T) uwResult { f := newUW(t); return f.run(t, f.wt, bad) }
		each(named,
			uwRC("a change name outside the allowlist ("+bad+") -> exit 2", 2),
			uwNoOut("a change name outside the allowlist ("+bad+") emits no verdict line"),
			uwErr("a rejected change name ("+bad+") names the failure", "check-unfinished-work: change name"))
	}
	// 8e-i. The allowlist does not depend on the caller's locale.
	for _, loc := range []string{"C", "en_US.UTF-8"} {
		inLocale := func(t *testing.T) uwResult {
			f := newUW(t)
			get := f.env.Getenv
			f.env.Getenv = func(k string) string {
				if k == "LC_ALL" {
					return loc
				}
				return get(k)
			}
			return f.run(t, f.wt, "démo")
		}
		each(inLocale,
			uwRC("a non-ASCII change name is rejected under LC_ALL="+loc, 2),
			uwNoOut("a non-ASCII change name emits no verdict line under LC_ALL="+loc))
	}
	// 8f. Traversal, end to end: a relative name that would land exactly on
	//     a planted, fully-ticked plan in a sibling temp dir is refused by the
	//     allowlist before any path is built.
	traversal := func(t *testing.T) uwResult {
		parent := t.TempDir()
		f := newUWBare(t, parent+"/wt")
		f.plan(t, "- [x] 1.1 done\n")
		f.setFindings(t, "fixed")
		writeFile(t, parent+"/planted/spectre/changes/clear/tasks.md", "- [x] 1.1 done\n")
		return f.run(t, f.wt, "../../../planted/spectre/changes/clear")
	}
	each(traversal,
		uwRC("a traversal-shaped name is rejected by the allowlist before any path is built", 2),
		uwNoOut("a traversal-shaped name reads nothing outside the worktree"))

	// 9. A worktree that cannot be read is not a verdict.
	each(func(t *testing.T) uwResult { return newUW(t).run(t, "/nonexistent/worktree", "demo") },
		uwNonZero("an unreadable worktree exits non-zero"),
		uwNoOut("an unreadable worktree writes nothing to stdout"),
		uwErr("an unreadable worktree names the failure on stderr", "check-unfinished-work: "))
	// 10. A missing argument is programmer error.
	each(func(t *testing.T) uwResult { return newUW(t).run(t, "", "") },
		uwRC("missing arguments -> exit 2", 2),
		uwNoOut("missing arguments: emits no verdict line"))

	// 11. The CLI's stderr diagnostics never reach the JSON parser: the
	//     stub prints a FLOW_ADDR line on stderr and [] on stdout, and the
	//     guard execs it (Env.Findings nil).
	diag := demo(func(t *testing.T, f *uwFix) {
		f.env.Findings = nil
		f.setFindings(t)
		f.setFile(t, "findings.diag", "")
	})
	each(diag,
		uwRC("a diagnostic on stderr does not corrupt the findings JSON", 0),
		uwTrue("a diagnostic on stderr never reaches jq", func(r uwResult) bool {
			return !strings.Contains(r.out+r.err, "jq failed")
		}))

	// 12. Satellites follow the link to their canonical plan.
	satCanon := func(line string, extra func(t *testing.T, canon string)) func(t *testing.T) uwResult {
		return func(t *testing.T) uwResult {
			f := newUWSatellite(t, t.TempDir(), "peerx")
			canon := uwCanonTree(t, t.TempDir(), line)
			if extra != nil {
				extra(t, canon)
			}
			return f.run(t, f.wt, "sat-demo", canon)
		}
	}
	each(satCanon("[x] done", nil),
		uwVerdict("a satellite whose canonical plan (passed explicitly) is fully checked is CLEAR", "CLEAR:"),
		uwTrue("the verdict names the satellite's own worktree, not the canonical one", func(r uwResult) bool {
			return strings.HasPrefix(r.out, "CLEAR: "+r.wt)
		}))
	each(satCanon("[ ] not done", nil),
		uwVerdict("a satellite whose canonical plan has an unchecked item is OUTSTANDING", "OUTSTANDING:"),
		uwReason("the reason counts against the canonical plan", "1 unchecked plan item(s)"),
		uwTrue(`a resolved satellite plan is never reported as "no plan at"`, func(r uwResult) bool {
			return !strings.Contains(r.out, "no plan at")
		}))
	// 12c. Through spectre/peers, no canonical-worktree argument.
	each(func(t *testing.T) uwResult {
		parent := t.TempDir()
		f := newUWSatellite(t, parent+"/sat-tree", "peery")
		if err := os.Rename(uwCanonTree(t, parent, "[ ] not done"), parent+"/peer-tree"); err != nil {
			t.Fatal(err)
		}
		writeFile(t, f.wt+"/spectre/peers", "peery ../peer-tree\n")
		return f.run(t, f.wt, "sat-demo")
	},
		uwVerdict("a satellite resolving through peers reaches the peer's canonical plan", "OUTSTANDING:"),
		uwReason("the peer-resolved reason counts against the peer's plan", "1 unchecked plan item(s)"))
	// 12d. A supplied canonical worktree with no plan there refuses, and is
	//      never retried against peers even though they would resolve.
	each(func(t *testing.T) uwResult {
		parent := t.TempDir()
		f := newUWSatellite(t, parent+"/sat-tree", "peerw")
		if err := os.Rename(uwCanonTree(t, parent, "[x] done"), parent+"/peer-tree"); err != nil {
			t.Fatal(err)
		}
		writeFile(t, f.wt+"/spectre/peers", "peerw ../peer-tree\n")
		empty := t.TempDir()
		mkdir(t, empty+"/spectre/changes")
		return f.run(t, f.wt, "sat-demo", empty)
	},
		uwRC("a canonical worktree with no plan there exits 2, not 0", 2),
		uwNoOut("an unresolvable satellite (canonical arg) emits no verdict line"),
		uwErr("an unresolvable satellite (canonical arg) says it cannot determine anything", "cannot determine anything"),
		uwTrue("an unresolvable satellite (canonical arg) never carries OUTSTANDING or CLEAR", func(r uwResult) bool {
			return !strings.Contains(r.out+r.err, "OUTSTANDING") && !strings.Contains(r.out+r.err, "CLEAR")
		}))
	// 12e. No canonical argument, and the declared peer is absent from disk.
	each(func(t *testing.T) uwResult {
		f := newUWSatellite(t, t.TempDir()+"/sat-tree", "ghost")
		writeFile(t, f.wt+"/spectre/peers", "ghost ../not-checked-out\n")
		return f.run(t, f.wt, "sat-demo")
	},
		uwRC("a satellite whose peer is absent exits 2, not 0", 2),
		uwNoOut("a satellite whose peer is absent emits no verdict line"),
		uwErr("a satellite whose peer is absent says it cannot determine anything", "cannot determine anything"))
	// 12f. A link.md with no ## Part of is not a satellite: the ordinary
	//      missing-plan case.
	each(demo(func(t *testing.T, f *uwFix) {
		rmAll("spectre/changes/demo")(t, f)
		writeFile(t, f.wt+"/spectre/changes/demo/link.md", "## Parts\n\n`peerz:some-part`\n\n## Merge order\n\n1. `.`\n2. `peerz`\n")
	}),
		uwVerdict("a link.md with no ## Part of is not a satellite, so a missing plan is still OUTSTANDING", "OUTSTANDING:"),
		uwReason("a link.md with no ## Part of names the missing-plan signal, same as no link.md at all", "no plan at"))
	// 12g. A plain change ignores a supplied canonical-worktree argument.
	each(func(t *testing.T) uwResult { f := newUW(t); return f.run(t, f.wt, "demo", t.TempDir()) },
		uwVerdict("a plain change ignores a supplied canonical-worktree argument when its own plan exists", "CLEAR:"))
	// 12h. The fix-sub-change sweep follows the plan into the canonical tree.
	each(satCanon("[x] done", func(t *testing.T, canon string) {
		writeFile(t, canon+"/spectre/changes/canon-demo-fix-1/tasks.md", "- [ ] 1.1 the fix is not done\n")
	}),
		uwVerdict("a satellite whose canonical tree carries an unchecked canonical-fix sibling is OUTSTANDING, not CLEAR", "OUTSTANDING:"),
		uwReason("the unchecked canonical-fix sibling is counted, even though the canonical plan itself is fully checked", "1 unchecked plan item(s)"))

	// 13. KAN-368: an unticked step beneath a checked task counts nothing.
	each(demo(func(t *testing.T, f *uwFix) {
		f.plan(t, "- [x] 1. done\n  - [ ] **Step 1: still shows unticked, and must not count**\n")
	}),
		uwVerdict("an unticked step beneath a checked task does not gate the guard", "CLEAR:"))

	// 20. The recorded -verdict argument equals the printed line, on both
	//     CLEAR and OUTSTANDING.
	recorded := func(r uwResult) bool { return strings.Contains(r.args("verdict"), "-verdict "+r.out) }
	each(func(t *testing.T) uwResult {
		clear, outstanding := demo(nil)(t), demo(unticked)(t)
		if !recorded(clear) || !recorded(outstanding) {
			t.Fatalf("recorded -verdict argument does not equal the printed line (CLEAR ok=%v, OUTSTANDING ok=%v)", recorded(clear), recorded(outstanding))
		}
		return clear
	},
		uwCheck{"case 20: the recorded -verdict argument equals the printed verdict line on both CLEAR and OUTSTANDING", func(*testing.T, uwResult) {}})

	// 21-25. Prior false positives are advisory; both store calls are too.
	const priorTwo = `[
  {"id":2,"guard":"check-unfinished-work","worktree":"/wt","verdict":"OUTSTANDING: /wt - x","recordedAt":"2026-08-30T12:00:00Z","falsePositive":true,"falsePositiveReason":"verified structural: plan lives in backend repo","flaggedAt":"2026-08-30T12:00:00Z","change":"kan-393"},
  {"id":1,"guard":"check-unfinished-work","worktree":"/wt","verdict":"OUTSTANDING: /wt - y","recordedAt":"2026-08-01T09:00:00Z","falsePositive":true,"falsePositiveReason":"older reason","flaggedAt":"2026-08-01T09:00:00Z","change":"kan-260"}
]
`
	noAdvisory := func(verdict string) func(r uwResult) bool {
		return func(r uwResult) bool {
			return r.rc == 0 && strings.HasPrefix(r.out, verdict) && !strings.Contains(r.err, "prior false positives")
		}
	}
	each(demo(func(t *testing.T, f *uwFix) { unticked(t, f); f.setFile(t, "verdicts.json", priorTwo) }),
		uwTrue("case 21: OUTSTANDING with two prior false positives prints the advisory line naming the count and newest reason", func(r uwResult) bool {
			return strings.HasPrefix(r.out, "OUTSTANDING:") &&
				strings.Contains(r.err, "prior false positives for this guard on this project: 2 — last: verified structural: plan lives in backend repo (kan-393, 2026-08-30)")
		}))
	each(demo(unticked),
		uwTrue("case 22: OUTSTANDING with no prior false positives prints no advisory line", noAdvisory("OUTSTANDING:")))
	each(demo(func(t *testing.T, f *uwFix) {
		unticked(t, f)
		f.setFile(t, "verdicts.json", `[{"flaggedAt":[1,2,3,4,5,6,7,8,9,10,11],"change":null,"falsePositiveReason":null}]`+"\n")
	}),
		uwTrue("an array flaggedAt is sliced and rendered as compact JSON, as jq did", func(r uwResult) bool {
			return strings.Contains(r.err, "— last: null (null, [1,2,3,4,5,6,7,8,9,10])")
		}))
	each(demo(func(t *testing.T, f *uwFix) { unticked(t, f); f.setFile(t, "verdicts.fail", "") }),
		uwTrue("case 23: verdicts unreachable prints no advisory line and leaves the verdict and exit 0 unchanged", noAdvisory("OUTSTANDING:")))
	each(demo(func(t *testing.T, f *uwFix) { f.setFile(t, "verdict.fail", "") }),
		uwVerdict("case 24: verdict write unreachable leaves the verdict line and exit 0 unchanged", "CLEAR:"))
	each(demo(func(t *testing.T, f *uwFix) {
		f.setFile(t, "verdicts.json", `[{"id":1,"guard":"check-unfinished-work","worktree":"/wt","verdict":"OUTSTANDING: /wt - x","recordedAt":"2026-08-01T09:00:00Z","falsePositive":true,"falsePositiveReason":"should never be read","flaggedAt":"2026-08-01T09:00:00Z","change":"kan-260"}]`+"\n")
	}),
		uwTrue("case 25: CLEAR never calls verdicts, so no advisory line ever prints", noAdvisory("CLEAR:")))

	// 26. KAN-260: the store calls anchor at the plan's project. These cases
	//     exec the stub (Env.Findings nil) so the -C it received is logged.
	// 26: absent-dir cross-repo, plan resolved from the canonical argument.
	absentDir := func(t *testing.T) uwResult {
		f := newUWBare(t, t.TempDir())
		f.env.Findings = nil
		canon := t.TempDir()
		writeFile(t, canon+"/spectre/changes/demo/tasks.md", "- [x] 1. done\n")
		r := f.run(t, f.wt, "demo", canon)
		r.canon = canon
		return r
	}
	each(absentDir,
		uwVerdict("case 26: an absent-dir cross-repo plan resolves and is CLEAR", "CLEAR:"),
		uwTrue("case 26: the findings query anchors at the canonical plan dir", func(r uwResult) bool {
			return r.anchored("findings", r.canon+"/spectre/changes/demo")
		}),
		uwTrue("case 26: the verdict write anchors at the canonical plan dir", func(r uwResult) bool {
			return r.anchored("verdict", r.canon+"/spectre/changes/demo")
		}),
		uwTrue("case 26: the verdict still names the judged worktree", func(r uwResult) bool {
			return strings.Contains(r.args("verdict"), "-worktree "+r.wt)
		}))
	// 26a: a worktree whose plan lives in another repo leaves the change's
	//      dispatch rows to the plan's worktree, so a multi-repo change
	//      lists each open row once.
	each(func(t *testing.T) uwResult {
		f := newUWBare(t, t.TempDir())
		f.env.Findings = nil
		f.setFile(t, "dispatches.json", `[{"key":"task-14-reviewer"}]`+"\n")
		canon := t.TempDir()
		writeFile(t, canon+"/spectre/changes/demo/tasks.md", "- [x] 1. done\n")
		return f.run(t, f.wt, "demo", canon)
	},
		uwVerdict("case 26a: a non-plan worktree does not repeat the change's open dispatch rows", "CLEAR:"))
	// 26b/26c: a local plan, and the no-plan fall-through, anchor at the
	// judged worktree.
	local := func(setup func(t *testing.T, f *uwFix)) func(t *testing.T) uwResult {
		return demo(func(t *testing.T, f *uwFix) {
			f.env.Findings = nil
			f.setFindings(t)
			if setup != nil {
				setup(t, f)
			}
		})
	}
	atWorktree := func(verb string) func(r uwResult) bool {
		return func(r uwResult) bool { return r.anchored(verb, r.wt) }
	}
	each(local(nil),
		uwVerdict("case 26b: a local plan is CLEAR", "CLEAR:"),
		uwTrue("case 26b: the findings query anchors at the judged worktree", atWorktree("findings")),
		uwTrue("case 26b: the verdict write anchors at the judged worktree", atWorktree("verdict")))
	each(local(func(t *testing.T, f *uwFix) {
		if err := os.Remove(f.wt + "/spectre/changes/demo/tasks.md"); err != nil {
			t.Fatal(err)
		}
	}),
		uwVerdict("case 26c: a missing plan is OUTSTANDING", "OUTSTANDING:"),
		uwTrue("case 26c: the fall-through findings query anchors at the judged worktree", atWorktree("findings")),
		uwTrue("case 26c: the fall-through verdict write anchors at the judged worktree", atWorktree("verdict")))

	// 27. KAN-267: with no canonical argument, the state record's worktrees
	//     map resolves the plan.
	storeResolved := func(t *testing.T) uwResult {
		f := newUWBare(t, t.TempDir())
		f.env.Findings = nil
		canon := t.TempDir()
		writeFile(t, canon+"/spectre/changes/demo/tasks.md", "- [x] 1. done\n")
		f.setFile(t, "find.json", `{"source":"store","complete":true,"records":[{"projectKey":"proj-a","name":"demo","state":"IN_PROGRESS","worktrees":{"`+
			f.wt+`":null,"`+canon+`":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"},"updatedAt":"2026-09-10T10:00:00Z","updatedBy":"/flow-fast"}]}`+"\n")
		r := f.run(t, f.wt, "demo")
		r.canon = canon
		return r
	}
	each(storeResolved,
		uwVerdict("case 27: the record's worktrees map resolves the plan and the change is CLEAR", "CLEAR:"),
		uwTrue("case 27: the store was asked for the change's own name", func(r uwResult) bool {
			return strings.Contains(r.args("find"), "state find demo")
		}),
		uwTrue("case 27: the findings query anchors at the resolved plan dir", func(r uwResult) bool {
			return r.anchored("findings", r.canon+"/spectre/changes/demo")
		}),
		uwTrue("case 27: the verdict still names the judged worktree", func(r uwResult) bool {
			return strings.Contains(r.args("verdict"), "-worktree "+r.wt)
		}))

	// 28. KAN-267: two projects' records each resolve a plan — refuse.
	each(func(t *testing.T) uwResult {
		f := newUWBare(t, t.TempDir())
		a, b := t.TempDir(), t.TempDir()
		writeFile(t, a+"/spectre/changes/demo/tasks.md", "- [x] 1. done\n")
		writeFile(t, b+"/spectre/changes/demo/tasks.md", "- [ ] 1. not done\n")
		f.setFile(t, "find.json", `{"source":"store","complete":true,"records":[`+"\n"+
			`{"projectKey":"proj-a","name":"demo","worktrees":{"`+a+`":"cccccccccccccccccccccccccccccccccccccccc"}},`+"\n"+
			`{"projectKey":"proj-b","name":"demo","worktrees":{"`+b+`":"dddddddddddddddddddddddddddddddddddddddd"}}]}`+"\n")
		return f.run(t, f.wt, "demo")
	},
		uwRC("case 28: an ambiguous record answer refuses outright", 2),
		uwNoOut("case 28: no verdict line on stdout"),
		uwTrue("case 28: the refusal names every project", func(r uwResult) bool {
			i := strings.Index(r.err, "ambiguous state-record resolution")
			return i >= 0 && strings.Contains(r.err[i:], "proj-a") &&
				strings.Index(r.err[i:], "proj-a") < strings.LastIndex(r.err[i:], "proj-b")
		}))
}

func TestCheckUnfinishedWorkCloseCommandCarriesEndedAt(t *testing.T) {
	f := newUW(t)
	f.setFile(t, "dispatches.json", `[{"key":"task-2-reviewer","role":"reviewer","sessionToken":"mf-r"},`+
		`{"key":"task-3-implementer","role":"implementer","sessionToken":"mf-i"}]`+"\n")
	r := f.run(t, f.wt, "demo")
	for _, want := range []string{
		"-key task-2-reviewer -session-token mf-r -outcome <outcome> -C " + f.wt + " -ended-at <when its work ended>`",
		"-key task-3-implementer -session-token mf-i -outcome <outcome> -C " + f.wt + " -ended-at <when its work ended> -commit <its task commit>`",
	} {
		if !strings.Contains(r.out, want) {
			t.Errorf("close command missing %q:\n%s", want, r.out)
		}
	}
}

// TestCheckUnfinishedWorkNoPanelPassRecorded pins signal eight (kan-970
// self-review): a change whose range carries a code commit and whose store
// holds no pass row at all is OUTSTANDING; a recorded pass, or only
// planning commits, clears it.
func TestCheckUnfinishedWorkNoPanelPassRecorded(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name, path, passes, want string
	}{
		{"a code commit with no pass row is OUTSTANDING", "src/a.go", "[]", "no review-panel pass recorded for code commit "},
		{"a code commit with a pass row is CLEAR", "src/a.go", `[{"id":1,"round":0,"note":"roster: full"}]`, "CLEAR:"},
		{"a planning-only commit with no pass row is CLEAR", "spectre/changes/demo/notes.md", "[]", "CLEAR:"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			f := newUW(t)
			gitRun(t, "", "init", "-q", f.wt)
			gitRun(t, f.wt, "commit", "-q", "--allow-empty", "-m", "base")
			base := tcfGit(t, f.wt, "rev-parse", "HEAD")
			writeFile(t, f.wt+"/"+c.path, "x\n")
			gitRun(t, f.wt, "add", c.path)
			gitRun(t, f.wt, "commit", "-q", "--date=2026-01-01T11:00:00Z", "-m", "fix: code")
			f.setFile(t, "dispatches.json", `[{"seq":9,"key":"verify","role":"verifier","outcome":"completed","startedAt":"2026-01-01T12:00:00Z"}]`+"\n")
			f.setFile(t, "passes.json", c.passes+"\n")
			r := f.run(t, f.wt, "demo", "", f.wt+"="+base)
			if r.rc != 0 || !strings.Contains(r.out, c.want) {
				t.Fatalf("exit %d stdout %q stderr %q, want exit 0 and %q", r.rc, r.out, r.err, c.want)
			}
		})
	}
}

// TestCheckUnfinishedWorkRefusesAnUnsplitRangeList pins the kan-970
// self-review fix: two ranges passed as one zsh-unsplit argument are refused
// at the argument, naming the cause, not later as an unreadable git range.
func TestCheckUnfinishedWorkRefusesAnUnsplitRangeList(t *testing.T) {
	t.Parallel()
	f := newUW(t)
	a := uwCommits(t, t.TempDir()+"/a", "base", "fix: a")
	b := uwCommits(t, t.TempDir()+"/b", "base", "fix: b")
	r := f.run(t, f.wt, "demo", "", a+" "+b)
	if r.rc != 2 || !strings.Contains(r.err, "zsh does not word-split an unquoted $VAR") {
		t.Fatalf("exit %d stderr %q, want exit 2 naming the unsplit list", r.rc, r.err)
	}
}

// TestCheckUnfinishedWorkVerifyFailureNotAddressed pins signal nine
// (decision verify-failure-signal, kan-994 self-review): a verify report's
// failed capture that no commit after the report touched, and no narrative
// line cites a range commit for, is OUTSTANDING.
func TestCheckUnfinishedWorkVerifyFailureNotAddressed(t *testing.T) {
	t.Parallel()
	const capture = "  1 failed — fidelity/full-app-suite.spec.ts:110 register-from-invite.png, 24px differ\n"
	const reason = "verify failure not addressed: fidelity/full-app-suite.spec.ts:110 register-from-invite.png (verify-report-visual-verify.md)"
	before, after := "2026-01-01T11:00:00Z", "2026-01-01T13:00:00Z"
	reportAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	commitAt := func(t *testing.T, dir, date, path string) string {
		t.Helper()
		if path != "" {
			writeFile(t, dir+"/"+path, date+"\n")
			gitRun(t, dir, "add", path)
		}
		cmd := exec.Command(fixtureGit, "-C", dir, "commit", "-q", "--allow-empty", "-m", "fix: "+path)
		cmd.Env = append(append(os.Environ(), fixtureGitEnv...),
			"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
			"GIT_AUTHOR_DATE="+date, "GIT_COMMITTER_DATE="+date)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("commit: %v\n%s", err, out)
		}
		return tcfGit(t, dir, "rev-parse", "HEAD")
	}
	for _, c := range []struct {
		name   string
		report string
		// setup makes the range's commits and returns the narrative body.
		setup  func(t *testing.T, wt, base string) string
		want   string
		wantIn bool
	}{
		{"no commit touches the capture", "- verify: exit 1\n" + capture,
			func(t *testing.T, wt, base string) string { return "" }, reason, true},
		{"a later commit touching the capture addresses it", "- verify: exit 1\n" + capture,
			func(t *testing.T, wt, base string) string {
				commitAt(t, wt, after, "fidelity/full-app-suite.spec.ts-snapshots/register-from-invite-darwin.png")
				return ""
			}, "verify failure not addressed", false},
		{"a commit before the report does not", "- verify: exit 1\n" + capture,
			func(t *testing.T, wt, base string) string {
				commitAt(t, wt, before, "fidelity/full-app-suite.spec.ts-snapshots/register-from-invite-darwin.png")
				return ""
			}, reason, true},
		{"a narrative line citing a range commit addresses it", "- verify: exit 1\n" + capture,
			func(t *testing.T, wt, base string) string {
				sha := commitAt(t, wt, after, "src/other.go")
				return "Verify failure addressed: fidelity/full-app-suite.spec.ts:110 — " + sha + "\n"
			}, "verify failure not addressed", false},
		{"a narrative line citing a sha outside the range does not", "- verify: exit 1\n" + capture,
			func(t *testing.T, wt, base string) string {
				return "Verify failure addressed: fidelity/full-app-suite.spec.ts:110 — " + base + "\n"
			}, reason, true},
		{"a failure naming no capture is addressed by a later touch of its spec", "- verify: exit 1\n  1 failed — tests/auth.spec.ts:42 times out\n",
			func(t *testing.T, wt, base string) string {
				commitAt(t, wt, after, "tests/auth.spec.ts")
				return ""
			}, "verify failure not addressed", false},
		{"a passing report contributes nothing", "- verify: exit 0\n" + capture,
			func(t *testing.T, wt, base string) string { return "" }, "verify failure not addressed", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			f := newUW(t)
			gitRun(t, "", "init", "-q", f.wt)
			base := commitAt(t, f.wt, before, "")
			if body := c.setup(t, f.wt, base); body != "" {
				writeFile(t, f.wt+"/spectre/changes/demo/narrative.md", body)
			}
			rp := f.wt + "/.superpowers/sdd/verify-report-visual-verify.md"
			writeFile(t, rp, "# Verify\n\n- build: exit 0\n"+c.report+"- lint: exit 0\n")
			if err := os.Chtimes(rp, reportAt, reportAt); err != nil {
				t.Fatal(err)
			}
			r := f.run(t, f.wt, "demo", "", f.wt+"="+base)
			if r.rc != 0 || strings.Contains(r.out, c.want) != c.wantIn {
				t.Fatalf("exit %d stdout %q stderr %q; want exit 0 and %q present=%v", r.rc, r.out, r.err, c.want, c.wantIn)
			}
		})
	}
}

// TestCheckUnfinishedWorkVerifyFailureRerunPassed pins kan-1022's
// self-review: a failed report whose next report, by mtime, passed a rerun
// of the failed specs is addressed; a passing rerun before the failure, or a
// failing next report, addresses nothing.
func TestCheckUnfinishedWorkVerifyFailureRerunPassed(t *testing.T) {
	t.Parallel()
	const failed = "- verify: exit 1\n  1) fidelity/templates.spec.ts:80 opens the template — timeout\n"
	const reason = "verify failure not addressed: fidelity/templates.spec.ts:80"
	first, second := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC), time.Date(2026, 1, 1, 13, 0, 0, 0, time.UTC)
	for _, c := range []struct {
		name           string
		failAt, nextAt time.Time
		next           string
		wantIn         bool
	}{
		{"a later passing rerun addresses it", first, second, "- verify: exit 0 — rerun of 19 failed specs at unchanged HEAD\n", false},
		{"an earlier passing rerun does not", second, first, "- verify: exit 0 — rerun of 19 failed specs at unchanged HEAD\n", true},
		{"a later failing rerun does not", first, second, "- verify: exit 1 — rerun of 19 failed specs at unchanged HEAD\n", true},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			f := newUW(t)
			gitRun(t, "", "init", "-q", f.wt)
			gitRun(t, f.wt, "commit", "-q", "--allow-empty", "-m", "base")
			base := tcfGit(t, f.wt, "rev-parse", "HEAD")
			for _, r := range []struct {
				name, body string
				at         time.Time
			}{{"visual-verify", failed, c.failAt}, {"visual-verify-2", c.next, c.nextAt}} {
				rp := f.wt + "/.superpowers/sdd/verify-report-" + r.name + ".md"
				writeFile(t, rp, "## Report\n"+r.body)
				if err := os.Chtimes(rp, r.at, r.at); err != nil {
					t.Fatal(err)
				}
			}
			r := f.run(t, f.wt, "demo", "", f.wt+"="+base)
			if r.rc != 0 || strings.Contains(r.out, reason) != c.wantIn {
				t.Fatalf("exit %d stdout %q stderr %q; want exit 0 and %q present=%v", r.rc, r.out, r.err, reason, c.wantIn)
			}
		})
	}
}

// TestCheckUnfinishedWorkVerifyFailureFlaggedEarlier pins kan-1010's
// self-review: screenshots.spec.ts:52 was flagged a false positive three
// times, each flag landing on that run's newest row only. An earlier
// false-positive verdict of this change naming the failure's `<spec>:<line>`
// clears it; another change's verdict, another line's, a longer line
// number's and a store that cannot answer do not.
func TestCheckUnfinishedWorkVerifyFailureFlaggedEarlier(t *testing.T) {
	t.Parallel()
	const reason = "verify failure not addressed: screenshots.spec.ts:52"
	row := func(change, why string) string {
		return `{"id":1,"change":"` + change + `","guard":"check-unfinished-work","falsePositive":true,"falsePositiveReason":"` + why + `","flaggedAt":"2026-10-22T02:09:55Z"}`
	}
	for _, c := range []struct {
		name     string
		verdicts string
		fail     bool
		wantIn   bool
	}{
		{"no verdict leaves it outstanding", "[]", false, true},
		{"this change's verdict naming it clears it", "[" + row("demo", "screenshots.spec.ts:52 does not recur: 5/5 passed") + "]", false, false},
		{"another change's verdict does not", "[" + row("other", "screenshots.spec.ts:52 does not recur") + "]", false, true},
		{"a verdict naming another line does not", "[" + row("demo", "screenshots.spec.ts:520 does not recur") + "]", false, true},
		{"a verdict naming another spec does not", "[" + row("demo", "fidelity/screenshots.spec.ts:52 does not recur") + "]", false, true},
		{"a store that cannot answer leaves it outstanding", "", true, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			f := newUW(t)
			gitRun(t, "", "init", "-q", f.wt)
			gitRun(t, f.wt, "commit", "-q", "--allow-empty", "-m", "base")
			base := tcfGit(t, f.wt, "rev-parse", "HEAD")
			writeFile(t, f.wt+"/.superpowers/sdd/verify-report-visual-verify.md",
				"# Verify\n\n- verify: exit 1\n  1) [functional] screenshots.spec.ts:52 sweep every reachable screen — test timeout\n")
			f.setFile(t, "verdicts.json", c.verdicts+"\n")
			if c.fail {
				f.setFile(t, "verdicts.fail", "")
			}
			r := f.run(t, f.wt, "demo", "", f.wt+"="+base)
			if r.rc != 0 || strings.Contains(r.out, reason) != c.wantIn {
				t.Fatalf("exit %d stdout %q stderr %q; want exit 0 and %q present=%v", r.rc, r.out, r.err, reason, c.wantIn)
			}
		})
	}
}

// TestCheckUnfinishedWorkMergeOrderRecord pins signal ten (decision
// merge-order-signal, kan-995 self-review): a change whose ranges span a
// second repository is OUTSTANDING until link.md's `## Merge order` names
// that repository by a peer the spectre/peers file resolves to it.
func TestCheckUnfinishedWorkMergeOrderRecord(t *testing.T) {
	t.Parallel()
	const reason = "merge order does not name "
	for _, c := range []struct {
		name   string
		link   string // "" writes no link.md
		peers  string
		ranges string // "both", "plan" or "same-repo"
		wantIn bool
	}{
		{"no link.md", "", "peer ../peer\n", "both", true},
		{"merge order names only the plan repository", "## Merge order\n\n1. `.`\n", "peer ../peer\n", "both", true},
		{"merge order names the peer", "## Merge order\n\n1. `.`\n2. `peer`\n", "peer ../peer\n", "both", false},
		{"a Part of section is not the merge order", "## Part of\n\n1. `.`\n2. `peer`\n", "peer ../peer\n", "both", true},
		{"one range reads nothing", "", "peer ../peer\n", "plan", false},
		{"two worktrees of one repository read nothing", "", "peer ../peer\n", "same-repo", false},
		{"a peer the peers file does not declare", "## Merge order\n\n1. `.`\n2. `peer`\n", "other ../other\n", "both", true},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			f := newUW(t)
			planRange := uwCommits(t, f.wt, "base", "fix: plan")
			writeFile(t, f.wt+"/spectre/peers", c.peers)
			if c.link != "" {
				writeFile(t, f.wt+"/spectre/changes/demo/link.md", c.link)
			}
			args := []string{f.wt, "demo", "", planRange}
			switch c.ranges {
			case "both":
				args = append(args, uwCommits(t, filepath.Dir(f.wt)+"/peer", "base", "fix: peer"))
			case "same-repo":
				linked := filepath.Dir(f.wt) + "/linked"
				gitRun(t, f.wt, "worktree", "add", "-q", linked)
				args = append(args, linked+"="+tcfGit(t, linked, "rev-parse", "HEAD~1"))
			}
			r := f.run(t, args...)
			if r.rc != 0 || strings.Contains(r.out, reason) != c.wantIn {
				t.Fatalf("exit %d stdout %q stderr %q; want exit 0 and %q present=%v", r.rc, r.out, r.err, reason, c.wantIn)
			}
			peer, err := filepath.EvalSymlinks(filepath.Dir(f.wt) + "/peer")
			if err != nil {
				peer = filepath.Dir(f.wt) + "/peer"
			}
			if c.wantIn && !strings.Contains(r.out, reason+peer+":") {
				t.Fatalf("the reason does not name the peer's main checkout: %s", r.out)
			}
		})
	}
}

// TestCheckUnfinishedWorkReportCount pins kan-1015's self-review: five panel
// reports counted 13 Critical and Important findings the store never held,
// and integrate read CLEAR. The gate now counts each report against the
// store as check-panel-findings-closed does.
func TestCheckUnfinishedWorkReportCount(t *testing.T) {
	t.Parallel()
	const reason = "panel-report-0-primary.md (report 2, recorded 0)"
	for _, c := range []struct {
		name, findings string
		want           bool
	}{
		{"a report the store under-counts is OUTSTANDING", "[]", true},
		{"a report the store holds is not refused", `[{"ref":"F1","status":"fixed","severity":"important","round":0,"slot":"primary"},{"ref":"F2","status":"fixed","severity":"important","round":0,"slot":"primary"}]`, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			f := newUW(t)
			f.findings = []byte(c.findings)
			writeFile(t, f.wt+"/.superpowers/sdd/panel-report-0-primary.md", "# primary\n\nCritical: 0\nImportant: 2\n")
			r := f.run(t, f.wt, "demo")
			if got := strings.Contains(r.out, reason); got != c.want || (c.want && !strings.HasPrefix(r.out, "OUTSTANDING:")) {
				t.Fatalf("exit %d stdout %q stderr %q; want reason %q present=%v", r.rc, r.out, r.err, reason, c.want)
			}
		})
	}
}

// TestUWVerifyFailuresSpecExcludesOpeningPunctuation pins kan-1023's
// self-review: a Playwright location wrapped in `(`, `[`, a backtick or a
// quote names the spec alone, so the citation the guard demands carries no
// stray character and a clean citation clears it.
func TestUWVerifyFailuresSpecExcludesOpeningPunctuation(t *testing.T) {
	t.Parallel()
	report := "- verify: exit 1\n" +
		"  1 failed (kan-341-route-b.spec.ts:224:5) and `fidelity/a.spec.ts:3` a.png\n" +
		"  [x/b.spec.ts:9] 'c.spec.ts:1' \"d.spec.ts:2\"\n"
	want := []uwVerifyFailure{
		{spec: "kan-341-route-b.spec.ts", line: "224"},
		{spec: "fidelity/a.spec.ts", line: "3", png: "a"},
		{spec: "x/b.spec.ts", line: "9"},
		{spec: "c.spec.ts", line: "1"},
		{spec: "d.spec.ts", line: "2"},
	}
	if got := uwVerifyFailures(report); !slices.Equal(got, want) {
		t.Fatalf("uwVerifyFailures = %+v, want %+v", got, want)
	}
}

// TestUnfinishedWorkRefusesAFixedRoundWithNoMutationRow pins kan-1018's
// self-review: a round holding a fixed Critical or Important finding needs a
// fix-mutation row at that round or the next, a `none — <reason>` row
// included; Minor-only rounds and withdrawn findings demand none.
func TestUnfinishedWorkRefusesAFixedRoundWithNoMutationRow(t *testing.T) {
	t.Parallel()
	const want = "fixed Critical/Important finding(s) in round(s) 0 and no fix-mutation row for that fix round"
	findings := `[{"ref":"F1","round":0,"status":"fixed","severity":"Important"},{"ref":"F2","round":0,"status":"fixed","severity":"Minor"},{"ref":"F3","round":1,"status":"withdrawn not real","severity":"Critical"}]`
	for _, c := range []struct {
		name, findings, mutations, want string
	}{
		{"fixed findings with no mutation row are OUTSTANDING", findings, "[]", want},
		{"a none row for the fix round is CLEAR", findings, `[{"id":1,"round":1,"path":"a","mutated":"none","test":"doc-only"}]`, "CLEAR:"},
		{"a row for the findings' own round is CLEAR", findings, `[{"id":1,"round":0,"path":"a","mutated":"none","test":"doc-only"}]`, "CLEAR:"},
		{"a row two rounds later is OUTSTANDING", findings, `[{"id":1,"round":2,"path":"a","mutated":"none","test":"doc-only"}]`, want},
		{"a round of fixed Minors is CLEAR", `[{"ref":"F1","round":0,"status":"fixed","severity":"Minor"}]`, "[]", "CLEAR:"},
		{"a lowercase important is OUTSTANDING", `[{"ref":"F1","round":0,"status":"fixed","severity":"important"},{"ref":"F2","round":0,"status":"fixed","severity":"minor"}]`, "[]", want},
		{"only withdrawn findings are CLEAR", `[{"ref":"F1","round":0,"status":"withdrawn not real"}]`, "[]", "CLEAR:"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			f := newUW(t)
			f.findings = []byte(c.findings)
			f.setFile(t, "findings.json", c.findings+"\n")
			f.setFile(t, "mutations.json", c.mutations+"\n")
			r := f.run(t, f.wt, "demo")
			if r.rc != 0 || !strings.Contains(r.out, c.want) {
				t.Fatalf("exit %d stdout %q stderr %q, want exit 0 and %q", r.rc, r.out, r.err, c.want)
			}
		})
	}
	t.Run("an unreadable mutation store exits 2", func(t *testing.T) {
		t.Parallel()
		f := newUW(t)
		f.findings = []byte(findings)
		f.setFile(t, "findings.json", findings+"\n")
		f.setFile(t, "mutations.fail", "")
		if r := f.run(t, f.wt, "demo"); r.rc != 2 {
			t.Fatalf("exit %d stdout %q, want exit 2", r.rc, r.out)
		}
	})
}

// uwVerifyCase runs signal seven on one code commit at 11:00 and one verify
// row started at 12:00 with outcome, a recorded panel pass beside it.
func uwVerifyCase(t *testing.T, outcome, startedAt string) uwResult {
	f := newUW(t)
	gitRun(t, "", "init", "-q", f.wt)
	gitRun(t, f.wt, "commit", "-q", "--allow-empty", "-m", "base")
	base := tcfGit(t, f.wt, "rev-parse", "HEAD")
	writeFile(t, f.wt+"/src/a.go", "x\n")
	gitRun(t, f.wt, "add", "src/a.go")
	gitRun(t, f.wt, "commit", "-q", "--date=2026-01-01T11:00:00Z", "-m", "fix: code")
	f.setFile(t, "dispatches.json", `[{"seq":9,"key":"verify","role":"verifier","outcome":"`+outcome+`","startedAt":"`+startedAt+`"}]`+"\n")
	f.setFile(t, "passes.json", `[{"id":1,"round":0,"note":"roster: full"}]`+"\n")
	return f.run(t, f.wt, "demo", "", f.wt+"="+base)
}

// TestCheckUnfinishedWorkSkippedVerifyRowCovers pins decision
// f1-skipped-is-verified (kan-1037): a cache-skipped verify row covers the
// code commits before it; a blocked one does not.
func TestCheckUnfinishedWorkSkippedVerifyRowCovers(t *testing.T) {
	t.Parallel()
	for outcome, want := range map[string]string{"skipped": "CLEAR:", "blocked": "OUTSTANDING:"} {
		t.Run(outcome, func(t *testing.T) {
			t.Parallel()
			if r := uwVerifyCase(t, outcome, "2026-01-01T12:00:00Z"); !strings.Contains(r.out, want) {
				t.Fatalf("exit %d stdout %q stderr %q, want %q", r.rc, r.out, r.err, want)
			}
		})
	}
}

// TestCheckUnfinishedWorkReasonNamesFreshVerifyKey pins decision
// f2-fresh-verify-key (kan-1037): the reason tells a re-run to take a fresh key.
func TestCheckUnfinishedWorkReasonNamesFreshVerifyKey(t *testing.T) {
	t.Parallel()
	r := uwVerifyCase(t, "completed", "2026-01-01T10:00:00Z")
	if want := "a re-run after a new commit takes a fresh `verify-<k>` key"; !strings.Contains(r.out+r.err, want) {
		t.Fatalf("exit %d stdout %q stderr %q, want %q", r.rc, r.out, r.err, want)
	}
}

// TestUnfinishedWorkFlagsAPanelFixCommitWithNoRow pins decision
// gate-panel-fix-trailer-row (kan-1044 self-review, F1-c): a range commit
// whose Flow-Key trailer names a panel-fix key no dispatch row carries is
// OUTSTANDING, whether HEAD or only the pre-reshape pin reaches it.
func TestUnfinishedWorkFlagsAPanelFixCommitWithNoRow(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name, rows string
		pinOnly    bool
		flagged    bool
	}{
		{"a trailer with no row on HEAD", `{"key":"panel-fix-1","role":"panel-fix","outcome":"completed"}`, false, true},
		{"a trailer with no row behind the pin", `{"key":"panel-fix-1","role":"panel-fix","outcome":"completed"}`, true, true},
		{"a trailer with its row", `{"key":"panel-fix-3","role":"panel-fix","outcome":"completed"}`, false, false},
		{"a round-low trailer with a row on its commit", `{"key":"panel-fix-4","role":"panel-fix","outcome":"completed","commitSha":"FULL"}`, false, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			f := newUW(t)
			gitRun(t, "", "init", "-q", f.wt)
			gitRun(t, f.wt, "commit", "-q", "--allow-empty", "-m", "base")
			base := tcfGit(t, f.wt, "rev-parse", "HEAD")
			writeFile(t, f.wt+"/src/a.go", "x\n")
			gitRun(t, f.wt, "add", "src/a.go")
			gitRun(t, f.wt, "commit", "-q", "-m", "fix: round 3\n\nFlow-Key: panel-fix-3")
			sha := tcfGit(t, f.wt, "rev-parse", "--short", "HEAD")
			full := tcfGit(t, f.wt, "rev-parse", "HEAD")
			if c.pinOnly {
				gitRun(t, f.wt, "update-ref", "refs/flow/pre-reshape/demo", "HEAD")
				gitRun(t, f.wt, "reset", "-q", "--hard", base)
			}
			f.setFile(t, "dispatches.json", `[{"seq":9,"key":"verify","role":"verifier","outcome":"completed","startedAt":"2099-01-01T12:00:00Z"},`+strings.ReplaceAll(c.rows, "FULL", full)+`]`+"\n")
			r := f.run(t, f.wt, "demo", "", f.wt+"="+base)
			got := strings.Contains(r.out, sha) && strings.Contains(r.out, "panel-fix-3")
			if got != c.flagged {
				t.Fatalf("flagged = %v, want %v; exit %d stdout %q stderr %q", got, c.flagged, r.rc, r.out, r.err)
			}
		})
	}
}

// TestUnfinishedWorkReadsHeadBesideThePin pins signal eleven reading HEAD and
// the pre-reshape pin together: a trailer committed on HEAD after the pin is
// still flagged.
func TestUnfinishedWorkReadsHeadBesideThePin(t *testing.T) {
	t.Parallel()
	f := newUW(t)
	gitRun(t, "", "init", "-q", f.wt)
	gitRun(t, f.wt, "commit", "-q", "--allow-empty", "-m", "base")
	base := tcfGit(t, f.wt, "rev-parse", "HEAD")
	gitRun(t, f.wt, "update-ref", "refs/flow/pre-reshape/demo", "HEAD")
	writeFile(t, f.wt+"/src/a.go", "x\n")
	gitRun(t, f.wt, "add", "src/a.go")
	gitRun(t, f.wt, "commit", "-q", "-m", "fix: round 5\n\nFlow-Key: panel-fix-5")
	sha := tcfGit(t, f.wt, "rev-parse", "--short", "HEAD")
	f.setFile(t, "dispatches.json", `[{"seq":9,"key":"verify","role":"verifier","outcome":"completed","startedAt":"2099-01-01T12:00:00Z"}]`+"\n")
	r := f.run(t, f.wt, "demo", "", f.wt+"="+base)
	if !strings.Contains(r.out, sha) || !strings.Contains(r.out, "panel-fix-5") {
		t.Fatalf("HEAD commit not flagged; exit %d stdout %q stderr %q", r.rc, r.out, r.err)
	}
}

// A gated reviewer's fixed Important needs no fix-mutation row, yet an open
// gated row still stops the gate.
func TestUnfinishedWorkIgnoresGatedRowsForFixRoundProof(t *testing.T) {
	t.Parallel()
	for _, c := range []struct{ name, findings, want string }{
		{"a fixed gated Important is CLEAR", `[{"ref":"F1","round":0,"slot":"gated-reviewer","status":"fixed","severity":"Important"}]`, "CLEAR:"},
		{"an open gated row is OUTSTANDING", `[{"ref":"F1","round":0,"slot":"gated-reviewer","status":"open","severity":"Important"}]`, "1 open finding(s)"},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			f := newUW(t)
			f.findings = []byte(c.findings)
			f.setFile(t, "findings.json", c.findings+"\n")
			f.setFile(t, "mutations.json", "[]\n")
			r := f.run(t, f.wt, "demo")
			if r.rc != 0 || !strings.Contains(r.out, c.want) {
				t.Fatalf("exit %d stdout %q stderr %q, want exit 0 and %q", r.rc, r.out, r.err, c.want)
			}
		})
	}
}
