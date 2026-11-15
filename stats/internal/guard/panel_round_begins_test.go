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
	wt := t.TempDir()
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
	code := panelRoundBegins([]string{wt, "kan-1", "4", "mf-k1", "primary:opus:low:a515c200", "test-sanity+call-efficiency:sonnet:low"}, env, &out, &errb)
	if code != 0 {
		t.Fatalf("exit %d, stderr %s", code, errb.String())
	}
	calls, _ := os.ReadFile(log)
	want := "record dispatch begin -C " + wt + " -change kan-1 -role reviewer -slot primary -model opus -effort low -diff-base a515c200 -key panel-4-primary -session-token mf-k1\n" +
		"record dispatch begin -C " + wt + " -change kan-1 -role reviewer -slot test-sanity+call-efficiency -model sonnet -effort low -key panel-4-test-sanity+call-efficiency -session-token mf-k1\n"
	if string(calls) != want {
		t.Errorf("flow calls:\n%s\nwant:\n%s", calls, want)
	}
	if !strings.Contains(out.String(), "panel-4-primary: recorded: dispatch 7\n") {
		t.Errorf("stdout = %q, want each key beside flow's answer", out.String())
	}

	out.Reset()
	if code := panelRoundBegins([]string{wt, "kan-1", "4", "mf-k1", "bad:opus:low", "primary:opus:low"}, env, &out, &errb); code != 1 {
		t.Errorf("a refused begin: exit %d, want 1", code)
	}
	if !strings.Contains(out.String(), "panel-4-primary: recorded") {
		t.Errorf("a refused begin stopped the round's later begins: %q", out.String())
	}
	for _, args := range [][]string{
		{wt, "kan-1", "$R", "mf-k1", "primary:opus:low"},
		{wt, "kan-1", "4", "mf-k1", "primary:opus"},
		{wt, "kan-1", "4", "mf-k1"},
	} {
		if code := panelRoundBegins(args, env, &out, &errb); code != 2 {
			t.Errorf("%q: exit %d, want 2", args, code)
		}
	}
}

// TestPanelRoundBeginsLeavesThePassRowToTheBegin pins kan-1086's
// self-review, F8: each slot's `flow record dispatch begin` writes its own
// pass row, so the guard makes the begins and no `record pass` call.
func TestPanelRoundBeginsLeavesThePassRowToTheBegin(t *testing.T) {
	dir := t.TempDir()
	wt := t.TempDir()
	log := filepath.Join(dir, "calls")
	stub := "#!/bin/sh\necho \"$*\" >> " + log + "\necho recorded\n"
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
	if code := panelRoundBegins([]string{wt, "kan-1", "4", "mf-k1", "primary:opus:low:a515c200", "test-sanity+call-efficiency:sonnet:low"}, env, &out, &errb); code != 0 {
		t.Fatalf("exit %d, stderr %s", code, errb.String())
	}
	calls, _ := os.ReadFile(log)
	lines := strings.Split(strings.TrimSpace(string(calls)), "\n")
	if len(lines) != 2 || strings.Contains(string(calls), "record pass") {
		t.Errorf("flow calls:\n%s\nwant the two begins and no record pass", calls)
	}
	if strings.Contains(out.String(), "pass:") {
		t.Errorf("stdout = %q, want no pass line", out.String())
	}
}

// TestPanelRoundBeginsSnapshotsThePlanTree pins self-review-kan-1110 F16: the
// begins call takes the round's plan-tree snapshot itself, once per round,
// and a failed snapshot records no begin.
func TestPanelRoundBeginsSnapshotsThePlanTree(t *testing.T) {
	dir := t.TempDir()
	log := filepath.Join(dir, "calls")
	stub := "#!/bin/sh\necho \"$*\" >> " + log + "\necho recorded\n"
	if err := os.WriteFile(filepath.Join(dir, "flow"), []byte(stub), 0o755); err != nil {
		t.Fatal(err)
	}
	env := Env{Getenv: func(k string) string {
		if k == "PATH" {
			return dir
		}
		return ""
	}}
	wt := t.TempDir()
	tasks := filepath.Join(wt, "spectre", "changes", "demo", "tasks.md")
	writeFile(t, tasks, "- [ ] 1. a\n")
	var out, errb bytes.Buffer
	begins := func(wt string) int {
		return panelRoundBegins([]string{wt, "demo", "2", "mf-x-20261114T120000Z", "primary:opus:low"}, env, &out, &errb)
	}
	if code := begins(wt); code != 0 {
		t.Fatalf("exit %d, stderr %s", code, errb.String())
	}
	snap := filepath.Join(wt, ".superpowers", "sdd", "plan-snapshot-panel-2.txt")
	first, err := os.ReadFile(snap)
	if err != nil {
		t.Fatalf("no snapshot written: %v", err)
	}
	verify := func() (int, string) {
		var o, e bytes.Buffer
		code := checkPlanUnchanged([]string{"verify", wt, "demo", snap}, env, &o, &e)
		return code, e.String()
	}
	if code, e := verify(); code != 0 {
		t.Fatalf("verify right after the snapshot: exit %d %s", code, e)
	}
	writeFile(t, tasks, "- [ ] 1. rewritten\n")
	later := panelRoundBegins([]string{wt, "demo", "2", "mf-x-20261114T120000Z", "mutation+failure-modes:opus:low"}, env, &out, &errb)
	if later != 0 {
		t.Fatalf("later call: exit %d, stderr %s", later, errb.String())
	}
	if again, _ := os.ReadFile(snap); !bytes.Equal(again, first) {
		t.Error("pass 1's later call re-snapshotted the plan tree")
	}
	if code, e := verify(); code != 1 || !strings.Contains(e, "tasks.md") {
		t.Errorf("verify after the rewrite: exit %d %s, want 1 naming tasks.md", code, e)
	}
	// A re-entered run reusing the round number overwrites the stale file.
	if code := begins(wt); code != 0 {
		t.Fatalf("re-entered call: exit %d, stderr %s", code, errb.String())
	}
	if code, e := verify(); code != 0 {
		t.Errorf("a re-entered round kept a stale snapshot: exit %d %s", code, e)
	}

	bad := t.TempDir()
	writeFile(t, filepath.Join(bad, ".superpowers"), "a file")
	_ = os.Remove(log)
	errb.Reset()
	if code := begins(bad); code != 2 {
		t.Errorf("unwritable snapshot dir: exit %d, want 2", code)
	}
	if calls, _ := os.ReadFile(log); strings.Contains(string(calls), "record dispatch begin") {
		t.Errorf("a failed snapshot still recorded a begin: %s", calls)
	}
}

// TestPanelRoundBeginsHoldsALoneMutation pins decision
// fix-latent-bundle-matchers: once the slot triggers shrink
// mutation+failure-modes to mutation alone, pass 1's later call still keeps
// the round's snapshot, while a call carrying primary beside mutation takes
// a fresh one.
func TestPanelRoundBeginsHoldsALoneMutation(t *testing.T) {
	dir := t.TempDir()
	stub := "#!/bin/sh\necho recorded\n"
	if err := os.WriteFile(filepath.Join(dir, "flow"), []byte(stub), 0o755); err != nil {
		t.Fatal(err)
	}
	env := Env{Getenv: func(k string) string {
		if k == "PATH" {
			return dir
		}
		return ""
	}}
	wt := t.TempDir()
	tasks := filepath.Join(wt, "spectre", "changes", "demo", "tasks.md")
	writeFile(t, tasks, "- [ ] 1. a\n")
	snap := filepath.Join(wt, ".superpowers", "sdd", "plan-snapshot-panel-1.txt")
	begins := func(slot string) {
		t.Helper()
		var out, errb bytes.Buffer
		if code := panelRoundBegins([]string{wt, "demo", "1", "mf-x-20261114T120000Z", slot}, env, &out, &errb); code != 0 {
			t.Fatalf("%s: exit %d, stderr %s", slot, code, errb.String())
		}
	}
	begins("primary:opus:low")
	first, err := os.ReadFile(snap)
	if err != nil {
		t.Fatalf("no snapshot written: %v", err)
	}
	writeFile(t, tasks, "- [ ] 1. rewritten\n")
	begins("mutation:opus:low")
	if again, _ := os.ReadFile(snap); !bytes.Equal(again, first) {
		t.Error("a later call recording mutation alone re-snapshotted the plan tree")
	}
	begins("primary+mutation:opus:low")
	if again, _ := os.ReadFile(snap); bytes.Equal(again, first) {
		t.Error("a call carrying primary beside mutation kept the stale snapshot")
	}
}
