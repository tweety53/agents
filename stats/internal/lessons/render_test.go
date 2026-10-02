package lessons

import (
	"strings"
	"testing"
)

func TestRenderBriefsThenMentions(t *testing.T) {
	res := &Result{
		Topic: "flake-bisection",
		Briefs: []Brief{{
			Repo:    "/ws/gymie",
			Path:    "docs/briefs/flake-bisection.md",
			Slug:    "flake-bisection",
			Title:   "Flake bisection brief",
			Content: "# Flake bisection brief\n\nStart small.",
		}},
		Mentions: []Mention{{
			Repo:   "/ws/gymie",
			Change: "kan-527-fix-the-flake",
			Path:   "spectre/changes/archive/kan-527-fix-the-flake/narrative.md",
			Lines:  []string{"- bisected on the smallest pair"},
		}},
	}
	out := Render(res)
	for _, want := range []string{
		"# Lessons: flake-bisection",
		"found: 1 briefs, 1 narrative mentions",
		"## brief: gymie/docs/briefs/flake-bisection.md",
		"# Flake bisection brief",
		"## narrative: gymie/spectre/changes/archive/kan-527-fix-the-flake/narrative.md",
		"- bisected on the smallest pair",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("render missing %q\n%s", want, out)
		}
	}
	if strings.Index(out, "## brief:") > strings.Index(out, "## narrative:") {
		t.Errorf("brief must render before any narrative mention\n%s", out)
	}
}

func TestRenderUnreadableNoteLeadsAndEmptyAnswerStatesItself(t *testing.T) {
	out := Render(&Result{Topic: "topic", Unreadable: []string{"/gone"}})
	if !strings.Contains(out, "note: /gone could not be read") {
		t.Errorf("render missing the unreadable note\n%s", out)
	}
	if !strings.Contains(out, "found: 0 briefs, 0 narrative mentions") {
		t.Errorf("empty answer must state its own emptiness\n%s", out)
	}
	if strings.Contains(out, "## ") {
		t.Errorf("an empty answer renders no sections\n%s", out)
	}
}
