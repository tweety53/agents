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
	Slot                  string
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
	script := "#!/usr/bin/env bash\ndir=\"$(dirname -- \"$0\")\"\n"
	if stderrLine != "" {
		script += "echo '" + stderrLine + "' >&2\n"
	}
	script += "case \"$*\" in\n"
	script += "  *\"record dispatches\"*) cat \"$dir/dispatches.json\" ;;\n"
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

func TestCheckPanelFindingsClosed(t *testing.T) {
	t.Parallel()
	wt := t.TempDir()
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
