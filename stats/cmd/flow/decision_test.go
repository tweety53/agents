package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// decisionRun drives `flow decision <args>` through run, the way the CLI
// dispatches it, with body written to a decision.json when non-empty.
func decisionRun(t *testing.T, body string, env testEnv, args ...string) (int, string, string) {
	t.Helper()
	full := []string{"decision"}
	for _, a := range args {
		if a == "<file>" {
			p := filepath.Join(t.TempDir(), "decision.json")
			if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
			a = p
		}
		full = append(full, a)
	}
	var out, errb bytes.Buffer
	rc := run(context.Background(), full, env.get, strings.NewReader(""), &out, &errb)
	return rc, out.String(), errb.String()
}

// TestDecisionRenderGolden pins the `## Decision` block: one line each for
// the groups, the panel's bundles and visual verification, a reason beside
// each group and bundle, overrides only when given, and any key a decision
// recorded before the decider carried ignored.
// smoke: core — the Decision block every run prints
func TestDecisionRenderGolden(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	cases := []struct {
		name, body, want string
	}{
		{"decider", `{"groups":[{"tasks":["1","2"],"reason":"2 builds on 1"},{"tasks":["3"],"reason":"independent"}],
			"panel":[{"slots":["primary","principles"],"reason":"always"},{"slots":["mutation"],"reason":"new logic"}],
			"visual":{"verify":"skipped","reason":"no page changes"},"overrides":[]}`,
			"## Decision\n\n" +
				"- **Groups:** 1+2 (2 builds on 1) · 3 (independent)\n" +
				"- **Panel:** primary+principles (always) · mutation (new logic)\n" +
				"- **Visual verify:** skipped — no page changes\n"},
		{"overrides and old keys", `{"class":"big","execution":"sdd","groups":[{"tasks":["1"],"bundles":[1],"reason":"one"}],
			"panel":[{"slots":["primary"],"reason":"r"}],
			"visual":{"verify":"not configured","reason":""},"overrides":["implementer opus/flow-medium: operator asked"]}`,
			"## Decision\n\n" +
				"- **Groups:** 1 (one)\n" +
				"- **Panel:** primary (r)\n" +
				"- **Visual verify:** not configured\n" +
				"- **Overrides:** implementer opus/flow-medium: operator asked\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			rc, out, errOut := decisionRun(t, c.body, env, "render", "-file", "<file>", "-session-model", "claude-opus-5-5")
			if rc != 0 {
				t.Fatalf("exit %d: %s", rc, errOut)
			}
			if out != c.want {
				t.Errorf("got:\n%s\nwant:\n%s", out, c.want)
			}
		})
	}
}

// TestDecisionRenderRefusals pins exit 2 on a usage error and on a body a
// consumer could not act on.
func TestDecisionRenderRefusals(t *testing.T) {
	t.Parallel()
	env := newTestEnv(t)
	ok := `"visual":{"verify":"required","reason":"r"}`
	cases := []struct{ name, body, want string }{
		{"not json", `nope`, "not a decision"},
		{"no groups", `{"panel":[],` + ok + `}`, "groups is empty"},
		{"a group without tasks", `{"groups":[{"tasks":[]}],` + ok + `}`, "groups[0] names no tasks"},
		{"a bundle of three", `{"groups":[{"tasks":["1"]}],"panel":[{"slots":["a","b","c"]}],` + ok + `}`, "panel[0] holds 3 slots"},
		{"an empty bundle", `{"groups":[{"tasks":["1"]}],"panel":[{"slots":[]}],` + ok + `}`, "panel[0] holds 0 slots"},
		{"no visual", `{"groups":[{"tasks":["1"]}]}`, "visual.verify"},
		{"a bad visual", `{"groups":[{"tasks":["1"]}],"visual":{"verify":"maybe"}}`, "visual.verify \"maybe\""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			rc, _, errOut := decisionRun(t, c.body, env, "render", "-file", "<file>")
			if rc != 2 || !strings.Contains(errOut, c.want) {
				t.Errorf("exit %d, stderr %q; want 2 naming %q", rc, errOut, c.want)
			}
		})
	}
	for _, args := range [][]string{{}, {"draw"}, {"render"}, {"render", "-reviewers", "x", "-file", "<file>"}} {
		if rc, _, _ := decisionRun(t, `{}`, env, args...); rc != 2 {
			t.Errorf("args %q: exit %d, want 2", args, rc)
		}
	}
}
