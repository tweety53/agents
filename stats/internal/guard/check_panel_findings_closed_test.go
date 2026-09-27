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
// **Panel re-runs**). The bash harness's stub `flow` printed a canned JSON
// array for `record findings`; here that stub sits on Env's PATH the same
// way, so every case runs the production exec path.

// cfcFinding is one row of `flow record findings`' JSON; an empty Severity
// is left out of the row, as the bash harness's ref/status pairs left it.
type cfcFinding struct {
	Ref, Status, Severity string
	Round                 int
}

func cfcJSON(fs ...cfcFinding) string {
	out := []map[string]any{}
	for _, f := range fs {
		row := map[string]any{"ref": f.Ref, "status": f.Status, "round": f.Round}
		if f.Severity != "" {
			row["severity"] = f.Severity
		}
		out = append(out, row)
	}
	b, _ := json.Marshal(out)
	return string(b)
}

// cfcBin is a directory holding a stub `flow` that prints stderrLine (when
// set) on stderr, then body on stdout, and exits rc.
func cfcBin(t *testing.T, body, stderrLine string, rc int) string {
	t.Helper()
	bin := t.TempDir()
	writeFile(t, filepath.Join(bin, "findings.json"), body)
	script := "#!/usr/bin/env bash\ndir=\"$(dirname -- \"$0\")\"\n"
	if stderrLine != "" {
		script += "echo '" + stderrLine + "' >&2\n"
	}
	script += "cat \"$dir/findings.json\"\nexit " + strconv.Itoa(rc) + "\n"
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
		want   int
		needle []string
		clean  bool // no stderr at all
	}{
		{label: "case 1: every finding closed exits 0 with no stderr", json: cfcJSON(cfcFinding{Ref: "F1", Status: "fixed"}, cfcFinding{Ref: "F2", Status: "fixed"}), want: 0, needle: []string{"FINDINGS-CLOSED"}, clean: true},
		{label: "case 2: one still-open finding exits 1, naming it", json: cfcJSON(cfcFinding{Ref: "F1", Status: "open"}, cfcFinding{Ref: "F2", Status: "fixed"}), want: 1, needle: []string{"check-panel-findings-closed: finding(s) still open: F1\n"}},
		{label: "case 3: several still-open findings exit 1, naming both", json: cfcJSON(cfcFinding{Ref: "F1", Status: "open"}, cfcFinding{Ref: "F2", Status: "fixed"}, cfcFinding{Ref: "F3", Status: "open"}), want: 1, needle: []string{"finding(s) still open: F1 F3\n"}},
		{label: "case 4: a withdrawn finding counts as closed, exits 0", json: cfcJSON(cfcFinding{Ref: "F1", Status: "withdrawn — not a real defect"}), want: 0},
		{label: "case 4b: a deferred finding counts as closed, exits 0", json: cfcJSON(cfcFinding{Ref: "F1", Status: "deferred cosmetic, not worth a fix round"}), want: 0},
		{label: "case 5: zero findings exits 0", json: "[]", want: 0},
		{label: "case 11: a Minor deferred beside an Important of the same round exits 1, naming it and the round",
			json: cfcJSON(cfcFinding{Ref: "F1", Severity: "Important", Status: "fixed"}, cfcFinding{Ref: "F2", Severity: "Minor", Status: "deferred cosmetic"}, cfcFinding{Ref: "F3", Severity: "minor", Status: "deferred doc-only"}),
			want: 1, needle: []string{"check-panel-findings-closed: round 0 raised a Critical or Important, so its Minor finding(s) go to that same fix, never deferred: F2 F3\n"}},
		{label: "case 12: a Minor deferred beside a Critical of the same round exits 1",
			json: cfcJSON(cfcFinding{Ref: "F1", Severity: "CRITICAL", Status: "fixed", Round: 2}, cfcFinding{Ref: "F2", Severity: "Minor", Status: "deferred cosmetic", Round: 2}),
			want: 1, needle: []string{"round 2 raised a Critical or Important", ": F2\n"}},
		{label: "case 13: a Minor-only round with its Minors deferred exits 0",
			json: cfcJSON(cfcFinding{Ref: "F1", Severity: "Minor", Status: "deferred cosmetic"}, cfcFinding{Ref: "F2", Severity: "Minor", Status: "deferred doc-only"}),
			want: 0, needle: []string{"FINDINGS-CLOSED"}, clean: true},
		{label: "case 14: Important and Minor of one round, the Minor fixed, exits 0",
			json: cfcJSON(cfcFinding{Ref: "F1", Severity: "Important", Status: "fixed"}, cfcFinding{Ref: "F2", Severity: "Minor", Status: "fixed"}),
			want: 0, needle: []string{"FINDINGS-CLOSED"}, clean: true},
		{label: "case 15: Minors deferred in a round whose Important is in another round exits 0",
			json: cfcJSON(cfcFinding{Ref: "F1", Severity: "Important", Status: "fixed", Round: 0}, cfcFinding{Ref: "F2", Severity: "Minor", Status: "deferred cosmetic", Round: 1}),
			want: 0, needle: []string{"FINDINGS-CLOSED"}, clean: true},
		{label: "case 16: a withdrawn Important took no fix round, so its round's Minors defer, exits 0",
			json: cfcJSON(cfcFinding{Ref: "F1", Severity: "Important", Status: "withdrawn — premise not in the primary source"}, cfcFinding{Ref: "F2", Severity: "Minor", Status: "deferred cosmetic"}),
			want: 0, needle: []string{"FINDINGS-CLOSED"}, clean: true},
		{label: "case 17: an open finding and a wrongly deferred Minor are both reported",
			json: cfcJSON(cfcFinding{Ref: "F1", Severity: "Important", Status: "open"}, cfcFinding{Ref: "F2", Severity: "Minor", Status: "deferred cosmetic"}),
			want: 1, needle: []string{"finding(s) still open: F1\n", "never deferred: F2\n"}},
		{label: "case 18: a non-string status cannot be judged, exits 2", json: `[{"ref":"F1","status":null}]`, want: 2, needle: []string{"check-panel-findings-closed: jq failed — cannot determine anything\n"}},
		{label: "case 19: empty store output cannot be judged, exits 2", json: "", want: 2, needle: []string{"jq failed — cannot determine anything"}},
	}
	for _, c := range cases {
		t.Run(c.label, func(t *testing.T) {
			t.Parallel()
			rc, out, errs := cfcRun(t, cfcBin(t, c.json, "", 0), wt, "demo")
			if rc != c.want {
				t.Fatalf("expected exit %d, got %d\nstdout: %s\nstderr: %s", c.want, rc, out, errs)
			}
			for _, n := range c.needle {
				if !strings.Contains(out+errs, n) {
					t.Fatalf("expected output to name %q, got:\nstdout: %s\nstderr: %s", n, out, errs)
				}
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
		rc, out, _ := cfcRun(t, cfcBin(t, cfcJSON(cfcFinding{Ref: "F1", Status: "fixed"}), "flow: using FLOW_ADDR=http://127.0.0.1:4174", 0), wt, "demo")
		if rc != 0 || out != "FINDINGS-CLOSED\n" {
			t.Fatalf("got exit %d, stdout %q", rc, out)
		}
	})
}
