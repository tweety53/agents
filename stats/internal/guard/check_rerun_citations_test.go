package guard

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

// The contract is scripts/check-rerun-citations.sh's header comment; the
// kan-964 case is the "all fixed" report with no verdict line per finding.
func TestCheckRerunCitations(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	diff := filepath.Join(dir, "fix-round-1.diff")
	writeFile(t, diff, `diff --git a/src/a.go b/src/a.go
--- a/src/a.go
+++ b/src/a.go
@@ -10,3 +10,4 @@ func A() {
 	x := 1
-	y := 2
+	y := 3
+	z := 4
 	return
diff --git a/src/gone.go b/src/gone.go
--- a/src/gone.go
+++ /dev/null
@@ -1,2 +0,0 @@
-package gone
-var G = 1
diff --git a/q.sql b/q.sql
--- a/q.sql
+++ b/q.sql
@@ -1,2 +1,2 @@
 select 1;
--- old comment
+++ new comment
@@ -40,2 +40,2 @@
 a
-b
+c
diff --git a/b/x.go b/b/x.go
--- a/b/x.go
+++ /dev/null
@@ -3 +0,0 @@
-package x
`)
	cases := []struct {
		name, report string
		refs         []string
		want         int
		out          string
	}{
		{"cited verdicts pass", "# Re-review\n- verdict: F1 — fixed — `src/a.go:12`\nverdict: F2 — fixed — agents:src/a.go:13\n* verdict: F3 — fixed — src/gone.go:2\nverdict: F4 — not fixed — none\n",
			[]string{"F1", "F2", "F3", "F4"}, 0, "CITED: F2 — fixed — agents:src/a.go:13"},
		{"kan-964: all fixed, no verdict line", "All findings fixed.\n", []string{"F1"}, 1, "UNCITED: F1 — no"},
		{"citation outside the hunks", "verdict: F1 — fixed — src/a.go:40\n", []string{"F1"}, 1, "lands in no hunk"},
		{"citation in an untouched file", "verdict: F1 — fixed — src/other.go:11\n", []string{"F1"}, 1, "lands in no hunk"},
		{"citation without a line", "verdict: F1 — fixed — src/a.go\n", []string{"F1"}, 1, "is not <path>:<line>"},
		{"neither verdict word", "verdict: F1 — looks good\n", []string{"F1"}, 1, "is neither"},
		{"comment lines inside a hunk are not headers", "verdict: F1 — fixed — q.sql:40\n", []string{"F1"}, 0, "CITED: F1"},
		{"a ./ prefix is stripped", "verdict: F1 — fixed — ./src/a.go:12\n", []string{"F1"}, 0, "CITED: F1"},
		{"only a/ is stripped from the old header", "verdict: F1 — fixed — b/x.go:3\n", []string{"F1"}, 0, "CITED: F1"},
		{"one finding missing", "verdict: F1 — fixed — src/a.go:11\n", []string{"F1", "F2"}, 1, "UNCITED: F2"},
	}
	for i, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			report := filepath.Join(dir, "report-"+string(rune('a'+i))+".md")
			writeFile(t, report, c.report)
			var out, errb bytes.Buffer
			got := checkRerunCitations(append([]string{report, diff}, c.refs...), Env{}, &out, &errb)
			if got != c.want || !strings.Contains(out.String(), c.out) {
				t.Fatalf("exit %d (want %d), stdout %q (want %q), stderr %q", got, c.want, out.String(), c.out, errb.String())
			}
		})
	}

	report := filepath.Join(dir, "report.md")
	writeFile(t, report, "verdict: F1 — fixed — src/a.go:11\n")
	for _, c := range []struct {
		name string
		args []string
		msg  string
	}{
		{"bad ref", []string{report, diff, "17"}, "not a finding ref"},
		{"missing diff", []string{report, filepath.Join(dir, "absent.diff"), "F1"}, "cannot read"},
		{"too few arguments", []string{report, diff}, "usage"},
	} {
		var out, errb bytes.Buffer
		if got := checkRerunCitations(c.args, Env{}, &out, &errb); got != 2 || !strings.Contains(errb.String(), c.msg) || out.Len() != 0 {
			t.Errorf("%s: exit %d, stdout %q, stderr %q; want 2 with %q", c.name, got, out.String(), errb.String(), c.msg)
		}
	}
}
