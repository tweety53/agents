package repotest

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// fenced is the body of the first ```<lang> block following marker in the file at rel.
func fenced(t *testing.T, rel, marker, lang string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repoRoot(t), rel))
	if err != nil {
		t.Fatal(err)
	}
	_, rest, ok := strings.Cut(string(b), marker)
	if !ok {
		t.Fatalf("%s: no %q", rel, marker)
	}
	_, body, ok := strings.Cut(rest, "\n```"+lang+"\n")
	if !ok {
		t.Fatalf("%s: no ```%s block after %q", rel, lang, marker)
	}
	body, _, _ = strings.Cut(body, "\n```\n")
	return body
}

// TestSelfReviewDetailExampleHasTheHeadings: the worked `suggested` -detail body in
// skills/flow-self-review/SKILL.md carries the required headings, in order, with only the two
// optional ones allowed absent, and the guard line its Effort and risk section asks for.
func TestSelfReviewDetailExampleHasTheHeadings(t *testing.T) {
	body := fenced(t, "skills/flow-self-review/SKILL.md", "A worked `suggested` `-detail`", "markdown")
	order := []string{"Proposal", "What happened", "Examples", "Current text", "Proposed change", "Effort and risk", "Benefit", "Done when", "Trace"}
	optional := map[string]bool{"Current text": true, "Proposed change": true}
	var got []string
	for _, m := range regexp.MustCompile(`(?m)^## (.+)$`).FindAllStringSubmatch(body, -1) {
		got = append(got, m[1])
	}
	var want []string
	for _, h := range order {
		if !optional[h] || slices.Contains(got, h) {
			want = append(want, h)
		}
	}
	if !slices.Equal(got, want) {
		t.Errorf("detail example headings = %q, want %q", got, want)
	}
	for _, l := range []string{"Effort: ", "Touches: ", "Risk: ", "Guard: "} {
		if !strings.Contains(body, "\n"+l) {
			t.Errorf("detail example's Effort and risk lacks a %q line", l)
		}
	}
	if !strings.Contains(body, "\n- [ ] ") {
		t.Error("detail example's Done when has no - [ ] item")
	}
}

// TestExecutorBriefExampleIsComplete: the worked executor brief in
// skills/flow-interactive/SKILL.md carries what **3. Execute** says a brief carries — the baseline
// pointer, the step range, the landing grant and the report shape.
func TestExecutorBriefExampleIsComplete(t *testing.T) {
	body := fenced(t, "skills/flow-interactive/SKILL.md", "A worked executor brief", "text")
	for _, s := range []string{
		"Role: operator-asked",
		"Before anything else, read `~/.claude/rules/agent-baseline.md` and follow it for this whole task.",
		"Include this instruction verbatim in any prompt you write for another agent.",
		"steps 3a–7 of `skills/flow-interactive/SKILL.md`",
		"landing grant",
		"Verdict, Findings, Decisions, Not done",
		"commit: request-",
		"base sha",
		"Red:",
	} {
		if !strings.Contains(body, s) {
			t.Errorf("executor brief example lacks %q", s)
		}
	}
}
