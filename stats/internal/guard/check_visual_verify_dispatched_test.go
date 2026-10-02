package guard

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// Every case of scripts/test-check-visual-verify-dispatched.sh at 2056def4
// (KAN-809's last commit to it), one subtest per ok: label, each pinning the
// whole stdout, stderr and exit the bash printed there, plus the branches
// that harness never reached. The harness's stub `flow` answered `record
// dispatches`, `record verdict` and `record verdicts` from canned files;
// here those are Env.Dispatches, Env.Verdict and Env.Verdicts, each logging
// the argv the CLI would have received, and the cases named "real flow" put
// a stub `flow` on Env's PATH instead, so the exec boundary is run too.
// The trigger question is visualTrigger, evaluated in-process against a real
// git repository's .flow/project.md, where the harness exec'd the real
// check-visual-trigger.sh.

const (
	vvdNotConfig = "VISUAL-VERIFY-OK: not configured\n"
	vvdNoUI      = "VISUAL-VERIFY-OK: no UI paths touched\n"
	vvdOK        = "VISUAL-VERIFY-OK: UI paths touched and a completed (or ended-but-unclosed) verifier dispatch is recorded for 'demo'\n"
	vvdMissing   = "VISUAL-VERIFY-MISSING: this change's diff touched a declared UI path but no completed or ended 'verifier' dispatch (key starting 'visual-verify') is recorded for 'demo' — flow.visual-verify's stage marks were written with no verifier ever dispatched, its report was never read to completion, or the dispatch's closing outcome was lost to a session restart\n"
	vvdSection   = "# project\n\n## visual verification\n\n| Setting | Value |\n|---------|-------|\n| `ui paths` | `app/src/**` |\n"
)

// vvdMaster is one master repository, built once: the harness's make_wt
// (with or without the declared section) plus its touch_paths commit.
type vvdMaster struct{ dir, base string }

// vvdMasters builds the harness's three sandbox shapes, plus "rename": a
// UI file moved out of the ui paths, which only --no-renames reports as
// touching one.
var vvdMasters = sync.OnceValues(func() (map[string]vvdMaster, error) {
	top, err := os.MkdirTemp(execFixtures.dir, "visual-verify-dispatched-master")
	if err != nil {
		return nil, err
	}
	git := func(dir string, args ...string) (string, error) {
		cmd := exec.Command(fixtureGit, append([]string{"-C", dir}, args...)...)
		cmd.Env = append(append(os.Environ(), fixtureGitEnv...),
			"GIT_AUTHOR_NAME=guard-test", "GIT_AUTHOR_EMAIL=guard-test@example.com",
			"GIT_COMMITTER_NAME=guard-test", "GIT_COMMITTER_EMAIL=guard-test@example.com")
		out, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("git %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out)), nil
	}
	write := func(path, body string) error {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		return os.WriteFile(path, []byte(body), 0o644)
	}
	shapes := []struct {
		name, project, baseFile string
		change                  func(dir string) error
	}{
		{"plain-docs", "# project\n", "", nil},
		{"declared-docs", vvdSection, "", nil},
		{"declared-ui", vvdSection, "", nil},
		{"rename", vvdSection, "app/src/Old.tsx", func(dir string) error {
			_, err := git(dir, "mv", "app/src/Old.tsx", "docs/New.tsx")
			return err
		}},
	}
	touched := map[string]string{"plain-docs": "docs/note.md", "declared-docs": "docs/note.md", "declared-ui": "app/src/Widget.tsx"}
	masters := map[string]vvdMaster{}
	for _, s := range shapes {
		dir := top + "/" + s.name
		if _, err := git(top, "init", "-q", dir); err != nil {
			return nil, err
		}
		if err := write(dir+"/.flow/project.md", s.project); err != nil {
			return nil, err
		}
		if s.baseFile != "" {
			if err := write(dir+"/"+s.baseFile, "x\n"); err != nil {
				return nil, err
			}
		}
		if _, err := git(dir, "add", "-A"); err != nil {
			return nil, err
		}
		if _, err := git(dir, "commit", "-q", "-m", "base"); err != nil {
			return nil, err
		}
		base, err := git(dir, "rev-parse", "HEAD")
		if err != nil {
			return nil, err
		}
		if p := touched[s.name]; p != "" {
			if err := write(dir+"/"+p, "x\n"); err != nil {
				return nil, err
			}
			if _, err := git(dir, "add", p); err != nil {
				return nil, err
			}
		}
		if s.change != nil {
			if err := os.MkdirAll(dir+"/docs", 0o755); err != nil {
				return nil, err
			}
			if err := s.change(dir); err != nil {
				return nil, err
			}
		}
		if _, err := git(dir, "commit", "-q", "-m", "change"); err != nil {
			return nil, err
		}
		masters[s.name] = vvdMaster{dir, base}
	}
	return masters, nil
})

// vvdRows is the harness's dispatch_json: rows from key/role/outcome triples.
func vvdRows(vals ...string) []byte {
	out := []map[string]string{}
	for i := 0; i+2 < len(vals); i += 3 {
		out = append(out, map[string]string{"key": vals[i], "role": vals[i+1], "outcome": vals[i+2]})
	}
	b, _ := json.Marshal(out)
	return b
}

type vvdCase struct {
	repo       string // vvdMasters key copied into the case's worktree
	dispatches []byte
	dErr       error
	decisions  []byte // Env.Decisions' answer; nil is "[]", no decision recorded
	decErr     error  // Env.Decisions fails
	flags      []byte // Env.Verdicts' answer; nil is "[]\n", the harness's flags.json
	fErr       error  // Env.Verdicts fails
	vErr       error  // Env.Verdict refuses the row
	flowStub   string // non-empty: every store hook nil, this `flow` on PATH answers them
	self       string // non-empty: FLOW_GUARD_SELF is this
	args       func(wt, base string) []string
	dir        func(t *testing.T) string // non-nil: the working directory
	relative   bool                      // the worktree passed relative to its parent, the working directory
}

type vvdResult struct {
	code             int
	out, err, wt, bn string
	base, cwd        string
	calls            string // every store call's argv, one line each: the hooks' log, or a real-flow stub's args file
}

// vvdRun copies the case's master into its own worktree and runs the guard
// in-process.
func vvdRun(t *testing.T, c vvdCase) vvdResult {
	t.Helper()
	fn := Registry["check-visual-verify-dispatched"]
	if fn == nil {
		t.Fatal("check-visual-verify-dispatched is not registered")
	}
	masters, err := vvdMasters()
	if err != nil {
		t.Fatal(err)
	}
	repo := c.repo
	if repo == "" {
		repo = "declared-ui"
	}
	m := masters[repo]
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gdcCopyTree(t, m.dir, wt)
	bin := t.TempDir()
	if c.flowStub != "" {
		writeExec(t, bin+"/flow", c.flowStub)
	}
	cwd := t.TempDir()
	if c.dir != nil {
		cwd = c.dir(t)
	}
	env := Env{Dir: cwd, Getenv: func(k string) string {
		switch k {
		case "PATH":
			return bin // no real `flow` is ever reachable from a test
		case "FLOW_GUARD_SELF":
			return c.self
		}
		return os.Getenv(k)
	}}
	var calls strings.Builder
	if c.flowStub == "" {
		flags := c.flags
		if flags == nil {
			flags = []byte("[]\n")
		}
		env.Dispatches = func(change string) ([]byte, error) {
			fmt.Fprintf(&calls, "record dispatches -change %s -C %s\n", change, wt)
			return c.dispatches, c.dErr
		}
		env.Decisions = func(change string) ([]byte, error) {
			fmt.Fprintf(&calls, "record decisions -change %s -C %s\n", change, wt)
			if c.decisions == nil {
				return []byte("[]"), c.decErr
			}
			return c.decisions, c.decErr
		}
		env.Verdict = func(change, guard, worktree, verdict string) error {
			fmt.Fprintf(&calls, "record verdict -change %s -guard %s -worktree %s -verdict %s -C %s\n", change, guard, worktree, verdict, wt)
			return c.vErr
		}
		env.Verdicts = func(guard string) ([]byte, error) {
			fmt.Fprintf(&calls, "record verdicts -guard %s -false-positive -C %s\n", guard, wt)
			return flags, c.fErr
		}
	}
	args := []string{wt, "demo", m.base}
	if c.relative {
		env.Dir, args[0] = filepath.Dir(wt), filepath.Base(wt)
	}
	if c.args != nil {
		args = c.args(wt, m.base)
	}
	var out, errb bytes.Buffer
	code := fn(args, env, &out, &errb)
	if c.flowStub != "" {
		b, _ := os.ReadFile(bin + "/args")
		calls.Write(b)
	}
	return vvdResult{code, out.String(), errb.String(), wt, bin, m.base, cwd, calls.String()}
}

// vvdCalls is the store calls a run through to a verdict makes, in order:
// the dispatches read, -- with decided -- the decisions read a run with no
// verifier makes, the verdict row, and -- with hint -- the false-positive
// read MISSING makes.
func vvdCalls(r vvdResult, verdict string, decided, hint bool) string {
	s := "record dispatches -change demo -C " + r.wt + "\n"
	if decided {
		s += "record decisions -change demo -C " + r.wt + "\n"
	}
	s += "record verdict -change demo -guard check-visual-verify-dispatched -worktree " + r.wt + " -verdict " + verdict + " -C " + r.wt + "\n"
	if hint {
		s += "record verdicts -guard check-visual-verify-dispatched -false-positive -C " + r.wt + "\n"
	}
	return s
}

// vvdFlowRows is a stub `flow` recording its argv beside itself, answering
// `record dispatches` with rows and `record verdicts` with flags, and
// refusing `record verdict` loudly on both streams -- output the guard must
// discard.
func vvdFlowRows(rows, flags string) string {
	return "#!/usr/bin/env bash\nprintf '%s\\n' \"$*\" >> \"$(dirname -- \"$0\")/args\"\n" +
		"if [ \"${1:-}\" = record ] && [ \"${2:-}\" = dispatches ]; then\n  printf '%s' '" + rows + "'\n  exit 0\nfi\n" +
		"if [ \"${1:-}\" = record ] && [ \"${2:-}\" = decisions ]; then\n  printf '%s' '[]'\n  exit 0\nfi\n" +
		"if [ \"${1:-}\" = record ] && [ \"${2:-}\" = verdicts ]; then\n  printf '%s' '" + flags + "'\n  exit 0\nfi\n" +
		"echo 'stub flow: refused'\necho 'stub flow: refused' >&2\nexit 2\n"
}

func TestCheckVisualVerifyDispatched(t *testing.T) {
	t.Parallel()
	const v, done = "verifier", "completed"
	p := vvdPrefix
	bad := vvdPrefix + "dispatch rows were not readable JSON — cannot answer\n"
	const (
		ended   = `[{"key":"visual-verify","role":"verifier","endedAt":"2026-09-20T10:00:00Z"}]`
		blocked = `[{"key":"visual-verify","role":"verifier","outcome":"blocked","endedAt":"2026-09-20T10:00:00Z"}]`
		open    = `[{"key":"visual-verify","role":"verifier"}]`
		flag    = `[{"falsePositive":true,"falsePositiveReason":"the verifier ran; the mark was lost to a session restart","change":"demo","flaggedAt":"2026-09-20T14:00:00Z"}]`
	)
	hint := func(n, last string) func(vvdResult) string {
		return func(vvdResult) string {
			return p + "prior false positives for this guard on this project: " + n + " — last: " + last + "\n"
		}
	}
	flagHint := hint("1", "the verifier ran; the mark was lost to a session restart (demo, 2026-09-20)")
	okCalls := func(r vvdResult) string { return vvdCalls(r, strings.TrimSuffix(vvdOK, "\n"), false, false) }
	missingCalls := func(r vvdResult) string { return vvdCalls(r, strings.TrimSuffix(vvdMissing, "\n"), true, true) }
	decision := func(visual string) []byte {
		return []byte(`[{"id":2,"decision":{"class":"small","visual":` + visual + `}},{"id":1,"decision":{"class":"small","visual":{"verify":"required","reason":"older row"}}}]`)
	}
	const skipReason = "actuator-only filter and deploy config; no page, no response a page consumes, no CORS/route the frontend uses"
	vvdSkipped := "VISUAL-VERIFY-OK: skipped at Decide — " + skipReason + "\n"
	skippedCalls := func(r vvdResult) string { return vvdCalls(r, strings.TrimSuffix(vvdSkipped, "\n"), true, false) }
	badDecision := p + "decision rows were not readable JSON — cannot answer\n"
	none := func(vvdResult) string { return "" }
	refused := errors.New("stub flow: refused")
	cases := []struct {
		label   string
		c       vvdCase
		want    int
		wantOut string
		wantErr func(r vvdResult) string // nil: no stderr
		calls   func(r vvdResult) string // non-nil: the store calls made, in order
	}{
		// ---- the harness's ok: labels ------------------------------------
		{"case 1: no declared section exits 0", vvdCase{repo: "plain-docs", dispatches: []byte("[]")}, 0, vvdNotConfig, nil, nil},
		{"case 1: the verdict is not-configured", vvdCase{repo: "plain-docs", dErr: errors.New("never read")}, 0, vvdNotConfig, nil, nil},
		{"case 2: no ui path touched exits 0", vvdCase{repo: "declared-docs", dispatches: []byte("[]")}, 0, vvdNoUI, nil, nil},
		{"case 2: the verdict is no-UI-paths", vvdCase{repo: "declared-docs", dErr: errors.New("never read")}, 0, vvdNoUI, nil, nil},
		{"case 3: completed verifier dispatch exits 0", vvdCase{dispatches: vvdRows("visual-verify-wt2", v, done)}, 0, vvdOK, nil, nil},
		{"case 3: the verdict is VISUAL-VERIFY-OK", vvdCase{dispatches: vvdRows("visual-verify-wt2", v, done)}, 0, vvdOK, nil, nil},
		{"case 3b: an ended row with no outcome exits 0", vvdCase{dispatches: []byte(ended)}, 0, vvdOK, nil, nil},
		{"case 3b: the verdict is VISUAL-VERIFY-OK", vvdCase{dispatches: []byte(ended)}, 0, vvdOK, nil, okCalls},
		{"case 4: an aborted dispatch is not evidence, exit 1", vvdCase{dispatches: vvdRows("visual-verify", v, "aborted")}, 1, vvdMissing, nil, nil},
		{"case 4: the verdict is VISUAL-VERIFY-MISSING", vvdCase{dispatches: vvdRows("visual-verify", v, "aborted")}, 1, vvdMissing, nil, nil},
		{"case 4b: a blocked row with an end is not evidence, exit 1", vvdCase{dispatches: []byte(blocked)}, 1, vvdMissing, nil, nil},
		{"case 4b: the verdict is VISUAL-VERIFY-MISSING", vvdCase{dispatches: []byte(blocked)}, 1, vvdMissing, nil, missingCalls},
		{"case 5: a non-verifier row is not evidence, exit 1", vvdCase{dispatches: vvdRows("visual-verify", "reviewer", done)}, 1, vvdMissing, nil, nil},
		{"case 6: an empty store is MISSING, exit 1", vvdCase{dispatches: []byte("[]")}, 1, vvdMissing, nil, nil},
		{"case 6b: an open row is not evidence, exit 1", vvdCase{dispatches: []byte(open)}, 1, vvdMissing, nil, nil},
		{"case 6b: the verdict is VISUAL-VERIFY-MISSING", vvdCase{dispatches: []byte(open)}, 1, vvdMissing, nil, missingCalls},
		{"case 7: a foreign key prefix is not evidence, exit 1", vvdCase{dispatches: vvdRows("some-other-key", v, done)}, 1, vvdMissing, nil, nil},
		{"case 8a: an unreachable store exits 2", vvdCase{dErr: errors.New("flow exit 1")}, 2, "",
			func(vvdResult) string { return p + "flow record dispatches failed for 'demo' — cannot answer\n" }, nil},
		{"case 8a: no verdict on an inability to answer", vvdCase{dErr: errors.New("flow exit 1")}, 2, "",
			func(vvdResult) string { return p + "flow record dispatches failed for 'demo' — cannot answer\n" }, nil},
		{"case 8b: non-JSON store output exits 2", vvdCase{dispatches: []byte("sorry, not json\n")}, 2, "",
			func(vvdResult) string { return bad }, nil},
		{"case 8c: a non-directory worktree exits 2",
			vvdCase{args: func(wt, base string) []string { return []string{wt + "/no-such-dir", "demo", base} }}, 2, "",
			func(r vvdResult) string { return p + "not a directory: " + r.wt + "/no-such-dir\n" }, nil},
		{"case 8d: an empty change name exits 2",
			vvdCase{args: func(wt, base string) []string { return []string{wt, "", base} }}, 2, "",
			func(vvdResult) string {
				return p + "usage: check-visual-verify-dispatched.sh <worktree> <change-name> <merge-base>\n"
			}, nil},
		{"case 8e: an empty merge-base exits 2",
			vvdCase{args: func(wt, _ string) []string { return []string{wt, "demo", ""} }}, 2, "",
			func(vvdResult) string {
				return p + "a merge-base is required — pass the state file's recorded value for this worktree\n"
			}, nil},
		{"case 8f: a change name outside the allowlist exits 2",
			vvdCase{args: func(wt, base string) []string { return []string{wt, "../evil", base} }}, 2, "",
			func(vvdResult) string {
				return p + "change name '../evil' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n"
			}, nil},

		{"case 9a: exactly one verdict row recorded on OK", vvdCase{dispatches: vvdRows("visual-verify-wt2", v, done)}, 0, vvdOK, nil, okCalls},
		{"case 9a: the recorded row names guard, change, worktree and verdict", vvdCase{dispatches: vvdRows("visual-verify-wt2", v, done)}, 0, vvdOK, nil, okCalls},
		{"case 9b: the MISSING verdict is recorded too", vvdCase{dispatches: vvdRows("visual-verify", v, "aborted")}, 1, vvdMissing, nil, missingCalls},
		{"case 10a: the not-configured verdict invokes no store call at all", vvdCase{repo: "plain-docs", dispatches: []byte("[]")}, 0, vvdNotConfig, nil, none},
		{"case 10b: the no-UI-paths verdict invokes no store call at all", vvdCase{repo: "declared-docs", dispatches: []byte("[]")}, 0, vvdNoUI, nil, none},
		{"case 11a: the advisory hint names count, reason, change and date", vvdCase{dispatches: []byte("[]"), flags: []byte(flag)}, 1, vvdMissing, flagHint, missingCalls},
		{"case 11a: the verdict still prints", vvdCase{dispatches: []byte("[]"), flags: []byte(flag)}, 1, vvdMissing, flagHint, nil},
		{"case 11b: no hint on an empty flag array", vvdCase{dispatches: []byte("[]")}, 1, vvdMissing, nil, missingCalls},
		{"case 11c: a failed verdicts read still reaches MISSING, exit 1", vvdCase{dispatches: []byte("[]"), flags: []byte(flag), fErr: refused}, 1, vvdMissing, nil, missingCalls},
		{"case 11c: no hint on a failed verdicts read", vvdCase{dispatches: []byte("[]"), flags: []byte(flag), fErr: refused}, 1, vvdMissing, nil, nil},
		{"case 12a: a refused write leaves the OK verdict standing", vvdCase{dispatches: vvdRows("visual-verify", v, done), vErr: refused}, 0, vvdOK, nil, okCalls},
		{"case 12a: the verdict is intact", vvdCase{dispatches: vvdRows("visual-verify", v, done), vErr: refused}, 0, vvdOK, nil, nil},
		{"case 12b: a refused write leaves the MISSING verdict standing", vvdCase{dispatches: vvdRows("visual-verify", "reviewer", done), vErr: refused}, 1, vvdMissing, nil, missingCalls},
		{"case 12b: the verdict is intact", vvdCase{dispatches: vvdRows("visual-verify", "reviewer", done), vErr: refused}, 1, vvdMissing, nil, nil},

		// ---- a visual-verification decision recorded at Decide ----------
		{"skipped at Decide with no verifier is OK, exit 0",
			vvdCase{dispatches: []byte("[]"), decisions: decision(`{"verify":"skipped","reason":"` + skipReason + `"}`)}, 0, vvdSkipped, nil, skippedCalls},
		{"a recorded verifier is OK before any decisions read",
			vvdCase{dispatches: vvdRows("visual-verify", v, done), decErr: refused}, 0, vvdOK, nil, okCalls},
		{"required at Decide with no verifier stays MISSING",
			vvdCase{dispatches: []byte("[]"), decisions: decision(`{"verify":"required","reason":"gateway CORS change the SPA calls"}`)}, 1, vvdMissing, nil, missingCalls},
		{"not configured at Decide with no verifier stays MISSING",
			vvdCase{dispatches: []byte("[]"), decisions: decision(`{"verify":"not configured","reason":"no section"}`)}, 1, vvdMissing, nil, missingCalls},
		{"only the newest decision row counts",
			vvdCase{dispatches: []byte("[]"), decisions: []byte(`[{"decision":{"visual":{"verify":"required","reason":"fix touched a page"}}},{"decision":{"visual":{"verify":"skipped","reason":"x"}}}]`)}, 1, vvdMissing, nil, missingCalls},
		{"a decision with no visual field (an older change) stays MISSING",
			vvdCase{dispatches: []byte("[]"), decisions: []byte(`[{"decision":{"class":"small"}}]`)}, 1, vvdMissing, nil, missingCalls},
		{"skipped with a blank reason is not a skip",
			vvdCase{dispatches: []byte("[]"), decisions: decision(`{"verify":"skipped","reason":"  "}`)}, 1, vvdMissing, nil, nil},
		{"a malformed visual value is not a skip",
			vvdCase{dispatches: []byte("[]"), decisions: decision(`"skipped"`)}, 1, vvdMissing, nil, nil},
		{"a multi-line reason folds onto the one verdict line",
			vvdCase{dispatches: []byte("[]"), decisions: decision(`{"verify":"skipped","reason":"actuator-only filter and deploy config;\nno page, no response a page consumes, no CORS/route the frontend uses"}`)}, 0, vvdSkipped, nil, nil},
		{"an unreachable decisions read exits 2, never read as skipped",
			vvdCase{dispatches: []byte("[]"), decErr: errors.New("flow exit 1")}, 2, "",
			func(vvdResult) string { return p + "flow record decisions failed for 'demo' — cannot answer\n" }, nil},
		{"non-JSON decisions exit 2",
			vvdCase{dispatches: []byte("[]"), decisions: []byte("sorry")}, 2, "", func(vvdResult) string { return badDecision }, nil},
		{"a non-object newest decision row exits 2",
			vvdCase{dispatches: []byte("[]"), decisions: []byte(`["skipped"]`)}, 2, "", func(vvdResult) string { return badDecision }, nil},
		{"real flow on PATH answers the decisions read",
			vvdCase{flowStub: "#!/usr/bin/env bash\nif [ \"$2\" = dispatches ]; then printf '[]'; exit 0; fi\nif [ \"$2\" = decisions ]; then printf '%s' '[{\"decision\":{\"visual\":{\"verify\":\"skipped\",\"reason\":\"" + skipReason + "\"}}}]'; exit 0; fi\nexit 2\n"}, 0, vvdSkipped, nil, nil},

		// ---- branches the harness never reached, pinned from the bash -----
		{"an ended row with an empty-string outcome is unclosed evidence",
			vvdCase{dispatches: []byte(`[{"key":"visual-verify","role":"verifier","outcome":"","endedAt":"x"}]`)}, 0, vvdOK, nil, okCalls},
		{"an ended row with a false outcome is unclosed evidence (jq's // reads false as absent)",
			vvdCase{dispatches: []byte(`[{"key":"visual-verify","role":"verifier","outcome":false,"endedAt":"x"}]`)}, 0, vvdOK, nil, nil},
		{"an endedAt of false is an end (jq's != null)",
			vvdCase{dispatches: []byte(`[{"key":"visual-verify","role":"verifier","endedAt":false}]`)}, 0, vvdOK, nil, nil},
		{"an endedAt of null is still open",
			vvdCase{dispatches: []byte(`[{"key":"visual-verify","role":"verifier","endedAt":null}]`)}, 1, vvdMissing, nil, nil},
		{"an ended row outside the key prefix is not evidence",
			vvdCase{dispatches: []byte(`[{"key":"visual","role":"verifier","endedAt":"x"}]`)}, 1, vvdMissing, nil, nil},
		{"an ended non-verifier row is not evidence",
			vvdCase{dispatches: []byte(`[{"key":"visual-verify","role":"reviewer","endedAt":"x"}]`)}, 1, vvdMissing, nil, nil},
		{"hint: a false-positive answer that is not an array prints nothing", vvdCase{dispatches: []byte("[]"), flags: []byte(`{"a":[1]}`)}, 1, vvdMissing, nil, nil},
		{"hint: a false-positive answer that is not JSON prints nothing", vvdCase{dispatches: []byte("[]"), flags: []byte(`[1] nope`)}, 1, vvdMissing, nil, nil},
		{"hint: two arrays print nothing", vvdCase{dispatches: []byte("[]"), flags: []byte(`[1] [2]`)}, 1, vvdMissing, nil, nil},
		{"hint: null fields render empty, a number flaggedAt as its text",
			vvdCase{dispatches: []byte("[]"), flags: []byte(`[{"falsePositiveReason":null,"change":false,"flaggedAt":20260920}]`)}, 1, vvdMissing, hint("1", " (, 20260920)"), nil},
		{"hint: an object flaggedAt is its compact text, key order kept, cut at ten",
			vvdCase{dispatches: []byte("[]"), flags: []byte(`[{"flaggedAt":{"b":1, "a":2}},{}]`)}, 1, vvdMissing, hint("2", ` (, {"b":1,"a")`), nil},
		{"hint: flaggedAt is cut at ten characters, not bytes",
			vvdCase{dispatches: []byte("[]"), flags: []byte(`[{"change":"c","flaggedAt":"ééééééééééxyz"}]`)}, 1, vvdMissing, hint("1", " (c, éééééééééé)"), nil},
		{"hint: a reason jq cannot add to a string renders last empty",
			vvdCase{dispatches: []byte("[]"), flags: []byte(`[{"falsePositiveReason":1}]`)}, 1, vvdMissing, hint("1", ""), nil},
		{"hint: an entry that is not an object renders last empty",
			vvdCase{dispatches: []byte("[]"), flags: []byte(`[[1]]`)}, 1, vvdMissing, hint("1", ""), nil},
		{"hint: one array among other values counts it and renders every value",
			vvdCase{dispatches: []byte("[]"), flags: []byte(`null [{"change":"x"}]`)}, 1, vvdMissing, hint("1", " (, )\n (x, )"), nil},
		{"hint: a value jq errors on before the last drops only its own line (jq exits on the last value's status)",
			vvdCase{dispatches: []byte("[]"), flags: []byte(`"s" [{"change":"x"}]`)}, 1, vvdMissing, hint("1", " (x, )"), nil},
		{"hint: an error on the last value renders last empty",
			vvdCase{dispatches: []byte("[]"), flags: []byte(`[{"change":"x"}] {}`)}, 1, vvdMissing, hint("1", ""), nil},
		{"hint: an empty entry array prints nothing", vvdCase{dispatches: []byte("[]"), flags: []byte("[]\n\n")}, 1, vvdMissing, nil, nil},
		{"no arguments names the missing worktree", vvdCase{args: func(string, string) []string { return nil }}, 2, "",
			func(vvdResult) string { return p + "not a directory: <missing>\n" }, nil},
		{"a relative worktree resolves against the working directory",
			vvdCase{dispatches: vvdRows("visual-verify", v, done), relative: true}, 0, vvdOK, nil, nil},
		{"a worktree argument read as a cd option is refused",
			vvdCase{args: func(_, base string) []string { return []string{"-x", "demo", base} },
				dir: func(t *testing.T) string { d := t.TempDir(); mkdir(t, d+"/-x"); return d }}, 2, "",
			func(vvdResult) string { return p + "worktree vanished: \n" }, nil},
		{"a merge-base that does not resolve is cannot-answer",
			vvdCase{args: func(wt, _ string) []string { return []string{wt, "demo", "nope"} }}, 2, "",
			func(r vvdResult) string {
				return p + "git diff failed against merge-base 'nope' in " + r.wt + " — cannot answer: fatal: ambiguous argument 'nope..HEAD': unknown revision or path not in the working tree.\nUse '--' to separate paths from revisions, like this:\n'git <command> [<revision>...] -- [<file>...]'\n"
			}, nil},
		{"a UI file renamed away still touches its UI path (--no-renames)", vvdCase{repo: "rename", dispatches: []byte("[]")}, 1, vvdMissing, nil, nil},
		{"a completed verifier row among others is found", vvdCase{dispatches: vvdRows("visual-verify", v, "aborted", "task-1-reviewer", "reviewer", done, "visual-verify-retry", v, done)}, 0, vvdOK, nil, nil},
		{"the outcome must be exactly completed", vvdCase{dispatches: vvdRows("visual-verify", v, "Completed")}, 1, vvdMissing, nil, nil},
		{"a null row selects nothing", vvdCase{dispatches: []byte(`[null,{"key":"visual-verify","role":"verifier","outcome":"completed"}]`)}, 0, vvdOK, nil, nil},
		{"rows missing fields read as empty", vvdCase{dispatches: []byte(`[{},{"role":"verifier","key":false,"outcome":"completed"},{"role":"verifier","key":null}]`)}, 1, vvdMissing, nil, nil},
		{"a non-string key on a non-verifier row is never read", vvdCase{dispatches: []byte(`[{"role":"reviewer","key":1}]`)}, 1, vvdMissing, nil, nil},
		{"a non-string key on a verifier row is cannot-answer", vvdCase{dispatches: []byte(`[{"role":"verifier","key":1}]`)}, 2, "", func(vvdResult) string { return bad }, nil},
		{"a non-object row is cannot-answer", vvdCase{dispatches: []byte(`["visual-verify"]`)}, 2, "", func(vvdResult) string { return bad }, nil},
		{"a top-level object is cannot-answer", vvdCase{dispatches: []byte(`{"a":{"key":"visual-verify","role":"verifier","outcome":"completed"}}`)}, 2, "", func(vvdResult) string { return bad }, nil},
		{"an empty store answer is cannot-answer", vvdCase{dispatches: []byte("")}, 2, "", func(vvdResult) string { return bad }, nil},
		{"two JSON values are cannot-answer", vvdCase{dispatches: []byte("[] []")}, 2, "", func(vvdResult) string { return bad }, nil},
		{"real flow on PATH answers the store read with the canonical worktree",
			vvdCase{flowStub: vvdFlowRows(`[{"key":"visual-verify","role":"verifier","outcome":"completed"}]`, "[]")}, 0, vvdOK, nil, okCalls},
		{"real flow: MISSING records its verdict, reads the prior false positives and hints",
			vvdCase{flowStub: vvdFlowRows("[]", flag)}, 1, vvdMissing, flagHint, missingCalls},
		{"real flow exiting non-zero is cannot-answer",
			vvdCase{flowStub: "#!/usr/bin/env bash\necho 'flow: connect: connection refused' >&2\nexit 1\n"}, 2, "",
			func(vvdResult) string { return p + "flow record dispatches failed for 'demo' — cannot answer\n" }, nil},
		{"real flow printing non-JSON is cannot-answer",
			vvdCase{flowStub: "#!/usr/bin/env bash\nprintf 'sorry, not json\\n'\n"}, 2, "", func(vvdResult) string { return bad }, nil},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()
			r := vvdRun(t, tc.c)
			wantErr := ""
			if tc.wantErr != nil {
				wantErr = tc.wantErr(r)
			}
			if r.code != tc.want || r.out != tc.wantOut || r.err != wantErr {
				t.Fatalf("got exit %d\nstdout %q\nstderr %q\nwant exit %d\nstdout %q\nstderr %q", r.code, r.out, r.err, tc.want, tc.wantOut, wantErr)
			}
			if tc.calls != nil {
				if want := tc.calls(r); r.calls != want {
					t.Fatalf("store calls\n%q\nwant\n%q", r.calls, want)
				}
			}
		})
	}
}

// TestCheckVisualVerifyDispatchedInProcessTrigger: the trigger question is
// answered in-process, so a shim whose checkout carries no
// scripts/check-visual-trigger.sh (FLOW_GUARD_SELF pointing into it) still
// reaches all three trigger outcomes -- fails if the guard execs a sibling
// again.
func TestCheckVisualVerifyDispatchedInProcessTrigger(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		label, repo string
		dispatches  []byte
		out         string
	}{
		{"not configured", "plain-docs", []byte("[]"), vvdNotConfig},
		{"no UI paths touched", "declared-docs", []byte("[]"), vvdNoUI},
		{"a dispatch verdict", "declared-ui", vvdRows("visual-verify", "verifier", "completed"), vvdOK},
	} {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()
			checkout := t.TempDir()
			mkdir(t, checkout+"/scripts")
			r := vvdRun(t, vvdCase{repo: tc.repo, dispatches: tc.dispatches, self: checkout + "/scripts/check-visual-verify-dispatched.sh"})
			if r.code != 0 || r.out != tc.out || r.err != "" {
				t.Fatalf("got exit %d stdout %q stderr %q, want 0 %q", r.code, r.out, r.err, tc.out)
			}
		})
	}
}
