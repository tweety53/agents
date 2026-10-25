package prompttuning

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sampleCase() (Case, Files) {
	c := Case{
		ID:           "demo-change-1",
		Type:         "reviewer",
		ReviewerKind: "miss",
		Scope:        "generic",
		Source: Source{
			Project: "agents", Change: "demo-change", Jira: "KAN-1", Repo: "/repo",
			FixSHA: "abc1234", FindingRef: "F-3", Round: 2, Slot: "primary", Severity: "Important",
		},
		Target:      "skills/flow/primary-reviewer-prompt.md",
		TargetSHA:   "def5678",
		DefectClass: "an error path left untested",
		WhyPicked:   "four parts",
		Before:      "old <script>alert(1)</script> text",
		After:       "new text",
	}
	f := Files{
		Finding:  "the finding sentence",
		BadDiff:  "--- a/x.go\n+++ b/x.go\n+bad",
		GoodDiff: "--- a/x.go\n+++ b/x.go\n+good",
		Task:     "none",
	}
	return c, f
}

func TestRenderEscapesAndOrders(t *testing.T) {
	c, f := sampleCase()
	out, err := Render(c, f)
	if err != nil {
		t.Fatal(err)
	}
	page := string(out)
	if strings.Contains(page, "<script") {
		t.Errorf("page carries a <script tag:\n%s", page)
	}
	if !strings.Contains(page, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Errorf("before text not escaped:\n%s", page)
	}
	if strings.Contains(page, "http") {
		t.Errorf("page carries an http URL")
	}
	last := -1
	for _, h := range sectionHeadings {
		i := strings.Index(page, "<h2>"+h+"</h2>")
		if i < 0 {
			t.Fatalf("heading %q missing", h)
		}
		if i < last {
			t.Errorf("heading %q out of order", h)
		}
		last = i
	}
	if !strings.HasSuffix(page, "</html>\n") || strings.HasSuffix(page, "\n\n") {
		t.Errorf("page does not end with a single newline")
	}
}

func TestRenderIsDeterministic(t *testing.T) {
	c, f := sampleCase()
	a, err := Render(c, f)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Render(c, f)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Error("two renders differ")
	}
}

func TestLoadNamesMissingFile(t *testing.T) {
	dir := t.TempDir()
	c, _ := sampleCase()
	data, _ := json.Marshal(c)
	for name, body := range map[string]string{
		"case.json": string(data), "finding.md": "x", "bad.diff": "x", "task.md": "none",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	_, _, err := Load(dir)
	if err == nil || !strings.Contains(err.Error(), "good.diff") {
		t.Errorf("Load error = %v, want one naming good.diff", err)
	}
}
