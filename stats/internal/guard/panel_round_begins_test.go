package guard

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPanelRoundBeginsRecordsLiteralKeys pins kan-1026's self-review: each
// slot's begin carries panel-<round>-<slot> as a literal key, in slot
// order, with its own model, effort and diff base.
func TestPanelRoundBeginsRecordsLiteralKeys(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	stub := "#!/bin/sh\necho \"$*\" >> " + log + "\ncase \"$*\" in *-key\\ panel-4-bad*) echo refused; exit 1;; esac\necho \"recorded: dispatch 7\"\n"
	if err := os.WriteFile(filepath.Join(dir, "flow"), []byte(stub), 0o755); err != nil {
		t.Fatal(err)
	}
	env := Env{Getenv: func(k string) string {
		if k == "PATH" {
			return dir
		}
		return ""
	}}

	var out, errb bytes.Buffer
	code := panelRoundBegins([]string{"/wt", "kan-1", "4", "mf-k1", "primary:opus:low:a515c200", "test-sanity+call-efficiency:sonnet:low"}, env, &out, &errb)
	if code != 0 {
		t.Fatalf("exit %d, stderr %s", code, errb.String())
	}
	calls, _ := os.ReadFile(log)
	want := "record dispatch begin -C /wt -change kan-1 -role reviewer -slot primary -model opus -effort low -diff-base a515c200 -key panel-4-primary -session-token mf-k1\n" +
		"record dispatch begin -C /wt -change kan-1 -role reviewer -slot test-sanity+call-efficiency -model sonnet -effort low -key panel-4-test-sanity+call-efficiency -session-token mf-k1\n"
	if string(calls) != want {
		t.Errorf("flow calls:\n%s\nwant:\n%s", calls, want)
	}
	if !strings.Contains(out.String(), "panel-4-primary: recorded: dispatch 7\n") {
		t.Errorf("stdout = %q, want each key beside flow's answer", out.String())
	}

	out.Reset()
	if code := panelRoundBegins([]string{"/wt", "kan-1", "4", "mf-k1", "bad:opus:low", "primary:opus:low"}, env, &out, &errb); code != 1 {
		t.Errorf("a refused begin: exit %d, want 1", code)
	}
	if !strings.Contains(out.String(), "panel-4-primary: recorded") {
		t.Errorf("a refused begin stopped the round's later begins: %q", out.String())
	}
	for _, args := range [][]string{
		{"/wt", "kan-1", "$R", "mf-k1", "primary:opus:low"},
		{"/wt", "kan-1", "4", "mf-k1", "primary:opus"},
		{"/wt", "kan-1", "4", "mf-k1"},
	} {
		if code := panelRoundBegins(args, env, &out, &errb); code != 2 {
			t.Errorf("%q: exit %d, want 2", args, code)
		}
	}
}
