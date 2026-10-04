package guard

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
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

	// 3. Signal one — the plan.
	each(demo(unticked),
		uwVerdict("an unchecked plan item is OUTSTANDING", "OUTSTANDING:"),
		uwReason("an unchecked plan item names its signal", "unchecked plan item"))
	// 3a. A change run 1 archived on its branch, with no live directory left,
	//     is judged by its archived plan — the one a later fix run appends to.
	archived := func(body string) func(t *testing.T) uwResult {
		return func(t *testing.T) uwResult {
			f := newUWBare(t, t.TempDir())
			f.setFindings(t, "fixed")
			writeFile(t, f.wt+"/spectre/changes/archive/demo/tasks.md", body)
			return f.run(t, f.wt, "demo")
		}
	}
	each(archived("- [x] 1.1 done\n"),
		uwVerdict("an archived change whose plan is all checked is CLEAR", "CLEAR:"))
	each(archived("- [x] 1.1 done\n- [ ] 2. appended by a fix run\n"),
		uwVerdict("an unchecked item in an archived change's plan is OUTSTANDING", "OUTSTANDING:"),
		uwReason("an archived change's unchecked item names its signal", "1 unchecked plan item(s)"))
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
