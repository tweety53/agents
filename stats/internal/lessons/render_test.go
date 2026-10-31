package lessons

import (
	"strings"
	"testing"
)

func TestRenderBriefs(t *testing.T) {
	res := &Result{
		Topic: "flake-bisection",
		Briefs: []Brief{{
			Repo:    "/ws/gymie",
			Path:    "docs/briefs/flake-bisection.md",
			Slug:    "flake-bisection",
			Title:   "Flake bisection brief",
			Content: "# Flake bisection brief\n\nStart small.",
		}},
	}
	out := Render(res)
	for _, want := range []string{
		"# Lessons: flake-bisection",
		"found: 1 briefs\n",
		"## brief: gymie/docs/briefs/flake-bisection.md",
		"# Flake bisection brief",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("render missing %q\n%s", want, out)
		}
	}
}

func TestRenderUnreadableNoteLeadsAndEmptyAnswerStatesItself(t *testing.T) {
	out := Render(&Result{Topic: "topic", Unreadable: []string{"/gone"}})
	if !strings.Contains(out, "note: /gone could not be read") {
		t.Errorf("render missing the unreadable note\n%s", out)
	}
	if !strings.Contains(out, "found: 0 briefs\n") {
		t.Errorf("empty answer must state its own emptiness\n%s", out)
	}
	if strings.Contains(out, "## ") {
		t.Errorf("an empty answer renders no sections\n%s", out)
	}
}
