package guard

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Every case of scripts/test-check-panel-findings-closed.sh, one subtest per
// ok: label, plus the Minor-deferral cases (the rule in review-panel.md's
// **Panel re-runs**) and the fixed-without-clean-rerun cases (the ordering
// rule in review-panel.md's **Recording findings**). The bash harness's stub
// `flow` printed a canned JSON array for `record findings`; here that stub
// sits on Env's PATH the same way, answers `record dispatches` from its own
// file, and every case runs the production exec path.

// cfcFinding is one row of `flow record findings`' JSON; an empty Severity
// or Slot is left out of the row, as the bash harness's ref/status pairs
// left it.
type cfcFinding struct {
	Ref, Status, Severity string
	Round                 int
	Slot, Supersedes      string
	Reproducer            string
}

func cfcJSON(fs ...cfcFinding) string {
	out := []map[string]any{}
	for _, f := range fs {
		row := map[string]any{"ref": f.Ref, "status": f.Status, "round": f.Round}
		if f.Severity != "" {
			row["severity"] = f.Severity
		}
		if f.Slot != "" {
			row["slot"] = f.Slot
		}
		if f.Supersedes != "" {
			row["supersedes"] = f.Supersedes
		}
		if f.Reproducer != "" {
			row["reproducer"] = f.Reproducer
		}
		out = append(out, row)
	}
	b, _ := json.Marshal(out)
	return string(b)
}

// cfcDispatch is one row of `flow record dispatches`' JSON.
type cfcDispatch struct {
	Key, Role, Slot, Outcome string
}

func cfcDispatchJSON(ds ...cfcDispatch) string {
	out := []map[string]any{}
	for _, d := range ds {
		out = append(out, map[string]any{"key": d.Key, "role": d.Role, "slot": d.Slot, "outcome": d.Outcome})
	}
	b, _ := json.Marshal(out)
	return string(b)
}

// cfcDefaultDispatches is the dispatch body every table case gets unless it
// names its own: one clean re-run of slot primary at each of rounds 1-3, so
// a case about the open or deferral predicates never trips the
// fixed-without-clean-rerun one.
var cfcDefaultDispatches = cfcDispatchJSON(
	cfcDispatch{Key: "panel-1-primary", Role: "reviewer", Slot: "primary", Outcome: "completed"},
	cfcDispatch{Key: "panel-2-primary", Role: "reviewer", Slot: "primary", Outcome: "completed"},
	cfcDispatch{Key: "panel-3-primary", Role: "reviewer", Slot: "primary", Outcome: "completed"},
)

// cfcBin is a directory holding a stub `flow` that prints stderrLine (when
// set) on stderr, answers `record dispatches` with cfcDefaultDispatches,
// every other call with body, and exits rc.
func cfcBin(t *testing.T, body, stderrLine string, rc int) string {
	t.Helper()
	return cfcBinDispatches(t, body, cfcDefaultDispatches, stderrLine, rc)
}

// cfcBinDispatches is cfcBin with the dispatch body spelled.
func cfcBinDispatches(t *testing.T, body, dispatches, stderrLine string, rc int) string {
	t.Helper()
	bin := t.TempDir()
	writeFile(t, filepath.Join(bin, "findings.json"), body)
	writeFile(t, filepath.Join(bin, "dispatches.json"), dispatches)
	writeFile(t, filepath.Join(bin, "state.json"), `{"worktrees":{}}`)
	script := "#!/usr/bin/env bash\ndir=\"$(dirname -- \"$0\")\"\n"
	if stderrLine != "" {
		script += "echo '" + stderrLine + "' >&2\n"
	}
	script += "case \"$*\" in\n"
	script += "  *\"record dispatches\"*) cat \"$dir/dispatches.json\" ;;\n"
	script += "  *\"record mutations\"*) [ -f \"$dir/mutations.fail\" ] && exit 1; if [ -f \"$dir/mutations.json\" ]; then cat \"$dir/mutations.json\"; else cat \"$dir/findings.json\"; fi ;;\n"
	script += "  *\"state get\"*) cat \"$dir/state.json\" ;;\n"
	script += "  *) cat \"$dir/findings.json\" ;;\n"
	script += "esac\nexit " + strconv.Itoa(rc) + "\n"
	writeExec(t, filepath.Join(bin, "flow"), script)
	return bin
}

func cfcRun(t *testing.T, bin string, args ...string) (int, string, string) {
	t.Helper()
	fn := Registry["check-panel-findings-closed"]
	if fn == nil {
		t.Fatal("check-panel-findings-closed is not registered")
	}
	var out, errb bytes.Buffer
	rc := fn(args, Env{Getenv: pathEnv(bin), Dir: t.TempDir()}, &out, &errb)
	return rc, out.String(), errb.String()
}

// A fixed Critical or Important round with no fix-mutation row is refused
// at the panel's close, not first at integrate's unfinished-work gate.
func TestCheckPanelFindingsClosedNeedsAFixMutationRow(t *testing.T) {
	t.Parallel()
	const line = "no fix-mutation row"
	important := cfcJSON(cfcFinding{Ref: "F1", Severity: "Important", Status: "fixed", Round: 1, Slot: "primary"})
	run := func(findings, muts string, fail bool) (int, string) {
		bin := cfcBinDispatches(t, findings, "[]", "", 0)
		writeFile(t, filepath.Join(bin, "mutations.json"), muts)
		if fail {
			writeFile(t, filepath.Join(bin, "mutations.fail"), "")
		}
		rc, _, errOut := cfcRun(t, bin, t.TempDir(), "demo")
		return rc, errOut
	}
	if rc, errOut := run(important, "[]", false); rc != 1 || !strings.Contains(errOut, line) || !strings.Contains(errOut, "round(s) 1 ") {
		t.Fatalf("no row: rc %d, stderr %q", rc, errOut)
	}
	if _, errOut := run(important, `[{"round":2,"path":"a","mutated":"none","test":"doc-only"}]`, false); strings.Contains(errOut, line) {
		t.Fatalf("a round-2 row proves round 1: stderr %q", errOut)
	}
	minor := cfcJSON(cfcFinding{Ref: "F1", Severity: "Minor", Status: "fixed", Round: 1, Slot: "primary"})
	if _, errOut := run(minor, "[]", false); strings.Contains(errOut, line) {
		t.Fatalf("a Minor fix needs no row: stderr %q", errOut)
	}
	if rc, errOut := run(important, "[]", true); rc != 2 {
		t.Fatalf("a failed mutations read: rc %d, stderr %q", rc, errOut)
	}
}

func TestCheckPanelFindingsClosed(t *testing.T) {
	t.Parallel()
	wt := t.TempDir()
	// Every clean re-run a passing case's dispatches name writes its report
	// (KAN-915); no count lines, so the count check skips them.
	sdd := filepath.Join(wt, ".superpowers", "sdd")
	if err := os.MkdirAll(sdd, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, r := range []string{"1", "2", "3"} {
		for _, id := range []string{"primary", "bugbot"} {
			writeFile(t, filepath.Join(sdd, "panel-report-"+r+"-"+id+".md"), "clean\n")
		}
	}
	// Every fixed Critical/Important a case records names its regression
	// test's none form, so no case trips the regression-test check.
	writeFile(t, filepath.Join(sdd, "panel-fix-report-1.md"), "regression-test: F1 — none — fixture\nregression-test: F2 — none — fixture\n")
	cases := []struct {
		label  string
		json   string
		disp   string
		want   int
		needle []string
		clean  bool   // no stderr at all
		absent string // must not appear on stderr
	}{
		{label: "case 1: every finding closed exits 0 with no stderr", json: cfcJSON(cfcFinding{Ref: "F1", Status: "fixed", Slot: "primary"}, cfcFinding{Ref: "F2", Status: "fixed", Slot: "primary"}), want: 0, needle: []string{"FINDINGS-CLOSED"}, clean: true},
		{label: "case 2: one still-open finding exits 1, naming it", json: cfcJSON(cfcFinding{Ref: "F1", Status: "open"}, cfcFinding{Ref: "F2", Status: "fixed", Slot: "primary"}), want: 1, needle: []string{"check-panel-findings-closed: finding(s) still open: F1\n"}},
		{label: "case 3: several still-open findings exit 1, naming both", json: cfcJSON(cfcFinding{Ref: "F1", Status: "open"}, cfcFinding{Ref: "F2", Status: "fixed", Slot: "primary"}, cfcFinding{Ref: "F3", Status: "open"}), want: 1, needle: []string{"finding(s) still open: F1 F3\n"}},
		{label: "case 4: a withdrawn finding counts as closed, exits 0", json: cfcJSON(cfcFinding{Ref: "F1", Status: "withdrawn — not a real defect"}), want: 0},
		{label: "case 4a: a bare withdrawn finding is open, exits 1, naming it (KAN-791)", json: cfcJSON(cfcFinding{Ref: "F1", Status: "withdrawn"}), want: 1, needle: []string{"check-panel-findings-closed: finding(s) still open: F1\n"}},
		{label: "case 4a2: a whitespace-only withdrawal reason is open, exits 1", json: cfcJSON(cfcFinding{Ref: "F1", Status: "withdrawn   "}), want: 1, needle: []string{"finding(s) still open: F1\n"}},
		{label: "case 4b: a deferred finding is open, exits 1, naming it", json: cfcJSON(cfcFinding{Ref: "F1", Status: "deferred cosmetic, not worth a fix round"}), want: 1, needle: []string{"check-panel-findings-closed: finding(s) still open: F1\n"}},
		{label: "case 5: zero findings exits 0", json: "[]", want: 0},
		{label: "case 11: Minors deferred beside an Important of the same round exit 1, naming both as open",
			json: cfcJSON(cfcFinding{Ref: "F1", Severity: "Important", Status: "fixed", Slot: "primary"}, cfcFinding{Ref: "F2", Severity: "Minor", Status: "deferred cosmetic"}, cfcFinding{Ref: "F3", Severity: "minor", Status: "deferred doc-only"}),
			want: 1, needle: []string{"check-panel-findings-closed: finding(s) still open: F2 F3\n"}},
		{label: "case 12: a Minor deferred beside a Critical of the same round exits 1, naming it as open",
			json: cfcJSON(cfcFinding{Ref: "F1", Severity: "CRITICAL", Status: "fixed", Round: 2, Slot: "primary"}, cfcFinding{Ref: "F2", Severity: "Minor", Status: "deferred cosmetic", Round: 2}),
			want: 1, needle: []string{"finding(s) still open: F2\n"}},
		{label: "case 13: a Minor-only round with its Minors deferred exits 1, naming both as open",
			json: cfcJSON(cfcFinding{Ref: "F1", Severity: "Minor", Status: "deferred cosmetic"}, cfcFinding{Ref: "F2", Severity: "Minor", Status: "deferred doc-only"}),
			want: 1, needle: []string{"finding(s) still open: F1 F2\n"}},
		{label: "case 14: Important and Minor of one round, the Minor fixed, exits 0",
			json: cfcJSON(cfcFinding{Ref: "F1", Severity: "Important", Status: "fixed", Slot: "primary"}, cfcFinding{Ref: "F2", Severity: "Minor", Status: "fixed", Slot: "primary"}),
			want: 0, needle: []string{"FINDINGS-CLOSED"}, clean: true},
		{label: "case 15: a Minor deferred in a round whose Important is in another round exits 1, naming it as open",
			json: cfcJSON(cfcFinding{Ref: "F1", Severity: "Important", Status: "fixed", Round: 0, Slot: "primary"}, cfcFinding{Ref: "F2", Severity: "Minor", Status: "deferred cosmetic", Round: 1}),
			want: 1, needle: []string{"finding(s) still open: F2\n"}},
		{label: "case 16: a Minor deferred beside a withdrawn Important exits 1, naming it as open",
			json: cfcJSON(cfcFinding{Ref: "F1", Severity: "Important", Status: "withdrawn — premise not in the primary source"}, cfcFinding{Ref: "F2", Severity: "Minor", Status: "deferred cosmetic"}),
			want: 1, needle: []string{"finding(s) still open: F2\n"}},
		{label: "case 17: an open finding and a deferred Minor are both reported open",
			json: cfcJSON(cfcFinding{Ref: "F1", Severity: "Important", Status: "open"}, cfcFinding{Ref: "F2", Severity: "Minor", Status: "deferred cosmetic"}),
			want: 1, needle: []string{"finding(s) still open: F1 F2\n"}},
		{label: "case 18: a non-string status cannot be judged, exits 2", json: `[{"ref":"F1","status":null}]`, want: 2, needle: []string{"check-panel-findings-closed: jq failed — cannot determine anything\n"}},
		{label: "case 19: empty store output cannot be judged, exits 2", json: "", want: 2, needle: []string{"jq failed — cannot determine anything"}},
		{label: "case 20: a fixed finding with no dispatch rows at all exits 1, naming it and its slot",
			json: cfcJSON(cfcFinding{Ref: "F1", Status: "fixed", Slot: "primary"}), disp: "[]",
			want: 1, needle: []string{"recorded fixed with no clean re-run dispatch of their slot in any later round: F1 (primary)\n"}},
		{label: "case 21: a later re-run of another slot does not cover the finding",
			json: cfcJSON(cfcFinding{Ref: "F1", Status: "fixed", Slot: "primary"}),
			disp: cfcDispatchJSON(cfcDispatch{Key: "panel-1-bugbot", Role: "reviewer", Slot: "bugbot", Outcome: "completed"}),
			want: 1, needle: []string{"no clean re-run dispatch of their slot in any later round: F1 (primary)\n"}},
		{label: "case 22: a re-run at the finding's own round or earlier does not cover it",
			json: cfcJSON(cfcFinding{Ref: "F1", Status: "fixed", Round: 1, Slot: "primary"}),
			disp: cfcDispatchJSON(cfcDispatch{Key: "panel-1-primary", Role: "reviewer", Slot: "primary", Outcome: "completed"}, cfcDispatch{Key: "panel-0-primary", Role: "reviewer", Slot: "primary", Outcome: "completed"}),
			want: 1, needle: []string{"no clean re-run dispatch of their slot in any later round: F1 (primary)\n"}},
		{label: "case 23: a re-run that did not end completed does not cover the finding",
			json: cfcJSON(cfcFinding{Ref: "F1", Status: "fixed", Slot: "primary"}),
			disp: cfcDispatchJSON(cfcDispatch{Key: "panel-1-primary", Role: "reviewer", Slot: "primary", Outcome: "timed-out"}),
			want: 1, needle: []string{"no clean re-run dispatch of their slot in any later round: F1 (primary)\n"}},
		{label: "case 24: a bundle re-run covering every + component satisfies a joined finding",
			json: cfcJSON(cfcFinding{Ref: "F1", Status: "fixed", Slot: "primary+bugbot"}),
			disp: cfcDispatchJSON(cfcDispatch{Key: "panel-1-bugbot+primary", Role: "reviewer", Slot: "bugbot+primary", Outcome: "completed"}),
			want: 0, needle: []string{"FINDINGS-CLOSED"}, clean: true},
		{label: "case 24b: a re-run covering only one + component of a joined finding does not satisfy it",
			json: cfcJSON(cfcFinding{Ref: "F1", Status: "fixed", Slot: "primary+bugbot"}),
			disp: cfcDispatchJSON(cfcDispatch{Key: "panel-1-primary", Role: "reviewer", Slot: "primary", Outcome: "completed"}),
			want: 1, needle: []string{"no clean re-run dispatch of their slot in any later round: F1 (primary+bugbot)\n"}},
		{label: "case 25: a handshake-retry key counts as the slot's clean re-run",
			json: cfcJSON(cfcFinding{Ref: "F1", Status: "fixed", Slot: "primary"}),
			disp: cfcDispatchJSON(cfcDispatch{Key: "panel-1-primary-retry", Role: "reviewer", Slot: "primary", Outcome: "completed"}),
			want: 0, needle: []string{"FINDINGS-CLOSED"}, clean: true},
		{label: "case 26: a panel-fix or implementer dispatch is never the slot's re-run",
			json: cfcJSON(cfcFinding{Ref: "F1", Status: "fixed", Slot: "primary"}),
			disp: cfcDispatchJSON(cfcDispatch{Key: "panel-fix-1", Role: "panel-fix", Slot: "", Outcome: "completed"}, cfcDispatch{Key: "task-1", Role: "implementer", Slot: "primary", Outcome: "completed"}),
			want: 1, needle: []string{"no clean re-run dispatch of their slot in any later round: F1 (primary)\n"}},
		{label: "case 27: two fixed findings without a re-run are both named",
			json: cfcJSON(cfcFinding{Ref: "F1", Status: "fixed", Slot: "primary"}, cfcFinding{Ref: "F2", Status: "fixed", Slot: "bugbot"}), disp: "[]",
			want: 1, needle: []string{"in any later round: F1 (primary) F2 (bugbot)\n"}},
		{label: "case 28: unreadable dispatch rows cannot be judged, exits 2",
			json: cfcJSON(cfcFinding{Ref: "F1", Status: "fixed", Slot: "primary"}), disp: `[{"key":`,
			want: 2, needle: []string{"jq failed — cannot determine anything"}},
		{label: "case 29: a fixed Minor with no later re-run of its slot exits 0",
			json: cfcJSON(cfcFinding{Ref: "F1", Severity: "Minor", Status: "fixed", Slot: "primary"}), disp: "[]",
			want: 0, needle: []string{"FINDINGS-CLOSED"}, clean: true},
		{label: "case 30: a fixed Minor needs no re-run while a fixed Important of another slot still does",
			json: cfcJSON(cfcFinding{Ref: "F1", Severity: "Important", Status: "fixed", Slot: "failure-modes"}, cfcFinding{Ref: "F2", Severity: "Minor", Status: "fixed", Slot: "primary"}), disp: "[]",
			want: 1, needle: []string{"with no clean re-run dispatch of their slot in any later round: F1 (failure-modes)\n"}, absent: "F2"},
		{label: "case 31: a row folded by -supersedes onto a same-round row of another role is covered by that row's slot re-run",
			json: cfcJSON(cfcFinding{Ref: "F1", Severity: "Important", Status: "fixed", Round: 1, Slot: "primary"}, cfcFinding{Ref: "F2", Severity: "Important", Status: "fixed", Round: 1, Slot: "bugbot", Supersedes: "F1"}),
			disp: cfcDispatchJSON(cfcDispatch{Key: "panel-2-primary", Role: "reviewer", Slot: "primary", Outcome: "completed"}),
			want: 0, needle: []string{"FINDINGS-CLOSED"}, clean: true},
		{label: "case 31b: a cross-round -supersedes re-raise still needs its own slot's re-run",
			json: cfcJSON(cfcFinding{Ref: "F1", Severity: "Important", Status: "fixed", Round: 1, Slot: "primary"}, cfcFinding{Ref: "F2", Severity: "Important", Status: "fixed", Round: 2, Slot: "bugbot", Supersedes: "F1"}),
			disp: cfcDispatchJSON(cfcDispatch{Key: "panel-3-primary", Role: "reviewer", Slot: "primary", Outcome: "completed"}),
			want: 1, needle: []string{"in any later round: F2 (bugbot)\n"}, absent: "F1 ("},
		{label: "case 32: a re-run that ended clean covers the finding",
			json: cfcJSON(cfcFinding{Ref: "F1", Status: "fixed", Slot: "primary"}),
			disp: cfcDispatchJSON(cfcDispatch{Key: "panel-1-primary", Role: "reviewer", Slot: "primary", Outcome: "clean"}),
			want: 0, needle: []string{"FINDINGS-CLOSED"}, clean: true},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			disp := c.disp
			if disp == "" {
				disp = cfcDefaultDispatches
			}
			rc, out, errs := cfcRun(t, cfcBinDispatches(t, c.json, disp, "", 0), wt, "demo")
			if rc != c.want {
				t.Fatalf("expected exit %d, got %d\nstdout: %s\nstderr: %s", c.want, rc, out, errs)
			}
			for _, n := range c.needle {
				if !strings.Contains(out+errs, n) {
					t.Fatalf("expected output to name %q, got:\nstdout: %s\nstderr: %s", n, out, errs)
				}
			}
			if c.absent != "" && strings.Contains(errs, c.absent) {
				t.Fatalf("expected stderr not to name %q, got:\n%s", c.absent, errs)
			}
			if c.clean && errs != "" {
				t.Fatalf("expected no stderr, got:\n%s", errs)
			}
		})
	}

	t.Run("case 6: an unreachable store is cannot-answer", func(t *testing.T) {
		t.Parallel()
		rc, _, errs := cfcRun(t, cfcBin(t, "", "flow: connect: connection refused", 1), wt, "demo")
		if rc != 2 || errs != "check-panel-findings-closed: cannot read findings for 'demo' from the store — cannot determine anything: flow: connect: connection refused\n" {
			t.Fatalf("got exit %d, stderr %q", rc, errs)
		}
	})

	t.Run("case 6b: an unreachable dispatches read is cannot-answer", func(t *testing.T) {
		t.Parallel()
		bin := t.TempDir()
		writeFile(t, filepath.Join(bin, "findings.json"), cfcJSON(cfcFinding{Ref: "F1", Status: "fixed", Slot: "primary"}))
		writeFile(t, filepath.Join(bin, "dispatches.json"), "[]")
		writeExec(t, filepath.Join(bin, "flow"), `#!/usr/bin/env bash
dir="$(dirname -- "$0")"
case "$*" in
  *"record dispatches"*) echo "flow: connect: connection refused" >&2; exit 1 ;;
  *) cat "$dir/findings.json" ;;
esac
exit 0
`)
		rc, _, errs := cfcRun(t, bin, wt, "demo")
		if rc != 2 || errs != "check-panel-findings-closed: cannot read dispatches for 'demo' from the store — cannot determine anything: flow: connect: connection refused\n" {
			t.Fatalf("got exit %d, stderr %q", rc, errs)
		}
	})

	t.Run("case 7: a change name outside the allowlist is rejected before the store is read", func(t *testing.T) {
		t.Parallel()
		bin := t.TempDir()
		marker := filepath.Join(bin, "called.marker")
		writeExec(t, filepath.Join(bin, "flow"), "#!/usr/bin/env bash\ntouch \"$(dirname -- \"$0\")/called.marker\"\necho '[]'\n")
		for _, bad := range []string{"../../../planted/clear", "demo*", "demo/../demo", ".hidden", "demo?x"} {
			rc, _, errs := cfcRun(t, bin, wt, bad)
			if rc != 2 || !strings.Contains(errs, "is not a plain change name") {
				t.Fatalf("%s: got exit %d, stderr %q", bad, rc, errs)
			}
			if _, err := os.Stat(marker); err == nil {
				t.Fatalf("change name %s reached the store", bad)
			}
		}
	})

	t.Run("case 8a: no arguments at all exits 2", func(t *testing.T) {
		t.Parallel()
		rc, _, errs := cfcRun(t, cfcBin(t, "[]", "", 0))
		if rc != 2 || errs != "check-panel-findings-closed: not a directory: <missing>\n" {
			t.Fatalf("got exit %d, stderr %q", rc, errs)
		}
	})

	t.Run("case 8b: a missing change name exits 2", func(t *testing.T) {
		t.Parallel()
		rc, _, errs := cfcRun(t, cfcBin(t, "[]", "", 0), wt)
		if rc != 2 || errs != "usage: check-panel-findings-closed.sh <worktree> <change-name>\n" {
			t.Fatalf("got exit %d, stderr %q", rc, errs)
		}
	})

	t.Run("case 9: non-directory argument exits 2 and names it", func(t *testing.T) {
		t.Parallel()
		f := filepath.Join(t.TempDir(), "file")
		writeFile(t, f, "")
		rc, _, errs := cfcRun(t, cfcBin(t, "[]", "", 0), f, "demo")
		if rc != 2 || errs != "check-panel-findings-closed: not a directory: "+f+"\n" {
			t.Fatalf("got exit %d, stderr %q", rc, errs)
		}
	})

	t.Run("case 10: a stderr diagnostic on the store read does not break a clean answer", func(t *testing.T) {
		t.Parallel()
		rc, out, _ := cfcRun(t, cfcBin(t, cfcJSON(cfcFinding{Ref: "F1", Status: "fixed", Slot: "primary"}), "flow: using FLOW_ADDR=http://127.0.0.1:4174", 0), wt, "demo")
		if rc != 0 || out != "FINDINGS-CLOSED\n" {
			t.Fatalf("got exit %d, stdout %q", rc, out)
		}
	})
}

// A slot's report counting more Critical and Important findings on its own
// count lines than the store holds for that round and slot fails the close:
// kan-1000's panel-report-11-primary.md raised an Important nobody recorded.
func TestCheckPanelFindingsClosedReportCounts(t *testing.T) {
	t.Parallel()
	cases := []struct {
		label  string
		files  map[string]string
		json   string
		disp   string
		want   int
		needle string
	}{
		{label: "a report counting an Important the store lacks exits 1",
			files: map[string]string{"panel-report-11-primary.md": "Critical: 0\nImportant: 1\n\nbody\n"},
			json:  "[]", want: 1, needle: "panel-report-11-primary.md (report 1, recorded 0)"},
		{label: "the Important recorded fixed with a clean re-run exits 0",
			files: map[string]string{
				"panel-report-11-primary.md":    "Critical: 0\nImportant: 1\n",
				"panel-report-12-primary.md":    "clean\n",
				"panel-report-12-principles.md": "clean\n",
				"panel-fix-report-11.md":        "regression-test: F10 — none — fixture\n",
			},
			json: cfcJSON(cfcFinding{Ref: "F10", Severity: "Important", Status: "fixed", Round: 11, Slot: "primary+principles"}),
			disp: cfcDispatchJSON(cfcDispatch{Key: "panel-12-primary+principles", Role: "reviewer", Slot: "primary+principles", Outcome: "completed"}),
			want: 0},
		{label: "the Important recorded at another round exits 1",
			files: map[string]string{"panel-report-11-primary.md": "Critical: 0\nImportant: 1\n"},
			json:  cfcJSON(cfcFinding{Ref: "F10", Severity: "Important", Status: "withdrawn dup", Round: 10, Slot: "primary"}),
			want:  1, needle: "report 1, recorded 0"},
		{label: "reports with no count lines or a placeholder are not compared",
			files: map[string]string{
				"panel-report-2-primary.md":    "#### Important\n- something\n",
				"panel-report-3-principles.md": "no verbatim report captured — timeout\n",
			},
			json: "[]", want: 0},
		{label: "a withdrawn Critical of the slot counts as recorded",
			files: map[string]string{"panel-report-3-test-sanity.md": "Critical: 1\nImportant: 0\n"},
			json:  cfcJSON(cfcFinding{Ref: "F3", Severity: "Critical", Status: "withdrawn duplicate of F2", Round: 3, Slot: "test-sanity"}),
			want:  0},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			wt := t.TempDir()
			sdd := filepath.Join(wt, ".superpowers", "sdd")
			if err := os.MkdirAll(sdd, 0o755); err != nil {
				t.Fatal(err)
			}
			for n, body := range c.files {
				writeFile(t, filepath.Join(sdd, n), body)
			}
			disp := c.disp
			if disp == "" {
				disp = "[]"
			}
			rc, out, errs := cfcRun(t, cfcBinDispatches(t, c.json, disp, "", 0), wt, "demo")
			if rc != c.want {
				t.Fatalf("expected exit %d, got %d\nstdout: %s\nstderr: %s", c.want, rc, out, errs)
			}
			if c.needle != "" && !strings.Contains(errs, c.needle) {
				t.Fatalf("expected stderr to name %q, got:\n%s", c.needle, errs)
			}
		})
	}
}

// A completed re-run covers a fixed finding only once its
// panel-report-<round>-<id>.md is written -- not absent, empty or the
// parent's placeholder (KAN-915).
func TestCheckPanelFindingsClosedRerunNeedsItsReportFile(t *testing.T) {
	t.Parallel()
	solo := cfcDispatchJSON(cfcDispatch{Key: "panel-2-primary", Role: "reviewer", Slot: "primary", Outcome: "completed"})
	joined := cfcDispatchJSON(cfcDispatch{Key: "panel-2-bugbot+primary", Role: "reviewer", Slot: "bugbot+primary", Outcome: "completed"})
	f1 := cfcJSON(cfcFinding{Ref: "F1", Severity: "Important", Status: "fixed", Round: 1, Slot: "primary"})
	cases := []struct {
		label, json, disp string
		files             map[string]string
		want              int
		needle            string
	}{
		{label: "no report file exits 1", json: f1, disp: solo, want: 1, needle: "panel-report-2-primary.md"},
		{label: "the placeholder exits 1", json: f1, disp: solo,
			files: map[string]string{"panel-report-2-primary.md": "no verbatim report captured — timeout\n"}, want: 1, needle: "panel-report-2-primary.md"},
		{label: "a written report exits 0", json: f1, disp: solo,
			files: map[string]string{"panel-report-2-primary.md": "Critical: 0\nImportant: 0\n"}, want: 0},
		{label: "a joined re-run covers a primary finding with only its primary report", json: f1, disp: joined,
			files: map[string]string{"panel-report-2-primary.md": "Critical: 0\nImportant: 0\n", "panel-report-2-bugbot.md": "no verbatim report captured — timeout\n"}, want: 0},
		{label: "a joined finding needs every component's report",
			json: cfcJSON(cfcFinding{Ref: "F1", Severity: "Important", Status: "fixed", Round: 1, Slot: "bugbot+primary"}), disp: joined,
			files: map[string]string{"panel-report-2-primary.md": "Critical: 0\nImportant: 0\n"}, want: 1, needle: "panel-report-2-bugbot.md"},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			wt := t.TempDir()
			sdd := filepath.Join(wt, ".superpowers", "sdd")
			if err := os.MkdirAll(sdd, 0o755); err != nil {
				t.Fatal(err)
			}
			for n, body := range c.files {
				writeFile(t, filepath.Join(sdd, n), body)
			}
			writeFile(t, filepath.Join(sdd, "panel-fix-report-1.md"), "regression-test: F1 — none — fixture\n")
			rc, out, errs := cfcRun(t, cfcBinDispatches(t, c.json, c.disp, "", 0), wt, "demo")
			if rc != c.want {
				t.Fatalf("expected exit %d, got %d\nstdout: %s\nstderr: %s", c.want, rc, out, errs)
			}
			if c.needle != "" && !strings.Contains(errs, c.needle) {
				t.Fatalf("expected stderr to name %q, got:\n%s", c.needle, errs)
			}
		})
	}
}

// cfcSDD makes <wt>/.superpowers/sdd and writes files into it.
func cfcSDD(t *testing.T, wt string, files map[string]string) {
	t.Helper()
	sdd := filepath.Join(wt, ".superpowers", "sdd")
	if err := os.MkdirAll(sdd, 0o755); err != nil {
		t.Fatal(err)
	}
	for n, body := range files {
		writeFile(t, filepath.Join(sdd, n), body)
	}
}

// Every completed reviewer dispatch leaves each role's report file, the
// placeholder accepted (kan-1034: F7's round-0 slot left none, unnoticed).
func TestCheckPanelFindingsClosedEveryRoleLeavesAReport(t *testing.T) {
	t.Parallel()
	pair := cfcDispatchJSON(cfcDispatch{Key: "panel-0-mutation+failure-modes", Role: "reviewer", Slot: "mutation+failure-modes", Outcome: "completed"})
	const line = "completed reviewer dispatch(es) left no report file: "
	cases := []struct {
		label, disp, json string
		files             map[string]string
		want              int
		needles           []string
		absent            string
	}{
		{label: "no report files exits 1 naming both", disp: pair, want: 1,
			needles: []string{line + "panel-report-0-mutation.md panel-report-0-failure-modes.md\n"}},
		{label: "one written names only the other", disp: pair, want: 1,
			files:   map[string]string{"panel-report-0-mutation.md": "clean\n"},
			needles: []string{line + "panel-report-0-failure-modes.md\n"}, absent: "panel-report-0-mutation.md"},
		{label: "both placeholders exit 0", disp: pair, want: 0,
			files: map[string]string{"panel-report-0-mutation.md": "no verbatim report captured — timeout\n", "panel-report-0-failure-modes.md": "no verbatim report captured — timeout\n"}},
		{label: "an empty file counts as absent", disp: pair, want: 1,
			files:   map[string]string{"panel-report-0-mutation.md": "clean\n", "panel-report-0-failure-modes.md": ""},
			needles: []string{line + "panel-report-0-failure-modes.md\n"}},
		{label: "an aborted dispatch needs no report", want: 0,
			disp: cfcDispatchJSON(cfcDispatch{Key: "panel-0-mutation", Role: "reviewer", Slot: "mutation", Outcome: "aborted"})},
		{label: "an implementer dispatch needs no report", want: 0,
			disp: cfcDispatchJSON(cfcDispatch{Key: "panel-0-mutation", Role: "implementer", Slot: "mutation", Outcome: "completed"})},
		{label: "a panel-fix key needs no report", want: 0,
			disp: cfcDispatchJSON(cfcDispatch{Key: "panel-fix-0", Role: "reviewer", Slot: "mutation", Outcome: "completed"})},
		{label: "a missing re-run report is named once", want: 1,
			json:    cfcJSON(cfcFinding{Ref: "F1", Severity: "Important", Status: "fixed", Round: 0, Slot: "primary"}),
			disp:    cfcDispatchJSON(cfcDispatch{Key: "panel-1-primary", Role: "reviewer", Slot: "primary", Outcome: "completed"}),
			needles: []string{"re-run(s) answered without writing their report file: panel-report-1-primary.md\n"}, absent: line},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			wt := t.TempDir()
			cfcSDD(t, wt, c.files)
			js := c.json
			if js == "" {
				js = "[]"
			}
			rc, out, errs := cfcRun(t, cfcBinDispatches(t, js, c.disp, "", 0), wt, "demo")
			if rc != c.want {
				t.Fatalf("expected exit %d, got %d\nstdout: %s\nstderr: %s", c.want, rc, out, errs)
			}
			for _, n := range c.needles {
				if !strings.Contains(errs, n) {
					t.Fatalf("expected stderr to name %q, got:\n%s", n, errs)
				}
			}
			if c.absent != "" && strings.Contains(errs, c.absent) {
				t.Fatalf("expected stderr not to name %q, got:\n%s", c.absent, errs)
			}
		})
	}
}

// A fixed finding's runnable reproducer resolves in the worktree or one of
// the state record's worktrees (kan-1034: F7's reproducer was lost).
func TestCheckPanelFindingsClosedFixedReproducerExists(t *testing.T) {
	t.Parallel()
	const tok = ".superpowers/sdd/reproducers/r0-x-1.sh"
	const line = "fixed finding(s) whose runnable reproducer resolves in no worktree of the change: "
	minor := func(status, repro string) string {
		return cfcJSON(cfcFinding{Ref: "F1", Severity: "Minor", Status: status, Slot: "primary", Reproducer: repro})
	}
	peer := t.TempDir()
	if err := os.MkdirAll(filepath.Join(peer, ".superpowers", "sdd", "reproducers"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(peer, tok), "#!/bin/sh\n")
	cases := []struct {
		label, json, state string
		stateRC            int
		inWT               bool
		want               int
		needle, absent     string
	}{
		{label: "present in the worktree exits 0", json: minor("fixed", tok+" --flag"), inWT: true, want: 0},
		{label: "absent everywhere exits 1 naming ref and token", json: minor("fixed", tok+" --flag"), want: 1, needle: line + "F1 (" + tok + ")\n"},
		{label: "present in a state-record worktree exits 0", json: minor("fixed", tok), state: `{"worktrees":{"` + peer + `":{}}}`, want: 0},
		{label: "the none form is never checked", json: minor("fixed", "none — cosmetic"), want: 0},
		{label: "an empty reproducer is never checked", json: minor("fixed", ""), want: 0},
		{label: "an open finding's absent reproducer is not on this line", json: minor("open", tok), want: 1, absent: line},
		{label: "a failed state read while a token is absent exits 2", json: minor("fixed", tok), stateRC: 1, want: 2},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			wt := t.TempDir()
			cfcSDD(t, wt, nil)
			if c.inWT {
				cfcSDD(t, wt, map[string]string{"reproducers/r0-x-1.sh": ""})
			}
			bin := cfcBinDispatches(t, c.json, "[]", "", 0)
			if c.state != "" {
				writeFile(t, filepath.Join(bin, "state.json"), c.state)
			}
			if c.stateRC != 0 {
				writeExec(t, filepath.Join(bin, "flow"), "#!/usr/bin/env bash\ndir=\"$(dirname -- \"$0\")\"\ncase \"$*\" in\n  *\"record dispatches\"*) cat \"$dir/dispatches.json\" ;;\n  *\"state get\"*) echo boom >&2; exit 1 ;;\n  *) cat \"$dir/findings.json\" ;;\nesac\n")
			}
			rc, out, errs := cfcRun(t, bin, wt, "demo")
			if rc != c.want {
				t.Fatalf("expected exit %d, got %d\nstdout: %s\nstderr: %s", c.want, rc, out, errs)
			}
			if c.needle != "" && !strings.Contains(errs, c.needle) {
				t.Fatalf("expected stderr to name %q, got:\n%s", c.needle, errs)
			}
			if c.absent != "" && strings.Contains(errs, c.absent) {
				t.Fatalf("expected stderr not to name %q, got:\n%s", c.absent, errs)
			}
		})
	}
}

// A fixed Critical or Important finding names a git-tracked suite test on a
// `regression-test:` line of a panel-fix report, or `none — <reason>`
// (kan-1019: three fixed Importants left only untracked reproducers).
func TestCheckPanelFindingsClosedFixedNeedsARegressionTest(t *testing.T) {
	t.Parallel()
	const line = "fixed Critical/Important finding(s) with no tracked regression test: "
	important := cfcJSON(cfcFinding{Ref: "F1", Severity: "Important", Status: "fixed", Round: 1, Slot: "primary"})
	critical := cfcJSON(cfcFinding{Ref: "F1", Severity: "Critical", Status: "fixed", Round: 1, Slot: "primary"})
	minor := cfcJSON(cfcFinding{Ref: "F1", Severity: "Minor", Status: "fixed", Round: 1, Slot: "primary"})
	rerun := cfcDispatchJSON(cfcDispatch{Key: "panel-2-primary", Role: "reviewer", Slot: "primary", Outcome: "completed"})
	// peer is a recorded worktree whose only tracked file is peer_test.go.
	peer := t.TempDir()
	gitRun(t, "", "init", "-q", peer)
	writeFile(t, filepath.Join(peer, "peer_test.go"), "package foo\n")
	gitRun(t, peer, "add", "peer_test.go")
	cases := []struct {
		label, json, report, state string
		stateRC, want              int
		needle                     string
	}{
		{label: "no line exits 1", json: important, want: 1, needle: "F1 (no regression-test: line)"},
		{label: "a fixed Critical with no line exits 1", json: critical, want: 1, needle: "F1 (no regression-test: line)"},
		{label: "an untracked path exits 1", json: important, report: "regression-test: F1 — untracked_test.go — TestX\n", want: 1, needle: "F1 (untracked_test.go is not tracked)"},
		{label: "a tracked path under .superpowers exits 1", json: important, report: "regression-test: F1 — .superpowers/sdd/reproducers/r.sh — r\n", want: 1, needle: "F1 (.superpowers/sdd/reproducers/r.sh is under .superpowers/)"},
		{label: "a tracked directory is not a test file", json: important, report: "regression-test: F1 — pkg — TestFoo\n", want: 1, needle: "F1 (pkg is not tracked)"},
		{label: "dot is not a test file", json: important, report: "regression-test: F1 — . — TestFoo\n", want: 1, needle: "F1 (. is not tracked)"},
		{label: "a committed test exits 0", json: important, report: "regression-test: F1 — pkg/foo_test.go — TestFoo\n", want: 0},
		{label: "a test tracked only in a recorded worktree exits 0", json: important, report: "regression-test: F1 — peer_test.go — TestPeer\n", state: `{"worktrees":{"` + peer + `":{}}}`, want: 0},
		{label: "an unreadable state record while the path is untracked here exits 2", json: important, report: "regression-test: F1 — peer_test.go — TestPeer\n", stateRC: 1, want: 2},
		{label: "the none form exits 0", json: important, report: "regression-test: F1 — none — prose-only fix\n", want: 0},
		{label: "the none form with no reason exits 1", json: important, report: "regression-test: F1 — none\n", want: 1, needle: "F1 (none form with no reason)"},
		{label: "a fixed Minor needs no line", json: minor, want: 0},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			wt := t.TempDir()
			gitRun(t, "", "init", "-q", wt)
			writeFile(t, filepath.Join(wt, "pkg", "foo_test.go"), "package foo\n")
			writeFile(t, filepath.Join(wt, "untracked_test.go"), "package foo\n")
			files := map[string]string{"panel-report-2-primary.md": "Critical: 0\nImportant: 0\n", "reproducers/r.sh": ""}
			if c.report != "" {
				files["panel-fix-report-1.md"] = c.report
			}
			cfcSDD(t, wt, files)
			gitRun(t, wt, "add", "pkg/foo_test.go")
			gitRun(t, wt, "add", "-f", ".superpowers/sdd/reproducers/r.sh")
			bin := cfcBinDispatches(t, c.json, rerun, "", 0)
			if c.state != "" {
				writeFile(t, filepath.Join(bin, "state.json"), c.state)
			}
			if c.stateRC != 0 {
				writeExec(t, filepath.Join(bin, "flow"), "#!/usr/bin/env bash\ndir=\"$(dirname -- \"$0\")\"\ncase \"$*\" in\n  *\"record dispatches\"*) cat \"$dir/dispatches.json\" ;;\n  *\"state get\"*) echo boom >&2; exit 1 ;;\n  *) cat \"$dir/findings.json\" ;;\nesac\n")
			}
			rc, out, errs := cfcRun(t, bin, wt, "demo")
			if rc != c.want {
				t.Fatalf("expected exit %d, got %d\nstdout: %s\nstderr: %s", c.want, rc, out, errs)
			}
			if c.needle != "" && !strings.Contains(errs, line+c.needle) {
				t.Fatalf("expected stderr to name %q, got:\n%s", line+c.needle, errs)
			}
		})
	}
}

// A gated per-task reviewer's rows (slot gated-reviewer, round 0) are no
// panel finding: closed by a clean task re-review, never by a panel re-run.
func TestCheckPanelFindingsClosedSkipsGatedRows(t *testing.T) {
	t.Parallel()
	const repro = "none — gated per-task review; its report file holds the evidence"
	findings := cfcJSON(
		cfcFinding{Ref: "F1", Severity: "Important", Status: "fixed", Slot: "gated-reviewer", Reproducer: repro},
		cfcFinding{Ref: "F2", Severity: "Critical", Status: "fixed", Slot: "gated-reviewer", Reproducer: repro},
	)
	dispatches := cfcDispatchJSON(cfcDispatch{Key: "task-1-reviewer", Role: "reviewer", Outcome: "clean"})
	fn := Registry["check-panel-findings-closed"]
	var out, errb bytes.Buffer
	env := Env{
		Getenv:     pathEnv(t.TempDir()),
		Dir:        t.TempDir(),
		Findings:   func(string) ([]byte, error) { return []byte(findings), nil },
		Dispatches: func(string) ([]byte, error) { return []byte(dispatches), nil },
	}
	if rc := fn([]string{t.TempDir(), "demo"}, env, &out, &errb); rc != 0 || !strings.Contains(out.String(), "FINDINGS-CLOSED") {
		t.Fatalf("rc %d stdout %q stderr %q, want 0 FINDINGS-CLOSED", rc, out.String(), errb.String())
	}
}
