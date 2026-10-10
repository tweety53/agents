package guard

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every case of scripts/test-check-task-reviewer-single-dispatch.sh at
// c5379c0a, one subtest per ok: label, plus the outputs that harness never
// asserted, pinned byte for byte as the bash printed them at c5379c0a. The
// harness's stub `flow` answered `record dispatches` from canned JSON; here
// that is Env.Dispatches. The decisions read has no Env hook, so a case that
// needs a class other than the default puts a stub `flow` on Env's PATH, and
// every other case runs with a PATH holding no `flow` at all -- the read
// fails and the class is `big`, as the harness's `[]` decisions made it.
// Group membership runs this checkout's REAL plan-dispatch-bundles.sh and
// plan-dispatch-groups.sh, as the harness did.

const trsdOK = "TASK-REVIEWER-SINGLE-DISPATCH-OK: every gated-per-task reviewer dispatch of this run is bundled by implementer group (bounds and retries included)\n"

// trsdPlan is the harness's fixture_tasks_md: two chains 1->3, 2->4 and a
// join 5, so the groups are `group 1: 1 3`, `group 2: 2 4`, `group 3: 5`.
var trsdPlan = func() string {
	var b strings.Builder
	task := func(id, files, after string) {
		b.WriteString("- [ ] " + id + ". Task " + id + "\n\n**Files:**\n- Create: `" + files + "`\n\n")
		if after != "" {
			b.WriteString("**After:** " + after + "\n\n")
		}
		b.WriteString("  - [ ] **Step 1: do it**\n\n")
	}
	task("1", "a.txt", "none")
	task("2", "b.txt", "none")
	task("3", "c.txt", "Task 1")
	task("4", "d.txt", "Task 2")
	task("5", "e.txt", "Task 3, 4")
	return b.String()
}()

type trsdCase struct {
	dispatches []byte
	dErr       error
	realFlow   bool   // Dispatches nil: the stub flow on PATH answers it
	decisions  string // non-empty: a stub flow on PATH answers `record decisions`
	plan       *string
	archived   bool              // the plan sits in the archived change directory, none live
	siblings   map[string]string // non-nil: FLOW_GUARD_SELF sits in a temp root's scripts/ beside these
	args       func(wt string) []string
}

// trsdStub is the stub flow: it records its argv beside itself and answers
// `record dispatches`/`record decisions` from the files next to it.
const trsdStub = `#!/usr/bin/env bash
d="$(dirname -- "$0")"
printf '%s\n' "$*" >> "$d/args"
case "$2" in
  dispatches) cat "$d/dispatches.json" ;;
  decisions) cat "$d/decisions.json" ;;
  *) exit 2 ;;
esac
`

// trsdRun runs the guard in-process and returns its exit, stdout, stderr,
// the canonical worktree and the stub's bin directory.
func trsdRun(t *testing.T, c trsdCase) (int, string, string, string, string) {
	t.Helper()
	fn := Registry["check-task-reviewer-single-dispatch"]
	if fn == nil {
		t.Fatal("check-task-reviewer-single-dispatch is not registered")
	}
	wt, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	plan := trsdPlan
	if c.plan != nil {
		plan = *c.plan
	}
	if plan != "" {
		dir := wt + "/spectre/changes/demo"
		if c.archived {
			dir = wt + "/spectre/changes/archive/demo"
		}
		writeFile(t, dir+"/tasks.md", plan)
	}
	bin := t.TempDir()
	if c.realFlow || c.decisions != "" {
		writeExec(t, bin+"/flow", trsdStub)
		writeFile(t, bin+"/dispatches.json", string(c.dispatches))
		writeFile(t, bin+"/decisions.json", c.decisions)
	}
	root := fpRepoRoot
	if c.siblings != nil {
		root = t.TempDir()
		for name, body := range c.siblings {
			writeExec(t, root+"/scripts/"+name, body)
		}
	}
	env := Env{Dir: t.TempDir(), Getenv: func(k string) string {
		switch k {
		case "PATH":
			return bin // no real `flow` is ever reachable from a test
		case "FLOW_GUARD_SELF":
			return root + "/scripts/check-task-reviewer-single-dispatch.sh"
		}
		return os.Getenv(k)
	}}
	if !c.realFlow {
		env.Dispatches = func(string) ([]byte, error) { return c.dispatches, c.dErr }
	}
	args := []string{wt, "demo", "mf-tok"}
	if c.args != nil {
		args = c.args(wt)
	}
	var out, errb bytes.Buffer
	code := fn(args, env, &out, &errb)
	return code, out.String(), errb.String(), wt, bin
}

func TestCheckTaskReviewerSingleDispatch(t *testing.T) {
	t.Parallel()
	const r, tok = "reviewer", "mf-tok"
	d := pfdDispatches
	const p = "check-task-reviewer-single-dispatch: "
	empty := ""
	noTasks := "no tasks here\n"
	cases := []struct {
		label   string
		c       trsdCase
		want    int
		wantOut string
		wantErr func(wt string) string // nil: no stderr
	}{
		{"case 1: one bundle per group exits 0 with no stderr",
			trsdCase{dispatches: d("task-1+3-reviewer", r, tok, "task-2+4-reviewer", r, tok, "task-5-reviewer", r, tok)}, 0, trsdOK, nil},
		{"case 2: original plus -retry exits 0",
			trsdCase{dispatches: d("task-1+3-reviewer", r, tok, "task-1+3-reviewer-retry", r, tok)}, 0, trsdOK, nil},
		{"case 3: two originals under one key exit 1, naming it",
			trsdCase{dispatches: d("task-1+3-reviewer", r, tok, "task-1+3-reviewer", r, tok)}, 1, "",
			func(string) string {
				return p + "bundle task-1+3-reviewer carries 2 reviewer dispatches -- the bundle goes to ONE dispatch, never one per task, slot or finding\n"
			}},
		{"case 4: same-group tasks split across two bundles exit 1",
			trsdCase{dispatches: d("task-1-reviewer", r, tok, "task-3-reviewer", r, tok)}, 1, "",
			func(string) string {
				return p + "tasks in group 1 are split across reviewer bundles 'task-1-reviewer' and 'task-3-reviewer' -- every gate-fired task of one implementer group goes out in ONE reviewer dispatch\n"
			}},
		{"case 5: malformed key exits 1, naming it verbatim",
			trsdCase{dispatches: d("task-1-reviewer-oops", r, tok)}, 1, "",
			func(string) string {
				return p + "out-of-shape gated-reviewer key 'task-1-reviewer-oops' -- the canonical shape is task-<n[+n+n...]>-reviewer[-fix-<k>][-retry]\n"
			}},
		{"case 6: one gated reviewer bundle on class small exits 1 (KAN-934: the panel reviews gate-fired tasks)",
			trsdCase{dispatches: d("task-1+2-reviewer", r, tok), decisions: `[{"decision":{"class":"small"}}]`}, 1, "",
			func(string) string {
				return p + "class 'small' carries 1 gated reviewer dispatch(es) (task-1+2-reviewer) -- on micro/small/regular no gated reviewer runs; the whole-branch panel reviews every gate-fired task\n"
			}},
		{"a gated reviewer bundle and its fix round on class micro exit 1",
			trsdCase{dispatches: d("task-1+3-reviewer", r, tok, "task-1+3-reviewer-fix-1", r, tok), decisions: `[{"decision":{"class":"micro"}}]`}, 1, "",
			func(string) string {
				return p + "class 'micro' carries 2 gated reviewer dispatch(es) (task-1+3-reviewer task-1+3-reviewer-fix-1) -- on micro/small/regular no gated reviewer runs; the whole-branch panel reviews every gate-fired task\n"
			}},
		{"no gated reviewer rows on class small exits 0",
			trsdCase{dispatches: d("panel-1-primary+principles", r, tok), decisions: `[{"decision":{"class":"small"}}]`}, 0, trsdOK, nil},
		{"case 7: two original bundles on class big (default) exits 0",
			trsdCase{dispatches: d("task-1+3-reviewer", r, tok, "task-2+4-reviewer", r, tok)}, 0, trsdOK, nil},
		{"case 8: original plus its own fix-round bundle exits 0",
			trsdCase{dispatches: d("task-1+3-reviewer", r, tok, "task-1+3-reviewer-fix-1", r, tok)}, 0, trsdOK, nil},
		{"case 9: unreachable store exits 2",
			trsdCase{dErr: errors.New("exit status 1")}, 2, "",
			func(string) string { return p + "flow record dispatches failed for 'demo' -- cannot answer\n" }},
		{"case 10: missing tasks.md exits 2",
			trsdCase{dispatches: []byte("[]"), plan: &empty}, 2, "",
			func(wt string) string {
				return p + "no tasks.md at " + wt + "/spectre/changes/demo/tasks.md -- cannot answer\n"
			}},
		{"a change run 1 archived reads its archived plan and exits 0",
			trsdCase{dispatches: d("task-1+3-reviewer", r, tok, "task-2+4-reviewer", r, tok, "task-5-reviewer", r, tok), archived: true}, 0, trsdOK, nil},
		{"case 11: empty token is a usage error, exits 2",
			trsdCase{dispatches: d("task-1+3-reviewer", r, tok), args: func(wt string) []string { return []string{wt, "demo", ""} }}, 2, "",
			func(string) string {
				return p + "session token is required and must be this run's own literal token\n"
			}},

		// Pinned beyond the harness, against the bash at c5379c0a.
		{"no arguments exits 2 naming <missing>",
			trsdCase{args: func(string) []string { return nil }}, 2, "",
			func(string) string { return p + "not a directory: <missing>\n" }},
		{"missing change name is a usage error, exits 2",
			trsdCase{args: func(wt string) []string { return []string{wt} }}, 2, "",
			func(string) string {
				return p + "usage: check-task-reviewer-single-dispatch.sh <worktree> <change-name> <session-token>\n"
			}},
		{"change name outside the allowlist exits 2",
			trsdCase{args: func(wt string) []string { return []string{wt, "../x", "mf-tok"} }}, 2, "",
			func(string) string {
				return p + "change name '../x' is not a plain change name — it must start with a letter or digit and contain only letters, digits, '.', '_' and '-'\n"
			}},
		{"no reviewer rows exits 0",
			trsdCase{dispatches: []byte("[]")}, 0, trsdOK, nil},
		{"retry-only and over-bound bundles exit 1, each named in first-seen order",
			trsdCase{dispatches: d("task-5-reviewer-retry", r, tok, "task-5-reviewer-retry", r, tok, "task-4-reviewer", r, tok,
				"task-4-reviewer-retry", r, tok, "task-4-reviewer-retry", r, tok)}, 1, "",
			func(string) string {
				return p + "bundle task-5-reviewer carries 2 -retry dispatch(es) and no original -- a retry is never a bundle's only dispatch\n" +
					p + "bundle task-4-reviewer carries 3 reviewer dispatches -- at most the original plus the handshake's one -retry\n"
			}},
		{"a gated reviewer bundle on the newest decision's class regular exits 1",
			trsdCase{dispatches: d("task-5-reviewer", r, tok),
				decisions: `[{"decision":{"class":"regular"}},{"decision":{"class":"big"}}]`}, 1, "",
			func(string) string {
				return p + "class 'regular' carries 1 gated reviewer dispatch(es) (task-5-reviewer) -- on micro/small/regular no gated reviewer runs; the whole-branch panel reviews every gate-fired task\n"
			}},
		{"foreign token, other role and a task outside the plan are ignored",
			trsdCase{dispatches: d("task-1-reviewer", r, tok, "task-3-reviewer", r, "other", "task-3-reviewer", "implementer", tok, "task-9-reviewer", r, tok)}, 0, trsdOK, nil},
		{"unreadable dispatch rows exit 2",
			trsdCase{dispatches: []byte("not json")}, 2, "",
			func(string) string { return p + "dispatch rows were not readable JSON -- cannot answer\n" }},
		// The bash also printed jq's own error line first; there is no jq.
		{"dispatch rows that are not an array of rows exit 2",
			trsdCase{dispatches: []byte(`{"a":1}`)}, 2, "",
			func(string) string { return p + "dispatch rows were not readable JSON -- cannot answer\n" }},
		// Rows the read returned but that do not parse fail closed: the bash
		// body exited 5 with jq's error line under set -e, and `big` is the
		// looser class here, so reading them as `big` would pass what the
		// bash blocked.
		{"unreadable decision rows exit 2",
			trsdCase{dispatches: d("task-1+3-reviewer", r, tok, "task-2+4-reviewer", r, tok), decisions: `[1]`}, 2, "",
			func(string) string {
				return p + "decision output was JSON but not an array of decision rows -- cannot answer\n"
			}},
		// The bash's `jq empty` gate left the class at `big` for output that
		// is not JSON at all -- kept.
		{"decision output that is not JSON reads as class big",
			trsdCase{dispatches: d("task-1+3-reviewer", r, tok, "task-2+4-reviewer", r, tok), decisions: `garbage`}, 0, trsdOK, nil},
		// jq iterated a top-level object's values (exit 1 here); the verb
		// prints an array, so an object is refused, never read as `big`.
		{"decision rows that are not an array exit 2",
			trsdCase{dispatches: d("task-1+3-reviewer", r, tok, "task-2+4-reviewer", r, tok), decisions: `{"a":{"decision":{"class":"small"}}}`}, 2, "",
			func(string) string {
				return p + "decision output was JSON but not an array of decision rows -- cannot answer\n"
			}},
		{"a plan with no tasks exits 0",
			trsdCase{dispatches: d("task-1-reviewer", r, tok, "task-3-reviewer", r, tok), plan: &noTasks}, 0, trsdOK, nil},
		{"plan-dispatch-bundles.sh failing exits 2 with its combined output",
			trsdCase{dispatches: []byte("[]"), siblings: map[string]string{
				"plan-dispatch-bundles.sh": "#!/usr/bin/env bash\necho 'bundles broke' >&2\necho 'bundle 1: 1'\nexit 3\n",
				"plan-dispatch-groups.sh":  "#!/usr/bin/env bash\necho 'group 1: 1'\n"}}, 2, "",
			func(wt string) string {
				return p + "plan-dispatch-bundles.sh failed on " + wt + "/spectre/changes/demo/tasks.md -- cannot answer\nbundles broke\nbundle 1: 1\n"
			}},
		{"plan-dispatch-groups.sh failing silently exits 2 with an empty line",
			trsdCase{dispatches: []byte("[]"), siblings: map[string]string{
				"plan-dispatch-bundles.sh": "#!/usr/bin/env bash\necho 'bundle 1: 1'\n",
				"plan-dispatch-groups.sh":  "#!/usr/bin/env bash\nexit 4\n"}}, 2, "",
			func(wt string) string {
				return p + "plan-dispatch-groups.sh failed on " + wt + "/spectre/changes/demo/tasks.md -- cannot answer\n\n"
			}},
		// The bash exported LC_ALL=C, so both siblings ran under it whatever
		// the caller's locale.
		{"plan-dispatch siblings run under LC_ALL=C",
			trsdCase{dispatches: []byte("[]"), siblings: map[string]string{
				"plan-dispatch-bundles.sh": "#!/usr/bin/env bash\necho \"LC_ALL=${LC_ALL:-unset}\"\nexit 1\n",
				"plan-dispatch-groups.sh":  "#!/usr/bin/env bash\necho 'group 1: 1'\n"}}, 2, "",
			func(wt string) string {
				return p + "plan-dispatch-bundles.sh failed on " + wt + "/spectre/changes/demo/tasks.md -- cannot answer\nLC_ALL=C\n"
			}},
		{"real flow: the store is read with -change and the canonical worktree",
			trsdCase{realFlow: true, dispatches: d("task-1-reviewer", r, tok, "task-3-reviewer", r, tok)}, 1, "",
			func(string) string {
				return p + "tasks in group 1 are split across reviewer bundles 'task-1-reviewer' and 'task-3-reviewer' -- every gate-fired task of one implementer group goes out in ONE reviewer dispatch\n"
			}},
	}
	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			t.Parallel()
			got, out, errs, wt, bin := trsdRun(t, tc.c)
			wantErr := ""
			if tc.wantErr != nil {
				wantErr = tc.wantErr(wt)
			}
			if got != tc.want || out != tc.wantOut || errs != wantErr {
				t.Fatalf("got exit %d, stdout %q, stderr %q\nwant exit %d, stdout %q, stderr %q", got, out, errs, tc.want, tc.wantOut, wantErr)
			}
			if tc.c.realFlow {
				args, err := os.ReadFile(bin + "/args")
				if err != nil {
					t.Fatal(err)
				}
				want := "record dispatches -change demo -C " + wt + "\nrecord decisions -change demo -C " + wt + "\n"
				if string(args) != want {
					t.Fatalf("flow argv %q, want %q", args, want)
				}
			}
		})
	}
}
